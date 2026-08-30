package template

import (
	"strings"
	"testing"
)

func withVersion(v string) []byte {
	base := minimalDocWithElements(nil, 1)
	return []byte(strings.Replace(string(base), `"version": "1.0"`, `"version": "`+v+`"`, 1))
}

// TestHigherMajorIsLoadError is AC6: a higher MAJOR than the library
// supports is a load error naming the declared version and the
// supported version, and no render is attempted.
//
// The literal moved from "2.0" to "3.0" when Story 7.3 raised
// SupportedMajor to 2 — 2.0 is now a version this library LOADS. The
// test's subject was always "higher than supported", never the literal
// two, so the constant it must stay above is SupportedMajor and the case
// is re-stated one MAJOR above the new ceiling.
func TestHigherMajorIsLoadError(t *testing.T) {
	_, err := ParseDocument(withVersion("3.0"))
	if err == nil {
		t.Fatal("expected a load error for a higher MAJOR version")
	}
	if !strings.Contains(err.Error(), "3.0") || !strings.Contains(err.Error(), SupportedVersion) {
		t.Fatalf("error must name both the declared and supported version, got: %v", err)
	}
}

// TestSupportedMajorIsLoadable is the other half of the move above, and
// it is what keeps that test honest: 2.0 must now LOAD and round-trip
// verbatim, or "3.0 is refused" would be satisfied by a library that
// refuses everything above 1.
func TestSupportedMajorIsLoadable(t *testing.T) {
	d, err := ParseDocument(withVersion("2.0"))
	if err != nil {
		t.Fatalf("2.0 is this library's own ceiling and must load: %v", err)
	}
	if d.Version != "2.0" {
		t.Fatalf("Version = %q, want 2.0", d.Version)
	}
	out, err := SerializeDocument(d)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if !strings.Contains(string(out), `"version": "2.0"`) {
		t.Fatalf("a 2.0 document must round-trip as 2.0:\n%s", out)
	}
}

// TestNewContentIsNeverStampedWithTheLibraryCeiling is D-7.2.1's
// guardrail asserted DIRECTLY rather than, as until now, only indirectly
// through the fixtures.
//
// SupportedVersion is 2.0 from Story 7.3 onward, so the failure this
// guards against is no longer hypothetical bookkeeping: a document that
// carries no version-raising content at all must still serialize 1.0,
// and one that carries only a 1.1 key must still serialize 1.1. Stamping
// either with the library's ceiling would orphan it from every reader
// that could in fact have read it — and, for the first time, would make
// a plain three-element document unreadable to the 1.x readers that
// exist.
func TestNewContentIsNeverStampedWithTheLibraryCeiling(t *testing.T) {
	for _, c := range []struct {
		label string
		style string
		want  string
	}{
		{"no style keys at all", "", "1.0"},
		{"a 1.1 key only", `"lineSpacing": 1.5, `, "1.1"},
		{"a 1.0 alignment", `"align": "center", `, "1.0"},
		{"the 2.0 alignment", `"align": "justify", `, "2.0"},
	} {
		t.Run(c.label, func(t *testing.T) {
			d, err := ParseDocument([]byte(lineSpacingRoundTripDocWithStyle(c.style)))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			out, err := SerializeDocument(d)
			if err != nil {
				t.Fatalf("serialize: %v", err)
			}
			if !strings.Contains(string(out), `"version": "`+c.want+`"`) {
				t.Fatalf("want version %s, got:\n%s", c.want, out)
			}
			if c.want != SupportedVersion && strings.Contains(string(out), `"version": "`+SupportedVersion+`"`) {
				t.Fatalf("the document was stamped with the library ceiling %s rather than the %s its content requires", SupportedVersion, c.want)
			}
		})
	}
}

// TestVersionRoundTripsVerbatim pins D-1.4.13: a 1.0 file round-trips
// as 1.0.
func TestVersionRoundTripsVerbatim_1_0(t *testing.T) {
	b := withVersion("1.0")
	d, err := ParseDocument(b)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if d.Version != "1.0" {
		t.Fatalf("Version = %q, want 1.0", d.Version)
	}
	out, err := SerializeDocument(d)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if !strings.Contains(string(out), `"version": "1.0"`) {
		t.Fatalf("serialized output does not carry version 1.0 verbatim:\n%s", out)
	}
}

// TestHigherMinorLoadsAndRoundTripsVerbatim is AC7: a higher MINOR
// loads, and a synthetic higher-MINOR file round-trips as that same
// MINOR with its unknown keys intact — never lowered to the library's
// own SupportedVersion.
func TestHigherMinorLoadsAndRoundTripsVerbatim(t *testing.T) {
	base := string(minimalDocWithElements(nil, 1))
	synthetic := strings.Replace(base, `"version": "1.0"`, `"version": "1.1"`, 1)
	synthetic = strings.Replace(synthetic, "\"locale\": \"en\",", "\"futureMinorKey\": true,\n  \"locale\": \"en\",", 1)

	d, err := ParseDocument([]byte(synthetic))
	if err != nil {
		t.Fatalf("a higher MINOR must load: %v", err)
	}
	if d.Version != "1.1" {
		t.Fatalf("Version = %q, want 1.1", d.Version)
	}
	found := false
	for _, f := range d.Extra {
		if f.Key == "futureMinorKey" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected unknown key futureMinorKey to survive parse, Extra=%v", d.Extra)
	}

	out, err := SerializeDocument(d)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if !strings.Contains(string(out), `"version": "1.1"`) {
		t.Fatalf("version must not be lowered to %s on save:\n%s", SupportedVersion, out)
	}
	if !strings.Contains(string(out), `"futureMinorKey": true`) {
		t.Fatalf("unknown key must survive serialize:\n%s", out)
	}
}

// TestVersionMustBeMajorDotMinor guards the format itself.
func TestVersionMustBeMajorDotMinor(t *testing.T) {
	for _, v := range []string{"1", "1.0.0", "a.b", ""} {
		if _, _, err := parseVersion(v); err == nil {
			t.Errorf("parseVersion(%q) should have failed", v)
		}
	}
}
