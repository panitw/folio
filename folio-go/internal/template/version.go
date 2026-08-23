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
// (D-1.4.9). SupportedVersion is the full "MAJOR.MINOR" this library
// itself would author for a brand-new document.
const (
	SupportedMajor   = 1
	SupportedVersion = "1.0"
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

// versionForSave is the single, well-named function isolating D-1.4.13's
// raise/lower rule: `version` is a property of the DOCUMENT, carried
// verbatim, raised only when content actually requiring a higher
// version is introduced, and lowered never.
//
// `version` describes the document, not the writer. A PDF 1.7 file
// edited by a tool that understands only 1.4 features does not become
// a PDF 1.4 file.
//
// At Story 1.4 the raise path is unreachable — no 1.1 exists yet, so
// there is no content this library can introduce that requires a
// higher MINOR than whatever was loaded. This function therefore
// always returns loaded, unchanged. It exists, and is called from
// exactly one place in serialize.go, so that the day a future MINOR
// exists, the raise path has exactly one place to be filled in — never
// by making the writer's own SupportedVersion the naive answer, which
// would violate BOTH halves of the rule at once (D-1.4.13): it would
// lower a document that already carries a higher MINOR, and it would
// gratuitously raise every unchanged older file, breaking
// Serialize(Parse(b)) == b for every one of them.
func versionForSave(loaded string) string {
	return loaded
}
