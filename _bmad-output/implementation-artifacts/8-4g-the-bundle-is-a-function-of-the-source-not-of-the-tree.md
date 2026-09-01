---
title: 'Story 8.4g: The bundle is a function of the source, not of the tree'
type: 'bugfix'
created: '2026-09-01'
status: 'done'
baseline_revision: '873757f038f66491aa431dc2b1f015ed42132e42'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: ['oversized']
deferred:
  - summary: >-
      The guard proves the absence of one known cause (VCS needles), but nothing in the repo
      compares two builds' digests, so a second tree-dependent input would leave every gate green
      while the story's headline claim is false.
    evidence: |-
      AC2's property is sha256(build_clean) == sha256(build_stray). Every automated artifact added
      here instead tests for the absence of four ASCII needles in the raw wasm. This is not
      hypothetical: DW-105, filed by this same story, records a measured input of exactly that kind
      (the checkout path). Raised independently by the blind-hunter and intent-alignment layers.
    location: >-
      folio-designer/scripts/wasm-vcs-stamp.mjs
    severity: medium
  - summary: >-
      Both red-proofs exist only as prose and cannot be re-run; the repo's own redProof harness in
      verify-offline-release.mjs is the established shape for exactly this and gained nothing.
    evidence: |-
      verify-offline-release.mjs:126 redProof(name, mutate, expected) already backs stale-wasm-byte,
      s1-total-mismatch, s1-delivery-fiction and dictionary-witness-mismatch, run by
      `npm run verify:offline:red`. The guard's red-proof was performed by hand twice this dispatch
      and recorded in prose, so it will rot silently. Not done here because Story 8.4f is about to
      land its own assertions in that same file and owns that surface.
    location: >-
      folio-designer/scripts/verify-offline-release.mjs:126
    severity: medium
  - summary: >-
      Story 8.4f's spec still carries the 2,203-byte two-builds-at-one-commit observation as a live
      unexplained drift, on a "clean tree" premise this story measured to be false.
    evidence: |-
      8-4f-a-bound-nobody-can-cross-silently.md:236 and :609. D-8.5.7 supersedes the explanation and
      this dispatch measured s1VisibleBytes identical (12,425,468) across a clean tree and a stray
      untracked file, but 8.4f is still backlog and nothing annotates the stale figure in its spec.
    location: >-
      _bmad-output/implementation-artifacts/8-4f-a-bound-nobody-can-cross-silently.md:236
    severity: low
  - summary: >-
      The workflow's step-1 regeneration of epic-8-context.md dropped scope constraints that the
      unbuilt Stories 8.5 and 8.6 were to be built against.
    evidence: |-
      Dropped in this dispatch's regeneration: the out-of-scope-by-decision list (bold/italic
      meaning, synthetic emboldening/obliquing, variable-font axes, live font services, enumerating
      host-installed fonts), the catalogue procurement rules (static, single-instance, prepared
      ahead of the build), and the chain-entry shape contract. Measured: every one retains an
      authority in epics.md, the PRD or the architecture spine, and the file declares itself a
      regenerable cache whose compile prompt instructs aggressive scoping — so this is designed
      lossiness over intact sources, not lost content. Recorded because the next regeneration may
      or may not restore them and 8.5/8.6 have not been built.
    location: >-
      _bmad-output/implementation-artifacts/epic-8-context.md
    severity: low
  - summary: >-
      The ambient environment remains an unrecorded input to the engine wasm build.
    evidence: |-
      build-wasm.mjs still passes env: { ...process.env, GOOS, GOARCH }, forwarding GOFLAGS,
      GOEXPERIMENT, CGO_ENABLED and toolchain selection. DW-105 records only the checkout path, so
      the set of inputs that are NOT the source is documented incompletely.
    location: >-
      folio-designer/scripts/build-wasm.mjs:44
    severity: low
---

<intent-contract>

## Intent

**Problem:** `folio-designer/scripts/build-wasm.mjs` invokes `go build` with no `-buildvcs` flag, so Go's default stamps `vcs.revision`, `vcs.time` and `vcs.modified` into the engine wasm — and Go derives `vcs.modified` from `git status`, where **a single untracked file is enough**. The bundle is therefore a function of the working tree, not of the source, and every byte figure this epic records is untrustworthy: this pipeline itself writes untracked files (halt files, result files, specs) between measurements.

**Approach:** Pass `-buildvcs=false` to the engine wasm build, add an assertion that fails if a VCS stamp ever returns to the emitted binary, and re-measure the discriminator to show the defect is actually closed.

## Boundaries & Constraints

**Always:**
- The engine wasm build passes `-buildvcs=false`.
- The emitted wasm carries **no** Go build-info VCS stamp — no `vcs.revision`, `vcs.time` or `vcs.modified` setting.
- The assertion is **red-proved**: with the flag removed, it must actually fail. A red-proof that is really commit ordering is not a red-proof.
- The fix is **re-measured**: a build on a clean tree and a build with one stray untracked file must produce the **same** wasm sha256. Both digests are reported. This is not optional (D-8.5.7).
- Every byte figure is recorded with its exact invocation, commit **and tree state**.
- Any probe file created is removed, and `git status --porcelain` is verified empty afterwards.
- The provenance loss is recorded as a **deliberate trade, not a free win**.

