# Moving a selection together

CAP-5 adds drag movement for an existing selection, whether formed by a rectangle, Shift-click, or both. This supersedes the earlier group-movement non-goal. No persistent group object is created. SPEC.md assumptions A-6 through A-8 identify band, snapping-reference, and click choices.

## Gesture and selection ownership

- An unmodified primary-pointer press on any already-selected component body captures the full selection and each member's committed bounds. Do not replace the selection with the pressed member. A selected table body follows this rule even over a column cell; table-column selection must not steal a multi-selection drag.
- Begin movement after the existing 2 CSS-pixel threshold. An unmodified press/release on a selected member without movement retains the group and sends no command. Dragging an unselected element instead selects and moves that element alone; Shift-click continues to toggle selection membership.
- Explicit resize handles, band boundaries, and editor controls retain their own actions. Group movement does not create group resize handles or convert line endpoint resizing into movement. Armed placement retains precedence.
- Capture IDs, start bounds, document/geometry context, snap setting, pointer position, and pressed occurrence. Translate each original position by the same delta throughout the preview. Never accumulate rounded frame deltas or rebase only some members on a new snapshot.
- All selected logical elements preview together, including echoes. A repeated header/footer ID moves only once; every occurrence reflects that result. Use the existing projected sheet mapping to translate the pointer path into document/column displacement across page seams. Page gaps must not introduce per-member drift.
- Release commits once and keeps the selected set intact. Suppress subsequent clicks that would collapse/clear it. During the drag, position fields reflect preview values under shared/mixed rules and cannot submit competing geometry edits.

## Translation, snapping, and boundaries

For every captured member i, the accepted move is `x'_i = x_i + dx` and `y'_i = y_i + dy` in its original band's coordinate system. Width, height, line thickness, style, content, and band identity do not change. Relative coordinate differences within the selection remain unchanged; screen positions continue to follow projected page layout.

With Snap off, preserve precise movement at the document's millipoint resolution. With Snap on, use the pressed element's origin as reference, align its proposed origin to the grid, and derive one shared delta. Apply that delta unchanged to every member. Members initially off-grid need not all land on grid: independently snapping them would change their spacing. The final preview and engine result must use the same accepted shared delta.

Intersect every member's legal displacement range to find the offsets that fit the whole selection. Band limits retain existing semantics: every band caps horizontally, header/footer cap vertically, and content is a vertically flowing column. At a boundary the entire group stops together; no member is independently clamped, resized, left behind, or reassigned to another band.

Choose a snapped reference position within the common feasible range. If that range has no grid-aligned reference position, retain the legal unsnapped common offset, following the existing rule that snapping must not invalidate a fitting placement. Do not use the fallback when a feasible grid-aligned offset exists.

## Transaction and cancellation

Use one engine-validated atomic move operation for all captured IDs. It may be an additive relative-movement command or an equivalent atomic per-target transaction; choose the concrete protocol during implementation. Existing `updateComponentProperties(ids, {x, y})` applies identical absolute values and cannot implement relative group motion for differently positioned elements. Sequential `moveComponent` calls are also insufficient: they can partially succeed, distort spacing through individual snapping, and create multiple undo entries.

The engine owns legality, final snapping/containment, and all-or-none installation. One successful release creates one history entry. One undo restores all original positions; redo reapplies the accepted group translation. A zero accepted displacement, click, or cancelled drag creates no document edit or history entry.

If any captured ID is missing or the command is otherwise invalid, reject the whole move, return the preview to committed positions, and show a located error where available. Do not fall back to moving only valid members. Retain the selection of IDs still present.

Escape, pointer cancellation/lost capture, document replacement, changed selection, external geometry change, zoom/mode change, or invalidated scroll coordinates cancels an uncommitted move. Delayed release and in-flight results cannot target a newer selection, reselect an old group, or overwrite a newer document. After dispatch, the captured transaction follows the existing command lifecycle; UI cancellation must not pretend an already committed move was undone.

## Acceptance cases

| ID | Given / When | Then |
|---|---|---|
| G-1 | Given three selected elements, when dragging any selected member | All three preview/move together; selection survives pointerdown, release, and following click. An unselected fourth stays unchanged. |
| G-2 | Given mixed text, rectangle, image, line, and table elements, when translating | Every member receives exactly one identical delta; relative offsets, dimensions, thickness, styles, content, and bands remain unchanged. |
| G-3 | Given off-grid members with unequal offsets, when moving with Snap on/off | On aligns the pressed reference using one common delta; off keeps precise delta. Neither changes relative spacing. |
| G-4 | Given members with different remaining band space, when dragging past an edge | The common movement stops at the first limit. If no snapped offset fits, the legal common fallback preserves all members and their spacing. |
| G-5 | Given preview movement, when cancelling or returning to zero displacement before release | All committed positions and selection remain; no revision/history entry. |
| G-6 | Given a successful move or a command invalid for one member | Success is one revision/history entry, undo restores all, redo reapplies all; failure changes none and reports the refusal. |
| G-7 | Given a selected table body, a line resize handle, or an unselected element | Table-body drag moves the selection without selecting a column; resize stays resize; the unselected-element drag moves only that element. |
| G-8 | Given repeated header/footer IDs or content across page seams at different zooms | Move each ID once; all occurrences reflect it; the same document delta preserves offsets without page-gap drift. |
| G-9 | Given an active/pending group move, when selection, geometry, or document changes | Cancel uncommitted work; captured commits never target the new group or restore stale selection. |
| G-10 | Given a multi-selection, when typing an absolute X/Y versus dragging the group | Property edit assigns the same value; group drag preserves different starting positions. Both remain atomic with their own single undo step. |

Use pure translation/snap/feasibility tests, real engine command and host undo/redo tests, and browser pointer tests. Include fractional offsets, a boundary with no feasible grid point, cancellation before/after the threshold, table-body hit targets, and the release-generated click.
