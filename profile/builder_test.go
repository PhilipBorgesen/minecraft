package profile

import (
	"encoding/base64"
	"io"
	"reflect"
	"testing"
	"time"
)

var testFillProfileInput = [...]struct {
	p          Profile
	m          map[string]interface{}
	expProfile Profile
	isDemo     bool
}{
	{
		p: Profile{ID: "x", Name: "y"},
		m: map[string]interface{}{
			"demo": true,
		},
		expProfile: Profile{ID: "x", Name: "y"},
		isDemo:     true,
	},
	{
		p: Profile{ID: "x", Name: "y"},
		m: map[string]interface{}{
			"id":   "087cc153c3434ff7ac497de1569affa1",
			"name": "Nergalic",
			"demo": true,
		},
		expProfile: Profile{ID: "x", Name: "y"},
		isDemo:     true,
	},
	{
		p: Profile{ID: "x", Name: "y"},
		m: map[string]interface{}{
			"id":   "cabefc91b5df4c87886a6c604da2e46f",
			"name": "AxeLaw",
			"demo": false,
		},
		expProfile: Profile{
			ID:   "cabefc91b5df4c87886a6c604da2e46f",
			Name: "AxeLaw",
		},
	},
	{
		p: Profile{ID: "x", Name: "y"},
		m: map[string]interface{}{
			"id":   "087cc153c3434ff7ac497de1569affa1",
			"name": "Nergalic",
		},
		expProfile: Profile{
			ID:   "087cc153c3434ff7ac497de1569affa1",
			Name: "Nergalic",
		},
	},
	{
		p: Profile{},
		m: map[string]interface{}{
			"id":     "087cc153c3434ff7ac497de1569affa1",
			"name":   "Nergalic",
			"legacy": false,
		},
		expProfile: Profile{
			ID:   "087cc153c3434ff7ac497de1569affa1",
			Name: "Nergalic",
		},
	},
	{
		p: Profile{},
		m: map[string]interface{}{
			"id":     "087cc153c3434ff7ac497de1569affa1",
			"name":   "Nergalic",
			"legacy": true,
		},
		expProfile: Profile{
			ID:          "087cc153c3434ff7ac497de1569affa1",
			Name:        "Nergalic",
			NameHistory: emptyHist,
		},
	},
	{ // Existing name history not overwritten
		p: Profile{NameHistory: make([]PastName, 1)},
		m: map[string]interface{}{
			"id":     "087cc153c3434ff7ac497de1569affa1",
			"name":   "Nergalic",
			"legacy": true,
		},
		expProfile: Profile{
			ID:          "087cc153c3434ff7ac497de1569affa1",
			Name:        "Nergalic",
			NameHistory: make([]PastName, 1),
		},
	},
}

