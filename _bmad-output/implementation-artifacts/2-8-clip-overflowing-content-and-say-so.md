# Story 2.8: Clip overflowing content and say so

Status: **done**. OD-1/OD-2/OD-3 were ALL RULED before development started (D-2.8.1 through
D-2.8.6) — the *DECISIONS NEEDED* section below is historical context for why AC6/AC7 were originally
left unwritten; both are now written and implemented. See "Do not re-open" and the Dev Agent Record
below for what actually shipped.

**Covers:** FR44 · AD-14, AD-24 · **last story in Epic 2** (the boundary gate runs after it)

---

## FIRST — the sign-off measurement D-000.41 requires, performed at creation

**Two of the three pinned digests have an owner sign-off pending against them. This story is the
first in the epic that has a measured, concrete reason to expect a golden to move, so the
measurement comes before anything else.**

Re-computed at `f651409` on a clean tree via `rtk proxy shasum -a 256 … > /tmp/dig.txt`
(D-000.12 as corrected — never through the wrapper's pipes), not inherited from the brief:

| artifact | digest at `f651409` | sign-off state |
|---|---|---|
| `fixtures/shaped-text/expected.pdf` | `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370` | **Thai READING sign-off pending** (D-2.3.5) |
| `fixtures/expected-breaks/expected_breaks.json` | `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de` | **Thai BREAK sign-off pending** (D-2.4.3) |
| `fixtures/multi-page/expected.pdf` | `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b` | recorded by Story 2.6 |

All three match the brief exactly.

**And here is the part that is new, and that the brief could not have known.** Story 2.7's creator
could write *"no committed fixture contains `{{page}}`"* and expect the digests to hold. **This
story cannot make the equivalent claim.** Measured (finding 4 below): **the committed goldens
already overflow their declared boxes — in eleven places across four fixtures**, and two of those
four are the two carrying pending sign-offs.

The consequence, stated once, plainly, so it reaches the owner rather than a diff:

- **On the HORIZONTAL axis**, exactly one element overflows: `fixtures/wrapped-text/input.folio`'s
  `e4`. Clipping it — which is what FR44 asks for, and what AD-25 names in terms — **moves
  `fixtures/wrapped-text/expected.pdf`** (`277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5`
  at `f651409`). There is no byte-neutral way to clip a thing that is currently painted. This story
  argues for that movement explicitly, below, as the brief's *"unless the story argues otherwise"*
  permits. **`wrapped-text` carries no pending sign-off** — the two pending ones bind to
  `shaped-text/expected.pdf` and `expected-breaks/expected_breaks.json` (D-000.26 refined: a
  sign-off binds to the artifact expressing the property judged).
- **On the VERTICAL axis**, ten elements overflow across `multi-page` (1), `shaped-text` (7) and
  `multi-script-fallback` (1). **Clipping on that axis moves both pending-sign-off goldens and the
  `multi-page` golden**, and — worse — it clips the `multi-page` fixture's flowed content, which is
  FR30's own subject. That is why the vertical axis is fenced OUT of this story below, and why the
  fence is escalated as **OD-2** rather than merely asserted.

---

## The epic's acceptance criteria, verbatim, cited BY ORDINAL

`_bmad-output/planning-artifacts/epics.md`, *Story 2.8: Clip overflowing content and say so*.
**Cited by ordinal and never by line number** — every line citation in Story 2.7 was wrong, the
orchestrator's included, and three of them shipped into source comments (D-2.7.4's process note;
Story 2.7 review Finding 6).

The section carries **two** `Given` blocks. Counted by reading the section, not by trusting a
summary (`[[brief-ac-list-is-a-subset]]`):

> **Source AC1**
> **Given** absolutely-positioned content exceeding its declared bounds
> **When** the document renders
> **Then** it is clipped at the component boundary
> **And** a diagnostic is reported naming the element id
> **And** it is never reflowed and never silently dropped

> **Source AC2**
> **Given** the diagnostic
> **When** the render completes
> **Then** PDF bytes are still returned — clipping degrades output but does not fail the render

**Source AC1 has three consequents and they do not have the same status.** *"never reflowed and
never silently dropped"* is buildable today and is written below. *"clipped at the component
boundary"* is buildable on one axis and moves a golden. *"a diagnostic is reported"* — and the
whole of source AC2 — **is not expressible on the current public surface**, which D-2.6.5 recorded
nine days ago as this story's to own.

---

## Baseline, measured at creation

HEAD is **`f651409`** on `main`. Working tree **clean** — `git status --porcelain` produced no
output, checked immediately before and after every probe below, and **every probe file was
deleted** and the clean state re-checked (D-2.6.9 correction: a count taken from a tree with
uncommitted work in it is a measurement of a moment, not of the repository).

**Full ordinary suite** — `rtk proxy go test ./... -count=1 -v` from `folio-go/`, output captured to
a file:

| measure | value |
|---|---|
| all-occurrences PASS (`--- PASS`) | **587** |
| all-occurrences FAIL (`--- FAIL`) | **1** |
| top-level PASS (`^--- PASS`) | **359** |
| top-level FAIL (`^--- FAIL`) | **1** |

