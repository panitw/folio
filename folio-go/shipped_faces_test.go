package folio

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf16"
)

// ---------------------------------------------------------------------
// D-000.22's semantic acceptance step, driven by a DECLARATIVE SPEC.
// ---------------------------------------------------------------------
//
// A hash guards against CHANGE. It says "this is the same as what we
// recorded", and it has nothing whatever to say about whether what we
// recorded was RIGHT — which matters because the moment a golden is
// committed it stops being an output and becomes an input, and every
// guard downstream then agrees with it, correctly and forever. A wrong
// first recording is not a bug that gets caught later; it is a bug that
// gets RATIFIED later. So the first recording is the only moment at
// which "is this right?" is answerable at all.
//
// Story 2.2 needed that step and did not have it. It shipped Simplified
// Chinese as THIN: NotoSansSC-VF's `wght` axis has default=100 (alone
// among the three faces), and the render seam pinned every axis to its
// default. Three guards agreed the artifact was correct — four-target
// byte-identity, the font-program digest, and the missing-glyph
// diagnostic — and all three were RIGHT. Each answered its own question
// accurately. None of them was asked whether the value MEANT what its
// name implied (D-000.21).
//
// WHY A SPEC TABLE RATHER THAN ASSERTIONS WRITTEN OUT PER FACE. Writing
// the checks face-by-face means the future Bold story adds faces and
// inherits none of them — which is precisely how a guard written for
// today's faces fails to cover tomorrow's. Here there is ONE record per
// shipped face and the assertion is "spec equals artifact", so a new
// face is covered by construction the moment it is added, and
// TestShippedSpecCoversEverythingShipped fails if someone adds a face
// to fonts.Shipped() and forgets the spec row.
//
// WHICH ARTIFACT CARRIES WHICH PROPERTY — this split is load-bearing and
// was got wrong once during this story. The embedded font programs in
// the PDF carry NO `name` table: textshape's subsetter lists `name` in
// `optionalTables` (subset/execute.go:531-545) and copies it only on
// explicit request, which nothing makes. That is not a defect — PDF
// §9.9.2 sanctions the reduced table set for an embedded CIDFontType2,
// because /CIDToGIDMap does the mapping. Asserting `name[1]`, `name[2]`,
// `name[6]` or "no record contains Thin" against the embedded program
// would therefore pass VACUOUSLY on all four counts: there are no name
// records, so no record contains anything.
//
//	property                                  | read off
//	------------------------------------------|---------------------------
//	name[1], name[2], name[6], table set,     | the shipped .ttf face file
//	usWeightClass                             |
//	usWeightClass, fvar/gvar absent           | the embedded program
//	/BaseFont == ^[A-Z]{6}\+<name6>$          | the produced PDF
//
// Hence the rule this file is written under, in its sharpened form:
// ASSERT ON THE ARTIFACT THAT CARRIES THE PROPERTY, AND PROVE IT CARRIES
// IT. Every assertion group below begins with a PRESENCE PRECONDITION —
// the table must exist before anything is claimed about its contents.
// That precondition is not defensive decoration; it is the half of the
// rule that stops it collapsing into a vacuous pass wherever a
// downstream stage has legitimately stripped the field.
//
// Note also what is NOT counted as coverage here. "No name record
// contains 'Thin'" is kept as belt-and-braces at the bottom of
// assertShippedFaceMatchesSpec, but it is keyed on the exact string that
// burned us: a face defaulting to Light, ExtraLight or Black sails
// straight past it. THE POSITIVE ASSERTIONS ARE THE GUARD — name[1]
// equals the exact family, name[2] == "Regular", name[6] ends
// "-Regular", usWeightClass == 400. A denylist entry must never be
// counted as coverage.

