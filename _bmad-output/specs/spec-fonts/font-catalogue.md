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

### The Story 16.1a batch — the tier grows to 31, by rule, with an owner and a re-run trigger

**Added 2026-09-03 by Story 16.1a (D-16.R.16, membership corrected by D-16.R.19).** The local face tier
was 21 families and is now **31**. The ten added are:

> **Arimo · DM Sans · Lora · Montserrat · Open Sans · Oswald · Plus Jakarta Sans · Roboto Condensed ·
> Roboto Mono · Roboto Slab**

**Why a batch at all.** Measured against the committed index snapshot, **38 of the 50 most popular
families are variable-only on the `google/fonts` mirror** and so cannot be added from the fetched tier
at all. Their **own project releases** publish perfectly ordinary single-weight statics; only the mirror
lacks them. So the head of the distribution is reachable by the mechanism this repository had already
executed 21 times — commit the upstream static, its unmodified `LICENSE*`, its `NOTICE.md` and its
manifest row — and by no other.

**PINNED TO THE SNAPSHOT COMMIT, not merely to the file** — `folio-designer/font-index.json` at
**`d6d51f16988cddf20d1a28697cd556b3d0a63f62`**. The file name alone does not pin a set: the index is regenerated by
`scripts/build-font-index.mjs`, and a regeneration landing between the survey and the build would
silently change the membership while every recorded step still read true. Worse, a later reader could
not tell whether the list was **wrong** or the **file moved underneath it** — two very different
failures with one appearance. Re-measured at implementation time rather than taken on report:
`git log -1 -- folio-designer/font-index.json` is exactly that commit, and this story's whole diff
against its baseline touches the file **not at all**.

**THE MEMBERSHIP RULE, and it is a GOAL rule rather than a margin rule.** *The refused families within
the **top 20 by `popularity` on the committed index snapshot**, minus CJK (a standing Non-goal), minus
the families the local tier already holds, minus the `shippedFamilies` collisions, minus anything with
no obtainable static from its own project upstream. Whatever that yields is the batch.* Refused means
`axes` non-empty on the snapshot row; the sort is `popularity` ascending with family name ascending as
the tie-break, so the set is reproducible offline from committed artifacts alone — from the snapshot at
`d6d51f16988cddf20d1a28697cd556b3d0a63f62` specifically, per the pin above.

**It is deliberately not "the most popular N that fit".** A rule with no goal term says yes until the
resource runs out and will consume whatever margin exists, on any run, forever (D-16.R.16). If the
goal-set ever **overflows** the slot budget, the batch **halts and returns the overflow** — it is never
truncated to fit, because truncation-by-resource silently converts a goal-bounded set back into a
margin-bounded one.

**The exclusions, itemised, because a set is only checkable against the things it left out.** **Two
different numbers are carried for every family below, because this data has two and they are not the
same one:** *position* is the rank in the snapshot's ascending sort, *popularity* is the field's own
value — 1,811 families carry only 1,100 distinct values, so ties are common and the two readings
diverge. Within the top 20 by position: `Roboto` (position 2, popularity 2) and `Inter` (position 5,
popularity 4) the tier already held · `Lato` (position 7, popularity 6), `Black Ops One` (position 10,
popularity 7), `Poppins` (position 11, popularity 7) and `Archivo Black` (position 14, popularity 12)
publish statics on the mirror and were never refused · `Noto Sans` (position 12, popularity 8) and
`IBM Plex Sans` (position 19, popularity 15) are `shippedFamilies` collisions that throw at
`build-wasm.mjs`'s family guard (D-16.R.17: two different problems sharing a symptom, both registered
rather than solved here) · **`Google Sans` (position 3, popularity 3) has no obtainable static** — `isBrandFont`, and absent
from all four licence directories · **`Jost` (position 20, popularity 16) has no obtainable
static under an admissible name** — set out in full below. **Twelve candidates, minus those two, is
ten.**

**THE "MINUS CJK" CLAUSE FIRED ZERO TIMES, and that is recorded rather than left to be inferred.** The
rule subtracts CJK families as a standing Non-goal, and a reader who sees no CJK family in the batch
cannot otherwise tell whether one was dropped by that clause or whether none was ever a candidate —
exactly the ambiguity D-16.R.16's *"exclusions itemised"* guardrail closes. **Measured on the committed
snapshot: it removed nobody, because there is no CJK family in the snapshot at all.** The index build
excludes them upstream — `publishedFamilies` **1,946**, `excludedCjkFamilies` **135**, and
`1,946 − 135 = 1,811`, the family count exactly. Probed directly: `Noto Sans JP`, `Noto Sans SC`,
`Noto Sans TC`, `Noto Serif SC` and `Noto Sans KR` are **absent**, and **zero** of the 1,811 rows
declares a Chinese, Japanese or Korean subset. **The clause is kept anyway**, because it states the
Non-goal at the place membership is decided; it is simply not what made this batch ten.

