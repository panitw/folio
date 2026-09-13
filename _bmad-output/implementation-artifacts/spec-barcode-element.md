---
title: 'Barcode element (Code 128)'
type: 'feature'
created: '2026-09-13'
status: 'done'
route: 'dispatch'
baseline_commit: '3bf1aad3a04153720f5649632179bc2f78d595ee'
review_loop_iteration: 1
context:
  - '{project-root}/_bmad-output/specs/spec-barcode-qr-elements/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-barcode-qr-elements/symbologies.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A folio template cannot print the Code 128 bill-payment barcode that Thai invoices need, so the code cannot come from the JSON data.

**Approach:** Add a `barcode` element type (CAP-1, CAP-3, CAP-4, CAP-5 and CAP-6 of `spec-barcode-qr-elements`, barcode half only). Its bindable `value` is encoded as Code 128, drawn as black filled rects inside its box, previewed on the designer canvas from the engine, and saved as format 4.0.

## Boundaries & Constraints

**Always:**
- Pure-Go encoder with no floats, no maps ranged, and no new module dependency. The code-set choice is deterministic and minimal-length (ISO/IEC 15417); the modulo-103 check character is added automatically.
- The quiet zone is at least 10 modules on each side, inside the box. The module width is the largest whole millipoint for which bars plus quiet zones fit the box width. Bars fill the full box height and the symbol is centred horizontally.
- `value` binds through `bind.BindTextSpans`, the same path a text value uses. Escapes: the designer field shows and accepts `\r`, `\n` and `\\`; the command layer stores real characters, so the `.folio` JSON carries `\r` as a JSON escape.
- A document containing a barcode saves as `4.0`. Documents without one keep their version and bytes, and every existing golden hash stays unchanged.
- Every diagnostic is registered in `internal/diag`, carries the element id, and has a production trigger.
- Decision (owner): failures caused by data — an unencodable character, or content that cannot fit — are Warnings. That barcode is omitted and the render completes; one bad record never stops a print run.
- The canvas shows the real bars from Go-computed geometry, with no browser measurement (AD-17).
- Decision (owner, review pass 1): the canvas never sees data, so a barcode whose value contains `{{ }}` draws illustrative bars with its placeholders filled by sample-like characters, never stand-in or sample data. Preview shows the real code.

**Never:**
- QR Code or any other symbology.
- Human-readable text.
- Colours, rotation, or `style.background` / `style.border` / font keys on a barcode. These are refused at load and by commands.
- Any change to `rectdoc.go` or `flipY`, or lowering a version on save.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Thai bill payment | `"\|0994000123456{{suffix}}\r{{ref1}}\r{{ref2}}\r{{amount}}"`, 300x50pt box | Code 128 bars that decode to exactly the resolved string, CRs included | N/A |
| Resolves empty or null | `{{ref}}` is null, or `value` is `""` | Nothing drawn | No diagnostic |
| Absent path | `{{missing}}` | — | Located Error (existing bind rule) |
| Non-ASCII character | Resolves to Thai text or a byte > 127 | Barcode not drawn | Static value: load Error. From data: located Warning `BARCODE_UNENCODABLE`; `-strict` fails |
| Module below 0.25 mm | Module width < 709 mp | Bars drawn | Located Warning `BARCODE_MODULE_TOO_SMALL`; `-strict` fails |
| Cannot fit | Module width would be < 1 mp | Barcode not drawn | Located Warning `BARCODE_DOES_NOT_FIT`; `-strict` fails |

</frozen-after-approval>

## Code Map

