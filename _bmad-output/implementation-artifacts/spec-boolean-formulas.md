---
title: Boolean formulas across existing expression inputs
type: feature
created: '2026-09-12'
status: done
review_loop_iteration: 0
baseline_commit: ae6cd70530b6059d8e66db448fb432533dd0df94
context:
  - '{project-root}/_bmad-output/specs/spec-boolean-formulas/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-boolean-formulas/formula-language.md'
  - '{project-root}/_bmad-output/specs/spec-boolean-formulas/brownfield.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Visibility rejects `loanAmount > 20000`.

**Approach:** Implement CAP-1–CAP-5 and the companion acceptance matrix across Go evaluation, editing, persistence and rendering.

## Boundaries & Constraints

**Always:** Strict boolean/null conditions; literal `true`, `false`, `null`; eight functions; exact bounded Decimals; static checking of both branches and lazy runtime selection; atomic commands; document-scope visibility; architecture guards.

**Ask First:** Changes to confirmed scope or decimal semantics.

**Never:** Formula-backed fixed booleans, new scopes/functions/operators beyond the contract, JavaScript evaluation, floating-point arithmetic, migration or another major version. This joins unreleased `2.0`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
| --- | --- | --- | --- |
| Threshold | Amount 25000 / 20000 / 19999 | Show / hide / hide | None |
| Null | JSON null / expression `null` | Visible / hidden | None |
| Lazy branch | `true ? true : missingFlag` | True | No unselected runtime error |
| Invalid branch | `false ? upper(1) : "ok"` | Reject | Located static error |
| Arithmetic | Zero divisor, overflow, exhausted limits | Fail | Located error; no partial command |
| Empty average | Compare empty `avg(items.amount)` using `!= 0` / `!= null` | True / false | One warning even when hidden |

</frozen-after-approval>

## Code Map

Go paths are relative to `folio-go/`; other paths are repository-relative.

- `internal/expr/`: parser/AST/checker/evaluator; `table.go` owns function constraints; `reduce.go` owns shared sum/average kernels. Preserve their routing and the closed Resolver interface.
- `folio_expr_validate.go`, `internal/bind/condition.go`: duplicate literal bans.
- `render_visibility.go` → `render.go`: currently discards caveats; preserve closed computation inputs.
- `component_commands.go`: clone/reparse/install; generic errors lose expression detail.
- `parameter_references.go`, `stand_in_data.go`, `internal/expr/footer.go`: AST walkers; footer derivation permits exactly two existing shapes.
- `serialize_template.go` → `internal/template/{serialize,version}.go`: derive expression requirements above template's dependency rank.
- `folio-designer/src/App.tsx`: `visibilityField`, `fxHint`, `conditionalContent`; commands/protocol/worker already carry opaque formulas and structured errors.

## Tasks & Acceptance

**Execution:**

- [x] `internal/expr/{ast,parser,check,table,eval,decimal,aggregate}.go` — add located nodes, precedence, literals, kind-aware checks, arithmetic and shared evaluation limits; retain function kernels and propagate selected-path caveats.
- [x] `internal/expr/*_test.go` — cover the source matrix and design rules, including discriminating precedence/association cases and all bounds.
- [x] `folio_expr_validate.go`, `internal/bind/{condition,text}.go`, `component_commands.go`, `render_visibility.go`, `render.go` — integrate checking/budgets, preserve typed causes/field context and return condition diagnostics once before body layout.
- [x] `parameter_references.go`, `stand_in_data.go`, `internal/expr/footer.go` — visit every child, implement conservative preview, retain footer restrictions and atomic alias-edit refusal.
- [x] `serialize_template.go`, `internal/template/{serialize,version}.go` — pass an AST-derived minimum version into serialization without mutating document state; never lower loaded versions.
- [x] `folio-designer/src/App.tsx`, `docs/expression-reference.{md,html}`, `_bmad-output/specs/spec-folio/folio-format.md` — update examples, grammar/version rules and accurate preview messaging; keep fixed boolean controls.
- [x] Root Go tests, `wasm/*_test.go`, `folio-designer/src/*test*`, new `folio-designer/e2e/boolean-formulas.spec.ts` — prove integration, atomicity and real worker/native persistence/render parity.

**Acceptance Criteria:**

- Given the complete source matrix, when exercised through engine consumers, then values, scales, errors and caveats match the contract.
- Given authored formulas, when committed, undone/redone and saved/reloaded, then text persists and real worker/native PDFs match for both threshold outcomes.
- Given invalid formulas, when loaded or committed, then field/element/cause and available token offset survive, with no partial mutation.
- Given formulas in all containers, when saved, then extension syntax requires `2.0`; quoted punctuation and ordinary paths do not falsely trigger it.