**Block If:**
- Any golden PDF digest moves. The wasm changing is expected; a golden moving is not — it would mean the stamp reached rendered output and would reopen D-8.5.7's conclusion that this is confined to the designer bundle. HALT.
- `TestShippedFacesReproduceFromUpstream` fails **with** `FOLIO_FONTGEN_PYTHON` set to the repo's `.fontgen-venv` interpreter. HALT.
- The re-measurement does not agree — i.e. clean and stray-file builds still differ after the fix. HALT rather than reporting the fix as landed.

**Never:**
- Never touch `maximumCacheAssets` or the S1 row pinning (Story 8.4f owns that bound).
- Never add, remove or rename an asset; no catalogue work (Story 8.5).
- Never set a byte threshold (Story 8.4d owns it).
- Never touch `folio-go/fonts/`, the engine `FontSet`, or any chrome token.
- Never add `-trimpath`. DW-100 lists it as optional; AC1 names exactly one flag, and a second variable in the same measurement would make the re-measurement unattributable.
- Never rewrite decision-log history. D-8.5.7 already supersedes D-8.4.29 in its own terms.
- Never push, never branch, never `git add -A`/`git add .`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Clean tree build | `git status --porcelain` empty | wasm emitted, zero `vcs.` build-info settings | No error expected |
| Stray untracked file | one untracked file present | wasm **byte-identical** to the clean-tree build | No error expected |
| Modified tracked file | a tracked file modified | wasm byte-identical to the clean-tree build | No error expected |
| Stamp returns | flag dropped from the build invocation | build **fails**, naming the stamp settings found | Throws, non-zero exit; wasm not published |
| Detector applied to a stamped binary | bytes containing `vcs.revision=` | reported as stamped | n/a — predicate returns the findings |
| Detector applied to a clean binary | bytes with no build-info VCS settings | reported as clean | n/a |

</intent-contract>

## Code Map

**Every anchor re-measured at `873757f`, this dispatch's HEAD.** Line anchors are navigation aids (D-8.4.13); assertions are named by what they assert.

### The one-line fix site

- `folio-designer/scripts/build-wasm.mjs`
  - `:18-22` the engine build: `execFileSync('go', ['build', '-o', wasmPath, './wasm/cmd/engine'], { cwd: join(repoRoot, 'folio-go'), env: { ...process.env, GOOS: 'js', GOARCH: 'wasm' }, stdio: 'inherit' })`. **No `-buildvcs` flag today, and no comment about determinism anywhere in the file.**
  - `:17` `wasmPath` = `src/generated/runtime/folio-engine.wasm` — the raw build output.
  - `:45-53` `fingerprint()` content-addresses each asset to `<stem>.<digest20><ext>`; `:66` produces `folio-engine.<digest20>.wasm`.
  - `:76` `rmSync(wasmPath, { force: true })` — **the raw wasm is deleted after fingerprinting**, so any check on the raw file must run before line 76.
- `folio-designer/package.json` `scripts` — `build:wasm` is a dependency of **`typecheck`, `test` and `build`**. A check inside `build-wasm.mjs` therefore runs in every designer gate.

### Where a test can live

- `folio-designer/vite.config.ts:33` — `test.include` is `['src/**/*.test.{ts,tsx}', 'scripts/**/*.test.mjs']`. **`scripts/*.test.mjs` is already a discovered suite** (`offline-release-contract.test.mjs`, `offline-service-worker-template.test.mjs`). `npm test` runs `build:wasm` first, so the generated wasm exists when the suite runs.

### The release verifier and its red-proof harness (reference; reuse the shape, not the file)

- `folio-designer/scripts/verify-offline-release.mjs:126` `redProof(name, mutate, expected)` — mutates built `dist/` output and requires verification to fail; `:157-161` the `s1-*` proofs are the pattern the epic text cites.
- `:97` already asserts Brotli sidecars are deterministic — unrelated to the wasm stamp.

### Read-only evidence — nothing in shipped code claims the drift

Grepped `buildvcs`, `vcs.revision`, `vcs.modified`, `vcs.time`, `commit stamp`, `s1VisibleBytes` across the repo excluding `.git`, `node_modules`, `dist`, `_bmad-output`: **no test, comment or product doc attributes byte drift to a commit stamp.** `README.md:123` and `folio-go/README.md:313` caveat reproducibility "on a given build toolchain" — that is about **rendered PDF bytes**, not the binary, and needs no correction.

### The record surface

