package template

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// This file is the parser (AD-9). json.Decoder + UseNumber decodes every
// leaf number as a preserved literal (AC26); the two numeric kinds
// (points, nextId) are converted through decimal.go's exact path
// (AC24). Closed-set validation (closedsets.go) is kept visibly
// separate from unknown-key passthrough (rawvalue.go, decodehelpers.go)
// — AC10, AC11 clause 3. Every load error is a *LoadError naming
// field, element id and value (AC41). Nothing is ever coerced (AC40).

// parseCtx threads document-wide id bookkeeping through the recursive
// descent (AC35: a duplicate id anywhere is a load error, across bands,
// elements and columns alike — the counter is document-wide, AC36).
type parseCtx struct {
	ids        map[string]int64 // id string -> decoded counter
	maxCounter int64
}

func newParseCtx() *parseCtx {
	return &parseCtx{ids: map[string]int64{}}
}

// claimID validates id's spelling (AC34), rejects a duplicate anywhere
// in the document (AC35), and records it toward AC37's nextId check.
func (c *parseCtx) claimID(field, id string) (ElementID, error) {
	counter, err := validateElementID(id)
	if err != nil {
		return "", newLoadError(field, id, id, err.Error())
	}
	if _, dup := c.ids[id]; dup {
		return "", newLoadError(field, id, id, "duplicate id: ids are unique document-wide (AD-10, AC35, AC36)")
	}
	c.ids[id] = counter
	if counter > c.maxCounter {
		c.maxCounter = counter
	}
	return ElementID(id), nil
}

// ParseDocument parses b into a canonicalised Document (AC4). It is the
// sole entry point internal/template exposes for parsing; package
// folio's ParseTemplate/LoadTemplate (folio-go/folio.go) call this.
func ParseDocument(b []byte) (*Document, error) {
	top, err := decodeObjectMap(json.RawMessage(b))
	if err != nil {
		return nil, fmt.Errorf("template: root: %w", err)
	}

	ctx := newParseCtx()
	doc := &Document{}
	consumed := map[string]bool{}

	// version (AC6, AC7)
	verRaw, ok := top["version"]
	if !ok {
		return nil, newLoadError("version", "", "", "missing required field")
	}
	consumed["version"] = true
	version, err := decodeStringRaw(verRaw)
	if err != nil {
		return nil, newLoadError("version", "", string(verRaw), "must be a string: "+err.Error())
	}
	if err := checkVersionLoadable(version); err != nil {
		return nil, err
	}
	doc.Version = version

	// locale (AD-12, AC5)
	if raw, ok := top["locale"]; ok {
		consumed["locale"] = true
		s, err := decodeStringRaw(raw)
		if err != nil {
			return nil, newLoadError("locale", "", string(raw), "must be a string: "+err.Error())
		}
		if !closedLocales[s] {
			return nil, newLoadError("locale", "", s, "not one of the closed set en, th, zh-Hans, ja (AD-12)")
		}
		doc.Locale = s
	} else {
		return nil, newLoadError("locale", "", "", "missing required field")
	}

	// utcOffset
	if raw, ok := top["utcOffset"]; ok {
		consumed["utcOffset"] = true
		s, err := decodeStringRaw(raw)
		if err != nil {
			return nil, newLoadError("utcOffset", "", string(raw), "must be a string: "+err.Error())
		}
		if !utcOffsetPattern.MatchString(s) {
			return nil, newLoadError("utcOffset", "", s, "must match ±HH:MM")
		}
		doc.UTCOffset = s
	} else {
		return nil, newLoadError("utcOffset", "", "", "missing required field")
	}

	// page
	if raw, ok := top["page"]; ok {
		consumed["page"] = true
		p, err := decodePage(raw)
		if err != nil {
			return nil, err
		}
		doc.Page = p
	} else {
		return nil, newLoadError("page", "", "", "missing required field")
	}

	// fonts
	if raw, ok := top["fonts"]; ok {
		consumed["fonts"] = true
		f, err := decodeFonts(raw)
		if err != nil {
			return nil, err
		}
		doc.Fonts = f
	} else {
		doc.Fonts = Fonts{}
		consumed["fonts"] = true
	}

	// bands
	if raw, ok := top["bands"]; ok {
		consumed["bands"] = true
		bands, err := decodeBands(ctx, raw)
		if err != nil {
			return nil, err
		}
		doc.Bands = bands
	} else {
		return nil, newLoadError("bands", "", "", "missing required field")
	}

	// assets
	if raw, ok := top["assets"]; ok {
		consumed["assets"] = true
		assets, err := decodeAssets(raw)
		if err != nil {
			return nil, err
		}
		doc.Assets = assets
	} else {
		doc.Assets = map[string]Asset{}
		consumed["assets"] = true
	}

	// nextId (AC32, AC33, AC37) — must be a plain decimal integer, never
	// base 36.
	nextRaw, ok := top["nextId"]
	if !ok {
		return nil, newLoadError("nextId", "", "", "missing required field (never repaired, never inferred, AD-10/AC37)")
	}
	consumed["nextId"] = true
	nextID, err := decodePlainInteger(nextRaw)
	if err != nil {
		return nil, newLoadError("nextId", "", string(nextRaw), "must be a plain decimal integer: "+err.Error())
	}
	if nextID <= ctx.maxCounter {
		return nil, newLoadError("nextId", "", fmt.Sprintf("%d", nextID),
			fmt.Sprintf("must be greater than the highest id present (%d) — never repaired, never renumbered (AD-10, AC37)", ctx.maxCounter))
	}
	doc.NextID = nextID

	extra, err := extraFields(top, consumed)
	if err != nil {
		return nil, fmt.Errorf("template: root: %w", err)
	}
	doc.Extra = extra

	return doc, nil
}

