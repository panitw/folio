---
title: 'Story 16.1: A font arrives from the web with its terms attached'
type: 'feature'
created: '2026-09-02'
status: 'done'
review_loop_iteration: 0
followup_review_recommended: true
baseline_commit: '9cbff85084ed7d8ff3e98e66cbea14d4c45d844b'
baseline_revision: 'bb14662900f6056e6987874acd565d8e352592ee'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-fonts/font-catalogue.md'
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-16-decision-log.md'
warnings: ['oversized']
deferred:
  - summary: >-
      No browser witness exists for any web-tier pick or refusal; every
      fetch/classify/embed claim is asserted in jsdom against a stubbed fetcher.
    evidence: |-
      e2e/font-embed-boundary.spec.ts asserts only that web rows EXIST; its
      embed loop still covers the 21 local faces alone. The empirical claims the
      mechanism rests on — that raw.githubusercontent.com sends
      access-control-allow-origin: *, and that fonts.google.com/metadata/fonts
      does not — are measured in prose and by curl, never by anything the suite
      runs. Intent-authorised: D-16.R.1's cadence denies this story an in-story
      browser run and names the run as debt.
    location: >-
      folio-designer/e2e/font-embed-boundary.spec.ts
    severity: medium
  - summary: >-
      A fetched face's recorded `source` names the `main` branch, not a commit,
      so the provenance a .folio publishes is not reproducible.
    evidence: |-
      font-source.ts builds source as `${host}/google/fonts/main/${dir}/${slug}/${file}`.
      Under AD-8/D-16.2 a face is identified by the SHA-256 of its bytes, so two
      authors picking the same family a month apart embed different bytes and
      record an identical, unfalsifiable source string. Pinning the fetch to a
      commit SHA is what would make DW-158's "a newer release is a different
      face, not a newer one" argument hold for the web tier too.
    location: >-
      folio-designer/src/font-source.ts
    severity: low
  - summary: >-
      66 addable families declare neither `latin` nor `thai`, so they resolve to
      `scripts: []` and contribute nothing to a script this product supports.
    evidence: |-
      Measured over the committed font-index.json: 66 non-variable, non-CJK
      families carry only Devanagari, Arabic, Hebrew or Ethiopic subsets.
      scriptsOf emits only 'latin' and 'thai', so such a pick proposes all three
      shipped fallbacks as its tail and the embedded face draws nothing.
      font-catalogue.md's "Script coverage — Latin and Thai" criterion is
      applied to the local tier and not to the web tier; the intent contract
      excludes CJK by name but is silent on other scripts, so this is a product
      question rather than a deviation.
    location: >-
      folio-designer/scripts/build-font-index.mjs
    severity: low
  - summary: >-
      `popularity` is carried on every emitted index row and read by nothing, so
      the 50-row paint cap shows the alphabetical head of the library.
    evidence: |-
      offeredFamilies orders local-tier-first and otherwise preserves the
      upstream index's own alphabetical order; renderedFamilyLimit then slices
      50. An author opening the control sees ABeeZee, ADLaM Display, AR One
      Sans. D-16.R.2's argument for hiding variable-only families turns on which
      families a user reaches first, so the ordering deserves to be a stated
      decision with a test rather than an artefact of the upstream response.
    location: >-
      folio-designer/src/font-index.ts
    severity: low
  - summary: >-
      Offline-release cache headroom has narrowed to 20 of 64 slots, and no
      build script checks it.
    evidence: |-
      Measured this dispatch: s1.assetCount is 44 against maximumCacheAssets 64.
      It was 23 at Epic 8. This story adds no asset — the snapshot module is
      source and consumes no slot, verified — but the remaining margin is now 20
      rather than the 41 the earlier note recorded. parseS1Payload returns
      undefined past the ceiling and only a Vitest test and the runtime catch it.
    location: >-
      folio-designer/src/release-payload.ts:33
    severity: low
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The designer shipped with twenty-one typefaces. It now also reaches the library the web publishes —
about two thousand families — and downloads the one an author picks at the moment they pick it. The
twenty-one stay as the faces this machine already holds, and picking one of those downloads nothing.

The scope fence first, because it is a limit rather than a feature. **The list is not live.** Whoever
publishes it will not let a browser read it, so the list ships inside the designer as a dated snapshot
and ages between releases exactly as the old catalogue did. Only the typeface is fetched, and the
control says so rather than leaving anyone to find out.

The paperwork now happens on the author's machine, mid-click. Before any byte enters the document, the
terms stated for that family are checked against the short list this product accepts. Anything else is
refused and named — and named differently depending on why: terms we know and decline say so, while
terms never classified say they were not recognised, so something new upstream never reads as a decision
nobody made. Nothing is admitted by default.

About a quarter of the library ships as one file holding every weight at once, which this product does
not accept. Those families are not listed, and the count on the control is what can actually be added.

One thing will look wrong later and is not: **none of this has been witnessed in a real browser.** Every
claim here is proved against a stand-in for the network. That run is owed, and recorded as owed.

<intent-contract>

## Intent

**Problem:** `SPEC-fonts` CAP-3 is a **bundled catalogue of 21 families**, curated by a build-time
licence gate. The owner has ruled (**D-16.1**) that the shortlist is not the product: the author
reaches the published library. That reverses the `## Non-goals` clause *"No live font service. No
Google Fonts API, no arbitrary URL"* — the second reversal of that clause on 2026-09-02, after
D-8.4d.1 struck its third phrase — and it moves the licence admission that D-8.5.2/D-8.5.3 made a
**build gate** into the **runtime**, which is where D-8.6.5's defect (17 of 21 faces carrying another
project's licence) would now be reachable with nothing watching.

**Approach:** A **snapshotted index** and a **live face fetch**, because CORS leaves no other shape
(D-16.3). Snapshot `fonts.google.com/metadata/fonts` at build time into a trimmed typed module; fetch
bytes, `OFL.txt` and `METADATA.pb` from `raw.githubusercontent.com/google/fonts` on pick; apply the
four-identifier allowlist to the fetched licence **before any byte is embedded**; hand the existing
`embedFontFamily` command exactly what it already requires. Amend the forbidden-host scan rather than
deleting it (**D-16.4**).

## Boundaries & Constraints

**Always:**
- **The word "live" is qualified everywhere it appears, in UI strings and in prose.** D-16.3 measured
  that `fonts.google.com/metadata/fonts` returns **no `access-control-allow-origin`**; the browser
  cannot read it. The index is a build-time snapshot and the product says so.
- **Bytes come from `raw.githubusercontent.com/google/fonts`, never from `fonts.googleapis.com/css2`.**
  Measured 2026-09-02: `css2` under a modern browser UA returns **`woff2`**, which
  `decodeRecognisedFont` refuses by design (`font/ttf` and `font/otf` only, `font/woff2` deliberately
  outside the set), **split by `unicode-range` into per-script subsets**, which would embed partial
  coverage into a document naming the whole family. The TTF that endpoint serves to a legacy UA is
  unreachable: a browser cannot set `User-Agent`.
- **Which file is Regular is READ, not constructed.** Take it from that family's `METADATA.pb`
  `fonts { filename, style: "normal", weight: 400 }` entry. A filename assembled from the family name
  is a guess that will be wrong.
- **`licenceText` is the upstream file, never a hand-copy.** `font-catalogue.md`'s existing rule — a
  hand-copy "would be a second authority on the terms" — transfers verbatim: fetch that family's
  `OFL.txt` (or the licence file `METADATA.pb` names) and carry its bytes.
- **`copyright` comes from the face's own `name` table, nameID 0**, "the one statement of provenance
  that cannot be edited from outside the binary". It is not taken from `METADATA.pb`.
- **The licence vocabulary is a CLOSED TOKEN TABLE, and the upstream token is not an SPDX id**
  (**D-16.R.4** — the ruling that stops this story shipping broken). Measured: `METADATA.pb` carries
  `license: "OFL"` / `"UFL"` / `"APACHE2"`, **not** `OFL-1.1` / `Ubuntu-font-1.0` / `Apache-2.0`.
  Applied literally as this spec first worded it, every family would be refused except by the accident
  that `UFL` matches — a partial pass that looks like it works. A new module
  `folio-designer/src/font-licence.ts` owns the mapping and **names no host**, so D-16.4's declared-host
  module stays a small subject for the scan's second half.
  - `OFL` → `OFL-1.1`; `APACHE2` → `Apache-2.0`; `UFL` → `Ubuntu-font-1.0`;
    `CC-BY-SA` → **present in the table and refused**, with its reason stated.
  - **THREE STATES, NEVER TWO:** *mapped-and-admitted*; *mapped-and-refused* (a named token, a stated
    reason); *unmapped* (refused, and the message says **the token was not recognised**, not that it is
    forbidden). `cc-by-sa` is a real top-level directory in `google/fonts`, and it is in the table
    precisely so **absent** and **refused** stop looking the same. A fifth upstream directory must read
    as "we have not classified this", never as a policy decision nobody made.
  - **An unmapped token is a refusal and never a default.** No fall-through, no permissive default, no
    warn-and-continue (D-8.5.2).
  - **`font.licence` in the `.folio` carries the SPDX id, never the upstream token.** Precedent, not
    preference: `font-catalogue.json` already writes `Ubuntu-font-1.0` rather than `UFL`, and
    `licenceSignatures` is keyed on SPDX. Two vocabularies in one field make a document unsortable by
    its own terms.
  - **MIT stays admissible and gets no table entry**, and the module says why: D-8.5.3's four
    identifiers are owner policy and are not amended here; `google/fonts` publishes no MIT token, so
    MIT has nothing to map. **Absence, not narrowing.**
  - A test asserts the table's admitted value set is a **subset** of D-8.5.3's four.
- **The nameID 13 tie is STORY 16.1b's, and it lands BEFORE this story** (D-16.R.7, D-16.R.8). Do not
  implement it here and do not grow a browser-side substitute for it: the guard is one byte-taking door
  in `internal/fontset` called from `embedFontFamily`, because the browser is not the only door. **This
  story must not be the thing that puts a fetched face into a document before that gate exists** —
  building the gate after the population it polices arrives is how D-8.6.5 shipped green.
- **The nameID 0 READER lives in the browser, and that is forced rather than chosen.** `copyright` is
  one of `embedFontFamily`'s twelve wire fields under `componentFields(raw, 12)`, and this story may not
  change that arity — so Go cannot supply an input to itself. Go reads the name table **again**, from
  the bytes, for its own different question (16.1b). Two readers answering two questions is correct
  here; one reader would be a check over its own input.
- **A family that cannot be added is FILTERED OUT, not listed and refused** (**D-16.R.2**, owner). The
  refusal was priced as 28.7% of a long tail; measured, **37 of the 50 most popular families are
  variable-only** — Roboto, Open Sans, Inter, Montserrat, Raleway, Nunito, Oswald, Playfair Display.
  Listing them and refusing means the most common first action in the product fails. **A hidden row is
  a presentation choice, never a guard:** the engine's refusal stays, for anything that reaches it.
