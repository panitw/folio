package fontset

import (
	"fmt"

	"github.com/boxesandglue/textshape/ot"
)

// THE VARIABLE-FACE REFUSAL, AND THERE IS EXACTLY ONE OF IT.
//
// Story 16.0 (D-16.6). Until this file existed, the `fvar` test and its
// message were an inline `if` in the body of New — the renderer's guard, and
// the only one. `embedFontFamily`'s structural gate is checkSfnt, whose own
// doc fences it at "the file is not lying about what it is, and nothing
// more", and which never inspects a table tag at all. So a pick of a variable
// face wrote a `.folio` that SAVED CLEANLY AND FAILED AT RENDER — the one
// outcome D-8.4d.1 and D-16.1 both promise cannot happen. The refusal was
// right; it arrived far too late.
//
// The command now refuses it too, and the engineering lead's guardrail at
// this story's plan gate is why the two refusals are ONE FUNCTION rather than
// two implementations that agree today: a command-side copy of this test
// would be a second authority by construction, and the two would drift the
// moment either message or predicate was touched. component_commands_test.go
// feeds ONE byte slice to both doors and asserts both refuse; deleting either
// call site reds that test.
//
// THE RENDERER'S GUARD IS KEPT, NOT MOVED. A `.folio` can be hand-written and
// the loader is not the only door, so `New` still calls this before it builds
// a shaper. fontset_test.go's TestNewRejectsVariableFace passes unchanged.
//
// variableFaceError owns both the test and the sentence. It is unexported and
// takes the already-parsed face, because New has one in hand and reparsing
// there would be waste.
func variableFaceError(name string, parsed *ot.Font) error {
	if !parsed.HasTable(ot.TagFvar) {
		return nil
	}
	return fmt.Errorf(
		"fontset: font %q: this is a VARIABLE font (it has an `fvar` table) and cannot be embedded: "+
			"PDF 1.7 cannot express a variable font, and folio will not silently pick an instance for you "+
			"(the default instance is not always Regular — Noto Sans SC's `wght` axis defaults to 100, i.e. Thin). "+
			"Instance it to a single static weight first, e.g. "+
			"`fonttools varLib.instancer --update-name-table %s.ttf wght=400 -o %s-Regular.ttf`, "+
			"pinning EVERY axis the face declares rather than just `wght` "+
			"(see tools/fontgen/instance_faces.py for the full recipe folio uses for its own shipped faces, "+
			"including the reproducibility pin the bare command above omits)",
		name, name, name,
	)
}

// RefuseVariableFace is the byte-taking door on to that one predicate, for a
// caller holding raw bytes rather than a parsed face — which is every caller
// outside this package, `embedFontFamily` included.
//
// It re-runs ot.ParseFont. That is a COST, not a correctness question: the
// command has already decoded and structurally checked these bytes, and
// paying one more parse at the moment of an author's click buys the
// single-authority property the alternative (exporting *ot.Font across the
// seam) would forfeit — D-1.5.10/AC17a forbids that type leaving this package
// at all.
//
// A face these bytes cannot be parsed as is NOT refused here, and the
// omission is deliberate. This function answers exactly one question — "is
// this a variable face?" — and an unparsable face is not one. New reports the
// parse failure in its own sentence; widening this helper to speak for that
// too would make it a second, partial copy of New's admission rules, which is
// the very shape this file exists to prevent.
func RefuseVariableFace(name string, data []byte) error {
	parsed, err := ot.ParseFont(data, 0)
	if err != nil {
		return nil
	}
	return variableFaceError(name, parsed)
}
