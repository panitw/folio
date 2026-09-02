package fontset

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode/utf16"
)

// ---------------------------------------------------------------------------
// STORY 16.1b: THE BINARY IS ASKED WHAT IT SAYS ABOUT ITSELF.
//
// THE TWO CONTROLS ARE COMMITTED BYTES AND SYNTHESISED BYTES, and neither is
// a new font file in the tree (D-16.R.9).
//
//   - CONTRADICTION, over real bytes: testdata/fonts/Roboto-Regular.ttf is
//     recorded `Apache-2.0` (lint/internal/licence/licencecensus_test.go) and
//     its own record 13 reads "Licensed under the Apache License, Version
//     2.0". Declaring it `OFL-1.1` is therefore a TRUE contradiction — the
//     bytes match a different ADMITTED licence's signature than the one
//     claimed — with no mislabelled binary committed anywhere.
//
//   - REFUSE-SIGNATURE, over bytes built in this process: no committed face
//     carries a copyleft statement and none should, so that half of the guard
//     gets a name table SYNTHESISED here. Committing a font whose sole
//     purpose is to lie about its own licence would trip the manifest guard,
//     the binary-identity walk and the licence census — three gates that
//     exist to keep exactly that file out — in a product whose whole subject
//     is licence honesty.
//
// The synthesis rebuilds the whole `name` table and REPOINTS the sfnt table
// directory record at it, rather than patchUnitsPerEm's same-length overwrite
// in place. That is a deliberate widening of that helper's method and not a
// departure from it: `name` records are variable-length behind a string
// storage pool, so a same-length substitution can only ever say something the
// same size as what is already there, and the arms below need records that
// are absent, present-but-unrecognised, and present in two different
// combinations of 13 and 0. The directory walk is patchUnitsPerEm's, verbatim
// in shape.

// nameRecord is one (nameID, string) pair for buildNameTable.
type nameRecord struct {
	id    uint16
	value string
}

// buildNameTable assembles a format-0 `name` table carrying exactly the
// given records, each as platform 3 (Windows) / encoding 1 (UCS-2) /
// language 0x0409 — the combination ot.ParseName decodes as UTF-16BE, and
// the one every real face in this repository uses for these records.
func buildNameTable(records []nameRecord) []byte {
	var directory, storage []byte
	for _, record := range records {
		encoded := encodeUTF16BEForTest(record.value)
		entry := make([]byte, 12)
		binary.BigEndian.PutUint16(entry[0:], 3)      // platformID: Windows
		binary.BigEndian.PutUint16(entry[2:], 1)      // encodingID: Unicode BMP
		binary.BigEndian.PutUint16(entry[4:], 0x0409) // languageID: en-US
		binary.BigEndian.PutUint16(entry[6:], record.id)
		binary.BigEndian.PutUint16(entry[8:], uint16(len(encoded)))
		binary.BigEndian.PutUint16(entry[10:], uint16(len(storage)))
		directory = append(directory, entry...)
		storage = append(storage, encoded...)
	}
	header := make([]byte, 6)
	binary.BigEndian.PutUint16(header[0:], 0) // format 0
	binary.BigEndian.PutUint16(header[2:], uint16(len(records)))
	binary.BigEndian.PutUint16(header[4:], uint16(6+12*len(records))) // storageOffset
	out := append(header, directory...)
	return append(out, storage...)
}

func encodeUTF16BEForTest(s string) []byte {
	units := utf16.Encode([]rune(s))
	out := make([]byte, 0, len(units)*2)
	for _, unit := range units {
		out = append(out, byte(unit>>8), byte(unit))
	}
	return out
}

