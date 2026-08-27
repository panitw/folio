package folio

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/pagemodel"
)

var alternatingBase = pagemodel.Color{R: 0x11, G: 0x22, B: 0x33}
var alternatingFill = pagemodel.Color{R: 0xDD, G: 0xEE, B: 0xFF}

func alternatingTableDoc(baseBackground, altBackground string) string {
	background := ""
	if baseBackground != "" {
		background = fmt.Sprintf(`, "background": %q`, baseBackground)
	}
	return fmt.Sprintf(`{
  "assets": {},
  "bands": {
    "content": {"elements": [
      {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
        "style": {"fontFamily": "latin", "fontSize": 8%s},
        "headerStyle": {"background": "#445566"},
        "altRowBackground": %q,
        "columns": [
          {"id": "e2", "label": "A", "width": 60, "bind": "{{row.a}}"},
          {"id": "e3", "label": "B", "width": 60, "bind": "{{row.b}}"}
        ]}
    ]},
    "pageFooter": {"elements": [], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"latin": ["Noto Sans"]},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "1.0"
}`, background, altBackground)
}

func fiveAlternatingRowsData(prefix string) string {
	return fmt.Sprintf(`{"items":[
    {"a":%q,"b":"0"},{"a":%q,"b":"1"},{"a":%q,"b":"2"},
    {"a":%q,"b":"3"},{"a":%q,"b":"4"}
  ]}`, prefix+"0", prefix+"1", prefix+"2", prefix+"3", prefix+"4")
}

func dataRowRectGroups(t *testing.T, pages []pagemodel.Page, columns int) [][]pagemodel.Rect {
	t.Helper()
	if len(pages) != 1 {
		t.Fatalf("got %d pages, want 1", len(pages))
	}
	rects := pages[0].Rects
	if len(rects) != columns*6 {
		t.Fatalf("got %d rects, want %d (one header plus five data rows, %d columns each)", len(rects), columns*6, columns)
	}
	groups := make([][]pagemodel.Rect, 5)
	for row := range groups {
		start := columns * (row + 1)
		groups[row] = rects[start : start+columns]
	}
	return groups
}

func assertRowFills(t *testing.T, groups [][]pagemodel.Rect, base *pagemodel.Color, alt pagemodel.Color) {
	t.Helper()
	for row, cells := range groups {
		for column, rect := range cells {
			if row%2 == 1 {
				if !rect.HasFill || rect.Fill != alt {
					t.Errorf("row %d column %d fill = {present:%v color:%+v}, want alternate %+v", row, column, rect.HasFill, rect.Fill, alt)
				}
				continue
			}
			if base == nil {
				if rect.HasFill {
					t.Errorf("row %d column %d unexpectedly filled %+v; no base background was declared", row, column, rect.Fill)
				}
			} else if !rect.HasFill || rect.Fill != *base {
				t.Errorf("row %d column %d fill = {present:%v color:%+v}, want base %+v", row, column, rect.HasFill, rect.Fill, *base)
			}
		}
	}
}

func TestAlternatingRowBackgroundAppliesToOddCollectionIndexes(t *testing.T) {
	for _, tc := range []struct {
		name string
		base string
		want *pagemodel.Color
	}{
		{name: "ordinary body background remains on even rows", base: "#112233", want: &alternatingBase},
		{name: "even rows remain unfilled without a body background"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pages := tablePagesForTest(t, alternatingTableDoc(tc.base, "#DDEEFF"), fiveAlternatingRowsData("R"))
			assertRowFills(t, dataRowRectGroups(t, pages, 2), tc.want, alternatingFill)
		})
	}
}

func TestAlternatingRowChoiceDependsOnTemplateAndCollectionIndexOnly(t *testing.T) {
	bluePages := tablePagesForTest(t, alternatingTableDoc("#112233", "#DDEEFF"), fiveAlternatingRowsData("A"))
	redPages := tablePagesForTest(t, alternatingTableDoc("#112233", "#CC0000"), fiveAlternatingRowsData("A"))
	changedDataPages := tablePagesForTest(t, alternatingTableDoc("#112233", "#DDEEFF"), fiveAlternatingRowsData("Z"))

	blueGroups := dataRowRectGroups(t, bluePages, 2)
	redGroups := dataRowRectGroups(t, redPages, 2)
	changedGroups := dataRowRectGroups(t, changedDataPages, 2)
	for row := range blueGroups {
		for column := range blueGroups[row] {
			if blueGroups[row][column].HasFill != changedGroups[row][column].HasFill || blueGroups[row][column].Fill != changedGroups[row][column].Fill {
				t.Errorf("row %d column %d changed its fill when unrelated cell values changed", row, column)
			}
			if row%2 == 0 && (!blueGroups[row][column].HasFill || blueGroups[row][column].Fill != alternatingBase) {
				t.Errorf("row %d column %d ordinary body fill = {present:%v color:%+v}, want base %+v", row, column, blueGroups[row][column].HasFill, blueGroups[row][column].Fill, alternatingBase)
			}
		}
	}
	if blueGroups[1][0].Fill == redGroups[1][0].Fill {
		t.Fatal("changing only altRowBackground did not change the odd row's page-model fill")
	}

	render := func(t *testing.T, alt string) []byte {
		t.Helper()
		tpl, err := ParseTemplate([]byte(alternatingTableDoc("#112233", alt)))
		if err != nil {
			t.Fatalf("ParseTemplate: %v", err)
		}
		res, err := Render(tpl, Data(fiveAlternatingRowsData("A")), nil, testShippedFontSet())
		if err != nil {
			t.Fatalf("Render: %v", err)
		}
		return res.Bytes
	}
	if bytes.Equal(render(t, "#DDEEFF"), render(t, "#CC0000")) {
		t.Fatal("templates differing only in altRowBackground produced byte-identical PDFs")
	}
}

