# Selection and bulk editing contract

This companion makes CAP-1 through CAP-4 testable; `group-movement.md` defines CAP-5. User-confirmed choices are replacement versus Shift-add, shared values versus blank mixed values, and selection across displayed pages and bands. SPEC.md identifies inferred interaction choices.

## Rectangle gesture

1. In Design mode, press the primary pointer on blank canvas. Capture the starting selection and current geometry/document context. An element body, resize handle, band boundary, editor, or toolbar is not a rectangle start; armed palette placement takes priority.
2. After the existing 2 CSS-pixel threshold, display a noninteractive rectangle from start to current point. Normalize its edges so all four directions behave identically. Suppress native text/image dragging for this gesture.
3. The rectangle may cross page gaps and bands; gaps contain no selectable elements. Compare in one canvas coordinate system using existing zoom and sheet transforms. Snap and Grid visibility do not affect membership.
4. On pointer release, calculate the enclosed set and install it once. Normal drag replaces selection; Shift held at gesture start unions the enclosed set with the starting selection. Do not let the release-generated click clear the result. Remove the overlay.
5. Escape, cancellation/lost capture, leaving Design mode, or invalidated coordinates aborts the gesture and retains the starting selection, intersected with IDs still present. A blank click below threshold retains current click-to-clear behavior; Shift-click on blank canvas leaves selection unchanged.

Selected IDs are unique and ordered deterministically by the current projection. The inspector count reflects logical elements, not drawn occurrences. Rectangle selection clears table-column selection and follows existing table-editor revocation and selection-error cleanup. Pressing an already-selected component body follows group-movement precedence instead of starting a rectangle or collapsing the selection.

## Full enclosure and occurrences

For element box B and rectangle R, enclosure includes equality on every boundary: `R.left <= B.left`, `R.top <= B.top`, `R.right >= B.right`, `R.bottom >= B.bottom`. Merely crossing a shape, covering its center, or covering only text ink is insufficient. Thin lines use their full thickness box; all fully enclosed overlapping elements qualify regardless of which is on top.

Use engine-projected x/y/width/height and the same sheet mapping as painting. Do not infer dimensions from labels, handles, shadows, clipped DOM boxes, or measured text.

Repeated headers/footers have multiple complete occurrences but one ID. Enclosing any complete occurrence selects that ID once. Content spanning page windows is one element: every projected portion of its complete box must lie inside the rectangle. A single clipped continuation does not qualify. Elements beyond the drawn-sheet cap or with an undrawn portion are not fully enclosed by a visible rectangle. Highlight logical selection without creating duplicate focusable occurrences.

## Shared property rules

Only sections with at least one common editable field appear in multi-selection. A control is common when its field, meaning, and offered operation/choice are supported by every target under the current authored-property policy. Do not expand capabilities merely because the engine can retain or serialize a field.

| Section / control | Common selection requirement | Bulk behavior |
|---|---|---|
| POSITION: X, Y | All five current element kinds | Set the same band-relative coordinate on every target. These fields assign absolute values; group dragging applies a relative delta. |
| POSITION: Width, Height | No selected table | Retain generic multi-selection dimensions. Tables keep derived size; line drag restrictions remain independent of explicit property edits. |
| TYPOGRAPHY | Every target is text or table | Offer shared font, size, spacing, bold, italic, text colour, and alignment. Justification requires support on every target; retain font availability restrictions. |
| BOX: background / colour / fill | All five current kinds | Use the common background property with a clear bulk label. Current absence on some elements does not make a supported field unavailable. |
| BOX: visibility | All five current kinds | Apply an explicit common value or supported clear action; a mixed blank does not make every element visible. |
| BOX: border width, colour, edges | No selected line | Follow the authored vocabulary: a selected line withholds border controls even though stored borders can exist. |
| CONTENT, IMAGE, TABLE configuration, BINDING | Single-element workflows | Omit in multi-selection; no empty sections or bulk controls guaranteed to be refused. |

