package profile

import (
	"time"
)

/*************
* CACHE TYPE *
*************/

// Cache is an interface allowing a caching mechanism to be used with the package
// through use of the Store type.
// Implementers are responsible for the thread safety of their implementations.
//
// The Load, LoadAtTime, LoadWithNameHistory, LoadWithProperties and LoadByID
// methods of a Store will seek its Cache for a matching entry before
// making a new server request. An entry matches if it contains sufficient
// information to serve the needs of the calling method, i.e.:
// ID and Name for Load, LoadAtTime and LoadByID;
// ID, Name and NameHistory for LoadWithNameHistory; and
// ID, Name and Properties for LoadWithProperties.
//
// When a Store method loads a profile from the server it will create a cache
// entry containing all retrieved information and pass it to the cache by calling
// Cache. The LoadAtTime method will additionally call CacheNameAtTime to cache the
// profile ID a username was associated with at a specific time in the past.
type Cache interface {

	// Cache the profile ID a username was associated with at a specific time.
	// These mappings will never become invalid.
	CacheNameAtTime(name string, tm time.Time, id string)

	// Cache a profile for later lookup by GetName, GetNameAtTime and GetID.
	Cache(e CacheEntry)

	// Retrieve the cached profile for a username.
	// Usernames should be compared ignoring case as Minecraft usernames are case-insensitive,
	// although case-preserving. "USER", "user", "User" and "uSeR" are all equivalent.
	// If no cache entry was found, the boolean return value is false, otherwise true.
	GetName(name string) (entry CacheEntry, ok bool)

	// Retrieve the cached profile for a username at a specific time.
	// Usernames should be compared ignoring case as Minecraft usernames are case-insensitive,
	// although case-preserving. "USER", "user", "User" and "uSeR" are all equivalent.
	// If no cache entry was found, the boolean return value is false, otherwise true.
	GetNameAtTime(name string, tm time.Time) (entry CacheEntry, ok bool)

	// Retrieve the cached profile for an ID.
	// If no cache entry was found, the boolean return value is false, otherwise true.
	GetID(id string) (entry CacheEntry, ok bool)
}

// CacheEntry represents an entry in a profile cache.
type CacheEntry struct {

	// The ID of the cached profile
	ID string

	// The cached username of the profile identified by ID
	Name string

	// The cached name history for the profile identified by ID.
	// nil if no name history is cached.
	// The slice should be treated as immutable by all clients.
	NameHistory []PastName

	// The cached properties for the profile identified by ID
	// nil if no properties are cached.
	Properties *Properties
}

/****************
* PROFILE STORE *
****************/

// NewStore constructs a new Store for loading profiles using c as
// the caching mechanism.
func NewStore(c Cache) Store {

	return Store{c}
}

// The default Store using no caching mechanism.
// Loading profiles using its methods is the same as using the identically named package functions.
var NoCacheStore = NewStore(nil)

// A Store provides methods for calling this package's load
// functions with additional caching functionality.
// Only if the store's caching mechanism not already have cached the
// requested profile information is the profile attempted loaded from the Mojang servers.
// When a profile is loaded from the Mojang servers it is automatically passed to
// the cache mechanism. See Cache for details.
type Store struct {
	cache Cache
}

// Cache returns the cache used by the store.
// If no cache is used, nil is returned.
func (s Store) Cache() Cache {

	return s.cache
}

// Load functions like the Load package function, additionally trying to
// fetch profiles from the Store's cache before issuing requests to
// the Mojang servers.
//
// Profiles successfully loaded from the Mojang servers will be passed to
// the Store's cache as a CacheEntry.
func (s Store) Load(username string) (*Profile, error) {

	c := s.cache

	// No caching?
	if c == nil {

		return Load(username)
	}

	// Profile is cached?
	if e, ok := c.GetName(username); ok {

		return cToP(&e, c), nil
	}

	// Load and cache
	p, err := Load(username)
	if err != nil {

		return p, err
	}
	p.cache = c

	c.Cache(pToC(p))

	return p, err
}

// LoadAtTime functions like the LoadTime package function, additionally
// trying to fetch profiles from the Store's cache before issuing requests
// to the Mojang servers.
//
// Profiles successfully loaded from the Mojang servers will be passed to
// the Store's cache as a CacheEntry.
func (s Store) LoadAtTime(username string, tm time.Time) (*Profile, error) {

	c := s.cache

	// No caching?
	if c == nil {

		return LoadAtTime(username, tm)
	}

	// Profile is cached?
	if e, ok := c.GetNameAtTime(username, tm); ok {

		return cToP(&e, c), nil
	}

	// Load and cache
	p, err := LoadAtTime(username, tm)
	if err != nil {

		return p, err
	}
	p.cache = c

	c.Cache(pToC(p))
	c.CacheNameAtTime(username, tm, p.id)

	return p, err
}

// LoadByID functions like the LoadByID package function, additionally
// trying to fetch profiles from the Store's cache before issuing requests
// to the Mojang servers.
//
// Profiles successfully loaded from the Mojang servers will be passed to
// the Store's cache as a CacheEntry.
func (s Store) LoadByID(id string) (*Profile, error) {

	c := s.cache

	// No caching?
	if c == nil {

		return LoadByID(id)
	}

	// Profile is cached?
	if e, ok := c.GetID(id); ok {

		return cToP(&e, c), nil
	}

	// Load and cache
	p, err := LoadByID(id)
	if err != nil {

		return p, err
	}
	p.cache = c

	c.Cache(pToC(p))

	return p, err
}

// LoadWithNameHistory functions like the LoadWithNameHistory package function,
// additionally trying to fetch profiles from the Store's cache before issuing
// requests to the Mojang servers.
//
// Profiles successfully loaded from the Mojang servers will be passed to
// the Store's cache as a CacheEntry.
func (s Store) LoadWithNameHistory(id string) (*Profile, error) {

	c := s.cache

	// No caching?
	if c == nil {

		return LoadWithNameHistory(id)
	}

	// Profile is cached?
	if e, ok := c.GetID(id); ok && e.NameHistory != nil {

		return cToP(&e, c), nil
	}

	// Load and cache
	p, err := LoadWithNameHistory(id)
	if err != nil {

		return p, err
	}
	p.cache = c

	c.Cache(pToC(p))

	return p, err
}

// LoadWithProperties functions like the LoadWithProperties package function,
// additionally trying to fetch profiles from the Store's cache before issuing
// requests to the Mojang servers.
//
// Profiles successfully loaded from the Mojang servers will be passed to
// the Store's cache as a CacheEntry.
func (s Store) LoadWithProperties(id string) (*Profile, error) {

	c := s.cache

	// No caching?
	if c == nil {

		return LoadWithProperties(id)
	}

	// Profile is cached?
	if e, ok := c.GetID(id); ok && e.Properties != nil {

		return cToP(&e, c), nil
	}

	// Load and cache
	p, err := LoadWithProperties(id)
	if err != nil {

		return p, err
	}
	p.cache = c

	c.Cache(pToC(p))

	return p, err
}

/************
* INTERNALS *
************/

// Simple constructor of profiles from cache entries
func cToP(e *CacheEntry, c Cache) *Profile {

	return &Profile{
		name:       e.Name,
		id:         e.ID,
		history:    e.NameHistory,
		properties: e.Properties,
		cache:      c,
	}
}

// Simple constructor of cache entries from profiles
func pToC(p *Profile) CacheEntry {

	return CacheEntry{
		Name:        p.name,
		ID:          p.id,
		NameHistory: p.history,
		Properties:  p.properties,
	}
}
