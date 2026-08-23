package folio

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// subprocessEnvVar, when set to "1", makes TestMain render the document
// to stdout and exit instead of running the test suite. This is how the
// determinism test in this file re-executes the test binary as a fresh OS
// process rather than comparing two calls inside one process (a
// same-process comparison would pass on shared memoised state, which is
// exactly what AC8 rules out).
const subprocessEnvVar = "FOLIO_SUBPROCESS_RENDER"

func TestMain(m *testing.M) {
	if os.Getenv(subprocessEnvVar) == "1" {
		b, err := Render()
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		n, werr := os.Stdout.Write(b)
		if werr != nil {
			os.Stderr.WriteString("write to stdout: " + werr.Error())
			os.Exit(1)
		}
		if n != len(b) {
			// A short write here would otherwise be indistinguishable
			// from a real determinism failure: the parent process would
			// see truncated bytes and report a byte-offset divergence
			// that points at the renderer, when the actual fault is this
			// pipe (this story's QA review, Nit 25).
			os.Stderr.WriteString("short write to stdout: wrote " + strconv.Itoa(n) + " of " + strconv.Itoa(len(b)) + " bytes")
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// assertWellFormedPDF performs toolchain-independent structural
// validation of the classic (non-stream) PDF this story's document is:
// it re-derives the document's self-described structure from its own
// bytes — no PDF-reading dependency, consistent with AC2 — and checks
// internal consistency, rather than checking for a handful of literal
// substrings.
//
// This story's QA review measured that the previous version (four
// checks: non-empty, header prefix, %%EOF suffix, exactly one
// "/Type /Page " substring) accepted a 118-byte stub with no catalog, no
// page tree, no content stream, no xref table and no trailer — every test
// that used it, including AC5's own validity test, passed on that stub
// (Major 4). It is also this function that AC8's vacuity guard 1 depends
// on to keep two identical failures from comparing equal, and — with
// Major 4's fix — it now doubles as Minor 18's PDF self-consistency
// assertions and Nit 26's fix for a page-count check that used to depend
// on an incidental trailing space rather than an actual dictionary
// boundary.
func assertWellFormedPDF(t *testing.T, label string, b []byte) {
	t.Helper()

	if len(b) == 0 {
		t.Fatalf("%s: output is empty", label)
	}
	if !bytes.HasPrefix(b, []byte("%PDF-1.7")) {
		t.Fatalf("%s: output does not start with %%PDF-1.7", label)
	}
	if !bytes.HasSuffix(b, []byte("%%EOF\n")) {
		t.Fatalf("%s: output does not end with %%%%EOF", label)
	}

	pageCount := countPageObjects(b)
	if pageCount != 1 {
		t.Fatalf("%s: expected exactly one /Type /Page object, found %d", label, pageCount)
	}

	if !bytes.Contains(b, []byte("/Type /Catalog")) {
		t.Fatalf("%s: no /Type /Catalog object found", label)
	}

	xrefOffset := assertStartxrefPointsAtXref(t, label, b)
	assertXrefEntriesPointAtTheirObjects(t, label, b, xrefOffset)
	assertStreamLengthsAreExact(t, label, b)
}

// countPageObjects counts occurrences of "/Type /Page" that are not the
// start of "/Type /Pages" — a substring match alone (as the previous
// version of this file relied on) cannot tell "/Type /Page " from
// "/Type /Page>>" apart from "/Type /Pages", except by accident of a
// trailing space (Nit 26).
func countPageObjects(b []byte) int {
	needle := []byte("/Type /Page")
	count := 0
	idx := 0
	for {
		rel := bytes.Index(b[idx:], needle)
		if rel == -1 {
			return count
		}
		pos := idx + rel
		end := pos + len(needle)
		if end < len(b) && b[end] == 's' { // "/Type /Pages", not a page object
			idx = end
			continue
		}
		count++
		idx = end
	}
}

// assertStartxrefPointsAtXref parses "startxref\n<offset>\n" and checks
// that offset genuinely indexes to the start of the "xref" keyword, and
// returns that offset.
func assertStartxrefPointsAtXref(t *testing.T, label string, b []byte) int {
	t.Helper()

	const kw = "startxref\n"
	six := bytes.LastIndex(b, []byte(kw))
	if six == -1 {
		t.Fatalf("%s: no startxref keyword found", label)
	}
	rest := b[six+len(kw):]
	eol := bytes.IndexByte(rest, '\n')
	if eol == -1 {
		t.Fatalf("%s: startxref value is not newline-terminated", label)
	}
	offsetStr := string(rest[:eol])
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		t.Fatalf("%s: startxref value %q is not an integer: %v", label, offsetStr, err)
	}
	if offset < 0 || offset >= len(b) || !bytes.HasPrefix(b[offset:], []byte("xref")) {
		t.Fatalf("%s: startxref offset %d does not point at the xref keyword", label, offset)
	}
	return offset
}

// assertXrefEntriesPointAtTheirObjects parses the classic xref table
// starting at xrefOffset and checks that every in-use entry's offset
// genuinely lands on "N 0 obj" for its own object number.
func assertXrefEntriesPointAtTheirObjects(t *testing.T, label string, b []byte, xrefOffset int) {
	t.Helper()

	const xrefKW = "xref\n"
	xrefBody := b[xrefOffset:]
	if !bytes.HasPrefix(xrefBody, []byte(xrefKW)) {
		t.Fatalf("%s: xref section does not start with %q", label, xrefKW)
	}
	xrefBody = xrefBody[len(xrefKW):]

	headerEnd := bytes.IndexByte(xrefBody, '\n')
	if headerEnd == -1 {
		t.Fatalf("%s: xref subsection header is not newline-terminated", label)
	}
	fields := strings.Fields(string(xrefBody[:headerEnd]))
	if len(fields) != 2 {
		t.Fatalf("%s: xref subsection header %q does not have two fields", label, xrefBody[:headerEnd])
	}
	start, err1 := strconv.Atoi(fields[0])
	count, err2 := strconv.Atoi(fields[1])
	if err1 != nil || err2 != nil || start != 0 || count < 1 {
		t.Fatalf("%s: unexpected xref subsection header %q", label, xrefBody[:headerEnd])
	}

	entries := xrefBody[headerEnd+1:]
	for i := 0; i < count; i++ {
		entryStart := i * 20
		if entryStart+20 > len(entries) {
			t.Fatalf("%s: xref entry %d is truncated", label, i)
		}
		entry := entries[entryStart : entryStart+20]
		if entry[10] != ' ' || entry[16] != ' ' || entry[18] != ' ' || entry[19] != '\n' {
			t.Fatalf("%s: xref entry %d %q is not exactly 20 bytes in the expected shape", label, i, entry)
		}
		offStr := string(entry[0:10])
		genStr := string(entry[11:16])
		kind := entry[17]
		off, oerr := strconv.Atoi(offStr)
		_, gerr := strconv.Atoi(genStr)
		if oerr != nil || gerr != nil {
			t.Fatalf("%s: xref entry %d %q has a non-numeric offset or generation field", label, i, entry)
		}
		switch kind {
		case 'f':
			if i != 0 {
				t.Fatalf("%s: xref entry %d is a free entry, but only object 0 should be", label, i)
			}
		case 'n':
			if i == 0 {
				t.Fatalf("%s: xref entry 0 must be the free entry, got in-use", label)
			}
			want := strconv.Itoa(i) + " 0 obj"
			if off < 0 || off+len(want) > len(b) || string(b[off:off+len(want)]) != want {
				t.Fatalf("%s: xref entry %d claims offset %d, but %q was not found there", label, i, off, want)
			}
		default:
			t.Fatalf("%s: xref entry %d has an unrecognised entry kind %q", label, i, string(kind))
		}
	}
}

// assertStreamLengthsAreExact finds every "/Length N" declaration
// followed by a stream body and checks that the body is exactly N bytes
// long, ending at a literal "endstream".
func assertStreamLengthsAreExact(t *testing.T, label string, b []byte) {
	t.Helper()

	found := 0
	idx := 0
	for {
		const kw = "/Length "
		rel := bytes.Index(b[idx:], []byte(kw))
		if rel == -1 {
			break
		}
		pos := idx + rel + len(kw)
		end := pos
		for end < len(b) && b[end] >= '0' && b[end] <= '9' {
			end++
		}
		if end == pos {
			t.Fatalf("%s: /Length at offset %d is not followed by digits", label, pos)
		}
		declared, err := strconv.Atoi(string(b[pos:end]))
		if err != nil {
			t.Fatalf("%s: /Length value %q is not an integer", label, b[pos:end])
		}

		const streamKW = "stream\n"
		streamRel := bytes.Index(b[end:], []byte(streamKW))
		if streamRel == -1 || streamRel > 40 {
			t.Fatalf("%s: no %q found shortly after /Length %d", label, streamKW, declared)
		}
		bodyStart := end + streamRel + len(streamKW)
		bodyEnd := bodyStart + declared
		if bodyEnd > len(b) {
			t.Fatalf("%s: /Length %d overruns the document", label, declared)
		}
		if !bytes.HasPrefix(b[bodyEnd:], []byte("endstream")) {
			t.Fatalf("%s: stream body does not end with endstream where /Length %d says it should", label, declared)
		}
		found++
		idx = bodyEnd
	}
	if found == 0 {
		t.Fatalf("%s: no stream object found to validate /Length against", label)
	}
}

// renderInSubprocess runs this same test binary as a fresh OS process with
// FOLIO_SUBPROCESS_RENDER=1 set, and the given GOMAXPROCS, and returns
// what it wrote to stdout.
func renderInSubprocess(t *testing.T, gomaxprocs string) []byte {
	t.Helper()

	// -test.run=^$ explicitly selects no test function to run; the
	// subprocessEnvVar check inside TestMain above is what actually
	// routes the child process. TestMain is not itself a test function,
	// so "-test.run=^TestMain$" (the previous flag here) matched nothing
	// either — harmlessly, since the env-var short-circuit runs before
	// m.Run() either way — but it named the wrong mechanism (this
	// story's QA review, Nit 24).
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(),
		subprocessEnvVar+"=1",
		"GOMAXPROCS="+gomaxprocs,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("subprocess render (GOMAXPROCS=%s) failed: %v\nstderr: %s", gomaxprocs, err, stderr.String())
	}
	return stdout.Bytes()
}

// firstDivergence returns the byte offset of the first byte at which a and
// b differ, and a short hex window around it, for diagnosing determinism
// failures.
func firstDivergence(a, b []byte) (offset int, window string) {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i, hexWindow(a, b, i)
		}
	}
	if len(a) != len(b) {
		return n, hexWindow(a, b, n)
	}
	return -1, ""
}

