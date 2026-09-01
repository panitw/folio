---
title: 'Story 8.4f: A bound nobody can cross silently'
type: 'bugfix'
created: '2026-09-01'
status: 'ready-for-dev'
baseline_revision: '548aa292294ac5093c4012fadeb3f141ee1248fc'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: [oversized]
deferred: []
---

## Frontmatter warnings — what they mean here

`oversized` — the spec exceeds the template's 1600-token guide. It is the house size for this run;
the length is Code Map and Verification, not intent. Nothing is deferred and no goal is split.

**`multiple-goals` was considered and deliberately NOT set.** The build gate (AC1–AC3) and the
runtime swallow (AC4) touch different files and look independently shippable, but they are not:
**AC4's chosen remedy is only defensible because AC1 lands with it.** The argument for narrowing the
swallow rather than removing it (Design Note 4) rests on the over-the-bound cause becoming
*unreachable in a shipped release* — which is true only once the build refuses it. Ship AC4 alone
and the narrowing is unargued; ship AC1 alone and the swallow stays. D-8.5.1(a) ruled them into one
story for this reason. Splitting is not merely discouraged here — it would break the reasoning.

## In plain terms (read this first if you just want the gist)

The designer's first-run load screen reads a small JSON payload that the build stamps into
`index.html`. The reader refuses a payload that lists more than **64** cached assets. The release
currently lists **23**, so there are 41 to spare — but nothing checks that number at build time.
If a future release ever crosses it, the reader hands back `undefined`, the caller turns any
failure into `undefined` as well, and **the first screen a user sees quietly loses its payload**:
no progress bar, no manifest, no error, nothing said. An all-clear that is indistinguishable from a
couldn't-look.

This story does two things. It teaches the offline release **verification job** to count the
release's assets and **fail the build** — naming the count and the bound — so the number can never
be crossed in silence again, and it proves that assertion by manufacturing a release over the bound
and watching it fail. And it stops the reader's blanket `catch` from turning *every* cause into the
same `undefined`, so "this bootstrap is not JSON" and "this release is over the bound" are no
longer the same event.

It does **not** raise the bound, touch the five S1 rows, or add, remove or rename a single asset.
Raising the ceiling belongs to whichever story needs the headroom, after this one lands.

<intent-contract>

## Intent

**Problem:** `folio-designer/src/release-payload.ts:8` declares `maximumCacheAssets = 64` and
`parseS1Payload` returns bare `undefined` when `assetCount` crosses it (`:13`). Its only caller,
`loadS1Payload` (`:33`), wraps the read in `try { … } catch { return undefined }` — soft twice over.
Measured at `548aa29`: the emitted release carries **23** assets, so the bound is real but distant,
and **no build script checks it**. `verify-offline-release.mjs` carries **thirteen** S1 assertions
(`:45`, `:47`, `:59`, `:61`, `:62`, `:65`, `:67`, `:69`, `:72`, `:73`, `:76`, `:77`, `:79` — counted,
not estimated) and **not one of them counts the assets against the bound.** The failure mode is
therefore silent at both ends: the build says
nothing, and the runtime "catches" the problem by being quiet. That is the run's signature defect —
an all-clear indistinguishable from a couldn't-look — sitting in the first thing a user sees.

**Approach:** Make the offline release verification job enforce the bound it already knows how to
read, deriving the number from the declaration in `release-payload.ts` rather than re-typing it, so
the two can never drift; fail the build naming the count and the bound; red-prove it by
manufacturing a release over the bound. Then give `parseS1Payload` a named reason per rejection and
narrow `loadS1Payload`'s catch to the one throw it exists for, so "malformed" and "over the bound"
stop being the same value.

## Boundaries & Constraints

**Always:**
- The verifier's bound is **derived from `folio-designer/src/release-payload.ts`**, never re-typed
  as a literal in `verify-offline-release.mjs`. A second copy in lockstep is the drift this repo
  has already been bitten by, and the build does not run Vitest, so a tie test alone would not
  close it.
- **The derivation must fail loudly when it cannot find what it is looking for.** If the constant is
  renamed, deleted, reformatted, or declared twice, verification fails with a message saying so.
  A reader that silently yields "no bound" would rebuild this story's own defect inside its fix.
- The failure message **names the measured count and the bound it crossed**, in the file's existing
  `fail('…')` voice.
- The new assertion is **red-proved** in `runRedProofs`, and the proof is **held to its own guard**
  by passing `redProof`'s third `expected` argument — the mechanism the file documents at `:133`–`:135`.
- Every guard added or widened is **mutation-proved**: delete it, and a **named** test or red proof
  must go red. A guard that survives its own deletion is decoration.
- The corpus does not move: 23 golden PDF digests unchanged, four AD-21 legs and
  `TestCrossTargetByteIdentity` green, `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`.
- Commit trailer `Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>`. `main` only,
  never a branch, never a push, never `git add -A` or `git add .`.

**Block If:**
- The measured release asset count at `548aa29` is **not 23**, or the declared bound is **not 64** —
  the story's premise has moved and the plan must be re-gated rather than adjusted in flight.
- Enforcing the bound cannot be done without changing which rows the S1 payload carries, or without
  adding, removing or renaming an asset. Both are out of scope by D-8.5.1(a); a route that needs
  either is the wrong route.
- Narrowing `loadS1Payload`'s catch is found to change what a user sees on any path other than the
  over-the-bound path — that is a product change this story is not scoped to make.
- The red proof cannot be made to fail on the **new** guard (it keeps tripping an older one first).
  Placement is the fix, not a weaker assertion; if placement cannot fix it, halt rather than accept
  a proof that fails for the wrong reason.

**Never:**
- **Never raise `maximumCacheAssets`.** Raising it is legitimate only *after* this story lands and
  only with stated headroom, and it belongs to whichever story needs it (D-8.5.1b). A bound tuned to
  the current population is one the code can move.