**The boundary of the cut, itemised — because a set is only checkable against what sits just outside
it.** Positions are the index snapshot's own order (`popularity` ascending, family name ascending as the
tie-break), computed from the committed `folio-designer/font-index.json` at `d6d51f1`
(`snapshotDate` 2026-09-03, 1,811 families) and from nothing live.

| position | family | popularity | why it is not in the batch |
|---|---|---|---|
| **20** | **Jost** | 16 | **Inside the cut and admitted by the rule**, then dropped at TASK 2 on evidence — see below. The last family the top-20 cut reaches. |
| 21 | Raleway | 17 | Outside the top-20 cut |
| 22 | Share Tech | 17 | Outside the cut, and static upstream — never refused, so never a candidate |
| 23 | Bebas Neue | 18 | Outside the cut, and static upstream |
| 24 | Nunito | 18 | Outside the top-20 cut |
| 25 | Fraunces | 19 | Outside the top-20 cut |
| 26 | Heebo | 19 | Outside the top-20 cut |

**`Raleway`, `Nunito` and `Fraunces` are named here deliberately.** They were the three families
D-16.R.19 removed when it recomputed the set from the committed snapshot, and without this row they
would appear in neither the batch nor the exclusions — vanishing silently, which is the exact failure
D-16.R.16's *"exclusions itemised"* guardrail exists to prevent. They are out because they sit **outside
the top-20 cut**, not on merit and not on the ceiling.

**`Jost` (position 20) is the one member the rule admitted and TASK 2 dropped.** It is inside the cut,
refused on the mirror, not local, not a `shippedFamilies` name — admissible on every stated criterion —
and it fails the last one: *no obtainable static from its own project upstream.*
`indestructible-type/Jost` publishes `fonts/ttf/Jost-400-Book.ttf` at release `3.5` and at HEAD; both
call the family **`Jost*`** in nameID 1, with nameID 2 `Regular`. The asterisk is outside
`build-wasm.mjs`'s `familyShape`, because that string is interpolated unescaped into
`font-family: '<name>'`; and `src/font-catalogue.test.ts` ties the manifest's `family` to nameID 1
byte-for-byte, on the stated ground that *"a family name is an assertion about bytes"*. Declaring it as
`Jost` — the spelling the index uses and the only one that would join — would publish a family name the
face's own bytes contradict. **No admissible static exists, so it is out, and nothing was promoted in
its place.**

**The size, measured, not chosen:** ten faces into **twenty** free precache slots
(`s1.assetCount` 44 → **54**, `maximumCacheAssets` 64), leaving a margin of **10**. The cap was not
moved and must not be moved by this epic under any circumstance.

**OWNER: the engineering lead**, as the standing holder of this epic's font-tier decisions — the same
authority that declined to supply an N and replaced the criterion instead (D-16.R.16). A change to the
**criterion** is an owner decision and is escalated; a **re-run under the existing criterion** is the
lead's. *(This appointment was made by the story rather than received from the gate, and is recorded
that way so it can be overridden by naming someone else here.)*

**RE-RUN TRIGGER — any one of these, and the batch is recomputed, not extended by hand:**

1. **The index snapshot is regenerated** (`scripts/build-font-index.mjs`) and the top-20-by-popularity
   membership moves. Popularity is a distribution that shifts; this batch is a snapshot of it, and a
   batch nobody re-runs is a queue (**DW-166**).
2. **Story 15.0 lands.** D-8.4d.1's fetch-on-first-pick is policy with no implementation today, so these
   faces are **precached**; when it ships, the slot constraint dissolves and the goal-set may be
   re-derived without it.
3. **A previously unobtainable family becomes obtainable** — `Google Sans` published under one of the
   four identifiers, or `Jost` shipping a static whose `name` table says `Jost`.

**And what a re-run may NOT do:** raise `maximumCacheAssets`, take bytes from the `google/fonts` mirror,
add a variable face, hand-copy a licence text, or admit a face whose `cmap` disagrees with its declared
`scripts` in either direction. Those are the same Block Ifs the first run was held to.

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

## The machine store — a fetched face stays on this machine

**Added 2026-09-03 by Story 16.2, under D-16.2 and D-16.R.33 R1.**

A face fetched from the web tier is **kept**, in an **origin-scoped IndexedDB store** keyed by the
**SHA-256 of the face bytes** — the same content address `.folio`'s `assets` map uses and the same
one Go derives at `embedFontFamily`. The store's read goes **in front of** the fetch, so picking a
family this designer has fetched before sends **no request at all**, and therefore works with the
network down. A stored family joins the family control's offered list as a third `FamilySource`
tier, `'stored'`, between the local tier and the web tier.

