---
title: Input Reconciliation — docs/folio-mvp-plan.md vs PRD + Addendum
input: docs/folio-mvp-plan.md (894 lines)
targets: prd.md, addendum.md
created: 2026-08-22
role: advocate for the source input
---

# Reconciliation — `docs/folio-mvp-plan.md`

Verdict in one line: the PRD is a **faithful capability inventory but an unfaithful
constraint inventory**. Nearly every *feature* in the plan survives; several of the plan's
*qualifiers, tiers, and hedges* do not, and in four places the PRD asserts requirements that
are strictly harder than — or in one case directly contradicted by — the source document,
without marking them as invented.

---

## Dropped entirely

### D1. Two of the seven template design goals — "CI/CD friendly" and "API friendly" (§3, Template Design Goals)

The plan lists seven co-equal goals: Human-readable, Git-friendly, Versionable, Diffable,
**CI/CD friendly**, **API friendly**, AI-editable.

PRD **FR11** carries: readable, diffable, mergeable-in-Git, human/AI-editable. **FR10/FR12**
carry versionable. Nothing in the PRD or addendum carries **API friendly** — the idea that a
`.folio` file is something you POST, store in a config service, or generate programmatically.
**CI/CD friendly** survives only obliquely, and only as validation (**FR41**), not as a
template-format property.

Addendum §C reproduces the template's JSON shape but omits the design-goals list entirely,
so neither document holds it.

### D2. Three output formats from the "do not attempt" list (§9)

Plan §9: *"Do not initially attempt to support: Excel, Word, HTML, CSV, PNG, PowerPoint."*

PRD §4.2 carries Excel, Word, HTML. **CSV, PNG, and PowerPoint are in neither the PRD nor
the addendum.** PNG is a live risk of omission: PRD §4.3 protects PNG as an
architecturally-possible later renderer, so without §9's explicit "not in MVP" the reader can
plausibly infer PNG output is fair game.

### D3. The §13 MVP Feature Matrix as a *priority instrument*

The plan's matrix encodes a five-tier priority scale: **Core** (Go library, above Must) >
**Must** (17 rows) > **Important** (streaming) > **Optional** (REST) > **Later** (Java/.NET/
Node SDKs, Excel export, charts, subreports).

The PRD has **no priority table at all**. Priority survives only as scattered prose:
"core MVP deliverable" (§5.8), "*(if capacity permits)*" (FR27), "*(optional…)*" (FR42),
"Priority: *important*" (NFR4). Consequences:

- **FR38 (streaming)** is stated with no priority marker at all in §5.8 — the "Important, not
  Must" tier lives only in NFR4, a different section. A reader of §5.8 alone will build it as
  a Must.
- The plan's distinction between **Later** (planned, deferred) and **out of scope** collapses:
  PRD §4.2 files Java/.NET/Node.js SDKs, Excel export, charts, and subreports under
  "explicitly out of scope", flattening a roadmap tier into a rejection. (§14 does postpone
  subreports/charts/Excel, so the tension is the plan's; the PRD resolves it in the harsher
  direction without saying so.)
- All 17 **Must** rows are now indistinguishable in weight from PRD-invented requirements
  (FR12, FR32, FR40, FR41).

### D4. Internationalization as an out-of-scope item (§14)

Plan §14 postpones **"Internationalization"** outright. The word appears nowhere in the PRD
§4.2 list; the PRD substitutes **"right-to-left text"** — a much narrower exclusion — and then
in §10 says "No i18n beyond script rendering."

This is not a wording change, it is a scope reversal. See W1.

### D5. Success criteria §17 items 2, 4, and 5 have no measure

The plan's §17 is a 12-item ordered acceptance list. PRD §7.2 compresses it to six measures.
Items 6–12 map onto S1/S2/S3/S5. Items **2 (the report can bind to JSON data)**,
**4 (a Go application can load that template)**, and **5 (the application can render it using
`folio.Render(template, data)`)** have no corresponding success measure — the entire
*integrator-side* half of the plan's acceptance definition is folded into S5's parenthetical
"render via `folio-go`". S1 ("4/4") is a single checkbox standing in for seven distinct plan
criteria (multi-page PDF, table pagination, repeated headers, headers/footers, page numbering).

### D6. Small concrete details

