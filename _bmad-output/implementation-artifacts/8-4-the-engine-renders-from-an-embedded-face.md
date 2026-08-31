---
title: 'The engine renders from an embedded face'
type: 'feature'
created: '2026-08-31'
status: 'draft'
baseline_revision: 'd52b1b073c621c9e4f3c7ea2c1091f93f96294fb'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized', 'multiple-goals']
deferred: []
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

**The boundary this story removes**
- `folio-go/render.go:1137-1147` `chainFaceNames` — the whole gap, 8 lines: `if entry.Embedded()
  { continue }`. Doc comment `:1109-1136` names this story. **Two non-test callers**:
  `render.go:1106` (inside `fontChain`) and `table_render.go:666` (the table **header-style** path,
  which does its own `Fonts.Chain` lookup at `:654` and does **not** route through `fontChain` —
  a second conversion site that must be changed too, or table headers silently keep the old drop).
- `folio-go/chain_face_names_test.go:83` and `:176` — the existing pins. `:176`
  `TestAllEmbeddedChainProducesTheExistingLocatedError` is **the shape to copy for AC4**: it already
  asserts load-accepts → Render-errors → **Validate errors identically**, including
  `verr.Error() != rerr.Error()` string equality at `:206` and `len(diags) != 0` at `:209`.

**Face resolution — what consumes names, and the seam to widen**
- `folio-go/render.go:1202-1215` `resolveRuneFace(chain []string, r rune, fs FontSet, cache
  *fontCache)` — declared order at `:1203`, **silent skip** of a name absent from the FontSet at
  `:1204-1206`, `HasGlyph` at `:1210`.
- **Three siblings apply the same silent-skip rule independently** and all take `chain []string`:
  `chainLineMetrics` (`wrap.go:550-565`, the "no member present" error in `verticalModel` at
  `:593-598`), `fontCache.get` (`render.go:1167-1181`, the hard-failure form at `:1173`), and
  `digitTableRun` (`page_number.go:436`, called from `render.go:989`). **This is the actual seam:**
  `resolveRuneFace`, `shapeSegments`, `chainLineMetrics`, `digitTableRun` and `fontCache` all take
  `(chain []string, fs FontSet)` and hold **no document**. Either widen all four, or materialise a
  per-render name→bytes view before the call. Choosing is this story's design work.
- `folio-go/fontset.go:20` `type FontSet map[string][]byte` — name → raw sfnt bytes.
- `folio-go/internal/fontset/fontset.go:85` `New(name string, data []byte) (*Font, error)` — **takes
  bytes and does not care where they came from.** The key affordance. Called from exactly one place:
  `fontCache.get` (`render.go:1176`).
- `folio-go/render.go:1152-1181` `fontCache` — `map[string]*fontset.Font`, keyed by **bare face
  name**, constructed twice in non-test code (`render.go:1630` once per `predictDocument`;
  `page_setup.go:1145` once per `CanvasWithTextPaint`). `pagemodel.TextRun.Face`
  (`internal/pagemodel/pagemodel.go:53-56`) and `pdf.EmbeddedFace.Name`
  (`internal/pdf/textdoc.go:17`) are both documented as "the caller's FontSet key" — **an embedded
  face needs a synthetic, non-colliding key and a stated precedence rule.**

**The render-time asset precedent to mirror**
- `folio-go/render.go:1873-1936` — the image loop, and the shape to copy exactly: dedupe asset keys
  (`:1873-1888`), `firstElementIDByAssetKey` for located errors (`:1882-1887`), sorted deterministic
  order (`:1889`), `visibleAssetKeys` (`:1900-1903`), presence check naming element+key
  (`:1906-1912`), `DecodeAssetBytes` (`:1913`), `DecodeImageForRender` (`:1919`), then **validate
  unconditionally, embed only if visible** (`:1924-1928`). Note `:1921` returns a bare
  `fmt.Errorf("folio: Render: %w", derr)` — **no `*RenderError`, no code, no `ElementID` field**;
  location is prose inside the message.
- `folio-go/internal/template/base64.go:41-43` `DecodeAssetBytes`. Not memoized anywhere.
- `folio-go/internal/template/fontasset.go` — **built by 8.3, zero non-test call sites.** `:25-38`
  `UnsupportedFontMediaTypeError{AssetKey, ElementID, MediaType}`; `:86-94` `DecodeFontForRender`;
  `:71-78` `decodeRecognisedFont` (`font/ttf`, `font/otf` only; woff/woff2/collection deliberately
  excluded at `:57-62`); `:137-193` `checkSfnt`. Comment `:80-85`: *"Story 8.4 is what gives it a
  call site on the render path."* **Note the error names AssetKey + ElementID + MediaType — it does
  NOT carry the chain name or entry index that AC4 requires.** Extending it, or wrapping it, is
  this story's work.
