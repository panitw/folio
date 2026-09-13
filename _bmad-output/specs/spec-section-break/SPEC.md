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

A **pain to solve**. Bank statements put a fixed block, such as the transaction-code legend, near the foot of the content band, under a `transactions[]` table whose length depends on the data. Folio never moves an element because a neighbour grew (AD-24). When the rows reach the legend they draw over it, and the statement is unreadable. The author has no way to say "this block comes after the table". A section break is a line the author places above the block. If content above the line runs past it, everything below the line moves to the next page, at the same position it was designed at.

## Capabilities

- **CAP-1**
  - **intent:** An author can add one section break to the content band, move it vertically and remove it on the design canvas.
  - **success:** On the canvas, add a break, drag it to a new offset, undo the drag, then remove the break. Each step round-trips through the engine and appears as exactly one undo entry. Moving or resizing an element across the line is refused, and so is dragging the line through an element. A refused edit leaves the document unchanged and tells the author which element blocked it.
- **CAP-2**
  - **intent:** Everything below the break lands on the page where the content above it ends, when that content ends at or above the line. Otherwise it lands on a new page that follows. In both cases each element keeps its declared position.
  - **success:** Render the golden statement with the legend below the break. At 5 rows the legend is on page 1 at its declared y. When the last row crosses the line on page 1, the legend is on page 2 at the same y, the page is empty above it, and no row overlaps it. When the rows fill page 1 and end above the line on page 2, the legend is on page 2 at its declared y.
- **CAP-3**
  - **intent:** A break that the content above it never crosses has no effect on the output.
  - **success:** When no data crosses the line, the template with a break and the same template without it render PDFs with identical hashes.
- **CAP-4**
  - **intent:** A page added for the below-line section is a full page of the document.
  - **success:** When the section moves to an added page, that page has the page header and footer, the page count includes it, and `Page X of Y` is correct on every page.
- **CAP-5**
  - **intent:** The break persists in the `.folio` text format, is documented there, and an invalid break is refused with a located error.
  - **success:** Load then save round-trips byte-identically. A break outside the content band, or a second break, is a registered load error naming the band. An element that sits on both sides of the break is a registered load error naming that element. `folio-format.md` documents the key and the pagination rule.
- **CAP-6**
  - **intent:** The canvas shows the break line and the below-line section with geometry that comes from the engine, matching what the PDF does.
  - **success:** The canvas projection carries the break offset and the section's membership. The section is drawn at its declared position only, with no copy at a page seam. A DOM-measurement ban test still passes. Preview of the golden statement matches the rendered PDF.

## Constraints

- **This is a narrow, deliberate exception to AD-24 and to the rule that a window collapses the gap above its first element.** It applies only to the below-line section. The section moves as a rigid block, only vertically, and only by whole pages. Each member keeps its declared `y` within the band. The exception must be written into the *Pagination* section of `folio-format.md`. No other element gains the ability to move.
- **No element may sit on both sides of the break.** An element whose top is above the line and whose bottom is below it is not allowed, so every element is unambiguously above or below the line, whatever its `x`. The section spans the full band width.
- "Where the content above the break ends" means the bottom of the last above-line item on its final page. A table's `minHeight` floor counts toward that bottom.
- **The Section Break wins over `keepTogether`.** A group with members on both sides of the break is split at the line. Members above the line stay where they are, and members below it move with the section. Members on the same side of the line still stay together. The split produces a registered warning naming the group, and `-strict` fails on it.
- The setting is called **Section Break** in the designer. In the `.folio` file it is the optional content-band key `sectionBreak`, holding an offset in points.
- The break is an optional new key, so it is a **MINOR** format change (4.0 → 4.1). Only documents that carry a break declare the new version. Documents without one keep their version and their bytes.
- **The break is never drawn in the PDF.** It is a marker on the design canvas only. A printed rule is still a `line` element.
- Rows of the above-line content may use the full page below the line on every page where the section does not land. The line reserves space only on the page where the section lands.
- Output stays byte-identical across darwin/arm64, linux/amd64, linux/arm64 and js/wasm (AD-21). All arithmetic is integer fixed-point. The engine is the only layout authority, including on the canvas.
- If the below-line section is taller than the space left on its landing page (for example, it holds a growing table), it continues onto later pages under the existing pagination rules, starting from its declared offset.
- The design canvas draws the below-line section only at its declared position. It never draws where the section would land after being pushed, because the canvas has no data.
- A golden fixture covering the behaviour in CAP-2 joins the four-platform hash matrix. Every new diagnostic is registered in the diagnostic registry.

## Non-goals

- More than one break per content band, or breaks chained so that one pushed section pushes another.
- Breaks in the page header or page footer.
- A page-break-always element, or any setting that forces a new page when nothing crosses the line.
- General reflow: pushing individual elements down by the amount a neighbour grew, or moving a pushed section to the top of the page.
- A printed or styled break line.
- Widow and orphan control for the rows above the break.

## Success signal

- The KKP deposit statement from the owner's screenshots, with a section break above the transaction-code legend, is rendered at 5, 40 and 80 transactions. On no page is the legend overdrawn. The legend appears exactly once, at its designed position, on the last page. The PDF hash is identical on all four platforms.