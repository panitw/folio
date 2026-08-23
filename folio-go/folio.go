// Package folio is the public entry point of the folio-go rendering
// library.
package folio

import (
	"os"

	"github.com/panitw/folio/folio-go/internal/template"
)

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
