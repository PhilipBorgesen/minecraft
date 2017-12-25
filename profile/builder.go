package profile

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"time"
)

var emptyHist = make([]PastName, 0, 0)

// fillProfile fills out p with basic profile information from m.
// m MUST contain string values for the keys "id" and "name".
// If available, "demo" and "legacy" MUST map to boolean values.
// fillProfile returns false if m represents a demo profile, otherwise
// true. If fillProfile returns false, p will not have been modified.
func fillProfile(p *Profile, m map[string]interface{}) bool {
	// Ensure demo accounts are not returned
	if t, demo := m["demo"]; demo && t.(bool) {
		return false
	}

	id := m["id"].(string)
	name := m["name"].(string)

	if p.NameHistory == nil {
		// Legacy Minecraft accounts have not migrated to Mojang accounts.
		// To change your Minecraft username you need to have a Mojang account.
		// Hence "legacy" flags a profile as having no name history.
		if t, legacy := m["legacy"]; legacy && t.(bool) {
			p.NameHistory = emptyHist
		}
	}

	p.ID = id
	p.Name = name

	return true
}

// buildHistory creates a username history (previous username first, original
// username last) and returns it along with the current username.
// a is an array of maps containing "name" and (possibly) "changedToAt" keys.
// The "name" values MUST be string and the "changedToAt" values MUST be integer.
// A "changedToAt" field is the "until" field of the previous PastName struct.
func buildHistory(arr []interface{}) (name string, hist []PastName) {
	if len(arr) == 0 {
		return "", nil
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

	return name, hist
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
	ps = &Properties{}
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
