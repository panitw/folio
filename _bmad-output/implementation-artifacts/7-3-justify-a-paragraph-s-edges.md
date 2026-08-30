---
title: 'Story 7.3: Justify a paragraph''s edges'
type: 'feature'
created: '2026-08-30'
status: 'done'
baseline_revision: '8fed42ffc5f9f5f9acf3c5f1277f9ff6616ac0ad'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: ['multiple-goals', 'oversized'] # multiple-goals: the justify feature and DW-24's corpus closure are separably shippable, and D-7.1.4 accepted that cost explicitly when it made 7.3 DW-24's owner. oversized: the ruling set (D-7.3.1, D-R7.9, D-7.1.5, D-7.2.1, D-7.2.6) plus a format MAJOR plus a six-site rounding closure are three wide surfaces that must be stated, not summarised.
deferred:
  - summary: >-
      The `valign` arm of `applyPropertyChanges` accepts any string, so the command layer can
      serialize a document its own loader refuses.
    evidence: |-
      Story 7.3 closed exactly this hole for the `align` arm (it previously "set style.align to
      whatever string arrived") by validating through `template.IsStyleAlign`. The adjacent
      `valign` arm was left unguarded, while `closedValigns` is a real load-time closed set and
      `serialize.go` writes `st.Valign.Value` verbatim. `updateComponentProperties` with
      `valign: "sideways"` therefore succeeds, serializes, and then fails to load back.
      Pre-existing, not caused by this story, but its sibling's fix makes the asymmetry live.
    location: >-
      folio-go/component_commands.go:1071-1080
    severity: medium
  - summary: >-
      `style.align: "justify"` on a table element or its `headerStyle` loads, forces the document
      to format 2.0, and then renders every cell at the start edge with no diagnostic.
    evidence: |-
      The intent contract requires this: it directs that `headerStyle.align: "justify"` must load
      and must raise the document to 2.0, while its Never list forbids implementing justified
      table cells and calls justified columns "a separate scope decision". So the diff is
      contract-correct, and the residue is a real product question the contract deliberately did
      not settle: a document unreadable to every 1.x reader renders identically to `align: left`.
      This story pins that fallback with a test and corrects the three `default:` arm comments
      that wrongly claimed the load-time check had already rejected such a value; deciding whether
      it should instead be a located load error, carry a diagnostic, or actually justify cells is
      the deferred scope decision.
    location: >-
      folio-go/table_render.go:701, :1044, :1228
    severity: medium
  - summary: >-
      Closing DW-24 removes the only tracked address for the image-centring rounding at
      render.go:505-506, which remains golden-uncovered.
    evidence: |-
      DW-24's closing note measures that site GREEN under a truncation mutation and records it as
      out of the entry's declared subject (it is unconditional on every image element rather than
      selected by a declared value). That reasoning is sound, but with DW-24 closed no open item
      names the gap, so the next re-derivation will meet it a third time with no owner.
    location: >-
      folio-go/render.go:505-506
    severity: low
  - summary: >-
      A large class of ordinary Thai cannot be rendered at all: any sequence that stacks two marks
      above a base (for example `ท` + `ั` + `้`) fails closed with
      `face Noto Sans Thai: CID N carries a non-zero vertical offset`.
    evidence: |-
      Measured while choosing this story's Thai fixture text. Every Thai codepoint renders in
      isolation (91/91 over U+0E01..U+0E5B), but a bisect over a natural Thai sentence found the
      four-rune window `าทั้` failing. It is independent of justification: the same string fails
      identically with `align: left`, so it is not a justification defect and is not this story's to
      fix. It is `internal/pdf`'s deliberate AC6 fail-closed branch (`textdoc.go:1006-1019`) and the
      refusal is correct as written, but it means the fixture text for any Thai document has to be
      chosen around mark stacking, which is not a property a document author can be expected to know.
      Nothing in the tree tracks this today.
    location: >-
      folio-go/internal/pdf/textdoc.go:1006-1019
    severity: medium
  - summary: >-
      `CanvasTextLine.Width` still projects the packer's ragged measurement for a justified line,
      though the line's fragments now span the full declared width.
    evidence: |-
      Canvas/PDF parity is asserted on fragment count, text and X, which is what the contract
      requires, and `Width` agrees between the two Go paths so no test fails. But the browser
      validator already compensates with `Math.max(paint.width, component.width)`, and any
      designer consumer using `paint.width` for hit-testing, caret placement or selection will be
      wrong on exactly the lines this story adds. Not contract-covered either way.
    location: >-
      folio-go/page_setup.go
    severity: low
  - summary: >-
      `fixtures/justified-thai/` is the first golden anywhere to insert visible inter-word space into
      continuous Thai, and it carries no human Thai sign-off record.
    evidence: |-
      The repository has a decided mechanism for exactly the irreducibly-human half of D-000.22:
      `fixtures/shaped-text/thai-signoff.json` (D-2.3.5, Thai mark placement) and
      `fixtures/expected-breaks/break-signoff.json` (D-2.4.3, "every marked seam falls between words
      and never inside one"), each enforced by a `//go:build matrix` red gate so the story commits
      green and the epic gate cannot pass. Neither binds here: D-000.26 binds a sign-off to the
      artifact expressing the property judged, and Story 4.7 created a third record rather than
      reusing an existing one.
      The precedent is NOT "every Thai-bearing golden gets a sign-off" -- measured at this revision,
      `fixtures/wrapped-text/` carries 47 Thai characters with no record, and `multi-script-fallback`
      one. Those fixtures only BREAK Thai at seams `expected-breaks` already signed off. This one
      inserts gaps of up to 3,528 mp between Thai words that Thai normally writes with no space at
      all, across 432 Thai characters -- more than any other golden. Whether that reads correctly is
      a question no machine test can answer and none of the new tests attempts.
      Not routed as an intent gap: the amendment enumerates AC-TH3's completion condition
      (`goldenDigestRecord`, `matrixDocuments`, the re-derived enumeration, a digest identical on all
      four targets) and the diff satisfies it exactly. Recorded here because only the owner can
      decide to commission a Thai reader, and an agent writing a `reader`/`date`/`examined` record
      would be fabricating an attestation.
    location: >-
      fixtures/justified-thai/ (precedent: folio-go/shaped_signoff_matrix_test.go)
    severity: medium
  - summary: >-
      No Thai instance exists anywhere in the canvas/PDF justification parity claim.
    evidence: |-
      `TestCanvasPaintMatchesTheShippingRunPathUnderJustification` runs on `justifyTemplateJSON`
      ("alpha beta gamma delta epsilon", Roboto), and the browser-side half validates a justified
      component against fabricated Latin fragments. This dispatch changed nothing under
      `folio-designer/`, so the contract's `Canvas parity` row and the AD-17 / Story 5.9 invariant are
      still witnessed only by Latin. The amendment's ACs are all PDF-surface (its own evidence table
      counts pieces, TJ arrays and diagnostics), so this is outside AC-TH1/2/3 rather than a miss --
      but a word-grained Thai line is where the projection would diverge first, and DW-25 already
      owns the neighbouring fragment-cap question for Story 7.4.
    location: >-
      folio-go/text_alignment_test.go (TestCanvasPaintMatchesTheShippingRunPathUnderJustification)
    severity: low
  - summary: >-
      `verticalOffsetError`, the AC6 fail-closed branch that refuses a large class of ordinary Thai,
      has no test anywhere.
    evidence: |-
      Pre-existing, surfaced while choosing this story's fixture text. Grepping
      `verticalOffsetError|non-zero vertical offset` across `*.go` returns only the four hits inside
      `internal/pdf/textdoc.go` itself -- the raise site, the type, and its `Error()` method. Nothing
      constructs a document that reaches it and nothing pins the message, even though it is a hard
      `Render` error rather than a diagnostic and it is what makes `ครั้ง`, `ทั้งนี้` and `ตั้งแต่`
      unrenderable. Distinct from the sibling entry above, which records the LIMIT; this one records
      that the refusal itself is unguarded and could regress in either direction unnoticed.
    location: >-
      folio-go/internal/pdf/textdoc.go
    severity: low
---

## In plain terms (read this first if you just want the gist)

*Non-normative: this section settles nothing.*

Legal body copy is set with both edges flush: every line reaches the same right-hand margin, and the leftover space is shared out between the words. This library can already push a paragraph left, right or centre; this story adds the fourth setting, with the rules that make it read correctly. The last line of a paragraph stays ragged, because stretching it would look absurd. So does any line the author ended by typing a break, and any line with nowhere to put extra space. Space is shared in whole units, the leftover going to the earliest gaps in a fixed order, so the document prints identically on every machine.

There is a real cost. Older readers of this format reject any alignment word they do not recognise, on purpose, so a file can never be drawn wrongly by a reader that misunderstands it. A document using the new setting is therefore unreadable to them rather than quietly wrong, and the file's own version number moves up a whole step to say so honestly. Nothing has been released and nobody depends on the old numbering, so this is free now and would not be later.

This story also closes a standing hole in the corpus: no example document had ever used the centred settings, so the arithmetic that halves leftover space was never checked on real output.

One test is expected to stay red. It is a deliberate standing marker, not a defect.

<intent-contract>

## Intent

**Problem:** A contract is body copy justified to both margins, and the format cannot express it: the alignment vocabulary is a closed set of `left`/`center`/`right`, enforced as a load error, and nothing distributes a line's leftover width across its gaps. Separately, the corpus has never contained a document declaring `center` or `middle`, so every rounding branch in the alignment feature has zero golden coverage and the four-target byte-identity claim is unfalsifiable for exactly the arithmetic most likely to break it.

**Approach:** Add `justify` to the *style* alignment set only, splitting the one shared closed set so table columns keep their existing three values, and derive both call sites' error message from the set rather than restating it. Distribute a justified line's slack across its interior break opportunities by one stated integer rule, applied in the shared placement layer both the PDF producer and the canvas projection already call, so the two cannot disagree. Because extending a closed set is a MAJOR change under D-1.4.12 and D-R7.9 rules the MAJOR free, raise the format to 2.0 and extend the content-derived version rule to report it. Close DW-24 with a second fixture that reaches every rounding site a fresh grep returns.

## Boundaries & Constraints

