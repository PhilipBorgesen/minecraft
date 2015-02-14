package profile

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
)

/********
* MODEL *
********/

// Model represents the player model type used by a profile.
type Model byte

const (
	Steve Model = iota // Classic player model aka "Steve"
	Alex               // Slim-armed player model aka "Alex"
)

var modelNames = [...]string{
	"Steve",
	"Alex",
}

func (m Model) String() string {

	return modelNames[m]
}

/*********************
* PROFILE PROPERTIES *
*********************/

// Properties contains additional information associated with a Profile.
type Properties struct {
	skinURL string
	capeURL string
	model   Model
}

// SkinURL returns a URL to the profile's custom skin texture if such one has been set.
// If a skin has been set, the boolean return value will be true.
// If no skin texture has been set for the profile, this method returns ("", false).
func (pp *Properties) SkinURL() (url string, ok bool) {

	url = pp.skinURL
	ok = url != ""
	return
}

// CapeURL returns a URL to the profile's cape texture if such one is associated.
// If a cape is associated, the boolean return value will be true.
// If no cape texture is associated with the profile, this method returns ("", false).
func (pp *Properties) CapeURL() (url string, ok bool) {

	url = pp.capeURL
	ok = url != ""
	return
}

// Model returns the player model type used by the profile.
func (pp *Properties) Model() Model {

	return pp.model
}

// SkinReader is a convenience method for retrieving a profile's custom skin texture.
// If no skin texture can be fetched from the URL returned by SkinURL, e.g. because
// no custom skin has been set for the profile, this method returns a ErrNoSkin error.
// If an error occurs, nil is returned instead of a ReadCloser.
//
// It is the client's responsibility to close the ReadCloser.
func (pp *Properties) SkinReader() (io.ReadCloser, error) {

	return readTexture(pp.skinURL, errNoSkin)
}

// CapeReader is a convenience method for retrieving a profile's cape texture.
// If no cape texture can be fetched from the URL returned by CapeURL, e.g. because
// the profile has no cape associated, this method returns a ErrNoCape error.
// If an error occurs, nil is returned instead of a ReadCloser.
//
// It is the client's responsibility to close the ReadCloser.
func (pp *Properties) CapeReader() (io.ReadCloser, error) {

	return readTexture(pp.capeURL, errNoCape)
}

/******************
* PROFILE METHODS *
******************/

// LoadProperties loads and returns the profile's skin, cape and model information.
// On success, the result will be memoized in a thread safe manner to be returned on future calls to this method.
// If an error occurs, nil is returned instead and the result is not memoized.
//
// A profile which was loaded by LoadWithProperties has this information preloaded.
//
// If the profile was loaded using a Store, if not already loaded that store will be used to load the properties.
//
// NB! For each profile, profile properties may only be requested once per minute.
func (p *Profile) LoadProperties() (*Properties, error) {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.properties == nil {

		var loaded *Profile
		var err error

		if c := p.cache; c != nil {

			loaded, err = Store{c}.LoadWithProperties(p.id)

		} else {

			loaded, err = LoadWithProperties(p.id)
		}

		if err != nil {

			return nil, err
		}

		p.properties = loaded.properties
	}

	return p.properties, nil
}

// Properties returns the profile's skin, cape and model information if loaded, otherwise nil.
func (p *Profile) Properties() *Properties {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.properties
}

/************
* INTERNALS *
************/

// Perform a GET request to URL, retrieve a ReadCloser to the returned result.
// If url is the empty string or the url creates a 404 response when requested,
// this function returns notFoundErr
func readTexture(url string, notFoundErr error) (_ io.ReadCloser, err error) {

	if url == "" {

		return nil, notFoundErr
	}

	// Fetch image
	resp, err := http.Get(url)
	if err != nil {

		if resp != nil && resp.StatusCode == 404 {

			return nil, notFoundErr
		}

		return nil, err
	}

	return resp.Body, nil
}

// Map of property => parser pairs
var propertyPopulators = map[string]func(base64 string, p *Properties) error{
	"textures": populateTextures,
}

// Constructs and returns a property set based on a JSON array of properties
func buildProperties(ps []interface{}) (*Properties, error) {

	pp := new(Properties)

	for _, p := range ps {

		prop := p.(map[string]interface{})
		name := prop["name"].(string)
		value := prop["value"].(string) // base64 encoded

		err := propertyPopulators[name](value, pp)
		if err != nil {

			return nil, err
		}
	}

	return pp, nil
}

// Parses the "textures" property and adds its info to the Properties struct
func populateTextures(enc string, pp *Properties) error {

	bs, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {

		return err
	}

	var j interface{}
	err = json.NewDecoder(bytes.NewBuffer(bs)).Decode(&j)
	if err != nil {

		return err
	}

	ts := j.(map[string]interface{})["textures"].(map[string]interface{})

	// Set skin URL and skin Model if present
	if s, set := ts["SKIN"]; set {

		skin := s.(map[string]interface{})

		// Set skin URL
		pp.skinURL = skin["url"].(string)

		// Set model
		if s, set := skin["metadata"]; set {

			skinMeta := s.(map[string]interface{})

			if m, set := skinMeta["model"]; set && m.(string) == "slim" {

				// Default model is Steve
				pp.model = Alex
			}
		}
	}

	// Set cape URL
	if c, ok := ts["CAPE"]; ok {

		cape := c.(map[string]interface{})

		pp.capeURL = cape["url"].(string)
	}

	return nil
}