func hexWindow(a, b []byte, at int) string {
	const radius = 8
	start := at - radius
	if start < 0 {
		start = 0
	}
	end := func(x []byte) int {
		e := at + radius
		if e > len(x) {
			e = len(x)
		}
		return e
	}
	var sb strings.Builder
	sb.WriteString("a=")
	sb.WriteString(hexOf(a[start:end(a)]))
	sb.WriteString(" b=")
	sb.WriteString(hexOf(b[start:end(b)]))
	return sb.String()
}

func hexOf(b []byte) string {
	const hexDigits = "0123456789abcdef"
	out := make([]byte, 0, len(b)*2)
	for _, c := range b {
		out = append(out, hexDigits[c>>4], hexDigits[c&0x0f])
	}
	return string(out)
}

// TestRenderIsByteIdenticalAcrossTwoProcesses is AC8: the same input
// rendered by two independently started OS processes produces
// byte-identical output. Child A runs with GOMAXPROCS=1 and child B with
// GOMAXPROCS=4, so the comparison also catches order-dependence introduced
// by concurrency (e.g. an accidental map iteration reaching an output
// byte, NFR1.d).
func TestRenderIsByteIdenticalAcrossTwoProcesses(t *testing.T) {
	a := renderInSubprocess(t, "1")
	b := renderInSubprocess(t, "4")

	assertWellFormedPDF(t, "child A (GOMAXPROCS=1)", a)
	assertWellFormedPDF(t, "child B (GOMAXPROCS=4)", b)

	if !bytes.Equal(a, b) {
		offset, window := firstDivergence(a, b)
		t.Fatalf("subprocess outputs differ (len a=%d, len b=%d); first divergence at byte offset %d; %s",
			len(a), len(b), offset, window)
	}
}

