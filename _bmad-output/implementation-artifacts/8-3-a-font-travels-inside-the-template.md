---
title: 'A font travels inside the template'
type: 'feature'
created: '2026-08-31'
status: 'done'
baseline_revision: 'f51dd5e4c7f8a993eb0b735496bf44c8164191df'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: ['oversized']
deferred:
  - summary: >-
      A font chain entry may name an asset that is not a font (e.g. an image/png
      asset), and nothing refuses or reports it at any surface today.
    evidence: |-
      decodeFontChainEntry checks only that the {"asset": key} is PRESENT in the
      assets map, never that the asset is a font. Refusing it at load would
      violate D-1.8.1 (as amended) — an unrecognised font media type is preserved
      at load and errors at render — so the load behaviour is correct as shipped.
      But chainFaceNames drops every embedded entry before face resolution, so it
      never errors at render either. The render-surface half becomes reachable
      only when Story 8.4 resolves an embedded entry to bytes; that is the story
      that must produce the located render error for a non-font asset.
    location: >-
      folio-go/internal/template/parse.go (decodeFontChainEntry)
    severity: medium
  - summary: >-
      An empty face name in a chain is now a load error, narrowing what already-
      shipped 1.x documents load, and the narrowing is not version-gated.
    evidence: |-
      Before this story decodeStringArrayRaw accepted "" as a chain entry, so a
      1.0 document containing {"fonts": {"body": [""]}} loaded. decodeFonts now
      refuses it at fonts.<name>[<i>]. The refusal is deliberate and tested (the
      "an empty face name" case), and an empty name could never resolve to a
      face — but it is a load narrowing applied to documents declaring an older
      version, which the I/O matrix's "Existing corpus" row cannot observe: that
      row measures only the 22 golden PDF digests. Precedent is DW-era
      canvasFontFamilies, which was recorded as deferred rather than reversed.
    location: >-
      folio-go/internal/template/parse.go (decodeFonts)
    severity: medium
  - summary: >-
      TestTheFontRecordCostsAnExistingDocumentNothing rests on a hand-maintained
      fixture count that no code derives.
    evidence: |-
      The assertion is `withAssets != 7`, and its own failure message concedes the
      number must be edited by hand whenever any fixture gains or loses a non-empty
      assets map. A fixture added without touching this line reduces the population
      silently rather than failing loudly. Same family of defect as DW-81 (a
      comment claiming a derivation the code does not perform), which this story
      closed elsewhere.
    location: >-
      folio-go/embedded_font_fixture_test.go
    severity: low
---

<intent-contract>

## Intent

**Problem:** A `.folio` cannot carry a font face. `Fonts` is `map[string][]string`
(`internal/template/model.go:147`) and `decodeFonts` (`parse.go:313-328`) routes every chain through
`decodeStringArrayRaw`, so a chain entry can only ever name a face the renderer already ships. A
document that wants any other typeface is not a contract between an author and an integrator — it is
an install instruction. This is the story that changes the format, and a shipped format version is
irreversible.

**Approach:** Add one new chain-entry shape — the one-key object `{"asset": "<key>"}` — resolved
against the existing top-level `assets` map, which already stores bytes keyed by the lowercase hex
SHA-256 of the decoded bytes with `data` base64 hard-wrapped at 76 columns. A font asset differs from
an image asset only in its `mediaType` and in carrying an additive `font` record (family, style,
licence, source). Storage, load, round-trip, located load errors and the canvas projection only —
rendering **from** an embedded face is Story 8.4.

## Boundaries & Constraints

**Always:**
- **Part 0 gates every other task.** AC5 is an ordering requirement: `folio-format.md` records the
  settled version rank **before any code lands**. The rank is settled below; Part 0 writes it down,
  it does not decide it.
- **`mediaType` stays an OPEN set at the format level.** D-1.8.1 (as amended) is binding and its own
  note predicted this recurrence "later for font formats". The lead already ruled this collision at
  run setup (`epic-7-8-decision-log.md:70-74`): *unrecognised media type is preserved at load and
  errors at render; a recognised type whose bytes do not decode stays a load error, as does a chain
  entry naming an absent asset key.* `TestClosedSetsNeverIncludeMediaType`
  (`closedsets_test.go:81-99`) enforces the mechanical half and fatals on a media-type-shaped key in
  `closedsets.go`.
- **`Validate` predicts `Render`, never a second rule system** (D-1.8.1 amended, binding).
- **The `assets` mechanism is reused whole, and nothing is added to it.** Key rule, 76-column split,
  dedup by key and emission order are unchanged. No new asset-key derivation site (see D-8.1.3
  measurement in Design Notes N3).
- **The projection's entry-shape validator moves in THIS commit** (D-8.2.8). Go and TypeScript land
  together or the canvas goes permanently blank with no attributable error.
- **`-count=1` on every Go gate.** A cache is not an anchor (D-7.9.5).
- A comment asserting a negative carries a test's evidentiary burden and must name the population it
  measured.

**Block If:**
- Any task appears to require **rendering text from an embedded face**, or shaping/subsetting one.
  That is Story 8.4. HALT rather than widen.
- The plan comes to believe **any of the 22 recorded golden digests moves**. A moved digest
  invalidates a human attestation by construction (D-4.7.1). HALT — do not re-record.
- Settling the version rank appears to require **moving `SupportedMajor` off 2**. It does not (see
  Design Notes N1); if it seems to, the reading is wrong. HALT.
- A named consumer is found that must **branch** on a font-specific diagnostic code. Minting one is a
  registry-policy act above a builder's authority (D-7.8.1). HALT.

**Never:**
- Never add an authored chain-key order (D-8.1.1). `fonts` is a mapping with no authored key order;
  `writeFonts` sorts keys and `writeObject` sorts again.
- Never build Story 8.4's rendering, 8.5's catalogue, or 8.6's pick-and-embed command / orphan
  collection.
- Never touch `rawNumberLiteral` or consolidate the designer's encoders (Story 15.2a).
- Never touch `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)` or its drift twin.
- Never write `reader`, `date` or `examined` in `fixtures/statement-signoff.json` or
  `fixtures/thai-stacked-marks/signoff.json`.
- Never `git add -A` / `git add .`; never modify, move or commit root `README.md`; never push; never
  branch.
