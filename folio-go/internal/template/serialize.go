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
		{"version", writeString(versionForSave(d.Version, d))},
		{"locale", writeString(d.Locale)},
		{"utcOffset", writeString(d.UTCOffset)},
		{"page", func(dst []byte, depth int) []byte { return writePage(dst, depth, d.Page) }},
		{"fonts", func(dst []byte, depth int) []byte { return writeFonts(dst, depth, d.Fonts) }},
		{"bands", func(dst []byte, depth int) []byte { return writeBands(dst, depth, d.Bands) }},
		{"assets", func(dst []byte, depth int) []byte { return writeAssets(dst, depth, d.Assets) }},
		{"nextId", writePlainInt(d.NextID)},
	}
	// unbreakableValues (Story 2.4, D-2.4.1) is OPTIONAL, and an absent
	// declaration must round-trip as an ABSENT KEY — not as "[]".
	// Emitting an empty array for every document that never used the
	// feature would move every existing golden's bytes, which AC12
	// forbids: exactly one recorded artifact moves in this story, and it
	// is not a template. The nil/empty distinction is therefore
	// load-bearing here in the same way D-1.4.3's fifth trap is.
	if len(d.UnbreakableValues) > 0 {
		paths := d.UnbreakableValues
		fields = append(fields, kv{"unbreakableValues", func(dst []byte, depth int) []byte {
			return writeStringArray(dst, depth, paths)
		}})
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
		fields = append(fields, kv{k, func(dst []byte, depth int) []byte { return writeFontChain(dst, depth, chain) }})
	}
	return writeObject(dst, depth, fields)
}

// writeFontChain is writeStringArray with a per-ENTRY writer (Story
// 8.3). It exists because writeStringArray can emit only JSON strings,
// and a chain entry may now be an object.
//
// An embedded entry is routed through writeObject — the ONE generic
// object emitter — rather than being spelled out as `{"asset": …}` here:
// a hand-written second emitter would be a second answer to indentation,
// escaping and key order, and the whole canonical-fixed-point property
// (P3) rests on there being one. A one-key object has no key order to
// get wrong today, which is exactly why it is the cheapest possible
// place to introduce a second emitter by accident.
//
// A string entry is emitted EXACTLY as writeStringArray emitted it, so a
// document with no embedded entry is byte-identical to what it was
// before this function existed — the recorded golden digests are the
// population that proves it (22 of them when Story 8.3 measured this;
// 23 since Story 8.4, which added the one fixture that HAS an embedded
// entry, and moved no other).
func writeFontChain(dst []byte, depth int, chain []FontChainEntry) []byte {
	if len(chain) == 0 {
		return append(dst, "[]"...)
	}
	dst = append(dst, '[')
	for i, entry := range chain {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = append(dst, '\n')
		dst = appendIndent(dst, depth+1)
		if entry.Embedded() {
			dst = writeObject(dst, depth+1, []kv{{"asset", writeString(entry.AssetKey)}})
			continue
		}
		dst = appendJSONString(dst, entry.Face)
	}
	dst = append(dst, '\n')
	dst = appendIndent(dst, depth)
	dst = append(dst, ']')
	return dst
}

// writeFontRecord emits assets[k].font (Story 8.3). Every key is
// optional in the three-way sense presence.go defines, and each is
// emitted only when it was SET — so an absent key stays absent and an
// explicit null round-trips as null, rather than the two collapsing into
// one spelling on the way out.
func writeFontRecord(dst []byte, depth int, r FontRecord) []byte {
	fields := make([]kv, 0, 6+len(r.Extra))
	for _, f := range []struct {
		key string
		val Presence[string]
	}{
		// Listed alphabetically for a reader; writeObject sorts anyway,
		// so `licenceText` lands after `licence` and `copyright` ahead of
		// `family` whatever order this literal is written in.
		{"copyright", r.Copyright},
		{"family", r.Family},
		{"licence", r.Licence},
		{"licenceText", r.LicenceText},
		{"source", r.Source},
		{"style", r.Style},
	} {
		if !f.val.Set {
			continue
		}
		if f.val.Null {
			fields = append(fields, kv{f.key, writeNull()})
			continue
		}
		fields = append(fields, kv{f.key, writeString(f.val.Value)})
	}
	fields = append(fields, extraKVs(r.Extra)...)
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
	// Story 7.7 (FR51). The unkeyed literal form is REQUIRED here, not a
	// style choice: drift_test.go's AST reader (extractGoKeys) sees a
	// kv literal only in the `kv{"key", …}` spelling, so a keyed
	// `kv{key: …, write: …}` would emit a key folio-format.md's drift
	// guard is structurally blind to.
	if e.KeepTogether.Set {
		if e.KeepTogether.Null {
			fields = append(fields, kv{"keepTogether", writeNull()})
		} else {
			fields = append(fields, kv{"keepTogether", writeString(e.KeepTogether.Value)})
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
		if e.Asset.Null {
			fields = append(fields, kv{"asset", writeNull()})
		} else {
			fields = append(fields, kv{"asset", writeString(e.Asset.Value)})
		}
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
		if t.HeaderStyle.Set {
			if t.HeaderStyle.Null {
				fields = append(fields, kv{"headerStyle", writeNull()})
			} else {
				hs := t.HeaderStyle.Value
				fields = append(fields, kv{"headerStyle", func(dst []byte, depth int) []byte { return writeStyle(dst, depth, hs) }})
			}
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
	if st.Color.Set {
		if st.Color.Null {
			fields = append(fields, kv{"color", writeNull()})
		} else {
			fields = append(fields, kv{"color", writeString(st.Color.Value)})
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
	if st.LineSpacing.Set {
		// Emitted in the SAME exact-decimal spelling it was authored in:
		// thousandths through appendPoints is the identical ×1000
		// decimal path decodePoints read it with, so 1500 round-trips as
		// `1.5` and 1000 as `1`. AD-9's edit-and-edit-back byte
		// identity, not a second number formatter.
		fields = append(fields, kv{"lineSpacing", writePoints(geom.Length(st.LineSpacing.Value))})
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
			// Story 8.3's one addition to an asset. Emitted only when
			// SET, so an image asset's bytes are untouched; writeObject
			// sorts, so "font" lands between "data" and "mediaType" and
			// moves neither's content.
			if a.Font.Set {
				font := a.Font
				assetFields = append(assetFields, kv{"font", func(dst []byte, depth int) []byte {
					if font.Null {
						return append(dst, "null"...)
					}
					return writeFontRecord(dst, depth, font.Value)
				}})
			}
			assetFields = append(assetFields, extraKVs(a.Extra)...)
			return writeObject(dst, depth, assetFields)
		}})
	}
	return writeObject(dst, depth, fields)
}
