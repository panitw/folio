# PRD Quality Review — Folio (MVP v0.1)

## Overall verdict

This is a genuinely opinionated PRD with a real thesis — *reliability over feature breadth*, enforced by a byte-reproducibility requirement (NFR1) that the counter-metrics actively police. Scope discipline, assumption honesty, and shape fit are all above the bar for a solo pre-architecture document, and the addendum does its job of keeping mechanism out of the body. What is at risk is done-ness: the highest-risk capability in the MVP (FR24, table page-breaking) is specified with the word "correctly" and nothing else, the nine Designer FRs carry no falsifiable consequences, and half the product's stated user population — the template author — has no success measure attached to it. The other exposure is strategic: NFR1 is stated as an absolute Must while Q1 asks whether it is achievable at all, and the PRD names no position to fall back to if the answer is no.

---

## Decision-readiness — adequate

The PRD does the hard thing in several places. §2 states a positioning bet as a bet ("**No migration path from JasperReports is offered in MVP.** Folio competes on the workflow, not on import fidelity") rather than hedging it. §4.3 makes a discriminating call — the layout model is protected for future PNG/SVG/HTML renderers, but "Excel is *not* protected in this way" with the reason given. §5.5 states a priority rule that resolves a real conflict in advance: "**Correct pagination takes priority over table styling** wherever the two compete." FR40 openly declares itself an invention ("this requirement is invented here because a library without one is not integrable"). These are decisions, not considerations, and a reader pushing back would find the trade-off named.

The Open Questions are also mostly real. Q5 and Q6 turn the source plan's own build order against it — Q6 asks "Does the designer really come last (Phase 5), when it is success criterion #1 and half the product's users?" That is a PRD arguing with its input, which is the correct posture.

Where decision-readiness thins is on the one question that matters most. NFR1 is written as an absolute — "**must produce a byte-identical PDF**" — and C3 declares that "Any determinism exception carved out to ship a feature" is a fatal signal. But Q1 asks whether byte-identical output is even achievable across WASM and native, and the PRD's own answer is that it "may rule out most off-the-shelf PDF libraries," i.e. it may require hand-rolling a PDF writer. A PRD that stakes its identity on a constraint whose feasibility is unresolved needs to state what happens if the constraint does not hold — does the MVP fall back to layout equivalence, does the WASM preview stop being hash-exact, or does the project stop? None of those are named. Relatedly, C4 ("Designer work starting before the engine renders the golden report" is a bad signal) and Q6 (should the designer really be last?) point in opposite directions with neither acknowledging the other.

There are also zero `[NOTE FOR PM]` callouts in the document. The `[ASSUMPTION]` prose partly absorbs that role — FR16's "this needs an explicit definition before implementation" is effectively a PM note — but the genuine unresolved tensions (NFR1 fallback, C4-vs-Q6) sit in no callout at all.

### Findings
- **critical** NFR1 is absolute but its feasibility is an open question, with no stated fallback (§6 NFR1 / §8 Q1 / §7.3 C3) — The PRD says byte-identity "**must**" hold and calls any exception "fatal," while Q1 concedes the approach may require hand-rolling a PDF writer. Architecture cannot proceed without knowing whether NFR1 is a hard gate or a target. *Fix:* add a decision rule to NFR1 — e.g. "if cross-target byte-identity proves infeasible within N weeks of spike, MVP degrades to layout-equivalence + a documented hash-per-target, and FR34's 'the previewed document is the production document' weakens to 'visually identical'." Name who decides and when.
- **high** C4 and Q6 give opposite sequencing guidance (§7.3 C4 / §8 Q6) — C4 treats designer-before-engine as a warning signal; Q6 argues the designer is being deferred too far. Downstream epic sequencing will pick whichever it reads last. *Fix:* resolve to one position, or state the reconciliation explicitly (e.g. "engine-first stands; Q6 concerns *preview integration*, not designer start").
- **high** NFR1 escalates the source plan without flagging it as an invention (§6 NFR1 vs. `docs/folio-mvp-plan.md` §10 / addendum §E) — The plan says "The same template and data should generate the same **layout** in every environment"; the addendum records this correctly as "Stated at **layout** granularity ('same layout'), never as byte-for-byte PDF equality." The PRD escalates to byte-identity and calls it "the single hardest constraint in the MVP" — a larger invention than FR40, which *is* tagged `[ASSUMPTION]`. *Fix:* tag NFR1 with `[ASSUMPTION]` and add it to §9, naming the escalation and its cost.

