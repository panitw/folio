package folio_test

// EVERY COMMITTED PDF GOLDEN MUST BE A PDF.
//
// WHY THIS FILE EXISTS, and it is worth more than the one-character bug that
// produced it.
//
// Story 2.6 recorded fixtures/multi-page/expected.pdf. The file PASSED its
// golden hash and PASSED the story's whole semantic acceptance step — page
// count at two independent sites, per-page header and footer placement
// against hand-derived literals, a pinned line->page partition, no text at a
// negative Y, byte-identity across two processes. And it was NOT A PDF: it
// emitted
//
//	<< /Type /Pages /Kids [8 0 R10 0 R] /Count 2 >>
//
// with no separator between the two indirect references. A PDF tokenizer
// reads "R10" as one unknown token, so NEITHER kid resolves, the page tree is
// empty, qpdf reports "file does not contain any pages", and a viewer shows
// "page 0 of 2".
//
// EVERY ONE OF THOSE ACCEPTANCE CHECKS WAS SATISFIED BY THE BROKEN FILE, and
// the reason is worth stating exactly:
//
//   - "/Type /Page objects == 2" counted OBJECT DEFINITIONS. Both page
//     objects were emitted, correctly, and both were reachable in the xref.
//     What was broken was the /Kids ARRAY that points at them — a different
//     thing, in a different object.
//   - "/Count 2" read the integer. The integer was right. The integer is a
//     CLAIM about the kids, not a fact derived from them.
//   - Every per-page assertion parsed the CONTENT STREAMS directly, by
//     splitting the file on stream boundaries. Content streams do not care
//     whether anything references them, so header/footer placement, the line
//     partition and the negative-Y check all read exactly as they would in a
//     healthy file.
//   - Byte-identity across two processes says the output is DETERMINISTIC. A
//     deterministically wrong file is byte-identical to itself.
//
// The common shape: every check asserted a property of a PART, and the defect
// was in a REFERENCE BETWEEN parts. A hash certifies that bytes have not
// changed since they were recorded; it says nothing whatever about whether
// the bytes were right when they were recorded. D-000.22's semantic
// acceptance step is supposed to be what covers that gap — and here it did
// not, because it was assembled from checks that were all true of an
// unrenderable document.
//
// SO THE REMEDY IS NOT "one more assertion about multi-page". It is a
// STRUCTURAL VALIDITY ORACLE applied to EVERY committed golden, asserting the
// property no per-part check can see: THE DOCUMENT RESOLVES. This test is
// hermetic — it parses the page tree itself rather than shelling out — so it
// runs on every target and in every story's plain `go test ./...`, not only
// where qpdf happens to be installed.
//
// PROPOSED STANDING RULE, recorded here and raised in the story's Delivery
// Log rather than assumed: *no recorded PDF golden's hash may be trusted
// until the artifact has passed a structural validity oracle.* A recording is
// an act of acceptance (D-000.44), and accepting bytes nobody has proven are
// a document is the failure this file was written after.

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

// kidsArrayRE captures the whole /Kids array body of the page tree.
var kidsArrayRE = regexp.MustCompile(`/Type /Pages /Kids \[([^\]]*)\] /Count (\d+)`)

// wellFormedRefRE matches ONE indirect reference, anchored, so a run-together
// pair like "8 0 R10 0 R" cannot match as a single reference.
var wellFormedRefRE = regexp.MustCompile(`^(\d+) 0 R$`)

// definedPageObjRE captures the object number of every "N 0 obj\n<< /Type
// /Page /Parent " definition — used to build the SET of defined page-object
// numbers for step (4)'s orphan check (finisher Finding 4: a COUNT of these
// occurrences cannot distinguish "every defined page is referenced" from
// "the counts happen to agree while one definition sits outside the
// referenced set and one reference is duplicated").
var definedPageObjRE = regexp.MustCompile(`(\d+) 0 obj\n<< /Type /Page /Parent `)

