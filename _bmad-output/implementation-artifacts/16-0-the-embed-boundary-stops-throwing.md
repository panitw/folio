---
title: 'Story 16.0: The embed boundary stops throwing, and refuses what render will refuse'
type: 'bug'
created: '2026-09-02'
status: 'ready-for-review'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: '09ab97d4ee04c1cfeededc567a8761d4014d63d4'
baseline_revision: '3c28c400c755941ece36740ee88b984559c746ba'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/8-6-picking-a-family-puts-it-in-the-file.md'
warnings: ['multiple-goals', 'oversized']
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Picking some typefaces fails. The panel says *"The engine returned an invalid response"* and the font
is not added. That message is not the engine refusing — the engine refusing says what is wrong and
where. It is the one thing the designer says when the part of it that talks to the engine breaks in a
way nobody anticipated, and it is deliberately the least useful sentence in the product.

The typefaces themselves are fine. Every one of the twenty-one was fed through the engine directly and
every one was accepted, so the fault is in the crossing, not the cargo. This story finds out what
actually breaks there, fixes it, and — separately and just as importantly — stops that crossing from
swallowing the reason. A failure an author can photograph but nobody can diagnose is two defects.

There is a second, quieter fault at the same place, found while planning this epic. Some typefaces —
about a quarter of what Google publishes — come as a single file holding every weight at once, and
Folio deliberately does not accept those. But the check that rejects them happens when the document is
printed, not when the typeface is added. So it is currently possible to add one, save the file, and
discover only at print time that the document cannot be printed. The refusal is right; it simply
arrives far too late. This story moves it to the moment of the click, where the engine's own message
already tells the author exactly what to do.

<intent-contract>

## Intent

**Problem:** Embedding a catalogue face fails intermittently at the WASM/worker boundary with
`WASM_PROTOCOL_FAILURE` / *"The engine returned an invalid response"*. Reported by the owner on
2026-09-02 against `main` at `a40c34d`, with a screenshot showing the message under the family
control. **Every acceptance in Epic 16 runs through this exact path**, so the epic cannot be verified
over it.

**Second problem, found at plan time and measured (D-16.6):** the **embed command accepts a variable
face that the renderer refuses.** `embedFontFamily`'s only structural check is `checkSfnt`, which does
not look at `fvar`; `fontset.New` refuses `fvar` at `internal/fontset/fontset.go:228`. **A pick can
therefore write a `.folio` that saves cleanly and fails at render** — the one outcome D-8.4d.1 and
D-16.1 both promise cannot happen. Reachable today, without any of Epic 16.

**Approach:** Diagnose the throw at the boundary with the Go layer already excluded by measurement, fix
the cause, and remove the boundary's ability to erase a cause — `execute`'s bare `catch` reports
`WASM_PROTOCOL_FAILURE` for every throw it sees and keeps none of them. Separately, add the missing
`fvar` refusal to `embedFontFamily`, so the engine's existing, already-actionable message reaches the
author at the pick instead of at the render.

**`multiple-goals` ACCEPTED, and the reason is recorded (D-16.6), not waived.** Both defects are the
same boundary admitting or erasing what it should have refused or reported; both are verified by the
same probe harness; and a split would put two halves of one guard in two stories.

## Boundaries & Constraints

**Always:**
- **The Go command layer is already cleared, and the measurement is not repeated on faith.** Measured
  2026-09-02 (wd `folio-go`, `go test ./wasm` over a probe applying `embedFontFamily` to every
  `folio-designer/public/fonts/*/*.ttf` through `wasm.Engine.Apply`, tree clean at `a40c34d`): **21 of
  21 catalogue faces pass.** The single failure was `notosanssc` (10,595,932 bytes), refused correctly
  and located: `folio: face exceeds the 6288384-byte supported size`. That face is not in the
  catalogue. **A diagnosis that lands in `component_commands.go` contradicts this measurement and must
  explain it rather than ignore it.**
- **The message's provenance is exact.** `The engine returned an invalid response` appears **once** in
  the repository, at `folio-designer/src/engine.worker.ts:82`, reached only from `execute`'s
  `catch`. It therefore means `host.handle(...)` threw, or `JSON.parse` of its output threw. It is
  **never** an engine refusal — those carry `result.diagnosticCode` and a located message through the
  arm at `:64-72`.
- **Report the cause, do not merely stop the throw.** If the fix removes the symptom without the
  boundary gaining the ability to say what happened, the second defect is still shipped.
- **A refusal must stay a refusal.** `notosanssc`'s located size error is correct behaviour and must
  still arrive as a located message, not be swept into whatever this story changes.
- **The `fvar` refusal is ADDED at the command, and the renderer's is KEPT.** `fontset.go:228` stays:
  a `.folio` can be hand-written and the loader is not the only door. This is the shape
  `component_commands.go` already states for `setComponentAsset` — *"bytes this build cannot read are
  refused before anything is written to `t.doc.Assets`"* — applied to the one class it currently
  misses.
