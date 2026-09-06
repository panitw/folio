---
title: 'Story 13.1: The preview keeps the PDF'
type: 'feature'
created: '2026-09-06'
status: 'done'
review_loop_iteration: 0
baseline_commit: '9608b46c6989a0855f8643929d02fe64edb40b0d'
context: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative — written after delivery. The frozen Intent below governs implementation; this section
describes what actually shipped.*

The designer can now hand the author the exact PDF on screen. A Save PDF control sits beside the local
render control on the preview screen. Pressing it writes the same bytes the engine produced — nothing is
re-rendered and nothing is re-encoded on the way out — and the save goes through the file machinery the
app already used for templates, taught to carry a file format rather than assume every save is a
template. With no render yet the control is present but disabled and says why; when the preview is out of
date it says so both before the press and in the completion message.

Two smaller changes to template file naming shipped as a deliberate consequence, and they were ruled
opposite ways on purpose. A title that already ends in a document extension no longer collects a second
one: the suggested name swaps the extension instead of stacking it, in both directions. But the
capitalisation the author typed is never rewritten, because changing case can fork one document into two
files a keystroke apart on some filesystems. The first is a suggestion anyone can overtype; the second is
silent divergence in the author's own file.

The browser-level test written here was type-checked but not executed locally; continuous integration now
runs it on every push. Six observations found along the way were recorded as deferred work rather than
fixed, including one about a hash the screen shows that nothing verifies.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The designer renders the exact production PDF but cannot let it leave the tab. There is no
export path anywhere in `folio-designer/src` — the only `download=` in the tree is the `.folio` anchor in
`file/input-download.ts` — so the only way to obtain the document on screen is to re-run the CLI.

**Approach:** Add a **Save PDF** control that writes `preview.bytes` — the `ArrayBuffer` the engine
returned, which is the buffer the displayed digest covers — through the existing two-tier `FileAccess`
interface, by parameterising that interface's `.folio` hardcodings rather than writing a second download
path.

## Boundaries & Constraints

**Always:**
- Write `preview.bytes` verbatim. Never re-render, never re-serialize, never decode to text and back.
- Route through the injected `FileAccess`. Capability detection stays in `file/capability.ts` alone —
  `file-access-contract.test.ts`'s `extraCapabilityChecks` fails any other file naming `showSaveFilePicker`.
- A PDF save is always a fresh target: pass `saveAs: true` with no `currentTarget`, and never write
  `title`, `target` or `savedRevision` from its result. The retained `.folio` handle must never be written
  with PDF bytes.
- Carry the format through **all three** `.folio` hardcodings — `folioName` (`file-access.ts:67`),
  `folioPickerType` (`file-system-access.ts:12`), and the blob MIME (`input-download.ts:49`). Three, not
  the two D-000.8 counted: a spec naming only the MIME ships a PDF saved as `.folio`, and a count wrong by
  one is how the third site stays hardcoded.
- Never pass `currentTarget`, and always pass `saveAs: true`. `FileSystemAccess.acquireSaveTarget` reuses a
  retained handle without a picker when `!saveAs && currentTarget?.kind === 'in-place'` — a PDF save that
  passed the template's target would **overwrite the author's `.folio` with PDF bytes**, irreversibly.
- Add nothing to `preview/pdf-viewer.tsx`. `preview-authority-contract.test.ts` forbids `createObjectURL`
  and `revokeObjectURL` there by name, and a new prop would join the render effect's dependency array,
  which is DW-191's live tear-down.
- Every new control is keyboard-reachable, labelled, and states its own disabled reason (UX-DR25).

**Ask First:**
- Any change to `preview/pdf-viewer.tsx`, to its render effect, or to `PDFPreviewViewState` — DW-191 is
  Story 13.2's, and fixing it here is an intent gap, not a courtesy.
- Any change to observable `.folio` open/save behaviour: its suggested name, picker type, blob MIME, or
  status wording.
