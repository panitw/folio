---
id: SPEC-boolean-formulas
companions:
  - formula-language.md
  - brownfield.md
sources: []
---

# Boolean formulas

## Why

Authors cannot enter a simple condition such as `loanAmount > 20000` in Visibility: the current expression grammar rejects it with a generic format-validation error. Allow comparisons, basic arithmetic and nested conditional expressions so authors can express these rules directly.

## Capabilities

- **CAP-1**
  - **intent:** Authors compare values to decide whether a condition holds.
  - **success:** `>`, `<`, `>=`, `<=` and `!=` produce the specified boolean results, including equal thresholds and invalid-type diagnostics.
- **CAP-2**
  - **intent:** Authors calculate values within conditions using basic math.
  - **success:** `+`, `-`, `*`, `/` and `%` compose with comparisons and parentheses; precedence, decimal behavior and arithmetic failure cases match the companion contract.
- **CAP-3**
  - **intent:** Authors select values using simple and nested ternary expressions.
  - **success:** Both `x ? y : z` and `x ? (a ? b : c) : z` work; only the selected branch evaluates, and nested expressions associate as documented.
- **CAP-4**
  - **intent:** Authors use consistent formulas wherever expressions consume boolean conditions.
  - **success:** Visibility, `if()` and ternary conditions accept the same boolean/null values, including `true`, `false` and `null` literals; final numeric/string results are rejected in boolean contexts. Fixed boolean properties remain fixed under confirmed Q1.
- **CAP-5**
  - **intent:** Authors can correct invalid formulas and retain valid ones throughout document editing.
  - **success:** Errors identify the affected field and cause; valid formulas survive commit, undo/redo and save/reload, then render correctly through the worker with datasets that exercise both outcomes.

## Constraints

- The complete syntax, semantic and acceptance contract is in `formula-language.md`; integration obligations and historical rules affected by this extension are in `brownfield.md`.
- Use the shared engine language across consumers; no browser-only formula parser or general-purpose script evaluation.
- Retain the eight named functions, data scopes, strict boolean/null condition handling and exact decimal representation.
- **A1 / Q1 (confirmed):** Condition scope is limited to existing inputs: Visibility, `if()` and ternary conditions. Bold/Italic and other fixed boolean properties remain fixed.
- **Q2 (confirmed):** Division uses max(operand decimal scales, 0) + 4 fractional places with round-half-to-even, applied at each division operator. Exact cases and bounds are in `formula-language.md`.
- **Release decision (confirmed):** No release has reached external users. Formulas join the unreleased `2.0` format; backward compatibility with earlier accepted expressions is not required. No separate major version, migration layer or legacy checker is needed for this extension.
- **A5 (revised and confirmed):** `true`, `false` and `null` are literal values, with no data lookup. This replaces the earlier decision to keep these bare words as field names.
- This request supersedes the historical prohibition on operators for the listed syntax. It does not implicitly authorize data-driven style controls.
- Preserve document-scope visibility and existing row/page restrictions.
- Bound parsing/evaluation work and surface syntax, type, arithmetic and resource-limit errors without partially applying editor commands.

## Non-goals

- General scripting, assignment, new named functions, logical operators, equality operators beyond `!=`, or spreadsheet formula syntax.
- Expanding data scopes, enabling row-level visibility, or changing pagination rules.
- Formula-backed fixed boolean properties, including Bold/Italic, or new data-driven styling controls.
- Implementing production code or generating story dispatch records in this specification run.

## Success signal

Enter `loanAmount > 20000` in Visibility and successfully render the element for 25000 but omit it for 20000. Demonstrate arithmetic and nested ternaries from the acceptance matrix through the same edit/save/reload/render path without precomputed boolean data.

## Assumptions

- **A2:** The requested operator list is exhaustive; no `==`, `===`, `!==`, `&&`, `||` or `!`.
- **A3:** Ordering compares numbers; `!=` uses the scalar/null rules in the language companion.
- **A4:** Ordinary precedence, grouping parentheses and unary signs are included.
- **A6:** Ternaries return selected values without coercion. Both ternaries and `if()` check both branches statically while evaluating only one; no legacy-checker exception is required.
- **A7:** Arithmetic stays decimal and bounded; modulo follows truncated-toward-zero remainder semantics.
- **A8:** Keywords are lowercase whole identifiers. They are reserved as bare expressions and path roots; keyword text after a dot remains a field name, and quoted text remains a string.

