# Expressions in a Folio template

Anything between double braces is an expression: `{{customer.name}}`. Folio evaluates it when the
report renders and substitutes the result into the document.

Expressions are deliberately small. There are **eight functions and no more** — no loops, no
variables, no arithmetic operators. If you find yourself wanting a ninth, the calculation probably
belongs in the data you supply rather than in the template.

> All eight functions are implemented. A function called with the wrong kind of argument, or with
> data that cannot support what it asks for, still produces an error naming the element rather than
> a plausible-looking value: a `sum` that quietly returned `0` on bad input would be a wrong total on
> a statement, and wrong totals are worse than errors.

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

## Totals

### `sum(collection.field)` · `count(collection)` · `avg(collection.field)`

`sum` and `avg` take a **projection path**: a collection, followed by the field to add up or
average across every element. `count` takes the **collection path alone** — it never looks at any
field, so an element missing the field `sum`/`avg` would need still counts.

**An aggregate is a number, and a number must be formatted before it can appear in text.** A bare
`{{sum(transactions.amount)}}` is an **error** — text bindings are never coerced (the same rule that
makes a bare `{{transactions.amount}}` an error today):

```
{{sum(transactions.amount)}}    Error — a number is never coerced to text
{{avg(transactions.amount)}}    Error — same rule
{{count(transactions)}}         12    (an exception, see below)
```

A number-valued aggregate must be wrapped in `formatNumber(...)` before it can appear in text — see
*Dates and numbers*, below. `count` is the one exception: it is already a plain non-negative whole
number, and needs no formatting to render as text.

```
{{formatNumber(sum(transactions.amount), "#,##0.00")}}     1,234.56
```

**Totals are exact.** Folio adds money as decimal digits, never as binary floating point, so a
statement total is correct to the last satang no matter how many rows it covers. `avg` divides at
the greatest number of decimal places any operand carries, plus a fixed number of extra digits —
four today; **the constant is illustrative, the rule is not** — with round-half-to-even, so a
repeating average never silently loses the tie-breaking digit.

**A total always covers the whole collection**, never just the rows printed on the current page.
There is no per-page subtotal, and no expression anywhere can refer to the page it is on.

**An explicit `null` value is a zero observation, not a missing one.** A row whose amount is JSON
`null` contributes zero to `sum` and counts as one observation in `avg`'s divisor — it pulls the
average down, exactly as a real zero would. A row whose amount is **absent entirely** is a different
thing: that is an error, naming the row and the field, because the document genuinely does not say
what belongs there. Two rows that look almost the same in JSON — `{"amount": null}` and `{}` — are
treated very differently for exactly this reason.

**The same rule extends to the collection itself.** A document where `transactions` is JSON `null`
(rather than a missing key, and rather than an empty list) is treated as one zero observation, the
same as a single `null` row would be: `sum` is `0`, `count` is `1`, and `avg` is `0` at its usual
scale. This is different again from `transactions` being genuinely **absent** from the data, which
is still an error.

**On an empty collection**, `sum` and `count` are legitimately zero. `avg` cannot divide by zero
observations, so it is not a number — but this is a **caveat the render survives**, not a failure:
the total column renders blank, and the render notes why, rather than refusing to produce the
document at all. This is different from an all-null collection, whose average **is** a real number
(zero, at the scale the rule above derives) — a collection with rows that happen to be blank and a
collection with no rows at all are not the same subject, and render differently on purpose.

---

## Dates and numbers

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
| Text | `upper` · `lower` | implemented |
| Choice | `if` | implemented |
| Totals | `sum` · `count` · `avg` | implemented |
| Formatting | `formatDate` · `formatNumber` | implemented |

Adding a ninth is not a small change and is not meant to be — the table is closed, and a new entry
has to be made deliberately and visibly.
