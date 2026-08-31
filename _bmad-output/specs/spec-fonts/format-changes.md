# `.folio` format changes

Field-level additions for SPEC-fonts. The base contract is [`../spec-folio/folio-format.md`](../spec-folio/folio-format.md);
everything not restated here is unchanged.

## `fonts` — chain entries may name an asset

Today a chain is an ordered list of **shipped face names**, resolved per rune for coverage:

```json
"fonts": { "body": ["Noto Sans", "Noto Sans Thai", "Noto Sans SC"] }
```

An entry may instead reference an embedded face by its `assets` key, written as a one-key object
so a name and a reference can never be confused for one another:

```json
"fonts": {
  "body": ["Noto Sans", "Noto Sans Thai", "Noto Sans SC"],
  "brand": [{ "asset": "a31f6086…c8bc27" }, "Noto Sans Thai", "Noto Sans SC"]
}
```

| Rule | Behaviour |
|---|---|
| Entry shapes | A JSON string is a shipped face name. A one-key object `{"asset": "<key>"}` is an embedded face. Any other shape is a load error naming the chain and the index. |
| Mixing | A chain may mix both shapes freely — an embedded Latin face over shipped CJK fallback is the expected case. |
| Order | Unchanged: first entry that covers the rune wins. |
| Absent key | A `{"asset": …}` entry whose key is not in `assets` is a load error naming the chain, the index and the key (CAP-4). |
| Absent shipped face | Unchanged from today's tolerance: a shipped name absent from the supplied `FontSet` is skipped, and a chain that ends up with no usable entry is the existing located error. |
| Empty chain | Unchanged: a chain with no entries is not a chain `fontFamily` may name. |

## `assets` — a font entry

The map, its key rule (lowercase hex SHA-256 of the raw bytes), its 76-column base64 wrapping, its
deduplication and its emission order are unchanged. A font entry differs from an image entry only
in `mediaType` and in carrying a licence record:

```json
"assets": {
  "a31f6086…c8bc27": {
    "data": ["AAEAAAAQAQAABAAA…"],
    "mediaType": "font/ttf",
    "font": {
      "family": "Inter",
      "style": "Regular",
      "licence": "OFL-1.1",
      "source": "Google Fonts, Inter v4.0, static Regular instance"
    }
  }
}
```

| Field | Rule |
|---|---|
| `mediaType` | **OPEN set, corrected at Story 8.3.** The "closed set" wording here contradicted binding D-1.8.1 (as amended), whose own note predicted this exact recurrence "later for font formats": a closed `mediaType` could only ever be extended by a MAJOR bump, making every new font container a breaking format change. The ruled position: a RECOGNISED font media type (`font/ttf` at minimum) whose bytes are not that format is a load error (CAP-4 holds for those); an UNRECOGNISED one loads clean and is preserved verbatim, and errors only at render, and only when a render actually needs that face. |
| `font.family` / `font.style` | Display identity, for the designer's panel and for a human reading the file. Never used to resolve or substitute a face — resolution is by asset key alone. |
| `font.licence` | Licence identifier the redistribution rests on. **OPTIONAL, corrected at Story 8.3** — as is `font` itself and every key inside it (`folio-format.md`, *A font asset*). "Required on every font asset" was a load rule this format does not have and the loader does not enforce: `TestPlainFontAssetNeedsNoRecord` pins a font asset with no `font` record at all round-tripping without the key. It remains the field a redistributor should fill in, and that is an authoring convention for whatever writes the asset — not something a reader may refuse a document over. |
| `font.source` | Where the face came from and which instance, so the embedding can be traced and replayed. |
| Order | `assets` emission order is unchanged (by key), so adding a font does not move an image. |

An image asset carries no `font` object, and a font asset carries one; the two are told apart by
`mediaType`, never by inspecting bytes at load.

## Version

**SETTLED at Story 8.3: MAJOR — a document carrying an embedded-face chain entry declares `2.0`.**

The `assets` additions really are additive, and taken alone they would point at a MINOR. The chain
ENTRY is not: `{"asset": …}` changes the legal shape of an existing value, and a pre-`2.0` reader
decodes a chain entry as a string and never coerces, so it REFUSES the file rather than mis-drawing
it — D-7.3.1's pre-reader test, failed harder than `align: "justify"` failed it. It joins the `2.0`
Story 7.3 already opened (owner decision D-R7.9: "Story 8.3 joins the same 2.0") rather than opening
a `3.0`, and `SupportedMajor` does not move.

Per D-1.4.13 the trigger is the ENTRY, not the asset: a document carrying a font asset that no chain
references stays at whatever its other content requires, because such a document loads and renders
correctly on a `1.x` reader.

## What does not change

- Canonical serialization (AD-9): same rules, same wrapping, same ordering, same byte-identical
  no-op round trip.
- Render-time behaviour (FR33): nothing is fetched, nothing is read from disk, no host font is
  enumerated.
- PDF embedding: the producer keeps subsetting once per document at render time. An embedded face
  and a shipped face reach that path identically — a `FontSet` entry is a `FontSet` entry.
- `style.fontFamily` still names a chain, never a face.
