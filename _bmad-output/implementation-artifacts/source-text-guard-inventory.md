# Source-text guard inventory

**Commissioned 2026-09-05.** An exhaustive census of every test in this repository that reads a source
file off disk and matches its **text**, classified by two properties that decide whether the guard can be
trusted: **does it strip comments**, and **does it break when a line wraps**.

**Method.** `/usr/bin/grep` throughout (the shell `grep` is a shadowing function — see D-000.20), every
absence claim paired with a positive control on the same file, and `-r` rather than `git grep` so
untracked files are included (D-15.2.1). Excluded: `node_modules/`, `.git/`, `folio-designer/dist/`,
`folio-designer/test-results/`, `.fontgen-venv/`.

**Vocabulary.** **STRIPS** — a real blanking step, or an AST/`go/types` parse, runs before matching, so a
commented-out occurrence is invisible. **DOES NOT STRIP** — raw text; a commented-out occurrence counts as
live. **N-A** — the guard's subject *is* a comment (`//go:embed`, a doc marker, a mandated prose
paragraph), or the file has no comment syntax. **wrap-fragile** — the pattern needs the matched phrase to
stay on one line (`^…$` with `m`, `line.startsWith(...)`, `strings.Contains` of a whole expression).

---

## The two findings

### 1. Comment stripping is inconsistent, and the inconsistency is mostly — but not entirely — deliberate

**Strips:** `canvas-authority-contract.test.ts`, `canvas-font-stack.test.ts`, `font-binary-identity.test.ts`
— three *independent copies* of the same quote-aware character scanner, duplicated on purpose so no suite
depends on another's helper. Plus `build-wasm.test.ts`, `font-index.test.ts`, `font-name-table.test.ts`,
`font-source.test.ts`, and the `font-store.ts` arm of `file-access-contract.test.ts`, all via
`forbidden-font-hosts.mjs`'s exported `blankComments`. Plus **every** AST / `go/types` guard in `folio-go`
and `lint`, comment-blind by construction.

**Partial — whole-line `//` only:** `command-json-soleness.test.ts`, `font-provenance.test.ts`,
`command_json_authority_wire_test.go`. Each documents the choice.

**Does not strip, where a commented-out occurrence counts as live:** `design-contract.test.ts` (the
colour-literal and radius bans), `preview-authority-contract.test.ts` (both tests),
`property-prose-height.test.ts` (all), **`engine-bounds-mirror.test.ts` (all 24 tests, including the
`'zh-Hans'` census over every `src/` file)**, the `.postMessage(` arm of `engine-ownership-contract.test.ts`,
the concatenation detector in `command-json-soleness.test.ts`, `font-catalogue.test.ts`'s generator parse,
`font-store.test.ts`, `font-licence.test.ts`'s host ban; and on the Go side `line_spacing_test.go`,
`preview_identity_test.go`, `component_commands_test.go`, `licencesignature_test.go`,
`vendorboundary_test.go`, `table_behaviour_suite_test.go`, plus lint's `retired_absence_test.go`.

**Two scanners read raw text ON PURPOSE and are tested for it.** `forbidden-font-hosts.mjs` and
`host-font-access.mjs` treat a host or an API call **in a comment** as a finding, and blank comments *only*
when computing the declaration exemption — so nobody can buy an exemption by writing a comment. Their
suites assert this in both directions (`fails on a forbidden host that appears inside a comment, in every
comment form`; `fails on a commented-out call, and refuses a declaration written in a comment`).

**Where not stripping has already cost something, measured in the repo's own words:**
- `font-binary-identity.test.ts`'s header: with three engine-named rules commented out, *"every test in
  this file stayed green while the emitted stylesheet dropped to three rules, which is the exact state
  that shipped the reported Thai overlap."*
- `canvas-font-stack.test.ts`: the seam module describes its own exception in a comment spelling
  `new FontFace`, so a raw scan reported the seam as registering **after the registration was removed**.
