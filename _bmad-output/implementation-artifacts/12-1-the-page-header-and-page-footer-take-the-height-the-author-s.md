---
title: 'Story 12.1: The page header and page footer take the height the author sets'
type: 'feature'
created: '2026-09-05'
status: 'done'
review_loop_iteration: 1
baseline_commit: 'fd4da07c569ce437458037f1478a6b0fc114cbb2'
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** `Band.Height` has no writer anywhere in the product. The loader accepts it, the renderer
and the canvas both consume it, and no command in either door sets it — so a letterhead's height is
whatever the starter file declared, and changing it means hand-editing the file the designer just saved.

**Approach:** Add one `setBandHeight` arm to `ApplyComponentCommand`, carrying a band name and a height.
It validates before it mutates, refusing with a located message when the height would leave no content
window or would strand a component in a vertically capping band, and returns a fresh projection. Two
rows in the PAGE SETUP panel read the projected heights and send that command. The panel enforces no
bound of its own.

## Boundaries & Constraints

**Always:**
- The refusal names the **act**, not the object: it names the height that was refused and the element
  that would be stranded. Reuse `containComponent` as the **predicate**; do not reuse its sentence
  (*"component geometry must stay within pageHeader"* answers a command the author did not send).
- The strand predicate evaluates **every** component in the band, not a selected one.
- The content-window invariant has **one** implementation, called by both `Canvas` and the new command.
  Each caller phrases its own refusal; neither re-derives the arithmetic. Two call sites of one
  function, never two implementations.
- Only `pageHeader` and `pageFooter` are settable — the members of `bandsCappingVertically`. `content`
  is refused; its height is derived.
- The command is atomic: on refusal the document is byte-unchanged.
- A band-height refusal reaches the author through `componentDiagnostic`, **never** `pageSetupDiagnostic`
  — the latter discards the engine's message for anything not carrying `PAGE_SETUP_INVALID`.
- The panel sends a band-height command only when that field's value **differs from the projected
  value** — it avoids needless commands and history entries. (Amended 2026-09-05: the original clause
  justified this by keeping an already-stranded hand-edited document editable. That document cannot be
  opened in the designer at all — see the Spec Change Log. The rule stands on the honest ground.)

**Ask First:**
- Any new `diag` registry code. The refusals here are `fmt.Errorf`/`componentFailure` strings; AD-14's
  permanent-surface cost must stay unpaid.
- Any change to `ApplyPageSetupCommand`'s arity, to `engineFailure`'s prefix routing, or to
  `pageSetupDiagnostic` — all three serve the whole page-setup panel.
- Any change to `containComponent`, `containEdgeY` or `bandsCappingVertically`, whose source text is
  pinned by `engine-bounds-mirror.test.ts:169,215-216`.

**Never:**
- **No browser-side floor, clamp or bound on the band-height rows.** Story 17.4 item 9 ruled that a
  band-edge bound is a property of the LAYOUT and must not be mirrored into the inspector, *whether or
  not the engine holds a bound*. The panel sends the value, the engine refuses it, the existing
  `role="alert"` path renders the refusal — consistency with typing is the property asserted.
