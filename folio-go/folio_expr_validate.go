package folio

// This file is ParseTemplate's Story 3.2 obligation (R1, R3, AC19,
// AC21): every "{{ }}" expression in the document is parsed and
// statically checked here, at load, and D-1.4.1's footerOf/
// footerFormat derivation runs here too — both forced up to this
// module-root package by F2 (internal/template can never import
// internal/expr).

import (
	"fmt"
	"strings"

	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/template"
)

// validateAndDeriveExpressions walks doc's three bands (AC19: every
// "{{ }}" occurrence in the document, not only footer-related ones)
// and returns D-1.4.1's derived-footer results, keyed by column id
// (R2: never written back into doc itself).
func validateAndDeriveExpressions(doc *template.Document) (map[template.ElementID]expr.DerivedFooter, error) {
	derived := map[template.ElementID]expr.DerivedFooter{}
	for _, band := range []template.Band{doc.Bands.PageHeader, doc.Bands.Content, doc.Bands.PageFooter} {
		for _, el := range band.Elements {
			// visibleIf is a common field on every element kind (QA
			// Finding 12, Minor): the field table (folio-format.md,
			// "visibleIf | *Optional.* An expression; …") and this
			// module's own fixtures (fixtures_test.go:298,
			// "visibleIf": "customer.hasTransactions") agree it holds
			// a BARE expression, unlike a text element's `value` or a
			// table column's `bind` — neither of which wraps it in
			// "{{ }}". checkVisibleIfExpression (below) parses and
			// Checks it directly for exactly that reason:
			// checkTextExpressions would scan for "{{ }}" occurrences
			// inside it and find none, passing silently no matter what
			// the string said — a vacuous check, not a real one. Story
			// 3.5 drives visibility from this exact field; a malformed
			// visibleIf (a bare operator, AC19's own named case) must
			// fail at load, the same as a malformed text value or
			// table bind, not load clean and surface only when 3.5
			// starts evaluating it.
			if el.VisibleIf.Set && !el.VisibleIf.Null {
				if err := checkVisibleIfExpression(el.VisibleIf.Value, el.ID); err != nil {
					return nil, err
				}
			}
			switch el.Type {
			case template.ElementText:
				if el.Value.Set && !el.Value.Null {
					if err := checkTextExpressions(el.Value.Value, el.ID); err != nil {
						return nil, err
					}
				}
			case template.ElementTable:
				if el.Table.Set {
					if err := validateTableColumns(el.Table.Value, derived); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return derived, nil
}

// checkTextExpressions parses and statically checks every non-reserved
// "{{ }}" occurrence in text (AC19, AC10, AC11): a syntax error, a
// wrong arity, an unknown function name, or a wrong-kind literal
// argument is a load error naming elementID and the offending
// expression text, verbatim, as the author wrote it.
func checkTextExpressions(text string, elementID template.ElementID) error {
	_, placeholders, _, serr := expr.ScanPlaceholders(text)
	if serr != nil {
		return fmt.Errorf("folio: ParseTemplate: element %s: %s", elementID, serr)
	}
	for _, ph := range placeholders {
		if ph.Reserved {
			// AC4/AD-4: {{page}}/{{pages}} are short-circuited before
			// any parse attempt, at load exactly as at render.
			continue
		}
		trimmed := strings.TrimSpace(ph.Inner)
		e, perr := expr.Parse(ph.Inner)
		if perr != nil {
			return fmt.Errorf("folio: ParseTemplate: element %s: %q is not a valid expression: %s", elementID, trimmed, perr)
		}
		if cerr := expr.Check(e); cerr != nil {
			return fmt.Errorf("folio: ParseTemplate: element %s: %s", elementID, cerr)
		}
	}
	return nil
}

// checkVisibleIfExpression parses and statically checks a visibleIf
// field's raw value as a BARE expression (QA Finding 12, Minor) — no
// "{{ }}" wrapping, and so no AD-4 reserved-token short-circuit
// either: "page"/"pages" are reserved only as the interpolation
// SYNTAX "{{page}}"/"{{pages}}", which visibleIf never contains; a
// visibleIf naming a bare "page" path (an unlikely condition, since
// AD-4 means it can only ever resolve against ordinary top-level data,
// never a page-number slot) is not a reservation conflict at all.
func checkVisibleIfExpression(raw string, elementID template.ElementID) error {
	e, perr := expr.Parse(raw)
	if perr != nil {
		return fmt.Errorf("folio: ParseTemplate: element %s: visibleIf %q is not a valid expression: %s", elementID, raw, perr)
	}
	if cerr := expr.Check(e); cerr != nil {
		return fmt.Errorf("folio: ParseTemplate: element %s: visibleIf: %s", elementID, cerr)
	}
	return nil
}

// validateTableColumns checks every column's bind (same static rules
// as checkTextExpressions) and, for a column requesting a sum/avg
// footer with footerOf omitted, runs D-1.4.1's derivation (AC21):
// derivable, the resolved value is recorded in derived; not derivable,
// a load error naming the column id (AC21's rejection arm — a bind
// like {{if(row.x, row.a, row.b)}} does not name a single numeric
// source the way the two derivable shapes do).
//
// footer: "count" is special (D-1.4.1): it always counts the table's
// own bound collection, never a per-column numeric source, so no
// derivation is attempted for it at all — and footerOf present
// alongside footer: "count" is already a load error at the template
// layer (AC43 check 1, internal/template/parse_bands.go).
func validateTableColumns(tbl template.TableExt, derived map[template.ElementID]expr.DerivedFooter) error {
	alias := resolvedRowAlias(tbl.As)
	collection := strings.TrimSuffix(tbl.Bind, "[]")

	for _, col := range tbl.Columns {
		if err := checkTextExpressions(col.Bind, col.ID); err != nil {
			return err
		}

		if !col.Footer.Set || col.FooterOf.Set {
			continue
		}
		if col.Footer.Value != "sum" && col.Footer.Value != "avg" {
			continue // "count" never derives a numeric source
		}

		result, derivable, err := expr.DeriveFooterOf(col.Bind, alias, collection)
		if err != nil {
			return fmt.Errorf("folio: ParseTemplate: column %s: %s", col.ID, err)
		}
		if !derivable {
			return fmt.Errorf(
				"folio: ParseTemplate: column %s: footer %q requires footerOf, and bind %q is not one of D-1.4.1's two derivable shapes — "+
					"name the numeric source explicitly with footerOf",
				col.ID, col.Footer.Value, col.Bind,
			)
		}
		derived[col.ID] = result
	}
	return nil
}