func TestFillProfile(t *testing.T) {
	for _, tc := range testFillProfileInput {
		profile := tc.p
		notDemo := fillProfile(&profile, tc.m)
		if !reflect.DeepEqual(profile, tc.expProfile) || notDemo != !tc.isDemo {
			t.Errorf(
				"\n"+
					"fillProfile(%#v, %#v)\n"+
					"was  %#v, %t\n"+
					"want %#v, %t",
				tc.p, tc.m,
				profile, notDemo,
				tc.expProfile, !tc.isDemo,
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
	enc           string
	expProperties *Properties
	expErr        error
}{
	{
		enc:           "!notBase64",
		expProperties: &Properties{},
		expErr:        base64.CorruptInputError(0),
	},
	{
		enc:           "",
		expProperties: &Properties{},
		expErr:        io.EOF,
	},
	{
		enc: "eyJ0aW1lc3RhbXAiOjE0OTM4NzUyMDcyMDYsInByb2ZpbGVJZCI6ImQ5MGI2OGJjODE3MjQzMjlhMDQ3ZjExODZkY2Q0MzM2IiwicHJvZmlsZU5hbWUiOiJha3Jvbm1hbjEiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvMzE3YTQxYzdhMzE1ODIxZTM2ZWU4YzdjOGMzOTQ3MTc0ZTQxYjU1MmViNDE2OGI3MTI3YzJkNWI4MmZhY2UwIn0sIkNBUEUiOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS9lYzgwYTIyNWIxNDVjODEyYTZlZjFjYTI5YWYwZjNlYmYwMjE2Mzg3NGQxYTY2ZTUzYmFjOTk5NjUyMjVlMCJ9fX0=",
		expProperties: &Properties{
			SkinURL: "http://textures.minecraft.net/texture/317a41c7a315821e36ee8c7c8c3947174e41b552eb4168b7127c2d5b82face0",
			CapeURL: "http://textures.minecraft.net/texture/ec80a225b145c812a6ef1ca29af0f3ebf02163874d1a66e53bac99965225e0",
			Model:   Steve,
		},
	},
	{
		enc: "eyJ0aW1lc3RhbXAiOjE0OTM4NzUwMTAxODEsInByb2ZpbGVJZCI6ImNhYmVmYzkxYjVkZjRjODc4ODZhNmM2MDRkYTJlNDZmIiwicHJvZmlsZU5hbWUiOiJBeGVMYXciLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvZDcyZDliMDBmM2Y2NDk0NjA3ZDIwZTU1N2U3ZjFiMjc2ZTczODZiYmZlNjk2NDliZTg3YmVjOGM0NDhkIn19fQ==",
		expProperties: &Properties{
			SkinURL: "http://textures.minecraft.net/texture/d72d9b00f3f6494607d20e557e7f1b276e7386bbfe69649be87bec8c448d",
			CapeURL: "",
			Model:   Steve,
		},
	},
	{
		enc: "eyJ0aW1lc3RhbXAiOjE0OTM4NzcwNzE4NzAsInByb2ZpbGVJZCI6IjM2ZGNjN2E4M2NhMDQzNzI4NjU3ODI4MTg1ODZjYjJjIiwicHJvZmlsZU5hbWUiOiJTYWt1cmFCZWxsIiwidGV4dHVyZXMiOnsiU0tJTiI6eyJtZXRhZGF0YSI6eyJtb2RlbCI6InNsaW0ifSwidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS9iYzJlMTc1MGMwNGMxNWU1YjdiMWYyYmFmZmEzNzEyMTM0ZmFmNzc0NGM0MTcyMzUxN2I1OTYwOGU0Yzk1NjgifX19",
		expProperties: &Properties{
			SkinURL: "http://textures.minecraft.net/texture/bc2e1750c04c15e5b7b1f2baffa3712134faf7744c41723517b59608e4c9568",
			CapeURL: "",
			Model:   Alex,
		},
	},
	{
		enc: "eyJ0aW1lc3RhbXAiOjE0OTM4Nzc4NTc0NTYsInByb2ZpbGVJZCI6ImVjNTYxNTM4ZjNmZDQ2MWRhZmY1MDg2YjIyMTU0YmNlIiwicHJvZmlsZU5hbWUiOiJBbGV4IiwidGV4dHVyZXMiOnt9fQ==",
		expProperties: &Properties{
			SkinURL: "",
			CapeURL: "",
			Model:   Steve,
		},
	},
}

func TestPopulateTextures(t *testing.T) {
	for _, tc := range testPopulateTexturesInput {
		var p Properties
		err := populateTextures(tc.enc, &p)
		if !reflect.DeepEqual(&p, tc.expProperties) || err != tc.expErr {
			t.Errorf(
				"populateTextures(%q, Properties) produced result:\n"+
					"      %#v, %s\n"+
					"want: %#v, %s",
				tc.enc,
				&p, err,
				tc.expProperties, tc.expErr,
			)
		}
	}
}

var testBuildPropertiesInput = [...]struct {
	props         []interface{}
	expProperties *Properties
	expErr        error
}{
	{
		props:         []interface{}{},
		expProperties: &Properties{},
	},
	{
		props: []interface{}{
			map[string]interface{}{
				"name":  "nonExistingProperty",
				"value": "dummy",
			},
		},
		expProperties: &Properties{},
	},
	{
		props: []interface{}{
			map[string]interface{}{
				"name":  "textures",
				"value": "!notBase64",
			},
		},
		expProperties: nil,
		expErr:        base64.CorruptInputError(0),
	},
	{
		props: []interface{}{
			map[string]interface{}{
				"name":  "textures",
				"value": "eyJ0aW1lc3RhbXAiOjE0OTM4NzUyMDcyMDYsInByb2ZpbGVJZCI6ImQ5MGI2OGJjODE3MjQzMjlhMDQ3ZjExODZkY2Q0MzM2IiwicHJvZmlsZU5hbWUiOiJha3Jvbm1hbjEiLCJ0ZXh0dXJlcyI6eyJTS0lOIjp7InVybCI6Imh0dHA6Ly90ZXh0dXJlcy5taW5lY3JhZnQubmV0L3RleHR1cmUvMzE3YTQxYzdhMzE1ODIxZTM2ZWU4YzdjOGMzOTQ3MTc0ZTQxYjU1MmViNDE2OGI3MTI3YzJkNWI4MmZhY2UwIn0sIkNBUEUiOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS9lYzgwYTIyNWIxNDVjODEyYTZlZjFjYTI5YWYwZjNlYmYwMjE2Mzg3NGQxYTY2ZTUzYmFjOTk5NjUyMjVlMCJ9fX0=",
			},
		},
		expProperties: &Properties{
			SkinURL: "http://textures.minecraft.net/texture/317a41c7a315821e36ee8c7c8c3947174e41b552eb4168b7127c2d5b82face0",
			CapeURL: "http://textures.minecraft.net/texture/ec80a225b145c812a6ef1ca29af0f3ebf02163874d1a66e53bac99965225e0",
			Model:   Steve,
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
				"buildProperties(%#v)\n"+
					"was:  %#v, %s\n"+
					"want: %#v, %s",
				tc.props,
				ps, err,
				tc.expProperties, tc.expErr,
			)
		}
	}
}
