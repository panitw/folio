package licence

import "strings"

// This file implements D-8.4i.1's and D-8.4i.2's rule: a licence text is
// classified by COLLECTING EVERY SIGNAL it carries and resolving them as
// one, rather than by returning on the first signal found.
//
// It is deliberately a separate file from classify.go. Story 8.4i's task
// 1 (D-8.4i.1's Block If, D-8.4i.6's hard constraint) requires the rule
// to be MEASURED over the whole existing population, in REPORT-ONLY
// mode, and that measurement to be committed, BEFORE any refusal becomes
// fatal. Landing the rule here, called only by the census test, makes
// "this commit changes no verdict" checkable by reading the diff: it
// touches no switch arm and no returned Family in classify.go.
//
// The semantics are not new policy. ClassifySPDXExpression
// (classify.go:14) has implemented them for a compound SPDX declaration
// since Story 1.3: start permissive, go copyleft if ANY term is
// copyleft, go unknown if ANY term is unrecognised. D-8.4i.1's ruling is
// that ClassifyLicenceText agrees with the function it sits beside.

// licenceName is one licence NAME this classifier can recognise in
// running text, together with the conditions under which that name
// RESOLVES to an SPDX identifier.
//
// The distinction between DETECTING a name and RESOLVING it is the whole
// of D-8.4i.2. Before this story the two were the same thing: a text
// whose licence name was half-recognised — the American spelling of the
// Ubuntu Font Licence, or OFL version 1.0 — simply missed its branch and
// fell into the MIT catch-all below it, and was published under MIT's
// label in lint/MANIFEST.md, a release artifact under AD-26. So the miss
// was SILENT, not loud, and D-2.1.3's fail-safe argument (a loud miss is
// fail-safe) did not apply to it at all.
//
// Here, detecting a name the classifier cannot resolve is itself a
// verdict: FamilyUnknown, which D-1.3.4 makes a build failure. Not
// knowing must not read as fine (D-8.5.2).
type licenceName struct {
	// canonical is the name spelling that, together with
	// requiredVersion, RESOLVES to id. Upper case; matched as a
	// substring of the upper-cased text.
	canonical string

	// otherSpellings are further spellings of the SAME licence name.
	// They are NAME SIGNALS — their presence means the text names a
	// licence — but they never resolve, so a text carrying only one of
	// them is FamilyUnknown.
	//
	// This is D-8.4i.2's ruling in code, and it is deliberately NOT
	// "accept both spellings": accepting the American spelling of the
	// Ubuntu Font Licence would fix one instance of a defect that has
	// now produced three (British/American UFL, OFL 1.0, and the
	// general near-miss class). Anchoring the name is the axis; the
	// spelling is an instance.
	otherSpellings []string

	// requiredVersion, when non-empty, must ALSO appear in the text for
	// the name to resolve. A name signal whose version conjunct is
	// absent is an unresolved name signal, not a miss.
	requiredVersion string

	// id is the SPDX identifier a resolved (name, version) pair yields.
	id string
}

// licenceNames is the ordered name table. ORDER IS LOAD-BEARING and is
// documented rather than incidental: it is the tie-break when several
// copyleft names appear in one text, and it preserves, exactly, the
// precedence the switch in ClassifyLicenceText had before Story 8.4i.
//
// Most specific first. The AGPL and LGPL legal codes both refer to the
// GNU General Public License in their own bodies, so an AGPL text
// carries BOTH the AGPL name signal and the GPL one; first-in-this-table
// wins, which reports AGPL-3.0 — the same identifier the old
// first-match switch reported. A gate whose message varies by map
// iteration order is not a gate.
var licenceNames = []licenceName{
	{canonical: "GNU AFFERO GENERAL PUBLIC LICENSE", id: "AGPL-3.0"},
	{canonical: "SERVER SIDE PUBLIC LICENSE", id: "SSPL-1.0"},
	{canonical: "GNU LESSER GENERAL PUBLIC LICENSE", id: "LGPL-3.0"},
	{canonical: "GNU GENERAL PUBLIC LICENSE", id: "GPL-3.0"},
	{canonical: "SIL OPEN FONT LICENSE", requiredVersion: "VERSION 1.1", id: "OFL-1.1"},
	{
		// THE BRITISH SPELLING "LICENCE" REMAINS THE ONLY RESOLVING
		// ONE — it is the licence's own spelling of its own name — but
		// the American spelling is now a NAME SIGNAL, so a file writing
		// "UBUNTU FONT LICENSE" reaches FamilyUnknown instead of
		// falling into MIT (DW-124).
		canonical:       "UBUNTU FONT LICENCE",
		otherSpellings:  []string{"UBUNTU FONT LICENSE"},
		requiredVersion: "VERSION 1.0",
		id:              "Ubuntu-font-1.0",
	},
	{canonical: "MIT LICENSE", id: "MIT"},
	{canonical: "APACHE LICENSE", id: "Apache-2.0"},
	{canonical: "BSD 3-CLAUSE", id: "BSD-3-Clause"},
	// PRE-EXISTING AND DELIBERATELY UNCHANGED HERE: a text naming the
	// 2-clause BSD licence has reported the 3-clause identifier since
	// Story 1.3 (classify.go's BSD case put both names in one
	// disjunct). It is a mislabel of exactly the class this story
	// exists to stop, but it is NOT one of the four findings this story
	// was chartered for, and D-8.4i.6 bounds Epic 8's licence-gate
	// insertions at 8.4i. Registered as DW-127 and routed to Epic 15's
	// release gate rather than fixed here. Not live: measured over all
	// 35 committed and dependency licence texts, no file in the
	// population reaches the BSD branch by NAME — the seven Go-style
	// BSD dependency licences reach it through the clause below, and
	// are genuinely 3-clause. permissiveSPDX carries "BSD-2-Clause" as
	// its own key, so the SPDX-line path is already correct.
	{canonical: "BSD 2-CLAUSE", id: "BSD-3-Clause"},
	{canonical: "CC0 1.0 UNIVERSAL", id: "CC0-1.0"},
}

