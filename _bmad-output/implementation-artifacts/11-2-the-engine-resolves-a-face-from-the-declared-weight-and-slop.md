---
title: 'The engine resolves a face from the declared weight and slope'
type: 'feature'
created: '2026-09-06'
status: 'done'
baseline_commit: '3ad4ede13b70ec04a2524d304f927410d22b476d'
review_loop_iteration: 0
context: []
---

## In plain terms (read this first if you just want the gist)

*Not normative, and rewritten after the fact: the frozen Intent below governs implementation, while
this section says what actually shipped.*

Asking a document for bold or italic now changes what it prints. Until this story those requests were
stored, edited and shown in the browser, and read by nothing that draws — two documents differing only in
a bold flag produced identical files.

A document's font list can now say which typeface cut stands in for bold, for italic, and for both.
Nothing is guessed: the engine never assembles or picks apart a typeface name to find a heavier one. It
still chooses a typeface by which letters that typeface can draw, and only then applies the weight the
document named within that choice. Where a document asks for bold and named no bold, it prints in the
book weight it already had and says so plainly rather than quietly substituting something close.

Saving a document that uses the new form now stamps a newer format number, because older readers
genuinely cannot read it. That was a real trap caught here: the two places deciding this had drifted
apart, and one would have written a number that lied.

None of this is visible in the designer yet. The canvas still paints book weight and the family control
still offers no cuts — the next two stories do that. No printed proof of a bold page ships either, and
that is a ruling, not an omission: nobody can vouch for a bold page before one can be produced.

Review caught two genuine bugs, both about measurement rather than the new feature itself; both are
fixed.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** `style.bold` and `style.italic` are parsed, stored, serialized, edited and projected to
the browser, and read by nothing that draws. `render.go`, `table_render.go` and `wrap.go` contain zero
occurrences of either field (measured, with a positive control). Two documents differing only by
`"bold": true` render to the same SHA-256. Story 11.1 shipped seven weighted and sloped faces, so
there is now something to resolve to, and nothing resolves.

**Approach:** A font chain entry gains an **object form** carrying optional style-variant siblings — a
FontSet face name on the shipped arm, an assets key on the embedded arm — from a **closed** set of three.
Per-rune coverage keeps deciding the entry on the **base** chain; the declared variant is then applied
**within** the entry that coverage already chose. An entry with no declared variant for the requested
style renders in **its own base face** and says so with a located Warning. Nothing is inferred, parsed
or constructed from a face name.

## Boundaries & Constraints

**Always:**
- **The `(family, weight, slope) → face` mapping is DECLARED on the chain entry and DERIVED FROM
  NOTHING** — not sfnt name ID 1, not `OS/2.usWeightClass`, not `fsSelection`, not `macStyle`, not
  `post.italicAngle`. FR57 says this itself: *"resolved per rune through the **declared** chain"*.
- **Coverage decides the entry on the BASE chain; style is resolved WITHIN the chosen entry.** Do NOT
  substitute variant names into the chain before coverage runs. *A variant whose `cmap` is narrower
  than its base would push a rune to the next entry and silently change the TYPEFACE to keep the
  WEIGHT*, which is the substitution AD-8 and D-B forbid by name. This trap ships if you get it wrong.
- **A face name is NEVER PARSED and NEVER CONSTRUCTED.** `entry.Face + " Bold"` is the foreclosed
  naming-convention weight carrier written in the other direction, ruled out on identical grounds to
  `strings.TrimSuffix(key, " Bold")`. The prohibition is written in five *locations with five different
  wordings*; here it binds **construction** identically, so a literal reading cannot honour the words
  while reinstating the mechanism.
- **A sibling's namespace must match the entry's discriminant (AD-8).** A `face` entry's siblings are
  **FontSet face names**; an `asset` entry's siblings are **asset keys**. A cross-namespace sibling —
  `{"asset":"myEmbeddedRoboto","bold":"Roboto Bold"}` — is exactly the substitution AD-8 forbids,
  arriving *inside a single entry where no precedence rule can see it*. **Located load error, both
  directions asserted.**
- **The variant key set is CLOSED: exactly `bold`, `italic`, `boldItalic`.** An open sub-object
  (`{"face":"X","variants":{…}}`) was considered and **rejected**: it surrenders the unknown-key refusal
  that is the whole property this change exists to preserve. **Extending the set later is a MAJOR
  change, and the format doc must say so.**
- **The one-key rule was the MECHANISM; the property is that an unknown key cannot ride along disguised
  as a decoration.** An exactly-one-of discriminant over a closed set preserves that property. This is a
  mechanism change in service of the invariant, not a relaxation of it.
- **Absence is a first-class result, never a nil.** An entry with no declared variant for the requested
  style resolves to **its own base face** and emits AC3's diagnostic. *"Nearest available face"* means
  **this entry's** base face — never a walk down the chain hunting for something bold.
- **One answer site: `chainFaceNames`, not `fontChain`.** Widening the six
  `(chain []string, FontSet, *fontCache)` signatures is rejected again.
- **AD-21 / AC5.** A document declaring neither bold nor italic must hash identically on all four
  targets. Bare-string chain entries must still serialise as bare strings.
- **AD-14 / I-8.** AC3's diagnostic is one type on one channel, from the closed **additive** registry,
  naming the element, the rune and the face.
- **AD-26 / I-7.** A variant **asset key** must clear the embedded-face licence requirement exactly as
  the `asset` value does. A sibling that skips it lets a document carry an unlicensed embedded bold.
- **I-5, no migration.** A document whose chain entries predate this story is not migrated, rewritten
  or repaired on open. It reports honestly.

**Ask First:**
- Any change to `folio.FontSet`'s shape (`folio-go/fontset.go:20`). D-B forecloses it.
- Any new **command kind**, or any widening of an existing component command's arity. The plan gate
  measured that surface; it belongs to Story 11.4 and is explicitly out of scope here.
- Projecting a variant into `CanvasFontChainEntry`. That is Story 11.3's call, and the exact-key guard
  makes it a blank-canvas hazard (see Code Map).
- Any change that moves a golden digest, or that changes the count or the set of shipped faces.
- **Any MAJOR version bump.** Ruled: none. See Design Notes.
- Anything outside this spec's Tasks: a file the tasks do not name, an adjacent bug, a refactor, a
  guard for a case no acceptance criterion demonstrates, or a new dependency.

**Never:**
- **Never commit, `git add`, stash, checkout, reset, revert, or restore.** The orchestrator makes every
  commit. Leave all work in the working tree and report what to commit. Reading git state is fine.
- **Never create a branch. Never push. Never open a pull request.**
- **Never `git add -A`.**
- Never synthesize bold or oblique, at emit time or anywhere (I-2). `SetSyntheticBold` /
  `SetSyntheticSlant` are exposed by the vendor and called nowhere; keep it that way.
- Never build DW-233's repo-wide `lint` rule. D-11.2.1 supersedes the register's stated discharge: the
  prohibition is discharged **behaviourally** by AC3's own test. The lint rule stays deferred.
- Never touch the canvas or the inspector. 11.3 paints; 11.4 teaches the family control.
- Never touch `folio-designer/public/templates/starter.folio`. That is 11.3's, by owner ruling D-11.0.1.
- Never widen the canvas fragment fallback stack (counts move, the stack does not).
- Never make the engine write `bold: false`. Off is absent, not false (designer rule, commit `3bacfec`).
- **Never pin a predicate against itself.** The version/serializer tie is proved behaviourally, output
  against output — never by asserting the shared predicate equals itself.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| No style declared | Element with no `bold`/`italic`; chain of bare strings | Byte-identical to today; no new diagnostic; version unchanged | N/A |