## Spec Change Log

## Design Notes

Use source-relative UTF-8 byte offsets. Reuse `EXPRESSION_INVALID` and `BINDING_PATH_ABSENT`. Group nodes are transparent to argument checks, but preserve footer syntax. Reject repeated unparenthesized comparisons at the same precedence level. Infer nullable function results conservatively; check each conditional branch against its consumer.

For `+`, `-`, `%`, result exponent is the minimum operand exponent; multiplication adds exponents; unary signs preserve scale. Do not normalize to bypass bounds. Keep direct MinInt64 literals valid; negating MinInt64 fails. Division follows confirmed Q2 exactly; derive rounding direction from operand signs, including `-3/50000 = -0.0001`.

Limit sources to 64 KiB, ASTs to 4096 nodes/depth 64; preflight malformed/cyclic caller-built trees. Validate Decimals before shifts. Share 1,000,000 work units per expression across recursion and concrete `bind.exprResolver`: charge nodes, processed string bytes, projection rows × path segments, and decimal shift digits before allocation/work. Preflight aggregate alignment without changing kernels or avg's divisor. Third-party resolvers own their external work bounds. The existing Visibility editor/projection field limit remains 512 bytes, separate from the engine parser’s 64 KiB work bound; valid over-limit editor input is refused atomically with field/limit context, while invalid syntax is validated first so its located cause survives.

Preview uses demand intersection: literals need no data; numbers default to zero, direct divisors prefer one. Infer `!=` from known kinds, otherwise allow compatible null. Traverse both branches. Conflicts/computed zero divisors refuse preview with sample-data guidance; valid commits remain accepted. Never promise generated data makes conditions true.

## Verification

- In `folio-go`: `rtk go test -count=1 ./...`.
- In `lint`: `rtk go test -count=1 ./...`; preserve architecture guards.
- In `folio-designer`: `rtk npm test`, `rtk npm run typecheck`, `rtk npm run lint`, `rtk npm run test:e2e:compile`, `rtk npm run test:e2e -- e2e/boolean-formulas.spec.ts`.
- Record any reproduced pre-existing `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` failure separately; do not weaken its corpus guard.


### Implementation verification (2026-09-12)

- Root Go suite: **2466 passed, 2 failed, 5 skipped** across 18 packages. The only failures are the pre-existing `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` floor and its parent (`got 7, need >=20`); the corpus guard is unchanged. Full log: `/Users/panitw/Library/Application Support/rtk/tee/1789192995_go_test.log`.
- Architecture suite (`lint`): **227 passed** across four packages.
- Frontend unit tests: **1515 passed** in 80 files. Typecheck, lint and e2e compilation passed; lint retains eight existing Fast Refresh warnings.
- Real js/wasm host test `TestWasmFormulaLongSyntaxCauseSurvivesWireBound`: passed using Go's pinned `go_js_wasm_exec`. A 600-space prefix retains `EXPRESSION_INVALID`, element `e1`, field `visibleIf`, cause and byte position 612 within the bounded wire message.
- Real browser command: **1 passed** (6.7-second test, 2.6-minute cold-build run). The test covers literal no-data worker previews (true differs from false; false/null match), accurate notice, authoring/history/atomic refusal, exact save/reload and native/worker PDF byte parity for 25000 and 20000 with distinct visibility outcomes. Report: `folio-designer/playwright-report/index.html`; saved template, data and native PDFs: `folio-designer/test-results/boolean-formulas-authored--d2692-ce-and-native-render-parity/boolean-formula-parity/`.
- Browser witness was corrected for existing UI behavior: undo/redo clears selection, and loading a second sample uses “Replace sample JSON”. No product change was needed for either.
- `git diff --check`: passed. No requested implementation work remains incomplete. The existing Visibility editor/projection limit is intentionally still 512 bytes, as documented above.


### Final review-patch verification (2026-09-12)

All R1–R13 from `_bmad-output/reviews/boolean-formulas-build-review-2026-09-12.md` are implemented and covered. Corrections restore outer Unicode whitespace without losing source offsets, safely reject unsupported AST wrappers, preserve diagnostic order, reconcile preview candidates without evaluating discarded inequality branches, retain preview/consumer/placeholder context, and use a generic no-data notice with ternary-only and quoted-question-mark coverage. No approved intent or decimal semantics changed. The frozen block SHA-256 remains `29563293ff89990711705c6235335aff94cc0481c91ba2d758f8441246bb1b67`.

