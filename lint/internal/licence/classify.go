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
//
// It is a thin wrapper over ClassifySPDXExpressionTerms, which is the
// SOLE term enumerator for an SPDX expression in this repository
// (Story 8.4j). This wrapper exists because both of ITS OWN callers —
// rules.ScanLicenceGraph and npm.go, neither of which gates per term —
// want only the family verdict.
func ClassifySPDXExpression(expression string) (Family, error) {
	family, _, err := ClassifySPDXExpressionTerms(expression)
	return family, err
}

// ClassifySPDXExpressionTerms is ClassifySPDXExpression with its term
// enumeration RETURNED rather than discarded. Story 8.4j (D-8.4j.1)
// added it for one reason: an asset gate must admit a compound
// expression only when EVERY term is on the list that governs that
// gate's population, and the alternative — splitting the expression
// string at the gate — is a second SPDX expression parser, a shape this
// run has now found four times. ONE parser, ONE term enumeration,
// THREE consumers: collectLicenceSignals, manifest's SITE A (the font
// gate, against the owner's four ids) and manifest's SITE B (the
// wordlist gate, against permissiveSPDX). The two gates keep their
// DIFFERENT POLICIES — which list governs which population — and share
// only the MECHANISM of testing a label term by term (D-8.4j.9).
//
// ITS RESOLUTION SEMANTICS ARE UNCHANGED from Story 1.3 and must stay
// so: trim, refuse parentheses and even field counts, require AND/OR at
// odd indices, start FamilyPermissive, ANY copyleft term makes the whole
// expression copyleft, ANY unrecognised term makes it unknown. The
// returned family and error are byte-for-byte what this function has
// always returned; only the terms are new.
//
// The terms slice carries the even-index fields discovered BEFORE the
// failure, so an unrecognised term is itself named in the slice — that
// is what lets the gate refuse naming the term that failed rather than
// naming only the expression. It is empty when the expression was too
// malformed to enumerate at all (parens, even field count), and a
// caller must treat "no terms" as "not admissible".
func ClassifySPDXExpressionTerms(expression string) (Family, []string, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" || strings.ContainsAny(expression, "()") {
		return FamilyUnknown, nil, fmt.Errorf("unsupported SPDX expression")
	}
	parts := strings.Fields(expression)
	if len(parts)%2 == 0 {
		return FamilyUnknown, nil, fmt.Errorf("malformed SPDX expression")
	}
	terms := make([]string, 0, (len(parts)+1)/2)
	family := FamilyPermissive
	for i, part := range parts {
		if i%2 == 1 {
			if part != "AND" && part != "OR" {
				return FamilyUnknown, terms, fmt.Errorf("unsupported SPDX operator %q", part)
			}
			continue
		}
		terms = append(terms, part)
		term := classifyBySPDX(part)
		if term == FamilyUnknown {
			return FamilyUnknown, terms, fmt.Errorf("unknown SPDX identifier %q", part)
		}
		if term == FamilyCopyleft {
			family = FamilyCopyleft
		}
	}
	return family, terms, nil
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

// spdxLineRE matches an "SPDX-License-Identifier: <expression>" marker,
// the preferred classification signal (AC30: "an SPDX identifier line or
// short marker" is all any fixture carries — never full licence text).
//
// STORY 8.4j (D-8.4j.1). The capture used to be `([A-Za-z0-9.\-+]+)`,
// ONE TOKEN, so a declaration of two licences at once —
// "SPDX-License-Identifier: OFL-1.1 OR GPL-3.0-only", an entirely
// ordinary way for a typeface to be published — was read as its FIRST
// TERM ALONE and classified (permissive, "OFL-1.1"). That id sits on the
// owner's four-id font allowlist, so the file passed the fail-closed
// asset gate and published under a permissive label in lint/MANIFEST.md.
// Reversing the two identifiers refused correctly: DW-125's
// order-dependence surviving in the one place Story 8.4i's collect-all
// rule did not reach, because collect-all worked ACROSS lines and this
// defect is WITHIN one. The capture is now the REST OF THE LINE, routed
// through ClassifySPDXExpression — the function that has resolved
// compound expressions correctly since Story 1.3 and which
// ClassifyLicenceText simply never asked.
//
// THREE DELIBERATE DETAILS, each with a reason:
//
//   - `[ \t]*` REPLACES `\s*` as the separator. `\s` matches newlines,
//     so the old pattern let a capture cross a line break. Harmless
//     against a one-token capture; against a rest-of-line capture it
//     would swallow an ARBITRARY FOLLOWING LINE — a copyright line, say
//     — as if it were an SPDX expression. The measured cost is that
//     "SPDX-License-Identifier:" with its identifier on the NEXT line
//     stops being a signal: it is not a valid SPDX declaration, no
//     population text uses it, and the direction is fail-closed.
//   - `\S` FIRST keeps a bare "SPDX-License-Identifier:" with nothing
//     after it a NON-SIGNAL, exactly as today, rather than a signal
//     carrying the empty expression.
//   - `[^\n\r]*` stops at the line end and tolerates CRLF.
//
// The known cost of reading the whole line is that trailing content
// after the identifier — "SPDX-License-Identifier: MIT */", a comment
// terminator or a parenthetical — now classifies (unknown, "") where it
// classified (permissive, "MIT"). That is the direct, intended price of
// reading the label whole, its direction is fail-closed and loud
// (D-1.3.4), and zero of the 35 population texts carry such a line. It
// is registered as a deferred finding on this story rather than
// papered over here.
var spdxLineRE = regexp.MustCompile(`(?i)SPDX-License-Identifier:[ \t]*(\S[^\n\r]*)`)

