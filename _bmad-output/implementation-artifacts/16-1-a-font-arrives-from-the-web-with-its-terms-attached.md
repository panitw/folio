---
title: 'Story 16.1: A font arrives from the web with its terms attached'
type: 'feature'
created: '2026-09-02'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: true
baseline_commit: '9cbff85084ed7d8ff3e98e66cbea14d4c45d844b'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-fonts/font-catalogue.md'
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-16-decision-log.md'
warnings: ['oversized']
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The designer has shipped with twenty-one typefaces, chosen when it was built. From here it **also**
reaches the library the web publishes — around two thousand families, thirty-four of them Thai — and
fetches the one the author picks, at the moment they pick it. The twenty-one do not go away. They stay
as the faces this machine already holds, and when the author picks one of those, nothing is downloaded
at all.

Three things about that are less simple than they sound.

The first is that the **list** cannot be fetched. Google publishes it, but not in a way a browser is
allowed to read, so the list ships with the designer and ages between releases exactly as the old
catalogue did. What is genuinely live is the typeface itself. Anyone describing this as a live font
browser without that qualification is saying something untrue.

The second is the paperwork, and it matters more than it sounds. A typeface put into a document carries
terms — whose it is, and what the person receiving the file may do with it. Until now those terms were
checked when the designer was built, by a gate that failed the build. Fetching at the moment of choosing
moves that check to the author's machine, in the middle of a click. **The half of that check that reads
the typeface's own bytes has already been built and shipped** — Story 16.1b, closed at `f6953da`, so
the engine now refuses a face whose bytes contradict the terms written beside them. What is left for
*this* story is the other half: deciding, from what the library says about a family, whether its terms
are ones this product accepts at all, and refusing before a single byte is embedded. The last time this
went wrong, seventeen of twenty-one typefaces travelled under another project's licence and nobody
noticed until review.

The third is that about a quarter of the library ships as a single file holding every weight at once,
which this product does not accept — accepting it would mean guessing which weight you meant, and would
make the same template print differently on different machines. Those families are **not shown**. They
are not listed and then refused: a row the author cannot act on is a row that should not be there, and
the number the browser reports is the number of families that can actually be added, and says so. The
engine still refuses such a face if one ever reaches it — hiding a row is a presentation choice, never
a guard.

One thing worth knowing, because an earlier draft of this page got it wrong. "Ships as one variable
file" is a fact about *the shelf we are looking at*, not about the typeface. Roboto and Inter are both
on that list, and Folio already carries both — as ordinary single-weight files, taken from those
projects' own releases rather than from Google's mirror. So they are offered from the twenty-one, and
the web list's opinion about them is never consulted. Nor is this repository in the business of
converting variable fonts: of the faces it ships, exactly one was produced that way, and every other
one is an upstream file copied byte for byte.

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

_No review pass has run. This dispatch halted after planning._

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
- **CUT 2 — the D-16.4 forbidden-host-scan amendment is delegated to a successor story**, taking the
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

Status: ready-for-dev
Blocking condition: none

**Dispatch:** second re-plan of Story 16.1 at HEAD `9cbff85`, `Halt after planning.`, 2026-09-03. Tree
clean at dispatch; the only file this dispatch modifies is this one. No code was written, no commit was
made, no branch, no push.

**The prior halt is discharged.** The `blocked` / `intent gap` result recorded in the Spec Change Log
above was answered by D-16.R.7 (three-valued, two-sided tie) and D-16.R.8 (the guard becomes Story
16.1b, built and closed at `f6953da`). Nothing in this story now depends on it being re-decided.

**Planning completed** — the story cut twice, `multiple-goals` cleared, all anchors re-measured at
`9cbff85`, the baseline re-taken, and one new Go task and one extraction task added from measurements
taken at this gate.

**Flagged for the gate, none of which blocks the build dispatch:**
1. Two **I/O-matrix rows inside the preserved contract are superseded** — *"nameID 13 absent or
   unparseable → Refused, saying which of the two"* is the D-16.R.5 contract; D-16.R.7 made it **NO
   EVIDENCE → ADMIT** and `licencesignature.go` implements the latter. The rows describe 16.1b's door.
   They may not be used as acceptance here. Amending the contract is the orchestrator's call.
2. Two **Block If clauses in the contract are discharged elsewhere** — the ≥50-family falsifier (run at
   the previous gate, 100 faces) and the before-the-tag narrowing (landed with 16.1b, registered as
   **DW-153**). Neither can trigger in this story.
3. The contract's **Approach names D-16.4's scan amendment**, which CUT 2 moves to a successor story.
   The lead pre-authorised the cut (D-16.R.8); the contract sentence now points at work this story does
   not do.
4. **The successor guard lands after the population it polices** — the inverse of 16.1b's ordering, and
   the shape D-16.R.8's own reasoning warns about. The exposure is repository hygiene (does a host string
   appear outside one module), not documents or bytes, so the class is materially lower than D-8.6.5's;
   recorded rather than absorbed.
5. **DW-152 masks its own class.** The standing red is already firing, so a new `getComputedStyle`
   violation would change the failure's array contents and not its status, and every file this story adds
   is auto-enrolled by directory walk. Read the array, not the count.
