package template

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/panitw/folio/folio-go/internal/diag"
	"github.com/panitw/folio/folio-go/internal/geom"
)

// decodeBands enforces AC5's "exactly these three keys" (FR6):
// content, pageFooter, pageHeader. Unlike an unknown key elsewhere in
// the document, a fourth band name or a missing one is a load error,
// never passthrough — bands is a closed set of exactly three names,
// not a place AC9 lists for opaque passthrough.
func decodeBands(ctx *parseCtx, raw json.RawMessage) (Bands, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Bands{}, fmt.Errorf("template: bands: %w", err)
	}
	want := []string{"content", "pageFooter", "pageHeader"}
	consumed := map[string]bool{}
	for _, k := range want {
		if _, ok := obj[k]; !ok {
			return Bands{}, newLoadError("bands."+k, "", "", "missing required band (FR6: bands must have exactly content, pageFooter, pageHeader)")
		}
		consumed[k] = true
	}
	if extra := unexpectedKeys(obj, consumed); len(extra) > 0 {
		return Bands{}, newLoadError("bands", "", strings.Join(extra, ","), "must have exactly the three keys content, pageFooter, pageHeader — no others (FR6)")
	}

	content, err := decodeBand(ctx, contentBandField, obj["content"], false)
	if err != nil {
		return Bands{}, err
	}
	footer, err := decodeBand(ctx, "bands.pageFooter", obj["pageFooter"], true)
	if err != nil {
		return Bands{}, err
	}
	header, err := decodeBand(ctx, "bands.pageHeader", obj["pageHeader"], true)
	if err != nil {
		return Bands{}, err
	}
	return Bands{Content: content, PageFooter: footer, PageHeader: header}, nil
}

// contentBandField is decodeBands' own name for the content band, and
// the value decodeElement's bandField carries for an element declared
// there. Named once rather than spelled twice, because Story 7.7's
// keep-together refusal turns on exactly this comparison.
const contentBandField = "bands.content"

// decodeBand decodes one band object. hasHeight controls whether a
// "height" key is required (pageHeader/pageFooter) or forbidden
// (content — AC5: "not on content", :94-96, "storing it would be a
// second source of truth").
func decodeBand(ctx *parseCtx, field string, raw json.RawMessage, hasHeight bool) (Band, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Band{}, fmt.Errorf("template: %s: %w", field, err)
	}
	consumed := map[string]bool{"elements": true}

	elemsRaw, ok := obj["elements"]
	if !ok {
		return Band{}, newLoadError(field+".elements", "", "", "missing required field")
	}
	items, err := decodeArrayRaw(elemsRaw)
	if err != nil {
		return Band{}, newLoadError(field+".elements", "", string(elemsRaw), "must be an array: "+err.Error())
	}
	elems := make([]Element, 0, len(items))
	for _, it := range items {
		el, err := decodeElement(ctx, field, it)
		if err != nil {
			return Band{}, err
		}
		elems = append(elems, el)
	}

	var height Presence[geom.Length]
	if hRaw, ok := obj["height"]; ok {
		consumed["height"] = true
		if !hasHeight {
			return Band{}, newLoadError(field+".height", "", string(hRaw), "content must not declare a height — it is derived (page height minus margins minus header minus footer)")
		}
		if rawIsNull(hRaw) {
			return Band{}, newLoadError(field+".height", "", "null", "height must not be null")
		}
		v, err := decodePointsRaw(field+".height", "", hRaw)
		if err != nil {
			return Band{}, err
		}
		height = present(v)
	} else if hasHeight {
		return Band{}, newLoadError(field+".height", "", "", "missing required field")
	}

	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Band{}, fmt.Errorf("template: %s: %w", field, err)
	}

	return Band{Elements: elems, Height: height, Extra: extra}, nil
}

