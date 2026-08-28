package folio

import "github.com/panitw/folio/folio-go/internal/template"

// SerializeTemplate returns the engine's canonical .folio bytes for t.
//
// The Template remains opaque: callers receive bytes, not the internal
// document or a field-by-field representation of it. This is the only save
// seam used by the browser worker (AD-15/AD-16).
func SerializeTemplate(t *Template) ([]byte, error) {
	if t == nil {
		return nil, errNilTemplate
	}
	return template.SerializeDocument(t.doc)
}
