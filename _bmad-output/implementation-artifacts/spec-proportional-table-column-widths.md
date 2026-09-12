---
title: 'Size table columns by proportion'
type: 'feature'
created: '2026-09-12'
status: 'done'
baseline_commit: '57536ce5b04768f7d02b63d2af2ca4691d7a8104'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Point widths require manual calculation and rebalancing; adding columns currently splits only one existing column.

**Approach:** Add Proportion sizing, dividing an authored table width according to positive column weights. New tables and columns default to proportion 1. The Table Editor exposes a total width in points and each column's proportion. Previously saved files need no compatibility or conversion workflow: the user confirmed the product has not been in production.

## Boundaries & Constraints

**Always:** Go owns allocation, validation, persistence, and canvas/PDF geometry. Preserve millipoint arithmetic, atomic history, editor sessions, bindings, and aligned controls. Store proportions persistently. Existing parser paths may remain where removing them would create unrelated fixture churn; backwards compatibility is not an acceptance requirement.

**Ask First:** Mixed units within one table or content-dependent sizing.

**Never:** Allocate geometry in React, silently clamp invalid ratios, or expand table drag/resize behavior.

## I/O & Edge-Case Matrix

| Scenario | Input | Outcome | Refusal |
|---|---|---|---|
| Equal | 500pt; ten or five weights of 1 | Each gets 50pt or 100pt | None |
| Unequal | 500pt; 1:2:1 | 125pt, 250pt, 125pt | None |
| Structure | Add/remove a column | New weight 1; retain total and surviving weights; redistribute | Existing count limit |
| Empty | Remove last, then add | Retain total; new column receives all of it | None |
| Round | 500pt; 1:1:1 | Exact sum; widths differ by at most 0.001pt | None |
| Invalid | Blank/nonpositive/malformed/overprecision or zero-width allocation | Restore committed controls and retain document/history | Located explanation |
| Total | 500pt becomes 400pt; 1:2:1 | 100pt, 200pt, 100pt | Reject beyond available band width |

</frozen-after-approval>

## Code Map

- `folio-go/internal/template/{model.go,parse_bands.go,serialize.go,version.go}`: typed format, table-width prohibition, canonical serialization, and content-triggered versions.
- `folio-go/component_commands.go`: `projectedSize`, creation, add/remove/update, containment, transactions.
- `folio-go/{table_render.go,page_setup.go,table_columns_projection.go}`: renderer width slice, `canvasTableColumns`, editor projection; these consume absolute widths directly today.
- `folio-designer/src/{engine-protocol.ts,table-column-command.ts,TableEditor.tsx,App.tsx,App.css}`: strict projection admission, opaque commands, controls, width budget, session-aware commit.
- Template geometry/drift/version tests and `folio-go/canvas_projection_wire_test.go` pin contracts that must change together.

## Tasks & Acceptance

**Execution:**
- [x] Template files plus `folio-go/internal/template/table_widths.go`: implement representation, validation, shared exact allocation, and version support.
- [x] Go command/render/projection files: support proportional authoring and structural edits; consume shared geometry.
- [x] Designer files: add total input, editable ratios, and resolved point widths; preserve focus, refusal, busy, Done, Add, and Cancel behavior. No migration or mode-conversion UI is required.
- [x] Go/template/WASM and designer component/protocol/browser tests: cover every matrix row, persistence, history, layout, and real PDF agreement.
- [x] Update `_bmad-output/specs/spec-folio/folio-format.md` and AD-13 in `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md`.

**Acceptance Criteria:**
- Given a new table, when opened, then it uses proportional sizing with one weight of 1 and the band's available total width.
- Given authored proportions, when saved/reopened, then total, ratios, bindings, and canvas/native/browser PDF geometry agree.
- Given sizing/structural edits, when Undo/Redo or Cancel runs, then total, weights, and columns restore together.
- Given formulas or expanded footers, when ratios change, then inputs remain horizontally aligned and keyboard reachable.

## Spec Change Log

- 2026-09-12: User confirmed previously saved files can be ignored because the product has not reached production. Removed legacy byte/output guarantees and mode-conversion requirements. Keep proportional allocation, exact totals, persisted ratios, and history behavior.

## Design Notes

Table-level `width` is the authored total only in proportional mode; each column then requires `proportion` instead of `width`. Without table width, the existing representation and point-input fallback may remain for fixtures. Expose no conversion workflow or sizing-mode selector. Reject mixed modes. Reuse `Element.Width`; extend `Column`. Keep the authored total when empty.

Accept positive proportions with at most three decimal places, using exact scaled integers; avoid unrelated line-spacing bounds. Allocate millipoints with overflow-safe arithmetic and largest remainders, ties by column order. Reject zero-width results. Keep layout unaware of proportions: feed it resolved millipoints.

Proportional content may use the already investigated 3.0 content requirement and supported ceiling; no migration or backwards-compatibility machinery is required. Keep format version handling coherent with existing content-triggered serialization.

## Verification

