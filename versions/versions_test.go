package versions

import (
	"testing"
)

// Test that Load succeeds and that all returned Versions data is populated.
func TestLoad(t *testing.T) {

	vs, err := Load()
	if err != nil {

		t.Fatalf("Load() failed to fetch a version listing: %s", err)
	}

	// Verify integrity of vs.Latest.Snapshot
	if s := vs.Latest.Snapshot; s == "" {

		t.Errorf("Load().Latest.Snapshot = %q; want non-empty", "")

	} else if _, ok := vs.Versions[s]; !ok {

		t.Errorf("Load().Versions contained no information for Load().Latest.Snapshot")

	} else if vt := vs.Versions[s].Type; vt != Snapshot {

		t.Errorf("Load().Latest.Snapshot denoted a version of type %s, not %s", vt, Snapshot)
	}

	// Verify integrity of vs.Latest.Release
	if r := vs.Latest.Release; r == "" {

		t.Errorf("Load().Latest.Release = %q; want non-empty", "")

	} else if _, ok := vs.Versions[r]; !ok {

		t.Errorf("Load().Versions contained no information for Load().Latest.Release")

	} else if vt := vs.Versions[r].Type; vt != Release {

		t.Errorf("Load().Latest.Release denoted a version of type %s, not %s", vt, Release)
	}

	// Verify every entry of vs.Versions
	for id, v := range vs.Versions {

		if id != v.ID {

			t.Errorf("Load().Versions[%q].ID = %q; want %q", id, v.ID, id)
		}

		if v.ID == "" {

			t.Errorf("Load().Versions[%q].ID = %q; want non-empty", id, "")
		}

		if v.Released.IsZero() {

			t.Errorf("Load().Versions[%q].Released = %s; want non-zero", id, v.Released)
		}

		if v.Type == "" {

			t.Errorf("Load().Versions[%q].Type = %s; want non-zero value", id, v.Type)
		}
	}
}