**Always:**
- `justify` joins the **style** alignment set only. `columns[].align` keeps exactly `{left, center, right}`. The two sets are separate declarations; neither is derived from the other by subtraction at a call site.
- Both closed-set rejection messages are **derived from an ordered token slice**, never from a map (ranging a map under `internal/` is a build failure, D-1.3.5) and never from a hand-written literal. After this story no message may claim a set of legal values that differs from the set actually enforced.
- Slack distribution is **integer-exact and total**: no float, no second rounding site, no division whose remainder is discarded. The distributed amounts must sum to the slack exactly.
- The remainder rule is **stated and ordered**: with `g` gaps and slack `s`, every gap receives `s / g`, and the first `s mod g` gaps in ascending position along the line each receive one additional millipoint. Ascending order is the line's own reading order and is not a function of map iteration, locale, or target.
- **Two independent conditions leave a line unjustified, and they are never collapsed into one flag:** (1) the line was ended by a mandatory break — read from the break-kind field Story 7.1 wrote for this consumer (D-7.1.5); (2) the line is the last line of the element — **derived from the line's index at this story's call site and never stored**, because that field answers "which break ended this line", not "is this the last line".
- A third condition, stated because the acceptance criteria are silent on it: a line with **zero interior break opportunities** has nowhere to place slack and is set at the element's natural start edge, exactly like a last line. AD-25 makes this reachable — an unknown Thai run is atomic and offers no interior opportunity.
- An element with **no declared width** has no box to justify to and is set at its natural start edge, consistent with the existing slack-only rule.
- Justification distributes **slack only**. A line whose measured width meets or exceeds the declared width has no slack; FR44's clip-and-warn applies unchanged and overflow detection continues to read the packer's own measured line width.
- The justified rule lives in **one shared function consumed identically by both producers** — the PDF path and the canvas paint projection — as the existing alignment rule already is. The canvas must show the same word positions the PDF prints (AD-17, the Story 5.9 invariant); it must achieve this by consuming engine-computed positions, never by enabling browser justification.
- The format version is a property of the **document**, raised only by the content it contains (D-1.4.13, D-7.2.1). A document using only `lineSpacing` or `color` still declares 1.1; one using neither still declares 1.0; one using `justify` declares 2.0. **All three coexist, and a brand-new document declares the lowest version its content requires — never the library's ceiling.** The corrected ceiling doc comment must not regress.
- The content-derived version must report the **highest** requirement in the document, not the first one found.
- **Corpus neutrality is asserted, not assumed.** Every recorded golden digest must be measured unchanged, and the record is invalidated in whole if any one moves.
- Every new fixture is registered at **every** surface that renders it, or it exists and nothing runs it.

**Block If:**
- DW-24 cannot be closed as specified here. It has been declined at 7.1 and 7.2 and **is not deferrable a third time** (D-7.2.6); a decline is an **escalation to the engineering lead**, not a fourth deferral entry. HALT with blocking condition `DW-24 cannot be closed`.
- Any recorded corpus digest moves. HALT with blocking condition `corpus digest moved`; do not re-record a golden to make a suite pass.
- Delivering the shared word positions to the canvas turns out to require **widening a canvas projection bound**. That bound is DW-25's subject and 7.4 owns it. HALT with blocking condition `canvas projection bound must widen`.
- The re-derived rounding enumeration returns a site that cannot be reached by any single document. HALT with blocking condition `rounding site unreachable by fixture`.

**Never:**
- Never add `justify` to the set `columns[].align` validates against, and never implement justified table cells. Justified columns are a separate scope decision, not a side effect of a map edit.
- Never propose `style.justified` or any additive-key spelling as an alternative to the MAJOR. Explicitly rejected by D-R7.9: it reintroduces the silently-wrong render D-1.4.12 exists to prevent.
- Never edit `internal/layout/paginate.go`. The window model is an input to Epic 7, not a target of it.
- Never do designer **editor** work: no inspector control offering the new alignment, no multi-line value editing. That is Story 7.4.
- Never widen the canvas projection bounds (DW-25, owned by 7.4) and never close DW-26.
- Never touch `TestCorpusMeetsP6ExerciseFloors` or the P6g floor. It is a mandated permanent red.
- Never bolt a new case onto a `statement-*` fixture: those four carry a human sign-off that any move invalidates in whole (D-4.7.1).
- Never add to the five-entry hard-pinned baseline acceptance table.
- Never re-record an existing golden. Never create a branch, never push, never `git add -A`, and never stage, modify or delete the repository-root `README.md`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Style justify accepted | element `style.align: "justify"` | Loads; round-trips verbatim | No error expected |
| Column justify still refused | `columns[].align: "justify"` | Located load error naming the column and field | Message lists exactly left, center, right |
| Derived message honesty | any illegal style align value | Located load error listing exactly the style set's members, including justify | Message derived from the ordered token slice |
| Interior line justified | justified paragraph, a non-final line with `g >= 1` interior opportunities and positive slack | Slack distributed across the gaps; the line's drawn right edge meets the declared width exactly | No error expected |
| Last line ragged | the final line of a justified element | Set at the natural start edge, unjustified | No error expected |
| Mandatory-break line ragged | a justified line whose break-kind field records a mandatory break | Set at the natural start edge, unjustified, even though it is not the last line | No error expected |
| Both conditions at once | final line that is also mandatory-break-ended | Unjustified; the two conditions are independent and neither is inferred from the other | No error expected |
| No interior opportunity | justified line over an atomic unknown Thai run | Natural start edge; no slack placed | No error expected |
| Exactly one gap | justified line with a single interior opportunity | The whole slack goes into that one gap | No error expected |
| Remainder placement | slack 7, gaps 3 | Gaps receive 3, 2, 2 in ascending order; total is exactly 7 | No error expected |
| Overflow unchanged | justified line wider than the declared width | Existing FR44 clip-and-warn, natural start edge, no negative distribution | Warning beside successful output |
| No declared width | justified element with no width | Natural start edge | No error expected |
| Byte neutrality | elements with no align, or left/center/right | Bytes identical to baseline at every recorded digest | Any movement is a HALT |
| Version raised by justify | document using `align: justify` | Serializes `2.0` | No error expected |
| Version not raised otherwise | document using only lineSpacing or color; document using neither | Serializes `1.1`; serializes `1.0` | Never raised to the library ceiling |
| Version never lowered | document declaring a higher version, no justify | Round-trips verbatim | No error expected |
| Highest requirement wins | document whose first styled element sets lineSpacing and a later one sets justify | Serializes `2.0`, not `1.1` | First-hit return is a defect |
| Higher MAJOR still refused | document declaring `3.0` | Load error naming declared and supported versions | No render attempted |
| Canvas parity | justified element projected to the canvas | Fragment count, text and X match the shipping run path per line and per fragment | Divergence fails the parity test |

</intent-contract>
## Scope amendment — 2026-08-30, owner request

**The owner asked directly: "For 7.3, I'd like to make sure Thai text can be justified as well."**
This section is normative for the amending dispatch and sits inside the contract's scope.

**Verified by the orchestrator before amending** (so this is coverage for behaviour that already works,
not a change of behaviour). Rendering the same Thai paragraph justified vs left, `Noto Sans` /
`Noto Sans Thai` / `Noto Sans SC` chain, 200pt box:

| case | pieces on line 1 | TJ arrays | diagnostics |
|---|---|---|---|
| continuous Thai (no spaces), `justify` | **9** | 10 | 0 |
| continuous Thai, left | 1 | 3 | 0 |
| spaced Thai, `justify` | 8+ | 14 | 0 |

Nine independently positioned pieces on one baseline proves the **dictionary-derived** break
opportunities are being used as justification gaps — Thai is not silently falling back to ragged
left. `justifiedLinePieces` distributes over the opportunity list and is script-agnostic.

**The gap is coverage, not behaviour.** `fixtures/justified-text/` is pure Latin, and no test in the
tree names Thai and justify together. This is the same absence that let `valign` ship uncovered and
cost DW-24 three stories.

### AC-TH1 — a justified Thai line is justified, not ragged
**Given** a text element with `align: "justify"` whose value is **continuous Thai with no spaces**
**When** it is rendered
**Then** each non-last line is justified across its dictionary-derived break opportunities, and its
right edge meets the declared width exactly — asserted at the same standard as the Latin case, in
integer millipoints.