- `file-access-contract.test.ts`'s header records an earlier whole-file skip that also dropped
  `localStorage` — the very prohibition D-16.2 names *at that module*.
- **`font-catalogue.test.ts` does not strip while its sibling `font-binary-identity.test.ts` strips over
  the same generator file.** Two guards, one subject, opposite policies.

### 2. Line-break fragility is concentrated, and it almost always fails loudly

**The measured precedent is written into `matrixdocs_source_test.go`:** two guards shared one regex
`(?m)^\s*slug:\s*"([^"]+)"`, and *"a seventh entry written as a single-line composite literal … is
gofmt-clean, compiles under `-tags matrix`, and BOTH guards report PASS."* **The fix was to remove the
parsing assumption — move to AST — rather than widen the regex.** That is the pattern to copy.

**Highest concentration:** `engine-bounds-mirror.test.ts` — 24 tests, every extraction `^…$` against an
exact one-line gofmt/prettier spelling. Then `property-prose-height.test.ts`, `design-contract.test.ts`'s
literal `toContain`s, `preview-authority-contract.test.ts`'s JSX attribute string,
`canvas-font-stack.test.ts`'s `line.startsWith('.canvas-text-fragment {')`, and on the Go side
`TestEveryBlockHeightCopyReadsTheModelsAdvance`, `TestPreviewIdentityIsStableFramedAndFiveInput`,
`TestTheBuildTimeTieIsStillInPlace`.

**The reassuring half, and the discipline is applied consistently:** in nearly all of these a broken
extraction produces a **RED** — an explicit non-vacuity floor, a `t.Fatal`, or a `throw` — rather than a
silent pass. The `*RedProof` tests that locate an injection point by exact one-line literal all
`t.Fatalf` with *"this red-proof's injection point is stale"*. `offline-release-contract.mjs` treats 0 or
2 matches as a **throw**, never first-match-wins, and its suite proves it on a commented-out and on an
indented occurrence.

**Only two places deliberately neutralise wraps before matching**, and both say why:
- `font-source.test.ts` — `source.replace(/\n\s*\*\s*/g, ' ')` before matching JSDoc prose.
- `statement_golden_fixture_test.go` — `strings.Join(strings.Fields(strings.ToLower(body)), " ")`, whose
  comment states the rule plainly: *"a required phrase that happens to fall across a line wrap is still
  recorded, and a guard that missed it would be checking the paragraph's line breaks rather than what it
  says."*

**One guard is wrap-DEPENDENT rather than wrap-fragile:** `font-licence.test.ts` requires
`/ABSENCE RATHER THAN\n\/\/ NARROWING/` — the phrase must stay split across those two comment lines.

---

## Scope of the census

**folio-designer / vitest:** 21 test files under `src/` read a file off disk and match its contents,
found by `/usr/bin/grep -rn "from ['\"](node:)?fs"` over all `*.test.ts[x]`; no other fs spelling exists.
Plus two npm-script scanners (`scan:font-hosts` over 631 files, floor 400; `scan:host-fonts` over 146,
floor 86) and `offline-release-contract.mjs`, all three run inside `npm run build` rather than vitest.

**folio-go:** ~24 raw source-text scanners, and a larger comment-blind group using `go/parser` /
`go/types` — the repo migrated deliberately in that direction.

**lint:** 11 rule scanners declaring 20 rule ids; **9 of 11 are comment-blind by construction.** Only
`embedfont.go` and `fontsassets.go` read source text line-by-line, and both are N-A because their subject
is a `//go:embed` directive — itself a comment. Three lint *tests* scan text directly:
`retired_absence_test.go` (does not strip), `stagerank_test.go` (N-A, HTML-comment markers),
`releasing_test.go` (N-A, prose).

**tools/ and hashmatrix/:** no tests of their own. But `tools/fontgen/instance_faces.py` **is read as
source text** by two folio-go tests, which extract its `UPSTREAM = […]` block by a regex requiring the
brackets alone on their own lines.

---

## What to do with this

1. **Prefer AST over regex for any new source guard.** The repo already knows this and wrote down why.
2. **A new guard must declare its comment policy and prove it**, in whichever direction it chose — the two
   host scanners are the model, asserting the behaviour in both directions.
3. **A wrap-fragile extraction must fail loudly.** Non-vacuity floor, `t.Fatal`, or `throw`. The repo is
   already consistent here; keep it that way.
4. **When editing a file a guard reads, check this inventory first.** `App.tsx`, `App.css`,
   `engine-protocol.ts`, `build-wasm.mjs` and `component_commands.go` are each read by several.
5. **`font-catalogue.test.ts` vs `font-binary-identity.test.ts` is an unreconciled disagreement** over the
   same file — worth a deliberate ruling rather than leaving two policies in place.

---

## AMENDMENT, 2026-09-06 — this census had the wrong axis, and three findings it could not see prove it

This document censused guards along **one** axis: *how does this test read source as text, and does it
strip?* That axis found real defects and it is not retracted. But three findings since have been
invisible to it **by construction**, and they are invisible for the same reason: they are not about
reading source text at all.

**The three it missed:**

1. **Directory-listing guards** (D-11.1.6). `lint`'s `fonts-asset-unaccounted` is hardcoded to
   `folio-go/fonts` and walks the tree; `manifest.ResolveAssets` walks the whole repo. Neither reads
   source as text, so neither could appear here — and the question *"which of these covers the designer
   side?"* had to be answered by measurement from scratch. **Recorded in the census's own body as a
   structural blind spot, and now hit twice.**
2. **A guard quantified over the wrong population** (D-11.1.7). The designer's shipped faces split into
   two populations: the catalogue's 31, over which `subfamily === 'Regular'` / `usWeightClass === 400` /
   `macStyle === 0` / `italicAngle === 0` are asserted per face, and the six hardcoded-slot faces, over
   which **nothing at all** is asserted. Both populations ship. Only one is checked. No amount of
   text-reading analysis surfaces that, because the assertion is correct — it is just quantified over
   half the subject.
3. **A record enforced in one direction only** (DW-230). Every consumer of `goldenDigestRecord` iterates
   the declared record, so record → disk is enforced and **disk → record is enforced by nothing**. A
   fixture shipping an `expected.pdf` that nobody registers gets zero coverage and turns nothing red.

**The four delivery axes a census must run along.** The original axis is the first of four, not the
whole job:

| Axis | The question | Found by it so far |
|---|---|---|
| **Source text** | does the guard read source as text, and does it strip comments/strings? | this document's original body |
| **Tree shape** | does the guard enumerate a directory, and *which* directory? | D-11.1.6 |
| **Build output** | does the guard read a built artifact, and was that artifact built the way the guard assumes? | DW-100's `-buildvcs` finding; DW-162's margin |
| **Runtime behaviour** | does anything assert what the surface a human actually reads does — as opposed to what the layer beneath it computes? | D-11.1.8's LoadScreen total |

**And two standing questions that cut across all four**, both of which have now earned their place:

- **Which population is this assertion quantified over — and is that the whole subject?** This is the
  question that produced findings 2 and 3 above, and neither was found by looking for a false zero. It
  is the cheapest high-yield question in the set.
- **Is this record enforced in both directions?** A declaration checked only from the declaration
  outward cannot see its own omissions. The vacuity guards already in place
  (`if len(goldenDigestRecord) == 0`) catch an **empty** record and are structurally blind to an
  **incomplete** one.

**Standing:** the four axes and the two cross-cutting questions apply to every remaining boundary gate.
This document's original body remains valid for the source-text axis and is **not** a census of the
other three; do not read a clean bill here as coverage of anything but reading source as text.
