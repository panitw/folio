// Package licence implements AD-26's licence boundary (D-1.3.8,
// D-1.3.9): resolving a Go module's dependency graph, classifying each
// dependency's licence, and generating the manifest AC19 requires.
package licence

import (
	"fmt"
	"regexp"
	"strings"
)

// ClassifySPDXExpression accepts a known SPDX identifier or a simple known
// conjunction/disjunction. Unknown syntax or terms fail closed.
func ClassifySPDXExpression(expression string) (Family, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" || strings.ContainsAny(expression, "()") {
		return FamilyUnknown, fmt.Errorf("unsupported SPDX expression")
	}
	parts := strings.Fields(expression)
	if len(parts)%2 == 0 {
		return FamilyUnknown, fmt.Errorf("malformed SPDX expression")
	}
	family := FamilyPermissive
	for i, part := range parts {
		if i%2 == 1 {
			if part != "AND" && part != "OR" {
				return FamilyUnknown, fmt.Errorf("unsupported SPDX operator %q", part)
			}
			continue
		}
		term := classifyBySPDX(part)
		if term == FamilyUnknown {
			return FamilyUnknown, fmt.Errorf("unknown SPDX identifier %q", part)
		}
		if term == FamilyCopyleft {
			family = FamilyCopyleft
		}
	}
	return family, nil
}

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
// CC0-1.0 joined this list at Story 2.1 (AC8, D-2.1.3), the first
// story to introduce a CC0 asset (PyThaiNLP's words_th wordlist) —
// measured, "CC0" appeared nowhere in the lint module before this.
//
// The general rule, so a later reader does not cite D-1.8.11 against
// this addition (D-2.1.3, binding): an allowlist whose miss is LOUD is
// a fail-safe; an allowlist whose miss is SILENT is a rotting list.
// permissiveSPDX is a fail-safe — an unrecognised licence classifies as
// FamilyUnknown, and D-1.3.4 deliberately made that a build failure, so
// a miss here is loud and adding a correct entry is ordinary
// maintenance. D-1.8.11 forbids extending manifest.go's fontExtensions
// instead: THAT list's miss is silent (an unrecognised extension is
// simply never scanned, and nothing is ever said) — same data
// structure shape, opposite failure mode. The two must not be
// conflated because they look alike in a diff.
var permissiveSPDX = map[string]bool{
	"MIT": true, "Apache-2.0": true, "BSD-2-Clause": true, "BSD-3-Clause": true,
	"ISC": true, "0BSD": true, "Unlicense": true, "CC0-1.0": true,
	"MIT-0": true, "BlueOak-1.0.0": true, "CC-BY-4.0": true,
	// OFL-1.1 joined this list at Story 2.2 (AC2, AD-26), the first
	// story to introduce an OFL-licensed asset (the three shipped Noto
	// faces) — measured, "OFL" appeared nowhere in the lint module
	// before this, and its full legal code was misclassified as "MIT"
	// (see ClassifyLicenceText) until this story's fix.
	"OFL-1.1": true,
	// Ubuntu-font-1.0 joined this list at Story 8.4h (2026-09-02,
	// D-8.5.3), the story that made the asset licence gate fail closed
	// against the owner's four-id allowlist {OFL-1.1, Apache-2.0, MIT,
	// UFL}. Measured, "Ubuntu" appeared nowhere in the lint module
	// before this, so the fourth member of that allowlist had no
	// implementation at all: a face under it would have been refused by
	// the very gate the owner's decision permits it through.
	//
	// THE IDENTIFIER IS "Ubuntu-font-1.0", VERIFIED, NOT REMEMBERED.
	// https://spdx.org/licenses/Ubuntu-font-1.0.html returns HTTP 200
	// ("Ubuntu Font Licence v1.0"); "UFL-1.0" and "Ufont-1.0" both
	// return 404. The owner's list spells its fourth member "UFL", the
	// community abbreviation, while spelling the other three as exact
	// SPDX identifiers — so "UFL" denotes the licence and this is its
	// canonical id. The bare alias is deliberately NOT a key here:
	// classifyBySPDX is an exact map lookup, so "UFL" could only ever be
	// produced by a LICENSE writing `SPDX-License-Identifier: UFL`,
	// which is not a valid SPDX line and which no real font ships. It
	// would be dead code that looks live — precisely the "advertises
	// detection that does not exist" failure this package's Family doc
	// comment was written to prevent (Design Note 3, Story 8.4h).
	//
	// NOTHING IN THIS REPOSITORY IS UBUNTU-LICENSED TODAY. The entry is
	// proved by the classifier table, by the SPDX-line fixture module
	// under testdata/licence/permissive/, and by nothing else; there is
	// no analogue of TestCommittedOFLTextClassifiesAsOFL11 to write
	// until Story 8.5 lands a face under it (Design Note 8).
	"Ubuntu-font-1.0": true,
}

