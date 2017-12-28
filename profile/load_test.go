package profile

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/PhilipBorgesen/minecraft/internal"
)

var testLoadInput = [...]struct {
	username   string
	transport  http.RoundTripper
	expProfile *Profile
	expErr     error
}{
	{
		username:   "",
		transport:  nil,
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
	{
		username: "doesNotExist",
		transport: errorTransport{
			&internal.FailedRequestError{
				StatusCode: 204,
			},
		},
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
	{
		username:   "demoAccount",
		transport:  http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
	{
		username:   "unexpectedFormat",
		transport:  http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr: &url.Error{
			Op:  "Parse",
			URL: "https://api.mojang.com/users/profiles/minecraft/unexpectedFormat",
			Err: internal.ErrUnknownFormat,
		},
	},
	{
		username:  "nergalic",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			Name: "Nergalic",
			ID:   "087cc153c3434ff7ac497de1569affa1",
		},
		expErr: nil,
	},
}

func TestLoad(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testLoadInput {
		client.Transport = tc.transport
		profile, err := Load(context.Background(), tc.username)
		if !reflect.DeepEqual(profile, tc.expProfile) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"Load(ctx, %q)\n"+
					" was: %#v, %s\n"+
					"want: %#v, %s",
				tc.username,
				profile, p(err),
				tc.expProfile, p(tc.expErr),
			)
		}
	}
}

func TestLoadContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx := context.WithValue(context.Background(), "", nil)
	ct := CtxStoreTransport{}

	client.Transport = &ct
	Load(ctx, "nergalic")

	if ct.Context != ctx {
		t.Error("Load(ctx, \"nergalic\") didn't pass context to underlying http.Client")
	}
}

var testLoadAtTimeInput = [...]struct {
	username   string
	time       time.Time
	transport  http.RoundTripper
	expProfile *Profile
	expErr     error
}{
	{
		username:   "",
		time:       time.Unix(0, 0),
		transport:  nil,
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
	{
		username: "doesNotExist",
		time:     time.Unix(0, 0),
		transport: errorTransport{
			&internal.FailedRequestError{
				StatusCode: 204,
			},
		},
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
	{
		username:   "unexpectedFormat",
		time:       time.Unix(1337, 564),
		transport:  http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr: &url.Error{
			Op:  "Parse",
			URL: "https://api.mojang.com/users/profiles/minecraft/unexpectedFormat?at=1337",
			Err: internal.ErrUnknownFormat,
		},
	},
}

func TestLoadAtTime(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testLoadAtTimeInput {
		client.Transport = tc.transport
		profile, err := LoadAtTime(context.Background(), tc.username, tc.time)
		if !reflect.DeepEqual(profile, tc.expProfile) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"LoadAtTime(ctx, %q, %s)\n"+
					" was: %#v, %s\n"+
					"want: %#v, %s",
				tc.username, tc.time,
				profile, p(err),
				tc.expProfile, p(tc.expErr),
			)
		}
	}
}

func TestLoadAtTimeContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx := context.WithValue(context.Background(), "", nil)
	ct := CtxStoreTransport{}

	client.Transport = &ct
	LoadAtTime(ctx, "nergalic", time.Now())

	if ct.Context != ctx {
		t.Error("LoadAtTime(ctx, \"nergalic\", time.Now()) didn't pass context to underlying http.Client")
	}
}

var testLoadWithNameHistoryInput = [...]struct {
	id         string
	transport  http.RoundTripper
	expProfile *Profile
	expErr     error
}{
	{
		id:         "",
		transport:  nil,
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
	{
		id: "00000000000000000000000000000000", // Doesn't exist
		transport: errorTransport{
			&internal.FailedRequestError{
				StatusCode: 204,
			},
		},
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
	{
		id:        "087cc153c3434ff7ac497de1569affa1",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			Name: "Nergalic",
			ID:   "087cc153c3434ff7ac497de1569affa1",
			NameHistory: []PastName{
				{
					Name:  "GeneralSezuan",
					Until: msToTime(1423047705000),
				},
			},
		},
		expErr: nil,
	},
}

func TestLoadWithNameHistory(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testLoadWithNameHistoryInput {
		client.Transport = tc.transport
		profile, err := LoadByID(context.Background(), tc.id) // Wrapper method used to test that as well
		if !reflect.DeepEqual(profile, tc.expProfile) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"LoadWithNameHistory(ctx, %q)\n"+
					" was: %#v, %s\n"+
					"want: %#v, %s",
				tc.id,
				profile, p(err),
				tc.expProfile, p(tc.expErr),
			)
		}
	}
}

func TestLoadWithNameHistoryContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx := context.WithValue(context.Background(), "", nil)
	ct := CtxStoreTransport{}

	client.Transport = &ct
	LoadByID(ctx, "dummy") // Wrapper method used to test that as well

	if ct.Context != ctx {
		t.Error("LoadWithNameHistory(ctx, \"dummy\") didn't pass context to underlying http.Client")
	}
}

