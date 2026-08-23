# Reviewer lens — good-spine rubric

Verdict: **PASS with amendments** (the adversarial lens's findings, applied).

| Criterion | Judgement |
|---|---|
| Fixes the real divergence points for the level below | Strong after amendment. The pre-amendment draft covered the determinism, ownership, and designer/engine seams well but left four core-semantics holes (data numerics, coordinate origin, visibility, aggregate scope) that epics would have hit in week one. |
| Every Rule is enforceable and prevents its stated divergence | Yes. AD-1, AD-2, AD-3, AD-5 and AD-21 are machine-checkable; AD-9's round-trip property is a test that ships with the format; AD-13 and AD-24 remove classes by making a value derived rather than stored. AD-15 and AD-16 are structural — there is no TypeScript document to drift. |
| Nothing under Deferred lets two units diverge | Holds. Every deferred item is either an added shell around an unchanged core (REST service, other-language SDKs), a measurement to record later (memory ceiling), or a UI affordance (font upload, sample-JSON persistence). The one to watch is template↔library compatibility policy: AD-9 validates the version field, which is enough for MVP but not for v1. |
| Named tech verified current | Yes — see `review-currency.md`. |
| Brownfield ratification | N/A. Repository is empty; no conventions existed to ratify. |
| Covers the driving spec's capabilities | FR1–FR45 and NFR1–NFR8 all land in the Capability map. FR28 (alternating row styling) is PRD-optional and correctly carries no invariant. FR45 is deferred with its consequence — a threat model — named rather than dropped. |
| Inherited parent spine | N/A — no parent. PRD-settled matter is inherited by reference and not re-decided. |
| Every dimension decided, deferred, or open | Yes. The operational envelope is explicitly covered in *Deployment and environments*, which is the dimension a domain-focused draft usually leaves silent: static hosting with brotli and immutable URLs as requirements, a module proxy, a CI matrix, and an explicit statement that Folio operates no runtime at all. |
| Seed stays minimal | Mostly. The source tree is one level deeper than strictly required, justified here because the package boundary *is* the determinism boundary (AD-1) — the tree is carrying a rule, not documenting the code. |

## Observations not raised as findings

- 24 ADs is high for a feature-altitude spine. Justified: the product's signature requirement is
  byte-identity, which converts many ordinarily-free choices into invariants. Counter-metric C3
  makes eroding any of them a defined failure.
- Q5 and Q6 (build order) are correctly left to `bmad-create-epics-and-stories`. The spine imposes
  no phase order, which is the right call — but it should keep flagging that Thai line breaking
  precedes text wrapping rather than refining it, and it does.
- The `[ASSUMPTION]` tags are concentrated on the genuinely expensive, hard-to-reverse calls
  (owning the PDF serializer, the millipoint unit, single wasm instance, DOM canvas, React) rather
  than sprinkled. That is the correct use of the Fast path.