- `_bmad-output/implementation-artifacts/deferred-work.md`
  - `:5092` **DW-100, OPEN, owner Story 8.4d** — "`s1VisibleBytes` is not yet a reproducible figure".
  - `:5119-5142` its Story 8.4c UPDATE states the mechanism as "the commit SHA, **the commit timestamp** and the working tree's dirty flag". **The timestamp half is refuted by D-8.5.7** — `vcs.time` is the timestamp of the *revision*, not of the build.
  - `:5140` its stated discharge condition is verbatim "Build the wasm with **`-buildvcs=false`**" — **this story satisfies it.**
  - `:5144` a **second, independent** defect in the same figure (no IBM Plex row; blind to 490,280 bytes of fonts) belongs to Story 8.4d. **So this story amends DW-100; it does not close it.**
- `_bmad-output/implementation-artifacts/epic-7-8-decision-log.md:4348` — D-8.5.7, the binding ruling. **Read-only.**
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — `:411` AD-21 (byte identity of **rendered output** across four targets, never of the compiled binary); `:423` AD-22 ("`go.mod` pins an exact `toolchain` directive and CI uses that version only"), which is what makes the provenance trade acceptable. **Read-only.**

## Tasks & Acceptance

**Execution:**

1. `folio-designer/scripts/wasm-vcs-stamp.mjs` (new) -- export a detector that returns the Go build-info VCS settings found in a wasm byte buffer (`vcs.revision`, `vcs.time`, `vcs.modified`), plus a thrower that fails naming what it found -- a separate module because `build-wasm.mjs` has top-level side effects and cannot be imported by a test without triggering a build.
2. `folio-designer/scripts/build-wasm.mjs` -- add `-buildvcs=false` to the `go build` argument list at `:18`, and immediately after the build (and **before** the `rmSync` at `:76`) apply the thrower to the freshly built `wasmPath` -- the flag is enforced at its point of use, in every gate that runs `build:wasm`. Comment must state the provenance trade and its ground (AD-22 pins the toolchain; the release manifest already carries `releaseId`/`pageId` from asset hashes).
3. `folio-designer/scripts/wasm-vcs-stamp.test.mjs` (new) -- unit-test the detector **both ways** on synthetic buffers (a buffer carrying `vcs.revision=`/`vcs.modified=` is reported stamped; one without is reported clean), **and** tie it to the real population by reading the generated `src/generated/runtime/folio-engine.*.wasm` and asserting it is clean -- a literal-driven unit test proves the predicate, never that it admits what actually ships.
4. `_bmad-output/implementation-artifacts/deferred-work.md` -- amend DW-100: correct the mechanism (drop the refuted "commit timestamp"; the operative input is `vcs.modified`, derived from `git status`, where an untracked file suffices), record that the `-buildvcs=false` discharge condition is satisfied by this story with the commit, and state plainly that DW-100 stays **OPEN** on its second half (the missing IBM Plex row), which is Story 8.4d's. Do not close it.
5. `_bmad-output/implementation-artifacts/deferred-work.md` -- add a new deferred entry recording the **newly measured residual**: the emitted wasm is still a function of the **checkout path**. Same commit, same flags, both unstamped, different directory → different bytes (measured: `c6018cac…` in a linked worktree vs `ff324971…` in the main checkout). Consequence: a baseline measured in a detached worktree is not comparable to one measured in the main checkout, and a linked worktree silently gets **no** VCS stamp at all because its `.git` is a file. `-trimpath` is the candidate remedy and is explicitly out of scope here.

**Acceptance Criteria:**

- Given the engine wasm build, when it runs, then it passes `-buildvcs=false` and the emitted wasm contains **zero** Go build-info VCS settings.
- Given the fix, when a build on a **clean tree** and a build with **one stray untracked file** are compared, then their wasm sha256 digests are **identical**, and both digests are reported with their invocation and tree state.
- Given the flag is removed from the build invocation, when `npm run build:wasm` runs, then it **fails** naming the VCS settings it found — the assertion is red-proved, not assumed.
- Given the loss of `vcs.revision` from the binary, when the trade is recorded, then it is stated as a deliberate trade with its ground (release manifest carries `releaseId`/`pageId` from asset hashes; AD-22 pins the toolchain) — stated, not assumed.
- Given the corpus, when the story is complete, then all 23 golden PDF digests are **unmoved** and `TestCrossTargetByteIdentity` passes — the stamp never reached rendered output.
- Given any probe file created during measurement, when the story is complete, then it is removed and `git status --porcelain` is empty.

## Spec Change Log

**Build dispatch, HEAD `873757f038f66491aa431dc2b1f015ed42132e42` — the same commit at which every
Code Map anchor was measured, so there is no drift to record.** No `bad_spec` loopback occurred and
`review_loop_iteration` stayed `0`; this entry is factual bookkeeping, not an amendment. The
`<intent-contract>` block was never edited: 3675 bytes, md5 `961a82c22f68b8299401f07db0ea8c9d`,
re-verified byte-identical after implementation, after review patching, and at finalize.

**One expectation in `## Verification` was stale as written and is corrected here rather than in the
contract:** it predicted a Go baseline of 1811 pass / 2 fail / 5 skip. The true baseline at
`873757f` is **1815 / 2 / 5** — Story 8.4e added four Go tests at `d055f62`. The spec said "to be
re-measured", and it was. No Go file is touched by this story, so baseline and post-change counts are
necessarily the same figure.

