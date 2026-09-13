# `fixtures/barcode-thai-bill-payment/` — a Code 128 bill-payment barcode

The golden for the **`barcode`** element (spec-barcode-qr-elements CAP-1). It is the first committed
document declaring **`"version": "4.0"`**, because it is the first carrying a barcode.

| File | What it is |
|---|---|
| `input.folio` | One barcode, `e1`, 400 x 50 pt, value `\|0994000123456{{suffix}}\r{{ref1}}\r{{ref2}}\r{{amount}}` |
| `data.json` | The record: suffix `01`, ref1 `1234567890`, ref2 `0000000001`, amount `150000` |
| `expected.pdf` | The render, sha256 `c04c1a848843b5adda8bb1852b74957cef99f90f522c09d14783e8d4f29c61ef` |

The bars must decode to exactly:

```
|099400012345601<CR>1234567890<CR>0000000001<CR>150000
```

where `<CR>` is a carriage return (byte 13). The static part — the biller's tax id and the three
separators — comes from the template; the suffix, references and amount come from the data.

## What the tests prove

- `TestBarcodeThaiBillPaymentGoldenFixture` pins `input.folio` and `data.json` to the Go constants,
  and the render's sha256 to `expected.json` and `expected.pdf`.
- `TestBarcodeThaiBillPaymentSemanticAcceptance` reads the page model's bars back into module runs
  and compares them with the Code 128 encoding of the resolved string, checks every module is the
  same whole number of millipoints, and checks the quiet zones. `internal/barcode`'s test-only
  decoder independently round-trips the same payload, carriage returns included.
- The document draws no text, so no face is embedded: every byte that could differ between the four
  matrix targets is integer bar geometry.

## Manual check (not automated)

Scan the barcode in `expected.pdf` with a phone scanner or ZXing. The decoded text must equal the
string above. This repository has no Code 128 decoder outside its own test code, so this scan is the
only check against an independent reader.

The payload layout is recalled from convention and has not been checked against the Thai Bankers'
Association document (spec-barcode-qr-elements, Assumptions).
