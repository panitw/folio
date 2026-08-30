package template

import (
	"fmt"
	"strconv"
	"strings"
)

// This file is version handling (D-1.4.9, D-1.4.13, AC6, AC7): a higher
// MAJOR than this library supports is a load error, never a
// best-effort render (FR13); a higher MINOR loads, and unknown content
// it carries passes through opaquely (rawvalue.go, parse.go).
//
// Blocker 4 / F-2 (this story's Dev Notes, measured at f9c27b3): this
// file's name and this function's error shape are exactly what
// folio-go/testdata/lint/numeric-formatting/template/version.go stands
// in for — fmt.Errorf is fine here; internal/pdf/numbers.go's
// strconv-confinement rule is scoped to internal/pdf only.

// SupportedMajor is the highest MAJOR version this library can load
// (D-1.4.9). SupportedVersion is the library's own CEILING: the highest
// full "MAJOR.MINOR" it knows how to read and write.
//
// IT IS NOT WHAT THIS LIBRARY AUTHORS FOR A BRAND-NEW DOCUMENT, and the
// doc comment that used to say so was wrong in the way D-1.4.13
// forbids. `version` is a property of the DOCUMENT, raised only by the
// content the document actually carries: a document using neither
// `style.lineSpacing` nor `style.color` declares 1.0 no matter what this
// constant reads, because declaring the library's ceiling would orphan
// a document from every older reader that could in fact have read it.
// versionForSave below is the one place that rule is expressed.
//
// Raised to 1.1 by Story 7.2 (D-7.2.1). 1.1 is the first MINOR the
// format has ever had; it adds two optional keys and changes the meaning
// of none — `style.lineSpacing` (Story 7.2) and, retrofitted, Epic 10's
// `style.color`, which shipped without moving the version and left
// colour-bearing documents declaring 1.0 while requiring 1.1.
//
// Raised to 2.0 by Story 7.3 (D-7.3.1, D-R7.9). `style.align` gained a
// fourth member, `justify` (FR47), and EXTENDING A CLOSED SET IS A MAJOR
// CHANGE under D-1.4.12: an older reader refuses an alignment word it
// does not recognise, on purpose, so that a file can never be drawn
// wrongly by a reader that misunderstands it. A justified document is
// therefore UNREADABLE to a 1.x reader rather than quietly ragged, and
// the MAJOR says so honestly. `style.justified`, or any other additive-
// key spelling that would have avoided the bump, was explicitly rejected
// (D-R7.9): it reintroduces the silently-wrong render D-1.4.12 exists to
// prevent.
//
// NONE OF THAT CHANGES WHAT AN EXISTING DOCUMENT DECLARES. A document
// using only `lineSpacing` or `color` still declares 1.1; one using
// neither still declares 1.0; only one that actually carries
// `align: "justify"` declares 2.0. All three coexist, and a brand-new
// document declares the LOWEST version its content requires.
const (
	SupportedMajor   = 2
	SupportedVersion = "2.0"
)

// baseVersion is the lowest version any document can declare, and the
// version a document whose content requires nothing newer keeps.
//
// minorFeatureVersion is the version introduced by the two 1.1 keys.
//
// majorFeatureVersion is the version introduced by the 2.0 closed-set
// extension, `style.align: "justify"` (Story 7.3).
//
// They are named rather than spelled inline so versionRequiredByContent
// reads as the rule rather than as string handling.
const (
	baseVersion         = "1.0"
	minorFeatureVersion = "1.1"
	majorFeatureVersion = "2.0"
)

// parseVersion splits a "MAJOR.MINOR" string into its two integer
// components. Both parts must be non-negative decimal integers with no
// sign and no extra components — the format is exactly two dot-separated
// integers (folio-format.md: `"version"` | `"MAJOR.MINOR"`).
func parseVersion(v string) (major, minor int, err error) {
	parts := strings.Split(v, ".")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("template: version %q is not of the form MAJOR.MINOR", v)
	}
	major, err = strconv.Atoi(parts[0])
	if err != nil || major < 0 || parts[0] == "" {
		return 0, 0, fmt.Errorf("template: version %q has an invalid MAJOR component", v)
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil || minor < 0 || parts[1] == "" {
		return 0, 0, fmt.Errorf("template: version %q has an invalid MINOR component", v)
	}
	return major, minor, nil
}

// checkVersionLoadable is AC6's gate: a higher MAJOR than
// SupportedMajor is a load error naming both the declared and the
// supported version, and no render is attempted (FR13). A higher MINOR
// is explicitly NOT a failure case (AC7, D-1.4.9) — it loads, and its
// unknown content passes through opaquely.
func checkVersionLoadable(declared string) error {
	major, _, err := parseVersion(declared)
	if err != nil {
		return err
	}
	if major > SupportedMajor {
		return fmt.Errorf("template declares version %s; supported version is %s (a higher MAJOR version is never loaded, FR13)", declared, SupportedVersion)
	}
	return nil
}