// TestEveryGoldenPDFResolvesItsPageTree is the structural oracle.
//
// It asserts, for every fixture declared in goldenDigestRecord:
//
//  1. The page tree exists and declares a /Count.
//  2. The /Kids array splits into exactly /Count references, each one
//     WELL-FORMED — this is the assertion that would have caught the bug.
//     "[8 0 R10 0 R]" splits into ONE token that matches no reference at all.
//  3. Every referenced object number is actually DEFINED in the file as
//     "N 0 obj" and carries /Type /Page, and the referenced numbers are
//     pairwise DISTINCT. A syntactically valid reference to a non-existent
//     object is still an empty page tree, and a duplicated reference is how
//     an orphaned page hides behind a correct-looking reference count.
//  4. The SET of referenced page-object numbers equals the SET of /Type
//     /Page definitions in the file — so a page object that exists but is
//     unreachable, which is the general shape of this defect, fails here
//     too. CORRECTED by the finisher (Finding 4): this used to compare
//     COUNTS, not SETS, and a probe — "[8 0 R 8 0 R]" with objects 8 and 10
//     both defined as /Type /Page — passed: len(refs) == declaredCount ==
//     2, both references resolve, and the defined-object COUNT equals 2,
//     while object 10 sits outside the referenced set entirely and page 2
//     is a silent duplicate of page 1.
func TestEveryGoldenPDFResolvesItsPageTree(t *testing.T) {
	root := repoRootForByteNeutrality(t)

	if len(goldenDigestRecord) == 0 {
		t.Fatal("vacuity guard: the golden declaration is empty, so this test checks nothing")
	}

	checked := 0
	for _, fx := range goldenDigestRecord {
		t.Run(fx.dir, func(t *testing.T) {
			path := filepath.Join(root, "fixtures", fx.dir, "expected.pdf")
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("presence precondition: %s could not be read: %v", path, err)
			}
			if len(b) == 0 {
				t.Fatal("presence precondition: the golden is empty")
			}
			if !bytes.HasPrefix(b, []byte("%PDF-")) {
				t.Fatalf("presence precondition: %s does not begin with %%PDF-", path)
			}

			// (1) The page tree.
			m := kidsArrayRE.FindSubmatch(b)
			if m == nil {
				t.Fatalf("no page tree of the form \"/Type /Pages /Kids [...] /Count N\" found in %s — the document has no page tree at all", path)
			}
			kidsBody := string(m[1])
			declaredCount, cerr := strconv.Atoi(string(m[2]))
			if cerr != nil || declaredCount < 1 {
				t.Fatalf("the page tree declares /Count %q, which is not a positive integer", m[2])
			}

			// (2) THE ASSERTION THAT WOULD HAVE CAUGHT THE BUG. Split the
			// array on whitespace-separated reference triples and require
			// each to be well formed. A missing separator collapses two
			// references into one unparseable token.
			refs := splitIndirectRefs(kidsBody)
			if len(refs) != declaredCount {
				t.Fatalf("the page tree declares /Count %d but its /Kids array yields %d reference(s): %q\n\n"+
					"This is the failure mode a hash CANNOT see and a page-object COUNT cannot see: both page "+
					"objects may exist and be correct while the array pointing at them is unparseable. "+
					"\"[8 0 R10 0 R]\" tokenizes as one unknown token, so NEITHER kid resolves and the page tree "+
					"is EMPTY — the /Count is a claim about the kids, not a fact derived from them.",
					declaredCount, len(refs), kidsBody)
			}
			for _, ref := range refs {
				if !wellFormedRefRE.MatchString(ref) {
					t.Errorf("the /Kids array contains %q, which is not a well-formed indirect reference (\"N 0 R\")", ref)
				}
			}
			if t.Failed() {
				return
			}

			// (3) Every reference resolves to a defined page object, AND
			// the referenced object numbers are pairwise DISTINCT.
			//
			// Distinctness is asserted here rather than left to (4)'s
			// count comparison, per the finisher's Finding 4: "[8 0 R 8 0
			// R]", with objects 8 and 10 both defined as /Type /Page, has
			// len(refs) == declaredCount == 2 and every reference resolves
			// — object 10 is a defined page object nothing references,
			// and page 2 is a silent duplicate of page 1. A COUNT
			// comparison cannot see this; a SET can.
			refNums := map[string]bool{}
			for _, ref := range refs {
				num := wellFormedRefRE.FindStringSubmatch(ref)[1]
				if refNums[num] {
					t.Errorf("the /Kids array references object %s more than once — a duplicated reference is how an orphaned page object stays invisible to a COUNT of referenced objects, even though this reference itself resolves fine", num)
					continue
				}
				refNums[num] = true

				def := []byte("\n" + num + " 0 obj\n")
				idx := bytes.Index(b, def)
				if idx < 0 {
					t.Errorf("the page tree references object %s, which is never defined in the file — a syntactically valid reference to a non-existent object is still an empty page tree", num)
					continue
				}
				rest := b[idx+len(def):]
				end := bytes.Index(rest, []byte("\nendobj"))
				if end < 0 {
					t.Errorf("object %s has no endobj", num)
					continue
				}
				if !bytes.Contains(rest[:end], []byte("/Type /Page ")) {
					t.Errorf("the page tree references object %s, which is not a /Type /Page", num)
				}
			}
			if t.Failed() {
				return
			}

			// (4) No page object is orphaned — compared as a SET against
			// the SET of definitions, not as a count against a count
			// (finisher Finding 4). bytes.Count("<< /Type /Page /Parent
			// ") == declaredCount is satisfied by [8 0 R 8 0 R] with
			// objects 8 AND 10 both defined: the count matches while
			// object 10 sits outside the referenced set entirely.
			definedNums := map[string]bool{}
			for _, m := range definedPageObjRE.FindAllSubmatch(b, -1) {
				definedNums[string(m[1])] = true
			}
			for num := range definedNums {
				if !refNums[num] {
					t.Errorf("object %s is defined as a /Type /Page but the page tree's /Kids array never references it — a page object that exists and is unreachable is invisible to a reader and to a hash alike, even when the page-object COUNT agrees with /Count", num)
				}
			}
			if len(definedNums) != len(refNums) {
				t.Errorf("the file defines %d distinct /Type /Page object(s) but the page tree references %d distinct one(s)", len(definedNums), len(refNums))
			}
			checked++
		})
	}

	if checked == 0 {
		t.Error("vacuity guard: no golden was actually checked")
	}
	t.Logf("structural validity: %d committed PDF goldens parsed; every page tree resolves", checked)
}

// splitIndirectRefs splits a /Kids array body into its whitespace-delimited
// reference triples: "8 0 R 10 0 R" -> ["8 0 R", "10 0 R"].
//
// It groups tokens in THREES rather than searching for "R", because grouping
// is what exposes the defect: "8 0 R10 0 R" tokenizes as
// ["8", "0", "R10", "0", "R"] — five tokens, which is not a multiple of
// three, and whose first triple "8 0 R10" is not a well-formed reference.
// A search for "R" would happily find two of them and report success.
func splitIndirectRefs(body string) []string {
	fields := bytes.Fields([]byte(body))
	var out []string
	for i := 0; i+2 < len(fields); i += 3 {
		out = append(out, string(fields[i])+" "+string(fields[i+1])+" "+string(fields[i+2]))
	}
	// A trailing partial group is itself a malformation; surface it as an
	// extra (unparseable) entry rather than discarding it, so the count
	// comparison above fails loudly instead of silently rounding down.
	if rem := len(fields) % 3; rem != 0 {
		var tail string
		for _, f := range fields[len(fields)-rem:] {
			tail += string(f) + " "
		}
		out = append(out, tail)
	}
	return out
}
