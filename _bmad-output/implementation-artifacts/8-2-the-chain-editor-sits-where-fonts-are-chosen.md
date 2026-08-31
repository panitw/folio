---
title: 'Story 8.2: The chain editor sits where fonts are chosen'
type: 'feature'
created: '2026-08-31'
status: 'done'
baseline_revision: '3f9205790598831d47a44e080daaa8d4d5a3245a'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized', 'multiple-goals']
deferred:
  - summary: >-
      The two hand-rolled JSON escapers in component-command.ts and component-asset-command.ts
      corrupt non-BMP characters, dropping the low surrogate of an astral pair.
    evidence: |-
      Both iterate by code point but escape from charCodeAt(0), so quote('a\u{1F600}b') emits
      "a\ud83db" and the low surrogate is lost. Demonstrated by executing the function during
      review. Go's encoding/json then substitutes U+FFFD, so a bind segment or asset key holding
      an astral character silently becomes a DIFFERENT path and the engine's refusal names a path
      the author never picked. component-command.test.ts's "complete JSON escaping" case uses only
      BMP inputs ('a.b', 'line\nbreak', '\u0000'), so the defect is invisible to it. This is
      PRE-EXISTING and owned by DW-32 / Story 15.2a, but the register records those encoders as
      merely unconverted rather than WRONG — 15.2a must know it is fixing a live bug.
    location: >-
      folio-designer/src/component-command.ts:52-68, folio-designer/src/component-asset-command.ts:18-34
    severity: high
  - summary: >-
      Nothing ties font-chain-command.ts's six command kinds and field arities to the Go dispatch
      table and componentFields counts they must match.
    evidence: |-
      Both sides assert against independently hand-written literals: font-chain-command.test.ts and
      App.test.tsx on the TypeScript side, component_commands_test.go on the Go side. Renaming
      "from"/"to" on moveFontChainEntry, or changing one componentFields(raw, N), in Go alone leaves
      every test in both languages green while the designer dispatches payloads the engine refuses
      with an arity error at every move. canvas_projection_wire_test.go closes exactly this seam for
      the Go->browser projection direction; the browser->Go command direction has no counterpart.
      Pre-existing for the five older encoders; this story adds six more kinds to the same gap.
    location: >-
      folio-designer/src/font-chain-command.ts:26-48 vs folio-go/component_commands.go:93-104
    severity: medium
  - summary: >-
      The isolable half of invalidatePreview() on the chain path — the token bump that revokes an
      in-flight render — remains unverified.
    evidence: |-
      A test asserting that an accepted chain command marks a rendered PDF stale was added and
      passes, but deleting invalidatePreview() leaves all tests green: re-entering Preview
      re-requests identity and renderPreview's own identity-mismatch branch sets the stale reason
      independently, so the user-facing property holds either way. Isolating the token bump needs a
      render held open ACROSS the chain command, which was not built. Reported by the implementer
      rather than claimed as covered.
    location: >-
      folio-designer/src/App.tsx:516
    severity: medium
  - summary: >-
      FontChainEditor reads its form values through document.getElementById on the global document.
    evidence: |-
      typed() resolves ids against the global document, so a second mount (two App instances in one
      test file, a future portal or preview pane) collides on ids and reads the wrong field. A useRef
      map would be local, typed, and immune. Works today because exactly one App mounts.
    location: >-
      folio-designer/src/FontChainEditor.tsx
    severity: low
---

## In plain terms (read this first if you just want the gist)

*Non-normative. The intent contract below governs the implementation; where the two differ, the
contract wins.*

A template hands a document a set of named font families, each one an ordered list of typefaces to
fall back through. Until now an author could pick one of those families but could never make, rename
or remove one, so a document was stuck with whatever its starter file happened to declare. This story
delivers the missing half: an author can now build those families from the very panel where fonts are
chosen — creating a family, renaming it, deleting it, and adding, reordering or removing the typefaces
inside it — without leaving that panel or opening a separate window.

The engine stays the only thing that decides whether an edit is allowed. The panel asks; it never
judges. When the engine refuses — a name already taken, a family still in use, a last remaining
typeface being removed — the author reads the engine's own words, unaltered, at the control they just
used. Nothing is second-guessed in the browser first, and two separate safeguards exist to keep it
that way.

Two faults sitting on this path were repaired: one that blanked the canvas the moment an author typed
certain names, and one that mangled pasted text before the engine could explain it.

What is not shown: all of this was measured in a simulated browser. Nobody has clicked it.

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
commands and re-projects; fix `quote()` in `component-property-command.ts` so it escapes the whole of
U+0000-U+001F rather than five characters; and correct the projection guard's ordering to Go's. **The
shared command-JSON authority and the numeric splice are Story 15.2a's, not this story's** (D-8.2.6). The engine remains the sole
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
- The six chain-editor commands encode every value with `JSON.stringify`, and `quote()` in
  `component-property-command.ts` is routed through `JSON.stringify` so it escapes all of U+0000-U+001F.
  **No other encoder is touched and no shared authority is built here** (D-8.2.5, D-8.2.6).