// shippedFaceSpec is one record per shipped face: what folio INTENDS
// that face to be. Every field is checked against the produced artifact
// rather than against the request that produced it (D-000.21) — these
// values are not passed to the generator, they are compared with what
// the generator emitted.
type shippedFaceSpec struct {
	// Key is the FontSet key, i.e. the name a .folio document's `fonts`
	// fallback chain references.
	Key string
	// Dir and File locate the committed face under folio-go/fonts/.
	Dir  string
	File string
	// Family is name record 1, exactly.
	Family string
	// Subfamily is name record 2, exactly. Regular-only for now, on
	// purpose: fontset.go exposes no way to request a non-default
	// instance, so a Bold face would be selectable by nothing (D-2.2.4).
	Subfamily string
	// PostScriptName is name record 6, exactly. This is also what
	// /BaseFont must carry after its six-letter subset tag.
	PostScriptName string
	// WeightClass is OS/2.usWeightClass, exactly. THE field this whole
	// apparatus exists for.
	WeightClass uint16
	// AxisPins records the instancer arguments the face was derived
	// with, for the reader's benefit. It is deliberately NOT asserted
	// against anything: it describes the request, and asserting the
	// request against itself is the defect D-000.21 names.
	AxisPins string
}

// shippedFaceSpecs is the whole intended shipped set. Adding a face here
// without adding it to fonts.Shipped() (or the reverse) fails
// TestShippedSpecCoversEverythingShipped in both directions.
var shippedFaceSpecs = []shippedFaceSpec{
	{
		Key: "Noto Sans", Dir: "notosans", File: "NotoSans-Regular.ttf",
		Family: "Noto Sans", Subfamily: "Regular",
		PostScriptName: "NotoSans-Regular", WeightClass: 400,
		AxisPins: "wght=400 wdth=100",
	},
	{
		Key: "Noto Sans Thai", Dir: "notosansthai", File: "NotoSansThai-Regular.ttf",
		Family: "Noto Sans Thai", Subfamily: "Regular",
		PostScriptName: "NotoSansThai-Regular", WeightClass: 400,
		AxisPins: "wght=400 wdth=100",
	},
	{
		// ONE axis. This is the face whose `wght` default is 100 — the
		// one D-000.21 was written about.
		Key: "Noto Sans SC", Dir: "notosanssc", File: "NotoSansSC-Regular.ttf",
		Family: "Noto Sans SC", Subfamily: "Regular",
		PostScriptName: "NotoSansSC-Regular", WeightClass: 400,
		AxisPins: "wght=400",
	},
}

// ---------------------------------------------------------------------
// sfnt reading, with presence preconditions throughout.
// ---------------------------------------------------------------------

// sfntTables returns the table directory of an sfnt as tag -> contents.
// It fails the test rather than returning an error, because every caller
// here is asserting about a font it has just been handed and a font it
// cannot parse is a test failure, never a skip.
func sfntTables(t *testing.T, label string, data []byte) map[string][]byte {
	t.Helper()
	if len(data) < 12 {
		t.Fatalf("%s: not an sfnt — only %d bytes", label, len(data))
	}
	switch string(data[0:4]) {
	case "\x00\x01\x00\x00", "true", "OTTO":
	default:
		t.Fatalf("%s: not an sfnt — bad magic %q", label, data[0:4])
	}
	numTables := int(binary.BigEndian.Uint16(data[4:6]))
	if numTables == 0 {
		t.Fatalf("%s: sfnt declares zero tables", label)
	}
	if len(data) < 12+numTables*16 {
		t.Fatalf("%s: sfnt truncated: %d tables declared, %d bytes total", label, numTables, len(data))
	}
	out := make(map[string][]byte, numTables)
	for i := 0; i < numTables; i++ {
		rec := data[12+i*16:]
		tag := string(rec[0:4])
		off := binary.BigEndian.Uint32(rec[8:12])
		length := binary.BigEndian.Uint32(rec[12:16])
		if int(off)+int(length) > len(data) {
			t.Fatalf("%s: table %q runs past end of file (off %d len %d, file %d)", label, tag, off, length, len(data))
		}
		out[tag] = data[off : off+length]
	}
	return out
}