- Any new dependency, any file not named under **Execution**, any adjacent defect discovered.
- Any responsive `@media` rule — `canvas-authority-contract.test.ts` asserts the media-query list equals
  exactly `['prefers-reduced-motion: reduce']`.

**Never:** commit, `git add`, stash, checkout, reset, revert, restore, push, create a branch, or open a
pull request — the human makes every commit in this run. Never build the evidence rail, the thumbnail
rail, the Re-render relocation, viewer navigation, or the no-data preview (13.2–13.5). Never touch Go,
the engine, the wasm boundary, or any golden byte.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Current preview, native tier | `admittedPreview(preview)`; both pickers present | picker offered `<base>.pdf` and a PDF-only accept entry; the handle is written exactly once with a buffer equal to `preview.bytes`; status names the saved file | N/A |
| Current preview, download tier | same state, no pickers | anchor `download` is `<base>.pdf`, blob type is `application/pdf`, blob bytes equal `preview.bytes` | N/A |
| Stale preview | `preview` present but `admittedPreview` false | it still saves, and says stale in both directions: the control names the save as stale before the press, and the completion status names the saved revision as stale | N/A |
| No render yet | `preview === undefined` | the control renders `disabled` with its reason in accessible text, never a control that fails when pressed | N/A |
| Author cancels the picker | `showSaveFilePicker` throws `AbortError` | surfaces as `FileAccessCancelled`; status cleared; no `role="alert"`; no state changed | silent, as `.folio` cancel already is |
| Write fails | `createWritable`/`write` throws | one `role="alert"` naming the PDF save; `preview`, `title`, `target`, `savedRevision` all unchanged | announced |
| A `.folio` target is held | `target` set from an earlier template save | the picker IS shown; the held `.folio` handle receives no write | N/A |
| Double press | a PDF save already in flight | the second press is a no-op; exactly one picker call | N/A |

</frozen-after-approval>

## Code Map

Anchors re-derived by symbol at `9608b46` (D-000.4); the 2026-09-05 survey's line numbers were SAMPLED.

- `folio-designer/src/file/file-access.ts` — the byte-opaque interface. `FileAccess` (:49), `SaveRequest`
  (:31), `SaveTargetRequest` (:35), `AcquiredSaveTarget` (:44), `FileAccessCancelled` (:55),
  `FileAccessFailure` (:59), `isFileAccessCancelled` (:63), and **`folioName` (:67)** — hardcoding #1, the
  `.folio` suffixer, imported by both tiers.
- `folio-designer/src/file/file-system-access.ts` — tier 1. **`folioPickerType` (:12)** — hardcoding #2,
  shared by `open()` (:21) and `acquireSaveTarget()` (:35). `writeSave` (:42) is format-blind.
  `localFailure` (:56) already normalises `AbortError` → `FileAccessCancelled`.
- `folio-designer/src/file/input-download.ts` — tier 2. `acquireSaveTarget` (:39) calls `folioName`;
  `writeSave` (:45) has **hardcoding #3** at `new Blob([request.bytes], { type: 'application/json' })`
  (:49). Revoke is deferred one microtask (:62) so the click can start.
- `folio-designer/src/file/capability.ts` — `selectFileAccess` (:19), the only permitted capability probe.
  Do not add a fourth selector; the format is a parameter, not a new tier.
- `folio-designer/src/App.tsx` — `PreviewRecord` (:140, `bytes: ArrayBuffer`); `preview` state (:203);
  `installPreview` (:524) is the single writer and already stores `result.bytes.slice(0)` (:590);
  `admittedPreview` (:1771) is the ready-made "is this the current, engine-authoritative preview" guard;
  `save` (:1703) is the two-phase acquire-then-write pattern to mirror; `announceFailure` (:1651);
  `fileBusy` (:159), `title` (:160), `target` (:161); the file-action buttons (:1897); the preview region
  (:1923); the **PREVIEW INPUTS** panel with `Render local PDF` (:1926).
