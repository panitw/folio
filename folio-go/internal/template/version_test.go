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
func TestHigherMajorIsLoadError(t *testing.T) {
	_, err := ParseDocument(withVersion("2.0"))
	if err == nil {
		t.Fatal("expected a load error for a higher MAJOR version")
	}
	if !strings.Contains(err.Error(), "2.0") || !strings.Contains(err.Error(), SupportedVersion) {
		t.Fatalf("error must name both the declared and supported version, got: %v", err)
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
