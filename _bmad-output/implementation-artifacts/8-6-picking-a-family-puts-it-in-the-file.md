---
title: 'Story 8.6: Picking a family puts it in the file'
type: 'feature'
created: '2026-09-02'
status: 'done'
review_loop_iteration: 0
followup_review_recommended: true
baseline_commit: '8d5059f8859ccb2b7c99d3fb4b16451d1793ddc9'
baseline_revision: 'b4885b2365235520ae650c278f017497b823ceeb'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-fonts/font-catalogue.md'
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
warnings: [multiple-goals, oversized]
deferred:
  - summary: >-
      The font-family listbox owns non-option children (`role="presentation"` list items) for its
      group headings, its empty state and the disk-font decline, which breaks the listbox's
      required-owned-elements rule for assistive technology.
    evidence: |-
      Pre-existing, not caused by this story: `git show b4885b2:folio-designer/src/App.tsx` already
      carried one such child at :1288 (the "This document declares no font chains." empty state).
      Story 8.6 added three more instances of the same pattern (two group headings and the
      permanently-rendered disk-font decline), taking the count from 1 to 4. The standard shape is
      `role="group"` with an `aria-label`, or moving the notes outside the `<ul>` and referencing
      them with `aria-describedby`.
    location: >-
      folio-designer/src/App.tsx (4 occurrences of role="presentation" inside the role="listbox")
    severity: low
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Choosing a typeface used to be a gesture that changed nothing. It now changes the file. When an
author picks a family, the typeface itself is copied into the document. Send that file to a
colleague, open it on a machine that has never had the typeface installed, and the pages come out
identical — the file is no longer asking the machine for anything. Picking the same family twice
stores it once, and a face nothing draws with any more is removed by the author's own action rather
than quietly at save time.

The file also carries whose the typeface is and on what terms: the copyright, and the full text of
the licence rather than merely its name. That is now a requirement rather than a courtesy — a
document that embeds a typeface without them is refused when it is opened, and told why.

One thing should be said plainly. For a while the catalogue was handing most of its typefaces the
wrong licence: seventeen of twenty-one travelled under another project's terms, reserving a name for
a typeface the document was not carrying. That is worse than saying nothing — it is a false
statement about what an author may pass on. Review caught it before release, and every face now
carries the text filed beside it.

Still absent, deliberately: bold and italic, a typeface from the author's own disk, and anything
fetched over the network.

<intent-contract>

## Intent

**Problem:** Story 8.5 shipped 21 catalogue typefaces into the designer, but no command can author an
embedded-face chain entry or write a font asset, so picking a family still changes nothing about the
document. Separately, a `.folio` can carry a face today with **no licence information at all** —
`font` and every key inside it is optional, and `licence` holds a short label, not the terms.

**Approach:** One new engine command embeds the picked face as a content-addressed asset, writes its
font record *including the licence text and copyright*, and declares a chain naming it by asset key —
one command, one history entry, one undo. **Make that record mandatory for an embedded face:** a
`.folio` whose chain names an asset lacking licence text or copyright is a **located load error**.
Fix DW-80 so the asset-reference walk can see font chains, then drop an unnamed font asset in the
**command** layer. The family control gains the catalogue as a second, visibly distinct group.

## Boundaries & Constraints

**Always:**
- **The drop is a command-layer mutation.** `writeAssets` is not touched: AD-9 / D-1.4.3's P1
  (`Parse(Serialize(d)) == d`) *forces* unconditional orphan preservation in the serializer, and its
  own comment says collecting orphans is a designer command, never a serializer side effect.
- **DW-80 is fixed before anything calls the reference walk for a font.** `assetKeyReferenced`
  currently returns `false` for every font asset. Extend it to walk `Fonts` entries' `AssetKey`, and
  add a test that reds if the font arm is deleted — mutate by deletion, not by falsifying a condition.
- **Dedupe by content hash**: insert only if the key is absent. A re-pick of an already-embedded
  family declares no second asset, no second chain, and no second undo entry.
- **A face resolves by asset key alone.** `family`/`style` are display identity and never substitute.
- **The pick reads bytes already present in the offline bundle.** Nothing is fetched from any host.
- **Version is derived, not asserted** (see Design Notes). `SupportedMajor` stays `2` — and the
  derivation must *show* it stays, not assume it.
- **Licence text and copyright are REQUIRED on an embedded face** (owner ruling, 2026-09-02: Folio is
  unreleased, there are no `.folio` documents to protect, and breaking is free until `folio-go/v0.1.0`
  / Story 15.3). Not optional, not best-effort. **The requirement is scoped to an asset a chain
  actually names** — an unreferenced font asset is not an embedded face and stays legal, which is what
  keeps D-1.4.13's "a font asset no chain references must not raise the document" intact.
- **Do not soften a rule to spare an existing document.** There are none. Where the old spec would
  have made something optional for compatibility, make it right instead.
- **"Already-touching, not going-looking" (D-000.15).** The freedom licenses this story to choose the
  *correct* shape over the *compatible* shape **in the format area it already touches** — the font
  asset record. It does **not** license opening any other format area. Make no `.folio` change outside
  the embedded-face record, however cheap breaking currently is.
- **Record what the freedom was spent on.** D-000.15 obliges Story 15.3 to be handed that list, so
  this story's narrowing must be registered when it lands, not reconstructed later.
- **23 golden digests byte-identical.** `maximumCacheAssets` stays `64`; 44 of 64 slots are used and
  this story adds **no** build asset.
- Every recorded measurement carries **command, commit, tree state and working directory** (D-8.4j.8).
- Standing reds are exactly three, by identity (see Verification). A fourth is a real failure.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **Any of the 23 golden digests moves.** This is *not* a back-compat obligation — it is the
  byte-identity witness, and C6 makes an unexplained moved hash a defect until proven intended. One
  fixture input must be edited (see Design Notes); its **PDF** digest is expected **not** to move, and
  a document declaring no font choice must hash identically. If a digest moves anyway, HALT: it is
  re-recorded only when deliberate, attributed and reasoned, never silently.
- **`SupportedMajor` would move.** The derivation below says it must not; if the work concludes
  otherwise, that is a halt, not a judgement call.
- `maximumCacheAssets` would be raised, or a build asset added.
- Anything would alter `Subset.Program` bytes or how the PDF producer subsets.
- A face's declared script coverage disagrees with that face's own `cmap`.
- A **licence-gate (`lint/`) defect**: this is *not* a halt — **register it to Epic 15 and keep
  going**. Do not fix it, do not add tests there, do not escalate it (D-000.11, D-000.13).

**Never:** bold, italic, oblique or variable axes (Epic 11, D-000.7) · CJK catalogue families · a live
font service, arbitrary URL, or download-on-first-use · enumerating host-installed fonts (AD-8) ·
embedding a font from the author's own disk — **declined explicitly, not by silence** · save-time
subsetting or any change to how the PDF producer subsets · a container format · `.woff2`/`.woff`/
`.ttc` · a byte threshold or budget figure (Story 8.4d owns it) · any `WITH`-form licence expression
(D-8.5.17 — **parser scope, not licence policy**) · a TypeScript model of the `fonts` map (AD-15).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| First pick of a family | Document declares no chain for it | One command: asset inserted under its SHA-256, `font` record written with family/style/licence/licenceText/copyright/source, chain declared with `{"asset":"<key>"}` head; document declares `2.0`; one undo entry | No error expected |
| Re-pick, same family | Asset key already in `Assets`, chain already names it | No second asset, no second chain; the existing chain is offered; canonical bytes unchanged so **no** history entry is pushed | No error expected |
| Re-pick after the chain was deleted | Asset key present, no chain names it | Chain re-declared naming the existing key; asset not duplicated | No error expected |
| Proposed fallback tail | Picked face covers Latin only | Tail is the shipped faces for the uncovered scripts; author may edit it in the chain editor | No error expected |
| Last chain entry naming a face removed | Asset becomes unnamed by any chain | Command drops that asset, scoped to the key just un-named — never a document-wide sweep | No error expected |
| Font asset still named by another chain | Two chains name one key; one removed | Asset **retained** — this is the arm DW-80 would have got wrong | No error expected |
| Chain names an asset with no licence text | Hand-written `.folio`; `font.licenceText` or `font.copyright` absent or explicitly `null` | **Load error**, located — naming the asset key and the chain entry that makes it an embedded face | Refused at load; never a warning, never a best-effort render |
| Font asset present but no chain names it | Orphan font asset, `font` record absent or partial | **Loads clean** — not an embedded face, so the requirement does not apply and the document is not raised to `2.0` | No error expected |
| Author asks to embed a disk font | — | Declined explicitly with a stated reason: the catalogue is the source | Refusal, located at the originating control |
| Catalogue list rendered | Document declares 3 of 21 families | Declared and not-yet-declared entries visibly distinct; picking moves an entry from the second group to the first | No error expected |