- **Never touch the S1 row pinning.** The five ids by ordered exact equality (`:58`–`:59`),
  `cachedRows.length === 4` (`:69`), `rows[4]` positionally (`:75`–`:76`), and `cjk-font` as the
  `Math.max` of the font rows (`:78`–`:79`) constrain **rows**; this bound counts **assets**.
  Conflating them is the error D-8.5.1 exists to correct.
- Never add, remove or rename any asset. No fonts, no catalogue work — Story 8.5 owns the
  catalogue, this story owns the bound.
- Never take on DW-101 / DW-103 (CI runs neither the Playwright specs nor the fontgen check).
- Never phrase an acceptance criterion at a layer nothing in this repo executes. The 12 Playwright
  specs have never run here and CI runs only `tsc --noEmit`. The live trigger is *"when CI executes
  the Playwright suite"*; *"when browser e2e arrives"* is void.
- Never write `reader`, `date` or `examined` into any attestation record, and never modify, move,
  delete or stage `README.md`.
- Never open a `deferred-work.md` entry for this story. Nothing here is deferred by design; if
  implementation surfaces something genuinely out of scope, record it in this spec's `deferred:`
  frontmatter and let the reviewer place it.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Release within the envelope | Emitted `dist/` with 23 assets, bound 64 | `verifyOfflineRelease` proceeds exactly as today; no new output | No error expected |
| Release over the bound | Manifest listing 65+ assets | Verification **fails**, message names the measured count and `64` | Build exits non-zero |
| Release under the floor | Manifest listing fewer than `minimumCacheAssets` (10) | Verification **fails**, message names the measured count and `10` | Build exits non-zero |
| Bound declaration unreadable | `maximumCacheAssets` renamed, commented out, or declared twice in `release-payload.ts` | Verification **fails**, saying the declaration could not be read as a single live constant | Build exits non-zero |
| Bootstrap is not JSON | `index.html` script node holds a truncated string | `loadS1Payload` reports the **malformed-json** reason; the load screen behaves exactly as today (`unavailable` + retry) | `SyntaxError` caught and named; any other throw rethrown |
| Payload over the bound at runtime | Hand-edited bootstrap with `assetCount` 65 | `loadS1Payload` reports the **over-the-bound** reason — a distinct value, producible by no other input | No throw; caller behaves as today |
| Payload malformed in shape | Surplus field, stale arithmetic, delivery fiction | Reported with their own reasons, none equal to the over-the-bound reason | No throw |
| Dev server, no bootstrap node | `import.meta.env.DEV`, no `#folio-release-bootstrap` | The dev bypass fires — and **only** on the absent-bootstrap reason, not on malformed or over-the-bound | No error expected |

</intent-contract>

## Code Map

All line anchors measured at `548aa292294ac5093c4012fadeb3f141ee1248fc`, clean tree. Re-measure
before the first edit and record any drift in `## Spec Change Log`; cite by what a thing asserts,
not by its line, if the two disagree (D-8.4.13).

### The declaration — the single authority for the bound

- `folio-designer/src/release-payload.ts:7-8` — `const minimumCacheAssets = 10` and
  `const maximumCacheAssets = 64`, each on its own line, unexported, module scope. **These two lines
  are the authority the verifier must derive from.** They are plain `^const <name> = <digits>$`
  declarations, which is what makes a source-anchored read viable.
- `:13` — the one long shape check: **seventeen `||`-joined clauses sharing a single
  `return undefined`.** Two of them are the bound (`candidate.assetCount < minimumCacheAssets` and
  `candidate.assetCount > maximumCacheAssets`); the other **fifteen** are unrelated shape checks.
  This is where "over the bound" becomes indistinguishable from everything else.
- `:10` `parseS1Payload(value: unknown): S1Payload | undefined` — **seven** `return undefined` sites
  in all: `:11`, `:13`, `:15`, `:16`, `:20`, `:22`, `:26`. Every one of them must acquire a distinct
  reason.
- `:30-34` `loadS1Payload()` — reads `#folio-release-bootstrap`, returns `undefined` when the node
  or its text is absent (`:32`), then `try { return parseS1Payload(JSON.parse(...).s1) } catch {
  return undefined }` (`:33`). **`parseS1Payload` never throws**, so this catch exists solely for
  `JSON.parse`. It currently also swallows any unexpected throw.
- `:36` `formatMiB` — untouched.

### The verification job — where the assertion lands

- `folio-designer/scripts/verify-offline-release.mjs:21` `verifyOfflineRelease(outputDir, {wasmWitness})`
  — the whole gate. Reads `dist/offline-release-manifest.json` at `:23-25`.
- `:12` `const fail = (message) => { throw new Error('offline release verification failed: ' + message) }`
  — the voice every assertion uses. The new one uses it too.
- `:27` release-identity shape check; `:29-32` the per-asset URL loop building `manifestUrls`.
  **`:32` is the end of that loop and the insertion point.**
- `:33-34` `runtimeOutputUrls(outputDir)` then `sameSet(manifestUrls, outputUrls)` →
  `fail('manifest and production runtime output are not an exact set')`. **This is the guard that
  will otherwise swallow the red proof**: any mutation that changes the manifest's asset list trips
  it. The new check must run *before* it.
- `:44-47` — the file already **regex-extracts** the bootstrap out of `index.html` text; `:49-50`
  does the same for `const RELEASE = …` in `sw.js`. **Reading a source artifact and failing when the
  pattern cannot be located is this file's established idiom**, not a new one.
- `:57-79` — the S1 block. `:59` carries `s1.assetCount !== release.assets.length`, which is the tie
  that makes bounding `release.assets.length` equivalent to bounding the parser's `assetCount`.
  `:58-59`, `:68-69`, `:75-76`, `:78-79` are the **row pinning — read-only, do not touch.**
- `:82-86` per-asset existence and digest; `:87-98` immutability and Brotli sidecars. Both loop over
  `release.assets` and both would also trip on manufactured assets — a second reason placement matters.
