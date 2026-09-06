---
title: 'Picking a family declares the cuts it has'
type: 'feature'
created: '2026-09-06'
status: 'done'
baseline_commit: '7cc02d14f7da3bf1cd22caddbfbf1d6ae92fd72e'
review_loop_iteration: 0
context: []
---

## In plain terms (read this first if you just want the gist)

*Not normative, and rewritten at close: the frozen Intent below governs implementation.*

The designer ships with a few font families already inside it — Roboto, Noto Sans and their relatives.
Picking one used to *embed a copy of it into your document*: several hundred kilobytes identical to bytes
the application already carried.

It now writes the family's **name** instead, with the cuts that family has — which weights and slopes are
real faces rather than something the browser fakes. Documents get smaller, and the bold and italic on
screen are the ones the engine will print. In practice only Roboto gains this, being the one shipped family
the picker offers; the owner had that correction before the work began.

**The fallback tail is computed by one piece of code, not two.** A font chain is a family plus fallbacks
covering the scripts it does not — Roboto covers Latin, so it needs Thai and Chinese behind it. Both paths
now share one function; two that merely agreed today would eventually disagree in silence, and a pick would
quietly lose coverage.

**A pick can no longer write a font asset by hand.** The refusal is structural now — the reader of the file
format rejects the shape outright — not a rule to remember.

**A document naming a family as its own bold is now refused when opened**, with the place and reason given.
Deliberate: such a file used to load and then print bold as ordinary text with no warning at all. Nothing
older is migrated, rewritten or repaired.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Story 11.2 taught the engine to resolve a declared cut and Story 11.3 taught the canvas to
paint it, but nothing in the product **writes** a cut into a document. `starter.folio` is the only file in
the world that declares one, and only because 11.3 hand-authored it. Worse, picking `Roboto` in the family
control today **embeds a byte-identical duplicate** of the `Roboto` the engine already ships
(`e688a215e0841b6e4edb…`, measured on both copies) — a Regular-only copy that can never bold while
`Roboto Bold` sits unreachable in the same FontSet. A second, unrelated door is still open: a chain entry
whose variant names its own base face (`{"face":"Roboto","bold":"Roboto"}`) is accepted silently and
renders bold-as-regular with **no** Warning — the one outcome AC3 exists to prevent.

**Approach:** Picking a family that the release already ships **names and declares** it — a plain
`face` entry carrying that family's cuts — instead of embedding a copy (D-16.5 says only *"the moment a
font starts travelling"* moved, and a face already on every machine is not travelling). The family→cuts
grouping becomes the product's first **declared** production table, mirrored to the engine's FontSet and
never derived from a face name. Picks that still embed get the same treatment for their proposed tail. And
D-11.2.11's narrowing lands: a variant naming its own base becomes a located load error, with the limit of
that check stated where a reader will meet it.

## Boundaries & Constraints

**Always:**
- **Resolution is DECLARED, never constructed, parsed, or inferred from binaries (D-11.2.1 / D-11.2.2).**
  `"Roboto" + " Bold"` is the foreclosed carrier written forwards; `TrimSuffix(key, " Bold")` is the same
  thing backwards. The family→cuts table is **hand-written data**, and a reader must not be able to mistake
  it for a naming convention.
- **The criterion for "this family is shipped" is membership in the declared mirror.** Never
  `build-wasm.mjs`'s `shippedFamilies`, never a parse of `fonts.go`, never a byte comparison. *Measured:
  `shippedFamilies` is wrong in **both** directions — it omits plain `Roboto` (13 names, no `Roboto`) and
  it includes `IBM Plex Sans` / `IBM Plex Mono` / `IBM Plex Sans Thai`, which have no `Shipped()` key at
  all. A membership test against it would answer "Roboto is not shipped" and would write
  `{"face":"IBM Plex Sans"}` into a document the engine then refuses.* It is the browser's **CSS family
  registry**, not the engine's FontSet.
- **The declared mirror's bases plus its cuts must equal `fonts.Shipped()`'s key set, asserted BOTH ways.**
  Four bases and seven cuts is eleven, which is `Shipped()` exactly. A tie test that only checks one
  direction cannot see a face that leaves the FontSet.
- **A pick writes FontSet FACE NAMES as variants, never asset keys.** This is what keeps
  `assetKeyReferenced`'s blind spot out of this story's reach (see Design Notes). If any path you build
  writes an asset sibling, **stop and report it** — the out-of-scope ruling flips.
- **A pick must never write a variant equal to its own base face.** DW-241 makes that a load error in this
  same story, so such a pick would author a document the product cannot reload. Assert it directly.
- **`Noto Sans SC` declares no cuts and takes AC3's Warning.** D-A's permanent shipped instance of a family
  with no face at a requested weight. Exercise it as **ordinary behaviour**, not as an edge case.
- **DW-241 (D-11.2.11): one predicate per arm, measured against THAT ARM's own discriminant value.** A
  `face` entry's variants are compared to `entry.Face`; an `asset` entry's to `entry.AssetKey`. **A
  cross-variant collision — `{"face":"Roboto","bold":"X","italic":"X"}` — is LEGAL and silent.** The base
  is privileged and that is the only axis.
- **The DW-241 refusal message is DERIVED from the closed set, never hand-written.** The three keys must
  not be spelled a second time in a sentence.
- **DW-245: disclose what the check does NOT catch**, in the code beside the check **and** in the format
  doc's chain-entry row. It is string equality, so `{"face":"Roboto","bold":"Roboto Copy"}` where two
  FontSet keys hold identical bytes is undetectable without reading binaries, which D-11.2.1 forbids.
- **AC3's trigger does not move.** It stays a *declaration* property, because under D-11.2.11 a
  self-referential variant never reaches the resolver.
- **I-5, no migration.** A document whose chain entries predate this story is not rewritten, repaired or
  normalised on open or on selection. It acquires variants only when its author re-picks the family.
- **AD-21.** A document whose author never touches the family control renders to identical bytes.

**Ask First:**
- **Running `npm run build`** (the full offline release build). `npm test` and `npm run typecheck` both run
  `build:wasm` themselves and are routine; the full build is not. Ask, and say what you measured.
- **Any change to the `command-json-soleness.test.ts` factory arrays or their Go twin.** Registering the
  Go list's drift is in scope; editing it is not unless this story touches that list anyway.
- Any change that moves a golden digest, a rendered byte, or the count or set of shipped faces. The 11.1
  guard table must stay green untouched; a move means this story drifted into 11.1.
- Any change to `fonts.Shipped()`, `folio.FontSet`'s shape, or the canvas fragment fallback stack.
- Adding a **third** group to the family control. 16.4 and 16.9 cut it to two deliberately.
- Anything outside this spec's Tasks: a file the tasks do not name, an adjacent bug, a refactor, a guard
  for a case no acceptance criterion demonstrates, or a new dependency.

**Never:**
- **Never commit, `git add`, stash, checkout, reset, revert, or restore. The orchestrator makes every
  commit.** Leave all work in the working tree and report what to commit. Reading git state is fine and
  encouraged — audit the orchestrator's commits; several agents have. One carve-out: `git init`/`add`/
  `commit` inside a `mkdtempSync` dir under `os.tmpdir()`, removed in `finally`.
- **Never create a branch. Never push. Never open a pull request.** Step-05 offers all three; decline all
  three.
- **Never `git add -A`.**
- Never build the multi-cut fetch. `parseFamilyMetadata` already parses every upstream
  `fonts { style, weight, filename }` block and `fetchWebFamily` discards it — that is registered, not
  built (see Design Notes).
- Never touch the canvas paint path or the B / I controls. 11.3 shipped those and their wording is reused
  verbatim, not re-authored.
- Never re-author `starter.folio`. It already declares its variants at version 2.0.
- Never widen the variant key set. It is CLOSED at `bold`, `italic`, `boldItalic`, and extending it later
  is a MAJOR change the format doc already prices.
- Never make a green `test:e2e:compile` read as coverage. It is a TYPECHECK; the suite is exercised at
  **epic boundaries**, not in CI and not per story (D-000.30).
- Never pin a predicate against itself, and never ship an assertion whose two sides could be equal
  (D-11.2.8).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Pick a shipped family | Author picks `Roboto` in the family control | A chain named `Roboto` is declared naming the **shipped** face and its three cuts, **followed by the same proposed fallback tail the embed path computes** — the shipped faces for the scripts the picked face does not cover, each declaring its own cuts. **The declare path and the embed path compute that tail by the same code, not by two implementations that agree today**; a pick must never yield a chain with less script coverage than the path it replaced. For Roboto (`scripts: ["latin"]`) that is exactly `starter.folio`'s three-entry chain — `Roboto` + `Noto Sans Thai` + `Noto Sans SC`. **No asset is embedded**; the document's `assets` map is untouched | N/A |  <!-- AMENDED BY THE ORCHESTRATOR after approval, 2026-09-06, per the `patch` ruling on the dropped fallback tail: the original row specified the cuts and omitted the tail, and the implementer built exactly what it said. Only a human may reopen a frozen block; the builder correctly declined. -->