- **The command's refusal reuses the engine's existing message.** `fontset.go:228` already ends with
  the `fonttools varLib.instancer` remedy and states why it names one: *"most Google Fonts downloads
  are variable builds today, so a caller hitting this needs an action, not a refusal."* Do not write a
  second, worse sentence for the same fault.
- **ONE predicate, not two authorities (engineering lead, at this plan gate).** The command's `fvar`
  refusal and the renderer's must be **the same code**, not two implementations that happen to agree
  today. There is nothing to share right now: the check is an inline `if parsed.HasTable(ot.TagFvar)`
  in the body of `fontset.New`, with its message inline beside it and no helper, no error type and no
  sentinel. Extract it into **one** exported helper in `internal/fontset` and call it from **both**
  sites, and add a test that feeds **the same bytes** to both and asserts both refuse.
  `internal/fontset` is the only honest home: package `folio` may import it (`render.go:13` and
  `wrap.go:7` already do), whereas siting the rule in `internal/template` beside `checkSfnt` would
  leave `fontset.New` carrying its own copy — a second authority by construction, because
  `internal/fontset` does not import `internal/template`.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **The diagnosis cannot be evidenced.** A plausible story about why it throws is not a diagnosis. If
  the cause cannot be observed, say so and report what was excluded rather than fixing a guess.
- **The `fvar` check would be moved rather than added.** Removing the renderer's guard because the
  command now has one leaves a hand-written `.folio` unguarded. Both, or halt.
- **The fix would loosen a bound.** `MAX_ENGINE_PAYLOAD_BYTES` (8 MiB), the 6,288,384-byte face cap and
  `base64ToBytesBounded`'s guard are limits, not obstacles. Raising one to make a symptom go away is a
  halt.
- **Any of the golden digests moves.** Nothing here should touch a rendered byte.

