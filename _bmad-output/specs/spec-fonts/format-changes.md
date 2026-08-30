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
| `mediaType` | Closed set, to be fixed at implementation: `font/ttf` at minimum. A media type outside the set is a load error; bytes that do not decode as that type are a load error (CAP-4). |
| `font.family` / `font.style` | Display identity, for the designer's panel and for a human reading the file. Never used to resolve or substitute a face — resolution is by asset key alone. |
| `font.licence` | Licence identifier the redistribution rests on. Required on every font asset. |
| `font.source` | Where the face came from and which instance, so the embedding can be traced and replayed. |
| Order | `assets` emission order is unchanged (by key), so adding a font does not move an image. |

An image asset carries no `font` object, and a font asset carries one; the two are told apart by
`mediaType`, never by inspecting bytes at load.

## Version

The additions are additive — a document without font assets is unchanged, and a reader that
understands them reads every older document. That points at a MINOR bump, but the version rule is
a load-time error path (a higher MAJOR is a load error, never a best-effort render), so the choice
is recorded as an open question in `SPEC.md` rather than settled here.

## What does not change

- Canonical serialization (AD-9): same rules, same wrapping, same ordering, same byte-identical
  no-op round trip.
- Render-time behaviour (FR33): nothing is fetched, nothing is read from disk, no host font is
  enumerated.
- PDF embedding: the producer keeps subsetting once per document at render time. An embedded face
  and a shipped face reach that path identically — a `FontSet` entry is a `FontSet` entry.
- `style.fontFamily` still names a chain, never a face.
