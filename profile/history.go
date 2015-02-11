package profile

import (
	"time"
)

/************
* PAST NAME *
************/

// PastName represents one of a profile's past usernames.
type PastName struct {
	name  string
	until time.Time
}

// Name returns the username represented by p.
func (p PastName) Name() string {

	return p.name
}

// Until returns the time at which the profile stopped using this username.
func (p PastName) Until() time.Time {

	return p.until
}

// String uses Name as its string representation.
func (p PastName) String() string {

	return p.name
}

/******************
* PROFILE METHODS *
******************/

// LoadNameHistory loads and returns a copy of the profile's past usernames sorted in ascending order.
// On success, the result will be memoized in a thread safe manner to be returned on future calls to this method.
// If an error occurs, nil is returned instead and the result is not memoized.
// A profile which was loaded by LoadWithNameHistory has this information preloaded.
func (p *Profile) LoadNameHistory() ([]PastName, error) {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.history == nil {

		loaded, err := LoadWithNameHistory(p.id)
		if err != nil {

			return nil, err
		}

		p.history = loaded.history
	}

	// Make copy
	hist := append(emptyHist, p.history...)

	return hist, nil
}

// NameHistory returns a copy of the profile's past usernames sorted in ascending order if loaded, otherwise nil.
func (p *Profile) NameHistory() []PastName {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	return append(emptyHist, p.history...)
}

/************
* INTERNALS *
************/

var emptyHist = make([]PastName, 0, 0)

// Creates an ascending username history from JSON decoded using the encoding/json package
// a is an array of objects containing "name" and (possibly) "changedToAt" fields
// A "changedToAt" field is the "until" field of the previous object
func createHistory(a []interface{}) (string, []PastName) {

	hist := make([]PastName, len(a)-1)
	var name string

	// ASSUMPTION: Mojang returns name history ascendingly
	// TODO: Verify
	for i, v := range a {

		m := v.(map[string]interface{})

		if v, ok := m["changedToAt"]; ok && i > 0 {

			hist[i-1].until = time.Unix(int64(v.(float64))/1000, 0)
		}

		if i == len(hist) {

			name = m["name"].(string)
			break
		}

		hist[i].name = m["name"].(string)
	}

	return name, hist
}
