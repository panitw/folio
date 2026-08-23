# Glossary — Folio

Terms the contract uses without defining inline. Several are Folio-specific and several are
industry terms Folio uses in a narrower sense than usual.

## The artifacts

**`.folio`** — the template file. A single canonical JSON document carrying a format version,
page setup, band content, and embedded image assets. The contract between the template author
and the integrating Go developer.

**`folio-go`** — the rendering library and the reference implementation of the engine. Compiles
to native targets and to `js/wasm`; the same engine in both.

**Golden report** — the Customer Account Statement acceptance fixture. See `acceptance.md`.

**Golden hash** — the recorded SHA-256 of a fixture's rendered PDF, stored per library version
*and* per Go toolchain version. A regression test is a comparison against it.

## The layout model

**Band** — one of exactly three regions: Page Header, Content, Page Footer. Structural, not
decorative: which band a component sits in determines whether it repeats on every page.

**Absolute placement** — within a band, components sit at fixed coordinates relative to that
band's top-left corner. They do not reflow and do not push each other down.

**Flow** — growth driven by data. In Folio, only a table flows. The Content band grows by
pagination — producing more pages — never by displacing siblings within a page.

**Clipping** — the defined response to content exceeding its declared bounds: cut at the
component boundary and report a diagnostic. Never reflowed, never silently dropped.

**Atomic row** — a table row is never split across a page boundary. If it does not fit in the
remaining space it moves whole to the next page.

**Orphaned footer** — a footer aggregate row separated from its data. Prevented: the footer
moves to the next page together with at least one preceding data row.

## The pipeline

**Two-pass** — lay out completely, then serialize. Required because `Page X of Y` needs the
total page count before page one can be emitted. Consequence: the writer-based API is an
ergonomic and API-stability choice, not a constant-memory guarantee.

**Page model** — the intermediate, renderer-agnostic structure produced by pass one. Names
geometry, glyph runs, images, and vector primitives; knows nothing about PDF. The seam that
makes future PNG, SVG, and HTML renderers possible.

**Late-bound slot** — text whose content depends on the finished document (total page count,
current page number), carried in the page model with its box already measured and substituted
between the passes. The mechanism that keeps pass two free of layout.

## Binding and expressions

**Row scope** — the current row inside a repeating region, addressed through an explicit alias
the region declares (`"as": "transaction"`, defaulting to `row`). A row never shadows the
document root.

**Parameter namespace** — `params.`, runtime values supplied at render time separately from
report data. The only route by which a date or time can reach output, since the engine reads no
clock.

**Conditional visibility** — hiding a component based on data. In scope. Distinct from
conditional *formatting* — data-driven styling — which is not.

## Determinism

**Byte-identical / byte-reproducible** — the same template, data, parameters, and library
version produce the same file hash on every machine and every compilation target. Folio's
signature claim and the constraint most other decisions bend around.

**FMA contraction** — a compiler fusing a multiply and an add into one instruction with a
single rounding. Go permits it and offers no flag to disable it; arm64 fuses, amd64 and wasm do
not. The divergence is invisible in rendered output and appears only in the hash. The reason
layout arithmetic is integer.

**Millipoint** — the engine's fixed-point unit: 1/1000 of a PDF point (a point being 1/72
inch), held as `int64`. Chosen so PDF number emission is exact at three decimal places.

**Subset tag** — the six-letter prefix on an embedded subsetted font's name (`ABCDEF+Noto`).
The PDF specification requires only that it be arbitrary and unique within a file, so Folio
derives it from a hash of the glyph set rather than at random.

**`SOURCE_DATE_EPOCH`** — the reproducible-builds convention for pinning document dates. Read
by the CLI and passed in as a parameter, never read by the library, which touches no
environment variable.

## Text

**Shaping** — turning characters into positioned glyphs using the font's OpenType `GSUB` and
`GPOS` tables. Thai vowel and tone-mark placement *is* GPOS mark positioning; a metrics-only
font layer cannot render Thai correctly.

**Dictionary-based line breaking** — the only way to break Thai, which does not delimit words
with spaces. Unicode UAX #14 assigns Thai to a complex-context class and defers to a
dictionary, so general segmenters yield effectively no break opportunities within Thai text.

**FontSet** — the explicit collection of font data handed to a render, with each family's
ordered fallback chain. Part of a render's identity: the same template with a different
FontSet is a different render, not a silent substitution. Never a host font query.

## The designer

**draw.io model** — open it in a browser, work on files on your own machine, no account, no
server-side storage, nothing uploaded.

**Design canvas** — the fast, interactive, explicitly *approximate* editing view.

**Preview** — the exact production document, rendered by the real engine compiled to
WebAssembly in the tab. Not an approximation and not a remote service.

**Stale preview** — a preview whose inputs have changed since it was rendered. Must be visibly
invalidated or re-rendered; presenting one as current is the single state failure that breaks
the product's central promise.
