# Formula language and acceptance

This companion expands CAP-1–CAP-5. A1/Q1 (existing condition inputs), A5 (literals), Q2 (division precision) and the unreleased-2.0 decision are user-confirmed. A2–A4 and A6–A8 remain inferred defaults.

## Input and scope

Q1 limits condition scope to existing inputs: Visibility, `if()` and ternary conditions. Bold/Italic and other fixed boolean properties remain fixed; no formula-backed styling or new property controls are added.

Visibility accepts a bare expression: `loanAmount > 20000`, without a leading `=` or `{{ }}`. Empty Visibility remains always visible. Existing expression wrappers in text remain unchanged.

The same expression syntax composes inside existing `if(condition, thenValue, elseValue)` calls and ternary conditions. Because the parser is shared, operators may also appear within existing text/value expression containers; their existing output-type and formatting rules still apply. Numeric results do not become strings automatically.

Final boolean conditions accept booleans and null: `true` is true; `false` and `null` are false. Literal null and a resolved JSON null have identical value semantics. Strings and numbers do not acquire truthiness. Arithmetic may therefore appear as an operand in `loanAmount + fee > 20000`, but `loanAmount + fee` alone is not a valid boolean result.

An absent Visibility field or JSON `"visibleIf": null` means no condition and keeps the element visible. JSON `"visibleIf": "null"` contains the null literal expression and hides it. Clearing the editor field still removes the condition.

## Syntax and precedence

Highest precedence first:

| Level | Syntax | Association |
| --- | --- | --- |
| Grouping / atoms | `(expr)`, data paths, string/number literals, `true`, `false`, `null`, existing function calls | Existing call rules |
| Unary signs | `+expr`, `-expr` | Right |
| Multiplicative | `*`, `/`, `%` | Left |
| Additive | `+`, `-` | Left |
| Ordering | `>`, `<`, `>=`, `<=` | No implicit chained-comparison semantics |
| Inequality | `!=` | No implicit chained-comparison semantics |
| Conditional | `condition ? thenExpr : elseExpr` | Right |

Recognize multi-character operators before their single-character prefixes. Whitespace is optional between unambiguous tokens. Preserve signed-number and scientific-notation parsing, dotted paths, function commas, and punctuation inside quoted strings.

`x ? y : a ? b : c` means `x ? y : (a ? b : c)`. Both requested forms `x ? y : z` and `x ? (a ? b : c) : z` are supported. Parentheses may group any expression. A chained numeric comparison such as `1 < amount < 10` must not silently acquire range semantics or boolean-to-number coercion.

Under revised A5, `true`, `false` and `null` are literal atoms, usable directly, inside parentheses, in comparisons, and in `if()` or ternary conditions and branches. They do not resolve against data.

Under A8, keywords match lowercase whole identifiers only. `trueFlag`, `falseValue`, `nullable` and `True` remain identifiers; `"true"` is a string. Bare `true`, `false` and `null` cannot name root data fields or functions: `true.field` and `true()` are syntax errors. Segments after a dot remain field names, so `record.true` and `params.null` are paths within existing scopes. No escaping syntax or additional data namespace is introduced.

## Value semantics

