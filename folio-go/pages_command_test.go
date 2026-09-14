package folio

import (
	"errors"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/template"
)

// SPEC-multi-pages story 2: the addPage, deletePage and setPageBreak engine
// commands, their shape changes (D-G.1) and their refusals.

func applyPageCommand(t *testing.T, tpl *Template, command string) CanvasProjection {
	t.Helper()
	projection, err := ApplyComponentCommand(tpl, []byte(command))
	if err != nil {
		t.Fatalf("%s: %v", command, err)
	}
	return projection
}

func refusePageCommand(t *testing.T, tpl *Template, command, wantPath string) {
	t.Helper()
	before, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ApplyComponentCommand(tpl, []byte(command))
	var failure *ComponentCommandError
	if !errors.As(err, &failure) {
		t.Fatalf("%s: err = %v, want a located refusal", command, err)
	}
	if failure.DataPath != wantPath {
		t.Errorf("%s: refused at %q, want %q (%s)", command, failure.DataPath, wantPath, failure.Message)
	}
	after, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Errorf("%s: a refused command changed the document", command)
	}
}

func reparse(t *testing.T, tpl *Template) *template.Document {
	t.Helper()
	saved, err := SerializeTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	d, err := template.ParseDocument(saved)
	if err != nil {
		t.Fatalf("the saved bytes do not load: %v\n%s", err, saved)
	}
	return d
}

func elementIDs(elements []template.Element) []string {
	var out []string
	for _, el := range elements {
		out = append(out, string(el.ID))
	}
	return out
}

func TestAddPageToAOnePageDocumentMovesItsContentAndBreakIntoPageOne(t *testing.T) {
	tpl := multiPageTemplate(t, sectionBreakStatementTemplateJSON)
	original := reparse(t, tpl)
	if original.PageCount() != 1 || !original.Bands.Content.SectionBreak.Set || len(original.Bands.Content.Elements) == 0 {
		t.Fatal("fixture precondition: a one-page document with elements and a section break")
	}
	projection := applyPageCommand(t, tpl, `{"kind":"addPage","version":1}`)
	d := reparse(t, tpl)
	if d.PageCount() != 2 || len(d.Bands.Content.Elements) != 0 || d.Bands.Content.SectionBreak.Set {
		t.Fatalf("pages = %d, bands.content elements = %d: want two pages and an empty bands.content", d.PageCount(), len(d.Bands.Content.Elements))
	}
	if strings.Join(elementIDs(d.Pages[0].Elements), ",") != strings.Join(elementIDs(original.Bands.Content.Elements), ",") || d.Pages[0].SectionBreak != original.Bands.Content.SectionBreak {
		t.Error("page 1 does not hold the old content and its break")
	}
	if len(d.Pages[1].Elements) != 0 || !d.Pages[1].PageBreak {
		t.Error("the new page is not empty with Page Break on")
	}
	if d.Version != "4.1" {
		t.Errorf("version = %s, want 4.1", d.Version)
	}
	if len(projection.PageBreaks) != 2 || !projection.PageBreaks[0] || !projection.PageBreaks[1] {
		t.Errorf("projected pageBreaks = %v", projection.PageBreaks)
	}
}

func TestAddPageInsertsAfterTheNamedPageOrAtTheEnd(t *testing.T) {
	tpl := multiPageTemplate(t, multiPageStatementTemplateJSON)
	applyPageCommand(t, tpl, `{"kind":"addPage","version":1}`)
	d := reparse(t, tpl)
	if d.PageCount() != 3 || len(d.Pages[2].Elements) != 0 || len(d.Pages[1].Elements) == 0 {
		t.Fatalf("without after: want the empty page appended as page 3")
	}
	// After the last page appends, exactly as no after does.
	applyPageCommand(t, tpl, `{"kind":"addPage","version":1,"after":2}`)
	d = reparse(t, tpl)
	if d.PageCount() != 4 || len(d.Pages[3].Elements) != 0 || !d.Pages[3].PageBreak || len(d.Pages[1].Elements) == 0 {
		t.Fatalf("after the last index: want an empty page appended as page 4")
	}
	applyPageCommand(t, tpl, `{"kind":"deletePage","version":1,"page":3}`)
	// Mark old page 3 so its new position is visible.
	applyPageCommand(t, tpl, `{"kind":"setPageBreak","version":1,"page":2,"pageBreak":false}`)
	applyPageCommand(t, tpl, `{"kind":"addPage","version":1,"after":1}`)
	d = reparse(t, tpl)
	if d.PageCount() != 4 {
		t.Fatalf("pages = %d, want 4", d.PageCount())
	}
	if len(d.Pages[2].Elements) != 0 || !d.Pages[2].PageBreak {
		t.Error("the page added after page 2 is not an empty page 3 with Page Break on")
	}
	if d.Pages[3].PageBreak {
		t.Error("old page 3 did not become page 4")
	}
}

func TestDeletePageRemovesThePageAndEveryElementOnIt(t *testing.T) {
	tpl := multiPageTemplate(t, multiPageStatementTemplateJSON)
	applyPageCommand(t, tpl, `{"kind":"addPage","version":1,"after":0}`)
	// Three pages: the statement, an empty page, the terms.
	terms := elementIDs(reparse(t, tpl).Pages[2].Elements)
	applyPageCommand(t, tpl, `{"kind":"deletePage","version":1,"page":1}`)
	d := reparse(t, tpl)
	if d.PageCount() != 2 || strings.Join(elementIDs(d.Pages[1].Elements), ",") != strings.Join(terms, ",") {
		t.Fatal("deleting page 2 did not make page 3 the new page 2")
	}
	projection := applyPageCommand(t, tpl, `{"kind":"deletePage","version":1,"page":1}`)
	d = reparse(t, tpl)
	if d.PageCount() != 1 || d.Pages != nil || len(d.Bands.Content.Elements) == 0 {
		t.Fatal("deleting down to one page did not return the one-page shape")
	}
	for _, id := range terms {
		for _, el := range d.Bands.Content.Elements {
			if string(el.ID) == id {
				t.Errorf("element %s of the deleted page survived", id)
			}
		}
		for _, component := range projection.Components {
			if component.ID == id {
				t.Errorf("element %s of the deleted page is still projected", id)
			}
		}
	}
	if len(projection.PageBreaks) != 1 || len(projection.ContentWindowPages) == 0 {
		t.Errorf("a one-page projection carries pageBreaks %v", projection.PageBreaks)
	}
}

