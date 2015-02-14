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

// FAKE PROFILES FOR READING

var (
	fakeID   = "23!€*-`"
	fakeName = "I_DONT_ËXIST_ÆØÅ39"

	// Assigned nil to enable LoadNameHistory and LoadProperties tests
	fakeHistory    []PastName  = nil
	fakeProperties *Properties = nil

	fakePastName = "I_DONT_ËXIST_ÆØÅ39_EITHER"
	fakePastTime = time.Unix(42, 0)
)

var fakeCacheEntry = CacheEntry{
	ID:          fakeID,
	Name:        fakeName,
	NameHistory: fakeHistory,
	Properties:  fakeProperties,
}

// REAL PROFILES FOR WRITING
// nerg* identifiers are from "load_test.go"

var realID = nergID
var realName = nergName

var (
	realPastName = nergHist[0].name
	realPastTime = nergHist[0].until
)

// These IDs could in theory stop being valid if their accounts are deleted.
// TODO: Substitute for profiles under author's control
var realID2 = "1796eb3bfc0346cda5fcdd139a2d87d2" // Forfal
var realID3 = "f8e273cca7c4499080327e15de919b8c" // Dushmursts

/*************
* TEST CACHE *
*************/

type cacheLogEntry struct {
	Name, ID string
	Time     time.Time
}

// A loggingTestCache caches values passed to it and logs every attempted read from it.
// An entry already existing is simply replaced by new values rather than merged.
type loggingTestCache struct {
	CacheReads []cacheLogEntry

	EntriesByID      map[string]CacheEntry           // ID --> profile info
	IDsByName        map[string]string               // Name --> ID
	IDsByNameAndTime map[string]map[time.Time]string // Past name + time --> ID
}

// Make a new LoggingTestCache
func newLoggingTestCache() *loggingTestCache {

	return &loggingTestCache{
		EntriesByID:      make(map[string]CacheEntry),
		IDsByName:        make(map[string]string),
		IDsByNameAndTime: make(map[string]map[time.Time]string),
	}
}

// METHODS -- CACHING

func (ts *loggingTestCache) CacheNameAtTime(name string, tm time.Time, id string) {

	// Normalise name
	norm_name := norm(name)

	// Create map of timestamps for name if no prior LoadAtTime queries have been done for name
	_, ok := ts.IDsByNameAndTime[norm_name]
	if !ok {

		ts.IDsByNameAndTime[norm_name] = make(map[time.Time]string)
	}

	// Store mapping
	ts.IDsByNameAndTime[norm_name][tm] = id
}

func (ts *loggingTestCache) Cache(e CacheEntry) {

	ts.EntriesByID[e.ID] = e
	ts.IDsByName[norm(e.Name)] = e.ID
}

// METHODS -- LOOKUP

func (ts *loggingTestCache) GetName(name string) (entry CacheEntry, ok bool) {

	// Log that the profile of name was sought
	ts.CacheReads = append(ts.CacheReads, cacheLogEntry{Name: name})

	// Lookup name --> ID
	id, ok := ts.IDsByName[norm(name)]
	if !ok {

		return
	}

	// Lookup ID --> profile info
	entry, ok = ts.EntriesByID[id]
	return
}

func (ts *loggingTestCache) GetNameAtTime(name string, tm time.Time) (entry CacheEntry, ok bool) {

	// Log that the profile of name at time was sought
	ts.CacheReads = append(ts.CacheReads, cacheLogEntry{Name: name, Time: tm})

	// Lookup past name --> time map
	m, ok := ts.IDsByNameAndTime[norm(name)]
	if !ok {

		return
	}

	// Lookup past time --> ID
	id, ok := m[tm]
	if !ok {

		return
	}

	// Lookup ID --> profile info
	entry, ok = ts.EntriesByID[id]
	return
}

func (ts *loggingTestCache) GetID(id string) (entry CacheEntry, ok bool) {

	// Log that the profile of id was sought
	ts.CacheReads = append(ts.CacheReads, cacheLogEntry{ID: id})

	// Lookup ID --> profile info
	entry, ok = ts.EntriesByID[id]
	return
}

/**********
* HELPERS *
**********/