- Never reuse `assetKeyReferenced` (`component_commands.go:790`) for a font asset — it walks image
  elements only and returns false for every font asset, so any orphan check built on it deletes live
  faces with no compile error. (No orphan check is in scope here; this fences the trap.)

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Embedded entry round-trips | `fonts.body = ["Noto Sans", {"asset":"<64hex>"}]`, asset present, `mediaType: "font/ttf"` | Parses; `Serialize(Parse(b)) == b`; entry re-emitted as a one-key object through `writeObject` | No error |
| Font `font` record round-trips | asset carries `font{family,style,licence,source}` | All four keys parsed into the model and re-emitted in sorted key order | No error |
| Dedup + order | two chains name the same asset key | One `assets` entry; `assets` key order unchanged; no image entry moves | No error |
| Absent asset key | `{"asset":"<key>"}`, key not in `assets` | Located load error naming **chain, entry index and key** | `LoadError`, field `fonts.<name>[<i>].asset`, code `TEMPLATE_FIELD_INVALID` |
| Bad entry shape | entry is a number, an array, or an object with keys other than exactly `asset` | Located load error naming chain and index | `LoadError`, field `fonts.<name>[<i>]` |
| Recognised font type, wrong bytes | `mediaType: "font/ttf"`, bytes are not sfnt | Located load error naming the asset | `LoadError`, field `assets.<key>.data` |
| Unrecognised media type | `mediaType: "font/woff2"` | **Loads clean.** Document is valid; asset preserved verbatim | Error only at Render, and only when a render needs the face (D-1.8.1 amended) |
| Version raised by content | document has an `{"asset":…}` entry | `versionRequiredByContent` returns `2.0` | No error |
| Version NOT raised | document has a font asset but no `{"asset":…}` entry | Stays at whatever its other content requires | No error |
| Existing corpus | any of the 22 goldens (none embeds a font) | Bytes unchanged; recorded digest still matches | No error |

</intent-contract>

## Code Map

**The version surface (Part 0's landing sites), measured at `83ab8c8`**
- `folio-go/internal/template/version.go:63-66` — `SupportedMajor = 2`, `SupportedVersion = "2.0"`.
  **Neither moves in this story.** `:88-93` the four version constants (`baseVersion` 1.0,
  `minorFeatureVersion` 1.1, `keepTogetherVersion` 1.2, `majorFeatureVersion` 2.0); `:261-268`
  `versionRank` iota; `:279-284` `versionForRank`. **No new constant and no new rank are needed** —
  an embedded-face entry maps onto the existing `rankMajorFeature`, which sidesteps D-7.7.2's
  renumbering guardrail entirely.
- `folio-go/internal/template/version.go:221-250` `versionRequiredByContent` — takes the **maximum**
  (`:192-197`), walks the three bands' elements only. `fonts` and `assets` are **not walked at all**,
  so the font trigger is a **new document-level probe outside the band loop** — the first probe in
  this function that is not per-element.
- `folio-go/internal/template/linespacing_test.go:311` `TestContentVersionNeverExceedsTheLibraryCeiling`
  — its comment at `:300-302` claims the bound is "derived from the constants themselves"; it is
  **hand-enumerated**. `:328` and `:342` are two further hand-enumerated builder loops. A font
  trigger needs a **third builder loop** or it is silently never checked.
- `folio-go/internal/template/version_test.go:22-30` `TestHigherMajorIsLoadError` hardcodes `"3.0"`;
  `:36-51` `TestSupportedMajorIsLoadable` hardcodes `"2.0"`; `:208-235`
  `TestVersionForRankIsStrictlyAscending` hand-enumerates the four ranks at `:229-232`. All three stay
  green because the ceiling does not move.
- `_bmad-output/specs/spec-folio/folio-format.md:67-72` the mandatory MINOR/MAJOR sentence; `:74-79`
  the "this has happened once" precedent paragraph; `:47` the `version` field row (the content ladder,
  which gains a rung); `:124-135` the `fonts` section (**states no entry-shape rule at all today** —
  string-ness is implied only by the example); `:408-432` the `assets` section (**no field table, no
  `mediaType` closed-set claim**; `mediaType` is absent from the eight-set list at `:68-70`).

**The asset mechanism being reused (read-only unless noted)**
- `folio-go/internal/template/model.go:37-38` `Assets map[string]Asset`, doc comment "keyed by
  lowercase hex SHA-256 of the raw bytes"; `:338-345` `Asset{Data []string; MediaType string; Extra
  []Field}`. **JSON keys are literals in the parser/serializer, not struct tags** (`parse.go:342,352`;
  `serialize.go:466-467`). **`Asset` gains one optional field: the `font` record.**
- `folio-go/internal/template/assetkey.go:14-25` `isSHA256HexKey`; `:29-32` `sha256HexOf` — the
  canonical derivation, on **decoded** bytes.
- `folio-go/internal/template/base64.go:55-71` `splitBase64Canonical`, `const width = 76` at `:60`
  (function-scoped, unexported); sole caller `serialize.go:463`. `:20-30` `decodeBase64Asset` —
  concatenates elements then `base64.StdEncoding.Strict()`. `:8-10` and `:47-49` state the canonical
  rule. **All reused verbatim; none edited.**
- `folio-go/internal/template/serialize.go:455-475` `writeAssets`, sorting at `:457`; doc comment
  `:433-454` (the 76-column re-wrap rule and the unconditional orphan-preservation rule).
  `:38-42` `writeObject`'s `sort.Slice` — **the mechanism AC2 rests on** (see N2).
- `folio-go/internal/template/parse.go:330-408` `decodeAssets` — the ten located refusals, key-shape
  check at `:368`, base64 decode `:375`, emptiness `:379`, **digest verification `:386-390`** (the
  loader verifies, it does not trust), image byte-sniff `:400-402` gated on `recognised`.
  `:392-399` is the comment that states D-1.8.1 (amended) in code. **Extended, not rewritten.**
- `folio-go/internal/template/image.go:96-115` `recognisedMediaTypes` / `decodeRecognisedImage` — the
  **shape to copy** for fonts, and its comment at `:98-104` is the binding statement that a media type
  set here is a *library capability* set and may never join `closedsets.go`. `:78-95`
  `UnsupportedMediaTypeError` and `:128-137` `DecodeImageForRender` — the render-time precedent.

**The entry shape, both sides — these land in ONE commit**
- `folio-go/internal/template/model.go:147` `Fonts map[string][]string`; `:150-171` `Fonts.Chain`,
  whose doc comment names the five callers that must not re-implement the emptiness test. **The
  element type changes; all five call sites are typed on `[]string`.**
- `folio-go/internal/template/parse.go:313-328` `decodeFonts` — the whole refusal is
  `decodeStringArrayRaw` at `:321-324`, which collapses the array into one error with field path
  `fonts.<name>` and **no entry index**. AC3's index does not exist today and must be threaded.
- `folio-go/internal/template/serialize.go:174-184` `writeFonts` → `:60-77` `writeStringArray`, which
  **can emit only JSON strings**. Needs a per-entry writer that routes an embedded entry through the
  existing `writeObject`. Note `writeFonts`' own comment already records why it sorts.
- `folio-go/page_setup.go:433-443` `CanvasFontChain{Name string; Entries []string}`; `:453-479`
  `canvasFontChains` (`slices.Clone(entries)` verbatim, no per-entry transformation); `:481-490`
  `canvasFontFamilyNames`; `:294` and `:287` the projection fields; `:39`
  `maxCanvasPropertyString = 512`; `:431` `maxCanvasFontFamilies = 256`;
  `component_commands.go:1918` `maxCanvasFontChainEntries = 64`.
- **`folio-designer/src/engine-protocol.ts:254-256` — THE VALIDATOR THAT MUST MOVE.** The clause
  `chain.entries.every((face) => typeof face === 'string' && face.length > 0 && face.length <=
  MAX_CANVAS_PROPERTY_STRING)` rejects an object entry outright; `hasExactKeys(chain,
  ['name','entries'])` at `:255` is an **exact** set, symmetric in both directions. `:165` the
  `fontChains` type; `:89-91` and `:211-212` the in-source statements of the consequence.
- `folio-designer/src/engine-client.ts:86-87` `#fail('PROTOCOL_INVALID', …)`; `:118-125` detaches and
  calls `this.worker.terminate()` — **no restart path**. A one-sided change is a permanently blank
  canvas with no element id and no attributable error.
