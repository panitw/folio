---
title: 'Story 16.2: A fetched face stays on this machine'
type: 'feature'
created: '2026-09-02'
status: 'done'
review_loop_iteration: 0
followup_review_recommended: true
baseline_commit: 'a40c34db6cff7372363b2a553710eff48759bef1'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-8-15-decision-log.md'
warnings: []
deferred:
  - 'DW-171 — the designer CI job stops at npm run test, so six later steps have never run'
  - 'DW-172 — the local-tier pick fetches inside the pick hold with no timeout'
  - 'DW-173 — the store never closes its connection and answers no onversionchange'
  - 'DW-174 — get() re-hashes every face, and the preview registration re-reads all stored faces per store mutation'
  - 'DW-175 — bytes can outlive the face record that names them'
  - 'DW-176 — real browser IndexedDB has no witness; ruled to Story 16.3 as its fourth case'
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

A typeface you download once is now kept. The next statement you build on this machine offers it
immediately, with no download and no wait — and offers it even with the network down, because the bytes
are already here. The typography panel gained a short list of what this machine holds, showing each
typeface's size and the day it arrived, with a button to let one go; letting one go frees the space and
nothing else, because a document that already uses that typeface carries its own copy inside it.

Two things this deliberately is not. It is not a list of the fonts installed on your computer: the
designer never looks at those, and the group in the font menu called *available locally* means
"typefaces this designer has downloaded before", not "typefaces Windows has". And it is not a second
copy of your document's fonts — a template still carries its own typefaces inside it, so a file you
send to someone else is unaffected by anything stored here.

One more thing changed while this was being built, and it fixes a way the font control could quietly
die. A download that FAILS has always told you so. A download that simply hangs — a hotel wifi login
page, a stuck proxy — used to leave the font control greyed out for the rest of your session, with no
message and no way back. It now gives up after thirty seconds, says what happened in its own words
(not "you have no network", which would be a lie while your network is fine), and hands the control
back. It does not retry on its own, because retrying a hang just hides it.

<intent-contract>

## Intent

**Problem:** D-16.1 makes every new family a network fetch. Without a store, the same family is
re-fetched for every document, the designer is useless offline for anything it has already downloaded,
and the design's third dropdown group (`AVAILABLE LOCALLY`) has nothing behind it.

**Approach:** An origin-scoped **IndexedDB** store, keyed by the **SHA-256 of the face bytes** — the
same content address `.folio`'s `assets` map uses — holding bytes, licence text, copyright and the
family metadata needed to offer the face again. Reads populate the dropdown group; a hit means no
fetch. The store is a **cache and a source**, never an authority: a document still carries its own
faces.

## Boundaries & Constraints

**Always:**
- **IndexedDB, and `localStorage` is refused on arithmetic** (D-16.2). The owner's words were "local
  storage"; the mechanism is the engineer's. `localStorage` is a ~5 MB per-origin **string** quota;
  storing bytes there means base64 at **+33%**; a measured `Sarabun-Regular.ttf` is **90,220 bytes →
  ~120 KB stored**; and its `QuotaExceededError` is synchronous with no partial-write path. **Write
  that reasoning into the module**, so the next person does not "simplify" it back.
- **Keyed by content hash, never by family name.** A family that changes upstream is a different key,
  not a silent substitution — the property AD-8 already buys the document.
- **The store never becomes an authority on a document.** A `.folio` carries its faces (CAP-2). The
  store shortens a fetch; it never stands in for what a file contains, and removing an entry never
  changes a saved document.
- **`AVAILABLE LOCALLY` is fetched faces, never host fonts** (D-16.2). The one Non-goal clause D-16.1
  leaves standing is *"No host fonts. Faces installed on the authoring or rendering machine are never
  enumerated or read."* The Local Font Access API is not used, referenced, or feature-detected.
- **Everything the embed command requires is stored with the bytes.** Licence identifier, licence text
  and copyright travel into the store, because a face offered from the store must be embeddable
  without a network — and `embedFontFamily` refuses without all three.
- **A STALLED fetch is a first-class degradation, not an unhandled case** (D-16.R.14, pulled in from
  DW-165). 16.2's subject is what happens when a fetch does not give you bytes, and its matrix already
  carries network-down and store-failure; **a stall is the one member of that class it did not cover,
  and it is the worst** — a rejection degrades with a message, a stall leaves the font control **dead
  for the session with no message at all**. Story 16.1's pick hold now spans the whole fetch chain with
  no timeout, so this is reachable today.
  - **A timeout is a number, and a number needs a basis** — derived from a measured fetch time against
    the real host, or an explicit "chosen because", never a magic constant. The measurement carries
    command, commit, tree state and working directory (D-8.4j.8).
  - **The hold must be PROVEN to release**, by a test that reds when the release is removed. This is the
    **third** instance of the state-lifetime class in two stories (D-16.R.15); assume a fourth.
  - **It must not silently retry.** A retry over a deterministic stall hides it — Story 16.0's `Never:`
    clause, same reasoning.
- **Storage failure is a degradation, stated.** A private window, cleared site data or a quota refusal
  leaves a working designer with an empty group, not an error the author cannot act on.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **The store would be read at render time.** FR33 is untouched: nothing is fetched or read from the
  machine at render. This store serves **authoring** only.
- **A stored face would reach a document without its licence record.** That would put a document its
  own parser refuses one step away.
- **`maximumCacheAssets` or the offline release contract would move.** This store is runtime data, not
  a release asset.

**Never:** enumerate host fonts · store in `localStorage` · key by family name · let a store miss
break a document that already embeds the face.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| First fetch of a family | Store empty | Bytes + licence text + copyright + metadata written under the SHA-256 | Write failure degrades to no-store |
| Re-pick, later document | Key present | **No fetch**; embed from stored bytes | No error |
| Re-pick, network down | Key present | Works; this is the store's whole point | No error |
| Family not stored, network down | Key absent | Stated: cannot add right now | Degradation |
| **Fetch STALLS rather than rejecting** | request never settles | **The timeout releases the hold and states the degradation** — the control must not be left dead for the session (D-16.R.14) | Degradation, stated; **never a silent retry** |
| Private window / storage blocked | IndexedDB throws on open | Group empty; picks still fetch; message states it | Degradation, once, not per pick |
| Quota exceeded on write | Large face, full origin | Fetch and embed still succeed; the caching is what failed | Degradation, named |
| Author removes an entry | Entry present | Removed, with what is being removed named; documents unaffected | No error |
| Stored bytes fail to decode later | Corrupt entry | Entry treated as absent and dropped; refetch on next pick | Self-healing, logged honestly |

</intent-contract>

## Code Map

**Designer (`folio-designer/`)**
- `src/App.tsx:216-243` — the document-scoped carried-face registration, and its comment on why the
  lifetime is document-scoped rather than per-component. **The precedent for lifetime reasoning**, and
  the machine store is a third lifetime that must be argued the same way.
- `src/embedded-face-registry.ts`, `src/embedded-face-family.ts` — how a face's asset key becomes a CSS
  family, and `isCarriedFaceAssetKey`, the key-shape predicate. **The store's keys must satisfy it.**
- `src/App.tsx:660` — `pickCatalogueFamily`, which Story 16.1 widens; the store's read goes in
  front of its fetch.