// IsPermissiveSPDX reports whether id is one of the SPDX identifiers
// this classifier recognises as permissive. It exists so a consumer in
// another package (manifest.resolveWordlistAssetRow, Story 8.4h) can
// consult THIS map rather than keep a second copy of the list: a
// duplicated list is a list the code can move, and the two copies drift
// apart silently the first time either moves (D-8.5.8c).
//
// NOTE FOR THE NEXT READER: this predicate is NOT the asset allowlist.
// manifest.go's font path enforces the owner's four-id decision
// (D-8.5.3) against its own, deliberately smaller, font-path-local
// constant. The two lists are independent in SCOPE, not merely in
// mechanism — CC0-1.0 is permissive and is legitimately the Thai
// wordlist's licence, and is emphatically not an acceptable font
// licence. Do not route the font path through here.
func IsPermissiveSPDX(id string) bool { return permissiveSPDX[id] }

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
	case strings.Contains(upper, "SIL OPEN FONT LICENSE") && strings.Contains(upper, "VERSION 1.1"):
		// Story 2.2 (AC2, AD-26): measured, real misclassification this
		// story is the first to exercise — the OFL 1.1 full legal code's
		// own grant clause opens "Permission is hereby granted, free of
		// charge, to any person obtaining a copy of the Font Software...",
		// the same MIT-detection substring the case below matches on,
		// so every one of this story's three OFL-licensed faces was
		// classified as "MIT" until this case was added ABOVE the MIT
		// check. "SIL OPEN FONT LICENSE" is the licence's own name,
		// appearing at the top of the committed OFL.txt text, and is not
		// a substring of the MIT, Apache, or BSD legal codes.
		//
		// "VERSION 1.1" is a REQUIRED conjunct, not decoration. Without
		// it this branch returned the SPDX id "OFL-1.1" for any text
		// containing the licence's name — including OFL **1.0**, whose
		// text carries the identical phrase, and including any
		// dependency LICENSE that merely bundles OFL text alongside its
		// own. The family verdict is permissive either way, so the
		// copyleft gate was never at risk; the LABEL was wrong, and the
		// label is what lands in lint/MANIFEST.md and attributes the
		// asset. The committed OFL text contains "Version 1.1" in its
		// own title line, so the conjunct costs nothing to satisfy and
		// makes a 1.0 file classify as FamilyUnknown — which D-1.3.4
		// deliberately makes a LOUD build failure rather than a quiet
		// mislabel. The CC0 marker below pins its version the same way.
		return FamilyPermissive, "OFL-1.1"
	case strings.Contains(upper, "UBUNTU FONT LICENCE") && strings.Contains(upper, "VERSION 1.0"):
		// Story 8.4h (2026-09-02, D-8.5.3): the fourth member of the
		// owner's asset allowlist, added LOUDLY — a map entry above and
		// this marker branch — rather than by widening any list
		// silently. Both halves are required and neither substitutes for
		// the other: this branch returns its id directly and never
		// consults permissiveSPDX, so a map entry alone leaves full UFL
		// text misclassifying as MIT; and a branch alone leaves
		// "SPDX-License-Identifier: Ubuntu-font-1.0" classifying as
		// FamilyUnknown.
		//
		// PLACED ABOVE THE MIT CASE, character-for-character the OFL
		// defect above. Measured against the SPDX-published UFL 1.0
		// text: it opens "UBUNTU FONT LICENCE Version 1.0" and carries
		// "Permission is hereby granted, free of charge, to any person
		// obtaining a" — the EXACT substring the MIT case below matches
		// on. Below MIT, every UFL-licensed face classifies as
		// (permissive, "MIT"): right family, WRONG LABEL, and the label
		// is what lands in lint/MANIFEST.md and attributes the asset.
		// Audited against every other branch: the UFL text contains
		// neither "SIL OPEN FONT LICENSE" nor "VERSION 1.1", and
		// collides with none of the four copyleft cases, nor Apache,
		// BSD or CC0.
		//
		// "VERSION 1.0" IS A REQUIRED CONJUNCT, on the OFL precedent's
		// stated reasoning: without it this branch returns the id
		// "Ubuntu-font-1.0" for any text merely containing the
		// licence's name — a future UFL 1.1, or a dependency LICENSE
		// that bundles a UFL notice beside its own licence. The
		// licence's own title line carries the version, so the conjunct
		// costs nothing to satisfy, and a version this classifier does
		// not know becomes FamilyUnknown — a LOUD build failure rather
		// than a quiet mislabel (D-1.3.4).
		//
		// THE BRITISH SPELLING "LICENCE" IS DELIBERATE, recorded here so
		// it is not "corrected" later. It is the licence's own spelling
		// of its own name. A file writing "UBUNTU FONT LICENSE" misses
		// this marker and classifies FamilyUnknown, which the asset gate
		// makes a build failure naming the directory — a LOUD miss, and
		// therefore fail-safe (D-2.1.3). Loosening the marker to accept
		// both spellings would trade that loud miss for a wider match
		// surface for no measured need: no such variant exists in this
		// repository, and none will until a face ships under it.
		return FamilyPermissive, "Ubuntu-font-1.0"
	case strings.Contains(upper, "MIT LICENSE") || strings.Contains(upper, "PERMISSION IS HEREBY GRANTED, FREE OF CHARGE"):
		return FamilyPermissive, "MIT"
	case strings.Contains(upper, "APACHE LICENSE"):
		return FamilyPermissive, "Apache-2.0"
	case strings.Contains(upper, "BSD 3-CLAUSE") || strings.Contains(upper, "BSD 2-CLAUSE") || strings.Contains(upper, "REDISTRIBUTION AND USE IN SOURCE"):
		return FamilyPermissive, "BSD-3-Clause"
	case strings.Contains(upper, "CC0 1.0 UNIVERSAL"):
		// Story 2.1 (AC8, D-2.1.3): the committed LICENSE-CC0-1.0.txt is
		// the CC0 1.0 Universal legal code in full (matching the
		// Apache/BSD full-text precedent above, not an SPDX-only
		// marker), so it needs its own fallback text match, not just
		// the SPDX-line path. The marker is deliberately narrow: an
		// earlier version of this fallback also matched "CREATIVE
		// COMMONS CORPORATION IS NOT A LAW FIRM", which is boilerplate
		// opening EVERY Creative Commons legal code (CC BY, CC BY-SA,
		// CC BY-NC, CC BY-ND, ...), not just CC0 — that over-broad
		// marker classified the entire CC family, including
		// NonCommercial and ShareAlike variants, as permissive CC0-1.0
		// (this story's second QA review, Major 1). "CC0 1.0
		// UNIVERSAL" alone is sufficient for the committed file and
		// does not appear in any other CC licence's legal code.
		return FamilyPermissive, "CC0-1.0"
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