func TestAlternatingRowBackgroundUsesExistingLocatedColourError(t *testing.T) {
	tpl, err := ParseTemplate([]byte(alternatingTableDoc("#112233", "not-a-colour")))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	res, err := Render(tpl, Data(fiveAlternatingRowsData("R")), nil, testShippedFontSet())
	if err == nil {
		t.Fatal("expected malformed table.altRowBackground to fail when collection index 1 is reached")
	}
	if len(res.Bytes) != 0 {
		t.Errorf("Render returned %d successful PDF bytes alongside the error, want none", len(res.Bytes))
	}
	var renderErr *RenderError
	if !errors.As(err, &renderErr) {
		t.Fatalf("errors.As(*RenderError) failed: %T: %v", err, err)
	}
	if renderErr.Diagnostic.Code != DiagCodeStyleColorInvalid {
		t.Errorf("Code = %q, want %q", renderErr.Diagnostic.Code, DiagCodeStyleColorInvalid)
	}
	if renderErr.Diagnostic.ElementID != "e1" {
		t.Errorf("ElementID = %q, want table element e1", renderErr.Diagnostic.ElementID)
	}
	if !strings.Contains(renderErr.Diagnostic.Message, "table.altRowBackground") {
		t.Errorf("message does not name table.altRowBackground: %q", renderErr.Diagnostic.Message)
	}
}

func alternatingPaginatedFooterDoc() string {
	doc := footerFixtureDoc("count", false)
	doc = strings.Replace(doc,
		`"style": {"fontFamily": "latin", "fontSize": 8}`,
		`"style": {"fontFamily": "latin", "fontSize": 8, "background": "#112233"}, "headerStyle": {"background": "#445566"}, "altRowBackground": "#DDEEFF"`, 1)
	return doc
}

// TestAlternatingRowBackgroundIsGeometryAndPaginationNeutral is the permanent
// relational witness for AD-13/AD-24.  It renders the exact same 20-row,
// three-page table with and without the optional field, then proves that the
// final page model has the same page partition, every run, and every
// non-fill rectangle field.  Fill presence/colour are intentionally the only
// permitted difference: they are the feature's whole output contract.
func TestAlternatingRowBackgroundIsGeometryAndPaginationNeutral(t *testing.T) {
	const rows = 20
	withAlternate := alternatingPaginatedFooterDoc()
	withoutAlternate := strings.Replace(withAlternate, `, "altRowBackground": "#DDEEFF"`, "", 1)
	data := footerFixtureData(rows)

	withPages := tablePagesForTest(t, withAlternate, data)
	withoutPages := tablePagesForTest(t, withoutAlternate, data)
	if len(withPages) < 3 {
		t.Fatalf("presence precondition: alternate fixture has %d pages, want at least 3", len(withPages))
	}
	if len(withPages) != len(withoutPages) {
		t.Fatalf("page partition changed: with alternate=%d pages, without=%d", len(withPages), len(withoutPages))
	}

	fillDifferences := 0
	for pageIndex := range withPages {
		withPage, withoutPage := withPages[pageIndex], withoutPages[pageIndex]
		if withPage.Width != withoutPage.Width || withPage.Height != withoutPage.Height ||
			withPage.MarginTop != withoutPage.MarginTop || withPage.MarginLeft != withoutPage.MarginLeft {
			t.Errorf("page %d geometry changed: with=%+v without=%+v", pageIndex, withPage, withoutPage)
		}
		if len(withPage.Rects) != len(withoutPage.Rects) {
			t.Fatalf("page %d rectangle cardinality changed: with alternate=%d, without=%d", pageIndex, len(withPage.Rects), len(withoutPage.Rects))
		}
		if len(withPage.Runs) != len(withoutPage.Runs) {
			t.Fatalf("page %d run cardinality changed: with alternate=%d, without=%d", pageIndex, len(withPage.Runs), len(withoutPage.Runs))
		}
		if !reflect.DeepEqual(withPage.Runs, withoutPage.Runs) {
			t.Errorf("page %d text runs changed when only altRowBackground was declared", pageIndex)
		}
		for rectIndex := range withPage.Rects {
			withRect, withoutRect := withPage.Rects[rectIndex], withoutPage.Rects[rectIndex]
			if withRect.HasFill != withoutRect.HasFill || withRect.Fill != withoutRect.Fill {
				fillDifferences++
			}
			withRect.HasFill, withRect.Fill = false, pagemodel.Color{}
			withoutRect.HasFill, withoutRect.Fill = false, pagemodel.Color{}
			if withRect != withoutRect {
				t.Errorf("page %d rectangle %d changed outside HasFill/Fill: with=%+v without=%+v", pageIndex, rectIndex, withRect, withoutRect)
			}
		}
	}
	if fillDifferences != 30 {
		t.Errorf("fill differences = %d, want 30 alternate cells (10 odd rows × 3 columns)", fillDifferences)
	}
}

