package profile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

/*********************
* EXPORTED CONSTANTS *
*********************/

// The maximum number of profiles which may be requested at once using the LoadMany function.
// If more are requested, the request may fail with a ErrMaxSizeExceeded error.
const LoadManyMaxSize int = 100

/*********************
* EXPORTED FUNCTIONS *
*********************/

// Load fetches the profile currently associated with a username.
func Load(username string) (*Profile, error) {

	url := fmt.Sprintf(loadURL, username)
	return load(url, username)
}

// LoadAtTime fetches the profile associated with a username at the specified instant of time.
func LoadAtTime(username string, tm time.Time) (*Profile, error) {

	url := fmt.Sprintf(loadAtTimeURL, username, tm.Unix())
	return load(url, username)
}

// LoadByID fetches the profile identified by an ID.
func LoadByID(id string) (*Profile, error) {

	return LoadWithNameHistory(id)
}

// LoadNameHistory fetches the profile identified by an ID.
// As a bonus, profiles loaded by this function already have their name history preloaded.
func LoadWithNameHistory(id string) (*Profile, error) {

	var nsu error = ErrNoSuchID{id}

	if id == "" {

		return nil, nsu
	}

	url := fmt.Sprintf(loadWithNameHistoryURL, id)
	j, err := getJson(url, nsu)
	if err != nil {

		return nil, err
	}

	a := j.([]interface{})
	name, hist := createHistory(a)

	p := &Profile{
		id:      id,
		name:    name,
		history: hist,
	}

	return p, nil
}

// LoadWithProperties fetches the profile identified by a ID.
// As a bonus, profiles loaded by this function already have skin, cape and model information preloaded.
//
// NB! For each profile, profile properties may only be requested once per minute.
func LoadWithProperties(id string) (*Profile, error) {

	var nsu error = ErrNoSuchID{id}

	if id == "" {

		return nil, nsu
	}

	url := fmt.Sprintf(loadWithPropertiesURL, id)
	j, err := getJson(url, nsu)
	if err != nil {

		return nil, err
	}

	p, err := buildProfile(j, nsu)
	if err != nil {

		return nil, err
	}

	m := j.(map[string]interface{})
	p.properties, err = buildProperties(m["properties"].([]interface{}))
	if err != nil {

		// Let the entire loading fail even if just property construction failed
		// May always be changed later if this is too drastic
		return nil, err
	}

	return p, nil
}

// LoadMany fetches multiple profiles based on current username/profile mappings.
// Usernames mapping to no profile are ignored and absent from the returned results.
// Duplicate usernames are only returned once.
//
// NB! Only a maximum of LoadManyMaxSize profiles may be fetched at once.
// If more are attempted loaded in the same operation an ErrMaxSizeExceeded error occurs.
func LoadMany(username ...string) ([]*Profile, error) {

	if len(username) > LoadManyMaxSize {

		return nil, ErrMaxSizeExceeded{len(username)}
	}

	// Remove empty usernames. They are not accepted by the Mojang API.
	var users []string
	for _, u := range username {

		if u != "" {

			users = append(users, u)
		}
	}

	// No need to request anything
	if len(users) == 0 {

		return nil, nil
	}

	body, err := json.Marshal(users)
	if err != nil {

		return nil, err
	}

	j, err := postJson(loadManyURL, body, nil)
	if err != nil {

		return nil, err
	}

	var res []*Profile
	for _, p := range j.([]interface{}) {

		pr, err := buildProfile(p, errDemo)
		if err != nil {

			// Skip demo accounts
			if err == errDemo {

				continue
			}

			// Otherwise fail
			return nil, err
		}

		res = append(res, pr)
	}

	return res, nil
}

/************
* INTERNALS *
************/

const (
	loadURL                = "https://api.mojang.com/users/profiles/minecraft/%s"
	loadAtTimeURL          = "https://api.mojang.com/users/profiles/minecraft/%s?at=%d"
	loadWithNameHistoryURL = "https://api.mojang.com/user/profiles/%s/names"
	loadWithPropertiesURL  = "https://sessionserver.mojang.com/session/minecraft/profile/%s"
	loadManyURL            = "https://api.mojang.com/profiles/minecraft"
)

// Common implementation used by Load and LoadAtTime.
func load(url, user string) (*Profile, error) {

	var nsu error = ErrNoSuchUser{user}

	if user == "" {

		return nil, nsu
	}

	j, err := getJson(url, nsu)
	if err != nil {

		return nil, err
	}

	p, err := buildProfile(j, nsu)
	if err != nil {

		return nil, err
	}

	return p, nil
}

// Common implementation used to make and fill out the basics of a Profile.
// j should be a map[string]interface{} with string values for the keys "id" and "name".
// If available, "demo" and "legacy" are expected to map to boolean values.
// demoErr is the error to return if the profile is a demo account.
func buildProfile(j interface{}, demoErr error) (*Profile, error) {

	m := j.(map[string]interface{})

	// Ensure demo accounts are not returned
	if t, demo := m["demo"]; demo && t.(bool) {

		return nil, demoErr
	}

	p := &Profile{
		id:   m["id"].(string),
		name: m["name"].(string),
	}

	// Legacy Minecraft accounts have not migrated to Mojang accounts.
	// To change your Minecraft username you need to have a Mojang account.
	// Hence "legacy" flags a profile as having no name history.
	if t, legacy := m["legacy"]; legacy && t.(bool) {

		p.history = emptyHist
	}

	return p, nil
}

// Common implementation used to make GET requests to the public Mojang API.
// url is what URL to request.
// ncErr is the error to return if a 204 No Content response is received.
// If ncErr is nil, no error is returned on a 204 No Content response.
func getJson(url string, ncErr error) (interface{}, error) {

	// Make request
	resp, err := http.Get(url)
	if err != nil {

		return nil, err
	}
	defer resp.Body.Close()

	return parseJson(resp, ncErr)
}

// Common implementation used to make POST requests to the public Mojang API.
// url is what URL to request.
// ncErr is the error to return if a 204 No Content response is received.
// If ncErr is nil, no error is returned on a 204 No Content response.
func postJson(url string, body []byte, ncErr error) (interface{}, error) {

	// Make request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {

		return nil, err
	}
	defer resp.Body.Close()

	return parseJson(resp, ncErr)
}

// Common implementation for parsing a JSON response from the public Mojang API.
// ncErr is the error to return if a 204 No Content response is received.
// If ncErr is nil, no error is returned on a 204 No Content response.
func parseJson(resp *http.Response, ncErr error) (interface{}, error) {

	// Profile exists?
	if resp.StatusCode == 204 && ncErr != nil {

		return nil, ncErr
	}

	// Decode JSON
	var j interface{}
	err := json.NewDecoder(resp.Body).Decode(&j)
	if err != nil {

		return nil, err
	}

	// Check for JSON errors
	if err := getJsonError(j); err != nil {

		return nil, err
	}

	return j, nil
}
