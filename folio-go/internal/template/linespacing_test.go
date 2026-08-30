package template

import (
	"strings"
	"testing"
)

// TestDecodeLineSpacingDomain is D-7.2.3's range, asserted as the domain
// it actually is rather than as a typographic opinion: a whole number of
// thousandths in [1, 1000000], nothing rounded, nothing clamped.
func TestDecodeLineSpacingDomain(t *testing.T) {
	for _, c := range []struct {
		literal string
		want    int64
		note    string
	}{
		{"1", 1000, "the neutral ratio"},
		{"1.5", 1500, "the worked case"},
		{"0.001", 1, "the floor — no 1.0 minimum: tight leading is what a filed contract is set in"},
		{"0.6", 600, "genuinely tight, and legal"},
		{"1000", 1000000, "the stated sanity ceiling"},
		{"1000.0", 1000000, "a non-canonical spelling of the ceiling is accepted and normalised"},
		{"1.500", 1500, "trailing zeros are a spelling, not a value"},
		{"1e0", 1000, "exponent notation is legal JSON and legal here"},
	} {
		got, err := DecodeLineSpacing(c.literal)
		if err != nil {
			t.Errorf("DecodeLineSpacing(%q) = error %v, want %d (%s)", c.literal, err, c.want, c.note)
			continue
		}
		if got != c.want {
			t.Errorf("DecodeLineSpacing(%q) = %d, want %d (%s)", c.literal, got, c.want, c.note)
		}
	}

	for _, c := range []struct {
		literal string
		reason  string
		note    string
	}{
		{"0", "range", "zero is below the floor"},
		{"-1", "range", "negative is below the floor"},
		{"-0.5", "range", "a negative ratio is not a ratio"},
		{"1000.001", "range", "one thousandth past the ceiling"},
		{"1.0005", "decimal places", "four decimal places is not a whole number of thousandths"},
		{"0.0001", "decimal places", "below the floor AND inexact — the exactness check runs first"},
	} {
		_, err := DecodeLineSpacing(c.literal)
		if err == nil {
			t.Errorf("DecodeLineSpacing(%q) was accepted; %s", c.literal, c.note)
			continue
		}
		if !strings.Contains(err.Error(), c.reason) {
			t.Errorf("DecodeLineSpacing(%q) = %q, want a reason mentioning %q (%s)", c.literal, err.Error(), c.reason, c.note)
		}
	}
}

// TestLineSpacingRoundTripsInItsAuthoredSpelling is AD-9's
// edit-and-edit-back byte identity for the new key at both ends of its
// domain: thousandths through appendPoints is the same ×1000 decimal path
// decodePoints read it with, so there is no second number formatter for
// the two to disagree about.
func TestLineSpacingRoundTripsInItsAuthoredSpelling(t *testing.T) {
	for _, c := range []struct{ authored, canonical string }{
		{"1", "1"},
		{"1.5", "1.5"},
		{"0.875", "0.875"},
		{"0.001", "0.001"},
		{"1000", "1000"},
		{"1.500", "1.5"},   // normalised, and stably so
		{"1e0", "1"},       // ditto
		{"1000.0", "1000"}, // ditto
	} {
		doc := lineSpacingRoundTripDoc(c.authored)
		d, err := ParseDocument([]byte(doc))
		if err != nil {
			t.Errorf("parse %s: %v", c.authored, err)
			continue
		}
		out, err := SerializeDocument(d)
		if err != nil {
			t.Errorf("serialize %s: %v", c.authored, err)
			continue
		}
		want := `"lineSpacing": ` + c.canonical
		if !strings.Contains(string(out), want) {
			t.Errorf("authored %s serialized without %s:\n%s", c.authored, want, out)
		}
		// A second pass must be a fixed point.
		d2, err := ParseDocument(out)
		if err != nil {
			t.Errorf("reparse %s: %v", c.authored, err)
			continue
		}
		out2, err := SerializeDocument(d2)
		if err != nil {
			t.Errorf("reserialize %s: %v", c.authored, err)
			continue
		}
		if string(out2) != string(out) {
			t.Errorf("authored %s is not a fixed point after one canonicalisation:\n%s\nvs\n%s", c.authored, out, out2)
		}
	}
}

