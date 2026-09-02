---
title: 'Story 16.1: A font arrives from the web with its terms attached'
type: 'feature'
created: '2026-09-02'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: true
baseline_commit: 'a40c34db6cff7372363b2a553710eff48759bef1'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-fonts/font-catalogue.md'
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-8-15-decision-log.md'
warnings: ['multiple-goals']
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The designer has shipped with twenty-one typefaces, chosen when it was built. From here it reaches the
library the web publishes — around two thousand families, thirty-four of them Thai — and fetches the
one the author picks, at the moment they pick it.

Two things about that are less simple than they sound, and the story is built around both.

The first is that the **list** cannot be fetched. Google publishes it, but not in a way a browser is
allowed to read, so the list ships with the designer and ages between releases exactly as the old
catalogue did. What is genuinely live is the typeface itself. Anyone describing this as a live font
browser without that qualification is saying something untrue.

The second is the paperwork, and it matters more than it sounds. A typeface put into a document
carries terms — whose it is, and what the person receiving the file may do with it. Until now those
terms were checked when the designer was built, by a gate that failed the build. Fetching at the
moment of choosing moves that check to the author's machine, in the middle of a click. This story
gives that check a home and a test, because the last time this went wrong seventeen of twenty-one
typefaces travelled under another project's licence, and nobody noticed until review.

One limit is deliberate and worth knowing up front. Around a quarter of the library ships as a single
file holding every weight at once, which this product does not accept, because accepting it would mean
guessing which weight you meant and would make the same template print differently on different
machines. Those families are refused, and said to be refused. The ones worth having are converted
ahead of time by the tool that already produced the three typefaces Folio ships — which is how the two
Thai families in that group are already available.

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
- **The nameID 13 tie transfers to runtime, and it lives in GO** (**D-16.R.5**). Without it this story
  replaces a build gate with something **strictly weaker** — exactly D-8.6.5's defect class. The
  build-time tie at `font-catalogue.test.ts:355-366` ties the declared SPDX id to the binary's own
  licence description, *"the one statement of a face's licence that cannot be edited from outside the
  binary"*; `licenceText` from `OFL.txt` and `licence` from `METADATA.pb` are **neither of them in the
  bytes**, so two of the three values this story carries are unearned without it.
  - Sited **inside `embedFontFamily`, beside the `fvar` check Story 16.0 added** — the browser is not
    the only door, and one implementation covers the local tier and the machine store too.
  - Compares the payload's SPDX `licence` against **nameID 13**, **substring, not prefix or equality**:
    measured at build time, `cascadiacode`'s description opens *"Microsoft supplied font…"* and carries
    the OFL sentence further in. An SPDX id with **no signature entry is a refusal**, not a skip.
  - **Absent or unparseable nameID 13: REFUSE, and say which of the two.**
  - **Both guards kept.** The build-time tie is not deleted or weakened because Go now checks it.
- **The engine command does not change.** `embedFontFamily` already requires family, style, licence,
  licenceText, copyright, source and refuses without them (`component_commands.go:2359-2410`). This
  story changes where those values come from and nothing about what the document must contain.
- **The forbidden-host scan is amended and grows, never deleted** (D-16.4). `fonts.googleapis.com` and
  `fonts.gstatic.com` stay **forbidden** — D-16.3 proved them unusable, so forbidding them is free and
  keeps the `woff2`/subset trap shut. Allowed hosts are declared in **one** module; an occurrence
  elsewhere still fails the build. The population floor and positive control extend to the new half.
- **A family that cannot be added is FILTERED OUT, not listed and refused** (**D-16.R.2**, owner). The
  refusal was priced as 28.7% of a long tail; measured, **37 of the 50 most popular families are
  variable-only** — Roboto, Open Sans, Inter, Montserrat, Raleway, Nunito, Oswald, Playfair Display.
  Listing them and refusing means the most common first action in the product fails. **A hidden row is
  a presentation choice, never a guard:** the engine's refusal stays, for anything that reaches it.
