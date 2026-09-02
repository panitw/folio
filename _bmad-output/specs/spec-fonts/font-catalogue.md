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
| Licence | One of the four identifiers D-8.5.3 admits — **OFL-1.1, Apache-2.0, MIT, UFL** — enforced the way AD-26's dependency ban is: fail the build, never warn, with an unclassifiable licence treated as failure (D-8.5.2). *(This row read "OFL-1.1 only" until Story 8.6 corrected it; the allowlist has been four identifiers since 8.5, and two shipped catalogue faces are `Ubuntu-font-1.0`.)* The licence identifier **and its text** ship with the face and are recorded in every document that embeds it — since Story 8.6 they are **required** there, and so is the copyright line, or the document is refused at load. |
| Instance | Static, single-instance faces derived ahead of the build, exactly as the shipped set is (`tools/fontgen`, committed outputs, replayable derivation, NOTICE per face). A face generated at build time makes the PDF a function of the build environment. |
| Script coverage | Latin and Thai scale. CJK families are excluded from the embeddable catalogue in this scope — a full SC face is 10.6 MB against 646 KB (Noto Sans) and 47 KB (Noto Sans Thai). |
| Size budget | Each catalogue face is counted against the offline bundle budget, and against the weight it will add to every document that embeds it. |

Which families, and how many, was an open question in `SPEC.md` and is **settled** there
(D-8.5.3, owner decision): 20+ families, admitted by the named permissive allowlist above.

## Per-entry record

Each catalogue entry carries what the designer shows and what the document will record:

- `family` — display name in the picker.
- `style` — the instance shipped (`Regular` initially; see the bold/italic open question).
- `licence` — identifier plus the text that travels into the document. Both reach the document:
  the identifier as `font.licence`, the text verbatim as `font.licenceText`. The text is read from
  the unmodified upstream `LICENSE*` committed beside the binary, never hand-copied into the
  manifest — a hand-copy would be a second authority on the terms, and the first binary swap would
  make the document publish terms its own bytes contradict.
- `copyright` — read from the face's own `name` table (nameID 0), the one statement of provenance
  that cannot be edited from outside the binary. Recorded as `font.copyright`, and **required** of
  an embedded face exactly as `licenceText` is.
- `source` — upstream release and the derivation that produced the shipped instance.
- `bytes` — the face itself, in the offline bundle.
- `scripts` — what it covers, so the designer can propose a sensible fallback tail. A **closed**
  vocabulary — `latin`, `thai`, `cjk` — because an unrecognised script proposes no fallback for
  itself and the chain silently draws tofu. Each face's declaration is checked against that face's
  own `cmap`, in both directions: a claimed script the binary cannot draw gets no fallback and
  renders tofu; an unclaimed script it can draw staples a redundant shipped face onto every chain
  that picks it.

## Picking a family

1. The author opens the family control and searches the catalogue. Catalogue entries are shown
   alongside the chains the document already declares; the two are visibly different things — a
   chain is in the document, a catalogue entry is not yet.
2. Picking a catalogue entry embeds the face as an asset (content-addressed, so picking the same
   family twice stores one copy) and adds a chain naming it. **The chain names it by ASSET KEY** —
   the entry is `{"asset": "<the lowercase hex SHA-256 of the face's bytes>"}`, never by family
   name, because a face is resolved by key alone and `font.family` is display identity (AD-8,
   `folio-format.md`). The chain is **named after the family**, and picking a family whose name a
   chain already takes is refused with that reason rather than silently merged.
3. The proposed chain's tail is the shipped fallback for the scripts the picked face's declared
   `scripts` do **not** cover, in the order `latin`, `thai`, `cjk` — `Noto Sans`, `Noto Sans Thai`,
   `Noto Sans SC` — and it is a **proposal**: the author edits it in the chain editor afterwards
   like any other chain (CAP-1), and nothing about it is privileged once it is in the document. A
   face covering every script the document renders proposes an empty tail, which is a chain of one
   embedded entry and is legal.
4. ~~Nothing is fetched. The bytes come from the bundle already on the machine — the face is one of
   the release's own content-addressed, precached assets, read exactly as `runtimeAssetUrls` assets
   are, so the pick works with the browser offline.~~
   **AMENDED 2026-09-02 by OWNER DECISION (D-8.4d.1), which reverses D-8.5.12; the original is
   preserved above verbatim.** A catalogue face is **fetched from the release's own origin on first
   pick** rather than precached at first load. It is still the release's own content-addressed
   asset and still no third party — what changed is **when** the bytes arrive, not where from.
   **Consequence, stated plainly because this file previously promised the opposite:** a family
   whose bytes have not yet been fetched **cannot be picked while the browser is offline**. It
   degrades to *"you cannot pick that family right now"*, **never to a document that will not
   render** — the three shipped Noto faces remain the coverage, and a face already embedded in a
   `.folio` travels inside the file. **Mechanism, caching and the offline path are the implementing
   story's design; this record states the policy change only.**
5. **Picking a family the document already carries stores nothing and declares nothing.** The
   content hash decides: the existing chain is what the author is offered, no second copy is
   stored, no second chain is declared, and — because the canonical bytes do not move — no history
   entry is pushed, so undo does not have to step over a pick that changed nothing.
6. **The whole pick is ONE command, one history entry, one undo.** The asset and the chain arrive
   together and leave together; there is no state in which the document carries bytes nothing names
   or names a key it does not carry.
7. **A typeface on the author's own disk cannot be embedded, and the control says so** (D-8.6.1).
   It is declined explicitly rather than by an absent button, because a control that is simply
   missing reads as an omission. See `SPEC.md`'s settled open question for why: a disk file arrives
   with no licence identifier, no licence text and no copyright the designer can record, and a
   `2.0` document embedding a face without them is refused at load.

## Removing a family

Removing the last chain entry that names a font asset leaves the asset unreferenced, and **the
command that un-named it drops it** — scoped to exactly that key, in that one action, so the author
sees the file get smaller because of something they did.

**It is a command, not a save-time side effect, and the distinction is not stylistic.** An earlier
draft of this document said unreferenced font assets are "dropped on save"; that is not
implementable and would be wrong if it were. The serializer's fixed point is
`Parse(Serialize(d)) == d` (AD-9, D-1.4.3's P1), so a save that collected an orphan would not be a
fixed point — and a document may legally carry an asset nothing references (D-1.4.13), which a
sweeping save would destroy without being asked. Collecting orphans is a designer feature. The
serializer preserves, unconditionally.

The scope matters as much as the timing: the drop asks only about the keys the author's action just
un-named, and only removes one when nothing else in the document still names it. A face a **second**
chain still names is kept. This keeps a file from accumulating megabytes of faces nothing draws
with, without ever removing one the author did not just let go of.

## What the panel shows

The typography panel's family control (already a search-and-select over the engine-projected list
of declared chains) gains the catalogue as a second, clearly separated group. The engine remains
the authority on what `fontFamily` may name: the designer never invents a family name, and a
catalogue pick becomes a chain through a command like every other document change.
