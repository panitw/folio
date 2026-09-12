**Boolean formulas specification review — 12 September 2026**

The requested operator families and core authoring path are covered. The owner has resolved property scope and division precision. The specification still has the additional integration, semantic, and acceptance gaps identified below. The owner has since resolved the release/version and legacy-checker compatibility concerns; their original findings are retained below with resolutions.

Reviewed SPEC.md, formula-language.md, brownfield.md, and .memlog.md in [_bmad-output/specs/spec-boolean-formulas/SPEC.md](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/SPEC.md), plus referenced engine, editor, and canonical-format code/documents. The configured project-context.md was absent. Independent adversarial, edge-case, and structure lenses were run; the coordinator verified source claims and added the association-test and alias-edit findings. This is a specification review, not an implementation test run. The initial review wrote only review artifacts. Subsequent owner clarifications were recorded in the canonical decision log and used to re-derive the spec and companions; production code has not been edited by this task.

**Decisions and questions**

- Latest owner decision supersedes the earlier A5 answer: `true`, `false`, and `null` are now literal values. The spec and companions record this revision, condition typing, null-versus-missing behavior, keyword boundaries (A8), integration obligations, and acceptance cases.
- Q1 resolved: existing condition inputs only. Bold/Italic and all other fixed boolean properties remain fixed; no new formula-backed styling controls or property schema are introduced.
- Q2 resolved: the owner accepted max(operand decimal scales, 0) + 4 fractional places with round-half-to-even. Apply the rule at each division, retain result scale, and enforce Decimal bounds. This gives `1 / 3 = 0.3333`, `1.00 / 3 = 0.333333`, and `1 / 100000 = 0.0000`; exact terminating quotients also use the selected scale. Signed, tie, exponent, tiny-result, repeated-division and overflow fixtures are now in the spec. These are specification expectations verified with decimal arithmetic, not current engine results.
- Release/version decision confirmed by the owner: no release has reached external users. New formula syntax joins unreleased 2.0. Backward compatibility with earlier accepted expressions is not required; no separate major version, migration layer, or legacy checker is needed. The spec now records this and applies its static rules uniformly to if() and ternaries.

**Coverage already present**

Boolean/null literals, all five requested comparisons, all five arithmetic operators, unary signs, ordinary precedence, grouping, simple and nested ternaries, right association, lazy selected-branch execution, strict boolean/null conditions, scalar inequality, missing-path errors, decimal-only arithmetic, zero-divisor errors, shared engine use, parameter discovery, field-specific errors, atomic commands, persistence, and real native/worker rendering are addressed. Logical operators, additional equality operators, row-level visibility, and new named functions are explicitly excluded. A requirement being mentioned does not by itself provide a discriminating acceptance test; findings below identify where the promised behavior still needs an executable decision or example.

**Adversarial findings**

**A01** — Formula documents introduce syntax older readers reject, but no required format version is assigned.

Resolution update: Resolved in the specification: the owner confirms there are no external releases. New formulas join unreleased 2.0, with content-based save requirements recorded; no new major version or migration layer is required.

Location: [_bmad-output/specs/spec-boolean-formulas/brownfield.md:30](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:30); [_bmad-output/specs/spec-boolean-formulas/SPEC.md:40](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/SPEC.md:40).

Required clarification or guard: Define the minimum .folio version for new formula syntax and update content-based version derivation. Test save/load from every expression container and preserve existing requirements for old documents. The canonical format permits minor increments only for new optional keys; decide whether formulas join the 2.0 ceiling or require another major version based on release status.

Consequence: Saved files can advertise compatibility with readers that cannot parse their contents.

Source evidence: [_bmad-output/specs/spec-folio/folio-format.md:47](/Users/panitw/Projects/folio/_bmad-output/specs/spec-folio/folio-format.md:47), [_bmad-output/specs/spec-folio/folio-format.md:67](/Users/panitw/Projects/folio/_bmad-output/specs/spec-folio/folio-format.md:67), [folio-go/internal/template/version.go:247](/Users/panitw/Projects/folio/folio-go/internal/template/version.go:247).

**A02** — An evaluated operand emits a caveat while its comparison returns a boolean: avg(items.amount) != 0 on an empty collection.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:52](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:52); [_bmad-output/specs/spec-boolean-formulas/brownfield.md:9](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:9).

