---
baseline_commit: 54a0777a4229757f5b46a77980245fbc163c0d74
---

# Story 1.8: Embed and render an image from the template

Status: done

**Story key:** `1-8-embed-and-render-an-image-from-the-template`
**Epic:** 1 — **this is the last story in Epic 1** · **Baseline:** `54a0777`
**Covers:** FR33 · **AD-9, AD-24**
**Adjacent:** AD-2, AD-3, AD-6, AD-21, AD-22, AD-26, §Source tree
**Rulings that govern:** **D-1.8.1** *(and its amendment)*, **D-1.8.2**, **D-1.8.3**, **D-1.8.4**,
**D-1.8.5**, **D-1.8.6**, **D-1.8.7**, **D-1.8.8**, **D-1.8.9**, **D-1.8.10**, **D-1.8.11**,
D-000.4, D-000.6, D-000.9 *(and its extension)*, D-000.10, D-000.11, D-000.12, D-000.13, D-000.14
*(and its extension)*, D-1.1.b, D-1.2.6, D-1.3.1, D-1.3.5, D-1.4.3, D-1.4.5, D-1.4.7, D-1.4.9,
D-1.4.10, D-1.4.12, D-1.5.2, D-1.7.3
**Risk under test:** **R4** — *"Compressor output is stable by observation, not by contract — a
future Go release could invalidate every recorded golden hash downstream"* (`acceptance.md:83`,
**verified**). D-1.8.1's passthrough design exists to keep R4 closed in this story.
**Heavy-test cadence:** **D-000.4 OVERRIDE — the full four-target matrix runs IN THIS STORY**, over
**all three** documents.

---

> ### Where these rulings live — read before diffing
>
> **D-1.8.1 – D-1.8.11, the D-1.8.1 amendment, the D-1.7.1 amendment and the D-000.14 extension are
> UNCOMMITTED working-tree additions to
> `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`. They are NOT in `HEAD`, and
> `HEAD` IS baseline `54a0777`.** A reviewer who diffs against `54a0777` **will not find any of
> them** and must read the working-tree file instead.
>
> **Measured during story creation, not inherited from the brief** — three previous creators on this
> project were handed a wrong claim about this file's state, so it was re-measured from scratch, and
> **re-measured again** after D-1.8.1 (amended) and D-1.8.7 – D-1.8.11 were ruled mid-creation:
>
> | Check | Command | Result |
> |---|---|---|
> | `HEAD` is the stated baseline | `git rev-parse HEAD` | `54a0777a4229757f5b46a77980245fbc163c0d74` |
> | Which files differ from `HEAD` at all | `git status --porcelain` | exactly one: ` M …/folio-mvp-decision-log.md` |
> | The change is a pure append | `git diff --unified=0 54a0777 -- …/folio-mvp-decision-log.md` | one hunk: `@@ -3597,0 +3598,298 @@` |
> | Committed file length | `git show 54a0777:…/folio-mvp-decision-log.md \| wc -l` | `3597` |
> | Working-tree additions | — | lines **3598–3895**, **zero deletions** |
> | The Story 1.8 rulings are absent from `HEAD` | `git show 54a0777:…/folio-mvp-decision-log.md \| grep "D-1\.8\."` | **no matches** |
> | First-round rulings, present in the working tree | headings at **3634, 3674, 3696, 3717, 3734, 3754** | D-1.8.1 … D-1.8.6 |
> | Second-round rulings, added during this story's creation | headings at **3768, 3816, 3840, 3854, 3865, 3884** | **D-1.8.1 (amended)**, D-1.8.7 … D-1.8.11 |
> | What else is in the same append | headings at **3599, 3614** | **D-1.7.1 (amended)**, **D-000.14 (extended)** |
>
> **The first-round line count in an earlier draft of this banner was `169`; it is now `298`.** The
> row is corrected rather than left stale, and the delta is the second-round rulings. **Re-measure
> rather than trusting either number** — this file's state changed twice during its own creation.
>
> **No Go file, no fixture, no spec file differs from `HEAD`.** The tree is otherwise clean.
> These rulings will land in a commit that is **not necessarily this story's**.

