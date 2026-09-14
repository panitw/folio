---
id: SPEC-multi-pages
companions:
  - brownfield.md
  - ../spec-folio/folio-format.md
  - ../spec-section-break/SPEC.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Multi-Page Design Canvas

## Why

This is a **pain to solve**. An author building a document that runs to several designed pages, such as a legal contract, cannot lay out page 2 today. A Folio document has one content column. The sheets on the canvas are windows onto that column, and where each one starts depends on the data, not on the author. The author has no page to add, no page to select, and no way to say "this clause block belongs on page 3", or to choose whether that block starts a fresh page or follows on from the content above it. The header and footer already repeat on every output page. Authors need that to hold for pages they design themselves, with one definition they can edit from any page.

## Capabilities

- **CAP-1**
  - **intent:** An author can add a new, empty page to the document from the canvas toolbar above the page.
  - **success:** In a one-page document, clicking Add Page in the canvas controls (not the document bar) shows page 2 on the canvas. The page is empty, shows the shared header and footer, has Page Break on, and becomes the selected page. The action is one undo entry: undo removes page 2 and redo restores it.
- **CAP-2**
  - **intent:** An author can select a page and delete it along with everything on it, after confirming. The last remaining page can never be deleted.
  - **success:** Select page 2 of a three-page document that has elements on it, then click Delete Page. A confirmation names page 2. Cancelling changes nothing. Confirming leaves two pages, and page 3's content has become page 2. One undo restores the page and every one of its elements. With one page left, Delete Page is disabled.
- **CAP-3**
  - **intent:** An author can place, move and edit elements on a specific page. Each content element belongs to exactly one page.
  - **success:** A palette element dropped on page 3 is saved as belonging to page 3, at the page-relative position where it was dropped. Dragging it onto page 1's content band moves it to page 1 in one undo entry. Elements on the other pages do not change.
- **CAP-4**
  - **intent:** The header and footer are a single definition for the whole document. They show on every page and can be edited from any page, and every copy updates together.
  - **success:** In a three-page document, select the header text on page 3 and change it. Pages 1–3 on the canvas and every page of the rendered PDF show the new text. The file holds one header definition, and one undo reverts the change everywhere. Only the current page's copy carries the component's accessible name.
- **CAP-5**
  - **intent:** A render outputs the designed pages in order. A page with Page Break on starts a new output page after the previous designed page, including all of its overflow pages, has ended. The page count includes every output page.
  - **success:** Render a two-page document with Page Break on for page 2, where page 1 holds a `transactions[]` table that runs 3 output pages. Page 2's content starts on output page 4 at its declared positions. The PDF has 4 pages, `Page X of Y` reads 1–4 of 4, and every output page has the header and footer.
- **CAP-6**
  - **intent:** Each designed page can have its own section break. It affects only that page's content and that page's overflow.
  - **success:** Page 1 and page 2 each have a break, and Page Break is on for page 2. When page 1's table crosses page 1's break, page 1's below-line section moves as the section-break rules require, and page 2's output is identical to rendering page 2 alone. The palette's Section Break entry is disabled only for a page that already has a break.
- **CAP-7**
  - **intent:** Pages and their Page Break setting are saved in the `.folio` text format and documented there. An invalid page structure is refused with a load error that says where the problem is.
  - **success:** A multi-page document round-trips through load and save byte-identically, with Page Break on and off. Every existing one-page fixture keeps its bytes, its declared version and its PDF hash. The following are each a registered load error that names the page: a document with zero pages, an element id duplicated across pages, and a keep-together group that spans pages. A later page with no Page Break value loads as on, and a Page Break value on page 1 loads without error and is ignored. `folio-format.md` documents the page structure, the Page Break key and both rules for how one page follows another.
- **CAP-8**
  - **intent:** The canvas draws each designed page, together with its overflow sheets, from geometry the engine computes, and labels each one with its page number.
  - **success:** The canvas projection says which sheets belong to which designed page. The canvas shows the page labels, and the selected page is visibly marked. The DOM-measurement ban test still passes. The Preview of the CAP-5 fixture matches the rendered PDF.
- **CAP-9**
  - **intent:** An author can turn Page Break off for a page, so its content follows directly after the previous page's content instead of always starting a new output page.
  - **success:**
    - **Setup:** in the CAP-5 document, select page 2 and clear Page Break in the page section of Page Setup. This is one undo entry.
    - **Fits:** page 1's rows end 200pt above the bottom of output page 3, and page 2's declared content is 150pt tall. Page 2's content is drawn on output page 3, starting where the rows end, with its internal offsets kept, and the PDF has 3 pages.
    - **Doesn't fit:** page 2's declared content is 250pt tall. It starts at the top of output page 4's content window.
    - **No overflow:** page 1 fits on one page. Page 2 renders as output page 2 at its declared positions.

