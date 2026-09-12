# Boolean formulas implementation review

Reviewed all tracked and untracked changes from `ae6cd70530b6059d8e66db448fb432533dd0df94` with independent blind, edge-case and verification-gap reviewers. All reviewers completed before triage. The approved intent block remains unchanged (SHA-256 `29563293ff89990711705c6235335aff94cc0481c91ba2d758f8441246bb1b67`).

The findings below are local implementation or verification corrections covered by the existing contract. They require no change to user intent or the language design. Repeated claims with the same corrective action are combined.

| ID | Severity | Route | Finding and required correction |
| --- | --- | --- | --- |
| R1 | low | patch | Restore outer whitespace handling in the parser while retaining offsets into the original UTF-8 source; add boundary whitespace cases. |
| R2 | high | patch | Reject unsupported value wrappers implementing `Expr` before pointer-only reflection; malformed caller-built ASTs must return located errors, never panic. |
| R3 | medium | patch | Merge visibility warnings into the existing band/declaration warning sequence; preserve the documented existing table-warning exception and relative order of pre-existing warnings. |
| R4 | medium | patch | Prefer a compatible null stand-in against mixed conditional scalar kinds, e.g. `x != (flag ? 1 : "a")`. |
| R5 | medium | patch | Reconcile direct unknown inequality paths when their existing admitted candidate sets share a legal scalar, e.g. date/number demands sharing numeric zero. This is candidate intersection, not solving condition outcomes. |
| R6 | high | patch | Do not evaluate discarded inequality subexpressions while generating preview data; retain both-branch demand traversal and the documented conservative divisor policy. |
| R7 | medium | patch | Preserve element/field context with deferred preview checks and include it in refusals. |
| R8 | medium | patch | Attach available expression locations to final runtime condition/text consumer-kind failures. |
| R9 | medium | patch | Keep the consuming field in missing-path diagnostic messages while retaining the missing data path in the structured location. |
| R10 | low | patch | Avoid asserting conditional syntax merely because a quoted string contains `?`; use accurate generic notice wording without introducing a JavaScript formula parser. |
| R11 | medium | patch | Identify which placeholder failed when a text field contains multiple expressions, while preserving bounded cause/offset reporting. |
| R12 | medium | patch | Test shared resolver/evaluation budget wiring through production `EvaluateCondition`; the existing tests construct their own wired resolver and do not protect this entry point. |
| R13 | low | patch | Cover a ternary-only text template without Visibility or `if()` in the no-data notice tests. |
| D1 | medium | defer | Aggregate evaluation with a nil resolver already panicked before this feature. Record separately; production bindings supply a concrete resolver. |

## Verification before review fixes

All six approved matrix rows have passing Go JSON events in `/Users/panitw/Library/Application Support/rtk/tee/1789192995_go_test.log`. The log is truncated after 1 MiB, but all named matrix pass events precede truncation. The real browser test passed literal previews, history, persistence, long syntax diagnostics and native/worker PDF parity for both threshold outcomes. Final patch verification is recorded below.

## Review fix audit

All R1–R13 fixes are implemented and audited against their regression coverage. The frozen intent checksum is unchanged, and `git diff --check` passes. D1 remains recorded separately in deferred work. No intent or design loopback was needed (`review_loop_iteration: 0`).

- Focused Go review regressions: **122 passed** across three packages.
- Full Go suite after fixes: **2473 passed, 2 known P6g failures, 5 skipped**. Formula matrix and new regression pass events are present in `/Users/panitw/Library/Application Support/rtk/tee/1789194854_go_test.log`; the only failures are the unchanged opaque-name corpus floor and its parent.
- Architecture: **227 passed**. Frontend: **1519 passed** across 80 files; typecheck, lint and e2e compilation pass with eight existing lint warnings.
- Real js/wasm host checks retain cause, source offset, field and second-placeholder identity within the bounded wire message.
- The production budget regression passes normally and fails when an isolated Go overlay removes only `budget: budget` from `EvaluateCondition`. Mutation evidence: `/tmp/folio-boolean-formulas-r12-mutant.log`.
- Final real browser rerun: **1 passed**, including authoring/history, bounded errors, literal no-data previews, exact save/reload and native/worker PDF equality for both threshold outcomes.
