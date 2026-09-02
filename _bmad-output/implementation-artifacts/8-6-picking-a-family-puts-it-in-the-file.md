---
title: 'Story 8.6: Picking a family puts it in the file'
type: 'feature'
created: '2026-09-02'
status: 'draft'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: '8d5059f8859ccb2b7c99d3fb4b16451d1793ddc9'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-fonts/font-catalogue.md'
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
warnings: [multiple-goals, oversized]
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The designer now offers twenty-one typefaces, but choosing one is currently a gesture that changes
nothing — the document comes out the same either way. This story makes the choice stick. When an
author picks a family, the typeface itself is copied into the document, and the document records
that it wants to be drawn with it. From then on the file carries the face: send it to a colleague,
open it on a machine that has never had that typeface installed, and the pages come out identical,
because the file is no longer asking the machine for anything.

Picking the same family twice does not store it twice. If a face ends up named by nothing, it is
removed rather than left to swell the file with typefaces nothing draws with — and that removal is
something the author's action does, not something the saving quietly does behind their back.

The file also records what the typeface is and whose it is: the family and style, where the bytes
came from, who holds the copyright, and the full terms under which they may be passed on — not
merely the name of a licence but its actual text. A font that travels without its terms is not a
font you may pass on, so a file that carries a typeface without them is not accepted at all; it is
refused when it is opened, and told why.

This story does not add bold, italic, or any weight axis; it does not let an author bring a typeface
in from their own disk; and it fetches nothing from the network.

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
| First pick of a family | Document declares no chain for it | One command: asset inserted under its SHA-256, `font` record written with family/style/licence/licenceText/source, chain declared with `{"asset":"<key>"}` head; document declares `2.0`; one undo entry | No error expected |
| Re-pick, same family | Asset key already in `Assets`, chain already names it | No second asset, no second chain; the existing chain is offered; canonical bytes unchanged so **no** history entry is pushed | No error expected |
| Re-pick after the chain was deleted | Asset key present, no chain names it | Chain re-declared naming the existing key; asset not duplicated | No error expected |
| Proposed fallback tail | Picked face covers Latin only | Tail is the shipped faces for the uncovered scripts; author may edit it in the chain editor | No error expected |
| Last chain entry naming a face removed | Asset becomes unnamed by any chain | Command drops that asset, scoped to the key just un-named — never a document-wide sweep | No error expected |
| Font asset still named by another chain | Two chains name one key; one removed | Asset **retained** — this is the arm DW-80 would have got wrong | No error expected |
| Chain names an asset with no licence text | Hand-written `.folio`; `font.licenceText` or `font.copyright` absent or explicitly `null` | **Load error**, located — naming the asset key and the chain entry that makes it an embedded face | Refused at load; never a warning, never a best-effort render |
| Font asset present but no chain names it | Orphan font asset, no `font` record | **Loads clean** — not an embedded face, so the requirement does not apply and the document is not raised to `2.0` | No error expected |
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
  source), and insert an `AssetEntry` into the chain. Shape it on `setComponentAsset` (`:667-770`).
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
- `_bmad-output/specs/spec-folio/folio-format.md` — document `font.licenceText` as optional, and
  record that the version trigger is unchanged.

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

## Review Triage Log

## Design Notes

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
   all. This is why one new key (`licenceText`) is sufficient and no separate `copyright` field is
   needed: the copyright, trademark and licence-description strings already travel inside the carried
   binary. That is a measurement, not an assumption.
2. **Inside the produced PDF, they are gone.** This is a real, previously-unmeasured gap against
   D-8.6.1's "including the strings inside the font binary". It is **out of this story's scope by a
   shipped Non-goal** ("no change to how the PDF producer subsets"), and acting on it would move
   every subset tag and every one of the 23 goldens. Register it; do not fix it here.

**Why the pick is not a fetch.** Story 8.5 took the bundled route: the 21 faces are already 21 of the
44 release assets, precached and content-addressed. The pick reads bytes **already on the machine**,
exactly as `runtimeAssetUrls` assets are read — which is what `font-catalogue.md` step 4 already
says. No new build asset, no new cache slot, 44 of 64 unchanged, and no middle tier is invented.

**Licence text: writer obligation, not format requirement.** The format stays permissive because
hand-written templates are first-class (FR12/S9); the *writer* always populates the record. This is
the half of D-8.6.1 the story can land without the owner. Making it **required** is the format change
that is the owner's alone — framed, not decided (see Block If).

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