---

## Substance over theater — strong

There are exactly two personas, both load-bearing, and §3 makes their relationship the product rather than decoration: "The `.folio` file is the **contract between them**. Every requirement in this document should be read against the question: *does this make the handoff more reliable?*" §3 also does the rarer thing of naming who is *not* a user — "no operator or administrator … the end recipient of the PDF is not a Folio user — they receive a document, not an experience."

The NFRs are mostly product-specific rather than boilerplate. NFR3 does not say "must support internationalization"; it names three scripts and their three distinct breaking algorithms, then adds the consequence that follows from NFR1 ("the breaking algorithm and any dictionary must be embedded and versioned with the library, never taken from the host platform"). NFR2 is a genuine architectural prohibition with a distinguishing test ("`folio-go` compiled to WebAssembly *is* `folio-go`, and satisfies this; browser DOM-based rendering does not"). The §1 governing principle is concrete enough to fail against — "a professional 20–50 page enterprise statement."

Two soft spots. NFR5 restates NFR1 and NFR2's consequence rather than adding a requirement; it is true but it is not independently testable. NFR8 states a posture rather than a requirement — "This is a property worth stating and protecting" is not something an engineer can build or verify. Neither is harmful, but both are the closest thing here to furniture.

### Findings
- **low** NFR5 has no independent content (§6 NFR5) — "The exact preview and the production render are the same engine and produce the same bytes (NFR1)" is a derivation of NFR1 + NFR2, not a separate constraint. *Fix:* fold into NFR1 as a consequence, or delete.
- **low** NFR8 states a posture, not a requirement (§6 NFR8) — Nothing here can be verified or violated. *Fix:* convert to a testable prohibition, e.g. "the designer makes no network request carrying template or data content; verified by a no-egress test."

---

## Strategic coherence — strong

The PRD has a thesis and the features follow from it rather than sitting beside it. The thesis is stated in §1 ("reliability over feature breadth") and given teeth in §2's property table, where every row is a deliberate difference from JasperReports rather than a feature claim. The scope decisions then trace back: exactly five components (FR4), exactly eight expression functions (FR17) with the explicit prohibition "It is not, and must not grow into, a general-purpose scripting language in MVP," and pagination privileged over styling (§5.5).

The counter-metrics are the strongest part of the document and are unusually well-targeted — they measure the thesis eroding, not activity. C1 (function count), C2 (palette count), and C3 (determinism exception) each name the specific way the *reliability over breadth* bet would be lost. C5 ("Time-to-first-PDF for a new integrator exceeding a few minutes") is the only user-experience counter-metric and it belongs to the Go developer.

The gap is that the §7.2 success measures do not cover both users of "equal standing." S1–S4 and S6 are engine properties; S5 is a round-trip that touches the designer only as a file producer. Nothing measures whether Ploy — the protagonist of UJ-1, and half the stated user population — can actually lay out and bind a report. C5 gives Anan a time-to-value metric; Ploy has none. Given §3's insistence that the two users are co-primary, this asymmetry undercuts the thesis the PRD says it is testing.

Scope kind is coherent — this is a problem-solving MVP (can the determinism/handoff workflow be made to work at all), and the acceptance-report framing in §7.1 matches that kind exactly.

