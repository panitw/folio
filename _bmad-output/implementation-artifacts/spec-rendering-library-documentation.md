---
title: 'Rendering library guide, published format reference and designer access'
type: 'feature'
created: '2026-09-15'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '563352e6f92adf8491c8c6c8c6df70299597396c'
context:
  - '{project-root}/_bmad-output/specs/spec-rendering-library-documentation/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-rendering-library-documentation/documentation-content.md'
  - '{project-root}/_bmad-output/specs/spec-rendering-library-documentation/api-inventory.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Go developers have no reference for installing Folio, rendering a PDF, the full exported API, or the formulas, barcode/QR, section breaks, designed pages and ruled tables that shape output; the `.folio` format reference exists only as a planning artifact, and Folio Designer offers no way to reach documentation.

**Approach:** Deliver the contract in `_bmad-output/specs/spec-rendering-library-documentation/` (CAP-1–CAP-6): publish `docs/folio-format.*`, write `docs/rendering-library.*` with verified example templates, link both READMEs, and add a document-bar link that opens the bundled guide in a new tab, offline included. `docs/` becomes the documentation source of truth.

## Boundaries & Constraints

**Always:**
- Current Go declarations and behavior at the delivery revision govern every statement; examples use public imports and handle every error and diagnostic.
- Guide content, ordering and caveats follow `documentation-content.md`; API coverage reconciles with a fresh `go doc -all` census of `folio`, `fonts`, `wasm` (baseline `api-inventory.md`).
- Format reference keeps every normative rule of `spec-folio/folio-format.md`, is corrected to `internal/template/version.go` (4.1) and `closedsets.go`, and drops planning labels (FR, AD-, D-, Story, AC, S9, SPEC-*), change-history blockquotes and internal file/planning-doc pointers.
- HTML pages are hand-written twins of their Markdown, reuse the expression-reference layout conventions, use system font stacks (no remote fonts), and give sections stable anchors.
- Designer link: always enabled, `aria-label` plus `data-tip`, no `title`, opens in a new tab so editor state survives; the "Local file actions" group stays six buttons.
- **Decision (bundling):** the designer bundles all three pages — guide, format reference and expression reference — as fingerprinted, precached `/assets/` entries with cross-page links rewritten to fingerprinted names. `docs/expression-reference.html` replaces its Google Fonts links with system font stacks; its content is otherwise unchanged.
- **Decision (asset cap, user 2026-09-15):** the committed release already holds 62 cached assets, so the three pages make 65; raise `maximumCacheAssets` in `folio-designer/src/release-payload.ts` from 64 to 65. No other cap or threshold changes.

**Never:** change rendering, public API or format rules; edit `spec-folio/folio-format.md` or the Go comments that cite it; push, tag or release; add a `go` fence before `folio-go/README.md`'s first fence; document CLI flags beyond a pointer.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| First PDF | fresh module, guide program + `docs/examples/first-pdf.folio` | valid nonempty PDF; `Render` and `RenderTo` bytes equal | every error checked |
| Writer failure | `RenderTo` into failing writer | documented: whole PDF buffered, partial write possible | returned error shown |
| Load refusal | ruled-table example with `minHeight` above content window | `ParseTemplate` fails `TABLE_MIN_HEIGHT_UNPLACEABLE` | `errors.As` `*folio.RenderError` |
| Feature render | each `docs/examples/*.folio` with its data | documented page count/placement and diagnostic codes | warnings listed, not fatal |
| Designer link | unsaved edits, click or Enter on link | guide opens in new tab; document, selection, undo unchanged | n/a |
| Offline | release installed, network off | bundled guide, format reference and expression reference load from cache, links between them work | n/a |

**Decision (install):** the primary install command is `go get github.com/panitw/folio/folio-go@main` (no tags exist; on 2026-09-15 it resolved to `v0.0.0-20260914045332-3121963bc2da`, which predates designed pages and unanchored section breaks). The user pushes `main` before install verification; implementation then runs `@main` from a fresh module, records the resolved pseudo-version in the guide, and explains that the reader's `go.mod` pins it. If the resolved commit still predates the guide's features at verification time, HALT and ask rather than documenting around it. This build never pushes.

</frozen-after-approval>

## Code Map