| Variant declared and used | `{"face":"Roboto","bold":"Roboto Bold"}`; element declares bold | Runes covered by Roboto shape and measure from **Roboto Bold** | N/A |
| **DW-233 tripwire** | Entry `"Roboto"` (bare, **no declared variant**); `FontSet` **does** contain `"Roboto Bold"`; element declares bold | Renders Roboto **Regular** + AC3 Warning. An implementation that constructs the name reds this. | Warning, render succeeds |
| Permanent absence (CJK) | Chain reaches `"Noto Sans SC"` (no bold cut exists); element declares bold | That rune renders in Noto Sans SC **Regular** + AC3 Warning; the Latin runs still take their declared bold | Warning, render succeeds |
| Permanent absence (Thai italic) | `"Noto Sans Thai"` (no italic cut exists); element declares italic | Noto Sans Thai **Regular** + AC3 Warning | Warning, render succeeds |
| Never walk the chain for weight | Entry A covers the rune and has no bold; entry B later has one | Renders **A's base face**, never B. Losing the weight is a smaller lie than changing the typeface. | Warning, render succeeds |
| Bold + italic together | Element declares both; entry declares `boldItalic` | Resolves to the `boldItalic` variant | N/A |
| Bold + italic, only `bold` declared | Element declares both; entry declares `bold` only | Base face + AC3 Warning — a partial match is an absence, not a nearest-fit search | Warning, render succeeds |
| Table header cascade | Table declares `style.bold`; `headerStyle.bold` also declared | `headerStyle` wins **for the header row alone**; body and footer cascade from `style` | N/A |
| Wrapping follows the face | Bold element whose bold face is wider | Breaks are the bold face's; canvas and PDF agree because both read one engine measurement | N/A |
| **Object form raises the version** | Any entry that serialises as an object (shipped **or** embedded) | `versionForSave` stamps **`2.0`** | N/A |
| **Bare string keeps the version** | Every entry a bare string | Version unchanged (`1.0` stays `1.0`) — the corpus does not move | N/A |
| **Cross-namespace sibling (embedded→shipped)** | `{"asset":"<key>","bold":"Roboto Bold"}` | Refused: an asset entry's siblings are asset keys | `fonts.<name>[<i>].bold` |
| **Cross-namespace sibling (shipped→embedded)** | `{"face":"Roboto","bold":"<assets key>"}` | Refused: a face entry's siblings are FontSet face names | `fonts.<name>[<i>].bold` |
| **Both discriminants** | `{"face":"Roboto","asset":"<key>"}` | Refused — exactly one of `face`\|`asset` | `fonts.<name>[<i>]` |
| **No discriminant** | `{"bold":"Roboto Bold"}` | Refused — a sibling cannot stand alone | `fonts.<name>[<i>]` |
| Unknown key on an entry | `{"asset":"<key>","weight":700}` | Refused — the key set stays CLOSED, no passthrough. Message **derived** from the closed set. | `fonts.<name>[<i>]` |
| Variant is empty string | `{"face":"Roboto","bold":""}` | Refused — an empty string names no face | `fonts.<name>[<i>].bold` |
| Wrong-typed variant | `{"face":"Roboto","bold":true}` | Refused, with a message saying which meaning applies **here** (a face name, not a boolean) | `fonts.<name>[<i>].bold` |
| Unlicensed variant asset | `{"asset":"<k1>","bold":"<k2>"}` where `<k2>` lacks licence fields | Refused, exactly as the `asset` value would be (AD-26) | `fonts.<name>[<i>].bold` |
| Variant names an absent FontSet face | `bold` names a face the supplied `FontSet` lacks | Existing FontSet tolerance applies: skipped, then AC3's absence arm | Warning, render succeeds |

</frozen-after-approval>

## Code Map

Every anchor **re-derived at `102e1fc`** (D-000.4). Cite by symbol; re-derive the line at the build's own
HEAD before editing.

### The resolution seam