- No floor on the `content` band.
- No second implementation of the content-window arithmetic, and no browser copy of it.
- No new projection field. `bands[].height` is already projected and already in `hasOnly`.
- No measurement of anything in the browser (AD-17).
- No change to `serialize.go` — AC5's behaviour already ships and only becomes testable here.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Sets a header height | A4 portrait, 36pt margins, `pageHeader` 20pt, no elements in it; `{band:"pageHeader",height:80}` | Accepted; `bands[0].height` = 80000; serialized `pageHeader.height` = 80 | N/A |
| Shortens to the lowest occupied edge | `pageHeader` holds `e1` at `y=50 h=30`; height 80 | **Accepted** — the boundary case, `y+h == height` | N/A |
| Strands a component | Same band; height 79 | Refused, document byte-unchanged | Located: ElementID `e1`, DataPath `bands.pageHeader.height`, message naming 79 and what would be stranded |
| Strands a component that is not the first | Band holds `e1` at `y=0 h=10` and `e2` at `y=50 h=30`; height 40 | Refused, naming `e2` | As above — the vacuity guard |
| Leaves no content window | Header 400, footer 400 on A4 (inner 769.89pt) | Refused, document byte-unchanged | Located, naming the quantity and the space available |
| Header plus footer exactly equal the inner height | Sums to exactly `innerH` | Refused — content must be strictly positive | As above |
| Content band | `{band:"content",height:100}` | Refused; `content` never gains a `height` key | Located, DataPath `bands.content.height` |
| Negative height | `{band:"pageFooter",height:-5}` | Refused | Located |
| Unknown band | `{band:"footer",height:20}` | Refused | Located |
| Wrong arity / duplicate keys | 3 or 5 keys; a repeated `band` | Refused by the door's existing gates | `componentFields` / `refuseDuplicateCommandKeys` |
| Re-sends the current height | Value equal to the projected one | Byte-identical document; no history entry (`wasm/engine.go:241-244`) | N/A |
| Nothing edited | Any document, no band-height command | Serializes byte-identically; an absent key stays absent | N/A |
| Refusal reaches the author | Any refusal above, from the panel | The engine's own located text renders via `role="alert"` | Not the fixed page-setup sentence |

</frozen-after-approval>

## Rulings Carried In

Settled by the engineering lead on 2026-09-05 in answer to this spec's Open Questions, which were raised
after five premises in D-12.A and the build dispatch were falsified at `fd4da07`. Recorded verbatim in
substance so a later reader does not re-open them.

- **Q1 — a new `setBandHeight` arm on `ApplyComponentCommand`.** *"The split is structural, not
  arbitrary: the command door's own validations are located; the projection's invariant checks are
  not."* `ApplyPageSetupCommand` gates on `len(raw) != 7`, so extending it changes its arity and every
  caller's shape; a new arm does not. The seven font-chain arms are the shipped precedent for a
  document-level mutation at that door.
- **Q2 — the engine refuses, and D-12.A's mechanism claim is withdrawn.** `containComponent` is **not**
  on the band-height path. Add the check to the writer, reusing `containComponent` as the predicate over
  a candidate band; one new call site, `containComponent` unchanged, no new diag code. The refusal's
  **subject is the act, not the object**. 12.1 and 12.5 call one predicate — *"my original 'one
  derivation, consumed twice' survives — relocated from the browser to the engine."*
- **Q3 — the command validates and refuses located; `Canvas`'s bare refusal stays as the hand-edit
  backstop.** *"The 'narrow the message, not add the check' principle does not bind here because these
  are two guards with two audiences: the command guards AUTHORING and owes a located message; `Canvas`
  guards LOADING and owes a refusal to project a document that cannot be laid out."* Condition: both
  call one predicate.
- **Q4 — REVERSED. The panel does not floor.** D-12.A's Fork-1 panel half is withdrawn, and the lead's
  own "a Go bound would make the floor legal" framing with it. Story 17.4 item 9's distinction is
  **field-property vs. layout-property**, not mirrored-vs-only-copy: a floor at `max(y + height)` over a
  band's components is layout-dependent by that definition's own terms, and item 9 forbids mirroring it
  *whether or not the engine holds a bound*. **This spec makes no claim to settle 17.4's band-extent
  question for any axis — it was already settled, the other way.** The D-7.4.5 mirror obligation and its
  red-proof are struck with it.
- **Q5 — DW-143 stays OPEN, owner unchanged.** The dispatch's re-pricing instruction is struck.
  Off-*page* geometry and a component rendering on the page overlapping content are different defects.

## Code Map

Measured at `fd4da07`. Every count marked CLOSED is fully enumerated.

**Go — the write path**
- `folio-go/internal/template/model.go:237-251` -- `Bands{Content, PageFooter, PageHeader}` (no `Extra`;
  the three-band set is closed) and `Band{Elements, Height Presence[geom.Length], Extra}`.
- `folio-go/internal/template/presence.go:16-38` -- `Presence[T]{Set, Null, Value}`; `present`/`presentNull`
  are **unexported**, so only package `template` can build one. Outside it, the exported composite literal
  is the only route.
