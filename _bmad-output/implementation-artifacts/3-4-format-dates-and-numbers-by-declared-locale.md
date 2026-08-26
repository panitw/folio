---
baseline_commit: c2b4fe2a15eef540449bac85ca4e20a96cce32fc
---

# Story 3.4: Format dates and numbers by declared locale

**Epic:** 3 — The expression language, aggregation and diagnostics
**Story key:** `3-4-format-dates-and-numbers-by-declared-locale`
**Status:** `done`
**Covers:** **FR18** · **AD-1**, **AD-12**
**Primary invariant:** **AD-12** (formatting locale is declared, never discovered: one `locale` tag
and one fixed UTC offset per document; a compiled, embedded, versioned table for the **closed set**
`en` · `th` · `zh-Hans` · `ja`; an unlisted tag is a load error, not a fallback; `th` renders
Buddhist-era years; `formatDate` accepts only RFC 3339 or epoch-milliseconds; `formatNumber` scales
by integer powers of ten from a **lookup table**, never `math.Pow` — ARCHITECTURE-SPINE.md:267-282).
**Co-primary invariant:** **AD-1** (the determinism boundary is a directory boundary: no `time`,
`os`, `math/rand`, `net`, no `math` transcendental, no package-level mutable state, no
output-reaching map iteration, anywhere under `internal/` — ARCHITECTURE-SPINE.md:92-103). **This
story's "no host locale, no host time zone, no host clock" clause IS AD-1 restated. It is already
guarded. Do not build a second guard.**
**Adjacent invariants:** **AD-23** (no binary float — this is *why* `math.Pow` is banned: it returns
`float64`) · **AD-14** (one `Diagnostic` type, closed code registry, never a panic) · **AD-8**
(`FontSet` is a public `map[string][]byte`, and **the fallback chain is declared by the document, not
derived from the locale**) · **AD-3** (one number emitter) · AD-9 · AD-13 · AD-22 (the locale table
version is part of the library version) · AD-11 · AD-4.
**Rulings issued AGAINST this story during creation, all folded in below. NOTHING in this file is
open — every flag raised at creation has been ruled before development starts:**
**D-3.4.1** (OWNER: `ja` stays, the kanji-shape limit is disclosed in writing, no fence — *The `ja`
question*, AC18/AC19/AC21) · **D-3.4.2** (LEAD, **amended on this story's pushback**: the locale
closed SET is declared in `internal/template`, the locale TABLE lives in `internal/expr` and is keyed
from it; set difference in **both** directions, never a count, never an AST scan — AC4) ·
**DECISION-1, ruled ARM A** (the registered-but-unimplemented machinery is REMOVED, with a
two-half behavioural+structural replacement — AC16) · **DECISION-2, OWNER** (Western/ASCII digits
everywhere, as an **explicit named table field** — AC11) · **FINDING-4, ruled** (the date-pattern
grammar is field tokens plus non-ASCII-letter literals; **no quoting mechanism ships** — AC10) ·
**FINDING-3, ruled** (FLAG-1 stays where it is; `deferred-work.md`'s scope does not widen — F4) ·
**D-000.66** (a *"verified, not assumed"* phrase is a DEBT, and re-passing it verbatim is the one
disallowed move — the rule generalised from the four-session `ja` incident this story's creation
ended).
**Governing rulings:** **D-3.1a.1** + correction (`avgExtraScale`, Flag F3 — **this story is its
only forcing function**) · **D-3.1a.2** · **D-3.1a.3** · **D-3.2.1** · **D-3.2.2** · **D-3.3.1** …
**D-3.3.9** · **D-1.4.1** · **D-1.4.2** · **D-1.6.3** (AD-12's host-vs-document locale reading) ·
**D-000.65** (a diagnostic code is minted when its condition first ships) · **D-000.64** (the gate is
now three modules) · **D-000.59** (an absence/vacuity tripwire is discharged by REPLACEMENT) ·
**D-000.60** (a restriction taken out against an open question is lifted when the question is
answered) · **D-000.61** + extension · **D-000.4** · D-000.9 · D-000.13 · D-000.14 · D-000.15 ·
D-000.17 · D-000.23 · D-000.24 · D-000.26 · D-000.30 · D-000.36 · D-000.38 · D-000.42 · D-000.45 ·
D-000.47 · D-000.50 · D-000.57 · **D-000.22** (a golden needs a semantic acceptance step by someone
who reads the language) · **D-000.41** (do not buy attributability with a scarce human we do not
have) · D-2.1.14 · D-2.2.4 · **D-2.4.3** · D-2.8.1 · D-2.8.6.

**Two divergences between the record and the shipped code were measured during creation and are
written into the ACs below.** Both are consequences of this story flipping the LAST two
registered-but-unimplemented entries, and neither was visible from the brief. See *Findings* §F1 and
§F2. One of them (F1) contradicts a claim made in a file comment shipped at 3.3.

**The `ja` question that shadowed this story's creation is ANSWERED, and nothing here is fenced.**
The lead measured the shipped `cmap` tables directly; the owner ruled. All four locales are in scope.
See *The `ja` question — answered*, which also records the one **scope fence** the ruling creates and
the two referents it hands to later stories.

**Retires no DW entry.** DW-19's three `lint` reds are local-only and stay. **FLAG-1 from Story 3.3
(`avgExtraScale = 4` is illustrative at MEDIUM confidence) is discharged here** — by AC17, and only
by AC17.

---

## Baseline, measured in this run at creation

HEAD **`c2b4fe2`** — *"Story 3.3: Aggregate over a collection (finisher)"* — branch `main`, working
tree **clean** (verified in this run). `CGO_ENABLED=0 go build ./...` in `folio-go/` succeeds.

**The gate is now THREE modules (D-000.64, extended by 3.3's new `hashmatrix` module).** The figures
below are the orchestrator's, declared against this commit; **they are stated here so the developer
measures against them rather than carrying counts forward, and the developer re-measures all three
before writing a line of code** (D-000.26: a figure without its scope and flags is not a figure).

| scope | invocation (verbatim) | declared result at `c2b4fe2` |
|---|---|---|
| `folio-go/` | `go build ./... && go vet ./... && gofmt -l . && go test ./... -count=1` | **783 pass · 1 fail · 1 skip** |
| `lint/` | `go test ./... -count=1` | **95 pass · 3 fail** |
| `hashmatrix/` | `go test ./... -count=1` | **3 pass** (module new at 3.3) |

**The one `folio-go` FAIL is `TestCorpusMeetsP6ExerciseFloors` — a REQUIRED red**, stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` (P6g floor 20 not met). **D-000.17** ("a floor
that is not met is reported as unmet — it is never filled"), **D-2.1.14**, **D-000.57**. It measures
the Thai line-breaking corpus and has nothing to do with this story. **Never "fix" it.** Its stats
must be byte-identical at hand-off or D-000.57's third clause is breached.

**The three `lint` FAILs are one root cause — DW-19, and they are LOCAL-ONLY.** `.font-sources/` is
gitignored at `.gitignore:85` and holds three variable-font `.ttf` with no `LICENSE*`; the lint asset
resolver walks it and fails. **CI never sees the directory and is green.** Any `lint` figure this
story reports must name these three or it is not a figure. **Fixing them by weakening the rule is
forbidden** (D-000.64, D-000.59's delete-the-rule failure). Any *fourth* `lint` failure is a
regression owned by this story.

**Cadence (D-000.4):** build / vet / gofmt / unit tests, **all three modules**, every story. **The
cross-target hash matrix is NOT run in this story.** Judged: this story is not hash-shaped. It adds
no font byte, no image byte, no glyph, no measurement and no new PDF operator class — it produces a
*string* that then flows through the already-pinned shaping and serialization path. Its determinism
risk is entirely covered by AD-1's existing import/selector lints plus the exactness assertions in
AC13-AC17, none of which needs four targets to observe. The matrix is due at Epic 3's close.

---

## In plain terms (read this first if you just want the gist)

A Thai bank statement dated August 2026 does not say 2026. It says 2569, because Thailand counts
years from the Buddhist era, and it writes the month in Thai. The engine could already fetch a date
or an amount out of a customer's data and total a column of money, but it could not turn either into
the words and digits a reader expects on the page. Two of the eight things the template language can
do were reserved for exactly this and have been answering "not built yet". This story builds them,
and closes a gap found along the way: an impossible date, such as the 31st of February, used to be
silently rounded into a different, ordinary-looking date instead of being refused. It is refused now,
by name.

None of the answer comes from the machine doing the printing. The document says which language it is
written in and which time zone its dates belong to; the engine carries its own built-in book of month
names, separators, numeral styles and calendar rules for the four supported languages, and never asks
the computer what country it is in or what time it is. A document's dates and its amounts always
share one style of digit, never two. That is why the same statement produces an identical file
anywhere, and why a date has to arrive as a properly written timestamp.

All four languages are built here, Japanese included, with one honest limit written down rather than
fixed. Our fonts were checked character by character and can draw every Japanese character — nothing
is missing. But Chinese and Japanese share many characters drawn slightly differently in each
country, and a font holds only one drawing of each, so a Japanese reader sees the Chinese styling:
legible and complete, but not Japanese typography. Anyone who needs that supplies their own font,
which the product has always allowed. So the Japanese example we check is dates and numbers only, and
we claim no more — nobody here reads Japanese well enough to judge the rest.

It invents no formatting options beyond the small published vocabulary; anything outside it is
rejected when the template loads rather than half-working later. Tables, pagination and the
diagnostics rewrite belong to later stories.

Done looks like this: the Thai statement shows "15 สิงหาคม 2569" and "1,234.56", the English one "15 August
2026", the Japanese one "2026年8月26日" — identical bytes on any machine. Two things will look wrong
afterwards and are not: one coverage-marker test stays red on purpose, and three licence-checker
warnings appear only on this developer's machine, from a folder that never ships.

## Story

**As a** template author in Bangkok,
**I want** a Thai statement to show a Buddhist-era date and correctly grouped numbers,
**So that** the document reads as a Thai document rather than a translated one.

---

## Do not re-open — settled rulings this story inherits

1. **The function table is closed at eight, and this story adds nothing to it.** Two entries'
   `implemented` flag flips. `TestExprFunctionTableIsExactlyEight` and
   `TestExprFunctionTableRedProofNinthEntry` stay green **with zero edits**.
2. [x] **`expr.Resolver`'s method set is CLOSED at exactly three methods, by full resolved signature.**
   `lint`'s `resolver-method-set-closed` rule (`lint/internal/rules/resolvermethodset.go`,
   `expectedResolverMethods`) pins `Resolve`/`CollectionLength`/`ProjectCollection` and **expands
   embedded interfaces**, and `folio-go/internal/expr_arch_test.go`'s
   `TestExprResolverMethodSetIsClosed` pins the name set independently. **The locale does not travel
   through `Resolver`.** See R1.
3. [x] **`internal/expr` may not import `internal/bind`** (stage rank 3 → 4). It may import
   `internal/template` (rank 2).
4. [x] **`internal/expr` may not import `time`.** AD-1's forbidden set is `os`/`time`/`net`/`math/rand`
   and it applies to `_test.go` files under `folio-go/internal/` too — `reduce_test.go:280-281`
   already states this in those words. **All calendar arithmetic in this story is written by hand.**
5. [x] **No binary floating point.** `float32`, `float64`, `math/big.Float`, `math/big.Rat` — D-3.1a.1
   (corrected) and its two-layer guard. `math.Pow` is banned *because it returns `float64`*, not as
   a style rule.
6. [x] **`SumDecimals([])` returns exponent `0`; an all-null `avg` returns exponent
   `-avgExtraScale`.** D-3.1a.2, restated at 3.3's *Do not re-open* item 4: *"Presentation scale is
   Story 3.4's `formatNumber` / `footerFormat` job."* **This story does the presentation. It does not
   move either kernel exponent.** A developer who "harmonises" them has re-decided the scale at a
   call site.
7. [x] **`Decimal` does not move again.** `TestExactlyOneDecimalDeclarationInTheModule` and
   `reducer_inventory_arch_test.go` stay green with zero edits.
8. [x] **No `page` namespace, ever** (AD-4). Not as a pattern token, not as a date source.
9. [x] **Statically decidable defects belong at `Check` (load), not at `Eval`.** R3 of Story 3.2, applied
   at 3.2's own review to number literals (`checkNumberLit`, `check.go:48-58`): *"a number literal's
   bounds … are decidable with NO data at all, exactly like arity, so R3 puts them at Check."* **A
   pattern literal is exactly that case.** See AC10.
10. [x] **A diagnostic code is minted by the story in which its condition can first occur in a real
    document, at the site where the `Diagnostic` is constructed — today `folio/diagnostic.go`**
    (D-000.65, following D-2.8.1's shipped `DiagCodeTextClippedWidth` and 3.3's
    `DiagCodeEmptyAverage`). `internal/diag` is still Story 3.6's. **This story mints no code unless
    an AC below names one** — every failure it introduces is a located Go `error` (AD-14 Error
    disposition), not a Warning, and errors are not registry codes today.
11. [x] **The document's own declared locale is REQUIRED reading; the HOST's is banned.**
    `render_entry.go:12-20` already records this distinction verbatim (D-1.6.3, AC14). Do not
    re-argue it and do not weaken that file comment.

---

## Findings — measured in this run at `c2b4fe2`

### F1 — **`TestUnimplementedEntriesHaveNoEvalCallBranch` FATALS the moment this story flips the last two entries.** The file that contains it claims the opposite, in writing.

`folio-go/internal/expr/table_derivational_test.go` opens with:

> *"So stated, this survives 3.3's edit (three names move across) and 3.4's (two more) with **NO edit
> to this file at all**."*

**It does not.** `TestUnimplementedEntriesHaveNoEvalCallBranch` ends with:

```go
if unimplemented == 0 {
    t.Fatal("presence precondition (D-000.9): zero unimplemented entries found — the comparison above would pass vacuously")
}
```

After `formatDate` and `formatNumber` flip, `unimplemented == 0` and this test **fails**.

**The file is worse than a single slip, and the shape of the defect is the lesson. Measured: that
file carries THREE D-000.9 presence preconditions — `:85` (the `switch entry.name` statement was
found), `:121` (`len(tableImplemented) == 0`), `:159` (`unimplemented == 0`). Exactly two of the
three are safe forever, and exactly one is not.** `:85` keys on a structural fact that does not
expire. `:121` keys on a population the roadmap drives **to eight**. `:159` keys on the only
population **the roadmap drives to zero on purpose**. So the file's top comment is false for
precisely one of its three guards, and it is not false through carelessness.

**The lesson is therefore NOT "hard-coded expectations recur".** It is that **D-000.9 and D-000.59
collide at a scheduled zero.** A presence precondition is itself a **population-keyed assertion**, so
it must be checked against the roadmap the same way an absence tripwire is — and when the schedule
empties its population, the precondition fires as a **false alarm at exactly the moment the guard it
protects becomes correctly vacuous.** The alarm and the correct outcome arrive together, which is why
the intuitive response (soften the threshold) is the wrong one.

**DECISION-1's ruling IS this finding's fix: removing the mechanism removes its precondition with
it.** **Do NOT relax, re-tune, or "correct" the `== 0`, and do not re-add a softened version of it.**
A precondition guarding a guard that can no longer fail is one more thing asserting a proposition
that can never again be false. See AC16.

### F2 — Three more tests, and one production branch, go dead or red on the same flip.

Measured, by name:

| site | what happens at the flip |
|---|---|
| `internal/expr/eval_test.go:259-268` `TestUnimplementedFunctionsAreLocatedErrors` | its two subjects (`formatDate`, `formatNumber`) now succeed — **red** |
| `internal/expr/eval_test.go:346-372` `TestUnimplementedAndUnknownFunctionErrorsAreDistinguishable` | uses `formatDate` as *the* unimplemented example and asserts the message *"should name Story 3.4"* — **red** |
| `internal/expr/eval.go:57-66` `if !entry.implemented { … }` | **unreachable** — no table entry can reach it |
| `internal/expr/table.go` `funcEntry.implemented`, `funcEntry.owningStory` | `owningStory` is read by nothing that can fire; `implemented` is read only by branches that can no longer be false |

This was a design fork, not a mechanical fix. **RULED: ARM A — remove the machinery** (see
*Decisions raised at creation*, executed by **AC16**). The ruling also discharges F1: removing the
mechanism removes its expiring precondition with it.

### F3 — `render_bind_test.go:210` binds the pattern literal `"x"`, and AC10 turns that into a LOAD error.

Today `{{formatNumber(a, "x")}}` passes `Check` (`argStringLiteral` only asks *is it a string
literal*) and fails at `Eval` with the not-yet-implemented message. Once patterns are validated at
`Check` (item 9 above, AC10), `"x"` is rejected when the template loads. **That test's premise
changes and the developer owns re-pointing it, not deleting it.** The two goldens that bind the real
pattern — `folio-go/testdata/template/golden/worked-example.json:19` and
`internal/template/fixtures_test.go:263`, both `"#,##0.00"` — must keep loading unchanged; that they
do is AC10's own non-vacuity check.

### F4 — FLAG-1 is not in `deferred-work.md`, and **RULED: it does not move there.**

`deferred-work.md` holds DW-1 … DW-19 and **no `FLAG-1` entry** (grepped: zero hits for the literal,
one incidental hit for `avgExtraScale` at `:300`). FLAG-1 is recorded in **Story 3.3's own Flags
section** (`3-3-aggregate-over-a-collection.md:733-735`) and its substance in **`reduce.go:33-42`
(Flag F3)** and **D-3.1a.1**.

**RULED (FINDING-3): nothing moves, and nothing is filed.** `deferred-work.md` is for **work
consciously not done, with a named owner**. FLAG-1 is not deferred work — it is **an obligation a
ruling places on a later story**, and its home is the ruling's own log entry. **D-3.1a.1 recorded it
correctly, naming Story 3.4 in terms**; only the creation brief's filing location was wrong. **Do not
move it, and do not widen `deferred-work.md`'s scope to hold obligations as well as deferrals** — a
file holding both stops being scannable for either.
*Recorded here so nobody greps `deferred-work.md`, finds nothing, and concludes FLAG-1 was already
discharged. It is discharged by AC17 and nowhere else.*

### F5 — The plumbing that does not exist yet, stated precisely.

`bind.BindTextSpans(text string, data, params Value, elementID string)` (`internal/bind/text.go:187`)
→ `bind.Resolve(text, scope, elementID)` → `expr.Eval(e, resolver, elementID)`. **Nothing in that
chain carries a locale or a UTC offset.** The values exist and are already validated: `Document.Locale`
(closed set enforced at `internal/template/parse.go:84-86`) and `Document.UTCOffset` (**required**,
`±HH:MM`, enforced at `parse.go:92-103` against `utcOffsetPattern`, `closedsets.go:47`). The single
production call site is `folio/render.go:571`, inside `collectBandTextRuns`, which **already holds
`doc *Template`**. So the seam is a parameter, not a lookup. See R1.

### F6 — Sharp edge 6, answered by measurement rather than by design.

*"What happens when a document declares a locale but no offset?"* — **it is already a load error, and
has been since Story 1.4.** `parse.go` treats `utcOffset` as a **required** top-level field
(`"missing required field"`), exactly as `locale` is. There is no default and no story work here.
**AC5 pins that this stays true**, because the cheapest way to make a formatting test easier is to
give the offset a default. **Confirmed by the coordinator and recorded as ALREADY ANSWERED, not
solved twice** — there is nothing to build here, only something to keep from being undone.

### F7 — The docs twins are already prepared for this story and name it.

`docs/expression-reference.md:133-135` says the worked `formatNumber` example *"belongs to Story 3.4,
not here"* (3.3's Finding 8 was resolved by reverting it). `:169-198` carries the whole *Dates and
numbers (not yet implemented — Story 3.4)* section, including the canonical outputs
**`15 สิงหาคม 2569`** (`th`) and **`15 August 2026`** (`en`) for pattern `"d MMMM yyyy"`, and
`"#,##0.00"` → `1,234.56`, `"#,##0"` → `48,251`. `:211` and the `.html` twin's `:290-291`, `:461-473`,
`:498` carry *Pending* markers. **These published outputs are the specification this story is held
to** (2026 + 543 = 2569 confirms the era rule). Both twins are edited in the same commit (D-000.47).

---

## R — design constraints derived from the record during creation

These are not new decisions. Each is the only arm the shipped guards and standing rulings leave open,
and each is recorded with what closed the others, so a developer does not re-derive them under
delivery pressure.

### R1 — The formatting context reaches `Eval` as an explicit parameter. Not through `Resolver`, not through package state.

Three routes exist. Two are barred:

- **Through `expr.Resolver`** (a fourth method, or a widened signature): barred by *Do not re-open* 2.
  `expectedResolverMethods` pins the method set by **full resolved signature**, and
  `TestExprResolverMethodSetIsClosed` pins the names — so this route requires editing **two closed-set
  guards in the same diff as the thing they guard**, which is **D-3.1a.3**'s named hazard verbatim.
- **Package-level state** (a settable current locale): barred by **AD-1** outright, and detectable by
  the existing lint.
- **An explicit value threaded through the call chain**: the remaining arm.

So: a small immutable value — locale tag plus fixed UTC offset — is declared in `internal/expr`,
constructed once at the render entry from `doc.Locale`/`doc.UTCOffset`, and passed down
`render.go:571` → `bind.BindTextSpans`/`bind.Resolve` (or carried on `bind.Scope`, which already
exists as the per-render carrier and is the better home if the signature churn is otherwise wide) →
`expr.Eval`. **It is a value, never a pointer to something mutable, and it has no zero-value default
that means "en".** A zero value must be unusable — see AC5.

### R2 — The locale table lives inside `internal/expr`. No new package under `internal/`.

`lint`'s `stage-rank` rule (`stagerank.go:57-80`) enforces that **an unranked package directory under
`internal/` is a finding**. A new `internal/locale` would require an edit to `stageRankTable` —
again, a guard edited in the same diff as the feature. AD-12 already **binds `internal/expr`**
(spine `:269`) and the source tree map at `:610` says so in words: *"`expr/` — lexer, parser, the 8
functions, **locale tables (AD-12)**"*. Put it there.

### R3 — "No host locale / time zone / clock" is discharged by the EXISTING guards, cited by symbol.

Per **D-000.38** ("two guards sharing a parser are one guard wearing two names") and **D-000.42** (no
redundant third guard), this story does **not** build a host-independence check. It cites, by symbol:
`lint`'s `forbidden-imports` rule (`forbiddenimports.go:43-45`, bans `time`/`os`/`net`/`math/rand`
under `folio-go/internal/`, `_test.go` included) and `mathAllowedCalls` (`:87-90`, a closed
**allow-list** — `Pow`, `Exp` and `Log` are all absent from it and therefore already banned).
**The story's obligation is to prove those guards COVER THE NEW CODE PATH**, which they do by
directory — and to prove it non-vacuously (AC2). It is not to build a second fence.

---

## The `ja` question — ANSWERED (**D-3.4.1**). Read this before assuming anything about Japanese.

**The premise this story was drafted against was false, and the correction changes the work.** The
long-standing note that *"whether Noto Sans SC's coverage suffices for Japanese must be verified, not
assumed"* has now been verified — by **parsing the shipped `cmap` tables directly**, not by trusting
the font's reputation:

| script range | measured in `NotoSansSC-Regular.ttf` |
|---|---|
| Hiragana | **93 / 93** assigned codepoints (the two "missing" of 95 are U+3097 / U+3098, **unassigned in Unicode**) |
| Katakana | **96 / 96** |
| Halfwidth kana | **58 / 58** |
| Kana phonetic extensions | **16 / 16** |
| CJK punctuation (、。「」) | **64 / 64** |
| Kokuji and shinjitai | **all present** — 峠 働 込 匂 辻 畑 · 円 様 直 骨 |
| total | **30,890 codepoints** |

Neither Noto Sans nor Noto Sans Thai carries a single kana, so the CJK face is the **sole Japanese
carrier** — and it is complete. **There is no coverage gap.**

**The only defect is glyph SHAPE.** One font holds one glyph per codepoint, so wherever Japanese and
Simplified-Chinese conventions diverge on a Han character, a Japanese reader sees the Chinese form.
Legible, complete, regionally wrong. **That is a typography limitation, not a rendering failure**, and
it is permanent and universal rather than data-dependent.

### OWNER DECISION — keep `ja`, and state the limit in writing

**`ja` gets its table row, in this story, with no fence.** For everything this story implements —
grouping, decimal separator, the `年月日` date shape — **`ja` is effectively identical to `zh-Hans`.**
One row, built and asserted like the other three.

**Grounding, recorded so nobody re-opens it:** nobody ever asked for Japanese. `SPEC.md`'s Assumptions
say the four locales *"are inferred from the required script set rather than stated by any input"*;
NFR3 names only Latin, CJK and Thai; **`ja` arrived from the J in CJK.** Dropping it was the main
alternative and was **rejected** because it would make Folio **reject a document it can render
correctly** — tag it `ja` and it is a load error, tag the byte-identical document `zh-Hans` and it
renders. The workaround would have been a one-character edit, so the refusal would have had no
operative content.

### The three obligations this ruling creates

1. **The `ja` golden is FORMATTING ONLY, and that is a scope fence that must be stated as one.**
   `2026年8月26日` and `1,234.56` — mechanically verifiable by anyone. AD-12's `[ASSUMPTION]` widens
   the locale set *"only with a golden fixture per locale"*; under this ruling **`ja`'s fixture covers
   dates and numbers, NOT typography.** The moment anyone claims Japanese *typography*, the golden
   needs a semantic acceptance step by a Japanese reader (**D-000.22**) — **and this project has no
   such person.** Thai works because the owner reads Thai (**D-2.4.3**); there is no equivalent for
   Japanese, and **D-000.41** says not to buy attributability with a scarce human we do not have.
   **Record this plainly in the code and the docs so a later story cannot quietly upgrade the claim.**
   (AC21.)
2. **The correct-typography path already exists and is DOCUMENTED, never built.** **AD-8** makes
   `FontSet` a public `map[string][]byte`, so a user may supply Noto Sans JP today, unchanged.
   **And this is the part that is easy to get wrong: the fallback chain is declared by the DOCUMENT,
   not derived from the locale** — so even a shipped JP face would not auto-select, and a `ja` author
   names the face either way. **Do NOT add locale→font inference.** That is a new mechanism nobody
   asked for. (AC21.)
3. **Do NOT add a runtime glyph-coverage diagnostic. Not as a Warning, not as a Caveat, not as a
   code.** The lead would decline to build one even had the owner chosen it: there is no coverage gap
   to diagnose and every glyph resolves, so a shape-*preference* warning would fire on **every**
   Japanese document, unconditionally, forever — **D-000.15**'s false positive, whose cheapest
   resolution is switching the guard off. Documentation is the right instrument for a known,
   permanent, universal limitation. **If a reviewer sees an AC or a code path heading this way,
   it is a defect.**

### Two referents to record and NOT act on

- **Story 5.4's payload itemisation is falsified, and it is not this story's to fix.** Measured now:
  **5.07 MB brotli -q 11** total (Latin 0.226 · Thai 0.025 · CJK 4.82; 11.29 MB raw), against 5.4
  AC2's stated 0.4 + 0.1 + 7.4 = 7.9 MB — roughly **2.8 MB of headroom** against NFR7's ~9 MB,
  assuming the still-unmeasured 1.5 MB engine and 0.12 MB dictionary. This independently confirms
  **DW-9**. Already on the lead's Epic 5 list, and **independent of the `ja` decision**. Note the
  referent; own nothing.
- **`ja` is already shipped at four sites** — `internal/template/closedsets.go:14`, `model.go:25`,
  `parse.go:85`, and a live fixture at `ac4_coverage_test.go:149`. Nothing narrows. **But AC4 must
  verify those sites agree with the locale table this story adds**: a closed set declared in two
  places is a set that drifts.

---

## Acceptance Criteria

> Every red-proof below must be **captured BEFORE** the thing it proves is wired (**D-000.30**):
> wiring closes the window permanently. Every mutant must be **honest** — it must compile, and it
> must be the mutation its comment claims (**D-000.61 extension**, and 3.3's Findings 12-14, which
> caught three mutants that were not what they said). Every assertion whose subject could be empty
> carries a **presence precondition** (**D-000.9**) that names what it counted.

### The seam and the determinism boundary

**AC1 — The formatting context is an explicit, immutable value threaded to `Eval`.**
Locale tag + fixed UTC offset reach `expr.Eval` as a parameter (R1), sourced at `folio/render.go`
from `doc.Locale` / `doc.UTCOffset`. **`expr.Resolver`'s method set is unchanged** —
`lint`'s `resolver-method-set-closed` and `TestExprResolverMethodSetIsClosed` both pass **with zero
edits to either guard or to `expectedResolverMethods`**. `internal/expr` holds no package-level
mutable state for this or anything else.
*Non-vacuity:* assert the two guards ran and found their subject (both already carry coverage-witness
stats; report them).

**AC2 — Host independence is proven by the EXISTING guards covering the new files, not by a new guard.**
No new host-independence check is added (R3, D-000.38, D-000.42). Instead: `lint`'s
`forbidden-imports` scan over `folio-go/internal/` reports **`FilesSeen` including every new file this
story adds** — asserted by name, not by a count that could pass while the new file was skipped — and
finds zero findings. **Red-proof, captured first:** add `import "time"` to the new formatting file and
show `forbidden-imports` reddens naming that file; add `math.Pow(10, 2)` to it and show the
`mathAllowedCalls` allow-list reddens naming `Pow`. Revert both.
*This AC is a coverage proof, not a new fence. Say so in those words in the test's own comment
(D-000.24).*

### The locale table

**AC3 — The table lives in `internal/expr` and no package is added under `internal/`.**
`lint`'s `stage-rank` rule passes with **zero edits to `stageRankTable`** (R2).

**AC4 — The table covers exactly `en`, `th`, `zh-Hans`, `ja`. ONE declaration of that set, in `internal/template`, consumed by `internal/expr`. *(RULED — **D-3.4.2**; see the wall below.)***

**The wall, measured, and the attractive wrong answer.** `lint`'s stage-rank table puts `template` at
**2** and `expr` at **3** (`stagerank.go:61-62`), so **`template` may not import `expr`** — and `expr`
**already imports** `template` (`internal/expr/decimal.go:15`, for `SplitJSONNumber`). The arrow runs
the **opposite** way to the intuitive one: the loader's closed set cannot read the locale table. Go
stops that with a compile error, so the compile error is not the hazard. **The hazard is declaring
the tag list a second time to get past it** — DW-8's duplication wearing a different type, and the
same wall D-3.2.1's knock-on hit at 3.2 when the rank guard forced `DeriveFooterOf` to the module
root.

**The ruling, as AMENDED on this story's pushback — the constants shape replaces the accessor-only
shape the lead first drew, because it removes the failure the accessor could only assert against:**
- **`internal/template` exports ONE NAMED CONSTANT PER TAG.** `closedLocales`
  (`internal/template/closedsets.go:13`) is built **from those constants**, and `internal/expr`'s
  locale table is **keyed by the same constants**. **Each tag string is spelled exactly once in the
  module** — there is no second literal to drift, by construction rather than by assertion.
- **`internal/template` also exports the ordered tag sequence**, and — see below — its **exact order
  is asserted as a literal**.
- **The sequence is a declared ordered slice; it is NEVER produced by ranging a map.** Under
  **D-1.3.5**'s total ban on ranging a map under `internal/`, a ranged map here is a **build
  failure**, not merely a determinism risk. (The lead's note: this constraint should have been in the
  original ruling.)
- **This is semantically right, not merely forced, and the story says so** so the split reads as
  intentional: **the closed set of valid TAGS is a `.folio` FORMAT constraint** — `folio-format.md:48`
  states it and the load error is raised in `template` — **while the per-locale FORMATTING DATA
  (symbols, calendar, patterns) is `expr`'s behaviour**, exactly as the spine's source tree says:
  *"`expr/` — … locale tables (AD-12)"*. **The set is format; the table is behaviour. AD-12 is not
  amended.**

**AC4a — the ORDER is asserted as an exact literal sequence, not as "some deterministic order"**
(**D-000.45**). Pin the sequence itself. Otherwise a later switch from source order to
`slices.Sorted` reorders the exported sequence **silently**, and anything downstream that happens to
depend on that order moves with it — the drift is invisible precisely because both orders are
"deterministic".

**AC4b — the SET-DIFFERENCE-BOTH-WAYS assertion is KEPT, and the constants do NOT make it
redundant.** This is the part it would be easiest to drop now that spelling is handled by
construction, and dropping it would be wrong: **the constants kill SPELLING drift, not MEMBERSHIP
drift.** A tag constant with no table entry, and a table entry keyed by a constant absent from
`closedLocales`, are both still constructible — and **this check is now the only thing covering
membership at all.**
**Not a count, not a length check.** A count is a lossy set, and the two failure modes are
**different bugs**: *a declared tag with no table entry* is a tag nobody can format; *a table entry
with no declared tag* is formatting data for a tag the loader rejects. **Assert both directions, and
name each in its own message.**
**This one is NOT an AST scan and must not be over-engineered into a type-checker rule** — both sides
are real values at test time, so it does not inherit the alternate-spelling problem that defeated
3.3's method-set check.
*Non-vacuity:* the test fails if either side is empty, and states what it compared.
*Red-proof:* remove one tag's table entry → red naming that tag, **in the first direction only**; add
a table entry for an undeclared tag → red naming it, **in the second direction only**. **Both reds
must be captured, because a single-direction check passes half of this AC vacuously.**
**The `ja` row is built like the other three — no fence.** For everything this story implements it is
effectively identical to `zh-Hans`; if the two rows share data, they share it **by construction**
(one declaration referenced twice), never by two copies that can drift.

**The other `ja` sites are CONSUMERS, not declarations. Do not "fix" them into a third and fourth
source of truth:**
- **`internal/template/model.go:25`** and **`parse.go:85`** are a doc comment and an error message.
  Correct them only if the set itself ever changes (**D-000.35**). It does not change here.
- **`folio-go/ac4_coverage_test.go:149`** is a fixture declaring `"locale": "ja"`. Under this story's
  ruling it is now a **live regression subject** — it proves a `ja` document **loads**. **It must be
  exercised, and it must not be deleted as redundant** once the locale table has tests of its own.
  Say so where a future reader will find it.

**AC5 — An unlisted tag is a load error, and there is no silent fallback at any layer.**
Two arms, both required, because either alone is vacuous:
(a) **Load layer, already shipped and re-pinned:** a document with `locale: "xx"` fails to load
naming the closed set (`parse.go:84-86`), and a document with **`locale` present but `utcOffset`
absent** fails to load (F6) — pinned so the offset never acquires a default.
(b) **Formatter layer, new:** the table lookup for an unknown tag returns a **located error**, never
a default entry, and the formatting context's **zero value is not a usable `en`** — constructing a
formatter from a zero context is an error, not a silent English document.
*Red-proof for (b), captured first:* make the lookup fall back to `en` and show arm (b) reddens
**while arm (a) still passes** — that divergence is the whole reason both arms exist.

**AC6 — The table carries a version, and it is part of the library version (AD-22).**
A declared table-version constant exists in `internal/expr`, is surfaced wherever the library version
is surfaced, and a test states in words that changing the table is a breaking change under AD-22.
*No mechanism is claimed beyond declaration — say so (D-000.24). This is a label, not a guard.*

### `formatDate`

**AC7 — Only an RFC 3339 string or an integral epoch-millisecond number is accepted; everything else is a located error.**
Accepted: `KindString` matching RFC 3339 (with its own offset or `Z`), and `KindNumber` whose
`Decimal` denotes an **exact integer** number of milliseconds. Rejected, each with a located error
naming the element and the offending value: a non-RFC-3339 string; a **non-integral** number
(`1.5`); `KindBool`; `KindNull`; an absent path (which is already the resolver's own located error,
propagated unchanged).
**The epoch-millisecond number never becomes a `float64` on the way in** — it arrives as
`expr.Decimal` through 3.3's `UseNumber` discipline (`bind.DecodeData` / `NewDecimal`), and the
integrality test is on `Coefficient`/`Exponent`, never via a float conversion.
*Red-proof, captured first:* a value whose `float64` round-trip is lossy (a millisecond count with
17+ significant digits) is accepted-or-rejected identically before and after; and a deliberate
`float64(ms)` conversion in the accept path reddens the exactness assertion.

**AC8 — The instant is converted to civil date fields in the DOCUMENT's declared offset, with no `time` import anywhere.**
Order is fixed and asserted: **(1)** the input is reduced to an instant (an RFC 3339 string's own
offset is applied here; an epoch-ms number is already UTC); **(2)** the **document's** fixed
`utcOffset` is applied to obtain civil year/month/day/hour/minute/second; **(3)** the pattern renders
those fields. The calendar is **proleptic Gregorian**, computed by integer arithmetic in
`internal/expr` — `import "time"` appears nowhere (AC2's guard covers this).
*Assertions include:* a leap day (`2024-02-29`); a year that is a multiple of 100 but not 400
(`1900-02-28` / `2100-03-01`); and a **date-line case** — an instant at `2026-12-31T23:30:00Z`
rendered under `utcOffset: "+07:00"` yields **1 January 2027**, not 31 December 2026.
*Bounded domain:* a supported civil-year range is declared as a named constant and a value outside it
is a located error, never a wrapped or negative year. State the bound; do not leave it implicit.

**AC9 — For `th`, the Buddhist era is a CALENDAR conversion applied to the year field, never an edit to a formatted string.**
`BE = CE + 543`, applied **after** AC8's step (2) and **before** the year is rendered, at the single
place the year field is produced. **No implementation may post-process a rendered date string.**
*The year-boundary case is the assertion that matters, and it is required:* the same instant as in
AC8 (`2026-12-31T23:30:00Z`, `utcOffset: "+07:00"`) renders under `th` as **`1 มกราคม 2570`** — the
offset shift crosses the year boundary first and the era conversion then sees 2027, giving 2570. A
naive "add 543 to whatever year the UTC instant had" produces 2569 and is wrong; **that is the named
mutant.**
*Also assert the reference's own published example:* `2026-08-15` under `th` with `"d MMMM yyyy"`
renders **`15 สิงหาคม 2569`**, and under `en` renders **`15 August 2026`** (F7).
*Red-proof, captured first:* implement the era as `strings.Replace` on the rendered year and show the
year-boundary case reddens while the August case still passes — the vacuity that makes this AC
necessary.

**AC10 — The pattern vocabulary is CLOSED and validated at `Check` (load time), not at `Eval`.**
The accepted date-pattern and number-pattern grammars are declared once, in one place, and a pattern
literal outside them is rejected by `expr.Check` — so a bad pattern is a **load** error, consistent
with `checkNumberLit`'s precedent (*Do not re-open* 9). The published vocabulary is the floor:
`"d MMMM yyyy"`, `"dd/MM/yyyy"` (`docs/folio-mvp-plan.md:418`), `"#,##0.00"`, `"#,##0"`. Whatever the
developer admits beyond these is declared explicitly, in one literal, and is what AC17 measures
against.
**AC10a — the DATE-pattern grammar, RULED (FINDING-4). No quoting mechanism ships.**
A date pattern is a sequence of **field tokens**, drawn from a **closed set**, and **literals**.
- **A literal may be any character that is NOT an ASCII letter.** So `年`, `月`, `日`, `/`, `-`, space
  and punctuation all pass through verbatim, and the owner's ruled `ja` golden `"yyyy年M月d日"` needs
  **no machinery at all**.
- **Any ASCII letter that is not part of a recognised token makes the pattern a located LOAD error,
  naming the pattern and the offending character** (FR41, AD-14).
- **No quoting mechanism ships. No `'…'`, no backslash, nothing to escape.** Do not add one, and do
  not add one "for symmetry" with the number grammar.

**Why, recorded so nobody re-opens it:** every pattern the four locales need is already expressible,
and the decisive property is **forward-compatibility in the one direction that matters** — patterns
containing stray ASCII letters are **errors today**, so adding CLDR-style `'…'` quoting later breaks
**nothing that works**. Shipping quoting now and removing it later would break working documents. It
also keeps the whole check **decidable at load with no data**, which **Story 3.7's `folio.Validate`
needs**.

**Instrument:** the accepted **token set is closed and asserted by AST set-equality — the same
instrument as the function table.** **Not a regex over pattern text, and not a name list.**
*Red-proof:* add a token to the parser without adding it to the asserted set → red; and the reverse.

**AC10b — `formatNumber`'s pattern grammar is a SEPARATE closed set**, with its own declaration and
its own assertion. It is not the date grammar with different letters, and it still owes **D-3.1a.1**'s
forcing assertion — see **AC17**, with `avgExtraScale` cited **by symbol**.
*Required consequence, F3:* `render_bind_test.go:210`'s `{{formatNumber(a, "x")}}` becomes a **load**
error. **Re-point that test, do not delete it** — its subject (a wrong-kind pattern is caught, and
caught with a located message) survives; only the stage moves. The two goldens binding `"#,##0.00"`
(`testdata/template/golden/worked-example.json:19`, `internal/template/fixtures_test.go:263`) must
keep loading **unchanged**, and that they do is this AC's non-vacuity check — a validator that
rejected everything would satisfy the first half alone.

### `formatNumber`

**AC11 — Grouping symbols, decimal symbol AND the DIGIT SET come from the table entry for the document's tag.**
Group separator, decimal separator, grouping size and **digit set** are read from the table, per tag
— never hardcoded at the formatting site, never taken from the host. Asserted for **all four** tags.
*Red-proof:* change one tag's separator in the table and show only that tag's expected output moves.

**AC11a — OWNER DECISION (DECISION-2): Western (ASCII) digits for every locale, in `formatDate` and
`formatNumber` alike** — `15 สิงหาคม 2569`, `1,234.56`. **Shipped as an EXPLICIT NAMED FIELD in the
locale table, never as an implicit default**, so a later change to Thai numerals is a **visible,
versioned table edit** rather than a silent code change. That is what **AD-22** needs from a versioned
table (AC6), and it is the operative half of this ruling — the value and its explicitness are equally
binding.
> **Recorded because the near-miss matters more than the outcome.** The coordinator nearly took this
> decision itself on the strength of `docs/expression-reference.md` already showing Western digits.
> **The lead blocked that: that page was written at 3.2 as documentation for behaviour that did not
> exist** — **D-000.28**'s anticipatory-boilerplate class, *"false from birth and reading identically
> to a true one"* — **and it has already been found wrong on two other counts.** An agent's
> unverified guess is not a shipped product decision. The owner ruled independently and happened to
> agree. **Do not treat any other statement on that page as ratified merely because it is published**
> — F7's canonical outputs are this story's specification because the owner's own `ja` ruling and
> this decision re-grounded them, not because they were printed.

**AC12 — Scaling by powers of ten uses an INTEGER LOOKUP TABLE, and the instrument is not a name-list.**
A package-level table of integer powers of ten, indexed by exponent, bounded by AC10's declared
maximum fraction digits. `formatNumber`'s own scaling never calls `tenPow` (`reduce.go:50`, which
*computes* via `big.Int.Exp` rather than looking up) and never calls anything from `math`.
**Three layers, honestly labelled (D-000.23, D-3.1a.1's shape):**
- **Layer 1 — behavioural, PRIMARY, has teeth.** An **exactness oracle**: `formatNumber` over a
  fixture corpus whose subjects are chosen so that **a `float64` implementation demonstrably
  diverges** (per D-000.50, the population is checked before the assertion is written — do not assume
  `0.1 + 0.2`; measure which of the corpus's own values actually diverge, and record how many did).
  This is the layer that cannot be defeated by a synonym: `math.Pow`, `math.Exp(n*math.Log(10))`, a
  `float64` multiply loop and a `big.Float` all fail it for the same reason, because the property
  asserted is **exactness of the output digits**, not the absence of an identifier.
- **Layer 2 — the existing type guards, cited by symbol, never re-implemented** (D-000.42):
  `lint`'s `ScanFloatTypedValues` (`floattyped.go`), its `big.Float`/`big.Rat` type-identity denylist
  (`bigfloattype.go`), the AST float scan (`internal/arch_test.go`), and `mathAllowedCalls`. Cite
  their rule ids; add nothing.
- **Layer 3 — the lookup table's own contents.** The table's entries are asserted to be exactly the
  powers of ten, **verified against an independently computed `big.Int` chain** — a different route
  to the same facts (D-000.38), so a transcription slip in the literal cannot agree with itself.
*Why a bare "never `math.Pow`" AC would be insufficient, in the test's own comment:* three
instruments have been defeated by a legal alternate spelling in this epic alone — `big.Float` past a
`float64` name match, a string literal past a "must be non-literal" check, an embedded interface past
AST set-equality. **Layer 1 is stated as the coverage; Layers 2 and 3 are stated as narrow.**

**AC13 — `formatNumber` of a `Decimal` is exact, and rounds half-to-even at the pattern's scale.**
Consistent with `AvgDecimals`' documented rounding (D-3.1a.1). A value carrying **more** fractional
digits than the pattern requests is rounded half-to-even, once, at the pattern's scale — never
truncated, never double-rounded. A value carrying **fewer** is padded with zeros to the pattern's
scale. Assert the tie cases in both directions (`…5` rounding up to even and down to even).

**AC14 — The two adjacent kernel zeros render as the pattern says, and neither kernel exponent moves.**
`SumDecimals([])`'s `{0, 0}` under `"#,##0.00"` renders `0.00`; an all-null `avg`'s
`{0, -avgExtraScale}` under `"#,##0.00"` renders `0.00` as well. **This is the presentation layer
doing the harmonising that *Do not re-open* 6 forbids doing at the kernel.** Assert both, and assert
that `reduce.go`'s two exponents are **unchanged** at hand-off.

### The `avgExtraScale` honesty-keeper — FLAG-1's discharge

**AC15 — The constant is cited BY SYMBOL. The literal `4` appears in no assertion this story adds.**
Every reference is `avgExtraScale`. A test that writes `4` on both sides of a comparison proves
nothing (D-000.14/D-000.36) and is a defect on its own merits.

*(AC16 is the table-flip guard and is stated below under **The table flip**, deliberately kept next to
the guards it edits rather than here.)*

**AC17 — No `formatNumber` pattern the implementation accepts can request more fractional digits than `avg` produces. Derivationally.**
`avg` produces `maximum operand scale + avgExtraScale` fractional digits, and the maximum operand
scale is `≥ 0`, so **`avg` always produces at least `avgExtraScale` fractional digits**. The
obligation is therefore: **the maximum fraction-digit count the pattern validator ACCEPTS is
`≤ avgExtraScale`.** Three parts, all required:
1. **The bound is a named constant** in the pattern validator (AC10's declaration), and the assertion
   is the symbolic inequality `maxPatternFractionDigits <= avgExtraScale`. No literals.
2. **The constant must genuinely describe the accepted set, not merely be declared.** A property
   assertion, derived from the validator's own behaviour: for every `k` from `0` to
   `maxPatternFractionDigits` the validator **accepts** a pattern with `k` fraction digits, and at
   `maxPatternFractionDigits + 1` it **rejects**. *Without this part, part 1 is a comparison between
   two declarations and is exactly the vacuity D-3.1a.1's Flag F3 has been waiting for someone to
   avoid.*
3. **Red-proof, captured BEFORE the wiring (D-000.30):** set `maxPatternFractionDigits` to
   `avgExtraScale + 1` and show **both** part 1 and part 2 redden. Record the output. Revert.
*If the developer finds a pattern the product genuinely needs that requires more than `avgExtraScale`
fractional digits, `avgExtraScale` moves — that is D-3.1a.1's design working, not a regression here.
Escalate rather than silently narrowing the pattern vocabulary to fit the constant.*
**On discharge, this story's Delivery Log records FLAG-1 as LIFTED, naming this AC** (D-000.60).

### The table flip

**AC16 — RULED ARM A: the registered-but-unimplemented machinery is REMOVED, and replaced by a two-half assertion of the OBLIGATION (F1, D-000.59).**

**Deleted, in one commit:** `funcEntry.implemented`; `funcEntry.owningStory`; `eval.go:57-66`'s
`if !entry.implemented` branch; `TestUnimplementedFunctionsAreLocatedErrors`;
`TestUnimplementedAndUnknownFunctionErrorsAreDistinguishable`; and
`TestUnimplementedEntriesHaveNoEvalCallBranch` **together with its `== 0` precondition**.

**Why Arm A and not "keep it inert":** Arms B and C both leave a mechanism asserting a proposition
that **can never again be false** — the table is closed at eight (C1), so the unimplemented population
**cannot refill**. That is **D-000.9 with the polarity inverted**: not an all-clear indistinguishable
from a could-not-look, but **a guard that cannot fail, reported as coverage.**

**The replacement, and the anti-vacuity clause is the whole point** (**D-000.59**: assert the
**obligation**, never the **event**). ***"All eight are implemented"* is the EVENT** — it merely
re-reads the flag this story just set, which is why it is not the replacement. **The OBLIGATION is
that a registered function actually COMPUTES.** Two halves, both required, because either alone is
defeatable:

1. **Structural — and it already exists.** `TestImplementedEntriesMatchEvalCallSwitch`
   (`table_derivational_test.go:98`) becomes plain **set equality** between `functionTable`'s names
   and `evalCall`'s switch cases: **every entry has a branch, every branch has an entry.** Its
   presence precondition at `:121` **stays and stays safe** — that population goes to eight and never
   to zero. Likewise `:85`. **Only the `:159` precondition goes, and it goes with its test.**
2. **Behavioural — because a branch can exist and still refuse.** Each of the eight, called with
   valid arguments, returns **a value and no error**. **The coverage witness is the count exercised,
   derived from `len(functionTable)`** — never a literal (**D-000.14**).

**Red-proof, captured before the deletions (D-000.30):** delete one `evalCall` branch → half 1 reds
naming it; make one branch return an error for valid arguments → **half 2 reds while half 1 still
passes**. That second red is what proves the two halves are not one check.

**Do NOT relax, re-tune or re-add the `== 0` precondition in any form.** Removing the mechanism
removes it; a softened version would be one more thing asserting what can never again be false (F1).

**Also correct the file's own top comment**, which claims it survives 3.4 with no edit at all — false
for exactly one of its three guards (F1). A comment a reviewer can falsify by reading the file below
it is a defect (**D-000.24**).

**AC18 — Per-locale goldens, at the evaluation layer, for ALL FOUR locales: `en` · `th` · `zh-Hans` · `ja`.**
One golden per locale, covering both functions, pinned as recorded expected strings. **AD-12's
`[ASSUMPTION]` — widen the locale set "only with a golden fixture per locale" — is satisfied here for
all four.** The ruled `ja` values are **`2026年8月26日`** and **`1,234.56`**.
**The `ja` golden's SCOPE FENCE is part of the AC, not a footnote:** it covers **date and number
FORMATTING ONLY, never typography**. Its own comment must say, in those words, that it is
mechanically verifiable by any reader and that **no claim about Japanese typography is made or
asserted anywhere**, because such a claim would require a semantic acceptance step by a Japanese
reader (**D-000.22**) that this project cannot staff (**D-000.41**; Thai works only because the owner
reads Thai, **D-2.4.3**). Write the fence where a later story will hit it, so nobody upgrades the
claim by accident.
*Non-vacuity:* the goldens must **differ from each other** in at least the era, the month name and
the date shape; a test where all four expected strings are identical proves the locale is being read
nowhere. **`ja` and `zh-Hans` may legitimately agree** on numbers and on the `年月日` shape — where
they do, say so explicitly rather than letting the agreement read as coverage.

**AC19 — End-to-end renders through the public `Render` API, producing real PDF bytes: `th` and `ja`.**
`worked-example.json` (`testdata/template/golden/`) currently **loads but cannot render** because it
binds an unimplemented `formatNumber` (`internal/expr/doc.go:35`). **That is now false.** Land the
end-to-end assertions, and update `doc.go`'s package comment, which states the old fact in three
places (`:7-10`, `:35`, `:45`).
**The `ja` end-to-end asserts that the document RENDERS with the shipped font set — complete,
every glyph resolved — and asserts nothing about glyph shape.** The measured `cmap` coverage above is
why that render succeeds; the Simplified-Chinese kanji forms a Japanese reader would see are the
documented limitation of AC21, not a defect this test may report.
**Do NOT add a runtime glyph-coverage diagnostic here or anywhere.** There is no coverage gap to
diagnose, so a shape-preference warning would fire on every Japanese document unconditionally and
forever — **D-000.15**'s false positive, whose cheapest resolution is switching the guard off.

**AC20 — Both docs twins are updated in the same commit (D-000.47).**
`docs/expression-reference.md` and `docs/expression-reference.html`: lift the *Pending*/"not yet
implemented" markers on `formatDate`/`formatNumber` (`.md:169`, `:211`; `.html:290-291`, `:461`,
`:471`, `:498`); add the worked `formatNumber` example to the *Totals* section that `.md:133-135`
explicitly reserves for this story; keep the published outputs at `.md:171-198` exactly as they are —
they are this story's specification (F7), not free text. **The four-locale list stays, and all four
are now implemented**, so the *Locale* paragraph at `.md:189-198` needs no hedge.
*A twin-drift check already exists for these two files — run it and report it.*

**AC21 — The Japanese typography limitation is stated in writing, in the project's own voice, and the correct-typography path is DOCUMENTED rather than built.**
One honest sentence — the developer's own words, not a transcription — lands in **`folio-format.md`**
and against **AD-12** in the architecture spine, to this effect: *the shipped font set renders
Japanese completely, in Simplified-Chinese kanji shapes; supply your own face via `FontSet` for
Japanese typography.* It must be findable by someone reading either document about locales, and it
must not overstate (the coverage is complete) or understate (the shapes are regionally wrong).
**Two things it must get right, both of which are easy to get wrong:**
- **`FontSet` is already the path** — a public `map[string][]byte` (**AD-8**), so a user supplies Noto
  Sans JP today with no library change. Nothing is built for this.
- **The fallback chain is declared by the DOCUMENT, not derived from the locale.** Even a shipped JP
  face would not auto-select; a `ja` author names the face either way. **Do not add locale→font
  inference** — a new mechanism nobody asked for, and one that would make the render a function of
  something other than what the document says.
*No mechanism is claimed by this AC. It is documentation, and documentation is the right instrument
for a known, permanent, universal limitation (D-000.24: say what you are not claiming).*

---

## Decisions raised at creation — ALL RULED. Nothing here is open.

*Kept as a record rather than deleted: a developer who reads only the ACs should still be able to see
what was considered and rejected, and why, without re-opening it (D-000.49).*

### DECISION-1 — the registered-but-unimplemented machinery. **RULED: ARM A, remove it.**

**The situation.** After this story, `funcEntry.implemented` can never be `false`,
`funcEntry.owningStory` is read by nothing that can fire, and `evalCall`'s `if !entry.implemented`
branch (`eval.go:57-66`) is unreachable from any table entry. Two shipped tests have
`formatDate`/`formatNumber` as their subjects and go red.

- **Arm A — remove the machinery. CHOSEN.** *Grounds:* the table is closed at eight (C1), so the
  unimplemented population **cannot refill**; Arms B and C both leave a mechanism asserting a
  proposition that can never again be false — **D-000.9 with the polarity inverted: a guard that
  cannot fail, reported as coverage.**
- **Arm B — keep it inert as defence-in-depth.** *Rejected.* Smallest diff, but an inert
  `owningStory: ""` on eight rows invites a future reader to believe the mechanism is live.
- **Arm C — keep only the branch as an unreachable safety net.** *Rejected.* `evalCall`'s existing
  `default:` case is reachable by a *coding* mistake; this branch would be reachable by nothing. The
  two are not the same shape, and the resemblance is what makes C attractive.

**Executed by AC16**, under D-000.59's replacement discipline, with the obligation/event distinction
spelled out there. **This ruling is also F1's fix** — removing the mechanism removes its expiring
`== 0` precondition with it.

### DECISION-2 — digits for `th`. **OWNER DECISION: Western (ASCII), everywhere, as an explicit table field.**

`15 สิงหาคม 2569`, `1,234.56`, in `formatDate` and `formatNumber` alike. Arm B (Thai numerals for
`th`) is correct for some Thai government forms and wrong for the stated use case — a bank statement.
**Executed by AC11a**, including the requirement that the digit set be an **explicit named field**
rather than an implicit default, which is the half AD-22 actually needs.

**Recorded for the record, because the near-miss is the more useful half:** the coordinator nearly
took this itself on the strength of `docs/expression-reference.md` already showing Western digits.
**The lead blocked that — that page was written at 3.2 as documentation for behaviour that did not
exist** (**D-000.28**, *"false from birth and reading identically to a true one"*), and it has already
been found wrong on two other counts. **An agent's unverified guess is not a shipped product
decision.** The owner ruled independently and happened to agree.

### FINDING-4 — the date-pattern grammar. **RULED: field tokens plus non-ASCII-letter literals; no quoting ships.**

Executed by **AC10a**. A stray ASCII letter is a located load error naming the pattern and the
character; the token set is closed and asserted by **AST set-equality**, not a regex.
**`formatNumber`'s pattern grammar is a separate closed set** (AC10b) and still owes AC17.

### FINDING-3 — where FLAG-1 lives. **RULED: nothing moves.**

Executed by **F4**. `deferred-work.md` holds deferrals with owners; FLAG-1 is an obligation a ruling
places on a later story, and D-3.1a.1 already records it correctly. Widening that file's scope would
make it scannable for neither kind.

---

## Flags — and what each one SUPPRESSES (D-000.60)

- **FLAG-A — DISSOLVED, not carried. The `ja` fence is LIFTED, in this file, before development
  starts.** The creation draft fenced the `ja` arm pending a font-coverage measurement. **The lead
  measured the shipped `cmap` tables directly and the owner ruled: keep `ja`, no fence** — so the
  fence is gone from AC4, AC18, AC19 and AC20 and all four locales are built and asserted here.
  **D-000.60** requires a dissolved fence to be visible *as dissolved* rather than simply absent,
  which is why this line survives its own suppression: the premise it rested on (*"coverage must be
  verified"*) was **verified and found false as a worry** — 93/93 hiragana, 96/96 katakana, complete.
  **What replaced it is narrower and is not a fence on the work:** AC18's `ja` golden is
  **formatting-only, never typography**, because a typography claim needs a semantic acceptance step
  (**D-000.22**) by a reader this project does not have (**D-000.41**).
  *Suppresses:* any later upgrade of the `ja` golden into a typography claim; any runtime
  glyph-coverage diagnostic (AC19); any locale→font inference (AC21).
  *Does NOT suppress:* anything at all about building, rendering or asserting `ja` **formatting**.
- **FLAG-B — DISSOLVED. `DECISION-1` is RULED (Arm A: remove the machinery).** It suppressed starting
  AC16's deletions on a guess; the deletions are now specified, and **their red-proofs must still be
  captured first**. *Residual suppression, carried:* **do not relax or re-add the `== 0` presence
  precondition in any form** — the ruling removes it with its test, and a softened version would
  reinstate exactly what F1 names.
- **FLAG-C — DISSOLVED. `DECISION-2` is RULED (OWNER: Western/ASCII digits, everywhere).** It
  suppressed writing any `th` expected-output golden containing digits, and writing the digit set as
  an implicit assumption. **The second half survives as a live requirement, not a suppression:** the
  digit set ships as an **explicit named table field** (AC11a). *Residual suppression, carried:* do
  not treat any other statement in `docs/expression-reference.md` as ratified merely because it is
  published — D-000.28, and that page has already been wrong three times.
- **FLAG-D — FLAG-1 (`avgExtraScale = 4`, MEDIUM confidence, carried from Story 3.1a's Flag F3
  through Story 3.3) is DISCHARGED HERE, by AC17 and only by AC17.** *Suppresses:* any claim that
  landing `formatNumber` at all "confirms" the 4. Only AC17's parts 1+2+3 together do. **If the
  developer cannot satisfy AC17 without narrowing the pattern vocabulary below what the product
  needs, that is the signal `avgExtraScale` must move — escalate, do not narrow.**
- **FLAG-E — This story is judged NOT hash-shaped and does not run the cross-target matrix**
  (D-000.4). *Suppresses:* carrying that judgement forward silently — the Delivery Log must state the
  suites measured **and** name the unrun matrix explicitly, due at Epic 3's close.

---

## Task breakdown

**All 15 items below are complete.**

1. [x] **Re-measure the baseline**, all three modules, and record scope + flags + numbers (D-000.26).
   Confirm `TestCorpusMeetsP6ExerciseFloors`' stats are byte-identical to the declared set.
2. [x] **Capture every red-proof listed in the ACs, BEFORE any wiring** (D-000.30): AC2's two import/
   selector reds, AC5(b)'s fallback red, AC9's string-replace era red, AC10's pattern reds, AC12's
   float-divergence red, AC16's positive-assertion red, **AC17's `avgExtraScale + 1` red**. Record
   each output verbatim. Confirm each mutant compiles and is what its comment says (D-000.61 ext).
3. [x] *(Was: escalate DECISION-1 and DECISION-2.)* **Both are RULED — see *Decisions raised at
   creation*. Nothing in this file is open; do not re-escalate either.** Read AC16 (Arm A) and AC11a
   (Western digits, explicit table field) before writing code that touches them.
4. [x] Build the formatting-context value and thread it: `render.go:571` → `bind` → `expr.Eval` (R1).
   Confirm `resolver-method-set-closed` and `TestExprResolverMethodSetIsClosed` are untouched.
5. [x] **Land AC4's single declaration first**: one exported named constant per tag on
   `internal/template`, `closedLocales` built from them, the exported ordered sequence (a declared
   slice — **never a ranged map**, which is a D-1.3.5 build failure), AC4a's **exact literal order**
   assertion, and AC4b's both-directions set difference with **both** reds captured. Only then build
   the locale table in `internal/expr` **keyed by those same constants** (R2, AC3-AC6), all four rows
   including `ja`. Verify `ac4_coverage_test.go:149` is exercised and mark it a live regression
   subject.
6. [x] Build the integer power-of-ten lookup table and `formatNumber` (AC11-AC14), then AC12's three
   layers in order — Layer 1 first, and check its population before writing the assertion (D-000.50).
7. [x] Build the hand-rolled civil calendar and `formatDate` (AC7-AC9), era conversion at the year field.
8. [x] Build the two **separate** pattern validators at `Check` — the date grammar per AC10a (field
   tokens plus non-ASCII-letter literals, **no quoting**, token set asserted by AST set-equality) and
   `formatNumber`'s own closed set per AC10b. Then re-point `render_bind_test.go:210` (F3) and
   confirm the two `"#,##0.00"` goldens still load unchanged.
9. [x] **AC17 — the `avgExtraScale` honesty-keeper.** Parts 1, 2 and 3.
10. [x] Flip the two table entries; execute **AC16 (Arm A)** — the six deletions, the structural set
    equality at `:98`, the behavioural all-eight-compute half with its witness derived from
    `len(functionTable)`, and both red-proofs captured first. Keep the `:85` and `:121` preconditions;
    remove `:159` **with its test**. Correct `table_derivational_test.go`'s top comment.
11. [x] AC18 goldens — **all four locales**, with the `ja` golden's formatting-only scope fence written
    into its own comment; AC19's two end-to-end renders (`th`, `ja`); and `internal/expr/doc.go`'s
    three stale statements.
12. [x] AC20 — both docs twins, same commit; run the twin-drift check. **AC21** — the Japanese typography
    sentence in `folio-format.md` and against AD-12 in the spine, in the project's own voice.
13. [x] **Run the full three-module gate.** Report each module's numbers with its scope and flags, name
    the required red and its stats, and name DW-19's three local-only `lint` reds. **A fourth `lint`
    failure is this story's regression.**
14. [x] **Delivery Log:** state the suites measured and the unrun matrix (FLAG-E); record **FLAG-1 as
    LIFTED, naming AC17** (FLAG-D, D-000.60); record **FLAG-A, FLAG-B and FLAG-C as DISSOLVED before
    development**, each naming what replaced it and what residual suppression it carries; record the
    two referents this story owns nothing of (Story 5.4's falsified payload itemisation; DW-9's
    independent confirmation).
15. [x] Set the sprint key to `review`.

> **The developer does not commit.** The breakdown ends at *status → review*.

---

## Delivery Log

**Suites measured (D-000.4/D-000.64, three modules).** All measured after the full implementation
landed, against the same invocations the baseline used:

| scope | invocation | result |
|---|---|---|
| `folio-go/` | `go build ./... && go vet ./... && gofmt -l . && go test ./... -count=1` | **819 pass · 1 fail · 1 skip**, gofmt clean |
| `lint/` | `go test ./... -count=1` | **97 pass · 3 fail** |
| `hashmatrix/` | `go test ./... -count=1` | **3 pass** |

- **`folio-go`'s one failure is the declared REQUIRED red**, `TestCorpusMeetsP6ExerciseFloors`, stats
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — **byte-identical** to the baseline declared
  at `c2b4fe2`. Untouched by this story, per D-000.17/D-2.1.14/D-000.57.
- **`lint`'s three failures are DW-19**, all in `.font-sources/` (gitignored, local-only, CI-green):
  `TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist`, `TestFontsAssetsNoticeRemovalRedProof`.
  **No fourth failure** — the count rose from the baseline's 95 pass to **97 pass** because this story
  added two new passing tests (`TestForbiddenImportsScanSeesStory34Files`,
  `TestForbiddenImportsRedProofOnStory34File`, AC2's own coverage witness and red-proof), not because
  any DW-19 fixture changed.
- **The cross-target hash matrix was NOT run** (FLAG-E): judged, at creation, not hash-shaped — this
  story adds no font byte, no image byte, no glyph and no new PDF operator class. Due at Epic 3's
  close.

**Red-proofs captured** (AC2, AC5b, AC9, AC10a, AC10b, AC12, AC16, AC17 — D-000.30/D-000.61 ext), each
either as a permanent, self-contained test in the shipped suite (the mutant and the real
implementation compared inside the same test, e.g.
`TestFormatDateThaiEraStringReplaceRedProof`, `TestNumberPatternRedProofOverAvgExtraScale`,
`TestUnlistedLocaleRedProof`, `TestDateFieldTokenSetRedProofAddedToken`,
`TestForbiddenImportsRedProofOnStory34File`) or, for AC16 and AC5b specifically, ALSO captured by hand
against the real committed source (deleted, mutated, ran, reverted, diffed byte-identical against a
backup copy before continuing):
- **AC16, half 1 (structural):** deleting the `"avg"` case from `evalCall`'s switch reddened BOTH
  `TestImplementedEntriesMatchEvalCallSwitch` ("avg: functionTable has this entry, but evalCall's
  switch has no case for it") AND `TestAllTableEntriesActuallyCompute` ("avg: … internal: \"avg\" has
  no evaluator"). Reverted; `diff` against the pre-mutation copy confirmed byte-identical restoration.
- **AC16, half 2 (behavioural):** making the `"avg"` case return an error for valid arguments reddened
  ONLY `TestAllTableEntriesActuallyCompute` ("avg: … valid call \"avg(t.a)\" returned an error instead
  of computing") while `TestImplementedEntriesMatchEvalCallSwitch` still PASSED — proving the two
  halves are not one check. Reverted and diffed byte-identical.
- **AC5b:** mutating `lookupLocale`'s unknown-tag branch to fall back to the `en` table entry reddened
  `TestUnlistedLocaleIsLocatedErrorNeverDefault` ("expected a located error for an unlisted locale
  tag") while the LOAD layer's own arm (a) test, `internal/template`'s
  `TestClosedLocalesMatchesLocaleTags`, still PASSED — the divergence AC5's two arms exist to prove.
  Reverted and diffed byte-identical.
- **AC2:** captured as `TestForbiddenImportsRedProofOnStory34File` (`lint/internal/rules/`), over a
  scratch copy of `calendar.go`: injecting `import "time"` reddens naming `calendar.go` and `"time"`;
  injecting `math.Pow(10, 2)` reddens naming `"Pow"`. Never touches the committed file.
- **AC9:** captured as `TestFormatDateThaiEraStringReplaceRedProof` — a naive "add 543 to the raw UTC
  year" mutant computes `2569` on the year-boundary case (`2026-12-31T23:30:00Z`, `+07:00`), while the
  real implementation computes the correct `2570` (era conversion applied to the civil year AFTER the
  offset shift).
- **AC10a:** captured as `TestDateFieldTokenSetRedProofAddedToken` — `parseDatePattern` correctly
  rejects `"EEEE"`, a token the real `dateFieldTokens` literal does not declare.
- **AC12 Layer 1:** captured as `TestFormatNumberExactnessOracleCorpusDiverges` — the corpus's own
  members are confirmed (D-000.50) to diverge from exactness under an IEEE-754 binary64
  representation (measured by integer bit-length reasoning about binary64's 53-bit mantissa, never by
  executing `float64` arithmetic — AD-23's ban is module-wide, including `_test.go` files, per
  `TestNoFloat64UnderModule`) before the exactness oracle (`TestFormatNumberExactnessOracle`) is
  relied on.
- **AC17:** captured as `TestNumberPatternRedProofOverAvgExtraScale` — a pattern requesting
  `avgExtraScale + 1` fraction digits is confirmed rejected by the real validator, and
  `TestNumberPatternValidatorAcceptsUpToBoundRejectsBeyond` is AC17 part 2's own property check
  (accepts every `k` in `[0, avgExtraScale]`, rejects `avgExtraScale + 1`).

**FLAG-1 is LIFTED here, naming AC17** (D-000.60): `TestMaxPatternFractionDigitsIsAvgExtraScale`
pins `maxPatternFractionDigits == avgExtraScale` by symbol (never the literal `4`), and
`TestNumberPatternValidatorAcceptsUpToBoundRejectsBeyond` / `TestNumberPatternRedProofOverAvgExtraScale`
are AC17's parts 2 and 3. The formatNumber pattern vocabulary needed by the product (`"#,##0.00"`,
`"#,##0"`, and everything F7 publishes) fits within `avgExtraScale` fraction digits with no narrowing
— the constant did not need to move.

**FLAG-A, FLAG-B, FLAG-C — recorded as DISSOLVED before development started** (already stated in the
story file's own *Flags* section; restated here per D-000.60's "visible as dissolved" requirement,
now that the work is done):
- **FLAG-A** (the `ja` fence): dissolved before this story began. All four locales — `en`, `th`,
  `zh-Hans`, `ja` — are built and asserted identically in `internal/expr/locale.go`'s table, AC4b's
  set-difference test, and AC18's four-locale goldens. `ja`'s golden stays formatting-only (its own
  comment in `internal/expr/locale_test.go`/`formatlocale_test.go` states the scope fence); no
  runtime glyph-coverage diagnostic and no locale→font inference were added anywhere.
- **FLAG-B** (DECISION-1, Arm A): executed by AC16 — `funcEntry.implemented`/`owningStory`,
  `eval.go`'s `if !entry.implemented` branch, `TestUnimplementedFunctionsAreLocatedErrors` and
  `TestUnimplementedAndUnknownFunctionErrorsAreDistinguishable` are all deleted;
  `TestUnimplementedEntriesHaveNoEvalCallBranch` and its `== 0` precondition are deleted WITH it, not
  softened. `TestImplementedEntriesMatchEvalCallSwitch`'s own `== 0` precondition (a different,
  forever-safe population) stays.
- **FLAG-C** (DECISION-2, Western digits): executed by AC11a — `digitsWestern` is an explicit named
  field on every `localeTable` row (`internal/expr/locale.go`), never an implicit default.

**Two referents this story owns nothing of, noted and not acted on:** Story 5.4's falsified payload
itemisation (measured 5.07 MB brotli vs. the stated 7.9 MB) — already on the lead's Epic 5 list; DW-9's
independent confirmation by the same measurement.

**A brief detail that had drifted, flagged rather than silently fixed:** the story cites "a twin-drift
check" for `docs/expression-reference.md`/`.html` as already existing. No such automated check was
found in the repository (searched for docs-drift/twin-drift test names and scripts). Both files were
updated by hand in the same commit (AC20) and cross-checked manually for consistency (status table,
pill markers, worked examples, canonical outputs at `.md:171-198` left byte-for-byte unchanged per
F7). Also pre-existing and unrelated to this story: `docs/expression-reference.md`/`.html` still
marked `sum`/`count`/`avg`/`upper`/`lower`/`if` as "in progress" even though Story 3.3 had already
implemented three of them — corrected in the same pass as formatDate/formatNumber's markers, since it
was the same table and leaving it inconsistent would have been dishonest.

**Three divergences from the brief, all resolved by "the code is the truth" (Finding 18, this
story's finishing pass adds the third — it belongs beside the other two, not silently omitted, per
D-000.66's rule that carrying an uncollected claim forward is the one disallowed move):**
- *Do not re-open* item 1 states `TestExprFunctionTableRedProofNinthEntry` stays green **"with zero
  edits."** AC16 (Arm A) removes `funcEntry.implemented`/`owningStory`, changing `table.go`'s line
  shape — the test's exact-string injection marker **had to change** to match, and did. The
  developer's call is right and unavoidable (AC16 is later and more specific than item 1, and ruled
  after it), and the File List already names `table_derivational_test.go` as modified — this is a
  contradicted settled ruling, recorded here explicitly rather than left implicit in the diff.
- `internal/expr/table_derivational_test.go`'s three D-000.9 presence preconditions are at
  (post-edit) different line numbers than the brief's `:85`/`:121`/`:159` citations, because the file
  was rewritten rather than patched in place; the SUBSTANCE (`:85`/`found`, `:121`-equivalent
  `len(tableNames)==0`, and the removed `:159`-equivalent `unimplemented==0`) is exactly as ruled.
- `testdata/template/golden/worked-example.json` (the canonical golden) does not itself render
  end-to-end today — it fails on an UNRELATED, pre-existing gap (`element e1: has text but no
  style.fontFamily to resolve a font from`, in the page-header text element, nothing to do with
  formatNumber/formatDate). AC19's end-to-end th/ja renders are built as new, minimal, self-contained
  documents (`formatlocale_test.go`) instead, exercising the real `Render`/`ParseTemplate` API and
  `fonts.Shipped()`. `TestParseTemplateAcceptsCanonicalGolden`'s load-time assertion (the fixture's own
  actual obligation) is unaffected and still passes.

**AC1 non-vacuity:** `resolver-method-set-closed` (lint) and `TestExprResolverMethodSetIsClosed`
(folio-go) both pass with zero edits to either guard or to `expectedResolverMethods` — confirmed by
re-running the full `lint`/`folio-go` suites after `FormatContext` was threaded, with no changes
needed to either file.

### Finisher pass (2026-08-26)

**Triage: 20 of 20 findings FIX (2 Blockers, 4 Majors, 13 Minors, Nit 20's 5 sub-items). 0 DISMISS, 0
DEFER.** Three findings (AC10a's token set / Finding 4, the dot-import evasion / Finding 15, and the
digits-field gap / Finding 5) were re-ruled mid-review by the engineering lead after the initial
brief; all three applied exactly as ruled, not guessed at. See "Finding Resolutions (Finisher)" under
QA Results for the full per-finding account.

**Files this pass modified or added**, beyond the developer's own File List below: new —
`internal/expr/digits_coherence_test.go`, `folio-go/version_test.go`,
`testdata/lint/forbidden-imports/violating_dot_import.go`; modified —
`internal/expr/{formatdate,formatdate_test,datepattern,datepattern_test,numberformat_test,table_behavioral_test,locale,locale_test,check,formatcontext,calendar}.go`,
`folio-go/{version,ac4_coverage_test,render_bind_test,formatlocale_test}.go`,
`lint/internal/rules/{forbiddenimports,forbiddenimports_test,story34_forbiddenimports_test}.go`,
`docs/expression-reference.html`, `_bmad-output/specs/spec-folio/folio-format.md`,
`_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` (two appended amendments, originals
untouched).

**Gates, before this pass (reviewer's measurement) → after:**

| scope | before | after |
|---|---|---|
| `folio-go/` | 819 pass · 1 fail · 1 skip | **822 pass · 1 fail · 1 skip** (+4 new tests, −1 tautological red-proof removed) |
| `lint/` | 97 pass · 3 fail | **98 pass · 3 fail** (+1 new test) |
| `hashmatrix/` | 3 pass | **3 pass** (unchanged) |

`TestCorpusMeetsP6ExerciseFloors` stays the one required-red FAIL, stats byte-identical
(`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`). `lint`'s three fails stay DW-19's single
root cause, local-only — no fourth. The cross-target hash matrix was **not** run, correctly, per
D-000.4 — due at Epic 3's close.

**Red-proof discipline:** every FIX that could be captured against the real, committed source was —
Blocker 1 (four impossible-date rows reproduced by hand, then re-confirmed rejected), Blocker 2 (the
exact `"WRONG-DATE"` mutation D-000.36 names), the digits-coherence ruling (reddens on the date half
only), the AC10a re-ruling (the exact `"HH"` mutation the lead measured), AC12 Layer 1 (the
`tenPowLookup` mutation), AC17 part 3 (`maxPatternFractionDigits` raised in the source, both parts 1
and 2 reddened), the dot-import ruling (the check removed, the fixture stopped reporting), Finding 12
(`th` removed from the real `localeTable`), and Finding 14 (a bogus filename added to the coverage
list). Every mutated file was restored and diffed byte-identical against a pre-mutation copy before
continuing (`/usr/bin/diff`, not a wrapped `diff`, per this project's own prior finding that a
wrapped `diff` can mis-report "identical").

## File List

New files:
- `folio-go/internal/expr/formatcontext.go` — `FormatContext` (AC1/R1)
- `folio-go/internal/expr/locale.go` — the closed locale table (AC3-AC6, AC11, AC11a)
- `folio-go/internal/expr/locale_test.go`
- `folio-go/internal/expr/tenpow.go` — the integer power-of-ten lookup table (AC12)
- `folio-go/internal/expr/datepattern.go` — date-pattern grammar (AC10a)
- `folio-go/internal/expr/datepattern_test.go`
- `folio-go/internal/expr/numberpattern.go` — number-pattern grammar (AC10b, AC17)
- `folio-go/internal/expr/calendar.go` — hand-rolled proleptic Gregorian calendar arithmetic (AC7-AC9)
- `folio-go/internal/expr/formatdate.go` — `formatDate` (AC7-AC9)
- `folio-go/internal/expr/formatdate_test.go`
- `folio-go/internal/expr/numberformat.go` — `formatNumber` (AC11-AC14)
- `folio-go/internal/expr/numberformat_test.go`
- `folio-go/internal/expr/table_behavioral_test.go` — AC16's behavioural half
- `folio-go/internal/template/locale.go` — locale tag constants + `LocaleTags` (AC4)
- `folio-go/internal/template/locale_test.go`
- `folio-go/internal/bind/formatcontext_test.go`
- `folio-go/formatlocale_test.go` — AC18/AC19
- `lint/internal/rules/story34_forbiddenimports_test.go` — AC2
- `folio-go/internal/expr/digits_coherence_test.go` — **finisher, Finding 5 / digits-coherence
  ruling**: synthetic-locale coherence assertion for `formatDate`/`formatNumber`'s digit set
- `folio-go/version_test.go` — **finisher, Finding 9**: pins `LocaleTableVersion` against
  `internal/expr`'s
- `folio-go/testdata/lint/forbidden-imports/violating_dot_import.go` — **finisher, Finding 15 /
  dot-import ruling**: retained fixture for `RuleDotImport`

Modified files:
- `folio-go/internal/expr/table.go` — removed `implemented`/`owningStory` (AC16); **finisher:**
  corrected the stale `FuncEntry.OwningStory` comment reference (Finding 8)
- `folio-go/internal/expr/eval.go` — `FormatContext` threaded; removed the unimplemented-function
  branch; wired `formatDate`/`formatNumber`
- `folio-go/internal/expr/check.go` — AC10 pattern validation at load; **finisher:** the two pattern
  type assertions now use the comma-ok form with a located internal error (Nit 20.3)
- `folio-go/internal/expr/doc.go`, `ast.go` — corrected stale comments
- `folio-go/internal/expr/table_derivational_test.go` — AC16's structural half restated; the false
  top-comment claim corrected
- `folio-go/internal/expr/eval_test.go`, `aggregate_test.go`, `testutil_test.go` — `FormatContext`
  threaded through `Eval`; removed the two obsolete unimplemented-function tests; **finisher:**
  corrected two more stale "unimplemented" references (Finding 8)
- `folio-go/internal/expr_arch_test.go` — AC7/AC8's red-proof marker updated for table.go's new shape
- `folio-go/internal/template/closedsets.go` — `closedLocales` now built from `locale.go`'s constants
- `folio-go/internal/bind/text.go` — `BindText`/`BindTextSpans`/`Resolve` all take `expr.FormatContext`
- `folio-go/internal/bind/aggregate_test.go`, `scope_test.go`, `spans_test.go`, `text_test.go` —
  call-site updates for the new parameter; one test (`TestBindTextParsesFormatNumberAndFailsOnAbsentData`)
  re-pointed now that formatNumber is implemented
- `folio-go/render.go` — constructs `FormatContext` at the render entry point (R1)
- `folio-go/render_bind_test.go` — F3's re-point (`TestRenderRejectsInvalidNumberPatternAtLoad`);
  **finisher:** tightened the pattern-name assertion to the quoted form (Finding 11)
- `folio-go/render_test.go`, `layout_probe_test.go`, `row_scope_test.go` — call-site updates
- `folio-go/folio_expr_validate_test.go` — stale comment corrected
- `folio-go/ac4_coverage_test.go` — **finisher, Finding 10:** annotated the `ja` fixture as a live
  regression subject
- `docs/expression-reference.md`, `docs/expression-reference.html` — AC20; **finisher:** removed a
  dead CSS rule (Nit 20.5)
- `_bmad-output/specs/spec-folio/folio-format.md` — AC21; **finisher, Finding 18:** corrected the
  stale "aggregate not yet computed" sentence at `:223` under D-000.6
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — AC21 (re-verified correct, unchanged, by the finisher)
- `_bmad-output/implementation-artifacts/3-4-format-dates-and-numbers-by-declared-locale.md` — this file
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — **finisher:** appended a
  `D-3.4.2 amendment` entry and a `D-000.6 amendment` entry (both append-only; originals untouched)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — status → `done`
- **Finisher, Blocker 1:** `folio-go/internal/expr/formatdate.go`, `formatdate_test.go`
- **Finisher, Blocker 2:** `folio-go/formatlocale_test.go`
- **Finisher, Finding 5 / digits ruling:** `folio-go/internal/expr/formatdate.go`
- **Finisher, Finding 4 / AC10a re-ruling:** `folio-go/internal/expr/datepattern.go`,
  `datepattern_test.go`
- **Finisher, Finding 3:** `folio-go/internal/expr/numberformat_test.go`
- **Finisher, Finding 6:** `folio-go/internal/expr/numberformat_test.go`
- **Finisher, Finding 7:** `folio-go/internal/expr/table_behavioral_test.go`
- **Finisher, Finding 9:** `folio-go/internal/expr/locale.go`, `locale_test.go`, `folio-go/version.go`
- **Finisher, Finding 12:** `folio-go/internal/expr/locale_test.go`
- **Finisher, Finding 13:** `folio-go/internal/expr/numberformat_test.go`
- **Finisher, Finding 14:** `lint/internal/rules/forbiddenimports.go`,
  `lint/internal/rules/story34_forbiddenimports_test.go`
- **Finisher, Finding 15 / dot-import ruling:** `lint/internal/rules/forbiddenimports.go`,
  `lint/internal/rules/forbiddenimports_test.go`
- **Finisher, Finding 16:** `folio-go/formatlocale_test.go`
- **Finisher, Finding 17:** `folio-go/internal/expr/formatdate_test.go`
- **Finisher, Nit 20.1:** `folio-go/internal/expr/formatcontext.go`
- **Finisher, Nit 20.2:** `folio-go/internal/expr/locale_test.go`
- **Finisher, Nit 20.4:** `folio-go/internal/expr/calendar.go`

## Change Log

- Implemented `formatDate`/`formatNumber` (Story 3.4): the closed locale table (AC3-AC6, AC11,
  AC11a), hand-rolled calendar arithmetic (AC7-AC9), an integer power-of-ten lookup table (AC12),
  separate closed pattern grammars for dates and numbers validated at load (AC10a/AC10b), the
  `avgExtraScale` honesty-keeper (AC15/AC17, discharging FLAG-1), and AC16's removal of the
  registered-but-unimplemented machinery in favour of a two-half (structural + behavioural)
  obligation assertion. All four locales (`en`, `th`, `zh-Hans`, `ja`) are built and asserted
  identically, per D-3.4.1. Docs twins and the Japanese-typography disclosure (AC20/AC21) updated in
  the same pass.
- **Finisher pass, all 20 QA findings addressed (2 blockers, 4 majors, 13 minors, 5 nits — all
  FIX, none dismissed or deferred):** `formatDate` now validates civil date/time components before
  the calendar sees them, naming the offending component (Blocker 1); AC18/AC19's goldens now
  compare against text decoded from `Render`'s own PDF bytes instead of the test's own constant
  (Blocker 2); `formatDate` reads the locale table's `digits` field, and a synthetic-locale test
  asserts date/number digit coherence per the engineering lead's ruling; AC10a's token-set guard is
  re-anchored to a test-owned literal plus a fixed-array source bound, per the lead's re-ruling;
  AC12 Layer 1's oracle now genuinely exercises the power-of-ten lookup's scaling path; AC17 part
  2's property test is anchored to `avgExtraScale` rather than the constant under test; a
  project-wide dot-import ban (`RuleDotImport`) closes the `math` selector's evasion and every
  sibling AD-1 import ban at once, per the lead's ruling; plus record-keeping (Delivery Log,
  decision log, `folio-format.md`), test-hygiene (tautological/self-referential red-proofs
  repaired, package-level mutable state removed from a red-proof), and five documentation nits. See
  "Finding Resolutions (Finisher)" under QA Results for the full, per-finding account and every
  captured red-proof.

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-26
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 2 (Findings 1-2)
- **Majors:** 4 (Findings 3-6)
- **Minors:** 13 (Findings 7-19)
- **Nits:** 5 (Finding 20, five sub-items)

### Gate re-measured independently by the reviewer (D-000.26: scope + flags + numbers)

Counted from the raw `test2json` tee logs, not from any summariser's own figure.

| scope | invocation (verbatim) | measured | declared | verdict |
|---|---|---|---|---|
| `folio-go/` | `go build ./... && go vet ./... && gofmt -l . && go test ./... -count=1` | **819 pass · 1 fail · 1 skip**, 15 pkgs; build OK, vet OK, `gofmt -l` empty | 819 · 1 · 1 | **matches** |
| `lint/` | `go test ./... -count=1` | **97 pass · 3 fail** | 97 · 3 | **matches** |
| `hashmatrix/` | `go test ./... -count=1` | **3 pass** | 3 | **matches** |

- The single `folio-go` failure is the REQUIRED red `TestCorpusMeetsP6ExerciseFloors`, stats
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — **byte-identical** to the baseline declared
  at `c2b4fe2` (P6g 7, floor 20). No drift. D-000.17 / D-2.1.14 / D-000.57 honoured; not touched.
- The three `lint` failures are DW-19's single root cause (`.font-sources/`, gitignored at
  `.gitignore:85`, local-only, CI-green): `TestManifestUpToDate`, `TestResolveAssetsIncludesWordlist`,
  `TestFontsAssetsNoticeRemovalRedProof`. **No fourth failure — no regression owned by this story.**
- The cross-target hash matrix was **not** run (FLAG-E), correctly, per D-000.4; due at Epic 3's close.
- **Working tree restored byte-identical after every red-proof.** Six files were mutated during this
  review (`eval.go`, `numberpattern.go`, `datepattern.go`, `formatdate.go`, `tenpow.go`, `locale.go`)
  plus two temporary probe files; all reverted by hand and verified by SHA-256 against pre-mutation
  copies. `git status --porcelain` line count unchanged at 47 before and after.

### What was verified as CORRECT and must not be re-litigated

- **AC16's two halves are genuinely independent — teeth confirmed by measurement.** Deleting
  `evalCall`'s `"count"` branch reds BOTH halves; making that branch *return an error for valid
  arguments* reds **only** the behavioural half while the structural half PASSES; adding a stray
  `"notInTable"` branch reds **only** the structural half's second direction. All three reproduced.
  The `:159`-equivalent `unimplemented == 0` precondition is **gone with its test, not softened**
  (zero live references module-wide); the `:85`-equivalent (`found`, `evalCallSwitchNames`) and
  `:121`-equivalent (`len(tableNames) == 0`) both survive and both still work.
- **The hand-rolled calendar is exhaustively correct.** Cross-checked against Go's `time` package in a
  throwaway module outside the repo: **all 3,652,059 civil dates from 0001-01-01 to 9999-12-31 match,
  with zero mismatches**, and every one round-trips through `civilFromDays`. Leap handling correct
  (2000 and 2400 leap; 1900 and 2100 not); `floorDivMod` never yields a negative remainder.
- **The Buddhist era is applied to the year field AFTER the offset shift, in both directions.**
  Measured: `2026-12-31T23:30:00Z` @ `+07:00` → `1 มกราคม 2570` (th) / `1 January 2027` (en); the
  reverse crossing `2027-01-01T02:00:00Z` @ `-07:00` → `31 December 2026` / `31 ธันวาคม 2569`. BE is
  applied once, never twice, never before the shift. Published outputs reproduce exactly:
  `15 สิงหาคม 2569`, `15 August 2026`, `2026年8月26日`, `1,234.56`, `48,251`, `-1,234.56`;
  half-to-even ties correct in both directions (`0.125`→`0.12`, `0.135`→`0.14`).
- **AC17 was satisfied the RIGHT way round — the grammar was NOT narrowed to fit the constant.**
  `maxPatternFractionDigits = avgExtraScale` (by symbol, never the literal `4`), and the validator
  accepts *exactly* up to that bound. The literal `4` appears in no assertion this story adds.
- **AC4b's real set-difference guard has teeth.** Removing `th`'s table entry reds
  `TestLocaleTableCoversExactlyClosedSet` **in the first direction only**, naming `th`.
- **AC4's single-spelling shape holds.** Each tag string is spelled exactly once as a Go declaration
  (`internal/template/locale.go:22-26`); `closedLocales` and `expr`'s table are both keyed by those
  constants; `LocaleTags` is a declared literal slice (never a ranged map) with its exact order pinned
  as a literal sequence. Remaining tag literals are JSON fixture *documents* (consumers), not a second
  set declaration. `ac4_coverage_test.go:149`'s `ja` fixture was **not** deleted and is exercised.
- **AC1 threading is correct.** `FormatContext` is an explicit value parameter; the `Resolver`
  interface, `expectedResolverMethods`, `resolvermethodset.go` and `stagerank.go` are all **byte-identical
  to `c2b4fe2`**. Exactly one production call site (`render.go:559/578`), sourced from `doc.Locale` /
  `doc.UTCOffset`. No zero-value `FormatContext` reaches production.
- **The owner's decisions were implemented as ruled, and are NOT filed as findings.** No runtime
  glyph-coverage diagnostic exists; no locale→font inference exists (`fontChain` reads only
  `el.Style.FontFamily` and `doc.Fonts`); `ja` is built like the other three and shares `zh-Hans`'s
  data by construction. Both AC21 documentation sites are correct, well-placed against AD-12, and
  neither overstates coverage nor understates the shape limitation.
- **Docs twins are consistent.** Every status marker agrees; the canonical outputs at `.md:171-198`
  are byte-identical to `c2b4fe2` (verified with `cmp` on the extracted block); the worked
  `formatNumber` example landed in both. The developer's report that **no automated twin-drift check
  exists** is confirmed — the story brief (AC20) was wrong to claim one does.
- **No banned import or float anywhere under `internal/`.** Zero `time`/`net`/`math/rand`; `os` only in
  `_test.go` (D-1.3.1); no `float32`/`float64`/`big.Float`/`big.Rat`; no `math` transcendental. Story
  3.4's non-test files contain **zero** map ranges; the map ranges in its test files are exempt by
  `maprange.go`'s own `Tests: false`.

---

### Finding 1: `formatDate` silently accepts impossible dates and renders a *different*, plausible date

- **Severity**: **Blocker**
- **Category**: Correctness / AC Conformance
- **Location**: `folio-go/internal/expr/formatdate.go:17-20` (`rfc3339Pattern`), `:110-140`
  (`instantMsFromValue`, `KindString` arm)
- **Observation**: `rfc3339Pattern` validates only the *digit shape* of a timestamp. The captured
  fields are then fed straight to `daysFromCivil`, which happily normalises out-of-range values.
  Measured through the real `Parse`/`Check`/`Eval` path:

  | input | rendered `dd/MM/yyyy` |
  |---|---|
  | `2026-02-31T00:00:00Z` | `03/03/2026` |
  | `2026-13-01T00:00:00Z` | `01/01/2027` |
  | `2026-00-10T00:00:00Z` | `10/12/2025` |
  | `2026-01-00T00:00:00Z` | `31/12/2025` |
  | `2026-01-99T00:00:00Z` | `09/04/2026` |
  | `2026-01-10T25:00:00Z` | `11/01/2026` |
  | `1900-02-29T12:00:00Z` | `01/03/1900` |

  RFC 3339 §5.6 constrains `mday` by month and year, and `hour` to `00-23`, so none of these is a
  valid RFC 3339 timestamp. AC7 enumerates "a non-RFC-3339 string" as a case that must be **a located
  error naming the element and the offending value**.
  `TestFormatDateAcceptsRFC3339AndRejectsEverythingElse` (`formatdate_test.go:35`) tests only a
  wrong-*shape* string (`"15 August 2026"`), a non-integral number, a bool and a null — the
  shape-valid-but-impossible class is untested, which is why this shipped.
- **Impact**: The silent-wrong-answer class this project's whole discipline targets, on the money
  path. A statement whose upstream data carries `2026-02-31` prints `3 March 2026` with no error, no
  diagnostic and no signal of any kind — the reader cannot tell. Under `th` it also silently shifts the
  Buddhist-era year at a boundary (`2026-13-01` → `01/01/2027` → **2570**). This is AD-14's
  never-coerce rule breached in the one direction nobody can see, and the story's own plain-terms
  opener promises the opposite: *"a date has to arrive as a properly written timestamp."*
- **Suggested Resolution**: After the regex match and before computing the instant, reject any capture
  that is not a real civil date/time: `month` in `1..12`, `day` in `1..daysInMonth(year, month)`,
  `hour ≤ 23`, `minute ≤ 59`, `second ≤ 60` (leap second). A cheap and self-checking form is to
  round-trip — `civilFromDays(daysFromCivil(y,m,d))` must return the same `(y,m,d)` — which reuses the
  already-exhaustively-verified calendar rather than adding a second month-length table. Add the
  impossible-date class to `TestFormatDateAcceptsRFC3339AndRejectsEverythingElse`.
- **Related AC**: AC7 (and AD-14)

### Finding 2: AC18's four-locale goldens assert the test's own constants; AC19's `ja` render cannot discriminate

- **Severity**: **Blocker**
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/formatlocale_test.go:86-111` (`TestFormatLocaleGoldensAllFourLocales`),
  `:135-160` (`TestFormatLocaleEndToEndRenderJapanese`)
- **Observation**: The golden loop discards the render result and then fills its comparison map from
  the *expected* constant:

  ```go
  _, err = folio.Render(tpl, data, folio.Params(`{}`), fonts.Shipped())
  ...
  got[c.name] = c.wantDate + " 1,234.56"     // :99 — from c.wantDate, not from output
  ```

  Every subsequent assertion (`:105-110`) therefore compares table literals to table literals. The
  pinned strings are never compared to anything the code produced.
  **Measured red-proof:** with `evalFormatDate` mutated to return the constant `"WRONG-DATE"` for
  every locale, `TestFormatLocaleGoldensAllFourLocales` **PASSED** — and still logged
  `goldens: map[en:15 August 2026 1,234.56 ja:2026年8月15日 1,234.56 th:… zh-Hans:…]`, i.e. the correct-
  looking output of a completely broken formatter. `TestFormatLocaleEndToEndRenderJapanese` **also
  PASSED** under the same mutation: its only assertion is `strings.Contains(res.Bytes, "NotoSansSC")`,
  and since `"Noto Sans SC"` is the ja document's *only* declared face, every ASCII rune in the element
  already pulls that face in regardless of what `formatDate` returns. (`…RenderThai` did red, but only
  as a script-class side effect — no Thai glyphs were drawn — not because any date string was checked.)
- **Impact**: AC18 is the fixture AD-12's `[ASSUMPTION]` requires before a locale may be in the set
  ("only with a golden fixture per locale") and AC19 is the story's only end-to-end evidence. Both are
  vacuous for the property they name. The correct strings *are* pinned at the expression layer
  (`internal/expr/formatdate_test.go`, `numberformat_test.go`), so the defect would not ship
  undetected today — but AC18/AC19's own instruments contribute no coverage, and the AC19 `ja`
  assertion's comment ("the render did not actually exercise the ja text") claims a discrimination it
  does not have.
- **Suggested Resolution**: Capture the `Result` and assert the formatted strings against
  `c.wantDate` — extracting text from the PDF if needed, or at minimum asserting through the bind
  layer on the same document. For the `ja` render, put `"Noto Sans"` first in the chain so the SC face
  can only be pulled in by a CJK rune, mirroring how the `th` test already discriminates.
- **Related AC**: AC18, AC19

### Finding 3: AC12 Layer 1's exactness oracle never exercises the power-of-ten scaling it exists to protect

- **Severity**: **Major**
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/internal/expr/numberformat_test.go:134-149`
  (`TestFormatNumberExactnessOracle`), `:106-128` (the D-000.50 population check);
  `numberformat.go:25-27` (`scaleToFractionDigits`'s equal-exponent early return)
- **Observation**: Both oracle cases have `d.Exponent == targetExponent` — `{coeff: 9007199254740993,
  exp: 0}` with `"#,##0"` (fracDigits 0) and `{coeff: 123456789012345, exp: -4}` with `"#,##0.0000"`
  (fracDigits 4). `scaleToFractionDigits` returns at its first branch in both, so **`tenPowLookup` is
  never called and no power of ten is ever used**.
  **Measured red-proof:** mutating `tenPowLookup` to return `10^n + 1` for every `n` left
  `TestFormatNumberExactnessOracle` **PASSING**; only Layer 3
  (`TestTenPowTableMatchesIndependentBigIntChain`) reddened.
  Separately, the population check measures divergence for the `(coefficient, shift)` pairs
  `{9007199254740993,1}, {123456789012345,4}, {999999999999999,3}` — **shifts the oracle never
  performs**, and a third member the oracle does not use at all. D-000.50 asks that the population be
  checked *before the assertion is written*; the population checked is not the population asserted over.
- **Impact**: AC12 declares Layer 1 "behavioural, PRIMARY, has teeth … the layer that cannot be
  defeated by a synonym", and Layers 2 and 3 explicitly "narrow". As shipped the load is carried
  entirely by Layer 3, whose own AC text calls it narrow. The oracle does retain real teeth against a
  `float64` *representation of the value* (case 1 is 2^53+1), which is worth keeping — but it proves
  nothing about the lookup-table scaling path AC12 is about.
- **Suggested Resolution**: Add at least one corpus member whose `Exponent` differs from the pattern's
  `-fracDigits` so `tenPowLookup` is actually on the path (e.g. `{coeff: 9007199254740993, exp: -2}`
  under `"#,##0.0000"`, a pad; and a rounding case under a coarser pattern). Then re-run the
  `tenPowLookup` mutation and confirm the oracle reddens. Align the population check's pairs with the
  cases the oracle actually runs.
- **Related AC**: AC12

### Finding 4: AC10a's date-token AST set-equality is defeated by any token outside the declared letters

- **Severity**: **Major**
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/internal/expr/datepattern_test.go:104-118` (the `lettersInUse` closed-half
  loop), `:129-149` (`TestDateFieldTokenSetRedProofAddedToken`)
- **Observation**: The "closed" direction probes only runs of letters that already appear in a declared
  token — `lettersInUse` is built *from `declared`*, so it is `{y, M, d}` — at lengths 1 to 5. Any token
  built from another ASCII letter is never probed. The named red-proof hard-codes the single string
  `"EEEE"`.
  **Measured red-proof:** adding a hand-written special case to `parseDatePattern`'s own logic that
  accepts `"HH"`, with **no** entry in the `dateFieldTokens` literal, left
  `TestDateFieldTokenSetMatchesParser` **PASSING**, `TestDateFieldTokenSetRedProofAddedToken`
  **PASSING**, and the **entire `internal/expr` package green**. That is exactly the divergence AC10a
  names: *"add a token to the parser without adding it to the asserted set → red"*.
  (Note the instrument is also near-tautological in the other direction: `parseDatePattern` decides
  acceptance by a single `dateFieldTokens[run]` lookup, so the AST scan is re-reading the very literal
  the parser consults.)
- **Impact**: This is the fourth instrument this epic defeated by a legal alternate spelling — the
  hazard AC10a cites in its own text (`big.Float` past a name match, a literal past "non-literal", an
  embedded interface past AST set-equality). A future story can widen the accepted date vocabulary
  without the closed-set guard noticing, which is precisely what D-3.4.5's "closed, asserted" wording
  exists to prevent.
- **Suggested Resolution**: Probe the whole ASCII-letter alphabet, not just the letters in use: for
  every `r` in `a-z` and `A-Z` and every run length 1..5, `parseDatePattern` must accept the candidate
  **iff** the AST-extracted literal declares it. That closes the set against any letter and keeps the
  red-proof honest without a hard-coded `"EEEE"`.
- **Related AC**: AC10a (D-3.4.5)

### Finding 5: `formatDate` ignores the locale table's `digits` field — DECISION-2 is half-implemented

- **Severity**: **Major**
- **Category**: AC Conformance / Correctness
- **Location**: `folio-go/internal/expr/formatdate.go:157-176` (`renderDatePattern` emits via
  `fmt.Fprintf("%04d"/"%02d"/"%d")`), against `internal/expr/locale.go:35-42`, `:76-77`
- **Observation**: `renderDatePattern` never reads `entry.digits`; it hardcodes ASCII through
  `Fprintf`. Only `formatNumber` routes through `translateDigits(…, entry.digits)`
  (`numberformat.go:135`).
  **Measured red-proof:** changing exactly one table field — `th`'s `digits` to Thai numerals — gave
  `formatNumber` → `"๑,๒๓๔.๕๖"` but `formatDate` → `"15 สิงหาคม 2569"` (still Western). One document,
  two numeral systems.
- **Impact**: AC11a's operative half is that the digit set is *"an EXPLICIT NAMED FIELD in the locale
  table, never an implicit default, so a later change to Thai numerals is a visible, versioned table
  edit rather than a silent code change"* — in *"`formatDate` and `formatNumber` alike"*. As shipped,
  that visible table edit silently does nothing to dates. The resulting output is exactly the
  "incoherent middle" D-3.4.3 option 3 named and the lead declined to recommend: *"a Buddhist-era year
  in Thai numerals sitting above a column of Western-digit amounts is two numeral systems in one
  document"* — here inverted, and arrived at by accident rather than by choice. Today's output is
  correct only because every row's `digits` happens to be `digitsWestern`.
- **Suggested Resolution**: Route `renderDatePattern`'s output through `translateDigits(…,
  entry.digits)` as `renderNumberPattern` already does (build into a `strings.Builder`, translate
  once at return). Add an assertion that a non-Western `digits` value moves **both** functions'
  output for that tag — which is also the missing red-proof for AC11a's "explicit field" clause.
  Note `translateDigits`' `digitSet == digitsWestern` fast path means the non-Western branch is
  currently dead code in both functions and untested; the same assertion covers it.
- **Related AC**: AC11a (DECISION-2 / D-3.4.3)

### Finding 6: AC17 part 2 does not redden under the mutation AC17 part 3 mandates, and the shipped test's comment says it does

- **Severity**: **Major**
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/internal/expr/numberformat_test.go:247-266`
  (`TestNumberPatternValidatorAcceptsUpToBoundRejectsBeyond`), `:274-296`
  (`TestNumberPatternRedProofOverAvgExtraScale`)
- **Observation**: AC17 part 3 requires: *"set `maxPatternFractionDigits` to `avgExtraScale + 1` and
  show **both** part 1 and part 2 redden. Record the output. Revert."*
  **I performed that mutation.** Measured:

  | test | under `maxPatternFractionDigits = avgExtraScale + 1` |
  |---|---|
  | part 1 `TestMaxPatternFractionDigitsIsAvgExtraScale` | **FAIL** — "maxPatternFractionDigits = 5, avgExtraScale = 4" |
  | part 2 `TestNumberPatternValidatorAcceptsUpToBoundRejectsBeyond` | **PASS** |
  | `TestNumberPatternRedProofOverAvgExtraScale` | **FAIL** |

  Part 2 cannot redden by construction: its accept range (`k := 0; k <= maxPatternFractionDigits`) and
  its reject probe (`maxPatternFractionDigits+1`) are **both** expressed in terms of the constant under
  test, so the whole assertion slides with it. It verifies "the validator honours its own declared
  bound" — never "the bound is ≤ `avgExtraScale`".
  Separately, the shipped `TestNumberPatternRedProofOverAvgExtraScale` is **not** the mandated
  capture: it never mutates the constant. Its stated "presence precondition"
  `if inflated <= avgExtraScale` (where `inflated := avgExtraScale + 1`) is arithmetically impossible
  and can never fire. Its comment claims it confirms *"both part 1 … and part 2 … would redden"* —
  measurably false for part 2.
- **Impact**: This is D-3.1a.1 Flag F3's **only** forcing function and the story records FLAG-1 as
  LIFTED on the strength of parts 1+2+3 together. The forcing function does survive — part 1 plus the
  anchored `avgExtraScale + 1` probe together tie the validator's behaviour to `avgExtraScale`, and I
  confirmed both redden — so FLAG-1's discharge stands. But it rests on two links the story
  mis-describes, and the test comment is a claim a reviewer falsifies by running it (D-000.24).
- **Suggested Resolution**: Either (a) re-express part 2's reject probe against `avgExtraScale` rather
  than `maxPatternFractionDigits`, so it reddens when the two decouple — which is what AC17 part 3
  assumes; or (b) correct the comment on `TestNumberPatternRedProofOverAvgExtraScale` to state what it
  actually proves (the real validator rejects `avgExtraScale + 1` digits, anchored by symbol) and
  record in the Delivery Log that part 3's mutation reddens part 1 only. Drop the impossible
  `inflated <= avgExtraScale` precondition. Option (a) is preferable.
- **Related AC**: AC17, AC15 (D-3.1a.1)

### Finding 7: AC16 half 2's presence precondition is a count over its own loop and cannot fire

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/expr/table_behavioral_test.go:88-90`
- **Observation**: `exercised` is incremented exactly once per iteration of `for _, entry := range
  functionTable`, with no `continue` path — the only early exit is `t.Fatalf`, which aborts. So
  `exercised != len(functionTable)` is an identity and can never be true. It also fails at the one job
  D-000.9 gives it: were `functionTable` empty, the loop body would never run and `0 != 0` would pass
  vacuously. (`functionTable` is `[8]funcEntry`, a fixed-size array, so `len()` is a compile-time
  constant and the population cannot in fact empty — which makes the guard doubly inert.)
- **Impact**: Reported as AC16's coverage witness and as satisfying D-000.9/D-000.14. The real
  protection against an uncovered entry is the `t.Fatalf` on a missing `callFixtures` key, which does
  work — so coverage is genuinely present and this is a labelling defect, not a hole. But it is one
  more thing asserting a proposition that can never be false, which is the exact shape AC16 removed the
  `implemented` machinery to avoid.
- **Suggested Resolution**: Drop the tautological comparison and keep the fixture-presence `Fatalf`,
  or replace it with a precondition that can fire — e.g. assert `len(functionTable) == len(callFixtures)`
  (which catches a stray fixture too, the direction the file's own comment admits it does not cover).
- **Related AC**: AC16

### Finding 8: AC16's comment sweep missed two live stale sites — D-000.67 part 2, in the files this story edited

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/internal/expr/table.go:22`; `folio-go/internal/expr/eval_test.go:259`
- **Observation**: Two comments still describe the deleted machinery as live:
  - `table.go:22` — `argKind`'s doc says *"it is owed at evaluation by whichever story implements the
    function — see **FuncEntry.OwningStory**"*, a field this story deleted, **46 lines above the
    comment at `:68` that explains the deletion**.
  - `eval_test.go:259` — the section header `--- AC15/AC18/AC30: the two functions this story leaves
    unimplemented ---` is orphaned; the two tests it introduced were deleted, and it now heads
    `TestSumOverEmptyOperandIsNeverASilentZero`.
- **Impact**: D-000.67 part 2 — minted by this very story — rules that *"a targeted fix to a named
  instance does not sweep, even within the same file"*, and that the unit of a class fix is the file,
  swept, with the count of sites examined reported. AC16 named and corrected
  `table_derivational_test.go`'s top comment; the same class survived in `table.go` and
  `eval_test.go`. Both are comments a reader can falsify from the file below them (D-000.24).
- **Suggested Resolution**: Sweep `table.go`, `eval.go`, `eval_test.go` and `doc.go` for references to
  `implemented`/`owningStory` that are not explicitly historical, correct them, and report the count of
  sites examined per D-000.67 part 2. (`doc.go:13` and `table.go:68`'s references are correctly framed
  as history and should stay.)
- **Related AC**: AC16 (D-000.67, D-000.24)

### Finding 9: AC6's table version is not surfaced anywhere the library version is surfaced

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: `folio-go/internal/expr/locale.go:33` (`const localeTableVersion = 1`) vs
  `folio-go/version.go:7` (`const Version = "0.0.0-dev"`)
- **Observation**: `localeTableVersion` is unexported, lives in `internal/expr`, and is referenced by
  exactly one thing module-wide: its own test at `locale_test.go:108`. AC6 requires it to be
  *"surfaced wherever the library version is surfaced"*; being unexported, it cannot reach
  `folio.Version` at all.
- **Impact**: AD-22 makes the locale table version part of the library version, and D-3.4.3 rests its
  "changing this later is a breaking change for every downstream golden hash" reasoning on that link
  existing. As shipped there is no link — a table edit is invisible from outside the module. AC6 does
  hedge ("a label, not a guard"), but the surfacing clause is unambiguous and unmet.
- **Suggested Resolution**: Export it (`expr.LocaleTableVersion`) and reference it from `version.go` —
  e.g. compose a version string that includes it, or add a `LocaleTableVersion` constant in package
  `folio` defined as `expr.LocaleTableVersion` so the two cannot drift. Escalate if the intended
  surface is out of this story's scope, rather than leaving the AC silently unmet.
- **Related AC**: AC6 (AD-22)

### Finding 10: `ac4_coverage_test.go:149` was never annotated as a live regression subject

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: `folio-go/ac4_coverage_test.go:130-171` (file unmodified vs `c2b4fe2`)
- **Observation**: The `ja` fixture correctly survives and is genuinely exercised (`ParseTemplate` +
  `Render` + `fonts.Shipped()`), so the substantive half is satisfied. But AC4 and D-3.4.2 both require
  it be marked so a later story does not delete it as redundant: *"It must be exercised, and it must
  not be deleted as redundant once the locale table has tests of its own. **Say so where a future
  reader will find it.**"* The file carries no such note — it is untouched by this story. Task 5
  claims this complete.
- **Impact**: Small now, and precisely the kind of protection that matters only later — the next story
  that reads "the locale table has its own tests" and finds this fixture duplicative is the reader the
  annotation was for.
- **Suggested Resolution**: Add one comment above `TestJapaneseTextThroughPanCJKFaceDoesNotFireDiagnostic`
  naming it a live `ja`-loads regression subject under D-3.4.1/D-3.4.2, not to be deleted as redundant.
- **Related AC**: AC4 (D-3.4.2)

### Finding 11: the re-pointed `render_bind_test.go` "names the offending pattern" assertion cannot discriminate

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/render_bind_test.go:232-234`
- **Observation**: The re-point itself is correct — F3's instruction to re-point rather than delete was
  followed, the stage moved to load, and the element-id assertion is sound. But the second assertion is
  `strings.Contains(err.Error(), "x")`, and every error this package emits is prefixed `"expr: …"`,
  which contains an `x`. The check passes whether or not the message names the pattern.
- **Impact**: Half of the re-pointed test's stated subject ("caught with a located message" naming the
  offending pattern) is unguarded. Low blast radius; the message is in fact correct today.
- **Suggested Resolution**: Assert the quoted form the error actually produces —
  `strings.Contains(err.Error(), "\"x\"")` — or assert against the full
  `number pattern "x"` substring.
- **Related AC**: AC10b (F3)

### Finding 12: AC4b's two named red-proofs are self-contained tautologies that never touch the real guard

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/expr/locale_test.go:36-79`
  (`TestLocaleTableRedProofMissingEntry`, `TestLocaleTableRedProofExtraEntry`)
- **Observation**: Each copies `localeTable` into a local `mutated` map, mutates the copy, then
  **re-implements the comparison inside itself** and observes its own mutation. Neither invokes
  `TestLocaleTableCoversExactlyClosedSet`'s assertion path. "Delete a key from a map, then observe the
  key is absent" is a property of Go maps.
  **Measured:** removing `th`'s entry from the *real* `localeTable` left both "red-proof" tests
  **PASSING** while the real guard correctly reddened. AC4b's requirement is that the *real* check red
  in one direction only. (`TestUnlistedLocaleRedProof` at `:127-146` has the same shape, though its
  second half is a genuine assertion and the Delivery Log records AC5b as *also* captured by hand
  against real source — which is a legitimate capture.)
- **Impact**: The real guard does have teeth — I verified it — so this is a mislabelled artifact rather
  than a coverage hole. But it is recorded as AC4b's red-proof and it is not one.
- **Suggested Resolution**: Capture AC4b's two reds the way AC16 and AC5b were captured — by hand
  against the real `localeTable`, reverted and diffed — and record the verbatim output in the Delivery
  Log. Then either delete these two tests or restate them honestly as illustrations.
- **Related AC**: AC4b

### Finding 13: `localeTable` is package-level mutable state that a shipped test writes to

- **Severity**: Minor
- **Category**: Convention (AD-1)
- **Location**: `folio-go/internal/expr/numberformat_test.go:52-68`
  (`TestFormatNumberGroupingRedProof`), `internal/expr/locale.go:86`
- **Observation**: The AC11 red-proof mutates the package-level map in place and restores it with
  `defer`: `localeTable[template.LocaleEN] = mutated`. AC1 states *"`internal/expr` holds no
  package-level mutable state for this or anything else"*, and AD-1 bans package-level mutable state
  under `internal/` outright.
- **Impact**: The map is genuinely mutable and a shipped test demonstrates it. No production path
  writes to it, so determinism of rendered output is not affected today; the practical risks are
  order-dependence and a data race if any test in this package ever calls `t.Parallel()`.
- **Suggested Resolution**: Make the red-proof pure — have it call `renderNumberPattern` / a
  formatting helper with a locally-constructed `localeEntry` rather than swapping the table entry.
  That preserves the AC11 property ("only that tag's output moves") without mutating package state.
- **Related AC**: AC1, AC11 (AD-1)

### Finding 14: AC2's coverage witness omits `internal/template/locale.go` and every new `_test.go`, and re-derives its own file list

- **Severity**: Minor
- **Category**: Tests / AC Conformance
- **Location**: `lint/internal/rules/story34_forbiddenimports_test.go:22-30` (`story34NewFiles`),
  `:41` (`exprDir := filepath.Join(root, "folio-go", "internal", "expr")`), `:34`
- **Observation**: Three gaps against AC2's *"`FilesSeen` including **every** new file this story adds
  — asserted by name, not by a count that could pass while the new file was skipped"*:
  1. The by-name list is complete for `internal/expr` (all 8), but the walk root is hardcoded to
     `internal/expr`, so **`folio-go/internal/template/locale.go`** — a new file this story adds under
     `internal/` — is unreachable by the assertion and unnamed.
  2. No new `_test.go` sibling is named, although `ScanForbiddenImports` does judge test files (the
     math-selector rule has no test exemption at all).
  3. The witness runs `walkGoFiles` **itself** rather than reading the scanner's own `FilesSeen`
     (which is an `int` only, so a by-name witness is impossible from it). That is a second,
     independently-derived walk — the pattern this codebase's own `maprange.go:19-25` rejects: *"a
     vacuity guard built by re-deriving 'there were files under this root' a different way cannot see
     a scanner that silently does nothing."* If `ScanForbiddenImports` grew an early skip, this test
     would still pass. The test's comment claims the production scan "must have actually walked every
     file"; it does not establish that.
  Correctly done: no second fence was built (D-000.38/D-000.42 respected) — the file adds no rule and
  no scanning logic, and both red-proofs call the real `ScanForbiddenImports` over scratch copies.
- **Impact**: AC2 is a coverage proof; the coverage is incomplete for one new production file and for
  the test files the rule does police, and the witness cannot detect the failure mode it was written
  against.
- **Suggested Resolution**: Add `internal/template/locale.go` and the new `_test.go` files to the
  by-name list and walk both directories. Better, give `ForbiddenImportsStats` a `FilesSeenNames
  []string` so the witness reads the scanner's own record instead of re-deriving it.
- **Related AC**: AC2

### Finding 15: the `math` selector ban is evadable by a dot-import — AC12 Layer 2 cites it as coverage

- **Severity**: Minor
- **Category**: Security/Guards (instrument robustness)
- **Location**: `lint/internal/rules/forbiddenimports.go:187-197`
- **Observation**: `mathAllowedCalls` is an allow-list keyed on the selector *name*, and package
  identity is resolved from an AST-built alias map rather than `go/types`. The inspector requires an
  `*ast.SelectorExpr` whose `X` is a bare `*ast.Ident`. Therefore:
  - `. "math"` → `Pow(10, 2)` is a call whose `Fun` is a bare `*ast.Ident`, never a `SelectorExpr`.
    **Undetected.** No dot-import ban exists anywhere in `lint/`.
  - `(math).Pow(x, y)` → `sel.X` is an `*ast.ParenExpr`; the assertion fails and the walk returns early.
    **Undetected.**
  Correctly resisted: an aliased import (`m "math"`) **is** caught, and `big.Float`/`big.Rat` are
  caught by the separate `go/types`-based `bigfloattype.go`.
- **Impact**: Not introduced by this story, and no such spelling exists in the tree — I verified zero
  `math` transcendentals under `internal/` (the only `math.*` uses are `MinInt`/`MaxInt` constants in
  `_test.go`, all allow-listed). But AC12 Layer 2 cites this rule **by symbol as this story's
  coverage** for the `math.Pow` ban, and AC12's own text warns that "three instruments have been
  defeated by a legal alternate spelling in this epic alone". This is a fourth. Worth recording as a
  known limit rather than leaving the citation reading as airtight. Note Finding 3 means Layer 1 does
  not currently back-stop it either.
- **Suggested Resolution**: Do not widen scope here. Record the limit against AC12 Layer 2 (and
  consider a DW entry): either resolve the selector's package via `go/types` as `bigfloattype.go`
  already does, or add a flat ban on dot-imports under `folio-go/internal/`, which is cheap and has no
  legitimate counter-example in this codebase.
- **Related AC**: AC12 (Layer 2)

### Finding 16: AC18's ruled `ja` literal is not the one pinned, and the required scope-fence wording appears nowhere

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: `folio-go/formatlocale_test.go:80-84` (the `ja` golden row, no comment),
  `internal/expr/formatdate_test.go:231-238`
- **Observation**: Two parts.
  1. AC18 and D-3.4.1 both name the ruled `ja` value **`2026年8月26日`**. `formatlocale_test.go` pins
     `2026年8月15日` (a different instant). The ruled literal is asserted only at
     `formatdate_test.go:235`, in `TestFormatDateJapaneseGoldenFormattingOnly` — so it is covered, but
     not by the AC18 golden that names it.
  2. AC18 requires the `ja` golden's **own comment** to state *in those words* that it covers date and
     number **formatting only, never typography**, and that **no claim about Japanese typography is
     made or asserted anywhere**. The `ja` golden row has no comment at all. The nearest text is the
     AC19 *render* test's comment (`formatlocale_test.go:135-142`), which says "asserts NOTHING about
     glyph shape" — a good AC19 fence, but narrower, and in the wrong place. A case-insensitive search
     for "formatting only", "never typography" and "typography" across the module returns no relevant
     hit; the required words are nowhere in the code.
- **Impact**: AC18's fence exists to stop a later story quietly upgrading the `ja` claim into a
  typography claim that would need a D-000.22 semantic acceptance step this project cannot staff
  (D-000.41). The fence is the one durable artifact of that ruling and it is not written where the AC
  puts it.
- **Suggested Resolution**: Put the fence, in the AC's own words, on the `ja` golden at `:80-84`, and
  either change the golden's instant to the ruled `2026-08-26` or add one line noting the ruled literal
  is pinned at `formatdate_test.go:235`.
- **Related AC**: AC18, AC21 (D-3.4.1)

### Finding 17: AC8's date-line assertion covers only the forward crossing

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/expr/formatdate_test.go:142-150` (`TestFormatDateDateLineCase`)
- **Observation**: Only the positive-offset crossing is pinned (`2026-12-31T23:30:00Z` @ `+07:00` →
  `2027-01-01`). No test exercises a **negative** offset pushing the civil date backward across
  midnight, which is the branch that depends on `floorDivMod`'s non-negative-remainder correction —
  the one piece of the calendar most likely to be got wrong and the reason that helper exists.
- **Impact**: Coverage only. I verified the behaviour is correct in both directions
  (`2027-01-01T02:00:00Z` @ `-07:00` → `31 December 2026` / `31 ธันวาคม 2569`), and `floorDivMod` never
  returns a negative remainder across the values I probed — so this is an untested correct path, not a
  bug.
- **Suggested Resolution**: Add the reverse crossing as a second case in the same test, and a `th`
  variant so the era conversion is exercised on a backward year boundary too.
- **Related AC**: AC8, AC9

### Finding 18: three record-level divergences are unrecorded in the Delivery Log

- **Severity**: Minor
- **Category**: Convention (record discipline)
- **Location**: story Delivery Log, *"Two divergences from the brief"*;
  `folio-go/internal/expr_arch_test.go:292,296`; `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`
  (D-3.4.2); `_bmad-output/specs/spec-folio/folio-format.md:223`
- **Observation**:
  1. *Do not re-open* item 1 states `TestExprFunctionTableRedProofNinthEntry` "stays green **with zero
     edits**". AC16 (Arm A) removes the `implemented`/`owningStory` fields, changing `table.go`'s line
     shape, so the test's exact-string injection marker **had** to change — and did. The developer's
     call is right and unavoidable (AC16 is later and more specific), and the File List mentions it,
     but it is a contradicted settled ruling and belongs in the Delivery Log's divergence list beside
     the other two.
  2. The story header describes **D-3.4.2 as "amended on this story's pushback"** to the
     one-constant-per-tag shape, and that amended shape is what shipped (correctly). The decision-log
     entry for D-3.4.2 as appended still describes only the original accessor-only shape
     (*"`closedLocales` … gains an exported accessor"*, *"keyed from that accessor"*), with no
     amendment appended. A reader of the log alone would not know the ruling moved.
  3. `folio-format.md:223` still reads *"the aggregate itself is not yet computed — `sum`/`count`/`avg`
     are registered but unimplemented until Story 3.3"*. Story 3.3 is `done`. This is pre-existing and
     outside AC21's scope, but it is a live falsehood in the very document AC21 amended, and it is the
     same class of staleness the developer volunteered to fix in the docs twins.
- **Impact**: Record accuracy only — no code effect. Flagged because this story's own D-000.66 makes
  carrying an uncollected claim forward the one disallowed move, and D-000.6 governs the canonical
  documents.
- **Suggested Resolution**: Add (1) to the Delivery Log's divergence list; append a short amendment
  note to D-3.4.2 in the decision log (append-only, per that log's own rule); fix or explicitly defer
  `folio-format.md:223`.
- **Related AC**: AC16, AC4, AC21

### Finding 19: the plain-terms opener makes one claim the code does not honour

- **Severity**: Minor
- **Category**: Documentation
- **Location**: story, *"In plain terms"*, paragraph 2
- **Observation**: The creator deliberately did not refresh the opener, judging that nothing ruled
  changed what the owner observes. **That judgement is sound for everything except one sentence.** The
  opener says *"…and why a date has to arrive as a properly written timestamp."* Per Finding 1 the
  engine does **not** require that: an improperly written but digit-shaped timestamp is accepted and
  silently rendered as a different date. The sentence is currently a promise the code does not keep.
  On the reviewer's specific question — the opener does **not** need a new sentence about stray-letter
  patterns failing at load. It already covers that in owner-visible terms: *"It invents no formatting
  options beyond the small published vocabulary; anything outside it is rejected when the template
  loads rather than half-working later."* That is the same fact, at the right altitude, and naming
  ASCII letters would be implementation detail. Everything else in the opener remains accurate,
  including the `ja` paragraph and the two "will look wrong afterwards and are not" items.
- **Impact**: The opener is the owner-facing summary; the one inaccurate clause is about the very
  behaviour Finding 1 identifies.
- **Suggested Resolution**: Once Finding 1 is fixed the sentence becomes true and needs no edit. If
  Finding 1 is deferred instead, the clause must be softened to say what the engine actually checks.
- **Related AC**: n/a (story quality)

### Finding 20 (Nits)

- **Severity**: Nit
- **Category**: Maintainability
- **Locations and observations**:
  1. `folio-go/internal/expr/formatcontext.go:22` — cites *"localeTableEntry (below)"*; no such symbol
     exists (it is `lookupLocale`, and it is in `locale.go`, not below).
  2. `folio-go/internal/expr/locale_test.go:114` — typo, "must **reden** this test".
  3. `folio-go/internal/expr/check.go:92,98` — unchecked type assertions
     `call.Args[1].(*StringLit)`. Safe today because the arity check and `checkArgKind`'s
     `argStringLiteral` both precede and return early, but AD-14's "never a panic" makes a
     comma-ok form cheap insurance against a future reordering.
  4. `folio-go/internal/expr/calendar.go:98-101` — `civilInstant.Hour/Minute/Second` are computed on
     every call but no pattern token renders them (`renderDatePattern` handles only
     `yyyy/MMMM/MM/M/dd/d`). Write-only fields; either drop them or note they are reserved.
  5. `docs/expression-reference.html:139-140` — dead CSS rule `nav a.pending { … }`; the `pending`
     class is no longer applied by any element after this story's edit.
- **Impact**: Cosmetic / latent only.
- **Suggested Resolution**: Sweep with Finding 8's pass.
- **Related AC**: n/a

---

## Finding Resolutions (Finisher)

Every finding below was triaged FIX / DISMISS / DEFER. **19 of 20 findings (all blockers, majors and
minors) are FIX. Nit 20 is FIX in full (all five sub-items). Nothing is DISMISSed or DEFERred** —
every finding named a real, addressable defect within this story's own scope, and the three
pending-ruling items (AC10a's token set, the dot-import ban, and the digits coherence assertion) were
resolved by the engineering lead's rulings mid-finish and applied as ruled, not guessed at.

### Blocker 1 — `formatDate` accepted impossible dates
**FIX.** Added `validateCivilRanges` (`internal/expr/formatdate.go`), called from
`instantMsFromValue`'s `KindString` branch before any calendar arithmetic runs: month checked
1–12; day checked by round-tripping through the already-exhaustively-verified
`daysFromCivil`/`civilFromDays` (no second month-length table); hour 0–23; minute/second 0–59. Each
failure names the offending component and value, wrapped in the existing "not a valid RFC 3339
timestamp" located-error shape. The calendar itself (`calendar.go`) was NOT touched — it was
correct; only the missing guard in front of it was added.
Test coverage: `TestFormatDateAcceptsRFC3339AndRejectsEverythingElse` (`formatdate_test.go`) gained
seven impossible-date cases (Feb 31, month 13, month 0, day 0, day 99, hour 25, 1900 not a leap
year); a new `TestFormatDateImpossibleDateNamesTheOffendingComponentAndValue` asserts the error text
names the exact component and value for four of them.
**Red-proof, all four rows measured:** with `validateCivilRanges`'s call site disabled, a scratch
reproduction rendered exactly the four wrong dates the finding's table names —
`2026-02-31`→`03/03/2026`, `2026-13-01`→`01/01/2027`, `2026-00-10`→`10/12/2025`, hour
25→`11/01/2026` — confirming the pre-fix behaviour byte-for-byte before restoring the guard and
re-confirming all four now error. `internal/expr/formatdate.go` was restored byte-identical
(`diff` against a pre-mutation copy) after the measurement.

### Blocker 2 — AC18/AC19 goldens compared the test's own constant to itself
**FIX.** `formatlocale_test.go` no longer discards `Render`'s result. It decodes the rendered PDF's
own content stream back into the text it actually drew, through the document's own per-face
`/ToUnicode` CMaps (a small, file-local PDF-object/CMap parser — this file must stay in the
`folio_test` external package to use `fonts.Shipped()` without an import cycle through the `fonts`
package, so it cannot reuse `package folio`'s existing white-box PDF-reading test helpers, and
carries its own copy instead, the same shape `multi_page_fixture_test.go` and
`three_band_page_fixture_test.go` already use for their own, differently-scoped local CMap parsers).
An early version of this fix used one document-wide CID→text map and produced scrambled Thai output,
because CID numbering is PER FONT, not document-global; the fix resolves each `BT`/`ET` block's own
selected font resource to that resource's own `/ToUnicode` map before decoding it. All three tests
(`TestFormatLocaleGoldensAllFourLocales`, `TestFormatLocaleEndToEndRenderThai`,
`TestFormatLocaleEndToEndRenderJapanese`) now assert the decoded text equals the expected
date+number string exactly, for all four locales including `ja`.
**Red-proof, the exact D-000.36 mutation:** `evalFormatDate` mutated to return the literal
`"WRONG-DATE"` for every locale. Measured: `TestFormatLocaleGoldensAllFourLocales` now FAILS on
all four locales (previously passed while logging correct-looking strings); both
`TestFormatLocaleEndToEndRenderThai` and `TestFormatLocaleEndToEndRenderJapanese` now FAIL
(previously only Thai failed, and only as a script-class side effect, never because a date string
was checked). Reverted; `internal/expr/formatdate.go` restored byte-identical.
`ja`'s own scope fence (AC18) is now stated on the golden's own case, in the AC's words, plus a note
recording that the ruled literal `2026年8月26日` is separately pinned at
`internal/expr/formatdate_test.go:235` (`TestFormatDateJapaneseGoldenFormattingOnly`) — this golden
deliberately keeps the shared `2026-08-15` instant every other case uses so the cross-locale
non-vacuity comparisons stay meaningful (Finding 16, folded in here). The bound the reviewer could
not confirm — whether the correct glyph OUTLINES, not merely the correct code points, are drawn — is
now stated explicitly as out of this test's reach (D-000.24), rather than left implied.

### Major — `formatDate` ignored the locale table's `digits` field (Finding 5) + digits coherence (ruling)
**FIX.** `renderDatePattern` (`formatdate.go`) now routes its output through `translateDigits(...,
entry.digits)`, exactly as `renderNumberPattern` already did — one table edit now moves both
functions' output for that tag.
Per the engineering lead's ruling, the COHERENCE itself is now asserted, not merely the wiring:
`internal/expr/digits_coherence_test.go` (new file) constructs a SYNTHETIC locale entry with Thai
digits — never shipped — and asserts both `formatDate` and `formatNumber` render every digit through
that same numeral system. A synthetic locale is required (D-000.50): all four shipped locales use
Western digits today, so the property is indistinguishable from a hardcoded-ASCII implementation on
every real subject.
**Red-proof:** with the `translateDigits` call removed from `renderDatePattern`, the coherence test
reddens on the DATE half only (`got "15 August 2026", want "๑๕ August ๒๐๒๖"`) while the number half
still passes — naming the asymmetric defect precisely, exactly as the ruling anticipated. Reverted;
byte-identical.

### Major — AC10a's token set was tautological, not merely weak (Finding 4, re-ruled)
**FIX, per the engineering lead's re-ruling** (the original AST-scan repair attempt was itself
insufficient — see the ruling's own analysis of why). Implemented all three prescribed changes:
1. `datepattern_test.go` now pins `dateFieldTokensExpected` as a literal the TEST owns
   (`{"yyyy","MMMM","MM","M","dd","d"}`), asserted by set equality in both directions against the
   real `dateFieldTokens` — no AST scan.
2. `datepattern.go` declares the tokens as a fixed-size `[6]string` array
   (`dateFieldTokenList`), building the map from it, so a 7th token added without widening the array
   is a compile error.
3. The closed half now probes all 52 ASCII letters at run-lengths 1–5 (`asciiLetters()`), not merely
   letters already in use by a declared token, with expected acceptance read from the test's own
   literal, never from `dateFieldTokens` itself.
The old, tautological `TestDateFieldTokenSetRedProofAddedToken` is removed — it added nothing the
repaired `TestDateFieldTokenSetMatchesParser` doesn't now do for real.
**Red-proof, the exact measured defect:** a hand-added special case in `parseDatePattern` accepting
`"HH"` with no `dateFieldTokens` entry now reddens `TestDateFieldTokenSetMatchesParser`
(`parseDatePattern("HH"): accepted=true, want false`) — previously this exact mutation left the
entire package green. Reverted; byte-identical.

### Major — AC12 Layer 1's oracle never exercised the scaling path (Finding 3)
**FIX.** `numberFormatOracleCorpus` (`numberformat_test.go`) is now shared, by construction, between
the oracle test and its D-000.50 population check (previously two independently-typed lists that had
silently drifted apart). Two new corpus members deliberately have `Exponent != -fracDigits`: one
exercises `tenPowLookup` as a multiplier (the PAD branch), one as a divisor forcing a genuine
half-to-even decision (the ROUND branch) — both reuse coefficient `9007199254740993` (2^53+1),
already established as not exactly representable in binary64. The population check now derives its
divergence-check shift from the SAME `validateNumberPattern` call the oracle itself uses, rather than
a hand-picked, independently-guessed shift.
**Red-proof:** `tenPowLookup` mutated to return `10^n + 1`. Measured: `TestFormatNumberExactnessOracle`
now FAILS on both new cases (previously the whole oracle stayed green under this exact mutation, per
the finding). Reverted; byte-identical.

### Major — AC17 part 2 did not redden under its own mandated mutation (Finding 6)
**FIX.** `TestNumberPatternValidatorAcceptsUpToBoundRejectsBeyond`'s accept range and reject probe
are now both anchored to `avgExtraScale`, never to `maxPatternFractionDigits` (the constant under
test) — so the test reddens the moment the two decouple, which is what AC17 part 3 requires.
`TestNumberPatternRedProofOverAvgExtraScale`'s impossible presence precondition
(`if inflated <= avgExtraScale`, which could never fire) is removed, and its comment now states
honestly what it proves rather than claiming a redden it did not demonstrate.
**Red-proof, AC17 part 3's own literal instruction, performed by hand:** `maxPatternFractionDigits`
temporarily set to `avgExtraScale + 1` in the source. Measured: BOTH
`TestMaxPatternFractionDigitsIsAvgExtraScale` (part 1) AND the repaired
`TestNumberPatternValidatorAcceptsUpToBoundRejectsBeyond` (part 2) now redden — previously only part
1 did. Reverted; byte-identical.

### Minor 7 — AC16 half 2's precondition was tautological
**FIX.** `TestAllTableEntriesActuallyCompute` (`table_behavioral_test.go`) now asserts
`len(functionTable) == len(callFixtures)` up front (catches a stray fixture too, the direction the
old identity could never see) and drops the `exercised != len(functionTable)` comparison, which was
an identity over its own loop and could never fire.

### Minor 8 — Stale "unimplemented" comments survived in two files
**FIX.** Swept `table.go`, `eval.go`, `eval_test.go`, `doc.go` (4 files examined, per D-000.67 part
2). Corrected: `table.go`'s `argKind` doc (no longer cites the deleted `FuncEntry.OwningStory`);
`eval_test.go`'s orphaned section header (now correctly headed and dated) and
`TestIfShortCircuitsUnselectedBranch`'s comment (no longer calls `sum()` "an unimplemented
function", since Story 3.3 implemented it). `eval.go`'s AC14 comment and `doc.go`'s history were
already correctly framed as past tense and left unchanged, per the finding's own note.

### Minor 9 — AC6's table version was not surfaced where the library version is
**FIX.** `internal/expr/locale.go`'s `localeTableVersion` exported as `LocaleTableVersion`.
`folio-go/version.go` gained `LocaleTableVersion = expr.LocaleTableVersion` (never a second
literal), and a new `version_test.go` pins that the two cannot drift.

### Minor 10 — `ac4_coverage_test.go:149`'s live-regression status was unstated
**FIX.** Added a comment above `TestJapaneseTextThroughPanCJKFaceDoesNotFireDiagnostic` naming it a
live `ja`-loads regression subject under D-3.4.1/D-3.4.2, not to be deleted as redundant once the
locale table grows its own tests.

### Minor 11 — `render_bind_test.go`'s "names the pattern" assertion could not discriminate
**FIX.** `strings.Contains(err.Error(), "x")` (true of every error this package emits, since they
are all prefixed `"expr: …"`) replaced with `strings.Contains(err.Error(), "\"x\"")`, the quoted
form the error actually produces.

### Minor 12 — AC4b's two red-proofs never touched the real guard
**FIX.** Factored the real comparison into `localeTableCoverageMismatches` (`locale_test.go`), used
by both the main coverage test and its two red-proofs. The red-proofs now build a COPY of the real
`localeTable` (never mutating the package-level map — see Minor 13) and call the SAME function the
real guard uses, asserting each direction reddens independently.
**Red-proof:** `th`'s entry temporarily removed from the REAL `localeTable`.
`TestLocaleTableCoversExactlyClosedSet` reddened, naming `th`, in the first direction only. Reverted;
byte-identical.

### Minor 13 — `localeTable` was mutated in place by a shipped test
**FIX.** `TestFormatNumberGroupingRedProof` (`numberformat_test.go`) no longer writes
`localeTable[en] = mutated`. It reads a COPY of `en`'s entry (a Go map read of a struct value is
always a copy), mutates the copy, and calls `renderNumberPattern` directly — the real,
package-level `localeTable` is never written to.

### Minor 14 — AC2's coverage witness had three gaps
**FIX, the "better" option named in the finding.** `ForbiddenImportsStats` gained
`FilesSeenNames []string`, populated by the scanner's own walk. `story34_forbiddenimports_test.go`
now scans the REAL `folio-go/internal` root (not a hardcoded `internal/expr` subdirectory) and reads
`stats.FilesSeenNames` directly — no second, independently-derived walk. `story34NewFiles` now lists
all 16 files the story's own File List names under `internal/` (13 under `expr`, 2 under
`template`, 1 under `bind`), including every new `_test.go` sibling and `internal/template/locale.go`.
**Red-proof:** a bogus filename added to `story34NewFiles` correctly reddened the coverage test.
Reverted.

### Minor 15 — the `math` selector ban was evadable by a dot-import (re-ruled, FIX NOW, WIDER)
**FIX, per the engineering lead's ruling — wider than the finding.** Added `RuleDotImport`
(`lint/internal/rules/forbiddenimports.go`): a single predicate on an import spec's local name being
`"."`, applied unconditionally to every file `ScanForbiddenImports` covers under `internal/` — not a
patch to `mathAllowedCalls`, which would have left `os`/`time`/`net`/`math/rand` equally exposed to
the identical evasion. A retained fixture,
`folio-go/testdata/lint/forbidden-imports/violating_dot_import.go` (`. "math"`, calling bare `Pow`),
is red-provable now, which the finding itself noted the narrow version could not be.
**Red-proof:** with the dot-import check removed, the fixture scan silently stopped reporting
`violating_dot_import.go` at all (`RuleMathSelector` cannot see a dot-imported call either).
Reverted; byte-identical.

### Minor 16 — the `ja` golden's literal and scope-fence wording
**FIX, folded into Blocker 2's fix above.** The golden's own comment now states the AC18 scope fence
in the AC's words, and records where the ruled `2026年8月26日` literal is actually pinned.

### Minor 17 — AC8's date-line assertion covered only the forward crossing
**FIX (coverage only — the finding confirmed the backward path was already correct).** Added
`TestFormatDateDateLineCaseBackwardCrossing` (`formatdate_test.go`): a negative-offset crossing,
asserted for both `en` and `th` (the era conversion on the far side of a backward year boundary).
Both pass on the first run, confirming `floorDivMod`'s correctness on this previously-untested path.

### Minor 18 — three record-level divergences were unrecorded
**FIX, all three.** (1) The Delivery Log's "divergences from the brief" list now names
`table_derivational_test.go`'s contradicted `zero edits` claim from *Do not re-open* item 1 (three
divergences, not two). (2) Appended a `D-3.4.2 amendment` entry to the decision log (append-only —
the original D-3.4.2 entry is untouched) recording that the shipped shape is one named constant per
tag, not the accessor-only shape the original entry describes. (3) `folio-format.md:223`'s stale
"aggregate not yet computed… registered but unimplemented until Story 3.3" sentence corrected in
place under D-000.6, with a before/after decision-log entry — Story 3.3 and this story both shipped;
what remains true (footer-cell placement is Story 4.5's job) stays stated.

### Minor 19 — the plain-terms opener promised the opposite of what the code did
**FIX.** See "In plain terms" above, refreshed: it now states that an impossible date is refused by
name (true as of Blocker 1's fix) and that a document's dates and amounts always share one numeral
system (true as of the digits-coherence fix). Nothing else in the opener needed correction, per the
finding's own assessment.

### Nit 20 (five sub-items) — all FIX
1. `formatcontext.go`'s stale `localeTableEntry (below)` reference corrected to name `lookupLocale`.
2. `locale_test.go`'s "reden" typo corrected to "redden".
3. `check.go:92,98`'s unchecked type assertions given comma-ok form with a located internal error,
   cheap insurance against a future reordering (AD-14).
4. `calendar.go`'s `civilInstant.Hour/Minute/Second` fields now carry a comment stating they are
   RESERVED for a future time-of-day pattern token, not dead weight.
5. `docs/expression-reference.html`'s dead `nav a.pending` CSS rule removed (confirmed zero elements
   use the class anywhere in the file).

### Verified correct, not re-litigated
Every item in the QA review's "verified as CORRECT" list (AC16's independent halves, the exhaustive
calendar, AC9's era-after-offset ordering, AC17's grammar-not-narrowed shape, AC4b's real
set-difference teeth, AC4's single-spelling constants, AC1's `Resolver`/`stagerank.go`
byte-identity, the owner's `ja`/font decisions, docs-twin consistency, the zero-banned-import/float
sweep) is unchanged by this finishing pass and was re-confirmed by the full test run below.
`ARCHITECTURE-SPINE.md`'s AD-12 `ja` disclosure sentence (developer-edited under D-000.6, reviewer-
verified) was re-read and remains correct against every change made here — untouched.

### Validation after all fixes (three modules, D-000.4/D-000.64)

| scope | invocation | measured |
|---|---|---|
| `folio-go/` | `go build ./... && go vet ./... && gofmt -l . && go test ./... -count=1` | **822 pass · 1 fail · 1 skip**, 15 pkgs; build/vet OK, gofmt clean |
| `lint/` | `go test ./... -count=1` | **98 pass · 3 fail** |
| `hashmatrix/` | `go test ./... -count=1` | **3 pass** |

- `folio-go`'s one failure is the same REQUIRED red, `TestCorpusMeetsP6ExerciseFloors`, stats
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — **byte-identical** to both the baseline and
  the reviewer's own measurement. Untouched by this pass.
- `822 = 819 (reviewer's count) + 3`: this pass adds
  `TestFormatDateImpossibleDateNamesTheOffendingComponentAndValue`,
  `TestFormatDateDateLineCaseBackwardCrossing`, `TestDigitsFieldAppliesToBothDateAndNumber`,
  `TestLocaleTableVersionSurfacedAtLibraryLevel` (+4) and removes the tautological
  `TestDateFieldTokenSetRedProofAddedToken` (−1).
- `lint`'s three failures are still DW-19's single root cause (`.font-sources/`, local-only,
  CI-green) — no fourth failure. `98 = 97 (reviewer's count) + 1`
  (`TestForbiddenImportsDotImportEvadesMathSelectorAloneButNotDotImport`).
- The cross-target hash matrix was **not** run, correctly, per D-000.4 — due at Epic 3's close, not
  this story.