| Plan | Detail | Status |
|---|---|---|
| §7 | Parameter example `preparedBy: "Operations"` | absent both docs (`reportDate`, `branchName` kept) |
| §2 | Sample data field `transactions[].description`, negative amount `-1200` | absent both docs |
| §3 | Recommended filename `customer-statement.folio` | survives only inside PRD §3.1 journeys |
| §9 | *"Different output formats have significantly different layout semantics"* — stated as a **general** law | PRD §4.3 applies it to Excel only |
| §1 | *"Basic component properties should **include**…"* (open list, a floor) | FR5 reads as a closed enumeration |
| §1 | *"Support only the most important components **initially**"* | FR4 + C2 turn five into a hard ceiling |
| Goal | *"**modern** report designer"*; the five-bullet focus list (simple visual design, JSON-first binding, deterministic PDF, developer-friendly integration, native Go library) | dissolved into §1/§2 prose; "developer-friendly integration" as a named goal is gone |

---

## Weakened or distorted

### W1. NFR3 (Latin/CJK/Thai) directly contradicts §14

**Plan §14** postpones *Internationalization*. **PRD NFR3** makes correct measurement,
breaking, and rendering of **Latin, CJK, and Thai** — including **dictionary-based Thai line
breaking with an embedded, versioned dictionary** — a hard MVP requirement, adds **S4** as a
success measure, makes it load-bearing for **NFR7** (font provisioning) and **Q2/Q9**.

Nothing in the 894-line plan mentions Thai, CJK, scripts, shaping, or word-breaking
dictionaries. The plan's only text-layout requirements are "Correct text wrapping" (§16) and
"Text measurement / Line wrapping" (§10) — script-agnostic. This is the single largest
requirement in the PRD with **zero** source basis, and it is **not** in the §9 Assumption
Register. It plausibly reflects real (unstated) author intent, but as written the PRD makes
the plan say the opposite of what it says.

### W2. NFR1 escalates "same layout" to "byte-identical PDF"

**Plan §10** (verbatim): *"The same template and data should generate the same layout in every
environment."* Its illustration is **Developer Laptop vs Production Linux Container** →
"Identical Pagination / Identical Line Wrapping / Identical Table Breaks". §17 item 11:
*"Fonts and layouts are deterministic across environments."*

**PRD NFR1** requires a **byte-identical PDF, same file hash**, and enumerates PDF-internal
pinning (creation timestamp, document ID, object ordering, compression output, font subset
generation) that appears nowhere in the plan. It then declares this "the single hardest
constraint in the MVP", concludes "no cgo", "deterministic float behaviour", and "this may
rule out most off-the-shelf PDF libraries", and promotes it to **Q1, the pivotal feasibility
question**, plus **S2, S3, S5** and counter-metric **C3**.

The addendum is honest about this — §E: *"Stated at **layout** granularity ('same layout'),
never as byte-for-byte PDF equality."* The PRD body is not: it presents byte-reproducibility
as the source's requirement, unmarked and absent from the Assumption Register. Note the
substantive difference: layout determinism is testable by page/position comparison and is
compatible with most PDF libraries; hash determinism may force a hand-rolled PDF writer. The
plan never asked for that trade.

Corollary loss: the plan's own seven determinism **control points** (§10 — font handling,
font embedding, text measurement, line wrapping, page dimensions, pagination rules, **image
sizing**) do not appear in NFR1, which substitutes a different list. They survive only in
addendum §E. **Image sizing** in particular has no PRD requirement anywhere.

### W3. FR34/FR35/NFR8 contradict the plan's §11 recommended preview flow

**Plan §11** recommends, in a diagram: `Designer → (template + sample JSON) → **Preview API**
→ folio-go → PDF`. That is a service call. The plan's PDF Preview requirements are exactly
three: *exact production rendering*, *generated using folio-go*, *same semantics as production
output*.

**PRD FR34** mandates `folio-go` **compiled to WebAssembly and running in the browser tab** —
and explicitly rules out *"a remote service"*, i.e. rules out the plan's own recommendation.
**FR35** adds fully-offline operation; **NFR8** adds a privacy posture; **S2** adds
native-vs-WASM hash equality; **Q9** adds a WASM bundle-size question. WebAssembly, offline,
and privacy appear nowhere in the plan.

