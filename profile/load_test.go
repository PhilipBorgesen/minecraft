package profile

import (
	"errors"
	"reflect"
	"testing"
	"time"
	"io"
	"encoding/base64"
	"net/http"
	"context"
	"github.com/PhilipBorgesen/minecraft/internal"
	"net/url"
)

var testBuildProfileInput = [...]struct {
	m          map[string]interface{}
	demoErr    error
	expProfile *Profile
	expErr     error
}{
	{
		m: map[string]interface{}{
			"demo": true,
		},
		demoErr:    testError,
		expProfile: nil,
		expErr:     testError,
	},
	{
		m: map[string]interface{}{
			"id":   "087cc153c3434ff7ac497de1569affa1",
			"name": "Nergalic",
			"demo": true,
		},
		demoErr:    testError,
		expProfile: nil,
		expErr:     testError,
	},
	{
		m: map[string]interface{}{
			"id":   "cabefc91b5df4c87886a6c604da2e46f",
			"name": "AxeLaw",
			"demo": false,
		},
		demoErr: testError,
		expProfile: &Profile{
			ID:   "cabefc91b5df4c87886a6c604da2e46f",
			Name: "AxeLaw",
		},
		expErr: nil,
	},
	{
		m: map[string]interface{}{
			"id":   "087cc153c3434ff7ac497de1569affa1",
			"name": "Nergalic",
		},
		demoErr: testError,
		expProfile: &Profile{
			ID:   "087cc153c3434ff7ac497de1569affa1",
			Name: "Nergalic",
		},
		expErr: nil,
	},
	{
		m: map[string]interface{}{
			"id":     "087cc153c3434ff7ac497de1569affa1",
			"name":   "Nergalic",
			"legacy": false,
		},
		demoErr: testError,
		expProfile: &Profile{
			ID:   "087cc153c3434ff7ac497de1569affa1",
			Name: "Nergalic",
		},
		expErr: nil,
	},
	{
		m: map[string]interface{}{
			"id":     "087cc153c3434ff7ac497de1569affa1",
			"name":   "Nergalic",
			"legacy": true,
		},
		demoErr: testError,
		expProfile: &Profile{
			ID:          "087cc153c3434ff7ac497de1569affa1",
			Name:        "Nergalic",
			NameHistory: emptyHist,
		},
		expErr: nil,
	},
}

func TestBuildProfile(t *testing.T) {
	for _, tc := range testBuildProfileInput {
		profile, err := buildProfile(tc.m, tc.demoErr)
		if !reflect.DeepEqual(profile, tc.expProfile) || err != tc.expErr {
			t.Errorf(
				"\n"+
					"buildProfile(%#v, %s)\n"+
					"was  %#v, %s\n"+
					"want %#v, %s",
				tc.m, tc.demoErr,
				profile, p(err),
				tc.expProfile, p(tc.expErr),
			)
		}
	}
}

var testBuildHistoryInput = [...]struct {
	arr     []interface{}
	expName string
	expHist []PastName
}{
	{
		arr:     nil,
		expName: "",
		expHist: nil,
	},
	{
		arr:     []interface{}{},
		expName: "",
		expHist: nil,
	},
	{
		arr: []interface{}{
			map[string]interface{}{
				"name": "A",
			},
		},
		expName: "A",
		expHist: emptyHist,
	},
	{
		arr: []interface{}{
			map[string]interface{}{
				"name": "B",
			},
			map[string]interface{}{
				"name":        "A",
				"changedToAt": float64(1423047705000),
			},
		},
		expName: "A",
		expHist: []PastName{
			{
				Name:  "B",
				Until: msToTime(1423047705000),
			},
		},
	},
	{
		arr: []interface{}{
			map[string]interface{}{
				"name": "C",
			},
			map[string]interface{}{
				"name":        "B",
				"changedToAt": float64(1000047705000),
			},
			map[string]interface{}{
				"name":        "A",
				"changedToAt": float64(1423047705000),
			},
		},
		expName: "A",
		expHist: []PastName{
			{
				Name:  "B",
				Until: msToTime(1423047705000),
			},
			{
				Name:  "C",
				Until: msToTime(1000047705000),
			},
		},
	},
}

func TestBuildHistory(t *testing.T) {
	for _, tc := range testBuildHistoryInput {
		name, hist := buildHistory(tc.arr)
		if name != tc.expName || !reflect.DeepEqual(hist, tc.expHist) {
			t.Errorf(
				"\n"+
					"buildHistory(%#v)\n"+
					"was  %q, %#v\n"+
					"want %q, %#v",
				tc.arr,
				name, hist,
				tc.expName, tc.expHist,
			)
		}
	}
}

