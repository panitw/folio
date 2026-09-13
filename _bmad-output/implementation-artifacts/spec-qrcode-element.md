---
title: 'QR Code element'
type: 'feature'
created: '2026-09-13'
status: 'done'
route: 'dispatch'
baseline_commit: '4bfc83c2163d7757049f9d1dd8ad5be411ecbe77'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/specs/spec-barcode-qr-elements/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-barcode-qr-elements/symbologies.md'
  - '{project-root}/_bmad-output/implementation-artifacts/spec-barcode-element.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Thai invoices also carry a QR code (Thai QR Payment), and a folio template cannot print one from JSON data.

**Approach:** Add a `qrcode` element (CAP-2, plus the QR half of CAP-3 to CAP-6) that reuses the barcode element's plumbing. Its bindable `value` is encoded as a QR Code at an authorable error-correction level, drawn as black filled rects, previewed on the canvas, and saved in the same format version 4.0.

## Boundaries & Constraints

**Always:**
- The pure-Go encoder lives in `internal/barcode`: ISO/IEC 18004 Model 2, byte mode over the value's UTF-8 bytes with no ECI, and the smallest version (1–40) that holds the content at the chosen level. It uses Reed–Solomon over GF(256), and chooses the mask with the lowest penalty score (ties go to the lowest mask number). No floats, no ranging over maps, and no new dependency.
- `errorCorrection` is an optional element key from the closed set `L`/`M`/`Q`/`H`; absent means `M`. Null, any other value, or the key on another element type is a load Error, and the commands refuse the same cases.
- Geometry: the symbol is square, with a 4-module quiet zone inside the box. The module width is the largest whole millipoint for which symbol plus quiet zone fits the smaller of the box's width and height. The symbol is centred on both axes.
- Shares the barcode's rules:
  - Binding goes through `bind.BindTextSpans`.
  - A `style` key and `{{page}}` are refused at load.
  - Null or empty content draws nothing.
  - `\r`, `\n` and `\\` escapes work in the designer field.
  - A bound value is drawn on the canvas as illustrative modules, with each `{{ }}` filled by `0123456789`.
  - The canvas and the render share one layout function.
- Owner decision, carried from the barcode build: failures caused by data are Warnings. That QR code is omitted and the render completes.
- A document containing a `qrcode` declares `4.0`; `SupportedMajor` and `SupportedVersion` do not change. The barcode fixture and every other existing golden hash stay unchanged.

**Never:**
- Numeric, alphanumeric or kanji modes, ECI, Micro QR, or Structured Append.
- Building a Thai QR/PromptPay payload or computing its CRC16.
- Logos, colours, or rotation.
- Changing the barcode's bars.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| EMVCo payload | Bound Thai QR Payment string, level M, 120x120pt box | Square QR that decodes to exactly the string | N/A |
| Thai UTF-8 | `"ชำระเงิน {{ref}}"`, level Q | Decodes to exactly the UTF-8 string | N/A |
| Each level | Same value at L, M, Q and H | Decoder reports that level; higher levels use a version at least as large | N/A |
| Level absent | No `errorCorrection` key | Encoded at M; saves without the key | N/A |
| Bad level | `"X"`, `null`, or the key on a `text` | — | Located load Error |
| Non-square box | 200x80pt | Symbol sized to 80pt, centred on both axes | N/A |
| Empty or null | Resolves to "" or null | Nothing drawn | No diagnostic |
| Too long, static | Static value over version-40 capacity at its level | — | Load Error `TEMPLATE_FIELD_INVALID` |
| Too long, data | Bound value over capacity | Not drawn | Located Warning `QRCODE_TOO_LONG` |
| Module below 0.5 mm | Module < 1418 mp | Drawn | Located Warning `QRCODE_MODULE_TOO_SMALL` |
| Cannot fit | Module would be < 1 mp | Not drawn | Located Warning `QRCODE_DOES_NOT_FIT` |

</frozen-after-approval>

## Code Map

