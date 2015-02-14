// Package versions fetches Mojang's versions listing of Minecraft, allowing clients to determine what the latest version is.
//
// This is useful to determine whether the latest version of the game is installed and/or construct download URLs for the latest versions of the game.
// For example:
//  vs, err := versions.Load()
//  if err != nil {
//
//    panic("Failed to fetch versions listing: " + err.Error())
//  }
//
//  if latest := vs.Latest.Release; latest != currentVersion {
//
//    url := fmt.Sprintf("http://s3.amazonaws.com/Minecraft.Download/versions/%s/%s.jar", latest, latest)
//
//    resp, err := http.Get(url)
//    ...
//  }
// For more information, see: http://wiki.vg/Game_Files
package versions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

/******************
* VERSION LISTING *
******************/

// Load fetches a listing of Minecraft versions from Mojang's servers.
//
// Load does not cache the result, and if an error occurs the returned Listing struct is uninitialised.
func Load() (Listing, error) {

	m, err := fetchJSON(versionsURL)
	if err != nil {

		return Listing{}, err
	}

	var l Listing
	l.Versions = make(map[string]Version)
	buildVersions(m, &l)

	return l, nil
}

// Listing represents a listing of Minecraft versions.
type Listing struct {
	Versions map[string]Version // Every known version of Minecraft indexed by version ID
	Latest   struct {           // The IDs of the latest snapshot and release versions.
		Snapshot string // The version ID of the latest development snapshot
		Release  string // The version ID of the latest Minecraft release
	}
}

// LatestRelease returns the version information for the latest release version.
// This is a convenience method for:
//  l.Versions[l.Latest.Release]
// If l.Versions does not contain the key l.Latest.Release, LatestRelease panics.
func (l Listing) LatestRelease() Version {

	v, ok := l.Versions[l.Latest.Release]

	if !ok {

		panic(fmt.Sprintf("l.Versions does not contain the l.Latest.Release key %q", l.Latest.Release))
	}

	return v
}

/***************
* VERSION TYPE *
***************/

// Type represents the release type of a version.
type Type string

const (
	Release  Type = "release"   // Release version
	Snapshot Type = "snapshot"  // Development snapshot
	Alpha    Type = "old_alpha" // An alpha version
	Beta     Type = "old_beta"  // An beta version
)

// String returns a description of the version type meant for humans:
//  Release.String()  = "release"
//  Snapshot.String() = "snapshot"
//  Alpha.String()    = "alpha"
//  Beta.String()     = "beta"
//
//  Type("").String()  = "???" // Zero value
//  Type("X").String() = "X"   // Unknown version types
func (t Type) String() string {

	switch t {

	case Release:
		return "release"

	case Snapshot:
		return "snapshot"

	case Alpha:
		return "alpha"

	case Beta:
		return "beta"

	case "":
		return "???"

	default:
		return string(t)

	}
}

/**********
* VERSION *
**********/

// Version contains information about a Minecraft version.
type Version struct {
	ID       string    // E.g. "1.8.1"
	Released time.Time // Time the version was released
	Type     Type      // Type of version, e.g. release or snapshot
}

// String returns the version's ID as its string representation.
func (v Version) String() string {

	return v.ID
}

/************
* INTERNALS *
************/

const versionsURL = "http://s3.amazonaws.com/Minecraft.Download/versions/versions.json"
const timeFormat = "2006-01-02T15:04:05-07:00"

// Fetch versions.json from Mojang servers
func fetchJSON(url string) (map[string]interface{}, error) {

	// Fetch JSON
	resp, err := http.Get(url)
	if err != nil {

		return nil, err
	}
	defer resp.Body.Close()

	// Decode JSON
	var j interface{}
	err = json.NewDecoder(resp.Body).Decode(&j)
	if err != nil {

		return nil, err
	}

	return j.(map[string]interface{}), nil
}

// Populate pre-initialised Versions struct with JSON data
func buildVersions(m map[string]interface{}, l *Listing) {

	latest := m["latest"].(map[string]interface{})
	l.Latest.Snapshot = latest["snapshot"].(string)
	l.Latest.Release = latest["release"].(string)

	vm := l.Versions
	for _, v := range m["versions"].([]interface{}) {

		var vers Version
		buildVersion(v.(map[string]interface{}), &vers)

		vm[vers.ID] = vers
	}
}

// Fill out Version struct with JSON data
func buildVersion(m map[string]interface{}, v *Version) {

	v.ID = m["id"].(string)
	v.Released = parseTime(m["releaseTime"].(string))
	v.Type = Type(m["type"].(string))
}

func parseTime(t string) time.Time {

	tm, _ := time.Parse(timeFormat, t)

	return tm
}