- `folio-go/canvas_projection_wire_test.go:44-68` `canvasProjectionWireKeys`; `:70-76`
  `canvasFontChainWireKeys = []string{"entries","name"}`; `:103-123` the Go half; `:176`/`:180` the
  regexes that extract the TS guard's own lists; `:184-214` the TS half. **Measured gap: it records
  key NAMES only, never value types or nesting depth** — changing `Entries []string` to a struct
  slice leaves it green while blanking the canvas. A third, entry-level recorded key set plus a
  matching TS extraction is the analogous tripwire and is this story's to add.
- `folio-designer/src/engine-bounds-mirror.test.ts:80-81` — ties Go's bounds to the TS constants by
  matching the **literal source text** `/chain\.entries\.length <= MAX_ENGINE_FONT_CHAIN_ENTRIES\b/`.
  Rewriting the entries validator must preserve that expression or update the mirror.
- `folio-designer/src/FontChainEditor.tsx:26-31` — the forward note written by Story 8.2, verbatim:
  *"THE DISPLAYED TEXT OF AN ENTRY IS THE PROJECTED ENTRY, UNMODIFIED … when Story 8.3 gives an entry
  a richer shape the author reads whatever the projection then carries, with no browser-side rule to
  update."* Render at `:131-132`. **`:74`'s change signature `chain.entries.join(' ')` would
  produce `[object Object]` on a richer entry** — it must become value-based over the new shape, or
  the focus logic silently breaks.
- `folio-designer/src/canvas-authority-contract.test.ts:65-69,99-103,167-178` — scans the panel for
  engine refusal vocabulary. The browser may not derive family/style itself; it must be projected.

**Command path — read-only in this story**
- `folio-go/component_commands.go:2194-2220` `moveFontChainEntry` and `:2222-2242`
  `removeFontChainEntry` are **index-only and shape-agnostic** — they work unchanged on a richer entry
  type. `:2045-2054` `fontChainFace` → `:1271-1281` `commandString` accepts a non-empty JSON string
  only, and `:2069-2117` `addFontChain` unmarshals `entries` as `[]string`. **Both correctly remain
  unable to express an embedded entry — that is Story 8.6's command.** `:2027-2043`
  `declaredFontChain` returns `[]string`; its signature follows the model.
- `folio-go/component_commands.go:2266-2277` `fontChainReferences` walks **both** `Element.Style` and
  `Element.Table.Value.HeaderStyle` (comment `:2255-2265` records the measurement). Unchanged here.

**Read-only evidence**
- `fixtures/*/expected.pdf` is **22** files; `goldenDigestRecord` (`byte_neutrality_test.go:92`) is 22.
  `matrixDocuments` is 23 (it includes `hidden-image`, which ships no `expected.pdf`) — **do not derive
  the digest count from it**. Of 28 fixture dirs, 24 carry an `input.folio`, all 24 declare `"assets"`,
  and only **6** have a non-empty map (`component-asset-import` 2, `image-embed`, `statement-1/5/20/50`
  1 each). Only `image/png` and `image/jpeg` appear anywhere under `fixtures/`.
- `folio-go/internal/template/drift_test.go:232` / `:269` — bidirectional, no allowlist. The
  `fonts`/`assets` exclusion is by **immediate-parent depth** (`:152`,`:166`), so inside an `assets`
  fence the SHA key becomes the parent and its children ARE registered: **`font`, `family`, `licence`
  and `source` each need a backticked/fence mention in `folio-format.md`** (`style`, `data`,
  `mediaType` already exist on both sides). `asset` is already registered from prose at
  `folio-format.md:273`.
- `folio-go/internal/template/goldenfixture_test.go:77` `TestAssetExampleMatchesGoldenFragment` —
  extracts the `` ## `assets` `` fence and compares byte-for-byte to
  `folio-go/testdata/template/golden/asset-example.json`, then re-verifies key == SHA-256(decoded) and
  the 76-column split. **Editing that fence obliges editing the golden fragment in the same commit.**
- `folio-go/internal/template/closedsets_test.go:81-99` `TestClosedSetsNeverIncludeMediaType` — fatals
  if `closedsets.go` gains a media-type-shaped key. `assets_test.go:121` / `:134` are AC11a/AC11b.
- `folio-go/internal/template/omitempty_test.go:134,154` — nil `Assets` must serialize as `{}`.
- `folio-designer/src/property-prose-height.test.ts:41-42` pins `property-value-prose` to exactly one
  `<textarea>` and one `<input>`. oxlint baseline is exactly **4** `only-export-components` warnings.
- Baseline red set at `83ab8c8` is expected to be **one**: `TestCorpusMeetsP6ExerciseFloors` /
  `P6g_(opaque_names)`. Re-measure rather than assume — this number has moved twice in this run.

## Tasks & Acceptance

**Execution:**

**Part 0 — the format record. No other task may start until this is written.**

