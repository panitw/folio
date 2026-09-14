# Documentation page content and acceptance

## Deliverables and navigation

Create `docs/rendering-library.md` and `docs/rendering-library.html`, and publish `docs/folio-format.md` and `docs/folio-format.html` (see *Format reference*). Link them from `README.md` and `folio-go/README.md`; fix conflicting existing usage snippets, including the old single-return `RenderTo` example. Preserve `TestREADMEExampleBlockMatchesSource` when touching the library README. Follow the existing expression-reference HTML visual conventions, with semantic headings, keyboard-accessible anchor navigation, readable code blocks and narrow-screen overflow. Add a clearly labeled Rendering library documentation action to Folio Designer’s top menu/document bar (`folio-designer/src/App.tsx`). It must be usable with mouse and keyboard and remain available independently of whether a template is loaded or rendering succeeds. Assume a separate browser tab, preserving unsaved template changes, selection and undo history.

Serve the HTML documentation as an asset of the designer in both development and production, respecting its configured base path; a repository-relative link outside the built application is insufficient. Include it in the applicable offline asset flow. Use local or system fonts instead of copying the expression page’s remote font imports. No separate hosting deployment is required.

## Installation

Explain module installation from an application directory:

```sh
go mod init example.com/folio-demo
go get github.com/panitw/folio/folio-go@<verified-version-or-commit>
go mod tidy
```

The final runnable guide must substitute a verified resolvable version or commit, state what was tested, and explain upgrades/pinning. Do not leave the placeholder in the primary copyable workflow. Use `@latest` only after verifying what it resolves to; do not invent `v0.1.0`. When remote installation is unavailable, state that limitation and provide a separately labeled local-checkout `replace` workflow rather than claiming Go module installation passed.

Import the module root as `folio` and shipped fonts from `github.com/panitw/folio/folio-go/fonts`. The checked-in module declares Go 1.25.0 minimum and toolchain go1.26.0. Explain that dependency toolchain directives do not pin the consuming application's build: reproducible bytes require the consumer to control its own toolchain and inputs. Consult `folio-go/go.mod` again when implementing.

## First PDF and common usage

Provide a complete runnable Go program and matching `.folio` JSON, adapted from `folio-go/testdata/example/first-pdf.folio`; avoid references to unavailable repository test paths. Include imports, template loading, `Data`, `Params`, `fonts.Shipped()`, `Render`, diagnostic handling, and checked file writing. Also show `ParseTemplate` for in-memory template bytes.

Show `RenderTo` using `diagnostics, err := folio.RenderTo(writer, tpl, data, params, fontSet)`. Explain full-document buffering, matching PDF bytes and writer failures/short writes; a failed writer may already have accepted bytes. If an HTTP example is included, distinguish pre-response render errors from output failures after headers/body have been committed.

Explain `SerializeTemplate` and that `Template` is opaque; callers obtain it through parsing/loading. Explain `Validate(templateBytes, data, params, fonts)` with the same real inputs intended for rendering, rather than promising structural-only validation or PDF output.

## Inputs, warnings and failures

- `Data` and `Params` are distinct defined types over raw JSON bytes. Use valid JSON data; nil/empty params means no runtime values. Explain reserved `params.*` bindings and absent-path failures.
- Preserve author-written numeric precision; avoid examples that first round data through `float64` decoding.
- Cover explicit `FontSet`, opt-in shipped font package, shipped families, fallback-chain names, custom font bytes and template-embedded font precedence/validation as established by current code and tests. Explain applicable font licensing and supported face restrictions without inventing policy.
- Document `Result.Bytes`, `Result.Diagnostics`, every `Diagnostic` field, severity values and `Severity.String()`. Warnings can accompany successful output; errors abort rendering. Use stable codes, not message parsing, and show `errors.As` with `*folio.RenderError` while retaining a fallback for ordinary errors.
- Document all exported diagnostic constants with their actual string values and meanings; validate against `diagnostic.go` and its registry tests.
- Explain document-controlled locale and dates, no ambient clock/font discovery in rendering, fixed-input/toolchain byte reproducibility, and the CLI-only nature of `SOURCE_DATE_EPOCH` if mentioned.
- Preserve shipped Japanese glyph-form, Thai declared unbreakable-value, Latin wrapping and CJK kinsoku limitations where still applicable. Link `docs/expression-reference.md` for expression syntax instead of duplicating it.
- Separate load-time refusal from render-time outcome. `ParseTemplate`/`LoadTemplate` refuse a table `minHeight` taller than the content window and an invalid section break or page structure; `PAGES_INVALID` and `SECTION_BREAK_INVALID` locate the band or page through `Diagnostic.DataPath`, not `ElementID`.

