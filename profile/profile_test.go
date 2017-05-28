package profile

import (
	"testing"
	"time"
	"context"
	"reflect"
	"net/http"
	"errors"
	"bytes"
	"io/ioutil"
	"net/url"
	"github.com/PhilipBorgesen/minecraft/internal"
)

func TestProfile_String(t *testing.T) {
	name := "TESTNAME&%"
	p := &Profile{Name: name}
	s := p.String()
	if s != name {
		t.Errorf(
			"&Profile{Name: %q}.String() was %q; want %q",
			name, s, name,
		)
	}
}

var testPastName_EqualInput = [...]struct {
	pn1    PastName
	pn2    PastName
	equals bool
}{
	{
		pn1:    PastName{},
		pn2:    PastName{},
		equals: true,
	},
	{
		pn1:    PastName{Name: "A"},
		pn2:    PastName{Name: "A"},
		equals: true,
	},
	{
		pn1:    PastName{Until: time.Unix(23, 42)},
		pn2:    PastName{Until: time.Unix(23, 42)},
		equals: true,
	},
	{
		pn1:    PastName{Name: "lolol", Until: time.Unix(52, 37)},
		pn2:    PastName{Name: "lolol", Until: time.Unix(52, 37)},
		equals: true,
	},
	{
		pn1:    PastName{Until: time.Unix(23, 42).In(time.FixedZone("Zone A", 2))},
		pn2:    PastName{Until: time.Unix(23, 42).In(time.FixedZone("Zone B", 7))},
		equals: true,
	},

	{
		pn1:    PastName{Name: "ABC", Until: time.Unix(52, 37)},
		pn2:    PastName{Name: "12t", Until: time.Unix(52, 37)},
		equals: false,
	},
	{
		pn1:    PastName{Name: "ABC", Until: time.Unix(52, 37)},
		pn2:    PastName{Name: "ABC", Until: time.Unix(23, 37)},
		equals: false,
	},
	{
		pn1:    PastName{Name: "ABC", Until: time.Unix(23, 66)},
		pn2:    PastName{Name: "ABC", Until: time.Unix(23, 37)},
		equals: false,
	},
}

func TestPastName_Equal(t *testing.T) {
	for _, tc := range testPastName_EqualInput {
		if res := tc.pn1.Equal(tc.pn2); res != tc.equals {
			t.Errorf(
				"\n"+
					"%#v.Equal(\n"+
					"%#v) was %t; want %t",
				tc.pn1, tc.pn2, res, tc.equals,
			)
		}
	}
}

func TestPastName_String(t *testing.T) {
	name := "TESTNAME&%"
	p := PastName{Name: name}
	s := p.String()
	if s != name {
		t.Errorf(
			"PastName{Name: %q}.String() was %q; want %q",
			name, s, name,
		)
	}
}

var testModel_StringInput = [...]struct {
	model  Model
	expStr string
}{
	{
		model:  Steve,
		expStr: "Steve",
	},
	{
		model:  Alex,
		expStr: "Alex",
	},
	{
		model:  Model(99),
		expStr: "???",
	},
}

func TestModel_String(t *testing.T) {
	for _, tc := range testModel_StringInput {
		s := tc.model.String()
		if s != tc.expStr {
			t.Errorf(
				"Model(%d).String() was %q; want %q",
				byte(tc.model), s, tc.expStr,
			)
		}
	}
}