## Review Triage Log

### 2026-09-01 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 5: (high 0, medium 1, low 4)
- defer: 5: (high 0, medium 2, low 3)
- reject: 19: (high 0, medium 0, low 19)
- addressed_findings:
  - `[medium]` `[patch]` **DW-105's path-dependence claim rested on a confound, not a control.** The linked-worktree arm differs from the main checkout in *two* ways at once — the absolute path, and `.git` being a file rather than a directory, which makes `-buildvcs=auto` silently omit the stamp entirely. So the original evidence could not distinguish "Go embeds absolute source paths" from "a worktree gets no stamp". Fixed by measuring a positive control: a full copy of the checkout at one commit, `.git` a real directory, identical tree state, built with the same flag. MAIN `ff324971…d90f732` vs COPY — different, and the mechanism confirmed directly at 87 embedded absolute source paths per binary, each naming its own root. The patch agent re-ran the control independently and got a *third* digest from its own copy path, which is the finding rather than a disagreement: three paths, three digests, one commit, MAIN unmoved. DW-105 now records both pairs, states that only the MAIN figure is reproducible by anyone else and only from that path, writes up the confound explicitly, and relabels the worktree observation as a separate finding rather than as the evidence.
  - `[low]` `[patch]` **The load-bearing build comment overclaimed.** It read "`-buildvcs=false` MAKES THE BUNDLE A FUNCTION OF THE SOURCE RATHER THAN OF THE TREE", broader than what was measured, since the same commit's DW-105 records a surviving tree-dependent input. Raised independently by the verification-gap and intent-alignment layers. Now leads "CLOSES ONE INPUT: THE TREE'S STATE", states the three measured tree states that agree, and names the residual and DW-105 in its own paragraph. The provenance-trade paragraph is untouched.
  - `[low]` `[patch]` **The detector's encodings disagreed for non-Buffer input.** `Buffer.from(source)` decodes a string as UTF-8 while every `indexOf` searched `latin1`, shifting every offset past the first non-ASCII byte. Fixed by decoding strings as `latin1` too, with the reason recorded at the site. Proved: for a source containing non-ASCII bytes before the needle, the string and buffer paths now return identical findings; previously they did not.
  - `[low]` `[patch]` **DW-100's amendment read as both closed and open** — "Closed by Story 8.4g's commit" sat a paragraph above "DW-100 STAYS OPEN". Now scoped explicitly: the reproducibility half *and only that half* is discharged; the missing IBM Plex row half stays OPEN and owned by Story 8.4d.
  - `[low]` `[patch]` **The real-population test could not distinguish an all-clear from a couldn't-look** — an absent `src/generated/runtime` threw a bare `ENOENT`. Now fails with a directed message naming `npm run build:wasm`, red-proved by moving the directory aside. The `toHaveLength(1)` assertion was preserved, not weakened.

**Rejected findings, each with the authority it was tested against** (DW-87 — the population is every finding returned by all four layers: blind-hunter 21, edge-case-hunter 18, verification-gap 3 plus an explicit `No verification gaps found`, intent-alignment 8 divergences; deduplicated to 29 distinct claims):