Required clarification or guard: Carry caveats from evaluated operands, conditions, and selected branches through Visibility and other consumers; unselected branches contribute none. Require exactly one empty-average diagnostic through native and worker rendering, with explicit behavior when the condition hides the element.

Consequence: Visibility can depend on an empty average while losing the diagnostic explaining the result.

Source evidence: [folio-go/render_visibility.go:109](/Users/panitw/Projects/folio/folio-go/render_visibility.go:109), [folio-go/internal/expr/aggregate.go:170](/Users/panitw/Projects/folio/folio-go/internal/expr/aggregate.go:170).

**A03** — Broader static inference is applied to existing lazy if() expressions without a compatibility boundary.

Resolution update: Resolved in the specification: backward compatibility with permissive legacy static checking is not required. The same static checks now apply to if() and ternary branches, with explicit load/commit rejection examples and unchanged runtime branch laziness.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:53](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:53); [_bmad-output/specs/spec-boolean-formulas/SPEC.md:40](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/SPEC.md:40).

Required clarification or guard: Specify whether new static type inference is restricted to new syntax or tightens legacy if(). Preserve or explicitly version examples such as if(flag, upper(1), "ok") with flag=false and Visibility if(flag, 1, fallback) with a boolean fallback. State the expected load/commit versus evaluation phase for each.

Consequence: Previously loadable documents that render successfully through a selected branch can start failing at load.

Source evidence: [folio-go/internal/expr/check.go:59](/Users/panitw/Projects/folio/folio-go/internal/expr/check.go:59), [folio-go/internal/expr/table.go:99](/Users/panitw/Projects/folio/folio-go/internal/expr/table.go:99), [folio-go/internal/expr/eval.go:148](/Users/panitw/Projects/folio/folio-go/internal/expr/eval.go:148).

**A04** — An exact arithmetic result has multiple representations or fits the coefficient bound only after normalization.

Resolution update: Q2 division precision is now confirmed, including retained result scale and coefficient bounds. This finding remains open for exact result scales and narrowing rules of addition, subtraction, multiplication, remainder and unary signs.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:45](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:45).

Required clarification or guard: Define result scale and narrowing for +, -, *, %, and unary signs independently of Q2. Specify the scale of 1.50 * 2.0 and cancellation results, and whether 9000000000000000000 + 1000000000000000000 errors or normalizes to 1e19. Reference exact coefficient, exponent, and alignment bounds.

Consequence: Implementations disagree on formatting, subsequent division precision, and overflow despite computing exact values.

Source evidence: [folio-go/internal/expr/decimal.go:18](/Users/panitw/Projects/folio/folio-go/internal/expr/decimal.go:18), [folio-go/internal/expr/reduce.go:66](/Users/panitw/Projects/folio/folio-go/internal/expr/reduce.go:66).

**A05** — A formula contains a long operator chain, deeply nested expressions, or repeated aggregate evaluation.

Resolution update: The compatibility-waiver removes the legacy-document reconciliation requirement from this finding; numeric source, depth, node, and work limits remain open.

Location: [_bmad-output/specs/spec-boolean-formulas/SPEC.md:44](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/SPEC.md:44); [_bmad-output/specs/spec-boolean-formulas/formula-language.md:108](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:108).

Required clarification or guard: Name source-byte, node/depth, and evaluation-work limits; state whether each applies per expression or per document and what consumes the budget. Test below, at, and above each boundary, including chains that parse iteratively but form deep ASTs. Reconcile new limits with legacy-document compatibility.

Consequence: Arbitrary rejection can satisfy the current acceptance row while expensive work or unsupported limits remain undefined.

**A06** — No-data preview encounters loanAmount > 20000 or amount / divisor > 1.

Location: [_bmad-output/specs/spec-boolean-formulas/brownfield.md:12](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:12); [_bmad-output/specs/spec-boolean-formulas/formula-language.md:108](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:108).

Required clarification or guard: Choose deterministic condition-satisfying values, safe typed values with disclosed hidden elements, or an explicit preview-only refusal policy. Specify nonzero-divisor handling, incompatible uses of a path, unsatisfied conditions, and the resulting editor display. Replace the catch-all acceptance outcome with specific expected results.

Consequence: The motivating formula can work with real data but hide its element or fail preview under an unspecified policy.

**A07** — Grouping wraps a syntactically constrained argument, such as formatNumber(amount, ("0.00")) or sum((items.amount)).

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:23](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:23); [_bmad-output/specs/spec-boolean-formulas/formula-language.md:33](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:33).