- `src/offline-lifecycle.ts` — the service worker's lifetime, and the origin whose storage this is.
- `src/image-file.ts`, `src/component-asset-command.ts` — the other place bytes cross from the browser
  into a command; same shape, no persistence.
- `src/release-payload.ts:33` — `maximumCacheAssets = 64`. Untouched; assert it.

**Go (`folio-go/`)**
- `folio-go/component_commands.go:2360-2410` — `embedFontFamily` (the file is at the module root, NOT under `internal/template/`). **The store feeds it exactly what a fetch
  feeds it**; Go cannot tell the difference and must not be able to.

**Anchors re-verified at the plan gate against `HEAD`, not against `baseline_commit`.** This spec was
planned at `a40c34d`; Stories 16.0, 16.1, 16.1b and 16.1a have landed since and moved every line-anchored
reference above except `release-payload.ts:33`. The originals pointed at unrelated code
(`186-224` at a ref block, `608-627` at `applyFontChain`) and at a path that does not exist
(`internal/template/component_commands.go`). **Re-verify these anchors again before relying on them** —
`baseline_commit` is stale by construction, and a Code Map is the one part of a spec that rots silently
because a wrong line number still reads as a right one.

## Tasks & Acceptance

**Execution:**
- `src/` — a font-store module over IndexedDB: `get(key)`, `put(record)`, `list()`, `remove(key)`, and
  one open path that degrades rather than throwing into callers. Records carry bytes, mediaType,
  family, style, licence, licenceText, copyright, source, scripts and the date fetched.
- **HOW THE STORE IS TESTED — RULED AT BUILD DISPATCH (2026-09-03). Use `fake-indexeddb` as a
  devDependency.** Measured at the gate: **jsdom 28.1.0 provides no IndexedDB at all**
  (`'indexedDB' in window` is `false`; `globalThis.indexedDB` is `undefined`), so without this the
  store's tests have no backing store to run against. `crypto.subtle` IS available, so the SHA-256
  addressing needs nothing.
  - **Why not a hand-rolled fake.** A fake written here is shaped by its author's reading of the
    IndexedDB request/transaction plumbing, so it agrees with the store by construction and proves the
    plumbing not at all — both sides move together, which is the vacuity this run exists to refuse. An
    independent implementation **can disagree with the code**, and that capacity is the whole value of
    the test.
  - **The dependency-count constraint does not exist in the form it is usually assumed.** Verified
    independently at the gate: D-1.3.6 invariant (b) — "do not add a dependency" — is scoped to
    **`folio-go`**, the Go engine, and `folio-designer/` is not in it. Story 5.5's project-structure
    note ("No new dependency is justified for pickers, JSON schema, storage, sync, or downloads …
    audit any unavoidable package through AD-26 before adding it") is about a **runtime** package doing
    a job the browser API already does, and its disposition clause is a **procedure, not a
    prohibition**. `fake-indexeddb` does not do IndexedDB's job in the product; it supplies IndexedDB
    where the environment has none, in tests, and never ships. AD-26 is a **licence boundary with CI
    automation**, not a count cap, and it already records `pdfjs-dist` as Apache-2.0 — so Apache-2.0 in
    the designer is **precedented, not novel**.
  - **FOUR CONDITIONS, and the fourth is a halt.**
    1. **`devDependencies` only, never `dependencies`** — plus a test or scan asserting **no shipping
       module under `src/` imports it**. A test-only package drifting into the bundle is the class
       `scripts/forbidden-font-hosts.mjs` exists to catch in another medium.
    2. **The AD-26 audit is over the TRANSITIVE tree, because AD-26 says "at any depth."** Report
       **every** package added, with its licence — not just the top-level one.
       *Audited at the gate before dispatch (`npm install --save-dev --dry-run fake-indexeddb`): the
       tree adds **exactly one** package, `fake-indexeddb@6.2.5`, **Apache-2.0**, with no transitive
       dependencies of its own. Re-confirm this after the real install, from the installed package's
       own `LICENSE` file and `npm ls fake-indexeddb`, and report it.*
    3. **Regenerate `lint/MANIFEST.md`** via `cd lint && go run ./cmd/genmanifest`, then
       `cd lint && go test -count=1 ./...` green. `TestManifestUpToDate` walks the tree, which is
       DW-168's exact class, so **`-count=1` is not optional here**.
    4. **HALT** if any package in the transitive tree carries a licence outside AD-26's admitted set.
       Do not substitute another package, do not vendor a subset, and do not fall back to a hand-rolled
       fake silently — return it to the engineering lead, whose decision it is.
  - **`pinnedCensus` and its derived floor do not move.** Verified: `committedLicenceFiles` shells out
    to `git ls-files` (`licencecensus_test.go:381-385`), so it enumerates **tracked** files only, and
    `node_modules` is untracked. Only `MANIFEST.md` moves.
  - **The honest residual, and where it goes.** Real browser IndexedDB is proven only by a browser
    witness under any option. Register it — **routed to Story 16.3's browser run**, NOT to the epic
    catch-up — as a **fourth case** beside DW-161's three: *a stored face survives a reload and is
    offered with the network disabled*. That run already reloads to exercise the offline path, so the
    case costs nothing and the residual retires **one story after it is created**.
- `src/` — hash the bytes to the key in the browser **for the store's own addressing**, and assert in a
  test that the key the store computes equals the asset key Go derives, so the two addressings cannot
  drift.
- `src/App.tsx` — read the store before fetching; register stored faces for preview alongside the
  document's carried faces, without disturbing the document-scoped effect's keying.
- `src/font-index.ts` — the store's listing joins the existing `FamilySource` union as a third arm,
  `{ tier: 'stored'; family: string; record: StoredFace }`, alongside `'local'` and `'web'`. The pick
  path handles all three. Story 16.4 groups this union; it does not reshape it. **Build the seam here,
  do not leave it to 16.4**: an exhaustive switch over `FamilySource` reds until the new arm is
  handled, so the tier cannot be silently dropped later (ruled, D-16.R.33 R1).
- `src/` — the removal affordance, naming what it removes and stating that documents are unaffected.
- `src/font-source.ts` — the fetch timeout goes in the **default fetcher** (`fetcher: Fetcher = (url)
  => fetch(url)`), not at the six call sites, and the signal is passed **into `fetch()`** so the abort
  reaches the **body** stream: the bytes are read by `response.arrayBuffer()` AFTER the fetcher
  returns, so a header-only timeout leaves the worst real stall uncovered. A test must stall during
  `arrayBuffer()`, not during the request, and assert the same stated degradation.
- The timeout's NUMBER is established by measurement, never invented: timed fetches against the real
  host `raw.githubusercontent.com` over both shapes in the chain (a small `METADATA.pb` and a **large**
  face). Take the **maximum** observed, not the
  mean or median — the number is a ceiling on patience, not an estimate of typical latency — times a
  stated factor with its reason recorded. Record it per D-8.4j.8 (command, commit, tree state, working
  directory). The constant carries its own measurement in its comment.
  **It may NOT be measured in jsdom against the stubbed fetcher** — timing our own stub measures the
  stub, the vacuity D-16.R.12 named. A number that "feels right" or is copied from another project is
  refused (D-16.R.14).
  - **RULED AT BUILD DISPATCH (2026-09-03), and this supersedes the sizing premise this bullet
    originally carried.** The superseded text said "SPEC-fonts records shipped faces up to 646 KB".
    That figure is true of the **31 committed local-tier faces** and false of the **fetchable**
    population the budget actually serves — a denominator error of the D-16.R.27 class, caught by
    measurement at the gate. **Size against the fetchable population's true maximum.**
    Measured denominator: of 1,811 index rows, **1,273** are addable after removing variable-only rows
    and the 31 local-tier families; **1,218** of those have a `<slug>-Regular.ttf` published in
    `google/fonts` at `main`. Their sizes: median 107,440 B, p90 420,092 B, p99 1,715,888 B,
    **max 24,271,604 B** (`Noto Color Emoji`, verified present in the snapshot as
    `variable: false`, so it is in `webFamilies` and offerable today — a legitimate pick, not a
    hypothetical).
  - **T = 30_000 ms, via `AbortSignal.timeout(30_000)` passed into `fetch()` in the default fetcher.**
    `AbortSignal.timeout` is chosen over a hand-armed `setTimeout` + `clearTimeout` deliberately: there
    is no disarm path, so the signal cannot be cleared when the headers arrive and therefore still
    covers `response.arrayBuffer()`.
  - **The constant's comment carries the ARITHMETIC, not prose.** It must contain: the measured shapes
    with their byte counts and observed maxima; the x10 factor with its reason; the line
    `2,097 x 10 = 20,970 <= 30,000`; and the sample's own limit — **one connection, one day, five
    repetitions** — because that is the honest statement of why the factor is x10 and not x2.
    Why 30_000 and not 20_000: `2,097 ms x 10 = 20,970 ms`, so a 20 s budget puts the single largest
    offerable face **outside** the budget the factor was chosen to cover, and the constant would ship
    carrying arithmetic that contradicts its own stated reason (D-16.R.25). The cost asymmetry settles
    the direction: too short **wrongly refuses a legitimate pick of a real font**, loudly and
    repeatably; too long only lengthens an already-bounded hold on a genuine stall.
- **The chain terminates on the FIRST abort, and that is asserted rather than assumed. NO chain
  deadline is added, and this is NOT registered as deferred work** (ruled at build dispatch,
  2026-09-03, superseding the "6 x T" worst case this bullet originally asserted).
  Measured and independently confirmed three ways: the probe catch at `font-source.ts:281-283`, the
  byte catch at `:339-341` and `readText`'s catch at `:381` each `return` a refusal, and the probe
  loop's only `continue` is on a **404**, which an abort never produces. So a stall ends the whole
  chain at the first timeout. **The STALL bound is therefore T plus the requests that already
  completed — roughly T + 3 s.** State that number, and its derivation, in the Delivery Log.
  - **AND THE QUALIFIER IS PART OF THE CLAIM, ADDED AT REVIEW (2026-09-03).** `T + completed` is the
    bound on a **STALL**, and only on a stall, because the chain terminates on the first abort. It is
    **not** the general worst case. `AbortSignal.timeout` is armed **per request**, so a host that is
    slow but ALIVE — every request answering just under T — is aborted by nothing, and the chain can
    hold for up to **n x T**, where n is the number of requests the pick makes. That is not an
    oversight: **no chain deadline was added, deliberately and by ruling**, so `n x T` is the accepted
    slow-but-alive bound and the ruling stands with the number stated rather than implied. The bullet
    originally said "not 6 x T" flat; that reading is true of the stall and false in general, and an
    unqualified worst-case claim is exactly the kind of statement this spec refuses elsewhere.
  - A deferral with a trigger is the wrong instrument for a property that can simply be asserted: a
    note ages, an assertion reds. **Add one table-driven test that the chain terminates on the first
    abort** — an injected fetcher that aborts on request *N* produces a stated refusal, for *N* across
    the probe loop, the licence read and the byte read, **with no further fetcher calls after it**
    (call count written as a literal, never computed from the code path that drives the fetches).
    **Red-proof: change one catch to `continue`; the test must red.**
- The stall path states its degradation in the located, anchored form the offline refusal uses at
  `font-source.ts:340`, and **must not reuse the offline wording** — "you cannot add a family without a
  network connection" is FALSE when the network is up and the host is hanging.
- Tests: store round trip; hit avoids fetch; miss with no network degrades as stated; open failure
  leaves a working designer; corrupt entry is dropped and refetched; and a **red-provable** assertion
  that no code path enumerates host fonts. That assertion is **source-level over the whole `src/`
  tree** — a test that only checks the store module would pass while `App.tsx` called
  `queryLocalFonts` — and it is red-proved by **deleting the guard**, never by falsifying a condition.
- The timeout's red-proof is THREE tests, and the second is the one usually skipped (D-16.R.15 named
  this class and said to assume a fourth instance):
  **(a) Release.** Never-settling fetcher + fake timers: after T the control is enabled and the
  degradation stated. Mutation: **delete the release path**; the test must red.
  **(b) The second carrier.** D-16.R.15's actual defect was a hold backed by both a ref and state,
  where the document-reset path cleared only the state copy. So: stall a fetch, **replace the document
  mid-stall**, then assert the hold is clear and a subsequent pick works. Proving (a) alone re-proves a
  shape that was already fixed.
  **(c) No silent retry.** Assert the fetcher was called **exactly** the expected number of times and
  not once more after the timeout, with the count written as a **literal** — never computed from the
  same code path that drives the fetches.
