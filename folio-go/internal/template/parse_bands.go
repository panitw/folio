package template

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

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

	content, err := decodeBand(ctx, "bands.content", obj["content"], false)
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

	consumed := map[string]bool{"id": true, "type": true, "x": true, "y": true}

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
			st, err := decodeStyle(string(id), styRaw)
			if err != nil {
				return Element{}, err
			}
			el.Style = present(st)
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
		s, err := decodeStringRaw(aRaw)
		if err != nil {
			return Element{}, newLoadError("asset", string(id), string(aRaw), "must be a string: "+err.Error())
		}
		el.Asset = present(s)
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
		if !closedAligns[s] {
			return Column{}, newLoadError("align", string(id), s, "not one of the closed set left, center, right")
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
		return Column{}, newLoadError("footerOf/footerFormat", string(id), "", "footerOf/footerFormat present with no footer — load error (D-1.4.2)")
	}

	if footerOfRaw, ok := obj["footerOf"]; ok {
		consumed["footerOf"] = true
		s, err := decodeStringRaw(footerOfRaw)
		if err != nil {
			return Column{}, newLoadError("footerOf", string(id), string(footerOfRaw), "must be a string: "+err.Error())
		}
		// AC43 check 1: footerOf present with footer: "count" — load error.
		if footer.Set && footer.Value == "count" {
			return Column{}, newLoadError("footerOf", string(id), s, `footerOf present alongside footer: "count" is a load error (D-1.4.2: storing it would be a second source of truth against bind, AD-13)`)
		}
		// AC43 check 3: footerOf must be prefixed by the table's own
		// collection path + "." — a string prefix test, no parser.
		prefix := collection + "."
		if !strings.HasPrefix(s, prefix) {
			return Column{}, newLoadError("footerOf", string(id), s, fmt.Sprintf("must be prefixed by the table's collection path %q (D-1.4.2)", prefix))
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

func decodeStyle(elementID string, raw json.RawMessage) (Style, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Style{}, newLoadError("style", elementID, string(raw), "must be an object: "+err.Error())
	}
	consumed := map[string]bool{}
	var st Style

	if r, ok := obj["align"]; ok {
		consumed["align"] = true
		s, err := decodeStringRaw(r)
		if err != nil {
			return Style{}, newLoadError("style.align", elementID, string(r), "must be a string: "+err.Error())
		}
		if !closedAligns[s] {
			return Style{}, newLoadError("style.align", elementID, s, "not one of the closed set left, center, right")
		}
		st.Align = present(s)
	}
	if r, ok := obj["valign"]; ok {
		consumed["valign"] = true
		s, err := decodeStringRaw(r)
		if err != nil {
			return Style{}, newLoadError("style.valign", elementID, string(r), "must be a string: "+err.Error())
		}
		if !closedValigns[s] {
			return Style{}, newLoadError("style.valign", elementID, s, "not one of the closed set top, middle, bottom")
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
				return Style{}, newLoadError("style.background", elementID, string(r), "must be a string: "+err.Error())
			}
			st.Background = present(s)
		}
	}
	if r, ok := obj["bold"]; ok {
		consumed["bold"] = true
		b, err := decodeBoolRaw(r)
		if err != nil {
			return Style{}, newLoadError("style.bold", elementID, string(r), "must be a bool: "+err.Error())
		}
		st.Bold = present(b)
	}
	if r, ok := obj["italic"]; ok {
		consumed["italic"] = true
		b, err := decodeBoolRaw(r)
		if err != nil {
			return Style{}, newLoadError("style.italic", elementID, string(r), "must be a bool: "+err.Error())
		}
		st.Italic = present(b)
	}
	if r, ok := obj["fontFamily"]; ok {
		consumed["fontFamily"] = true
		s, err := decodeStringRaw(r)
		if err != nil {
			return Style{}, newLoadError("style.fontFamily", elementID, string(r), "must be a string: "+err.Error())
		}
		st.FontFamily = present(s)
	}
	if r, ok := obj["fontSize"]; ok {
		consumed["fontSize"] = true
		v, err := decodePointsRaw("style.fontSize", elementID, r)
		if err != nil {
			return Style{}, err
		}
		st.FontSize = present(v)
	}
	if r, ok := obj["padding"]; ok {
		consumed["padding"] = true
		pd, err := decodePadding(elementID, r)
		if err != nil {
			return Style{}, err
		}
		st.Padding = present(pd)
	}
	if r, ok := obj["border"]; ok {
		consumed["border"] = true
		b, err := decodeBorder(elementID, r)
		if err != nil {
			return Style{}, err
		}
		st.Border = present(b)
	}

	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Style{}, fmt.Errorf("template: element %s style: %w", elementID, err)
	}
	st.Extra = extra

	return st, nil
}

func decodePadding(elementID string, raw json.RawMessage) (Padding, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Padding{}, newLoadError("style.padding", elementID, string(raw), "must be an object: "+err.Error())
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
			v, err := decodePointsRaw("style.padding."+kv.key, elementID, r)
			if err != nil {
				return Padding{}, err
			}
			*kv.dst = present(v)
		} // omitted edges default to 0 for layout purposes, but the key
		// itself stays absent (P3) — see Padding's doc comment.
	}
	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Padding{}, fmt.Errorf("template: style.padding: %w", err)
	}
	p.Extra = extra
	return p, nil
}

func decodeBorder(elementID string, raw json.RawMessage) (Border, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Border{}, newLoadError("style.border", elementID, string(raw), "must be an object: "+err.Error())
	}
	consumed := map[string]bool{}
	var b Border
	if r, ok := obj["color"]; ok {
		consumed["color"] = true
		s, err := decodeStringRaw(r)
		if err != nil {
			return Border{}, newLoadError("style.border.color", elementID, string(r), "must be a string: "+err.Error())
		}
		b.Color = present(s)
	}
	if r, ok := obj["width"]; ok {
		consumed["width"] = true
		v, err := decodePointsRaw("style.border.width", elementID, r)
		if err != nil {
			return Border{}, err
		}
		b.Width = present(v)
	}
	if r, ok := obj["edges"]; ok {
		consumed["edges"] = true
		edges, err := decodeStringArrayRaw(r)
		if err != nil {
			return Border{}, newLoadError("style.border.edges", elementID, string(r), "must be an array of strings: "+err.Error())
		}
		for _, e := range edges {
			if !closedBorderEdges[e] {
				return Border{}, newLoadError("style.border.edges", elementID, e, "not one of the closed set top, right, bottom, left")
			}
		}
		b.Edges = present(edges)
	}
	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Border{}, fmt.Errorf("template: style.border: %w", err)
	}
	b.Extra = extra
	return b, nil
}
