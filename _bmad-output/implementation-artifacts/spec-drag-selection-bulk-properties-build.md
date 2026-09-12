---
title: 'Rectangle selection, group movement, and common property editing'
type: 'feature'
created: '2026-09-12'
status: 'done'
baseline_commit: 'c74df9b7a7444647850f43fa8e1b3a4790e16207'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-drag-selection-bulk-properties/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-drag-selection-bulk-properties/interaction-contract.md'
  - '{project-root}/_bmad-output/specs/spec-drag-selection-bulk-properties/group-movement.md'
  - '{project-root}/_bmad-output/specs/spec-drag-selection-bulk-properties/brownfield.md'
---

<frozen-after-approval reason="User explicitly requested Build this spec after its scope was reviewed">

## Intent

**Problem:** Authors cannot rectangle-select elements, drag a selection together, or consistently edit its common properties.

**Approach:** Implement CAP-1–CAP-5 and every AC/G case in the canonical spec and three companions listed in context. Those files carry the complete approved interaction contract; do not narrow the scope.

## Boundaries & Constraints

**Always:** Preserve relative group offsets, authored values, unique IDs, one-command atomicity and undo. Use engine geometry and the existing input-coordinate seam; no browser layout measurement. Preserve prior line-resize/snapping fixes. Use existing theme tokens. Primary input is desktop pointer; test actual Chromium gestures.

**Ask First:** A product behavior change incompatible with the canonical spec. Routine additive command/query/projection changes needed to implement that spec are authorized.

**Never:** Move elements between bands, scale groups, introduce persistent groups, touch unrelated documentation, alter golden fixtures, weaken authority scans, or commit/stash/reset/push the existing mixed worktree. The prior line changes and unrelated rendering-documentation spec predate this build.

## I/O & Edge-Case Matrix

The normative matrix is `interaction-contract.md` AC-1–AC-14 and `group-movement.md` G-1–G-10. Cover every row with passing tests; include real browser gestures and engine history rather than relying on static fake snapshots.

</frozen-after-approval>

## Code Map

- `folio-designer/src/App.tsx`: `select` (~1397), `sheetSurface` (~2700), `CanvasComponent` (~4976), `ComponentEcho`, `clearInteraction`; current pointerdown collapses selection and drag state holds one ID. Own gesture lifecycle/cancellation and per-ID preview without breaking placement/resize/table editing.
- `folio-designer/src/sheet-stack.ts`: `sheetPitch`, `columnForStackY`, projected window origins/home/echo/cap. New pure selection geometry can live alongside this. Use an always-present positioned stack plus known-width canvas-body/gutters for deterministic local input origins; blank band/page/gap starts reuse `placementPoint`, then only client deltas. Do not infer outer centered canvas origin from unrelated local offsets.
- `App.tsx` `placementPoint` (~4961) is the approved native local coordinate seam; `canvasDisplay.documentDelta` converts deltas. `canvas-authority-contract.test.ts` prohibits browser metrics throughout source and tests. Keep the ban intact.
- `App.tsx` `ComponentProperties`, `PropertyDraft`, specialized controls: existing shared values/mixed logic and selection-key remounts; fix common-field filtering, untouched default blur commits, mixed swatches/booleans and stale drafts. Keep single-selection behavior. Group preview disables competing geometry edits.
- `folio-go/component_commands.go`: `ApplyComponentCommand`, `moveComponent`, `updateComponentProperties`, `findComponent`, `projectedSize`, `containComponent`. Add a pure shared-delta solver plus atomic `moveComponents` command (`ids`, `referenceId`, `dx`, `dy`, `snap`), validating all IDs and derived safe lengths. Intersect member movement ranges; snap reference within feasible grid points, precise fallback only if none fits.
- `folio-go/wasm/engine.go`: `Engine.Apply` already owns clone/byte-equality/no-op/undo/redo. Add read-only group preview using the same solver, returning accepted delta and revision. Fence final command against captured revision. Public command must remain atomic even outside host.
- `folio-go/wasm/cmd/engine/main.go`, `folio-designer/src/engine-protocol.ts`, `engine-client.ts`, `engine.worker.ts`: mirror `table-columns` query plumbing for group preview; coalesce pointer queries (one in flight), reject stale gesture/sequence/revision replies, await final requested preview on release. No TypeScript copy of engine snapping policy.
- `folio-designer/src/component-command.ts`, `command-json.ts`: centrally encode additive group command/query payload; no per-ID command loop or identical absolute X/Y assignment.
- `folio-go/page_setup.go`: `CanvasComponent`, `canvasComponents`, `applyCanvasStyle` collapse null/absent and omit nonpainting border values. Add bounded typed authored inspector evidence separately from paint fields, preserving absent/null/value and explicit false/zero/empty values. Update protocol guards; a states-only map cannot recover hidden authored border values.
- Existing tests: `App.test.tsx`, `line-rect-vocabulary.test.tsx`, protocol/client/worker/encoder suites; Go `component_commands_test.go`, `wasm/engine_test.go`, `wasm/cmd/engine/main_test.go`; browser fixtures/config under `folio-designer/e2e/` and `playwright.config.ts`.