func TestMsToTime(t *testing.T) {
	const s2ms int64 = 1000
	const ms2ns int64 = 1000 * 1000

	t1 := time.Unix(1423047705, 123*ms2ns)
	ms := t1.Unix()*s2ms + int64(t1.Nanosecond())/ms2ns
	t2 := msToTime(ms)

	if !t1.Equal(t2) {
		t.Errorf("msToTime(%s) was %s; want %s", t1, t2, t1)
	}
}

func TestIsEven(t *testing.T) {
	for dec, hex := range [...]uint8{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'} {
		expEven := false
		if dec%2 == 0 {
			expEven = true
		}
		if even := isEven(hex); even != expEven {
			t.Errorf("isEven(%q) was %t; want %t", hex, even, expEven)
		}
	}
}

var testDefaultModelInput = [...]struct {
	uuid     string
	expModel Model
}{
	{
		uuid:     "087cc153c3434ff7ac497de1569affa1", // Nergalic
		expModel: Steve,
	},
	{
		uuid:     "3fe136c0cd434f7783fc94b9b86eed6d", // Feathertail
		expModel: Alex,
	},
}

func TestDefaultModel(t *testing.T) {
	for _, tc := range testDefaultModelInput {
		if model := defaultModel(tc.uuid); model != tc.expModel {
			t.Errorf("defaultModel(%q) was %s; want %s", tc.uuid, model, tc.expModel)
		}
	}
}

var testPopulateTexturesInput = [...]struct {
	enc string
	expProperties *Properties
	expErr error
}{
	{
		enc: "!notBase64",
		expProperties: &Properties{},
		expErr: base64.CorruptInputError(0),
	},
	{
		enc: "",
		expProperties: &Properties{},
		expErr: io.EOF,
	},
	{
		enc: "eyJ0aW1lc3RhbXAiOjE0OTM4NzUyMDcyMDYsInByb2ZpbGVJZCI6ImQ5MGI2OGJjODE3MjQzMjlhMDQ3ZjExODZkY2Q0MzM2IiwicHJvZmlsZU5hbWUiOiJha3Jvbm1hbjEiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvMzE3YTQxYzdhMzE1ODIxZTM2ZWU4YzdjOGMzOTQ3MTc0ZTQxYjU1MmViNDE2OGI3MTI3YzJkNWI4MmZhY2UwIn0sIkNBUEUiOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS9lYzgwYTIyNWIxNDVjODEyYTZlZjFjYTI5YWYwZjNlYmYwMjE2Mzg3NGQxYTY2ZTUzYmFjOTk5NjUyMjVlMCJ9fX0=",
		expProperties: &Properties{
			SkinURL:"http://textures.minecraft.net/texture/317a41c7a315821e36ee8c7c8c3947174e41b552eb4168b7127c2d5b82face0",
			CapeURL:"http://textures.minecraft.net/texture/ec80a225b145c812a6ef1ca29af0f3ebf02163874d1a66e53bac99965225e0",
			Model: Steve,
		},
	},
	{
		enc: "eyJ0aW1lc3RhbXAiOjE0OTM4NzUwMTAxODEsInByb2ZpbGVJZCI6ImNhYmVmYzkxYjVkZjRjODc4ODZhNmM2MDRkYTJlNDZmIiwicHJvZmlsZU5hbWUiOiJBeGVMYXciLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvZDcyZDliMDBmM2Y2NDk0NjA3ZDIwZTU1N2U3ZjFiMjc2ZTczODZiYmZlNjk2NDliZTg3YmVjOGM0NDhkIn19fQ==",
		expProperties: &Properties{
			SkinURL: "http://textures.minecraft.net/texture/d72d9b00f3f6494607d20e557e7f1b276e7386bbfe69649be87bec8c448d",
			CapeURL: "",
			Model: Steve,
		},
	},
	{
		enc: "eyJ0aW1lc3RhbXAiOjE0OTM4NzcwNzE4NzAsInByb2ZpbGVJZCI6IjM2ZGNjN2E4M2NhMDQzNzI4NjU3ODI4MTg1ODZjYjJjIiwicHJvZmlsZU5hbWUiOiJTYWt1cmFCZWxsIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJtZXRhZGF0YSI6eyJtb2RlbCI6InNsaW0ifSwidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS9iYzJlMTc1MGMwNGMxNWU1YjdiMWYyYmFmZmEzNzEyMTM0ZmFmNzc0NGM0MTcyMzUxN2I1OTYwOGU0Yzk1NjgifX19",
		expProperties: &Properties{
			SkinURL: "http://textures.minecraft.net/texture/bc2e1750c04c15e5b7b1f2baffa3712134faf7744c41723517b59608e4c9568",
			CapeURL: "",
			Model: Alex,
		},
	},
	{
		enc: "eyJ0aW1lc3RhbXAiOjE0OTM4Nzc4NTc0NTYsInByb2ZpbGVJZCI6ImVjNTYxNTM4ZjNmZDQ2MWRhZmY1MDg2YjIyMTU0YmNlIiwicHJvZmlsZU5hbWUiOiJBbGV4IiwidGV4dHVyZXMiOnt9fQ==",
		expProperties: &Properties {
			SkinURL: "",
			CapeURL: "",
			Model: Steve,
		},
	},
}

