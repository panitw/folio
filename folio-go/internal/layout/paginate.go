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
// ONE EXCEPTION, AND IT IS RULED RATHER THAN CHOSEN (Story 4.6, AD-14,
// D-4.6.2). A GROUP (Story 4.3's ItemGroup — a table row) taller than the
// window is not a template error: AD-14 says over-tall rows are Warnings
// returned alongside PDF bytes, never fatal. Such a group is given a page
// of its own and CLIPPED at that page's content bottom, recorded in
// Pagination.Clipped. Rules 1-4 are untouched by it: the column is not
// mutated, no line is split (whole lines are dropped, never halves), no
// sibling moves, and the exception is scoped to grouped items because a
// row's height comes from DATA while a font size and an image box come
// from the author's own keyboard.
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
	"fmt"
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

	// Runs, Images and Rects are the item's content. Exactly one is
	// populated: an item is a line, an image, OR a table's vector chrome,
	// never more than one of the three — a line is atomic by rule 3, an
	// image by rule 4, and a table's header rects (Story 4.1) are atomic
	// for the same reason a header row does not split within this story
	// (AC3/AC9). A table's row LABELS/CELLS are separate items (kind
	// "line", built the same way any text is): Story 4.1 gave the header
	// row and its label the SAME extent so the two land on the same page
	// as a side effect, with no fourth exclusivity case; Story 4.3's
	// Group field (below) is the GENERAL mechanism that keeps a row's
	// chrome and every one of its (possibly several) physical lines
	// together, whether or not their extents happen to coincide.
	Runs   []TextRunRef
	Images []ImageRef
	Rects  []RectRef

	// Group is Story 4.3's grouping concept (DECISION-1), ORTHOGONAL to
	// the Runs/Images/Rects exclusivity check above (R1): it is a
	// SEPARATE statement about which items must land on the same page,
	// never a relaxation of "exactly one kind per item". The zero value,
	// ItemGroup{}, means "not grouped" (Present is false), so every item
	// built before this story — and every item this story's callers do
	// not themselves tag — is completely unaffected (R2).
	//
	// The identity is copied VERBATIM from the row-generating code's own
	// output (package folio's tableRectSource/textRunSource), never
	// reconstructed from ElementID, extent or emission order (D-4.2.2,
	// R3). A table's HEADER row is one group and each DATA row is its
	// own group — "a row", not "a run of rows" (4.5 is a different rule
	// and stays out, R7's own scoping).
	Group ItemGroup
}

// ItemGroup names the set of ColumnItems that Paginate must place on one
// page TOGETHER (Story 4.3, DECISION-1): if the window does not entirely
// contain the group's FULL extent (the union of every member's Top..Bottom,
// resolved in the code comment "Things the schema and record could not
// resolve" — the union rather than any one member's own chrome rect,
// because the two are indistinguishable at this commit but only the union
// stays correct once they diverge), the window slides to the group's
// EARLIEST Top and every member is re-tested against it — so a group
// behaves, for placement purposes, like one atomic item spanning from its
// earliest Top to its latest Bottom, without becoming one ColumnItem and
// without touching MixedItemError (R1).
type ItemGroup struct {
	// Present distinguishes "grouped" from the zero value: an ungrouped
	// item (Present == false) is placed exactly as before this story,
	// regardless of what Key holds.
	Present bool

	// Key names one group. Two grouped items with equal Key belong to the
	// same group. It is a plain comparable struct — never fed to a map
	// whose ITERATION would reach output order (R5); Paginate looks keys
	// up, it never ranges over a map of them to decide emission order.
	Key ItemGroupKey
}

// ItemGroupKey is one group's identity: a table element scopes it to one
// table (two different tables' "row 0" never collide), and IsHeader/Index
// scope it to one row within that table's own rows. Index is meaningless
// (and always the zero value) when IsHeader is true — the header is one
// row, not an indexed member of a family of headers.
type ItemGroupKey struct {
	ElementID string
	IsHeader  bool
	Index     int
}

// TextRunRef and ImageRef index the caller's own run/placement slices rather
// than carrying copies. Pagination decides WHICH PAGE and BY HOW MUCH TO
// SHIFT; it does not rebuild content, so it never has an opportunity to
// reorder or drop a run.
type TextRunRef int

// ImageRef indexes the caller's placement slice, for the same reason.
type ImageRef int

// RectRef indexes the caller's pagemodel.Rect slice (Story 4.1), for the
// same reason ImageRef indexes the caller's placement slice: this
// package never constructs or inspects a Rect's own fields, so it names
// content only by position in a slice the caller owns.
type RectRef int

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
	Rects  []RectRef
}

