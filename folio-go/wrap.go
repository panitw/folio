package folio

import (
	"fmt"

	"github.com/panitw/folio/folio-go/internal/bind"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/text"
)

// This file is Story 2.4's measurement and line-packing core: given one
// text element's already-shaped face segments and the positions at
// which it may break, it produces the lines that element occupies.
//
// It emits no PDF and decides no vertical placement. That separation is
// deliberate: everything here is decided by WIDTHS, and widths are
// exactly the part that must not disagree with what is drawn.

// faceSegment is one maximal run of an element's text whose runes all
// resolve to the SAME face, together with the shaper's answer for it.
//
// runeStart/runeEnd are ELEMENT-GLOBAL rune indices, while
// ShapedGlyph.Cluster is SEGMENT-LOCAL. The two units meet in exactly
// one place — segmentLocal — so there is one site that can get the
// conversion wrong rather than one per call.
type faceSegment struct {
	face         string
	segText      string
	runeStart    int
	runeEnd      int
	glyphs       []text.ShapedGlyph
	clusterTexts []string
	unitsPerEm   int64
}

func (s faceSegment) segmentLocal(elementRune int) int { return elementRune - s.runeStart }

// glyphRangeForRunes returns the half-open glyph index range covering
// the element-global rune range [from, to) within this segment.
//
// AC10, AND THIS IS THE WHOLE OF IT: the slice boundary is the FIRST
// GLYPH WHOSE Cluster >= r. Break opportunities never fall inside a
// cluster (AD-25's cluster absolute, and CJK/whitespace breaks are
// cluster boundaries trivially), so that position is exact — no glyph
// straddles the cut.
//
// Cluster IS A RUNE INDEX, NOT A BYTE OFFSET (D-2.3.2, pinned by
// TestClusterValuesAreRuneIndices). Reading it as a byte offset would
// agree on ASCII and diverge on the first Thai character, where byte
// offsets run at three times the rune index. Nothing here converts to
// bytes.
//
// Clusters are non-decreasing across the slice for the left-to-right
// scripts folio ships, but this scan does not ASSUME contiguity: a
// ligature's cluster is followed by the cluster of its last component
// plus one (ClusterTexts' own finding), so the search is a linear scan
// for the first index satisfying the predicate, not arithmetic on
// cluster values.
func (s faceSegment) glyphRangeForRunes(from, to int) (lo, hi int) {
	localFrom := s.segmentLocal(from)
	localTo := s.segmentLocal(to)

	lo = len(s.glyphs)
	for i, g := range s.glyphs {
		if g.Cluster >= localFrom {
			lo = i
			break
		}
	}
	hi = len(s.glyphs)
	for i, g := range s.glyphs {
		if g.Cluster >= localTo {
			hi = i
			break
		}
	}
	if hi < lo {
		hi = lo
	}
	return lo, hi
}

// advance1000 sums the per-glyph advances of glyphs[lo:hi], each scaled
// INDIVIDUALLY to the 1000-unit em.
//
// PER GLYPH, THEN SUMMED — never summed and then scaled. That is the
// already-rounded space the viewer's pen consumes, and it is the same
// order splitByFace's segment cursor and appendShapedRun's advance
// correction use. Changing the order here would reintroduce a second
// derivation of the same quantity, which is precisely Story 2.3's
// Blocker 1.
//
// The advances come from the SAME []text.ShapedGlyph the run is drawn
// from. fontset.AdvanceForRune — the per-rune hmtx path — must never
// appear on this path: shaped advances include GPOS kerning and hmtx
// advances do not, and text drawn kerned but measured unkerned is the
// defect 2.3 removed.
func (s faceSegment) advance1000(lo, hi int) int64 {
	var total int64
	for _, g := range s.glyphs[lo:hi] {
		total += int64(geom.ScaleRound(geom.Length(int64(g.XAdvance)), 1000, s.unitsPerEm))
	}
	return total
}

