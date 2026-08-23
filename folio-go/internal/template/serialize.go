package template

import (
	"maps"
	"slices"
	"sort"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// This file is the serializer (AD-9). Every object at every level is
// built as a slice of (key, value-writer) pairs — known fields AND
// passthrough Extra fields together — then sorted once by byte-order
// key and emitted as one sequence (AC8: "merged back in sorted order …
// one sorted key sequence, not a known block followed by an unknown
// block", AC18). No `omitempty`-shaped skipping happens anywhere except
// through Presence.Set (D-1.4.3 extended): every map and slice that
// reaches writeObject/writeArray is already non-nil (constructed empty
// by the parser), so an authored `{}` or `[]` always round-trips as
// itself rather than vanishing (AC22, AC23).

// SerializeDocument renders d to its canonical byte form (AD-9): keys
// sorted, two-space indent, LF, no trailing whitespace, a trailing
// newline, HTML escaping off, literal UTF-8, minimal escaping.
func SerializeDocument(d *Document) ([]byte, error) {
	var buf []byte
	buf = writeDocument(buf, d)
	buf = append(buf, '\n') // AD-9's trailing newline (M-2: Encoder.Encode's behaviour, reproduced by hand here).
	return buf, nil
}

// kv is one (key, value-writer) pair at some object nesting level.
type kv struct {
	key   string
	write func(dst []byte, depth int) []byte
}

func writeObject(dst []byte, depth int, fields []kv) []byte {
	if len(fields) == 0 {
		return append(dst, "{}"...)
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].key < fields[j].key })
	dst = append(dst, '{')
	for i, f := range fields {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = append(dst, '\n')
		dst = appendIndent(dst, depth+1)
		dst = appendJSONString(dst, f.key)
		dst = append(dst, ':', ' ')
		dst = f.write(dst, depth+1)
	}
	dst = append(dst, '\n')
	dst = appendIndent(dst, depth)
	dst = append(dst, '}')
	return dst
}

func writeStringArray(dst []byte, depth int, items []string) []byte {
	if len(items) == 0 {
		return append(dst, "[]"...)
	}
	dst = append(dst, '[')
	for i, s := range items {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = append(dst, '\n')
		dst = appendIndent(dst, depth+1)
		dst = appendJSONString(dst, s)
	}
	dst = append(dst, '\n')
	dst = appendIndent(dst, depth)
	dst = append(dst, ']')
	return dst
}

func writeBool(b bool) func(dst []byte, depth int) []byte {
	return func(dst []byte, depth int) []byte {
		if b {
			return append(dst, "true"...)
		}
		return append(dst, "false"...)
	}
}

func writeString(s string) func(dst []byte, depth int) []byte {
	return func(dst []byte, depth int) []byte { return appendJSONString(dst, s) }
}

func writePoints(v geom.Length) func(dst []byte, depth int) []byte {
	return func(dst []byte, depth int) []byte { return appendPoints(dst, v) }
}

func writePlainInt(v int64) func(dst []byte, depth int) []byte {
	return func(dst []byte, depth int) []byte { return appendPlainInt(dst, v) }
}

func writeNull() func(dst []byte, depth int) []byte {
	return func(dst []byte, depth int) []byte { return append(dst, "null"...) }
}

// extraKVs converts the passthrough Field list into kv entries.
func extraKVs(fields []Field) []kv {
	out := make([]kv, 0, len(fields))
	for _, f := range fields {
		f := f
		out = append(out, kv{key: f.Key, write: func(dst []byte, depth int) []byte {
			return appendRaw(dst, f.Value, depth)
		}})
	}
	return out
}