</intent-contract>

## Code Map

**Go — engine (`folio-go/`)**
- `internal/template/model.go:393-414` — `FontRecord`: `Family`/`Style`/`Licence`/`Source`, all
  `Presence[string]`, plus `Extra []Field` passthrough. **No struct tags** — hand-serialized, so JSON
  key names live in the writer/decoder. `Presence` means an explicit `null` round-trips distinctly
  from absence; write guards for both arms.
- `internal/template/model.go:416-432` — `Asset{Data, MediaType, Font Presence[FontRecord], Extra}`;
  `Font` emitted only when `Set`, which is why image assets stayed byte-identical at 8.3.
- `internal/template/model.go:~163-197` — `FontChainEntry` (`Face`/`AssetKey` partition, `Embedded()`
  discriminant, `FaceEntry`/`AssetEntry` constructors); `Fonts map[string][]FontChainEntry`.
- `internal/template/serialize.go:503-528` — `writeAssets`. **READ-ONLY.** The comment at `:510-514`
  is the constraint: orphans "PRESERVED here unconditionally — D-1.4.3's P1 forces this; there is no
  policy latitude to drop one", and `:514-517` "garbage-collecting orphans is a designer feature …
  never a serializer side effect." The loop at `:526-528` has no reference test at all.
- `internal/template/version.go:77-79` — `SupportedMajor = 2`, `SupportedVersion = "2.0"`. Do not
  move. Guards at `:141` and `linespacing_test.go:317-318`, `version_test.go:32-36`.
- `component_commands.go:790-799` — `assetKeyReferenced`. **DW-80 lives here**: the only `true` arm
  requires `el.Type == template.ElementImage && el.Asset.Set && !el.Asset.Null`, and the walk never
  reads `t.doc.Fonts`, so it returns `false` for every font asset. Comment at `:784-789` warns that
  `assetKeyReferenced`, `findComponent` and `addCanvasImagePaint` must be updated together.
- `component_commands.go:667-770` — **the precedent to copy** (`setComponentAsset`, D-5.13.3): closed
  payload, Go decodes/bounds/validates/hashes, inserts only if the key is absent (`:754`), and drops
  the orphan scoped to the one key just repointed away from (`:766-768`).
- `component_commands.go:1901-2297` — the 8.1/8.3 font-chain commands (`addFontChain`,
  `renameFontChain`, `deleteFontChain`, `addFontChainEntry`, `moveFontChainEntry`,
  `removeFontChainEntry`), `applyFontChainCommand` transaction wrapper at `:1977-2009`, reference
  walk `fontChainReferences` at `:2274-2297`, bound `maxCanvasFontChainEntries = 64` at `:1913`.
  **Gap:** `addFontChainEntry` (`:2179-2200`) only ever builds a `FaceEntry`; **nothing can author an
  `AssetEntry` or write a font asset.** That is this story's central addition.
- `wasm/engine.go:210-262` — `Apply`; the no-op short-circuit at `:243-245` and the single
  `pushUndo` at `:258-260` are why one command kind gives AC1 one undo entry **for free**.
- `internal/fontset/fontset.go:730` — `(*Font).Subset`, the single subsetting site. **READ-ONLY.**

**Designer (`folio-designer/`)**
- `font-catalogue.json` — 21 entries `{id, directory, file, family, licence}`. **No `scripts` field**,
  though `font-catalogue.md` says there should be one. Read today only by the build.
- `scripts/build-wasm.mjs:109-190` — validates and fingerprints the catalogue into
  `src/generated/runtime/catalogue-<id>.<hash>.ttf`, emits `src/generated/runtime-fonts.css`.
- `src/generated/offline-assets.ts` — exports only the 9 named slots; **catalogue faces have no
  TypeScript-side export**, so `src` cannot enumerate the catalogue at all today.
- `src/App.tsx:1243-1294` — `FontFamilyProperty`: a combobox over `CanvasProjection.fontFamilies`,
  i.e. **only chains the document already declares**. `choose(name)` at `:1269`, listbox at
  `:1289-1292`, empty state "This document declares no font chains."
- `src/font-chain-command.ts` — opaque builders; its comment: "THE FIELD COUNT IS PART OF THE
  CONTRACT". `src/font-chain-control.ts` — `FontChainControl.action` union, refusal anchoring.
  `src/App.tsx:575-591` — `applyFontChain` dispatch. `src/App.tsx:599-628` — `applyImageAsset`, the
  browser-side shape to copy.
- `src/release-payload.ts:33` — `const maximumCacheAssets = 64`, line-anchored and regex-read by
  `scripts/offline-release-contract.mjs:75-81`. Must keep the exact single-line form.
- `src/font-catalogue.test.ts` — reads assertions **out of each binary's own tables** (nameID 13 vs
  the declared SPDX id). This is the precedent for verifying declared coverage against `cmap`.

**Docs to amend**
- `_bmad-output/specs/spec-fonts/SPEC.md` `## Open Questions` — two questions resolved here.
- `_bmad-output/specs/spec-fonts/font-catalogue.md` — `## Picking a family`, `## Removing a family`,
  `## Per-entry record`, the stale OFL-only Licence row, and the stale line-26 pointer.
- `_bmad-output/specs/spec-folio/folio-format.md` — the font-asset section (~`:509-555`).

## Tasks & Acceptance

**Execution:**
- `folio-go/component_commands.go` — extend `assetKeyReferenced` (`:790-799`) to also walk
  `t.doc.Fonts` chain entries' `AssetKey` — **DW-80's fix, landed before any font orphan drop calls
  it.** Update `findComponent`/`addCanvasImagePaint` in step per the `:784-789` warning.
- `folio-go/component_commands_test.go` — a test that embeds a face, names it from a chain, and
  asserts `assetKeyReferenced` is `true`; **red-prove it by deleting the font arm**, not by
  falsifying the image condition. Add the two-chain retention case from the I/O matrix.
- `folio-go/internal/template/model.go` — add `LicenceText` and `Copyright` (`Presence[string]`) to
  `FontRecord`. Emit/decode them in the hand-written writer and decoder, alphabetically sorted
  (`copyright`, `family`, `licence`, `licenceText`, `source`, `style`).
- `folio-go/internal/template/parse.go` — **the new required-field validation.** For every asset a
  chain names by `{"asset": key}`, refuse the document unless `font.licence`, `font.licenceText` and
  `font.copyright` are all present and non-null and non-empty. The error locates the asset key **and**
  the chain entry (`fonts.<chain>[<i>]`) that makes it an embedded face. An **unreferenced** font
  asset is untouched by this rule. Cover the `Presence` null arm explicitly — an explicit JSON `null`
  must be refused, not admitted by a guard written only for the absent case.
- `folio-go/internal/template/parse_test.go` — the refusal's positive and negative arms, plus the
  orphan-asset control that must still load clean.