func TestDeletingDownToOnePageKeepsPageOnesSectionBreak(t *testing.T) {
	tpl := multiPageTemplate(t, sectionBreakStatementTemplateJSON)
	breakBefore := reparse(t, tpl).Bands.Content.SectionBreak
	applyPageCommand(t, tpl, `{"kind":"addPage","version":1}`)
	applyPageCommand(t, tpl, `{"kind":"deletePage","version":1,"page":1}`)
	d := reparse(t, tpl)
	if d.Pages != nil || d.Bands.Content.SectionBreak != breakBefore {
		t.Fatalf("the break is not back in bands.content: %+v", d.Bands.Content.SectionBreak)
	}
}

func TestDeletingPageOneTakesItsBreakAndMakesPageTwoTheSinglePage(t *testing.T) {
	tpl := multiPageTemplate(t, sectionBreakStatementTemplateJSON)
	applyPageCommand(t, tpl, `{"kind":"addPage","version":1}`)
	applyPageCommand(t, tpl, `{"kind":"setPageBreak","version":1,"page":1,"pageBreak":false}`)
	applyPageCommand(t, tpl, `{"kind":"deletePage","version":1,"page":0}`)
	d := reparse(t, tpl)
	if d.Pages != nil || len(d.Bands.Content.Elements) != 0 || d.Bands.Content.SectionBreak.Set {
		t.Fatal("old page 2 (empty, no break) did not become the single page")
	}
	saved, _ := SerializeTemplate(tpl)
	if strings.Contains(string(saved), "pageBreak") {
		t.Error("a one-page document still writes a pageBreak")
	}
}

func TestSetPageBreakWritesTheValueOnALaterPage(t *testing.T) {
	tpl := multiPageTemplate(t, multiPageStatementTemplateJSON)
	projection := applyPageCommand(t, tpl, `{"kind":"setPageBreak","version":1,"page":1,"pageBreak":false}`)
	saved, _ := SerializeTemplate(tpl)
	if !strings.Contains(string(saved), `"pageBreak": false`) {
		t.Error("the saved bytes do not state pageBreak false")
	}
	if len(projection.PageBreaks) != 2 || !projection.PageBreaks[0] || projection.PageBreaks[1] {
		t.Errorf("projected pageBreaks = %v, want [true false]", projection.PageBreaks)
	}
}

func TestPageCommandsRefuseWhatTheyCannotDo(t *testing.T) {
	one := multiPageTemplate(t, sectionBreakStatementTemplateJSON)
	refusePageCommand(t, one, `{"kind":"deletePage","version":1,"page":0}`, "bands.content")
	refusePageCommand(t, one, `{"kind":"setPageBreak","version":1,"page":0,"pageBreak":false}`, "bands.content.pageBreak")
	refusePageCommand(t, one, `{"kind":"addPage","version":1,"after":1}`, "pages")

	many := multiPageTemplate(t, multiPageStatementTemplateJSON)
	refusePageCommand(t, many, `{"kind":"setPageBreak","version":1,"page":0,"pageBreak":false}`, "pages[0].pageBreak")
	refusePageCommand(t, many, `{"kind":"setPageBreak","version":1,"page":1,"pageBreak":null}`, "pages[1].pageBreak")
	refusePageCommand(t, many, `{"kind":"setPageBreak","version":1,"page":2,"pageBreak":true}`, "pages")
	refusePageCommand(t, many, `{"kind":"deletePage","version":1,"page":-1}`, "pages")
	refusePageCommand(t, many, `{"kind":"deletePage","version":1,"page":null}`, "pages")
	refusePageCommand(t, many, `{"kind":"deletePage","version":1}`, "pages")
	refusePageCommand(t, many, `{"kind":"addPage","version":1,"after":null}`, "pages")
	refusePageCommand(t, many, `{"kind":"addPage","version":1,"after":0,"extra":1}`, "pages")
}

func TestContentComponentsProjectTheirDesignedPage(t *testing.T) {
	tpl := multiPageTemplate(t, multiPageStatementTemplateJSON)
	projection := shippedProjection(t, tpl)
	d := reparse(t, tpl)
	want := map[string]int{}
	for page, content := range d.Pages {
		for _, el := range content.Elements {
			want[string(el.ID)] = page
		}
	}
	laterPage := 0
	for _, component := range projection.Components {
		if component.Band != bandContent {
			if component.Page != 0 {
				t.Errorf("%s component %s projects page %d, want 0", component.Band, component.ID, component.Page)
			}
			continue
		}
		if component.Page != want[component.ID] {
			t.Errorf("component %s projects page %d, want %d", component.ID, component.Page, want[component.ID])
		}
		if component.Page == 1 {
			laterPage++
		}
	}
	if laterPage == 0 {
		t.Fatal("fixture precondition: no page-2 component was projected")
	}
	for _, component := range shippedProjection(t, multiPageTemplate(t, sectionBreakStatementTemplateJSON)).Components {
		if component.Page != 0 {
			t.Errorf("a one-page document projects component %s on page %d", component.ID, component.Page)
		}
	}
}