func writeDocument(dst []byte, d *Document) []byte {
	fields := []kv{
		{"version", writeString(versionForSave(d.Version))},
		{"locale", writeString(d.Locale)},
		{"utcOffset", writeString(d.UTCOffset)},
		{"page", func(dst []byte, depth int) []byte { return writePage(dst, depth, d.Page) }},
		{"fonts", func(dst []byte, depth int) []byte { return writeFonts(dst, depth, d.Fonts) }},
		{"bands", func(dst []byte, depth int) []byte { return writeBands(dst, depth, d.Bands) }},
		{"assets", func(dst []byte, depth int) []byte { return writeAssets(dst, depth, d.Assets) }},
		{"nextId", writePlainInt(d.NextID)},
	}
	fields = append(fields, extraKVs(d.Extra)...)
	return writeObject(dst, 0, fields)
}

func writePage(dst []byte, depth int, p Page) []byte {
	fields := []kv{
		{"margin", func(dst []byte, depth int) []byte { return writeMargin(dst, depth, p.Margin) }},
		{"orientation", writeString(p.Orientation)},
	}
	if p.SizeIsName {
		fields = append(fields, kv{"size", writeString(p.SizeName)})
	} else {
		fields = append(fields, kv{"size", func(dst []byte, depth int) []byte {
			return writeObject(dst, depth, []kv{
				{"height", writePoints(p.SizeCustom.Height)},
				{"width", writePoints(p.SizeCustom.Width)},
			})
		}})
	}
	fields = append(fields, extraKVs(p.Extra)...)
	return writeObject(dst, depth, fields)
}

func writeMargin(dst []byte, depth int, m Margin) []byte {
	fields := []kv{
		{"top", writePoints(m.Top)},
		{"right", writePoints(m.Right)},
		{"bottom", writePoints(m.Bottom)},
		{"left", writePoints(m.Left)},
	}
	fields = append(fields, extraKVs(m.Extra)...)
	return writeObject(dst, depth, fields)
}

func writeFonts(dst []byte, depth int, f Fonts) []byte {
	fields := make([]kv, 0, len(f))
	// D-1.3.5/NFR1.d: ranging a map is forbidden under internal/;
	// writeObject sorts fields by key anyway (AC18), but the sorted-key
	// idiom is used here too rather than relying on that downstream step.
	for _, k := range slices.Sorted(maps.Keys(f)) {
		chain := f[k]
		fields = append(fields, kv{k, func(dst []byte, depth int) []byte { return writeStringArray(dst, depth, chain) }})
	}
	return writeObject(dst, depth, fields)
}

func writeBands(dst []byte, depth int, b Bands) []byte {
	return writeObject(dst, depth, []kv{
		{"content", func(dst []byte, depth int) []byte { return writeBand(dst, depth, b.Content) }},
		{"pageFooter", func(dst []byte, depth int) []byte { return writeBand(dst, depth, b.PageFooter) }},
		{"pageHeader", func(dst []byte, depth int) []byte { return writeBand(dst, depth, b.PageHeader) }},
	})
}

func writeBand(dst []byte, depth int, band Band) []byte {
	fields := []kv{
		{"elements", func(dst []byte, depth int) []byte { return writeElements(dst, depth, band.Elements) }},
	}
	if band.Height.Set {
		fields = append(fields, kv{"height", writePoints(band.Height.Value)})
	}
	fields = append(fields, extraKVs(band.Extra)...)
	return writeObject(dst, depth, fields)
}

func writeElements(dst []byte, depth int, elems []Element) []byte {
	if len(elems) == 0 {
		return append(dst, "[]"...)
	}
	dst = append(dst, '[')
	for i, e := range elems {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = append(dst, '\n')
		dst = appendIndent(dst, depth+1)
		dst = writeElement(dst, depth+1, e)
	}
	dst = append(dst, '\n')
	dst = appendIndent(dst, depth)
	dst = append(dst, ']')
	return dst
}

