---
title: Folio PRD — Addendum (Technical-How)
status: final
created: 2026-08-22
updated: 2026-08-23
---

# Addendum — Technical Depth for Downstream Documents

Implementation detail that does not belong in the PRD body, which states capabilities rather
than mechanisms. This is the handoff payload for `bmad-architecture`.

Sections A–I are extracted from `docs/folio-mvp-plan.md` (section numbers refer to it).
Sections J–N record findings from the verification pass and supersede parts of A–I where
noted.

---

## A. System Decomposition

**Component chain** (§ MVP Core Deliverables):

```
Folio Designer → .folio Template → folio-go → PDF
                                   ├─ Expression Engine
                                   ├─ Layout Engine
                                   ├─ Pagination Engine
                                   └─ PDF Renderer
```

**Internal module tree** (§5):

```
template/    parser, schema, validator
expression/  evaluator, functions
layout/      text, table, pagination, measurement
renderer/    pdf
folio/       public API
```

**Processing pipeline** (§5):
`Template → Expression Evaluation → Layout Calculation → Page Model → PDF Renderer`

**Rejected anti-pattern** (§5): "Template → Draw Directly Onto PDF". An intermediate
page/layout model is mandatory — and per PRD NFR4 the pipeline is two-pass regardless, since
`Page X of Y` needs the total page count before serialization. It is also the seam that makes
future PNG / SVG / HTML renderers possible.

**Excluded from the shared layout model** (§5): Excel — different layout semantics, must not
be forced into the same model.

**Intermediate data structure** (§5):

```go
type Page struct {
    Width    float64
    Height   float64
    Elements []Element
}
```

> **Superseded by NFR1.a.** These `float64` fields must become integer fixed-point (see §K).

---

## B. Public API Surface (folio-go)

Suggested module path (§4, **undecided**): `github.com/folio-reports/folio` *or*
`github.com/folio-reports/folio-go`. The code example imports the former; the section title
and feature matrix call the component "folio-go". Public API identity unresolved.

```go
folio.LoadTemplate(path string) (Template, error)
folio.Render(template Template, data any) ([]byte, error)
folio.RenderTo(w io.Writer, template Template, data any) error
```

Data passed as `map[string]any` / `[]map[string]any` in the plan's examples (§4). Parameters
must also appear in the signature per FR40 — the plan's examples omit them.

**On `RenderTo`:** per PRD NFR4 (and FR39) this is an API-shape decision, not a
constant-memory guarantee. Keeping the writer signature now lets genuine incremental
rendering arrive later without a breaking change.

---

## C. Template Format Internals

Text-based, JSON internally, extension `.folio` (§3):

```json
{
  "version": "1.0",
  "page": {
    "size": "A4",
    "orientation": "portrait",
    "margin": { "top": 0, "right": 0, "bottom": 0, "left": 0 }
  },
  "body": [
    { "type": "text", "value": "..." }
  ]
}
```

Binding syntax literals (§2, §7):

| Form | Meaning |
|---|---|
| `{{path.to.field}}` | Scalar field binding by dotted path |
| `transactions[]` | Collection binding for a repeating region |
| `{{params.reportDate}}` | Runtime parameter namespace |
| `{{sum(transactions.amount)}}` | Aggregation over a collection field |
| `{{transaction.amount}}` | Per-row scope inside a table — **scope rule undefined (Q3)** |

FR33 (images embedded inside the template, never fetched by URL or read from disk) is in
tension with §3's "Git-friendly, diffable" goal. Consider whether image payloads can live in
a separate, clearly-delimited region of the file so text diffs stay readable.

---

## D. Expression Engine

MVP function catalogue (§6) — **eight functions**:

- **Aggregation** — `sum()`, `count()`, `avg()`
- **Formatting** — `formatDate()`, `formatNumber()` with pattern arguments
  (`"dd/MM/yyyy"`, `"#,##0.00"`)
- **String** — `upper()`, `lower()`
- **Logic** — `if()`