- `internal/template/fonts_embedded_test.go:927` `TestEmbeddedEntryIsInertUntilStory84` — asserts
  `DecodeFontForRender` has **zero** call sites in that package. **This story must change it
  deliberately**, not delete it.

**`Validate` — why AC4's third half is free if wired correctly**
- `folio-go/validate.go:52` `Validate(b, d, p, f) ([]Diagnostic, error)`; doc `:17-19`: *"a DRY-RUN
  PREDICTOR of Render, not a second rule system … and NO RENDER IS ATTEMPTED"*; `:22-24`: shares
  its implementation via `predictDocument`.
- `folio-go/render.go:1625` `predictDocument` — the single shared derivation. `Render` reaches it
  via `renderDocument` (`:1529`) → `buildPageModel` (`:1606`); `Validate` calls it directly at
  `validate.go:68`. **The image decode at `render.go:1919` is already inside it**, so `Validate`
  today already predicts the unrecognised-image-media-type error — **but no test asserts that
  agreement** (`render_image_test.go:258` asserts Render only). The font analogue must not repeat
  that omission.
- `folio-go/render_arch_test.go:326` `TestValidateNeverReachesRenderOrInternalPDF` — AST guard.
- `folio-go/render_error.go:45` `*RenderError` (carries `Diagnostic`, `Err`; `:72`
  `newRenderError(code, elementID, dataPath, err)`); `folio-go/diagnostic.go:128-341` the 17-code
  closed registry — **no code exists for an asset media-type failure**.

**The canvas — measurement is already shared; paint is not**
- `folio-go/page_setup.go:1130` `addCanvasTextPaint` — calls the render path's own `fontChain`
  (`:1162`), `shapeSegments` (`:1192`), `chainVerticalModel` (`:1229`), `positionSegments`
  (`:1288`/`:1298`). Comment `:1225-1227`: *"the canvas consumes the IDENTICAL advance the renderer
  does … the browser never measures text."* **So the canvas inherits embedded resolution for free
  the moment `fontChain` stops dropping — and that is what AC5's Then-clause asks for.**
- `ARCHITECTURE-SPINE.md:361` **AD-17** — *"The browser never measures text, including on the
  canvas … The browser contributes rasterization only."*
- `folio-designer/src/App.css:118` `.canvas-text-fragment` — `font-family: 'IBM Plex Sans', 'IBM
  Plex Sans Thai', 'IBM Plex Mono', sans-serif`. **Four literals, no `var()`, no document input.**
  Contrast `:116` `.canvas-text-paint`, which *does* take document values (size, weight, style, ink).
- `folio-designer/scripts/build-wasm.mjs:68-70,79` — the three shipped **Noto** files registered
  under **IBM Plex** family names (and `IBM Plex Mono` → the **CJK** file). Engine face names are
  `Noto Sans` / `Noto Sans Thai` / `Noto Sans SC` (`folio-go/fonts/fonts.go:57-63`). **Zero
  intersection**, pinned by `canvas-font-stack.test.ts:100-106`.
- `folio-designer/src/canvas-font-stack.test.ts` — three tripwires, all verified present:
  stylesheet-constant `:85-98`, non-intersection `:100-106`, no-chain-entry-to-CSS-family `:108-121`
  (self-limited at `:110-115` as a single-line scan). **If the generated family names ever change,
  `:62-65`, `:100` and `:126` must be re-authored, not merely re-run.**
- `folio-designer/src/canvas-authority-contract.test.ts:18-51,155` — scans **all** production, unit
  and e2e sources for `measureText`, `getBoundingClientRect`, `getComputedStyle` etc. Red-proved at
  `:200-206`. Note `:145` neutralizes `document.fonts` before the scan, so a `FontFace` injection
  would trip **nothing** today.
- `folio-go/asset_bytes.go:19-23` `AssetBytes(t, key) ([]byte, string, error)` — key-shape gated,
  **no media-type restriction**; surfaced over the worker as the `'asset'` message
  (`engine-client.ts:142`), consumed at `App.tsx:1335`. **The byte channel a paint fix would use
  already exists and has never been used for fonts.**
- Projection today: `page_setup.go:164,181` `CanvasComponent.FontFamily` (the chain **name**);
  `:441` `CanvasFontChain`; `:487-492` `CanvasFontChainEntry{Face,AssetKey,Family,Style}`; `:545`
  `projectFontChainEntry`. Comment `:281-288`: *"the font BYTES are projected by nothing."*
  TS mirror `engine-protocol.ts:171`, guards `:220-225` (`isFontChainEntry`, exact-key + XOR
  discriminant), `:294`. **No per-fragment face attribution is on the wire.**

**Subsetting**
- `folio-go/render.go:1818-1836` — `glyphsByFace` union `:1818-1824`, deterministic face order
  `:1826` (`slices.Sorted`), `font.Subset(...)` at `:1834` (**the only non-test call site**),
  `pdf.EmbeddedFace` populated `:1843-1859`. Once per render, once per face.
