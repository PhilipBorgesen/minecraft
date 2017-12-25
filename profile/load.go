package profile

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PhilipBorgesen/minecraft/internal"
)

// LoadManyMaxSize is the maximum number of profiles which may be requested at
// once using LoadMany. If more are requested, the request may fail with an
// ErrMaxSizeExceeded error.
const LoadManyMaxSize int = 100

// Load fetches the profile currently associated with username. ctx must be
// non-nil. If no profile currently is associated with username, Load returns
// ErrNoSuchProfile. If an error is returned, p will be nil.
func Load(ctx context.Context, username string) (p *Profile, err error) {
	if username == "" {
		return nil, ErrNoSuchProfile
	}
	endpoint := fmt.Sprintf(loadURL, username)
	return loadByName(ctx, endpoint)
}

// LoadAtTime fetches the profile associated with username at the specified
// instant of time. ctx must be non-nil. If no profile was associated with
// username at the specified instant of time, LoadAtTime returns
// ErrNoSuchProfile. If an error is returned, p will be nil.
func LoadAtTime(ctx context.Context, username string, t time.Time) (p *Profile, err error) {
	if username == "" {
		return nil, ErrNoSuchProfile
	}
	endpoint := fmt.Sprintf(loadAtTimeURL, username, t.Unix())
	return loadByName(ctx, endpoint)
}

// Common implementation used by Load and LoadAtTime.
func loadByName(ctx context.Context, endpoint string) (p *Profile, err error) {
	js, err := internal.FetchJSON(ctx, client, endpoint)
	if err != nil {
		return nil, transformError(err)
	}

	defer func() { // If JSON data isn't structured as expected
		if r := recover(); r != nil {
			p = nil
			err = &url.Error{Op: "Parse", URL: endpoint, Err: internal.ErrUnknownFormat}
		}
	}()

	p = &Profile{}
	if !fillProfile(p, js.(map[string]interface{})) {
		return nil, ErrNoSuchProfile
	}

	return p, nil
}

// LoadByID fetches the profile identified by id. ctx must be non-nil. If no
// profile is identified by id, LoadByID returns ErrNoSuchProfile. If an error
// is returned, p will be nil.
func LoadByID(ctx context.Context, id string) (p *Profile, err error) {
	return LoadWithNameHistory(ctx, id)
}

// LoadWithNameHistory fetches the profile identified by id, incl. its name
// history. ctx must be non-nil. If no profile is identified by id,
// LoadWithNameHistory returns ErrNoSuchProfile. If an error is returned,
// p will be nil.
func LoadWithNameHistory(ctx context.Context, id string) (p *Profile, err error) {
	if id == "" {
		return nil, ErrNoSuchProfile
	}
	pr := Profile{ID: id}
	_, err = pr.LoadNameHistory(ctx, true)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

// LoadWithProperties fetches the profile identified by id, incl. its
// properties. ctx must be non-nil. If no profile is identified by id,
// LoadWithProperties returns ErrNoSuchProfile. If an error is returned,
// p will be nil.
//
// NB! For each profile, profile properties may only be requested once per
// minute.
func LoadWithProperties(ctx context.Context, id string) (p *Profile, err error) {
	if id == "" {
		return nil, ErrNoSuchProfile
	}
	pr := Profile{ID: id}
	_, err = pr.LoadProperties(ctx, true)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

// LoadMany fetches multiple profiles by their currently associated usernames.
// Usernames associated with no profile are ignored and absent from the
// returned results. Duplicate usernames are only returned once, and ps will be
// nil if an error occurs. ctx must be non-nil.
//
// NB! Only a maximum of LoadManyMaxSize profiles may be fetched at once.
// If more are attempted loaded in the same operation, an ErrMaxSizeExceeded
// error is returned.
func LoadMany(ctx context.Context, usernames ...string) (ps []*Profile, err error) {
	if len(usernames) > LoadManyMaxSize {
		return nil, ErrMaxSizeExceeded{len(usernames)}
	}

	c := 0
	var users [LoadManyMaxSize]string
	for _, u := range usernames {
		// Remove empty usernames. They are not accepted by the Mojang API.
		if u != "" {
			users[c] = u
			c++
		}
	}

	if c == 0 {
		return nil, nil // No need to request anything
	}

	js, err := internal.ExchangeJSON(ctx, client, loadManyURL, users[:c])
	if err != nil {
		return nil, transformError(err)
	}

	defer func() { // If JSON data isn't structured as expected
		if r := recover(); r != nil {
			err = &url.Error{Op: "Parse", URL: loadManyURL, Err: internal.ErrUnknownFormat}
			ps = nil
		}
	}()

	arr := js.([]interface{})
	ps = make([]*Profile, 0, len(arr))

	var pr *Profile
	for _, p := range arr {
		if pr == nil {
			pr = &Profile{} // Reuse allocation of skipped demo profile
		}
		if !fillProfile(pr, p.(map[string]interface{})) {
			continue
		}
		ps = append(ps, pr)
		pr = nil
	}
	return ps, nil
}

var client = &http.Client{}

func transformError(src error) error {
	if e, ok := internal.UnwrapFailedRequestError(src); ok {
		if e.StatusCode == 204 {
			return ErrNoSuchProfile
		} else if e.ErrorCode == "TooManyRequestsException" {
			return ErrTooManyRequests
		}
	}
	return src
}
