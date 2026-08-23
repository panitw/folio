package template

import (
	"path/filepath"
	"reflect"
	"testing"
)

// canonicalFixtures is the corpus of canonical `.folio` bytes this
// story's P1/P2/P3 tests run over (AC12–AC17). Every entry here is
// itself asserted, by mustBeCanonical, to already be a fixed point
// under the serializer before it is used to test anything else — a
// fixture that was never actually canonical would make every
// downstream assertion about it meaningless.
func canonicalFixtures(t *testing.T) map[string][]byte {
	t.Helper()
	root := repoRootFromTest(t)
	golden := mustReadFile(t, filepath.Join(root, "folio-go", "testdata", "template", "golden", "worked-example.json"))

	minimal := []byte(`{
  "assets": {},
  "bands": {
    "content": {
      "elements": []
    },
    "pageFooter": {
      "elements": [],
      "height": 40
    },
    "pageHeader": {
      "elements": [],
      "height": 80
    }
  },
  "fonts": {},
  "locale": "en",
  "nextId": 1,
  "page": {
    "margin": {
      "bottom": 36,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
`)

	return map[string][]byte{
		"worked-example": golden,
		"minimal":        minimal,
		"unknown-keys":   unknownKeysFixture,
		"html-trap":      htmlEscapeTrapFixture,
		"utf8-trap":      utf8TrapFixture,
		"escape-trap":    minimalEscapeTrapFixture,
		"maximal":        maximalFixture,
		"null-field":     nullFieldFixture,
	}
}

func mustBeCanonical(t *testing.T, name string, b []byte) {
	t.Helper()
	d, err := ParseDocument(b)
	if err != nil {
		t.Fatalf("[%s] parse: %v", name, err)
	}
	out, err := SerializeDocument(d)
	if err != nil {
		t.Fatalf("[%s] serialize: %v", name, err)
	}
	if string(out) != string(b) {
		t.Fatalf("[%s] fixture is not itself canonical (P3 precondition failed):\nwant:\n%s\ngot:\n%s", name, b, out)
	}
}

// TestP3FixturesAreCanonical is P3's forward half over the whole corpus:
// b canonical => Serialize(Parse(b)) == b.
func TestP3FixturesAreCanonical(t *testing.T) {
	for name, b := range canonicalFixtures(t) {
		t.Run(name, func(t *testing.T) { mustBeCanonical(t, name, b) })
	}
}

// TestP1RoundTripThroughValue is AC12: for every Document d obtained by
// parsing a valid file, Parse(Serialize(d)) == d — equality on the
// parsed value, not on bytes.
func TestP1RoundTripThroughValue(t *testing.T) {
	for name, b := range canonicalFixtures(t) {
		t.Run(name, func(t *testing.T) {
			d, err := ParseDocument(b)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			out, err := SerializeDocument(d)
			if err != nil {
				t.Fatalf("serialize: %v", err)
			}
			d2, err := ParseDocument(out)
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			if !reflect.DeepEqual(d, d2) {
				t.Fatalf("Parse(Serialize(d)) != d:\nd:  %#v\nd2: %#v", d, d2)
			}
		})
	}
}

// TestP2NormalisationIsIdempotent is AC13: for every valid input b,
// Serialize(Parse(b)) == Serialize(Parse(Serialize(Parse(b)))).
func TestP2NormalisationIsIdempotent(t *testing.T) {
	inputs := canonicalFixtures(t)
	inputs["non-canonical-numbers"] = nonCanonicalNumberFixture

	for name, b := range inputs {
		t.Run(name, func(t *testing.T) {
			d, err := ParseDocument(b)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			out1, err := SerializeDocument(d)
			if err != nil {
				t.Fatalf("serialize: %v", err)
			}
			d2, err := ParseDocument(out1)
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			out2, err := SerializeDocument(d2)
			if err != nil {
				t.Fatalf("re-serialize: %v", err)
			}
			if string(out1) != string(out2) {
				t.Fatalf("normalisation is not idempotent:\nfirst:\n%s\nsecond:\n%s", out1, out2)
			}
		})
	}
}

// TestP3NonCanonicalNormalises is P3's "iff": a non-canonical but valid
// input does NOT round-trip byte-identically to itself, yet normalises
// to exactly the same bytes as the already-canonical form of the same
// document (AC27: legality is a property of the value, not the
// spelling).
func TestP3NonCanonicalNormalises(t *testing.T) {
	d, err := ParseDocument(nonCanonicalNumberFixture)
	if err != nil {
		t.Fatalf("parse non-canonical: %v", err)
	}
	out, err := SerializeDocument(d)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if string(out) == string(nonCanonicalNumberFixture) {
		t.Fatalf("non-canonical fixture unexpectedly round-tripped byte-identically to itself")
	}

	canonical := canonicalFixtures(t)["minimal"]
	if string(out) != string(canonical) {
		t.Fatalf("normalised non-canonical input does not match the canonical form of the same document:\nnormalised:\n%s\ncanonical:\n%s", out, canonical)
	}
}

// nonCanonicalNumberFixture is semantically identical to the "minimal"
// canonical fixture but spells its numbers non-canonically (AC27:
// "36.0000" -> "36" style normalisation, on margin top/right/bottom/left
// and the two band heights) and is deliberately squeezed onto few lines
// (also non-canonical whitespace) to keep this file smaller — whitespace
// outside string/number tokens carries no semantic weight, only the
// decoded value does.
var nonCanonicalNumberFixture = []byte(`{"assets":{},"bands":{"content":{"elements":[]},"pageFooter":{"elements":[],"height":4.0e1},"pageHeader":{"elements":[],"height":80.000}},"fonts":{},"locale":"en","nextId":1,"page":{"margin":{"bottom":36.0000,"left":3.6e1,"right":36.0,"top":36},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"1.0"}`)