var testProfile_LoadNameHistoryInput = [...]struct {
	profile *Profile
	force bool
	transport http.RoundTripper
	expProfile *Profile
	expHist []PastName
	expErr error
} {
	{ // Load when history is unknown
		profile: &Profile{
			ID: "087cc153c3434ff7ac497de1569affa1",
		},
		force: false,
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			Name: "Nergalic",
			ID: "087cc153c3434ff7ac497de1569affa1",
			NameHistory: []PastName {
				{
					Name: "GeneralSezuan",
					Until: msToTime(1423047705000),
				},
			},
		},
		expHist: []PastName{
			{
				Name: "GeneralSezuan",
				Until: msToTime(1423047705000),
			},
		},
		expErr: nil,
	},
	{ // Don't load when unforced and history is known
		profile: &Profile{
			ID: "087cc153c3434ff7ac497de1569affa1",
			NameHistory: []PastName{
				{
					Name: "NameInFakeHistory",
					Until: msToTime(1234047705000),
				},
			},
		},
		force: false,
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			Name: "",
			ID: "087cc153c3434ff7ac497de1569affa1",
			NameHistory: []PastName{
				{
					Name: "NameInFakeHistory",
					Until: msToTime(1234047705000),
				},
			},
		},
		expHist: []PastName{
			{
				Name: "NameInFakeHistory",
				Until: msToTime(1234047705000),
			},
		},
		expErr: nil,
	},
	{ // Load when forced despite history is known
		profile: &Profile{
			ID: "087cc153c3434ff7ac497de1569affa1",
			NameHistory: []PastName{
				{
					Name: "NameInFakeHistory",
					Until: msToTime(1234047705000),
				},
			},
		},
		force: true,
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			Name: "Nergalic",
			ID: "087cc153c3434ff7ac497de1569affa1",
			NameHistory: []PastName{
				{
					Name: "GeneralSezuan",
					Until: msToTime(1423047705000),
				},
			},
		},
		expHist: []PastName{
			{
				Name: "GeneralSezuan",
				Until: msToTime(1423047705000),
			},
		},
		expErr: nil,
	},
	{ // Old history returned (but not updated) on error
		profile: &Profile{
			ID: "",
			NameHistory: []PastName{
				{Name: "NotReplaced", Until: time.Unix(0, 0)},
			},
		},
		force: true,
		transport: nil,
		expProfile: &Profile{
			NameHistory: []PastName{
				{Name: "NotReplaced", Until: time.Unix(0, 0)},
			},
		},
		expHist: []PastName{
			{Name: "NotReplaced", Until: time.Unix(0, 0)},
		},
		expErr: errors.New("p.ID was not set"),
	},
	{ // Unforced: Old history returned (but not updated) on error
		profile: &Profile{
			ID: "00000000000000000000000000000000", // Doesn't exist
		},
		force: false,
		transport: statusOverrideTransport{status: 204, transport: http.NewFileTransport(http.Dir("testdata"))},
		expProfile: &Profile{
			ID: "00000000000000000000000000000000",
		},
		expHist: nil,
		expErr: ErrNoSuchProfile,
	},
}

func TestProfile_LoadNameHistory(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testProfile_LoadNameHistoryInput {
		client.Transport = tc.transport
		profile := *tc.profile

		hist, err := profile.LoadNameHistory(context.Background(), tc.force)
		if !reflect.DeepEqual(&profile, tc.expProfile) || !reflect.DeepEqual(hist, tc.expHist) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"%#v.LoadNameHistory(ctx, %t) produced result:\n" +
					"      %#v, %#v, %s\n" +
					"want: %#v, %#v, %s",
				tc.profile, tc.force,
				&profile, hist, p(err),
				tc.expProfile, tc.expHist, p(tc.expErr),
			)
		}
	}
}

func TestProfile_LoadNameHistory_ContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx, _ := context.WithCancel(context.Background())
	ct := CtxStoreTransport{}

	client.Transport = &ct

	profile := Profile {ID: "dummy"}
	profile.LoadNameHistory(ctx, true)

	if ct.Context != ctx {
		t.Error("Profile{ID: \"dummy\"}.LoadNameHistory(ctx, true) didn't pass context to underlying http.Client")
	}
}