// versionForSave is the single, well-named function isolating
// D-1.4.13's raise/lower rule: `version` is a property of the DOCUMENT,
// carried verbatim, raised only when content actually requiring a higher
// version is present, and lowered never.
//
// `version` describes the document, not the writer. A PDF 1.7 file
// edited by a tool that understands only 1.4 features does not become a
// PDF 1.4 file — and, symmetrically, a plain 1.0 document does not
// become 1.1 merely because the writer knows what 1.1 is.
//
// Story 1.4 left this a stub returning `loaded` unchanged, with its own
// comment recording why: "the day a future MINOR exists, the raise path
// has exactly one place to be filled in." Story 7.2 is that day
// (D-7.2.1). The rule, in full:
//
//   - derive the version the document's CONTENT requires
//     (versionRequiredByContent);
//   - return the HIGHER of that and what was loaded;
//   - never lower, so a file already declaring a higher MINOR — or
//     carrying opaque content from one — round-trips verbatim.
//
// A loaded version this package cannot parse is returned untouched: it
// reached here only through checkVersionLoadable, and inventing a
// version for it would be a silent rewrite of a field the caller owns.
//
// "Requires nothing newer" is checked BEFORE the comparison, not through
// it. versionRequiredByContent returns baseVersion as its FLOOR — the
// answer "this document needs no 1.1 key" — not as a demand that the
// document declare 1.0. Comparing that floor numerically would raise any
// document declaring a LOWER version than the floor, and MAJOR 0 is
// loadable: parseVersion admits `major >= 0` and checkVersionLoadable
// refuses only `major > SupportedMajor`, so a `"0.9"` document loads and
// would be silently restamped `"1.0"` on save despite introducing no
// content that requires it. That is the AD-9 edit-and-edit-back break
// this branch exists to prevent, and the same rule the format spec
// states: saving raises the version only when content requiring it is
// introduced.
func versionForSave(loaded string, d *Document) string {
	required := versionRequiredByContent(d)
	if required == baseVersion {
		return loaded
	}
	loadedMajor, loadedMinor, err := parseVersion(loaded)
	if err != nil {
		return loaded
	}
	// parseVersion on a package constant cannot fail; the error is
	// ignored deliberately rather than propagated through a signature
	// no caller could act on.
	requiredMajor, requiredMinor, _ := parseVersion(required)
	if requiredMajor > loadedMajor || (requiredMajor == loadedMajor && requiredMinor > loadedMinor) {
		return required
	}
	return loaded
}

// versionRequiredByContent is D-1.4.13's "raised only by content" half,
// stated once: the LOWEST version that can express what this document
// actually contains.
//
// It is the MAXIMUM over every attachment point, never the first answer
// found (Story 7.3). Until 7.3 there was only one non-base answer, so a
// first-hit return was indistinguishable from a maximum; with two,
// a document whose first styled element sets `lineSpacing` and whose
// LATER one sets `align: "justify"` would have reported 1.1 and shipped
// a file its own 1.x reader must refuse to draw. The walk therefore
// visits every element and keeps the highest rank it saw.
//
// The questions, lowest requirement first:
//
//   - does any style block set `lineSpacing` or `color`? Both are 1.1's
//     optional keys; a document using neither is expressible in 1.0 and
//     must keep declaring it.
//   - does any style block set `align: "justify"`? That is 2.0's closed-
//     set extension, and no 1.x reader may draw it.
//
// Presence.Set, not "has a non-empty value" — but that only bites for
// `color`, and the asymmetry is stated rather than papered over.
// `color: null` is a legal, meaningful value ("no colour"), and it is
// still the key appearing in a file a 1.0 reader would not recognise, so
// it requires 1.1. `lineSpacing: null` never reaches here at all: it is
// refused at load (DecodeLineSpacingRaw), because an absent leading and
// a null one are the same thing and the format admits only the absent
// spelling.
func versionRequiredByContent(d *Document) string {
	if d == nil {
		return baseVersion
	}
	highest := rankBase
	for _, band := range []Band{d.Bands.PageHeader, d.Bands.Content, d.Bands.PageFooter} {
		for _, el := range band.Elements {
			if el.Style.Set && !el.Style.Null {
				if r := styleVersionRank(el.Style.Value); r > highest {
					highest = r
				}
			}
			if el.Table.Set && !el.Table.Null {
				hs := el.Table.Value.HeaderStyle
				if hs.Set && !hs.Null {
					if r := styleVersionRank(hs.Value); r > highest {
						highest = r
					}
				}
			}
		}
	}
	return versionForRank[highest]
}

// versionRank orders the versions the content rule can require, lowest
// first, so "the highest requirement in the document" is a comparison
// rather than a second string parse per element.
//
// It ranks REQUIREMENTS, not versions in general: a document's declared
// version can be anything (0.9, 1.9, 2.1) and versionForSave compares
// those numerically. This type only answers "which of the versions THIS
// LIBRARY's own content rules can demand is the highest one demanded
// here".
type versionRank int

const (
	rankBase versionRank = iota
	rankMinorFeature
	rankMajorFeature
)

// versionForRank maps a rank back to the version string it names.
// Indexed by rank rather than ranged over, so it stays deterministic and
// stays clear of D-1.3.5's map-range build failure.
var versionForRank = [...]string{
	rankBase:         baseVersion,
	rankMinorFeature: minorFeatureVersion,
	rankMajorFeature: majorFeatureVersion,
}

// styleVersionRank is the lowest version that can express ONE style
// block — the successor to styleNeedsMinorVersion, which was a bool
// because there was only ever one answer above the floor.
//
// Presence.Set, not "has a non-empty value", for the 1.1 keys — see
// versionRequiredByContent's own note. `align` is different and
// deliberately so: an align that is Set carries a VALUE from a closed
// set, and only one member of that set is new in 2.0, so the rank turns
// on the value rather than on the key's presence. `align: "left"` is a
// 1.0 document.
func styleVersionRank(st Style) versionRank {
	rank := rankBase
	if st.LineSpacing.Set || st.Color.Set {
		rank = rankMinorFeature
	}
	if st.Align.Set && !st.Align.Null && st.Align.Value == AlignJustify {
		rank = rankMajorFeature
	}
	return rank
}