- **The browser's family count is the ADDABLE count**, and it says which it is. Not 1,946.
- **Only ONE class of refusal is filterable before a pick, and the spec says so** (D-16.R.7), so nobody
  builds a filter that cannot be fed: **(a) variable-only — YES**, from the snapshot's `axes`;
  **(b) licence-token refusals (`cc-by-sa`, unmapped) — NO**, the index carries no licence field so it
  is unknowable until `METADATA.pb` is fetched, i.e. after the pick; **(c) contradictions between a
  face's declared licence and its own bytes — NO**, by construction. The browser **filters (a)** and
  **states (b) and (c) at pick time with their reasons**.
- **The bundled catalogue SURVIVES as the LOCAL FACE TIER** (**D-16.R.3**). `pickCatalogueFamily`
  **gains a source; it does not swap one.** The 21 committed faces and their whole provenance regime —
  `font-catalogue.json`, `build-wasm.mjs`'s generated module, per-face `LICENSE*` and `NOTICE.md`, and
  `font-catalogue.test.ts` — are untouched. **Do not call it the "derived-static tier":** measured,
  `instance_faces.py` drives a hardcoded three-entry `UPSTREAM` list of ENGINE faces and **none of the
  21 is derived** (every `NOTICE.md`: *"NO DERIVATION APPLIES"*).
  - **Join key: exact `family` string equality.** No case-folding, no whitespace normalisation, no
    fuzzy match. `Inter Display` and `Source Serif 4 Display` have **no index row at all**, so 2 of 21
    are unjoinable under any normalisation, while `Geist` / `Geist Mono` / `Geist Pixel` is exactly the
    neighbourhood a loose matcher gets wrong. **A local face with no index row is local-tier-only, and
    that is correct behaviour, not a defect.**
  - **Local wins, with NO fetch at all** — no `METADATA.pb`, no `OFL.txt`. The committed bytes carry a
    **stronger** record than any fetch can produce; preferring a fetch would replace a verified record
    with an unverified one.
  - **Divergence is deliberately NOT reconciled.** Under AD-8 and D-16.2 a face is identified by the
    SHA-256 of its bytes, so "upstream released a newer version" is a **different face**, not a newer
    one. **No staleness check, no update prompt, no version compare in this epic** — register it in
    `deferred-work.md` with its trigger.
- **`Roboto` and `Inter` are already in the local tier**, as byte-for-byte upstream statics from
  `github.com/googlefonts/roboto-classic` and `github.com/rsms/inter` v4.1. They appear among the 37
  only because the **`google/fonts` mirror** carries VF-only builds of them. **"Variable-only" is a
  property of the byte source, not of the family** (D-16.R.2a) — so a family present in the local tier
  is offered from it and never judged by the index's `axes` field.
- **`axes` is a PREDICTION, and the spec says so where it would otherwise read as fact.** Measured:
  **all 558 axes-declaring families still list a `400` key** in `fonts`, so that map is an
  offered-weights list, not a static inventory, and `axes != []` is the only signal available. It is a
  good heuristic — verified on Roboto and six others — and it is a heuristic. The authority stays Go.
- **The family DIRECTORY is derived, then CONFIRMED — never trusted, and never read as a licence**
  (**D-16.R.6**). The index carries **no path and no licence field**, so the directory must be resolved.
  - **Slug rule, exact: lowercase the family name, then delete every character outside `[a-z0-9]`.**
    Verified 8 of 8 on deliberately awkward families (`Press Start 2P` → `pressstart2p`,
    `Baloo Bhai 2` → `baloobhai2`, `Source Serif 4` → `sourceserif4`, `Noto Sans Thai Looped` →
    `notosansthailooped`).
  - **Closed by verification**, which is why it is not the guess this story forbids one level down: the
    resolved directory is accepted **only if `METADATA.pb`'s `name` string-equals the index family the
    author picked.** A mismatch is a **refusal**, never a fallback to the next directory.
  - **`METADATA.pb` always wins on licence; the probe result is NEVER evidence of terms.** Probe order
    `ofl`, `apache`, `ufl`, `cc-by-sa`. Measured: **upstream moves families between directories —
    `apache/roboto` now 404s and Roboto lives in `ofl/`** — so reading layout as a licence assertion
    would let a family that moved silently change the terms a document publishes. That is AD-26's
    Prevents exactly.
  - **Layout disagreement is an observation, not a refusal.** Resolved at `ofl/x` while the token says
    `APACHE2` → `METADATA.pb` wins and `Apache-2.0` is admitted; the divergence is recorded.
  - **Probing is once per pick**, never at index render and never on a keystroke. The build step **may**
    carry the resolved directory in the snapshot as a **path** (a path is not a claim about terms),
    re-confirmed by the `name`-equality check at pick time; the snapshot **must not** carry the licence
    token, which would be a second licence authority ageing on its own schedule.
- **No licence is knowable before a pick**, and the UI must not pretend otherwise. It follows from the
  index having no licence field: every licence refusal is necessarily post-pick, after network
  round-trips. A `cc-by-sa` family cannot be pre-filtered, so the browser will show families it later
  refuses — say what that looks like rather than leaving it to the implementer.
- **Offline degrades, never breaks.** No network means no new family. It never means a document that
  will not render: the three shipped Noto faces are the coverage, and an embedded face travels inside
  the `.folio`.
- Every recorded measurement carries **command, commit, tree state and working directory** (D-8.4j.8).
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **A variable face would reach the embed command.** D-16.6 measured that the command currently accepts
  one and the renderer refuses it, i.e. a pick can write an unrenderable document. Story 16.0 closes
  that; if this story is somehow built first, it must not be the thing that makes the gap reachable.
- **Browser-side instancing would be introduced.** D-16.5 refused option (c) explicitly: it makes the
  embedded face a function of the author's runtime, which is AD-22's drift class at the asset layer.
  Any library that instances a variable font client-side is out of scope by ruling, not by omission.
- **The runtime licence check would be best-effort.** If the four-identifier allowlist cannot be
  applied to a fetched family with certainty, the family is refused, not admitted with a caveat.
- **A face would reach `Assets` before its licence and copyright are in hand.** The order is
  non-negotiable: classify, then embed.
- **Any of the 23 golden digests moves**, or `SupportedMajor` would move.
- **Story 16.0 is not closed.** Every acceptance here ends in an embed through the boundary 16.0 fixes.
  *(Closed 2026-09-03 at `4aa610a`.)*
- **The nameID 13 guard would ship without its falsifier.** The build-time tie is measured over **21**
  faces, not 1,946, so there is no false-positive population for the open library. This story must
  sample **≥50 families across `ofl`/`apache`/`ufl`** and report the rate at which the guard would
  refuse them **before** it ships. A rate materially above zero means the signature table is too narrow
  and comes back to the engineering lead as a finding — **never fixed by softening the guard to a
  warning.**
- **The nameID 13 narrowing would slip past the tag.** It narrows `embedFontFamily`, which is reachable
  through the exported API, so under D-7.8.3/D-8.2.2 it joins the **before-the-tag set**.
  `folio-go/v0.1.0` is verified not cut. **It lands in this epic or it does not land**, and it goes on
  D-000.15's running list of what the format freedom was spent on, for Story 15.3.