// sfntWeightClass reads OS/2.usWeightClass (offset 4). PRESENCE
// PRECONDITION: the OS/2 table must exist and be long enough. A font
// with no OS/2 table has no weight class at all, and reporting 0 (or
// silently passing) would be exactly the vacuous-pass failure this file
// exists to prevent.
func sfntWeightClass(t *testing.T, label string, tables map[string][]byte) uint16 {
	t.Helper()
	os2, ok := tables["OS/2"]
	if !ok {
		t.Fatalf(
			"%s: PRESENCE PRECONDITION FAILED — no OS/2 table, so usWeightClass cannot be read at all. "+
				"Any weight assertion against this artifact would pass vacuously.", label,
		)
	}
	if len(os2) < 6 {
		t.Fatalf("%s: OS/2 table is %d bytes, too short to contain usWeightClass", label, len(os2))
	}
	return binary.BigEndian.Uint16(os2[4:6])
}

// sfntNameRecords returns every `name` table record as (nameID -> value),
// preferring the Windows/Unicode (platform 3) entry. PRESENCE
// PRECONDITION: the `name` table must exist and declare at least one
// record — see this file's header for why that precondition is the
// whole point rather than a courtesy.
func sfntNameRecords(t *testing.T, label string, tables map[string][]byte) map[uint16]string {
	t.Helper()
	nameTable, ok := tables["name"]
	if !ok {
		t.Fatalf(
			"%s: PRESENCE PRECONDITION FAILED — this artifact has NO `name` table, so every assertion "+
				"about name[1], name[2], name[6] or 'no record contains Thin' would pass VACUOUSLY. "+
				"Read name records off the shipped .ttf face file, not off an embedded PDF font program "+
				"(textshape drops `name`; PDF §9.9.2 permits it).", label,
		)
	}
	if len(nameTable) < 6 {
		t.Fatalf("%s: `name` table is %d bytes, too short for a header", label, len(nameTable))
	}
	count := int(binary.BigEndian.Uint16(nameTable[2:4]))
	stringOffset := int(binary.BigEndian.Uint16(nameTable[4:6]))
	if count == 0 {
		t.Fatalf("%s: PRESENCE PRECONDITION FAILED — `name` table declares zero records", label)
	}
	out := map[uint16]string{}
	for i := 0; i < count; i++ {
		off := 6 + i*12
		if off+12 > len(nameTable) {
			t.Fatalf("%s: `name` table truncated at record %d of %d", label, i, count)
		}
		rec := nameTable[off : off+12]
		platformID := binary.BigEndian.Uint16(rec[0:2])
		nameID := binary.BigEndian.Uint16(rec[6:8])
		length := int(binary.BigEndian.Uint16(rec[8:10]))
		strOff := int(binary.BigEndian.Uint16(rec[10:12]))
		start := stringOffset + strOff
		if start+length > len(nameTable) {
			continue
		}
		raw := nameTable[start : start+length]
		var value string
		if platformID == 3 || platformID == 0 {
			if length%2 != 0 {
				continue
			}
			u16 := make([]uint16, 0, length/2)
			for j := 0; j+1 < length; j += 2 {
				u16 = append(u16, binary.BigEndian.Uint16(raw[j:j+2]))
			}
			value = string(utf16.Decode(u16))
		} else {
			value = string(raw)
		}
		// Prefer platform 3 (Windows/Unicode); do not let a Macintosh
		// record shadow it.
		if existing, seen := out[nameID]; seen && platformID != 3 {
			_ = existing
			continue
		}
		out[nameID] = value
	}
	return out
}

// sfntAllNameValues returns every decoded name value, for the
// belt-and-braces denylist check only.
func sfntAllNameValues(t *testing.T, label string, tables map[string][]byte) []string {
	t.Helper()
	recs := sfntNameRecords(t, label, tables)
	out := make([]string, 0, len(recs))
	for _, v := range recs {
		out = append(out, v)
	}
	return out
}

// ---------------------------------------------------------------------
// The assertions.
// ---------------------------------------------------------------------

