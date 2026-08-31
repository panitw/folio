---
title: 'Story 8.2: The chain editor sits where fonts are chosen'
type: 'feature'
created: '2026-08-31'
status: 'ready-for-dev'
baseline_revision: 'bc671da1e7f3abca93045f9281d8be62e8fe02ed'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized', 'multiple-goals']
deferred: []
---

<intent-contract>

## Intent

**Problem:** Story 8.1 shipped six engine commands that create, rename and delete font chains and add,
move and remove their entries — and **no TypeScript sender exists for any of them**. `fontChains` is
projected, validated and consumed by nothing in designer production code. An author can pick a chain in
the family control but cannot make, change or remove one, so a template still offers whatever chains its
starter file declared. Two browser-side defects sit directly on the path this story opens: the designer's
command encoders escape author text five different ways (one of them incompletely, and one splicing
author text into JSON unquoted), and the projection guard sorts chain names in UTF-16 order while Go
sorts by bytes — which, once an author can name a chain, terminates the engine worker and blanks the
canvas from two keystrokes.

**Approach:** Add an inline chain-editor section to the typography panel that sends the six existing
commands and re-projects; introduce one shared command-JSON authority and route every designer command
encoder through it; and correct the projection guard's ordering to Go's. The engine remains the sole
owner of the document and of every rule: the panel dispatches, displays the engine's own refusal text
verbatim, and holds no model, no validation and no copy of any refusal rule.

## Boundaries & Constraints

**Always:**
- **AD-15 — there is no TypeScript document model.** Every chain name, every entry and every ordering the
  panel shows is read from `CanvasProjection.fontChains` on the latest snapshot. No local list, no
  optimistic mutation, no derived cache that survives a snapshot. After any edit the panel re-renders from
  the engine's answer.
- **AD-16 — commands are opaque.** Each of the six edits is one command producing one revision and one
  undo entry (UX-DR20). The browser never re-implements, re-words, re-orders or pre-empts an engine rule.
- **AD-17 — the browser never measures text.** `canvas-authority-contract.test.ts`'s banned identifiers
  and its `placementPoint`/`pageStyle` adjacency seam hold unchanged.
- **The refusal text the author reads is the engine's own string, byte for byte** (bounded to 512 at the
  host). The panel supplies only *where* to put it, from the control it dispatched from — never *what* it
  says.
- **AD-22/AD-21 — all 22 golden digests unmoved.** This story changes no Go file, so byte identity holds
  by construction; it is still measured, not assumed.
- Every command payload is produced by one shared encoder that uses `JSON.stringify` for every value.
- Existing property-panel conventions are followed rather than re-invented: `role="alert"` +
  `aria-invalid` + `aria-errormessage` at the control, accessible names on every icon-only control,
  keyboard operability for reordering (UX-DR25).

**Block If:**
- Any change is required under `folio-go/` to satisfy an AC. This story is designer-only; a Go change
  means an AC was mis-scoped or a command kind is missing, and Story 8.1 owns that vocabulary. HALT with
  `blocked` / `go change required for a designer-only story`.
- Any of the 22 digests in `shasum -a 256 fixtures/*/expected.pdf` differs from the baseline captured
  before the first edit. HALT with `blocked` / `golden digest moved`. Never write `reader`, `date` or
  `examined` in `fixtures/statement-signoff.json` or `fixtures/thai-stacked-marks/signoff.json`.
- A new `.folio` format field, a new command kind, or a new projection key is found to be necessary.
  HALT with `blocked` / `format or command surface change required`.

**Never:**
- Never build Story 8.3's embedded-face entry shape (`{"asset": "<key>"}`), 8.4's rendering, or 8.5/8.6's
  catalogue. Never add a font media type, an asset-backed chain entry, or a catalogue list.
- Never change `folio-go/internal/layout/` or the `.folio` format. Never add an authored chain-key order
  (D-8.1.1).
- Never copy, paraphrase, pre-empt or predict an engine refusal in TypeScript — no client-side duplicate
  check, no client-side empty-chain check, no client-side orphan check, no message table.
- Never open a dialog for the chain editor: AC1 requires the same panel, no separate mode, no dialog
  stack.