## Template features that change rendered output

Summarise what each feature does to the PDF, the format version a template must declare to use it, and its diagnostics. Name the keys a developer needs and link the matching section of `docs/folio-format.md` (*Designed pages*, *Pagination*, `table`, *Elements*, *Expressions*) for full field rules. Each feature gets a minimal template verified by rendering.

## Format reference

Publish `_bmad-output/specs/spec-folio/folio-format.md` as `docs/folio-format.md` and an equivalent `docs/folio-format.html` using the same visual, navigation and offline-font conventions as the guide, served with the designer's documentation assets.

- Keep every normative rule: units, document keys, version derivation, `page`, `fonts`, `bands`, designed pages, pagination, elements, `table`, expressions, `style`, `assets`, line breaking, what is not in the file, and the worked example.
- Reconcile against the implementation revision: version ladder and `SupportedVersion` (`internal/template/version.go`), closed sets, load errors and their codes, and save canonicalization. Correct stale statements rather than preserving them.
- Remove planning traceability labels (FR, AD, D-, Story, QA references) and internal source-file pointers; restate any rule they carried in plain terms.
- Give sections stable anchors the guide can link. Link it from the guide and both READMEs; defer expression syntax to `docs/expression-reference.md`.
- Do not edit the planning copy. Once published, `docs/folio-format.md` is canonical and the planning copy is a historical snapshot; later format changes update `docs/`.

| Feature | Rendering behavior to document | Min. version | Diagnostics |
|---|---|---|---|
| Formulas | Comparisons, exact-decimal arithmetic, nested ternaries and `true`/`false`/`null` in Visibility, `if()` and text; a bare number in text prints its exact decimal. Link the expression reference for syntax, limits and division scale. | 2.0; 3.3 when a text expression can return a number | `EXPRESSION_INVALID`, `BINDING_PATH_ABSENT`, `AGGREGATE_EMPTY_AVERAGE` |
| Proportional table widths | Column widths allocated from proportions against the table width. | 3.0 | allocation errors located by `DataPath` `table.width` or `column.<field>` (verify code) |
| Table frame, rules, `minHeight` | A table's `style.border`/`background` frame each page slice once (breaking change from per-cell chrome); author-controlled rules between columns; a ruled area taller than its rows. Multi-line column labels and cell padding. | 3.1 | `TABLE_MIN_HEIGHT_UNPLACEABLE` |
| Column header alignment | Per-column `headerAlign` overriding the header style. | 3.2 | — |
| Barcode and QR code | `barcode` (Code 128) and `qrcode` (L/M/Q/H, default M) elements bind content like text, draw vector modules of whole-millipoint width inside the box with quiet zone, never distorted. Unencodable/too-long static content fails load; at render the code is omitted with a warning; undersized modules draw with a warning. Control characters are real characters in the file. | 4.0 | `BARCODE_UNENCODABLE`, `BARCODE_DOES_NOT_FIT`, `BARCODE_MODULE_TOO_SMALL`, `QRCODE_TOO_LONG`, `QRCODE_DOES_NOT_FIT`, `QRCODE_MODULE_TOO_SMALL` |
| Section break | Content-band `sectionBreak` offset: when content above crosses it, the below-line section moves as a rigid block. Anchored (default) moves whole pages at declared positions; `sectionBreakAnchor: false` follows where the content ends if it fits. Never drawn in the PDF; no effect when uncrossed. Wins over `keepTogether`, with a warning. | 4.1 | `SECTION_BREAK_INVALID`, `SECTION_BREAK_STRADDLED`, `SECTION_BREAK_SPLITS_KEEP_TOGETHER` |
| Designed pages | Top-level `pages` array renders designed pages in order with one shared header/footer and page count across all output pages. `pageBreak` on starts a new output page after the previous page and its overflow; off continues where it ended when the block fits. One section break per page. | 4.1 | `PAGES_INVALID` |

