// This package allows the username, ID, skin and username history of Minecraft profiles to be retrieved by either username or ID.
// It is a binding for the public Mojang API described at: http://wiki.vg/Mojang_API
//
// Since Mojang's API historically have been inconsistent on whether demo profiles are returned or not,
// to ensure consistency this package have been written never to return those.
//
// Please note that the public Mojang API is request rate limited, so if you expect heavy usage you should cache the results.
// For more information see the documentation for ErrTooManyRequests and LoadWithProperties.
package profile

import "sync"

/**********
* PROFILE *
**********/

// Profile represents the profile of a Minecraft user account.
// A Profile struct should not be copied.
type Profile struct {
	id   string
	name string

	history    []PastName
	properties *ProfileProperties
	mutex      sync.Mutex
}

// ID returns the universially unique ID of the profile.
func (p *Profile) ID() string {

	return p.id
}

// Name returns the profile's username.
func (p *Profile) Name() string {

	return p.name
}

// String uses Name as its string representation.
func (p *Profile) String() string {

	return p.name
}