## Tasks & Acceptance

- [x] Implement pure rectangle/occurrence geometry, surface input mapping and selection gesture.
- [x] Implement engine group solver, atomic command, read-only preview and guarded transport.
- [x] Integrate group gesture, coalesced preview, final release, cancellation and selection retention.
- [x] Add authored property evidence, common-field intersections and inert mixed/default draft behavior.
- [x] Run unit, Go/host, type/lint and real browser tests covering all AC/G rows; record evidence.

Given the approved canonical spec, when each AC-1–AC-14 and G-1–G-10 scenario runs, then its stated result holds. Given the previous line fix, when a line is resized or moved, then thickness and Snap behavior remain correct. Given cancellation/refusal/stale work, when it completes, then no partial mutation or new-selection targeting occurs.

## Design Notes

User approval is the explicit 2026-09-12 instruction to build the reviewed spec. The working tree is intentionally dirty from the prior authorized fix; backup is `/var/folders/3c/6839rkd565d4dgdfpm5yxxx00000gn/T/folio-selection-baseline-39o9fjca`. Do not restore whole files from HEAD. Work sequentially; no further implementation delegation. Finish all approved behavior and tests; report concrete blockers rather than silently omitting cases.

Additional canvas findings: capture on a stable ancestor so moving a component between page homes does not lose pointer ownership. Recompute preview homes/echoes from moved geometry; preserve committed text-paint offsets rather than blindly replacing its origin. Content enclosure must cover the whole column interval across effective visible windows (respect seam masks and drawing cap), not just all listed occurrences. `commitComponent` lacks the generation guards of `applyProperties`; the group lifecycle must fence late document/revision responses.

## Verification

Run focused tests during development, then all designer unit tests; forced `tsc -b --force`, `oxlint`, `npm run test:e2e:compile`, focused Go command/host suites and relevant architecture guards, `npm run build:wasm`, and real Chromium selection/group/bulk integration tests. Build the production designer when required by browser config. Report any suites not run and pre-existing failures separately. Record exact commands/results and an AC/G evidence map here at completion.


### Implementation evidence — 2026-09-12

The initial implementation evidence below was collected before independent review; final reviewed evidence follows the triage section. Existing line-resize changes and unrelated rendering documentation were preserved. No commit, stash, reset, push, golden-fixture change, or authority-scan relaxation was performed.

Implemented projection-based full-box enclosure across sheet windows, rectangle replacement/Shift union, stable ancestor capture and cancellation, selected-body group movement (including table cells and repeated/continued occurrences), one-flight revision-fenced engine previews, and atomic shared-delta Go commits/history. The inspector now uses separate bounded authored absent/null/value evidence, intersects capabilities across kinds, preserves mixed states and explicit clear/null operations, and fences drafts/responses by selection and document generation. Single-selection inherited-default behavior is preserved; reviewed multi-selection Clear also reconciles to the accepted shared default. The visible grid now shares the engine's band-relative origin, including content-window phase.

Commands below ran from `folio-designer/` unless otherwise stated:

| Command | Executed result |
|---|---|
| `rtk proxy npx vitest run --reporter=json --outputFile=/tmp/folio-selection-unit-final.json` | **1,506 passed, 0 failed**; complete designer suite, including authority scans and prior line regressions. Log: `/tmp/folio-selection-unit-final.log`. |
| `rtk proxy npx tsc -b --force` | Passed after final source/test changes. |
| `rtk proxy npm run lint` | Passed, with the existing eight Fast Refresh export warnings and no errors. Log: `/tmp/folio-selection-lint-final2.log`. |
| `rtk proxy npm run test:e2e:compile` | Passed. |
| `rtk proxy npm run build` | Passed: both font scans, Go WASM compilation, TypeScript, Vite, offline generation and offline verification. Existing warnings: large JS chunks and 62 of 64 offline cache assets. Log: `/tmp/folio-selection-build-complete.log`. |
| `rtk proxy go test ./...` from `folio-go/` | All packages passed except the pre-existing `internal/text` corpus-floor failure: `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` has 7 entries, needs 20. Root component tests and `wasm` history tests passed. Log: `/tmp/folio-selection-go-all-final.log`. |
| `rtk proxy go test . ./wasm -count=1` from `folio-go/` | Passed both native component/command and host suites. Log: `/tmp/folio-selection-go.log`. |
| `rtk proxy env GOOS=js GOARCH=wasm go test -exec=/opt/homebrew/Cellar/go/1.26.0/libexec/lib/wasm/go_js_wasm_exec ./wasm/cmd/engine` from `folio-go/` | Passed real JavaScript/WASM host transport execution, including group preview/commit/refusal/no-op history. Log: `/tmp/folio-selection-jswasm-final.log`. |
| `rtk proxy env PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' npx playwright test -c .selection-playwright.config.ts e2e/selection-group-properties.spec.ts --reporter=list` | **11 passed in 21.1 s**, actual Chromium over the production assets above. Temporary runner config spread the existing Playwright config and changed only `webServer` to `npm run preview -- --host 127.0.0.1 --port 4173`, with the same URL and a 30 s startup timeout, avoiding another identical production build. Temporary config removed afterward. Log: `/tmp/folio-selection-browser-final.log`. The normal config remains the reproducible full-build entry point. |
| `rtk proxy git diff --check` from repository root | Passed. |

Visual evidence: `/tmp/folio-selection-grid.png` shows the selected group, Snap/Grid on, and mixed inspector at 150% with margin left 37 pt / top 35 pt and header 61 pt. The parent independently inspected it and found the band-relative grid and inspector visually correct.

### Acceptance coverage

All browser references below refer to the executed `folio-designer/e2e/selection-group-properties.spec.ts`, not merely compiled test titles.

| Acceptance | Passing behavior evidence |
|---|---|
| AC-1 | Browser four-direction rectangles at three zooms select A/B and exclude partial C; `canvas-selection.test.ts` includes exact edge equality. |
| AC-2 | Browser all-kind enclosure and table-body drag; pure geometry tests cover overlapping five kinds, full thickness, and text-ink-only exclusion. |
| AC-3 | Browser replacement, Shift union, empty and Shift-empty gestures; hook cancellation preserves selection without a command. |
| AC-4 | Browser rectangle crosses displayed pages and repeated header IDs once; pure occurrence tests reject partial split content, undrawn gaps and drawing-cap portions. |
| AC-5 | Browser 50/100/150% four-direction cases repeat membership after toggling Grid/Snap. |
| AC-6 | Browser group table-cell hit, unselected-body movement, line endpoint resize and placement through selected repeated echo; existing full-suite boundary/placement regressions also pass. |
| AC-7 | Browser mouse release and following native click retain membership and remove the overlay. |
| AC-8 | Browser mixed text/rect/line/table intersection plus full App/vocabulary suites verify text/table, line/rect, table/rect and removal of single-only sections. |
| AC-9 | `App.test.tsx` common-property scope tests and authored Go/protocol tests cover shared/mixed values, absent/null, false/zero/empty, boolean/segment/swatches, and explicit Clear of null. |
| AC-10 | Browser focus/blur of mixed and shared inherited-default fields, and draft returned to blank, leave revision unchanged; App tests also cover shared/defaulted drafts. |
| AC-11 | Browser bulk font-size set and Undo; existing Go property batch/host atomic history suites and App command-capture assertions pass. |
| AC-12 | Browser width valid for one target but invalid for another reports refusal without revision change; Go atomic property batch tests pass. |
| AC-13 | App deferred responses prove captured selection targets and refusal to install an old-document response; fresh selection drops stale blur/draft ownership. |
| AC-14 | Hook rectangle matrix covers Escape, lost capture, zoom, mode, geometry, document, and scroll invalidation with no selection commit; browser demonstrates page-space capture/release and scroll cancellation. |
| G-1 | Browser three-member group preserves selection on press/release/click, moves all previewed members equally, leaves the unselected rectangle unchanged. |
| G-2 | Browser mixed five-kind group retains dimensions/thickness and common delta; Go all-kind tests also preserve authored properties/content/bands. |
| G-3 | Go fractional millipoint Snap on/off reference tests; browser reference grid alignment at 150% and precise Snap-off movement. |
| G-4 | Browser group stops at common horizontal cap; Go intersected ranges test feasible grid choice, legal no-grid fallback, header limits and flowing content. |
| G-5 | Browser Escape and return-to-start preserve revision/positions; hook zero accepted delta and engine byte/history no-op checks pass. |
| G-6 | Browser one-revision move and one Undo/Redo; native/public command, host and js/WASM tests refuse invalid/missing/duplicate members atomically. |
| G-7 | Browser selected table cell moves group without column selection; unselected drag moves one; line endpoint resize preserves thickness. |
| G-8 | Browser repeated-header body drag and Shift toggle at 50%, plus a real flowing-text continuation pressed on page 2 and dragged across a seam; logical ID translation and Undo remain coherent. Pure geometry covers repeat/split occurrences. |
| G-9 | Hook deferred preview matrix exercises active and released/pending gestures under selection/document/geometry/zoom/mode/Escape/scroll invalidation; stale replies cannot commit/reselect. Expected automatic capture loss after release is separately retained while awaiting the final preview. |
| G-10 | Browser bulk X=42 and Y=56 assign identical absolute coordinates to both targets, each undone once; group drag retains differing original positions and uses one history step. |

