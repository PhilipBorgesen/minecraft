package profile

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PhilipBorgesen/minecraft/internal"
)

// LoadManyMaxSize is the maximum number of profiles which may be requested at
// once using LoadMany. If more are requested, the request may fail with an
// ErrMaxSizeExceeded error.
const LoadManyMaxSize int = 100

// Load fetches the profile currently associated with username. ctx must be
// non-nil. If no profile currently is associated with username, Load returns
// ErrNoSuchProfile. If an error is returned, p will be nil.
func Load(ctx context.Context, username string) (p *Profile, err error) {
	if username == "" {
		return nil, ErrNoSuchProfile
	}
	endpoint := fmt.Sprintf(loadURL, username)
	return loadByName(ctx, endpoint)
}

// LoadAtTime fetches the profile associated with username at the specified
// instant of time. ctx must be non-nil. If no profile was associated with
// username at the specified instant of time, LoadAtTime returns
// ErrNoSuchProfile. If an error is returned, p will be nil.
func LoadAtTime(ctx context.Context, username string, t time.Time) (p *Profile, err error) {
	if username == "" {
		return nil, ErrNoSuchProfile
	}
	endpoint := fmt.Sprintf(loadAtTimeURL, username, t.Unix())
	return loadByName(ctx, endpoint)
}

// Common implementation used by Load and LoadAtTime.
func loadByName(ctx context.Context, endpoint string) (p *Profile, err error) {
	j, err := internal.FetchJSON(ctx, client, endpoint)
	if err == nil {
		defer func() { // If JSON data isn't structured as expected
			if r := recover(); r != nil {
				err = &url.Error{Op: "Parse", URL: endpoint, Err: internal.ErrUnknownFormat}
			}
		}()
		return buildProfile(j.(map[string]interface{}), ErrNoSuchProfile)
	} else {
		return nil, transformError(err)
	}
}

// LoadByID fetches the profile identified by id. ctx must be non-nil. If no
// profile is identified by id, LoadByID returns ErrNoSuchProfile. If an error
// is returned, p will be nil.
func LoadByID(ctx context.Context, id string) (p *Profile, err error) {
	return LoadWithNameHistory(ctx, id)
}

// LoadWithNameHistory fetches the profile identified by id, incl. its name
// history. ctx must be non-nil. If no profile is identified by id,
// LoadWithNameHistory returns ErrNoSuchProfile. If an error is returned,
// p will be nil.
func LoadWithNameHistory(ctx context.Context, id string) (p *Profile, err error) {
	if id == "" {
		return nil, ErrNoSuchProfile
	}
	endpoint := fmt.Sprintf(loadWithNameHistoryURL, id)
	j, err := internal.FetchJSON(ctx, client, endpoint)
	if err == nil {
		defer func() { // If JSON data isn't structured as expected
			if r := recover(); r != nil {
				err = &url.Error{Op: "Parse", URL: endpoint, Err: internal.ErrUnknownFormat}
			}
		}()
		name, hist := buildHistory(j.([]interface{}))
		return &Profile{
			ID:          id,
			Name:        name,
			NameHistory: hist,
		}, nil
	} else {
		return nil, transformError(err)
	}
}

// LoadWithProperties fetches the profile identified by id, incl. its
// properties. ctx must be non-nil. If no profile is identified by id,
// LoadWithProperties returns ErrNoSuchProfile. If an error is returned,
// p will be nil.
//
// NB! For each profile, profile properties may only be requested once per
// minute.
func LoadWithProperties(ctx context.Context, id string) (p *Profile, err error) {
	if id == "" {
		return nil, ErrNoSuchProfile
	}
	endpoint := fmt.Sprintf(loadWithPropertiesURL, id)
	j, err := internal.FetchJSON(ctx, client, endpoint)
	if err == nil {
		defer func() { // If JSON data isn't structured as expected
			if r := recover(); r != nil {
				err = &url.Error{Op: "Parse", URL: endpoint, Err: internal.ErrUnknownFormat}
				p = nil
			}
		}()

		m := j.(map[string]interface{})
		p, err = buildProfile(m, ErrNoSuchProfile)
		if err != nil {
			return nil, err
		}

		p.Properties, err = buildProperties(m["properties"].([]interface{}))
		if err != nil {
			// Let the entire loading fail even if just property construction fails.
			// May always be changed later if this is too drastic.
			return nil, &url.Error{Op: "Parse", URL: endpoint, Err: err}
		}

		return p, nil
	} else {
		return nil, transformError(err)
	}
}

// LoadMany fetches multiple profiles by their currently associated usernames.
// Usernames associated with no profile are ignored and absent from the
// returned results. Duplicate usernames are only returned once, and ps will be
// nil if an error occurs. ctx must be non-nil.
//
// NB! Only a maximum of LoadManyMaxSize profiles may be fetched at once.
// If more are attempted loaded in the same operation, an ErrMaxSizeExceeded
// error is returned.
func LoadMany(ctx context.Context, usernames ...string) (ps []*Profile, err error) {
	if len(usernames) > LoadManyMaxSize {
		return nil, ErrMaxSizeExceeded{len(usernames)}
	}

	c := 0
	var users [LoadManyMaxSize]string
	for _, u := range usernames {
		// Remove empty usernames. They are not accepted by the Mojang API.
		if u != "" {
			users[c] = u
			c++
		}
	}

	if c == 0 {
		return nil, nil // No need to request anything
	}

	j, err := internal.ExchangeJSON(ctx, client, loadManyURL, users[:c])
	if err == nil {
		defer func() { // If JSON data isn't structured as expected
			if r := recover(); r != nil {
				err = &url.Error{Op: "Parse", URL: loadManyURL, Err: internal.ErrUnknownFormat}
				ps = nil
			}
		}()

		j := j.([]interface{})
		ps = make([]*Profile, 0, len(j))

		for _, p := range j {
			var pr *Profile
			pr, err = buildProfile(p.(map[string]interface{}), ErrNoSuchProfile)
			if err != nil {
				if err == ErrNoSuchProfile {
					// Skip demo accounts
					continue
				}
				return nil, err
			}
			ps = append(ps, pr)
		}
		return ps, nil
	} else {
		return nil, transformError(err)
	}
}

