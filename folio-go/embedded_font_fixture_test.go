package folio

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtures/embedded-font/ is Story 8.3's artifact: THE FIRST `.folio` IN THIS
// REPOSITORY THAT CARRIES A FONT FACE.
//
// WHAT IT RED-PROVES: that a face can travel inside a template — stored,
// loaded, round-tripped and projected — through the same `assets` mechanism
// images already use, and that a document declaring one declares format
// version 2.0.
//
// WHAT IT DOES NOT COVER: rendering FROM the embedded face. That is Story 8.4.
// The chain's FIRST entry is the shipped Noto Sans, and that is the face the
// page is drawn with; the embedded entry contributes nothing, which is why
// this fixture's own matrix guard asserts the rendered PDF embeds EXACTLY ONE
// font program. When 8.4 lands, that guard is what will have to change, and
// deliberately.
//
// IT SHIPS NO expected.pdf, on the hidden-image precedent (Story 3.5). An
// `expected.pdf` is a human-attested artifact under AD-21/D-4.7.1, and this
// story cannot produce the one that matters — a page drawn WITH the embedded
// face. Recording a page drawn with the shipped face under the name
// "embedded-font" would attest the wrong thing. So goldenDigestRecord stays at
// 22 and this fixture's acceptance is STRUCTURAL: it is registered in
// matrixDocuments with a recorded cross-target hash, and its input.folio is
// pinned byte-for-byte against the template constant below.

// embeddedFontAssetBytes is the face the fixture carries: the SHIPPED Noto
// Sans Thai, embedded a second time as an ASSET rather than supplied through
// the FontSet. NO NEW BINARY ENTERS THE REPOSITORY — this is the same
// folio-go/fonts/notosansthai/NotoSansThai-Regular.ttf the module already
// commits, reached through the test binary's own embed (testfont_embed_test.go)
// so the bytes travel to every cross-target leg, js/wasm included.
//
// Thai deliberately, and Latin text on the page deliberately: the face the
// document carries is the one the page does NOT need, so an implementation
// that silently started rendering from the embedded face would change the
// page's bytes and be caught rather than absorbed.
func embeddedFontAssetBytes() []byte { return testShippedNotoSansThai }

// embeddedFontAssetKey is the key the format's own rule produces: the
// lowercase hex SHA-256 of the DECODED bytes. It is derived here rather than
// written down, because a written-down digest is a second authority on what
// the key is, and this fixture exists partly to show there is only one.
func embeddedFontAssetKey() string {
	sum := sha256.Sum256(embeddedFontAssetBytes())
	return fmt.Sprintf("%x", sum)
}

// embeddedFontTemplateJSON is fixtures/embedded-font/input.folio, BUILT rather
// than transcribed.
//
// Every other matrix document keeps its template as a hand-copied Go string
// constant. This one cannot: its asset is ~47 KB of base64, and a hand-copied
// constant of that size is a diff nobody reads and a drift nobody sees. So the
// document is generated from the shipped bytes by the format's own rules —
// SHA-256 for the key, canonical 76-column wrapping for the data — and
// TestEmbeddedFontFixtureMatchesInputFolio pins the committed file against it.
// The tie is the same one every other fixture has; only its direction is
// reversed (the constant is derived, the file is checked), and that is
// strictly stronger, because the derivation is the format's rule rather than a
// human's copy of its output.
func embeddedFontTemplateJSON() string {
	encoded := base64Wrapped76(embeddedFontAssetBytes())
	var data strings.Builder
	for i, line := range encoded {
		if i > 0 {
			data.WriteString(",\n")
		}
		data.WriteString("        \"" + line + "\"")
	}
	return `{
  "assets": {
    "` + embeddedFontAssetKey() + `": {
      "data": [
` + data.String() + `
      ],
      "font": {
        "family": "Noto Sans Thai",
        "licence": "SIL Open Font License 1.1",
        "source": "folio-go/fonts/notosansthai/NotoSansThai-Regular.ttf — the shipped static Regular instance; see that directory's NOTICE.md for the derivation",
        "style": "Regular"
      },
      "mediaType": "font/ttf"
    }
  },
  "bands": {
    "content": {
      "elements": [
        {
          "height": 40,
          "id": "e1",
          "style": {
            "fontFamily": "body",
            "fontSize": 12
          },
          "type": "text",
          "value": "A font travels inside the template.",
          "width": 400,
          "x": 0,
          "y": 0
        }
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [],
      "height": 20
    }
  },
  "fonts": {
    "body": [
      "Noto Sans",
      {
        "asset": "` + embeddedFontAssetKey() + `"
      }
    ]
  },
  "locale": "en",
  "nextId": 2,
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
  "version": "2.0"
}
`
}