- `folio-go/internal/template/model.go:403-461`, `closedsets.go:21`, `parse_bands.go:111-334` (text branch at :277 is the pattern; list at :141) -- type constant, closed set, decoding and style refusal. `serialize.go:394` already writes `Value`.
- `folio-go/internal/template/version.go:103-106,142-167,311-380,393-420` -- `barcodeVersion = "4.0"` with `rankBarcode` appended; SupportedMajor 4. Precedent: headerAlign at :365. Fix `version_test.go:24` (4.0 becomes 5.0) and `linespacing_test.go:313-345`.
- `folio-go/internal/barcode/` (new) -- Code 128 encoder returning module runs. Register in `lint/internal/rules/stagerank.go:57-81`.
- `folio-go/render.go:741-766` (bind pattern), `:2357-2439`, `element_box.go:64` -- produce `tableRectSource` bars (`pagemodel.Rect`, `HasFill`, black) for visible barcodes; keepTogether and bands apply through `elementID`.
- `folio-go/internal/diag/diag.go:71-412` -- codes, `allCodes`, dispositions; census `diagnostic_registry_census_test.go:17`.
- `folio-go/component_commands.go:1383-1407,1839,1877-1928,1961-2020,966-1013` -- create/drop allow-lists, default size, default `Value`, value/expression permission, binding; escape encode and decode.
- `folio-go/page_setup.go:258-312,948,1079,1288`, `canvas_authored_properties.go:66` -- barcode canvas paint (bars in mp, or an unavailable reason) modelled on `addCanvasImagePaint`.
- `folio-go/stand_in_data.go:311-326` -- stand-in rule like text's.
- `folio-designer/src/component-command.ts:25`, `App.tsx:122,230,3485,3566,3640-3690,5274,5331,5555`, `DataPanel.tsx:67`, `engine-protocol.ts:229-230,476,808,814` -- palette, glyph, single-line CONTENT field, hidden BOX controls, `BarcodePaint`, bindable types, protocol keys.
- `fixtures/barcode-thai-bill-payment/` (new) -- register in `byte_neutrality_test.go:101`, `matrix_test.go:1583`, `.github/workflows/matrix.yml:271`, plus a golden test.
- `_bmad-output/specs/spec-folio/folio-format.md` Elements -- barcode row, "six types", 4.0 (drift tests `internal/template/drift_test.go:232,269`).

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/internal/barcode/` -- Code 128 encoder and tests: set selection, check character, known vectors, and a test-only decoder round-trip including CR -- the core of CAP-1.
- [x] `folio-go/internal/template/*`, `lint/internal/rules/stagerank.go` -- barcode type, parse, style refusal, static non-ASCII load error, version 4.0 -- CAP-6.
- [x] `folio-go/internal/diag/diag.go`, `render.go`, `element_box.go` -- bind, encode, fit, emit rects and diagnostics -- CAP-3 and CAP-4.
- [x] `folio-go/component_commands.go`, `page_setup.go`, `canvas_authored_properties.go`, `stand_in_data.go` and their tests -- commands, escape round-trip, canvas paint.
- [x] `folio-designer/src/*` and tests (`App.test.tsx:1950`, `control-vocabulary-contract.test.tsx`, `engine-bounds-mirror.test.ts:745-770`, `engine-protocol.test.ts:571`, `canvas-selection.test.ts:21`, `DataPanel.test.tsx`) -- palette, properties, `BarcodePaint`, binding -- CAP-5.
- [x] `fixtures/barcode-thai-bill-payment/` and registrations, `folio-format.md` -- golden fixture and format documentation.

**Acceptance Criteria:**
- Given a document without a barcode, when it is loaded, saved and rendered, then its bytes and every existing golden hash are unchanged.
- Given a barcode document, when it is saved, then it declares `4.0` and round-trips byte-identically.
- Given the designer, when a user places a Barcode and edits a static value with `\r` in the content field, then the canvas bars match the rendered PDF's rects.
- Given a barcode bound with `{{ref}}` in the Data panel, when the canvas projects it, then it draws illustrative bars (never "No barcode content") and a template using `{{params.x}}` in a barcode lists `x` among runtime parameters.

## Implementation Notes

- Encoder in `internal/barcode` (stage rank 1, also in ARCHITECTURE-SPINE); render in `barcode_render.go`; load checks, escapes and canvas paint in `barcode_element.go`. Canvas and render share `layoutBarcode`.
- Static unencodable text and `{{page}}`/`{{pages}}` in a barcode value are load Errors (`TEMPLATE_FIELD_INVALID`). A bound barcode's canvas bars resolve against engine stand-in data, not the author's loaded sample.
- `version.go`: the proportional-table probe now guards `rankProportionalTable > highest`, so a later proportional table can no longer lower a higher rank; saving never lowers a loaded version, so no existing file changes.
- Verification fix: `canvas_projection_wire_test.go` record gained `barcode`/`barcodeUnavailable` (the implementation run left it red).
- Independent check: the fixture PDF's 106 bars, rasterised and decoded with zxing-cpp, read `|099400012345601\r1234567890\r0000000001\r150000` exactly.
- Not run: the four-target matrix legs (owed at the next matrix gate); payload layout still unverified against the TBA document.

## Spec Change Log

- Review pass 1, B1/B2 (intent_gap): a bound barcode resolved against stand-in data projected "" and showed "No barcode content", and mixed values drew partial codes unlike the PDF. The owner chose illustrative canvas bars and a fix in place over revert and rebuild. Amended: the frozen decision bullet, the third acceptance criterion (static values only), a new bound-canvas and runtime-parameters criterion, and Design Notes. Known-bad state avoided: a canvas that shows nothing, or a wrong partial code, for the core bind flow. KEEP: encoder, render path, diagnostics, version 4.0, escapes, fixture and its hash, and the protocol guard. The fixture decodes with zxing-cpp.

## Review Triage Log

Pass 1 (blind-hunter, edge-case-hunter, verification-gap):

| # | Finding | Verdict | Evidence | Route |
|---|---|---|---|---|
| B1 | Bound barcode `{{ref}}` shows "No barcode content" on the canvas | high | The text demand's lowest stand-in is `standInEmptyString` (`stand_in_data.go:115,169`), so `addCanvasBarcodePaint` resolves "" and projects neither paint nor reason; the preview with the loaded sample draws bars | intent_gap (with B2) |
| B2 | Canvas bars for bound values ignore the loaded sample | high | The canvas projection is data-free by design (`page_setup.go:1425`); a mixed static and bound value draws a partial code that differs from the PDF | intent_gap (with B1) |
| B3 | Canvas gives no too-small warning | low | Real, but the render Warning reaches the preview; the fix adds a UI branch | reject |
| B4 | Only `\r`, `\n` and `\\` escapes; no GS or FNC1 | false | The owner fixed the escape set in the frozen intent; GS1 is outside intent | reject |
| B5/E2 | Escaping can push a value past `maxCanvasBodyText` and fail the canvas without an element id | low | Needs a value above ~512 KiB of CR/LF/backslash; the fix adds a bound check | reject |
| B6 | CAP-4 "too long" dropped; Encode runs over huge input | false | Code 128 has no length limit, so too long surfaces as `BARCODE_DOES_NOT_FIT` (tested); any other fix edits the spec | reject |
| B7 | QR will force a second major unless both ship together | low | A process risk already recorded in `deferred-work.md`; no code defect | reject |
| B8 | No barcode tests for header/footer, keepTogether or `-strict` | medium / low | Header/footer grouped with V5; keepTogether and `-strict` run through generic elementID and warning paths already covered | V5 patch / reject |
| B9 | Encoder tie-breaks unpinned; unfinished `"123"` test comment | low | The comment is a direct correction; minimality is covered by brute force and the bars decode with zxing-cpp | patch (comment) |
| B10 | Stale binding comment above the text-only binding test in `engine-protocol.test.ts` | low | Its claim that `text` is the only scalar-binding kind is now false | patch |
| B11 | Stale "five"/"sixth kind"/HOIST/3.0-ceiling wording in tests and comments | low | Direct corrections | patch |
| B12 | Spec record names files the diff does not touch | low | The fix edits this build's spec | reject |
| B13 | `BARCODE_DOES_NOT_FIT` message units wrong | false | At the 1 mp minimum module, N modules need N mp, so the figure is right | reject |
| B14 | Fixture `re f` count is fragile; semantic test skips centring | low | The fixture draws bars only by construction; centring is covered by the unit test | reject |
| E1 | Raw CR/LF inside `{{ }}` is not escaped for the single-line field | low | Needs a hand-edited placeholder with a line break; the fix adds a branch | reject |
| E3 | A mixed selection hides border and fill for the other kinds | false | Same pattern as `line`; a mixed commit carrying `background` is refused by Go for the barcode | reject |
| E4 | The version guard changes non-barcode multi-table documents | false | Those documents declared a lower version than their content requires; the guard corrects that and stops a barcode followed by a proportional table declaring 3.0 | reject |
| V1 | `ParameterReferences` skips barcode values (`parameter_references.go:45`) | medium | Confirmed text-only gate; barcode `params.x` never reaches the Data panel | patch |
| V2 | Canvas window-count exactness untested for bound and unfit barcodes | medium | Pre-verified by layer | patch |
| V3 | `barcodeUnavailable` production untested in Go | medium | Pre-verified by layer | patch |
| V4 | No designer test for `BarcodePaint` or the barcode inspector | medium | Pre-verified by layer | patch |
| V5 | No barcode test in the page header or footer band | medium | Pre-verified by layer | patch |
| V6 | Version guard (barcode then proportional table) untested | medium | Pre-verified by layer | patch |

## Design Notes

Escapes live in the command/canvas projection, not the engine. `.folio` stores the real characters (JSON `"\r"`), so hand-editors use ordinary JSON escapes and the engine stays escape-free. The designer field is single-line, so Enter commits rather than inserting a line feed.

Illustrative canvas bars: each `{{ }}` placeholder in a barcode value is replaced by the ten characters `0123456789` before `layoutBarcode` runs. No data is resolved, so the `unresolved` reason no longer exists. The page-count exactness rule for bound barcodes is unchanged.

## Verification

**Commands:**
- `cd folio-go && gofmt -l . && go vet ./... && go test -count=1 ./...` -- expected: no gofmt output, all tests pass.
- `cd lint && go test -count=1 ./...` -- expected: pass (stage rank, float ban).
- `cd folio-designer && npm run lint && npm run typecheck && npm test` -- expected: pass.

**Manual checks:**
- Scan the barcode in `fixtures/barcode-thai-bill-payment/expected.pdf` with a phone scanner or ZXing; the decoded text must equal the fixture's resolved value.