- `folio-go/internal/template/parse_bands.go:85-100` -- the **only non-test writer** of `Band.Height`
  (`:108`). Required on pageHeader/pageFooter, forbidden on content, refuses `null`, decodes points→
  millipoints via `decimal.go:30-67` (`big.Int` only, ≤3 decimal places). **No positivity check and no
  upper bound**: `"height": -5000` and `"height": 999999999` both parse cleanly.
- `folio-go/internal/template/serialize.go:269-278` -- `writeBand`; emits `height` on `band.Height.Set`
  **alone**, ignoring `.Null` (a directly-constructed `{Set:true, Null:true}` would serialize `height: 0`).
  Key order within a band: `elements`, `height`, `Extra`. This is AC5's whole implementation and it is
  already correct; AC5 becomes a real test only once a writer exists.
- Readers of `Band.Height` — **CLOSED, 3, all non-test**: `page_setup.go:698-699` (reads `.Value` **without**
  checking `.Set`/`.Null`), `render.go:305-309` (**does** check), `serialize.go:273-274`. Zero test readers.
  Writers outside `parse_bands.go`: **zero** (`omitempty_test.go:138-139` is in-package test-only).

**Go — validation and projection**
- `folio-go/page_setup.go:726-782` -- `Canvas`. Three refusals at `:735-747`, all bare `fmt.Errorf`, none
  carrying ElementID or DataPath. The band gate is `:743-746`: `header < 0 || footer < 0 || header >= innerH-footer`
  (content strictly positive; `header+footer == innerH` is refused). Bands built at `:765-769`; the content
  height comes from the one derivation `layout.ContentHeight` (`internal/layout/band.go:76`).
- `folio-go/page_setup.go:685-701` -- `canvasPageGeometry`, the sole entry of band heights into the canvas.
- `folio-go/page_setup.go:1643+` -- `canvasComponents`. Bounds only `MaxCanvasMillipoints` and string
  lengths; **never** compares a component's extent to its band's height.
- `folio-go/component_commands.go:2067` -- `var bandsCappingVertically = []string{bandPageHeader, bandPageFooter}`
  (a slice). Consulted at exactly 2 sites: `:2092` (`containComponent`) and `:2112` (`containEdgeY`).
  Mirrored in `folio-designer/src/engine-protocol.ts`; `engine-bounds-mirror.test.ts:169` parses the Go
  source with an anchored regexp and `:215-216` pins both gates' source text.
- `folio-go/component_commands.go:2069-2098` -- `containComponent` and the *"ONE band-extent validation in
  the designer command path"* comment. One message for every violation:
  `"folio: component geometry must stay within %s"`. **11 call sites** (the comment says "eleven"), all
  component commands — D-12.A's "10" is wrong.
- `folio-go/internal/diag/diag.go` -- **17** `Code =` declarations, not 19. CLOSED and corroborated four
  ways (`allCodes` 17, `dispositions` 17, `diag_test.go` `codePins` 17, `diag_bridge_test.go` 17). **No
  band-overflow code** — the substance of D-12.A holds; only the number was wrong. Adding one costs 4
  coordinated edits plus a real production trigger in `diagnostic_registry_census_test.go`.

**Go — the doors**
- `folio-go/page_setup.go:1903-2007` -- `ApplyPageSetupCommand`. `refuseDuplicateCommandKeys(command, pageSetupCommandPath)`
  at `:1914`; **`len(raw) != 7`** at `:1922`; `len(margins) != 4` at `:1944-1948`. Mutates `t.doc.Page` in
  place with a `restorePage` rollback (`:2008`).
- `folio-go/component_commands.go:202-276` -- `ApplyComponentCommand`; **24** `case` arms (CLOSED) + `default`.
  `componentFields(raw, N)` (`:1481`) is the per-arm arity gate.
- `folio-go/component_commands.go:2286-2320` -- `addFontChain`, the document-level precedent: no element id,
  writes `t.doc.Fonts`, refuses `componentFailure("", fontChainPath(name), …)`.
- `folio-go/wasm/engine.go:220-233` -- the **sole** dispatch, a binary `if/else` on the literal `"pageSetup"`.
  `:241-244` short-circuits when canonical bytes are unchanged (so re-applying the current height is not a
  committed mutation and does not touch undo) — relevant to AC5.
