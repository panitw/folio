---
title: 'Story 15.2b: The font-host scan sees the whole tree'
type: 'bugfix'
created: '2026-09-05'
status: 'done'
baseline_commit: '314ad06631f7cff2ee808c15e630ca8bb7776f22'
plan_gate_revision: '7025b340e04d34da97d160e05026145209036f66' # plan gate measured here; 314ad06 changed only the epic's citation line
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-15-context.md'
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs. Rewritten at close, to
describe what actually shipped.*

Two build checks scan the source for banned font-provider addresses. Both chose what to read by
asking version control for its list of known files, which omits anything written but never
registered — and no story in this run registers its own work. So every file a story created was
invisible to both checks. One earlier story added five files, and the checks reported clean over a
tree holding none of them.

Neither script was dishonest: each bounded its verdict to what it had read, and one already named
this very hole in its own opening comment. The fault was in the quoting. What this buys is not a
bigger number — the checks could not tell "I looked and found nothing" from "I never looked at what
you just wrote", and now they can.

Both now read the whole tree bar what is deliberately excluded. That opens one trap: build-generated
files are excluded one name at a time, so a new one would trip a guard nobody touched. A new test
closes it, naming any emission that is not excluded.

Review found three real defects, all fixed — chiefly that the headline property was untested in both
checks, and that a nested repository was skipped in silence: this story's own bug, returning through
its own fix. The manual per-story workaround this replaced now stands down. One unrelated text
failure stays deliberately red; the heavier cross-platform and browser suites run only at the end of
the epic.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** `npm run scan:font-hosts` builds its population from `git ls-files -z`
(`scannedPopulation` in `folio-designer/scripts/forbidden-font-hosts.mjs`) — **tracked files
only**. Because subagents in this run never stage, every file a story creates is untracked for
that story's whole life; Story 12.2 added five files and the scan was green over a tree
containing none of them. **The script is not at fault and must not be "fixed" as though it
were**: its header names *"a host in an untracked file"* as a known hole and every message is
worded to claim only the scanned population. The defect was in the consumer. **What the gate
cannot do today is distinguish "no forbidden host found" from "I did not look at the files you
just wrote", and that collapse — not a bigger number — is what this story buys.**

**Approach:** Widen the **file listing only** to include `git ls-files --others
--exclude-standard` alongside the tracked listing, correct every message that claims the
tracked bound, and close the trap the widening creates — an unignored newly emitted build
artifact now enters the scan the moment it is emitted — with an assertion that the set
`build-wasm.mjs` emits is a subset of the set `.gitignore` ignores.

**And the same widening reaches the sibling guard `scripts/host-font-access.mjs`, narrowly**
(ruled at this plan gate; see **Rulings**). It runs in the same `npm run build` chain, has a
structurally identical tracked-only `scannedPopulation`, and **has already been fooled once by
this exact defect and wrote that down** — its `POPULATION_FLOOR` docblock records a figure it
had to correct because six of its own new files were still untracked when it was measured.
Its refusal message states this story's thesis in the guard's own words — *"an unobtainable
population must never read as all-clear"* — **and it still has the hole.** Fixing only the
measured half would relocate the drift rather than discharge it: the consumer sentence is one
sentence, *"the host scans are clean"*, and it would keep inheriting the identical false green
through the other door.

## Boundaries & Constraints

**Always:**

- **Widen the file listing, nothing else.** `SCANNED_ROOTS` and `SCANNED_EXTENSIONS` are
  untouched **in both scans**. Changing *which directories* are scanned is a different decision
  and is not this story's — it is why `_bmad-output/break-signoff-review-sheet.html`, which
  really does contain a forbidden host, stays out of scope by design.
- **The sibling's scope is exactly three things**: the widened listing, the message corrections,
  and one hermetic red arm. **No floor change and no subset assertion on the sibling.** Where its
  docblock's recorded numbers are stale, they are corrected as **recorded measurements**, which is
  a different act from re-measuring a floor: `POPULATION_FLOOR = 86` does not move, and neither
  does `400`.
- **The second `git` call gets its own refusal, in both scans.** A failing or unobtainable
  `--others` listing must throw, never silently degrade to the tracked list. Each
  `scannedPopulation`'s existing throws exist because *a scan that could not look must never read
  as an all-clear*; the new call inherits that rule rather than being exempt from it. **The
  sibling already spells that principle in its own message and still had the hole — do not let
  the new call be the one place it is not applied.**
- **Every claim of the tracked bound is corrected in the same commit, in both scans.** The
  forbidden-host scan's header hole sentence, its two population throws and its CLI
  `"N tracked source files"` line; and the sibling's two population throws. All assert a bound
  this change makes false. A guard whose honesty is its subject may not ship a stale claim about
  itself.
- **The red-proof is HERMETIC.** A throwaway repository under `os.tmpdir()`: `git init`, one
  file committed, a second left untracked carrying a forbidden host, the scan invoked against
  that root. After this widening an untracked file containing a forbidden host is *exactly what
  the scan is designed to catch*, so planting one inside this repository would redden any
  concurrent story's gate, and an interrupted prover would leave it behind to poison every later
  scan. A committed fixture cannot serve: **the property under test is being untracked**, which
  a committed fixture cannot have.
- **The population floors stay floors.** `POPULATION_FLOOR` remains `400` here and `86` in the
  sibling, and both assertions remain `toBeGreaterThan`. Measured at this plan gate: **no test
  anywhere pins either population to an exact number** (see Design Notes), so the story's "update
  it deliberately" clause is discharged by there being nothing to update — recorded as a
  decision, not as silence.
- **The exemption list is re-checked by measurement and the measurement is recorded in the
  code**, replacing the stale figures in both docblocks — **630** here (recorded: 579) and
  **145** in the sibling (recorded: 129), plus the sibling's now-false cross-citation of this
  scan's ratio.
- **Cite by symbol, not line number**, in every comment and message this story writes.

**Ask First:**

- Changing `SCANNED_ROOTS`, `SCANNED_EXTENSIONS`, or the **value** of either `POPULATION_FLOOR`.
- Widening any guard beyond the two named here (`forbidden-font-hosts.mjs` and
  `host-font-access.mjs`), or growing the sibling's scope past its three ruled items.
- Any edit to `.gitignore` itself, or making a currently ignored artifact tracked.
- **Any finding that either widened population is not clean.** A real occurrence surfaced by the
  widening is the story working, not a bug to sweep — stop and report it.

**Never:**