- `fixtures/embedded-font/input.folio` — add the required `licenceText` and `copyright` to the
  embedded face's `font` record; it is otherwise invalid under the new rule. Update
  `fixtures/embedded-font/README.md` to say why. **`expected.pdf` must not change** — see Design Notes.
- `folio-go/component_commands.go` — add ONE new command kind that embeds the face and declares the
  chain in a single `applyFontChainCommand`-style transaction: decode + bound the bytes, hash to the
  asset key, **insert only if absent**, attach the `FontRecord` (family, style, licence, licenceText,
  copyright, source — the command must refuse to embed a face whose catalogue row cannot supply them,
  so the writer can never produce a document its own parser would reject), and insert an `AssetEntry`
  into the chain. Shape it on `setComponentAsset` (`:667-770`).
- `folio-go/component_commands.go` — the font orphan drop: on removal of the last chain entry naming
  a key, `delete(t.doc.Assets, key)` **scoped to that key**, guarded by the now-correct
  `assetKeyReferenced`. Never a document-wide sweep. `serialize.go` is not touched.
- `folio-go/wasm/*_test.go` — assert the pick is **one** history entry with working undo/redo, and
  that a re-pick of an already-embedded family pushes **no** second entry (the `:243-245`
  canonical-bytes short-circuit).
- `folio-designer/font-catalogue.json` — add the `scripts` field `font-catalogue.md` already
  specifies, one row per face.
- `folio-designer/src/font-catalogue.test.ts` — verify each declared `scripts` value against that
  face's own `cmap`, following the existing nameID-13 precedent. A disagreement is a Block If.
- `folio-designer/scripts/build-wasm.mjs` — emit a typed catalogue module beside `runtime-fonts.css`
  so `src` can enumerate families, ids, licence and coverage. **No new build asset.**
- `folio-designer/src/App.tsx` — split `FontFamilyProperty`'s list (`:1262`, `:1289-1292`) into
  declared / not-yet-declared groups and fork `choose()` (`:1269`): a declared name keeps today's
  `fontFamily` property commit; a catalogue family sends the new command. Read the face bytes from
  the precached content-addressed URL, as `runtimeAssetUrls` assets are already read.
- `folio-designer/src/font-chain-command.ts`, `src/font-chain-control.ts` — one new opaque builder and
  one new `FontChainControl.action` member for refusal anchoring.
- `_bmad-output/specs/spec-fonts/SPEC.md` — strike through **in place**, in the existing `~~…~~` +
  `**SETTLED (Story 8.6, …): …**` + derivation form: (1) *"Does the licence record live inline on each
  font asset, or in one document-level notice block?"* → **inline, with the actual text** (D-8.6.1);
  (2) *"May an author embed a font file from their own disk…?"* → **declined; the catalogue is the
  source** — declined explicitly rather than by silence.
- `_bmad-output/specs/spec-fonts/font-catalogue.md` — amend `## Removing a family` so the drop is
  stated as a **command**, not a save-time side effect; pin `## Picking a family` step 2's chain
  naming and step 3's tail; correct the OFL-only Licence row to D-8.5.3's four identifiers; fix the
  stale line-26 pointer.
- `_bmad-output/specs/spec-folio/folio-format.md` — the font-asset section (~`:509-555`). Amend the
  sentence "`font` is **optional**, and every key inside it is optional": it stays true for an
  **unreferenced** font asset, and is now **false** for an asset a chain names — there,
  `licence`, `licenceText` and `copyright` are required and their absence is a load error. Document
  the two new keys, and record explicitly that **the version trigger and `SupportedMajor` do not
  move**, with the derivation.
- `_bmad-output/implementation-artifacts/deferred-work.md` — register two things: (1) **what the
  format freedom was spent on here** (the embedded-face record narrowing), for Story 15.3 per
  D-000.15; (2) **the subsetting attribution gap** — the produced PDF's subset carries no `name`
  table, so nameIDs 0/7/13/14 do not reach the reader's font program. Out of scope by a shipped
  Non-goal; register it, do not fix it.

**Acceptance Criteria:**
- Given a catalogue pick, when it is committed, then one command embeds the face and declares a chain
  naming it, as one history entry, and one undo removes both.
- Given a family already embedded in this document, when it is picked again, then no second copy is
  stored — the content hash decides — and the existing chain is offered rather than a duplicate
  declared.
- Given a picked face that does not cover every script the document may render, when the chain is
  proposed, then its tail is the shipped faces for the uncovered scripts, and the author can edit
  that tail in the chain editor.
- Given the family control, when the catalogue is shown, then entries the document already declares
  and entries it does not are visibly distinct, and picking is what moves an entry from the second
  group to the first.
- Given a font asset no chain names any longer, when the author's action un-names it, then it is
  dropped by the command — never by the serializer — so a file cannot accumulate megabytes of faces
  nothing draws with.
- Given any document that embeds a catalogue face, when it is saved, then the asset's `font` record
  carries family, style, licence identifier, **licence text**, **copyright** and source.
- Given a `.folio` whose chain names a font asset lacking licence text or copyright, when it is
  loaded, then it is **refused with a located error** naming the asset and the chain entry — not
  warned about, not rendered best-effort — while a font asset no chain names still loads clean.

## Spec Change Log

**1. `serialize.go` DOES have a diff, and the spec asked for both.** The `## Verification` manual
check says "`folio-go/internal/template/serialize.go` shows **no diff**", while `## Tasks &
Acceptance` says to "Emit/decode them in the hand-written writer and decoder". Those cannot both
hold: `writeFontRecord` lives in `serialize.go`, and two keys that are not emitted do not round-trip,
which breaks D-1.4.3's P1 immediately. Resolved in favour of Tasks & Acceptance, which is the
executable contract. **The constraint that actually mattered is honoured exactly:** `writeAssets`
(`:503-528`) and its orphan-preservation loop are untouched — the whole diff to that file is six
lines inside `writeFontRecord` adding `copyright` and `licenceText` to its key list, plus a comment.
Verified by reading the diff, not asserted.

**2. The licence requirement reaches one rule beyond the font-asset record, and it is registered
rather than silently absorbed.** The rule as written in Tasks & Acceptance is "for every asset a
chain names by `{"asset": key}`" — no media-type carve-out — so it also applies to a chain entry
naming a **non-font** asset, which D-1.8.1 (as amended) says loads clean and errors at render. Such
a document is now refused at LOAD if it does not state terms; a document that does state them still
loads and still fails at render, so D-1.8.1's promise is intact downstream of the new check. The
alternative — scoping the rule to recognised font media types — was rejected because it is evadable:
declaring `font/woff2` would let a face travel with no terms at all, which is the exact thing the
rule exists to stop. Registered in `deferred-work.md` DW-138 as part of what the format freedom was
spent on.

**3. The proposed fallback tail is computed in the DESIGNER and sent as a `[]string`.** The spec
does not say which side derives it. It is the browser's, on the precedent `addFontChain` already
sets (`entries` arrives as a `[]string`): the tail is a list of SHIPPED FACE NAMES, and the engine
gains no knowledge of script coverage it did not have before. The one entry that names an asset is
the one the command builds itself from the bytes it just hashed, so no caller can put a second
embedded entry into a chain by writing one down.

**4. The pick does not also set `fontFamily` on the selection.** Not stated either way by the spec.
They are two decisions — "carry this typeface" and "draw this box with it" — and fusing them would
make one undo ambiguous, against AC1's "one undo removes both" being about the asset and the chain.

**5. `deleteFontChain` drops orphaned faces too, not only `removeFontChainEntry`.** Tasks name "the
last chain entry naming a key"; deleting a whole chain un-names every entry in it at once, which is
the same event. AC5 is written about "the author's action un-names it", so both actions collect,
each scoped to exactly the entries it removed.

**6. `scripts` is a closed three-value vocabulary (`latin`, `thai`, `cjk`).** `font-catalogue.md`
asks for the field without fixing its values. Closed because an unrecognised script proposes no
fallback for itself and the chain silently draws tofu. Measured against all 21 committed faces: 19
Latin-only, 2 Thai-only, none CJK — which agrees with the catalogue's own CJK exclusion.