- `folio-go/wasm/cmd/engine/main.go:229-269` -- `engineFailure`. `*ComponentCommandError` matched **first**
  → `COMPONENT_INVALID` with ElementID/DataPath; page-setup fallback at `:262-268` reconstructs DataPath by
  substring match and needs prefix `folio: page.`.

**Designer**
- `folio-designer/src/App.tsx:1606-1607` -- the `PageSetup` component, rendered at `:1478` when nothing is
  selected. `<Field label="Top margin (pt)" …>` is the row to copy.
- `folio-designer/src/App.tsx:2888-2895` -- `type Draft`, `points()`, `draftFor(canvas)`. `App.tsx:2889` --
  `Field`, a plain `<label><input aria-label inputMode="decimal">`, raw string passthrough, no clamp.
- `folio-designer/src/App.tsx:1228` -- `updateDraft`; `:738-748` -- `applyPageSetup`; `:2908` --
  `pageSetupDiagnostic` (**discards the engine message unless `PAGE_SETUP_INVALID`**).
- `folio-designer/src/page-setup-command.ts:13-26` -- `pageSetupCommand`, the 7-key encoder.
- `folio-designer/src/component-property-command.ts:32-44` -- `POSITIVE_LENGTH_FIELDS` and the comment Q4
  corrects. Its **one** runtime read is `App.tsx:2150`, inside `step` (`:2155-2201`).
- `folio-designer/src/engine-protocol.ts:178` -- the band projection (an inline member of `CanvasProjection`,
  **no named `CanvasBand` type in TS**); `:179` -- components, carrying `band`, `x`, `y`, `width`, `height`
  as required non-negative safe integers. `:316` -- `hasOnly(band, ['name','x','y','width','height'])`;
  `:275` -- the projection's own `hasOnly`, pinned from Go by `canvas_projection_wire_test.go:346`.
- A floor **is** computable from the projection alone — and is deliberately **not** built (Q4). Recorded
  only so a later reader knows the omission was ruled, not overlooked.
- `bandHeight|headerHeight|footerHeight` across `folio-designer/src/` -- **zero**, CLOSED over 109
  `.ts/.tsx/.css` files. Positive control `borderWidth` = **6** files (D-12.A's "7" counts the compiled
  `src/generated/runtime/*.wasm` blob under `grep -a`). Band height reaches the browser only as
  `bands[i].height`.

**Source-text contracts that will fire (re-derived, D-12.A superseded)**
- `canvas-authority-contract.test.ts` **strips comments** (`withoutComments`, `:131-151`, quote-aware) before
  scanning its 14 prohibited patterns, and asserts that property at `:228`. **D-12.A.1's "scans raw text
  WITHOUT stripping comments" is false**; a comment about measuring a band is safe. Two real raw-text
  exceptions: `:535-547` `withoutApprovedLocalPointerInput` runs on raw App.tsx and **throws** unless
  `placementPoint` is immediately followed by `function pageStyle`; and `:238-241` App.css may contain
  **exactly one** `@media` query.
- `design-contract.test.ts` — **this** is where the hex/rgb/hsl rule (`:87`, whole-file, comments included)
  and the exact-once `var(--type-display)` / `var(--type-numeric-lg)` counts (`:145-151`) live. It reads
  App.css and **never** App.tsx (grep rc=1; positive control `App.css` = 3).
- `property-prose-height.test.ts:143-148` — exactly **one** `<input>` and one `<textarea>` carrying
  `property-value-prose` in App.tsx. A new page-setup row must not use that class. `:118-125` —
  `.property-field` must carry **no** `position:`.
- `page-setup-command.test.ts:36` — `Object.keys(command)` equals the 7 keys **in order**, plus two
  byte-exact payload strings at `:13-16`. Fires only on Q1(a).
- `command-json-soleness.test.ts:87` — allowlist: any production module building command JSON itself is
  flagged; a new factory using `commandBytes` from `command-json.ts` passes. `:94` names six factories
  by name and requires each to import the authority — a seventh should be added there to be covered by
  the `\bString\(` check at `:114`.
- `canvas-authority-contract.test.ts:431-435` — population floors: `e2e` 15, `.tsx` 8, `.css` 3 sit **exactly
  on** their floors. Adding a `.ts` file is safe; deleting any file of those kinds is not.