- Existing property-panel conventions are followed rather than re-invented: `role="alert"` +
  `aria-invalid` + `aria-errormessage` at the control, accessible names on every icon-only control,
  keyboard operability for reordering (UX-DR25).

- **SCOPE SETTLED 2026-08-31 (D-8.2.4 … D-8.2.7). This story carries DW-70 and a MINIMAL `quote()`
  fix. It does NOT carry DW-32.**

- **DW-70 is IN SCOPE as a precondition** — this story is what makes an author able to name a chain,
  so shipping the editor without it ships a feature whose **second keystroke terminates the worker**.
  Same category as Story 8.0: a defect the story makes reachable is a precondition of the story, not
  a competitor to it.
  **⚠ THE GUARDRAIL IS THE IMPORTANT HALF: Go's byte ordering is NORMATIVE and must not change.**
  `fonts` keys are sorted into the canonical `.folio` under AD-9, so **Go's sort IS the byte order of
  the document.** The mismatch is one line to fix on either side, and **changing the Go comparator
  would move golden bytes** for any document whose chain names cross the boundary. **The TS guard
  adopts code-point ordering to match Go** — which for UTF-8 byte order is the same sequence —
  **never the reverse.** Say this in the story, or someone fixes the cheaper-looking side.

- **`quote()` IS INCOMPLETE, and that IS on this story's path.** `component-property-command.ts:41`
  escapes `\`, `"`, `\n`, `\r`, `\t` — **and nothing else.** JSON requires escaping **all** of
  U+0000–U+001F. A chain name carrying any other control character (pasted, most plausibly) produces
  **invalid JSON**, which Go rejects with a generic failure rather than *"that name is not
  allowed"*. This inverts the dispatch's premise a second way: chain names do route through the
  quoter rather than the splice — **and the quoter is itself broken**, and this story is the first
  to make chain names author-supplied on that path.
  **Fix it MINIMALLY: route `quote()` through `JSON.stringify`. One line.** It is **not** the
  shared-authority consolidation and **must not be allowed to grow into it** — that is Story 15.2a's.
  **Engine-side name validation does not substitute:** the JSON is **malformed before Go can see the
  name**, so the engine's rule cannot run. An ordering fact, not a preference.

- **DW-32 IS OUT OF SCOPE — it is Story 15.2a**, filed in Epic 15 and sequenced before the tag. Do
  **not** build the shared command-JSON authority, do **not** touch `rawNumberLiteral`, and do
  **not** consolidate the five encoders. **HALT if an AC seems to require any of it.**
  **Both stories touch `component-property-command.ts`.** This one changes `quote()` only.

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

