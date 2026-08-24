# `fixtures/wrapped-text/` — line breaking in all three scripts (Story 2.4, S4)

The golden for **line breaking and measurement**. Every earlier text fixture *fits* its boxes — all
ten of their text elements were measured, the widest at 26% of its declared width — so none of them
can tell a build that wraps from one that does not. This is the only fixture whose elements are
measured to **overflow** their boxes, and therefore the only one whose bytes move if wrapping breaks.

## What each element is for

| Element | Script | Box | Full text measures | Lines | What it proves |
|---|---|---|---|---|---|
| `e1` | Latin | 150 pt | 331,683 mp | 3 | breaks after a whitespace run, and the run is **consumed** — drawn on neither line |
| `e2` | Thai | 150 pt | 192,830 mp | 2 | breaks come from the embedded dictionary, under AD-25's absolutes |
| `e3` | CJK | 150 pt | 198,000 mp | 2 | breaks between adjacent Han characters |
| `e4` | Thai, bound | **20 pt** | 46,585 mp | 2 | **the declared unbreakable value** |

All four boxes are narrower than the text they hold. That is measured, and asserted as a precondition
before any wrapping assertion runs — "it wrapped" must not be satisfiable by an input that fitted.

## `e4` is the discriminating element — read this before changing its width

`e4` binds `customer.name`, which the document declares in `unbreakableValues`. Its value is
`ศรีสุข` — D-2.1.6's worked case, a surname that is character-for-character the two common
dictionary words `ศรี` and `สุข`.

Its box is **deliberately narrower (20,000 mp) than the value itself measures (24,585 mp)**. That is
what makes the declaration load-bearing rather than decorative:

- **declared** (as shipped): `ผู้รับ` / `ศรีสุข` — the surname stays whole and **overflows its box
  visibly**, which is AD-25's own prescription and Story 2.8's clipping to finish.
- **undeclared** (the same template with the list removed, nothing else changed): `ผู้รับ` / `ศรี` /
  `สุข` — the surname **splits**.

`TestWrappedTextDeclarationIsLoadBearing` asserts **both**. Widening this box breaks the test on
purpose: at any width ≥ 24,585 mp the two polarities produce the same layout and the fixture stops
proving anything about the declaration.

## Rules

- **Hand-check and re-record are different things.** This fixture's *break placement* is certified by
  `fixtures/expected-breaks/`, not here — one human judgment, at the layer where it is legible
  (D-2.4.3, DN-4). There is deliberately **no third sign-off** for this document.
- **A hash change here is a defect until proven otherwise** (AD-21/AD-22). Do not regenerate to make a
  test pass. Every semantic property — line counts, per-line widths against the box, no boundary
  inside a Thai cluster, no boundary inside a declared span, the embedded faces, the `/ToUnicode`
  section sizes — is asserted at recording and none of it is deferrable (D-000.22, D-2.3.5).
- **The vertical model is welded in.** Both spans come from the *declared* font stack, each axis
  maximised **independently** (D-2.4.2 as **amended**, Story 2.5a):

  | span | rule | this fixture, at 11 pt |
  |---|---|---|
  | top → first baseline | `max(A)` | `1160/1000 em` (Noto Sans SC) → **12,760 mp** |
  | baseline → baseline | `max(A) + max(D) + max(gap)` | `1160 + 450 + 0 = 1610` → **17,710 mp** |

  The superseded form — `max(ascent - descent + lineGap)`, the largest *single* face — gave 1511/em
  and 16,621 mp, **99 units per em short**, because it assumed one face supplies both axes when Noto
  Sans Thai wins the descent (450) and Noto Sans SC the ascent (1160). Changing that rule re-records
  this fixture and every one recorded after it.

  **This is the ONLY fixture in the repository on which the baseline-to-baseline span is observable
  at all** — it is the only one with a multi-line element on a multi-face stack, and for a
  single-face stack the amended and superseded forms are identical. Measured by mutation: reinstating
  the superseded rule reddens this fixture and nothing else.
- `input.folio` is kept byte-identical to `folio-go/wrapped_text_template.go`'s
  `wrappedTextTemplateJSON` by hand, as `font-text` and `multi-script-fallback` already are.

## Cross-target status

Registered in `matrixDocuments` **and** in `.github/workflows/matrix.yml`, and — unlike `shaped-text`
— its four legs were **actually run in its own story**, because D-000.4 names 2.4 as a per-story
matrix override ("line breaking feeds every measurement"). All four targets agree:

```
darwin/arm64   277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5  72738 bytes
linux/amd64    277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5  72738 bytes
linux/arm64    277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5  72738 bytes
js/wasm        277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5  72738 bytes
```

## Provenance — external structural validation (D-000.53)

No golden artifact is accepted — first recording **or re-recording** (D-000.44) — until a reader this
project did not write parses it and resolves it into the semantic objects it claims to contain.

| | |
|---|---|
| reader | `qpdf` **12.4.0** |
| invocation | `qpdf --check fixtures/wrapped-text/expected.pdf` |
| result | exit **0** — `No syntax or stream encoding errors found` |
| invocation | `qpdf --show-npages fixtures/wrapped-text/expected.pdf` |
| result | **1** page(s), matching the declared `/Count` |
| validated at | `50ad6c8` (Story 2.6) |

**Settled**, validated unchanged at `50ad6c8`. This artifact predates D-000.53 and was validated retroactively; it is recorded here so the row ends settled rather than carried (D-000.29).

The external reader is the **acceptance instrument, at recording time only** — run off-leg on the
recording machine, hand-checked, output pasted here. It is never a runtime or CI dependency (AD-25,
`TestModuleGraphAllowlist`), and it is deliberately **not** gated to "the legs that have qpdf": a check
that runs on some legs and not others reproduces D-000.9's failure — an "all clear" indistinguishable
from "I could not look" — one level up, at the leg. The standing every-leg regression guard is the
in-repo checker `folio-go/golden_structural_validity_test.go`, which is hermetic and covers all four
targets including `js-wasm`.
