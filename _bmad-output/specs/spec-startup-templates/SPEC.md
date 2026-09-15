---
id: SPEC-startup-templates
companions:
  - ./example-templates.md
  - ../../planning-artifacts/ux-designs/ux-startup-templates-2026-09-15/mockups/Main.dc.html
  - ../../planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md
  - ../../planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md
  - ../../../docs/folio-format.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# folio8 Designer — Start from Blank or a Working Example

## Why

**A pain to solve, for the new author.** The Designer opens straight onto an empty A4 page. A
first-time author has no template, no sample JSON, and so no way to see what bindings, tables,
pagination or a real Preview look like without building all of it first; Preview with no data
shows only stand-in values. A startup dialog offering Blank or one of four realistic examples —
each carrying its own made-up sample data — puts a finished, previewable document in front of
the author in one click, and doubles as a reference for how such documents are built.

## Capabilities

- **CAP-1 — Startup dialog**
  - **intent:** When the Designer launches, the author chooses Blank or one of the bundled
    examples from a dialog showing each option's name and a first-page thumbnail.
  - **success:** Every fresh app open shows the dialog after the engine is ready; choosing
    Blank produces the same document today's starter produces (byte-identical serialization);
    the dialog works with the browser offline.

- **CAP-2 — Example opens previewable**
  - **intent:** Choosing an example loads its template and its sample data together and shows
    it in Preview, rendered from that sample.
  - **success:** After choosing any example the Designer is in Preview, the render uses the
    example's sample JSON (the "No sample data" notice is absent), and the sample tree is
    available for binding exactly as if the author had opened that JSON file.

- **CAP-3 — Reopen from the document bar**
  - **intent:** The author reaches the same dialog later from the document bar's New action,
    and can back out of it there.
  - **success:** The toolbar's Start blank becomes New…, which opens the dialog; Cancel or
    Escape leaves the current document untouched. At launch, dismissing the dialog is Blank.

- **CAP-4 — Four bundled examples**
  - **intent:** The Designer ships Invoice, Bank Statement, Legal Contract and Electricity Bill
    examples with fictional English data, each exercising engine features an author of that
    document type needs.
  - **success:** Each example matches its entry in `example-templates.md`, loads and renders
    against its sample JSON with zero diagnostics, and the build fails if any does not.

- **CAP-5 — An example is an ordinary document**
  - **intent:** An opened example behaves like any unsaved template: editable, undoable, and
    savable to a location the author chooses, without altering the bundled original.
  - **success:** Save on a freshly opened example prompts for a location and writes a `.folio`
    that reopens identically; choosing the same example again from New… yields the original.

- **CAP-6 — Unsaved changes are never lost silently**
  - **intent:** Before the dialog replaces a document that has unsaved changes — with Blank, an
    example, or an opened file — the author is warned and must confirm.
  - **success:** With unsaved edits, confirming any dialog choice shows a warning naming the
    document; Keep editing returns to the dialog with the document and its undo history intact;
    only Discard replaces it. A document with no unsaved changes is replaced without asking.

- **CAP-7 — Save the sample data**
  - **intent:** The author saves the currently loaded sample JSON to a local file, so an
    example's sample data survives a reload and can be opened again as sample data.
  - **success:** After opening an example and saving its sample, reopening the saved `.folio`
    and the saved JSON renders a Preview identical to the one the example first showed; the
    saved file's bytes equal the loaded sample's bytes.

- **CAP-8 — Open an existing file from the dialog**
  - **intent:** The dialog offers Open existing file, so a returning author skips the choice
    of Blank or example.
  - **success:** Open existing file in the dialog uses the same local open path as the
    document bar's Open; cancelling the file picker returns to the dialog; a successful open
    closes it.

## Constraints

- **Offline.** Examples, their sample JSON and thumbnails ship in the verified offline bundle
  alongside `starter.folio`; nothing is fetched at runtime.
- **Sample data stays outside the `.folio`** (format rule: the template never stores sample
  data or its path). An example is a template file plus a sibling JSON file.
- **Only shipped capabilities.** Examples use existing element types (`text`, `image`, `table`,
  `line`, `rect`, `barcode`, `qrcode`) and fonts the Designer already carries — no charts, no
  font fetched on open.
- **Thumbnails come from the engine.** Each is produced at build time from the engine's own
  render of the example with its sample data, so a thumbnail cannot drift from what opens.
- **No new per-user persistence.** No localStorage or IndexedDB entry records dialog state;
  the dialog shows on every launch.
- **Visual language is DESIGN.md's.** Square corners, the existing sheet shadow over the scrim,
  cyan for the selected option, amber only for data, stroked SVG icons, no emoji.

## Non-goals

- A user-saved template library, "save as template", or remote/downloadable templates.
- A "don't show again" setting or first-run-only behaviour.
- Localized or Thai-language examples.
- A guided tour, tooltips walkthrough, or explanatory copy beyond each option's name and one
  line of description.
- Charts or any element type the engine does not ship.

## Success signal

A new author opens the Designer offline, picks Bank Statement, and within one interaction is
looking at a multi-page Preview of a statement with realistic transactions — then switches to
Design, rebinds a column against the loaded sample, and saves the result as their own `.folio`.

## Assumptions

- An opened example's document title is the example's name; it has no file target until Save.
- The unsaved-changes warning offers Keep editing and Discard only — no Save-first action.
- Save sample data writes the loaded sample's exact bytes and lives with the existing sample
  data controls, available for any loaded sample, not only an example's.