// licenceClause is a licence GRANT CLAUSE — a sentence out of a
// licence's body rather than its name.
//
// A clause is a WEAKER signal than a name, and that asymmetry is the
// repair D-8.4i.2 ordered. "PERMISSION IS HEREBY GRANTED, FREE OF
// CHARGE" is shared verbatim by MIT, OFL and UFL, so as a peer of the
// name signals it would make the committed OFL 1.1 text carry two
// identifiers (OFL-1.1 and MIT) and classify as a conflict. Clauses are
// therefore consulted ONLY when the text names no licence at all — which
// is precisely the bare-grant-clause case the MIT fallback was written
// for, and the shape seven of the nine dependency LICENSE files rely on
// (Design Note 1: go-cmp, x/mod, x/net, x/sync, x/sys, x/telemetry,
// x/tools are Go-style BSD texts with no name signal and no SPDX line).
type licenceClause struct {
	clause string
	id     string
}

// licenceClauses is ordered for the same reason licenceNames is.
var licenceClauses = []licenceClause{
	{clause: "PERMISSION IS HEREBY GRANTED, FREE OF CHARGE", id: "MIT"},
	// A CLAUSE, NOT A NAME (Design Note 1, and this story's task 3
	// verifies it stays one): the Go standard library's own licence and
	// six of the eight other dependency licences reach BSD-3-Clause
	// through this sentence and through nothing else. Treating it as a
	// name would turn seven of the nine dependency licences into
	// FamilyUnknown and red the dependency scan on Go's own licence.
	{clause: "REDISTRIBUTION AND USE IN SOURCE", id: "BSD-3-Clause"},
}

// licenceSignals is everything one licence text says about itself.
type licenceSignals struct {
	// copyleftIDs are the copyleft identifiers the text carries, in the
	// deterministic order described on resolveLicenceSignals.
	copyleftIDs []string
	// permissiveIDs are the DISTINCT permissive identifiers the text
	// carries, in order of first appearance.
	permissiveIDs []string
	// unresolved is true if the text carries a signal this classifier
	// recognises as a licence signal but cannot resolve: a licence name
	// with no known (name, version) pair, or an SPDX identifier that is
	// on neither the permissive list nor the copyleft prefix list.
	unresolved bool
	// unresolvedID is the FIRST unrecognised SPDX identifier the text
	// declared, or "" if the only unresolvable signal was a licence
	// NAME.
	//
	// It exists so an unrecognised SPDX line still NAMES the identifier
	// it could not place (AD-14: a diagnostic that locates). This is
	// pre-existing behaviour, kept deliberately: the wordlist gate's
	// three refusal arms are proved DISJOINT by
	// TestWordlistSiteEnforcesThePermissiveSetNotTheFontAllowlist, and
	// a declared-but-unknown id such as CC-BY-SA-4.0 must reach its own
	// "not a permissive licence" arm rather than the neighbouring
	// "could not be classified" one. An unresolved NAME has no
	// identifier to name — that is exactly what makes it unresolved —
	// so it correctly returns "".
	unresolvedID string
}