// OverflowError is FR44's located diagnostic for an item that fits in NO
// window — the residual case once the window can slide, and therefore
// exactly "the item is taller than the content window".
//
// ONE OVERFLOW RULE, TWO SUBJECTS — AND, SINCE STORY 4.6, ONE SUBJECT ONLY.
// D-2.6.1 ruled the image case: an image whose declared box exceeds the
// content window "fits nowhere", and that is a TEMPLATE error with a located
// message rather than a render-time surprise. A LINE taller than the window
// (reachable with a font size larger than the content band) is the same case
// and gets the same answer, so the implementation is never left to split,
// loop or panic. Both of those subjects are UNGROUPED items, and both keep
// this error verbatim.
//
// What is NO LONGER a subject of this error is a GROUPED item — a table row
// (Story 4.3's ItemGroup). AD-14 rules that "over-tall rows (FR25) and
// clipped content (FR44) are Warnings returned alongside PDF bytes, never
// silent and never fatal", and Story 4.6 brought the code to that spine: an
// over-tall GROUP is clipped to a fresh page and recorded in
// Pagination.Clipped, from which package folio builds a Warning
// (TABLE_ROW_CLIPPED_HEIGHT). The distinction is AD-13's own line and D-4.6.2's
// ruling: a row's height is DERIVED FROM DATA the author may never have seen,
// while an image's declared box and a font size are things the author TYPED.
// A failure derived from unauditable data must not be fatal; a typo should be.
//
// NEVER A STRADDLE AND NEVER A SILENT CLIP: both are what this error exists
// to prevent, and Story 4.6's clip weakens neither. It is not a straddle (it
// drops WHOLE lines and truncates a chrome rectangle's coordinate — no line
// is ever drawn in halves, AD-24) and it is not silent (every clip carries a
// located Warning on Result.Diagnostics). "Never a silent clip" was always
// the operative word; the clip that arrived is loud.
//
// REACHABILITY OF Kind, recorded so a future reader does not treat a dead
// branch as live (Story 4.6, AC8). Kind == "table" is produced only for an
// item carrying Rects, and every tableRectSource in package folio is built at
// exactly one of three sites (table_render.go's header, data-row and footer
// constructors), all three of which carry a PRESENT ItemGroup — so every item
// that could set Kind "table" now takes the clip branch instead. Kind ==
// "table" is therefore NO LONGER PRODUCED FROM PACKAGE FOLIO. It remains
// constructible by a direct caller of this package that builds an ungrouped
// Rects item, which is why the derivation is kept rather than deleted.
type OverflowError struct {
	ElementID     string
	ItemHeight    geom.Length
	ContentHeight geom.Length
	Kind          string // "line", "image" or "table" (Story 4.1)
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
		return "element " + e.ElementID + ": a column item carries neither a text run, an image nor a " +
			"table rect, so it would occupy column space while drawing nothing — an invisible item can " +
			"push visible content onto another page"
	}
	return "element " + e.ElementID + ": a column item carries MORE THAN ONE of {text runs, an image, " +
		"table rects}. Each is atomic for a different reason and an overflow diagnostic about this " +
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
//
// Story 4.4 (FR26) adds a SECOND, NARROWER displacement channel —
// PageAssignment.HeaderRepeats and RowDisplacement — that is scoped to one
// table on one page and is layered ON TOP OF Shift, never a redefinition of
// it (DECISION-3, ruled): Shift keeps exactly its pre-4.4 meaning, and an
// element that is not part of a repeating table is positioned by Shift
// alone, exactly as before this story.
type Pagination struct {
	Pages []PageAssignment

	// Suppressed records every (table, page) where FR26's repeat could
	// not be honoured (Story 4.4, DECISION-2): the next unplaced row of
	// that table fit the BARE window on that page but not the window
	// under the header's own reserved height. The repeat is silently
	// absent from that ONE page only, so pagination always terminates
	// (AC6) rather than looping or turning a document that renders
	// today into a hard error.
	//
	// This is DATA for a caller to build a located Warning from — the
	// row height and the space it fit inside are carried here, straight
	// from the computation that decided the suppression, so a caller
	// never re-derives the fit arithmetic a second time (D-4.2.2).
	// Appended in the sweep's own deterministic order (a slice, never a
	// map ranged for emission — R5).
	Suppressed []TableHeaderSuppressed

	// Clipped records every GROUP that was taller than the content
	// window and was therefore placed alone on a fresh page and cut off
	// at that page's content bottom (Story 4.6, FR25/AD-14). One entry
	// per clipped group, appended in the sweep's own deterministic order
	// (a slice, never a map ranged for emission — R5).
	//
	// This is DATA for a caller to build a located Warning from, exactly
	// as Suppressed is: the group's identity, its own height and the
	// height it was measured against are carried here straight from the
	// comparison that decided the clip, so a caller never re-derives the
	// fit arithmetic a second time (D-4.2.2). The MESSAGE is built in
	// package folio — this package has no author-facing prose in it and
	// does not know a footer sentinel from a row index.
	//
	// An UNGROUPED over-tall item is not here: it still returns
	// *OverflowError, per D-4.6.2 and OverflowError's own doc comment.
	Clipped []TableRowClipped
}

