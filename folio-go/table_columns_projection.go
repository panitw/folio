package folio

import (
	"fmt"

	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/template"
)

const maxTableColumns = 128

// TableColumnsProjection is a bounded paint/editor projection. It deliberately
// carries only the bounded table configuration the focused editor can paint.
// It never contains canonical bytes, sample data, parsed sample shape, or
// aggregate values.
type TableColumnsProjection struct {
	TableID    string `json:"tableId"`
	Collection string `json:"collection"`
	Alias      string `json:"alias"`

	// STORY 12.3 — the table-level header and row properties, sixteen
	// members, and the arithmetic is 7 x 2 + 1 + 1.
	//
	// TWO MEMBERS PER HEADER-STYLE FIELD. The committed one is what the
	// document actually declares — empty/zero when the key is absent —
	// so the control can tell SET from UNSET and "clear it back to
	// absent" stays a thing the author can express. The resolved one is
	// what the document will USE, which for an absent field is the
	// table's own `style.<field>` and then that field's documented
	// default. One member cannot serve both: a single resolved value
	// makes absence unrepresentable, and a single committed value
	// forces the browser to run the cascade (D-12.3.1).
	//
	// THE RESOLVED HALF IS THE ENGINE'S OWN ANSWER, not a second
	// implementation of it. resolveHeaderStyle (table_render.go) is the
	// ONE cascade in this program and it is called here, unchanged and
	// unexported, because this file is package folio too. Nothing in
	// TypeScript may re-derive, mirror or approximate it (AC2, AC3,
	// AD-15, AD-17).
	//
	// HeaderHeight AND AltRowBackground CARRY ONE MEMBER EACH, and that
	// asymmetry is deliberate — do not "fix" it. A resolved member
	// answers "what will be used when this is absent". HeaderHeight is
	// REQUIRED (parse_bands.go hard-errors on its absence), so it is
	// never absent; AltRowBackground is a flat override on odd
	// zero-based collection indexes with no fallback level of its own,
	// so absence resolves to nothing rather than to an inherited value.
	// For both, the question has no content and committed IS resolved.
	//
	// ABSENCE IS SPELLED AS THE ZERO VALUE, the same convention
	// TableColumnProjection already uses for Binding, Footer, FooterOf
	// and FooterFormat: "" for a string, 0 for a length or a ratio. It
	// is the only spelling available to a wire record whose key set is
	// pinned exactly in both directions (canvas_projection_wire_test.go),
	// where an `omitempty` would make a key appear only sometimes and
	// the browser's hasExactKeys reject exactly those documents. The one
	// place it is lossy is a hand-edited `headerStyle.fontSize: 0`,
	// which no command can write (the arm refuses a non-positive size)
	// and which DW-26 already records as unbounded at the loader.
	HeaderHeight     int64  `json:"headerHeight"`
	AltRowBackground string `json:"altRowBackground"`

	HeaderFontFamily          string `json:"headerFontFamily"`
	HeaderFontFamilyResolved  string `json:"headerFontFamilyResolved"`
	HeaderFontSize            int64  `json:"headerFontSize"`
	HeaderFontSizeResolved    int64  `json:"headerFontSizeResolved"`
	HeaderLineSpacing         int64  `json:"headerLineSpacing"`
	HeaderLineSpacingResolved int64  `json:"headerLineSpacingResolved"`
	HeaderBackground          string `json:"headerBackground"`
	HeaderBackgroundResolved  string `json:"headerBackgroundResolved"`
	HeaderColor               string `json:"headerColor"`
	HeaderColorResolved       string `json:"headerColorResolved"`
	HeaderValign              string `json:"headerValign"`
	HeaderValignResolved      string `json:"headerValignResolved"`
	HeaderAlign               string `json:"headerAlign"`
	HeaderAlignResolved       string `json:"headerAlignResolved"`

	Columns []TableColumnProjection `json:"columns"`
}