// decodePlainInteger decodes a JSON number literal that must be a plain
// decimal integer (no '.', no exponent) — AC32's nextId spelling.
func decodePlainInteger(raw json.RawMessage) (int64, error) {
	n, err := decodeNumberRaw(raw)
	if err != nil {
		return 0, err
	}
	return n.Int64()
}

func decodePage(raw json.RawMessage) (Page, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Page{}, fmt.Errorf("template: page: %w", err)
	}
	var p Page
	consumed := map[string]bool{"margin": true, "orientation": true, "size": true}

	marginRaw, ok := obj["margin"]
	if !ok {
		return Page{}, newLoadError("page.margin", "", "", "missing required field")
	}
	m, err := decodeMargin(marginRaw)
	if err != nil {
		return Page{}, err
	}
	p.Margin = m

	orRaw, ok := obj["orientation"]
	if !ok {
		return Page{}, newLoadError("page.orientation", "", "", "missing required field")
	}
	or, err := decodeStringRaw(orRaw)
	if err != nil {
		return Page{}, newLoadError("page.orientation", "", string(orRaw), "must be a string: "+err.Error())
	}
	if !closedPageOrientations[or] {
		return Page{}, newLoadError("page.orientation", "", or, "not one of the closed set portrait, landscape")
	}
	p.Orientation = or

	sizeRaw, ok := obj["size"]
	if !ok {
		return Page{}, newLoadError("page.size", "", "", "missing required field")
	}
	if s, err := decodeStringRaw(sizeRaw); err == nil {
		if !closedPageSizeNames[s] {
			return Page{}, newLoadError("page.size", "", s, `not one of the closed set "A4", "Letter", or an object with height/width`)
		}
		p.SizeIsName = true
		p.SizeName = s
	} else {
		sizeObj, oerr := decodeObjectMap(sizeRaw)
		if oerr != nil {
			return Page{}, newLoadError("page.size", "", string(sizeRaw), `must be "A4", "Letter", or an object with height/width`)
		}
		hRaw, hok := sizeObj["height"]
		wRaw, wok := sizeObj["width"]
		if !hok || !wok {
			return Page{}, newLoadError("page.size", "", string(sizeRaw), "custom size object must carry both height and width")
		}
		h, err := decodePointsRaw("page.size.height", "", hRaw)
		if err != nil {
			return Page{}, err
		}
		w, err := decodePointsRaw("page.size.width", "", wRaw)
		if err != nil {
			return Page{}, err
		}
		if len(sizeObj) != 2 {
			return Page{}, newLoadError("page.size", "", string(sizeRaw), "custom size object must carry exactly height and width")
		}
		p.SizeCustom = PageSize{Height: h, Width: w}
	}

	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Page{}, fmt.Errorf("template: page: %w", err)
	}
	p.Extra = extra

	return p, nil
}

func decodeMargin(raw json.RawMessage) (Margin, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return Margin{}, fmt.Errorf("template: page.margin: %w", err)
	}
	var m Margin
	consumed := map[string]bool{}
	for _, kv := range []struct {
		key string
		dst *geom.Length
	}{
		{"top", &m.Top}, {"right", &m.Right}, {"bottom", &m.Bottom}, {"left", &m.Left},
	} {
		r, ok := obj[kv.key]
		if !ok {
			return Margin{}, newLoadError("page.margin."+kv.key, "", "", "missing required field")
		}
		consumed[kv.key] = true
		v, err := decodePointsRaw("page.margin."+kv.key, "", r)
		if err != nil {
			return Margin{}, err
		}
		*kv.dst = v
	}
	extra, err := extraFields(obj, consumed)
	if err != nil {
		return Margin{}, fmt.Errorf("template: page.margin: %w", err)
	}
	m.Extra = extra
	return m, nil
}

func decodeFonts(raw json.RawMessage) (Fonts, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return nil, fmt.Errorf("template: fonts: %w", err)
	}
	out := Fonts{}
	for _, k := range slices.Sorted(maps.Keys(obj)) {
		v := obj[k]
		chain, err := decodeStringArrayRaw(v)
		if err != nil {
			return nil, newLoadError("fonts."+k, "", string(v), "must be an array of strings: "+err.Error())
		}
		out[k] = chain
	}
	return out, nil
}