- `_bmad-output/specs/spec-folio/folio-format.md` -- publication source; history blockquotes at 14–15, 68–138, 652–662, 907–926; stale version prose at 75, 88, 919, 924; "nine closed sets" at 69–71/102 (add qrcode `errorCorrection`, `rules.between`, table align).
- `docs/expression-reference.html` -- layout to copy: no `<html>` wrapper, inline `<style>`, `.wrap` → header → nav anchors → sections, `.sample pre`; lines 2–4 are remote font links to replace with system font stacks.
- `folio-go/render_entry.go:246`, `validate.go`, `render_error.go`, `diagnostic.go`, `internal/diag/diag.go` (`allCodes`, `dispositions`) -- signatures, error wrapping, code strings/severity; table-width errors are `TEMPLATE_FIELD_INVALID` with `DataPath` `table.width`/`column.<field>`.
- `folio-go/example_test.go`, `readme_example_test.go:23` -- README first-fence sync; `folio-go/README.md:239-245` stale `RenderTo` snippet.
- `fixtures/{barcode-thai-bill-payment,qrcode-payments,section-break-statement,section-break-unanchored,multi-page-flow}/` and `folio-go/*_template.go` -- sources to adapt minimal examples; formulas and table rules have no fixture (see `table_rules_test.go`, `boolean_formulas_test.go`).
- `folio-go/section_break_test.go`, `pages_section_break_test.go`, `multi_page_fixture_test.go` (`mpExtractRuns`), `barcode_element_test.go` (`barcodePages`) -- page/placement probes to reuse in docs example tests.
- `tools/codescan/scan.py` -- barcode/QR decode check via `uv run --with zxing-cpp --with pillow`.
- `folio-designer/src/App.tsx:3338-3405` -- inline document bar; insert link as its own `role="group"` after `.mode-switch`; `toolTip` at line 44.
- `folio-designer/src/toolbar-icons.tsx:7` -- `ToolGlyph` union; add a `docs` glyph.
- `folio-designer/scripts/build-wasm.mjs:56-110` -- `fingerprint` copy into ignored `src/generated/runtime`; model for copying docs HTML; `src/engine-client.ts:209` shows `new URL(..., import.meta.url)` emission.
- `folio-designer/scripts/generate-offline-release.mjs:13-18`, `verify-offline-release.mjs`, `src/release-payload.ts:42` (cap 64, now 61; `warnCacheAssets` 56 already exceeded) -- precache path.
- `folio-designer/scripts/offline-service-worker-template.mjs:3-6,78-86` -- `isCacheableStaticRequest` rejects every navigation; only `/` and `/index.html` navigations are served from cache.
- `folio-designer/scripts/forbidden-font-hosts.mjs:118-135` -- scans generated HTML; its `docs/expression-reference.html` Google Fonts exception comment becomes stale and must be updated.
- `folio-designer/src/App.test.tsx:1183`, `src/control-vocabulary-contract.test.tsx:997`, `e2e/document-bar-fit.spec.ts:61-72` -- bar contracts to extend, not loosen.

## Tasks & Acceptance

**Execution:**
- [x] `docs/folio-format.md` -- publish reconciled, label-free format reference -- linkable source of truth.
- [x] `docs/examples/*.folio` + data JSON -- first PDF, formula visibility, barcode+QR, section break anchored and unanchored, two designed pages (Page Break on/off), ruled table with `minHeight` -- verified inputs the guide embeds verbatim.
- [x] `folio-go/docs_examples_test.go` -- render each example asserting page count, placement and codes; assert guide embeds each file verbatim; assert every exported identifier of `folio`, `fonts`, `wasm` (via `go/doc`) appears in `docs/rendering-library.md` -- keeps docs honest as code changes.
- [x] `docs/rendering-library.md` -- install, first PDF, inputs/diagnostics, load vs render, template features (links into `folio-format.md` anchors), full API reference and command catalog -- CAP-1–4, CAP-6.
- [x] `docs/rendering-library.html`, `docs/folio-format.html` -- equivalent HTML twins.
- [x] `docs/expression-reference.html`, `folio-designer/scripts/forbidden-font-hosts.mjs` -- system font stacks instead of Google Fonts; update the stale exception comment -- bundling decision.
- [x] `README.md`, `folio-go/README.md` -- link the docs; fix `RenderTo` snippet -- CAP-5, stale example.
- [x] `folio-designer/scripts/build-wasm.mjs` -- fingerprint-copy the three HTML pages, rewrite cross-page links, expose the guide URL to the app -- production/offline assets.
- [x] `folio-designer/scripts/offline-service-worker-template.mjs` + its tests -- serve same-origin navigations to precached `/assets/*.html` entries from cache -- CAP-5 offline.
- [x] `folio-designer/src/App.tsx`, `src/toolbar-icons.tsx`, `src/App.css` -- documentation link group -- CAP-5.
- [x] `folio-designer/src/App.test.tsx`, `src/control-vocabulary-contract.test.tsx`, `e2e/document-bar-fit.spec.ts`, new `e2e/documentation-link.spec.ts` -- link contract, fit at 1024px in both modes, new-tab open with unsaved edits preserved.

