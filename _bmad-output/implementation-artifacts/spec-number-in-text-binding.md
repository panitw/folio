---
title: 'A number bound into text prints as its exact decimal'
type: 'feature'
created: '2026-09-13'
status: 'done'
route: 'dispatch'
baseline_commit: 'cc8bd8f461a82bb8e384e89f32708b9f05aedcdc'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
  - '{project-root}/docs/expression-reference.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A text element or table cell bound to a JSON number (e.g. `{{header.balance_available}}`) fails the whole render with "resolved to a number, not a string — text bindings are never coerced", so a statement with numeric data can't preview or print without wrapping every field in `formatNumber`.

**Approach:** Owner decision (2026-09-13): a number that resolves in a TEXT binding prints as its exact decimal text. There is no grouping, locale, rounding or float; `formatNumber` stays the way to get styled output. This revises AD-14's "never coerced" rule for **numbers in text only**.

## Boundaries & Constraints

**Always:**
- The exact decimal text comes from the engine's Decimal. The coefficient digits print with a `.` placed by the exponent, and the scale is kept: `1234.50` → `1234.50`, `2067071865` → `2067071865`, `-3.5` → `-3.5`, `1e3` → `1000`, `-0` → `0`, `0.000` → `0.000`. There is no exponent notation.
- The same printer serves data values and computed numbers (arithmetic, `count`/`sum`/`avg`) in every text consumer: text elements, table data cells, and footer cells.
- Booleans, arrays and objects in text stay located errors, unchanged. `visibleIf` (boolean) and footer-source (number) kind checks are unchanged.
- The load/commit static check (`CheckText`) admits the number kind for text, so `{{1}}`, `{{count(items)}}` and `{{a + b}}` are valid in text.
- No-data preview: a path used bare in text may take a zero stand-in, so a path shared by text and `formatNumber` no longer refuses preview.
- Format version: `SupportedVersion` rises to **3.3**. On save, a document rises to `3.3` when any text expression is statically known to be able to return a number without data (`KnownKinds` includes number). A plain path whose kind depends on data cannot be detected. That case is DISCLOSED in folio-format.md, following the `style.border` precedent, and not mechanised.

**Never:**
- No change to `formatNumber`, `formatDate`, `upper`/`lower` operand rules, conditions, or footer aggregate typing.
- No locale digits, grouping separators or thousands formatting in the default text.
- No designer-side type rule; the designer keeps showing engine results.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Data number in text | `{{v}}`, data `{"v": 1234.50}` | prints `1234.50` | N/A |
| Integer / negative / exponent | `2067071865`, `-3.5`, `1e3`, `-0` | `2067071865`, `-3.5`, `1000`, `0` | N/A |
| Computed number | `{{count(items)}}` over 3 items; `{{a + b}}` with 1.5 and 2 | `3`; `3.5` | overflow/zero divisor stay located errors |
| Mixed placeholder text | `Total: {{v}} THB` with v = 10 | `Total: 10 THB` | N/A |
| Table cell / footer | column bind `{{row.amount}}` with 99.9 | cell prints `99.9` | N/A |
| Boolean in text | `{{flag}}` with true; `{{true}}` | unchanged: located error | existing message (booleans only) |
| Static check | commit text value `{{1}}` | accepted; saved document declares ≥ `3.3` | N/A |
| No-data preview | `{{v}}` and `{{formatNumber(v, "0")}}` in one document | preview renders (stand-in 0) | N/A |

</frozen-after-approval>

## Code Map