- `:126-137` `redProof(name, mutate, expected)` — the harness. `:132` requires the mutation to
  produce *some* failure; `:133-135` carries the comment *"Where a proof names the guard it is
  proving, hold it to that guard rather than to any failure at all"* and honours `expected` by
  substring. **The new proof passes `expected`.**
- `:139-174` `runRedProofs`. The two patterns named by the epic: `s1-total-mismatch` (`:157`) and
  `s1-delivery-fiction` (`:158`) — both read the manifest, mutate one field, write it back
  un-pretty-printed, and restore from the original buffer. `:142-143` show the *other* shape, which
  is the relevant one here: mutate `release.assets` then `rewriteRelease` so identities stay
  consistent. `:172` shows `expected` in use (`'development offline bypass shipped in'`).
- `:119-124` `rewriteRelease(outputDir, release)` — recomputes `pageId` and `id` from the assets and
  rewrites both the manifest and `sw.js`. **This is what lets a proof add assets without tripping
  the identity checks at `:40-41`.**
- `:176-181` the CLI entry: `--red-proof`, `--red-only`, `--wasm-witness`.

### Where the bound reader can live, and how it is tested

- `folio-designer/scripts/offline-release-contract.mjs` — 26 lines, the shared contract module,
  imported by the generator (`:6`), the verifier (`:8`) and its own test. Imports `node:crypto`, so
  it is **build-time only and must never be imported from `src/`**.
- `folio-designer/scripts/offline-release-contract.test.mjs` — 14 lines, two cases. The pattern for
  unit-testing a scripts-side helper.
- `folio-designer/vite.config.ts:33` — `test: { … include: ['src/**/*.test.{ts,tsx}',
  'scripts/**/*.test.mjs'] }`. **Vitest already collects `scripts/*.test.mjs`**, so a helper placed
  there gets a test that runs under `npm test` with no config change.
- `folio-designer/tsconfig.app.json` — `"include": ["src"]`. `src/` imports nothing from `scripts/`
  (verified: zero import statements, four comment mentions). **The sharing is one-directional by
  construction** — which is precisely why the verifier must read the constant rather than import it.

### The runtime consumers of `loadS1Payload` / `parseS1Payload`

- `folio-designer/src/main.tsx:41` — the only caller of `loadS1Payload`, inside `startObservation`.
  Note the control flow: `render()` runs at `:39` **before** the `try` at `:40`, and the block is
  `try { … } finally { observationInFlight = false }` with **no catch**. `startObservation` is
  invoked as `void startObservation()` at module scope.
- `:45` — `if (import.meta.env.DEV && !payload) { … dev-bypass … }`. **Currently fires on any
  falsy payload**, so in dev a malformed bootstrap reads as "no bootstrap".
- `:46` — `registerOfflineLifecycle(expectedPageId, payload, …)`.
- `folio-designer/src/offline-lifecycle.ts:45` — with no payload, publishes `'unavailable'`, which
  is what produces the load screen's *"Offline cache unavailable"* phase and its **Retry
  preparation** button (`LoadScreen.tsx:19-20`, `:30`). **This is the behaviour that must not
  change**, and the reason full removal of the catch is unsafe (Design Note 4).
- `folio-designer/src/release-payload.test.ts` — 22 lines, three cases. `:16-21` already asserts the
  parser rejects `assetCount = 9` and `assetCount = 65`. **That test is the defect, not the fix:**
  it certifies the silent `undefined`. It must be re-pointed at the named reasons.
- `folio-designer/src/engine-ownership-contract.test.ts:66` — `release-payload.ts` is on the
  allowlist of files permitted to parse JSON. Do not disturb that list.

### Read-only evidence (do not modify)

- `folio-designer/dist/offline-release-manifest.json` — a build artifact, gitignored. **Measured
  twice this dispatch at `548aa29`, clean tree, and the two disagree.** The pre-existing `dist/`
  read before any build: `assets: 23`, `s1.assetCount: 23`, `s1.cacheAssets: 23`,
  `s1.cachedBytes: 38,460,398`, `s1VisibleBytes: 12,428,555`. After
  `cd folio-designer && npm run build` this dispatch: **identical except**
  `s1VisibleBytes: 12,426,352` — a **2,203-byte** difference between two builds at one commit, which
  is the drift D-8.4.35a records as still unexplained. **`s1VisibleBytes` is a four-needle sum and is
  not a metric (D-8.4.29) — quote it with its command, never reason from it.** The figure this story
  actually depends on, **23 assets against a bound of 64**, was identical in both reads.

