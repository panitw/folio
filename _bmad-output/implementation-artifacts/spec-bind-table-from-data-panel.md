---
title: 'Bind table collections from the Data panel'
type: 'bugfix'
created: '2026-09-12'
status: 'done'
baseline_commit: 'c94f973a503a4413650c14fbd604c243723a4641'
review_loop_iteration: 0
context:
  - /Users/panitw/.codex/RTK.md
  - /Users/panitw/Projects/folio/_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md
---

<frozen-after-approval reason="user-requested interaction">

## Intent

**Problem:** Selecting a table and clicking a sample JSON array only expands the Data tree. The user expects that click to assign the array to the table, just as picking a scalar binds text.

**Approach:** Offer root-addressable arrays when a whole table is selected. Clicking or activating an array immediately binds the table's collection, updates its canvas label, and permits undo in one step.

## Boundaries & Constraints

**Always:** Preserve table alias (including absent alias), columns and their expressions, layout, and all other document settings. Go owns canonical syntax, admission and document mutation. TypeScript sends decoded key segments through opaque commands. Use existing history, preview invalidation, in-flight and stale-response guards. Sample data stays local.

**Never:** Bind a scalar to a whole table, bind a collection to text or a selected column, admit runtime params or nested arrays inside rows as root collections, silently reinterpret a JSON key containing punctuation as several keys, or change column sizing/placement. Keep existing TableEditor controls available.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected behavior |
| --- | --- | --- |
| Collection pick | Whole table; populated, empty or nested-object array | Click, Enter or Space sends one collection bind; returned canvas shows selected collection |
| Browsing | Array with children | Activation binds and toggles disclosure; ArrowLeft/Right only browse, without binding |
| Selection gates | Text, image, line, rectangle, missing selection/kind, or column | Existing scalar/column behavior; no whole-table collection dispatch |
| Scope gates | params array/descendant, root JSON array, nested row array, truncated ancestor key | Explain unavailable binding; browsing remains possible; no command |
| Invalid key | Decoded key such as a.b, empty, unicode, control character | Go rejects unsupported collection grammar without mutation; located Data-panel error |
| Preservation/history | Explicit or implicit alias, existing column bindings | Only collection changes; one undo restores it; redo reapplies; repeated same bind is history-neutral |
| Concurrency/error | Pending request, selection/sample/document replacement | Shared binding latch prevents duplicates; stale projection/error fences behave as scalar binding does |

</frozen-after-approval>

## Code Map

- `folio-designer/src/DataPanel.tsx`: `rowFor`, `DataTree` activation currently prioritize branch expansion; scalar and column gates and scoped error UI already exist. Collection mode needs its own row reasons and context, keeping scalar TABLE ONLY badge outside table mode.
- `folio-designer/src/App.tsx`: `bindPickedPath` owns Data-panel latch and generation/revision fencing; `selectedComponent.tableBind` is authoritative displayed collection. Reuse this handler with command selection by current engine component kind, avoiding a second state machine. `bindPickedColumn` remains independent.
- `folio-designer/src/component-command.ts`, `table-column-command.ts`: central JSON encoders; transport segments with no path-building in TypeScript.
- `folio-go/component_commands.go`: dispatch, `applyTableColumnCommand`, `configureTableBinding`, scalar picked-segments validation. Add `bindTableCollection` accepting only id and segments, preserving alias in Go without a query/configure race. Each segment must be an existing bounded identifier; resulting collection must fit existing root-collection grammar/bound and refuse params. Transactional seam supplies canonical reparse.
- `folio-designer/src/sample-data.ts`: collection segments must remain available for empty/populated root arrays. `array` currently ignores rootScoped; honor it so row arrays and descendants of truncated keys cannot masquerade as root candidates. Other readers use these segments for collection discovery.
- `folio-designer/src/DataPanel.test.tsx`, `table-column-binding.test.tsx`: existing App mocks, selection gates, refusals and data tree coverage. Update obsolete TableEditor-only expectations narrowly.
- `folio-go/component_commands_test.go`, `folio-go/wasm/*test.go`: Go admission, preservation and engine history evidence.
- `folio-designer/e2e/table-column-binding.spec.ts`, `table-placement.spec.ts`: production pointer tests and local-file fixtures; use Playwright boundingBox, never browser geometry APIs in app/tests. Fixture statement table e8 has alias txn and existing columns.

## Tasks & Acceptance

- [x] Add bounded Go collection-pick command and escaped TS factory; verify transactional refusals, preservation and history.
- [x] Enable whole-table array picks, keyboard disclosure, accurate context/current binding and errors; preserve text/column mode.
- [x] Correct collection discovery scope, with focused regression evidence for the matrix.
- [x] Add a production browser flow: select whole table, bind another sample array, verify canvas and one-step undo/redo, then pick a column field from the new collection.
- [x] Run appropriate Go, frontend and production-browser checks and report exact outcomes.

**Acceptance:** The screenshot's fresh `items[]` table can be mapped to `transactions[]` directly by clicking that array in DATA. Existing table alias and column expressions survive collection changes, and column selection still offers only the selected collection's fields.

