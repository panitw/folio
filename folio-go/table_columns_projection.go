package folio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/pagemodel"
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

	// STORY 12.3 — the table-level header and row properties, twenty-six
	// members since Story 14.8, and the arithmetic is 12 x 2 + 1 + 1.
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
	//
	// AND, SINCE STORY 11.3, THE TWO BOOLEANS — a SECOND disclosed
	// limit, of the same kind and a wider one. `false` is the zero
	// value, so HeaderBold/HeaderItalic collapse committed-ABSENT with
	// committed-`false`, and unlike the fontSize case a command CAN
	// write the losing value: `tableHeaderStyleFields` carries `bold`
	// and `italic`, and `{"bold": null}` decodes to present(false).
	// The collapse is accepted rather than answered with a tri-state
	// for one field, on D-11.3.3's ground: CanvasProjection is
	// engine<->browser, both in this repo, moving in one commit — it is
	// NOT tag-bound, so a tri-state can be added the day something needs
	// one, and consistency inside one struct beats a second idiom until
	// then. Nothing today needs one: the panel's third state is
	// declared-true-but-no-face, which survives the collapse. AN
	// UNDISCLOSED LIMIT AGES INTO FALSE REASSURANCE, so it is disclosed.
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
	HeaderBold                bool   `json:"headerBold"`
	HeaderBoldResolved        bool   `json:"headerBoldResolved"`
	HeaderItalic              bool   `json:"headerItalic"`
	HeaderItalicResolved      bool   `json:"headerItalicResolved"`

	// STORY 14.8's THREE PAIRS, AND THE DOTTED KEY NAMES ARE FORCED RATHER
	// THAN CHOSEN. `tableHeaderStyleFields` spells the authorable attributes
	// `border.width`, `border.color` and `border.edges` — dotted, so that
	// `table.headerStyle.` + field is a path the document actually has —
	// and table_header_style_test.go derives THESE key names from THOSE
	// field names by string transformation (strip `header`, lowercase the
	// first letter) and requires the two sets to be equal. So the dot
	// propagates out of the command spelling, into the json tag, into the
	// browser's hasExactKeys list, and into quoted TypeScript property
	// access. There is no independent naming decision here to make.
	//
	// ⚠ THE RESOLVED TRIO IS NOT A FIELD-BY-FIELD CASCADE, and that is the
	// one thing about these six members a reader has to know.
	// resolveHeaderStyle takes the header's border WHOLE: the moment
	// `headerStyle.border` exists at all, the table's own `style.border`
	// stops contributing, and every sub-key the header does not declare
	// falls to the FORMAT's default (0.5pt, #000000, all four edges) rather
	// than to the table's value for it. These members report that, computed
	// by the same three functions the RENDERER calls (table_render.go's
	// resolvedBorderWidth/Color/Edges) so the panel cannot show a number the
	// PDF does not draw. The panel discloses the block-granularity in words;
	// nothing here softens it.
	//
	// HeaderBorderEdgesResolved IS EMPTY EXACTLY WHEN NOTHING IS PAINTED —
	// both when no border resolves at all and when a declared `edges: []`
	// leaves no side to stroke — and it is the ONLY member that can say
	// so about the PAINT. The colour members spell "no border resolves"
	// as this projection's ordinary absence, "".
	//
	// ⚠ AND THE WIDTH PAIR IS THE ONE PLACE ON THIS PROJECTION WHERE A
	// LENGTH IS SPELLED AS A STRING RATHER THAN AS THOUSANDTHS OF A POINT.
	// "" is absent, "0" is a declared zero, "500" is a declared half point.
	// THE UNITS ARE UNCHANGED — still integer thousandths, exactly as
	// HeaderHeight and HeaderFontSize — only the SPELLING OF ABSENCE moves.
	//
	// The reason is the loader's own, and it is why this pair differs from
	// HeaderHeight rather than being inconsistent with it.
	// internal/template/parse_bands.go says of a border width: "ZERO IS
	// VALID and stays accepted: it is the thinnest device line PDF can
	// draw, not an absent border. Only a NEGATIVE width is refused." So a
	// header border authored as nothing but `{"width": 0}` is a real,
	// declared, painted border — and a numeric member whose absence is
	// spelled 0 cannot tell it apart from a header that declares no border
	// at all. It did not, and the panel therefore told an author "nothing
	// here is set, so this header row takes the table's own border" about a
	// header that had taken the border over. HeaderHeight never needed this
	// spelling because a zero header height is not a meaningful declaration
	// of anything; a zero border width is.
	//
	// BOTH HALVES OF THE PAIR ARE STRINGS, and the resolved half is
	// formatted in Go. Every other pair on this struct shares one type
	// across the pair (HeaderFontSize/HeaderFontSizeResolved are both
	// int64); a committed string beside a resolved number would be the
	// first breach of that, and it would breach it on the one pair a reader
	// is most likely to mis-read. Both-strings also makes the border trio
	// uniform — six string members, "" meaning absent throughout, the same
	// as the colour and edge pairs beside them.
	HeaderBorderWidth         string `json:"headerBorder.width"`
	HeaderBorderWidthResolved string `json:"headerBorder.widthResolved"`
	HeaderBorderColor         string `json:"headerBorder.color"`
	HeaderBorderColorResolved string `json:"headerBorder.colorResolved"`
	HeaderBorderEdges         string `json:"headerBorder.edges"`
	HeaderBorderEdgesResolved string `json:"headerBorder.edgesResolved"`

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