> **⚠ CORRECTED 2026-09-01 (D-8.5.7 / Story 8.4g's close). THE 2,203-BYTE FIGURE IS NOT WHAT THIS
> PASSAGE SAYS IT IS, ON TWO COUNTS — and the second is worse than the first.**
>
> **(1) The "clean tree" premise is FALSE.** `go build` defaults to `-buildvcs`, stamping
> `vcs.modified` into the wasm, and Go derives that flag from `git status` where **an untracked file
> is enough**. This pipeline writes untracked files into the tree — halt files, result files, specs —
> so **the instrument was perturbing the specimen.**
>
> **(2) It was never two builds at one commit.** It compared a **pre-existing `dist/` of unknown
> provenance** against **one fresh build.** That stale `dist`'s engine row was stamped with a
> **different commit and a different dirty flag**, which changed its compressed size. **So the figure
> is UNATTRIBUTED, not merely mis-premised** — a distinction worth keeping, because "measured under a
> wrong assumption" and "not a comparison at all" are different defects.
>
> **Settled by discriminator at `92cd590` and re-measured with a control at `c985b9c`:** two builds
> with provably identical tree state give a **byte-identical** wasm; one stray untracked file changes
> it. **The build is deterministic.** Story 8.4g fixed it with `-buildvcs=false`; after the fix, clean
> and stray-file builds produce **the same digest and a byte-identical manifest in whole**.
>
> **THIS STORY IS NOT INVALIDATED.** The figure it actually depends on — **23 assets against a bound
> of 64** — was **identical in both reads** and is unaffected. Re-measure at the current HEAD rather
> than carrying any figure forward from this passage.

- `_bmad-output/implementation-artifacts/epic-7-8-decision-log.md` `### D-8.5.1` — the binding
  ruling; part (a) is this story, part (b) is Story 8.5's. Do not amend it.
- `_bmad-output/implementation-artifacts/bmad-build-auto-result-8-5-…​.md` finding 6 — where this
  defect was first measured.
- `folio-designer/scripts/generate-offline-release.mjs` — the producer. It builds the payload but
  has no concept of the bound; **it is not where the assertion goes** (the epic names the verifier),
  and it is not modified.

## Tasks & Acceptance

> **TWO TASKS ADDED 2026-09-01 BY THE ORCHESTRATOR, ROUTED FROM STORY 8.4g'S CLOSE (DW-106, DW-107).**
> They are here in **Tasks** and not only in prose, because **a ruling recorded in every artifact
> except the one a builder executes is a ruling that did not happen** (D-8.4.31). Both land in
> `verify-offline-release.mjs` — **the file this story already owns** — which is why routing them now
> is cheap and routing them later is not.
>
> **T-A (DW-107): make Story 8.4g's two AC2 red-proofs RE-RUNNABLE.** They are currently **prose in a
> Delivery Log and cannot be executed.** The `redProof` harness in this file is their natural home.
> A red-proof that exists only as a sentence has been performed once and can never fail again.
>
> **T-B (DW-106): the new bound guard watches a CAUSE, not the PROPERTY.** Story 8.4g's guard asserts
> the `vcs.` stamp is absent; the property that matters is **two builds of one commit agree**. That is
> the **proxy-versus-purpose defect the engineering lead recorded against its own ruling at
> D-8.4.30**. Add the property-level check.
>
> **State its known limit rather than overclaiming it:** a two-builds-agree check **would NOT have
> caught DW-105**, because it holds the checkout path fixed while DW-105 is path dependence — the same
> commit built from three different paths gives three different binaries, each embedding **87
> occurrences of its own absolute root**. **Its value is the NEXT input, not the last one.** Say that
> in the assertion's own comment, so a later reader does not read it as covering more than it does.

**Execution:**

1. `folio-designer/scripts/offline-release-contract.mjs` — add one exported helper that reads
   `src/release-payload.ts` from disk and returns the declared `{ minimumCacheAssets,
   maximumCacheAssets }`. It must match each constant with a **line-anchored, multiline** pattern
   (`^const <name> = (\d+)$`) so a commented-out copy cannot satisfy it, and must **throw** unless
   each name matches **exactly once** — zero matches (renamed, deleted, reformatted) and two matches
   (a second live declaration) are both failures, with a message naming which constant and which
   case. Resolve the source path from the module's own location (the pattern `verify-offline-release.mjs:10`
   already uses), never from `process.cwd()`, and take the **source text as an optional argument
   defaulting to that file** so Task 2 can exercise the failure modes without editing the real one.
   Rationale: this makes the TypeScript declaration the single authority, and makes a rename a loud
   build failure instead of a silently disabled bound. Placing it here rather than inline gives it a
   test that `npm test` already collects.
2. `folio-designer/scripts/offline-release-contract.test.mjs` — extend with cases for the helper:
   it returns the two numbers actually declared today (asserted against a re-read of the real file,
   **not** against re-typed literals); it throws when the constant is absent; it throws when the
   only occurrence is inside a comment; it throws when the constant is declared twice. Feed the
   last three from fixture strings through an injectable source argument rather than by editing the
   real file. Rationale: the reader is the new single point of failure for the whole guard, and a
   comment-blind or first-match-wins regex is the recorded way this class of guard has failed here
   before.
3. `folio-designer/scripts/verify-offline-release.mjs` — call the helper and assert the envelope
   **immediately after the manifest URL loop that ends at `:32`, before `runtimeOutputUrls` at
   `:33`**. One check, one `fail(…)`, message naming the measured `release.assets.length` and the
   bound it crossed (both ends: over `maximumCacheAssets`, under `minimumCacheAssets`). Add a short
   comment recording (a) why the count is taken from `release.assets.length` — the tie at `:59`
   makes it the same quantity the parser bounds, and it is available earlier — and (b) why the
   check sits before the set comparison. Rationale: this is the epic's named home, and the placement
   is what lets the red proof fail on this guard rather than on `sameSet`.
4. `folio-designer/scripts/verify-offline-release.mjs` — add one red proof to `runRedProofs`,
   following the `:142` shape rather than the `:157` one: read the manifest, **append** synthetic
   asset entries until `release.assets.length` exceeds the declared maximum, `rewriteRelease` so the
   identities stay consistent, restore from the saved buffers. **Pass `expected`** so it is held to
   the new guard's own message (`:133`–`:135`). Name it in the file's existing convention
   (`s1-…`/`…-count` style, chosen to read as what it proves). Rationale: AC3 — an assertion that
   has never failed has not been shown to discriminate.
5. `folio-designer/src/release-payload.ts` — give every rejection a named reason. `parseS1Payload`
   returns a discriminated result rather than `S1Payload | undefined`; the `assetCount` bound moves
   out of the seventeen-clause condition at `:13` into its own check so its reason is reachable by
   **that cause and no other**. `loadS1Payload` returns the same shape, reporting a distinct reason
   for an absent bootstrap node, and **narrows its catch**: a `SyntaxError` from `JSON.parse`
   becomes the malformed reason; **anything else is rethrown**, not swallowed. Keep the exported
   `S1Payload`/`S1Row` types and `formatMiB` unchanged. Rationale: AC4 — the two causes stop sharing
   one outcome, and the blanket swallow is removed for every cause except the single throw it exists
   for.