- `folio-designer/src/file/file-access.test.ts` — reusable fakes: `handle()` (:10) records
  `['picker','write:…','close']` ordering; the injected `{createObjectURL, revokeObjectURL}` and anchor
  doubles (:102–106); the `AbortError` idiom (:70).
- `folio-designer/src/App.test.tsx` — `vi.mock('./preview/pdf-viewer', …)` (:126) replaces the viewer with
  an `Admit local PDF` button, which is what flips status to `current`; the inline three-`vi.fn()`
  `FileAccess` literal (e.g. :993); the render-arm engine stub (:673–701).
- `folio-designer/src/file/file-access-contract.test.ts` — `extraCapabilityChecks` (:73),
  `opaqueByteRewrites` (:77, bans `new Blob([` followed by a string literal or `text`/`decoded`/`content`),
  `fileNetworkCalls` (:81). Read-only for this story.
- `folio-designer/src/preview/preview-authority-contract.test.ts` — 13 forbidden tokens in
  `pdf-viewer.tsx` (:7) and six pinned `App.tsx` literals (:17–22), including the freshness `<p>`'s exact
  attribute spelling and order. Read-only; do not disturb the pinned literals.
- `folio-designer/e2e/local-file-actions.spec.ts` — the model for both tiers in a real browser: forcing
  tier 2 by nulling both pickers (:11), the fake handle stashing bytes on `window.__folioWrites` (:26–34),
  and the `AbortError` cancel case (:44–59). Compiled every story; executed at the epic-boundary gate, and
  soon in CI (DW-268).

## Tasks & Acceptance

**Execution:**
- [x] `folio-designer/src/file/file-access.ts` -- add an exported `LocalFileFormat`
  (`{ description; mimeType; extension }`) with `folioFileFormat` and `pdfFileFormat` constants; replace
  `folioName` with `localFileName(name, format)` that strips a trailing extension belonging to either
  known format before appending `format.extension`; add a required `format` field to `SaveTargetRequest`
  and to `AcquiredSaveTarget` -- one module owns all three hardcodings, and `AcquiredSaveTarget` carrying
  it makes it structurally impossible for the write to disagree with the picker.
- [x] `folio-designer/src/file/file-system-access.ts` -- build the picker `types` entry from
  `request.format` and the suggested name from `localFileName`; return the format on the acquired target.
  `open()` keeps `folioFileFormat` unchanged -- opening is still `.folio`-only.
- [x] `folio-designer/src/file/input-download.ts` -- take the download name and the blob `type` from the
  acquired target's format. `open()`'s `input.accept` is unchanged.
- [x] `folio-designer/src/App.tsx` -- add `exportPreviewPdf()` mirroring `save()`'s acquire-inside-the-
  gesture shape, guarded by a new `exportInFlight` ref and setting `fileBusy`; it passes
  `{ suggestedName: title, saveAs: true, format: pdfFileFormat }`, writes `preview.bytes`, and sets only
  `fileStatus`/`fileError`. Render a **Save PDF** button beside `Render local PDF` in the PREVIEW INPUTS
  panel (:1926), disabled with a stated reason when `preview` is undefined, and naming the save as stale
  when `admittedPreview(preview)` is false.
- [x] `folio-designer/src/file/file-access.test.ts` -- extend both tiers over the PDF format: picker
  options, suffixing (including a `.folio`-named title becoming `.pdf`), blob MIME, byte equality, and
  write/close ordering. Mutation-prove each new assertion by deleting the line it guards.
- [x] `folio-designer/src/App.test.tsx` -- integration over the matrix: enabled/disabled states and their
  reasons, the stale telling in both directions, cancel silence, failure alert, the held-`.folio`-target
  case (picker shown, handle unwritten), double-press, and byte equality between the written buffer and
  the engine stub's bytes.