- **The browser's family count is the ADDABLE count**, and it says which it is. Not 1,946.
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
| nameID 13 disagrees with the token | `METADATA.pb` says `OFL`, the binary's description does not carry the SIL sentence | **Refused** before any byte reaches `Assets` | Located refusal naming the family |
| nameID 13 absent or unparseable | face carries no licence description | **Refused, saying which of the two** | Located refusal |
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

**Measured endpoints (2026-09-02, wd `/Users/panitw/Projects/folio`, `curl` with `Origin: http://localhost:5173`, tree clean at `a40c34d`)**
- `https://fonts.google.com/metadata/fonts` — 200, 2,699,611 bytes, **1,946 families**, **34 Thai**,
  **no ACAO**. Fields present and usable: `family`, `category`, `subsets`, `designers`, `fonts` (style
  keys), `axes`, `popularity`, `trending`, `size`, `isOpenSource`.
- `https://fonts.googleapis.com/metadata/fonts` — **404**. Not the endpoint.
- `https://raw.githubusercontent.com/google/fonts/main/ofl/<dir>/<Family>-Regular.ttf` — 200, **ACAO
  `*`**, full static TTF (`Sarabun-Regular.ttf` = 90,220 bytes).
- `.../ofl/<dir>/OFL.txt` — 200, ACAO `*`, 4,387 bytes for Sarabun, opening with its copyright line.
- `.../ofl/<dir>/METADATA.pb` — 200, ACAO `*`, carries `designer`, `license`, `category`,
  `fonts { name, style, weight, filename }`.

**Designer (`folio-designer/`)**
- `scripts/build-wasm.mjs` — validates and fingerprints `font-catalogue.json` into
  `src/generated/runtime/catalogue-<id>.<hash>.ttf` and emits `src/generated/font-catalogue.ts` and
  `runtime-fonts.css`. **The index snapshot is generated here**, beside them.
- `font-catalogue.json` + `src/generated/font-catalogue.ts` — the 21-entry catalogue and its typed
  projection (`catalogueFaces`, `scriptFallbackFaces`, `CatalogueFace`). **The shape the snapshot's
  module should echo**, so `App.tsx` reads one kind of thing.
- `src/App.tsx:608-627` — `pickCatalogueFamily`: fetch bytes from a precached URL, compute the
  fallback tail from `scriptFallbackFaces`, send `embedFontFamilyCommand`. **The seam this story
  widens**: same shape, different source, plus the licence and metadata fetches.
- `src/font-chain-command.ts:69-98` — `embedFontFamilyCommand`. Its comment: *"THE FIELD COUNT IS PART
  OF THE CONTRACT"* — 12 fields, matching `componentFields(raw, 12)` in Go. **Do not change it.**
- `scripts/forbidden-font-hosts.mjs` — `FORBIDDEN_FONT_HOSTS`, `DECLARATION_MARKER`
  (`folio:font-host-declaration`), `SCANNED_ROOTS`. Its header states what the scan may and may not be
  used to claim; **the amendment must keep that discipline**.
- `src/forbidden-font-hosts.test.ts` — positive control, population floor, comment-direction test. The
  three tests that keep the scan from being green before it is correct.
- `src/offline-lifecycle.ts`, `src/release-payload.ts:33` (`maximumCacheAssets = 64`) — the offline
  release contract. A snapshot module is source, not a cache asset; **assert that it does not consume
  a slot.**
- `src/font-catalogue.test.ts` — reads assertions out of each binary's own tables (nameID 13 vs the
  declared SPDX id). **The precedent for the runtime licence and `cmap` checks.**

**Go (`folio-go/`)**
- `component_commands.go:2359-2410` — `embedFontFamily`: `componentFields(raw, 12)`,
  `embeddedFontRecord`, `embeddedFaceBytes`, `DecodeFontForRender`, hash, dedupe via
  `assetKeyReferenced`, `maxCanvasFontFamilies` (256, `page_setup.go:513`). **Unchanged by this story.**
- `internal/template/fontasset.go:178-206` — `decodeRecognisedFont` / `DecodeFontForRender` and the
  face-size cap. **The reason `woff2` is not an option.**
