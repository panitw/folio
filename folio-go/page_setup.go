package folio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/template"
	"github.com/panitw/folio/folio-go/internal/text"
)

// GridIncrement is the fixed six-point grid used by the designer projection.
// SnapNearest's documented midpoint rule is away from zero.
const GridIncrement int64 = 6000

// MaxCanvasMillipoints keeps every document value emitted to the JSON/JS paint
// boundary within Number.MAX_SAFE_INTEGER. The page-setup command has the same
// bound, so a successful command can never strand the worker with an
// unrepresentable projection.
const MaxCanvasMillipoints int64 = 9007199254740991

// maxCanvasPropertyString bounds an IDENTIFIER, a COLOUR or an EXPRESSION —
// a font-family name, `color`, `background`, `border.color`, `visibleIf` and
// a table's `bind`. Seven sites, all legitimately short, and all of them
// still ABORT the projection: Epic 7 makes none of them newly reachable, so
// that residue is recorded rather than fixed (DW-25).
//
// It used to bound document body text as well, which is the two-jobs
// conflation D-7.4.2 §3 ruled must be SPLIT rather than raised: 512 bytes is
// ~80 English words, or ~170 Thai/CJK characters at three bytes each, which
// is less than one numbered contract clause. Body text now has its own
// derived bounds below; this constant governs identifiers only.
const maxCanvasPropertyString = 512

// maxCanvasTextFragments is the PER-LINE fragment guard, unchanged at 512 and
// deliberately a different quantity from the cumulative per-element budget
// below. It bounds one degenerate line; the browser mirrors only the
// cumulative count, so the two are not each other's mirror and the tie
// assertion does not pair them.
//
// Its sibling `maxCanvasTextLines = 256` is GONE, not renamed: 256 lines is
// about five pages at 11pt — maxCanvasBodyTextLines' own derivation below
// establishes 48 lines to an A4 page at that size — an order of magnitude
// short of the epic's own
// forty-page target, and D-7.4.2 rejected raising it in place — the cliff
// was the defect, its position was not.
const maxCanvasTextFragments = 512

// maxCanvasBodyText is a CHANNEL-REPRESENTABILITY BACKSTOP, not a paint
// bound, and it is the ONE body-text site that still refuses rather than
// degrades. Degradation lives on the paint side alone (D-7.4.2 §1): a
// component's Value is what the properties panel edits and SAVES, so
// shortening it would write the truncation into the author's document.
//
// Criterion: it must not be able to bind before the paint bounds do. The
// largest document the paint bounds admit is maxCanvasBodyTextLines lines of
// ~90 characters, and the worst case is three bytes per character (Thai/CJK,
// which NFR3 makes first-class): 1920 × 90 × 3 = 518 400 bytes. The next
// power of two above that is 524 288 (512 KiB); this constant is the one
// ABOVE it, and that extra doubling is deliberate rather than arithmetic —
// the ~90-character line is an ESTIMATE of a line's width in characters, not
// a bound on it, so the backstop is set a full power of two clear of the
// estimate. Epic 7's own input cannot reach either figure. Recorded, not
// fixed — following D-7.2.3's precedent for a stated sanity ceiling.
const maxCanvasBodyText = 1048576

// maxCanvasBodyTextLines is Epic 7's own forty-page target, measured:
//
//	40 pages × ⌊729890 mp content-band height ÷ 14982 mp advance⌋ = 40 × 48
//
// where 729890 mp is ContentHeight for a canonical A4 page (841890 mp tall,
// internal/layout/band.go) with 36pt margins and 20pt header and footer, and
// 14982 mp is the measured Advance of the shipped ["Noto Sans"] chain at
// 11pt. (At 12pt the advance is 16344 mp and a page holds 44 lines, so 11pt
// is the admitting figure of the two.) Past this the element paints its
// first N lines and sets Truncated — it never aborts the projection.
const maxCanvasBodyTextLines = 1920

// maxCanvasBodyTextFragments is the CUMULATIVE per-element fragment budget,
// mirroring the browser validator's own cumulative count (Go's per-line
// maxCanvasTextFragments bounds a different quantity, and the Go side must
// not emit what the browser will reject).
//
// Criterion: the same forty-page document, justified at full A4 content
// width, where Story 7.3 makes a justified line project one fragment per
// word-piece.
//
// MEASURED, at the closing revision of Story 7.4 and with the value cap
// lifted, through CanvasWithTextPaint itself: justified English contract
// prose at 11pt in the shipped ["Noto Sans"] chain — the same face and size
// maxCanvasBodyTextLines is derived from — across 523.276 pt of A4 content
// width gives 18.05 fragments per line over 101 lines (1 823 fragments for
// 1 824 words). 1920 × 18.05 = 34 656, and the next power of two above that
// is 65 536.
//
// 65 536 also clears a SHORT-WORD worst case measured the same way — "the cat
// sat on a mat" prose packs 30.86 fragments per line, and 1920 × 30.86 =
// 59 251 — so the forty-page criterion holds for text denser than a
// contract's, not only for the corpus it was measured on. (The earlier 16.72
// figure was a thirteen-line sample, where a justified block's short last
// line still moves the average; the earlier 19.35 was the Roboto-Regular TEST
// face rather than the shipped chain. Both are superseded by the figure
// above, which deferred-work.md and epic-7-8-decision-log.md now also carry.)
//
// The geometry-free law behind it: a justified component's cumulative
// fragment count ≈ the value's WORD COUNT, at any column width.
const maxCanvasBodyTextFragments = 65536

// CanvasTextFragment is a shaped, positioned paint fragment. It is not a
// document text node: x is the engine-owned, band-relative paint origin.
type CanvasTextFragment struct {
	Text string `json:"text"`
	X    int64  `json:"x"`
}

// CanvasTextLine is one pre-broken engine line. All coordinates are
// band-relative, top-left/Y-down millipoints. Advance is retained so the
// browser never derives a following line's origin from CSS metrics.
type CanvasTextLine struct {
	Top       int64                `json:"top"`
	Baseline  int64                `json:"baseline"`
	Advance   int64                `json:"advance"`
	Width     int64                `json:"width"`
	Fragments []CanvasTextFragment `json:"fragments"`
}

// CanvasTextPaint is the closed browser paint plan for one text component.
// It deliberately carries no CSS, browser metric, or document-schema input.
type CanvasTextPaint struct {
	Overflow bool `json:"overflow"`
	// Truncated says this paint is a PREFIX of the element's text: the value
	// is intact in the document and renders whole to PDF, but the projection
	// stopped at a painting bound (D-7.4.2 §2).
	//
	// It exists because without it a degraded element and an EMPTY element
	// are indistinguishable — both used to project `Lines: []`, the all-clear
	// wearing the face of could-not-look. It is a projection disposition, not
	// a document validity rule: no diag.Diagnostic, no registry entry, and
	// the render path has no such cap.
	Truncated bool             `json:"truncated"`
	Lines     []CanvasTextLine `json:"lines"`
}

// SnapToGrid is the reusable core-command seam for Story 5.7 placement.
// It uses the fixed six-point grid and half-away-from-zero rule; callers pass
// millipoints and never browser pixels.
func SnapToGrid(proposed geom.Length) (geom.Length, bool) {
	return proposed.SnapNearest(geom.Length(GridIncrement))
}