// replaceNameTable returns a COPY of a well-formed sfnt whose `name` table
// record points at table instead. The new table is appended at the end of
// the file on a 4-byte boundary; the old bytes are left where they are and
// simply stop being referenced, which is what keeps every other table's
// offset valid.
//
// Checksums are not recomputed, deliberately: ot.ParseFont validates none of
// them (ot/font.go), and a test helper that quietly repaired the file would
// be asserting over bytes no author could produce by hand.
func replaceNameTable(t *testing.T, src, table []byte) []byte {
	t.Helper()
	out := make([]byte, len(src))
	copy(out, src)
	for len(out)%4 != 0 {
		out = append(out, 0)
	}
	offset := uint32(len(out))
	out = append(out, table...)

	numTables := int(binary.BigEndian.Uint16(out[4:6]))
	for i := 0; i < numTables; i++ {
		record := out[12+i*16 : 12+i*16+16]
		if string(record[0:4]) != "name" {
			continue
		}
		binary.BigEndian.PutUint32(record[8:12], offset)
		binary.BigEndian.PutUint32(record[12:16], uint32(len(table)))
		return out
	}
	t.Fatal("the subject face carries no `name` table record to repoint, so nothing this helper builds could be read back")
	return nil
}

// removeNameTable drops the `name` table's directory record entirely, so the
// face HAS no name table rather than an empty one. Same directory walk.
func removeNameTable(t *testing.T, src []byte) []byte {
	t.Helper()
	out := make([]byte, len(src))
	copy(out, src)
	numTables := int(binary.BigEndian.Uint16(out[4:6]))
	for i := 0; i < numTables; i++ {
		offset := 12 + i*16
		if string(out[offset:offset+4]) != "name" {
			continue
		}
		copy(out[offset:12+(numTables-1)*16], out[offset+16:12+numTables*16])
		for j := 12 + (numTables-1)*16; j < 12+numTables*16; j++ {
			out[j] = 0
		}
		binary.BigEndian.PutUint16(out[4:6], uint16(numTables-1))
		return out
	}
	t.Fatal("the subject face carries no `name` table record to remove, so no assertion about its absence can be made")
	return nil
}

// faceWithNames is the synthesis in one call: the committed Roboto with its
// name table replaced by exactly the records given.
func faceWithNames(t *testing.T, records ...nameRecord) []byte {
	t.Helper()
	return replaceNameTable(t, testFontBytes(t), buildNameTable(records))
}

const (
	silSentence    = "This Font Software is licensed under the SIL Open Font License, Version 1.1."
	apacheSentence = "Licensed under the Apache License, Version 2.0"
	ubuntuSentence = "Licensed under the Ubuntu Font Licence 1.0."
)

// TestLicenceGuardPreconditions asserts the two committed controls really do
// say what this file's other tests assume, BEFORE anything asserts an outcome
// over them. A fixture that stopped carrying its licence sentence would
// otherwise turn every CONTRADICTION assertion below into a NO EVIDENCE one
// and take the whole file green for the wrong reason.
func TestLicenceGuardPreconditions(t *testing.T) {
	roboto, present := ReadLicenceStatement(testFontBytes(t))
	if !present {
		t.Fatal("precondition: testdata/fonts/Roboto-Regular.ttf makes no readable licence statement, so it cannot serve as the contradiction control")
	}
	if !strings.Contains(roboto, "Apache License") {
		t.Fatalf("precondition: Roboto-Regular.ttf's licence statement is %q and no longer names the Apache License — the contradiction control has stopped being a contradiction", roboto)
	}

	noto, present := ReadLicenceStatement(variableNotoBytes(t))
	if !present {
		t.Fatal("precondition: NotoSansThai-VF.ttf makes no readable licence statement, so it cannot serve as the confirmation control")
	}
	if !strings.Contains(noto, "SIL Open Font License") {
		t.Fatalf("precondition: NotoSansThai-VF.ttf's licence statement is %q and no longer names the SIL OFL", noto)
	}
}

// variableNotoBytes is the OFL confirmation control. It is VARIABLE, which is
// why every assertion over it in this file goes through the licence door
// DIRECTLY: routed through embedFontFamily the `fvar` refusal fires first
// (component_commands.go:2414) and would mask the licence outcome entirely —
// a green assertion about an arm that was never reached.
func variableNotoBytes(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fonts", "notosansthai-variable-testonly", "NotoSansThai-VF.ttf")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read the confirmation control %s: %v", path, err)
	}
	return data
}