- `folio-go/internal/fontset/fontset.go:730` `Subset`; `:910-924` `deriveTag`. **Measured: the tag
  is FNV-1a over the COMPLETE emitted subset program bytes, NOT the glyph set** — and the
  glyph-id-set reading was **explicitly rejected** (comment `:885-909`: *"B6: two pinned instances
  of one face collided under that reading"*, and D-1.5.8). See Design Notes.

**The fixture surface — the complete registration checklist**
- `fixtures/embedded-font/` — `input.folio` (903 lines), `expected.json` (hash
  `db400698…e513ad`), `README.md`. **No `expected.pdf`.** The chain is `["Noto Sans", {"asset":
  "c94562c1…3caf"}]`; the embedded bytes are the shipped **Noto Sans Thai**; the page text is
  **Latin**, deliberately — *"the face the document carries is the one the page does not need"*.
  **So as it stands the embedded face still would not draw**: the fixture must change (Thai text, or
  a chain whose covering entry is the embedded one) for AC6 to mean anything.
- README's *"Why there is no `expected.pdf`"*: *"this story cannot produce the one that would matter
  — a page drawn with the embedded face … **When Story 8.4 lands, that guard must change,
  deliberately.**"*
- `folio-go/matrix_test.go:1963-1970` the `embedded-font` entry; `:1994-2007`
  **`requireEmbeddedFaceStaysOffThePage`**, which fatals unless exactly **one** font program is
  embedded. Doc `:1981-1993`: *"WHEN STORY 8.4 LANDS THIS GUARD MUST CHANGE."*
- **Registration sites for a golden that gains an `expected.pdf`:** (1)
  `byte_neutrality_test.go:92` `goldenDigestRecord` — a `dir` + `sha256` + every `site`; note the
  test **`os.ReadFile`s `expected.pdf` and errors "presence precondition" if missing**
  (`:601-614`), so the entry and the file are one commit. This automatically enrols the fixture in
  `TestEveryGoldenPDFResolvesItsPageTree` (`golden_structural_validity_test.go:90`) and
  `TestNoPreStory80GoldenCarriesATextRise` (`byte_neutrality_test.go:1396`). (2)
  `byte_neutrality_test.go:1363` **`textRiseExemptGoldens`** — **only if** the new golden's stream
  carries ` Ts`, which a page drawn with an embedded **Thai** face plausibly does; guarded both ways
  at `:1432` and `:1436`. (3) `matrix_test.go` — the entry already exists; the `extraGuard` changes.
  (4) `.github/workflows/matrix.yml:267` `docs="…"` and the four `hash.<target>.<slug>.txt` upload
  paths — already present for `embedded-font`; enforced by `matrix_registration_test.go:39,87-95`.
  (5) `byte_neutrality_test.go:853` `declaredEpic2GateObligations` — `matrix-document:
  embedded-font` **already declared**; a **new** obligation may not be added without a ruling saying
  so explicitly (failure text `:1046-1052`). (6) `embedded_font_fixture_test.go:317`
  `withAssets != 7` and `:307`'s `entry.Name() != "embedded-font"` literal — **both bite only if a
  SECOND fixture is added**; reusing `embedded-font` avoids them. (7)
  `missing_glyph_corpus_test.go:201-212` + `beyondBaselineAcceptance` `:249-258`, set-compared both
  ways at `:265-275`.
- `folio-go/embedded_font_fixture_test.go:73` `embeddedFontTemplateJSON()` derives `input.folio` from
  the shipped bytes; `:171-181` pins the committed file against it. Editing the fixture means
  editing the generator.
- **Not affected, stated so nobody chases them:** `TestCorpusMeetsP6ExerciseFloors` /
  `TestCorpusP6StatsMatchDeclaredBaseline` (`internal/text/corpus_test.go:169,243`) read **only**
  `fixtures/thai-break-corpus/corpus.json`; the declared baseline is a Go `const` block at
  `:244-252` (`baselineP6g = 7`). Adding a fixture directory never enters that population.
- **A real hole, recorded:** there is **no reverse-direction guard** — a fixture shipping an
  `expected.pdf` but absent from `goldenDigestRecord` is invisible to every test.

**Read-only evidence**
- Baseline red set must be **re-measured**, not assumed; it has moved three times in this run.
- oxlint baseline is exactly **4** `only-export-components` warnings; Vitest baseline **323 tests /
  35 files** at 8.3's close.
- `folio-designer/src/engine-bounds-mirror.test.ts:80-81` matches literal source text.
- `lint`'s `TestFloatTypedTestScopeInventory` pins five sites **by line number** in
  `shaping_expectations_test.go` and `internal/fontset/{fontset,vendorboundary}_test.go` — this
  story edits `internal/fontset/fontset.go` only if it must, and any line shift there reddens it.

## Tasks & Acceptance

**Execution:**

1. `folio-go/render.go` — replace `chainFaceNames`' unconditional drop with a resolution that
   carries the embedded arm, and give `fontCache`/`resolveRuneFace`/`shapeSegments` access to the
   document's font bytes. Pick **one** of the two seams named in the Code Map and state the choice
   in a doc comment. Rationale: this is the whole of AC1 and AC2.
2. `folio-go/render.go` — add the font analogue of the image loop (`:1873-1936`) inside
   `predictDocument`: dedupe embedded asset keys, sorted order, `firstElementIDByAssetKey`,
   `DecodeAssetBytes` → `template.DecodeFontForRender` → `fontset.New`. Rationale: this is AC4's
   Render half **and** AC4's `Validate` half, because `Validate` calls `predictDocument` directly.
3. `folio-go/internal/template/fontasset.go` — extend `UnsupportedFontMediaTypeError` (or add a
   located wrapper) so the message names the **chain** and the **entry index**, not only the asset
   key and element id. Rationale: AC4 requires both, and the built error carries neither.
4. `folio-go/table_render.go` — apply the same resolution at the header-style site (`:654-666`).
   Rationale: it is a second, independent chain→names conversion; leaving it silently keeps the old
   drop for table headers.
5. `folio-go/render.go` — confirm the synthetic face key for an embedded face (the asset key is the
   natural candidate) does not collide with a caller-supplied FontSet name, and state the precedence
   rule in a doc comment. Rationale: `fontCache`, `pagemodel.TextRun.Face` and `pdf.EmbeddedFace.Name`
   are all keyed on "the caller's FontSet key".
6. `folio-go/chain_face_names_test.go` — re-point the two existing mutation pins to the new
   boundary; keep them, never delete. Rationale: they are the red-proof that the drop is gone.
7. `folio-go/internal/template/fonts_embedded_test.go:927` — change
   `TestEmbeddedEntryIsInertUntilStory84` deliberately (rename and invert, or retire with the
   ruling that discharged it). Rationale: it asserts the zero-call-site fact this story removes.
8. `folio-go/matrix_test.go:1994` — change `requireEmbeddedFaceStaysOffThePage` deliberately to
   require the embedded program **on** the page. Rationale: the guard was written to force exactly
   this edit rather than let a golden move silently.
9. `fixtures/embedded-font/` + `folio-go/embedded_font_fixture_test.go:73` — change the document so
   the embedded face **actually draws** (the page text must need a rune only the carried face
   covers), regenerate `input.folio` from the generator, re-record `expected.json`, and ship
   `expected.pdf`. Rationale: AC6; and the fixture as it stands would not exercise the feature.
10. `folio-go/byte_neutrality_test.go:92` — add the `embedded-font` `goldenDigestRecord` entry with
    every site; check `textRiseExemptGoldens` (`:1363`) if the new stream carries ` Ts`. Rationale:
    AC6's "recorded digest like every other fixture".
11. **Tests for the I/O matrix**, each red-proved by reverting its production expression: mixed
    chain resolves in declared order; shared family name resolves by asset key; shipped-FontSet-only
    render; non-font asset **drawn** errors on **both** `Render` and `Validate` with identical
    strings (copy `chain_face_names_test.go:176-211`); non-font asset **never drawn** renders clean;
    load still accepts. Rationale: AC4's `Validate` half is the one that disappears without a test
    that fails when only the Render half is built — assert the Validate arm **separately**, so
    deleting the Validate call reddens a named test on its own.
12. `folio-go/page_setup.go` + a canvas test — pin that `CanvasWithTextPaint` on the embedded-face
    document produces fragment x-positions and advances **identical** to the PDF path's measurement.
    Rationale: AC5's Then-clause, and it reddens if anyone later carves the canvas out of the shared
    path. **Subject to the Block If — see Design Notes.**
13. `_bmad-output/specs/spec-folio/folio-format.md` — record that a chain entry naming a non-font
    asset is accepted at load and errors at render, if the format doc does not already say so.
14. `_bmad-output/implementation-artifacts/deferred-work.md` — close DW-83; record the disposition
    of DW-35 as ruled; record any new deferral. Rationale: DW-87 — the register is the audit trail.

**Acceptance Criteria:**
- Given a chain mixing an embedded and a shipped face, when text is shaped, then the embedded face
  joins per-rune coverage resolution in declared order, and where an embedded and a shipped face
  share a family name the **asset key** decides with no name-based substitution.
- Given a `FontSet` containing only the shipped set, when the document names an embedded face, then
  it renders from the document's own bytes, reading no network, no host-installed font and no path
  on disk.
- Given the same document on `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm`, when the
  outputs are compared, then they are byte-identical, subsetting happened once per render inside the
  PDF producer, and no face was subset at save time.
- Given a chain entry naming a non-font asset, when the document is **loaded**, then load accepts it.
- Given that same document, when something must actually **draw** from that entry, then `Render`
  returns a located error naming the chain, the entry index and the asset key.
- Given that same document and identical inputs, when `Validate` is called, then it returns the
  **identical** error — asserted by a test that fails if the `Validate` path is removed while the
  `Render` path remains.
- Given a chain entry naming a non-font asset that nothing draws from, when the document is
  rendered, then it renders clean.
- Given the designer's canvas paint projection and a component using an embedded face, when the
  preview is produced, then it measures with that same face through the same engine path, and its
  fragment positions equal the PDF path's.
- Given the embedded-face fixture, when it is added to the corpus, then it carries a recorded digest
  in `goldenDigestRecord`, and all 22 pre-existing digests are unmoved.

## Spec Change Log

## Review Triage Log

## Design Notes

**The measurement/paint seam — why this halts, and what I recommend.**
AC5 reads: *"the preview **measures** with that same face through the same engine path, so the
canvas and the PDF keep one **measurement** authority."* Under this project's own vocabulary that
is a precise claim: **AD-17** (`ARCHITECTURE-SPINE.md:361`) separates measurement from rasterization
by name — *"The browser never measures text … The browser contributes rasterization only"* — and the
epic context restates AC5 in the same words. Measured at HEAD, the measurement half is **already
structurally shared**: `addCanvasTextPaint` calls the render path's own `fontChain`,
`shapeSegments` and `chainVerticalModel` (`page_setup.go:1162,1192,1229`), so once Task 1 lands the
canvas measures with the embedded face **by construction**. Task 12 is the test that burden requires.

**But DW-35 is about paint, not measurement**, and the two are being treated as one:
- DW-35's register entry asserts *"Story 8.4's **AC4** … 'the preview measures with that same face …'
  — **is this entry written as an acceptance criterion**"*. Two errors: it is **AC5**, not AC4 (AC4
  is DW-83, per D-8.3.5), and DW-35's own stated fix is *"the canvas fragment rule must derive its
  family list from the projected component's own resolved chain rather than from a stylesheet
  constant"* — **rasterization**, which AC5's Then-clause does not mention.