// assertShippedFaceMatchesSpec is the whole semantic acceptance step for
// ONE face, read off the shipped .ttf — the artifact the generator
// produced, never the flags it was handed (D-000.21).
func assertShippedFaceMatchesSpec(t *testing.T, spec shippedFaceSpec, data []byte) {
	t.Helper()
	label := spec.Key + " (" + spec.File + ")"
	tables := sfntTables(t, label, data)

	// --- outline format and staticness, off the table directory ---
	if _, ok := tables["glyf"]; !ok {
		t.Errorf("%s: NFR7 requires a glyf/TrueType outline build, but there is no `glyf` table", label)
	}
	if _, ok := tables["CFF2"]; ok {
		t.Errorf("%s: NFR7 takes glyf over CFF, but this face carries a `CFF2` table", label)
	}
	for _, tag := range []string{"fvar", "gvar", "avar"} {
		if _, ok := tables[tag]; ok {
			t.Errorf(
				"%s: still carries a %q table, so it is NOT a static instance. "+
					"D-2.2.4 ships statics precisely so the float `gvar` interpolation path is unreachable "+
					"from the render; a face arriving here with %q means the instancer did not pin every "+
					"axis the face declares (Latin and Thai have TWO: wght AND wdth).",
				label, tag, tag,
			)
		}
	}

	// --- the weight class, off OS/2, presence-precondition first ---
	if got := sfntWeightClass(t, label, tables); got != spec.WeightClass {
		t.Errorf(
			"%s: OS/2.usWeightClass is %d, want %d.\n"+
				"THIS IS THE ASSERTION STORY 2.2 DID NOT HAVE. Asking for Regular and getting Regular are "+
				"two different facts (D-000.21); %s's `wght` axis default is not necessarily 400.",
			label, got, spec.WeightClass, spec.Family,
		)
	}

	// --- the names, off `name`, presence-precondition first ---
	names := sfntNameRecords(t, label, tables)
	if got := names[1]; got != spec.Family {
		t.Errorf("%s: name[1] (family) is %q, want exactly %q", label, got, spec.Family)
	}
	if got := names[2]; got != spec.Subfamily {
		t.Errorf(
			"%s: name[2] (subfamily) is %q, want exactly %q. Without fontTools'"+
				" --update-name-table the instanced Regular keeps naming itself after the axis DEFAULT"+
				" (for Noto Sans SC that is 'Thin').",
			label, got, spec.Subfamily,
		)
	}
	if got := names[6]; got != spec.PostScriptName {
		t.Errorf("%s: name[6] (PostScript name) is %q, want exactly %q", label, got, spec.PostScriptName)
	}
	if !strings.HasSuffix(names[6], "-Regular") {
		t.Errorf("%s: name[6] is %q, which does not end \"-Regular\"", label, names[6])
	}

	// --- belt-and-braces only; NOT counted as coverage ---
	//
	// This is keyed on the exact string that burned us. A face
	// defaulting to Light, ExtraLight or Black passes it while being
	// just as wrong. The positive assertions above are the guard.
	for _, v := range sfntAllNameValues(t, label, tables) {
		if strings.Contains(v, "Thin") {
			t.Errorf(
				"%s: a name record still reads %q. (Belt-and-braces check only — the positive "+
					"name[1]/name[2]/name[6] and usWeightClass assertions above are what actually guard this.)",
				label, v,
			)
		}
	}
}

// TestShippedFacesMatchSpec is D-000.22's semantic acceptance step over
// every shipped face, read off the committed .ttf files themselves.
func TestShippedFacesMatchSpec(t *testing.T) {
	root := repoRootFromTest(t)
	checked := 0
	for _, spec := range shippedFaceSpecs {
		t.Run(spec.Key, func(t *testing.T) {
			path := filepath.Join(root, "folio-go", "fonts", spec.Dir, spec.File)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read shipped face %s: %v", path, err)
			}
			if len(data) == 0 {
				t.Fatalf("shipped face %s is EMPTY — a comparison against it would be vacuous", path)
			}
			assertShippedFaceMatchesSpec(t, spec, data)
		})
		checked++
	}
	// Coverage witness (D-000.9): a run that examined nothing must not
	// report the same "no failures" a healthy run reports.
	if checked != len(shippedFaceSpecs) {
		t.Fatalf("coverage witness: checked %d of %d shipped faces", checked, len(shippedFaceSpecs))
	}
	if checked == 0 {
		t.Fatal("coverage witness: the shipped-face spec is EMPTY, so this test asserted nothing")
	}
}