### AC-TH2 — the no-interior-opportunity fallback is exercised, not just stated
**Given** a Thai run that yields **zero** interior break opportunities (AD-25's atomic-unknown-run)
**When** it is justified
**Then** it is set at its natural start edge, exactly like a last line — the third ragged condition
the ACs were originally silent on. **This must be a test, not a comment**: pick a Thai run the
dictionary does not cover, and assert the precondition (zero interior opportunities) with a
`t.Fatalf` so the case cannot go vacuous if the dictionary later covers it.

### AC-TH3 — a Thai justified fixture joins the corpus
**Given** the golden corpus
**When** this story closes
**Then** it contains a fixture whose justified content is Thai, registered at every golden surface
(`goldenDigestRecord`, `matrixDocuments`, and the rest of the enumeration DW-24's closure re-derived),
with a recorded digest identical on all four targets.

**Scope fence.** This adds **coverage and its fixture only**. It changes no justification behaviour,
no break rule, and no format field. If the Thai case turns out NOT to justify correctly, that is a
finding to report — **HALT** rather than changing `justifiedLinePieces` to make a new fixture green.

**Known pre-existing limit, NOT this story's to fix and NOT a justification defect.** Some Thai fails
to render at all: `face Noto Sans Thai: CID 20 carries a non-zero vertical offset (-57), which a TJ
array cannot express`. The orchestrator verified it fails **identically** with `align: left`, so it is
independent of justification and is a deliberate documented refusal, not a crash. **Choose fixture
text that renders** (the statement fixtures' Thai does). If you cannot find Thai that both renders and
exercises AC-TH2, say so rather than working around it — that combination is itself a finding.


## Code Map

### The closed set and its two enforcement sites

- `folio-go/internal/template/closedsets.go:30-32` -- `closedAligns = {left, center, right}`, the ONE map shared by both align checks. This is the trap: adding `justify` here silently legalises justified table cells.
- `folio-go/internal/template/parse_bands.go:393-403` -- **columns[].align**. Rejects with the hand-written literal `"not one of the closed set left, center, right"` at **:400**. Field name is the bare `"align"`; `id` is the column id. **Keeps the three-value set.**
- `folio-go/internal/template/parse_bands.go:525-535` -- **style.align / headerStyle.align**, inside `decodeStyle(elementID, raw, fieldPrefix)` (`:508-518`). Same hand-written literal at **:532**. `fieldPrefix` is `"style"` or `"headerStyle"` so the error is located at the right one. **This is the set that gains `justify`.**
- **Derived-message precedent, and the only one in the repo:** `folio-go/internal/expr/locale.go:126-133` formats the legal set into the message from `template.LocaleTags`, an **ordered `[]string`** (`internal/template/locale.go:39-48`), never from the map — because ranging a map under `internal/` is a **build failure** (D-1.3.5, enforced by `ScanMapRange` with its corpus at `folio-go/testdata/lint/map-range/`), and because the order must be an asserted literal sequence. `TestClosedLocalesMatchesLocaleTags` pins slice and map against each other. Copy this shape exactly.
- `folio-go/internal/template/closedsets_test.go:39` `scanClosedSets` extracts set names and every string-literal value, but `TestClosedSetsNeverIncludeMediaType` (`:82`) only checks media-type shapes. **No test anywhere enumerates `closedAligns`' members** — so the split needs its own pinning test.

### The line packer and the seam Story 7.1 left

- `folio-go/wrap.go:169-173` -- `wrappedLine{from, to int; width geom.Length; endedBy text.BreakKind}`. Rune indices are element-global, half-open. **`endedBy` is 7.1's field, written there and read HERE** (D-7.1.5); its doc at `:155-168` states that a line no break ended carries the zero value `BreakOptional`, and that **7.3 must derive the last-line case from the index, not from this field**.
- `folio-go/internal/text/opportunity.go:27-48` -- `BreakKind` with `BreakOptional` (zero) and `BreakMandatory`. A **named kind, not a bool**, precisely so the third case is not re-derived wrongly.
- `folio-go/internal/text/opportunity.go:62-66` -- `Opportunity{LineEnd, NextStart int; Kind BreakKind}`. `LineEnd != NextStart` **only** where the break consumes text — today exactly one case, a whitespace run (`:196-215`). A CJK/Thai seam is zero-width (`LineEnd == NextStart`).
- `folio-go/wrap.go:207` -- `packLines(segs, ops, totalRunes, fontSize, maxWidth) []wrappedLine`. The greedy pick at `:276-302` holds the winning opportunity index in `chosen`/`first`, and the append at `:305-326` copies only `LineEnd`, the re-measured width and `Kind`. **Interior opportunities are not retained** — but they need no re-derivation: the SAME `ops` slice is live at both consumers (`render.go:834`, `page_setup.go:450`), so a justified line's gaps are recovered by **filtering the ops the caller already holds**, not by computing breaks a second time.
- `folio-go/wrap.go:349` -- `packMandatoryOnly`, the `maxWidth <= 0` path. Consistent with "no declared width, no justification".
- `folio-go/wrap.go:118-131` -- `measureRuneRange`. Its doc states it **reproduces `positionSegments`' cursor arithmetic exactly, because it IS that arithmetic**. This identity is what makes per-piece measurement and per-piece positioning agree, and it is the reason the justified right edge can land on the declared width exactly.

### Placement — the shared rule both producers call

- `folio-go/text_alignment.go` -- the whole placement rule, shared by the PDF producer and the canvas projection **so that the two cannot drift** (doc `:8-26`). `elementAlignment` `:30`; `textAlignOffset` `:47-60` (**`geom.ScaleRound(slack, 1, 2)` at `:56`** — a DW-24 site); `textValignOffset` `:72-85` (**`ScaleRound(slack, 1, 2)` at `:81`** — a DW-24 site); `textBlockHeight` `:91-96`. `textAlignOffset`'s `default:` arm returns 0 for any unrecognised align, so `justify` reaching it unchanged would silently produce a ragged left-edge line — correct for the ragged cases, and the reason the justified path must be an explicit branch rather than a fallthrough.
- `folio-go/render.go:1384-1432` -- `positionSegments(segs, from, to, x, y, fontSize, baselineOffset, slots) ([]textRunSource, error)`. **One cursor per call**, one run per face segment overlapping `[from,to)`, each stamped `x: cursor` (`:1403`), cursor advanced by `geom.ScaleRound(advance1000(lo,hi), fontSize, 1000)` (`:1429`). Run boundaries today are **face-segment** boundaries, so an all-Latin line is ONE run at ONE x. `slots` are `{{page}}` reservations; a slot straddling a run boundary is a **located error, never a panic** (`:1418-1424`) — piece splitting widens the set of inputs that can reach it, so it needs a tripwire.
- `folio-go/render.go:898-901` -- the PDF line loop: `lineY`, then `lineX := el.X + textAlignOffset(...)`, then `positionSegments`. `boxWidth` at `:836-839`; overflow detected from `ln.width` at `:848` **before** placement; `align, valign` at `:882`; `elementY` with the valign offset at `:894`.
- `folio-go/page_setup.go:484-485` -- the canvas projection, **the identical two lines**. That parity is the whole of the current canvas/PDF agreement.
- Interior inter-word spaces **are drawn glyphs** with their own advance: consumption applies only at the break actually taken, so within a line the range `[from,to)` is contiguous and every space is shaped, positioned and measured. Justification therefore adds space *between pieces*; it does not delete or re-shape anything.

### The canvas channel — already word-position-capable

- `folio-go/page_setup.go:30-53` -- `CanvasTextFragment{Text, X}`, `CanvasTextLine{Top, Baseline, Advance, Width, Fragments}`, `CanvasTextPaint{Overflow, Lines}`. The fragment array is **already an array of positioned pieces**; only the *granularity* is face-segment today.
- `folio-designer/src/App.tsx:1168-1171` + `folio-designer/src/App.css:75,79-81` -- each fragment is an absolutely positioned span at `left: var(--text-fragment-x)` with `white-space: pre`. **No wire-schema change, no CSS change, and no validator-clause change is needed** for word-grained fragments.
- `folio-designer/src/canvas-authority-contract.test.ts:28,58` -- browser `text-align: justify` is **contractually banned** across production, unit and e2e sources, and `:58` red-proofs the scanner. This stays green and is the correct outcome: the engine positions, the browser never justifies.
- `folio-go/canvas_text_paint_test.go:49` -- `TestCanvasTextPaintExactlyMatchesTheShippingRunPath` compares **per line AND per fragment**: fragment count (`:86-88`), `Top`/`Baseline` (`:97-99`), and **`fragment.Text` and `fragment.X` per fragment** (`:100-103`), plus `Width` (`:105-107`). This is the guard that catches a PDF-only justification, and it pins word positions for free once fragments are word-grained.
- `folio-go/text_alignment_test.go:140` -- `TestCanvasPaintMatchesTheShippingRunPathUnderAlignment`, the align-specific parity claim over five components; the natural place to extend for `justify`.
- **The bound to measure, not to widen:** `folio-go/page_setup.go:28` `maxCanvasTextFragments = 512` is enforced **per line** (`page_setup.go:503`, and exceeding it returns an error that aborts the whole projection), while `folio-designer/src/engine-protocol.ts:236` enforces `fragments <= 512` **cumulatively across the whole component**. Word-grained fragments multiply the count by words-per-line, so the TS cumulative cap is the cliff a long justified paragraph reaches first. **This is DW-25's subject and 7.4 owns it** — measure the reachability and record it; do not raise either number.
- `folio-designer/src/engine-protocol.ts:157` -- `['left','center','right'].includes(component.align)`; type at `:61` `align?: 'left' | 'center' | 'right'`. **Both must admit `justify`** or a justified document blanks the entire projection (`:239` drops the whole response). `TableColumn.align` at `:29` and the columns check at `:111` are the **columns** vocabulary and stay the triple.
- `folio-designer/src/App.tsx:881-885` -- `alignGlyphs` and `alignSegments`, the inspector's segmented control, typed over the literal triple independently of the component type. **Story 7.4's**; leave it. Verify only that widening the component type does not break `npm run typecheck` — and if it does, fix the type flow without adding the control.

### Format version

- `folio-go/internal/template/version.go:20-41` -- `SupportedMajor = 1`, `SupportedVersion = "1.1"`, with the doc comment 7.2 **corrected**: "IT IS NOT WHAT THIS LIBRARY AUTHORS FOR A BRAND-NEW DOCUMENT … `version` is a property of the DOCUMENT, raised only by the content the document actually carries". **Do not regress this.**
- `folio-go/internal/template/version.go:43-52` -- `baseVersion = "1.0"`, `minorFeatureVersion = "1.1"`. A third constant is needed.
- `folio-go/internal/template/version.go:74-88` -- `checkVersionLoadable`; refuses only `major > SupportedMajor` (`:84`).
- `folio-go/internal/template/version.go:127-144` -- `versionForSave(loaded, d)`. `required == baseVersion` **short-circuits before any numeric comparison** so a `"0.9"` document is not restamped; then returns the higher of loaded and required. Comparison already handles MAJOR (`requiredMajor > loadedMajor || (equal && requiredMinor > loadedMinor)`), so **2.0 vs 1.1 needs no new comparison logic**. Exactly one caller: `serialize.go:118`.
- `folio-go/internal/template/version.go:163-181` -- `versionRequiredByContent`. Walks the three bands, each element's `Style` and its `Table.Value.HeaderStyle`, both guarded `Set && !Null`. **It returns `minorFeatureVersion` on the FIRST hit** — so a document whose first styled element sets `lineSpacing` and a later one sets `justify` would report 1.1. This is a real restructure to a maximum, not a one-line add.
- `folio-go/internal/template/version.go:185-187` -- `styleNeedsMinorVersion(st) bool { return st.LineSpacing.Set || st.Color.Set }`. Becomes a rank.
- **Raising `SupportedMajor` breaks a test by design:** `folio-go/internal/template/version_test.go:16` `TestHigherMajorIsLoadError` uses `withVersion("2.0")` and asserts a load error; it must move to `"3.0"`. The public twin `folio-go/template_test.go:106-111` uses `"9.0"` and is safe.
- `folio-go/internal/template/linespacing_test.go:111` `TestVersionForSaveIsRaisedOnlyByContent` — a nine-case table including the never-lowered and `"0.9"`-floor pins. `:178` `TestContentVersionNeverExceedsTheLibraryCeiling` asserts `ceilingMajor == SupportedMajor`, hard-pins `baseVersion == "1.0"` (`:196-198`), and enumerates `[]string{baseVersion, minorFeatureVersion}` at `:186` (and a style-shape loop at `:203`) — **hand-maintained lists that go vacuous unless the new constant is added.**
- **Nothing changes for existing documents:** `versionForSave` never consults `SupportedMajor`/`SupportedVersion`. 18 `fixtures/*/input.folio` declare `1.0`, `fixtures/line-spacing/input.folio` declares `1.1`, and `folio-go/testdata/template/golden/worked-example.json:106` and `testdata/example/first-pdf.folio` declare `1.0`. All keep declaring exactly what they declare today.
- The de facto new-document seed is `folio-designer/public/templates/starter.folio:19` (`"version": "1.0"`). There is **no Go constructor that authors a version for a blank document**; the lowest-version rule is pinned only indirectly, which is why it needs an explicit test here.

### Property command path

- `folio-go/component_commands.go:290-294` -- **columns** align: a hardcoded `left/center/right` triple with its own message. `folio-go/component_commands_test.go:336` already uses `"justify"` as its invalid-column-value case and asserts rejection — **that test stays green and becomes a red-proof of the split.** Leave both alone.
- `folio-go/component_commands.go:1048-1057` -- **style** align arm: sets `st.Align` to whatever string arrives, with **no closed-set validation at all**. Pre-existing, and now the one place the two vocabularies could be conflated. Validate against the style token set through the same single source.

### Test fences a new closed-set VALUE trips (and the ones it does not)

- `folio-go/internal/template/drift_test.go` -- `TestDriftGoToDoc` `:232`, `TestDriftDocToGo` `:269`, `TestDriftASTMatchesRuntimeEmission` `:402` compare **key names only**; `fenceKeyAnywhere` (`:103`) requires a trailing colon and its comment (`:124-127`) says it "never picks up a closed-set VALUE". **Adding `justify` requires no doc change to keep these green** — the doc change below is for correctness, not to satisfy a fence.
- `folio-go/internal/template/fixtures_test.go:236` `maximalFixture` already exercises align at four sites and declares `"version": "1.1"` at `:406`. It needs no `justify` — and if anything adds one, its declared version must become `2.0` or the round-trip assertions will fight `versionForSave`.
- `folio-go/internal/template/numeric_classification_test.go` -- **untouched**: `justify` adds no numeric key.
- `folio-go/internal/template/goldenfixture_test.go:16-25` + `folio-go/testdata/template/golden/worked-example.json` -- **leave the worked example alone**; it uses no justify and stays 1.0.

### DW-24 — the entry, and the re-derivation this story owes

- `_bmad-output/implementation-artifacts/deferred-work.md:251` -- DW-24. Owner **Story 7.3** with the orchestrator's gate checklist as a second standing address (D-7.1.4). Trigger has fired twice and it is **not deferrable a third time** (`:277`, D-7.2.6); 7.3 treats it as an **acceptance criterion**. Closure conditions verbatim at `:342-347` and `:367-380`; falsifier at `:389-399`.
- **The mandated closure step:** `:342-347` requires the enumeration be **re-derived by grep and recorded in the closing note**, never read off the hand-list, "because the list is being amended today precisely because it had already rotted once".
- **Re-derivation performed at planning (HEAD `0cd9491`), and it confirms the hand-list has rotted a second time — every line anchor has drifted:**

| DW-24 hand-list | actual at `0cd9491` | branch required |
|---|---|---|
| `text_alignment.go:56` | `text_alignment.go:56` | `align: center` |
| `text_alignment.go:74` | **`text_alignment.go:81`** | `valign: middle` |
| `table_render.go:687` | **`table_render.go:705`** | header cell `align: center` |
| `table_render.go:698` | **`table_render.go:716`** | header cell `valign: middle` |
| `table_render.go:1017` | **`table_render.go:1038`** | body cell `align: center` |
| `table_render.go:1193` | **`table_render.go:1214`** | footer cell `align: center` |
| *(absent from the hand-list)* | **`table_render.go:966`** `slack / 2` | table element `valign: middle`, body row |

  `table_render.go:966` is an integer **line-count** split, not a `geom.ScaleRound` and not a cross-target float hazard (its own comment at `:953-955` argues it out of scope) — but it is a `valign: middle`-only branch with zero golden coverage, which is the absence DW-24 exists to record. Record it in the closing note and reach it with the same fixture. `render.go:505-506` (image centring) is **unconditional on every image element** and remains out of subject, as `:338` already notes.
- **Census confirming the hazard, measured at planning:** across all `fixtures/*/input.folio` — **16 `left`, 8 `right`, 0 `center`, and 0 `valign` of any value**. All 24 declarations live in the four statement fixtures (five `columns[].align` and one element `style.align: right` each). `headerStyle` appears in only two fixtures and neither declares align or valign. Go template consts declare none. `worked-example.json:53` has a `center` but is a **serializer round-trip golden, never rendered**, so it is not matrix coverage.

### Golden fixture registration — five surfaces, per fixture

- `fixtures/<name>/` -- `README.md`, `input.folio`, `expected.json`, `expected.pdf`. Precedent: `fixtures/mandatory-break/` (7.1) and `fixtures/line-spacing/` (7.2).
- A Go template const **kept byte-identical to `input.folio`**, plus a data const. Precedent `folio-go/mandatory_break_template.go:67` (`mandatoryBreakTemplateJSON`) and `:97` (data), asserted by `folio-go/mandatory_break_fixture_test.go:415-421` inside `TestMandatoryBreakGoldenFixture` (`:406`) — which deliberately runs **after** the semantic tests, because "a hash frozen before anyone checked what it contained certifies only that the bytes have not changed" (D-000.22).
- `folio-go/byte_neutrality_test.go:92` `goldenDigestRecord` — IS the list (D-000.47). Non-statement entries carry three sites (`expected.json`, `second-literal`, `readme`) and **no sign-off**; statement entries carry a fourth. Last entry `line-spacing` at `:365`.
- `folio-go/matrix_test.go:1317` `matrixDocuments` — 18 entries; `mandatory-break` `:1595-1596`, `line-spacing` `:1618-1619`.
- `folio-go/missing_glyph_corpus_test.go:41` corpus table **and** `beyondBaselineAcceptance` at **`:145`**. The accounting identity at `:185-186` fatals unless `len(fixtures) == len(baselineAcceptanceFixtures) + len(beyondBaselineAcceptance)`, so **both** declarations must gain each new fixture, with a non-empty story-naming reason (`:176`).
- `.github/workflows/matrix.yml:243` `docs=` list, **plus four per-target upload blocks** (`:69-70`, `:107-108`, `:145-146`, `:183-184`, each `if-no-files-found: error`). Five edits per fixture in this file.
- **Do NOT add to** `folio-go/first_baseline_acceptance_test.go:100` `baselineAcceptanceFixtures` — `:258-260` fatals on any count but five.

### Recorded baseline digests to assert unchanged

`folio-go/byte_neutrality_test.go`: `statement-1` `:231` `114df1d6…` 76,744 B; `statement-5` `:245` `70dce051…` 127,363 B; `statement-20` `:264` `56bfbbd9…` 269,884 B; `statement-50` `:280` `5d090b0f…` 555,829 B; `mandatory-break` `:338` `7cf743de…` 56,681 B; `line-spacing` `:365` `de212115…` 57,770 B. The record "is invalidated IN WHOLE if any one moves" (`:225`). The two new fixtures make eight.

### Documentation

- `_bmad-output/specs/spec-folio/folio-format.md:66-70` version rules (the closed-set MAJOR sentence, which names `align` literally); `:47` the `version` row stating the concrete content rule; `:319` the style defaults row `| align | left · also center, right |`; `:245`/`:246`/`:252` the `columns[]` rows and the column-over-style cascade; `:333-345` the Story 3.5 prose asserting a column's align is "already governed by its closed left/center/right set". `## Worked example` at `:514`, version at `:624` — **untouched**.

## Tasks & Acceptance

**Execution:**

- `folio-go/internal/template/closedsets.go` -- Replace `closedAligns` with **two ordered token slices and two maps built from them**: a style set `{left, center, right, justify}` and a column set `{left, center, right}`. Order is the literal sequence, asserted. -- Rationale: one shared map is the trap D-7.3.1 names; splitting at the declaration makes justified columns impossible by construction rather than by discipline.
- `folio-go/internal/template/parse_bands.go` -- Point `:399` at the **column** set and `:531` at the **style** set, and replace both hand-written literals (`:400`, `:532`) with a message **derived from the relevant ordered slice**, following `internal/expr/locale.go:126-133`. -- Rationale: two sets now exist; two literals that restate one of them would ship messages that lie about what is legal.
- `folio-go/internal/template/closedsets_test.go` -- Add a test pinning each ordered slice against its map (both directions, exact sequence), and asserting the style set contains `justify` while the column set does not. -- Rationale: nothing enumerates these sets today, so the split has no guard.
- `folio-go/internal/template/parse_bands_test.go` (or the nearest existing load-error test file) -- Assert `style.align: "justify"` and `headerStyle.align: "justify"` load; assert `columns[].align: "justify"` is a **located** load error; assert each message lists exactly its own set's members. -- Rationale: this is the split's behavioural red-proof.
- `folio-go/internal/template/version.go` -- Add `majorFeatureVersion = "2.0"`; set `SupportedMajor = 2` and `SupportedVersion = "2.0"`; replace `styleNeedsMinorVersion` with a **rank** (base / minor / major) and restructure `versionRequiredByContent` to take the **maximum rank over every attachment point** instead of returning on first hit. Extend the ceiling doc comment to name 2.0 and `align: justify` **without weakening its "not what this library authors for a brand-new document" correction**. -- Rationale: first-hit return would report 1.1 for a document that also uses justify; the doc comment is D-7.2.1's discharged debt and must not regress.
- `folio-go/internal/template/version_test.go` -- Move `TestHigherMajorIsLoadError` from `"2.0"` to `"3.0"`. -- Rationale: 2.0 is now loadable; the test's subject is "higher than supported", not the literal 2.
- `folio-go/internal/template/linespacing_test.go` (and/or a new justify version test) -- Add `majorFeatureVersion` to the `:186` and `:203` enumerations in `TestContentVersionNeverExceedsTheLibraryCeiling`; extend `TestVersionForSaveIsRaisedOnlyByContent` with justify cases: justify alone → 2.0; justify at the headerStyle attachment point → 2.0; lineSpacing on an earlier element and justify on a later one → **2.0**; justify with a loaded `2.1` → 2.1 (never lowered); neither key → unchanged. -- Rationale: the ceiling guard goes vacuous for an unenumerated constant, and the maximum-rule needs the ordering case that a first-hit implementation fails.
- `folio-go/internal/template/version_test.go` or `serialize_test.go` -- Assert a document carrying **no** version-raising content still serializes the lowest version its content requires and is never stamped with `SupportedVersion`. -- Rationale: the lowest-version rule is currently pinned only indirectly by fixtures; D-7.2.1's guardrail deserves a direct assertion.
- `folio-go/text_alignment.go` -- Add the justification rule as the **one shared function** beside the existing placement rules: given the line, its index, the line count, the element's declared width, the packer's `ops` and the segments, return either "natural start edge" or an ordered list of `(pieceFrom, pieceTo, pieceX)`. Gaps are the ops with `LineEnd > ln.from && LineEnd < ln.to`, in ascending order. Piece widths come from `measureRuneRange` — **the packer's own function, not a second measurement**. Slack is `boxWidth − Σ pieceWidths`; if it is `<= 0`, or `boxWidth <= 0`, or there are no gaps, or `ln.endedBy == BreakMandatory`, or the index is the last, return the natural start edge. Otherwise `base = slack / g`, `r = slack − base*g`, and the first `r` gaps take `base+1`. -- Rationale: basing slack on the summed piece widths (rather than `ln.width`) makes the drawn right edge land on the declared width **exactly**, because `measureRuneRange` is documented to reproduce `positionSegments`' cursor arithmetic; and putting the rule where `textAlignOffset` already lives keeps "one rule, both producers" true by construction.
- `folio-go/text_alignment.go` -- Extend the file's doc comment: the SLACK ONLY paragraph now covers a fourth alignment that distributes slack **between** pieces rather than before the line, and the three independent ragged conditions are named. -- Rationale: the comment is the first thing a future reader meets and currently describes only whole-line offsets.
- `folio-go/render.go` -- In the line loop at `:898-901`, when the element is justified, place each piece with its own `positionSegments` call at the piece's X and concatenate the runs, passing `slots` to each call; otherwise leave the existing single call byte-for-byte unchanged. -- Rationale: a piece is a rune range at a position, which is exactly `positionSegments`' existing contract; not branching for the unjustified case is what keeps the corpus byte-identical.
- `folio-go/page_setup.go` -- Apply the identical piece placement at `:484-485`. -- Rationale: AD-17 and the Story 5.9 invariant; the two call sites are one line apart today and must stay that way.
- `folio-go/render.go` or the nearest test file -- Add a **tripwire** asserting that a `{{page}}` reservation straddling a justified piece boundary produces the existing **located error, not a panic**, and state whether the construct is reachable through the shipped set. -- Rationale: piece splitting widens the set of inputs that can reach `:1418-1424`, and that path's own comment records that a public entry point must not let an internal panic cross it.
- `folio-go/component_commands.go` -- Validate the **style** align arm (`:1048`) against the style token set through the single shared source. Leave the column arm (`:290-294`) exactly as it is. -- Rationale: with two vocabularies live, an unvalidated style arm is the one remaining place they could be conflated.
- `folio-designer/src/engine-protocol.ts` -- Admit `justify` in the **component** align validator (`:157`) and its type (`:61`). Leave `TableColumn.align` (`:29`) and the column check (`:111`) as the triple. -- Rationale: the validator gates the projection, so a justified document would otherwise blank the entire canvas; the inspector control that offers the choice is 7.4's.
- `folio-designer/src/engine-protocol.test.ts` -- Assert a `justify` component is admitted, a `justify` **column** is rejected, and word-grained fragments with ascending X are admitted. -- Rationale: the split must hold on the browser side too.
- `fixtures/justified-text/` -- New golden: `input.folio` declaring **`"version": "2.0"`** with a justified multi-line paragraph whose lines have differing gap counts, at least one line ended by a mandatory break, and a final ragged line; plus `README.md`, `expected.json`, `expected.pdf`. -- Rationale: every feature ships its golden fixture, and this one red-proves the version raise as well as the placement.
- `folio-go/justified_text_template.go` (new) -- The template const kept byte-identical to `input.folio`, plus the data const, following `mandatory_break_template.go`. -- Rationale: the established hand-sync precedent; data reaches a fixture through a Go const, not a data file.
- `folio-go/justified_text_fixture_test.go` (new) -- Semantic tests **first** (gap counts, the ragged last line, the ragged mandatory-break line, the exact right edge, the remainder order), byte-identity of const vs `input.folio`, then the golden. -- Rationale: D-000.22 — a hash frozen before anyone checked what it contained certifies only that the bytes have not changed.
- `fixtures/alignment-rounding/` -- New golden closing DW-24: one document reaching **every site the re-derived enumeration returns** — a text element with `align: center`, a text element with `valign: middle` (and one with `valign: bottom`, unrounded but equally undeclared), and a table with a centred column carrying a **footer** so header, body and footer cells all round, plus `headerStyle.valign: middle` and a table-element `valign: middle` to reach the integer line-count split. **At least one box must leave an ODD millipoint slack** so the half-to-even tie is actually taken. Declares `"version": "1.0"` — it uses no 1.1 or 2.0 key. -- Rationale: a centred text element does not cover the table sites; different code, same rounding.
- `folio-go/alignment_rounding_template.go` + `folio-go/alignment_rounding_fixture_test.go` (new) -- Same const/byte-identity/semantic-then-golden shape, plus an assertion that the declared slacks really are odd. -- Rationale: an even slack silently avoids the tie the fixture exists to take, and nothing else would notice.
- `folio-go/byte_neutrality_test.go`, `folio-go/matrix_test.go`, `folio-go/missing_glyph_corpus_test.go` (**both** the fixtures table and `beyondBaselineAcceptance`), `.github/workflows/matrix.yml` (`docs=` **and** all four upload blocks) -- Register both new fixtures at every surface. -- Rationale: miss one and the fixture exists but nothing renders it; miss the `beyondBaselineAcceptance` half and the accounting identity fatals.
- `_bmad-output/specs/spec-folio/folio-format.md` -- Update `:47` (the content rule gains its 2.0 arm), `:319` (the style align row gains `justify`, marked style-only), `:245`/`:246`/`:252` (state that `columns[].align` does **not** admit it), `:333-345` (split the "closed left/center/right set" sentence now that only columns are that triple), and record at `:66-70` that align was extended and the format is 2.0. Leave the Worked example at `:624` untouched. -- Rationale: the drift tests are key-name based and would not catch any of this, so the doc can only be kept honest deliberately.
- `_bmad-output/implementation-artifacts/deferred-work.md` -- **Close DW-24**, recording the grep command, its verbatim output at the closing revision, the reconciliation against the hand-list (all six anchors drifted; `table_render.go:966` was missing), and the falsifier result. **Amend DW-25's scope** to record that word-grained justified fragments make the cumulative browser-side fragment cap reachable far sooner, with the measured worst case; owner remains Story 7.4. -- Rationale: D-7.1.4's own precedent — the scope amendment lands in the story that discovers it, because a deferral known to be wrong is worse than one merely open.

**Acceptance Criteria:**

- Given the closed-set split, when the whole suite runs, then no code path validates a style align against the column set or a column align against the style set, and each rejection message lists exactly the members of the set that rejected it.
- Given DW-24's falsifier, when `textValignOffset` is multiplied by zero, and separately when each rounding site the re-derived enumeration returns is changed from half-to-even to truncation, then the golden suite goes **red at every one of those sites**. While the item was open the suite stayed green under both; that must now be false, and the closing note must record the result per site.
- Given the two new fixtures, when they are removed from any one of their five registration surfaces, then some test or workflow step fails — no surface is optional.
- Given the full corpus, when the goldens are rendered, then all six recorded digests are **measured** unchanged and reported as measured, not assumed.
- Given a justified line, when its pieces are drawn, then the distributed amounts sum to the slack exactly and the last piece's right edge equals the element's declared right edge exactly.
- Given the same justified document rendered on all four targets, when the hashes are compared, then they are identical.
- Given a justified element, when the canvas projection and the PDF run path are compared, then fragment count, text and X agree per line and per fragment, and no browser justification is used anywhere.
- Given the working tree at the end of the story, when it is inspected, then the repository-root `README.md` is byte-identical to its committed state and appears in no commit.

## Spec Change Log

## Review Triage Log

### 2026-08-30 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 5: (high 1, medium 1, low 3)
- defer: 4: (high 0, medium 2, low 2)
- reject: 11: (high 0, medium 2, low 9)
- addressed_findings:
  - `[high]` `[patch]` `TestJustifiedLinePiecesRemainderRule` never called production code — it re-transcribed `base`/`remainder` in the test body and asserted that against its own literals, so it stayed green for ANY implementation of the shipped rule, leaving the I/O matrix row "Remainder placement | slack 7, gaps 3 | 3, 2, 2" with no real coverage. Rewritten to shape real text, drive `justifiedLinePieces`, and recover each gap's granted amount from the returned piece offsets; covers 7/3, 6/3, 2/3 (a gap legitimately gets nothing), 5/2, single-gap and zero-slack. Independently red-proofed by reversing the remainder to the last gaps — three subtests fail (`granted [2 2 3], want [3 2 2]`) — and `text_alignment.go` restored byte-identical afterwards.
  - `[medium]` `[patch]` `TestAlignmentRoundingSlacksAreOdd`'s discriminating guard was vacuous: `ScaleRound(odd,1,2) % 2 != 0` can never fire, and the case that actually matters — a slack ≡ 1 (mod 4), which halves DOWN to the number a truncating `slack/2` also produces — was only `t.Logf`'d and passed. Three documents (DW-24's closing note, the fixture README, the template comment) claimed `slack ≡ 3 (mod 4)` was asserted "rather than left to luck"; it was not. Now asserts `slack % 4 == 3` plus `ScaleRound(slack,1,2) == slack/2 + 1`. Independently red-proofed by moving one box to a slack of 25,021 (≡ 1 mod 4) — fails with the intended message — then reverted byte-identically.
  - `[low]` `[patch]` The three table-cell align `default:` arm comments claimed they caught "any value the load-time closed-set check already rejected". False after the split: `justify` cascades in through `alignFallback` and is drawn at the start edge. Comments corrected at `table_render.go:701`, `:1044`, `:1228`, and `TestTableCellsCascadedJustifyIsDrawnAtTheStartEdge` added to pin that fallback (sha256 equal to the `left` render, with a `center` leg as the vacuity guard). No load error added and no justified cells implemented — both are forbidden by the intent contract.
  - `[low]` `[patch]` `justifyTemplateJSON`'s doc comment claimed "four justified text elements … plus a control" covering the ragged, no-width and overflow cases; the constant holds two elements and covers none of those. Comment rewritten to describe what it actually contains and what actually covers those rows.
  - `[low]` `[patch]` The `{{page}}` slot-straddle located error named only a "face-segment or line boundary"; a justified piece boundary is now a third way to reach it. Message extended so the diagnostic tells the author the truth.

### 2026-08-30 — Review pass (Thai scope amendment)
- intent_gap: 0
- bad_spec: 0
- patch: 6: (high 0, medium 4, low 2)
- defer: 3: (high 0, medium 1, low 2)
- reject: 11: (high 0, medium 0, low 11)
- addressed_findings:
  - `[medium]` `[patch]` The AC-TH2 fixture's atomic run was documented in four places as "a given name the shipped dictionary does not cover". Verified false: `กานต์`, a suffix of `ณัฐกานต์`, is line 3084 of `internal/text/wordlist/words_th.txt`, and the greedy matcher does propose a break at rune 3 — D-2.1.9's both-sides-coverable filter (`internal/text/tileable.go`) then withdraws it, because the preceding `ณัฐ` cannot be tiled by dictionary entries at all. The asserted behaviour was always right (the precondition is computed from production `text.Opportunities`/`text.Dictionary()`, not from a literal); only the stated reason was wrong, and a reader checking it by grepping the wordlist would have concluded the test was broken. Corrected in the fixture README, the template const comment, the test comments, the `t.Fatalf` message and this spec, with the real mechanism named after it was verified rather than assumed.
  - `[medium]` `[patch]` `justifiedThaiAssertGeometry`'s discriminating relation — `e1 runs > e2 runs`, the one assertion that a silent ragged-left fallback for Thai cannot satisfy — rested on an unasserted premise: that the control element e2 carries the same string in the same box at the same size as e1. Nothing checked it, so a drifted control would have left the relation green and meaningless. Added `justifiedThaiControlPremise`, asserting value, declared width and font size equality through one parse path, called immediately before the relation. Red-proofed by pointing the comparison at e3.
  - `[medium]` `[patch]` Nothing asserted that a justified line's pieces still spell the line. e1's lines are split into 8/7/6/10 separately positioned runs, and splitting Thai per word is exactly where a combining mark could be dropped or reordered across a piece boundary — a failure the golden digest reports only as "hash mismatch". AC-TH1 now concatenates each line's piece texts in ascending x and compares against the control element's single-run line, so the expected string comes from production rather than a hand transcription. Red-proofed by dropping the first piece.
  - `[medium]` `[patch]` The two ragged lines' stated causes were claimed in comments ("only the break-kind field can answer this", "only the line INDEX can answer") but asserted nowhere — the tests checked only "one piece at the natural start edge", so a build that set line 2 ragged for the wrong reason stayed green. This is the same defect class as the vacuous remainder test the previous pass found. Added `TestJustifiedThaiRaggedLinesAreRaggedForTheReasonsClaimed`, asserting line 2's `wrappedLine.endedBy == text.BreakMandatory` and non-final index, line 5's final index and `BreakOptional`, and that BOTH hold interior break opportunities so the zero-gap condition is not what made them ragged. Measured: line 2 mandatory with 4 interior ops, line 5 optional with 2. Red-proofed by swapping the rows and by forcing the interior count to zero.
  - `[low]` `[patch]` Two hand-maintained-table walks could panic instead of reporting: `wantCounts[i]` index-out-of-range if the run-count and baseline tables ever disagree in length (which would kill the whole test binary, including every matrix leg that calls this as its guard), and `slack / gaps` divide-by-zero for a `justifiedThaiLines` row authored with `gaps: 0` and no ragged reason. Both now fail with a message naming the inconsistent tables.
  - `[low]` `[patch]` The comment deriving the fourteen baseline literals mixed units — every constant in millipoints, `y` in points — and it is the sole record of how those literals were obtained. Rewritten so the arithmetic resolves as written.

**Not addressed, and why.** Three findings were routed to `defer` and recorded in frontmatter: the absent human Thai sign-off for the first golden to insert inter-word space into continuous Thai (only the owner can commission a Thai reader; an agent writing that record would be fabricating an attestation), the absence of any Thai instance in the canvas parity claim (7.4 owns that surface), and `verticalOffsetError` being untested (pre-existing). Eleven were rejected, the notable ones being: the per-leg matrix guard not checking a right edge (true, and identical to the Latin precedent this amendment names as its standard), the new limitation not being promoted into `deferred-work.md` (the workflow's `defer` route is frontmatter; promotion to a DW number is the closer's), and `epics.md` not carrying AC-TH1/2/3 (the amendment lives in this spec by the orchestrator's own design).

**One finding this pass could not act on.** AC-TH2's own wording, inside the owner's scope amendment, directs the implementer to "pick a Thai run the dictionary does not cover" — the same framing corrected in the first patch above. The behaviour it asks for is right and is delivered; the stated test for it is not. That text is the orchestrator's amendment, not this workflow's to rewrite, so it is reported rather than edited. Suggested rewording: "pick a Thai run the segmenter proposes no interior opportunity inside".

## Design Notes

**Why the slack is based on the summed piece widths rather than the packer's line width.** Splitting a line into pieces means each piece's advance is rounded on its own, and a sum of roundings is not the rounding of a sum — so a justified line positioned from `ln.width` could miss the declared right edge by a millipoint or two. `measureRuneRange`'s doc comment states it *is* `positionSegments`' cursor arithmetic, so measuring the pieces with the same function the packer already uses and taking `slack = boxWidth − Σ pieceWidths` makes the right edge land exactly, with one derivation and no new measurement path. Overflow detection keeps reading `ln.width`, so FR44 is untouched.

**Why justification splits pieces instead of widening spaces.** Interior spaces are real shaped glyphs inside one face-segment run, and the PDF emitter has no word-spacing operator available (under Identity-H, `Tw` applies only to single-byte code 32, which never occurs in a two-byte CID stream). Inflating advances in place is worse than unavailable: the shaped glyph slice is an aliased view shared by every line of the element, so mutating it would corrupt the others. Splitting also happens to be what the canvas requires — the browser lays out a fragment's own string, so only a real fragment boundary can move a word. One mechanism satisfies both producers, which is why AD-17 selects it rather than leaving a choice.

**Worked remainder.** Slack 7 millipoints across 3 gaps: `base = 2`, `r = 1`, so the gaps take 3, 2, 2 in reading order and sum to 7. Slack 6 across 3: 2, 2, 2. Slack 2 across 3: `base = 0`, `r = 2`, so 1, 1, 0 — a gap may legitimately receive nothing.

**Why the three ragged conditions stay three.** The break-kind field answers "which break ended this line"; the index answers "is this the last line"; the gap count answers "is there anywhere to put slack". A single boolean collapsing any two would make the third case re-derivable wrongly, which is the hazard D-7.1.5 created a named kind rather than a bool to prevent.

## Verification

7.3's correctness is byte-identity-shaped, so it carries the heavy tests regardless of the per-epic cadence (D-R7.1). Report measured pass/fail counts, never "green".

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expected: **exactly ONE** failure, `TestCorpusMeetsP6ExerciseFloors` (P6g got 7, need >=20), the mandated permanent red. Never touch it or the P6g floor. Anything else red is a defect.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean.
- `gofmt -l folio-go` -- expected: no output.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` -- expected: run **once per leg** with `FOLIO_MATRIX_TARGET` set (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`; the list is `matrix_test.go:69-74`). **Unset, this test logs "asserts NOTHING" and a no-op is not a pass** — name the legs that ran.
- `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` -- expected: pass; the all-four-in-one-process local gate.
- `cd lint && go test ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expected: pass, 214 tests before this story plus whatever this story adds (4 pre-existing `only-export-components` lint warnings are not a regression).

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` fails under `-tags=matrix` because `fontTools` is not installed here; `lint/internal/rules/licencegraph_test.go` is not gofmt-clean (DW-23, owned by Story 15.2).

**Manual checks:**
- Re-run `grep -rn 'ScaleRound(.*1, 2)' folio-go --include='*.go' | grep -v _test` at the closing revision and paste the output into DW-24's closing note. Close against **that**, never against the hand-list.
- Measure and record the worst-case cumulative fragment count a realistic justified paragraph projects, against the browser-side cap of 512, and record it in DW-25's amended scope.

**Delivery Log obligations (for the closer):**
- All six existing corpus digests reported as **measured unchanged**, plus the two new digests recorded.
- DW-24 reported as **closed**, with the re-derived enumeration and the per-site falsifier result — not as deferred.
- The format version reported as `2.0` with confirmation that 1.0 and 1.1 documents still serialize unchanged.

## Auto Run Result

### Dispatch 1 — plan only (2026-08-30)

Status: `ready-for-dev`
Blocking condition: none
Baseline: `0cd9491399fd679629c1862d518bb968bfaad95f`, tree clean, branch `main`
Commits: none (plan-only dispatch; no code written)

**Route.** Classic-intent dispatch, epic-story path. `epic-7-context.md` cache was valid (correct header, no planning-artifacts file newer) and was loaded as primary planning context; no raw planning docs were read. Previous-story continuity loaded from `7-2-set-the-space-between-a-paragraph-s-lines.md` (`status: done`, the highest done story below 7.3); no `in-review` spec exists for Epic 7. No `project-context.md` exists in this repository.

**Warnings.** `multiple-goals` carried from step 1: the justify feature and DW-24's corpus closure are separably shippable. D-7.1.4 weighed and accepted that cost when it made 7.3 DW-24's owner ("a criterion that yields to a budget stops being a criterion"), so this is recorded, not split and not blocked. `oversized` added at step 6.

**No intent gap.** Four candidate forks were examined and all four were resolved rather than escalated:

1. *How justification reaches the page* — widening drawn spaces versus splitting the line into positioned pieces. Selected by architecture, not preference: the PDF emitter has no usable word-spacing operator under Identity-H, the shaped glyph slice is an aliased view shared across every line of the element, and the canvas paints a fragment's own string, so only a real fragment boundary can move a word. AD-17 plus the existing per-fragment parity test therefore admit one mechanism.
2. *Which interior opportunities count as gaps* — the acceptance criterion says "break opportunities" unqualified, and the dispatch's own underflow rule presupposes counting them generally. All interior opportunities count, whatever their kind. An interior opportunity is never mandatory, because a mandatory break always ends its line.
3. *The zero-gap case the acceptance criteria do not cover* — stated as the dispatch directed: a line with no interior opportunities is set at its natural start edge, exactly like a last line. No disagreement to flag.
4. *Exactness of the justified right edge* — piece-wise rounding is not the rounding of the sum, so the slack basis is the summed piece widths rather than the packer's line width. This makes the edge land exactly while keeping one derivation, because `measureRuneRange` is documented to be `positionSegments`' own cursor arithmetic.

**DW-24 planned as an acceptance criterion, not a deferral,** per D-7.1.4 and D-7.2.6. The enumeration was re-derived by grep at the baseline rather than read off the hand-list, and **the hand-list had rotted a second time**: all six recorded anchors have drifted (`text_alignment.go:74`→`:81`; `table_render.go:687`→`:705`, `:698`→`:716`, `:1017`→`:1038`, `:1193`→`:1214`; only `text_alignment.go:56` still matches), and a **seventh** unexercised branch the entry never listed was found — `table_render.go:966`, an integer `slack / 2` reachable only through a table element's `valign: middle`. The entry's own claim that the six were "confirmed by grep at those exact lines" is false at HEAD. The closing note must re-run the grep at the closing revision and close against that.

**Corpus census measured, not assumed:** 16 `left`, 8 `right`, 0 `center`, 0 `valign` across every `fixtures/*/input.folio`, all 24 declarations confined to the four statement fixtures. The one `center` in `worked-example.json` is a serializer round-trip golden that is never rendered, so it is not matrix coverage. DW-24's stated hazard is confirmed intact.

**A canvas-bound interaction was found and deliberately left to Story 7.4.** Word-grained fragments multiply the projected fragment count by words-per-line. Go bounds fragments **per line** (512), while the browser-side validator bounds them **cumulatively per component** (512), so a long justified paragraph reaches the browser cap first and blanks the whole projection. That is DW-25's subject and 7.4 owns it, and the dispatch fences this story off from widening it. The spec therefore requires the reachability to be measured and recorded as an amendment to DW-25's scope — following D-7.1.4's precedent that a scope amendment lands in the story that discovers it — and makes "the canvas requires a bound to widen" a HALT condition rather than a licence.

**Two tests will go red by design and are planned as work, not defects:** `TestHigherMajorIsLoadError` uses `withVersion("2.0")` and must move to `"3.0"` once 2.0 is loadable, and `TestContentVersionNeverExceedsTheLibraryCeiling`'s hand-maintained version enumeration goes vacuous unless the new constant is added to it.

**Nothing was implemented.** No source file, fixture, workflow or deferred-work record was modified by this dispatch; the only file written is this spec.

### Dispatch 2 — implement, review, commit (2026-08-30)

Status: `done`
Blocking condition: none
Baseline: `20ccefaaa1ed2860ed68289354a1345ce678885c`, tree clean, branch `main`

**Summary.** `justify` is now a member of the *style* alignment vocabulary only, the closed set having been split in two at its declaration so that `columns[].align` cannot be widened by accident. A justified line is placed by one shared rule, `justifiedLinePieces`, called identically by the PDF line loop and the canvas paint projection, distributing slack across the line's interior break opportunities by an integer-exact, ordered remainder rule with no float and no second rounding site. Extending a closed set is a MAJOR under D-1.4.12, so the format ceiling moves to 2.0 and the content-derived version rule became a maximum over every attachment point rather than a first-hit return. DW-24 is closed with a second golden fixture, and DW-25's scope is amended with a measured reachability result.

**Files changed.**
- `folio-go/internal/template/closedsets.go` — one `closedAligns` map replaced by two ordered token slices and two maps built from them; `IsStyleAlign` exported as the single source.
- `folio-go/internal/template/parse_bands.go` — each align check points at its own set; both hand-written literals replaced by messages derived from the relevant ordered slice.
- `folio-go/internal/template/version.go` — `SupportedMajor` 2, `SupportedVersion` "2.0", new `majorFeatureVersion`; `styleNeedsMinorVersion` became a rank and `versionRequiredByContent` takes the maximum. Ceiling doc comment extended without weakening D-7.2.1's correction.
- `folio-go/text_alignment.go` — the justification rule and the extended file doc comment naming the three independent ragged conditions.
- `folio-go/render.go`, `folio-go/page_setup.go` — per-piece placement at the two call sites that must stay one line apart; unjustified path unchanged byte-for-byte.
- `folio-go/component_commands.go` — the style align arm, previously accepting any string with no validation at all, now validates through the shared style set; the column arm untouched.
- `folio-designer/src/engine-protocol.ts` — component align validator and type admit `justify`; `TableColumn.align` stays the triple.
- `folio-go/table_render.go` — the three cell-align `default:` comments corrected; no behaviour change.
- New: `fixtures/justified-text/`, `fixtures/alignment-rounding/`, their template/data consts and fixture tests.
- Registration: `byte_neutrality_test.go` (incl. `declaredEpic2GateObligations`), `matrix_test.go`, `missing_glyph_corpus_test.go` (both the corpus table and `beyondBaselineAcceptance`), `.github/workflows/matrix.yml` (`docs=` plus all four upload blocks), `render_test.go` subprocess selectors.
- Docs: `_bmad-output/specs/spec-folio/folio-format.md`; `_bmad-output/implementation-artifacts/deferred-work.md` (DW-24 closed, DW-25 scope amended).

**Review findings breakdown.** 5 patches applied (1 high, 1 medium, 3 low); 4 items deferred (2 medium, 2 low, recorded in frontmatter `deferred`); 11 rejected. No intent gap and no bad-spec loopback: the two findings that looked like scope defects — that a table's `style.align`/`headerStyle.align` admits `justify` and renders it as `left`, and that a load error "should" be added — are the intent contract's own explicit requirements (it directs that `headerStyle.align: "justify"` load and raise the document to 2.0, and forbids implementing justified table cells), so they were routed to defer and reject respectively on the contract's authority rather than escalated.

**Follow-up review recommendation:** `true`. Patched findings by severity: high 1, medium 1, low 3. Any high severity sets the flag; the score `3 × 1 + 1 × 3 = 6` also clears the threshold of 5.

**Verification performed (measured, not assumed).**
- `cd folio-go && go test -count=1 ./...` — **1479 pass, 5 skip, 2 fail entries constituting exactly ONE distinct failure**: `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest (got 7, need >=20), the mandated permanent red. Untouched. Nothing else red.
- `cd folio-go && go vet -tags=matrix ./...` — clean, exit 0.
- `gofmt -l folio-go` (run from the repository root) — no output.
- `cd lint && go test ./...` — 4 packages ok, 0 fail.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — typecheck clean; lint exactly the 4 pre-existing `only-export-components` warnings, 0 errors; **30 test files, 215 tests passed** (214 before this story, +1 added).
- `TestTargetRenderHash` — run once per leg with `FOLIO_MATRIX_TARGET` actually set, so none was the unset no-op that "asserts NOTHING": **`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` all pass**, all four agreeing on both new documents.
- `TestCrossTargetByteIdentity` — pass (21.0s), the all-four-in-one-process local gate.
- **All six recorded corpus digests measured unchanged**, by direct `shasum`, not inferred from a green suite: `statement-1` 76,744 B `114df1d6…`; `statement-5` 127,363 B `70dce051…`; `statement-20` 269,884 B `56bfbbd9…`; `statement-50` 555,829 B `5d090b0f…`; `mandatory-break` 56,681 B `7cf743de…`; `line-spacing` 57,770 B `de212115…`. Two new: `justified-text` 59,894 B `6da3b12e…`; `alignment-rounding` 61,346 B `986400a1…`.
- Matrix test audit: every I/O matrix row is covered by a test that ran and passed. Two gaps found and closed during this dispatch — the "Both conditions at once" row had no covering test at all, and the "Remainder placement" row's test did not call production code.
- Manual check 1 — the DW-24 grep was re-run at the closing revision and its output recorded verbatim in the closing note. The set of sites is identical at both revisions; only the two `text_alignment.go` anchors moved (`:56`→`:85`, `:81`→`:110`) because this story inserted the justification rule above them.
- Manual check 2 — the canvas fragment-cap reachability was measured and recorded as an amendment to DW-25's scope. No bound was widened.

**Residual risks.**
- `fixtures/alignment-rounding/` depends on the shipped Noto Sans metrics for its `slack ≡ 3 (mod 4)` property. A face change would redden `TestAlignmentRoundingSlacksAreOdd` loudly rather than silently degrading the fixture's discriminating power — which is now a real assertion rather than the vacuous one this review found.
- The four deferred items in frontmatter, in particular the unguarded `valign` command arm and the justified-table-styling scope question.
- `followup_review_recommended: true` — a high-severity patch (a test that could not fail) was applied in this pass.

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` fails under `-tags=matrix` when `fontTools` is absent; `lint/internal/rules/licencegraph_test.go` is not gofmt-clean (DW-23, owned by Story 15.2).

### Dispatch 3 — the owner's Thai scope amendment (2026-08-30)

Status: `done`
Blocking condition: none
Baseline: `8fed42ffc5f9f5f9acf3c5f1277f9ff6616ac0ad`, branch `main`

**Summary.** AC-TH1/AC-TH2/AC-TH3 only. No justification behaviour, break rule or format field
changed: `folio-go/text_alignment.go` is byte-identical to its committed state, and so is every
other production file. What this dispatch adds is the coverage the amendment says was missing —
`fixtures/justified-thai/`, a golden whose justified content is Thai, plus the tests that make Thai
justification falsifiable.

**Verified BEFORE any byte was recorded** (the scope fence's own requirement — if Thai did not
justify, the instruction was to HALT rather than change the rule). Rendering the amendment's own
cases at the baseline, `justifiedLinePieces` distributes over the dictionary-derived opportunity
list unchanged: a spaceless Thai paragraph in a 220 pt box packs into six lines of which four are
justified into 7, 6, 5 and 9 interior gaps, every one of them a wordlist seam, with each line's last
piece ending on 220,000 mp exactly.

**Files added.**
- `fixtures/justified-thai/` — `input.folio` (three elements, `"version": "2.0"`), `README.md`,
  `expected.json`, `expected.pdf`.
- `folio-go/justified_thai_template.go` — the template const kept byte-identical to `input.folio`,
  plus the empty data const, following `justified_text_template.go`.
- `folio-go/justified_thai_fixture_test.go` — semantic acceptance, AC-TH1, AC-TH2, the version
  declaration, and the golden LAST (D-000.22).

**Files changed — registration only.** `folio-go/render_test.go` (a fifth subprocess selector),
`folio-go/matrix_test.go` (capture + per-leg feature guard + `matrixDocuments`),
`folio-go/byte_neutrality_test.go` (`goldenDigestRecord` + `declaredEpic2GateObligations`),
`folio-go/missing_glyph_corpus_test.go` (corpus table + `beyondBaselineAcceptance`),
`.github/workflows/matrix.yml` (`docs=` + all four upload blocks).

**What the fixture covers.**
- `e1` — `align: "justify"`, two Thai paragraphs, **not one space character** in the value, so every
  interior opportunity is a dictionary seam (asserted, not assumed: the test fails outright if a
  space appears). Six lines: four justified (gaps 7/6/5/9, slacks 13,893 / 12,771 / 7,942 / 31,757,
  remainders 5/3/2/5 — all non-zero, three of the four larger than half the gap count), one ragged
  by mandatory break and not last, one ragged as the last line.
- `e2` — the control: same string, chain, size and box, no align. One run per line at the left edge.
- `e3` — AC-TH2. `"ณัฐกานต์ ปฐพี"` in a 50 pt box. The segmenter proposes **zero** break
  opportunities strictly inside its first line `"ณัฐกานต์"` — measured through the production
  `text.Opportunities` against the shipped `text.Dictionary()`, not assumed; that line is justified,
  is not the last line, was not ended by a mandatory break, and leaves **9,179 mp of positive
  slack** — each of those exclusions asserted — and is still set at the natural start edge. The
  zero-opportunity precondition is a `t.Fatalf`, so the case cannot go vacuous if the segmenter's
  answer for this run ever changes.

  *The reason is not "the wordlist lacks the name", and the docs and comments that said so have been
  corrected.* `words_th.txt` does contain `กานต์`, a suffix of the run: the greedy matcher matches it
  and proposes a break in front of it, and D-2.1.9's both-sides-coverable filter
  (`internal/text/tileable.go`) then withdraws that proposal because the preceding stretch cannot be
  tiled by dictionary entries at all. What the fixture asserts is the net answer the justification
  rule actually reads — no interior opportunity — whatever entries the wordlist happens to hold.

**Verification performed (measured, not assumed).**
- `cd folio-go && go test -count=1 ./...` — **1486 pass, 5 skip, 2 fail entries constituting exactly
  ONE distinct failure**: `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest
  (got 7, need >=20), the mandated permanent red. Untouched. Nothing else red.
- `cd folio-go && go vet -tags=matrix ./...` — clean, exit 0.
- `gofmt -l folio-go` — no output.
- `cd lint && go test ./...` — 4 packages ok, 0 fail.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — typecheck clean; lint
  exactly the 4 pre-existing `only-export-components` warnings, 0 errors; **30 test files, 215 tests
  passed** — unchanged, as the amendment adds no browser-side behaviour.
- `TestTargetRenderHash` run once per leg with `FOLIO_MATRIX_TARGET` actually set: **`darwin/arm64`,
  `linux/amd64`, `linux/arm64`, `js/wasm` all pass, and all four report the SAME digest for
  justified-thai**, `58ca4777…`, 15,079 bytes — AC-TH3's four-target identity.
- `TestCrossTargetByteIdentity` — pass (26.8s).
- **All eight previously recorded corpus digests measured unchanged**, by direct `shasum`, not
  inferred from a green suite: `statement-1` `114df1d6…`; `statement-5` `70dce051…`; `statement-20`
  `56bfbbd9…`; `statement-50` `5d090b0f…`; `mandatory-break` `7cf743de…`; `line-spacing`
  `de212115…`; `justified-text` `6da3b12e…`; `alignment-rounding` `986400a1…`. No existing fixture
  file was touched.

**Red-proofs (each mutation reverted byte-identically, verified by sha256).**
1. *The failure this fixture exists to catch.* Making `justifiedLinePieces` skip zero-width
   opportunities (`op.LineEnd == op.NextStart`) — a script-shaped bug that justifies Latin correctly
   and silently sets Thai ragged left — leaves **every `TestJustifiedText*` assertion green** and
   reddens the Thai ones: four lines drawn as 1 run instead of 8/7/6/10, plus
   "the justified element draws 6 runs and the unaligned control 6". Before this fixture, no
   recorded byte in the repository could tell that build from a correct one.
2. *The ordered remainder, over dictionary gaps.* Reversing the remainder to the last gaps reddens
   every justified Thai line (`gap 0 is 1984, want 1985 …`).
3. *AC-TH2's non-vacuity.* Simulating the segmenter proposing an interior opportunity inside the
   atomic run fires the precondition `t.Fatalf` with its intended message, so the case cannot pass
   by having quietly stopped being AD-25's atomic run.

**Registration red-proofed too:** dropping `justified-thai` from `matrix.yml`'s `docs=` list reddens
`TestMatrixDocumentSlugsAreRegisteredInCI`, which also pins the four per-target upload paths.

**Finding, reported rather than worked around.** The amendment's "known pre-existing limit" was
characterised: every Thai codepoint renders in isolation (91/91 over U+0E01..U+0E5B), but any
sequence that **stacks two marks above a base** (for example `ท` + `ั` + `้`) fails closed with
`CID N carries a non-zero vertical offset`. It fails identically under `align: left`, so it is
independent of justification and is not a justification defect; it is `internal/pdf`'s deliberate
AC6 fail-closed branch. It is recorded as a new frontmatter `deferred` entry because nothing in the
tree tracked it and it constrains what Thai any future fixture can contain. Thai that both renders
and exercises AC-TH2 was found, so no HALT condition was reached.

**Review patches applied to this dispatch (test and documentation only; no production file touched,
no golden re-recorded).**
- *A factual claim corrected in four places.* "A given name the shipped dictionary does not cover"
  was false — see the e3 bullet above. The prose in `fixtures/justified-thai/README.md`, in the
  template const's comment, in the test's comments and messages, and here now states the property
  that is actually asserted.
- *The control's premise is asserted.* `justifiedThaiAssertGeometry`'s `e1 <= e2` relation is
  evidence about justification only if e2 carries the same string in the same box at the same size.
  `justifiedThaiControlPremise` now asserts all three with `t.Fatalf` (and that e1's declared width
  is `justifiedThaiBoxWidthMP`), recovered from the parsed document. The function takes `*testing.T`
  as a result; `matrix_test.go`'s call was updated with it.
- *The justified pieces must still spell the line.* AC-TH1 now concatenates each line's piece texts
  in ascending x and compares them to the same line as the **control** sets it — an expected string
  from production, never a hand-transcribed one — for the ragged lines too.
- *Two table-walk panics turned into reported failures.* A length disagreement between
  `justifiedThaiRunsPerBaseline()` and `justifiedThaiBaselinesMP` (an index-out-of-range that would
  have killed the binary, matrix legs included) and a `justifiedThaiLines` row authored `gaps: 0`
  with no ragged reason (a divide-by-zero) now each fail with a message naming the inconsistency.
- *The ragged lines' causes are asserted.* New
  `TestJustifiedThaiRaggedLinesAreRaggedForTheReasonsClaimed` packs e1 through the production path
  and asserts line 2's `wrappedLine.endedBy` really is `text.BreakMandatory` and is not the last
  index, line 5's index really is the final one and no mandatory break ended it, and that **both**
  hold interior break opportunities — so neither is quietly AC-TH2's zero-gap case.
- *Units made explicit* in the fourteen-baseline derivation comment, which mixed millipoint
  constants with a point-valued `y`.

Each of the six was red-proofed by mutation and reverted; the whole suite's only failure remains
`TestCorpusMeetsP6ExerciseFloors` (P6g, the mandated permanent red), and all nine recorded corpus
digests are unchanged.

**Review findings breakdown (this dispatch).** 6 patches applied (0 high, 4 medium, 2 low); 3 items
deferred (1 medium, 2 low, recorded in frontmatter `deferred`); 11 rejected. No intent gap and no
bad-spec loopback: the amendment enumerates AC-TH3's completion condition — `goldenDigestRecord`,
`matrixDocuments`, the re-derived enumeration, and a digest identical on all four targets — and the
diff satisfies it exactly, so the one finding that looked like a scope defect (the missing human Thai
sign-off) was routed to `defer` on the intent's own authority rather than escalated.

**Follow-up review recommendation:** `true`. Patched findings this pass by severity: high 0,
medium 4, low 2. No high severity, but the score `3 × 4 + 1 × 2 = 14` clears the threshold of 5.

**Verification re-performed by the reviewing pass (measured independently, not taken on report).**
- `cd folio-go && go test -count=1 ./...` — **1487 pass, 5 skip, 2 fail entries constituting exactly
  ONE distinct failure**: `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest
  (got 7, need >=20), the mandated permanent red. Untouched. Nothing else red. (1486 before the
  patches; `TestJustifiedThaiRaggedLinesAreRaggedForTheReasonsClaimed` is the +1.)
- The 5 skips are the two always-on ones (`TestBrowserAuthoredRoundTripWitness`,
  `TestXrefEntriesRejectsMalformedSubprocess`) plus the three `FOLIO_HEAVY`-gated table tests, which
  were then run with `FOLIO_HEAVY=1` and all three pass. No skip conceals an I/O-matrix row.
- `cd folio-go && go vet -tags=matrix ./...` — clean, exit 0.
- `gofmt -l folio-go` from the repository root — no output.
- `cd lint && go test -count=1 ./...` — 4 packages ok, 0 fail (run uncached).
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — typecheck exit 0; lint
  exactly the 4 pre-existing `only-export-components` warnings; **30 test files, 215 tests passed**,
  unchanged, as this dispatch adds no browser-side behaviour.
- `TestTargetRenderHash` with `FOLIO_MATRIX_TARGET` actually set on each of four legs — none was the
  unset no-op that "asserts NOTHING": `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm` all
  pass, and all four report the identical `justified-thai` digest `58ca4777…`, 15,079 bytes.
- `TestCrossTargetByteIdentity` — pass (22.1s).
- **All nine corpus digests measured by direct `shasum`, not inferred from a green suite.** The eight
  pre-existing ones are byte-identical: `statement-1` 76,744 `114df1d6…`; `statement-5` 127,363
  `70dce051…`; `statement-20` 269,884 `56bfbbd9…`; `statement-50` 555,829 `5d090b0f…`;
  `mandatory-break` 56,681 `7cf743de…`; `line-spacing` 57,770 `de212115…`; `justified-text` 59,894
  `6da3b12e…`; `alignment-rounding` 61,346 `986400a1…`. The ninth is `justified-thai` 15,079
  `58ca4777…`.
- **Independent red-proof of the fixture's reason to exist**, performed by the reviewing pass rather
  than accepted on report: mutating `justifiedLinePieces` to skip zero-width opportunities
  (`op.LineEnd == op.NextStart`) — a script-shaped bug that justifies Latin correctly and silently
  sets Thai ragged left — leaves the entire Latin `TestJustifiedText*` suite **green** and reddens
  three Thai tests. `text_alignment.go` was then restored and proven byte-identical by sha256 and an
  empty `git diff`. Before this fixture, no recorded byte in the repository could tell that build
  from a correct one.

**Residual risks.**
- The absent human Thai sign-off, recorded as a deferred item. `fixtures/justified-thai/` is the
  first golden anywhere to insert visible space between Thai words, it is now defended on four
  targets in CI, and whether that reads correctly is a question no test here answers.
- The fixture's text was chosen around a hard render refusal: stacked-mark sequences (vowel-above
  plus tone-above) cannot be rendered at all, which excludes very common vocabulary. This dispatch
  characterised the limit and recorded it; it did not fix it, and it is independent of justification.
- Line breaks, gap counts and slacks are hand-stated and depend on the shipped Noto Sans Thai metrics
  and the shipped wordlist. A face or wordlist change reddens this loudly rather than degrading it
  silently.
- The four earlier deferred items from dispatch 2 stand, in particular the unguarded `valign` command
  arm and the justified-table-styling scope question.