### Findings
- **high** No success measure covers the template author (§7.2 S1–S6 vs. §3) — §3 declares "two primary users of equal standing," but every measure except S5 is an engine property, and the only time-to-value counter-metric (C5) belongs to the Go developer. *Fix:* add a measure for UJ-1 — e.g. "S7: the golden statement can be authored end to end in the designer (page setup, five bindings, one bound table with header repeat and sum footer) in under N minutes by the author, without hand-editing `.folio`."

---

## Done-ness clarity — thin

This is the weakest dimension and the one that will hurt most downstream. Several FRs are excellent — FR32 states both the requirement and the two prohibited implementations with reasons ("does **not** fetch images by URL and does **not** read them from the filesystem at render time; both would break determinism"). FR40 defines its error contract by content ("naming the template element and the data path") across four named failure classes. FR31's consequence is directly testable ("renders identically on a machine that has none of them installed"). FR12 gives a failure mode ("fails clearly rather than rendering incorrectly").

But the highest-risk requirement in the MVP is one word. FR24 reads in full: "Break a table across pages correctly." §5.5 has already told us table rendering is "the highest-risk area of the MVP," and the addendum §F escalates that to "expected to be one of the highest-risk engineering areas." Nothing in the PRD says what correct means: a row taller than a page, a row whose cell text wraps across the page boundary, orphan/widow rules for the last row before a break, a header-plus-first-row that does not fit, a footer aggregate row that lands alone on a continuation page. Each is a real decision, each will be made by whoever implements it, and none is here. FR25 ("Repeat the table header at the top of every continuation page") is testable; FR24 is not.

The Designer FRs are a second cluster. FR1–FR8 are capability names with no verifiable condition: FR3 is "Align work using a grid with snapping" (grid size? snap threshold? togglable?); FR5 enumerates thirteen editable properties but says nothing about what editing one must do; FR8 says "with no server round-trip and no account," which is at least falsifiable, but that is the exception. FR33 asks for a "fast **Design Canvas**" with no bound, and FR37 asks for an API that is "extremely simple" — a quoted adjective standing in for a contract, though C5's time-to-first-PDF partly rescues it.

Verification methods in §7.2 are honest but soft in two rows: S4 is "Verified by inspection" and S6 is bare "Verified." For a solo project, inspection is a legitimate method; naming *what* is inspected would still cost one line each.

### Findings
- **critical** FR24 specifies the MVP's highest-risk behaviour as "correctly" (§5.5 FR24) — No definition of correct page-breaking: oversized rows, mid-row text wrapping across a boundary, orphan/widow policy, header+first-row atomicity, or an isolated footer-aggregate row. The PRD names this area as highest-risk and then leaves its acceptance semantics to the implementer. *Fix:* expand FR24 into 4–6 named cases with the required behaviour for each, or add a §5.5 acceptance block enumerating them; the golden report at 50 pages should exercise every case.
- **high** Designer FRs carry no testable consequence (§5.1 FR1–FR7) — "Align work using a grid with snapping" (FR3) and "Edit component properties: position, size, font family …" (FR5) name capabilities without conditions. Downstream story creation will invent acceptance criteria for the entire designer. *Fix:* add one verifiable consequence per FR (e.g. FR3: "default grid 5mm, snapping toggleable, snap applies to both position and resize edges"), or add a §5.1 acceptance block covering the UJ-1 flow end to end.
- **medium** "Extremely simple" is the only bound on the library API (§5.8 FR37) — The adjective is quoted from the source plan, and the parenthetical "a load call and a render call, nothing ceremonial" is the real content. *Fix:* state the contract as a shape — "rendering a template to a writer requires no more than three calls and no configuration object" — and let C5 remain the outcome check.
- **medium** "Fast" Design Canvas has no bound (§5.7 FR33) — Interactivity target unspecified, which matters because NFR2 forbids using the browser's own layout engine to get it. *Fix:* give it a budget (e.g. "drag/resize feedback under 16ms at 200 components on a 50-page template").
- **low** Two success measures name no verification method (§7.2 S4, S6) — "Verified by inspection" and "Verified." *Fix:* name the artefact — S4: which strings, at which widths, checked against which reference; S6: peak RSS ceiling for the 50-page render.

