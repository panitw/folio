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

	"github.com/boxesandglue/textshape/ot"
	"github.com/boxesandglue/textshape/subset"

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/text"
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
	psName     string // name table record 6, read directly (see readPostScriptName)

	// shaper is this face's ONE OpenType shaper, built at construction
	// and reused for every shaping call against this face — the
	// vendor's documented contract ("A Shaper is created once per font
	// and reused across shaping calls"). Its type is folio's own
	// text.Shaper, not *ot.Shaper, so no vendor pointer crosses this
	// package's seam (AC17a, D-1.5.10).
	shaper *text.Shaper
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

	// Story 2.2 / D-2.2.4 (binding): a caller-supplied VARIABLE face is
	// REJECTED here, at face ingestion — not instanced, and not
	// mid-render. PDF 1.7 cannot express a variable font (AD-7 pins the
	// profile), so a variable face cannot be embedded as supplied, and
	// the two ways of coping with that in-process are both closed:
	//
	//   - Pinning every axis to its DEFAULT (what this package did
	//     before) is not a neutral act. NotoSansSC-VF's `wght` axis
	//     defaults to 100, so "the default instance" embedded
	//     Simplified Chinese as THIN beside Regular Latin and Regular
	//     Thai, and every guard in the project agreed the artifact was
	//     correct because each was asked whether the value CHANGED,
	//     never whether it meant what its name implied (D-000.21).
	//   - Pinning an explicitly chosen weight is unreachable: the
	//     vendor's PinAxisLocation requires the identifier `float32`,
	//     which internal/arch_test.go:54 bans under internal/ AND the
	//     module root (AD-23). And it would not have worked anyway —
	//     textshape@v0.0.15 subset/execute.go:496-499 copies `OS/2`
	//     VERBATIM and never updates usWeightClass (there is no writer
	//     for that field anywhere in textshape), so pinning wght=400 on
	//     a variable face yields Regular outlines carrying metadata that
	//     still claims Thin — strictly worse than the defect it fixes,
	//     because outlines and metadata would then disagree.
	//
	// So the seam is DELETED rather than fixed. folio's own shipped
	// faces arrive already static (folio-go/fonts/, derived by
	// tools/fontgen/instance_faces.py, which pins the axes AND lets
	// fontTools write the correct usWeightClass), and a caller bringing
	// their own variable face is told, early and by name, to instance it
	// first. This removes the float `gvar` interpolation path from the
	// render entirely — the FMA hazard D-2.2.0 measured stops being
	// monitored and stops existing.
	//
	// The message NAMES THE REMEDY on purpose: most Google Fonts
	// downloads are variable builds today, so a caller hitting this
	// needs an action, not a refusal.
	if parsed.HasTable(ot.TagFvar) {
		return nil, fmt.Errorf(
			"fontset: font %q: this is a VARIABLE font (it has an `fvar` table) and cannot be embedded: "+
				"PDF 1.7 cannot express a variable font, and folio will not silently pick an instance for you "+
				"(the default instance is not always Regular — Noto Sans SC's `wght` axis defaults to 100, i.e. Thin). "+
				"Instance it to a single static weight first, e.g. "+
				"`fonttools varLib.instancer --update-name-table %s.ttf wght=400 -o %s-Regular.ttf`, "+
				"pinning EVERY axis the face declares rather than just `wght` "+
				"(see tools/fontgen/instance_faces.py for the full recipe folio uses for its own shipped faces, "+
				"including the reproducibility pin the bare command above omits)",
			name, name, name,
		)
	}

	psName, perr := readPostScriptName(parsed)
	if perr != nil {
		return nil, fmt.Errorf("fontset: font %q: read name table: %w", name, perr)
	}

	shaper, serr := text.NewShaper(name, face)
	if serr != nil {
		return nil, fmt.Errorf("fontset: font %q: %w", name, serr)
	}

	return &Font{
		name:       name,
		face:       face,
		unitsPerEm: upem,
		created:    created,
		modified:   modified,
		psName:     psName,
		shaper:     shaper,
	}, nil
}