**What travels into the store, and why it is not negotiable.** Bytes, media type, family, style,
licence identifier, licence text, copyright, `source`, scripts and the date fetched. A face offered
from the store must be embeddable **without a network**, and `embedFontFamily` refuses without the
licence identifier, the licence text and the copyright — so a store that kept the bytes and dropped
the terms would put a document its own parser refuses one step away.

**Three things it deliberately is NOT.**

- **Not a list of the fonts installed on your computer.** The designer never enumerates, reads or
  feature-detects a host-installed face: SPEC-fonts' *"No host fonts"* Non-goal is the one clause
  D-16.1 left standing, and it is untouched. The dropdown group called **AVAILABLE LOCALLY** means
  *"typefaces this designer has fetched before, plus the ones it ships with"*. A source scan over
  the whole designer (`scripts/host-font-access.mjs`) fails on any spelling of the Local Font Access
  API, and it is red-proved by deleting the guard rather than by falsifying a condition.
- **Not a second copy of a document's fonts.** A `.folio` carries its own faces (CAP-2). The store
  shortens a fetch; it never stands in for what a file contains, and **removing an entry never
  changes a saved document**. It is a cache and a source, never an authority — which is why the
  `source` provenance string is carried back **byte-identically** and asserted, at retrieval,
  through the same tripwire both writers are held to.
- **Not a render-time input.** FR33 is untouched: nothing is fetched or read from the machine at
  render. This store serves **authoring** only, and it consumes no release cache slot —
  `maximumCacheAssets` does not move.

**Why IndexedDB and not `localStorage`, on arithmetic.** `localStorage` is a ~5 MB per-origin
**string** quota, so bytes must be base64 at **+33%**: a measured `Sarabun-Regular.ttf` of 90,220
bytes costs ~120 KB stored, and the largest face this designer can legitimately offer — `Noto Color
Emoji`, 24,271,604 bytes — is ~32 MB base64 and does not fit at all in an empty origin. Its
`QuotaExceededError` is thrown synchronously from the assignment with no partial-write path, so a
store built on it cannot degrade; it can only fail. The reasoning is written into
`src/font-store.ts` so it is not "simplified" back.

**Why the content hash and never the family name.** The store answers *"do I already have these
bytes"*. Keying by name would make it answer *"do I have something called Sarabun"* — and the day
upstream changes the face, the store would hand over the old bytes under the new name and the
document would carry a face nobody chose. Under AD-8 a family that changed upstream is a **different
key**, not a newer one.

**What "on this machine" honestly means.** This browser, this profile, this origin. Not the
operating system, not synced, not shared with another browser on the same machine. The UI does not
imply more than that.

**Every failure is a stated degradation, never a failed pick.**

| What happened | What the author gets |
|---|---|
| The store cannot be opened (private window, cleared site data, storage blocked) | An empty group, a working designer, and one message — stated **once**, not per pick |
| A write is refused (quota) | The fetch and the embed still succeed; the message names the family and says the **caching** is what failed |
| A stored entry is unreadable, has lost its bytes, or no longer hashes to its own key | Treated as absent, dropped, and refetched on the next pick — self-healing, and logged honestly |
| A family is neither stored nor reachable | *"You cannot add a family right now"* — never a document that will not render |
| A fetch **stalls** rather than failing | After 30 seconds the attempt is stopped, the hold on the control is released, and the degradation is stated in its **own** words — never the offline wording, which would be false while the network is up. Nothing is retried automatically. |

**The removal affordance.** The panel lists what this machine holds — family, style, size, and the
day it was downloaded — with one control per row that names what it removes. Removing is a
**machine** action: no command is sent, no revision moves, no undo entry is pushed, and the panel
says so, because an author who has to guess whether a delete button reaches their saved work will
never press it.

**Not yet proven in a real browser.** Every claim above is asserted against an independent IndexedDB
implementation in the test environment, because jsdom provides none. The browser witness — *a stored
face survives a reload and is offered with the network disabled* — is routed to Story 16.3's browser
run as a fourth case beside DW-161's three.

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

**Since Story 16.2 the offered rows come from three tiers, and each row says which.** The committed
local faces read *"already on this machine"*; a face this designer has fetched before reads
*"already downloaded to this machine"*; a family from the build-time snapshot carries the plain
note, because picking it fetches. Story 16.4 groups these under headings; 16.2 builds the seam the
grouping hangs on, as a third arm of the `FamilySource` union that every exhaustive switch must
handle.
