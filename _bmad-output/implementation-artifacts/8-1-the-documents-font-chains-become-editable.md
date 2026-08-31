---
title: "Story 8.1: The document's font chains become editable"
type: 'feature'
created: '2026-08-31'
status: 'done'
baseline_revision: 'b119831059cce3ddfa362f4122d1e48bb18a6e79'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: ['oversized']
deferred:
  - summary: >-
      A pre-existing .folio declaring a chain with more than 64 entries, or a face name over 512
      bytes, now fails to open at all rather than merely being uneditable.
    evidence: |-
      canvasFontChains' two new refusals (page_setup.go:463 and :467) run inside
      CanvasWithTextPaint, which Engine.load calls at wasm/engine.go:119, so the error surfaces
      from Load. decodeFonts (internal/template/parse.go:313) bounds neither entry count nor face
      length, and render.go's fontChain never counted entries, so such a document parsed and
      rendered before this story. The bound itself was directed by the spec's Task 6 and mirrors
      the pre-existing canvasFontFamilies shape, so it is not a defect; the compatibility
      narrowing is simply unrecorded and has no matrix row.
    location: >-
      folio-go/page_setup.go:463
    severity: medium
  - summary: >-
      Go sorts projected chain names by byte order while the browser guard checks strict ascending
      order in UTF-16 code units; the two disagree on names mixing astral-plane and U+E000-U+FFFF
      characters, which drops the whole snapshot and blanks the canvas.
    evidence: |-
      engine-protocol.ts's sorted/unique check predates this story and already applied to
      fontFamilies, so the mismatch is pre-existing rather than caused here. This story does make
      it newly reachable through the Go command API, but the designer sends no chain command until
      Story 8.2, so it is not reachable through the product yet.
    location: >-
      folio-designer/src/engine-protocol.ts
    severity: medium
  - summary: >-
      The wasm host boundary has no test that dispatches a font-chain command, so the refusal
      messages are specified at the host wire and measured one layer below it.
    evidence: |-
      folio-go/wasm/cmd/engine is //go:build js && wasm, so go test ./... never compiles it and no
      CI job executes its host-boundary assertions. Its own sibling test states the hazard
      verbatim: "Every Go-side assertion would still have been green." The package's dormancy is a
      standing deferred item, not introduced here; P11 added a source-reading tripwire tying the
      two bound literals, which is the most a compiled test can currently reach.
    location: >-
      folio-go/wasm/cmd/engine/main.go:236
    severity: medium
  - summary: >-
      FR52's chain-level (key) reorder is delivered by no story in Epic 8 and is not expressible in
      the .folio format.
    evidence: |-
      Design Notes R1 rules this out against four independent authorities (AD-9's sorted-keys rule,
      folio-format.md:390's "no authored key order" as D-4.1.1's discharged debt, format-changes.md's
      Unchanged Order row, and D-R7.9 placing the epic's format change at 8.3). This story delivers
      entry-level reordering, which the format does store. Recorded here because the epic text still
      promises the chain-level form.
    location: >-
      _bmad-output/planning-artifacts/epics.md:113
    severity: low
  - summary: >-
      Index refusals reach the author without the "folio:" prefix every sibling refusal carries.
    evidence: |-
      fontChainIndex passes commandInt's error through verbatim, and commandInt
      (component_commands.go:509) - unlike commandString - does not prefix its messages. An author
      sees "index must be an integer" beside "folio: name must be a non-empty string". The
      inconsistency lives in commandInt and predates this story.
    location: >-
      folio-go/component_commands.go:509
    severity: low
---

<intent-contract>

## Intent

**Problem:** Nothing anywhere can write the `.folio` document's `fonts` map. Measured at `b2fdaa1`: `Document.Fonts` (`internal/template/model.go:147`, `map[string][]string`) has exactly two write sites and both are the loader (`parse.go:126,128`). There is no command, no public API and no designer affordance. So every template created in the designer offers a font list of length one — whatever the starter file declared — and `style.fontFamily` names a chain the author never chose. Story 8.2 puts an editor on the typography panel, 8.3–8.6 embed and catalogue faces; all of them need an engine that can be *told* to change a chain, and none of that exists.

**Approach:** Add a font-chain command vocabulary to the engine's existing opaque, versioned command surface, and extend the canvas projection to carry the chains themselves so the designer can re-project rather than model them. The refusals are the substance: deleting a chain something still names, renaming without carrying the elements, and emptying a chain are each refused inside the one transaction that already makes every command atomic. No `.folio` format change, no new fixture, no panel.

## Boundaries & Constraints

**Always:**
- **One command, one history entry, engine-owned.** Every font-chain edit is a new `kind` in the versioned vocabulary dispatched by `ApplyComponentCommand` (`component_commands.go:57`), reached through `wasm.Engine.Apply` (`wasm/engine.go:210`), which clones → applies → serializes → reparses → projects → installs and calls `pushUndo` exactly once (`:259`). A multi-part edit is one entry **by construction**; do not add a second push and do not build a second transaction.
- **A rename carries the elements in the same handler.** `fontFamily` has exactly **two** attachment points in the model — `Element.Style` (`model.go:199`) and `TableExt.HeaderStyle` (`model.go:249`), both `Presence[Style]` with `Style.FontFamily` at `model.go:282`. Both are live references: `render.go:1098-1105` resolves the first, `table_render.go:315-318` → `:653-663` resolves the second. A rename rewrites **both**, across all three bands (`Bands.PageHeader/Content/PageFooter`, `model.go:158-161`), inside the one handler.
- **One authority for "is this a chain `fontFamily` may name".** Today that rule is open-coded **five** times: `component_commands.go:1221` (`knownFontFamily`), `:1511` (`defaultFontFamily`), `page_setup.go:433`, `render.go:1103`, `table_render.go:656`. Create the single authority as a method on `template.Fonts` returning `(chain []string, ok bool)` — present **and** non-empty — and route all five through it. Do not add a sixth copy.
- **Author-supplied strings go through the existing `commandString`** (`component_commands.go:1259-1268`), which already refuses absent and empty values with `"folio: %s must be a non-empty string"`. That is the empty-name refusal, free, through the shared helper.
- **Projection bounds are referenced, never restated.** `maxCanvasPropertyString` (512, `page_setup.go:39`) and `maxCanvasFontFamilies` (256, `page_setup.go:424`) are the constants; a command refusal cites those symbols.
- **A refusal is `componentFailure`** (`component_commands.go:29-31`), whose `Message` reaches the author verbatim: the wasm host matches `*ComponentCommandError` **before** `*RenderError` (`wasm/cmd/engine/main.go:236-244`) and emits `bounded(msg, 512)` with no `reportableMessage` filtering.
- **The corpus hashes identically.** All **22** `goldenDigestRecord` digests unmoved; `shasum -a 256 fixtures/*/expected.pdf` is 22 lines, zero deltas.
- Every commit ends `Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>`.

**Block If:**
- **Any of the 22 golden digests moves.** State it and STOP. A moved digest invalidates a human attestation by construction — `fixtures/statement-signoff.json` (in whole, all four) and `fixtures/thai-stacked-marks/signoff.json`. **No agent writes `reader`, `date` or `examined`** (D-R7.6, D-R7.7).
- **The plan cannot be expressed without building Story 8.2's chain-editor panel.** Say so and halt rather than annexing it.
- **A `.folio` format change turns out to be required** — an authored order for chain *keys*, a new key, or a version bump. That is Story 8.3's ground and D-R7.9's decision; halt rather than take it here. (Design Notes **R1** settles the one question that could raise this — chain-key ordering — and is CLOSED; re-open only on new evidence, and keep the R1 heading alive.)
- **Minting a diagnostic registry code appears necessary.** Design Notes **R3** settles this against three measurements. If that reasoning is falsified, halt — a shipped code is irreversible under AD-14.

**Never:**
- Never touch `README.md` at the repo root. Never `git add -A` or `git add .`. Never push. Never create a branch.
- Never change `.folio`'s canonical byte form. `writeFonts` (`serialize.go:174-184`) sorts chain keys and `writeObject` (`:42`) sorts them again; entry order is the slice. Leave both.
- Never add an authored key order to the `fonts` map. `folio-format.md:390` states verbatim that "`fonts` is a mapping with no authored key order" — a discharged debt from D-4.1.1. Do not regress it.
- Never build Story 8.2's chain editor, 8.3's asset entries (`{"asset": "<key>"}`), 8.4's embedded-face rendering, or 8.5/8.6's catalogue. TypeScript changes are limited to the **protocol contract** (`engine-protocol.ts` and its tests, plus test fixtures the new required field breaks). `App.tsx:1146-1186` (`FontFamilyProperty`) is not touched.
- Never make the empty-chain rule a **load-time** refusal. `decodeFonts` (`parse.go:313-328`) accepts `"body": []` today and a `.folio` in the wild may carry one; narrowing the loader is D-7.8.3 / DW-54 territory and is not this story.
- Never add a second `Apply*` entry point. `wasm/engine.go:228` special-cases only `"pageSetup"`; everything else already routes to `ApplyComponentCommand`.
- Never write a new golden fixture or a new `//go:build matrix` file. Neither is owed here.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Add a chain | `{"kind":"addFontChain","version":1,"name":"heading","entries":["Noto Sans"]}` on a doc declaring `body` | `fonts` gains `heading`; projection lists `body`,`heading` sorted; one revision, one undo step | No error expected |
| Add with a taken name | `name` already a key in `fonts` | Refused; document byte-identical, revision and both history branches unchanged | `componentFailure("", "fonts.heading", "a font chain named \"heading\" already exists")` |
| Add with no entries | `"entries":[]` | Refused | `…"a font chain must declare at least one entry"` |
| Add with an empty name | `"name":""` | Refused by `commandString` before any mutation | `"folio: name must be a non-empty string"` |
| Rename, elements carried | `{"kind":"renameFontChain","version":1,"name":"body","to":"brand"}`; 3 elements + 1 `headerStyle` name `body` | Key renamed, all 4 references rewritten, **one** history entry; one undo restores map and elements together | No error expected |
| Rename onto a taken name | `to` already a key | Refused; the destination chain is never silently destroyed | `…"a font chain named \"x\" already exists"` |
| Delete an unreferenced chain | `{"kind":"deleteFontChain","version":1,"name":"unused"}`, nothing names it | Key removed; projection loses it | No error expected |
| **Delete an orphaning chain** | `name` is still named by `style.fontFamily` on `e2`,`e7` and by `headerStyle.fontFamily` on `t1` | **Refused**, document unmutated | `componentFailure("", "fonts.body", "font chain \"body\" is still named by e2, e7, t1")` — element ids in document order (pageHeader, content, pageFooter), the `headerStyle` bearer included |
| Remove the last entry | `{"kind":"removeFontChainEntry","version":1,"name":"body","index":0}` on a one-entry chain | **Refused** | `…"removing that entry would leave font chain \"body\" with no entries"` |
| Remove a non-last entry | 3-entry chain, `index:1` | Entry removed, order of the rest preserved | No error expected |
| Move an entry | `{"kind":"moveFontChainEntry","version":1,"name":"body","from":0,"to":2}` | Slice reordered; `.folio` entry order follows verbatim | Out-of-range index → `…"entry index is out of range"` |
| Add an entry | `{"kind":"addFontChainEntry","version":1,"name":"body","index":1,"face":"Noto Sans Thai"}` | Entry inserted at `index`; a face absent from the render `FontSet` is **accepted** (the format's standing tolerance, `render.go:1144-1149`) | Index out of range → refused |
| Any command naming an absent chain | `name` not a key in `fonts` | Refused | `…"no font chain named \"x\" is declared"` |
| Name over the projection bound | `name` longer than `maxCanvasPropertyString` | Refused **at the command**, located, before the projection's unlocated bare error fires | `…"font chain name exceeds the projection bound"` |
| 257th chain | doc already declares `maxCanvasFontFamilies` chains | Refused at the command | `…"document declares more font chains than the projection bound"` |
| **Edit and edit back** | rename `body`→`brand`, then `brand`→`body` | Final canonical bytes **byte-identical** to the original, `fonts` and `bands` both; two revisions, two undo steps | No error expected |
| No-op | a command whose canonical bytes equal the current bytes | `wasm/engine.go:244` short-circuit: revision, dirty state and both history branches unchanged | No error expected |

</intent-contract>

## Code Map

**The model and the format (read-only unless noted)**
- `folio-go/internal/template/model.go:147` — `type Fonts map[string][]string`; `:32` `Document.Fonts`. `:199` `Element.Style`, `:249` `TableExt.HeaderStyle`, `:282` `Style.FontFamily` — **the complete reference set**. `:158-161` `Bands{Content,PageFooter,PageHeader}`.
- `folio-go/internal/template/parse.go:313-328` — `decodeFonts`. **Validates nothing about the key** and accepts an empty array. Read-only here.
- `folio-go/internal/template/serialize.go:174-184` — `writeFonts`, `slices.Sorted(maps.Keys(f))`; `:42` `writeObject` sorts again; `:60-77` `writeStringArray` preserves slice order. **Read-only** — this is why edit-and-edit-back is byte-safe (Design Notes R2).
- `folio-go/internal/template/parse_bands.go:677-684` — `style.fontFamily` decode: must be a string, nothing more.
- `_bmad-output/specs/spec-folio/folio-format.md:124-136` (`fonts`), `:390` (no font default, no authored key order).

**The command surface (edited)**
- `folio-go/component_commands.go:36` `ApplyComponentCommand`; `:57-94` the `switch kind` — **the six new arms go here**. `:1251-1256` `componentFields` (exact arity, counts `kind` and `version`); `:1259-1268` `commandString`; `:20-31` `ComponentCommandError` + `componentFailure`.
- `folio-go/component_commands.go:654-757` `setComponentAsset` — **the structural precedent** (D-5.13.1): one command that writes a document-level map, repoints an element and drops an orphan, in one transaction. `:759-786` `assetKeyReferenced` — the reference-walk shape to copy, including its warning at `:770-776` that there is no shared element enumerator.
- `folio-go/component_commands.go:1035-1041` (`case "fontFamily":`), `:1220-1223` `knownFontFamily`, `:1504-1516` `defaultFontFamily`, `:1467` its adoption in `createComponentInBand`.
- `folio-go/wasm/engine.go:210-263` `Apply` — the transaction, the `bytes.Equal` no-op short-circuit at `:244`, the single `pushUndo` at `:259`. `:221-232` the kind dispatch. **Read-only.**
- `folio-go/wasm/cmd/engine/main.go:229-254` `engineFailure` — `*ComponentCommandError` matched **before** `*RenderError`; `:272-281` `reportableMessage`, which the command channel never reaches. **Read-only, and the evidence that no code needs minting.**

**The projection (edited)**
- `folio-go/page_setup.go:266` `CanvasProjection`; `:280-287` `FontFamilies` and its doc comment — *"It is names only — never the chains…"* — **the sentence AC5 outgrows**. `:421-424` `maxCanvasFontFamilies`, `:39` `maxCanvasPropertyString`, `:427-445` `canvasFontFamilies` (already refuses an over-long name and an over-count, with a stated reason).
- `folio-designer/src/engine-protocol.ts:126` `CanvasProjection`; `:153-156` `fontFamilies`; `:194` `hasOnly` (**a subset check — an extra key Go sends is REJECTED**); `:211` the key list; `:217` the sorted/unique/length rule; `:14-15` `MAX_ENGINE_FONT_FAMILIES`.
- `folio-designer/src/engine-bounds-mirror.test.ts:69-74` — the Go/TS constant pair list. `MAX_ENGINE_FONT_FAMILIES` is **absent** from it today; the new entry-count bound must be added.

**Test surface**
- `folio-go/component_commands_test.go:17` `componentTemplate` (drives `ApplyComponentCommand` directly); `folio-go/wasm/engine_test.go` (drives `Engine.Apply` — revision/history/no-op precedents at `:284`, `:311`, `:476`, `:507`).
- `folio-go/text_alignment_test.go:193` and `folio-go/table_render_test.go:54,89,654` — the only multi-chain documents in the repo; **none is byte-compared**.
- `folio-go/byte_neutrality_test.go:92-524` `goldenDigestRecord` (**22 entries**), `:566` `TestGoldenDigestAgreesAtEveryDeclaredSite`, `:853` `declaredEpic2GateObligations` (**do not add to it — no new matrix file is owed**).
- `folio-designer/src/engine-protocol.test.ts`, `src/App.test.tsx:19` (the shared canvas fixture — a new required projection key breaks it).

**Read-only evidence recorded at `b2fdaa1`**
- No test anywhere pins the string `"names a chain with no entries"` (grep: zero hits in `*_test.go`).
- No fixture declares more than one chain, and every `style.fontFamily` in `fixtures/` is the literal `"body"` (155 occurrences, 22 files). `component-asset-import` and `image-embed` declare `"fonts": {}`.
- `versionRequiredByContent` does not consult `fonts`, so no chain edit can move `version`.

## Tasks & Acceptance

**Execution:**

1. `folio-go/internal/template/model.go` — add `func (f Fonts) Chain(name string) ([]string, bool)`, returning the chain only when the key is present **and** the chain is non-empty. Doc-comment it as **the** authority for "a chain `fontFamily` may name", and state the measured population it replaces (five open-coded copies, enumerated).
2. `folio-go/component_commands.go`, `folio-go/page_setup.go`, `folio-go/render.go`, `folio-go/table_render.go` — route all five copies through `Fonts.Chain`. `render.go:1103` and `table_render.go:656` keep their own message text; only the predicate is shared. Correct `page_setup.go:427`'s comment, which claims it projects "exactly the names knownFontFamily accepts" while re-implementing the test three lines later.
3. `folio-go/component_commands.go` — add six `switch` arms and their handlers: `addFontChain`, `renameFontChain`, `deleteFontChain`, `addFontChainEntry`, `moveFontChainEntry`, `removeFontChainEntry`. Every author string through `commandString`; every arity through `componentFields`; every refusal through `componentFailure` with `ElementID: ""` and `DataPath: "fonts." + name`.
4. `folio-go/component_commands.go` — add `fontChainReferences(t, name) []string`: the element ids naming `name` through **either** `Element.Style.FontFamily` **or** `Element.Table.Value.HeaderStyle.FontFamily`, across all three bands, in document order. Model it on `assetKeyReferenced` (`:759`) and carry a doc comment naming the two attachment points it measured, per D-8.0.1's rule about comments that assert a negative. It backs the delete refusal.
5. `folio-go/component_commands.go` — `renameFontChain` rewrites both attachment points in the same handler before returning, so `Apply`'s single `pushUndo` covers map and elements.
6. `folio-go/page_setup.go` — add `CanvasFontChain{Name string; Entries []string}` and `CanvasProjection.FontChains []CanvasFontChain`, built from the same sorted, non-empty key walk that produces `FontFamilies`, so `FontChains[i].Name == FontFamilies[i]` holds by construction. Add a per-chain entry-count bound constant and refuse over it with a stated reason, the way `canvasFontFamilies` already does.
7. `folio-designer/src/engine-protocol.ts` — add `fontChains` to the `CanvasProjection` type, to `isCanvas`'s `hasOnly` key list at `:211`, and to its validation: bounded count, bounded name and entry lengths, entry array non-empty, and `fontChains.map(c => c.name)` deep-equal to `fontFamilies`. Add the entry-count constant and its mirror pair.
8. `folio-designer/src/engine-bounds-mirror.test.ts` — add the new entry-count pair, **and** the long-absent `MAX_ENGINE_FONT_FAMILIES` / `maxCanvasFontFamilies` pair, so the new field is not a one-sided edit like its neighbour.
9. `folio-designer/src/App.test.tsx`, `src/engine-protocol.test.ts` — extend the canvas fixture with `fontChains`; add validator cases for a missing key, an extra key, a name/`fontFamilies` disagreement, an empty entry list, and an out-of-order chain list.
10. `folio-go/component_commands_test.go` — cover every I/O matrix row against `ApplyComponentCommand`, including the two-attachment-point delete refusal and the document-order id list.
11. `folio-go/wasm/engine_test.go` — assert one revision and one undo step per command; assert **one** undo restores map **and** elements after a rename; assert a refused command leaves bytes, revision and both history branches unchanged.
12. `folio-go/wasm/engine_test.go` — the edit-and-edit-back byte test: a **multi-chain** document (this is the repo's first byte-level pin on multi-chain emission), rename out and back, assert the final canonical bytes equal the original exactly.
13. `folio-go/component_commands_test.go` — assert `defaultFontFamily`'s coupling explicitly: after renaming the sorted-first chain, a newly created text element adopts the new sorted-first name. D-4.1.1 rejected an alphabetical default on exactly this hazard; record the behaviour rather than leave it unmeasured.

**Acceptance Criteria:**
- Given the six new commands, when each is applied through `wasm.Engine.Apply`, then exactly one revision advances and exactly one undo entry is pushed per accepted command, and a refused command advances none.
- Given a rename of a chain named by three elements and one `headerStyle`, when it is committed, then all four references and the map key change together and a single undo restores every one of them.
- Given the delete refusal, when the chain is named by an element in each of the three bands and by a table's `headerStyle`, then the refusal message names all four ids in document order and the document is byte-identical afterwards.
- Given each new refusal, when its guard is **deleted** and the suite re-run, then a specific named test goes red — proved per guard, not inferred. The `headerStyle` arm of `fontChainReferences` is proved separately from the `style` arm, because it is the subject a `style`-only test never reaches.
- Given `FontChains` is added to `CanvasProjection` and **not** to `isCanvas`'s key list, when `parseInbound` runs, then it returns `undefined` — the `hasOnly` subset check is measured, in both directions, not assumed.
- Given the whole change, when `shasum -a 256 fixtures/*/expected.pdf` is run, then it prints 22 lines identical to the baseline's.

## Spec Change Log

## Review Triage Log

### 2026-08-31 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 12: (high 1, medium 5, low 6)
- defer: 5: (high 0, medium 3, low 2)
- reject: 8: (high 0, medium 0, low 8)
- addressed_findings:
  - `[high]` `[patch]` The pageFooter arm of `fontChainReferences` was entirely unmeasured. Measured, not inferred: dropping `t.doc.Bands.PageFooter.Elements` from `fontChainBands` left the whole folio-go suite green but for the mandated permanent red, while the AC demands the delete refusal name an element in each of the three bands plus the `headerStyle` bearer. The fixture gained a pageFooter element naming the chain under test; the refusal now names `e2, e7, e9, e11` in document order, and the same mutation now reddens `TestFontChainRenameCarriesEveryElementNamingIt` and `TestFontChainDeleteRefusesAChainAnElementStyleNames`.
  - `[medium]` `[patch]` `fontChainOrphanMessage` had zero tests: its trimming branch, its `" and N more"` suffix, its whole-id-boundary guarantee and its `"%d elements"` fallback were all unexecuted, both delete tests being ~55 bytes against a ~477-byte budget. Added `TestFontChainOrphanMessageTrimsOnAWholeIdBoundary` and `TestFontChainOrphanMessageFitsEveryNameTheCommandAccepts`.
  - `[medium]` `[patch]` `fontChainOrphanMessage`'s budget could go negative and defeat its own doc comment: `fontChainName` admits a name up to 512 bytes while the prefix is roughly `len(name)+28`, so every branch overran the host's 512-byte cut and the message was cut mid-name. The name is now trimmed first, on a rune boundary, measured against the formatted prefix so `%q` escaping cannot defeat it.
  - `[medium]` `[patch]` The projection's two new refusals (entry count, face length) were never exercised on a loaded document — both `if` blocks were deletable with every Go test still green, and without them an over-long chain reaches the wire, `isCanvas` returns false and the canvas blanks with nothing to attribute it to. Added `TestCanvasFontChainEntryCountIsBoundedOnALoadedDocument` on a document parsed from bytes, and a `fonts chain entry` row to the identifier-bounds table.
  - `[medium]` `[patch]` Several reachable refusal branches had no test, against the AC that each new refusal be proved by deleting its guard: `entries` absent, `entries` not a string array, an empty-string entry, an entry over the bound, and a rename destination over the name bound. Added `TestFontChainAddRefusesEveryMalformedEntryList` and `TestFontChainRenameRefusesADestinationOverTheProjectionBound`.
  - `[medium]` `[patch]` Nothing tied `CanvasFontChain`'s JSON tags to the browser's `hasExactKeys(chain, ['name', 'entries'])`. `fontChains` is the first nested object the projection has gained, so a new Go field would make `isCanvas` false and blank the canvas — verbatim the failure `canvas_projection_wire_test.go` exists to prevent. The recorded-key tripwire now covers the nested chain object on both sides.
  - `[low]` `[patch]` Dangling doc references to `canvasFontFamilies`, a function this story renamed out of existence, at `internal/template/model.go:160` and `page_setup.go:445` (plus two further stale mentions the patch pass found). Corrected.
  - `[low]` `[patch]` The matrix row for "Move an entry" expects `.folio` entry order to follow verbatim, but the test asserted the in-memory slice only; the sole byte-level assertions were the rename round trip and a no-op move. Added `TestEngineFontChainMoveIsFollowedByTheFolioBytes`.
  - `[low]` `[patch]` A TypeScript assertion conflated two conditions — `fontChains: [], fontFamilies: [], extraProjectionKey: 1` is rejected by the extra key alone — and the positive zero-chain case was missing, though `component-asset-import` and `image-embed` both declare `"fonts": {}`. Split, and the positive case added.
  - `[low]` `[patch]` The host bounds `DataPath` at 256 while `fontChainPath` built `"fonts." + name` for names up to 512, so the over-long-name refusal — the one case where locating the name is the whole point — arrived truncated, and `bounded` slices by bytes so it could split a rune. `fontChainPath` now cuts at `maxComponentDataPathBytes` on a rune boundary.
  - `[low]` `[patch]` `maxComponentFailureMessageBytes = 512` hand-copied the host's literal with nothing tying the two — the same one-sided-constant defect this story fixed for `maxCanvasFontFamilies`. `TestComponentFailureBoundsMatchTheHostsOwnLiterals` now reads `wasm/cmd/engine/main.go`, which no compiled test can see.
  - `[low]` `[patch]` `canvasFontChains`' comment still asserted the list was "exactly the names knownFontFamily accepts" — the drift the spec's Task 2 called out. It now states what the code does: it asks `Fonts.Chain`, declared and non-empty.

**Rejected, with the authority each was tested against.** (1) No replace/set-entry command — a vocabulary expansion the intent does not ask for and Story 8.2's ground. (2) No TypeScript command builders — the intent's "Never" limits TS changes to the protocol contract. (3) Duplicate entries in a chain are accepted — consistent with the format's standing FontSet tolerance (`render.go:1163-1177`); the intent requires no dedup. (4) `addFontChain`'s nil-map guard is unreachable because `parse.go:128` always sets `Fonts{}` — harmless defensive code. (5) `addFontChain` counts raw keys rather than projectable chains — "declares" is a defensible reading of the matrix row, and refusing marginally early is the safe direction. (6) Two engine assertions are coupled to serializer text shape — brittle, but deliberately byte-level and passing. (7) `maxCanvasFontChainEntries` is declared in `component_commands.go` rather than beside `maxCanvasFontFamilies` — file placement only; the mirror pair locates it. (8) `engine-bounds-mirror.test.ts` was called out of TS scope — the spec's Task 8 directs that change explicitly.

## Design Notes

**R1 — "Reorder" is entry-level, and the epic text overpromises. RULED; do not re-open.** FR52 (`epics.md:113`) says "Create, rename, reorder and delete the document's font chains **and their entries**", and AC1 (`:2716`) says "a chain added, renamed, reordered". Chain-*key* reordering is not expressible: `Fonts` is a Go map with no stored order, `writeFonts` sorts, and four independent authorities forbid adding one here — AD-9's Rule says `.folio` has one legal serialization with "object keys sorted"; `folio-format.md:390` states there is no authored key order, as D-4.1.1's discharged debt; `format-changes.md`'s `fonts` table lists the epic's format changes and its Order row reads "Unchanged"; and D-R7.9 places the epic's format change at 8.3. So "reorder" here is reordering a chain's **entries**, which the `[]string` already carries. **FR52's chain-level reorder is delivered by no story in Epic 8 and is not expressible in the format** — reported, not resolved.

**R2 — Edit-and-edit-back is byte-safe by construction, and that is why it needs a test.** A chain's emitted byte position is a total function of its key (sorted at `serialize.go:180`, sorted again at `:42`), and entry order is the slice. Rename `X`→`Y`→`X` restores the identical key set and identical slices, so `writeFonts` emits identically — **provided** the rename also restores every `style.fontFamily` and `headerStyle.fontFamily` it rewrote, which is where the `bands` bytes could move instead. `versionForSave` never consults `fonts`, so `version` cannot drift. No document in the repo currently pins multi-chain emission order at the byte level; task 12 is the first.

**R3 — No diagnostic code is minted, and the reason is measured, not deferred.** D-7.8.1 rules that the general code is the default and a specific one is minted only when a **named consumer must branch on it**. Three measurements at `b2fdaa1`: (a) a command refusal travels as `*ComponentCommandError`, which `main.go:236-244` matches **before** `*RenderError` and emits with `bounded(msg, 512)` — it never reaches `reportableMessage`, so the message already arrives at the author intact; (b) `diagnostic_registry_census_test.go:131` hard-requires `errors.As(err, &renderErr)` for every registered code, so a code minted for a command refusal could not be made green without routing the refusal through `RenderError` — a structural change nothing asks for; (c) no TypeScript consumer branches on any registry code — `componentDiagnosticDetail` (`App.tsx:1231`) reads `elementId`, `dataPath` and `message` and never `code`, and `preview/diagnostic-presenter.tsx:9-10` says it "deliberately has no code registry". The only code-branching consumer in the designer is `pageSetupDiagnostic` on `PAGE_SETUP_INVALID`, which is a host-local string, not a registry code. **Named consumer that would branch: none. Mint nothing.** The `diag.go:253-258` reservation that once guarded this is spent — it was answered by `CodeTemplateFieldInvalid`, and D-7.8.1's guardrail 3 scopes that code to the **load stage** anyway, so it is not reusable here either.

**R4 — Located-ness without an element id.** `ComponentCommandError.ElementID` is single-valued and there is no precedent anywhere for a refusal naming a *list* of element ids. A chain command is not addressed to an element, so `ElementID` stays empty and `DataPath` carries `fonts.<name>` — the same shape page-setup refusals use. The element list lives in `Message`, which is where AC2's "naming the elements" is satisfiable; `formatFontChain` (`render.go:1184`) is the existing precedent for joining names into a message. **Bounding rule:** the host cuts at 512 bytes, so name the ids in document order and, if the list would not fit, name as many as fit whole and append `" and N more"`. Do not let the host truncate mid-id.

**R5 — What the "existing rule" in AC4 actually is, measured.** AC4 says the empty-chain refusal upholds "the existing rule that a chain with no entries is not a chain `fontFamily` may name". That rule exists at **render time only** — `render.go:1097-1106`, mirrored at `table_render.go:655-662` — and is *filtered* rather than enforced at `component_commands.go:1039` and `page_setup.go:433`. `decodeFonts` accepts `"body": []` at load. So this story's refusal is an **additional guard at the command path**, not a relocation of the existing one, and `Fonts.Chain` is how it routes through the same authority instead of copying it. Note also that the render-time message **conflates "chain absent" with "chain empty"** into one sentence; the command path distinguishes them, and that is an improvement, not a divergence.

**R6 — The projection must grow, and the TS change is not optional.** `FontFamilies`'s own doc comment (`page_setup.go:283-285`) says it is "names only — never the chains… so it cannot reconstruct the fonts map". AC5 requires the projection to carry **the chains**, and a move/remove-entry command is otherwise invisible to the designer. `hasOnly` (`engine-protocol.ts:194`) is a subset check, so shipping `fontChains` from Go without adding it at `:211` makes `isCanvas` false → `PROTOCOL_INVALID` → the worker is terminated and the canvas blanks with no attributable error. Both halves land in one commit. Keeping `FontChains` to exactly the chains `FontFamilies` names preserves the single authority and gives a free cross-check.

**R7 — DW-35 disposition (this plan gate is its named owner).** DW-35 — the canvas hard-codes one CSS font stack regardless of the document's `fonts` map — is **re-owned to Story 8.2**, not closed here. Reachability test: 8.1 is engine-side and the designer sends no chain command until 8.2's editor exists, so the divergence is not reachable *through the product* until 8.2; and DW-35's own fix needs the projected per-component chain, which 8.2 needs anyway. A named story with a position in the sequence, per D-8.0.5. **DW-32** (the command encoder splices author text into JSON unquoted) is likewise flagged to 8.2: 8.1 sends nothing from TypeScript, but 8.2 will carry author-typed chain names through that same splice, where a `"` or `\` is worse than the numeric case DW-32 describes.

**Findings against the epic text, measured at `b2fdaa1`.** (i) Story 8.1's `Covers:` line (`epics.md:2711`) names AD-9, AD-15 and AD-22 but **not AD-16**, whose Rule is the substance of AC1; AD-8 (the font invariant) and AD-14 are likewise unnamed and bind. (ii) FR52 promises chain-level reorder, which the format cannot express (R1). (iii) AC2's refusal is written against `style.fontFamily` alone; `headerStyle.fontFamily` is an equally live reference (`table_render.go:315-318`) and the AC's own principle — "never accepted with the orphaned elements left to fail at render" — reaches it. (iv) 8.1's ACs enumerate two refusals plus the empty-chain one; the **duplicate-name refusal** is stated only in **Story 8.2's** AC4 ("an orphaning delete, an empty chain, a duplicate name"), and since 8.2's panel only reports what the engine answers, that refusal belongs to this story. (v) The dispatch's premise that a single existing authority validates a chain name is **false**: nothing validates the name at all, and the nearest rule is open-coded five times — the authority is created here, not found.

## Verification

**Commands:**
- `cd folio-go && go test -count=1 ./...` — expected: exactly ONE distinct red, `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (got 7, need >=20), the mandated permanent red. Anything else is a regression. Report pass/fail counts, not "green".
- `cd folio-go && go vet -tags=matrix ./...` — expected: clean.
- `gofmt -l folio-go` from the repo root — expected: no output. (`lint/…/licencegraph_test.go` carries a known finding, DW-23, and is outside this path.)
- `cd folio-go && FOLIO_MATRIX_TARGET=darwin/arm64 go test -count=1 -tags=matrix -run TestTargetRenderHash -v .` — and once each for `linux/amd64`, `linux/arm64`, `js/wasm`. **Plus an unset control** (`FOLIO_MATRIX_TARGET` not exported) proving the legs are not no-ops: the control must show the assertion skipped while the four exported legs assert. Report all five.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity .` — expected: pass.
- `cd folio-go && go test -count=1 -tags=matrix -run TestThaiStackedMarksSemanticSignOffIsRecorded .` — expected: pass.
- `cd lint && go test -count=1 ./...` — **`-count=1` is MANDATORY.** The `rules` package walks a directory and Go's cache does not track `ReadDir`, so a cached `ok` here is no measurement at all (D-7.9.5). Expected: clean. Note `TestFloatTypedTestScopeInventory` pins five sites **by line number** in `shaping_expectations_test.go` and `internal/fontset/*_test.go` — do not shift lines in those three files.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` — expected: typecheck and e2e-compile clean; 4 pre-existing `only-export-components` lint warnings; the suite reports **284 tests / 34 files** at baseline — report the number after, do not gate on it.
- `shasum -a 256 fixtures/*/expected.pdf` — expected: **22 lines, byte-identical to the baseline's**. Capture the baseline list before any edit and diff the two.
- Environmental, not yours to fix: `TestShippedFacesReproduceFromUpstream` fails or skips under `-tags=matrix` without `fontTools`. Say so explicitly rather than folding it into a count.

**Mutation proofs (run, never inferred):**
- Per new refusal, **delete the guard** and name the test that goes red. Deletion is the cheapest screen and catches the class value-mutation cannot: a subject the tests never *reach*.
- Prove the `headerStyle.fontFamily` arm of `fontChainReferences` **separately** from the `style.fontFamily` arm — the `headerStyle` population is exactly the one a `style`-only test never reaches.
- Prove `isCanvas`'s new key both ways: Go sends `fontChains` and TS does not list it → `parseInbound` returns `undefined`; TS lists it and Go omits it → likewise.
- `Fonts.Chain` moves the "declared and non-empty" population out of five local guards into one. Re-point each caller's test at a member still in scope **and** assert the departed population under its new treatment — an empty chain must still be invisible to the projection and still refused by the property command.

**Manual checks:**
- Confirm `git status --porcelain` lists only intended paths before any commit, and that `README.md` is untouched (`md5 078d7d80d518d54af2fc04fb270d46b8`, 8470 bytes).
- Confirm no line was added to `goldenDigestRecord` or `declaredEpic2GateObligations`; this story owes neither.
- Confirm `fixtures/statement-signoff.json` and `fixtures/thai-stacked-marks/signoff.json` are unmodified.

## Auto Run Result

Status: done
Blocking condition: none

**Implemented change.** The `.folio` document's `fonts` map became writable. Six new command kinds
— `addFontChain`, `renameFontChain`, `deleteFontChain`, `addFontChainEntry`, `moveFontChainEntry`,
`removeFontChainEntry` — join the versioned vocabulary `ApplyComponentCommand` dispatches, each
reached through `wasm.Engine.Apply`, whose single `pushUndo` makes every edit one history entry by
construction. A rename rewrites **both** attachment points — `Element.Style.FontFamily` and
`TableExt.HeaderStyle.FontFamily` — across all three bands inside the one handler, so one undo
restores the map and the elements together. The refusals are the substance: a duplicate name, an
orphaning delete, an empty chain, an out-of-range index and the two projection bounds are each
refused as a `componentFailure` with `ElementID: ""` and `DataPath: "fonts.<name>"`.

The story also created the single authority the dispatch assumed already existed. "Declared **and**
non-empty" had been open-coded five times; `template.Fonts.Chain(name) ([]string, bool)` is now the
one rule, and all five sites route through it. No `.folio` format change, no authored key order, no
new fixture, no new matrix obligation, and **nothing minted in the diagnostic registry** — Design
Notes R3 held on all three measurements.

**Files changed.**
- `folio-go/internal/template/model.go` — `Fonts.Chain`, the single authority, doc-commented with the five copies it replaces.
- `folio-go/component_commands.go` — six command arms, `applyFontChainCommand`, `fontChainReferences` (both attachment points, three bands, document order), `fontChainOrphanMessage` (whole-id trimming under the host's 512-byte cut), `fontChainPath` (rune-safe under the host's 256-byte `DataPath` cut).
- `folio-go/page_setup.go` — `CanvasFontChain`, `CanvasProjection.FontChains`, `canvasFontChains`; `FontFamilies` now derived from it, so `FontChains[i].Name == FontFamilies[i]` holds by construction.
- `folio-go/render.go`, `folio-go/table_render.go` — routed through `Fonts.Chain`, own message text kept.
- `folio-designer/src/engine-protocol.ts` — `fontChains` on the type, in `isCanvas`'s `hasOnly` list, and validated against `fontFamilies`; `MAX_ENGINE_FONT_CHAIN_ENTRIES`.
- `folio-go/canvas_projection_wire_test.go` — the cross-language key tripwire, extended to the first nested projection object.
- `folio-go/component_commands_test.go`, `folio-go/wasm/engine_test.go`, `folio-go/canvas_body_text_bounds_test.go`, `folio-designer/src/{engine-protocol,engine-bounds-mirror,App,DataPanel,sheet-stack}.test.*` — coverage and the fixtures the new required projection field breaks.
- `folio-designer/src/App.tsx` is untouched; no chain-editor panel was built.

**Review findings breakdown.** 12 patches applied (1 high, 5 medium, 6 low); 5 items deferred; 8
rejected; 0 intent gaps and 0 bad-spec loopbacks, so `review_loop_iteration` stays 0. The high
finding is the one worth carrying forward: the pageFooter arm of `fontChainReferences` was
completely unmeasured, and the whole band could be deleted with the suite green — the AC's
"an element in each of the three bands" was asserted by no test until this pass.

**Follow-up review recommendation: true.** Patched counts by severity: high 1, medium 5, low 6. A
high-severity patched finding alone sets it; the score is also 3x5 + 1x6 = 21, well past 5.

**Verification performed** (measured, re-run independently of the implementing subagent):
- `cd folio-go && go test -count=1 ./...` — 1675 pass, 5 skip, exactly ONE distinct red: `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest (got 7, need >=20), the mandated permanent red. 13 packages ok, 1 failed. Baseline before the story was 1661 pass with the same single red.
- `go vet -tags=matrix ./...` — clean, exit 0. `gofmt -l folio-go` from the repo root — no output.
- Matrix, four legs each with `FOLIO_MATRIX_TARGET` exported: `darwin/arm64` ok 1.10s, `linux/amd64` ok 6.56s, `linux/arm64` ok 5.04s, `js/wasm` ok 10.96s — every leg printing real per-fixture digests. **Unset control**: passes in 0.00s and logs "this test asserts NOTHING and is a deliberate no-op", proving the four legs are not no-ops.
- `TestCrossTargetByteIdentity` ok 22.59s; `TestThaiStackedMarksSemanticSignOffIsRecorded` ok 0.39s.
- `cd lint && go test -count=1 ./...` — 4 packages ok (`genmanifest`, `licence`, `manifest`, `rules`). `-count=1` used throughout, per D-7.9.5.
- `cd folio-designer` — typecheck exit 0; lint exactly 4 pre-existing `only-export-components` warnings; `npm test` 34 files / 285 tests all passing (284 at baseline); `test:e2e:compile` exit 0.
- `shasum -a 256 fixtures/*/expected.pdf` — 22 lines, `diff` against a baseline captured **before any edit** is empty. Both human attestations (`fixtures/statement-signoff.json`, `fixtures/thai-stacked-marks/signoff.json`) have zero diff; `byte_neutrality_test.go` has zero diff, so no line was added to `goldenDigestRecord` or `declaredEpic2GateObligations`.
- `README.md` untouched: md5 `078d7d80d518d54af2fc04fb270d46b8`, 8470 bytes.
- `TestShippedFacesReproduceFromUpstream` did not run — it is reached only by a full `-tags=matrix` suite run, which the Verification section does not call for. No environmental failure is folded into any count.

**Mutation proofs.** Independently re-run rather than taken on report, and each mutation confirmed
to land before its result was believed — one candidate mutation aborted on an occurrence-count
assertion when a second, pre-existing `PageFooter.Elements` site (`assetKeyReferenced`) appeared,
and was redone by line number. Both arms of `fontChainReferences` were proved in both directions:
deleting the `style` arm reddens the style test while the headerStyle-only test stays green, and
deleting the `headerStyle` arm reddens the headerStyle-only test with the population vanishing
entirely. Deleting the whole pageFooter band was green before this review pass and now reddens two
named tests. `isCanvas`'s new key was proved both ways — removing `fontChains` from the `hasOnly`
list reddens 14 protocol tests, and tolerating a missing `fontChains` reddens the chain guard test.
`Fonts.Chain`'s non-empty half, deleted, reddens `TestEmptyFontChainIsInvisibleToTheProjectionAndRefusedByTheProperty`
with "Fonts.Chain accepted an empty chain". Each new refusal guard, deleted one at a time, reddens
exactly one specifically named test.

**Residual risks.** The five deferred items carry the real ones. The sharpest is that this story's
new per-chain entry bound is enforced at projection time, and `Engine.Load` projects — so a
pre-existing `.folio` with a chain over 64 entries, or a face name over 512 bytes, now fails to
open rather than merely being uneditable. The bound was directed by Task 6 and mirrors the
pre-existing `canvasFontFamilies` shape, so it is recorded rather than reversed. Second, the
refusal messages are specified at the wasm host wire but measured one layer below it, because
`folio-go/wasm/cmd/engine` is `//go:build js && wasm` and no compiled test can reach it; P11's
source-reading tripwire ties the two bound literals, which is the most that surface currently
allows. Third, `maxCanvasFontChainEntries = 64` is a number this story invented; nothing in the
repo declares more than 3 entries, so nothing existing is affected, but an embedding story may want
to revisit it.