**Never:** change what the engine accepts · change the command's field-count contract
(`font-chain-command.ts`'s "THE FIELD COUNT IS PART OF THE CONTRACT") · add a retry that hides a
deterministic fault.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Pick a small face | e.g. `notoserifthai`, 26,496 bytes | Embeds; chain declared | No error |
| Pick a large catalogue face | e.g. `cascadiacode`, 598,060 bytes → ~1.06 MB doubly-encoded | Embeds; chain declared | No error |
| Face over the engine's cap | 10,595,932-byte face reaching the command | **Located refusal** naming the cap | Refusal, anchored at the control |
| Payload over the protocol bound | > `MAX_ENGINE_PAYLOAD_BYTES` | Admission refusal, unchanged | Refusal, not a throw |
| Genuine WASM fault | `host.handle` throws | A message that names what threw | Reported, not erased |
| Boot failure | wasm never registered | `WASM_INITIALIZATION_FAILED`, unchanged | Lifecycle failure path |
| Embed a variable face | `Anuphan[wght].ttf`, 231,712 bytes, carries `fvar` | **Refused at the command**, located, with the instancer remedy named | Refusal anchored at the control, not a render-time surprise |
| Load a hand-written `.folio` embedding a VF face | Command bypassed entirely | Renderer still refuses at `fontset.go:228` | Unchanged; the guard is not moved |
| Embed a static face | `Kanit-Regular.ttf`, no `fvar` | Accepted, unchanged | No error |

</intent-contract>

## Code Map

**Every anchor below was re-measured at `09ab97d4ee04c1cfeededc567a8761d4014d63d4` (this plan gate).**
The line numbers the hand-authored draft carried were measured against an earlier tree and most were
off by one or more; the corrections are recorded in the Spec Change Log. Trust these, not the draft's.

**Designer (`folio-designer/`)**
- `src/engine.worker.ts:58-83` — `execute`. **The subject.** `:59-63` builds the request; `:62-63`
  base64-encode through `bytesToBase64`; `:64` `JSON.parse`; `:65-73` the engine-refusal arm
  (`if (!result.ok || !result.snapshot)`, `diagnosticCode` at `:67`); `:81-83` the bare `catch`, with
  `WASM_PROTOCOL_FAILURE` at `:82` — **and it keeps nothing**.
- `src/engine.worker.ts:82` — `'The engine returned an invalid response'`. **Verified unique at HEAD**:
  the only occurrence in any source, test or e2e file repo-wide (the remaining hits are `_bmad-output/`
  prose). It is never an engine refusal.
- `src/engine.worker.ts:96` — `bytesToBase64`: `for (const byte of new Uint8Array(bytes)) text +=
  String.fromCharCode(byte)`. Byte-at-a-time concatenation over the whole payload. Contrast the
  **chunked** encoder at `src/font-chain-command.ts:93-99`, whose comment at `:89-92` explains that
  `String.fromCharCode(...bytes)` on a 480 KB face "spreads half a million arguments across the call
  stack and throws". **The worker's encoder never got that treatment.**
- `src/engine.worker.ts:97-104` — `base64ToBytesBounded`. Throws
  `new Error('WASM byte response exceeds its transport limit')` at **`:100`** (pre-decode arithmetic)
  and again at **`:102`** (post-`atob` recheck). Both land in the `:81` catch, so a **response**-side
  breach is today indistinguishable from a **request**-side fault.
- `src/engine-protocol.ts:6` `MAX_ENGINE_PAYLOAD_BYTES = 8 * 1024 * 1024`; `:7`
  `MAX_ENGINE_RENDER_PDF_BYTES = 32 * 1024 * 1024`; `:15` `MAX_ENGINE_FONT_FAMILIES = 256`.
- `src/engine-worker-admission.ts:13` — `admit` is a **method on class `EngineRequestAdmission`**
  (class at `:10`), not a free function. Refusals here are `fatal` or per-request and do **not** reach
  `execute`.
- `src/App.tsx:608-627` — `pickCatalogueFamily`: fetch at `:612`, `arrayBuffer()` at `:614`, its own
  fetch-failure refusal text at `:621`, one `embedFontFamilyCommand` via `applyFontChain` at `:626`.
  **It hardcodes `mediaType: 'font/ttf'`** — worth knowing before suspecting media-type negotiation.
- `src/font-chain-command.ts:69-87` — `embedFontFamilyCommand`; `:15` carries
  `// THE FIELD COUNT IS PART OF THE CONTRACT.` It emits 11 named fields plus `kind`/`version`,
  matching `componentFields(raw, 12)` in Go.
- `src/font-catalogue.test.ts:138` — **the designer already knows how to look for `fvar`**, in Node,
  over the shipped catalogue: `variableTables: ['fvar','gvar','avar','HVAR','MVAR'].filter(...)`. This
  is *not* a second authority to reconcile — it inspects committed build inputs, never author bytes.
- `src/generated/font-catalogue.ts:32` — **21 catalogue entries** (21 unique ids and families); that is
  what "all 21 families" means. `public/fonts/` holds **27** directories; the six not catalogued are
  `ibmplexmono`, `ibmplexsans`, `ibmplexsansthai`, `notosans`, `notosanssc`, `notosansthai` — the last
  three being the `scriptFallbackFaces` at `:30`. **Every catalogue face is static by construction**:
  each `public/fonts/*/NOTICE.md` asserts "no `fvar`/`gvar`", and `scripts/build-wasm.mjs:191-210`
  validates every entry at build time.

**Go (`folio-go/`)**
- `component_commands.go:2359` — `embedFontFamily` (doc from `:2330`). **The function does not end at
  `:2410`** as the draft said; that line is mid-body and the body runs past `:2440`.
  `componentFields(raw, 12)` at `:2360`; the **only** structural gate is `template.DecodeFontForRender`
  at `:2391`; the comment at `:2384-2390` states the very rule this story extends — *"bytes this build
  cannot read are refused before anything is written to `t.doc.Assets`"*. Content-hash dedupe at
  `:2397-2399`.
- `component_commands.go:667` — **`maxComponentAssetBytes` is DERIVED, not a literal**:
  `(engineProtocolMaxPayloadBytes - maxComponentAssetPayloadOverheadBytes) * 3 / 4` = `(8388608 - 4096)
  * 3 / 4` = **6288384**, from `:637` and `:648`. The face refusal is formatted at `:2513` inside
  `embeddedFaceBytes` (declared `:2496`): `"face exceeds the %d-byte supported size"`. **Grepping code
  for `6288384` finds nothing** — that number exists only in prose.
- `component_commands.go:672-683` and `:749-752` — `setComponentAsset`'s doc and its AC1 comment,
  carrying the codebase's own name for this story's guardrail: *"reusing image.go's
  `DecodeImageForRender` rather than a second capability check"*.
- `internal/template/fontasset.go:178-187` — `decodeRecognisedFont` (`font/ttf`, `font/otf` only, at
  `:180`); `:197-207` `DecodeFontForRender`; **`checkSfnt` at `:249`**, its doc at `:230-248`.
  **`checkSfnt` is the gap, and its own doc fences it**: *"it establishes that the file is not lying
  about what it is, and nothing more … does not validate a single table's contents; all of that belongs
  to the shaper, at render — internal/fontset."* Its loop computes `record := headerSize + i*recordSize`
  and reads offset/length at `record+8`/`record+12` while **never inspecting the tag at `record+0`**.
  `DecodeFontForRender`'s own fence (`:194-196`) is the sharper one: *"can this build read these bytes
  as a single face, and nothing more"* — and a variable face **is** readable as a single face. That is
  why the new rule does not belong here.
- `internal/fontset/fontset.go:226` — `if parsed.HasTable(ot.TagFvar)`, inside `New` (declared `:85`),
  message inline at `:228-236` naming `fonttools varLib.instancer` and pointing at
  `tools/fontgen/instance_faces.py`. **`parsed` is an `*ot.Font` from the vendored `textshape/ot`
  library, not a hand-rolled walk**, and there is **no named error type, no sentinel and no constant** —
  a bare `fmt.Errorf`. `:782-796` (note at `:785`) explains why ingestion is where it lived. Sibling
  `HasTable` uses at `:344` and `:509`.
- `render.go:1469` — `fontset.New(name, raw)` in `fontCache.parseEmbedded`: **this is where a variable
  face fails today, long after the document was saved.** The caller-FontSet arm is `:1445`.
- `render.go:13`, `wrap.go:7` — package `folio` **already imports `internal/fontset`**, so the shared
  helper is reachable from `component_commands.go`, whose imports at `:3-19` are stdlib plus
  `internal/expr`, `internal/geom` and `internal/template` (`internal/fontset` must be added).
- `render_arch_test.go:461` — **`TestFolioMethodNamesAreInjective`, and it constrains this story.** No
  two receiver types in non-test root files may share a method name. `Error` is already declared on a
  root receiver, so **a new error type cannot be declared in package `folio` root** — it could not have
  an `Error() string`. Vacuity guard at `:504` (`seenAtEpic4Planning = 7`). This is a second,
  independent reason the shared predicate belongs in `internal/fontset`, where a bare `fmt.Errorf`
  needs no type at all.
- `testdata/fonts/notosansthai-variable-testonly/NotoSansThai-VF.ttf` (213.5 KB, with `LICENSE-OFL.txt`
  and `NOTICE.md`) — **the committed variable face. The same-bytes test needs no network and no new
  fixture.** It is already embedded into **package `folio`** — the very package `embedFontFamily` lives
  in — as `testNotoSansThaiVariableFontBytes` at `testfont_embed_test.go:32`. It is also reachable in
  `internal/fontset` via `fontset_test.go:441`'s `testVariableFontBytes(t)`, which feeds
  `TestNewRejectsVariableFace` at `fontset_test.go:573`.
- **No test today feeds variable bytes through the command.** `grep` for `ariable`/`VF` over
  `component_commands_test.go`, `component_asset_command_test.go` and `embedded_font_fixture_test.go`
  returns zero hits. That disjointness is exactly the shape that lets two sites drift apart.
- `byte_neutrality_test.go:100` `goldenDigestRecord`, `:611` `goldenDigestSearchScope`, `:631`
  `TestGoldenDigestAgreesAtEveryDeclaredSite` — the golden guard, executed by plain `go test ./...`.
- **`tools/fontgen/instance_faces.py` is at the REPO ROOT, not under `folio-go/`.** The draft said
  `folio-go/tools/...`; there is no `folio-go/tools/` directory at all. Its header at `:44-52` explains
  why its output is committed rather than generated.

## Tasks & Acceptance

**Execution:**
- **Reproduce in a real browser first, and keep the harness.** `folio-designer/e2e/` is
  compile-checked and never executed (D-000.4), which is how this reached the owner. Add a spec under
  `folio-designer/e2e/` that drives `pickCatalogueFamily` over **all 21 entries of
  `src/generated/font-catalogue.ts`** — importing the catalogue rather than hardcoding a list, so a
  22nd family cannot silently escape it — and records, per family: embedded, or refused with its
  located text, or `WASM_PROTOCOL_FAILURE`. Run it (the Verification section below carries the exact
  invocation) and record the per-family table **before** changing anything, then again after.
  - **Preconditions already measured at this gate, so none of these is a research task:** the pinned
    Chromium revision **1208 is already installed** at
    `~/Library/Caches/ms-playwright/chromium-1208/chrome-mac-arm64/Google Chrome for Testing.app`.
    **`npx playwright install chromium` is NOT required and must not be run** — the download is
    unnecessary, and that pinned archive has previously returned HTTP 400 from the download host.
    `playwright.config.ts` exposes `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` for exactly this case, and
    prior epic boundary gates used it. Note `chromium_headless_shell-1208` is **absent** (only
    1217/1223/1228 are present), so pass the headed executable explicitly rather than letting
    Playwright resolve a headless shell.
  - `playwright.config.ts`'s `webServer` runs a full `npm run build` — which includes `build:wasm`, so
    the **Go toolchain must be on `PATH`** — with `reuseExistingServer: false` and a 180 s timeout. A
    prior single-spec gate run took ~1.9 minutes end to end. Budget for it; do not mistake it for a
    hang.
  - **The browser cannot prove the second goal, and must not be made to.** Every catalogue face is
    static by construction, so no catalogue pick can exercise the `fvar` refusal. **Do not add a
    variable face to the catalogue in order to create one** — `scripts/build-wasm.mjs` and the
    per-family `NOTICE.md` both stand against it. The browser proves the throw is gone; the `fvar`
    guard is proven Go-side over the committed fixture.
- **Instrument the boundary before changing it.** Capture the actual exception at
  `engine.worker.ts:81` — name, message and, where available, whether it came from `host.handle` or
  `JSON.parse` — and record the observation with command, commit, tree state and working directory
  (D-8.4j.8).
- **Fix the cause the evidence names.** If it is the encoder, chunk it the way
  `font-chain-command.ts:93-99` already does, and say in the code why the two encoders now agree. If it
  is elsewhere, the fix goes where the evidence points.
- `src/engine.worker.ts` — **the second defect, fixed regardless of the first.** Split the `catch` so a
  transport-bound breach, a `JSON.parse` failure and a `host.handle` throw are distinguishable, and
  carry the thrown value's message into the reported error, bounded like every other string on this
  protocol.
- `src/engine-protocol.test.ts` / a worker-level test — red-prove the new reporting: a forced throw
  must produce a message naming it, and the test must fail if the message is reverted to the bare
  string.
- Keep `notosanssc`'s located size refusal arriving as a located refusal; assert it.
- `folio-go/internal/fontset/fontset.go` — **extract the predicate first, before adding any caller.**
  Split the inline `if parsed.HasTable(ot.TagFvar)` block at `:226` into one helper owning both the
  test and the message, and export a byte-taking wrapper (the command holds `[]byte`, not an
  `*ot.Font`, so the wrapper re-runs `ot.ParseFont` — a cost, not a correctness question). `New` must
  then call that helper rather than keep a copy: **if `New` still contains its own `HasTable` test
  after this task, the task is not done.**
- `folio-go/component_commands.go` — **add the `fvar` refusal to `embedFontFamily`** (D-16.6) by
  calling that one helper, beside the existing `DecodeFontForRender` call at `:2391` and before
  anything is written to `t.doc.Assets`. Add `internal/fontset` to the import block at `:3-19`.
  **Declare no new error type in package `folio` root** — `render_arch_test.go:461` forbids it — and
  route the helper's message through `componentFailure("", fontChainPath(name), …)` as the sibling
  refusals do.
- `folio-go/component_commands_test.go` — **the same-bytes test the engineering lead requires.** Feed
  `testNotoSansThaiVariableFontBytes` (already embedded in this package at `testfont_embed_test.go:32`)
  to **both** `embedFontFamily` **and** `fontset.New` within one test, and assert **both** refuse.
  Feeding one byte slice to both is the whole point: two tests over two fixtures would re-create the
  disjointness this guard exists to close. Also assert a static face (`testRobotoFontBytes`) is still
  accepted, so the guard cannot be over-broad, and assert **no key was written to `t.doc.Assets`** on
  the refusal — "returned an error" is a weaker claim than "wrote nothing".
- **Red-prove by DELETION, not by falsifying a condition.** Deleting a call site proves it is reached;
  inverting a predicate only proves arm order. Delete the new call in `embedFontFamily`, confirm the
  same-bytes test fails, restore it; then do the same for `New`'s call.
- `folio-go/internal/fontset/fontset_test.go` — `TestNewRejectsVariableFace` (`:573`) must still pass
  unchanged, so a later reader cannot conclude the command's check made the renderer's redundant.

**Acceptance Criteria:**
- Given every catalogue family, when each is picked in a real browser, then each is embedded or
  refused with a located reason, and none produces `The engine returned an invalid response`.
- Given a fault at the WASM boundary, when it is reported, then the message names what threw rather
  than only that something did.
- Given a face larger than the engine's supported size, when it is picked, then the located cap
  refusal still arrives unchanged.
- Given the diagnosis, when it is recorded, then it carries command, commit, tree state and working
  directory, and it either explains the Go-layer measurement or is consistent with it.
- Given a variable face, when it is embedded, then the **command** refuses it with the located message
  that names the instancer remedy — and no byte reaches `t.doc.Assets`.
- Given a hand-written `.folio` that embeds a variable face without passing through the command, when
  it is rendered, then the renderer's own guard still refuses it.
- Given one slice of variable-face bytes, when it is put through the command and through `fontset.New`
  in a single test, then both refuse, and both refusals originate in **one** shared helper — evidenced
  by deleting that helper's call sites and observing each deletion red the test.
- Given a static face, when it is embedded, then nothing about today's behaviour changes.

## Spec Change Log

**2026-09-02 — plan-gate hardening dispatch (`bmad-build-auto`, halt after planning), at HEAD
`09ab97d4ee04c1cfeededc567a8761d4014d63d4`.** This spec was hand-authored and committed at
`ready-for-dev` before the dispatch. It was set to `draft` for the duration so step-01 would route to
planning rather than straight to implementation, then returned to `ready-for-dev`. The
`<intent-contract>` block was carried across **verbatim** apart from the single orchestrator-directed
amendment in item 1; the slab was diffed against the committed version to prove it.

1. **One addition inside the intent contract, to Boundaries & Constraints "Always": the ONE-predicate
   guardrail**, supplied by the engineering lead through the dispatch. Ten lines added; nothing
   removed, reworded or reordered. It was **not** already implied. The draft required the command's
   refusal to *reuse the engine's message* and required the renderer's guard to be *kept*, but nothing
   in it required the two to be **the same code**, and nothing required a test over **the same bytes** —
   the draft's two tests used two different fixtures in two different packages, which is precisely the
   shape that lets two sites drift apart.
2. **Code Map re-measured wholesale at HEAD.** The draft's anchors were measured against an earlier
   tree. Line corrections: `execute` `:57-83`→`:58-83`; the refusal arm `:64-72`→`:65-73`; the catch
   `:80-82`→`:81-83`; `bytesToBase64` `:99`→**`:96`**; `base64ToBytesBounded` `:100-108`→**`:97-104`**;
   the chunked encoder `:92-98`→`:93-99`; `decodeRecognisedFont` `:178-184`→`:178-187`;
   `DecodeFontForRender` `:197-206`→`:197-207`; `checkSfnt` now anchored at `:249` with its doc at
   `:230-248`; the `fvar` refusal `:226-239`→`:226-238`. Three were **not** line drift but errors of
   fact: `admit` is a **method on `EngineRequestAdmission`** (`:13`), not a free function;
   `embedFontFamily` does **not** end at `:2410`, which is mid-body; and `instance_faces.py` sits at
   the **repo root** — there is no `folio-go/tools/` directory at all.
3. **Facts added that change how the work must be done**, none of which the draft carried: package
   `folio` already imports `internal/fontset` (`render.go:13`, `wrap.go:7`), so the shared helper is
   reachable; `render_arch_test.go:461` **forbids a new error type in package `folio` root**, which
   independently forces the helper into `internal/fontset`; the variable-face fixture is **already
   embedded in package `folio`** as `testNotoSansThaiVariableFontBytes`, so the same-bytes test needs
   neither network nor new fixture; `6288384` is derived at `component_commands.go:667` and appears as
   a literal nowhere in code; and every catalogue face is static by construction, so the browser run
   structurally cannot exercise the `fvar` half.
4. **Verification rewritten from claims into commands.** The draft named five things to check and gave
   a runnable command for two. It also instructed a re-run of the D-16.6 probe over
   **`Anuphan[wght].ttf` — a file that exists nowhere in this repository**; it was fetched from
   `raw.githubusercontent.com/google/fonts` for the original throwaway measurement. That instruction
   was unexecutable in a repo built around an offline bundle and a `scan:font-hosts` gate, and it is
   replaced by the committed fixture. `Anuphan` deliberately **remains** in the I/O matrix inside the
   contract, where it is correct: there it is the measured *subject* of D-16.6, not a repo artifact.
5. **The heavy-test cadence override is now stated in the spec itself**, with the Story 1.2 precedent
   named, rather than living only in the dispatch prompt.
6. **`oversized` added to `warnings`** — the spec is 33,716 bytes (~8,400 tokens) against a 1,600-token
   threshold. It is not clearable at this size and the flag is a project-wide convention gap, not a
   defect here: siblings `8-4`, `8-4c` and `15-1` all carry it.

## Design Notes

**Why this is Story 16.0 and not a defect ticket.** Every acceptance in Epic 16 ends in "the font is
added", and every one of those runs through `execute`. A story that cannot distinguish "my feature is
broken" from "the boundary is throwing again" cannot be reviewed. It is placed first for the same
reason 15.1 was pulled forward as story zero of the Epic 7–8 run: the run's own gates are unevaluable
while it is red.

**The strongest available hypothesis, stated as a hypothesis.** `bytesToBase64` is the one function on
this path that this repository has already fixed **somewhere else for a documented reason**, and the
worker's copy predates that fix. That makes it the first place to look. It does not make it the
answer: a byte-at-a-time concatenation is slow rather than throwing, so if it is the cause, the
mechanism has to be shown, not assumed.

**D-16.R.2 does NOT make the second defect redundant, and a later reader will be tempted to think it
does.** That owner decision changes 16.1 and 16.3 from *show-and-refuse* to *filter-out*: a
variable-only family is simply not offered in the font browser. It would be easy to conclude from that
that a command-level `fvar` refusal has nothing left to refuse. The decision itself forecloses that
reading in one sentence — *"the refusal path stays in the engine for anything that slips through — a
hidden row is a presentation choice, never a guard."* Hiding a row is UI; this story is the guard
underneath it, and it is reachable today with none of Epic 16 built.

**Why the second defect belongs here and not in 16.1.** 16.1 is about where a font comes from. This
is about what the boundary will accept once it has one — the same boundary, the same probe harness, and
a guard whose absence is a defect whether or not Epic 16 is ever built. Placing it in 16.1 would make
a shipped bug look like a consequence of the new feature, which is the opposite of true.

**What "some fonts" is worth as a signal.** The owner reported failure on *some* picks. The catalogue
spans 26 KB to 646 KB, so a size-dependent boundary fault is consistent with that report and a
content-dependent one is not — the Go layer accepted all 21 sets of bytes.

## Verification

**This story overrides the run's heavy-test cadence, and the override is already a recorded decision —
not this dispatch's invention.** D-16.R.1 sets the run to `end-of-run` and then names **Story 16.0** as
one of exactly two stories that override it: *"its acceptance criterion is a browser run … its whole
subject is a boundary that no Go test can observe."* The
same shape of override was taken once before, for Story 1.2, on the same ground: the heavy suite *is*
the deliverable. The defect here is browser-shaped and reached the owner **precisely because**
`folio-designer/e2e/` is compile-checked and never executed. A verification pass that proved
this story without a real browser would be reproducing the conditions that produced the bug. Run in
this order, and report what each command printed rather than that it was green.

1. **The browser run — required, not optional.** From `folio-designer/`, with the Go toolchain on
   `PATH`:
   ```
   PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH="$HOME/Library/Caches/ms-playwright/chromium-1208/chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing" \
     npx playwright test e2e/<new-spec>.spec.ts --reporter=list
   ```
   Report the **per-family table across all 21 catalogue families**, before and after the fix, plus the
   pass/fail line. If the run cannot be made to happen, **say so explicitly and name the obstacle** — do
   not substitute `npm run test:e2e:compile`, which is only `tsc --noEmit`, and never report a compiled
   spec as a verified one. **Do not run `npx playwright install chromium`**: revision 1208 was measured
   present on disk at this gate, and that archive has previously returned HTTP 400 from the host.
2. **The same-bytes guard, red-proved by deletion.**
   `cd folio-go && go test -run 'VariableFace|Fvar|EmbedFontFamily' ./... -v`. Then delete the call site
   in `embedFontFamily`, re-run, confirm it **fails**, restore. Repeat for `fontset.New`'s call site.
   Report both directions for both sites; a green run over correct code proves nothing on its own.
3. `cd folio-go && go test ./...` — **expect exactly one failure, the mandated P6g red.** A second
   failure is a regression. This run is also what executes the golden guard
   (`TestGoldenDigestAgreesAtEveryDeclaredSite`).
4. `cd folio-go && go vet ./...`, and `cd lint && go test ./...`.
5. `cd folio-designer && npm run typecheck && npm run lint && npm run test && npm run test:e2e:compile`
   — expect 4 pre-existing `only-export-components` lint warnings; they are not a regression.
6. **The 23 golden digests, unmoved.** From the **repo root**: `shasum -a 256 fixtures/*/expected.pdf`.
   Measured at this plan gate (HEAD `09ab97d4`): **23 files**, rolled digest
   `892a1505e5e7fff0184310d5f70eb7bfcfa10d18cda9af4e2aecf262a0630ce9` — unchanged from the Epic 9/10
   boundary gate. Report the file count **and** the rolled digest: a changed count is as much a failure
   as a changed digest, and an empty listing must never be read as "nothing moved".
7. **`notosanssc`'s located size refusal still arrives located.** It is **not** in the pickable
   catalogue, so step 1's browser run will not reach it — assert it Go-side instead. The text is
   `"face exceeds the %d-byte supported size"` (`component_commands.go:2513`) over the **derived**
   `maxComponentAssetBytes` (`:667`); do not hardcode `6288384` in a new assertion.
8. **The D-16.6 disagreement, re-measured at implementation HEAD**, using the committed fixture rather
   than the unobtainable `Anuphan[wght].ttf`: `testNotoSansThaiVariableFontBytes` must be refused at
   the command **and** still refused at ingestion, both from the one shared helper.
## The Diagnosis (D-8.4j.8 form: command, commit, tree state, working directory)

**The hypothesis in Design Notes is REFUTED, and the cause is elsewhere.** `bytesToBase64`'s
byte-at-a-time concatenation is slow, not throwing; it was never reached with a bad value.

**Step 1 — the throw, observed in a real browser.** Measured 2026-09-02, wd `folio-designer`, at
HEAD `3c28c400c755941ece36740ee88b984559c746ba` with a working tree carrying only the new probe
spec:

```
PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=".../chromium-1217/.../Google Chrome for Testing" \
  npx playwright test e2e/font-embed-boundary.spec.ts --reporter=list
```

19 of 21 families embedded; **Cascadia Code (598,060 B) and Cascadia Mono (575,912 B)** — the two
largest catalogue faces — reported `The engine returned an invalid response`. Ubuntu Sans (487,492 B)
is the largest that passed, so the fault is size-dependent, which is what "some fonts" was worth as
a signal. With the boundary's catch split (this story's second fix) the same run reported instead:

> `The engine's response could not be parsed: SyntaxError: "undefined" is not valid JSON`

i.e. **`host.handle(...)` returned `undefined`**, not a string. That is the throw, and it took one
run to obtain once the boundary stopped erasing it — which is the argument for the second fix in one
sentence.

**Step 2 — the cause, isolated in Node against the built wasm.** Measured 2026-09-03, wd
`folio-designer`, same HEAD, working tree carrying the probe spec only; a throwaway script drove
`FolioWasmHost.handle` directly over every `public/fonts/*/*.ttf`. Go printed a full fatal trace
before exiting:

```
runtime: out of memory: cannot allocate 524288-byte block (4234510336 in use)
fatal error: out of memory
  template.decodeBase64Asset  internal/template/base64.go:23
  template.writeAssets        internal/template/serialize.go:534
  template.writeDocument      internal/template/serialize.go:141
  folio.SerializeTemplate     serialize_template.go:14
  folio.applyFontChainCommand component_commands.go:2015
  wasm.(*Engine).Apply        wasm/engine.go:231
  main.dispatch               wasm/cmd/engine/main.go:149
```

`decodeBase64Asset` joined the canonical 76-column split with `joined += part` — one fresh string
per element, copying everything accumulated so far. A 598,060-byte face is 10,493 elements, so the
total allocated is `sum(76, 152, …, 797,416) ≈ 4.18 GB`. **The re-added quadratic join measures
4,226,609,792 bytes in `TestDecodeBase64AssetJoinsInLinearSpace`**, against the wasm runtime's
reported `4234510336 in use`. Go's runtime then exits, `FolioWasmHost.handle` is gone, and the
wrapper returns `event.result` — `undefined`.