1. `_bmad-output/specs/spec-folio/folio-format.md` — record the settled rank and the entry-shape rule.
   Under `fonts` (`:124-135`), state the **two** legal entry shapes, that any other shape is a load
   error naming the chain and the index, that a `{"asset":…}` key absent from `assets` is a load error
   naming chain, index and key, and that chain-key order remains unauthored. Under `assets`
   (`:408-432`), add the `font` record (`family`, `licence`, `source` — and `style`, already
   registered) and state **explicitly that `mediaType` remains an OPEN set** for fonts exactly as for
   images, citing D-1.8.1 (amended): an unrecognised font media type loads clean and errors at render.
   Update the `version` row at `:47` to add the new rung: `2.0` if any style sets `align: "justify"`
   **or any chain declares an embedded-face entry**. Extend the `:74-79` precedent paragraph to record
   this as the second MAJOR-class extension joining the same `2.0`. **Every new key must appear
   backticked or in a fence or `TestDriftDocToGo` fails.** If the `` ## `assets` `` fence is edited,
   update `folio-go/testdata/template/golden/asset-example.json` byte-identically in the same commit.
2. `_bmad-output/specs/spec-fonts/SPEC.md` and `format-changes.md` — close the open question ("MINOR or
   MAJOR `.folio` version bump?") with the settled answer and the decisions it derives from, and
   **correct `format-changes.md:56`'s "Closed set" claim for `mediaType`**, which contradicts binding
   D-1.8.1 (amended). Leave the other open questions alone.

**Part 1 — model, parse, serialize.**

3. `folio-go/internal/template/model.go` — give `Fonts` an entry type carrying either a shipped face
   name or an asset key, and add the optional `font` record to `Asset`. Update `Fonts.Chain`'s
   signature and its doc comment's list of callers. Keep the record's fields as `Presence[T]` where the
   surrounding style uses it, so absence and explicit `null` stay distinguishable
   (`presence.go:16-20`). **An explicit JSON `null` is not absence**: a refusal written only in the
   non-null branch lets `"font": null` and `"family": null` past every guard, so each new optional key
   needs its null case asserted as well as its missing case.
4. `folio-go/internal/template/parse.go` — `decodeFonts` (`:313-328`) walks the chain array itself
   rather than delegating to `decodeStringArrayRaw`, so the **entry index reaches the field path**.
   Refuse: a non-string, non-object entry; an object whose key set is not exactly `{asset}`; a
   non-string or empty asset value; and an asset key absent from `assets`. Field paths
   `fonts.<name>[<i>]` and `fonts.<name>[<i>].asset`. **`decodeAssets` must run before `decodeFonts`**
   or the absent-key check has nothing to consult — state the ordering in a comment. Extend the
   `mediaType` handling by adding a **font** recognised-type predicate beside
   `decodeRecognisedImage` at the `:400-402` call site, so a *recognised* font type whose bytes are not
   sfnt is a load error and an *unrecognised* one is never inspected. Parse the `font` record.
   **All refusals use `newLoadError`** — no code is minted (N4).
5. `folio-go/internal/template/fontasset.go` (new) — the font media-type capability set and the sfnt
   structural check, mirroring `image.go:96-137`'s shape: a `recognisedFontMediaTypes` set that is
   deliberately **not** in `closedsets.go`, a decode predicate, and an `UnsupportedMediaTypeError`
   analogue for the render surface. Carry `image.go:98-104`'s comment forward in substance, naming the
   population it measured.
6. `folio-go/internal/template/serialize.go` — `writeFonts` (`:174-184`) gains a per-entry writer that
   emits a string entry exactly as today and routes an embedded entry through `writeObject`. Emit the
   `font` record from `writeAssets` (`:455-475`). **Do not touch `splitBase64Canonical`, the sort at
   `:38-42`, or `writeAssets`' key sort at `:457`.**
7. `folio-go/internal/template/version.go` — `versionRequiredByContent` (`:221-250`) gains a
   document-level probe over `d.Fonts` for an embedded-face entry, raising to the **existing**
   `rankMajorFeature`. No new constant, no new rank. Update `:39-48`'s doc comment to record the second
   reason `2.0` exists. Note in the comment that the probe is outside the band loop and why.

**Part 2 — the projection, both sides, one commit.**

8. `folio-go/page_setup.go` — `CanvasFontChain.Entries` becomes a slice of a new projected entry
   struct carrying the discriminated shape plus, for an embedded entry, the **family and style read
   from the asset's `font` record**. Apply `maxCanvasPropertyString` to every projected string and keep
   `maxCanvasFontChainEntries` enforced at the same site. Update the type's doc comment.
9. `folio-designer/src/engine-protocol.ts` — move the entry validator at `:254-256` to accept the new
   shape and reject everything else, and extend the `CanvasProjection` type at `:165`. **Preserve the
   literal expression `chain.entries.length <= MAX_ENGINE_FONT_CHAIN_ENTRIES`** or update
   `engine-bounds-mirror.test.ts:80-81` deliberately.
10. `folio-go/canvas_projection_wire_test.go` — add the **entry-level** recorded key set and a TS-side
    extraction for it, closing the measured gap that the existing nested record covers key names only.
    The failure text must say what it says for the chain object: a field added on one side only blanks
    the canvas.
11. `folio-designer/src/FontChainEditor.tsx` — render an embedded entry as its projected **family and
    style**, and a string entry exactly as today. Replace `:74`'s `entries.join(' ')` change
    signature with one that is value-based over the new shape. Update the `:26-31` comment to what is
    now true. **Derive nothing in the browser** — family and style come from the projection.

**Part 3 — evidence.**

12. `fixtures/embedded-font/` (new) — `input.folio` declaring a chain with an embedded-face entry
    whose bytes are an **already-shipped face from `folio-go/fonts/`** (no new binary enters the repo),
    plus a README stating what it red-proves and what it does **not** cover (rendering from the face is
    Story 8.4). **Ships no `expected.pdf`**, on the `hidden-image` precedent, because this story cannot
    render from an embedded face — so `goldenDigestRecord` stays at **22** and AC6 is structural.
    Register in `matrixDocuments` (`matrix_test.go`), the `docs=` slug list and all four
    `hash.<target>.<slug>.txt` upload blocks in `.github/workflows/matrix.yml`
    (`matrix_registration_test.go:40` cross-checks these two automatically), the subprocess const +
    `TestMain` arm in `render_test.go`, and `declaredEpic2GateObligations`
    (`byte_neutrality_test.go:814`) with an inline comment naming this story. Confirm
    `expected.json`'s `goToolchain` matches `matrixDocuments[0]`'s or `assertFixturesShareToolchain`
    fatals. If the fixture is text-bearing, update **both** `missing_glyph_corpus_test.go` declarations
    together (`beyondBaselineAcceptance`, never `baselineAcceptanceFixtures`).
13. `folio-go/internal/template/fonts_embedded_test.go` (new) — the I/O matrix, every row. Round-trip
    fixed-point both directions (`Parse(Serialize(d)) == d` and `Serialize(Parse(b)) == b`). Each
    located refusal asserted on its **exact field path including the index**, red-proved by mutation
    per site and in both directions — a green suite over correct code proves nothing: revert the
    guard, watch the test go red; restore, watch it pass. The **unrecognised-media-type positive control** — a font
    asset with `mediaType: "font/woff2"` loads clean — is the arm that proves D-1.8.1 (amended) was
    honoured, and it is the one a "closed set" implementation would fail.
14. `folio-go/internal/template/linespacing_test.go` — add the **third** builder loop to
    `TestContentVersionNeverExceedsTheLibraryCeiling` for the font trigger, and assert both directions:
    a document with an embedded-face entry requires `2.0`, and one carrying only a font **asset**
    (no `{"asset":…}` entry) does **not**. **Mutate the enumeration, not only the predicate** — both
    arms of a probe can be proved while the whole loop is silently never walked.
15. `folio-go/internal/template/closedsets_test.go` — extend `TestClosedSetsNeverIncludeMediaType`'s
    population, or add a sibling, so it also refuses a **font**-media-type-shaped key. The comment
    asserting "font media types are not a closed set" must name the population it measured.
16. `folio-designer/src/engine-protocol.test.ts` and `folio-designer/src/App.test.tsx` — the new entry
    shape is accepted; a malformed entry (unknown key, missing `asset`, a number, an object with an
    extra key) still returns `undefined`; the panel renders family+style for an embedded entry and the
    raw name for a string entry; the change-signature fix is proved by a snapshot in which two entries
    differ only in a field `join` would have flattened.
17. `_bmad-output/implementation-artifacts/deferred-work.md` — record the wire-test gap now closed, and
    file: `assetKeyReferenced`'s blindness to font assets (a data-loss trap for 8.6), and
    `linespacing_test.go:300-302`'s comment claiming a derivation the code does not perform.

**Acceptance Criteria:**
- Given the whole change, when `shasum -a 256 fixtures/*/expected.pdf` is run, then it prints **22**
  lines byte-identical to a baseline captured **before the first edit**, and `goldenDigestRecord` still
  holds 22 entries.
- Given Part 0, when `git log -p` for this commit is read, then the `folio-format.md` version rank and
  the entry-shape rule are present in the **same commit as** or an **earlier commit than** any code
  change — AC5 is an ordering requirement, not a documentation task.
- Given a `.folio` carrying a font asset with an unrecognised `mediaType`, when it is loaded, then it
  loads clean and `Validate` is silent — proved by a positive control, not inferred.
- Given the Go projection and the TypeScript guard, when either side's entry shape is changed alone,
  then `canvas_projection_wire_test.go` goes red — proved by making the one-sided change, not inferred.
- Given `closedsets.go`, when the whole change is inspected, then no font media type appears in it and
  `TestClosedSetsNeverIncludeMediaType` passes.
- Given each of the three located load errors, when the guard producing it is reverted, then exactly the
  test asserting it goes red — per site, both directions.
- Given the six fixtures that already carry a non-empty `assets` map, when they are re-serialized, then
  their bytes are unchanged, proving the `font` record's absence costs nothing.

## Spec Change Log

## Review Triage Log

### 2026-08-31 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 9: (high 2, medium 4, low 3)
- defer: 3: (high 0, medium 2, low 1)
- reject: 6
- addressed_findings:
  - `[high]` `[patch]` `chainFaceNames` (`folio-go/render.go`) is the story's ONE
    behavioural boundary between 8.3 and 8.4, and nothing tested it: deleting
    `if entry.Embedded() { continue }` left the whole Go suite green. Its comment also
    claimed both consequences were "pinned by test … (fonts_embedded_test.go)", a file in
    `package template` that structurally cannot call `Render`. Added
    `folio-go/chain_face_names_test.go` in `package folio` pinning the drop, order
    preservation, an embedded entry in FIRST position not shadowing the named face behind
    it, and the located all-embedded-chain error at BOTH `Render` and `Validate` with the
    two messages asserted equal. Comment corrected to cite the test that now exists.
    Re-proved after the fix: the same deletion reddens
    `TestChainFaceNamesDropsEmbeddedEntriesAndKeepsOrder`.
  - `[high]` `[patch]` `projectFontChainEntry` (`folio-go/page_setup.go`) had no Go test:
    deleting the `out.Family = entry.AssetKey` fallback left the suite green, while any
    font asset carrying no `font.family` would then project `family: ""`, which
    `isFontChainEntry` rejects → `parseInbound` undefined → `worker.terminate()` →
    permanently blank canvas. Exactly the D-8.2.8 / DW-82 failure mode, shipping green.
    Added `folio-go/canvas_font_chain_entry_test.go` driving the real `Canvas()` entry
    point over 8 record shapes plus the `Family`/`Style` bound arms. Re-proved after the
    fix: the same deletion reddens
    `TestProjectedEmbeddedEntryCarriesTheDiscriminantAndTheRecord`.
  - `[medium]` `[patch]` `checkSfnt` returned nil for a `ttcf` tag on the false ground that
    "no recognised media type reaches here with it anyway" — `mediaType` is author-declared,
    so `font/ttf` over TTC bytes was recognised, reached the check and loaded clean, against
    the format text's promise that a file lying about itself is a load error. `ttcf` moved
    out of the single-face tag set into its own named refusal, with the coupling note that
    the refusal must become conditional if `font/collection` ever joins the recognised set.
  - `[medium]` `[patch]` `_bmad-output/specs/spec-fonts/format-changes.md` still described
    `font.licence` as "Required on every font asset", contradicting `folio-format.md`'s
    "every key inside it is optional" and `TestPlainFontAssetNeedsNoRecord`. The story had
    corrected the `mediaType` row two lines above and left this one. Corrected to optional.
  - `[medium]` `[patch]` `page_setup.go` cited `canvas-authority-contract.test.ts` as
    enforcing that `FontChainEditor` "holds no rule of its own"; that file never names
    FontChainEditor and enforces one narrower rule (no engine refusal vocabulary, by
    directory walk). Both comments rewritten to the enforcement that exists, and
    `FontChainEditor.tsx` now concedes that `entryLabel` branches on `assetKey` and composes
    family+style — the claim is authority, not absence of code.
  - `[medium]` `[patch]` The `font` record's per-key presence handling was exercised only
    for `family`, though the spec's own argument ("a refusal written only in the non-null
    branch lets `"family": null` past every guard") is per-key. Table-driven over a
    `fontRecordKeys` list bound to `FontRecord` by reflection, so a fifth key cannot arrive
    uncovered; all four keys now get present / explicitly-null / empty-string / absent arms
    plus wrong-type refusals.
  - `[low]` `[patch]` `decodeFonts`' chain-level refusal still read "must be an array of
    strings", untrue now an entry may be an object, and unpinned because the test asserted
    only the Field. Message corrected and the reason pinned, with an explicit guard against
    the old wording reappearing.
  - `[low]` `[patch]` `maximalFixture`'s font asset declared `"source": "hand-built 32-byte
    sfnt"` for bytes that decode to 156 — and `font.source` is defined as where the bytes
    came from. Corrected.
  - `[low]` `[patch]` `collectBandTableRuns` (`folio-go/table_render.go`) called
    `chainFaceNames` before checking `if !ok`, doing needless work on the not-found path and
    turning `chain` from nil into a non-nil empty slice before the error branch. Moved below
    the check.


## Design Notes

**N1 — The version rank: MAJOR, `2.0`, and the library ceiling does not move.**
The rank is **settled from shipped decisions**, so no halt is warranted. Three of them, in order of
authority:
- **D-R7.9 (owner decision, 2026-08-30)** names this story and this version literally: *"Epic 7 takes
  the MAJOR bump to `.folio` 2.0 freely at Story 7.3, Epic 8's format change joins the same 2.0 at
  Story 8.3"*, and *"Story 8.3 joins the same 2.0."* It also dissolves the only argument against —
  the lead's pre-story finding (`epic-7-8-decision-log.md:64-68`) worried that *"Story 15.3 is about
  to cut `folio-go/v0.1.0`, so a 2.0 document would be unreadable by the first tagged release
  forever."* D-R7.9 answers exactly that: *"We haven't released anything to production yet so no need
  to tag now … Story 15.3 is not imminent and must not be treated as a deadline."* **The premise of
  the objection is gone, and the owner ruled after it.**
- **D-7.3.1's pre-reader test** confirms it independently: *would a pre-V reader refuse it or render
  it wrong?* A 1.x reader meets an object entry in `decodeStringRaw` (`decodehelpers.go:51-57`, which
  "never coerces") and **refuses the file** — D-1.4.12's stated mechanism exactly. Declaring anything
  below 2.0 would be, in D-7.7.2's words, *"a version that lies"*: it would claim a reader sufficient
  for content that reader cannot load.
- **D-1.4.13** makes `version` a property of the **document**. So the trigger is the **entry shape**,
  not the presence of a font asset: a document carrying a font asset that no chain references loads
  and renders correctly on a 1.x reader (the `font` record rides through as passthrough under
  D-1.4.9, and `mediaType` is only required to be a present string), and therefore must **not** be
  raised. AC-tested both directions in Task 14.

**`SupportedMajor` stays 2.** D-R7.9's "joins the same 2.0" means the ceiling is already where it
needs to be; the story adds a **second reason** a document declares 2.0, mapping onto the existing
`rankMajorFeature`. That is why no new version constant and no new `versionRank` member are needed,
and why D-7.7.2's `versionForRank` renumbering guardrail is never engaged. `TestHigherMajorIsLoadError`'s
hardcoded `"3.0"` and `TestSupportedMajorIsLoadable`'s `"2.0"` both stay green untouched.

**N2 — AC2's guarantee, measured.** *"`assets` emission order is unchanged, so adding a font never
moves an image"* is structural, not incidental, and it holds at **two** independent layers:
`writeAssets` sorts its own keys (`serialize.go:457`, `slices.Sorted(maps.Keys(assets))`) and the
single generic object emitter sorts again (`serialize.go:38-42`,
`sort.Slice(fields, …fields[i].key < fields[j].key)`). `writeObject` is the emitter for every object at
every level, so an asset's byte position is a **total function of its key**. Since the key is the
SHA-256 of the bytes and each entry is an independent `kv` written by its own closure, inserting one
entry can only shift entries **after** it in key order, and reproduces everything before it
byte-for-byte. The same holds inside one asset object: adding `"font"` sorts between `"data"` and
`"mediaType"` and moves neither's content. **Dedup** is likewise free — `Assets` is a Go map keyed by
the digest, so two chains naming the same face are one entry by construction, with no dedup pass to
write. This is measured, and it is why AC2 needs no new code at all.

**N3 — D-8.1.3 applied: the asset mechanism's "single authority" is THREE things, and only one of
them is sole.** Verified rather than assumed, per the standing rule:
- **Key derivation is NOT sole.** `sha256HexOf` (`assetkey.go:29-32`) is the canonical definition, but
  `component_commands.go:743-744` re-derives it inline (`sha256.Sum256` + `fmt.Sprintf("%x", …)`)
  because package `folio` cannot see the unexported helper. The **shape** check is duplicated too:
  `isSHA256HexKey` (`assetkey.go:14-25`) and `isAssetKeyShape` (`asset_bytes.go:41-52`), the latter's
  comment acknowledging the copy. **This story adds no fourth site** — it stores no asset itself; a
  font asset arrives by being loaded or, later, by Story 8.6's command. The measurement is recorded so
  8.6 does not add one, and so no task here claims a soleness that does not exist.
- **The base64 wrapper IS sole**, in both directions: `splitBase64Canonical` (`base64.go:55-71`) has
  exactly one caller, `writeAssets` (`serialize.go:463`), and `decodeBase64Asset` (`base64.go:20-30`)
  is the one decoder. Reused untouched.
- **Digest verification IS sole and it verifies rather than trusts** — `parse.go:386-390` compares
  `sha256HexOf(decoded)` to the key and refuses a mismatch. A font asset gets this for free.
So AC1's *"the same asset mechanism images already use, with no second storage shape"* is honoured by
**not writing an asset path at all**: the only additions are one optional `font` object inside the
existing record and one recognised-media-type predicate beside the existing image one.

**N4 — No diagnostic code is minted, and the consumer test is why.** D-7.8.1 replaced `diag.go`'s
reservation with a rule: *"The general code is the default. A specific code is minted only when a named
consumer must BRANCH on it to behave differently."* `newLoadError` (`errors.go:257-259`) already
supplies `TEMPLATE_FIELD_INVALID`, which survives the wasm host's `reportableMessage` destruction
(only `TEMPLATE_MALFORMED` is replaced), so all three located load errors reach the author in words
with no mint. **No consumer branches:** the designer's diagnostic presenter is deliberately
code-agnostic and does one thing — locate, name the field, show the value and the reason — and it
discriminates on `Field`, which now carries the chain name **and the entry index**. Minting
`FONT_ASSET_MISSING` would buy it nothing `Field` does not already give. Task 4's field paths are what
make that true, so they are load-bearing rather than cosmetic.

**N5 — D-8.1.2's standing check, answered explicitly.** **This story walks no `style.X`.** Its version
probe is document-level over `d.Fonts`, outside `versionRequiredByContent`'s band loop, so
`headerStyle` is reached vacuously — there is nothing on `headerStyle` for a `fonts` probe to walk.
Two related facts, measured, so the check is answered rather than dodged: `versionRequiredByContent`
**already** walks both `el.Style` and `el.Table.Value.HeaderStyle` (`version.go:229-238`) and keeps
doing so unchanged; and `fontChainReferences` (`component_commands.go:2266-2277`) already walks both
attachment points for rename/delete safety. The asymmetry that remains is **pre-existing and untouched
here**: `page_setup.go` projects no `headerStyle` at all (zero occurrences), so a table header's chain
is authoritative in Go and invisible to the designer. Recorded, not fixed.

**N6 — Where the family/style display lands, and why it is HERE and not 8.4.**
**Here.** Three measured reasons:
1. The projection change is **forced into this commit regardless** (D-8.2.8): `engine-protocol.ts:256`'s
   `typeof face === 'string'` rejects an object entry, `parseInbound` returns `undefined`,
   `engine-client.ts:121` terminates the worker, and the canvas is permanently blank with no
   attributable error. Since the entry's projected shape must change anyway, projecting the family and
   style **on** it is marginal work, not a second feature.
2. The alternative ships a release in which the panel displays a **64-character hex digest** to the
   author. `FontChainEditor.tsx:26-31` forbids the browser from doing anything about that ("no key
   detection, no parsing"), and `canvas-authority-contract.test.ts` enforces it — so the panel could
   not paper over it even if it wanted to. Deferring produces a knowingly-wrong UI.
3. Story 8.4 is about **rendering** from an embedded face — a different surface with a different
   acceptance shape. Putting a panel-display obligation there would repeat exactly the mismatch
   D-8.2.8(b) was written about.
This discharges Story 8.2's forward obligation in full: 8.2 delivered the negative half (an entry
displays unmodified) and named 8.3 for the positive half.

**N7 — What this story deliberately does NOT do, and the test that must say so.** After this story a
chain may contain an embedded entry that nothing can render: `resolveRuneFace` (`render.go:1163-1176`)
looks entries up in the supplied `FontSet` and **silently skips** an absent one, and a chain left with
no usable entry produces the existing located render error. That is the honest interim state — Story
8.4 makes it render — but it is a negative assertion, so per Epic 7's boundary record it **carries a
test's evidentiary burden**: Task 13 must pin the interim behaviour and name the population, rather
than leave it to a comment. If pinning it appears to require implementing the resolution, that is the
`Block If` firing.

## Verification

**Commands** (run from the repo root unless stated; `-count=1` on every Go gate, D-7.9.5):
- Capture the digest baseline **before the first edit**: `shasum -a 256 fixtures/*/expected.pdf > <scratch>/digests.before` -- expected: **22** lines.
- `cd folio-go && go test -count=1 ./...` -- expected: exactly **one** distinct failure, `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)` (mandated permanent red; never "fix" it). Re-measure the baseline red set at `83ab8c8` in a detached worktree rather than assuming this number — it has moved twice in this run.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean.
- `gofmt -l folio-go` -- expected: empty. (`lint/…/licencegraph_test.go` is a known pre-existing gofmt offender, DW-23, outside this path.)
- `cd folio-go && go test -count=1 -tags=matrix ./...` -- **the overdue full matrix sweep.** `TestShippedFacesReproduceFromUpstream` needs `fontTools`; report explicitly whether it **FAILED** or **SKIPPED**, never folded into a count. This is the story that touches font bytes, so this sweep is not optional.
- `cd folio-go && for T in darwin/arm64 linux/amd64 linux/arm64 js/wasm; do FOLIO_MATRIX_TARGET=$T go test -count=1 -tags=matrix -run TestTargetRenderHash -v .; done` -- expected: four legs, each asserting. Then **an unset control**: `cd folio-go && go test -count=1 -tags=matrix -run TestTargetRenderHash -v .` -- expected: a deliberate no-op pass. Report all five.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` -- expected: pass.
- `cd folio-go && go test -count=1 -tags=matrix -run TestThaiStackedMarksSemanticSignOffIsRecorded .` -- expected: pass (this is NOT the same test as `TestGoldenDigestAgreesAtEveryDeclaredSite`).
- `cd lint && go test -count=1 ./...` -- expected: pass. **`-count=1` is mandatory**: the rules package walks the `folio-go` tree with `ReadDir`, which Go's test cache does not track, so a cached `ok` here is no measurement at all. `TestFloatTypedTestScopeInventory` pins five sites **by line number** in `shaping_expectations_test.go` and `internal/fontset/{fontset,vendorboundary}_test.go` — if this story shifts lines there, it reddens.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` -- expected: typecheck clean; lint at exactly **4** pre-existing `only-export-components` warnings; tests **at or above the 319 / 35-files baseline** — report the actual numbers; e2e compiles (it is compile-only and executes nowhere by design).
- `shasum -a 256 fixtures/*/expected.pdf | diff - <scratch>/digests.before` -- expected: **empty diff**, 22 lines. This is AC6.
- `cd folio-go && GOOS=js GOARCH=wasm go test -count=1 -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" ./wasm/cmd/engine/` -- run **only if** a located load error's message must be shown to reach the author; this package is invisible to `go test ./...` and "it compiles" is not evidence.

**Manual checks:**
- `git status --porcelain` before committing -- expected: only this story's paths. Stage **explicit paths only**; never `git add -A` / `git add .`. Root `README.md` must be untouched (md5 `078d7d80d518d54af2fc04fb270d46b8`, 8,470 bytes).
- Confirm neither `fixtures/statement-signoff.json` nor `fixtures/thai-stacked-marks/signoff.json` appears in the diff.
- Confirm the Part 0 documentation change is in the same commit as, or an earlier commit than, any code change (AC5's ordering requirement).

## Auto Run Result

Status: done
Blocking condition: none

Implementation dispatch. Baseline `f51dd5e4c7f8a993eb0b735496bf44c8164191df` on `main`; one local
commit, nothing pushed, no branch created.

**Summary.** A `.folio` can now carry a font face. `Fonts` becomes `map[string][]FontChainEntry` — a
discriminated entry that is either a shipped face name or the one-key object `{"asset": "<key>"}`
resolved against the existing top-level `assets` map. `Asset` gains one optional `font` record
(`family`, `style`, `licence`, `source`), each field `Presence[string]` so absence and explicit
`null` stay distinguishable. `decodeFonts` walks each chain array itself so every refusal carries the
ENTRY INDEX (`fonts.<name>[<i>]`, `fonts.<name>[<i>].asset`) — plumbing that did not exist before, not
a formatting detail. `mediaType` remains an OPEN set: bytes that do not decode as their declared font
media type are a located load error, while a media type this build does not recognise is preserved at
load and errors only at render (D-1.8.1 as amended). The format version joins the existing `2.0` via
the existing `rankMajorFeature`; `SupportedMajor` stays 2 and no constant or rank was added. The
projection's entry-shape validator moved in the same commit on both sides, carrying family and style
so the panel never shows the author a 64-character digest. Rendering FROM an embedded face remains
Story 8.4; `chainFaceNames` is the single boundary where this story stops.

**Files changed** (38 files, one commit):
- `_bmad-output/specs/spec-folio/folio-format.md` — Part 0: the settled MAJOR rank, the two legal
  entry shapes, the located refusals, and `mediaType` stated explicitly as an OPEN set for fonts.
- `_bmad-output/specs/spec-fonts/SPEC.md`, `format-changes.md` — the MINOR-or-MAJOR open question
  closed; the incorrect "Closed set" claim for `mediaType` and the "Required" claim for `font.licence`
  both corrected.
- `folio-go/internal/template/model.go` — `FontChainEntry`, `FontRecord`, `Fonts.Chain` retyped.
- `folio-go/internal/template/parse.go` — `decodeAssets` before `decodeFonts`; per-entry walk with the
  index in the field path; the font byte check at the existing call site.
- `folio-go/internal/template/fontasset.go` (new) — the recognised-font-media-type capability set
  (deliberately NOT in `closedsets.go`), the sfnt structural check, the render-surface error.
- `folio-go/internal/template/serialize.go` — per-entry writer routing an embedded entry through the
  existing `writeObject`; the `font` record emitted from `writeAssets`.
- `folio-go/internal/template/version.go` — a document-level probe outside the band loop.
- `folio-go/page_setup.go`, `render.go`, `table_render.go`, `component_commands.go` — the projected
  entry struct, and the one filter where 8.3 stops.
- `folio-designer/src/engine-protocol.ts`, `FontChainEditor.tsx` — the entry guard and family/style
  display; the bounds-mirror literal preserved.
- `folio-go/canvas_projection_wire_test.go` — the entry-level recorded key set (DW-82 closed).
- `fixtures/embedded-font/` (new) — ships NO `expected.pdf` on the `hidden-image` precedent.
- New tests: `fonts_embedded_test.go`, `embedded_font_fixture_test.go`, `chain_face_names_test.go`,
  `canvas_font_chain_entry_test.go`.
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-81 and DW-82 closed; DW-80 left open.

**Review findings breakdown.** 9 patched (2 high, 4 medium, 3 low), 3 deferred, 6 rejected, 0
intent_gap, 0 bad_spec. Both high findings were verification gaps that a green suite was hiding:
deleting the 8.3/8.4 boundary filter, and deleting the projection's family fallback, each left the
entire Go suite green before the patch and each reddens a named test after it.

**Follow-up review recommendation: true.** Patched counts: high 2, medium 4, low 3. Any high patched
finding sets this true on its own; the score `3 x 4 + 1 x 3 = 15` is also >= 5.

**Verification performed** (every Go gate with `-count=1`, D-7.9.5):
- 22 golden digests captured BEFORE the first edit and diffed after: EMPTY DIFF, 22 lines. AC6 holds.
- `go test -count=1 ./...` — exactly ONE distinct red, `TestCorpusMeetsP6ExerciseFloors` /
  `P6g_(opaque_names)` (got 7, need >=20), the mandated permanent red.
- `go vet -tags=matrix ./...` clean; `gofmt -l folio-go` from the repo root empty.
- Full `go test -count=1 -tags=matrix ./...` sweep (overdue, now measured):
  `TestShippedFacesReproduceFromUpstream` **FAILED — it did not skip** ("fontgen: fontTools is not
  importable by this interpreter"). Confirmed pre-existing by running it in a detached worktree at
  `f51dd5e`, where it fails identically; the test fails rather than skips by deliberate design, and
  `folio-go/fonts/` is untouched by the diff.
- `TestTargetRenderHash` four legs with `FOLIO_MATRIX_TARGET` exported — darwin/arm64 1.465s,
  linux/amd64 8.069s, linux/arm64 5.252s, js/wasm 11.239s — plus the unset control at 0.377s, a
  deliberate no-op. The timing contrast is the evidence the legs actually asserted.
- `TestCrossTargetByteIdentity` ok 23.227s; `TestThaiStackedMarksSemanticSignOffIsRecorded` ok.
- `cd lint && go test -count=1 ./...` — all four packages ok.
- Designer: typecheck clean; oxlint exactly 4 pre-existing warnings; 323 tests / 35 files (baseline
  319 / 35); e2e compiles.
- `GOOS=js GOARCH=wasm` `wasm/cmd/engine` ok.
- Mutation proofs re-run independently, each asserting its anchor occurrence count before editing:
  the entry index proved with a failing entry NOT first (an always-0 index reddens the `[1]` and `[2]`
  cases while `[0]` stays green); the projection validator proved BOTH ways (reverting to the pre-8.3
  `typeof face === 'string'` rule reddens 16 tests, and malformed entries are still rejected after).
- No closed set of font media types was created: `closedsets.go` is untouched by the commit and
  contains no font media type; all three closed-set tests pass. Nothing was minted: no diag file
  appears in the commit and no new code constant appears anywhere in the diff.

**Residual risks.**
- `TestShippedFacesReproduceFromUpstream` cannot pass in this environment (no `fontTools`), so the
  shipped-face reproduction claim is unverified here — pre-existing, and now measured rather than
  assumed.
- Three deferred items are recorded in frontmatter; the first bears directly on Story 8.4, which is
  where a non-font asset named by a chain entry must produce a located render error.
- DW-80 (`assetKeyReferenced` blind to font assets) remains open and owned by Story 8.6. Nothing
  collects orphans today, so it is not reachable; no collector was added.
