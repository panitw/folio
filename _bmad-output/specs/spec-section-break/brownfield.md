# Existing integration points

Inspected on 2026-09-13. Paths are relative to the repository root, and line anchors may move. These are starting points and rules that must not break. They do not widen the scope.

## Engine

| Source | Current behaviour | Implication |
|---|---|---|
| `folio-go/internal/template/model.go` | The content band has `Elements` and no stored height. Element `X`/`Y` are relative to the band. | The break is the optional content-band field `sectionBreak`. Reject any element that sits on both sides of it. Validate it against the derived content height (`layout.ContentHeight`), never a stored height. |
| `folio-go/internal/layout/paginate.go`, `Paginate` (~664) | The content is one column of unbounded height. Each window starts at the first item that did not fit, and the gap above that item collapses. | The below-line section is the one exception to that collapse. It starts a window at its declared offset, not at the top of the page. Items above the line paginate exactly as they do today. |
| `folio-go/render.go`, `paginateDocument` (~2748), `contentColumnItems` (~2475) | Builds the column items and applies per-page shifts (~3229). | Split the items at the break by declared top edge. Paginate the items above it, then decide the section's landing page from where they end compared with the break offset. |
| `folio-go/internal/layout/table_slice.go`, `ElementPush`; `render.go` `elementPushFor` (~3233–3290) | The only existing push. It moves elements below a table by the space its `minHeight` reserves. It does not push elements beside the table. | Do not overload it. The floored bottom it produces counts toward where the above-line content ends. |
| `render.go`, keep-together (`keepTogetherTags`, `orKeepTogether`, ~2757–2907) | A tagged group is placed as one unit. | A group with members on both sides of the break is split at the line, and the Section Break wins. Each side stays together with its own members. The split raises a registered warning naming the group. |
| `folio-go/table_render.go`, `collectBandTableRuns` (~761) | Places rows one after another down the column, and repeats the header on continuation pages. | Unchanged. Rows may run past the break on every page except the landing page. |
| `folio-go/page_number.go` | Computes the page count for `Page X of Y`. | It must include a page added for the section. |
| `folio-go/page_setup.go` (~426), `CanvasProjection` (`folio-designer/src/engine-protocol.ts:415`) | The canvas geometry comes from the engine, including window origins. | Add the break offset and section membership to the projection. The designer computes neither. |

## Format

- `_bmad-output/specs/spec-folio/folio-format.md`, *Pagination* (~355): add a rule for the section break, and qualify the AD-24 sentence ("no element is ever displaced…") and the gap-collapse note so that the exception is stated in the document.
- *Bands* (~340): document the new content-band key.
- The `version` derivation (line 47): a break raises the document's version to `4.1`. Nothing else about the version changes.

## Designer

- `folio-designer/src/band-boundary.ts` and `e2e/band-boundary-drag.spec.ts` are the precedent for a draggable horizontal line that the engine validates. Reuse that interaction and command shape, and label the control "Section Break".
- `folio-go/component_commands.go`, `moveComponent` (~1820) and `updateComponentProperties` (~800): the engine refuses any move or resize that would leave an element on both sides of the break, and refuses a break drag that would put the line through an element. Group moves are refused as a whole.
- `folio-designer/src/App.tsx`: band sections (~2953), the `page-seam` marker, and `sheet-stack.ts` occurrences. The section is drawn at its declared position only; do not add a pushed copy at the page seam.
- `folio-designer/src/canvas-authority-contract.test.ts` bans measuring the DOM for geometry. The break must follow that ban.

## Verification locations

- Add Go tests beside `folio-go/table_pagination_test.go` for the three landing cases in CAP-2, for no-crossing byte identity (CAP-3), and for page count (CAP-4).
- Add a golden fixture under `fixtures/` that is registered in `hashmatrix/`.
- Add a template round-trip and load-error test for each new diagnostic.
- Add a designer e2e test for add, drag, undo and remove of the break.