// ClassifyLicenceText inspects the content of a module's LICENSE file
// (or a short marker standing in for one) and returns its Family and,
// where recognised, its SPDX identifier. An unrecognised or absent
// marker returns FamilyUnknown — AC19/AC29 require that to fail the
// build, not pass silently (V10, V11): "a silent pass on an
// unidentifiable licence is the realistic failure mode" (D-1.3.8).
//
// STORY 8.4i (2026-09-02, D-8.4i.1 / D-8.4i.2). The body of this
// function was a first-match-and-return: one regexp for an SPDX line,
// then a switch of substring markers, returning on the first thing that
// matched. It now COLLECTS EVERY SIGNAL in the text and resolves them as
// one (licencesignals.go). Two defects made that necessary, and both had
// been measured before the change:
//
//   - DW-125, A GATE BYPASS. The SPDX regexp returned on the FIRST match
//     anywhere in the file, so a full GNU GPL v3 text carrying a stray
//     "SPDX-License-Identifier: MIT" line below it classified
//     (permissive, "MIT") and PASSED the fail-closed asset gate Story
//     8.4h had just built — the copyleft branches never ran. Measured
//     again at 8.4i's plan gate: the bypass works with a copyleft SPDX
//     LINE too ("MIT" line then "GPL-3.0-only" line classified
//     (permissive, "MIT")), which is why the fix is "collect every
//     signal" and not "a copyleft marker outranks an SPDX line".
//   - DW-124, A GREEDY CATCH-ALL. The MIT case matched on
//     "PERMISSION IS HEREBY GRANTED, FREE OF CHARGE", a clause MIT, OFL
//     and UFL share verbatim, so a NEAR-MISS on a licence NAME never
//     reached FamilyUnknown — it landed in MIT and was published under
//     MIT's label in lint/MANIFEST.md, a release artifact under AD-26.
//
// The semantics adopted are not new policy: ClassifySPDXExpression, at
// the top of this same file, has implemented them for a compound SPDX
// declaration since Story 1.3.
func ClassifyLicenceText(text string) (Family, string) {
	return classifyByAllSignals(text)
}