func decodeElement(ctx *parseCtx, bandField string, raw json.RawMessage) (Element, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Element{}, fmt.Errorf("template: %s.elements[]: %w", bandField, err)
	}

	idRaw, ok := obj["id"]
	if !ok {
		return Element{}, newLoadError(bandField+".elements[].id", "", "", "missing required field")
	}
	idStr, err := decodeStringRaw(idRaw)
	if err != nil {
		return Element{}, newLoadError(bandField+".elements[].id", "", string(idRaw), "must be a string: "+err.Error())
	}
	id, err := ctx.claimID(bandField+".elements[].id", idStr)
	if err != nil {
		return Element{}, err
	}

	consumed := map[string]bool{"id": true, "type": true, "x": true, "y": true, "keepTogether": true}

	typeRaw, ok := obj["type"]
	if !ok {
		return Element{}, newLoadError("type", string(id), "", "missing required field")
	}
	typeStr, err := decodeStringRaw(typeRaw)
	if err != nil {
		return Element{}, newLoadError("type", string(id), string(typeRaw), "must be a string: "+err.Error())
	}
	if !closedElementTypes[typeStr] {
		return Element{}, newLoadError("type", string(id), typeStr, "not one of the closed set text, image, table, line, rect (FR4)")
	}
	el := Element{ID: id, Type: ElementType(typeStr)}

	xRaw, ok := obj["x"]
	if !ok {
		return Element{}, newLoadError("x", string(id), "", "missing required field")
	}
	el.X, err = decodePointsRaw("x", string(id), xRaw)
	if err != nil {
		return Element{}, err
	}
	yRaw, ok := obj["y"]
	if !ok {
		return Element{}, newLoadError("y", string(id), "", "missing required field")
	}
	el.Y, err = decodePointsRaw("y", string(id), yRaw)
	if err != nil {
		return Element{}, err
	}

	wRaw, wok := obj["width"]
	hRaw, hok := obj["height"]
	if el.Type == ElementTable {
		if wok {
			return Element{}, newLoadError("width", string(id), string(wRaw), "a table declares x and y only — never width (AD-13, AC5)")
		}
		if hok {
			return Element{}, newLoadError("height", string(id), string(hRaw), "a table declares x and y only — never height (AD-13, AC5)")
		}
	} else {
		if !wok {
			return Element{}, newLoadError("width", string(id), "", "missing required field")
		}
		if !hok {
			return Element{}, newLoadError("height", string(id), "", "missing required field")
		}
		consumed["width"] = true
		consumed["height"] = true
		v, err := decodePointsRaw("width", string(id), wRaw)
		if err != nil {
			return Element{}, err
		}
		el.Width = present(v)
		v, err = decodePointsRaw("height", string(id), hRaw)
		if err != nil {
			return Element{}, err
		}
		el.Height = present(v)
	}

	if viRaw, ok := obj["visibleIf"]; ok {
		consumed["visibleIf"] = true
		if rawIsNull(viRaw) {
			el.VisibleIf = presentNull[string]()
		} else {
			s, err := decodeStringRaw(viRaw)
			if err != nil {
				return Element{}, newLoadError("visibleIf", string(id), string(viRaw), "must be a string expression: "+err.Error())
			}
			el.VisibleIf = present(s)
		}
	}

	if styRaw, ok := obj["style"]; ok {
		consumed["style"] = true
		if rawIsNull(styRaw) {
			el.Style = presentNull[Style]()
		} else {
			st, err := decodeStyle(string(id), el.Type, styRaw, "style")
			if err != nil {
				return Element{}, err
			}
			el.Style = present(st)
		}
	}

	// Story 7.7 (FR51): the author-declared keep-together tag. Decoded
	// on `visibleIf`'s exact shape — absent and explicit null both mean
	// "no tag declared" — with two refusals that are local to the
	// element being decoded, which is why the declaration is an element
	// key at all (D-7.7 Ruling B).
	//
	//  1. NOT OUTSIDE THE CONTENT BAND. FR51 scopes the feature to
	//     content-band elements; a page header or page footer is
	//     repeated verbatim on every page and is never paginated, so a
	//     tag there could never do anything. bandField is decodeBands'
	//     own name for the band being decoded, so this is the same hook
	//     the band-scoped `height` refusal above uses.
	//  2. NOT ON A TABLE. A table's own items already carry a row key
	//     (tableRectSource.chromeRowGroup / textRunSource.lineRowGroup),
	//     and layout.ColumnItem.Group is ONE key per item — honouring
	//     both would be a second grouping model, which this story's
	//     contract forbids outright.
	//
	// An EMPTY tag is refused everywhere: the value names a group, and
	// "" names nothing while silently joining every other ""-tagged
	// element into one union extent.
	//
	// ⚠ THE TWO REFUSALS ABOVE ARE ABOUT THE KEY, NOT ITS VALUE, and so
	// they are tested BEFORE the null branch. The contract's matrix rows
	// say "the key on a pageHeader or pageFooter element" and "the key on
	// type: table" — not "a non-empty key" — and there is no reading on
	// which `"keepTogether": null` on one of those elements could ever be
	// honoured. Leaving it inside the non-null arm accepted it silently
	// AND raised the document to 1.2 (versionRequiredByContent probes
	// KeepTogether.Set, which an explicit null sets) for a key that can
	// never do anything. Where the key IS allowed — a content-band
	// non-table element — null keeps meaning exactly "ungrouped".
	if ktRaw, ok := obj["keepTogether"]; ok {
		if bandField != contentBandField {
			return Element{}, newLoadError("keepTogether", string(id), string(ktRaw), "valid only on a content-band element (FR51) — a page header or page footer is repeated verbatim on every page and is never paginated, so a keep-together group there could never be honoured")
		}
		if el.Type == ElementTable {
			return Element{}, newLoadError("keepTogether", string(id), string(ktRaw), "not valid on a table — a table's rows already carry their own grouping identity, and one item can belong to only one group, so a tagged table would need a second grouping model")
		}
		if rawIsNull(ktRaw) {
			el.KeepTogether = presentNull[string]()
		} else {
			s, err := decodeStringRaw(ktRaw)
			if err != nil {
				return Element{}, newLoadError("keepTogether", string(id), string(ktRaw), "must be a string naming the keep-together group: "+err.Error())
			}
			if s == "" {
				return Element{}, newLoadError("keepTogether", string(id), `""`, "must name a group — an empty tag names none (FR51)")
			}
			el.KeepTogether = present(s)
		}
	}

	switch el.Type {
	case ElementText:
		vRaw, ok := obj["value"]
		if !ok {
			return Element{}, newLoadError("value", string(id), "", "missing required field for a text element")
		}
		consumed["value"] = true
		if rawIsNull(vRaw) {
			el.Value = presentNull[string]()
		} else {
			s, err := decodeStringRaw(vRaw)
			if err != nil {
				return Element{}, newLoadError("value", string(id), string(vRaw), "must be a string: "+err.Error())
			}
			el.Value = present(s)
		}
	case ElementImage:
		aRaw, ok := obj["asset"]
		if !ok {
			return Element{}, newLoadError("asset", string(id), "", "missing required field for an image element")
		}
		consumed["asset"] = true
		if rawIsNull(aRaw) {
			// An image box the author placed but has not filled yet. The
			// field stays required — a silently absent asset is still a
			// load error — but null says "no image chosen", which Render
			// draws nothing for and the designer shows as a placeholder.
			el.Asset = presentNull[string]()
		} else {
			s, err := decodeStringRaw(aRaw)
			if err != nil {
				return Element{}, newLoadError("asset", string(id), string(aRaw), "must be a string: "+err.Error())
			}
			el.Asset = present(s)
		}
	case ElementTable:
		tbl, tblConsumed, err := decodeTableExt(ctx, string(id), obj)
		if err != nil {
			return Element{}, err
		}
		el.Table = present(tbl)
		for _, k := range slices.Sorted(maps.Keys(tblConsumed)) {
			consumed[k] = true
		}
	case ElementLine, ElementRect:
		// no extra fields
	}

	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Element{}, fmt.Errorf("template: element %s: %w", id, err)
	}
	el.Extra = extra

	return el, nil
}