**Never:** host fonts (the one Non-goal clause D-16.1 leaves standing) · `woff2` · a `unicode-range`
subset · an API key in the client bundle (D-16.3's refused alternative) · CJK families ·
save-time subsetting · bold, italic or variable axes.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Pick a static family | `Kanit`, index snapshot present, network up | `METADATA.pb` → `Kanit-Regular.ttf` + `OFL.txt` fetched; licence classified; one embed command | No error |
| Pick a family already in the machine store | Store holds the key | **No fetch**; embed from stored bytes (Story 16.2) | No error |
| Pick a family already in this document | Asset key present | Dedupe by content hash; no second asset, no second history entry | No error |
| VF-only family, not in the local tier | `Anuphan[wght].ttf` is the only file upstream | **Filtered out of results** (D-16.R.2). Not listed, not refused in the UI | The engine's refusal remains for anything that reaches it |
| VF upstream, present in the local tier | `Roboto`, `Inter`, `Noto Sans Thai`, `Noto Serif Thai` | **Offered normally, from the local tier, with no fetch** — `axes` is not consulted for a family the local tier holds | No error |
| Local face with no index row | `Inter Display`, `Source Serif 4 Display` | Local-tier-only; offered from the family control, absent from the web browser's results | Correct behaviour, not a defect |
| Token maps to a refused licence | `METADATA.pb` says `CC-BY-SA` | **Refused, named token, stated reason** — post-pick, because no licence is knowable before one | Refusal at the control |
| Token not in the table | a fifth upstream directory appears | **Refused**, and the message says the token was **not recognised** — never that it is forbidden | Refusal, distinguishable from a policy decision |
| Face contradicts its declared licence | fetched face whose name table names a different licence | **The engine refuses it** — Story 16.1b's guard, shipped at `f6953da`. This story surfaces that refusal at the control | Located refusal, necessarily post-pick |
| Face states nothing readable about its licence | no signature matched, or the name table absent | **Admitted** (D-16.R.7: NO EVIDENCE admits). This story must not add a browser-side check that disagrees | No error |
| `METADATA.pb` name mismatch | slug resolves a directory whose `name` is not the picked family | **Refused** — never a fallback to the next directory | Refusal |
| Directory layout disagrees with the token | resolved at `ofl/x`, token says `APACHE2` | `METADATA.pb` wins; `Apache-2.0` admitted; divergence **recorded** | Not a refusal |
| Unclassifiable licence | `METADATA.pb` names something outside the four | **Refused**, named, before any byte is embedded | Refusal at the control |
| `OFL.txt` missing or empty | Licence file absent upstream | **Refused** — the document may not carry a face without its terms | Refusal, stating why |
| `name` table has no nameID 0 | Face carries no copyright | **Refused** — `copyright` is required of an embedded face | Refusal, stating why |
| Network down | Fetch rejects | *"You cannot add a family right now"*, stated; machine store still offered | Degradation, not an error state |
| Upstream 404 | Family in a stale snapshot, gone upstream | Named refusal that says the snapshot is stale | Refusal, actionable |
| Face over the engine's cap | > 6,288,384 bytes | The engine's existing located refusal | Refusal, anchored |

</intent-contract>

## Code Map

**Every anchor below was re-measured at `9cbff85` (2026-09-03, clean tree, wd `/Users/panitw/Projects/folio`).**
The previous gate measured at `384c8ac`; Story 16.1b (`f6953da`) has landed since and moved most of the
Go anchors. Anchors quoted **inside `<intent-contract>`** cannot be edited — the stale ones are listed
at the end of this section, and **where this section disagrees with the contract's prose, this section
is correct on addresses** (never on rules).

**Measured endpoints (2026-09-02, re-confirmed 2026-09-03; `curl` with `Origin: http://localhost:5173`)**
- `https://fonts.google.com/metadata/fonts` — 200, 2,699,611 bytes, **1,946 families**, **34 Thai**,
  **no ACAO**. Usable fields: `family`, `category`, `subsets`, `designers`, `fonts` (style keys),
  `axes`, `popularity`, `trending`, `size`, `isOpenSource`. **No path field and no licence field.**
- `https://fonts.googleapis.com/metadata/fonts` — **404**. Not the endpoint.
- `https://raw.githubusercontent.com/google/fonts/main/ofl/<dir>/<Family>-Regular.ttf` — 200, **ACAO
  `*`**, full static TTF (`Sarabun-Regular.ttf` = 90,220 bytes).
- `.../ofl/<dir>/OFL.txt` — 200, ACAO `*`, 4,387 bytes for Sarabun.
- `.../ofl/<dir>/METADATA.pb` — 200, ACAO `*`, carries `designer`, `license`, `category`,
  `fonts { name, style, weight, filename }`.

### Already built — DO NOT re-implement (Story 16.1b, `f6953da`)

- `folio-go/internal/fontset/licencesignature.go` (359 lines). Exported surface is **exactly two
  functions, no exported types**:
  - `ReadLicenceStatement(data []byte) (string, bool)` — **`:233`** (doc `177-232`). Reads nameID 13;
    consults nameID 0 **only when 13 is absent** (`:200` records why).
  - `RefuseContradictedLicence(name, declared string, data []byte) error` — **`:298`** (doc `271-297`).
  - `admitLicenceSignatures` — an **ordered slice** at **`:95`**, rows `96-98`: `OFL-1.1` →
    `(?i)SIL Open Font License`; `Ubuntu-font-1.0` → `(?i)Ubuntu Font Licence`; `Apache-2.0` →
    `(?i)Apache License,?\s+Version 2\.0`. **`MIT` deliberately has no row** (D-16.R.10: absence is
    NO EVIDENCE and admits).
  - `refuseLicenceSignatures` — ordered slice at **`:113`**, rows `118-121`: GPL/LGPL/AGPL (two
    patterns), SSPL, ShareAlike/CC-BY-SA. Checked against **every** face regardless of what it declares.
  - Three outcomes, and **NO EVIDENCE ADMITS** (`:301`, `:334`). Refuse half runs first (`307-317`).
- `folio-go/component_commands.go:2443` — the call site:
  `fontset.RefuseContradictedLicence(name, record.Licence.Value, decoded)`, wrapped at `:2444` as
  `componentFailure("", fontChainPath(name), lerr.Error())`. It sits **after** `RefuseVariableFace`
  (`:2414`) and **before** the dedupe (`:2452`) and the first `t.doc.Assets` touch (`:2465`).
  **This is what consumes this story's SPDX id**: the browser's `licence` wire field is `record.Licence.Value`.
- `folio-go/internal/fontset/licencesignature_test.go:691` —
  `TestGoLicenceTableSubsumesTheDesignerTable`, **the mirror contract**. It reads
  `folio-designer/src/font-catalogue.test.ts` **as source text** with a non-greedy regex over
  `const licenceSignatures`, carries a vacuity guard, and **hardcodes the ids the TS table is known to
  declare: `{OFL-1.1, Ubuntu-font-1.0}`**. **It does not read this story's new module.** See the task
  that extends it.

### Go — read-only for this story

- `component_commands.go` — `embedFontFamily` **`2360-2495`** (doc from `2331`). `componentFields(raw, 12)`
  `:2361`; `embeddedFontRecord` `:2368`; `embeddedFaceBytes` `:2376`; `DecodeFontForRender` `:2392`;
  `assetKeyReferenced` `:2452`; `maxCanvasFontFamilies` `:2458`. **Unchanged by this story.**
- `component_commands.go:668` — `maxComponentAssetBytes`, *computed*:
  `(8388608 − 4096) × 3/4` = **6,288,384**. Enforced in `embeddedFaceBytes` at **`:2565`** (not `:743`,
  which is the image path).
- `internal/template/fontasset.go` — `decodeRecognisedFont` **`:178`**, `DecodeFontForRender` **`:197`**.
  Accepted media types are **exactly `font/ttf`, `font/otf`** (`:180`). This is why `woff2` is not an option.
- `internal/fontset/variableface.go:69` — `RefuseVariableFace(name string, data []byte) error`. Still
  **returns `nil` for an unparsable face**, deliberately (`70-73`); DW-150 is **CLOSED as reconciled**
  under D-16.R.7, so this story does **not** register it.
- `internal/fontset/fontset.go` — package doc `1-16` states `*ot.Font` may not cross the package seam
  (D-1.5.10 / AC17a). `readPostScriptName` `:502`; `PostScriptName()` `:465`.
- `internal/template/parse.go` — `requireEmbeddedFaceLicence` **`:471`** (doc `435-470`), called from
  `decodeFontChainEntry` (def `:384`) at **`:424`**. **Measured: there is no comment anywhere in Go
  source about a one-door asymmetry, a deferred second door, or D-16.R.11** — that text exists only in
  the `9cbff85` commit message. This story writes it.
- `folio-go/byte_neutrality_test.go:100` — `goldenDigestRecord`, **exactly 23 records** (counted).
  `internal/template/version.go:77` — `SupportedMajor = 2`.
- `folio-go/render_arch_test.go:461` — `TestFolioMethodNamesAreInjective`. **A new error *type* may not
  be declared in the module-root package `folio`.** Not reached by this story unless Go is touched.
- `page_setup.go:513` — `maxCanvasFontFamilies = 256`; enforced `page_setup.go:609` and
  `component_commands.go:2458`.

### Designer (`folio-designer/`) — where this story works

- `src/App.tsx:608-627` — `pickCatalogueFamily`. **The seam this story widens.** `fetch(face.url)`
  `:612`; fallback tail from `scriptFallbackFaces` `:625`; `embedFontFamilyCommand({...})` `:626`, where
  **`mediaType: 'font/ttf'` is hardcoded at the call site**, not read from the face. `onPickFamily` is
  wired at `:903`; the prop hops are `:1112` → `:1145` → `FontFamilyProperty` `:1296` → invoked `:1335`.
- `src/font-chain-command.ts` — `embedFontFamilyCommand` **`69-87`** (not `69-98`; `89-99` is a separate
  `base64` helper). The *"THE FIELD COUNT IS PART OF THE CONTRACT"* comment is at **`15-18`**.
  **Wire keys in emission order (`81-86`):** `kind`, `version`, `name`, `family`, `style`, `licence`,
  `licenceText`, `copyright`, `source`, `mediaType`, `data`, `tail`. ⚠ **Two TS parameter names differ
  from their wire keys — `chain` → `name`, `bytes` → `data`** — so an implementer counting the TS
  signature will miscount the twelve. **Do not change the arity.**
- ⚠ **A nameID-0 reader ALREADY EXISTS IN TYPESCRIPT, TWICE, and neither copy is importable
  production source.** `src/font-catalogue.test.ts:95` `nameTableString(view, tables, nameID)` (table
  directory at `:81`, `DataView` at `:76`) and `scripts/build-wasm.mjs:247` `nameTableString` (directory
  `:236`, used for copyright at `:268-269`). Both are hand-written `DataView` walks, and
  `font-catalogue.test.ts:71` records **why there is no parser dependency** — adding one is a new
  dependency in a bundle under a measured payload budget. This story needs the reader **at runtime**, so
  it is the first production copy; **extract one module and have the existing two use it rather than
  writing a third hand-copy.** `package.json` has no font-parsing dependency (`dependencies` are exactly
  `pdfjs-dist`, `react`, `react-dom`) and **this story adds none**.
- `src/font-catalogue.test.ts:197-200` — the **build-time** signature table, exactly two rows
  (`'OFL-1.1': /SIL Open Font License/i`, `'Ubuntu-font-1.0': /Ubuntu Font Licence/i`). The nameID-13
  tie assertions are **`355-367`** inside a test spanning `317-405`; the nameID-0 assertions are
  `397-402`. The catalogue population floor is **`:303`, `toBeGreaterThanOrEqual(20)`** — there is **no
  literal "21-face floor"**; "21" appears only in prose.
- `folio-designer/font-catalogue.json` — **21 entries**, `licence` tally **`OFL-1.1` × 19,
  `Ubuntu-font-1.0` × 2**. The field is spelled `licence` (British); there is no `license` key. The whole
  local tier sits inside the two-row build-time table, which is why that tie is green and says nothing
  about the wider population.
- `scripts/build-wasm.mjs` — emits `src/generated/font-catalogue.ts` `:308`,
  `src/generated/offline-assets.ts` `:323`, `src/generated/pdfjs-assets.ts` `:325`,
  `src/generated/runtime-fonts.css` `:362`, and `src/generated/runtime/`. ⚠ **`font-catalogue.ts` and
  `runtime-fonts.css` are GITIGNORED** (`.gitignore:66-71`); only `pdfjs-assets.ts` is tracked under
  `src/generated/`. Consequently `npm run typecheck` and `npm test` both prepend `build:wasm`.
  **The index snapshot is generated here, beside them** — and whether it is tracked or generated is a
  choice this story must make explicitly, not inherit.
- `scripts/forbidden-font-hosts.mjs` — `FORBIDDEN_FONT_HOSTS` **`41-44`**, exactly two entries
  (`fonts.googleapis.com`, `fonts.gstatic.com`); `DECLARATION_MARKER` **`:47`**
  (`folio:font-host-declaration`); `SCANNED_ROOTS` **`:74`**; `POPULATION_FLOOR` **`:233` = 400**.
  Marker direction, precisely: `occurrencesIn` (`171-182`) splits the **RAW** source (`:172`) while
  `exemptLineNumbers` (`157-165`) computes exemptions over **comment-blanked** source (`:158`), so a
  marker written in a comment declares nothing (`95-101` says so). ⚠ **Neither
  `raw.githubusercontent.com` nor `fonts.google.com` is a forbidden host today**, so this story's new
  hosts do not trip the scan and it stays green **unamended**. The D-16.4 amendment is cut to a
  successor story — see the Spec Change Log.
- `src/forbidden-font-hosts.test.ts` — **nine** `it(` cases (not three, not eight). The
  comment-direction test is at **`:164`**.
- `src/release-payload.ts:32-33` — `minimumCacheAssets = 10`, `maximumCacheAssets = 64`.
  `scripts/offline-release-contract.mjs:66-70` reads them with a **line-anchored regex**
  (`^const <name> = (\d+)$`, `gm`) that must match **exactly once**, so reformatting either line breaks
  the release build. **The snapshot module is source, not a cache asset — assert it consumes no slot.**
- `src/canvas-authority-contract.test.ts:186-194` — the standing red (**DW-152**). The corpus is a
  directory walk of `src/**` and `e2e/**` with **no registration list**, so any file this story adds is
  auto-enrolled and must not use `getComputedStyle`, `measureText` or `getBoundingClientRect`.
- `src/engine.worker.ts` — Story 16.0 replaced the bare `catch` with a five-value `BoundaryStage`
  machine plus `describeThrow`, so an embed failure now names which boundary stage threw. **This is the
  channel every refusal in this story surfaces through** — read it before writing a refusal message.
- **Absent at HEAD — all five files this story creates:** `src/font-licence.ts`, `src/font-source.ts`,
  `src/font-index.ts`, `src/font-name-table.ts`, `scripts/build-font-index.mjs`, plus the generated
  snapshot module. Nearest neighbours for style: `src/embedded-face-family.ts`,
  `src/embedded-face-registry.ts`, `src/shipped-face-family.ts` (all exist). **The names above are the
  spec's; a better neighbourhood name is a review finding, not an implementer's free choice** — the
  mirror-contract test reads `src/font-licence.ts` **by path**.

### Anchors quoted inside `<intent-contract>` that are now stale

The contract's addresses were written at `a40c34d`. These four are wrong and are corrected above; the
**rules** they carry are unaffected.

| In the contract | Correct at `9cbff85` |
|---|---|
| `src/font-chain-command.ts:69-98` | `69-87` |
| `component_commands.go:2359-2410` | `2360-2495`, doc from `2331` |
| `internal/template/fontasset.go:178-206` … "the face-size cap" | `decodeRecognisedFont` `:178`, `DecodeFontForRender` `:197`; the cap is **`component_commands.go:668`**, enforced `:2565` |
| `src/release-payload.ts:33` read "by line anchor" | read by a `gm` regex that must match once — the effect (reformatting breaks the build) is right, the mechanism is not |

**And two I/O-matrix rows in the contract are superseded by D-16.R.7 and by shipped code.** The rows
*"nameID 13 disagrees with the token → Refused"* and *"nameID 13 absent or unparseable → **Refused,
saying which of the two**"* were written under D-16.R.5. D-16.R.7 replaced the second: **absent or
unparseable is NO EVIDENCE and ADMITS**, and `licencesignature.go:301,334` implements it that way. Both
rows describe **Story 16.1b's** door, which is built and closed. **Neither is this story's to implement
and neither may be used as an acceptance here.** The contract is preserved verbatim by rule; this note
is the correction.

## Tasks & Acceptance

**Execution:**

*Settled and not to be re-opened:* D-16.5 (owner) · D-16.R.2 / D-16.R.2a (filter, not refuse; the
derive-ahead batch is Story 16.1a's, not this story's) · D-16.R.3 (local face tier) · D-16.R.4 (closed
token table) · D-16.R.6 (slug-and-confirm) · D-16.R.7 / D-16.R.8 (the Go guard is 16.1b's and is
**done**) · D-16.R.10 · D-16.R.11 (the second door is deferred; this story writes its comment).

- `folio-designer/scripts/build-font-index.mjs` (**new**) — a build step that snapshots `fonts.google.com/metadata/fonts`, trims it
  to the rendered fields (`family`, `category`, `subsets`, `axes`, `popularity`, `fonts` style keys),
  and emits a typed module beside the existing generated catalogue, echoing `CatalogueFace`'s shape so
  `App.tsx` reads one kind of thing. **Record the snapshot's date and family count in the module** so
  the UI can state its own staleness. **State in the task whether the module is committed or
  gitignored-and-generated, and why** — `font-catalogue.ts` is gitignored, `pdfjs-assets.ts` is tracked,
  so the precedent does not decide it. The build step **must not require network at `npm run build`
  time**: an offline release build is a shipped gate.
- `folio-designer/src/font-licence.ts` (**new**) — the closed token→SPDX table with its **three states**
  (mapped-and-admitted; mapped-and-refused, named, with a reason; unmapped, refused, saying the token
  was **not recognised**). `OFL`→`OFL-1.1`, `APACHE2`→`Apache-2.0`, `UFL`→`Ubuntu-font-1.0`,
  `CC-BY-SA`→present and refused. No fall-through, no permissive default. **The module names no host.**
  Plus the test asserting its admitted set is a **subset of D-8.5.3's four**.
- `folio-go/internal/fontset/licencesignature_test.go` — **extend
  `TestGoLicenceTableSubsumesTheDesignerTable` to also read `src/font-licence.ts`'s admitted SPDX set.**
  Rationale, and this is the one Go change this story owns beyond a comment: the mirror contract today
  reads only `font-catalogue.test.ts`'s **build-time** table and hardcodes `{OFL-1.1,
  Ubuntu-font-1.0}`. This story creates a **second** TS table that decides what SPDX id reaches
  `RefuseContradictedLicence`, and nothing watches it. D-16.R.10 names exactly this gap and nominates
  this test as the thing that should catch it. Carry over the existing **vacuity guard** and the
  **known-ids list** idiom so a rotted extraction reds instead of passing over an empty read.
- `folio-designer/src/font-name-table.ts` (**new**) — a **shared nameID-0/name-table reader** extracted into one production module,
  with `src/font-catalogue.test.ts` and `scripts/build-wasm.mjs` switched onto it. `copyright` for a
  fetched face comes from that reader over the fetched bytes. **No new dependency.** Do not write a
  third hand-copy of the `DataView` walk.
- `folio-designer/src/font-source.ts` (**new**) — one **font-source module**, the single declared home of every allowed host.
  It resolves a family to its `METADATA.pb`, its Regular filename (**read from
  `fonts { style: "normal", weight: 400 }`, never constructed**), its face bytes, and its licence text
  (the upstream file, never a hand-copy). It carries the `folio:font-host-declaration` marker **as real
  code on the line naming the host**, so the successor host-scan story has a subject.
- `folio-designer/src/font-source.ts` — the **directory slug-and-confirm rule** (D-16.R.6): lowercase the family name
  and delete every character outside `[a-z0-9]`; probe `ofl`, `apache`, `ufl`, `cc-by-sa`; accept only
  if `METADATA.pb`'s `name` **string-equals** the picked family; a mismatch is a refusal, never a
  fallback to the next directory. Probe **once per pick**. A test that the probe result never sets
  `font.licence`.
- `folio-designer/src/font-index.ts` (**new**) — the **local-tier join** (D-16.R.3): **exact `family` string equality** against
  `font-catalogue.json`; local wins with **no fetch at all**; `axes` is not consulted for a family the
  local tier holds; a local face with no index row stays offered from the local tier. A test that a
  local-tier pick issues **no** third-party request.
- `folio-designer/src/font-index.ts` — **filter out** rows the snapshot's `axes` marks variable-only **and** which
  the local tier does not hold. The rendered count is the **addable** count and says so.
- `folio-designer/src/App.tsx` — widen `pickCatalogueFamily` (`608-627`) to take a family from the
  snapshot as well as a bundled `CatalogueFace`, keeping the fallback-tail proposal and
  `embedFontFamilyCommand` **untouched at twelve fields**.
- `folio-go/internal/template/parse.go` — **a comment naming the one-door asymmetry deliberately**
  (D-16.R.11's guardrail; measured absent at `9cbff85`). It must say the second door is **deferred past
  `folio-go/v0.1.0` with no migration owed**, and name the trigger: **adding a refuse-signature after
  the tag is itself the narrowing**. Without it the next reader closes the gap as a tidy-up and ships an
  unpriced narrowing.
- `folio-designer/scripts/forbidden-font-hosts.mjs` — **grow the scan's SECOND HALF** (D-16.4).
  `raw.githubusercontent.com` and `fonts.google.com` are permitted **only** on a line carrying both the
  host and `folio:font-host-declaration` **as real code**; anywhere else under `SCANNED_ROOTS` fails the
  build. The first half is untouched — `fonts.googleapis.com` and `fonts.gstatic.com` stay forbidden
  outright, because D-16.3 measured them unusable and forbidding them keeps the `woff2`/subset trap shut.
- `folio-designer/src/forbidden-font-hosts.test.ts` — extend the **positive control** and the
  **population floor** to the new half, preserving the marker direction: the scan runs over **raw**
  source while the exemption runs over **comment-blanked** source, so a marker in a comment declares
  nothing. **The new half must red when `src/font-source.ts`'s declaration is removed** — red-prove by
  deleting the half, never by falsifying a condition.
- `_bmad-output/implementation-artifacts/deferred-work.md` — register (a) the local-tier **divergence**
  (no staleness check, no update prompt, no version compare in this epic) with its trigger, and (b) what
  moving the licence gate to runtime leaves unwatched at build time. **Do not register DW-150** — it is
  already CLOSED as reconciled under D-16.R.7.
- Docs: `_bmad-output/specs/spec-fonts/SPEC.md` `## Non-goals` — strike the *"No live font service"*
  clause in place in the `~~…~~` + **AMENDED** form D-8.4d.1 used, preserving the original wording, and
  amend D-8.4d.1's own surviving sentence which this contradicts. `font-catalogue.md`
  `## Why bundled rather than fetched` — now false about the bytes, **still true about the list**; say
  exactly that. `spec-folio/folio-format.md` font-asset section — the record is unchanged, its **source**
  is not.