- `folio-go/render.go:1216` `chainFaceNames(chain []template.FontChainEntry) []string` — **THE answer
  site.** **Exactly TWO non-test call sites** (measured; D-11.2.1's "three" is wrong): `render.go:1156`
  inside `fontChain`, and `table_render.go:707` inside `collectBandTableRuns`' **header** arm, which
  reaches it via `lookupFontChain` at `table_render.go:699`. ⚠ **The header arm bypasses `fontChain`
  entirely — resolution placed in `fontChain` alone silently skips every table header.** Test call
  sites: `chain_face_names_test.go:157`, `:208`.
- `folio-go/render.go:1148` `fontChain(doc, el) ([]string, error)`. Three production callers:
  `render.go:786`, `table_render.go:866`, and **`page_setup.go:1347` `addCanvasTextPaint`** — which is
  why 11.3 inherits this for free (it feeds the same `[]string` to `shapeSegments` at `:1370` and
  `chainVerticalModel` at `:1459`).
- `folio-go/render.go:1572` `resolveRuneFace(chain []string, r rune, fs, cache)` — per-rune coverage.
  **No element id in scope.** Skips a chain member absent from the FontSet (`:1574`); `("", false, nil)`
  means "no coverage" and is not an error.
- `folio-go/render.go:1692` `shapeSegments(elementID string, chain []string, …)` — **`elementID` is the
  first parameter.** AC3's Warning belongs here: zero plumbing, and a hidden element's diagnostics are
  already discarded at `render.go:842-861`.
- ⚠ **The plumbing problem.** `chainFaceNames` flattens `[]FontChainEntry` → `[]string` *before* the
  per-rune loop, so at `shapeSegments` **the entry is gone; only its name survives.** Thread a parallel
  styled-name slice (same length, same order) alongside the base chain — coverage reads the base slice,
  shaping and metrics read the styled one. Do **not** substitute into the base slice.

### The six documentless signatures — still six, and **none needs a style parameter**

`resolveRuneFace` (`render.go:1572`), `shapeSegments` (`:1692`), `digitTableRun` (`page_number.go:436`),
`chainLineMetrics` (`wrap.go:550`), `chainVerticalModel` (`:756`), `lineAdvance` (`:767`). Every one is
fed a `[]string` that came verbatim from `chainFaceNames` with no re-derivation. **AC4 is therefore
free.**
⚠ `chainFaceNames`' doc table (`render.go:1194-1213`) is **STALE**: Story 8.4 gave `formatFontChain`
(`:1613`) and `missingGlyphMessage` (`:1632`) a `*fontCache`. The split is **6/2/2, not 6/4**.

### The second mint site — fails silently

- `folio-go/embedded_face.go:250` `newEmbeddedFaceIndex` — independently computes
  `embeddedFaceName(entry.AssetKey)` (`:262`), walking only `entry.AssetKey`. **A variant asset key is
  invisible to it** → `declares` false → `resolveRuneFace` skips it → falls back with **no diagnostic**.
  ⚠ Its comment pins determinism on chains in **sorted name order** and entries in **authored order**.
  Siblings are a **third axis**: visit them in the closed set's own fixed order.

### ⚠ THE VERSION DEFECT — two copies of one predicate that only agree today

```go
// internal/template/version.go:358  fontsRequireMajor
if entry.Embedded() { return true }              // raises to 2.0
// internal/template/serialize.go:215  writeFontChain
if entry.Embedded() { dst = writeObject(...) }   // emits an object
```

They agree **only because object-form and embedded are currently the same set.** The object form
separates them: `{"face":"Roboto","bold":"Roboto Bold"}` has `AssetKey == ""`, so `Embedded()` is false,
so nothing raises the version, so `versionForSave` stamps **`1.0`** on a document **no 1.x reader can
decode**. `version.go`'s own comment says a 1.x reader *"decodes a chain entry as a string and never
coerces, so it refuses the file outright"*; `folio-format.md:86` names the failure as *"a version that
lies."* **It would ship green — nothing asserts the negative direction for a non-embedded object entry,
because there has never been one.**
**Fix: ONE predicate on `FontChainEntry` — "serialises as an object" — consumed by BOTH sites**, so they
cannot disagree. D-7.4.5 mirrored invariant: it moves in one commit, with `folio-format.md:47`.

### The cascade (AC2)

- `folio-go/table_render.go:313` `resolveHeaderStyle` — nine identical three-line switches, header arm →
  base arm → default. Called from `:670` (render) and `table_columns_projection.go:152` (canvas).
  **Copy the nine-sibling spelling exactly**: both arms test `.Set && !.Null`. `:358-368` records a live
  defect from omitting `!.Null` on the `color` arm — an explicit null won and killed the fall-through.
- `folio-go/table_render.go:438` `resolveBodyStyle` — deliberately **no header arm** (D-000.76, `:404-418`).
- ⚠ `folio-go/component_commands.go:2521` `tableHeaderStyleFields` + its comment `:2510-2520` say
  bold/italic *"have NOWHERE TO RESOLVE FROM"*. **A tripwire written for this story.** Retire by editing.

### The format half

- `folio-go/internal/template/model.go:164` `FontChainEntry{Face, AssetKey string}` — two **plain
  strings**, no tags, **no `Extra`**, and **no `Presence` field**. ⚠ D-11.2.1 §1's phrase *"using the
  existing `Presence` idiom"* was carried out of D-B unchecked and **the lead has withdrawn it**: there
  is no such idiom here. **Mirror the entry's own convention** — plain `string`, `""` means absent, empty
  refused at parse, exactly as `Face` and `AssetKey` already work. `Presence[T]` (`presence.go:16`) is
  real and used by `Style`/`Asset`, but importing it here would make this entry the only mixed-idiom
  struct in the model for no gain. Amend the `:157-163` doc comment; correct the stale five-call-sites
  claim at `:213-219` (at HEAD there are **four**).
- `folio-go/internal/template/parse.go:392` `decodeFontChainEntry` — arm chosen by the **first non-space
  byte**. `:416` `if !ok || len(entryObj) != 1` is exact cardinality. **Three refusal messages go stale
  the moment the set widens** and must be derived, not hand-written: `:407` (*"exactly one key,
  \"asset\""*), `:417` (*"…carries no other key"*), and the `default` branch (*"either a face name (a
  string) or an embedded face"*). Story 8.3's own comment says why: *"unpinned wording in a refusal is
  wording that goes stale silently and sends the author to fix the one thing that was not wrong."*
- `folio-go/internal/template/parse.go:433` `requireEmbeddedFaceLicence` — called for the `asset` value.
  **Must also run for every variant asset key.**
- `folio-go/internal/template/serialize.go:204` `writeFontChain` — decider is `entry.Embedded()` at
  `:215`. **Replace with the shared object-form predicate.** `writeObject` (`:38`) sorts keys byte-wise,
  so emission order is deterministic by construction — but state the sibling order rule anyway.
- `folio-go/internal/template/version.go:270` `fontsRequireMajor` call site; `:358` the predicate.
  `SupportedMajor` (2) / `SupportedVersion` ("2.0") at `:76-79` **do not move**.

### The diagnostic half

- `folio-go/internal/diag/diag.go` — 17 codes. Minting touches **eight** sites: the const block
  (`:66-326`), `allCodes` (`:333`), `dispositions` (`:368`), the `folio.DiagCodeX` bridge const
  (`diagnostic.go`, pattern `:177`), `diag_test.go`'s `codePins` (`:25-47`), `diag_bridge_test.go`'s
  `diagCodeBridgePins` (`:27-49`), the construction site (`diag_no_empty_code_test.go:56`), and the
  census trigger. No docs table, no lint rule, no designer enumeration, no wasm change (verified with
  positive controls).
- `folio-go/diagnostic_registry_census_test.go:154-243` — a **real production trigger** returning
  non-empty `Bytes`, `SeverityWarning`, a non-blank message and a **non-empty `ElementID` or
  `DataPath`**; cardinality asserted both ways (`:242`). Copy the fixture-precondition guard (`:44-46`).
- `folio-go/render.go:1745-1758` — the pattern AC3 copies: coalesced first-occurrence over
  `seenMissingRunes`, one Diagnostic per (element, distinct rune), a **slice with a linear scan, never a
  map** (determinism, `:1700-1709`).
- `folio-go/render.go:1613` `formatFontChain` — spells an embedded entry by **display name**, not its
  `asset:<64 hex>` render-path name. AC3's message must go through it or an equivalent.
- ⚠ `folio-go/page_number.go:437` passes `elementID = ""` and discards diags. Propagating them would
  produce an unlocated Warning that trips the census location assertion.

### Blast radius — MEASURED per test, so nobody has to guess

**Exactly ONE existing test reds from the new Warning:**
- `folio-go/component_commands_test.go:156` `TestPlacedImageStartsEmptyAndSurvivesTheRoundTripAndRender`
  — renders `componentTemplate(t)` (= `worked-example.json`) at `:197` and asserts
  `len(result.Diagnostics) != 0 → Fatal` at `:201`. **WOULD RED. Expected; update the expectation and say
  why.**

Everything else is unaffected, and the reason is structural rather than lucky:
- `component_properties_test.go` (~20 tests), `canvas_window_count_test.go`, `canvas_projection_wire_test.go`,
  `element_box_test.go`, `command_json_authority_wire_test.go` — use `Canvas`/`CanvasWithTextPaint` only,
  which **has no diagnostics channel at all** and discards them at `page_setup.go:1370`.
- `table_header_style_test.go:360`, `page_setup_test.go` (4 tests), `serialize_template_test.go:10`,
  `folio_expr_validate_test.go:383`, `goldenfixture_test.go:16`, `internal/template/roundtrip_test.go`,
  all `wasm/*` consumers — **parse/serialize/command only; never render.**
- `component_commands_test.go:205` and `browser_roundtrip_witness_test.go:187` render but **never inspect
  `Diagnostics`**. The latter also `t.Skip`s without `FOLIO_ROUNDTRIP_DIR`.
- `diagnostic_registry_census_test.go` — the Error loop returns only `error` and discards `Result`; the
  Warning loop **scans for its own code and `continue`s past extras** (`:222-234`). Extra diagnostics
  cannot red it. It reds only structurally, from registering a new code.
- **`maximalFixture` consumers are all in `internal/template`, which cannot render** — search:
  `grep -rn "func Render\|shapeSegments\|Diagnostic" folio-go/internal/template/*.go` → 1 hit, a comment
  (`fontasset.go:37`); positive control `grep -rn "func ParseDocument"` → `parse.go:53`.

**Nothing pins `worked-example.json`'s rendered bytes.** `grep -c "worked-example"` → 0 in
`byte_neutrality_test.go`, `matrix_test.go`, `statement_golden_fixture_test.go`; positive control
`grep -c "statement-1" byte_neutrality_test.go` → 5. And `grep -rn '"bold"\|"italic"' fixtures/` → zero
across all 30 fixture dirs; positive control `grep -rln '"fontFamily"' fixtures/` → many. **So no pinned
`expected.pdf` renders a bold- or italic-declaring document.**
⚠ **That asymmetry is itself a hazard:** if the resolver ever fell back to a `"<family> Bold"` name
lookup, `e1` would silently switch face, no Warning would fire, the one red test would stay green, and
**nothing in the tree pins those bytes** — so it would go unnoticed. The DW-233 tripwire is the only
thing standing there. `e1` does reach shaping (page header collected at `render.go:2151`,
`shapeSegments` at `:838`; `grep -c "visibleIf" worked-example.json` → 0, positive control on
`fixtures_test.go` → 4), and its Warning lands at **`Result.Diagnostics[0]`** (header diags concatenated
first, `render.go:2172`) — relevant to anything index-keyed.

⚠ **`worked-example.json` cannot be edited alone**: `goldenfixture_test.go:16` byte-compares it against
the `## Worked example` fence in `folio-format.md` (its `"bold": true` is at `folio-format.md:876`). Do
not strip `bold` from the fixture to dodge the red — that is the wrong repair and it moves the doc.

### Guards, and the format doc's four coupled edits

- `drift_test.go:232/:269/:402` — the drift trio. ⚠ **If the variant keys are named `bold`/`italic` these
  stay green with no doc edit at all**, because `style.bold`/`style.italic` already occupy those tokens
  (`folio-format.md:497`). That is why a targeted assertion is required (see Tasks). ⚠ Never write a
  **keyed** literal `kv{key: "x", …}` in `serialize.go` (`drift_test.go:459`).
- `folio-go/internal/template/fixtures_test.go:259` `maximalFixture` — must emit any genuinely new key.
- `fonts_embedded_test.go:620` — ⚠ `{"face": "Noto Sans"}` is pinned **today as a load error** and this
  story makes it **VALID**. Replace the row.
- `fonts_embedded_test.go:621` — `{"asset":…,"weight":700}`: **the subject has not moved, the ground
  has** (from "exactly one key" to "not in the closed set"). `requireLoadError` asserts the **field
  only**, so it cannot see the change.
- `render_arch_test.go:326` — matches **method names package-wide regardless of receiver**. Prefer a
  **free function**; `TestFolioMethodNamesAreInjective` also forbids a second root-file error type.
- `font_cache_sites_test.go` — fixes the **two** `newDocumentFontCache` sites (`render.go:2043`,
  `page_setup.go:1312`). Do not add a third.
- **`_bmad-output/specs/spec-folio/folio-format.md` — FOUR edits that move together (D-7.4.5):**
  **(1)** `:192` *"A chain entry has exactly two legal shapes"* → three, with the object form's grammar.
  **(2)** `:47` the 2.0 trigger, *"an embedded-face entry"* → *"an object-form entry"* — **load-bearing,
  not editorial.** **(3)** `:598` the licence requirement, currently reaching only *"an asset a chain
  names by `{"asset": "<key>"}`"*, must widen to reach a **variant** asset key. **(4)** the chain-entry
  row itself: the three new keys, the closure and its price, and the **type difference**.
- **The 11.1 guard table is a font-set table and none of it fires here.** This story adds no faces:
  `TestFontsShippedMatchesExpectedFaceSet` (`shipped_faces_ext_test.go:72`),
  `TestShippedSpecCoversEverythingShipped` (`shipped_faces_test.go:525`), `wantDerivedShippedFaces = 7`
  (`fontgen_matrix_test.go:264`), and the designer `toBe(11)`/`toBe(13)` counts must all stay green
  untouched. **If any moves, the story has drifted into 11.1's territory.**

### Facts the implementer will otherwise re-derive

- `folio-go/fonts/fonts.go:150` `Shipped()` — eleven keys: `Noto Sans`, `Noto Sans Bold`,
  `Noto Sans Italic`, `Noto Sans Bold Italic`, `Noto Sans Thai`, `Noto Sans Thai Bold`, `Noto Sans SC`,
  `Roboto`, `Roboto Bold`, `Roboto Italic`, `Roboto Bold Italic`. **Noto Sans** and **Roboto** have all
  four cuts; **Noto Sans Thai** has Regular + Bold only; **Noto Sans SC** is Regular only. The last two
  are AC3's permanent absence subjects. `fonts.go:147` carries the prohibition itself.
- `fontCache` (`render.go:1256`) is keyed by **bare face name**, so `"Noto Sans Bold"` is just another
  `fs[name]` lookup — two independent cache entries, parsed once each. Correct, not a leak.
- `Style.Bold`/`Italic` are `Presence[bool]` (`model.go:358`/`:359`), **five** production readers in four
  files, none of them in the draw path. Neither has a `presentNull` path — and ⚠ **CORRECTED during
  implementation: `"bold": null` is NOT a load error.** `decodeBoolRaw` (`decodehelpers.go:60`) delegates
  to `json.Unmarshal`, and Go reads JSON `null` into a `bool` as **`false` with no error** (measured with
  a throwaway program). So `null` decodes to `present(false)`, `.Null` is never set for these two fields,
  and the `!Null` half of the new cascade arms is **inert today**. Carry it anyway to match the nine
  siblings — `table_render.go:358-368` records the live defect from omitting it — but say in the test
  that it is currently unreachable rather than shipping an untestable claim.
- `Validate` reaches `predictDocument` directly (`validate.go:68`); `Render` via
  `renderDocument`→`buildPageModel` (`render.go:2007`). **AC3's Warning is emitted by both for free.**

## Tasks & Acceptance

**Execution:**

- [x] `folio-go/internal/template/model.go` -- Add the three optional variant siblings to
  `FontChainEntry` as **plain `string` fields** (mirroring `Face`/`AssetKey`: `""` means absent), plus
  **one exported-to-the-package predicate reporting whether the entry serialises as an object**. Amend
  the `:157-163` doc comment to state the mechanism change and the preserved property; correct the stale
  call-site count at `:213-219`. -- One predicate, two consumers, so they cannot disagree.
- [x] `folio-go/internal/template/parse.go` -- In `decodeFontChainEntry`: admit the object form; hold the
  variant key set **once, as a named enumeration — the authority**; enforce exactly-one-of
  `face`|`asset`; enforce **namespace matching** (a `face` entry's siblings are face names, an `asset`
  entry's are asset keys) in **both** directions; refuse empty and wrong-typed variants with a message
  saying which meaning applies here; run `requireEmbeddedFaceLicence` on **every variant asset key**;
  and **derive all three refusal messages from the enumeration** (`:407`, `:417`, the `default` branch)
  so none can go stale. Keep every refusal located at `fonts.<name>[<i>]` / `…[<i>].<key>`.
- [x] `folio-go/internal/template/serialize.go` -- `writeFontChain` emits a bare string iff the shared
  object-form predicate is false; otherwise an object through `writeObject`, with siblings visited in
  the closed set's fixed order. Positional `kv{"key", writer}` literals only. -- AC5 lives in the
  bare-string branch.
- [x] `folio-go/internal/template/version.go` -- Replace `fontsRequireMajor`'s `entry.Embedded()` with
  the **shared object-form predicate**. `SupportedMajor`/`SupportedVersion` do not move; no new constant,
  no new rank. -- Otherwise a document carrying an object entry is stamped `1.0` and lies.
- [x] `folio-go/internal/template/` tests -- **Behavioural red proof of the version tie**: a literal
  table of entry shapes asserting `(the serialised entry begins with '{') ⟺ (versionForSave raises to
  2.0)`. **Pin the serializer's real output against the version's real output — never the predicate
  against itself.**
- [x] `folio-go/render.go` -- Resolve style **within** the entry coverage chose. `chainFaceNames` keeps
  producing the base names coverage walks; derive a parallel, same-length, same-order styled-name list
  that shaping and measurement consume. Emit AC3's Warning from `shapeSegments`, coalesced per (element,
  distinct rune) with a linear slice scan. Route the face's display name through `formatFontChain` or an
  equivalent. Correct the stale 6/4 doc table at `:1194-1213` to 6/2/2.
- [x] `folio-go/table_render.go` -- Give `resolveHeaderStyle` bold and italic arms in the exact
  nine-sibling spelling (`.Set && !.Null` on **both** arms), and carry the resolved style into the header
  chain at `:707` and the body/footer chains. -- AC2.
- [x] `folio-go/component_commands.go` -- Add `bold`/`italic` to `tableHeaderStyleFields` (`:2521`) and
  **retire the `:2510-2520` ruling comment by EDITING it, never deleting**, so the history reads.
- [x] `folio-go/embedded_face.go` -- Teach `newEmbeddedFaceIndex` the variant asset keys, visiting
  siblings in the closed set's fixed order so the walk stays deterministic on all three axes.
- [x] `folio-go/internal/diag/diag.go` + `folio-go/diagnostic.go` -- Mint one new **Warning** code across
  all eight registry sites. Do **not** reuse `TEXT_MISSING_GLYPH` (see Design Notes).
- [x] `folio-go/diagnostic_registry_census_test.go` -- Add the new code's Warning trigger: a real
  `ParseTemplate` + `Render` producing bytes and a **located** Warning, with a fixture-precondition guard.
- [x] `_bmad-output/specs/spec-folio/folio-format.md` -- The **four coupled edits** listed in the Code
  Map: `:192` (three shapes), `:47` (object-form version trigger), `:598` (licence reaches variant asset
  keys), and the chain-entry row (three keys, **the closure and its price — extending the set later is a
  MAJOR change**, and **the type difference**: `style.bold` is a boolean, a chain entry's `bold` is a
  face name or asset key).
- [x] `folio-go/internal/template/` -- **ONE targeted test** asserting (a) a **literal** list of the three
  keys equals the parser's enumeration, and (b) `folio-format.md` carries a row per key **under the
  chain-entry row**. **Literal-vs-derived, never derived-vs-derived** — if the doc assertion reads the
  same enumeration the parser reads, both sides move together and the guard is vacuous. **Red-prove by
  deleting that row while leaving `style`'s row intact.** Do **not** widen it to police edits 1–3.
  **PRICED AT THE PLAN GATE — one test, and the separation is structural:** `` ## `fonts` `` occurs
  **exactly once** (`:182`), its section ends at the next `^## ` (`:256`), and `style`'s
  `| `bold`, `italic` |` row is at **`:497`** inside `### style` (`:473`–`:526`) — outside the slice by
  construction. ⚠ Assert the heading matches exactly once before slicing, and slice to the next `^## `,
  never a hardcoded line.
- [x] `folio-go/internal/template/fonts_embedded_test.go` -- **`:620` inverts: `{"face": "Noto Sans"}` is
  pinned today as a load error and becomes VALID. Replace that row** with the genuine no-discriminant
  case (e.g. `{"bold": "Roboto Bold"}`), do not re-explain it. **Add a both-discriminants row**
  (`{"face":"Roboto","asset":"<key>"}`) or "exactly one of" is asserted in neither direction. **`:621`
  keeps its name but gains a message assertion** — `requireLoadError` checks the field only, so today it
  cannot see that the ground moved.
- [x] `folio-go/internal/template/fixtures_test.go` -- Extend `maximalFixture` to exercise any genuinely
  new serialized key, or `TestDriftASTMatchesRuntimeEmission` reds as `astOnly`.
- [x] `folio-go/` tests -- The acceptance tests: **the DW-233 tripwire as its own named test**, both
  permanent-absence arms (Noto Sans SC bold; Noto Sans Thai italic), the never-walk-the-chain case, the
  **AC2 table-header** case, an AC4 wrapping-difference test, the AD-8 cross-namespace refusals in both
  directions, and an AC5 byte-identity test. Assert AC3's **message content** (the face it names), not
  merely the element id. -- A refusal asserted only by its address is not asserted (D-000.25).
- [x] `folio-go/component_commands_test.go:156` -- Update
  `TestPlacedImageStartsEmptyAndSurvivesTheRoundTripAndRender`'s zero-diagnostics assertion. **State in
  the test why it moved** — the fixture declares `bold` on `e1` against a chain with no variants, so it
  now earns AC3's Warning — so the next reader can tell a deliberate change from a silenced one.

**Acceptance Criteria:**

- Given a text element declaring `style.bold`, `style.italic`, or both, and a chain whose covering entry
  declares the matching variant, when the document renders, then the glyphs are shaped **and measured**
  from that variant face's own metrics, never the base face's.
- Given a table whose **header row** declares bold through `headerStyle`, when the document renders, then
  the header cells resolve to the declared variant. (Its own test **because the header arm calls
  `chainFaceNames` directly at `table_render.go:707`, bypassing `fontChain`**: a `fontChain`-only
  implementation passes every text AC and silently fails this one.)
- Given a covering entry with no declared variant for the requested style, when the document renders,
  then the rune renders in **that entry's own base face**, the render still succeeds, and a Warning from
  the additive registry names the element, the rune and the face.
- Given a chain entry `"Roboto"` with no declared variant, a `FontSet` that **does** contain
  `"Roboto Bold"`, and an element declaring bold, when the document renders, then it renders Roboto
  **Regular** and emits the Warning.
  **Do not delete or weaken this test, and this is why — it is not here because a register asked for
  it.** Measured: no pinned golden renders a bold- or italic-declaring document
  (`grep -rn '"bold"\|"italic"' fixtures/` → 0 across all 30 fixture dirs; positive control
  `'"fontFamily"'` → 23 files), and nothing pins `worked-example.json`'s rendered bytes
  (`grep -c "worked-example"` → 0 in `byte_neutrality_test.go`, `matrix_test.go`,
  `statement_golden_fixture_test.go`; positive control `"statement-1"` → 5). So an implementation that
  falls back to constructing `"<family> Bold"` would **silently switch `e1`'s face, fire no Warning,
  leave the one expected red green, and move no golden.** This test is the only thing standing between
  the repository and a silent face swap that nothing else can see.
- Given any chain entry that serialises as an object — shipped or embedded — when the document is saved,
  then the version is raised to `2.0`; and given every entry a bare string, then the version is unchanged.
- Given an entry whose sibling names a different namespace than its discriminant, when the document is
  loaded, then it is a located load error — asserted in **both** directions.
- Given bold or italic set on an element, when its lines are broken and measured, then the breaks are
  those of the face actually used, and canvas and PDF agree because both read one engine measurement.
- Given a document declaring neither bold nor italic, when it is rendered on every target, then nothing
  about its resolution changes, its serialized bytes and its version string are unchanged, and all 24
  recorded digests and all four AD-21 target legs are unmoved (AD-21).
- Given a document whose chain entries predate this story, when it is opened, then nothing is migrated,
  rewritten or repaired; it reports honestly (I-5).

## Spec Change Log

**2026-09-06 — Code Map correction: `"bold": null` is NOT a load error.** The Code Map asserted that
`Style.Bold`/`Italic` have no `presentNull` path *and* that an explicit null is refused. The first half
is true, the second is false. `decodeBoolRaw` (`decodehelpers.go:60`) delegates to `json.Unmarshal`, and
Go reads JSON `null` into a `bool` as **`false` with no error** — measured with a throwaway program, not
inferred. So `null` decodes to `present(false)` and `.Null` is never set for these two fields, which
makes the `!Null` half of the new cascade arms **inert today**. The arms still carry `!Null` to match the
nine siblings (`table_render.go:358-368` records the live defect from omitting it), and the test says the
branch is currently unreachable rather than claiming coverage it does not have. **Known-bad state
avoided:** shipping a test that asserts an explicit-null cascade path which no input can reach — a
guard that cannot fail. The wrong claim reached the Code Map from a survey subagent and was caught only
because a second agent contradicted it; see the entry below on relayed absences.

**2026-09-06 — the Matrix Test Audit found a production defect, not just a missing test.** The row
*"Variant names an absent FontSet face"* had no covering test: all four variant values in the acceptance
suite named faces the shipped set supplies. Writing the missing test **aborted the render** —
`none of the fallback chain's faces [Roboto Condensed Bold] is present in the supplied FontSet, so no
line height can be derived from it`. The two gates disagreed about the long-standing FontSet tolerance:
shaping applied it (`faceCovers` opens with `cache.declares`), while `drawnFaceNames` coalesced the
styled name onto the drawn chain **unconditionally**, so an unsupplied variant name reached
`chainLineMetrics`. Glyphs came from the base face while the leading was being derived from a face
nobody supplied, and on a single-entry chain that is a hard render failure **on a document the format
calls valid**. Fixed by giving `drawnFaceNames` the same `cache.declares` gate — one tolerance asked at
both gates — across its four call sites. **KEEP on re-derivation:** the gate, its doc comment recording
why it exists, and the test's precondition guard proving the named variant really is unsupplied.
**Known-bad state avoided:** a valid document that renders everywhere the variant happens to be supplied
and aborts everywhere it is not. This was not an intent gap — the matrix row, the frozen Boundaries and
the Design Notes all specified the behaviour identically; only the metrics path failed to honour it.

## Design Notes

**FR57's own wording already says "declared".** `epics.md:121`: *"Realize `style.bold` and `style.italic`
as real weighted and sloped faces, **resolved per rune through the declared chain**, with no synthetic
emboldening or obliquing anywhere."* The no-inference rule is not a preference layered on the
requirement; it is what the requirement said. Cite this before re-litigating the resolution source.

**Why the one-key rule may change: mechanism versus property.** `parse.go:411-418` states the property in
its own words — *an unknown key cannot ride along disguised as a decoration*. Exactly-one-of over a
**closed** set preserves that property exactly. And D-B's reason for calling this "not a new format axis"
was never an idiom: it was that `writeFontChain` **already** emits an entry as either a bare string or an
object, which is precisely what the object form uses.

**Why no MAJOR bump — a ruling, not a default.** The doc's own test is *would a pre-2.0 reader refuse this
file or render it wrong?* It refuses, on the entry shape, so the existing 2.0 trigger already covers it.
Minting 3.0 was considered and rejected on `folio-format.md:617-634`'s grounds: Folio is unreleased, and a
bump would make **every** document declare 3.0 — including the twenty-two fixtures that make no font
choice at all — moving their bytes and goldens for a reason unrelated to fonts. We widen what 2.0 *means*;
every reader of 2.0 that has ever existed lives in this repo and moves in the same commit.

**"The same treatment an uncoverable rune already gets" is about TREATMENT, not IDENTITY.** It means
*stated, not silent* — a Warning naming element, rune and face. `TEXT_MISSING_GLYPH` names element + rune
+ **the whole chain** and means the rune was **dropped** (no glyph, no advance); AC3 names **one face**
and means the rune was **drawn at the wrong weight**. D-4.5.1's two-limb test — same author action AND
same thing happened to the document — fails on both limbs, and reuse would make the shipped message text
false, which AD-14 makes a breaking change. **RULED: mint.**

**Why a parallel styled list rather than a substituted chain.** If the bold name replaces the base name
*before* coverage runs, a variant whose `cmap` is narrower pushes a rune to the **next entry**, silently
changing the typeface to keep the weight. Coverage must walk the base names. The minimal shape satisfying
that and "one answer site" is one function producing two aligned slices, with the index carrying the
correspondence.

**When the declared variant itself lacks the glyph.** Coverage chose the entry on its base face, so the
variant may not cover that rune. Treat it as **absence**: render in the entry's own base face and emit
AC3's Warning. This keeps "its own base face, never a walk down the chain" true on every path rather than
inventing a second fallback policy.

**A partial style match is an absence.** Bold+italic against an entry declaring only `bold` resolves to
the base face and warns. Choosing the bold cut would be a nearest-fit search — the inference this design
refuses.

**Example — the entry shapes, AC5's first:**

```json
"fonts": {
  "body": [
    "Noto Sans Thai",
    { "face": "Roboto", "bold": "Roboto Bold", "italic": "Roboto Italic" },
    { "asset": "9f86d0…", "bold": "1b4f0e…" }
  ]
}
```

The first entry is untouched and must serialise byte-for-byte as today — that is AC5, satisfied by
construction. The second raises the document to `2.0`; so does the third, and under the current predicate
only the third would.

## Verification

Run each module's commands in **its own** invocation — a `cd` persists through a compound command and
makes later relative paths resolve elsewhere, printing `lstat …` lines that read like passes. Use
`-count=1` on every Go run: the cache serves a stale PASS for tests that walk the filesystem. Never
`&&`-chain these; a conjunction silently drops everything after its first failing term.

**Commands:**

- `cd folio-go && go test -count=1 ./...` -- baseline at `102e1fc` was **2124 pass / 2 fail / 5 skip**,
  exit 1. The ONLY acceptable pre-existing failures are `TestCorpusMeetsP6ExerciseFloors` and its
  `P6g_(opaque_names)` child — a mandated permanent red, never to be "fixed". **Any third distinct
  failure that is not the one expected red named below is a hard stop: report it, do not triage around
  it.**
- `cd folio-go && FOLIO_FONTGEN_PYTHON=/private/tmp/claude-501/-Users-panitw-Projects-folio/c910e6ef-835c-4b93-9710-f11720b44f3a/scratchpad/fontgen-venv/bin/python go test -count=1 -tags=matrix ./...`
  -- **the FULL matrix suite, UNFILTERED. Never a `-run` filter.** Baseline **2136 / 2 / 5**. This story
  puts AD-21 directly at risk, so a green from anything narrower discharges nothing. Confirm
  `fontgen: derived and compared 7 of 7 faces`; without the env var the test degrades to a
  could-not-execute, and a form that cannot run is not a form that can pass.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` -- PASS, **four
  legs** (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`), **24 documents each**. Quote the
  per-leg counts.