- `folio-go/barcode_element.go`, `barcode_render.go` -- share the band walk, escapes, illustrative filling, static check and placement check with QR through a per-kind layout function. Keep `layoutBarcode` and the barcode golden hash.
- `folio-go/internal/barcode/` (rank 1, `stagerank.go:61`) -- new QR encoder beside `code128.go`. Integer GF(256) tables, block and capacity tables, alignment positions, BCH format bits (mask 0x5412) and version bits.
- `folio-go/internal/template/closedsets.go:22,138-173`, `model.go:414,543`, `parse_bands.go:141,215,284,621-632`, `serialize.go:354-410` (unkeyed `kv` after `value`) -- type, `errorCorrection` set/lookup (a slice, not a map range), Presence field, decode and refusals.
- `folio-go/internal/template/version.go:104-110,351,412,432` -- QR maps to the 4.0 rank; no new major.
- `folio-go/internal/template/drift_test.go:232,269,402`, `fixtures_test.go:270` -- document and add `errorCorrection` to the maximal fixture.
- `folio-go/internal/diag/diag.go:368-454`, `diag_test.go:49`, `diagnostic.go:317`, `diag_bridge_test.go:51`, `diagnostic_registry_census_test.go:251` -- three `QRCODE_*` Warning codes with witnesses.
- `folio-go/render.go:2437,2516`, `folio_expr_validate.go:78`, `stand_in_data.go:321`, `parameter_references.go:45` -- widen every barcode type check to include qrcode.
- `folio-go/component_commands.go:1013,1382-1386,1555-1600,1900-1951,2027` -- scalar bind, `errorCorrection` set/clear modelled on `align`, palette type, 72x72pt drop, starter value `Folio`.
- `folio-go/page_setup.go:313,970,1092,1203,1879`, `canvas_authored_properties.go:17-76`, `canvas_projection_wire_test.go:676` -- paint `qrcode: {moduleWidth, rects[{x,y,width,height}]}` (dark modules merged into horizontal runs per row), `qrcodeUnavailable: 'tooLong'|'doesNotFit'`, authored `errorCorrection`, wire keys.
- `folio-designer/src/component-command.ts:25`, `DataPanel.tsx:67,114`, `engine-protocol.ts:230,398,476,631-641,808-929`, `App.tsx:122,236,3490,3619,3684-3687,5035,5280,5342,5572`, `App.css:409` -- palette, glyph, CONTENT section, L/M/Q/H segmented control, `QRCodePaint`, protocol guard (exact keys, integers, square modules, rows ordered, inside the box), binding list.
- `fixtures/qrcode-payments/` (new), `render_test.go:883,1070`, `matrix_test.go:737,1999`, `byte_neutrality_test.go:494,1041`, `.github/workflows/matrix.yml:78,124,170,216,275` -- one page with QR codes at L, M, Q and H, a Thai UTF-8 value, and an EMVCo payload; registered everywhere.
- `_bmad-output/specs/spec-folio/folio-format.md` -- seven types, `qrcode` paragraph, version rule wording.

## Tasks & Acceptance

**Execution:**
- [x] `folio-go/internal/barcode/qr*.go` -- encoder and tests: RS codewords against the ISO/IEC 18004 Annex I example, format/version BCH values, capacity edges per level, mask penalty determinism -- CAP-2 core.
- [x] `folio-go/internal/template/*` -- type, `errorCorrection`, refusals, static too-long load Error, 4.0 mapping, drift and maximal fixture -- CAP-6.
- [x] `folio-go/internal/diag/*`, `diagnostic*.go`, `render.go`, shared `barcode_*` helpers -- fit, rects, Warnings, all types widened -- CAP-3 and CAP-4.
- [x] `folio-go/component_commands.go`, `page_setup.go`, `canvas_authored_properties.go`, `stand_in_data.go`, `parameter_references.go` and tests -- commands, paint, illustrative canvas, header-band render.
- [x] `folio-designer/src/*` and tests -- palette, CONTENT and error-correction controls, `QRCodePaint`, guard, binding -- CAP-5.
- [x] `fixtures/qrcode-payments/` with registrations, `folio-format.md`, `symbologies.md` -- golden and docs.

**Acceptance Criteria:**
- Given the `qrcode-payments` fixture PDF, when its rects are rasterised and decoded with zxing-cpp, then every symbol decodes to its exact resolved string at its declared error-correction level.
- Given any document without a qrcode, when rendered and saved, then its bytes, its version and every existing golden hash, including the barcode fixture, are unchanged.
- Given the designer, when a user places a QR Code, sets level H and binds its value, then the document saves `errorCorrection: "H"` at 4.0 and the canvas shows illustrative modules.

## Implementation Notes

