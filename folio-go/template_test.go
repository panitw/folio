package folio

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimalTemplateJSON is a well-formed `.folio` document (AC4): a
// version, page setup, and ordered (empty) band content.
const minimalTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": []
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
`

// TestParseTemplate is AC4/AC1: a well-formed `.folio` file parses with
// no error via ParseTemplate.
func TestParseTemplate(t *testing.T) {
	tpl, err := ParseTemplate([]byte(minimalTemplateJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	if tpl == nil {
		t.Fatal("ParseTemplate returned a nil *Template with no error")
	}
	if tpl.Version != "1.0" {
		t.Fatalf("Version = %q, want 1.0", tpl.Version)
	}
}

// TestLoadTemplate is AC1/AC4: LoadTemplate reads a path from disk and
// delegates to ParseTemplate.
func TestLoadTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "minimal.folio")
	if err := os.WriteFile(path, []byte(minimalTemplateJSON), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	tpl, err := LoadTemplate(path)
	if err != nil {
		t.Fatalf("LoadTemplate: %v", err)
	}
	if tpl.Version != "1.0" {
		t.Fatalf("Version = %q, want 1.0", tpl.Version)
	}
}

// TestLoadTemplateMissingFile confirms LoadTemplate surfaces the
// underlying os error rather than swallowing it.
func TestLoadTemplateMissingFile(t *testing.T) {
	_, err := LoadTemplate(filepath.Join(t.TempDir(), "does-not-exist.folio"))
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

// TestHigherMajorVersionIsRejected is AC6: a template declaring a MAJOR
// format version newer than the library supports fails to load, naming
// the declared and supported version, and no render is attempted (there
// is no *Template to pass to Render at all).
func TestHigherMajorVersionIsRejected(t *testing.T) {
	bad := `{"assets":{},"bands":{"content":{"elements":[]},"pageFooter":{"elements":[],"height":20},"pageHeader":{"elements":[],"height":20}},"fonts":{},"locale":"en","nextId":1,"page":{"margin":{"bottom":36,"left":36,"right":36,"top":36},"orientation":"portrait","size":"A4"},"utcOffset":"+00:00","version":"9.0"}`
	_, err := ParseTemplate([]byte(bad))
	if err == nil {
		t.Fatal("expected a load error for a higher MAJOR version")
	}
	// This story's finisher review (Finding 19, Minor): the assertion
	// used to stop at err != nil, though the comment above already
	// claimed the error names both versions — the stronger check existed
	// only in internal/template/version_test.go, one layer down. Assert
	// it here too, at the module-root public API this AC actually names.
	if !strings.Contains(err.Error(), "9.0") || !strings.Contains(err.Error(), "1.0") {
		t.Fatalf("error must name both the declared (9.0) and supported (1.0) version, got: %v", err)
	}
}

// TestRenderAcceptsATemplate is AC2: Render's signature is
// func Render(t *Template) ([]byte, error).
func TestRenderAcceptsATemplate(t *testing.T) {
	tpl, err := ParseTemplate([]byte(minimalTemplateJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	b, err := Render(tpl)
	if err != nil {
		t.Fatalf("Render(tpl): %v", err)
	}
	if len(b) == 0 {
		t.Fatal("Render(tpl) produced no bytes")
	}
}

// TestRenderIgnoresItsArgumentToday is D-1.4.16's required pinning test
// (this story's finisher review, Finding 7, Major): "documentation alone
// decays into prose nobody re-reads" — Render's PROVISIONAL doc comment
// (folio.go) states it accepts but does not yet consume its *Template
// argument, and this test asserts the fact rather than only documenting
// it. Render(tpl) and Render(nil) must produce byte-identical output
// TODAY. THIS TEST IS DESIGNED TO FAIL AND BE DELETED: the moment
// Story 1.5 or 1.6 makes Render read its argument (fonts, layout, real
// data), Render(tpl) and Render(nil) will legitimately differ — that
// failure is expected and is this test's own expiry notice, not a
// regression to chase.
func TestRenderIgnoresItsArgumentToday(t *testing.T) {
	tpl, err := ParseTemplate([]byte(minimalTemplateJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	withTemplate, err := Render(tpl)
	if err != nil {
		t.Fatalf("Render(tpl): %v", err)
	}
	withNil, err := Render(nil)
	if err != nil {
		t.Fatalf("Render(nil): %v", err)
	}
	if !bytes.Equal(withTemplate, withNil) {
		t.Fatalf("Render(tpl) and Render(nil) differ — if this is because Story 1.5/1.6 made Render consume its argument, DELETE this test rather than fix it; that is what it exists to force")
	}
}
