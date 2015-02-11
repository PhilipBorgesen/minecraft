package profile_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/PhilipRasmussen/minecraft/profile"
)

/************
* TEST DATA *
************/

// 1. known account of author
const (
	nergName = "Nergalic"
	nergUUID = "087cc153c3434ff7ac497de1569affa1"
	// Model is expected to be Steve
	// Cape is not expected
)

var nergHist = []pastName{{"GeneralSezuan", time.Unix(1423047705, 0)}}

// 2. known account of author
const (
	breeName = "BreeSakana"
	breeUUID = "d9a5b542ce88442aaab38ec13e6c7773"
)

var breeHist = []pastName{}

// akronman1, the holder of the 1.000.000th-Minecraft-copy cape
// Profile may in theory be deleted at any moment...
//
// TODO: Substitute for a cape profile under author's control
const capeUUID = "d90b68bc81724329a047f1186dcd4336"

/******************
* TEST STRUCTURES *
******************/

// Username --> profile info
// Must maintain consistency with the other test structures
var loadTestUsers = map[string]user{
	nergName:                  {nergUUID, nergName},
	strings.ToUpper(nergName): {nergUUID, nergName}, // Casing should not matter
	breeName:                  {breeUUID, breeName},
}

// Should match loadTestUsers without duplicates
var loadManyTestRes = []string{breeName, nergName}

// Keys should match entries in loadTestUsers
var pastNames = map[string][]pastName{
	breeName: breeHist,
	nergName: nergHist,

	// TODO: Add profiles with longer name history to verify order
}

/*************
* TEST TYPES *
*************/

type user struct {
	uuid string
	name string
}

type pastName struct {
	name  string
	until time.Time
}

type propertySet struct {

	// Do not check their values, simply check their presence
	skinURL, capeURL *bool // expected skin and cape (if any)

	model *Model // expected model (if any)

	user string // expect profile to match loadTestUsers[user]
}

/**********
* HELPERS *
**********/

var oneSec time.Duration = time.Unix(1, 0).Sub(time.Unix(0, 0))

// Verifies the basics of a loaded profile, i.e. that no errors happened and that
// the UUID and name attributes are as expected.
func assertBasicInfo(t *testing.T, fn string, err error, got *Profile, expect user) {

	checkForErr(t, err, fn)

	if uuid := got.UUID(); uuid != expect.uuid {

		t.Errorf("%s.UUID() = %q; want %q", fn, uuid, expect.uuid)
	}
	if name := got.Name(); name != expect.name {

		t.Errorf("%s.Name() = %q; want %q", fn, name, expect.name)
	}
}

// Verifies that no error occurred, otherwise reports it.
// err is the error, fn is the function invocation which returned the error.
func checkForErr(t *testing.T, err error, fn string) {

	if err != nil {

		t.Logf("%s returned error: %s", fn, err)

		if _, tmr := err.(ErrTooManyRequests); tmr {

			t.Log("Likely problem: Test has been executed too frequently. Wait 1-10 minutes before retrying.")
		}

		t.FailNow()
	}
}

/*************
* TEST CASES *
*************/

// Test that the correct UUID and case-corrected name gets loaded
func TestLoad(t *testing.T) {

	for n, expect := range loadTestUsers {

		got, err := Load(n)
		assertBasicInfo(t, fmt.Sprintf("Load(%q)", n), err, got, expect)
	}
}

// Test that the correct UUID and case-corrected name gets loaded for past names
func TestLoadAtTime(t *testing.T) {

	for n, hist := range pastNames {

		expect := loadTestUsers[n]

		for _, p := range hist {

			tmfmt := p.until.Format(time.RFC3339)

			got, err := LoadAtTime(p.name, p.until)
			assertBasicInfo(t, fmt.Sprintf("LoadAtTime(%q, %s)", p.name, tmfmt), err, got, expect)
		}
	}
}

// No profile should be found for these usernames at these time instants.
func TestLoadAtTimeFailure(t *testing.T) {

	for n, hist := range pastNames {

		for _, p := range hist {

			p.until = p.until.Add(oneSec)
			tmfmt := p.until.Format(time.RFC3339)

			got, err := LoadAtTime(p.name, p.until)

			if _, nsp := err.(ErrNoSuchUser); !nsp {

				checkForErr(t, err, fmt.Sprintf("LoadAtTime(%q, %s)", p.name, tmfmt))

				t.Errorf("LoadAtTime(%q, %s) = %s; want ErrNoSuchUser error", n, tmfmt, got)
			}
		}
	}
}

