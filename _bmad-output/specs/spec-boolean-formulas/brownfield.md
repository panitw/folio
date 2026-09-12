# Brownfield integration

## Observed implementation

- `folio-designer/src/App.tsx`: Visibility edits `visibleIf` as a condition; helper text currently advertises only a boolean field or call.
- `folio-go/internal/expr/parser.go`, `ast.go`, `check.go`, `eval.go`: shared grammar, AST, static checking and evaluation. Current grammar has no operators; current literal checks assume no boolean literals. Revised A5 requires boolean/null AST leaves and kind-aware condition checks.
- `folio-go/internal/expr/table.go`: eight existing named functions. Operators do not require adding named functions.
- `folio-go/internal/expr/decimal.go` and `reduce.go`: bounded scaled-integer Decimal and existing half-even average behavior. Q2 now adopts operand-scale-plus-four and half-even rounding for division; the language companion fixes exact expectations.
- `folio-go/internal/bind/condition.go`, `folio-go/folio_expr_validate.go`, `folio-go/render_visibility.go`: bare visibility expressions, load-time validation and document-scope evaluation.
- `folio-go/component_commands.go`: property commands can surface generic component format failures; Bold/Italic currently require literal JSON booleans.
- `folio-go/parameter_references.go`: recursively visits paths and calls; new node children must not disappear from parameter discovery.
- `folio-go/stand_in_data.go`: demand inference walks current node variants. New operators require deliberate handling for no-data preview; generated data must not be reported as satisfying a condition when it does not.
- `folio-go/internal/expr/footer.go`: inspect derivation assumptions and preserve existing footer restrictions when adding AST forms.
- `folio-go/internal/template/version.go`: derives the saved format requirement from serialized content; the current format ceiling is `2.0`.

## Obligations

Extend one shared parser/checker/evaluator and audit all exhaustive AST switches, walkers and literal assumptions. Preserve function argument constraints and runtime branch laziness, and apply bounds across every new nesting form. Apply the language companion's static checks consistently to `if()` and ternaries. Update authored grammar documentation, help text, format validation and obsolete tests that assert all operators are invalid or no boolean/null literals exist.

Replace condition-slot checks that reject every literal with checks that accept boolean/null literals and reject numeric/string results. Keep aggregate path-only arguments and formatting pattern string-only arguments independently constrained. Audit `IsLiteralExpr`, function argument kinds, bare Visibility validation and evaluation, parameter discovery, and stand-in traversal: new literal nodes contribute no data references and require no generated values.

Keep the existing AST-to-binding dependency direction and decimal invariants. No frontend JavaScript evaluator, external scripting engine or float-based arithmetic shortcut.

Editor errors should identify `visibleIf` (or the actual condition field) and the expression cause, with token location when available. Keep runtime data failures distinct from static validation; do not reject a valid data-dependent expression merely because sample data is absent. Preserve undo/redo and command atomicity.

Verify persistence using real serialized templates and native/worker engine paths. Unaffected document output, runtime-lazy `if()`, null conditions, parameter discovery, sample/stand-in preview and footer derivation need regression evidence appropriate to the new AST. Fixtures that depended on permissive static checking may change to match the new contract.

## Release and historical contract changes

The user confirms that no release has reached external users. The requested syntax joins the existing unreleased `2.0` format; it does not require a separate major release, migration layer, legacy parser or compatibility mode. The earlier blanket requirement to keep every previously accepted expression valid is superseded.

Under the existing content-based save rule, expressions using this extension, including standalone boolean/null literals, in Visibility, text or table-column bindings require `2.0`. Saving does not lower the loaded version. No additional format ceiling is introduced by this feature.

This request replaces the no-operator grammar restriction while retaining the eight-function registry, strict condition typing and decimal invariants. Revised A5 explicitly supersedes the earlier field-name decision: `true`, `false` and `null` are literal values. A8 defines lowercase keyword boundaries and dotted-field handling; no compatibility mode is required.

The current project-wide spec lives at `../spec-folio/SPEC.md`; its format companion is `../spec-folio/folio-format.md`. This feature spec records the override without rewriting that separate canonical workspace; implementation must update its grammar and content-version documentation to match.

## Confirmed scope and precision

Q1 retains the prohibition on data-driven formatting (AD-24): fixed Bold/Italic and other fixed boolean properties remain literal boolean controls. The feature adds no new property expression slots, styling schema or styling evaluation path. Existing expression containers continue to use the shared language under their existing output-type rules.

Q2 fixes division at max(operand decimal scales, 0) + 4 fractional places using round-half-to-even. The result retains that scale and the existing Decimal bounds; exact terminating, repeating, tie, negative, tiny-result and overflow expectations are specified in the language companion.

Configured `project-context.md` was absent at specification and review time; these observations come from source inspection.