- [x] `folio-designer/e2e/pdf-export.spec.ts` -- a browser witness of both tiers, modelled on
  `local-file-actions.spec.ts` and carrying a header saying it is compile-checked here. Write it as a real
  test, not a placeholder: CI is being changed under DW-268 to run the full Playwright suite on every push,
  so it will likely execute before the epic-boundary gate. It must use no identifier
  `canvas-authority-contract.test.ts` forbids -- every `e2e/` file is auto-enrolled by directory walk.

**Acceptance Criteria:**
- Given a preview rendered from an engine stub that returned a **literal** byte fixture, when Save PDF is
  pressed in either tier, then the bytes captured at the write seam equal that literal element-for-element,
  and their SHA-256 equals the digest the stub reported. Both sides are pinned to the fixture, never read
  back out of `preview.bytes` — an assertion whose two sides come from the same place is not an assertion
  (D-11.2.8). The fixture must be long enough and non-uniform enough that a truncation or a re-encode
  changes it.
- Given a preview is on screen, when Save PDF is pressed, then the engine receives no request at all — the
  test counts requests across the press and the count does not move. Nothing is re-rendered or
  re-serialized.
- Given a template previously saved to a `.folio` handle, when Save PDF is pressed, then
  `showSaveFilePicker` is called and that handle's `createWritable` is never called. **Mutation-prove this
  one specifically**: change the call to pass `currentTarget`, confirm the criterion reds, restore by `cp`
  and report the digest. A guard against destroying the author's file must be shown to fire.
- Given a save completes, when the designer state is inspected, then `title`, `target` and `savedRevision`
  are exactly what they were before the press.
- Given `preview === undefined`, when Preview is shown, then the Save PDF control is present, `disabled`,
  and its reason is in text a screen reader reaches.
- Given every guard added or changed by this story, when the line it guards is deleted, then that guard
  fails — and the source is restored by `cp` from a snapshot, with the restore digest reported.

## Spec Change Log

- **Plan gate, 2026-09-06 (not a review loopback).** D-000.8 records that "13.1 has two hardcodings, not
  one." The count is **three**: `folioName` (`file-access.ts:67`), `folioPickerType`
  (`file-system-access.ts:12`), and the blob MIME (`input-download.ts:49`). Re-derived by symbol at
  `9608b46`. The engineering lead ruled that this supersedes D-000.8's figure. Known-bad state avoided: a
  count wrong by one leaves the third site hardcoded and ships a PDF written with the wrong MIME or the
  wrong suffix. **KEEP on any re-derivation: all three sites named individually, never as "the `.folio`
  hardcodings".**

- **Step-04 triage, 2026-09-06 (patch, not a loopback).** This spec contradicted itself: the **Execution**
  clause mandated stripping "a trailing extension belonging to either known format", while **Ask First**
  forbade "any change to observable `.folio` open/save behaviour: its suggested name". Stripping *is* such
  a change — `folioName('statement.pdf')` gave `statement.pdf.folio`, `localFileName` gives
  `statement.folio`. **The Task clause was the one honoured**, ruled by the engineering lead: one symmetric
  rule in both directions, because `statement.pdf.folio` was an artifact of appending without stripping
  and the asymmetric variant needs a comment that will eventually be simplified away. Known-bad state
  avoided: an asymmetric suffixer that re-introduces the double-extension bug from the other side.
  **KEEP on any re-derivation:** the identity case `localFileName('report.folio', folioFileFormat) ===
  'report.folio'` must stay a no-op, and the reason for symmetric stripping must live in a comment at the
  strip site, not only here.

## Design Notes

**Plan-gate rulings, carried verbatim (engineering lead, 2026-09-06).**

- Q1 — telling the author a preview is stale: *"(b). … 'In the same breath' means the control and the
  completion status say it, not a live region that may have spoken minutes ago. And (c) is ruled out on a
  hard fact rather than taste: `file/file-access.ts:41-43` records that Chromium's picker is gated on the
  click's transient user activation, and a confirmation step spends it. A safety prompt that breaks the
  save is not a safety prompt."*
