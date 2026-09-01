---
title: 'Story 8.4f: A bound nobody can cross silently'
type: 'bugfix'
created: '2026-09-01'
status: 'done'
baseline_revision: '8dd6c6e8235008fdf79abf742868885974fcc091'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: [oversized]
deferred:
  - summary: >-
      declaredCacheAssetBounds() reads src/release-payload.ts at verify time, so verifying a dist/
      detached from its source tree throws an unframed Node ENOENT instead of a stated verification
      failure, and an old dist/ verified against a newer src/ is silently checked against the new bounds.
    evidence: |-
      The manifest carries no record of the bound the release was built under; the bound is resolved
      from the working tree at the moment of verification, not from the artifact. Reachable by an
      extracted release or a deploy-stage-only CI job. Not reachable in this repo's current cadence,
      where verify:offline always runs beside its own source tree.
    location: >-
      folio-designer/scripts/offline-release-contract.mjs
    severity: low
  - summary: >-
      The DW-106 tree-state probe file survives a SIGINT/SIGKILL landing between the two ~40s engine
      builds, leaving an untracked file in folio-go/ that trips the pipeline's own clean-tree gates.
    evidence: |-
      Cleanup is a finally block, which a signal does not run. The obvious remedy - a .gitignore entry -
      is specifically WRONG here and is now warned against in the code: the probe's visibility to git IS
      the perturbation being measured, so ignoring it would make the check vacuous. The correct remedy is
      a signal handler, which was out of proportion to this story.
    location: >-
      folio-designer/scripts/verify-offline-release.mjs
    severity: low
---

## Frontmatter warnings — what they mean here

**`followup_review_recommended` was set `true` by the build (2 medium + 3 low patched) and is CLEARED
at close (2026-09-02) — by the closer's own adversarial pass, not on the build's or the orchestrator's
judgement.** The flag records whether a follow-up review is *owed*, and one was performed: both flagged
guards were re-verified by mutation. **One held** (the DW-106 probe-visibility assertion, re-proved to
fire when the probe is hidden from git). **One did not** — `main.tsx`'s two startup decisions were found
to be **narrowed, not closed**: their extracted predicates are now executed and have teeth, but the call
sites are still run by no gate, and the exact mutation the review demonstrated **still ships green
today, measured**. That residual is **not left in this flag** — it is filed as **DW-110** with a named
owner. See `## Delivery Log` for both measurements.

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

*Non-normative — a plain-language summary of what shipped. The intent contract below governs
implementation.*

The designer's first-run load screen reads a small summary the build stamps into the page. The reader
refuses a summary listing more than **64** cached items; the release lists **23**. Nothing counted
that number at build time, and had a future release crossed it, the first screen a user sees would
have lost its summary in silence — no progress, no message, nothing said.

The build now counts, and refuses to finish when the number falls outside the declared range, naming
both the count and the limit it crossed. It takes that limit from the one place that declares it
rather than keeping a second copy, and stops outright if that declaration becomes unreadable.
Releases manufactured outside the range were shown to fail, at both ends. The reader also stopped
answering every different problem with the same silence: each refusal now carries its own name, and
the one meaning "over the limit" can be produced by nothing else.

Three things will look wrong and are deliberate. Nothing was raised, renamed, added or removed — the
headroom stays for whichever story needs it. One acceptance criterion asked for a failure worded a
way its own design note shows to be unreachable; that wording was corrected at close, loudly and with
the original kept, rather than quietly satisfied. And the startup code consuming all of this is still
run by no test here: the decisions it makes are proved elsewhere, the wiring is not, and that is
recorded rather than closed.

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
  deleted from the source, when the red proof runs, then the `asset-count-over-bound` leg still fails
  the gate non-zero, **on a message that is not the new guard's** — the guard does not survive its own
  deletion, and nothing else in the verifier measures the count.

> **⚠ AC3's SECOND CLAUSE WAS AMENDED AT CLOSE, 2026-09-02, BY THE STORY CLOSER. SAYING SO, WITH THE
> ORIGINAL PRESERVED, BECAUSE AN AC REWRITTEN TO MATCH AN IMPLEMENTATION IS NORMALLY THE DEFECT.**
>
> **The original wording, verbatim:** *"And given the new check is deleted from the source, when the
> red proof runs, then it fails as `escaped verification` — the guard does not survive its own
> deletion."*
>
> **Why this is D-8.4.2 (an error to correct) and NOT D-8.4.24 / D-8.4.29 (a threshold moved to match
> a measurement).** Those two rulings forbid moving a **bar** to match a **result**. The bar here is
> AC3's own headline — *the guard is red-proved and proved to discriminate; it does not survive its own
> deletion* — and that bar is **met, measured, and unmoved**. What was wrong is a **factual premise
> about the mechanism**: the original clause assumes an over-bound manifest violates **only** the
> bound, so that removing the bound guard leaves nothing to catch it. **`## Design Notes` 3, in this
> same spec, states the opposite and explains why** — an over-bound manifest necessarily also breaks
> the manifest/output exact-set check, and the digest and Brotli loops besides. So the spec asserted a
> proposition in `## Tasks & Acceptance` that it refuted in `## Design Notes`. The contradiction is
> **AC3 against Task 4 and Design Note 3, both inside this document** — not AC3 against reality.
>
> **The phenomenon is OVERDETERMINATION of the mutation, not ordering of the guards.** The new guard is
> the **earliest**, which is why the positive proof works. On deletion, a **later and independent**
> invariant catches the same mutation, because a manifest carrying 65 assets violates two invariants at
> once. `redProof`'s `escaped verification` string means *no failure at all*, which no faithful
> over-bound mutation of this shape can produce.
>
> **The literal string is reachable, at a measured price — so "unsatisfiable" would overstate it.** The
> regenerate-after-mutation shape (write real synthetic asset files, then regenerate so every other
> figure is restored and the count is the only remaining fault) would produce it. **Measured
> independently at close: `npm run build:offline` takes 45.73s and `npm run verify:offline:red` at
> `68c548e` takes 183.33s, so that shape adds ~91s — about +50% — to every red run, for a different
> word in a failure message.** Rejected on that cost, recorded rather than silently chosen.
>
> **The amended clause is not weaker than the original — it is checkable, and it is what was measured.**
> See `## Delivery Log` for both directions, run at close.
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

