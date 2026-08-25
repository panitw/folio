# Expressions in a Folio template

Anything between double braces is an expression: `{{customer.name}}`. Folio evaluates it when the
report renders and substitutes the result into the document.

Expressions are deliberately small. There are **eight functions and no more** — no loops, no
variables, no arithmetic operators. If you find yourself wanting a ninth, the calculation probably
belongs in the data you supply rather than in the template.

> **Status while the engine is being built.** Every function below is described as it is specified.
> Functions marked *not yet implemented* are recognised by the template — the document still loads —
> but calling one produces an error naming the element, rather than a plausible-looking value. This
> is deliberate: a `sum` that quietly returned `0` before it was real would be a wrong total on a
> statement, and wrong totals are worse than errors.

---

## Reading values

### A plain path

Paths read from the data you supply with the report.

```
{{customer.name}}                    Somchai Srisuk
{{account.number}}                   123-4-56789-0
{{statement.closingBalance}}         48250.75
```

A path always reads from the top of your data, wherever it appears in the document.

### Parameters

Values supplied separately from the report data — a run date, a branch code, a report title — live in
their own namespace.

```
{{params.reportDate}}
{{params.branchCode}}
```

`params` can never be shadowed. A path beginning `params.` always means the parameter you supplied,
even inside a repeating region that happens to use the same word.

### Rows in a repeating region

A repeating region binds to a collection and gives each row a name:

```json
{ "bind": "transactions[]", "as": "transaction" }
```

Inside that region, the name you chose reads from the current row:

```
{{transaction.date}}
{{transaction.description}}
{{transaction.amount}}
```

Leave `as` out and the rows are called `row`. An unqualified path still reads from the top of your
data — **a row never shadows the document root**, so `{{customer.name}}` inside a transaction row
still means the customer, not a `name` field on the transaction.

You cannot name a region `params`, `page`, or `pages`. Those are reserved: `params` because it can
never be shadowed, and `page`/`pages` because nothing in a Folio template may ever refer to the page
it sits on. Using one is an error naming the element, raised **when the report renders**.

---

## Text

### `upper(text)` · `lower(text)`

```
{{upper(customer.name)}}             SOMCHAI SRISUK
{{lower(customer.email)}}            somchai@example.com
```

---

## Choosing between two values

### `if(condition, then, else)`

```
{{if(customer.isVip, "VIP", "Standard")}}
{{if(hasDiscount, discount.amount, "N/A")}}
```

**The condition must be true or false** — an actual boolean in your data. Folio does not treat `0`,
`""`, or an empty list as false. If the condition is some other kind of value, that is an error
naming the element, not a guess about what you meant.

Three cases worth knowing, because they differ:

| the condition is | what happens |
|---|---|
| `true` or `false` | takes that branch |
| **missing entirely** | **an error**, naming the element and the path — usually a typo |
| **explicitly `null`** | **treated as false**, silently |

Only the branch actually taken is evaluated. That is what makes the second example above work:
`discount.amount` does not exist on rows without a discount, and Folio never looks at it on those
rows.

The cost of that convenience is worth stating plainly: **a mistake in the branch that is not taken
goes unnoticed.** If `{{if(hasDiscount, discount.amout, "N/A")}}` has a typo in the branch that only
runs for discounted rows, you will not hear about it until a discounted row renders. Validating the
template cannot catch it either, because validation does not know which branch any given row will
take.

---

## Totals *(not yet implemented — Story 3.3)*

### `sum(collection.field)` · `count(collection)` · `avg(collection.field)`

```
{{sum(transactions.amount)}}
{{count(transactions)}}
{{avg(transactions.amount)}}
```

**Totals are exact.** Folio adds money as decimal digits, never as binary floating point, so a
statement total is correct to the last satang no matter how many rows it covers.

**A total always covers the whole collection**, never just the rows printed on the current page.
There is no per-page subtotal, and no expression anywhere can refer to the page it is on.

On an empty collection, `sum` and `count` return zero. `avg` reports an error instead of dividing by
zero, because the average of nothing is not zero — it does not exist.

---

## Dates and numbers *(not yet implemented — Story 3.4)*

### `formatDate(value, pattern)`

```
{{formatDate(statement.date, "d MMMM yyyy")}}

  locale th     15 สิงหาคม 2569        ← Buddhist era
  locale en     15 August 2026
```

The value must be either an RFC 3339 date string (`2026-08-15T00:00:00Z`) or a number of milliseconds
since the epoch. Anything else is an error.

### `formatNumber(value, pattern)`

```
{{formatNumber(transaction.amount, "#,##0.00")}}     1,234.56
{{formatNumber(statement.balance,  "#,##0")}}        48,251
```

### Locale

The document declares its locale and a fixed UTC offset. Folio ships tables for exactly four:

`en` · `th` · `zh-Hans` · `ja`

Any other tag is reported when the template loads — Folio will not quietly fall back to something
close. **The machine rendering the report never affects the output**: its locale, its time zone and
its clock are all ignored, so the same template and the same data produce the same bytes anywhere.

---

## The whole list

Eight, and the engine enforces that count.

| | function | status |
|---|---|---|
| Text | `upper` · `lower` | in progress |
| Choice | `if` | in progress |
| Totals | `sum` · `count` · `avg` | not yet implemented |
| Formatting | `formatDate` · `formatNumber` | not yet implemented |

Adding a ninth is not a small change and is not meant to be — the table is closed, and a new entry
has to be made deliberately and visibly.
