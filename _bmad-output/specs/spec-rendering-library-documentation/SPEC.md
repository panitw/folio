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

Go developers need one discoverable reference to install Folio, render a PDF from their own inputs, and understand every available API. Existing README guidance is incomplete and includes a stale `RenderTo` call shape; a source-checked guide closes that integration gap.

## Capabilities

- **CAP-1**
  - **intent:** Developers can install the rendering library in a Go module.
  - **success:** Fresh-module instructions identify prerequisites, correct dependency/import paths, a verified resolvable version and consumer toolchain considerations.
- **CAP-2**
  - **intent:** Developers can render their first PDF and choose returned bytes or writer output.
  - **success:** Complete examples supply a valid template, data, params and fonts, compile, produce a valid PDF, and handle errors and diagnostics.
- **CAP-3**
  - **intent:** Developers can supply suitable inputs and act on validation, rendering and output failures.
  - **success:** The guide explains data versus params, fonts, validation, diagnostics, error inspection and rendering limitations, with the behaviors specified in documentation-content.md.
- **CAP-4**
  - **intent:** Developers can look up every available exported library API.
  - **success:** Reference coverage reconciles with a fresh public export census, including methods, constants and fields; each entry explains its signature and relevant behavior, units, mutation and errors.
- **CAP-5**
  - **intent:** Developers can open and navigate the documentation from the Folio Designer top menu and repository entry points.
  - **success:** An accessible top-menu action opens the documentation without losing unsaved work; it resolves in development and the production build. Both READMEs link to it, anchors resolve, narrow-screen code is usable, and Markdown/HTML technical content agrees.

## Constraints

- Current Go declarations and behavior govern documentation; correct conflicting README snippets and preserve the existing tested example-block synchronization.
- Use the installation and API boundaries in documentation-content.md and the inventory in api-inventory.md; recheck both against the implementation revision.
- Runnable examples use public imports and handle every returned error; do not invent APIs or omit diagnostics silently.
- Preserve toolchain, font, locale, line-breaking and buffering caveats; make no unsupported concurrency or release guarantees.
- No implementation decision depends on the missing project-context.md; its absence is recorded in the memory log.

## Non-goals

- Rendering or public API changes, release creation, hosting, designer UI changes beyond documentation access, expanded CLI reference, or replacement of the template/expression language documentation.

## Success signal

A developer starting outside this repository follows the installation and first-PDF instructions successfully, observes matching output from the bytes and writer examples, and can find every exported symbol without reading implementation code. A designer user opens the same guide from the top menu without losing edits. Example execution, export coverage, menu/navigation checks and HTML inspection establish completion.

## Assumptions

- Publish the documentation as docs/rendering-library.md plus docs/rendering-library.html, following the existing expression-reference pair.
- “All APIs” includes the folio and fonts packages and the wasm integration package, with authoring/browser APIs clearly identified.
- Release resolution is an implementation-time verification step; this specification does not assert remote module availability.
- Open the top-menu documentation action in a separate tab to preserve the active editor session.