6. `folio-designer/src/main.tsx` — adapt `:41`–`:46` to the new shape at exactly one place, mapping
   a rejected result to the `undefined` payload `registerOfflineLifecycle` already expects, so the
   user-visible behaviour is unchanged. **Narrow the DEV bypass at `:45` to the absent-bootstrap
   reason only** — the comment above it says the dev server emits no bootstrap, and that is the only
   case it is meant to cover; a malformed or over-bound bootstrap must not read as "no bootstrap".
   Rationale: the same "two causes, one outcome" defect, one line away, and closing it costs one
   condition.
7. `folio-designer/src/release-payload.test.ts` — re-point `:16`–`:21` from `toBeUndefined()` to the
   named reasons, and add the discrimination assertions: an over-the-bound payload reports the
   bound reason and **no other input in the file reports it**; a malformed (non-JSON) bootstrap
   reports the malformed reason; the `SyntaxError`-only catch is proved by confirming a non-syntax
   throw propagates. Rationale: this file currently certifies the silent `undefined`; it must
   certify the distinction instead.
8. Run the whole `## Verification` cadence, including its **Manual checks**, and record **measured
   counts**, not adjectives. Rationale: the offline release build is this story's own subject, and a
   claim about it that was not measured this dispatch is not evidence.

**Acceptance Criteria:**

- **AC1 — the build fails, naming the count and the bound.** Given a release whose asset count
  exceeds the declared `maximumCacheAssets`, when `npm run verify:offline` runs, then it exits
  non-zero with a message that contains both the measured count and the bound, and the verification
  does **not** return a release record. Given the release as it stands at `548aa29` (23 assets),
  when the same command runs, then it passes exactly as before with no new output.
- **AC2 — the assertion lands in the named file, on the existing pattern.** Given the new check,
  when the diff is read, then it is in `verify-offline-release.mjs`, uses that file's `fail(…)`
  voice, and its red proof is registered in `runRedProofs` alongside `s1-total-mismatch` and
  `s1-delivery-fiction`, passing `redProof`'s `expected` argument so it is held to its own guard.
- **AC3 — red-proved, and proved to discriminate.** Given a release manufactured over the bound,
  when `npm run verify:offline:red` runs, then the new proof reports the failure and the failure
  message is the **new guard's**, not `sameSet`'s or an identity check's. And given the new check is
  deleted from the source, when the red proof runs, then it fails as `escaped verification` — the
  guard does not survive its own deletion.
- **AC4 — the swallow no longer merges two causes.** Given a bootstrap that is not JSON and a
  bootstrap that is valid JSON but over the bound, when `loadS1Payload` reads each, then it reports
  **two different named reasons**, the over-the-bound reason is produced by that cause and by no
  other input asserted in `release-payload.test.ts`, and a throw that is not a `JSON.parse`
  `SyntaxError` **propagates** rather than being converted to a rejection. Given a document with no
  bootstrap node under the dev server, when the app starts, then the dev bypass still fires; given a
  malformed or over-bound bootstrap under the dev server, then it does **not**.
- **AC5 — the bound's authority is single, and unreadable means loud.** Given
  `maximumCacheAssets` is renamed, commented out, or declared a second time in `release-payload.ts`,
  when verification runs, then it fails saying the declaration could not be read as a single live
  constant — the number is never re-typed in the verifier and a rename can never silently disable
  the guard.
- **AC6 — nothing else moved.** Given the corpus, when the story is complete, then
  `shasum -a 256 fixtures/*/expected.pdf` is **23** lines with unchanged digests, the four AD-21
  legs and `TestCrossTargetByteIdentity` pass, `README.md` md5 is unchanged, the S1 row pinning is
  byte-identical to `548aa29`, `maximumCacheAssets` is still **64**, and the suites show exactly the
  standing reds recorded in `## Verification` and no others.

## Spec Change Log

## Review Triage Log

## Design Notes

**1. Why the verifier reads the constant instead of re-typing it — and why a tie test is not
enough.** The obvious alternative is `const maximumCacheAssets = 64` in the verifier plus a Vitest
test asserting the two agree. That fails on a measured fact: `npm run build` is
`build:wasm && tsc -b && vite build && build:offline && verify:offline` — **it does not run
Vitest.** A drifted copy would therefore pass the build and only redden in a suite the build never
consults, which is the same "declared but unenforced" shape this story exists to remove. Reading the
declaration makes drift impossible rather than detectable. The idiom is already this file's own:
it regex-extracts the bootstrap from `index.html` (`:44`) and `const RELEASE = …` from `sw.js`
(`:49`), and its `worker-manifest-drift` proof (`:146`) fails outright when its pattern cannot be
located. Import is not an option in the other direction: `offline-release-contract.mjs` imports
`node:crypto`, `tsconfig.app.json` includes only `src`, and `src/` imports nothing from `scripts/` —
verified, not assumed.

**2. The reader is now the single point of failure, so it gets the strictest treatment in the
story.** Three failure modes have bitten guards in this repo before and all three are cheap to close
here: a raw-text scan that counts **commented-out** code as live (line-anchor the pattern); a
**first-match-wins** read that silently prefers one of two declarations (require exactly one match);
and a **soft miss** that yields a default (throw, never default). Task 2 exists so each of those is
proved, not asserted. Note the asymmetry that makes this safe: a false *failure* here stops a build
loudly; a false *success* re-creates the defect invisibly. Bias every ambiguous case toward failing.

**3. Why the check goes before `sameSet`, and why that is a correctness question rather than
tidiness.** `runtimeOutputUrls` walks `dist/assets` and `sameSet` demands the manifest and the
emitted output be an exact set (`:33`–`:34`). Any red proof that adds assets to the manifest breaks
that set. The digest loop (`:82`–`:86`) and the Brotli loop (`:87`–`:98`) would also trip on
synthetic assets. So a bound check placed anywhere after `:33` **cannot be red-proved on its own
message** — the proof would pass `redProof`'s bare "something failed" test while proving nothing
about the bound, which is exactly what `expected` was added to this file to prevent. Placing the
check at `:32` and passing `expected` closes both halves. If it still fails for the wrong reason,
that is a Block If, not something to soften the assertion around.