- **Never run, in this repository, and this applies to every subagent: `git commit`, `git add`,
  `git stash`, `git checkout`, `git reset`, `git revert`, `git restore`, `git push`, `git
  merge`, `git rebase`, `git cherry-pick`, or the creation of a branch.** The human makes every
  commit. Read-only git is fine.
  **Ratified exception, and it is the only one:** `git init`, `git add` and `git commit` are
  permitted **inside a fresh `mkdtempSync` directory outside this repository, removed in a
  `finally`** — never against `/Users/panitw/Projects/folio`. The prohibition is about *this
  repository's history*, which the human controls; it was never about the `git` binary, and the
  shipped suite already builds throwaway repos this way at four sites.
- Never create a file containing a font host anywhere under this repository — not in
  `fixtures/`, not temporarily, not "just to check".
- Never treat the script's bounded claim as dishonesty, and never delete the bound: the runtime,
  dependency and third-party clauses of its header stay true and stay written.
- Never adjust a count, a floor or a recorded measurement to whatever the scan now reports.
- Never add a manifest of generated artifacts as a *third* declaration. The emitter knows what
  it emits and `.gitignore` states what it ignores; the assertion reads both, and inventing a
  hand-maintained list in between just adds a place to drift.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Clean tree, widened | This repository at a clean checkout; 0 untracked non-ignored files | Green. Count is `tracked + untracked`, and the printed sentence no longer calls the population "tracked" | N/A |
| The property being bought | Hermetic temp repo: one committed clean file, one **untracked** file spelling a forbidden host | **RED.** The untracked path appears in `scannedPopulation()`, and the scan throws naming `path:line` | Throws, names file and line |
| Untracked listing unobtainable | The `--others` invocation fails | Throws "could not look" | **Never** falls back to the tracked list |
| Generated artifacts present | `src/generated/` fully populated by a local build | None of the five ignored artifacts enter the population; `pdfjs-assets.ts` is in it only because it is tracked, as it already was | N/A |
| A new emitted artifact, unignored | `build-wasm.mjs` emits `src/generated/<new>.ts` with no `.gitignore` line | Assertion reds, **naming that artifact**, telling the author to ignore it or justify tracking it | Names `pdfjs-assets.ts` as the deliberate tracked exception |
| Declared-only host in an emitted file | `src/generated/font-index.ts` carries `fonts.google.com` in its generated header | Stays out of the population because its ignore line holds — asserted, not assumed | N/A |
| Sibling, clean tree | `scan:host-fonts` over this repository | Green. Baseline **145 files, floor 86**; message no longer calls the population tracked | N/A |
| Sibling, untracked offender | Hermetic temp repo: one committed clean file, one **untracked** file spelling a Local Font Access API | **RED**, naming `path:line` | Throws |
| Sibling, `--others` unobtainable | The second git call fails | Throws "could not look" | **Never** falls back to the tracked list |

</frozen-after-approval>

## Rulings — settled at this plan gate, 2026-09-05

Recorded here because a ruling that exists only in the orchestration channel does not exist.
Quoted, not paraphrased.

**R1 — the sibling is in scope, narrowly.** *"A guard that has already been fooled once by the
defect we are fixing, and that documented being fooled, is not a candidate for deferral."* And:
*"The sibling articulates the principle and still has the hole. That is the most persuasive fact
in the checkpoint."* Scope is the widened listing, the message corrections and one hermetic red
arm — **no floor change, no subset assertion on the sibling.** Its stale recorded figure and its
stale cross-citation of this scan's ratio are corrected, and the spec must *"say that you are
correcting a recorded measurement, not re-measuring a floor."*

**R2 — git write commands in a throwaway temp repo are permitted.** *"The prohibition is about
this repository's history, which is mine to control; it was never about the `git` binary."*
Permitted: `git init`, `git add`, `git commit` **inside a fresh `mkdtempSync` directory outside
this repository, removed in `finally`. Never against `/Users/panitw/Projects/folio`.** Carry the
implementation note — `-c user.name= -c user.email= -c commit.gpgsign=false` — because *"a test
that hangs on someone else's laptop is a test that gets deleted."*

**R3 — `D-000.x` is a per-log namespace, not a global one, and nothing is renumbered.** All three
run logs number their own series from 1; **27 numbers are defined in more than one log,
deliberately.** So the collision is not a misnumbering. *"What is genuinely defective is the bare
cross-log citation"* — inside a log, bare is unambiguous; from a spec in
`implementation-artifacts/` it names two different rules. **Disambiguate cross-log citations by
log and title; never change a number.** The owner corrects `epics.md`'s bare citation separately.