func decodeTableExt(ctx *parseCtx, id string, obj map[string]json.RawMessage) (TableExt, map[string]bool, error) {
	consumed := map[string]bool{"bind": true, "columns": true, "headerHeight": true}
	var t TableExt

	bindRaw, ok := obj["bind"]
	if !ok {
		return TableExt{}, nil, newLoadError("bind", id, "", "missing required field for a table")
	}
	bind, err := decodeStringRaw(bindRaw)
	if err != nil {
		return TableExt{}, nil, newLoadError("bind", id, string(bindRaw), "must be a string: "+err.Error())
	}
	t.Bind = bind

	if asRaw, ok := obj["as"]; ok {
		consumed["as"] = true
		s, err := decodeStringRaw(asRaw)
		if err != nil {
			return TableExt{}, nil, newLoadError("as", id, string(asRaw), "must be a string: "+err.Error())
		}
		t.As = present(s)
	}

	colsRaw, ok := obj["columns"]
	if !ok {
		return TableExt{}, nil, newLoadError("columns", id, "", "missing required field for a table")
	}
	items, err := decodeArrayRaw(colsRaw)
	if err != nil {
		return TableExt{}, nil, newLoadError("columns", id, string(colsRaw), "must be an array: "+err.Error())
	}
	collection := strings.TrimSuffix(bind, "[]")
	cols := make([]Column, 0, len(items))
	for _, it := range items {
		col, err := decodeColumn(ctx, id, collection, it)
		if err != nil {
			return TableExt{}, nil, err
		}
		cols = append(cols, col)
	}
	t.Columns = cols

	hhRaw, ok := obj["headerHeight"]
	if !ok {
		return TableExt{}, nil, newLoadError("headerHeight", id, "", "missing required field for a table")
	}
	hh, err := decodePointsRaw("headerHeight", id, hhRaw)
	if err != nil {
		return TableExt{}, nil, err
	}
	t.HeaderHeight = hh

	if arbRaw, ok := obj["altRowBackground"]; ok {
		consumed["altRowBackground"] = true
		s, err := decodeStringRaw(arbRaw)
		if err != nil {
			return TableExt{}, nil, newLoadError("altRowBackground", id, string(arbRaw), "must be a string: "+err.Error())
		}
		t.AltRowBackground = present(s)
	}

	// headerStyle (Story 4.1, owner ruling): same Style vocabulary and
	// the same null-vs-absent handling as an element's own "style"
	// above — a header-only override, never required.
	if hsRaw, ok := obj["headerStyle"]; ok {
		consumed["headerStyle"] = true
		if rawIsNull(hsRaw) {
			t.HeaderStyle = presentNull[Style]()
		} else {
			// Statically a table: decodeTableExt's only caller sits
			// under `case ElementTable:`. The constant is passed
			// rather than a variable because there is no type in
			// scope to pass, and a headerStyle exists on nothing else.
			hs, err := decodeStyle(id, ElementTable, hsRaw, "headerStyle")
			if err != nil {
				return TableExt{}, nil, err
			}
			t.HeaderStyle = present(hs)
		}
	}

	return t, consumed, nil
}