- The provenance is a **recommendation, never a ruling**. It originates in Story 8.2's spec Design
  Note N3 (*"Recommended owner: Story 8.4"*), was copied into the register, and reaches this
  dispatch as a settled fact. **The decision log contains no ruling on it** — `awk` over
  `epic-7-8-decision-log.md` finds zero occurrences of `DW-35`.
- That same note calls the fix **"a design-system decision above a builder's authority"**: the
  browser's `@font-face` families are `IBM Plex *` while engine face names are `Noto *`, with **zero
  intersection** (`canvas-font-stack.test.ts:100-106`), so a chain's entries cannot be used as CSS
  family names at all. And for an **embedded** face there is no shipped file to rename — closing it
  needs a runtime `FontFace`/`@font-face` injection fed from `AssetBytes` (`asset_bytes.go:19`), a
  channel that exists for images and **has never been used for fonts**.

**Why this is not mine to rule.** 8.4 does not merely inherit DW-35; it **widens** it into a new
class. Today an embedded entry is dropped, so the canvas paints with whatever the shipped chain
resolves to. After Task 1 the engine measures with the embedded face while the browser paints with a
fixed IBM Plex stack — and for an arbitrary embedded face there is **no declared family at all**, so
the browser falls through to `sans-serif`. That is precisely the owner-reported defect fixed at
`c6e4d03` (Thai overlapping at every space), re-created for this story's own headline use case.
Absorbing the fix needs the design-system decision; **routing it out** to a named story needs
authorisation that neither this dispatch nor the register grants (my memory of D-8.0.5 and the 7.4
precedent: routing a goal out requires explicit authorisation); and ruling AC5 as measurement-only
leaves DW-35 with **no named owner**, since 8.4 was its last one. Three dispositions, none available
at a builder's authority — so it is bundled into the halt with a recommendation rather than decided.