type CanvasBand struct {
	Name   string `json:"name"`
	X      int64  `json:"x"`
	Y      int64  `json:"y"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}
type CanvasComponent struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Band      string `json:"band"`
	X         int64  `json:"x"`
	Y         int64  `json:"y"`
	Width     int64  `json:"width"`
	Height    int64  `json:"height"`
	Resizable bool   `json:"resizable"`
	// The following explicitly named optional values are the minimum committed
	// property-panel projection. This is not a generic style or document bag.
	Value *string `json:"value,omitempty"`
	// Binding is a bounded, Go-derived paint label for a direct text binding.
	// It is not a general expression/template projection and cannot be used to
	// reconstruct canonical document bytes in the browser.
	Binding    *string `json:"binding,omitempty"`
	VisibleIf  *string `json:"visibleIf,omitempty"`
	FontFamily *string `json:"fontFamily,omitempty"`
	FontSize   *int64  `json:"fontSize,omitempty"`
	// LineSpacing is style.lineSpacing in THOUSANDTHS, the unit the format
	// and the property command both carry it in (template.LineSpacingUnit).
	// It is dimensionless — a ratio applied to the vertical model's Advance —
	// so it is not a geom.Length and is never treated as one.
	LineSpacing   *int64            `json:"lineSpacing,omitempty"`
	Bold          *bool             `json:"bold,omitempty"`
	Italic        *bool             `json:"italic,omitempty"`
	Align         *string           `json:"align,omitempty"`
	Valign        *string           `json:"valign,omitempty"`
	Background    *string           `json:"background,omitempty"`
	Color         *string           `json:"color,omitempty"`
	BorderWidth   *int64            `json:"borderWidth,omitempty"`
	BorderColor   *string           `json:"borderColor,omitempty"`
	BorderEdges   []string          `json:"borderEdges,omitempty"`
	TableBind     *string           `json:"tableBind,omitempty"`
	PaddingTop    *int64            `json:"paddingTop,omitempty"`
	PaddingRight  *int64            `json:"paddingRight,omitempty"`
	PaddingBottom *int64            `json:"paddingBottom,omitempty"`
	PaddingLeft   *int64            `json:"paddingLeft,omitempty"`
	TextPaint     *CanvasTextPaint  `json:"textPaint,omitempty"`
	Image         *CanvasImagePaint `json:"image,omitempty"`
	// ImageUnavailable is a small, bounded discriminant set ONLY when this
	// is an image element and Image is absent (Finding 9, review of
	// 2026-08-29): "missing" when the element's own asset key is not in
	// the document's assets map, or "undecodable" when the key resolves
	// but the bytes fail to decode or the media type is one this library
	// version cannot render. D-5.13.2's "one Go-side signal drives both"
	// governed the media-type case only; collapsing a dangling asset
	// reference into that same undecodable text was a defect — the media
	// type there is fine, the asset is simply gone. This does not widen
	// the projection's authority: it is still Go stating which of two
	// bounded, enumerated reasons applies, never bytes or a path.
	ImageUnavailable *string `json:"imageUnavailable,omitempty"`
}

// imageUnavailableMissing / imageUnavailableUndecodable are
// ImageUnavailable's only two values (Finding 9). Kept as named constants,
// not inline literals, so the Go producer and any future consumer cannot
// drift on spelling.
const (
	imageUnavailableMissing     = "missing"
	imageUnavailableUndecodable = "undecodable"
)

// CanvasImagePaint is Story 5.13's read-only, paint-only projection of one
// placed image element's Go-owned display data: the declared media type,
// the asset's content-addressed key, VALIDATED intrinsic pixel dimensions,
// and the fit-and-centre draw rectangle already computed for the PDF
// (resolveImagePlacement), in BAND-RELATIVE millipoints (matching this
// component's own X/Y, D-5.13.2's "Frame" clause). It carries no asset
// BYTES (AD-17: a paint-only projection must not carry anything that
// reconstructs canonical bytes or the assets map): AssetKey is only a
// LOOKUP TOKEN for the separate, explicit, per-key bytes request
// (AssetBytes/wasm.Engine.AssetBytes) the canvas uses to obtain what it
// paints — a key alone cannot reconstruct the assets map or canonical
// bytes any more than a table id (already sent on every projection) can.
// The inspector abbreviates this same key for DISPLAY (a formatting choice
// over a value Go already supplied, same as it formats millipoints as
// "12.5pt"); Go does not truncate it on the wire, because the canvas needs
// the real key to ask for bytes.
//
// The whole field is present only when the referenced asset decodes
// successfully through the recognised-image path — D-5.13.2's "Absence,
// not zero": DecodedImage.Width()/Height() are reachable only through
// decodeRecognisedImage, so a legally-loaded asset of an unrecognised media
// type (or one whose bytes fail to decode) has no dimensions and no
// computable rectangle. Rather than carry two independently-absent signals
// (a known media type but a missing rectangle), the ENTIRE paint is absent
// together — ONE Go-side signal drives both AC2's inspector failure text
// and AC3's canvas placeholder, never two.
type CanvasImagePaint struct {
	MediaType  string `json:"mediaType"`
	AssetKey   string `json:"assetKey"`
	Width      int64  `json:"width"`
	Height     int64  `json:"height"`
	DrawX      int64  `json:"drawX"`
	DrawY      int64  `json:"drawY"`
	DrawWidth  int64  `json:"drawWidth"`
	DrawHeight int64  `json:"drawHeight"`
}

const maxCanvasBindingString = 256

type CanvasProjection struct {
	Width         int64             `json:"width"`
	Height        int64             `json:"height"`
	Orientation   string            `json:"orientation"`
	Preset        string            `json:"preset"`
	MarginTop     int64             `json:"marginTop"`
	MarginRight   int64             `json:"marginRight"`
	MarginBottom  int64             `json:"marginBottom"`
	MarginLeft    int64             `json:"marginLeft"`
	GridIncrement int64             `json:"gridIncrement"`
	CommandWidth  int64             `json:"commandWidth"`
	CommandHeight int64             `json:"commandHeight"`
	Bands         []CanvasBand      `json:"bands"`
	Components    []CanvasComponent `json:"components"`
	// FontFamilies is the closed set style.fontFamily may name in THIS
	// document: the declared, non-empty font chains, by name, sorted so the
	// projection is deterministic. It exists so the designer can offer the
	// author exactly the families the engine will accept (knownFontFamily),
	// instead of a free text field whose every rejection is a round trip. It
	// is still names only — the faces live in FontChains below, and the font
	// BYTES are projected by nothing.
	FontFamilies []string `json:"fontFamilies"`
	// FontChains is the same set, WITH the ordered faces behind each name —
	// entry for entry, in the document's own authored order. It is exactly the
	// chains FontFamilies names, in the same positions, so the two can never
	// disagree about which chains a document declares; it exists so a chain
	// editor re-projects the engine's answer instead of modelling the fonts
	// map a second time in the browser.
	FontChains []CanvasFontChain `json:"fontChains"`
	// DefaultFontSize is the size the producer draws a text element at when
	// its style carries no fontSize, in millipoints. It is projected rather
	// than restated in the browser for the ordinary reason: it is the
	// engine's number, and a second copy of it in the designer would be a
	// second authority on what an unset size means.
	DefaultFontSize int64 `json:"defaultFontSize"`
	// ContentWindowHeight is ONE page's worth of content column, in
	// millipoints: internal/layout's ContentHeight, which is the single
	// function permitted to derive it (AD-13). It is the same number
	// bands[1].height carries — the content band rectangle IS one window —
	// and it is named separately because the two stop being interchangeable
	// the moment the designer draws a second sheet: the band is where page
	// one's content sits, the window is the distance between sheets.
	//
	// It is projected rather than recomputed in the browser because a second
	// spelling of it would be a second authority on where a page ends, and
	// the divergence would be invisible: the canvas and the engine would draw
	// different pages and still agree on the bytes.
	ContentWindowHeight int64 `json:"contentWindowHeight"`
	// ContentWindowCount is how many of those windows the content column
	// occupies, from internal/layout's Paginate — the ONE function that
	// decides how many pages a column has. It is never `ceil(lowestBottom /
	// ContentWindowHeight)`, a spelling paginate.go forbids by name: the
	// window advances to the first item that did not fit, so an element
	// declared ten windows below the text starts the NEXT window rather than
	// generating nine empty ones.
	//
	// WHAT THIS NUMBER IS A NUMBER ABOUT. It describes the column AS THE
	// CANVAS CURRENTLY PAINTS IT, not the document that will render. The
	// canvas has no data, so a bound table contributes only its header
	// height (projectedSize) and every row it will grow is absent: for any
	// document with a bound table in the content band this count is a FLOOR,
	// never a prediction. The finished document may run longer; it can never
	// run shorter.
	//
	// It is never derived from CanvasTextPaint. The paint truncates — at a
	// line budget and at a fragment budget — and a count that read a paint's
	// line list would shorten with it, so the canvas would draw the wrong
	// number of sheets for exactly the documents long enough to need them.
	// The extents fed to Paginate come from the FULL shaped line list and the
	// vertical model, which truncation never touches.
	//
	// Always at least 1: a column with nothing in it is one page, not zero.
	ContentWindowCount int64 `json:"contentWindowCount"`
	// ContentWindowOrigins is where each of those windows BEGINS, in the
	// content column's own band-relative frame — the same frame
	// CanvasComponent.Y is already in for a content component. origins[0] is
	// 0, because window one starts at the top of the column and internal/
	// layout guarantees that unconditionally; every later entry is the
	// column offset the engine slid that window to. There is exactly one
	// entry per window, so len(ContentWindowOrigins) == ContentWindowCount.
	//
	// They come from internal/layout's own PageAssignment.Shift — the value
	// Paginate had already computed while deciding the count — and are NEVER
	// `index * ContentWindowHeight`. That closed form is the spelling
	// paginate.go forbids by name for the count, and origins expose it more
	// sharply than the count does: the window advances to the TOP OF THE
	// FIRST ITEM THAT DID NOT FIT, never by a fixed height, so three
	// elements a round 728pt apart begin at 0, 728000 and 1456000 where the
	// closed form answers 0, 727890 and 1455780 — adrift by 110 millipoints
	// per window — and a column with a declared ten-window gap begins two
	// windows where the closed form answers eleven.
	//
	// The window HEIGHT does not vary: window i spans
	// [origins[i], origins[i]+ContentWindowHeight). Only the tops slide.
	//
	// It is ALWAYS non-empty. A nil slice marshals to JSON null, the browser
	// protocol rejects it, and rejecting one field discards the whole
	// snapshot — which blanks the canvas with nothing to attribute the blank
	// to.
	ContentWindowOrigins []int64 `json:"contentWindowOrigins"`
	// ContentWindowCountIsExact states, as a value rather than only in the
	// comment above, whether ContentWindowCount can be TRUSTED as the number
	// of pages this content column occupies. The ENGINE reports it, because
	// only the engine knows every cause; the designer states the consequence
	// in words and never decides for itself that, say, a table means more
	// pages — that would be a second authority on a question this flag
	// answers exactly.
	//
	// ITS ZERO VALUE IS THE SAFE CLAIM, and that is why it is spelled this
	// way round rather than as `…IsApproximate`. `false` reads "do not trust
	// this count", so a projection path that forgets to set it degrades to
	// the HONEST claim. The inverse field would have had a forgotten set
	// CLAIM EXACTNESS — which is precisely the defect that produced this
	// field's rename, rebuilt into its default. A hazard indicator must not
	// fail toward the quiet variant.
	//
	// It is false when any of these is so:
	//
	//	(a) a content-band table carries a non-empty binding, so the column
	//	    being counted holds that table's header and none of the rows its
	//	    data will grow;
	//	(b) Paginate could not place the column at all — a component taller
	//	    than one window — and the count degraded to the documented one,
	//	    or the pagination produced an origin sequence the browser
	//	    protocol would refuse;
	//	(c) a content-band text element contributed no extents because its
	//	    font chain would not resolve, so its lines are absent from the
	//	    column the count measures;
	//	(d) a content-band element's VISIBILITY DEPENDS ON DATA — it carries
	//	    a visibleIf, which this file only projects as a string and which
	//	    nothing on the canvas path evaluates, because evaluating it needs
	//	    the data the canvas has never been given. The canvas places the
	//	    element and the render may omit it, and AD-24 makes a hidden
	//	    element absent WITH NO GAP, so the column is simply shorter.
	//	    UNDISCLOSED SINCE STORY 7.5 shipped the count: it applies to an
	//	    UNGROUPED visibleIf element exactly as much as to a grouped one.
	//	    Story 7.9's grouping work is how it was found, not what caused it.
	//
	// GROUPING IS NOT AMONG THEM, and never becomes one. keepTogetherTags
	// takes the *Template and nothing else, so an author-declared
	// keep-together group is a pure template property the canvas holds every
	// input for: being wrong about it is a defect to fix, never a shortfall
	// to disclose. parse_bands.go's refusal of keepTogether on a table is
	// what keeps that true, by stopping a group inheriting (a)'s data
	// dependency.
	//
	// DIRECTION WAS DELIBERATELY DROPPED, and this sentence is here because
	// without it a future reader restores the floor claim mistaking a choice
	// for lost fidelity. The causes do not agree on a direction — (a) and (c)
	// make the canvas count too LOW (a floor), while (d) makes it too HIGH (a
	// ceiling), and a document carrying both is wrong in either direction —
	// so no single direction is honest for the general case, and the field
	// this replaced was named `ContentWindowCountIsFloor` and set true on
	// ceiling causes. Direction also informs no decision: a floor means there
	// may be more sheets than drawn and a ceiling fewer, and neither is a
	// safe side to act on. It belongs WITH THE CAUSES — a cause knows its own
	// direction — so if this projection ever carries the cause set, direction
	// can be derived there without this flag re-acquiring a claim. The
	// projection carries only the boolean today.
	ContentWindowCountIsExact bool `json:"contentWindowCountIsExact"`
}

// maxCanvasFontFamilies bounds the projected name list the way every other
// list in this projection is bounded. A document declaring more chains than
// this is refused a projection with a stated reason, never silently cut.
const maxCanvasFontFamilies = 256

// CanvasFontChain is one declared font chain AS THE DESIGNER SEES IT: the name
// style.fontFamily may carry, and the ordered face names behind it. Story 8.1
// adds the second half. FontFamilies' own doc comment used to say the
// projection was "names only — never the chains", and that was exactly the
// limitation a chain editor could not be built on: a moved or removed entry
// changes nothing the browser can observe, so the panel would have to model
// the fonts map itself rather than re-project it.
type CanvasFontChain struct {
	Name string `json:"name"`
	// Entries carries the ordered entries, and since Story 8.3 each is
	// an OBJECT rather than a string: an entry may name a face the
	// renderer is given or a face the document itself carries, and the
	// browser must be able to tell which without inspecting the value.
	Entries []CanvasFontChainEntry `json:"entries"`
}

// CanvasFontChainEntry is one chain entry AS THE DESIGNER SEES IT.
//
// THE SHAPE IS DISCRIMINATED, AND THE DISCRIMINANT IS PROJECTED RATHER
// THAN INFERRED. Exactly one of Face and AssetKey is non-empty. The
// browser is forbidden from deriving which kind an entry is — no key
// detection, no parsing, no length heuristic on a 64-character string —
// so the engine states it, and the designer's guard asserts it.
//
// Family and Style are EMPTY for a named face (its name is the whole
// identity the document gives it) and non-empty for an embedded one.
// They come from the asset's own `font` record, read HERE — the browser
// may display what this projection carries and derive nothing from it,
// which is why the family falls back below rather than being left for
// the panel to patch up.
//
// WHAT MECHANICALLY ENFORCES THAT, STATED NARROWLY. Nothing tests "the
// panel holds no rule" as such, and claiming otherwise was this
// comment's own defect (review finding 5).
// canvas-authority-contract.test.ts walks every production source file
// under folio-designer/src — FontChainEditor.tsx among them, by
// directory walk rather than by name — and fails if any of them restates
// the ENGINE'S REFUSAL VOCABULARY. That is one rule, not all of them.
// The rest of this paragraph is an engineering rule the reviewer of a
// browser change enforces, and the Go-side half of the contract — that
// the engine really emits the shape the browser's guard requires — is
// pinned by canvas_font_chain_entry_test.go.
//
// Family is NEVER EMPTY for an embedded entry. When the asset declares
// no `font.family`, the ASSET KEY is projected as the family — the
// engine chooses what the panel shows, so the browser never has to
// decide what to do with an empty name. Showing a 64-character digest is
// the honest answer for a document that named its own face nothing;
// inventing a name here would be the engine guessing.
//
// All four keys are ALWAYS emitted (no omitempty, deliberately). The
// browser checks this object with an exact-key guard, so a key that
// appears only for some entries is a key that rejects the whole snapshot
// for some documents — and the symptom is a blank canvas.
type CanvasFontChainEntry struct {
	Face     string `json:"face"`
	AssetKey string `json:"assetKey"`
	Family   string `json:"family"`
	Style    string `json:"style"`
}

// canvasFontChains is the projection of the document's declared chains, in
// sorted key order: every chain template.Fonts.Chain accepts — declared AND
// non-empty — and no other. It ASKS Fonts.Chain that question rather than
// re-implementing it, which is what makes the sentence above a description of
// the code instead of a second copy of it. The comment this replaced named a
// different function as the authority ("exactly the names knownFontFamily
// accepts") while spelling the test out again three lines later; naming a
// caller rather than the rule is exactly how that drift started.
func canvasFontChains(t *Template) ([]CanvasFontChain, error) {
	chains := make([]CanvasFontChain, 0, len(t.doc.Fonts))
	// slices.Sorted(maps.Keys(...)) is the module's one way to walk a map:
	// map order is not an order, and this list is projected output.
	for _, name := range slices.Sorted(maps.Keys(t.doc.Fonts)) {
		entries, ok := t.doc.Fonts.Chain(name)
		if !ok {
			continue
		}
		if len(name) > maxCanvasPropertyString {
			return nil, fmt.Errorf("folio: font family name exceeds the projection bound")
		}
		if len(entries) > maxCanvasFontChainEntries {
			return nil, fmt.Errorf("folio: font chain declares more entries than the projection bound")
		}
		projected := make([]CanvasFontChainEntry, 0, len(entries))
		for _, entry := range entries {
			p, perr := projectFontChainEntry(t, entry)
			if perr != nil {
				return nil, perr
			}
			projected = append(projected, p)
		}
		chains = append(chains, CanvasFontChain{Name: name, Entries: projected})
	}
	if len(chains) > maxCanvasFontFamilies {
		return nil, fmt.Errorf("folio: document declares more font families than the projection bound")
	}
	return chains, nil
}

// projectFontChainEntry projects ONE entry, and applies
// maxCanvasPropertyString to EVERY string it puts on the wire — the face
// name, the asset key, the family and the style alike. A bound applied
// to three of four fields is a bound on nothing: the projection is
// refused with a stated reason rather than silently cut, which is the
// rule every other list in this projection already follows.
//
// The family and style are read from the asset's `font` record. An
// explicit `null` there is treated as absence for DISPLAY purposes —
// the file keeps the distinction (Presence round-trips it), but a panel
// has nothing to draw for a null, and it is not the browser's job to
// decide that.
func projectFontChainEntry(t *Template, entry template.FontChainEntry) (CanvasFontChainEntry, error) {
	var out CanvasFontChainEntry
	if entry.Embedded() {
		out.AssetKey = entry.AssetKey
		out.Family = entry.AssetKey
		if asset, ok := t.doc.Assets[entry.AssetKey]; ok && asset.Font.Set && !asset.Font.Null {
			record := asset.Font.Value
			if record.Family.Set && !record.Family.Null && record.Family.Value != "" {
				out.Family = record.Family.Value
			}
			if record.Style.Set && !record.Style.Null {
				out.Style = record.Style.Value
			}
		}
	} else {
		out.Face = entry.Face
	}
	for _, s := range []string{out.Face, out.AssetKey, out.Family, out.Style} {
		if len(s) > maxCanvasPropertyString {
			return CanvasFontChainEntry{}, fmt.Errorf("folio: font chain entry exceeds the projection bound")
		}
	}
	return out, nil
}

// canvasFontFamilyNames is FontFamilies, derived from FontChains rather than
// walked a second time: FontChains[i].Name == FontFamilies[i] then holds BY
// CONSTRUCTION, which is what lets the browser cross-check the two lists
// against each other and lets the single Fonts.Chain authority govern both.
func canvasFontFamilyNames(chains []CanvasFontChain) []string {
	names := make([]string, 0, len(chains))
	for _, chain := range chains {
		names = append(names, chain.Name)
	}
	return names
}

// canvasPageGeometry is THE one layout.PageGeometry the canvas builds, and
// every canvas consumer of a page-geometry quantity reads it: the content
// band rectangle, the projected window height and the window count all come
// from this single struct, so they cannot diverge from one another.
//
// It is deliberately NOT render.go's pageGeometryOf. That one routes through
// pageDimensions, which hard-errors on "Letter" by design — failing loudly is
// more honest than a silent A4 substitution when a PDF is about to be
// produced. canvasDimensions supports Letter, and a Letter document projects
// a canvas today, so routing the canvas through the render path's spelling
// would break a projection that works.
func canvasPageGeometry(t *Template) (layout.PageGeometry, error) {
	w, h, err := canvasDimensions(t)
	if err != nil {
		return layout.PageGeometry{}, err
	}
	m := t.doc.Page.Margin
	return layout.PageGeometry{
		Width:            w,
		Height:           h,
		MarginTop:        m.Top,
		MarginBottom:     m.Bottom,
		MarginLeft:       m.Left,
		MarginRight:      m.Right,
		PageHeaderHeight: t.doc.Bands.PageHeader.Height.Value,
		PageFooterHeight: t.doc.Bands.PageFooter.Height.Value,
	}, nil
}

// Canvas returns immutable paint geometry. It intentionally exposes neither
// template fields nor elements, canonical bytes, or browser measurements.
//
// ContentWindowCount is a documented ONE window here, declared NOT EXACT.
// Counting windows needs shaped lines, which needs a FontSet this entry point
// does not receive — and it does not need to: every projection that reaches
// the browser is a CanvasWithTextPaint (wasm/engine.go's three seams),
// because every mutating command's own Canvas(t) is discarded and recomputed
// there with fonts. One window is what a column with nothing placeable in it
// occupies anyway; a silent zero would be a page count no document has. The
// number is not a claim of any kind — neither a floor nor a ceiling — which
// is exactly what the flag below says.
//
// ContentWindowOrigins and ContentWindowCountIsExact are the SAME admission,
// spelled in the two fields that carry it: one window beginning at column
// offset zero, declared NOT EXACT. ⚠ THE SENSE OF THAT FIELD IS INVERTED
// FROM THE ONE IT REPLACED, and this literal is the site where a mechanical
// rename would have converted a documented shortfall into a claim of
// exactness — the flag reads `false` here for the same reason it used to
// read `true`. The struct is shared with the entry point that can shape, so
// these values never reach the browser — but a shared struct's values must
// be honest wherever they are set, and a `nil` origins slice would marshal to
// a JSON null the protocol rejects.
func Canvas(t *Template) (CanvasProjection, error) {
	if t == nil {
		return CanvasProjection{}, errNilTemplate
	}
	g, err := canvasPageGeometry(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	w, h := g.Width, g.Height
	m := t.doc.Page.Margin
	header, footer := g.PageHeaderHeight, g.PageFooterHeight
	if w <= 0 || h <= 0 || m.Left < 0 || m.Right < 0 || m.Top < 0 || m.Bottom < 0 || m.Left >= w-m.Right || m.Top >= h-m.Bottom {
		return CanvasProjection{}, fmt.Errorf("folio: page setup leaves no positive content region")
	}
	for _, v := range []geom.Length{w, h, m.Top, m.Right, m.Bottom, m.Left, header, footer} {
		if v < 0 || v > geom.Length(MaxCanvasMillipoints) {
			return CanvasProjection{}, fmt.Errorf("folio: page setup exceeds the JavaScript-safe geometry bound")
		}
	}
	innerW, innerH := w-m.Left-m.Right, h-m.Top-m.Bottom
	if header < 0 || footer < 0 || header >= innerH-footer {
		return CanvasProjection{}, fmt.Errorf("folio: page setup leaves no positive content region")
	}
	preset := "custom"
	if t.doc.Page.SizeIsName {
		preset = t.doc.Page.SizeName
	}
	commandW, commandH := w, h
	if !t.doc.Page.SizeIsName {
		commandW, commandH = t.doc.Page.SizeCustom.Width, t.doc.Page.SizeCustom.Height
	}
	// AD-13: the content band's height is derived by ONE function, in
	// internal/layout. The inline `innerH - header - footer` that used to
	// stand here was arithmetically identical and still a second spelling of
	// a derived quantity — and the projection now REPORTS this number as the
	// page-height window, which is what the designer draws sheet boundaries
	// from, so a divergence would show up as the canvas and the engine
	// drawing different pages while agreeing on every byte.
	window := layout.ContentHeight(g)
	bands := []CanvasBand{
		{Name: bandPageHeader, X: int64(m.Left), Y: int64(m.Top), Width: int64(innerW), Height: int64(header)},
		{Name: bandContent, X: int64(m.Left), Y: int64(m.Top + header), Width: int64(innerW), Height: int64(window)},
		{Name: bandPageFooter, X: int64(m.Left), Y: int64(h - m.Bottom - footer), Width: int64(innerW), Height: int64(footer)},
	}
	components, err := canvasComponents(t, bands)
	if err != nil {
		return CanvasProjection{}, err
	}
	chains, err := canvasFontChains(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	return CanvasProjection{Width: int64(w), Height: int64(h), Orientation: t.doc.Page.Orientation, Preset: preset, MarginTop: int64(m.Top), MarginRight: int64(m.Right), MarginBottom: int64(m.Bottom), MarginLeft: int64(m.Left), GridIncrement: GridIncrement, CommandWidth: int64(commandW), CommandHeight: int64(commandH), Bands: bands, Components: components, FontFamilies: canvasFontFamilyNames(chains), FontChains: chains, DefaultFontSize: int64(defaultFontSizePt), ContentWindowHeight: int64(window), ContentWindowCount: 1, ContentWindowOrigins: []int64{0}, ContentWindowCountIsExact: false}, nil
}

// CanvasWithTextPaint returns Canvas geometry augmented with a read-only,
// production-parity text paint plan. It is session output only: it never
// mutates the template or its canonical serialization.
func CanvasWithTextPaint(t *Template, fs FontSet) (CanvasProjection, error) {
	projection, err := Canvas(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	// One shaping, two consumers. addCanvasTextPaint shapes the content
	// band's text once; the paint plan is one consumer of the extents that
	// shaping produced and the window count is the other. A second shaping
	// pass would be a second derivation of the same numbers, which is the
	// thing internal/layout's ColumnItem doc forbids.
	column := canvasColumnExtents{Items: make([]layout.ColumnItem, 0)}
	if err := addCanvasTextPaint(t, &projection, fs, &column); err != nil {
		return CanvasProjection{}, err
	}
	if err := addCanvasImagePaint(t, &projection); err != nil {
		return CanvasProjection{}, err
	}
	if err := addCanvasWindowCount(t, &projection, column); err != nil {
		return CanvasProjection{}, err
	}
	return projection, nil
}

// canvasColumnExtents is what addCanvasTextPaint hands addCanvasWindowCount:
// the content column's per-line extents, and the ONE thing about them the
// count cannot see for itself.
//
// Items is the same slice this function used to pass on its own. FontChain-
// Degraded is the addition: a content-band text element whose font chain
// would not resolve is skipped with an empty paint a few lines into
// addCanvasTextPaint, and the extents it would have contributed are simply
// absent from Items. The count that measures Items is therefore a FLOOR for
// that document, and nothing downstream of Items could tell — the missing
// lines look exactly like an element that had nothing to say. Carrying the
// fact beside the extents is what keeps the flag an ENGINE fact rather than
// a browser rule about what an empty paint might mean.
type canvasColumnExtents struct {
	Items             []layout.ColumnItem
	FontChainDegraded bool
}

// canvasWindowOrigins reads the window tops out of the very Pagination the
// count is taken from: PageAssignment.Shift IS window i's band-relative
// column offset, already computed by the one function permitted to decide
// where a page begins. Nothing is derived, and in particular nothing is
// multiplied — `index * ContentWindowHeight` is the closed form
// internal/layout/paginate.go forbids by name, and it is wrong here by 110
// millipoints per window on a column of round 728pt spacing and by nine whole
// windows on a column with a declared gap.
//
// It STATES the three properties the browser protocol independently requires
// — a non-empty sequence, a first origin of zero, strictly increasing — and
// refuses rather than returning a sequence that would fail them. A refused
// sequence degrades exactly as a refused pagination does; a sequence the
// protocol rejects would discard the whole snapshot and blank the canvas with
// nothing to attribute the blank to.
func canvasWindowOrigins(plan layout.Pagination) ([]int64, bool) {
	if len(plan.Pages) == 0 {
		return nil, false
	}
	origins := make([]int64, 0, len(plan.Pages))
	for i, page := range plan.Pages {
		shift := page.Shift
		if shift < 0 || shift > geom.Length(MaxCanvasMillipoints) {
			return nil, false
		}
		if i == 0 && shift != 0 {
			return nil, false
		}
		if i > 0 && int64(shift) <= origins[i-1] {
			return nil, false
		}
		origins = append(origins, int64(shift))
	}
	return origins, true
}

// canvasContentBandHasBoundTable is cause (a) of ContentWindowCountIsExact's
// register, and the one that reads as a FLOOR: a table in the content band
// with a non-empty binding. projectedSize gives
// such a table its header height and not one row, because the canvas has
// never been given the data — so the column being counted is one header tall
// however many hundred rows the finished document runs to.
func canvasContentBandHasBoundTable(t *Template) bool {
	for _, element := range t.doc.Bands.Content.Elements {
		if element.Type == template.ElementTable && element.Table.Set && !element.Table.Null && element.Table.Value.Bind != "" {
			return true
		}
	}
	return false
}

// canvasElementIsPlaced answers, for a NON-TEXT content element, whether the
// render path would contribute a content-column item for it — the question
// addCanvasWindowCount's own arm must answer the same way, or the canvas
// counts a column the document does not have.
//
// Three routes reach the render path's contentColumnItems, and the kinds
// divide by which ones they can take (page_number.go). NO KIND IS PLACED
// UNCONDITIONALLY — each route has a condition, every one of them is a pure
// property of the template, and this predicate is the conjunction the canvas
// can therefore evaluate with no data at all:
//
//	table — its header rect, from collectBandTableRuns, and never through
//	        element_box.go, which excludes a table by name. Placed while it
//	        declares at least one COLUMN: `"columns": []` parses, and a
//	        table with no columns has nothing to lay out or draw.
//	image — its image run, from collectImageRuns, which is placed while the
//	        element's `asset` is present and non-null; createComponent gives
//	        every newly dropped image a NULL asset, so an unfilled box is
//	        the designer's ordinary state and not an exotic one. An image
//	        may ALSO declare a box, and a styled one reaches a column item
//	        that way even with no file chosen — so the two routes are OR-ed
//	        rather than ranked.
//	rect,
//	line  — element_box.go's rect source, and NOTHING ELSE. So these two are
//	        placed exactly where they declare a box.
//
// Each condition is the render path's own predicate, called rather than
// restated: one authority per route, two callers each, the same shape as
// keepTogetherTags. elementDeclaresBox carries BOTH halves of the box rule —
// the style declaration and the positive rectangle — so a styled element of
// zero height is placed by neither side.
//
// WHAT THIS DOES NOT ANSWER, deliberately: whether the element is VISIBLE.
// The render path consults its visibility verdicts before asking any of
// these questions, and the canvas has no data to resolve a visibleIf with —
// which is why conditional visibility is a registered cause of inexactness
// rather than a clause here.
func canvasElementIsPlaced(element template.Element) bool {
	switch element.Type {
	case template.ElementTable:
		return tableDrawsColumns(element)
	case template.ElementImage:
		return imageDrawsItsAsset(element) || elementDeclaresBox(element)
	default:
		return elementDeclaresBox(element)
	}
}

// canvasContentBandHasConditionalVisibility is the cause named for what it
// IS — an element whose visibility depends on DATA — rather than for the
// story that found it.
//
// page_setup.go only PROJECTS VisibleIf, as a string; nothing on this path
// evaluates it, because evaluating it needs the data the canvas has never
// been given. So the canvas places the element and the render may omit it,
// and AD-24 makes a hidden element ABSENT WITH NO GAP — no sibling moves up
// into the space, the column is simply shorter, and the two counts differ.
//
// UNDISCLOSED SINCE STORY 7.5, which shipped the count, and 7.6, which
// shipped the flag. It applies to an UNGROUPED element carrying visibleIf
// exactly as much as to a grouped one: grouping did not create this cause, it
// is only how it was found, because a conditional member makes a group's
// whole slide conditional and that is loud enough to measure.
//
// Content band only, like every other cause here: this flag is a claim about
// the content column, and a conditional element in a repeated band changes
// what that band draws, never how many windows the column occupies.
//
// PRESENT AND NON-NULL IS THE WHOLE TEST. It carried a third clause,
// `Value != ""`, which was unreachable — the loader refuses an empty
// expression outright ("visibleIf \"\" is not a valid expression: empty
// expression") — and unreachable in the unsafe direction: had a document
// ever carried one, that clause would have made this flag claim exactness
// for an element whose placement the canvas cannot decide. A hazard test
// does not need an escape hatch for a document the loader cannot produce.
func canvasContentBandHasConditionalVisibility(t *Template) bool {
	for _, element := range t.doc.Bands.Content.Elements {
		if element.VisibleIf.Set && !element.VisibleIf.Null {
			return true
		}
	}
	return false
}

// addCanvasWindowCount is the THIRD paint producer, beside addCanvasTextPaint
// and addCanvasImagePaint: it reports how many page-height windows the
// content column occupies, WHERE EACH ONE BEGINS, and whether that number can
// be trusted as the printed document's page count.
//
// It calls internal/layout's Paginate — the one function that decides how
// many pages a column has — and never a second pagination of its own. What it
// supplies is only the extents: the per-line tops and advances the paint plan
// already computed (textItems), plus one item per non-text content component
// from the box canvasComponents already projects. Paginate's signature takes
// PageGeometry and ColumnItems, no data, no bindings and no template, because
// receiving caller-derived extents is exactly what it is for.
//
// A NON-TEXT COMPONENT CONTRIBUTES EXACTLY WHERE THE RENDER PATH PLACES ONE
// — canvasElementIsPlaced above is that whole rule, calling each route's own
// predicate rather than restating it — and the sentence here used to say the
// opposite ("every content component contributes, styled or not"). It was
// wrong, and Story 7.9 is where it started to matter: an unstyled rect
// reaches no column item on the render path, so a canvas that counted it drew
// a window the printed document does not have. That was invisible while the
// two answers were both ungrouped and merely differed on origins; the moment
// a declared group made the canvas's partition matter, it became a count that
// was confidently wrong.
//
// ⚠ THE SENTENCE IS QUALIFIED FOR A REASON, AND THE QUALIFIER IS TEXT. It
// holds for the kinds this arm carries and NOT in the other direction for
// text, which is why "non-text" is not decoration. Text contributes one item
// per SHAPED LINE and never its box, so a text element that shapes no lines
// at all — an unset, null or empty value, or a font chain that cannot be
// resolved — contributes nothing here, exactly as the render path treats a
// value that binds to empty. But collectElementBoxRects accepts TEXT as well
// (element_box.go's four eligible kinds are text, image, rect and line), so a
// content-band text element declaring a background or a border ALSO
// contributes a full-declared-box column item on the render path, and the
// canvas contributes only its shaped lines. A short value in a tall styled
// box therefore occupies more column on the render path than on the canvas.
// Measured: a styled text element at y 700 with a declared height of 200 and
// a one-line value gives a canvas count of 1 against a real render of 2.
//
// That divergence is LEFT AS IT IS, deliberately and out of this arm's
// subject: closing it means adding a second canvas item source for a text
// element — its box, beside its lines — which is new placement rather than
// the removal of placement this arm performs, and it is not what RULING B
// asked for. It is recorded here so the invariant above is never read wider
// than it holds.
func addCanvasWindowCount(t *Template, projection *CanvasProjection, column canvasColumnExtents) error {
	g, err := canvasPageGeometry(t)
	if err != nil {
		return err
	}
	// THREE of the flag's causes are known before Paginate runs — a bound
	// table, a degraded font chain, and an element whose visibility depends
	// on data; the fourth is the degradation branch below. They are OR-ed
	// rather than ranked because the flag reports that the count cannot be
	// trusted, not which cause made it so. ⚠ The SENSE is inverted from the
	// field this replaced: `exact` is true only when NONE of them applies.
	exact := !(column.FontChainDegraded ||
		canvasContentBandHasBoundTable(t) ||
		canvasContentBandHasConditionalVisibility(t))
	// Story 7.9 (FR51): the same index addCanvasTextPaint tagged its line
	// items with, from the same one authority. Grouping is a pure property
	// of the Template — keepTogetherTags takes nothing else — so the canvas
	// already holds every input it needs to be RIGHT about it, and being
	// wrong about it is a defect rather than a cause to register beside the
	// four above.
	keepTogether := keepTogetherTags(t)
	items := make([]layout.ColumnItem, 0, len(column.Items)+len(t.doc.Bands.Content.Elements))
	items = append(items, column.Items...)
	for _, element := range t.doc.Bands.Content.Elements {
		if element.Type == template.ElementText {
			// Text contributes one item PER SHAPED LINE, never its box: a
			// paragraph splits between windows at a line, which is what
			// makes the count a slide rather than a division.
			continue
		}
		if !canvasElementIsPlaced(element) {
			continue
		}
		_, height := projectedSize(element)
		items = append(items, layout.ColumnItem{
			ElementID: string(element.ID),
			Top:       element.Y,
			Bottom:    element.Y + height,
			// The dummy-ref idiom page_number.go already sanctions:
			// Paginate's exclusivity pre-pass requires exactly one of
			// Runs/Images/Rects to be non-empty, and this Pagination is
			// discarded except for len(Pages), so the value is never read
			// back.
			Rects: []layout.RectRef{0},
			// Story 7.9 (FR51): a rect, line, image or table box joins its
			// author-declared group here, so a signature's ruled line rides
			// with the name above it in the canvas's column exactly as it
			// does in the render's. This arm carries every non-text content
			// kind, so tagging only the text arm would group the column in
			// halves and the count would still diverge.
			Group: keepTogether.keepTogetherGroup(string(element.ID)),
		})
	}
	// ONE translation, in one place. Every extent above is band-relative,
	// exactly as the author declared it and as CanvasComponent carries it;
	// Paginate reads the printable frame, whose content origin is the page
	// header's height. MarginTop is deliberately NOT added — Origins measures
	// downward from the printable top edge, inside the margin, while
	// CanvasBand.Y is paper-absolute.
	origin := layout.Origins(g).Content
	for i := range items {
		items[i].Top += origin
		items[i].Bottom += origin
	}
	plan, err := layout.Paginate(g, items)
	if err != nil {
		// A pagination failure DEGRADES THE COUNT; it never fails the
		// projection. The reachable case is a content component taller than
		// one window, which this story newly makes authorable — turning the
		// render path's overflow into a canvas refusal would make a
		// canvas bound into a document validity rule. One window is
		// Paginate's own answer for a column it cannot place, and it is the
		// same shape as this file's other degradations: dispose of the
		// number, keep the canvas.
		//
		// The origins degrade with the count they describe — one window
		// beginning at the top of the column — and the flag says the number
		// is NOT EXACT, because a column Paginate could not place is
		// emphatically not a prediction of the document's length.
		projection.ContentWindowCount = 1
		projection.ContentWindowOrigins = []int64{0}
		projection.ContentWindowCountIsExact = false
		return nil
	}
	origins, ok := canvasWindowOrigins(plan)
	if !ok {
		// The same degradation, for the same reason: a sequence that would
		// not survive the browser's own validation must never be sent, and
		// discarding the number is cheaper than discarding the snapshot.
		projection.ContentWindowCount = 1
		projection.ContentWindowOrigins = []int64{0}
		projection.ContentWindowCountIsExact = false
		return nil
	}
	projection.ContentWindowCount = int64(len(plan.Pages))
	projection.ContentWindowOrigins = origins
	projection.ContentWindowCountIsExact = exact
	return nil
}

// addCanvasImagePaint is addCanvasTextPaint's sibling for image components
// (D-5.13.2's "Producer" clause): a paint PRODUCER invoked from
// CanvasWithTextPaint, never computed inside setComponentAsset or any other
// command — every mutating command's own Canvas(t) is discarded and
// recomputed by wasm/engine.go, so the paint must be derivable from
// template state alone, exactly like text paint.
//
// It builds each run in the BAND frame (element.X/Y untranslated by band
// origin, matching this component's own X/Y) and calls the same
// resolveImagePlacement collectImageRuns/renderDocument use for the PDF —
// never a second fit computation. canvas_image_paint_test.go asserts that
// this band-relative rectangle is exactly a translation of the page-absolute
// one collectImageRuns/resolveImagePlacement produce, rather than assuming
// it from the two call sites merely sharing a function.
//
// A missing asset key or a decode failure (unrecognised media type or
// malformed bytes) leaves this component's Image field absent and does NOT
// fail the whole projection — Render (render.go) is the located, fatal
// diagnostic for a genuinely broken document; this paint-only projection
// must stay paintable (AC3: "not a crash") even for a document a save
// cannot yet produce a clean render from.
func addCanvasImagePaint(t *Template, projection *CanvasProjection) error {
	components := make(map[string]*CanvasComponent, len(projection.Components))
	for i := range projection.Components {
		component := &projection.Components[i]
		components[component.ID] = component
	}
	for _, band := range []struct {
		name     string
		elements []template.Element
	}{
		{bandPageHeader, t.doc.Bands.PageHeader.Elements},
		{bandContent, t.doc.Bands.Content.Elements},
		{bandPageFooter, t.doc.Bands.PageFooter.Elements},
	} {
		for _, element := range band.elements {
			if element.Type != template.ElementImage {
				continue
			}
			component := components[string(element.ID)]
			if component == nil || component.Band != band.name {
				return fmt.Errorf("folio: canvas image component %q is missing from geometry projection", element.ID)
			}
			if !element.Width.Set || !element.Height.Set || !element.Asset.Set {
				// Load-time validation (parse_bands.go) already makes these
				// required for a successfully parsed document — handled
				// rather than assumed, never reached in practice.
				continue
			}
			if element.Asset.Null {
				// A placed but unfilled box. Neither Image nor
				// ImageUnavailable is set: there is nothing to paint and
				// nothing has gone wrong, and that pairing is what tells the
				// designer to draw its empty placeholder instead of one of
				// the two failure texts.
				continue
			}
			assetKey := element.Asset.Value
			asset, ok := t.doc.Assets[assetKey]
			if !ok {
				missing := imageUnavailableMissing
				component.ImageUnavailable = &missing
				continue
			}
			raw, err := template.DecodeAssetBytes(asset)
			if err != nil {
				undecodable := imageUnavailableUndecodable
				component.ImageUnavailable = &undecodable
				continue
			}
			img, err := template.DecodeImageForRender(asset.MediaType, raw, assetKey, string(element.ID))
			if err != nil {
				undecodable := imageUnavailableUndecodable
				component.ImageUnavailable = &undecodable
				continue
			}
			run := imageRunSource{elementID: string(element.ID), assetKey: assetKey, x: element.X, y: element.Y, boxW: element.Width.Value, boxH: element.Height.Value}
			drawX, drawY, drawW, drawH := resolveImagePlacement(run, img)
			width, err := canvasDerived("image intrinsic width", geom.Length(img.Width()))
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			height, err := canvasDerived("image intrinsic height", geom.Length(img.Height()))
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			dx, err := canvasDerived("image draw x", drawX)
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			dy, err := canvasDerived("image draw y", drawY)
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			dw, err := canvasDerived("image draw width", drawW)
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			dh, err := canvasDerived("image draw height", drawH)
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			component.Image = &CanvasImagePaint{
				MediaType:  asset.MediaType,
				AssetKey:   assetKey,
				Width:      int64(width),
				Height:     int64(height),
				DrawX:      int64(dx),
				DrawY:      int64(dy),
				DrawWidth:  int64(dw),
				DrawHeight: int64(dh),
			}
		}
	}
	return nil
}

func addCanvasTextPaint(t *Template, projection *CanvasProjection, fs FontSet, column *canvasColumnExtents) error {
	components := make(map[string]*CanvasComponent, len(projection.Components))
	for i := range projection.Components {
		component := &projection.Components[i]
		components[component.ID] = component
	}
	// Story 7.9: the document's OWN keep-together declarations, read
	// through the render path's single authority (keepTogetherTags,
	// render.go) rather than re-read here. It takes the *Template and
	// nothing else — no data, no params, no FontSet — which is precisely
	// why grouping is knowable canvas-side and is a DEFECT to omit rather
	// than a shortfall to disclose. See addCanvasWindowCount's own use of
	// it for the non-text arm; the two arms must tag from the same index
	// or the column is grouped in halves.
	keepTogether := keepTogetherTags(t)
	// FROM THE DOCUMENT (Story 8.4), and this is the second of the two
	// sites that must be — predictDocument (render.go) is the other. The
	// canvas consumes the IDENTICAL advance the renderer does (AD-17), so
	// a canvas cache that could not see the document's carried faces would
	// measure a document the PDF does not print.
	cache := newDocumentFontCache(t)
	// degrade disposes of ONE element and carries on, and it is spelled
	// once because two different conditions reach it (D-7.4.2: DEGRADE
	// THIS ELEMENT, NEVER ABORT THE PROJECTION). See both call sites
	// below for what each one is.
	degrade := func(band string, element template.Element, component *CanvasComponent) {
		// AND THE WINDOW COUNT LOSES THIS ELEMENT'S EXTENTS. Nothing
		// downstream of the column could tell — an element with nothing
		// to say and an element that could not be shaped contribute the
		// identical nothing — so the fact is recorded here, where it is
		// known, and reported as the honesty flag. Only an element that
		// actually HAS a value loses anything: an unset or empty one
		// would have contributed no lines with a perfectly good chain.
		if band == bandContent && element.Value.Set && !element.Value.Null && element.Value.Value != "" {
			column.FontChainDegraded = true
		}
		component.TextPaint = &CanvasTextPaint{Lines: []CanvasTextLine{}}
	}
	for _, band := range []struct {
		name     string
		elements []template.Element
	}{
		{bandPageHeader, t.doc.Bands.PageHeader.Elements},
		{bandContent, t.doc.Bands.Content.Elements},
		{bandPageFooter, t.doc.Bands.PageFooter.Elements},
	} {
		for _, element := range band.elements {
			if element.Type != template.ElementText {
				continue
			}
			component := components[string(element.ID)]
			if component == nil || component.Band != band.name {
				return fmt.Errorf("folio: canvas text component %q is missing from geometry projection", element.ID)
			}
			chain, err := fontChain(t, element)
			if err != nil {
				// Existing designer documents can be structurally valid while
				// incomplete for production rendering (for example, a text
				// component without a chosen font chain). They remain loadable;
				// there is simply no honest measured paint to display yet.
				degrade(band.name, element, component)
				continue
			}
			// SCOPED TO THIS ELEMENT'S CHAIN (fontCache.forChain), shadowing
			// the shared cache for the rest of the iteration: the canvas is
			// the surface an author repairs a chain on, so a message about a
			// chain entry must name the chain their element draws through.
			cache := cache.forChain(element.Style.Value.FontFamily.Value)
			paint := &CanvasTextPaint{Lines: []CanvasTextLine{}}
			if !element.Value.Set || element.Value.Null || element.Value.Value == "" {
				component.TextPaint = paint
				continue
			}
			fontSize := defaultFontSizePt
			if element.Style.Set && !element.Style.Null && element.Style.Value.FontSize.Set && !element.Style.Value.FontSize.Null {
				fontSize = element.Style.Value.FontSize.Value
			}
			segs, _, err := shapeSegments(string(element.ID), chain, element.Value.Value, fs, cache, breaksAreConsumed)
			if err != nil {
				// A CHAIN ENTRY THIS BUILD CANNOT DRAW WITH IS A DOCUMENT
				// THE FORMAT CALLS VALID, and it degrades exactly as an
				// unresolvable chain does above (D-7.4.2, stated at the
				// truncation arm below in this same function).
				//
				// Story 8.4 made this reachable from document CONTENT: a
				// chain entry naming a non-font asset loads (D-1.8.1 as
				// amended) and is refused at coverage resolution, which is
				// right for Render — the page would be wrong — and wrong
				// here. The designer is the ONE surface on which an author
				// can repair that entry, and a projection that returns an
				// error opens no document at all, so the defect would lock
				// the author out of its own repair.
				//
				// SCOPED TO THE CAPABILITY ERROR, deliberately. Only
				// template.UnsupportedFontMediaTypeError is tolerated: a
				// genuine internal shaping fault is not a document
				// property, has no author repair, and still aborts the
				// projection as it always did.
				var unsupported *template.UnsupportedFontMediaTypeError
				if !errors.As(err, &unsupported) {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				degrade(band.name, element, component)
				continue
			}
			atomic := atomicSpansFor(t.doc.UnbreakableValues, nil)
			ops := text.Opportunities(text.Dictionary(), element.Value.Value, atomic)
			boxWidth := geom.Length(0)
			if element.Width.Set && !element.Width.Null {
				boxWidth = element.Width.Value
			}
			lines := packLines(segs, ops, len([]rune(element.Value.Value)), fontSize, boxWidth)
			// D-7.4.2: DEGRADE THIS ELEMENT, NEVER ABORT THE PROJECTION.
			// This used to `return` an error, and that error was the
			// function's own — one over-long clause blanked the whole
			// canvas for a document that renders to a perfectly good PDF.
			// The shape reused here is the fontChain path a few lines
			// above: dispose of the one element and carry on. What is new
			// is that the element keeps its first N lines and SAYS it was
			// cut, rather than presenting an empty paint.
			//
			// The lines dropped here are dropped from the PAINT ONLY.
			// element.Value is untouched, is what the properties panel
			// saves, and renders whole.
			painted := lines
			if len(painted) > maxCanvasBodyTextLines {
				painted = painted[:maxCanvasBodyTextLines]
				paint.Truncated = true
			}
			// Overflow and the vertical origin below are still derived from
			// the FULL line list, so the prefix paints at exactly the
			// coordinates it occupies in the whole block: truncation must
			// not silently move the text the author can still see.
			_, paint.Overflow = detectWidthOverflow(string(element.ID), lines, boxWidth)
			// AC6 / the Story 5.9 invariant: the canvas consumes the
			// IDENTICAL advance the renderer does, ratio included — the
			// browser never measures text and never adjudicates what the
			// engine measured.
			vm, err := chainVerticalModel(chain, fontSize, styleLineSpacing(element.Style), fs, cache)
			if err != nil {
				return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
			}
			// The same slack-only alignment rule the PDF producer applies
			// (text_alignment.go), from the same committed style: the canvas
			// has to show what prints. Both offsets are non-negative, so every
			// projected coordinate stays inside the band-relative, JS-safe
			// bound canvasLineTop and canvasDerived check.
			align, valign := elementAlignment(element)
			boxHeight := geom.Length(0)
			if element.Height.Set && !element.Height.Null {
				boxHeight = element.Height.Value
			}
			originY := element.Y + textValignOffset(valign, boxHeight, textBlockHeight(len(lines), vm))
			// THE WINDOW COUNT'S EXTENTS, taken here and nowhere else.
			//
			// It iterates `lines` — the FULL, untruncated list — and never
			// `painted`, `budget`, `oversized` or the placed runs, so the
			// count is identical whether this element paints every line, a
			// truncated prefix, or (the first-line-too-tall path) none at
			// all. That independence is the whole point: a canvas that
			// stopped drawing at line 1920 must not also stop counting
			// pages there.
			//
			// The extent is the render path's, term for term rather than
			// re-derived: `top` here is exactly the `lineY` render.go
			// places a line at, and the bottom adds the same
			// FirstBaseline + LastDescent from the same vertical model.
			if band.name == bandContent {
				for i := range lines {
					top, err := canvasLineTop(originY, i, vm.Advance)
					if err != nil {
						return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
					}
					column.Items = append(column.Items, layout.ColumnItem{
						ElementID: string(element.ID),
						Top:       top,
						Bottom:    top + vm.FirstBaseline + vm.LastDescent,
						Runs:      []layout.TextRunRef{0},
						// Story 7.9 (FR51): the SAME group the render
						// path's own line items carry
						// (contentColumnItems / paginateDocument). A
						// canvas line has no prior group — only a
						// table's row items do, and a table's cells are
						// not text elements — so the group is taken
						// directly rather than through orKeepTogether,
						// exactly as the render path's image arm does.
						// An untagged element gets the ZERO ItemGroup,
						// which is what it carried before this story and
						// is what keeps an ungrouped document identical.
						Group: keepTogether.keepTogetherGroup(string(element.ID)),
					})
				}
			}
			budget := canvasFragmentBudget{}
			for i, line := range painted {
				top, err := canvasLineTop(originY, i, vm.Advance)
				if err != nil {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				lineX := element.X + textAlignOffset(align, boxWidth, line.width)
				// THE IDENTICAL BRANCH render.go's line loop carries,
				// from the identical shared rule (Story 7.3): the
				// canvas shows the word positions the PDF prints, and
				// it gets them by consuming engine-computed offsets —
				// never by asking the browser to justify, which
				// canvas-authority-contract.test.ts bans across every
				// production, unit and e2e source.
				var placed []textRunSource
				if pieces := justifiedLinePieces(align, line, i, len(lines), segs, ops, fontSize, boxWidth); pieces != nil {
					for _, piece := range pieces {
						pieceRuns, pieceErr := positionSegments(segs, piece.from, piece.to, element.X+piece.offset, top, fontSize, vm.FirstBaseline, nil)
						if pieceErr != nil {
							return fmt.Errorf("folio: canvas text element %s: %w", element.ID, pieceErr)
						}
						placed = append(placed, pieceRuns...)
					}
				} else {
					var perr error
					placed, perr = positionSegments(segs, line.from, line.to, lineX, top, fontSize, vm.FirstBaseline, nil)
					if perr != nil {
						return fmt.Errorf("folio: canvas text element %s: %w", element.ID, perr)
					}
				}
				baseline, err := canvasDerivedSum(top, vm.FirstBaseline)
				if err != nil {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				advance, err := canvasDerived("line advance", vm.Advance)
				if err != nil {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				width, err := canvasDerived("line width", line.width)
				if err != nil {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				paintLine := CanvasTextLine{Top: int64(top), Baseline: int64(baseline), Advance: int64(advance), Width: int64(width), Fragments: []CanvasTextFragment{}}
				// A fragment's text is BODY TEXT, not an identifier: this
				// site is the second of the two maxCanvasPropertyString
				// conflations DW-25 undercounted, and a value that got past
				// the value cap used to abort here instead.
				oversized := false
				for _, fragment := range placed {
					if len(fragment.text) > maxCanvasBodyText {
						oversized = true
						break
					}
					x, err := canvasDerived("fragment x", fragment.x)
					if err != nil {
						return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
					}
					paintLine.Fragments = append(paintLine.Fragments, CanvasTextFragment{Text: fragment.text, X: int64(x)})
				}
				// Painting stops at the last WHOLE line that fits. A half
				// line would be a worse lie than a short one: the author
				// would read a sentence the document does not contain.
				if oversized || !budget.admits(len(paintLine.Fragments)) {
					paint.Truncated = true
					break
				}
				budget.take(len(paintLine.Fragments))
				paint.Lines = append(paint.Lines, paintLine)
			}
			component.TextPaint = paint
		}
	}
	return nil
}

// canvasFragmentBudget is the two fragment bounds a painted line must satisfy,
// held in one place because they bound DIFFERENT QUANTITIES and are easy to
// mistake for one another.
//
//   - maxCanvasTextFragments is PER LINE. It is Go's own long-standing guard
//     on one degenerate line and has no counterpart in the browser.
//   - maxCanvasBodyTextFragments is CUMULATIVE across the whole element. It
//     exists to mirror engine-protocol.ts's `fragments` counter, which is
//     declared once per component and never reset — so a projection whose
//     every line is per-line legal can still be refused there, and a refusal
//     there drops the ENTIRE engine response with no attributable error.
//
// The asymmetry is why the engine has to carry the cumulative count itself:
// satisfying the per-line guard says nothing about the browser's bound.
type canvasFragmentBudget struct{ used int }

// admits reports whether a line carrying count fragments can still be
// painted. Both bounds must hold; neither implies the other.
func (b *canvasFragmentBudget) admits(count int) bool {
	return count <= maxCanvasTextFragments && b.used+count <= maxCanvasBodyTextFragments
}

func (b *canvasFragmentBudget) take(count int) { b.used += count }

func canvasDerived(name string, value geom.Length) (geom.Length, error) {
	if value < 0 || value > geom.Length(MaxCanvasMillipoints) {
		return 0, fmt.Errorf("%s exceeds the JavaScript-safe projection bound", name)
	}
	return value, nil
}

func canvasDerivedSum(left, right geom.Length) (geom.Length, error) {
	if left < 0 || right < 0 || left > geom.Length(MaxCanvasMillipoints)-right {
		return 0, fmt.Errorf("derived canvas coordinate exceeds the JavaScript-safe projection bound")
	}
	return left + right, nil
}

func canvasLineTop(elementY geom.Length, index int, advance geom.Length) (geom.Length, error) {
	if index < 0 || advance < 0 || elementY < 0 || advance > 0 && geom.Length(index) > (geom.Length(MaxCanvasMillipoints)-elementY)/advance {
		return 0, fmt.Errorf("derived canvas line origin exceeds the JavaScript-safe projection bound")
	}
	return canvasDerivedSum(elementY, geom.Length(index)*advance)
}

func canvasComponents(t *Template, bands []CanvasBand) ([]CanvasComponent, error) {
	out := make([]CanvasComponent, 0)
	for _, projected := range bands {
		var elements []template.Element
		switch projected.Name {
		case bandPageHeader:
			elements = t.doc.Bands.PageHeader.Elements
		case bandContent:
			elements = t.doc.Bands.Content.Elements
		case bandPageFooter:
			elements = t.doc.Bands.PageFooter.Elements
		}
		for _, element := range elements {
			width, height := projectedSize(element)
			for _, value := range []geom.Length{element.X, element.Y, width, height} {
				if value > geom.Length(MaxCanvasMillipoints) {
					return nil, fmt.Errorf("folio: component exceeds the JavaScript-safe geometry bound")
				}
			}
			component := CanvasComponent{ID: string(element.ID), Type: string(element.Type), Band: projected.Name, X: int64(element.X), Y: int64(element.Y), Width: int64(width), Height: int64(height), Resizable: element.Type != template.ElementTable}
			if element.Type == template.ElementText && element.Value.Set && !element.Value.Null {
				// BODY TEXT, so the body-text backstop — not the identifier
				// bound this used to share. At 512 bytes this returned nil
				// for the ENTIRE component list, and because the property
				// command re-projects inside its own transaction
				// (component_commands.go's updateComponentProperties) it did
				// not merely blank the canvas: it REJECTED the author's edit
				// at about eighty English words, or a hundred and seventy
				// Thai characters. The refusal stays — a value may not be
				// truncated, only its paint (D-7.4.2 §1) — but it is now a
				// megabyte-scale channel backstop Epic 7's input cannot
				// reach, not a clause-length cap.
				if len(element.Value.Value) > maxCanvasBodyText {
					return nil, fmt.Errorf("folio: component value exceeds the projection bound")
				}
				component.Value = stringPointer(element.Value.Value)
				if binding := directCanvasBinding(element.Value.Value); binding != "" {
					component.Binding = stringPointer(binding)
				}
			}
			if element.VisibleIf.Set && !element.VisibleIf.Null {
				if len(element.VisibleIf.Value) > maxCanvasPropertyString {
					return nil, fmt.Errorf("folio: component visibleIf exceeds the projection bound")
				}
				component.VisibleIf = stringPointer(element.VisibleIf.Value)
			}
			if element.Type == template.ElementTable && element.Table.Set && !element.Table.Null {
				if len(element.Table.Value.Bind) > maxCanvasPropertyString {
					return nil, fmt.Errorf("folio: component table bind exceeds the projection bound")
				}
				component.TableBind = stringPointer(element.Table.Value.Bind)
			}
			if element.Style.Set && !element.Style.Null {
				if err := applyCanvasStyle(&component, element.Type, element.Style.Value); err != nil {
					return nil, err
				}
			}
			out = append(out, component)
		}
	}
	return out, nil
}

func directCanvasBinding(value string) string {
	literal, placeholders, trailing, err := expr.ScanPlaceholders(value)
	if err != nil || len(literal) != 1 || literal[0] != "" || len(placeholders) != 1 || trailing != "" || placeholders[0].Reserved {
		return ""
	}
	parsed, err := expr.Parse(placeholders[0].Inner)
	if err != nil {
		return ""
	}
	path, ok := parsed.(*expr.PathExpr)
	if !ok || path.Raw == "" || len(path.Raw) > maxCanvasBindingString {
		return ""
	}
	return path.Raw
}

func stringPointer(value string) *string     { return &value }
func boolPointer(value bool) *bool           { return &value }
func int64Pointer(value int64) *int64        { return &value }
func lengthPointer(value geom.Length) *int64 { rendered := int64(value); return &rendered }
func canvasPropertyLength(name string, value geom.Length) (*int64, error) {
	if value < 0 || value > geom.Length(MaxCanvasMillipoints) {
		return nil, fmt.Errorf("folio: component %s exceeds the projection bound", name)
	}
	return lengthPointer(value), nil
}

func applyCanvasStyle(component *CanvasComponent, elementType template.ElementType, style template.Style) error {
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.FontFamily.Set && !style.FontFamily.Null {
		if len(style.FontFamily.Value) > maxCanvasPropertyString {
			return fmt.Errorf("folio: component fontFamily exceeds the projection bound")
		}
		component.FontFamily = stringPointer(style.FontFamily.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.FontSize.Set && !style.FontSize.Null {
		value, err := canvasPropertyLength("fontSize", style.FontSize.Value)
		if err != nil {
			return err
		}
		component.FontSize = value
	}
	// style.lineSpacing, projected for the first time (Story 7.4). Its range
	// is already settled at load and on the property-command path by the one
	// validator both call (template.DecodeLineSpacing, D-7.2.3), so this is a
	// projection of a committed value and not a second opinion on it.
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.LineSpacing.Set && !style.LineSpacing.Null {
		component.LineSpacing = int64Pointer(style.LineSpacing.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.Bold.Set && !style.Bold.Null {
		component.Bold = boolPointer(style.Bold.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.Italic.Set && !style.Italic.Null {
		component.Italic = boolPointer(style.Italic.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.Align.Set && !style.Align.Null {
		component.Align = stringPointer(style.Align.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.Valign.Set && !style.Valign.Null {
		component.Valign = stringPointer(style.Valign.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.Color.Set && !style.Color.Null {
		if len(style.Color.Value) > maxCanvasPropertyString {
			return fmt.Errorf("folio: component color exceeds the projection bound")
		}
		component.Color = stringPointer(style.Color.Value)
	}
	if style.Background.Set && !style.Background.Null {
		if len(style.Background.Value) > maxCanvasPropertyString {
			return fmt.Errorf("folio: component background exceeds the projection bound")
		}
		component.Background = stringPointer(style.Background.Value)
	}
	if style.Border.Set && !style.Border.Null {
		border := style.Border.Value
		if border.Width.Set && !border.Width.Null {
			value, err := canvasPropertyLength("borderWidth", border.Width.Value)
			if err != nil {
				return err
			}
			component.BorderWidth = value
		}
		if border.Color.Set && !border.Color.Null {
			if len(border.Color.Value) > maxCanvasPropertyString {
				return fmt.Errorf("folio: component borderColor exceeds the projection bound")
			}
			component.BorderColor = stringPointer(border.Color.Value)
		}
		if border.Edges.Set && !border.Edges.Null {
			component.BorderEdges = append([]string(nil), border.Edges.Value...)
		}
	}
	if style.Padding.Set && !style.Padding.Null {
		padding := style.Padding.Value
		if padding.Top.Set && !padding.Top.Null {
			value, err := canvasPropertyLength("paddingTop", padding.Top.Value)
			if err != nil {
				return err
			}
			component.PaddingTop = value
		}
		if padding.Right.Set && !padding.Right.Null {
			value, err := canvasPropertyLength("paddingRight", padding.Right.Value)
			if err != nil {
				return err
			}
			component.PaddingRight = value
		}
		if padding.Bottom.Set && !padding.Bottom.Null {
			value, err := canvasPropertyLength("paddingBottom", padding.Bottom.Value)
			if err != nil {
				return err
			}
			component.PaddingBottom = value
		}
		if padding.Left.Set && !padding.Left.Null {
			value, err := canvasPropertyLength("paddingLeft", padding.Left.Value)
			if err != nil {
				return err
			}
			component.PaddingLeft = value
		}
	}
	return nil
}

func canvasDimensions(t *Template) (geom.Length, geom.Length, error) {
	var width, height geom.Length
	switch {
	case t.doc.Page.SizeIsName && t.doc.Page.SizeName == "A4":
		width, height = 595276, 841890
	case t.doc.Page.SizeIsName && t.doc.Page.SizeName == "Letter":
		width, height = 612000, 792000
	case !t.doc.Page.SizeIsName:
		width, height = t.doc.Page.SizeCustom.Width, t.doc.Page.SizeCustom.Height
	default:
		return 0, 0, fmt.Errorf("folio: unsupported page size")
	}
	if t.doc.Page.Orientation == "landscape" {
		width, height = height, width
	}
	return width, height, nil
}

// ApplyPageSetupCommand decodes the one versioned, Go-defined opaque command.
// Numeric input stays a JSON literal until exact millipoint conversion; it is
// never decoded through float64.
func ApplyPageSetupCommand(t *Template, command []byte) (CanvasProjection, error) {
	if t == nil {
		return CanvasProjection{}, errNilTemplate
	}
	dec := json.NewDecoder(bytes.NewReader(command))
	dec.UseNumber()
	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil || dec.More() {
		return CanvasProjection{}, fmt.Errorf("folio: page setup command is malformed")
	}
	if len(raw) != 7 || !equalString(raw["kind"], "pageSetup") || !equalNumber(raw["version"], "1") {
		return CanvasProjection{}, fmt.Errorf("folio: unknown page setup command")
	}
	preset, orientation := stringField(raw, "preset"), stringField(raw, "orientation")
	if orientation != "portrait" && orientation != "landscape" {
		return CanvasProjection{}, fmt.Errorf("folio: page.orientation must be portrait or landscape")
	}
	if preset != "A4" && preset != "Letter" && preset != "custom" {
		return CanvasProjection{}, fmt.Errorf("folio: page.size must be A4, Letter, or custom")
	}
	var width, height geom.Length
	if preset == "custom" {
		var err error
		width, err = lengthField(raw, "width")
		if err != nil {
			return CanvasProjection{}, fmt.Errorf("folio: page.width: %w", err)
		}
		height, err = lengthField(raw, "height")
		if err != nil {
			return CanvasProjection{}, fmt.Errorf("folio: page.height: %w", err)
		}
	}
	marginRaw, ok := raw["margin"]
	if !ok {
		return CanvasProjection{}, fmt.Errorf("folio: page.margin is required")
	}
	var margins map[string]json.RawMessage
	if json.Unmarshal(marginRaw, &margins) != nil || len(margins) != 4 {
		return CanvasProjection{}, fmt.Errorf("folio: page.margin must contain top, right, bottom, left")
	}
	readMargin := func(name string) (geom.Length, error) {
		v, err := lengthField(margins, name)
		if err != nil {
			return 0, fmt.Errorf("folio: page.margin.%s: %w", name, err)
		}
		if v < 0 {
			return 0, fmt.Errorf("folio: page.margin.%s must not be negative", name)
		}
		return v, nil
	}
	top, err := readMargin("top")
	if err != nil {
		return CanvasProjection{}, err
	}
	right, err := readMargin("right")
	if err != nil {
		return CanvasProjection{}, err
	}
	bottom, err := readMargin("bottom")
	if err != nil {
		return CanvasProjection{}, err
	}
	left, err := readMargin("left")
	if err != nil {
		return CanvasProjection{}, err
	}
	if preset == "A4" {
		width, height = 595276, 841890
	} else if preset == "Letter" {
		width, height = 612000, 792000
	}
	if width <= 0 || height <= 0 {
		return CanvasProjection{}, fmt.Errorf("folio: page.size width and height must be positive")
	}
	if preset == "custom" && (width <= 0 || height <= 0) {
		return CanvasProjection{}, fmt.Errorf("folio: custom page size is required")
	}
	before, err := SerializeTemplate(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	page := &t.doc.Page
	page.Orientation = orientation
	page.Margin = template.Margin{Top: top, Right: right, Bottom: bottom, Left: left}
	if preset == "custom" {
		page.SizeIsName = false
		page.SizeName = ""
		page.SizeCustom = template.PageSize{Width: width, Height: height}
	} else {
		page.SizeIsName = true
		page.SizeName = preset
		page.SizeCustom = template.PageSize{}
	}
	projection, err := Canvas(t)
	if err != nil {
		restorePage(t, before)
		return CanvasProjection{}, err
	}
	return projection, nil
}

func restorePage(t *Template, canonical []byte) {
	restored, err := ParseTemplate(canonical)
	if err == nil {
		t.doc = restored.doc
		t.derivedFooters = restored.derivedFooters
	}
}
func equalString(raw json.RawMessage, want string) bool {
	var got string
	return json.Unmarshal(raw, &got) == nil && got == want
}
func equalNumber(raw json.RawMessage, want string) bool { return string(raw) == want }
func stringField(raw map[string]json.RawMessage, key string) string {
	var value string
	_ = json.Unmarshal(raw[key], &value)
	return value
}
func lengthField(raw map[string]json.RawMessage, key string) (geom.Length, error) {
	v, ok := raw[key]
	if !ok {
		return 0, fmt.Errorf("%s is required", key)
	}
	literal := string(v)
	if strings.ContainsAny(literal, "eE") {
		return 0, fmt.Errorf("%s must be a decimal with at most three places", key)
	}
	if dot := strings.IndexByte(literal, '.'); dot >= 0 && len(literal)-dot-1 > 3 {
		return 0, fmt.Errorf("%s must have at most three decimal places", key)
	}
	var n json.Number
	if json.Unmarshal(v, &n) != nil {
		return 0, fmt.Errorf("%s must be a finite number", key)
	}
	value, err := parseMillipoints(literal, key)
	if err != nil {
		return 0, err
	}
	if value > geom.Length(MaxCanvasMillipoints) || value < -geom.Length(MaxCanvasMillipoints) {
		return 0, fmt.Errorf("%s exceeds the JavaScript-safe geometry bound", key)
	}
	return value, nil
}
func parseMillipoints(literal, key string) (geom.Length, error) {
	negative := strings.HasPrefix(literal, "-")
	if negative {
		literal = literal[1:]
	}
	parts := strings.Split(literal, ".")
	if len(parts) > 2 || len(parts[0]) == 0 {
		return 0, fmt.Errorf("%s must be a number", key)
	}
	whole := int64(0)
	for _, c := range parts[0] {
		if c < '0' || c > '9' || whole > (1<<63-1)/10 {
			return 0, fmt.Errorf("%s overflows millipoints", key)
		}
		whole = whole*10 + int64(c-'0')
	}
	frac := int64(0)
	if len(parts) == 2 {
		for _, c := range parts[1] {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("%s must be a number", key)
			}
			frac = frac*10 + int64(c-'0')
		}
		for len(parts[1]) < 3 {
			parts[1] += "0"
			frac *= 10
		}
	}
	if whole > (1<<63-1)/1000 || (whole == (1<<63-1)/1000 && frac > (1<<63-1)%1000) {
		return 0, fmt.Errorf("%s overflows millipoints", key)
	}
	value := whole*1000 + frac
	if negative {
		value = -value
	}
	return geom.Length(value), nil
}
