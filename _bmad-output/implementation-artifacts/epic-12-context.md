# Epic 12 Context: The inspector reaches the engine that is already there

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Six capabilities already ship in the engine — accepted by the loader, consumed by the renderer, covered by goldens — and no control in the designer reaches any of them. Page-header and page-footer band heights have no writer at all, so a letterhead is stuck at whatever the starter file declared. The document's `locale` and `utcOffset` have no control anywhere, so every date and number formats as `en` in a product whose reason to exist is Thai and CJK statements. A table's header height, header style and alternating row background are authorable only by hand-editing the file the designer just saved. This epic closes that gap from the panel side. It is the mirror of Epic 9 — there the panel offered controls the engine did not honour; here the engine honours values the panel does not offer — and both leave the author editing JSON by hand to finish what the designer started, when hand-editability was meant as a second door, not the only way through. The epic adds **no new rendering behaviour**: every value it makes authorable is one the engine already accepts and consumes, so a document declaring none of them must hash identically.

## Stories

- Story 12.1: The page header and page footer take the height the author sets
- Story 12.2: The document declares its locale and its UTC offset
- Story 12.3: A table's header and its alternating rows are authorable
- Story 12.4: Padding is honoured outside a table, or the format says it is not
- Story 12.5: A band boundary is dragged on the canvas

## Requirements & Constraints

- **Band heights are authorable**, in points, alongside the page margins, and reachable both as a typed number and as a canvas gesture. A height that would leave no content window — header plus footer at or beyond the page's inner height — is refused with a located message naming the quantity and the space available, and the document is left unchanged. A height reduced below the content its band already holds is *accepted*; the overflow clips with the diagnostic the engine already defines, never a second overflow rule.
- **Document formatting settings are authorable**: the locale as a closed set (`en`, `th`, `zh-Hans`, `ja`) and the UTC offset as a `±HH:MM` field, both seeded from the engine's current values. A malformed offset is refused with a located message and no document change. A locale change must take effect for every date and number expression on the next preview, including Buddhist-era years for `th`, with nothing else in the document altered.
- **Table header and row styling are authorable** from the table editor that today configures columns only: header height, the header-row style fields, and the alternating row background — each showing the engine's committed value and each clearable back to *absent*, so the documented cascade still applies.
- **Padding is ruled, not built.** The ruling is settled: `style.padding` stays an engine property — a loaded document keeps and renders it, the panel never authors it, and the command layer refuses padding fields on non-table component kinds while the loader keeps accepting them for compatibility. The format states the restriction. The former two-way fork (insets text vs table-only) is closed.
- **Byte-identity is the acceptance gate for the whole epic.** A document whose new values are not edited serializes to identical bytes, including a key that was absent staying absent, and no document in the corpus changes hash.
- Every value made authorable ships with a golden fixture, and a hash change is investigated as a defect until proven to be an intended, versioned behaviour change.

## Technical Decisions

- **The engine owns the document.** There is no TypeScript model of a `.folio`. Every committed mutation travels as one versioned opaque command over the single channel; Go validates and applies it; the canvas re-projects from the engine's accepted document, never from a browser-side model of the bands. Transient interaction state — a drag in flight, an uncommitted keystroke — lives in the UI and never enters the document. Undo/redo is engine-side history over committed commands.
- **The content window is derived, never stored.** Page height minus margins minus the two band heights, computed by one function in layout. Nothing may cache a second copy of that arithmetic, and a table's width remains the sum of its column widths.
- **Formatting locale is declared, never discovered.** The locale set is closed; an unlisted tag is a load error rather than a fallback. The engine reads no host time zone or locale on any render path.
- **Boxes are absolute and nothing negotiates.** Shrinking a band never moves, deletes or reflows the components inside it — they clip, with a diagnostic, and stay where the author put them.
- **Diagnostics are one type on one channel** with stable codes: clipped content is a warning returned alongside successful PDF bytes, and it locates the offending element.
- **The style cascade already exists; do not reimplement it.** The header-row style governs the header row exclusively; data cells cascade from the table's own style; a column's own alignment still wins for that column. An absent header-style field falls back to the table's style and then to that field's documented default. The alternating background follows the collection index (so it does not reset per page) and wins over the table background on the rows where it applies.
- These stories add no format keys, so the format's document version is not raised.

## UX & Interaction Patterns

- **Document-level settings live in the page setup panel.** No mockup draws locale or UTC offset — the drawn inspector covers properties, position, typography, content, bands and components only. Follow the shipped page-setup panel's existing row idiom rather than inventing a surface.
- **Band tabs are the reach point for band height.** Each tab sits at the top-left of the band it names. Dragging the CONTENT tab moves the header/content boundary (the page header's height); dragging the PAGE FOOTER tab moves the content/footer boundary. The PAGE HEADER tab sits against the page margin, not a band boundary, and must not resize anything — a control that looks draggable and is not is worse than one that does not.
- **A boundary drag is a transient local proposal**, exactly like a component drag: the canvas paints the proposed boundary, the height reads out in points as it moves, nothing is committed, and pointer release sends the one band-height command whose accepted value replaces the proposal. It snaps to the same grid increment as every other drag.
- **The engine's refusal becomes a drag limit.** The boundary stops at the point where no content window would remain, so an impossible document is never proposed and then rejected.
- **The gesture is an affordance on top of the command, never the only way in.** The boundary is keyboard-focusable and resizable by the same nudge keys that move a component; every new control is keyboard-reachable with visible focus, and the table editor keeps behaving as a data grid under keyboard navigation.
- Refusal and diagnostic copy stays terse and technical: state the fact, name the location, offer no comfort.
- Relevant mockups: the workspace artboard (document bar's page-setup summary, inspector, band tabs) and the table editor artboard (HEADER and CELLS sections beside the column matrix), under `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/`.

## Cross-Story Dependencies

- **12.5 follows 12.1**: the drag commits through 12.1's command, so the number lands before the gesture. The epic is not complete with only the number.
- **12.1–12.4 are otherwise independent** of one another and parallelisable.
- **Epic 14's table-editor presentation story follows 12.3**: 12.3 makes the header and alternating-row values authorable and owns their storage; building the presentation first would force it to invent storage 12.3 already owns.
- **12.3 consumes existing engine behaviour**: alternating-row painting and the header/cell cascade already ship; this story adds no second implementation of either.
- **12.4 has no code fork left** — the owner ruling is recorded and the format already states the restriction; what remains is the command-layer refusal and the format's own wording.
