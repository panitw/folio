package folio

import (
	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/template"
)

// SerializeTemplate returns the engine's canonical .folio bytes for t.
//
// The Template remains opaque: callers receive bytes, not the internal
// document or a field-by-field representation of it. This is the only save
// seam used by the browser worker (AD-15/AD-16).
func SerializeTemplate(t *Template) ([]byte, error) {
	if t == nil {
		return nil, errNilTemplate
	}
	return template.SerializeDocumentWithMinimumVersion(t.doc, expressionMinimumVersion(t.doc))
}

// Parse the actual expressions, so quoted operator punctuation and ordinary
// paths never accidentally raise the saved format requirement.
func expressionMinimumVersion(doc *template.Document) string {
	formula := func(raw string) bool { e, err := expr.Parse(raw); return err == nil && expr.UsesFormulas(e) }
	textFormula := func(raw string) bool {
		_, parts, _, err := expr.ScanPlaceholders(raw)
		if err != nil {
			return false
		}
		for _, part := range parts {
			if !part.Reserved && formula(part.Inner) {
				return true
			}
		}
		return false
	}
	for _, band := range []template.Band{doc.Bands.PageHeader, doc.Bands.Content, doc.Bands.PageFooter} {
		for _, el := range band.Elements {
			if el.VisibleIf.Set && !el.VisibleIf.Null && formula(el.VisibleIf.Value) {
				return "2.0"
			}
			if el.Value.Set && !el.Value.Null && textFormula(el.Value.Value) {
				return "2.0"
			}
			if el.Table.Set && !el.Table.Null {
				for _, col := range el.Table.Value.Columns {
					if textFormula(col.Bind) {
						return "2.0"
					}
				}
			}
		}
	}
	return ""
}
