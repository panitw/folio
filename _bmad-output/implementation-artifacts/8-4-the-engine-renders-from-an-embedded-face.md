---
title: 'The engine renders from an embedded face'
type: 'feature'
created: '2026-08-31'
replanned: '2026-09-01'
status: 'done'
baseline_revision: '15ca0ddbc4565d935fde026bfbad463be8ddd182'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: ['oversized', 'multiple-goals']
deferred:
  - summary: >-
      fixtures/embedded-font/expected.pdf is a Thai-bearing golden that no human
      has read, and it is now the only one in the corpus without a reading
      attestation.
    evidence: |-
      Story 8.4 recorded this fixture's first expected.pdf, and its page is Thai
      drawn from the embedded face. Every OTHER Thai-bearing golden carries a
      human reading sign-off declared as a gate obligation (shaped-text's
      thai-signoff.json under D-2.3.5; thai-stacked-marks' signoff.json under
      Story 8.0), and this one carries none. No agent may write one.

      IT IS FILED AS LOW RATHER THAN AS A HALT, on a measurement rather than on
      convenience: the drawn string is สัญญา, which is byte-for-byte the string
      fixtures/thai-stacked-marks/ draws as its e2 ZERO-OFFSET CONTROL, at the
      same 12pt, shaped by the SAME face — folio-go/fonts/notosansthai/
      NotoSansThai-Regular.ttf, whose bytes this document merely carries a second
      copy of. So the shaping and the reading of these five glyphs are already
      attested by the owner (2026-08-31); what is unattested is this particular
      subset and page. The full -tags=matrix sweep leaves NO sign-off gate red
      and invalidates no existing attestation — checked, not assumed — so AC6's
      three grounds for closing without a new sign-off all hold, and this entry
      exists so the gap is disclosed rather than discovered.
    location: >-
      fixtures/embedded-font/expected.pdf
    severity: low
  - summary: >-
      The new Thai golden's sign-off obligation exists only as a register
      entry; unlike every other attested golden it has no red gate that a
      person must clear.
    evidence: |-
      Every other attested fixture pairs its attestation with a failing
      //go:build matrix gate (folio-go/shaped_signoff_matrix_test.go,
      thai_stacked_marks_signoff_matrix_test.go,
      statement_signoff_matrix_test.go) and a goldenDigestRecord site of
      {kind: "signoff"}. fixtures/embedded-font/'s record uses neither, so the
      obligation is held only where someone has to go looking for it -- the
      condition D-2.3.5 calls "an obligation nobody trips over".

      It is NOT actionable inside this story: the spec's Block If halts the
      story if a //go:build matrix sign-off gate would be left red, so minting
      the gate here would force the halt it is meant to prevent. It needs a
      ruling that authorises the gate together with the human reading.
    location: >-
      folio-go/byte_neutrality_test.go (goldenDigestRecord, embedded-font entry)
    severity: low
  - summary: >-
      A readable embedded entry that no rune ever resolves to still contributes
      its metrics to the vertical model, so line advance can differ from
      pre-8.4 for such a document.
    evidence: |-
      fontCache.metricsFace treats any embedded name as declared and decodes
      it, so chainLineMetrics/chainVerticalModel now see the carried face even
      when nothing draws with it. Measured: only fixtures/embedded-font/ has an
      embedded chain entry, so no committed golden can observe the difference
      and all 22 pre-existing digests are unmoved -- but an integrator's
      document with a carried face it never draws from could see its line
      advance change across this version boundary. Nothing pins the intended
      answer either way.
    location: >-
      folio-go/render.go (fontCache.metricsFace) and folio-go/wrap.go
    severity: low
  - summary: >-
      A chain whose ONLY entry is an unreadable embedded one is untested; the
      error it produces today is right by ordering rather than by assertion.
    evidence: |-
      Both new tests that exercise a non-font chain entry
      (TestNonFontAssetDrawnErrorsAtRenderAndAtValidate and
      TestChainOfOnlyUnusableEntriesProducesTheExistingLocatedError) place the
      failing entry BEHIND a shipped face that supplies the vertical model. A
      probe of the single-entry case returns the located capability error
      rather than the generic empty-metrics error, so the behaviour is correct
      today -- but which of the two errors wins depends on evaluation order
      that no test fixes.
    location: >-
      folio-go/chain_face_names_test.go
    severity: low
  - summary: >-
      Two error branches the code itself documents as structurally unreachable
      are untested.
    evidence: |-
      predictDocument's `if cache.isEmbedded(name)` arm is commented
      "Unreachable in practice: a name only reaches this loop by having already
      been parsed", and embeddedFaceSource.decode's `if !s.present` arm is
      "structurally unreachable today". Neither is exercised, and an untested
      unreachable error path is the one that is wrong when a later change makes
      it fire. This repo's own idiom elsewhere is an unreachability tripwire
      that reddens when the branch becomes reachable.
    location: >-
      folio-go/render.go (predictDocument) and folio-go/embedded_face.go
    severity: low
---

<intent-contract>

## Intent

**Problem:** A `.folio` can carry a font face (Story 8.3) but nothing can draw with one.
`chainFaceNames` (`render.go:1137`) drops every embedded entry before face resolution, so an
embedded face contributes nothing to coverage, nothing to metrics and no font program to the PDF.
Three consequences follow. FR54's promise — a template renders on a machine that has never seen its
faces — is undelivered. A chain entry naming a **non-font** asset (an image, say) is accepted at
load, correctly under D-1.8.1 as amended, but errors **nowhere** at render either, so a document
naming an image as a font renders **silently wrong** (DW-83; D-8.3.5 makes this an acceptance
criterion, not a note, because the `Validate` half has no user-visible symptom when missing).
And the canvas inherits the drop **twice over**: it cannot measure with a face the PDF will use,
and — once this story makes it measure — it has **no CSS family for an embedded face at all** and
falls through to `sans-serif`, which is the owner-reported defect fixed at `c6e4d03` rebuilt for this
story's headline use case (DW-35; **D-8.4.1** rules it this story's, because this story *creates* the
condition rather than inheriting it).

**Approach:** Resolve an embedded chain entry to bytes on the render path, from the document's own
`assets` map, reusing the image precedent verbatim (`render.go:1873-1936`): dedupe asset keys, walk
them in sorted order, `DecodeAssetBytes` → `DecodeFontForRender` → `fontset.New`, carrying the
element id for located errors. `DecodeFontForRender` and `UnsupportedFontMediaTypeError` were built
by Story 8.3 and are deliberately unwired (`internal/template/fontasset.go:80-94`); this story is
their call site. Because the single `predictDocument` derivation (`render.go:1625`) is what both
`Render` and `Validate` call, wiring the resolution there makes AC4's `Validate` half true **by
construction** rather than by a second rule system. Because `CanvasWithTextPaint` calls the same
`fontChain` (`page_setup.go:1162`), the canvas measures with the same face for the same reason — and
that sharing is **asserted by a test rather than assumed**, so a later story cannot quietly fork it.
For the paint half, the face's bytes reach the browser through the projection and are registered as a
`FontFace` under a CSS family name **derived from the asset key** (D-8.4.1), parallel to what
`c6e4d03` already built for the shipped faces.

**On the chain-lookup duplication — the layer matters, because it changes the repair.**
`chainFaceNames` **is** the single boundary: `table_render.go:665` *calls* it, so that is a second
caller, not a second conversion. What **is** duplicated is `fontChain`'s **lookup and its error**,
hand-mirrored at `table_render.go:653-660` under a comment that says so verbatim. So: **extract the
LOOKUP, not the filter.** The real remaining seam is the four documentless
`(chain []string, fs FontSet)` consumers — the set `chainFaceNames`' own comment names as the thing
it exists to avoid widening. **Verify all of this in the tree before building on it; it is a finding
to confirm, not a premise to accept.**

## Boundaries & Constraints

**Always:**
- **Load continues to ACCEPT a chain entry naming a non-font asset.** D-1.8.1 as amended: an
  unrecognised or wrong-kind media type is preserved at load and errors at render. Do **not**
  tighten `decodeFontChainEntry`. The load behaviour is correct as shipped.
- **Resolution is by asset key, never by name.** AD-8: an embedded face and a shipped face sharing a
  family name never substitute for one another. The asset key decides.