- Never `git add -A` / `git add .`; never modify, move or commit `README.md`; never push; never branch.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Open the editor | A text component selected; document declares chains | The family control offers an affordance that reveals the chain editor **in the same panel section**; the list shown is `projection.fontChains`, in the engine's order | No error expected |
| No chains declared | `fontChains: []` | The editor states the document declares no chains and offers only "Add font chain" | No error expected |
| Add a chain | Author types name `heading`, one face | `addFontChain` sent with `{version, kind, name, entries}`; panel re-projects; the new chain appears because the engine returned it | No error expected |
| Add a duplicate name | Name equals a declared chain | Command **is sent** (no local pre-check); engine refuses `a font chain named "x" already exists` at `DataPath fonts.x` | Message rendered verbatim at the name control; document unchanged |
| Rename onto an existing key | `renameFontChain{name, to}` where `to` is declared | Engine refuses with the same string, `DataPath fonts.<to>` — the **destination**, not the source | Verbatim at the rename control |
| Delete an in-use chain | Chain named by elements and/or a table `headerStyle` | Engine refuses `font chain "x" is still named by e1, e2 and 3 more` | Verbatim at the delete control; the id list is the engine's, never recomputed in TS |
| Remove the last entry | Chain has one entry | Engine refuses `removing that entry would leave font chain "x" with no entries` | Verbatim at that entry's remove control |
| Reorder entries | `moveFontChainEntry{name, from, to}` | Entry order shown afterwards is the projection's, not a local splice | Out-of-range index → `entry index is out of range`, verbatim |
| Author name containing `"` and `\` | Name `a"b\c` | Command JSON is well-formed; engine accepts and the chain is projected back with that exact name | No error expected — a malformed-command refusal here is a defect |
| Author name containing a C0 control | Name `a\u0001b` | `JSON.stringify` emits `\u0001`; the command parses; the engine decides | No malformed bytes reach the engine |
| Chain names disagreeing across sort orders | Chains `\uE000` and `\u{1F600}` both declared | Guard accepts: it compares in **byte order**, matching Go, so the snapshot is not dropped | Canvas keeps rendering; worker is not terminated |
| Unlocated refusal | Command refused below `componentFailure` (arity, projection bound) → `ENGINE_REJECTED`, no `dataPath` | Message still shown at the control the panel dispatched from | Never silently swallowed |
| Chain entry naming a face this build does not ship | Entry `Helvetica` | Accepted by the engine; displayed as the projected string, unchanged | No error expected |

</intent-contract>

## Code Map

**Where the panel is, and the exact conventions to follow (edited)**
- `folio-designer/src/App.tsx:1146-1187` — `FontFamilyProperty`. A **search-and-select combobox**, not a
  `FieldSpec`: `role="combobox"`, `aria-label="Font family"`, `aria-expanded/-controls/-autocomplete/
  -activedescendant`, popup `<ul role="listbox" aria-label="Declared fonts">` with `role="option"`.
  Typed text is filter-only and is **never** committed; commit is choosing an option. `×` = "Clear Font
  family"; `⌄/⌃` = "Show fonts"/"Hide fonts". **The chain-editor affordance is a third button here.**
- `folio-designer/src/App.tsx:1007` — the `<PropertySection title="TYPOGRAPHY">` the editor lives inside;
  `:771` `fontFamilies` fed from `canvas.fontFamilies`. **`canvas.fontChains` has zero production
  consumers today** — it exists only in `engine-protocol.ts` and test fixtures.
- `folio-designer/src/App.tsx:891` `PropertyCommitError{field, selectionKey, elementId?, dataPath?,
  message}`; `:475-494` `applyProperties` (drops a stale refusal when generation or selection key moved);
  `:1000-1007` the per-control fan-out; `:1231` `componentDiagnosticDetail`; `:1232` `componentDiagnostic`.
  **This is the anchor-by-originating-control mechanism to copy** — the panel keys the error by what it
  *asked*, not by what the engine returned.
- `folio-designer/src/App.tsx:1130` (`PropertyDraft`) and `:1184` (`FontFamilyProperty`) — the two
  per-control error renderings: `<p id={errorId} role="alert" className="property-error">`, with
  `aria-invalid` + `aria-errormessage` on the input. `role="status" aria-live="polite"` is reserved for
  **non-error** global lines (`:764`, `:767`, `:783`) — do not use it for a refusal.
- `folio-designer/src/TableEditor.tsx` — the **only** ordered add/remove/reorder precedent. Copy its
  keyboard idioms (roving tabindex `matrixCell`, arrow/Home/End, 1-indexed positional labels `Move column
  N earlier/later`, `Remove column N`, `Add column after column N`, uncontrolled text cells committing on
  blur with a changed-value guard). **Do not copy its shell** — it is `role="dialog" aria-modal="true"`,
  which AC1 forbids. Its errors are per-dialog; the chain editor's are per-control.

**The command encoders — five different answers to one question (edited)**
- `folio-designer/src/component-property-command.ts:25` — the sole construction site, a **template
  literal**. `:28-32` `rawNumberLiteral` returns the author's string **verbatim** for the 11 fields in
  `pointFields` (`:7`) and `ratioFields` (`:17`). `:40` `quote` escapes only `\ " \n \r \t` — **not** C0
  controls, **not** lone surrogates. **This is DW-32's site.**