// Shaper returns this face's single, reused OpenType shaper (Story 2.3,
// AC1). One per Font, constructed in New, never one per call and never
// one per run.
//
// The returned type is folio's own internal/text.Shaper. This accessor
// deliberately does NOT expose *ot.Face or *ot.Shaper: D-1.5.10's
// binding form of D-1.5.2 — "a vendor type we cannot constrain must not
// become part of a seam we rely on being constrained" — and the whole
// reason this package wraps the parsed face privately in the first
// place.
func (f *Font) Shaper() *text.Shaper { return f.shaper }

// PostScriptName returns the face's own PostScript name — `name` table
// record 6, read off the supplied font program itself, never derived
// from the FontSet key the caller happened to file it under.
//
// It exists because ISO 32000-1 Table 117 (CIDFontType2) requires
// /BaseFont to be "the value of the CIDFontName entry in the CIDFont
// program". Before Story 2.2 folio spelled /BaseFont from the FontSet
// key, so the declared name and the embedded program disagreed
// (`NotoSansSC` vs `NotoSansSC-Regular`) — a real conformance defect:
// PDF/A validators flag it, and it is the name a viewer falls back to
// when the embedded program fails to load.
//
// Fixing it also converts an INVISIBLE property into a visible one,
// which is the more valuable half. The Thin defect (D-000.21) hid
// precisely because nothing in the output carried the font's own
// identity; now that /BaseFont reflects the embedded program, a future
// weight defect of this class is legible by reading the PDF, not only
// through the dedicated assertion we happened to remember to write.
//
// Returns "" when the supplied program declares no usable name record;
// the caller decides what to do about that (internal/pdf falls back to
// the FontSet key, which is the pre-2.2 behaviour). This accessor
// returns a plain string and never leaks *ot.Font, *ot.Name or any
// other vendor type through the seam (AC17a, D-1.5.10).
func (f *Font) PostScriptName() string { return f.psName }

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

