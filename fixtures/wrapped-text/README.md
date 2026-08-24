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
- **The leading is welded in.** Line-to-line advance is the maximum `ascent - descent + lineGap` over
  the *declared* font stack (D-2.4.2): 1511/1000 em from Noto Sans Thai, 16,621 mp at 11 pt. Changing
  that rule re-records this fixture and every one recorded after it.
- `input.folio` is kept byte-identical to `folio-go/wrapped_text_template.go`'s
  `wrappedTextTemplateJSON` by hand, as `font-text` and `multi-script-fallback` already are.

## Cross-target status

Registered in `matrixDocuments` **and** in `.github/workflows/matrix.yml`, and — unlike `shaped-text`
— its four legs were **actually run in its own story**, because D-000.4 names 2.4 as a per-story
matrix override ("line breaking feeds every measurement"). All four targets agree:

```
darwin/arm64   3845da37ae198beae3d3ef98211678b02a397a87336cea025e2e8286a712288e  72743 bytes
linux/amd64    3845da37ae198beae3d3ef98211678b02a397a87336cea025e2e8286a712288e  72743 bytes
linux/arm64    3845da37ae198beae3d3ef98211678b02a397a87336cea025e2e8286a712288e  72743 bytes
js/wasm        3845da37ae198beae3d3ef98211678b02a397a87336cea025e2e8286a712288e  72743 bytes
```