Verify version numbers against `internal/template/version.go` and codes against `diagnostic.go` at implementation time; saving never lowers a loaded version. Adapt minimal templates from the golden fixtures under `fixtures/` and the `*_template.go` examples rather than inventing layouts.

## Complete API reference

Use api-inventory.md as the starting census, then regenerate against the implementation revision. Organize entries into rendering/validation, fonts, diagnostics, template/asset helpers, authoring/canvas (including group movement and the command-kind catalog), and wasm integration. Cover all exported functions, types, constants, variables, methods and struct fields in the three listed packages. Do not mistake the core Render/RenderTo pair for the entire API.

Each callable entry needs an exact signature, purpose, argument and return meanings, preconditions, errors, and mutation/ownership behavior where applicable. Record field types, units, JSON keys and absence/default semantics where those affect callers. Explain command payloads accepted by the public byte-based authoring methods, with supported operations and source-checked examples; untyped JSON inputs still require documentation. Clearly identify browser/editor helpers and their stability constraints.

`SnapToGrid` and the `FontSize`/`BorderWidth` fields of `CanvasAuthoredProperties` currently mention `geom.Length` from an internal package. Describe that external callers cannot import/name that internal type; show only externally compilable usage, and do not quietly change the API as part of this documentation work. State only concurrency or ownership guarantees supported by current implementation/tests.

## Verification

1. Verify dependency resolution from an external consumer module and record the selected version/commit. Local replacement can validate working-tree examples but cannot establish remote installation availability.
2. Compile and run the primary example and the writer variant against that revision; inspect valid PDF output and compare bytes for identical inputs. Check validation and diagnostic/error examples against actual outcomes.
3. Render each feature's minimal template: formula visibility on both sides of a threshold; barcode and QR decoding to the bound string; a section break uncrossed (hash equal to no break), crossed anchored and crossed unanchored; two designed pages with Page Break on and off; a ruled table with `minHeight`. Confirm documented page placement and diagnostic codes, including the load-time refusals.
4. Reconcile exported declarations, methods and fields from all three packages with reference entries, including grouped constants. `go doc -all .`, `go doc -all ./fonts` and `go doc -all ./wasm`, run inside folio-go, supply a readable census; use AST/type inspection if needed for exact coverage.
5. Exercise the top-menu action with and without a loaded template and with unsaved edits; verify keyboard access, the opened page and preserved editor state. Verify the documentation URL in the built application with its configured base path and after offline caching. Verify README links and page anchors, including guide links into the published format reference in the repository and the built designer. Check the format reference's version ladder and load-error codes against source and confirm no planning traceability labels remain. Inspect the rendered HTML at desktop and narrow widths, including keyboard navigation and long signatures. Check Markdown/HTML content equivalence.
6. Run affected documentation/example checks, including the README example synchronization test if touched. No renderer behavior changes are required.

## Source anchors

Implementation evidence: `folio-go/go.mod`, `folio-go/render_entry.go`, `folio-go/validate.go`, `folio-go/folio.go`, `folio-go/diagnostic.go`, `folio-go/render_error.go`, `folio-go/fontset.go`, `folio-go/fonts/fonts.go`, `folio-go/internal/template/version.go`, `folio-go/barcode_render.go`, `folio-go/section_break.go`, `folio-go/content_pages.go`, `folio-go/table_frame.go`, the `folio-go/*_template.go` feature examples, `fixtures/`, the exported source files recorded in the API inventory and their adjacent tests. Existing presentation/content: `_bmad-output/specs/spec-folio/folio-format.md` (format reference source), root and library READMEs, `RELEASING.md`, `docs/expression-reference.md`, `docs/expression-reference.html`. These are implementation-time evidence, not additional adopted contract companions.
