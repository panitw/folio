# Symbologies

Reference for the encoders behind CAP-1 to CAP-4 in `SPEC.md`.

## Code 128 (`barcode`)

| Aspect | Rule |
|---|---|
| Character set | ASCII 0–127. Code set A covers control characters such as CR, B covers printable characters, C covers digit pairs. |
| Set selection | Deterministic, and minimal in symbol length: start in C for runs of four or more digits, and use SHIFT or CODE switches per ISO/IEC 15417. The same input always gives the same bars. |
| Check character | A mandatory modulo-103 check before STOP. It is never shown to or authored by the user. |
| Quiet zone | At least 10 modules on each side, inside the element box. |
| Bar height | Fills the box height. No human-readable text. |
| Unencodable | A character above 127, including Thai script, is a located diagnostic (CAP-4). |

### Thai bill-payment payload (fixture shape; unverified — see SPEC Assumptions)

```
|<TaxID 13 digits><Suffix 2 digits>\r<Ref1>\r<Ref2>\r<Amount in satang, no decimal point>
```

Authored as `|0994000123456{{suffix}}\r{{ref1}}\r{{ref2}}\r{{amount}}`, where `\r` is a carriage return (byte 13), not `\n`.

- Bank of Thailand standard: Code 128, at most 62 characters, bar height at least 1 cm.
- The payload arrives already composed, from the data or from an expression. The element never assembles it.

## QR Code (`qrcode`)

| Aspect | Rule |
|---|---|
| Standard | ISO/IEC 18004 (Model 2). |
| Encoding | Byte mode over the UTF-8 bytes of the content, with no ECI. Numeric and alphanumeric segmentation may be used when it shortens the symbol, and must be deterministic. |
| Error correction | Authorable L / M / Q / H, recovering about 7 / 15 / 25 / 30% damage; default M. A higher level needs more modules in the same box. |
| Version | The smallest version (1–40) that holds the content at the chosen level. Content too long for version 40 is a located diagnostic. |
| Mask | The lowest penalty score across the eight masks per the standard. Ties go to the lowest mask number. |
| Quiet zone | 4 modules on each side, inside the element box. |
| Geometry | Square. The module width is the largest whole-millipoint size such that the symbol plus quiet zone fits min(width, height). The symbol is centred. |

As built (spec-qrcode-element), the encoder uses byte mode only: numeric and alphanumeric segmentation are permitted by the rule above but not used, which keeps segmentation trivially deterministic at the cost of at most a version or two for digit-heavy payloads. The level is the element's optional `errorCorrection` key.

## Module-size diagnostics (CAP-4)

| Case | Barcode | QR | Severity | Drawn? |
|---|---|---|---|---|
| Module fits but below minimum | < 0.25 mm (709 millipoints) | < 0.5 mm (1418 millipoints) | Warning (`-strict` fails) | Yes |
| Symbol plus quiet zone cannot fit at 1 millipoint per module | — | — | Warning (`-strict` fails) | No |

Millipoint values round 0.25 mm and 0.5 mm up to the nearest whole millipoint.

### Thai QR Payment

- An EMVCo Merchant-Presented Mode payload (tag 29 PromptPay, tag 30 bill payment) ending in a CRC16 tag 63.
- The element encodes the finished string only. Payload building and CRC are non-goals.