- **This discharges DW-165.** Close that entry with the measured number recorded, not merely marked
  done.
- A face stored and read back carries a `source` byte-identical to the one it was stored with,
  asserted through `assertProvenanceShape` at the RETRIEVAL side — the store is a carrier, and a
  carrier that normalises, truncates or re-derives a provenance record has become an authority
  (contract Boundary: *"the store never becomes an authority on a document"*). This makes an existing
  Boundary checkable rather than adding scope, and it puts a **third real call site** on the shared
  provenance helper, whose current two call sites both feed it known-good values — the precise
  condition that let its assertions go inert (D-16.R.34 F1).
- Docs: `font-catalogue.md` gains the store's description — what it is, what it is not, and that it is
  authoring-only.

**Acceptance Criteria:**
- Given a fetched face, when the fetch succeeds, then it is stored under the SHA-256 of its bytes with
  everything the embed command requires.
- Given a stored family, when it is picked again in any document on this machine, then nothing is
  fetched and the embed runs over the stored bytes.
- Given a stored family and no network, when it is picked, then it is embedded successfully.
- Given a populated store, when its listing is read, then it returns exactly the faces this designer
  has fetched and stored, and nothing else — and a red-provable test asserts that no code path in the
  designer enumerates, reads, or feature-detects host fonts (no `queryLocalFonts`, no
  `navigator.fonts`, no Local Font Access API by any spelling).
- Given storage that cannot be opened or written, when the designer runs, then it still works and says
  what is degraded.
- Given a fetch that stalls rather than rejecting, when the timeout fires, then the hold is released,
  the degradation is stated, nothing is silently retried — and a test reds if the release is removed.
- Given a stored face the author removes, when it is removed, then documents that already embed it are
  unchanged.

## Design Notes

