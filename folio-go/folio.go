// Package folio is the public entry point of the folio-go rendering
// library.
package folio

import (
	"os"

	"github.com/panitw/folio/folio-go/internal/pdf"
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

// Render produces the fixed, single-page PDF 1.7 document described by
// Story 1.1: one filled rectangle on an A4 page, with no compression, no
// /Info dictionary, and no /CreationDate or /ModDate.
//
// PROVISIONAL — Story 1.1 through Story 1.7. Per D-1.1.c's
// one-parameter-per-story convergence, this story (1.4) gives Render
// its template parameter: func Render(t *Template) ([]byte, error).
// The parameter is accepted and NOT YET consumed — internal/layout and
// internal/pdf do not read a Template yet (neither package exists in a
// template-consuming form; that lands with Story 1.5's layout engine).
// Render still emits Story 1.1's fixed rectangle regardless of what t
// contains. Story 1.6 gives it data and runtime parameters, and
// Story 1.7 the RenderTo(w io.Writer, …) writer form — which arrives
// without breaking this call, per the architecture spine's Deferred
// row. Do not build against this signature as if it were final.
func Render(t *Template) ([]byte, error) {
	_ = t // not yet consumed — see the PROVISIONAL note above.
	return pdf.Serialize(), nil
}
