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

// licenceNames is the ordered name table.
//
// ⚠ WHERE THE OLD BRANCHES' COMMENTS WENT. Two of these entries — OFL
// and the Ubuntu Font Licence — replace switch branches whose comments
// asserted a safety that did not exist. Those comments were NOT deleted:
// they are preserved VERBATIM, with their corrections, as CORRECTION 1
// and CORRECTION 2 in classify.go, immediately below ClassifyLicenceText.
// Read them before changing either entry's requiredVersion or spelling —
// the reasoning that looks obvious here is the reasoning that was
// measured false there.
//
// ⚠ EVERY id FIELD IN THIS TABLE AND IN licenceClauses MUST ALSO BE
// RECOGNISED BY classifyBySPDX — i.e. be a key of permissiveSPDX or
// carry a copyleftSPDXPrefixes prefix. This is a COUPLING STORY 8.4i
// INTRODUCED: addID routes every identifier, however it was signalled,
// through classifyBySPDX, so an id here that no list recognises does not
// classify as itself — it silently marks the whole text UNRESOLVED. The
// upside is that a name-table id is provably a real identifier rather
// than a free-text label; the cost is that the two structures can no
// longer be mutated independently (see
// TestUbuntuFontLicenceSPDXLineIsPermissive's note on the resolution
// this lost).
//
// ORDER IS LOAD-BEARING and is documented rather than incidental: it is the tie-break when several
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