Constraint (§6): "Avoid creating a general-purpose scripting language in the first MVP."

**Two PRD constraints apply here.** Formatting functions imply a locale model, but FR43
forbids reading the host locale — the model must be explicit and carried in the template or
parameters (Q10). Per NFR1.c, a `formatNumber` implementation must not reach for
`math.Pow(10, n)` to do decimal scaling; that is a realistic accident that would break
byte-reproducibility.

---

## E. Determinism Control Points

Named in §10: font handling · font embedding · text measurement · line wrapping · page
dimensions · pagination rules · image sizing.

> **Superseded.** The source plan stated determinism at *layout* granularity. PRD NFR1
> requires byte-identity, and §K below replaces this list with the verified constraint set.
> The plan's list is retained only as the origin of the requirement.

---

## F. Table Pagination

Page-break behaviour (§8): the header row repeats at the top of each continuation page.
Framed as "one of the most important technical areas in the entire MVP" and "expected to be
one of the highest-risk engineering areas" (§15 Phase 3).

Priority rule (§8): "Correct pagination should take priority over advanced table styling."

The source plan specified this as the single word "correctly". PRD FR25 now defines atomic
rows, over-tall row handling, orphan-footer protection, and fixed column widths.

---

## G. Preview and Service Surfaces

**Preview flow** (§11): `Designer → (template + sample JSON) → Preview API → folio-go → PDF`

> **Superseded by FR35.** The "Preview API" is dissolved: `folio-go` compiles to WebAssembly
> and runs in the browser tab. There is no preview service, no network hop, and no separate
> preview renderer. See §J.2.

**REST contract shape** (§12, *Optional* — PRD FR45):

```
POST /api/reports/{name}/render
{ "parameters": {...}, "data": {...} }
→ Content-Type: application/pdf
```

> The `{name}` path segment presupposes server-side template storage. Per FR45 the template
> travels in the request or is resolved from an operator-supplied path — there is no
> server-side report repository. See §J.1.

**Recommended data architecture** (§2):
`Database → Application / Microservice → JSON → Folio → PDF`

---

## H. Proposed Build Order (§15)

| Phase | Content |
|---|---|
| 1 | `.folio` schema, parser, validator, text rendering, page model, PDF renderer, font handling → single-page `folio.Render` works |
| 2 | Text wrapping, element measurement, page boundaries, headers, footers, page numbers, multi-page |
| 3 | Tables — collection binding, rows, column layout, page breaking, repeated headers, aggregation |
| 4 | Expression engine — field binding, parameters, formatting, aggregation, conditional logic |
| 5 | Designer — canvas, components, properties panel, drag/resize, binding UI, save/load |
| 6 | Production preview integration — the WebAssembly build, not a preview service (§G) |

Sequencing rule (§15): "The Go rendering engine should be implemented before investing
heavily in the designer."

**Two unresolved objections** (Q5, Q6): Phase 3 depends on aggregation delivered in Phase 4;
and the designer sits at Phase 5 despite carrying success measure S5 and serving half the
user base. **The Thai line-breaking work (§M) is not in this build order at all and must be
inserted** — it is a prerequisite for Phase 2's text wrapping, not a refinement of it.

---

## I. Technology Choices NOT Made in the Plan

Open for `bmad-architecture`: PDF generation library · font engine and shaping library ·
designer frontend framework and language · expression parser approach · template schema and
validation technology · Go version floor · persistence layer for the designer.

§L and §M below narrow several of these substantially.

---

## J. Tensions Carried Forward — Status

1. **REST service presupposes a template repository.** §12 addresses templates by name on the
   server, which implies server-side storage — but §14 excludes a report repository.
   **Status: resolved by FR45** — the template travels in the request or comes from an
   operator-supplied path.