- Boolean literals produce boolean values and `null` produces the explicit null value; none invokes the resolver or creates a parameter/data reference.
- Ordering requires two numeric Decimals and compares their mathematical values; decimal scale does not change comparison results. Boolean/null operands are invalid.
- `!=` compares same-kind scalar values: numeric mathematical value, exact string value, or boolean value. `null != null` is false; null versus a non-null scalar is true. Non-null mixed types and collections/objects error. Missing paths retain existing resolution errors: `customer.middleName != null` is false for a present null and true for a present non-null scalar, including an empty string, but errors if the path is missing.
- `+`, `-`, `*`, `/`, `%` and unary signs accept numbers only. Boolean/null operands, string concatenation and numeric-string coercion are invalid.
- Addition, subtraction and multiplication use exact decimal arithmetic within explicit result/resource bounds; no binary floating-point intermediates.
- Remainder is `a - trunc(a / b) * b`, computed exactly without using rounded division. It carries the dividend's sign: `-5 % 2` is -1 and `5 % -2` is 1.
- Division by zero, remainder by zero, overflow or exceeded bounds produce located errors; no NaN, infinity or silent false.
- **Q2 confirmed:** For evaluated operand Decimals with exponents `ea` and `eb`, division chooses `S = max(-ea, -eb, 0) + 4` fractional places. Compute the exact quotient and round once at that scale using round-half-to-even; no binary floating-point intermediate. Equivalently, round `(a / b) * 10^S` to the nearest integer coefficient, choosing the even coefficient at a tie, then return exponent `-S`.
- Apply that rule at each division operator, including exact terminating quotients. Preserve operand scales when choosing `S` and retain the result's scale, including trailing zeros; do not normalize to change subsequent precision or avoid coefficient bounds. For example, `1 / 3` gives `0.3333`, whereas `1.00 / 3` gives `0.333333`.
- Tiny nonzero quotients may round to zero: `1 / 100000` gives `0.0000`. Positive and negative results rounded to zero use the existing zero coefficient with exponent `-S`; there is no signed-zero value. This specified rounding is not an arithmetic failure.
- Division retains Decimal's `int64` coefficient and exponent bounds. Check the rounded result and intermediate resource requirements; a coefficient that does not fit at the required result scale is an error, even if removing trailing zeros would make it fit. Zero divisors remain errors.
- A ternary checks its condition using existing boolean/null semantics and returns the selected branch's value unchanged. Only that branch is resolved/evaluated; unselected branches cannot emit runtime errors or caveats.
- Parse and statically validate both branches of ternaries and `if()`. Invalid syntax, unknown functions or statically provable type misuse must not hide in an unselected branch. Runtime-dependent type failures arise only on the evaluated path. These rules apply uniformly without preserving the earlier checker's accepted forms.
- Branches need not have identical kinds in general value contexts; the selected result must satisfy its consuming field's type. Reject statically provable invalid boolean outcomes where possible.

## Acceptance matrix