1. *Three of the four needles are unanchored, unlike `build\tvcs=`.* Raised by three layers. Refuted on direction: an unanchored needle can only produce a **loud false red** that fails every designer gate, never a silent pass, and anchoring would *narrow* a safety guard. The bare `vcs` needle is anchored precisely because `vcs=` is short enough to collide with payload bytes, and that asymmetry is documented at the site.
2. *No test for a stamped binary with `vcs.modified=false`.* The guard fires on the presence of any setting, not on the modified value; measured probe A (clean tree, default build) carries three settings, and the red-proof fired on `vcs.revision`/`vcs.time`/`vcs=git` regardless. Not a distinct code path.
3. *Findings are ordered by needle rather than byte position.* Deterministic given a fixed `VCS_SETTINGS`; the test asserts that order. Cosmetic.
4. *`readValue`'s 200-byte cap truncates silently.* Reachable only on a false positive in payload bytes, which itself hard-fails the build loudly. Affects the readability of an already-fatal message.
5. *The `found` array and error message are uncapped.* Same reachability as 4; bounded at four in any legitimate binary.
6. *An empty or truncated wasm would be reported clean.* `execFileSync` throws on a non-zero `go build` exit, so a zero-length wasm cannot reach the check; and the emitted asset is content-addressed, so truncation changes its digest.
7. *The guard throws after `rmSync` already destroyed the previous generated tree.* The throw fails the build, and every gate runs `build:wasm` first, which repopulates. No persistent bad state.
8. *The label hardcodes a path already computed as `wasmPath`.* Cosmetic. The test's regex pins the emitted *asset* shape, a different fact from the raw output path.
9. *`import.meta.dirname` deviates from the repo's `fileURLToPath` convention.* The verification-gap layer confirmed it resolves correctly under Vitest 4. Style only.
10. *CI's `go build ./...` and the matrix/probe builds were not given the flag.* Traced by the verification-gap layer: they build test binaries and probes and compare **rendered output**, never binary bytes, so they share no observable contract with the bundle. The intent names the engine wasm build.
11. *`sprint-status.yaml` still says `backlog`.* Explicitly the closer's file; `bmad-build-auto` never touches it.
12. *The spec says `in-progress` with empty logs while `deferred-work.md` already says "landed".* An artifact of reviewing a mid-dispatch commit; both are written at finalize, which is this step.
13. *No post-fix `s1VisibleBytes` baseline was recorded for Story 8.4d.* Out of scope on the intent's own authority ("Do not set a byte threshold — Story 8.4d owns that"). Measured anyway, twice, and it held identical.
14. *`epic-8-context.md` was rewritten wholesale inside a build-flag commit, deleting epic scope content unreviewed.* Refuted as to attribution: the file is the workflow's own step-1 regenerated cache, declares itself generated, and its compile prompt instructs aggressive scoping. Measured: every flagged constraint retains an authority in `epics.md`, the PRD or the architecture spine. The residual orchestration signal is **deferred**, not rejected.
15. *The Stories list order is non-monotonic with no note.* It reflects the epics file's own resequencing (8.4g before 8.4f, 8.4d last), stated in that file's Cross-Story Dependencies.
16. *DW-105's owner is disjunctive and it does not require `-trimpath` before 8.4d's threshold.* Assigning routing and sequencing obligations is the orchestrator's and lead's call, not a build dispatch's. A deferral records the fact and the candidate remedy.
17. *DW-100's amendment records no commit SHA.* The SHA was unknowable at write time — this commit was amended during the dispatch — and the entry names the story, which is stable.
18. *The repository is not clean at the end.* The only uncommitted file was this spec's own status flip, committed at finalize.
19. *The real-population test could validate a stale artifact if `vitest` is run directly.* The suite's own entry point (`npm test`) runs `build:wasm` first, and CI uses that entry point. The absent-directory case is now a directed failure (patch 5). Retained as evidence under the deferred digest-comparison item rather than as a separate defect.

## Design Notes

**1. Why the diagnosis was re-confirmed rather than adopted, and what that turned up.** D-8.5.7's table was run at `92cd590`; this dispatch re-ran the discriminator at `873757f`. Result (main checkout, `folio-go/`, `GOOS=js GOARCH=wasm go build -o <tmp> ./wasm/cmd/engine`):

| probe | tree state | wasm sha256 | `vcs.modified` |
|---|---|---|---|
| A | clean | `68ee2569…c03117` | `false` |
| B | clean | `68ee2569…c03117` — identical | `false` |
| C | one stray untracked file | `1de602ca…3cc470` — differs | **`true`** |

Confirmed exactly. **Two traps were hit and are worth recording.** First, `go version -m` reports `unrecognized file format` for a wasm binary — so grepping its output for `vcs` prints nothing **because the tool failed**, not because the stamp is absent. The stamp must be read from the raw bytes. Second, the first attempt ran the probes in a **linked git worktree**, where `.git` is a file; Go's default `-buildvcs=auto` then silently omits the stamp entirely, so the stray file changed nothing and the run looked like a refutation. It was a non-measurement. A positive control in the main checkout is what distinguished the two.

**2. Why the guard lives in `build-wasm.mjs` and not only in a test.** `build:wasm` is a dependency of `typecheck`, `test` and `build`, so a check there fires in every designer gate and fails fast, before the raw wasm is deleted at `:76`. The reachable failure it catches is "the flag was dropped". The Vitest test catches the strictly weaker case where **both** the flag and the in-script guard are removed, and additionally proves the detector discriminates on synthetic input. Neither is decoration: each has a mutation that reddens it and nothing else.

**3. Why the provenance trade is acceptable — stated, not assumed.** `-buildvcs=false` removes `vcs.revision` from the binary, so the artifact no longer self-identifies its commit. Two things already cover that: the release manifest carries `releaseId` and `pageId` derived from asset hashes, which identify the *bundle* more precisely than a commit does, and AD-22 pins an exact toolchain and makes bumping it a release event. This is a deliberate exchange of binary provenance for reproducibility, not a free win.

**4. The corollary, written down because it changes how this pipeline measures.** The pipeline writes untracked files into the tree — halt files, result files, specs. Before this fix, a run that wrote an artifact between two measurements **changed the stamp it was measuring**. That is why one earlier pair of builds at one commit agreed and another did not; the "clean tree" qualifier on Story 8.4f's 2,203-byte observation was false in the relevant sense. The instrument perturbed the specimen.

**5. What this does NOT close.** The binary remains a function of the **checkout path** (Task 5). `-buildvcs=false` closes the tree-*state* defect that D-8.5.7 names and AC2 tests; it does not make two checkouts of one commit agree. Recorded as a deferred item with its measurement rather than fixed here, because `-trimpath` is a second variable and the AC names one flag.