func decodeAssets(raw json.RawMessage) (map[string]Asset, error) {
	obj, err := decodeObjectMap(raw)
	if err != nil {
		return nil, fmt.Errorf("template: assets: %w", err)
	}
	out := map[string]Asset{}
	for _, k := range slices.Sorted(maps.Keys(obj)) {
		v := obj[k]
		aObj, err := decodeObjectMap(v)
		if err != nil {
			return nil, newLoadError("assets."+k, "", string(v), "must be an object: "+err.Error())
		}
		consumed := map[string]bool{}
		dataRaw, ok := aObj["data"]
		if !ok {
			return nil, newLoadError("assets."+k+".data", "", "", "missing required field")
		}
		consumed["data"] = true
		data, err := decodeStringArrayRaw(dataRaw)
		if err != nil {
			return nil, newLoadError("assets."+k+".data", "", string(dataRaw), "must be an array of strings: "+err.Error())
		}
		mtRaw, ok := aObj["mediaType"]
		if !ok {
			return nil, newLoadError("assets."+k+".mediaType", "", "", "missing required field")
		}
		consumed["mediaType"] = true
		mt, err := decodeStringRaw(mtRaw)
		if err != nil {
			return nil, newLoadError("assets."+k+".mediaType", "", string(mtRaw), "must be a string: "+err.Error())
		}

		// AC6a (D-1.8.8): validate SHAPE first, then VALUE — two
		// distinct error classes. A key that is not even 64 lowercase
		// hex characters is evidence nothing looked at the key at all;
		// that is a different diagnosis from a well-formed key that
		// does not match its data, and the shape check is the cheaper
		// one, so it runs first.
		if !isSHA256HexKey(k) {
			return nil, newLoadError("assets."+k, "", k, "asset key is not a 64-character lowercase hex digest (AC6a)")
		}

		// AC4/AC1-AC2 (D-1.8.2): accept ANY input wrapping, decode
		// strictly. AC4: invalid base64 and an empty decoded asset are
		// both load errors.
		decoded, derr := decodeBase64Asset(data)
		if derr != nil {
			return nil, newLoadError("assets."+k+".data", "", string(dataRaw), derr.Error())
		}
		if len(decoded) == 0 {
			return nil, newLoadError("assets."+k+".data", "", string(dataRaw), "decoded asset data is empty — it cannot render, and its key would be the SHA-256 of nothing (AC4)")
		}

		// AC5/AC6 (D-1.8.3): the key is the SHA-256 of the DECODED
		// bytes, validated on load; a mismatch is a load error naming
		// both digests.
		gotDigest := sha256HexOf(decoded)
		if gotDigest != k {
			return nil, newLoadError("assets."+k, "", k,
				fmt.Sprintf("asset key does not match the SHA-256 of its data (expected %s, got %s) (AC6)", k, gotDigest))
		}

		// AC9/AC11a/AC11b (D-1.8.1 as amended): a RECOGNISED mediaType
		// whose bytes are not that format is a load error — the file
		// lies about itself, reader-independent. An UNRECOGNISED
		// mediaType is never inspected here and never refused at load
		// (D-1.8.1 amended); it only becomes a located error at render
		// time, only when an element actually needs to draw it
		// (DecodeImageForRender, image.go) — this loader never calls
		// that predicate.
		if _, recognised, ierr := decodeRecognisedImage(mt, decoded); recognised && ierr != nil {
			return nil, newLoadError("assets."+k+".data", "", mt, ierr.Error())
		}

		extra, err := extraFields(aObj, consumed)
		if err != nil {
			return nil, fmt.Errorf("template: assets.%s: %w", k, err)
		}
		out[k] = Asset{Data: data, MediaType: mt, Extra: extra}
	}
	return out, nil
}

// unexpectedKeys returns the keys of obj not present in consumed, sorted
// (D-1.3.5/NFR1.d bans ranging a map anywhere under internal/, so the
// sorted-then-index idiom is used even though only the first result is
// ever read by a caller).
func unexpectedKeys(obj map[string]json.RawMessage, consumed map[string]bool) []string {
	var out []string
	for _, k := range slices.Sorted(maps.Keys(obj)) {
		if !consumed[k] {
			out = append(out, k)
		}
	}
	return out
}

// decodePointsRaw decodes a points-kind numeric field (AC24) from raw.
func decodePointsRaw(field, elementID string, raw json.RawMessage) (geom.Length, error) {
	n, err := decodeNumberRaw(raw)
	if err != nil {
		return 0, newLoadError(field, elementID, string(raw), "must be a JSON number: "+err.Error())
	}
	v, err := decodePoints(string(n))
	if err != nil {
		return 0, newLoadError(field, elementID, string(n), err.Error())
	}
	return v, nil
}