**R4 — the delta arm's repair is upheld.** *"An assertion that cannot fail on the tree it runs
against is the run's dominant defect in its purest form."* `untracked === 1` and `files ===
tracked + untracked` belong in the hermetic repo, where they are falsifiable — **never as a
clean-tree-only delta check.**

**R5 — the epic's stylesheet framing is corrected in the open, not quietly narrowed.** The epic
text *"was written from the emission site rather than the file."* Measured: `runtime-fonts.css`
carries **no** host; the font-index snapshot is the sole carrier. Say so.

## Code Map

Everything below was measured at `7025b34` on 2026-09-05. Anchors are symbols; the line numbers
quoted in the epic text (`:253`, `:8-13`, `:68-75`) were correct when written and will move.

**The scan itself — `folio-designer/scripts/forbidden-font-hosts.mjs`**

- `scannedPopulation(root)` — **the one function to change.** Runs `git -C <root> ls-files -z`,
  filters through `insideTheWalk` then `SCANNED_EXTENSIONS` then an `existsSync`/`isFile` check.
  Three throws, all worded around *"could not look"*: the `execFileSync` catch, `tracked.length
  === 0`, and `files.length === 0` (message contains `carries a scanned extension` — a test
  matches on that phrase, keep it). The last two messages say **"tracked"** and must be reworded.
- `scanForbiddenFontHosts(root, {floor})` returns `{ files, floor, findings }`. If the widening
  adds `tracked`/`untracked` counts to this object (recommended — it is what makes the delta
  assertion falsifiable), **the sidecar must be hand-updated**; see below.
- The CLI block at the bottom prints `"... in ${result.files} tracked source files ..."`. This
  sentence is the guard's public claim and is the most visible thing that goes stale.
- Module header — the paragraph naming *"a host in an untracked file"* as a known hole. That
  clause becomes **false for this scan** and must be amended; the runtime / dependency /
  not-written-here clauses stay.
- `POPULATION_FLOOR = 400`, whose docblock records **"MEASURED AT STORY 8.5: 579 tracked
  files"**. Measured now: **630**. Stale figure, and it is the input the floor's 67–69% ratio was
  derived from.
- `insideTheWalk` already rejects `.git` and `node_modules` segments and anything whose first
  segment is not in `SCANNED_ROOTS` — so untracked files under `_bmad-output/` cannot enter.
- **Reuse:** `blankComments(source, extension)` is exported here and already imported by five
  other tests. Use it in the new subset assertion so a *commented-out* emission is not counted
  as live.

**Its test — `folio-designer/src/forbidden-font-hosts.test.ts`** (the new arm belongs here;
convention is `<script-basename>.test.ts` under `src/`)

- Four existing `it(...)` blocks already build throwaway git repos under `os.tmpdir()` with
  `git init -q` + `git add -A` and clean up in `finally`. **Copy that shape.** They `add` but
  never `commit`, and `git ls-files` counts staged files as tracked — which is exactly why the
  existing positive control *would stay green through a broken widening*, and why AC1 needs its
  own arm.
- `'reads a real population ...'` asserts `result.files > POPULATION_FLOOR` and derives `reached`
  (first path segments) `.toEqual([...SCANNED_ROOTS].sort())` — an **exact set**. It stays green
  only while `insideTheWalk` keeps filtering untracked files outside the roots. Do not weaken it.
- `'throws rather than reporting an empty population it could not obtain'` matches
  `/could not look/` and `/carries a scanned extension/`.

**Type sidecar — `folio-designer/scripts/forbidden-font-hosts.d.mts`**

- `tsc -b` never opens the `.mjs`: `allowJs` is off and no tsconfig includes `scripts/`. The
  `.d.mts` *is* the module to the compiler, resolved from the `.mjs` specifier under
  `moduleResolution: "bundler"`. **Nothing checks the sidecar against the implementation** — a
  missing export is `TS2305` at the importer; a wrong signature is accepted silently and fails at
  runtime. Any new export or changed result shape must be mirrored here by hand.
- `.mts` is in `SCANNED_EXTENSIONS`, so the sidecar is itself scanned.

**The emitters, for the subset assertion**

- `folio-designer/scripts/build-wasm.mjs` — writes into `generatedDir` (`src/generated/`):
  `font-catalogue.ts`, `offline-assets.ts`, `pdfjs-assets.ts`, `runtime-fonts.css`; writes the
  whole `outputDir` (`src/generated/runtime/`); and calls `emitFontIndexModule()` at the end.
  **It cannot be imported by a test** — it runs `go env GOROOT` and the whole build at module
  top level. The assertion must read its *source text*.
- `folio-designer/scripts/build-font-index.mjs` — `emitFontIndexModule()` writes
  `src/generated/font-index.ts`. Must be read too, or `font-index.ts` is missed.
- `.gitignore`'s WASM-build-output block ignores `folio-designer/src/generated/runtime/`,
  `offline-assets.ts`, `runtime-fonts.css`, `font-catalogue.ts`, `font-index.ts` — **five
  entries, one per file, plus the runtime directory**. `pdfjs-assets.ts` has no line and is
  tracked on purpose.

**Measured, so the story does not have to assume it** (`git check-ignore -q` rc / `git ls-files
--error-unmatch` rc, per file):

| Artifact | ignored | tracked |
|---|---|---|
| `src/generated/runtime/folio-engine.wasm` | yes | no |
| `src/generated/offline-assets.ts` | yes | no |
| `src/generated/runtime-fonts.css` | yes | no |
| `src/generated/font-catalogue.ts` | yes | no |
| `src/generated/font-index.ts` | yes | no |
| `src/generated/pdfjs-assets.ts` | **no** | **yes** |

**The one host in the generated tree, and why the ignore line is load-bearing.**
`/usr/bin/grep -rn` for all four scanned hosts across the entire populated `src/generated/`
tree returns **exactly one** occurrence: `fonts.google.com`, in the generated header **comment**
of `font-index.ts`. The scan reads raw text and exempts only a marker written *in code*, so that
file would be a `declared-only` **finding** if it ever entered the population.
`build-font-index.mjs` itself spells the host **zero** times — it composes the comment from
`familyIndexHost`, imported from `src/font-source.ts`, the single declared home. **The generator
passes the scan and its output would not.** `runtime-fonts.css` carries no host at all (its
`url()`s are local), so the epic's precautionary framing of the stylesheet is broader than the
measurement; the font-index snapshot is the real carrier.

**The sibling, IN SCOPE under R1 — `folio-designer/scripts/host-font-access.mjs`**

- `scannedPopulation(root)` — the same `git ls-files -z` walk, the same shape, **three throws
  down to two**: the `execFileSync` catch, `tracked.length === 0`, and `files.length === 0`. The
  last two say **"tracked"**. Its catch message already reads *"an unobtainable population must
  never read as all-clear"* — the thesis, stated by the guard that lacks it.
- `scanHostFontAccess(root, {floor})` → `{ files, floor, findings }`, and
  `assertNoHostFontAccess` above it. Same `tracked`/`untracked` treatment as the other scan.
- `SCANNED_ROOTS = ['folio-designer/src', 'folio-designer/scripts', 'folio-designer/e2e']` —
  **two-segment roots**, unlike the other scan's single segments. Its test derives `reached` with
  `.split('/').slice(0, 2).join('/')` accordingly. Do not "harmonise" them.
- `POPULATION_FLOOR = 86`, docblock **"MEASURED AT STORY 16.2: 129"**. Measured now: **145**
  (`host-font scan: 145 files scanned under … floor 86, 0 occurrences of 4 spellings`, rc 0).
  The docblock also asserts **"86 is 2/3 of 129 (66.7%)"** — at 145 it is **59.3%** — and
  cross-cites this scan as **"floors 400 against a measured 579-600, which is 67-69%"**, now
  400/630 = **63.5%**. **Three stale recorded numbers in one docblock, all corrected as
  measurements; the value 86 does not move** (R1).
- Its docblock is also the evidence for R1: *"that 123 was measured while this story's own six
  new files were still UNTRACKED, so the walk — which reads `git ls-files` — could not see them."*
- `folio-designer/scripts/host-font-access.d.mts` — same hand-mirrored sidecar rule.
  `HostFontScanResult` must gain whatever `FontHostScanResult` gains.
- `folio-designer/src/host-font-access.test.ts` — already has two `mkdtempSync` + `git init -q` +
  `git add -A` arms to copy, a `toBeGreaterThan(POPULATION_FLOOR)` check, and an arm asserting the
  guard scans its own file. **One** new hermetic red arm joins them.
- `host-font-access.mjs` re-exports `blankComments` from `forbidden-font-hosts.mjs` — the two
  modules already share code, so this is not a new coupling.

**Where the scan runs.** `folio-designer/package.json` `build` = `scan:font-hosts &&
scan:host-fonts && build:wasm && tsc -b && vite build && ...`, and CI's `folio-designer` job runs
`npm run build`. **`scan:font-hosts` runs before `build:wasm`**, so in a fresh CI checkout no
generated artifact exists yet and the widened population is unchanged there.

## Tasks & Acceptance

**Execution:**

- [x] `folio-designer/scripts/forbidden-font-hosts.mjs` — in `scannedPopulation`, add a second
      `execFileSync('git', ['-C', root, 'ls-files', '--others', '--exclude-standard', '-z'])`
      and union its result with the tracked listing before the existing filters. Give it its own
      `try`/`catch` throwing a *"could not look"* message naming `--others`; a failure here must
      never degrade to the tracked-only list. Keep the union order-stable and duplicate-free.
- [x] `folio-designer/scripts/forbidden-font-hosts.mjs` — carry the two counts through
      `scanForbiddenFontHosts`'s result (`tracked`, `untracked`, with `files === tracked +
      untracked`) so the delta is a value a test can read, not an inference from a printed string.
- [x] `folio-designer/scripts/forbidden-font-hosts.mjs` — correct every stale claim, in one pass:
      the header's *"a host in an untracked file"* clause (amend, do not delete the bound); both
      population throw messages; and the CLI sentence, which must report the widened population
      and its split rather than calling it "tracked". Re-record `POPULATION_FLOOR`'s docblock
      measurement as **630** with today's date, keeping the value at 400 and saying why.
- [x] `folio-designer/scripts/forbidden-font-hosts.d.mts` — mirror the changed result shape and
      any new export by hand. Nothing checks this file against the implementation.
- [x] `folio-designer/src/forbidden-font-hosts.test.ts` — add the **hermetic red arm**: a
      throwaway `mkdtempSync` repo, `git init -q`, one clean file committed (pass
      `-c user.name=… -c user.email=… -c commit.gpgsign=false`, or the commit fails or hangs on a
      machine with signing configured), a second file left **untracked** spelling a forbidden
      host. Assert (a) the untracked path is in `scannedPopulation(scratch)` — the positive
      control on the population itself, so a globally-excluded temp dir cannot pass as a green,
      (b) `assertNoForbiddenFontHosts` throws naming `path:line`, and (c) `untracked === 1` and
      `files === tracked + untracked`. Remove the tree in `finally`.
- [x] `folio-designer/src/forbidden-font-hosts.test.ts` — add the **exemption measurement** as an
      executable assertion: for each of the five ignored generated artifacts, `git check-ignore`
      reports ignored; for `pdfjs-assets.ts`, not ignored and tracked. Record the measured
      one-host finding in `font-index.ts` beside it as the reason the check exists.
- [x] `folio-designer/src/build-wasm.test.ts` (new) — the **subset assertion**. Read the
      comment-blanked source of `build-wasm.mjs` and `build-font-index.mjs`, extract every
      `src/generated/` path they write, assert the extracted set equals the six known emissions
      (five files plus `runtime/`) so a seventh reds **naming it**, then assert every extracted
      entry is git-ignored **except** `pdfjs-assets.ts`, which is asserted tracked and named in
      the assertion as the deliberate exception. The failure message must tell the author to add
      an ignore line or justify tracking it.
- [x] `.gitignore` — beside the WASM-build-output block, a comment recording that the ignores are
      per-file on purpose (it is what lets `pdfjs-assets.ts` stay tracked), that the widened scan
      makes a new unignored artifact live immediately, and that `src/build-wasm.test.ts` is what
      enforces it. **Comment only — no pattern changes.**
- [x] `folio-designer/scripts/host-font-access.mjs` (R1, narrow) — the same three things and
      nothing more: union `--others --exclude-standard` into `scannedPopulation` with its own
      *"could not look"* throw; carry `tracked`/`untracked` through `scanHostFontAccess`; correct
      the two throw messages that say "tracked". Then correct the **recorded measurements** in
      `POPULATION_FLOOR`'s docblock — 129 → **145**, the "2/3 (66.7%)" fraction → **59.3%**, and
      the cross-citation of the other scan → **400 against 630, 63.5%** — stating in the comment
      that these are corrected recorded measurements and that the floor value **86 is unchanged**.
      **Do not** change `SCANNED_ROOTS`, `SCANNED_EXTENSIONS`, the floor value, or add a subset
      assertion here.
- [x] `folio-designer/scripts/host-font-access.d.mts` — mirror the changed result shape by hand.
- [x] `folio-designer/src/host-font-access.test.ts` — **one** hermetic red arm, same shape as the
      other scan's: throwaway repo, one committed clean file, one untracked file spelling a Local
      Font Access API; assert the untracked path is in `scannedPopulation(scratch)`, that
      `assertNoHostFontAccess` throws naming `path:line`, and that `untracked === 1`. Nothing else
      in this file changes.

**Acceptance Criteria:**

- Given a hermetic throwaway repository with one committed file and one untracked file spelling
  a forbidden host, when the scan runs against that root, then it is **red**, names the untracked
  file and line, and that path is present in `scannedPopulation()` — proved by mutation:
  reverting `scannedPopulation` to the tracked listing alone makes this arm green while every
  pre-existing arm stays green.
- Given the widened population on a clean tree, when the scan runs, then it is green, `files ===
  tracked + untracked`, and the reported sentence no longer describes the population as tracked.
- Given the exemption list, when it is re-checked against the widened population, then every
  generated artifact is measured — not assumed — to be excluded, `pdfjs-assets.ts` is measured
  tracked, and the check is executable rather than prose.
- Given `POPULATION_FLOOR`, when the widening lands, then it is still `400`, still a floor, and
  its recorded measurement is the freshly measured number with the date it was taken.
- Given the set `build-wasm.mjs` and `build-font-index.mjs` emit, when the assertion runs, then
  it proves that set is a subset of the set `.gitignore` ignores, fails naming any new artifact,
  and names `pdfjs-assets.ts` as the deliberate tracked exception rather than merely omitting it
  — red-proved by adding a seventh emission to a copy of the source and confirming the message
  names it.
- Given the sibling `host-font-access.mjs`, when an untracked file in a hermetic repo spells a
  Local Font Access API, then its scan is **red** — red-proved by the same mutation, reverting its
  `scannedPopulation` to the tracked listing alone — and on this repository it stays green at
  **145 files** with `POPULATION_FLOOR` still `86` and its three stale recorded numbers corrected.
- Given the compensating per-story grep put in force by **D-000.21 *"A gate that states its own
  hole honestly"* in `epic-11-14-decision-log.md`** — not the identically numbered D-000.21
  *"Assert on the produced thing"* in `folio-mvp-decision-log.md`, which is a different decision —
  when this story lands, then that grep is stood down, time-boxed to exactly this story, and the
  Delivery Log says so, so it has a named end rather than becoming permanent by habit.

## Spec Change Log

_Empty — no `bad_spec` loopback has occurred._

## Design Notes

**Why the floor does not move, stated as a decision.** Measured at this plan gate with a positive
control: **there is no assertion anywhere in this repository that pins the font-host scan's
population to an exact number.** Both consumers (`src/forbidden-font-hosts.test.ts` and the
sibling's test) use `toBeGreaterThan(POPULATION_FLOOR)`, and the script compares `result.files <
result.floor`. So the story's "any exact count is updated deliberately" clause has no subject.
`400` stays: it exists to catch a walk that collapsed, not to fence in growth, and raising it
because the population grew is the tuning-to-fit this project forbids. What *does* move is the
docblock's recorded measurement, because a count written beside the thing it counts is a claim.

**The delta arm is vacuous in this repository, and that is the finding.** `git ls-files --others
--exclude-standard` returns **0** at `7025b34` — 12.2's files are committed now. So on this tree
the widened count is `630 + 0`, and an assertion phrased as "higher by exactly the number of
untracked files" is `630 === 630`: true, and proof of nothing. **The hermetic repo is the only
place the delta arm is falsifiable**, which is why AC1 carries `untracked === 1` and `files ===
tracked + untracked` rather than leaving them to a clean-tree check. Do not write a
clean-tree-only delta assertion and call the criterion met.

**The justification the subset assertion must carry, in one sentence (R5).** Across the entire
populated `src/generated/` tree there is **exactly one** occurrence of any scanned host —
`fonts.google.com`, in the generated header **comment** of `font-index.ts` — while
`build-font-index.mjs` spells that host **zero** times, composing the comment from
`familyIndexHost` in `src/font-source.ts`. **The generator passes the scan and its output would
not.** It is live only if that ignore line ever lapses, which is precisely what the subset
assertion guards. And the epic's framing of *"the emitted stylesheet"* as a host carrier was
written from the emission site rather than from the file: measured, `runtime-fonts.css` carries
**no** host at all — its `url()`s are local — so the snapshot is the sole carrier. Recorded
openly rather than quietly narrowed.

**Why the emitted set is read from the emitters rather than declared.** A hand-maintained list of
generated artifacts would be a third place for the truth to live, and the thing it is supposed to
catch — someone adding an emission — is exactly the moment they would forget to update it. The
emitter's own source and `git check-ignore` are both authorities that cannot be forgotten. The
extraction must run over **comment-blanked** source: a raw-text scan counts a commented-out
`writeFileSync` as a live emission, and the blanker for that already exists in
`forbidden-font-hosts.mjs`. And the extracted set is checked as a **closed set**, not just fed to
a subset test — a subset assertion over an extraction that silently stopped matching is green
over nothing.

## Verification

Each command's exit code is read from `$?` on its **own line**, never through a pipe (`zsh`:
`${pipestatus[1]}`). No `cd` inside a compound command — it re-roots later relative paths. The
heavy-test cadence is end-of-epic, so **the Go matrix suite and Playwright are not in this
story's set**.

**Commands:**

- `cd folio-designer && npm run scan:font-hosts` — expected exit 0. **Report the population count
  as measured**, and report the tracked/untracked split. Baseline at `7025b34`: `0 occurrence(s)
  in 630 tracked source files ... (floor 400)`.
- `cd folio-designer && npm run scan:host-fonts` — expected exit 0, **separate step, separate exit
  code**. Baseline at `7025b34`: `145 files scanned under folio-designer/src,
  folio-designer/scripts, folio-designer/e2e, floor 86, 0 occurrences of 4 spellings`.
- The **hermetic red-proof**, run as part of the suite below and reported separately by name: it
  must be red with the widening reverted and green with it applied. **Mutate first, confirm the
  arm reds, then apply** — a proof that is red both before and after means the fix is not on the
  path the proof exercises.
- `cd folio-designer && npx vitest run` — the full suite, no filter. Baseline floor: **60 files /
  844 tests**. Report the delivered counts and confirm no pre-existing test regressed.
- `cd folio-designer && npx tsc -b` — expected: no errors, exit 0. **`npx tsc --noEmit`
  type-checks zero files in this repository and is not a measurement** (DW-207).
- `cd folio-designer && npx oxlint` — expected **exactly 4** `only-export-components` warnings,
  freshly measured, 0 errors. A fifth is a regression.
- `cd folio-go && go test -count=1 ./...` — expected **exit 1 with EXACTLY TWO failures**:
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child, the one sanctioned red,
  never to be "fixed". **A third failure is a hard stop — report before triaging.** Count from
  `-json` test-level `Action`; package `ok` lines give no counts.

**Manual checks:**

- `git status --porcelain` after every mutation probe, and the mutated line echoed back to prove
  the edit applied. **A mutation that did not apply produces a green that looks exactly like a
  passing proof.** Restore with an absolute path.
- Confirm no file containing a font host exists anywhere under this repository at the end of the
  run: `/usr/bin/grep -rl` for the four hosts over the working tree, with a positive control.

## Delivery Log — 2026-09-05

**What landed.** Both font-host scans build their population from `git ls-files -z` **unioned with**
`git ls-files --others --exclude-standard -z`. The listing is the only thing that widened:
`SCANNED_ROOTS`, `SCANNED_EXTENSIONS`, `insideTheWalk` and both `POPULATION_FLOOR` values are
untouched, so `_bmad-output/break-signoff-review-sheet.html` stays out of scope exactly as designed.

- `folio-designer/scripts/forbidden-font-hosts.mjs` — `scannedPopulation` now delegates to a new
  `splitPopulation`, which makes **two** git calls, each with **its own** *"could not look"* refusal
  naming the half it could not obtain and stating that it must not degrade to the other listing.
  The union is order-stable (tracked half first, in git's order) and duplicate-free.
  `scanForbiddenFontHosts` returns `{ files, tracked, untracked, floor, findings }`.
- Every claim of the tracked bound corrected in the same commit: the header's *"a host in an
  untracked file"* clause is **amended, not deleted** (the runtime / dependency / not-written-here
  clauses stay written, and the amendment says which clause went false and why); the
  `tracked.length === 0` throw became a combined *"neither tracked nor untracked-but-not-ignored"*
  throw; the `files.length === 0` throw no longer says *"no tracked file"* (the phrase
  `carries a scanned extension`, which a test matches on, is preserved); and the CLI sentence now
  reads `N source files … (T tracked + U untracked-but-not-ignored, floor 400)`.
- `folio-designer/scripts/host-font-access.mjs` — the same three things and nothing more (R1):
  widened listing with its own refusal, `tracked`/`untracked` carried through `scanHostFontAccess`,
  both throw messages corrected. Its CLI line gained the same split. **No floor change and no subset
  assertion here.**
- Both `.d.mts` sidecars mirror the new result shape **by hand**, with the reason written in.

**The measurements, and what moved.** `POPULATION_FLOOR` is **still 400** here and **still 86** in
the sibling; both assertions are still `toBeGreaterThan`. What changed is the **recorded
measurements** written beside them, which is a different act from re-measuring a floor:

| Docblock claim | Was | Now |
|---|---|---|
| `forbidden-font-hosts.mjs` population | 579 (Story 8.5) | **631**, 2026-09-05 |
| `host-font-access.mjs` population | 129 (Story 16.2) | **146**, 2026-09-05 |
| sibling's floor fraction | 86 is 2/3 of 129 (66.7%) | **86 is 59.3% of 146** |
| sibling's cross-citation of this scan | 400 against 579–600, 67–69% | **400 against 631, 63.5%** |

Measured at the plan gate with a positive control and re-confirmed: **no assertion anywhere in this
repository pins either population to an exact number** — both consumers use
`toBeGreaterThan(POPULATION_FLOOR)` — so the *"update any exact count deliberately"* clause had no
subject. Recorded as a decision, not as silence.

**The widening is visibly working on this tree.** At the plan gate `git ls-files --others
--exclude-standard` returned 0, so the widened count would have been `630 + 0`. This story's own new
test file is untracked, so the delivered run reports `631 = 630 tracked + 1 untracked` — the guard
reading a file the story just wrote, which is the entire point.

**The trap the widening opens, and the assertion that closes it.**
`folio-designer/src/build-wasm.test.ts` (new) reads the **comment-blanked** source of
`build-wasm.mjs` and `build-font-index.mjs`, extracts every path they write into `src/generated/`,
and checks the extracted set as a **closed set** of six (`runtime/`, `font-catalogue.ts`,
`offline-assets.ts`, `pdfjs-assets.ts`, `runtime-fonts.css`, `font-index.ts`). It then asserts via
`git check-ignore` that every one is ignored **except** `pdfjs-assets.ts`, which it asserts is
tracked and names in the message as the deliberate exception. No third declaration was added: the
emitter's source and `.gitignore` are the two authorities, read directly. A companion arm reports
any reference that constructs a generated path in a shape the extractor cannot read, with a positive
control, so a subset test can never run green over an extraction that stopped matching.
`.gitignore` gained a **comment only** — no pattern changes — recording why the ignores are per-file,
that an unignored emission is now live immediately, and that this test is what enforces it.

**R5, recorded openly rather than quietly narrowed.** Across the entire populated `src/generated/`
tree there is **exactly one** occurrence of any scanned host: the family-index host, in the generated
header **comment** of `font-index.ts`. `build-font-index.mjs` spells that host **zero** times,
composing the comment from `familyIndexHost` in `src/font-source.ts` — **the generator passes the
scan and its output would not.** The epic's framing of the emitted stylesheet as a carrier was
written from the emission site rather than from the file: measured, `runtime-fonts.css` carries **no
host at all**; its `url()`s are local. The snapshot is the sole carrier. This is written into
`src/build-wasm.test.ts`'s header and asserted (when a local build has populated the tree) in
`forbidden-font-hosts.test.ts`.

**Red proofs — mutated first, confirmed red, then restored.** Every mutation was echoed back and
`git status --porcelain` read afterwards.

| Proof | Mutation | Result |
|---|---|---|
| Forbidden-host hermetic arm | `splitPopulation`'s `untrackedListing` forced to `[]` (the tracked listing alone) | **2 of 18 red** — the new hermetic arm and the new refusal arm. **All 15 pre-existing arms stayed green**, which is exactly why the story needed its own arm: the existing positive control `git add`s its fixture, and `git ls-files` counts staged files as tracked. |
| Sibling hermetic arm | the same mutation in `host-font-access.mjs` | **1 of 8 red** — the new arm only; 7 pre-existing green. |
| Subset assertion | a seventh `writeFileSync(join(generatedDir, 'seventh-artifact.ts'), …)` appended to a **copy** of the source | the real check reds and the message **names `seventh-artifact.ts`**. Run in-test against mutated text, never against the file on disk. |
| Unparsed-reference guard | `join(generatedDir, 'nested', 'deep.ts')` appended to a copy | reported by name rather than silently dropped. |
| Comment direction | a commented-out emission | not counted as live — which is why the source is blanked before extraction. |

The hermetic repos are `mkdtempSync` directories under `os.tmpdir()`, `git init`/`add`/`commit`ed
with `-c user.name= -c user.email= -c commit.gpgsign=false` and removed in `finally` (**R2**).
Nothing was ever committed, staged or planted inside this repository, and no file containing a font
host was created anywhere under it.

**Verification, as measured.**

| Command | Expected | Delivered |
|---|---|---|
| `npm run scan:font-hosts` | exit 0 | **rc 0** — `0 occurrence(s) in 631 source files … (630 tracked + 1 untracked-but-not-ignored, floor 400)` |
| `npm run scan:host-fonts` | exit 0, separate step | **rc 0** — `146 files scanned … (145 tracked + 1 untracked-but-not-ignored), floor 86, 0 occurrences of 4 spellings` |
| `npx vitest run` | no regression; baseline 60 files / 844 tests | **rc 0 — 61 files / 853 tests, all passing** (+1 file, +9 tests; no pre-existing test changed behaviour) |
| `npx tsc -b` | no errors | **rc 0** |
| `npx oxlint` | exactly 4 `only-export-components` warnings, 0 errors | **rc 0, exactly 4** |
| `cd folio-go && go test -count=1 ./...` | exit 1 with EXACTLY TWO failures | **rc 1, exactly two**: `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child — the one sanctioned red. No third failure. |