- Q2 — placement: *"(a) — beside `Render local PDF`. 13.3 then relocates one row instead of hunting two
  controls."*
- Q2's architectural fork — *"follow the AC — and AD-20 settles it, not precedent. Its rule ends: 'One
  capability check at startup selects the tier; the rest of the app talks to one file-access interface and
  never branches again.' Parameterising the picker's type and suggested name IS that rule; a fourth narrow
  interface is the branch it forbids."*
- Q3 — no runtime digest check: *"(a), no runtime check. AC1's obligation is a property of the PATH —
  never re-rendered, never re-serialized — and a test proves a path. Adding `crypto.subtle.digest` inside
  an activation-gated gesture is new runtime behaviour no AC asks for, in the one place where an extra
  async hop is most expensive."*
- Q4 — *"(a), and your honesty about it is the condition. Add the compile-only witness … and never
  reported as executed coverage."* Plus: CI is being changed now to run the full Playwright suite on every
  push (DW-268), so write it as a real test.

**The `image-file.ts` precedent, checked rather than relayed — and it argues the same way.** `grep -an
"AD-20"` over `folio-designer/src` returns exactly two hits (positive control: `AD-17` returns five):
`file/capability.ts:35` and **`image-file.ts:5`**. So `image-file.ts` *does* cite AD-20 — but for *"a
narrow, read-bytes-only interface with **no save**, no handle retention and no document semantics."* It is
a precedent for a narrow **open** path, not for a second **save** path, and this story adds a save. The
precedent and AD-20 therefore agree with the AC; there is no outlier to reconcile.

**Why `preview.bytes` is already exact.** The chain is `folio-go/wasm/engine.go` `Engine.Render` (the
digest is taken over that same slice) → base64 over the wasm reply → `engine.worker.ts` decode →
structured-clone *transfer* → `engine-client.ts` `copyBytes` → `App.tsx:590` `result.bytes.slice(0)`.
Every hop is value-preserving. `pdf-viewer.tsx:47` hands pdf.js `new Uint8Array(bytes.slice(0))`, a private
copy, so the rasterizer never neuters the buffer the export reads — **that `.slice(0)` is load-bearing for
this story and must not be removed.**

**Why the freshness predicate is `admittedPreview`, not `previewStatus === 'current'`.** `'current'` in
this codebase additionally means *pdf.js admitted the bytes* (`App.tsx:597`, `:701`). A viewer failure
leaves valid, byte-exact bytes under a `'stale'` status — savable, and correctly told as stale. Using
`admittedPreview` reuses the existing predicate rather than minting a second definition of freshness that
could drift from the status line's.

**The digest is a claim, not a check.** Nothing in the browser verifies `preview.digest` against
`preview.bytes` today — it is computed in Go (`wasm/engine.go` `Engine.Render`), carried on the wire, and
shape-checked only (`engine-protocol.ts` `isPreview`, `/^[a-f0-9]{64}$/`). Ruled Q3(a): this story proves
the equality **in test** and adds no runtime crypto. The standing gap is being registered separately in
`deferred-work.md` by the orchestrator; it is not this story's to close.

**Format shape, for the implementer:**

```ts
export type LocalFileFormat = Readonly<{ description: string; mimeType: string; extension: string }>
export const folioFileFormat: LocalFileFormat = { description: 'Folio template', mimeType: 'application/json', extension: '.folio' }
export const pdfFileFormat: LocalFileFormat = { description: 'PDF document', mimeType: 'application/pdf', extension: '.pdf' }
```

**Placement is provisional and 13.3 relocates it.** The design pairs Save PDF with Re-render at the foot of
an evidence rail that does not exist yet. Pairing it with `Render local PDF` where that control lives today
keeps the two together, so 13.3 moves one row rather than hunting two controls.

## Verification

