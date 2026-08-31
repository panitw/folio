# `embedded-font`

**The first `.folio` in this repository that carries a font face.**

Story 8.3 (FR53, FR56). The document declares one chain:

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

## What it red-proves

- A face can **travel inside a template**: stored, loaded, canonically round-tripped and projected
  to the designer, through the `assets` mechanism, with no second storage shape.
- A document declaring an embedded-face entry declares **format version `2.0`**. That is the
  version its content requires, and `TestEmbeddedFontFixtureIsCanonicalAndDeclaresTwoPointZero`
  asserts the declared version and the canonical fixed point together.
- The `font` record's four keys (`family`, `licence`, `source`, `style`) survive a round trip in
  sorted key order, and `font` sorts between `data` and `mediaType` without moving either.

## What it does NOT cover

**Rendering from the embedded face. That is Story 8.4.**

The chain's *first* entry is the shipped `Noto Sans`, and that is the face the page is actually
drawn with. The embedded entry contributes nothing: at Story 8.3 an embedded entry resolves to no
face, exactly as a named face absent from the `FontSet` resolves to none. That is the honest
interim state, and it is asserted rather than assumed —
`matrix_test.go`'s `requireEmbeddedFaceStaysOffThePage` runs on **every** cross-target leg and
requires the rendered PDF to embed **exactly one** font program. **When Story 8.4 lands, that guard
must change, deliberately.**

The text on the page is Latin, and the carried face is Thai, on purpose: the face the document
carries is the one the page does not need, so an implementation that quietly started drawing with
it would move this fixture's bytes rather than be absorbed by them.

## Why there is no `expected.pdf`

On the `hidden-image` precedent (Story 3.5). An `expected.pdf` is a **human-attested** artifact
under AD-21 / D-4.7.1, and this story cannot produce the one that would matter — a page drawn
*with* the embedded face. Recording a page drawn with the shipped face under the name
`embedded-font` would attest the wrong thing. So this directory ships `input.folio` and
`expected.json` only, `goldenDigestRecord` stays at **22**, and this fixture's acceptance is
structural: it is registered in `matrixDocuments` with a recorded cross-target hash.

## Files

| File | What it is |
|---|---|
| `input.folio` | The document. Canonical bytes; a serializer fixed point. |
| `expected.json` | The recorded render hash, toolchain and library version for the matrix legs. |

**All four legs were run in-story**, not merely the native one D-000.54 requires: `darwin/arm64`,
`linux/amd64`, `linux/arm64` and `js/wasm` all produce
`db400698567a45b5fc529849453c72845c4074958a9945390fced41459e513ad` at 55,513 bytes, and
`TestCrossTargetByteIdentity` agrees. The question those legs answer — does ~47 KB of carried asset
survive four toolchains identically? — is this document's whole reason to exist, so deferring them
to the Epic 8 boundary gate would have deferred the only thing worth measuring.