**Re-measured at `8dd6c6e8235008fdf79abf742868885974fcc091` before the first edit, clean tree, branch
`main`.** The spec's anchors were taken at `548aa29`; `8dd6c6e` is that commit plus this spec's own
repair commit, and **every cited line anchor held exactly** — `release-payload.ts:7-8` (`const
minimumCacheAssets = 10` / `const maximumCacheAssets = 64`), `:13` (the seventeen-clause condition),
`:30-34` (`loadS1Payload`), and `verify-offline-release.mjs:12` (`fail`), `:32` (end of the manifest
URL loop), `:33-34` (`runtimeOutputUrls` / `sameSet`), `:126-137` (`redProof`), `:139-174`
(`runRedProofs`). **No anchor drift to record.**

**Block If premise, re-measured, holds.** `cd folio-designer && npm run build` (exit 0) then reading
`dist/offline-release-manifest.json`: **`release.assets.length` 23, `s1.assetCount` 23,
`s1.cacheAssets` 23, `s1.cachedBytes` 38,460,370, `s1VisibleBytes` 12,425,468** — 23 against a
declared bound of **64**, 41 to spare. (`s1VisibleBytes` is quoted with its command and is not
reasoned from, per D-8.4.29. It differs from the figures in this spec's Read-only-evidence passage,
which D-8.5.7 already ruled unattributed; this reading is from one build at one commit.)

**1. AC3's second clause was met in substance and NOT in its literal wording — measured both ways.**
AC3 asks that, with the new check deleted, the red proof "fails as `escaped verification`". With the
proof shape **Task 4 prescribes** (`:142`-style: append synthetic assets, `rewriteRelease`, restore
from buffers) that string is unreachable, and **Design Note 3 already says why**: an over-bound
manifest necessarily also breaks `sameSet`, so `verifyOfflineRelease` still throws and `redProof`'s
"something failed" test is still satisfied. What AC2 additionally mandates — passing `expected` —
then catches it. Measured with the guard commented out:

- `npm run verify:offline:red` reddens on the new leg with
  `red proof asset-count-over-bound failed for the wrong reason: offline release verification failed:
  manifest and production runtime output are not an exact set`.
- A direct probe of `verifyOfflineRelease` over the same mutated manifest reports
  `manifest and production runtime output are not an exact set` for **both** the over-bound and the
  under-bound release — i.e. **nothing else in the file measures the count**.

So the guard does not survive its own deletion (the gate exits non-zero on the new leg, by name), and
nothing else in the verifier covers for it. Only the failure *string* differs from AC3's wording. The
alternative that would produce the literal string is the `:157`/`dev-bypass-shipped` shape — write the
synthetic asset files into `dist/assets` and call `generateOfflineRelease` so every other figure is
regenerated and the count is the only remaining fault. It was **measured and rejected on cost**: one
`npm run build:offline` takes **43.2s** (Brotli quality 11 over the full ~38 MB release), and that
shape needs two of them, adding **~86s to every `verify:offline:red` run** — for a different word in
a failure message. Recorded here rather than silently chosen.

**2. Guard added that the spec did not name, to honour a Boundary it did name.** `JSON.parse('null')`
SUCCEEDS, so once the `try` is narrowed to the parse alone, reading `.s1` off the result would throw a
`TypeError` that the narrowed catch no longer admits — changing what a user sees on a path that is
neither the over-the-bound path nor malformed JSON, which the Block If forbids. `loadS1Payload` now
guards the property read so a `null` bootstrap still reaches the parser and still rejects
(`not-an-object`), exactly as it did before. Asserted in `release-payload.test.ts`.

**3. Both bound arms carry SEPARATE reasons.** Design Note 7 enforces both ends with one message
shape; the parser goes one step further and reports `asset-count-over-maximum` and
`asset-count-under-minimum` as **distinct** reasons, so AC4's exclusivity claim ("produced by that
cause and by no other") is literally true of the over-the-bound reason rather than true of a shared
"out of bounds" value.

**4. T-A (DW-107) and T-B (DW-106), as routed by the orchestrator — how they landed.**

- A **single authority for the engine's compile** was factored into
  `folio-designer/scripts/wasm-vcs-stamp.mjs` (`ENGINE_BUILD_FLAGS`, `buildEngineWasm`), and
  `build-wasm.mjs` now calls it. This is the same doctrine AC5 imposes on the bound: the verifier must
  not re-type the argv it is proving something about. Nothing asserts over `build-wasm.mjs`'s go
  invocation, so the move is safe (`font-binary-identity.test.ts` and `canvas-font-stack.test.ts` read
  that file only for its `@font-face` rules).
- **T-B** is `assertEngineWasmIsTreeIndependent()`, run inside `verifyOfflineRelease`'s existing
  `wasmWitness` block — the expensive-witness section, reached by `npm run verify:offline:wasm` and
  **not** by `npm run build`. It builds the engine twice, writes one stray untracked file between the
  runs, and requires the digests to agree, reporting both. Its comment states its limit rather than
  overclaiming it: it holds the checkout path fixed and **would not have caught DW-105**.
- **T-A** is `proveVCSStampGuardDiscriminates()`, called from `runRedProofs`. It deliberately does
  **not** go through `redProof`: that harness mutates `dist/` and re-runs `verifyOfflineRelease`, and
  the guard under proof lives in `build-wasm.mjs`. The shape is kept — build with the flag dropped,
  observe a failure, hold it to the guard's own message (all four settings named), restore — and the
  probe builds to a temp file, so `src/generated/` is never touched and there is nothing to undo by
  hand. It fails loudly rather than passing vacuously when the checkout has no `.git`.
- **DW-106 and DW-107 are discharged by this story but their register entries were NOT edited** — the
  spec forbids this story touching `deferred-work.md`, so closing them is left to the reviewer/closer.

