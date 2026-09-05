---
title: 'What closes over a moving set is re-derived from it'
type: 'bugfix'
created: '2026-09-05'
status: 'done'
baseline_commit: 'd365fd47a1b3fa2389d82d3ad2ce5c1d3434d4b8'
review_loop_iteration: 0
context: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs. Written at close to describe
what actually shipped.*

An earlier story added a fourth typeface to the set this product ships, and made one of them the typeface
a new document starts in. Four checks elsewhere were still written against the old arrangement. One
noticed and complained; three did not, because each was checking a number it had worked out from itself,
or a fact it had been told rather than looked up. A check that counts its own inputs can only ever report
that everything is accounted for.

This story rewrites all of them to work the answer out from the real source each time, so the next change
to the shipped set makes them fail rather than quietly agree. The provenance record for each shipped
typeface is now checked against the actual bytes, and every face must be accounted for by exactly one
named route, with no catch-all. A stale assertion the product could no longer ever satisfy was deleted
rather than left standing as an expected failure.

No product behaviour changed, and nothing reported as broken turned out to be broken. An earlier report
that the default typeface could not be chosen was wrong, and the record now says so.

Two things look wrong on purpose. One long-standing text-corpus failure is deliberately left red and is
unrelated to this work. And the heavier cross-platform and browser suites were not re-run at close — this
project runs those once, at the end of the epic.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Story 16.8 (`4d2b27e`) grew `fonts.Shipped()` from three faces to four and renamed
`starter.folio`'s only chain from `body` to `Roboto` — making a catalogue family document-declared and
the default typeface in one move. Four things closed over the old state. One failed loudly and is fixed
(`28cd225`). Three did not: `fontgen`'s reproduction guard takes its denominator from its own manifest
and can only ever say `3 of 3` while four faces ship (D-000.12/D-000.14); Roboto's engine-side
`NOTICE.md` digest table is the one engine-side provenance record nothing asserts; and two assertions in
`e2e/font-embed-boundary.spec.ts` closed over "the starter declares no catalogue family". A fourth
assertion there closed over a sentence Story 16.10 deleted, and is now permanently unsatisfiable.
**No product behaviour is wrong. Every behaviour named here is correct.**

**Approach:** Re-derive each from its actual source and make it fail when that source moves. An untagged
Go accounting guard total over `fonts.Shipped()`; the fontgen witness's denominator moved onto
`fonts.Shipped()`; the two e2e expectations derived from the starter template rather than from a premise;
and the assertion whose subject the product no longer has, deleted.

## Boundaries & Constraints

**Always:**
- Every expectation derives from an **independent** source, never from the artifact it checks
  (D-000.14). Say in the code where each expectation comes from.
- The accounting guard is **untagged**, in `folio-go/fonts`, pure file reading — no fontTools, no venv,
  no network — so it runs per-commit. Behind `matrix` it would re-create D-000.11 exactly.
- Both obligations are **TOTAL over `fonts.Shipped()`** and fail in **both** directions, modelled on
  `TestShippedSpecCoversEverythingShipped`.
- **Exactly one named route per face** — `derived` (in fontgen's manifest) or `static-upstream` (its
  NOTICE records `copied unmodified, no derivation`). **No fallback bucket, no default arm.** Unaccounted
  fails; a face in *both* routes fails. Every message names the face and takes its denominator from
  `fonts.Shipped()`, never a bare ratio.
- A parser that extracts nothing **throws**. An empty expectation is the failure mode being designed
  against, not an outcome.
- Red-prove by mutating the **SUBJECT**, never the expectation. Echo the mutated line back and say where
  the failure landed. **The silent direction is mandatory:** add a fifth face to `Shipped()` with no
  NOTICE and no manifest entry and prove RED. A proof that only *removes* a face re-tests the direction
  that already failed loudly.

**Ask First:**
- A face that is neither `derived` nor `copied unmodified, no derivation`. The answer is a third **named**
  route, never a fallback arm — and that is a ruling, not an implementation choice.
- Any change to product behaviour, or to `choose`/`commit` in `App.tsx`.

**Never:**
- Do not restate Roboto's digest as a literal in Go (`fonts_test.go:27-34` records why). The NOTICE pin is
  a **third comparison**, not a fourth copy of the number.
- No second manifest, and do not bend `instance_faces.py`'s `src`/`out`/instancer shape around a face with
  no derivation step — a route that has to lie about having an instancer step is worse than a second
  named route.
- No hand-written Go table mapping face name → directory: a fourth authority and a D-000.14 tautology.
- Never special-case the name `Roboto`, never hardcode a subtraction — both rot at the next starter change.
- Out of scope: the third e2e failure (fresh authored sessions, 300 s), unreproduced and untriaged; a
  load-time charset constraint on `component.id` (exists — `internal/template/ids.go:validateElementID`);
  `epics.md:5693`'s stale three-group family control.
