package layout

// Story 2.6's pagination, ruled by D-2.6.1 and amended by the sliding-window
// ruling that resolved its residual arithmetic.
//
// THE MODEL, in one sentence: the content band is a WINDOW onto one
// unbounded content column, and a longer document is more windows, never
// rearranged furniture.
//
// FOUR RULES, and each is stated as an OUTCOME rather than as a mechanism:
//
//  1. WINDOW. A page shows a page-height window onto the content column.
//     Window 0 begins at the top of the content band. Window N+1 begins at
//     the top of the FIRST ITEM THAT DID NOT FIT in window N.
//  2. LINE GRANULARITY, AND IT IS FORCED, NOT CHOSEN. The atom is the LINE,
//     not the element. It is forced because a single text element can wrap
//     past a page: with element granularity, an element taller than a page
//     would be UNPLACEABLE. This is not a preference and must not be
//     revisited as one.
//  3. FIT ENTIRELY. No line is ever split across a page boundary. A line is
//     placed on the first page whose window contains it ENTIRELY — from
//     `baseline − max(ascent)` to `baseline + max(descent)`, per D-2.4.2 as
//     amended. Whitespace at the foot of a page is CORRECT TYPESETTING, not
//     a defect. The rejected reading — "a line belongs to whichever page
//     contains its top" — permits a line drawn half on each of two pages,
//     and NOTHING in the repository would diagnose it: FR44's overflow
//     machinery concerns content exceeding its BOX, not the page edge. A
//     bank statement cannot ship a half-line.
//  4. IMAGES ARE ATOMIC, same rule. An image goes to the first page whose
//     window contains its DECLARED BOX entirely. Because AD-24 already
//     scales the image to fit that box, an image that fits nowhere is a
//     statement about the BOX — a template error with a located message,
//     never a silent clip and never a straddle.
//
// WHY THE WINDOW SLIDES RATHER THAN SITTING ON A FIXED GRID, recorded
// because the rejected alternatives are the ones a reader will re-propose:
//
//   - A FIXED GRID at [N·H, (N+1)·H) with each straddling item bumped flush
//     to the next window's top was ELIMINATED BY MEASUREMENT. On the Noto
//     Sans 12pt chain (advance 16,344mp) with H = 727,890mp, line 44 spans
//     719,136..735,480 and straddles. Bumping it alone puts line 44 at 0 and
//     line 45 at 7,590 — less than the 16,344 advance, so the two lines
//     OVERLAP. The model produces a defect worse than the one it fixes.
//   - A FIXED GRID with the bump PROPAGATING through the column (every later
//     item shifted by the same delta) keeps the leading intact but violates
//     AD-24 by its own text: "a hidden element is absent from the PageModel
//     and leaves no gap; SIBLINGS NEVER MOVE, BECAUSE NOTHING IN A BAND EVER
//     REFLOWS." Under it an absolutely-positioned sibling with nothing to do
//     with the overflow moves. That is the case the clause describes, not an
//     analogy to it.
//
// So the window slides and the COLUMN IS NEVER MUTATED. Every declared
// column Y is untouched — which is why "elements never reposition" and "a
// paragraph's leading survives a page break" are both true here rather than
// traded off: the gap between line 43 and line 44 is still exactly one
// advance; the two merely fall in different windows.
//
// PAGE COUNT IS NOT A CLOSED FORM. It falls out of the advance. In
// particular it is NOT `ceil(lowestBottom / H)`, and that spelling must not
// be reintroduced: the window advances to the first UNPLACED item, not by a
// fixed H, so an element declared far below the text STARTS THE NEXT WINDOW
// rather than generating blank pages before it. NO PAGE IS EVER EMPTY, and
// TestPaginateNeverProducesAnEmptyPage asserts it rather than trusting it.
//
// AD-4 and AD-24 both hold here and neither is weakened: this is pass ONE.
// internal/pdf receives a finished []pagemodel.Page and decides nothing.
// ContentHeight is still a function of PageGeometry alone — pagination
// CONSUMES it and no item's measurement reaches it.