- Focused formula/preview/condition regressions: **122 passed** in three packages.
- Full root Go suite: **2473 passed, 2 failed, 5 skipped** in 18 packages. Only the unchanged pre-existing P6g corpus-floor subtest and parent fail (`got 7, need >=20`). Current full log: `/Users/panitw/Library/Application Support/rtk/tee/1789194854_go_test.log`.
- Full architecture suite: **227 passed** in four packages.
- Full frontend suite: **1519 passed** in 80 files. Typecheck, lint and e2e compilation passed; the same eight Fast Refresh lint warnings remain.
- Real js/wasm host: **both formula wire tests passed**. Long Visibility syntax retains its cause and byte offset; the second placeholder's identity, cause and byte offset also survive the bounded wire message.
- Production budget regression: `TestFormulaConditionProductionResolverSharesBudget` passed with 250,000 rows sharing one immutable row object. An isolated Go overlay removing only `budget: budget` from `EvaluateCondition` made this test fail with `production resolver budget was not shared: <nil>`, proving the regression detects the reviewed mutant. Workspace source was never mutated. Evidence: `/tmp/folio-boolean-formulas-r12-mutant.log`.
- Prescribed real browser rerun: **1 passed** (6.1-second test; 2.6-minute cold-build command), including no-data literal previews, accurate generic notice, history, atomic diagnostics, persistence, and native/worker PDF byte parity for both threshold outcomes. Current report: `folio-designer/playwright-report/index.html`; parity artifacts: `folio-designer/test-results/boolean-formulas-authored--d2692-ce-and-native-render-parity/boolean-formula-parity/`.
- `git diff --check` passed. No patch finding remains incomplete. The pre-existing nil-resolver aggregate panic remains separately deferred as D1; this patch batch does not change that behavior. This verification preceded the final local build commit.

## Suggested Review Order

**Language and evaluation**

- Start with Visibility: parsing, strict kinds and shared work limits meet here.
  [condition.go:38](/Users/panitw/Projects/folio/folio-go/internal/bind/condition.go:38)

- Read precedence, literal recognition and offsets in the shared parser.
  [parser.go:33](/Users/panitw/Projects/folio/folio-go/internal/expr/parser.go:33)

- Both branches are checked before runtime selects one.
  [check.go:82](/Users/panitw/Projects/folio/folio-go/internal/expr/check.go:82)

- Inspect exact arithmetic, scale retention and half-even division.
  [arithmetic.go:54](/Users/panitw/Projects/folio/folio-go/internal/expr/arithmetic.go:54)

- Only the selected branch executes and contributes caveats.
  [eval.go:193](/Users/panitw/Projects/folio/folio-go/internal/expr/eval.go:193)

- Malformed trees and expensive work fail within explicit bounds.
  [limits.go:56](/Users/panitw/Projects/folio/folio-go/internal/expr/limits.go:56)

**Editing, rendering and persistence**

- Cloned edits preserve atomicity and return the original expression cause.
  [component_commands.go:804](/Users/panitw/Projects/folio/folio-go/component_commands.go:804)

- Condition warnings retain declaration order without reordering existing body warnings.
  [render.go:3499](/Users/panitw/Projects/folio/folio-go/render.go:3499)

- Missing paths retain their code while naming the consuming field.
  [render_error.go:143](/Users/panitw/Projects/folio/folio-go/render_error.go:143)

- Expression syntax raises serialization requirements to unreleased 2.0.
  [serialize_template.go:22](/Users/panitw/Projects/folio/folio-go/serialize_template.go:22)

**Preview and existing consumers**

- Preview intersects path demands and preserves lazy execution during final checks.
  [stand_in_data.go:387](/Users/panitw/Projects/folio/folio-go/stand_in_data.go:387)

- Every expression child contributes references, including unselected branches.
  [parameter_references.go:67](/Users/panitw/Projects/folio/folio-go/parameter_references.go:67)

- Formula support preserves the two existing footer derivation shapes.
  [footer.go:135](/Users/panitw/Projects/folio/folio-go/internal/expr/footer.go:135)

**Author guidance and verification**

- The reference documents the language, decimal rules and examples.
  [expression-reference.md:1](/Users/panitw/Projects/folio/docs/expression-reference.md:1)

- Engine integration tests cover outcomes, warnings, diagnostics and atomic edits.
  [boolean_formulas_test.go:25](/Users/panitw/Projects/folio/folio-go/boolean_formulas_test.go:25)

- This production-entry test detects disconnected resolver budget wiring.
  [formula_budget_test.go:53](/Users/panitw/Projects/folio/folio-go/internal/bind/formula_budget_test.go:53)

- The real browser witness checks history, persistence and native/worker PDF parity.
  [boolean-formulas.spec.ts:60](/Users/panitw/Projects/folio/folio-designer/e2e/boolean-formulas.spec.ts:60)