// TestRenderProducesAValidPDF exercises AC5 directly in-process: the
// output is non-empty, has exactly one page, and passes the well-formed
// structural checks above. Independent-reader validation (qpdf --check)
// is recorded in the Delivery Log rather than wired into this test, since
// this module intentionally has no PDF-reading dependency (AC2).
func TestRenderProducesAValidPDF(t *testing.T) {
	b, err := Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	assertWellFormedPDF(t, "Render()", b)
}

// TestRenderHasNoCreationOrModDate is AC9: neither key appears anywhere in
// the output, and no /Info dictionary is emitted at all. It also checks
// the neighbouring metadata/compression keys the Dev Notes and fixture
// README promise are absent but that nothing previously asserted:
// /Producer, /Creator and /Metadata (XMP) are the next places a
// timestamp, tool name or machine name could leak into the bytes — the
// same AD-7 hazard AC9 exists to close, one key over — and /Filter is the
// R4 compression risk this story's Dev Notes explicitly remove from the
// variable set (this story's QA review, Minor 20).
func TestRenderHasNoCreationOrModDate(t *testing.T) {
	b, err := Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	forbidden := []string{"/CreationDate", "/ModDate", "/Info", "/Producer", "/Creator", "/Metadata", "/Filter"}
	for _, key := range forbidden {
		if bytes.Contains(b, []byte(key)) {
			t.Errorf("output contains %s", key)
		}
	}
}

// TestRenderIDEntriesAreIdentical is AC10's identity half: both /ID array
// entries are byte-identical to each other and 16 bytes each (32 hex
// characters).
func TestRenderIDEntriesAreIdentical(t *testing.T) {
	b, err := Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	idx := bytes.Index(b, []byte("/ID ["))
	if idx == -1 {
		t.Fatal("output does not contain an /ID array")
	}
	rest := b[idx+len("/ID ["):]

	first, rest, ok := extractAngleBracketed(rest)
	if !ok {
		t.Fatal("could not parse first /ID entry")
	}
	second, _, ok := extractAngleBracketed(rest)
	if !ok {
		t.Fatal("could not parse second /ID entry")
	}

	if len(first) != 32 {
		t.Errorf("first /ID entry is %d hex characters, want 32 (16 bytes)", len(first))
	}
	if len(second) != 32 {
		t.Errorf("second /ID entry is %d hex characters, want 32 (16 bytes)", len(second))
	}
	if first != second {
		t.Errorf("/ID entries differ: %q != %q", first, second)
	}
}

func extractAngleBracketed(b []byte) (content string, rest []byte, ok bool) {
	if len(b) == 0 || b[0] != '<' {
		return "", b, false
	}
	end := bytes.IndexByte(b, '>')
	if end == -1 {
		return "", b, false
	}
	return string(b[1:end]), b[end+1:], true
}