1. `folio-designer/src/component-property-command.ts` — route `quote()` (`:41`) through
   `JSON.stringify`. **One line.** It escapes `\`, `"`, `\n`, `\r`, `\t` and nothing else, while JSON
   requires all of U+0000-U+001F, so a chain name carrying any other control character produces invalid
   JSON that Go rejects generically instead of naming the field. **Leave `rawNumberLiteral` untouched, and
   leave the other four encoders untouched** — the consolidation and the numeric splice are Story 15.2a's
   (D-8.2.6), which must re-read this file rather than assume its earlier shape.

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
15. `_bmad-output/implementation-artifacts/deferred-work.md` — at the story's close: mark **DW-70**
    CLOSED with the closing revision. **DW-32 is NOT closed by this story** — it is Story 15.2a's, and its
    severity re-rating to HIGH and the injection mechanism behind it are recorded there; amend **DW-35** with the measured name-mapping obstacle and
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
- Given the six chain builders and `quote()`, when each is fed a value containing `"`, `\` and a C0
  control **other than the five `quote()` already handled**, then every produced payload is valid JSON —
  proved per builder. **The other four encoders are out of scope and are not exercised here.**
- Given `fontChains` containing `\uE000` before `\u{1F600}`, when `parseInbound` runs, then it returns a
  snapshot rather than `undefined`, and the worker is not terminated.
- Given the whole change, when `shasum -a 256 fixtures/*/expected.pdf` is run, then it prints **22** lines
  identical to the baseline captured before the first edit.

## Spec Change Log

## Review Triage Log

### 2026-08-31 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 12: (high 0, medium 6, low 6)
- defer: 4: (high 1, medium 2, low 1)
- reject: 5
- addressed_findings:
  - `[medium]` `[patch]` `followFocus.current` was cleared only inside the `[chains]` effect, so a refused command left stale target ids that the next unrelated snapshot consumed — reproduced by refusing a move then committing `Align center`, which stole focus onto `Move entry 2 of font chain 2 earlier`. Replaced with a settle-driven `pending` ref consumed when the command settles rather than when the list next changes. The implementer found a second defect while writing the test: deciding "the list moved" by `chains` ARRAY IDENTITY is wrong, because Go may return an equal list and a fixture may reuse one array across snapshots, which also made the test unfalsifiable; it now compares a value signature over names and entries.
  - `[medium]` `[patch]` A refusal left focus on `document.body`: pressing a control set `busy`, which set `disabled`, which blurred it, and only the success path restored focus — so the `role="alert"` was unreachable by keyboard and `aria-errormessage` hung off a control the author had left. Focus now returns to the originating control on refusal.
  - `[medium]` `[patch]` `applyFontChain` early-returned on `fileBusy` while the editor received `busy={fontChainBusy}` only, so during a save/open every chain click was a silent no-op. Now `busy={fontChainBusy || fileBusy}`, with a test holding `fileAccess.open` open.
  - `[medium]` `[patch]` Coverage gap proved by mutation: replacing `applyFontChain`'s body with a bare request plus `setCurrentSnapshot` — deleting `invalidatePreview()`, `setHistoryAvailability(...)` and the revision-monotonic install guard — left all 308 tests green. Tests added for stale-preview, Undo availability, and non-advancing-revision install. `invalidatePreview()` itself remains only partly isolable; see `deferred`.
  - `[medium]` `[patch]` Coverage gap proved by mutation: deleting the `documentGeneration`/`selectionKey` refusal drop, the generation-guarded busy reset, and the `clearDocumentInteraction` reset left all 308 green — shipping a chain command in flight across a document open would strand `fontChainBusy` true and kill the editor for the session. Two tests added on the existing asset-path precedent; the stale-refusal one needed the replacing document to declare a same-named chain, or the message would drop for the wrong reason and prove nothing.
  - `[medium]` `[patch]` The new refusal-vocabulary scan read RAW file text, so its patterns matched ordinary English in comments — a tax this diff had already paid, its only two unrelated hunks being prose rewrites in `App.tsx` and `sheet-stack.ts` made solely to escape it. Comments are now stripped by a character scanner (quotes tested before comment openers, so an apostrophe in a comment cannot open a string and `//` inside a URL cannot open a comment); both reworded comments were restored verbatim as the proof, and `sheet-stack.ts` consequently left the diff entirely. Five further refusal sentences Go actually emits were then added safely. Scan stays production-only; all twelve literals red-proved.
  - `[low]` `[patch]` `canvas-font-stack.test.ts`'s stylesheet-constant test was vacuous under a CSS reformat (`.find(...) ?? ''` then `exec('')?.[1] ?? ''` makes `expect('').not.toMatch(/var\(/)` pass proving nothing); both halves now assert they were found. The implementer notes the file also throws at module scope in that case, so this is defence in depth rather than a live hole.
  - `[low]` `[patch]` The DW-35 tripwire regex fires only when a chain reference sits on the same line as the font-family property, while its comment claimed any future wiring would go red; the comment was reworded to exactly what the scan proves, naming the non-intersection test as the stronger evidence.
  - `[low]` `[patch]` `applyProperties` did not clear `fontChainError`, so a refused chain edit stayed rendered while the author went on committing font size or bold. Cleared symmetrically, with a test.
  - `[low]` `[patch]` The add-chain and add-entry inputs were never cleared after an accepted add, so a second press re-dispatched — and since the engine has no duplicate-entry rule, silently appended a duplicate face. Fields are now emptied only when the add LANDED, asserted both ways (a refusal keeps the text so it can be corrected).
  - `[low]` `[patch]` The rename-refusal test's "document unchanged" assertion read the uncontrolled name fields, which show the author's draft and would look identical had the rename succeeded. Both such assertions now read the entry lists and the declared-fonts listbox.
  - `[low]` `[patch]` `.font-chain-name` and `.font-chain-face` were applied in JSX with zero rules in `App.css`; the ruleless class names were dropped.

Rejected (dropped): a lone-surrogate case for `compareCodePoints` (Go cannot emit a lone surrogate — `encoding/json` substitutes U+FFFD — so the divergence is unreachable); a `Number.isSafeInteger` guard in `font-chain-command.ts`'s `index()` (every call site passes an array index); the editor being reachable only while a component is selected (AC1 mandates that placement); the compile-only e2e spec exercising no refusal path (it executes nowhere by design); and "Add font chain" with empty fields dispatching (that IS the required anti-pre-emption behaviour).

**Two deviations recorded rather than repaired.**
1. Task 9 says the prohibition covers `src/**`; it is implemented over PRODUCTION source only. This is forced by the spec itself: the AC requiring the rendered refusal to be `===` to the engine's `message` obliges a test to hold that sentence as a fixture, so a literal `src/**` scan would redden on its own evidence. Production is where a rule would have to live to reach an author, so the AC's substance is served; the narrower scope is documented in the test.
2. Design Note N4 still narrates "the authority is created (**Story 15.2a's Task 1, not this story's** — re-pointed 2026-08-31)" and "Raise DW-32 to HIGH — **done 2026-08-31 by the orchestrator; DW-32 is Story 15.2a's and is NOT closed by this story**" — residue of the pre-amendment scope that amendment item 5's second half did not re-point at Story 15.2a. It is explanatory prose, not a task, AC, boundary or mutation proof, and nothing in the story is unsatisfiable without the consolidation, so the contract's halt trigger ("HALT if an AC seems to require any of it") was not met. Flagged for the orchestrator.

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
is D-8.1.3's exact shape, so the authority is created (**Story 15.2a's Task 1, not this story's** — re-pointed 2026-08-31) and that is a real change in this story's
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
  which is an accident, not a guard. Raise DW-32 to HIGH — **done 2026-08-31 by the orchestrator; DW-32 is Story 15.2a's and is NOT closed by this story**.

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
- Restore `quote()`'s five-character escape table and confirm the C0-control case goes red. **The
  departed population, re-asserted:** the five characters `quote()` already handled, plus lone surrogates,
  must still escape correctly after routing through `JSON.stringify` — the fix widens what is escaped and
  must not narrow it.
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
and Design Notes carry the measured investigation this gate was asked for) and `multiple-goals`, which was
**RESOLVED BY SPLITTING, 2026-08-31 (D-8.2.6)**. The gate identified three independently shippable goals:
the chain editor, the shared command-JSON authority closing DW-32, and the projection-guard ordering
closing DW-70. **DW-32 split out to Story 15.2a**, in Epic 15 before the tag, because its subject is
command encoding rather than font chains and it is reachable today without any of this story. **DW-70
stayed** as a precondition — this story is what makes it reachable, and shipping the editor without it
ships a feature whose second keystroke terminates the worker. A **minimal** `quote()` fix stayed with it,
on the same test, and **must not grow into the consolidation**.

Dispositions ruled at this gate, with grounds recorded in Design Notes: AC3's positive half is
unsatisfiable before Story 8.3 and is delivered as a forward-compatible display rule (N2); DW-35 does not
close here and its stated placement ground is falsified (N3); DW-32 closes here, understated in the
register and raised to HIGH (N4); DW-70 becomes fully reachable and closes here (N5 iv). No intent gap was
found.

---

### Auto Run Result — dispatch 2 (implementation), 2026-08-31

Status: blocked
Blocking condition: `intent contract contradicts itself on DW-32 scope; an acceptance criterion requires the excluded work`

Dispatched at HEAD `f4e88863071358a6f69f6e026b89cf1019fa3697` on `main`, tree clean, spec at
`ready-for-dev`, directive "do not halt after planning". **No implementation subagent was launched. No
file under `folio-designer/`, `folio-go/` or `lint/` was written, no code was produced, nothing was
committed.** The only file modified by this dispatch is this spec: its `status` frontmatter and this
block.

**Why the run stopped before step-03's implementation handoff.**

The `<intent-contract>` block both mandates and forbids the DW-32 consolidation, in three passages that
cannot all be honoured:

- **Line 29 (Intent -> Approach), mandates it:** "...introduce one shared command-JSON authority and
  route every designer command encoder through it..."
- **Line 50 (Boundaries & Constraints -> Always), mandates it:** "Every command payload is produced by
  one shared encoder that uses `JSON.stringify` for every value."
- **Lines 81-84 (Boundaries & Constraints -> SCOPE SETTLED 2026-08-31, D-8.2.4 ... D-8.2.7), forbids
  it:** "**DW-32 IS OUT OF SCOPE — it is Story 15.2a** ... Do **not** build the shared command-JSON
  authority, do **not** touch `rawNumberLiteral`, and do **not** consolidate the five encoders. **HALT if
  an AC seems to require any of it.**"

The halt is therefore the contract's own instruction rather than a builder's judgement call, and its
stated trigger condition is met. **Line 290 is an acceptance criterion that requires the excluded work:**

> "Given the shared encoder, when each of the six chain builders **and each existing encoder** is fed a
> value containing `"`, `\` and `U+0001`, then every produced payload is valid JSON — proved per builder."

Five further passages outside the contract are derived from the mandating half and are equally
unsatisfiable under the settled scope:

- **Task 1** (line 209) — create `folio-designer/src/command-json.ts`, "the **single command-JSON
  authority**".
- **Task 2** (lines 214-218) — "`rawNumberLiteral`'s verbatim splice is replaced by the validating
  encoder". The settled scope forbids touching `rawNumberLiteral` and limits this file to `quote()`.
- **Tasks 3-4** (lines 219-223) — route `page-setup-command.ts`, `component-command.ts`,
  `component-asset-command.ts` and `table-column-command.ts` through Task 1's module and "delete the two
  byte-identical hand-rolled copies". This is the consolidation of the five encoders, by name.
- **Task 15** (line 273) — mark **DW-32** CLOSED with the closing revision.
- **`## Verification` -> Mutation proofs** (lines 462, 464-467) — "Restore `rawNumberLiteral`'s verbatim
  splice and confirm the `a\"b\\c` and injection cases go red", and "The departed population,
  re-asserted. Task 1 moves escaping out of four local encoders into one." Neither mutation has a site to
  act on if the consolidation is not built.

**Why this could not be worked around inside the workflow.** Step-03 requires the implementation handoff
to be passed **verbatim** — "the spec is the subagent's sole source of truth" — and explicitly forbids
adding "parent-authored goal restatements, file lists, ownership boundaries, or acceptance criteria" to
it. There is consequently no sanctioned lever by which a subagent reading this spec would build the
minimal `quote()` fix instead of Task 1's authority: it would read Tasks 1-4 and build exactly the three
things D-8.2.4 ... D-8.2.7 forbid. Narrowing the scope silently would also leave step-04 reviewing an
implementation against an AC and two mutation proofs that implementation cannot satisfy, and would put a
story into history whose recorded acceptance does not match its code.

**What must change before re-dispatch.** All of it is the orchestrator's to amend; `<intent-contract>` is
read-only to this workflow.

> **AMENDED 2026-08-31 — all seven items below are DISCHARGED.** The contradiction was the
> orchestrator's: a scope block settling D-8.2.4…D-8.2.7 was appended to a contract whose Approach,
> Always bullet, one AC, Tasks 1–4, Task 15 and two mutation proofs had all been written under the
> opposite assumption. **The halt was correct and is the reason nothing was built against a
> self-contradicting contract.** This block is retained as the record of it rather than deleted.

1. **Line 29** — strike the shared-authority clause from the Approach, so it reads as the chain editor,
   DW-70, and the minimal `quote()` fix.
2. **Line 50** — restate the Always bullet minimally: the values the **chain-editor** commands encode are
   produced with `JSON.stringify`, and `quote()` in `component-property-command.ts` is routed through
   `JSON.stringify`. Drop "one shared encoder" and "every command payload".
3. **Line 290** — re-scope the AC to the six chain builders plus `quote()`, dropping "and each existing
   encoder".
4. **Tasks 1-4** — replace with one task: route `component-property-command.ts:41` `quote()` through
   `JSON.stringify`, one line, leaving `rawNumberLiteral` and the other four encoders untouched.
5. **Task 15** — DW-32 is **not** closed by this story; it is Story 15.2a. DW-70 still closes here. Design
   Note N4's "Raise DW-32 to HIGH — **done 2026-08-31 by the orchestrator; DW-32 is Story 15.2a's and is NOT closed by this story**" and its framing ("the authority is created (**Story 15.2a's Task 1, not this story's** — re-pointed 2026-08-31) and that
   is a real change in this story's size") should be re-pointed at 15.2a. The severity re-rating is
   evidence worth keeping wherever it lands.
6. **Mutation proofs** — drop the `rawNumberLiteral` restoration and the "departed population" proof.
   Replace with the control-character assertion on `quote()` (a C0 control other than the five it already
   handled) plus lone-surrogate and five-original-character re-assertions on `quote()` alone.
7. Reset `status` to `ready-for-dev`, or step-01 of the next dispatch halts on `blocked spec supplied`
   before implementation restarts.

Nothing else in the spec is implicated. The chain editor (Tasks 5, 7, 8, 9, 11, 13, 14), DW-70 (Tasks 6,
12), the 22-digest baseline, the anti-pre-emption assertions and the discriminating re-projection proof
are all coherent with the settled scope and were not disturbed by this halt.

---

### Auto Run Result — dispatch 3 (implementation), 2026-08-31

Status: done
Blocking condition: none

Dispatched at HEAD `3f9205790598831d47a44e080daaa8d4d5a3245a` on `main`, tree clean, spec at
`ready-for-dev` after the orchestrator's repair, directive "do not halt after planning". The
self-contradiction that stopped dispatch 2 is gone: Approach, the Always bullet, the AC, Tasks 1-4,
Task 15 and both mutation proofs now agree that DW-32 is Story 15.2a's. Implementation proceeded.

**Implemented change.** An inline font-chain editor in the typography panel that dispatches Story
8.1's six existing commands and re-projects, holding no document model, no validation and no copy of
any engine rule; the DW-70 projection-guard ordering fix; and a one-line `quote()` fix. **No file
under `folio-go/` or `lint/` is in the diff**, so the byte-identity claim is structural as well as
measured.

**Files changed.**
- `folio-designer/src/FontChainEditor.tsx` (new) — the inline editor: per chain a name control,
  rename and delete actions, an ordered entry list with per-entry move-earlier/later/remove and an
  add-entry control, plus add-chain. Reads `projection.fontChains` every render. Keyboard-operable
  reordering with focus following the moved entry.
- `folio-designer/src/font-chain-command.ts` (new) — six builders at the exact arities
  `componentFields` counts (4/4/3/5/5/4), every value through `JSON.stringify`.
- `folio-designer/src/font-chain-control.ts` (new) — anchor-by-originating-control types. A `.ts`,
  keeping the oxlint baseline at 4.
- `folio-designer/src/component-property-command.ts` — `quote()` routed through `JSON.stringify`.
  One line. `rawNumberLiteral` and the other four encoders untouched.
- `folio-designer/src/engine-protocol.ts` — DW-70: `compareCodePoints` (= Go's UTF-8 byte order)
  replaces `>=` on UTF-16 code units. **On the TypeScript side only; Go's comparator is normative.**
- `folio-designer/src/App.tsx` — `applyFontChain` on the same command path property commits use,
  with stale-generation/selection drop; refusal state keyed by originating control; the third button
  on `FontFamilyProperty`; the editor rendered inside the TYPOGRAPHY section.
- `folio-designer/src/canvas-authority-contract.test.ts` — the refusal-vocabulary prohibition over
  production source, comment-stripped, twelve literals, each red-proved.
- `folio-designer/src/engine-ownership-contract.test.ts` — `documentJson` allowlist widened to admit
  the two scalar encoders. A forced consequence of routing `quote()` through `JSON.stringify`.
- `folio-designer/src/{App.test.tsx,engine-protocol.test.ts,canvas-font-stack.test.ts,component-property-command.test.ts,font-chain-command.test.ts}`, `App.css`, `e2e/component-properties.spec.ts`.
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-70 CLOSED; DW-32 amended (its
  `quote()` half landed here, 15.2a must re-read the file); DW-35 amended; DW-73 and DW-74 filed.

**Review findings breakdown.** 12 patches applied (0 high, 6 medium, 6 low), 4 items deferred
(1 high, 2 medium, 1 low), 5 rejected, 0 intent_gap, 0 bad_spec. Two of the twelve were surfaced by
a reviewer's own executed mutation, not by inspection.

**Follow-up review recommendation: true.** Patched counts this pass: high 0, medium 6, low 6.
Score = 3x6 + 1x6 = 24, which is >= 5. No patched finding was high severity.

**Verification performed** (every command re-run by the parent after patching, not taken on report).
- `cd folio-go && go test -count=1 ./...` — exactly ONE distinct red,
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`, the mandated permanent red; 13 packages `ok`.
- `go vet -tags=matrix ./...` clean (exit 0); `gofmt -l folio-go` from the repo root: no output.
- `TestTargetRenderHash` five ways: `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` each
  exported and asserting with byte-identical hashes across all four, plus an **unset control** that
  emits no target lines and passes in 0.375s as the deliberate no-op.
- `TestCrossTargetByteIdentity` PASS (21.93s); `TestThaiStackedMarksSemanticSignOffIsRecorded` PASS.
- `cd lint && go test -count=1 ./...` — all four packages `ok`; `TestFloatTypedTestScopeInventory`'s
  line-number pins did not move.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` —
  typecheck and e2e-compile clean; oxlint **exactly 4** pre-existing `only-export-components`
  warnings and nothing else; **319 tests / 35 files**, all passing (baseline 285/34).
- `shasum -a 256 fixtures/*/expected.pdf | diff - <pre-edit baseline>` — **22 lines, empty diff**.
  Baseline captured before the first edit. `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`,
  8470 bytes, unchanged; both signoff files byte-identical; no `goldenDigestRecord` line added.

**Mutation proofs, run by the parent and each confirmed to have LANDED before any conclusion.**
`quote()` five-char table restored -> the C0 assertion reddens (`SyntaxError: Bad control character`).
DW-70 comparator reverted to UTF-16 -> the acceptance case reddens (`parseInbound` returns undefined,
the blank-canvas path); the ordering check deleted entirely -> the same test reddens on its
out-of-order half, so both directions are proved. A real refusal literal inserted into a real
production file -> the prohibition reddens naming the file and pattern; after the comment-stripping
patch, re-proved BOTH ways — the same literal is clean in a comment and red in a string. A local
duplicate-name rule added to `addChain` -> the anti-pre-emption assertion reddens. A local
rename pre-check added -> the rename anti-pre-emption assertion reddens (the first attempt produced
a syntax error and "no tests", which was discarded as a broken mutant and redone as valid code).

**Matrix Test Audit.** All 13 rows covered by tests that ran and passed. Two rows were uncovered on
the implementer's first pass and were filled before review: "Rename onto an existing key" had only a
dispatch test and no refusal test, and the "Reorder entries" out-of-range error cell was covered only
by a differently-worded unlocated refusal.

**Residual risks — what is NOT proven.**
`npm run test:e2e:compile` is `tsc --noEmit`. Playwright appears in no workflow and browser e2e has
never executed anywhere in this run (D-000.4). **This is a panel story, so the surface an author
actually touches is the one nothing executes.** Specifically unproven: that the editor is
keyboard-operable end to end in a real browser; that focus is visibly rendered per UX-DR25; that
`role="alert"` is actually announced; that the six commands survive the real worker round trip; that
the DW-70 fix keeps a real worker alive; and that the new affordance does not disturb
`e2e/browser-native-roundtrip.spec.ts:84`, the repository's only cross-boundary authoring witness,
which is compiled and never run. jsdom applies no stylesheet, so no unit test proves the editor's
appearance or focus ring. None of this may be reported as verified.
Also unproven: `invalidatePreview()`'s isolable half on the chain path (see `deferred`); and the
refusal sentences are coupled to Go by transcription, since no test reads both languages.

## Delivery Log

### 2026-08-31 — planned

Planned at baseline `bc671da` on a clean tree, halt-after-planning. The plan gate raised two warnings
and resolved the load-bearing one by **splitting**: the spec had three independently shippable goals,
and DW-32's shared command-JSON authority left for **Story 15.2a** in Epic 15, sequenced before the
tag (D-8.2.6). DW-70 **stayed** as a precondition — this story is what first lets an author type a
chain name — and a deliberately minimal `quote()` fix stayed with it. The gate also ruled AC3's
positive half unsatisfiable before Story 8.3 and delivered its negative half as a forward-compatible
display rule (D-8.2.5, D-8.2.7, with D-8.1.2/D-8.1.3 standing behind them).

### 2026-08-31 — built

**The first implementation dispatch halted, and the contradiction was the orchestrator's own.** The
scope block settling D-8.2.4…D-8.2.7 was **appended** to an intent contract whose Approach, one
Always bullet, one acceptance criterion, four tasks and two mutation proofs had all been written
under the opposite assumption — that this story *would* build the shared authority. The contract
therefore both mandated and forbade the same work, and its own stated trigger (*"HALT if an AC seems
to require any of it"*) was met. **The halt was correct, and the reason it was correct is
structural:** the implementation handoff passes the spec **verbatim**, as the subagent's sole source
of truth, so there was no sanctioned lever by which a builder reading that contract would have built
the minimal fix. It would have built exactly the three things the rulings forbid, and the review
layer would then have graded it against acceptance criteria its own code could not satisfy. Nothing
was written, nothing was committed; the orchestrator amended all seven items and re-dispatched. This
is the halt working as designed, not a builder being timid.

The third dispatch implemented against the repaired contract and committed at `e3ba0a2` (17 files,
+1513/-19). **No file under `folio-go/` or `lint/` is in the diff**, so the byte-identity claim is
structural as well as measured.

**DW-70 was a two-keystroke worker kill, not a cosmetic disagreement.** Go sorts the projected chain
names by byte; the browser's guard compared UTF-16 code units. They disagree wherever a name mixes
the astral planes with U+E000–U+FFFF, and the consequence was not a dropped frame — the guard failed,
the whole snapshot was discarded, and the engine worker was **terminated**, leaving the canvas
permanently blank. Reachability was total the moment this story shipped, because the engine accepts
any non-empty bounded name. **The fix had to go on the TypeScript side and must stay there:** those
keys are sorted into the canonical document's own `fonts` order under AD-9, so Go's comparator *is*
the document's byte order and is normative. Fixing the cheaper-looking side would have moved golden
bytes. The spec says so twice, and it was right to.

**`quote()` was the third premise about that one file to be falsified inside a single story.** The
dispatch asserted that chain names would travel through the numeric splice; measurement showed they
do not — they route through the quoter. The plan gate then cleared that route. And the cleared route
was **still broken one level down**: the quoter escaped five characters where JSON requires all of
U+0000–U+001F, so a pasted control character produced malformed bytes and the engine answered with a
generic parse failure instead of the located refusal that would have named the field. Engine-side
validation could not have substituted, because the bytes never reached the rule. Each premise was
narrower than the last and each was wrong; the file was only understood by being executed.

**The refusal-vocabulary prohibition initially matched ordinary English, and edited the codebase to
suit itself.** Its first version scanned raw file text, so its patterns — *already exists*, *is
declared*, *is out of range* — collided with prose that had nothing to do with an engine rule. Two
unrelated comments, in the panel and in the sheet stack, were reworded purely to get the guard green:
the guard editing the codebase rather than the codebase answering to the guard, and it would have
recurred on every future occurrence. It now strips comments with a character scanner that tests
quotes before comment openers. **Both reworded comments were restored verbatim as the fix's own
proof, and the sheet-stack file consequently left the diff entirely.** Five further refusals Go
actually emits were then added safely, taking the guard from seven literals to twelve.

Review: 12 patched (0 high, 6 medium, 6 low), 4 deferred, 5 rejected, 0 intent gaps, 0 bad spec.
Two of the twelve were surfaced by an executed mutation rather than by inspection.

### 2026-08-31 — done

Closed at `e3ba0a2` on `main`, baseline `3f92057`. Every gate below was **re-run at close**, not taken
on the build's report, and every number is the closer's own measurement.

**Measured green.** `go test -count=1 ./...` — 13 packages `ok`, exactly **one** distinct red,
`TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`, the mandated permanent red; since this story
changes no Go file, that is the whole Go story. `go vet -tags=matrix ./...` exit 0; `gofmt -l folio-go`
from the repo root, no output. `TestTargetRenderHash` on **four exported legs** — `darwin/arm64`
(0.73s), `linux/amd64` (7.33s), `linux/arm64` (4.66s), `js/wasm` (10.58s) — each emitting 23 real
per-document hashes, **plus the unset control**, which passes in 0.375s while printing that it asserts
nothing. The control is what proves the four legs were not the same no-op.
`TestCrossTargetByteIdentity` ok (22.2s); `TestThaiStackedMarksSemanticSignOffIsRecorded` ok.
`cd lint && go test -count=1 ./...` — four packages `ok`, the mandatory `-count=1` per D-7.9.5, and
`TestFloatTypedTestScopeInventory`'s line-number pins did not move. The designer gate — typecheck,
oxlint, unit, e2e-compile — all four exit 0: **oxlint exactly 4** pre-existing `only-export-components`
warnings and nothing else, and **319 tests / 35 files** passing against a 285/34 baseline. All **22**
golden digests diffed against a baseline reconstructed from the pre-story tree: **empty**.

**Re-derived, not relayed.** The two anti-pre-emption locks were reproduced by adding real local rules:
a duplicate-name pre-check and an emptying-remove pre-check each turn the dispatch assertion red with
*"expected [] to deeply equal [Array(1)]"* — nothing sent. That is the assertion that matters, because
a passing *"the error shows"* test is satisfied equally well by a TypeScript copy of the rule. The
prohibition was proved **both ways with the same literal**: clean in a comment, red in a string, naming
the file and both patterns — so the comment-stripping fix is measured, not asserted. DW-70 was proved
in **both directions**: reverting the comparator to UTF-16 reddens the accepted private-use/astral pair (U+E000 before U+1F600),
and deleting the ordering check outright reddens the out-of-order half. **No Go comparator moved** —
zero `folio-go/` paths in the diff, mechanically confirmed. `quote()`'s C0 assertion reddens under the
restored five-character table (*Bad control character*), and the departed population — those five
characters plus a lone surrogate — is re-asserted in the same test, so the fix widens and does not
narrow. Both Matrix Test Audit cells filled before review — rename onto an existing key, and the
out-of-range reorder — carry real refusal assertions and are **non-vacuous**: making the panel invent
its own refusal text reddens all six refusal tests. Scope was verified mechanically: zero paths under
`folio-go/` or `lint/`, no `command-json.ts`, the four other encoders untouched, DW-32 not closed, and
`rawNumberLiteral` and `role="dialog"` each appearing in the diff **only inside a comment**.

**Not measured, and not to be read as verified.** Browser e2e never executed — `test:e2e:compile` is a
type-check only, and Playwright appears in no workflow (D-000.4). **This is a panel story, so the
surface an author actually touches is the one nothing runs.** Unproven: keyboard operability end to
end in a real browser, the visible focus ring, whether the alert is announced, whether the six commands
survive the real worker round trip, and whether the DW-70 fix keeps a real worker alive. jsdom applies
no stylesheet, so no unit test can reach appearance or focus either. The full `-tags=matrix ./...`
sweep was **not** run, so `TestShippedFacesReproduceFromUpstream` is unmeasured here and comes due at
Epic 8's close; `lint/…/licencegraph_test.go`'s gofmt deviation remains DW-23 and is outside this path.

**Deferrals filed at close.** All four of the build's deferrals were recorded only in the spec's
frontmatter and are now in the register: **DW-75 (HIGH)**, **DW-76**, **DW-77**, **DW-78** — plus
**DW-79**, found by the closer's own mutation screen. DW-75 is the one that matters: the two
hand-rolled escapers iterate by code point but escape from the first UTF-16 unit, so an astral
character is emitted as a **lone surrogate**, Go substitutes U+FFFD, and a bind segment or asset key
**binds to a different path than the author typed**, silently. Re-confirmed at close **by executing
the function**, not by reading it. It is Story 15.2a's, and DW-32's owner line now points at it so that
story knows its consolidation is a repair rather than a tidy-up. DW-79 records that the value-signature
fix the review pass made is correct and load-bearing (forcing it false reddens 4 tests, true reddens 1)
but that **no test distinguishes it from the array-identity check it replaced** — reverting that
specific defect would go unnoticed by all 319.

**`followup_review_recommended` cleared to `false`.** The flag was raised on patch volume (12 patches,
score 24), with **no high-severity finding among them**. Every load-bearing claim was independently
re-derived at close, as recorded above; the one new gap found is DW-79, filed at LOW. Nothing was
re-opened.