func transformError(src error) error {
	if e, ok := internal.UnwrapFailedRequestError(src); ok {
		if e.StatusCode == 204 {
			return ErrNoSuchProfile
		} else if e.ErrorCode == "TooManyRequestsException" {
			return ErrTooManyRequests
		}
	}
	return src
}

var emptyHist = make([]PastName, 0, 0)

var client = &http.Client{}

// buildProfile makes and fills out the basics of a Profile.
// m MUST contain string values for the keys "id" and "name".
// If available, "demo" and "legacy" MUST map to boolean values.
// demoErr is the error to return if the profile is a demo account.
// If an error is returned, p will be nil.
func buildProfile(m map[string]interface{}, demoErr error) (p *Profile, err error) {
	// Ensure demo accounts are not returned
	if t, demo := m["demo"]; demo && t.(bool) {
		return nil, demoErr
	}

	p = &Profile{
		ID:   m["id"].(string),
		Name: m["name"].(string),
	}

	// Legacy Minecraft accounts have not migrated to Mojang accounts.
	// To change your Minecraft username you need to have a Mojang account.
	// Hence "legacy" flags a profile as having no name history.
	if t, legacy := m["legacy"]; legacy && t.(bool) {
		p.NameHistory = emptyHist
	}

	return p, nil
}

// buildHistory creates a username history (previous username first, original
// username last) and returns it along with the current username.
// a is an array of maps containing "name" and (possibly) "changedToAt" keys.
// The "name" values MUST be string and the "changedToAt" values MUST be integer.
// A "changedToAt" field is the "until" field of the previous PastName struct.
func buildHistory(arr []interface{}) (name string, hist []PastName) {
	if len(arr) == 0 {
		return
	}

	hist = make([]PastName, len(arr)-1)

	h := len(hist) - 1
	for i, v := range arr {
		m := v.(map[string]interface{})

		if v, ok := m["changedToAt"]; ok && i > 0 {
			hist[h+1].Until = msToTime(int64(v.(float64)))
		}

		if i == len(hist) {
			name = m["name"].(string)
			break
		} else {
			hist[h].Name = m["name"].(string)
			h--
		}
	}

	return
}

func msToTime(ms int64) time.Time {
	s := ms / 1000
	ns := (ms - s*1000) * 1000000
	return time.Unix(s, ns)
}

// buildProperties returns a property set based on a JSON array of properties.
// props MUST consist of map[string]interface{} maps, each map containing
// string values for the keys "name" and "value".
func buildProperties(props []interface{}) (ps *Properties, err error) {
	ps = new(Properties)
	for _, p := range props {
		prop := p.(map[string]interface{})
		name := prop["name"].(string)
		value := prop["value"].(string) // base64 encoded

		if parser, ok := propertyPopulators[name]; ok {
			err = parser(value, ps)
			if err != nil {
				return nil, err
			}
		}
	}
	return ps, nil
}

// propertyPopulators is a map of property name/value parser pairs.
// Each parser takes the base64 encoded value, decodes it, and populates p with
// the parsed data.
var propertyPopulators = map[string]func(base64 string, p *Properties) error{
	"textures": populateTextures,
}

// populateTextures parses the base64 encoded "textures" property enc and adds
// its information to the Properties struct.
func populateTextures(enc string, props *Properties) error {
	bs, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return err
	}

	var j map[string]interface{}
	err = json.NewDecoder(bytes.NewBuffer(bs)).Decode(&j)
	if err != nil {
		return err
	}

	ts := j["textures"].(map[string]interface{})

	// Set skin URL and skin Model if present
	if s, set := ts["SKIN"]; set {
		skin := s.(map[string]interface{})
		props.SkinURL = skin["url"].(string)

		props.Model = Steve // Steve unless explicitly overridden
		if s, set := skin["metadata"]; set {
			skinMeta := s.(map[string]interface{})
			if m, set := skinMeta["model"]; set && m.(string) == "slim" {
				props.Model = Alex
			}
		}
	} else {
		// Default skin and model depends on player ID
		props.Model = defaultModel(j["profileId"].(string))
	}

	// Set cape URL
	if c, ok := ts["CAPE"]; ok {
		cape := c.(map[string]interface{})
		props.CapeURL = cape["url"].(string)
	}

	return nil
}

// defaultModel implementation is inspired by https://git.io/vSF4a.
// Credit goes to Minecrell for compacting Java's 'uuid.hashCode() & 1' into the below.
//
// Copyright (c) 2014, Lapis <https://github.com/LapisBlue>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
func defaultModel(uuid string) Model {
	if (isEven(uuid[7]) != isEven(uuid[16+7])) != (isEven(uuid[15]) != isEven(uuid[16+15])) {
		return Alex
	} else {
		return Steve
	}
}

func isEven(c uint8) bool {
	switch {
	case c >= '0' && c <= '9':
		return (c & 1) == 0
	case c >= 'a' && c <= 'f':
		return (c & 1) == 1
	default:
		panic("minecraft/profile: invalid digit '" + string(c) + "' in player uuid")
	}
}