func writeElement(dst []byte, depth int, e Element) []byte {
	fields := []kv{
		{"id", writeString(string(e.ID))},
		{"type", writeString(string(e.Type))},
		{"x", writePoints(e.X)},
		{"y", writePoints(e.Y)},
	}
	if e.Width.Set {
		fields = append(fields, kv{"width", writePoints(e.Width.Value)})
	}
	if e.Height.Set {
		fields = append(fields, kv{"height", writePoints(e.Height.Value)})
	}
	if e.VisibleIf.Set {
		if e.VisibleIf.Null {
			fields = append(fields, kv{"visibleIf", writeNull()})
		} else {
			fields = append(fields, kv{"visibleIf", writeString(e.VisibleIf.Value)})
		}
	}
	if e.Style.Set {
		if e.Style.Null {
			fields = append(fields, kv{"style", writeNull()})
		} else {
			st := e.Style.Value
			fields = append(fields, kv{"style", func(dst []byte, depth int) []byte { return writeStyle(dst, depth, st) }})
		}
	}
	if e.Value.Set {
		if e.Value.Null {
			fields = append(fields, kv{"value", writeNull()})
		} else {
			fields = append(fields, kv{"value", writeString(e.Value.Value)})
		}
	}
	if e.Asset.Set {
		fields = append(fields, kv{"asset", writeString(e.Asset.Value)})
	}
	if e.Table.Set {
		t := e.Table.Value
		fields = append(fields, kv{"bind", writeString(t.Bind)})
		if t.As.Set {
			fields = append(fields, kv{"as", writeString(t.As.Value)})
		}
		cols := t.Columns
		fields = append(fields, kv{"columns", func(dst []byte, depth int) []byte { return writeColumns(dst, depth, cols) }})
		fields = append(fields, kv{"headerHeight", writePoints(t.HeaderHeight)})
		if t.AltRowBackground.Set {
			fields = append(fields, kv{"altRowBackground", writeString(t.AltRowBackground.Value)})
		}
	}
	fields = append(fields, extraKVs(e.Extra)...)
	return writeObject(dst, depth, fields)
}

func writeColumns(dst []byte, depth int, cols []Column) []byte {
	if len(cols) == 0 {
		return append(dst, "[]"...)
	}
	dst = append(dst, '[')
	for i, c := range cols {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = append(dst, '\n')
		dst = appendIndent(dst, depth+1)
		dst = writeColumn(dst, depth+1, c)
	}
	dst = append(dst, '\n')
	dst = appendIndent(dst, depth)
	dst = append(dst, ']')
	return dst
}

func writeColumn(dst []byte, depth int, c Column) []byte {
	fields := []kv{
		{"id", writeString(string(c.ID))},
		{"label", writeString(c.Label)},
		{"width", writePoints(c.Width)},
		{"bind", writeString(c.Bind)},
	}
	if c.Align.Set {
		fields = append(fields, kv{"align", writeString(c.Align.Value)})
	}
	if c.Footer.Set {
		fields = append(fields, kv{"footer", writeString(c.Footer.Value)})
	}
	if c.FooterOf.Set {
		fields = append(fields, kv{"footerOf", writeString(c.FooterOf.Value)})
	}
	if c.FooterFormat.Set {
		fields = append(fields, kv{"footerFormat", writeString(c.FooterFormat.Value)})
	}
	fields = append(fields, extraKVs(c.Extra)...)
	return writeObject(dst, depth, fields)
}

func writeStyle(dst []byte, depth int, st Style) []byte {
	var fields []kv
	if st.Align.Set {
		fields = append(fields, kv{"align", writeString(st.Align.Value)})
	}
	if st.Valign.Set {
		fields = append(fields, kv{"valign", writeString(st.Valign.Value)})
	}
	if st.Background.Set {
		if st.Background.Null {
			fields = append(fields, kv{"background", writeNull()})
		} else {
			fields = append(fields, kv{"background", writeString(st.Background.Value)})
		}
	}
	if st.Bold.Set {
		fields = append(fields, kv{"bold", writeBool(st.Bold.Value)})
	}
	if st.Italic.Set {
		fields = append(fields, kv{"italic", writeBool(st.Italic.Value)})
	}
	if st.FontFamily.Set {
		fields = append(fields, kv{"fontFamily", writeString(st.FontFamily.Value)})
	}
	if st.FontSize.Set {
		fields = append(fields, kv{"fontSize", writePoints(st.FontSize.Value)})
	}
	if st.Padding.Set {
		pd := st.Padding.Value
		fields = append(fields, kv{"padding", func(dst []byte, depth int) []byte { return writePadding(dst, depth, pd) }})
	}
	if st.Border.Set {
		bd := st.Border.Value
		fields = append(fields, kv{"border", func(dst []byte, depth int) []byte { return writeBorder(dst, depth, bd) }})
	}
	fields = append(fields, extraKVs(st.Extra)...)
	return writeObject(dst, depth, fields)
}