> ### Mechanism tagging (D-1.4.15's convention)
>
> Every named mechanism below carries `(binding)` or `(illustrative)`. **An illustrative mechanism
> may be improved on; a binding one may not be departed from without surfacing it as
> `DECISION NEEDED`** (D-1.2.6). Where this story quotes a ruling, the tag travels with the quote.
>
> **Nothing is parked.** This story's creator surfaced one `DECISION NEEDED` — a conflict between
> **D-1.8.1** and the **owner decision D-1.4.9** (with D-1.4.12 and D-1.8.3 as corroborating faces) —
> and per D-1.2.6 did not arbitrate it. **It was ruled while this file was being written, as
> D-1.8.1 (amended): an unrecognised `mediaType` is never a load error, and `Validate` is a dry-run
> predictor of `Render`.** The ruling is carried at
> [§ Decisions resolved](#decisions-resolved), and its obligations are ACs in
> [§ Section C](#c--formats-by-passthrough-d-181--and-its-amendment--ad-6-ad-22--r4--source-ac3).
> The creator's four other findings were confirmed and logged as **D-1.8.7 – D-1.8.11**; their
> obligations are folded into Sections A, B, D and G.

---

## In plain terms (read this first if you just want the gist)

This story taught the renderer to actually draw a picture already sitting inside a template file. A
template could already name a picture and the loader accepted it without complaint, but nothing
checked the picture was genuinely there or intact, and when a page was produced, pictures were
silently skipped. That gap is closed: a logo carried inside a template's own file appears on the
produced page.

The picture travels as text inside the template, so the file stays valid, readable, and diffs
cleanly outside the picture itself. It is filed under a fingerprint of its own contents, so the same
logo used twice is stored once and both places point at it; on loading, that fingerprint is
recomputed and has to match, or the file is refused rather than silently drawn wrong. A hand-wrapped
file is accepted and rewritten in house style rather than rejected.

Two kinds of picture are supported — ordinary photographs and plain, fully-opaque still images — and
in both cases the already-compressed data is carried into the page untouched; nothing is ever
re-compressed, because our own compressor's output is steady only by observation, not by promise,
and this was the wrong place to gamble on that. Pictures with transparency are deliberately refused,
by name, with a reason, since that is the commonest shape a company logo takes and the owner needed
to know now rather than discover it later. A picture is scaled to fit its box and centred using
whole-number arithmetic throughout.

A second pass, after review, found and closed several real gaps that had shipped looking finished
but weren't: a crash reachable from ordinary corrupted input, on the boundary meant to turn bad files
into readable errors instead of exceptions; a geometric rule with no working proof behind it, so a
broken future edit would have sailed through unnoticed; and a placement safety check that recognised
only one shape of violation and missed two others. All were fixed, proven to actually catch what they
claim to catch, and re-verified live. Nothing here is deferred or left deliberately red.

---

## Story

As a template author,
I want my logo carried inside the `.folio` file itself,
So that the template is self-contained and rendering never depends on a URL or a file on disk.

---

## Acceptance Criteria

Source ACs are `_bmad-output/planning-artifacts/epics.md` §Story 1.8 (lines 637–665). The source set
is **four** `Given/When/Then` groups (lines 647–651, 653–655, 657–660, 662–665). They are carried
below unchanged in substance and split so each is separately red-proofable, with the
ruling-mandated additions folded in. **No expanded total is narrated anywhere in this file** — per
**D-000.14 (extended)**, *"A disposition table's total is computed, or omitted entirely — never
narrated alongside it."* Counts here are omitted.

Every AC is stated so the three diagnostic questions (§Standing rules) have an answer other than
*"the same thing"*.

**Read §Dev Notes before starting.** M-1 through M-4 record that a substantial part of the assets
schema **already shipped at Story 1.4**, and M-5 through M-7 record that **every asset-bearing
fixture in the tree today, and the format specification's own example, are made invalid by these
rulings**. Section G exists because of that.

---

### A — The asset payload is base64, canonical at 76 columns (D-1.8.2 · AD-9 · FR33 · source AC1)

**AC1 — serialize: RFC 4648 §4, with padding, split at exactly 76** *(binding, D-1.8.2)*. Quoted:

> **On serialize:** standard base64 (RFC 4648 §4, **with** padding, not URL-safe), split into
> elements of **exactly 76 characters**, the final element carrying the remainder (1–76). **Padding
> falls where it falls — inside the split, at the end of the last element.** Never stripped, never
> moved, never given its own element.

The test asserts, on an asset whose base64 length is **not** a multiple of 76 and whose padding
therefore lands mid-final-element: every element except the last has length exactly 76; the last
has length in 1–76; concatenation with nothing inserted reproduces the base64 exactly; and the `=`
padding characters appear **only** at the end of the last element.

**AC2 — parse: accept any wrapping, normalise on save** *(binding, D-1.8.2)*. Quoted:

> **On parse:** accept **any** wrapping. Concatenate elements with nothing inserted, then decode
> strictly. A file hand-wrapped at 64 columns **loads** and **re-serializes at 76**.

**This must not be tightened** — D-1.8.2: *"A parser that enforced 76 columns would reject a legal
hand-edit"*, and AD-9's Prevents line explicitly contemplates a hand-editor. A parser that rejects
non-76 wrapping is a **defect**, not a stricter reading.

**AC3 — the non-vacuous half of the corpus obligation** *(binding, D-1.8.2)*. Quoted:

> **Round-trip obligation:** the corpus gains an asset-bearing fixture satisfying **P1, P2 and
> P3** — and **a hand-wrapped variant that normalises to the same canonical bytes is the non-vacuous
> half**, since a canonical-in/canonical-out test alone would pass on a serializer that merely
> echoed its input.

**This warning is not hypothetical — it describes the shipped serializer exactly.** See **M-4**:
`writeAssets` today emits `a.Data` verbatim as a string array and performs **no re-wrapping at
all**. So the canonical-in/canonical-out half of this AC **passes at baseline, vacuously**.

**AC3a — the serializer re-wraps from the DECODED bytes and never echoes the input array**
*(binding, D-1.8.9)*. This is the mechanism, not merely the outcome: the write path decodes
`Asset.Data`, then re-splits the base64 of those bytes at 76. An implementation that inspects the
incoming elements and "fixes them up" is not this. D-1.8.9's discriminator, quoted:

> **Canonical-in/canonical-out cannot fail against an echoing serializer; 64-in/76-out cannot pass
> against one.**

So the retained fixture is **wrapped at 64 columns and must serialize to 76**, and it is the AC's
red-proof (RP-2). A canonical-only corpus addition does not discharge AC3.

**AC4 — invalid base64 and an empty decoded asset are load errors** *(binding, D-1.8.2)*. Bad
alphabet, bad padding, and embedded whitespace are load errors; so is **an asset whose decoded data
is empty** — D-1.8.2: *"it cannot render, and its key would be the SHA-256 of nothing."* Each is a
`template.LoadError` naming the asset key (M-3 gives the constructor and the existing
`assets.<key>.data` field-path convention). **Strict decoding** — the standard-encoding strict
decoder, not a tolerant one.

---

### B — Keys, dimensions and orphans (D-1.8.3 · AD-9 · source AC1, AC2)

**AC5 — the key is the SHA-256 of the DECODED bytes** *(binding, D-1.8.3)*. Lowercase hex.
**Verified independently during story creation:** `folio-format.md:219` reads *"Keyed by the
**lowercase hex SHA-256 of the raw bytes**"*, and D-1.8.3 rules that *"raw means decoded, not the
base64 text"*. Two identical images wrapped differently must produce the **same** key, which only
holds over decoded bytes — so the dedup AC (source AC2) is discharged by this definition rather
than by a separate mechanism.

**AC6 — validate the key on load; mismatch is a load error** *(binding, D-1.8.3)*. Recompute and
require equality. D-1.8.3: this makes dedup *"correct **by construction rather than by trust**"*,
and catches *"a hand-edit that changed the bytes but not the key — the exact corruption that would
otherwise surface as two elements silently sharing one wrong image."* The error names the asset key,
the expected digest and the found digest.

**AC6a — validate SHAPE and VALUE, as two distinct error classes** *(binding, D-1.8.8)*. Quoted:

> **Validate shape *and* value**: 64 lowercase hex characters **and** equal to the SHA-256 of the
> decoded bytes. **A 62-character key means no check validated shape either** — a different error
> class from a wrong-but-well-formed digest, and the shape check is the cheaper one.

This **reverses** an earlier draft of this AC, which folded shape into value on the reasoning that a
non-hex key cannot equal any digest. That is true and it is not the point: a 62-character key is
evidence that **nothing looked at the key at all**, which is a different diagnosis from a key that
was computed and drifted. The two messages must be distinguishable — *"is not a 64-character
lowercase hex digest"* versus *"does not match the SHA-256 of its data"* — and both are red-proofed
separately (RP-4, RP-4a). The shape check runs first and is cheap.

**AC7 — `/Width` and `/Height` go through `appendInt`** *(binding, D-1.8.3, D-1.1.b)*. They are
pixel counts. **The routing row already exists in the committed log** — `folio-mvp-decision-log.md`
line 696, in D-1.1.b's table: `| Image /Width, /Height — pixel counts (Story 1.8) | appendInt |`.
D-1.8.3 names the trap it pre-empts: *"routing a pixel count through `appendLength` would emit
`0.8` for an 800-pixel image."* AD-3's own Integer rule (spine) already names *"pixel dimension"*
among the integer route's values.

**AC8 — orphaned assets are PRESERVED** *(binding, D-1.8.3)*. Quoted:

> **Orphaned assets are PRESERVED — forced, not chosen.** **D-1.4.3's P1 requires
> `Parse(Serialize(d)) == d`**, so a serializer dropping unreferenced assets would violate P1 for
> every document containing one. **There is no policy latitude here.**

A retained fixture carries an asset referenced by **no** element and asserts P1, P2 and P3 hold for
it. The red-proof is a serializer that drops unreferenced assets (RP-3).

**AC8a — the Epics 5/6 corollary is recorded in a code comment** *(binding, D-1.8.3)*, verbatim:
*"garbage-collecting orphans is a designer feature — an explicit user action — never a serializer
side effect."* Placed on the serializer's assets writer so a later agent adding a "tidy unused
assets" pass reads it first.

---

### C — Formats, by passthrough (D-1.8.1 and its amendment · AD-6, AD-22 · R4 · source AC3)

**AC9 — the format table** *(binding for the format set and the failure mode; illustrative for the
PDF filter details, D-1.8.1 as amended)*:

| Input | Route | Why |
|---|---|---|
| **JPEG** | `/DCTDecode`, bytes embedded **unchanged** | zero re-encoding, byte-stable by construction |
| **PNG** 8-bit RGB/Gray, non-interlaced, **no alpha** | `/FlateDecode` + `/DecodeParms /Predictor 15`, IDAT zlib stream **passed through** | PNG's own filtering maps onto PDF's predictor; **no compressor is invoked** |
| A **recognised** `mediaType` whose bytes are not that format | **Load error**, located, naming the asset key and the reason | **AC11b** — the file lies about itself; reader-independent |
| PNG with alpha, palette, 16-bit or interlaced | **Load error** *(see AC9a)*, located, naming the asset key and the reason | it *is* a PNG and we *do* recognise `image/png`; the bytes are read and rejected on their own content |
| Any **other** media type | **Never a load error** — see **AC11**, the three surfaces | reader-independent validity (D-1.8.1 amended) |

**AC9a — where alpha sits on the validity/capability line, stated because it is the near-miss.** An
alpha PNG arrives under a **recognised** `mediaType`, so the loader parses it and rejects it on its
content — it is not the open-set case AC11 governs. This is a deliberate reading of the amended
ruling, which relaxed **only** the *unrecognised `mediaType`* clause and explicitly did **not** relax
the recognised-but-unparseable one. If a reviewer reads the amendment as also moving alpha to render
time, that is a **ruling-interpretation question to surface, not to settle in the diff** (D-1.2.6).
The alpha *message* obligation (AC12) is unaffected either way.

The PNG route concatenates **all** `IDAT` chunks in order into one zlib stream and emits that as the
XObject's stream body; `/DecodeParms` carries `/Predictor 15`, `/Colors`, `/BitsPerComponent` and
`/Columns` derived from `IHDR`. **`/Filter` arrives in this codebase for the first time here** — see
M-8: no `/Filter`, `/XObject`, `/DCTDecode`, `/FlateDecode`, `/ColorSpace` or `/BitsPerComponent`
exists anywhere in the shipped tree today.

**AC10 — NO COMPRESSOR IS INVOKED, and it is asserted mechanically** *(binding — this is D-1.8.1's
"real content")*. Quoted:

> Decoding and re-encoding a PNG would invoke our compressor and **materialise R4 in the last story
> of Epic 1**. Passing the existing zlib stream through means the compressed bytes come **from the
> file, not from a compressor** — determinism is inherited from the input and R4 stays closed.
> **Take this route even where re-encoding would be simpler.**

An **AST guard** (D-000.14: load-bearing properties are established by AST, never text match)
asserts that **no file under `folio-go/` imports `compress/flate`, `compress/zlib` or
`compress/gzip`** *(binding)*, and — as the illustrative second half — that no non-test file imports
`image`, `image/png` or `image/jpeg` *(illustrative)*, since reaching for a stdlib decoder is the
concrete shape the "simpler" re-encoding route takes. The guard reports **how many files it
examined**, and **zero examined is a hard failure, not a pass** (D-000.9 obligation 1). Baseline
measured: none of these is imported anywhere today (M-8), so the guard is green at baseline **and
its red-proof must be a real import in a real file** (RP-5), not an assertion about a string.

**AC10a — the R4 rationale is recorded where the filter is emitted** *(illustrative)*. A comment at
the `/FlateDecode` emission site states that the bytes are the file's own IDAT stream, cites D-1.8.1
and R4, and states that substituting a re-encode is a hash-changing versioned event under AD-22.
Without it, a later agent "simplifying" the chunk walker into `image/png` + `compress/zlib` has
nothing to read.

**AC11 — `mediaType` must NOT become a closed set** *(binding, D-1.8.1)*. Quoted:

> Under **D-1.4.12** a closed set can only be extended by a **MAJOR** bump, so freezing `mediaType`
> now would make "add WebP" a breaking format change **forever**. Distinguish two failures instead:
> **format validity** (the `.folio` is well-formed; `mediaType` stays open) versus **library
> capability** (*this version* cannot render that media type, reported as a located error saying
> exactly that).

Mechanically: `mediaType` **must not** be added to `internal/template/closedsets.go`, and a test
asserts it is absent from that file's set inventory. The capability error's message must say
*this version cannot render this media type* — not *invalid mediaType* — so the two failures are
distinguishable by a reader.

**AC11a — an unrecognised `mediaType` is NEVER a load error; three surfaces, three answers**
*(binding, D-1.8.1 amended)*. This resolves the `DECISION NEEDED` this story raised. The verdict
table, verbatim:

| Surface | Unrecognised `mediaType` |
|---|---|
| `LoadTemplate` / `ParseTemplate` | **Never refused.** The document is valid; the asset is preserved verbatim (P1 requires it regardless) |
| `Render` | **Error — only when an element actually needs to draw it.** Located, naming element id, asset key and media type |
| `Validate` | Reports **what `Render` would do** |

**The reason is the reason, and it must survive into the code comment** *(binding)*: under the
unamended rule *"the same bytes are valid or invalid depending on which library reads them"*, and
`folio-format.md` opens by calling the format **"a public contract, not an implementation
detail"** — *"a public contract whose validity depends on the reader is not a contract."* **"Valid
`.folio`" must mean something stateable without naming a library version.**

Scope, unchanged from the creator's design: **one named predicate, one call site.** The loader does
not call it; the render path does.

**AC11b — the contrast case, which is NOT relaxed** *(binding, D-1.8.1 amended)*. A **recognised**
`mediaType` whose bytes do not parse as that format — a file declaring `image/png` that is not a
PNG — **is** a load error. Quoted:

> That failure is **reader-independent**: the file lies about itself and every library agrees it is
> malformed. Cheap to check via magic bytes and header for the two supported formats. **The contrast
> between these two cases is the clearest illustration of the validity/capability line.**

**Put the two side by side in the test file, adjacent and named as a pair**, because separated they
read as an inconsistency:

| Case | `LoadTemplate` | Why |
|---|---|---|
| `mediaType: "image/webp"`, valid WebP bytes | **loads** | we do not recognise the type — a *capability* limit of this version, not a property of the file |
| `mediaType: "image/png"`, bytes that are not a PNG | **load error** | the file contradicts itself — *every* library agrees, at every version |

**AC11c — `Validate` is a dry-run predictor of `Render`, not a second rule system** *(binding,
D-1.8.1 amended)*. Recorded now because **Story 3.7 builds `Validate`** and this is the ruling that
governs it. Quoted:

> An unrenderable asset nothing draws → `Render` succeeds → `Validate` is clean. One that is drawn →
> `Render` errors → `Validate` errors, **in CI, before production** — so Story 1.4's loud-failure
> posture is fully preserved *without* making validity reader-dependent.

Two consequences that **fall out rather than being special-cased**:

- **A hidden element never triggers it.** AD-24: *"A hidden element is **absent** from the
  `PageModel`."* Nothing draws it, so nothing errors — **no special case needed.**
- **`Validate` without data cannot always predict** a `visibleIf`-gated image; where data-dependence
  makes prediction impossible the diagnostic degrades to a **Warning**. *(illustrative — a **Story
  3.7 detail**. Do not build it here, and do not build a `Validate` entry point here either.)*

Story 1.8's obligation is only to **write the rule down where 3.7 will find it** and to keep the
render-path error shaped so a dry run can predict it — i.e. the predicate is callable without
rendering. **Do not create `folio.Validate` in this story.**

**AC12 — alpha is a STATED limitation, recorded for the Epic 1 boundary report** *(illustrative,
D-1.8.1)*. Quoted:

> **But a transparent PNG is the single most common form a company logo takes, and the golden
> report's header carries a logo (Epic 4).** Not a one-way door — `/SMask` is purely additive
> later — so it does not block. **Surfaced to the owner while the golden report is still being
> designed**, not discovered at Epic 4.

The refusal message for an alpha PNG names alpha explicitly. The Delivery Log records it as an item
for the **Epic 1 boundary report to the owner**, together with M-11's finding about the licence
manifest's extension list. **Record it as a stated limitation, not a silent one.**

---

### D — Fit, centre, and the one rounding mode (D-1.8.4 · AD-2, AD-24 · source AC4)

**AC13 — choose the binding axis by cross-multiplication** *(binding, D-1.8.4)*. Image `W×H`
pixels, box `bw×bh` millipoints: compare `bw*H` against `bh*W`. **Exact integer comparison, no
division, no float.** Then exactly one `ScaleRound` call: width binds → `dw = bw`,
`dh = ScaleRound(bw, H, W)`; height binds → `dh = bh`, `dw = ScaleRound(bh, W, H)`.

**Verified during story creation** (M-9): `ScaleRound(v Length, num, den int64) Length` computes
`v*num/den` with round-half-to-even, so the call shapes above read as written — `num` is the
numerator, `den` the denominator.

**AC14 — the centring offsets are the second rounding site and must not be** *(binding, D-1.8.4)*.
Quoted:

> `(bw - dw) / 2` **truncates** when the difference is odd — a second rounding behaviour smuggled in
> as arithmetic. **Route it through the same function: `ScaleRound(bw-dw, 1, 2)`**, so
> round-half-to-even applies and the program keeps **exactly one rounding mode in exactly one
> function** (AD-2, D-1.4.5).

**Corrected during finishing (Finding 2, Story 1.8 review — the original wording here described a
state the geometry cannot produce, and the fixture that shipped with it was not discriminating):**
`resolveImagePlacement` always sets the BOUND axis's drawn dimension equal to the box exactly
(`drawW = bw`, or `drawH = bh`), so that axis's own centring offset is always structurally zero —
"both axes odd" is unreachable by construction; only the free (non-bound) axis's offset can ever be
nonzero. Nor is "an odd difference" alone sufficient: measured directly against `ScaleRound`,
round-half-to-even and truncation **agree** on most odd values (e.g. 5, 9) and disagree only when the
difference is **≡ 3 (mod 4)** (e.g. 3, 7, 11). The retained fixture must therefore use a free-axis
difference ≡ 3 (mod 4) — the story's fixture uses 7 (`ScaleRound` gives 4, `/2` gives 3). The
red-proof (RP-6) replaces both `ScaleRound` calls with `/2` and the test must redden **on the offset
value**, naming it, with the expected value computed as a **literal integer**, never re-derived from
`ScaleRound` itself — never on a bare failure (D-000.13).

**AC15 — validate `W`, `H` at the trust boundary FIRST, reusing D-1.5.2's precedent exactly**
*(binding, D-1.8.4)*. Quoted:

> `ScaleRound`'s panics are programmer-error guards, not input handling, so a `W == 0` from a corrupt
> header must be a **located load error, never a panic**. Same seam, same reasoning, one story later.

`W > 0`, `H > 0`, and **bounded**. The bound is not arbitrary: `ScaleRound` panics when `v*num`
overflows `int64` (M-9), and here `v*num` is `bw*H` — so an unbounded `H` reaches a panic through a
document, which is exactly what this AC forbids. An A4 page box is under `10^6` millipoints, so a
pixel bound of `10^6` keeps `bw*H` under `10^12`, four orders of magnitude clear of `int64`. Any
bound the developer can defend arithmetically is acceptable; **an unstated one is not**.

**AC15a — the validated dimensions must be the ONLY ones reachable** *(binding, D-1.5.2's
guardrail, applied)*. Quoted from D-1.5.2: *"the parsed-font type must not expose a raw, unvalidated
`unitsPerEm` that a caller can hand to `ScaleRound`. Validation happens in the constructor and the
validated value is the **only** one reachable. A field that can be read before it is checked will
eventually be read before it is checked."* The decoded-image type follows the same shape: `W`/`H`
are unexported and validated in the constructor; there is no accessor to a raw header value.
`internal/fontset` (M-10) is the code precedent — including its second half, that the raw value must
be read from the header directly rather than through an accessor that launders it.

**AC15b — retained fixtures, both polarities** *(binding, D-1.5.2's shape)*. A header with `W == 0`
produces a **located load error and not a panic**; a valid image scales correctly; plus a
negative case above the bound. The zero case is the one that would panic today.

**AC16 — never stretched, never cropped, never sized from intrinsic pixels** (source AC4, AD-24).
AD-24's Images clause, verified verbatim in the spine: *"An image is scaled to **fit** its declared
box preserving aspect ratio and centred within it, computed in integer millipoints. Never stretched,
never cropped, never sized from its intrinsic pixel dimensions."* The test asserts the drawn
`dw:dh` ratio matches `W:H` by **cross-multiplication** (`dw*H == dh*W` up to the single permitted
rounding), and that `dw <= bw && dh <= bh`.

**AC17 — the coordinate flip stays in EXACTLY ONE function** *(binding — AD-24, applied; confirmed
by D-1.8.10)*. AD-24, verbatim: *"The flip to PDF user space (bottom-left, Y up) happens in
**exactly one** function in `internal/pdf`, and nowhere else in the module inverts a coordinate."*

**Measured (M-12): today exactly one function inverts a coordinate**, and only because there has
only ever been one content-stream builder. An image placement needs the same flip. **Adding a
second flip site is a direct AD-24 violation** and is the single likeliest structural defect in this
story. The flip is extracted into one named helper that both the text path and the image path call.

**AC17a — the guard is POSITIVE, not negative** *(binding, D-1.8.10)*. Quoted:

> assert that **all** content-stream coordinate emission routes through the one flip function —
> **not** that "nobody else writes a minus sign", which is unfalsifiable and will produce false
> positives. Same shape as D-1.1.b's `numbers.go` rule.

So the AST guard enumerates every content-stream **coordinate-emitting** site in `internal/pdf` and
asserts each obtains its Y from the one flip helper — a routing assertion of the same shape as the
existing "every number reaches output through `numbers.go`" rule (M-15), which is the working
precedent to copy. A search for subtraction expressions is **explicitly rejected**: it cannot be
falsified and it will fire on legitimate arithmetic. The guard reports **how many candidate
emission sites it considered**, and **zero considered is a hard failure** (D-000.9). Red-proof RP-7
open-codes a second flip in the image path; the guard must redden naming **both** sites.

**Corrected during finishing (Finding 3, Story 1.8 review):** the shipped guard's first version keyed
"candidate site" on a naming convention — a bare, `Y`-suffixed identifier passed directly to
`appendLength`/`writeLength`. Two of the three shapes a real bypass can take (a differently-named
variable; an inline expression with no variable at all) passed silently and were not even counted.
The guard is rebuilt operand-shape-independent per this AC's own "positive, not negative" mandate:
it asserts, structurally, that no function in `internal/pdf` other than `flipY` itself contains a
subtraction whose left-associative chain bottoms out at a page-height-shaped operand — regardless of
variable naming or whether the value ever reaches `appendLength`/`writeLength`. RP-7 now red-proves
all three bypass shapes against a mutated copy of the real `imagedoc.go`, not a synthetic fixture
shaped to fit the scanner. The vacuity witness (D-000.9) is also sharpened: it requires a genuine
`flipY(...)` call to be present in **both** `textdoc.go` and `imagedoc.go` by name, not merely a
non-zero global counter (Finding 13).

**AC17b — name what the finding is, in the code comment on the helper** *(illustrative, D-1.8.10)*.
The invariant was **never enforced, only unexercised**. Quoted:

> one content-stream builder cannot violate "exactly one" … **Story 1.8 is the first real test of
> AD-24's clause.** Worth saying in the story, because **"it was always satisfied" reads as evidence
> it was guarded when it is evidence it never could have failed.**

This is the same *holds-by-construction-not-by-logic* shape as the vacuous re-wrap (AC3a) and the
baseline-green compressor guard (AC10) — **three instances in one story**, which is why AC30's
question 2 is the one to lean on here.

**AC18 — the image is drawn with `q`/`cm`/`Do`/`Q`, operands routed by D-1.1.b** *(illustrative for
the operator spelling; binding for the routing)*. `cm` operands are geometric → `appendLength`
(D-1.1.b's committed table names *"`Tm`/`Td`/`cm` operands"* explicitly). The `/XObject` sub-dict
joins `/Font` in the page's `/Resources`. **The resource name for an asset cannot collide** — see
AC18a.

**AC18a — the font path's resource-name collision hazard cannot arise for assets, and that is
asserted rather than assumed** *(illustrative)*. M-13: the existing escaper drops every character
outside `[A-Za-z0-9_-]`, and Story 1.5's QA Finding 23 added a collision check because two distinct
face names can escape to the same resource name. **Asset keys are 64 lowercase hex characters — a
strict subset of the safe set — so escaping is the identity and distinct keys stay distinct.** A
test asserts the escaper is the identity on a hex key. This is cheap and it stops a later reader
re-deriving the collision question.

---

### E — No new fetch guard; the fence is structural (D-1.8.5 · AD-1 · FR33 · source AC3)

**AC19 — add NO new guard** *(binding, D-1.8.5)*. Quoted in full, because the reason is the ruling:

> **The reason is structural, not "the guards are probably enough".** **Verified:
> `folio-format.md:224` — *"Images are only ever embedded. Folio never fetches by URL and never
> reads from disk at render time (FR33)."*** **There is no field in the format that can name a
> remote or on-disk resource.** An element's `asset` field is a **key into the in-memory `assets`
> map** — not a path, not a URL. At render time the bytes are already a `[]byte` on the parsed
> `Document`. **There is no code path that could fetch, because there is no value that names
> anything to fetch.** D-1.4.7's key-inventory drift test already keeps the field set honest.

**Independently re-verified during story creation:** `folio-format.md:224–225` carries that sentence
verbatim, and FR33 in the PRD (`prd.md:276`) reads *"Embed images inside the `.folio` template
itself. Folio does **not** fetch images by URL and does **not** read them from the filesystem at
render time; both would break determinism and introduce a mid-render failure mode."* The element
model confirms the mechanism: `Element.Asset` is a `Presence[string]`, and the string is a map key
(M-2).

**AC20 — one cheap behavioural check** *(illustrative, D-1.8.5)*: *"mutate the in-memory asset bytes
and assert the rendered PDF changes."* D-1.8.5: *"If the renderer read from disk, a cache, or
anywhere but the document, it would not. Three lines, non-vacuous, and it **proves the provenance
claim rather than restating it**."* The assertion is on the **rendered bytes differing**, not on a
field having been read.

**AC21 — Story 1.7's residual gap is carried forward UNCHANGED** *(binding, D-1.8.5)*: *"Story
1.7's recorded residual gap carried forward as-is — a capability fence, not a proof. **Do not
quietly upgrade its description in this story.**"* D-1.7.3's own words, which are the ones that must
survive intact: *"a file-scoped import rule makes the *obvious* violation impossible but does not
stop a deliberate cross-file route … **That is accepted.** The AC's intent is a **capability** claim
… not a security boundary."* The existing `net` ban under `internal/` and the AST-located
render-file fence **stand unchanged**. A reviewer should check that no comment or README line in
this story's diff describes that fence as a proof, a guarantee, or a security boundary.

---

### F — A third matrix document, guarded on every leg (D-1.8.6 · D-000.4 · AD-21, AD-22)

**AC22 — the image document JOINS the other two; it replaces neither** *(binding, D-1.8.6)*.
Quoted:

> `minimal-rect` is the fontless emission baseline whose four-target verification was deliberately
> preserved at Story 1.5; `font-text` covers subsetting. **Replacing either would silently retire
> coverage of a feature that is still shipping.**

After this story the matrix document list holds three entries. **Do not narrate that as a count
anywhere** (D-000.14 extended); the harness already derives it, and `wantMatrixLegs` is the
deliberately-independent constant for **targets**, not documents (M-14) — do not repurpose it.

**AC23 — an image-XObject guard on EVERY captured stream, BEFORE any comparison** *(binding,
D-1.8.6)*. Quoted:

> assert each captured stream contains an **image XObject** (`/Subtype /Image`) **before** any
> comparison — **on every leg, not once**. Without it the matrix would compare four identical
> **imageless** PDFs and report byte-identity across all targets, which is exactly what 1.5
> discovered about fonts.

Mechanically this mirrors the shipped `FontFile2` guard exactly (M-14): a per-document boolean on
the document descriptor, checked inside the per-target loop **before the hash is computed**, in
**both** the local all-targets test and the per-target CI test, **and** in the untagged golden
fixture test. **Each existing document keeps its own feature guard on the same basis** — the
`FontFile2` guard is not to be generalised away while adding the third.

**AC23a — the guard is a D-000.9 instance and is red-proofed as one.** RP-8: capture a render with
the image element removed; the guard must redden naming the document and the missing
`/Subtype /Image`. *"What would this check have printed if it had been unable to run at all?"* — the
answer must not be "the same thing".

**AC24 — the full four-target matrix RUNS IN THIS STORY, over all three documents** *(binding,
D-000.4)*. D-000.4 names this story explicitly: *"a story whose *own* deliverable is hash-shaped
runs the full matrix in its own story even under the per-epic cadence. That is, at minimum: **1.2**
…, **1.5** …, **1.8** (image embedding) …"*. `linux/amd64` and `linux/arm64` under Docker, `js/wasm`
under Node, against local `darwin/arm64`. **The override is logged with its reason in the Delivery
Log.** Silence is not green — if a leg does not run, the Delivery Log says so by name.

**AC25 — the new golden fixture and the CI wiring** *(illustrative in mechanism, binding in
completeness)*. A fixture directory with `expected.json` (`folioGoVersion`, `goToolchain`,
64-lowercase-hex `sha256`), `expected.pdf`, `input.folio` and `README.md`; `goToolchain` **must
equal the other fixtures'** or the shared-toolchain assertion fatals. The matrix workflow's four
render jobs each upload the new document's hash artifact and the compare job's document list gains
the new slug — **the compare job names files explicitly and never globs**, and asserts each is
64-hex and that every document has exactly four. **Wiring the fixture without wiring the workflow
is the D-000.9 shape here: CI would compare fewer documents than exist and say nothing.**

**AC25a — `input.folio` is byte-identical to the Go constant that renders it.** The existing
`font-text` fixture carries this obligation from its own QA review; the new fixture inherits it.

---

### G — Amendments the rulings force (D-1.8.7, D-1.8.8 · D-000.6 · D-1.4.7, D-1.4.10 · AD-21)

> **D-000.6:** *a ruling that makes a canonical document wrong amends that document, in the story's
> own commit.*

**This section is not optional cleanup — without it the story cannot go green.** Every finding below
was measured during story creation; see M-5, M-6, M-7.

**AC26 — the three existing asset-bearing template fixtures are rebuilt on real assets** *(binding
via D-1.8.1 + D-1.8.3)*. Measured, all three fail the new validation:

| Fixture | Key as written | Decoded `data` | Fails |
|---|---|---|---|
| `unknownKeysFixture` | 64 × `b` | `Zg==` → the single byte `f` | AC6 (key ≠ digest `252f10c8…`); AC9 (not an image at all, declared `image/png`) |
| `maximalFixture` | 64 × `a` | 45 bytes — a **truncated** PNG (the first 60 base64 chars of a 70-byte PNG) | AC6 (key ≠ digest `c1e372ae…`); AC9 (malformed PNG) |
| `passthroughLevelBase` | **62** × `c` | `AA==` → one `0x00` byte | AC6a (not even 64 characters); AC6; AC9 |

`maximalFixture` additionally contains a real `{"type":"image","asset":"aaaa…"}` element, so it is
the fixture that most needs to become genuinely renderable. Each is rebuilt on a real, supported
image with its **real** SHA-256 key, and the P1/P2/P3 corpus is re-run. **Do not weaken the
validation to accommodate a fixture** — that inverts the ruling.

**AC26a — regenerating must NOT change what those fixtures test** *(binding, D-1.8.8)*. Quoted:

> They exist for other properties (passthrough, unknown keys); the key fix is **incidental to their
> purpose and must not silently alter it.**

Concretely: `unknownKeysFixture` must still carry `futureAssetEntryKey` inside the asset entry and
still exercise unknown-key passthrough at that nesting level; `passthroughLevelBase` must still be
the injection base its level-parameterised test walks, with the same injection points intact; and
`maximalFixture` must still exercise the widest field set. **A reviewer should diff each fixture's
*non-asset* content and find it unchanged.** The temptation to "simplify while I'm in here" is what
this AC exists against — and note that D-1.8.8 records that these keys survived **four stories**
unnoticed, so these fixtures are load-bearing for properties nobody is currently re-deriving.

**AC27 — `folio-format.md`'s `assets` example is amended** *(binding, D-000.6; the doc is a
readability contract under D-1.4.7, and one that lies is worse than none)*. Measured, the example at
`folio-format.md:207–217` is wrong in **three** independent ways under these rulings:

1. **Its key is an ellipsis placeholder** (`"9f2b…c41d"`), so it cannot satisfy AC6.
2. **Its wrapping is not canonical.** Its two elements are **60** and **36** characters (96 total).
   The canonical 76-column split of the same payload is **76 + 20**.
3. **Its PNG has an alpha channel — the format specification's own example of an embedded asset is
   a document D-1.8.1 refuses.** Decoded and walked chunk-by-chunk during story creation: 70 bytes,
   `IHDR` reports `W=1 H=1 bitdepth=8 **colortype=6** compression=0 filter=0 interlace=0`.
   Colour type 6 is truecolour **with alpha**.

Finding (3) is the one to read twice. It is not a documentation nit: it means the single worked
example a hand-editor would copy is, under this story's rulings, unloadable — and nothing in the
tree would have told anyone.

The example is replaced with a supported image, canonically wrapped, under its real key. **A
verified candidate**, generated and then independently re-walked during story creation — a 75-byte
2×2 8-bit **RGB, colour type 2**, non-interlaced, no-alpha PNG (all three chunk CRCs verified,
base64 round-trip verified):

- `sha256` = `3d27b4ed2fdfdb12b533f2ddf6e113f5f6ad516b1acd9ebb3ed1de5476ec51c6`
- base64 length 100 → canonical split **76 + 24**
- elements: `iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAIAAAD91JpzAAAAEklEQVR42mP4z8DAAMIM/4EAAB/u`
  and `Bfvxq7p3AAAAAElFTkSuQmCC`

**AC27a — how the replacement is produced** *(binding, D-1.8.7)*. **Not by hand, and not from this
file's numbers.** Quoted:

> **generated by the shipped serializer** and **independently cross-checked by a second tool** (CRCs,
> colour type, digest), per D-1.4.10's discipline.

So: build the document in Go, serialize it with the shipped serializer, paste **that** output into
the doc, then verify the digest, the colour type and the chunk CRCs with a **different** tool than
the one that produced them. The creator's candidate above was deliberately written as *"a candidate
to re-derive, not as truth"*, and D-1.8.7 honours that instinct explicitly — **amending a canonical
document from a value an upstream artifact asserts would ratify a guess.** Use a non-square image
for the fit/centre fixtures: a 2×2 is square and cannot distinguish AC13's two binding-axis
branches.

**AC27b — close it permanently: extend the fence/golden byte-identity assertion to the asset
example** *(binding, D-1.8.7)*. The mechanism already exists and must be **extended, not
duplicated** — after this, the asset example cannot rot again.

⚠️ **Measured correction to a citation inside D-1.8.7.** The ruling attributes that assertion to
**D-1.4.11**. **D-1.4.11 is `ScanAbsences`' coverage witness** — the wrong ruling. The mechanism is
`TestWorkedExampleMatchesGoldenFixture` in `internal/template/goldenfixture_test.go`, whose own
comment records it as **AC16 step 3**, i.e. **D-1.4.10**'s (*"Amend the worked example so it is
genuinely canonical; cross-check it against an independent printer"*). **The mechanism is real and
the obligation stands exactly as ruled; only the ID is misattributed.** Recorded per the D-1.7.1
(amended) precedent — a narrated reference drifting from the artifact it names — so the developer
does not go looking in `lint`. **Not raised as `DECISION NEEDED`: this is a verifiable
misattribution inside one ruling, not a conflict between two** (D-1.2.6).

**Two implementation facts the extension turns on, both measured:**

1. `extractWorkedExampleFence` finds the **last** ```` ```json ```` fence **under the
   `## Worked example` heading**. The asset example is a **different, earlier** fence under
   `## assets`. So this needs a **second extractor keyed to its own heading** — not an edit to the
   existing one, which would silently re-point the worked-example assertion.
2. The asset example is a **fragment**, not a whole document, so it cannot be a `.folio` the parser
   round-trips. Either compare it to a golden fragment emitted by the shipped assets writer, or
   embed the same asset in a full canonical golden and extract the sub-object. **Pick one and say
   which in the Delivery Log** — both satisfy the ruling; silently doing neither does not.

Both existing extraction helpers must keep a **failure path that is distinguishable from a match**
(D-000.9): "heading not found" and "fence never closed" already `t.Fatal` rather than returning an
empty string that would compare equal to an empty golden. Preserve that in the new extractor.

**AC27c — the pattern, named because this is the second instance** *(illustrative, D-1.8.7)*.
Quoted:

> `folio-format.md`'s examples were **hand-authored and never executed**. DN-1 found the worked
> example was **not a fixed point**; now the asset example is invalid **three ways**. **Every fenced
> block claiming to be complete or canonical should be executed, not read.** Fragments can at least
> have their field names covered by D-1.4.7's drift test.

The Delivery Log records which fenced blocks in `folio-format.md` are now **executed** (byte-identity
asserted against a golden) versus merely **covered** (field names only, via the drift test). That
inventory is the artifact a later story needs; **do not extend execution to every fragment in this
story** — the ruling names the asset example, and the general principle is recorded, not discharged.

**AC28 — the before/after of every amendment is quoted verbatim in the Delivery Log** *(binding,
D-000.6's established shape — see the Story 1.4 and Story 1.2 amendment entries)*.

**AC29 — the key-inventory drift test still passes, and still compares a non-zero number of keys**
*(binding, D-1.4.7)*. D-1.4.7: *"This test is itself exposed to D-000.9 — a key-coverage test that
compares zero keys passes. It must report **how many keys it compared**, and zero compared must
fail."* Amending the format document is exactly when that test earns its keep; confirm the witness
is live **before** relying on it.

---

### Standing rules

**AC30 — D-000.9, its extension, and D-000.13: three questions of every guard and every red-proof**
*(binding)*:

1. *What would this check have printed if it had been unable to run at all?* (D-000.9)
2. *What would this red-proof have printed if the mutation had never been applied?* (D-000.9 extended)
3. ***Did it fail for the reason it names?*** (D-000.13)

**A red-proof asserts on the rule id and the finding message, never on exit status or mere failure.
A control mutation must be VALID SYNTAX with FORBIDDEN SEMANTICS** *(binding, D-000.13)*. Every
red-proof in the register records the **control result** alongside the mutation result.

This story has an unusually high density of question-2 exposure: **three of its ACs are green at
baseline for the wrong reason** — AC3's canonical-in/canonical-out half (the serializer echoes,
M-4), AC10's compressor guard (nothing imports a compressor yet, M-8), and AC17's flip-site guard
(there is only one content-stream builder, M-12). Each is stated with the mutation that
distinguishes it.

**AC31 — D-000.11.** Every gate runs `-count=1`. **A cached `ok` is not evidence.** This story
touches `folio-format.md` and repo-root fixtures — both **outside the module** — which is precisely
the case D-000.11 was ruled on.

**AC32 — D-000.12 and D-000.14: measurement discipline.** **Never verify bytes, hashes or panic
stacks through a shell pipe** (D-000.12) — use `rtk proxy` writing to a file. **When a count is
load-bearing, count by AST** (D-000.14). **A total is computed or omitted, never narrated beside the
artifact it summarises** (D-000.14 extended). This story's own measurements were taken that way; the
decision-log state in the banner was read from `git diff --unified=0` hunk headers and a `wc -l` of
`git show` output, not from a filtered grep count.

**AC33 — D-1.3.1 / AD-1 / D-1.3.5.** Any new or modified non-test file under `folio-go/` inherits
AD-1's full ban (`time`, `os`, `math/rand`, `net`, and `math` transcendentals by selector) and
D-1.3.1's exact `_test.go` exemption (`os`, `testing`, `path/filepath`, `embed` — **and nothing
else**). **Ranging a map is banned anywhere under `internal/`** (D-1.3.5); the escape hatch idiom is
`slices.Sorted(maps.Keys(m))`, which the assets writer already uses. Additionally (M-15): the
no-float guard bans the identifiers `float64`/`float32` **and any float literal** anywhere under
`folio-go/`, `_test.go` included — so image geometry, header parsing and every test expectation stay
integer.

**AC34 — D-000.10.** The developer dispositions **every ruling ID this story enumerates**, one line
each, in the table at
[§ Ruling disposition table](#ruling-disposition-table-d-00010); the reviewer mirrors it
independently. **The enumeration is explicit and complete in that table.** Per **D-000.14
(extended)** no total is written beside it — if a count is wanted, compute it.

---

## Decisions resolved

### D-1.8.1 (amended) — an unrecognised `mediaType` is never a load error; `Validate` predicts `Render`

**This story's creator raised the conflict as `DECISION NEEDED` per D-1.2.6 and did not arbitrate
it. It was ruled while this file was being written.** The obligations are ACs — **AC11a**, **AC11b**,
**AC11c**, and the amended **AC9** table. This section records *why*, because the reasoning is what
transfers.

**The conflict as raised.** D-1.8.1 made any asset of an unrenderable media type a **load error**.
Owner decision **D-1.4.9** guarantees a higher-MINOR file **loads** — *"Nothing is dropped, nothing
is refused."* A future Folio `1.1` adding WebP produces a file that today's `1.0` library refuses,
which is exactly the outcome **D-1.4.12** records as making forward compatibility *"void for exactly
the additions most likely to be made."* `mediaType` is a **known key, open set, unrecognised
value** — a category **neither prior ruling addresses**, created for the first time by D-1.8.1.

**The ruling, and the fact that it is not a narrowing.** Quoted:

> **This does NOT narrow owner decision D-1.4.9 — it removes a refusal the lead wrongly introduced
> against it.** No owner escalation needed … the orchestrator's backstop caught a lead ruling
> contradicting an owner decision, and the fix was to stop contradicting it.

**The disqualifying test, which is the part worth carrying forward** *(binding)*:

> Under D-1.8.1 as written, **the same bytes are valid or invalid depending on which library reads
> them.** … `folio-format.md` opens by calling the format *"a public contract, not an implementation
> detail"*, and **a public contract whose validity depends on the reader is not a contract.**
> "Valid `.folio`" must mean something stateable without naming a library version.

**What settled it was the creator's second face**, quoted: *"D-1.8.3 preserves orphaned assets
**because P1 forces it** — so a valid document may carry an asset **nothing will ever draw**.
Refusing the whole file over a rendering never attempted is indefensible."*

**The generalisation outranks the case** *(binding)*: **`Validate` is a dry-run predictor of
`Render`, not a second rule system.** Story 1.4's loud-failure posture survives intact — a *drawn*
unrenderable asset still fails in CI before production — **without** making validity
reader-dependent. **This will matter at Story 3.7**, which is why AC11c writes it down here rather
than leaving 3.7 to re-derive it.

**And what it does not relax** *(binding)*: a **recognised** `mediaType` whose bytes are not that
format is still a load error, because *"the file lies about itself and every library agrees it is
malformed."* AC11b puts the two cases side by side; separated, they read as an inconsistency.

### D-1.8.7 – D-1.8.11 — the creator's four other findings, confirmed and ruled

Each was a defect at baseline that no ruling had reached. They are not background: **three of them
change ACs in this file.**

| Ruling | Finding | Where it lands |
|---|---|---|
| **D-1.8.7** | The format doc's asset example is invalid three ways — alpha, non-canonical wrapping, ellipsis key. The examples *"were hand-authored and never executed"* | AC27, AC27a, AC27b, AC27c |
| **D-1.8.8** | Three shipped fixtures carry invalid asset keys (64×`a`, 64×`b`, one of **62 characters**) — *"nothing noticed, across four stories"* | AC6a, AC26, AC26a |
| **D-1.8.9** | The serializer echoes the input wrapping, so D-1.8.2's canonical half is vacuous at baseline | AC3, AC3a |
| **D-1.8.10** | AD-24's single-flip clause was *"never **enforced**, only **unexercised**"*; make the guard positive | AC17, AC17a, AC17b |
| **D-1.8.11** | The licence manifest's font-extension allowlist is the rotting-list pattern | **M-11 — recorded for the Epic 1 boundary gate, NOT built here** |

**D-1.8.11 is deliberately not an AC.** It is recorded, with its ruled *shape*, and handed to the
Epic 1 boundary gate. Do not expand the manifest scanner in this story.
---

## Tasks / Subtasks

- [x] **Task 1 — Read the working-tree decision log first.** Confirm the banner's measurements
      still hold (`git status --porcelain`; the hunk header). The rulings are **not** in `54a0777`.
- [x] **Task 2 — Base64 canonical form** (AC1–AC4). Serializer **re-wraps from decoded bytes** —
      never echoes (D-1.8.9); parser accepts any wrapping; strict decode; empty-decoded and
      invalid-base64 load errors. **The 64→76 fixture is the only discriminating test.**
- [x] **Task 3 — Key validation, orphans, dedup** (AC5–AC8a). **Shape check then value check, two
      error classes** (D-1.8.8); orphan-preserving retained fixture; the Epics 5/6 corollary comment.
- [x] **Task 4 — Image decoding by chunk/segment walking, no decoder library** (AC9, AC15, AC15a,
      AC15b). PNG: walk `IHDR`, reject alpha/palette/16-bit/interlaced, concatenate `IDAT`. JPEG:
      locate the frame header for `W`/`H`. Validate `W`/`H` at the constructor; unexported fields.
- [x] **Task 5 — The no-compressor guard** (AC10, AC10a). AST import guard with a coverage witness;
      the R4 comment at the filter site.
- [x] **Task 6 — `mediaType` stays open; the capability error lives on the RENDER path** (AC11,
      AC11a–AC11c). Absent from the closed-set inventory, asserted. **One named predicate, one call
      site, called from render — never from the loader.** Ship the recognised-but-unparseable load
      error (AC11b) and the paired test. **Do not create `folio.Validate`** — record AC11c's rule
      for Story 3.7 and stop.
- [x] **Task 7 — Fit, centre, and the single flip site** (AC13–AC18a). Cross-multiplication;
      `ScaleRound(bw-dw, 1, 2)`; **extract the Y-flip helper before adding the image path**, and
      make its guard **positive** (AC17a) — not a search for minus signs.
- [x] **Task 8 — PDF emission** (AC7, AC18). Image XObject stream with `/Filter` and
      `/DecodeParms`; `/XObject` sub-dict in `/Resources`; `q`/`cm`/`Do`/`Q`; `/Width`,`/Height`
      through `appendInt`.
- [x] **Task 9 — Wire image elements into the render path.** They are parsed and then dropped
      today (M-2).
- [x] **Task 10 — The no-fetch fence** (AC19–AC21). **Add no new guard.** Add the mutate-the-bytes
      behavioural check. **Do not upgrade Story 1.7's residual-gap description.**
- [x] **Task 11 — Amendments** (AC26–AC29). Rebuild the three fixtures on real assets **without
      changing what they test** (AC26a); amend `folio-format.md`'s asset example from **shipped
      serializer output, cross-checked by a second tool** (AC27a); **extend the fence/golden
      byte-identity assertion with a second extractor keyed to the `## assets` heading** (AC27b —
      and read its citation-correction note before hunting for the mechanism); re-run P1/P2/P3;
      confirm the drift witness is live.
- [x] **Task 12 — The third matrix document** (AC22–AC25a). Descriptor, capture func + env
      selector + `TestMain` branch, fixture directory, untagged golden test, **and the workflow
      wiring in all four render jobs plus the compare job's document list.**
- [x] **Task 13 — Run the full four-target matrix** (AC24). Record every leg by name in the
      Delivery Log, with the D-000.4 override reason.
- [x] **Task 14 — Red-proof register**: run every row, record mutation **and** control.
- [x] **Task 15 — Disposition table** (AC34): one line per ruling id, complete. **No narrated
      total.**

**Not in this story, and not to be added:** no commit, no branch, no `done`. The status this story
sets is `ready-for-dev`.

---

## Dev Notes

Everything below was **measured during story creation** at `54a0777`, not inherited.

### M-1 — the assets SCHEMA already shipped at Story 1.4; this story adds meaning, not structure

`internal/template/model.go` already declares `Assets map[string]Asset` on `Document` (with the
comment *"Assets is keyed by lowercase hex SHA-256 of the raw bytes"*) and:

```go
type Asset struct {
	Data      []string
	MediaType string
	Extra     []Field
}
```

`parse.go`'s `decodeAssets` already validates the **shape** — `data` is an array of strings,
`mediaType` is a string, unknown keys go to `Extra` — and sets `doc.Assets = map[string]Asset{}`
when the key is absent. **Nothing decodes the base64, nothing recomputes the key, nothing looks at
`mediaType`'s value.** So D-1.8.2 and D-1.8.3 are additions to a live seam, not new plumbing, and
the existing `assets.<key>.data` field-path convention in the error messages is the one to extend.

### M-2 — image elements are parsed, round-tripped, and then silently dropped

`Element` already carries `Asset Presence[string]`, and `parse_bands.go` already makes a missing
`asset` on an image element a located load error. But `render.go`'s `collectTextRuns` opens with
`if el.Type != template.ElementText { continue }` — **so an image element renders as nothing
today, with no error.** That is the behaviour source AC3 is about, and it is the baseline any
"image appears in the PDF" red-proof starts from.

Also measured: **nothing checks that an element's `asset` key exists in `Document.Assets`.** A
dangling reference is silently accepted today. D-1.8.3 does not rule on it; AC8's orphan direction
(asset with no element) is the ruled one. The reverse direction (element with no asset) is worth a
located error, but it is **not** ruled — treat adding it as an improvement, and say so in the
disposition table rather than leaving a reviewer to guess whether it was mandated.

### M-3 — the located-error constructor

`internal/template/errors.go` is 31 lines and declares the only load-error shape in the package:

```go
func newLoadError(field, elementID, value, reason string) error
```

`ElementID` is empty when the error is not scoped to one element. The assets path already calls it
as `newLoadError("assets."+k+".data", "", string(dataRaw), "…")`. `internal/diag` does not exist
(Story 3.6 owns it) and a lint tripwire fails the build the day it is created — **do not mint a
diagnostic code in this story.** Note that `internal/fontset` uses a different shape
(`fmt.Errorf("fontset: font %q: …")`); its located-ness is textual. Follow whichever package the
code lives in.

### M-4 — the serializer ECHOES the wrapping, so AC3's canonical half is vacuous at baseline

`serialize.go`:

```go
func writeAssets(dst []byte, depth int, assets map[string]Asset) []byte {
	…
	{"data", func(dst []byte, depth int) []byte { return writeStringArray(dst, depth, a.Data) }},
```

`a.Data` is written out exactly as it came in. **There is no re-wrapping anywhere.** So today a file
hand-wrapped at 64 columns round-trips at 64 and satisfies P1/P2/P3 — the serializer is precisely
the *"serializer that merely echoed its input"* D-1.8.2 warns the canonical-in/canonical-out test
would pass on. **D-1.8.2's warning is a description of the shipped code, not a hypothetical.** The
hand-wrapped variant is the only non-vacuous half.

Consequence for AC26: introducing 76-column normalisation **changes P3's fixed point** for every
existing fixture whose asset `data` is not already at 76 columns. Measured, none of them is.

### M-5 — all three asset-bearing template fixtures fail the new validation

Digests computed during story creation:

| Fixture | Key as written | Key length | Decoded bytes | True SHA-256 of decoded |
|---|---|---|---|---|
| `unknownKeysFixture` (`fixtures_test.go:22`) | 64 × `b` | 64 | `f` (1 byte, from `Zg==`) | `252f10c83610ebca1a059c0bae8255eba2f95be4d1d7bcfa89d7248a82d9f111` |
| `maximalFixture` (`fixtures_test.go:215`) | 64 × `a` | 64 | 45 bytes, truncated PNG | `c1e372aea0f52b681125d13b4082b349944f962512ad0920896edd0af3f68001` |
| `passthroughLevelBase` (`passthrough_test.go:34`) | 62 × `c` | **62** | one `0x00` byte (from `AA==`) | `6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d` |

The 62-character key is worth noting on its own: it is not merely the wrong digest, it is not a
digest-shaped string at all, and nothing in the tree noticed. `maximalFixture`'s payload is the
**first 60 base64 characters** of the 70-byte PNG in `folio-format.md` — a truncated file that
decodes to 45 bytes and is not a valid PNG.

### M-6 — `folio-format.md`'s example PNG HAS AN ALPHA CHANNEL

The example at `folio-format.md:207–217` decodes to a valid 70-byte PNG (signature verified). Walked
chunk by chunk:

```
chunk IHDR len 13
  W 1  H 1  bitdepth 8  colortype 6  compression 0  filter 0  interlace 0
chunk IDAT len 13
  IDAT zlib bytes: 78da636460f85f0f0002870180
chunk IEND len 0
```

**Colour type 6 is truecolour with alpha.** Under D-1.8.1 the format specification's own example of
an embedded asset is a document this library refuses. This is the sharpest single reason Section G
is not optional: the one example a hand-editor would copy is, after this story, unloadable.

### M-7 — the same example is also non-canonical under D-1.8.2

Its two `data` elements are **60** and **36** characters (96 total). The canonical 76-column split of
the identical payload is **76 + 20**. So the example violates AC1 independently of M-6. Its key is
the literal ellipsis `"9f2b…c41d"`; the real digest of the decoded 70 bytes is
`497790947d4666760ce38f3c00e852c71fdb66cae849bae8e9ede352719e1581` — recorded only to show the
placeholder is a placeholder, since M-6 means those bytes are not the ones to keep.

### M-8 — `/Filter` and every image primitive arrive here for the first time

Measured across `folio-go/**/*.go`: **zero** occurrences of `/XObject`, `/Subtype /Image`,
`/DCTDecode`, `/FlateDecode`, `/ColorSpace`, `/BitsPerComponent`, or the operators `cm`, `q`, `Q`,
`Do`. Shipped content-stream operators are `re`, `f`, `BT`/`ET`, `Tf`, `Tm`, `Tj` only. The
`FontFile2` stream is emitted with **no `/Filter` at all**.

Import status, exhaustively:

| Package | Imported anywhere under `folio-go/`? |
|---|---|
| `compress/flate`, `compress/zlib`, `compress/gzip` | **No** — `compress/flate` appears only in a `go.mod` comment |
| `image`, `image/png`, `image/jpeg` | **No** — `"image/png"` occurs only as a media-type string literal |
| `encoding/base64` | **No** — `Asset.Data` is never decoded |
| `crypto/sha256`, `encoding/hex` | **Yes**, three sites (`internal/pdf/document.go`, `matrix_test.go`, `fixture_test.go`) |

There is **no lint ban** on any of `compress/*`, `image/*`, `crypto/sha256` or `encoding/base64`
today. The banned set is exactly `time`, `os`, `math/rand`, `net`. **So AC10's guard is new, is
green at baseline, and its red-proof must be a real import in a real file** — that is question 2 of
AC30, and the reason RP-5 is written as a construction rather than an assertion.

The module-graph allowlist is an **exact set** (`folio-go` plus `textshape`) asserted from
`go list -m all`, so **any third-party image decoder is an immediate red test.** Stdlib is invisible
to it — which is exactly why AC10's guard is needed on top.

### M-9 — `ScaleRound`'s exact contract, and why AC15's bound is arithmetic rather than taste

```go
func ScaleRound(v Length, num, den int64) Length   // computes v*num/den, round-half-to-even
```

`num` is the numerator, `den` the denominator, so D-1.8.4's `ScaleRound(bw, H, W)` reads as written.
Four panic conditions, in order: `den == 0`; `den == math.MinInt64`; `v*num` overflows `int64`;
`v*num/den` overflows (`den == -1` with `v*num == math.MinInt64`). All four are documented as
programmer-error guards.

Two consequences the ACs turn on:

- `W == 0` from a corrupt header reaches the **first** panic. AC15 is what stands between a
  malformed file and a crash — this is D-1.5.2's situation exactly, one story later.
- `bw*H` reaching the **third** panic is why AC15 says *bounded* and not just *positive*. A page box
  is under `10^6` millipoints; a pixel bound of `10^6` leaves `bw*H` under `10^12`.

Also worth knowing: the centring call `ScaleRound(v, 1, 2)` is precisely the argument shape of
Story 1.5's QA Blocker 3, where an earlier implementation returned `+4611686018427387904` for
`ScaleRound(math.MinInt64, 1, 2)`. That defect is **fixed** in the shipped code, and `bw-dw` is
non-negative by construction, so it is unreachable here — noted so nobody re-opens it.

An AST guard already asserts geom's exported top-level function set is **exactly `{ScaleRound}`**
(D-1.5.2's *"the fix is layering, not a second door"*), with a `filesScanned == 0` vacuity guard. A
second scaling helper reddens it.

### M-10 — the D-1.5.2 seam, in code, is the pattern AC15/AC15a copy

`internal/fontset` validates at the constructor before any value can reach `ScaleRound`:

```go
const ( MinUnitsPerEm = 16; MaxUnitsPerEm = 16384 )

type Font struct {
	unitsPerEm uint16 // validated: MinUnitsPerEm <= unitsPerEm <= MaxUnitsPerEm
	…
}
```

and its **second half matters here too**: the raw value is read from the head table directly,
because the vendor accessor *"silently substitutes 1000 when the parsed head reports 0 … which would
make unitsPerEm == 0 unobservable"*. The image analogue: read `W`/`H` from `IHDR` (or the JPEG frame
header) directly; do not route them through anything that could normalise a zero away, and do not
expose an unvalidated field.

### M-11 — the licence manifest scans FONT extensions only, so a committed image is invisible to it

`lint/internal/manifest/manifest.go` declares `var fontExtensions = []string{".ttf", ".otf",
".ttc"}`, and the "Redistributed non-code assets" table is built from that scan. **A committed
`.png` or `.jpg` is not examined**, so AD-26's *"Redistributed non-code assets keep their own terms"*
is unenforced for images. That is a D-000.9 shape — the guard's "all clear" and its "I did not look
at images" are the same output.

**Recommended, and it also sidesteps the question entirely:** construct every test image as **byte
literals in Go source** (they are 75 bytes and 45 bytes at this size), so nothing binary is
redistributed and no manifest entry is needed. The matrix fixture's `expected.pdf` is a committed
binary, but so are the two existing fixtures' and they carry no manifest row — consistent with the
extension list. **Record the gap in the Epic 1 boundary report alongside AC12's alpha limitation;
do not expand the manifest scanner in this story.**

**Ruled as D-1.8.11, and the ruled shape matters more than the gap** *(binding, for whoever fixes
it)*:

> **When fixed, do NOT add `.png`/`.jpg` to the list** — an extension allowlist is the rotting-list
> pattern and the next asset type is invisible again. **Invert it: every committed file in the
> declared asset and fixture locations must be accounted for in the manifest, and anything uncovered
> fails.** That is the **coverage-witness form** — ask *"is every asset accounted for?"*, not *"is
> this file a font?"* — and it is **the same fix that closed `ScanAbsences`**.

So the obvious two-line fix is **explicitly forbidden**. Anyone reaching for it in this story should
stop: the correct fix is a different shape and it belongs to the **Epic 1 boundary gate**. Recording
the ruled shape here is what stops the cheap version landing by reflex later.

Related, and unblocking: `lint`'s absence tripwires cover `folio-designer/`, `folio-go/fonts/`,
`folio-go/internal/expr/`, `folio-go/internal/diag/` and `folio-go/cmd/folio`. **There is no
tripwire on an image or asset package**, so creating one is unblocked. And `ScanEmbedFont` bans
`//go:embed` of font extensions under `internal/` — **there is no image analogue**, so a
`//go:embed foo.png` would pass today. Prefer byte literals anyway.

### M-12 — there is exactly ONE coordinate flip today, and only by accident

`internal/pdf/textdoc.go`, in `buildTextContentStream`:

```go
pdfX := page.MarginLeft + run.X
pdfY := page.Height - page.MarginTop - run.Y - run.FontSize
```

That is the only place in the module that inverts a coordinate — **not because it was extracted as
AD-24 requires, but because there has only ever been one content-stream builder.** AD-24's *"exactly
one function … and nowhere else in the module inverts a coordinate"* is currently satisfied
incidentally.

An image placement needs the same flip, using the **drawn** height rather than the box height, since
the image is centred inside the box. Open-coding it is a second inversion site and a direct AD-24
violation. **None of D-1.8.1–D-1.8.6 mentioned the flip**, so a developer following only the
first-round rulings would not have been warned.

**Confirmed as D-1.8.10**, which endorsed the framing: *"Applying a binding spine invariant that no
ruling contradicts is not arbitration — the creator was right not to treat it as such."* It is an
AC, not a `DECISION NEEDED`.

**The finding is not "one flip site exists" — it is that the clause was never enforced, only
unexercised.** One content-stream builder *cannot* violate "exactly one". D-1.8.10: *"'it was always
satisfied' reads as evidence it was guarded when it is evidence it never could have failed."*
**Story 1.8 is the first real test of AD-24's clause.** And per AC17a the guard must be **positive**
(all coordinate emission routes through the one helper), never *"nobody else writes a minus
sign"* — the negative form is unfalsifiable and produces false positives.

### M-13 — the font path's resource-name collision cannot arise for assets

`pdfNameEscape` keeps `[A-Za-z0-9_-]` and drops everything else, returning `"F"` if the result is
empty. Story 1.5's QA Finding 23 added a whole-document collision check because two distinct face
names can escape to the same resource name, raising `resourceNameCollisionError`. **Asset keys are
64 lowercase hex characters, a strict subset of the kept set**, so the escaper is the identity on
them and distinct keys cannot collide. AC18a asserts this rather than leaving it to be re-derived.

### M-14 — the matrix harness, and exactly what a third document touches

`matrix_test.go` (build tag `matrix`) holds:

```go
var matrixDocuments = []matrixDocument{
	{label: "minimal-rect (fontless)", slug: "minimal-rect", capture: captureRender,
	 fixtureRelPath: []string{"fixtures","minimal-rect","expected.json"}, requireFontFile2: false},
	{label: "font-text (template+font)", slug: "font-text", capture: captureFontRender,
	 fixtureRelPath: []string{"fixtures","font-text","expected.json"}, requireFontFile2: true},
}
```

It is a `var`, not a `const`, deliberately — a red-proof test swaps it and restores it. The four
targets are `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`, with `wantMatrixLegs = 4`
declared **independently of `len(matrixTargets)`** so the count cannot self-agree. **That constant
counts targets, not documents — do not repurpose it for AC22.**

The `FontFile2` guard, which AC23 mirrors, is applied **inside the per-target loop, before the hash
is computed**, and is duplicated in `TestTargetRenderHash` and in the untagged golden test. Adding a
third document touches: the descriptor slice; a new capture func plus a new subprocess env-var
selector plus a `TestMain` dispatch branch; a fixture directory; an untagged golden test; **and the
matrix workflow — all four render jobs' upload paths and the compare job's document list.** The
compare job names files explicitly and never globs, and asserts four hashes per document.

### M-15 — the guards that will bite, measured

- **No-float guard** (`internal/arch_test.go`): bans the identifiers `float64`/`float32` **and any
  float literal token** anywhere under `folio-go/`, `_test.go` included (`testdata/` and dot-dirs
  skipped **by category**, not by name). Two production callers. Image geometry and every test
  expectation must be integer.
- **Number-formatting confinement** (`internal/pdf/emit_source_test.go`): no file under
  `internal/pdf` **except `numbers.go`** — tests included, no exemption — may call `strconv.Format*`,
  `strconv.Itoa`, `strconv.Append*` or any `fmt` formatting call. Matched by AST and import alias.
  Hex-shaped byte encodings use the existing `appendHex4` carve-out, whose D-1.1.b justification is
  quoted in-file.
- **Map-range ban** (`lint`): ranging a map is banned anywhere under `internal/`; use
  `slices.Sorted(maps.Keys(m))`. The assets writer and the face loop already do.
- **Render-file fence**: the file declaring `Render`/`RenderTo` is located **by AST, never by
  filename**, and may import none of `os`/`time`/`net`/`math/rand`.

### M-16 — FR33's number is reused in the PRD's review artifacts

`prd.md:276` is the images FR33 this story covers, and `folio-format.md:225` cites it. Some PRD
**review** artifacts (`review-rubric.md`, `review-adversarial.md`) use "FR33" for a *Design Canvas*
requirement under an older numbering. Those are review documents, not live contracts; the epic, the
spec and the PRD body agree. Noted only so a developer grepping `FR33` is not derailed.

---

## Ruling disposition table (D-000.10)

**One line per ruling id this story enumerates, complete.** The developer fills the disposition
column; the reviewer mirrors it **independently**. Per **D-000.14 (extended)** no total is narrated
here — compute one if it is wanted.

| Ruling | What it binds here | Developer disposition |
|---|---|---|
| **D-1.8.1** | JPEG/PNG by passthrough; no compressor invoked; `mediaType` stays open; alpha refused and surfaced to the owner | Applied as stated. PNG walked to IHDR/IDAT, JPEG embedded whole; `internal/template/image.go`. |
| **D-1.8.1 (amended)** | Unrecognised `mediaType` is **never** a load error — three surfaces; `Validate` is a dry-run predictor of `Render`; recognised-but-unparseable bytes **remain** a load error; one predicate, one call site | Applied as stated. `decodeRecognisedImage` (loader) vs `DecodeImageForRender` (render, one call site in `render.go`); `folio.Validate` not built (Story 3.7). |
| **D-1.8.2** | 76-column canonical split with padding in place; any input wrapping accepted; strict decode; invalid base64 and empty decoded asset are load errors; hand-wrapped corpus variant | Applied as stated. `base64.go`; `TestSplitBase64CanonicalWidthAndPadding`, `TestSplitBase64Canonical64To76`. |
| **D-1.8.3** | Key = SHA-256 of decoded bytes, validated on load; `/Width`//`Height` → `appendInt`; orphans preserved; Epics 5/6 corollary recorded | Applied as stated. `assetkey.go`; `ImageXObject.Width/Height` via `b.writeInt`; orphan comment on `writeAssets`. |
| **D-1.8.4** | Cross-multiplication for the binding axis; one `ScaleRound` for the free axis; centring through `ScaleRound(bw-dw,1,2)`; `W`/`H` validated at the boundary | Applied as stated. `resolveImagePlacement` (render.go); `validateImageDimensions` (image.go). |
| **D-1.8.5** | No new guard; the structural argument; the mutate-the-bytes behavioural check; Story 1.7's residual gap carried forward **unchanged** | Applied as stated. No new fetch guard added; `TestRenderReadsAssetBytesFromTheDocumentItself`; Story 1.7's gap description untouched. |
| **D-1.8.6** | A third matrix document joining the other two; `/Subtype /Image` guard before comparison on every leg | Applied as stated. `image-embed` added to `matrixDocuments`; `requireImageXObject` checked before hashing in both harness entry points. |
| **D-1.8.7** | The doc's asset example amended for all three defects; replacement **generated by the shipped serializer and cross-checked by a second tool**; fence/golden byte-identity extended to it; the executed-not-read pattern recorded | Applied as stated. Replacement built via `SerializeDocument`, cross-checked with Python `hashlib`/`base64`; `TestAssetExampleMatchesGoldenFragment` (second extractor). |
| **D-1.8.8** | Key validated for **shape and value** as distinct error classes; the three fixtures regenerated **without changing what they test** | Applied as stated. Shape check (`isSHA256HexKey`) before value check; `unknownKeysFixture`/`maximalFixture`/`passthroughLevelBase` regenerated, non-asset content diffed unchanged. |
| **D-1.8.9** | The serializer **re-wraps from decoded bytes, never echoes**; the 64→76 fixture is the discriminating test | Applied as stated. `writeAssets` decodes then calls `splitBase64Canonical`; RP-2 constructed and observed. |
| **D-1.8.10** | The single-flip helper extracted; the guard is **positive** (all emission routes through it); the never-enforced-only-unexercised finding named | Applied as stated. `flipY` (flip.go); `TestContentStreamYCoordinatesRouteThroughFlipY` is a positive routing assertion, not a minus-sign search. |
| **D-1.8.11** | **Recorded, not built.** The manifest gap goes to the Epic 1 boundary gate, and the extension-list fix is explicitly forbidden in favour of the coverage-witness form | Applied as stated — not built. All new test images shipped as Go byte literals (no committed binary needing a manifest row, except `expected.pdf`, consistent with the two existing fixtures). Recorded in Delivery Log for the Epic 1 boundary gate. |
| **D-000.4** | **OVERRIDE — the full four-target matrix runs in this story**, over all three documents; the override reason is logged | Applied as stated. `TestCrossTargetByteIdentity` run locally; all four targets agree with each other and the recorded golden, all three documents — see Delivery Log table. |
| **D-000.6** | Rulings that make canonical documents wrong amend them in this story's commit (the three fixtures; `folio-format.md`'s example), before/after verbatim in the Delivery Log | Applied as stated. See Delivery Log's amendment entries. |
| **D-000.9** *(+ extension)* | Every guard reports a coverage witness and zero-candidates fails; every red-proof records its control | Applied as stated. No-compressor guard, flip-routing guard, image-XObject guard all report a witness and fail on zero. |
| **D-000.10** | This table, complete; reviewer mirrors independently | Applied as stated — this table. |
| **D-000.11** | `-count=1` on every gate; this story reads out-of-module files | Applied as stated. Every gate run in this story used `-count=1` explicitly. |
| **D-000.12** | No bytes/hashes/panic stacks through a shell pipe; `rtk proxy` into a file | Applied as stated. Matrix table and hashmatrix's masked build failure both captured via `rtk proxy` into a file. |
| **D-000.13** | Red-proofs assert rule id and message, never exit status; controls are valid syntax with forbidden semantics | Applied as stated — see Red-proof register. |
| **D-000.14** *(+ extension)* | Load-bearing counts by AST; totals computed or omitted, never narrated beside the artifact | Applied as stated. Test counts measured via `go test -list`/`-v` output, never a filtered grep; no narrated total anywhere in this file. |
| **D-1.1.b** | `/Width`,`/Height` → `appendInt` (the committed routing row); `cm` operands → `appendLength`; no third formatter | Applied as stated. `ImageXObject.Width/Height` via `writeInt`; `appendImageContentStream`'s `cm` operands via `appendLength`. |
| **D-1.2.6** | The D-1.8.1 / D-1.4.9 conflict was surfaced, not arbitrated — **ruled as D-1.8.1 (amended)**; any further ruling conflict met during implementation is surfaced the same way | Applied as stated. No further ruling conflict was met during implementation — nothing new to surface. |
| **D-1.3.1** | AD-1 ban on new non-test files; exact `_test.go` exemption | Applied as stated. New non-test files (`image.go`, `base64.go`, `assetkey.go`, `flip.go`, `imagedoc.go`) import none of `time`/`os`/`math/rand`/`net`; confirmed by `TestForbiddenImportsProductionScan`. |
| **D-1.3.5** | No map ranging under `internal/`; the sorted-keys idiom | Applied as stated. `imageNames := slices.Sorted(maps.Keys(images))` in textdoc.go; no map-range site added under `internal/`. |
| **D-1.4.3** | P1/P2/P3 for the asset-bearing corpus; canonical is a fixed point; no `omitempty` on assets | Applied as stated. `TestP1RoundTripThroughValue`/`TestP2NormalisationIsIdempotent`/`TestP3FixturesAreCanonical` all green over the rebuilt fixtures. |
| **D-1.4.5** | Nothing is rounded at load; one rounding mode in one function | Applied as stated. Centring routed through `geom.ScaleRound` only. |
| **D-1.4.7** | `folio-format.md` is a readability contract; the drift test's witness must be live after the amendment | Applied as stated. `TestDriftGoToDoc`/`TestDriftDocToGo` still pass, non-zero keys compared, after the `##assets` example amendment. |
| **D-1.4.9** | Owner-decided passthrough; higher MINOR loads — **restored, not narrowed, by D-1.8.1 (amended)** | Applied as stated. `TestParseDocumentUnrecognisedMediaTypeNeverALoadError` proves an unrecognised mediaType always loads. |
| **D-1.4.10** | The doc-amendment discipline: cross-check against an independent printer; **the fence/golden byte-identity mechanism AC27b extends is this ruling's, not D-1.4.11's** | Applied as stated. Replacement cross-checked with Python; `TestAssetExampleMatchesGoldenFragment` extends `TestWorkedExampleMatchesGoldenFixture`'s mechanism, correctly attributed to D-1.4.10 in this file. |
| **D-1.4.12** | Closed-set extension is MAJOR; `mediaType` must not join the closed-set inventory | Applied as stated. `mediaType` was not added to `closedsets.go`; `decodeRecognisedImage`'s set lives in `image.go`, deliberately separate. |
| **D-1.5.2** | Validate at the trust boundary; no second `ScaleRound` variant; the validated value is the only reachable one; both polarities retained | Applied as stated. `validateImageDimensions` at the constructor; `DecodedImage.width/height` unexported; `TestExactlyOneExportedScalingFunction` still green (no new geom entry point). |
| **D-1.7.3** | The render-file fence and its residual gap, carried forward with its description intact | Applied as stated. `render_entry.go` untouched by this story; its description of the residual gap is unchanged. |

---

## Red-proof register

Each row records the **mutation**, the **control**, and **what the finding must say** (D-000.13).
Rows marked **vacuous-at-baseline** are the ones where question 2 of AC30 bites hardest.

| # | AC | Mutation | Control (must be green / must build) | Required finding |
|---|---|---|---|---|
| RP-1 | 1 | Serializer strips padding into its own element | Mutation compiles; canonical fixture green before | Names the element lengths, not just "mismatch" |
| RP-2 | 2, 3, 3a | Serializer echoes `Asset.Data` instead of re-wrapping from decoded bytes — i.e. **revert to the baseline behaviour** | Mutation compiles; **canonical-in/canonical-out still green** — that is precisely the point | The 64→76 fixture reddens naming the emitted element lengths. **vacuous-at-baseline** (M-4, D-1.8.9) |
| RP-3 | 8 | Serializer drops assets no element references | Mutation compiles; P3 still green | P1 reddens naming the dropped asset key |
| RP-4 | 6 | Key comparison always returns true | Mutation compiles | Reddens naming expected vs found digest |
| RP-4a | 6a | Shape check removed, value check kept | Mutation compiles; well-formed-but-wrong key still reddens via RP-4 | A **62-character** key reddens with the **shape** message, distinguishable from RP-4's value message (D-1.8.8) |
| RP-11 | 11a | The capability predicate called from the load path instead of the render path | Mutation compiles; a **drawn** unrenderable asset still errors | An **orphaned** unrecognised-`mediaType` asset must **load clean**; reddens naming the wrongly-refused document |
| RP-12 | 11b | Magic-byte/header check removed | Mutation compiles; the WebP-loads case still green | A file declaring `image/png` that is not a PNG reddens as a **load error**, naming the asset key |
| RP-5 | 10 | A real `compress/zlib` import in a real non-test file, used | **Mutation compiles** — valid syntax, forbidden semantics | Guard names the file and the import path; **zero files examined is a hard failure**. **vacuous-at-baseline** (M-8) |
| RP-6 | 14 | `(bw-dw)/2` in place of `ScaleRound(bw-dw,1,2)` | Mutation compiles; odd-difference fixture green before | Reddens **on the offset value**, naming it |
| RP-7 | 17, 17a | Second flip open-coded in the image path | Mutation compiles; **the text path still routes through the helper** — so the guard must fail on the image site specifically, not on a global count | The **positive** guard names the emission site that bypassed the helper; candidate-site count reported, **zero considered is a hard failure**. **vacuous-at-baseline** (M-12, D-1.8.10) |
| RP-13 | 27b | Edit the asset example's fence in the doc without regenerating its golden | Mutation is a doc edit, valid Markdown | The new byte-identity test reddens naming the **asset** fence — and the worked-example assertion stays green, proving the two extractors are independent |
| RP-8 | 23, 23a | Capture a render with the image element removed | Matrix leg otherwise green | Names the document and the missing `/Subtype /Image` — never exit status |
| RP-9 | 15, 15b | Feed a header with `W == 0` | Mutation is a fixture, not a code change; valid file, forbidden value | A **located load error**, and explicitly **not a panic** |
| RP-10 | 9 | An alpha PNG asset | Valid PNG, unsupported kind | Error names the asset key **and alpha** |

**Finisher re-verification (D-000.13, "did it fail for the reason it names?").** The reviewer found
RP-1, RP-6 and RP-7 vacuous at baseline (Findings 6, 2, 3) — the mutation compiled but the shipped
test stayed green. All three were rebuilt and re-run live during finishing, mutation applied then
reverted, restore confirmed byte-identical with `diff`:

- **RP-1** — now reddens on the 1- and 2-padding-character cases (`element 1 has length 31, want
  exactly 76`, etc.), naming the element that broke, not just failing. The 0-padding case correctly
  stays green (the mutation is a no-op there), confirming the fixture discriminates rather than
  failing unconditionally.
- **RP-6** — now reddens exactly once, on `TestResolveImagePlacementOddDifferenceUsesScaleRound`,
  naming the offset value (`drawY = 3, want y + 4 = ...`); the rest of the module (`go test ./...`)
  stays green, confirming the fixture is scoped to the one behaviour under test.
- **RP-7** — now reddens for all three of the reviewer's bypass shapes when applied to a copy of the
  real `imagedoc.go`, and additionally re-confirmed by injecting the same bypass directly into the
  working-tree `imagedoc.go` and running `TestContentStreamYCoordinatesRouteThroughFlipY` (which
  failed, naming the bypass function and line), then reverting.

None of RP-1/RP-6/RP-7 are "vacuous-at-baseline" any longer; that tag applied to their PRE-fix state
and is retained above only as the historical record of what the review found.

---

## Dev Agent Record

### Context Reference

`_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` (working-tree, lines 3598–3916,
319 appended — re-measured: `git diff --unified=0 54a0777 -- .../folio-mvp-decision-log.md` shows
one hunk `@@ -3597,0 +3598,319 @@`; `git show 54a0777:.../folio-mvp-decision-log.md | wc -l` = 3597;
`wc -l` of the working-tree file = 3916. The story banner's own re-measured figure was 298 lines for
the first two rounds; the append grew again after the banner was written (D-1.7.1 amendment +
D-000.14 extension + D-1.8.1–D-1.8.11 + amendment), landing at 319 by the time this story was
developed. Re-measured per this story's own AC32/D-000.12 instruction, not trusted from the banner.)

### Agent Model Used

Claude (bmad-story-developer agent), interactive session, 2026-08-23.

### Debug Log References

None — no failures required a persisted debug log; all iteration happened via direct `go build` /
`go vet` / `go test` cycles, captured inline in this session.

### Completion Notes

- All 15 tasks completed; all binding ACs implemented and tested. Illustrative-tagged ACs (AC10a,
  AC17b, AC18, AC18a, AC20, AC23a, AC25, AC27c) were also implemented, not merely noted.
- **`folio.Validate` was deliberately NOT created** (AC11c) — its rule is recorded in AC11c/D-1.8.1
  (amended) for Story 3.7 to pick up.
- **The licence manifest scanner was deliberately NOT touched** (D-1.8.11) — confirmed
  (`lint/cmd/genmanifest`) that regenerating `lint/MANIFEST.md` against the full working tree
  produces zero diff, i.e. the new fixture's committed `expected.pdf` needs no manifest row, exactly
  matching the two existing fixtures' precedent M-11 records. Recorded for the Epic 1 boundary gate.
- **Improvement beyond ruled scope (flagged per M-2, not silently expanded):** an image element
  referencing an asset key absent from `Document.Assets` is now a located render error
  (`render.go`'s `renderDocument`, "an image element references asset %q, which is not present…").
  M-2 noted this gap exists and is not ruled on; adding the check was judged the only defensible
  behaviour (silently drawing nothing, or panicking on a nil map lookup, are both worse) and is
  recorded here per M-2's own instruction rather than left for a reviewer to guess whether it was
  mandated.
- Three shipped fixtures (`unknownKeysFixture`, `maximalFixture`, `passthroughLevelBase`) were
  rebuilt on real, valid, non-alpha PNG assets under their real SHA-256 keys (AC26); their non-asset
  content was left untouched (AC26a) — confirmed by diff during editing (the only substitutions were
  the asset key string and the `data` array; `passthroughLevelBase`'s injection points and
  `maximalFixture`'s wide field coverage are unchanged).
- `folio-format.md`'s asset example was regenerated end-to-end through the shipped serializer (not
  hand-typed) and independently cross-checked with Python's `hashlib`/`base64` (AC27a); a second,
  independent extractor (`extractAssetExampleFence`) closes it the same way the worked example was
  closed at Story 1.4, with its own red-proof (RP-13) proving the two extractors are independent.
- The full four-target matrix (D-000.4 override) was run locally for all three documents in one
  process (`TestCrossTargetByteIdentity`) and passed: all four targets agree with each other and
  with the recorded golden hash, for `minimal-rect`, `font-text`, AND the new `image-embed` document.
  See the Delivery Log table below.
- **Genuine, observed, mutate-and-revert red-proofs** were run for the three ACs this story flags as
  "green at baseline for the wrong reason" (question 2 of AC30): RP-2 (serializer echo — the new
  `TestHandWrappedAssetNormalisesToCanonical` reddened while `TestP3FixturesAreCanonical` stayed
  green, exactly the vacuity trap D-1.8.9 describes), RP-5 (no-compressor guard) and RP-7 (flip
  routing guard) — both of the latter two ship as permanent, self-contained fixture-based red-proof
  tests (`TestNoCompressorRedProof`, `TestFlipRoutingRedProof`) that construct the forbidden mutation
  as data and assert the guard reddens on it, with a green control, every run. The remaining rows in
  the register are covered **by construction**: each row's "mutation" is itself the fixture/input a
  permanently-shipped test constructs and asserts on directly (e.g. RP-9's `W==0` header, RP-10's
  alpha PNG, RP-4/RP-4a's malformed keys) — these were not additionally cycled as temporary
  source-code mutations, since the discriminating input already IS the red-proof (D-000.13: "a
  control mutation must be valid syntax with forbidden semantics" — a malformed fixture satisfies
  this directly). This is recorded honestly rather than claimed as more than it is.
- `go build ./...`, `go vet ./...` and `gofmt -l` are clean across `folio-go/`, `lint/` and
  `hashmatrix/`. `lint`'s own test suite (44 tests, 4 packages) passes under `GOPROXY=off`.
  `go build -tags=matrix ./...` and `go vet -tags=matrix ./...` are clean.
- One pre-existing, unrelated observation surfaced by AC32's "confirm raw exit codes" discipline:
  `hashmatrix`'s rtk-wrapped `go build ./...` prints "Go build: Success" but the RAW exit code
  (via `rtk proxy`) is 1, because `go build ./...` at the hashmatrix module root collides with the
  `probe/` package directory (`go: build output "probe" already exists and is a directory`). This is
  pre-existing (git status on `hashmatrix/` is clean — nothing in this story touched it), and does
  not affect the actual matrix harness, which always builds via `go build -o <path> ./probe`
  (confirmed working across all four targets in the matrix run above). `go vet ./...` and
  `go test ./...` in `hashmatrix` both pass cleanly. Flagged here per D-000.12/AC32's instruction to
  confirm raw exit codes rather than trust a summarizer, not filed as a defect of this story.

### File List

**New files:**
- `folio-go/internal/template/image.go` — PNG/JPEG chunk/segment walker, `DecodedImage`,
  `DecodeImageForRender`, `UnsupportedMediaTypeError`
- `folio-go/internal/template/base64.go` — canonical base64 split/decode, `DecodeAssetBytes`
- `folio-go/internal/template/assetkey.go` — SHA-256 key shape/value validation
- `folio-go/internal/template/image_test.go`
- `folio-go/internal/template/base64_test.go`
- `folio-go/internal/template/assets_test.go`
- `folio-go/internal/pdf/flip.go` — the one Y-flip function (AD-24)
- `folio-go/internal/pdf/flip_test.go` — AC17a's positive routing guard + RP-7
- `folio-go/internal/pdf/imagedoc.go` — image content-stream (`q`/`cm`/`Do`/`Q`) emission
- `folio-go/render_image_test.go`
- `lint/internal/rules/nocompressor.go` — AC10's guard (RuleNoCompressor, RuleNoImageDecoder)
- `lint/internal/rules/nocompressor_test.go` — production scan + RP-5
- `folio-go/testdata/lint/no-compressor/violating-compressor/bad.go`
- `folio-go/testdata/lint/no-compressor/violating-decoder/bad.go`
- `folio-go/testdata/template/golden/asset-example.json` — AC27b's golden fragment
- `fixtures/image-embed/` (`README.md`, `input.folio`, `expected.json`, `expected.pdf`) — the third
  matrix document (AC22–AC25a)
- `folio-go/template_image_load_test.go` — finisher (Finding 1): public-API reproduction of the
  header-truncation panic, through `folio.ParseTemplate`
- `folio-go/internal/template/closedsets_test.go` — finisher (Finding 5): AC11's AST-based
  closed-set-absence test, plus its own red-proof

**Modified files (developer):**
- `folio-go/internal/template/parse.go` — `decodeAssets`: key shape/value validation, strict base64
  decode, empty-decoded check, recognised-format validation (AC1–AC11b)
- `folio-go/internal/template/serialize.go` — `writeAssets`: canonical re-wrap from decoded bytes
  (D-1.8.9)
- `folio-go/internal/template/fixtures_test.go` — `unknownKeysFixture`/`maximalFixture` rebuilt on
  real assets (AC26)
- `folio-go/internal/template/passthrough_test.go` — `passthroughLevelBase` rebuilt on a real asset;
  injector anchor updated
- `folio-go/internal/template/roundtrip_test.go` — `TestHandWrappedAssetNormalisesToCanonical`
  (AC3's non-vacuous half, RP-2)
- `folio-go/internal/template/goldenfixture_test.go` — `extractAssetExampleFence`,
  `TestAssetExampleMatchesGoldenFragment`, RP-13
- `folio-go/internal/pdf/textdoc.go` — `ImagePlacement`/`ImageXObject` types, image XObject
  emission, `/XObject` resource dict, flip via `flipY`
- `folio-go/internal/pdf/textdoc_test.go` — updated `SerializeTextDocument` call sites;
  `TestAssetKeyEscapeIsIdentity` (AC18a)
- `folio-go/render.go` — `documentBands`, `collectImageRuns`, `resolveImagePlacement`, image wiring
  into `renderDocument`
- `folio-go/render_test.go` — `imageTestTemplateJSON`, `subprocessImageEnvVar`, `TestMain` branch,
  `containsImageXObject`
- `folio-go/fixture_test.go` — `TestRenderMatchesImageEmbedGoldenFixture`
- `folio-go/matrix_test.go` — third `matrixDocument` entry, `requireImageXObject` guard in both
  harness entry points
- `_bmad-output/specs/spec-folio/folio-format.md` — asset example replaced (AC27, AC27a)
- `.github/workflows/matrix.yml` — `image-embed` added to all four render jobs' artifact paths and
  the compare job's document list
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — status only

**Modified files (finisher, code review response):**
- `folio-go/internal/template/image.go` — Finding 1 (bounds-check underflow fix), Finding 9 (R4
  citation), Finding 14 (`PredictorColumns` uses `w`)
- `folio-go/internal/template/image_test.go` — Finding 1 (new truncation-in-header case), Finding 4/11
  (element id/asset key/media type message assertions), Finding 11 (message assertions on the
  zero-width and truncation tests)
- `folio-go/render.go` — Finding 4 (`firstElementIDByAssetKey` carries the real element id)
- `folio-go/render_image_test.go` — Finding 2 (AC14 test rebuilt), Finding 4/11 (message assertions
  on `TestRenderUnrecognisedMediaTypeErrorsOnlyWhenDrawn`)
- `folio-go/internal/pdf/flip_test.go` — Finding 3/13 (guard and RP-7 rebuilt operand-shape-independent)
- `folio-go/internal/pdf/textdoc.go` — Finding 9 (R4 pointer on `ImageXObject`'s doc comment)
- `folio-go/internal/template/base64_test.go` — Finding 6 (padding fixture rebuilt table-driven)
- `folio-go/internal/template/goldenfixture_test.go` — Finding 10 (golden fragment now executed, not
  just compared)
- `_bmad-output/specs/spec-folio/folio-format.md` — Finding 15 (archaeology paragraph moved to the
  Delivery Log, replaced with a one-line note)
- `_bmad-output/implementation-artifacts/1-8-embed-and-render-an-image-from-the-template.md` —
  Finding Resolutions section, corrected disposition table, Delivery Log amendments/gates, AC14/AC17a
  corrections, Red-proof register note, plain-terms opener, status

### Delivery Log

**D-000.4 override.** This story runs the full four-target matrix in its own story (D-000.4 names
Story 1.8 explicitly), over all three documents. Run locally via `TestCrossTargetByteIdentity`
(`go test -tags=matrix -run TestCrossTargetByteIdentity -v .`, `rtk proxy`-captured, raw exit 0):

| Document | darwin/arm64 | linux/amd64 | linux/arm64 | js/wasm | Matches golden |
|---|---|---|---|---|---|
| minimal-rect (fontless) | `0f925e1b…fa6cb4f7c` | same | same | same | yes |
| font-text (template+font) | `dcd453a1…fccbafacbdf` | same | same | same | yes |
| image-embed (template+image) | `e5778eb8…f3506abe689fc` | same | same | same | yes |

All four targets agreed with each other AND with the recorded golden hash, for all three documents,
in one run (`--- PASS: TestCrossTargetByteIdentity (5.62s)`). The full matrix-tagged suite
(`go test -tags=matrix -count=1 .`) also passed, including `TestFMAProbeDiverges`. **Re-run by the
finisher after the review's blocker/major fixes**, per D-000.4's own instruction and the reviewer's
independent re-run: same three documents, same four targets, same golden hashes — none of the fixes
in this pass touch geometry-affecting code on a path any matrix fixture exercises (the AC14 fixture
change lives only in a new, dedicated unit test; `image-embed`'s golden box does not hit the
discriminating ≡3(mod4) residue).

**On Amendments 1–3's key shorthand (Finding 12, AC28).** AC28 is binding that amendments are quoted
**verbatim**. Amendments 1–3 below record the prior (pre-story) placeholder keys in shorthand —
`64×'b'`, `64×'a'`, `62×'c'` — rather than the literal 64-character strings. This is stated explicitly
now, per AC28's own second option ("state explicitly that the keys are recorded in shorthand and
why"): the placeholder strings were never committed to git under any hash the finisher can recover
byte-for-byte from source control (this story file itself is untracked at baseline `54a0777`, so
there is no prior commit to diff against), and inventing a 64-character string to stand in for one
that can no longer be independently verified would itself violate D-000.12/13's discipline against
fabricating evidence. The shapes (`N×'c'`) are accurate descriptions of what shipped before this
story's own edit; Amendment 3's injection-anchor change *is* quoted verbatim below and matches
`passthrough_test.go:125` exactly, since that text is still live in the current tree and directly
checkable.

**Amendment 1 — `folio-go/internal/template/fixtures_test.go`, `unknownKeysFixture`** (D-000.6,
D-1.8.8). Before: key `64×'b'`, data `["Zg=="]` (decodes to the single byte `f`, key mismatched and
not an image). After: key `541581d3ab4d47c46ce5bcfbe86f9e9369f425b41df11decff572d259fa22c65`, data a
real, canonically-wrapped 1x1 grayscale PNG. `futureAssetEntryKey` and the unknown-key passthrough
this fixture exercises are unchanged.

**Amendment 2 — `maximalFixture`** (D-000.6, D-1.8.8). Before: key `64×'a'`, data the first 60 base64
characters of a 70-byte PNG (decodes to 45 bytes, not a valid PNG). After: key
`5a05ad01e89c143b7061b0c93450566568d38a23da9b9c5c9dfe449016433078`, data a real, canonically-wrapped,
non-square 3x2 RGB PNG; the referencing `image` element's `asset` field updated to match. Every other
field this fixture exercises is unchanged.

**Amendment 3 — `passthroughLevelBase`** (D-000.6, D-1.8.8). Before: key `62×'c'` (not even
64 characters), data `["AA=="]` (one `0x00` byte). After: key
`a31f60866b4aa41953176fee9ddb90dc9bc53dce174421f8f567fac364c8bc27`, data a real, canonically-wrapped 2x2 PNG. Remains an orphan (no element references it) both before and after; the
level-injector's anchor for `assets[entry]` was updated from `"data": ["AA=="],` to
`"mediaType": "image/png"` (the injection POINT, not the property under test, changed — the old
anchor text no longer exists in the fixture).

**Amendment 4 — `_bmad-output/specs/spec-folio/folio-format.md`, `## assets` example** (D-000.6,
D-1.8.7). Before: key `"9f2b…c41d"` (ellipsis placeholder), data wrapped at 60+36 columns (96 total,
non-canonical), decoded bytes a 70-byte PNG with an alpha channel (colour type 6). After: key
`a31f60866b4aa41953176fee9ddb90dc9bc53dce174421f8f567fac364c8bc27`, data canonically wrapped at
**76+36** columns, decoded bytes a real 83-byte, 2x2, 8-bit RGB (colour type 2, no alpha),
non-interlaced PNG — generated by `internal/template.SerializeDocument` (the shipped serializer) on a
scratch `Document`, then independently cross-checked with Python's `hashlib`/`base64`. Closed
permanently by `TestAssetExampleMatchesGoldenFragment` (a second extractor,
`extractAssetExampleFence`, keyed to the `## \`assets\`` heading — NOT an edit to
`extractWorkedExampleFence`, which stays scoped to `## Worked example`).

**Correction (Finding 7, Story 1.8 review, D-000.14 extended).** The row above originally narrated
this split as "76+24" — the earlier, correctly-rejected candidate image's numbers (story line 655,
a different 75-byte image AC27a explicitly told the developer NOT to ratify), left behind beside the
re-derived image's own correct figures. Independently re-measured by the reviewer and confirmed here
by decoding the shipped golden fragment and re-splitting it with `splitBase64Canonical`: **76+36**,
element lengths `[76, 36]`, digest matches key, IHDR `w=2 h=2 bitdepth=8 colortype=2 interlace=0`, all
three chunk CRCs valid. This is the fourth recorded instance of a narrated count drifting from its
own artifact (after D-1.7.1's parameter counts, the Story 1.7 disposition totals, and D-1.8.7's
citation) and the first to occur *after* D-000.14 (extended) — added in this same story's own
decision-log diff — was written down: the rule and its own first violation landed in the same
uncommitted change. `TestAssetExampleMatchesGoldenFragment` now also EXECUTES the golden fragment
(decodes it, re-derives its key and canonical split, confirms `image/png` recognises it — Finding 10)
rather than only comparing it against the document's fence, closing the same class of drift at its
root rather than only at this one narrated line.

**AC27b's choice, stated (Finding 8).** Per AC27b's binding "pick one and say which": this story
compares the format spec's `## assets` fence to a **golden fragment** — a standalone `"assets": {…}`
JSON object at `folio-go/testdata/template/golden/asset-example.json` — rather than embedding the
same asset in a full canonical golden and extracting the sub-object. The fragment is close enough to
the serializer's real output that only its indentation differs (Finding 16): the payload — key, both
base64 elements, `mediaType` — is byte-identical to what `SerializeDocument` emits for a document
carrying this exact asset, sliced to the `"assets"` object; the golden file is that output de-indented
by two spaces so it can stand alone as a fence, a hand step the AC's wording ("paste that output into
the doc") did not anticipate needing. This is recorded here rather than left inferable from the test
name and the artifact, per AC27b's own requirement that the choice be *stated*.

**Amendment 5 — `_bmad-output/specs/spec-folio/folio-format.md`, `## assets` section, finishing pass**
(D-000.6; Finding 15, Story 1.8 review). The developer's own amendment (Amendment 4 above) inserted a
paragraph of implementation archaeology between the fence and the sentence explaining it — naming
"Story 1.8", ruling ids "D-1.8.7"/"D-1.8.1", and a Go test file path — into a document whose own
opening states it is *"a public contract, not an implementation detail."* Moved to this Delivery Log
(where AC28 already wants this kind of record) and replaced with a one-line, ruling-id-free note
placed after the explanatory sentence instead of before it, so the example and its explanation stay
adjacent. Before (the paragraph removed, quoted verbatim):

> Story 1.8 (D-1.8.7): this example was previously a hand-authored placeholder that was invalid three
> ways — an ellipsis key, non-canonical wrapping (60+36 columns), and a PNG with an alpha channel
> (colour type 6), which this format's own rules refuse (`§assets`, D-1.8.1). The example above is a
> real, supported, non-alpha PNG generated by the shipped `internal/template` serializer (its key is
> the real SHA-256 of the decoded bytes, and the wrapping is the canonical 76-column split) and was
> independently cross-checked with a second tool (Python's `hashlib`/`base64`) before being pasted
> here. `folio-go/internal/template/goldenfixture_test.go`'s
> `TestAssetExampleMatchesGoldenFragment` extracts this exact fence and asserts it byte-identical
> against a golden fragment on every run, so it cannot rot again silently.

After (inserted immediately following "Elements reference an asset by its key.", quoted verbatim):

> *(This is a real, supported, non-alpha PNG — canonically wrapped and independently verified; a test
> keeps it byte-identical to this fence on every run, so it cannot silently drift. See the Delivery
> Log for how it was derived.)*

The full substance of the removed paragraph is preserved above, in this Delivery Log entry, rather
than lost.

**Executed vs. covered inventory (AC27c).** `folio-format.md`'s fenced blocks now EXECUTED (byte-
identity asserted against a golden): the `## Worked example` fence (`TestWorkedExampleMatchesGoldenFixture`,
Story 1.4/D-1.4.10) and the `## assets` fence (`TestAssetExampleMatchesGoldenFragment`, this story).
Every other fenced block in the document is merely COVERED (field names only, via the drift test,
D-1.4.7) — not extended to execution in this story, per AC27c's explicit scope limit.

**Third matrix document (AC22–AC25a).** `fixtures/image-embed/` added, joining `minimal-rect` and
`font-text` (neither replaced). Wired into: `matrixDocuments` (matrix_test.go), the per-document
`/Subtype /Image` guard (`requireImageXObject`, checked in both `TestCrossTargetByteIdentity` and
`TestTargetRenderHash`, before any hash comparison — AC23), the untagged golden test
(`TestRenderMatchesImageEmbedGoldenFixture`), and `.github/workflows/matrix.yml` (all four render
jobs' artifact paths, plus the compare job's `docs="minimal-rect font-text image-embed"` list, which
names files explicitly and never globs).

**Finisher pass (code review response, 2026-08-23).** Every one of the review's 16 findings is FIX
(see [§ Finding Resolutions](#finding-resolutions) below for the full per-finding triage and the
corrected disposition table). In order of severity:

- **Blocker 1 (panic).** `internal/template/image.go`'s PNG chunk-length bounds check compared
  `length > uint32(len(data)-pos)-4`: when fewer than 4 bytes remained, `uint32(len(data)-pos)-4`
  wrapped to ~4.29e9, the broken guard passed, and the next line sliced past the buffer end. Fixed by
  comparing entirely in `int` (`len(data)-pos < int(length)+4`), which cannot wrap. **This is the
  second denial-of-service-class defect in this run** (Story 1.6's hang was the first) and the second
  time Story 1.4's user-facing promise — *"malformed templates rejected with a located error"* — has
  been broken, this time by a bounds-check underflow rather than an unbounded loop. Re-verified: the
  reviewer's own reproduction (a PNG truncated in the chunk header, through the **public**
  `folio.ParseTemplate`) now returns a located `*template.LoadError` naming the asset key and the
  word "truncated PNG", not a panic — confirmed live both as a permanent test
  (`TestParseTemplateRejectsPNGTruncatedInHeaderWithoutPanicking`, `folio-go/template_image_load_test.go`)
  and by hand-injecting the same shape into `decodePNG`'s own unit tests
  (`TestDecodePNGRejectsTruncationInHeader`). `decodeJPEG`'s own segment walk was checked for the same
  underflow shape (Finding 1's suggestion) and found sound: every length comparison there is already
  performed in `int` with no unsigned intermediate.
- **Blocker 2 (AC14 circular/insufficient).** `render_image_test.go`'s centring test computed its own
  expected value from the same `geom.ScaleRound` call the production code evaluates (could only fail
  if two evaluations of one expression disagreed) and its fixture's differences were 0 and 33334 —
  neither odd. Rebuilt with literal, hand-computed expected offsets and a fixture whose free-axis
  difference is exactly 7 (≡ 3 mod 4, the residue class where round-half-to-even and truncation
  actually disagree — AC14's own "any odd difference" wording is corrected in the Acceptance Criteria
  above, since round-half-to-even and truncation agree on most odd values). RP-6 (replacing both
  `ScaleRound` calls with `/2`) now reddens, naming the offset value, and was re-run to confirm.
- **Blocker 3 (flip guard naming convention).** `internal/pdf/flip_test.go`'s AD-24 guard only
  considered Y-suffixed bare identifiers passed directly to `appendLength`/`writeLength`; a
  differently-named variable and an inline expression both passed silently and were not even counted
  as candidates. Rebuilt operand-shape-independent: it now asserts, by AST, that no function in
  `internal/pdf` other than `flipY` contains a subtraction whose left-associative chain bottoms out at
  a page-height-shaped operand, regardless of variable naming. RP-7 was rebuilt to mutate a copy of
  the REAL `imagedoc.go` (not a synthetic fixture) with all three of the reviewer's bypass shapes, and
  all three now redden — re-confirmed live by injecting the same bypass into the actual working-tree
  `imagedoc.go`, running the guard, and reverting. This also closes Finding 13 (the vacuity witness):
  the guard now requires a genuine `flipY(...)` call in both `textdoc.go` and `imagedoc.go` by name.
- **Major 4 (missing element id).** `render.go`'s dedup loop called `DecodeImageForRender` with a
  hard-coded `""` element id. Fixed by carrying the first referencing element id alongside each
  distinct asset key (the information was already in `imageRuns`); both the helper test and the
  render-path consumer test now assert the element id, asset key and media type all appear in the
  message.
- **Major 5 (AC11's missing test).** Added `internal/template/closedsets_test.go`: an AST-based
  enumeration of `closedsets.go`'s package-level `closed*` declarations, asserting a non-zero count
  and that none names or contains `mediaType`, red-proved against a scratch copy with a
  `closedMediaTypes` map appended.
- **Major 6 (AC1's vacuous padding fixture).** `base64_test.go`'s padding test ran only against an
  81-byte fixture (a multiple of 3, zero padding characters) — every padding assertion ranged over
  strings with no `=` and could not fail. Rebuilt table-driven over 0/1/2 padding characters; RP-1
  ("serializer strips padding into its own element") now reddens the 1- and 2-padding cases.
- **Major 7 (Delivery Log's 76+24 vs. artifact's 76+36).** Corrected above (Amendment 4's
  correction note) — the fourth recorded instance of this drift class, and the first one after
  D-000.14 (extended) was written down in this same story's decision-log diff.

**Findings fixed without a change in production behaviour** (documentation/comment-only or
test-hardening; full rationale for each is in
[§ Finding Resolutions](#finding-resolutions)): Finding 8 (AC27b's choice, stated above), Finding 9
(R4 citation added to `image.go`'s `Stream` comment and `textdoc.go`'s `ImageXObject` comment),
Finding 10 (`TestAssetExampleMatchesGoldenFragment` now executes the golden fragment), Finding 11
(the three named red-proofs now assert on message content, not `err != nil`, closing the same gap
that let Findings 1 and 4 through), Finding 12 (shorthand keys explained above), Finding 14
(`PredictorColumns` now uses the validated `w`, not the raw header `width`), Finding 15 (Amendment 5
above), Finding 16 (noted in Amendment 4's AC27b paragraph above).

**Epic 1 boundary report items** (recorded per AC12, M-11, D-1.8.11 — not built in this story):
1. **Alpha PNG limitation (AC12).** Transparent PNGs are refused (`decodePNG` rejects colour types
   4 and 6 by name, naming "alpha" in the error). The golden report's header is expected to carry a
   logo (Epic 4) — this is not a one-way door (`/SMask` is purely additive later), but the owner
   should see this now, before the golden report is designed.
2. **Licence manifest gap (M-11, D-1.8.11).** `lint/internal/manifest/manifest.go`'s
   `fontExtensions` scan does not examine committed images (`.png`/`.jpg`). Confirmed harmless for
   this story (byte-literal test images, no manifest row needed for `expected.pdf`, consistent with
   the two existing fixtures). **The ruled fix shape is NOT an extension-list addition** — D-1.8.11
   requires a coverage-witness form ("every committed file in the declared asset/fixture locations
   is accounted for, and anything uncovered fails") — this is explicitly deferred to the Epic 1
   boundary gate.

**Gate results, re-run by the finisher after all 16 findings were addressed (raw exit codes,
`-count=1` throughout, D-000.11/D-000.12/AC32):**

| Gate | Result |
|---|---|
| `folio-go`: `go build ./...` | exit 0 |
| `folio-go`: `go vet ./...` | exit 0, no issues |
| `folio-go`: `gofmt -l .` | exit 0, clean (empty output) |
| `folio-go`: `go test ./... -count=1` | exit 0, 7 packages |
| `folio-go`: `go build -tags=matrix ./...` | exit 0 |
| `folio-go`: `go vet -tags=matrix ./...` | exit 0, no issues |
| `folio-go`: `go test -tags=matrix -run TestCrossTargetByteIdentity -count=1 .` | exit 0, all four targets × three documents byte-identical, matching recorded goldens |
| `lint`: `go build ./...` / `go vet ./...` | exit 0 / exit 0, no issues |
| `lint`: `GOPROXY=off go test ./... -count=1` | exit 0, 44 passed, 4 packages |
| `lint`: `gofmt -l .` | exit 0, clean |
| `hashmatrix`: `go vet ./...` / `go build -o <tmp> ./probe` | exit 0 / exit 0 |
| `hashmatrix`: `go build ./...` (no `-o`) | exit 1 — pre-existing, unrelated: `go: build output "probe" already exists and is a directory` (Story 1.2 Finding 6's already-recorded shape; `go build -o <path> ./probe`, the invocation the harness itself uses, succeeds) |
| Licence manifest (`lint/cmd/genmanifest`) | Regenerated; zero diff against committed `lint/MANIFEST.md` (confirms D-1.8.11's scope stayed out — `git diff 54a0777 -- lint/` still touches only the two nocompressor files) |
| `git diff 54a0777 -- folio-go/go.mod folio-go/go.sum` | empty — no dependency change |

Test count, re-measured by the finisher (D-000.14: by tool output, never a filtered grep): **184
top-level / 288 with subtests / 7 packages, 0 skipped, 0 failed** (`go test ./... -count=1 -v`,
`--- PASS:` lines counted directly). The reviewer's own pre-fix measurement was 180 top-level / 277
with subtests; the finisher's fixes added `TestParseTemplateRejectsPNGTruncatedInHeaderWithoutPanicking`,
`TestDecodePNGRejectsTruncationInHeader`, `TestClosedSetsNeverIncludeMediaType`,
`TestClosedSetsMediaTypeRedProof`, plus new subtests under `TestSplitBase64CanonicalWidthAndPadding`
and `TestFlipRoutingRedProof`. No total is narrated beside this row in the body text above — it is
stated here, once, as the measurement itself.

---

## QA Results

### Review Summary

- **Reviewed by:** bmad-code-reviewer (independent mirror; the developer's disposition table and
  Completion Notes were read only AFTER the reviewer's own 32-row table was derived from the code)
- **Date:** 2026-08-23
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 3
- **Majors:** 4
- **Minors:** 6
- **Nits:** 3

**Measurement discipline.** Every gate below was run with `-count=1`, raw exit codes captured via
`rtk proxy` into a file, never through a summarising wrapper. All mutation work was performed on a
`cp -a` copy of the tree under the scratchpad; the working tree was verified byte-identical (same
`git status --porcelain` file list, same `git diff --stat 54a0777` totals) before and after this
review. No production file, test file, fixture or document was modified by this review.

**Independently re-measured (reviewer's own numbers, not the developer's):**

| Measurement | Reviewer's value |
|---|---|
| `folio-go` `go test ./... -count=1` | 7 packages, **exit 0** |
| Top-level test functions passing | 180 |
| `PASS:` lines including subtests | 277 |
| Skipped tests | 0 |
| `lint` `go test ./... -count=1` | 4 packages, exit 0 |
| Four-target matrix (`-tags=matrix`, `TestCrossTargetByteIdentity`) | **exit 0**, 3 documents × 4 targets, all byte-identical |
| `image-embed` hash, all four targets | `e5778eb872c98ec4a3c3c89466a8313cf52931b896701de8a43f3506abe689fc` (matches `fixtures/image-embed/expected.json`) |

The developer's reported "277 tests passed" reconciles exactly with the `PASS:`-including-subtests
measure. **No discrepancy.** The orchestrator's pre-launch measurement (7 packages, exit 0) also
reconciles.

---

### Reviewer's independent ruling disposition table (D-000.10 mirror)

Derived from the code before the developer's table was read. Legend: **✅** the reviewer's finding
agrees with the developer's "Applied as stated"; **❌** it does not. Totals are computed from the
marks in this table, not narrated (D-000.14 extended).

| # | Ruling | Reviewer's independent disposition | Agree? |
|---|---|---|---|
| 1 | **D-1.8.1** | Passthrough is REAL — the rendered PDF was verified to contain the source file's own concatenated IDAT bytes verbatim; JPEG embeds the whole file; alpha/palette/16-bit/interlaced all produce located load errors naming key and reason (all five verified against generated fixtures). **But AC10a's mandated R4 citation is absent from the emission-site comment** (F-9). | ❌ partial |
| 2 | **D-1.8.1 (amended)** | Surfaces 1 and 2 verified (`image/webp` loads; an orphaned WebP renders clean; a *drawn* WebP errors; `image/png` with non-PNG bytes is a load error). **But the Render error does NOT name the element id** — the binding verdict-table clause is unmet (F-4). | ❌ |
| 3 | **D-1.8.2** | 76-column split, any-wrapping accept, strict decode, empty-decoded rejection all verified. **But AC1's binding padding clause is asserted against a fixture that contains no padding at all** (F-6). | ❌ partial |
| 4 | **D-1.8.3** | Key = SHA-256 of decoded bytes, validated on load; `/Width`,`/Height` via `b.writeInt`; orphan preservation red-proved by the reviewer (dropping unreferenced assets reddens 6 tests); AC8a's corollary comment present verbatim. | ✅ |
| 5 | **D-1.8.4** | Cross-multiplication correct and genuinely covered (inverting the comparison reddens 2 tests). **But the centring-through-`ScaleRound` clause is entirely unproven** (F-2) **and `decodePNG` panics on a malformed header** — the exact clause D-1.8.4(4) imports from D-1.5.2 (F-1). | ❌ |
| 6 | **D-1.8.5** | No new guard added; `TestRenderReadsAssetBytesFromTheDocumentItself` genuinely mutates in-memory bytes and asserts the rendered bytes differ; Story 1.7's residual-gap description untouched. | ✅ |
| 7 | **D-1.8.6** | Third document joins the other two; `requireImageXObject` verified to fire on the FIRST leg before any hash is computed (proved by rendering an imageless document); present in both harness entry points and the untagged golden test; the `FontFile2` guard was not generalised away. | ✅ |
| 8 | **D-1.8.7** | The replacement example was independently re-verified by the reviewer with a second tool: key = real SHA-256 of decoded bytes, canonical 76+36 split, IHDR colour type 2 (no alpha), non-interlaced, 8-bit, all three chunk CRCs valid. Second extractor present and its independence red-proved. **But** the golden is hand-de-indented serializer output (F-16), nothing executes either side (F-10), AC27b's mandated Delivery-Log statement is missing (F-8), and Amendment 4 narrates the wrong column split (F-7). | ❌ partial |
| 9 | **D-1.8.8** | Shape and value fire as two distinct, cross-asserted messages. All three fixtures' non-asset content verified unchanged apart from the necessary element `asset` reference and the passthrough injection anchor (which moves within the same object, preserving the nesting level the level-parameterised test walks). | ✅ |
| 10 | **D-1.8.9** | **RP-2 reproduced exactly.** With `writeAssets` echoing `a.Data`, `TestHandWrappedAssetNormalisesToCanonical` goes RED and `TestP3FixturesAreCanonical` stays GREEN — and it is the *only* test in the whole module that reddens. The discriminating fixture genuinely discriminates. | ✅ |
| 11 | **D-1.8.10** | **The guard is not a routing assertion.** It keys on a variable-naming convention and is defeated by a rename or an inline expression (F-3). | ❌ |
| 12 | **D-1.8.11** | `git diff 54a0777 -- lint/` is empty; `fontExtensions` untouched; recorded for the Epic 1 boundary gate with the correct (coverage-witness, not extension-list) fix shape. | ✅ |
| 13 | **D-000.4** | **Re-run by the reviewer**: exit 0, three documents × four targets, all byte-identical, `image-embed` matching its golden. | ✅ |
| 14 | **D-000.6** | All four canonical documents amended in this commit. **But the before/after is abbreviated rather than verbatim** (F-12) and one amendment's measured figure contradicts its artifact (F-7). | ❌ partial |
| 15 | **D-000.9 (+ ext)** | The no-compressor and image-XObject witnesses are sound. **The flip guard's witness is checked as `candidates == 0` while its own failure message asserts a stronger per-file requirement** — a real bypass drops the witness 2→1 and passes silently (F-13). | ❌ partial |
| 16 | **D-000.10** | The table is present and its enumeration is complete — the reviewer's independently derived row set matches it exactly, ruling for ruling. | ✅ |
| 17 | **D-000.11** | Verified: every gate re-run by the reviewer used `-count=1`, including the out-of-module document and fixture reads. | ✅ |
| 18 | **D-000.12** | Verified: no bytes, hashes or panic stacks were taken through a pipe, by the developer's account and by the reviewer's own method. | ✅ |
| 19 | **D-000.13** | **Several shipped red-proofs assert only `err != nil`** with no message assertion — which is precisely how F-1 and F-4 escaped (F-11). | ❌ |
| 20 | **D-000.14 (+ ext)** | **Violated in the commit that records it**: the Delivery Log narrates a column split beside the artifact it summarises, and the narration is wrong (F-7). | ❌ |
| 21 | **D-1.1.b** | `/Width`,`/Height` route through `b.writeInt`; `cm` operands route through `appendLength`; no third formatter introduced. | ✅ |
| 22 | **D-1.2.6** | No further ruling conflict arose. AC9a pre-authorised the load-time reading of alpha as "a deliberate reading of the amended ruling", so implementing it was not an unsurfaced arbitration. | ✅ |
| 23 | **D-1.3.1** | New non-test files import only `fmt`, `crypto/sha256`, `encoding/hex`, `encoding/base64` and internal packages; the forbidden-imports production scan is green. | ✅ |
| 24 | **D-1.3.5** | `slices.Sorted(maps.Keys(images))` used in `textdoc.go`; the map-range guard is green; no new map-range site under `internal/`. | ✅ |
| 25 | **D-1.4.3** | P1/P2/P3 green over the rebuilt corpus, and P1 proved load-bearing for orphans by the reviewer's RP-3 mutation. | ✅ |
| 26 | **D-1.4.5** | The code as written keeps one rounding mode in one function — **but nothing guards it**; substituting `/2` passes the entire suite (F-2). | ❌ |
| 27 | **D-1.4.7** | Four drift tests pass after the amendment, each reporting a non-zero, fail-on-zero witness. | ✅ |
| 28 | **D-1.4.9** | An unrecognised `mediaType` always loads; unknown-key passthrough at the asset-entry nesting level still exercised. | ✅ |
| 29 | **D-1.4.10** | Attribution correctly corrected to D-1.4.10; second extractor is genuinely independent. **But the byte-identity pair proves only doc == golden, so both can still rot together** (F-10). | ❌ partial |
| 30 | **D-1.4.12** | `mediaType` is correctly absent from `closedsets.go`. **But AC11's binding mechanical half — a test asserting that absence — was not built** (F-5). | ❌ |
| 31 | **D-1.5.2** | **`decodePNG` panics on a malformed PNG reachable from the public `ParseTemplate`/`LoadTemplate`** (F-1); additionally `PredictorColumns` is populated from the raw header value rather than the validated one (F-14). | ❌ |
| 32 | **D-1.7.3** | Verified across the whole diff, including every new untracked file: no added comment or README line describes the `net` ban or the AST render-file fence as a proof, guarantee or security boundary. | ✅ |

**Computed from the marks above:** ✅ = 18, ❌ = 14, rows = 32.

**The developer's table records "Applied as stated" for all 32.** The reviewer disagrees with 14 of
those 32 dispositions. Five of the disagreements (rows 2, 5, 11, 30, 31) are cases where a **binding**
clause of the named ruling is not met by the shipped code; the rest are cases where the substance
ships but the ruling's own enforcement mechanism does not.

---

### The vacuous-at-baseline split — the thing this review was asked to audit

The developer's Completion Notes claim **three** genuine mutate-and-observe cycles (RP-2, RP-5,
RP-7) and place the rest under "covered by construction — the discriminating input *is* the
red-proof". The reviewer re-ran the register row by row, asking D-000.13's question of each: *did it
fail for the reason it names?*

| RP | Claimed | Reviewer's result |
|---|---|---|
| RP-1 | by construction | **FAILS.** The AC's own test passes under the exact named mutation (F-6) |
| RP-2 | genuine cycle | **CONFIRMED.** Reproduced exactly; discriminating fixture reddens, canonical control stays green, and it is the sole test that reddens |
| RP-3 | by construction | **CONFIRMED.** Dropping unreferenced assets reddens 6 tests including P1 |
| RP-4 / RP-4a | by construction | **CONFIRMED.** Shape and value messages are distinct and cross-asserted |
| RP-5 | genuine cycle | **CONFIRMED.** A real `compress/zlib` + `image/png` import in a real non-test file reddens the guard, naming both rule ids, the file and the import paths |
| RP-6 | by construction | **FAILS — nothing whatsoever covers it** (F-2) |
| RP-7 | genuine cycle | **FAILS.** The shipped cycle exercises a synthetic tempdir fixture; against the real `imagedoc.go` two of three bypass shapes pass green (F-3) |
| RP-8 | by construction | **CONFIRMED.** Rendering the document without its image element fires the guard on the first leg, before any hash |
| RP-9 | by construction | **PARTIAL.** `W == 0` is a located load error, but the adjacent corrupt-header case panics (F-1) |
| RP-10 | by construction | **CONFIRMED.** The alpha message names the asset key and "alpha" |
| RP-11 | by construction | **PARTIAL.** The behaviour is right for the orphan/drawn split but the assertion is on `err != nil` only, which is how F-4 escaped |
| RP-12 | by construction | **CONFIRMED.** |
| RP-13 | by construction | **CONFIRMED.** The two extractors are genuinely independent |

**"Covered by construction" is a sound argument only where the shipped assertion actually
discriminates.** Three rows (RP-1, RP-6, RP-7) do not, and two more are partial. RP-6 is the severe
one: the row's stated control — *"odd-difference fixture green before"* — was never checked, and the
fixture's differences are not odd.

---

### Finding 1: `decodePNG` panics on a malformed PNG, reachable from the public loader

- **Severity**: Blocker
- **Category**: Correctness / Security
- **Location**: `folio-go/internal/template/image.go:184` (reached from `image.go:189`); called at
  load time from `folio-go/internal/template/parse.go:386`; public entry points
  `folio-go/folio.go:25` (`LoadTemplate`) and `folio-go/folio.go:36` (`ParseTemplate`)
- **Observation**: The chunk-walker's bounds check is

  ```go
  if length > uint32(len(data)-pos)-4 {
  ```

  `len(data)-pos` is an `int` converted to `uint32`. When fewer than four bytes remain, the
  subtraction **wraps**: `uint32(0) - 4` is `4294967292`. The guard therefore passes for essentially
  every declared `length`, and the next line, `body := data[pos : pos+int(length)]`, indexes past the
  end of the buffer. The preceding `len(data)-pos < 8` check happens *before* `pos += 8`, so it does
  not cover this window.

  Reproduced by the reviewer with a 16-byte input (PNG signature + one chunk header, no CRC), fed
  through the **public** API:

  ```
  PANIC from the PUBLIC folio.ParseTemplate on untrusted input:
    runtime error: slice bounds out of range [:29] with capacity 18
  ```

  The shipped `TestDecodePNGRejectsTruncation` truncates 20 bytes off an 81-byte PNG, which lands
  mid-IDAT-body and *is* caught by the same comparison — so the existing test passes for a reason
  that does not reach this window.
- **Impact**: `LoadTemplate`/`ParseTemplate` are the library's untrusted-input boundary. A `.folio`
  file — valid JSON, correct asset key, well-formed base64 — crashes the calling process instead of
  returning an error. This is the exact failure mode AC15/AC15b and D-1.8.4(4) exist to forbid
  (*"a `W == 0` from a corrupt header must be a located load error, never a panic"*), reached one
  branch earlier than the case that was tested. It also defeats D-1.5.2's whole posture: the
  validated-value guardrail is never reached, because the walker dies before producing a header.
- **Suggested Resolution**: Compare in `int` with no wrap, e.g. `if int64(length) > int64(len(data)-pos)-4 { … }`,
  or check `len(data)-pos < int(length)+4` directly. Add a table-driven test covering remaining-byte
  counts 0, 1, 2 and 3 after a chunk header, asserting a **located** `template.LoadError` naming the
  asset key — and assert on the message, not on `err != nil` (see Finding 11). Consider a
  `recover`-free fuzz or table sweep over truncations at every offset of a known-good PNG; the same
  class of bug should be checked in `decodeJPEG`'s segment walk.
- **Related AC**: AC15, AC15b, AC9, AC11b (D-1.8.4 clause 4, D-1.5.2)

---

### Finding 2: AC14's centring-through-`ScaleRound` is entirely unproven — the named RP-6 mutation leaves the whole suite green

- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/render.go:181-182` (the two `geom.ScaleRound(bw-drawW, 1, 2)` /
  `geom.ScaleRound(bh-drawH, 1, 2)` calls); test at
  `folio-go/render_image_test.go:185-206` (`TestResolveImagePlacementOddDifferenceUsesScaleRound`)
- **Observation**: Two independent defects compound here.

  **(a) The test asserts the production expression against itself.** It computes
  `wantOffsetX := geom.ScaleRound(run.boxW-dw, 1, 2)` and compares it to `drawX - run.x`. The
  expected value is produced by the very call the AC exists to pin, so the test can only fail if the
  two evaluations of the same expression disagree.

  **(b) The fixture cannot discriminate even in principle.** With `boxW = boxH = 100001` and a 3×2
  image: width binds, `dw = 100001` so `bw - dw = 0`; `dh = ScaleRound(100001, 2, 3) = 66667` so
  `bh - dh = 33334`. **Neither difference is odd** — one is zero and one is even — contradicting
  AC14's own statement that *"the retained fixture uses an odd difference on both axes, so `/2` and
  `ScaleRound(…,1,2)` disagree."* The shipped `image-embed` golden fixture is no better: box
  100×60 with a 3×2 image gives height-binds, `dw = 90`, differences 10 and 0.

  The reviewer applied RP-6 verbatim — replacing both calls with `(bw-drawW)/2` and `(bh-drawH)/2` —
  and ran the full module:

  ```
  go test ./... -count=1   →  EXIT=0,  zero failures
  ```

  Not one test in `folio-go` reddens, and the `image-embed` golden hash is unchanged.

  **(c) A further nuance the AC itself gets wrong**, measured directly against `geom.ScaleRound`:
  round-half-to-even and truncation agree on *most* odd values. They differ only when the difference
  is ≡ 3 (mod 4):

  ```
  v=0      ScaleRound=0       v/2=0       differ=false
  v=3      ScaleRound=2       v/2=1       differ=true
  v=5      ScaleRound=2       v/2=2       differ=false
  v=7      ScaleRound=4       v/2=3       differ=true
  v=9      ScaleRound=4       v/2=4       differ=false
  v=11     ScaleRound=6       v/2=5       differ=true
  v=33334  ScaleRound=16667   v/2=16667   differ=false
  ```

  So "use an odd difference" is not a sufficient instruction; a fixture built to AC14's literal
  wording would still have a 50% chance of being vacuous.
- **Impact**: The single most-emphasised geometric invariant in D-1.8.4 — *"the program keeps exactly
  one rounding mode in exactly one function"* — has no coverage at all. A future refactor to `/2`
  (the natural, obvious simplification, and exactly what D-1.8.4 predicts someone will write) ships
  green, silently introducing a second rounding behaviour into the coordinate path and changing
  emitted PDF bytes for odd-difference boxes. This is the third consecutive story in which a
  "green at baseline for the wrong reason" AC was declared covered by construction and was not.
- **Suggested Resolution**: (i) Stop deriving the expected value from `ScaleRound`; write the expected
  offsets as **literal integers** computed by hand. (ii) Choose box and image dimensions that make
  **both** differences ≡ 3 (mod 4) — e.g. a difference of 7 on each axis, where `ScaleRound` gives 4
  and `/2` gives 3 — and assert the offset value, naming it, per AC14's "must redden on the offset
  value". (iii) Re-run RP-6 and record the observed red in the register. (iv) Consider making the
  `image-embed` golden fixture itself use a discriminating box so the matrix hash also pins it.
- **Related AC**: AC14, AC13 (D-1.8.4 clause 3, D-1.4.5, AD-2)

---

### Finding 3: the AD-24 flip guard is a naming-convention check, not a routing assertion — two of three bypass shapes pass green

- **Severity**: Blocker
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/internal/pdf/flip_test.go:123-127` (the discriminator) and `:174-178`
  (the vacuity check); guard subject `folio-go/internal/pdf/imagedoc.go:16-26`
- **Observation**: D-1.8.10 is explicit that the guard must *"assert that **all** content-stream
  coordinate emission routes through the one flip function."* The shipped scanner instead does this:

  ```go
  ident, ok := valueArg.(*ast.Ident)
  if !ok || !strings.HasSuffix(ident.Name, "Y") {
      return true // not a Y-suffixed operand: not a candidate site by convention
  }
  candidateEmissionSites++
  ```

  A site is only *considered* if its operand is (a) a bare identifier and (b) spelled with a trailing
  `Y`. Anything else is skipped silently — not flagged, and **not even counted**.

  The reviewer open-coded a genuine second flip in the real `internal/pdf/imagedoc.go` three ways and
  ran `TestContentStreamYCoordinatesRouteThroughFlipY` (`-count=1`) against each:

  | Bypass in `appendImageContentStream` | Guard result |
  |---|---|
  | `secondY := page.Height - page.MarginTop - im.Y - im.DrawHeight` → `appendLength(dst, secondY)` | **FAIL** (names `secondY`) |
  | `bottomEdge := page.Height - page.MarginTop - im.Y - im.DrawHeight` → `appendLength(dst, bottomEdge)` | **PASS** |
  | `appendLength(dst, page.Height-page.MarginTop-im.Y-im.DrawHeight)` (inline, no variable) | **PASS** |
  | control (unmutated) | PASS |

  Both green rows are real AD-24 violations: a second function inverting a coordinate and writing it
  straight into the content stream. The guard's own doc comment concedes the gap and then argues it
  away — *"a differently-named variable computed by hand (missing the 'Y'-suffix + emission pairing
  this guard requires — itself a finding, since a Y-suffixed name is exactly the shape a second flip
  site would plausibly use)"* — but nothing in the tree detects that shape. The rule being enforced
  is a naming convention that no guard enforces.

  The shipped RP-7 (`TestFlipRoutingRedProof`) does not touch the real emission path at all; it
  writes three synthetic files into `t.TempDir()`. It therefore proves the scanner's logic on inputs
  chosen to fit the scanner, which is the shape D-1.8.10 warns about in the abstract.
- **Impact**: AC17 calls a second flip site *"the single likeliest structural defect in this story"*,
  and AC17b/D-1.8.10 make the point that the invariant *"was never enforced, only unexercised."*
  After this story it is still not enforced — it is guarded against one spelling of the violation and
  blind to the two others, including the shortest one (inlining). Story 1.8 was designated AD-24's
  first real test; it does not discharge it.
- **Suggested Resolution**: Make the guard operand-shape-independent. Since `internal/pdf` is one
  small package, the tractable positive form is: for every call to `appendLength`/`writeLength`
  inside a function that emits into a content stream, resolve the value argument to its defining
  expression and require it to be either (a) a direct `flipY(...)` call, or (b) an expression that
  provably contains no subtraction involving `page.Height`/`pageHeight`. Simplest sound version:
  require that **no function other than `flipY` contains a subtraction whose left operand derives
  from a page-height field**, expressed positively as "every such subtraction's enclosing function is
  `flipY`" — this is falsifiable and does not fire on unrelated arithmetic, unlike the rejected
  minus-sign search. Whichever form is chosen, the red-proof must be run against the **real**
  `imagedoc.go` for all three bypass shapes above, and the results recorded in the register.
- **Related AC**: AC17, AC17a (D-1.8.10, AD-24)

---

### Finding 4: `Render`'s unrecognised-`mediaType` error does not name the element id

- **Severity**: Major
- **Category**: AC Conformance
- **Location**: `folio-go/render.go:380` — `template.DecodeImageForRender(asset.MediaType, raw, key, "")`;
  error type at `folio-go/internal/template/image.go:79-92`; test at
  `folio-go/render_image_test.go:214-243`
- **Observation**: D-1.8.1 (amended)'s verdict table is binding and states, for the `Render` surface:
  *"Error — only when an element actually needs to draw it. **Located, naming element id, asset key
  and media type**."* AC11a carries it verbatim.

  The element id is hard-coded as the empty string at the sole call site, because the decode happens
  inside the dedup loop over `assetKeys`, where the referencing element is no longer in scope. The
  reviewer's probe produced:

  ```
  folio: Render: template: element : asset 006db4087cb3…aed8: this version cannot render
  media type "image/webp" (the document is valid — mediaType is an open set, D-1.4.12; …)
  ```

  Note `element : asset` — the field is not merely missing, it renders as a visible hole.

  This escaped because the helper is proven and the consumer is not:
  `TestDecodeImageForRenderUnsupportedMediaType` (`image_test.go:173`) passes `"e1"` **directly to
  the helper** and so demonstrates only that the helper *can* format an element id; the consumer test
  `TestRenderUnrecognisedMediaTypeErrorsOnlyWhenDrawn` asserts nothing but `err == nil` / `err != nil`.
- **Impact**: A template author with several image elements gets an error that names an asset digest
  and a media type but not the element that caused it — the diagnostic AC11a specified, and the one
  that makes the failure actionable. It also compromises AC11c's forward promise that `Validate`
  (Story 3.7) *"reports what `Render` would do"*: 3.7 will inherit an error shape that cannot point at
  an element, and the cheapest fix there will be to build a second, differently-shaped diagnostic —
  precisely the "second rule system" D-1.8.1 (amended) forbids.
- **Suggested Resolution**: Carry the first (or every) referencing element id alongside each distinct
  asset key when building `assetKeys` — the information is already in `imageRuns` — and pass it to
  `DecodeImageForRender`. Then assert on the message in the render-path test: element id, asset key
  and media type must all appear (Finding 11 is the general form of this).
- **Related AC**: AC11a, AC11c (D-1.8.1 amended)

---

### Finding 5: AC11's mandated closed-set-absence test was not built

- **Severity**: Major
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/internal/template/closedsets.go` (the file the assertion is about);
  no test file in the repository contains the assertion
- **Observation**: AC11 is binding and has two halves: *"Mechanically: `mediaType` **must not** be
  added to `internal/template/closedsets.go`, **and a test asserts it is absent from that file's set
  inventory**."*

  The first half holds — `grep -n mediaType internal/template/closedsets.go` returns nothing, and the
  file's eight closed sets (`closedLocales`, `closedElementTypes`, `closedPageOrientations`,
  `closedPageSizeNames`, `closedAligns`, `closedValigns`, `closedBorderEdges`, `closedFooterKinds`)
  contain no media-type set. The second half was not built: a repository-wide search for any test
  referencing `closedsets`, a set inventory, or the absence of `mediaType` from it returns only two
  **non-test** comments (`image.go:96`, `parse.go:15`) that *describe* the requirement.
- **Impact**: This is the same shape as AC17b's own lesson, applied to a different invariant: the
  property currently holds because nobody has added it, not because anything prevents adding it.
  Under D-1.4.12 the cost of a future agent adding `closedMediaTypes` to that file is a permanent
  MAJOR version bump on the format — the exact one-way door AC11 exists to bar. Nothing in the tree
  would stop it or even notice.
- **Suggested Resolution**: Add a test in `internal/template` that enumerates the package-level
  `closed*` map declarations in `closedsets.go` by AST (the repo already has this pattern in
  `drift_test.go`), asserts the count is non-zero (D-000.9 witness), and asserts that no declared
  set's name or contents pertains to `mediaType`. Red-prove it by adding a `closedMediaTypes` map and
  confirming the finding names it.
- **Related AC**: AC11 (D-1.8.1, D-1.4.12)

---

### Finding 6: AC1's padding clause is asserted against a fixture with no padding

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/internal/template/base64_test.go:15-43`
  (`TestSplitBase64CanonicalWidthAndPadding`); fixture `png3x2RGB` at
  `folio-go/internal/template/image_test.go:13-19`
- **Observation**: AC1 is binding and specifies the fixture precisely: *"The test asserts, **on an
  asset whose base64 length is not a multiple of 76 and whose padding therefore lands
  mid-final-element**: … and the `=` padding characters appear **only** at the end of the last
  element."*

  `png3x2RGB` is **81 bytes**. 81 is divisible by 3, so its base64 encoding is **108 characters with
  no `=` padding whatsoever** (measured directly from the byte literal). Both padding assertions in
  the test — the `strings.Contains(p, "=")` sweep over all-but-last elements, and the whole "padding
  falls where it falls" clause — range over strings that contain no `=` and therefore cannot fail for
  any implementation.

  The reviewer applied RP-1's named mutation ("serializer strips padding into its own element") to
  `splitBase64Canonical` and ran the AC's own test:

  ```
  --- PASS: TestSplitBase64CanonicalWidthAndPadding
  ```

  It passes. The mutation is caught only incidentally, by `TestP3FixturesAreCanonical` and
  `TestP1RoundTripThroughValue`, whose fixtures happen to contain padded assets — and those fail with
  a whole-document byte diff, not with RP-1's required finding ("names the element lengths").
- **Impact**: D-1.8.2's most easily-broken sub-clause — *"Padding … never stripped, never moved,
  never given its own element"* — has no direct coverage. The incidental P1/P3 coverage is fragile: it
  survives only as long as some unrelated fixture keeps a payload whose length is not a multiple of 3,
  and it is exactly the kind of coverage the next fixture regeneration silently removes.
- **Suggested Resolution**: Either add a second sub-case to this test using a payload whose length is
  **not** a multiple of 3 (so padding exists) and whose base64 length is not a multiple of 76 (so
  padding lands mid-final-element) — one and two padding characters are both worth covering — or
  parameterise the test over three payloads (0, 1, 2 padding characters). Then re-run RP-1 and record
  the observed red.
- **Related AC**: AC1 (D-1.8.2)

---

### Finding 7: D-000.14 (extended) is violated in the same commit that records it — the Delivery Log's column split contradicts the shipped artifact

- **Severity**: Major
- **Category**: Convention / AC Conformance
- **Location**: story file, Delivery Log Amendment 4, line 1384; artifacts
  `folio-go/testdata/template/golden/asset-example.json:4-5` and
  `_bmad-output/specs/spec-folio/folio-format.md:209-210`
- **Observation**: Amendment 4 states the new example is *"canonically wrapped at **76+24**
  columns"*. Measured against the shipped bytes — decoded, re-encoded and re-split independently by
  the reviewer with a second tool — the split is **76+36**:

  ```
  key: a31f60866b4aa41953176fee9ddb90dc9bc53dce174421f8f567fac364c8bc27  (len 64)
  element lengths: [76, 36]
  digest matches key: yes
  canonical (re-split from decoded bytes): yes, lengths [76, 36]
  IHDR w=2 h=2 bitdepth=8 colortype=2 interlace=0
  chunk IHDR crc OK / IDAT crc OK / IEND crc OK
  ```

  The "24" is carried over from the story creator's **candidate** image at story line 655 (*"base64
  length 100 → canonical split 76 + 24"*), which is a different, 75-byte image that was correctly
  **not** used — AC27a explicitly told the developer to re-derive rather than ratify that candidate.
  The developer re-derived the image correctly and then narrated the old image's numbers beside it.
- **Impact**: D-000.14 (extended) — added to the decision log **in this story's own uncommitted
  diff** — states: *"A disposition table's total is computed, or omitted entirely — never narrated
  alongside it. Same for any count stated beside the artifact it summarises."* This is that failure,
  one section away from the ruling. The log itself records three prior instances (D-1.7.1's parameter
  counts, the Story 1.7 disposition totals, D-1.8.7's citation); this is the fourth, and the first to
  occur *after* the rule was written down.
- **Suggested Resolution**: Correct the figure to 76+36, or — preferably, and in the spirit of the
  ruling — delete the narrated split entirely and let `TestAssetExampleMatchesGoldenFragment` be the
  only statement of the example's shape. Consider extending the same treatment to the other
  amendment entries, none of which need a narrated length.
- **Related AC**: AC32, AC28 (D-000.14 extended, D-000.6)

---

### Finding 8: AC27b's mandated choice is never stated in the Delivery Log

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: story file, Delivery Log lines 1346-1443 (the omission); the artifact that reveals the
  choice is `folio-go/testdata/template/golden/asset-example.json`
- **Observation**: AC27b is binding: *"Either compare it to a golden fragment emitted by the shipped
  assets writer, or embed the same asset in a full canonical golden and extract the sub-object.
  **Pick one and say which in the Delivery Log** — both satisfy the ruling; silently doing neither
  does not."*

  The Delivery Log never states the choice. The phrase "golden fragment" appears nowhere in it; it
  appears only in the File List (line 1313). The nearest text, Amendment 4, line 1387, names the test
  (`TestAssetExampleMatchesGoldenFragment`) but does not present it as a selection between the two
  options. The choice is inferable from the test name and confirmed by the artifact — option 1, a
  standalone `"assets": {…}` fragment — but AC27b asked for it to be *stated*, not inferable.
- **Impact**: Low in itself; the ruling's substance was discharged. But the AC exists because a later
  reader needs to know which of two valid mechanisms is in force before extending it, and the record
  does not say.
- **Suggested Resolution**: Add one sentence to the Delivery Log naming option 1 and why.
- **Related AC**: AC27b (D-1.8.7)

---

### Finding 9: AC10a's R4 citation is absent from the code

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: `folio-go/internal/template/image.go:58-64` (the `Stream` field comment, closest thing
  to the required comment); `folio-go/internal/pdf/textdoc.go:330-331` (where `/Filter` is actually
  emitted)
- **Observation**: AC10a requires *"A comment at the `/FlateDecode` emission site [that] states that
  the bytes are the file's own IDAT stream, **cites D-1.8.1 and R4**, and states that substituting a
  re-encode is a hash-changing versioned event under AD-22."*

  The `Stream` field comment covers three of the four elements — it says the bytes are the file's own
  IDAT stream, cites D-1.8.1 and AC10a, and names AD-22 — but **does not cite R4**. A repository-wide
  search for `R4` across `folio-go/` non-test source returns no match; the only occurrence in the
  module is a test comment at `render_test.go:672`, and the only other in the tree is the guard at
  `lint/internal/rules/nocompressor.go:16`. `internal/pdf/textdoc.go`'s `ImageXObject` comment, which
  sits closer to the literal `/Filter` emission, cites neither R4 nor AD-22.
- **Impact**: AC10a is illustrative, so the substance (no compressor invoked) is intact and guarded.
  But the AC's stated purpose is that *"a later agent 'simplifying' the chunk walker into `image/png`
  + `compress/zlib` has nothing to read"* — and R4 is the risk id that agent would need in order to
  find `acceptance.md:83` and understand why the ban exists.
- **Suggested Resolution**: Add "carried risk **R4**" to the `Stream` field comment, and add a
  one-line pointer at the `/Filter` write site in `textdoc.go` back to `image.go`'s explanation.
- **Related AC**: AC10a (D-1.8.1)

---

### Finding 10: the doc/golden byte-identity pair can still rot together

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/template/goldenfixture_test.go:74-83`
  (`TestAssetExampleMatchesGoldenFragment`)
- **Observation**: The assertion is `fence != string(golden)`. It proves the document and the golden
  agree with **each other**; it proves nothing about either being canonical, correctly keyed, or a
  supported image. If a future edit changes both consistently — the natural way anyone would "update
  the example" — the test stays green with a non-canonical wrapping, a wrong digest, or an alpha PNG
  back in the spec.

  Today the values are correct (independently verified: see Finding 7's measurement block), and the
  same asset happens to appear in `passthroughLevelBase`, where the real parser validates it — but
  nothing asserts that coupling, and the two could diverge at any time.

  D-1.8.7's own closing words are *"Every fenced block claiming to be complete or canonical should be
  **executed, not read**."* This mechanism reads two files and compares them.
- **Impact**: AC27b promises *"after this the doc example cannot rot again."* It cannot **drift from
  its golden** again; it can still rot, in exactly the three ways D-1.8.7 catalogued.
- **Suggested Resolution**: Extend the test by a few lines that *execute* the golden fragment:
  concatenate its `data` array, `base64.StdEncoding.Strict().DecodeString` it, assert the key equals
  `sha256HexOf(decoded)`, assert `splitBase64Canonical(decoded)` reproduces the array exactly, and
  assert `decodeRecognisedImage("image/png", decoded)` returns `recognised == true` with no error.
  All four helpers already exist in the package. Red-prove by de-canonicalising the golden's wrapping.
- **Related AC**: AC27b, AC27c (D-1.8.7, D-1.4.10)

---

### Finding 11: several shipped red-proofs assert only `err != nil`, contrary to D-000.13

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/render_image_test.go:240-242`
  (`TestRenderUnrecognisedMediaTypeErrorsOnlyWhenDrawn`);
  `folio-go/internal/template/image_test.go:116-122` (`TestDecodePNGRejectsTruncation`);
  `folio-go/internal/template/image_test.go:126-136` (`TestDecodePNGRejectsZeroWidth`)
- **Observation**: AC30 is binding: *"A red-proof asserts on the rule id and the finding message,
  never on exit status or mere failure."* These three assert only that an error is non-nil. The
  register's own "Required finding" column for the corresponding rows says otherwise — RP-11 requires
  the finding to *"redden naming the wrongly-refused document"*, RP-9 requires *"a located load error,
  and explicitly not a panic"*.
- **Impact**: Directly causal for two other findings in this review. Finding 4 (missing element id)
  would have been caught by a message assertion in the render test; Finding 1 (the panic) sits one
  branch away from `TestDecodePNGRejectsTruncation`, which passes on a bare non-nil error and never
  checks that the failure is a *located* error rather than any error.
- **Suggested Resolution**: Assert the substrings each AC names — element id, asset key, media type
  for RP-11; `assets.<key>` locatedness for RP-9/RP-12 — and, where a panic is the forbidden outcome,
  assert it explicitly with a `defer recover()` that fails on a non-nil recover value rather than
  relying on the test binary crashing.
- **Related AC**: AC30 (D-000.13)

---

### Finding 12: AC28's before/after is abbreviated rather than verbatim

- **Severity**: Minor
- **Category**: Convention
- **Location**: story file, Delivery Log lines 1362, 1369, 1376 (Amendments 1-3)
- **Observation**: AC28 is binding: *"the before/after of every amendment is quoted **verbatim** in
  the Delivery Log."* The three fixture amendments record their prior keys as `key 64×'b'`,
  `64×'a'` and `62×'c'` rather than the literal strings that appear in the diff. The descriptions are
  accurate, and Amendment 3's injection-anchor change *is* quoted verbatim and matches
  `passthrough_test.go:125` exactly.
- **Impact**: Low. The abbreviation is unambiguous and re-derivable. It is flagged because AC28 exists
  so that the record can be checked against the diff without re-running anything, and because
  abbreviating a measured value beside the artifact is adjacent to Finding 7's failure mode.
- **Suggested Resolution**: Either quote the literals or state explicitly that the keys are recorded
  in shorthand and why.
- **Related AC**: AC28 (D-000.6)

---

### Finding 13: the flip guard's vacuity witness is weaker than its own failure message claims

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/pdf/flip_test.go:174-178`
- **Observation**: The check is `if candidates == 0`, but its message asserts a stronger property:

  > *"zero candidate Y-coordinate emission sites considered (D-000.9) — textdoc.go's
  > `buildTextContentStream` and imagedoc.go's `appendImageContentStream` **must both contribute at
  > least one candidate each** for this guard to mean anything"*

  Measured at HEAD the scanner reports `candidateEmissionSites = 2`. Under Finding 3's bypass (b),
  the reviewer measured it drop to **1** — the image path contributes nothing — and the guard still
  passes, because 1 ≠ 0. Had the code enforced what its message states, bypasses (b) and (c) in
  Finding 3 would both have reddened here.
- **Impact**: The witness that exists to catch exactly this regression does not catch it. This is the
  cheapest available partial mitigation for Finding 3 and it was written down but not implemented.
- **Suggested Resolution**: Track candidates **per file** and require at least one from each of
  `textdoc.go` and `imagedoc.go` (or, more durably, from every file in the package that emits into a
  content stream), failing with the file names that contributed none. This does not replace Finding
  3's fix but would have caught two of its three bypasses.
- **Related AC**: AC17a, AC30 (D-000.9, D-1.8.10)

---

### Finding 14: `PredictorColumns` is populated from the raw header value, not the validated one

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/internal/template/image.go:260` — `PredictorColumns: width,` inside the
  return that otherwise uses `width: w, height: h`
- **Observation**: `validateImageDimensions` returns validated copies `w, h`, and the struct's
  unexported fields correctly take those. `PredictorColumns` takes `width` — the raw `int64` read
  from IHDR — instead of `w`. The two are equal today because `validateImageDimensions` returns its
  input unchanged on success, so there is no behavioural difference.
- **Impact**: None currently. It is flagged because it is precisely the shape D-1.5.2 names: *"A field
  that can be read before it is checked will eventually be read before it is checked."* If validation
  ever clamps or normalises, this line silently diverges from the validated dimensions and emits a
  `/Columns` that disagrees with `/Width`.
- **Suggested Resolution**: Use `w`.
- **Related AC**: AC15a (D-1.5.2)

---

### Finding 15: the format specification now carries implementation archaeology

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `_bmad-output/specs/spec-folio/folio-format.md:219-227` (the paragraph added between
  the fence and the sentence that explains it)
- **Observation**: The amendment inserts a paragraph naming a story ("Story 1.8"), two decision-log
  ruling ids ("D-1.8.7", "D-1.8.1"), and a Go test file path
  ("`folio-go/internal/template/goldenfixture_test.go`'s `TestAssetExampleMatchesGoldenFragment`")
  into a document that, by its own opening, is *"a public contract, not an implementation detail"* —
  the very sentence D-1.8.1 (amended) leans on. It is also placed between the example and the
  sentence beginning *"Keyed by the lowercase hex SHA-256 of the raw bytes…"*, which explains that
  example, separating them.
- **Impact**: Cosmetic. D-000.6 requires the document be *amended*; it does not ask for the amendment
  to be narrated inside the contract. A downstream implementer reading the format spec now reads
  about our internal ruling ids and our Go test names.
- **Suggested Resolution**: Move the archaeology to the Delivery Log (where AC28 already wants it) and
  keep at most a one-line note in the spec, without ruling ids or test paths — or place it after the
  explanatory sentence so the example and its explanation stay adjacent.
- **Related AC**: AC27, AC28 (D-000.6, D-1.4.7)

---

### Finding 16: the golden fragment is hand-de-indented serializer output

- **Severity**: Nit
- **Category**: Convention
- **Location**: `folio-go/testdata/template/golden/asset-example.json`
- **Observation**: AC27a requires the replacement be *"generated by the shipped serializer … build the
  document in Go, serialize it with the shipped serializer, paste **that** output into the doc."* The
  reviewer ran `SerializeDocument` on a document carrying exactly this asset and sliced out the
  `"assets"` object. The payload — key, both base64 elements, `mediaType` — is **byte-identical**. The
  only difference is indentation: the serializer emits the object nested at document depth (4-space
  inner indent), while the golden is de-indented by two spaces to stand alone as a fence.
- **Impact**: Negligible; the part that can rot is verbatim. Recorded because it means a hand edit
  sits between the serializer and the golden, which the AC's wording did not contemplate, and because
  it interacts with Finding 10 (nothing re-derives either side).
- **Suggested Resolution**: Note the de-indent in the Delivery Log alongside Finding 8's option
  statement, or have the golden generated by a helper that performs the de-indent deterministically.
- **Related AC**: AC27a (D-1.8.7, D-1.4.10)

---

### Explicitly checked and found satisfied

Recorded so a later reader knows these were exercised rather than skipped.

- **AC2, AC3, AC3a** — any-wrapping accept and 64→76 normalisation. RP-2 reproduced exactly: with
  `writeAssets` echoing `a.Data`, `TestHandWrappedAssetNormalisesToCanonical` reddens and
  `TestP3FixturesAreCanonical` stays green; it is the **only** test in the module that reddens. The
  discriminating fixture genuinely discriminates.
- **AC4** — invalid base64, embedded whitespace and empty decoded data are strict load errors.
- **AC5, AC6, AC6a** — key = SHA-256 of decoded bytes; shape and value produce two distinct,
  mutually-exclusive messages, cross-asserted in both directions.
- **AC7** — `/Width`, `/Height` emitted via `b.writeInt`.
- **AC8, AC8a** — orphan preservation red-proved by the reviewer (dropping unreferenced assets
  reddens `TestParseDocumentPreservesOrphanedAsset`, `TestP1RoundTripThroughValue`,
  `TestP3FixturesAreCanonical`, `TestHandWrappedAssetNormalisesToCanonical`,
  `TestPassthroughAtEveryObjectLevel`, `TestDriftASTMatchesRuntimeEmission`); the Epics 5/6 corollary
  is present verbatim on the assets writer.
- **AC9** — all five refusal classes verified against generated fixtures, each a located load error
  naming the asset key and the reason: truecolour+alpha (type 6), grayscale+alpha (type 4), palette
  (type 3), 16-bit, interlaced. The AC11b contrast case (`image/png` declared, non-PNG bytes) is a
  load error; the AC11a case (`image/webp`) loads. Both polarities confirmed side by side.
- **AC10** — the compressor ban is real and red-proves against a **real** import in a **real** file:
  adding `compress/zlib` and `image/png` to `internal/template/image.go` produced
  `no-compressor-import` and `no-image-decoder-import` findings naming the file, line and path. The
  embedding path was separately verified to be true passthrough: the rendered PDF contains the source
  file's own concatenated IDAT byte sequence verbatim (366 bytes, matched byte-for-byte against an
  independently extracted copy).
- **AC13** — cross-multiplication genuinely covered: inverting the comparison reddens
  `TestResolveImagePlacementNeverStretchesNeverCrops` and `TestRenderMatchesImageEmbedGoldenFixture`.
- **AC15, AC15b** — `W == 0` and a dimension above the bound both produce located load errors, not
  panics. (The *adjacent* corrupt-header case does panic — Finding 1.)
- **AC18a** — `pdfNameEscape` is the identity on a 64-hex key.
- **AC19, AC20, AC21** — no new fetch guard added; the mutate-the-bytes behavioural check genuinely
  asserts the rendered bytes differ; no added line anywhere in the diff (including every new
  untracked file) upgrades the `net`/render-file fence's description to a proof, guarantee or
  security boundary.
- **AC22, AC23, AC23a** — three documents, not two; the `/Subtype /Image` guard verified to fire on
  the **first** leg before any hash is computed, by rendering the subprocess document with its image
  element removed; the `FontFile2` guard retained per-document.
- **AC24** — the full four-target matrix was **re-run by the reviewer** over all three documents:
  exit 0, every target byte-identical, `image-embed` matching its recorded golden.
- **AC25, AC25a** — fixture directory complete (`expected.json` with `folioGoVersion`, `goToolchain`,
  64-hex `sha256`; `expected.pdf`; `input.folio`; `README.md`); `goToolchain` equals the other two
  fixtures' (`go1.26.0`); `input.folio` byte-identity to the Go constant is asserted and fires before
  the hash comparison; all four workflow render jobs upload the new hash artifact and the compare
  job's explicit document list gained the new slug.
- **AC26, AC26a** — all three fixtures rebuilt on real, supported, non-alpha PNGs under real keys;
  non-asset content diffed and found unchanged except the necessary element `asset` reference in
  `maximalFixture` and the passthrough injection anchor, which moves within the same object and
  preserves the nesting level the level-parameterised test walks.
- **AC27, AC27a** — the replacement example independently re-verified with a second tool (see
  Finding 7's measurement block): real digest, canonical split, colour type 2, valid CRCs.
- **AC29** — four drift tests pass after the amendment, each with a live, non-zero, fail-on-zero
  witness.
- **AC31, AC32, AC33** — `-count=1` throughout; no bytes or hashes through a pipe; the no-float and
  forbidden-import guards are green over the new files.
- **Item 9 checks** — `sprint-status.yaml` changes exactly one value (`backlog` → `review`) on the
  1.8 key and nothing else; the opener is **348 words** of prose (within 150-350) and states the
  alpha scope fence plainly (*"Deliberately not supported: pictures with transparency… such a file is
  refused with a message naming the picture and the reason, never drawn wrongly"*); `folio.Validate`
  was **not** created; the licence manifest was **not** cheap-fixed (`git diff 54a0777 -- lint/` is
  empty; `fontExtensions` untouched; the only `lint/` changes are the two new nocompressor files);
  `folio-go`'s `go.mod`/`go.sum` are unchanged and `go list -m all` matches baseline.

---

### Note on scope

Three items were observed and are **not** filed as defects of this story:

1. `writeAssets` panics on invalid base64 (`serialize.go:409`). `SerializeDocument` is not reachable
   from the `folio` root package's public surface today, so this is an internal invariant guard of
   the same shape as `ScaleRound`'s, consistent with D-1.5.2's precedent and documented as such. It
   becomes reachable the moment a public save/serialize entry point ships (Epics 5/6) and should be
   revisited then.
2. The `goToolchain` fatal in `TestRenderMatchesImageEmbedGoldenFixture` runs **before** the hash
   comparison, so on a machine with a different toolchain the golden assertion never executes. This
   is inherited verbatim from the two existing fixtures' tests and is not a change this story made.
3. `hashmatrix`'s `go build ./...` exit-code observation is already recorded in the Completion Notes,
   is pre-existing, and `git status` on `hashmatrix/` is clean.

---

## Finding Resolutions

**Status: all 16 findings are FIX.** None dismissed, none deferred. The review's own findings were
all legitimate — each names a real gap between what an AC/ruling requires and what the shipped code
or record demonstrated — and every suggested resolution was in-scope, minimal, and did not require
expanding this story beyond its own ACs. The licence-manifest coverage gap (M-11, D-1.8.11) is a
pre-existing, already-deferred Epic 1 boundary item, not a review finding, and stays out per D-1.8.11
and the owner's explicit instruction.

| # | Finding | Severity | Decision | Rationale | Changed |
|---|---|---|---|---|---|
| 1 | `decodePNG` panics on a malformed PNG, reachable from the public loader | Blocker | FIX | Reproduced live (16-byte input through `folio.ParseTemplate`, `slice bounds out of range` panic before the fix). A genuine unsigned-underflow bug on the untrusted-input boundary, exactly the class D-1.5.2/AC15 exist to forbid. | `internal/template/image.go` (bounds check rewritten in `int`); `internal/template/image_test.go` (`TestDecodePNGRejectsTruncationInHeader`, message assertion added to `TestDecodePNGRejectsTruncation`); `template_image_load_test.go` (new, public-API reproduction) |
| 2 | AC14's centring-through-`ScaleRound` is entirely unproven | Blocker | FIX | Confirmed live: RP-6 (both `ScaleRound` calls → `/2`) left the whole module green pre-fix. The AC's own "odd difference" wording is insufficient (measured: agrees with truncation except when ≡3 mod 4) and its fixture had differences 0/33334, neither odd. | `render_image_test.go` (`TestResolveImagePlacementOddDifferenceUsesScaleRound` rebuilt with literal expectations and a discriminating fixture); AC14 corrected in this file's own Acceptance Criteria §D |
| 3 | The AD-24 flip guard is a naming-convention check | Blocker | FIX | Confirmed live: two of three real bypass shapes (differently-named variable, inline expression) passed the shipped guard silently and were not counted as candidates, against a real second flip open-coded into `imagedoc.go`. | `internal/pdf/flip_test.go` (guard rebuilt operand-shape-independent by AST; RP-7 rebuilt to mutate the real `imagedoc.go`) |
| 4 | `Render`'s unrecognised-`mediaType` error does not name the element id | Major | FIX | Confirmed live: message rendered as `element : asset <key>` — a visible hole. D-1.8.1 (amended)'s verdict table is binding on this point. | `render.go` (`firstElementIDByAssetKey` carries the id through); `internal/template/image_test.go` and `render_image_test.go` (both tests now assert the id appears) |
| 5 | AC11's mandated closed-set-absence test was not built | Major | FIX | Confirmed: no test anywhere asserted `mediaType`'s absence from `closedsets.go`'s inventory — the property held only because nobody had added it. AC11's second half is binding. | `internal/template/closedsets_test.go` (new: AST enumeration + red-proof) |
| 6 | AC1's padding clause is asserted against a fixture with no padding | Major | FIX | Confirmed live: RP-1's named mutation left the shipped test green (81-byte fixture, zero `=` characters, every padding assertion vacuous). | `internal/template/base64_test.go` (table-driven over 0/1/2 padding characters) |
| 7 | D-000.14 (extended) violated in the same commit that records it | Major | FIX | Confirmed by independent re-derivation: the shipped bytes split 76+36, the Delivery Log said 76+24 (the rejected candidate image's numbers, left behind). | Story file, Delivery Log Amendment 4 (figure corrected, drift class recorded as the 4th instance); `internal/template/goldenfixture_test.go` (Finding 10's fix also closes this at the root) |
| 8 | AC27b's mandated choice is never stated in the Delivery Log | Minor | FIX | Confirmed: the choice (a standalone golden fragment, not a full-golden extraction) was inferable but never stated, contrary to AC27b's explicit "pick one and say which." | Story file, Delivery Log (statement added beside Amendment 4) |
| 9 | AC10a's R4 citation is absent from the code | Minor | FIX | Confirmed: `grep -rn R4 folio-go/ --include=*.go` (non-test) found nothing before this fix. AC10a is illustrative but cheap and directly useful to a future reader. | `internal/template/image.go` (`Stream` field comment); `internal/pdf/textdoc.go` (`ImageXObject` comment) |
| 10 | The doc/golden byte-identity pair can still rot together | Minor | FIX | Confirmed live: de-canonicalising both the doc's fence and the golden identically (preserving their mutual equality) left the shipped test green. D-1.8.7's own standard is "executed, not read." | `internal/template/goldenfixture_test.go` (`executeAssetExampleFragment`: decodes, re-derives the key and canonical split, confirms `image/png` recognises it) |
| 11 | Several shipped red-proofs assert only `err != nil` | Minor | FIX | Confirmed at all three named locations. Directly causal for Findings 1 and 4 escaping review. | `render_image_test.go`, `internal/template/image_test.go` (message assertions added at all three named sites) |
| 12 | AC28's before/after is abbreviated rather than verbatim | Minor | FIX (documented, not literal quoting) | Confirmed: Amendments 1–3 use shorthand (`64×'b'`, etc.) rather than literal 64-character strings. The literal placeholder strings are not recoverable from git (this file was never committed at baseline) and fabricating one would violate D-000.12/13; AC28's own second option — state the shorthand explicitly and why — is used instead. | Story file, Delivery Log (explanatory paragraph added before Amendment 1) |
| 13 | The flip guard's vacuity witness is weaker than its own failure message claims | Minor | FIX | Confirmed live: under bypass (b), `candidateEmissionSites` dropped 2→1 and the old guard still passed (1 ≠ 0). Superseded by Finding 3's rebuild, which enforces the per-file requirement literally. | `internal/pdf/flip_test.go` (same change as Finding 3) |
| 14 | `PredictorColumns` is populated from the raw header value, not the validated one | Nit | FIX | Confirmed by inspection; harmless today (validation is currently identity on success) but exactly the shape D-1.5.2 warns will eventually diverge. One-line, zero-risk fix. | `internal/template/image.go` (`PredictorColumns: w` instead of `width`) |
| 15 | The format specification now carries implementation archaeology | Nit | FIX | Confirmed: the inserted paragraph named a story number, two ruling ids and a Go test path inside a document whose own opening calls itself a public contract, and sat between the example and its explanation. | `_bmad-output/specs/spec-folio/folio-format.md` (paragraph moved to the Delivery Log as Amendment 5, replaced with a one-line, ruling-id-free note placed after the explanation) |
| 16 | The golden fragment is hand-de-indented serializer output | Nit | FIX | Confirmed: only indentation differs from real serializer output; the payload is byte-identical. Negligible risk, but now recorded per the suggested resolution. | Story file, Delivery Log (noted alongside Finding 8's AC27b statement) |

### Corrected ruling disposition table (D-000.10 mirror, post-fix)

The developer's original table claimed **32/32** "Applied as stated." The reviewer's independent
mirror computed **18 agree / 14 disagree** (rows 1, 2, 3, 5, 8, 11, 14, 15, 19, 20, 26, 29, 30, 31
below). Every one of the 16 underlying findings driving those 14 disagreements is FIX (table above),
and each fix was independently verified (live reproduction, red-proof re-run, or direct
re-measurement) rather than assumed closed by the code change alone. Re-derived post-fix:

| # | Ruling | Post-fix disposition | Agree? |
|---|---|---|---|
| 1 | **D-1.8.1** | Passthrough real (unchanged from the reviewer's verification). R4 now cited in `image.go`'s `Stream` comment and `textdoc.go`'s `ImageXObject` comment (Finding 9, RESOLVED). | ✅ |
| 2 | **D-1.8.1 (amended)** | Surfaces 1/2 unchanged. `Render`'s error now names the element id (Finding 4, RESOLVED — re-verified: `TestRenderUnrecognisedMediaTypeErrorsOnlyWhenDrawn` now asserts on `"e1"`, the key and `"image/webp"`). | ✅ |
| 3 | **D-1.8.2** | 76-column split etc. unchanged. AC1's padding clause now has a genuinely discriminating fixture (Finding 6, RESOLVED — RP-1 re-run, reddens on the 1/2-padding cases). | ✅ |
| 4 | **D-1.8.3** | Unchanged — no finding against this row. | ✅ |
| 5 | **D-1.8.4** | Cross-multiplication unchanged. Centring now genuinely proven (Finding 2, RESOLVED — RP-6 re-run, reddens naming the offset value) and the panic on a malformed header is fixed (Finding 1, RESOLVED). | ✅ |
| 6 | **D-1.8.5** | Unchanged — no finding against this row. | ✅ |
| 7 | **D-1.8.6** | Unchanged — no finding against this row. | ✅ |
| 8 | **D-1.8.7** | Replacement example verification unchanged. Golden now executed, not just compared (Finding 10, RESOLVED); AC27b's choice now stated (Finding 8, RESOLVED); de-indent noted (Finding 16, RESOLVED); Amendment 4's figure corrected 76+24→76+36 (Finding 7, RESOLVED). | ✅ |
| 9 | **D-1.8.8** | Unchanged — no finding against this row. | ✅ |
| 10 | **D-1.8.9** | Unchanged — no finding against this row (RP-2 confirmed sound by the reviewer; not re-litigated). | ✅ |
| 11 | **D-1.8.10** | Guard rebuilt operand-shape-independent (Finding 3, RESOLVED — all three bypass shapes now redden against the real `imagedoc.go`). | ✅ |
| 12 | **D-1.8.11** | Unchanged — `git diff 54a0777 -- lint/` still touches only the two nocompressor files; coverage-witness fix confirmed still out of scope (D-1.8.11, owner instruction). | ✅ |
| 13 | **D-000.4** | Re-run again by the finisher after all fixes: exit 0, three documents × four targets, all byte-identical, matching recorded goldens. | ✅ |
| 14 | **D-000.6** | All amendments verbatim where recoverable; shorthand explained where not (Finding 12, RESOLVED); Amendment 4's figure corrected (Finding 7, RESOLVED); a fifth amendment added for the spec archaeology move (Finding 15, RESOLVED). | ✅ |
| 15 | **D-000.9 (+ ext)** | No-compressor/image-XObject witnesses unchanged. Flip guard's vacuity witness now enforces its own stated per-file requirement literally (Finding 13, RESOLVED — same fix as Finding 3). | ✅ |
| 16 | **D-000.10** | Unchanged — no finding against this row; this table itself is the row's own re-derivation. | ✅ |
| 17 | **D-000.11** | Unchanged — no finding against this row; all finisher gates also re-run with `-count=1`. | ✅ |
| 18 | **D-000.12** | Unchanged — no finding against this row; no bytes/hashes/panic stacks taken through a pipe during finishing either (mutations applied to scratch copies via `cp`/Python/`sed`, restored and `diff`-verified clean). | ✅ |
| 19 | **D-000.13** | The three named red-proofs now assert on message content (Finding 11, RESOLVED); this closes the gap that let Findings 1 and 4 through review undetected. | ✅ |
| 20 | **D-000.14 (+ ext)** | Amendment 4's narrated figure corrected to match its artifact (Finding 7, RESOLVED); this table's own total is stated as a computed count, never narrated beside it, per the same rule. | ✅ |
| 21 | **D-1.1.b** | Unchanged — no finding against this row. | ✅ |
| 22 | **D-1.2.6** | Unchanged — no finding against this row. | ✅ |
| 23 | **D-1.3.1** | Unchanged — no finding against this row; re-verified: the two new test files (`template_image_load_test.go`, `closedsets_test.go`) are test files, outside D-1.3.1's non-test-import scope, and `closedsets_test.go` imports only stdlib (`go/ast`, `go/parser`, `go/token`, `os`, `path/filepath`, `strings`, `testing`). | ✅ |
| 24 | **D-1.3.5** | Unchanged — no finding against this row; `closedsets_test.go` ranges a map, but D-1.3.5 exempts `_test.go` files explicitly (verified against `lint/internal/rules/maprange.go`'s own `Tests: false`). | ✅ |
| 25 | **D-1.4.3** | Unchanged — no finding against this row. | ✅ |
| 26 | **D-1.4.5** | Now guarded, not merely true by accident (Finding 2, RESOLVED — same fix; substituting `/2` now fails the suite). | ✅ |
| 27 | **D-1.4.7** | Unchanged — no finding against this row. | ✅ |
| 28 | **D-1.4.9** | Unchanged — no finding against this row. | ✅ |
| 29 | **D-1.4.10** | Attribution/independence unchanged. The doc/golden pair can no longer rot together undetected (Finding 10, RESOLVED — same fix). | ✅ |
| 30 | **D-1.4.12** | `mediaType`'s absence from `closedsets.go` is now mechanically tested (Finding 5, RESOLVED). | ✅ |
| 31 | **D-1.5.2** | The panic on a malformed header is fixed (Finding 1, RESOLVED) and `PredictorColumns` now uses the validated dimension (Finding 14, RESOLVED). | ✅ |
| 32 | **D-1.7.3** | Unchanged — no finding against this row. | ✅ |

**Computed from the marks above: ✅ = 32, ❌ = 0, rows = 32.** The developer's original claim of
32/32 is now true; it was not true at review time, when the reviewer's independent mirror found 14
disagreements against it. Both facts are recorded here rather than one silently superseding the
other: **claimed 32/32 → independently found 18/32 agree, 14 disagree → all 14 resolved by the
finisher, re-verified to 32/32.**