func TestAlternatingRowBackgroundContinuesAcrossPagesAndExcludesHeaderFooter(t *testing.T) {
	const rows = 20
	doc := alternatingPaginatedFooterDoc()
	data := footerFixtureData(rows)
	plan, _, sources := paginateContentTableForTest(t, doc, data)
	if len(plan.Pages) < 3 {
		t.Fatalf("presence precondition: got %d pages, want at least 3", len(plan.Pages))
	}

	discriminatingBoundary := false
	for pageIndex, page := range plan.Pages {
		first := -1
		for _, ref := range page.ContentRects {
			source := sources[ref]
			if !source.isDataRow {
				continue
			}
			if first == -1 {
				first = source.rowIndex
			}
			want := alternatingBase
			if source.rowIndex%2 == 1 {
				want = alternatingFill
			}
			for column, rect := range source.rects {
				if !rect.HasFill || rect.Fill != want {
					t.Errorf("page %d row %d column %d fill = {present:%v color:%+v}, want %+v from collection parity", pageIndex, source.rowIndex, column, rect.HasFill, rect.Fill, want)
				}
			}
		}
		if pageIndex > 0 && first%2 == 1 {
			discriminatingBoundary = true
		}
	}
	if !discriminatingBoundary {
		t.Fatal("presence precondition: no continuation page begins on an odd collection index; page-local reset would be indistinguishable")
	}

	header := pagemodel.Color{R: 0x44, G: 0x55, B: 0x66}
	seenHeader, seenFooter := 0, 0
	for _, source := range sources {
		switch {
		case source.isHeaderRow:
			seenHeader++
			for _, rect := range source.rects {
				if !rect.HasFill || rect.Fill != header {
					t.Errorf("source header fill = {present:%v color:%+v}, want header %+v", rect.HasFill, rect.Fill, header)
				}
			}
		case source.isFooterRow:
			seenFooter++
			for _, rect := range source.rects {
				if !rect.HasFill || rect.Fill != alternatingBase {
					t.Errorf("source footer fill = {present:%v color:%+v}, want body %+v", rect.HasFill, rect.Fill, alternatingBase)
				}
			}
		}
	}
	if seenHeader != 1 || seenFooter != 1 {
		t.Fatalf("source groups: header=%d footer=%d, want one of each", seenHeader, seenFooter)
	}

	tpl, err := ParseTemplate([]byte(doc))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	pages, _, _, _, err := buildPageModel(tpl, mustDecodeData(t, data), mustDecodeParams(t), testShippedFontSet())
	if err != nil {
		t.Fatalf("buildPageModel: %v", err)
	}
	if len(pages) != len(plan.Pages) {
		t.Fatalf("buildPageModel pages=%d, pagination pages=%d", len(pages), len(plan.Pages))
	}
	for pageIndex, page := range pages {
		headerCells := 0
		for _, rect := range page.Rects {
			if rect.HasFill && rect.Fill == header {
				headerCells++
			}
		}
		if headerCells != 3 {
			t.Errorf("page %d carries %d header-colour cells, want 3 (original/repeated header retained)", pageIndex, headerCells)
		}
	}
	// The source assertions above prove the footer's assigned colour before
	// pagination. Count the final page-model rectangles as well: 20 rows ×
	// three cells gives 30 base cells at even indexes and 30 alternate cells at
	// odd indexes, while the three footer cells add only to the base colour.
	// This catches a pagination-stage mix-up that recolours the footer after
	// collectBandTableRuns has returned.
	baseCells, alternateCells := 0, 0
	for _, page := range pages {
		for _, rect := range page.Rects {
			switch {
			case rect.HasFill && rect.Fill == alternatingBase:
				baseCells++
			case rect.HasFill && rect.Fill == alternatingFill:
				alternateCells++
			}
		}
	}
	if baseCells != 33 {
		t.Errorf("final page-model carries %d body-colour cells, want 33 (30 even data cells plus 3 footer cells)", baseCells)
	}
	if alternateCells != 30 {
		t.Errorf("final page-model carries %d alternate-colour cells, want 30 (10 odd data rows × 3 columns)", alternateCells)
	}

	res, err := Render(tpl, Data(data), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	assertWellFormedPDF(t, "alternating paginated footer", res.Bytes, len(plan.Pages))
}