- `folio-go/internal/bind/text.go:266-277` -- `Resolve` kind switch; add a `KindNumber` arm that writes the Decimal text. Update the rule comment at ~L136-140.
- `folio-go/internal/bind/value.go:56`, `text.go:417-428` -- the number is decoded via `UseNumber` and converted with `AsDecimal`; the Decimal is what gets printed.
- `folio-go/internal/expr/decimal.go:27` -- `Decimal{Coefficient int64, Exponent int}`; add an exact plain-text method (e.g. `Text()`). Do NOT reuse `UnformattedPattern`/`renderNumberPattern` (they cap fraction digits or add locale formatting).
- `folio-go/internal/expr/check.go:88-94` -- `CheckText` allowed kinds; add `numberKind`, keep bool rejected, update the message.
- `folio-go/folio_expr_validate.go:72,145,309` -- text/column callers of `CheckText` (no change expected beyond the new acceptance).
- `folio-go/render.go:742`, `table_render.go:1237,1479` -- text consumers going through `Resolve`. Update the stale comments at `table_render.go:72-77`, `internal/bind/condition.go:4-9`.
- `folio-go/stand_in_data.go:159-165,529,666,765` -- `demandText` allowed stand-ins; add the zero stand-in (and one, if the divisor logic needs it) and update the comment.
- `folio-go/serialize_template.go:21-45`, `internal/template/version.go:88,138,366,384` -- save-time expression version floor; add a 3.3 raise when a text expression's `KnownKinds` includes number, and raise `SupportedVersion` to 3.3.
- Tests pinning the old rule: `internal/bind/text_test.go` (~L55 `row3_wrong_kind_is_error_no_coercion`, ~L283 `TestBindTextWellFormedNumberIsStillWrongKind`); `render_bind_test.go:162`; `wasm/cmd/engine/main_test.go:267`; `internal/expr/formulas_test.go:~100` (`CheckText("1")`); `folio_expr_validate_test.go:~158` (`{{a + b}}` in text); `stand_in_data_test.go:277` ("text and formatNumber share nothing"); stale comments at `render_empty_average_diagnostic_test.go:9-16`, `stand_in_data_test.go:117-120`. Tests that set `SupportedVersion`/the fixture version to 3.2 may need 3.3.
- Docs: `_bmad-output/specs/spec-folio/folio-format.md` (~L47 version row, ~L602 "Numbers … require formatNumber in text"); `docs/expression-reference.md` L41, L45, L154-165 (and regenerate/update `docs/expression-reference.html` if it mirrors the .md); `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md:317-336` (AD-14) — record the revision.

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/internal/expr/decimal.go` + test -- exact `Text()` for Decimal covering every row of the I/O number-formatting cases -- one printer.
- [x] `folio-go/internal/bind/text.go` + `text_test.go` -- print numbers in `Resolve`; keep bool/array/object errors (message names the kind; drop "never coerced" wording only where it now lies) -- runtime.
- [x] `folio-go/internal/expr/check.go` + `formulas_test.go`, `folio_expr_validate_test.go` -- admit number kind in `CheckText` -- load/commit.
- [x] `folio-go/stand_in_data.go` + test -- text demand admits a zero stand-in -- no-data preview.
- [x] `folio-go/serialize_template.go`, `internal/template/version.go` + tests -- 3.3 raise for statically number-kind text expressions; SupportedVersion 3.3 -- format version.
- [x] `folio-go/render_bind_test.go`, `table_render*_test.go`, `wasm/cmd/engine/main_test.go` -- flip the pinned refusals into rendering assertions (text element, table cell, mixed placeholder, computed count) and add a boolean-still-errors case -- coverage.
- [x] `folio-format.md`, `docs/expression-reference.md` (+ `.html` if generated from it), `ARCHITECTURE-SPINE.md` AD-14 -- state the new rule, the exact-text form, 3.3, and the disclosed undetectable-path case -- docs.

**Acceptance Criteria:**
- Given the statement template with `{{header.balance_available}}` bound to a number, when previewed, then the PDF renders with that value printed as its exact decimal and no render failure.
- Given the Go suite and designer suite, when run, then everything that passed before still passes (except `internal/text` `TestCorpusMeetsP6ExerciseFloors/P6g`, which already fails on the baseline).

## Design Notes

Printing (sketch):

```go
// Text is the exact decimal: coefficient digits with '.' placed by Exponent.
func (d Decimal) Text() string // {123450,-2}→"1234.50"; {1,3}→"1000"; {-35,-1}→"-3.5"; {0,-3}→"0.000"
```

Mind the int64 minimum when taking the absolute value.

Why 3.3 and not a gate on paths: the kind of `{{row.amount}}` is data, not document. An older reader given such a document still loads it and fails at render with a located error, which is no silent wrong output. Only statically number-typed expressions are detectable, and those raise.

## Verification

**Commands:**
- `cd folio-go && go test ./...` -- expected: pass except the pre-existing `internal/text` P6g corpus failure
- `cd folio-go && gofmt -l .` -- expected: empty
- `cd folio-go/wasm && go test ./...` (if a separate module) -- expected: pass
- `cd folio-designer && npx tsc -b && npx vitest run` -- expected: pass
- `cd folio-designer && npx playwright test e2e/table-column-formulas.spec.ts` -- expected: pass

## Implementation Notes

- One printer: `expr.Decimal.Text()` (uint64 magnitude, so `math.MinInt64` is safe; a zero coefficient never prints a sign or appends zeros for a positive exponent). `bind.Resolve` gains a `KindNumber` arm; the wrong-kind message now reads "resolved to a bool, not a string or number — a bool is never coerced to text".
- `CheckText` admits `numberKind`. Text stand-in demand admits `standInZero` and `standInOne` (after the empty string, so text-only paths still stand in as `""`); the one-stand-in lets `{{d}} {{10 / d}}` preview.
- Version: `template.TextNumberExpressionVersion = "3.3"` is exported and applied from `serialize_template.go`'s `expressionMinimumVersion` (it dominates the 2.0 formula floor). It is not a `versionRank` member, because the trigger is expression-derived rather than content-key-derived. `{{avg(x)}}` and `{{flag ? 1 : "x"}}` raise, since their `KnownKinds` include number. `if(flag, 1, path)` does not raise: `KnownKinds` is nil when a branch is data-dependent (same disclosed undetectable class).
- Pinned tests flipped beyond the Code Map: `boolean_formulas_test.go` (placeholder-location values now true/false), `component_column_expression_test.go` and `wasm/table_column_expression_test.go` (`{{row.amount * 1.07}}` replaced by boolean-kind `{{row.amount > 1.07}}` as the refused bind), `render_empty_average_diagnostic_test.go` (non-empty avg now renders `1.5000` with no diagnostic), `folio_expr_validate_test.go` (fixture now `{{nosuch(a + b)}}`). New: `internal/expr/decimal_text_test.go`, `render_number_in_text_test.go`.
- Table footer keeps `formatNumber` wrapping (unchanged); only its stale comment was updated.
- Acceptance probe (temporary, deleted): statement-1 with `Balance: {{header.balance_available}}` and data `2067071865.50` drew `Balance: 2067071865.50`; its stand-in document previewed without error.

## Spec Change Log

## Review Triage Log

| # | Source | Finding | Verdict | Evidence | Route |
|---|--------|---------|---------|----------|-------|
| 1 | blind | `docs/expression-reference.html` still states the old rule | low | L360: "Text still requires string or null … Numbers require formatNumber." | patch |
| 2 | blind, edge (×2) | Printing a positive exponent (data `1e100000`) is uncharged: an 8-byte value writes a 100,001-char run per placeholder/row | low | Exponent bounded at 100000 (arithmetic.go:9), but `Decimal.Text()` in `Resolve` runs after `EvalWithBudget` with no `budget.Charge`; the budget is in scope (text.go:258) | patch |
| 3 | blind | Version test's plain-path check accepts "3.0" or "1.0" | low | Loose assertion; direct correction to the exact version | patch |
| 4 | blind | Render test uses substrings ("3", "10") already present in other values | low | "3" is in "1234.50", "10" in "0.10"/"101.00"; direct correction to whole-run or per-element matches | patch |
| 5 | blind, edge | Docs (folio-format.md, AD-14, test header) claim footer cells print exact decimals; footers always go through formatNumber | low | table_render.go footerCellExprText always wraps formatNumber, so no bare number reaches a footer; direct doc/comment correction | patch |
| 6 | blind | Arrays/objects in text have no new tests; the message can only say bool | low | Arrays are refused before kind dispatch (TestBindTextArrayPathIsLocatedNotScalarError, still green); expr.Kind has no array/object value | reject |
| 7 | blind | No test pins `if(flag, 1, path)` staying below 3.3 | low | Developer-only; behaviour documented in Implementation Notes | reject |
| 8 | blind | Docs don't say a 3.2 reader refuses a 3.3 document with `{{1}}` at load (CheckText), only the plain-path render case | low | Direct doc sentence | patch |
| 9 | blind | "never rounded" wording hides that division/avg results carry padded, rounded scale (`1.5000`, `0.3333`) | low | The printer doesn't round, but division does; a one-sentence doc clarification | patch |
| 10 | blind | Acceptance evidence is a deleted probe, not the real statement fixture | low | The statement template isn't in the repo; TestNoDataPreviewPrintsNumberStandIns and the render test cover the shape | reject |
| 11 | blind | Each placeholder parsed twice on save | low | Save-time only; negligible cost | reject |
| 12 | blind | Awkward folio-format SupportedVersion parenthetical; repeated kind in the error text; stale table_render.go comments ~L1226/~L1484 | low | Prose and comments; direct corrections (error text left as is) | patch |
| 13 | edge | Boolean error message changed although the matrix says "existing message" | low | The message still contains "not a string" and names the kind and element; no caller matches on the text (designer grep: none) | reject |
| 14 | blind | Spec status/triage log stale | low | The fix would edit this build's spec | reject |
