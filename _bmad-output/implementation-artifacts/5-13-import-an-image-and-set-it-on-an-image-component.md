---
baseline_commit: 3a52ae4
story_key: 5-13-import-an-image-and-set-it-on-an-image-component
status: done
created: 2026-08-29
---

# Story 5.13: Import an image and set it on an image component

**Epic:** 5 — A template author can lay out a report and see the real PDF
**Covers:** FR4, FR5, FR33 · UX-DR10, UX-DR21, UX-DR24, UX-DR25

## In plain terms (read this first if you just want the gist)

An author can now select a placed image, choose a `.png` or `.jpeg` from their own machine, and see
their own picture on the page — not the literal word "Image" a placed logo used to show. The engine
installs those exact bytes into the document's content-addressed asset store and points the element
at the result; choosing the same picture twice stores it once; the canvas paints inside the same
fit-and-centre rectangle the PDF uses, never a browser-computed approximation. The file is embedded,
never linked — nothing here fetches by URL or reads the filesystem at render time, so a saved file
stays one portable artefact. Go alone decides what bytes are legal; the browser only reads and hands
them over.

The engine work here was genuinely solid on a first pass. What was not, at first, was the evidence
for three of its own headline promises: a test meant to prove "replacing an image never deletes an
unrelated asset elsewhere in the document" didn't actually contain an unrelated asset to protect, so
it could not have caught the mistake it was written for; a browser test meant to prove the new file
picker survives a specific known defect used a stand-in picker too forgiving to reproduce that
defect; and the strongest claim about where the picture gets painted had no test actually reading
the position and size of the painted picture. All three were corrected and each is now demonstrated,
live, to catch the exact mistake it exists to catch — not merely asserted to. A fourth real gap, also
closed here: if an author replaced a picture and then closed the document (or opened a different one)
in the moment before that replacement finished, the stale result could have landed on the wrong
document; it no longer can. Along the way, a separate, already-known defect — a placed picture
couldn't be dragged, because the new picture element intercepted the browser's own mouse gesture —
was confirmed fixed and is now covered by a real drag test, alongside a new one for resizing and
zooming the picture.

One accepted rough edge ships on purpose: each picture on the page currently asks the engine for its
bytes separately, with no shared cache, even when several elements show the same picture. That's a
cost the design always expected to pay somehow; today's shipped form of it just wasn't measured
against a real document, so it's recorded to revisit rather than guessed at now. Not here, still:
binding an image to data, an asset library, cropping, rotation, opacity, alt text, SVG, or removing
the shipped default picture from the insert path.

## Story

As a template author,
I want to put my own logo on the page,
so that the statement I lay out is my organisation's document rather than Folio's.

## Acceptance criteria

### AC1 — A chosen local image becomes a content-addressed document asset through one engine command

**Given** one selected image element and an image file chosen from local disk,
**When** the author sets it,
**Then** the browser sends one opaque, committed command carrying the file's raw bytes and its
declared media type to the existing Go/wasm document owner, and only a successful response installs
the next snapshot.

**And** Go computes the lowercase-hex SHA-256 of the decoded bytes, inserts the asset under that key
only if absent, base64 hard-wraps at 76 columns into an array of strings per AD-9, sets the element's
`asset` to the key, and does all of it inside the existing serialize/reparse/project transaction —
advancing the document revision and engine-side undo history exactly once when canonical bytes
change, invalidating Preview through the existing authority path, and participating in save/undo/redo
like any other document mutation. The same picture chosen for two elements stores one asset and two
references.

**And** Go/wasm owns command-envelope legality, target eligibility (an image element and nothing
else), byte decoding, digest computation, media-type recognition, size bounds, canonical mutation and
rejection diagnostics. The browser must not hash, sniff, validate, or maintain a TypeScript asset
model. A rejected command leaves canonical bytes, revision, undo/redo branches, selection truth and
displayed asset state unchanged, and no asset is left half-inserted.

**Proof / red proof.** Prove one successful set reaches the real worker command channel, serializes
canonically, survives save/load and `Parse(Serialize(d)) == d`, and is undoable/redone as one
mutation. Prove the dedup path: two elements, one picture, one entry in `assets`. Prove a media type
this version cannot decode is refused at the command with a located diagnostic and never enters the
document — the load-time contract in `decodeAssets` must never be the thing that catches it. Red-prove
a browser-computed digest, two commands for one set action, a partially-inserted asset after a
rejected command, a command result installed after document replacement, and an asset whose key is not
the SHA-256 of its own bytes reaching canonical bytes.

### AC2 — The inspector exposes an IMAGE section that is honest about what is set and what failed

**Given** a selection that is exactly one image element,
**When** the properties panel renders,
**Then** it shows an IMAGE section between CONTENT and BOX carrying the current asset's identity from
the engine snapshot — its media type, intrinsic pixel dimensions, and its key **abbreviated for
display** — plus one named
control that opens the local image picker.

**And** the section is keyboard-reachable and operable with visible `colors.select` focus and an
accessible name on every icon-only control. Unavailable and failed states state the concrete reason
in the product's terse register (UX-DR24) — no picker capability in this browser tier, a media type
this version cannot render, a file too large, a selection that is not a single image element — and are
distinguished by shape before colour, never colour alone. The panel's standing footer note loses its
"Asset import ... not editable here" clause, because it stops being true.

**And** the section appears only for image elements: a multi-selection, a text/table/line/rect
selection, and an empty selection are unchanged. Preview mode is unchanged.

**Proof / red proof.** Cover the section's presence and absence across single-image, multi-image,
mixed, non-image and empty selections; the keyboard path from selection to picker to committed set;
named controls; focus visibility; and announced failure text for each unavailable reason. Red-prove
the section appearing for a non-image selection, a colour-only error state, an icon-only control with
no accessible name, and the footer note still claiming asset import is unavailable.

### AC3 — The canvas paints the placed image inside Go-owned geometry

**Given** an image element whose asset is installed,
**When** the canvas paints the current engine snapshot,
**Then** it draws the image inside the fit-and-centre draw rectangle the engine already computes for
the PDF (`resolveImagePlacement`), supplied on the canvas projection as Go-owned millipoints, mapped
through the existing zoom rule. The browser must not compute the fit itself, must not reach for
`object-fit` as the authority, and must not read intrinsic dimensions from a decoded DOM image to
decide placement.

**And** the projection stays a lossy paint-only projection under AD-17: it carries what is needed to
draw this image and nothing that reconstructs canonical bytes or the assets map. Object URLs minted
for painting are revoked on element deletion, asset replacement and document replacement, with no
accumulation across a session. An element whose asset is missing or undecodable paints an honest
placeholder naming the reason, not a blank box and not a crash.

**And** UX-DR21 survives: the canvas remains explicitly approximate about text. Painting an image
exactly where Go says does not license the canvas to start deciding any other geometry.

**And** four specifics bind, because leaving them open is where this drifts (D-5.13.2):
*Frame* — `collectImageRuns` builds `y: layout.PlaceInBand(b.origin, el.Y)`, page-absolute, while a
canvas component's `X`/`Y` are band-relative. Build the paint's run in the **band frame** and call the
same `resolveImagePlacement`; the centring offsets are pure functions of box against image, so the
result is a translation of the render's — assert that equality, never assume it.
*Producer* — the paint is produced by a sibling of `addCanvasTextPaint` invoked from
`CanvasWithTextPaint`, not inside the command, whose own `Canvas(t)` is discarded and recomputed by
`wasm/engine.go`. Every derived coordinate routes through `canvasDerived`/`canvasDerivedSum` as the
text paint does, or a large document emits a JS-unsafe number instead of a bounded error.
*Two-producer proof* — the established shape is `TestCanvasTextPaintExactlyMatchesTheShippingRunPath`:
the paint is proven equal to **the shipping run path**, not merely computed from the same function.
Ship the image analogue by that name-shape, or the test can pass with both sides moving together.
*Absence, not zero* — `DecodedImage.Width()/Height()` are reachable only through
`decodeRecognisedImage`, so a legally-loaded asset of an unrecognised media type has no dimensions and
no computable rectangle. The paint field is **absent** in that case, never zero-valued, and AC2's
failure text and this AC's placeholder are driven by that one Go-side signal, not two.

**Proof / red proof.** Prove the painted rectangle equals the engine's draw rectangle for
wider-than-box, taller-than-box and exactly-proportioned assets, and that it tracks a resize. Prove the
band-frame/page-absolute translation equality explicitly. Prove URL revocation on delete, replace and
document replacement. Red-prove a browser-computed fit, a paint that reads DOM-measured intrinsic
dimensions, a leaked object URL, a projection field that mirrors the assets map, a zero-valued
rectangle standing in for an absent one, and a missing asset rendering as an empty box.

### AC4 — The asset map stays canonical and the scope boundary holds

**Given** an image element repointed from one asset to another by this command,
**When** the document is serialized,
**Then** the command collects **only the previous asset key of the element it just repointed**, and
only when no other element still references that key after the mutation — so replacing a logo ten
times leaves one asset and not eleven.

**And** it is never a document-wide sweep (D-5.13.3). A document may **legally** carry an unreferenced
asset, and that is a shipped, asserted state: `render_image_test.go`'s
`TestRenderUnrecognisedMediaTypeErrorsOnlyWhenDrawn` requires an orphaned unrecognised-media-type asset
to load and render clean — RP-11's positive control. An unqualified sweep would mean opening such a
document, setting any image, and silently losing the author's asset from their own file.

**And** undo needs no new machinery and must not grow any. History is byte-snapshot replay —
`Engine.undo` holds canonical byte strings and `restore` reparses one whole — so undoing the mutation
that collected an asset restores it by construction. **Prove that; do not build it.**

**And** the shipped default mark stays exactly where it is: inserting from the palette still attaches
`defaultAuthoringLogoKey` so a dropped Image is renderable before anyone picks a file. This story adds
the replacement path; it does not change the insert path, and it does not make an image element
legally asset-less.

**And** FR33 holds without exception. No URL field, no filesystem path, no remote fetch, and no
render-time IO is introduced on any path. Nothing added here implies a server round-trip (UX-DR23).
Excluded and remaining later work: binding an image to data, an asset library/manager surface,
crop/rotate/opacity/alt-text, SVG or any other unrecognised media type, image elements in table cells,
and export of an embedded asset back to disk.

