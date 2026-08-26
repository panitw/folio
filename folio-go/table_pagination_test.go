package folio

// Story 4.3: the repository's first MULTI-PAGE table fixture (R8 lifts
// 4.2's own "every table fits on one page" fence, here and only here).
//
// These tests exercise the REAL row-generating pipeline
// (collectBandTableRuns/collectBandTextRuns, table_render.go) feeding
// layout.Paginate through the SAME chromeRowGroup()/lineRowGroup()
// direct-field-lookup methods paginateDocument/contentColumnItems use in
// production (render.go/page_number.go) — this file's own itemsForTest
// helper duplicates only the "index into the caller's own slice" wiring
// paginateDocument already has, never Paginate's placement logic, so a
// green result here is a statement about the SHIPPED grouping mechanism.

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/panitw/folio/folio-go/internal/layout"
)

// multiRowTableDoc is a small custom page size (R8's fixture, chosen
// small so a modest row count already crosses a page boundary) with one
// two-column table bound to `items[]`. footerPageCount, when true, adds
// a page-header/page-footer "Page {{page}} of {{pages}}" construct —
// AC3's own witness needs both passes to see the SAME table.
func multiRowTableDoc(footerPageCount bool) string {
	footerElements := `[]`
	if footerPageCount {
		// Finding 3 (this story's finisher review): "pf1" violated the
		// schema's own id pattern (^e[0-9a-z]+$) and never even parsed —
		// the true branch was dead AND broken. "e4" is free (the table's
		// own elements use e1/e2/e3) and nextId (5) already exceeds it.
		footerElements = `[{"id": "e4", "type": "text", "x": 0, "y": 0, "width": 180, "height": 8, "value": "Page {{page}} of {{pages}}", "style": {"fontFamily": "latin", "fontSize": 6}}]`
	}
	return fmt.Sprintf(`{
  "assets": {},
  "bands": {
    "content": {"elements": [
      {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
        "style": {"fontFamily": "latin", "fontSize": 8},
        "columns": [
          {"id": "e2", "label": "A", "width": 80, "bind": "{{row.a}}"},
          {"id": "e3", "label": "B", "width": 80, "bind": "{{row.b}}"}
        ]}
    ]},
    "pageFooter": {"elements": %s, "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"latin": ["Noto Sans"]},
  "locale": "en",
  "nextId": 5,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`, footerElements)
}

// tableRowJSON is one bound row: A/B carry an "RxW-" marker (this
// story's own row-identity witness pattern, matching
// TestDataRowIdentityIsConsistentAndDistinct) so a run's own SourceText
// names the row it claims, independent of rowIndex bookkeeping.
type tableRowJSON struct {
	A string `json:"a"`
	B string `json:"b"`
}

// multiRowTableData builds `rows` bound rows. wrapRow, if >= 0, gets a
// long value in column "a" so that row wraps to several physical lines
// (AC1's "at least one row wrapped" precondition) — narrow enough (80pt
// at 8pt Noto Sans) that eight ~5-letter marked words will not fit one
// line.
func multiRowTableData(rows, wrapRow int) string {
	items := make([]tableRowJSON, rows)
	for i := range items {
		marker := fmt.Sprintf("R%dW-", i)
		a := marker + "x"
		if i == wrapRow {
			a = marker + "Alpha " + marker + "Bravo " + marker + "Charlie " + marker + "Delta " + marker + "Echo " + marker + "Foxtrot " + marker + "Golf " + marker + "Hotel"
		}
		items[i] = tableRowJSON{A: a, B: marker + "b"}
	}
	b, err := json.Marshal(map[string]any{"items": items})
	if err != nil {
		panic(err) // test-fixture construction only; a marshal failure here is a test bug, not reachable through Render
	}
	return string(b)
}

// itemsForTest mirrors paginateDocument's/contentColumnItems' own
// "index into the caller's slice" wiring for a CONTENT-band-only table
// (this file's fixtures declare no page-header/page-footer table and no
// images), carrying each item's Group via the SAME
// chromeRowGroup()/lineRowGroup() methods production code calls. It is
// NOT a second implementation of Paginate's placement logic — only of
// the bookkeeping that turns collectBandTableRuns/collectBandTextRuns'
// output into layout.ColumnItems, which is what lets RectRef/TextRunRef
// be traced straight back to tableRects[i]/contentRuns[i] in the
// assertions below (unlike contentColumnItems' own PHASE-A rects, which
// deliberately carry dummy refs — see that function's own doc comment).
func itemsForTest(contentRuns []textRunSource, tableRects []tableRectSource) []layout.ColumnItem {
	var items []layout.ColumnItem
	for i, r := range tableRects {
		items = append(items, layout.ColumnItem{
			ElementID: r.elementID, Top: r.top, Bottom: r.bottom,
			Rects: []layout.RectRef{layout.RectRef(i)},
			Group: r.chromeRowGroup(),
		})
	}
	for i := 0; i < len(contentRuns); i++ {
		j := i
		item := layout.ColumnItem{
			ElementID: contentRuns[i].elementID,
			Top:       contentRuns[i].itemTop, Bottom: contentRuns[i].itemBottom,
			Group: contentRuns[i].lineRowGroup(),
		}
		for j < len(contentRuns) &&
			contentRuns[j].elementID == contentRuns[i].elementID &&
			contentRuns[j].lineIndex == contentRuns[i].lineIndex {
			item.Runs = append(item.Runs, layout.TextRunRef(j))
			j++
		}
		items = append(items, item)
		i = j - 1
	}
	return items
}