Manual sweep at end of run: `/usr/bin/grep -rl` for all four hosts over the working tree returns only
pre-existing carriers — the scanner, its test, `src/font-source.ts` and its test, the gitignored
`src/generated/font-index.ts`, and gitignored `dist/` output — plus the known `_bmad-output/` and
`docs/` occurrences that `SCANNED_ROOTS` deliberately excludes. Positive control fired (a scratch
file outside the repository). Nothing new was left behind.

**Neither widened population is anything but clean.** No real occurrence was surfaced by the
widening, so there is nothing to stop and report under the *Ask First* clause.

**The compensating per-story grep is STOOD DOWN, effective this story.** `epic-11-14-decision-log.md`'s
**D-000.21 — *"A gate that states its own hole honestly"*** — not the identically numbered D-000.21
*"Assert on the produced thing"* in `folio-mvp-decision-log.md`, which is a different decision
(**R3**: `D-000.x` is a per-log namespace, 27 numbers are defined in more than one log, and nothing is
renumbered; cross-log citations are disambiguated by log and title) — put an interim manual grep over
each story's new files in force, explicitly *"time-boxed to 15.2b landing, so it has a named end
rather than becoming permanent by habit."* That end is here: `scan:font-hosts` and `scan:host-fonts`
now read every non-ignored file in the tree, tracked or not, so the population the manual grep was
compensating for no longer exists. **No further story needs to run it.** The bound it compensated for
is gone; the bounds that remain — a host assembled at runtime, a host inside a dependency, a request
made by something this repository did not write — are unchanged, still true and still written into
the guard.