Required clarification or guard: Define grouping transparency for literal-only patterns and path-only aggregate operands, while keeping computed replacements subject to existing restrictions. Give grouped row-path and footer-derivation acceptance cases without silently widening the two existing footer shapes.

Consequence: Equivalent-looking grouped expressions can unexpectedly fail or bypass function and footer restrictions.

Source evidence: [folio-go/internal/expr/check.go:151](/Users/panitw/Projects/folio/folio-go/internal/expr/check.go:151), [folio-go/internal/expr/footer.go:125](/Users/panitw/Projects/folio/folio-go/internal/expr/footer.go:125).

**A08** — An author writes a != b != c or amount > limit != override.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:27](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:27); [_bmad-output/specs/spec-boolean-formulas/formula-language.md:33](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:33).

Required clarification or guard: Define whether repeated unparenthesized comparisons are syntax errors and, if allowed, their association. Distinguish numeric chains from typed boolean comparisons; give acceptance cases for a != b != c, (a != b) != c, and amount > limit != override.

Consequence: The phrase 'no implicit chained-comparison semantics' allows different accepted grammars and error phases.

**A09** — Arithmetic or resource failures cross the native or worker boundary to the editor.

Location: [_bmad-output/specs/spec-boolean-formulas/brownfield.md:24](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:24); [_bmad-output/specs/spec-boolean-formulas/formula-language.md:71](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:71).

Required clarification or guard: Map syntax/static-type, runtime-type, zero-divisor, numeric-bound, and resource-bound failures to stable diagnostic codes, reusing existing codes when appropriate. Define required element/field identity and source-offset conventions and assert the payload received by the worker client.

Consequence: An internally located error can still surface as the generic format failure that motivated this change.

Source evidence: [folio-go/render_error.go:10](/Users/panitw/Projects/folio/folio-go/render_error.go:10).

**A10** — Unary parsing splits -9223372036854775808 into an out-of-range positive literal and a minus node.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:24](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:24); [_bmad-output/specs/spec-boolean-formulas/formula-language.md:31](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:31).

Required clarification or guard: Add signed-boundary acceptance: preserve -9223372036854775808, reject positive 9223372036854775808, and report overflow when negating a MinInt64 coefficient. Cover unary chains and scientific notation without requiring an invalid positive intermediate.

Consequence: The parser can reject an existing valid literal or negation can wrap despite the preservation requirement.

Source evidence: [folio-go/internal/expr/decimal.go:116](/Users/panitw/Projects/folio/folio-go/internal/expr/decimal.go:116).

**A11** — The acceptance matrix is used to verify operator association or ungrouped arithmetic precedence.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:67](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:67); [_bmad-output/specs/spec-boolean-formulas/formula-language.md:68](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:68).

Required clarification or guard: Use discriminating predicates: 10 - 3 - 2 < 6 is true only with left association; 12 / 3 / 2 < 3 is true only with left association; 2 + 3 * 4 < 15 distinguishes multiplicative precedence from left-to-right evaluation. Retain the grouped case as a separate check.

Consequence: A parser with incorrect association can pass both existing association examples.

**A12** — A table alias is edited while a column contains a formula referencing the old alias.

Location: [_bmad-output/specs/spec-boolean-formulas/brownfield.md:13](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:13); [_bmad-output/specs/spec-boolean-formulas/brownfield.md:18](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:18).

Required clarification or guard: Explicitly include expressionUsesRoot, RewriteRowBinding, and ProjectRowBinding in the integration inventory. Preserve the current atomic refusal for unsupported alias migrations, unless expanded deliberately. Test formatNumber(row.amount + row.fee, "0.00") and old-alias references inside either ternary branch; complex formulas must remain protected from destructive row-field editing.

Consequence: An alias change can miss references under new nodes and leave a previously valid column binding broken.

Source evidence: [folio-go/internal/expr/footer.go:40](/Users/panitw/Projects/folio/folio-go/internal/expr/footer.go:40), [folio-go/internal/expr/footer.go:70](/Users/panitw/Projects/folio/folio-go/internal/expr/footer.go:70), [folio-go/component_commands.go:557](/Users/panitw/Projects/folio/folio-go/component_commands.go:557).

**Edge-case findings**

These four findings independently overlap A04 (decimal representation), A06 (preview), A02 (caveats), and A07 (grouping). They are retained to show agreement between lenses rather than counted as four additional distinct omissions.