- `folio-designer/src/page-setup-command.ts:6-7` — the **same** unquoted-number splice, from free-text
  inputs (`App.tsx:887`). DW-32 does not name it; it must be fixed with it or the authority is not sole.
- `folio-designer/src/table-column-command.ts:3` — uses `JSON.stringify`. **The correct model.**
- `folio-designer/src/component-command.ts:52-67` (`bindComponentScalar`) and
  `folio-designer/src/component-asset-command.ts:18-34` — complete hand-rolled escapers, byte-identical
  copies of each other. `component-command.ts:17-19` `createComponentCommand` splices `type`/`band` into
  quotes, safe only because both are closed unions.

**The projection guard (edited)**
- `folio-designer/src/engine-protocol.ts:157-165` `fontChains` type; `:220` `isCanvas`'s `hasOnly` key
  list (a **subset** check); `:226` the sorted/unique rule — `names[index-1] >= name`, **UTF-16 code
  units**; `:235` `hasExactKeys(chain, ['name','entries'])`; `:266` the `CanvasComponent` `hasOnly` list.
  **This is DW-70's site.**
- `folio-designer/src/engine-client.ts:87` `#fail('PROTOCOL_INVALID', …)`; `:117-124` — detaches, calls
  `this.worker.terminate()`, rejects `ready` and every in-flight request. A guard disagreement is a dead
  worker and a permanently blank canvas, not a dropped frame.

**Read-only evidence, measured at `bc671da`**
- `folio-go/component_commands.go:93-104` the six kinds; `:1977` `applyFontChainCommand`; `:2011`
  `fontChainName`; `:2027` `declaredFontChain`; `:2045` `fontChainFace`; `:2056` `fontChainIndex`; `:2069`
  `addFontChain`; `:2119` `renameFontChain`; `:2156` `deleteFontChain`; `:2171` `addFontChainEntry`;
  `:2194` `moveFontChainEntry`; `:2222` `removeFontChainEntry`; `:2290` `fontChainOrphanMessage`;
  `:2266-2277` the reference walk over **both** `Element.Style` and `Element.Table.Value.HeaderStyle`.
  `:1952` `fontChainPath` → `DataPath` is `fonts` or `fonts.<name>`; `ElementID` is **always empty**.
- `folio-go/wasm/cmd/engine/main.go:236-244` — `*ComponentCommandError` matched **before** `*RenderError`
  (`:245`), emitted as `COMPONENT_INVALID` with `bounded(msg,512)`, `bounded(elementId,128)`,
  `bounded(dataPath,256)`; **`reportableMessage` (`:276`) is not on this path.** `:269` `ENGINE_REJECTED`
  reports a bare error verbatim.
- `folio-go/internal/template/parse.go:313-328` `decodeFonts` → `decodehelpers.go:95` `decodeStringArrayRaw`
  — **every chain entry must be a JSON string**. `model.go:147` `Fonts map[string][]string`. No
  `FontRecord`, no font media type, no `{"asset":…}` entry anywhere. **AC3's input cannot exist at HEAD.**
- `folio-go/page_setup.go:440-443` `CanvasFontChain{Name, Entries}`; `:266-427` `CanvasProjection` (no
  usage count, no referee list); `:181` `CanvasComponent.FontFamily *string`, set at `:1406-1410` only for
  text/table with `style.fontFamily` explicitly set. **`headerStyle` appears nowhere in `page_setup.go`.**
- Bounds, enforced at **both** the command path and the projection: `maxCanvasFontChainEntries` = 64
  (`component_commands.go:1918`), `maxCanvasPropertyString` = 512 bytes for chain and face names
  (`page_setup.go:39`), `maxCanvasFontFamilies` = 256 chains (`page_setup.go:431`).
- `folio-designer/src/property-prose-height.test.ts:41-42` — pins that `property-value-prose` is worn by
  **exactly one `<textarea>` and one `<input>`**. A chain-editor input reusing that class reddens it.
- oxlint baseline is **exactly 4** `only-export-components` warnings (`pdf-viewer.tsx:16,17`;
  `App.tsx:1218,1225`). Keep new non-component exports in `.ts`, not `.tsx`.
- `folio-designer/e2e/browser-native-roundtrip.spec.ts:84` addresses `getByRole('combobox', {name: 'Font
  family'})`. That name must not change.
- `folio-go/byte_neutrality_test.go:92` `goldenDigestRecord` — **22** entries, equal to the 22 dirs holding
  an `expected.pdf`. `matrix_test.go`'s `matrixDocuments` is **23** (it includes `hidden-image`, which
  ships no `expected.pdf`) — do not derive the digest count from it.

## Tasks & Acceptance

**Execution:**