**4. Why `loadS1Payload`'s catch is narrowed rather than removed — measured from the control flow,
not assumed.** Removing it entirely lets a `JSON.parse` `SyntaxError` propagate out of
`startObservation`. Read `main.tsx:37`–`:48`: `render()` runs at `:39`, **before** the `try` at
`:40`; the block has a `finally` and **no catch**; and the function is called as
`void startObservation()` at module scope. So the throw does not blank the page — it does something
worse to diagnose. Today an unparseable bootstrap reaches `registerOfflineLifecycle` with
`payload === undefined`, which publishes `'unavailable'` (`offline-lifecycle.ts:45`), and the load
screen states *"Offline cache unavailable"* and offers **Retry preparation** (`LoadScreen.tsx:20`,
`:30`). With the catch removed, `registerOfflineLifecycle` is never reached: the screen sits on
*"Checking cache"* forever, with no message and **no retry button**. That trades a stated failure
for an unstated hang — a *worse* silent outcome, produced in the name of removing silence. Hence:
keep a catch, narrow it to the one throw it exists for, rethrow everything else, and carry a named
reason. **The implementer must re-verify this control flow before relying on it**; it is the whole
argument for choosing "distinguish" over "remove", and the dispatch requires the reason to be
stated, not asserted.

**5. What AC4 does and does not change for a user, said plainly so a reviewer judges it rather than
discovers it.** After this story, a malformed bootstrap and an over-bound bootstrap still produce
the same *screen*. What changes is that they stop producing the same *value*: two distinct named
reasons, one of which no other input can produce, asserted in a suite that executes. That is
defensible only because AC1 lands with it — once the build refuses an over-bound release, the only
way that reason can be reached at runtime is a tampered `index.html` or a deploy that skipped
`verify:offline`, neither of which is a product state the load screen should be redesigned around.
Changing the load screen's copy is not asked for by the epic, is not in D-8.5.1(a)'s scope, and
would be UI work in a story about a build gate. **If a reviewer disagrees, the disagreement is with
this paragraph, which is the point of writing it down.**

**6. What this story's gates can and cannot prove.** `npm run verify:offline` and
`verify:offline:red` execute the real gate against the real emitted release — this is one of the few
acceptance surfaces in this run that is genuinely *run* rather than compiled. Vitest executes the
parser and the reader. `test:e2e:compile` is `tsc --noEmit` and proves nothing was run; the 12
Playwright specs have never executed on this machine and CI does not invoke them (DW-101). **No AC
here is phrased at the browser layer**, and none needs to be: every claim this story makes is
observable in Node or jsdom. Say so at close rather than letting a compile pass read as a run.

**7. Both ends of the envelope, and why that is not scope creep.** The epic's AC names only the
maximum. The minimum (`minimumCacheAssets = 10`) fails identically and silently, is read by the same
helper from the same two adjacent lines, and is enforced by the same `if` with the same message
shape — enforcing one and leaving the other would need a second story to close a defect this one is
already standing on. It adds no new mechanism, no new authority and no new file. The maximum arm is
the one the epic requires to be red-proved; red-proving the minimum arm as well is cheap (filter
assets down rather than append) and recommended, but the maximum arm is the AC.

## Verification

**Baseline MEASURED at `548aa292294ac5093c4012fadeb3f141ee1248fc`, clean tree, in place — not
assumed.** Re-measure before the first edit and record any drift in `## Spec Change Log`. Record
every byte figure with its exact invocation, commit and tree state.

**Every figure below was MEASURED at `548aa29` this dispatch, not carried forward.**

**Commands:**

- `cd folio-go && go test -count=1 ./...` — measured **1815 pass / 2 fail / 5 skip**. The two are the
  standing red set below; nothing else.
- `cd folio-go && go test -count=1 -tags=matrix ./...` — measured **1826 pass / 3 fail / 5 skip**.
  The third is `TestShippedFacesReproduceFromUpstream` (a could-not-execute, see below).
- `cd folio-go && FOLIO_FONTGEN_PYTHON=/Users/panitw/Projects/folio/.fontgen-venv/bin/python go test
  -count=1 -tags=matrix -run TestShippedFacesReproduceFromUpstream ./...` — measured **PASS in
  8.23s, non-vacuously**: three `matches the recorded derivation` lines and
  `fontgen: derived and compared 3 of 3 faces`. **Sweep both ways and
  report both. If it fails WITH the variable set, HALT** — that is a real divergence in a committed
  face and it moves every font-bearing golden. Never edit the test; never put the variable in a
  committed file.
- `cd folio-go && go vet -tags=matrix ./...` — measured: no output.
- `gofmt -l folio-go` — **run from the REPO ROOT**; measured: no output. After a `cd folio-go` in the
  same compound command it prints an `lstat` error that reads like success; treat any `lstat` line
  as a non-measurement.
- `cd lint && go test -count=1 ./...` — measured: four `ok` (`cmd/genmanifest`, `internal/licence`,
  `internal/manifest`, `internal/rules`), no FAIL. **`-count=1` always**: these rules walk
  directories and Go's cache does not track `ReadDir`, so a cached `ok` is not a measurement.
  **Note this `cd` breaks later relative paths** — run the digest and README checks from the root.
