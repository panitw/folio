---
id: SPEC-section-break
companions:
  - brownfield.md
  - ../spec-folio/folio-format.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Section Break

## Why

A **pain to solve**. Bank statements put a fixed block, such as the transaction-code legend, near the foot of the content band, under a `transactions[]` table whose length depends on the data. Folio never moves an element because a neighbour grew (AD-24). When the rows reach the legend they draw over it, and the statement is unreadable. The author has no way to say "this block comes after the table". A section break is a line the author places above the block. If content above the line runs past it, everything below the line moves out of the way. When the break is **anchored** the section moves to the next page at the same position it was designed at. When it is **not anchored** the section follows directly after the content that pushed it, on the same page if there is room.

## Capabilities

- **CAP-1**
  - **intent:** An author can add one section break to the content band from the palette, then select, move, edit and remove it on the design canvas as if it were an element.
  - **success:** With no break, arm the Section Break palette entry and click the content band: the break lands at that height (snapped), and the entry disables. Enter on the armed content band places it at the middle of that page's content window. Select it by clicking, drag it, arrow-nudge it (1pt, Shift 10pt), type its Y in Properties, and delete it with Delete. Each step round-trips through the engine as exactly one undo entry. Moving or resizing an element across the line is refused, and so is placing, dragging or typing the line through an element. A refused edit leaves the document unchanged and tells the author which element blocked it.
- **CAP-2**
  - **intent:** With an anchored break, everything below the break lands on the page where the content above it ends, when that content ends at or above the line. Otherwise it lands on a new page that follows. In both cases each element keeps its declared position.
  - **success:** Render the golden statement with an anchored break above the legend. At 5 rows the legend is on page 1 at its declared y. When the last row crosses the line on page 1, the legend is on page 2 at the same y, the page is empty above it, and no row overlaps it. When the rows fill page 1 and end above the line on page 2, the legend is on page 2 at its declared y.
- **CAP-3**
  - **intent:** A break that the content above it never crosses has no effect on the output, anchored or not.
  - **success:** When no data crosses the line, the template with a break (Anchor on or off) and the same template without it render PDFs with identical hashes.
- **CAP-4**
  - **intent:** A page added for the below-line section is a full page of the document.
  - **success:** When the section moves to an added page, that page has the page header and footer, the page count includes it, and `Page X of Y` is correct on every page.
- **CAP-5**
  - **intent:** The break and its Anchor setting persist in the `.folio` text format, are documented there, and an invalid break is refused with a located error.
  - **success:** Load then save round-trips byte-identically, anchored and unanchored. A break outside the content band, a second break, or an Anchor setting with no break is a registered load error naming the band. An element that sits on both sides of the break is a registered load error naming that element. `folio-format.md` documents both keys and both pagination rules.
- **CAP-6**
  - **intent:** The canvas shows the break line and the below-line section with geometry that comes from the engine, matching what the PDF does.
  - **success:** The canvas projection carries the break offset, its Anchor setting and the section's membership. The section is drawn at its declared position only, with no copy at a page seam. A DOM-measurement ban test still passes. Preview of the golden statement matches the rendered PDF.
- **CAP-7**
  - **intent:** An author can turn Anchor off so that, when content above crosses the line, the section is pushed down to sit right where that content ends instead of jumping to its declared position on a new page.
  - **success:** Render the golden statement with an unanchored break 20pt above the legend. At 5 rows the legend is at its declared y. When the rows end 120pt below the line on page 1 and the legend fits in the room left, the legend is on page 1, 20pt below the end of the rows, and no row overlaps it. When it does not fit, it is on page 2 with the line at the top of the content window and the legend 20pt below it. When the rows fill page 1 and end on page 2 higher than the line's declared offset, the legend is on page 2, 20pt below the end of the rows. Toggling Anchor in the Section Break Properties is one engine command and one undo entry, and it shows or hides the anchor icon in the canvas tab.

## Constraints