- `cd lint && go test -count=1 ./...` -- four `ok`, **227 tests**, exit 0.
- `gofmt -l /Users/panitw/Projects/folio/folio-go /Users/panitw/Projects/folio/lint` -- from the **repo
  root with absolute paths**; empty output. Any `lstat` line is a non-measurement.
- `cd folio-designer && npx tsc -b --force` -- exit 0.
- `cd folio-designer && npm test` -- **64 files / 953 tests, all passing**, exit 0. Never run
  `npx vitest` from the repo root: it picks up the Playwright specs and produces a false mass-failure.
- No release build. This story changes no designer-side asset; do not run `npm run build`.

**Manual checks:**

- **The ONE expected red, and it is expected.**
  `folio-go/component_commands_test.go:156` `TestPlacedImageStartsEmptyAndSurvivesTheRoundTripAndRender`
  asserts zero diagnostics while rendering `worked-example.json`, whose `e1` declares `bold` against a
  variant-free chain. **It must be updated, not silenced, and the updated expectation must say why it
  moved.** Every other consumer of that fixture and of `maximalFixture` is unaffected — see the Code
  Map's measured per-test table. If anything *else* reds, it is not this.
- **Mutation-proof AC3, the DW-233 tripwire, and the version tie.** A green suite over correct code
  proves nothing. For each: reverse the behaviour at the production site (make the resolver construct
  `entry.Face + " Bold"`; make absence silent; make `fontsRequireMajor` read `Embedded()` again) and
  confirm the named test reds; restore and confirm it greens. **Anchor every mutation and check WHERE it
  landed** — an unanchored edit reds an unrelated test and proves the wrong thing. Report which test
  caught which mutation.
