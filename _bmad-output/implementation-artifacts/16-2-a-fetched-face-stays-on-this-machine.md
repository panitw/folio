---
title: 'Story 16.2: A fetched face stays on this machine'
type: 'feature'
created: '2026-09-02'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: 'a40c34db6cff7372363b2a553710eff48759bef1'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-8-15-decision-log.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

A typeface fetched once is kept. The next statement you build on this machine offers it immediately,
with no download and no wait — and offers it even with the network down, because the bytes are already
here.

Two things this deliberately is not. It is not a list of the fonts installed on your computer: the
designer never looks at those, and the group in the font menu called *available locally* means
"typefaces this designer has fetched before", not "typefaces Windows has". And it is not a second copy
of your document's fonts — a template still carries its own typefaces inside it, so a file you send to
someone else is unaffected by anything stored here.

<intent-contract>

## Intent

**Problem:** D-16.1 makes every new family a network fetch. Without a store, the same family is
re-fetched for every document, the designer is useless offline for anything it has already downloaded,
and the design's third dropdown group (`AVAILABLE LOCALLY`) has nothing behind it.

**Approach:** An origin-scoped **IndexedDB** store, keyed by the **SHA-256 of the face bytes** — the
same content address `.folio`'s `assets` map uses — holding bytes, licence text, copyright and the
family metadata needed to offer the face again. Reads populate the dropdown group; a hit means no
fetch. The store is a **cache and a source**, never an authority: a document still carries its own
faces.

## Boundaries & Constraints

**Always:**
- **IndexedDB, and `localStorage` is refused on arithmetic** (D-16.2). The owner's words were "local
  storage"; the mechanism is the engineer's. `localStorage` is a ~5 MB per-origin **string** quota;
  storing bytes there means base64 at **+33%**; a measured `Sarabun-Regular.ttf` is **90,220 bytes →
  ~120 KB stored**; and its `QuotaExceededError` is synchronous with no partial-write path. **Write
  that reasoning into the module**, so the next person does not "simplify" it back.
- **Keyed by content hash, never by family name.** A family that changes upstream is a different key,
  not a silent substitution — the property AD-8 already buys the document.
- **The store never becomes an authority on a document.** A `.folio` carries its faces (CAP-2). The
  store shortens a fetch; it never stands in for what a file contains, and removing an entry never
  changes a saved document.
- **`AVAILABLE LOCALLY` is fetched faces, never host fonts** (D-16.2). The one Non-goal clause D-16.1
  leaves standing is *"No host fonts. Faces installed on the authoring or rendering machine are never
  enumerated or read."* The Local Font Access API is not used, referenced, or feature-detected.
- **Everything the embed command requires is stored with the bytes.** Licence identifier, licence text
  and copyright travel into the store, because a face offered from the store must be embeddable
  without a network — and `embedFontFamily` refuses without all three.
- **Storage failure is a degradation, stated.** A private window, cleared site data or a quota refusal
  leaves a working designer with an empty group, not an error the author cannot act on.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **The store would be read at render time.** FR33 is untouched: nothing is fetched or read from the
  machine at render. This store serves **authoring** only.
- **A stored face would reach a document without its licence record.** That would put a document its
  own parser refuses one step away.
- **`maximumCacheAssets` or the offline release contract would move.** This store is runtime data, not
  a release asset.

**Never:** enumerate host fonts · store in `localStorage` · key by family name · let a store miss
break a document that already embeds the face.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| First fetch of a family | Store empty | Bytes + licence text + copyright + metadata written under the SHA-256 | Write failure degrades to no-store |
| Re-pick, later document | Key present | **No fetch**; embed from stored bytes | No error |
| Re-pick, network down | Key present | Works; this is the store's whole point | No error |
| Family not stored, network down | Key absent | Stated: cannot add right now | Degradation |
| Private window / storage blocked | IndexedDB throws on open | Group empty; picks still fetch; message states it | Degradation, once, not per pick |
| Quota exceeded on write | Large face, full origin | Fetch and embed still succeed; the caching is what failed | Degradation, named |
| Author removes an entry | Entry present | Removed, with what is being removed named; documents unaffected | No error |
| Stored bytes fail to decode later | Corrupt entry | Entry treated as absent and dropped; refetch on next pick | Self-healing, logged honestly |