1. `folio-designer/src/command-json.ts` (new) — create the **single command-JSON authority**: a `quote`
   built on `JSON.stringify`, and a numeric encoder that returns a discriminated refusal (not a spliced
   string) when the author's text is not a finite number. Doc-comment it with the measured population it
   replaces: five encoders, three different escaping answers, two unquoted-number splices. Per D-8.1.3 this
   authority is **created**, not found.
2. `folio-designer/src/component-property-command.ts` — route `quote` and the 11 numeric fields through
   Task 1's module; `rawNumberLiteral`'s verbatim splice is replaced by the validating encoder, so an
   unparseable entry is refused at the field before any bytes are built (this is what DW-32 asks for).
   Preserve the deliberate property documented at `:8-16` — Go alone decides the unit — by re-emitting the
   parsed number, not by reformatting it.
3. `folio-designer/src/page-setup-command.ts` — same substitution at `:6-7`. Unnamed by DW-32 and carrying
   the identical defect; leaving it makes Task 1's authority not sole.
4. `folio-designer/src/component-command.ts`, `folio-designer/src/component-asset-command.ts`,
   `folio-designer/src/table-column-command.ts` — route their three escapers through Task 1's module and
   delete the two byte-identical hand-rolled copies.
5. `folio-designer/src/font-chain-command.ts` (new) — six builders for the six kinds, using Task 1's
   encoder for every value: `addFontChain{name, entries[]}`, `renameFontChain{name, to}`,
   `deleteFontChain{name}`, `addFontChainEntry{name, index, face}`, `moveFontChainEntry{name, from, to}`,
   `removeFontChainEntry{name, index}`, each with `version` and `kind` and **exactly** the field arity
   `componentFields` expects (4/4/3/5/5/4). A `.ts` file, not `.tsx`, to keep the oxlint baseline at 4.
6. `folio-designer/src/engine-protocol.ts:226` — replace the UTF-16 `>=` comparison with a **byte-order**
   comparator matching Go's `slices.Sorted` over UTF-8, applied to `fontFamilies` (and to `fontChains`'
   names, which must equal it). Comment it with the measured pair that motivated it. **This closes DW-70.**
7. `folio-designer/src/FontChainEditor.tsx` (new) — the inline editor, rendered inside the TYPOGRAPHY
   `PropertySection` and revealed by a new affordance on `FontFamilyProperty`. Reads
   `projection.fontChains` on every render. Per chain: a name control, a rename action, a delete action, an
   ordered entry list with per-entry move-earlier / move-later / remove actions and an add-entry control;
   plus an add-chain control. Every icon-only control carries an accessible name including its 1-indexed
   position, on `TableEditor`'s pattern. Reordering is operable by keyboard alone (UX-DR25). Do **not**
   reuse the `property-value-prose` class.
8. `folio-designer/src/App.tsx` — wire the editor: an `onCommand` that posts the built command through the
   same path property commits use, and a refusal state keyed by the **originating control** (chain name,
   entry index, action), mirroring `PropertyCommitError`'s `field`+`selectionKey` shape at `:891`/`:475-494`
   including the stale-generation drop. Render the engine's `message` **verbatim** at that control as
   `<p role="alert" className="property-error">`, with `aria-invalid`/`aria-errormessage` on the control.
   Handle a refusal carrying **no** `dataPath` (`ENGINE_REJECTED`) identically — the anchor is what the
   panel asked, not what the engine returned.
9. `folio-designer/src/canvas-authority-contract.test.ts` — add a prohibition, alongside the existing
   banned-identifier set, on the engine's refusal vocabulary appearing anywhere in `src/**`: the literals
   `already exists`, `is still named by`, `with no entries`, `must declare at least one entry`,
   `is out of range`, `exceeds the projection bound`, `is declared`. Red-proof it by adding one of them to
   a source file and observing the failure. This is the mechanism that stops a TypeScript copy of a rule.
10. `folio-designer/src/font-chain-command.test.ts` (new) — assert the encoded bytes for every builder,
    including a chain name containing **both `"` and `\`** (`a"b\c`) and one containing `\u0001`, and
    assert each parses as JSON with the expected shape and arity. Assert the old encoder's output for the
    same inputs would not parse, so the fix is measured and not asserted.
11. `folio-designer/src/App.test.tsx` — panel behaviour: the editor opens **within** the typography section
    (no `role="dialog"` appears); the list shown equals the fixture's `fontChains`; each of the six actions
    dispatches the expected command; a refusal renders the engine's exact string at the originating
    control with `role="alert"`; and — the anti-pre-emption assertion — **submitting a duplicate name and
    an emptying remove each still dispatch a command**, proving no local rule blocked them. Extend the
    shared canvas fixture with a multi-chain, multi-entry `fontChains`.
12. `folio-designer/src/engine-protocol.test.ts` — add the DW-70 cases: `['\uE000', '\u{1F600}']` (Go's byte
    order) is **accepted**, and a genuinely out-of-byte-order pair is still rejected. Red-proof by
    restoring the UTF-16 comparator and observing the first case fail.