- **An embedded face's CSS family name is derived from its ASSET KEY, never from `font.family`**
  (**D-8.4.1**, the design decision Story 8.2's Design Note N3 escalated, now made by the lead).
  `font.family`/`font.style` are display identity, *never used to resolve or substitute a face*.
  Deriving the CSS family from `font.family` would let an embedded "Inter" collide with a shipped
  "Inter" in the browser's font registry — AD-8's own hazard, one layer down.
- **AC3's tag is a hash of the emitted subset PROGRAM BYTES, not of the glyph set** (**D-8.4.2**).
  The epic text has been corrected. AD-7 names the glyph-set reading and rejects it; D-1.5.8 rejected
  it after two instances collided. The old wording implemented literally moves all 22 digests.
- **`fixtures/embedded-font/`'s document must be changed to draw text the embedded face COVERS**
  (**D-8.4.4b**). It currently draws Latin while carrying a Thai face, so its digest cannot observe
  the embedded face being used and AC6 asserts nothing. Re-recording that one fixture's digest is
  deliberate and correct **within this story** — it is the story that owns the fixture. This is the
  single sanctioned exception to the "every existing golden renders byte-identically" bullet above;
  the other 21 must not move.
- **Every existing golden renders byte-identically.** The 22 recorded digests are the assertion.
  Capture `shasum -a 256 fixtures/*/expected.pdf` **before the first edit** and diff after.
- **`Validate` predicts `Render` by sharing `predictDocument`, never by re-implementing.** Every new
  check lands inside `predictDocument`'s reach. `TestValidateNeverReachesRenderOrInternalPDF`
  (`render_arch_test.go:326`) is the guard that makes adding it anywhere else awkward on purpose.
- **Subsetting stays once per render inside the PDF producer** (`render.go:1834`, the only non-test
  `.Subset(` call site in the module). No save-time subsetting is introduced.
- Commits land on `main`. Never push, never branch. Stage explicit paths; never `git add -A`.
- The repo-root `README.md` is the owner's file (md5 `078d7d80d518d54af2fc04fb270d46b8`): never
  modified, moved, deleted or staged. `fixtures/statement-signoff.json` and
  `fixtures/thai-stacked-marks/signoff.json` are human attestations; no agent writes their
  `reader`, `date` or `examined` fields.

**Block If:**
- **The canvas PAINT question (DW-35) is RULED — it is NOT a Block If any more, and it is IN SCOPE.**
  **D-8.4.1**: the paint half belongs to this story, because this story *creates* the condition
  rather than widening it (before it, an embedded face cannot render at all, so no author can reach
  the state). The design decision that blocked it — the CSS family name — has been **made** by the
  lead and is an Always bullet above. **Do not re-open it and do not route it out.**
  **Sizing is this gate's call:** if the paint half genuinely makes this two shippable deliverables,
  return `multiple-goals` and say so — it then splits to a story sequenced **immediately after 8.4**,
  not "later in Epic 8", and this story's record must state the canvas limitation **explicitly**, so
  it is disclosed rather than discovered. Splitting on size is legal; dropping it silently is not.
- A new human sign-off record is required — i.e. an existing attestation is invalidated, or a
  `//go:build matrix` sign-off gate would be left red. Never write `reader`/`date`/`examined`.
- Closing the tag derivation onto the glyph set (see Design Notes, "AC3's false premise").

**Never:**
- Never mint a diagnostic code unless a **named consumer must branch on it** (D-7.8.1). The image
  precedent (`render.go:1921`) returns a bare wrapped error with prose location and no code; follow
  it unless a consumer is named.
- Never add a font media type to `closedsets.go`. It is a **capability** set, not a format set, and
  `TestClosedSetsNeverIncludeMediaType` (`internal/template/closedsets_test.go:81`) fatals on it.
- Never absorb **DW-80** (`assetKeyReferenced` blind to font assets). It is Story 8.6's.
- Never "fix" either standing red (see Verification).
- Never edit `fixtures/thai-break-corpus/corpus.json`. The P6 tests read only that file; a new
  golden fixture does not touch them.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Mixed chain, embedded face covers | `["Noto Sans", {"asset": K}]`, K a `font/ttf` asset; text needs a rune only K covers | Rune shapes from K, in declared order; K's program embedded in the PDF; subset tag derived from the emitted program | No error expected |
| Shared family name | K's `font.family` == a shipped face name in the FontSet | The asset key decides; no name-based substitution either way | No error expected |
| FontSet is the shipped set alone | `fonts.Shipped()` only; document names an embedded face | Renders from the document's own bytes; no network, no host font, no disk path read | No error expected |
| Non-font asset, drawn | Chain entry names an `image/png` asset, and something must draw from it | **Render errors, located**, naming the chain, the entry index and the asset key | `UnsupportedFontMediaTypeError` via `DecodeFontForRender` |
| Non-font asset, same document, `Validate` | Identical inputs to the row above | **`Validate` returns the identical error** — same string, zero diagnostics | Must match Render byte-for-byte |
| Non-font asset, never drawn | Chain entry names an image, but no element resolves a rune to that entry | Loads and renders clean — the error fires only when something must actually draw it | No error expected |
| Load, non-font asset | Any of the above, at `LoadTemplate` | **Accepts** — unchanged from Story 8.3 | No error expected |
| Four targets | The embedded-face fixture on `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` | Byte-identical output; one recorded hash | No error expected |
| Canvas MEASURES with the embedded face | A component whose chain resolves a rune to an embedded entry | The canvas paint projection's metrics come from the engine's own `fontChain`/`shapeSegments`/`chainVerticalModel` — **asserted by a test**, so the path cannot be forked later | No error expected |
| Canvas PAINTS with the embedded face | Same component, in the browser | The face's bytes reach the browser through the projection and register as a `FontFace` under a family derived from the **asset key**; the fragment rule asks for that family | No error expected |
| Embedded and shipped share `font.family` | Embedded asset K has `font.family: "Inter"`; a shipped face is also "Inter" | The two get **distinct** CSS family names, because the embedded one's derives from K's asset key — no registry collision | No error expected |
| The embedded-face fixture draws covered text | `fixtures/embedded-font/` after this story | The document draws runes the embedded face **covers**, so the digest observes the face being used; that one digest is re-recorded deliberately | No error expected |

</intent-contract>

## Code Map

**Verified at `de87bef` on re-plan. Where a ruling asked for confirmation, the measurement is stated.**

### The boundary, and the layer that was wrong the first time

- `folio-go/render.go:1109-1136` `chainFaceNames`' doc comment — **it IS the single boundary**, in its
  own words: *"chainFaceNames is THE one boundary between the document's chain and the render path's
  face-name list, and it is where Story 8.3 stops."* **D-8.4.4(a) confirmed in the tree.**
- `folio-go/table_render.go:666` — `chain = chainFaceNames(entries)`. A second **caller**, not a
  second conversion. There is **no `entry.Embedded()` anywhere in `table_render.go`**; the only
  non-test `Embedded()` sites are `render.go:1140`, `page_setup.go:547`,
  `internal/template/version.go:361`, `internal/template/serialize.go:213`.
- **What IS duplicated is the LOOKUP and its error**, and the comment admits it verbatim —
  `table_render.go:656-659`: *"Mirrors fontChain's own error, verbatim in shape (render.go) — a text
  element with the same defect fails the same way, plain-wrapped, no `*RenderError`."*
  `render.go:1102` `doc.doc.Fonts.Chain(chainName)` vs `table_render.go:654`
  `doc.doc.Fonts.Chain(hs.fontFamily)`; error text duplicated at `render.go:1104` vs
  `table_render.go:660` (which hand-prefixes the `folio: Render: element %s:` wrapping its caller
  otherwise applies). **A second message is duplicated too**: `render.go:1099` *"has text but no
  style.fontFamily to resolve a font from"* vs `table_render.go:688`. **So: extract the LOOKUP.**
- `folio-go/render.go:1097-1107` `fontChain` — three parts: the `el.Style…FontFamily` presence check
  with its own error, the `Fonts.Chain` lookup with its error, and `return chainFaceNames(chain), nil`.
  **Only the middle part is duplicated.** It and `collectBandTableRuns` (`table_render.go:602`) are
  the **only two functions holding a `*Template` at the lookup**, i.e. upstream of the `[]string`
  narrowing.

### The seam — and a correction to the "four" that both the contract and the code comment carry

`chainFaceNames`' comment (`render.go:1133-1136`) names four consumers it exists to avoid widening:
*"resolveRuneFace/chainLineMetrics/shapeSegments/formatFontChain … those four would each have to
answer the same question, and four answers is how they come to differ."*

**Measured: `chain []string` is a parameter of TEN non-test functions, not four.** The comment's list
is neither the population nor the risk set — one of its four takes no `FontSet` at all, and three
that DO take `FontSet` + `*fontCache` are **not named**:

| # | file:line | `FontSet`? | `*fontCache`? | named in the comment? |
|---|---|---|---|---|
| 1 | `render.go:1202` `resolveRuneFace` | yes | yes | yes |
| 2 | `render.go:1223` `formatFontChain` | **no** | no | yes |
| 3 | `render.go:1232` `missingGlyphMessage` | no | no | no |
| 4 | `render.go:1292` `shapeSegments` | yes | yes | yes |
| 5 | `page_number.go:436` `digitTableRun` | yes | yes | **no** |
| 6 | `wrap.go:550` `chainLineMetrics` | yes | yes | yes |
| 7 | `wrap.go:592` `verticalModel` | no | no | no |
| 8 | `wrap.go:699` `scaleAdvanceByLineSpacing` | no | no | no |
| 9 | `wrap.go:750` `chainVerticalModel` | yes | yes | **no** |
| 10 | `wrap.go:761` `lineAdvance` | yes | yes | **no** |

**SIX take `(FontSet, *fontCache)` and are the real seam** (1, 4, 5, 6, 9, 10). **None of the ten has
any access to a `*Template` / `*template.Document`**, hence none can reach `Assets`. The four
non-`FontSet` ones (2, 3, 7, 8) consume the chain only for **messages and vertical arithmetic**, so
they need the names but never the bytes — that asymmetry is what makes a per-render name→bytes view
attractive over widening six signatures. **Whichever seam is chosen, correct the comment's "four" in
the same commit** (D-8.2.3: a hole in one arm of an enumeration is evidence about the enumeration).

### Render-time resolution — the precedent and the built-but-unwired door

- `folio-go/render.go:1873-1936` — the image loop, the shape to mirror: dedupe asset keys
  (`:1873-1888`), `firstElementIDByAssetKey` for location (`:1882-1887`), sorted deterministic order
  (`:1889`), presence check naming element+key (`:1906-1912`), `DecodeAssetBytes` (`:1913`),
  `DecodeImageForRender` (`:1919`), **validate unconditionally, embed only if visible**
  (`:1924-1928`). Returns a bare `fmt.Errorf("folio: Render: %w", derr)` at `:1921` — no
  `*RenderError`, no code, prose location.
- `folio-go/internal/template/fontasset.go:25-38` `UnsupportedFontMediaTypeError{AssetKey,
  ElementID, MediaType}`; `:86-94` `DecodeFontForRender`; `:71-78` `decodeRecognisedFont`
  (`font/ttf`, `font/otf` only); `:137-193` `checkSfnt`. **Zero non-test call sites**; comment
  `:80-85` names this story as the one that gives it one. **The error names neither the chain nor
  the entry index AC4 requires — extend it (D-8.4.3: additive, and message text is safe under
  AD-14 because callers match the code, never the text).**
- `folio-go/internal/template/base64.go:41-43` `DecodeAssetBytes`. Not memoized anywhere.
- `folio-go/render.go:1148-1181` `fontCache` — `map[string]*fontset.Font`, **keyed by bare face
  name, which is also the `FontSet` key** (`data, ok := fs[name]` at `:1171`). *"Looked up and
  written only by key — NEVER ranged (ScanMapRange, D-2.2.3)."* `fontset.New` is called at `:1175`
  and that is the **only** path into the cache. **An embedded face needs a distinct key or it
  collides in this namespace**; `pagemodel.TextRun.Face` and `pdf.EmbeddedFace.Name` are both
  documented as "the caller's FontSet key", so the choice propagates.
- `folio-go/internal/fontset/fontset.go:85` `New(name string, data []byte)` — takes bytes, does not
  care where they came from. The key affordance.
- `folio-go/page_setup.go:545` `projectFontChainEntry` — the **existing precedent** for resolving an
  embedded entry against `t.doc.Assets` (`:550`).

### `Validate` — why AC4's third half is free, and the trap that would drop it

- `folio-go/validate.go:52` `Validate`; doc `:17-19` *"a DRY-RUN PREDICTOR of Render, not a second
  rule system … NO RENDER IS ATTEMPTED"*; `:22-24` shares its implementation via `predictDocument`;
  it calls it directly at `:68`.
- `folio-go/render.go:1625` `predictDocument` — the single shared derivation. `Render` reaches it via
  `renderDocument` (`:1529`) → `buildPageModel` (`:1606`). **The image decode at `:1919` is already
  inside it**, so `Validate` already predicts that error — **but no test asserts it**
  (`render_image_test.go:258` asserts `Render` only). **Do not copy the image test.**
- `folio-go/chain_face_names_test.go:176-211` `TestAllEmbeddedChainProducesTheExistingLocatedError`
  — **the shape to copy**: load accepts → `Render` errors → `Validate` errors identically, with
  `verr.Error() != rerr.Error()` string equality at `:206` and `len(diags) != 0` at `:209`.
- `folio-go/render_arch_test.go:326` `TestValidateNeverReachesRenderOrInternalPDF` — the AST guard
  that makes a `Render`-only placement structurally visible.

### Subsetting

- `folio-go/render.go:1818-1836` — `glyphsByFace` union `:1818-1824`, deterministic face order
  `:1826`, `font.Subset(...)` at `:1834` (**the only non-test `.Subset(` call site in the module**),
  `pdf.EmbeddedFace` `:1843-1859`. Once per render, once per face. **No save-time subsetting exists.**
- `folio-go/internal/fontset/fontset.go:910-924` `deriveTag` — FNV-1a over the **complete emitted
  subset program bytes**. Doc `:885-909` records the glyph-set reading as rejected. AC3 is now
  corrected in the epic (**D-8.4.2**); **AD-7** names and rejects the glyph-set reading
  (*"unlike a hash of the sorted glyph-id set alone"*) and **D-1.5.8** rejected it after two
  instances collided. **Assert the current derivation; change nothing here.**

### The fixture (AC6) — modify in place, and four traps that would pass while proving nothing

- `folio-go/embedded_font_fixture_test.go:73-148` `embeddedFontTemplateJSON()` — **only two things
  are parameterised** (the asset key, twice at `:84`/`:128`, and the base64 `data` lines). The drawn
  text `"A font travels inside the template."` is a **hard-coded literal at `:108`**, mirrored once
  in `fixtures/embedded-font/input.folio:865`. Changing the text is a one-line edit in each.
- **Coverage, measured from the shipped cmaps:** `NotoSans-Regular.ttf` covers **zero** codepoints in
  U+0E00–U+0E7F; `NotoSansThai-Regular.ttf` covers **87** of them (and also Latin). So **any Thai
  rune** falls through `"Noto Sans"` to the embedded entry. **A pure-Thai string is the sharp
  witness — a mixed Latin+Thai string is not**, because the Thai face also covers Latin.
- **Ts-free Thai already committed** (all in `goldenDigestRecord`, none in `textRiseExemptGoldens`,
  so their streams are already proven to carry no ` Ts`): `สัญญา`
  (`fixtures/thai-stacked-marks/input.folio:7`, documented at
  `folio-go/thai_stacked_marks_template.go:37-45` as **the zero-offset control**),
  `ณัฐวุฒิ เกิด กรุงเทพ` (`fixtures/shaped-text/input.folio:7`),
  `ใบเสร็จรับเงินค่าน้ำประปาประจำเดือนมกราคม` (`fixtures/wrapped-text/input.folio:7`).
  **Avoid** `ทั้ง`-class clusters (`ั` + tone mark) — those carry a non-zero `YOffset`
  (`internal/pdf/textdoc.go:940-945`, `:1002-1007`) and would emit ` Ts`, requiring a
  `textRiseExemptGoldens` entry (`byte_neutrality_test.go:1363`).
- **Modify in place, do not add a fixture.** `embedded_font_fixture_test.go:307`
  (`entry.Name() != "embedded-font"`) keys on the directory **name**, and `:317` (`withAssets != 7`)
  counts asset-bearing directories — **both stay untouched under modify-in-place**, and both would
  break if a second fixture were added.
- **TRAP 1 — `embedded_font_fixture_test.go:218` `len(programs) != 1` certifies nothing after the
  change.** With chain `["Noto Sans", <embedded Thai>]` and pure-Thai text, a *correct* 8.4 embeds
  **one** program (the Thai face) — the same count as today's *incorrect* behaviour. **The check
  must become identity-based** (which face is embedded), never a count.
- **TRAP 2 — `matrix_test.go:1980-2004` `requireEmbeddedFaceStaysOffThePage` has the identical
  defect.** Its doc says *"WHEN STORY 8.4 LANDS THIS GUARD MUST CHANGE, and having to change it
  deliberately … is exactly why it is written down."* Inverting it to `!= 2` would be **wrong**;
  invert it to an identity assertion. The `matrixDocuments` entry (`:1963-1969`) otherwise stands.
- **TRAP 3 — `expected.json` has no "second literal".** `embedded-font` is not in
  `goldenDigestRecord`, so the second-literal discipline does not apply and its hash can be
  re-recorded invisibly. Nothing regenerates it: `TestTargetRenderHash`
  (`matrix_test.go:2260-2273`) writes the digest to `folio-go/.matrix-build/hash.<target>.<slug>.txt`
  and the test log, and a **human copies it in**. `assertFixturesShareToolchain` (`matrix_test.go:391`)
  requires `goToolchain`/`folioGoVersion` to stay consistent across all fixtures.
- **TRAP 4 — two tests go red the moment the text changes and before the feature lands, and both are
  more temptingly "fixed" than implemented.** `chain_face_names_test.go:125-127`
  (`TestEmbeddedEntryIsSkippedAtRender` fatals on any diagnostic — Thai text on a `["Noto Sans"]`
  chain warns per rune) and `missing_glyph_corpus_test.go:33`
  `TestCorpusFixturesProduceNoMissingGlyphWarnings` (asserts **zero** `DiagCodeTextMissingGlyph`).
  **These reds are the correct red-proof of the feature.** Never relax the diagnostic check.
- **Prose that becomes false and must be rewritten in the same commit:**
  `fixtures/embedded-font/README.md` (the whole *"Why there is no `expected.pdf`"* section and
  *"What it does NOT cover"*), `byte_neutrality_test.go:885`, `matrix_test.go:1932-1962`,
  `embedded_font_fixture_test.go:14-48`, `missing_glyph_corpus_test.go:191-201` and its
  `beyondBaselineAcceptance` string at `:256` (*"rendering from an embedded face is Story 8.4"*).
- **Registration for the new `expected.pdf`:** `byte_neutrality_test.go:92` `goldenDigestRecord`
  (dir + sha256 + every site; the test `os.ReadFile`s `expected.pdf` and errors "presence
  precondition" if missing, `:601-614`). That one entry auto-enrols the fixture in
  `TestEveryGoldenPDFResolvesItsPageTree` (`golden_structural_validity_test.go:90`),
  `TestNoPreStory80GoldenCarriesATextRise` (`:1396`) and `tounicode_corpus_test.go:138`.
  `declaredEpic2GateObligations` (`:853`) **already** declares `matrix-document: embedded-font`; a
  **new** obligation may not be added without a ruling saying so (`:1046-1052`).
- **Not affected:** `TestCorpusMeetsP6ExerciseFloors` / `TestCorpusP6StatsMatchDeclaredBaseline`
  (`internal/text/corpus_test.go:169,243`) read **only** `fixtures/thai-break-corpus/corpus.json`;
  the baseline is a Go `const` block at `:244-252`.

### The canvas — measurement (in scope) and paint (split out; see Design Notes)

- `folio-go/page_setup.go:1130` `addCanvasTextPaint` calls the render path's own `fontChain`
  (`:1162`), `shapeSegments` (`:1192`), `chainVerticalModel` (`:1229`), `positionSegments`
  (`:1288`/`:1298`). Comment `:1225-1227`: *"the canvas consumes the IDENTICAL advance the renderer
  does … the browser never measures text."* **AD-17** (`ARCHITECTURE-SPINE.md:361`).
  **So AC5's measurement half arrives with Task 1 and needs a test, not a feature.**
- `folio-go/page_setup.go:1145` — the canvas builds its **own** `fontCache`, separate from
  `predictDocument`'s (`render.go:1630`). Whatever seam Task 1 chooses must be reachable from
  **both**, or the canvas silently keeps the old behaviour.
- **Paint-half evidence, measured and recorded here so the successor story does not re-derive it:**
  `CanvasTextFragment` (`page_setup.go:115-121`) carries `Text` and `X` **only**; `fragment.face` is
  in scope at `page_setup.go:1344` and **discarded**. A fragment is **exactly one face by
  construction** (`wrap.go:28-36` `faceSegment.face` is scalar; `render.go:1459-1508`
  `positionSegments` emits one run per segment and never merges), so attribution must be
  **per-fragment** — per-component would be wrong for any mixed-script element, which is the
  `c6e4d03` case. `canvas_projection_wire_test.go` records key sets for the **top level and the two
  font-chain levels only** (`:40-60`, `:68`, `:89`) — **nothing for the paint tree**, so a new field
  on `CanvasTextFragment` would **not** redden it. The TS guard is `engine-protocol.ts:452`
  (`hasOnly(fragment, ['text','x'])`, a **subset** check → an unlisted key blanks the canvas with no
  diagnostic). `AssetBytes` (`asset_bytes.go:18`) is **media-type agnostic end to end** and usable
  for font bytes as-is; the only image-specific hop is the consumer at `App.tsx:1336`.
  **There is no runtime font registration anywhere in the designer** — no `new FontFace`, no
  `document.fonts.add`, no dynamic style injection; all three faces are build-time
  (`build-wasm.mjs:79`). `canvas-authority-contract.test.ts:145` silently rewrites every
  `document.fonts` occurrence to `fontReadinessOnly` **before** the prohibition scan, so its
  `/\bdocument\.fonts\b/` rule (`:24`) is **dead** and `document.fonts.add` trips nothing today.

### Read-only evidence

- oxlint baseline exactly **4** `only-export-components` warnings; Vitest baseline **323 tests /
  35 files** at 8.3's close.
- `lint`'s `TestFloatTypedTestScopeInventory` pins five sites **by line number** in
  `shaping_expectations_test.go` and `internal/fontset/{fontset,vendorboundary}_test.go` — any line
  shift there reddens it.
- Baseline red set must be **re-measured at `de87bef`**, not assumed; it has moved three times in
  this run.

## Tasks & Acceptance

**Execution:**

1. `folio-go/render.go` — carry the embedded arm past `chainFaceNames`. Choose **one** seam (widen
   the six `(FontSet, *fontCache)` consumers, or materialise a per-render name→bytes view upstream of
   the `[]string` narrowing) and state the choice and its rejected alternative in the doc comment.
   Must be reachable from **both** `fontCache` construction sites (`render.go:1630`,
   `page_setup.go:1145`). Rationale: AC1, AC2, and AC5's measurement half all follow from this.
2. `folio-go/render.go:1133-1136` — correct `chainFaceNames`' doc comment: it names four consumers
   where ten take `chain []string` and six take `(FontSet, *fontCache)`. Rationale: D-8.2.3 — the
   enumeration, not the arm.
3. `folio-go/render.go` + `folio-go/table_render.go:653-660` — **extract the chain LOOKUP and its
   error** into one function both `fontChain` and `collectBandTableRuns` call, retiring the
   hand-mirrored copy and the comment that admits it. Rationale: D-8.4.4(a) — the filter is already
   sole; the lookup is what drifted. Keep the two error strings byte-identical unless a test proves
   the change.
4. `folio-go/render.go` — add the font analogue of the image loop (`:1873-1936`) **inside
   `predictDocument`**: dedupe embedded asset keys, sorted order, `firstElementIDByAssetKey`,
   `DecodeAssetBytes` → `template.DecodeFontForRender` → `fontset.New`. Rationale: AC4's Render half
   **and** its `Validate` half, the latter by construction.
5. `folio-go/internal/template/fontasset.go` — extend `UnsupportedFontMediaTypeError` with the chain
   name and entry index and say so in the message. Re-point **and re-measure** every existing test
   asserting the old string; never merely update one. Rationale: AC4 + D-8.4.3.
6. `folio-go/render.go` — choose the embedded face's cache/`FontSet` key so it cannot collide with a
   caller-supplied face name, and state the precedence rule in a doc comment. Rationale:
   `fontCache`, `pagemodel.TextRun.Face` and `pdf.EmbeddedFace.Name` all key on "the caller's
   FontSet key"; AD-8 makes the asset key the resolver.
7. `folio-go/chain_face_names_test.go` — re-point the two mutation pins to the new boundary; keep
   them. `internal/template/fonts_embedded_test.go:927` `TestEmbeddedEntryIsInertUntilStory84` —
   change deliberately (rename and invert, or retire with the ruling that discharged it).
   Rationale: both assert facts this story removes.
8. `folio-go/embedded_font_fixture_test.go:108` + `fixtures/embedded-font/input.folio:865` — change
   the drawn text to a **pure-Thai, Ts-free** string already committed elsewhere in the corpus, so
   the embedded face is the only face that can draw it. Rationale: AC6; D-8.4.4(b).
9. `folio-go/embedded_font_fixture_test.go:218` and `folio-go/matrix_test.go:1998` — replace **both**
   `len(programs) != 1` count checks with **identity** assertions naming which face is embedded.
   Rationale: TRAP 1/2 — a correct 8.4 also embeds exactly one program, so the counts certify
   nothing either way.
10. `fixtures/embedded-font/expected.pdf` (new) + `folio-go/byte_neutrality_test.go:92` — ship the
    golden and add its `goldenDigestRecord` entry with every site. Re-record
    `fixtures/embedded-font/expected.json` from the matrix legs. Rationale: AC6/AD-21.
11. `fixtures/embedded-font/README.md`, `folio-go/byte_neutrality_test.go:885`,
    `folio-go/matrix_test.go:1932-1962`, `folio-go/embedded_font_fixture_test.go:14-48`,
    `folio-go/missing_glyph_corpus_test.go:191-201,256` — rewrite the prose that asserts the
    now-false negative ("stays at 22", "no `expected.pdf`", "rendering from an embedded face is
    Story 8.4"). Rationale: D-8.2.3 — a surviving spelling of a retracted claim reads as
    authoritative.
12. **Tests for the I/O matrix**, each red-proved by reverting its own production expression:
    mixed chain resolves in declared order; shared family name resolves by asset key; render with
    `fonts.Shipped()` alone; non-font asset **drawn** errors on `Render` **and** on `Validate` with
    identical strings (copy `chain_face_names_test.go:176-211`, **not** the image test); non-font
    asset **never drawn** renders clean; load still accepts. **Assert the `Validate` arm
    separately**, so removing the `Validate` call reddens a named test on its own. Rationale: AC4,
    and D-8.3.5's warning that the third half is the one that disappears.
13. `folio-go/page_setup.go` + a canvas test — **pin** that `CanvasWithTextPaint` on the
    embedded-face document measures identically to the PDF path (same fragment origins and
    advances). Rationale: AC5's measurement half, stated as *verified* rather than deleted
    (D-8.4.1a), so a later story cannot quietly fork the path.
14. `folio-designer/src/canvas-font-stack.test.ts` — update the DW-35 tripwire to record the **new**
    state: the engine now measures with an embedded face and the browser has **no CSS family for
    it**, naming the successor story. Rationale: the split's disclosure obligation, discharged by a
    test rather than a comment (8.2's precedent).
15. `_bmad-output/specs/spec-folio/folio-format.md` — record that a chain entry naming a non-font
    asset is accepted at load and errors at render, if the format doc does not already say so.
16. `_bmad-output/implementation-artifacts/deferred-work.md` — close **DW-83**; re-own **DW-35** to
    the named successor story with the paint scope and the asset-key derivation rule recorded;
    repair the entry's two contradictory `Owner:` bullets. Rationale: DW-87 — the register is the
    audit trail.

**Acceptance Criteria:**
- Given a chain mixing an embedded and a shipped face, when text is shaped, then the embedded face
  joins per-rune coverage resolution in declared order, and where an embedded and a shipped face
  share a family name the **asset key** decides with no name-based substitution.
- Given a `FontSet` containing only the shipped set, when the document names an embedded face, then
  it renders from the document's own bytes, reading no network, no host-installed font and no path
  on disk.
- Given the same document on all four AD-21 targets, when the outputs are compared, then they are
  byte-identical; subsetting happened once per render inside the PDF producer; no face was subset at
  save time; and the subset tag is a deterministic hash of the **emitted subset program bytes**.
- Given a chain entry naming a non-font asset, when the document is **loaded**, then load accepts it.
- Given that same document, when something must actually **draw** from that entry, then `Render`
  returns a located error naming the chain, the entry index and the asset key.
- Given that same document and identical inputs, when `Validate` is called, then it returns the
  **identical** error — asserted by a test that fails if the `Validate` path is removed while the
  `Render` path remains.
- Given a chain entry naming a non-font asset that nothing draws from, when the document is
  rendered, then it renders clean.
- Given the embedded-face document, when the canvas paint projection is produced, then its fragment
  origins and advances are **identical** to the PDF path's for the same inputs, asserted by a test.
- Given `fixtures/embedded-font/`, when this story closes, then its document draws text only the
  embedded face covers, it ships an `expected.pdf` recorded in `goldenDigestRecord`, the embedded
  face's presence on the page is asserted **by identity and not by count**, and the other 21 digests
  are unmoved.

## Spec Change Log

### 2026-09-01 — re-planned after the intent-gap ruling (D-8.4.1 … D-8.4.5)

The first plan gate halted `blocked` / `intent gap` on AC5. The `<intent-contract>` above is the
orchestrator's amended block, preserved verbatim; everything below it was re-derived at `de87bef`.

**What changed, and the known-bad state each amendment avoids:**
- **AC5 ruled (D-8.4.1).** The measurement half is *verified* and **pinned by test** (Task 13) rather
  than deleted — avoids "shared today, shared by luck tomorrow". The paint half is 8.4's subject by
  right (this story *creates* the condition), and its blocking design decision is **made**: the CSS
  family derives from the **asset key**, never from `font.family`. **Sizing was delegated to this
  gate and is exercised** — see Design Notes, "The sizing call".
- **AC3 corrected in the epic (D-8.4.2)**, with **AD-7** as the stronger ground alongside D-1.5.8.
  Avoids re-injecting the falsehood into every later reader.
- **AC4 joins the before-the-tag set, now three (D-8.4.3).** Avoids it being quietly deferred out.
- **D-8.4.4(a) corrected my own finding at the layer that changes the repair** — the filter is sole,
  the **lookup** drifted (Task 3). Avoids extracting the wrong thing.
- **`Covers:` swept (D-8.4.5).**

**KEEP on any re-derivation:** the four fixture traps in the Code Map (two count checks that certify
nothing, the second-literal-less `expected.json`, and the two reds that are the feature's own
red-proof); the instruction to copy `chain_face_names_test.go:176-211` and **not**
`render_image_test.go:258`; and the measured coverage fact that a **pure-Thai** string is the sharp
witness while a mixed one is not.

## Review Triage Log

### 2026-09-01 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 9: (high 0, medium 3, low 6)
- defer: 4: (high 0, medium 0, low 4)
- reject: 6: (high 0, medium 0, low 6)
- addressed_findings:
  - `[medium]` `[patch]` **The canvas aborted the whole projection on a document the format calls
    valid.** `page_setup.go`'s `shapeSegments` arm returned an error, and this story made that
    pre-existing `return` reachable from document content. Verified by probe: `CanvasWithTextPaint`
    over chain `["Noto Sans", {asset: image/png}]` with Thai text errored, where the same document
    projected clean at baseline `15ca0dd`; `wasm/engine.go:119,255,294` all propagate it, so the
    designer could not open the one document whose bad chain entry it was needed to repair. Fixed by
    tolerating `*template.UnsupportedFontMediaTypeError` via `errors.As` and degrading the element the
    way the adjacent `fontChain` arm does under **D-7.4.2** ("DEGRADE THIS ELEMENT, NEVER ABORT THE
    PROJECTION"), setting `column.FontChainDegraded`. Any other shaping error still aborts. Pinned by
    `TestCanvasDegradesRatherThanAbortingOnANonFontChainEntry`; re-probed: canvas now returns nil.
  - `[medium]` `[patch]` **The located error could name a chain the element does not draw through.**
    `newEmbeddedFaceIndex` stored one `FontChainSite` per asset key, first occurrence in sorted chain
    order. Verified by probe: with chains `aaa` and `body` both naming asset K and element `e1`
    drawing through `body`, Render reported `font chain "aaa" entry 1` — sending the author to edit
    the wrong entry, against AC4's requirement that the error name the chain. Fixed by carrying every
    occurrence and scoping resolution to the drawn chain via `fontCache.forChain`, which shares the
    cache's maps so no chain-consumer signature widened and no face is re-parsed. Pinned by
    `TestTheLocatedErrorNamesTheChainTheElementDrawsThrough`; re-probed: now names `body`.
  - `[medium]` `[patch]` **The author-facing missing-glyph diagnostic printed a 64-hex asset digest.**
    Verified by probe: `no face in chain [Noto Sans, asset:c94562c1…73caf] covers U+6F22 (漢)`. The
    assertion that forbade this was deleted in the implementation — correctly, since its premise was
    that an embedded entry never reaches the chain as a face name, which this story reverses — but the
    harm it predicted became real and unguarded, against D-000.37's "here is what to fix". Fixed with a
    display spelling used in the message path ONLY (`embedded "Noto Sans Thai"`); the reserved
    `asset:<key>` name is unchanged everywhere it resolves, caches, names a run or reaches the PDF.
    Pinned by `TestMissingGlyphMessageSpellsACarriedFaceForAHuman`; re-probed: reads
    `[Noto Sans, embedded "Noto Sans Thai"]`, and no rendered byte moved.
  - `[low]` `[patch]` **A false mechanism claim in the permanent register.** `deferred-work.md` stated
    the embedded entry "is deliberately **not** decoded by the vertical-model walk
    (`chainLineMetrics`)". Verified false by reading `fontCache.metricsFace`: it calls `declares()`
    (true for any embedded name) then `get()`, which runs `parseEmbedded` and base64-decodes; only the
    *failure* is swallowed. Corrected to describe what the code does, with `metricsFace`'s own doc
    comment brought into line, and the entry's "three parts" heading reconciled with its four-row table.
  - `[low]` `[patch]` **Golden-count prose off by one.** Verified by counting `goldenDigestRecord`
    entries at both revisions: 22 at `15ca0dd`, 23 now — so 22 others, not 21. Fixed in
    `byte_neutrality_test.go` and `embedded_font_fixture_test.go`, and the sentence made
    artifact-specific: `expected.json`'s recorded sha256 genuinely moved
    (`db400698…e513ad` → `f533b04b…d851832`) while `expected.pdf` ships for the first time.
    `fixtures/embedded-font/README.md` restored to recording a measured RESULT (the digest at 3,225
    bytes) rather than the present-tense procedure claim that had replaced 8.3's record.
  - `[low]` `[patch]` **A comment claimed a guard that did not exist.** `embedded_face.go` said
    `TestCanvasMeasuresWithTheEmbeddedFace` "reddens if a third site ever calls `newFontCache()` with a
    `*Template`"; that test only exercises `addCanvasTextPaint`, so a new third production site would
    have shipped green. Replaced with a real AST scan,
    `TestOnlyTheTwoDocumentAwareSitesBuildAFontCache`, pinning the two `newDocumentFontCache` sites
    (`render.go:predictDocument`, `page_setup.go:addCanvasTextPaint`) and zero production
    `newFontCache` callers; red-proved by swapping the canvas site.
  - `[low]` `[patch]` **Task 14's disclosure test named three mechanisms and scanned two.**
    `canvas-font-stack.test.ts` asserted "None exists anywhere in this build. When one arrives, this
    reddens" while scanning only `new FontFace` and `document.fonts.add` — an injected `@font-face`
    with a `data:`/`blob:` `src`, the third mechanism its own comment names, would have left it green.
    Since this test *is* the split's disclosure obligation, it must not overclaim. Fixed with a
    `registersAFaceAtRuntime()` detector covering all three, plus a test that the detector matches four
    spellings and rejects three benign ones so the negative scan cannot go vacuous.
  - `[low]` `[patch]` **An undocumented, untested semantic asymmetry.** Verified by probe: chain
    `["Noto Sans", {asset: image/png}, "Noto Sans Thai"]` over Thai text fails the whole render at
    entry 1 even though entry 2 covers every rune, while the control
    `["Noto Sans", "No Such Face At All", "Noto Sans Thai"]` renders clean — a non-font asset refuses
    where an unsupplied face skips. Behaviour KEPT (a non-font asset is a document defect that travels
    with the file; an absent face is a deployment condition; the spec puts the decode at
    `resolveRuneFace`); what was missing was the pin and the disclosure. Added
    `TestANonFontAssetRefusesWhereAnUnsuppliedFaceSkips` fixing both arms, and stated the rule in
    `folio-format.md` — which also gained the `Validate`-returns-the-identical-error half it had
    omitted, the half D-8.3.5 says always gets dropped.
  - `[low]` `[patch]` **The story removed an exact-count bound without replacing it.** The pre-8.4
    guards bounded the render to exactly one embedded program; replacing them with identity-only
    assertions left a third, unrelated subset face able to reach the page unnoticed on every
    cross-target leg, since the identity scan only names two faces. TRAP 1 is right that the count
    cannot BE the identity check — 1 is the answer on both sides of the story — but as a bound beside
    the identity it is not the same claim. Restored in the shared helper
    `requireEmbeddedFaceDrewThePage` (used by the in-process test and all four matrix legs) and
    red-proved by flipping the expected count.

**Deferred (4), all low, recorded in the frontmatter `deferred:` list:** the new Thai golden's
sign-off obligation exists only as a register entry with no red gate, and cannot be given one inside
this story without firing its own `Block If`; a readable-but-never-drawn embedded entry now
contributes metrics to the vertical model, so line advance can differ from pre-8.4 for such a
document (no committed golden can observe it, since only `embedded-font` has an embedded entry); a
chain whose ONLY entry is an unreadable embedded one is untested, its correct error owed to
evaluation order nothing fixes; and two branches the code documents as structurally unreachable are
untested.

**Rejected (6), enumerated with the ground each was refused on (DW-87). Each ground refutes the
specific claim at the cited location.**

1. **Claim:** "`Validate` now returns errors whose text says `Render`" — cited at `render.go:787/810/887`
   reached through `predictDocument`. **Refused:** not caused by this change, and refusing it is
   required by the story's own AC. Verified at baseline `15ca0dd`: `render.go:1921` already returned
   `fmt.Errorf("folio: Render: %w", derr)` for the image decode, inside `predictDocument`, which
   `Validate` has called since before this story — so the prefix predates the diff on a path the diff
   did not touch. Independently, AC4 requires `Validate`'s and `Render`'s strings be byte-identical, so
   altering the prefix on the `Validate` side would break the criterion the finding sits beside.
2. **Claim:** "A plain string chain entry spelled `asset:<key>` silently resolves from the document's
   assets, shadowing the caller's `FontSet`" — cited at `render.go:1195-1205`. **Refused:** verified by
   reading `fontCache.get`: the embedded arm is consulted only via `c.embedded[name]`, a map populated
   solely from the document's own embedded chain entries. A name in the reserved namespace therefore
   resolves to the document's own asset of exactly that name and to nothing else, which is the stated
   precedence rule at `render.go:1214-1221` and is pinned by
   `TestEmbeddedFaceWinsOverACollidingFontSetKey`. Where the document does not carry that key the name
   falls through to the `FontSet` unchanged.
3. **Claim:** "`nonFontAssetKey` / `nonFontAssetData` are copy-pasted and nothing ties the key to the
   bytes — edit the data and the key silently stops being its SHA-256" — cited at
   `chain_face_names_test.go:81-83`. **Refused:** the posited silent failure does not exist. Verified at
   `internal/template/parse.go:487`: the load path computes `gotDigest := sha256HexOf(decoded)` and
   compares it to the key, so a mismatched pair fails loudly at `ParseTemplate` rather than surfacing
   later as something unrelated.
4. **Claim:** "The `lookupFontChain` extraction is unrelated refactoring riding along in a story
   already flagged `oversized`/`multiple-goals`" — cited at `table_render.go`. **Refused:** on the
   intent's own authority, not the spec's. The `<intent-contract>`'s Approach section names this work
   verbatim — *"So: extract the LOOKUP, not the filter"* — and the same paragraph identifies the
   hand-mirrored copy at `table_render.go:653-660` as its target. It is inside the contract, not
   alongside it.
5. **Claim:** "The story's `deferred:` frontmatter entry carries no DW identifier and never reaches
   `deferred-work.md`" — cited at the spec frontmatter. **Refused:** this is the recorded convention
   working as designed, not a gap. DW-83's own register entry states it: a build-filed deferral lives
   in the spec's frontmatter and is entered into the register at close. Close is the story-closer's
   step, not this dispatch's; writing it here would duplicate the entry under a DW number this
   dispatch has no authority to mint.
6. **Claim:** "`pdfEscapedEmbeddedFaceName` re-implements `internal/pdf`'s name escaping by hand, so a
   change to the real rule leaves the test helper green" — cited at `embedded_font_fixture_test.go`.
   **Refused:** the helper is in `package folio` and `internal/pdf`'s escaper is unexported, so the
   duplication is a package-boundary consequence rather than a choice, and the assertion it feeds is
   corroborated on the same bytes by two independent checks that do not use it (the `/BaseFont`
   identity scan and the exact-count bound). A drift in the escape rule would move the fixture's
   recorded digest and redden all four matrix legs before it could make this assertion vacuous.

## Design Notes

**AC3's false premise — CLOSED by D-8.4.2; do not re-open.** This heading is kept alive because the
preserved `<intent-contract>`'s third **Block If** bullet points at it by name, and the contract is
not this spec's to edit. The premise was that the subset tag hashes the **glyph set**. Measured:
`deriveTag` (`internal/fontset/fontset.go:910-924`) is FNV-1a over the **complete emitted subset
program bytes**, and **AD-7 names the glyph-set reading and rejects it** — *"unlike a hash of the
sorted glyph-id set alone"* — after **D-1.5.8** found two pinned instances of one variable face
collided under it. **The epic text is corrected (D-8.4.2), so the Block If can no longer fire from
the AC's wording.** It remains live in its literal sense: closing the derivation onto the glyph set
is still forbidden, would still move all 22 digests, and is still a halt if anyone proposes it.
Assert the current derivation; change nothing in `deriveTag`.

**The sizing call — `multiple-goals` STANDS, and the paint half splits to a named successor.**
The contract delegates this to the gate. Measured at `de87bef`, the paint half is a second
deliverable, on all three of the tests this run has used:

1. **Separably shippable.** The engine half discharges FR54 and the story's own user sentence
   *"As an integrating Go developer … render on a machine that has never seen them"* completely. The
   paint half serves a **different user on a different surface** (the designer author). Neither
   blocks the other: the engine half ships correct PDFs today, and the paint half is additive.
2. **It carries a mechanism with no precedent.** There is **no runtime font registration anywhere in
   the designer** — no `new FontFace`, no `document.fonts.add`, no dynamic style injection. All three
   shipped faces are build-time (`build-wasm.mjs:79`). The whole registration path is built from
   scratch, plus a **new** projection field on `CanvasTextFragment` with its TS guard in the same
   commit (`engine-protocol.ts:452`, a `hasOnly` subset check — get it wrong and the canvas blanks
   with no diagnostic), plus a **fragment-level wire tripwire that does not exist** (the wire test
   records the top level and the two font-chain levels only).
3. **It requires re-authoring two guards that were written to forbid exactly this shape.**
   `canvas-font-stack.test.ts:100-106` asserts the fragment stack contains **no `var(`** — the
   mechanism trips it immediately; `:123-132` forbids any source naming a chain entry in a
   font-family position. A third assertion (`:108-121`, the non-intersection tripwire) survives
   mechanically but **becomes false in spirit**. That is a rewrite that must *keep* its shipped-face
   guarantees while adding the embedded exception with its own teeth — not an edit.

Bundled into 8.4, this doubles the largest story in the epic and puts a from-scratch browser
mechanism in the same diff as a six-signature engine seam, a golden re-record and four
certify-nothing traps. **Split.**

**Which parts of the preserved contract the split reassigns, named so no pointer dangles.** The
`<intent-contract>` above is the orchestrator's and is not edited; its **Block If** authorises this
split in its own words (*"if the paint half genuinely makes this two shippable deliverables, return
`multiple-goals`"*). Two of its I/O matrix rows are therefore the **successor's**, not 8.4's:
**"Canvas PAINTS with the embedded face"** and **"Embedded and shipped share `font.family`"** — both
describe browser rasterization and the CSS-family derivation. Every other row, including **"Canvas
MEASURES with the embedded face"**, is 8.4's and is delivered by Task 13. The asset-key derivation
rule in the **Always** bullets stays binding on the successor and is quoted into its deferral entry
by Task 16, so the ruling travels with the work rather than staying behind in this file.

**What the successor owns, so it is disclosed rather than discovered.** Per-fragment face attribution
on `CanvasTextFragment` (the value is already in scope at `page_setup.go:1344` and discarded) with
its TS type, guard and a new fragment-level wire record; a named asset-key → CSS-family derivation
module; runtime `FontFace` registration hoisted to document scope (**not** per component — the
`ImagePaint` effect at `App.tsx:1326-1341` is the closest pattern and is the *wrong* lifetime); the
`.canvas-text-fragment` `font-family` var head; the `canvas-font-stack.test.ts` re-author; and a
deliberate narrowing of `canvas-authority-contract.test.ts:24`/`:145`, whose blanket
`document.fonts` → `fontReadinessOnly` rewrite makes its own prohibition dead and would let a
measurement call in unnoticed. **Sequenced immediately after 8.4**, per D-8.4.1 — not "later in
Epic 8". It needs writing into the epics file and DW-35 re-owned to it by name (Task 16); an
unwritten story is a deferral wearing a story's name.

**The disclosure is a test, not a comment (Task 14).** 8.2's precedent: assert the absence so its
arrival trips something. After 8.4 the true state is *"the engine measures with an embedded face and
the browser has no CSS family for it"*, and `canvas-font-stack.test.ts` is where that is already
recorded — its existing DW-35 tripwire comment cites "Story 8.4, AC4" and states the obstacle as
unresolved, both now wrong.

**The seam choice, and why it is left to the implementer.** Six functions take
`(chain []string, fs FontSet, cache *fontCache)` and none can reach `Assets`; four more take the
chain for messages and vertical arithmetic only. Widening six signatures spreads the embedded
question across six answer sites — precisely what `chainFaceNames`' comment says it exists to
prevent. Materialising a per-render name→bytes view keeps one answer site but must be built at
**both** `fontCache` construction points (`render.go:1630`, `page_setup.go:1145`), or the canvas
silently keeps the old behaviour and Task 13's pin is what catches it. Either is defensible; the
comment must be corrected either way, because its "four" is neither the population nor the risk set.

**`Validate`'s half, and the trap that drops it.** Wiring the decode inside `predictDocument` makes
AC4's third half true by construction. The failure mode D-8.3.5 names is placing the check somewhere
`Render` reaches and `Validate` does not — `renderDocument` or `buildPageModel`. Two defences:
`TestValidateNeverReachesRenderOrInternalPDF` makes that structurally visible, and Task 12's separate
`Validate` assertion reddens on its own. **The image precedent has exactly this omission already**
(`render_image_test.go:258` asserts `Render` only), which is why it is named here as the thing not
to copy.

**No diagnostic code is minted** (D-7.8.1): no consumer branches. `folio.Validate` is unreachable
from the wasm engine (`wasm/engine.go:199` reparses rather than calling it) and has **no call site
at all** in `folio-designer/src`; the CLI (`cmd/folio/main.go:235`) prints messages, not codes.

**`Covers:` re-checked at this gate (D-8.2.8(a)), post-sweep.** The amended line is
`FR54, FR33 · AD-7, AD-8, AD-14, AD-17, AD-21, AD-22, D-1.8.1, NFR1, NFR1.d, NFR1.e — and DW-83
(AC4) and DW-35 (AC5)`. Read against the ACs: AC1→AD-8; AC2→FR33, FR54; AC3→AD-7, AD-22, NFR1;
AC4→AD-14, D-1.8.1; AC5→AD-17; AC6→AD-21. **No omission remains.**

**AC6's human sign-off, checked rather than assumed.** Story 8.0's measured precedent: a new golden
that invalidates no existing attestation and leaves no `//go:build matrix` sign-off gate red closes
`done`. The three grounds hold here — the other 21 digests are unmoved, no existing sign-off's
`sha256` is touched, and no agent writes `reader`/`date`/`examined`. **But the fixture's subject is
now Thai rendered from an embedded face**, and if the full `-tags=matrix` sweep leaves a sign-off
gate red, that is the `Block If` firing. Record any new human-reading question as a **deferral** with
the README stating the gap — not a halt, and never a fabricated attestation.

## Verification

**Commands** (from the repo root unless stated; `-count=1` on **every** Go gate — D-7.9.5, a warm
cache is not an anchor):
- **Before the first edit:** `shasum -a 256 fixtures/*/expected.pdf > <scratch>/digests.before` —
  expected **22** lines. Re-measure the **baseline red set** in a detached worktree at `de87bef`
  (`git worktree add --detach`), never by assumption: it has moved three times in this run.
- `cd folio-go && go test -count=1 ./...`
- `cd folio-go && go test -count=1 -tags=matrix ./...` — the **full** sweep, not a subset.
- `cd folio-go && go vet -tags=matrix ./...` — expected clean.
- `gofmt -l folio-go` — expected empty (`lint/…/licencegraph_test.go` is a known pre-existing
  offender, DW-23, outside this path).
- `cd folio-go && for T in darwin/arm64 linux/amd64 linux/arm64 js/wasm; do FOLIO_MATRIX_TARGET=$T
  go test -count=1 -tags=matrix -run TestTargetRenderHash -v .; done` — **all four AD-21 targets.**
  `TestTargetRenderHash` **asserts nothing unless `FOLIO_MATRIX_TARGET` is set**, so also run the
  **unset control** and report it as the deliberate no-op it is. Report all five legs with timings —
  the timing contrast is the evidence the legs actually asserted.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .`
- `cd lint && go test -count=1 ./...` — **`-count=1` is mandatory**: the rules package walks the
  `folio-go` tree with `ReadDir`, which Go's test cache does not track, so a cached `ok` here is no
  measurement at all.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` —
  typecheck clean; oxlint exactly **4** pre-existing `only-export-components` warnings; Vitest at or
  above the **323 tests / 35 files** baseline (report actual numbers); e2e **compiles only** and
  executes nowhere by design.
- `shasum -a 256 fixtures/*/expected.pdf | diff - <scratch>/digests.before` — **the 21 digests other
  than `embedded-font` must be byte-identical.** `embedded-font` appears as a new line (it ships no
  `expected.pdf` today); its digest is the one deliberate, story-owned re-record (D-8.4.4b). Any
  other movement is a failure, not a re-record.
- `cd folio-go && GOOS=js GOARCH=wasm go test -count=1
  -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" ./wasm/cmd/engine/` — this package is invisible
  to `go test ./...` (`//go:build js && wasm`), so "it compiles" is not evidence.

**Mutation proofs required** (each states its anchor before editing, and reverts after): the
embedded-entry resolution (deleting it must redden a named test, not merely change a count); the
`Validate` arm **on its own**; the fixture's face-identity assertions, proved by pointing them at
the wrong face; and the canvas measurement pin, proved by forking the canvas off the shared path.

**There are exactly TWO standing reds. Both are named; neither may be hidden, skipped or "fixed".**
1. `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` — got 7, need ≥20, under `./...`. The
   mandated corpus-sourcing red, with its drift twin `TestCorpusP6StatsMatchDeclaredBaseline`
   (`internal/text/corpus_test.go:169,243`; `baselineP6g = 7` at `:251`).
2. `TestShippedFacesReproduceFromUpstream` — under `-tags=matrix` **only**
   (`fontgen_matrix_test.go:64`), failing with `fontgen: fontTools is not importable`. Registered as
   **DW-86 (D-8.3.4)**. It **fails rather than skips deliberately** (`fontgen_matrix_test.go:26-32,
   68-82`) — *"the sources were not present"* must never read as *"the faces reproduce"*.
   **Confirm it at THIS story's baseline `de87bef` before calling it pre-existing.**

**Expect two TRANSIENT reds while the fixture text changes and before the feature lands**, and treat
them as the feature's red-proof rather than as failures to silence: `chain_face_names_test.go:125-127`
and `missing_glyph_corpus_test.go:33`. **Never relax a diagnostic check to clear them.**

**Any third standing red is a real failure. Report it; do not absorb it.**

**Manual checks:**
- `git status --porcelain` before committing — only this story's paths. Stage **explicit paths
  only**; never `git add -A` / `git add .`.
- Root `README.md` untouched (md5 `078d7d80d518d54af2fc04fb270d46b8`).
- Neither `fixtures/statement-signoff.json` nor `fixtures/thai-stacked-marks/signoff.json` appears
  in the diff.
- Every commit message ends with the trailer
  `Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>`.
- **`## Review Triage Log`: every REJECTED finding is enumerated the way `addressed_findings`
  records a patch — the claim, its cited location, and the ground on which it was refused.** A bare
  `reject: N` cannot be audited; that is **DW-87**. A rejection is sound only when it refutes the
  *specific claim at the cited location*; a true fact about nearby code is not a refutation.

## Auto Run Result

Status: done
Blocking condition: none

Implement-review-commit dispatch. Baseline `15ca0ddbc4565d935fde026bfbad463be8ddd182` on `main`.
The `<intent-contract>` was preserved **byte-identical** (sha256 of the block verified against the
baseline copy before and after every edit). Never pushed, never branched; explicit paths staged only.

**What shipped.** `chainFaceNames` now MAPS an embedded chain entry to a reserved face name derived
from its asset key (`asset:<key>`, D-8.4.1 — never `font.family`) instead of dropping it, and
`fontCache` became the single answer site for "where do this face's bytes come from", with a stated
precedence: a name in the reserved namespace resolves from the document's own assets and the supplied
`FontSet` is never consulted for it. The seam chosen was a per-render name→bytes view upstream of the
`[]string` narrowing — the alternative, widening the six `(FontSet, *fontCache)` consumers, is
recorded as rejected in `embedded_face.go`'s doc comment — so none of the ten chain-taking signatures
moved. Both `fontCache` construction sites now build from the document (`predictDocument`,
`addCanvasTextPaint`), which makes AC4's `Validate` half true by construction and AC5's measurement
half true by sharing rather than by a second rule system. The decode is lazy, at `resolveRuneFace`,
which is what keeps "an entry nothing draws from renders clean" true.

**Files changed** (24 in the implementation commit, 15 more touched by the review patches):
`folio-go/embedded_face.go` (new — the seam, the index, the lazy decode); `folio-go/render.go`
(`chainFaceNames` mapping, `fontCache` embedded arm + `declares`/`metricsFace`/`forChain`, the
extracted `lookupFontChain`, the corrected "four consumers" comment); `folio-go/table_render.go`
(hand-mirrored lookup retired); `folio-go/page_setup.go` (document-aware canvas cache; the
capability-error degrade); `folio-go/internal/template/fontasset.go` (`FontChainSite`, the widened
`UnsupportedFontMediaTypeError`); `fixtures/embedded-font/{input.folio,expected.pdf,expected.json,README.md}`
(text → `สัญญา`, first golden); `folio-go/byte_neutrality_test.go` (`goldenDigestRecord` 22 → 23);
`folio-go/{chain_face_names,embedded_font_fixture,canvas_embedded_face,font_cache_sites,matrix,missing_glyph_corpus,render}_test.go`;
`folio-designer/src/canvas-font-stack.test.ts` (the DW-35 / 8.4a disclosure);
`_bmad-output/specs/spec-folio/folio-format.md`; `_bmad-output/implementation-artifacts/deferred-work.md`
(DW-83 closed, DW-35 re-owned to 8.4a); `_bmad-output/planning-artifacts/epics.md`.

**Review findings: 9 patched (3 medium, 6 low), 4 deferred (all low), 6 rejected — each rejection
enumerated with its refuting ground in the Review Triage Log above, per DW-87.** No `intent_gap` and
no `bad_spec`: the three medium patches were a canvas projection that aborted where D-7.4.2 says to
degrade, a located error that could name a chain the element does not draw through, and an
author-facing diagnostic that printed a 64-hex asset digest. All three were confirmed by direct probe
before triage and re-probed after the fix.

**Follow-up review recommended: true.** Patched findings only: 0 high, 3 medium, 6 low →
`3 × 3 + 1 × 6 = 15`, at or above the threshold of 5.

**Verification — what it actually printed** (every Go gate `-count=1`; all figures measured in this
dispatch, not carried over from the subagents' reports):
- `go test -count=1 ./...` — **1798 pass, 2 fail, 5 skip**; the two failures are
  `TestCorpusMeetsP6ExerciseFloors` and its subtest `P6g_(opaque_names)` (got 7, need ≥20).
- `go test -count=1 -tags=matrix ./...` — **1808 pass, 3 fail**; the same P6g pair plus
  `TestShippedFacesReproduceFromUpstream` (`fontgen: fontTools is not importable`, DW-86).
- **Both standing reds confirmed AT THIS STORY'S BASELINE** `15ca0dd` in a detached worktree before
  any edit — plain 2 fails, matrix 3 fails, byte-identical sets. **No third red at any point.**
- `go vet -tags=matrix ./...` exit 0; `gofmt -l folio-go` empty.
- Four AD-21 legs, all `ok`, `embedded-font` → `f533b04b…d851832` on every one: darwin/arm64 1.075s,
  linux/amd64 6.809s, linux/arm64 5.456s, js/wasm 11.265s. **Unset control** `ok` in 0.383s — the
  deliberate no-op; the timing contrast is the evidence the four legs asserted.
- `TestCrossTargetByteIdentity` `ok` 23.178s.
- `cd lint && go test -count=1 ./...` — 4 packages `ok`.
- Designer: typecheck clean; oxlint **exactly 4** pre-existing `only-export-components` warnings;
  Vitest **325 passed / 35 files** (baseline 323/35); `test:e2e:compile` clean.
- `GOOS=js GOARCH=wasm go test ./wasm/cmd/engine/` — `ok` 0.411s.
- **Golden digests: `shasum -a 256 fixtures/*/expected.pdf` captured BEFORE the first edit (22 lines)
  and diffed after (23 lines) — exactly one ADDED line, zero moved.** `fixtures/embedded-font/`
  ships its first `expected.pdf`; `expected.json`'s recorded sha256 moved
  `db400698…e513ad` → `f533b04b…d851832`, the one deliberate story-owned re-record (D-8.4.4b).
- Root `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8` unchanged; neither human sign-off JSON
  appears in the diff.

**Mutation proofs run by this dispatch, independently of the implementer's own:** reverting
`chainFaceNames`' embedded arm to the pre-8.4 drop reddens **11 named tests**, including the fixture
identity pin and both canvas pins — so the coverage is not vacuous. Making `Validate` swallow
`predictDocument`'s error while leaving `Render` intact reddens **4 named tests**, including
`TestNonFontAssetDrawnErrorsAtRenderAndAtValidate` — AC4's third half reddens on its own. The
restored exact-count bound was red-proved by flipping the expected count. All mutations reverted; tree
verified clean after each.

**Matrix test audit.** Every I/O row is covered by a test that ran and passed, except the two rows the
plan gate reassigned to **Story 8.4a** ("Canvas PAINTS with the embedded face" and "Embedded and
shipped share `font.family`"), on the authority of the contract's own `Block If`, which authorises the
split in its own words and requires the limitation be disclosed. The disclosure is **Task 14's test**,
not a comment — `canvas-font-stack.test.ts` now records that the engine measures with a carried face
while the browser has no CSS family for it, and scans for all three registration mechanisms so it
reddens when 8.4a lands.

**Residual risks.** The carried face's CSS-family rule is honoured one layer below where the contract's
Always bullet names it — the Go `FontSet` namespace, not the browser's font registry — because the
paint half is 8.4a's; nothing here asserts anything about CSS family names. A readable but never-drawn
embedded entry now contributes metrics to the vertical model, so an integrator's document could see its
line advance change across this version boundary; no committed golden can observe it. And
`fixtures/embedded-font/expected.pdf` is the corpus's only Thai-bearing golden with no human reading
attestation — filed as a deferral with the README stating the gap, never a fabricated sign-off. Its
grounds for closing were checked rather than assumed: the full `-tags=matrix` sweep leaves no sign-off
gate red and invalidates no existing attestation.