// committedStyleBool is the same reading for a Presence[bool], and it is
// the site of the collapse the struct comment discloses: absent, null and
// an explicit `false` all come back `false`, because `false` is the only
// spelling of absence a bool has on a wire whose key set is pinned exactly
// in both directions. It is its own function rather than an inline
// expression so there is ONE place to change if that stops being
// acceptable.
func committedStyleBool(value template.Presence[bool]) bool {
	return value.Set && !value.Null && value.Value
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
		HeaderBold:                committedStyleBool(committed.Bold),
		HeaderBoldResolved:        resolved.bold,
		HeaderItalic:              committedStyleBool(committed.Italic),
		HeaderItalicResolved:      resolved.italic,
		HeaderBorderWidth:         committedBorderWidth(committedHeaderBorder(committed).Width),
		HeaderBorderWidthResolved: resolvedHeaderBorderWidth(resolved),
		HeaderBorderColor:         committedStyleString(committedHeaderBorder(committed).Color),
		HeaderBorderColorResolved: resolvedHeaderBorderColor(resolved),
		HeaderBorderEdges:         canonicalEdgeList(committedHeaderBorder(committed).Edges),
		HeaderBorderEdgesResolved: resolvedHeaderBorderEdges(resolved),
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

// committedBorderWidth is committedLength's STRING sibling, and it exists for
// exactly one member. It reads the SAME UNITS — integer thousandths of a point
// — and differs only in how it spells absence: "" rather than 0.
//
// It cannot be committedLength, because for a border width 0 is a LEGAL
// DECLARED VALUE. parse_bands.go: "ZERO IS VALID and stays accepted: it is the
// thinnest device line PDF can draw, not an absent border." So `{"width": 0}`
// and no `width` at all are two different documents that a numeric member
// reports identically — and the panel's "nothing here is set, so this header
// row takes the table's own border" was said about the first of them, which is
// false. Absence needs a spelling of its own here; it does not for
// committedLength's other caller, because a zero font size is not a
// declaration.
func committedBorderWidth(value template.Presence[geom.Length]) string {
	if value.Set && !value.Null {
		return strconv.FormatInt(int64(value.Value), 10)
	}
	return ""
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

// committedHeaderBorder reads the header's border block AS THE DOCUMENT
// DECLARES IT. An absent or explicitly null `border` yields the zero Border,
// whose three Presence members all read as absent — which is exactly what it
// means, and what makes "clear this attribute back to absent" expressible.
func committedHeaderBorder(committed template.Style) template.Border {
	if !committed.Border.Set || committed.Border.Null {
		return template.Border{}
	}
	return committed.Border.Value
}

// THE THREE RESOLVED BORDER MEMBERS, AND ALL THREE ASK THE RENDERER'S OWN
// FUNCTIONS. resolvedBorderWidth, resolvedBorderColor and resolvedBorderEdges
// live in table_render.go beside the emitter that consumes them, so what this
// projection tells the author is what the PDF draws — by construction, not by
// two implementations agreeing. `resolved.hasBorder` is resolveHeaderStyle's own
// verdict on whether ANY border reaches the header row.
// resolvedHeaderBorderWidth is a STRING for the same reason its committed twin
// is — the pair shares one type, as every pair on this struct does — and the
// formatting happens HERE, in Go, so the browser is never the place a length
// acquires a spelling. Thousandths, unchanged; "" only when no border reaches
// the header row at all.
func resolvedHeaderBorderWidth(resolved resolvedHeaderStyle) string {
	if !resolved.hasBorder {
		return ""
	}
	return strconv.FormatInt(int64(resolvedBorderWidth(resolved.border)), 10)
}

func resolvedHeaderBorderColor(resolved resolvedHeaderStyle) string {
	if !resolved.hasBorder {
		return ""
	}
	return resolvedBorderColor(resolved.border)
}

// resolvedHeaderBorderEdges is the member that carries "nothing is painted",
// and it says so by being EMPTY — for both of the two ways that happens: no
// border resolves at all, and a border whose declared `edges` names no side.
// internal/pdf/rectdoc.go's emission gate is exactly that disjunction, so an
// empty string here and no stroke in the PDF are the same condition.
func resolvedHeaderBorderEdges(resolved resolvedHeaderStyle) string {
	if !resolved.hasBorder {
		return ""
	}
	return edgeListOf(resolvedBorderEdges(resolved.border))
}

// canonicalEdgeList and edgeListOf are the ONE spelling of an edge set on this
// wire: the four names in the format's own order, comma-joined, and "" for none.
// The order is canonical rather than the author's, because the browser's guard
// admits a canonical list and a re-ordered one would be refused — and because a
// set has no order to preserve. An unknown name in a hand-edited document is
// dropped here exactly as the renderer's own switch drops it; the loader is the
// door that refuses it, and this projection is not a second one.
func canonicalEdgeList(edges template.Presence[[]string]) string {
	if !edges.Set || edges.Null {
		return ""
	}
	declared := pagemodel.RectEdges{}
	for _, edge := range edges.Value {
		switch edge {
		case "top":
			declared.Top = true
		case "right":
			declared.Right = true
		case "bottom":
			declared.Bottom = true
		case "left":
			declared.Left = true
		}
	}
	return edgeListOf(declared)
}

func edgeListOf(edges pagemodel.RectEdges) string {
	names := make([]string, 0, 4)
	for _, edge := range []struct {
		name string
		on   bool
	}{{"top", edges.Top}, {"right", edges.Right}, {"bottom", edges.Bottom}, {"left", edges.Left}} {
		if edge.on {
			names = append(names, edge.name)
		}
	}
	return strings.Join(names, ",")
}