13. `folio-designer/src/canvas-font-stack.test.ts` — add a **tripwire** recording that the fragment stack is
    a stylesheet constant that does not consult the document's chains, naming DW-35 and the measured
    obstacle (the browser declares its three `@font-face` families as `IBM Plex Sans` / `IBM Plex Sans
    Thai` / `IBM Plex Mono` while the engine's face names are `Noto Sans` / `Noto Sans Thai` / `Noto Sans
    SC`, so no name mapping exists on either side). A comment asserting a negative carries a test's
    evidentiary burden; this is that test.
14. `folio-designer/e2e/component-properties.spec.ts` — extend the compile-only witness to exercise the
    editor's controls, and confirm `browser-native-roundtrip.spec.ts:84`'s `'Font family'` accessible name
    is unchanged. State in the spec's own header that this is compiled, never executed.
15. `_bmad-output/implementation-artifacts/deferred-work.md` — at the story's close: mark **DW-32** and
    **DW-70** CLOSED with the closing revision; amend **DW-35** with the measured name-mapping obstacle and
    the corrected owner recommendation (Story 8.4, whose AC4 is DW-35 stated as an acceptance criterion);
    file the two new findings recorded in Design Notes N4 and N5.

**Acceptance Criteria:**
- Given the whole change, when `git diff --stat` is inspected, then **no file under `folio-go/` is
  modified** — the story is designer-only, and that is what makes its byte-identity claim structural.
- Given a chain editor open on a document with two chains, when any of the six edits is committed and
  accepted, then exactly one revision advances, exactly one undo entry is pushed, and the values displayed
  afterwards are re-read from the new projection — proved by asserting the panel renders a snapshot the
  test supplies rather than the value the test typed.
- Given each of the three named refusals — orphaning delete, emptying remove, duplicate name — when the
  engine refuses, then the string rendered to the author is `===` to the engine's `message` field, at the
  control that dispatched it, with `role="alert"` and `aria-errormessage` wired.
- Given Task 9's prohibition, when any of the seven refusal-vocabulary literals is added to `src/**`, then
  the contract test goes red — proved by adding one, not inferred.
- Given the shared encoder, when each of the six chain builders and each existing encoder is fed a value
  containing `"`, `\` and `\u0001`, then every produced payload is valid JSON — proved per builder.
- Given `fontChains` containing `\uE000` before `\u{1F600}`, when `parseInbound` runs, then it returns a
  snapshot rather than `undefined`, and the worker is not terminated.
- Given the whole change, when `shasum -a 256 fixtures/*/expected.pdf` is run, then it prints **22** lines
  identical to the baseline captured before the first edit.

## Spec Change Log

## Review Triage Log

## Design Notes

**N1 — How the panel shows a refusal while holding no rule of its own, and what stops a TypeScript copy.**
Three properties, each separately measured:
(a) *The text is the engine's.* A chain refusal travels as `*ComponentCommandError`, matched at
`main.go:236-244` **before** `*RenderError`, emitted as `COMPONENT_INVALID` with `bounded(msg, 512)` and
**never** through `reportableMessage` (`:276`). So the author's sentence already exists, verbatim, before
the browser sees it. The panel's only job is placement.
(b) *The placement is browser knowledge, not engine knowledge.* The panel knows which control it
dispatched from; it does **not** need `DataPath` to place the message. This matters because three chain
refusal paths carry no location at all — `componentFields` arity (`:1264`), the `Canvas(installed)`
projection-bound errors surfaced raw at `:1996`, and serialize/parse failures at `:1979/:1983/:1987` — all
of which reach the browser as `ENGINE_REJECTED` with no `dataPath`. Anchoring by origin handles located
and unlocated refusals with one rule. This is exactly `PropertyCommitError`'s existing shape.
(c) *Nothing is pre-empted.* The panel performs **no** validation before dispatching: a duplicate name, an
emptying remove and an orphaning delete are all sent and all refused by the engine. Task 11 asserts the
dispatch happens, which is the only way to prove a local rule is absent — a passing "the error shows"
test is satisfied equally well by a TypeScript copy. Task 9's source-level prohibition is the second
lock, on `canvas-authority-contract.test.ts`'s existing banned-regex idiom.

**N2 — AC3 is not satisfiable before Story 8.3, measured; the deliverable is its negative half.**
`decodeFonts` (`parse.go:313-328`) routes every chain through `decodeStringArrayRaw`
(`decodehelpers.go:95`), which requires each element to be a JSON **string**. `Fonts` is
`map[string][]string`. There is no `FontRecord` type in the repository, no font media-type handling, and
`"asset"` in the format refers exclusively to image elements. `CanvasFontChain.Entries` is `[]string` and
`canvasFontChains` clones the raw face-name slice. **At `bc671da` a chain entry cannot name an embedded
face at all**, so AC3's positive half — "reads as the face's family and style from the projection" — has
no projection to read: there is no family/style pair anywhere.
Per the dispatch's authorization, the **forward-compatible display rule** is adopted rather than a halt:
*the displayed text of a chain entry is the projected entry, unmodified.* The panel performs no parsing,
no key detection, no extension stripping and no file-name handling — so whatever shape 8.3 gives an entry
is what the author reads, with no browser-side rule to update. AC3's negative half ("never as an asset key
or a file name") is therefore delivered and assertable **today**, and its positive half is 8.3's to
deliver by putting a family and style into the projection. `hasExactKeys(chain, ['name','entries'])` at
`engine-protocol.ts:235` rejects in **both** directions, so 8.3 cannot change the entry shape without
this panel's validator changing in the same commit. Do not build the entry shape here.

**N3 — DW-35 does NOT close here, and its stated ground is falsified.** Three measurements:
(i) The register says the fix "needs the projected per-component chain, which 8.2 needs anyway". Both
halves are wrong at HEAD. Story 8.1 **already** landed `CanvasComponent.FontFamily` (`page_setup.go:181`,
set at `:1406-1410`) and `CanvasProjection.FontChains[].Entries` (`:294`, `:440-443`), so the projection is
present; and 8.2 edits the document-level `fonts` map and **never needs the per-component join**. The
efficiency argument for co-locating them does not exist.
(ii) The real obstacle is unrecorded anywhere: the browser's three `@font-face` families are declared as
`'IBM Plex Sans'`, `'IBM Plex Mono'` and `'IBM Plex Sans Thai'` (`build-wasm.mjs:79`), mapping the shipped
**Noto** files under IBM Plex names, while the engine's face names in a chain are `"Noto Sans"`,
`"Noto Sans Thai"`, `"Noto Sans SC"`. A chain's entries therefore **cannot** be used as CSS family names.
The fix needs a face-name → CSS-family mapping that exists on neither side, and a decision between
renaming the generated `@font-face` families (rippling into `tokens.css:11` `--font-page`, its three type
tokens, and `design-contract.test.ts:34`) and generating a map. That is a design-system decision above a
builder's authority.
(iii) What 8.2 **does** make newly reachable is narrow and should be stated rather than fixed: an author
can now create a chain whose first covering entry differs from `Noto Sans` (e.g. `["Noto Sans Thai"]`),
after which the engine measures with that face while the browser paints the fixed Latin-first stack. Task
13 records this as a tripwire test rather than a comment. Recommended owner: **Story 8.4**, whose AC4 —
"the preview measures with that same face through the same engine path, so the canvas and the PDF keep one
measurement authority" — *is* DW-35 written as an acceptance criterion.

**N4 — DW-32 is understated in the register, and the dispatch's premise about it is false.** The dispatch
says this story "carries author-typed chain names through that same splice". Measured: it does not.
`rawNumberLiteral` (`component-property-command.ts:28-32`) is reached only by the 11 numeric fields in
`pointFields`/`ratioFields`; every string field, `fontFamily` included, already routes through `quote`
(`:40`), which does escape `"` and `\`. And 8.2 writes a **new** sender module regardless, so no chain name
would have touched that splice. The conclusion survives in a sharper form: there is **no single command-JSON
authority** — five encoders give three different answers (`JSON.stringify`; two byte-identical complete
hand-rolled copies; one incomplete copy escaping only five characters; and none at all for numbers) — which
is D-8.1.3's exact shape, so the authority is created (Task 1) and that is a real change in this story's
size.
Two defects found while measuring, both worse than recorded:
- `quote` at `:40` does not escape C0 controls. `quote("a\u0001b")` emits a raw `0x01` inside a JSON
  string — invalid JSON — and the prose textarea's paste path (`App.tsx:1107-1126`) accepts pasted
  `text/plain` carrying such bytes.
- **The numeric splice is a command-shape injection, not merely malformed bytes.** The register rates
  DW-32 MEDIUM on "no bad value reaches the document"; that is false. A value of
  `0}},"ids":["other"],"changes":{"width":{"op":"set","value":10` produces **valid** JSON with duplicate
  keys, which Go decodes into `map[string]json.RawMessage` (last key wins) while `componentFields(raw, 4)`
  still counts 4 — so the command mutates a different component's different property. Escalation to a
  different `kind` is blocked only by an arity coincidence (`componentFields(raw, 3)` at `:1721` seeing 5),
  which is an accident, not a guard. Raise DW-32 to HIGH at close.

**N5 — findings against the epic and register text, measured at `bc671da`.**
(i) Story 8.2's `Covers:` line names `FR52 · UX-DR13, UX-DR24, UX-DR25` but omits **AD-15** and **AD-16**,
whose Rules are the literal substance of AC2 and AC1 — the same omission corrected on Story 8.1's line at
its own plan gate.
(ii) The epic's claim that "the family control is already a search-and-select over the engine-projected
list of declared chains" is **true** (`App.tsx:1146-1187`), with the precision that the list it searches is
`fontFamilies` (names), while `fontChains` (names + entries) is projected, validated and consumed by
nothing in production code.
(iii) **No AC in this story asks for a usage count, and the panel must not invent one.** Per D-8.1.2's
standing check: a count derived from the projection would walk `style.fontFamily` only, because
`CanvasComponent.FontFamily` is the sole projected attachment point and `headerStyle` appears nowhere in
`page_setup.go` — so a chain shown as "unused" could still be refused deletion naming an id the panel had
no field for. The engine's own refusal already names every id across **both** attachment points
(`component_commands.go:2266-2277`), in document order. The panel shows no count and defers to that
message; that is the answer to the standing check.
(iv) DW-70's "no attributable error" is loose: `PROTOCOL_INVALID` / "The engine sent an invalid protocol
message" is raised, but names no field, chain or element, and the worker is **terminated**
(`engine-client.ts:117-124`) rather than the frame dropped. Reachability is total: `fontChainName`
(`:2011`) accepts any non-empty ≤512-byte UTF-8 string, so two chain names brick the canvas.
(v) `canvas_projection_wire_test.go` records the top-level key list (`:47`) and the nested
`CanvasFontChain` list (`:76`) but **not** `engine-protocol.ts:266`'s `CanvasComponent` key list. A future
per-component projection field would blank the canvas with every test on both sides green. This story adds
no projection field, so it is reported rather than fixed — but it is the gap DW-35's eventual fix walks
into.
(vi) `matrixDocuments` is **23**, not 22 (`hidden-image` ships no `expected.pdf`). The digest count is 22,
via `goldenDigestRecord` and `fixtures/*/expected.pdf`, which agree exactly.
(vii) `commandInt` (`:510`) still omits the `folio: ` prefix its `commandString` sibling (`:1271`) carries,
so an author moving a chain entry reads `"index must be an integer"` beside `"folio: name must be a
non-empty string"` (DW-72). This story is the first to put both in front of an author but adds no Go
command field, so DW-72's stated owner does not become this story; report it, do not fix it.

**N6 — UX-DR13's ten states, and which have a referent here.** Populated (chains listed), Empty (a
document declaring no chains), Diagnostic and Error (a refused edit) are the four this surface owns and
must draw. Loading, Rendering, Unsaved changes and Stale preview are already owned globally by the app
shell and the editor inherits them. *Empty — no sample data* and *Empty — table with no columns* have no
referent in a chain editor. A chain with no entries is **unreachable** by construction: the engine refuses
both the create (`a font chain must declare at least one entry`) and the emptying remove, so it is not a
state to draw. Stating this is the obligation; drawing an unreachable state is not.