// base64Wrapped76 reproduces the format's canonical `data` wrapping. It is a
// SECOND site for that rule, and the only reason it is tolerable is that
// TestEmbeddedFontFixtureIsCanonical below asserts the whole document is a
// serializer fixed point — so if this wrapping ever disagreed with
// splitBase64Canonical, the fixture would stop round-tripping and this file
// would go red rather than shipping a document the engine rewrites on save.
func base64Wrapped76(raw []byte) []string {
	const width = 76
	encoded := base64.StdEncoding.EncodeToString(raw)
	var out []string
	for len(encoded) > width {
		out = append(out, encoded[:width])
		encoded = encoded[width:]
	}
	return append(out, encoded)
}

// TestEmbeddedFontFixtureMatchesInputFolio ties the committed file to the
// template the matrix actually renders. Without it the two drift, and the
// fixture stops documenting what the matrix measured.
func TestEmbeddedFontFixtureMatchesInputFolio(t *testing.T) {
	path := filepath.Join(repoRootFromTest(t), "fixtures", "embedded-font", "input.folio")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(raw) == 0 {
		t.Fatal("presence precondition: fixtures/embedded-font/input.folio is empty")
	}
	if string(raw) != embeddedFontTemplateJSON() {
		t.Errorf("embeddedFontTemplateJSON and fixtures/embedded-font/input.folio have DIVERGED — the matrix renders one and the repository documents the other")
	}
}

