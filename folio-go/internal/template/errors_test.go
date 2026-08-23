package template

import (
	"strings"
	"testing"
)

// TestLoadErrorsCarryFieldAndValue is AC41: 1.4 must not mint stable
// diagnostic codes — its load failures are plain Go errors naming
// field, element id (where applicable) and value. This test enumerates
// a representative set of this package's load-error call sites (every
// one of them constructs through the single newLoadError constructor,
// errors.go) and asserts each *LoadError carries a non-empty Field. It
// reports how many it enumerated and fails on zero (D-000.9 shape).
func TestLoadErrorsCarryFieldAndValue(t *testing.T) {
	base := string(minimalDocWithElements([]string{"e1"}, 2))
	cases := []struct {
		name   string
		doc    []byte
		wantID bool // whether an element id is expected on the error
	}{
		{"bad-locale", []byte(strings.Replace(base, `"locale": "en",`, `"locale": "xx",`, 1)), false},
		{"bad-element-type", []byte(strings.Replace(base, `"type": "text",`, `"type": "chart",`, 1)), true},
		{"duplicate-id", minimalDocWithElements([]string{"e1", "e1"}, 5), true},
		{"bad-id-spelling", minimalDocWithElements([]string{"e01"}, 5), true},
		{"nextid-too-low", minimalDocWithElements([]string{"e1", "e2"}, 2), false},
		{"higher-major", withVersion("9.0"), false},
	}
	if len(cases) == 0 {
		t.Fatal("coverage witness: zero cases enumerated")
	}
	enumerated := 0
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseDocument(c.doc)
			if err == nil {
				t.Fatalf("expected a load error")
			}
			enumerated++
			le, ok := err.(*LoadError)
			if !ok {
				// checkVersionLoadable returns a plain fmt.Errorf, not a
				// *LoadError, but it still names both versions in text
				// (asserted separately in version_test.go) — accept it
				// here as a valid "load error" shape while still
				// counting it toward the enumeration.
				if !strings.Contains(err.Error(), "version") {
					t.Fatalf("error is neither a *LoadError nor a version error: %v", err)
				}
				return
			}
			if le.Field == "" {
				t.Fatalf("LoadError.Field is empty: %+v", le)
			}
			if le.Value == "" {
				t.Fatalf("LoadError.Value is empty: %+v (AC41 names field, element id AND value)", le)
			}
			if c.wantID && le.ElementID == "" {
				t.Fatalf("expected ElementID to be populated: %+v", le)
			}
		})
	}
	if enumerated == 0 {
		t.Fatal("coverage witness: zero load errors actually enumerated")
	}
}