- `rtk go test -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` in Go root (the same documented quarantine as CI); `rtk go test ./...` in WASM. The unchanged corpus baseline was separately reproduced at P6g=7, below its required 20; its baseline-drift test must stay green.
- `rtk npm test`, `rtk npm run typecheck`, `rtk npm run lint`, and `rtk npm run test:e2e:compile` in designer.
- Production browser tests: new `e2e/table-proportions.spec.ts` plus column-authoring, column-formulas, table-editor, matrix-layout, and browser-native-roundtrip suites. Inspect aligned controls in a screenshot.
- All relevant tests pass; no new warnings; `rtk git diff --check` clean.

### Implementation verification results (2026-09-12)

- Go root: `rtk go test -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` passed 2,640 tests in 18 packages after review fixes. The unchanged corpus exercise-floor test is explicitly quarantined by CI; its baseline-drift test passed.
- WASM: `rtk go test ./...` passed 62 tests, including structural refusal diagnostics, unchanged bytes/IDs/revision, and intact Undo/Redo history.
- Designer: 1,607 tests passed in 80 files; the targeted TableEditor suite passed 125 tests. Type checking and browser-test compilation passed. Lint passed with the eight existing Fast Refresh warnings and no new warnings.
- All 16 relevant production browser tests passed against the final build: 12 in the combined run and all four proportional tests in the focused rerun after correcting a test-only wait for repeated file loading. The exact-decimal, total-blur Tab/click, and dirty-total Add/Escape regressions passed.
- Browser/native PDF bytes match for saved proportional columns with formulas and an expanded footer. Browser coverage also checks authored totals and ratios, resolved canvas widths, the user's equal-weight examples, numeric refusals, and pending-input actions.
- The final aligned-controls screenshot at `/tmp/folio-aligned-proportions.png` was visually inspected. Header, binding, proportion, alignment, and footer controls share the same row alignment; expanded footer fields sit below their primary control.
- The parent independently ran the lint module's architecture/production scans: 227 tests passed in four packages.
- Review patches and `rtk git diff --check` pass. The production build retains the existing chunk-size and offline asset-budget warnings.

## Review Triage

The three prescribed review layers completed. Duplicate findings were consolidated; no intent changes were required. All patches below are implemented, independently inspected, and verified; no findings remain open.

| Finding | Severity | Route | Resolution |
|---|---|---|---|
| Dirty total plus Add strands keyboard focus | medium | patch | Restore focus after commits without stealing intentional navigation; exercise Escape directly. |
| Structural allocation errors lose diagnostic location | medium | patch | Wrap allocator errors and verify removal, movement, and Add refusals remain atomic. |
| Proportional editor shows obsolete 72pt split help | low | patch | Restrict that explanation to point sizing. |
| Resolved-width outputs create many live regions | low | patch | Match the existing budget's explicit `aria-live="off"`. |
| Native proportion steppers can round exact large values | medium | patch | Use decimal text entry and retain exact authored strings. |
| Overflow message describes proportions as millipoints | low | patch | Report the dimensionless limit explicitly. |
| Unequal remainder ordering lacks an expected allocation | medium | patch | Add allocations that distinguish remainders from authored weights. |
| Empty persistence test saves only after adding again | medium | patch | Save and reload the empty table before adding its replacement column. |

Structural history assertions accompany the diagnostic patch, including ID allocation rollback. Total blur also preserves the native Tab/click destination if a save temporarily disables that control, while subsequent intentional focus takes precedence.

## Suggested Review Order

**Authoring**

- Start with the total-width control and per-column proportions visible to authors.
  [TableEditor.tsx:607](../../folio-designer/src/TableEditor.tsx#L607)

**Exact allocation**

- Resolve one exact total using largest remainders, shared by every geometry consumer.
  [table_widths.go:43](../../folio-go/internal/template/table_widths.go#L43)

- Validate dimensionless decimal weights without floating-point conversion.
  [table_widths.go:20](../../folio-go/internal/template/table_widths.go#L20)

**Atomic editing and diagnostics**

- Apply column changes on a candidate so refusal preserves the original document.
  [component_commands.go:319](../../folio-go/component_commands.go#L319)

- Validate total-width edits against allocation and available band space.
  [component_commands.go:584](../../folio-go/component_commands.go#L584)

- Retain the affected column and field when structural redistribution fails.
  [render_error.go:123](../../folio-go/render_error.go#L123)

**Canvas and PDF agreement**

- Paint the canvas using the same resolved widths consumed by PDF layout.
  [page_setup.go:1931](../../folio-go/page_setup.go#L1931)

- Feed resolved millipoints into the existing table geometry and pagination path.
  [table_render.go:768](../../folio-go/table_render.go#L768)

**Format and verification**

- Persist authored proportions while keeping resolved widths out of proportional columns.
  [model.go:483](../../folio-go/internal/template/model.go#L483)

- Raise the format requirement only when proportional content needs it.
  [version.go:117](../../folio-go/internal/template/version.go#L117)

- Check equal shares, unequal remainders, exact totals, and overflow-safe arithmetic.
  [table_widths_test.go:26](../../folio-go/internal/template/table_widths_test.go#L26)

- Prove structural refusals preserve document bytes, IDs, revision, Undo, and Redo.
  [table_proportions_test.go:102](../../folio-go/wasm/table_proportions_test.go#L102)

- Exercise exact editing, keyboard focus, persistence, alignment, and browser/native output.
  [table-proportions.spec.ts:134](../../folio-designer/e2e/table-proportions.spec.ts#L134)
