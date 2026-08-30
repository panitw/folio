# The designer's font catalogue

The catalogue backing CAP-3. It is a **shipped artifact of the designer**, not a service the
designer talks to.

## Why bundled rather than fetched

The designer ships an offline release behind a service worker with verified asset URLs, and the
product's claim is that it opens with no account, no upload and no server. A live
`fonts.google.com` list would make the font control the one part of authoring that stops working
on a plane, and would put a third-party request in the middle of a flow whose selling point is
that no template or data byte leaves the machine. The catalogue therefore ships with the app, at a
known size, verified like every other release asset.

The cost is honest and stated: the list changes only when the designer is released.

## Selection criteria

| Criterion | Rule |
|---|---|
| Licence | OFL-1.1 (or equivalently redistribution-permitting). The licence identifier and text ship with the face and are recorded in every document that embeds it. |
| Instance | Static, single-instance faces derived ahead of the build, exactly as the shipped set is (`tools/fontgen`, committed outputs, replayable derivation, NOTICE per face). A face generated at build time makes the PDF a function of the build environment. |
| Script coverage | Latin and Thai scale. CJK families are excluded from the embeddable catalogue in this scope — a full SC face is 10.6 MB against 646 KB (Noto Sans) and 47 KB (Noto Sans Thai). |
| Size budget | Each catalogue face is counted against the offline bundle budget, and against the weight it will add to every document that embeds it. |

Which families, and how many, is an open question in `SPEC.md`.

## Per-entry record

Each catalogue entry carries what the designer shows and what the document will record:

- `family` — display name in the picker.
- `style` — the instance shipped (`Regular` initially; see the bold/italic open question).
- `licence` — identifier plus the text that travels into the document.
- `source` — upstream release and the derivation that produced the shipped instance.
- `bytes` — the face itself, in the offline bundle.
- `scripts` — what it covers, so the designer can propose a sensible fallback tail.

## Picking a family

1. The author opens the family control and searches the catalogue. Catalogue entries are shown
   alongside the chains the document already declares; the two are visibly different things — a
   chain is in the document, a catalogue entry is not yet.
2. Picking a catalogue entry embeds the face as an asset (content-addressed, so picking the same
   family twice stores one copy) and adds a chain naming it.
3. The proposed chain's tail is the shipped fallback for coverage the picked face lacks — a Latin
   face followed by the shipped Thai and SC faces — and the author can edit that tail (CAP-1).
4. Nothing is fetched. The bytes come from the bundle already on the machine.

## Removing a family

Removing the last chain entry that references a font asset leaves the asset unreferenced. The
document is the author's, so the rule is stated rather than inferred: unreferenced font assets are
dropped on save, the same way an unreferenced image asset is treated. This keeps a file from
accumulating megabytes of faces nothing draws with.

## What the panel shows

The typography panel's family control (already a search-and-select over the engine-projected list
of declared chains) gains the catalogue as a second, clearly separated group. The engine remains
the authority on what `fontFamily` may name: the designer never invents a family name, and a
catalogue pick becomes a chain through a command like every other document change.