**Exactly one expected failure**, and it is not this story's: `TestCorpusMeetsP6ExerciseFloors`, the
intentional Story 2.1 red, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` —
reproduced character-for-character from this run, not inherited from the brief. It is a floor that
is not met, reported as unmet, and it is **not to be "fixed"** (D-000.17).

**Scope note (D-000.14, extended).** These are scoped figures: *whole module, no build tags,
`-count=1`, all-occurrences and top-level reported separately*. Any figure this story reports later
must carry the same three-part scope or it is not comparable
(`[[carried-gate-figures-are-scoped]]`).

---

## In plain terms (read this first if you just want the gist)

When someone designs a report, they draw a box and put words in it. Sometimes the words do not fit
— a customer has a longer name than the designer imagined, or a Thai surname that the engine is
forbidden to break in half. Before this story the extra text was simply painted outside the box,
over whatever happened to be next to it, and nobody was told. On a statement that goes out fifty
thousand times, the first person to notice was a customer.

That no longer happens unannounced. The engine now measures whether each piece of text actually
fits its declared box, and when it does not, the overrunning part is cut off cleanly at the box's
edge instead of spilling across the page — and whoever produced the document is told, in the same
breath the finished pages come back, which element overflowed and why. The complaint reaches the
person who can fix it, not a customer holding the printed page. The older promises hold too:
overflowing content is never rearranged to fit and never quietly thrown away, only trimmed. The
cut only ever happens sideways; cutting the top and bottom of a box was deliberately left alone,
because several of the project's own sample documents already spill past those edges by design.

Before sign-off, a second engineer hunted for checks that could not actually fail, on the theory
that such a check is worse than none. Three of the new checks meant to prove this all works had
exactly that flaw: one meant to pin a specific label down instead compared a value against itself;
one could not tell "set correctly" from "never set" because the two looked identical; and one
would crash the whole test run, rather than fail cleanly, if the defect it existed to catch
reappeared — silently hiding everything checked alongside it. Two were tightened and proven, by
deliberately reintroducing each mistake, to now catch it. The third needs a small change to a
shared building block before it can be made trustworthy; that call was left for whoever next
touches that block, and is recorded as open rather than quietly accepted.

---

## Story

**As a** template author,
**I want** content that exceeds its box cut at the boundary with a diagnostic rather than silently
lost,
**So that** I find out on the preview instead of from a customer.

---

## Do not re-open — settled rulings this story inherits

- **D-2.6.5** — an item that fits in **no page window** is a **located error naming the element**,
  for a line and an image alike, returned from `Render`/`RenderTo`. `layout.OverflowError` is the
  ruled shape. **This story does not soften it, does not reclassify it, and does not route it
  through whatever channel it eventually gains.** D-2.6.5 says in its own text that any such
  reclassification *"is the owner's call, not a re-ruling."*
- **D-2.7.3** — `{{page}}` in a content band is a located template error. Same category:
  *a declaration-level impossibility, knowable from the template alone with no data.*
- **D-2.6.1 (amended)** — the sliding window; nothing repositions; across a window boundary
  declared vertical gaps collapse.
- **AD-24** — boxes are absolute and nothing negotiates. An image is scaled to **fit** its declared
  box and centred; never stretched, **never cropped**.
- **AD-14** — one `Diagnostic` value: `Severity` (`Error` aborts, `Warning` accompanies a
  successful render), a stable string code from a closed registry, an optional element id, an
  optional data path, a message. **Over-tall rows (FR25) and clipped content (FR44) are `Warning`s
  returned alongside PDF bytes, never silent and never fatal.**
- **AD-25** — an unknown Thai run is atomic and *"overflows visibly under FR44 — clipped, with a
  located diagnostic — rather than being split at a guess."*
- **D-1.7.1** — no options struct on `Render`'s **arguments**, because it would make the `FontSet`
  omittable at compile time and turn an AD-8 violation from a compile error into a runtime one.

### The apparent contradiction, and why it is not one — stated so a developer meeting both rules does not have to guess

The project now holds **two rulings that say "error"** (D-2.6.5, D-2.7.3) and **one spec clause
that says "diagnose and continue"** (this story). A developer who meets all three without being
told the distinction will guess, and will guess differently from the next developer.

**The distinction is the SUBJECT of the sentence "it does not fit", and it is drawn by whether a
correct rendering of the element exists at all.**

| | D-2.6.5 / D-2.7.3 | this story |
|---|---|---|
| what does not fit | the item vs. **the page window** | the content vs. **its own declared box** |
| is a correct output available? | **No.** There is no page the item can be placed on. Any output would be a fabrication. | **Yes.** The pages are correct, the element is where the author put it; only its own box is too small for what the author asked to put in it. |
| decidable from the template alone? | **Yes** — leading is a function of declared chain and size (D-2.4.2 c.1); the window is page geometry minus declared bands. **No data can change it.** | **No.** It is a function of the *data*: `ศรีสุข` overflows and `Lee` does not, in the same template. |
| disposition | **located error**, render fails | **located diagnostic**, render succeeds |

**That third row is the load-bearing one, and it is the reason the two dispositions compose rather
than conflict.** D-2.6.5's own Ground 2 makes declaration-level decidability the criterion: *"Two
declaration-level impossibilities → one disposition."* This story's case fails that criterion by
construction — the very fixture FR44 exists for (`wrapped-text` e4) overflows only because of the
customer name in the report data. **A template that is fine for one customer and broken for another
cannot be a load error**, because there is no load at which it is knowable. Which leaves exactly
one place to put the news: alongside the bytes.

D-2.6.5 anticipated this and said so: *"A ruling that forces a choice between returning bytes and
diagnosing will have the diagnostic dropped, which is the half that ships silently wrong output."*

**Nothing here is re-opened.** If the engineering lead reads the table and disagrees, that is a
ruling to make; it is not a thing the developer may settle in an editor.

---

## Measured findings — read all of these before writing code

All measurements at `f651409`, clean tree, probe files deleted afterwards and the clean state
re-checked.

### 1. A text element's declared `height` is consulted by NOTHING today; its `width` is consulted by exactly one thing

Every read of an element's declared box in the module root's renderer:

- **Image elements** — the box is required (a missing `width`/`height` on an image is a located
  render error) and is the input to AD-24's fit-and-centre arithmetic. **An image therefore cannot
  overflow its box**: it is scaled to fit, by ruling and by construction. Every remaining `Height`
  read in the renderer is a *band* height or the *page* size, not an element's.
- **Text elements** — the declared `width` is passed to the line packer as the wrap width. The
  declared **`height` is read by nothing at all.** There is no code path in which a text element's
  height influences layout, pagination, or emission.

**Consequence for scope: FR44's subject in this story is TEXT, and effectively one axis.** An image
is already fitted; a table does not exist yet (Epic 4) and declares no box at all under AD-13;
`line` and `rect` elements are parsed but rendered by nothing.

### 2. Today, overflow is painted, silently, on both axes — measured, not inferred (D-000.30: captured BEFORE any fix)

Two throwaway probes through the **public** `folio.Render`, both deleted:

**Vertical.** A single text element declaring `width: 60, height: 10` at `fontSize: 12`, holding one
sentence:

```
Render SUCCEEDED, 57963 bytes, error nil
runs=10  distinct line origins=10
line origins span 781062 mp … 633966 mp  →  extent 147096 mp against a declared 10000 mp box
clip-path operators ('W n') across every content stream in the document: 0
```

**14.7× the declared height, painted, with a nil error and not one clip operator anywhere in the
produced document.**

**Horizontal.** The 34-rune atomic token `Supercalifragilisticexpialidocious` at `fontSize: 12`,
rendered twice — once in a 20 pt box, once in a 600 pt box:

```
width= 20pt: render OK, runs=1 lines=1 totalCIDs=34
width=600pt: render OK, runs=1 lines=1 totalCIDs=34
```

**Identical glyph sequence, one line, nothing dropped, nothing re-broken.** This is the red-proof
for the *never-reflowed / never-dropped* half **and** the negative control for the clip half in one
measurement: the narrow box paints all 34 glyphs today, so a test asserting "all 34 are still laid
out after the change" would pass vacuously unless it is paired with an assertion that they are no
longer all *painted*. Write both or neither (D-000.36's hazard, inherited).

### 3. The horizontal case is already documented in the packer as ruled behaviour

The line packer has two branches for "the smallest available unit does not fit": one when a break
opportunity exists after it, one when none does. Both place the unit and both are commented as
Story 2.4's AC11 — *"the first available unit does not fit. It goes on this line and overflows
visibly"* and *"one atomic unit runs to the end of the element. Same rule: it overflows."*

**So "overflows visibly" is not an accident to be repaired; it is the ruled behaviour, and FR44 is
the clause that says what "visibly" is supposed to look like.** AD-25 completes the sentence:
*"clipped, with a located diagnostic."* The packer must not change.

### 4. THE FIXTURE POPULATION — D-000.50's pre-flight question, and the answer is the opposite of Story 2.6's

D-000.50: *"Before writing a guard, ask whether any subject can express the defect."* For Story 2.5
and Story 2.6 the answer was *no subject can*. **Here the answer is: the subjects are the committed
goldens themselves.**

Measured by laying out every text element of every committed fixture through the same functions the
renderer uses (`shapeSegments` → `atomicSpansFor` → break opportunities → the line packer → the
ruled vertical model), i.e. the existing per-element layout probe, not a second implementation.
`font-text` is excluded: its chain is the Roboto test face, which the shipped font set does not
carry, so it cannot be laid out by this probe (recorded rather than silently skipped).

**HORIZONTAL — widest packed line vs. declared `width`. Exactly ONE element in the entire committed
fixture set overflows:**

| fixture | element | widest line | declared box | overflow |
|---|---|---|---|---|
| `wrapped-text` | **`e4`** | **24 585 mp** | **20 000 mp** | **+4 585 mp (+22.9 %)** |
| `wrapped-text` | `e1` | 139 073 mp | 150 000 mp | fits |
| `wrapped-text` | `e2` | 107 107 mp | 150 000 mp | fits |
| `wrapped-text` | `e3` | 143 000 mp | 150 000 mp | fits |
| every other fixture element | — | — | — | fits |

`wrapped-text` `e4` is `"ผู้รับ {{customer.name}}"` in a 20 pt box against
`{"customer":{"name":"ศรีสุข"}}` — **AD-25's worked case, placed there deliberately by Story 2.4**:
a surname that is character-for-character two common dictionary words, declared unbreakable, in a
box too narrow for it. It is, as far as the repository is concerned, *the* FR44 subject.

**VERTICAL — laid-out extent (first line top → last line bottom, i.e. the ruled model's
`FirstBaseline + LastDescent` on the last line) vs. declared `height`. TEN elements across THREE
fixtures overflow:**

| fixture | element(s) | lines | extent | declared | overflow |
|---|---|---|---|---|---|
| `multi-page` | `e1` | 29 | 947 952 mp | 700 000 mp | **+247 952 mp (+35.4 %)** |
| `shaped-text` | `e1`–`e7` (all seven) | 1 each | 25 760 mp | 24 000 mp | **+1 760 mp (+7.3 %) each** |
| `multi-script-fallback` | `e1` | 1 | 22 540 mp | 20 000 mp | **+2 540 mp (+12.7 %)** |

`three-band-page`, `page-count-1/5/20/50` and `wrapped-text` all fit vertically.

**Two things follow, and they are the shape of this whole story.**

### 5. `shaped-text`'s 7.3 % is systematic, not authorial — the ruled line box is simply taller than a hand-declared "font size plus a bit"

All seven `shaped-text` elements declare `height: 24` at `fontSize: 16` and lay out to
**25 760 mp**. That number is not content-dependent: it is `max(hhea ascent) + max(descent)` over
the **declared chain**, scaled to 16 pt — the vertical model D-2.4.2 (as amended) ruled and Story
2.5a shipped. Every single-line element in this project sized on the intuition *"height ≈ font size
× 1.5"* is under it.

**So "the author declared the box too small" is the wrong description of seven of these ten
elements.** A vertical clip would cut the descenders off every line of the Thai/CJK reading fixture
— the artifact the owner has been asked to read and sign off on — and would do so for a reason the
author had no way to anticipate from the format documentation. That is a product decision, not a
bug fix.

### 6. `multi-page` `e1` is FR30's own subject, and clipping it vertically would contradict a shipped FR

`multi-page` `e1` declares `height: 700` and lays out **29 lines / 947 952 mp**, flowing across
pages under D-2.6.1's sliding window. That is the fixture's entire purpose. If a text element's
declared `height` became a clip bound, this element would be cut at 700 pt — roughly 21 of its 29
lines — the document would lose its later pages' content, and **FR30 ("a document that grows
produces more pages") would stop being true of the fixture that proves it.**

**Nothing in the format makes a content-band text element's `height` mean anything.** Finding 1
measured that it is read by nothing. Making it a clip bound is not "implementing FR44"; it is
**minting a new format semantic**, and D-000.6 would require `folio-format.md` to be amended to say
so. That is OD-2.

### 7. AD-14 already fixes the diagnostic's VALUE and its DISPOSITION — but not its SIGNATURE, and the package does not exist

AD-14 is not silent here; it is unusually specific, and it should be read before any option is
weighed:

> one `Diagnostic` value carries `Severity` (`Error` aborts the render, `Warning` accompanies a
> successful one), a **stable string code** from a closed registry, an optional element id (AD-10),
> an optional data path, and a message. […] **Over-tall rows (FR25) and clipped content (FR44) are
> `Warning`s returned alongside PDF bytes, never silent and never fatal.**

So the **payload type is already ruled** (`diag.Diagnostic`, with those five fields) and the
**disposition is already ruled** (`Warning`, alongside bytes). **What is not ruled is the shape of
the seam that carries it out of `Render`.** No invented bespoke type; no `[]string`; no `error`
carrying a list. That is a real constraint on the options below and it narrows them usefully.

`internal/diag` **does not exist** at `f651409`. Story 3.6 mints it and its code registry (DW-6).

### 8. DW-6's anti-rot tripwire DOES NOT EXIST — measured, and it matters directly to this story

`deferred-work.md`'s DW-6 says: *"Anti-rot mechanism: a test asserts `folio-go/internal/diag` is
**absent**. The day Story 3.6 creates it, the assertion goes red and forces these two codes to be
wired before the build can pass again."*

**There is no such test.** Searched the whole tree for the path `internal/diag`; it appears in
exactly two places, and neither is an assertion:

1. `folio-go/testdata/lint/stage-rank/pagemodel/violating_equal_rank.go` — a lint **fixture** that
   *imports* the path as a deliberate rank violation. It is testdata, not module code, and it
   asserts nothing about the package's existence.
2. a prose comment in `folio-go/internal/template/errors.go` saying the package does not exist yet.

**Consequence.** If this story (or any option below) mints `internal/diag`, **nothing reddens**,
DW-6's two codes (`TABLE_FOOTER_SOURCE_UNRESOLVED` / `TABLE_FOOTER_SOURCE_FORBIDDEN`) are silently
skipped, and the forcing function the deferral was granted on turns out never to have been built.
This is `[[named-sinks-may-not-exist]]` and `[[verify-a-ruling-actually-shipped]]` in one: a
deferral's premise, believed for four stories, that no one checked. **It is flagged, not fixed,
below** — but any option that creates the package owes the tripwire *first*, in the same commit.

### 9. Call sites, at a named commit, on a tree with no uncommitted changes (D-2.6.9 correction)

Counted at **`f651409`**, `git status --porcelain` empty, by parsing each `.go` file with comments,
string literals and raw strings stripped so a mention inside prose or inside a generated-program
literal cannot be counted as a call:

| symbol | call expressions | files | of which NON-test |
|---|---|---|---|
| `Render` | **72** | 25 | **5** — `render_entry.go` (RenderTo's own call), and `testdata/swapproof/{good,bad,convert}/main.go` (the argument-order proof; `bad` is a deliberate must-not-compile subject), plus `example_test.go` which is the godoc example and is byte-bound to `README.md` |
| `RenderTo` | **7** | 1 | **0** — all in `renderto_test.go` |

**Zero call sites exist outside the `folio-go` module.** There is no designer, no wasm consumer and
no second module in the repository that calls either function.

`example_test.go` counts as one call site but **two artifacts**: a committed check asserts
`README.md`'s single fenced Go block is byte-identical to `example_test.go`, so any signature change
that touches the example must land in both or the suite goes red. `README.md` separately documents
the argument-order guarantee in prose.

### 10. The forbidden-import guard follows the NAMES `Render` and `RenderTo`, and nothing else

AD-1's import guard locates the file(s) to scan by walking the AST of every non-test file directly
under `folio-go/` and matching top-level function declarations named **exactly** `Render` or
`RenderTo`; it then scans those files for `time`/`os`/`net`/`math`/`rand`. It has a vacuity guard
for "no declaring file found", but **no** guard for "a new public entry point exists under a
different name."

**Measured consequence, which is a real cost of one option below:** a sibling entry point called,
say, `RenderWithDiagnostics`, declared in a new file, would be **outside the scan** — a public
render path with no forbidden-import protection, and nothing would say so. This is D-000.15 (key a
guard on its **purpose**, never on a proxy) with the proxy being a function *name*.

### 11. `v0.1.0` has not shipped — the brief's premise, corrected (D-000.31)

The brief says *"This is a public API shape at v0.1.0."* Measured: the repository has **one** git
tag, `pre-email-rewrite`, and the module's own version constant is **`"0.0.0-dev"`**, whose comment
says *"A real release tag is `folio-go/v0.1.0` … that is not this story's business."*

**So there are zero external consumers by construction, and "breaking change" today costs a diff,
not a major version.** This does not make the decision less the owner's — it is still the published
shape of the product's front door, and AD-22 makes any rendered-byte change breaking for downstream
regardless. But it materially changes the *price* of options (B) and (C) below, and the owner should
weigh them knowing it.

---

## DECISIONS NEEDED — escalate before development starts

### OD-1 (BLOCKING, and it is the owner's or the lead's) — how does a successful render carry a warning?

**Source AC1's diagnostic clause and the whole of source AC2 cannot be written down until this is
ruled, and they are deliberately left unwritten below.**

The surface today:

```go
func Render(t *Template, d Data, p Params, f FontSet) ([]byte, error)
func RenderTo(w io.Writer, t *Template, d Data, p Params, f FontSet) error
```

**Neither can express "succeeded, with warnings."** That is the whole difficulty and it has been
this story's to own since D-2.6.5.

Two things are already settled and must not be re-litigated inside this decision: **the payload is
AD-14's `diag.Diagnostic`** (finding 7), and **`Error` aborts / `Warning` accompanies** is AD-14's
wording. Only the seam is open.

| | option | signature | breaking? | call sites that move (at `f651409`) | what it does to `RenderTo` |
|---|---|---|---|---|---|
| **A** | **Bytes with a non-nil error** | unchanged | no | **0** | unchanged — and the flaw is identical |
| **B** | **A result struct returned** | `func Render(t, d, p, f) (Result, error)` where `Result{PDF []byte; Diagnostics []diag.Diagnostic}` | **yes** | **72** `Render` calls in 25 files, + `README.md`'s fenced block, + `testdata/swapproof/{good,bad,convert}` | **`RenderTo` gains a second return**, `(Diagnostics, error)` — its bytes went to the writer, so only the diagnostics come back. **7** call sites |
| **C** | **A diagnostics sink parameter** | `func Render(t, d, p, f, sink diag.Sink) ([]byte, error)` (or a `func(diag.Diagnostic)` callback) | **yes** — sixth positional argument | **72** + the same three artifacts | `RenderTo` gains a **seventh** argument. **7** call sites. **And it re-opens D-1.7.1's question**: a sixth/seventh positional argument on a five-argument function whose ordering is already protected by a compile-time swap proof |
| **D** | **A sibling entry point** | `func RenderWithDiagnostics(t, d, p, f) ([]byte, []diag.Diagnostic, error)`, `Render` untouched | **no** — purely additive | **0** | needs a sibling too (`RenderToWithDiagnostics`) or `RenderTo` silently never warns. **2 new public functions, 4 total to keep in step** |
| **E** | **Prior art the codebase suggests: nothing, and that is the finding** | — | — | — | — |

**Option A must be disposed of on the record rather than omitted.** It is almost certainly wrong,
and precisely because it *is* the zero-cost option someone will reach for. Every Go convention says
discard results when `err != nil`; `errors.As` on a warning type requires a caller who already knows
to look; and the project's own `RenderTo` **returns early on a non-nil error from `Render` and never
writes the bytes** — so under (A) a warning would make `RenderTo` produce *no output at all* for a
document that rendered perfectly. **(A) is not merely unidiomatic here; it is measurably broken for
half the public surface.** Recorded so it is disposed of, not skipped.

**On (B) and prior art — do not over-apply D-1.7.1.** D-1.7.1 rejects an options struct and it is
quoted in the source, but read what it rejects: *"an options struct would make `f` **omittable at
compile time**… turning an AD-8 violation from a compile error into a runtime one."* That is an
argument about **arguments**, not about returns. A `Result` **return** type omits nothing and makes
nothing optional. **D-1.7.1 does not reach (B), and citing it against (B) would be exactly the
vacuous citation D-000.55's neighbours warn about.** It *does* reach (C), squarely.

**On (D) — the prior art in this project is real, and so is its measured cost.** `Render`/`RenderTo`
*are* a sibling pair, added additively by Story 1.7 under D-1.1.c (*"`folio.Render`'s signature is
untouched"*), which is the closest thing here to an established convention. **But finding 10
measures the cost:** AD-1's forbidden-import guard matches the *names* `Render` and `RenderTo`, so a
new entry point is outside it until the guard is re-keyed on purpose rather than on name. And (D)
doubles the surface permanently: four entry points, two of which cannot report what the other two
can, which is the shape that produces "why is my warning missing" bug reports for the life of the
library.

**Looked for prior art before inventing, and here is the honest result (E).** The module root
declares no result-carrying struct on any public path; `Template`, `Data`, `Params`, `FontSet` are
all inputs; `layout.OverflowError` is a typed *error*, not a warning carrier; and there is no
existing collector, sink, callback or context-like value anywhere on the render path. **The project
has no existing shape that fits, which is itself worth reporting** — the answer is not hiding in the
codebase, and no option below is "the one that matches what we already do."

**`RenderTo`'s problem is genuinely not identical to `Render`'s, and that asymmetry should be part
of the ruling, not discovered after it.** `Render` returns bytes, so a caller who ignores an added
return value is merely uninformed. `RenderTo` returns only an `error`; its bytes have already gone
to the writer. Under (B) or (D) its warning channel is a return value nobody currently reads;
under (C) it is a parameter nobody currently passes. **There is no option under which `RenderTo`'s
existing seven call sites keep compiling *and* start seeing warnings**, except (A), which breaks
`RenderTo` outright.

**DO NOT PICK ONE.** This is the shape of the product's front door. The developer must not choose;
the story below is written so that development can proceed without the answer.

### OD-2 — is a text element's declared `height` a clip bound at all?

**This story fences the vertical axis OUT, and argues the fence below. It is surfaced rather than
merely asserted, because it is a scope fence with teeth and because the measurements behind it are
new.**

The argument for the fence:

1. **Nothing reads it today** (finding 1). Making `height` a clip bound mints a new format
   semantic, and D-000.6 would require `folio-format.md` to be amended to state it.
2. **It contradicts FR30 as shipped** (finding 6). `multi-page` `e1` declares 700 pt and flows 29
   lines across pages by design. A vertical clip cuts it to ~21 lines and the fixture stops proving
   the FR it was built for.
3. **Seven of the ten vertical overflows are systematic, not authorial** (finding 5): the ruled
   line box at 16 pt exceeds a declared 24 pt by 7.3 %. Clipping there shaves the descenders off
   every line of the Thai reading fixture — **the artifact under a pending owner sign-off.**
4. **The spine's only concrete FR44 instance is horizontal.** AD-25 is the one place in the
   architecture that names a specific thing FR44 clips, and it is an unbreakable Thai run running
   past the *side* of its box.
5. **The story's own subject supports it.** The single element in the whole fixture corpus that FR44
   was visibly designed around — `wrapped-text` `e4` — overflows horizontally and fits vertically.

**If the lead or owner rules the vertical axis IN, that is a separate story**, and it must come with
a disposition for `multi-page` `e1` and for the seven `shaped-text` elements, both of which move
goldens under review. **Do not implement it inside this one on a reading of "declared bounds".**

### OD-3 — may the clip SHIP before the diagnostic channel exists?

This is the second-order consequence of OD-1 and it deserves an explicit answer rather than a
developer's judgement call at 6 pm.

**The tension:** source AC1 asks for clip **and** diagnostic, and forbids *silently* dropping.
Shipping the clip alone would cut `wrapped-text` `e4`'s Thai surname at the box edge **with no way
whatever to report it** — which is strictly worse than today's visible overflow: today the author
can at least *see* the problem on the preview, which is the exact remedy the story's own "So that"
clause asks for. A clip with no channel converts a visible defect into an invisible one.

**Three dispositions, framed, not chosen:**

- **(i)** Clip ships only with the channel. This story delivers detection, identification and the
  never-reflow/never-drop guarantees; the clip and the diagnostic land together once OD-1 is ruled
  — in this story if the ruling arrives during development, or in a follow-up. **FR44 does not ship
  in Epic 2**, and the boundary gate must be told so.
- **(ii)** Clip ships now, diagnostic follows. `wrapped-text`'s golden moves, FR44's clip half is
  cross-target-proven at the gate, and the library ships a known-silent lossy path for one story.
- **(iii)** Neither ships; this story is detection-only by design and FR44 moves wholesale to a
  story after `internal/diag` exists.

**The ACs below are written for (i)** — they are the subset that is true under all three — **and the
clip-emission ACs are left unwritten and clearly marked.** If (ii) is ruled, the unwritten ACs are
written then, against the argument in *Golden movement* below, which is already made.

---

## The gate obligation question — argued against D-2.6.2's criterion, and it comes out NO

The brief asks whether the Epic 2 gate's **six** obligations (D-2.7.4) should become **seven**,
argued rather than asserted.

**D-2.6.2's criterion, verbatim:** *a gate obligation is warranted when it is **the only
cross-target artifact for a shipped FR***. Its settling sentence: *"If refused, FR30 ships with no
cross-target artifact."*

Apply it:

1. **Does FR44 ship in this story?** **Under OD-3 (i), no** — detection ships, clipping and the
   diagnostic do not. The criterion is about *a shipped FR*, and an unshipped FR cannot lack a
   cross-target artifact for its behaviour. **Under (i) the question does not arise.**
2. **If FR44 does ship (OD-3 (ii)), does it already have a cross-target artifact?** **Yes.**
   Measured: `wrapped-text` is already a registered `matrixDocuments` entry (Story 2.4, legs RUN
   in-story under a D-000.4 override), and it contains **the one element in the corpus that FR44's
   clip acts on** (`e4`, +22.9 % horizontal, finding 4). The moment the clip lands, `wrapped-text`'s
   four legs are a cross-target proof of FR44's clip behaviour. **Refusing a seventh entry does not
   leave FR44 with no cross-target artifact**, which is precisely the sentence D-2.6.2 says settles
   it — and here it settles it the other way.
3. **Therefore: no seventh obligation, under either disposition.** `declaredEpic2GateObligations`
   stays byte-unchanged. This is the same *decline* D-000.42 calls a third guard category applied to
   a gate entry: adding one that duplicates existing coverage is not caution, it is cost.

**Two consequences that follow from that answer and must not be lost:**

- **D-000.54's native leg is NOT owed**, because it attaches to a **newly registered** matrix
  document and none is registered. **But `wrapped-text` is an existing registered document whose
  bytes this story may move**, and re-recording a registered document's golden is a different
  obligation from registering a new one: its four legs must agree on the *new* bytes before the
  digest is trusted, and Story 2.4's in-story override is the precedent for how. **State which of
  the two you are doing; do not borrow one's cadence for the other.**
- **Any re-recorded golden owes D-000.53 provenance** — accepted only once a reader we did not write
  resolves it into the objects it claims to contain — plus a `validated at <commit>` anchor (Story
  2.7 review, Finding 15).

---

## Golden movement — argued explicitly, as the brief permits

**The claim: `fixtures/wrapped-text/expected.pdf` MOVES, and every other committed golden does
not.** Made now so the argument exists before the bytes do; it must still be **measured** after
implementation, never inherited from this section.

**Why it must move.** `wrapped-text` `e4` is painted 4 585 mp outside its box today (finding 4).
FR44 says that paint is clipped. **There is no byte-neutral way to stop painting something that is
currently painted.** Any implementation that satisfies source AC1's clip clause changes this
document's content stream. Refusing to move it means refusing to implement FR44 for the only
subject in the repository that exhibits it.

**Why nothing else moves.** The clip must be emitted **only for elements that actually overflow** —
not wrapped around every element defensively. That is not a stylistic preference: it is forced by
the brief's byte-identity constraint, since a clip path around every element would change **all**
eight pre-existing goldens plus `page-count-20`. Measured, exactly one element in the corpus
overflows horizontally, so exactly one document's bytes change.

**Why the two pending sign-offs survive.** They bind to `shaped-text/expected.pdf` and
`expected-breaks/expected_breaks.json` (D-000.26 refined). Neither contains a horizontally
overflowing element; `shaped-text`'s seven elements overflow **vertically**, which OD-2 fences out.
**If OD-2 is ruled the other way, both sign-offs are invalidated and this paragraph is void.**

**Why `expected_breaks.json` cannot move regardless.** It is a record of **break opportunities**.
Clipping changes what is painted, never where a line may break — the packer is untouched (finding
3). Its digest moving would itself be the defect signal.

---

## Scope fence — what this story is NOT

- **Not the diagnostic channel.** No new public function, no changed signature, no new return
  value, no new parameter, until OD-1 is ruled. **The developer must not choose.**
- **Not the vertical axis** (OD-2). A text element's declared `height` stays what it is today: read
  by nothing. Do not "just also clip the bottom."
- **Not `internal/diag`.** Story 3.6 mints it and its code registry (DW-6). If OD-1's ruling
  requires it earlier, that is a ruling to be recorded, and DW-6's missing tripwire (finding 8) is
  owed **first, in the same commit**.
- **Not a re-litigation of D-2.6.5 or D-2.7.3.** Page-window overflow stays a located error. This
  story does not route `layout.OverflowError` through anything.
- **Not images.** AD-24 fits and centres them; an image cannot overflow its box, and *"never
  cropped"* is AD-24's word.
- **Not tables, lines or rects.** Tables are Epic 4 and declare no box (AD-13); `line` and `rect`
  are parsed and rendered by nothing.
- **Not the line packer.** *"It goes on this line and overflows visibly"* is Story 2.4's ruled
  behaviour and stays exactly as it is. Clipping is about paint, never about breaking.
- **Not a seventh gate obligation.** `declaredEpic2GateObligations` stays byte-unchanged; the
  argument is made above rather than assumed either way.
- **Not `epic-2: backlog`.** That line stays. The **gate** flips it, not this story.
- **Not a fix for DW-6's missing tripwire.** Flagged below; fixing it belongs to whoever mints the
  package.

---

## Acceptance Criteria

**AC1–AC5 are written and developable now. AC6 and AC7 are DELIBERATELY UNWRITTEN and the developer
must not write them** — see the marked section after AC5.

### AC1 — Overflow is DETECTED per element, on the horizontal axis, and identified by element id

Source AC1's *"naming the element id"*, delivered as the value that every disposition of OD-1 needs
and none of them changes.

**Given** a text element whose widest packed line exceeds its declared `width`,
**When** the document is laid out,
**Then** the engine holds a per-element record carrying **the element's id** (AD-10: *"every
diagnostic and every error that concerns a template element carries its id"*), the axis, the
declared bound and the measured extent — **and the render still succeeds**, returning bytes, exactly
as it does today.

Assert on **the fields**, never on "something was detected" (D-000.21, and D-2.6.5's guardrail
applied one story on). A count of overflowing elements is not the assertion; the id is.

**Red-proof, and it must be captured BEFORE the change (D-000.30):** the probe in finding 2 —
34 CIDs in a 20 pt box, `Render` returning nil — is the pre-state. Commit the measurement, not a
recollection of it.

**Subject population (D-000.50):** `wrapped-text` `e4` is a real, committed subject and must be one
of the cases. It must **not** be the only one: a synthetic template exercising a Latin atomic token
belongs beside it, because a single-subject assertion would tie the property to the Thai dictionary
path.

### AC2 — Overflowing content is NEVER reflowed and NEVER dropped, and this is proven by a subject that could exhibit either

Source AC1's third consequent, in full.

**Given** an element whose content exceeds its declared `width`,
**When** it is laid out,
**Then** its glyph sequence, its line count and its break positions are **identical** to the same
element laid out in a box wide enough to hold it on one line — nothing is re-broken, nothing is
substituted, nothing is omitted — **and** every sibling element's position is byte-for-byte
unchanged from a document in which the overflow does not occur (AD-24: *"siblings never move,
because nothing in a band ever reflows"*).

Finding 2's two-width measurement is the shape of this assertion. **Both halves are required**: an
assertion that the glyphs survive, alone, is satisfied by the current code and would ship as a
tautology.

### AC3 — The declared `height` remains inert, and the guard says so on purpose

The OD-2 fence, made mechanical rather than left as prose.

**Given** any text element,
**When** the document renders,
**Then** changing **only** its declared `height` changes **no** emitted byte.

This is a **forward guard** (D-000.24: a named, permitted category) with a genuine red-proof
available — the guard fails the moment someone wires `height` into layout or emission — and it is
keyed on the **purpose** (*height does not affect output*), not on the proxy *"no code reads
`el.Height`"*, which a future refactor could satisfy while breaking the property (D-000.15).

**Subject:** it must be run against an element that would visibly change if `height` became a clip
bound — `shaped-text`-shaped geometry (single line, 16 pt, `height: 24`), and a multi-line element.

### AC4 — The eight pre-existing goldens and `page-count-20` are byte-identical, MEASURED not argued

**Given** the implementation,
**When** every committed golden is re-rendered,
**Then** all eight pre-existing goldens plus `page-count-20` hash exactly as recorded at
`f651409`, **including** the three pinned digests reproduced at the top of this file — **and** the
measurement is performed and recorded, not inferred from the scope fence.

`fixtures/wrapped-text/expected.pdf` is **in this set under AC1–AC5**, because nothing is clipped
yet. It leaves this set only if AC6 is written (see below), and only under the argument in *Golden
movement*.

Story 2.6a could derive non-movement from its scope fence because it edited only test files. **This
story edits the renderer's layout path** and can move every golden in the repository. Measure it
first, not last.

### AC5 — The baseline is preserved, every delta is itemised, and every enumeration names its commit

**Given** the delivered work,
**When** the ordinary suite runs (`./... -count=1`, no build tags),
**Then** the counts are reported against the recorded baseline — **587 / 1** all-occurrences and
**359 / 1** top-level at `f651409` — with **every** delta itemised test-by-test, and the one
expected failure still `TestCorpusMeetsP6ExerciseFloors` with its stats reproduced.

Every count, inventory or enumeration this story records **names the commit it was taken at and
states the tree was clean in the files counted** (D-2.6.9 correction). A figure without its scope
(*whole module / build tags / `-count=1` / all-occurrences vs top-level*) is not comparable and must
not be written down (`[[carried-gate-figures-are-scoped]]`).

And per D-000.55: `go vet -tags matrix` is a **compile gate, not an honesty check**. If a matrix
file is vetted and not run, say those two things separately. **The compound phrasing
"compiled and vetted, deliberately not run" is banned.**

---

### AC6 and AC7 — WRITTEN AND IMPLEMENTED (superseded from "deliberately unwritten")

**The premise of this section's original heading no longer holds.** OD-1 was ruled by the owner
(D-2.8.3: `Result`; D-2.8.6: `RenderTo` returns `([]Diagnostic, error)`) and OD-3 resolved to "ship
both, or neither" → **both** (D-2.8.2), before this story left `ready-for-dev`. The text below is kept
verbatim as the historical record of why the ACs were originally left blank; it does not describe what
shipped. See the Dev Agent Record for AC6/AC7 as actually implemented and tested
(`render_clip_diagnostic_test.go`).

**AC6 (the clip itself) and AC7 (the diagnostic channel) are the two clauses of the epic that this
story cannot state, and leaving them blank is the deliverable, not an omission.** This is the same
disposition Story 2.7's creator took with its DN-1 width question, which worked.

**AC7 is blocked on OD-1 and nothing else.** Source AC1's *"a diagnostic is reported"* and the whole
of source AC2 (*"PDF bytes are still returned"*) both describe a seam that does not exist. Writing
an AC for it would fix a signature by implication, and the signature is the owner's. **Whatever the
ruling, AD-14 already fixes the payload** (`Severity`/code/element id/path/message, `Warning`
alongside bytes) — so when AC7 is written, it is written against AD-14's value, not a new one.

**AC6 is blocked on OD-3, which is blocked on OD-1.** It is technically writable — clip at the
element's declared left/right edge, no mark painted outside it, `wrapped-text`'s golden re-recorded
under the *Golden movement* argument — but under OD-3 (i) shipping it without AC7 converts a
**visible** defect into an **invisible** one, which is the opposite of this story's own "So that"
clause. **The developer must not decide that trade.**

**What the developer does instead:** build AC1–AC5, which every disposition of OD-1 and OD-3 needs
and none of them invalidates, and stop. If a ruling arrives mid-development, the ACs are written
then, by whoever rules — not inferred from the ruling by the person holding the keyboard.

**And state it for the gate:** under OD-3 (i), **FR44 does not ship in Epic 2.** The boundary gate
must be told that in terms, because Epic 2's FR coverage table claims FR44 and the gate is the
owner's observation point (D-000.3).

---

## Heavy-test cadence — proposed DECLINED, stated so it can be refused

**D-000.4's override criterion is DECLINED, and the argument is the same one D-2.6.2 and D-2.7.4
both reached.** Overflow *detection* is a comparison of two `geom.Length` values already computed by
the line packer and the vertical model: **no float, no vendor call, no compressor, no new
dependency, no new source of cross-target divergence.** D-000.4's own warning — that on a weaker
trigger *"nearly every story would qualify"* — applies exactly.

**If AC6 is later written**, re-run this argument rather than inheriting it: emitting a clip changes
a content stream, which changes compressed bytes, which is the one place this epic has repeatedly
found divergence. The disposition may legitimately flip. **Do not carry this paragraph forward
unexamined.**

---

## Task breakdown

**Superseding note, read before the checklist below.** OD-1, OD-2 and OD-3 — open when this story was
created — are now ALL RULED (D-2.8.1 through D-2.8.6, recorded at `7f97ef7` and `278520b`, both
ancestors of this story's development baseline `278520b`). The owner chose option (B), a `Result`
struct (D-2.8.3); `RenderTo` returns `([]Diagnostic, error)`, not `Result` (D-2.8.6); the vertical axis
stays fenced out (D-2.8.1, unchanged from this story's own argument); and OD-3 resolved to **(i) ship
both, or neither** — **both landed** (D-2.8.2). So items 9-11 below, written for the old "AC6/AC7 stay
unwritten" disposition, are **superseded**: AC6 (the clip) and AC7 (the diagnostic channel) ARE built
in this story, per the rulings, and `declaredEpic2GateObligations`/`folio-format.md`/the line packer
are handled exactly as the rulings require (declaredEpic2GateObligations byte-unchanged per D-2.8.4;
folio-format.md amended per D-000.6; the line packer untouched).

1. [x] **Re-measure the baseline and the three pinned digests** on a clean tree at the development
   commit `278520b`. `go test ./... -count=1 -v` (whole module, no build tags, all-occurrences and
   top-level counted separately): reproduced the same **587/1** all-occurrences, **359/1** top-level,
   with the single expected failure `TestCorpusMeetsP6ExerciseFloors` at the identical stats
   `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. The three pinned digests (`shaped-text`,
   `expected-breaks`, `multi-page`) reproduced exactly as stated at the top of this file.