**E01** — An exact result fits Decimal bounds only after removing trailing zeros.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:45](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:45).

Guard: Specify result exponent and trailing-zero rules, including whether normalization precedes overflow checks.

Consequence: Implementations disagree on overflow and subsequent division precision for mathematically identical results.

**E02** — No sample data exists for loanAmount > 20000 or amount / divisor > 1.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:108](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:108).

Guard: Define deterministic stand-ins or preview-only refusal for comparisons, divisors, and ternaries, including editor display.

Consequence: A valid formula prevents preview or hides components without a specified authoring outcome.

**E03** — An evaluated operand or selected branch produces an empty-average caveat.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:52](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:52).

Guard: Propagate caveats from every evaluated operand, condition, and selected branch in evaluation order.

Consequence: A rendered result silently loses the existing empty-average warning.

**E04** — Grouping wraps a collection-path argument or a required literal formatting pattern.

Location: [_bmad-output/specs/spec-boolean-formulas/formula-language.md:33](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:33).

Guard: Make grouping transparent to existing argument restrictions; computed replacements remain rejected.

Consequence: Valid grouped calls fail, or grouping bypasses existing function argument restrictions.

**Structure finding**

This spec exists to help implementers and reviewers agree on formula behavior, integrate it safely, and verify complete coverage. Structure model: Strategic/Context for SPEC.md; Reference/Database for its companions.

| Pass | Original text | Revised text | Changes |
| --- | --- | --- | --- |
| structure | [_bmad-output/specs/spec-boolean-formulas/brownfield.md:26](/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:26), final regression-evidence paragraph; section total: 143 words | MOVE regression-evidence requirements into formula-language.md's acceptance matrix; retain an integration cross-reference. | The language companion is identified as the complete acceptance contract, but old-document and footer regression obligations remain outside its matrix. Estimated reduction: 0 words. |

One structure recommendation. Exact combined count at the original structure review, before the owner-clarification update: 2,091 words. Estimated reduction: 0 words (0%). No length target was supplied; no comprehension trade-off is proposed.

**Raw findings**

The original findings are retained with resolution updates. The editorial finding uses its lens-specific table fields. The same array is available as [JSON](/Users/panitw/Projects/folio/_bmad-output/reviews/boolean-formulas-review-2026-09-12.json).

