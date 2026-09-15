# Example templates

The four bundled examples (CAP-4). Data is fictional and English: invented company, people,
addresses and account numbers — no real organization, and account/meter numbers that cannot be
mistaken for real ones. Currency is neutral (two decimals, thousands separators via
`formatNumber`, no locale-specific symbol). Each example is `<name>.folio` + `<name>.sample.json`.

"Must exercise" is the load-bearing column: the example exists to show an author that feature
working. Layout details beyond it are the implementer's.

| Example | Pages with sample | Must exercise | Sample data shape (top level) |
|---|---|---|---|
| **Blank** | 1 | — (today's `starter.folio`, unchanged) | none |
| **Invoice** | 1 | Header with issuer/bill-to blocks; line-item `table` bound to a collection with qty × price columns and a column-footer `sum`; subtotal/tax/total text with `formatNumber`; `formatDate` for issue/due dates; payment reference as `qrcode` | `invoice{number, issued, due}`, `seller{…}`, `customer{…}`, `items[]{description, qty, unitPrice, amount}` (6–10 rows), `totals{subtotal, tax, total}` |
| **Bank Statement** | 2–3 | Transactions `table` that paginates with repeated header; page header/footer with "Page X of Y"; alternating rows; opening/closing balance; a static legend placed under a section break | `account{holder, number, period{from, to}}`, `balances{opening, closing}`, `transactions[]{date, description, debit, credit, balance}` (enough rows to span ≥2 pages) |
| **Legal Contract** | 2–3 | Long flowing body text across pages; numbered clause headings; parties block bound from data; `keepTogether` signature block that does not split; page footer with document reference and page numbers | `agreement{title, reference, effective}`, `partyA{…}`, `partyB{…}`, `clauses[]{heading, body}` (8–12), `signatories[]{name, role}` |
| **Electricity Bill** | 1 | Account/meter summary; usage `table` (previous/current reading, units, rate, charge) with footer total; 6-month usage history as a table (no chart element exists); conditional `visibleIf` overdue notice; payment `barcode` | `customer{…}`, `account{number, meter}`, `period{from, to}`, `readings[]{…}`, `charges[]{label, units, rate, amount}`, `history[]{month, units}`, `amountDue`, `dueDate`, `overdue` |

## Rules every example follows

- Declares the lowest format version its features require and stays in canonical form, so a
  no-op load/save round trip is byte-identical.
- Renders against its sample with zero diagnostics — checked in the build, which also emits the
  thumbnail from that render.
- Sample JSON is accepted by `acceptSampleData` (the tree display may truncate long collections;
  the render receives the whole file).