| The picked family bolds | That element then declares `style.bold` | It paints and renders in `Roboto Bold` — no Warning, because the cut is declared | N/A |
| Pick a catalogue family | Author picks `Inter` (31 catalogue faces, all Regular) | Still embeds, as today. The picked entry declares **no cuts** — it has none | N/A |
| The tail declares its own cuts | That same pick's proposed tail | `Noto Sans Thai` carries `bold`; `Noto Sans SC` carries none and stays a bare string | N/A |
| Absent cut is declared absent | Bolding a catalogue-face element | The B control states the absence in **11.3's own words**, and the engine emits AC3's Warning | Warning, render succeeds |
| **`Noto Sans SC`, ordinary behaviour** | A CJK run in a chain whose SC entry declares nothing, element declares bold | Renders SC Regular + AC3's Warning. **Permanent shipped condition (D-A), not an edge case** | Warning, render succeeds |
| A pick never writes self-reference | Any pick, any family | No variant written equals its own base face — otherwise the pick authors a document this story makes unloadable | N/A |
| **Self-referential variant, face arm** | `{"face":"Roboto","bold":"Roboto"}` | **Located load error** at `fonts.<chain>[<i>].bold`, message derived from the closed set | `fonts.<name>[<i>].bold` |
| **Self-referential variant, embedded arm** | `{"asset":"k","bold":"k"}` | Located load error at `fonts.<chain>[<i>].bold` | `fonts.<name>[<i>].bold` |
| **Cross-variant collision stays LEGAL** | `{"face":"Roboto","bold":"X","italic":"X"}` | **Loads and renders.** The narrowing must not quietly become "no two variants may agree" | N/A |
| Self-reference the check cannot see | `{"face":"Roboto","bold":"Roboto Copy"}`, two FontSet keys, identical bytes | Accepted. **DW-245** — string equality cannot see it, and detecting it needs binary reading D-11.2.1 forbids. Disclosed in code and in the format doc | N/A |
| No migration | A document whose entries predate this story is opened | Nothing is rewritten, repaired or normalised. It reports honestly (I-5) | N/A |
| Untouched document | Author never touches the family control | Serialized bytes, version string and rendered bytes all unchanged (AD-21) | N/A |

</frozen-after-approval>

## Code Map

Every anchor **measured at `7cc02d1`** with a clean tree, and `baseline_commit` stays there per the
workflow's preserve rule. Two commits landed during planning and **both are `_bmad-output`-only with zero
code delta**, verified rather than assumed: `887b288..7cc02d1` touches only `deferred-work.md` (DW-238's
discharge) and `7cc02d1..447b307` only `epic-11-14-decision-log.md` (D-11.4.2, the owner ruling) —
`git diff --name-only <a>..<b> | grep -v '^_bmad-output/'` is empty for both. **Implementation dispatches
from `447b307`; every anchor below holds at all three commits.**
**D-11.3.4: a Code Map is measured before the story edits the files it anchors. Cite by symbol; re-derive
every line before editing it, including the ones below.**

### The picking surface, and why exactly one family is affected

- `folio-designer/src/App.tsx:2978` `FontFamilyProperty`, rendered at `:2084`. Two groups only
  (`groups`, `:3090`) over one flat `matches` array (`:3086`): **`IN THIS TEMPLATE`** = `declared`, from
  `canvas.fontFamilies`; **`AVAILABLE LOCALLY`** = `onThisMachine` (`:3058`) =
  `offeredFamilies(query, storedFaces).filter(s => !families.includes(s.family) && familyIsInstalled(s))`.
- `folio-designer/src/App.tsx:3142` **`choose` — THE FORK, and the site this story edits.**
  ```ts
  if (match.source) { void commitFirstUse(match.source); return }
  void commit({ field: 'fontFamily', operation: 'set', value: match.name })
  ```
  A **declared** name is a property commit and writes **no chain entry at all**; a `source`-bearing row
  goes to `commitFirstUse` (`:3114`) → `onUseFamily` → `dispatchEmbed`.
- `folio-designer/src/App.tsx:1208` `dispatchEmbed` — **the one embed dispatch.** `:1209` computes the
  tail: `scriptFallbackFaces.filter(([script]) => !face.scripts.includes(script)).map(([, shipped]) => shipped)`.
- ⚠ **THE POPULATION, MEASURED — this is why AC1's non-vacuous set has one member.**
  `catalogueFaces` (`src/generated/font-catalogue.ts`) is **31 faces, ALL `style: "Regular"`, one per
  family** (extracted and counted; `build-wasm.mjs:428` hardcodes `style: "Regular"` — it is not in the
  source JSON). Script coverage: **29 latin-only, 2 thai-only, 0 cjk**. The `stored` tier holds only what
  `fetchWebFamily` produces, which takes `regularFilename(metadata)` and hardcodes `style: 'Regular'`.
  **Of the four shipped base families, only `Roboto` appears in the catalogue at all** — `Noto Sans`,
  `Noto Sans Thai` and `Noto Sans SC` are uncatalogued (D-11.1.5) and the control never offers them.
- ⚠ `folio-designer/public/fonts/roboto/Roboto-Regular.ttf` and `folio-go/fonts/roboto/Roboto-Regular.ttf`
  are **byte-identical**, `e688a215e0841b6e4edb1207c93f88f4c609f82e870884349a7257e449eb9355`. That is not
  luck: `folio-go/fonts/fonts_test.go` `TestShippedRobotoMatchesDesignerCatalogue` makes *"there is exactly
  one Roboto"* machine-checked (Story 16.8; `robotoCuts` at `fonts_test.go:51` carries all four).
  **It is a one-family invariant and it is NOT the criterion** — see Boundaries.

### The declared mirror — the product's first family→cuts table

`Shipped()` (`folio-go/fonts/fonts.go:159`) is a **flat** 11-key map whose own comment (`:156`) reads
*"NEVER PARSE ONE OF THESE KEYS (D-B). They are readable strings for a human writing a chain, not an
encoding."* The grouping exists in **exactly one place in the tree** and it is document data:
`folio-designer/public/templates/starter.folio:10-11`. (Population: all `.ts/.tsx/.go/.mjs/.py/.json/
.folio` outside `node_modules`/`dist`, searched for `Roboto Bold` and again for `boldItalic`. Every other
hit is a flat list, a prose comment, a test fixture or the projection type. Positive control: the same
search *does* return the starter's grouping.)

So the mirror is **hand-written**, and its content is forced:

| base family | `bold` | `italic` | `boldItalic` |
|---|---|---|---|
| `Noto Sans` | `Noto Sans Bold` | `Noto Sans Italic` | `Noto Sans Bold Italic` |
| `Noto Sans Thai` | `Noto Sans Thai Bold` | — | — |
| `Noto Sans SC` | — | — | — |
| `Roboto` | `Roboto Bold` | `Roboto Italic` | `Roboto Bold Italic` |

**4 bases + 7 cuts = 11 = `Shipped()`'s key count exactly.** That equality is the tie, and it is asserted
both ways.

**The repo's established idiom for a two-language list tie** — copy it, do not invent one. The strongest
instance is `folio-designer/src/engine-bounds-mirror.test.ts`: `goSources` (`:45`) reads **four Go files as
text**, `pairs` (`:86`) declares each Go constant with its TS mirror **and the call sites that must consume
it**, `it('finds every declared pair on both sides…')` (`:114`) is the **non-vacuity gate**, and
`it('holds every Go bound and its TypeScript mirror at the same number')` (`:133`) is the tie. The
shipped-face precedent is `canvas-font-stack.test.ts` `shippedFaceNames()` (`:86`), which parses
`Shipped()` out of `fonts.go` with `/func Shipped\(\) folio\.FontSet \{[\s\S]*?\n\}/` then
`"([^"]+)":\s*\w+,`.
⚠ **D-11.1.24's trap, second instance:** that parse is **test-side** and unimportable by production.
Production must hold its own declared copy; a **test** ties it to Go. Pointing production at a Go-source
parse is the exact AC 11.1 had to reword.

**Two existing tables already encode the grouping and must ASSERT the mirror rather than BE it:**
`folio-go/shipped_faces_test.go:143` `shippedFaceSpecs` (11 rows, `Key`/`Family`/`Subfamily`, verified
against name ID 1/2/6 and `usWeightClass` **read off the binary**), and
`folio-designer/src/font-catalogue.test.ts:625` `shippedSlotFaces` (13 rows,
`{cssFamily, family, subfamily, usWeightClass, bold, italic, oblique}`, also binary-verified). Both are
**test-only** and neither can be imported by production.

⚠ **D-8.4.14 rejected a table of this shape once**, and a reviewer will cite it:
`src/shipped-face-family.ts:19-20` says *"a face-name → CSS-family table (a second authority maintained in
lockstep with `fonts.Shipped()`). **A table is what this module is NOT.**"* That table answered **naming**;
this one answers **grouping**, and the difference must be written down beside it or the objection lands.