```json
[
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:30; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/SPEC.md:40",
    "trigger_condition": "Formula documents introduce syntax older readers reject, but no required format version is assigned.",
    "guard_snippet": "Define the minimum .folio version for new formula syntax and update content-based version derivation. Test save/load from every expression container and preserve existing requirements for old documents. The canonical format permits minor increments only for new optional keys; decide whether formulas join the 2.0 ceiling or require another major version based on release status.",
    "potential_consequence": "Saved files can advertise compatibility with readers that cannot parse their contents.",
    "evidence": [
      "/Users/panitw/Projects/folio/_bmad-output/specs/spec-folio/folio-format.md:47",
      "/Users/panitw/Projects/folio/_bmad-output/specs/spec-folio/folio-format.md:67",
      "/Users/panitw/Projects/folio/folio-go/internal/template/version.go:247"
    ],
    "resolution": "Resolved in the specification: the owner confirms there are no external releases. New formulas join unreleased 2.0, with content-based save requirements recorded; no new major version or migration layer is required."
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:52; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:9",
    "trigger_condition": "An evaluated operand emits a caveat while its comparison returns a boolean: avg(items.amount) != 0 on an empty collection.",
    "guard_snippet": "Carry caveats from evaluated operands, conditions, and selected branches through Visibility and other consumers; unselected branches contribute none. Require exactly one empty-average diagnostic through native and worker rendering, with explicit behavior when the condition hides the element.",
    "potential_consequence": "Visibility can depend on an empty average while losing the diagnostic explaining the result.",
    "evidence": [
      "/Users/panitw/Projects/folio/folio-go/render_visibility.go:109",
      "/Users/panitw/Projects/folio/folio-go/internal/expr/aggregate.go:170"
    ]
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:53; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/SPEC.md:40",
    "trigger_condition": "Broader static inference is applied to existing lazy if() expressions without a compatibility boundary.",
    "guard_snippet": "Specify whether new static type inference is restricted to new syntax or tightens legacy if(). Preserve or explicitly version examples such as if(flag, upper(1), \"ok\") with flag=false and Visibility if(flag, 1, fallback) with a boolean fallback. State the expected load/commit versus evaluation phase for each.",
    "potential_consequence": "Previously loadable documents that render successfully through a selected branch can start failing at load.",
    "evidence": [
      "/Users/panitw/Projects/folio/folio-go/internal/expr/check.go:59",
      "/Users/panitw/Projects/folio/folio-go/internal/expr/table.go:99",
      "/Users/panitw/Projects/folio/folio-go/internal/expr/eval.go:148"
    ],
    "resolution": "Resolved in the specification: backward compatibility with permissive legacy static checking is not required. The same static checks now apply to if() and ternary branches, with explicit load/commit rejection examples and unchanged runtime branch laziness."
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:45",
    "trigger_condition": "An exact arithmetic result has multiple representations or fits the coefficient bound only after normalization.",
    "guard_snippet": "Define result scale and narrowing for +, -, *, %, and unary signs independently of Q2. Specify the scale of 1.50 * 2.0 and cancellation results, and whether 9000000000000000000 + 1000000000000000000 errors or normalizes to 1e19. Reference exact coefficient, exponent, and alignment bounds.",
    "potential_consequence": "Implementations disagree on formatting, subsequent division precision, and overflow despite computing exact values.",
    "evidence": [
      "/Users/panitw/Projects/folio/folio-go/internal/expr/decimal.go:18",
      "/Users/panitw/Projects/folio/folio-go/internal/expr/reduce.go:66"
    ],
    "resolution": "Q2 division precision is now confirmed, including retained result scale and coefficient bounds. This finding remains open for exact result scales and narrowing rules of addition, subtraction, multiplication, remainder and unary signs."
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/SPEC.md:44; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:108",
    "trigger_condition": "A formula contains a long operator chain, deeply nested expressions, or repeated aggregate evaluation.",
    "guard_snippet": "Name source-byte, node/depth, and evaluation-work limits; state whether each applies per expression or per document and what consumes the budget. Test below, at, and above each boundary, including chains that parse iteratively but form deep ASTs. Reconcile new limits with legacy-document compatibility.",
    "potential_consequence": "Arbitrary rejection can satisfy the current acceptance row while expensive work or unsupported limits remain undefined.",
    "resolution": "The compatibility-waiver removes the legacy-document reconciliation requirement from this finding; numeric source, depth, node, and work limits remain open."
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:12; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:108",
    "trigger_condition": "No-data preview encounters loanAmount > 20000 or amount / divisor > 1.",
    "guard_snippet": "Choose deterministic condition-satisfying values, safe typed values with disclosed hidden elements, or an explicit preview-only refusal policy. Specify nonzero-divisor handling, incompatible uses of a path, unsatisfied conditions, and the resulting editor display. Replace the catch-all acceptance outcome with specific expected results.",
    "potential_consequence": "The motivating formula can work with real data but hide its element or fail preview under an unspecified policy."
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:23; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:33",
    "trigger_condition": "Grouping wraps a syntactically constrained argument, such as formatNumber(amount, (\"0.00\")) or sum((items.amount)).",
    "guard_snippet": "Define grouping transparency for literal-only patterns and path-only aggregate operands, while keeping computed replacements subject to existing restrictions. Give grouped row-path and footer-derivation acceptance cases without silently widening the two existing footer shapes.",
    "potential_consequence": "Equivalent-looking grouped expressions can unexpectedly fail or bypass function and footer restrictions.",
    "evidence": [
      "/Users/panitw/Projects/folio/folio-go/internal/expr/check.go:151",
      "/Users/panitw/Projects/folio/folio-go/internal/expr/footer.go:125"
    ]
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:27; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:33",
    "trigger_condition": "An author writes a != b != c or amount > limit != override.",
    "guard_snippet": "Define whether repeated unparenthesized comparisons are syntax errors and, if allowed, their association. Distinguish numeric chains from typed boolean comparisons; give acceptance cases for a != b != c, (a != b) != c, and amount > limit != override.",
    "potential_consequence": "The phrase 'no implicit chained-comparison semantics' allows different accepted grammars and error phases."
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:24; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:71",
    "trigger_condition": "Arithmetic or resource failures cross the native or worker boundary to the editor.",
    "guard_snippet": "Map syntax/static-type, runtime-type, zero-divisor, numeric-bound, and resource-bound failures to stable diagnostic codes, reusing existing codes when appropriate. Define required element/field identity and source-offset conventions and assert the payload received by the worker client.",
    "potential_consequence": "An internally located error can still surface as the generic format failure that motivated this change.",
    "evidence": [
      "/Users/panitw/Projects/folio/folio-go/render_error.go:10"
    ]
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:24; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:31",
    "trigger_condition": "Unary parsing splits -9223372036854775808 into an out-of-range positive literal and a minus node.",
    "guard_snippet": "Add signed-boundary acceptance: preserve -9223372036854775808, reject positive 9223372036854775808, and report overflow when negating a MinInt64 coefficient. Cover unary chains and scientific notation without requiring an invalid positive intermediate.",
    "potential_consequence": "The parser can reject an existing valid literal or negation can wrap despite the preservation requirement.",
    "evidence": [
      "/Users/panitw/Projects/folio/folio-go/internal/expr/decimal.go:116"
    ]
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:67; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:68",
    "trigger_condition": "The acceptance matrix is used to verify operator association or ungrouped arithmetic precedence.",
    "guard_snippet": "Use discriminating predicates: 10 - 3 - 2 < 6 is true only with left association; 12 / 3 / 2 < 3 is true only with left association; 2 + 3 * 4 < 15 distinguishes multiplicative precedence from left-to-right evaluation. Retain the grouped case as a separate check.",
    "potential_consequence": "A parser with incorrect association can pass both existing association examples."
  },
  {
    "lens": "adversarial",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:13; /Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/brownfield.md:18",
    "trigger_condition": "A table alias is edited while a column contains a formula referencing the old alias.",
    "guard_snippet": "Explicitly include expressionUsesRoot, RewriteRowBinding, and ProjectRowBinding in the integration inventory. Preserve the current atomic refusal for unsupported alias migrations, unless expanded deliberately. Test formatNumber(row.amount + row.fee, \"0.00\") and old-alias references inside either ternary branch; complex formulas must remain protected from destructive row-field editing.",
    "potential_consequence": "An alias change can miss references under new nodes and leave a previously valid column binding broken.",
    "evidence": [
      "/Users/panitw/Projects/folio/folio-go/internal/expr/footer.go:40",
      "/Users/panitw/Projects/folio/folio-go/internal/expr/footer.go:70",
      "/Users/panitw/Projects/folio/folio-go/component_commands.go:557"
    ]
  },
  {
    "lens": "edge-case-hunter",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:45",
    "trigger_condition": "An exact result fits Decimal bounds only after removing trailing zeros.",
    "guard_snippet": "Specify result exponent and trailing-zero rules, including whether normalization precedes overflow checks.",
    "potential_consequence": "Implementations disagree on overflow and subsequent division precision for mathematically identical results."
  },
  {
    "lens": "edge-case-hunter",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:108",
    "trigger_condition": "No sample data exists for loanAmount > 20000 or amount / divisor > 1.",
    "guard_snippet": "Define deterministic stand-ins or preview-only refusal for comparisons, divisors, and ternaries, including editor display.",
    "potential_consequence": "A valid formula prevents preview or hides components without a specified authoring outcome."
  },
  {
    "lens": "edge-case-hunter",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:52",
    "trigger_condition": "An evaluated operand or selected branch produces an empty-average caveat.",
    "guard_snippet": "Propagate caveats from every evaluated operand, condition, and selected branch in evaluation order.",
    "potential_consequence": "A rendered result silently loses the existing empty-average warning."
  },
  {
    "lens": "edge-case-hunter",
    "location": "/Users/panitw/Projects/folio/_bmad-output/specs/spec-boolean-formulas/formula-language.md:33",
    "trigger_condition": "Grouping wraps a collection-path argument or a required literal formatting pattern.",
    "guard_snippet": "Make grouping transparent to existing argument restrictions; computed replacements remain rejected.",
    "potential_consequence": "Valid grouped calls fail, or grouping bypasses existing function argument restrictions."
  },
  {
    "lens": "structure",
    "Pass": "structure",
    "Original Text": "brownfield.md §Obligations, final regression-evidence paragraph (line 23); section total: 143 words.",
    "Revised Text": "MOVE regression-evidence requirements into formula-language.md §Acceptance matrix; retain a cross-reference under integration obligations.",
    "Changes": "SPEC.md calls the language companion the complete acceptance contract, but existing-document output and footer-derivation coverage remain outside its matrix. Reorganization only; estimated reduction 0 words."
  }
]
```