- `engine-ownership-contract.test.ts:11,24` — any TS type carrying ≥2 of
  `{version, page, bands, elements, assets}` is reported as a schema mirror and must be empty.

**Read-only / do not touch**
- `folio-designer/e2e/browser-native-roundtrip.spec.ts` — modified in the tree by the orchestrator's
  in-flight Playwright run. Not churn, not mine.
- `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child — the mandated permanent red.

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/page_setup.go` -- extract the content-window invariant at `:743-746` into one named
  predicate that reports whether a (header, footer, innerH) triple leaves a positive content region, and
  have `Canvas` call it. `Canvas`'s message is unchanged. -- Q3's condition: one implementation, two
  audiences.
- [x] `folio-go/component_commands.go` -- add a `setBandHeight` arm (25th) to `ApplyComponentCommand`'s
  switch and its handler: `componentFields(raw, 4)` over `kind`, `version`, `band`, `height`; `band`
  restricted to `bandsCappingVertically`; `height` via the existing length helper. Validate the
  content-window predicate against the proposed height and the *other* band's current height, then run
  `containComponent` over **every** element of that band against a candidate
  `CanvasBand{Name, Width: innerW, Height: proposed}`. Only then write
  `t.doc.Bands.<band>.Height` and return `Canvas(t)`. Each refusal is a `componentFailure` carrying
  DataPath `bands.<band>.height` and, for a strand, the stranded element's ID. -- The one writer; its
  refusals name the height and the element, never `containComponent`'s own sentence.
- [x] `folio-go/band_height_command_test.go` -- new: cover every Matrix row, including the
  `y+h == height` boundary acceptance and the **not-the-first-component** strand. Assert the located
  fields (ElementID, DataPath) and byte-identity of the document after each refusal. -- The Matrix's
  unit tests.
- [x] `folio-go/band_height_command_test.go` -- add the red-proofs, each mutating the **subject** and
  echoing the mutated line: delete the strand check; make it examine only the band's first element;
  delete the command's content-window call; change the shared predicate and show that both `Canvas`'s
  test and the command's test red together. -- D-000.14, and the vacuity the lead asked to be checked.
- [x] `folio-designer/src/band-height-command.ts` -- new factory built on `commandBytes` from
  `command-json.ts`, emitting `kind`, `version`, `band`, `height` in that order. -- The encoder;
  soleness requires it route through the authority.
- [x] `folio-designer/src/band-height-command.test.ts` -- new: byte-exact payload and exact key order,
  matching the sibling factories' test shape. -- Pins the wire.
- [x] `folio-designer/src/command-json-soleness.test.ts` -- add `band-height-command.ts` to the named
  factory list at `:94`. -- Otherwise the new encoder is outside the `\bString(` check; the list exists
  so "the scan found nothing" cannot mean "the scan never looked at it".
- [x] `folio-designer/src/App.tsx` -- add `Page header height (pt)` and `Page footer height (pt)` rows to
  `PageSetup` beside the margins, seeded from `canvas.bands[].height` through `draftFor`; extend `Draft`
  with the two keys. On **Apply page setup**, send a band-height command for each row **whose value
  differs from the projected value**, before the `pageSetup` command, stopping at the first refusal;
  surface a refusal with `componentDiagnostic`, not `pageSetupDiagnostic`. No clamp, no floor, no
  stepping. -- AC1 and AC2; the difference test avoids needless commands and history entries.