var testProfile_LoadPropertiesInput = [...]struct {
	profile *Profile
	force bool
	transport http.RoundTripper
	expProfile *Profile
	expProps *Properties
	expErr error
} {
	{ // Load when properties are unknown
		profile: &Profile{
			ID: "087cc153c3434ff7ac497de1569affa1",
		},
		force: false,
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			Name: "Nergalic",
			ID: "087cc153c3434ff7ac497de1569affa1",
			Properties: &Properties{
				SkinURL: "http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e",
				CapeURL: "",
				Model: Steve,
			},
		},
		expProps: &Properties{
			SkinURL: "http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e",
			CapeURL: "",
			Model: Steve,
		},
		expErr: nil,
	},
	{ // Don't load when unforced and properties are known
		profile: &Profile{
			ID: "087cc153c3434ff7ac497de1569affa1",
			Properties: &Properties{
				SkinURL: "dummy",
				CapeURL: "dummy",
				Model: Alex,
			},
		},
		force: false,
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			Name: "",
			ID: "087cc153c3434ff7ac497de1569affa1",
			Properties: &Properties{
				SkinURL: "dummy",
				CapeURL: "dummy",
				Model: Alex,
			},
		},
		expProps: &Properties{
			SkinURL: "dummy",
			CapeURL: "dummy",
			Model: Alex,
		},
		expErr: nil,
	},
	{ // Load when forced despite properties are known
		profile: &Profile{
			ID: "087cc153c3434ff7ac497de1569affa1",
			Properties: &Properties{
				SkinURL: "dummy",
				CapeURL: "dummy",
				Model: Alex,
			},
		},
		force: true,
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			Name: "Nergalic",
			ID: "087cc153c3434ff7ac497de1569affa1",
			Properties: &Properties{
				SkinURL: "http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e",
				CapeURL: "",
				Model: Steve,
			},
		},
		expProps: &Properties{
			SkinURL: "http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e",
			CapeURL: "",
			Model: Steve,
		},
		expErr: nil,
	},
	{ // Old properties returned (but not updated) on error
		profile: &Profile{
			ID: "",
			Properties: &Properties{
				SkinURL: "notReplaced",
				CapeURL: "notReplaced",
				Model: Alex,
			},
		},
		force: true,
		transport: nil,
		expProfile: &Profile{
			Properties: &Properties{
				SkinURL: "notReplaced",
				CapeURL: "notReplaced",
				Model: Alex,
			},
		},
		expProps: &Properties{
			SkinURL: "notReplaced",
			CapeURL: "notReplaced",
			Model: Alex,
		},
		expErr: errors.New("p.ID was not set"),
	},
	{ // Unforced: Old properties returned (but not updated) on error
		profile: &Profile{
			ID: "fictiveDemo", // Doesn't exist
		},
		force: false,
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile{
			ID: "fictiveDemo",
		},
		expProps: nil,
		expErr: ErrNoSuchProfile,
	},
}

func TestProfile_LoadProperties(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testProfile_LoadPropertiesInput {
		client.Transport = tc.transport
		profile := *tc.profile

		props, err := profile.LoadProperties(context.Background(), tc.force)
		if !reflect.DeepEqual(&profile, tc.expProfile) || !reflect.DeepEqual(props, tc.expProps) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"%#v.LoadProperties(ctx, %t) produced result:\n" +
					"      %#v, %#v, %s\n" +
					"want: %#v, %#v, %s",
				tc.profile, tc.force,
				&profile, props, p(err),
				tc.expProfile, tc.expProps, p(tc.expErr),
			)
		}
	}
}

func TestProfile_LoadProperties_ContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx, _ := context.WithCancel(context.Background())
	ct := CtxStoreTransport{}

	client.Transport = &ct

	profile := Profile {ID: "dummy"}
	profile.LoadProperties(ctx, true)

	if ct.Context != ctx {
		t.Error("Profile{ID: \"dummy\"}.LoadProperties(ctx, true) didn't pass context to underlying http.Client")
	}
}

var testProperties_SkinReaderInput = [...] struct{
	props *Properties
	transport http.RoundTripper
	expTexture []byte
	expErr error
} {
	{
		props: &Properties{
			SkinURL: "",
			Model: Steve,
		},
		transport: http.NewFileTransport(http.Dir("testdata")),
		expTexture: (func() []byte { b, _ := ioutil.ReadFile("testdata/SkinTemplates/steve.png"); return b})(),
	},
	{
		props: &Properties{
			SkinURL: "",
			Model: Alex,
		},
		transport: http.NewFileTransport(http.Dir("testdata")),
		expTexture: (func() []byte { b, _ := ioutil.ReadFile("testdata/SkinTemplates/alex.png"); return b})(),
	},
	{
		props: &Properties{
			SkinURL: "",
			Model: Model(99),
		},
		transport: nil,
		expTexture: nil,
		expErr: errors.New("SkinReader() encountered unknown Model"),
	},
	{
		props: &Properties{
			SkinURL: "http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e",
		},
		transport: http.NewFileTransport(http.Dir("testdata")),
		expTexture: (func() []byte { b, _ := ioutil.ReadFile("testdata/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e"); return b})(),
	},
	{
		props: &Properties{
			SkinURL: "://",
		},
		transport: nil,
		expErr: &url.Error{
			Op: "parse",
			URL: "://",
			Err: errors.New("missing protocol scheme"),
		},
	},
	{
		props: &Properties{
			SkinURL: alexSkinURL,
		},
		transport: errorTransport{testError},
		expErr: &url.Error{
			Op: "Get",
			URL: alexSkinURL,
			Err: testError,
		},
	},
	{
		props: &Properties{
			SkinURL: "http://example.com/does/not/exist.png",
		},
		transport: http.NewFileTransport(http.Dir("testdata")),
		expErr: &url.Error{
			Op: "Get",
			URL: "http://example.com/does/not/exist.png",
			Err: &internal.ErrFailedRequest{StatusCode: 404},
		},
	},
}

