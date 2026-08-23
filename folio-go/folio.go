// Package folio is the public entry point of the folio-go rendering
// library.
package folio

import (
	"os"

	"github.com/panitw/folio/folio-go/internal/template"
)

// Template is a parsed, canonicalised `.folio` document (AC1). It is an
// OPAQUE handle wrapping internal/template.Document — deliberately NOT
// a type alias (Epic 1 boundary gate finding): an alias made every
// exported field of internal/template.Document, and all 32 exported
// declarations reachable from it, part of the public surface, and let
// a caller construct a Template field-by-field, bypassing
// ParseTemplate entirely — sidestepping asset-key validation, the
// version rules, the eight closed sets, exact-decimal discipline and
// nextId monotonicity. Construction is only through LoadTemplate/
// ParseTemplate; there are no exported fields and no accessors are
// shipped pre-emptively (add them only when a real need appears — a
// MINOR addition later, versus a MAJOR break if the surface had to
// close after being opened). See testdata/templateopaque/ for the
// compile-time proof that composite-literal construction of Template
// does not type-check.
//
// A defined type without the "=" (type Template template.Document)
// would NOT close this hole: it still exposes every exported field of
// template.Document and still permits composite-literal construction
// (folio.Template{Version: "1.0"} would compile). Do not take that
// route.
type Template struct {
	doc *template.Document
}

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
	doc, err := template.ParseDocument(b)
	if err != nil {
		return nil, err
	}
	return &Template{doc: doc}, nil
}