// measureRuneRange returns the width, in exact integer millipoints, of
// the element-global rune range [from, to) across segments.
//
// Per segment: scale each glyph advance to the 1000-em, sum, then scale
// by font size. Across segments: sum those. This reproduces
// splitByFace's cursor arithmetic exactly, because it IS that
// arithmetic — a line's width and the position of the next face
// segment on it are the same number computed once.
//
// No float32 or float64 appears anywhere on this path (AD-23, D-000.25).
func measureRuneRange(segs []faceSegment, from, to int, fontSize geom.Length) geom.Length {
	var total geom.Length
	for _, s := range segs {
		if to <= s.runeStart || from >= s.runeEnd {
			continue
		}
		lo, hi := s.glyphRangeForRunes(maxInt(from, s.runeStart), minInt(to, s.runeEnd))
		if hi <= lo {
			continue
		}
		total += geom.ScaleRound(geom.Length(s.advance1000(lo, hi)), int64(fontSize), 1000)
	}
	return total
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// wrappedLine is one laid-out line of a text element: the element-global
// rune range it draws, and the width that range measured to.
//
// Width is recorded rather than recomputed by the caller, so the number
// the packer decided with and the number anything downstream asserts on
// are the same number.
type wrappedLine struct {
	from, to int
	width    geom.Length
}

// packLines is the greedy line breaker: given an element's segments,
// its break opportunities and its declared box width, it returns the
// lines the element occupies, in order.
//
// GREEDY, AND DELIBERATELY SO. Each line takes the LAST opportunity
// that still fits. There is no Knuth-Plass paragraph optimisation here
// and none is implied — folio wraps a statement line, not a book.
//
// THE OVERFLOW RULE (AC11, AD-25's own words: "it overflows visibly
// under FR44 — clipped, with a located diagnostic"). When not even the
// FIRST available opportunity fits, the line takes it anyway and
// overflows its box. It is NOT re-broken at a guess, NOT squeezed, and
// NOT silently dropped. The atomic unit that overflowed is an
// unbreakable declared value, an uncoverable Thai run, or an unbroken
// Latin word — in every case something the engine has been told, or has
// determined, it may not split. Clipping and the diagnostic are Story
// 2.8's; this function's obligation is to not paper over it.
//
// maxWidth <= 0 means "no declared box": the element is one line,
// exactly as it rendered before this story. That is what keeps every
// existing golden still (AC13) — though see the Delivery Log for which
// population that claim was measured over.
func packLines(segs []faceSegment, ops []text.Opportunity, totalRunes int, fontSize, maxWidth geom.Length) []wrappedLine {
	if totalRunes == 0 {
		return nil
	}
	if maxWidth <= 0 {
		return []wrappedLine{{from: 0, to: totalRunes, width: measureRuneRange(segs, 0, totalRunes, fontSize)}}
	}

	var lines []wrappedLine
	start := 0
	for start < totalRunes {
		// Does everything that is left fit? Then it is the last line.
		if w := measureRuneRange(segs, start, totalRunes, fontSize); w <= maxWidth {
			lines = append(lines, wrappedLine{from: start, to: totalRunes, width: w})
			break
		}

		chosen := -1
		first := -1
		for i, op := range ops {
			if op.LineEnd <= start {
				continue
			}
			if first < 0 {
				first = i
			}
			// Not short-circuited on the first miss: advances are
			// normally non-negative so width grows with the end
			// position, but GPOS may legitimately produce a negative
			// advance, and a packer that assumed monotonicity would
			// then stop early and under-fill the line for reasons no
			// test would explain.
			if measureRuneRange(segs, start, op.LineEnd, fontSize) <= maxWidth {
				chosen = i
			}
		}

		switch {
		case chosen >= 0:
			op := ops[chosen]
			lines = append(lines, wrappedLine{
				from:  start,
				to:    op.LineEnd,
				width: measureRuneRange(segs, start, op.LineEnd, fontSize),
			})
			start = op.NextStart
		case first >= 0:
			// AC11: the first available unit does not fit. It goes on
			// this line and overflows visibly.
			op := ops[first]
			lines = append(lines, wrappedLine{
				from:  start,
				to:    op.LineEnd,
				width: measureRuneRange(segs, start, op.LineEnd, fontSize),
			})
			start = op.NextStart
		default:
			// No break opportunity remains at all — one atomic unit
			// runs to the end of the element. Same rule: it overflows.
			lines = append(lines, wrappedLine{
				from:  start,
				to:    totalRunes,
				width: measureRuneRange(segs, start, totalRunes, fontSize),
			})
			start = totalRunes
		}
	}
	return lines
}

// atomicSpansFor maps a document's declared unbreakable data paths onto
// the rune spans one element's binding actually substituted.
//
// The declaration is document-level and names DATA PATHS (D-2.4.1); the
// spans are per-element and come from internal/bind's report of what it
// substituted where. This function is the only place the two meet, and
// it produces the parameter internal/text receives — internal/text
// never learns where the spans came from (D-000.16).
//
// Coverage is by CONSTRUCTION over the substitution set: every
// substituted span whose path is declared, whatever its script and
// whatever its content. There is no list of sample values anywhere on
// this path (D-000.23).
func atomicSpansFor(declared []string, subs []bind.Substitution) []text.Span {
	if len(declared) == 0 || len(subs) == 0 {
		return nil
	}
	declaredSet := make(map[string]bool, len(declared))
	for _, p := range declared {
		declaredSet[p] = true
	}
	var spans []text.Span
	for _, s := range subs {
		if declaredSet[s.Path] {
			spans = append(spans, text.Span{Start: s.Start, End: s.End})
		}
	}
	return spans
}

// lineAdvance is THE leading rule (D-2.4.2, ruled): the distance from
// one baseline to the next, for an element drawn with the declared
// chain at fontSize.
//
// It is the MAXIMUM, over the faces of the DECLARED chain that are
// present in fs, of that face's hhea (ascent - descent + lineGap),
// scaled to fontSize. One function, never open-coded at a call site
// (AD-2: "font scaling is one exported function").
//
// # Why the maximum over the chain, and why that is not content-dependent
//
//	A chain declares what MAY appear in an element. Leading must
//	accommodate what MAY appear — not what DOES appear.
//
// "What does appear" is content-dependent, and content-dependent leading
// is forbidden: adding one CJK character would reflow the element, which
// AD-24's "boxes are absolute, and nothing negotiates" rules out. "What
// may appear" is exactly the declared chain — a static property of the
// template, identical for every value the element is ever bound to. So
// this depends on the chain and on nothing that is drawn.
//
// # Why not the chain's first face
//
// Measured, not argued. Against the shipped chain ["Noto Sans", "Noto
// Sans Thai", "Noto Sans SC"], the first face gives 1362/1000 em while
// Noto Sans Thai requires 1511 — at 16 pt that is 21.79 pt of leading
// for text needing 24.18, so Thai below-vowels collide with the next
// line's above-vowels, in the DEFAULT chain, on the script this epic
// exists to support.
//
// # Why hhea, and not OS/2 typo metrics or a multiple of the size
//
// Also measured. Noto Sans SC declares USE_TYPO_METRICS = false
// (fsSelection 0x0040), so its sTypoAscender/Descender of 880/-120 are
// explicitly NOT its line metrics — yet 1000/em is a perfectly plausible
// number, and it is below that face's own 1448. That is the same class
// of fiction as a substituted /CapHeight. And a fixed multiple fails
// outright: 1.2 em is below all three shipped faces (1362, 1511, 1448),
// and even 1.5 em is below Noto Sans Thai.
//
// The bounded cost, stated: a Latin-only element in a Latin+Thai+CJK
// chain gets ~11% taller lines than Noto Sans alone would need. The
// author declared that chain; an author who wants Latin metrics declares
// a Latin-only chain. No element pays for a face its own chain does not
// name.
func lineAdvance(chain []string, fontSize geom.Length, fs FontSet, cache *fontCache) (geom.Length, error) {
	var maxUnits int64
	present := 0
	for _, name := range chain {
		if _, ok := fs[name]; !ok {
			// A chain member the caller did not supply cannot appear in
			// the element, so it does not constrain the leading. This is
			// the same "first chain entry present in fs" tolerance
			// fontChain and resolveRuneFace already apply.
			continue
		}
		f, err := cache.get(name, fs)
		if err != nil {
			return 0, err
		}
		lm := f.LineMetrics()
		units := lm.Ascent - lm.Descent + lm.LineGap
		if units > maxUnits {
			maxUnits = units
		}
		present++
	}
	if present == 0 {
		return 0, fmt.Errorf(
			"folio: none of the fallback chain's faces %v is present in the supplied FontSet, so no line height can be derived from it",
			chain,
		)
	}
	if maxUnits <= 0 {
		return 0, fmt.Errorf(
			"folio: the fallback chain %v yields a line height of %d font units — every face in it declares an hhea ascent no greater than its descent",
			chain, maxUnits,
		)
	}
	return geom.ScaleRound(geom.Length(maxUnits), int64(fontSize), 1000), nil
}