// decodeColumn decodes one table column, including the footer schema's
// three decidable checks (AC42, AC43, AC44). collection is the table's
// own bind with "[]" stripped (D-1.4.1).
func decodeColumn(ctx *parseCtx, tableID, collection string, raw json.RawMessage) (Column, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Column{}, fmt.Errorf("template: table %s: column: %w", tableID, err)
	}

	idRaw, ok := obj["id"]
	if !ok {
		return Column{}, newLoadError("columns[].id", tableID, "", "missing required field")
	}
	idStr, err := decodeStringRaw(idRaw)
	if err != nil {
		return Column{}, newLoadError("columns[].id", tableID, string(idRaw), "must be a string: "+err.Error())
	}
	id, err := ctx.claimID("columns[].id", idStr)
	if err != nil {
		return Column{}, err
	}

	consumed := map[string]bool{"id": true, "label": true, "width": true, "bind": true}
	var col Column
	col.ID = id

	labelRaw, ok := obj["label"]
	if !ok {
		return Column{}, newLoadError("label", string(id), "", "missing required field")
	}
	col.Label, err = decodeStringRaw(labelRaw)
	if err != nil {
		return Column{}, newLoadError("label", string(id), string(labelRaw), "must be a string: "+err.Error())
	}

	widthRaw, ok := obj["width"]
	if !ok {
		return Column{}, newLoadError("width", string(id), "", "missing required field")
	}
	col.Width, err = decodePointsRaw("width", string(id), widthRaw)
	if err != nil {
		return Column{}, err
	}

	if alignRaw, ok := obj["align"]; ok {
		consumed["align"] = true
		s, err := decodeStringRaw(alignRaw)
		if err != nil {
			return Column{}, newLoadError("align", string(id), string(alignRaw), "must be a string: "+err.Error())
		}
		// THE COLUMN SET (Story 7.3, D-7.3.1): exactly left, center,
		// right. `justify` is a STYLE value and never a column one, and
		// the message is derived from ColumnAlignTokens so it can never
		// name a value this check does not actually admit.
		if !closedColumnAligns[s] {
			return Column{}, newLoadError("align", string(id), s, closedSetMessage(ColumnAlignTokens))
		}
		col.Align = present(s)
	}

	bindRaw, ok := obj["bind"]
	if !ok {
		return Column{}, newLoadError("bind", string(id), "", "missing required field")
	}
	col.Bind, err = decodeStringRaw(bindRaw)
	if err != nil {
		return Column{}, newLoadError("bind", string(id), string(bindRaw), "must be a string: "+err.Error())
	}

	var footer Presence[string]
	if footerRaw, ok := obj["footer"]; ok {
		consumed["footer"] = true
		s, err := decodeStringRaw(footerRaw)
		if err != nil {
			return Column{}, newLoadError("footer", string(id), string(footerRaw), "must be a string: "+err.Error())
		}
		// UNCODED, DELIBERATELY, AND ON THIS GROUND (D-000.67 part 2;
		// Story 4.5's review, Minor 8 corrected the ground recorded for
		// it). Unlike the three checks around it, this is NOT a type
		// failure — the value is a well-formed string. It is a
		// CLOSED-SET violation: a footer KIND outside {sum, count, avg}.
		// TABLE_FOOTER_SOURCE_UNRESOLVED / _FORBIDDEN name a failure of
		// the footer's numeric SOURCE (which collection path its value
		// comes from), and a bad KIND is a different statement entirely,
		// so coding it with either would corrupt both meanings — worse
		// than leaving it to surface as TEMPLATE_MALFORMED through
		// wrapTemplateError, which is what it is.
		if !closedFooterKinds[s] {
			return Column{}, newLoadError("footer", string(id), s, "not one of the closed set sum, count, avg")
		}
		footer = present(s)
	}
	col.Footer = footer

	_, hasFooterOfRaw := obj["footerOf"]
	_, hasFooterFormatRaw := obj["footerFormat"]

	// AC43 check 2: footerOf/footerFormat present with no footer —
	// pure field presence.
	if (hasFooterOfRaw || hasFooterFormatRaw) && !footer.Set {
		return Column{}, newLoadErrorCoded("footerOf/footerFormat", string(id), "", "footerOf/footerFormat present with no footer — load error (D-1.4.2)", diag.CodeTableFooterSourceForbidden)
	}

	if footerOfRaw, ok := obj["footerOf"]; ok {
		consumed["footerOf"] = true
		s, err := decodeStringRaw(footerOfRaw)
		if err != nil {
			return Column{}, newLoadError("footerOf", string(id), string(footerOfRaw), "must be a string: "+err.Error())
		}
		// AC43 check 1: footerOf present with footer: "count" — load error.
		if footer.Set && footer.Value == "count" {
			return Column{}, newLoadErrorCoded("footerOf", string(id), s, `footerOf present alongside footer: "count" is a load error (D-1.4.2: storing it would be a second source of truth against bind, AD-13)`, diag.CodeTableFooterSourceForbidden)
		}
		// AC43 check 3: footerOf must be prefixed by the table's own
		// collection path + "." — a string prefix test, no parser.
		prefix := collection + "."
		if !strings.HasPrefix(s, prefix) {
			// D-1.4.1: TABLE_FOOTER_SOURCE_UNRESOLVED covers "underivable
			// OR out-of-collection source" — this is the out-of-collection
			// arm. Routed here by the engineering lead (Story 4.5): a
			// Story 3.6 absorption gap (the two FORBIDDEN checks beside
			// this one were coded at 3.6; this one was left as a plain
			// newLoadError), swept and closed in this story rather than
			// carried further.
			return Column{}, newLoadErrorCoded("footerOf", string(id), s, fmt.Sprintf("must be prefixed by the table's collection path %q (D-1.4.2)", prefix), diag.CodeTableFooterSourceUnresolved)
		}
		col.FooterOf = present(s)
	}

	if footerFormatRaw, ok := obj["footerFormat"]; ok {
		consumed["footerFormat"] = true
		s, err := decodeStringRaw(footerFormatRaw)
		if err != nil {
			return Column{}, newLoadError("footerFormat", string(id), string(footerFormatRaw), "must be a string: "+err.Error())
		}
		col.FooterFormat = present(s)
	}

	// Story 3.5, AC3: a pure field-presence check, alongside the three
	// AC43 column checks above. Without it, "visibleIf" on a column
	// falls through into Extra below (extraFields absorbs any unknown
	// key opaquely) and round-trips silently — verified by execution at
	// this story's creation, not assumed. AD-24 permits a condition on
	// an ELEMENT only, never on a table ROW: a column IS one row's
	// per-record slice, so a column-level visibleIf is exactly the
	// row-level visibility AD-24 forbids, because it would make
	// pagination a function of data in a way FR25 does not define.
	// folio-format.md has said "Not valid on a table column" since the
	// format was written; this is the first place anything enforces it.
	if _, ok := obj["visibleIf"]; ok {
		return Column{}, newLoadError("visibleIf", string(id), "", "visibility applies to elements only, never a table column/row (AD-24) — a condition here would make pagination a function of data")
	}

	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Column{}, fmt.Errorf("template: table %s column %s: %w", tableID, id, err)
	}
	col.Extra = extra

	return col, nil
}