**My recommendation, stated so the lead can ratify rather than re-derive:** rule AC5 as the
**measurement** claim (Task 12 delivers and pins it), and re-own DW-35's **paint** half to a named
story with the design-system question framed for the owner — the natural candidate is Story 8.5/8.6,
where a catalogue face is picked and the canvas must show it. That keeps the seam clean in the same
way Story 7.4 kept the product half and gave away the format half.

**AC3's false premise — measured, and it must not be implemented literally.** AC3 says the subset
tag *"remains a deterministic hash of the **glyph set**"*. It is not one and never was:
`deriveTag` (`internal/fontset/fontset.go:910-924`) is FNV-1a over the **complete emitted subset
program bytes**, and its doc comment `:885-909` records that the glyph-id-set reading was
**deliberately rejected** — *"two pinned instances of one face collided under that reading, since
instancing changes outlines without changing which glyph ids are retained"* (D-1.5.8). Implementing
AC3 literally would re-introduce that collision **and move all 22 golden digests**, which the epic's
own standing constraint (*every existing golden renders byte-identically after every story here*)
forbids. **Ruled:** read AC3 as *"the tag remains a deterministic function of the emitted subset
program"* — satisfied today, asserted rather than assumed. Recorded here as an epic-text correction
per D-8.2.3, with the original wording quoted, rather than silently reinterpreted.