**7. The catalogue's licence text and copyright are DERIVED, not declared.** `font-catalogue.json`
gained only `scripts`. `licenceText` is read at build time from the unmodified upstream `LICENSE*`
committed beside each binary, and `copyright` from the face's own `name` table (nameID 0) — so
neither is a hand-copy that could contradict the bytes it describes.

**Corrected at review (P1).** The emitted module first keyed licence texts by SPDX identifier, on the
assumption that "the OFL is the OFL". That is FALSE OF THE FILES: the SIL OFL carries a per-project
preamble — a copyright line and a Reserved Font Name — so two OFL-1.1 faces ship two different texts.
The cache was filled from whichever face reached each identifier first, which gave **17 of 21 faces
another project's licence text** (every OFL-1.1 face emitted cascadiacode's LICENSE, "with Reserved
Font Name Cascadia Code" and all), inverting the story's central promise. It is now read per face
from that face's own directory; the bundle carries 21 texts instead of 2, which is the correct trade
and is stated in the generator rather than left to be rediscovered. No build asset is added and the
release cache stays at 44.

**It shipped green because nothing observed the generated module.** `font-catalogue.test.ts` read the
manifest and the binaries; `App.test.tsx` asserted only `toBeTruthy()`. The artifact BETWEEN them —
the only thing the pick actually reads — was checked by nothing, and that gap is now closed per face
(P2).

## Review Triage Log

### 2026-09-02 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 7: (high 1, medium 3, low 3)
- defer: 1: (high 0, medium 0, low 1)
- reject: 10: (high 0, medium 1, low 9)
- addressed_findings:
  - `[high]` `[patch]` `build-wasm.mjs` keyed `licenceTexts` by SPDX identifier, so 17 of 21 faces emitted another project's licence text (every OFL-1.1 face carried cascadiacode's, Reserved Font Name clause included). Fixed to read each face's own `LICENSE*`; re-measured independently at 0 mismatches of 21.
  - `[medium]` `[patch]` Nothing observed the generated catalogue module — only `toBeTruthy()`. Added per-face assertions tying emitted `licenceText` to that face's own `LICENSE*` and emitted `copyright` to nameID 0 read by a second, independent sfnt reader.
  - `[medium]` `[patch]` `embedFontFamily`'s chain-name collision refusal was exercised by nothing — deleting the branch left the whole `folio-go` suite green while a pick silently overwrote a declared chain. Added `TestEmbedFontFamilyRefusesAChainNameTheDocumentAlreadyTakes`, mutation-proved by deletion.
  - `[medium]` `[patch]` The chain-entry-scoped rule's reach to a non-font asset — the story's widest registered consequence — had no positive test. Added `TestANonFontAssetNamedByAChainMustAlsoStateItsTerms` with the with-terms control in the same test.
  - `[low]` `[patch]` `requireEmbeddedFaceLicence` accepted whitespace-only terms (`Value == ""`). Now `strings.TrimSpace`, with blank rows for all three keys, and the same trim on the command side.
  - `[low]` `[patch]` `scriptFallbacks`' three face names were tied to nothing; a rename would propose a fallback the engine skips. Added a guard against `shippedFamilies`.
  - `[low]` `[patch]` `TestEmbedFontFamilyRefusesAFaceThatCannotStateItsTerms` asserted a substring matching every row. Now asserts the exact per-field message and the located `DataPath`.

**Deferred (1).** `role="presentation"` non-option children inside `role="listbox"` — the pattern pre-existed at `App.tsx:1288` at `b4885b2`; this story added three more instances. Pre-existing, surfaced incidentally.

**Rejected (10), each with the population or path verified.**
1. *Catalogue re-shows a family after its chain is renamed; the pick then silently no-ops.* Verified against matrix row 2 ("chain already names it"): every expected output holds — no second asset, no second chain, "No error expected", and the existing chain is offered in the declared group under its current name. `App.tsx:1030` filters on `CanvasProjection.fontFamilies`, which carries names only and cannot reconstruct asset keys, so a hash-keyed filter needs a projection change the contract does not authorise. Severity medium; recorded as a residual risk.
2. *Concurrent picks race / no fetch timeout.* Both paths verified: two picks of DIFFERENT families produce two independently legal commands; two picks of the SAME family both reach `assetKeyReferenced` (`component_commands.go:2400`) and the second returns `nil`. Neither path yields a duplicate chain or asset.
3. *`cmapCoverage`'s 70000 format-12 bound and last-record-wins subtable choice.* Population verified: all 21 catalogue rows declare only `latin` (19) or `thai` (2); every probe in those scripts is BMP and reachable in a format-4 subtable, so the bound is never approached. The test also asserts `covered.size > 0` and passed both directions for all 21.
4. *Tail could name the picked family itself or repeat a name.* Population verified: the three fallback families (`Noto Sans`, `Noto Sans Thai`, `Noto Sans SC`) are disjoint from all 21 catalogue family names (nearest: `Noto Sans Thai Looped`, `Noto Serif`, `Noto Serif Thai`), and `scriptFallbacks` holds one entry per script so the tail cannot repeat.
5. *`style: "Regular"` hardcoded in the generator.* Population verified: all 21 rows are Regular instances; weighted and italic faces are excluded by a shipped ruling the contract restates under **Never**.
6. *Designer imports a gitignored generated module.* Verified: `.gitignore:68-70` already ignores `src/generated/runtime/`, `offline-assets.ts` and `runtime-fonts.css`, all imported by `src`; `npm run typecheck` is literally `npm run build:wasm && tsc -b`, so generation precedes every typecheck. Pre-existing pattern.
7. *`font-catalogue.md`'s CJK-exclusion criterion vs a vocabulary admitting `cjk`.* Population verified: no catalogue row declares `cjk`; the key exists only so a Latin face receives a CJK *fallback*, which is the exclusion working rather than contradicting itself.
8. *Two normalisations of "verbatim" (`TrimRight("\n")` vs `.trimEnd()`).* Path verified: `embeddedFontLicenceText()` is read only by the fixture generator and `.trimEnd()` only by the designer generator; no value crosses between them.
9. *DW-138/DW-139 lack an assigned action; `SPEC.md` reads as if attribution is solved.* Verified: the settled text at `SPEC.md:165-166` answers where the record *lives* (inline, with the text), not the PDF question, which DW-139 records separately as owner-owned.
10. *Frontmatter carries both `baseline_commit` and `baseline_revision`, and the triage log was empty.* Verified: both are this workflow's own bookkeeping (`baseline_commit` = the plan gate's measurement commit, `baseline_revision` = this dispatch's HEAD), and the log is written by this step.

**Detail of the seven patches, as applied.**

- **P1 [HIGH] — the generated catalogue gave 17 of 21 faces another project's licence text.**
  `build-wasm.mjs`'s `licenceTexts` was keyed by SPDX identifier. Fixed: read per face from that
  face's own directory (`licenceTextOf`), the SPDX cache and its `continue` skip deleted — which
  also restores the "exactly one `LICENSE*` beside it" check for every face rather than for the
  first face of each identifier. Re-measured after the fix: **0 mismatches of 21**.
- **P2 [MEDIUM] — nothing observed the generated module, which is why P1 shipped green.**
  `src/font-catalogue.test.ts` now imports `src/generated/font-catalogue.ts` and asserts, per face,
  that the emitted `licenceText` equals that face's own `LICENSE*` and that the emitted `copyright`
  equals nameID 0 read by the test file's OWN sfnt walk (`instanceOfFile` gained it). Plus a
  top-level population check, so a truncated or stale module reds in one sentence instead of face by
  face.