// TestVersionForSaveIsRaisedOnlyByContent pins D-1.4.13 at the one
// function that expresses it, including the retrofit D-7.2.1 obliged:
// Epic 10 shipped style.color and bumped nothing, so colour-bearing
// documents declared 1.0 while requiring 1.1.
func TestVersionForSaveIsRaisedOnlyByContent(t *testing.T) {
	for _, c := range []struct {
		label  string
		style  string
		loaded string
		want   string
	}{
		{"neither key, 1.0 in", "", "1.0", "1.0"},
		{"lineSpacing, 1.0 in", `"lineSpacing": 1.5, `, "1.0", "1.1"},
		{"color, 1.0 in", `"color": "#112233", `, "1.0", "1.1"},
		{"explicit null color still declares the key", `"color": null, `, "1.0", "1.1"},
		{"neither key, 1.1 in — never lowered", "", "1.1", "1.1"},
		{"neither key, 1.9 in — never lowered", "", "1.9", "1.9"},
		{"lineSpacing, 1.9 in — the higher wins", `"lineSpacing": 1.5, `, "1.9", "1.9"},
		// MAJOR 0 is LOADABLE: parseVersion admits major >= 0 and
		// checkVersionLoadable refuses only major > SupportedMajor. A 0.x
		// document that introduces no 1.1 key must round-trip verbatim.
		// Raising it to the baseVersion FLOOR would rewrite a field the
		// document owns, on a save that changed nothing (AD-9).
		{"neither key, 0.9 in — a floor is not a demand", "", "0.9", "0.9"},
		{"lineSpacing, 0.9 in — real content still raises", `"lineSpacing": 1.5, `, "0.9", "1.1"},
	} {
		t.Run(c.label, func(t *testing.T) {
			doc := strings.Replace(lineSpacingRoundTripDocWithStyle(c.style), `"version": "1.0"`, `"version": "`+c.loaded+`"`, 1)
			d, err := ParseDocument([]byte(doc))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := versionForSave(d.Version, d); got != c.want {
				t.Errorf("versionForSave(%q) = %q, want %q", c.loaded, got, c.want)
			}
		})
	}

	// The rule must reach BOTH attachment points, not only element.style:
	// a table that sets lineSpacing on headerStyle alone still requires
	// 1.1, and a version that missed it would misdeclare exactly the way
	// style.color did.
	d, err := ParseDocument([]byte(lineSpacingHeaderStyleDoc))
	if err != nil {
		t.Fatalf("parse headerStyle doc: %v", err)
	}
	if got := versionRequiredByContent(d); got != minorFeatureVersion {
		t.Errorf("a document whose only 1.1 key is on headerStyle requires %q, want %q", got, minorFeatureVersion)
	}

	// And an unparseable loaded version is returned untouched rather than
	// invented over — it reached here only through checkVersionLoadable.
	if got := versionForSave("not-a-version", d); got != "not-a-version" {
		t.Errorf("versionForSave over an unparseable version = %q, want it returned untouched", got)
	}
}

// TestContentVersionNeverExceedsTheLibraryCeiling states the relation the
// three version constants currently only IMPLY by coincidence.
//
// versionForSave stamps versionRequiredByContent onto documents it saves.
// If a future key ever raised that above SupportedVersion, this library
// would write files ITS OWN LOADER refuses — checkVersionLoadable rejects
// a higher MAJOR outright — and it would do so silently, because nothing
// compares the two. The failure would surface as "the engine cannot
// reopen what it just wrote", which is the worst possible place to
// discover it.
//
// The bound is asserted over every version the content rule can return,
// derived from the constants themselves rather than from a list someone
// remembered to extend.
func TestContentVersionNeverExceedsTheLibraryCeiling(t *testing.T) {
	ceilingMajor, ceilingMinor, err := parseVersion(SupportedVersion)
	if err != nil {
		t.Fatalf("SupportedVersion %q does not parse: %v", SupportedVersion, err)
	}
	if ceilingMajor != SupportedMajor {
		t.Errorf("SupportedVersion %q declares MAJOR %d but SupportedMajor is %d — checkVersionLoadable and versionForSave would disagree about what this library can load", SupportedVersion, ceilingMajor, SupportedMajor)
	}
	for _, v := range []string{baseVersion, minorFeatureVersion} {
		major, minor, perr := parseVersion(v)
		if perr != nil {
			t.Errorf("%q does not parse: %v", v, perr)
			continue
		}
		if major > ceilingMajor || (major == ceilingMajor && minor > ceilingMinor) {
			t.Errorf("the content rule can return %q, which is ABOVE the library's own ceiling %q — this library would write documents its own loader refuses", v, SupportedVersion)
		}
	}
	// baseVersion must also be the FLOOR: a document requiring nothing
	// must not be stamped above the lowest version the format has.
	if baseVersion != "1.0" {
		t.Errorf("baseVersion = %q, want the format's own lowest version 1.0 — a document that needs nothing new must declare it", baseVersion)
	}
	// And the rule's own output is inside the bound for every document
	// shape it distinguishes, not merely for the constants in isolation.
	for _, style := range []string{"", `"lineSpacing": 1.5, `, `"color": "#112233", `} {
		d, perr := ParseDocument([]byte(lineSpacingRoundTripDocWithStyle(style)))
		if perr != nil {
			t.Fatalf("parse: %v", perr)
		}
		got := versionRequiredByContent(d)
		major, minor, _ := parseVersion(got)
		if major > ceilingMajor || (major == ceilingMajor && minor > ceilingMinor) {
			t.Errorf("versionRequiredByContent returned %q, above the ceiling %q", got, SupportedVersion)
		}
	}
}

func lineSpacingRoundTripDoc(authored string) string {
	return lineSpacingRoundTripDocWithStyle(`"lineSpacing": ` + authored + `, `)
}

func lineSpacingRoundTripDocWithStyle(style string) string {
	return `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 40, "value": "v", "style": {` + style + `"fontFamily": "body", "fontSize": 11}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 2,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
}

const lineSpacingHeaderStyleDoc = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "rows[]", "as": "row", "headerHeight": 20,
          "columns": [{"id": "e2", "label": "L", "width": 100, "bind": "{{row.v}}"}],
          "headerStyle": {"lineSpacing": 1.5}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
