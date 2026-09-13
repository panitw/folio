---
id: SPEC-barcode-qr-elements
companions:
  - symbologies.md
  - ../spec-folio/folio-format.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Barcode and QR Code Elements

## Why

Thai invoices and utility bills carry a Code 128 bill-payment barcode and a QR code so payers can settle them at ATMs, bank counters, convenience stores and in mobile banking apps. Today a folio template cannot produce either. Authors must paste in pre-rendered images, which breaks folio's JSON-first model: the code cannot come from the data. This work adds `barcode` and `qrcode` as element types. It supersedes the "out of scope" listing in PRD §5.3 and in `spec-folio/SPEC.md`.

## Capabilities

- **CAP-1**
  - **intent:** An author can place a `barcode` element whose bindable content renders as a Code 128 symbol.
  - **success:** A fixture rendering a Thai bill-payment payload scans back through a standard decoder (e.g. ZXing) to exactly the bound string, control characters included.
- **CAP-2**
  - **intent:** An author can place a `qrcode` element whose bindable content renders as a QR Code, at a chosen error-correction level (L/M/Q/H, default M).
  - **success:** Fixtures at each supported error-correction level, including a UTF-8 Thai string and an EMVCo Thai QR Payment payload, decode to exactly the bound string.
- **CAP-3**
  - **intent:** A code fits inside its declared box with its required quiet zone and is never distorted.
  - **success:** Every module is a whole number of millipoints wide and all modules share one width. A QR code stays square and centred. The symbol plus quiet zone never exceeds the box. All of this is asserted against the page model.
- **CAP-4**
  - **intent:** The author is told, and no wrong code is drawn, when content cannot be encoded, is too long for the symbology, or cannot fit the box. The author is warned when the code fits only below the scannable minimum module size.
  - **success:** Unencodable, too-long and cannot-fit content each produce a registered Error naming the element id, and nothing is drawn. A module narrower than 0.25 mm (barcode) or 0.5 mm (QR) still draws and produces a registered Warning naming the element id, which `-strict` turns into a failure.
- **CAP-5**
  - **intent:** A designer user can add both elements from the palette, edit their content and options in the properties panel, and see the real code on the canvas.
  - **success:** Dragging each element onto a band, binding its content to a data field and previewing shows a code that matches the rendered PDF.
- **CAP-6**
  - **intent:** Both elements persist in the `.folio` text format and are documented there.
  - **success:** Load → save round-trips byte-identically. A document without either element keeps its current version and bytes. `folio-format.md` lists both types and their fields.

## Constraints

- `barcode` and `qrcode` extend the closed element-type set. That is a MAJOR format change (D-1.4.12, D-1.4.13): only documents that carry one declare the new version, and older readers refuse them.
- The code is drawn only with the PDF writer's existing filled-rectangle primitive. There is no third-party PDF writer (AD-3) and no raster image of the code.
- Encoders are pure Go with no cgo and no floating point: geometry is in integer millipoints. Output is byte-identical on darwin/arm64, linux/amd64, linux/arm64 and js/wasm (AD-21).
- A new Go module must pass the licence graph (AD-26: no GPL/LGPL/AGPL/SSPL/commercial at any depth) and a deliberate edit to the `gomod_test.go` pin.
- Content binds through the same expression evaluator and `{{ }}` syntax as a text element's `value`. There is no second binding path.
- Control characters such as field separators are authored as backslash escapes (`\n`, `\r`) in the static part of the content, outside `{{ }}`. Each escape encodes exactly the byte it names, with no silent conversion. Expression string literals stay escape-free. The designer's content field must let an author enter them.
- Each element ships a golden fixture under `fixtures/` in the four-platform hash matrix. Every new diagnostic is registered in the diagnostic registry.
- The designer holds no TypeScript document model: the canvas preview comes from the wasm engine.

## Non-goals

- 1D symbologies other than Code 128 (EAN-13/UPC-A, Code 39, ITF), and 2D symbologies other than QR (Data Matrix, PDF417, Aztec, Micro QR).
- A Thai QR Payment / PromptPay payload builder or a CRC16 expression function. The data or expression supplies the finished payload string.
- Rotation, colours other than black modules on the page background, logos inside QR codes, and inverted codes.
- Decoding or validating the payment semantics of a payload.
- Human-readable text printed with the barcode. Authors place their own text element.

## Success signal

- Render an invoice template bound to JSON with a Tax ID, Ref1, Ref2 and amount. The PDF's Code 128 barcode and QR code both scan in a Thai mobile banking app to the intended bill, and the PDF hash is identical on all four platforms.

## Assumptions

- The QR version and mask are chosen automatically per ISO/IEC 18004.
- Content that resolves to `null` hides the element silently, matching `visibleIf` null semantics. A path absent from the data is a located Error.
- The Thai bill-payment payload layout in `symbologies.md` is recalled from convention and has not been checked against the Thai Bankers' Association document.