Heavy-test cadence for this run: unit + lint + build every story; the full Playwright suite is an
epic-boundary gate (D-000.30) and is **not** executed here — this story's behaviour is fully observable in
jsdom, and its browser witness is compile-checked only at this gate. Report it as compile-checked, never as
executed coverage. (CI is being changed under DW-268 to run Playwright on every push, so the witness may
execute there sooner than the epic gate — that is a reason to write it properly, not a reason to claim it
ran here.)

**Commands** (all paths absolute; never `cd` before a relative restore path):

- `cd /Users/panitw/Projects/folio/folio-designer && npm test` -- expected: **65+ files, ≥976 tests, 0
  failing**. Report the new totals and name every added test file.
- `cd /Users/panitw/Projects/folio/folio-designer && npx tsc -b --force` -- expected: exit 0.
- `cd /Users/panitw/Projects/folio/folio-designer && npx oxlint` -- expected: **exactly 4**
  `only-export-components` warnings, 0 errors. Baseline sites at `9608b46`: `preview/pdf-viewer.tsx:16,17`
  and `App.tsx:3585,3592`. The count and the rule are the invariant; the lines are not.
- `cd /Users/panitw/Projects/folio/folio-designer && npm run test:e2e:compile` -- expected: exit 0.
- `cd /Users/panitw/Projects/folio/folio-go && go test -count=1 ./...` -- expected: **2242 pass / 2 fail /
  5 skip**, failing ONLY `TestCorpusMeetsP6ExerciseFloors` and `P6g_(opaque_names)`, checked by enumerated
  name and never by exit code. A third distinct failure is a hard stop. (Measured at `9608b46`; this story
  touches no Go, so any movement is a signal.)
- `gofmt -l /Users/panitw/Projects/folio/folio-go /Users/panitw/Projects/folio/lint` -- expected: empty.
  Absolute paths only; a persisted `cd` makes this emit phantom `lstat` lines.
- Do **NOT** run `npm run build`.

**Manual checks:**
- `git status --porcelain` after every mutation restore, before the next probe.
- Confirm no file outside the **Execution** list is modified, and that `.github/workflows/*.yml` and
  `deferred-work.md` are untouched — another agent is editing those concurrently.

## Suggested Review Order

**The format parameter — the one idea the whole change turns on**