2. **Preview API is unscoped work.** ~~A *Must* capability depends on infrastructure that
   appears in no feature matrix and no build phase. If the designer is a web app and folio-go
   is a Go library, a server-side preview path is unavoidable.~~
   **Status: resolved, and the earlier conclusion was wrong.** A server-side path is *not*
   unavoidable: `folio-go` compiles to WebAssembly and runs in the browser (FR35). No preview
   service exists. This supersedes the original entry.

3. **Build order contradicts dependency order.** Phase 3 tables need `sum`/`count`/`avg` from
   Phase 4. **Status: open — Q5.**

4. **Build order back-loads the acceptance definition.** Designer at Phase 5, preview at
   Phase 6, while §17 lists the designer as criterion 1. **Status: open — Q6.**

5. **`Page X of Y` needs total-page resolution.** **Status: resolved** — two-pass layout,
   with the streaming memory claim dropped (NFR4).

6. **Conditional visibility vs. conditional formatting.** **Status: resolved by FR20** —
   visibility in, formatting out.

7. **Layout model unstated.** Absolute placement (FR1, FR5) versus flowing content (FR22,
   FR25) was never reconciled. **Status: resolved by PRD §6** — banded layout, absolute
   within bands, flow confined to tables, clipping on overflow (FR44).

---

## K. Byte-Reproducibility — Verified Mechanics

Measured on Go 1.26, across `darwin/arm64`, `darwin/amd64`, and `js/wasm`. This section
replaces §E.

### K.1 What actually varies — and what does not

| Source | Verdict |
|---|---|
| **FMA contraction in layout arithmetic** | **The real threat.** wasm forbids fusion by spec; arm64 fuses. `w += adv*size/1000.0` yields `2664.369755534079` on arm64 and `2664.3697555340796` on wasm. Go has no flag to disable it. Visually identical, different hash. |
| **`math` transcendentals** | **Not portable.** `Exp` differs on amd64; `Pow`, `Sinh`, `Erf` differ across all three targets. Safe set: `+ - * /`, comparison, `Sqrt`, `Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod`. |
| **Go map iteration** | Randomized per process — the dominant bug class in real PDF writers. Any map ranged into output must be sorted first. |
| **`compress/flate`** | **Stable.** Byte-identical across arm64/amd64/wasm and across Go 1.21–1.26. Stable by observation, not by contract — see PRD Risk R4. |
| **PDF object ordering / xref** | Deterministic if no map iteration reaches emission. Not an inherent problem. |
| **`/CreationDate`, `/ModDate`, `/ID`** | Trivially pinnable. Optional in PDF 1.7. |

The first draft of this PRD emphasized compression and object ordering and treated float
behaviour as a footnote. That was backwards.

### K.2 Spec position

- ISO 32000-1:2008 §14.4: `/ID` is a **should**, with an explicit note that *"the calculation
  of the file identifier need not be reproducible."* A fixed or content-derived `/ID` is
  conforming.
- Table 317: `/CreationDate` optional; `/ModDate` required only if `PieceInfo` is present.
- Exactly one `shall`-level randomness requirement exists in the specification, inside
  public-key encryption.
- **PDF 2.0 makes `/ID` mandatory** — the value may still be a content digest, but it must be
  supplied explicitly. Bears on Q11.
- §7.6.2 Algorithm 1(d) mandates a random 16-byte IV per encrypted string/stream — hence
  NFR1.g.

### K.3 Subset tags

ISO 32000-1 §9.6.4 requires the six-letter prefix to be *"arbitrary"* and unique within a
file. A deterministic hash over the glyph set satisfies both. Random tags are the industry
default and are widely treated as a defect.

### K.4 Prior art

Typst derives `/ID` from a hash of the serialized PDF and reuses those values for XMP, so
metadata adds no entropy; it honours `SOURCE_DATE_EPOCH`. pdfTeX, ReportLab, and qpdf
(`--deterministic-id`) all provide equivalent mechanisms. `SOURCE_DATE_EPOCH` is the
established convention and is adopted by NFR1.f.

---

## L. PDF Writer Candidates — Measured

Reproducibility measured across **separate OS processes** (same-process loops hide
map-iteration nondeterminism).