2. [x] **Capture the pre-change red-proof** (D-000.30): finding 2's two probes are the pre-state this
   story's `TestWidthOverflowDetectedPerElement`/negative-control pairing and
   `TestOverflowingContentNeverReflowedNeverDropped` now formalise as permanent tests
   (`render_clip_diagnostic_test.go`) — the "34 CIDs painted in a 20pt box, nil error" shape is now the
   WIDE-box negative control (`clipWideTemplate`) proving nothing is dropped even where nothing clips.
3. [x] **Re-ran the fixture-population measurement** (finding 4) at `278520b`: the eleven overflows and
   the single horizontal one (`wrapped-text` e4) reproduced exactly. No drift.
4. [x] **Built AC1**: `detectWidthOverflow` (render.go) — per-element horizontal overflow record
   carrying element id, declared width, measured width. **Public surface change and emission change
   BOTH happened**, per the now-ruled OD-1/OD-3 — see the superseding note above.
5. [x] **Built AC2**: never-reflowed / never-dropped, both halves, plus the sibling-position assertion
   (`TestOverflowingContentNeverReflowedNeverDropped`).
6. [x] **Built AC3**: the `height`-is-inert forward guard (`TestDeclaredHeightIsInert`), single-line AND
   multi-line subjects, extreme height deltas in both directions.