- `cd folio-go && FOLIO_MATRIX_TARGET=<t> go test -count=1 -tags=matrix -run TestTargetRenderHash -v .`
  for each of `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` — measured: all four PASS
  (0.73s / 6.51s / 4.88s / 11.56s), **24 documents hashed per leg** (counted). Then the same command
  with the variable **unset** — measured: PASS in **0.00s** while asserting **nothing**, saying so at
  `matrix_test.go:2199`. **It is a control, never a fifth leg.**
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` — measured:
  PASS in 23.77s.
- `cd folio-designer && npm run typecheck` — measured: exit 0, clean.
- `cd folio-designer && npm run lint` — measured: **exactly 4** `only-export-components` warnings at
  `preview/pdf-viewer.tsx:16,17` and `App.tsx:1324,1331`. **The count of 4 and the rule name are the
  invariant; the line numbers move whenever `App.tsx` does.**
- `cd folio-designer && npm test` — use the project's own script. `npx vitest run` **from the repo
  root** picks up the 12 Playwright specs and produces a false mass-failure;
  `--reporter=basic` is not a valid reporter here and exits 1. Measured baseline: **39 files / 383
  tests, all passing** (16.45s). Expect that plus the tests this story adds.
- `cd folio-designer && npm run test:e2e:compile` — measured: exit 0. **This is
  `tsc -p tsconfig.e2e.json --noEmit`. Do not report it as a run.**
- **This story's own subject, and the gate that matters most here:**
  `cd folio-designer && npm run build`, then `npm run verify:offline:red` and
  `npm run verify:offline:wasm` — measured this dispatch: **all three exit 0** on node
  **v24.16.0** (`RELEASE_RUNTIME` pins it exactly; measured `v24.16.0`). `npm run build` already
  ends in `verify:offline`. Emitted release measured: **`release.assets.length` 23,
  `s1.assetCount` 23, `s1.cacheAssets` 23, `s1.cachedBytes` 38,460,398** — i.e. **23 against a bound
  of 64, 41 to spare.** **Do not relay any `s1VisibleBytes` figure without re-measuring it and
  quoting the command** (DW-100, D-8.4.27b); see the Read-only evidence note for the two disagreeing
  readings taken at this one commit.
- `shasum -a 256 fixtures/*/expected.pdf` — **run from the REPO ROOT**; measured: **23** lines.
  **The corpus is 23 documents and none may move; a changed digest is a defect, not a re-record.**
- `md5 -q README.md` — measured: `078d7d80d518d54af2fc04fb270d46b8`, unchanged.

**The standing reds, BY IDENTITY (D-8.4.34) — a count is a lossy set:**

1. `TestCorpusMeetsP6ExerciseFloors` and its subtest `P6g_(opaque_names)` in `internal/text`
   (`P6g floor not met: got 7, need >=20`), with drift twin `TestCorpusP6StatsMatchDeclaredBaseline`.
   **The one genuine standing red. Mandated permanent; never "fix" it.**
2. `TestShippedFacesReproduceFromUpstream` **is NOT a standing red.** Its failing assertion is
   verbatim `fontgen: fontTools is not importable by this interpreter` — a **misconfigured
   interpreter**, a could-not-execute that never compared bytes (DW-86). With
   `FOLIO_FONTGEN_PYTHON` set it passes non-vacuously.

**Any red beyond those two identities is a real failure.**

**23 golden digests, 24 AD-21 documents per leg, and the manifest's `assetCount` (23) are THREE
DIFFERENT POPULATIONS.** Do not conflate them, and do not "correct" one to another.

**Manual checks:**

- **Mutation-prove the new bound check by deleting it.** The red proof must report
  `escaped verification`. A guard that survives its own deletion is decoration.
- **Mutation-prove the reader's strictness, one mode at a time**, against a fixture source rather
  than the real file: constant absent → throws; only occurrence inside a comment → throws; declared
  twice → throws. Confirm each throws for **its own** reason, not a shared one.
- **Confirm the red proof fails on the right guard.** Temporarily drop the `expected` argument and
  observe which message the mutation actually produces; if it is `sameSet`'s or an identity check's,
  the placement is wrong. Restore `expected` afterwards.
- **Prove the over-the-bound reason is exclusive.** Run every rejection case in
  `release-payload.test.ts` and confirm exactly one produces the bound reason. Then mutate the bound
  check away and confirm that case reports some *other* reason rather than still reporting the bound.
- **Re-verify Design Note 4's control-flow premise before relying on it**: `render()` at
  `main.tsx:39` precedes the `try`; there is no `catch`; `registerOfflineLifecycle` publishes
  `'unavailable'` with no payload; the retry button is gated on `unavailable`. If any of that has
  moved, the "narrow rather than remove" argument must be re-derived, not assumed.
- **Confirm the dev bypass narrowing.** Under `import.meta.env.DEV`, an absent bootstrap still
  bypasses; a malformed one no longer does.
- **Confirm the S1 row pinning is byte-identical** to `548aa29` — `git diff` over
  `verify-offline-release.mjs` must show no change to the four pinning sites.
- **Confirm `maximumCacheAssets` is still 64** in the final diff. Raising it is out of scope and
  belongs to whichever story needs the headroom (D-8.5.1b).
- **Use explicit paths on every `git add`.** Never `git add -A` or `git add .`. Never push, never
  branch.

## Auto Run Result

Status: ready-for-dev
Blocking condition: none

Dispatch: classic intent, plan-only (`Halt after planning.`), at
`548aa292294ac5093c4012fadeb3f141ee1248fc`, tree clean, branch `main`. **No code was written.** The
only tracked file modified this dispatch is `epic-8-context.md`, recompiled by step-01 because
`epics.md` (16:01:46) was newer than the cache (13:04:47); that is the workflow's own step-1 output,
not pre-existing churn.

**Every premise the dispatch handed to this gate was checked. All came back TRUE:**

- `folio-designer/src/release-payload.ts:8` — `const maximumCacheAssets = 64`. TRUE.
- The bound is enforced only inside the seventeen-clause condition at `:13`, sharing one
  `return undefined` with fifteen unrelated shape checks. TRUE.
- `loadS1Payload:33` — `try { … } catch { return undefined }`, and `parseS1Payload` never throws, so
  the catch exists solely for `JSON.parse`. TRUE, and soft twice over as described.
- No build script checks the bound: `verify-offline-release.mjs` carries thirteen S1 assertions and
  none counts assets against it; `generate-offline-release.mjs` has no concept of it. TRUE.
- `s1-total-mismatch` (`:157`) and `s1-delivery-fiction` (`:158`) exist as red proofs, and
  `redProof` (`:126`) takes an `expected` argument documented at `:133`–`:135` for exactly the
  "held to its own guard" purpose. TRUE — the pattern to follow is real.
- The S1 row pinning is the four sites the dispatch named: `:58`–`:59` ordered exact equality,
  `:69` `cachedRows.length === 4`, `:75`–`:76` `rows[4]` positionally, `:78`–`:79` `cjk-font` as the
  `Math.max` of the font rows. TRUE — and it constrains rows, not assets.
- Measured `assetCount` of **23**. TRUE, twice: from the pre-existing `dist/`, and again from a
  fresh `npm run build` this dispatch. 41 assets of headroom.

**One premise was refined rather than falsified.** The dispatch called `verify-offline-release.mjs`
"the natural home", which it is — but the bound is declared in a **TypeScript** module the Node
verifier cannot import, and the sharing is one-directional by construction
(`offline-release-contract.mjs` imports `node:crypto`; `tsconfig.app.json` includes only `src`;
`src/` imports nothing from `scripts/`). So the assertion cannot simply reference the constant. The
spec resolves this by having the verifier **derive** the number from the declaration, on the
file's own established idiom of regex-extracting from a source artifact and failing when the pattern
cannot be located — because `npm run build` does not run Vitest, so a duplicated constant plus a tie
test would leave the build passing over drift. See Design Note 1.

**AC4 was ruled, not halted.** The dispatch's own decision procedure — prefer removal; if unsafe,
say why and make the causes distinguishable — selects between the readings, so this is not an intent
gap. Removal is unsafe, and the ground is measured control flow rather than caution: `main.tsx:39`
calls `render()` **before** the `try` at `:40`, the block has a `finally` and no `catch`, and
`startObservation` is invoked as `void startObservation()` at module scope. Today an unparseable
bootstrap reaches `registerOfflineLifecycle` with no payload, which publishes `'unavailable'`
(`offline-lifecycle.ts:45`) and gives the user *"Offline cache unavailable"* plus a **Retry
preparation** button. Remove the catch and that call is never reached: the screen sits on *"Checking
cache"* forever, no message, no retry — a *worse* silent outcome produced in the name of removing
silence. The spec therefore narrows the catch to `SyntaxError`, rethrows everything else, and gives
every rejection a named reason. Design Note 5 states plainly what that does and does not change for
a user, so a reviewer judges it rather than discovers it.

**`multiple-goals` was considered and deliberately not set** — see the frontmatter warnings section.

## Verification measured this dispatch

Baseline probes only at `548aa29`, clean tree, in place. No code was written, so nothing could
regress; these establish the floor the implementer measures against.

- `cd folio-go && go test -count=1 ./...` → **1815 pass / 2 fail / 5 skip**. The two are
  `TestCorpusMeetsP6ExerciseFloors` and its subtest `P6g_(opaque_names)`. Mandated permanent red.
- `cd folio-go && go test -count=1 -tags=matrix ./...` → **1826 pass / 3 fail / 5 skip**. The third
  is `TestShippedFacesReproduceFromUpstream`.
- Same test with `FOLIO_FONTGEN_PYTHON=/Users/panitw/Projects/folio/.fontgen-venv/bin/python` →
  **PASS in 8.23s, non-vacuously**: three `matches the recorded derivation` lines and
  `fontgen: derived and compared 3 of 3 faces`. **It is a misconfigured interpreter, never a byte
  divergence** (DW-86). No committed face has diverged; no golden is at risk on this ground.
- `go vet -tags=matrix ./...` → no output. `gofmt -l folio-go` (from the repo root) → no output.
- `cd lint && go test -count=1 ./...` → four `ok`, no FAIL, uncached.
- Four AD-21 legs PASS (`darwin/arm64` 0.73s, `linux/amd64` 6.51s, `linux/arm64` 4.88s, `js/wasm`
  11.56s), **24 documents hashed per leg** (counted). Unset control PASSES in **0.00s** asserting
  nothing at `matrix_test.go:2199` — a control, never a fifth leg.
- `TestCrossTargetByteIdentity` → PASS in 23.77s.
- Designer: `npm run typecheck` exit 0; `npm run lint` **exactly 4** `only-export-components`
  warnings (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1324,1331`); `npm test` **39 files / 383
  tests, all passing**; `npm run test:e2e:compile` exit 0 — a **compile**, not a run.
- `shasum -a 256 fixtures/*/expected.pdf` → **23** lines. `md5 -q README.md` →
  `078d7d80d518d54af2fc04fb270d46b8`.
- **This story's own subject:** `cd folio-designer && npm run build` → exit 0 (ends in
  `verify:offline`); `npm run verify:offline:red` → exit 0; `npm run verify:offline:wasm` → exit 0.
  Node `v24.16.0`. Emitted release: **23 assets, `s1.assetCount` 23, `s1.cachedBytes` 38,460,398**.
- `s1VisibleBytes` read twice at this one commit and the readings **disagree**: 12,428,555 from the
  pre-existing `dist/`, 12,426,352 after this dispatch's own build — a 2,203-byte variance,
  D-8.4.35a's still-unexplained drift. Quoted, never reasoned from (D-8.4.29).

  **⚠ SUPERSEDED (D-8.5.7). Not a variance between two builds at one commit** — a **stale `dist/` of
  unknown provenance** against one fresh build, with the stale side stamped from a **different commit
  and dirty flag**. The drift is **explained and FIXED** by Story 8.4g's `-buildvcs=false`; it is no
  longer "unexplained", and this figure should not be carried forward.

**Exactly the two standing red identities, no third.** 23 golden digests, 24 AD-21 documents per
leg, and the manifest's 23 assets are three different populations and are not conflated here.