⚠ `src/canvas-font-stack.test.ts:1043` is a **closed six-entry allowlist** of every expression permitted in
a `font-family` position across all designer sources. The new module must not add a seventh — it never
reaches a font-family position. And `:288`
`shippedFaceDerivedFamily = /^shippedFaceFamily\(fragment\.face\)$/` is anchored to the identifier
`fragment` **by name** precisely because a chain **entry** has a `face` too. **Never put a chain entry in a
font-family position.**

### Where the declared table lives, and the generated-module trap

The closest precedent is `scriptFallbacks` → `scriptFallbackFaces`: declared by hand at
`folio-designer/scripts/build-wasm.mjs:227`
(`const scriptFallbacks = { latin: 'Noto Sans', thai: 'Noto Sans Thai', cjk: 'Noto Sans SC' }`), emitted at
`:426` into `src/generated/font-catalogue.ts`, guarded by three build-time throws (`:255`, `:257-258`,
`:269`), and consumed **on the pick path** at `App.tsx:1209`.

⚠ **But `src/generated/` is gitignored.** `git ls-files folio-designer/src/generated/` returns exactly one
file, `pdfjs-assets.ts`; `font-catalogue.ts` is `.gitignore:101`. A generated table is invisible in the
diff and unreviewable. `src/build-wasm.test.ts:290` also asserts **exactly six emissions** and would red on
a seventh (its own red-proof is at `:398`).
✅ **`npm test` = `npm run build:wasm && vitest run` and `npm run typecheck` = `build:wasm && tsc -b`**, so
generated modules do exist for every gate in the cadence. That removes the *existence* objection but not
the *reviewability* one.
**RECOMMENDED: a hand-written, git-tracked production module in `folio-designer/src/`.** It is the artifact
the mirror test ties to Go, it appears in the diff, and it needs no new emission. See Design Notes for the
alternative and why it loses.

### The wire — **RULED: route C** (the fork and its reasoning are in Design Notes)

- `folio-go/component_commands.go:1504` `componentFields(raw, want)` — the guard is
  `if len(raw) != want`, an **exact count** over an already-deduplicated map
  (`refuseDuplicateCommandKeys`, `:103`, runs first), counting `kind` and `version`.
- ⚠ **"An optional argument is structurally inexpressible" is FALSE, and both prior measurements missed
  the counter-example.** `folio-go/component_commands.go:2551` `tableCommandOp(raw, id, path, base)`
  implements **discriminant-driven arity**: `componentFields(raw, base+1)` at `:2558` for `op == "set"`,
  `componentFields(raw, base)` at `:2567` for `op == "clear"`. Callers `setTableAltRowBackground` (`:2650`,
  base 4 → **4 or 5**) and `updateTableHeaderStyle` (`:2702`, base 5 → **5 or 6**). The precedent is
  complete end-to-end: `folio-designer/src/table-style-command.ts:91` `operationFields` emits a
  **variable number of fields** through the authority and is green under every factory list.
- `folio-go/component_commands.go:3007` `addFontChain` — `componentFields(raw, 4)` at `:3008`; refuses a
  taken name at `:3015-3017`; `entries` unmarshalled as **`[]string`** at `:3022`; builds
  `template.FaceEntry(face)` at `:3052`.
- `folio-go/component_commands.go:3260` `embedFontFamily` — `componentFields(raw, 12)` at `:3261`; refuses
  a taken name at **`:3355-3357`**, `"a font chain named %q already exists"` (DW-238's `:3356` re-derived
  and still exactly right); builds `template.AssetEntry(key)` at `:3389` and `template.FaceEntry(face)` per
  tail item at `:3391`.
- `folio-go/component_commands.go:3480` `embeddedFontTail` — unmarshals `tail` into `var tail []string` and
  refuses anything else. Its comment at `:3470-3477` states the reason: *"every entry it can express is a
  FACE NAME, and the ONE entry that names an asset is the one this command builds itself."* **That
  reasoning survives every route below and is the structural guarantee that a pick cannot write an asset
  sibling.**
- `folio-go/component_commands.go:3483` — the codebase's own idiom for optional content in a mandatory key:
  *"the proposed fallback tail is required — write [] for a face that needs none."*
- `folio-designer/src/font-chain-command.ts` — seven builders, `:34`–`:77`. **Only `embedFontFamilyCommand`
  (`:77`) has a production caller** (`App.tsx:20` import, `App.tsx:1210` call). The other six are
  uncalled since Story 16.9 deleted `FontChainEditor.tsx`.
- `folio-designer/src/font-chain-command.test.ts:27-42` — the `arity` table,
  `[payload, fieldCount, orderedKeys]`, asserted with `toEqual` on `Object.keys`. **Any wire change adds or
  edits a row here.**
- `folio-go/wasm/engine.go:210` `Apply` — branches only on `kind == "pageSetup"` (`:228`) and forwards
  opaque bytes otherwise. **A new component command kind needs zero change at the wasm boundary**, and
  there is no central registry of command kinds.

### The factory lists — THREE, not two, and one has already drifted

- `folio-designer/src/command-json-soleness.test.ts:98` and `:119` — **nine** entries each, spelled twice
  deliberately; the file says at `:93-96` and `:114-116` that both must move together.
  `productionFiles` (`:35`) is a **recursive readdir** of every non-test `.ts`/`.tsx`, so `:88`
  (`commandJsonBuilders(sources())` must equal `[AUTHORITY]`) covers every production file whether or not
  it is named.
- ⚠ `folio-go/command_json_authority_wire_test.go:180`
  `TestEveryDesignerCommandFactoryRoutesThroughTheAuthority` lists **six** (`:181-188`), missing
  `band-height-command.ts`, `document-settings-command.ts` and `table-style-command.ts`. It is **green**,
  because it only ever checks the six it names.
  **The drift is a genuine FALSE NEGATIVE, not mere under-reporting, and here is the measured hole:** the
  `charCodeAt(` check at `:208` runs **inside** that six-name loop, and **no designer-side test scans
  factories for `charCodeAt(` at all** (searched every designer `*.test.ts*`; the only hits are prose and
  fixture decoding — positive control: `charCodeAt` appears in six designer files, so the search finds it).
  So for those three factories a hand-rolled escape table is seen by **nothing**. `component-asset-command.ts`
  does contain `charCodeAt` but only at `:10`, inside a line comment, which is why the guard strips line
  comments and why the suite is green. **REGISTER; fix in-story only if this story touches that list.**

### DW-241 and DW-245 — the parse narrowing

- `folio-go/internal/template/parse.go:464` `decodeFontChainEntry`. Arm chosen by first non-space byte
  (`:467-470`, `:479`, `:583`). Exactly-one-of at `:488-495`. Closed-set scan at `:499-506` over
  `slices.Sorted(maps.Keys(entryObj))` — sorted so the *named* key is deterministic. `entry.Face` is set at
  **`:517`**, `entry.AssetKey` at **`:534`** — **both before the sibling loop**, which is what makes one
  check site serve both arms.
- **The sibling loop is `:540-581`**, `for _, v := range fontChainVariants`, with `sibField := field + "." + v.key`
  built at `:546`, the empty check at `~:558`, the namespace block at `:563-579`, and the write
  `*v.field(&entry) = name` at **`:580`**.
  **PUT THE SELF-REFERENCE CHECK AFTER THE EMPTY CHECK AND BEFORE THE NAMESPACE BLOCK.** A self-reference
  is trivially in-namespace, so the namespace check can never fire on it; placing it first also avoids a
  redundant second `requireEmbeddedFaceLicence` call on the same asset. The comparand is `entry.AssetKey`
  when `entry.Embedded()`, else `entry.Face`.
- **The enumeration — `fontChainVariants`, `folio-go/internal/template/model.go:235`**, declared *"THE
  authority for the closed variant set"*, projected by `fontChainVariantKeys()` (`:247`).
  **The formatter is `quotedKeyList`, `parse.go:387`**, output `"bold", "italic", "boldItalic"`. Existing
  derived-message helpers: `fontChainEntryGrammar()` `:405`, `noDiscriminantReason()` `:424`,
  `bothDiscriminantsReason()` `:431`, `unknownEntryKeyReason(key)` `:441`. **The new refusal joins these —
  it does not spell the keys again.**