7. [x] **Built AC4**: all nine pre-existing goldens plus `page-count-20` measured byte-identical
   EXCEPT `fixtures/wrapped-text/expected.pdf`, which moves under the *Golden movement* argument —
   re-recorded, re-validated (qpdf, D-000.53) and re-run across all four matrix legs
   (`TestCrossTargetByteIdentity`, PASS).
8. [x] **Built AC5**: baseline delta below, itemised, commit named throughout.
9. [x] **`declaredEpic2GateObligations` untouched** (D-2.8.4: six confirmed, seventh declined).
   **`folio-format.md` WAS amended**, deliberately, per D-000.6 (the `x,y,width,height` row now states
   FR44's width-clip/height-inert outcome) — this is the ruling superseding the old "do not touch"
   instruction, not a deviation from it. The line packer is untouched. Both public signatures changed,
   per the owner's ruling (D-2.8.3/D-2.8.6) — the one deliberate exception to "no public surface
   change" this task list originally specified.
10. [x] **Delivery Log written** (below), stating in terms that AC6/AC7 are now WRITTEN (the old
    disposition is superseded) and that FR44 SHIPS in Epic 2. `epic-2-boundary-gate.md` appended with
    the resolution (OD-1 ruled, both halves shipped, six obligations confirmed) rather than the
    now-moot "unruled OD-1" framing.
11. [x] **Story set to `review`.**

---

## Flagged, not fixed

1. **DW-6's anti-rot tripwire does not exist** (finding 8). `deferred-work.md` states that a test
   asserts `folio-go/internal/diag` is absent; no such test is in the tree at `f651409`. The
   deferral's forcing function has been believed since Story 1.4 and was never built. **Owed by
   whoever mints the package, in the same commit.**
2. **`epics.md`'s FR coverage table claims FR44 for Epic 2.** Under OD-3 (i) that becomes false at
   the Epic 2 boundary. D-000.6 would make it an amendment owed by whichever story actually ships
   FR44 — recorded here so the gate sees it.
3. **AD-1's forbidden-import guard is keyed on two function names** (finding 10). Independent of
   OD-1, that is a proxy where a purpose belongs (D-000.15). Not this story's to fix; it becomes
   urgent the moment a third public entry point exists.
4. **`shaped-text`'s seven elements are under-declared by 7.3 % against the ruled line box**
   (finding 5). Whatever OD-2 decides, `folio-format.md` documents no relationship between a
   declared `height` and the line box an author will actually get, and an author has no way to
   compute one. Recorded as a documentation gap, not repaired here.
5. **`multi-script-fallback` `e1` overflows its declared height by 12.7 %** and, unlike
   `shaped-text`, by a margin that is a mixed-script line-box effect. Recorded with the others.

---

## Dev Agent Record

### Delivery Log

**Baseline.** Development commit `278520b` (main). Tree clean at start except the story creator's own
uncommitted `sprint-status.yaml` line (`2-8: backlog → ready-for-dev`) and this story file. `go test
./... -count=1 -v` reproduced **587/1** all-occurrences, **359/1** top-level, single expected failure
`TestCorpusMeetsP6ExerciseFloors` at identical stats. All three pinned digests reproduced exactly.

**All three open decisions (OD-1/OD-2/OD-3) were ruled BEFORE development started** — D-2.8.1 through
D-2.8.6, recorded at `7f97ef7` and `278520b`, both ancestors of the development baseline. This story
therefore builds AC1–AC7 in full, not AC1–AC5 with AC6/AC7 left blank as originally planned. See "Do
not re-open" at the top of this file and the superseding notes inline in the Task breakdown and AC6/
AC7 sections above.

**What was built, against the rulings:**

1. **`folio.Result` / `folio.Diagnostic` / `folio.Severity`** (`diagnostic.go`, new file, package
   `folio` at the module root — NOT `internal/diag`, per D-2.8.3/D-2.8.6's unqualified `Diagnostic`).
   `Diagnostic` carries exactly AD-14's five fields (`Severity`, `Code`, `ElementID`, `DataPath`,
   `Message`). One code minted: `DiagCodeTextClippedWidth = "TEXT_CLIPPED_WIDTH"`.
2. **`Render` returns `(Result, error)`**; **`RenderTo` returns `([]Diagnostic, error)`, not
   `Result`** — the asymmetry D-2.8.6 rules deliberately, documented in both functions' doc comments
   (`render_entry.go`). `RenderTo`'s early-return defect (a Warning would previously have produced no
   output at all under a bare-error seam) is closed by construction: a Warning never makes `Render`'s
   `error` non-nil, so `RenderTo` always proceeds to write complete bytes on the Warning-only path.
   Red-proved: `TestRenderToWritesCompleteBytesDespiteWarning`.