func TestPopulateTextures(t *testing.T) {
	for _, tc := range testPopulateTexturesInput {
		var p Properties
		err := populateTextures(tc.enc, &p)
		if !reflect.DeepEqual(&p, tc.expProperties) || err != tc.expErr {
			t.Errorf(
				"populateTextures(%q, Properties) produced result:\n" +
				"      %#v, %s\n" +
				"want: %#v, %s",
				tc.enc,
				&p, err,
				tc.expProperties, tc.expErr,
			)
		}
	}
}

var testBuildPropertiesInput = [...]struct {
	props []interface{}
	expProperties *Properties
	expErr error
}{
	{
		props: []interface{} {},
		expProperties: &Properties{},
	},
	{
		props: []interface{} {
			map[string]interface{} {
				"name": "nonExistingProperty",
				"value": "dummy",
			},
		},
		expProperties: &Properties{},
	},
	{
		props: []interface{} {
			map[string]interface{} {
				"name": "textures",
				"value": "!notBase64",
			},
		},
		expProperties: nil,
		expErr: base64.CorruptInputError(0),
	},
	{
		props: []interface{} {
			map[string]interface{} {
				"name": "textures",
				"value": "eyJ0aW1lc3RhbXAiOjE0OTM4NzUyMDcyMDYsInByb2ZpbGVJZCI6ImQ5MGI2OGJjODE3MjQzMjlhMDQ3ZjExODZkY2Q0MzM2IiwicHJvZmlsZU5hbWUiOiJha3Jvbm1hbjEiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvMzE3YTQxYzdhMzE1ODIxZTM2ZWU4YzdjOGMzOTQ3MTc0ZTQxYjU1MmViNDE2OGI3MTI3YzJkNWI4MmZhY2UwIn0sIkNBUEUiOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS9lYzgwYTIyNWIxNDVjODEyYTZlZjFjYTI5YWYwZjNlYmYwMjE2Mzg3NGQxYTY2ZTUzYmFjOTk5NjUyMjVlMCJ9fX0=",
			},
		},
		expProperties: &Properties{
			SkinURL:"http://textures.minecraft.net/texture/317a41c7a315821e36ee8c7c8c3947174e41b552eb4168b7127c2d5b82face0",
			CapeURL:"http://textures.minecraft.net/texture/ec80a225b145c812a6ef1ca29af0f3ebf02163874d1a66e53bac99965225e0",
			Model: Steve,
		},
	},
	// Other cases:
	// - Multiple properties
	// - Same property appearing twice
	// -
}

func TestBuildProperties(t *testing.T) {
	for _, tc := range testBuildPropertiesInput {
		ps, err := buildProperties(tc.props)
		if !reflect.DeepEqual(ps, tc.expProperties) || err != tc.expErr {
			t.Errorf(
				"buildProperties(%#v)\n" +
					"was:  %#v, %s\n" +
					"want: %#v, %s",
				tc.props,
				ps, err,
				tc.expProperties, tc.expErr,
			)
		}
	}
}

var testLoadInput = [...]struct {
	username string
	transport http.RoundTripper
	expProfile *Profile
	expErr error
} {
	{
		username: "",
		transport: nil,
		expProfile: nil,
		expErr: ErrNoSuchProfile,
	},
	{
		username: "doesNotExist",
		transport: errorTransport{
			&internal.FailedRequestError{
				StatusCode: 204,
			},
		},
		expProfile: nil,
		expErr: ErrNoSuchProfile,
	},
	{
		username: "unexpectedFormat",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr: &url.Error{
			Op: "Parse",
			URL: "https://api.mojang.com/users/profiles/minecraft/unexpectedFormat",
			Err: internal.ErrUnknownFormat,
		},
	},
	{
		username: "nergalic",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile {
			Name: "Nergalic",
			ID: "087cc153c3434ff7ac497de1569affa1",
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
				"Load(ctx, %q)\n" +
					" was: %#v, %s\n" +
					"want: %#v, %s",
				tc.username,
				profile, p(err),
				tc.expProfile, p(tc.expErr),
			)
		}
	}
}