import (
	"slices"
	"strconv"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// ColumnItem is one ATOMIC, indivisible unit of the content column: one
// LINE of text (never a fraction of one), or one whole image placement.
//
// Top and Bottom are the item's vertical extent in the SAME page-absolute
// coordinate the runs and placements already carry — the caller has already
// resolved the band origin through PlaceInBand, so nothing here adds one
// (AD-24: "bands are placed on the page by internal/layout alone", and they
// were, upstream of this call).
//
// WHY THE EXTENT IS CARRIED RATHER THAN DERIVED HERE. A line's extent is
// `baseline − max(ascent)` .. `baseline + max(descent)` over the element's
// DECLARED font chain (D-2.4.2 amended) — a quantity the vertical model
// already computed once. Re-deriving it here would be a second derivation of
// the same number, which is exactly what D-2.4.2's amendment exists to
// prevent. An image's extent is its DECLARED BOX, not its drawn box, per
// rule 4 above.
type ColumnItem struct {
	// ElementID names the document element this item came from. It is for
	// the located overflow diagnostic ONLY — no geometry is derived from it.
	ElementID string

	// Top, Bottom are the item's vertical extent. Bottom >= Top.
	Top, Bottom geom.Length

	// Runs and Images are the item's content. Exactly one is populated: an
	// item is a line OR an image, never both, because the two have different
	// atomicity arguments (a line is atomic by rule 3, an image by rule 4).
	Runs   []TextRunRef
	Images []ImageRef
}

// TextRunRef and ImageRef index the caller's own run/placement slices rather
// than carrying copies. Pagination decides WHICH PAGE and BY HOW MUCH TO
// SHIFT; it does not rebuild content, so it never has an opportunity to
// reorder or drop a run.
type TextRunRef int

// ImageRef indexes the caller's placement slice, for the same reason.
type ImageRef int

// BandContent is the page-header or page-footer band's finished content,
// which is REPEATED VERBATIM on every page.
//
// It is repeated rather than recomputed: the running header on page 34 is
// the same bytes as on page 1, at the same band origin, which is what "page
// 34 is as complete as page 1" means and what makes the per-page header/
// footer assertion a comparison of equal literals rather than of two
// derivations.
type BandContent struct {
	Runs   []TextRunRef
	Images []ImageRef
}

// OverflowError is FR44's located diagnostic for an item that fits in NO
// window — the residual case once the window can slide, and therefore
// exactly "the item is taller than the content window".
//
// ONE OVERFLOW RULE, TWO SUBJECTS. D-2.6.1 ruled the image case: an image
// whose declared box exceeds the content window "fits nowhere", and that is
// a TEMPLATE error with a located message rather than a render-time
// surprise. A LINE taller than the window (reachable with a font size larger
// than the content band) is the same case and gets the same answer, so the
// implementation is never left to split, loop or panic.
//
// NEVER A STRADDLE AND NEVER A SILENT CLIP: both are what this error exists
// to prevent.
type OverflowError struct {
	ElementID     string
	ItemHeight    geom.Length
	ContentHeight geom.Length
	Kind          string // "line" or "image"
}

// millipoints spells a geom.Length for a HUMAN-READABLE diagnostic. It is
// not an output-format emitter: nothing in a PDF, a page model or a golden
// passes through here, so AD-3's "one number emitter" is untouched (that
// rule governs internal/pdf's numbers.go, and lint's numeric-formatting
// scan is scoped to that package).
func millipoints(v geom.Length) string {
	return strconv.FormatInt(int64(v), 10) + "mp"
}

func (e *OverflowError) Error() string {
	return "element " + e.ElementID + ": " + e.Kind +
		" is taller than the content window (" + millipoints(e.ItemHeight) +
		" against a content height of " + millipoints(e.ContentHeight) +
		"), so it fits on no page. Under AD-24 the content band does not grow to fit, and a line is " +
		"never split across a page boundary — reduce the element's font size, or its declared box " +
		"height for an image, or increase the page's content height by reducing its margins or its " +
		"page-header/page-footer heights."
}

// MixedItemError reports a ColumnItem that is neither exactly one line nor
// exactly one image.
//
// It is a PROGRAMMING error, not a template error, and it is deliberately a
// distinct type from OverflowError so a caller cannot confuse "this document
// does not fit" with "this item was built wrongly". An item carrying both
// runs and images would have its OverflowError.Kind mislabelled; an item
// carrying neither occupies column space while drawing nothing, which would
// let an invisible item push visible content onto another page.
type MixedItemError struct {
	ElementID string
	// Empty distinguishes "both" from "neither", because the two have
	// different causes and a single message for both would send a reader
	// looking in the wrong place.
	Empty bool
}

func (e *MixedItemError) Error() string {
	if e.Empty {
		return "element " + e.ElementID + ": a column item carries neither a text run nor an image, " +
			"so it would occupy column space while drawing nothing — an invisible item can push visible " +
			"content onto another page"
	}
	return "element " + e.ElementID + ": a column item carries BOTH text runs and an image. " +
		"A line and an image are atomic for different reasons and an overflow diagnostic about this " +
		"item would name the wrong kind; build one item for each"
}

// Pagination is Paginate's answer: for each page, which of the caller's runs
// and placements appear on it, and the vertical SHIFT to apply to each.
//
// The shift is expressed once per page rather than per run because every
// item on a page shares it: it is `windowStart(page) − contentBandOrigin`,
// the distance the window has slid. A caller applies `Y − Shift` and nothing
// else, which is why pagination cannot introduce a per-item displacement
// even by accident.
type Pagination struct {
	Pages []PageAssignment
}

// PageAssignment is one page's content and its window shift.
type PageAssignment struct {
	// Shift is subtracted from every content run's and image's Y to move it
	// from column space into this page's space. It is ZERO on page 0 of any
	// document whose content fits one window — which is what keeps every
	// pre-2.6 golden byte-identical.
	//
	// CHECKED, NOT CHANGED (Story 2.6 finisher, Finding 14). The review
	// claimed this comment overstates the guarantee — that window 0 begins
	// at "the top of the FIRST ITEM", so a single element positioned low in
	// the content band would still get a non-zero page-0 shift. Measured
	// against the actual rule (rule 1, this file's package doc, unchanged
	// by this story) and against the code: window 0 begins at the CONTENT
	// BAND'S OWN TOP, unconditionally — only window N+1 (N>=0) begins at
	// the first item that did not fit in window N. So Shift is
	// unconditionally zero on page 0 for ANY document, not merely one whose
	// content "fits one window" (that phrase, if anything, understates the
	// guarantee rather than overstating it). Confirmed empirically: a
	// single item declared 80,000mp below the content band's top, well
	// inside one window, produces Shift 0 and is placed at its own declared
	// Y — nothing pulls it to the band's top. The finding's claim does not
	// hold against this file's own ruled model; the comment is left as it
	// was.
	Shift geom.Length

	// ContentRuns/ContentImages are the CONTENT band's items on this page,
	// in the caller's ORIGINAL ORDER, not in the order the sweep visited
	// them. Emission order reaches output bytes (object and operator order),
	// and the sweep visits in COLUMN order, which is not necessarily
	// authored order. Preserving authored order is what makes a one-page
	// document byte-identical to its pre-2.6 self.
	ContentRuns   []TextRunRef
	ContentImages []ImageRef
}

// Paginate slices the content column into windows and returns one assignment
// per page. It is the ONE function that decides how many pages a document
// has.
//
// Integer arithmetic on geom.Length (int64 millipoints) throughout: no
// division, no rounding, no float (AD-23, AD-2). The only operations are
// addition, subtraction and comparison, so the page partition is exact and
// identical on every target — which is what lets the multi-page golden be
// compared byte-for-byte across four platforms.
//
// A document with NO content items is ONE page, not zero: a page-header and
// page-footer with an empty content band is a legitimate document, and a
// zero-page PDF is not.
func Paginate(g PageGeometry, items []ColumnItem) (Pagination, error) {
	contentTop := Origins(g).Content
	height := ContentHeight(g)

	// One page always exists. Growth adds windows to this.
	pages := []PageAssignment{{Shift: 0}}

	if len(items) == 0 {
		return Pagination{Pages: pages}, nil
	}

	// EXCLUSIVITY, ASSERTED RATHER THAN ASSUMED (D-2.6.5's guardrail).
	// OverflowError.Kind is derived as "image" when the item carries any
	// image ref and "line" otherwise, so an item carrying BOTH would be
	// MISLABELLED in a diagnostic a template author is expected to act on.
	// The two construction sites in package folio build one or the other and
	// never both — but "never both" is a claim about a caller, and a claim
	// about a caller is exactly the kind that stops being true quietly. It
	// is cheap to check here, so it is checked here.
	for _, it := range items {
		if len(it.Runs) > 0 && len(it.Images) > 0 {
			return Pagination{}, &MixedItemError{ElementID: it.ElementID}
		}
		if len(it.Runs) == 0 && len(it.Images) == 0 {
			return Pagination{}, &MixedItemError{ElementID: it.ElementID, Empty: true}
		}
	}

	// The sweep visits items in COLUMN order. Authored order is not
	// necessarily column order — an element declared second may sit above
	// one declared first — and the window's advance is a fact about the
	// column, not about the authoring. Ties keep authored order (stable), so
	// two items sharing a top edge cannot be reordered by the sort.
	order := make([]int, len(items))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		switch {
		case items[a].Top < items[b].Top:
			return -1
		case items[a].Top > items[b].Top:
			return 1
		default:
			return 0
		}
	})

	// pageOf[i] is the page authored-item i landed on. Filled by the sweep,
	// read by the emission below — which is what lets the sweep run in column
	// order while emission runs in authored order.
	pageOf := make([]int, len(items))

	windowStart := contentTop
	page := 0
	pageHasItem := false

	for _, idx := range order {
		it := items[idx]

		// Does this item fit ENTIRELY in the window as it currently stands?
		if it.Bottom > windowStart+height {
			// It does not. The window slides to begin at THIS ITEM'S TOP.
			//
			// The page only advances if the current page already carries
			// something. That single condition is what guarantees no page is
			// ever empty: an item far below everything placed so far starts
			// the CURRENT page's window rather than leaving blank pages
			// between, and the very first item of a document never pushes a
			// blank page 0 ahead of itself.
			if pageHasItem {
				page++
				pages = append(pages, PageAssignment{})
				pageHasItem = false
			}
			windowStart = it.Top
			pages[page].Shift = windowStart - contentTop
		}

		// The residual case, and the ONLY one left once the window can
		// slide: the item is taller than the window itself, so no window of
		// any position contains it. FR44's located diagnostic.
		if itemHeight := it.Bottom - it.Top; itemHeight > height {
			kind := "line"
			if len(it.Images) > 0 {
				kind = "image"
			}
			return Pagination{}, &OverflowError{
				ElementID:     it.ElementID,
				ItemHeight:    itemHeight,
				ContentHeight: height,
				Kind:          kind,
			}
		}

		pageOf[idx] = page
		pageHasItem = true
	}

	// Emission, in AUTHORED order per page. See PageAssignment.ContentRuns.
	for i := range items {
		p := pageOf[i]
		pages[p].ContentRuns = append(pages[p].ContentRuns, items[i].Runs...)
		pages[p].ContentImages = append(pages[p].ContentImages, items[i].Images...)
	}

	return Pagination{Pages: pages}, nil
}