**AC4's third half, and how to make it undroppable.** `Validate` predicts `Render` because both run
`predictDocument` (`validate.go:68`; `render.go:1606`). So wiring the font decode **inside**
`predictDocument` (Task 2) makes the `Validate` half true by construction. The failure mode
D-8.3.5 warns about is real but subtle: a build could satisfy "Render errors" by adding the check
somewhere `Render` reaches and `Validate` does not — `renderDocument` or `buildPageModel`. Two
defences: `TestValidateNeverReachesRenderOrInternalPDF` (`render_arch_test.go:326`) makes that
placement structurally visible, and Task 11's **separate** Validate assertion reddens on its own.
Note the image precedent has exactly this omission — `render_image_test.go:258` asserts Render only,
never `Validate` — so the pattern to copy is `chain_face_names_test.go:176-211`, not the image test.

**No diagnostic code is minted.** D-7.8.1: *"a specific code is minted only when a named consumer
must BRANCH on it."* No consumer branches — `folio.Validate` is unreachable from both the wasm engine
(`wasm/engine.go:199` reparses, it does not call `folio.Validate`) and the designer (`'validate'` is
in the protocol union but **has no call site** in `folio-designer/src`), and the CLI
(`cmd/folio/main.go:235`) prints messages, not codes. The image precedent returns a bare wrapped
error with prose location. Follow it.

