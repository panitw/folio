# Epic 12 Context: The inspector reaches the engine that is already there

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Several capabilities already ship in the engine — accepted by the loader, consumed by the renderer, covered by goldens — and no control in the designer reaches any of them. The page-header and page-footer band heights have no writer at all, so a letterhead is stuck at whatever height the starter file declared. The document's locale and UTC offset have no control anywhere, so every date and number formats as `en`, in a product whose reason to exist is Thai and CJK statements. A table's header height, header style and alternating row background are authorable only by hand-editing the file the designer just saved. This epic closes that gap from the panel side. It is the mirror of Epic 9 — there the panel offered controls the engine did not honour; here the engine honours values the panel does not offer — and both leave the author editing JSON by hand to finish what the designer started, when hand-editability was meant as a second door, not the only way through. The epic adds **no new rendering behaviour**: every value it makes authorable is one the engine already accepts and consumes, so a document declaring none of them must hash identically.

## Stories

- Story 12.1: The page header and page footer take the height the author sets
- Story 12.2: The document declares its locale and its UTC offset
- Story 12.3: A table's header and its alternating rows are authorable
- Story 12.4: Padding is honoured outside a table, or the format says it is not
- Story 12.5: A band boundary is dragged on the canvas

## Requirements & Constraints

- **Band heights are authorable** in points, beside the page margins, reachable both as a typed number and as a canvas gesture. A height that would leave no content window — header plus footer at or beyond the page's inner height — is refused with a located message naming the quantity and the space available, and the document is unchanged. A height reduced below the content its band already holds is *accepted*; the overflow clips with the diagnostic the engine already defines, never a second overflow rule.
- **Document formatting settings are authorable**: locale as a closed set (`en`, `th`, `zh-Hans`, `ja`) and UTC offset as a `±HH:MM` field, both seeded from the engine's current values. A malformed offset is refused with a located message and no document change. A locale change takes effect for every date and number expression on the next preview, including the Buddhist-era year `th` already produces, with nothing else altered.
- **Table header and row styling are authorable** from the table editor that today configures columns only: header height, header-row style fields, and alternating row background — each showing the engine's committed value and each clearable back to *absent*, so the documented cascade still applies.
- **Padding is ruled, not built.** The ruling is settled and recorded: `style.padding` stays an engine property — a loaded document keeps and renders it, the panel never authors a control for it, and the command layer stops accepting `padding*` on non-table component kinds while the loader keeps accepting them for compatibility. The format states the restriction. The two-way fork the story text offers (insets text vs table-only) is closed; do not re-open it.
- **Byte-identity is the epic's acceptance gate.** A document whose new values are not edited serializes to identical bytes, including a key that was absent staying absent, and no corpus document changes hash. Every value made authorable ships a golden fixture; a hash change is a defect until proven an intended, versioned change.

## Technical Decisions

- **The engine owns the document.** No TypeScript model of a `.folio` exists. Every committed mutation travels as one versioned opaque command over the single channel; Go validates and applies it; the canvas re-projects from the engine's accepted document, never from a browser-side model of the bands. Transient state — a drag in flight, an uncommitted keystroke — lives in the UI and never enters the document. Undo/redo is engine-side history over committed commands.
- **The content window is derived, never stored.** Page height minus margins minus the two band heights, by one function in layout. Nothing caches a second copy of that arithmetic, and a table's width stays the sum of its column widths.
- **Formatting locale is declared, never discovered.** The set is closed; an unlisted tag is a load error, not a fallback. No render path reads a host locale or time zone. The locale table is versioned with the library.
- **Boxes are absolute and nothing negotiates.** Shrinking a band never moves, deletes or reflows the components inside it — they clip, with a diagnostic, and stay where the author put them.
- **Diagnostics are one type on one channel** with stable codes: clipped content is a warning returned alongside successful PDF bytes and names the element. Errors (render cannot proceed) and warnings (render proceeded, with a caveat) must not look alike.
- **The table styling cascade already exists; do not reimplement it.** An absent header-style field falls back to the table's own style, then to that field's documented default. Alternating backgrounds follow the row's index in the collection, so alternation does not reset per page.
- No new format keys, so the document format version is not raised.

## UX & Interaction Patterns

- **Document-level settings live in the page setup panel.** No mockup draws locale or UTC offset — the drawn inspector covers properties, position, typography, content, bands and components only. Follow the shipped page-setup panel's row idiom rather than inventing a surface.
- **Band tabs are the reach point for band height.** Dragging the CONTENT tab moves the header/content boundary (the page header's height); dragging the PAGE FOOTER tab moves the content/footer boundary. The PAGE HEADER tab sits against the page margin, not a boundary, and must not resize anything — a control that looks draggable and is not is worse than one that does not.
- **A boundary drag is a transient local proposal**, exactly like a component drag: the canvas paints the proposed boundary, the height reads out in points as it moves, nothing is committed, and release sends the one band-height command whose accepted value replaces the proposal. It snaps to the same grid increment as every other drag, and stops where no content window would remain — so an impossible document is never proposed and then rejected.
- **The gesture is an affordance on top of the command, never the only way in.** The boundary is keyboard-focusable and resizable by the same nudge keys that move a component; every new control is keyboard-reachable with visible focus and an accessible name, and the table editor keeps behaving as a data grid under keyboard navigation.
- Band boundaries are drawn dashed cyan, a reserved meaning; band identity must stay legible at all times, because which band a component sits in determines whether it repeats on every page.
- Copy stays terse and technical: name the element and the location, never apologize, prefer a noun label over a sentence, and use the format's own vocabulary (*band*, *page header*, *row scope*).
- Mockups live under `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/`: the workspace artboard (page-setup summary, inspector, band tabs) and the table editor artboard (HEADER and CELLS sections beside the column matrix).

## Cross-Story Dependencies

- **12.5 follows 12.1**: the drag commits through 12.1's command, so the number lands before the gesture. The epic is not complete with only the number.
- **12.1–12.4 are otherwise independent** and parallelisable.
- **12.3 consumes existing engine behaviour**: alternating-row painting and the header/cell cascade already ship; this story adds no second implementation and no second place to store the values.
- **Epic 14's table-editor section story follows 12.3.** Reduced by owner decision to presentation only, it groups the controls 12.3 ships into the drawn HEADER/CELLS/BORDERS sections and invents no field the format lacks; building it first would force it to invent storage 12.3 owns.
- **12.4 gates a later cell-padding control** — settling what padding means is why the ruling comes first, and any designer-side padding write path introduced elsewhere must not outlive it.