**5. Final measurement, after the change, at this dispatch's last build.** `cd folio-designer &&
npm run build` exit 0 (it ends in `verify:offline`); `npm run verify:offline:wasm` exit 0 (it now also
runs the DW-106 property check); `npm run verify:offline:red` exit 0 (it now also runs the two bound
legs and the DW-107 stamp proof). Emitted release: **`release.assets.length` 23, `s1.assetCount` 23,
`s1.cacheAssets` 23, `s1.cachedBytes` 38,460,805, `s1VisibleBytes` 12,425,468.** `cachedBytes` moved
+435 from the pre-change build in this same dispatch **because this story changed `src/main.tsx` and
`src/release-payload.ts`, so the emitted bundle and therefore `index.html` are larger** — the figure
is a sum that includes `index.html`'s own size. `s1VisibleBytes` (the four cached rows: engine wasm
and three fonts) is **unchanged at 12,425,468** across both builds, which is what a source-only change
to TypeScript should do. Neither figure is reasoned from (D-8.4.29); both are quoted with the build
that produced them.

## Review Triage Log

### 2026-09-01 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 5: (high 0, medium 2, low 3)
- defer: 2: (high 0, medium 0, low 2)
- reject: 24: (high 0, medium 0, low 24)
- addressed_findings:
  - `[medium]` `[patch]` **`main.tsx`'s two new decisions were executed by no test.** Nothing in the
    repo imports `main.tsx`; Vitest collects only `src/**/*.test.{ts,tsx}` and `scripts/**/*.test.mjs`,
    and the Playwright suite is compile-only. Demonstrated: replacing `payload = result.ok ? … :
    undefined` with `payload = undefined` type-checks and leaves every gate green while shipping an app
    permanently stuck on "Offline cache unavailable". This was also the one I/O-matrix row (row 8, the
    dev bypass) the step-03 matrix audit could not tie to an executed test. Fixed by extracting
    `payloadForLifecycle` and `isDevBypassReason` as pure exported functions and driving the bypass
    predicate over the rejection table, with the reason union held to a
    `Readonly<Record<S1PayloadRejection, true>>` so a future unnamed reason is a typecheck failure.
    Mutation-proved both ways.
  - `[medium]` `[patch]` **The DW-106 tree-independence check could not tell "tree-independent" from
    "my probe perturbed nothing".** Both arms build with `-buildvcs=false`, which closes the only
    measured tree→binary channel, so the stray file had no route into the output — an all-clear
    indistinguishable from a couldn't-look, inside this story's own fix for that defect. Fixed by
    asserting the probe is visible to `git status --porcelain` before relying on the second build, and
    by stating in place what the check does and does not establish, including that the probe's
    binary-moving capability was measured at Story 8.4g's close and is not re-measured here.
    Mutation-proved by hiding the probe via `.git/info/exclude` (never `.gitignore`).
  - `[low]` `[patch]` `readDeclaredConstant`'s failure message hardcoded `src/release-payload.ts` even
    when a source string was injected — which is how every failure test drives it — so failures
    misattributed themselves to the real file. Now takes a source label.
  - `[low]` `[patch]` Nothing checked that `minimumCacheAssets <= maximumCacheAssets`; an inverted pair
    would have blamed the release for an incoherent declaration. Now fails naming both numbers and the
    declaration as the fault.
  - `[low]` `[patch]` One test asserted two mutations (commented-out and indented) under a name
    describing one, contradicting that file's own "one mutation at a time" principle stated two tests
    above. Split.

**Rejected findings, enumerated with the ground that refuted each (DW-87).** Two were refuted by
direct measurement rather than by argument:

1. **CRLF fragility of the derived-constant reader** (`offline-release-contract.mjs`) — claim: `$`
   under `/m` will not match before `\r`, so a CRLF checkout reads zero matches and fails the build.
   **Refuted by measurement:** JS `$` in multiline mode matches before `\r` as well as `\n`; a CRLF
   source yields exactly 1 match, identical to LF.
2. **`rewriteRelease` may leave `index.html` drifted for proofs that run after the two bound legs**
   (`verify-offline-release.mjs`) — **refuted by reading the function**: it writes only
   `offline-release-manifest.json` and `sw.js`, which are exactly the two files both new proofs
   snapshot and restore.
3. **The production path drops the named reason (no log, telemetry or distinct message)** — raised
   independently by three layers. Rejected on the authority of **Design Note 5**, which rules this
   explicitly and invites the disagreement in writing, plus the intent's **Block If** forbidding any
   user-visible change outside the over-the-bound path. There is also no diagnostic channel to write
   to: `src/` and `scripts/` contain zero `console.*` calls (DW-93).
4. **`release-payload.test.ts` re-types `65`/`9`, a second authority** — rejected because that drift
   fails **loud**, not green: a fixture that stopped being over-bound reports `accepted` and the
   assertion reddens. The doctrine this story enforces is against drift that ships green.
5. **An empty bootstrap node reads as `no-bootstrap`** — pre-existing and unchanged; the prior code was
   `if (!node?.textContent) return undefined` and the prior bypass fired on any falsy payload. No
   regression caused by this change. (Same ground rejects the whitespace-only `trim()` variant.)
6. **`JSON.parse` `RangeError` should also map to `malformed-json`** — the spec instructs verbatim that
   a `SyntaxError` becomes the malformed reason and **anything else is rethrown**; widening contradicts
   an explicit instruction, and the bootstrap is build-emitted rather than user input.
7. **`GOFLAGS` could re-add `-buildvcs=false` to the flag-dropped probe** — fails **safe**: the probe
   would report `escaped verification` and fail the build loudly. Never a false pass.
8. **Split `S1PayloadRejection` into parse/load unions** — type-modelling preference; the contract asks
   for a named reason per rejection, which is satisfied.
9. **`goModuleRoot` / `.git` path arithmetic re-derived in the caller** — cosmetic; the no-re-typing
   doctrine applies to the value being proved, not to path joins.
10. **Three vocabularies (`over-bound` / `over-maximum` / `declared maximum`)** — cosmetic; each name
    is apt at its own layer.
11. **`proveVCSStampGuardDiscriminates()`'s position between two `redProof` calls** — no defect; it
    mutates nothing in `dist/` and restores nothing, so ordering is immaterial.
12. **The change log quantifies the rejected option's cost but not the accepted one's** — documentation
    completeness, not a defect in the artifact.
13. **The unreadable-declaration guard is only asserted in Vitest, which `npm run build` never runs** —
    the **guard itself** is on the build path (the throw propagates through `declaredCacheAssetBounds()`
    during `verify:offline`); only its failure-mode tests live in Vitest. The drift argument the intent
    makes is about a duplicated **value**, and there is no duplicate.
14. **Exclusivity is proved over a fixture table, not the input space** — the test names itself
    honestly and AC4 asks verbatim for "no other input asserted in `release-payload.test.ts`".
15. **The baseline moved from `548aa29` to `8dd6c6e`** — the dispatch directed re-measurement at HEAD;
    the premise (23 assets against a bound of 64) was re-verified and recorded.
16. **`minimumCacheAssets < 3` would break the under-bound proof's slice** — hypothetical on a constant
    no one is changing, and the failure would be a loud build failure. Partly covered by the new
    coherence check.
17. **A missing `git` binary, or a missing Go toolchain, escapes the `fail()` contract** — `go` is
    already a hard prerequisite of the whole gate (`npm run build` begins with `build:wasm`), and the
    new probe-visibility check now fails in the file's own voice when git cannot be run.
18. **AC3 is unmet literally and nothing records it** — moot; it is recorded here and in
    `## Auto Run Result`, and reported to the orchestrator.