---

## Scope honesty — strong

Omissions are explicit almost everywhere a reader could otherwise infer them. §4.2 lists twenty-plus out-of-scope items by name rather than gesturing at "advanced features." §4.3 is better still — it separates "deferred" from "deferred *and* architecturally protected," and then excludes Excel from that protection with a reason. Optional and capacity-dependent items are marked inline rather than silently ranked: FR27 "*(if capacity permits)*", FR42 "*(optional, only if capacity permits)*" with the priority stated ("The Go library remains the higher-priority runtime integration").

§10 does real work rather than performing thoroughness. Each omission carries a re-entry condition: the security model "must be revisited the moment the optional REST service (FR42) becomes real"; accessibility is "worth flagging as a likely requirement for any enterprise buyer, should Folio ever go public"; NFR3 is explicitly bounded away from a locale model. That is the difference between "we didn't do this" and "here is when this becomes your problem."

Open-items density is appropriate to stakes: nine Open Questions, nine inline assumptions, on a solo pre-architecture PRD whose next stop is `bmad-architecture` rather than a build green-light. Two of the assumptions are unusually honest about their own status — FR19 says "This boundary is drawn here for the first time — the source plan leaves it undrawn," and FR40 admits it is invented. Q7 records a non-decision as a decision ("Currently undecided by choice"), which is the right handling for the distribution question.

No findings.

---

## Downstream usability — adequate

This PRD is chain-top — it explicitly hands off to `bmad-architecture` via the addendum — so traceability matters more than it would for a standalone document. The mechanics are clean: FR1–FR42 and NFR1–NFR8 are contiguous with no gaps or duplicates, Q-items name what they block ("Blocks" column), and cross-references resolve (Q3→FR16, Q8→FR9, §10→FR42, NFR7→FR31/NFR1/NFR3). Sections mostly stand alone; §5.5's priority rule and §5.8's framing paragraph each carry their own context.

Two gaps. There is no Glossary, and the document leans on domain nouns that drift: "region" appears as "content region" (UJ-1) and "**Content**" (FR6); "sample JSON" appears as "sample JSON file" (UJ-1), "sample JSON data document" (FR9), "sample data" (NFR8), and "sample JSON" (Q8); "PDF Preview" (FR34) becomes "exact preview" (NFR5) and "production preview" (addendum §H Phase 6). None of these will mislead a human, but a story-generation pass that extracts nouns will produce three names for one thing.

More consequential: the addendum has drifted out of sync with the PRD on one point. Addendum §J.2 records "Preview API is unscoped work … If the designer is a web app and folio-go is a Go library, a server-side preview path is unavoidable — yet the only server surface in the plan is marked *Optional*." But the PRD resolved exactly that tension in FR34/FR35 by compiling `folio-go` to WebAssembly and running it in the tab — no server preview path is needed. Since the addendum is described as "the handoff payload for `bmad-architecture`," architecture will receive a live tension that the PRD already closed.

Q6 also references "success criterion #1," which is a criterion of `docs/folio-mvp-plan.md` §17, not of this PRD's §7.2 (where S1 is the golden report, not the designer). A reader working from the PRD alone cannot resolve the reference.