## Verification

Capture `shasum -a 256 fixtures/*/expected.pdf > /tmp/digests-baseline.txt` **before the first edit**.

**Commands:**
- `cd folio-go && go test -count=1 ./...` — expected: exactly ONE distinct red,
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (got 7, need >=20), the mandated permanent red.
  Report pass/fail counts, not "green". Anything else is a regression. This story changes no Go file, so a
  new red here means something outside the diff moved — say so rather than absorbing it.
- `cd folio-go && go vet -tags=matrix ./...` — expected: clean.
- `gofmt -l folio-go` from the repo root — expected: no output. (`lint/…/licencegraph_test.go` carries
  DW-23 and is outside this path.)
- `cd folio-go && FOLIO_MATRIX_TARGET=darwin/arm64 go test -count=1 -tags=matrix -run TestTargetRenderHash -v .`
  — and once each with `linux/amd64`, `linux/arm64`, `js/wasm` **exported**, **plus one unset control**.
  The control must show the assertion skipped while the four exported legs assert; unset, this test is a
  no-op that passes. Report all five.
- `cd folio-go && go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity .` — expected: pass.
- `cd folio-go && go test -count=1 -tags=matrix -run TestThaiStackedMarksSemanticSignOffIsRecorded .` —
  expected: pass. (Note this is **not** `TestGoldenDigestAgreesAtEveryDeclaredSite`; do not conflate them.)