19. **DW-106 / DW-107 are discharged in code but left open in the register** — correct by the spec's
    own prohibition on this story touching `deferred-work.md`; the closer owns it, and it is surfaced
    in the dispatch report.
20. **The DW-106/DW-107 scope is invisible in the `<intent-contract>`** — accurate as an observation,
    but the orchestrator routed T-A and T-B into `## Tasks` as settled decisions and the contract is
    frozen. Surfaced upward rather than actioned.
21-24. Four further cosmetic or restated observations from the blind and edge-case layers (comment
    wording, naming, ordering, and a restatement of the production-reason finding) carrying no distinct
    required action.

**Population check.** The four layers returned **38** raw findings (blind hunter 18, edge-case hunter
10, verification-gap 3, intent-alignment 7 divergence observations). Deduplication merged **7** into
other claims, leaving **31** distinct claims. Routes: 5 patch + 2 defer + 24 reject = **31**. The
routes sum to the declared population.

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


---

## Auto Run Result — build dispatch, 2026-09-01

Status: **done**
Blocking condition: **none**

Dispatch: existing spec at `status: ready-for-dev`, so step-01 routed straight to step-03 and nothing
was re-derived. Baseline `8dd6c6e8235008fdf79abf742868885974fcc091`, tree clean, branch `main`.
Commits `7a18079` (implementation) and the review-patch commit that follows it. Nothing pushed, no
branch created, no `git add -A`.

### Summary of the implemented change

The offline release verification job now counts the emitted release's assets against the bound the
TypeScript parser declares and fails the build naming both the measured count and the bound crossed,
at both ends of the envelope. The number is **derived** from the `const` lines in
`src/release-payload.ts` rather than re-typed, because `npm run build` never runs Vitest and a
duplicate-plus-tie-test would ship green through the very gate this story adds; the reader is
line-anchored, requires exactly one live match, and fails loudly on a renamed, reformatted,
commented-out, duplicated or inverted declaration. The check sits immediately after the manifest URL
loop and **before** the manifest/output set comparison, because an over-bound manifest breaks that set
too and a proof failing there would prove nothing about the bound. Both arms are red-proved with
`redProof`'s `expected` argument.

`parseS1Payload` now returns a discriminated result in which every rejection carries its own name, with
the bound moved out of the seventeen-clause shape condition into its own checks so
`asset-count-over-maximum` is reachable by that cause and no other. `loadS1Payload`'s blanket catch is
narrowed to the `JSON.parse` `SyntaxError` it existed for and rethrows everything else; it is narrowed
rather than removed because removal is measurably worse for a user, and the dev-server bypass is gated
on the absent-bootstrap reason alone.

T-A (DW-107) and T-B (DW-106), routed here from Story 8.4g's close, are both discharged in
`verify-offline-release.mjs`: the flag-dropped stamp red-proof is now executable rather than prose, and
the determinism guard now asserts the **property** (two builds of one commit agree) alongside the
cause, with its limit stated in place.

### Files changed

- `folio-designer/scripts/offline-release-contract.mjs` — new `declaredCacheAssetBounds()`; derives both
  bounds from the declaration, exactly-one-live-match, labelled messages, envelope-coherence check.
- `folio-designer/scripts/offline-release-contract.test.mjs` — the derived reader's failure modes, one
  mutation per case, asserted against an independent re-read rather than re-typed literals.
- `folio-designer/scripts/verify-offline-release.mjs` — the envelope guard at its pinned insertion
  point; two red proofs; `proveVCSStampGuardDiscriminates` (T-A) and
  `assertEngineWasmIsTreeIndependent` (T-B) with its probe-visibility control.
- `folio-designer/scripts/wasm-vcs-stamp.mjs` — `ENGINE_BUILD_FLAGS` / `buildEngineWasm`, one authority
  for the engine's compile argv so neither new proof re-types it.
- `folio-designer/scripts/build-wasm.mjs` — calls that single authority.
- `folio-designer/src/release-payload.ts` — discriminated result, eleven named reasons, narrowed catch,
  null-bootstrap guard, and the two extracted pure helpers.
- `folio-designer/src/release-payload.test.ts` — reasons, exclusivity over the whole rejection table,
  the rethrow, and the previously untested startup decisions.
- `folio-designer/src/main.tsx` — maps a rejection to `undefined` at one place; dev bypass narrowed.

### Review findings breakdown

5 patched (2 medium, 3 low), 2 deferred, 24 rejected, 0 intent_gap, 0 bad_spec. The four layers
returned 38 raw findings; dedup merged 7, leaving 31 distinct claims, and 5 + 2 + 24 = 31.
Every rejected finding is enumerated with its refuting ground in `## Review Triage Log`; two of them
were refuted by direct measurement rather than argument.

**Follow-up review recommended: true.** Patched findings only: medium 2, low 3 → 3x2 + 1x3 = **9**,
which is >= 5.

### Verification performed

Measured at `8dd6c6e` before the first edit and again after the patches; exit codes read from the
command itself, never through a pipe.

- `go test -count=1 ./...` **1815 pass / 2 fail / 5 skip**, unchanged from baseline.
  `-tags=matrix` **1826 / 3 / 5**. Reds by identity: `TestCorpusMeetsP6ExerciseFloors` +
  `P6g_(opaque_names)`, and matrix-only `TestShippedFacesReproduceFromUpstream`, a could-not-execute
  which **passes non-vacuously** with `FOLIO_FONTGEN_PYTHON` set (`derived and compared 3 of 3 faces`).
  Exactly the two standing red identities, no third.
- `go vet -tags=matrix ./...` no output. `gofmt -l folio-go` **from the repo root** no output.
  `cd lint && go test -count=1 ./...` four `ok`, uncached.
- Four AD-21 legs PASS, **24 documents hashed per leg**; the unset control passes in 0.00s asserting
  nothing at `matrix_test.go:2199` — a control, never a fifth leg. `TestCrossTargetByteIdentity` PASS.
- Designer: `typecheck` exit 0; `lint` exit 0 with **exactly 4** `only-export-components` warnings and
  no other rule; `npm test` **40 files / 409 tests all passing** (baseline 40 / 387);
  `test:e2e:compile` exit 0 — a **compile**, not a run.