**Why the content hash and not the family name.** The store answers "do I already have these bytes",
which is the same question `assetKeyReferenced` and the `assets` map answer. Keying by name would make
the store answer a *different* question — "do I have something called Sarabun" — and the day upstream
changes the face, the store would hand over the old bytes under the new name and the document would
carry a face nobody chose.

**Why this is a separate story from 16.1.** 16.1 can ship without persistence and be correct; it would
simply re-fetch. Fusing them would put a storage failure in the same blast radius as a licence
refusal, and those are different failures with different right answers.

**What "on this machine" honestly means.** Origin-scoped browser storage: this browser, this profile,
this origin. Not the OS, not synced, not shared with another browser on the same machine. The UI
should not imply more than that.

## Verification

- `cd folio-designer && npm run test` — **run as its own command**, and report the file/test counts and
  the identity of every failure.
- `cd folio-designer && npm run build` — **run as its own command, and report it separately.**
  **RULED AT BUILD DISPATCH (2026-09-03): the previous single line `npm run test && npm run build` was
  a gate that never fired.** `npm test` exits **1** at baseline on the DW-152 red
  (`src/canvas-authority-contract.test.ts:190`), so the `&&` short-circuits and the build has been
  silently not-measured for as long as that red has existed. Two lines, two reported results.
- `cd folio-designer && npm run typecheck && npm run lint && npm run scan:font-hosts && npm run
  scan:host-fonts` — the lint warning count and rule are the invariant (expect exactly 4
  `only-export-components`); the line numbers never are.
  **`scan:host-fonts` was added to this list at review (2026-09-03), and to the `build` script beside
  `scan:font-hosts`.** The host-font guard shipped defined in `package.json` and wired into nothing:
  `build` ran only its model, `scan:font-hosts`. A guard nothing runs is a guard that is not running,
  and this one was in fact RED at the time it was reported green — see the Delivery Log's corrected
  table. It now fails the build, which is the same standing its model has.
- `cd lint && go run ./cmd/genmanifest` (required — a devDependency is being added), then
  `cd lint && go test -count=1 ./...`
- `cd folio-go && go test -count=1 ./...`
  **`-count=1` is mandatory and is not decoration** (DW-168, narrowed by D-16.R.31): CI already passes
  it everywhere, so the live residue is exactly this by-hand path. This story touches no Go, which is
  precisely the condition under which a filesystem-walking Go test replays a stale green.
- Browser: fetch a family; reload; confirm it is offered with the network disabled; remove it; confirm
  a document that embedded it still renders.
  **NOT RUN IN THIS STORY — deferred by the run's heavy-test cadence (`end-of-run`, D-16.R.1), and
  routed to Story 16.3's browser run** rather than to the epic catch-up, as the fourth case beside
  DW-161's three (ruled at build dispatch, 2026-09-03). The browser e2e specs, the matrix corpora, the
  AD-21 legs and `TestCrossTargetByteIdentity` are likewise **not** run here. Name them as unrun; never
  report them as passing.