// The marker branches this function used to carry now live in
// licencesignals.go's licenceNames and licenceClauses tables. TWO OF
// THEIR COMMENTS ASSERTED A SAFETY THAT DID NOT EXIST, and they are
// corrected here rather than deleted, with their original wording
// preserved VERBATIM — because a comment asserting a branch is safe is
// exactly why nobody looked (D-8.0.1), and this is that pattern's third
// occurrence. Story 8.4h's review REJECTED the American-spelling finding
// BY QUOTING the second of these comments. Deleting the wording would
// leave the next reviewer the same trap with no record of it.
//
// ────────────────────────────────────────────────────────────────────
// CORRECTION 1 — the OFL branch (was classify.go:193-205, Story 2.2).
//
// ORIGINAL, VERBATIM:
//
//	"VERSION 1.1" is a REQUIRED conjunct, not decoration. Without
//	it this branch returned the SPDX id "OFL-1.1" for any text
//	containing the licence's name — including OFL **1.0**, whose
//	text carries the identical phrase, and including any
//	dependency LICENSE that merely bundles OFL text alongside its
//	own. The family verdict is permissive either way, so the
//	copyleft gate was never at risk; the LABEL was wrong, and the
//	label is what lands in lint/MANIFEST.md and attributes the
//	asset. The committed OFL text contains "Version 1.1" in its
//	own title line, so the conjunct costs nothing to satisfy and
//	makes a 1.0 file classify as FamilyUnknown — which D-1.3.4
//	deliberately makes a LOUD build failure rather than a quiet
//	mislabel. The CC0 marker below pins its version the same way.
//
// CORRECTION, MEASURED 2026-09-02 AT STORY 8.4i's PLAN GATE (baseline
// 7e4b2c4), BY PROBE, NOT INFERRED: the sentence "makes a 1.0 file
// classify as FamilyUnknown" was FALSE, and had been since Story 2.2.
// The SPDX-published OFL 1.0 text carries "Permission is hereby granted,
// free of charge, to any person obtaining a copy of the Font Software,
// to use, study, copy, merge, embed, modify, redistribute, and sell
// modified and unmodified copies of the Font Software, subject to the
// following conditions:" under the title line "SIL OPEN FONT LICENSE
// Version 1.0 - 22 November 2005", and contains "Version 1.1" NOWHERE.
// So it missed this branch, fell into the MIT catch-all immediately
// below it, and classified (permissive, "MIT") — a SILENT pass under the
// wrong label, not the loud failure claimed. Everything else in the
// original stands: the conjunct is required, and the label is what lands
// in lint/MANIFEST.md.
//
// D-8.4i.2 predicted this as "the fourth instance" and required the
// story to MEASURE it rather than take it on authority. It holds.
//
// WHAT NOW MAKES THE CLAIM TRUE: the name anchor. A text naming the SIL
// Open Font License without "VERSION 1.1" is an UNRESOLVED NAME SIGNAL,
// and an unresolved name signal is FamilyUnknown by itself — it can no
// longer fall through to a neighbouring permissive branch, because the
// grant clauses are consulted only when the text names no licence at
// all.
//
// ────────────────────────────────────────────────────────────────────
// CORRECTION 2 — the Ubuntu Font Licence branch (was classify.go:
// 231-249, Story 8.4h).
//
// ORIGINAL, VERBATIM:
//
//	"VERSION 1.0" IS A REQUIRED CONJUNCT, on the OFL precedent's
//	stated reasoning: without it this branch returns the id
//	"Ubuntu-font-1.0" for any text merely containing the
//	licence's name — a future UFL 1.1, or a dependency LICENSE
//	that bundles a UFL notice beside its own licence. The
//	licence's own title line carries the version, so the conjunct
//	costs nothing to satisfy, and a version this classifier does
//	not know becomes FamilyUnknown — a LOUD build failure rather
//	than a quiet mislabel (D-1.3.4).
//
//	THE BRITISH SPELLING "LICENCE" IS DELIBERATE, recorded here so
//	it is not "corrected" later. It is the licence's own spelling
//	of its own name. A file writing "UBUNTU FONT LICENSE" misses
//	this marker and classifies FamilyUnknown, which the asset gate
//	makes a build failure naming the directory — a LOUD miss, and
//	therefore fail-safe (D-2.1.3). Loosening the marker to accept
//	both spellings would trade that loud miss for a wider match
//	surface for no measured need: no such variant exists in this
//	repository, and none will until a face ships under it.
//
// CORRECTION, MEASURED 2026-09-02 (DW-124 at Story 8.4h's close, and
// again at 8.4i's plan gate), BY PROBE: both loud-miss claims were
// FALSE. A real Ubuntu Font Licence text spelled "UBUNTU FONT LICENSE"
// classified (permissive, "MIT") and PASSED the asset gate
// mis-attributed, because every real UFL text carries the shared grant
// clause and the MIT case sat immediately below this branch. The claim
// was true only of the bare two-line synthetic the story's own
// version-lookalike test used. D-2.1.3's fail-safe argument does not
// apply to a miss that lands in a NEIGHBOURING PERMISSIVE BRANCH rather
// than in FamilyUnknown — the miss was SILENT, which is the opposite of
// what the comment claimed, and "for no measured need" was written
// before anyone measured.
//
// THIS COMMENT IS WHY THE DEFECT SURVIVED REVIEW. Story 8.4h's review
// rejected the missing American-spelling test by quoting it.
//
// WHAT DID NOT CHANGE, AND WHY IT IS NOT "ACCEPT BOTH SPELLINGS"
// (D-8.4i.2, explicitly rejected): "UBUNTU FONT LICENCE" is still the
// only spelling that RESOLVES — it is the licence's own spelling of its
// own name. What changed is that "UBUNTU FONT LICENSE" is now a NAME
// SIGNAL: the text names a licence this classifier cannot pin to a known
// (name, version) pair, so it returns FamilyUnknown. That is the general
// anchor, not a spelling patch — it covers OFL 1.0, a future UFL 1.1,
// and every near-miss nobody has thought of yet.

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
