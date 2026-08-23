# Acceptance — Folio MVP v0.1

How Folio is judged done. The golden report is the fixture; the measures decide pass; the
counter-metrics decide whether passing happened in the right direction.

## The golden report

A **Customer Account Statement**, the primary acceptance fixture and the subject of most
measures below.

| Element | Requirement |
|---|---|
| Header | Embedded logo, customer information, account information, statement period |
| Table | Date, Description, Debit, Credit, Balance — five fixed-width columns, repeated header row, sum footer |
| Footer | Confidentiality text, generated date, `Page X of Y` |
| Generated date | Supplied through `params` — **never** read from the clock. A fixture stamping wall-clock time cannot satisfy a hash-equality test, and the engine cannot read a clock at all. |
| Script coverage | Latin, Thai, and CJK text **in the same table** |
| Wrapping | At least one row whose cell content wraps to multiple lines |
| Images | At least one embedded image |
| Page counts | Must render correctly at **1, 5, 20, and 50 pages** |

A **second, structurally different template** (S7) exists so the engine is not fitted to one
document shape.

## Success measures

Each names what would otherwise let it pass vacuously.

| # | Measure | Target |
|---|---|---|
| S1 | Golden report renders correctly at 1, 5, 20, 50 pages — verified against recorded reference renders, not merely "produces a file" | 4/4 |
| S2 | Native-vs-WASM hash equality on the golden report at all four page counts, with the reference render confirmed non-blank and page-count-correct | Identical |
| S3 | Cross-platform hash equality across `darwin/arm64`, `linux/amd64`, **and `linux/arm64`** | Identical |
| S4 | Thai line breaks match a hand-checked expected-break fixture; CJK breaks per character; Thai marks positioned by GPOS | Fixture matches exactly |
| S5 | Round trip — author in the browser, save locally, render via `folio-go` — where the template was *authored in this session*, not merely re-saved | Byte-identical to the preview |
| S6 | A 50-page render completes within a recorded memory ceiling, established as a baseline and enforced against regression | No regression |
| S7 | A second, structurally different template renders correctly and reproducibly | Pass |
| S8 | Every error case produces a located, actionable message | All cases covered |
| S9 | A template hand-edited in a text editor — never opened in the designer — loads and renders | Pass |

## Counter-metrics

Signals that MVP is succeeding in the wrong direction.

| # | Counter-metric | Why it matters |
|---|---|---|
| C1 | Expression function count exceeding **eight** | The language is becoming a scripting language |
| C2 | Component palette exceeding **five** | Breadth is displacing reliability |
| C3 | Any determinism exception carved out to ship a feature | Byte-identity is the product; erosion is fatal |
| C4 | Designer work starting before the engine renders the golden report | The sequencing rule, violated — but see PRD Q6, which holds this open |
| C5 | Time-to-first-PDF for a new integrator exceeding a few minutes | The "extremely simple API" goal has failed |
| C6 | Golden hashes regenerated rather than investigated when CI goes red | The regression suite has stopped being a test |

## Verification harness

The harness is a deliverable, not a by-product — it is the only thing that can observe most of
the constraints in `SPEC.md`.

- **Golden-hash fixtures** per page count, recorded per `folio-go` version **and per Go
  toolchain version**. A toolchain bump is a versioned breaking change.
- **CI matrix must include an arm64 target** alongside amd64 and wasm. FMA contraction is an
  architecture property, not an operating-system one: arm64 contracts, amd64 and wasm do not.
  A matrix covering only amd64 and wasm passes while every arm64 user receives different bytes.
- **A wasm target executed under Node** or an equivalent runtime, hash-compared against native.
- **An expected-break fixture** for Thai and CJK line breaking (S4), hand-checked once and then
  frozen.
- **An import-restriction lint** banning `math` transcendentals outside an allow-listed shim.
- **A map-iteration check** confirming no unordered iteration reaches PDF emission.
- **Format round-trip tests** shipping with the template format, not after it:
  `Parse(Serialize(d)) == d` and `Serialize(Parse(b)) == b` for canonical `b`.
- **No change lands without a fixture covering it.** A hash change is investigated as a defect
  until proven to be an intended, versioned behaviour change (C6).

## Carried risks

Scope was deliberately held rather than cut, so these are carried explicitly.

| # | Risk | Severity |
|---|---|---|
| R1 | FMA contraction silently breaks byte-identity; divergence is invisible in output and an amd64-and-wasm-only CI matrix passes anyway | Critical |
| R2 | Thai line breaking must be written from scratch, on the critical path — no maintained pure-Go segmenter exists, and it is a prerequisite of text wrapping rather than a refinement of it | High |
| R3 | The strongest candidates for shaping and subsetting are pre-1.0 with explicitly unstable APIs | Medium |
| R4 | Compressor output is stable by observation, not by contract — a future Go release could invalidate every recorded golden hash downstream | Medium |
| R5 | Solo capacity against held scope — the dominant risk to delivery | High |
| R6 | `Page X of Y` forecloses true streaming for the life of MVP | Low (accepted) |
