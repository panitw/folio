---
title: 'Edit table column formulas in aligned binding inputs'
type: 'feature'
created: '2026-09-12'
status: 'done'
baseline_commit: 'ad33ac77f8a8b64f01d42d219fd200e5e8082cef'
review_loop_iteration: 0
context: ['/Users/panitw/.codex/RTK.md']
---

<frozen-after-approval reason="user-requested follow-up to table column authoring">

## Intent

**Problem:** Displaying a relative field input above a duplicate binding label makes the column controls misalign. The input also prevents authors from manipulating values with the engine's formula language.

**Approach:** Use a single full-binding input, including `{{row.date}}`, and permit formula editing directly. Remove the duplicate label and keep each row's primary controls aligned, even when footer details expand.

## Boundaries & Constraints

**Always:** Go owns document state, expression validation, limits and history. Reuse the transaction and editor session fences. Preserve DATA relative-field picking and existing Add width fitting. Submit full input text unchanged; allow valid literals and interpolation under existing engine rules. Sample suggestions show full bindings scoped to the current collection/alias.

**Ask First:** A new expression language or persisted schema is outside this request.

**Never:** Parse/evaluate formulas in the browser, silently rewrite existing formulas, weaken footer/alias validation, or change unrelated geometry. This request supersedes the previous simple-only input and duplicate binding-display design.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Simple or sample pick | Full `{{row.date}}` or scoped option | Same full binding appears in one editable control | Engine validates before commit |
| Formula | Existing or entered `{{upper(row.trn_code)}}`, date or numeric formatting | Exact binding remains editable, persists and renders through existing engine | Invalid syntax/function/type rejected atomically with located diagnostic |
| Clear / literal | Empty input or valid text/interpolation | Preserve canonical existing text semantics; empty clears; unchanged is no-op | Footer dependencies retain canonical refusal |
| Alignment | Bound/unbound rows, expanded footer fields | Header, binding, width and aggregate primary controls stay aligned | Verify browser geometry |
| Actions/history | Dirty formula then Add, Done, Escape or Cancel | Save before Add/close; Cancel discards; undo/redo/reopen preserve exact bytes | Rejection/session revocation cannot dispatch a later stale action |
| Boundaries | Missing/null/nonstring/overlong input; implicit aggregate source; alias change | Existing limits and supported migrations hold; explicit footer source permits formulas | Refusal changes neither document nor history |

</frozen-after-approval>

## Code Map

- `folio-designer/src/TableEditor.tsx:96,530`: pending-field/action helpers must compare complete `column.binding`; replace input/output/explanation with one enabled input named `Binding for column N`. Preserve refusal resets and action sequencing. Use normal Left/Right/Home/End caret behavior in this input; Alt+Left/Right can move between matrix cells, Alt+Down retains native suggestions.
- `folio-designer/src/App.css:849,878`: matrix rows center cells whose heights differ. Remove duplicate display and align primary controls at row start; keep grid tracks and narrow viewport scroller.
- `folio-designer/src/App.tsx:95,1164,3030`: safe candidate paths are identifier segments; binding callback uses the fenced commitTableColumn result. DATA continues using updateTableColumnBindingCommand.
- `folio-designer/src/table-column-command.ts`: add opaque explicit-expression encoder. Full datalist options may format projected alias and safe candidate path for presentation only; filter against existing binding bound, never parse input or construct a document AST.
- `folio-go/component_commands.go:315,670`: add sibling `updateTableColumnExpression` command with id/columnId/binding; explicit required string, empty allowed, non-null, at most maxCanvasBindingString bytes. Dispatch through applyTableColumnCommand; retain old relative-field command unchanged.
- `folio-go/folio_expr_validate.go:304`, `internal/expr/check.go`, `table_columns_projection.go:320`: existing ParseTemplate validates complete text and derives footer sources; complete Binding already projected. No new DTO/schema or parser needed. RowFieldEditable remains projection compatibility, not input edit gating.
- `folio-go/internal/expr/footer.go`: migration supports simple paths/direct formatNumber and refuses unsupported alias rewrites. Preserve rules and verify atomicity.
- `folio-designer/src/TableEditor.test.tsx`, `App.test.tsx`, command/protocol tests and existing `e2e/*.spec.ts`: replace obsolete relative-input/output expectations with full input values and retain previous interaction checks.
- PRD FR18 requires existing formatting/string/logic functions. `{{formatNumber(row.amount * 1.07, "#,##0.00")}}` is valid; unformatted numeric arithmetic as text is intentionally refused.

## Tasks & Acceptance

**Execution:**
- [x] Engine command, encoder and App wiring accept full binding text without changing DATA picks.
- [x] TableEditor and scoped CSS provide aligned full-expression input, suggestions and formula-friendly caret behavior.
- [x] Go/WASM tests verify command admission, formulas, bounds, footers and history; frontend tests cover input, actions and refusals; update affected existing browser specs.
- [x] Coordinator adds `folio-designer/e2e/table-column-formulas.spec.ts` for browser alignment, formulas, save/reopen, history and invalid refusal, and runs production browser checks.
- [x] Review and final checks complete.