- `internal/template/parse.go` — the Story 8.6 rule refusing a chain-named asset without
  `licence`/`licenceText`/`copyright`. **The parser this story must never be able to feed.**

**Docs to amend**
- `_bmad-output/specs/spec-fonts/SPEC.md` `## Non-goals` — strike the *"No live font service"* clause
  in place, in the existing `~~…~~` + **AMENDED** form D-8.4d.1 used, preserving the original wording
  verbatim; and amend D-8.4d.1's own surviving sentence, which this contradicts.
- `_bmad-output/specs/spec-fonts/font-catalogue.md` — `## Why bundled rather than fetched` is now
  false about the bytes and **still true about the list**; say exactly that.
- `_bmad-output/specs/spec-folio/folio-format.md` — the font-asset section: the record is unchanged,
  its **source** is not.

## Tasks & Acceptance

**Execution:**
- **D-16.5 is RULED** (owner, 2026-09-02): refuse variable-only, derive the ones worth having, move the
  refusal to the pick. Implement to the ruling; do not re-open it. The derive-ahead batch is
  `tools/fontgen`'s work and is **not** this story's — this story must simply not refuse a family whose
  static face this repository already carries.
- `src/` — **filter out** rows the snapshot's `axes` field marks variable-only **and** which the local
  face tier does not hold (D-16.R.2). Predictive only; Go still decides, and the engine's refusal stays
  reachable.
- `src/font-licence.ts` — the closed token→SPDX table, three states, no host named, plus the test
  asserting its admitted set is a subset of D-8.5.3's four (D-16.R.4).
- `src/` — the local-tier join: exact `family` equality against `font-catalogue.json`, local wins with
  no fetch, and a test that a local-tier pick issues **no** third-party request (D-16.R.3).
- `src/` — the slug rule and the `METADATA.pb` `name`-equality confirmation, with the probe order and a
  test that the probe never sets `font.licence` (D-16.R.6).
- `folio-go/component_commands.go` — the **nameID 13 tie inside `embedFontFamily`**, beside Story
  16.0's `fvar` check: substring match against a closed signature table, refusal on no-signature-entry
  and on absent/unparseable, distinguishing the two (D-16.R.5). Red-prove by deleting the guard, and
  assert a face with a correct nameID 13 still embeds so the guard cannot be over-broad.
- **The ≥50-family sample**, run and reported before the guard is proposed for merge.
- `deferred-work.md` — register the local-tier divergence (no staleness check in this epic, with its
  trigger), and **DW-150** (the unparsable-face residual, assigned to this story by Story 16.0's close,
  which is where faces stop being catalogue-built).
- `scripts/` — a build step that snapshots `fonts.google.com/metadata/fonts`, trims it to the rendered
  fields, and emits a typed module beside the existing generated catalogue. **Record the snapshot's
  date and family count in the module** so the UI can state its own staleness.
- `src/` — one **font-source module**, the single declared home of every allowed host (D-16.4). It
  resolves a family to its `METADATA.pb`, its Regular filename, its face bytes and its licence text,
  and it is the only file in the repository allowed to name those hosts.
- `src/` — the **runtime licence admission**: the four-identifier allowlist applied to the fetched
  licence before any embed, refusal on anything else, with the refusal's text written for an author.
- `src/` — a **nameID 0 reader** over the fetched face, or an engine-side derivation if the plan gate
  prefers Go to own it; either way `copyright` comes from the binary.
- `src/App.tsx` — widen `pickCatalogueFamily` to take a family from the snapshot rather than a bundled
  `CatalogueFace`, keeping the fallback-tail proposal and the existing command untouched.
- `scripts/forbidden-font-hosts.mjs` + `src/forbidden-font-hosts.test.ts` — the amendment and its
  second half, with the positive control and population floor extended. **The new half must red when
  the font-source module's declaration is removed.**
- Tests: a family resolving end to end from snapshot row to command payload; an unclassifiable licence
  refused; a missing `OFL.txt` refused; a missing nameID 0 refused; offline degrading to the stated
  message; the `woff2` route asserted **absent** from source.
- Docs: the three amendments above, and register in `deferred-work.md` what moving the licence gate to
  runtime now leaves unwatched at build time.