- **Red-prove the doc-row assertion** by deleting the chain-entry row while leaving `style`'s row at
  `folio-format.md:497` intact. If the deletion stays green the assertion is not doing its job.
- **AC5 witness, as a closed population.** Measured over tracked files: **29 `.folio` files declare
  `bold`/`italic` zero times**, positive control `"fontFamily"` in **24**. State the population as
  `.folio` files **plus** `folio-go/testdata/template/golden/*.json` **plus** the in-source canonical
  fixtures — the `.folio` corpus alone is not the whole document population, and scoping to it is how
  invariant I-6 came to certify "zero" while two documents declared bold. Record that as a **census-scope**
  error: I-6 was correct about what it looked at.
- **The two bold-declaring documents are a genuine positive control, not a formality** — they prove the
  corpus can register a change at all. Confirm their **bytes and version strings** are unchanged.
- **This story ships NO bold golden fixture, and that is a ruling — not an oversight for the closer to
  flag.** No pinned `expected.pdf` renders a bold-declaring document, and this story deliberately does
  not add one. Under DW-12 a new pinned document needs its own golden plus the full four-target matrix
  in-story, and an `expected.pdf` is a **human-attested artifact** (AD-21 / D-4.7.1) — attestation is an
  owner action, and nobody can honestly attest a bold page before **Story 11.3** makes bold reachable
  and `starter.folio` declares variants. **Registered as deferred work owned by Story 11.3.** The DW-233
  tripwire covers the mechanism in the meantime; the golden will cover the bytes.