No approved behavior is intentionally left incomplete. The full unrelated browser suite was not run; the focused new integration suite and full designer unit suite were run. The known Go corpus-floor failure remains outside this change. Independent review/triage is owned by the parent build after this implementation handoff.


### Independent review triage — 2026-09-12

The three review layers completed before triage. Blind and edge reviewers were fresh independent agents. The verification reviewer reused its earlier read-only engine-investigation session because the runtime refused another fresh agent slot; it had not implemented this feature. Duplicate findings with the same fix were merged. The implementation agent was no longer available with its context, so the parent applied the patches sequentially.

All accepted findings below are `patch`: local implementation or verification corrections under the already explicit approved contract; no product decision, frozen scope change, or authored vocabulary expansion is needed.

| Finding | Severity | Resolution |
|---|---|---|
| Last border checkbox sent a refused empty array | medium | Preserve the existing last-edge Clear operation; real-engine browser set/clear/undo coverage. |
| Clear of projected parent-null border state did nothing | medium | Clear Edges handles an internal null border consistently with the other border controls; native regression. The current file parser normalizes a parent border:null to an empty block, so no child-null Clear affordance appears for that loaded spelling; parser semantics are preserved. |
| Group preview clipped beyond the last displayed window | medium | Open affected windows when accepted home or continuation geometry extends into an undrawn interval or beyond the final window; clip earlier-page prefixes on opened echoes. |
| Continuous pointer input starved preview updates | medium | Paint monotonically newer accepted replies while coalescing and waiting for the final requested sequence on release. |
| Preview snapped back while commit was pending | medium | Retain accepted geometry, fence further canvas pointer starts until settlement, and preserve captured transaction lifecycle. |
| A pre-release preview refusal allowed the release click to clear selection | medium | Cancel with click suppression before reporting the refusal. |
| Outer and below-page blank canvas did not start rectangles | medium | Extend the stack's native hit target with a fixed pseudo-element clipped to the canvas region; no layout measurement or added scroll extent. Preserve backdrop click behavior. |
| Page traversal anchored to component top rather than pressed point | medium | Capture body-local pointer input through the approved seam; selected table cells belong to their component during multi-selection. Verify exact vertical displacement across a seam. |
| Shift-right-click toggled selected echoes | low | Apply the same primary-button admission guard as home components. |
| Group preview repeatedly projected the document for each ID | medium | One projection and member index per operation. A native 200-rectangle probe improved from about 30 ms to 0.34 ms; this is a local diagnostic, not a browser performance guarantee. |
| Loaded-image offsets and painted-grid phase lacked exercised coverage | medium | Add loaded-image preview/cancel checks and raster comparison with Grid off, including a nonzero continuation phase. No golden fixture or authority guard changes. |
| Mixed Clear reconciled through stale pre-commit consensus | medium | Resolve shared state and inherited defaults from the accepted projection. Restore the browser assertion to actual value 12 and prove untouched blur is inert. |

The existing authored property policy does not permit setting an empty border-edge array from this editor. The supported Clear behavior is preserved. Review fixes retain all passing geometry, atomicity, history, transport, line-resize, and property-scope behavior.


Review follow-up also checked the fixes themselves: gap handling now includes intermediate skipped intervals; Escape/scroll after dispatch keeps accepted geometry until command settlement. Context replacement still discards obsolete preview state. Hook coverage distinguishes pending preview from pending commit.


### Final reviewed verification — 2026-09-12

All accepted review findings are resolved. No change to the approved scope or authored-property vocabulary was required. Prior line length/thickness and Snap fixes remain included; unrelated rendering-library documentation is untouched. No commit, stash, reset, push, golden change, or authority-scan exception was made. Sprint sync was skipped because this spec has no story key.