// licenceClauses is ordered for the same reason licenceNames is, and its
// id fields carry the same obligation: each must be recognised by
// classifyBySPDX, or the clause marks the text unresolved instead of
// classifying it.
var licenceClauses = []licenceClause{
	{clause: "PERMISSION IS HEREBY GRANTED, FREE OF CHARGE", id: "MIT"},
	// A CLAUSE, NOT A NAME (Design Note 1, and this story's task 3
	// verifies it stays one). MEASURED by TestLicenceSignalCensus over
	// the nine dependency licences the three Go module graphs resolve
	// to: SEVEN of the nine — go-cmp, x/mod, x/net, x/sync, x/sys,
	// x/telemetry and x/tools, the same seven enumerated above — reach
	// BSD-3-Clause through this sentence and through nothing else,
	// carrying no SPDX line and no licence name at all. (The remaining
	// two, textshape and goldmark, classify MIT by name. The Go
	// standard library is NOT in this population — it is not a module
	// in any of the three graphs.) Treating this sentence as a name
	// would turn those seven into FamilyUnknown and red the dependency
	// scan.
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

// collectLicenceSignals gathers every licence signal in text. Three
// kinds:
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
//
// THIS IS A COLLECTION ORDER, NOT A PRECEDENCE ORDER, and an earlier
// draft of this comment called the three "decreasing strength" with SPDX
// first — which resolveLicenceSignals contradicts. Only ONE of the three
// is ranked against the others here: a clause is weaker than a name, and
// is consulted only in a name's absence. SPDX lines and names are peers;
// their identifiers go into the same two slices and are deduplicated
// together.
//
// Everything else is decided by resolveLicenceSignals, where the
// UNRESOLVED tier sits ABOVE the single-identifier arm. So an unresolved
// NAME outranks the text's own explicit declaration:
//
//	"SPDX-License-Identifier: MIT" + prose naming "UBUNTU FONT LICENSE"
//	→ (unknown, "")
//
// THAT DIRECTION IS DELIBERATE AND IS THE FAIL-SAFE ONE. A file that
// declares MIT while also naming a licence this classifier cannot place
// is a file whose governing licence is not established — which is
// exactly DW-125's shape with the roles swapped, and letting the
// declaration win is what the bypass did. Refusing costs a maintainer
// one SPDX line; believing the declaration costs a wrong label on a
// release artifact, or a forbidden font on a green build. Pinned by
// TestClassifyCollectsEverySignal.
func collectLicenceSignals(text string) licenceSignals {
	upper := strings.ToUpper(text)
	var sig licenceSignals

	seen := map[string]bool{}

	// declaredTerms holds the TERMS of every SPDX expression that
	// resolved in step (1). It is DELIBERATELY NOT the `seen` map, and
	// it is consulted by steps (2) and (3) ONLY — the NAME and CLAUSE
	// signal space. See markDeclaredTerms for why it exists and
	// D-8.4j.10 for why its scope is exactly this narrow.
	declaredTerms := map[string]bool{}

	// addID is the NAME and CLAUSE path only (steps 2 and 3). Step 1
	// stopped calling it at Story 8.4j, because an SPDX line no longer
	// carries a single identifier — see below. It therefore never sets
	// unresolvedID: an unresolved licence NAME names nothing, because
	// there is none to give (resolveLicenceSignals arm 2), and that was
	// already this closure's behaviour under the `fromSPDXLine` flag it
	// used to consult, which was false for every call it ever received
	// from steps 2 and 3.
	addID := func(id string) {
		if seen[id] || declaredTerms[id] {
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
		}
	}

	// (1) Every SPDX line, in document order. Collected first so that a
	// file's own explicit declaration is the identifier named in a
	// copyleft refusal when the text also carries a copyleft NAME.
	//
	// STORY 8.4j — THE COMPOSITION RULE. One SPDX line contributes
	// exactly ONE signal, WHATEVER ITS ARITY. Resolution WITHIN a line
	// is ClassifySPDXExpression's job; composition ACROSS lines and
	// against name signals is resolveLicenceSignals' job. Neither
	// reaches into the other, and this loop is the whole of the seam.
	//
	// Why one line is one signal: a declaration of "MIT OR Apache-2.0"
	// is a file saying ONE thing — "either of these" — not a file saying
	// two conflicting things. Treating its terms as two permissive
	// identifiers would hit resolveLicenceSignals' arm 3 (two or more
	// distinct permissive ids → unknown) and refuse a file that is
	// perfectly clear about its own terms.
	//
	// The capture is normalized to single-spaced fields first, so
	// "MIT  OR   Apache-2.0" and "MIT OR Apache-2.0" are ONE signal and
	// dedup through the same `seen` map every other signal uses. The
	// dedup key is the WHOLE EXPRESSION and nothing else: a second SPDX
	// LINE is never suppressed by an earlier line's terms (D-8.4j.10 —
	// suppressing it made the verdict depend on which line came first,
	// a NEW order dependence introduced by the story whose subject is
	// order dependence). Two SPDX lines are two explicit declarations,
	// and a file declaring "MIT" on one line and "MIT OR Apache-2.0" on
	// another genuinely says two things: arm 3 fires, IN BOTH ORDERS.
	//
	// A whole expression must NOT be routed through classifyBySPDX: that
	// is an exact map lookup, which would call every compound expression
	// unrecognised.
	for _, m := range spdxLineRE.FindAllStringSubmatch(text, -1) {
		fields := strings.Fields(m[1])
		expression := strings.Join(fields, " ")
		if seen[expression] {
			continue
		}
		seen[expression] = true

		family, terms, err := ClassifySPDXExpressionTerms(expression)
		switch {
		case err != nil || family == FamilyUnknown:
			sig.unresolved = true
			// unresolvedID is set ONLY for a SINGLE-TERM expression —
			// today's behaviour, unchanged. A COMPOUND expression that
			// fails to resolve names nothing, exactly as an unresolved
			// licence NAME names nothing. THIS DETAIL IS LOAD-BEARING:
			// naming a malformed expression would move a font from the
			// asset gate's "could not be classified" arm to its "not one
			// of the permitted licences" arm, whose message would then
			// assert that the text CLASSIFIES AS something it
			// demonstrably does not. It would also break the
			// arm-disjointness
			// TestWordlistSiteEnforcesThePermissiveSetNotTheFontAllowlist
			// pins, and move DW-131's requested pin for the
			// parenthesised form "(MIT OR Apache-2.0)" away from
			// (unknown, "").
			if len(fields) == 1 && sig.unresolvedID == "" {
				sig.unresolvedID = expression
			}
		case family == FamilyCopyleft:
			sig.copyleftIDs = append(sig.copyleftIDs, expression)
			markDeclaredTerms(declaredTerms, terms)
		default:
			sig.permissiveIDs = append(sig.permissiveIDs, expression)
			markDeclaredTerms(declaredTerms, terms)
		}
	}

	// (2) Every licence name, in table order.
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

// markDeclaredTerms records the TERMS of a RESOLVED SPDX expression, so
// that a later NAME or CLAUSE signal for one of the expression's own
// terms does not count a second time (Story 8.4j).
//
// WHY IT IS REQUIRED: a real dual-licensed font ships the full text of
// ONE of its licences. Without this, "SPDX-License-Identifier: OFL-1.1
// OR Apache-2.0" above the committed OFL 1.1 body yields the expression
// AND the body's own OFL-1.1 name signal — two distinct permissive
// identifiers — and resolveLicenceSignals' arm 3 refuses a file that
// says exactly one thing. Measured: (unknown, "") without this marking,
// (permissive, "OFL-1.1 OR Apache-2.0") with it.
//
// IT WRITES ITS OWN SET, NOT THE `seen` MAP THE SPDX LINES DEDUP
// THROUGH (D-8.4j.10). The justification above is entirely about a
// duplicated BODY signal, and it does not extend to a second SPDX
// LINE: two SPDX lines are two explicit declarations, and a file
// declaring "MIT" on one line and "MIT OR Apache-2.0" on another
// genuinely says two things. Written into the shared map instead, this
// marking swallowed the later line and made the verdict depend on
// which order the two lines appeared in — measured: (unknown, "") one
// way and (permissive, "MIT OR Apache-2.0") the other. A NEW order
// dependence, in the story whose whole subject is order dependence.
//
// IT MARKS TERMS, NOT WHITESPACE-SEPARATED FIELDS. Fields would mark
// the OPERATORS "OR" and "AND" as declared ids — inert today only
// because no licence is named "OR" — and the terms come from the
// SINGLE enumerator (ClassifySPDXExpressionTerms), never from a
// re-split here. That is the same one-parser guardrail the asset
// gates' per-term admission follows.
//
// IT CANNOT MASK A DIFFERENT LICENCE'S SIGNAL, because only the terms
// the expression ITSELF declared are marked: a "MIT OR Apache-2.0"
// header over a GPL body still refuses as copyleft (re-measured after
// this scoping changed, not carried forward on inspection).
func markDeclaredTerms(declaredTerms map[string]bool, terms []string) {
	for _, t := range terms {
		declaredTerms[t] = true
	}
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