- `folio-go/internal/template/parse.go:671` `requireEmbeddedFaceLicence` — already runs for every variant
  asset key (called at `:573` inside the sibling loop, plus `:531` for the entry's own `asset`).
- `folio-go/internal/template/model.go:272` `SerialisesAsObject()` — true for
  `{"face":"Roboto","bold":"Roboto"}` today. **After the narrowing it can never be parsed, so no serializer
  or version path can observe it. The predicate needs no change.**
- `folio-go/embedded_face.go:284` — **DW-241's second finding CONFIRMED.** `FontChainSite` records are
  appended inside `for _, assetKey := range entry.EmbeddedAssetKeys()`, and `EmbeddedAssetKeys()`
  (`model.go:310`) appends `e.AssetKey` then each non-empty variant **with no de-duplication**, so
  `{"asset":"k","bold":"k"}` yields two byte-identical `FontChainSite` records. **It is inert** — `siteIn`
  (`:166`) returns the first match, nothing counts or iterates for output — **and the narrowing makes it
  unreachable.** No separate fix; say so.
- **Regression scan, run before planning:** searched all `.go/.json/.folio/.ts/.tsx` outside
  `node_modules`/`.git`/`_bmad-output` for `"(bold|italic|boldItalic)"\s*:\s*"` — 38 hits, and **no
  fixture, test document or shipped file declares a variant equal to its own base**, `starter.folio`
  included. Positive control: the same search returns
  `style_face_resolution_test.go:324` (`{"face": "Noto Sans SC", "bold": "Noto Sans Thai"}`), so it would
  have found a self-referential pair. **The narrowing breaks no existing document.**

### The format doc, and the guard that reads it

- `_bmad-output/specs/spec-folio/folio-format.md` — `` ## `fonts` `` at **line 182**, the **only**
  occurrence; the section ends at the next `^## ` (`` ## `bands` ``, **line 299**). The chain-entry table
  header is at `:225`, `` `face` `` at `:227`, `` `asset` `` at `:228`, and **the variant row at `:229`**.
  Namespace prose at `:231-236`; entry-level refusals at `:254-258`.
  **DW-245's disclosure goes in the second cell of `:229`**, or as a paragraph after `:236`.
- `folio-go/internal/template/font_chain_variants_test.go:36`
  `TestTheVariantKeySetIsClosedAndTheFormatDocSaysSo` — reads the doc, requires **exactly one** line equal
  to `` ## `fonts` `` (counting, not an awk range — the re-trigger trap is already closed), slices to the
  next `## `, and per key inspects **only the FIRST table cell** via `strings.SplitN(trimmed, "|", 3)`.
  **Two constraints on the disclosure: the first cell must stay `` `bold`, `italic`, `boldItalic` ``, and
  the sentence must contain no raw `|`.** Adding prose to the second cell does **not** break it.
- `folio-go/internal/template/drift_test.go:242` / `:279` — `extractDocKeys` (`:129`) harvests backticked
  tokens from a row's **first cell only** (`:185-189`). A disclosure in the description cell adds no keys;
  **a new table row whose first cell backticks a non-serializer token would red `TestDriftDocToGo`.**

### Where the tests go

- **The two self-reference rows** → `folio-go/internal/template/fonts_embedded_test.go:763`
  `TestADefectiveVariantIsRefusedAtTheSiblingItself` — the per-sibling `.<key>` field table (5 rows today).
  ⚠ Its helper `requireLoadError` (`:121`) asserts the **field only**, never the message, so a
  **message** assertion needs its own test beside `TestAnUnknownEntryKeyIsRefusedByTheCLOSEDSetAndSaysSo`
  (`:714`), whose derived-key loop at `:718` is the pattern to copy.
- **The cross-variant-collision positive direction** → parse/serialize half beside
  `font_chain_variants_test.go:190` `TestTheObjectFormRoundTripsAsAFixedPoint`; **render** half in
  `folio-go/style_face_resolution_test.go`, which already drives real renders over object-form chains
  (`:297`, `:527`).
- **What a pick writes** → `folio-designer/src/App.test.tsx:2808`,
  `it('embeds and then commits the property, as two commands, when a family this machine holds is picked')`
  inside `describe('typography controls over the engine-projected closed sets')` (`:2446`). It inspects the
  **command bytes**, and `:2861` already asserts `payload['tail']` equals `['Noto Sans Thai', 'Noto Sans SC']`
  — **that is the exact assertion the tail's cuts extend.** Other payload tests live in
  `App.font-store.test.tsx` via `embedPayloads()` (`:171`).
  ⚠ **Nothing in the designer suite inspects a resulting document.** The closest, `robotoChain`
  (`App.test.tsx:5855`), is a **hand-built projection**, not a document the designer produced.
- **The document-level assertion** → `folio-go/starter_template_test.go` is the file that already does this
  shape of thing. `TestTheShippedStarterDeclaresOnlyCutsTheEngineSupplies` (`:61`) checks non-vacuity by
  count (`:83`), that every face and variant is a `Shipped()` key (`:90`, `:99`), and **already errors when
  `variant.name == entry.Face`, citing D-11.2.11 and naming this story** (`:102`).
  `TestANewDocumentFromTheStarterCanActuallyBold` (`:161`) is the end-to-end shape to copy — it derives its
  expectation from the file rather than restating it (`declaredBoldOf`, `:225`).
- **e2e** `folio-designer/e2e/font-embed-boundary.spec.ts:241`, `:327` drive the real dropdown against the
  real starter but derive expectations from `Object.keys(starter.fonts)` only (`:76-85`), **so variants are
  invisible to it**. Compile-only; `test:e2e:compile` must stay green.

### Guards that must stay green untouched

`canvas_projection_wire_test.go:200`/`:414` (the chain-entry wire tie); `canvas-font-stack.test.ts`
`:796` `toBe(11)`, `:819` `toHaveLength(10)`, `:1043` the six-entry allowlist; `font-catalogue.test.ts:461`
(*"every catalogue face as a single upright static Regular"*) and `:524-537` (`@font-face` rules carry no
`font-weight`/`font-style` **descriptor**); the 11.1 shipped-face guard table
(`shipped_faces_ext_test.go:72`, `shipped_faces_test.go:525`, `fontgen_matrix_test.go:264`
`wantDerivedShippedFaces = 7`, the designer `toBe(11)`/`toBe(13)` counts); `build-wasm.mjs`'s three
build-time throws at `:255`, `:257-258`, `:269`; `src/build-wasm.test.ts:290` (exactly six emissions).
**Any move means this story drifted.**

## Tasks & Acceptance

**Execution:**

- [x] `folio-designer/src/shipped-face-cuts.ts` (new) -- Declare the family→cuts mirror as **hand-written
  production data**: four base families, seven cuts, the table in the Code Map. State in a comment that
  this is the **only** place the grouping is declared, and why it is **not** the face-name→CSS-family table
  D-8.4.14 refused — that one answers *naming*, this answers *grouping*. The name deliberately sits beside
  `shipped-face-family.ts`, whose `:19-20` carries that refusal, so a reader meets both at once.
  -- Or the next story writes a second one.
- [x] `folio-designer/src/shipped-face-cuts.test.ts` (new) -- **The mirror tie, copying
  `engine-bounds-mirror.test.ts`'s idiom**: a non-vacuity gate that both reads found something, then
  **bases + cuts == `Shipped()`'s key set asserted BOTH ways**, reading `fonts.go` as text the way
  `canvas-font-stack.test.ts:86` does. Tie it to `shippedSlotFaces` (`font-catalogue.test.ts:625`) as the
  binary-verified witness. **Red-prove by mutating a synthetic copy of the Go source**, never the real one.
  -- A one-directional tie cannot see a face that leaves the FontSet.
- [x] `folio-designer/src/App.tsx` -- In `choose` (`:3142`), route a pick whose family is a member of the
  mirror to a **name-and-declare** path instead of `commitFirstUse`'s embed: declare the chain naming the
  shipped face with its cuts, **followed by the same proposed fallback tail the embed path computes**, then
  commit the property, still as **two** commands and two undo entries (`:3120-3140` records why fusing them
  is refused). Leave the declared-name branch and the catalogue branch otherwise untouched.
  ⚠ **The tail is computed by ONE function both paths call, not by two implementations that agree today.**
  `dispatchEmbed` already computes it; extract and share it. Two copies of one rule is exactly the shape
  that produced 11.2's version defect — `writeFontChain` and `fontsRequireMajor` held two copies of
  `entry.Embedded()` that agreed until an input separated them, silently. **The testable property is that a
  pick never yields a chain with less script coverage than the path it replaced**, and that can be asserted
  without knowing what the correct tail is.
- [x] `folio-designer/src/App.tsx` -- In `dispatchEmbed` (`:1208`), give the **proposed tail** its declared
  cuts from the mirror. The picked (embedded) entry declares none — it has none. -- AC2, and it is required
  under either route.
- [x] `folio-designer/src/font-chain-command.ts` + `folio-go/component_commands.go` -- **ROUTE C, ruled at
  CHECKPOINT 1.** `addFontChain`'s `entries` (`:3022`) and `embedFontFamily`'s `tail` (`embeddedFontTail`,
  `:3480`) stop being `[]string` and take **the format's own chain-entry shape**: a bare string, or an
  object with `face` plus the closed variant keys. **No arity constant moves** (`addFontChain` stays 4,
  `embedFontFamily` stays 12); the builder stays in this **existing** file so no factory list moves.
  **The object form MUST refuse `asset`** — that is what makes "variant keys are face names, never asset
  keys" enforced by the decoder instead of remembered by a person, and the
  `assetKeyReferenced` out-of-scope ruling rests on it.
  ⚠ **`embeddedFontTail`'s comment (`:3470-3477`) states the INVARIANT and then its mechanism. Preserve
  the reason verbatim and replace only the mechanism clause**, so a reader sees the invariant survived
  rather than that it was rewritten. `[]string` was never the point; *"a caller cannot put a second asset
  entry in a chain by writing one down"* is.
  ⚠ **Do not lose `"AN EMPTY TAIL IS LEGAL"`.** Its named witness is
  `TestLoadNeitherResolvesNorRefusesAnEmbeddedEntry` — assert it still holds under the new shape.
- [x] `folio-designer/src/font-chain-command.test.ts` -- Update/extend the `arity` table at `:27-42`
  (`[payload, fieldCount, orderedKeys]`) for the changed wire shape. **The field counts must not move** —
  route C changes a field's shape, not the arity.
- [x] `folio-go/component_commands_test.go` -- **The 15.2a obligation, made testable** (ruling in Design
  Notes): for the command shapes that exist today, the accept/refuse **verdict is unchanged at every
  input**. Take a representative refusal that exists today on an **existing** field set — an arity
  violation, and a malformed entry from `TestFontChainAddRefusesEveryMalformedEntryList` (`:1629`) — and
  show it **still refuses** after C. **Not vacuous:** the point is that widening a field's *shape* did not
  soften a *verdict*, so a case that only shows the new shape being accepted proves nothing about it.
- [x] `folio-go/internal/template/parse.go` -- **DW-241.** In `decodeFontChainEntry`'s sibling loop
  (`:540-581`), after the empty check and before the namespace block, refuse a variant equal to that arm's
  own discriminant (`entry.AssetKey` when `entry.Embedded()`, else `entry.Face`), located at `sibField`.
  **Derive the message from `fontChainVariantKeys()` via `quotedKeyList`** — join the existing
  `*Reason()` helpers, do not spell the keys. **Write DW-245's limit in a comment beside the check.**
- [x] `_bmad-output/specs/spec-folio/folio-format.md` -- Add the self-reference refusal and **DW-245's
  limit** to the chain-entry row's **second cell** at `:229` (or a paragraph after `:236`). **Keep the
  first cell exactly `` `bold`, `italic`, `boldItalic` `` and use no raw `|`** — `font_chain_variants_test.go:36`
  reads the first cell only, and `drift_test.go`'s `extractDocKeys` harvests from it.
- [x] `folio-go/internal/template/fonts_embedded_test.go` -- Add the two self-reference rows to
  `TestADefectiveVariantIsRefusedAtTheSiblingItself` (`:763`), face arm and embedded arm. Add a **message**
  test beside `:714` (derived-from-the-closed-set), because `requireLoadError` (`:121`) asserts the field
  only and cannot see the wording.
- [x] `folio-go/internal/template/font_chain_variants_test.go` + `folio-go/style_face_resolution_test.go`
  -- **The other direction, and it is not optional (D-11.2.11):** a cross-variant collision
  `{"face":"Roboto","bold":"X","italic":"X"}` **loads, round-trips and RENDERS**. Without it the narrowing
  quietly becomes "no two variants may agree."
- [x] `folio-designer/src/App.test.tsx` -- Extend the pick tests at `:2808`/`:2861`: a **shipped-family**
  pick declares the cuts and embeds **nothing** (`assets` untouched); a **catalogue** pick still embeds and
  its **tail** now declares its cuts. Assert **both directions** — a test that only checks the shipped case
  cannot see the catalogue case regress.
- [x] `folio-go/` tests -- **AC1's real assertion, at document level:** pick a family, **change** it, and
  confirm the variants are present in the resulting document. Copy
  `starter_template_test.go:161`'s shape — derive the expectation from the document, never restate it.
  Include a `Noto Sans SC` case as **ordinary behaviour**: bold a CJK run, get SC Regular plus AC3's
  Warning. And assert **no pick writes a variant equal to its base**, which this story's own parse change
  would make unloadable.
- [x] `_bmad-output/implementation-artifacts/deferred-work.md` -- Register **three** entries, none of them
  in scope, each with the measurement that makes it cheap to pick up: **(a)** the multi-cut fetch — frame
  it as *the* load-bearing entry for "bold works for the fonts people actually pick", and record that
  `parseFamilyMetadata` (`font-source.ts:138`) **already parses every upstream `fonts { style, weight,
  filename }` block** and `fetchWebFamily` discards it, so the capability is **half-built**; **(b)** the
  duplicate embed this story closes, left **open in one direction** — a *fetched* face byte-identical to a
  shipped one still duplicates, and detecting that needs the binary comparison D-11.2.1 refuses, so state
  the limit rather than letting the fix read as complete; **(c)** the drifted Go factory list, with the
  measured `charCodeAt(` false negative. Also register **the no-self-contained-copy limit**, re-priced by
  the first request for a `.folio` that renders against a foreign FontSet. Use `### DW-<n>` headings — the
  workflow's raw `- source_spec:` block form is invisible to a `### DW-` census (D-11.0.2).

**Acceptance Criteria:**

- Given an author picking a family the release already ships, when the pick is applied, then the chain
  entry it writes **names the shipped face and declares that family's cuts**, no asset is embedded, and the
  document's `assets` map is untouched.
- Given that element then declaring `style.bold`, when the document renders, then it renders in the
  declared bold cut with **no** AC3 Warning — the cut was declared, not inferred.
- Given an author picking a family the release does **not** ship, when the pick is applied, then it still
  embeds, the picked entry declares **no** cuts because it has none, the **proposed tail** declares the
  cuts its shipped faces have, and the B or I control states the absence **in Story 11.3's own words**.
- Given a chain entry whose variant names its own base face, on **either** arm, when the document is
  loaded, then it is a located load error at `fonts.<chain>[<i>].<key>` whose message is **derived from the
  closed set**. **And given a cross-variant collision, then it LOADS AND RENDERS** — both directions, or
  the narrowing has silently become something wider than it was ruled.
- Given any pick this story can produce, when the resulting document is reloaded, then it loads — no pick
  writes a variant equal to its own base face.
- Given a document whose chain entries predate this story, when it is opened, then nothing is migrated,
  rewritten or repaired; it reports honestly (I-5).
- Given a document whose author never touches the family control, when it is rendered, then its serialized
  bytes, its version string and its rendered bytes are unchanged on all four targets (AD-21).

## Spec Change Log

**2026-09-06 — the Matrix Test Audit found one uncovered row, and it was the DISCLOSURE row.** Matrix row
*"Self-reference the check cannot see"* (`{"face":"Roboto","bold":"Roboto Copy"}`, two FontSet keys holding
identical bytes → **Accepted**, DW-245) had **no covering test**. The limit existed only as prose — a
comment at `parse.go:597` and a paragraph in `folio-format.md` — and prose is not a covering test. Searched
`Roboto Copy` across `folio-go` and `folio-designer/src`: one hit, the comment. Positive control: the same
search idiom over `folio-go/internal/template` for `Roboto Bold` returns three files, so it finds what is
there.

**Why this row was the one to slip, and it generalises.** Every other row asserts something *happens*; this
one asserts something *does not happen and is disclosed instead*. There was nothing to red, so nothing was
written — which is exactly the shape DW-245 warns about, arriving in the guard for DW-245 itself.

**Fixed by `TestTheSelfReferenceCheckIsSTRINGEQUALITYAndSaysSo`**
(`folio-go/internal/template/font_chain_variants_test.go`): the nearest legitimate spelling **loads**, a
positive control shows one edit (`"Roboto Copy"` → `"Roboto"`) refuses the same document, and **both
required statements of the limit are pinned** — neither had any guard at all, because the doc guard reads a
table row's first cell only and nothing reads the parse.go comment. Mutation-proved by **deletion**, three
ways: removing the check reds the control, and removing either disclosure reds its site. Both mutated files
were restored from `cp` snapshots and verified byte-identical by sha256 (`git checkout` was not used — it
would have destroyed the uncommitted implementation).

**Known-bad state avoided:** shipping DW-245's disclosure as prose that any later edit could delete
silently, leaving a check whose stated limit had quietly become an unstated one — the precise failure the
entry exists to prevent.

**2026-09-06 — review round 1: MY TASKS UNDER-SPECIFIED THE DECLARE PATH, and the implementer built
exactly what they said.** The review found that a shipped-family pick emitted a **one-entry chain**,
dropping the Thai and CJK fallback that both `starter.folio` and the embed path it replaced provide — so a
pick produced *less script coverage than the path it replaced*. The cause is in this section, not in the
code: matrix row 1 and the `choose` task both described "declare the chain naming the shipped face with its
cuts" and never carried the tail onto that path, while the tail clause went only into the `dispatchEmbed`
task. **The record should not read as an implementer error, because it was not one.**

**Triaged `patch`, not `bad_spec`, and the discriminator is worth keeping** (orchestrator ruling, review
round 1): *`bad_spec` exists for a spec defect whose consequences are **pervasive** — where the implementer
derived a chain of decisions from a wrong frame and patching one symptom leaves the frame wrong. The test
is whether the wrongness is local or pervasive, not whether the spec or the code is at fault.* This
omission is **additive and local**: nothing written was wrong, nothing else in ~1,545 lines was derived
from its absence, and the fix had **no design freedom** — `starter.folio` *is* the target chain and
`dispatchEmbed` three lines away already computed it. A re-derivation would have produced the same code
plus three lines.

**Amended:** the `choose` task now carries the tail clause, and the orchestrator amended the frozen matrix
row (the builder correctly declined to touch it). Both now require the tail to come from **one shared
function**, not two implementations that agree today — the same mirrored-invariant shape as 11.2's
`entry.Embedded()` version defect.

**KEEP on re-derivation:** the shared `proposedFallbackTail` extraction called by *both* paths; the Go-side
assertion that a Roboto pick's chain equals `starter.folio` **read off disk** rather than a restated
literal; and the coverage-not-reduced property, which is assertable without knowing the correct tail.

**Known-bad state avoided:** shipping a pick that silently narrows a document's script coverage — Thai and
CJK text losing its fallback — while every gate stayed green, because nothing compared the declare path's
output against the embed path's or against the product's own default document.

## Design Notes

### THE COMMAND-SURFACE FORK — **RULED AT CHECKPOINT 1: ROUTE C.** Apply, do not re-open.

**All three routes violated Story 15.2a's stated norm in letter, so that norm did not discriminate between
them.** 15.2a (`15-2a-….md:107-109`, reaffirmed `:669`) says *"the accept-set of both doors is narrowed or
unchanged at every input. No draft may be accepted that was refused before."* A new kind, a widened arity
and a widened field shape **all** accept a draft that hits `default:` or the arity gate today. **It is
ruled below rather than left to review.**

### THE 15.2a NORM — RULED at CHECKPOINT 1 (2026-09-06). Do not re-derive it at review.

**The norm binds the JUDGMENT OF EXISTING INPUT SHAPES, not the SET OF EXPRESSIBLE COMMANDS.** *"No draft
may be accepted that was refused before"* is a **hardening** rule: it forbids the doors becoming more
permissive about inputs **they already judge**. Read as a bar on the product ever gaining a capability it
would make 15.2a a permanent freeze on the command surface, which no hardening story can have meant and
which its own *"Valid command, unchanged"* matrix row (`:208`) contradicts.

**The obligation 11.4 inherits, stated so it is testable:** for every command shape that exists today, the
accept/refuse **verdict must be unchanged at every input**. A draft that hits `default:` because its kind
does not exist is not a judgment about a shape; a draft refused by an arity gate on an **existing** field
set is. **Assert the second class explicitly** — take a representative refusal that exists today and show
it still refuses after C.

### Why route C, and the deciding argument is not the cheap one

Points 1–3 below are **cost** arguments. **Point 4 is a kind argument, and it is why C won.** Under A or B,
*"variant keys are face names, never asset keys"* stays **a rule someone has to remember**, and the
`assetKeyReferenced` out-of-scope ruling rests on that rule holding. Under C **the decoder enforces it**,
so the ruling rests on the compiler. This run has spent most of its findings on the gap between a rule and
a guard; a route that closes that gap for free wins.

**C does not impose a new restriction — it re-implements an invariant the code already claims.** Today
`[]string` is the *mechanism* enforcing "no caller-written asset entry"; under C the object form refusing
`asset` is the mechanism, and **the property is unchanged.** That is the mechanism-versus-property
distinction D-11.2.2 used to admit the object form over the one-key rule, applied a second time — which is
why rewriting `embeddedFontTail`'s prose is an edit to the reasoning rather than a contradiction of it.

**The `variants` sidecar was refused on the spec's own ground:** it invents a second grammar for what the
format already says, which is exactly D-11.2.2's objection to `{"face":"X","variants":{…}}`. Consistency
with a ruling made three stories ago outweighs the arity saving.

| | **A — new command kind** | **B — widen arity** | **C — widen the FIELD's shape** |
|---|---|---|---|
| Go | new `case` + handler + `componentFields(raw, N)` | `addFontChain` 4→5, `embedFontFamily` 12→13 | **no arity constant moves** |
| TS | new builder in the existing `font-chain-command.ts` | two builders change | two builders change |
| Arity table `:27-42` | new row | two rows move | rows keep their counts |
| Factory lists | untouched (existing file) | untouched | untouched |
| wasm boundary | untouched (`engine.go:228` is kind-agnostic) | untouched | untouched |
| Serves BOTH paths? | **no** — `embedFontFamily`'s tail still needs its own change | yes | yes |

**RULED: C.** `addFontChain`'s `entries` and `embedFontFamily`'s `tail` stop being `[]string` and
become **the format's own chain-entry shape** — a bare string, or an object with `face` plus the closed
variant keys. Four reasons:

1. **One grammar, not two.** A parallel `variants` sidecar keyed by index invents a second way to say what
   the format already says — the same objection D-11.2.2 used to refuse `{"face":"X","variants":{…}}`.
2. **It serves both paths with one mechanism**, which A does not.
3. **No arity constant, no new kind, no new file, no factory list.** That keeps the drifted Go list out of
   scope exactly as ruled.
4. ⚠ **It makes the asset guarantee structural rather than disciplinary.** `embeddedFontTail`'s comment
   (`:3470-3477`) already says every entry a caller can express is a **face name**. If the object form
   admits `face` and the three variant keys and **refuses `asset`**, then a pick *cannot* write an asset
   sibling — which is the precondition the `assetKeyReferenced` out-of-scope ruling rests on. Under A or B
   that stays a rule someone must remember.

The one thing to weigh against C: `addFontChain`'s `entries` being `[]string` is load-bearing prose in two
places, and C rewrites it. That is an edit to the reasoning, not a contradiction of it.

### Where the mirror lives — and why not the generated module

`scriptFallbacks` → `scriptFallbackFaces` is the closest precedent and it is tempting: hand-declared beside
the shipped-face list, emitted into `font-catalogue.ts`, consumed on the pick path, guarded by build-time
throws. **It loses on one point: `src/generated/` is gitignored** (`git ls-files` returns only
`pdfjs-assets.ts`). A generated mirror never appears in a diff, so the one artifact this story exists to
introduce would be unreviewable — and the mirror is precisely the thing a reviewer must check by eye
against `Shipped()`. `npm test` runs `build:wasm` first, so existence is not the problem; **reviewability
is.** A hand-written tracked module also gives the tie test two *sources* to compare, which is what
`engine-bounds-mirror.test.ts` does.

### What this actually buys — say it in these words, in the Delivery Log

**Exactly one family becomes boldable by a pick: `Roboto`.** The three Notos are uncatalogued shipped faces
the control never offers; all 31 catalogue faces and every fetched face are Regular-only **by
construction**. So AC1 stops being vacuous, but **its non-vacuous population has one member** — and that
member is the family `starter.folio` already declares.

That is worth having: a document that re-picks Roboto keeps its bold, and **~348 KB of duplicate stops
shipping**. But it is **not** *"picked families can now bold"*, and the Delivery Log must say so plainly
rather than letting the epic read as delivering more than it did.

**This is an ACCEPTED LIMIT, not a shortfall discovered late — and the difference matters to how the log
reads.** The measurement above falsified how this story was described when the owner funded it (*"any
existing document whose author re-picks their family, which is most of them in practice"*). The
orchestrator took that to the owner rather than narrating it afterwards, and the owner re-decided with the
correct measurement in hand: **D-11.4.2, "ship as ruled, register the rest"** — Roboto boldable by pick,
the duplicate embed removed, DW-241's narrowing carried; **declined**, widening the picker to the shipped
Notos (it would reopen the UI decision 16.4/16.9 made) and scheduling the general upstream-cuts capability
now. So the Delivery Log records a limit the owner has seen and accepted, in the words above.

### DW-238 — DISCHARGED BY MEASUREMENT. No tasks. Do not "fix" it.

DW-238 said the chain commands destroy declared variants and made itself Story 11.4's hard precondition.
**There is no destroying code.** Measured independently here, at `7cc02d1`, and confirmed by the
orchestrator's own separate reading:

```
addFontChainEntry     Fonts[name] = slices.Insert(slices.Clone(chain), index, FaceEntry(face))   :3140
moveFontChainEntry    moved := slices.Clone(chain) … reorder                                     :3160-3163
removeFontChainEntry  Fonts[name] = slices.Delete(slices.Clone(chain), index, index+1)           :3188
addFontChain          refuses a taken name                                                       :3015-3017
embedFontFamily       refuses a taken name                                                       :3355-3357
```

**Every entry-mutating command clones the chain and copies each entry by value; both whole-chain
constructors refuse to overwrite.** Only a *newly inserted* entry is a bare `FaceEntry`, which is correct —
a new entry has no variants yet. DW-238 is wrong on three of its four evidence claims (`setFontChain` does
not exist; `addFontChain` cannot touch an existing chain; the three real chain edits all preserve), and its
own later paragraph contradicts its quoted evidence. **The register entry has been corrected by the
orchestrator; do not re-open it.**

**What survived the discharge is an obligation, not a bug:** the control must write variants **at the
moment it constructs the chain**, or the author loses them — not because old code destroys them, but
because new code would never write them. That is AC1, and it is why the pick test **changes** the family
rather than only setting it once.

### `assetKeyReferenced` — out of scope, on a test you can check

`assetKeyReferenced` (`component_commands.go:1003-1028`) tests only
`entry.Embedded() && entry.AssetKey == key` at `:1022` and **never calls `EmbeddedAssetKeys()`**
(`model.go:310`, whose only production caller is `embedded_face.go:277`). Since 11.2 a variant sibling of an
`asset` entry **is** an assets key, so the walk is blind to a class of reference. Two consequences:
`deleteFontChain` → `dropUnnamedFontAssets` (`:3216`) can delete an asset a variant still names, and the
reparse then dies as the **unlocated** `"font chains did not pass format validation"` (`:2933`); and
removing `{"asset":"K1","bold":"K2"}` orphans K2, against the function's own contract at `:3193-3194`.

**Ruled out of 11.4 on a checkable test: this story writes FACE-name variants, never ASSET-key variants**,
so it does not make the defect reachable — it stays hand-authored-only, exactly as today. **Verify that
holds against what you actually build. If any path writes an asset sibling, stop and report it; the ruling
flips.** Registered MEDIUM with both consequences.

### The starter is already the witness, and it already names this story

`folio-go/starter_template_test.go:102` **already errors when `variant.name == entry.Face`**, citing
D-11.2.11 and saying in terms that *"Story 11.4's parse check turns it into a load error on the shipped
starter."* So the starter is a live, executing precondition for the narrowing — and the narrowing must not
red it. The file derives every expectation from the document rather than restating it (`:110-133`,
`declaredBoldOf` `:225`); **copy that discipline, do not add a second copy of the starter's contents.**

## Verification

Run each module's commands in **its own** invocation — a `cd` persists through a compound command and makes
later relative paths resolve elsewhere, printing `lstat …` lines that read like passes. Use `-count=1` on
every Go run: the cache serves a stale PASS for tests that walk the filesystem. **Never `&&`-chain these**;
a conjunction silently drops everything after its first failing term, and `$?` after a pipe is the tail
command's.

**Commands:**

- `cd folio-go && go test -count=1 ./...` -- baseline **2204 pass / 2 fail / 5 skip**. The ONLY acceptable
  failures are `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child — a mandated permanent
  red, never to be "fixed". **A third distinct failure is a HARD STOP: report it, do not triage around
  it.** Count from `go test -json` `Action` events carrying a `Test` field; a plain run prints no totals
  and a carried-forward figure is not a measurement.
- `cd folio-go && FOLIO_FONTGEN_PYTHON=/private/tmp/claude-501/-Users-panitw-Projects-folio/c910e6ef-835c-4b93-9710-f11720b44f3a/scratchpad/fontgen-venv/bin/python go test -count=1 -tags=matrix ./...`
  -- **UNFILTERED, all four targets, in-story, because this story changes the FORMAT (DW-241 is a parse
  narrowing).** Without the env var `TestShippedFacesReproduceFromUpstream` degrades to a
  could-not-execute rather than a failure. Expect `fontgen: derived and compared 7 of 7 faces`.
  ⚠ **A `-run` filter here produces a green from tests that assert nothing (D-000.28).**
- `cd lint && go test -count=1 ./...` -- **227 pass**, four packages `ok`, exit 0.
- `gofmt -l folio-go lint` (from the repo root, absolute paths) -- **empty output**, exit 0.
- `cd folio-designer && npx tsc -b --force` -- exit 0. **`--force` is required**: `tsc -b` exits 0 without
  typechecking when its build info is current.
- `cd folio-designer && npm test` -- baseline **64 files / 965 tests**, all passing. (This runs
  `build:wasm` first, which is routine and not the release build.)
- `cd folio-designer && npx oxlint` -- exit 0 with **exactly 4** `only-export-components` warnings, all
  pre-existing (two in `pdf-viewer.tsx`, two in `App.tsx`). A fifth is this story's.
- `cd folio-designer && npm run test:e2e:compile` -- exit 0. **A green here is a TYPECHECK and not
  coverage**: the suite is exercised at **epic boundaries**, not in CI and not per story (D-000.30). Say it
  that way — the shorter "compiled and never executed" is false and would mis-scope Story 15.2.
- **`npm run build` is Ask First.** Do not run it without asking.

**Manual checks:**

- **Mutation-proof every new guard**, and do it by **deletion** where the guard is an arm, not by
  substitution: revert the self-reference check and confirm the two refusal rows red; revert the mirror's
  cut data and confirm the tie test reds; revert the tail's cuts and confirm the payload test reds.
  A green suite over correct code proves nothing.
- **Run the DW-241 predicate against the nearest legitimate spelling**, not only against the defect
  (D-11.3.7): a cross-variant collision must stay green. That pattern has fired in both of the last two
  stories, and in 11.3 the rules were written *with the warning in hand* and still had four holes, found
  only by executing them.
- Confirm no path you built writes an **asset** sibling (Boundaries), and say how you confirmed it.

## Suggested Review Order

**Start here — the artifact the story exists to introduce**

- The product's first declared family→cuts grouping; hand-written data, never a naming convention.
  [`shipped-face-cuts.ts:77`](../../folio-designer/src/shipped-face-cuts.ts#L77)

- Membership in the mirror is the criterion — never `shippedFamilies`, which is wrong in both directions.
  [`shipped-face-cuts.ts:102`](../../folio-designer/src/shipped-face-cuts.ts#L102)

- Returns the format's own entry shape, so a cut-less family degrades to a bare face name.
  [`shipped-face-cuts.ts:119`](../../folio-designer/src/shipped-face-cuts.ts#L119)

**The pick, and the fork that replaced an embed with a declaration**

- The fork: a shipped family is named and declared; everything else still embeds.
  [`App.tsx:3318`](../../folio-designer/src/App.tsx#L3318)

- Declares the chain and no asset — ~348 KB of duplicate Roboto stops shipping.
  [`App.tsx:1298`](../../folio-designer/src/App.tsx#L1298)

- ONE tail computation both paths call; two that merely agree is 11.2's version defect again.
  [`App.tsx:126`](../../folio-designer/src/App.tsx#L126)

- Both arms share one ordering helper, so the two-command undo cannot drift apart.
  [`App.tsx:3258`](../../folio-designer/src/App.tsx#L3258)

**Route C — the wire, where the guarantee became structural rather than remembered**

- Decodes to a map first: `asset` refused on presence, case-sensitive keys, nulls refused.
  [`component_commands.go:3646`](../../folio-go/component_commands.go#L3646)

- One decoder serves both doors, so `entries` and `tail` cannot drift in what they admit.
  [`component_commands.go:3576`](../../folio-go/component_commands.go#L3576)

- The closed cut set spelled once and tied to the format's own unexported authority.
  [`component_commands.go:3506`](../../folio-go/component_commands.go#L3506)

- Invariant preserved verbatim, mechanism clause replaced — `[]string` was never the point.
  [`component_commands.go:3487`](../../folio-go/component_commands.go#L3487)

**DW-241's narrowing, and DW-245's stated limit**

- One predicate per arm against its own discriminant; the limit disclosed beside the check.
  [`parse.go:472`](../../folio-go/internal/template/parse.go#L472)

- The refusal, and the reason its load-bearing clause now fits inside the 256-rune render cap.
  [`folio-format.md:238`](../specs/spec-folio/folio-format.md#L238)

**Guards — the ones that would have caught what review caught**

- Bases plus cuts equal `Shipped()`, asserted both ways; red-proved on a synthetic copy.
  [`shipped-face-cuts.test.ts:83`](../../folio-designer/src/shipped-face-cuts.test.ts#L83)

- A Roboto pick's chain is compared against `starter.folio` read off disk, not a restated literal.
  [`pick_declares_cuts_ext_test.go:168`](../../folio-go/pick_declares_cuts_ext_test.go#L168)

- Every arm of the closed key set, including the nulls that made route C real.
  [`component_commands_test.go:2924`](../../folio-go/component_commands_test.go#L2924)

- 15.2a's obligation made testable: existing shapes keep their verdict at every input.
  [`component_commands_test.go:2814`](../../folio-go/component_commands_test.go#L2814)

- The disclosure row the Matrix Test Audit found uncovered; asserts what the author reads.
  [`font_chain_variants_test.go:297`](../../folio-go/internal/template/font_chain_variants_test.go#L297)

- Build-time throw tying `scriptFallbacks` to the mirror; they could drift and lose cuts silently.
  [`build-wasm.mjs:283`](../../folio-designer/scripts/build-wasm.mjs#L283)

## Delivery Log

### 2026-09-06 — done

Baseline `7cc02d1`. Shipped at `58f3331` — 21 files, +3690/−67 — as one commit, `Make a shipped-family
pick name and declare its cuts instead of embedding a duplicate`.

**What actually shipped.** A pick of a family the release already ships now **names** it and **declares its
cuts**, and embeds nothing: the ~348 KB byte-identical Roboto duplicate stops travelling, and the entry it
writes can bold because the cuts are declared rather than inferred from a name. The family→cuts grouping
landed as the product's first **declared** production table (`shipped-face-cuts.ts`), tied to
`fonts.Shipped()`'s key set **in both directions** — four bases plus seven cuts is eleven — so a face that
leaves the FontSet cannot leave the mirror behind. Route C reshaped `addFontChain`'s `entries` and
`embedFontFamily`'s `tail` from `[]string` into the format's own chain-entry shape with **no arity
movement**, and the object form **refuses `asset`** structurally in the decoder, which is what keeps
`assetKeyReferenced`'s blind spot out of this story's reach. DW-241's self-reference narrowing landed as a
located load error derived from the closed key set, with DW-245's limit written beside the check and into
`folio-format.md`. The fallback tail is now computed by **one** function both paths call.

**What it turned out not to do, and the owner knew before it was built.** Under D-11.4.1's (2′) criterion
**exactly one family becomes boldable by a pick: Roboto** — the three shipped Notos are uncatalogued and
the control does not offer them, and every catalogue and fetchable face is Regular-only by construction.
D-11.4.2 records that the orchestrator had recommended this story to the owner on a description
(*"any existing document whose author re-picks their family, which is most of them in practice"*) that
measurement falsified, took the correction back to the owner rather than burying it in a Delivery Log, and
the owner ruled **ship as ruled, register the rest**. AC1 is non-vacuous with a population of one.

**Decisions applied.** D-11.0.1 (owner: the story exists at all), D-11.4.1 (2′ — name and declare; the
membership criterion is the declared mirror, and `shippedFamilies` is wrong in *both* directions),
D-11.4.2 (owner: ship as ruled), D-11.4.3 (the three registrations and what stays open in each), D-11.4.4
(the search instrument is evidence), D-11.2.1/.2 (declared, never constructed or parsed), D-11.2.8
(vacuity), D-11.2.11 (the narrowing must not become "no two variants may agree" — asserted in both
directions), D-11.3.7 (run the predicate against the nearest legitimate spelling), D-000.28 (no `-run`
filter on the matrix leg), D-000.30 (how the e2e gate is stated).

**Triage.** **No patched/deferred/rejected tally exists in any artifact readable at close** — the story
file carries no triage section and the builder's report is not on disk. Recording the absence rather than
a number. What *is* on the record: **one review round**, whose finding was that a shipped-family pick
emitted a **one-entry chain** and so produced *less script coverage than the path it replaced*; and the
**Matrix Test Audit**, which found one matrix row uncovered. Both are written up in the Spec Change Log.

**The `bad_spec` vs `patch` discriminator, kept because it generalises.** The one-entry-chain finding was
triaged `patch`, not a `bad_spec` loopback, on this rule (orchestrator ruling, review round 1): *`bad_spec`
exists for a spec defect whose consequences are **pervasive** — where the implementer derived a chain of
decisions from a wrong frame and patching one symptom leaves the frame wrong. The test is whether the
wrongness is local or pervasive, not whether the spec or the code is at fault.* Here it was additive and
local: nothing written was wrong, nothing else in ~1,545 lines derived from its absence, and the fix had no
design freedom. The record should not read as an implementer error, because it was not one — the tasks
under-specified the declare path and the implementer built exactly what they said.

**The uncovered row was the DISCLOSURE row, and that is a general lesson.** Every other matrix row asserts
something *happens*; the DW-245 row asserts something *does not happen and is disclosed instead*. **There
was nothing to red, so nothing was written** — and it arrived inside the guard for DW-245 itself. Generalise
it: *a row whose content is an absence plus a disclosure has no natural failing test, so test-first
discipline silently skips it. Those rows need a positive control and a mutation, or they ship as prose.*

**We shipped a vacuous assertion, and it is fixed. Say it plainly.** A cut-key check written as
`Object.keys(payload).filter(k => VARIANTS.includes(k))` was applied to an `embedFontFamily` payload that
**never carried cut keys at all**, so its two sides could not disagree and it asserted nothing. This run has
made D-11.2.8 — *an assertion whose two sides could be equal is not an assertion* — a standing rule we hold
every other agent to, and we exempted ourselves from it in the same story. A rule we exempt ourselves from
is not a rule. The replacement asserts against the declared mirror and states why the payload cannot carry
a cut by the back door.

**DW-256, and the near-miss that is the better half of it.** Found by the orchestrator while verifying
these gates, not by the product: two literal NUL bytes at `App.tsx:3201-3202` make the designer's core
source file *binary* to the project's ugrep-based `grep -I`, which skips it with **exit 1, no output, no
warning**. Every review layer, Code Map and investigation in this program has searched that file with a
tool that silently declined to open it, so a no-match there has meant nothing. Ruled **D-11.4.4**: a search
tool is part of the evidence, not a neutral window onto it. **The near-miss is instructive** — the broken
grep contradicted a *truthful* builder report, and trusting the instrument over the claim would have filed
a false defect against correct work and sent a builder chasing it. Verifying a report does not mean
assuming the report is the thing that is wrong.

**Measured gates, re-run at `58f3331` for this close.** `folio-go` plain: **2237 pass / 2 fail / 5 skip**,
counted from `go test -json` `Action` events carrying a `Test` field. `folio-go` **unfiltered**
`-tags=matrix` with `FOLIO_FONTGEN_PYTHON` set: **2249 pass / 2 fail / 5 skip**, and
`fontgen: derived and compared 7 of 7 faces`. In **both** runs the only failures are, by enumerated name,
`TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child — the mandated permanent red; **no
third failure**. `TestCrossTargetByteIdentity`, `TestGoldenDigestAgreesAtEveryDeclaredSite` and
`TestShippedRobotoMatchesDesignerCatalogue` all **pass**, so no golden moved. `lint`: **227 pass / 0 fail**,
four packages `ok`, exit 0. `gofmt -l folio-go lint`: **empty (0 bytes)**, exit 0. `npx tsc -b --force`:
exit 0. `npm test`: **65 files / 976 tests**, all passing. `npx oxlint`: exit 0 with **exactly 4**
`only-export-components` warnings, all pre-existing (two in `pdf-viewer.tsx`, two in `App.tsx`).
`npm run test:e2e:compile`: exit 0 — and a green there is a **TYPECHECK, not coverage**: the suite is
exercised at **epic boundaries**, not in CI and not per story (D-000.30). **Not run, deliberately:**
`npm run build`, which is Ask First in this spec's Verification.

**Deferred.** Nine entries by the builder — **DW-247** (the multi-cut fetch is half-built:
`parseFamilyMetadata` already parses every upstream cut and `fetchWebFamily` throws them away; registered
as **load-bearing**, not an enhancement, per D-11.4.2), **DW-248** (the duplicate embed, closed for shipped
families and left **open in one direction** for a fetched face byte-identical to a shipped one, because
detecting that needs the binary comparison D-11.2.1 refuses — the limit is stated so the fix is never read
as complete), **DW-249** (the Go factory list names six designer factories and there are nine, with a
measured false negative), **DW-250** (a `.folio` naming shipped faces is not self-contained and the format
does not say which FontSet it was written against), **DW-251** (`assetKeyReferenced` never walks a chain
entry's variant siblings), **DW-252** (every chain-entry refusal reason is over the 256-rune render cap, so
authors have read truncated sentences since 11.2), **DW-253** (the mirror's tie test parses a sibling test
file by regex, so the `fonts.go` parse now exists twice), **DW-254** (a cut can be declared only at chain
*creation*; no command adds or changes one afterwards), **DW-255** (a command's refusal names the chain but
not which entry, while the loader names the sibling exactly). Plus **DW-256** by the orchestrator. **All ten
carry `### DW-` headings and are census-visible**; none is in the workflow's raw `- source_spec:` block
form, which is invisible to a `### DW-` census (D-11.0.2). Owners: DW-247 is Epic 16-shaped; the rest are
unassigned and want placement at Epic 11's boundary gate.

**Register housekeeping discharged at this close.** D-000.31a's second survivor — *"ten blocks under two
known headings is a close-sized task… folded into the next close rather than scheduled"* — is done here.
The ten unindexed raw blocks lodged inside **DW-189** (seven, from Story 17.4) and **DW-190** (three, from
Story 17.5) are now **DW-257 … DW-266**, each with a `### DW-` heading, its original `summary` and
`evidence` text preserved verbatim, and a pointer left in the parent entry. Owners and severities on those
ten are marked **the closer's read, not a lead ruling**; DW-266 additionally collides with D-000.30 on its
"run e2e in CI" half and is flagged unassigned for that reason. After this the register holds **269**
`### DW-` headings and **six** remaining `- source_spec:`+`summary:` blocks, all of them within two lines
of their own heading — genuine provenance metadata, not orphans (D-000.31a).
