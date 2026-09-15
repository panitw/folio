# `fixtures/qrcode-payments/` — QR codes at every error-correction level

The golden for the **`qrcode`** element (spec-barcode-qr-elements CAP-2). It declares
**`"version": "4.0"`**, the major the `barcode` element introduced and the `qrcode` element joined.

| File | What it is |
|---|---|
| `input.folio` | Four QR codes on one page: `e1` `{{payload}}` with no `errorCorrection` (M), 120 x 120 pt; `e2` `{{payload}}` at H, 120 x 120 pt; `e3` `ชำระเงิน {{ref}}` at Q, 120 x 120 pt; `e4` static `INV-2026-000001` at L in a 200 x 80 pt box |
| `data.json` | The record: an EMVCo Thai QR Payment shaped `payload` and `ref` `INV-2026-000001` |
| `expected.pdf` | The render, sha256 `a14dc0aec702a28ef741fca2ebb09c79a0edd7122d181c6f305ce378e03bd84a` |

Each symbol must decode to exactly its resolved string, at its declared level:

| Element | Level | Decodes to |
|---|---|---|
| `e1` | M | `00020101021230810016A00000067701011201150994000123456780214INV2026000001030900000000153037645406150.005802TH62100706INV01263049A3F` |
| `e2` | H | the same payload |
| `e3` | Q | `ชำระเงิน INV-2026-000001` (UTF-8, no ECI) |
| `e4` | L | `INV-2026-000001` |

The payload has the EMVCo Merchant-Presented Mode shape (tag 30 bill payment, CRC tag 63). Its CRC
is illustrative: the element encodes the finished string and never builds or checks a payload.

## What the tests prove

- `TestQRCodePaymentsGoldenFixture` pins `input.folio` and `data.json` to the Go constants
  (`qrcode_payments_template.go`), and the render's sha256 to `expected.json` and `expected.pdf`.
- `TestQRCodePaymentsSemanticAcceptance` reads each QR code's rectangles back out of the page model
  and checks they cover exactly the dark modules of its string's encoding at its declared level, on
  one whole-millipoint module grid of the largest size that fits, with the quiet zone inside the box.
- The document draws no text, so no face is embedded: every byte that could differ between the four
  matrix targets is integer module geometry.

## Independent check

Run `uv run --with zxing-cpp --with pillow python3 tools/codescan/scan.py fixtures/qrcode-payments/expected.pdf`
from the repository root. It rasterises the page's filled rectangles and decodes them with zxing-cpp;
it must print all four strings above at their declared levels (L, M, Q, H).