**Out of scope by ruling, and left alone deliberately:** `.gitignore` patterns (comment only);
`SCANNED_ROOTS` / `SCANNED_EXTENSIONS`; both floor values; any guard beyond the two named; and
`epics.md`'s bare cross-log citation, which **R3** assigns to the owner separately.

### 2026-09-05 — done

Baseline `314ad06`; shipped at **`3c4abcb`**, *"Let the host scans read the files a story just
wrote"*, one commit, 12 files, on `main` and already pushed. Closed by the story closer; the section
above is the builder's own delivery record and is left as written — this entry adds what it could not
know, and corrects two of its figures that the commit itself made stale.

**Neither script was lying, and that is the whole point of the story.** Both bounded every claim they
made to the population they had actually read, and the forbidden-host scanner went further: it named
*"a host in an untracked file"* as a known hole **in its own module header**. Nothing in either guard
was dishonest. **The defect was in the consumer** — in the one sentence *"the host scans are clean"*,
quoted downstream without the bound that made it true. A bounded claim becomes a false claim the
moment somebody drops its bound, and the guard cannot stop that from the inside.

**What the story bought is not a bigger number.** The population moved by one file. The property is
that **"no forbidden host found" and "I did not look at the files you just wrote" stopped being the
same output.** Before this, those two states were indistinguishable at the call site; the gate could
collapse *all-clear* into *couldn't-look* and print the same green. That collapse is what closed. Any
count in this record is incidental to it.