// decodeStyle decodes a Style block, whichever of the document's two
// attach points it was read from: an element's own "style"
// (fieldPrefix "style") or a table's "headerStyle" (fieldPrefix
// "headerStyle", Story 4.1's owner-ruled scope addition). Every load
// error this function (and decodePadding/decodeBorder, below) raises
// names fieldPrefix rather than a hardcoded "style", so a mistyped
// headerStyle field is located at headerStyle, not at its sibling
// (finisher fix, Story 4.1 review Finding 5 — previously EVERY
// diagnostic raised inside a headerStyle block named "style", sending
// the template author to the wrong block).
//
// elementType is the CONSUMER (Story 7.8). A Style block's meaning is
// not fixed by the key it was written under — a table's `style.align`
// is consumed by the cell renderer, exactly like its `columns[].align`,
// while a text element's is consumed by the paragraph justifier — so
// the closed set `align` validates against is selected by the element
// that owns the block, never by the key path. Both callers already hold
// the type: decodeElement has el.Type live, and decodeTableExt is
// statically a table. See closedsets.go's header for why this is the
// partition that holds.
func decodeStyle(elementID string, elementType ElementType, raw json.RawMessage, fieldPrefix string) (Style, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Style{}, newLoadError(fieldPrefix, elementID, string(raw), "must be an object: "+err.Error())
	}
	consumed := map[string]bool{}
	var st Style

	if r, ok := obj["align"]; ok {
		consumed["align"] = true
		s, err := decodeStringRaw(r)
		if err != nil {
			return Style{}, newLoadError(fieldPrefix+".align", elementID, string(r), "must be a string: "+err.Error())
		}
		// THE SET IS CHOSEN BY THE CONSUMER (Story 7.8), not by the
		// key this block was written under. A table's style.align and
		// headerStyle.align feed the same alignFallback its
		// columns[].align does, so they admit the same three values;
		// every other element's style.align is the paragraph
		// justifier's own input and admits `justify` (FR47). The
		// message is derived from whichever ordered slice did the
		// rejecting, so it can never name `justify` as legal for an
		// element that cannot carry it.
		tokens, admits := StyleAlignTokens, closedStyleAligns
		if elementType == ElementTable {
			tokens, admits = TableStyleAlignTokens, closedTableStyleAligns
		}
		if !admits[s] {
			return Style{}, newLoadError(fieldPrefix+".align", elementID, s, closedSetMessage(tokens))
		}
		st.Align = present(s)
	}
	if r, ok := obj["valign"]; ok {
		consumed["valign"] = true
		s, err := decodeStringRaw(r)
		if err != nil {
			return Style{}, newLoadError(fieldPrefix+".valign", elementID, string(r), "must be a string: "+err.Error())
		}
		if !closedValigns[s] {
			return Style{}, newLoadError(fieldPrefix+".valign", elementID, s, "not one of the closed set top, middle, bottom")
		}
		st.Valign = present(s)
	}
	if r, ok := obj["background"]; ok {
		consumed["background"] = true
		if rawIsNull(r) {
			st.Background = presentNull[string]()
		} else {
			s, err := decodeStringRaw(r)
			if err != nil {
				return Style{}, newLoadError(fieldPrefix+".background", elementID, string(r), "must be a string: "+err.Error())
			}
			st.Background = present(s)
		}
	}
	if r, ok := obj["color"]; ok {
		consumed["color"] = true
		if rawIsNull(r) {
			st.Color = presentNull[string]()
		} else {
			s, err := decodeStringRaw(r)
			if err != nil {
				return Style{}, newLoadError(fieldPrefix+".color", elementID, string(r), "must be a string: "+err.Error())
			}
			st.Color = present(s)
		}
	}
	if r, ok := obj["bold"]; ok {
		consumed["bold"] = true
		b, err := decodeBoolRaw(r)
		if err != nil {
			return Style{}, newLoadError(fieldPrefix+".bold", elementID, string(r), "must be a bool: "+err.Error())
		}
		st.Bold = present(b)
	}
	if r, ok := obj["italic"]; ok {
		consumed["italic"] = true
		b, err := decodeBoolRaw(r)
		if err != nil {
			return Style{}, newLoadError(fieldPrefix+".italic", elementID, string(r), "must be a bool: "+err.Error())
		}
		st.Italic = present(b)
	}
	if r, ok := obj["fontFamily"]; ok {
		consumed["fontFamily"] = true
		s, err := decodeStringRaw(r)
		if err != nil {
			return Style{}, newLoadError(fieldPrefix+".fontFamily", elementID, string(r), "must be a string: "+err.Error())
		}
		st.FontFamily = present(s)
	}
	if r, ok := obj["fontSize"]; ok {
		consumed["fontSize"] = true
		v, err := decodePointsRaw(fieldPrefix+".fontSize", elementID, r)
		if err != nil {
			return Style{}, err
		}
		st.FontSize = present(v)
	}
	if r, ok := obj["lineSpacing"]; ok {
		// Story 7.2 / D-7.2.3, D-7.2.5. The entry in `consumed` is not
		// bookkeeping: without it the key falls through to extraFields
		// and round-trips opaquely through Extra, silently ignored by
		// every construction site — a documented format key the engine
		// would appear to accept and then never honour.
		consumed["lineSpacing"] = true
		v, err := DecodeLineSpacingRaw(r)
		if err != nil {
			// CODED, not plain. An uncoded LoadError becomes
			// TEMPLATE_MALFORMED at folio.ParseTemplate's boundary, and
			// wasm/cmd/engine's reportableMessage replaces THAT code's
			// message with "The template could not be processed" — so an
			// uncoded lineSpacing refusal is destroyed before the author
			// ever sees which element and which range it was about.
			// Located at fieldPrefix, so a headerStyle value is located
			// at headerStyle.lineSpacing rather than at its sibling.
			return Style{}, newLoadErrorCoded(fieldPrefix+".lineSpacing", elementID, string(r), err.Error(), diag.CodeStyleLineSpacingInvalid)
		}
		st.LineSpacing = present(v)
	}
	if r, ok := obj["padding"]; ok {
		consumed["padding"] = true
		pd, err := decodePadding(elementID, r, fieldPrefix+".padding")
		if err != nil {
			return Style{}, err
		}
		st.Padding = present(pd)
	}
	if r, ok := obj["border"]; ok {
		consumed["border"] = true
		b, err := decodeBorder(elementID, r, fieldPrefix+".border")
		if err != nil {
			return Style{}, err
		}
		st.Border = present(b)
	}

	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Style{}, fmt.Errorf("template: element %s %s: %w", elementID, fieldPrefix, err)
	}
	st.Extra = extra

	return st, nil
}