// readPostScriptName reads `name` table record 6 directly, rather than
// through (*ot.Face).PostscriptName(). That vendor accessor returns the
// literal string "Unknown" when the parsed font carries no name table
// (ot/metrics.go:415-420) — a silent substitution of exactly the kind
// the readUnitsPerEm comment below documents for Upem(), and one that
// would put the PDF name /Unknown into /BaseFont while every assertion
// downstream reported a well-formed string. A missing name table is
// reported here as "", which is observably absent.
//
// A missing or unparseable `name` table is NOT an ingestion error: it is
// not part of what this seam validates (AC16/AC17 validate unitsPerEm),
// and a face that renders correctly should not become unloadable over a
// metadata table. The guard that this value is RIGHT for the faces folio
// actually ships is the semantic acceptance assertion on the produced
// PDF (D-000.22), which pins /BaseFont to ^[A-Z]{6}\+<name6>$ per face.
func readPostScriptName(font *ot.Font) (string, error) {
	if !font.HasTable(ot.TagName) {
		return "", nil
	}
	data, err := font.TableData(ot.TagName)
	if err != nil {
		return "", err
	}
	parsed, err := ot.ParseName(data)
	if err != nil {
		return "", err
	}
	return parsed.PostScriptName(), nil
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

// HasGlyph reports whether f's cmap maps r to any glyph (Story 2.2,
// AC4): the COVERAGE test the missing-glyph diagnostic is built on —
// evaluated per rune, against each candidate face's own cmap, in chain
// order. Never a proxy such as a locale check or "is this the preferred
// face for the script" (D-2.2-D4).
func (f *Font) HasGlyph(r rune) bool {
	cmap := f.face.Cmap()
	if cmap == nil {
		return false
	}
	_, ok := cmap.Lookup(ot.Codepoint(r))
	return ok
}

// AdvanceForRune returns r's raw `hmtx` horizontal advance in f, scaled
// to a 1000-unit em (the same scale (*Subset).WidthForGlyph uses), and
// whether r resolved to a glyph at all.
//
// NOT A POSITIONING FUNCTION, and this is a correction rather than a
// caveat. Story 2.2 used it to place a multi-face run's second and later
// face-segments; Story 2.3's finisher removed that call site (Blocker 1)
// because the value is the UNKERNED `hmtx` advance and runs are now
// drawn KERNED. Summing it over a segment's runes yields a cursor that
// disagrees with what was drawn wherever GPOS kerns — measured at 640
// millipoints for "AV ก" at 16 pt in the shipped chain. The segment
// cursor is derived from the SHAPED advances instead
// (folio-go/render.go's splitByFace), and nothing in the render path
// calls this any more.
//
// It is retained deliberately, not left as an oversight: it is one of
// the two `ot.Face.HorizontalAdvance` (`float32`) call sites that
// `2-3a-audit-the-vendor-boundary` (D-000.25) owns and audits, and
// deleting it here would silently shrink that story's subject. Any
// future caller must first establish that the raw `hmtx` number, not
// the shaped one, is what it wants.
func (f *Font) AdvanceForRune(r rune) (int64, bool) {
	cmap := f.face.Cmap()
	if cmap == nil {
		return 0, false
	}
	gid, ok := cmap.Lookup(ot.Codepoint(r))
	if !ok {
		return 0, false
	}
	adv := f.face.HorizontalAdvance(gid)
	return int64(geom.ScaleRound(geom.Length(int64(adv)), 1000, int64(f.unitsPerEm))), true
}

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
// SOURCE GLYPH IDS name (plus whatever closure the vendor subsetter
// itself adds — composite components, GSUB-reachable glyphs, .notdef),
// a six-letter tag derived from the program bytes (AC6, AC7), and
// enough per-glyph data to drive a PDF Type0/CIDFontType2 embedding.
//
// Story 2.3, AC5: the input is the set of glyphs the SHAPED runs
// actually contain, never the set of runes the document contains. That
// is D-1.5.8's own rule applied one level up — "assert on the produced
// thing, never on the thing you asked for" — and the two sets are
// measurably different: shaping "office" in Noto Sans draws glyph 1656,
// the `ffi` ligature, which no rune maps to at all.
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

	// GlyphForSource maps each requested SOURCE glyph id to its glyph id
	// in the OUTPUT (subset) numbering, obtained from plan.MapGlyph.
	//
	// It is NOT a CID map. Story 2.3 allocates CIDs per (subset glyph,
	// /ToUnicode context) rather than per glyph — see internal/pdf's
	// CIDEntry — because one glyph legitimately needs two different
	// /ToUnicode answers in one document (the tail of "น้ำ" and a
	// standalone "า" are the same glyph). Anything inverting this map
	// must not assume CIDs and glyph ids are 1:1.
	GlyphForSource map[uint16]uint16

	// WidthForGlyph maps every OUTPUT glyph id to its horizontal
	// advance, scaled to a 1000-unit em (the PDF /W array's unit),
	// rounded via geom.ScaleRound — this module's one scaling function
	// (AD-2, AC18).
	WidthForGlyph map[uint16]int64
}