3. **`detectWidthOverflow`** (`render.go`) — AC1's per-element horizontal overflow record (element id,
   declared width, measured width — the widest packed line's `wrappedLine.width`, already computed by
   `packLines`; no second measurement). Wired into `collectBandTextRuns`, which now also returns
   `[]Diagnostic` in element-declaration order within its own band.
4. **`renderDocument`** concatenates each band's diagnostics in BAND order (header, content, footer —
   D-2.8.6's document-order rule), which is NOT the order the three `collectBandTextRuns` calls
   execute in (content is collected first, under Story 2.7's Phase A/B split, to learn the page count
   before header/footer resolve `{{page}}`). Tested directly:
   `TestDiagnosticsOrderIsDocumentOrder` (four overflowing elements across all three bands).
5. **Empty-is-nil** (D-2.8.6's second determinism rule): `Diagnostics` is never reassigned to a
   non-nil empty slice; a fitting document's `Result.Diagnostics` is `nil`, asserted directly
   (`TestDiagnosticsEmptyIsNil`, `!= nil` check, not `len() == 0`).
6. **The clip (AC6)** — `pagemodel.TextRun` gained `ClipToBox`/`ClipX`/`ClipWidth`
   (`internal/pagemodel/pagemodel.go`); `internal/pdf/textdoc.go`'s `buildTextContentStream` wraps a
   clipped run's `BT..ET` block in `q / <x> 0 <width> <pageHeight> re / W n / ... / Q`. The clip
   rectangle's vertical extent is the WHOLE PAGE (`0` to `page.Height`), never the element's own
   declared height — D-2.8.1's fence, made mechanical: `TestDeclaredHeightIsInert` proves changing
   only an element's declared height changes no emitted byte, for both a single-line (shaped-text
   geometry, the case D-2.8.1's grounds name as the one that WOULD visibly clip descenders if height
   ever bound) and a multi-line subject, at extreme height deltas in both directions.
   `TestClipRestrictsOnlyTheHorizontalAxis` reads the produced clip rectangle back off the actual PDF
   bytes and confirms it uses the overflowing element's own box, never the fitting sibling's.
7. **AC2, both halves**: `TestOverflowingContentNeverReflowedNeverDropped` compares a narrow-box
   (overflowing) render against a wide-box (fitting) control of the IDENTICAL atomic word
   (`Supercalifragilisticexpialidocious`, finding 2's own probe subject) — glyph CIDs, origin, AND the
   untouched sibling element's position and glyphs, all asserted equal.
8. **AC7's producer/consumer agreement**: `TestRenderAndRenderToDiagnosticsAgree` renders one clipping
   document through both entry points and asserts `Render`'s `Result.Diagnostics` equals `RenderTo`'s
   returned slice via `reflect.DeepEqual`, THEN pins `Render`'s side against a literal `[]Diagnostic`
   so the pair cannot both drift together undetected (D-2.8.6's explicit requirement).

**`folio-format.md` amended** (D-000.6): the `x, y, width, height` row now states, as an outcome, that
a text element's declared `width` bounds and clips laid-out content (FR44) while `height` does not,
and that an image's `height` is honoured (AD-24 fit-and-centre) and reserved for `valign`.

**`deferred-work.md`**: DW-6's anti-rot line corrected (D-2.8.4 — it names a lint rule, not a test;
this story did NOT create `internal/diag`, so the obligation stays open and un-owed by this story).
New entry **DW-17** files D-2.8.5's forward obligation (surfacing a returned `Diagnostic` to a human is
owned by Stories 3.7/5.12/6.6, on the presented output — explicitly NOT recorded as an accepted risk
against AD-14, per D-000.49).

**Golden movement.** `fixtures/wrapped-text/expected.pdf` moved (the only committed golden this story
touches): `277bc5c0…` → `07c38cf7…` (72,738 → 72,790 bytes). Cause: `e4`'s Thai surname, previously
painted 4,585 mp outside its 20,000 mp box, is now clipped there and carries one `Diagnostic`. Every
other pre-existing golden (all eight, plus `page-count-20`) is byte-identical, measured
(`TestGoldenDigestAgreesAtEveryDeclaredSite`, all sites agree). The two pending sign-offs
(`shaped-text`, `expected-breaks`) are untouched — neither contains a horizontally overflowing
element, and the vertical axis stays fenced out (D-2.8.1). `wrapped-text`'s four cross-target legs were
re-run against the new bytes (`go test -tags=matrix -run TestCrossTargetByteIdentity .`, PASS — all
four targets agree on `07c38cf7…`, 72,790 bytes). External-reader provenance re-validated:
`qpdf --check` (exit 0, no syntax/stream errors) and `qpdf --show-npages` (1 page, matches `/Count`),
qpdf **12.4.0**. `fixtures/wrapped-text/README.md` and `expected.json` updated accordingly; full
argument at `fixtures/wrapped-text/README.md`'s new "Story 2.8" section.

**Gate note appended** to `epic-2-boundary-gate.md` ("Story 2.8" section): FR44 now ships in Epic 2
(superseding the earlier "FR44 does not ship" framing, which anticipated an unruled OD-1); gate
obligations stay at SIX, seventh declined (D-2.8.4); `wrapped-text`'s digest movement noted
explicitly so it is not mistaken for an undeclared obligation; `internal/diag` was NOT created by this
story, so DW-6's tripwire stays untouched and green.

**Call-site count, re-derived at this story's own (pre-commit) working tree, method named** (D-2.6.9
correction — not inherited from either prior number): a `go/ast`-based `CallExpr` scan (structurally
excludes comments, string literals and declarations — no text-stripping heuristic needed) found **75**
`Render` call expressions across **26** files (4 non-test: `render_entry.go` and the three
`testdata/swapproof/*/main.go` fixtures) and **9** `RenderTo` call expressions across **2** files (0
non-test). This is a THIRD number, deliberately not reconciled to the creator's 72 (25 files, `f651409`)
or the orchestrator's 55 (`7f97ef7`) — each was taken at a different commit with a different method,
and this story's own diff legitimately added new call sites in its own new test file. Full detail in
`epic-2-boundary-gate.md`'s "Story 2.8" section, item 7.