### Findings
- **medium** Addendum §J.2 contradicts the PRD's resolution of the preview tension (addendum §J.2 vs. §5.7 FR34/FR35) — FR34 specifies preview "produced by the real `folio-go` engine compiled to WebAssembly and running in the browser tab — not by a browser-based approximation, and not by a remote service," and FR35 requires fully offline operation. The addendum still carries a server-side preview path as an unresolved architecture tension. *Fix:* rewrite §J.2 to state the resolution and what it costs architecture instead (WASM toolchain, bundle size per Q9, NFR1 across compile targets).
- **medium** No Glossary, with drift on three domain nouns (throughout) — region / Content region; sample JSON / sample JSON data document / sample data; PDF Preview / exact preview / production preview. *Fix:* add a short Glossary fixing one canonical form for each and normalize usages.
- **low** Q6 cites a criterion that lives in the source plan, not this PRD (§8 Q6) — "it is success criterion #1" resolves against `docs/folio-mvp-plan.md` §17, not §7.2. *Fix:* cite the source explicitly or restate the criterion inline.

---

## Shape fit — strong

The document is shaped like what it is: a solo, engine-first, chain-top PRD for a technical product with two real users. Rigor is light where light is right — no team, timeline, budget, monetization, or GTM apparatus, all named as deliberate in §10 with reasons — while the substance bar holds on the parts that determine whether the thing works.

UJ density is correctly calibrated. Two journeys, one per primary user, each with a named protagonist carrying enough context inline that neither needs §3 to be read first: Ploy is "customer communications at a mid-sized bank," Anan "maintains the statements microservice." They are load-bearing rather than decorative — UJ-1 walks the exact FR set (FR2 page setup, FR7 binding, FR21–FR26 table, FR34 preview, FR8 save) and UJ-2 encodes NFR1 as a behaviour a developer performs ("compare the SHA-256 against the recorded hash. It matches the file Ploy saw in her browser"). A single-operator tool would not have earned two UJs; a two-sided handoff product does, and the PRD says why in §3.

Success criteria are correctly operational rather than adoption-based, and §7 says so up front: "Because Folio is presently built to scratch its author's own itch, success is defined by **capability and reliability, not adoption**." That is the right call for the stated stakes, and it is stated rather than left as an absence.

The mechanism/capability split between `prd.md` and `addendum.md` is well-observed — the body says `Page X of Y` is required and explicitly leaves the two-pass question to architecture (FR30), while the addendum carries the Go structs, module tree, and build phases. No over-formalization, no under-formalization.

No findings.

---

## Mechanical notes

- **Expression function count is inconsistent.** FR17 enumerates exactly eight functions (`sum`, `count`, `avg`, `formatDate`, `formatNumber`, `upper`, `lower`, `if`), but counter-metric C1 reads "Expression function count exceeding the **ten** specified." The counter-metric that guards against language creep is itself calibrated two functions loose. *(§5.3 FR17 vs. §7.3 C1 — worth fixing, since C1's whole job is to be countable.)*
- **Assumption Index roundtrip has one break.** Nine inline `[ASSUMPTION]` tags exist (§3.1, FR9, FR16, FR19, FR30, FR40, NFR4, NFR6, NFR7); §9 lists ten rows. The extra row — "§7 | Success is acceptance-based, not adoption-based, per current distribution intent" — has no corresponding inline tag in §7. Either tag it inline or drop the row.
- **ID continuity is clean.** FR1–FR42 contiguous, no duplicates, no gaps. NFR1–NFR8 likewise. S1–S6, C1–C5, Q1–Q9 all contiguous.
- **Cross-references resolve.** Q3→FR16, Q4→FR30, Q8→FR9, §10→FR42, NFR7→FR31/NFR1/NFR3, NFR5→NFR1, FR35→FR34 all check out. `docs/folio-mvp-plan.md` and `addendum.md` both exist at the referenced paths.
- **Zero `[NOTE FOR PM]` callouts.** The `[ASSUMPTION]` prose partly covers the same ground, but the two live tensions identified above (NFR1 fallback, C4 vs. Q6) sit in no callout of any kind.
- **UJ protagonists are named and self-contained.** Both UJ-1 and UJ-2 carry their protagonist's role inline; no floating UJs.