**Acceptance Criteria:**
- Given the census test, when a new exported identifier is added without documentation, then `go test ./...` fails naming it.
- Given `docs/folio-format.md`, when grepped for `FR\d|AD-\d|D-\d|Story \d|SPEC-|\.go\b`, then nothing matches, and its version ladder names 4.1 as supported.
- Given `npm run build`, when the offline release is verified, then it passes with all three pages precached and no more than 65 assets.
- Given the designer at 1024px in Design and Preview, when the bar is measured, then it still fits and the link is reachable by Tab.

## Implementation Notes

- Planning measured the release at 61 cached assets; the designer build at baseline actually emits 62, so bundling three pages gave 65 and `npm run build` failed at the 64 cap. User chose to raise the cap to 65 (recorded in the frozen block); implementation was paused for that decision and resumed after it.
- User pushed `main` on request (3121963..563352e); `go get …@main` resolved to `v0.0.0-20260914182357-563352e6f92a`, recorded in the guide. A fresh module ran both guide programs against it: equal `Render`/`RenderTo` bytes and the documented failing-writer error.
- Implementation delegated designer work and the guide HTML to helper agents; `docs/folio-format.html` had a stray duplicated link list after `</nav>`, removed.
- The source accepts three component kinds the planning inventory omitted (`bindComponentScalar`, `bindTableCollection`, `configureTableBinding`); the guide documents all 44 plus `pageSetup`.
- Raising the cap also moved `font-store.test.ts`'s pinned `maximumCacheAssets` line and `release-payload.test.ts`'s over-bound fixture (65 → 66), and the present-tense cap in `src/vendor/pdfjs/PROVENANCE.md`. `control-vocabulary-contract.test.tsx` floors moved 33 → 39 and pinned 32 → 38 because the new group renders in every state; `vite.config.ts` passes the docs' fingerprinted names through unhashed so cross-page links resolve.
- Verified by orchestrator: the 11 covering Go tests (docs examples, verbatim embedding, export census, guide programs, README sync, diagnostic census) pass; full `go test ./...` passes except `internal/text` `TestCorpusMeetsP6ExerciseFloors`, red by design (`corpus_test.go:216`) and untouched; service-worker tests 6/6; `npm run build` passes offline verify at 65/65 assets (margin 0 warning); both Playwright specs 6/6 (bar spare room 270px Design, 198px Preview). Banned-label grep on `docs/folio-format.md` is empty.
- Offline matrix row is covered by the service-worker fetch-handler test (cached `/assets/*.html` navigation served), the release verifier's cross-link check and the e2e emitted-link test; no browser test opens the pages with the network disabled.
- User direction during review (2026-09-15): the three HTML twins take the claude.ai/design "Documentation site redesign" shell — FOLIO8 brand, sticky filterable sidebar listing every section of all three pages (cross-page entries as `stem.html#id`, rewritten to fingerprinted names by the bundler), document pills, the design's typography and callouts, a small inline script for filter and active section, and build-time syntax highlighting of Go, JSON and shell samples (`tk-*` spans; text unchanged once tags are stripped; expression-reference samples keep their hand colouring). Three separate pages kept (bundling, verifier, cache 65 unchanged); page content unchanged.

## Spec Change Log

## Review Triage Log