The addendum does record the tension (§G: the Preview API "appears in the §11 diagram but in
neither the §13 feature matrix nor the §15 phase list"; §J.2: "a server-side preview path is
unavoidable"). But addendum §J.2's conclusion and PRD FR34 point in opposite directions, and
the PRD body presents WASM as settled fact, unflagged, not in the Assumption Register.

### W4. The "draw.io model" is an invention presented as positioning

PRD §2 (positioning table), §5.1 preamble, **FR8** ("no server round-trip and no account"),
**FR35**, **NFR8** all rest on: a web app in the draw.io mould, local files only, no account,
no server-side template storage.

The plan says only **"Web visual designer | Must"** (§13) and *"Template save/load"* (§15
Phase 5). It specifies nothing about accounts, local-first storage, or server absence — and it
does describe a Preview API (§11) and an optional REST service (§12). PRD §2's row
*"Integration | An in-process Go library, **not a server you have to operate**"* also overstates
against §12, which explicitly permits a REST rendering service in MVP if time permits.

### W5. NFR2 forbids what the plan permits in the design canvas

**Plan §11** defines the Design Canvas purpose as *"Approximate visual representation"* — its
whole point is that it is fast and inexact, with PDF Preview as the exact counterpart.

**PRD NFR2** extends layout authority to `folio-go` *"including in the design canvas"* and says
"browser DOM-based rendering does not [satisfy this]". Under the plan, a DOM-drawn approximate
canvas is explicitly fine. NFR2 turns the plan's two-tier preview design into a one-tier
mandate and materially raises designer implementation cost.

### W6. Counter-metric C1 miscounts the expression catalogue

PRD **C1**: *"Expression function count exceeding **the ten specified**"*.

Plan §6 specifies **eight**: `sum`, `count`, `avg`, `formatDate`, `formatNumber`, `upper`,
`lower`, `if`. PRD **FR17** itself lists eight. The counter-metric's own threshold contradicts
the requirement it polices and grants two free functions of scope creep.

### W7. §14 "Report repository" narrowed to "server-side report repository"

Plan §14 postpones *"Report repository"* unqualified. PRD §4.2 writes *"server-side report
repository"*, and **FR42** then permits templates "resolved from a path supplied by the
operator". That is a real narrowing that makes FR42's phrasing self-consistent — but it is the
PRD resolving the plan's §12/§14 tension in favour of §12, silently. (Addendum §J.1 records the
tension correctly.)

### W8. Database independence hardened from MVP-scoped to permanent

Plan §2: *"Database connectivity **should not be part of the first MVP**"* — an MVP scoping
statement. PRD §2 and **FR13**: *"Folio **never** touches a database"* / *"Folio never connects
to a database"* — a product law. The plan's own §2 architecture diagram
(`Database → Application/Microservice → JSON → Folio`) supports the philosophy, but "not in
MVP" and "never" are different commitments.

### W9. Excel hedge promoted to mandate

Plan §5: *"Excel **should likely** use different layout semantics and **should not be forced**
into the same rendering model."* PRD §4.3: *"Excel is *not* protected in this way — it has
different layout semantics and **must not** be forced into the same model."* Addendum §A quotes
it correctly. Minor, but a hedge became a constraint.

### W10. "No migration path from JasperReports" over-generalizes §14

Plan §14 excludes exactly one thing: *"Jasper `.jrxml` compatibility"*. PRD §2 states
**"No migration path from JasperReports is offered in MVP"** — broader (excludes, e.g., a
documented manual conversion guide, which the plan never rules out).

### W11. FR32 (images embedded in the template) has no source

**FR32** forbids URL fetch and filesystem reads at render time and requires images to be
embedded in the `.folio` file. The plan says only "Image" (component, §1), "Images" (§16
acceptance) and "Image sizing" (§10 determinism). The rule is defensible and serves NFR1, but
it is an invented constraint with a real authoring cost (large base64 blobs inside a file the
plan wants "Human-readable", "Git-friendly", and "Diffable" — §3), and it is not flagged.

### W12. `folio.Render` signature vs FR39

Plan §4's API is `folio.Render(template, data)` and `folio.RenderTo(writer, template, data)` —
**two arguments, no parameters**, reproduced verbatim in addendum §B. Plan §7 requires runtime
parameters and §9's pipeline shows Template + Data + **Parameters** as three inputs. PRD
**FR39** requires parameters at render time. Nobody records that this breaks the plan's — and
§17 criterion 5's — literal signature, which is itself an acceptance criterion.

---

## Qualitative losses

### Q-L1. "Simple" is the plan's most repeated word, and it is nearly gone

The plan asserts simplicity as a design value seven times: *"Simple visual report design"*
(Goal), *"Support only the most important components initially"* (§1), *"Keep report structure
simple for the MVP"* (§1), *"The API should be **extremely simple**"* (§4), *"The MVP needs a
**small** expression language"* (§6), *"Avoid creating a general-purpose scripting language"*
(§6), *"optimize for reliability rather than feature breadth"* (Principle).

