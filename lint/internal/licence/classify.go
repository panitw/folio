// Package licence implements AD-26's licence boundary (D-1.3.8,
// D-1.3.9): resolving a Go module's dependency graph, classifying each
// dependency's licence, and generating the manifest AC19 requires.
package licence

import (
	"regexp"
	"strings"
)

// Family is a coarse licence classification.
//
// There is no FamilyCommercialEULA (Finding 9, this story's QA review):
// AD-26's Rule names "GPL, LGPL, AGPL, SSPL, or a commercial EULA" as
// forbidden families, but this classifier has no marker table or SPDX
// convention to detect a commercial EULA specifically — no real project
// spells one identically enough to pattern-match the way the GNU family
// texts do. A commercial EULA still correctly FAILS the build today: it
// falls through every recognised marker to FamilyUnknown, and AC19/AC29
// already make an unresolvable licence a build failure. Reintroducing a
// FamilyCommercialEULA member with no classifier path reaching it, no
// fixture and no test (as this story originally shipped) is worse than
// this comment: it advertises detection that does not exist. If real
// EULA detection is ever built, it belongs here with its own SPDX/marker
// table, a classifier test case, and a fixture stub — not as a bare enum
// value.
type Family int

const (
	FamilyUnknown Family = iota
	FamilyPermissive
	FamilyCopyleft
)

func (f Family) String() string {
	switch f {
	case FamilyPermissive:
		return "permissive"
	case FamilyCopyleft:
		return "copyleft"
	default:
		return "unknown"
	}
}

// permissiveSPDX and copyleftSPDX are closed allow/deny lists of SPDX
// identifiers this classifier recognises. AD-26's Rule names the
// forbidden families explicitly: "GPL, LGPL, AGPL, SSPL, or a
// commercial EULA".
var permissiveSPDX = map[string]bool{
	"MIT": true, "Apache-2.0": true, "BSD-2-Clause": true, "BSD-3-Clause": true,
	"ISC": true, "0BSD": true, "Unlicense": true,
}

var copyleftSPDXPrefixes = []string{"GPL-", "LGPL-", "AGPL-", "SSPL-"}

// spdxLineRE matches an "SPDX-License-Identifier: <id>" marker, the
// preferred classification signal (AC30: "an SPDX identifier line or
// short marker" is all any fixture carries — never full licence text).
var spdxLineRE = regexp.MustCompile(`(?i)SPDX-License-Identifier:\s*([A-Za-z0-9.\-+]+)`)

// ClassifyLicenceText inspects the content of a module's LICENSE file
// (or a short marker standing in for one) and returns its Family and,
// where recognised, its SPDX identifier. An unrecognised or absent
// marker returns FamilyUnknown — AC19/AC29 require that to fail the
// build, not pass silently (V10, V11): "a silent pass on an
// unidentifiable licence is the realistic failure mode" (D-1.3.8).
func ClassifyLicenceText(text string) (Family, string) {
	if m := spdxLineRE.FindStringSubmatch(text); m != nil {
		id := m[1]
		return classifyBySPDX(id), id
	}

	upper := strings.ToUpper(text)
	switch {
	case strings.Contains(upper, "GNU AFFERO GENERAL PUBLIC LICENSE"):
		return FamilyCopyleft, "AGPL-3.0"
	case strings.Contains(upper, "SERVER SIDE PUBLIC LICENSE"):
		return FamilyCopyleft, "SSPL-1.0"
	case strings.Contains(upper, "GNU LESSER GENERAL PUBLIC LICENSE"):
		return FamilyCopyleft, "LGPL-3.0"
	case strings.Contains(upper, "GNU GENERAL PUBLIC LICENSE"):
		return FamilyCopyleft, "GPL-3.0"
	case strings.Contains(upper, "MIT LICENSE") || strings.Contains(upper, "PERMISSION IS HEREBY GRANTED, FREE OF CHARGE"):
		return FamilyPermissive, "MIT"
	case strings.Contains(upper, "APACHE LICENSE"):
		return FamilyPermissive, "Apache-2.0"
	case strings.Contains(upper, "BSD 3-CLAUSE") || strings.Contains(upper, "BSD 2-CLAUSE") || strings.Contains(upper, "REDISTRIBUTION AND USE IN SOURCE"):
		return FamilyPermissive, "BSD-3-Clause"
	}
	return FamilyUnknown, ""
}

func classifyBySPDX(id string) Family {
	if permissiveSPDX[id] {
		return FamilyPermissive
	}
	for _, p := range copyleftSPDXPrefixes {
		if strings.HasPrefix(id, p) {
			return FamilyCopyleft
		}
	}
	return FamilyUnknown
}