func TestLoad_ContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx, _ := context.WithCancel(context.Background())
	ct := CtxStoreTransport{}

	client.Transport = &ct
	Load(ctx, "nergalic")

	if ct.Context != ctx {
		t.Error("Load(ctx, \"nergalic\") didn't pass context to underlying http.Client")
	}
}

var testLoadAtTimeInput = [...]struct {
	username string
	time time.Time
	transport http.RoundTripper
	expProfile *Profile
	expErr error
} {
	{
		username: "",
		time: time.Unix(0, 0),
		transport: nil,
		expProfile: nil,
		expErr: ErrNoSuchProfile,
	},
	{
		username: "doesNotExist",
		time: time.Unix(0, 0),
		transport: errorTransport{
			&internal.FailedRequestError{
				StatusCode: 204,
			},
		},
		expProfile: nil,
		expErr: ErrNoSuchProfile,
	},
	{
		username: "unexpectedFormat",
		time: time.Unix(1337, 564),
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr: &url.Error{
			Op: "Parse",
			URL: "https://api.mojang.com/users/profiles/minecraft/unexpectedFormat?at=1337",
			Err: internal.ErrUnknownFormat,
		},
	},
	{
		username: "GeneralSezuan",
		time: time.Unix(0, 0),
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile {
			Name: "Nergalic",
			ID: "087cc153c3434ff7ac497de1569affa1",
		},
		expErr: nil,
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
				"LoadAtTime(ctx, %q, %s)\n" +
					" was: %#v, %s\n" +
					"want: %#v, %s",
				tc.username, tc.time,
				profile, p(err),
				tc.expProfile, p(tc.expErr),
			)
		}
	}
}

func TestLoadAtTime_ContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx, _ := context.WithCancel(context.Background())
	ct := CtxStoreTransport{}

	client.Transport = &ct
	LoadAtTime(ctx, "nergalic", time.Now())

	if ct.Context != ctx {
		t.Error("LoadAtTime(ctx, \"nergalic\", time.Now()) didn't pass context to underlying http.Client")
	}
}

var testLoadWithNameHistoryInput = [...]struct {
	id string
	transport http.RoundTripper
	expProfile *Profile
	expErr error
} {
	{
		id: "",
		transport: nil,
		expProfile: nil,
		expErr: ErrNoSuchProfile,
	},
	{
		id: "00000000000000000000000000000000", // Doesn't exist
		transport: errorTransport{
			&internal.FailedRequestError{
				StatusCode: 204,
			},
		},
		expProfile: nil,
		expErr: ErrNoSuchProfile,
	},
	{
		id: "unexpectedFormat",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr: &url.Error{
			Op: "Parse",
			URL: "https://api.mojang.com/user/profiles/unexpectedFormat/names",
			Err: internal.ErrUnknownFormat,
		},
	},
	{
		id: "087cc153c3434ff7ac497de1569affa1",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile {
			Name: "Nergalic",
			ID: "087cc153c3434ff7ac497de1569affa1",
			NameHistory: []PastName {
				{
					Name: "GeneralSezuan",
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
				"LoadWithNameHistory(ctx, %q)\n" +
					" was: %#v, %s\n" +
					"want: %#v, %s",
				tc.id,
				profile, p(err),
				tc.expProfile, p(tc.expErr),
			)
		}
	}
}

func TestLoadWithNameHistory_ContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx, _ := context.WithCancel(context.Background())
	ct := CtxStoreTransport{}

	client.Transport = &ct
	LoadByID(ctx, "dummy") // Wrapper method used to test that as well

	if ct.Context != ctx {
		t.Error("LoadWithNameHistory(ctx, \"dummy\") didn't pass context to underlying http.Client")
	}
}

