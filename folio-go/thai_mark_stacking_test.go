package folio

import (
	"strings"
	"testing"
)

// DW-28's characterization, and the reason it is written HERE rather
// than beside the branch it exercises.
//
// internal/pdf's fail-closed branch for a glyph carrying a vertical
// offset (textdoc.go's appendShapedRun) has been exercised since Story
// 2.3 by handing the emitter a SYNTHETIC run. Both that test and the
// branch's own comment stated, as measured fact, that the branch is
// "UNREACHABLE through the render path with the shipped set and cannot
// be red-proved through it".
//
// THAT CLAIM WAS FALSE, and it was load-bearing: it is why no
// render-path test was ever built, and it is what a reader hitting the
// refusal in production would have been told when they went looking.
// Story 2.3 measured ITS OWN SAMPLES and reported on THE SHIPPED SET —
// two different populations. The samples happened to contain no Thai
// sequence stacking two marks over one base; ordinary Thai does, and
// the shipped Noto Sans Thai gives those marks a non-zero YOffset.
//
// So this file exists to hold the branch to a REAL DOCUMENT. It pins
// the message a document author actually receives, and it proves the
// branch is reached through ParseTemplate + Render — the same public
// entry point a caller uses — rather than through a hand-built run.
//
// IT DOES NOT FIX THE LIMIT. Stacked Thai marks still refuse; that is
// Epic 8's opening story. This is the characterization that must exist
// FIRST, so the fix has something to move: a test that says what the
// engine does today cannot be quietly widened or narrowed while the fix
// is written, and the two arms below will become the fix's own
// before/after.

// thaiStackedMarksTemplate is the smallest document that reaches the
// refusal: one text element whose value is the single word ทั้งสิ้น
// ("in total" / "altogether"), in which ั and ้ both sit above ท.
//
// It is deliberately ONE WORD. A whole clause reaches the same branch
// (see the owner's-document arm below), but a one-word document keeps
// the subset small enough that the CID in the pinned message is stable.
const thaiStackedMarksTemplate = `{
  "assets": {},
  "bands": {
    "content": {"elements": [
      {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 40,
       "value": "ทั้งสิ้น", "style": {"fontFamily": "body", "fontSize": 12}}
    ]},
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans", "Noto Sans Thai"]},
  "locale": "th",
  "nextId": 2,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36},
           "orientation": "portrait", "size": "A4"},
  "utcOffset": "+07:00",
  "version": "1.0"
}`

// thaiUnstackedTemplate is the NON-VACUITY CONTROL, and it is the whole
// reason the test above proves anything about mark stacking rather than
// about Thai.
//
// สัญญา ("contract") is the same script, the same face, the same font
// size and the same box. Its only difference is that no base carries
// two stacked marks. If this document ever stops rendering, the arm
// above stops being evidence about stacking and the failure is
// somewhere else entirely.
const thaiUnstackedTemplate = `{
  "assets": {},
  "bands": {
    "content": {"elements": [
      {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 200, "height": 40,
       "value": "สัญญา", "style": {"fontFamily": "body", "fontSize": 12}}
    ]},
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans", "Noto Sans Thai"]},
  "locale": "th",
  "nextId": 2,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36},
           "orientation": "portrait", "size": "A4"},
  "utcOffset": "+07:00",
  "version": "1.0"
}`

// thaiStackedMarksMessage is the message pinned VERBATIM, because the
// point of pinning it is that this is the text a document author reads
// when their document will not render. A reworded refusal is a changed
// product, not a changed internal, and it should have to be edited here
// deliberately.
const thaiStackedMarksMessage = "internal/pdf: face Noto Sans Thai: CID 3 carries a non-zero " +
	"vertical offset (-57), which a TJ array cannot express. Emitting the glyph without its " +
	"offset would place it wrongly with no observable difference in the output bytes, so this " +
	"fails rather than degrades."

// TestThaiStackedMarksAreRefusedThroughTheRenderPath is the reachability
// proof DW-28 records as missing: the branch IS reachable through the
// public entry point, with the shipped font set, on ordinary Thai.
func TestThaiStackedMarksAreRefusedThroughTheRenderPath(t *testing.T) {
	tpl, err := ParseTemplate([]byte(thaiStackedMarksTemplate))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	res, err := Render(tpl, Data("{}"), nil, testShippedFontSet())
	if err == nil {
		t.Fatalf("ordinary Thai stacking two marks over one base must refuse today; got %d bytes and no error", len(res.Bytes))
	}
	if got := err.Error(); !strings.Contains(got, thaiStackedMarksMessage) {
		t.Fatalf("the refusal an author reads has changed.\n got: %s\nwant it to contain: %s", got, thaiStackedMarksMessage)
	}
	// A refusal must produce NO bytes. A half-written document would be
	// worse than the refusal, because a caller that ignores the error
	// would ship it.
	if len(res.Bytes) != 0 {
		t.Fatalf("a refused render must emit no bytes, got %d", len(res.Bytes))
	}
}

// TestThaiWithoutStackedMarksStillRenders is the control. Without it the
// test above passes for a document that was never going to render.
func TestThaiWithoutStackedMarksStillRenders(t *testing.T) {
	tpl, err := ParseTemplate([]byte(thaiUnstackedTemplate))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	res, err := Render(tpl, Data("{}"), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("Thai without stacked marks must render; the refusal is about STACKING, not about Thai: %v", err)
	}
	if len(res.Bytes) == 0 {
		t.Fatal("the control rendered no bytes, so it witnesses nothing")
	}
}

// TestAnOrdinaryThaiClauseIsRefused is the arm that says why this is
// worth a story rather than a footnote.
//
// The value is a contractor-liability clause from a real Thai contract,
// reported by the owner against the shipped designer. It is not
// constructed to trip the branch — it trips it because ordinary Thai
// legal prose contains stacked marks (ทั้งสิ้น here, and ครั้ง, ทั้งนี้,
// ตั้งแต่ elsewhere). The CID is NOT pinned for this arm: a longer
// document subsets more glyphs, so which CID reports first is an
// artefact of the subset rather than a fact about the language.
func TestAnOrdinaryThaiClauseIsRefused(t *testing.T) {
	const clause = "การที่ผู้รับจ้างเป็นผู้ประกอบธุรกิจทวงถามหนี้ เป็นผู้มีความรู้ ความชำนาญ " +
		"มีความสามารถเป็นพิเศษในงานที่รับจ้าง จึงทราบและเข้าใจในพระราชบัญญัติ การทวงถามหนี้ พ.ศ. 2558 " +
		"และกฎหมายอื่น ๆ ที่เกี่ยวข้องกับกิจการงานที่มารับจ้างตามที่ระบุอยู่ในสัญญาจ้างเป็นอย่างดี " +
		"ผู้รับจ้างย่อมต้องรับผิดเป็นการส่วนตัวทั้งสิ้น"
	tpl, err := ParseTemplate([]byte(strings.Replace(thaiStackedMarksTemplate, `"ทั้งสิ้น"`, `"`+clause+`"`, 1)))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := Render(tpl, Data("{}"), nil, testShippedFontSet()); err == nil {
		t.Fatal("an ordinary Thai contract clause must reach the same refusal; got no error")
	} else if !strings.Contains(err.Error(), "carries a non-zero vertical offset") {
		t.Fatalf("the clause failed for a different reason than mark stacking: %v", err)
	}
}