- Tests for the I/O matrix rows this story owns: a family resolving end to end from snapshot row to
  command payload; an unclassifiable licence refused; a `CC-BY-SA` token refused with its reason; an
  unmapped token refused as **not recognised**; a missing `OFL.txt` refused; a missing nameID 0 refused;
  a `METADATA.pb` `name` mismatch refused; offline degrading to the stated message; and **the `woff2`
  route asserted absent from source**.

**Acceptance Criteria:**

- Given the family index, when the designer is built, then a trimmed snapshot ships with it carrying its
  own date and family count, and the browser **never fetches the index at runtime**.
- Given a picked family, when its bytes are fetched, then they are a **full static TTF** from the
  declared host, its filename read from `METADATA.pb`'s `style:"normal" weight:400` entry, and never a
  `woff2` and never a `unicode-range` subset.
- Given any face that reaches a document, when it is embedded, then `licenceText` is the upstream
  licence file fetched with it, `copyright` is the face's own nameID 0, and `licence` is one of the four
  admitted SPDX identifiers — or the pick is **refused before any byte is written**.
- Given a fetched family, when its licence token is classified, then an admitted token maps to an SPDX
  id the `.folio` carries, a refused token is named with its reason, and an unrecognised token says it
  was **not recognised** rather than that it is forbidden.
- Given a licence outside the allowlist, when it is classified, then the pick is refused and named,
  **never warned about**.
- Given `src/font-licence.ts`'s admitted SPDX set, when the Go suite runs, then the mirror-contract test
  reads it and **reds if it admits an id `admitLicenceSignatures` has no row for** — red-proved by
  adding such a row to the TS module, and non-vacuous because the extraction's vacuity guard reds on an
  empty read.
- Given a family present in the local face tier, when it is picked, then it is embedded from the
  committed bytes with **no fetch at all**, and the index's `axes` field is not consulted for it.
- Given a local face with no index row, when the family control is opened, then it is still offered from
  the local tier and does not appear among the web browser's results.