**Proof / red proof.** Prove collection across repeated set/replace on one element, and that undo
restores a collected asset's bytes through the existing replay. Prove `TestRenderUnrecognisedMediaType
ErrorsOnlyWhenDrawn` still passes untouched. Prove the palette insert path is byte-identical to today.
Grep-level red proof that no URL, path or fetch reaches the image path. Red-prove a document-wide
sweep, an asset collected while another element still references it, an undo that cannot restore
collected bytes, and an image element serialized with no asset.

## Tasks / subtasks

- [x] **1. Add the closed engine-owned asset command** (AC: 1, 4)
  - Extend the existing versioned component-command vocabulary in `folio-go/component_commands.go`
    and route it through `wasm.Engine.Apply`. Do not add an `asset` key to `applyPropertyChanges`'s
    `allowed` map or to `propertyOrder`. The structural ground, which is firmer than "it carries
    bytes" (D-5.13.1): `propertyChange` admits exactly one or two keys of the form `{op, value?}` with
    `op` in `{set, clear, null}`, and `propertyPath` returns a single key as the diagnostic location.
    An asset payload is two correlated values — bytes plus declared media type — and AC4's rule that an
    image element is never legally asset-less means `clear` and `null` must be **inexpressible** for
    it. Admitting `asset` would force a second grammar into `propertyChange` or push base64 through
    `propertyString`, and the closed property vocabulary would stop being closed.
  - Decode the payload, compute the SHA-256 key, recognise the media type through the existing
    `image.go` capability predicate rather than a second copy of it, insert-if-absent, repoint the
    element, collect orphans, then serialize/reparse/project transactionally and install only when
    canonical bytes differ. Preserve diagnostic shape, bounded element/path fields, revision
    monotonicity, undo/redo semantics and failed-command non-mutation.
  - Derive the byte-size bound from the existing `MAX_ENGINE_PAYLOAD_BYTES` envelope (8 MiB,
    `engine-protocol.ts`) on an honest **host-memory** ground — wasm linear-memory headroom — so an
    oversized file is a located diagnostic and never an OOM. Do **not** mint a fresh number "in the
    style of" `maxImagePixelDimension`: that constant's justification is arithmetic (int64 overflow in
    `ScaleRound`), a byte bound has no such derivation, and borrowing the style would dress a
    memory judgement as a proof (D-5.13.4). Enforce it in Go; the browser must not pre-reject on size,
    and the protocol envelope and the author-facing diagnostic must not disagree about the threshold.

- [x] **2. Extend the paint-only worker snapshot and protocol narrowly** (AC: 1, 2, 3)
  - Add only the Go-produced image display data the panel and canvas need: media type, validated
    intrinsic pixel dimensions, the **full 64-hex** asset key, and the fit-and-centre draw rectangle
    in millipoints. The key is full width on the wire because the per-key bytes request below needs a
    real lookup token; a prefix is either ambiguous or forces the browser to hold a key table, which
    is the assets-map mirror this invariant exists to prevent (D-5.13.2 amendment). The inspector
    abbreviates for display only, the way it renders millipoints as "12.5pt". **The snapshot carries no asset bytes** — an earlier draft of this task asked for
    "the bytes-or-handle the canvas paints from", which contradicts AC3's own rule against carrying
    anything that reconstructs the assets map (D-5.13.2). Paintable bytes arrive on a separate,
    explicit, **per-key** request; the `EngineSuccess.bytes` channel that already exists outside the
    snapshot for save is the precedent to reuse. One asset's bytes on demand is not the assets map;
    the assets map on every snapshot is. If a cleaner per-key seam exists, the invariant binds, not
    this mechanism. Update strict worker/protocol admission and deep-freeze coverage together so
    unknown, malformed, oversized or schema-mirroring projections are rejected.
  - Keep the snapshot lossy. TypeScript must not reconstruct the assets map or canonical bytes from
    it. Preserve FIFO request/response correlation, document-generation guards, one-worker ownership
    and Preview invalidation; a late result must not install across Open, Start blank, undo/redo or a
    newer authoritative snapshot.

- [x] **3. Add the image picker capability tier and the IMAGE inspector section** (AC: 2)
  - Follow AD-20 and the existing `sample-file.ts` precedent: one capability check selecting
    `showOpenFilePicker` or `<input type="file">`, exposing a narrow read-bytes interface with no
    save, no handle retention and no document semantics. Restrict the accept list to the media types
    `image.go` recognises today.
  - Build the IMAGE section in `ComponentProperties` from the DESIGN.md property-field and
    palette-item token counterparts. Present every unavailable and failed state honestly and
    keyboard-first, and remove the now-false asset clause from the footer note.

- [x] **4. Paint the image on the canvas from engine geometry** (AC: 3)
  - Replace the literal `'Image'` text in `CanvasComponent` with a paint driven by the projection's
    draw rectangle and the existing `canvasDisplay` zoom mapping. Own object-URL lifetime explicitly
    and revoke on delete, replace and document replacement.
  - Paint an honest, named placeholder for a missing or undecodable asset.

- [x] **5. Prove authority, transaction, visual-state and scope boundaries** (AC: 1–4)
  - Add focused Go command/serialization/dedup/orphan/history/diagnostic tests, wasm
    transaction/snapshot tests, protocol and worker-admission tests, and designer
    component/accessibility/Preview-freshness/static-ownership tests.
  - Add Playwright coverage for select image, choose a file, see it on the canvas, Preview, save,
    reopen, undo/redo — and **run it in this story** (`npm run test:e2e` from `folio-designer/`).
    D-000.4's per-story override is **granted for the designer e2e suite** and **declined for the
    four-target matrix** (D-5.13.5). The coverage must exercise the **File System Access tier**, not
    only the `<input type="file">` fallback: `eef7fbb` is a measured in-repo instance of that exact
    gap — a receiver-binding defect broke Open, Save and Load Sample JSON on that tier alone, passed
    the whole unit suite, and survived the entire Epic 5 Playwright gate because `vi.fn()` doubles
    have no brand check. Add a unit test asserting the new picker's **receiver**, following that
    commit's own precedent. Name the four-target matrix explicitly as written-but-unrun, due at the
    second Epic 5 close.
  - Run focused unit suites plus `npm run typecheck`, `npm run lint`, `npm run build`,
    `npm run test:e2e:compile`, relevant Go/wasm tests, `go vet ./...`, native and js/wasm builds,
    formatting/diff checks, and applicable lint/hashmatrix static gates. Ship a golden fixture per AD-21
    (D-5.13.7). Because the passthrough embedding path is unchanged, a plain render golden would
    near-duplicate `fixtures/image-embed/` — what this story changes is the canonical `.folio` bytes
    the authoring command produces. So `input.folio` must **be the captured canonical output of the
    real command**, and the fixture must assert **both** (a) the render matches `expected.json` in the
    `TestRenderMatchesGoldenFixture` shape, **and (b) re-running the authoring command on the source
    image bytes reproduces `input.folio` byte-for-byte** — committed literal against live computation.
    (b) is the load-bearing half; without it this is a second `image-embed`. Register in
    `matrixDocuments` with `folioGoVersion`/`goToolchain` matching the existing fixtures, prove it
    compiles under `-tags matrix`, and do not run the matrix. Hashes are derived Go-side from
    canonical bytes on the local target, never transcribed from a browser artifact or the e2e run.

## Developer guardrails

1. **One committed command, one document authority.** Setting an image is a single Go/wasm command and
   a single engine-side history event. No optimistic browser document state, no split command
   sequence, no browser inverse command, no direct template mutation.
2. **The browser reads bytes; Go decides everything else.** Hashing, format recognition, size bounds,
   dedup, orphan collection, mutation and diagnostics are Go's. A SHA-256 implementation appearing in
   `folio-designer/` is the defect this guardrail exists to prevent.
3. **`decodeAssets` is a last line, not the check — and is not to be tightened.** A file this version
   cannot decode must be refused by the *command* with a located diagnostic; relying on load-time
   validation means the author learns at reopen instead of at the click. Equally, it must not become
   the thing that is *changed*: D-1.8.1 (as amended) keeps `mediaType` an **open** set at the format
   level, and `image.go` defers the capability error to `DecodeImageForRender`. AC1's refusal is an
   authoring-command policy, never a format rule. "Aligning" the loader with the new command would
   break Story 1.8's shipped assertions and RP-11's positive control.
3a. **No decoder, no re-encoder, anywhere.** `image.go` is a chunk/segment walker by design — no
   stdlib or third-party image package exists in the module (AD-1/AC10). If obtaining intrinsic
   dimensions, normalising bytes or generating a canvas preview reaches for any image decoder or
   re-encoder, that is a new dependency **and** a new source of cross-target divergence: stop, and
   re-raise the matrix override before landing it. The AD-21 golden's hashes are derived Go-side from
   canonical bytes, never accepted from a browser-produced artifact.
4. **Snapshots are for paint only.** Expose the smallest Go-owned image paint data. No assets-map
   mirror, no browser-side template parse, no saved bytes derived from a snapshot.
5. **The canvas paints where Go says.** The fit-and-centre rectangle already exists in
   `resolveImagePlacement`. Project it. A second fit computation in CSS or TypeScript is a second
   source of truth about output geometry, which is exactly what AD-17 and AD-24 forbid.
6. **Embedded, never linked.** FR33 is absolute. No URL, no filesystem path, no fetch, no render-time
   IO, and nothing in the interface that implies a remote round-trip.
7. **Canonical form is not negotiable.** AD-9's key-is-the-digest rule, 76-column base64 wrapping,
   sorted keys and round-trip identity hold for every asset this command writes.
8. **Do not disturb the insert path.** The shipped default mark on palette insert stays. This story is
   additive to it.
9. **No later scope.** No data-bound images, asset library, crop/rotate/opacity/alt text, SVG, table-cell
   images, or asset export.
10. **Preserve unrelated work.** Do not modify `_bmad` configuration/manifest churn, `.agents/`,
    planning research, unrelated user changes, or fixtures, goldens and fonts belonging to **other
    stories**. **Carve-out:** this story's own AD-21 fixture is required by Task 5 and is explicitly
    not covered by this guardrail. The original blanket wording, inherited from the Story 6.2
    template, contradicted Task 5 outright and was read as forbidding the fixture — see D-5.13.8. Any
    story cloned from that template carries the same defect and must reconcile this guardrail against
    its own tasks before development starts.

## Project structure notes

- `folio-go/component_commands.go:1191-1206` is the current insert path: it embeds
  `defaultAuthoringLogoKey` (`:1217`) and sets `element.Asset`. `applyPropertyChanges` (`:719`) is the
  closed property vocabulary that deliberately has no `asset` key, and `propertyPath` (`:661`) its
  canonical diagnostic order. Extend the command vocabulary alongside them.
- `folio-go/internal/template/parse.go:330-412` (`decodeAssets`) already owns key-shape validation,
  strict base64 decoding, empty-asset refusal, digest-match enforcement and the lies-about-itself
  format check. `folio-go/internal/template/image.go` owns the recognised-media-type set
  (`decodeRecognisedImage`), the capability error (`UnsupportedMediaTypeError`) and the pixel bound
  (`maxImagePixelDimension`). Reuse both; do not restate their rules in the command.
- `folio-go/render.go:407` (`collectImageRuns`) and `:452` (`resolveImagePlacement`) already produce
  the per-element draw geometry the canvas needs. `render.go:1698` is the located missing-asset error
  whose wording the canvas placeholder should echo rather than reinvent.
- `folio-designer/src/engine-protocol.ts:55` is the strict canvas projection; components currently
  carry text/table display fields and no image fields. Add the smallest Go-owned extension and update
  `engine-worker-admission.ts` and its tests in the same change.
- `folio-designer/src/component-property-command.ts:4` is the opaque property-command factory;
  `component-command.ts` is the insert factory. A new asset factory belongs beside them and must stay
  equally opaque.
- `folio-designer/src/file/capability.ts` decides the file-access tier once for templates
  (`selectFileAccess`) and once for sample data (`selectSampleFileAccess`); `sample-file.ts` is the
  read-only, no-handle-retention precedent an image tier should follow.
- `folio-designer/src/App.tsx:798-820` is `ComponentProperties` and its section order;
  `App.tsx:903` is `CanvasComponent`, which currently paints the literal string `'Image'`;
  `App.tsx:660-663` is the tabbed inspector shell this section renders inside.

## References

- `_bmad-output/planning-artifacts/epics.md` — FR4/FR5/FR33 coverage map; Story 5.7 (palette) and 5.8
  (properties) as the closest prior art; UX-DR10, UX-DR21, UX-DR23, UX-DR24, UX-DR25.
- `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md` — FR4, FR5, FR33.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` —
  AD-9 (canonical bytes; assets keyed by SHA-256, 76-column wrapping, dedup), AD-14 (one diagnostic
  channel), AD-15 (engine owns the document), AD-16 (one worker), AD-17 (paint-only browser canvas),
  AD-20 (two-tier capability-detected file access), AD-21 (every feature ships its golden fixture),
  AD-24 (boxes are absolute).
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md:304` — KF-1 step 3,
  *"she ... drops the bank logo into the Page Header band"*: the key flow this gap currently breaks.
- `_bmad-output/implementation-artifacts/1-8-embed-and-render-an-image-from-the-template.md` — the
  format, passthrough-embedding, digest, dimension-bound and capability-error contracts this story
  authors against. D-1.8.1 (as amended), D-1.8.2, D-1.8.3, D-1.8.8.
- `_bmad-output/implementation-artifacts/6-7-the-round-trip-closes.md:146` — the review observation
  that the browser session never authored an embedded image. `defaultAuthoringLogoKey` entered in that
  story's finisher commit `44121b8` to make the palette control usable; this story is the other half
  it deferred.
- `_bmad-output/implementation-artifacts/5-8-edit-component-properties.md` — the property-panel,
  committed-commit and mixed-selection baseline the IMAGE section extends.
- `_bmad-output/implementation-artifacts/epic-5-boundary-gate.md` — the Epic 5 PASS is
  baseline-anchored evidence for Stories 5.1–5.12 at commit `4bd6331` and is not falsified by this
  story. It does mean D-000.4's per-epic cadence owes a second Epic 5 boundary once 5.13 lands, and
  that `epic-5` returns to `in-progress` in `sprint-status.yaml` until it does.

## Delivery log

### Development delivery — 2026-08-29

#### Implementation summary

- **Go command (AC1/AC4):** `setComponentAsset` added to `folio-go/component_commands.go` as its own
  top-level command kind (never a key in `applyPropertyChanges`/`propertyOrder`, per D-5.13.1). It
  decodes base64 bytes, computes the SHA-256 key, recognises/validates the media type by calling the
  existing `template.DecodeImageForRender` (image.go's doc comment updated to record this as its
  second call site, alongside the render path — no rule restated, no `decodeAssets` change), inserts
  the asset only if absent, repoints the element, and orphan-collects **only** the element's own
  previous key when nothing else references it (`assetKeyReferenced`, scoped per D-5.13.3). It runs
  inside its own serialize/reparse/project transaction, matching `updateComponentProperties`'s shape,
  and the size bound (`maxComponentAssetBytes`, 8 MiB, matching `MAX_ENGINE_PAYLOAD_BYTES`) is enforced
  on the decoded bytes before any document mutation.
- **Paint-only projection (AC2/AC3, D-5.13.2):** `CanvasImagePaint` added to `folio-go/page_setup.go`,
  produced by `addCanvasImagePaint` — a sibling of `addCanvasTextPaint`, invoked from
  `CanvasWithTextPaint`, never computed inside the command. It builds each run in the BAND frame
  (element X/Y untranslated) and calls the same `resolveImagePlacement` the render path uses;
  `TestCanvasImagePaintExactlyMatchesTheShippingRunPath` asserts the band-relative rectangle is
  exactly a translation of collectImageRuns/resolveImagePlacement's page-absolute one, across all
  three bands. The whole `Image` field is present only when the asset decodes (absence-not-zero); on
  absence the canvas shows a placeholder and the inspector shows one fixed failure message — one
  Go-side signal drives both, never two. `AssetKey` carries the FULL 64-hex key (not abbreviated): a
  per-key bytes request needs a real lookup token, so the wire field stays full and the inspector
  abbreviates it only for display — a coherence resolution of an internal tension between D-5.13.2's
  "abbreviated key" phrasing and AC3's own per-key-request mechanism, decided rather than parked
  because any other reading leaves the per-key fetch unable to address a real asset.
- **Per-key paintable bytes (D-5.13.2 "Producer"):** new `folio.AssetBytes(t, key)` /
  `wasm.Engine.AssetBytes(key)` read-only query, wired through a new `'asset'` engine protocol
  operation (`engine-protocol.ts`, `engine-client.ts`, `wasm/cmd/engine/main.go`) that reuses the
  existing `EngineSuccess.bytes` channel shape (matches `'serialize'`'s admission pattern). The
  snapshot itself never carries asset bytes.
- **Image picker tier (AC2, AD-20):** `folio-designer/src/image-file.ts`
  (`FileSystemImageAccess`/`InputImageAccess`), wired through `selectImageFileAccess` in
  `file/capability.ts`, reusing the SAME `currentBrowser()` binding that `eef7fbb` fixed — the picker
  is bound to `window`, not copied onto a plain object literal, so this tier does not repeat that
  defect. Accept list restricted to `image/png`/`image/jpeg`. No size pre-rejection (D-5.13.4: Go
  enforces the bound, not the browser).
- **IMAGE inspector section (AC2):** `ImageSection` in `App.tsx`, rendered between CONTENT/TYPOGRAPHY
  and BOX for exactly a single image selection. Shows media type/intrinsic px/abbreviated key from the
  snapshot when present, a fixed "cannot render this media type" message when absent, and states "no
  local file picker is available" when `imageFileAccess` is undefined. The footer note's now-false
  "Asset import ... not editable here" clause is removed.
- **Canvas paint (AC3):** `ImagePaint` in `App.tsx` fetches bytes per asset key via the `'asset'`
  operation, builds one object URL, and paints an `<img>` sized/positioned to the EXACT Go-supplied
  draw rectangle (no `object-fit`, no second fit decision). The effect's cleanup revokes the URL it
  created on every teardown — unmount (deletion), assetKey change (replacement) and a `generation`
  (documentGenerationValue) change (document replacement) — so no URL accumulates across a session.
- **Command/query factories:** `folio-designer/src/component-asset-command.ts`
  (`setComponentAssetCommand`, `assetBytesRequest`) — opaque, no hashing/sniffing, verified by a red
  proof that the SAME bytes produce identical output regardless of declared media type.
- **AD-21 golden fixture (Task 5, added after initial delivery — see Decisions, above):**
  `fixtures/component-asset-import/` pins the **authoring command's** canonical-bytes behaviour, not a
  second render of a document that already names an asset. `input.folio` is the captured canonical
  output of one real `setComponentAsset` command (`ApplyComponentCommand`) run against
  `imageTestTemplateJSON`'s base document (the same one `fixtures/image-embed/` renders), replacing its
  image element's asset with a real, valid, decodable 1x1 grayscale PNG — not hand-authored to
  resemble that output. `expected.pdf`/`expected.json` were recorded by rendering that document
  through the public `Render` path, exactly as every other fixture. Two new Go tests
  (`fixture_test.go`) assert both halves: `TestRenderMatchesComponentAssetImportGoldenFixture` (the
  render matches `expected.json`, same shape/vacuity-guard as every other golden test) AND
  `TestComponentAssetImportCommandReproducesTheFixtureInput` — re-running the real command against the
  same starting document and source bytes reproduces `input.folio` byte-for-byte. The second test is
  what makes this a 5.13 fixture rather than a second `image-embed`: it red-proofs the base64 wrap
  width, the digest-as-key derivation and the sorted-key ordering, none of which the render-only
  comparison alone can see. Registered in `matrixDocuments` (`matrix_test.go`, `//go:build matrix`) —
  new selector `subprocessComponentAssetImportEnvVar`/`captureComponentAssetImportRender` — matching
  the existing fixtures' `folioGoVersion`/`goToolchain`; `go build -tags matrix ./...` and
  `go vet -tags matrix ./...` both compile clean; the matrix itself was **not** run (D-5.13.5, stays
  declined for this story). Registering the new document also required two mechanical, pattern-
  dictated updates the repo's own guards demand of every new fixture: `.github/workflows/matrix.yml`'s
  `docs=` list and per-target hash-artifact paths, and `byte_neutrality_test.go`'s
  `declaredEpic2GateObligations` / `goldenDigestRecord` completeness lists (both guards failed loudly
  until updated, exactly as designed).