// Subset builds one subset over the union of SOURCE GLYPH IDS supplied
// (AC9: one subset call per font per document, over everything the
// document draws). The caller's ordering is IRRELEVANT and this function
// never sorts its input (AC8, D-1.5.7): the slice is deduplicated by
// first occurrence, so any permutation of the same glyph set produces
// byte-identical output. Sorting here would make that property
// untestable (AC8a: "a defensive normalisation upstream of a check can
// delete that check's discriminating power") — do not reintroduce it.
//
// Story 2.2 (D-2.2.4, binding): this function does NO instancing. PDF
// 1.7 cannot express a variable font (AD-7), and folio's answer is that
// a variable face never gets this far — New() rejects one at ingestion
// with a located diagnostic naming the remedy. Everything arriving here
// is already a single static instance, so there is no axis to pin, no
// `gvar` to interpolate, and no float arithmetic on the render path.
//
// This package still exposes NO way for a caller to request a
// particular instance, and that is deliberate rather than unfinished:
// reaching the vendor's PinAxisLocation requires the identifier
// `float32`, which internal/arch_test.go:54 bans under internal/ and
// the module root (AD-23). Bold, when it arrives, is a `wght` instance
// derived AHEAD of the build by tools/fontgen/instance_faces.py and
// shipped as its own static face — and it inherits D-2.2.1's standing
// condition (its own golden, plus a four-target matrix run) so it does
// not have to rediscover why.
//
// This package's own tag-discrimination test (V5) exercises a second,
// non-default instance by calling the vendor subsetter directly,
// entirely from the test file, never through this method.
func (f *Font) Subset(sourceGlyphs []uint16) (*Subset, error) {
	// Deduplicate by ranging the CALLER'S SLICE, never a map (D-1.3.5's
	// ScanMapRange bans ranging a map anywhere under internal/, with no
	// site-specific exception — and D-1.5.7 separately forbids sorting
	// this particular site, since a sort here would make AC8's
	// permutation-invariance proof vacuous, AC8a). Both rulings are
	// satisfied at once by never constructing a map whose KEYS this
	// function ranges: `seen` below is read by direct index, only ever
	// written and looked up — never ranged — so gids ends up in exactly
	// the caller's order (deduplicated by first occurrence), which is
	// what varies from call to call when AC8's repeat/permutation-
	// invariance tests permute their input.
	// Every source id is validated against the FACE's own glyph count
	// before it reaches the vendor (Story 2.3 finisher, Finding 8).
	//
	// This is a real guard on a reachable path, and it is what makes the
	// "not retained" check below honest. Since AC5 re-keyed this function
	// on GLYPH IDS, a caller can hand it an id the face does not have —
	// and the vendor does not report that, it FABRICATES a glyph.
	// Measured against fonts/notosans/NotoSans-Regular.ttf (4,515
	// glyphs): Subset([]uint16{65535}) returned a 460-byte program and a
	// complete mapping, with no error and nothing unmapped. That is
	// D-000.25's Finding 2 shape — "vendor accessors substitute plausible
	// defaults for missing data" — arriving at a call site this story
	// created, and the resulting bytes are a wrong glyph embedded in a
	// document that reports success.
	numGlyphs := f.face.Font.NumGlyphs()
	if numGlyphs <= 0 {
		return nil, fmt.Errorf("fontset: font %q: face reports %d glyphs", f.name, numGlyphs)
	}

	seen := map[uint16]bool{}
	var gids []ot.GlyphID
	var uniqueSources []uint16
	for _, g := range sourceGlyphs {
		if seen[g] {
			continue
		}
		if int(g) >= numGlyphs {
			return nil, fmt.Errorf(
				"fontset: font %q: glyph id %d does not exist in this face, which has %d glyphs (0..%d) — "+
					"the subsetter would fabricate one rather than report it missing",
				f.name, g, numGlyphs, numGlyphs-1,
			)
		}
		seen[g] = true
		uniqueSources = append(uniqueSources, g)
		gids = append(gids, ot.GlyphID(g))
	}

	input := subset.NewInput()
	input.AddGlyphs(gids...) // do NOT sort gids first — AC8, D-1.5.7.

	// There is NO instancing seam here any more (D-2.2.4, binding). Every
	// font reaching this point is already static: New() rejects a face
	// carrying `fvar` at ingestion, with a located diagnostic that names
	// the remedy. See New()'s comment for why the seam was deleted
	// rather than fixed — pinning to defaults embedded Thin, and pinning
	// explicitly is both unreachable (AD-23 bans `float32`) and
	// ineffective (textshape never rewrites OS/2.usWeightClass).
	//
	// Do not reintroduce a pin here. The consequence of this deletion is
	// that the float `gvar` interpolation path — D-2.2.0's measured FMA
	// hazard, the reason this story carries a four-target matrix at all
	// — is not merely monitored, it is unreachable from the render.
	plan, err := subset.CreatePlan(f.face.Font, input)
	if err != nil {
		return nil, fmt.Errorf("fontset: font %q: create subset plan: %w", f.name, err)
	}

	program, err := plan.Execute()
	if err != nil {
		return nil, fmt.Errorf("fontset: font %q: execute subset: %w", f.name, err)
	}

	// AC6 / D-2.2.2 (superseded): the six-letter subset tag hashes the
	// RETURNED PROGRAM BYTES — the embedded font program itself, in
	// full — never plan.GlyphSet() alone (which collided for two
	// pinned instances of one face sharing a glyph-id set, B6) and
	// never any axis coordinate in any form, exact or float (D-1.5.8:
	// a tag keyed on the request lies about what is embedded).
	tag, err := deriveTag(program)
	if err != nil {
		return nil, fmt.Errorf("fontset: font %q: derive subset tag: %w", f.name, err)
	}

	// Ranges `uniqueSources` (a slice), never a map — same ScanMapRange
	// reasoning as above.
	//
	// A shaped glyph the plan did not retain is a LOCATED ERROR naming
	// the face and the glyph id (AC5), never a silent substitution and
	// never .notdef: if this fires, the set handed to the subsetter and
	// the set the renderer is about to draw have disagreed, and emitting
	// a blank box would make that disagreement invisible.
	//
	// UNREACHABLE VENDOR-CONTRACT ASSERTION, LABELLED AS SUCH RATHER THAN
	// COUNTED AS COVERAGE (D-000.24; Story 2.3 finisher, Finding 8).
	// plan.MapGlyph reads p.glyphMap, and createCompactMapping inserts
	// every member of glyphSet — which is this function's input plus its
	// closure — so a glyph passed to AddGlyphs is ALWAYS mapped. Coverage
	// confirms it: this branch's count is 0, and no test in the repo
	// matches "was not retained" against this package. It reads like
	// coverage in a diff and cannot fire; saying so is the honest form.
	//
	// It is kept, not deleted, because it is the assertion that would
	// catch the vendor changing that contract — and because AC5's LIVE
	// guard is the render-side one (render.go's buildShapedPDFRuns,
	// red-proved by dropping the `ffi` ligature from the subset union),
	// not this. The reachable half of this seam is the glyph-count
	// validation above, which IS red-proved.
	glyphForSource := map[uint16]uint16{}
	for _, src := range uniqueSources {
		newGID, ok := plan.MapGlyph(ot.GlyphID(src))
		if !ok {
			return nil, fmt.Errorf(
				"fontset: font %q: shaped glyph id %d was not retained by the subset plan",
				f.name, src,
			)
		}
		glyphForSource[src] = newGID
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
		Program:        program,
		Tag:            tag,
		NumGlyphs:      numOut,
		GlyphForSource: glyphForSource,
		WidthForGlyph:  widthForGlyph,
	}, nil
}