The matrix covers existing main-inspector controls, not every engine property. Returning to one selected element restores its normal kind-specific panel. During group movement, shared/mixed position display reflects the common preview delta; position editing is unavailable until the gesture ends.

## Values, commits, and async behavior

- Compare committed values without collapsing distinct absent, null, false, zero, and empty states. If all targets agree, show the shared value using existing display rules; otherwise show blank with a visible and accessible `Mixed` indication.
- Mixed booleans/segments are indeterminate or unchosen; a swatch must not imply an arbitrary fallback colour is shared. Users can explicitly choose a supported value.
- Untouched controls never commit, including inherited defaults. Returning a typed draft to its initial blank without choosing Clear/Null leaves the document unchanged. Explicit Clear and Null remain distinct, offered only where legal for every target.
- A field commit captures selected IDs and submits one `updateComponentProperties` command containing only the intended field/operation. Do not send all visible fields or one independent command per element.
- Success updates field display from the engine. One undo restores every target's previous value, including divergence and absence/null. Unselected elements and other fields remain unchanged.
- If any target rejects the value, none changes. Show an error by the relevant control and the failing element where supplied; retain the draft for correction. Prevent duplicate submission while pending.
- A new selection gets fresh drafts. Unsubmitted values, timers, stale validation, or delayed blur cannot apply to it. In-flight commands retain captured targets and cannot reselect the old group.

## Acceptance cases

| ID | Given / When | Then |
|---|---|---|
| AC-1 | Given enclosed A/B and partially crossed C, when dragging in each of four directions | Exactly A/B are selected; touching edges qualify. |
| AC-2 | Given a thin line, image, table, text, and rectangle, when full boxes are enclosed | Every kind is eligible, including overlaps; text ink alone is insufficient. |
| AC-3 | Given a selection, when completing normal, Shift-add, empty, or cancelled drags | Replace, union, clear on normal empty drag, or preserve on cancellation; Shift-empty preserves. |
| AC-4 | Given multiple pages/bands, when crossing gaps, repeated headers, and split content | All qualifying logical IDs are selected once; partial continuations are excluded. |
| AC-5 | Given the same geometry at 50%, 100%, and 150% zoom, with Snap/Grid on or off | Equivalent rectangles select the same IDs. |
| AC-6 | Given a selected body, unselected body, resize/line endpoint, band handle, or armed placement | Selected-body drags move the group; other targets retain their specified behavior. No competing rectangle appears. |
| AC-7 | Given a completed rectangle, when the browser dispatches the subsequent click | Selection remains and the overlay is gone. |
| AC-8 | Given text/text, text/table, text/rectangle, line/rectangle, and table/rectangle selections | Fields and choices follow the intersection matrix; no empty single-item sections appear. |
| AC-9 | Given identical or differing numeric, colour, boolean, enum, absent/null states | Shared values display correctly; differences are blank/indeterminate and accessible as mixed. |
| AC-10 | Given a multi-selection, when focusing/blurring untouched shared, mixed, defaulted fields | No command, dirty revision, or history entry. |
| AC-11 | Given a mixed selection, when setting a common field | Every target receives it once, others remain unchanged, and one undo restores prior values. |
| AC-12 | Given a value invalid for one target, when committing to the selection | No partial mutation; identify the failing target where available. |
| AC-13 | Given a dirty draft or pending commit, when changing selection/document | Async work never targets the new group or restores stale selection. |
| AC-14 | Given an active rectangle, when Escape, cancellation, capture loss, geometry/zoom/mode change occurs | Abort without document mutation or stale-coordinate selection. |

Verify pure enclosure/intersection arithmetic, UI selection and drafts, real-engine atomicity/undo, and actual browser capture, release-click ordering, scrolling, zoom, and page-spanning selection. jsdom alone cannot demonstrate canvas gestures.