func writePadding(dst []byte, depth int, p Padding) []byte {
	var fields []kv
	if p.Top.Set {
		fields = append(fields, kv{"top", writePoints(p.Top.Value)})
	}
	if p.Right.Set {
		fields = append(fields, kv{"right", writePoints(p.Right.Value)})
	}
	if p.Bottom.Set {
		fields = append(fields, kv{"bottom", writePoints(p.Bottom.Value)})
	}
	if p.Left.Set {
		fields = append(fields, kv{"left", writePoints(p.Left.Value)})
	}
	fields = append(fields, extraKVs(p.Extra)...)
	return writeObject(dst, depth, fields)
}

func writeBorder(dst []byte, depth int, b Border) []byte {
	var fields []kv
	if b.Color.Set {
		fields = append(fields, kv{"color", writeString(b.Color.Value)})
	}
	if b.Edges.Set {
		edges := b.Edges.Value
		fields = append(fields, kv{"edges", func(dst []byte, depth int) []byte { return writeStringArray(dst, depth, edges) }})
	}
	if b.Width.Set {
		fields = append(fields, kv{"width", writePoints(b.Width.Value)})
	}
	fields = append(fields, extraKVs(b.Extra)...)
	return writeObject(dst, depth, fields)
}

// writeAssets re-wraps every asset's base64 from its DECODED bytes at
// the canonical 76-column split and NEVER echoes the input array
// (D-1.8.9, AC3a): it decodes a.Data (whatever wrapping it happens to
// carry — any wrapping is accepted on parse, D-1.8.2) and calls
// splitBase64Canonical on the result, so a serializer that merely
// "fixed up" the existing elements would not satisfy this — the
// canonical output is derived from the bytes, not from the input
// array's shape. AC8: assets with no referencing element (orphans) are
// PRESERVED here unconditionally — D-1.4.3's P1 (Parse(Serialize(d)) ==
// d) forces this; there is no policy latitude to drop one.
//
// Epics 5/6 corollary (AC8a), verbatim: garbage-collecting orphans is a
// designer feature — an explicit user action — never a serializer side
// effect. A future "tidy unused assets" pass belongs elsewhere, never
// here.
//
// Invariant this function assumes: a.Data is always valid, previously
// parse-validated base64 (ParseDocument's decodeAssets already decoded
// and validated it, per D-1.8.2/D-1.8.3). A hand-built Document whose
// Asset.Data is not valid base64 violates that invariant; this function
// panics rather than silently emitting invalid output, the same
// programmer-error-guard shape ScaleRound uses (D-1.5.2's precedent).
func writeAssets(dst []byte, depth int, assets map[string]Asset) []byte {
	fields := make([]kv, 0, len(assets))
	for _, k := range slices.Sorted(maps.Keys(assets)) {
		a := assets[k]
		decoded, err := decodeBase64Asset(a.Data)
		if err != nil {
			panic("template: SerializeDocument: asset " + k + " carries invalid base64 (invariant violation — assets must be parse-validated before serialization): " + err.Error())
		}
		canonical := splitBase64Canonical(decoded)
		fields = append(fields, kv{k, func(dst []byte, depth int) []byte {
			assetFields := []kv{
				{"data", func(dst []byte, depth int) []byte { return writeStringArray(dst, depth, canonical) }},
				{"mediaType", writeString(a.MediaType)},
			}
			assetFields = append(assetFields, extraKVs(a.Extra)...)
			return writeObject(dst, depth, assetFields)
		}})
	}
	return writeObject(dst, depth, fields)
}