**AD-1 confirmed, not assumed** (D-2.8.6's guardrail): `render_entry.go` still declares top-level
functions named exactly `Render` and `RenderTo`, so lint's forbidden-import scan keeps locating them —
confirmed by running `lint`'s full suite (`go test ./...` under `lint/`, 85 passed) rather than
inferred from the names being unchanged.

**`go vet -tags matrix ./...`**: executed (not merely compiled), clean, zero issues (D-000.55: stating
what was executed, on which target, by name — never the banned "compiled and vetted, deliberately not
run" compound).

**Test suite, final state** (corrected by the finisher — review Findings 4 and 5; see "Finding
Resolutions" below): `go test ./... -count=1 -v` from `folio-go/` — **600 PASS / 1 FAIL
all-occurrences, 368 PASS / 1 FAIL top-level**, whole module, no build tags, both counts named per
AC5's own scope rule. The one failure is the same pre-existing `TestCorpusMeetsP6ExerciseFloors`,
unrelated to this story, reproduced at the identical stats. **9** new top-level test functions
added (`render_clip_diagnostic_test.go`, `grep -c '^func Test' render_clip_diagnostic_test.go`), all
passing — not 13 as originally stated; the baseline top-level count (359) plus 9 correctly predicts
the measured 368.

### File List

**New:**
- `folio-go/diagnostic.go` — `Result`, `Diagnostic`, `Severity`, `DiagCodeTextClippedWidth`.
- `folio-go/render_clip_diagnostic_test.go` — AC1/AC2/AC3/AC6/AC7 tests.
- `_bmad-output/implementation-artifacts/2-8-clip-overflowing-content-and-say-so.md` — this file
  (created by the story creator; developer edits are the permitted sections only).

**Modified — production code:**
- `folio-go/render.go` — `detectWidthOverflow`, `widthOverflow`, `millipoints`; `textRunSource` gained
  clip fields; `collectBandTextRuns` returns `[]Diagnostic`; `renderDocument` returns
  `([]byte, []Diagnostic, error)` and orders diagnostics in band order; `buildShapedPDFRuns` threads
  clip fields into `pagemodel.TextRun`.
- `folio-go/render_entry.go` — `Render` returns `(Result, error)`; `RenderTo` returns
  `([]Diagnostic, error)`; doc comments updated.
- `folio-go/internal/pagemodel/pagemodel.go` — `TextRun` gained `ClipToBox`/`ClipX`/`ClipWidth`.
- `folio-go/internal/pdf/textdoc.go` — `buildTextContentStream` emits the clip (`q/re/W n/Q`).

**Modified — mechanical call-site adaptation to the new `Render`/`RenderTo` signatures** (no logic or
assertion changes; `Result`/`[]Diagnostic` unwrapping only):
`folio-go/ac4_coverage_test.go`, `folio-go/byte_neutrality_test.go`,
`folio-go/collect_text_runs_composition_test.go`, `folio-go/example_test.go`,
`folio-go/first_baseline_acceptance_test.go`, `folio-go/fixture_test.go`,
`folio-go/internal/fontset/vendorboundary_test.go`, `folio-go/multi_page_composition_test.go`,
`folio-go/multi_page_fixture_test.go`, `folio-go/page_count_matrix_test.go`,
`folio-go/render_bind_test.go`, `folio-go/render_image_test.go`, `folio-go/render_test.go`,
`folio-go/renderto_test.go`, `folio-go/reserved_placeholders_test.go`,
`folio-go/segment_origin_test.go`, `folio-go/shaped_fixture_test.go`, `folio-go/template_test.go`,
`folio-go/three_band_page_fixture_test.go`, `folio-go/wrapped_text_fixture_test.go`.
- `folio-go/README.md` — resynced verbatim from `example_test.go` (`TestREADMEExampleBlockMatchesSource`).

**Modified — fixtures (the one deliberate golden movement):**
- `fixtures/wrapped-text/expected.pdf`, `fixtures/wrapped-text/expected.json`,
  `fixtures/wrapped-text/README.md`.

**Modified — docs/tracking:**
- `_bmad-output/specs/spec-folio/folio-format.md` — `x,y,width,height` row amended (D-000.6).
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-6 corrected; DW-17 added.
- `_bmad-output/implementation-artifacts/epic-2-boundary-gate.md` — "Story 2.8" section appended.
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — `2-8-…`: `ready-for-dev` → `review`.

### Change Log

| Date | Change |
|---|---|
| 2026-08-25 | Story 2.8 implemented: `Result`/`Diagnostic`/`Severity` types; `Render`/`RenderTo` signatures changed per D-2.8.3/D-2.8.6; FR44's width clip and diagnostic built and shipped together (D-2.8.2); `fixtures/wrapped-text/expected.pdf` re-recorded and re-validated across all four matrix legs; `folio-format.md`, `deferred-work.md` (DW-6 correction, DW-17 added) and `epic-2-boundary-gate.md` updated; status set to `review`. |
| 2026-08-25 | Story 2.8 finished: review findings triaged (1 Blocker, 2 Majors, 2 Minors, 1 Nit — see "Finding Resolutions"). Blocker (the `Code` pin) and one Major (the panic-on-index in the agreement test) fixed and red-proved in `render_clip_diagnostic_test.go`; the other Major (`Severity`'s unfalsifiable zero value) deferred and filed as DW-18 in `deferred-work.md`, since closing it means changing a public AD-14-governed type, which is not this pass's call; both Minors (Delivery Log's test count and missing top-level figure) corrected; the Nit (a stale doc comment in `template_test.go`) fixed. No production code changed, no golden moved. Status set to `done`. |

### Completion Notes

**All tasks complete; all acceptance criteria (AC1–AC7) built and tested** — AC6/AC7 were originally
scoped as deliberately unwritten pending OD-1/OD-3, but both were ruled before development started
(D-2.8.1–D-2.8.6), so this story delivers the full epic clause, not the AC1–AC5 subset originally
planned.

**AC verification:**
- **AC1** (per-element horizontal detection, id/axis/bound/extent): `TestWidthOverflowDetectedPerElement`
  (real subject `wrapped-text`/e4 + synthetic Latin atomic token), `TestDetectWidthOverflowRecordsTheBoundAndExtent`
  (internal record fields).
- **AC2** (never reflowed, never dropped, siblings unmoved): `TestOverflowingContentNeverReflowedNeverDropped`.
- **AC3** (height inert, forward guard): `TestDeclaredHeightIsInert`.
- **AC4** (all nine pre-existing goldens byte-identical, `wrapped-text` moves, both measured):
  `TestGoldenDigestAgreesAtEveryDeclaredSite`, `TestWrappedTextGoldenFixture`,
  `TestEveryGoldenPDFResolvesItsPageTree`.
- **AC5** (baseline preserved, deltas itemised, commits named): this Delivery Log; `go test ./...
  -count=1` at 600 passed / 1 failed (the one pre-existing, unrelated failure), reproduced against the
  `278520b` baseline of 587/1.
- **AC6** (the clip, superseding "unwritten"): `TestClipRestrictsOnlyTheHorizontalAxis`,
  `fixtures/wrapped-text/expected.pdf`'s re-recording.
- **AC7** (the diagnostic channel, superseding "unwritten"): `TestRenderAndRenderToDiagnosticsAgree`,
  `TestRenderToWritesCompleteBytesDespiteWarning`, `TestDiagnosticsOrderIsDocumentOrder`,
  `TestDiagnosticsEmptyIsNil`.

**Regression tests**: full ordinary suite green except the one pre-existing, documented failure.
`go vet ./...` and `go vet -tags matrix ./...` both clean. `lint`'s own suite (85 tests) green,
confirming AD-1's forbidden-import scan and DW-6's `absence-diag-package` tripwire both still function
correctly against this story's changes.

**Not done, and correctly not done** (scope fence, unchanged by the rulings): the vertical axis
(D-2.8.1, unchanged); `internal/diag` (not created — `Diagnostic`/`Result` live in package `folio`
per D-2.8.3/D-2.8.6's unqualified naming); a seventh gate obligation (D-2.8.4); an AST guard asserting
every `Render`/`RenderTo` call site reads its diagnostics (D-2.8.5, declined); recording D-2.8.5's
concern as an accepted AD-14 gap (D-000.49 — filed as DW-17, a forward obligation with named owners,
instead).

---

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-25
- **Review baseline:** `278520b`, uncommitted working tree (32 files changed, 531 insertions, 137
  deletions — per-file diffstat re-measured identical at the end of the review, proving every
  mutation below was reverted by hand and nothing but this section was left changed)
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 1 · **Majors:** 2 · **Minors:** 2 · **Nits:** 1

**Method.** This review is execution-first. Every finding below was produced by *running* something,
and every mechanism the story claims was red-proved by mutating the production code and watching the
named test go red — then reverting by hand (never `git checkout`) and re-verifying the file's SHA-256
against its pre-mutation value. Eight mutations were applied and reverted. All figures are scoped:
*whole module, no build tags, `-count=1`, all-occurrences and top-level reported separately*.

### Suites MEASURED (not merely read)

| suite | invocation | result |
|---|---|---|
| ordinary unit suite | `go test ./... -count=1 -v` from `folio-go/` | **600 PASS / 1 FAIL** all-occurrences; **368 PASS / 1 FAIL** top-level |
| the one expected failure | — | `TestCorpusMeetsP6ExerciseFloors`, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — reproduced character-for-character, the intentional Story 2.1 red (D-000.17) |
| cross-target matrix | `go test -tags=matrix -run TestCrossTargetByteIdentity -count=1 -v .` | **PASS**, all four targets, all nine documents |
| `lint`'s own suite | `go test ./... -count=1` under `lint/` | all four packages green |
| `go vet -tags matrix ./...` | executed | exit 0, clean |
| external reader | `qpdf --check` / `qpdf --show-npages` on `fixtures/wrapped-text/expected.pdf` | qpdf **12.4.0**, exit 0, `No syntax or stream encoding errors found`, **1** page |

**The orchestrator's figures are confirmed, not contradicted:** 600/1 all-occurrences and 368/1
top-level. The story's own Delivery Log quotes 600/1 and declines to quote top-level — see Minor 5.

### Verified by execution — things that are RIGHT, recorded so the gate can rely on them

1. **All four matrix legs genuinely ran and agree** (attack item 2 — this is NOT "the native leg plus
   a vet", and there is no D-000.55 violation). Every one of the nine registered documents reported
   an identical digest on `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm`, including
   `wrapped-text` at `07c38cf765a39d86376c1a3c78bfb6f0a96f089f19792c9bfeeaa1dc754269d6` / **72,790
   bytes**. The Linux legs cannot be faked on this host: the harness builds ELF binaries that a
   darwin kernel cannot exec, runs them under `docker run --rm --platform …`, and treats failure as
   `t.Fatalf` ("*docker run for target %s FAILED — this is a failure, not a skip (AC11)*"); the wasm
   leg `t.Fatalf`s if `node` is absent. There is no silent-degradation path.
2. **The pinned digests are unmoved, confirmed independently** — `shaped-text` =
   `6c040ef7…c85370`, `expected_breaks.json` = `a545e042…9324de`, `multi-page` = `66ce0ee4…95869b`.
   `git diff --stat 278520b -- fixtures/` shows `wrapped-text` as the *only* fixture touched. Both
   pending owner sign-offs survive.
3. **The vertical axis is genuinely fenced out.** Nothing clips vertically and no diagnostic is
   emitted for vertical overflow. The clip rectangle read back off the shipped golden is
   `36 0 20 841.89 re / W n` — x = `margin.left` 36 + `e4.x` 0, width = `e4.width` 20, and the
   vertical extent is the **whole A4 page** (0 → 841.89), never anything derived from a declared
   height. The corpus's ten vertical overflows are untouched.
4. **The clip's PDF mechanics are sound** (attack item 7). `q`/`Q` cannot become unbalanced: both the
   `len(run.Glyphs) == 0` `continue` and both error returns in `buildTextContentStream` precede the
   `q` emission, so every `q` is reached only on a path that also reaches its `Q`. Measured on the
   shipped golden: `q`=2, `Q`=2, `W n`=2. Measured on a purpose-built 3-page probe: `q`=4, `Q`=4. No
   clip state leaks across elements or pages.
5. **Multi-page dedup is correct** (an intersection this story's own tests do not cover — probed
   directly): a **footer** element that overflows and carries `{{page}}`, rendered across **3** pages,
   produces exactly **ONE** `Diagnostic` (not one per page), with the page number still substituted
   correctly and the clip emitted per page. Digit-table runs (`digitTableBandIndex`) are explicitly
   skipped in `paginateDocument` and never drawn, so they cannot escape a clip.
6. **AD-1's forbidden-import scan still locates the entry points** (attack item 8) — confirmed by
   execution, not by the names being unchanged. Adding `"os"` to `render_entry.go` produced
   `TestRenderEntryFileHasNoForbiddenImports`: *"AC12: the file declaring Render must import none of
   os/time/net/math/rand, found: render_entry.go:37: forbidden import \"os\""*. Reverted; SHA-256
   restored.
7. **DW-6's tripwire exists, is a lint rule, and still fires** (attack item, and the
   `deferred-work.md` correction is right). `lint/internal/rules/absences.go` registers
   `absence-diag-package`; creating `folio-go/internal/diag/` made `TestAbsencesProductionScan` go
   red naming that rule. Directory removed; `git status` re-verified clean. `folio-go/internal/diag`
   does **not** exist — `Diagnostic`/`Result` live in package `folio`, exactly as the story says.
8. **`declaredEpic2GateObligations` is byte-unchanged.** The moved digest lives in
   `goldenDigestRecord` (`byte_neutrality_test.go:181`), a different structure from
   `declaredEpic2GateObligations` (line 479). The gate note states the digest-moves/obligations-don't
   distinction explicitly. Six obligations, seventh declined. `epic-2: backlog` is untouched
   (`sprint-status.yaml:49`).
9. **The call-site count re-derivation is exactly reproducible** (attack item 10). I wrote an
   independent `go/ast` `CallExpr` walker and got **75** `Render` call expressions across **26**
   files, **4** non-test — and the same four files (`render_entry.go`, `testdata/swapproof/{good,bad,
   convert}/main.go`) — plus **9** `RenderTo` across **2** files, **0** non-test. The developer's
   number, method and tree are all correct and correctly named; the refusal to reconcile 72/55/75
   across three different commits and methods is the right call (D-2.6.9).
10. **DW-17 is filed as a forward obligation, not an accepted risk** (D-000.49 honoured), with named
    owners 3.7 / 5.12 / 6.6 and a "how we'd know it was forgotten" clause. No AST guard was built
    (correctly declined). No "accepted gap against AD-14" is recorded anywhere.
11. **`folio-format.md` is amended as an OUTCOME, never a mechanism** (D-000.6) — it states what an
    author observes (width clips, height does not, image height is honoured), naming no function,
    file or code path.
12. **The API asymmetry is the owner's** — `Render` returns `(Result, error)`, `RenderTo` returns
    `([]Diagnostic, error)`. The implementation did **not** "helpfully" symmetrise them. Correct per
    D-2.8.6, and documented in both doc comments.
13. **The ~20 mechanically migrated test files carry no killed detector.** Every file's diff was read
    line by line; no `t.Fatal*`/`t.Error*` was deleted, no `break`/`continue` introduced, no error
    path turned into `_`, and every `renderto_test.go` error-path assertion (nil writer, mid-write
    failure, short write, render-error propagation, exact byte count) survives verbatim. The one
    theoretically-weakened control — `vendorboundary_test.go`'s `len(bad) != 0` — was probed by
    mutation (making `Render` return `Result{Bytes: []byte("LEAKED")}` on its error path) and
    **`TestFolioDeclinesEverySubstitutionAtIngestion` went red**, so the public-boundary property it
    guards is still live. **Not a finding.**

### Red-proofs executed (mutation → named test red → revert → SHA-256 re-verified)

| # | mutation | result |
|---|---|---|
| M2 | suppress the diagnostic attach (`if false && overflows`) | `TestWidthOverflowDetectedPerElement` red on **both** subtests — the real committed `wrapped-text`/e4 subject *and* the synthetic. FR44's acceptance test is genuinely red-proved, not merely described (D-000.52). Also exposed Major 3. |
| M3 | flip band concatenation to footer/content/header | `TestDiagnosticsOrderIsDocumentOrder` red. Document order is genuinely asserted — and the test's four elements across three bands really do distinguish band order from the Phase-A collection order. |
| M4 | `var diags []Diagnostic` → `diags := []Diagnostic{}` | `TestDiagnosticsEmptyIsNil` red. Empty-is-nil is genuinely asserted (`!= nil`, not `len() == 0`). |
| M5 | suppress the clip emission (`clipped := false && …`) | `TestClipRestrictsOnlyTheHorizontalAxis` **and** `TestWrappedTextGoldenFixture` red. |
| M6 | restore the old early-return defect (`err != nil \|\| len(res.Diagnostics) > 0`) | `TestRenderToWritesCompleteBytesDespiteWarning` **and** `TestRenderAndRenderToDiagnosticsAgree` red. The "no output at all for a document that rendered fine" breakage is genuinely closed and genuinely guarded (attack item 5). |
| M1 | change `DiagCodeTextClippedWidth`'s **value** | **suite stayed GREEN** → Blocker 1 |
| M8 | delete `Severity: SeverityWarning` from the construction | **suite stayed GREEN** → Major 2 |
| M7 | make `Render` leak bytes on its error path | `TestFolioDeclinesEverySubstitutionAtIngestion` red → dismissed as a non-finding (item 13 above) |

---

### Finding 1: The D-2.8.6 literal pin does not pin the diagnostic `Code`'s value — and the test's own comment claims it does

- **Severity**: **Blocker**
- **Category**: Tests
- **Location**: `folio-go/render_clip_diagnostic_test.go:368-381` (the `want` literal and its comment);
  `folio-go/diagnostic.go:87` (the constant); also
  `render_clip_diagnostic_test.go:118-120` and `:136-138`
- **Observation**: Proved by execution (M1). Changing the constant's **value** —
  `const DiagCodeTextClippedWidth = "TEXT_CLIPPED_WIDTH"` → `"MUTANT_CODE_XYZ"` — leaves the entire
  suite green (`go test ./... -count=1`: 600 PASS / 1 FAIL, unchanged). `grep -rn TEXT_CLIPPED_WIDTH
  folio-go/` returns exactly **one** source hit: the declaration itself. Every assertion in the tree
  compares `d.Code` against `DiagCodeTextClippedWidth`, i.e. the constant against itself, and the
  "pinned literal" does the same:

  ```go
  // Pinned literal (D-2.8.6): if BOTH producers drifted together —
  // e.g. the code constant were accidentally renamed — the equality
  // check above would still pass. This catches that.
  want := []Diagnostic{{
      Severity:  SeverityWarning,
      Code:      DiagCodeTextClippedWidth,   // <- the constant, not the wire string
      ElementID: "e1",
      Message:   res.Diagnostics[0].Message,
  }}
  ```

  The comment states the exact hazard the pin exists to catch, and the pin does not catch it. Only
  `ElementID` (`"e1"`) and the slice length are actually pinned against literals.
- **Impact**: AD-14 requires *"a stable string code from a **closed registry**"*, and `diagnostic.go`'s
  own comment says *"Additive only (AD-14, verbatim: 'changing a code's meaning is a breaking
  change'): once shipped, this string's meaning is permanent."* Nothing in the repository can detect
  a change to that wire value. `Code` is the field AD-14 designates as the *programmatic* identity —
  the one a CLI or designer dispatches on (`Message` is explicitly not parsed, D-000.21) — so this is
  the one field whose value most needs a pin. This is also the **first** code in the registry that
  Story 3.6 will extend with DW-6's `TABLE_FOOTER_SOURCE_UNRESOLVED` / `TABLE_FOOTER_SOURCE_FORBIDDEN`
  and AD-14's remaining codes: the pattern established here is the pattern that registry inherits.
  **The story's Delivery Log presents this as satisfying "D-2.8.6's explicit requirement"** (item 8:
  *"THEN pins `Render`'s side against a literal `[]Diagnostic` so the pair cannot both drift together
  undetected"*), and the boundary gate runs next and would inherit that assurance. A guard whose
  comment claims a property it does not have is the shape D-000.49/D-000.55 exist to prevent.
- **Suggested Resolution**: Pin the wire value with a bare string literal — either use `Code: "TEXT_CLIPPED_WIDTH"`
  in `want`, or add a one-line registry assertion (`if DiagCodeTextClippedWidth != "TEXT_CLIPPED_WIDTH"
  { t.Errorf(...) }`) sited so Story 3.6 will extend it rather than duplicate it. Do **not** resolve
  this by softening the comment. Red-proof the fix by re-running M1 and confirming the suite goes red.
- **Related AC**: AC7 (D-2.8.6's agreement/determinism requirement); AD-14

### Finding 2: `Diagnostic.Severity` is asserted three times and cannot fail any of them — `SeverityWarning` is the iota zero value

- **Severity**: **Major**
- **Category**: Tests
- **Location**: construction at `folio-go/render.go:532-533`; assertions at
  `folio-go/render_clip_diagnostic_test.go:113-115`, `:136-138`, and `:372` (the `want` literal);
  type at `folio-go/diagnostic.go:26-28`
- **Observation**: Proved by execution (M8). Deleting the line `Severity:  SeverityWarning,` from the
  `Diagnostic` literal in `collectBandTextRuns` leaves the **entire module suite green** (600 PASS /
  1 FAIL — only the pre-existing `TestCorpusMeetsP6ExerciseFloors`). The cause is structural:
  `SeverityWarning Severity = iota` makes `SeverityWarning == 0`, so an **unset** `Severity` field is
  byte-identical to a correctly set one. Every check — `d.Severity != SeverityWarning`,
  `Severity: SeverityWarning` inside the pinned literal — is satisfied by a field nobody assigned.
- **Impact**: `Severity` is one of AD-14's five ruled fields, and AC1 requires assertion **on the
  fields** rather than on "something was detected" (D-000.21). Three assertions read as coverage of
  it; none is falsifiable. The forward risk is concrete rather than theoretical: `SeverityError` is
  already declared and reserved for a future path, and AD-14's whole disposition rule ("Error aborts
  the render, Warning accompanies a successful one") turns on this field. The first code path that
  constructs an `Error`-severity `Diagnostic` and forgets to set the field will silently produce a
  `Warning` returned alongside bytes — the failure mode AD-14 was written to prevent — and nothing
  here would catch it.
- **Suggested Resolution**: Make the zero value non-conforming or assert against a non-zero value.
  Preferred: introduce an unexported `severityUnset Severity = iota` as the zero value so
  `SeverityWarning` becomes 1 and an unset field is detectable (`Severity.String()` already has a
  `default:` branch for exactly this). Alternatively, add a case that constructs and asserts a
  `SeverityError` value. Red-proof by re-running M8.
- **Related AC**: AC1, AC7; AD-14

### Finding 3: The agreement test panics instead of failing, and the panic blacks out the rest of the package's test run

- **Severity**: **Major**
- **Category**: Tests
- **Location**: `folio-go/render_clip_diagnostic_test.go:376` — `Message: res.Diagnostics[0].Message`
- **Observation**: Observed during M2. With the diagnostic attach suppressed,
  `TestRenderAndRenderToDiagnosticsAgree` indexes `res.Diagnostics[0]` with no length guard and
  produces:

  ```
  panic: runtime error: index out of range [0] with length 0 [recovered, repanicked]
  github.com/panitw/folio/folio-go.TestRenderAndRenderToDiagnosticsAgree(...)
      /Users/panitw/Projects/folio/folio-go/render_clip_diagnostic_test.go:376
  ```

  A panic aborts the whole test binary. Measured consequence in that run:
  `TestRenderToWritesCompleteBytesDespiteWarning`, `TestDiagnosticsOrderIsDocumentOrder` and
  `TestDiagnosticsEmptyIsNil` **never executed** — they appear in the `-v` output as neither `PASS`
  nor `FAIL`. Only two top-level failures were reported for a mutation that should have reddened at
  least four tests.
- **Impact**: The single most likely regression in this story's own subject — the diagnostic stops
  being attached — converts a clean, localised, well-messaged failure into a suite-wide blackout of
  the module's largest package, taking three of this story's own AC7 mechanisms down with it and
  reporting a stack trace instead of the carefully written failure messages. The boundary gate runs
  next; a gate reading "panic" cannot tell which properties still hold. Note the test one line above
  *does* guard correctly (`reflect.DeepEqual(res.Diagnostics, toDiags)` with a `t.Fatalf`) — the
  omission is only at the literal-pin step.
- **Suggested Resolution**: Add `if len(res.Diagnostics) != 1 { t.Fatalf("presence precondition: want
  exactly 1 Diagnostic, got %d", len(res.Diagnostics)) }` before constructing `want`. (Fixing Finding
  1 by using the bare string literal for `Code` removes one of the two reasons this line reaches into
  the slice, but the `Message` field still does — guard the length regardless.)
- **Related AC**: AC7

### Finding 4: The Delivery Log's new-test count is wrong — 13 claimed, 9 shipped

- **Severity**: **Minor**
- **Category**: AC Conformance
- **Location**: story Delivery Log, *"Test suite, final state"*; subject file
  `folio-go/render_clip_diagnostic_test.go`
- **Observation**: The story states *"13 new top-level test functions added
  (`render_clip_diagnostic_test.go`), all passing."* Measured: `grep -c '^func Test'
  render_clip_diagnostic_test.go` = **9** (`TestWidthOverflowDetectedPerElement`,
  `TestDetectWidthOverflowRecordsTheBoundAndExtent`, `TestOverflowingContentNeverReflowedNeverDropped`,
  `TestDeclaredHeightIsInert`, `TestClipRestrictsOnlyTheHorizontalAxis`,
  `TestRenderAndRenderToDiagnosticsAgree`, `TestRenderToWritesCompleteBytesDespiteWarning`,
  `TestDiagnosticsOrderIsDocumentOrder`, `TestDiagnosticsEmptyIsNil`). The measured top-level delta
  corroborates 9, not 13: **359 → 368**.
- **Impact**: AC5 requires every delta itemised and every enumeration to name its commit and method.
  A count that is wrong by four is exactly the kind of figure a later story inherits uncritically —
  the hazard D-2.6.9's correction and `[[carried-gate-figures-are-scoped]]` exist to stop — and it is
  the load-bearing input to Finding 5's justification.
- **Suggested Resolution**: Correct the number to 9 and name the method (`grep -c '^func Test'` on the
  one new file) alongside it.
- **Related AC**: AC5

### Finding 5: AC5's top-level count is omitted, and the reason given for omitting it is falsified by measurement

- **Severity**: **Minor**
- **Category**: AC Conformance
- **Location**: story Delivery Log, *"Test suite, final state"*
- **Observation**: The Delivery Log reports only *"600 passed, 1 failed (all-occurrences; **top-level
  not separately re-quoted here as it moves 1:1 with the new top-level test functions added**)"*. AC5
  and the story's own scope rule require both figures (*"all-occurrences and top-level reported
  separately"*; a figure without its scope *"must not be written down"*). The stated justification is
  also arithmetically false on its own inputs: the baseline is 359 top-level and the story claims 13
  new functions, which predicts **372**; the measured value is **368**. (It moves 1:1 with the *real*
  count of 9 — so the reasoning is sound and only the input is wrong, but the omission means the
  discrepancy was never surfaced.)
- **Impact**: The boundary gate is the next consumer of these figures and D-000.3 makes it the
  owner's observation point. Handing it one of the two required numbers, with a stated reason that
  does not survive arithmetic, is precisely the unscoped-figure hazard the story itself quotes.
- **Suggested Resolution**: Record both, scoped: **600 PASS / 1 FAIL all-occurrences, 368 PASS / 1
  FAIL top-level**, whole module, no build tags, `-count=1`, at this story's own tree atop `278520b`,
  with the single expected failure `TestCorpusMeetsP6ExerciseFloors` at its recorded stats. These are
  measured and reproduced in this review.
- **Related AC**: AC5

### Finding 6: `TestRenderAcceptsATemplate`'s doc comment states a signature that has not existed for two stories

- **Severity**: **Nit**
- **Category**: Maintainability
- **Location**: `folio-go/template_test.go`, `TestRenderAcceptsATemplate`'s doc comment
- **Observation**: The comment still reads *"AC22: Render's signature is `func Render(t *Template, d
  Data, f FontSet) ([]byte, error)`"* — wrong on the `Params` argument (stale since Story 1.7) and now
  also on the return type. The migration updated the call site beneath it and not the comment.
- **Impact**: Cosmetic; no assertion is affected. Recorded because this project has repeatedly found
  stale comments hardening into inherited claims (D-2.7.4's process note; Story 2.7 review Finding 6),
  and this one names a public signature.
- **Suggested Resolution**: Update the comment to `func Render(t *Template, d Data, p Params, f
  FontSet) (Result, error)`.
- **Related AC**: —

---

### AC-by-AC disposition

| AC | disposition | evidence |
|---|---|---|
| **AC1** — per-element horizontal detection, identified by element id | **satisfied** | Red-proved by M2 on both required subjects (the committed `wrapped-text`/e4 *and* the synthetic Latin atomic token — D-000.50's two-subject requirement honoured). Fields asserted individually, not "something was detected". `detectWidthOverflow`'s own unit test covers the `<=` boundary (a line exactly as wide as the box does not overflow) and the `boxWidth == 0` no-declared-width case. **Caveat: two of the five AD-14 fields carry unfalsifiable assertions — Findings 1 and 2.** |
| **AC2** — never reflowed, never dropped, siblings unmoved | **satisfied** | Both halves present, as finding 2 of the story required: the glyph-survival half is paired with a wide-box negative control of the *identical* word, and the sibling's origin and CIDs are asserted unchanged. Read back off produced PDF bytes, not renderer internals. |
| **AC3** — declared `height` is inert | **satisfied** | `TestDeclaredHeightIsInert` renders pairs differing only in height at three extreme deltas in both directions (24/999999, 24/1, 1000/5) across a single-line (`shaped-text` geometry, 16pt/height 24) and a multi-line subject, asserting both byte equality **and** `Diagnostics` equality. Keyed on the purpose, not the `el.Height` proxy. Genuine forward guard. |
| **AC4** — nine pre-existing goldens byte-identical, `wrapped-text` moves | **satisfied** | Independently confirmed three ways: `git diff --stat 278520b -- fixtures/` (only `wrapped-text` touched), direct `shasum -a 256` of the three pinned digests, and all four matrix legs agreeing on all nine documents. |
| **AC5** — baseline preserved, deltas itemised, commits named | **partially satisfied** | Baseline preserved and the expected failure reproduced exactly; commits and methods named throughout (the call-site re-derivation is exemplary). **Findings 4 and 5**: the new-test count is wrong and the required top-level figure is omitted on a falsified justification. |
| **AC6** — the clip | **satisfied** | Red-proved by M5. Clip rectangle verified against the shipped golden's actual bytes (`36 0 20 841.89 re / W n`), horizontal-only, vertical extent the full page. `q`/`Q` balance proved structurally and measured on both the golden (2/2) and a 3-page probe (4/4). |
| **AC7** — the diagnostic channel | **partially satisfied** | Document order (M3), empty-is-nil (M4) and the `RenderTo` complete-bytes-despite-warning fix (M6) are all genuinely red-proved, and `Render`/`RenderTo` agreement is asserted via `reflect.DeepEqual`. **Findings 1, 2 and 3** all land here: the literal pin does not pin `Code`'s value, `Severity` is unfalsifiable, and the pin step panics rather than failing. |

### Notes for the finisher

- Findings 1, 2 and 3 are all inside `render_clip_diagnostic_test.go` and are cheap to fix together;
  none requires touching production behaviour, which is correct as shipped. **No golden moves as a
  result of any fix below**, so `fixtures/wrapped-text/expected.pdf` must remain
  `07c38cf765a39d86376c1a3c78bfb6f0a96f089f19792c9bfeeaa1dc754269d6` / 72,790 bytes and the four
  matrix legs need not be re-run.
- After fixing, re-run M1 and M8 (mutate the constant's value; delete the `Severity:` line) and
  confirm each now goes **red** — a fix to a vacuity finding is not done until the mutation reddens.
- Findings 4 and 5 are edits to the story's own Delivery Log; the corrected figures are measured and
  quoted above and can be used verbatim.
- Nothing in this review requires reopening D-2.8.1–D-2.8.6. The `Render`/`RenderTo` asymmetry, the
  vertical-axis fence, the six-obligation gate set, the declined AST guard and the `folio` package
  placement of `Diagnostic`/`Result` were each checked against their ruling and are all correct.

---

## Finding Resolutions (finisher)

All work below is confined to `folio-go/render_clip_diagnostic_test.go`, `folio-go/template_test.go`
(a doc comment) and this story's own Delivery Log/Dev Agent Record — **no production code changed,
no golden moved**. `fixtures/wrapped-text/expected.pdf` reproduced at
`07c38cf765a39d86376c1a3c78bfb6f0a96f089f19792c9bfeeaa1dc754269d6` / 72,790 bytes, unchanged; the
four matrix legs were not re-run because nothing that could move them changed (spot-checked anyway,
see Validation below).

### Finding 1 (Blocker) — the pin that pins nothing: **FIX**

`want`'s `Code` field now compares against the bare wire literal `"TEXT_CLIPPED_WIDTH"` instead of
the `DiagCodeTextClippedWidth` constant, so a changed constant value and a changed comparison value
can no longer drift together undetected. The comment above it was rewritten to state accurately
what the pin does and doesn't catch (it no longer claims a guarantee the prior version lacked).
**Red-proved**: re-ran M1 (`DiagCodeTextClippedWidth`'s value changed to `"MUTANT_CODE_XYZ"`) —
`TestRenderAndRenderToDiagnosticsAgree` now fails, quoting the mismatched `Code`. Reverted by hand;
`diagnostic.go`'s SHA-256 confirmed identical to its pre-mutation value
(`d85e86cffc3e622b0def986570887483567e37b280862025489e0551df1fbb4e`).

### Finding 2 (Major) — `Severity` cannot fail its assertions: **DEFERRED, not fixed** (filed as DW-18)

This one is not closed by a test-only change, and I'm not making it one. `SeverityWarning` is
`iota`'s zero value, which is also Go's zero value for an unset field — the mutation (deleting the
`Severity: SeverityWarning,` assignment) produces a runtime value that is bit-for-bit identical to
the correctly-set one. No comparison written in the test file can tell those two states apart;
closing the gap requires either changing `Severity`'s zero-value semantics (an unexported sentinel
ahead of `SeverityWarning` in the `iota` sequence, so the constant's numeric value shifts) or an
equivalent change to the public `Severity` type — both are changes to a public, AD-14-governed
surface, which is exactly the call this review itself flags as "AD-14 territory" rather than a
finisher's to make unilaterally while closing out a test-quality pass.

**What was done instead**: a comment was added at each of the three assertion sites (the two
per-subject checks in `TestWidthOverflowDetectedPerElement` and the pinned literal in
`TestRenderAndRenderToDiagnosticsAgree`) recording the limitation in place, so a future reader sees
documented intent rather than an implied guarantee. **Filed as DW-18** in `deferred-work.md`, owned
by whoever next touches `folio.Severity` — plausibly Story 3.6, the first story to construct a real
`SeverityError` value. **Re-ran M8** (deleting the assignment) after this pass to confirm the
honest state: the suite **stays green**, exactly as before — this is not a regression, it is the
structural fact the deferral records. Reverted by hand; `render.go`'s SHA-256 confirmed identical
(`4835bc0f8ef9faad9dd35919750a8f48ed2b67f8289547f6f27877f9b0ec8736`).

### Finding 3 (Major) — the unguarded index that hides three tests: **FIX**

Added a presence-precondition length guard (`if len(res.Diagnostics) != 1 { t.Fatalf(...) }`)
immediately before `TestRenderAndRenderToDiagnosticsAgree` builds its `want` literal, so a
suppressed attach fails that one test cleanly instead of panicking and blacking out the rest of the
package's run. Checked the rest of `render_clip_diagnostic_test.go` for the same pattern (per the
review's own instruction): the only other two `res.Diagnostics[0]` indexing sites, in
`TestWidthOverflowDetectedPerElement`'s two subtests, were already guarded by an equivalent
`len(res.Diagnostics) != 1` check immediately above them — not a second instance of the defect.
**Red-proved**: re-ran M2 (suppress the attach) with the fix in place —
`TestRenderAndRenderToDiagnosticsAgree` now fails cleanly with the precondition message (no panic,
no stack trace), and `TestRenderToWritesCompleteBytesDespiteWarning`,
`TestDiagnosticsOrderIsDocumentOrder` and `TestDiagnosticsEmptyIsNil` all ran to completion and
reported their own PASS/FAIL (two correctly failed under the mutation, one correctly passed) —
confirming the blackout is closed. Reverted by hand; `render.go`'s SHA-256 confirmed identical.

### Finding 4 (Minor) — Delivery Log's new-test count wrong (13 vs 9): **FIX**

Delivery Log corrected to **9**, with the method named (`grep -c '^func Test'
render_clip_diagnostic_test.go`), and the top-level delta cross-checked (359 baseline + 9 = 368
measured — arithmetically consistent, unlike the original 359 + 13 = 372 ≠ 368).

### Finding 5 (Minor) — AC5's top-level count omitted on a false justification: **FIX**

Delivery Log now states both required figures, scoped: **600 PASS / 1 FAIL all-occurrences, 368
PASS / 1 FAIL top-level**, whole module, no build tags, `-count=1`, `go test ./... -count=1 -v`, at
this story's own tree, with `TestCorpusMeetsP6ExerciseFloors` at its recorded stats.

### Finding 6 (Nit) — stale signature in a doc comment: **FIX**

`folio-go/template_test.go`'s `TestRenderAcceptsATemplate` doc comment updated to
`func Render(t *Template, d Data, p Params, f FontSet) (Result, error)`.

### Validation performed by the finisher

- `go build ./...` (folio-go module): clean.
- `go test ./... -count=1 -v` (folio-go module): **600 PASS / 1 FAIL all-occurrences, 368 PASS / 1
  FAIL top-level** — matches the review's own measured baseline and the epic gate's stated
  expectation exactly. Sole failure `TestCorpusMeetsP6ExerciseFloors`, stats
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`, reproduced character-for-character.
- `go vet ./...` and `go vet -tags matrix ./...` (folio-go module): both clean.
- `go test ./... -count=1` under `lint/` (sibling module): all four packages green.
- `go test -tags=matrix -run TestCrossTargetByteIdentity -count=1 -v .`: **PASS**, all four targets
  agree on all nine documents, including `wrapped-text` at
  `07c38cf765a39d86376c1a3c78bfb6f0a96f089f19792c9bfeeaa1dc754269d6` / 72,790 bytes.
- Pinned digests re-measured directly: `shaped-text` = `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370`,
  `expected-breaks` = `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de`,
  `multi-page` = `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b` — all three unmoved.
  `git diff --stat 278520b -- fixtures/` confirms `wrapped-text` is the only fixture touched.
- Each red-proof (M1, M8, M2) was executed, observed, and reverted by hand with a SHA-256
  before/after comparison, per D-000.30/D-000.40, as itemised above.

### Summary

**1 Blocker fixed, 2 Majors (1 fixed, 1 deferred with a filed follow-up DW-18), 2 Minors fixed, 1
Nit fixed.** No production code changed; no golden moved; `epic-2: backlog` untouched;
`declaredEpic2GateObligations` untouched (not read by this pass). Story set to **done**.
