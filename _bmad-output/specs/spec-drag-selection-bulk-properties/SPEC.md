---
id: SPEC-drag-selection-bulk-properties
companions:
  - interaction-contract.md
  - group-movement.md
  - brownfield.md
sources: []
---

# Designer rectangle selection, group movement, and bulk property editing

## Why

Authors need to select a group spatially, move its elements together, and change shared properties together. The designer currently requires selecting individual elements, moves one element per drag, and exposes an inconsistent set of controls for mixed selections. Rectangle selection, movement that preserves spacing, and a panel based on shared capabilities make these edits predictable.

## Capabilities

- **CAP-1**
  - **intent:** Authors can draw a selection rectangle anywhere in the canvas, across displayed pages and bands, to select fully enclosed elements.
  - **success:** A visible rectangle follows a drag in any direction; completion selects every eligible element whose full projected box is enclosed, including edge equality, and excludes partial overlaps. Selection does not modify the document.
- **CAP-2**
  - **intent:** Authors can replace or extend a selection while retaining existing selection interactions.
  - **success:** Normal rectangle selection replaces selected IDs; Shift-drag adds enclosed IDs without duplicates or toggling existing IDs off. Cancellation restores the starting selection. Existing click and Shift-click selection remain available, with selected-member drag precedence defined in `group-movement.md`.
- **CAP-3**
  - **intent:** Authors can see and edit the properties common to all selected elements.
  - **success:** The panel shows only fields editable for every target, grouped in nonempty sections. Identical values are shown; differing values are blank with an accessible mixed indication. This applies regardless of how the selection was made.
- **CAP-4**
  - **intent:** Authors can apply a chosen property value to the entire selection without disturbing other properties.
  - **success:** One explicit field commit updates every selected ID atomically and creates one undo step; failure changes none. Untouched blanks do nothing, and an edit never migrates to a later selection.
- **CAP-5**
  - **intent:** Authors can drag selected elements together to reposition their arrangement.
  - **success:** Dragging any selected component body moves every selected element by one shared translation, preserving relative offsets, dimensions, and selection. Snap preserves the arrangement; boundary handling moves the group together. One release commits atomically, and one undo restores all original positions.

## Constraints

- Full enclosure and movement use engine-projected geometry and existing canvas coordinate transforms, never text ink bounds or browser layout measurements. Zoom, band offsets, and projected page origins must be respected.
- Rectangle selection is local UI state: no geometry mutation, serialized selection, dirty revision, undo entry, or grid snapping of the rectangle. Group movement is an explicit document edit committed on release.
- Shared fields and valid choices are intersections across the selection. A shared section title alone is insufficient; unavailable controls and empty sections are omitted. See `interaction-contract.md`.
- Blank/mixed is not a stored null, clear, zero, false, or default. Selecting, focusing, and blurring untouched fields sends no property command. Clearing or setting null requires an explicit supported action.
- Use the existing bulk property command for property edits and the central encoder for all commands. Group translation requires an atomic engine operation that can preserve different starting positions; serial individual moves or identical absolute assignments are insufficient.
- A group uses one movement delta and one snapping reference. Never independently snap or clamp members. Preserve table-derived dimensions, line thickness, and existing resize behavior.
- Capture targets and geometry for each operation. Changed selections/documents invalidate unsubmitted drafts and gestures; in-flight results stay scoped to captured IDs and cannot restore stale selection. Clear incompatible table-column/editor state.

## Non-goals

- Group scaling, persistent grouping/ungrouping, lasso or partial-overlap selection, and new keyboard group-nudging behavior.
- New bulk content replacement, image replacement, binding, table structure, or orientation-remapping operations.
- A panel-wide Apply button or multi-field staging transaction; existing per-field commit timing remains.
- Moving elements into different bands, automatic edge scrolling, new touch/multitouch gestures, or changes to the document format.

## Success signal

An author draws a rectangle across displayed pages, selecting two fully enclosed text elements while excluding a partially intersecting rectangle, then Shift-drags to add another text element. Dragging any member moves all three without changing their spacing; Snap aligns the drag reference without distorting the group, and one undo restores every original position. A shared font size appears if equal and blank if different. Entering a font size changes all three; one undo restores their distinct prior values. Adding a rectangle hides typography while leaving shared controls usable. Selection and untouched fields never change the document revision.

## Assumptions

- **A-1:** Rectangle selection starts with the primary pointer on blank canvas and uses the existing 2 CSS-pixel threshold. Resize handles and armed placement retain their gestures; Escape/pointer cancellation aborts the active gesture.
- **A-2:** Properties commit on Enter/blur or explicit picker/toggle choice. Numeric assignments remain absolute per-element values, including band-relative X/Y; dragging a group instead applies a shared delta.
- **A-3:** Bulk property scope is the existing main inspector. Multi-selection keeps generic Width/Height where both are editable for every target; no semantic Length/Thickness remapping across differently oriented lines is added.
- **A-4:** Geometry, document, selection, mode, or zoom changes cancel an uncommitted canvas gesture. Unsupported scrolling also cancels rather than applying stale coordinates; no new auto-scroll behavior is required.
- **A-5:** Any fully enclosed occurrence of a repeating header/footer selects that logical ID once. A content element split across pages requires enclosure of its complete projected box; clipped or undrawn portions do not qualify alone.
- **A-6:** Moving a group preserves each element's band. Limit the shared displacement by the first member to reach a legal boundary; content retains its existing uncapped vertical column movement. Preserve document-coordinate offsets across page windows.
- **A-7:** Snap uses the pressed selected element's origin as reference and chooses one feasible shared offset. If no grid-aligned offset fits, preserve a legal common offset rather than distorting the group.
- **A-8:** An unmodified click on a selected member keeps the group selected; dragging an unselected element selects/moves only it. Shift-click remains a membership toggle and Shift-drag on blank canvas remains additive rectangle selection.