**`Covers:` omits ADs its own ACs state — D-8.2.8(a), third occurrence in four stories.** The epic's
line is `FR54 · AD-8, AD-14, AD-22, D-1.8.1, NFR1, NFR1.d, NFR1.e`. Missing: **AD-17** (*"the
browser never measures text, including on the canvas"* — AC5's entire substance), **AD-21**
(*"every feature ships its golden fixture"* — AC6's substance, and AC6's own text cites it), and
**FR33**, which AC2's own text cites parenthetically while `Covers:` names only FR54.

**The before-the-tag set (D-8.2.2) gains a candidate.** AC4's render refusal is an **unshipped
narrowing on the tagged surface**: a document whose chain names an image renders today (when another
entry covers) and errors after. `Render` is exported, so D-8.2.2(a) is satisfied and (b)'s
"unshipped narrowing" test fires. Recommend adding it to D-8.2.7's set, which stands at two.

## Verification

**Commands** (from the repo root unless stated; `-count=1` on **every** Go gate — D-7.9.5, a warm
cache is not an anchor):
- **Before the first edit:** `shasum -a 256 fixtures/*/expected.pdf > <scratch>/digests.before` —
  expected **22** lines. Re-measure the **baseline red set** in a detached worktree at
  `d52b1b0` (`git worktree add --detach`), never by assumption: it has moved three times in this run.
- `cd folio-go && go test -count=1 ./...`
- `cd folio-go && go test -count=1 -tags=matrix ./...` — the **full** sweep, not a subset.
- `cd folio-go && go vet -tags=matrix ./...` — expected clean.
- `gofmt -l folio-go` — expected empty (`lint/…/licencegraph_test.go` is a known pre-existing
  offender, DW-23, outside this path).
- `cd folio-go && for T in darwin/arm64 linux/amd64 linux/arm64 js/wasm; do FOLIO_MATRIX_TARGET=$T
  go test -count=1 -tags=matrix -run TestTargetRenderHash -v .; done` — **all four AD-21 targets.**
  `TestTargetRenderHash` **asserts nothing unless `FOLIO_MATRIX_TARGET` is set**, so also run the
  unset control and report it as the deliberate no-op it is. Report all five legs with timings — the
  timing contrast is the evidence the legs actually asserted.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .`
- `cd lint && go test -count=1 ./...` — **`-count=1` is mandatory**: the rules package walks the
  `folio-go` tree with `ReadDir`, which Go's test cache does not track, so a cached `ok` here is no
  measurement at all.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` —
  typecheck clean; oxlint exactly **4** pre-existing `only-export-components` warnings; Vitest at or
  above the **323 tests / 35 files** baseline (report actual numbers); e2e **compiles only** and
  executes nowhere by design.
- `shasum -a 256 fixtures/*/expected.pdf | diff - <scratch>/digests.before` — the golden digest
  diff. Expected: **empty for the 22 pre-existing digests**; a 23rd line appears only if this story
  ships `fixtures/embedded-font/expected.pdf`, and that addition must be deliberate and recorded.
- `cd folio-go && GOOS=js GOARCH=wasm go test -count=1
  -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" ./wasm/cmd/engine/` — this package is invisible
  to `go test ./...` (it is `//go:build js && wasm`), so "it compiles" is not evidence.

**There are exactly TWO standing reds. Both are named; neither may be hidden, skipped or "fixed".**
1. `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` — got 7, need ≥20, under `./...`. The
   mandated corpus-sourcing red, with its drift twin `TestCorpusP6StatsMatchDeclaredBaseline`
   (`internal/text/corpus_test.go:169,243`; `baselineP6g = 7` at `:251`).
2. `TestShippedFacesReproduceFromUpstream` — under `-tags=matrix` **only**
   (`fontgen_matrix_test.go:64`), failing with `fontgen: fontTools is not importable`. Registered at
   Story 8.3's close as **DW-86 (D-8.3.4)**. It **fails rather than skips deliberately**
   (`fontgen_matrix_test.go:26-32,68-82`) — *"the sources were not present"* must never read as
   *"the faces reproduce"*. Environmental and pre-existing; confirmed identical at baseline `f51dd5e`
   with sources supplied. **Do not claim it is pre-existing without checking it at this story's own
   baseline.**

**Any third red is a real failure. Report it; do not absorb it.**

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

Status: blocked
Blocking condition: intent gap

Plan-gate dispatch, halt-after-planning. Baseline `d52b1b073c621c9e4f3c7ea2c1091f93f96294fb` on
`main`. No code written, no commit, no branch, nothing staged.

**The spec above is complete and investigation-drained.** It halts on one bundled question, not on
missing work: re-dispatch after the ruling should need planning only for what the ruling changes.

### The gap (ONE question, with two sub-parts kept together on purpose)

**Q1 — Does AC5 oblige this story to close DW-35's canvas PAINT gap, or only the MEASUREMENT
claim its Then-clause states?**

Two defensible readings with observably different outcomes:
- **(A) Measurement.** AC5's Then says *"the preview **measures** with that same face through the
  same engine path, so the canvas and the PDF keep one **measurement** authority."* AD-17
  (`ARCHITECTURE-SPINE.md:361`) separates measurement from rasterization by name, and the epic
  context restates AC5 in the same words. Measured: the canvas already shares `fontChain`,
  `shapeSegments` and `chainVerticalModel` (`page_setup.go:1162,1192,1229`), so this half becomes
  true by construction once the drop is removed, and needs a test (Task 12), not a feature.
- **(B) Paint.** DW-35's register entry asserts this AC *"**is** this entry written as an acceptance
  criterion"*, and DW-35's stated fix is that *"the canvas fragment rule must derive its family list
  from the projected component's own resolved chain rather than from a stylesheet constant"* —
  rasterization, which AC5's Then-clause does not mention.

**Why I did not rule it.** Reading B requires a face-name → CSS-family mapping that exists on
neither side (`App.css:118`'s four IBM Plex literals vs. the engine's Noto names — **zero
intersection**, pinned at `canvas-font-stack.test.ts:100-106`) plus, for an embedded face, a runtime
`FontFace` injection from `AssetBytes` that has never existed for fonts. **Story 8.2's own Design
Note N3 — the source of the DW-35→8.4 assignment — calls that "a design-system decision above a
builder's authority."**

Three dispositions exist and **none is available at a builder's authority**: absorbing needs that
design-system decision; **routing DW-35 out** to a named story needs authorisation neither this
dispatch nor the register grants; and ruling (A) leaves DW-35 with **no named owner**, 8.4 having
been its last one.

**And 8.4 actively WIDENS the defect, which is why this cannot simply be recorded and passed on.**
Today an embedded entry is dropped, so the canvas paints with whatever the shipped chain resolves
to. After Task 1 the engine measures with the embedded face while the browser paints a fixed IBM
Plex stack — and for an arbitrary embedded face **no family is declared at all**, so the browser
falls through to `sans-serif`. That is the owner-reported defect fixed at `c6e4d03` (Thai
overlapping at every space), re-created for this story's own headline use case.

**Recommendation (stated so the lead can ratify rather than re-derive):** rule AC5 as the
**measurement** claim — Task 12 delivers and pins it — and re-own DW-35's **paint** half to a named
story, framing the design-system question for the owner. The natural candidate is 8.5/8.6, where a
catalogue face is picked and the canvas must show it. Same clean seam as Story 7.4's keep-the-
product-half / give-away-the-format-half split.

### Two premises in the dispatch, corrected against the tree

1. **"the fourth AC is DW-35 stated as an acceptance criterion" is FALSE.** The fourth AC is
   **DW-83** — the non-font-asset three-halves criterion — which is exactly what the dispatch's own
   first premise says, and what **D-8.3.5** rules. DW-35 corresponds to the **fifth** AC (the canvas
   one). The register entry makes the same off-by-one, calling it "Story 8.4's AC4" while quoting the
   canvas AC. DW-83 and DW-35 are both re-owned here, but by **different** criteria.
2. **The DW-35→8.4 assignment was never ruled.** It originates as *"Recommended owner: Story 8.4"* in
   Story 8.2's spec Design Note N3, was copied into `deferred-work.md`, and reached this dispatch as
   settled. `awk` over `epic-7-8-decision-log.md` finds **zero** occurrences of `DW-35`. The register
   entry also carries **two contradictory `Owner:` bullets** (Story 8.2 and Story 8.4), an editing
   artifact worth repairing whichever way Q1 is ruled.

### Ruled at the gate, recorded rather than escalated

- **AC3's premise is factually false and must not be implemented literally.** AC3 says the subset tag
  *"remains a deterministic hash of the **glyph set**"*. Measured: `deriveTag`
  (`internal/fontset/fontset.go:910-924`) is FNV-1a over the **complete emitted subset program
  bytes**, and its doc comment `:885-909` records the glyph-id-set reading as **deliberately
  rejected** — *"two pinned instances of one face collided under that reading"* (D-1.5.8).
  Implementing it literally re-introduces that collision and moves all 22 golden digests, which the
  epic's own byte-identity constraint forbids. Ruled as *"a deterministic function of the emitted
  subset program"*; recorded as an epic-text correction per D-8.2.3 with the original wording quoted.
- **No diagnostic code is minted** (D-7.8.1): no consumer branches. `folio.Validate` is unreachable
  from the wasm engine (`wasm/engine.go:199` reparses rather than calling it) and has **no call site
  at all** in `folio-designer/src`.
- **`Covers:` omits ADs its own ACs state — D-8.2.8(a), now the third occurrence in four stories.**
  Missing **AD-17** (AC5's entire substance), **AD-21** (AC6's substance, cited in AC6's own text),
  and **FR33** (cited parenthetically in AC2's own text while `Covers:` names only FR54).
- **D-8.2.2 / D-8.2.7:** AC4's render refusal is an **unshipped narrowing on the tagged surface**
  (`Render` is exported; a document whose chain names an image renders today and errors after), so it
  is a candidate for the before-the-tag set, which stands at two.
- **AC6 is achievable without a human sign-off**, on Story 8.0's measured precedent: a new golden
  invalidates no existing attestation and leaves no `//go:build matrix` sign-off gate red. No agent
  writes `reader`/`date`/`examined`. If the embedded-face page's subject turns out to be a new
  irreducibly-human property, that is a **deferral** with the README stating the gap — not a halt.

### Findings the plan gate measured that the dispatch did not carry

- **`chainFaceNames` is NOT the single boundary.** `table_render.go:654-666` performs a second,
  independent chain→names conversion for table **header** styles and does not route through
  `fontChain`. Changing only `chainFaceNames` silently leaves table headers on the old behaviour.
- **The real seam is wider than one function.** `resolveRuneFace`, `shapeSegments`, `chainLineMetrics`
  (`wrap.go:550`), `digitTableRun` (`page_number.go:436`) and `fontCache` all take
  `(chain []string, fs FontSet)` and hold **no document**. Per **D-8.1.3**, "route it through X" is a
  claim to verify: there is no single authority here, and choosing between widening the four
  signatures and materialising a per-render name→bytes view is real design work inside the story.
- **The built error does not name what AC4 requires.** `UnsupportedFontMediaTypeError`
  (`internal/template/fontasset.go:25`) carries `AssetKey`, `ElementID`, `MediaType` — **not** the
  chain name or entry index. It must be extended or wrapped.
- **The fixture as it stands would not exercise the feature.** `fixtures/embedded-font/input.folio`
  draws **Latin** text while carrying a **Thai** face, deliberately (*"the face the document carries
  is the one the page does not need"*). AC6 is vacuous unless the document changes so the embedded
  face actually draws.
- **The `Validate`-half trap is concrete, and the image precedent is the wrong thing to copy.**
  `render_image_test.go:258` asserts `Render` only and never `Validate` — the exact omission D-8.3.5
  warns about, already present in the tree. Copy `chain_face_names_test.go:176-211` instead, which
  asserts both surfaces with string equality.
- **A structural hole, recorded not fixed:** there is **no reverse-direction guard** — a fixture
  shipping an `expected.pdf` but absent from `goldenDigestRecord` is invisible to every test.
- **The dispatch's P6 premise needed checking and holds for the wrong reason:** the P6 tests read
  **only** `fixtures/thai-break-corpus/corpus.json` (`internal/text/corpus_test.go:31`), so adding a
  golden fixture cannot move them. The declared baseline is a Go `const` block at `:244-252`, not a
  file.

### Resume note for the orchestrator

The workflow's designed resume path for an intent-gap halt is `status: draft`, which routes step-01
to step-02 and triggers step-02's Draft resume check — that check preserves the
`<intent-contract>` block verbatim. Re-dispatching against this file while it reads `status:
blocked` halts immediately at step-01 with `blocked spec supplied`. So: amend the
`<intent-contract>` with the ruling, set `status: draft`, re-dispatch.