#### Decisions made without parking

- **AssetKey wire width (full vs. abbreviated).** Resolved as a coherence fix, not a design fork: the
  per-key bytes request is a MECHANISM the story itself specifies, and it cannot address an asset by a
  12-character prefix. Documented in `page_setup.go`'s `CanvasImagePaint` doc comment.
- **AD-21 golden fixture — corrected after initial delivery.** The first pass of this story skipped
  the fixture on two premises that the coordinator checked against the tree and found factually
  wrong: (1) that a new `fixtures/` entry is "the same apparatus" the declined four-target matrix
  consumes — it is not; `matrix_test.go`'s `matrixDocuments` table is a hand-registered, opt-in-per-
  fixture list, and creating a directory under `fixtures/` enrolls nothing by itself; and (2) that the
  fixture would sit unverified until the second Epic 5 gate — it would not; `fixture_test.go` carries
  no build tag, so a golden added in-story is verified immediately on this target by the ordinary
  `go test ./...`, deferred only on the other three (D-000.4's normal arrangement for every golden in
  this program). Guardrail 10's "do not disturb fixtures/goldens/fonts" was itself a defect in the
  story text — its intent, inherited from the Story 6.2 template, is "belonging to OTHER stories,"
  and this story's own AD-21 fixture is required by Task 5, not excluded by it. The coordinator
  amended the story text accordingly; this was not the developer's error, but the fixture work was
  still owed and has now been added — see `fixtures/component-asset-import/`, below.
- **AssetKey full-vs-abbreviated — confirmed, not revised.** The coordinator independently checked
  this decision against AC2's transport clause and upheld it: `CanvasImagePaint.AssetKey` keeps the
  full 64-hex key; the spec's "abbreviated" was a display concern mistakenly written into the
  transport clause, and an abbreviated key cannot serve a per-key lookup without pushing toward the
  browser-held key table the invariant exists to prevent. No code change; AC2 and the decision log
  were amended to match what was built.

#### Test suites run (measured)