| # | Source | Finding | Verdict | Route | Evidence |
|---|---|---|---|---|---|
| 1 | blind, edge | README `statementHandler` returns empty 200 on render error | medium | patch | Baseline had `http.Error`; diff replaced it with log-and-return, contradicting the guide's io.Writer section. |
| 2 | blind, gap-other | Census and verbatim tests read only the `.md` twin; bundled `.html` can drift | medium | patch | Both tests open only `docs/rendering-library.md`; designer ships the `.html`. |
| 3 | blind | `#formulas-and-visibility` vs `#formulas` anchor drift between twins | false | reject | Each link resolves in its own format: `expression-reference.md` heading "Formulas and visibility", `expression-reference.html` `id="formulas"`. |
| 4 | blind | HTML CLI pointer has no README link | low | reject | Bundled offline page cannot reach `../README.md`; plain text is deliberate. |
| 5 | blind, edge | Census matches common-word identifiers anywhere in prose | low | reject | Real but qualified matching adds complexity; rare in everyday use. |
| 6 | edge | Census misses embedded fields and interface methods | low | reject | No exported interfaces in public packages today; fix adds branches. |
| 7 | blind | "tested on every change" overclaims unasserted outcomes | medium | patch | Decoded barcode contents, byte counts, some load failures and exact y are not asserted. |
| 8 | blind, gap | Bundled docs pages unguarded against remote font hosts | medium | patch | Pre-verified: host scan runs before copy and skips `docs/`; verifier has no host check. |
| 9 | blind, edge | Verifier link check would fail absolute external links | low | reject | No such link exists; guarding adds a branch for an unshown case. |
| 10 | blind, edge | Consumer-module test fails on cold module cache | false | reject | It runs inside `go test` of folio-go, which has already fetched every dependency; `go.sum` is copied. |
| 11 | blind | `docs/examples/*/main.go` outside go vet | low | reject | Compiled and run by the consumer-module test. |
| 12 | blind | Concurrency guidance contradictory | false | reject | Lines 456 and 896 agree: no guarantee; do not share without locking. |
| 13 | blind | Page Break off rule unclear or wrong | false | reject | Wording matches `folio-go/content_pages.go:155-159` (no overflow, new output page). |
| 14 | blind | Cache at 65/65 margin 0; Code Map stale | false | reject | Cap 65 is a frozen user decision; Code Map fix edits this spec. |
| 15 | blind | No network-off browser test | low | reject | Covered by SW unit test, verifier and e2e link test (noted); offline e2e is substantial new harness. |
| 16 | blind | Designer labels "invalid Anchor", "Page Break on/off" in Go guide | low | patch | Direct wording fix to field names. |
| 17 | blind | build-wasm comment claims base-path support | low | patch | No `base` in vite config; SW, verifier, e2e hard-code `/assets/`. Delete phrase. |
| 18 | edge | `.DS_Store` in docs/examples fails verbatim test | low | patch | WalkDir takes every file; macOS Finder creates it; one-condition fix. |
| 19 | edge | CRLF checkout breaks verbatim test | low | reject | Unlikely in this repo's workflow; fix adds normalisation. |
| 20 | edge | e2e `toHaveTitle('')` passes for untitled page | low | reject | All pages have titles; guard for unshown state. |
| 21 | gap | Group-digest renaming rule untested | low | defer | Pre-verified; test needs helper extraction. Logged in deferred-work.md. |
| 22 | gap-other | Barcode test message prints `len(pages)` as rects | low | patch | Direct correction at `docs_examples_test.go:148`. |

## Verification

**Commands:**
- `cd folio-go && go test ./...` -- expected: pass, including docs example, census and README sync tests.
- Scratch consumer module: `go mod init`, `go get github.com/panitw/folio/folio-go@main` (plus local `replace` run for working-tree examples), run guide's first-PDF and writer programs -- expected: valid PDF, equal bytes; record resolved pseudo-version.
- `uv run --with zxing-cpp --with pillow python3 tools/codescan/scan.py <barcode example pdf>` -- expected: decodes to bound strings.
- `cd folio-designer && npm run lint && npm run typecheck && npm test && npm run build` -- expected: pass, offline verify green.
- `cd folio-designer && npx playwright test e2e/document-bar-fit.spec.ts e2e/documentation-link.spec.ts` -- expected: pass.

**Manual checks (if no CLI):**
- All three HTML pages at desktop and 400px: nav anchors by keyboard, long signatures scroll inside code blocks, Markdown/HTML content equivalent.
