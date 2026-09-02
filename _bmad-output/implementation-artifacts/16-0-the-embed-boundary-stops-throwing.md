---
title: 'Story 16.0: The embed boundary stops throwing, and refuses what render will refuse'
type: 'bug'
created: '2026-09-02'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: 'a40c34db6cff7372363b2a553710eff48759bef1'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/8-6-picking-a-family-puts-it-in-the-file.md'
warnings: ['multiple-goals']
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

**Designer (`folio-designer/`)**
- `src/engine.worker.ts:57-83` — `execute`. **The subject.** `:59-62` builds the request; `:61`
  base64-encodes the payload through `bytesToBase64`; `:63` `JSON.parse`; `:64-72` the engine-refusal
  arm; `:80-82` the bare `catch` that reports `WASM_PROTOCOL_FAILURE` **and keeps nothing**.
- `src/engine.worker.ts:99` — `bytesToBase64`: `for (const byte of new Uint8Array(bytes)) text +=
  String.fromCharCode(byte)`. **Byte-at-a-time string concatenation over the whole command payload**,
  which for a 598,060-byte face is ~800 KB in and ~1.06 MB out. Contrast
  `src/font-chain-command.ts:92-98`, whose own comment explains why *that* encoder is chunked:
  `String.fromCharCode(...bytes)` on a 480 KB face "spreads half a million arguments across the call
  stack and throws". **The worker's encoder never got that treatment.**
- `src/engine.worker.ts:100-108` — `base64ToBytesBounded`; throws on a transport-limit breach, and that
  throw lands in the same `catch`. A response-side breach is currently indistinguishable from a
  request-side fault.
- `src/engine-protocol.ts:5-15` — the bounds: `MAX_ENGINE_PAYLOAD_BYTES = 8 * 1024 * 1024`,
  `MAX_ENGINE_RENDER_PDF_BYTES`, `MAX_ENGINE_FONT_FAMILIES`.
- `src/engine-worker-admission.ts` — `admit`; refusals here are `fatal` or per-request and do **not**
  reach `execute`.
- `src/App.tsx:608-627` — `pickCatalogueFamily`: fetch the face, then one command. The fetch arm has
  its own refusal text, so a failure reported as *"invalid response"* is past it.
- `src/font-chain-command.ts:69-98` — `embedFontFamilyCommand` and its chunked base64.

**Go (`folio-go/`)**
- `component_commands.go:2359-2410` — `embedFontFamily`. Cleared by measurement; read for context only.
- `internal/template/fontasset.go:197-206` — `DecodeFontForRender`; `:178-184` `decodeRecognisedFont`
  (`font/ttf`, `font/otf` only); `checkSfnt` below it. The face-size cap lives on this path.
  **`checkSfnt` is the gap:** its own comment bounds it — *"it establishes that the file is not lying
  about what it is, and nothing more"* — so it reads the version tag and the table directory and never
  asks whether `fvar` is present.
- `internal/fontset/fontset.go:226-239` — the `fvar` refusal and its message. **Read it before writing
  the command's.** `:785` carries the design note for why ingestion is where it lived until now.
- `tools/fontgen/instance_faces.py` — the derivation D-16.5(d) uses, and the header explaining why its
  output is committed rather than generated.

## Tasks & Acceptance

**Execution:**
- **Reproduce in a browser first.** `folio-designer/e2e/` is compile-checked and never executed
  (D-000.4), which is how this reached the owner. Drive a real browser over every catalogue family,
  record which fail and with what, and keep the harness — a spec that is run, or an explicit statement
  that it is not.
- **Instrument the boundary before changing it.** Capture the actual exception at
  `engine.worker.ts:80` — name, message and, where available, whether it came from `host.handle` or
  `JSON.parse` — and record the observation with command, commit, tree state and working directory
  (D-8.4j.8).
- **Fix the cause the evidence names.** If it is the encoder, chunk it the way
  `font-chain-command.ts:92-98` already does, and say in the code why the two encoders now agree. If it
  is elsewhere, the fix goes where the evidence points.
- `src/engine.worker.ts` — **the second defect, fixed regardless of the first.** Split the `catch` so a
  transport-bound breach, a `JSON.parse` failure and a `host.handle` throw are distinguishable, and
  carry the thrown value's message into the reported error, bounded like every other string on this
  protocol.
- `src/engine-protocol.test.ts` / a worker-level test — red-prove the new reporting: a forced throw
  must produce a message naming it, and the test must fail if the message is reverted to the bare
  string.
- Keep `notosanssc`'s located size refusal arriving as a located refusal; assert it.
- `folio-go/component_commands.go` — **add the `fvar` refusal to `embedFontFamily`** (D-16.6), beside
  the existing `DecodeFontForRender` call, refusing before anything is written to `t.doc.Assets`, and
  carrying the engine's existing remedy message rather than a new one.
- `folio-go/component_commands_test.go` — a variable face refused at the command, **red-proved by
  deleting the new guard**, not by falsifying a condition; and a static face still accepted, so the
  guard cannot be over-broad.
- `folio-go/internal/fontset/fontset_test.go` — assert the renderer's guard **still** fires, so a later
  reader cannot conclude the command's check made it redundant.

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
- Given a static face, when it is embedded, then nothing about today's behaviour changes.

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

**Why the second defect belongs here and not in 16.1.** 16.1 is about where a font comes from. This
is about what the boundary will accept once it has one — the same boundary, the same probe harness, and
a guard whose absence is a defect whether or not Epic 16 is ever built. Placing it in 16.1 would make
a shipped bug look like a consequence of the new feature, which is the opposite of true.

**What "some fonts" is worth as a signal.** The owner reported failure on *some* picks. The catalogue
spans 26 KB to 646 KB, so a size-dependent boundary fault is consistent with that report and a
content-dependent one is not — the Go layer accepted all 21 sets of bytes.

## Verification

- A browser run over all 21 families, recorded per family.
- `cd folio-go && go test ./...`
- `cd folio-designer && npm run test && npm run build`
- The 23 golden digests, unmoved.
- The D-16.6 probe re-run at implementation HEAD: `Anuphan[wght].ttf` refused at the command, and still
  refused at ingestion.