// committedHeaderStyle is the header style block AS THE DOCUMENT
// DECLARES IT — never cascaded, never defaulted. An absent or explicitly
// null `headerStyle` yields the zero Style, which reads as "every field
// absent", which is exactly what it means.
//
// An explicitly null FIELD inside a present block (`headerStyle: {"color":
// null}`, reachable only by hand-editing) also reads as absent here: the
// projection has one member for "what is committed" and no third state to
// put a null in. The cascade's own answer for that case still travels
// intact in the resolved member beside it, because resolveHeaderStyle
// falls through a null exactly as it falls through an absent field.
func committedHeaderStyle(table template.TableExt) template.Style {
	if !table.HeaderStyle.Set || table.HeaderStyle.Null {
		return template.Style{}
	}
	return table.HeaderStyle.Value
}

// committedStyleString reads one Presence[string] the way the projection
// spells absence: the value when it is genuinely set, "" otherwise.
func committedStyleString(value template.Presence[string]) string {
	if value.Set && !value.Null {
		return value.Value
	}
	return ""
}

type TableColumnProjection struct {
	ID               string `json:"id"`
	Header           string `json:"header"`
	Width            int64  `json:"width"`
	Align            string `json:"align"`
	Binding          string `json:"binding"`
	RowField         string `json:"rowField"`
	RowFieldEditable bool   `json:"rowFieldEditable"`
	Footer           string `json:"footer"`
	FooterOf         string `json:"footerOf"`
	FooterFormat     string `json:"footerFormat"`
}