- **P3 [MEDIUM] — `embedFontFamily`'s chain-name collision refusal was exercised by nothing.**
  Every existing embed test either deduped first or picked a free name. Added
  `TestEmbedFontFamilyRefusesAChainNameTheDocumentAlreadyTakes`, which picks a NEW face (Noto Sans,
  so the dedupe short-circuit cannot answer first) into the declared `body` chain and asserts the
  located refusal, the untouched chain and the unstored asset.
- **P4 [MEDIUM] — the rule's reach past font assets had no positive test.**
  `chain_face_names_test.go` had been amended only to SATISFY the new rule, making the story's widest
  registered consequence a precondition asserted nowhere. Added
  `TestANonFontAssetNamedByAChainMustAlsoStateItsTerms`: an `image/png` named by a chain with no
  terms is refused at load, located at both `assets.<key>.font.licence` and `fonts.body[1]` — with
  the control in the same test showing that the identical document WITH terms still loads and still
  fails at render, so D-1.8.1 is gated rather than replaced.
- **P5 [LOW] — whitespace-only terms passed.** `requireEmbeddedFaceLicence` now tests
  `strings.TrimSpace(...) == ""`; blank rows added for all three keys. The **command** got the same
  trim: `commandString` refuses `""` and stops there, so without it a pick could hand the
  transaction's reparse a document its own parser rejects, arriving as the unlocated "font chains did
  not pass format validation".
- **P6 [LOW] — `scriptFallbacks`' three face names were tied to nothing.** One guard now asserts
  every value is in `shippedFamilies` (itself checked against the hand-written `@font-face` rules),
  with the reason in the message: the engine SKIPS a chain entry naming a face it was not given, so a
  rename here would silently propose a fallback that draws nothing.
- **P7 [LOW] — a substring assertion that matched every row.** The refusal sentence for a missing
  `licence` contains "licence" and so does the one for `licenceText`, so the rows did not distinguish
  between themselves. Now asserts the exact per-field message plus the DataPath, and each key is
  driven through both the empty and the blank spelling.


## Design Notes

**Where the measurements were taken.** Every code, digest and line-number measurement in this spec was
taken at `8d5059f` (branch `main`, tree clean). HEAD advanced to `133f14a` mid-plan-gate when the owner
ruling and D-000.14/D-000.15 were recorded; `git diff --stat 8d5059f..133f14a` touches only
`epic-8-15-decision-log.md` and this spec, so **no code moved** and every anchor below still holds. The
implementer should still re-check the Code Map anchors against the dispatch's own HEAD before relying
on a line number.

**The version trigger, derived from the pre-reader test — not asserted.** `folio-format.md:47`: a
document declares the **lowest version its own content requires**. D-7.3.1's test is *would a
pre-`2.0` reader refuse the file, or render it wrong?*

- **Does this story add a new trigger?** No. The required keys can only appear on an asset a chain
  names by `{"asset": key}` — and that entry shape **already** forces `2.0` (settled at Story 8.3; a
  `1.x` reader refuses an object chain entry outright). So every document carrying the new keys is a
  `2.0` document for a reason that already shipped. The keys are a **record about the asset**; they
  reach no output byte, so a pre-`2.0` reader could not render such a file wrong — it refuses it
  first, on the entry shape. Scoping the requirement to *referenced* assets is what keeps D-1.4.13
  intact: an unreferenced font asset still leaves the document at `1.x`.
- **Does `SupportedMajor` have to move?** No — and it must **not**. A bump would make every document
  declare `3.0`, including the twenty-two fixtures that make no font choice at all; their bytes would
  change and their PDFs would be re-rendered, which **moves goldens for a reason unrelated to this
  story's content**. That is precisely the failure C6 exists to catch. `SupportedMajor` stays `2`.
- **What about tightening validation on existing `2.0` documents?** A document has no way to declare
  "I am missing licence text", so the new refusal is a *reader strictness* change, not a version
  trigger — version describes the document, never the writer. Normally that tension would force a
  MAJOR bump so older files kept loading; the owner's ruling removes it, because **there are no older
  files**. Breaking is free now and expensive after `folio-go/v0.1.0` (Story 15.3), which is the
  argument for making the format right here rather than deferring it.

**The one fixture that must change, and why no digest should move.**
`fixtures/embedded-font/input.folio` is the only fixture with an embedded-face chain entry
(`fonts.body[1]`, asset `c94562c1…`). Its `font` record carries `family`, `licence`, `source` and
`style` but **no licence text and no copyright**, so under the new rule it is invalid at load and must
be amended. Measured at `8d5059f`: the recorded digest in both `fixtures/embedded-font/expected.json`
and `signoff.json` is `f533b04b…`, which is the SHA-256 of **`expected.pdf`**, not of `input.folio` —
so amending the input moves no recorded digest **provided** the `font` record reaches no output byte.
Two independent reasons say it does not: `folio-format.md` states the engine derives none of it from
the bytes and none of it is required to render; and the subsetting measurement below shows the PDF
carries no `name`-table strings at all. **Expected outcome: all 23 digests hold.** This is an
expectation with a stated basis, not a guarantee — if one moves, that is the Block If.

**THE SUBSETTING MEASUREMENT — taken at this gate, and the answer is bad.** D-8.6.1 left this
unverified and required it before 8.6 ships. Measured at `8d5059f`: folio subsets through
`github.com/boxesandglue/textshape v0.0.15` (`internal/fontset/fontset.go:730`). Its
`handleOptionalTables` (`subset/execute.go:531-546`) copies `name`/`post`/`gasp` **only** when
`PassThroughTable(ot.TagName)` or `FlagPassUnrecognized` is set; folio sets neither. Against the real
shipped catalogue face `catalogue-inter.<hash>.ttf`, **the entire `name` table is absent from the
subset** — nameIDs 0 (copyright), 1 (family), 6 (PostScript name), 7 (trademark), 13 (licence
description) and 14 (licence URL) are all dropped; confirmed on a second face, Roboto.

Two consequences, and they point in opposite directions:
1. **Inside the `.folio`, attribution is intact.** Subsetting is render-time only — a non-goal
   forbids save-time subsetting — so the document carries the **unsubsetted** face, name table and
   all: nameIDs 0, 7, 13 and 14 travel with the bytes. The explicit `copyright` and `licenceText`
   keys are therefore *redundancy with a purpose* rather than the only carrier — they put the terms
   where a person reading the JSON can see them, and where a check can require them, without
   depending on a reader that can parse a font binary. That the binary also carries them is a
   measurement, not an assumption.
2. **Inside the produced PDF, they are gone.** This is a real, previously-unmeasured gap against
   D-8.6.1's "including the strings inside the font binary". It is **out of this story's scope by a
   shipped Non-goal** ("no change to how the PDF producer subsets"), and acting on it would move
   every subset tag and every one of the 23 goldens. Register it; do not fix it here.

**Why the pick is not a fetch.** Story 8.5 took the bundled route: the 21 faces are already 21 of the
44 release assets, precached and content-addressed. The pick reads bytes **already on the machine**,
exactly as `runtimeAssetUrls` assets are read — which is what `font-catalogue.md` step 4 already
says. No new build asset, no new cache slot, 44 of 64 unchanged, and no middle tier is invented.

**Licence text and copyright: required, by owner ruling.** D-8.6.1 left this as a format-change
question because making it mandatory would strand earlier documents. The owner settled it on
2026-09-02: **Folio is unreleased, no `.folio` documents exist, and the format may be broken.** So it
is specified as required rather than framed as a question — a chain that names an asset lacking
licence text or copyright is refused at load. Hand-written templates remain first-class (FR12/S9);
they are simply held to the same rule, and the rule is stated in the format document rather than
enforced only by the designer's writer. Note the requirement is **conditional on reference**, which is
what stops it colliding with D-1.4.13.