</intent-contract>

## Code Map

**Designer (`folio-designer/`)**
- `src/App.tsx:186-224` — the document-scoped carried-face registration, and its comment on why the
  lifetime is document-scoped rather than per-component. **The precedent for lifetime reasoning**, and
  the machine store is a third lifetime that must be argued the same way.
- `src/embedded-face-registry.ts`, `src/embedded-face-family.ts` — how a face's asset key becomes a CSS
  family, and `isCarriedFaceAssetKey`, the key-shape predicate. **The store's keys must satisfy it.**
- `src/App.tsx:608-627` — `pickCatalogueFamily`, which Story 16.1 widens; the store's read goes in
  front of its fetch.
- `src/offline-lifecycle.ts` — the service worker's lifetime, and the origin whose storage this is.
- `src/image-file.ts`, `src/component-asset-command.ts` — the other place bytes cross from the browser
  into a command; same shape, no persistence.
- `src/release-payload.ts:33` — `maximumCacheAssets = 64`. Untouched; assert it.

**Go (`folio-go/`)**
- `component_commands.go:2359-2410` — `embedFontFamily`. **The store feeds it exactly what a fetch
  feeds it**; Go cannot tell the difference and must not be able to.

## Tasks & Acceptance

**Execution:**
- `src/` — a font-store module over IndexedDB: `get(key)`, `put(record)`, `list()`, `remove(key)`, and
  one open path that degrades rather than throwing into callers. Records carry bytes, mediaType,
  family, style, licence, licenceText, copyright, source, scripts and the date fetched.
- `src/` — hash the bytes to the key in the browser **for the store's own addressing**, and assert in a
  test that the key the store computes equals the asset key Go derives, so the two addressings cannot
  drift.
- `src/App.tsx` — read the store before fetching; register stored faces for preview alongside the
  document's carried faces, without disturbing the document-scoped effect's keying.
- `src/` — the removal affordance, naming what it removes and stating that documents are unaffected.
- Tests: store round trip; hit avoids fetch; miss with no network degrades as stated; open failure
  leaves a working designer; corrupt entry is dropped and refetched; and a **red-provable** assertion
  that no code path enumerates host fonts.
- Docs: `font-catalogue.md` gains the store's description — what it is, what it is not, and that it is
  authoring-only.

**Acceptance Criteria:**
- Given a fetched face, when the fetch succeeds, then it is stored under the SHA-256 of its bytes with
  everything the embed command requires.
- Given a stored family, when it is picked again in any document on this machine, then nothing is
  fetched and the embed runs over the stored bytes.
- Given a stored family and no network, when it is picked, then it is embedded successfully.
- Given the dropdown, when it is opened, then `AVAILABLE LOCALLY` lists exactly the stored faces and no
  operating-system font.
- Given storage that cannot be opened or written, when the designer runs, then it still works and says
  what is degraded.
- Given a stored face the author removes, when it is removed, then documents that already embed it are
  unchanged.

## Design Notes

**Why the content hash and not the family name.** The store answers "do I already have these bytes",
which is the same question `assetKeyReferenced` and the `assets` map answer. Keying by name would make
the store answer a *different* question — "do I have something called Sarabun" — and the day upstream
changes the face, the store would hand over the old bytes under the new name and the document would
carry a face nobody chose.

**Why this is a separate story from 16.1.** 16.1 can ship without persistence and be correct; it would
simply re-fetch. Fusing them would put a storage failure in the same blast radius as a licence
refusal, and those are different failures with different right answers.

**What "on this machine" honestly means.** Origin-scoped browser storage: this browser, this profile,
this origin. Not the OS, not synced, not shared with another browser on the same machine. The UI
should not imply more than that.

## Verification

- `cd folio-designer && npm run test && npm run build`
- Browser: fetch a family; reload; confirm it is offered with the network disabled; remove it; confirm
  a document that embedded it still renders.
- The 23 golden digests, unmoved. `maximumCacheAssets` still 64.