| Library | Result | Licence | Notes |
|---|---|---|---|
| **`boxesandglue/boxesandglue`** | Reproducible with `SetSuppressInfo(true)` or `SOURCE_DATE_EPOCH` | **BSD-3** | Content-derived `/ID`. The only Go library with a designed, tested reproducibility story. Pre-1.0, API expected to change. |
| **`signintech/gopdf`** | **Reproducible by default**, including subsetted TTFs and images | **MIT** | Actively maintained, widely used. Two caveats: emits no `ABCDEF+` subset tag (violates §9.6.4, will fail PDF/A validation), and `ImportPage` ranges a map — avoid templates. |
| `seehuhn.de/go/pdf` | Reproducible; cleanest API | **GPL-3.0** | Almost certainly disqualifying for this project. |
| `benoitkugler/pdf` | Reproducible with Info and `/ID` pinned | MIT | Immature, solo, `go 1.16`. |
| `codeberg.org/go-pdf/fpdf` | Conditionally reproducible | — | Requires `SetCatalogSort(true)` **and** no two images sharing a width — the sort key is image width and `SliceStable` preserves map order for ties. Present in all fpdf lineages. |
| `pdfcpu` | **Disqualified** | Apache-2.0 | Clobbers dates unconditionally; hashes `time.Now()` into `/ID`; maintainer has explicitly declined determinism requests. |
| `unidoc/unipdf` | **Disqualified** | Commercial EULA | No longer AGPL. Requires a licence key; obfuscated source. |
| `maroto` v2 | **Disqualified** | — | No `/ModDate` control; never sorts image objects. |
| `tdewolff/canvas` | Two lines from reproducible | MIT | Hardcodes `time.Now()` as `/CreationDate` with no setter. |

**Archive note:** `jung-kurt/gofpdf` and `github.com/go-pdf/fpdf` are both archived. The live
fork is `codeberg.org/go-pdf/fpdf` — a genuine module path change.

**Every candidate above builds for `js/wasm` and `wasip1/wasm` with `CGO_ENABLED=0`.** The
first draft's worry that cgo would disqualify PDF libraries was unfounded; the real
disqualifiers are timestamps and map iteration.

---

## M. Text Stack — Measured

### M.1 Shaping and subsetting

| Library | Verdict |
|---|---|
| **`boxesandglue/textshape`** (MIT) | **Recommended.** Pure-Go HarfBuzz port *and* `hb-subset` port in one module, zero dependencies. Full GSUB/GPOS, correct Thai, CFF + CFF2 + glyf, variable-font instancing. **Subset output is byte-stable across processes**, with a determinism test in-tree. Beta, `v0.0.x`. |
| **`go-text/typesetting`** (UNLICENSE/BSD-3, HarfBuzz subdir MIT) | **Recommended as cross-validation.** Passes 1,869 HarfBuzz conformance cases including 52 Thai SARA AM cases. Used by Fyne, Gio, Ebitengine. Produces byte-identical Thai output to textshape. **No subsetting.** Do not import `fontscan` in a wasm build. |
| `tdewolff/font` | **Disqualified by NFR1.e** — stamps `time.Now()` into `head.modified`, producing different bytes every run. |
| `seehuhn.de/go/sfnt` | Technically strong, deterministic — but GPL-3.0. |
| `x/image/font/sfnt` | **Not sufficient.** No `GSUB` at all; `GPOS` limited to Latin `kern` pair positioning, so no mark-to-base — fatal for Thai. No encoder, so no subsetting. Silently renders the default instance of a variable font. Explicitly disclaims hardening against malicious input. |

### M.2 Thai line breaking — must be built

Unicode UAX #14 assigns Thai to class SA (complex context) and defers to a dictionary. Every
Go implementation applies only the default rule, so a 47-rune Thai sentence yields **one**
break opportunity — the end.