- Encoder `internal/barcode/qr.go`: ISO/IEC 18004 tables, GF(256) exp/log arrays, Reed–Solomon, BCH format/version bits, zigzag placement and the four penalty rules (light border counted for N3). Tests pin the Annex I RS vector, format/version BCH values, byte capacity at versions 1, 2, 7, 10, 27, 40 and every version edge, mask-choice minimality with ties to the lower mask, and rect merging.
- Cross-check: 100 symbols (lengths 1–2953, all four levels, versions 1–40) decoded with zxing-cpp at the right level. segno agreed on every version but not on modules: segno writes an extra zero pad byte after the terminator, which this encoder follows ISO in not doing.
- Shared plumbing: `barcode_render.go` gains `layoutQRCode` and the per-kind `layoutCode`; `collectBarcodeRects`, `addCanvasBarcodePaint`, `canvasBarcodeIsPlaced` and `canvasContentBandHasBoundBarcode` now walk both kinds. `layoutBarcode` and the barcode golden hash are unchanged. `checkBarcodeValue` became `checkCodeValue` (static Code 128 non-ASCII, or static QR bytes over version-40 capacity at the element's level).
- The maximal fixture gained a qrcode with `errorCorrection` (now 4.0); `TestAFullyExercisedDocumentIsNotVersionINFLATED` strips qrcode elements the way it strips the other ceiling keys.
- Fixture `qrcode-payments` sha256 `a14dc0aec702a28ef741fca2ebb09c79a0edd7122d181c6f305ce378e03bd84a`; its 2007 `re f` rects, rasterised and decoded with zxing-cpp, read all four strings at L, M, Q and H.
- Not run: the four-target matrix legs (owed at the next matrix gate).

## Spec Change Log

## Review Triage Log

Pass 1 (blind-hunter, edge-case-hunter, verification-gap):

| # | Finding | Verdict | Evidence | Route |
|---|---|---|---|---|
| B1 | No full-symbol check against an outside encoder; module tests use `EncodeQR` as their own oracle | medium | The encoder is correct: all four fixture symbols rebuilt from the PDF match python-qrcode module for module (e1 v8 mask 2, e2 v11 mask 2, e3 v4 mask 0, e4 v1 mask 3). But no Go test pins that independently | patch |
| B2 | zxing-cpp scan script lives only in a scratchpad | low | Acceptance criterion 1 and the fixture README rely on an uncommitted script | patch |
| B3 | Byte mode without ECI means strict ISO-8859-1 readers garble Thai | low | ECI is excluded by the frozen intent (Never: ECI) | reject |
| B4 | Fixture semantic test never checks symbol position or centring | low | `assertRectsEncode` takes the origin from the first rect; the comment and README claim centring | patch |
| B5/E1 | Header-band test has a dead check and no per-page origin check | low | `if i > 0 && rects[0].Y != 0 && len(rects) == 0` can never fire | patch |
| B6 | No test that raising the level pushes static text over capacity | low | The command path re-validates (`component_commands.go:1316`), but no test covers it | patch |
| B7 | Any encoder error maps to `QRCODE_TOO_LONG` | false | The only other error is an unknown level, which the loader and commands already refuse; unreachable | reject |
| B8 | Canvas can say tooLong for the illustrative fill of many placeholders | low | Needs 100+ placeholders; the fix adds a branch | reject |
| B9 | Error-correction control shows no segment for default M; explicit "M" is a second spelling | low | The panel note states "None selected is M"; "M" is a legal member of the closed set | reject |
| B10 | `isQRCodePaint` does not check x alignment and has no size cap | low | Admits Go-computed geometry that is always aligned; the fix adds guard logic | reject |
| B11 | Stale "six fixed palette components" and "NEW seventh element kind" wording | low | Direct corrections | patch |
| B12 | `render.go` comment still says barcode only; helpers keep barcode names; Code Map lists `render.go` | low | The comment is a direct correction; renames are churn; the Code Map is this build's spec | patch (comment) / reject |
| B13 | `layoutCode` comment says canvas paint calls it, but canvas calls `layoutQRCode` | low | Direct comment correction | patch |
| B14 | QR Code / QR code / qrcode spellings differ | low | Follows existing per-context naming (the barcode kind is lowercase too) | reject |
| B15 | `drawCodewords` comment "Remainder modules stay light" is wrong after masking | low | Direct comment correction | patch |
| E2/E3 | Exported `QRMaxBytes`/`QRVersionFor`/`QRFormatBits` panic on out-of-range input | low | Internal package; every caller passes validated values | reject |
| V1 | No test checks the JSON Go emits for `qrcode`/`qrcodeUnavailable` | medium | Pre-verified by layer: the wire-key fixture has no qrcode, so a tag rename would pass every test | patch |

## Design Notes

Byte mode only keeps segmentation trivially deterministic. Thai QR Payment strings are short enough that the numeric and alphanumeric modes would save at most a version or two, and the spec's "may" permits omitting them.

## Verification

**Commands:**
- `cd folio-go && gofmt -l . && go vet ./... && go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` -- expected: no gofmt output, all tests pass.
- `cd lint && go test -count=1 ./...` -- expected: pass.
- `cd folio-designer && npm run lint && npm run typecheck && npm test` -- expected: pass.
- `uv run --with zxing-cpp --with pillow python3 <scan script>` on `fixtures/qrcode-payments/expected.pdf`, adapting `/private/tmp/claude-501/-Users-panitw-Projects-folio/92c7a974-8111-457f-9788-796c179c3d68/scratchpad/scan.py` (it rasterises the `re f` rects) -- expected: every symbol decodes to its exact string with the right level.
