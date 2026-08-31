# `embedded-font`

**The first `.folio` in this repository that carries a font face — and the first whose page is drawn
with one.**

Story 8.3 (FR53, FR56) made the face travel. Story 8.4 (FR54) made it draw. The document declares
one chain:

```json
"fonts": {
  "body": [
    "Noto Sans",
    { "asset": "c94562c15cbff8c9af93042adb1c63981b5deeeba40693ea8d98cd3b33b73caf" }
  ]
}
```

The second entry is the new shape. Its key is an entry of the top-level `assets` map — the same
mechanism images have always used: the key is the lowercase hex SHA-256 of the decoded bytes,
`data` is base64 hard-wrapped at 76 columns, and the asset carries a `font` record naming the
family, style, licence and source.

The bytes are the **shipped** Noto Sans Thai
(`folio-go/fonts/notosansthai/NotoSansThai-Regular.ttf`), embedded here as an asset rather than
supplied through the `FontSet`. **No new binary entered the repository for this fixture**, and
`input.folio` is not hand-transcribed: `embeddedFontTemplateJSON()`
(`folio-go/embedded_font_fixture_test.go`) derives it from those shipped bytes by the format's own
rules, and `TestEmbeddedFontFixtureMatchesInputFolio` pins the committed file against it.

## The text is pure Thai, and that is the whole measurement

The drawn value is `สัญญา` ("contract"). Measured from the shipped cmaps:

| face | codepoints covered in U+0E00–U+0E7F |
|---|---|
| `NotoSans-Regular.ttf` (the chain's **first** entry) | **0** |
| `NotoSansThai-Regular.ttf` (the face the document **carries**) | **87** |

So every rune on this page falls through the shipped entry and can only be drawn by the face the
document carries. **That is what makes the golden evidence.** Until Story 8.4 this document drew
*Latin* — which the carried Thai face also covers — so its bytes could not tell a build that renders
from a carried face from one that ignores it entirely, and its digest observed nothing. A **pure**
Thai string is the sharp witness; a mixed one is not.

`สัญญา` is also the **zero-offset control** `fixtures/thai-stacked-marks/` already documents: none of
its glyphs carries a vertical offset, so its run emits no ` Ts` operator and this golden stays out of
`textRiseExemptGoldens`.

## What it red-proves

- A face can **travel inside a template**: stored, loaded, canonically round-tripped and projected
  to the designer, through the `assets` mechanism, with no second storage shape.
- The engine **renders from it**: the font program on this page was base64-decoded out of the
  document, checked against the recognised-font set, parsed, subset once and embedded — with no
  network read, no host-installed font and no path on disk consulted for it.
- Resolution goes by **asset key, never by name** (AD-8, D-8.4.1). The carried face's `font.family`
  is `"Noto Sans Thai"` and the `FontSet` handed to the render holds a *shipped* face under exactly
  that name — and it is not the one on the page. The PDF's font resource is named from the **asset
  key**, which is why the produced bytes themselves are the identity witness.
- A document declaring an embedded-face entry declares **format version `2.0`**, and
  `TestEmbeddedFontFixtureIsCanonicalAndDeclaresTwoPointZero` asserts the declared version and the
  canonical fixed point together.
- The `font` record's four keys (`family`, `licence`, `source`, `style`) survive a round trip in
  sorted key order, and `font` sorts between `data` and `mediaType` without moving either.

## Asserted by identity, never by a count

`matrix_test.go`'s `requireEmbeddedFaceDrawsThePage` runs on **every** cross-target leg, before any
byte comparison, and asks the produced bytes *which* face drew the page: the resource name derived
from the asset key is present, the shipped Latin `+NotoSans-Regular` is absent, and `/BaseFont`
names `NotoSansThai-Regular`.

It replaced a guard that asserted the render embeds **exactly one** font program. That count is
`1` **both** before Story 8.4 and after it — the shipped Latin face then, the carried Thai face now
— so it passes identically on both implementations and certifies neither. Inverting it to `2` would
have been wrong for the same reason: the shipped Latin face covers none of this page, so a correct
build never subsets it.

## What it does NOT cover

**The designer canvas *painting* with the carried face.** The engine now **measures** with it — the
canvas paint projection's fragment origins and advances come from the render path's own
`fontChain`/`shapeSegments`/`chainVerticalModel`, asserted by
`folio-go/canvas_embedded_face_test.go` — but the browser has **no CSS family for a carried face at
all** and falls through to generic `sans-serif`. That is **Story 8.4a** (DW-35), and the gap is
recorded by a test rather than by a comment:
`folio-designer/src/canvas-font-stack.test.ts`.

## Files

| File | What it is |
|---|---|
| `input.folio` | The document. Canonical bytes; a serializer fixed point. |
| `expected.pdf` | The golden, recorded by Story 8.4 and registered in `goldenDigestRecord`. |
| `expected.json` | The recorded render hash, toolchain and library version for the matrix legs. |

The golden's digest is
`f533b04b7a4ccb20587f096c9e3173a48fbc870b8c718a73fecf869c6d851832`, at 3,225 bytes. Story 8.3
recorded `db400698567a45b5fc529849453c72845c4074958a9945390fced41459e513ad` at 55,513 bytes in
`expected.json` and shipped no `expected.pdf` at all. Both numbers moved for the same reason: the
drawn text became Thai, so the page is drawn with the carried face and only the five glyphs it
actually uses are subset into the file.

**No human reading sign-off is recorded for this page**, and that is stated rather than left to be
noticed. Every other Thai-bearing golden carries one (`fixtures/shaped-text/thai-signoff.json`,
`fixtures/thai-stacked-marks/signoff.json`); this one does not, and no agent may write one. What
stands in its place is a measurement: the drawn string `สัญญา` is byte-for-byte the string
`fixtures/thai-stacked-marks/` draws as its **zero-offset control**, at the same 12 pt, shaped by
the same face — the shipped `NotoSansThai-Regular.ttf`, of which this document simply carries a
second copy. Those five glyphs are therefore already read and attested; what is unattested is this
particular subset and page.

**All four legs were run in-story and all four produce
`f533b04b7a4ccb20587f096c9e3173a48fbc870b8c718a73fecf869c6d851832` at 3,225 bytes** — `darwin/arm64`,
`linux/amd64`, `linux/arm64` and `js/wasm`, with `TestCrossTargetByteIdentity` agreeing — rather than
merely the native leg D-000.54 requires. That is the recorded RESULT and not a statement about
procedure: "the legs are run" is true of a run that disagreed. The question those legs answer — does
a font program decoded out of the document's own base64, subset and embedded, survive four
toolchains identically? — is this document's whole reason to exist, so deferring them to the Epic 8
boundary gate would defer the only thing worth measuring.