// TableRowClipped is Pagination.Clipped's element type — one group that
// did not fit any window and was cut off (Story 4.6).
type TableRowClipped struct {
	// ElementID names the element the clipped group belongs to — the
	// TABLE's id for a table row, which is what the group's own members
	// carry.
	ElementID string

	// Key is the group's identity, carried VERBATIM rather than
	// re-derived (D-4.2.2), so a caller reads the row index and the
	// header/footer role straight off the group that was clipped. This
	// package attaches no meaning to Index's value beyond equality —
	// the footer's sentinel is package folio's own convention.
	Key ItemGroupKey

	// ItemHeight is the group's UNION height (latest Bottom minus
	// earliest Top) — the quantity that exceeded the window, not any one
	// member's own height.
	ItemHeight geom.Length

	// ContentHeight is the content window's height it was measured
	// against. A caller's message can state both directly.
	ContentHeight geom.Length

	// Page is the page the clipped group landed on, 0-based (a caller
	// adds one to speak to a human, exactly as the Suppressed loop does).
	Page int
}

// RectClip truncates ONE rect's bottom edge on ONE page (Story 4.6): the
// caller draws the rect it already owns, but no lower than Bottom.
//
// It is a COORDINATE, in the same page-absolute column space every other
// extent in this package uses, and it is a BOUND rather than a height:
// the caller applies `min(rect.Bottom, Bottom)` against the rect's own
// geometry, which this package never inspects (RectRef's whole contract).
// That keeps a group of several rects with different extents correct with
// one number, and keeps AD-5 intact — no PDF clip path is implied or
// needed, because truncating a rectangle is a change to a rectangle.
//
// Apply it BEFORE PageAssignment.Shift: Bottom is a column coordinate,
// like Top/Bottom on a ColumnItem, and Shift is what moves column space
// into page space.
type RectClip struct {
	Ref    RectRef
	Bottom geom.Length
}

// TableHeaderRepeat is one continuation page's redrawn copy of a table's
// header row (Story 4.4, FR26): the SAME Rects/Runs the table's own header
// ColumnItem already carries (R6 — no new glyphs, no second producer),
// positioned immediately above that table's own first row on THIS page
// (DECISION-3, ruled: never at the page's own top — that would displace
// unrelated siblings, which AD-24 forbids).
type TableHeaderRepeat struct {
	// ElementID names the table this repeat belongs to.
	ElementID string

	// Rects/Runs are the header's OWN content refs, copied verbatim from
	// the header's ColumnItem — index into the SAME caller-owned slices
	// PageAssignment.ContentRects/ContentRuns already index into.
	Rects []RectRef
	Runs  []TextRunRef

	// Shift is subtracted from each ref's own page-absolute Y (the same
	// convention PageAssignment.Shift uses) to place the repeat on this
	// page. It is a SEPARATE quantity from PageAssignment.Shift — the
	// page's own Shift is untouched and continues to govern every other
	// element on the page (DECISION-3).
	Shift geom.Length
}

// TableRowDisplacement is the EXTRA downward displacement (Story 4.4),
// beyond PageAssignment.Shift, applied to one table's OWN rows on one page
// to make room for that table's repeated header immediately above them.
// Scoped by ElementID: no other element's position is touched (R2 — the
// column itself is never mutated; DECISION-3 — displacement is per-table,
// never page-wide).
type TableRowDisplacement struct {
	ElementID string
	Amount    geom.Length
}