## Constraints

- **Add Page and Delete Page live in the canvas controls toolbar above the page**, beside zoom, grid, snap, duplicate and delete. They never go in the document bar.
- **Delete Page always asks for confirmation, naming the page number** (for example "Delete page 3?"). This applies even to an empty page. Confirming is one engine command and one undo entry that removes the page and its elements. The document always has at least one page.
- **Page Break is a per-page checkbox, on by default.**
  - **On:** the page starts a new output page after the previous designed page and all of its overflow.
  - **Off:** the page's content, as one block with its page-relative offsets, begins where the previous page's content ends on that page's last output page. That happens only when the whole block fits in the space left there. Otherwise it begins at the top of the next output page's content window. The block never moves up onto an earlier output page.
  - **Page 1:** Page Break does not apply, and the checkbox is disabled. A value on page 1 in the file is ignored and is not an error.
- "Where the previous page's content ends" means the bottom of that page's last content item on its final output page, counting a table's `minHeight` floor and that page's below-line section, the same measure the section-break rules use.
- **Selecting a page shows Page Setup.** The document-wide settings appear first, then a section for the selected page that holds its Page Break checkbox. A page becomes selected when the author clicks empty space on any of its sheets or clicks its page label.
- **The header and footer are document-level and single-source.** No page can have its own header or footer, override them, or leave them out. Every output page, overflow pages included, draws them. Every page's copy on the canvas can be selected and edits the one definition. For assistive technology, only the current page's copy is named and interactive. The current page is the page the author last selected something on. Clicking page 3's header copy makes page 3 current. Otherwise it is the page of the selected element, or page 1 when nothing is selected.
- **Page setup is document-wide.** Size, orientation, margins and header/footer band heights apply to every page.
- **An element belongs to exactly one page.** Element ids are unique across the whole document. A keep-together group cannot span pages, and a move that would split a group across pages is refused as a whole.
- **Section-break rules apply per page, unchanged** (see `../spec-section-break/SPEC.md`). There is at most one break per page, and it only affects that page's own content column and overflow.
- **The engine remains the only authority on layout and undo.** Adding, deleting and selecting pages, moving an element across pages, and toggling Page Break all go through engine commands or projections. The designer never measures the DOM to get geometry.
- **Format versioning is a 4.x MINOR change.** A one-page document keeps its file bytes, its declared version and its PDF hash. Only documents with more than one page declare the new minor. A multi-page file lists every page, page 1 included, in a top-level `pages` array, and `bands.content` holds no elements. A one-page file keeps today's `bands.content` shape. Content in both places, and a `pages` array with a single entry, are load errors.
- **Page Break is always saved explicitly.** Every page after page 1 stores the value as `true` or `false`, so a restored file states the setting rather than relying on a default. A later page with no value in the file loads as on.
- Output stays byte-identical across darwin/arm64, linux/amd64, linux/arm64 and js/wasm (AD-21). Within a page, AD-24 (no element is displaced because a neighbour grew) still holds. A multi-page golden fixture covering CAP-5 and CAP-9 joins the four-platform hash matrix, and every new diagnostic is registered.
- The canvas sheet cap (`MAX_CANVAS_SHEETS`) applies to the whole stack, across all pages. The canvas draws each page's content at its declared position only, never where Page Break off would place it after data pushes it.

## Non-goals

- A page break inside a designed page, such as a break element placed partway down a page. Page Break exists only as the per-page setting.
- Reordering, duplicating, or copying pages.
- Page setup per page, such as mixed sizes or orientations.
- A header or footer per page, a first-page-only header or footer, or hiding the header or footer on specific pages.
- More than one section break per page.
- Content that flows across designed pages, as in a linked text frame. Page Break off moves a page's content as one block and never reflows it.
- Pulling a page with Page Break off up onto an earlier output page when the previous page has no overflow.

## Success signal

- An author opens a blank document and uses Add Page twice to build a three-page contract, with Page Break on. They put clauses on pages 1 and 2 and a signature block on page 3, then edit the header's party name once from page 2. Then they turn Page Break off for page 3 and render with a long clause table on page 2. Every PDF page shows the new party name. The signature block sits directly under the last table row, or at the top of the next page when it doesn't fit. The hash is the same on all four platforms.

## Assumptions

- Add Page inserts the new page after the selected page. When no page is selected, it goes at the end.
- Dragging an element onto another page's content band moves it to that page.
- The new minor version follows the Anchor precedent. It joins 4.1 while 4.1 is unreleased, and otherwise becomes the next 4.x minor.
- The confirmation also applies to an empty page, which keeps the rule to one behaviour.
- A Page Break value on page 1 is dropped on save and produces no diagnostic, matching how an explicit `sectionBreakAnchor: true` is dropped.