// TestShippedSpecCoversEverythingShipped closes the drift channel that
// made the golden-count guard vacuous: the spec and the set actually
// embedded must agree in BOTH directions, by key and by bytes.
//
// testShippedFontSet() is a hand-copy of fonts.Shipped() (the import
// would be a cycle — `fonts` imports `folio`). Comparing it to the spec
// here, and the spec to fonts.Shipped() in shipped_faces_ext_test.go,
// chains the two together: a face added to fonts.Shipped() and nowhere
// else now fails, instead of silently never reaching the matrix fixture.
func TestShippedSpecCoversEverythingShipped(t *testing.T) {
	root := repoRootFromTest(t)
	embedded := testShippedFontSet()

	if len(embedded) != len(shippedFaceSpecs) {
		t.Fatalf(
			"testShippedFontSet() carries %d faces but shippedFaceSpecs has %d rows. "+
				"Every shipped face needs a spec row, or its weight/name/staticness is asserted by nothing "+
				"(D-2.2.1's V7 hazard: the golden set covering fewer pairs than actually ship).",
			len(embedded), len(shippedFaceSpecs),
		)
	}
	for _, spec := range shippedFaceSpecs {
		got, ok := embedded[spec.Key]
		if !ok {
			t.Fatalf("shippedFaceSpecs names face %q, which testShippedFontSet() does not carry", spec.Key)
		}
		if len(got) == 0 {
			t.Fatalf("face %q embedded as ZERO bytes — any comparison against it is vacuous", spec.Key)
		}
		path := filepath.Join(root, "folio-go", "fonts", spec.Dir, spec.File)
		onDisk, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if len(onDisk) != len(got) {
			t.Fatalf(
				"face %q: the embedded bytes (%d) differ in length from %s (%d) — the go:embed directive "+
					"and the spec row are naming different files",
				spec.Key, len(got), path, len(onDisk),
			)
		}
	}
	// The other direction: nothing embedded may be missing from the spec.
	specKeys := map[string]bool{}
	for _, s := range shippedFaceSpecs {
		specKeys[s.Key] = true
	}
	for key := range embedded {
		if !specKeys[key] {
			t.Fatalf(
				"testShippedFontSet() carries face %q with no row in shippedFaceSpecs — it ships with "+
					"no weight, name or staticness assertion at all", key,
			)
		}
	}
}

// TestShippedFaceKeysDifferFromPostScriptNames is the ANTI-VACUITY
// witness for the /BaseFont conformance fix (D-2.2.6).
//
// The fix makes /BaseFont carry the embedded program's own PostScript
// name instead of the FontSet key. The `font-text` fixture cannot see
// it: that fixture keys its face "Roboto-Regular", which is exactly
// Roboto's name[6], so /BaseFont reads HXRYNT+Roboto-Regular either way
// and its golden did not move. If the shipped faces ever coincided the
// same way, EVERY /BaseFont assertion in this repo would pass whether
// the fix were present or reverted.
//
// They do not coincide — "Noto Sans SC" vs "NotoSansSC-Regular" — and
// this test fails the day that stops being true, rather than letting the
// multi-script fixture quietly stop discriminating.
func TestShippedFaceKeysDifferFromPostScriptNames(t *testing.T) {
	distinct := 0
	for _, spec := range shippedFaceSpecs {
		if spec.Key != spec.PostScriptName {
			distinct++
		}
	}
	if distinct == 0 {
		t.Fatal(
			"every shipped face's FontSet key equals its PostScript name, so the multi-script fixture's " +
				"/BaseFont assertions can no longer tell the ISO 32000-1 Table 117 fix from the pre-2.2 " +
				"behaviour of spelling /BaseFont from the FontSet key. Add a face whose key differs, or " +
				"move the discrimination somewhere that still has it.",
		)
	}
	if distinct != len(shippedFaceSpecs) {
		t.Logf(
			"note: %d of %d shipped faces have a FontSet key distinct from their PostScript name",
			distinct, len(shippedFaceSpecs),
		)
	}
}