- The 23 golden digests, unmoved. `maximumCacheAssets` still 64 — **and report the measured
  `s1.assetCount` after the build** (DW-162's margin is halved to 10 of 64 and is watched by nothing).
  The store is runtime IndexedDB data and must consume **no** cache slot; reporting the number means a
  store implementation that accidentally lands anything in the release manifest is caught by an
  assertion rather than by a release failing. The mechanism stays deferred; this is the assertion half.

## Spec Change Log

### 2026-09-03 — build dispatch, two forks ruled by the engineering lead; `<intent-contract>` byte-untouched

Story 16.2 entered its build dispatch at `status: ready-for-dev` with CHECKPOINT 1 already passed and
the `<intent-contract>` (lines 30-105) frozen. Two forks were raised as Open Questions **before**
implementation, because both change the shape of the code and the step-03 handoff is verbatim — the
spec is the implementer's sole source of truth, so a ruling that is not written here does not reach it.
**The frozen block was not edited, and the check IS a digest** (restated at review, 2026-09-03; the
original wording of this sentence was "every edit in this entry is at line >= 106 … and the check is
the diff-hunk one, not a digest", and that line claim was not true of the commit this entry shipped
in). The `<intent-contract>` block — its two delimiter lines included, which is lines 39-114 of this
file — is **byte-identical** before and after, sha256
`193937cf232c65708ceca9b5a9a99779f170b28128f90437c918da300e63ce4a`, verified over `a378acd` (the last
commit before this entry existed) and `2a0c92a` (the commit that added it), and the same digest holds
back at `0e3b576`. That is the claim worth making, and it is stronger than the line-number one.

The line claim itself was wrong: the commit that carried this entry also changed the frontmatter
`status` at **line 5** (`ready-for-dev` → `in-review`) and rewrote the plain-terms opener at roughly
**lines 18-41** — both below 106, and both correctly outside the frozen block. Nothing about the
rulings below changes; only the sentence describing what was verified.

**Rulings already carried into this spec at dispatch and NOT re-opened:** D-16.R.33 **R1** (the store's
listing joins `FamilySource` as a third `'stored'` arm and the seam is built HERE, not left to 16.4;
16.2 builds no new dropdown group) and D-16.R.33 **R4** (every timeout requirement, including that the
signal reach `fetch()` so the abort propagates to the **body** stream, and that a test stall **during**
`response.arrayBuffer()`).

**Q1 — the timeout's number, and whether 6 x T needs a chain deadline.**
Three changes, all in `## Tasks & Acceptance`:

1. **The sizing premise was measured false and is superseded.** The spec sized the budget against
   "shipped faces up to 646 KB". That is true of the **31 committed local-tier faces** and false of the
   **1,273 fetchable families** the budget serves, whose published Regular statics reach
   **24,271,604 B** — a denominator error of the D-16.R.27 class. Measuring against the fetchable
   maximum is strictly more conservative and is what the spec's own reasoning asks for. **Accepted by
   the lead, who independently confirmed `Noto Color Emoji` is `variable: false` in the snapshot and so
   is offerable today.** Noted but deliberately not acted on: that face is one of DW-163's 66
   script-less families, and if 16.3 filters those the maximum drops to `Noto Sans SignWriting`
   (7,881,488 B) — **a later filter can only add margin, never remove it**, so no revisit clause.
2. **T = 30_000, not the 20_000 the builder proposed — overruled on the builder's own arithmetic.**
   `2,097 ms x 10 = 20,970 ms > 20,000 ms`, so a 20 s budget excludes the single largest offerable face
   from the very budget the x10 factor was chosen to cover, and would ship a constant whose stated
   reason its own arithmetic contradicts (D-16.R.25). Cost asymmetry settles the direction: too short
   wrongly refuses a legitimate pick, loudly and repeatably; too long only lengthens an
   already-bounded hold on a genuine stall.
3. **No chain deadline, and no deferral either — the property is asserted instead.** The builder's
   measurement that a stall ends the chain at the **first** abort (every catch `return`s; the loop's
   only `continue` is on a 404, which an abort never produces) was verified independently by the
   orchestrator and by the lead. So the worst-case hold is **T + the requests that already completed,
   roughly T + 3 s — not 6 x T**, which the superseded text asserted. A deferral with a trigger is the
   wrong instrument for a property that can be asserted: a note ages, an assertion reds. A
   table-driven first-abort-terminates test is added, red-proved by changing one catch to `continue`.

**Q2 — how the IndexedDB store is tested: `fake-indexeddb` as a devDependency.**
Measured at the gate: jsdom 28.1.0 provides **no IndexedDB**, so the store's tests had no backing store
under any reading of the spec, which is silent on this. The builder recommended the real package over a
hand-rolled fake **against its own convenience**, on the ground that a fake it wrote would agree with
its code by construction — and that argument is what the lead ruled on. The dependency constraint the
builder deferred to was checked by three readers and **does not exist in the assumed form**: D-1.3.6
invariant (b) is scoped to `folio-go`; Story 5.5's note governs a **runtime** package doing a job the
browser already does and states a **procedure**, not a prohibition; AD-26 is a licence boundary, not a
count cap, and already admits Apache-2.0 via `pdfjs-dist`. Four conditions attach, the fourth a halt;
all four are written into `## Tasks & Acceptance`.

**A gate that was never firing, found while reading the spec's own commands.** `## Verification` said
`npm run test && npm run build`. `npm test` exits **1** at baseline on the DW-152 red, so the `&&`
short-circuited and **the build has not been measured by this line for as long as that red has
existed**. Split into two commands, each reported separately.

**Baseline measured at `3c45993`, clean tree, before any edit:** `folio-go`
`go test -count=1 ./...` → **1910 pass / 2 fail / 5 skip**, the two being
`TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`, the mandated permanent red;
`folio-designer` `npm test` → **48 files / 529 tests, 1 failing**, that failure being DW-152 at
`src/canvas-authority-contract.test.ts:190` with the received array exactly
`["e2e/e9-5-border-no-ink.spec.ts: /\bgetComputedStyle\s*\(/"]`. Note `baseline_commit` in the
frontmatter remains the planning-time `a40c34d` per the workflow's preserve rule; **`3c45993` is where
these pre-implementation baseline suites were measured, and is not the commit the story was built on —
that is `227befe`, as the Delivery Log records** (corrected at review, 2026-09-03: this sentence used
to say "the commit this story was actually built on is `3c45993`", which contradicted the Delivery Log
two sections below. Three commits are in play and each has one job: `a40c34d` planned it, `3c45993`
measured its baseline, `227befe` built it).

## Delivery Log

### 2026-09-03 — built

Built on `main` at baseline `227befe` (the frontmatter's `baseline_commit` remains the planning-time
`a40c34d` per the workflow's preserve rule). Nothing pushed, no branch created.

**Anchors re-verified against `HEAD` before anything was written**, per the Code Map's own warning.
`src/App.tsx:216-243` (the document-scoped carried-face effect) and `src/App.tsx:660`
(`pickCatalogueFamily`) had both moved again since the spec was written; `folio-go/component_commands.go`
is at the module root as the spec's correction says, with `embedFontFamily` at `:2360`;
`src/release-payload.ts:33` is still `const maximumCacheAssets = 64`.

**What shipped.**

- **`src/font-store.ts`** — the machine store over IndexedDB: `get`, `put`, `list`, `remove`, one open
  path that degrades rather than throwing into callers, and `storedFaceKey`. Records carry bytes,
  mediaType, family, style, licence, licenceText, copyright, source, scripts and the date fetched. Two
  object stores under one key, written and deleted in one transaction, so a listing renders a menu
  without deserializing every face. **The `localStorage` refusal and its arithmetic are written into the
  module**, as the contract demands.
- **`src/font-index.ts`** — `FamilySource` gains its third `'stored'` arm and `offeredFamilies` takes
  the listing. A stored family **replaces** its web row rather than sitting beside it; the local tier is
  never displaced; a stored family the snapshot has stopped listing is still offered.
  `familySourceNote` is the exhaustive switch that makes the 16.4 hand-off a mechanism.
- **`src/App.tsx`** — the store is opened once for the session, its read goes in front of the fetch, a
  fetched face is kept **after** the embed (so a quota refusal cannot decide whether an author gets
  their font), stored faces are registered for preview in a **separate machine-scoped** registration
  that does not disturb the document-scoped effect's keying, and the removal affordance names what it
  removes and states that documents are unaffected.
- **`src/font-source.ts`** — the fetch timeout, in the default fetcher, with the signal passed into
  `fetch()`; `readText` returns a union so a stalled licence read is not reported as a missing licence
  file; the stall's own degradation wording.
- **`scripts/host-font-access.mjs` + `src/host-font-access.test.ts`** — the source scan over the whole
  designer for the Local Font Access API, shaped on `forbidden-font-hosts.mjs`.
- **Docs** — `font-catalogue.md` gains *"The machine store"*, and its panel section gains the third
  tier.

**Verification, each command run on its own and reported separately** (the previous single
`npm run test && npm run build` line was a gate that never fired):

⚠ **THIS TABLE WAS WRONG WHEN IT WAS WRITTEN, AND WHY IS WORTH MORE THAN THE NUMBERS.** It is
reproduced below in its corrected form, re-measured command by command at review (2026-09-03) after the
review patches landed. Two of its rows had been false at the time they were recorded:

- It reported `npm run test` as **1 failing**. The true figure at the commit it describes was **2**, and
  the second failure was **this story's own guard**, `src/host-font-access.test.ts:68`, red against
  itself: the positive control at `:89` wrote `queryLocalFonts` whole in an expected value, and the
  whole-tree scan — which reads RAW source and which does read its own test file — found it. Lines 82
  and 85 of the same file split that spelling deliberately; line 89 was missed. The story's central
  red-provable assertion was therefore shipped RED, reported GREEN.
- It reported `scan:host-fonts` as 0 occurrences in **123** files. The real population is **129**, and
  at the time the command **exited 1**.

**THE MEASUREMENT-ORDER TRAP, RECORDED FOR THE NEXT STORY.** Both wrong numbers have one cause: the
figures were taken while this story's own six new files were still **untracked**. Both scans enumerate
`git ls-files`, so `font-store.ts`, `App.font-store.test.tsx` and the scanner's own test were invisible
to a scan run before `git add` — the population was six files short, and the file carrying the
violation had not been read. A scan whose population comes from the index measures the INDEX, not the
working tree. **Run the tracked-file gates after staging, or the green they report is a green over a
smaller repository than the one being shipped.** The same trap is what made `123` and `1 failing` agree
with each other and with nothing else.

Re-measured, each command run on its own (the previous single `npm run test && npm run build` line was
a gate that never fired):

| Command | Result |
|---|---|
| `cd folio-designer && npm run test` | **51 files / 584 tests, 1 failing** — the pre-existing DW-152 red at `src/canvas-authority-contract.test.ts:190`, received array exactly `["e2e/e9-5-border-no-ink.spec.ts: /\bgetComputedStyle\s*\(/"]`, which is **not this story's** and is not fixed here. It is now the ONLY failure. As shipped it was **579 tests, 2 failing** — DW-152 plus this story's own `src/host-font-access.test.ts:68`. Pre-implementation baseline was 48 / 529 / the same single DW-152 failure. |
| `cd folio-designer && npm run build` | **Green**, exit 0. `verify:offline` passed. `build` now runs `scan:host-fonts` beside `scan:font-hosts`; it previously ran only the latter, so the new guard was wired into nothing. |
| `cd folio-designer && npm run typecheck` | Clean, exit 0. |
| `cd folio-designer && npm run lint` | Exit 0, **exactly 4 `only-export-components` warnings**, the invariant — `src/preview/pdf-viewer.tsx` x2 and `src/App.tsx` x2. |
| `cd folio-designer && npm run scan:font-hosts` | Exit 0. 0 occurrences in **606** tracked source files (floor 400). |
| `cd folio-designer && npm run scan:host-fonts` | Exit 0. 0 occurrences of 4 spellings in **129** files (floor **86**, raised from 50 — see below). |
| `cd lint && go run ./cmd/genmanifest` then `cd lint && go test -count=1 ./...` | Regenerated; **`MANIFEST.md` is unchanged by this review** (it records licences and dependencies, and a new `_test.go` file adds no row — the "one row added" in the original entry was `fake-indexeddb`, already committed). All 4 packages green. |
| `cd folio-go && go test -count=1 ./...` | **Only the two mandated permanent reds** — `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`, `got 7, need >=20`. Every other package green, including the new `TestStoredFaceKeyTie`. The 23 golden digests under `fixtures/*/expected.pdf` are unmoved (`git status` reports no change under `fixtures/`). |
| `cd folio-designer && npm run test:e2e:compile` | Clean, exit 0. |

**The population floor was raised with the corrected number.** `POPULATION_FLOOR` was 50 against a real
population of 129 — about 39%, so a walk that collapsed to 60 files would still have reported
all-clear. It is now **86, two thirds of 129**, the same fraction this guard's model runs
(`forbidden-font-hosts.mjs`: 400 against 579-606, 67-69%). The fraction is written into the comment,
along with the note that the original 123 was measured over an untracked tree.

**`maximumCacheAssets` is still 64, and the measured `s1.assetCount` after the build is 54** — margin
10 of 64, exactly where DW-162 left it. The store consumed no cache slot, which is the assertion half
of the Block If.

**The timeout's measurement, per D-8.4j.8.**
Command: `node <scratchpad>/measure-fetch.mjs` (five repetitions per target, `fetch` + `arrayBuffer`,
wall clock). Commit `227befe`. Tree state: clean except `_bmad-output/` (the spec and sprint-status);
no source file modified at the time of measurement. Working directory `/Users/panitw/Projects/folio`.
Host: the declared repository host, `main`.

| Target | Bytes | min | mean | **max** |
|---|---|---|---|---|
| `ofl/kanit/METADATA.pb` | 4,957 | 9 ms | 80 ms | **359 ms** |
| `ofl/kanit/OFL.txt` | 4,383 | 12 ms | 65 ms | **275 ms** |
| `ofl/kanit/Kanit-Regular.ttf` | 175,148 | 14 ms | 26 ms | **58 ms** |
| `ofl/notocoloremoji/NotoColorEmoji-Regular.ttf` | 24,271,604 | 270 ms | 418 ms | **805 ms** |

**The sizing figure is the build gate's 2,097 ms for that same face, not this run's 805 ms**, because a
ceiling sized on the faster of two samples is the tighter and therefore the wronger one. The two agree
in kind — this connection is faster today — and neither loosens nor tightens the ruling.
`2,097 x 10 = 20,970 <= 30,000`. The constant carries the arithmetic, the x10 factor, and the sample's
own limit (one connection, one day, five repetitions).

**The STALL hold, with its derivation, as the spec requires it stated.** A stall ends the chain at
the **first** abort, so the hold is **T plus the requests that already completed**. The longest such
prefix is the byte read: four probes at most (max observed 359 ms each) plus the licence read (275 ms),
so **30,000 + (4 x 359) + 275 = 31,711 ms, i.e. about T + 1.7 s** on this measurement — and T + 3 s
using the gate's slower sample.

**AND THAT DERIVATION IS TRUE OF A STALL ONLY, WHICH THE ORIGINAL ENTRY DID NOT SAY** (corrected at
review, 2026-09-03; the derivation itself is unchanged and was re-checked, only the scope of the claim
was wrong). The entry ended "**Not 6 x T**" flat. That is right for a stall and wrong in general.
`AbortSignal.timeout` is armed **per request**, not once for the chain, so a host that is **slow but
alive** — every request answering just under T — never aborts anything, and a pick making n requests
can hold the control for up to **n x T**. There is no chain deadline to stop it.

**That is a ruling, not a gap.** No chain deadline was added, deliberately (Q1(3) in the Spec Change
Log below, which records the same ruling and whose "not 6 x T" is likewise the stall reading). So the
two bounds, both stated:

| Case | Bound | Why |
|---|---|---|
| The host **stalls** (nothing answers) | **T + the requests already completed** — measured 31,711 ms, about T + 1.7 s | The first abort `return`s a refusal from every catch; the loop's only `continue` is on a 404, which an abort never produces. Asserted by the table-driven first-abort test, not assumed. |
| The host is **slow but alive** (each request answers just under T) | **up to n x T** | The signal is per request, so nothing aborts. No chain deadline exists, by ruling. |

The stall is the case this story exists for, and it is the one the timeout bounds tightly. The
slow-but-alive case is bounded only by the ruling's own choice, and saying so is the honest form of the
claim.

**Red-proofs run, each by reinstating or deleting the guard rather than weakening an assertion.**

1. **Delete the release path** (`holdFontChain(false)` in `pickCatalogueFamily`'s `finally`) → *"releases
   the hold and states the degradation when the fetch times out"* reds on the still-disabled combobox.
2. **The second carrier** (D-16.R.15's actual defect: the document-reset path clears only the state copy
   of the hold, `holdFontChain(false)` → `setFontChainBusy(false)`) → *"is clear after the document is
   replaced mid-stall"* reds. This is the instance D-16.R.15 predicted, and proving (1) alone would have
   re-proved a shape that was already fixed.
3. **Change the probe catch to `continue`** → 4 tests red in `font-source.test.ts`, including the
   table-driven first-abort-terminates row on its literal call count.
4. **Empty `HOST_FONT_ACCESS_APIS`** (deleting the guard, not falsifying a condition) → 3 tests red in
   `host-font-access.test.ts`.

**Two defects the independent IndexedDB implementation found that a hand-rolled fake could not have.**
This is the concrete payoff of D-16.R.42's ruling, and it is recorded because the argument for the
dependency was made in the abstract:

1. **The transaction auto-commit.** `put` originally issued its two requests with an `await` between
   them. A transaction commits as soon as its last outstanding request settles and control returns to
   the event loop, so the second request landed on an inactive transaction: the metadata was written and
   the **bytes were not**, silently, with `put` still reporting success. The store appeared to work and
   remembered nothing. Every caller now issues its requests synchronously and combines the promises
   afterwards, and the rule is written above `transact` because it is silent when broken.
2. **`instanceof ArrayBuffer` is realm-scoped.** A buffer read back out of the store reports
   `constructor.name === 'ArrayBuffer'` and the right `byteLength` while `value instanceof ArrayBuffer`
   is **false**, because the structured clone produced it in another realm. The `instanceof` gate threw
   away every sound entry as a lost-its-bytes corruption. Replaced with a realm-independent brand check.

**The AD-26 audit, over the transitive tree, because AD-26 says "at any depth."** `npm install
--save-dev fake-indexeddb` added **exactly one** package: `fake-indexeddb@6.2.5`, **Apache-2.0**, with
**no dependencies of its own** — confirmed after the real install from `npm ls fake-indexeddb`, from the
lockfile diff (one `node_modules/` key added), from the installed package's `package.json`, and from its
own `LICENSE` file, which is the Apache License Version 2.0 and contains none of GPL, LGPL, AGPL or
SSPL. Apache-2.0 is precedented in this designer by the shipped `pdfjs-dist`. **Condition 4 did not
fire.** `lint/MANIFEST.md` gained exactly one row; `pinnedCensus` and its derived floor did not move, as
predicted — `committedLicenceFiles` enumerates tracked files and `node_modules` is untracked. A test
asserts the package is in `devDependencies` only, that `dependencies` is still exactly
`pdfjs-dist, react, react-dom`, and that **no shipping module under `src/` imports it** (walk asserted
non-vacuous first).

**One contract test was amended, narrowly and with an argument.** `src/file/file-access-contract.test.ts`
forbids `localStorage|sessionStorage|indexedDB|…` anywhere in production source — Story 5.5's "no
durable/browser-cloud fiction" rule, whose subject is **documents**. Story 16.2 adds durable state on
purpose, ruled and scoped to fetched font bytes. `font-store.ts` is exempted **by name** (the shape
`canvas-authority-contract.test.ts` already uses for `embedded-face-registry.ts`), deliberately not by
pattern, and a new test holds the exemption to its terms: the file must exist, must really be the store,
must be the **only** exempt module (asserted by re-running the filter with the name changed), and must
contain no document-shaped state (`saveTemplate`, `documentBytes`, `serialize`, `revision`, `snapshot`).
Two prose mentions of the API in `App.tsx` and `font-index.ts` were reworded to name the module rather
than the API, which is the same convention the repository already holds for font hosts. **The
prohibition is not weakened anywhere else.**

**⚠ AND THE EXEMPTION AS SHIPPED WAS FAR WIDER THAN THAT ARGUMENT — NARROWED AT REVIEW (2026-09-03).**
It returned early on the filename, which dropped the **whole** prohibition for that module: not only
`indexedDB` but `localStorage`, `sessionStorage`, `caches.open`, `showDirectoryPicker`,
`navigator.storage.getDirectory` and the entire `cloud|sync|recent files|collaborator|account` fiction
half. That is exactly backwards here. The intent contract's `Never:` clause names `localStorage` and
names it **at this module** ("Never: … store in `localStorage`"), so the one module the contract most
explicitly forbids it in had become the one module where it was unguarded. The exemption existed only
because D-16.2 **mandates** that this module write the `localStorage` refusal arithmetic into its own
prose, which a raw-text scan cannot tell from a use. It is now narrowed to exactly that: the exempt
file is scanned like every other, over a source with **comments blanked** (the idiom
`canvas-authority-contract.test.ts` already uses, so the mandated arithmetic is invisible) and the
single spelling `indexedDB` removed. Every other prohibition applies to `font-store.ts` again, and a
**mutation proof** asserts it rather than the comment claiming it: a synthetic `font-store.ts` carrying
`localStorage.setItem(…)` in CODE is reported under its own name, as are the other five prohibitions
and the fiction half, while the same words in a comment are not — and `globalThis.indexedDB` in any
OTHER file still is. Red-proved by restoring the whole-file skip, which reds that test.

**This discharges DW-165**, closed in `deferred-work.md` with the measured number, the arithmetic, the
worst-case hold and its derivation — not merely marked done.

**Deliberately NOT run, and never reported as passing:** the browser e2e specs, the matrix corpora, the
AD-21 legs and `TestCrossTargetByteIdentity`. The browser leg of this story — *fetch a family; reload;
confirm it is offered with the network disabled; remove it; confirm a document that embedded it still
renders* — is **not run here**, per the run's `end-of-run` heavy-test cadence (D-16.R.1), and is routed
to **Story 16.3's browser run** as the fourth case beside DW-161's three (ruled at build dispatch).
**Real browser IndexedDB is proven by nothing in this story.** Every store claim above rests on an
independent implementation in the test environment, which is a genuinely independent witness of the
plumbing and is not a browser.

**One risk stated rather than left implicit.** The machine-scoped preview registration reads every
stored face's bytes once per session, so an author who has fetched many large families pays that memory
at start-up. The store grows by one face per newly-picked family, so this is a handful of megabytes in
practice; the cost is written at the effect. If Story 16.3's browser run finds it material, the fix is
to register lazily from the browser dialog rather than eagerly from the listing.

### 2026-09-03 — review patches applied

Fifteen findings from a four-layer review, applied on `main` at `2a0c92a`. Nothing pushed, no branch
created, `<intent-contract>` byte-untouched (sha256 `193937cf232c65708ceca9b5a9a99779f170b28128f90437c918da300e63ce4a`
over lines 39-114 inclusive, unchanged). The verification table above is re-measured from these, and
the two corrections it carries — the second test failure and the scan population — are recorded there
with the measurement-order trap that produced them.

**The four that were red-proved by re-making the defect, not by argument.**

1. **The story's own host-font guard was RED against itself.** `src/host-font-access.test.ts:89` spelled
   the forbidden API whole in an expected value, and the whole-tree scan reads raw source and reads its
   own file. Split, the way lines 82 and 85 of that file already do. **Not** fixed by excluding the test
   file from the walk: a guard that does not read itself is a guard nobody is holding, so a new case
   asserts `scannedPopulation` **contains** `folio-designer/src/host-font-access.test.ts` and the
   scanner beside it. *Proof:* writing the spelling whole again reds the population test; excluding the
   file from the walk reds the new one.
2. **The durable-state exemption was a whole-file skip.** Narrowed to comments-blanked plus the single
   `indexedDB` spelling, with a mutation proof. *Proof:* restoring the early return on the filename reds
   the mutation proof. (Full argument in the paragraph above.)
3. **Two stored faces of one family collapsed arbitrarily.** `offeredFamilies` built a `Map` from the
   listing, so the LAST record in `list()` order — that is, hash order — silently won. Now
   `mostRecentlyFetched`: the greater `fetchedAt` (a `YYYY-MM-DD` string, so plain comparison is
   chronological), ties to the lexicographically smaller `key`, which is also the face `list()` sorts
   first within a family, so the menu and the listing agree. Two tests, each run in **both** arrival
   orders; the newer face is deliberately given the SMALLER key so a rule that was really sorting by key
   reds. **No new dropdown group** — presenting both styles is Story 16.4's. *Proof:* restoring the
   one-line `Map` build reds both.
4. **The machine-scoped preview registration was pinned by nothing.** A reviewer mutation-proved it:
   replacing `carriedFaces={paintableFaces}` with `carriedFaces={carriedFaces}` at both canvas call
   sites left the whole suite green, so the one machine-scoped registration in the application had no
   test. Added one: a document declaring only a shipped face (so `carriedFaces` is empty and the engine
   is never asked for an asset) whose projection attributes a fragment to a face **only the store
   holds** — the fragment can acquire a family only if the set the canvas is given is larger than the
   set the document carries. *Proof:* that exact swap at both sites reds it, on the exact assertion
   (`expected '' to be 'folio-carried-9b689b36…'`).

**The rest.**

- **`scan:host-fonts` was wired into nothing.** Added to `build` beside `scan:font-hosts` (its model was
  already there) and to `## Verification`. Done only after (1), or it would have broken the build.
- **The stall message asserted what the code cannot know.** It said "Your network is reachable — the
  font host is not answering". A timeout knows neither half: the same abort fires on a blackholed link
  (the captive portal this story cites as its own trigger), on a hanging DNS lookup, and on a
  connection merely too slow for a 24 MB face. That is the same class of false statement the story
  rightly refuses the offline wording for, failing in the other direction. It now says only that the
  request was made and did not complete in time, and names the three things it cannot distinguish
  between. Two tests now assert in **opposite** directions — not the offline sentence, and not a claim
  the network is fine.
- **The Go/browser key agreement was a transcribed literal, not a tie.** The digest appeared in exactly
  one designer test file; changing Go's derivation reddened Go and left that claim green. Added
  `folio-go/stored_face_key_tie_test.go` in the package that owns `embedFontFamily`: it rebuilds the
  same 110-byte fixture from the same written-down sfnt layout, derives the digest two ways
  (`sha256.Sum256`+`%x`, and the streaming API with `encoding/hex`), and asserts the **same constant**.
  Each file's comment names the other by path and symbol. The byte-length guard is kept on both sides,
  each saying the digest must be **re-derived, not adjusted**. Stated plainly in both: the tie is **two
  suites pinned to one shared constant**. A second Go test drives the real `embedFontFamily` over a real
  committed face and asserts the document is keyed by that same derivation, so the constant is tied to
  the code path and not only to itself.
- **Unobserved transaction rejection.** In `font-store.ts`'s `transact`, `settled(transaction)` was
  created before `await work(transaction)`; a rejecting request threw out of `work`, the `catch`
  returned a failed outcome, and the transaction promise then rejected with no handler. A no-op
  `catch` is now attached at creation. `await done` still sees the same rejection and still turns it
  into a failed outcome — nothing is swallowed.
- **A test whose name did not match what it exercised.** `openFontStore(undefined)` **triggers** the
  default parameter (`= globalThis.indexedDB`) rather than bypassing it, so the test named "the
  environment has no IndexedDB at all" was really exercising `globalThis.indexedDB` and passed only
  because that file's environment happens to have none. It now asserts that as a stated precondition,
  so the day a global fake is installed in that file it reds and says why.
- **A comment contradicted the frozen contract.** `font-index.ts` said `AVAILABLE LOCALLY` "means this
  arm and the local arm". The frozen block says it is "fetched faces, never host fonts"; the
  plain-terms opener says "typefaces this designer has fetched before"; the panel heading says
  TYPEFACES THIS DESIGNER HAS DOWNLOADED. The comment now agrees with all three, and records that what
  the dropdown group of that name ends up containing is **Story 16.4's to decide**. The panel heading is
  untouched.
- **The worst-case-hold claim was true only of a stall and was stated flat.** Qualified in both places
  it appears (see the two-row table above and the `## Tasks & Acceptance` bullet): `T + completed` is
  the **stall** bound; the slow-but-alive bound is **n x T**, by the deliberate ruling that added no
  chain deadline. The derivation is unchanged and was re-checked; only the scope of the claim was wrong.
- **The population floor was weak.** 50 against 129 (~39%) → **86** (two thirds, the fraction the model
  runs), with the fraction and its reason in the comment.
- **Two inaccurate claims in the Spec Change Log**, corrected in place with the substance kept: the
  frozen-block check is a **digest** (`193937cf…`, lines 39-114 inclusive, verified at `0e3b576`,
  `a378acd` and `2a0c92a`) and not the line-number claim the entry made — the commit it shipped in also
  changed the frontmatter `status` at line 5 and rewrote the opener at roughly lines 18-41, both below
  106 and both correctly outside the block; and `3c45993` is where the **pre-implementation baseline
  suites were measured**, not the commit the story was built on, which is `227befe` as the Delivery Log
  says.

**Still deliberately not run, and still never reported as passing:** the browser e2e specs, the matrix
corpora, the AD-21 legs and `TestCrossTargetByteIdentity`. Real browser IndexedDB is proven by nothing
here either.

## Suggested Review Order

**The store itself — what it keeps, and why not `localStorage`**

- Start here: the arithmetic that decided the mechanism, written into the module so it survives.
  [`font-store.ts:33`](../../folio-designer/src/font-store.ts#L33)

- The record shape: everything `embedFontFamily` refuses a face without, so a store hit needs no network.
  [`font-store.ts:122`](../../folio-designer/src/font-store.ts#L122)

- The content address, computed in the browser; this is the key, never the family name.
  [`font-store.ts:182`](../../folio-designer/src/font-store.ts#L182)

- The one open path that degrades instead of throwing into callers.
  [`font-store.ts:286`](../../folio-designer/src/font-store.ts#L286)

- Self-healing: a stored face whose bytes no longer hash to its key is dropped, not served.
  [`font-store.ts:396`](../../folio-designer/src/font-store.ts#L396)

**The timeout, and what a stall is allowed to claim**

- The whole timeout: one signal into `fetch()`, so the abort reaches the body stream too.
  [`font-source.ts:343`](../../folio-designer/src/font-source.ts#L343)

- The stall's own sentence — deliberately not the offline wording, and it claims nothing it cannot know.
  [`font-source.ts:388`](../../folio-designer/src/font-source.ts#L388)

- The three catches that each end the chain; this is why a stall costs T, not six of them.
  [`font-source.ts:412`](../../folio-designer/src/font-source.ts#L412)

**The seam Story 16.4 inherits**

- The third arm. An exhaustive switch over this union reds until 16.4 handles `stored`.
  [`font-index.ts:69`](../../folio-designer/src/font-index.ts#L69)

- Two stored faces of one family: the choice is stated and deterministic, not arrival order.
  [`font-index.ts:207`](../../folio-designer/src/font-index.ts#L207)

**Where the store meets the pick**

- The store read sits in front of the fetch; a hit sends no request at all.
  [`App.tsx:325`](../../folio-designer/src/App.tsx#L325)

- Machine-held faces join document-carried ones for preview, without disturbing the document-scoped keying.
  [`App.tsx:332`](../../folio-designer/src/App.tsx#L332)

**The two guards that had to be got right**

- The host-font prohibition, source-level over the whole designer, by every spelling.
  [`host-font-access.mjs:53`](../../folio-designer/scripts/host-font-access.mjs#L53)

- The durable-state exemption, narrowed to one spelling in one file rather than to the file.
  [`file-access-contract.test.ts:66`](../../folio-designer/src/file/file-access-contract.test.ts#L66)

**Peripherals**

- The Go half of the key tie: both suites pinned to one shared constant.
  [`stored_face_key_tie_test.go:115`](../../folio-go/stored_face_key_tie_test.go#L115)

- The population floor, at the ratio its model uses rather than a round number.
  [`host-font-access.mjs:98`](../../folio-designer/scripts/host-font-access.mjs#L98)