// deriveTag is AC6/AC6a's ruled derivation (`D-2.2.2 (superseded)`): the
// six-letter A-Z subset tag, derived from a hash of `program` — the
// COMPLETE embedded font program `subset.Subset()` (here, plan.Execute())
// returned, byte for byte. Not the glyph-id set (B6: two pinned
// instances of one face collided under that reading, since instancing
// changes outlines without changing which glyph ids are retained), not
// any axis coordinate in any form, exact or float (D-1.5.8: a tag keyed
// on the request lies about what is embedded).
//
// Zero float anywhere in this function, by construction: `program` is
// already a []byte, so there is no map to sort and no ScanMapRange
// concern at this site at all — this derivation needs no escape hatch
// because it never ranges a map in the first place.
//
// D-1.1.b, verbatim, on why this six-letter tag is not routed through
// internal/pdf's numeric emitters: "Glyph ids under Identity-H take
// neither route — they are a big-endian hex pair inside a string
// literal, i.e. a byte encoding like /ID, not a number. Same for the
// six-letter subset tag. This must be said in a code comment or a
// later agent will 'unify' it."
func deriveTag(program []byte) (string, error) {
	if len(program) == 0 {
		return "", fmt.Errorf("empty font program")
	}

	digest := fnv64aOverBytes(program)

	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var tag [6]byte
	v := digest
	for i := 0; i < 6; i++ {
		tag[i] = letters[v%26]
		v /= 26
	}
	return string(tag[:]), nil
}

// fnv64aOverBytes folds a byte sequence into a single uint64 using
// FNV-1a — a fixed, deterministic, integer-only digest (no floating
// point, no wall-clock, no randomness) suitable for AC6's "hash of the
// returned program bytes".
func fnv64aOverBytes(data []byte) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	h := uint64(offset64)
	for _, c := range data {
		h ^= uint64(c)
		h *= prime64
	}
	return h
}