- The three `.folio` hardcodings become one value carried to both tiers.
  [`file-access.ts:81`](../../folio-designer/src/file/file-access.ts#L81)

- Naming: case-folded recognition, symmetric strip, casing never rewritten — read the comment first.
  [`file-access.ts:128`](../../folio-designer/src/file/file-access.ts#L128)

- The picker entry is derived from the format, so filter and suffix cannot disagree.
  [`file-system-access.ts:14`](../../folio-designer/src/file/file-system-access.ts#L14)

- The blob MIME reads off the acquired target, never a literal.
  [`input-download.ts:51`](../../folio-designer/src/file/input-download.ts#L51)

**The export itself**

- `saveAs: true`, no `currentTarget`, nothing kept from the result — the safety comment states why.
  [`App.tsx:1824`](../../folio-designer/src/App.tsx#L1824)

- One label drives the control and its disabled reason, so they cannot name different controls.
  [`App.tsx:1815`](../../folio-designer/src/App.tsx#L1815)

- A second in-flight latch, deliberately not `saveInFlight`; each path refuses while the other holds.
  [`App.tsx:239`](../../folio-designer/src/App.tsx#L239)

**Tests**

- Nine behaviours over the real tiers: byte-verbatim writes, stale telling, latch release, tab-switch alerts.
  [`App.test.tsx:6368`](../../folio-designer/src/App.test.tsx#L6368)

- The casing pin restored beside the PDF case — separate properties, neither substitutes for the other.
  [`file-access.test.ts:167`](../../folio-designer/src/file/file-access.test.ts#L167)

- Browser witness for both tiers; compile-checked locally, executed in CI since `adf905a`.
  [`pdf-export.spec.ts:1`](../../folio-designer/e2e/pdf-export.spec.ts#L1)

## Delivery Log

### 2026-09-07 — done

Baseline `9608b46`. Shipped in one commit, `1091645` — 11 files, +1412/−28, no loopback
(`review_loop_iteration: 0`). Save PDF writes `preview.bytes` verbatim through the existing two-tier
`FileAccess`, with the three `.folio` hardcodings collapsed into one `LocalFileFormat` carried from picker
to write. The plan gate's correction of D-000.8 — three sites, not two — held: had the count stayed at
two, the third site would have shipped a PDF written with the wrong suffix or the wrong MIME, which is
exactly the failure a story named for removing those hardcodings would have been assumed to have closed.

**Two `.folio` behaviour changes shipped, ruled opposite ways deliberately — record the asymmetry, do not
"simplify" it away.** Stripping is now **symmetric**: a template titled `statement.pdf` suggests
`statement.folio`, not `statement.pdf.folio`. Casing is **preserved**: `REPORT.FOLIO` stays itself. Both
are changes to observable `.folio` naming, both sat behind the same Ask-First clause, and they were ruled
in opposite directions on **blast radius**, not on taste. A wrong suggested name is a string the author
overtypes in the picker before anything happens. A case-folded name **forks one document into two files a
keystroke apart on a case-sensitive filesystem**, with nothing to say which one the next Open picks up —
and on a case-insensitive one it silently overwrites. The two rules look inconsistent side by side; they
are not, and the reason lives in a comment at the strip site as well as here, because a later reader
tidying "one of these must be wrong" would reintroduce the worse of the two.

**Both Ask-First gates fired on the same function**, which is the builder's own observation and the more
useful finding. `localFileName` was mandated a change by a Task clause and forbidden that same change by
an Ask-First clause — the spec contradicted itself, and the contradiction was only visible at the moment
of implementing it. Ruled at step-04 triage: the Task clause wins for stripping; the Ask-First clause wins
for casing. A fence that trips twice in one place was drawn wrong, not obeyed wrongly. Recorded in the
Spec Change Log above and in D-13.1.2 / the step-04 triage entry.

**The `exportInFlight` latch was invisible to 286 passing tests.** Deleting its reset left the entire suite
green, because every test pressed Save PDF exactly once. A leaked latch would have made the control work
**once per session**, silently, with the button still enabled and no error anywhere — the worst shape of
defect this run has produced, because the failure mode is a control that stops responding and never says
so. Now red-proved by pressing twice and asserting two picker calls and two writes, on each of the
completed, cancelled and failed paths that share the `finally`.

**A fix whose correctness no test can distinguish from the bug is not a fix.** The builder's own patch-1
diagnosis for the naming defect named the wrong line, and the fix it proposed would have been caught by no
assertion in the suite — it would have passed review as a fix and shipped as a no-op. The separating input
had to be added before the patch meant anything. Treat "the patch is in and the suite is green" as
evidence of nothing until an input exists on which the two implementations disagree.

**A later fix can silently retire an earlier fix's only separating input** — DW-275, and the rule is worth
more than the entry. The single-longest-match strip was red-proved against a chained fold earlier in this
same story. Then the case-preserving early return, ruled at the second Ask-First gate and entirely
correct, returned the one input on which the two implementations disagreed unchanged, short-circuiting the
comparison before either strip ran. Nothing regressed; the property simply stopped being observable, and
the recorded proof became a claim about a tree that no longer exists. **So: re-run a function's mutation
proofs after every subsequent change to that function, not only after the change that introduced them.**

**A test pin moved to a neighbouring case, and the baseline it protected stopped being checked while still
reading as a guard.** Caught and restored — the casing pin now sits beside the PDF case as its own
assertion, because they are separate properties and neither substitutes for the other. Same family as
Epic 11's vacuous set assertions: the assertion is present, the file is green, and the thing it was
written to defend is no longer among the things it can fail on.

**Also fixed by executing rather than by reading:** `localFileName` returned `report.folio.folio` for one
input and a stem-less `.pdf` for another (the untitled fallback ran before the strip, not after), and its
extension match assumed lowercase constants. The `fileError`/`fileStatus` pair moved out of the hidden
tabpanel, where switching to the DATA tab mid-save dropped `role="alert"` from the accessibility tree — an
alert that tests as present and behaves as absent. And a test cast erased the very type the story had just
made required.

**Findings triaged: no tally was recorded, and I am not inventing one.** The population I searched is the
whole of this spec file plus `1091645`'s commit message; this story carries no `## Implementation Notes`
and no triage section, and neither source states a findings count or names a single rejection. What is
countable is the deferred arm: **6 deferrals**, each a census-visible `### DW-` heading whose
`source_spec` names this story. DW-270 — the displayed digest is never checked against the bytes it
claims to cover (pre-existing; natural home Story 13.3, which promotes that hash to an evidence block).
DW-271 — two of `exportPreviewPdf`'s three pre-await captures are defensive reads nothing can
discriminate; **the byte capture IS guarded and reds**, and the entry exists so the defending comment is
never mistaken for a proof. DW-272 and DW-273 — two status-announcement gaps, the second being the half
of a pre-existing silence this story did not close. DW-274 — AD-20 says "one file-access interface" and
there are three, all read-only, so the rule's purpose is intact and only its literal has drifted.
DW-275 — above. All six are LOW / OPEN / unassigned. The patched arm is named rather than counted, in the
paragraphs above and in the commit message; **the rejected arm has no record at all**, so read "0
rejected" as unverified rather than measured.

The orchestrator placed DW-270 … DW-275 and I verified rather than re-added them: all six are proper
`### DW-` headings, and the register is internally consistent at **278** headings with **max DW-275** —
278 = 275 numbers + the three known pre-existing duplicate headings (DW-100, DW-162, DW-238), which are
left exactly as found. Per D-000.31a the `- source_spec:` lines are provenance metadata on numbered
entries, not orphan markers; no unnumbered block is lodged inside any entry, and the file ends inside
DW-275 with nothing trailing it.

**Gates, re-measured at `1091645` by the closer, not carried forward:** `npm test` **65 files / 993 tests,
0 failing** (baseline 976). The vitest **file** count did not move: the story added **no new vitest file**,
and all +17 tests landed as additions to the two existing files `src/App.test.tsx` and
`src/file/file-access.test.ts`. The one test file this story *added* is `e2e/pdf-export.spec.ts`, which
Playwright owns and vitest never loads — so it is not one of the 65 and contributes none of the 993.

The rest, in order: `tsc -b --force` exit 0; `oxlint` exit 0 with
**exactly 4** `only-export-components` warnings, 0 errors, all pre-existing (`preview/pdf-viewer.tsx:16,17`
and `App.tsx:3665,3672` — the spec's baseline named `App.tsx:3585,3592` at `9608b46`; the count and the
rule are the invariant, the lines are not); `npm run test:e2e:compile` exit 0; `go test -count=1 ./...`
**2242 pass / 2 fail / 5 skip**, counted from `-json` events, failing only `TestCorpusMeetsP6ExerciseFloors`
and its `P6g_(opaque_names)` subtest by enumerated name, as required by D-000.17/D-2.1.14 and unchanged by
this story, which touches no Go; `gofmt -l` over `folio-go` and `lint` with absolute paths, empty.
`npm run build` was **not** run, per the spec.

**Not executed here: the full Playwright suite.** The heavy-test cadence places it at the epic boundary,
and Epic 13 has four stories left before that gate. `e2e/pdf-export.spec.ts` is **compile-checked only**
locally and must never be reported as executed coverage from this run — but per D-13.1.4 it does execute
in CI on every push and pull request since `adf905a` (D-000.4's epic-boundary cadence superseded on that
point by the owner's D-11.6.1), so the first push carries its first real execution, and the first Linux
evidence for the designer e2e job at all.