// collectLicenceSignals gathers every licence signal in text.
//
// Three kinds, in decreasing strength:
//
//  1. SPDX-License-Identifier lines — the text's own explicit
//     declaration. EVERY line is collected, not just the first. The
//     first-match-and-return this replaces is DW-125: a full GNU GPL v3
//     text carrying a stray "SPDX-License-Identifier: MIT" line anywhere
//     below it classified (permissive, MIT) and passed the fail-closed
//     asset gate Story 8.4h had just built.
//  2. Licence NAMES — see licenceName. A name that does not resolve
//     marks the whole text unresolved.
//  3. Grant CLAUSES — see licenceClause. Consulted only when the text
//     carries no name signal at all.
func collectLicenceSignals(text string) licenceSignals {
	upper := strings.ToUpper(text)
	var sig licenceSignals

	seen := map[string]bool{}
	fromSPDXLine := true
	addID := func(id string) {
		if seen[id] {
			return
		}
		seen[id] = true
		switch classifyBySPDX(id) {
		case FamilyCopyleft:
			sig.copyleftIDs = append(sig.copyleftIDs, id)
		case FamilyPermissive:
			sig.permissiveIDs = append(sig.permissiveIDs, id)
		default:
			// An identifier this classifier does not recognise is not
			// nothing — it is a licence the build cannot show to be
			// permitted (D-1.3.4, D-1.3.8).
			sig.unresolved = true
			if fromSPDXLine && sig.unresolvedID == "" {
				sig.unresolvedID = id
			}
		}
	}

	// (1) Every SPDX line, in document order. Collected first so that a
	// file's own explicit declaration is the identifier named in a
	// copyleft refusal when the text also carries a copyleft NAME.
	for _, m := range spdxLineRE.FindAllStringSubmatch(text, -1) {
		addID(m[1])
	}

	// (2) Every licence name, in table order.
	fromSPDXLine = false
	namedAny := false
	for _, n := range licenceNames {
		spelled := strings.Contains(upper, n.canonical)
		canonicalSpelling := spelled
		for _, alt := range n.otherSpellings {
			if strings.Contains(upper, alt) {
				spelled = true
			}
		}
		if !spelled {
			continue
		}
		namedAny = true
		if !canonicalSpelling || (n.requiredVersion != "" && !strings.Contains(upper, n.requiredVersion)) {
			// The text names a licence this classifier cannot pin to a
			// known (name, version) pair. Half-recognised is not
			// recognised.
			sig.unresolved = true
			continue
		}
		addID(n.id)
	}

	// (3) Grant clauses, only when nothing named a licence.
	if !namedAny {
		for _, c := range licenceClauses {
			if strings.Contains(upper, c.clause) {
				addID(c.id)
			}
		}
	}

	return sig
}

// resolveLicenceSignals turns the collected signals into one verdict,
// in D-8.4i.1's FIXED order. The order is not stylistic:
//
//  1. ANY copyleft signal → FamilyCopyleft, NAMING THE IDENTIFIER.
//     Copyleft detection is not replaced by conflict detection and it
//     outranks it: a maintainer who reads "conflicting identifiers" adds
//     an SPDX line, while one who reads "GPL detected" removes the
//     dependency. Hazard indicators fail toward the loudest, never
//     toward the most precise.
//  2. Any unresolvable signal → FamilyUnknown, naming the unrecognised
//     SPDX identifier if there was one. A half-recognised licence name,
//     or an SPDX identifier on neither list, is not knowing — and not
//     knowing must not read as fine (D-8.5.2, D-1.3.4). An unresolved
//     NAME yields no identifier, because there is none to give.
//  3. Two or more DISTINCT permissive identifiers → FamilyUnknown. The
//     file says two things; the classifier does not pick one.
//  4. Exactly one identifier → that identifier, as before.
//  5. Nothing at all → FamilyUnknown, as before.
//
// THE COPYLEFT TIE-BREAK, when several copyleft signals appear: the
// FIRST entry of copyleftIDs wins, and that slice is built in a fixed
// order — SPDX lines in document order first, then licence names in
// licenceNames' most-specific-first table order. So an AGPL text (whose
// body also names the GNU GPL) reports AGPL-3.0, and a text declaring
// "SPDX-License-Identifier: AGPL-3.0-only" over a GPL body reports the
// declared identifier. Both are pinned by tests: the verdict must not
// vary by map iteration order.
func resolveLicenceSignals(sig licenceSignals) (Family, string) {
	if len(sig.copyleftIDs) > 0 {
		return FamilyCopyleft, sig.copyleftIDs[0]
	}
	if sig.unresolved {
		return FamilyUnknown, sig.unresolvedID
	}
	if len(sig.permissiveIDs) != 1 {
		// Zero, or two-or-more: either way this classifier cannot say
		// what licence the file is under.
		return FamilyUnknown, ""
	}
	return FamilyPermissive, sig.permissiveIDs[0]
}

// classifyByAllSignals is D-8.4i.1's and D-8.4i.2's classifier. Story
// 8.4i's task 1 commit calls it from the census test ONLY; task 2 makes
// ClassifyLicenceText's body a call to it.
func classifyByAllSignals(text string) (Family, string) {
	return resolveLicenceSignals(collectLicenceSignals(text))
}
