package profile

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

/********
* MODEL *
********/

// Model represents the player model type used by a profile.
type Model int

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

// ProfileProperties contains additional information associated with a Profile.
type ProfileProperties struct {
	skinURL string
	capeURL string
	model   Model
}

// SkinURL returns a URL to the profile's custom skin texture if such one has been set.
// If a skin has been set, the boolean return value will be true.
// If no skin texture has been set for the profile, this method returns ("", false).
func (pp *ProfileProperties) SkinURL() (url string, ok bool) {

	url = pp.skinURL
	ok = url != ""
	return
}

// CapeURL returns a URL to the profile's cape texture if such one is associated.
// If a cape is associated, the boolean return value will be true.
// If no cape texture is associated with the profile, this method returns ("", false).
func (pp *ProfileProperties) CapeURL() (url string, ok bool) {

	url = pp.capeURL
	ok = url != ""
	return
}

// Model returns the player model type used by the profile.
func (pp *ProfileProperties) Model() Model {

	return pp.model
}

/******************
* PROFILE METHODS *
******************/

// LoadProperties loads and returns profile skin, cape and model information.
// On success, the result will be memoized in a thread safe manner to be returned on future calls to this method.
// A profile which was loaded by LoadWithProperties has this information preloaded.
//
// NB! For each profile, profile properties may only be requested once per minute.
func (p *Profile) LoadProperties() (*ProfileProperties, error) {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.properties == nil {

		loaded, err := LoadWithProperties(p.id)
		if err != nil {

			return nil, err
		}

		p.properties = loaded.properties
	}

	return p.properties, nil
}

// Properties simply returns profile skin, cape and model information if loaded, otherwise nil.
func (p *Profile) Properties() *ProfileProperties {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.properties
}

/************
* INTERNALS *
************/

// Map of property => parser pairs
var propertyPopulators = map[string]func(base64 string, p *ProfileProperties) error{
	"textures": populateTextures,
}

// Constructs and returns a property set based on a JSON array of properties
func buildProperties(ps []interface{}) (*ProfileProperties, error) {

	pp := new(ProfileProperties)

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

// Parses the "textures" property and adds its info to the ProfileProperties struct
func populateTextures(enc string, pp *ProfileProperties) error {

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