The PRD keeps two of these verbatim (FR37's "extremely simple", §5.3's closing line) and the
governing principle in §1. But the *reasons* are gone: "Keep report structure simple" becomes
FR6's neutral "Group headers, report headers, and group footers are deferred" — the deferral
survives, the stated motive does not. FR4's rationale ("cover the majority of common enterprise
reporting use cases") is preserved; FR6's is not.

### Q-L2. The plan's build-order conviction is quarantined into the addendum, then attacked

Plan §15 opens with an unhedged instruction: **"The Go rendering engine should be implemented
before investing heavily in the designer."** This is the plan's strongest sequencing statement
and it is the reason §15 exists.

In the PRD it appears nowhere in the body. It survives in addendum §H (quoted correctly) and is
then challenged from three directions — **Q5**, **Q6** ("Does the designer really come last…"),
and addendum §J.3/§J.4. Only **C4** carries it forward as a live rule, and only as a
counter-metric. A downstream reader of the PRD alone gets the impression the build order is
disputed; the plan states it as settled and gives it its own top-level section with six phases.

### Q-L3. Table rendering's rhetorical weight is preserved but relocated

Plan §8 opens: *"Table rendering is one of the most important technical areas in **the entire
MVP**."* §15 Phase 3: *"expected to be one of the **highest-risk** engineering areas."* These
are two different claims — importance and risk. PRD §5.5 merges them into one
("highest-risk area of the MVP"), losing the *importance* framing. The priority rule
(*"Correct pagination should take priority over advanced table styling"*) is correctly carried
in both docs — note the PRD drops the word "advanced", slightly broadening it.

### Q-L4. "AI-editable" survives as a claim but not as a peer

Asked directly: the plan gives **no** rationale for "AI-editable" — it is item 7 of 7 in a flat
bullet list of template design goals (§3), with equal weight to "Diffable" and "CI/CD friendly".
The PRD **promotes** it: it appears in the §2 positioning table as a headline property
("editable by an AI") and in **FR11**. So the *idea* is not lost — it is amplified, while two
of its six peers are deleted (D1). The plan's actual qualitative intent — *the template is a
plain text artifact you own and every tool in your pipeline can touch it* — is only partially
reconstructible from the PRD, because the pipeline half (API, CI/CD) is the half that went
missing.

### Q-L5. The §17 criteria lose their ordering and their voice

The plan's §17 is a **narrative walk-through of one workflow in order** — designer → bind →
save → Go loads → Go renders → PDF → tables → headers → footers → page numbers → determinism →
preview parity. It is the "Design → Bind → Preview → Render" thesis restated as acceptance. PRD
§7.2's S1–S6 is a **matrix of independent checks**, reordered around determinism (S2, S3, S5 are
all hash tests). The workflow-shaped acceptance narrative is gone, and with it the plan's
implicit claim that criterion 1 (the designer) comes first.

### Q-L6. §7's "scratch its author's own itch" framing has no source

PRD §7 opens: *"Because Folio is presently built to scratch its author's own itch, success is
defined by capability and reliability, not adoption."* The plan says nothing about the author,
the audience, or adoption. Its own framing is product-level and market-facing — "positioned as
an alternative to JasperReports", "cover the majority of common **enterprise** reporting use
cases", "one representative **enterprise** report". The PRD's personal framing (and Q7's
"Currently undecided by choice") is a plausible interview product, but it changes the register
of the document from "product" to "side project", which is not the plan's register.

### Q-L7. Diagrams as intent

