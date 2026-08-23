---
title: Folio PRD — Technical Feasibility Review
status: draft
created: 2026-08-22
reviewer: Principal engineer (pre-architecture feasibility gate)
scope: NFR1 (byte-reproducible rendering across compilation targets) and adjacent requirements
---

# Folio PRD — Technical Feasibility Review

**Subject:** `prd.md` + `addendum.md` (MVP v0.1)
**Central question:** Is NFR1 — byte-identical PDF output across compilation targets, WASM-in-browser hash-matching native-on-Linux — achievable?

---

## 0. Verdict

**NFR1 is achievable, but NOT as written, and not for the reason the PRD thinks.**

The PRD names the wrong risks. It flags "compression output" and "internal object ordering" as the hard parts. Those are the easy parts — I tested them and they are stable. The requirement's actual load-bearing risk is **floating-point fused multiply-add (FMA) contraction**, which the PRD mentions only in passing as "deterministic float behaviour across targets," as though it were an assumption one can adopt rather than a property one must engineer.

I ran the experiment. On Go 1.26, **ordinary text-measurement arithmetic already produces different `float64` results on `darwin/arm64` than on `js/wasm`**, and that difference reaches the emitted PDF bytes. The PRD's own success criteria S2, S3 and S5 fail today on an Apple Silicon development machine — which, given the author is developing on `darwin/arm64`, is the machine this project will actually be built on.

The good news is that the divergence has exactly two sources, both are enumerable, and both are controllable in pure Go without cgo. NFR1 survives if the PRD converts it from an aspiration into a **set of engineering constraints on the rendering engine's own arithmetic**, plus a **normative numeric-emission rule**. Section 9 states the strongest achievable version.

**Bottom line for the architecture phase:** do not treat NFR1 as "pick a PDF library that lets us pin the timestamp." Treat it as "design the layout engine's number system." The PDF-level variability (Q1) is a solved, low-risk problem. The arithmetic is the product risk.

---

## 1. Evidence base

Everything in Sections 2–4 is from executed experiments, not recollection. Environment:

- Go 1.26.0, `darwin/arm64` (Apple Silicon)
- Native `darwin/arm64`; native `darwin/amd64` executed under Rosetta 2; `js/wasm` executed under Node v24.16.0 with Go's own `wasm_exec_node.js`
- Cross-version checks against downloaded toolchains go1.21.13, go1.22.12, go1.23.4, go1.24.0, go1.25.0

Library claims were checked by fetching the modules into the Go module cache and inspecting source and build behaviour directly, not from documentation. Harnesses are in the session scratchpad (`dettest/`, `sweep/`, `layout/`, `pdfbytes/`, `rate/`, `mag/`, `maps/`, `wasmsize/`, `libcheck/`, `shape/`, `thai/`).

Where a claim is an estimate rather than a measurement, it is marked as such inline.

---

## 2. Finding F1 — FMA contraction breaks arm64 ↔ WASM equality in ordinary layout code

**Severity: CRITICAL.** This is the finding that decides NFR1.

### The claim in the PRD

> "Directly implied: no cgo in the rendering path, deterministic float behaviour across targets…" (NFR1)

The PRD treats "deterministic float behaviour across targets" as a consequence that follows from choosing the right library. It does not follow from anything. Go actively permits the compiler to break it.

### The evidence

The Go specification (verbatim, from `$GOROOT/doc/go_spec.html`, Go 1.26):

> "…fused operation, possibly across statements, and produce a result that differs from the value obtained by executing and rounding the instructions individually. An explicit floating-point type conversion rounds to the precision of the target type, preventing fusion that would discard that rounding. For instance, some architectures provide a 'fused multiply and add' (FMA) instruction that computes `x*y + z` without rounding the intermediate result `x*y`."
>
> ```
> // FMA allowed for computing r, because x*y is not explicitly rounded:
> r  = x*y + z
> r  = z;   r += x*y
> t  = x*y; r = t + z
> // FMA disallowed for computing r, because it would omit rounding of x*y:
> r  = float64(x*y) + z
> r  = z; r += float64(x*y)
> ```

So `w += adv*size/1000.0` — the single most common statement in any text layout engine — is explicitly licensed to produce architecture-dependent results.

I wrote that exact statement into a realistic glyph-advance measurement loop (400 glyphs, non-round advance widths, font size 10.7pt, tracking, horizontal scale) and ran it on all three targets:

| Target | `measure()` — `w += (adv+kern)*size/1000.0*hscale + tracking` | `measureBarrier()` — same, with `float64()` barriers |
|---|---|---|
| `darwin/arm64` | `40a4d0bd5098ecc5` | `40a4d0bd5098ed03` |
| `darwin/amd64` | `40a4d0bd5098ecc6` | `40a4d0bd5098ed03` |
| `js/wasm` | `40a4d0bd5098ecc6` | `40a4d0bd5098ed03` |

**arm64 differs from wasm in the last bit.** The barrier version is identical on all three.

This is not academic. It reaches the file. The same run, formatting the measured width the way a naive PDF writer emits a content-stream coordinate:

| Target | `strconv.FormatFloat(w, 'f', -1, 64)` |
|---|---|
| `darwin/arm64` | `2664.369755534079` |
| `darwin/amd64` | `2664.3697555340796` |
| `js/wasm` | `2664.3697555340796` |

Different string, different length, different content stream, **different SHA-256**. NFR1 fails, S2 fails, S5 fails — with no visual difference whatsoever in the rendered page.

### Why this happens, precisely