- This story's own subject: `npm run build` **exit 0** (ends in `verify:offline`),
  `verify:offline:red` **exit 0**, `verify:offline:wasm` **exit 0**, on node `v24.16.0`.
  Emitted release: **`release.assets.length` 23, `s1.assetCount` 23, `s1.cacheAssets` 23,
  `s1.cachedBytes` 38,460,833**; `s1VisibleBytes` (a **top-level** manifest key —
  `release.s1.s1VisibleBytes` is `undefined` and looks like absence) reads **12,425,468**, quoted with
  its command and never reasoned from (D-8.4.29).
- `shasum -a 256 fixtures/*/expected.pdf` from the repo root: **23** lines, **byte-identical** to the
  baseline capture. `md5 -q README.md` `078d7d80d518d54af2fc04fb270d46b8`, unchanged.
- **Independent mutation proofs run by the dispatch itself, not taken on report:** replacing the
  over-bound proof's `expected` with an impossible needle showed the mutation trips the **new guard**
  (`release carries 65 cache assets, over the declared maximum of 64`), not `sameSet` — clearing the
  Block If. Commenting the guard out showed the proof still reddens, on `sameSet`'s message. The
  derived reader was driven against fixture sources for the absent, commented-out, duplicated and
  reformatted cases, each throwing with its own named case.

### Residual risks

1. **AC3's second clause is met in substance but NOT in its literal wording.** With the proof shape
   Task 4 prescribes, deleting the guard reddens as `failed for the wrong reason` rather than
   `escaped verification`, because an over-bound manifest also breaks `sameSet` — which Design Note 3
   itself predicts — and AC2's mandated `expected` then catches it. The guard demonstrably does not
   survive its own deletion, and nothing else in the verifier measures the count. The shape that would
   produce the literal string was measured and rejected at ~86s added to every `verify:offline:red`.
   **The AC text was deliberately NOT amended to match the code.**
2. **DW-106 and DW-107 are discharged in code but their `deferred-work.md` entries are untouched**, as
   this spec's Boundaries require. Closing them belongs to the closer.
3. `verify:offline:red` and `verify:offline:wasm` now shell out to `go build` (once and twice
   respectively). `go` was already a hard prerequisite of `npm run build`, but these two commands now
   carry that dependency as well, and are correspondingly slower.
4. No claim here is phrased at the browser layer. The 12 Playwright specs still never execute on this
   machine and CI runs only `tsc --noEmit` (DW-101).

---

## Delivery Log

### 2026-09-02 — done