## Verification

**Baseline to be MEASURED at `873757f` by the implementer, in the main checkout, recording tree state with every figure.** Note the tree carries the workflow's own `M _bmad-output/implementation-artifacts/epic-8-context.md` from step 1; a genuinely clean-tree probe must account for it.

**Commands:**

- `cd folio-go && go test -count=1 ./...` -- expected: **1811 pass / 2 fail / 5 skip** at this baseline, to be re-measured. The two are `TestCorpusMeetsP6ExerciseFloors` and its subtest `P6g_(opaque_names)` (`got 7, need >=20`), with drift twin `TestCorpusP6StatsMatchDeclaredBaseline`. **The one genuine standing red; never "fix" it.**
- `cd folio-go && go test -count=1 -tags=matrix ./...` -- expected: **1822 pass / 3 fail / 5 skip**. The third is `TestShippedFacesReproduceFromUpstream` and it is a **could-not-execute, not a byte divergence** — verbatim `fontgen: fontTools is not importable by this interpreter` (DW-86).
- **Sweep that test both ways.** Re-run with `FOLIO_FONTGEN_PYTHON=/Users/panitw/Projects/folio/.fontgen-venv/bin/python`; expected **PASS non-vacuously**, printing `derived and compared 3 of 3 faces`. **If it fails with the variable set, HALT.** Do not edit the test; do not commit the variable.
- **Exactly TWO standing reds by identity. Any third is a real failure.**
- `cd folio-go && go vet -tags=matrix ./...` -- expected: no output.
- `gofmt -l folio-go` -- **run from the REPO ROOT**, expected: no output. After a `cd folio-go` it prints an `lstat` error that reads like success.
- `cd lint && go test -count=1 ./...` -- expected: four `ok`, no FAIL. **`-count=1` always**: these rules walk directories and Go's cache does not track `ReadDir`.
- `cd folio-go && FOLIO_MATRIX_TARGET=<t> go test -count=1 -tags=matrix -run TestTargetRenderHash -v .` for each of `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` -- expected PASS, documents hashed per leg (re-measure the count). Then the same command **unset** -- passes in ~0.00s while asserting **nothing**, saying so at `matrix_test.go:2199`. **A control, never a fifth leg.**
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` -- expected PASS.
- `cd folio-designer && npm run typecheck` / `npm run lint` / `npm test` / `npm run test:e2e:compile` -- lint expected **exactly 4** `only-export-components` warnings (count and rule name are the invariant, not the line numbers). `test:e2e:compile` is `tsc --noEmit`; **do not report it as a run.**
- `shasum -a 256 fixtures/*/expected.pdf` -- expected **23** lines, **byte-identical to the pre-dispatch snapshot** (diff it, do not eyeball). **A moved golden is a HALT, not a re-record.**
- `md5 -q README.md` -- expected `078d7d80d518d54af2fc04fb270d46b8`, unchanged.
- Offline release, required because this story changes what reaches the bundle: `cd folio-designer && npm run build`, then `npm run verify:offline`, `npm run verify:offline:red`, `npm run verify:offline:wasm` -- expected pass, node **v24.16.0**. Quote any `s1VisibleBytes` with its command; it is a four-needle sum and **not a metric** (D-8.4.29). Never reason from it.

**Manual checks:**

- **The re-measurement, which is the story (D-8.5.7).** In the main checkout: build with a clean tree, then build with one stray untracked file present, and confirm the two wasm sha256 digests are **identical**. Report both. Then remove the probe file and confirm `git status --porcelain` is empty.
- **Red-prove the guard.** Remove `-buildvcs=false` from the build invocation, run `npm run build:wasm`, and confirm it **fails naming the stamp**. Restore with an **absolute** path — a `cd` earlier in the same compound command breaks a relative restore and silently leaves the mutation in place. Then `git status --porcelain` before continuing.
- **Red-prove the test.** Defeat the detector and confirm the named Vitest test reddens; confirm the synthetic stamped-buffer case fails when the detector is made to always report clean.
- **Read the stamp from bytes, never from `go version -m`** — it cannot parse wasm and its empty output is not evidence of absence.
- **Confirm zero `vcs.` occurrences** in the emitted wasm after the fix, against **three** before it.

## Auto Run Result

Status: done
Blocking condition: none

**Dispatch:** classic-intent, plan and implement in one dispatch (no `Halt after planning.`).
`baseline_revision` `873757f038f66491aa431dc2b1f015ed42132e42`, tree clean at start. Commit
`77cd80ed36af693bd1a284f149207d868e12e61d` (amended in place from `4f8516a`; no second commit).
`<intent-contract>` never edited — 3675 bytes, md5 `961a82c22f68b8299401f07db0ea8c9d`, verified
byte-identical after implementation, after review patching, and at finalize.

**Implemented change.** The engine wasm is now built with `-buildvcs=false`, so Go no longer stamps
`vcs.revision` / `vcs.time` / `vcs.modified` into it and the emitted binary stops being a function of
whether someone left a file lying in the checkout. A detector reads the stamp out of the raw bytes and
the build asserts on it at the point of use, so the flag cannot be dropped silently; a Vitest suite
proves the detector discriminates and ties it to the artefact this build actually emitted.

**The diagnosis was re-confirmed before anything changed, and re-confirming it caught two traps.**
`go version -m` answers `unrecognized file format` for a wasm binary, so grepping its output for `vcs`
prints nothing **because the tool failed** — the stamp must be read from raw bytes. And the first
attempt ran the probes in a linked git worktree, where `.git` is a file and `-buildvcs=auto` silently
omits the stamp entirely; the stray file changed nothing and the run looked like a refutation. It was a
non-measurement. A positive control in the main checkout is what separated the two.

**Files changed.**
- `folio-designer/scripts/build-wasm.mjs` — `-buildvcs=false` added to the `go build` argument list;
  `assertNoVCSStamp` applied to the freshly built wasm before the raw file is deleted by fingerprinting.
  The comment states what the flag closes (tree state), the residual it does not (checkout path, DW-105),
  and the provenance trade with its ground.
- `folio-designer/scripts/wasm-vcs-stamp.mjs` (new) — `findVCSStampSettings` / `assertNoVCSStamp`,
  reading Go build-info VCS records from raw bytes, `latin1` throughout so a needle cannot straddle a
  multi-byte decode.
- `folio-designer/scripts/wasm-vcs-stamp.test.mjs` (new) — four cases: a synthetic stamped buffer
  reported with values, the thrower naming them, a synthetic clean buffer with a near-miss mention of
  `vcs` reported clean, and the real emitted wasm asserted clean with `toHaveLength(1)` so it cannot
  pass vacuously.
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-100's reproducibility half discharged
  (mechanism corrected: the refuted "commit timestamp" removed, `vcs.modified` named as the operative
  input), DW-100 itself left **OPEN** on the IBM Plex row half owned by Story 8.4d; DW-105 opened for
  the measured checkout-path residual.
- `_bmad-output/implementation-artifacts/epic-8-context.md` — the workflow's own step-1 regenerated
  cache (`epics.md` was newer). Not story content.

**Review findings breakdown.** 5 patches applied (1 medium, 4 low), 5 deferred (2 medium, 3 low),
19 rejected — each with the authority it was tested against, enumerated in `## Review Triage Log`.
0 intent gaps, 0 bad-spec loopbacks; `review_loop_iteration` stayed 0.

**Follow-up review recommendation: `true`.** Patched findings only: high 0, medium 1, low 4 →
`3x1 + 1x4 = 7`, which is >= 5.

**The re-measurement (AC2), measured in this dispatch, in the main checkout, tree state recorded.**
Invocation `cd folio-go && GOOS=js GOARCH=wasm go build [-buildvcs=false] -o <tmp> ./wasm/cmd/engine`,
and for the post-fix pair also the real build path `cd folio-designer && npm run build:wasm`.

| probe | commit | tree state | flag | wasm sha256 | `vcs.` settings |
|---|---|---|---|---|---|
| A | `873757f` | clean | default | `68ee2569…c03117` | 3, `vcs.modified=false` |
| B | `873757f` | clean (repeat) | default | `68ee2569…c03117` — **identical** | 3 |
| C | `873757f` | one stray untracked | default | `1de602ca…3cc470` — **differs** | 3, **`vcs.modified=true`** |
| D | `873757f` | clean | `false` | `ff324971…d90f732` | **0** |
| E | `873757f` | one stray untracked | `false` | `ff324971…d90f732` — **identical** | **0** |
| F | `77cd80e` | clean (post-patch, via `npm run build:wasm`) | `false` | `ff324971…d90f732` | **0** |
| G | `77cd80e` | one stray untracked (post-patch) | `false` | `ff324971…d90f732` — **identical** | **0** |
| H | `77cd80e` | one modified **tracked** file | `false` | `ff324971…d90f732` — **identical** | **0** |

**Both digests, as required: clean `ff324971091afd641151d1658020852ad0120687c225e5760b05888d4d90f732`,
stray-untracked-file `ff324971091afd641151d1658020852ad0120687c225e5760b05888d4d90f732`.** Every probe
file was removed and `git status --porcelain` verified after each.

**And the payoff measured at the surface that consumes it.** Two full `npm run build` runs at `77cd80e`,
one on a clean tree and one with a stray untracked file, both report `s1VisibleBytes` **12,425,468**,
`s1.cachedBytes` **38,460,370**, `assetCount` **23**. Story 8.4f measured that figure moving 2,203 bytes
between two builds it believed were at one commit on a clean tree. It now holds still. Quoted with its
command; `s1VisibleBytes` is a four-needle sum and **not a metric** (D-8.4.29), and no threshold was set.

**Red proofs, all three run independently by this dispatch, none of them commit ordering.**
- *The guard.* Flag removed → `npm run build:wasm` exits **1**: `src/generated/runtime/folio-engine.wasm
  carries a Go build-info VCS stamp … vcs.revision=…, vcs.time=…, vcs.modified=true, vcs=git. Build it
  with -buildvcs=false.` Restored from an absolute-path backup; tree re-verified clean.
- *The test's real-population arm.* Flag **and** in-script guard removed → the emitted wasm was stamped
  and exactly the fourth test reddened (1 failed, 3 passed).
- *The test's synthetic arm.* Detector forced to always report clean → the two stamped cases reddened
  and the two clean cases stayed green, which is precisely why the stamped pair exists.

**Verification performed — measured in this dispatch, after the patches.**
- `cd folio-go && go test -count=1 ./...` → **1815 pass / 2 fail / 5 skip**. The two are
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest, verbatim `P6g (opaque names)
  floor not met: got 7, need >=20` — the mandated permanent red. Drift twin
  `TestCorpusP6StatsMatchDeclaredBaseline` passed.
- `cd folio-go && go test -count=1 -tags=matrix ./...` → **1826 pass / 3 fail / 5 skip**. The third is
  `TestShippedFacesReproduceFromUpstream`, a **could-not-execute, not a byte divergence** — verbatim
  `fontgen: fontTools is not importable by this interpreter`. It never compared bytes (DW-86).
  **No third distinct failure by identity.**
- **Swept both ways.** With `FOLIO_FONTGEN_PYTHON=…/.fontgen-venv/bin/python` the same test **PASSES
  non-vacuously** — `fontgen: derived and compared 3 of 3 faces`, all three committed faces reproducing
  from their recorded derivation. The variable was set on the command line only; it is in no committed
  file and the test was not edited.
- No `.go` file and no `fixtures/` file is touched by this story (measured: 0 such paths in
  `git diff --name-only 873757f..HEAD`), so the baseline and post-change Go figures are necessarily the
  same numbers.
- `go vet -tags=matrix ./...` empty. `gofmt -l folio-go` **run from the repo root** empty.
  `cd lint && go test -count=1 ./...` four `ok`, no FAIL.
- Four AD-21 legs (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`) **PASS, 24 documents hashed
  each**. The unset control **passes in 0.00s asserting nothing**, saying so at `matrix_test.go:2199` —
  a control, never a fifth leg. `TestCrossTargetByteIdentity` **PASS**, all four targets byte-identical.
- Designer: `typecheck` exit 0; oxlint **exactly 4** `only-export-components` warnings
  (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1324,1331`); `npm test` **40 files / 387 tests, all passing**
  (baseline 39/383 — this story adds one file and four tests). The new suite was confirmed to actually
  run, not merely exist. `npm run test:e2e:compile` clean — **this is `tsc --noEmit`, not a run.**
- `shasum -a 256 fixtures/*/expected.pdf` → **23 lines**, and `git diff 873757f..HEAD -- fixtures/` is
  empty: **no golden moved.** The stamp never reached rendered output, so D-8.5.7's confinement to the
  designer bundle stands. `md5 -q README.md` → `078d7d80d518d54af2fc04fb270d46b8`, unchanged.
- Offline release on node **v24.16.0**: `npm run build`, `verify:offline`, `verify:offline:red`,
  `verify:offline:wasm` all exit 0.

**Matrix test audit.** All six rows covered by verification that ran: rows 5 and 6 (detector on stamped
and on clean bytes) by the Vitest suite, which ran and passed; row 1 (clean build, zero settings) by the
suite's real-population case **and** by measured probes D/F; rows 2 and 3 (stray untracked file, modified
tracked file) by measured probes E/G/H; row 4 (stamp returns) by the guard's red-proof. Rows 2, 3 and 4
are covered by measured manual probes rather than suite tests — two full wasm builds per row cannot sit
in a unit suite — and the spec's `## Verification` designates them manual checks. Recorded as a
disclosure, not claimed as automated coverage.

**Residual risks.**
1. **The guard watches one cause, not the property.** Nothing compares two builds' digests, so a second
   tree-dependent input would keep every gate green while the story's title is false. Deferred.
2. **The checkout path is still an input, and this is measured, not suspected.** Same commit, same flag,
   both unstamped, three different absolute paths → three different digests, with the main checkout's
   figure stable throughout; each binary carries 87 occurrences of its own root's absolute source paths.
   `-trimpath` is the candidate remedy and was deliberately not applied — a second variable in the same
   measurement would make the re-measurement unattributable. Recorded as DW-105.
3. **Both red-proofs are prose and cannot be re-run.** The repo's `redProof` harness in
   `verify-offline-release.mjs` is the established shape for exactly this; Story 8.4f owns that file next.
4. **A standing constraint was breached mid-dispatch and repaired.** The implementation subagent used
   root `README.md` as the modified-tracked-file probe, which the dispatch names as never modified. It was
   restored, its md5 re-verified `078d7d80d518d54af2fc04fb270d46b8`, and it is absent from the commit — so
   the end state honours the invariant, but the act breached it. The independent re-measurement of that
   same matrix row (probe H) used a different tracked file.