**Acceptance Criteria:**
- Given the family index, when the designer is built, then a trimmed snapshot ships with it carrying
  its own date and count, and the browser never fetches the index at runtime.
- Given a picked family, when its bytes are fetched, then they are a full static TTF from the declared
  host, chosen from `METADATA.pb`, and never a `woff2` or a `unicode-range` subset.
- Given any face that reaches a document, when it is embedded, then `licenceText` is the upstream
  licence file fetched with it, `copyright` is the face's own nameID 0, and `licence` is one of the
  four admitted identifiers — or the pick is refused before any byte is written.
- Given a licence outside the allowlist, when it is classified, then the pick is refused and named,
  never warned about.
- Given the forbidden-host scan, when an allowed host appears outside the declaring module, then the
  build fails; and when the declaring module is deleted, the new half's control reds.
- Given no network, when the author opens the browser, then it states that a family cannot be added
  right now and offers what the machine already holds.
- Given a family that cannot be added, when results are rendered, then it is **filtered out** rather
  than listed and refused, and the browser's count is the addable count and says so.
- Given a family present in the local face tier, when it is picked, then it is embedded from the
  committed bytes with **no fetch at all**, and the index's `axes` field is not consulted for it.
- Given a local face with no index row, when the family control is opened, then it is still offered
  from the local tier, and it does not appear among the web browser's results.
- Given a fetched family, when its licence token is classified, then an admitted token maps to an SPDX
  id that the `.folio` carries, a refused token is named with its reason, and an unrecognised token
  says it was **not recognised** rather than that it is forbidden.
- Given any face about to be embedded, when the command runs, then its SPDX licence is tied to the
  binary's own nameID 13 by substring, and an absent or unparseable description is refused with which
  of the two it was.
- Given the nameID 13 guard, when it is proposed for merge, then a **≥50-family sample across
  `ofl`/`apache`/`ufl`** has been run and its refusal rate reported — and a rate materially above zero
  is a finding for the engineering lead, never a reason to soften the guard.
- Given a family directory, when it is resolved, then the slug rule produced it and `METADATA.pb`'s
  own `name` confirmed it, and the directory the probe found is never used as evidence of the licence.

## Design Notes

**Why the index is snapshotted and that is not a compromise being hidden.** The alternative that would
make it genuinely live is the Developer API, and D-16.3 refused it on arithmetic: it needs an API key
in the client bundle — a secret nobody can keep, on someone else's quota — and it returns no licence
text, so `raw.githubusercontent.com` would still be required. It buys one property and costs two.

**The licence check is the real subject of this story.** The UI work is Story 16.3. What changes
here is that an admission decision which used to fail a build now happens on an author's machine over
bytes nobody reviewed. D-8.6.5 is the precedent and it is not reassuring: 17 of 21 catalogue faces
shipped under another project's licence, green, until a review caught it. That defect was possible
with a build gate watching. This story removes the gate and must replace it with something testable.

**Why 28.7% of the library is refused on purpose.** D-16.5(c) — instancing in the browser — would buy
those families by making the embedded face a function of the author's browser and library version: a
different runtime, different bytes, a different PDF from the same template. `instance_faces.py`'s
header states the same hazard for the build environment and is why its output is committed. Folio has
no backend, so there is no third place to do it. The 28.7% is the price of *"the same template renders
the same PDF everywhere"*, paid deliberately.

**Why `embedFontFamily` is untouched.** It already demands exactly what a `.folio` requires and
refuses without it, so the writer still cannot produce a document its own parser rejects. Changing the
source of those values while leaving that guard in place is what keeps this story from reaching the
format at all.

## Verification

- `cd folio-designer && npm run scan:font-hosts && npm run test && npm run build`
- `cd folio-go && go test ./...`
- The 23 golden digests, unmoved; `SupportedMajor` still 2.
- A browser run: pick a static family with the network up; with it down; and one whose licence is
  outside the allowlist.
- The measured endpoint table in D-16.3 re-taken at implementation HEAD, since it is an external
  dependency and this story's whole mechanism rests on it.