// baseFontPattern is the shape ISO 32000-1 §9.6.4 + Table 117 require:
// exactly six upper-case letters, ONE plus sign, then the CIDFont
// program's own name. Asserting the shape (rather than only the value)
// is what catches a double-applied tag — `ABCDEF+ABCDEF+Noto...` — and a
// stale name, in one assertion.
var baseFontPattern = regexp.MustCompile(`^([A-Z]{6})\+(.+)$`)

// assertBaseFontNames checks every /BaseFont in a produced PDF against
// the spec's PostScript names.
func assertBaseFontNames(t *testing.T, label string, pdfBytes []byte, want map[string]bool) {
	t.Helper()
	re := regexp.MustCompile(`/BaseFont\s*/([^\s/\[\]<>()]+)`)
	matches := re.FindAllSubmatch(pdfBytes, -1)
	if len(matches) == 0 {
		t.Fatalf(
			"%s: PRESENCE PRECONDITION FAILED — the PDF contains no /BaseFont entry at all, so every "+
				"assertion below would pass vacuously", label,
		)
	}
	seen := map[string]bool{}
	for _, m := range matches {
		got := string(m[1])
		parts := baseFontPattern.FindStringSubmatch(got)
		if parts == nil {
			t.Errorf(
				"%s: /BaseFont %q does not match ^[A-Z]{6}\\+.+$ — ISO 32000-1 §9.6.4 requires exactly "+
					"a six-upper-case-letter subset tag, one '+', then the CIDFont program's name",
				label, got,
			)
			continue
		}
		psName := parts[2]
		if strings.Contains(psName, "+") {
			t.Errorf(
				"%s: /BaseFont %q carries MORE THAN ONE '+' — the subset tag looks doubly applied",
				label, got,
			)
			continue
		}
		if !want[psName] {
			t.Errorf(
				"%s: /BaseFont %q names PostScript font %q, which is not one of the expected shipped "+
					"PostScript names %v.\nISO 32000-1 Table 117: /BaseFont shall be the CIDFontName in the "+
					"CIDFont program — not the FontSet key the caller happened to file the face under.",
				label, got, psName, sortedKeys(want),
			)
			continue
		}
		seen[psName] = true
	}
	// Coverage witness: every expected name must actually have appeared.
	for name := range want {
		if !seen[name] {
			t.Errorf("%s: expected PostScript name %q never appeared in any /BaseFont", label, name)
		}
	}
}