- **No commits, adds, stashes, checkouts, resets, reverts, restores, branches or pushes.** Work stays in
  the working tree.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| All accounted | 4 faces; 3 in the manifest, Roboto's NOTICE says `copied unmodified, no derivation` | Guard passes; each NOTICE's digest and size equal its embedded bytes | N/A |
| Fifth face, no NOTICE, no manifest entry | `Shipped()` returns 5 | **FAIL**, naming the unaccounted face | Hard failure, never a skip |
| Face in BOTH routes | Roboto added to the manifest while its NOTICE still says `copied unmodified` | **FAIL**, naming the face and both routes | Hard failure |
| NOTICE digest drifts | one byte of a shipped binary changes | **FAIL**, naming the face, the recorded digest and the computed one | Hard failure |
| NOTICE row unreadable | the digest or `Size` row matches 0 or ≥2 times | **THROW** — never read as "no constraint" | Hard failure |
| Starter declares a catalogue family | `starter.folio` declares `Roboto` | Roboto under `IN THIS TEMPLATE`, absent from `AVAILABLE LOCALLY`; picking it sends the command and the engine returns its stable snapshot | No `p.property-error`; revision does not advance |

</frozen-after-approval>

## Code Map

Anchors measured at `71627a5`. **Re-verify before editing.** Recursive `grep` returns false zeros here —
use `/usr/bin/grep -a` by absolute path or `git grep --untracked` (plain `git grep` is tracked-only,
D-000.15) — and state which form you used beside every count.

**The shipped set and its records**
- `folio-go/fonts/fonts.go:79-85` — `Shipped() folio.FontSet`, i.e. `map[string][]byte`
  (`folio-go/fontset.go:20`). Keys: `Noto Sans`, `Noto Sans Thai`, `Noto Sans SC`, `Roboto`. Four
  `//go:embed` at `:41,:44,:47,:66`. **No build tags in the package.**
- The 4 `folio-go/fonts/*/NOTICE.md`. Rows identical in all four, and therefore the only safe parse
  targets: `| **sha256 of the SHIPPED (produced) file** | \`<64hex>\` |`, `| Size | <n,nnn> bytes |`,
  `| **Shipped file** | \`<name>\` |`. ⚠ The derivation statement is **not** uniform — a blockquote in
  the three Noto files (`> **This file is a DERIVATIVE…**`, `:8`) and a **multi-line table cell** in
  roboto (`| Relation to source | **copied unmodified, no derivation** …`, `:65-67`). Do not line-anchor
  one regex across both.
- `folio-go/fonts/fonts_test.go` — `TestShippedRobotoMatchesDesignerCatalogue` (`:35-98`) is the package's
  **only** test. `:27-34` forbids a restated digest literal. `:48-49,:83-84` are the package's disk-read
  idiom: `filepath.Join` on relative segments + `os.ReadFile`.
- **Nothing reads the CONTENT of a `folio-go/fonts/*/NOTICE.md`.** Every TS NOTICE reader targets
  `folio-designer/public/fonts/*`; `lint/internal/manifest/manifest.go:410-430` reads the Go-side ones but
  parses **only** the `Copyright` line. (Population and control recorded in the plan-gate report.)
- **The model to copy:** `folio-go/shipped_faces_test.go:389-444` — both directions, a cardinality
  pre-check, a zero-byte non-vacuity floor, a fresh on-disk length comparison. It is `package folio` and
  cannot see `fonts`. Parser idiom: `folio-designer/src/font-catalogue.test.ts:139-152` (exactly-one-match
  or throw) and `e2e/font-embed-boundary.spec.ts:62-69` (throws rather than returning an empty set).

**The fontgen route**
- `tools/fontgen/instance_faces.py:117-168` — `UPSTREAM`, a literal 3-entry list, no walk, no glob; fields
  include `"key"` and `"dir"`. `:366` `total = len(UPSTREAM)`; `:367` prints
  `derived and compared {compared} of {total} faces`. `roboto` appears nowhere in `tools/`.
- **No machine-readable mode** — flags are only `--sources` (required), `--repo-root`, `--out`,
  `--verify-only` (`:228-254`). You cannot ask it what it derives without ~20 MB of sources and three
  fontTools runs, so **the manifest must be read as text.**
- `folio-go/fontgen_matrix_test.go` — 129 lines, `//go:build matrix`, **`package folio`**; `:117`
  `const wantWitness = "derived and compared 3 of 3 faces"`; `:84-87` `FOLIO_FONTGEN_PYTHON` or bare
  `python3`. ⚠ **It cannot import `folio-go/fonts` — `fonts` imports `folio` (`fonts.go:26`), a cycle.**
  `package folio_test` is the available move; `shipped_faces_ext_test.go:129-149` already provides
  `repoRootForExtTest`.
- **Join key, measured:** `UPSTREAM[i]["key"]` is byte-identical to the `fonts.Shipped()` map key for all
  three — exact equality, no normalisation. `"dir"` maps literally to the directory
  (`instance_faces.py:332,:336`). Roboto joins to neither, deliberately, documented at `fonts.go:50-57`
  and `roboto/NOTICE.md:18-22`.