var testLoadWithPropertiesInput = [...]struct {
	id string
	transport http.RoundTripper
	expProfile *Profile
	expErr error
} {
	{
		id: "",
		transport: nil,
		expProfile: nil,
		expErr: ErrNoSuchProfile,
	},
	{
		id: "00000000000000000000000000000000", // Doesn't exist
		transport: errorTransport{
			&internal.FailedRequestError{
				StatusCode: 204,
			},
		},
		expProfile: nil,
		expErr: ErrNoSuchProfile,
	},
	{
		id: "fictiveDemo",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr: ErrNoSuchProfile,
	},
	{
		id: "noSkinAndBadUUID",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr: &url.Error{
			Op: "Parse",
			URL: "https://sessionserver.mojang.com/session/minecraft/profile/noSkinAndBadUUID",
			Err: internal.ErrUnknownFormat,
		},
	},
	{
		id: "badProperties",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: nil,
		expErr: &url.Error{
			Op: "Parse",
			URL: "https://sessionserver.mojang.com/session/minecraft/profile/badProperties",
			Err: base64.CorruptInputError(0),
		},
	},
	{
		id: "tooManyRequests",
		transport: statusOverrideTransport{
			status: 429,
			transport: http.NewFileTransport(http.Dir("testdata")),
		},
		expProfile: nil,
		expErr: ErrTooManyRequests,
	},
	{
		id: "087cc153c3434ff7ac497de1569affa1",
		transport: http.NewFileTransport(http.Dir("testdata")),
		expProfile: &Profile {
			Name: "Nergalic",
			ID: "087cc153c3434ff7ac497de1569affa1",
			Properties: &Properties{
				SkinURL: "http://textures.minecraft.net/texture/5b40f251f7c8db60943495db6bf54353102d6cad20d2299d5f973f36b4f3677e",
				CapeURL: "",
				Model: Steve,
			},
		},
		expErr: nil,
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
				"LoadWithProperties(ctx, %q)\n" +
					" was: %#v, %s\n" +
					"want: %#v, %s",
				tc.id,
				profile, p(err),
				tc.expProfile, p(tc.expErr),
			)
		}
	}
}

func TestLoadWithProperties_ContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx, _ := context.WithCancel(context.Background())
	ct := CtxStoreTransport{}

	client.Transport = &ct
	LoadWithProperties(ctx, "dummy")

	if ct.Context != ctx {
		t.Error("LoadWithProperties(ctx, \"dummy\") didn't pass context to underlying http.Client")
	}
}

var testLoadManyInput = [...]struct {
	ids []string
	transport http.RoundTripper
	expProfiles []*Profile
	expErr error
} {
	{
		ids: []string{},
		transport: nil,
		expProfiles: nil,
		expErr: nil,
	},
	{
		ids: []string{""},
		transport: nil,
		expProfiles: nil,
		expErr: nil,
	},
	{
		ids: make([]string, LoadManyMaxSize + 1, LoadManyMaxSize + 1),
		transport: nil,
		expProfiles: nil,
		expErr: ErrMaxSizeExceeded{LoadManyMaxSize + 1},
	},
	{
		ids: []string{"dummy"},
		transport: http.NewFileTransport(http.Dir("testdata/LoadMany/unexpectedFormat")),
		expProfiles: nil,
		expErr: &url.Error{
			Op: "Parse",
			URL: "https://api.mojang.com/profiles/minecraft",
			Err: internal.ErrUnknownFormat,
		},
	},
	{
		ids: []string{"nergalic", "AxeLaw", "demo", "doesNotExist"},
		transport: http.NewFileTransport(http.Dir("testdata/LoadMany/success")),
		expProfiles: []*Profile{
			{
				ID: "087cc153c3434ff7ac497de1569affa1",
				Name: "Nergalic",
			},
			{
				ID: "cabefc91b5df4c87886a6c604da2e46f",
				Name: "AxeLaw",
				NameHistory: emptyHist,
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
				"LoadMany(ctx, %q)\n" +
					" was: %#v, %s\n" +
					"want: %#v, %s",
				tc.ids,
				profiles, p(err),
				tc.expProfiles, p(tc.expErr),
			)
		}
	}
}

func TestLoadMany_ContextUsed(t *testing.T) {
	origTransport := client.Transport
	defer func() { client.Transport = origTransport }()

	ctx, _ := context.WithCancel(context.Background())
	ct := CtxStoreTransport{}

	client.Transport = &ct
	LoadMany(ctx, "dummy")

	if ct.Context != ctx {
		t.Error("LoadMany(ctx, \"dummy\") didn't pass context to underlying http.Client")
	}
}

///////////////////

var testError = errors.New("testError")

func p(x interface{}) interface{} {
	if x == nil {
		return "<nil>"
	} else {
		return x
	}
}

///////////////////

type statusOverrideTransport struct {
	status int
	transport http.RoundTripper
}

func (sot statusOverrideTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = sot.transport.RoundTrip(req)
	resp.StatusCode = sot.status
	return
}

///////////////////

type errorTransport struct {
	err error
}

func (et errorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, et.err
}

///////////////////

type CtxStoreTransport struct {
	Context context.Context
}

func (ct *CtxStoreTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ct.Context = req.Context()
	return nil, errors.New("RoundTrip was called")
}