**Why the sibling could not be deferred.** `host-font-access.mjs` runs in the same `npm run build`
chain and had the structurally identical tracked-only population. Two facts made deferral untenable
rather than merely untidy. Its `POPULATION_FLOOR` docblock **records having been misled by this exact
bias** — a figure it had to correct because six of that story's own new files were still untracked
when it was measured, written down in the guard's own comment. And its refusal message **already
states this story's thesis in the guard's own words**: *"an unobtainable population must never read
as all-clear."* A guard that has articulated the principle, and written down being fooled by its
absence, and still has the hole, is not a deferral candidate — fixing only the measured half would
have relocated the drift rather than discharged it, because the consumer sentence would keep
inheriting the identical false green through the other door.

**The three things the review caught.** All three were verified before being acted on, and all three
are patched in `3c4abcb`.

1. **The headline property was not actually tested — in either scanner.** Both refusal arms invoked
   the population builder against a non-repository path, but **the tracked call throws first**, so
   the `--others` refusal was never on the path the arm exercised; the tests then fell back to
   grepping the scanner's own source text. A catch that silently swallowed the untracked listing
   would have left both arms green. Fixed with a `PATH` git shim that fails **only** on `--others`,
   red-proved in both suites. This is the sharpest finding in the story: the arm written to prove the
   new refusal could not have failed if the new refusal had been absent.