// assertSubsetTagsAppearOnlyInNames is V5a's OTHER half — the one that
// was specified and not built.
//
// V5a asks for TWO assertions. The first, that the tag never occurs
// inside the program bytes it was derived from, exists in
// internal/fontset (TestSubsetTagNonCircularity) and is red-proved. The
// second, that the tag's appearances in the PDF are confined to name
// positions, did not: render_test.go asserts the tag IS at /BaseFont,
// which is positive membership, not exclusivity. A test that only ever
// asks "is it here" cannot see it turning up somewhere else as well.
//
// Exclusivity is what keeps the derivation non-circular in practice. The
// tag is a hash of the final program bytes, so if anything ever wrote it
// somewhere that feeds back into those bytes, the derivation would eat
// its own output. Confining every occurrence to "immediately after a '/'
// and immediately before a '+'" is the checkable form of that.
func assertSubsetTagsAppearOnlyInNames(t *testing.T, label string, pdfBytes []byte) {
	t.Helper()
	re := regexp.MustCompile(`/BaseFont\s*/([A-Z]{6})\+`)
	matches := re.FindAllSubmatch(pdfBytes, -1)
	if len(matches) == 0 {
		t.Fatalf("%s: PRESENCE PRECONDITION FAILED — no tagged /BaseFont found, so exclusivity is vacuous", label)
	}
	tags := map[string]bool{}
	for _, m := range matches {
		tags[string(m[1])] = true
	}

	for tag := range tags {
		occurrences := 0
		nameShaped := 0
		for i := 0; i+len(tag) <= len(pdfBytes); i++ {
			if string(pdfBytes[i:i+len(tag)]) != tag {
				continue
			}
			occurrences++
			// A name-position occurrence reads "/TAG+" — preceded by the
			// PDF name solidus, followed by the subset-tag separator.
			okBefore := i > 0 && pdfBytes[i-1] == '/'
			okAfter := i+len(tag) < len(pdfBytes) && pdfBytes[i+len(tag)] == '+'
			if okBefore && okAfter {
				nameShaped++
			}
		}
		if occurrences == 0 {
			t.Fatalf("%s: tag %q was matched at /BaseFont but then not found by scan — test bug", label, tag)
		}
		if nameShaped != occurrences {
			t.Errorf(
				"%s: subset tag %q occurs %d times in the PDF but only %d are in a name position "+
					"(\"/%s+\"). V5a requires the tag's appearances to stay confined to PDF-level names: an "+
					"occurrence anywhere else — above all inside an embedded font program — would make the "+
					"derivation CIRCULAR, since the tag is a hash of those very bytes. Revisit the "+
					"derivation; do not relax this test.",
				label, tag, occurrences, nameShaped, tag,
			)
		}
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// tiny insertion sort — avoids pulling "sort" in for three items
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// assertEmbeddedProgramsAreStaticRegular is the embedded-program half of
// the split. The embedded programs DO carry OS/2, so usWeightClass is
// readable and meaningful there; they do NOT carry `name`, so nothing
// name-related is asserted here (see this file's header).
func assertEmbeddedProgramsAreStaticRegular(t *testing.T, label string, programs [][]byte, wantWeight uint16) {
	t.Helper()
	if len(programs) == 0 {
		t.Fatalf("%s: PRESENCE PRECONDITION FAILED — no embedded font programs extracted", label)
	}
	for i, p := range programs {
		plabel := fmt.Sprintf("%s: embedded program #%d", label, i)
		tables := sfntTables(t, plabel, p)
		if _, ok := tables["glyf"]; !ok {
			t.Errorf("%s: no `glyf` table", plabel)
		}
		if _, ok := tables["CFF2"]; ok {
			t.Errorf("%s: carries a `CFF2` table", plabel)
		}
		for _, tag := range []string{"fvar", "gvar", "avar"} {
			if _, ok := tables[tag]; ok {
				t.Errorf("%s: still carries %q — not genuinely instanced", plabel, tag)
			}
		}
		// Documented on purpose: if a future change enables name-table
		// pass-through, this stops being true and the split above should
		// be revisited — rather than silently letting the .ttf-side
		// assertions look like they cover the embedded program.
		if _, ok := tables["name"]; ok {
			t.Logf(
				"%s: NOTE — this embedded program now HAS a `name` table, which it did not when the "+
					"assertion split in shipped_faces_test.go was written. Name assertions could now be "+
					"made here too.", plabel,
			)
		}
		if got := sfntWeightClass(t, plabel, tables); got != wantWeight {
			t.Errorf(
				"%s: OS/2.usWeightClass is %d, want %d.\n"+
					"textshape copies OS/2 VERBATIM from the source face (subset/execute.go:496-499) and "+
					"never rewrites usWeightClass, so this value is the SHIPPED FACE's weight showing "+
					"through. A wrong value here means a wrong face was shipped, not a subsetting bug.",
				plabel, got, wantWeight,
			)
		}
	}
}