*(The close ran across the local date boundary: gates began 2026-09-01, the record and the closing
commit landed 2026-09-02. Every other date in this spec is the build's own day, 2026-09-01.)*

Baseline `8dd6c6e`. Shipped in `7a18079` (implementation) and `68c548e` (review patches); closed by
this file's own commit on `main`, **unpushed** — `origin/main` is at `c985b9c`, and the commits above
it are the owner's to push. Never pushed, never branched, no `git add -A`.

**THE RUN PAUSES HERE at the owner's request.** This entry is written to be resumable cold: what is
still owed, and by whom, is spelled out at the end rather than left to be re-derived.

**What shipped.** The offline release verification job now counts the emitted release's assets against
the envelope the TypeScript parser declares, and fails the build naming both the measured count and the
end it crossed. The number is derived from the declaration rather than re-typed, line-anchored, and
required to match exactly once; a renamed, reformatted, commented-out, duplicated or inverted
declaration fails the build loudly. The check is placed before the manifest/output set comparison so
its red proof can fail on its own message. Every parser rejection acquired a distinct name, the bound
moved out of the seventeen-clause shape condition so its reason is reachable by that cause alone, and
the blanket runtime catch was narrowed to the one throw it existed for. Nothing was raised, added,
renamed or removed: `maximumCacheAssets` is still **64**, and the S1 row pinning is byte-identical to
`548aa29` — verified at close by extracting every `s1.`-bearing line from both revisions and diffing
them; the only difference is two added comment lines, and no pre-existing line changed.

**Decisions applied by ID.** D-8.5.1(a) scoped the story and forbade the split; D-8.5.1(b) reserved
raising the bound for whichever story needs the headroom. D-8.4.30's proxy-versus-purpose defect is what
T-B answers. D-8.4.31 is why T-A and T-B were routed into `## Tasks` rather than into prose. D-8.4.29
governs every byte figure quoted below. D-8.5.7 supersedes this spec's own 2,203-byte passage. D-8.4.34
governs the standing reds by identity. D-8.4.2 is the ruling under which AC3 was corrected; D-8.4.24 and
D-8.4.29 are the two it was tested against and found not to be.

---

#### The ruling this close was asked for: AC3

**AC3 is MET.** Its first clause and its headline are met and measured. Its second clause carried a false
premise and **was corrected in place at close, loudly, with the original preserved verbatim** — see the
amendment block under `## Tasks & Acceptance`.

**Both directions re-measured at close, at `68c548e`, clean tree — not taken on the build's report:**

- **Guard present, `expected` replaced with an impossible needle** (`cd folio-designer && npm run
  verify:offline:red`): **exit 1**, `red proof asset-count-over-bound failed for the wrong reason:
  offline release verification failed: release carries 65 cache assets, over the declared maximum of 64`.
  The mutation trips **the new guard**, with its own message. **The Block If is clear and the placement
  is right.**
- **Guard neutered in place** (`if (false && …)`), `expected` restored: **exit 1**, `red proof
  asset-count-over-bound failed for the wrong reason: offline release verification failed: manifest and
  production runtime output are not an exact set`. **The guard does not survive its own deletion** — the
  gate fails non-zero on the new leg, by name — but not with the word AC3 asked for.

**The ruling, stated so it can be disagreed with.** The orchestrator's provisional reading — *the guard
is shown to discriminate; what cannot be done is proving it through that particular route, because an
earlier guard legitimately catches the same mutation first* — is **right in its conclusion and wrong in
its mechanism, and the mechanism matters.** No earlier guard preempts anything: the bound guard **is**
the earliest, which is exactly why the positive proof produces its own message. What happens on deletion
is that a **later and independent** invariant catches the mutation, because a manifest carrying 65
assets **violates two invariants at once**. That is **overdetermination of the mutation**, not ordering
of the guards. `escaped verification` means *nothing failed at all*, and no faithful over-bound mutation
of this shape can produce it — the second invariant is not optional.

**"Unsatisfiable" would also overstate it, and the overstatement is worth removing.** The literal string
**is** reachable, by the regenerate-after-mutation shape that restores every other figure and leaves the
count as the only fault. It was rejected on cost, and **the cost was re-measured at close rather than
relayed**: `cd folio-designer && npm run build:offline` → **45.73s real**; `cd folio-designer && npm run
verify:offline:red` at `68c548e` → **183.33s real, exit 0**. Two extra generations add **~91s, about
+50%**, to every red run — for a different word in a failure message. So AC3's three clauses **can** all
hold; they cannot all hold **under the proof shape Task 4 prescribes**. The contradiction is between AC3,
Task 4 and Design Note 3 — all three inside this one document.

**Which of the two rules this is, and why.** It is **D-8.4.2** — an error to correct — and **not**
D-8.4.24 / D-8.4.29. Those forbid moving a **bar** to match a **result**; the bar here (*the guard is
red-proved, discriminates, and does not survive its own deletion*) is met, measured and unmoved. What was
wrong is a **factual premise about the blast radius of the mutation**, which this spec's own Design Note 3
already refutes. The corrected clause is **checkable and not weaker** — it names which guard caught the
deletion, where the original named a string that no faithful mutation can produce. The original is
preserved verbatim in the amendment block so nobody has to take this on trust.

---

#### `followup_review_recommended: true` — the hard adversarial pass, and its outcome

**Cleared on this closer's own measurements, not on the build's report or the orchestrator's judgement.**
Both flagged guards were re-verified independently. **One held. One did not, and the disposition below is
a correction to the build's own account.**

**1. The DW-106 probe-visibility assertion — VERIFIED; the patch closed it.** The code says what it
claims: the check reads `git status --porcelain` at the repository root before relying on the second
build, fails in the file's own `fail(…)` voice when git cannot be run at all, and its comment carries the
⚠ prohibition that the probe filename must never be added to `.gitignore`, with the reason (Go derives
`vcs.modified` from git, so an ignored probe makes both builds see an identical tree).
**Re-proved by mutation at close:** `.folio-tree-state-probe-*` appended to `.git/info/exclude` (never
`.gitignore`), then `cd folio-designer && npm run verify:offline:wasm` → **exit 1**, `offline release
verification failed: the tree-state probe folio-go/.folio-tree-state-probe-71235 is invisible to git, so
both builds below see the same tree and their agreement would assert nothing`. The exclude file was then
restored byte-identically, re-confirmed with `git check-ignore` (the pattern is **not** ignored at HEAD),
and no stray probe was left behind. **The check can no longer pass by perturbing nothing.** Its remaining
limit is stated in the code rather than hidden: both arms still build with `-buildvcs=false`, so the only
measured tree→binary channel is closed and the check is *expected* to be quiet — it watches for a future
input, not for the closed one.

**2. `main.tsx`'s payload mapping and dev-bypass narrowing — the patch NARROWED the gap. It did NOT close
it, and the build's account reads as though it did.** This is the one correction this close makes to the
dispatch report.

What the patch genuinely bought, and it is real: both decisions are now pure exported functions that
Vitest executes over the whole rejection table, the reason union is held to an exhaustive record type so
a future unnamed reason is a typecheck failure, and **both have teeth — proved at close, not assumed**:

- Neutering the parser's over-maximum check (`if (false && …)`) reddens **five named tests** in
  `release-payload.test.ts`, and reddens them for the right reason — the over-bound fixture becomes
  `accepted`, so the new behaviour is the thing under test and no sibling guard is covering for it.
- Re-widening the bypass predicate to `!result.ok` reddens exactly *"fires the dev bypass on the
  absent-bootstrap reason and on no other reason that exists"* (1 failed / 14 passed).

What the patch did **not** buy, measured at close: **the two call sites in `main.tsx` are still executed
by nothing, and the exact mutation the build demonstrated still ships green today.**

- **A first attempt at the mutation was rejected as unfaithful, and the trap is worth recording.**
  Replacing the mapping with a bare `undefined` orphans the now-unused import, and `tsc -b` fails with
  **TS6133 `'payloadForLifecycle' is declared but its value is never read`**. That is an incidental
  compile error, **not a test detecting the defect** — a mutant that will not compile proves nothing.
- **Reinstated faithfully** (mapping neutered **and** the orphaned import dropped, which is what a real
  refactor would do): `npm run typecheck` **exit 0**; `npm test` **exit 0, 40 files / 409 tests passed**;
  `npm run lint` **exit 0**; `npm run test:e2e:compile` **exit 0**; `npm run build` **exit 0**;
  `npm run verify:offline:red` **exit 0**. **Every gate green, shipping an app permanently reporting
  "Offline cache unavailable".**
- **Second mutation, run separately** (bypass condition reverted to the pre-story `import.meta.env.DEV &&
  !payload`, orphaned import dropped): `npm run typecheck` **exit 0**; `npm test` **exit 0, 40 / 409**;
  `npm run build` **exit 0**. The narrowing this story added can be undone with every gate green.

`main.tsx` was restored from git after each mutation and `git status --porcelain` re-run empty each time.
**Filed as DW-110**, owner named. Closing it needs either a harness able to import a module whose top
level calls `void startObservation()`, or the executed browser assertion Epic 8 already owes (D-8.4.25d)
— both outside this story. **The story is correct as shipped; what is unproved is that it stays correct.**

**Recorded plainly, as the operating rules require: this is the third consecutive story in which the
build's own I/O-matrix row 8 could not be tied to an executed test.** The patch moved the untested surface
from an expression to a call site. That is progress, and it is not closure.

---

#### Triage audit (DW-87)

**The reconciliation was verified arithmetically and it holds — unlike Story 8.4e's, which failed by
two.** The layers returned 18 + 10 + 3 + 7 = **38** raw findings; dedup merged **7**, leaving **31**
distinct claims; routes 0 intent_gap + 0 bad_spec + 5 patch + 2 defer + 24 reject = **31**. Cross-checked
against the artefacts rather than against the summary line: `addressed_findings` lists exactly **5**; the
frontmatter `deferred:` block carries exactly **2**; the rejection enumeration accounts for exactly **24**.

**The one caveat, and it is DW-87's own unfinished half.** Rejections **1–20** are enumerated individually
with the ground that refuted each. Rejections **21–24** are recorded as a single line — *"four further
cosmetic or restated observations … carrying no distinct required action"* — with **no claim and no
location each**. Those four **cannot be spot-checked**, and this entry says so rather than implying they
were. DW-87 stays **OPEN**; this story discharged 20 of 24, which is the best showing in the run so far.

**Spot-checked at the cited locations, by measurement where measurement was possible:**

- **#1 CRLF fragility of the derived reader — CONFIRMED refuted, re-measured at close.** A CRLF source
  yields **1** match, identical to LF: JavaScript's `$` under `/m` matches before `\r` as well as `\n`.
- **#2 `rewriteRelease` leaving `index.html` drifted — CONFIRMED refuted by reading the function.** It
  writes `offline-release-manifest.json` and `sw.js` and nothing else; both new proofs snapshot and
  restore exactly those two files.
- **#4 the test re-typing `65`/`9` as a second authority — CONFIRMED refuted.** Traced: were the bound to
  move, the fixture would parse as `accepted` and the reason-table assertion reddens. The drift fails
  **loud**, which is the distinction the story's doctrine actually draws.
- **#7 `GOFLAGS` re-adding the flag to the flag-dropped probe — CONFIRMED refuted by reading the code.**
  It falls through to `fail('… escaped verification …')`. Never a false pass.
- **#13 the unreadable-declaration guard being Vitest-only — CONFIRMED refuted.** The guard itself is on
  the build path; only its failure-mode fixtures live in Vitest.
- **#17 a missing `git` binary — CONFIRMED refuted**, and the review patch is what makes it true: the
  probe-visibility check now catches a git failure in the file's own voice. The missing-Go half is weaker
  (an unframed `ENOENT` rather than a framed `fail`), but it still fails the build non-zero and `go` is
  already a hard prerequisite of the whole gate.

**#3 — the reason-dropping finding, raised independently by THREE layers, rejected on Design Note 5. This
closer's ruling: the rejection is SCOPE-CORRECT, and the observation is STILL TRUE.**

Scope-correct on three grounds, each checked rather than accepted. AC4 is phrased at the
**function-return** layer and is met there. The Block If forbids changing what a user sees on any path but
the over-the-bound one. And the "no diagnostic channel" ground is **measured, not asserted** — re-counted
at close, `src/` and `scripts/` contain **zero** non-test `console.*` calls. There is also a real consumer
for the reason: the dev-server bypass predicate reads it and changes behaviour on it, so the reason is
load-bearing rather than decorative.

And still true: in a production build, an over-bound bootstrap and a malformed one remain **observably
identical** to a user and to an operator. Design Note 5 says so in advance and **invites the disagreement
in writing**, which is precisely why it should not be left in a rejected-findings list where nothing
sweeps it. **Filed as DW-111.** A rejection can be scope-correct and still true; the build has no way to
carry that forward, and the closer does.

---

#### The step-03 commit breach — INSTANCE THREE

The step-03 subagent **committed `7a18079` on its own**, ahead of the point in the workflow that owns
committing. The commit is **content-clean** (nine files, all this story's), carries the required trailer,
was unpushed, and touched no protected path — so the builder appended `68c548e` rather than rewriting
history, per Finalize's keep-commits-already-created rule. That was the right trade.

**But the timing breach is real, and this class now stands at THREE: D-8.4.9c, D-8.4.18, and this.** The
standing note against the first two was that **re-measurement is a recovery, not a repeatable guarantee**.
**The recovery held a third time** — every claim in both commits was re-measured at this close, and every
one reproduced. That is exactly the argument for **pricing this class rather than absorbing it again**:
three consecutive recoveries is evidence both that the recovery works and that the breach is systematic
rather than incidental. It is put to the orchestrator in those terms, not filed as closed.

---

#### Verification — every gate re-run at `68c548e`, `-count=1`, clean tree, branch `main`

Nothing below is carried forward from the dispatch report. Every figure was produced by the invocation
quoted beside it, at `68c548e`, with `git status --porcelain` empty.

- `cd folio-go && go test -count=1 ./...` → **1815 pass / 2 fail / 5 skip**, counted from a `-count=1
  -json` re-run rather than estimated.
- `cd folio-go && go test -count=1 -tags=matrix ./...` → **1826 pass / 3 fail / 5 skip**.
- **Standing reds BY IDENTITY — exactly two, no third.** (1) `TestCorpusMeetsP6ExerciseFloors` and its
  subtest `P6g_(opaque_names)`, `P6g (opaque names) floor not met: got 7, need >=20`. (2) matrix-only
  `TestShippedFacesReproduceFromUpstream`, failing verbatim on `fontgen: fontTools is not importable by
  this interpreter` — a **could-not-execute**, never a byte divergence (DW-86).
- **Fontgen swept both ways, and the passing arm proved non-vacuous.** `cd folio-go &&
  FOLIO_FONTGEN_PYTHON=/Users/panitw/Projects/folio/.fontgen-venv/bin/python go test -count=1
  -tags=matrix -run TestShippedFacesReproduceFromUpstream -v .` → **PASS in 8.38s**, printing **three**
  `matches the recorded derivation` lines and `fontgen: derived and compared 3 of 3 faces`. It asserted
  something. **No committed face has diverged; no golden is at risk on this ground.**
- `cd folio-go && go vet -tags=matrix ./...` → no output, exit 0.
- `gofmt -l folio-go` **run from the REPO ROOT** → no output, exit 0. (Under a `cd folio-go` it prints an
  `lstat` line that reads like success; that form was not used.)
- `cd lint && go test -count=1 ./...` → four `ok` (`cmd/genmanifest` 0.855s, `internal/licence` 0.438s,
  `internal/manifest` 2.285s, `internal/rules` 7.174s), no FAIL, uncached.
- **Four AD-21 legs**, `FOLIO_MATRIX_TARGET=<t> go test -count=1 -tags=matrix -run TestTargetRenderHash
  -v .` → all four PASS, **24 documents hashed per leg, counted from the output rather than assumed**:
  `darwin/arm64` 0.76s, `linux/amd64` 6.55s, `linux/arm64` 4.96s, `js/wasm` 11.29s. **The unset control**
  → PASS in **0.00s**, hashing **0** documents and saying so at `matrix_test.go:2199`. It is a control,
  never a fifth leg.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` → **PASS in
  23.62s**.
- `shasum -a 256 fixtures/*/expected.pdf` **from the repo root** → **23** lines. **Compared against a
  baseline reconstructed OUT OF GIT** — `git show 8dd6c6e:<path>` piped to `shasum` for each of the 23 —
  and the two sets are **byte-identical across all 23**, over the same 23 fixture names at both
  revisions. **No golden moved; no HALT.**
- `md5 -q README.md` → `078d7d80d518d54af2fc04fb270d46b8`, unchanged. **`README.md` appears in neither
  commit**, and it was never used as a scratch or probe file at any point in this close.
- **The three attestation records** — `fixtures/statement-signoff.json`,
  `fixtures/thai-stacked-marks/signoff.json` and `fixtures/embedded-font/signoff.json` — **appear in
  neither commit**, and no `reader`, `date` or `examined` key was added anywhere in the story's diff
  (searched across `8dd6c6e..68c548e`: zero hits). The embedded-font record remains a **transferred**
  reading and carries none.
- **Designer**, node **v24.16.0** (`RELEASE_RUNTIME` pins it exactly): `npm run typecheck` **exit 0**;
  `npm run lint` **exit 0 with exactly 4** `react(only-export-components)` warnings and no other rule, at
  `preview/pdf-viewer.tsx:16,17` and `App.tsx:1324,1331` — **the count of 4 and the rule name are the
  invariant; the line numbers move whenever `App.tsx` does**; `npm test` **exit 0, 40 files / 409 tests
  passed** (baseline 40 / 387); `npm run test:e2e:compile` **exit 0 — this is `tsc -p tsconfig.e2e.json
  --noEmit`, a COMPILE and not a run.**
- **This story's own subject:** `cd folio-designer && npm run build` **exit 0** (it ends in
  `verify:offline`); `npm run verify:offline:red` **exit 0** (183.33s); `npm run verify:offline:wasm`
  **exit 0**.

**Byte figures, each with its invocation, commit and tree state.** All read from
`folio-designer/dist/offline-release-manifest.json` immediately after `cd folio-designer && npm run build`
(exit 0), at **`68c548e`**, **clean tree**, branch `main`, node `v24.16.0`:

- `release.assets.length` = **23**
- `s1.assetCount` = **23**
- `s1.cacheAssets.length` = **23**
- `s1.cachedBytes` = **38,460,833**
- **top-level** `s1VisibleBytes` = **12,425,468**

**`s1VisibleBytes` is a TOP-LEVEL manifest key.** `release.s1.s1VisibleBytes` is `undefined` — **confirmed
by direct read at close**, where it dropped out of the serialised probe entirely and would read to a later
agent as *absence* rather than as *wrong path*. It is **not a metric**; it is quoted here with its command
and is **never reasoned from** (D-8.4.29, DW-100). Every figure above reproduces the dispatch report
exactly — recorded as an independent reproduction, not as a relay.

**23 golden digests, 24 AD-21 documents per leg, and the manifest's 23 assets are THREE DIFFERENT
POPULATIONS.** They are not conflated here, and none was "corrected" to another.

**What was NOT run, said plainly.** The 12 Playwright specs did not execute — CI runs only `tsc --noEmit`
and the browser install remains broken (DW-101, D-8.4.25d). `npm run test:e2e:compile` passed and is a
**compile**. Rejections **21–24** were not spot-checked, because they carry no claim and no location to
check against. No AC in this story is phrased at the browser layer, and none needs to be.

---

#### Register work this story owes — all of it done here

- **DW-106 — CLOSED** by `7a18079` / `68c548e`. The property (two builds of one commit agree) is now
  asserted directly, with the probe's visibility to git asserted first. **Its stated limit is recorded in
  the entry and in the code rather than overclaimed: a two-builds-agree check would NOT have caught
  DW-105**, which is **path dependence** — the same commit built from three different paths gives three
  different binaries, each embedding **87 occurrences of its own absolute root** — because this check
  holds the checkout path fixed by construction. **Its value is the NEXT input, not the last one.**
- **DW-107 — CLOSED** by `7a18079`. Story 8.4g's flag-dropped stamp proof is now executed by `npm run
  verify:offline:red` instead of existing as prose in a delivery log.
- **DW-105 — stays OPEN.** `-trimpath` was deliberately not applied by this story.
- **DW-100 — its reproducibility half is discharged; DW-100 itself is NOT CLOSED.** The missing IBM Plex
  row half belongs to **Story 8.4d**.
- **DW-110 (new)** — `main.tsx`'s two startup decisions are still executed by no test at their call sites;
  measured above.
- **DW-111 (new)** — the named rejection reason has no production consumer: the three-layer finding,
  scope-correctly rejected and still true.
- **DW-112 (new)** — the derived bound is resolved from the working tree at verify time, so a `dist/`
  detached from its source tree throws an unframed ENOENT (this spec's `deferred:` entry 1, placed).
- **DW-113 (new)** — the tree-state probe survives a signal landing between the two engine builds (this
  spec's `deferred:` entry 2, placed). **Its obvious remedy is specifically WRONG**: a `.gitignore` entry
  would make the visibility check vacuous.

---

#### Because the run pauses here — what a resumer must not re-derive

**Remaining in the sprint: 8.5 → 8.6 → 8.4d.** `epic-8` stays **`backlog`**; this closer does not mark
epics.

**What Epic 8 still owes at its gate: the executed browser assertion (D-8.4.25d).** It is **owed, not
attempted-and-passed.** The 12 Playwright specs exist, the config is real, and the harness runs — but the
Chrome for Testing install has never completed (the cache entry stays at 428 KB), so **nothing the browser
would have told us is established.** Do not let a passing `test:e2e:compile` read as a run.

**Open registers a resumer must NOT re-derive — all still OPEN at this commit:** **DW-100** (the
`s1VisibleBytes` figure is not yet reproducible; Story 8.4d), **DW-101** (twelve executable Playwright
specs that CI never runs), **DW-103** (a real policy check that CI never runs — DW-101's twin),
**DW-105** (the engine wasm is still a function of the checkout path), **DW-108** (scope constraints
dropped from the regenerated epic context), **DW-109** (the ambient environment is an unrecorded build
input). Plus **DW-87** (rejections still not fully enumerable) and the four new entries above.

**DW-108 IS DUE BEFORE STORY 8.5 IS DISPATCHED.** Its discharge is to confirm that 8.5's spec carries the
catalogue procurement rules and the out-of-scope list **sourced from
`_bmad-output/specs/spec-fonts/SPEC.md`'s `## Non-goals`** — **not** from the regenerated
`epic-8-context.md`, which is a declared-regenerable cache and is where the constraints were lost. The
next story the loss can damage is 8.5, and 8.5 is next.

**Working tree left clean. Nothing pushed. No branch created.**