| Check | Final evidence |
|---|---|
| Complete designer unit suite | **1,512 passed, zero failed**, `/tmp/folio-selection-reviewed-unit.json`. |
| Affected App, gesture, and authority checks after final preview fixes | **449 passed**, `/tmp/folio-selection-reviewed-recheck-final.json`. |
| Actual Chromium integration over completed production assets | **19 passed in 32.9 seconds**, `/tmp/folio-selection-reviewed-browser-complete.log`. Ran `rtk proxy env PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' npx playwright test -c .selection-playwright.config.ts e2e/selection-group-properties.spec.ts --reporter=list`; temporary preview-only runner config was removed afterward. |
| Production build | `rtk proxy npm run build` passed; `/tmp/folio-selection-reviewed-build-final.log`. Font-host scans, Go WASM build, TypeScript, Vite, offline generation and verification all completed. Existing chunk-size and cache-budget warnings remain. |
| Types and lint | Forced TypeScript passed; final E2E compilation passed (`/tmp/folio-selection-reviewed-e2ets-complete.log`); lint passed with the same eight existing Fast Refresh warnings (`/tmp/folio-selection-reviewed-lint-final.log`). |
| Native command and host suites | `rtk proxy go test . ./wasm -count=1` passed after engine patches; `/tmp/folio-selection-review-go2.log`. |
| Real JS/WASM host | `rtk proxy env GOOS=js GOARCH=wasm go test -exec=/opt/homebrew/Cellar/go/1.26.0/libexec/lib/wasm/go_js_wasm_exec ./wasm/cmd/engine` passed; `/tmp/folio-selection-reviewed-jswasm.log`. |
| Full Go run | Earlier `go test ./...` reported only the unchanged known P6g corpus exercise-floor failure (7 opaque names versus required 20). Subsequent engine changes passed the affected native and JS/WASM suites above. |
| Diff hygiene | `rtk proxy git diff --check` passed. |

The expanded browser coverage proves outer/below-page rectangle starts, loaded-image paint offsets, last-edge/explicit Clear with atomic Undo, preview geometry beyond the final window and within skipped intervals, and continuation tails beyond the final foot. G-8 now compares exact authored Y displacement using engine-declared page pitch; its start is verified to hit a selected continuation. Pixel tests compare Grid on/off raster output at the expected pitch and independently derived continuation phase. Focus setup preserves scroll position and body-drag tests avoid explicit resize handles at 50% zoom.

The complete unrelated browser suite was not run. No approved acceptance case is intentionally deferred. The known Go corpus exercise-floor failure is outside this change.

## Suggested Review Order

**Gesture ownership**

- Start here: one captured gesture owns selection, previews, cancellation, and atomic release.
  [use-canvas-selection.ts:19](../../folio-designer/src/use-canvas-selection.ts#L19)

- Connect stable canvas capture to engine commands and inspector state.
  [App.tsx:1100](../../folio-designer/src/App.tsx#L1100)

- Select complete logical boxes across repeated and split page occurrences.
  [canvas-selection.ts:11](../../folio-designer/src/canvas-selection.ts#L11)

**Atomic movement**

- Intersect group constraints and snap one reference without changing relative offsets.
  [group_movement.go:62](../../folio-go/group_movement.go#L62)

- Apply the accepted common delta through one atomic template replacement.
  [group_movement.go:153](../../folio-go/group_movement.go#L153)

- Fence preview and commit against engine revision and existing history.
  [engine.go:133](../../folio-go/wasm/engine.go#L133)

**Common properties and painting**

- Expose bounded authored state separately from paint defaults and hidden borders.
  [canvas_authored_properties.go:42](../../folio-go/canvas_authored_properties.go#L42)

- Intersect editable fields and choices across the selected kinds.
  [App.tsx:3443](../../folio-designer/src/App.tsx#L3443)

- Keep untouched drafts inert and reconcile accepted shared values.
  [App.tsx:3603](../../folio-designer/src/App.tsx#L3603)

- Translate image and text paint with their boxes while preserving internal offsets.
  [canvas-selection.ts:37](../../folio-designer/src/canvas-selection.ts#L37)

**Verification and transport**

- Validate the additive read-only preview transport.
  [engine-protocol.ts:333](../../folio-designer/src/engine-protocol.ts#L333)

- Exercise deferred previews, pending commits, cancellation, and click suppression.
  [use-canvas-selection.test.ts:24](../../folio-designer/src/use-canvas-selection.test.ts#L24)

- Verify common movement, atomic refusals, precision, and read-only previews.
  [group_movement_test.go:63](../../folio-go/group_movement_test.go#L63)

- Run actual gestures, bulk edits, painted-grid checks, and engine Undo in Chromium.
  [selection-group-properties.spec.ts:33](../../folio-designer/e2e/selection-group-properties.spec.ts#L33)
