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

**AMENDED 2026-09-02 by OWNER DECISION (D-16.1), and this section is now half true.** The paragraph
above is preserved verbatim because its reasoning is what the decision was taken against. **What
changed:** the catalogue is no longer the only source. A family's **face bytes, its upstream licence
text and its per-family metadata are fetched from a third party at the moment the author picks it** —
`raw.githubusercontent.com/google/fonts`, which serves a full static TTF with `access-control-allow-
origin: *`, never `fonts.googleapis.com/css2`, which serves `woff2` this engine refuses by design and
`unicode-range` subsets that would embed partial coverage. **What did NOT change, and it is the
sentence above that survives:** the **list** still ships with the app. `fonts.google.com/metadata/fonts`
sends no CORS header, so a browser cannot read it; the index is snapshotted at build time and *"the
list changes only when the designer is released"* remains exactly as true, and exactly as costly, as
it was for the 21 bundled families. Anyone describing the result as a live font browser without that
qualification is describing something this product does not do. See D-16.1, D-16.3 and Epic 16.

## Selection criteria

| Criterion | Rule |
|---|---|
| Licence | One of the four identifiers D-8.5.3 admits — **OFL-1.1, Apache-2.0, MIT, UFL** — enforced the way AD-26's dependency ban is: fail the build, never warn, with an unclassifiable licence treated as failure (D-8.5.2). *(This row read "OFL-1.1 only" until Story 8.6 corrected it; the allowlist has been four identifiers since 8.5, and two shipped catalogue faces are `Ubuntu-font-1.0`.)* The licence identifier **and its text** ship with the face and are recorded in every document that embeds it — since Story 8.6 they are **required** there, and so is the copyright line, or the document is refused at load. |
| Instance | Static, single-instance faces derived ahead of the build, exactly as the shipped set is (`tools/fontgen`, committed outputs, replayable derivation, NOTICE per face). A face generated at build time makes the PDF a function of the build environment. **Extended 2026-09-02 by D-16.5 to the fetched path:** a family that ships **variable-only** upstream — 558 of 1,946, 28.7%, measured — has no static Regular to fetch and is **refused, visibly and with its reason**; the ones worth having are put through `tools/fontgen` ahead of the build like the shipped set, which is why `Noto Sans Thai` and `Noto Serif Thai` are offered despite being variable upstream. **Instancing in the browser is refused**: it makes the embedded face a function of the *author's runtime*, which is the same hazard one step further out, and folio has no backend to do it anywhere else. A row is refused because no static face is obtainable, never because upstream happens to be variable. **CORRECTED 2026-09-03 by Story 16.1, under D-16.R.2 and D-16.R.2a; the sentence above is preserved verbatim and two of its clauses are now false.** (1) **Variable-only families are FILTERED OUT, not refused.** Measured: **37 of the 50 most popular families are variable-only** — Roboto, Open Sans, Inter, Montserrat, Raleway, Nunito, Oswald, Playfair Display — so listing them and refusing means the most common first action in the product fails. A row the author cannot act on is a row that should not be there, and the count the browser reports is the count of families it can actually add, and says so. **A hidden row is a presentation choice, never a guard:** the engine's refusal stays for anything that reaches it. (2) **The derivation premise was false.** `tools/fontgen/instance_faces.py` drives a hardcoded three-entry list of **engine** faces; **none of the 21 catalogue faces is derived** (every `NOTICE.md` reads *"NO DERIVATION APPLIES"*), and of the two Thai families cited, only `Noto Sans Thai` is a derivation. (3) **"Variable-only" is a property of the BYTE SOURCE, not of the family.** `Roboto` and `Inter` are committed here as byte-for-byte upstream *statics*, and appear among the 558 only because the `google/fonts` mirror carries VF-only builds of them — so a family the local tier holds is offered from it and the mirror's `axes` field is never consulted for it. **Instancing in the browser stays refused**, unchanged. |
| Script coverage | Latin and Thai scale. CJK families are excluded from the embeddable catalogue in this scope — a full SC face is 10.6 MB against 646 KB (Noto Sans) and 47 KB (Noto Sans Thai). |
| Size budget | Each catalogue face is counted against the offline bundle budget, and against the weight it will add to every document that embeds it. |

Which families, and how many, was an open question in `SPEC.md` and is **settled** there
(D-8.5.3, owner decision): 20+ families, admitted by the named permissive allowlist above.

## The local face tier, and the tier beside it

**Added 2026-09-03 by Story 16.1 (D-16.R.3).** The 21 families this file describes did not go away when
D-16.1 opened the published library. They survive **unchanged**, as the **local face tier** — this
manifest, the committed binaries, the `LICENSE*` and `NOTICE.md` beside each one, and the build-time
gate over all of it. `pickCatalogueFamily` **gained a source; it did not swap one.**

*(The term is the **local face tier**, not the "derived-static tier". None of the 21 is derived — see
the correction in the Instance row above.)*

- **Join key: exact `family` string equality.** No case-folding, no whitespace normalisation, no fuzzy
  match. Measured: `Inter Display` and `Source Serif 4 Display` have **no upstream index row at all**,
  so 2 of the 21 are unjoinable under any normalisation — while `Geist` / `Geist Mono` / `Geist Pixel`
  is exactly the neighbourhood a loose matcher gets wrong. **A local face with no index row is
  local-tier-only, and that is correct behaviour, not a defect.**
- **Local wins, with no third-party fetch at all** — no `METADATA.pb`, no upstream licence file. The
  committed bytes carry a **stronger** record than any fetch can produce; preferring a fetch would
  replace a verified record with an unverified one. In the browser a local-tier row says so, and the
  pick works offline.
- **Divergence is deliberately not reconciled.** Under AD-8 and D-16.2 a face is identified by the
  SHA-256 of its bytes, so *"upstream released a newer version"* is a **different face**, not a newer
  one. No staleness check, no update prompt, no version compare — registered as **DW-158** with its
  trigger.

**And what the tier beside it does NOT ship.** The family **list** is a build-time snapshot:
`fonts.google.com/metadata/fonts` sends no `access-control-allow-origin`, so a browser cannot read it,
and *"the list changes only when the designer is released"* is as true of the ~1,800 snapshot families
as it was of these 21. Only the **face, its licence text and its per-family metadata** are live. The
snapshot carries **no licence field**, deliberately: a licence token cached beside the list would be a
second authority on the terms, ageing on its own schedule. **No licence is knowable before a pick.**

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
