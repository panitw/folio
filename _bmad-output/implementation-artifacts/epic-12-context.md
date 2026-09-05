# Epic 12 Context: The inspector reaches the engine that is already there

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Six capabilities already ship in the engine — validated by the loader, consumed by the renderer, covered by goldens — and no control in the designer reaches any of them. Band heights for the page header and page footer have no writer at all; the document's `locale` and `utcOffset` have no control anywhere, so every date and number formats as `en` in a product whose reason to exist is Thai and CJK statements; a table's header height, header style and alternating row background are authorable only by hand-editing the file the designer just wrote. This epic closes that gap from the panel side. It is the mirror of Epic 9 — there the panel offered controls the engine did not honour; here the engine honours values the panel does not offer — and both leave the author hand-editing JSON to finish what the designer started, when hand-editability was meant as a second door, not the only way through. The epic adds **no new rendering behaviour**: every value it makes authorable is one the loader already accepts and the renderer already consumes.

## Stories

- Story 12.1: The page header and page footer take the height the author sets
- Story 12.2: The document declares its locale and its UTC offset
- Story 12.3: A table's header and its alternating rows are authorable
- Story 12.4: Padding is honoured outside a table, or the format says it is not
- Story 12.5: A band boundary is dragged on the canvas

## Requirements & Constraints

- **Band heights authorable** — page-header and page-footer heights are settable from the designer, and reachable both as a number and as a canvas gesture.
- **Document formatting settings authorable** — `locale` and `utcOffset` are settable from the designer; a locale change must take effect for every date and number expression in the document on the next preview.
- **Table chrome authorable** — a table's header height, header-row style fields, and alternating row background are settable from the table editor, each clearable back to absent so the documented cascade still applies.
- **Byte stability is the acceptance gate.** A document that declares none of these values must serialize to identical bytes, including keys that were absent staying absent, and the rendered corpus must hash unchanged. Every visual behaviour ships with a golden fixture; a moved hash is investigated as a defect, never regenerated.
- **Refusals are located.** Every rejected value produces a diagnostic naming the element/quantity and leaves the document unchanged. Errors (render cannot proceed) and diagnostics (render proceeded with a caveat) must not look alike.
- **Overflow reuses the existing answer.** A band shortened below the content it holds clips that content with the established overflow diagnostic; components are never moved, deleted or reflowed.
- **A document must always have a content window.** Header plus footer at or beyond the page's inner height is refused, with the message naming the quantity and the space available.
- **Accessibility floor.** Every new control and the band boundary itself are keyboard-reachable and operable, with visible focus and accessible names on icon-only controls. A drag is an affordance on top of a command, never the only route to a value.

## Technical Decisions

- **The engine owns the document.** There is no TypeScript model of a `.folio`; the wasm engine parses, holds, mutates, validates and serializes it. The UI paints from an immutable snapshot and sends every committed mutation as one versioned opaque command over the single channel. Band heights therefore travel exactly as the existing page-setup command does, and the canvas re-projects from the engine's accepted document — never from a browser-side model of the bands.
- **Transient interaction state never enters the document.** A drag in flight is a local proposal: the canvas paints the proposed boundary and shows the height in points, nothing is committed, and pointer release sends the one command whose accepted value replaces the proposal. Undo/redo is engine-side history over committed commands.
- **Locale is a declared, closed set.** `en`, `th`, `zh-Hans`, `ja` only — an unlisted tag is a load error, not a fallback. `th` renders Buddhist-era years. The engine never reaches for host locale or time zone; the UTC offset is a fixed `±HH:MM` the document declares.
- **Table geometry is derived, not stored twice.** The content area's height is computed as page height minus margins minus header height minus footer height by one function in the layout package; nothing else may recompute it.
- **Boxes are absolute and nothing negotiates.** An element's coordinates are relative to its band's top-left; bands are placed on the page by layout alone; nothing in a band reflows.
- **Style cascade is not reimplemented.** Alternating row backgrounds paint through the existing renderer rule, and an absent header-style field falls back to the table's own style and then to that field's documented default via the cascade that already exists. No second implementation.
- **Padding (12.4) is a ruling first, code second.** The format documents `padding` with a default of `0` on all four edges scoped to no element kind, the command layer accepts all four keys on any component, and the only consumer on any render path is a table's cell chrome — inert on text, image, rect and line. The owner must rule, and the ruling be recorded in the decision log, **before any code is written**: either padding insets a text element's content box (reusing the same helper a table cell already computes, with the width-overflow diagnostic measuring against the inset width), or padding is documented as table-only, refused by the command layer on the other four kinds, still accepted by the loader for compatibility, and the restriction stated in the format's Style table.

## UX & Interaction Patterns

- **Where the controls live.** Band heights and the document's locale/UTC offset belong in the inspector's PAGE SETUP panel, beside the margins, in points — that panel is already where document-level settings live. Header height, alternating row background and header-style fields belong in the table editor's HEADER and CELLS sections, beside the column matrix, which stays presented as a matrix (columns against attributes) rather than a repeated single-column form.
- **The band tabs are the drag handle.** Each tab already sits at its band's top-left. Dragging the **CONTENT** tab moves the header/content boundary (the page-header height); dragging the **PAGE FOOTER** tab moves the content/footer boundary (the page-footer height). The **PAGE HEADER** tab sits on the page margin, not on a band boundary, and must not resize anything — a control that looks draggable and is not is worse than one that does not look draggable.
- **Drag behaves like every other drag.** It snaps to the same grid increment when snapping is on, and releases the same way. A drag that would leave no content window, or a negative height, stops the boundary at that limit rather than proposing a document the engine will refuse.
- **Keyboard parity.** The boundary is focusable and resizable by the same nudge keys that move a component.
- **Voice.** Prefer the noun over the sentence for labels (`Page header`, not `Set the page header`). Use the product's own vocabulary — *band*, *binding*, *aggregate*, *row scope*; the `.folio` file uses these words too.

## Cross-Story Dependencies

- **12.5 depends on 12.1.** The command comes first because the drag commits through it; the epic is not complete with only the number. 12.5 also enforces 12.1's engine rule (no content window, no negative height) as a drag limit.
- **12.3 depends on the existing table-cascade and alternating-row rendering already delivered** — it is an authoring surface over rules that ship, not new rendering.
- **12.4 is blocked on an owner ruling** recorded in the decision log before implementation; its outcome determines whether the work lands in the render path or in the command layer plus the format doc.
- **Epic 9 is the counterpart** — this epic assumes the box-printing work it did, and must not re-open it.