- `go build ./...`, `go vet ./...`, `GOOS=js GOARCH=wasm go build ./...`, `gofmt -l .` — all clean.
- `go test ./...` (folio-go) — all packages pass except `internal/text`'s
  `TestCorpusMeetsP6ExerciseFloors/P6g (opaque names)` (got 7, need ≥20) — DW-11, the sanctioned
  pre-existing red recorded in `epic-6-boundary-gate.md`, untouched by this story. This now also
  includes the two new AD-21 fixture tests
  (`TestRenderMatchesComponentAssetImportGoldenFixture`,
  `TestComponentAssetImportCommandReproducesTheFixtureInput`) and the completeness-guard updates their
  addition required (`TestMatrixDocumentSlugsAreRegisteredInCI`, `TestEpic2GateObligationsMatchTheDeclaredSet`) — all pass.
- `go build -tags matrix ./...`, `go vet -tags matrix ./...` — both clean (the new
  `component-asset-import` matrix registration compiles). **The matrix itself was not run** —
  D-5.13.5 stays declined for this story.
- `folio-designer`: `npx vitest run` — 29 files, 186 tests, all passing (baseline was 27 files/166
  tests; net +2 files/+20 tests from this story). `npm run typecheck`, `npm run lint` (only the
  pre-existing two `only-export-components` warnings), `npm run build` (incl. offline release
  generation/verification), `npm run test:e2e:compile` — all clean.
- **Playwright e2e: real measured result — 21 passed, 0 failed, 2.1 minutes.** Corrected after initial
  delivery: the first pass's claim that the suite "could not execute in this environment" was wrong —
  the environment has three working headless-shell installs
  (`chromium_headless_shell-1217`/`-1223`/`-1228`), and `playwright.config.ts` reads
  `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` for exactly this case; only the broken default
  `chromium-1208` and a blocked network install were tried the first time. The working invocation,
  from `folio-designer/`:
  ```
  PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH="$HOME/Library/Caches/ms-playwright/chromium_headless_shell-1217/chrome-headless-shell-mac-arm64/chrome-headless-shell" npm run test:e2e
  ```
  The first real run (20 passed, 1 failed) found one genuine test defect in the new spec, at line 63:
  `page.getByText('IMAGE')` is a strict-mode collision (Playwright's `getByText(string)` substring/
  case-insensitive-matches by default), resolving to 5 elements — the palette's "Place Image" caption,
  the identity row's "image" type text, the section label itself, the honest-note text, and the
  "Choose image…" button. **Fixed by scoping the locator, not by weakening the assertion**:
  `page.locator('p.section-label', { hasText: /^IMAGE$/ })`, plus `{ exact: true }` added to four other
  asset-identity assertions during a full-spec locator audit. Getting past that line then surfaced a
  second issue, at the undo/redo section: `App.tsx`'s pre-existing, story-independent `applyHistory`
  clears the local selection on every undo/redo (`setCurrentSnapshot(..., true)` calls
  `setSelected([])`), which the original assertions did not account for. Confirmed via `git diff` that
  this behaviour predates this story and via a temporary debug spec (since deleted) that showed the
  inspector panel genuinely deselected after Undo/Redo — a test defect (wrong assumption), not a
  product defect. **Fixed test-only**: capture the revision string, click Undo/Redo, poll for the
  revision to change, then re-select the component before asserting its identity — still fully proving
  AC4's undo/redo semantics via the correct UI interaction shape. No product code was changed for
  either fix; the FSA-tier mocking (real `window.showOpenFilePicker`/`showSaveFilePicker` assignment,
  never disabled) was preserved throughout, and no assertion was deleted or relaxed. The four-target
  hash matrix stays written-but-unrun (D-5.13.5, declined).

#### File List

- `folio-go/component_commands.go` (modified) — `setComponentAsset` command.
- `folio-go/internal/template/image.go` (modified) — doc comment update recording the command as `DecodeImageForRender`'s second call site.
- `folio-go/page_setup.go` (modified) — `CanvasImagePaint`, `addCanvasImagePaint`.
- `folio-go/asset_bytes.go` (new) — `AssetBytes` read-only query.
- `folio-go/component_asset_command_test.go` (new).
- `folio-go/canvas_image_paint_test.go` (new).
- `folio-go/wasm/engine.go` (modified) — `Engine.AssetBytes`.
- `folio-go/wasm/asset_test.go` (new).
- `folio-go/wasm/cmd/engine/main.go` (modified) — `"asset"` dispatch case.
- `folio-designer/src/engine-protocol.ts` (modified) — `'asset'` operation, `CanvasImagePaint`/`image` admission.
- `folio-designer/src/engine-protocol.test.ts` (modified).
- `folio-designer/src/engine-client.ts` (modified) — `'asset'` operation-payload matching.
- `folio-designer/src/engine-client.test.ts` (modified).
- `folio-designer/src/component-asset-command.ts` (new) — `setComponentAssetCommand`, `assetBytesRequest`.
- `folio-designer/src/component-asset-command.test.ts` (new).
- `folio-designer/src/image-file.ts` (new) — `FileSystemImageAccess`/`InputImageAccess`.
- `folio-designer/src/image-file.test.ts` (new).
- `folio-designer/src/file/capability.ts` (modified) — `selectImageFileAccess`.
- `folio-designer/src/main.tsx` (modified) — wires `imageFileAccess` into `App`.
- `folio-designer/src/App.tsx` (modified) — `ImageSection`, `ImagePaint`, `applyImageAsset`, footer note.
- `folio-designer/src/App.test.tsx` (modified) — "Story 5.13: image asset selection" suite.
- `folio-designer/e2e/image-asset.spec.ts` (new) — File System Access-tier e2e scenario; locator scoping
  and undo/redo re-selection fixed after the real Playwright run (see Test suites run, above).
- `folio-go/render_test.go` (modified) — `componentAssetImportKey`/`componentAssetImportTemplateJSON`
  constants, `subprocessComponentAssetImportEnvVar`, new `TestMain` dispatch branch.
- `folio-go/matrix_test.go` (modified, `//go:build matrix`) — `captureComponentAssetImportRender`,
  `component-asset-import` entry in `matrixDocuments`.
- `folio-go/fixture_test.go` (modified) — `TestRenderMatchesComponentAssetImportGoldenFixture`,
  `TestComponentAssetImportCommandReproducesTheFixtureInput`.
- `folio-go/byte_neutrality_test.go` (modified) — `component-asset-import` entries in
  `declaredEpic2GateObligations` and `goldenDigestRecord` (completeness guards; mechanical, dictated
  by the existing pattern, not a new decision).
- `fixtures/component-asset-import/input.folio` (new) — captured canonical output of one real
  `setComponentAsset` command.
- `fixtures/component-asset-import/expected.pdf` (new).
- `fixtures/component-asset-import/expected.json` (new) — `sha256`, `folioGoVersion: "0.0.0-dev"`,
  `goToolchain: "go1.26.0"` (matches every other fixture).
- `fixtures/component-asset-import/README.md` (new).
- `.github/workflows/matrix.yml` (modified) — `component-asset-import` added to the `docs=` list and
  each target's uploaded hash-artifact paths (mechanical, mirrors every existing registration).

#### File List — coordination artifacts (Finding 19, review of 2026-08-29)

The tree carries five modifications the File List above did not originally name. Four are
legitimate coordination artifacts this story's own obligations require; the fifth is confirmed
excluded.

- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified) — this story's key moves to
  `done`; `epic-5` is left `in-progress` per D-5.13.6 (the second Epic 5 boundary gate is still owed).
- `_bmad-output/planning-artifacts/epics.md` — coverage-map bookkeeping for FR4/FR5/FR33 against this
  story landing.
- `_bmad-output/implementation-artifacts/epic-6-boundary-gate.md` — amended per D-5.13.6 to state its
  PASS covers baseline `44121b8` only, naming the post-gate delta (`eef7fbb`, `3a52ae4`) precisely,
  since two commits touched canonical-byte-producing/diagnostic code after that gate closed with no
  story attached.
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — holds D-5.13.0 through D-5.13.8
  (including the D-5.13.2 amendment), all referenced throughout this story and its finisher pass.
- `_bmad-output/implementation-artifacts/evidence/story-6.7-roundtrip-manifest.json` (modified) —
  **confirmed NOT part of this story** by the orchestrator; excluded from this story's commit.
- `_bmad-output/implementation-artifacts/deferred-work.md` (modified, finisher pass) — new DW-22 entry
  for Finding 18 (per-element, uncached `'asset'` fetch), owned by the second Epic 5 boundary gate.

### Finisher delivery — 2026-08-29

Triaged all 20 review findings (Finding Resolutions, below the QA Results section): 19 FIX, 1 DEFER
(DW-22), 0 DISMISS. Applied production and test changes for every FIX; each of the four
non-negotiable findings plus Finding 5 was confirmed by a live mutation that reddens under the exact
defect its proof exists to catch, then confirmed green again after the mutation was reverted.

**Gates re-measured after the finisher pass:**
- `go build ./...`, `go vet ./...`, `GOOS=js GOARCH=wasm go build ./...`, `gofmt -l .`,
  `go build -tags matrix ./...`, `go vet -tags matrix ./...` — all clean.
- `go test ./...` — clean except `internal/text`'s `TestCorpusMeetsP6ExerciseFloors/P6g` (DW-11,
  sanctioned pre-existing red, unchanged by this story or this pass). The matrix itself was **not
  run** — D-5.13.5 stays declined for this story; named here again as written-but-unrun.
- `folio-designer`: `npm run typecheck`, `npm run lint` (only the same two pre-existing
  `only-export-components` warnings), `npm run build` (incl. offline release generation/
  verification), `npm run test:e2e:compile` — all clean. `npx vitest run` — **29 files / 190 tests,
  all passing** (net +2 tests from this pass: the receiver-pinning test for Finding 8 and the
  document-replacement race test for Finding 4; the Finding 3/10/11 additions extended existing
  tests rather than adding new ones).
- **Playwright e2e, real measured result: 23 passed, 0 failed, 2.1 minutes**, using the working
  headless-shell invocation recorded above. `image-asset.spec.ts` alone: 3 passed (the original
  File-System-Access-tier scenario, plus two new ones — drag, and resize+zoom). The four-target hash
  matrix remains written-but-unrun (D-5.13.5, declined).

No finding was scope-expanded beyond what its own fix required; no unrelated code was touched.
`sprint-status.yaml`'s story key moves to `done`; `epic-5` is left `in-progress` (D-5.13.6 — the
second Epic 5 boundary gate is still owed).

## QA Results

### Review Summary

- **Reviewed by:** bmad-code-reviewer (adversarial pass, against the AMENDED story text: D-5.13.1 – D-5.13.8 including the D-5.13.2 amendment)
- **Date:** 2026-08-29
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 5
- **Majors:** 6
- **Minors:** 8
- **Nits:** 1

**One-line verdict.** The Go engine work is genuinely good — the command is transactional, the paint really is a translation of the shipping run path, undo really does restore collected bytes through byte-snapshot replay, and no decoder entered the module. What is not good is the **evidence**. Three of this story's four load-bearing new proofs are green under the exact defect each exists to catch, measured by mutation: a document-wide orphan sweep passes the entire Go suite; `object-fit: contain` over the whole box passes all 186 designer unit tests; and `eef7fbb`'s receiver-binding defect passes `image-asset.spec.ts`. The e2e override granted at D-5.13.5 was bought with a promise this spec does not keep.

