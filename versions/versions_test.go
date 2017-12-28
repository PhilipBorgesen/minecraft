package versions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/PhilipBorgesen/minecraft/internal"
)

var loadExpectations = [...]Version{
	{
		ID:       "rd-132211",
		Released: time.Date(2009, 05, 13, 20, 11, 00, 00, time.UTC),
		Type:     Alpha,
	},
	{
		ID:       "b1.0",
		Released: time.Date(2010, 12, 19, 22, 00, 00, 00, time.UTC),
		Type:     Beta,
	},
	{
		ID:       "1.0",
		Released: time.Date(2011, 11, 17, 22, 00, 00, 00, time.UTC),
		Type:     Release,
	},
	{ // Snapshot versions need to be updated along with testdata/version_manifest.json
		ID:       "16w50a",
		Released: time.Date(2016, 12, 15, 14, 38, 52, 00, time.UTC),
		Type:     Snapshot,
	},
}

// Test that specific known versions are decoded correctly
func TestLoadSpecifics(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	client.Transport = http.NewFileTransport(http.Dir("testdata/cached"))
	vs, err := Load(context.Background())
	if err != nil {
		t.Fatalf("Load(ctx) failed to fetch a version listing: %s", err)
	}

	for _, le := range loadExpectations {
		if v, ok := vs.Versions[le.ID]; !ok {
			t.Errorf("Load(ctx).Versions[%q] is missing", le.ID)
		} else if !v.Equal(le) {
			t.Errorf("Load(ctx).Versions[%q] was:\n"+
				"      %s\n"+
				"want: %s",
				le.ID, pVersion(v), pVersion(le))
		}
	}
}

func TestLoadContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx := context.WithValue(context.Background(), dummy, nil)
	ct := CtxStoreTransport{}

	client.Transport = &ct
	Load(ctx)

	if ct.Context != ctx {
		t.Error("Load(ctx) didn't pass context to underlying http.Client")
	}
}

// Test that Load succeeds and that all returned Versions data is populated.
func TestLoadInvariants(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	client.Transport = http.NewFileTransport(http.Dir("testdata/cached"))
	vs, err := Load(context.Background())
	if err != nil {
		t.Fatalf("Load(ctx) failed to fetch a version listing: %s", err)
	}

	// Verify integrity of vs.Latest.Release
	if r := vs.Latest.Release; r == "" {
		t.Errorf("Load(ctx).Latest.Release = %q; want non-empty", "")
	} else if _, ok := vs.Versions[r]; !ok {
		t.Error("Load(ctx).Versions contained no information for Load(ctx).Latest.Release")
	} else if vs.LatestRelease() != vs.Versions[r] {
		t.Error("Load(ctx).LatestRelease() differed from Load(ctx).Versions[Load().Latest.Release]")
	} else if vt := vs.Versions[r].Type; vt != Release {
		t.Errorf("Load(ctx).Latest.Release denoted a version of type %s; want %s", vt, Release)
	}

	// Verify integrity of vs.Latest.Snapshot
	if s := vs.Latest.Snapshot; s == "" {
		t.Errorf("Load(ctx).Latest.Snapshot = %q; want non-empty", "")
	} else if _, ok := vs.Versions[s]; !ok {
		t.Error("Load(ctx).Versions contained no information for Load(ctx).Latest.Snapshot")
	} else if vt := vs.Versions[s].Type; vt != Snapshot {
		t.Errorf("Load(ctx).Latest.Snapshot denoted a version of type %s; want %s", vt, Snapshot)
	}

	// Verify every entry of vs.Versions
	for id, v := range vs.Versions {
		if id != v.ID {
			t.Errorf("Load(ctx).Versions[%q].ID = %q; want %q", id, v.ID, id)
		}
		if v.ID == "" {
			t.Errorf("Load(ctx).Versions[%q].ID = %q; want non-empty", id, "")
		}
		if v.Released.IsZero() {
			t.Errorf("Load(ctx).Versions[%q].Released = %s; want non-zero", id, v.Released)
		}
		if v.Type == "" {
			t.Errorf("Load(ctx).Versions[%q].Type = %s; want non-zero value", id, v.Type)
		}
		if s := v.String(); s != v.ID {
			t.Errorf("Load(ctx).Versions[%q].String() = %q; want %q", id, s, id)
		}
	}
}

var testLoadErrorsInput = [...]struct {
	transport http.RoundTripper
	op        string
	errStr    string
}{
	{
		transport: http.NewFileTransport(http.Dir("testdata/nonexisting")),
		op:        "Get",
		errStr:    (&internal.FailedRequestError{StatusCode: 404}).Error(),
	},
	{
		transport: http.NewFileTransport(http.Dir("testdata/malstructured")),
		op:        "Parse",
		errStr:    internal.ErrUnknownFormat.Error(),
	},
}

func TestLoadError(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testLoadErrorsInput {
		expErr := &url.Error{
			Op:  tc.op,
			URL: versionsURL,
			Err: errors.New(tc.errStr),
		}

		client.Transport = tc.transport
		vs, err := Load(context.Background())

		if !urlErrorAlike(expErr, err) || !reflect.DeepEqual(vs, Listing{}) {
			t.Errorf("Load(ctx) returned result:\n"+
				"      %s, %s\n"+
				"want: %s, %s",
				vs, err,
				Listing{}, expErr)
		}
	}
}

func TestLatestReleasePanic(t *testing.T) {
	var l Listing
	l.Versions = make(map[string]Version)
	l.Latest.Release = "doesNotExist"

	test := func() (panicked bool) {
		defer func() { recover() }()
		panicked = true
		l.LatestRelease()
		return false
	}

	if panicked := test(); !panicked {
		t.Error("LatestRelease() didn't panic as expected")
	}
}

var knownTypes = [...]struct {
	t Type
	s string
}{
	{
		t: Alpha,
		s: "alpha",
	},
	{
		t: Beta,
		s: "beta",
	},
	{
		t: Release,
		s: "release",
	},
	{
		t: Snapshot,
		s: "snapshot",
	},
	{ // None
		t: Type(""),
		s: "???",
	},
	{ // Not handled by String method
		t: Type("UNDEFINED"),
		s: "UNDEFINED",
	},
}

func TestTypeString(t *testing.T) {
	for _, kt := range knownTypes {
		if s := kt.t.String(); s != kt.s {
			t.Errorf("Type(%q).String() was %q; want %q", string(kt.t), s, kt.s)
		}
	}
}

/*************
* TEST UTILS *
*************/

var dummy struct{}

func pVersion(v Version) string {
	return fmt.Sprintf("Version{ID: %q, Released: %s, Type: %s}", v.ID, v.Released, v.Type)
}

func urlErrorAlike(exp *url.Error, err error) bool {
	e, ok := err.(*url.Error)
	if !ok {
		return false
	}

	if e == nil {
		if exp != nil {
			return false
		}
	} else if exp == nil {
		return false
	}

	return e.Op == exp.Op && e.URL == exp.URL && e.Err.Error() == exp.Err.Error()
}

type CtxStoreTransport struct {
	Context context.Context
}

func (ct *CtxStoreTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ct.Context = req.Context()
	return nil, errors.New("RoundTrip was called")
}