**The e2e spec** — `folio-designer/e2e/font-embed-boundary.spec.ts`, 294 lines, the one spec in this
directory that is actually executed.
- `:39-41` `families` read from `folio-designer/font-catalogue.json` (31 rows), never hardcoded.
- `:130` `expect([...local].sort()).toEqual([...families].sort())` — `local` is the `AVAILABLE LOCALLY`
  group's `.property-option-name` texts.
- `:142-150` the **registered pre-existing** disclosure assertion. `:130` throws first today, so `:148` is
  never reached; Story 16.10 deleted `familyIndexDisclosure` and its last consumer, so it is now
  **permanently unsatisfiable**, not merely stale.
- `:168` (test name) and `:225-226` (comment over `:227`) are the **only two homes** of the universal
  embed sentence — it is not in `epics.md`. `16-0-…md:370` states the weaker disjunction ("embedded **or**
  refused with a located reason"), which `4d2b27e` does **not** falsify.
- `:179-228` the loop: bare `'EMBEDDED'` literal at `:187,:193,:197`, compared at `:227`; `:194`
  `timeout: 120_000`. `:169-171`'s ONE-FAMILY-PER-BLANK-DOCUMENT note guards the *cross-iteration*
  confound only — not the within-iteration one this story is about.
- `:75-92` `placeAndSelectText` (`page.goto('/')` per iteration) and `currentRevision` (reads
  `REVISION (\d+)` off `getByTestId('engine-snapshot')`).

**Why the two e2e assertions fail — both product behaviours are CORRECT**
- `folio-designer/public/templates/starter.folio` (455 bytes) declares **exactly one** chain:
  `"Roboto": ["Roboto", "Noto Sans Thai", "Noto Sans SC"]`. Node-readable (precedent:
  `folio-go/wasm/cmd/engine/main_test.go:229`); the e2e spec does not read it today.
- `App.tsx:2580` `declared` → `IN THIS TEMPLATE`; `:2604`
  `onThisMachine = offeredFamilies(...).filter((s) => !families.includes(s.family) && familyIsInstalled(s))`
  → `AVAILABLE LOCALLY`. `families` **is the document's declared chain names** (`:2524→1769→1741→1478
  canvas.fontFamilies`; Go side `page_setup.go:662,:778`), so Roboto sits in `IN THIS TEMPLATE` and
  `local` holds 30 of the spec's 31. **Only TWO groups exist** — `ADDED FROM WEB FONTS` is in no shipped
  source.
- `component_commands.go:1722-1734` `defaultFontFamily` (sole caller `:1685` in `createComponent`) —
  first declared non-empty chain in sorted key order, so **every placed text element is born with
  `fontFamily: Roboto`.**
- `App.tsx:2688-2692` `choose` — a `source`-bearing row goes to `commitFirstUse` (embed + commit), a
  declared row to a plain `commit`. `commitFirstUse` has one call site (`:2690`) and `onThisMachine` is
  the only producer of a `source`-bearing row (`:2632`), **so no `IN THIS TEMPLATE` pick can ever embed.**
- ⚠ **`choose` sends the command unconditionally** — no comparison against `committed` (`:2527`, read only
  at `:2746` and `:2820`). The revision does not move because the **engine** no-ops it:
  `folio-go/wasm/engine.go:240-246`, `if bytes.Equal(canonical, e.bytes) { return e.Snapshot(), nil }` —
  *"valid but leave canonical bytes unchanged … not committed mutations: preserve revision, dirty state,
  preview authority, and both history branches exactly as they were."* **That is the AD-15 property: the
  engine decides what is a mutation, not the UI.**

**The AD-15 unit test's home** — `App.test.tsx` (4889 lines), block
`typography controls over the engine-projected closed sets` (**1613-2237**), which owns `select()`
(`:1614-1620`) and every family-control test.
- Model: `:665-672` `keeps a no-op command non-dirty and out of browser history when the engine returns
  its stable snapshot` — **page setup only; nothing covers a property commit or `fontFamily`.**
- ⚠ **Do not add a no-send guard.** `:2960-2973` is a deliberate ruling the other way —
  `WRITES THE SHOWN DEFAULT when the author commits it unchanged`, whose comment names the no-send
  behaviour *"the tempting wrong implementation"*.
- ⚠ Two confirmed traps. `select()`'s default mock returns a snapshot with **no canvas** and
  `App.tsx:1227` replaces the snapshot wholesale, so after the first command resolves the inspector
  unmounts and a later `queryBy…` can pass because the panel vanished — render your own canvas with
  `{ ...textComponent, fontFamily: 'body' }` instead of using `select()`. And Vitest runs `css: false`, so
  `toBeVisible()` proves nothing.
- Negation idiom in this file: the inline `const sent: ArrayBuffer[] = []` recorder, **settled** with
  `await Promise.resolve()` before reading a zero, **plus** a positive control (see `:2216-2218` and the
  vacuous-zero incident noted at `:2226-2233`).

**Rulings implemented:** `epic-11-14-decision-log.md` — D-000.11 (`:910`), **D-000.12 (`:989`)**,
D-000.14 (`:1121`), D-000.15 (`:1159`).

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/fonts/` (new **untagged** `_test.go`) — the accounting guard. **(a)** every shipped face
      has a `NOTICE.md` beside it whose recorded shipped digest equals the embedded bytes and whose
      recorded size equals their length. **(b)** every shipped face is accounted for by exactly one named
      route — `derived` (its key is in the manifest, read as **text** from `tools/fontgen/instance_faces.py`)
      or `static-upstream` (its NOTICE records `copied unmodified, no derivation`). Both total over
      `fonts.Shipped()`, both failing in both directions.
- [x] Same file — **derive the face→directory relation; do not tabulate it.** Recommended: walk
      `folio-go/fonts/*/`, read each NOTICE's `**Shipped file**` row, hash that binary, join to
      `Shipped()` **by bytes** — both directions then fall out and no new authority is introduced. Any
      alternative must name, in a comment, where its expectation comes from (D-000.14).
- [x] Same file — the digest and size parsers **throw** on 0 or ≥2 matches; the manifest reader throws if
      it extracts no keys. Handle both derivation-statement shapes explicitly.
- [x] `folio-go/fontgen_matrix_test.go` — denominator onto `fonts.Shipped()`: report *"derived and
      compared 3 of 4 shipped faces; 1 accounted static-upstream"*, every number derived, or refuse. Keep
      asserting fontgen's own unchanged `3 of 3` witness — that is the manifest's internal self-check and
      a different claim. Needs `package folio` → `package folio_test` (cycle; use `repoRootForExtTest`).
- [x] `e2e/font-embed-boundary.spec.ts:130` — derive the expected `AVAILABLE LOCALLY` set as **catalogue
      families MINUS the starter's declared chains**, reading `public/templates/starter.folio` the way
      `:39-40` reads `font-catalogue.json`. **Add the missing positive case:** assert the subtracted
      families ARE present under `IN THIS TEMPLATE` — removing a family from one group's expectation
      without asserting the other group holds it converts a false red into a real blind spot.
- [x] `e2e/font-embed-boundary.spec.ts:168-228` — **partition the loop by that same derived predicate**,
      one model, no family named. A starter-declared family is an `IN THIS TEMPLATE` row that structurally
      cannot embed, so its arm asserts: the revision does **not** advance, no `p.property-error` appears,
      the document already declares a chain for it, **and** — the non-vacuity leg — the pick actually
      landed (the dropdown closed and the field reads the family). Without that leg the arm passes when
      the click misses entirely. **Both arms must be non-empty, with counts derived from the partition.**
      "No command" is not observable from a browser and must not be asserted here.
- [x] Same file — **rename the test and define the label.** `:168` and `:225-226` are the only two homes
      of the universal sentence; both must now state the asymmetry (a declared family is *picked*, not
      *embedded*). Give `'EMBEDDED'` a defined meaning at its declaration or rename it — a label reading
      EMBEDDED for a pick that embedded nothing is the defect this story exists to remove.
- [x] Same file — **delete `:147-150` and its `:142-146` note**, citing Story 16.10's ruling in the
      deletion's own comment, so the next reader finds out why rather than discovering a hole.
- [x] `folio-designer/src/App.test.tsx` (block 1613-2237) — one new test: picking the family a component
      already has **sends** the command and the engine answers with its **stable snapshot** — revision
      does not advance, document stays non-dirty, Undo stays disabled. Model on `:665-672`; note that
      nothing covers a property commit or `fontFamily` today. **Positive control in the same test:** pick
      a different declared chain and assert exactly one command carrying the changed value.
- [x] **Red-proof every new guard by mutating the SUBJECT**, echoing the mutated line back and saying
      where the failure landed. Mandatory arms: the fifth-face silent direction, in-both-routes, and
      digest drift. Restore every file by absolute path and confirm green afterwards.

**Acceptance Criteria:** (the guard's own failure arms are the I/O matrix; these are the system-level
behaviours it does not cover)
- Given `TestShippedFacesReproduceFromUpstream` with the pinned toolchain, when it runs, then it reports
  three faces derived of four shipped with one accounted static-upstream, every number derived.
- Given `go test -count=1 ./...` in `folio-go`, when it runs, then the accounting guard executes with no
  build tag and without fontTools, a venv, or a network.
- Given the starter declaring a catalogue family, when the e2e spec runs, then that family is absent from
  `AVAILABLE LOCALLY`, present under `IN THIS TEMPLATE`, and both loop arms are non-empty.
- Given a component that already carries a font family, when that same family is picked, then a command
  is sent, the revision does not advance, the document stays non-dirty and Undo stays disabled — while
  picking a different declared chain sends exactly one command carrying the changed value.
- Given `npx playwright test e2e/font-embed-boundary.spec.ts`, when it runs, then it exits 0 — no
  assertion is left red for a human to read past.

## Spec Change Log

## Design Notes

**The question that finds this defect class:** *what would have to change for this guard to fail?*
D-000.14 records four instances this run, all shipped, none caught by review; two are this story's
subject. A guard sized too large announces itself; a guard sized too small congratulates you — which is
why the mandatory red-proof is the ADD direction.

**Why reproduction stays tagged and accounting does not.** Reproduction genuinely needs Python 3.12.13 +
fontTools 4.63.0. Accounting is file reading. Putting the cheap guard inside the expensive suite is
D-000.11 exactly. **CI does not have the fontgen toolchain and cannot acquire it** — a standing property,
and the second argument for keeping accounting toolchain-free.

**Two kinds of assurance under one heading.** The three Noto faces are *reproduced*; Roboto is
*provenance-only* — a vendored static upstream file whose source and shipped sha256 are the same value,
and that identity IS the record. **What is missing is the assertion, not the record.**

**The e2e partition is the honest repair; the tempting one is not.** Setting the element to a different
family first and then picking would make the loop pass — and record `EMBEDDED` for a family that was
never embedded, because a declared row takes the plain-commit branch. A test that goes green by making
its own label false is worse than the red it replaced.

## Verification

**Nothing in Epic 16 is CI-verified** (DW-171); every gate is a local measurement. Say what you did not
run, and why. Exit codes from `$?` **immediately**, never through a pipe; zsh, so `${pipestatus[1]}`. A
test run is evidence about the tree it ran on and nothing else.

**Commands:**
- `cd folio-go && go test -count=1 ./...` — **exit 1 with EXACTLY two failures**,
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest. **A third is a real
  regression.** This is where the new accounting guard runs.
- `cd folio-go && go vet -tags=matrix ./...` — rc=0. **Load-bearing this story:**
  `fontgen_matrix_test.go` changes package, and this proves the tagged tree still compiles.
- `cd folio-go && FOLIO_FONTGEN_PYTHON=/private/tmp/claude-501/-Users-panitw-Projects-folio/c910e6ef-835c-4b93-9710-f11720b44f3a/scratchpad/fontgen-venv/bin/python go test -count=1 -tags=matrix -run TestShippedFacesReproduceFromUpstream .`
  — verified present: Python 3.12.13 / fontTools 4.63.0, and `/Users/panitw/Projects/folio/.font-sources/`
  holds all three VF sources. Quote the new witness line verbatim.
- `cd folio-designer && npx vitest run` — baseline **58 files / 804 tests**, exit 0; expect 805. Report
  the number, not the adjective.
- `cd folio-designer && npx tsc --noEmit` — rc=0.
- `cd folio-designer && npx oxlint` — rc=0, **exactly 4** `only-export-components` warnings
  (`preview/pdf-viewer.tsx:16,17`; `App.tsx:2896,2903`). A fifth is a regression.
- `cd folio-designer && PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH="$HOME/Library/Caches/ms-playwright/chromium-1217/chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing" npx playwright test e2e/font-embed-boundary.spec.ts --reporter=list`
  — **expect exit 0**, and print the per-family table. ⚠ The env var is **load-bearing, not optional**:
  `chromium-1208` no longer exists on this machine at all, `playwright.config.ts:11` passes the var
  through with no fallback, and `chromium_headless_shell-1243` has no headed counterpart. Complete:
  1217 (336M), 1223 (341M), 1228 (344M). `reuseExistingServer: false` + `npm run build` means **every run
  rebuilds wasm (~2 min) and Go must be on `PATH`** — that is the build, not a hang.
- `cd lint && go build ./...` · `go vet ./...` · `test -z "$(gofmt -l .)"` · `go test -count=1 ./...` —
  **four separate commands, four separate exit codes. Never chain with `&&`;** a chain stopping early has
  already been misread as success on this run. `-count=1` is mandatory: the rules package walks
  directories and Go's test cache does not track `ReadDir`.
  ⚠ **A RED THIS GATE CAUGHT, AND ITS REPAIR — recorded because the gate is why it was found.**
  This `go test` command was **not** in the dispatch (which listed only build/vet/gofmt for `lint`); it
  was added to this spec, and on first run it exited **1** on `TestManifestUpToDate`. Diagnosis, proved
  read-only with `git show 71627a5 --stat` — no worktree, no git write: baseline `71627a5` ("Unpin
  Playwright so the e2e gate can actually start") moved `@playwright/test` to `^1.63.0` in
  `package.json`/`package-lock.json` and **did not regenerate `lint/MANIFEST.md`**, which still recorded
  `1.58.2` and was last touched at `4d2b27e`. Pre-existing, unrelated to this story, and never pushed
  (`origin/main` was still at `b4a5372`). **REPAIRED at `bd1fa44`** by `cd lint && go run
  ./cmd/genmanifest` (3 insertions, 4 deletions to `lint/MANIFEST.md`). Re-measured after it: **rc=0,
  four `ok` packages.** So the expectation for this command is **rc=0** — the red is history, not a
  standing allowance. **No expected-failure line is recorded here on purpose:** a legible expected red
  makes a real defect permanent, and it is the shape that let 16.8's breakage sit for four days.

**Manual checks:**
- If you touch the spec's header, correct `:22-28` — it still describes a 428 KB `chromium-1208` stub
  directory that has been removed entirely.

## Delivery Log

**Delivered 2026-09-05 on the working tree at `e1ce892`** (baseline `d365fd4`; `bd1fa44` and `e1ce892`
landed under this story and touched only `lint/MANIFEST.md` and a decision log — no anchor moved). Four
files: one new untagged Go guard, and three test files repaired. **No production code changed. No
behaviour changed.** Uncommitted by instruction — the orchestrator commits.

⚠ **WHERE THIS STORY'S COVERAGE ENDS, STATED PLAINLY. The `:130` repair — the `AVAILABLE LOCALLY`
subtraction — is asserted in `e2e/font-embed-boundary.spec.ts`, and that suite does not run
per-commit.** `ci.yml:249` runs `test:e2e:compile`, which is `tsc --noEmit`; `playwright test` appears
in no workflow. Measured demonstration: delete `!families.includes(source.family)` from `App.tsx:2604`
and the entire vitest suite, the typecheck and the lint all stay green — only the unexecuted spec
reddens. **That is the same property that made the original defect invisible.** It is partly closed
here by a new vitest case (`App.test.tsx:2355`) asserting the subtraction where CI can see it, but the
browser-level assertion itself remains guarded only by a suite a human must remember to run. Closing
that is DW-193's, not this story's.

**The reproduction half is also unautomated** (DW-200): `TestShippedFacesReproduceFromUpstream` runs in
no workflow — `ci.yml` only compiles the tag, `matrix.yml` name-filters to two other tests. Its pinned
toolchain is one CI does not have and cannot acquire (D-000.12(4)). **This is exactly why the
accounting half was required to be untagged**, and the accounting half does now run per-commit.

**Measured gates, on the final tree, re-measured independently by the builder and again by the
orchestrator:**

| gate | result |
|---|---|
| `folio-go go test -count=1 ./...` | rc=1 — **1952 pass / 2 fail / 5 skip**; the two are `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child. No third. |
| `folio-go/fonts` | **ok**, untagged, in that same run |
| `go vet -tags=matrix ./...` | rc=0 — the `package folio` → `package folio_test` change compiles |
| `gofmt -l .` | empty |
| fontgen matrix, pinned venv | rc=0 — `fontgen: derived and compared 3 of 4 shipped faces (Noto Sans, Noto Sans SC, Noto Sans Thai); 1 accounted static-upstream`, beside fontgen's own unchanged `3 of 3` |
| `npx vitest run` | rc=0 — **58 files / 806 tests** (baseline 804; +2) |
| `npx tsc --noEmit` · `npx oxlint` | rc=0 · rc=0 with exactly 4 `only-export-components` |
| `npx playwright test e2e/font-embed-boundary.spec.ts` | **rc=0, 3 passed (2.9m)** — 30 ADVANCED, 1 UNCHANGED (`Roboto`, carried `Roboto`) |
| `lint`: build · vet · gofmt · `go test -count=1` | rc=0 · rc=0 · rc=0 · rc=0, four separate commands |

**A red this story's own Verification found and did not cause.** `lint`'s `TestManifestUpToDate` failed
on first measurement: baseline `71627a5` unpinned Playwright in `package.json` without regenerating
`lint/MANIFEST.md`, which still recorded `1.58.2`. Attribution settled read-only with
`git show 71627a5 --stat`, reported rather than worked around, and **repaired by the orchestrator at
`bd1fa44`**; re-measured green afterwards. No expected-failure line is recorded for that gate on
purpose — a legible expected red is what let 16.8's breakage sit for four days.

**Three defects were found in this story's own instructions before any of them shipped**, all corrected
by ruling rather than absorbed: the premise that picking the current family *"issues no command"* (it
issues one unconditionally — `App.tsx:2688-2692`; the engine is what declines to mutate, at
`wasm/engine.go:240-246`); a correction message that was dispatched, queued and never delivered, which
a paraphrase then made look received (**D-000.17**); and the e2e partition coupling below.

**The review's best finding, and it was against this story's own subject.** The declared arm partitioned
by *"does the starter declare this family"* and asserted the revision does not move — but the actual
cause of no-move is *"the element already carries this family"*. Those coincide only because the starter
declares exactly one chain and `defaultFontFamily` takes the first in sorted key order; a second declared
catalogue chain sorting earlier would have made the test red over correct behaviour, while its header
claimed the opposite. Triaged **`patch`, not `bad_spec`** — the frozen intent was right, the non-frozen
encoding of it was imprecise, the failure mode was loud rather than silent, and a loopback would have
destroyed 521 verified lines to repair one predicate. The arm now partitions on the carried value,
derived by `defaultFontFamily`'s own rule, and separately asserts every element really was born carrying
it, so the licence for deriving from it is itself measured.

**Sixteen review patches applied**, all verified against the tree before dispatch; five findings rejected
with reasons (including a missing `LICENSE` check already enforced by `lint/internal/manifest`
`ResolveAssets`, and "the new file is untracked" — a commit this agent is forbidden to make). Two
deferrals filed: **DW-200** (fontgen reproduction in no CI workflow — Story 15.2's) and **DW-201** (two
NOTICE parsers of differing strictness, neither citing the other).

**Anchor rot caused and fixed inside this story, which is its own subject in miniature:** correcting the
stale `chromium-1208` header added 12 lines and silently invalidated `accounting_test.go`'s citation of
`font-embed-boundary.spec.ts`, which a patch had *just* corrected. Re-derived to `:130-136` and swept
repo-wide for others (one citation exists; it is right). The spec's own Code Map was also off by one on
`wasm/engine.go` and now reads `:240-246`.

### 2026-09-05 — done

Baseline `d365fd4`; shipped at **`bc321d7`**, "Re-derive what closes over a moving set, instead of
transcribing it". One commit, seven files, no production code and no product behaviour. One review loop.
Triage **16 patched / 2 deferred / 5 rejected / 0 intent_gap / 0 bad_spec**, every route ruled by the
orchestrator. **CI is green on `bc321d7`** — both `Build, vet, and guardrails` and `Cross-target byte
identity` concluded success — which is the first time this epic's work has had any CI signal at all
(DW-171 says nothing in Epic 16 is CI-verified; that remains true of the *e2e and reproduction* halves,
not of the untagged accounting guard, which now runs in the per-commit Go suite).

**Gates re-measured at close, on `bc321d7`, working tree clean before and after except this file and the
tracker.** Each exit code captured on the command's own line (D-000.18 — `$?` is clobbered by any
intervening command, `echo` included, and that produced a false green inside this very story):

| gate | measured at close |
|---|---|
| `folio-designer` `npx vitest run` | rc=0 — **58 files / 806 tests passed**, 0 failed (baseline before this story: 804) |
| `folio-designer` `npx tsc --noEmit` | rc=0 — "No errors found" |
| `folio-go` `go test -count=1 ./...` | rc=1 — **1952 pass / 2 fail / 5 skip**, counted from `-json` events, not from a summary line |
| — the two failures | `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child, in `internal/text`. **No third.** That is the one sanctioned red |
| — `folio-go/fonts` | **ok**, untagged, inside that same run — the accounting guard executes per-commit as designed |
| `lint` `go build ./...` | rc=0 |
| `lint` `go vet ./...` | rc=0 |
| `lint` `test -z "$(gofmt -l .)"` | rc=0 |
| `lint` `go test -count=1 ./...` | rc=0 — four `ok` packages. Four separate commands, four separate exit codes, never chained |

**Not re-run at close, by cadence, not by omission.** The heavy suites are on an **end-of-epic** cadence
and come due at Epic 16's boundary gate: the `matrix` reproduction (`TestShippedFacesReproduceFromUpstream`
on the pinned Python 3.12.13 / fontTools 4.63.0 venv) and Playwright
(`e2e/font-embed-boundary.spec.ts`). Both were run and recorded by the builder on this same tree —
Playwright rc=0, 3 passed, 30 ADVANCED and 1 UNCHANGED; the matrix run rc=0 emitting the new witness
**`fontgen: derived and compared 3 of 4 shipped faces (Noto Sans, Noto Sans SC, Noto Sans Thai);
1 accounted static-upstream`** beside fontgen's own unchanged internal `3 of 3`. Those two lines are the
**builder's** measurement carried forward with attribution, not this closer's; `go vet -tags=matrix` and
`npx oxlint` likewise stand on the builder's record.

**Preserved deliberately: the defect this story committed against itself.** Correcting the stale
`chromium-1208` header added 12 lines and silently rotted a citation in the new accounting guard that a
review patch had *just* fixed. The builder re-derived it and swept the repo for others. A story whose whole
subject is stale premises manufacturing one of its own is worth keeping in the record rather than tidying
away, and it is the cheapest available argument for why the guards it ships derive rather than transcribe.

**The coverage boundary stands as the builder wrote it, and is restated here so no later reader has to
find it.** The `AVAILABLE LOCALLY` subtraction is asserted in a browser suite that **does not run
per-commit** — measured by deleting the filter clause from `App.tsx:2604` and watching vitest, `tsc` and
lint all stay green while only the unexecuted spec reddened. That is the same property that made the
original defect invisible. The new vitest case closes it partly, where CI can see it; the browser-level
half is **DW-193's**, not this story's.

**Neither e2e failure was a product defect.** Both behaviours were correct. The orchestrator's earlier
report that the default typeface could not be chosen was **wrong**, and this entry records that
correction rather than leaving the stronger claim standing — a picked declared family sends its command,
and the engine's stable snapshot is what holds the revision still.

**Deferred, with owners — verified present in `deferred-work.md` at close, not merely in this spec.**
Two were **filed by this story**: **DW-200** (the fontgen reproduction runs in no workflow; pre-existing,
but 16.11 is what made it matter — owner **unassigned, properly Story 15.2's**, still `backlog`) and
**DW-201** (two NOTICE parsers of differing strictness, neither citing the other — owner: the next story
to change a `NOTICE.md` row shape on either side). The browser-level half of the coverage boundary above
is carried by **DW-193**, which this story did **not** file — it was raised by Story 17.6's review,
marked PRE-EXISTING, and is still OPEN and unassigned. Five findings were rejected with reasons,
including "the new file is untracked" — a commit the builder is forbidden to make, and one that has
since happened at `bc321d7`.

**Epic 16 is NOT closed by this story.** `epic-16` stays `in-progress`. The boundary gate closes it, and
still owes a characterisation of the third e2e failure (`browser-native-roundtrip.spec.ts`, "fresh
authored sessions close exactly through admitted Preview and native Folio"): it reproduces standalone and
exceeds its own 300 s budget, but has **not** been distinguished between genuinely hung and merely slower
than that budget. Until it is, the gate has an unmeasured item, not a green one.

## Suggested Review Order

**The accounting guard — the story's core**

- Obligation (a): every shipped face's NOTICE digest and size equal its embedded bytes.
  [`accounting_test.go:430`](../../folio-go/fonts/accounting_test.go#L430)

- Obligation (b): exactly one named route, no fallback arm, no default.
  [`accounting_test.go:494`](../../folio-go/fonts/accounting_test.go#L494)

- The face→directory relation joined BY BYTES, so no hand-written table becomes a fourth authority.
  [`accounting_test.go:267`](../../folio-go/fonts/accounting_test.go#L267)

- `static-upstream` now asserts the SOURCE digest, not just the relation phrase — a route satisfied by a sentence is not a route.
  [`accounting_test.go:541`](../../folio-go/fonts/accounting_test.go#L541)

- Both derivation-statement shapes handled, and the static phrase matched in full so roboto's licence sentence cannot be misread.
  [`accounting_test.go:147`](../../folio-go/fonts/accounting_test.go#L147)

- Every NOTICE row parser throws on 0 or ≥2 matches; an empty extraction is never "no constraint".
  [`accounting_test.go:105`](../../folio-go/fonts/accounting_test.go#L105)

- The fontgen manifest read as text, because the script has no machine-readable mode.
  [`accounting_test.go:388`](../../folio-go/fonts/accounting_test.go#L388)

**The fontgen denominator**

- The denominator moves onto `fonts.Shipped()`, with an independent pin rather than one derived from UPSTREAM.
  [`fontgen_matrix_test.go:239`](../../folio-go/fontgen_matrix_test.go#L239)

- `package folio_test`, forced by the `fonts` → `folio` import cycle.
  [`fontgen_matrix_test.go:48`](../../folio-go/fontgen_matrix_test.go#L48)

- Header corrected: three of the four shipped faces are derived, not all four.
  [`fontgen_matrix_test.go:5`](../../folio-go/fontgen_matrix_test.go#L5)

**The e2e repair — derived from the template, not from a premise**

- The partition: what the element is BORN carrying, by `defaultFontFamily`'s own sorted-first-chain rule.
  [`font-embed-boundary.spec.ts:107`](../../folio-designer/e2e/font-embed-boundary.spec.ts#L107)

- The starter's declared chains read from the file the build serves.
  [`font-embed-boundary.spec.ts:77`](../../folio-designer/e2e/font-embed-boundary.spec.ts#L77)

- Labels state what was measured and drop the causal claim the browser cannot see.
  [`font-embed-boundary.spec.ts:160`](../../folio-designer/e2e/font-embed-boundary.spec.ts#L160)

- The loop: group membership by declaredness, outcome by carried value, both arms non-empty.
  [`font-embed-boundary.spec.ts:319`](../../folio-designer/e2e/font-embed-boundary.spec.ts#L319)

- The permanently-unsatisfiable disclosure assertion deleted, citing 16.10 and naming where the count lives now.
  [`font-embed-boundary.spec.ts:262`](../../folio-designer/e2e/font-embed-boundary.spec.ts#L262)

- The chromium header rewritten as a measurement; the 1208 stub it warned about no longer exists.
  [`font-embed-boundary.spec.ts:22`](../../folio-designer/e2e/font-embed-boundary.spec.ts#L22)

**Designer unit coverage — where CI can actually see it**

- The AD-15 property: the command IS sent; the engine's stable snapshot is what holds everything still.
  [`App.test.tsx:2264`](../../folio-designer/src/App.test.tsx#L2264)

- The subtraction asserted in a suite that runs per-commit, with a derived count and a positive control.
  [`App.test.tsx:2355`](../../folio-designer/src/App.test.tsx#L2355)
