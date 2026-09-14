---
id: SPEC-rendering-library-documentation
companions:
  - documentation-content.md
  - api-inventory.md
sources: []
---

> **Canonical contract.** This SPEC and its companions define what to build, test, and validate.

# Rendering library documentation

## Why

Go developers need one discoverable reference to install Folio, render a PDF from their own inputs, and understand every available API and template feature that shapes output. Existing README guidance is incomplete and includes a stale `RenderTo` call shape, and nothing documents the formulas, barcode/QR codes, section breaks, designed pages and ruled tables that have since shipped; a source-checked guide closes that integration gap.

## Capabilities

- **CAP-1**
  - **intent:** Developers can install the rendering library in a Go module.
  - **success:** Fresh-module instructions identify prerequisites, correct dependency/import paths, a verified resolvable version and consumer toolchain considerations.
- **CAP-2**
  - **intent:** Developers can render their first PDF and choose returned bytes or writer output.
  - **success:** Complete examples supply a valid template, data, params and fonts, compile, produce a valid PDF, and handle errors and diagnostics.
- **CAP-3**
  - **intent:** Developers can supply suitable inputs and act on validation, rendering and output failures.
  - **success:** The guide explains data versus params, fonts, validation, load-time refusal versus render warnings, diagnostics, error inspection and rendering limitations, with the behaviors specified in documentation-content.md.
- **CAP-4**
  - **intent:** Developers can look up every available exported library API.
  - **success:** Reference coverage reconciles with a fresh public export census, including methods, constants, fields and authoring command kinds; each entry explains its signature and relevant behavior, units, mutation and errors.
- **CAP-5**
  - **intent:** Developers can open and navigate the documentation from the Folio Designer top menu and repository entry points.
  - **success:** An accessible top-menu action opens the documentation without losing unsaved work; it resolves in development and the production build. Both READMEs link to it, anchors resolve, links to the published format reference resolve in the repository and the designer build, narrow-screen code is usable, and Markdown/HTML technical content agrees.
- **CAP-6**
  - **intent:** Developers can predict how template features affect rendered output and failures.
  - **success:** Each feature in documentation-content.md (formulas, barcode and QR code, anchored and unanchored section break, designed pages with Page Break on/off, table frame/rules/`minHeight`) has a minimal template that renders with the documented placement, its required format version, and its load errors and render warnings by stable code.

## Constraints

- Current Go declarations and behavior govern documentation; correct conflicting README snippets (including `folio-go/README.md`'s single-return `RenderTo`) and preserve the existing tested example-block synchronization.
- Use the installation, API and feature boundaries in documentation-content.md and the inventory in api-inventory.md (census at `15871c7`); recheck both against the implementation revision.
- Runnable examples use public imports and handle every returned error; do not invent APIs or omit diagnostics silently. Examples using newer features declare the format version they require.
- Link `docs/expression-reference.md` for formula syntax and the published `docs/folio-format.md` for field rules rather than duplicating them; claims about barcode/QR, section break and page placement are limited to what fixtures and tests verify.
- Publish the format reference from `spec-folio/folio-format.md`: keep every normative rule, reconcile it with the implementation's supported version and load/save behavior, and strip planning traceability labels and internal source pointers. Do not edit the planning copy.
- `docs/` is the source of truth for user-facing documentation: published pages reflect the implementation at the delivery revision, later documentation changes land in `docs/`, and the planning format reference becomes a historical snapshot once published.
- Preserve toolchain, font, locale, line-breaking and buffering caveats; make no unsupported concurrency or release guarantees.
- No implementation decision depends on the missing project-context.md; its absence is recorded in the memory log.

## Non-goals

- Rendering or public API changes, release creation, hosting, designer UI changes beyond documentation access, expanded CLI reference, new `.folio` format rules or a rewrite of the format reference beyond publishing it, or replacement of the template/expression language documentation.

## Success signal

A developer starting outside this repository follows the installation and first-PDF instructions successfully, observes matching output from the bytes and writer examples, renders a barcode, a section break and a two-page document from the guide's templates with the described result, and finds every exported symbol without reading implementation code. A designer user opens the same guide from the top menu without losing edits. Example execution, feature-template rendering, export coverage, menu/navigation checks and HTML inspection establish completion.

## Assumptions

- Publish the documentation as docs/rendering-library.md plus docs/rendering-library.html, following the existing expression-reference pair.
- “All APIs” includes the folio and fonts packages and the wasm integration package, with authoring/browser APIs clearly identified.
- Release resolution is an implementation-time verification step; this specification does not assert remote module availability.
- Open the top-menu documentation action in a separate tab to preserve the active editor session.
- Publish the format reference as docs/folio-format.md plus docs/folio-format.html, shipped with the designer's documentation assets so guide links resolve in production and offline.