| Capability | Example / action | Expected |
| --- | --- | --- |
| CAP-1, CAP-5 | `loanAmount > 20000`; amounts 25000, 20000, 19999 | Visible, hidden, hidden |
| CAP-1 | Repeat boundary cases for `<`, `>=`, `<=`, `!=` | Correct below/equal/above outcomes for each |
| CAP-1 | `1.0 != 1.00` | False |
| CAP-1 | String compared with number using ordering or `!=` | Located type error |
| CAP-1 | `true != false`; `null != null`; `null != true`; `null != 0`; `"null" != null` | True, false, true, true, true |
| CAP-1 | `true != 1`; `false < true` | Static type errors; no boolean-to-number coercion |
| CAP-1 | `customer.middleName != null` with null, empty string, non-empty string, missing path | False, true, true, located missing-path error |
| CAP-2 | `2 + 3 * 4 > 10`; `(2 + 3) * 4 > 19` | Both true |
| CAP-2 | `10 - 3 - 2 > 4`; `12 / 3 / 2 > 1` | Both true; left association |
| CAP-2 | `5 % 2 != 0`; `-5 % 2 < 0`; `5 % -2 > 0` | All true |
| CAP-2 | `0.1 + 0.2 >= 0.3` | True with decimal arithmetic |
| CAP-2 | Division/remainder by zero, result overflow | Located errors |
| CAP-2 | `1 / 3`; `2 / 3`; `1.00 / 3`; `1 / 8` | Decimals `0.3333`, `0.6667`, `0.333333`, `0.1250` with the shown scales |
| CAP-2 | Half-even ties: `1 / 32`; `3 / 32`; `-1 / 32`; `-3 / 32`; `1 / -32` | `0.0312`, `0.0938`, `-0.0312`, `-0.0938`, `-0.0312` |
| CAP-2 | `1e3 / 3`; `1e-3 / 3` | `333.3333`; `0.0003333`, using the evaluated operands' Decimal exponents |
| CAP-2 | `1 / 100000`; `-1 / 100000`; `1 / 100000 > 0` | Both numeric results are coefficient 0, exponent -4; comparison false |
| CAP-2 | `1 / 3 > 0.33332`; `1.00 / 3 > 0.33332` | False; true, reflecting accepted operand-scale-dependent rounding |
| CAP-2 | `12 / 3 / 2` | `2.00000000`: first division returns `4.0000`, then the next chooses scale 8 |
| CAP-2 | `1000000000000000 / 1` | Located coefficient-bound error: required scale 4 needs coefficient `10000000000000000000`, outside int64 |
| CAP-2 | `true + 1`; `null * 2`; `-false` | Static type errors |
| CAP-3 | `x ? y : z`, boolean data | Selected branch determines result |
| CAP-3, CAP-4 | `true ? true : false`; `if(false, true, false)`; `null ? true : false`; `if(null, true, false)` | True, false, false, false |
| CAP-3, CAP-4 | `manualOverride ? true : loanAmount > 20000`; `blocked ? false : loanAmount >= 20000` | Override shows; blocked hides; otherwise evaluate the comparison |
| CAP-3, CAP-4 | `vip ? true : (blocked ? false : loanAmount > 20000)` | Exercise VIP, blocked and threshold outcomes using literal branches |
| CAP-3 | `true ? true : missingFlag`; `if(false, missingFlag, true)` with the path absent | True without resolving the unselected missing path |
| CAP-3 | `x ? (a ? b : c) : z`, boolean data | Exercise x false, x/a true, x true/a false |
| CAP-3 | `x ? y : a ? b : c` | Same as right-parenthesized form |
| CAP-3 | `x ? y : amount / divisor > 1`, x true, divisor zero | Returns y without division error |
| CAP-3 | Unselected path is missing versus unselected syntax is malformed | Missing path is not resolved; syntax is rejected |
| CAP-3, CAP-4 | `if(flag, upper(1), "ok")`; `flag ? upper(1) : "ok"` | Static type error at load/commit for either form, including when flag would be false |
| CAP-4 | `if(loanAmount > 20000, eligible, fallback)` | Same condition behavior as Visibility |
| CAP-4 | `loanAmount + fee` in Visibility | Boolean-result error |
| CAP-4 | `if(flag, 1, fallback)`; `flag ? 1 : fallback` in Visibility | Static boolean-result error for either form, even if sample data would select a boolean fallback |
| CAP-4 | `true`, `false`, `null` and `(true)` in Visibility with supplied data `{}` | Visible, hidden, hidden, visible; no data lookup |
| CAP-4 | JSON `"visibleIf": null`, JSON `"visibleIf": "null"`, cleared editor field | Visible, hidden, visible |
| CAP-4 | `"true"`, `"false"` or `"null"` as the expression text in Visibility | Static type error: quoted words are strings |
| CAP-4 | Existing boolean path, null path, blank Visibility | Preserve current true/false, null-false and always-visible semantics |
| CAP-4 | Fixed Bold/Italic and other fixed boolean properties | Existing literal-boolean controls/schema remain; no expression editor or formula serialization for them |
| CAP-4 | `sum(true)`; `count(null)`; `formatNumber(1, false)` | Static argument errors; condition-slot literal support does not relax collection-path or string-pattern requirements |
| CAP-4 | `trueFlag`, `falseValue`, `nullable`, `True`, `record.true`, `params.null` | Paths under A8, resolved through existing scope rules |
| CAP-4 | Bare `true` with supplied data `{"true": false}`; `true.field`; `true()` | Literal true; syntax error; syntax error |
| CAP-4 | Text `{{true ? "Yes" : "No"}}`; text `{{null}}`; text `{{true}}` or `{{false}}` | Yes; empty text; consumer type error under existing text-output rules, without boolean-to-string coercion |
| CAP-5 | Valid expression committed, undone/redone, saved/reloaded | Formula retained and behavior reproduced |
| CAP-5 | Save bare literals in Visibility or literal-containing valid expressions in text/table-column bindings | Serialized document requires `2.0`; no new major ceiling, migration or legacy checker |
| CAP-5 | `loanAmount >` or missing ternary colon | Field-specific syntax error; no partial document mutation |
| CAP-5 | Parameters inside arithmetic and either ternary branch | All references discovered even when a branch is unselected |
| CAP-5 | `true`, `false`, `null`; `params.enabled ? true : null` in parameter discovery | No references for the literals; only `params.enabled` for the ternary |
| CAP-5 | Literal-only conditions in no-sample-data preview; literal-containing formulas committed, undone/redone, saved/reloaded and rendered natively/through the worker | Preview needs no fabricated data for literals; true shows, false/null hide; authored formula text and typed results survive the round trip |
| CAP-5 | Excessive nesting/size, no-data preview, native/worker evaluation | Bounded failure or documented supported behavior; no crash or divergent semantics |

Verify real engine/worker commits and rendered output; a mocked acceptance response does not demonstrate support. Division expectations above assert Decimal value and scale in engine checks, and comparison outcomes through boolean consumers; numeric results still require existing formatting functions in text.