// Normalises a string for case insensitive matching.
func norm(s string) string {

	return strings.ToLower(s)
}

// Verifies that:
// - Cache was seeked exactly once
// - Expected cache method was used
func verifyCacheRead(t *testing.T, c *loggingTestCache, fn string, expectedLog cacheLogEntry) {

	// Verify number of cache seeks
	if len(c.CacheReads) != 1 {

		t.Errorf("%s seeked the cache %d times; it should seek exactly once", fn, len(c.CacheReads))

		// Verify correct cache method used
	} else if c.CacheReads[0] != expectedLog {

		t.Errorf("%s did not seek the cache (no matching log entry found)", fn)
	}
}

// Verifies that:
// - No error occurred
// - Cache was seeked exactly once
// - Expected cache method was used
func verifyCacheWrite(t *testing.T, c *loggingTestCache, fn string, expectedCached cacheLogEntry, err error) {

	// Check for errors
	if err != nil {

		t.Errorf("%s returned error: %s", fn, err)
		return
	}

	id := expectedCached.ID
	name := expectedCached.Name
	tm := expectedCached.Time

	switch {

	// Verify profile was cached for ID
	case id != "":
		if _, ok := c.GetID(id); !ok {

			t.Errorf("%s did not cache its results.", fn)
		}

	// Verify profile was cached for past username
	case !tm.IsZero() && name != "":
		if _, ok := c.GetNameAtTime(name, tm); !ok {

			t.Errorf("%s did not cache its results.", fn)
		}

	// Verify profile was checked for current username
	case name != "":
		if _, ok := c.GetName(name); !ok {

			t.Errorf("%s did not cache its results.", fn)
		}

	default:
		panic("Expected cacheLogEntry struct had neither Name or ID set.")
	}
}

/*************
* TEST CASES *
*************/

// READ TEST:
// Methods attempt to read fake profiles from a pre-populated cache.
// All methods on Store are tested along with LoadProperties and
// LoadNameHistory on loaded profiles.
// LoadProperties, LoadNameHistory, LoadWithProperties and LoadWithNameHistory
// will all fail to load with an error since the cache has no properties/history
// info and no profiles exists for the fake ID when the missing info afterwards
// is attempted fetched.
// After each load the cache log is inspected to verify that it was sought.
func TestCacheRead(t *testing.T) {

	// SETUP
	c := newLoggingTestCache()
	c.Cache(fakeCacheEntry)
	c.CacheNameAtTime(fakePastName, fakePastTime, fakeID)

	ps := NewStore(c)

	// TESTS Load method
	ps.Load(fakeName)
	fn := fmt.Sprintf("ps.Load(%q)", fakeName)
	verifyCacheRead(t, c, fn, cacheLogEntry{Name: fakeName})
	c.CacheReads = nil

	// TESTS LoadAtTime method
	ps.LoadAtTime(fakePastName, fakePastTime)
	fn = fmt.Sprintf("ps.LoadAtTime(%q, %s)", fakePastName, fakePastTime)
	verifyCacheRead(t, c, fn, cacheLogEntry{Name: fakePastName, Time: fakePastTime})
	c.CacheReads = nil

	// TESTS LoadWithNameHistory method
	ps.LoadWithNameHistory(fakeID)
	fn = fmt.Sprintf("ps.LoadWithNameHistory(%q)", fakeID)
	verifyCacheRead(t, c, fn, cacheLogEntry{ID: fakeID})
	c.CacheReads = nil

	// TESTS LoadWithProperties method
	ps.LoadWithProperties(fakeID)
	fn = fmt.Sprintf("ps.LoadWithProperties(%q)", fakeID)
	verifyCacheRead(t, c, fn, cacheLogEntry{ID: fakeID})
	c.CacheReads = nil

	// TESTS LoadByID method
	p, _ := ps.LoadByID(fakeID)
	fn = fmt.Sprintf("ps.LoadByID(%q)", fakeID)
	verifyCacheRead(t, c, fn, cacheLogEntry{ID: fakeID})
	c.CacheReads = nil

	// TESTS LoadProperties method on Profile
	p.LoadProperties()
	fn = "p.LoadProperties()"
	verifyCacheRead(t, c, fn, cacheLogEntry{ID: fakeID})
	c.CacheReads = nil

	// TESTS LoadNameHistory method on Profile
	p.LoadNameHistory()
	fn = "p.LoadNameHistory()"
	verifyCacheRead(t, c, fn, cacheLogEntry{ID: fakeID})
	c.CacheReads = nil
}