The plan communicates heavily through ASCII diagrams (the MVP stack, the report-sections tree,
the data architecture chain, the table page-break illustration, the determinism laptop/container
figure, the renderer fan-out, the preview flow, the rendering pipeline). Addendum §A/§G preserve
five of them. Two are lost as figures: the **§1 report-sections tree** (Page Header / Content /
Page Footer, which conveys nesting and ordering that FR6's prose flattens) and the **§8 table
page-break illustration** (which shows *the header repeating immediately after the break*, the
precise behaviour FR25 states abstractly).

---

## Correctly carried (spot check)

| Plan | Where it lands | Note |
|---|---|---|
| §1 Canvas: drag-drop, page boundaries, A4/Letter/custom, portrait/landscape, margins, grid/snap, resize, positioning | FR1–FR3 | complete |
| §1 Components: exactly Text, Image, Table, Line, Rectangle + "majority of enterprise use cases" rationale | FR4 | complete, rationale kept |
| §1 Properties: all 12 (X/Y, W/H, font family, size, bold, italic, text align, vertical align, border, padding, background, visibility, data binding) | FR5 | complete |
| §1 Sections: Page Header / Content / Page Footer; group + report headers deferred | FR6 | complete |
| §2 JSON-first; app prepares data; `Database → App → JSON → Folio → PDF` | FR13, addendum §G | complete |
| §2/§7 Binding literals `{{customer.name}}`, `transactions[]`, `{{params.reportDate}}`, `{{sum(transactions.amount)}}`, `{{transaction.amount}}` | FR14–FR18, FR20, addendum §C | complete, incl. the unspecified row-scope flag |
| §3 `.folio`, text-based, JSON internally, `version`/`page`/`body` shape | FR10, addendum §C | complete |
| §4 folio-go is **Core**, not a later SDK; reference implementation; "extremely simple" API; streaming rationale ("avoids temporary files", HTTP handler) | §5.8, FR36–FR38, addendum §B | complete |
| §4 Package-name ambiguity (`folio` vs `folio-go`) | addendum §B | correctly flagged as unresolved |
| §5 Module tree, pipeline, `Page` struct, "Draw Directly Onto PDF" anti-pattern, PDF/PNG/SVG/HTML fan-out | addendum §A, PRD §4.3 | complete |
| §6 All 8 functions + pattern literals `"dd/MM/yyyy"`, `"#,##0.00"` + no-scripting rule | FR17, addendum §D | complete (but see W6) |
| §8 All 11 MVP table features + optional alternating rows | FR21–FR27 | complete, 1:1 |
| §8 Pagination-over-styling priority | §5.5, addendum §F | complete |
| §10 Seven determinism control points incl. image sizing | addendum §E | complete in addendum only |
| §10 "The browser should not be the canonical renderer" | NFR2 | carried and correctly strengthened in spirit |
| §11 Two preview modes; canvas = fast/approximate; "avoids discrepancies" | FR33, NFR5 | complete |
| §12 REST optional, Go library higher priority, `POST /api/reports/{name}/render`, `application/pdf`, future `/render/{pdf,html,xlsx}` | FR42, addendum §G | complete |
| §14 18 of 21 out-of-scope items + "do not reproduce Jasper feature-for-feature" | §2, §4.2 | see D4/W7/W10 for the three exceptions |
| §15 All six phases with full contents + Phase 1 goal + Phase 3 risk note | addendum §H | complete |
| §16 Golden report: logo, customer info, account info, statement period, 5-column table (Date/Description/Debit/Credit/Balance), footer (confidentiality, generated date, Page X of Y); 1/5/20/50 pages; all 11 reliability items | §7.1 | complete, verbatim-faithful — the strongest section of the PRD |
| Product Principle: reliability over feature breadth; 20–50 page enterprise statement from JSON via `.folio` + `folio-go` | §1 | complete, quoted |

---

## Recommended repairs, in order

1. **Tag NFR1, NFR3, FR34/FR35, and NFR8 as `[ASSUMPTION]` and add them to §9.** Four
   hard requirements currently read as source-derived when they are PRD inventions; one (NFR3)
   inverts an explicit §14 exclusion. This is the highest-value fix — downstream architecture is
   about to be sized against them (Q1, Q2, Q9).
2. **Reinstate a priority table.** Restore §13's Core / Must / Important / Optional / Later
   tiers as a column on the FR list, so FR38 is visibly *Important* and Java/.NET/Node/Excel/
   charts/subreports are visibly *Later*, not *rejected*.
3. **Fix C1**: ten → eight.
4. **Restore CSV, PNG, PowerPoint** to §4.2, and reconcile PNG against §4.3's protection.
5. **Restore "CI/CD friendly" and "API friendly"** to FR11, and reconcile "API friendly" +
   "human-readable / diffable" against FR32's embedded image blobs.
6. **Move §15's sequencing rule into the PRD body** as a stated constraint, rather than leaving
   it in the addendum under attack from Q5/Q6.
7. **Reconcile FR34 with addendum §J.2** — the PRD says "not a remote service", the addendum
   says a server-side preview path is unavoidable. One of them is wrong.
8. **Add success measures for §17 items 2, 4, 5** (JSON binding; Go app loads template; Go app
   renders via `folio.Render`) — currently the integrator half of acceptance is unmeasured.
9. **Resolve the `folio.Render` arity** (FR39 vs §17 criterion 5 / addendum §B).