- **WebAssembly forbids contraction by specification.** The Wasm core spec requires IEEE 754 round-to-nearest-ties-to-even and states that implementations "are not permitted to contract or fuse operations to elide intermediate rounding steps." WASM is therefore the *most* deterministic target Folio has, not the least.
- **arm64 fuses.** Go's compiler auto-emits `FMADD` on arm64 (also ppc64, s390x, riscv64, loong64).
- **amd64 does not fuse today** — see [golang/go#71204](https://github.com/golang/go/issues/71204), "no automatic use of fused multiply-add on amd64 even with GOAMD64=v3". This is why `amd64` and `wasm` agree in my table. **That agreement is an implementation accident, not a guarantee**, and it is precisely the kind of thing that changes when GOAMD64 baselines move.
- The governing upstream issue is [golang/go#17895](https://github.com/golang/go/issues/17895) ("spec: allow the use of fused multiply-add floating point instructions"), whose resolution is the spec text quoted above. The Go team's explicit position is that **applications cannot rely on results being identical across GOARCH-dependent code paths.**

### There is no compiler flag

Unlike C/C++ (`-ffp-contract=off`) or nvcc (`--fmad=false`), **Go provides no flag to disable FMA contraction.** Go's answer was to add `math.FMA` as a *guaranteed-fused* intrinsic ([golang/go#25819](https://github.com/golang/go/issues/25819)) — the opposite of what Folio needs. The only mechanism to *prevent* fusion is the explicit `float64(...)` conversion, applied by hand, at every multiply-add site, forever.

That means the NFR1 discipline is **unenforced by the toolchain and invisible in code review**. A single `w += a*b` added in month six silently breaks byte-reproducibility on arm64 only, in a way that CI on an amd64 Linux runner will not catch.

### How bad is it in practice?

I measured the magnitude over 200,000 randomized measurement runs (modelling fused vs. unfused with `math.FMA` vs. explicit barriers):

- 32.5% of raw measurements differ at the bit level
- Absolute divergence: p50 = 1.14e-13 pt, p100 = 6.82e-13 pt

So the error is real but tiny. Over 300,000 randomized trials I found **zero** cases where it changed a line-break index and **zero** cases where it survived quantization to 4 decimal places. Analytically, the residual risk that a divergence straddles a rounding boundary is roughly `maxdiff / quantum`:

| Emission precision | p(divergence per number) | p(≥1 divergence in a 50,000-number document) |
|---|---|---|
| 6 dp | 6.8e-07 | 3.3e-02 |
| 4 dp | 6.8e-09 | 3.4e-04 |
| 3 dp | 6.8e-10 | 3.4e-05 |
| 2 dp | 6.8e-11 | 3.4e-06 |

**Read this correctly.** Quantizing emitted coordinates to 2–3 decimals reduces the failure probability to ~1 in 300,000 documents. That is *not* byte-reproducibility. It is a rare, silent, input-dependent, non-reproducible-on-demand hash mismatch — the worst possible failure mode for a requirement whose entire value proposition is that a downstream user asserts `sha256(output) == recorded_hash` in a CI test (UJ-2, S5). A 1-in-300,000 flake in a determinism guarantee is a support burden, not a guarantee.

**Quantization is a mitigation, not a fix.** The fix is to eliminate the divergence at the source.

### Recommendation

The PRD should stop asserting deterministic float behaviour and start *requiring* it, with a named mechanism. Add to NFR1:

> **NFR1.a — Layout arithmetic is contraction-free.** All layout and measurement arithmetic in `folio-go` must produce bit-identical `float64` results on every supported GOARCH. Because Go permits FMA contraction (Go spec, "Floating-point operators") and offers no flag to disable it, the engine must either:
>  - **(preferred) use integer fixed-point arithmetic** for all position, advance, and dimension math — e.g. `int64` in 1/1000 pt units, following the model of TeX's `scaled point` and FreeType's 26.6/16.16 fixed formats. Integer arithmetic is exact and contraction is not applicable, which makes the guarantee *structural* rather than *disciplinary*; or
>  - **(fallback) confine float arithmetic to a `measure` package** whose every multiply-add is written with explicit `float64()` rounding barriers, enforced by a project-specific `go vet`/analysis pass in CI that fails on any un-barriered `a*b+c` pattern in that package.
>
> **NFR1.b — Numeric emission is quantized.** All numbers written into the PDF must be emitted through a single normative formatter at a fixed decimal precision (recommended: 3 dp for coordinates, sufficient at 72 dpi where 0.001 pt ≈ 0.35 µm). Free-form `%v`/`FormatFloat(…, -1, …)` of a float must never reach the output stream.

I strongly recommend the fixed-point option. It converts NFR1 from "a rule every future contributor must remember" into "a property of the type system." Given C3 ("any determinism exception carved out to ship a feature — NFR1 is the product; erosion is fatal"), a disciplinary guarantee is inconsistent with the PRD's own stated stance.

---

## 3. Finding F2 — Go's `math` library is not bit-identical across GOARCH

**Severity: CRITICAL** if the engine uses transcendentals; **LOW** if it does not. The PRD does not say which, so it is currently unbounded risk.

### The evidence

I hashed 4,000 evaluations of each `math` function across the three targets. Results (`✓` = all three agree):

| Function | arm64 | amd64 | wasm | Verdict |
|---|---|---|---|---|
| `Sqrt` | A | A | A | ✓ safe |
| `Floor`, `Ceil`, `Round`, `Trunc` | A | A | A | ✓ safe |
| `Cbrt`, `Mod` | A | A | A | ✓ safe |
| `Sin`, `Cos`, `Tan`, `Asin`, `Atan`, `Atan2` | A | B | B | **arm64 differs** |
| `Log`, `Log2`, `Log10`, `Log1p`, `Exp2`, `Gamma` | A | B | B | **arm64 differs** |
| `Exp` | A | **B** | A | **amd64 differs** |
| `Pow` | A | B | **C** | **all three differ** |
| `Sinh`, `Erf` | A | B | **C** | **all three differ** |

Two distinct mechanisms are visible:

1. **FMA contraction inside pure-Go polynomial kernels.** `Sin`, `Cos`, `Log` etc. have no arm64 assembly in `$GOROOT/src/math` — they diverge because their Horner-scheme polynomials get fused on arm64. Same root cause as F1.
2. **Per-architecture assembly implementations.** `$GOROOT/src/math` contains `exp_amd64.s`, `exp_arm64.s`, `log_amd64.s`, `floor_wasm.s`, and a large `*_s390x.s` family. `math.Exp` is the clean demonstration: `exp_amd64.s` produces a result that differs by 1 ULP from both the arm64 assembly and the pure-Go wasm path. **This one cannot be fixed by rounding barriers in Folio's code** — the divergence is inside the standard library.

`Pow`, `Sinh` and `Erf` produce three *different* answers on three targets, because they compose both mechanisms.

### Why this matters for Folio

A report renderer that only does boxes, glyph advances, and line breaking needs `+ - * /`, comparison, `Sqrt`, and `Floor`/`Ceil`/`Round` — **all of which I verified are identical across all three targets** (IEEE 754 requires correctly-rounded results for `+ - * / sqrt`, and Wasm mandates the same).

It starts needing transcendentals the moment someone implements rotated text, a gradient, an arc/rounded rectangle, a dash-phase calculation, a chart (out of scope, but the addendum protects for SVG/PNG renderers), or a `formatNumber()` that reaches for `math.Pow(10, n)` to do decimal scaling. That last one is a genuinely likely accident — my own quantization helper used `math.Pow(10, n)` before I noticed.

### Recommendation

Add to NFR1:

> **NFR1.c — Restricted numeric surface.** The rendering path may use only floating-point operations that IEEE 754 requires to be correctly rounded (`+`, `-`, `*`, `/`, `sqrt`) plus exact operations (`Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod`). Use of `math.Sin/Cos/Tan/Log/Exp/Pow` and other transcendental functions in the rendering path is **prohibited**, because Go's implementations are not bit-identical across GOARCH (verified: `Exp` differs between amd64 and arm64; `Pow` differs across all three of amd64/arm64/wasm). If a future feature requires a transcendental, it must ship with a vendored, architecture-independent pure-Go implementation written with explicit rounding barriers.

Add an import-restriction lint (`depguard` or a `go vet` analyzer) banning `math` transcendentals outside an allow-listed shim package. This is cheap and catches the accident class above.

---

## 4. Finding F3 — Compression and PDF structure are NOT the risk the PRD thinks they are

**Severity: LOW (currently mis-prioritized as high).** Good news, but the PRD should be corrected so architecture does not spend its risk budget here.

### The claim in the PRD

> "It requires pinning every source of variation: document creation timestamp, document ID, internal object ordering, compression output, and font subset generation." (NFR1)

Of these five, **compression output and internal object ordering are not sources of variation at all** in a correctly-written Go program, and the first two are trivially pinnable.

### The evidence — `compress/flate` is byte-stable

I compressed a 1.2 MB synthetic PDF content stream at levels 1, 6, 9, and `HuffmanOnly`, plus `zlib`, and hashed the output.

- **Across architectures** (arm64, amd64, wasm at Go 1.26): byte-identical at every level.
- **Across Go versions** (1.21.13, 1.22.12, 1.23.4, 1.24.0, 1.25.0, 1.26.0 on arm64): byte-identical at every level.

`compress/flate` is pure Go with no architecture-specific assembly and no parallelism; its output is a deterministic function of input and level. It is *observably* stable over a five-year span of Go releases.

**The caveat that matters:** this is stability by observation, not by contract. The [Go 1 compatibility promise](https://go.dev/doc/go1compat) says nothing about compressor output bytes — it guarantees API and behavioural compatibility, explicitly disclaiming performance and saying nothing about byte-level encoder output. Go's DEFLATE encoder has been substantially reworked in the past (the compression-performance work merged around the Go 1.7 era), and any such rework can change output bytes without breaking any API. A future rewrite would silently invalidate every recorded golden hash in every downstream user's CI.

### Object ordering and xref

"Internal object ordering" varies only if the writer iterates a Go `map` to emit objects, resources, or font glyph sets — Go randomizes map iteration order deliberately. I confirmed how aggressive this is: iterating the same 6-key `map[string]int` six times **within a single process** gives six different orders:

```
[Im0 GS0 F3 Im1 F1 F2] [F2 Im0 GS0 F3 Im1 F1] [F1 F2 Im0 GS0 F3 Im1]
[GS0 F3 Im1 F1 F2 Im0] [F1 F2 Im0 GS0 F3 Im1] [GS0 F3 Im1 F1 F2 Im0]
```

So a PDF resource dictionary emitted by map iteration would not be reproducible **between two consecutive renders on the same machine**, let alone across targets. This is a **self-inflicted** hazard with a one-line fix (sort keys, or use insertion-ordered slices). It is not an inherent PDF property. Likewise xref offsets are a deterministic function of the bytes already written.

Usefully, `encoding/json` *does* sort map keys when marshalling, which I verified — so `.folio` template round-tripping through `map[string]any` is stable. The hazard is confined to hand-written emission code.

Similarly: if the engine ever parallelizes page layout across goroutines and collects results without deterministic ordering, output becomes nondeterministic *on the same machine*. Worth a one-line prohibition.

### Recommendation

Rewrite NFR1's second paragraph. It currently misdirects architecture. Replace with:

> This is stronger than layout equivalence. It requires pinning: the document creation and modification timestamps, the trailer `/ID`, and any font subset tag — all of which are inputs the engine chooses; and it requires the engine to avoid self-inflicted nondeterminism: no `map` iteration in emission order, no clock or environment reads, no unordered concurrency in the layout or writing path. Compression output (`compress/flate`) and xref offsets are deterministic functions of their input and are not independent risks, but the DEFLATE encoder's exact output is not covered by the Go 1 compatibility promise — see NFR6.

And add to NFR6 (versioning), which currently does not contemplate this at all:

> **NFR6.a** — Byte-reproducibility (NFR1) is guaranteed only for a fixed `folio-go` version **compiled with a declared Go toolchain version range**. The Go toolchain version is part of the reproducibility contract and must be recorded in release metadata and in the golden-hash fixtures. Consider vendoring the DEFLATE encoder (or pinning `github.com/klauspost/compress` at an exact version) so that a Go upgrade cannot silently change output.

This last point is a real product decision the PRD currently hides: **a `folio-go` user's recorded SHA-256 will break when they upgrade Go**, unless the encoder is vendored. UJ-2 has Anan writing exactly that test. He deserves to be told the conditions under which it holds.

---

## 5. Finding F4 — `formatDate()` is a determinism bomb, and the clock is unmentioned

**Severity: HIGH.** Cheap to fix, catastrophic if missed, and currently invisible in the PRD.

### The evidence

FR17 specifies `formatDate()` with a pattern argument, and FR20 specifies `{{params.reportDate}}`. The PRD never says what timezone `formatDate` resolves in. Go's default for `time.Time.Format` on a parsed timestamp depends on the `Location` carried by that value, and `time.Local` is read from the host: `/etc/localtime` or `$TZ` natively, and from the JavaScript environment under `js/wasm`.

I tested formatting the instant `2026-08-22T23:30:00Z` as `dd/MM/yyyy` under different host timezones:

| Host TZ | `t.Local()` renders | `t.UTC()` renders |
|---|---|---|
| `UTC` | `22/08/2026` | `22/08/2026` |
| `America/Los_Angeles` | `22/08/2026` | `22/08/2026` |
| `Asia/Bangkok` | **`23/08/2026`** | `22/08/2026` |
| `Pacific/Kiritimati` | **`23/08/2026`** | `22/08/2026` |

A Thai user previewing in a browser in `Asia/Bangkok` and a Linux container running `UTC` render **different dates** — different visible content, different string widths, potentially different line wrapping, and certainly a different hash. This defeats NFR1, NFR5, S2, S3 and S5 simultaneously, and unlike F1 it is *visible to the end recipient of the statement*.

There is a second, related hazard the PRD never addresses: **nothing forbids the engine from reading the clock.** If `folio-go` defaults `/CreationDate` to `time.Now()` — which is what essentially every PDF library does by default — NFR1 fails on the first render, trivially. The PRD lists "document creation timestamp" as something to pin but never states the rule.

### Recommendation

Add an explicit NFR:

> **NFR9 — The rendering path is a pure function.** `folio-go` must not read the system clock, the host timezone, the locale, environment variables, the filesystem, or the network during rendering. Render output must be a pure function of (template, data, parameters, library version). Specifically:
>  - `/CreationDate` and `/ModDate` default to a fixed epoch, overridable only by an explicit caller-supplied value (honour `SOURCE_DATE_EPOCH` semantics if a default is wanted).
>  - The trailer `/ID` is derived deterministically — e.g. the SHA-256 of the document body — never from a clock or RNG.
>  - `formatDate()` resolves in UTC unless the template declares an explicit IANA timezone; it must never consult `time.Local`.
>  - `formatNumber()` / `formatDate()` patterns resolve against a locale declared in the template, never the host locale. (The PRD already notes in §10 that no locale model exists — this is the requirement that makes the gap load-bearing rather than cosmetic.)

This also resolves an assumption the PRD flags but leaves open in §10 ("No i18n beyond script rendering… does not provide a locale model for `formatDate`/`formatNumber`"). That gap is not merely a missing feature — it is an NFR1 violation.

---

## 6. Finding F5 — `Page X of Y` (FR30) directly contradicts streaming (FR38 / NFR4 / S6)

**Severity: HIGH.** This is a genuine architectural contradiction, not a gap.

### The claim

- **FR30**: render `Page X of Y`, mechanism deferred to architecture.
- **FR38 / NFR4 / S6**: "Render to a writer as a stream, so large reports avoid temporary files"; "Large reports must stream to a writer rather than materialising fully in memory or on disk"; S6 = "50-page render streams without materialising the whole document — Verified".

These cannot both hold in their strong forms. `Page X of Y` on page 1 requires `Y`, which is only known after the last page is laid out. Therefore either:

1. **Two-pass layout** — lay out everything to a page model, then write. The *layout* is fully materialised in memory; only the PDF byte-writing streams. Or
2. **Buffer the content streams** and patch page-count placeholders before writing. Same memory profile.
3. **PDF-level trickery** — a form XObject holding the page count, referenced by every page and written last. This *does* permit true single-pass streaming, but it is an unusual construction, complicates the deterministic-object-ordering rule, and interacts badly with text measurement: the string "Page 1 of 7" and "Page 1 of 128" have different widths, so any centred or right-aligned page-number field cannot be positioned until `Y` is known. Option 3 only works for left-aligned page numbers in a fixed-width font, which is not a constraint the PRD accepts.

The addendum records this as tension #5 but frames it as "never discussed as a constraint on the pipeline" — understating it. It is not just unconstrained; it is contradictory with a *Must*-adjacent NFR and a success criterion.

Note also that the addendum's own pipeline (`Template → Expression Evaluation → Layout Calculation → Page Model → PDF Renderer`) and its mandated `Page` struct **already imply full materialisation of the page model**. The architecture the addendum specifies is inherently two-pass. S6 as written is unsatisfiable by the architecture the addendum mandates.

### Recommendation

Resolve this in the PRD rather than deferring it, because it changes what S6 tests:

> **FR38 (revised)** — Render to an `io.Writer` incrementally, so that the PDF byte stream is never buffered in full and no temporary files are used. The *layout page model* is materialised in memory; the *output bytes* are streamed.
>
> **NFR4 (revised)** — Peak memory must scale with the page model, not with the size of the output PDF. Working target: a 50-page, 5,000-row statement renders in bounded memory without temporary files.
>
> **S6 (revised)** — 50-page render writes incrementally to an `io.Writer` with no temporary file and no full-document buffer; peak RSS recorded and bounded.

If genuinely constant-memory streaming for arbitrarily large reports is wanted, that is a different product with a different pagination model (no `Page X of Y`, no forward-referencing aggregates) and should be stated as out of scope for v0.1.

---

## 7. Finding F6 — `{{sum(transactions.amount)}}` in a page header contradicts streaming too

**Severity: MEDIUM.** Same family as F5, worth catching now.

FR18 permits aggregation over a collection, and FR6 puts a Page Header on every page. Nothing forbids `{{sum(transactions.amount)}}` in the page header — indeed a statement's header commonly carries a closing balance. Any aggregate referenced *before* the rows that feed it forces a full traversal of the data before page 1 can be emitted.

Combined with FR13 (data arrives as JSON from the caller — `map[string]any` per the addendum, i.e. already fully in memory), this means the **data** is fully materialised regardless. NFR4's "streaming" was never going to reduce data-side memory; it only concerns output-side memory. The PRD should say so, or an integrator will read NFR4 as a promise about handling a million-row dataset, which the `folio.Render(template, data any)` signature already forecloses.

### Recommendation

State the memory model honestly in NFR4: input JSON is fully materialised by the caller and by `folio-go`; the page model is materialised; only the output byte stream is incremental. If large-dataset streaming input is a future goal, note it as architecturally protected (an `iter.Seq` / row-cursor data source) but out of scope for v0.1.

---

## 8. Finding F7 — Q1 is asked at the wrong altitude; the answer is "hand-roll the writer," and that is fine

**Severity: MEDIUM (scoping).**

Q1 asks: "Can byte-identical output be achieved across WASM and native targets with an available PDF approach, or must the PDF writer be hand-rolled?"

Given F1–F4, the PDF *writer* is the least of the problem. But the answer is still "hand-roll it," for a reason the PRD hasn't articulated: **NFR1 requires control of the layout engine's arithmetic, and any library that does its own text measurement and pagination takes that control away.** Folio needs a PDF *serializer* (objects, streams, xref, font embedding), not a PDF *document library* (which is what gofpdf/maroto/gopdf are — they own layout).

Writing a conforming PDF 1.7 serializer for the feature set in scope — text with embedded Type0/CIDFontType2 fonts, images (XObjects), lines, rectangles, a page tree, and Flate-compressed streams — is a genuinely tractable, well-specified body of work, on the order of a few thousand lines. It is dramatically simpler than the layout engine sitting above it. The PRD should say this plainly so the architecture phase does not spend weeks evaluating libraries against a bar none of them were built to meet.

What the PDF spec actually requires here is permissive:

- `/CreationDate` and `/ModDate` live in the Info dictionary and are **optional** — they can be omitted entirely, which is the cleanest determinism answer.
- The trailer `/ID` is a two-element array of byte strings. The spec *recommends* deriving it from file characteristics including the current time, but that is guidance for uniqueness, not a conformance requirement — a deterministically derived `/ID` (e.g. hash of document content) is conforming. Note that `/ID` feeds the encryption key derivation, so if PDF encryption is ever added, a content-derived `/ID` must be computed before encryption.
- Font subset tags (`ABCDEF+Helvetica`) are six uppercase letters, conventionally arbitrary. Deriving them from a hash of the subsetted glyph set is both conforming and deterministic — this is the standard trick and should be stated as a requirement, not left to chance.

### Recommendation

Replace Q1 with a decision already taken:

> **D1 (decided)** — `folio-go` implements its own PDF serializer. No third-party PDF document library is used in the rendering path, because NFR1 requires the engine to own text measurement, pagination, and numeric emission, and because no surveyed Go PDF library exposes control over all of `/ID`, timestamps, object ordering, and subset-tag generation. The serializer targets PDF 1.7, omits `/CreationDate`/`/ModDate` by default, derives `/ID` and font subset tags by hash of content, and emits all numbers through a single quantizing formatter.

Then open a narrower question: whether a *font* library (parsing, subsetting) can be used as a dependency, which is the real remaining build-vs-buy call.

---

## 9. The strongest achievable version of NFR1

NFR1 as written — "regardless of operating system, machine, or compilation target" — is achievable, but only if the PRD stops asserting it and starts constraining the engine. Here is the version I would sign off on:

> ### NFR1 — Byte-reproducible rendering *(the signature requirement)*
>
> For a fixed `folio-go` version, built with a Go toolchain in the declared supported range, the same template + data + parameters **must produce a byte-identical PDF** — the same SHA-256 — on every supported target, including `js/wasm` in a browser and native `linux/amd64`, `linux/arm64`, and `darwin/arm64`.
>
> This is stronger than layout equivalence and is achieved by construction, not by testing. It requires all of:
>
> - **NFR1.a — Contraction-free arithmetic.** All layout, measurement, and positioning arithmetic uses `int64` fixed-point (1/1000 pt). Where floating point is unavoidable, every multiply-add carries an explicit `float64()` rounding barrier, enforced by a CI analyzer. Rationale: the Go spec permits FMA contraction, arm64 performs it and wasm forbids it, and Go offers no flag to disable it.
> - **NFR1.b — Quantized emission.** All numbers reach the PDF through one formatter at fixed precision (3 dp). No `%v` or shortest-representation float formatting in the output path.
> - **NFR1.c — Restricted numeric surface.** Only IEEE-correctly-rounded operations (`+ - * / sqrt`) and exact operations (`Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod`). Transcendental `math` functions are prohibited in the rendering path; Go's implementations differ across GOARCH.
> - **NFR1.d — Purity.** No clock, timezone, locale, environment, filesystem, or network reads during render (see NFR9). `/CreationDate` and `/ModDate` omitted by default; `/ID` and font subset tags derived by hash of content.
> - **NFR1.e — Ordered emission.** No `map` iteration order, and no unordered concurrency, may influence output. Object numbering follows a deterministic traversal.
> - **NFR1.f — Vendored encoders.** The DEFLATE implementation and the font subsetter are vendored or version-pinned, so that a Go toolchain upgrade cannot change output bytes.
> - **NFR1.g — Declared scope of the guarantee.** Reproducibility holds across OS, machine, and compilation target for a fixed (`folio-go` version, Go toolchain range, embedded font set version, Thai dictionary version). Changing any of these four may change output and is a breaking change under NFR6.
>
> **Verification:** CI must build the golden report on `linux/amd64`, `linux/arm64`, `darwin/arm64`, and `js/wasm`, and assert a single shared SHA-256 across all four. A three-target CI matrix that omits `arm64` **does not test this requirement** — arm64 is the only target that performs FMA contraction, and is where the guarantee will break first.

That last sentence is the single most important operational recommendation in this review. The PRD's S3 says "macOS dev, Linux container" — if the Linux container is amd64, that pair plus WASM would have *passed* on the naive implementation while `darwin/arm64` silently disagreed. **The test as specified would not have caught the defect I found in ten minutes.**

---

## 10. Other requirements that are harder than they read

*(The two HIGH-severity font/script findings, F15 and F16, have their own section — see Section 11.)*

### F10 — FR11 "readable, diffable, and mergeable in Git" vs FR32 "embed images inside the `.folio` template"
**Severity: MEDIUM.** These fight each other. FR32 requires embedding image bytes in the template; FR10/FR11 require the template to be human-readable, diffable, and *mergeable*. A base64-encoded logo is a single multi-megabyte JSON string line — it is not diffable, it defeats line-based merge, and it will dominate the file. Every change to the template carries the blob through Git history.

**Recommendation:** state the tradeoff and pick. Either (a) accept it and require images be stored as a separate top-level `"assets"` map with one asset per line, keyed by content hash, so that layout edits produce clean diffs and asset changes are isolated; or (b) make `.folio` a container (a zip, like `.docx`/`.odt`) — which forfeits FR11's plain-text diffability entirely and should then be stated as such. Option (a) is the better fit for the stated values. Note also that FR32's determinism rationale is sound and should be kept.

### F11 — FR34's "the previewed document is the production document" is stronger than NFR1 supports
**Severity: MEDIUM.** FR34 and NFR5 promise the preview *is* production. But the browser preview renders with the fonts embedded in the WASM bundle, while the server render uses whatever font set that `folio-go` build carries. If a user supplies a custom font (NFR7, unspecified), the designer must embed that font's bytes into the `.folio` file or the preview and production diverge — and NFR7 explicitly flags custom font supply as unspecified. This is the same class of problem as FR32 and needs the same answer: **fonts referenced by a template must be resolved from content the template carries or from a versioned set the library pins — never from the host.** Currently the PRD says this for the Thai dictionary (NFR3) but not for fonts.

### F12 — FR40's error contract vs NFR1
**Severity: LOW.** Sound requirement, no conflict, but note that "located, actionable error naming the template element and the data path" implies stable element identity in `.folio`. If elements are addressed by array index, error messages and template diffs both degrade. Recommend requiring stable element `id` fields in the format (FR10), which also improves FR11 mergeability.

### F13 — C4 vs Q6 is an unresolved contradiction the PRD notices but does not resolve
**Severity: LOW (process).** C4 makes "designer work starting before the engine renders the golden report" a counter-metric — i.e. a failure signal — while Q6 asks whether the designer should really come last given it is success criterion #1. The PRD is simultaneously enforcing and questioning the same sequencing rule. Pick one before epics are written; as written, a story-planning agent could reasonably produce either order and cite the PRD for it.

### F14 — Q7 (licence) is load-bearing for NFR7 and should be pulled forward
**Severity: MEDIUM.** Q7 is marked "Deferred; revisit before any public repo." But NFR7 requires shipping embedded fonts with CJK + Thai coverage, and font licences (even the permissive SIL OFL that Noto uses) impose conditions on redistribution and, for OFL, on *bundling under a reserved name*. The font set decision (Q2) cannot be finalised without the distribution decision (Q7). Recommend promoting Q7 above Q9 in the blocking order.

---

## 11. Fonts, text shaping, and the WASM bundle (NFR3, NFR7, FR31, FR34, Q2, Q9)

### F15 — Q9 is not an open question; it is a trilemma, and the answer as specified is ~13 MB
**Severity: HIGH.**

Q9 asks "What is the WASM bundle size ceiling, given embedded fonts and a Thai breaking dictionary?" and marks the impact "may be significant." I measured it. It is significant.

**Measured: the Go/WASM engine itself.** I compiled a skeleton with `folio-go`'s realistic dependency surface — `encoding/json`, `golang.org/x/image/font/sfnt`, `compress/zlib`, `crypto/sha256`, `syscall/js`, sorting and string building — for `GOOS=js GOARCH=wasm` on Go 1.26:

| Artifact | Size |
|---|---|
| raw `.wasm` | 3.83 MB |
| `gzip -9` | 1.05 MB |
| `brotli -q11` | **0.78 MB** |
| stripped (`-ldflags="-s -w"`) | 3.75 MB raw / 1.03 MB gzip |

That is a *skeleton*. A real engine adds the layout engine, pagination, table logic, expression evaluator, text shaper, and font subsetter. A realistic estimate is 6–10 MB raw, **1.5–2.5 MB brotli**. That alone is acceptable — comparable to a mid-sized SPA, and it is the *good* news in this section.

**Measured: the fonts.** This is where it collapses. Real font files on this machine:

| Font | Raw | brotli -q11 |
|---|---|---|
| `Arial Unicode.ttf` — single face, Latin + CJK + Thai | 22.20 MB | **11.74 MB** |
| `Hiragino Sans GB.ttc` — CJK only | 22.43 MB | 15.79 MB |
| `Songti.ttc` | 63.83 MB | — |
| `STHeiti Light.ttc` | 53.20 MB | — |

**Fonts barely compress** — roughly 47% off with brotli at maximum quality, because glyph outline data is already dense. This is the number that decides Q9: a single pan-Unicode face covering NFR3's Latin + CJK + Thai costs **~12 MB over the wire, compressed**. Noto Sans CJK is in the same class — the full multi-language collection runs to the high tens/low hundreds of MB, and a single-language, single-weight cut such as Noto Sans SC Regular is on the order of 10–20 MB raw. *(These Noto figures are indicative, not measured here; the Arial Unicode and Hiragino numbers above are measured.)*

**Total for the PRD as written:** ~0.8–2.5 MB engine + ~12 MB fonts + Thai dictionary = **~13–15 MB before the user sees a canvas.** On a typical connection that is tens of seconds. It does not sink the *product*, but it decisively sinks the **"open it in a browser like draw.io"** framing in §5.1 and FR8, which is a stated positioning pillar in §2.

### The trilemma

Three requirements cannot all hold in their strong forms:

1. **FR31 / NFR7** — embed all fonts, so the PDF renders anywhere; Latin + CJK + Thai coverage.
2. **FR34 / NFR5 / NFR1** — the WASM preview is byte-identical to production, which requires the browser to have *exactly* the font bytes the server will use.
3. **FR35 / NFR8** — the designer and its exact preview operate **fully offline**, with nothing leaving the machine.

Lazy-loading fonts on demand (the obvious fix for #1) breaks #3 on first use, and complicates #2 unless the font set is content-addressed and version-pinned.

### Recommendation

Resolve Q9 in the PRD rather than deferring it, and split the requirement:

> **FR34 (revised)** — The designer offers an exact PDF Preview produced by `folio-go` compiled to WebAssembly. The engine module and the *core Latin font set* load with the application. **CJK and Thai font sets load on demand**, the first time a template references them, and are cached persistently (Cache Storage / service worker) so that subsequent sessions are fully offline.
>
> **FR35 (revised)** — After first load of the font sets a template requires, the designer and its exact preview operate fully offline. Neither template nor data ever leaves the user's machine at any point, online or offline. *(The privacy guarantee of NFR8 is preserved unconditionally; only the availability guarantee is conditioned on a first fetch.)*
>
> **NFR7 (revised)** — `folio-go` ships a versioned font set. Font identity in a `.folio` template is a (family, version, content-hash) triple, and both the native library and the WASM build resolve it to identical bytes. A font the library does not carry must be embedded in the template by the designer, or the render fails with a located error (FR40) — never a silent substitution, which would violate NFR1.
>
> **Q9 (answered)** — Measured: Go/WASM engine ≈ 0.8–2.5 MB brotli; a pan-Unicode CJK+Thai face ≈ 12 MB brotli. Eager bundling gives a ~13–15 MB first load and is rejected. Lazy, cached, per-script font loading is adopted.

Two further notes for architecture:

- **Consider subsetting the shipped fonts at build time** to the CJK ranges that enterprise statements actually need. A Chinese statement realistically needs GB 2312 level 1 (~3,755 hanzi) rather than all ~30,000 CJK glyphs, which cuts the font by roughly an order of magnitude. This must be a *build-time, versioned* subset so that NFR1 still holds — never a runtime dynamic subset driven by observed data, which would make the font bytes data-dependent.
- **TinyGo is not a viable size fix here.** It produces much smaller binaries but its reflection support is incomplete, which `encoding/json` depends on heavily, and its runtime/goroutine semantics differ from upstream Go. More decisively for this project: using a different compiler for the WASM build **directly threatens NFR1**, since it is a different code generator with different floating-point lowering. The WASM and native builds must come from the same toolchain. State this as a constraint.

### F16 — NFR3 (Thai) is one line of PRD hiding the second-largest engineering item in the MVP
**Severity: HIGH.**

NFR3 says Thai needs "dictionary-based line breaking, as Thai does not delimit words with spaces," and requires the dictionary be "embedded and versioned with the library, never taken from the host platform." That instinct is exactly right and is the correct NFR1 posture. But the requirement understates the work in three ways:

1. **Thai needs shaping, not just breaking.** Thai stacks vowels and tone marks above and below base consonants (e.g. U+0E48–U+0E4B tone marks over U+0E31/U+0E34–U+0E37 vowels over a consonant). Correct rendering requires OpenType GSUB/GPOS mark positioning — a HarfBuzz-class shaper. Naive per-codepoint glyph placement produces visibly broken Thai. NFR3 says "measure, break, and render" without acknowledging that "render" here means shaping.
2. **cgo is excluded by NFR1**, so HarfBuzz proper is unavailable. A pure-Go shaper is required — one does exist, see F17.
3. **Dictionary-based segmentation is ambiguous.** Thai word segmentation has genuinely ambiguous cases; different algorithms (longest-matching vs. maximal-matching vs. ICU's dictionary+model approach) give different breaks. Whichever is chosen becomes part of the reproducibility contract (NFR1.g) — and changing it later changes pagination, which changes output, which breaks every downstream golden hash. This is a **one-way door** and should be recognised as such.

**Recommendation:**

> **NFR3 (revised)** — MVP must correctly measure, shape, break, and render Latin, CJK, and Thai text.
>  - Latin — whitespace/UAX #14 line breaking, standard OpenType shaping (kerning, basic ligatures).
>  - CJK — per-character breaking with UAX #14 prohibition rules (no line-start closing punctuation, no line-end opening punctuation).
>  - Thai — OpenType GSUB/GPOS mark positioning for vowel/tone stacking, plus dictionary-based word segmentation for line breaking.
>
> The shaper, the line-breaking algorithm, the segmentation dictionary, and any Unicode data tables are **embedded and versioned with the library**, never taken from the host platform, and are named in the reproducibility contract (NFR1.g). Changing any of them is a breaking change under NFR6, because it changes pagination and therefore output bytes.
>
> **Add to §7.2:** an S4 that is testable rather than "verified by inspection" — a fixture set of Thai and CJK strings with *expected break positions* recorded, so segmentation regressions are caught mechanically.

Also: promote this to a first-class risk in the build order. The addendum's Phase 1 lists "font handling" as a Phase 1 item and Phase 2 lists "text wrapping" — but Thai shaping and segmentation are not mentioned in any phase, despite being a *Must* in NFR3 and appearing in the golden report (S4, and UJ-1's "Thai customer names wrapping correctly"). That is unscoped work on the critical path, the same defect the addendum already identifies for the Preview API (tension #2).

### F17 — The pure-Go shaping stack exists and is good; the Thai *word breaker* does not exist

**Severity: HIGH**, but with a much clearer path than F16 alone suggested. I verified all of this by building and running it, not by reading about it.

**The good news — shaping is solved.** `github.com/go-text/typesetting` v0.3.4 is a **pure-Go port of HarfBuzz**:

- Contains a full `harfbuzz/` package including **`ot_shape_thai.go`** — a dedicated Thai shaper — plus `ot_shape_complex.go`, `ot_layout.go` and OpenType language tables.
- **Zero cgo** (verified: no `import "C"` anywhere in the module).
- **Builds cleanly for `GOOS=js GOARCH=wasm`** (verified by an actual build against `shaping`, `font`, `segmenter`, `di`).
- Ships a `segmenter/` package implementing **UAX #14** line breaking (`unicode14_rules.go`) and **UAX #29** word/grapheme breaking (`unicode29_rules.go`).

So NFR3's shaping requirement, and its Latin and CJK breaking requirements, are satisfiable with an existing, maintained, cgo-free, wasm-safe dependency. This substantially de-risks the "HarfBuzz-equivalent without cgo" question — the answer is yes, it exists.

**The bad news — Thai line breaking is genuinely absent.** I ran the segmenter on three strings and counted line-break opportunities:

| Input | Runes | Break opportunities found | Correct? |
|---|---|---|---|
| `Hello my name is Somchai` | 24 | 5 | ✅ correct (whitespace) |
| `你好我叫史密斯` | 7 | 7 | ✅ correct (per-character) |
| `สวัสดีครับผมชื่อสมชาย` ("Hello, my name is Somchai" — 5 Thai words, no spaces) | 21 | **1** | ❌ **the entire string is one unbreakable unit** |

The cause is visible in the source: `segmenter/unicode14_rules.go:312` — *"apply rule LB1 to resolve break classes AI, SG, XX, SA and CJ"* with `case ucd.LB_SA:`. Thai is Unicode line-break class **SA (Southeast Asian)**, and UAX #14's rule LB1 resolves SA to AL/ID — i.e. **no internal break opportunities**. UAX #14 explicitly defers SA to dictionary-based resolution and does not provide one. This is correct UAX #14 behaviour, and it is useless for Thai.

Concretely: **a Thai customer name or address in the golden report would not wrap at all** — it would overflow its cell or be clipped. UJ-1's beat "Thai customer names wrapping correctly" does not happen with the off-the-shelf stack.

**What exists to fill the gap.** The only pure-Go Thai segmenter I could locate and fetch is `github.com/narongdejsrn/go-thaiwordcut` (MIT, last commit **2019 — unmaintained for ~7 years**). Inspecting it:

- Bundles the **LEXiTRON** dictionary: **1,061,924 bytes / 42,221 words**.
- Algorithm is **greedy longest-match forward** (`findSegment` walks forward taking the longest trie match, advancing one byte on failure) — the simplest and least accurate approach, materially weaker than ICU's dictionary-plus-model method for Thai, and it produces known errors on ambiguous segmentations.
- **Loads the dictionary from disk at runtime** via `os.Open(filePath)` resolved through `runtime.Caller`. This is **fatal for `js/wasm`** (no filesystem) and directly violates NFR3's own rule that the dictionary be "embedded and versioned with the library, never taken from the host platform." Using it requires forking it to `go:embed`.
- The LEXiTRON dictionary carries **its own licence from NECTEC, separate from the wrapper's MIT** — a distinct obligation that feeds Q7/Q2.

**A useful correction to Q9:** the PRD pairs "embedded fonts and a Thai breaking dictionary" as jointly responsible for bundle size. Measured, the dictionary is **~1 MB** against **~12 MB** for one CJK-capable font. **Fonts dominate by an order of magnitude**; the dictionary is not a bundle-size concern at all. Q9's framing mis-attributes the cost.

### Recommendation

> **Adopt `github.com/go-text/typesetting` as the shaping and Latin/CJK line-breaking dependency**, pinned to an exact version and named in the NFR1.g reproducibility contract. It is pure Go, wasm-safe, and carries a Thai OpenType shaper.
>
> **Scope Thai word segmentation as first-party, build-it work** — it is the one NFR3 capability with no usable off-the-shelf pure-Go answer. Give it its own story in the build order (it currently appears in no phase). The minimum viable form is a vendored, `go:embed`-ed dictionary plus a segmentation algorithm; prefer **longest-matching with a documented tie-break rule** over greedy, and record the algorithm and dictionary version in the reproducibility contract, because changing either changes pagination and breaks every downstream golden hash (a one-way door, per F16).
>
> **Add to §7.2:** an S4 fixture set of Thai strings with *expected break positions*, so segmentation is tested mechanically rather than "verified by inspection."
>
> **Add to Q2/Q7:** the Thai dictionary's licence (e.g. LEXiTRON/NECTEC) is a distribution obligation distinct from the font licences and must be settled alongside them.

---

## 12. Survey of Go PDF libraries against NFR1 (answering Q1 with primary evidence)

I fetched the candidate libraries into the module cache and inspected the actual source and build behaviour rather than relying on documentation.

| Library | Version | cgo? | Builds for `js/wasm`? | Pin `CreationDate`? | Trailer `/ID` | Clock / RNG in path |
|---|---|---|---|---|---|---|
| `github.com/go-pdf/fpdf` | v0.9.0 | **No** (verified: zero `import "C"`) | **Yes** (verified: builds clean) | **Yes** — `SetCreationDate`, `SetModificationDate`, `SetDefaultCreationDate` | Emits constant `/ID [()()]` — deterministic by construction | `time.Now()` default for creation date (`fpdf.go:4874`); `math/rand` only in `protect.go` (encryption) |
| `github.com/signintech/gopdf` | v0.38.0 | **No** | **Yes** (verified: builds clean) | No such API found | Not emitted | `rand.Seed(time.Now()…)` in `pdf_protection.go` (encryption only) |
| `github.com/pdfcpu/pdfcpu` | v0.15.0 | **No** | not tested | No generator API — it is a *processor* | Reads/validates `/ID` | 24 × `time.Now()`, 6 × rand across the package |
| `github.com/unidoc/unipdf` | — | — | — | — | — | **Commercial licence** — disqualified for an undecided-licence project (Q7) |
| `github.com/johnfercher/maroto` | v2 | — | — | — | — | Layout framework built *on top of* gofpdf; inherits its constraints and adds its own layout engine |

### What this actually shows

The surprise is that the PDF-writing layer is **not** the blocker. `go-pdf/fpdf` is pure Go, compiles to `js/wasm` today, lets you pin both timestamps, and already emits a *constant* `/ID [()()]`. Its own doc comment is telling:

> "SetCreationDate fixes the document's internal CreationDate value. By default, the time when the document is generated is used for this value. **This method is typically only used for testing purposes to facilitate PDF comparison.**"

The maintainers built exactly the hook NFR1 needs — but framed it as a test affordance, not a supported guarantee. Nothing in that library's contract promises byte-stability across versions.

### Why I still recommend hand-rolling the serializer

Not because the PDF layer is hard, but because of **what these libraries take away**:

1. **They own text measurement and layout.** `fpdf.CellFormat`, `MultiCell`, and friends do their own width computation and line breaking in `float64`, with no rounding barriers and no control over emitted numeric precision. NFR1.a/1.b/1.c cannot be enforced from outside the library. This is decisive.
2. **Font support does not reach NFR3.** `fpdf` uses a font-definition-file model oriented at Latin Type1/TrueType; full CID/Type0 embedding with OpenType GSUB/GPOS shaping for Thai is not what it does. `gopdf` has better CID support but still no shaper.
3. **No subsetting-determinism contract.** None of them documents subset-tag generation as deterministic.
4. **They are the wrong shape.** The addendum already mandates an intermediate page model and explicitly rejects "Template → Draw Directly Onto PDF." A document library wants to be driven imperatively; Folio wants a serializer fed by a finished page model.

**However** — `go-pdf/fpdf` is genuinely useful as a **differential oracle**: render simple fixtures through both fpdf and Folio's serializer and compare with a PDF parser to catch spec-conformance bugs early. Worth a line in the architecture, and cheaper than discovering conformance problems in Acrobat.

### Recommendation

Answer Q1 as decided (see F7/D1), and add the supporting rationale:

> No surveyed Go PDF library is disqualified by cgo or by WASM incompatibility — `go-pdf/fpdf` and `signintech/gopdf` are pure Go and compile for `GOOS=js GOARCH=wasm` today, and `go-pdf/fpdf` can pin both document timestamps and emits a constant trailer `/ID`. They are disqualified instead because they **own text measurement and layout**, which NFR1.a–1.c require `folio-go` to own, and because none supports CID/Type0 embedding with OpenType shaping at the level NFR3 requires. `folio-go` therefore implements its own PDF serializer, and may use `go-pdf/fpdf` in tests as a conformance oracle.

---

## 13. Findings summary

| # | Severity | Finding | Requirement affected |
|---|---|---|---|
| **F1** | **CRITICAL** | FMA contraction makes `darwin/arm64` produce different `float64` results from `js/wasm` in ordinary text-measurement code; the difference reaches the emitted PDF bytes. Go offers no flag to disable it. **Verified by execution.** | NFR1, S2, S3, S5 |
| **F2** | **CRITICAL** (conditional) | Go's `math` transcendentals are not bit-identical across GOARCH — `Exp` differs on amd64; `Pow`, `Sinh`, `Erf` differ across all three targets. `+ - * / Sqrt Floor Ceil Round Trunc Mod` are safe. **Verified by execution.** | NFR1 |
| **F4** | **HIGH** | `formatDate()` resolving in host-local time renders *different dates* (22 vs 23 Aug) in `Asia/Bangkok` vs `UTC`. No requirement forbids the engine from reading the clock, timezone, or locale. **Verified by execution.** | NFR1, NFR3, FR17, FR20 |
| **F5** | **HIGH** | `Page X of Y` (FR30) is contradictory with true streaming (FR38 / NFR4 / S6). The addendum's own page-model pipeline is inherently two-pass, making S6 unsatisfiable as written. | FR30, FR38, NFR4, S6 |
| **F15** | **HIGH** | Q9 measured: engine ≈ 0.8–2.5 MB brotli, but a pan-Unicode CJK+Thai face is **11.7 MB brotli** (fonts compress only ~47%). Eager bundling ⇒ ~13–15 MB first load, sinking the "open it like draw.io" positioning. FR31 + FR34 + FR35 form a trilemma. | Q9, FR31, FR34, FR35, NFR7 |
| **F16** | **HIGH** | NFR3's one line hides Thai OpenType mark-positioning (shaping, not just breaking), a pure-Go shaper requirement (cgo excluded ⇒ no HarfBuzz), and a segmentation-algorithm choice that is a one-way door baked into the reproducibility contract. Unscoped in every build phase. | NFR3, S4, build order |
| **F17** | **HIGH** | Shaping is solved: `go-text/typesetting` is a pure-Go HarfBuzz port, cgo-free, wasm-buildable, with a Thai shaper (**verified**). But its UAX #14 segmenter gives Thai **1 break opportunity in a 5-word string** (**verified**) — Thai will not wrap. The only pure-Go Thai segmenter is unmaintained since 2019, greedy-longest-match, and reads its dictionary from disk (fatal for wasm). Build-it work, unscoped. | NFR3, S4, UJ-1, build order |
| **F7** | **MEDIUM** | Q1 asked at the wrong altitude. The PDF layer is the easy part; hand-roll the serializer because layout control is required, not because PDF writing is hard. | Q1 |
| **F6** | **MEDIUM** | Aggregates in page headers force full data traversal; NFR4's "streaming" never applied to input data anyway, given `Render(tpl, data any)`. | NFR4, FR18 |
| **F10** | **MEDIUM** | FR32 (embed images in template) fights FR11 (readable, diffable, mergeable in Git). | FR10, FR11, FR32 |
| **F11** | **MEDIUM** | FR34's "the previewed document *is* the production document" is unsupported while custom-font provisioning (NFR7) is unspecified. Fonts must be template-carried or library-pinned, never host-resolved. | FR34, NFR5, NFR7 |
| **F14** | **MEDIUM** | Q7 (licence) blocks Q2 (font set) — font redistribution terms cannot be settled without the distribution decision. Promote Q7 above Q9. | Q2, Q7, NFR7 |
| **F3** | **LOW** (mis-prioritized as high) | `compress/flate` is byte-stable across arm64/amd64/wasm and Go 1.21→1.26 (**verified**). Object ordering is only a hazard via `map` iteration — which varies *within one process* (**verified**) — and is a one-line fix. Neither is an inherent risk. But flate stability is not covered by the Go 1 compatibility promise. | NFR1, NFR6 |
| **F12** | **LOW** | FR40's located errors imply stable element identity; recommend explicit `id` fields in `.folio`, which also helps FR11. | FR40, FR10 |
| **F13** | **LOW** | C4 enforces a build order that Q6 simultaneously questions. Unresolvable by a downstream planning agent. | C4, Q6, build order |

---

## 14. What I would change in the PRD, in priority order

1. **Rewrite NFR1** into the constrained form in Section 9 (NFR1.a–1.g). This is the single highest-value edit. As written, NFR1 is a wish; rewritten, it is a specification an architect can build against and a test suite can enforce.
2. **Add NFR9 (purity)** — no clock, timezone, locale, environment, filesystem, or network in the render path. Cheap, and closes F4 permanently.
3. **Fix the CI matrix requirement.** S2/S3 must name `linux/amd64`, `linux/arm64`, `darwin/arm64`, and `js/wasm`. **A matrix without arm64 does not test NFR1** — it is the only target that contracts FMA. Add this to §7.2 explicitly.
4. **Resolve F5** — decide that layout is two-pass and that "streaming" means the output byte stream, then restate FR38 / NFR4 / S6 accordingly.
5. **Answer Q9 with the measured numbers** and adopt lazy, cached, per-script font loading; restate FR35 as "offline after first font fetch," preserving NFR8's privacy guarantee unconditionally.
6. **Expand NFR3** to name shaping, the pure-Go shaper constraint, and the segmentation dictionary as a versioned, one-way-door component. Adopt `go-text/typesetting` for shaping and Latin/CJK breaking, and **create an explicit story for first-party Thai word segmentation** — it exists in no build phase today, has no usable off-the-shelf pure-Go answer, and without it Thai text does not wrap at all (F17).
7. **Close Q1 as decided** (own serializer, with `go-pdf/fpdf` as a test oracle) so architecture does not re-litigate it.
8. **Promote Q7 above Q9** in the blocking order, since font licensing gates the font set.
9. **Resolve C4 vs Q6** before epics are generated.

---

## 15. Closing assessment

The PRD is unusually good at knowing what it does not know — the `[ASSUMPTION]` register and the addendum's "internal tensions carried forward" are better than most PRDs manage. My criticism is not that it is vague; it is that **it has correctly identified NFR1 as the signature requirement and the hardest constraint, and then under-specified the one mechanism that actually determines whether NFR1 holds.**

The PRD says NFR1 "may rule out most off-the-shelf PDF libraries." That is true but incidental. What NFR1 actually rules out is **writing `w += a*b` in the layout engine** — and no library choice, no architecture diagram, and no amount of testing on amd64 will surface that. I found it in ten minutes because I looked for it. A team that does not know to look for it will ship a v0.1 that passes its own CI, and then receive a bug report from the first user with an Apple Silicon laptop and a Linux amd64 server, reporting that the hash Anan recorded in UJ-2 does not match. Diagnosing that from the outside would be brutal.

**NFR1 is achievable.** WASM is deterministic by specification and forbids fusion; IEEE 754 guarantees `+ - * / sqrt` are correctly rounded everywhere; `compress/flate` is empirically stable; and the remaining variability in a PDF is entirely under the writer's control. The conditions are: fixed-point or barrier-disciplined arithmetic, a restricted numeric surface, quantized emission, a pure render function, ordered emission, vendored encoders, and a declared version scope — plus an arm64 leg in CI. Every one of those is achievable in pure Go with no cgo.

Adopt Section 9's NFR1 and the project's signature requirement becomes real. Ship the current wording and it becomes a claim the product cannot keep.
