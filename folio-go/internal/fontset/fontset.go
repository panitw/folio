// Package fontset resolves and subsets fonts supplied by the caller
// (AC2). It performs the module's font seam entirely against explicit
// input bytes: it never embeds font data (AD-8's Rule: "no package
// under internal/ embeds font data") and never queries the host for
// installed fonts (AC4).
//
// The parsed-font type this package exposes deliberately does not leak
// *ot.Font, *ot.Head or *subset.Plan to any caller — D-1.5.10's binding
// form of D-1.5.2's guardrail (AC17a): "A vendor type we cannot
// constrain must not become part of a seam we rely on being
// constrained." The third-party ot.Head.UnitsPerEm field is an exported,
// unvalidated uint16 (F-4); this package's own Font type wraps the
// parsed face privately and exposes only UnitsPerEm(), which is the
// validated value and the ONLY one reachable through this seam (AC16,
// AC17).
package fontset

import (
	"fmt"
	"maps"
	"slices"

	"github.com/boxesandglue/textshape/ot"
	"github.com/boxesandglue/textshape/subset"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// MinUnitsPerEm and MaxUnitsPerEm are AC16's valid unitsPerEm range,
// inclusive at both ends. Zero is explicitly out of range — a font
// reporting unitsPerEm == 0 is a located load error, not a panic
// (AC19), because 0 would reach geom.ScaleRound as a zero denominator.
const (
	MinUnitsPerEm = 16
	MaxUnitsPerEm = 16384
)

// Font is a caller-supplied font, parsed and validated at construction
// (AC16, AC17). Every exported accessor returns only validated,
// primitive-typed data — never a vendor pointer type (AC17a).
type Font struct {
	name       string
	face       *ot.Face
	unitsPerEm uint16 // validated: MinUnitsPerEm <= unitsPerEm <= MaxUnitsPerEm
	created    int64  // head.created, offset 20, LONGDATETIME (F-5)
	modified   int64  // head.modified, offset 28, LONGDATETIME (F-5)
}

// New parses data as an OpenType/TrueType font named name (the name the
// caller's FontSet keyed it under, used only for error messages) and
// validates unitsPerEm before returning. Validation happens here, in the
// constructor, so that the validated value is the only one reachable
// through this seam (AC17, D-1.5.2: "the parsed-font type must not
// expose a raw, unvalidated unitsPerEm that a caller can hand to
// ScaleRound"). An out-of-range value — including 0 — is a located
// error naming the font and the value, never a panic (AC16, AC19).
func New(name string, data []byte) (*Font, error) {
	parsed, err := ot.ParseFont(data, 0)
	if err != nil {
		return nil, fmt.Errorf("fontset: font %q: parse: %w", name, err)
	}
	face, err := ot.NewFace(parsed)
	if err != nil {
		return nil, fmt.Errorf("fontset: font %q: build face: %w", name, err)
	}

	// Read the head table's unitsPerEm DIRECTLY, rather than trusting
	// (*ot.Face).Upem(): that vendor accessor silently substitutes 1000
	// when the parsed head reports 0 ("Default for CFF" — ot/metrics.go
	// NewFace), which would make unitsPerEm == 0 unobservable through
	// Face and defeat AC19's "including 0" retained fixture.
	rawUpem, herr0 := readUnitsPerEm(parsed)
	if herr0 != nil {
		return nil, fmt.Errorf("fontset: font %q: read head table: %w", name, herr0)
	}

	upem := rawUpem
	if upem < MinUnitsPerEm || upem > MaxUnitsPerEm {
		return nil, fmt.Errorf(
			"fontset: font %q: unitsPerEm %d is out of the valid range [%d,%d]",
			name, upem, MinUnitsPerEm, MaxUnitsPerEm,
		)
	}

	created, modified, herr := readHeadTimes(parsed)
	if herr != nil {
		return nil, fmt.Errorf("fontset: font %q: read head table: %w", name, herr)
	}

	return &Font{name: name, face: face, unitsPerEm: upem, created: created, modified: modified}, nil
}

// parseHead reads and parses the head table once. *ot.Font and *ot.Head
// never leave this function or its two callers below (AC17a).
func parseHead(font *ot.Font) (*ot.Head, error) {
	data, err := font.TableData(ot.TagHead)
	if err != nil {
		return nil, err
	}
	return ot.ParseHead(data)
}

// readHeadTimes reads the head table's created/modified fields directly
// (offsets 20 and 28, LONGDATETIME).
func readHeadTimes(font *ot.Font) (created, modified int64, err error) {
	head, err := parseHead(font)
	if err != nil {
		return 0, 0, err
	}
	return head.Created, head.Modified, nil
}

// readUnitsPerEm reads the head table's unitsPerEm field (offset 18)
// directly — see the caller's comment for why this must not go through
// (*ot.Face).Upem(), which silently substitutes a default of 1000 when
// the raw value is 0.
func readUnitsPerEm(font *ot.Font) (uint16, error) {
	head, err := parseHead(font)
	if err != nil {
		return 0, err
	}
	return head.UnitsPerEm, nil
}

// Name returns the caller-supplied face name this Font was constructed
// with.
func (f *Font) Name() string { return f.name }

// UnitsPerEm returns the validated unitsPerEm value (AC17): the only
// unitsPerEm reachable through this seam. Every value derived from a
// caller-supplied font that reaches geom.ScaleRound is scaled against
// this validated value, never against a raw vendor field.
func (f *Font) UnitsPerEm() uint16 { return f.unitsPerEm }

// HeadTimes returns the source face's head.created / head.modified
// values, verbatim (LONGDATETIME: seconds since 1904). AC11a asserts
// the embedded subset's copies equal these exactly.
func (f *Font) HeadTimes() (created, modified int64) { return f.created, f.modified }

// Metrics is the subset of font-wide metrics a PDF FontDescriptor
// needs, scaled to a 1000-unit em via geom.ScaleRound — this module's
// one scaling function (AD-2, AC18) — so no caller outside this package
// ever converts a raw font-unit value itself.
type Metrics struct {
	Ascent    int64
	Descent   int64
	CapHeight int64
	BBoxXMin  int64
	BBoxYMin  int64
	BBoxXMax  int64
	BBoxYMax  int64
}

// Metrics returns f's font-wide metrics, scaled to a 1000-unit em.
func (f *Font) Metrics() Metrics {
	scale := func(v int16) int64 {
		return int64(geom.ScaleRound(geom.Length(int64(v)), 1000, int64(f.unitsPerEm)))
	}
	xMin, yMin, xMax, yMax := f.face.BBox()
	return Metrics{
		Ascent:    scale(f.face.Ascender()),
		Descent:   scale(f.face.Descender()),
		CapHeight: scale(f.face.CapHeight()),
		BBoxXMin:  scale(xMin),
		BBoxYMin:  scale(yMin),
		BBoxXMax:  scale(xMax),
		BBoxYMax:  scale(yMax),
	}
}

// Subset is one face's subsetted, embeddable output (AC5, AC9): a
// standalone TrueType program covering exactly the glyphs the supplied
// runes resolve to (plus whatever closure the vendor subsetter itself
// adds — composite components, GSUB-reachable glyphs, .notdef), a
// six-letter tag derived from the closure set (AC6, AC7), and enough
// per-glyph data to drive a PDF Type0/CIDFontType2 embedding.
type Subset struct {
	// Program is the FontFile2 payload: a complete, standalone TrueType
	// font containing only the retained glyphs.
	Program []byte

	// Tag is AC6's six-letter (A-Z) subset tag.
	Tag string

	// NumGlyphs is the number of glyphs in the OUTPUT numbering
	// (always a dense {0..NumGlyphs-1} range — AC7a's rejected-reading
	// warning: this is NOT what the tag is derived from).
	NumGlyphs int

	// GlyphForRune maps each requested, resolvable rune to its glyph id
	// in the OUTPUT (subset) numbering — the CID a content stream Tj
	// operator addresses under Identity-H.
	GlyphForRune map[rune]uint16

	// WidthForGlyph maps every OUTPUT glyph id to its horizontal
	// advance, scaled to a 1000-unit em (the PDF /W array's unit),
	// rounded via geom.ScaleRound — this module's one scaling function
	// (AD-2, AC18).
	WidthForGlyph map[uint16]int64
}

// Subset builds one subset over the union of runes supplied (AC9: one
// subset call per font per document, over everything the document
// uses). The caller's rune ordering is IRRELEVANT and this function
// never sorts its input (AC8, D-1.5.7): runes is deduplicated into a
// Go map before anything is handed to the vendor subsetter, so any
// permutation of the same rune set produces byte-identical output.
// Sorting here would make that property untestable (AC8a: "a defensive
// normalisation upstream of a check can delete that check's
// discriminating power") — do not reintroduce it.
func (f *Font) Subset(runes []rune) (*Subset, error) {
	cmap := f.face.Cmap()
	if cmap == nil {
		return nil, fmt.Errorf("fontset: font %q has no cmap table", f.name)
	}

	// Deduplicate by ranging the CALLER'S SLICE, never a map (D-1.3.5's
	// ScanMapRange bans ranging a map anywhere under internal/, with no
	// site-specific exception — and D-1.5.7 separately forbids sorting
	// this particular site, since a sort here would make AC8's
	// permutation-invariance proof vacuous, AC8a). Both rulings are
	// satisfied at once by never constructing a map whose KEYS this
	// function ranges: `seen` and `oldGIDForRune` below are read by
	// direct index, only ever written and looked up — never ranged —
	// so gids ends up in exactly the caller's rune order (deduplicated
	// by first occurrence), which is what varies from call to call when
	// AC8's repeat/permutation-invariance tests permute their input.
	seen := map[rune]bool{}
	oldGIDForRune := map[rune]ot.GlyphID{}
	var gids []ot.GlyphID
	for _, r := range runes {
		if seen[r] {
			continue
		}
		seen[r] = true
		gid, ok := cmap.Lookup(ot.Codepoint(r))
		if !ok {
			return nil, fmt.Errorf("fontset: font %q: no glyph for rune %U", f.name, r)
		}
		oldGIDForRune[r] = gid
		gids = append(gids, gid)
	}

	input := subset.NewInput()
	input.AddGlyphs(gids...) // do NOT sort gids first — AC8, D-1.5.7.

	plan, err := subset.CreatePlan(f.face.Font, input)
	if err != nil {
		return nil, fmt.Errorf("fontset: font %q: create subset plan: %w", f.name, err)
	}

	program, err := plan.Execute()
	if err != nil {
		return nil, fmt.Errorf("fontset: font %q: execute subset: %w", f.name, err)
	}

	tag, err := deriveTag(plan.GlyphSet())
	if err != nil {
		return nil, fmt.Errorf("fontset: font %q: derive subset tag: %w", f.name, err)
	}

	// Ranges `runes` (a slice) again, not oldGIDForRune (a map) — same
	// ScanMapRange reasoning as above.
	glyphForRune := map[rune]uint16{}
	for _, r := range runes {
		if _, already := glyphForRune[r]; already {
			continue
		}
		oldGID := oldGIDForRune[r]
		newGID, ok := plan.MapGlyph(oldGID)
		if !ok {
			return nil, fmt.Errorf("fontset: font %q: rune %U's glyph was not retained by the subset plan", f.name, r)
		}
		glyphForRune[r] = newGID
	}

	numOut := plan.NumOutputGlyphs()
	widthForGlyph := make(map[uint16]int64, numOut)
	for newGID := 0; newGID < numOut; newGID++ {
		oldGID, ok := plan.OldGlyph(ot.GlyphID(newGID))
		if !ok {
			return nil, fmt.Errorf("fontset: font %q: output glyph %d has no source glyph", f.name, newGID)
		}
		adv := f.face.HorizontalAdvance(oldGID)
		width1000 := geom.ScaleRound(geom.Length(int64(adv)), 1000, int64(f.unitsPerEm))
		widthForGlyph[uint16(newGID)] = int64(width1000)
	}

	return &Subset{
		Program:       program,
		Tag:           tag,
		NumGlyphs:     numOut,
		GlyphForRune:  glyphForRune,
		WidthForGlyph: widthForGlyph,
	}, nil
}

// deriveTag is AC6/AC7: the six-letter A-Z subset tag, derived from a
// hash of Plan.GlyphSet() — the closure set, in SOURCE-font glyph
// numbering (D-1.5.8) — sorted ascending BEFORE hashing. This is the
// one site in this story where sorting is correct and required (AC7,
// AC6): the ordering is ours, it reaches an output byte (the tag), and
// it comes from ranging a map, so Story 1.3's ScanMapRange guard is
// load-bearing here and must not be worked around — the escape hatch
// idiom (slices.Sorted(maps.Keys(...))) is used deliberately, not
// incidentally.
//
// AC7a records the two rejected readings and why: hashing the
// requested set (not the closure) lets two different embedded programs
// share a tag; hashing the OUTPUT numbering is catastrophic, because
// createCompactMapping always produces {0..n-1} regardless of which
// glyphs those are, making the tag a function of glyph count alone.
//
// D-1.1.b, verbatim, on why this six-letter tag is not routed through
// internal/pdf's numeric emitters: "Glyph ids under Identity-H take
// neither route — they are a big-endian hex pair inside a string
// literal, i.e. a byte encoding like /ID, not a number. Same for the
// six-letter subset tag. This must be said in a code comment or a
// later agent will 'unify' it."
func deriveTag(glyphSet map[ot.GlyphID]bool) (string, error) {
	if len(glyphSet) == 0 {
		return "", fmt.Errorf("empty glyph set")
	}

	sorted := slices.Sorted(maps.Keys(glyphSet)) // ScanMapRange-compliant: sort before use.

	digest := fnv64aOverGIDs(sorted)

	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var tag [6]byte
	v := digest
	for i := 0; i < 6; i++ {
		tag[i] = letters[v%26]
		v /= 26
	}
	return string(tag[:]), nil
}

// fnv64aOverGIDs folds a sorted glyph-id sequence into a single uint64
// using FNV-1a, big-endian per glyph id — a fixed, deterministic,
// integer-only digest (no floating point, no wall-clock, no
// randomness) suitable for AC6's "hash of the sorted glyph-id set".
func fnv64aOverGIDs(sorted []ot.GlyphID) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	h := uint64(offset64)
	for _, gid := range sorted {
		b := [2]byte{byte(gid >> 8), byte(gid)}
		for _, c := range b {
			h ^= uint64(c)
			h *= prime64
		}
	}
	return h
}