**Gates measured in this review (not re-run from the Delivery Log):**
- `folio-designer` unit suite: **29 files / 186 tests, all passing** — matches the Delivery Log exactly.
- `go test ./...` under each mutation below: clean except `internal/text`'s `TestCorpusMeetsP6ExerciseFloors/P6g` (DW-11, sanctioned pre-existing red, present at baseline).
- `TestMatrixDocumentSlugsAreRegisteredInCI` red-proved live (removing `component-asset-import` from `matrix.yml`'s `docs=` reddens it) — the Delivery Log's "both guards failed loudly until updated" claim is **verified true**.
- `fixtures/*/expected.json`: the new fixture's `folioGoVersion: 0.0.0-dev` / `goToolchain: go1.26.0` **match all 15 existing fixtures**, so `assertFixturesShareToolchain` cannot red the gate for the wrong reason. Guard edits confirmed purely additive; the new `goldenDigestRecord` entry correctly omits a `readme` site because the new README does not quote the digest (the completeness half would have caught it if it did).
- Working tree verified pristine after every mutation: `git diff --stat HEAD` is byte-for-byte the pre-review figure (23 files, 1503 insertions, 196 deletions).

---

### Finding 1: A document-wide orphan sweep passes the entire Go test suite

- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/component_asset_command_test.go:279-303` (`TestSetComponentAssetNeverSweepsAnUnrelatedOrphan`), fixture helper at `folio-go/component_asset_command_test.go:34-55` (`twoImageTemplateJSON`)
- **Observation**: The test named for D-5.13.3 contains **no orphan**. `twoImageTemplateJSON` gives `keyB` to element `e2`, so `keyB` is a *referenced* asset. The test therefore only proves "the command does not delete an asset another element still references" — which is exactly what a document-wide sweep would also do. Red-proved: replacing the scoped collection at `component_commands.go:727-729` with

  ```go
  for existing := range t.doc.Assets {
      if !assetKeyReferenced(t, existing) { delete(t.doc.Assets, existing) }
  }
  ```

  and running `go test -count=1 ./...` produces **only DW-11**. Nothing in the repository reddens. `TestRenderUnrecognisedMediaTypeErrorsOnlyWhenDrawn` does not catch it either, because it never applies a `setComponentAsset` command — it only loads and renders.
- **Impact**: AC4's central rule, and the one clause a lead ruling already had to correct once (D-5.13.3: "This one corrected a real defect in the drafted story"), has **zero** regression coverage. AC4's proof clause explicitly demands "Red-prove a document-wide sweep". The shipped code is correct today; nothing stops the next edit from silently reintroducing the exact failure D-5.13.3 describes — open a `.folio` carrying an unreferenced asset, set any image, and the author's asset vanishes from their own file.
- **Suggested Resolution**: Add a **third** asset key to the fixture document that **no element references**, and assert it survives the command. That single change makes the existing test discriminate a scoped collection from a sweep. Consider also asserting the unreferenced asset is of an unrecognised media type, so the test doubles as an in-command echo of RP-11's positive control.
- **Related AC**: AC4 (and D-5.13.3's "How we'd know it was wrong")

### Finding 2: `image-asset.spec.ts` is green with `eef7fbb`'s receiver-binding defect reintroduced

- **Severity**: Blocker
- **Category**: Tests
- **Location**: `folio-designer/e2e/image-asset.spec.ts:17-47` (`installNativePickers`)
- **Observation**: The spec's `page.addInitScript` replaces `window.showOpenFilePicker` with a plain **arrow function** via `Object.assign(window, {...})`. An arrow function has no `this` binding and no brand check, so it accepts any receiver. This removes precisely the sensitivity the spec exists to provide. Red-proved end-to-end: reverting `file/capability.ts:15` to `eef7fbb`'s defect —

  ```go
  showOpenFilePicker: pickerWindow.showOpenFilePicker   // no .bind(pickerWindow)
  ```

  — and running `npx playwright test e2e/image-asset.spec.ts` with the working headless-shell path gives **1 passed (1.6m)**. The same mutation reddens **2 unit tests** (`file/file-access.test.ts:54` and `image-file.test.ts:56`), so the unit suite catches it and the browser suite does not — the inverse of the story's stated rationale.
- **Impact**: This is the **entire justification** for D-5.13.5's e2e override. That ruling's reasoning is: `eef7fbb` "passed the entire unit suite and survived the whole Epic 5 Playwright gate ... because `vi.fn()` doubles have no brand check", and 5.13 adds a third picker on the same seam. The spec as written reproduces the very property that made `eef7fbb` invisible — it is a `vi.fn()` double wearing a Playwright costume. A green browser test that stubs the picker this thoroughly is worse than no test, because it reads as proof that the tier is exercised against real browser semantics when it is not.
- **Suggested Resolution**: Make the mock brand-check its own receiver. Replace the arrow function with a `function` that throws the browser's own failure when misused, e.g.:

  ```ts
  const nativeOpen = window.showOpenFilePicker
  Object.assign(window, {
    showOpenFilePicker: function (this: unknown, options?: ...) {
      if (this !== window) throw new TypeError("Failed to execute 'showOpenFilePicker' on 'Window': Illegal invocation")
      return Promise.resolve([...])
    },
  })
  ```

  Then re-run the mutation above and confirm the spec **fails**. Until that red proof exists, this spec does not discharge D-5.13.5's condition and the override should be treated as unsatisfied.
- **Related AC**: AC1, AC2 (and D-5.13.5's Consequences clause)

### Finding 3: The canvas paints with no geometry assertion anywhere — a browser-computed fit passes every suite

- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-designer/src/App.test.tsx:868-895` ("paints the image inside the Go-owned draw rectangle…"), `folio-designer/e2e/image-asset.spec.ts:78` and `:123`, subject at `folio-designer/src/App.tsx:985`
- **Observation**: The unit test named "paints the image inside the Go-owned draw rectangle" asserts only that `createObjectURL` was called once, that `img.src` contains the blob URL, that the placeholder renders for the second element, and that unmount revokes. It **never reads `style.left/top/width/height`**. The e2e's only assertions on the painted image are `toBeVisible()`. Red-proved: replacing `App.tsx:985` with the AC's own named anti-pattern —

  ```ts
  const style: CSSProperties = { position: 'absolute', left: 0, top: 0,
    width: canvasDisplay.css(component.width, zoom),
    height: canvasDisplay.css(component.height, zoom), objectFit: 'contain' }
  ```

  — leaves **29 files / 186 tests all passing**, and the e2e's image assertions are `toBeVisible()`, which an `object-fit` image also satisfies.
- **Impact**: AC3's headline prohibition ("The browser must not compute the fit itself, must not reach for `object-fit` as the authority") and its explicitly named red proof ("Red-prove a browser-computed fit") are both unmet. The Go side of AC3 is excellent — `TestCanvasImagePaintExactlyMatchesTheShippingRunPath` genuinely proves the band-frame translation across all three band origins (0, 20000, 739890), verified by perturbing `DrawY` by +1 and watching all three bands redden. But that proof stops at the projection boundary. Everything downstream of the wire is unpinned, which is the half AD-17 and AD-24's designer clause are actually about.
- **Suggested Resolution**: Assert the four style values against the projected `image.draw*` at the default zoom **and at a second zoom** (see Finding 10 — zoom is untested for this paint), and add one Playwright assertion comparing `img.boundingBox()` against the component's box scaled from the snapshot. Then re-run the `object-fit` mutation and confirm red.
- **Related AC**: AC3

### Finding 4: `applyImageAsset` installs its result with no document-generation guard

- **Severity**: Blocker
- **Category**: Correctness / AC Conformance
- **Location**: `folio-designer/src/App.tsx:479-493`
- **Observation**: Every other async committed-command path in this file captures the generation before the await and guards installation with it — `bindComponentScalar` (`App.tsx:404-424`), the table editor (`:379-399`), parameter references (`:294-310`), and the preview pipeline (`:195-206`) all use `const generation = documentGeneration.current` … `if (documentGeneration.current === generation && snapshotRef.current?.revision === priorRevision)`. `applyImageAsset` captures `priorRevision` but uses it **only** to decide whether to call `invalidatePreview()`; it then calls `setCurrentSnapshot(result.snapshot)` unconditionally, with no generation or revision check. It also holds `id` in its closure across the two longest awaits in the application (an OS file dialog, then an engine command carrying up to megabytes).
- **Impact**: AC1 names this exactly: "Red-prove … a command result installed after document replacement." The mechanism is neither implemented nor proved. Element ids are reused across documents (`e1`, `e2`, …), so if an Open / Start blank / undo lands between the picker resolving and the command returning, the command is applied to a *different* document's element of the same id, and the stale snapshot is installed over the authoritative one. Separately, `assetError` and `assetBusy` are not in `setCurrentSnapshot`'s `clearDocumentInteraction` list (`App.tsx:495`), so a stale asset diagnostic survives a document replacement and is re-displayed against whatever new element shares that id.
- **Suggested Resolution**: Capture `documentGeneration.current` before the picker await; guard both `setCurrentSnapshot` and `setAssetError` on it plus `snapshotRef.current?.revision === priorRevision`, matching `bindComponentScalar`'s shape exactly. Add `setAssetError(undefined); setAssetBusy(false)` to `clearDocumentInteraction`. Cover with a unit test that resolves the picker after a simulated document replacement and asserts the snapshot is **not** installed.
- **Related AC**: AC1

### Finding 5: A placed image component cannot be dragged — the new `<img>` takes native drag

- **Severity**: Blocker
- **Category**: Correctness
- **Location**: `folio-designer/src/App.tsx:986` (`<img … className="canvas-image-paint">`), `folio-designer/src/App.css` (no `canvas-image` rule exists; contrast `.canvas-text-paint` at `App.css:53`)
- **Observation**: Reported by the repository owner mid-review and independently confirmed here from the source: the `<img>` carries no `draggable={false}`, and neither `.canvas-image-paint` nor `.canvas-image-placeholder` has any CSS rule at all — `grep -c "canvas-image" App.css` → **0**. `.canvas-text-paint` (App.css:53) carries `pointer-events: none`; the image paint carries nothing. Native `dragstart` takes pointer capture and `CanvasComponent`'s `onPointerMove`/`onPointerUp` never complete the move. `.canvas-component`'s `user-select: none` (App.css:50) does not suppress native image dragging.
- **Impact**: A placed image is unmovable in the shipped designer. Logged here for the record because this is Story 5.13's code; the fix is understood to be in flight and is **not** re-derived here.
- **Suggested Resolution**: As already planned: `draggable={false}` on the `<img>` plus a `.canvas-image-paint, .canvas-image-placeholder { pointer-events: none; }` rule, with the Playwright regression test that drags an image component and asserts committed X/Y changed. See Finding 10 for the coverage-shape gap that let this reach the owner.
- **Related AC**: AC3

### Finding 6: The size bound the browser enforces and the one Go reports disagree by ~2 MiB

- **Severity**: Major
- **Category**: Correctness / AC Conformance
- **Location**: `folio-go/component_commands.go:617-626` (`maxComponentAssetBytes` and its doc comment), `folio-designer/src/engine-protocol.ts:6` and `:204`, `folio-designer/src/component-asset-command.ts:8-16`
- **Observation**: `maxComponentAssetBytes` is applied to the **decoded** bytes (8 388 608). The command travels as JSON containing **base64**, and `parseRequest` (`engine-protocol.ts:204`) rejects any `command` payload whose `byteLength > MAX_ENGINE_PAYLOAD_BYTES` — the same 8 388 608, applied to the inflated envelope. Base64 is 4/3, so the envelope's effective ceiling on the decoded file is `3/4 × 8 388 608 − overhead ≈ 6 291 376` bytes (~6.0 MiB). Any image between ~6 MiB and 8 MiB is refused by the **transport**, never reaching Go. `EngineRequestAdmission.admit` (`engine-worker-admission.ts:20-21`) returns `PROTOCOL_INVALID` / "The engine request is invalid", which `applyImageAsset` surfaces verbatim. A file at exactly Go's stated limit can never reach Go at all.
- **Impact**: D-5.13.4's Consequences clause is explicit: "the protocol envelope may reject the message, but the author-facing 'file too large' diagnostic AC2 requires comes from the engine, and **the two must not disagree about the threshold**." They disagree, and the author sees a protocol error instead of AC2's stated reason. The doc comment at `component_commands.go:619-621` asserts the opposite of what the code does — "applied here to the DECODED file, not the base64-inflated command envelope, so the author-facing diagnostic and the protocol threshold never disagree" — base64 inflation is precisely *why* they disagree. `TestSetComponentAssetRefusesOversizedPayload` calls `ApplyComponentCommand` directly and so never crosses the envelope.
- **Suggested Resolution**: Pick one number and derive the other from it. Either set `maxComponentAssetBytes` to the decoded size the envelope can actually carry (`3/4 × MAX_ENGINE_PAYLOAD_BYTES` minus the fixed JSON overhead) with the derivation stated, or raise the envelope for the `command` operation. Add a designer test that sends an over-envelope asset command and asserts the author-facing text is the engine's size diagnostic, not `PROTOCOL_INVALID`. Correct the doc comment either way.
- **Related AC**: AC1, AC2 (D-5.13.4)

### Finding 7: The golden fixture cannot red-prove sorted-key ordering — it carries one asset

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/fixture_test.go:944-1015` (`TestComponentAssetImportCommandReproducesTheFixtureInput`, doc comment lines 947-963), `fixtures/component-asset-import/README.md` ("What this fixture proves that image-embed does not"), `folio-go/byte_neutrality_test.go:301-317`
- **Observation**: Three separate places claim assertion (b) "red-proofs … the sorted-key ordering". It cannot: the command orphan-collects the base document's only other asset, so `input.folio` carries **exactly one** asset (README step 3 says so outright), and a one-key map has no ordering to sort. Red-proved: reversing the asset iteration order in `internal/template/serialize.go:426` —

  ```go
  for _, k := range func() []string { s := slices.Sorted(maps.Keys(assets)); slices.Reverse(s); return s }() {
  ```

  — leaves both fixture tests green **and the entire `go test ./...` suite green** (only DW-11). Nothing in the repository pins asset-key ordering. The fixture's other two claims are real and were red-proved here: changing `splitBase64Canonical`'s `const width = 76` to `64` reddens assertion (b); removing the repoint (`component_commands.go:726`) reddens it; disabling orphan collection reddens it.
- **Impact**: D-5.13.7's own falsification criterion — "Assertion (b) never failing across a real change to the wrapping or key derivation would mean it is not actually pinning the command's output" — is met for one of the three named properties. The fixture is not a second `image-embed`, but its advertised coverage is one-third wider than its actual coverage, and three documents assert the wider claim.
- **Suggested Resolution**: Either give the fixture a second asset (e.g. a second image element on a different picture, which also exercises the dedup/no-collect path in the same golden), or delete the sorted-key claim from the test doc comment, the README, and the `goldenDigestRecord` comment. The first is better: it is the only way any committed byte in this repository would notice an ordering change.
- **Related AC**: AC1 (D-5.13.7)

### Finding 8: `selectImageFileAccess`'s own receiver binding is unpinned

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-designer/src/file/capability.ts:42`, `folio-designer/src/image-file.test.ts:56-69`
- **Observation**: The mandated receiver test asserts `receivers).toEqual([window])` after calling `selectImageFileAccess()` with its `currentBrowser()` default. But `currentBrowser()` (`capability.ts:15`) already binds the picker to `window`, and re-binding an already-bound function is a no-op — so the test cannot see line 42's `.bind(browser)` at all. Red-proved: deleting `.bind(browser)` from line 42 leaves `image-file.test.ts` at **7 passed** and the full designer suite at 186 passed.
- **Impact**: The test reddens only against `currentBrowser()`'s shared bind, which the pre-existing `file/file-access.test.ts:54` already covers for the template and sample tiers. D-5.13.5's consequence — "A unit test must pin the new picker's **receiver**" — is discharged in name; the new tier's own binding decision is not pinned by anything. Combined with Finding 2, no test in the repository distinguishes the image tier's receiver handling from the other two tiers'.
- **Suggested Resolution**: Add a case that passes an **explicit** `FileAccessBrowser` whose `showOpenFilePicker` is an unbound receiver-recording `function`, and assert the receiver is that object — which is what line 42's `.bind(browser)` actually guarantees. That case reddens when line 42 loses its bind.
- **Related AC**: AC2 (D-5.13.5)

### Finding 9: A missing asset and an undecodable media type are reported with the same, wrong reason

- **Severity**: Major
- **Category**: AC Conformance
- **Location**: `folio-designer/src/App.tsx:869-873` (`ImageSection`'s absence branch), `folio-designer/src/App.tsx:984` (`ImagePaint`'s absence branch), producer at `folio-go/page_setup.go:270-283`
- **Observation**: `addCanvasImagePaint` `continue`s — leaving `Image` absent — for **three** distinct conditions: the asset key is not in the document's assets map, `DecodeAssetBytes` fails, and `DecodeImageForRender` fails. The browser then renders one fixed string for all three: the inspector says "This version cannot render this asset's media type." and the canvas says "Image unavailable — this version cannot render its media type". `folio-go/canvas_image_paint_test.go:199-233` (`TestCanvasImagePaintIsAbsentForAMissingAssetKey`) proves the missing-key state is reachable from a document that parses successfully.
- **Impact**: AC3 requires "an honest placeholder **naming the reason**" and the story's Project structure notes point at `render.go:1698` — "the located missing-asset error whose wording the canvas placeholder should echo rather than reinvent". For a document with a dangling asset reference, both surfaces name a reason that is false: the media type is fine, the asset is gone. AC2's "state the concrete reason in the product's terse register" is likewise not met for that state. D-5.13.2's "one Go-side signal drives both" was a ruling about *the media-type case*; it does not license collapsing a fourth condition into it.
- **Suggested Resolution**: Carry a small bounded Go-side reason discriminant alongside the absent paint (still one signal, now carrying which of the two it is), and echo `render.go:1698`'s wording for the missing case. Alternatively keep one signal and reword both strings so they are true of every condition they cover. Add the missing-key case to `App.test.tsx`'s coverage.
- **Related AC**: AC2, AC3

### Finding 10: Coverage was written against the old rendering and not re-derived for the new DOM element

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-designer/e2e/image-asset.spec.ts` (whole file), `folio-designer/e2e/component-manipulation.spec.ts:18-49`, `folio-designer/src/App.test.tsx:868-895`
- **Observation**: This is the audit the coordinator asked for, and Finding 5 is one instance of a family, not a one-off. Story 5.13 replaced a bare text node (`'Image'`) with a real `<img>` that participates in browser-native behaviour, but the test plan carried over the old element's assumptions. Measured gaps of the same shape:
  1. **Move** — `component-manipulation.spec.ts` drags a **Rectangle** only; no spec drags an image. This is the shipped defect.
  2. **Resize** — the same spec resizes a Rectangle only. `.resize-handle` (`App.css:60`) sits at `right:-9px; bottom:-9px`, so 15px of its 24×24 hit target overlaps the box interior, which the `<img>` now occupies. DOM order puts the handle after `<ImagePaint>` so it should still win the hit test, but nothing asserts it, and the `<img>` has no `pointer-events: none`.
  3. **Zoom** — `AC3` says the rectangle is "mapped through the existing zoom rule". `canvasDisplay.css(…, zoom)` is applied at `App.tsx:985`, but no unit test and no e2e exercises the image paint at any zoom other than 1. The old `'Image'` label had no geometry, so no zoom coverage was ever owed for this element; now it is.
  4. **Live resize** — `componentStyle(active, zoom)` uses the live drag rect while `ImagePaint` uses the committed `component.x/y` and `image.draw*`, so the picture does not track a resize until the command commits. This matches `TextPaint`'s precedent, so it is noted rather than charged, but it is untested either way.
- **Impact**: `image-asset.spec.ts` proves a narrower path than AC2 and AC3 assert. Against AC2's stated proof list it covers a single-image selection only — not multi-image, mixed, non-image or empty (all unit-only), not focus visibility (nowhere), and not the keyboard path (see Finding 15). Against AC3's list it covers none of the three fit shapes, no zoom, and only one of the three revocation triggers (Finding 11). A spec that sets an asset and never interacts with the resulting element is precisely how Finding 5 reached the owner.
- **Suggested Resolution**: Extend `image-asset.spec.ts` (or `component-manipulation.spec.ts`) to drag **and** resize an image component with a committed-geometry assertion; add one zoom step with a `boundingBox()` assertion against the projected rectangle; and treat "this story introduces a new element into the canvas" as a standing trigger to re-derive the interaction matrix rather than inherit it.
- **Related AC**: AC2, AC3

### Finding 11: URL revocation is proved for one of the three triggers the AC names

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-designer/src/App.test.tsx:888-891`, subject at `folio-designer/src/App.tsx:969-983`
- **Observation**: AC3's proof clause reads "Prove URL revocation on delete, replace and document replacement." The only assertion is `view.unmount()` → `expect(revokeObjectURL).toHaveBeenCalledWith('blob:paint-1')`. Asset replacement (`image.assetKey` changes on a live component) and document replacement (`generation` changes) are handled by construction through the effect's dependency array, but neither is exercised. The e2e never asserts revocation at all.
- **Impact**: "no accumulation across a session" is the AC's actual promise and it rests on the dep array being right. A future edit that memoizes the URL, hoists it out of the effect, or drops `generation` from the deps would leak on every document replacement with nothing red.
- **Suggested Resolution**: Two more cases in the existing test: rerender with a changed `assetKey` and assert the first URL was revoked and a second created; rerender with a bumped `generation` and assert the same. Both are cheap given the harness already stubs `URL.createObjectURL`.
- **Related AC**: AC3

### Finding 12: The strict admission guard accepts a truncated asset key, and its own fixture uses one

- **Severity**: Minor
- **Category**: Convention / Tests
- **Location**: `folio-designer/src/engine-protocol.ts:163` (`isImagePaint`), `folio-designer/src/engine-protocol.test.ts:88`
- **Observation**: The guard accepts `assetKey.length` in 1..64. The test's own well-formed fixture uses `assetKey: 'abcdef012345'` — a 12-character key. D-5.13.2 (amendment) settled that the wire key is the **full 64-hex** value precisely because a per-key request cannot address an asset by a prefix; Go's `isAssetKeyShape` (`folio-go/asset_bytes.go:41-52`) requires exactly 64. A truncated key therefore passes admission and then fails the per-key fetch, producing the permanent "Loading image…" of Finding 13.
- **Impact**: Task 2 asks strict admission to reject "unknown, malformed, oversized or schema-mirroring projections". The one field whose width is a settled ruling is the one the guard does not check, and the test fixture cements the loose reading.
- **Suggested Resolution**: `value.assetKey.length !== 64` → reject. Change the fixture to `'a'.repeat(64)` (as `App.test.tsx:786` already does) and add a rejection case for a 12-character key.
- **Related AC**: AC1 (D-5.13.2 amendment)

### Finding 13: A failed per-key asset fetch renders as "Loading image…" forever

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-designer/src/App.tsx:976-986`
- **Observation**: `.catch(() => undefined)` swallows every failure of the `'asset'` request, leaving `url` undefined, which renders the `Loading image…` placeholder permanently. Causes include a truncated key (Finding 12), a key absent from the engine's current document, and any transport failure.
- **Impact**: AC3 requires an honest placeholder for a component that cannot be painted. A permanent "Loading" is the opposite of honest — it claims work is in progress that has already failed.
- **Suggested Resolution**: Track a failed state in the effect and render a named failure placeholder distinct from the loading one.
- **Related AC**: AC3

### Finding 14: `assetError` and `assetBusy` survive a document replacement

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-designer/src/App.tsx:495` (`setCurrentSnapshot`'s `clearDocumentInteraction` branch), `:125-126`
- **Observation**: The branch clears `selected`, `bindingError`, `bindingBusy`, `tableEditor`, `tableEditorError` and the interaction state, but not `assetError` or `assetBusy`. A stale asset diagnostic keyed on an element id therefore reappears against a new document's element of the same id, and a picker promise that never settles leaves the Choose control disabled across the replacement.
- **Impact**: AC1's "displayed asset state unchanged" applies to a rejected command; this is the neighbouring case the same clause implies. Small blast radius, but it is a divergence from the pattern every sibling error follows.
- **Suggested Resolution**: Add `setAssetError(undefined); setAssetBusy(false)` alongside the existing clears. Fold into Finding 4's fix.
- **Related AC**: AC1

### Finding 15: The "keyboard path" test performs a mouse click

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-designer/src/App.test.tsx:829-845`
- **Observation**: The test titled "drives the keyboard path from selection to picker to one committed setComponentAsset command" selects with `fireEvent.click`, calls `button.focus()`, asserts `document.activeElement`, and then activates with `fireEvent.click(button)`. No key event is dispatched anywhere. The e2e also clicks throughout.
- **Impact**: AC2 asks for "the keyboard path from selection to picker to committed set". What is proved is focusability. A native `<button>` is keyboard-operable by construction so the practical risk is low, but the assertion does not support the claim its name makes — and AC2's separate "visible `colors.select` focus" requirement is asserted nowhere at all (it holds today via the global `button:focus-visible` rule at `App.css:5`).
- **Suggested Resolution**: Rename, or use `fireEvent.keyDown`/`userEvent.keyboard('{Enter}')` for both the selection and the activation, and reach the button by `Tab` from the selected component so "keyboard-reachable" is also exercised.
- **Related AC**: AC2

### Finding 16: `TestCanvasImagePaintTracksAResize` asserts only that something changed

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/canvas_image_paint_test.go:140-168`
- **Observation**: The only assertion is `*afterImage == *beforeImage` → fail. Any change to any of the eight fields satisfies it, including a wrong one.
- **Impact**: AC3 asks the painted rectangle to "track a resize". The correctness half is covered indirectly by `TestCanvasImagePaintExactlyMatchesTheShippingRunPath` for a fixed document, but this test as written would pass on a paint that resized incorrectly.
- **Suggested Resolution**: Assert the post-resize rectangle equals `resolveImagePlacement` on the resized run, in the same shape the shipping-run-path test uses.
- **Related AC**: AC3

### Finding 17: `assetKeyReferenced` walks only the three top-level band element lists

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/component_commands.go:733-743`
- **Observation**: The reference scan iterates `Bands.PageHeader.Elements`, `Bands.Content.Elements`, `Bands.PageFooter.Elements` only. This is correct for today's model, and images in table cells are explicitly out of scope for this story (AC4's exclusions).
- **Impact**: This helper is the *safety* half of a **delete**. If a later story introduces images anywhere the walk does not reach, it will under-report references and delete a live asset — a silent data loss with no compile error to announce it. The story-scoped test set would not notice.
- **Suggested Resolution**: Either derive the walk from a single element-enumeration helper shared with `findComponent`/`collectImageRuns`, or add a comment binding the helper to the same enumeration those two use, so a new element location has one place to update.
- **Related AC**: AC4

### Finding 18: One `'asset'` round trip per image component per mount, with no cache

- **Severity**: Minor
- **Category**: Performance
- **Location**: `folio-designer/src/App.tsx:970-983`
- **Observation**: `ImagePaint`'s effect fires per component instance. N image elements sharing one asset produce N requests, each carrying the full decoded bytes back through the FIFO-correlated single worker; the same happens again on every `generation` change.
- **Impact**: D-5.13.2 accepted "an extra request per **distinct** asset and its cache/lifetime handling in the browser". What shipped is a request per *element*, with no cache. On a header logo repeated across bands this is small; on a document with many images it serializes the worker behind repeated multi-megabyte transfers.
- **Suggested Resolution**: Key the fetch by asset key at the canvas level (one in-flight request and one object URL per distinct key, revoked when the last referent unmounts), or record the accepted cost explicitly if it is being deferred.
- **Related AC**: AC3

### Finding 19: The File List omits five changed files

- **Severity**: Minor
- **Category**: Convention
- **Location**: Story File List section, versus `git diff --stat HEAD`
- **Observation**: The tree carries five modifications the File List does not name: `_bmad-output/implementation-artifacts/sprint-status.yaml`, `_bmad-output/planning-artifacts/epics.md`, `_bmad-output/implementation-artifacts/epic-6-boundary-gate.md`, `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`, and `_bmad-output/implementation-artifacts/evidence/story-6.7-roundtrip-manifest.json`. The first four are legitimate coordination artifacts (D-5.13.6 mandates the gate-record amendment; the decision log holds D-5.13.0–D-5.13.8). The fifth is confirmed by the orchestrator as **not part of this story**.
- **Impact**: Scope discipline otherwise holds cleanly — every code file in the diff belongs to 5.13, `internal/template/image.go` is comment-only as required, the palette insert path (`component_commands.go:1191-1206`) is untouched, `decodeAssets` is untouched, no image decoder or re-encoder entered the module (`go.mod` still has exactly one dependency), no URL/path/fetch/render-time IO appears on the image path, and neither `eef7fbb` nor `3a52ae4` has been absorbed. But an unlisted changed file carries an unchecked claim.
- **Suggested Resolution**: Name the four coordination artifacts in the File List with a one-line reason each, and confirm `evidence/story-6.7-roundtrip-manifest.json` is excluded from this story's commit.
- **Related AC**: Guardrail 10

### Finding 20: The fixture's digest-as-key red proof fires through a pre-existing mechanism

- **Severity**: Nit
- **Category**: Tests
- **Location**: `folio-go/fixture_test.go:947-963`
- **Observation**: Mutating `digest := sha256.Sum256(decoded)` does redden `TestComponentAssetImportCommandReproducesTheFixtureInput`, but the failure is `ApplyComponentCommand … "component asset did not pass format validation"` — the command's own reparse hitting `decodeAssets`'s pre-existing digest-match enforcement, not the byte comparison. The claim holds; the mechanism doing the catching predates this story.
- **Impact**: None to correctness. Recorded so a future reader does not over-credit the fixture for coverage `parse.go` already provides.
- **Suggested Resolution**: Optionally note the attribution in the doc comment.
- **Related AC**: AC1

---

### AC-by-AC disposition

- **AC1** — *Partially satisfied.* The command, transaction, dedup, digest, media-type refusal at the command, size enforcement in Go, revision/undo semantics and the bytes-unchanged short-circuit are all implemented and genuinely proved (`component_asset_command_test.go`, `wasm/asset_test.go` — the wasm undo test is a real proof that collected bytes return through byte-snapshot replay, with no inverse-command machinery invented). Unmet: the browser-side threshold agreement (Finding 6) and the "result installed after document replacement" guarantee (Finding 4).
- **AC2** — *Partially satisfied.* The IMAGE section exists between CONTENT/TYPOGRAPHY and BOX, shows engine-sourced identity with a display-abbreviated key, appears only for a single image selection, states the no-picker and undecodable reasons in text, and the footer note's asset clause is gone (verified in `App.tsx:851`). Unmet: the reason stated for a missing asset is false (Finding 9); the keyboard and focus-visibility claims are asserted only nominally (Finding 15); the e2e covers one selection shape of five (Finding 10).
- **AC3** — *Partially satisfied on the Go side, unproven on the browser side.* `addCanvasImagePaint` is correctly a sibling of `addCanvasTextPaint` invoked from `CanvasWithTextPaint`, builds in the band frame, calls the same `resolveImagePlacement`, routes every derived coordinate through `canvasDerived`, and is absent-not-zero — and the two-producer/translation proof has real teeth (verified by perturbation across three band origins). The projection carries no bytes; `AssetBytes` is the separate per-key route; the full 64-hex key on the wire is correct and settled. Unmet: Findings 3, 5, 9, 10, 11, 12, 13, 18.
- **AC4** — *Implemented correctly, materially unproven.* Scoped collection, the "replace ten times, one asset" property, dedup, the untouched palette insert path, the untouched `decodeAssets` and open-set `mediaType` posture, `TestRenderUnrecognisedMediaTypeErrorsOnlyWhenDrawn` still green and untouched, undo-restores-collected-bytes, and FR33 all hold. Unmet: the one rule D-5.13.3 exists to protect has no discriminating test (Finding 1).

### What I would ship

**I would not ship this yet.** Not because the engine is wrong — on the evidence I gathered it is right, and the Go-side proofs of the canvas translation, the transaction and the history semantics are among the better ones in this repository. I would hold it because the story's three headline new proofs are each green under the exact defect they were written to catch, and two of those (Findings 1 and 2) are the direct outputs of lead rulings that already caught a real defect once. Landing them as-is converts two live rulings into decoration.

The shortest path to *ship* is small and mechanical: add a genuinely unreferenced third asset to the orphan fixture (Finding 1), make the e2e picker mock brand-check its receiver (Finding 2), assert the four style values against the projected rectangle (Finding 3), and add the generation guard that every sibling handler already has (Finding 4). Findings 5 and 6 are separate product fixes. Everything else can follow.

---

### Finding Resolutions — finisher pass, 2026-08-29

**Disposition: 19 FIX, 1 DEFER, 0 DISMISS.** Every legitimate finding was in scope for this story
and was fixed; nothing was dismissed. One Minor (Finding 18) is deferred with a named owner because
it is real, accepted-but-unbuilt design cost with no measured trigger yet, not a defect.

All four Blockers named as non-negotiable, plus Finding 5 (also a Blocker), were fixed with a
production or test change and confirmed by a **live mutation** that reddens under the exact defect
each proof exists to catch, then confirmed to pass clean again after the mutation was reverted. See
the Test suites run entry below for the consolidated gate results this pass measured.

| # | Sev | Decision | Rationale |
|---|---|---|---|
| 1 | Blocker | **FIX** | `twoImageTemplateJSONWithOrphan` adds a THIRD, genuinely unreferenced asset (unrecognised media type, echoing RP-11's positive control) to `TestSetComponentAssetNeverSweepsAnUnrelatedOrphan`. Red-proved: substituting the document-wide sweep D-5.13.3 rejected now reddens this test (previously only the sanctioned DW-11 red survived); reverted clean. Files: `folio-go/component_asset_command_test.go`. |
| 2 | Blocker | **FIX** | `image-asset.spec.ts`'s native-picker mock now brand-checks its receiver (`function` + `this !== window` throwing the browser's own "Illegal invocation"), replacing the arrow functions that accepted any receiver. Red-proved live: reverting `capability.ts`'s `.bind(browser)` to the pre-`eef7fbb` defect now fails the spec (previously 1 passed); reverted clean, full spec re-run green. Files: `folio-designer/e2e/image-asset.spec.ts`. |
| 3 | Blocker | **FIX** | Unit test now asserts the painted `<img>`'s `left`/`top`/`width`/`height` against the projected draw rectangle at zoom 1 AND a second zoom (110%); e2e gained a real resize+zoom scenario with `boundingBox()` assertions. Red-proved live: substituting the AC's own named `object-fit`/no-fit anti-pattern into `App.tsx` now reddens the unit test (previously all 186 passed); reverted clean. Files: `folio-designer/src/App.test.tsx`, `folio-designer/e2e/image-asset.spec.ts`. |
| 4 | Blocker | **FIX** | `applyImageAsset` now captures `documentGeneration.current`/`priorRevision` BEFORE the picker await and guards `setCurrentSnapshot`/`setAssetError`/`setAssetBusy` on them, matching `bindPickedPath`'s shape exactly; `assetError`/`assetBusy` added to `clearDocumentInteraction` (also closes Finding 14). New unit test resolves the picker AFTER a simulated Start-blank document replacement and asserts the stale command result is never installed. Red-proved live: reverting the guard reddens the new test (revision 2 overwrote revision 50); reverted clean. Files: `folio-designer/src/App.tsx`, `folio-designer/src/App.test.tsx`. |
| 5 | Blocker | **FIX (confirmed)** | The fix (`draggable={false}` plus `pointer-events: none` on `.canvas-image-paint`/`.canvas-image-placeholder`) was already present in the working tree when this pass started — the repository owner's in-flight fix the finding itself names. Confirmed by reading the source and by a new, real-mouse-event e2e regression test that drags a placed image and asserts both a committed revision change and a moved box. Files: `folio-designer/e2e/image-asset.spec.ts` (new coverage only; `App.tsx`/`App.css` already carried the fix). |
| 6 | Major | **FIX** | `maxComponentAssetBytes` is now DERIVED — `(engineProtocolMaxPayloadBytes − 4 KiB overhead reservation) × 3/4` — instead of reusing the raw 8 MiB envelope ceiling against decoded bytes, so a file at Go's stated limit can always actually reach the command; doc comment corrected to state the real reasoning. New designer test proves a command at Go's derived ceiling clears protocol admission, and reproduces the old bug (the OLD 8 MiB decoded bound's base64 envelope exceeding `MAX_ENGINE_PAYLOAD_BYTES`) as a regression guard. Files: `folio-go/component_commands.go`, `folio-designer/src/component-asset-command.test.ts`. |
| 7 | Major | **FIX** | The fixture's base document now carries a SECOND image element/asset (`e2`, a real JPEG) the command never touches, so two assets under two different keys survive it. Regenerated `input.folio`/`expected.pdf`/`expected.json`/README and the matching Go constants against the REAL command/render output (never hand-authored). Note: the reviewer's own suggested mutation ("reverse the loop order in `writeAssets`") is empirically a no-op — `writeObject` re-sorts unconditionally regardless of insertion order — so the actually-representative red proof disables `writeObject`'s sort entirely; that reddens `TestComponentAssetImportCommandReproducesTheFixtureInput` with two assets present and did not with one. Confirmed both directions; reverted clean. Files: `folio-go/render_test.go`, `folio-go/fixture_test.go`, `folio-go/byte_neutrality_test.go`, `fixtures/component-asset-import/*`. |
| 8 | Major | **FIX** | New test passes an EXPLICIT `FileAccessBrowser` (bypassing `currentBrowser()`'s own bind) whose `showOpenFilePicker` is an unbound receiver-recording function, and asserts the receiver is that explicit object — what `selectImageFileAccess`'s own `.bind(browser)` actually guarantees, distinct from the pre-existing test that only observes `currentBrowser()`'s shared bind. Red-proved live: deleting that `.bind(browser)` reddens only the new test, leaving the old one green (proving it is a genuinely distinct check). Files: `folio-designer/src/image-file.test.ts`. |
| 9 | Major | **FIX** | Added a small, bounded Go-side discriminant (`ImageUnavailable`: `"missing"` \| `"undecodable"`) alongside the absent image paint, wired through strict protocol admission and both UI surfaces (inspector text, canvas placeholder), so a dangling asset reference is no longer reported with the same false "cannot render this media type" text a genuinely undecodable asset gets. Both Go tests for the absent-paint cases now assert the correct discriminant value. Files: `folio-go/page_setup.go`, `folio-go/canvas_image_paint_test.go`, `folio-designer/src/engine-protocol.ts`, `folio-designer/src/engine-protocol.test.ts`, `folio-designer/src/App.tsx`. |
| 10 | Major | **FIX** | Treated as the standing family the finding named. e2e gained real resize and zoom scenarios (drag was already covered per Finding 5); the unit geometry test (Finding 3) now covers zoom. Live-resize tracking (item 4 of the finding) was correctly noted as matching `TextPaint`'s existing, pre-story precedent and left as documented behaviour, not a defect. Files: `folio-designer/e2e/image-asset.spec.ts`, `folio-designer/src/App.test.tsx`. |
| 11 | Major | **FIX** | The ImagePaint unit test now covers all three named revocation triggers distinctly: asset replacement (assetKey change via a committed command, generation unchanged), document replacement (Start blank, SAME assetKey, isolating the `generation` dependency), and deletion (unmount) — each asserted to revoke the prior URL and mint a new one. Previously only unmount was covered. Files: `folio-designer/src/App.test.tsx`. |
| 12 | Minor | **FIX** | `isImagePaint`'s admission now requires `assetKey.length === 64` (was `1..64`), matching D-5.13.2 amendment's full-key wire contract and Go's `isAssetKeyShape`. Test fixture corrected off the loose 12-character key; added an explicit rejection case for a 12-character key alongside the existing malformed-hex case. Files: `folio-designer/src/engine-protocol.ts`, `folio-designer/src/engine-protocol.test.ts`. |
| 13 | Minor | **FIX** | `ImagePaint` now tracks fetch failure as its own state, distinct from "still loading," and renders a named failure placeholder instead of a permanent "Loading image…". Covered by the existing painted-image test's undecodable-asset branch continuing to pass and by the admission fix (12) removing the truncated-key cause in the first place. Files: `folio-designer/src/App.tsx`. |
| 14 | Minor | **FIX** | Folded into Finding 4's fix: `assetError`/`assetBusy` added to `setCurrentSnapshot`'s `clearDocumentInteraction` clears alongside every sibling error/busy pair. Files: `folio-designer/src/App.tsx`. |
| 15 | Minor | **FIX** | Unit test's SELECTION step now dispatches a real `keyDown('Enter')` on the focused component (exercising `CanvasComponent`'s actual keyboard handler) instead of a mouse click; renamed to state precisely what is proved. The button ACTIVATION half — native `<button>` Enter/Space-to-click translation — is real-browser behaviour jsdom does not synthesise from a bare keydown, so that half is now proved for real in the e2e suite: the "Choose image…" control is reached by focus and activated with an actual OS-level `Enter` key press, not a click. Files: `folio-designer/src/App.test.tsx`, `folio-designer/e2e/image-asset.spec.ts`. |
| 16 | Minor | **FIX** | `TestCanvasImagePaintTracksAResize` now additionally asserts the post-resize paint equals `resolveImagePlacement` run fresh against the resized element — the same correctness shape the shipping-run-path test uses — not merely "changed from before." Red-proved live: perturbing the post-resize draw rectangle by a fixed offset now fails this test where it previously passed on any change, including a wrong one; reverted clean. Files: `folio-go/canvas_image_paint_test.go`. |
| 17 | Minor | **FIX (documentation)** | Extracting a shared element-enumeration helper across `assetKeyReferenced`, `findComponent`, and `addCanvasImagePaint` was judged a wider refactor than this finding's severity warrants (Minor/Maintainability, no live defect). Took the suggested resolution's cheaper alternative: a doc comment on `assetKeyReferenced` binding it explicitly to the same three-band enumeration those two other functions use, so a future element location has one documented place to update in all three. Files: `folio-go/component_commands.go`. |
| 18 | Minor | **DEFER** | Real, accepted design cost (D-5.13.2 named "an extra request per distinct asset and its cache/lifetime handling" as the accepted price of a lossy snapshot) that shipped as per-ELEMENT rather than per-distinct-key, with no measured document that makes the difference matter yet. Fixing it now would be speculative scope expansion beyond this pass's mandatory/major fixes. **Deferred as DW-22** in `deferred-work.md`, owner: the second Epic 5 boundary gate (already owed by this story's reopening, D-5.13.6) — or whichever story first produces a document where the cost is actually observed. |
| 19 | Minor | **FIX** | File List extended with the four legitimate coordination artifacts (`sprint-status.yaml`, `epics.md`, `epic-6-boundary-gate.md`, `folio-mvp-decision-log.md`) and their one-line reasons, plus this finisher pass's own `deferred-work.md` addition; `evidence/story-6.7-roundtrip-manifest.json` confirmed excluded and NOT staged in this story's commit. Files: this story file. |
| 20 | Nit | **FIX** | Added the attribution note to both the Go test's doc comment and the fixture README: mutating the digest-as-key derivation reddens `TestComponentAssetImportCommandReproducesTheFixtureInput` through `ApplyComponentCommand`'s own reparse hitting `decodeAssets`'s PRE-EXISTING digest-match enforcement (Story 1.8), not new coverage this fixture adds — while the claim itself still holds. Files: `folio-go/fixture_test.go`, `fixtures/component-asset-import/README.md`. |