- Given a family that cannot be added, when results are rendered, then it is **filtered out** rather than
  listed and refused, and the browser's count is the **addable** count and says so.
- Given a family directory, when it is resolved, then the slug rule produced it and `METADATA.pb`'s own
  `name` confirmed it, and the directory the probe found is **never** used as evidence of the licence.
- Given the forbidden-host scan, when an allowed host appears in source outside the one module that
  declares it, then the build fails; and when that declaring module is removed, the scan's new half
  reds.
- Given no network, when the author opens the browser, then it states that a family cannot be added
  right now and offers what the machine already holds — a degradation, never an error state.
- Given `internal/template/parse.go`, when it is read, then a comment names the one-door asymmetry, says
  the second door is deferred past the tag with no migration owed, and names the refuse-signature
  trigger.
- Given the whole story, when the suites run, then the **23** golden digests are unmoved, `SupportedMajor`
  is still 2, `maximumCacheAssets` is still 64 and the snapshot module consumes **no cache slot**.

## Design Notes

**Why the index is snapshotted, and that is not a compromise being hidden.** The alternative that would
make it genuinely live is the Developer API, and D-16.3 refused it on arithmetic: it needs an API key in
the client bundle — a secret nobody can keep, on someone else's quota — and it returns no licence text,
so `raw.githubusercontent.com` would still be required. It buys one property and costs two.

**What this story is now, after two cuts.** It is the **fetch path and its admission decision**, and
nothing else. The byte-side guard is built (16.1b, `f6953da`); the host-scan amendment is a successor
story. What remains is one goal: *a family the author picks becomes a face in the document, with terms
this product accepts, or it does not become one at all.* The licence-token mapping stays here because it
**is** that admission decision — D-16.R.8 forbids cutting it, for the reason D-16.6 refused an earlier
split: the fetch and its gate in two stories is how a population arrives before the thing that polices it.