**Acceptance Criteria:**
- Given the screenshot's bound columns, when the editor opens, then each full binding appears only in its aligned input and can be edited.
- Given a date or code field, when a valid formatting/string formula is entered, then the saved/reopened binding and engine-rendered value reflect it.
- Given an invalid formula, when committed, then the previous binding remains intact and the editor reports the engine diagnostic.

## Spec Change Log

- Review identified native input line-break stripping and multiline caret interception. Preserve existing CRLF bytes on untouched open/close; promote pasted or Shift+Enter drafts before sanitization; retain action fences, exact engine binding text and aligned primary controls.
- Review corrected alias-specific examples, mock binding metadata and existing browser expectations for full bindings, native matrix navigation and parser diagnostics.
- Browser evidence confirmed alignment and persistence. The PDF fixture now uses the engine's required RFC3339 date value; the caret test uses the native macOS end-of-line gesture. Neither correction changes formula or keyboard semantics in production.

## Design Notes

The full-expression command is distinct from DATA's path intent. Presentation-only suggestion strings are drafts; Go authorizes every commit. Keep whitespace and literals accepted by canonical text semantics. The implementation subagent owns production and existing tests; the coordinator owns the new formula browser regression and browser runs. The subagent does not commit or run competing browser builds; the coordinator creates the reviewed local commit. Local evidence may refine implementation within these boundaries.

Existing bindings containing line breaks use a one-row textarea because native single-line inputs strip CR/LF. Pending/blur detection compares each control against its mounted DOM value, preserving canonical CRLF bytes on untouched open/close while intentional edits submit the control text unchanged. The same action fences, refusal resets and focus trap cover both controls.

Review refinement: multiline paste and Shift+Enter promote a single-line draft to a compact textarea before the browser can strip CR/LF. Promotion preserves replacement selection and caret placement, remains local until normal commit, and keeps the textarea mode for the current editor session so resets preserve focus. Multiline Up/Down and selection keys stay native; vertical resizing exposes the complete text. Help uses the projected alias and identifies single-line-only datalist shortcuts. Go remains the sole expression and migration authority.

## Verification

- Go root and WASM: `rtk proxy go test -count=1 . ./wasm/...`.
- Frontend: focused Vitest, application TypeScript, E2E compilation and lint.
- Coordinator production browser: new formula spec plus table-column-authoring, table-editor, table-matrix-layout, table-column-binding and browser-native-roundtrip specs. Inspect screenshot and measured control coordinates.
- `rtk git diff --check` and local reviewed commit; no push.

## Verification Results

- Final full Vitest run: 80 files and 1,593 tests passed. The implementation patch pass also verified 626 tests across six focused suites.
- Final production browser run: all 16 tests passed, including four new formula cases, previous column authoring and DATA workflows, keyboard/grid layout, and exact browser/native PDF roundtrip.
- Verified PDF text contains the formula results `12/09/2026`, `ABC-12`, and `107.00`. Browser geometry confirms aligned primary controls for expanded footers and multiline inputs, including a narrow viewport; the final table screenshot was visually inspected.
- Go root and WASM suites passed; targeted expression suites passed after review.
- Application TypeScript and E2E compilation passed. Lint has eight existing Fast Refresh warnings and no errors. Diff whitespace check passed.
- Follow-up review of multiline promotion, reset, caret and pending-action handling returned no additional findings.
- Native datalist option content and invocation are covered; operating-system popup selection is not claimed as automated coverage.

## Suggested Review Order

**Binding input and layout**

- Edit complete expressions in one aligned control, preserving existing multiline text.
  [TableEditor.tsx:545](../../folio-designer/src/TableEditor.tsx#L545)

- Preserve pasted line breaks and selection before browser input sanitization.
  [TableEditor.tsx:110](../../folio-designer/src/TableEditor.tsx#L110)

- Align primary controls even when a footer expands.
  [App.css:849](../../folio-designer/src/App.css#L849)

**Engine admission**

- Keep explicit expression intent separate from DATA field picks.
  [table-column-command.ts:26](../../folio-designer/src/table-column-command.ts#L26)

- Validate bounded full text through the existing atomic engine transaction.
  [component_commands.go:718](../../folio-go/component_commands.go#L718)

**Regression evidence**

- Verify exact text, malformed expressions, footer dependencies and alias migration.
  [component_column_expression_test.go:20](../../folio-go/component_column_expression_test.go#L20)

- Exercise aligned authoring, rendered PDF values, persistence and multiline actions in the browser.
  [table-column-formulas.spec.ts:79](../../folio-designer/e2e/table-column-formulas.spec.ts#L79)