// TestEmbeddedFontFixtureIsCanonicalAndDeclaresTwoPointZero is the fixture's
// own acceptance, and it is what makes AC6 structural rather than absent.
func TestEmbeddedFontFixtureIsCanonicalAndDeclaresTwoPointZero(t *testing.T) {
	source := embeddedFontTemplateJSON()
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if string(out) != source {
		t.Fatalf("the fixture is not a serializer fixed point — the engine would rewrite it on save:\n--- got ---\n%s\n--- want ---\n%s", firstBytes(string(out)), firstBytes(source))
	}
	// The version the fixture DECLARES is the version its content REQUIRES:
	// an embedded-face entry is a 2.0 document (Story 8.3).
	if !strings.Contains(source, `"version": "2.0"`) {
		t.Error("a document declaring an embedded-face entry must declare version 2.0")
	}
	// And it really renders, with the shipped face, against the shipped set.
	res, err := Render(tpl, Data(`{}`), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if len(res.Bytes) == 0 {
		t.Fatal("the fixture rendered zero bytes")
	}
	// THE INTERIM STATE, PINNED. Exactly one font program reaches the page:
	// the shipped Noto Sans the chain's FIRST entry names. The embedded Noto
	// Sans Thai contributes nothing, because rendering from an embedded face
	// is Story 8.4. When that story lands this assertion must change, and
	// changing it deliberately is the point of writing it down.
	programs := extractAllFontFile2Programs(t, res.Bytes)
	if len(programs) != 1 {
		t.Fatalf("the fixture's render embeds %d font programs, want exactly 1 — the embedded face must not reach the page until Story 8.4", len(programs))
	}
}

// firstBytes bounds a mismatch report: the fixture is ~64 KB, and dumping two
// copies of it into a test log describes nothing.
func firstBytes(s string) string {
	const limit = 400
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "\n… (truncated; the documents differ somewhere in " + fmt.Sprint(len(s)) + " bytes)"
}

// TestUnrecognisedFontMediaTypeIsValidToo is AC's positive control at the
// PUBLIC boundary: `Validate` must be SILENT over a document carrying a font
// asset whose media type this build does not recognise.
//
// The internal half of this is proved in internal/template
// (TestUnrecognisedFontMediaTypeLoadsClean). This is the half that matters to
// an integrator, and it is asserted rather than inferred, because `Validate`
// predicts `Render` (D-1.8.1 amended) and a `Validate` that refused what
// `Render` accepts — or accepted what `Render` refuses — would be a second
// rule system. It is the arm a "closed set of font media types" implementation
// fails.
func TestUnrecognisedFontMediaTypeIsValidToo(t *testing.T) {
	source := strings.Replace(embeddedFontTemplateJSON(), `"mediaType": "font/ttf"`, `"mediaType": "font/woff2"`, 1)
	if !strings.Contains(source, "font/woff2") {
		t.Fatal("fixture assumption violated: the media type was not substituted")
	}
	diags, err := Validate([]byte(source), Data(`{}`), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("a font asset with an unrecognised mediaType must be VALID (D-1.8.1 amended), got: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("Validate must be SILENT over an unrecognised font media type, got %d diagnostic(s): %+v", len(diags), diags)
	}
	// And it still renders — the page is drawn with the shipped face the chain
	// names first, and the carried asset is simply along for the ride.
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := Render(tpl, Data(`{}`), nil, testShippedFontSet()); err != nil {
		t.Fatalf("render: %v", err)
	}
}

// TestTheFontRecordCostsAnExistingDocumentNothing is the story's last AC,
// stated as something that is actually true and actually measured.
//
// WHAT IS NOT ASSERTED HERE, AND WHY. "Re-serializing a committed fixture
// reproduces its bytes" is FALSE at HEAD and was false before this story:
// several committed `input.folio` files are hand-written for readability
// (compact bands, one-line style blocks) and are not canonical. Writing that
// assertion would have meant either changing those files — moving the very
// bytes this AC is about — or quietly restricting the population until it
// passed. Both are worse than saying what holds.
//
// WHAT IS ASSERTED. Over EVERY committed document: the serializer's output is
// a fixed point (so nothing this story added makes an existing document
// unstable), and the output carries NO `font` key for any document that did
// not declare one — an absent Presence stays absent and never reappears as an
// authored null or an empty object. The asset-bearing population is counted
// rather than named, so a fixture that gains or loses an assets map is read
// here rather than slipping past.
//
// The stronger statement — that no committed byte moved at all — is carried by
// the 22 recorded golden PDF digests (AC6, unchanged) and by the diff itself,
// which touches no pre-existing fixture.
func TestTheFontRecordCostsAnExistingDocumentNothing(t *testing.T) {
	dir := filepath.Join(repoRootFromTest(t), "fixtures")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read fixtures/: %v", err)
	}
	withAssets, checked := 0, 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name(), "input.folio")
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			continue // not every fixture ships an input.folio
		}
		checked++
		canonical := serializeTwice(t, entry.Name(), raw)
		if entry.Name() != "embedded-font" && strings.Contains(canonical, `"font"`) {
			t.Errorf("%s: a document that declares no font record must not acquire one on serialize", entry.Name())
		}
		if assetCount(t, raw) > 0 {
			withAssets++
		}
	}
	if checked == 0 {
		t.Fatal("vacuity guard: no fixture input.folio was read, so this test asserted nothing")
	}
	if withAssets != 7 {
		t.Errorf("fixtures carrying a non-empty assets map = %d, want 7 (the six that predate Story 8.3 plus embedded-font) — if a fixture gained or lost an asset map, say so here deliberately", withAssets)
	}
	t.Logf("byte-neutrality witness — %d committed documents serialized, %d of them carrying a non-empty assets map", checked, withAssets)
}

// serializeTwice returns a document's canonical bytes, having asserted they are
// a fixed point: Serialize(Parse(Serialize(Parse(b)))) == Serialize(Parse(b)).
func serializeTwice(t *testing.T, name string, raw []byte) string {
	t.Helper()
	once := serializeOnce(t, name, raw)
	twice := serializeOnce(t, name, []byte(once))
	if twice != once {
		t.Errorf("%s: the serializer is not a fixed point over this document", name)
	}
	return once
}

func serializeOnce(t *testing.T, name string, raw []byte) string {
	t.Helper()
	tpl, err := ParseTemplate(raw)
	if err != nil {
		t.Fatalf("%s: parse: %v", name, err)
	}
	out, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatalf("%s: serialize: %v", name, err)
	}
	return string(out)
}

// assetCount reads how many entries a committed document's `assets` map holds,
// off the JSON rather than off the parsed model, so the population this test
// reports is the one the FILE declares.
func assetCount(t *testing.T, raw []byte) int {
	t.Helper()
	var doc struct {
		Assets map[string]json.RawMessage `json:"assets"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal a committed document: %v", err)
	}
	return len(doc.Assets)
}