**Where the two licence checks meet, and why that is not duplication.** The browser decides *whether
these terms are admissible at all* from what the library says (`METADATA.pb`'s token). Go decides
*whether the bytes agree with the terms written beside them* (`RefuseContradictedLicence`). Two
questions, two authorities, and the browser's answer is literally Go's input — `record.Licence.Value` at
`component_commands.go:2443`. That is why the mirror-contract extension matters: an SPDX id this story
can emit that Go has no signature row for lands in **NO EVIDENCE and admits**, with no tie at all.
D-16.R.10 accepted that as a bounded gap **for today's closed token set**; the moment this module can
emit a fourth id, the gap is no longer bounded by anything a test reads.

**Why `embedFontFamily` is untouched.** It already demands exactly what a `.folio` requires and refuses
without it, so the writer still cannot produce a document its own parser rejects. Changing the *source*
of those values while leaving that guard in place is what keeps this story from reaching the format.

**Why the hidden families stay hidden.** D-16.5(c) — instancing in the browser — would buy them by
making the embedded face a function of the author's browser and library version: a different runtime,
different bytes, a different PDF from the same template. Folio has no backend, so there is no third place
to do it. Do not "fix" the filter by reaching for a client-side instancer; it is refused by ruling.

**The nameID-0 reader is the quiet cost in this story.** It exists twice already and neither copy can be
imported — one is a test, one is a build script — and the deliberate absence of a font-parsing dependency
(`font-catalogue.test.ts:71`) means the third consumer cannot simply `npm install` its way out. Extract
once, switch both, add nothing.

**Why the second seam was declined (orchestrator, at the re-plan gate).** D-16.R.8 pre-authorised
cutting the host-scan amendment *"if 16.1 is still `oversized` after the split"*. That condition is
vacuously true — `oversized` fires at a 1,600-token threshold nothing in this programme has ever met —
and the dispatch measured the cut as buying **~400 of ~15,900 tokens, about 3%**. Meanwhile
**`multiple-goals` had already cleared** from the 16.1b split alone, and that is the warning a cut can
actually act on. The authorisation's *purpose* was therefore already served, and taking the seam anyway
would buy 3% at three costs: a whole story's dispatch overhead — plan, build, close — for one file and
one test; a guard separated from the module that is its only subject; and, decisively, **the gate would
land AFTER the population it polices**, the inverse of the ordering D-16.R.8 itself insisted on for
16.1b on the reasoning that *"building the gate after the population it polices arrives is how D-8.6.5
shipped green."* Applying that reasoning consistently keeps the scan amendment beside the fetch module
it exists to police. The risk class here is repository hygiene rather than documents or bytes, so the
argument is weaker than it was for 16.1b — but it points the same way, and 3% is not a reason to point
the other way.

## Verification

**Cadence: end-of-run (D-16.R.1).** This story is **not** one of the named overrides — 16.0, 16.2 and
16.3 are — so it gets **no in-story browser run**. Unit, lint and build run in-story, plus a
**compile-only** check of the e2e specs so a spec that fails to compile under its build tag cannot
silently skip. The unrun suites are named as debt in the Delivery Log and are never reported as green.

**In-story:**
- `cd folio-designer && npm run typecheck`
- `cd folio-designer && npm run lint` — expect **exactly 4** `react(only-export-components)` warnings and
  0 errors. Pre-existing, not a regression.
- `cd folio-designer && npm test` — **against the baseline table below, not against "green".**
- `cd folio-designer && npm run test:e2e:compile` — must stay clean.
- `cd folio-designer && npm run scan:font-hosts` — **this story does not amend the scan and must keep it
  green.** Expect exit 0, `0 occurrence(s)`, population **above the floor of 400** (586 files at
  `9cbff85`). A population that *falls* is a finding, not a pass.
- `cd folio-designer && npm run build` (node **exactly `v24.16.0`**), then `npm run verify:offline:red`
  and `npm run verify:offline:wasm`. Assert `maximumCacheAssets` is still 64 and the snapshot module
  consumes no slot.
- `cd folio-go && go test -count=1 ./...`
- `cd folio-go && go vet ./...`; `gofmt -l folio-go` **from the repo root** — run inside `folio-go/` it
  prints an `lstat` error that reads like a pass.
- `cd lint && go test -count=1 ./...` — **`-count=1` is mandatory.** The rules package walks the
  `folio-go` tree and Go's test cache does not track `ReadDir`, so a new file never invalidates it. A
  cached `ok` here is not a measurement.
- The **23** golden digests (`folio-go/byte_neutrality_test.go:100`) unmoved, and `SupportedMajor` still
  2 (`internal/template/version.go:77`).

**Baseline MEASURED at `9cbff85`** — clean tree, wd `/Users/panitw/Projects/folio`, 2026-09-03, node
`v24.16.0`. Report against these numbers.

| Suite | Measured at `9cbff85` | Was, at `384c8ac` |
|---|---|---|
| `folio-go` `go test -count=1 ./...` | **1910 pass / 2 fail / 5 skip** | 1896 / 2 / 5 |
| `folio-go` `go vet ./...` | exit 0, no output | — |
| `gofmt -l folio-go` (repo root) | empty | empty |
| `lint` `go test -count=1 ./...` | four `ok`, no FAIL | four `ok` |
| designer `npm run typecheck` | clean | clean |
| designer `npm run lint` | 4 warnings / 0 errors | 4 / 0 |
| designer `npm test` | **43 files / 437 tests — 1 file / 1 test FAILING** (436 pass) | 43 / 437, 1 failing |
| designer `npm run test:e2e:compile` | clean | clean |
| designer `npm run scan:font-hosts` | exit 0, 0 occurrences over **586** files | 584 files |

The Go count rose 1896 → 1910 because Story 16.1b added `licencesignature_test.go`. The scan population
rose 584 → 586 for the same reason.

**The two baseline reds, and neither is this story's:**
1. `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest — `got 7, need >=20`. The
   **mandated permanent red**. Never "fix" it, never report it as a regression.
2. ⚠ **`src/canvas-authority-contract.test.ts:190`** — **DW-152**, Epic 9/10 lane. Expects `[]`, receives
   `["e2e/e9-5-border-no-ink.spec.ts: /\\bgetComputedStyle\\s*\\(/"]`. Arrived with `a40c34d`. **Not this
   story's to fix** and must not be swept into its commit. If still red at the build dispatch the count
   to hold is **436 pass / 1 fail**; if someone has resolved it, 437 / 0. ⚠ **Because it is already
   firing it cannot fire again** — a second file adopting `getComputedStyle` changes the failure's
   contents, not its status. Every file this story adds is auto-enrolled in that corpus by directory
   walk, so **read the failure's array contents, not just the count.**

**Deferred to the end-of-run catch-up, named as debt in the Delivery Log:** the browser run (a pick with
the network up, with it down, and one whose licence is outside the allowlist); `-tags=matrix` and the
four AD-21 legs; `TestCrossTargetByteIdentity`. Note `TestShippedFacesReproduceFromUpstream` fails under
`-tags=matrix` as a **could-not-execute** (`fontgen: fontTools is not importable by this interpreter`) —
never report that as a byte divergence.

**Re-take at implementation HEAD:** the measured endpoint table in the Code Map is an external dependency
and this story's whole mechanism rests on it. It held on 2026-09-03; it may not hold at build time.

## Review Triage Log

### 2026-09-03 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 7: (high 0, medium 3, low 4)
- defer: 5: (high 0, medium 1, low 4)
- reject: 10: (high 0, medium 0, low 10)
- addressed_findings:
  - `[medium]` `[patch]` `font-name-table.ts:128` — `faceCopyright` walked untrusted fetched bytes with the UNCHECKED `sfntTableDirectory` while `requireStaticTrueTypeTables` exists in the same module for exactly that purpose (its own comment at `:37` claims the guard for it). Switched to the checked reader, added a `byteLength < 12` floor, and bounded every `name`-record slice against `min(name.offset + name.length, view.byteLength)`. Safe for the other two callers: all 27 committed `.ttf` faces measured sfnt version `00010000`. Tests added for `OTTO`/`wOFF`/`wOF2` magic, an HTML error-page body, a 2-byte body and an out-of-range string offset.
  - `[medium]` `[patch]` `App.tsx` — `pickCatalogueFamily` guarded re-entry on `fontChainBusy`, but `setFontChainBusy(true)` was only reached inside `applyFontChain`, i.e. AFTER up to six sequential cross-origin round-trips. Two overlapping picks both passed the guard and two embeds could commit. The hold is now taken once per pick and released in a `finally`; the command half moved to `sendFontChain`, which does not re-take it. A `fontChainBusyRef` backs the flag because two picks dispatched in one tick both read the state value they closed over. The `responseGeneration`/`selectionKey` drop rule is unchanged.
  - `[medium]` `[patch]` `font-index.test.ts:42-45` — the CJK-exclusion test was VACUOUS in its load-bearing half and the real filter was executed by nothing. `scriptsOf` only ever pushes `latin`/`thai`, so `row.scripts.includes('cjk')` could never be true whatever `trimSnapshot` did, and no test imported `trimSnapshot` or `trimFamily`. A test now runs the real `trimSnapshot` over a hand-written `familyMetadataList` (one family per `cjkSubsets` entry, one axes-declaring, one static Thai) and asserts survivors and counts directly. Red-proved by deleting `'korean'` from `cjkSubsets`.
  - `[low]` `[patch]` `font-source.ts:300,315` — `layoutDivergence` was computed, typed, returned and asserted in tests, and read by no consumer; the contract says the divergence "is recorded". It is now written to the browser log at the pick, deliberately not to a UI surface (the contract is explicit that it is an observation, never a refusal), and the consumer is asserted.
  - `[low]` `[patch]` `font-source.ts:78-83` — `licenceFileFor` carried an `MIT` row for an SPDX id `classifyLicenceToken` can never emit, contradicting `font-licence.ts`'s "absence, not narrowing" reasoning about the same identifier. Row dropped; the map is now documented as keyed on what the token table can emit, and a missing row is a stated refusal instead of a fetch of `.../undefined`.
  - `[low]` `[patch]` `font-source.ts:130-146` — `parseFamilyMetadata` decremented `depth` without a floor, so an extra `}` drove it negative and a later `{` returned it to 0, after which a NESTED face `name` could be read as the top-level family name. That is exactly the confusion the `name`-equality confirmation exists to prevent. Depth floored; a malformed file can now only fail to resolve, never resolve to the wrong string.
  - `[low]` `[patch]` `font-source.test.ts` — two test defects. `.replace(/robotoslab/g, 'robotoslab')` was a literal no-op and left the fixture registering a URL with an unencoded space, a shape production had never been shown to handle; and `if (outcome.ok) return` sat inside a `for` loop, so the empty-licence-text case was skipped whenever the 404 case regressed. Fixture corrected to `RobotoSlab-Regular.ttf` with an assertion that no asked URL contains a space; the loop is now `it.each`, so both cases always run and report independently.

**Verified before triage, not taken on reviewer confidence.** Every routed finding was checked against the code first. Three claims that arrived as findings were rejected on measurement: the service worker does NOT break the new cross-origin fetches (its `fetch` handler returns early for non-cacheable requests and there is no CSP `connect-src`); the e2e "narrowing" is not a lost bound (the removed clause was made false by this story, and the exact-equality bound on the local tier survives alongside new staleness assertions); and the absence of a browser-side byte cap is the contract's own matrix row, which assigns that refusal to the engine.

**The ten rejects, with the authority each was tested against.** No byte cap on a fetched face (matrix row assigns it to the engine's existing located refusal) · CSP/service-worker breakage (measured false) · e2e narrowing (bound made false by the story itself) · no fetch timeout/`AbortSignal` (not required by the contract; offline is covered by its own row) · undeclared Node type-stripping (`RELEASE_RUNTIME` already pins v24.16.0) · `npm test` writing gitignored `src/generated/font-index.ts` (`npm test` already prepends `build:wasm`, which writes identical content) · the `558` axes figure (prose inside a preserved struck-through block) · `refreshSnapshot` having no ceiling (a manual step whose output lands as a reviewed diff) · an `fvar` check on the fetched bytes (Go is the declared authority; D-16.R.2a) · a pre-emptive offline disclosure (the matrix row specifies a stated refusal, which is what ships).

## Spec Change Log

### 2026-09-03 — plan-gate re-dispatch at `384c8ac` (halt after planning)

Dispatched as a re-plan on the contract the orchestrator amended at `384c8ac` under D-16.R.2,
D-16.R.2a, D-16.R.3, D-16.R.4, D-16.R.5 and D-16.R.6.

- **`status` was set to `draft` for the dispatch and is restored below.** Step-01 routes a
  `ready-for-dev` spec straight to step-03 IMPLEMENT; `draft` is the workflow's own re-plan route. A
  spec being re-planned is not ready for dev, so this is a factual correction, not an improvisation.
- **`<intent-contract>` was preserved VERBATIM** (step-02 item 1, `preserved_intent_contract`) and is
  byte-identical to `384c8ac`: md5 `554cc9140f311a087061c1cc336d5163`, 17,923 bytes, verified before
  and after every edit in this dispatch. Unlike the Story 8.5 re-plan, the orchestrator supplied **no
  replacement scope** — it supplied rulings it had *already applied to the contract itself* — so
  preservation is the correct handling and declining it would have been the builder editing a block it
  may not edit.
- **The plain-terms opener was rewritten.** It was written before D-16.R.2 and D-16.R.2a and had gone
  false in two places: it said variable-only families are *"refused, and said to be refused"* (now
  filtered out), and it said the ones worth having are *"converted ahead of time by the tool that
  already produced the three typefaces Folio ships — which is how the two Thai families in that group
  are already available"*. Measured: `tools/fontgen/instance_faces.py` drives a hardcoded three-entry
  list of **engine** faces, **none of the 21 designer catalogue faces is derived**, and of the two Thai
  families only `Noto Sans Thai` is a derivation. It also never mentioned the local face tier, which
  D-16.R.3 makes central. It now describes the story being built.
- **Every Code Map anchor was re-measured at `384c8ac`** and the corrections recorded in a new
  subsection rather than by rewriting the originals, because two of the rotted anchors are also quoted
  inside `<intent-contract>`, which cannot be edited. Six anchors were wrong; the most consequential is
  that **Story 16.0's `fvar` check is not inside `embedFontFamily` at all**.
- **`## Verification` was rewritten to the run's end-of-run cadence** (D-16.R.1). The previous version
  called for *"a browser run"*, which contradicts the cadence: 16.1 is not one of the two named
  overrides. It also omitted lint, the e2e compile check, `-count=1` on the `lint` module, and
  `npm run scan:font-hosts`'s population floor.
- **A second, undeclared baseline red was found** — `canvas-authority-contract.test.ts`, introduced by
  `a40c34d`. Recorded under Verification. It is not this story's and must not be fixed here.

### The nameID 13 falsifier was taken at the plan gate and it came back RED

The contract's Block If — *"the nameID 13 guard would ship without its falsifier … sample ≥50 families
across `ofl`/`apache`/`ufl` and report the rate at which the guard would refuse them **before** it
ships"* — was **taken at the plan gate rather than deferred to the build's first task**, so that a red
halts before a design is committed to. The build must still run it; a disagreement between the build's
run and this record is itself a halt.

**Method.** 133 families attempted across all 44 `apache/` directories, 86 `ofl/` families spread A–Z,
and all 5 `ufl/` families; `METADATA.pb` parsed for its `license:` token and its
`style:"normal" weight:400` filename; that TTF fetched from `raw.githubusercontent.com/google/fonts/main`
and its sfnt `name` table parsed in pure Python (no `fontTools`) for nameID 13 and nameID 0. **100 faces
had a usable static Regular and parsed; 33 were variable-only and skipped; 4 could not be fetched; zero
sfnt parse errors.** Run 2026-09-03 from the session scratchpad; nothing was written to the repository.

**Result: 50 of 100 sampled faces (50.0%) would be REFUSED by the guard as this spec specifies it.**

| Cause | n |
|---|---|
| No signature entry for its SPDX id — **all 41 are `Apache-2.0`** | 41 |
| nameID 13 **absent** (4 `ofl`, 3 `ufl`) | 7 |
| Signature entry exists but nameID 13 does not match | 2 |
| **Pass** | **50** |

**The four findings, in the order they matter:**

1. **`Apache-2.0` is admitted by D-16.R.4's token table and has no entry in D-16.R.5's signature
   table.** Measured: the build-time table (`src/font-catalogue.test.ts:197-200`) has **exactly two**
   entries, `OFL-1.1` and `Ubuntu-font-1.0`, and this spec instructs Go to mirror it. So every
   `APACHE2` family passes the licence gate and is then unconditionally refused one step later — 41% of
   the sample. A signature **is** cleanly derivable: 33 of 33 non-empty apache nameID 13 values are the
   byte-identical string `Licensed under the Apache License, Version 2.0`, so
   `/Apache License,?\s+Version 2\.0/i` matches with zero variance. Adding it drops refusals from 50%
   to **17%**. But minting it is the decision D-16.R.5 explicitly reserves — *"comes back to the
   engineering lead as a finding, never fixed by softening the guard"* — and this gate will not mint it.
2. **The `Ubuntu-font-1.0` signature entry never fires positively on the upstream corpus.** All three
   static `ufl/` families (`ubuntu`, `ubuntucondensed`, `ubuntumono`) carry **no nameID 13 at all**;
   their licence sentence is in **nameID 0** (`Copyright 2011 Canonical Ltd.  Licensed under the Ubuntu
   Font Licence 1.0`). The regex's British spelling is right and it is pointed at the wrong record.
   **This is a mechanism problem, not a table problem** — no added row fixes it — and it is invisible to
   the build-time tie because the two Ubuntu faces committed here are not the upstream `ufl/` files.
3. **14% of faces have no nameID 13.** The contract says absent → refuse, which is coherent; the
   measurement is that this rejects roughly one face in seven for a reason the author cannot act on.
   Families: `apache/` — `jsmathcmbx10`, `jsmathcmex10`, `jsmathcmmi10`, `jsmathcmr10`, `jsmathcmsy10`,
   `jsmathcmti10`, `yellowtail`; `ofl/` — `candal`, `gfsneohellenic`, `tangerine`, `monoton`; `ufl/` —
   all three. **nameID 0 was present and non-empty on 100 of 100**, which is what makes a nameID 0
   fallback viable at all.
4. **Two faces where the binary contradicts its own metadata — which is the guard working.**
   `apache/mountainsofchristmas` declares `APACHE2` in `METADATA.pb` while its nameID 13 states the
   **SIL OFL 1.1**; `ofl/nanumbrushscript` has `NHN Corporation` in the licence-description slot, naming
   no licence. Note the first would be caught only *incidentally*, as a signature miss, not diagnosed as
   a contradiction. Separately, `ofl/wdxllubrifonttc` carries a correct OFL 1.1 statement **in
   Traditional Chinese** — semantically compliant, invisible to any ASCII substring signature, and a
   floor that widening the regex cannot get under.

**Why this halts rather than being ruled here.** The `ofl` leg alone would be a tuning question. The
`apache` and `ufl` legs are not: one requires minting a signature that D-16.R.5 reserves to the lead,
and the other requires changing the guard's **shape** (a nameID 0 fallback), which no ruling
contemplates and which the contract's *"absent or unparseable nameID 13: REFUSE"* currently forbids.
Two settled rulings point at observably different outcomes for an entire upstream directory, and
nothing in the intent selects between them. That is an intent gap by the workflow's definition, and it
is the escalation this story's own Block If names.


### 2026-09-03 — second plan-gate re-dispatch at `9cbff85` (halt after planning), after the story was CUT

The intent gap recorded above is **resolved**: D-16.R.7 replaced D-16.R.5's guard contract with a
three-valued, two-sided one; D-16.R.8 split the Go guard into **Story 16.1b**, which has since been
built, reviewed and closed at `f6953da`; D-16.R.9–D-16.R.12 carry the supporting rulings. This entry
records what the cut did to this spec.

- **`status` was set to `draft` for the dispatch and is restored to `ready-for-dev` below.** Step-01
  routes a `ready-for-dev` spec straight to step-03 IMPLEMENT; `draft` is the workflow's own re-plan
  route. A spec being re-planned is not ready for dev.
- **`<intent-contract>` was preserved VERBATIM** (step-02 item 1, `preserved_intent_contract`) and is
  byte-identical to `9cbff85`: sha-256 `c23d45c32e858cb28d1931eb41752ce6c8ed4c5e02f5c5d94b3b425229e24ec7`,
  17,578 bytes, verified before and after every edit in this dispatch. The orchestrator supplied
  **rulings it had already applied to the contract**, not replacement scope, so preservation is the
  correct handling.
- **CUT 1 — the Go guard is gone from this story.** Removed: the
  `folio-go/component_commands.go` nameID-13 task, the ≥50-family sample task, and the two ACs that
  carried them. It is built: `fontset.RefuseContradictedLicence` at
  `internal/fontset/licencesignature.go:298`, called from `component_commands.go:2443`. This story must
  not reimplement it and must not grow a browser-side substitute.
- **CUT 2 IS REVERSED BY THE PLAN GATE (orchestrator, 2026-09-03).** The forbidden-host-scan amendment
  **stays in this story** — see `## Design Notes` → *"Why the second seam was declined"*. The paragraph
  below is preserved as the dispatch proposed it, and is superseded.
- ~~**CUT 2 — the D-16.4 forbidden-host-scan amendment is delegated to a successor story**, taking the~~
  seam D-16.R.8 pre-authorised. Removed: the
  `scripts/forbidden-font-hosts.mjs` + `src/forbidden-font-hosts.test.ts` task and its AC. What remains
  here is only the **declaration marker in the font-source module**, so the successor has a subject.
  ⚠ **Measured, and the reason the cut is clean:** neither `raw.githubusercontent.com` nor
  `fonts.google.com` is in `FORBIDDEN_FONT_HOSTS` (`forbidden-font-hosts.mjs:41-44`, exactly two
  entries), so this story's new hosts do not trip the scan and it stays green unamended.
- **`multiple-goals` was re-evaluated after the cut rather than carried forward unread**
  (D-16.R.8's guardrail) and **has been REMOVED**. With both cuts taken, what remains is one goal — the
  fetch path and its admission decision — plus one one-line record (the `parse.go` comment) that is not
  a shippable goal. `oversized` **stays**: the flag's threshold is 1,600 tokens and no spec in this
  programme has ever been near it.
- **The `deferred: []` register instruction for DW-150 was removed** — DW-150 is **CLOSED as
  reconciled** under D-16.R.7 (`deferred-work.md:7023`), so the old task would have re-opened a closed item.
- **Two tasks were ADDED from measurements taken at this gate**, both recorded in the Code Map: the
  mirror-contract extension onto `src/font-licence.ts` (D-16.R.10 names the gap and nominates the test),
  and the extraction of the nameID-0 reader that already exists twice in unimportable form.
- **Every Code Map anchor was re-measured at `9cbff85`** and the previous *"Anchor corrections"*
  subsection folded into a single measured map. **`## Verification` was re-baselined** at the same
  commit: `folio-go` 1896 → **1910 pass**, scan population 584 → **586 files**, both because 16.1b landed.

## Auto Run Result

Status: done
Blocking condition: none

**Dispatch:** build of Story 16.1 at `baseline_revision` `bb14662`, 2026-09-03. Tree clean at dispatch
and clean at finalization. Commits on `main` only; no branch, no push, no `git add -A`.

### What was implemented

The **fetch path and its admission decision**: a family the author picks becomes a face in the
document with terms this product accepts, or it does not become one at all. The build-time index
snapshot ships with the designer and the browser never fetches the list; the typeface itself is
fetched at the moment of a pick, from `raw.githubusercontent.com/google/fonts` and never from
`fonts.googleapis.com/css2`. The 21 committed faces survive untouched as the local face tier, joined
on exact `family` equality, and a local pick issues no request at all.

### Files changed

- `folio-designer/src/font-licence.ts` (new) — the closed token→SPDX table, three states, no
  fall-through and no permissive default. Names no host.
- `folio-designer/src/font-source.ts` (new) — the single declared home of every allowed host;
  slug-and-confirm, the probe order, the Regular filename READ from `style:"normal" weight:400`, the
  upstream licence file, and classify-before-fetch enforced by ordering.
- `folio-designer/src/font-index.ts` (new) — the local-tier join, the variable-only filter, and the
  addable-count disclosure.
- `folio-designer/src/font-name-table.ts` (new) — the extracted nameID-0 reader; the two pre-existing
  hand-written copies now use it. No font-parsing dependency added.
- `folio-designer/scripts/build-font-index.mjs` (new) + committed `font-index.json` → the gitignored
  `src/generated/font-index.ts`. The emit step reaches no network, because an offline release build is
  a shipped gate.
- `folio-designer/scripts/forbidden-font-hosts.mjs` + `src/forbidden-font-hosts.test.ts` — the scan's
  second half, beside the module it polices.
- `folio-designer/src/App.tsx` — `pickCatalogueFamily` gains a source; `embedFontFamilyCommand` stays
  at twelve fields.
- `folio-go/internal/fontset/licencesignature_test.go` — the mirror contract now reads
  `src/font-licence.ts`'s admitted set, not only the build-time table.
- `folio-go/internal/template/parse.go` — the one-door asymmetry comment (D-16.R.11).
- `SPEC.md`, `font-catalogue.md`, `folio-format.md`, `deferred-work.md` (DW-158, DW-159).

### Review findings

7 patched (0 high, 3 medium, 4 low) · 5 deferred · 10 rejected · 0 intent_gap · 0 bad_spec.
See `## Review Triage Log`. Follow-up review recommended: **true** — `3 × 3 + 1 × 4 = 13`, at or above
the threshold of 5. No patched finding was `high`.

### Verification performed

Re-run by the builder after the patches, not relayed from the implementation subagent. Baseline was
re-measured at `bb14662` on a clean tree before any change and matched the spec's `9cbff85` table
exactly.

| Suite | Baseline `bb14662` | After patches | wd |
|---|---|---|---|
| `go test -count=1 ./...` | 1910 pass / 2 fail / 5 skip | **1910 / 2 / 5** | `folio-go` |
| `go vet ./...` | exit 0, no output | exit 0, no output | `folio-go` |
| `gofmt -l <abs>/folio-go` | empty | empty | repo root |
| `go test -count=1 ./...` | four `ok` | four `ok` | `lint` |
| `npm run typecheck` | clean | clean | `folio-designer` |
| `npm run lint` | 4 warn / 0 err | **4 warn / 0 err** | `folio-designer` |
| `npm test` | 43 files / 437, 1 fail | **47 files / 506, 1 fail (505 pass)** | `folio-designer` |
| `npm run test:e2e:compile` | clean | clean | `folio-designer` |
| `npm run scan:font-hosts` | exit 0, 0 occ / 586 files | **exit 0, 0 occ / 598 files** | `folio-designer` |
| `npm run build` + `verify:offline:red` + `verify:offline:wasm` | — | all exit 0, node `v24.16.0` | `folio-designer` |

**The two baseline reds are unchanged and neither is this story's.** `TestCorpusMeetsP6ExerciseFloors`
and its `P6g_(opaque_names)` subtest are the mandated permanent red. `canvas-authority-contract.test.ts:190`
is DW-152, and its **received array contents were read, not just its status**: still exactly
`["e2e/e9-5-border-no-ink.spec.ts: /\bgetComputedStyle\s*\(/"]`, so none of the seven files this
story adds to that directory-walked corpus introduced a violation.

**Invariants asserted.** 23 golden digests unmoved (`go test` set byte-identical between
`bb14662` and now: no test gained, lost or changed status). `SupportedMajor` still 2. `maximumCacheAssets`
still 64, `s1.assetCount` **44**, and the snapshot module consumes **no** cache slot (verified absent
from the emitted manifest). `font-catalogue.json` and the 21 committed faces untouched
(`git diff --stat` empty for both paths).

**Three red-proofs taken by the builder, by DELETION rather than by falsifying a condition.**
1. Removing `src/font-source.ts`'s declaration markers makes `npm run scan:font-hosts` **exit 1** with
   both hosts reported as `(declared-only)` findings and guidance naming the module. Restored by `cp`,
   byte-identical.
2. Emptying `DECLARED_ONLY_FONT_HOSTS` reds **3 of the 14** tests in `forbidden-font-hosts.test.ts`, so
   the new half is non-vacuous.
3. Adding a fourth SPDX id to `font-licence.ts` reds `TestGoLicenceTableSubsumesTheDesignerTable` with
   the message naming the hazard — that Go's signature table has no row for it, so a face declaring it
   would be tied against nothing and admitted as NO EVIDENCE.

**Matrix test audit: satisfied.** Every I/O matrix row has at least one covering test that RAN and
PASSED. The two rows the contract assigns to the engine are covered by Story 16.1b's shipped Go tests,
both confirmed green in this dispatch's run:
`TestEmbedFontFamilyRefusesAFaceWhoseOwnBytesContradictTheDeclaredLicence` (contradiction refuses) and
`TestRefuseContradictedLicenceAdmitsSilence` (nothing readable admits).

**Endpoints re-measured at implementation HEAD**, as `## Verification` requires: the index still 200
with **no** `access-control-allow-origin`; `METADATA.pb`, `OFL.txt` and the static TTF all 200 with
ACAO `*`. The index body measured 2,700,544 bytes against the plan gate's 2,699,611 — a live endpoint
drifting, not a discrepancy.

### Residual risks

1. **No browser run, and it is the sharpest unverified edge.** Deferred by D-16.R.1's cadence, which
   denies this story an in-story browser run. Every web-tier claim is witnessed in jsdom against a
   stubbed fetcher; the CORS facts the whole mechanism rests on are measured by `curl`, never by
   anything the suite executes. Registered as the first `deferred:` item.
2. **`-tags=matrix`, the four AD-21 legs and `TestCrossTargetByteIdentity` were not run** — end-of-run
   catch-up per the cadence. Not reported as green.
3. **The snapshot ages between releases**, by design and by CORS. The list's staleness is stated in the
   UI; only the typeface is live.
4. Cache headroom is now 20 of 64 slots, down from 41 at Epic 8, and no build script checks it.

## Delivery Log

### 2026-09-03 — done

Baseline `bb14662`. Closed on `main` at the commit named at the end of this entry; nothing pushed —
`origin/main` is still `c985b9c`, five commits behind.

**What shipped.** The fetch path and its admission decision, as the cut left it: a build-time index
snapshot the browser never fetches at runtime, a live face fetch from the one declared repository host,
the closed token→SPDX table with its three states, the local-face-tier join on exact `family` equality,
the variable-only filter with an addable count that says which count it is, the slug-and-confirm
directory rule, the forbidden-host scan's second half beside the module it polices, the extracted
nameID-0 reader, the extended Go mirror contract, and the `parse.go` one-door asymmetry comment. The 21
committed faces, `font-catalogue.json` and every per-face licence file are untouched — verified by an
empty `git diff` over those paths across `bb14662..HEAD`.

**Findings triaged by the build: 7 patched / 5 deferred / 10 rejected**, 0 intent_gap, 0 bad_spec. The
routes reconcile against the declared population (7 + 5 + 10 = 22). The ten rejections are **enumerated**
with the authority each was tested against, so they could be spot-checked rather than taken on a count.

**The follow-up review (`followup_review_recommended: true`) was given, and all three mediums have
teeth.** Each was probed by reinstating the finding's own defect, not by weakening an assertion:

1. *Unchecked table directory over fetched bytes.* `faceCopyright` now calls the checked reader. Putting
   the unchecked directory back reddened **2 tests across 2 files** (scope: `font-name-table.test.ts` and
   `font-source.test.ts`), including the fetched-body case, which stops refusing a non-TrueType container
   and walks it instead. The guard is applied **before** the walk and a malformed directory is refused.
2. *Re-entry hold released before the network work.* Reinstating the pre-patch shape exactly — the hold
   taken only inside the command half, and read from React state rather than the ref — makes
   `engine.request('command', …)` fire **twice** on two overlapping picks. **The double-commit is real
   and observable**, and the new test asserts it. Worth recording precisely: under the mutation the test
   reds at its *first* assertion (the control is not disabled), so the load-bearing assertion was probed
   separately by suspending that line — with it suspended the second pick performs a **second full
   resolution (2 `METADATA.pb` fetches, not 1)**, and with the ref restored to the command half as well,
   a **second embed command**. All three arms were re-run at close; the file was restored byte-exact by
   `cp` and md5-verified.
3. *Vacuous CJK exclusion.* Deleting `'korean'` from `cjkSubsets` now reds the new fixture test, with the
   survivor list showing 3 families where 2 are expected. Confirmed at close, restored byte-exact.

**A regression introduced BY the concurrency patch was found at close and fixed here.** The patch backs
the busy flag with a ref because a state read inside a handler is stale — correct — but the flag now
lives in **two places** and the document-reset path in `setCurrentSnapshot` cleared only the **state**
copy. Replace the document while a pick is in flight (Start blank is reachable during the six-round-trip
window; it is guarded on `fileBusy`, not on the font hold) and the pick's own `finally` declines to
release a hold that no longer matches its generation, while the reset clears the half the guard does not
read. `fontChainBusyRef` then stays `true` for the rest of the session: the combobox looks perfectly
enabled and **every subsequent pick silently does nothing**. This is the same defect class the ref was
introduced to fix — one flag, two copies, one of them updated. Fixed by routing that reset through
`holdFontChain`, with a test that reds without the fix (measured: the second pick sends no command at
all) and passes with it. One line of production change, inside the story's own patch, so it is in scope.

**Gates measured at close, on a clean tree, wd `/Users/panitw/Projects/folio`, node `v24.16.0`.** These
are my numbers, not the build's.

| Suite | Measured at close | Baseline `bb14662` |
|---|---|---|
| `folio-go` `go test -count=1 ./...` | **1910 pass / 2 fail / 5 skip** | 1910 / 2 / 5 |
| `folio-go` `go vet ./...` | exit 0, no output | exit 0 |
| `gofmt -l <abs>/folio-go` (repo root) | empty | empty |
| `lint` `go test -count=1 ./...` | four `ok`, no FAIL | four `ok` |
| designer `npm run typecheck` | clean | clean |
| designer `npm run lint` | 4 warnings / 0 errors | 4 / 0 |
| designer `npm test` | **47 files / 507 tests — 506 pass, 1 fail** | 43 / 437, 1 fail |
| designer `npm run test:e2e:compile` | clean | clean |
| designer `npm run scan:font-hosts` | exit 0, 0 occurrences over **598** files | 586 |
| designer `npm run build` + `verify:offline:red` + `verify:offline:wasm` | all exit 0 | — |

The designer total is **507**, one above the build's 506, because this close added the regression test
described above. **23 golden digests unmoved** (`byte_neutrality_test.go` byte-identical to baseline, 23
records counted); `SupportedMajor` still **2**; `maximumCacheAssets` still **64** with **44** assets and
`s1.assetCount` **44**, and the index snapshot consumes **no** cache slot (absent from the emitted
manifest). Headroom is therefore **20 slots**, which is registered as DW-162.

**The two standing reds are unchanged and neither is this story's.** `TestCorpusMeetsP6ExerciseFloors`
and its `P6g_(opaque_names)` subtest are the mandated permanent red. DW-152 was verified **independently
at `bb14662`**, not taken on the build's word: the contract test file is byte-identical between baseline
and HEAD, and running it in a detached worktree at `bb14662` produced the **same single-element received
array** as at HEAD — `["e2e/e9-5-border-no-ink.spec.ts: /\bgetComputedStyle\s*\(/"]`. So none of the
files this story adds to that directory-walked corpus introduced a violation, and the array's
*contents* were read rather than its status. **The "pre-existing" claim reproduces.** (A third
countable claim in the build's record is off: it says **seven** files were added to that corpus.
Measured, `git diff --name-status bb14662..HEAD` adds **nine** tracked files under `src/`, plus the
gitignored generated snapshot module, which the walk also enrols. The conclusion is unaffected — none
of them violates — but the number was not measured.)

**The Go count discrepancy is resolved, and the build's account of it is confirmed by a second route.**
A patch subagent reported 1908 against the build's 1910; measured at close, it is **1910**. The stronger
check is structural: `git diff bb14662..HEAD -- folio-go/` adds **no `func Test` and no `t.Run`** — the
mirror-contract extension grew an existing test's body — so the suite is genuinely unchanged in shape,
not merely equal in total.

**Two figures in the build's own record disagree with each other, and the commit message is the stale
one.** `d6d51f1`'s body says the scan was green over **597** files; the `## Auto Run Result` says **598**;
measured at close it is **598**. Since `04068d6` adds **no file** (`git diff --name-status` shows only
modifications), the population was 598 at `d6d51f1` too, so 597 was transcribed rather than measured. The
same commit body's "47 files / 497 tests" is not a discrepancy — it is the pre-patch count, correctly
stated for its moment. Separately, the `## Spec Change Log`'s claim that the intent contract is
"byte-identical to `9cbff85`" is **now false as a statement about this file**: the contract was changed by
the orchestrator at `c1accf6` and has been byte-stable from `c1accf6` through this close (17,721 bytes,
sha-256 `aaea1949…`), including across both build commits and this one. The build preserved it correctly
from its own baseline; the record simply quotes a superseded gate.

**Rejections spot-checked.** The ten are enumerated, so each could be tested against the authority it
cites. Nine hold as stated. **One is scope-correct and its consequence has since changed:** the build
rejected "no fetch timeout / `AbortSignal`" because the contract does not require one and offline has its
own matrix row — true, and confirmed: there is no `AbortSignal` anywhere on the pick path. But the same
dispatch's concurrency patch moved the busy hold to **before** the fetch chain, so a fetch that *hangs*
(as opposed to rejecting, which is the offline case the row covers) now disables the control for the
session. That is not a defect in the patch and the trade is the right way round; it is a residual the
rejection could not have priced, and nothing re-read it afterwards. Filed as **DW-165** rather than fixed
here, because adding a timeout policy is a design decision with a blast radius beyond this story.

**Red-proofs re-run at close, not relayed.** Stripping `src/font-source.ts`'s two declaration markers
makes `scan:font-hosts` throw with exactly **2 findings**, both labelled `(declared-only)`, naming file,
line and host with the guidance that points at the one declaring module. Emptying
`DECLARED_ONLY_FONT_HOSTS` reds **3 of the 14** cases in `forbidden-font-hosts.test.ts` — the build's
count, confirmed. Both files restored by `cp` and md5-verified; the scan is green over 598 files
afterwards and the release build was re-run after every mutation, so no gate here was measured against a
mutated or half-generated tree.

**Suites NOT run, deferred to the end-of-run heavy-test catch-up (D-16.R.1), and not reported as green:**
`-tags=matrix`; the **four AD-21 legs**; `TestCrossTargetByteIdentity`; and **every browser spec** —
`npm run test:e2e` was not executed, only `test:e2e:compile`. Consequently **this story has no browser
witness for any web-tier pick or refusal**: every fetch, classify, refuse and embed claim is asserted in
jsdom against a stubbed fetcher, and the CORS facts the mechanism rests on are `curl` measurements in
prose. That is intent-authorised by the cadence and it is stated here rather than implied. Note that
`TestShippedFacesReproduceFromUpstream` fails under `-tags=matrix` as a **could-not-execute**
(`fontgen: fontTools is not importable by this interpreter`) — never report that as a byte divergence.

**Deferrals registered.** The build wrote **DW-158** and **DW-159** into `deferred-work.md` as its spec
tasks required, but left all **five triaged deferrals in this spec's `deferred:` frontmatter only**, where
nothing sweeps them. All five are now in the register, plus one filed from the re-read rejection:

- **DW-160** — the recorded `source` names the `main` branch, not a commit, so a `.folio`'s provenance is
  not reproducible. **Filed LOW by the build and raised to MEDIUM here.** The premise of the format is
  that the file travels alone and states what it redistributes (D-8.6.1); a branch name is not a fixed
  point, so two authors embedding different bytes record an identical, unfalsifiable string — and it
  leaves DW-158's "a newer release is a different face" argument enforceable on the local tier and
  unenforceable on the web tier. Owner: the engineering lead, before Story 16.2 makes a fetched face
  travel.
- **DW-161** — no browser witness for the web tier; the largest evidence gap in the epic so far.
  Ordering-linked to **DW-101**, which is the CI half of the same obligation.
- **DW-162** — cache headroom is **20 of 64**, not the 41 an Epic 8 note records, and **no build script
  checks the margin**. Over the ceiling fails; approaching it is watched by nothing. Trigger: the next
  story that adds a cache asset.
- **DW-163** — 66 addable families declare neither Latin nor Thai and embed a face that draws nothing.
- **DW-164** — `popularity` is emitted and read by nothing, so the browser shows the alphabetical head of
  the library while D-16.R.2's own argument turns on which families a user reaches first.
- **DW-165** — the stalled-fetch residual described above.

**What was NOT done, so silence is not read as verification.** No browser suite ran. `-tags=matrix`, the
AD-21 legs and `TestCrossTargetByteIdentity` were not run. The upstream endpoints were **not** re-measured
at close — the build re-measured them at implementation HEAD and that measurement is taken on report; a
CORS posture change since then would be invisible here, which is DW-161's whole point. Rejections were
spot-checked against their cited authorities but not independently re-derived from the reviewer's original
wording, which this record does not carry.
