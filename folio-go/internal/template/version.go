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
const (
	SupportedMajor   = 1
	SupportedVersion = "1.1"
)

// baseVersion is the lowest version any document can declare, and the
// version a document whose content requires nothing newer keeps.
//
// minorFeatureVersion is the version introduced by the two 1.1 keys.
// They are named rather than spelled inline so versionRequiredByContent
// reads as the rule rather than as string handling.
const (
	baseVersion         = "1.0"
	minorFeatureVersion = "1.1"
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
// Today that is exactly one question — does any style block anywhere in
// the document set `lineSpacing` or `color`? Both are 1.1's optional
// keys; a document using neither is expressible in 1.0 and must keep
// declaring it.
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
	for _, band := range []Band{d.Bands.PageHeader, d.Bands.Content, d.Bands.PageFooter} {
		for _, el := range band.Elements {
			if el.Style.Set && !el.Style.Null && styleNeedsMinorVersion(el.Style.Value) {
				return minorFeatureVersion
			}
			if el.Table.Set && !el.Table.Null {
				hs := el.Table.Value.HeaderStyle
				if hs.Set && !hs.Null && styleNeedsMinorVersion(hs.Value) {
					return minorFeatureVersion
				}
			}
		}
	}
	return baseVersion
}

// styleNeedsMinorVersion reports whether one style block carries a key
// that 1.0 cannot express.
func styleNeedsMinorVersion(st Style) bool {
	return st.LineSpacing.Set || st.Color.Set
}