No usable library exists. `mapkha` is archived and LGPL; `go-thaiwordcut` is stale with a
UTF-8 slicing bug; `go-wordseg` ships no licence; `gse` has no Thai support. **Two of them
resolve their dictionary via `runtime.Caller` at runtime — unconditionally fatal under
WebAssembly.** No Go port of libthai or ICU4X exists.

**Recommendation:** embed a dictionary and write longest-match/DAG segmentation directly.

| Source | Words | Compiled | Licence |
|---|---|---|---|
| ICU `thaidict.txt` | 26,383 | **~123 KB** as a BytesTrie | Unicode License |
| PyThaiNLP `words_th.txt` | 62,106 | larger | **CC0-1.0** |

libthai's own trie is ~581 KB for a smaller wordlist — prefer a BytesTrie representation.

Standard ICU builds use dictionary matching, not the LSTM model, for Thai; the model is
excluded from default builds. Dictionary segmentation is the mainstream behaviour, not a
compromise.

### M.3 Font payload

Outline format dominates every other consideration:

| Format | Compression |
|---|---|
| TTF/glyf static | ~35–40% of original |
| TTF/glyf variable | ~44–51% |
| **OTF/CFF (CJK)** | **~70–73%** |

CFF barely compresses because WOFF2 defines transforms only for `glyf`, `loca`, and `hmtx`.

> **Take the glyf/TrueType variable build.** `NotoSansSC[wght].ttf` is 17.77 MB on disk →
> **7.42 MB** compressed, covering nine weights. `NotoSansCJKsc-Regular.otf` is 16.44 MB on
> disk → **10.90 MB** compressed, covering one. Larger uncompressed, 3.5 MB smaller over the
> wire.

Thai is cheap: Noto Sans Thai (Thai-only, upstream) is 94,872 bytes as a variable font
against 218,652 for the Google Fonts build that redundantly includes Latin. Strip hinting if
rasterizing independently.

### M.4 WebAssembly budget

| Component | Compressed |
|---|---|
| Go runtime floor (`fmt`) | ~0.57 MB |
| + `boxesandglue/textshape` | ~0.82 MB |
| Engine, PDF writer, Thai trie | ~1.5 MB estimated |
| **CJK font** | **~7.4 MB** |
| **Total** | **~9 MB** |

Fonts dominate roughly five to one. The first draft attributed the payload problem partly to
the Thai dictionary; at ~123 KB it is negligible.

**Standard Go, not TinyGo.** TinyGo is ~5.6× smaller on realistic programs but: `recover()`
does not work on wasm, so any nil dereference while parsing an untrusted font traps the whole
module; the `unsafe` zero-copy `[]byte`↔`string` idiom — common in font and parsing libraries
— **silently miscompiles**, failing at runtime with corrupted data rather than at build time;
and TinyGo trails the Go release.

Go wasm binaries have grown roughly 27–40% from Go 1.21 to 1.26; there is no future version
to wait for. `-ldflags="-s -w"` saves about 2%. `wasm-opt` is effectively a no-op once brotli
is applied. **Serving brotli is worth more than every compiler flag combined.**

---

## N. Verification Harness

Implied by NFR1 and §9 but not yet designed:

- Golden-hash fixtures per page count, recorded per `folio-go` version **and Go toolchain
  version** (Risk R4).
- CI matrix **must** include an arm64 target alongside amd64 and wasm. Contraction is an
  architecture property, not an OS one — arm64 contracts, amd64 and wasm do not (measured on
  `darwin/arm64`, §K.1). An amd64-and-wasm-only matrix passes while arm64 users receive
  different bytes.
- A wasm target executed under Node or an equivalent runtime, hash-compared against native.
- An expected-break fixture for Thai and CJK line breaking (S4), hand-checked once and then
  frozen.
- An import-restriction lint banning `math` transcendentals outside an allow-listed shim
  (NFR1.c), and an analysis pass for un-barriered multiply-add if the fallback float approach
  is ever taken instead of fixed-point (NFR1.a).
- A check that no map iteration reaches PDF emission (NFR1.d).