- [x] `folio-designer/src/App.test.tsx` -- assert each row shows the engine's value; that the header row
  sends `pageHeader` and the footer row sends `pageFooter` (the key→band map pinned, not just "a command
  was sent"); that a refusal renders the engine's own located text and **not** the fixed page-setup
  sentence; and that an unchanged row sends nothing. -- The Story 12.4 lesson: assert what would have to
  change for this to fail.

**Acceptance Criteria:**
- Given a document whose band heights are never edited, when it is opened and serialized, then the bytes
  are identical and `content` still carries no `height` key.
- Given a band-height row the author did not change, when Apply is pressed, then no band-height command
  is sent for it.
- Given a band-height refusal, when it reaches the panel, then the author reads the engine's own message
  naming the height and the element, through the existing `role="alert"` path.
- Given the designer at this commit, when `folio-designer/src/` is searched for a browser-side band-height
  bound, then there is none — the panel proposes values the engine may refuse, exactly as typing already does.

## Spec Change Log

**2026-09-05 — review pass 1. `review_loop_iteration` 0 -> 1. The frozen block was amended by explicit
human authorisation; the implementation was NOT reverted, by the same ruling.**

**1. THE TRIGGERING FINDING: a frozen justification rested on a false measurement, and the measurement
was mine.** At the plan gate I reported that a projection carrying a stranded component is admitted by
the browser — *"the projection is well-formed and the TS guard admits it; the snapshot is NOT dropped"* —
and on that basis D-12.A's original claim that the TS mirror drops the snapshot was recorded as
corrected. **It was right and I was wrong.** `folio-designer/src/engine-protocol.ts:345`:

    if (BANDS_CAPPING_VERTICALLY.includes(component.band as string) && !(box.y + box.height <= band.height)) return false

For the strand actually measured (`y=50000`, `h=30000`, `band.height=20000`), `80000 <= 20000` is false,
so `isCanvas` returns false, `parseInbound` returns `undefined`, and `EngineClient.#fail('PROTOCOL_INVALID')`
**terminates the worker**. How the error was made, because the mechanism outlives the fact: the claim came
from a surveying subagent's summary of the numeric-range check **fourteen lines earlier**, and the vertical
clause was never read. **A surveyor's summary of a guard is not the guard**, and the tell was fourteen
lines away.

**2. WHAT WAS AMENDED.** (a) The frozen Always clause justifying the send-only-if-different rule by
*"so an already-stranded hand-edited document can still accept a margin change"* — struck; that document
cannot be opened in the designer at all. (b) The acceptance criterion asserting the same unreachable
scenario — struck, replaced by the honest property (an unchanged row sends no command). (c) The rule
itself — **kept**, on the ground that it avoids needless commands and history entries. A rule can be
right for a reason other than the one first given.

**3. THE ENGINE CHECK'S REASON IS REPLACED, AND IS MUCH STRONGER THAN WHAT IT REPLACED.** Without
`setBandHeight`'s check, a band-height command creates a strand and the very next projection fails
`isCanvas`, terminating the worker **mid-session, with no attributable error**. *"Later commands are
refused"* was an inconvenience; this is data loss shaped like a crash. Recorded in Design Notes.

**4. KNOWN-BAD STATE AVOIDED.** Shipping a spec whose frozen contract justified a real rule with an
unreachable scenario, and an acceptance criterion whose only test was green because it mocks the engine
and never runs the real guard (`App.test.tsx` carries **0** references to `isCanvas`/`parseInbound`
against 305 to `engine`). The next reader would have inherited both as settled.

**KEEP — what worked and must survive re-derivation.**
- **The pageFooter rotation red-proof, which outranks everything else here.** Deleting the band-pointer
  swap so a `pageFooter` command writes `PageHeader` left the **entire Go suite green** except the
  mandated P6g red. Every accepting, strand, boundary and content-window test used `pageHeader`; the one
  `pageFooter` row returns above the swap. Both bands must be mirrored, and the rotation must stay as the
  red-proof.
- The shared-predicate red-proof: one edit to `bandsLeaveContentWindow` reddens `Canvas`'s test and the
  command's test **together**. That is the only evidence distinguishing one predicate from two
  implementations that currently agree, and it is the condition Q3 imposed, made checkable.
- The vacuity red-proof as a **pair**: restricting the strand loop to the band's first element must red
  the not-the-first test and nothing else.
- Validating before mutating, holding the previous `Presence` and restoring it on a failed projection.
- Refusing the content band inside the command — that is what keeps `height` absent from it forever.
- Naming the `Draft` keys for the bands themselves, so there is no row-to-band map that can be rotated.
- Routing band-height refusals through `componentDiagnostic`, never `pageSetupDiagnostic`.


## Design Notes

**Baseline measured at `fd4da07`**, not inherited: `cd folio-go && go test -count=1 ./...` -> **1956 pass /
2 fail / 5 skip**, exit 1, and the two failures are exactly `TestCorpusMeetsP6ExerciseFloors` and its
`P6g_(opaque_names)` child. Designer: **58 files / 806 tests**, exit 0; `npx tsc --noEmit` clean; `npx oxlint`
exactly 4 `only-export-components` (`preview/pdf-viewer.tsx:16,17`; `App.tsx:2896,2903`).

**Why the panel enforces nothing, written down so it is not "fixed" later.** The obvious kindness here is
a floor on the band-height field at the lowest occupied edge. Story 17.4 item 9 ruled against exactly that
shape, and the ruling's reason is the part that binds: a bound the browser can only compute from *layout*
— a component's position, its band's height, the page — does not belong in the inspector, however
convenient. A `max(y + height)` floor is layout-dependent by that test. The author therefore types a
number the engine may refuse, and reads why. That is what typing already does everywhere else in this
panel, and consistency with typing is the property the tests assert.

**Why the refusal cannot borrow `containComponent`'s sentence.** The predicate is right; the words are
answers to a different question. *"folio: component geometry must stay within pageHeader"* tells an author
who just set a header height that they moved a component, which they did not. The predicate is reused, the
sentence is new, and that costs nothing under AD-14 because command refusals are error strings routed by
prefix, not registry codes.

**Why a row is sent only when it changed.** Re-sending an unchanged band height is not free: it is a
command, a round trip and — if its bytes moved — a history entry, for a value the author did not touch.
The comparison is two strings, both the engine's own spelling of its own numbers; it computes nothing
about layout and is not a bound. **The original justification for this rule was different and was
false** — it claimed the rule kept an already-stranded hand-edited document editable. Such a document
cannot be opened at all (see below), so that scenario is unreachable. The rule survived the correction
on the ground above; see the Spec Change Log.

**A strand terminates the worker, and that is the real reason the engine must refuse one.**
`engine-protocol.ts:345` reads `if (BANDS_CAPPING_VERTICALLY.includes(component.band) && !(box.y +
box.height <= band.height)) return false`. So a projection carrying a stranded component fails
`isCanvas`, `parseInbound` returns `undefined`, and `EngineClient.#fail('PROTOCOL_INVALID')` calls
`worker.terminate()`. Without `setBandHeight`'s check, a band-height command would create that strand
and **the very next projection would brick the editor mid-session, with no attributable error** — not
merely leave the element awkward to edit. Refusing at the command door is what keeps that unreachable.

**Ordering on Apply.** Band heights are sent before `pageSetup` and the sequence stops at the first
refusal, so the common failure leaves the document wholly unchanged. The residue is disclosed rather than
designed away: if a band height is accepted and the `pageSetup` command is then refused, the band height
stands. Each command is individually atomic, which is what AC3 asks; the gesture is not.

**AC5 reads oddly and this is its honest form.** `parse_bands.go` makes `height` **required** on
`pageHeader`/`pageFooter` and **forbidden** on `content`, so the only band whose key can be absent is
`content`, and what keeps it absent is that `setBandHeight` refuses that band. AC5 is therefore tested as
byte-identity on an unedited document plus the content-band refusal, not as an optional-key emission test.

## Operating Constraints

Binding on every agent that touches this story, including subagents.

**Version control — you may NOT `commit`, `add`, `stash`, `checkout`, `reset`, `revert`, or `restore`.**
Every commit on this story is made by a human. Never push. Never create a branch. Leave your work
unstaged in the working tree and report what you changed. `folio-designer/e2e/browser-native-roundtrip.spec.ts`
is modified by a measurement running outside this story: do not touch it, do not run Playwright, and do
not report it as churn.

**Measurement rules — this repository produces false zeros by four unrelated mechanisms.**
- Recursive `grep` misses files here. Use `/usr/bin/grep -a` by absolute path, or `git grep --untracked`
  (plain `git grep` is tracked-only — D-000.15).
- **Capture exit status on the command's own line** — `cmd; rc=$?`. `$?` is clobbered by ANY intervening
  command, including `echo` (D-000.18). Prefer `if cmd; then … fi`, which cannot be clobbered.
- State the population beside every zero, with a positive control that actually fires. A control that
  fires for the wrong reason licenses a zero it never tested: `grep -a` matches binaries, so a compiled
  blob in the population makes any text count a sample rather than a census.
- **Red-prove by mutating the SUBJECT, not the expectation** (D-000.14), and echo the mutated line back.
- `folio-go/internal/arch_test.go`'s `TestNoFloat64UnderModule` scans `_test.go` files too — **no
  `float64` anywhere in that module, tests included.** Use `json.Number` for numeric read-back.
- Ask of every assertion: **what would have to change for this to fail?** A table test asserting
  "accepted, and some byte moved" stayed green when a key→edge map was rotated so every command wrote
  the opposite edge.
- Run the `lint` module with `-count=1` always: its rules walk a directory, and Go's test cache does not
  track `ReadDir`, so a cached `ok` there is no measurement at all.

## Verification

Cadence is **end of epic** — no matrix suite, no Playwright. Exit code captured on each command's own line.

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expect exit 1 with **exactly two** failures,
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child. A third is a real regression.
- `cd folio-designer && npx vitest run` -- baseline 58 files / 806 tests, exit 0.
- `cd folio-designer && npx tsc -b` -- exit 0. **NOT `npx tsc --noEmit`**, which is VACUOUS here: the
  root `tsconfig.json` is a solution file (`"files": []`, two references), so without `-b` tsc checks
  nothing and prints success over real errors. Red-proved 2026-09-05 by appending
  `export const DELIBERATE_TYPE_ERROR: number = "this is a string"` to a `src/` module: `--noEmit`
  returned **rc 0**; `-b` reported it. `package.json`'s own `typecheck` script is `tsc -b`.
- `cd folio-designer && npx oxlint` -- exactly 4 `only-export-components`, no fifth.
- `cd lint && go build ./...` · `go vet ./...` · `test -z "$(gofmt -l .)"` · `go test -count=1 ./...` --
  four separate commands, four exit codes, all 0.

## Suggested Review Order

**The writer, and what it refuses**

- Start here: the only writer of `Band.Height` outside the loader; validates before it mutates.
  [`component_commands.go:2184`](../../folio-go/component_commands.go#L2184)

- The 25th arm — a document-level command on the component door, as the font-chain arms already are.
  [`component_commands.go:275`](../../folio-go/component_commands.go#L275)

- The content-window invariant, in one place, asked by `Canvas` while loading and by the command while authoring.
  [`page_setup.go:729`](../../folio-go/page_setup.go#L729)

- The bound the refusal quotes comes from the predicate, so message and check cannot disagree.
  [`page_setup.go:737`](../../folio-go/page_setup.go#L737)

- DataPath for a refusal that names a band, not an element; bounded at a rune boundary for the host's cut.
  [`component_commands.go:2141`](../../folio-go/component_commands.go#L2141)

**The panel, which enforces nothing**

- Apply is now a sequence: band heights first, only where changed, stopping at the first refusal.
  [`App.tsx:780`](../../folio-designer/src/App.tsx#L780)

- The two rows, rendered only when the engine actually projected that band.
  [`App.tsx:1683`](../../folio-designer/src/App.tsx#L1683)

- Returns `undefined` rather than inventing a height; absent is not zero.
  [`App.tsx:2979`](../../folio-designer/src/App.tsx#L2979)

- The encoder, routed through the command-JSON authority like every other factory.
  [`band-height-command.ts:37`](../../folio-designer/src/band-height-command.ts#L37)

**The band list, kept to one copy**

- The type and the iterable array, both derived from the mirrored Go list rather than respelled.
  [`engine-protocol.ts:111`](../../folio-designer/src/engine-protocol.ts#L111)

**Tests worth reading as evidence, not as coverage**

- The footer mirror: without it, a footer command could write the header with the whole suite green.
  [`band_height_command_test.go:581`](../../folio-go/band_height_command_test.go#L581)

- The vacuity guard: a strand below the band's first element, so a one-element check cannot pass.
  [`band_height_command_test.go:215`](../../folio-go/band_height_command_test.go#L215)

- Sends back the bound the message names, then one millipoint more.
  [`band_height_command_test.go:278`](../../folio-go/band_height_command_test.go#L278)

- Exported so a root-package diagnostic can quote a length in the author's units.
  [`decimal.go:256`](../../folio-go/internal/template/decimal.go#L256)