// WRITE TEST:
// Methods attempt to read real profiles from the server.
// After each load the loaded profile is tested for being in the cache.
func TestCacheWrite(t *testing.T) {

	// SETUP
	c := newLoggingTestCache()
	ps := NewStore(c)

	// TESTS Load method
	_, err := ps.Load(realName)
	fn := fmt.Sprintf("ps.Load(%q)", realName)
	verifyCacheWrite(t, c, fn, cacheLogEntry{Name: realName}, err)
	c = newLoggingTestCache()
	ps = NewStore(c)

	// TESTS LoadAtTime method
	_, err = ps.LoadAtTime(realPastName, realPastTime)
	fn = fmt.Sprintf("ps.LoadAtTime(%q, %s)", realPastName, realPastTime)
	verifyCacheWrite(t, c, fn, cacheLogEntry{Name: realPastName, Time: realPastTime}, err)
	c = newLoggingTestCache()
	ps = NewStore(c)

	// TESTS LoadWithNameHistory method
	_, err = ps.LoadWithNameHistory(realID)
	fn = fmt.Sprintf("ps.LoadWithNameHistory(%q)", realID)
	verifyCacheWrite(t, c, fn, cacheLogEntry{ID: realID}, err)
	c = newLoggingTestCache()
	ps = NewStore(c)

	// TESTS LoadWithProperties method
	_, err = ps.LoadWithProperties(realID2)
	fn = fmt.Sprintf("ps.LoadWithProperties(%q)", realID2)
	verifyCacheWrite(t, c, fn, cacheLogEntry{ID: realID2}, err)
	c = newLoggingTestCache()
	ps = NewStore(c)

	// TESTS LoadByID method
	_, err = ps.LoadByID(realID)
	fn = fmt.Sprintf("ps.LoadByID(%q)", realID)
	verifyCacheWrite(t, c, fn, cacheLogEntry{ID: realID}, err)
	c = newLoggingTestCache()
	ps = NewStore(c)
}

// WRITE TEST:
// Loads a profile by ID, then clears the cache separate its cache write from the following
// invocation of LoadProperties. LoadProperties is then called and its cache write is verified.
func TestCacheWriteLoadProperties(t *testing.T) {

	// SETUP
	c := newLoggingTestCache()
	ps := NewStore(c)
	fn := fmt.Sprintf("ps.LoadByID(%q)", realID3)

	// TEST
	p, err := ps.LoadByID(realID3)
	if err != nil {

		t.Errorf("%s returned error: %s", fn, err)
		t.Error("Could not test cache write behaviour of LoadProperties() method on Profile.")

	} else {

		// Reset cache to verify that LoadProperties really writes to the cache
		c.EntriesByID = make(map[string]CacheEntry)

		_, err = p.LoadProperties()

		verifyCacheWrite(t, c, "p.LoadProperties()", cacheLogEntry{ID: realID3}, err)
	}
}

// WRITE TEST:
// Loads a profile by ID, then clears the cache separate its cache write from the following
// invocation of LoadNameHistory. LoadNameHistory is then called and its cache write is verified.
func TestCacheWriteLoadNameHistory(t *testing.T) {

	// SETUP
	c := newLoggingTestCache()
	ps := NewStore(c)
	fn := fmt.Sprintf("ps.Load(%q)", realName)

	// TEST -- LoadByID would preload name history under the hood
	p, err := ps.Load(realName)
	if err != nil {

		t.Errorf("%s returned error: %s", fn, err)
		t.Error("Could not test cache write behaviour of LoadNameHistory() method on Profile.")

	} else {

		// Reset cache to verify that LoadNameHistory really writes to the cache
		c.EntriesByID = make(map[string]CacheEntry)

		_, err = p.LoadNameHistory()

		verifyCacheWrite(t, c, "p.LoadNameHistory()", cacheLogEntry{ID: realID}, err)
	}
}