Inline-with-the-text is not this story's invention either: D-8.6.1 answers the `spec-fonts` open
question *"inline on each font asset, or one document-level notice block?"* with **inline, with the
actual text**, and `font-catalogue.md` already promised "identifier plus the text that travels into
the document". The cost is duplication — a document embedding three OFL families carries three copies
of ~4 KB of near-identical text. That is accepted deliberately: an asset that is passed on alone must
carry its own terms, which a document-level block would not survive.

## Verification

**D-000.4 override, invoked with its reason:** the cadence is per-epic, but this story changes **what
a `.folio` carries** — a new format key, a new asset-writing command, and a mutation that deletes
assets. Its own correctness is format- and round-trip-shaped, so the serialization/round-trip
integration commands belong in this story's own `## Verification`, not at the epic boundary.
The four `FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity` and Playwright stay **excluded** —
those are Epic 8's boundary gate.

**Three standing reds, by identity. A fourth is a real failure.** (1) `folio-go`
`TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)` — **1815 pass / 2 fail / 5 skip**; (2)
`gofmt -l` → `lint/internal/rules/licencegraph_test.go` — **do not reformat**; (3) designer
`npm run lint` → **exactly 4** `only-export-components`. Designer baseline: **42 files / 424 tests**.
Record every invocation with command, commit, tree state **and working directory** (D-8.4j.8).

**Commands:**
- `shasum -a 256 fixtures/*/expected.pdf > <scratch>/digests.before` — **repo root, before the first
  edit**; expect exactly 23 lines.
- `cd folio-go && go test -count=1 ./...` — standing red 1 only.
- `cd folio-go && go test -count=1 -run 'Serialize|Parse|RoundTrip|Asset|Font' ./internal/template/ -v`
  — the round-trip/serialization integration set, per the D-000.4 override above. `Parse(Serialize(d))
  == d` must hold for a document carrying an unreferenced font asset.
- `cd folio-go && go vet ./...` — no output.
- `cd folio-go && go vet -tags=matrix ./...` — matrix-tagged code still compiles.
- `gofmt -l folio-go lint` — **from the REPO ROOT** (from `lint/` it prints `lstat` errors that read
  as clean); standing red 2 and nothing else.
- `cd lint && go test -count=1 ./...` — **`-count=1` always**; a cached `ok` here is no measurement.
- `cd lint && go run ./cmd/genmanifest` — **from inside `lint/`** — then from the **repo root**:
  `git diff --exit-code -- lint/MANIFEST.md`.
- `cd folio-designer && npm run typecheck` — exit 0.
- `cd folio-designer && npm run lint` — exit 0 with standing red 3.
- `cd folio-designer && npm test` — at or above the 42-file / 424-test baseline.
- `cd folio-designer && npm run test:e2e:compile` — exit 0. **Compile only; do not report as a run**
  (DW-101).
- `cd folio-designer && npm run build` — node v24.16.0.
- `cd folio-designer && npm run verify:offline` / `verify:offline:red` / `verify:offline:wasm` — exit 0.
- `shasum -a 256 fixtures/*/expected.pdf | diff <scratch>/digests.before -` — **empty diff. A moved
  golden is a HALT.**
- `grep -c 'maximumCacheAssets = 64' folio-designer/src/release-payload.ts` — expect `1`.

**Manual checks:**
- The declared `scripts` value for each catalogue face agrees with that face's own `cmap`.
- `folio-go/internal/template/serialize.go` shows **no diff**.
- The disk-font decline is present as stated text at the control, not merely absent behaviour.

## Auto Run Result

Status: done
Blocking condition: none

### Summary of implemented change

Picking a catalogue family now writes it into the document. One new engine command, `embedFontFamily`,
decodes and bounds the bytes, hashes them to a content-addressed asset key, inserts the asset only if
that key is absent, attaches a six-key `FontRecord` (family, style, licence, **licenceText**,
**copyright**, source) and declares a chain whose head is an `AssetEntry` naming the key — one
`applyFontChainCommand` transaction, so `wasm.Engine.Apply` gives it one history entry and one undo for
free, and a re-pick short-circuits to byte-identical canonical output and pushes nothing.

The record is now **mandatory for an embedded face**: for every asset a chain names by
`{"asset": key}`, `licence`, `licenceText` and `copyright` must each be present, non-null and
non-blank, or the document is refused at load with an error located at both `assets.<key>.font.<field>`
and the chain entry `fonts.<chain>[<i>]`. An **unreferenced** font asset is untouched by the rule and
still leaves the document at `1.x`, which is what keeps D-1.4.13 intact.

DW-80 is fixed first: `assetKeyReferenced` now walks `t.doc.Fonts` over sorted keys, so the new
command-layer orphan drop (`dropUnnamedFontAssets`, called from `removeFontChainEntry` and
`deleteFontChain`) removes only a face nothing names. `serialize.go`'s `writeAssets` is untouched.

### Files changed

- `folio-go/internal/template/model.go` — `FontRecord` gains `LicenceText` and `Copyright`.
- `folio-go/internal/template/serialize.go` — six lines in `writeFontRecord` only; `writeAssets` and its
  orphan loop untouched.
- `folio-go/internal/template/parse.go` — decodes the two keys; `requireEmbeddedFaceLicence` is the new
  located, reference-scoped load refusal.
- `folio-go/component_commands.go` — DW-80's font arm; the `embedFontFamily` command; the orphan drop.
- `folio-go/internal/template/*_test.go`, `folio-go/*_test.go`, `folio-go/wasm/embed_font_test.go` — the
  refusal table, the DW-80 and retention cases, the collision refusal, the non-font-asset arm, and the
  one-history-entry/undo/redo and no-second-entry assertions.
- `fixtures/embedded-font/input.folio`, `README.md` — the one fixture invalid under the new rule,
  regenerated with terms derived from the committed `LICENSE-OFL.txt`.
- `folio-designer/font-catalogue.json` — the `scripts` field the record already specified.
- `folio-designer/scripts/build-wasm.mjs` — validates `scripts` against a closed vocabulary, ties the
  fallback faces to `shippedFamilies`, and emits a typed catalogue module carrying each face's own
  licence text and nameID-0 copyright. No new build asset.
- `folio-designer/src/App.tsx`, `App.css`, `font-chain-command.ts`, `font-chain-control.ts` — the split
  family control, `pickCatalogueFamily`, the new opaque builder and the `'embed'` refusal anchor.
- `folio-designer/src/font-catalogue.test.ts`, `App.test.tsx`, `font-chain-command.test.ts` — coverage
  against each binary's own `cmap` and `name` table, and the payload on the wire.
- `_bmad-output/specs/spec-fonts/SPEC.md`, `font-catalogue.md`, `spec-folio/folio-format.md`,
  `deferred-work.md` — the two closed questions, the amended sections, and DW-138 / DW-139.
- `.gitignore` — the generated catalogue module, beside the three generated modules already ignored.

### Review findings breakdown

Patches applied: 7 (high 1, medium 3, low 3). Deferred: 1 (low). Rejected: 10, each enumerated in the
Review Triage Log with the caller, path or population verified. No intent_gap, no bad_spec, no loopback.