// TableColumns returns only the selected table's structural column paint
// state. The browser may display it but cannot use it as a template model.
func TableColumns(t *Template, tableID string) (TableColumnsProjection, error) {
	if t == nil {
		return TableColumnsProjection{}, errNilTemplate
	}
	if tableID == "" || len(tableID) > 128 {
		return TableColumnsProjection{}, fmt.Errorf("folio: table id is invalid")
	}
	_, _, _, element, err := findComponent(t, tableID)
	if err != nil {
		return TableColumnsProjection{}, fmt.Errorf("folio: table was not found")
	}
	if element.Type != template.ElementTable || !element.Table.Set || element.Table.Null {
		return TableColumnsProjection{}, fmt.Errorf("folio: component is not a table")
	}
	if len(element.Table.Value.Columns) > maxTableColumns {
		return TableColumnsProjection{}, fmt.Errorf("folio: table has too many columns for editor projection")
	}
	collection := element.Table.Value.Bind
	alias := "row"
	if element.Table.Value.As.Set && !element.Table.Value.As.Null {
		alias = element.Table.Value.As.Value
	}
	if collection == "" || !rootCollectionPath.MatchString(collection) || len(collection) > 256 || alias == "" || len(alias) > 64 || !boundedIdentifier.MatchString(alias) || reservedRowAlias(alias) {
		return TableColumnsProjection{}, fmt.Errorf("folio: table configuration cannot be projected")
	}
	// THE CASCADE, ASKED ONCE, HERE. resolveHeaderStyle is the engine's
	// only header cascade; calling it is what makes the resolved members
	// the engine's answer rather than the browser's guess.
	resolved := resolveHeaderStyle(*element)
	committed := committedHeaderStyle(element.Table.Value)
	projection := TableColumnsProjection{
		TableID:                   tableID,
		Collection:                collection,
		Alias:                     alias,
		HeaderHeight:              int64(element.Table.Value.HeaderHeight),
		AltRowBackground:          committedStyleString(element.Table.Value.AltRowBackground),
		HeaderFontFamily:          committedStyleString(committed.FontFamily),
		HeaderFontFamilyResolved:  resolvedHeaderFontFamily(resolved),
		HeaderFontSize:            committedLength(committed.FontSize),
		HeaderFontSizeResolved:    int64(resolved.fontSize),
		HeaderLineSpacing:         committedRatio(committed.LineSpacing),
		HeaderLineSpacingResolved: resolved.lineSpacing,
		HeaderBackground:          committedStyleString(committed.Background),
		HeaderBackgroundResolved:  resolvedHeaderBackground(resolved),
		HeaderColor:               committedStyleString(committed.Color),
		HeaderColorResolved:       committedStyleString(resolved.inkStyle.Color),
		HeaderValign:              committedStyleString(committed.Valign),
		HeaderValignResolved:      resolved.valign,
		HeaderAlign:               committedStyleString(committed.Align),
		HeaderAlignResolved:       resolved.alignFallback,
		Columns:                   make([]TableColumnProjection, 0, len(element.Table.Value.Columns)),
	}
	for _, column := range element.Table.Value.Columns {
		if len(column.Label) > 256 || column.Width <= 0 || len(column.ID) == 0 || len(column.ID) > 128 {
			return TableColumnsProjection{}, fmt.Errorf("folio: table column cannot be projected")
		}
		align := "left"
		if column.Align.Set && !column.Align.Null {
			align = column.Align.Value
		}
		if align != "left" && align != "center" && align != "right" {
			return TableColumnsProjection{}, fmt.Errorf("folio: table column cannot be projected")
		}
		footer, footerOf, footerFormat := "", "", ""
		if column.Footer.Set && !column.Footer.Null {
			footer = column.Footer.Value
		}
		if column.FooterOf.Set && !column.FooterOf.Null {
			footerOf = column.FooterOf.Value
		}
		if column.FooterFormat.Set && !column.FooterFormat.Null {
			footerFormat = column.FooterFormat.Value
		}
		if footer != "" && footer != "sum" && footer != "avg" && footer != "count" || (footer == "" && (footerOf != "" || footerFormat != "")) || (footer == "count" && footerOf != "") || len(column.Bind) > 256 || len(footerOf) > 256 || len(footerFormat) > 256 {
			return TableColumnsProjection{}, fmt.Errorf("folio: table column cannot be projected")
		}
		// A new column has no bind yet. It is deliberately editable so the
		// normal Table Editor can complete it through the Go command boundary;
		// once a non-empty expression exists, retain the stricter projection
		// rules for arbitrary/unsupported expressions.
		row := expr.RowBinding{Editable: column.Bind == ""}
		if column.Bind != "" {
			row, err = expr.ProjectRowBinding(column.Bind, alias)
			if err != nil {
				return TableColumnsProjection{}, fmt.Errorf("folio: table column cannot be projected")
			}
		}
		projection.Columns = append(projection.Columns, TableColumnProjection{ID: string(column.ID), Header: column.Label, Width: int64(column.Width), Align: align, Binding: column.Bind, RowField: row.Field, RowFieldEditable: row.Editable, Footer: footer, FooterOf: footerOf, FooterFormat: footerFormat})
	}
	return projection, nil
}

// committedLength and committedRatio are committedStyleString's two
// numeric siblings: the declared value, or 0 for "the document does not
// declare one".
func committedLength(value template.Presence[geom.Length]) int64 {
	if value.Set && !value.Null {
		return int64(value.Value)
	}
	return 0
}

func committedRatio(value template.Presence[int64]) int64 {
	if value.Set && !value.Null {
		return value.Value
	}
	return 0
}

// resolvedHeaderFontFamily and resolvedHeaderBackground read the two
// cascade results that carry a presence flag beside their value. Empty
// means the cascade found nothing to resolve from — a font chain is
// genuinely optional at this level (the render path raises its own
// located error where a header label actually needs one), and a header
// with no background paints none.
func resolvedHeaderFontFamily(resolved resolvedHeaderStyle) string {
	if resolved.hasFontFamily {
		return resolved.fontFamily
	}
	return ""
}

func resolvedHeaderBackground(resolved resolvedHeaderStyle) string {
	if resolved.hasBackground {
		return resolved.background
	}
	return ""
}