// TestRefuseContradictedLicenceRefusesTheDeclaredIdTheBytesContradict is the
// POSITIVE CONTROL the ruling requires: without it the guard could ship
// having never been observed to refuse anything at all.
//
// RED-PROVED BY DELETING THE GUARD, not by falsifying a condition (recorded
// 2026-09-03, wd folio-go, `go test ./internal/fontset/`): with the body of
// RefuseContradictedLicence replaced by `return nil` this test reds, and so
// does the copyleft control below. Inverting a predicate would only have
// proved arm order.
func TestRefuseContradictedLicenceRefusesTheDeclaredIdTheBytesContradict(t *testing.T) {
	err := RefuseContradictedLicence("Roboto", "OFL-1.1", testFontBytes(t))
	if err == nil {
		t.Fatal("a face whose own name table names the Apache License was admitted under a declared OFL-1.1 — the tie the build gate holds the 21 committed faces to does not exist at runtime, and Epic 16 has replaced that gate with something strictly weaker (D-8.6.5)")
	}
	// THE MESSAGE MUST NAME BOTH SIDES. Either alone is unactionable: the
	// author cannot tell whether the catalogue row is wrong or the binary is
	// the wrong binary without seeing the two statements together.
	for _, want := range []string{"Roboto", "OFL-1.1", "Apache License"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
}

// TestRefuseContradictedLicenceAdmitsWhatTheBytesConfirm is the
// over-broadness control. A guard that refused everything would satisfy the
// test above and be worthless, so the same two faces are declared the licence
// their own bytes actually name and both must be admitted.
func TestRefuseContradictedLicenceAdmitsWhatTheBytesConfirm(t *testing.T) {
	if err := RefuseContradictedLicence("Roboto", "Apache-2.0", testFontBytes(t)); err != nil {
		t.Errorf("a face declared Apache-2.0 whose own name table names the Apache License was refused: %v", err)
	}
	if err := RefuseContradictedLicence("Noto Sans Thai", "OFL-1.1", variableNotoBytes(t)); err != nil {
		t.Errorf("a face declared OFL-1.1 whose own name table names the SIL OFL was refused: %v", err)
	}
}

// TestRefuseContradictedLicenceRefusesCopyleftWhateverIsDeclared is the
// SECOND positive control, and it covers the half the committed fixtures
// cannot reach (D-16.R.9). This is the part D-16.R.7 added as NEW — the
// build-time tie never had it — and a new guard with no control is exactly
// the vacuity the ruling exists to prevent.
//
// The declared id is `OFL-1.1` in every row, which is the whole point: the
// declaration is the thing under suspicion, so the refuse half must not
// consult it. AD-26 binds ALL, and its Prevents is copyleft arriving through
// a plausible-looking package.
func TestRefuseContradictedLicenceRefusesCopyleftWhateverIsDeclared(t *testing.T) {
	for _, statement := range []string{
		"Licensed under the GNU General Public License version 3",
		"Licensed under the GPL",
		"This font is released under the GPLv2 with the font exception",
		"Licensed under the LGPL-2.1",
		"Licensed under the GNU Affero General Public License",
		"Licensed under AGPL-3.0",
		"Licensed under the Server Side Public License",
		"Licensed under SSPL",
		"Creative Commons Attribution-ShareAlike 4.0 International",
		"Licensed under CC BY-SA 4.0",
	} {
		face := faceWithNames(t, nameRecord{id: 13, value: statement})
		err := RefuseContradictedLicence("Copyleft Face", "OFL-1.1", face)
		if err == nil {
			t.Errorf("a face whose own name table reads %q was admitted under a declared OFL-1.1 — the copyleft floor does not exist, and a share-alike font program would be embedded and subset into every PDF the author produces", statement)
			continue
		}
		if !strings.Contains(err.Error(), statement) {
			t.Errorf("the refusal of %q does not quote what the bytes said: %v", statement, err)
		}
	}

	// AND THE OVER-BROADNESS CONTROL FOR THIS HALF TOO. The refuse patterns
	// must not fire on the permissive sentences real faces carry, or every
	// admitted face becomes a refusal and the guard is turned off within a
	// week.
	for _, statement := range []string{silSentence, apacheSentence, ubuntuSentence} {
		face := faceWithNames(t, nameRecord{id: 13, value: statement})
		if err := RefuseContradictedLicence("Permissive Face", "OFL-1.1", face); err != nil && strings.Contains(err.Error(), "copyleft") {
			t.Errorf("a refuse-signature fired on the permissive statement %q: %v", statement, err)
		}
	}
}

// TestRefuseContradictedLicenceAdmitsSilence is the arm that will look like a
// hole to the next reader, which is why the reasoning is in the code beside
// it and asserted here rather than merely described.
//
// The threat is a face travelling under ANOTHER PROJECT'S terms — a FALSE
// statement. A face that says nothing has made no statement to be false, so
// refusing it catches none of that; and it was MEASURED to cost about a sixth
// of the library (50 of 100 sampled upstream faces refused under the original
// "absent or unparseable => REFUSE" contract). Genuinely unreadable bytes are
// refused one step later by template.DecodeFontForRender / checkSfnt.
func TestRefuseContradictedLicenceAdmitsSilence(t *testing.T) {
	// No name table at all.
	nameless := removeNameTable(t, testFontBytes(t))
	if _, present := ReadLicenceStatement(nameless); present {
		t.Fatal("precondition: the name-stripped face still reports a licence statement, so the strip did not take effect and the arm below is untested")
	}
	if err := RefuseContradictedLicence("Nameless", "OFL-1.1", nameless); err != nil {
		t.Errorf("a face with no name table at all was refused: %v", err)
	}

	// Bytes that are not a font at all: unparseable, and admitted here
	// because this door answers one question only. checkSfnt refuses them.
	if err := RefuseContradictedLicence("Rubbish", "OFL-1.1", []byte("not a font")); err != nil {
		t.Errorf("unreadable bytes were refused by the LICENCE door, which is not its question: %v", err)
	}

	// A statement that is present and recognised by nothing in either table.
	unknown := faceWithNames(t, nameRecord{id: 13, value: "All rights reserved by a licence this table has never heard of."})
	if err := RefuseContradictedLicence("Unknown Terms", "OFL-1.1", unknown); err != nil {
		t.Errorf("a face whose statement matches no signature was refused: %v", err)
	}

	// A declared id with NO admit-signature row (D-16.R.10). MIT is the
	// reachable case: D-8.5.3 admits it and google/fonts publishes no MIT
	// token for a signature to be minted against — "absence, not narrowing".
	// Refusing here would refuse a licence the OWNER has admitted.
	if slices.ContainsFunc(admitLicenceSignatures, func(s licenceSignature) bool { return s.id == "MIT" }) {
		t.Fatal("this test's subject has gone: admitLicenceSignatures now has an MIT row, so `a declared id with no signature entry` is no longer reachable through MIT — re-derive the case rather than deleting the check")
	}
	if err := RefuseContradictedLicence("MIT Face", "MIT", faceWithNames(t, nameRecord{id: 13, value: "Licensed under the MIT License"})); err != nil {
		t.Errorf("a face declared MIT — an id D-8.5.3 admits and this table deliberately does not cover — was refused: %v", err)
	}
}

// TestNameID0IsConsultedOnlyWhenNameID13IsAbsent holds BOTH halves of
// D-16.R.7's widening, and the second half is the one that keeps it a
// widening rather than a weakening.
//
// Measured at Story 16.1's plan gate: all three static `ufl/` families carry
// NO record 13 and state their terms in record 0, so a tie that looked only
// at 13 would have refused every genuine Ubuntu-licensed family. But if 0
// were consulted whenever 13 failed to MATCH, a face whose 13 says GPL could
// be laundered by a permissive-sounding 0 — defeating the contradiction check
// with the very thing it exists to catch.
func TestNameID0IsConsultedOnlyWhenNameID13IsAbsent(t *testing.T) {
	// ABSENT 13, terms in 0: the `ufl/` shape. Consulted, and it confirms.
	ubuntu := faceWithNames(t, nameRecord{id: 0, value: "Copyright 2011 Canonical Ltd. " + ubuntuSentence})
	statement, present := ReadLicenceStatement(ubuntu)
	if !present || !strings.Contains(statement, "Ubuntu Font Licence") {
		t.Fatalf("record 0 was not consulted for a face carrying no record 13: present=%v statement=%q", present, statement)
	}
	if err := RefuseContradictedLicence("Ubuntu", "Ubuntu-font-1.0", ubuntu); err != nil {
		t.Errorf("a face declaring Ubuntu-font-1.0 and stating those terms in record 0 was refused: %v", err)
	}
	// And the contradiction arm works through record 0 too, or the widening
	// would have opened a door that only ever admits.
	if err := RefuseContradictedLicence("Ubuntu", "OFL-1.1", ubuntu); err == nil {
		t.Error("a face stating the Ubuntu Font Licence in record 0 was admitted under a declared OFL-1.1 — record 0 is being read as a licence-shaped decoration rather than as a statement")
	}

	// PRESENT AND MISMATCHED 13, PERMISSIVE 0: the laundering attempt. 13
	// alone decides, so this must refuse.
	laundered := faceWithNames(t,
		nameRecord{id: 0, value: "Copyright 2020. " + silSentence},
		nameRecord{id: 13, value: "Licensed under the GNU General Public License version 3"},
	)
	err := RefuseContradictedLicence("Laundered", "OFL-1.1", laundered)
	if err == nil {
		t.Fatal("a face whose record 13 names the GPL was admitted because its record 0 sounded permissive — record 0 is being used as a fallback rather than as a widening, which defeats the contradiction check with the very thing it exists to catch")
	}
	if !strings.Contains(err.Error(), "General Public License") {
		t.Errorf("the refusal did not read record 13: %v", err)
	}
	if strings.Contains(err.Error(), "SIL Open Font License") {
		t.Errorf("the refusal quoted record 0, which must not have been consulted at all: %v", err)
	}
}

// TestLicenceSignatureTablesAreOrderedAndNotEmpty is AD-1's guard in this
// file: an ordered slice, never a map, so a face matching two rows produces
// the same sentence on every run. It also stops both tables being emptied
// into vacuous green.
func TestLicenceSignatureTablesAreOrderedAndNotEmpty(t *testing.T) {
	if len(admitLicenceSignatures) == 0 {
		t.Fatal("vacuity guard: the admit table is empty, so every CONFIRMATION and CONTRADICTION assertion in this file would be answered by NO EVIDENCE")
	}
	if len(refuseLicenceSignatures) == 0 {
		t.Fatal("vacuity guard: the refuse table is empty, so the copyleft floor does not exist")
	}
	for _, signature := range admitLicenceSignatures {
		if signature.id == "" || signature.label == "" || signature.pattern == nil {
			t.Errorf("an admit row is incomplete: %+v", signature)
		}
	}
	for _, signature := range refuseLicenceSignatures {
		if signature.id != "" {
			t.Errorf("a refuse row carries an SPDX id (%q); the refuse half is checked against every face whatever it declares and must be keyed by nothing", signature.id)
		}
		if signature.label == "" || signature.pattern == nil {
			t.Errorf("a refuse row is incomplete: %+v", signature)
		}
	}
}

// ---------------------------------------------------------------------------
// THE MIRROR CONTRACT (D-7.4.5 / DW-25, applied by D-16.R.7).

// designerLicenceSignatureTable extracts the object literal the designer's
// build-time tie is keyed off. Anchored on the TS declaration's OWN
// identifier, because an object literal is that file's commonest shape and an
// unanchored match would read some other table and compare it to this one — a
// green test asserting the wrong thing.
var designerLicenceSignatureTable = regexp.MustCompile(`(?s)const licenceSignatures[^=]*=\s*\{(.*?)\n\}`)

// designerLicenceSignatureKey pulls the SPDX ids out of that literal.
var designerLicenceSignatureKey = regexp.MustCompile(`(?m)^\s*'([^']+)'\s*:`)

// TestGoLicenceTableSubsumesTheDesignerTable is the mirror contract, enforced
// by a test and not by a comment.
//
// TWO TABLES IN TWO LANGUAGES, and the Go one is the authority for documents:
// the designer's runs at BUILD time over the 21 committed faces, while this
// one runs at the moment bytes are written into a `.folio`, over faces nobody
// has reviewed. If the designer's table admitted an id Go's did not, a face
// the build gate had passed would be refused at the pick — or, worse, the two
// halves would disagree about what a licence sentence looks like and only one
// of them would be watched.
//
// It reads the TypeScript as SOURCE TEXT, which is the established idiom here
// (canvas_projection_wire_test.go does the same to engine-protocol.ts). The
// designer file exports nothing, and no shared generated artifact is
// introduced for this: the only Go/TS bridge in the repository,
// folio-designer/src/generated/font-catalogue.ts, is gitignored and TS-only.
//
// SUBSUMPTION IS STRICT TODAY, so this is not vacuous on landing: TS declares
// {OFL-1.1, Ubuntu-font-1.0} and Go declares those plus Apache-2.0. Red-prove
// it by adding a row to the TS table with no Go counterpart.
func TestGoLicenceTableSubsumesTheDesignerTable(t *testing.T) {
	path := filepath.Join("..", "..", "..", "folio-designer", "src", "font-catalogue.test.ts")
	source, err := os.ReadFile(path)
	if err != nil {
		// NOT A SKIP. The guard on the other side of this seam is half of
		// "both, or halt"; a missing one is a finding, not an excuse.
		t.Fatalf("read the designer's build-time licence tie %s: %v", path, err)
	}
	match := designerLicenceSignatureTable.FindSubmatch(source)
	if match == nil {
		t.Fatal("font-catalogue.test.ts no longer declares `const licenceSignatures` where this test can read it; if the table was restructured, re-derive this extraction rather than deleting the check — D-16.R.7 requires BOTH guards or a halt")
	}
	var designerIDs []string
	for _, key := range designerLicenceSignatureKey.FindAllSubmatch(match[1], -1) {
		designerIDs = append(designerIDs, string(key[1]))
	}
	// VACUITY GUARD. "Every TS id is covered by Go" is also true of a table
	// this test read as empty, which is what a restructured literal or a
	// changed quote style would silently produce.
	if len(designerIDs) == 0 {
		t.Fatal("vacuity guard: the extraction read 0 ids out of the designer's licenceSignatures table, so `Go subsumes TS` says nothing")
	}

	goIDs := make([]string, 0, len(admitLicenceSignatures))
	for _, signature := range admitLicenceSignatures {
		goIDs = append(goIDs, signature.id)
	}
	for _, id := range designerIDs {
		if !slices.Contains(goIDs, id) {
			t.Errorf("the designer's build-time tie admits %q and the Go table — the authority for what goes into a document — has no row for it. The two are a mirror contract: they move in ONE commit. Go declares %v.", id, goIDs)
		}
	}
}

// TestTheBuildTimeTieIsStillInPlace is the other half of "both, or halt". The
// designer's tie is not deleted or weakened because Go now checks it: Go
// guards the DOCUMENT, and the designer's guards the 21 faces this repository
// itself redistributes, which Go never sees a pick of.
func TestTheBuildTimeTieIsStillInPlace(t *testing.T) {
	path := filepath.Join("..", "..", "..", "folio-designer", "src", "font-catalogue.test.ts")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read the designer's build-time licence tie %s: %v", path, err)
	}
	for _, needle := range []string{
		"const licenceSignatures",
		"licenceSignatures[face.licence]",
		".toMatch(signature as RegExp)",
	} {
		if !strings.Contains(string(source), needle) {
			t.Errorf("font-catalogue.test.ts no longer contains %q — the build-time nameID 13 tie has been deleted or restructured. Story 16.1b ADDS a runtime tie; it does not replace this one (D-16.R.7: both, or halt).", needle)
		}
	}
}