var testLoadWithPropertiesInput = [...]struct {
	id         string
	transport  http.RoundTripper
	expProfile *Profile
	expErr     error
}{
	{
		id:         "",
		transport:  nil,
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
	{
		id: "00000000000000000000000000000000", // Doesn't exist
		transport: errorTransport{
			&internal.FailedRequestError{
				StatusCode: 204,
			},
		},
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
	{
		id:        "087cc153c3434ff7ac497de1569affa1",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			Name: "Nergalic",
			ID:   "087cc153c3434ff7ac497de1569affa1",
			Properties: &Properties{
				SkinURL: "http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e",
				CapeURL: "",
				Model:   Steve,
			},
		},
		expErr: nil,
	},
	{
		id:         "fictiveDemo",
		transport:  http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr:     ErrNoSuchProfile,
	},
}

func TestLoadWithProperties(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testLoadWithPropertiesInput {
		client.Transport = tc.transport
		profile, err := LoadWithProperties(context.Background(), tc.id)
		if !reflect.DeepEqual(profile, tc.expProfile) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"LoadWithProperties(ctx, %q)\n"+
					" was: %#v, %s\n"+
					"want: %#v, %s",
				tc.id,
				profile, p(err),
				tc.expProfile, p(tc.expErr),
			)
		}
	}
}

func TestLoadWithPropertiesContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx := context.WithValue(context.Background(), "", nil)
	ct := CtxStoreTransport{}

	client.Transport = &ct
	LoadWithProperties(ctx, "dummy")

	if ct.Context != ctx {
		t.Error("LoadWithProperties(ctx, \"dummy\") didn't pass context to underlying http.Client")
	}
}

var testLoadManyInput = [...]struct {
	ids         []string
	transport   http.RoundTripper
	expProfiles []*Profile
	expErr      error
}{
	{
		ids:         []string{},
		transport:   nil,
		expProfiles: nil,
		expErr:      nil,
	},
	{
		ids:         []string{""},
		transport:   nil,
		expProfiles: nil,
		expErr:      nil,
	},
	{
		ids:         make([]string, LoadManyMaxSize+1, LoadManyMaxSize+1),
		transport:   nil,
		expProfiles: nil,
		expErr:      ErrMaxSizeExceeded{LoadManyMaxSize + 1},
	},
	{
		ids:         []string{"dummy"},
		transport:   http.NewFileTransport(http.Dir("testdata/LoadMany/unexpectedFormat")),
		expProfiles: nil,
		expErr: &url.Error{
			Op:  "Parse",
			URL: "https://api.mojang.com/profiles/minecraft",
			Err: internal.ErrUnknownFormat,
		},
	},
	{
		ids:       []string{"nergalic", "AxeLaw", "demo", "doesNotExist"},
		transport: http.NewFileTransport(http.Dir("testdata/LoadMany/success")),
		expProfiles: []*Profile{
			{
				ID:          "cabefc91b5df4c87886a6c604da2e46f",
				Name:        "AxeLaw",
				NameHistory: emptyHist,
			},
			{
				ID:   "087cc153c3434ff7ac497de1569affa1",
				Name: "Nergalic",
			},
		},
		expErr: nil,
	},
}

func TestLoadMany(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testLoadManyInput {
		client.Transport = tc.transport
		profiles, err := LoadMany(context.Background(), tc.ids...)
		if !reflect.DeepEqual(profiles, tc.expProfiles) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"LoadMany(ctx, %q)\n"+
					" was: %s, %s\n"+
					"want: %s, %s",
				tc.ids,
				profiles, p(err),
				tc.expProfiles, p(tc.expErr),
			)
		}
	}
}

func TestLoadManyContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx := context.WithValue(context.Background(), "", nil)
	ct := CtxStoreTransport{}

	client.Transport = &ct
	LoadMany(ctx, "dummy")

	if ct.Context != ctx {
		t.Error("LoadMany(ctx, \"dummy\") didn't pass context to underlying http.Client")
	}
}

/***************
*  TEST UTILS  *
***************/

var testError = errors.New("testError")

func p(x interface{}) interface{} {
	if x == nil {
		return "<nil>"
	} else {
		return x
	}
}

type statusOverrideTransport struct {
	status    int
	transport http.RoundTripper
}

func (sot statusOverrideTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = sot.transport.RoundTrip(req)
	resp.StatusCode = sot.status
	return
}

type errorTransport struct {
	err error
}

func (et errorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, et.err
}

type CtxStoreTransport struct {
	Context context.Context
}

func (ct *CtxStoreTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ct.Context = req.Context()
	return nil, errors.New("RoundTrip was called")
}
