// Package profile allows the username, ID, skin and username history of Minecraft
// profiles to be retrieved by either username or ID.
// It is a binding for the public Mojang API described at: http://wiki.vg/Mojang_API.
//
// Since Mojang's API historically have been inconsistent on whether demo profiles
// are returned or not, to ensure consistency this package have been written never
// to return those.
//
// Please note that the public Mojang API is request rate limited, so if you expect
// heavy usage you should cache the results.
// For more information on rate limits see the documentation for ErrTooManyRequests.
package profile

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/PhilipBorgesen/minecraft/internal"
)

// Profile represents the profile of a Minecraft user account.
type Profile struct {
	// ID is the profile's universally unique identifier, which never changes.
	ID string
	// Name is the profile's currently associated username, subject to change.
	Name string
	// NameHistory is the profile's past usernames incl. when each username
	// stopped being used. The profile's last username is first, its original
	// username is last. Unless explicitly loaded, NameHistory may be nil.
	NameHistory []PastName
	// Properties contains the skin, model and cape used by the profile.
	// Unless explicitly loaded, Properties may be nil.
	Properties *Properties

	_ struct{} // Ensure Profile is constructed using named parameters.
}

// String returns p.Name.
func (p *Profile) String() string {
	return p.Name
}

// LoadNameHistory loads and returns p.NameHistory, which contains the
// profile's past usernames. If force is true, p.NameHistory will be loaded
// anew from the Mojang servers even though it already is present. If force
// is false, p.NameHistory will only be loaded if nil.
//
// ctx must be non-nil and p.ID must be set. When the name history is loaded,
// p.Name will also be updated if it has changed.
//
// No matter whether the loading succeeds or not, p.NameHistory will be
// returned as hist, which thus only will be nil if the loading fails and
// p.NameHistory was nil beforehand.
//
// A profile which was loaded by LoadWithNameHistory has p.NameHistory
// pre-loaded.
func (p *Profile) LoadNameHistory(ctx context.Context, force bool) (hist []PastName, err error) {
	if p.NameHistory == nil || force {
		var loaded *Profile
		loaded, err = LoadWithNameHistory(ctx, p.ID)
		if err != nil {
			if err == ErrNoSuchProfile && p.ID == "" {
				err = ErrUnsetPlayerID
			}
		} else {
			p.Name = loaded.Name
			p.NameHistory = loaded.NameHistory
		}
	}
	return p.NameHistory, err
}

// LoadProperties loads and returns p.Properties, which contains the profile's
// skin, cape and model. If force is true, p.Properties will be loaded anew
// from the Mojang servers even though it already is present. If force is
// false, p.Properties will only be loaded if nil.
//
// ctx must be non-nil and p.ID must be set. When properties are loaded, p.Name
// will also be updated if it has changed.
//
// No matter whether the loading succeeds or not, p.Properties will be returned
// as ps, which thus only will be nil if the loading fails and p.Properties was
// nil beforehand.
//
// A profile which was loaded by LoadWithProperties has p.Properties pre-loaded.
//
// NB! For each profile, profile properties may only be requested once per minute.
func (p *Profile) LoadProperties(ctx context.Context, force bool) (ps *Properties, err error) {
	if p.Properties == nil || force {
		var loaded *Profile
		loaded, err = LoadWithProperties(ctx, p.ID)
		if err != nil {
			if err == ErrNoSuchProfile && p.ID == "" {
				err = ErrUnsetPlayerID
			}
		} else {
			p.Name = loaded.Name
			p.Properties = loaded.Properties
		}
	}
	return p.Properties, err
}

/*// UploadSkin sets s as the skin for the profile identified by p.ID.
// authToken is a valid Mojang authentication token that can be retrieved
// using the minecraft/auth package. ctx must be non-nil.
func (p *Profile) UploadSkin(ctx context.Context, authToken string, s *SkinUpload) error {
	return nil
}*/

// PastName represents one of a profile's past usernames.
// PastName values should be used as map or database keys with caution as they
// contain a time.Time field. For the same reasons, do not use == with PastName
// values; use Equal instead.
type PastName struct {
	// Name is a username used by the profile in the past.
	Name string
	// Until is the time instant the profile stopped using Name as username.
	// Prior past usernames may be consulted to determine when this username
	// was taken into use.
	Until time.Time

	_ struct{} // Ensure PastName is constructed using named parameters.
}

// Equal reports whether p and q represents the same past username of a
// profile, i.e. whether p.Name == q.Name and p and q were used until the same
// time instant. Do not use == with PastName values.
func (p PastName) Equal(q PastName) bool {
	return p.Name == q.Name && p.Until.Equal(q.Until)
}

// String returns p.Name.
func (p PastName) String() string {
	return p.Name
}

// Properties contains additional information associated with a Profile.
type Properties struct {
	// SkinURL is an URL to the profile's custom skin texture.
	// If SkinURL == "", no skin texture has been set and the profile uses the
	// default skin for Model.
	SkinURL string
	// CapeURL is an URL to the profile's cape texture.
	// If CapeURL == "", no cape is associated with the profile.
	CapeURL string
	// Model is the profile's player model type.
	Model Model

	_ struct{} // Ensure Properties is constructed using named parameters.
}

// SkinReader is a convenience method for retrieving the skin texture at
// p.SkinURL. ctx must be non-nil. If p.SkinURL == "", the default texture for
// p.Model will be attempted to be retrieved instead.
//
// It is the client's responsibility to close the ReadCloser. When an error is
// returned, ReadCloser is nil.
func (p *Properties) SkinReader(ctx context.Context) (io.ReadCloser, error) {
	url := p.SkinURL
	if url == "" {
		url = p.Model.defaultSkinURL()
		if url == "" {
			return nil, ErrUnknownModel
		}
	}
	return loadTexture(ctx, url)
}

// CapeReader is a convenience method for retrieving the cape texture at
// p.CapeURL. ctx must be non-nil. If p.CapeURL == "", ErrNoCape is returned as
// error.
//
// It is the client's responsibility to close the ReadCloser. When an error is
// returned, ReadCloser is nil.
func (p *Properties) CapeReader(ctx context.Context) (io.ReadCloser, error) {
	if p.CapeURL == "" {
		return nil, ErrNoCape
	}
	return loadTexture(ctx, p.CapeURL)
}

func loadTexture(ctx context.Context, endpoint string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		err = &url.Error{
			Op:  "Get",
			URL: endpoint,
			Err: &internal.FailedRequestError{StatusCode: resp.StatusCode},
		}
		return nil, err
	}

	return resp.Body, nil
}

// Model represents the player model type used by a profile.
type Model byte

const (
	Steve Model = iota // Classic player model aka "Steve".
	Alex               // Slim-armed player model aka "Alex".
)

// String returns a string representation of m.
//	Steve.String() = "Steve"
//	Alex.String()  = "Alex"
// String returns "???" for models not declared by this package.
func (m Model) String() string {
	switch m {
	case Steve:
		return "Steve"
	case Alex:
		return "Alex"
	default:
		return "???"
	}
}

// defaultSkinURL returns a URL to the default skin of m
func (m Model) defaultSkinURL() string {
	switch m {
	case Steve:
		return steveSkinURL
	case Alex:
		return alexSkinURL
	default:
		return ""
	}
}