// Verify that name histories are reported with expected stats and in ascending order.
// Also verify that LoadWithNameHistory actually preloads the profile's name history
//
// TODO: Supply test samples to verify ascending order
func TestLoadWithNameHistory(t *testing.T) {

	for n, hist := range pastNames {

		expect := loadTestUsers[n]
		fn := fmt.Sprintf("LoadWithNameHistory(%q)", expect.uuid)

		got, err := LoadWithNameHistory(expect.uuid)

		if got.NameHistory() == nil {

			t.Errorf("%s.NameHistory() = nil; should already be loaded", fn)
		}

		assertBasicInfo(t, fn, err, got, expect)

		// Test name history for correctness and ascending order
		h, err := got.LoadNameHistory()
		
		// Never ought to happen
		checkForErr(t, err, fn)

		var nameMatch, timeMatch bool
		for i, p := range hist {

			nameMatch = h[i].Name() == p.name
			if !nameMatch {

				t.Errorf("%s.LoadNameHistory()[%d].Name() = %q; want %q", fn, i, h[i].Name(), p.name)
				continue
			}

			timeMatch = h[i].Until() == p.until
			if !timeMatch {

				t.Errorf("%s.LoadNameHistory()[%d].Until() = %s; want %s",
					fn, i, h[i].Until().Format(time.RFC3339), p.until.Format(time.RFC3339))
			}
		}
	}
}

// Verify that:
// 1) LoadWithProperties successfully loads a profile with correct name and UUID
// 2) LoadWithProperties preloads properties
// 3) Skins are handled successfully, a) present or b) not
// 4) Model are handled successfully, a) present or b) not (Alex/Steve)
// 5) Capes are handled successfully, b) present or b) not
//
// TODO: 3b, 4a
func TestLoadWithProperties(t *testing.T) {

    ///////////////////////////
	// TEST 1, 2, 3a, 4b, 5b //
    ///////////////////////////

	u1 := loadTestUsers[nergName]
	
	fn := fmt.Sprintf("LoadWithProperties(%q)", u1.uuid)
	p1, err := LoadWithProperties(u1.uuid)
	
	// 1) Correct name and UUID, no errors
	assertBasicInfo(t, fn, err, p1, u1)
	
	// 2) Properties preloaded
	if p1.Properties() == nil {
	
		t.Errorf("%s.Properties() = nil; want preloaded", fn)
	}
	
	pp1, err := p1.LoadProperties()
	
	// Never ought to happen
	checkForErr(t, err, fn)
	
	// 3a) Skin set
	if _, ok := pp1.SkinURL(); !ok {
	
		t.Errorf("%s.Properties().SkinURL() = \"\"; want URL", fn)
	}
	
	// 4b) Model is Steve
	if m := pp1.Model(); m != Steve {
	
		t.Errorf("%s.Properties().Model() = %s; want %s", fn, m, Steve)
	}
	
	// 5b) No cape
	if c, ok := pp1.CapeURL(); ok {
	
		t.Errorf("%s.Properties().CapeURL() = %q; want %q", fn, c, "")
	}
	
	////////////////
	// TEST 2, 5a //
	////////////////

	fn = fmt.Sprintf("LoadWithProperties(%q)", capeUUID)
	
	p2, err := LoadWithProperties(capeUUID)
	checkForErr(t, err, fn)
	
	// 2 + 5a) Check for cape and blow up if not preloaded
	if p2.Properties().CapeURL == nil {
	
		t.Errorf("%s.Properties().CapeURL() = \"\"; want URL", fn)
	}
}

// Test that ErrTooManyRequests is returned if we exceed the rate limit.
// Allow the operation to succeed in case Mojang changes the rate limit.
func TestLoadWithPropertiesFailure(t *testing.T) {

	uuid := loadTestUsers[breeName].uuid

	for i := 0; i < 3; i++ {

		_, err := LoadWithProperties(uuid)
		if _, nsp := err.(ErrTooManyRequests); !nsp && err != nil {

			t.Fatalf("LoadWithProperties(%q) returned non-ErrTooManyRequests error: %s", uuid, err)
		}
	}
}

// Test that multiple users may be loaded (correctly) at once.
// Also test that empty and non-existing usernames are ignored
// and duplicates only are returned once.
func TestLoadMany(t *testing.T) {

	testUsers := []string{"", "I_DONT_ËXIST_ÆØÅ39"}

	for n, _ := range loadTestUsers {

		testUsers = append(testUsers, n)
	}

	fn := fmt.Sprintf("LoadMany(%#v)", testUsers)

	ps, err := LoadMany(testUsers...)
	checkForErr(t, err, fn)

	expect := loadManyTestRes

	// Verify that the correct number of profiles were returned
	if len(ps) != len(expect) {

		t.Errorf("len(%s) = %d; want %d", fn, len(ps), len(expect))
		t.Errorf("%s = %s; want %s", fn, ps, expect)

		t.FailNow()
	}

	// Verify that the loaded data was correct.
	for i := range ps {

		got := ps[i]
		assertBasicInfo(t, fn, nil, got, loadTestUsers[expect[i]])
	}
}

// Test that LoadMany actually succeeds at requesting LoadManyMaxSize profiles.
func TestLoadManyMax(t *testing.T) {

	var genUsers []string
	for i := 0; i < LoadManyMaxSize; i++ {

		genUsers = append(genUsers, "user"+string(i))
	}

	_, err := LoadMany(genUsers...)

	// Verify that no error occurred
	checkForErr(t, err, "LoadMany(<LoadManyMaxSize USERNAMES>)")
}