func decodePadding(elementID string, raw json.RawMessage, fieldPrefix string) (Padding, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Padding{}, newLoadError(fieldPrefix, elementID, string(raw), "must be an object: "+err.Error())
	}
	var p Padding
	consumed := map[string]bool{}
	for _, kv := range []struct {
		key string
		dst *Presence[geom.Length]
	}{
		{"top", &p.Top}, {"right", &p.Right}, {"bottom", &p.Bottom}, {"left", &p.Left},
	} {
		if r, ok := obj[kv.key]; ok {
			consumed[kv.key] = true
			v, err := decodePointsRaw(fieldPrefix+"."+kv.key, elementID, r)
			if err != nil {
				return Padding{}, err
			}
			*kv.dst = present(v)
		} // omitted edges default to 0 for layout purposes, but the key
		// itself stays absent (P3) — see Padding's doc comment.
	}
	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Padding{}, fmt.Errorf("template: %s: %w", fieldPrefix, err)
	}
	p.Extra = extra
	return p, nil
}

func decodeBorder(elementID string, raw json.RawMessage, fieldPrefix string) (Border, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Border{}, newLoadError(fieldPrefix, elementID, string(raw), "must be an object: "+err.Error())
	}
	consumed := map[string]bool{}
	var b Border
	if r, ok := obj["color"]; ok {
		consumed["color"] = true
		s, err := decodeStringRaw(r)
		if err != nil {
			return Border{}, newLoadError(fieldPrefix+".color", elementID, string(r), "must be a string: "+err.Error())
		}
		b.Color = present(s)
	}
	if r, ok := obj["width"]; ok {
		consumed["width"] = true
		v, err := decodePointsRaw(fieldPrefix+".width", elementID, r)
		if err != nil {
			return Border{}, err
		}
		// ISO 32000-1 §8.4.3.2: a line width is a NON-NEGATIVE number.
		// A negative one reached the emitter verbatim and produced
		// `-5 w` in a content stream, which is not a valid PDF — and
		// the product of this module is a byte-identical PDF, of which
		// a byte-identical non-PDF is not an instance.
		//
		// REFUSED AT LOAD, not at render, and that is a layering
		// decision rather than a preference. The value flows into
		// buildCellRectWithBackgroundField, which is SHARED with table
		// cell chrome, so `-5 w` has been reachable since Epic 4 and
		// Epic 9 only widened its reach to four more element kinds; one
		// load-time refusal closes both paths at one authority instead
		// of two render-time checks that can drift. It is legal here
		// only because `border.width` is a geom.Length and this package
		// already imports internal/geom — the same check on a COLOUR is
		// structurally unavailable at load (parseHexColor lives in the
		// module root, which AD-1 forbids this package to import; see
		// linespacing.go's note).
		//
		// ZERO IS VALID and stays accepted: it is the thinnest device
		// line PDF can draw, not an absent border. Only a NEGATIVE
		// width is refused.
		//
		// The code is the general load code newLoadError supplies
		// (TEMPLATE_FIELD_INVALID, D-7.8.1) — no consumer branches on
		// this condition, so it discriminates on Field like every other
		// located field refusal.
		if v < 0 {
			return Border{}, newLoadError(fieldPrefix+".width", elementID, string(r), "must not be negative: a PDF line width is non-negative (ISO 32000-1 8.4.3.2); use 0 for the thinnest line")
		}
		b.Width = present(v)
	}
	if r, ok := obj["edges"]; ok {
		consumed["edges"] = true
		edges, err := decodeStringArrayRaw(r)
		if err != nil {
			return Border{}, newLoadError(fieldPrefix+".edges", elementID, string(r), "must be an array of strings: "+err.Error())
		}
		for _, e := range edges {
			if !closedBorderEdges[e] {
				return Border{}, newLoadError(fieldPrefix+".edges", elementID, e, "not one of the closed set top, right, bottom, left")
			}
		}
		b.Edges = present(edges)
	}
	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Border{}, fmt.Errorf("template: %s: %w", fieldPrefix, err)
	}
	b.Extra = extra
	return b, nil
}