- **This is a narrow, deliberate exception to AD-24 and to the rule that a window collapses the gap above its first element.** It applies only to the below-line section, and no other element gains the ability to move. The section always moves as a rigid block, only vertically, and each member keeps its offset from the break line. The exception must be written into the *Pagination* section of `folio-format.md`.
- **Anchored (the default):** the section moves only by whole pages, and each member keeps its declared `y` within the band.
- **Unanchored:** when the content above ends below the line, the line moves down to that end. If the section's declared extent below the line (its lowest member bottom, counting a table's `minHeight` floor) fits in the room left on that page, it is drawn there. Otherwise the whole section moves to a new page with the line at the top of the content window. On page 1 an unanchored section is never pulled up: when the content above ends there at or above the line, the section sits at its declared `y`, exactly as an anchored one does. When the content above runs onto a later page, the line sits where that content ends on its last page, whether that is above or below the declared line, under the same fit rule.
- **No element may sit on both sides of the break.** An element whose top is above the line and whose bottom is below it is not allowed, so every element is unambiguously above or below the line, whatever its `x`. The section spans the full band width.
- "Where the content above the break ends" means the bottom of the last above-line item on its final page. A table's `minHeight` floor counts toward that bottom.
- **The Section Break wins over `keepTogether`.** A group with members on both sides of the break is split at the line. Members above the line stay where they are, and members below it move with the section. Members on the same side of the line still stay together. The split produces a registered warning naming the group, and `-strict` fails on it.
- **In the designer the break behaves like an element; in the file it is a band key.** It is added from a palette entry that is disabled while a break exists, and it has its own selection that never joins component bulk edits, group moves, copy or select all. On the canvas it is a solid 1px cyan line with a mono SECTION BREAK tab, drawn once at its declared position; elements below it carry no marking. When Anchor is on, the tab also shows an anchor icon; when Anchor is off, it shows no icon.
- Band-height and page-setup changes that would leave the break at or below the content band's bottom are refused, naming the break.
- The setting is called **Section Break** in the designer, and its option is called **Anchor**, on by default. Anchor is a checkbox in the Section Break Properties panel. In the `.folio` file the break is the optional content-band key `sectionBreak`, holding an offset in points, and Anchor is the optional boolean content-band key `sectionBreakAnchor`. Declaring `sectionBreakAnchor` without `sectionBreak` is `SECTION_BREAK_INVALID` naming the band, and an explicit `true` is dropped on save.
- **Anchor defaults to on, so existing files and output do not change.** The Anchor key is written only when Anchor is off. It is part of the same **MINOR** format change as `sectionBreak` (4.0 → 4.1, unreleased), with no further bump. Only documents that carry a break declare 4.1. Documents without one keep their version and their bytes.
- **The break is never drawn in the PDF.** It is a marker on the design canvas only. A printed rule is still a `line` element.
- Rows of the above-line content may use the full page below the line on every page where the section does not land. The line reserves space only on the page where the section lands.
- Output stays byte-identical across darwin/arm64, linux/amd64, linux/arm64 and js/wasm (AD-21). All arithmetic is integer fixed-point. The engine is the only layout authority, including on the canvas.
- If the below-line section is taller than the space left on its landing page (for example, it holds a growing table), it continues onto later pages under the existing pagination rules. Anchored, it starts from its declared offset. Unanchored and moved to a new page, it starts from the top of that page's content window.
- The design canvas draws the below-line section only at its declared position. It never draws where the section would land after being pushed, because the canvas has no data.
- A golden fixture covering CAP-2 and CAP-7 joins the four-platform hash matrix. Every new diagnostic is registered in the diagnostic registry.

## Non-goals

- More than one break per content band, or breaks chained so that one pushed section pushes another.
- Breaks in the page header or page footer.
- A page-break-always element, or any setting that forces a new page when nothing crosses the line.
- General reflow: pushing individual elements down by the amount a neighbour grew. Only the below-line section, as one block, ever moves.
- Pulling an unanchored section up on page 1 when the content above ends there short of the line.
- A printed or styled break line.
- Widow and orphan control for the rows above the break.

## Success signal

- The KKP deposit statement from the owner's screenshots, with a section break above the transaction-code legend, is rendered at 5, 40 and 80 transactions, once with Anchor on and once with Anchor off. On no page is the legend overdrawn. The legend appears exactly once on the last page: at its designed position when anchored, and directly after the last row when unanchored and there is room. The PDF hash is identical on all four platforms.
