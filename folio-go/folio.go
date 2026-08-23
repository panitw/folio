// Package folio is the public entry point of the folio-go rendering
// library.
package folio

import (
	"errors"
	"os"

	"github.com/panitw/folio/folio-go/internal/template"
)

// errNilTemplate is AC14b's located error for Render(nil, f).
var errNilTemplate = errors.New("folio: Render: template is nil (a document must be loaded via LoadTemplate/ParseTemplate first)")

// Template is a parsed, canonicalised `.folio` document (AC1). It is a
// type alias for internal/template.Document — the model, parser and
// serializer all live under internal/template (AD-9); package folio is
// the "os" boundary D-1.4.6 draws around it: internal/template never
// sees a path, and no internal/ package imports "os" (AC3).
type Template = template.Document

// LoadTemplate reads path from disk and parses it as a `.folio`
// document (AC1). Per D-1.4.6: "So LoadTemplate(path string) and
// ParseTemplate(b []byte) live in package folio at the module root,
// where os is permitted; internal/template never sees a path — it
// takes []byte and returns a *Template or an error. AD-1 is untouched
// because no internal/ package imports os." LoadTemplate delegates to
// ParseTemplate — "this is honest factoring rather than scope growth."
func LoadTemplate(path string) (*Template, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseTemplate(b)
}

// ParseTemplate parses b as a `.folio` document (AC1). It is the sole
// bridge into internal/template's parser (AD-9: "internal/template
// owns both the parser and the serializer").
func ParseTemplate(b []byte) (*Template, error) {
	return template.ParseDocument(b)
}

// Render produces a PDF 1.7 document from t, embedding every font the
// document's text elements use as a subset (AC1, D-1.5.5): the f
// parameter is inserted into its final target position, never appended
// — target is Render(t, d, p, f), so 1.6 adds d before f, 1.7 adds p
// before f (D-1.5.5, verbatim: "No call site is ever reordered, only
// extended").
//
// t must be non-nil (AC14b): Story 1.1's fixture document
// (fixtures/minimal-rect/) predates the public API accepting a
// template at all and is pinned via internal/pdf.Serialize() directly
// (AC14a) — a public Render documented as silently ignoring its
// argument is worse than a located error on nil (D-1.5.9).
//
// PROVISIONAL: text is placed using a provisional band-relative origin
// convention (AC28) that stands in for AD-24's real placement until
// internal/layout exists (Story 2.5); TestProvisionalBandOriginIsPinned
// (render_test.go) fails the day that package exists, forcing this to
// be revisited. Story 1.6 gives Render its data parameter and Story 1.7
// its io.Writer form (D-1.1.c).
func Render(t *Template, f FontSet) ([]byte, error) {
	if t == nil {
		return nil, errNilTemplate
	}
	return renderDocument(t, f)
}