// TableHeaderSuppressed is Pagination.Suppressed's element type — see its
// doc comment.
type TableHeaderSuppressed struct {
	// ElementID names the table whose repeat was suppressed on Page.
	ElementID string
	Page      int

	// RowHeight is the row that forced the decision — its own height,
	// which fits Available but not Available minus the header's height.
	RowHeight geom.Length

	// Available is the space the row ACTUALLY had to fit inside WITHOUT
	// the reservation — `windowStart + height - effectiveTop`, the room
	// remaining from wherever this row's own window happened to land,
	// not the page's bare content height (finisher fix, Story 4.4
	// review Finding 8/Major: the two coincide only when the window
	// slid to this row's own top; when an EARLIER element on the page
	// slid the window first — reachable, see
	// TestReservedHeaderSuppressedDiagnosticReportsTheRoomTheRowActuallyHad
	// — the row's real headroom is strictly less than the bare content
	// height, and reporting the bare height overstated the room the
	// author actually had to work with, D-000.37's "executable by a
	// human" requirement). A caller's diagnostic message can state
	// RowHeight and Available directly, without re-deriving either.
	Available geom.Length

	// HeaderHeight is the table's own reserved header height on this
	// page — carried so a caller's message can name the "reduce
	// headerHeight" lever WITH a number, rather than leaving the author
	// to go measure it themselves (finisher fix, Finding 8).
	HeaderHeight geom.Length
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
	//
	// Story 4.4, DECISION-3 (ruled): this field's meaning is UNCHANGED by
	// FR26. A repeated table header is never folded into Shift — see
	// HeaderRepeats/RowDisplacement below, a second and narrower channel.
	Shift geom.Length

	// ContentRuns/ContentImages are the CONTENT band's items on this page,
	// in the caller's ORIGINAL ORDER, not in the order the sweep visited
	// them. Emission order reaches output bytes (object and operator order),
	// and the sweep visits in COLUMN order, which is not necessarily
	// authored order. Preserving authored order is what makes a one-page
	// document byte-identical to its pre-2.6 self.
	ContentRuns   []TextRunRef
	ContentImages []ImageRef
	ContentRects  []RectRef

	// HeaderRepeats — Story 4.4, FR26: this page's copy of a table's
	// header, one entry per table continuing onto this page under an
	// honoured reservation. Empty for a page that is not a continuation
	// of any table, and always empty for the page carrying that table's
	// OWN header (DECISION-1: that page's header is not a "repeat").
	// Appended in the sweep's own deterministic order.
	HeaderRepeats []TableHeaderRepeat

	// RowDisplacement — Story 4.4: the extra downward displacement
	// (beyond Shift) this page applies to a table's own rows, keyed by
	// ElementID, when that table's header repeats on this page. A slice,
	// walked by a caller in declared order — never a map ranged for
	// emission (R5).
	RowDisplacement []TableRowDisplacement

	// ClippedRects — Story 4.6: the rects on this page whose bottom edge
	// must be truncated because the group they belong to was taller than
	// the window. Empty for every page that carries no clipped group,
	// which is every page of every document that has none. Appended in
	// the sweep's own deterministic order.
	ClippedRects []RectClip
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
		populated := 0
		if len(it.Runs) > 0 {
			populated++
		}
		if len(it.Images) > 0 {
			populated++
		}
		if len(it.Rects) > 0 {
			populated++
		}
		if populated > 1 {
			return Pagination{}, &MixedItemError{ElementID: it.ElementID}
		}
		if populated == 0 {
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

	// GROUPING (Story 4.3, DECISION-1), computed ONCE here rather than
	// re-derived per item in the sweep below.
	//
	// groupExtent is one group's UNION extent across every member
	// (earliest Top, latest Bottom) — the "Things the schema and record
	// could not resolve" note's choice, taken because it stays correct
	// once a group's members no longer share one rect's own bounds
	// (4.5/4.8 are both about to add items into a row's span). Today the
	// union and a data row's own chrome rect ARE equal by construction
	// (table_render.go: rowBottom = last line's bottom + padBottom), so
	// no fixture at this commit can distinguish the two choices — this
	// picks the one that keeps working when that stops being true.
	//
	// R7's ORIGINAL PREMISE — "a group's members are contiguous in column
	// order by construction" — is FALSE for real templates (this story's
	// own finisher review, Finding 1, measured against a legal document:
	// a caption, or a second table, sharing a row's own Top sorts BETWEEN
	// two of that row's members on the stable sort's tie order, which is
	// ordinary authoring, not a caller bug). Enforcing contiguity as a
	// render-time internal error therefore rejected legal input.
	//
	// The fix removes the DEPENDENCE on contiguity instead of asserting a
	// premise that does not hold. This pre-pass needs only each group's
	// UNION EXTENT, a property of the group's own members regardless of
	// what else the sweep visits between them, so it is computed by one
	// scan over `items` with no ordering requirement at all. It is the
	// SWEEP below (not this pre-pass) that then makes interleaving
	// harmless: a group's page is resolved ONCE, at whichever member the
	// column-order sweep visits FIRST, and every later-visited member of
	// the same group is assigned that SAME page directly, without
	// re-running the fit test — so an intervening item (grouped or not)
	// that ties a member's Top and sorts between two members can advance
	// the window/page for ITSELF without ever being able to split the
	// group.
	type groupExtent struct {
		top, bottom geom.Length
	}
	groups := make(map[ItemGroupKey]*groupExtent)
	for i := range items {
		g := items[i].Group
		if !g.Present {
			continue
		}
		e, ok := groups[g.Key]
		if !ok {
			groups[g.Key] = &groupExtent{top: items[i].Top, bottom: items[i].Bottom}
			continue
		}
		if items[i].Top < e.top {
			e.top = items[i].Top
		}
		if items[i].Bottom > e.bottom {
			e.bottom = items[i].Bottom
		}
	}

	// Story 4.4 (FR26): a table's repeat candidacy and its header's
	// height are both read from the SAME `groups` pre-pass above by
	// DIRECT LOOKUP — headerExtent below is a thin helper for
	// `groups[ItemGroupKey{ElementID: t, IsHeader: true}]`, never a
	// second, independent derivation (R3/D-4.2.2) and never a RANGE over
	// `groups` (R5/D-1.3.5 total ban — a lookup is fine, an iteration is
	// not). FR26 is unconditional (no schema opt-out — "Things the
	// schema and the record could not resolve" note 1), so no
	// caller-supplied list of which tables repeat is needed at all: a
	// table's header group existing among `items` IS the authorization
	// to repeat it. A page-header/page-footer band's table never
	// reaches Paginate as a ColumnItem (it is repeated verbatim via
	// BandContent instead, AC5), so this can never affect it.
	headerExtent := func(table string) (*groupExtent, bool) {
		e, ok := groups[ItemGroupKey{ElementID: table, IsHeader: true}]
		return e, ok
	}
	// headerPageOf[T] is the page table T's OWN header landed on —
	// populated by the sweep below, the moment that page is decided.
	// Consulted only by direct lookup, never ranged for emission (R5).
	headerPageOf := make(map[string]int)

	// tablePageKey names one (table, page) pair — Story 4.4's decision
	// granularity for whether a repeat is honoured there (DECISION-2/3).
	type tablePageKey struct {
		table string
		page  int
	}
	// reservation[key] is the reserved height ADOPTED for that table on
	// that page — headerHeightOf[table] if the repeat is honoured there,
	// or 0 if DECISION-2's fallback applied (suppressed on that one
	// page only). Decided ONCE per key, the first time a row of that
	// table is resolved onto that page; every later row of the same
	// table on the same page reads the SAME decision, by direct lookup
	// — never re-derived (D-4.2.2).
	reservation := make(map[tablePageKey]geom.Length)
	var suppressed []TableHeaderSuppressed
	var clipped []TableRowClipped

	// dropped[i] marks an authored item whose RUNS AND IMAGES the Story
	// 4.6 clip DESTROYED: a line (or image) of an over-tall group whose
	// own extent lies past the content bottom of the one page that group
	// could be given. The emission below omits them — leaving a ref out
	// of a page's ContentRuns/ContentImages IS the drop, since those
	// slices are already a per-page SUBSET of the caller's own slices
	// (AD-5: this package decides pages, it never rebuilds content).
	//
	// It does NOT gate the item's Rects. A rectangle is a coordinate and
	// is TRUNCATED (ClippedRects) rather than dropped, so the row still
	// reads as a row; only a line, which is atomic, is destroyed whole.
	//
	// It is a []bool indexed by AUTHORED position, not a set, so the
	// emission loop reads it by direct index and no iteration order is
	// introduced (R5).
	dropped := make([]bool, len(items))

	// pageOf[i] is the page authored-item i landed on. Filled by the sweep,
	// read by the emission below — which is what lets the sweep run in column
	// order while emission runs in authored order.
	pageOf := make([]int, len(items))

	windowStart := contentTop
	page := 0
	pageHasItem := false

	// groupPage records, for each group the sweep has already resolved,
	// the page every one of its members lands on. It is what replaces
	// R7's contiguity requirement (Finding 1): a group's page is decided
	// ONCE, at whichever member the column-order sweep visits FIRST, and
	// every LATER-visited member of the same group — however far away in
	// column order, and whatever else the sweep visited in between — is
	// assigned that same page directly below, without re-running the fit
	// test. An intervening item (grouped or not) that happens to tie a
	// member's Top can advance the window/page for ITSELF; it can never
	// split the group, because the group's remaining members never ask
	// the window a second question.
	groupPage := make(map[ItemGroupKey]int, len(groups))

	for _, idx := range order {
		it := items[idx]

		if it.Group.Present {
			if p, ok := groupPage[it.Group.Key]; ok {
				// This group's page was already decided at an
				// earlier-visited member. Ride along on that page
				// without touching window state — that is precisely
				// what keeps an interleaved, non-contiguous member (or
				// an unrelated item sorting between two members) from
				// being able to move this item anywhere else.
				pageOf[idx] = p
				continue
			}
		}

		// effectiveTop/effectiveBottom are the quantity the fit test and
		// the overflow test use: the ITEM's own extent, unless it belongs
		// to a group, in which case it is the GROUP's union extent —
		// computed above with no ordering requirement on the group's
		// members. This branch runs only for the FIRST-visited member of
		// a group (or for an ungrouped item), by construction of the
		// short-circuit above.
		effectiveTop, effectiveBottom := it.Top, it.Bottom
		if it.Group.Present {
			e := groups[it.Group.Key]
			effectiveTop, effectiveBottom = e.top, e.bottom
		}

		// STORY 4.6 (FR25, AD-14): the group is taller than the window
		// itself, so no window of any position contains it. AD-14 rules
		// this NEVER FATAL — "over-tall rows (FR25) and clipped content
		// (FR44) are Warnings returned alongside PDF bytes" — so the
		// group is given a page of its own and cut off at that page's
		// content bottom, and the render carries on.
		//
		// This is tested BEFORE the fit test and the FR26 reservation
		// below, and it does its own page advance, because both of those
		// are answers to "which window does this fit in?" and this group
		// fits none: the fresh page is not a window decision the slide
		// happens to make, it is this branch's own first step, and it is
		// deletable on its own (D-000.85's per-observable screen).
		//
		// Keyed on Group.Present, and NOT on the item's kind: measured
		// at 45cf812, an over-tall table row and an over-tall plain text
		// element are indistinguishable at the public API (both 0 bytes,
		// both CONTENT_UNLAYOUTABLE, both Kind "line"), because the
		// shipped path appends text before rects and the row's first
		// LINE ties the chrome on Top. A branch keyed on Kind == "table"
		// would clip nothing (D-4.6.2).
		if it.Group.Present {
			if itemHeight := effectiveBottom - effectiveTop; itemHeight > height {
				// (1) A FRESH PAGE FOR THE GROUP. The same
				// no-empty-page condition the slide uses: a page that
				// carries nothing yet IS the fresh page.
				if pageHasItem {
					page++
					pages = append(pages, PageAssignment{})
					pageHasItem = false
				}
				windowStart = effectiveTop
				pages[page].Shift = windowStart - contentTop

				// (1b) THE REPEATED HEADER THIS PAGE STILL GETS
				// (FR26/DECISION-3 composed with the clip, D-4.6.4).
				//
				// This branch used to `continue` past Story 4.4's whole
				// DECISION-2/DECISION-3 block below, so a clipped row's
				// page silently lost its column headers. That was not a
				// missing RECORD — it was 4.4's REMEDY applied where
				// 4.4's TRIGGER never fired. DECISION-2 arm (c)
				// suppresses a repeat precisely when reserving the
				// header would leave no room for a row; here there IS a
				// row on the page. And the substance runs the same way:
				// FR26 exists so a continuation page stays readable, and
				// of every page in the document the one carrying a
				// TRUNCATED row is the one that can least afford to lose
				// the labels that say what its surviving cells mean.
				//
				// So the two rules COMPOSE rather than one
				// short-circuiting the other: repeat the header, then
				// clip the row into what is left below it. The
				// reservation narrows the cut by exactly the header's
				// own height, which is why `contentBottom` is computed
				// from `reserved` rather than from `height` alone.
				//
				// The composition terminates, and its floor is 4.4's
				// own: if reserving the header leaves room for not even
				// ONE of the row's lines, the repeat buys nothing and
				// costs a line, so DECISION-2 arm (c) fires on its own
				// terms — suppressed, and RECORDED through the very
				// TableHeaderSuppressed channel 4.4 built for it.
				reserved := geom.Length(0)
				if !it.Group.Key.IsHeader {
					// The table whose row this is, and whose header is
					// already placed on an EARLIER page — DECISION-2's
					// own condition, computed here because `table`
					// below is not yet in scope. An over-tall HEADER is
					// never its own repeat, hence the IsHeader guard.
					tbl := it.Group.Key.ElementID
					hdr, hasHeader := headerExtent(tbl)
					hp, headerPlaced := headerPageOf[tbl]
					if _, decided := reservation[tablePageKey{tbl, page}]; hasHeader && headerPlaced && hp != page && !decided {
						hh := hdr.bottom - hdr.top

						// Does even one of the row's own lines survive
						// the narrowed cut? A chrome rect is not a
						// line: it is truncated either way, so it can
						// never be the thing that justifies the
						// reservation.
						reservedBottom := windowStart + height - hh
						linesUnderReservation := 0
						for j := range items {
							if !items[j].Group.Present || items[j].Group.Key != it.Group.Key {
								continue
							}
							if len(items[j].Rects) == 0 && items[j].Bottom <= reservedBottom {
								linesUnderReservation++
							}
						}

						if linesUnderReservation > 0 {
							// DECISION-3, honoured on a clipped page.
							reserved = hh
							reservation[tablePageKey{tbl, page}] = hh
							hdrRects, hdrRuns := headerContentOf(items, tbl)
							pages[page].RowDisplacement = append(pages[page].RowDisplacement,
								TableRowDisplacement{ElementID: tbl, Amount: hh})
							pages[page].HeaderRepeats = append(pages[page].HeaderRepeats,
								TableHeaderRepeat{
									ElementID: tbl,
									Rects:     hdrRects,
									Runs:      hdrRuns,
									Shift:     hdr.top - effectiveTop + pages[page].Shift,
								})
						} else {
							// DECISION-2 arm (c), on its own terms and
							// through its own path. Recorded, never
							// silent.
							reservation[tablePageKey{tbl, page}] = 0
							suppressed = append(suppressed, TableHeaderSuppressed{
								ElementID:    tbl,
								Page:         page,
								RowHeight:    itemHeight,
								Available:    height,
								HeaderHeight: hh,
							})
						}
					}
				}

				// (2) THE CUT. Everything at or above contentBottom is
				// kept; everything crossing or below it is destroyed.
				// FIT ENTIRELY (rule 3) is unchanged and is exactly what
				// makes this a clip rather than a straddle: a line whose
				// extent CROSSES the bottom is dropped whole, never drawn
				// in halves. A chrome rect is not a line — it is a
				// coordinate, and it is truncated (RectClip) rather than
				// dropped, so the row still reads as a row.
				contentBottom := windowStart + height - reserved
				for j := range items {
					if !items[j].Group.Present || items[j].Group.Key != it.Group.Key {
						continue
					}
					pageOf[j] = page
					if items[j].Bottom <= contentBottom {
						continue
					}
					// A MEMBER'S RECTS AND ITS RUNS ARE ANSWERED
					// INDEPENDENTLY, never as an either/or. Each kind
					// gets the treatment its own nature admits: a
					// rectangle is a coordinate, so it is TRUNCATED; a
					// line is atomic, so it is DROPPED WHOLE (AD-24
					// forbids drawing half of one).
					//
					// This used to short-circuit — `if len(Rects) > 0 {
					// …; continue }` — so a member carrying BOTH had its
					// rects bounded and its runs KEPT, drawn past the
					// content bottom. Content retained silently, on the
					// one code path whose entire job is destroying
					// content deliberately. Unreachable today
					// (paginateDocument and itemsForTest both build
					// rect-items and run-items disjointly), and handling
					// the two kinds separately is cheaper than asserting
					// the disjointness the old shape assumed (this
					// story's reviewer, Finding 11).
					for _, ref := range items[j].Rects {
						pages[page].ClippedRects = append(pages[page].ClippedRects,
							RectClip{Ref: ref, Bottom: contentBottom})
					}
					dropped[j] = true
				}

				// (3) RECORDED, NEVER SILENT. The caller turns this into
				// a located Warning; this package builds no prose.
				clipped = append(clipped, TableRowClipped{
					ElementID:     it.ElementID,
					Key:           it.Group.Key,
					ItemHeight:    itemHeight,
					ContentHeight: height,
					Page:          page,
				})

				// (4) PLACED. Marking the group resolved is what makes
				// this a single forward pass: every later-visited member
				// takes the groupPage short-circuit at the top of the
				// loop and never asks the window a second question.
				pageHasItem = true
				groupPage[it.Group.Key] = page
				if it.Group.Key.IsHeader {
					headerPageOf[it.Group.Key.ElementID] = page
				}
				continue
			}
		}

		// Story 4.4: is this item one of table T's OWN ROWS (never the
		// header itself) whose header this sweep has already resolved?
		// table == "" means "not applicable" throughout what follows.
		table := ""
		if it.Group.Present && !it.Group.Key.IsHeader {
			if _, ok := headerExtent(it.Group.Key.ElementID); ok {
				table = it.Group.Key.ElementID
			}
		}

		// ceilingFor is the usable window top edge for THIS item on page
		// p — height reduced by whatever reservation already stands for
		// (table, p), or the RAW ceiling if none has been decided yet
		// (deciding happens below, once p is FINAL for this item, never
		// here: a page's reservation must never be decided against a
		// window this item is only tentatively being tested against).
		ceilingFor := func(p int, ws geom.Length) geom.Length {
			if table != "" {
				if amt, ok := reservation[tablePageKey{table, p}]; ok {
					return ws + height - amt
				}
			}
			return ws + height
		}

		// Does this item (or its group) fit ENTIRELY in the window as it
		// currently stands?
		if effectiveBottom > ceilingFor(page, windowStart) {
			// It does not. The window slides to begin at THIS ITEM'S TOP
			// — or, grouped, at the GROUP'S earliest Top, which is what
			// keeps a wrapped row's continuation lines from starting a
			// window that excludes the row's own first line.
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
			windowStart = effectiveTop
			pages[page].Shift = windowStart - contentTop
		}

		// Story 4.4 (FR26/DECISION-2/DECISION-3): p/windowStart are now
		// FINAL for this item. If it is table T's row and (T, p) has no
		// decision yet, decide it exactly once, against the window this
		// item actually landed in.
		if table != "" {
			key := tablePageKey{table, page}
			if _, decided := reservation[key]; !decided {
				if hp, known := headerPageOf[table]; known && hp != page {
					headerItem, _ := headerExtent(table)
					hh := headerItem.bottom - headerItem.top
					reservedCeiling := windowStart + height - hh
					if effectiveBottom <= reservedCeiling {
						// DECISION-3: honoured. This row (or group) fits
						// even with the header's height reserved above
						// it, so the repeat is drawn on this page,
						// immediately above the table's own first row
						// here — effectiveTop, not windowStart, per
						// DECISION-3's ruling that the repeat sits at
						// the top of the TABLE on that page, not the
						// top of the page.
						reservation[key] = hh
						hdrRects, hdrRuns := headerContentOf(items, table)
						pages[page].RowDisplacement = append(pages[page].RowDisplacement,
							TableRowDisplacement{ElementID: table, Amount: hh})
						pages[page].HeaderRepeats = append(pages[page].HeaderRepeats,
							TableHeaderRepeat{
								ElementID: table,
								Rects:     hdrRects,
								Runs:      hdrRuns,
								Shift:     headerItem.top - effectiveTop + pages[page].Shift,
							})
					} else {
						// DECISION-2, arm (c): even alone, this row does
						// not fit under the reservation. Suppress the
						// repeat on THIS page only and place the row
						// under the RAW ceiling instead (guaranteed to
						// fit: the row already passed the raw fit test
						// above, or the window was just set to its own
						// top, which always admits an item that is not
						// itself an overflow). Recorded, never silent.
						reservation[key] = 0
						suppressed = append(suppressed, TableHeaderSuppressed{
							ElementID:    table,
							Page:         page,
							RowHeight:    effectiveBottom - effectiveTop,
							Available:    windowStart + height - effectiveTop,
							HeaderHeight: hh,
						})
					}
				} else {
					// This IS table T's own header page (or the header
					// has not yet been placed — unreachable given the
					// sort order, since a header's Top always precedes
					// every one of its rows' Tops, but guarded rather
					// than assumed): never a repeat here (DECISION-1).
					reservation[key] = 0
				}
			}
		}

		// The residual case, and the ONLY one left once the window can
		// slide: the item is taller than the window itself, so no window
		// of any position contains it. FR44's located diagnostic.
		//
		// Story 4.6 narrowed this to UNGROUPED items only — a grouped
		// item took the clip branch above and never reaches here. What
		// remains is D-2.6.1's own two subjects: a line whose font size
		// exceeds the content band, and an image whose DECLARED BOX
		// does. Both are things the author TYPED, and both stay a
		// located template error rather than a silent truncation of a
		// picture (D-4.6.2, AD-13).
		if itemHeight := effectiveBottom - effectiveTop; itemHeight > height {
			kind := "line"
			if len(it.Images) > 0 {
				kind = "image"
			}
			if len(it.Rects) > 0 {
				kind = "table"
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
		if it.Group.Present {
			groupPage[it.Group.Key] = page
			if it.Group.Key.IsHeader {
				headerPageOf[it.Group.Key.ElementID] = page
			}
		}
	}

	// R6, belt and suspenders: every member of a group landed on the SAME
	// page. The short-circuit above makes this true BY CONSTRUCTION now
	// (a group's page is decided once and every other member copies it,
	// with no dependence on contiguity — Finding 1) — checked anyway,
	// once more, in a single forward pass over `order`, because a
	// caller-visible invariant this central deserves an assertion rather
	// than only an argument in a comment.
	for pos, idx := range order {
		g := items[idx].Group
		if !g.Present {
			continue
		}
		if p := groupPage[g.Key]; p != pageOf[idx] {
			return Pagination{}, fmt.Errorf(
				"layout: Paginate: internal error: group %+v split across pages %d and %d at column "+
					"position %d (element %q) — row atomicity failed despite the group-page short-circuit; "+
					"this should be unreachable",
				g.Key, p, pageOf[idx], pos, items[idx].ElementID)
		}
	}

	// Emission, in AUTHORED order per page. See PageAssignment.ContentRuns.
	//
	// Story 4.6: a dropped item's RUNS AND IMAGES contribute nothing to
	// any page. That is the whole of the drop — there is no second
	// channel and no marker left behind, because a page's content refs
	// are already a per-page subset of the caller's slices and an absent
	// ref is an absent line.
	//
	// Its RECTS still emit, because a rect is never dropped: the clip
	// bounds it (ClippedRects) and the caller applies that bound. The
	// distinction is inert for every input this module builds today —
	// nothing carries both kinds — and it is spelled out anyway so that
	// the day something does, a rectangle is truncated and a line is
	// destroyed, rather than whichever the branch order happened to pick
	// (this story's reviewer, Finding 11).
	for i := range items {
		p := pageOf[i]
		if !dropped[i] {
			pages[p].ContentRuns = append(pages[p].ContentRuns, items[i].Runs...)
			pages[p].ContentImages = append(pages[p].ContentImages, items[i].Images...)
		}
		pages[p].ContentRects = append(pages[p].ContentRects, items[i].Rects...)
	}

	return Pagination{Pages: pages, Suppressed: suppressed, Clipped: clipped}, nil
}

// headerContentOf returns table's header ColumnItem(s)' own Rects and Runs,
// found by a single SLICE WALK over items (never a map range, R5) and
// matched by DIRECT FIELD LOOKUP on Group.Key (never reconstructed from
// ElementID/extent/order, D-4.2.2). A table's header is built as exactly
// two ColumnItems sharing one Group.Key{ElementID: table, IsHeader: true}
// — one carrying the header's cell chrome (Rects), one carrying its column
// labels (Runs), collectBandTableRuns'/table_render.go's own shape — so at
// most one of each is ever found, but both are accumulated defensively
// rather than assumed.
func headerContentOf(items []ColumnItem, table string) (rects []RectRef, runs []TextRunRef) {
	key := ItemGroupKey{ElementID: table, IsHeader: true}
	for i := range items {
		if !items[i].Group.Present || items[i].Group.Key != key {
			continue
		}
		rects = append(rects, items[i].Rects...)
		runs = append(runs, items[i].Runs...)
	}
	return rects, runs
}