// paginateContentTableForTest collects a content-band table's runs/rects
// through the real production functions and paginates them, returning
// enough to assert row atomicity/order directly against tableRects/
// contentRuns' own identity fields.
func paginateContentTableForTest(t *testing.T, tplJSON, dataJSON string) (plan layout.Pagination, contentRuns []textRunSource, tableRects []tableRectSource) {
	t.Helper()
	tpl, err := ParseTemplate([]byte(tplJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	bands, err := documentBands(tpl)
	if err != nil {
		t.Fatalf("documentBands: %v", err)
	}
	data := mustDecodeData(t, dataJSON)
	params := mustDecodeParams(t)
	geometry, err := pageGeometryOf(tpl)
	if err != nil {
		t.Fatalf("pageGeometryOf: %v", err)
	}
	fc := testFormatContext()
	contentRuns, tableRects, _, err = collectBandTableRuns(tpl, bands, contentBandIndex, data, params, fc, testShippedFontSet(), newFontCache(), nil)
	if err != nil {
		t.Fatalf("collectBandTableRuns: %v", err)
	}
	textRuns, _, _, err := collectBandTextRuns(tpl, bands, contentBandIndex, data, params, testShippedFontSet(), newFontCache(), contentBandResolver, nil)
	if err != nil {
		t.Fatalf("collectBandTextRuns: %v", err)
	}
	contentRuns = append(textRuns, contentRuns...)
	plan, err = layout.Paginate(geometry, itemsForTest(contentRuns, tableRects))
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}
	return plan, contentRuns, tableRects
}

// TestDataRowNeverSplitAcrossPages is AC1: for every page, the set of
// rowIndex values among that page's CHROME items equals the set among
// its LINE items (D-000.68: set equality by identity, never a count),
// and every run bearing a row's own content marker lands on that same
// page (the independent content anchor).
func TestDataRowNeverSplitAcrossPages(t *testing.T) {
	const rows = 20
	const wrapRow = 5
	plan, contentRuns, tableRects := paginateContentTableForTest(t, multiRowTableDoc(false), multiRowTableData(rows, wrapRow))

	if len(plan.Pages) < 2 {
		t.Fatalf("presence precondition: fixture must paginate to >= 2 pages, got %d — widen the fixture", len(plan.Pages))
	}
	wrappedLines := 0
	for _, r := range contentRuns {
		if r.isTableRowLine && r.rowIndex == wrapRow {
			wrappedLines++
		}
	}
	if wrappedLines < 2 {
		t.Fatalf("presence precondition: row %d must wrap to >= 2 physical lines, got %d run(s) carrying its marker", wrapRow, wrappedLines)
	}

	markerOf := func(text string) (row int, ok bool) {
		for i := range rows {
			marker := fmt.Sprintf("R%dW-", i)
			if containsSubstring(text, marker) {
				return i, true
			}
		}
		return 0, false
	}

	// CORE invariant, survivable under mutation (a) (no chrome at all):
	// every LINE item bearing one rowIndex lands on exactly ONE page, and
	// the marker its own SourceText carries names that same row. This
	// does not consult tableRects at all, so removing a data row's
	// chrome (mutation (a)) cannot make it vacuous — with no chrome, the
	// row's LINES are the only thing left to hold together, and this is
	// the assertion that they still do.
	contentChecked := 0
	linePages := map[int]map[int]bool{} // rowIndex -> set of pages its lines appear on
	for p, pg := range plan.Pages {
		for _, ref := range pg.ContentRuns {
			r := contentRuns[ref]
			if !r.isTableRowLine {
				continue
			}
			if linePages[r.rowIndex] == nil {
				linePages[r.rowIndex] = map[int]bool{}
			}
			linePages[r.rowIndex][p] = true
			row, ok := markerOf(r.text)
			if !ok {
				t.Fatalf("run %q carries no row marker — test fixture defect", r.text)
			}
			if row != r.rowIndex {
				t.Errorf("run %q carries marker for row %d but rowIndex = %d", r.text, row, r.rowIndex)
			}
			contentChecked++
		}
	}
	if contentChecked == 0 {
		t.Fatal("presence precondition: no table-row-line runs were examined")
	}
	for row, pages := range linePages {
		if len(pages) != 1 {
			t.Errorf("row %d's own physical lines are spread across %d distinct pages — a row's lines must all land on ONE page", row, len(pages))
		}
	}

	// SECONDARY invariant, only meaningful when the chrome exists (the
	// UNMUTATED tree, and mutation (b)'s subject): the chrome's rowIndex
	// set on a page must equal the line rowIndex set on that SAME page —
	// D-000.68's set equality, never a count. Skipped (not merely
	// vacuous-true) for a row with no chrome at all, which is exactly
	// mutation (a)'s effect and exactly why the CORE check above exists
	// independently.
	chromePage := map[int]int{}
	for p, pg := range plan.Pages {
		for _, ref := range pg.ContentRects {
			r := tableRects[ref]
			if r.isDataRow {
				chromePage[r.rowIndex] = p
			}
		}
	}
	for row, pages := range linePages {
		cp, hasChrome := chromePage[row]
		if !hasChrome {
			continue
		}
		var lp int
		for pg := range pages {
			lp = pg
		}
		if cp != lp {
			t.Errorf("row %d: chrome lands on page %d but its lines land on page %d — chrome and lines must share a page (D-000.68)", row, cp, lp)
		}
	}
	for row := range chromePage {
		if _, hasLines := linePages[row]; !hasLines {
			t.Errorf("row %d has a chrome item but no line items at all", row)
		}
	}
}

// TestTableRowsAppearExactlyOnceInDataOrderAcrossPages is AC2: every row
// of the bound collection appears on EXACTLY ONE page (no duplication,
// no omission — asserted as two SEPARATE failures, D-000.79's own
// discipline), and page p's rows all precede page p+1's (a table's own
// rows are laid out strictly downward by construction — rowTop = the
// previous row's rowBottom, table_render.go — so within one table this
// is the row-loop order already; this test pins it as an assertion
// rather than an assumption).
func TestTableRowsAppearExactlyOnceInDataOrderAcrossPages(t *testing.T) {
	const rows = 20
	plan, _, tableRects := paginateContentTableForTest(t, multiRowTableDoc(false), multiRowTableData(rows, -1))
	if len(plan.Pages) < 2 {
		t.Fatalf("presence precondition: fixture must paginate to >= 2 pages, got %d", len(plan.Pages))
	}

	// Finding 8 (this story's finisher review, Nit): a `rowPage` slice
	// used to be recorded here and read back only by an `if ... { continue
	// }` as the LAST statement of the loop body — a no-op with no effect
	// on any assertion. Removed rather than wired into a real check: each
	// row's own page is already asserted more precisely below (the union
	// check, and the per-page ordering check further down), so recording
	// it a second time here would just be a second, unused derivation.
	seenCount := make([]int, rows)
	maxRowSeenOnPage := make([]int, len(plan.Pages))
	for i := range maxRowSeenOnPage {
		maxRowSeenOnPage[i] = -1
	}
	for p, pg := range plan.Pages {
		for _, ref := range pg.ContentRects {
			r := tableRects[ref]
			if !r.isDataRow {
				continue
			}
			seenCount[r.rowIndex]++
			if r.rowIndex > maxRowSeenOnPage[p] {
				maxRowSeenOnPage[p] = r.rowIndex
			}
		}
	}
	for row := range rows {
		if seenCount[row] == 0 {
			t.Errorf("row %d is OMITTED — appears on no page", row)
		}
		if seenCount[row] > 1 {
			t.Errorf("row %d is DUPLICATED — its chrome appears %d times", row, seenCount[row])
		}
	}
	// Union of rowIndex values across all pages is exactly {0..rows-1}.
	union := map[int]bool{}
	for _, pg := range plan.Pages {
		for _, ref := range pg.ContentRects {
			r := tableRects[ref]
			if r.isDataRow {
				union[r.rowIndex] = true
			}
		}
	}
	if len(union) != rows {
		t.Errorf("union of rowIndex across pages has %d members, want %d", len(union), rows)
	}
	// Order: page p's rows all precede page p+1's.
	for p := 0; p+1 < len(plan.Pages); p++ {
		if maxRowSeenOnPage[p] == -1 || maxRowSeenOnPage[p+1] == -1 {
			continue
		}
		minNext := rows
		for _, ref := range plan.Pages[p+1].ContentRects {
			r := tableRects[ref]
			if r.isDataRow && r.rowIndex < minNext {
				minNext = r.rowIndex
			}
		}
		if maxRowSeenOnPage[p] >= minNext {
			t.Errorf("page %d's highest rowIndex (%d) is >= page %d's lowest (%d) — rows are not in the collection's order across pages", p, maxRowSeenOnPage[p], p+1, minNext)
		}
	}
}

// TestBothPaginationPassesAgreeOnRowPartition is AC3: PHASE A
// (contentColumnItems, page_number.go) and PHASE B (paginateDocument's
// own item order, render.go) must produce the SAME page count and the
// SAME per-page rowIndex partition for one table, even though the two
// builders append rects/text in different orders (D3) — because both
// now carry the SAME Group identity (chromeRowGroup/lineRowGroup),
// direct field lookup, on the SAME underlying tableRects/contentRuns.
func TestBothPaginationPassesAgreeOnRowPartition(t *testing.T) {
	const rows = 20
	tplJSON := multiRowTableDoc(false)
	dataJSON := multiRowTableData(rows, 5)

	tpl, err := ParseTemplate([]byte(tplJSON))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	bands, err := documentBands(tpl)
	if err != nil {
		t.Fatalf("documentBands: %v", err)
	}
	data := mustDecodeData(t, dataJSON)
	params := mustDecodeParams(t)
	geometry, err := pageGeometryOf(tpl)
	if err != nil {
		t.Fatalf("pageGeometryOf: %v", err)
	}
	fc := testFormatContext()
	visible, err := computeVisibility(bands, data, params, fc)
	if err != nil {
		t.Fatalf("computeVisibility: %v", err)
	}
	imageRuns, err := collectImageRuns(tpl)
	if err != nil {
		t.Fatalf("collectImageRuns: %v", err)
	}
	contentTableRuns, contentTableRects, _, err := collectBandTableRuns(tpl, bands, contentBandIndex, data, params, fc, testShippedFontSet(), newFontCache(), visible)
	if err != nil {
		t.Fatalf("collectBandTableRuns: %v", err)
	}
	contentRuns, _, _, err := collectBandTextRuns(tpl, bands, contentBandIndex, data, params, testShippedFontSet(), newFontCache(), contentBandResolver, visible)
	if err != nil {
		t.Fatalf("collectBandTextRuns: %v", err)
	}
	// PHASE A's OWN order (page_number.go's predictDocument caller):
	// this band's text runs, THEN this band's table runs.
	contentRuns = append(contentRuns, contentTableRuns...)

	// PHASE A: the real production function.
	phaseAPlan, err := layout.Paginate(geometry, contentColumnItems(contentRuns, imageRuns, contentTableRects, visible))
	if err != nil {
		t.Fatalf("PHASE A Paginate: %v", err)
	}

	// PHASE B: paginateDocument's OWN item order (render.go) — rects
	// first, then text — reproduced here with the SAME contentRuns/
	// contentTableRects slices phase A used, so any divergence is about
	// ORDER/GROUPING alone, not about different underlying data.
	var phaseBItems []layout.ColumnItem
	for i, r := range contentTableRects {
		phaseBItems = append(phaseBItems, layout.ColumnItem{
			ElementID: r.elementID, Top: r.top, Bottom: r.bottom,
			Rects: []layout.RectRef{layout.RectRef(i)},
			Group: r.chromeRowGroup(),
		})
	}
	for i := 0; i < len(contentRuns); i++ {
		j := i
		item := layout.ColumnItem{
			ElementID: contentRuns[i].elementID,
			Top:       contentRuns[i].itemTop, Bottom: contentRuns[i].itemBottom,
			Group: contentRuns[i].lineRowGroup(),
		}
		for j < len(contentRuns) &&
			contentRuns[j].elementID == contentRuns[i].elementID &&
			contentRuns[j].lineIndex == contentRuns[i].lineIndex {
			item.Runs = append(item.Runs, layout.TextRunRef(j))
			j++
		}
		phaseBItems = append(phaseBItems, item)
		i = j - 1
	}
	phaseBPlan, err := layout.Paginate(geometry, phaseBItems)
	if err != nil {
		t.Fatalf("PHASE B Paginate: %v", err)
	}

	if len(phaseAPlan.Pages) != len(phaseBPlan.Pages) {
		t.Fatalf("PHASE A produced %d pages, PHASE B produced %d — {{page}}/{{pages}} (D-2.7.2) would print the WRONG total", len(phaseAPlan.Pages), len(phaseBPlan.Pages))
	}
	if len(phaseAPlan.Pages) < 2 {
		t.Fatalf("presence precondition: fixture must paginate to >= 2 pages, got %d", len(phaseAPlan.Pages))
	}

	// D-000.68: SETS of pages per row, per phase — not a single
	// "last-write-wins" page number. A row whose lines are SPLIT across
	// two pages within one phase would otherwise silently report
	// whichever page its LAST line happened to land on, which can
	// coincide with the other phase's (correct) page purely by chance
	// and mask exactly the divergence this test exists to catch.
	rowPagesPhaseA := make([]map[int]bool, rows)
	rowPagesPhaseB := make([]map[int]bool, rows)
	for i := range rowPagesPhaseA {
		rowPagesPhaseA[i] = map[int]bool{}
		rowPagesPhaseB[i] = map[int]bool{}
	}
	for p, pg := range phaseAPlan.Pages {
		for _, ref := range pg.ContentRuns {
			r := contentRuns[ref]
			if r.isTableRowLine {
				rowPagesPhaseA[r.rowIndex][p] = true
			}
		}
	}
	for p, pg := range phaseBPlan.Pages {
		for _, ref := range pg.ContentRuns {
			r := contentRuns[ref]
			if r.isTableRowLine {
				rowPagesPhaseB[r.rowIndex][p] = true
			}
		}
	}
	for row := range rows {
		if len(rowPagesPhaseA[row]) != 1 {
			t.Errorf("row %d: PHASE A itself spreads this row's lines across %d distinct pages (%v) — PHASE A must keep a row whole too, or its own page count is unreliable", row, len(rowPagesPhaseA[row]), rowPagesPhaseA[row])
		}
		if len(rowPagesPhaseB[row]) != 1 {
			t.Errorf("row %d: PHASE B itself spreads this row's lines across %d distinct pages (%v)", row, len(rowPagesPhaseB[row]), rowPagesPhaseB[row])
		}
		if len(rowPagesPhaseA[row]) == 1 && len(rowPagesPhaseB[row]) == 1 {
			var pa, pb int
			for p := range rowPagesPhaseA[row] {
				pa = p
			}
			for p := range rowPagesPhaseB[row] {
				pb = p
			}
			if pa != pb {
				t.Errorf("row %d: PHASE A assigns page %d, PHASE B assigns page %d — the two passes disagree on the partition", row, pa, pb)
			}
		}
	}
}

// TestRowTallerThanContentWindowStillReturnsLocatedTableOverflow is AC4,
// scoped EXPLICITLY to itemsForTest's own item order (PHASE B: rects
// before text — Finding 4, this story's finisher review): a row (its own
// chrome+lines UNION) taller than the whole content window fails EXACTLY
// as at 903bf8f on THIS ordering — a located *layout.OverflowError, Kind
// "table" (the chrome rect, visited first under PHASE B's order, sets
// Kind) — and the call RETURNS (it does not hang). Story 4.6 owns the
// real answer (clip to a fresh page, report a Diagnostic); THIS test
// records CURRENT behaviour and does not endorse it.
//
// This is deliberately NOT the ordering the shipped path uses — see
// TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender
// immediately below, which pins the OTHER arm (PHASE A, via the public
// Render(), Kind "line") that Finding 4 found unmeasured. Do not leave
// one arm pinned and the other unmeasured — both are recorded, and 4.6
// is told which one its own users will actually see.
// tooTallRowDoc's fontSize 200pt makes a single physical line's own
// height alone exceed the ~100,000mp content window this fixture's
// 150pt-tall page leaves (150-10-10-10-10 = 110pt content height): the
// row is unplaceable in ANY window, the residual case FR44 exists for.
// Shared by both AC4 tests below (PHASE B via itemsForTest, PHASE A via
// the public Render()) so the two arms are pinned against the SAME
// document, not two documents that merely resemble each other.
func tooTallRowDoc() string {
	return `{
  "assets": {},
  "bands": {
    "content": {"elements": [
      {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
        "style": {"fontFamily": "latin", "fontSize": 200},
        "columns": [{"id": "e2", "label": "A", "width": 80, "bind": "{{row.a}}"}]}
    ]},
    "pageFooter": {"elements": [], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"latin": ["Noto Sans"]},
  "locale": "en",
  "nextId": 5,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
}

func TestRowTallerThanContentWindowStillReturnsLocatedTableOverflow(t *testing.T) {
	doc := tooTallRowDoc()
	data := multiRowTableData(1, -1)

	tpl, err := ParseTemplate([]byte(doc))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	bands, err := documentBands(tpl)
	if err != nil {
		t.Fatalf("documentBands: %v", err)
	}
	dataVal := mustDecodeData(t, data)
	params := mustDecodeParams(t)
	geometry, err := pageGeometryOf(tpl)
	if err != nil {
		t.Fatalf("pageGeometryOf: %v", err)
	}
	fc := testFormatContext()
	contentRuns, tableRects, _, err := collectBandTableRuns(tpl, bands, contentBandIndex, dataVal, params, fc, testShippedFontSet(), newFontCache(), nil)
	if err != nil {
		t.Fatalf("collectBandTableRuns: %v", err)
	}
	textRuns, _, _, err := collectBandTextRuns(tpl, bands, contentBandIndex, dataVal, params, testShippedFontSet(), newFontCache(), contentBandResolver, nil)
	if err != nil {
		t.Fatalf("collectBandTextRuns: %v", err)
	}
	contentRuns = append(textRuns, contentRuns...)

	// Presence precondition: the row's own chrome really is taller than
	// the content window, or the overflow case is not exercised.
	contentHeight := layout.ContentHeight(geometry)
	var rowHeight int64
	for _, r := range tableRects {
		if r.isDataRow {
			rowHeight = int64(r.bottom - r.top)
		}
	}
	if rowHeight <= int64(contentHeight) {
		t.Fatalf("presence precondition: the row is %dmp tall, which FITS the %dmp window — the overflow case is not exercised", rowHeight, contentHeight)
	}

	// No explicit timeout wrapper: the sweep is a SINGLE forward pass
	// (R6) — if this call ever hangs, the test binary's own default
	// timeout fails the package, which is itself the "did it hang or
	// did it answer wrong" information 4.6 wants recorded (see the
	// Delivery Log).
	_, perr := layout.Paginate(geometry, itemsForTest(contentRuns, tableRects))
	if perr == nil {
		t.Fatal("Paginate accepted a row taller than the content window; FR44's diagnostic must fire for a too-tall ROW exactly as it does for a too-tall LINE")
	}
	var overflow *layout.OverflowError
	if !errors.As(perr, &overflow) {
		t.Fatalf("Paginate returned %T, want *layout.OverflowError", perr)
	}
	if overflow.Kind != "table" {
		t.Errorf("overflow.Kind = %q, want %q", overflow.Kind, "table")
	}
	if overflow.ElementID != "e1" {
		t.Errorf("overflow.ElementID = %q, want %q", overflow.ElementID, "e1")
	}
}

// TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender is
// AC4's OTHER arm — Finding 4 (this story's finisher review, Major): the
// shipped path reaches layout.Paginate through PHASE A first
// (contentColumnItems, page_number.go), which appends TEXT before RECTS,
// so at the default zero top padding the row's first LINE item ties the
// chrome on Top and is visited first — Kind comes out "line", not
// "table". This is NOT a regression (the identical probe at a 903bf8f
// worktree produces the identical message), but the pin in the sibling
// test above asserts an ordering production never reaches. This test
// goes through the actual public Render(), the same entry point AC4's
// own "exactly as at 903bf8f" claim is about, and pins what a caller of
// this library actually observes — the pin 4.6 must build its
// clip-and-diagnose behaviour against.
func TestRowTallerThanContentWindowReturnsLocatedOverflowThroughRender(t *testing.T) {
	tpl, err := ParseTemplate([]byte(tooTallRowDoc()))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	_, rerr := Render(tpl, Data(multiRowTableData(1, -1)), nil, testShippedFontSet())
	if rerr == nil {
		t.Fatal("Render accepted a document whose only row is taller than the content window; FR44's diagnostic must fire")
	}
	var overflow *layout.OverflowError
	if !errors.As(rerr, &overflow) {
		t.Fatalf("Render returned %T (chain: %v), want an error wrapping *layout.OverflowError", rerr, rerr)
	}
	if overflow.Kind != "line" {
		t.Errorf(`overflow.Kind = %q, want %q — this is the ordering (PHASE A, text before rects) the shipped path actually uses; the mislabelling ("line" rather than "table" for a table row) predates this story and belongs to 4.6, not this test`, overflow.Kind, "line")
	}
	if overflow.ElementID != "e1" {
		t.Errorf("overflow.ElementID = %q, want %q", overflow.ElementID, "e1")
	}
}

// headerPushedToNextPageDoc places the table far enough down its content
// band (el.Y past contentHeight-headerHeight) that the HEADER ROW itself
// does not fit page 0's window at all — AC5's own fixture shape, applied
// to the header rather than a data row.
func headerPushedToNextPageDoc() string {
	return `{
  "assets": {},
  "bands": {
    "content": {"elements": [
      {"id": "e6", "type": "text", "x": 0, "y": 0, "width": 180, "height": 8, "value": "filler", "style": {"fontFamily": "latin", "fontSize": 8}},
      {"id": "e1", "type": "table", "x": 0, "y": 105, "bind": "items[]", "headerHeight": 10,
        "style": {"fontFamily": "latin", "fontSize": 8},
        "columns": [
          {"id": "e2", "label": "A", "width": 80, "bind": "{{row.a}}"},
          {"id": "e3", "label": "B", "width": 80, "bind": "{{row.b}}"}
        ]}
    ]},
    "pageFooter": {"elements": [], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"latin": ["Noto Sans"]},
  "locale": "en",
  "nextId": 7,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
}

// TestHeaderRowMovesWholeToNextPageAndIsNotRepeated is AC5: a header
// that does not fit the remaining content height moves — chrome AND
// column labels TOGETHER — to the next page, exactly like a data row
// (D2's own note: this already holds by the chrome accident at 903bf8f;
// the mechanism this story ships must hold it for a REASON). The second
// half of AC5 is the fence: the header must NOT be repeated on any later
// page (4.4's job), and this test confirms it stays absent past the one
// page it landed on.
func TestHeaderRowMovesWholeToNextPageAndIsNotRepeated(t *testing.T) {
	const rows = 60
	plan, contentRuns, tableRects := paginateContentTableForTest(t, headerPushedToNextPageDoc(), multiRowTableData(rows, -1))

	if len(plan.Pages) < 3 {
		t.Fatalf("presence precondition: fixture must paginate to >= 3 pages so 'not repeated' is observable on a page the header did NOT land on, got %d", len(plan.Pages))
	}
	// Presence precondition: page 0 carries the FILLER (so the sweep has
	// already committed to page 0 before reaching the table — otherwise
	// the "no page is ever empty" rule slides window 0 to the header's
	// own top instead of advancing the page, and the header would "move"
	// to a page that never had anything else on it, proving nothing) but
	// NO table content at all — no chrome rect, no header label.
	if len(plan.Pages[0].ContentRects) != 0 {
		t.Fatalf("presence precondition: page 0 carries %d rect(s) — the header must not fit page 0's window at all", len(plan.Pages[0].ContentRects))
	}
	for _, ref := range plan.Pages[0].ContentRuns {
		if contentRuns[ref].isHeaderLabel || contentRuns[ref].isTableRowLine {
			t.Fatalf("presence precondition: page 0 carries table content — the header must not fit page 0's window at all, or this fixture does not exercise AC5")
		}
	}

	// CORE invariant, survivable under mutation (a) (no header chrome at
	// all): every header-LABEL run lands on the SAME page as every other
	// — the header's own labels stay together even with no chrome to
	// carry them, which is what proves this is a mechanism and not the
	// chrome accident.
	headerLabelPages := map[int]bool{}
	for p, pg := range plan.Pages {
		for _, ref := range pg.ContentRuns {
			if contentRuns[ref].isHeaderLabel {
				headerLabelPages[p] = true
			}
		}
	}
	if len(headerLabelPages) == 0 {
		t.Fatal("no header-label runs were found at all — test fixture defect")
	}
	if len(headerLabelPages) != 1 {
		t.Errorf("header labels are spread across %d distinct pages; they must all land on ONE page", len(headerLabelPages))
	}
	var headerPage int
	for p := range headerLabelPages {
		headerPage = p
	}
	if headerPage == 0 {
		t.Errorf("header labels landed on page 0 — the header must not fit page 0's window at all (presence precondition failed silently)")
	}

	// SECONDARY invariant, only meaningful when the chrome exists (the
	// UNMUTATED tree, and mutation (b)'s subject): the header's chrome
	// must be on the SAME page as its labels.
	headerChromePages := map[int]bool{}
	for p, pg := range plan.Pages {
		for _, ref := range pg.ContentRects {
			if tableRects[ref].isHeaderRow {
				headerChromePages[p] = true
			}
		}
	}
	if len(headerChromePages) > 0 {
		if len(headerChromePages) != 1 || !headerChromePages[headerPage] {
			t.Errorf("header chrome pages = %v, want exactly {%d} (the same page as the header's labels)", headerChromePages, headerPage)
		}
	}

	// FENCE, both halves: the header must appear on EXACTLY ONE page —
	// repeating it is Story 4.4's job, not this story's.
	for p := range plan.Pages {
		if p == headerPage {
			continue
		}
		if headerChromePages[p] {
			t.Errorf("page %d carries a header rect — the header must appear on EXACTLY ONE page (repeating it is Story 4.4's job, not this story's)", p)
		}
		if headerLabelPages[p] {
			t.Errorf("page %d carries a header label — the header must appear on EXACTLY ONE page", p)
		}
	}
}

// TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter is
// Finding 3 (this story's finisher review) and Major 3's coverage gap
// together: no test this story added ever called the public Render() —
// every AC1/AC2/AC3/AC5 test above stops at layout.Paginate — so a
// rendering-level regression (Finding 1) could not have been caught by
// anything in this file. This test takes the `{{page}} of {{pages}}`
// branch through Render() (the fixture's own id/nextId bug that made it
// unparseable is fixed above), and cross-checks PHASE A's page count
// (the value {{page}}/{{pages}} print, D-2.7.2) against the PRODUCED
// PDF's own /Type /Page object count and its /Count entry — two
// independent sites, D-000.33, not a conservation law.
func TestMultiRowTableRendersThroughPublicRenderWithPageCountFooter(t *testing.T) {
	table := []struct {
		rows      int
		wrapRow   int
		wantPages int
	}{
		{rows: 20, wrapRow: 5, wantPages: 3},
		{rows: 40, wrapRow: -1, wantPages: 5},
		{rows: 60, wrapRow: -1, wantPages: 7},
	}
	for _, row := range table {
		t.Run(fmt.Sprintf("rows=%d", row.rows), func(t *testing.T) {
			tplJSON := multiRowTableDoc(true)
			tpl, err := ParseTemplate([]byte(tplJSON))
			if err != nil {
				t.Fatalf("ParseTemplate: %v — the footerPageCount=true branch must parse", err)
			}
			res, err := Render(tpl, Data(multiRowTableData(row.rows, row.wrapRow)), nil, testShippedFontSet())
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if len(res.Bytes) == 0 {
				t.Fatal("presence precondition: Render returned zero bytes")
			}

			// PHASE A's own answer (the value {{page}}/{{pages}} print),
			// derived independently via layout.Paginate over the SAME
			// content-band table, mirroring what predictDocument does
			// before the footer/header bands are collected.
			plan, _, _ := paginateContentTableForTest(t, multiRowTableDoc(false), multiRowTableData(row.rows, row.wrapRow))
			if len(plan.Pages) != row.wantPages {
				t.Fatalf("presence precondition: the content-band-only pagination itself produced %d pages, want %d — widen the fixture", len(plan.Pages), row.wantPages)
			}

			if got := countPageObjects(res.Bytes); got != row.wantPages {
				t.Errorf("rendered %d /Type /Page object(s) for %d row(s); PHASE A (and this test's declared value) says %d", got, row.rows, row.wantPages)
			}
			if got := readDeclaredCount(t, res.Bytes); got != row.wantPages {
				t.Errorf("/Count is %d for %d row(s); the declared value is %d", got, row.rows, row.wantPages)
			}
		})
	}
}

// tableBesideSameYElementDoc places a table and an ORDINARY text element
// at the SAME y within the content band — Finding 1's exact shape (this
// story's finisher review): a caption, a column heading, or any other
// element level with a table is a canonical, LEGAL report layout, not a
// caller bug. The two elements do not overlap horizontally (x 0..120 for
// the table's two 60pt columns, x 130..175 for the note) — Finding 1
// measured that horizontal overlap is not even required to trigger the
// pre-fix defect, only a shared y.
func tableBesideSameYElementDoc() string {
	return `{
  "assets": {},
  "bands": {
    "content": {"elements": [
      {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
        "style": {"fontFamily": "latin", "fontSize": 8},
        "columns": [
          {"id": "e2", "label": "A", "width": 60, "bind": "{{row.a}}"},
          {"id": "e3", "label": "B", "width": 60, "bind": "{{row.b}}"}
        ]},
      {"id": "e4", "type": "text", "x": 130, "y": 0, "width": 45, "height": 8, "value": "Note", "style": {"fontFamily": "latin", "fontSize": 8}}
    ]},
    "pageFooter": {"elements": [], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"latin": ["Noto Sans"]},
  "locale": "en",
  "nextId": 5,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
}

// TestTableBesideSameYElementRenders is Finding 1's regression, at the
// Render() level (this story's finisher review): before the fix, this
// exact document returned "internal error: group ... is not contiguous
// in column order" — an ordinary, legal template made unrenderable by
// R7's contiguity check firing on a caller who did nothing wrong. The
// finisher review measured the identical document rendering successfully
// against a 903bf8f worktree, which is what makes this a REGRESSION
// rather than a pre-existing limitation. It must render here too, and a
// table beside a same-y element is exactly the coverage hole Finding 3
// names as the reason no test in this story called Render() at all.
func TestTableBesideSameYElementRenders(t *testing.T) {
	tpl, err := ParseTemplate([]byte(tableBesideSameYElementDoc()))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	if _, err := Render(tpl, Data(multiRowTableData(3, -1)), nil, testShippedFontSet()); err != nil {
		t.Fatalf("Render: %v — a table beside an ordinary element merely sharing its y must render; this is Finding 1's regression", err)
	}
}