2. **An untracked nested repository is listed as a single directory entry and was dropped in
   silence.** `git ls-files --others` does not recurse into a nested repository; it emits one
   directory entry, which then failed the `isFile` filter and vanished — a whole subtree unread while
   the guard claimed the tree. **That is this story's own defect re-entering through its own
   widening**, and it would have shipped inside the fix for it. Both scanners now refuse on such an
   entry rather than skipping it.
3. **`build-wasm.test.ts` would have reddened on any fresh clone.** `git check-ignore` returns 1 for
   a directory pattern when the directory does not exist, so the ignore probe for the generated
   `runtime/` directory fails before a build has created it. Local and CI runs survived only
   because `build:wasm` runs ahead of `vitest` in both. Fixed at the probe.

**The compensating per-story font-host grep STANDS DOWN, effective this story, and the reason is on
the record.** `epic-11-14-decision-log.md`'s **D-000.21 — *"A gate that states its own hole
honestly"*** (**not** `folio-mvp-decision-log.md`'s identically numbered D-000.21 *"Assert on the
produced thing"*, a different rule; per **R3** `D-000.x` is a per-log namespace) imposed a manual
grep over each story's new files, explicitly time-boxed to *"15.2b landing, so it has a named end
rather than becoming permanent by habit."* That end is here. **The reason it stands down is that the
scanned population now includes untracked-but-not-ignored files**, so the gap the manual grep was
compensating for no longer exists — not because anyone tired of running it. The bounds that remain —
a host assembled at runtime, a host inside a dependency, a request from code this repository did not
write — are unchanged, still true, and still written into the guard. No further story needs the grep.

**Two figures in the section above went stale when the commit landed, and are corrected here rather
than edited there.** Both were correct when the builder measured them, at a moment when this story's
own new test file was still untracked:

| Claim above | Measured then (pre-commit) | Measured at close on `3c4abcb` |
|---|---|---|
| `scan:font-hosts` split | `631 = 630 tracked + 1 untracked` | **`631 = 631 tracked + 0 untracked`** |
| `scan:host-fonts` split | `146 = 145 tracked + 1 untracked` | **`146 = 146 tracked + 0 untracked`** |
| `vitest` | 61 files / **853** tests | **61 files / 866 tests** (+13 from the review patches) |

**The changed split is not a regression — it is the proof the widening works.** `build-wasm.test.ts`
was the `+1 untracked` before the commit and is the `+1 tracked` after it; the *total* is 631 in both
readings, and it is 631 in both readings **only because the widened listing counted the file while it
was unstaged.** A tracked-only scan would have reported 630 then and 631 now. `git ls-files --others
--exclude-standard | wc -l` returns **0** on this tree, which is why the untracked half is now empty.

**Gates re-measured at close on `3c4abcb`**, tree clean before the run, each exit code captured on the
command's **own** line (`$?` is clobbered by any intervening command, `echo` included), and no `cd`
inside a compound command:

| gate | measured at close |
|---|---|
| `folio-designer` `npm run scan:font-hosts` | **rc 0** — `0 occurrence(s) in 631 source files under .. (631 tracked + 0 untracked-but-not-ignored, floor 400)` |
| `folio-designer` `npm run scan:host-fonts` | **rc 0** — `146 files scanned … (146 tracked + 0 untracked-but-not-ignored), floor 86, 0 occurrences of 4 spellings` |
| `folio-designer` `npx vitest run` | **rc 0 — 61 files / 866 tests passed**, 0 failed |
| `folio-designer` `npx tsc -b` | **rc 0**, no output |
| `folio-designer` `npx tsc -b --force` | **rc 0**, no output — run because plain `-b` is incremental and can exit 0 having checked nothing |
| `folio-designer` `npx oxlint` | **rc 0 — exactly 4** `only-export-components` warnings, 0 errors, freshly counted (not carried forward) |
| `folio-go` `go test -count=1 ./...` | **rc 1 — exactly two failures**: `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child, the one sanctioned red. **No third.** |
| `folio-go` `gofmt -l .` | **rc 0**, zero lines |
| `folio-go` `go vet ./...` | **rc 0** |
| `lint` `go build ./...` / `go vet ./...` / `gofmt -l .` / `go test -count=1 ./...` | **rc 0 / rc 0 / rc 0, zero lines / rc 0** (4 packages ok) |

**Not run, and deliberately so:** the Go cross-target matrix suite and the Playwright browser suite.
This run's heavy-test cadence is **end of epic**; they come due at Epic 15's catch-up run, before
`epic-15` may be marked `done`. Nothing in this story touched engine output or browser behaviour.

**Manual checks re-run at close.** `/usr/bin/grep -rlE` for all four scanned spellings across the
working tree (excluding `node_modules`, `.git`, `dist`) names **only pre-existing carriers** — the
scanner, its test, `src/font-source.ts` and its test, the gitignored `src/generated/font-index.ts`,
and the `_bmad-output/`, `docs/` and UX-mockup occurrences that `SCANNED_ROOTS` excludes by design.
**A positive control fired** on a scratch file outside the repository. No file this story created
carries a host. `git ls-files --others --exclude-standard` returns 0, so nothing was left behind.

**The frozen block is byte-identical.** Lines 34–150, `<frozen-after-approval>` through
`</frozen-after-approval>`, 9195 bytes, **sha256
`90e04a0a04f08e4f2b7c83890b09462799879ff6b639cdddef264f722e8eaf03`** before this entry was written
and the same sha256 after. Only the plain-terms opener — which sits **outside** the block precisely
so it can be rewritten after approval — and this Delivery Log entry were changed in this file.

**Review-order anchors re-checked at close.** All **17** `file#Lnnn` links in *Suggested Review Order*
were resolved by printing the cited line at `3c4abcb`; every one lands on the construct its caption
names. No anchor rotted.

**Deferrals.** Three, all filed by the builder and already in `3c4abcb`, not re-priced here:
[[DW-213]] (neither host scan's CLI block is exercised by any test, and this story rewrote both
sentences), [[DW-214]] (nothing detects a `.gitignore` line for an emission that no longer exists),
and [[DW-215]] (the two scanners now duplicate their whole population builder, three of its
behaviours unverified). **DW-215 is the story's own edge, not pre-existing** — the widening is what
doubled the duplicated surface — and DW-213 is sharpened by it, because this story rewrote exactly
the untested sentences.

**Triage route counts are not recorded anywhere in this story's record, and none is invented here.**
`review_loop_iteration` is `0` and the spec has no findings section; the three findings above are
reconstructed from the commit message, where they are stated in full. What can be said as measured:
**3 findings patched, 3 items deferred with DW numbers, 0 `intent_gap`, 0 `bad_spec`** (the *Spec
Change Log* is empty, and says so). **No rejection tally exists** — so none is quoted, rather than a
zero being implied.

**Out of scope by ruling and left alone:** `.gitignore` patterns (comment only); `SCANNED_ROOTS` and
`SCANNED_EXTENSIONS`; both floor values, which did not move; any guard beyond the two named; and
`epics.md`'s bare cross-log citation, which **R3** assigns to the owner separately — that one has
since been corrected at `314ad06`, this story's baseline.

## Suggested Review Order

**The widening itself — one function, twice**

- The whole change in one place: two git listings, each with its own refusal.
  [`forbidden-font-hosts.mjs:300`](../../folio-designer/scripts/forbidden-font-hosts.mjs#L300)

- The second call. Everything else follows from this line existing.
  [`forbidden-font-hosts.mjs:315`](../../folio-designer/scripts/forbidden-font-hosts.mjs#L315)

- A nested repo is listed as one directory and never recursed — so it refuses.
  [`forbidden-font-hosts.mjs:339`](../../folio-designer/scripts/forbidden-font-hosts.mjs#L339)

- Untracked files vanish mid-scan; ENOENT skips rather than aborting the gate.
  [`forbidden-font-hosts.mjs:424`](../../folio-designer/scripts/forbidden-font-hosts.mjs#L424)

- The structural twin, ruled in because it had already been fooled once.
  [`host-font-access.mjs:205`](../../folio-designer/scripts/host-font-access.mjs#L205)

**The claims the guards make about themselves**

- Floor unchanged at 400; only the recorded measurement moved.
  [`forbidden-font-hosts.mjs:408`](../../folio-designer/scripts/forbidden-font-hosts.mjs#L408)

- Three stale recorded numbers corrected; the floor value 86 deliberately did not move.
  [`host-font-access.mjs:122`](../../folio-designer/scripts/host-font-access.mjs#L122)

**The trap the widening creates**

- Emitted set read from the emitters, never declared a third time.
  [`build-wasm.test.ts:63`](../../folio-designer/src/build-wasm.test.ts#L63)

- Directory emissions need the trailing slash, or check-ignore lies when absent.
  [`build-wasm.test.ts:231`](../../folio-designer/src/build-wasm.test.ts#L231)

- Detects an emission the extractor cannot read, so the closed set cannot go blind.
  [`build-wasm.test.ts:168`](../../folio-designer/src/build-wasm.test.ts#L168)

- Per-file ignores are deliberate; the comment explains, the test enforces.
  [`.gitignore:65`](../../.gitignore#L65)

**The proofs — where the property is actually falsifiable**

- The property the story buys, provable only in a hermetic repo.
  [`forbidden-font-hosts.test.ts:395`](../../folio-designer/src/forbidden-font-hosts.test.ts#L395)

- Fails only the --others call; the old arm never reached it.
  [`forbidden-font-hosts.test.ts:471`](../../folio-designer/src/forbidden-font-hosts.test.ts#L471)

- Proves --exclude-standard behaviourally, which keeps build output out.
  [`forbidden-font-hosts.test.ts:589`](../../folio-designer/src/forbidden-font-hosts.test.ts#L589)

- The sibling's three mirrored arms, added because its coverage was asymmetric.
  [`host-font-access.test.ts:320`](../../folio-designer/src/host-font-access.test.ts#L320)

- A seventh emission reds naming itself.
  [`build-wasm.test.ts:398`](../../folio-designer/src/build-wasm.test.ts#L398)

**Peripherals**

- Result shape mirrored by hand; nothing checks a sidecar against its implementation.
  [`forbidden-font-hosts.d.mts:1`](../../folio-designer/scripts/forbidden-font-hosts.d.mts#L1)