Follow-up review recommendation: **true** — a `high`-severity finding was patched (the generated
catalogue was stating the wrong project's licence terms for 17 of 21 faces). Score by the formula:
3 × 3 medium + 1 × 3 low = 12, which also clears 5 on its own.

### Verification performed

All commands re-run by the workflow after the patches, at `b4885b2` with the story's edits in the tree.
Working directory recorded per D-8.4j.8.

- `folio-go/` — `go test -count=1 ./...`: **1877 pass / 2 fail / 5 skip**; the two fails are standing red
  1 by identity (`TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest, got 7 need 20).
- `folio-go/` — `go test -count=1 -run 'Serialize|Parse|RoundTrip|Asset|Font' ./internal/template/ -v`: ok.
- `folio-go/` — `go vet ./...` and `go vet -tags=matrix ./...`: both exit 0, no output.
- **repo root** — `gofmt -l folio-go lint`: `lint/internal/rules/licencegraph_test.go` and nothing else
  (standing red 2, not reformatted). Run once from `folio-go/` first, where it printed the two `lstat`
  lines that read as clean; re-run from the root, which is the measurement that counts.
- `lint/` — `go test -count=1 ./...`: 4 packages ok, uncached.
- `lint/` — `go run ./cmd/genmanifest`, then **repo root** `git diff --exit-code -- lint/MANIFEST.md`:
  exit 0. Run once from `lint/` first, where that pathspec matches nothing and exits 0 vacuously.
- `folio-designer/` — `npm run typecheck` 0; `npm run lint` 0 with **exactly 4** `only-export-components`
  (standing red 3); `npm test` **42 files / 432 tests passed** (baseline 42 / 424); `npm run build` 0 on
  node v24.16.0; `verify:offline`, `:red`, `:wasm` and `test:e2e:compile` all 0 (the last compile-only,
  DW-101).
- **repo root** — `shasum -a 256 fixtures/*/expected.pdf | diff digests.before -`: **empty. All 23 golden
  digests byte-identical**, taken before the first edit and again after the patches.
- **repo root** — `grep -c 'maximumCacheAssets = 64' …/release-payload.ts`: 1. Release manifest
  `assetCount`: 44, unchanged.
- Block-If probes: `SupportedMajor = 2` unmoved; `folio-go/internal/fontset/` untouched (no subsetting
  change); `lint/` untouched; the spec's `<intent-contract>` byte-identical to `b4885b2`.

**Mutation proofs taken by the workflow, not accepted from the report.**
- DW-80: deleting the font arm of `assetKeyReferenced` outright (not falsifying the image condition)
  reds four tests — `TestAssetKeyReferencedSeesAFontChain`, `TestAFaceASecondChainStillNamesIsRetained`,
  `TestEmbedFontFamilyRePickStoresNoSecondCopy`, `TestEmbedFontFamilyRePicksAfterTheChainWasDeleted`.
  Restored byte-identically and re-run green.
- The collision refusal: deleting the whole `if _, exists := t.doc.Fonts[name]` branch left the entire
  `folio-go` suite green — which is what routed it to patch. It now reds one named test.
- P1: the emitted catalogue module was parsed and each row compared to the `LICENSE*` in that face's own
  directory — **17 mismatches of 21 before the patch, 0 of 21 after**.

**Manual checks.** The declared `scripts` value agrees with each face's own `cmap` in both directions,
asserted per face with a non-vacuity guard. The disk-font decline is stated text at the control
(`App.tsx:1366`) and asserted by string in `App.test.tsx`. `serialize.go` does **not** show "no diff" —
see the residual risk below.

### Residual risks

1. **The `## Verification` manual check "`serialize.go` shows no diff" contradicts the Tasks**, which
   require the two new keys emitted by the hand-written writer — and that writer lives in
   `serialize.go`. Resolved by judgement in favour of Tasks, because unemitted keys break P1
   (`Parse(Serialize(d)) == d`). The constraint that actually carries the intent — *`writeAssets` is not
   touched, the orphan drop is not a serializer side effect* — holds exactly: the diff is six lines in
   `writeFontRecord`'s field list.
2. **The new load rule reaches one rule outside the font-asset record.** Because it keys off the chain
   entry, as the matrix specifies with no media-type carve-out, a chain entry naming a *non-font* asset
   must also state terms. Chosen over a carve-out, which `font/woff2` would evade. Registered in DW-138
   and now pinned by its own test — but it is the widest consequence of the format freedom and deserves
   an owner glance.
3. **A family whose chain was renamed reappears in the catalogue group, and picking it is a silent
   accepted no-op.** Rejected against matrix row 2 (every expected output holds, including "No error
   expected"), but the fix would need asset keys in `CanvasProjection`, which carries names only.
4. **DW-139 is live and owner-owned**: the produced PDF's subset carries no `name` table, so nameIDs
   0/7/13/14 do not reach the reader. Out of scope by a shipped Non-goal; the `.folio` itself carries the
   unsubsetted face, so attribution is intact in the document but not in the PDF.

### What the format freedom was spent on (for Story 15.3, per D-000.15)

`licenceText` and `copyright` are **required** on a font asset a chain names; a `.folio` embedding a
face without them is invalid at load. Chosen over an optional, absent-by-default field purely because
there are no documents to protect. Registered as **DW-138**. `SupportedMajor` stays `2` and no new
version trigger was introduced — the `{"asset": …}` entry shape already forces `2.0`.

## Delivery Log

### 2026-09-02 — planned

Planned against `8d5059f`, with HEAD advancing to `133f14a` mid-gate as the owner's ruling and
D-000.14/D-000.15 were recorded; that advance touched only the decision log and this spec, so no code
moved under the Code Map. **D-8.6.1** was the gate: the owner's own research on redistributing Google
Fonts, registered a story earlier, said a `.folio` must carry the licence *text* and the copyright —
and the format let a face travel with neither. **D-8.6.2** discharged the measurement that entry made
a precondition: the produced PDF's subset drops the font's entire `name` table, while the `.folio`
carries the unsubsetted face and keeps it. That split the obligation cleanly — this story can make the
document right and cannot make the PDF right, so the PDF half was registered rather than attempted.

**D-8.6.3 — `multiple-goals` accepted, no split.** Each candidate split (the orphan drop, DW-80's fix,
the `scripts` field, the TypeScript export) is required for the pick's own correctness. The stronger
reason was standing: this epic had already split six times, and a seventh split in the last feature
story of a run course-corrected for exactly that is not a judgement call. `oversized` was left set.
**D-8.6.4** endorsed four plan-gate calls (declare `scripts` and verify it against each face's own
`cmap`; the two explicit keys as redundancy with a purpose rather than the sole carrier; a typed
catalogue module so `src` can enumerate the faces at all; the pick reads precached bytes and is not a
fetch), and derived the version answer rather than asserting it: the `{"asset": …}` entry shape
already forces `2.0`, so there is no new trigger and `SupportedMajor` must not move.

**D-000.15's freedom was spent here on purpose.** Making licence text required is a narrowing, and
narrowings are free exactly now — already-touching, not going-looking.

### 2026-09-02 — built

Baseline `b4885b2`. **One commit, `49eb7d7`, made at Finalize — no step-03 subagent commit; instance
seven did not recur, and the count stays at seven.** Review triage: **7 patched (1 high, 3 medium,
3 low) / 1 deferred (low) / 10 rejected**, 0 `intent_gap`, 0 `bad_spec`, no loopback.

**The high is the story, and it is recorded at D-8.6.5.** The generated catalogue module keyed licence
text by SPDX identifier, so **17 of 21 faces emitted Cascadia Code's LICENSE** — reserved-name clause
included. A document embedding Inter would have travelled stating Microsoft's terms, which is worse
than carrying none: it is a false statement about the terms, naming a typeface the document is not
carrying. The root error is that *"the OFL is the OFL"* is false of the files — the SIL licence carries
a per-project preamble and reserved names. Read per face now; the bundle carries 21 texts instead of 2,
and no build asset was added. **Why it shipped green is the transferable half:** nothing observed the
generated module. The manifest was right, every binary was right, and the artifact between them — the
only thing the pick reads — was asserted `toBeTruthy()`.

The other six patches: the per-face tie of emitted text and copyright to each binary's own `LICENSE`
and name table (the gap that let the high through); the chain-name collision refusal, whose deletion
left the whole engine suite green while a pick silently overwrote a declared chain; a positive test for
the rule's reach past font assets, which had been amended *into* existing tests without one; whitespace-
only terms accepted as stated terms, on both the load and the command side; three fallback face names
tied to nothing, so a rename would have proposed a face the engine skips; and a refusal assertion whose
substring matched every row it was meant to distinguish.

**Deferred (1, low):** the family listbox owns `role="presentation"` children for its group headings,
its empty state and the disk-font decline — pre-existing at one instance, taken to four here.
**Rejected (10), enumerated with the population or path each verified:** (1) a renamed chain's family
reappears in the catalogue group and re-picking is a silent no-op — every I/O-matrix row-2 expectation
holds, and the fix needs asset keys in a projection that carries names only (**D-8.6.7**, and it is
registered rather than buried); (2) concurrent picks — both paths traced, neither yields a duplicate;
(3) the coverage reader's format-12 bound — no catalogue row reaches it, all 21 are BMP; (4) the
proposed tail naming the picked family — the three fallback families are disjoint from all 21;
(5) `style: "Regular"` hardcoded — all 21 rows are Regular by a shipped exclusion; (6) the designer
importing a gitignored generated module — the pre-existing pattern, and generation precedes every
typecheck; (7) a `cjk` value in the vocabulary against the catalogue's CJK exclusion — the key exists so
a Latin face receives a CJK *fallback*; (8) two spellings of "verbatim" — no value crosses between the
two generators; (9) DW-138/DW-139 lacking an action — one is a ledger entry, the other owner-owned;
(10) the frontmatter carrying two baselines — both are this workflow's own bookkeeping.

**D-8.6.6** endorsed the widest consequence deliberately: the rule keys off the chain entry with no
media-type carve-out, so a chain entry naming a *non-font* asset must also state terms. A carve-out
keyed on "is it a font?" would be evaded by `font/woff2` on an open media-type set. **D-8.6.8** records
the working-directory trap firing twice more in this dispatch, both times producing false clean output.

### 2026-09-02 — done

Baseline `b4885b2`, story commit `49eb7d7`, closed at `main` with nothing pushed — `origin/main` is
still `c985b9c`. **Epic 8's last feature story.**

**Gates re-measured at HEAD by the close, not relayed** — every invocation with its working directory
(D-8.4j.8). `[folio-go]` `go test -count=1 ./...` → **1877 pass / 2 fail / 5 skip** (baseline 1815;
**+62 tests**), the two failures `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest
by identity, *got 7 need >=20*; `go vet ./...` and `go vet -tags=matrix ./...` both exit 0 and silent.
**`[repo root]`** `gofmt -l folio-go lint` → `lint/internal/rules/licencegraph_test.go` and nothing
else, not reformatted. `[lint]` `go test -count=1 ./...` → *227 passed in 4 packages*, uncached.
`[lint]` `go run ./cmd/genmanifest` → rewrote `MANIFEST.md`, then **`[repo root]`**
`git diff --exit-code -- lint/MANIFEST.md` exit 0 — a no-op, and the pathspec was proved to resolve
rather than accepted as a bare exit 0. `[folio-designer]` (node v24.16.0) `npm run typecheck` 0;
`npm run lint` 0 with **exactly 4** `only-export-components`; `npm test` **42 files / 432 tests**
(baseline 42/424); `npm run build` 0 with the offline manifest at **44 assets** of 64;
`verify:offline`, `:red` and `:wasm` all 0; `test:e2e:compile` 0 — compile only, DW-101.
**`[repo root]`** the 23 golden `expected.pdf` digests hashed against a **worktree checked out at
`b4885b2`**: **byte-identical, all 23**. `maximumCacheAssets = 64` present once; `SupportedMajor`
unmoved at 2; `internal/fontset/` and `lint/` carry no diff against `b4885b2` at all.

**Heavy suites, per D-000.4's per-epic cadence: written and compiling, NOT RUN.** The four
`FOLIO_MATRIX_TARGET` legs and `TestCrossTargetByteIdentity` compile — `go vet -tags=matrix ./...` is
clean — and the Playwright suite compiles under `test:e2e:compile`. **All of them come due at Epic 8's
boundary gate**, after Story 8.4d. The D-000.4 override this story invoked for its own format- and
round-trip-shaped correctness was honoured separately and independently: the load rule was exercised
end to end through the real CLI, and a document carrying an unreferenced font asset was proved a fixed
point of parse-then-serialize.

**The close's own measurements, taken rather than accepted.** The emitted catalogue module was parsed
independently and each face's `licenceText` compared to the `LICENSE*` beside its own binary and each
`copyright` to nameID 0 read by an independent sfnt walk: **0 mismatches of 21, both fields**, with a
disequality control confirming the checker reddens when a value is altered. The 17 distinct texts
across 21 faces are four same-project sibling pairs sharing an identical upstream file, not the cache
returning. The load-bearing refusal was driven end to end: all three keys × absent / explicit `null` /
whitespace-blank are refused at load, each located at both the asset key and `fonts.body[1]`, while the
unmodified document loads clean and an unreferenced termless font asset still loads clean.
`serialize.go`'s whole diff against `b4885b2` is six lines inside `writeFontRecord`'s field list —
`writeAssets` and its orphan loop are untouched, and an orphaned font asset was round-tripped twice to
canonical byte-stability, so AD-9's P1 holds. **DW-80's deletion mutation reds FIVE tests, not four:**
the build's count was scoped to the root package and missed
`TestEngineApplyEmbedFontFamilyRePickPushesNoSecondEntry` in `wasm`. Neutering the load rule reds 16
subtests plus the non-font arm. Both mutants were restored and verified byte-identical by digest.
Both `spec-fonts` Open Questions are closed in the file: struck through with the answer, and the
disk-font one declined explicitly, rendered as stated text at the control and asserted by string.

**`followup_review_recommended: true` is discharged without a second review dispatch**, on D-8.4j.16's
rule. The profile says the layers did their job: 0 `intent_gap`, 0 `bad_spec`, no loopback, the high
found *by* the review and mutation-proved, and all 10 rejections enumerated with what each verified.
**The closer carried the scrutiny instead** — the independent per-face re-measurement, the end-to-end
refusal drive, both mutation re-runs at module scope, and the audit below.

**One defect found by the close and NOT patched, registered as DW-140.** A `.folio` may legally carry
an unreferenced font asset with no terms — the I/O matrix says so and D-1.4.13 requires it. If the
author then picks that same catalogue family, the command inserts the asset **only if absent** — correct
for dedupe by content hash — so the termless record survives, and the chain the command declares makes
it an embedded face. The transaction wrapper catches the result and refuses safely, leaving the document
untouched, but the message is the unlocated *"font chains did not pass format validation"* — the exact
failure mode patch P5 was written to prevent on the command-string path and did not reach on the
asset-reuse path. Reproduced at `456ee84`. Diagnostic quality, not correctness: no corruption, no
history entry, no bad document written.

**Registered at this close with owners:** **DW-140** the pick over an existing termless orphan (owner:
the next story touching the embed command); **DW-141** the listbox's non-option children (owner: the next
story touching the family control); **DW-142** the renamed-chain silent no-op of D-8.6.7, three reviewers
raised and rejected against the contract (owner: **whichever story next touches the canvas projection**,
since the fix needs asset keys there). **DW-138** already records what the format freedom was spent on
and the non-font-asset reach; **DW-139** the PDF attribution gap.

**What the format freedom was spent on, for Story 15.3's list (D-000.15):** exactly one thing —
`licenceText` and `copyright` are **required** on a font asset a chain names, and a `.folio` embedding a
face without them is invalid at load. Chosen over an optional absent-by-default field purely because
there are no documents to protect. Its one consequence outside the record is the non-font-asset reach.
`SupportedMajor` stays 2, no new version trigger, no other format area opened, all 23 goldens hold.

**Provenance audit (D-8.4j.12's standing step), reported by SHA.** `49eb7d7`: 32 files, all this
story's — engine, designer, the one fixture, the three spec documents, the deferred-work register and
one `.gitignore` line; nothing under `lint/` or `internal/fontset/`. Author and committer both
`Panit Wechasil <panitw@hotmail.com>`; branch `main`; both required trailers present and exact.
`456ee84` is the decision-log record commit. **No step-03 subagent commit occurred — instance seven did
not recur, and the count stays at seven.** Nothing pushed.