- `cd lint && go test -count=1 ./...` — **`-count=1` is MANDATORY** (D-7.9.5): the `rules` package walks a
  directory and Go's cache does not track `ReadDir`, so a cached `ok` is no measurement at all. Expected:
  clean. `TestFloatTypedTestScopeInventory` pins five sites **by line number** in
  `shaping_expectations_test.go` and `internal/fontset/{fontset,vendorboundary}_test.go`; this story
  touches none of them, so a red there means a line moved and must be reported, not re-pinned.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` —
  **this is the real gate for this story.** Expected: typecheck and e2e-compile clean; oxlint exactly
  **4** pre-existing `only-export-components` warnings (a fifth is a regression — keep new non-component
  exports in `.ts`). Baseline is **285 tests / 34 files**; report the number after, do not gate on it.
  What the suite must **prove**, named test by named test: (a) the six builders emit valid JSON for `"`,
  `\` and `\u0001`; (b) the panel dispatches on a duplicate name and on an emptying remove, so no local
  rule pre-empts the engine; (c) the rendered refusal string is `===` to the engine's `message`;
  (d) `parseInbound` accepts Go's byte order for `['\uE000','\u{1F600}']`; (e) no `role="dialog"` appears
  when the editor opens; (f) the refusal-vocabulary prohibition fires.
- `shasum -a 256 fixtures/*/expected.pdf | diff - /tmp/digests-baseline.txt` — expected: **22 lines, empty
  diff**. Do not use `matrixDocuments`' 23 as the count.

**Mutation proofs (run, never inferred):**
- **Deletion is the cheapest screen.** For each new guard and each new prohibition, delete it and name the
  test that goes red — including Task 9's banned-vocabulary regex, proved by adding one of the seven
  literals to a source file.
- Restore the UTF-16 comparator at `engine-protocol.ts:226` and confirm the DW-70 acceptance case goes red;
  then confirm a genuinely out-of-order pair is still rejected. Both directions, per D-7.9.6's rule.
- Restore `rawNumberLiteral`'s verbatim splice and confirm the `a"b\c` and injection cases go red.
- **The departed population, re-asserted.** Task 1 moves escaping out of four local encoders into one. For
  each caller, re-point its existing test at a member still in scope **and** assert the departed population
  under its new treatment: the two complete hand-rolled escapers must still escape lone surrogates and C0
  controls after routing through `JSON.stringify`.
- The panel's re-projection is proved by supplying a snapshot whose value differs from what the test typed
  and asserting the **snapshot's** value renders — a test that types and reads back the same string cannot
  distinguish re-projection from a local model.

**What remains UNPROVEN, and must be said plainly rather than implied by a compile pass:**
`npm run test:e2e:compile` is `tsc --noEmit` only; `npm run test:e2e` (Playwright) appears in no workflow
and browser e2e is deferred by D-000.4. **This is a panel story, so the surface an author actually touches
is the one nothing executes.** Specifically unproven: that the editor is keyboard-operable end to end in a
real browser; that focus is visibly rendered per UX-DR25; that `role="alert"` is actually announced; that
the six commands survive the real worker round trip; that the DW-70 fix keeps a real worker alive; and that
the affordance opens without disturbing the combobox at
`e2e/browser-native-roundtrip.spec.ts:84`, which is the repository's only cross-boundary authoring witness
and is compiled, never run. jsdom applies no stylesheet, so no unit test can prove the editor's appearance
or focus ring either. State this in the Delivery Log; never report the e2e work as verified.

**Manual checks:**
- `git status --porcelain` lists only intended paths before any commit; `README.md` untouched
  (`md5 078d7d80d518d54af2fc04fb270d46b8`, 8470 bytes).
- `fixtures/statement-signoff.json` and `fixtures/thai-stacked-marks/signoff.json` unmodified; no line
  added to `goldenDigestRecord` or `declaredEpic2GateObligations`.
- `git diff --name-only` contains **no** path under `folio-go/`.

## Auto Run Result

Status: ready-for-dev
Blocking condition: none

Planned at baseline `bc671da1e7f3abca93045f9281d8be62e8fe02ed` on `main`, tree clean. Dispatch directed
`Halt after planning.`, so no implementation code was written and nothing was committed. The workflow's
own step 1 regenerated `_bmad-output/implementation-artifacts/epic-8-context.md` (the cached context was
stale against `epics.md`); that modification is left uncommitted for the orchestrator.

Warnings: `oversized` (the spec is ~5,300 words against the template's 900-1600 token target; the Code Map
and Design Notes carry the measured investigation this gate was asked for) and `multiple-goals` (the chain
editor, the shared command-JSON authority closing DW-32, and the projection-guard ordering closing DW-70
are three independently shippable goals, all re-owned to this story by the deferred register; not split,
per the workflow's own rule).

Dispositions ruled at this gate, with grounds recorded in Design Notes: AC3's positive half is
unsatisfiable before Story 8.3 and is delivered as a forward-compatible display rule (N2); DW-35 does not
close here and its stated placement ground is falsified (N3); DW-32 closes here, understated in the
register and raised to HIGH (N4); DW-70 becomes fully reachable and closes here (N5 iv). No intent gap was
found.