## Design Notes

Use `bindTableCollection` rather than resetting alias through `configureTableBinding`: aliases are document state absent from the main canvas projection. Choosing an array preserves the alias and column expressions; only the collection and its qualified footer-source prefixes change. Reuse existing code paths where possible; adjust implementation detail when local evidence requires it while preserving the intent and boundaries above. Do not commit; the coordinating agent performs review and commits.

Review clarification: preserving an explicit footer source means preserving its row-relative field, footer operation, and formatting. Its collection prefix must follow the changed table collection, because canonical parsing rejects a footer pointing outside that collection. Share that prefix update with the existing Configure columns command. Keep ordinary column expressions and aliases unchanged. Collection admission must match Configure columns: individual decoded segments must be identifiers, while the existing 256-byte whole-collection bound supplies the length limit; do not add scalar-picker-specific key/count limits.

## Verification

- `rtk proxy go test -count=1 . ./wasm/...` in `folio-go`.
- Focused Vitest checks for DataPanel, column binding, sample discovery, command escaping and ownership/protocol contracts; TypeScript build checks.
- Production Playwright collection-binding and existing column/placement tests. Run `rtk npm run test:e2e:compile` before browser checks. A browser run performs the production/WASM build itself; avoid concurrent preview servers on port 4173.

## Spec Change Log

- 2026-09-12: Clarified preservation of relative footer fields and shared collection admission after review. Preserve aliases, column expressions, layout, formatting, and transactional history.

## Review Notes

- Fixed explicit footer sources preventing collection changes; both collection-editing routes now preserve the relative footer field and rebase its collection prefix. Overlong resulting sources are refused atomically within the existing editor limit.
- Unified collection admission with Configure columns, removing scalar-specific segment length/count restrictions.
- Added real-browser tree focus, Enter/Space activation, and arrow-only browsing checks using a different collection to distinguish accidental binding from a no-op.
- Recorded pre-existing picked-row highlighting and duplicate-JSON-key discovery issues in deferred work. Busy-state notices, generic malformed-envelope errors, rendering coverage, and shared revision-fence coverage did not require changes to this feature.
- Edge-case and verification-gap reviewers returned no additional findings.

## Verification Results

All final checks passed on 2026-09-12.

- In `folio-go`: `rtk proxy go test -count=1 . ./wasm/...` — root package passed in 11.467s; WASM package passed in 1.223s.
- In `folio-designer`: `rtk proxy npx vitest run src/DataPanel.test.tsx src/table-column-binding.test.tsx src/sample-data.test.ts src/component-command.test.ts src/engine-ownership-contract.test.ts src/engine-bounds-mirror.test.ts src/command-json-soleness.test.ts src/command-json.test.ts src/engine-protocol.test.ts` — 9 files, 228 tests passed after all review fixes.
- Broader frontend run: `rtk proxy npx vitest run` — 80 files, 1,550 tests passed before the final refusal test and footer-only review fixes; the final affected-file run above includes the added test.
- `rtk npm run test:e2e:compile` — passed.
- `rtk npm run test:e2e -- e2e/table-column-binding.spec.ts e2e/table-placement.spec.ts` — full production/WASM/offline rebuild, then 8 browser tests passed in 2.9 minutes. Includes real clicks, keyboard focus/activation, browsing without mutation, fresh and aliased table binding, undo/redo, column scope, and existing placement/zoom checks.
- Targeted oxlint and `rtk git diff --check` — passed.
- Inspected the regenerated fresh-table screenshot: selected table spans the content width at 24pt height, canvas and DATA both show `transactions[]`, and arrays are offered as table bindings.

Screenshot evidence: [fresh table bound from DATA](../../folio-designer/test-results/table-placement-maps-a-fre-d42da-ransactions-array-from-DATA/fresh-table-collection-binding.png).

## Suggested Review Order

- Enable collection picks while preserving column and scalar selection modes.
  [DataPanel.tsx:99](../../folio-designer/src/DataPanel.tsx#L99)

- Share existing command fencing, preview invalidation, and engine-owned state.
  [App.tsx:1321](../../folio-designer/src/App.tsx#L1321)

- Validate decoded keys without changing aliases or column expressions.
  [component_commands.go:527](../../folio-go/component_commands.go#L527)

- Rebase footer sources atomically within the editor’s existing limit.
  [component_commands.go:570](../../folio-go/component_commands.go#L570)

- Keep row arrays and incomplete ancestor paths out of root discovery.
  [sample-data.ts:95](../../folio-designer/src/sample-data.ts#L95)

- Verify actual click and keyboard workflows, including undo and redo.
  [table-placement.spec.ts:10](../../folio-designer/e2e/table-placement.spec.ts#L10)

- Verify alias preservation and subsequent column mapping in the browser.
  [table-column-binding.spec.ts:76](../../folio-designer/e2e/table-column-binding.spec.ts#L76)

- Exercise explicit, derived, and absent footer sources in one shared fixture.
  [table-collection-footers.folio:1](../../folio-go/testdata/commands/table-collection-footers.folio#L1)