- Confirm no shipped-face count assertion moved (the 11.1 guard table). If one did, the story has drifted.
- Report `git status --porcelain` and confirm **nothing was committed**: HEAD must still be the
  dispatch's baseline.

## Suggested Review Order

**Start here — the design in one type**

- The whole design: three optional plain-string siblings, and the one predicate two consumers share.
  [`model.go:186`](../../folio-go/internal/template/model.go#L186)

- THE authority for the closed set — key spelling, `FontStyle`, and struct field tied in one table.
  [`model.go:235`](../../folio-go/internal/template/model.go#L235)

**The version tie — the defect this story nearly shipped**

- The shared predicate. Two copies of `Embedded()` agreed only until object-form and embedded diverged.
  [`model.go:272`](../../folio-go/internal/template/model.go#L272)

- Consumer one: bare string iff not object-form. This branch is AC5.
  [`serialize.go:210`](../../folio-go/internal/template/serialize.go#L210)

- Consumer two: raises to 2.0. Reading `Embedded()` here stamped 1.0 on a file no 1.x reader can decode.
  [`version.go:382`](../../folio-go/internal/template/version.go#L382)

**Resolution — coverage on the base chain, style within the chosen entry**

- Two aligned slices. Coverage walks `base`; only shaping and metrics see `styled`.
  [`render.go:1353`](../../folio-go/render.go#L1353)

- The metrics chain. It APPENDS the variant beside its base — both faces can draw, so both belong.
  [`render.go:1244`](../../folio-go/render.go#L1244)

- Returns the entry INDEX, so style resolves within the entry coverage already chose.
  [`render.go:1758`](../../folio-go/render.go#L1758)

- The absence arm: variant applied by index, or the entry's own base face plus AC3's Warning.
  [`render.go:1976`](../../folio-go/render.go#L1976)

- One tolerance for "supplied" and "covers this rune", asked identically on both gates.
  [`render.go:1738`](../../folio-go/render.go#L1738)

**The format — an object form that keeps the unknown-key refusal**

- Exactly-one-of `face`|`asset`, a closed sibling set, and namespace matching in both directions (AD-8).
  [`parse.go:464`](../../folio-go/internal/template/parse.go#L464)

- Every refusal sentence derived from the enumeration, so none can go stale as the set moves.
  [`parse.go:405`](../../folio-go/internal/template/parse.go#L405)

- Three shapes, the closure and its MAJOR price, and the boolean-vs-face-name type difference.
  [`folio-format.md:182`](../specs/spec-folio/folio-format.md#L182)

**The cascade and the embedded arm**

- Bold and italic arms in the nine-sibling spelling; `headerStyle` wins for the header row alone.
  [`table_render.go:321`](../../folio-go/table_render.go#L321)

- The ruling that bold/italic "have nowhere to resolve from" retires here — edited, not deleted.
  [`component_commands.go:2527`](../../folio-go/component_commands.go#L2527)

- Variant asset keys are a third determinism axis; missing them fails silently.
  [`embedded_face.go:269`](../../folio-go/embedded_face.go#L269)

- Digit table and slot run must shape through one resolution, or page numbers get foreign CIDs.
  [`page_number.go:451`](../../folio-go/page_number.go#L451)

**The diagnostic**

- Minted, not reused: this rune was DRAWN at the wrong weight, not dropped.
  [`diag.go:145`](../../folio-go/internal/diag/diag.go#L145)

- Coalesced per (element, distinct rune) across table rows, by linear scan for determinism.
  [`render.go:1858`](../../folio-go/render.go#L1858)

**Tests and peripherals**

- The DW-233 tripwire and every absence arm — the only guard over a silent face swap.
  [`style_face_resolution_test.go`](../../folio-go/style_face_resolution_test.go)

- The embedded half, end-to-end: a carried bold face drawn from its own asset key.
  [`style_face_embedded_test.go`](../../folio-go/style_face_embedded_test.go)

- Closed set vs the doc, and the version tie proved output-against-output.
  [`font_chain_variants_test.go`](../../folio-go/internal/template/font_chain_variants_test.go)

- The one pre-existing test whose expectation moved, narrowed rather than silenced.
  [`component_commands_test.go:198`](../../folio-go/component_commands_test.go#L198)

## Delivery Log

### 2026-09-06 — done

Baseline `3ad4ede`. Shipped in **`d0ded7e`** on `main` — **36 files, +3746 / −282** (`--numstat`, summed;
the `--stat` line is a single column and is not a split). Local and unpushed at this close: `main` sits
**12 commits ahead of `origin/main`** and nothing here was pushed. Decisions **D-11.2.1 through
D-11.2.10** live in [`epic-11-14-decision-log.md`](./epic-11-14-decision-log.md) and are not restated.
This closer touched only this story file, `sprint-status.yaml` and `deferred-work.md`, and staged nothing.

**What shipped.** A font chain entry may now be an **object** declaring its own cuts. Per-rune coverage
still decides the entry on the **base** chain, and the declared variant is applied **within** the entry
coverage already chose — never by constructing or parsing a face name, and never by inferring one from
the binaries. An entry with no declared variant for the requested style renders in **its own base face**
and emits a newly minted Warning naming the element, the rune and the face. Bare-string entries serialise
byte-for-byte as before, so the corpus does not move.

**The defect this story nearly shipped, and it was found before code rather than by a test.** The version
predicate and the serializer held **two copies of `entry.Embedded()`** that agreed only because
object-form and embedded had always been the same set. The object form separates them, and the old
predicate would have stamped **`1.0`** on a document **no 1.x reader can decode** — green, silently, with
nothing asserting the negative direction because no such entry had ever existed. One shared predicate now
feeds both sites, red-proved behaviourally output-against-output rather than by asserting the predicate
equals itself.

**What review caught, and it was not a guard gap.** Two real correctness defects, both in *measurement*
rather than in the new resolution logic. First, the metrics chain **replaced** the base face instead of
**adding** it: a variant that did not cover a rune drew glyphs from the base while deriving leading from
the variant, and on a single-entry chain the base was excluded outright — a hard render failure on a
document the format calls valid. Second, fixing that made a **latent** page-number defect reachable:
digit CIDs are injected without re-shaping, so the faces must match, and nothing pinned that. **They had
to land in one commit**, which is why the review's two findings are inseparable in the diff. The
Matrix Test Audit is what surfaced the first one — an uncovered matrix row led to a missing test led to a
production bug, the complete chain (D-11.2.9).

**AC2 caught nothing, and that is the criterion working.** D-11.2.5 wrote the failure mode *into* the
acceptance criterion — that the table-header arm reaches `chainFaceNames` directly and bypasses
`fontChain` — and the implementer routed style through `chainFaceNames` from the start. A guard that
changes behaviour before it can fire appears in no findings count and paid for itself anyway.

**Review triage: 16 patched / 5 deferred / 0 rejected / 0 loopbacks** (`review_loop_iteration` 0). Taken
as the build loop's own tally and **not re-litigated at this close**. The count is stated, not
enumerated, in the build's report, so this close did not spot-check the individual patches.

**THE REGISTER DEBT, and this close discharged it — seven entries minted, DW-238 through DW-244.** The
build wrote its five deferrals in the workflow's prescribed `- source_spec:` block form, which is
**invisible to a `### DW-` census** — the exact defect that lost four of Story 11.1's five deferrals
(D-11.0.2). To its credit the build put them under their own heading and said in the file that they
needed real numbers, so nothing was buried inside another entry's body this time. Three things worth
recording about the conversion:

- **Its own heading said "four" twice; there were five blocks.** An undercount inside the very note
  warning about a census gap. The fifth (`TableColumnsProjection` has no header bold/italic) is now
  **DW-240**.
- **One deferral became two entries.** Per **D-11.2.10**, *"whichever story lands first must close it"*
  names two owners and is how an item is dropped by both. The **destruction** half is **DW-238, Story
  11.4's hard precondition**; the **projection** half is **DW-239, Story 11.3's**. Reachability was
  measured rather than assumed: the chain-command builders have no production caller and `embedFontFamily`
  cannot rebuild an existing chain, so **severity is LOW today and HIGH the moment 11.4 builds the
  surface**. **Epic 11 does not close with either open.**
- **A seventh entry came from the decision log, not from review.** **DW-244** registers D-11.2.9's "number
  to watch": AC3's Warning is per (element, **distinct rune**). Re-measured at this close rather than
  relayed — `worked-example.json`'s `e1` binds `"Statement for Ada Lovelace"`, **16 distinct runes, so 16
  Warnings from one element on one render**. Distinct-rune coalescing is bounded by the *alphabet*, and
  for CJK body text the alphabet is the text: `Noto Sans SC` is Regular-only, so a bold CJK page takes the
  absence arm on **every** rune. Per spec, not a defect, unpriced at scale.

Owners for **DW-238**, **DW-239** and **DW-244** were ruled by the orchestrator. **DW-240**, **DW-242**
and **DW-243** were routed by this closer and each says so in its own entry; **DW-243** (the pre-existing
footer cache) is the one I am least confident about and is registered unassigned to get an owner rather
than to claim one. **DW-241** — a variant naming its own base face, accepted silently — is routed to the
**engineering lead** because it needs a **ruling**, not a fix: the behaviour is self-consistent under the
rules this story shipped, so no implementer can settle whether it is legal.

**Measured gates at this close — my figures, at `d0ded7e` with a clean tree.**

| Gate | Result |
|---|---|
| `cd folio-go && go test -count=1 ./...` | **2187 pass / 2 fail / 5 skip**, exit 1 |
| `cd lint && go test -count=1 ./...` | **227 pass**, four packages `ok`, exit 0 |
| `gofmt -l folio-go lint` (repo root, absolute paths) | **empty**, exit 0 |
| `cd folio-designer && npx tsc -b --force` | exit 0 |
| `cd folio-designer && npm test` | **64 files / 953 tests passed**, exit 0 |

The two Go failures are **`TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child** — the
mandated permanent red, never to be "fixed". **There was no third failure**, and that is asserted from the
enumerated failing test names rather than from the exit code: the pass/fail/skip totals were counted from
`go test -json` `Action` events carrying a `Test` field, because a plain run prints no totals and a
carried-forward figure is not a measurement. `tsc` was run with `--force`, since `tsc -b` exits 0 without
typechecking when its build info is current.

**Not run by this closer, and recorded as the builder's figures rather than mine.** The full
`-tags=matrix` suite (**2199 / 2 / 5**, `fontgen: derived and compared 7 of 7 faces`) and
`TestCrossTargetByteIdentity` (**24 documents × 4 legs**, every document's four hashes checked
individually rather than by trusting the PASS) were run **unfiltered by the build** and re-measured by the
orchestrator at commit time. This close did not re-execute them. No release build was run — this story
changes no designer-side asset.

**Tracker and output tree.** `sprint-status.yaml` changed by exactly one line, `review` → `done`;
`epic-11` stays `in-progress` with 11.3 and 11.4 open. Its 512 comment lines were left intact: a value
histogram over all 169 keys returned only bare status tokens, and the comment blocks around the Epic 11
keys are the owner's and the lead's own record (D-A, D-11.0.1, D-11.2.1), written deliberately by prior
closers and recorded nowhere else. **No agent narrative was found to move.** The output tree carried no
debris from this story: nothing untracked anywhere under `_bmad-output/`, no orphaned review prompts, no
patch files, no duplicate slug specs, no unresolved result files. `epic-11-context.md` is **not stale** —
it was recompiled inside this story's own commit, after `epics.md` last moved at `102e1fc`.