func TestProperties_SkinReader(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testProperties_SkinReaderInput {
		var buf bytes.Buffer
		client.Transport = tc.transport

		reader, err := tc.props.SkinReader(context.Background())
		if reader != nil {
			buf.ReadFrom(reader)
		}
		texture := buf.Bytes()

		if !reflect.DeepEqual(texture, tc.expTexture) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"%#v.SkinReader(ctx)\n" +
					" was: %#v, %s\n" +
					"want: %#v, %s",
				tc.props,
				texture, p(err),
				tc.expTexture, p(tc.expErr),
			)
		}
	}
}

func TestProperties_SkinReader_ContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx, _ := context.WithCancel(context.Background())
	ct := CtxStoreTransport{}

	client.Transport = &ct

	props := Properties {
		SkinURL: "http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e",
	}
	props.SkinReader(ctx)

	if ct.Context != ctx {
		t.Error("Properties{SkinURL: ...}.LoadProperties(ctx) didn't pass context to underlying http.Client")
	}
}

var testProperties_CapeReaderInput = [...] struct{
	props *Properties
	transport http.RoundTripper
	expTexture []byte
	expErr error
} {
	{
		props: &Properties{
			CapeURL: "",
		},
		transport: nil,
		expTexture: nil,
		expErr: ErrNoCape,
	},
	{
		props: &Properties{
			CapeURL: "http://textures.minecraft.net/texture/ec80a225b145c812a6ef1ca29af0f3ebf02163874d1a66e53bac99965225e0",
		},
		transport: http.NewFileTransport(http.Dir("testdata")),
		expTexture: (func() []byte { b, _ := ioutil.ReadFile("testdata/texture/ec80a225b145c812a6ef1ca29af0f3ebf02163874d1a66e53bac99965225e0"); return b})(),
	},
	{
		props: &Properties{
			CapeURL: "://",
		},
		transport: nil,
		expErr: &url.Error{
			Op: "parse",
			URL: "://",
			Err: errors.New("missing protocol scheme"),
		},
	},
	{
		props: &Properties{
			CapeURL: alexSkinURL,
		},
		transport: errorTransport{testError},
		expErr: &url.Error{
			Op: "Get",
			URL: alexSkinURL,
			Err: testError,
		},
	},
	{
		props: &Properties{
			CapeURL: "http://example.com/does/not/exist.png",
		},
		transport: http.NewFileTransport(http.Dir("testdata")),
		expErr: &url.Error{
			Op: "Get",
			URL: "http://example.com/does/not/exist.png",
			Err: &internal.ErrFailedRequest{StatusCode: 404},
		},
	},
}

func TestProperties_CapeReader(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	for _, tc := range testProperties_CapeReaderInput {
		var buf bytes.Buffer
		client.Transport = tc.transport

		reader, err := tc.props.CapeReader(context.Background())
		if reader != nil {
			buf.ReadFrom(reader)
		}
		texture := buf.Bytes()

		if !reflect.DeepEqual(texture, tc.expTexture) || !reflect.DeepEqual(err, tc.expErr) {
			t.Errorf(
				"%#v.CapeReader(ctx)\n" +
					" was: %#v, %s\n" +
					"want: %#v, %s",
				tc.props,
				texture, p(err),
				tc.expTexture, p(tc.expErr),
			)
		}
	}
}

func TestProperties_CapeReader_ContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx, _ := context.WithCancel(context.Background())
	ct := CtxStoreTransport{}

	client.Transport = &ct

	props := Properties {
		CapeURL: "http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e",
	}
	props.CapeReader(ctx)

	if ct.Context != ctx {
		t.Error("Properties{CapeURL: ...}.LoadProperties(ctx) didn't pass context to underlying http.Client")
	}
}