**It EXPLAINS the plan gate's Go-layer measurement rather than contradicting it.** That probe ran
`go test ./wasm` natively on darwin/arm64: 64-bit address space, a real GC, and 4.18 GB of
short-lived allocation is merely slow. **The 4 GiB ceiling is js/wasm's 32-bit linear memory**, and
nothing about the bytes differs between the two — which is exactly why 21 of 21 passed natively and
2 of 21 failed in the browser. The diagnosis therefore lands in `internal/template/base64.go`, on
the path `component_commands.go:2015` reaches, without landing *in* `component_commands.go`.

**No bound was loosened.** `MAX_ENGINE_PAYLOAD_BYTES`, the derived 6,288,384-byte face cap and
`base64ToBytesBounded`'s guard are all untouched; the fix removes an allocation, not a limit.

## Auto Run Result

Status: implemented, awaiting review

**Browser run, after the fix (the acceptance):** all 21 catalogue families EMBEDDED, none producing
`The engine returned an invalid response`; both specs pass in 2.2 minutes. The per-family table is
in the run output and in the section above for the "before" state.

**One plan-gate precondition was WRONG and is corrected here.** The gate recorded Chromium revision
**1208 as installed** at `~/Library/Caches/ms-playwright/chromium-1208/...`. The path exists; **the
browser does not.** That directory is **428 KB** and contains only `Contents/MacOS` and
`Contents/Resources` — the `Contents/Frameworks/Google Chrome for Testing Framework.framework` is
absent, and launching it aborts with a `dlopen` failure. It is a truncated download, consistent with
the HTTP 400 the gate noted. `npx playwright install chromium` was **not** run, as directed.
**Revision 1217 was used instead** (336 MB, complete, present on disk), passed through the same
`PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` the config already exposes. A future run of this spec should
use 1217, 1223 or 1228 — not 1208 — until 1208 is re-downloaded.

**Standing reds, by identity.** `go test ./...`: exactly one, `TestCorpusMeetsP6ExerciseFloors/P6g`
(`got 7, need >=20`), the mandated P6g red. `npm run test`: one,
`canvas-authority-contract.test.ts` flagging `getComputedStyle` in `e2e/e9-5-border-no-ink.spec.ts`
— **verified pre-existing** by stashing every change in this story (including the new e2e spec) and
re-running against the clean tree at `3c28c400`, where it fails identically. Not a regression, and
not this story's file.

**Golden digests unmoved:** 23 files, rolled digest
`892a1505e5e7fff0184310d5f70eb7bfcfa10d18cda9af4e2aecf262a0630ce9`.
