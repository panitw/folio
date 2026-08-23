---
title: Folio PRD — Adversarial Review
status: draft
created: 2026-08-22
reviewer: adversarial pass
target: prd.md (MVP v0.1) + addendum.md, against docs/folio-mvp-plan.md
---

# Adversarial Review — Folio PRD (MVP v0.1)

## Verdict

This is a well-written document about a project that cannot be built as specified. The prose
quality is actively dangerous here: it reads like a finished plan, so nobody will notice that
the two decisions that determine the entire codebase — **what the layout model is** and **what
the expression language is** — are never made, and that the document's signature requirement
(NFR1, cross-target byte equality) was invented by the PRD, is not in the source plan, is not
in the assumption register, and is simultaneously listed as an open feasibility question (Q1).

The PRD also silently reverses the source plan on the single largest scope item in the project.
`docs/folio-mvp-plan.md` §14 puts **Internationalization** explicitly out of scope. The PRD makes
Thai and CJK text a hard Must (NFR3) and does not acknowledge the reversal anywhere. That one
edit adds an OpenType shaping engine, a CID font subsetter, and a Thai dictionary breaker — in
pure Go, with no cgo permitted — to a solo project. Nothing in §5, §6, §7, §8, or §9 accounts for it.

Underneath that: the MVP is four products (a typesetting engine, a deterministic PDF writer, a
programming language, and a visual design tool). One person does not ship that. The success
criteria in §7.2 can all pass on a renderer that handles exactly one template — I show how below.
And §7.1's golden report contains a **generated date** in its footer, which makes S2 and S3
(hash equality) literally unsatisfiable as written.

Fix the two undecided models, register the requirement you invented, cut the scope you smuggled
in, and rewrite the success measures so a demo cannot pass them. Then this is a real plan.

---

# Findings

Ranked by severity, most severe first.

---

## F1 — CRITICAL — The layout model is never chosen, and every other FR depends on it

**Location:** FR1, FR5, FR6, FR21, FR24, FR29, FR33; addendum §C (`"body": [...]`)

FR1 and FR5 describe drag-and-drop placement with **X/Y position and width/height** — absolute
positioning. FR6 describes three regions. FR21/FR24 describe a table that **grows dynamically and
breaks across pages** — flow layout. The addendum's template shape is an **ordered array**
(`body: [...]`), which reads as flow. The PRD contains both models and commits to neither.

Answer this question and the whole build changes: *when a table bound to `transactions[]` grows
from 3 rows to 300, what happens to the Text component positioned below it?*

- Absolute model → it stays at Y=400 and the table renders straight through it. Product is broken.
- Flow model → it pushes down, and FR5's X/Y properties are a lie, and FR1's "resize directly" is
  a lie, and the design canvas is not a canvas.
- Hybrid (absolute inside bands, flow between them) → that is a real design, and it is not in
  this document.

**Why it matters:** two engineers handed this PRD build two incompatible products, and neither is
wrong. This is not a detail for architecture to settle — it determines what the designer *is*, what
`.folio` *is*, and whether the golden report is even expressible. It is also, damningly, **absent
from §8 Open Questions** — the list of things that block downstream work does not contain the thing
that blocks everything.

**Do:** Add an FR that names the model explicitly. Recommended: page-header band (absolute,
fixed height) + content region (vertical flow of blocks; a block may internally position children
absolutely within its own fixed height) + page-footer band (absolute, fixed height). Then delete
X/Y from the property list for content-region blocks, or scope it to "within the parent block."
Add Q0 to §8 and mark it blocking until answered.

---

## F2 — CRITICAL — NFR1 is unsourced, unregistered, and is simultaneously an open feasibility question

**Location:** NFR1; Q1; §9 Assumption Register; addendum §E

The source plan (§10) says: *"The same template and data should generate the same layout in every
environment."* **Layout.** The addendum records this honestly: *"Stated at layout granularity
(`same layout`), never as byte-for-byte PDF equality."*

The PRD escalates this to: byte-identical PDF, same SHA-256, **across compilation targets** —
WASM in a browser must hash-match native on Linux — and calls it *"the signature requirement"* and
*"the single hardest constraint in the MVP."* It then admits in Q1 that nobody knows if it is
achievable and that it *"may rule out most off-the-shelf PDF libraries."*

So the hardest requirement in the document is (a) invented by the PRD, (b) not tagged
`[ASSUMPTION]` while far softer claims are, (c) absent from the §9 register, and (d) a listed
open question about its own feasibility. A requirement whose feasibility is unknown is an
assumption wearing a MUST.

**Why it matters:** if Q1 comes back "you must hand-roll the PDF writer including deterministic
TrueType/CFF subsetting," you have added 3–6 months of solo work to produce a file format, not a
feature. If Q1 comes back "cross-target byte equality is not reachable in budget," then NFR1,
NFR2's entire justification, NFR5, FR34, S2, S3, S5, and C3 collapse together — that is the
product's positioning, gone. This one question invalidates more of the document than any other.
It is not deferrable; it is the go/no-go.

**Do:** (1) Register it in §9. (2) Split it: **NFR1a** byte-identical on a single target (cheap,
testable, real value for CI); **NFR1b** byte-identical across targets (expensive, gates FR34).
Ship v0.1 on NFR1a. (3) Spike Q1 *before* any epics are written — a two-week prototype that emits
a deterministic PDF with an embedded subset font from both native and `GOOS=js` and diffs the
bytes. If it fails, you have saved the project.

---

## F3 — CRITICAL — NFR3 requires a text shaping engine that appears nowhere in the document, and reverses the source plan's scope

**Location:** NFR3, NFR7, FR31, Q2, §4.2, §10; source plan §14

The source plan puts **Internationalization** in §14 "Explicitly Out of Scope for MVP." The PRD
makes Latin + CJK + **Thai** a Must and never mentions that it reversed the source. §4.2's
out-of-scope list quietly drops "internationalization" and substitutes "right-to-left text,"
which is a much smaller concession.

Worse, NFR3 frames the problem as *measuring, breaking, and rendering* — as if it is a line-break
algorithm. It is not. Thai requires **glyph shaping**: combining vowels and tone marks stack above
and below the base consonant and must be repositioned (`GSUB`/`GPOS` mark attachment) or the
output is visibly wrong — collided marks on descenders, wrong tone-mark height. CJK requires
**Type0/CID composite fonts**, `Identity-H` encoding, CMap and `ToUnicode` generation, and
subsetting across a 20,000+ glyph face.

The word "shaping" does not appear in the PRD or the addendum. There is no FR for it. And the
standard solution — HarfBuzz — is **C**, which NFR1 forbids ("no cgo in the rendering path"). So
NFR3 + NFR1 together require: *write an OpenType shaping engine in pure Go.*

Add NFR7's admission that the font set, custom-font supply path, licensing, and missing-glyph
fallback are **all unspecified**, and FR31's demand that all fonts be embedded, and Q2's note that
"Thai + CJK coverage with deterministic subsetting is non-trivial." Non-trivial is doing a lot of
work in that sentence.

**Why it matters:** this is the largest single body of hidden work in the PRD, it is concealed
inside one NFR bullet, and it is load-bearing for NFR1, FR31, S2, S3, S4, and the golden report
(which is a Thai bank statement per UJ-1). Getting it wrong does not degrade the product — it
means Thai output is unshippable to a Thai bank, which is the author's own itch.

**Do:** Cut CJK from v0.1. Keep Latin + Thai (that is the actual use case) and add explicit FRs:
"shape Thai text using embedded OpenType GSUB/GPOS tables," "subset and embed a simple TrueType
font deterministically." Decide Q2 *now* — pick one Thai-covering face with a permissive licence
(e.g. an SIL OFL Thai face) and freeze it. Note the Thai break dictionary's licence: the obvious
sources (ICU, libthai) carry terms that interact with Q7, and no one has checked.

---

## F4 — CRITICAL — There is no expression language here, only a list of eight function names

**Location:** FR16, FR17, FR18, FR19; addendum §D; Q3

FR17 lists functions. It does not specify a grammar, and the omissions are fatal:

- **No operators.** `if()` is listed. `if(x, a, b)` requires a boolean `x`. There is no `>`, `<`,
  `==`, `&&`, `!`. As specified, `if()` can only branch on truthiness of a bare path, which makes
  it nearly useless and makes FR19 (conditional visibility) unimplementable for any real condition
  like "hide when balance is zero."
- **No literals.** `formatDate(d, "dd/MM/yyyy")` implies string literals. Numbers? Booleans? Null?
- **No type system and no null/missing semantics.** What is `{{customer.middleName}}` when absent —
  empty string, the literal text, or an FR40 error? What is `sum()` over an empty array — `0` or an
  error? What is `1 + "x"`? These are not edge cases; a 50-page statement will hit all of them.
- **No row scope (FR16, Q3).** The PRD admits this: the plan writes `{{transaction.amount}}`
  (singular) with no scoping rule ever named. Is it the singularized collection name? An implicit
  `row`? A `this`? Positional?
- **Two incompatible syntaxes for the same array.** FR15 binds a region with `transactions[]`.
  FR18 aggregates with `sum(transactions.amount)` — a dotted path *through* an array with no `[]`
  and an implicit map-over-collection that is never defined as a language feature.

**Why it matters:** this is the textbook "two engineers implement incompatibly" case. One builds
a Jinja-ish evaluator with `.` traversal and truthiness; the other builds a typed mini-language
with explicit `row`. Both satisfy FR17. Neither can read the other's templates — and the `.folio`
file is supposedly "the contract" between two users (§3).

**Do:** Write the grammar into the PRD or the addendum before a single story. EBNF, ~30 lines.
Name the row scope. Name the null policy. Name the arity and types of all eight functions. Cut
`if()` from v0.1 if you do not want to specify an operator set — and then cut FR19 with it, because
it depends on `if()`.

---

## F5 — CRITICAL — §7.1's golden report contains a generated date; §7.2's S2/S3 require hash equality. Pick one.

**Location:** §7.1 (footer: "confidentiality text, **generated date**, and Page X of Y") vs S2, S3, NFR1

The primary acceptance artifact has a generation timestamp printed in its footer. The primary
success measures demand that two renders of it produce identical SHA-256 hashes. Two renders one
minute apart cannot both be byte-identical and carry a real generated date.

NFR1's list of variation sources to pin — *"document creation timestamp, document ID, internal
object ordering, compression output, font subset generation"* — covers PDF **metadata**
timestamps. It does not cover a timestamp that is *page content*.

**Why it matters:** it is a direct, mechanical contradiction between the acceptance report and its
own pass criteria, sitting two paragraphs apart. It also exposes a real determinism hazard nobody
listed: any expression whose value depends on wall-clock or the host timezone (`formatDate` on a
"now" value) breaks NFR1 silently, on some machines, in some months.

**Do:** State that the generated date is supplied as a **parameter** (`{{params.generatedAt}}`),
that `folio-go` has no access to the clock, and that `formatDate` requires an explicit timezone
argument or renders in UTC. Add "wall-clock and host timezone" to NFR1's pinning list.

---

## F6 — CRITICAL — All six success measures pass on a renderer that handles exactly one template

**Location:** §7.2 (S1–S6), §7.1

Here is the exploit, step by step. It requires no bad faith — it is the path of least resistance
for a solo builder under pressure.

- **S1** *(golden report renders at 1/5/20/50 pages)* — every measure in §7.2 is scoped to "the
  golden report." Hardcode the five column widths. Tune the pagination against this one table.
  Ship a renderer that has never seen a second template.
- **S2 / S3** *(hash equality, native-vs-WASM and macOS-vs-Linux)* — hash equality is a property
  of **determinism, not correctness**. A renderer that emits a blank A4 page passes S2 and S3
  perfectly. So does one that renders Thai with every tone mark in the wrong place, as long as it
  does so consistently. These two measures — half the scorecard — cannot detect a wrong document.
- **S4** *(Thai and CJK wrap at correct boundaries)* — **"Verified by inspection."** By whom? The
  solo author, who also wrote the line breaker, inspecting output against no reference. There is no
  corpus, no comparison against ICU/libthai, no count of test strings, no definition of "correct."
  This is not a measure; it is a mood.
- **S5** *(round trip — design in browser, save locally, render)* — satisfied by: open the golden
  `.folio`, click Save, render. **Zero designer capability is exercised.** It does not require
  authoring anything, binding anything, or placing a component.
- **S6** *(50-page render streams without materialising the whole document)* — **"Verified."** No
  memory ceiling, no threshold, no instrument. NFR4 explicitly disclaims having any target. There
  is no observable pass/fail. See F8.

**Nothing in §7.2 measures:** a second template · a template authored from scratch in the designer
· the expression language (FR17) · parameters (FR20) · conditional visibility (FR19) · the error
contract (FR40) · standalone validation (FR41) · column alignment (FR22) · page setup (FR2) ·
anyone other than the author successfully producing a report.

**Why it matters:** you can score 6/6 and hold a demo, not a product. And because §7 declares
"success is defined by capability and reliability, not adoption," there is no downstream signal
that will ever correct the error.

**Do:** Add S7: *a second, structurally different report (multi-column invoice with conditional
blocks) renders correctly from a template the golden report's code path has never seen.* Add S8:
*every FR40 error class is triggered by a fixture and produces an error naming the element and the
path.* Rewrite S4 against a fixture corpus with expected break positions. Rewrite S5 to require
authoring the golden template from an empty canvas. Give S6 a number.

---

## F7 — CRITICAL — This is four products. One person does not ship it.

**Location:** §5 in aggregate; addendum §H; §10 ("solo project, sequencing over dates")

Honest enumeration of what §5 + §6 actually require, built from zero, in pure Go with no cgo:

1. **A PDF writer** — object model, xref, deterministic compression, TrueType/CFF **subsetting**,
   Type0/CID composite fonts, CMap + ToUnicode, image XObjects. (NFR1 likely forbids using an
   existing one — Q1.)
2. **A typesetting engine** — text measurement from raw font tables, **OpenType shaping** for
   Thai, UAX#14 line breaking for CJK, dictionary breaking for Thai, a layout model (F1), table
   pagination with repeated headers and footer aggregates.
3. **A programming language** — lexer, parser, evaluator, type/null semantics, eight functions
   (F4).
4. **A visual design tool** — canvas, drag/resize/snap, a properties panel, a *table* editor
   (columns, widths, alignments, header toggle, aggregate footers — see F11), JSON-path discovery
   and binding UI, local file open/save, plus a **WASM build with an embedded font and dictionary**
   for FR34.

Plus a cross-platform, cross-target determinism harness in CI (S2, S3).

Any one of these is a credible solo year. The build order in addendum §H compresses item 4 — the
largest and least specified — into a single line, "Phase 5 — Designer." §10 says "sequencing over
dates," which is a graceful way of declining to notice that the sequence has no end.

**Why it matters:** the failure mode is not "late." It is "Phase 1–3 take 14 months, the designer
never starts, and the two-user premise in §3 was never tested." C4 ("designer work starting before
the engine renders the golden report") actively *encourages* this failure by making early designer
work a counter-metric.

**Do:** Take the cut list in **§ The v0.1 That Can Actually Ship** below. If you refuse to cut
cross-target byte equality (F2), then cut the designer from v0.1 entirely and ship `folio-go` with
hand-written templates — that is an honest fork. You cannot have both the hardest rendering
constraint in the industry and a visual design tool in the same solo release.

---

## F8 — HIGH — FR30 (`Page X of Y`) and NFR4 (streaming) are mutually exclusive as written, and S6 cannot be failed

**Location:** FR30, NFR4, FR38, S6; addendum §J5

You cannot print "Page 1 of 34" in the footer of page 1 until you know there are 34 pages, and you
do not know that until you have laid out all 34. NFR4 says large reports "must stream to a writer
**rather than materialising fully in memory**." Those cannot both hold naively.

They are reconcilable — lay out fully into the page model, then *stream the write* — but that
means NFR4's actual guarantee is "the PDF byte stream is not buffered," not "the document is not
materialised." The page model for 50 pages *is* materialised, in memory, always. The PRD says the
opposite.

And S6 — "50-page render streams without materialising the whole document / **Verified**" — has no
threshold, no instrument, and NFR4 explicitly states that no memory ceiling, throughput target, or
row-count limit exists anywhere. **There is no way to fail S6.** A requirement that cannot be
failed is decoration.

Compounding: NFR4 is marked *"important, below the Must-level requirements"* while FR38 (render to
a writer) is an unqualified FR and S6 is a top-line success measure. The priority is inconsistent
across three places.

**Do:** Rewrite NFR4 to what is true and testable: *"the PDF byte stream is written incrementally
to the `io.Writer`; peak RSS for a 50-page / 5,000-row statement stays under N MB."* Pick N (256MB
is defensible). Rewrite S6 to measure that number. State in FR30 that total-page resolution
requires full layout before write, so the architect stops treating it as an open trade (Q4).

---

## F9 — HIGH — FR39 requires parameters at render time; the API in the addendum has nowhere to put them

**Location:** FR39, FR20 vs addendum §B

```go
folio.Render(template Template, data any) ([]byte, error)
folio.RenderTo(w io.Writer, template Template, data any) error
```

FR20 requires a distinct `params.*` namespace. FR39 requires parameters "at render time alongside
the data." Neither signature accepts them. The source plan has the same hole (§7 defines parameters,
§4 shows an API without them), and the addendum faithfully reproduces the hole instead of flagging it.

**Why it matters:** this is the exact seam between the two users of §3, and it is broken in the one
concrete artifact a Go developer will copy-paste. It also forces an early, permanent API decision:
a third positional argument, an options struct, or variadic functional options — and F5's fix
(generated date as a parameter) makes it mandatory, not optional.

**Do:** Fix the signature in the addendum now: `folio.Render(tmpl Template, in Input) ([]byte, error)`
with `Input{Data any; Params map[string]any}`. Also settle the module path — addendum §B records
that `github.com/folio-reports/folio` vs `.../folio-go` is **undecided**, which means the import
line in every doc and example is currently wrong 50% of the time.

---

## F10 — HIGH — Text overflow policy is undefined, and FR40 makes the common case an error

**Location:** FR5, FR40, FR23, FR33

A Text component has a fixed width and height (FR5). A bound value is longer than fits. What
happens? Clip? Shrink-to-fit? Grow the box? Grow the row? Push the next element? Drop it? Throw?

The PRD's only statement is FR40's error class: *"content that cannot be laid out."* In enterprise
reporting, content that does not fit its box is the **single most common runtime condition** —
long customer names, long transaction descriptions, Thai text at ~1.4x the width of the Latin
mockup. If overflow is an error, the library fails on real data. If it silently clips, the "professional
statement" claim is dead the first time a name is truncated mid-glyph.

Table cells inherit the same hole: FR23 requires borders and padding, but nothing says how row
height is computed when a cell wraps to three lines, or whether all cells in a row share a height,
or what happens when a single row is taller than a page (FR24 says "break a table across pages
correctly" — a row taller than a page is the case that defines "correctly").

**Do:** Add an FR: per-component overflow policy from `{clip, wrap, grow}` with a stated default;
table rows size to the tallest cell; a row taller than the content region is an FR40 error naming
the row index. This is a five-line addition that removes a month of rework.

---

## F11 — HIGH — The designer is co-primary in §3 and an afterthought in every other section

**Location:** §3 vs addendum §H (Phase 5/6), §7.2 (S5), C4, Q6, FR5, UJ-1

§3 declares "two primary users of **equal standing**" and says the relationship "is the product."
Now audit what the document actually does for the template author:

- **Build order** puts the Designer at Phase 5 of 6 and Preview at Phase 6 — dead last, after the
  four hardest engine phases.
- **C4** makes starting designer work early a *counter-metric* — i.e. serving user #1 first is
  officially defined as failure.
- **Q6** asks "does the designer really come last, when it is success criterion #1 and half the
  product's users?" — and then does not answer. The PRD raises the objection to its own premise
  and files it as a question.
- **§7.2** gives the designer exactly one weak measure (S5, which F6 shows is passable by opening
  and re-saving a file).
- **§7.1**, the acceptance test, is a **PDF**. It can be produced entirely by hand-writing JSON.
  The designer is not required to exist for the primary acceptance test to pass.
- **FR5** lists a flat property set (X/Y, W/H, font, bold, align, border, padding, background,
  visibility, binding). **None of it edits a table.** UJ-1 has Ploy "set five columns with fixed
  widths" and "turn on a repeated header row and a sum footer" — there is no FR for a column
  editor, per-column width/alignment, a header-repeat toggle, or an aggregate-footer picker. The
  designer's most complex UI surface is entirely unrequirement'd, hidden behind FR4's word "Table."
- **UJ-2** has Anan compare his hash "against the file Ploy saw in her browser" — which requires
  Ploy to **save the previewed PDF out of the designer**. No FR provides that.

**Why it matters:** §3 is the document's thesis, and the FR set does not support it. This is not a
prioritisation choice — it is an unacknowledged one. Say it out loud or fix it.

**Do:** Either (a) demote the template author to secondary for v0.1 and rewrite §3 honestly, or
(b) reorder: make the designer's *file round-trip* Phase 2, so the `.folio` contract is exercised
by both sides from month two. Add FRs for the table property editor and for exporting the preview
PDF. Delete C4 or rewrite it as "engine capability trailing designer capability," which is the
real risk.

---

## F12 — HIGH — FR32 (embed images in the template) destroys FR11 (readable, diffable, mergeable)

**Location:** FR11, FR10, FR32; §2 positioning table

FR10/FR11 sell the `.folio` file as "a plain text file the user owns — readable, diffable,
versionable in Git, editable by an AI." FR32 requires images to be embedded **inside the template**.
A bank logo is 50–500KB; base64'd into a JSON string it is a single 700,000-character line.

That file is not human-readable. It is not meaningfully diffable — every logo change is one
enormous line change. It is not mergeable. It will blow past the context window of the AI agent
FR11 promises can edit it. And JSON has no comments, which further undercuts "human-readable."

Separately, "mergeable" is doing unearned work: `body` is a positional array (addendum §C). Two
authors reordering elements produce a conflict Git cannot resolve semantically. FR11 has **no
observable pass/fail** — how does one test "readable" or "AI-editable"? It is a marketing claim
in the FR list.

**Do:** Either allow a sidecar assets directory / `.folio` bundle (a directory or zip with the
template as plain JSON plus binary assets) — which keeps determinism as long as the bundle is
self-contained, satisfying FR32's actual goal (no network, no filesystem lookup at render time) —
or accept that FR11 applies to the layout, not the assets, and say so. Make FR11 testable: "the
template round-trips through the designer with stable key order and stable formatting; a
single-property change produces a diff of ≤3 lines."

---

## F13 — HIGH — The PRD forbids itself from specifying mechanisms, then mandates WebAssembly

**Location:** FR34, NFR2, FR35, Q9; addendum preamble ("the PRD body states capabilities, not mechanisms"); source §11

The addendum opens by declaring the division of labour: the PRD states capabilities, the addendum
carries mechanisms, architecture decides. Then FR34 mandates **"the real `folio-go` engine compiled
to WebAssembly and running in the browser tab."** That is a mechanism, an architecture decision,
and the most expensive one in the document — made in the PRD, by fiat, in violation of the PRD's
own stated rule.

The source plan does not say this. Source §11 shows `Designer → Preview API → folio-go → PDF` — a
**service**. The addendum flags this honestly (§G: "Its nature — service, local process, or
in-designer call — is unstated"; §J2: "Preview API is unscoped work"). The PRD resolved that
ambiguity toward the single most expensive option and did not mark it as a decision.

And Q9 quietly notes the cost: *"What is the WASM bundle size ceiling, given embedded fonts and a
Thai breaking dictionary? — may be significant."* Do the arithmetic the PRD declines to: a Go WASM
binary (~2–10MB), plus a Thai-covering font, plus a CJK font (Noto CJK is ~16MB uncompressed; even
aggressively subsetted for preview you cannot subset ahead of unknown data), plus a Thai break
dictionary (~1MB). A 30–60MB first load, in a tool whose selling point is "open it and work."
"May be significant" is the understatement that will kill FR34 in month 14.

**Do:** Demote WASM from requirement to option. State the capability: *"the exact preview is
produced by the same `folio-go` build that renders in production, with no data leaving the user's
machine."* A localhost helper process satisfies that, satisfies NFR8, satisfies NFR2, satisfies
NFR5, and drops NFR1's cross-target clause — the single hardest constraint (F2) — at a stroke.
Answer Q9 with a number before committing to WASM.

---

## F14 — HIGH — "Fully offline web app with local file save" assumes browser capabilities no FR requires

**Location:** FR8, FR35, NFR8, §5.1

FR8: open and save local files "with no server round-trip and no account." FR35: operate "fully
offline." §5.1: "a web application following the draw.io model."

Unstated dependencies, every one of them load-bearing:

- **File System Access API** (`showOpenFilePicker` / `showSaveFilePicker`) is the only way to
  *save back to the same file*. It ships in Chromium. It does **not** ship in Firefox or Safari.
  The fallback (`<input type=file>` + download) cannot save in place — Ploy in UJ-1 "saves the file
  back to her laptop" would instead get a new file in ~/Downloads every time, breaking the Git
  workflow the whole positioning rests on.
- **Offline operation** of a web app requires a service worker, a cache manifest, and a PWA install
  path — plus a strategy for a 30–60MB WASM payload (F13). No FR mentions any of this.
- **Where the app is served from** is never stated. A "web app" with "no server" still needs a
  first load from somewhere, and that origin must be pinned for the service worker.
- **No browser support matrix exists anywhere** in the PRD.

**Why it matters:** "works on my machine in Chrome" is a plausible v0.1 answer, but it must be
*written down*, because it changes UJ-1 from "she saves the file back" to "she re-downloads it."

**Do:** Add an FR naming the supported browser(s) and the file-access mechanism, and an FR for the
offline delivery mechanism (PWA/service worker) or an explicit deferral of FR35.

---

## F15 — HIGH — Table footer aggregates are undefined at exactly the point that matters

**Location:** FR26, FR18, FR24, FR25, §7.1

FR26: "render footer aggregates: sum, count, and average." The golden report is a bank statement
with **Date, Description, Debit, Credit, Balance** across up to 50 pages.

Unanswered: is the footer aggregate a **grand total** (all rows), a **page subtotal** (rows on
this page), or a **running total** (rows so far)? Does the footer appear on every continuation page
(FR25 only promises the *header* repeats), only the last page, or both? For a Balance column,
"carried forward / brought forward" subtotals per page are the actual banking requirement, and
none of the three FRs acknowledge the concept exists.

This also interacts with F8: a page subtotal is computable during layout; a grand total in a
*header* is not.

**Do:** Specify. Recommended minimum for v0.1: grand total only, rendered once at the end of the
table; page subtotals and carried-forward explicitly deferred and listed in §4.2 so nobody assumes
them from the golden report's description.

---

## F16 — HIGH — FR40 and FR41 have no observable pass/fail and no success measure

**Location:** FR40, FR41; §7.2

FR40 requires "a located, actionable error — naming the template element and the data path" for
four classes: malformed templates, unresolvable bindings, invalid expressions, and unlayoutable
content. The PRD honestly flags that it **invented** this requirement because the source plan shows
only `panic(err)`.

But it stops at prose. There is no error taxonomy, no error type, no code scheme, no statement of
whether errors are returned singly or as a list (a malformed template has many problems at once —
CI wants all of them), no definition of "element" (an index into `body`? a stable ID? — note that
**no FR requires template elements to have stable IDs**, so there is nothing to name), and no
success measure exercising any of it.

FR41 (validate independently of rendering) has the same problem plus one more: validate *against
what*? Schema only, or schema + expression parse + binding resolution against a sample document?
Those are three different features with three different costs.

**Why it matters:** "actionable error" is the kind of requirement that is declared done by
returning `fmt.Errorf("layout failed: %w", err)`. Two engineers ship wildly different products here,
and the Go developer of §3 — the user who cares about "API ergonomics" — is the one who pays.

**Do:** Add an FR requiring stable element IDs in `.folio` (which also fixes F12's diffability and
the designer's selection model). Define an error type carrying `{ElementID, Path, Kind, Message}`.
Specify multi-error collection for FR41. Add the S8 measure from F6.

---

## F17 — HIGH — Units, origin, and coordinate system are never named

**Location:** FR2, FR5; addendum §C

FR2 sets margins. FR5 sets X/Y and width/height. Addendum §C shows `"margin": {"top": 0, ...}`.
Nowhere in the PRD, the addendum, or the source plan is there a **unit**. Points? Millimetres?
Pixels at 96dpi? Twips? Is the origin the page corner or the margin box? Is Y down or up (PDF's
native coordinate space is Y-**up**, which is the classic trap)? Is font size in points while
positions are in millimetres?

**Why it matters:** this is the most literal "two engineers implement incompatibly" defect in the
document. The designer author picks CSS pixels; the engine author picks PDF points; every template
renders at 75% scale and nobody notices until the golden report is measured with a ruler. It costs
one sentence to prevent.

**Do:** State it in FR10: *"All lengths in `.folio` are in points (1/72 inch), floating point.
Origin is the top-left of the page, Y increases downward; the renderer converts to PDF space."*

---

## F18 — HIGH — §3's "contract between them" is enforced by nothing

**Location:** §3, FR8, FR40, FR41, FR12

§3's thesis: "the `.folio` file is the **contract**" between the two users, and every requirement
should be read against "does this make the handoff more reliable?"

Now find the requirement that enforces the contract. There isn't one.

- Nothing requires the **designer to emit only templates `folio-go` accepts**. FR8 says open and
  save; FR41 says the *library* can validate. The designer can happily write a template the engine
  rejects, and the failure surfaces in Anan's CI, not in Ploy's browser.
- Nothing requires the **designer to surface FR40/FR41 errors** in its UI.
- Nothing requires the designer to **preserve unknown fields** when opening a hand-edited or
  AI-edited template (FR11's whole promise). If Ploy opens a template an AI enriched and saves it,
  does the designer silently drop what it didn't understand? Round-trip fidelity is unspecified.
- Nothing requires a **schema artifact** shipped alongside the format (a JSON Schema would let both
  sides validate, and would let the AI of FR11 write correct templates).

**Do:** Add: *"the designer must validate the template against the shipped schema before save and
must refuse to write an invalid template"*; *"opening and saving a template without edits must
produce a byte-identical file"*; *"the format ships a machine-readable schema."* These three lines
do more for §3's thesis than half of §5.

---

## F19 — MEDIUM — FR27 and FR42 are not requirements; they are wishes with numbers

**Location:** FR27 ("if capacity permits"), FR42 ("optional, only if capacity permits")

A requirement conditional on capacity is unfalsifiable and unschedulable. Neither appears in §7.2,
so shipping without them is invisible; shipping with them is unrewarded. They will consume review
attention in every downstream document — architecture will design for FR42's HTTP surface, epics
will carry stories for it — and then be cut silently.

**Do:** Delete both from §5. Move alternating row styling and the REST service to a "v0.2 candidates"
list. If they matter, requirement them properly with a success measure. FR42 in particular is not
"a few hours" — see F20.

---

## F20 — MEDIUM — FR42 is an internet-facing code execution surface, and §10 declines to model it

**Location:** FR42, §10 ("No security or threat model"), §4.2; addendum §J1

§10: *"MVP has no server, no accounts, and no untrusted multi-tenant surface. This must be revisited
the moment the optional REST service (FR42) becomes real."* But FR42 **is in §5**. It is in scope,
in the FR list, numbered. It is already real enough for architecture to design.

FR42 accepts "the template in the request **or** resolved from a path supplied by the operator."
Read that carefully:

- Template in the request body = accepting an attacker-authored template into an **expression
  evaluator**, an **image decoder**, and a **layout engine with no memory ceiling** (NFR4 admits
  there is none). `{{sum(...)}}` over a crafted structure, a 100,000-row table, a decompression
  bomb in an embedded image — every one is a trivial DoS.
- "Path supplied by the operator" is ambiguous in the worst way: if any part of that path derives
  from the request (`/api/reports/{name}/render` — the `{name}` in the source plan's own example is
  request-controlled), you have arbitrary file read.

Addendum §J1 flags the related contradiction — addressing templates by name implies a server-side
repository, which §4.2 puts out of scope — and leaves it unresolved.

**Do:** Cut FR42 (see F19). If it survives: template in the request body only, never a path; a
hard limit on rows, pages, template bytes, and image bytes; a render timeout; and a threat-model
paragraph replacing §10's exemption.

---

## F21 — MEDIUM — FR5's font-family property has no legal values

**Location:** FR5, FR31, NFR7, §4.2 ("advanced font management UI" out of scope)

FR5 gives the author a **font family** property. FR31 requires all fonts embedded. NFR7 admits that
which fonts ship, how a user supplies a custom font, the licensing, and missing-glyph fallback are
**all unspecified**. §4.2 puts font management UI out of scope.

So: the designer offers a font picker with an undefined list, backed by an undefined shipped set,
with no way to add a font, and undefined behaviour when a glyph is missing. What does the picker
show? What happens when Ploy types a family name the engine doesn't have — silent substitution
(breaking NFR1's promise of identical rendering), or an FR40 error?

**Do:** For v0.1: ship exactly one family with Latin+Thai coverage in regular/bold/italic/bold-italic,
make FR5's font-family an enum of that one family, and state the missing-glyph policy (recommended:
render `.notdef` and emit a warning — never silently substitute, which is a determinism hazard).
Custom fonts to v0.2.

---

## F22 — MEDIUM — FR12 requires enforcing a compatibility policy that NFR6 says does not exist

**Location:** FR12, NFR6, FR10

FR12: the library validates the format version on load so "a template authored against a future
format fails clearly rather than rendering incorrectly." NFR6: *"Forward and backward compatibility
rules between template versions and library versions are undefined and need a policy before v1."*

FR12 cannot be implemented without the policy NFR6 says is missing. Is `"1.0"` semver? Does
`folio-go v0.1` accept `1.1`? Reject `2.0` but accept `1.x`? Reject anything not exactly equal?
Addendum §C shows `"version": "1.0"` in a **v0.1** product — the template format is already
versioned ahead of the library, which is at minimum confusing and at worst two version schemes
nobody has reconciled.

**Do:** Decide it in one line now: *"`version` is `MAJOR.MINOR`; the library accepts equal MAJOR
and MINOR ≤ its own; anything else is a fatal load error naming both versions."* Start the format
at `0.1` to match the library.

---

## F23 — MEDIUM — `formatDate` / `formatNumber` require a locale model that §10 explicitly refuses to provide

**Location:** FR17, §10 ("No i18n beyond script rendering"), NFR3, UJ-1; addendum §D

§10 states that NFR3 covers rendering Thai and CJK *text* but provides "no locale model for
`formatDate` / `formatNumber`." The addendum agrees: "Format-pattern conventions imply a locale
model that the plan never defines."

The golden report is **a Thai bank statement**. Thai financial documents routinely use the
**Buddhist era** (2569, not 2026) and Thai month names. `formatDate(d, "dd/MM/yyyy")` with no
locale, no calendar, and no timezone renders a date that is wrong by 543 years for the exact
document the MVP is validated against. `formatNumber("#,##0.00")` needs a decimal and grouping
separator — fine for Thai/US, wrong for most of Europe.

And per F5, an unpinned timezone is a determinism hazard: the same render on a UTC container and a
UTC+7 laptop can print different dates and **different bytes**, breaking S3.

**Do:** State the v0.1 answer explicitly: Gregorian, UTC, `.` decimal, `,` grouping, English month
names, no locale parameter. Then either accept that the golden report shows Gregorian dates, or
add a Buddhist-era pattern token — but decide, and do not let §10's disclaimer sit next to a Thai
banking acceptance test unremarked.

---

## F24 — MEDIUM — Q5 and Q6 are not questions; they are known defects in the build order

**Location:** Q5, Q6; addendum §H, §J3, §J4

Q5: *"Does the build order stand, given tables (Phase 3) depend on aggregation from the expression
engine (Phase 4)?"* — No. It does not. FR26 requires sum/count/avg footers; FR21 requires collection
binding; the expression engine is Phase 4. Phase 3 even *duplicates* "Aggregation" and "Data
collection binding" in its own list. The addendum states this flatly in §J3. This is a defect with a
known fix (swap phases 3 and 4, or split expressions into "paths + aggregation" early and
"formatting + logic" late), dressed up as an open question.

Q6 is the same move applied to F11: the objection to the plan's own priorities, filed rather than
answered.

**Why it matters:** questions get deferred; defects get fixed. Filing a defect as a question buys
another planning cycle of nobody touching it, and §8 claims to be "ordered by how much they block
downstream work" — Q5 blocks epic sequencing, which is the very next artifact.

**Do:** Delete Q5 and Q6. Publish the corrected order in the PRD: expressions and path resolution
before tables; designer file round-trip early (F11).

---

## F25 — MEDIUM — FR9 and Q8 ask the same question twice and neither answers it, while preview correctness depends on it

**Location:** FR9, Q8, FR7, FR34, §9

FR9 asserts (as `[ASSUMPTION]`) that sample JSON is a separate local file "the user re-associates
per session." Q8 then asks "where does sample JSON live relative to the template during design?"
The register logs both. Three entries, zero decisions.

It is not a UX detail. It determines: whether opening a template in a fresh session can preview at
all without a second file dialog; whether the paths offered by FR7's binding UI survive a reload;
whether a template can be validated in CI (FR41) against a known-good sample; and whether a
teammate pulling the `.folio` from Git can preview it at all. "Re-associates per session" is a
prediction that Ploy will re-pick a file every single time she opens the tool — a UX tax the PRD
imposes without noticing.

**Do:** Decide: a sibling convention (`customer-statement.folio` + `customer-statement.sample.json`,
auto-loaded when present) with an optional recorded relative path in the template. Costs nothing,
removes the tax, makes CI validation possible.

---

## F26 — MEDIUM — The Go API has no concurrency contract, no cancellation, and no defined input types

**Location:** FR36–FR39; addendum §B

The §3 user who "cares about API ergonomics, determinism, streaming, and memory behaviour under
load" is given six FRs, and none of them answer the questions that user actually asks:

- Is a loaded `Template` safe for **concurrent** use across goroutines? A microservice loads it
  once at startup and renders from N handlers. If not, every integrator gets a data race.
- Is there a `context.Context` for cancellation and timeouts? FR38 streams to an
  `http.ResponseWriter`; when the client disconnects mid-50-page render, nothing stops it. That is
  the memory-under-load failure the same user cares about.
- What does `data any` accept — `map[string]any` only (the plan's examples), arbitrary structs via
  reflection, `json.RawMessage`, `[]byte`? Each is a different implementation with different
  determinism properties (struct field ordering, `json` tags, map iteration order).
- What happens on partial failure mid-stream, after bytes are already on the wire?

**Do:** Add FRs: template is immutable and safe for concurrent renders; `context.Context` as the
first parameter of both render calls; input types stated as `map[string]any` + `json.RawMessage`
for v0.1 (reflection over structs deferred); mid-stream failure returns an error and the caller is
responsible for the truncated response.

---

## F27 — MEDIUM — Image support is one word in FR4 and one constraint in FR32

**Location:** FR4, FR32, FR5

"Image" is a palette entry. Nothing states supported formats, and the answer changes the project
size by an order of magnitude:

- PNG/JPEG — straightforward embedding as PDF XObjects.
- **SVG** — requires a vector interpreter and, for text-bearing SVGs, the whole font stack again.
- PNG with alpha — requires a soft-mask (`SMask`), which most naive embedders get wrong.
- CMYK JPEG, 16-bit PNG, animated GIF — each a special case.

Also unspecified: fit/scale mode (stretch, contain, cover, none), DPI interpretation, and whether
the designer downsamples on embed (which interacts with F12's file-size problem).

**Do:** State it: *"PNG (8-bit RGB/RGBA) and JPEG (baseline, YCbCr) only; fit mode from
{contain, stretch}; images embedded base64 in the template; no downsampling."* Everything else
explicitly out.

---

## F28 — MEDIUM — NFR2 and FR33 contradict each other about the design canvas

**Location:** NFR2, FR33, NFR5

NFR2: the browser's text and layout engines "must never determine pagination, line breaking, or
measurement — **including in the design canvas**."

FR33: the design canvas is a "**fast** ... interactive editing" view, "understood to be an
**approximate** visual representation."

Both cannot hold. If the browser may not measure text, then every keystroke and every drag must
round-trip into WASM for measurement — which is where "fast" goes to die, and no performance
requirement exists to define what "fast" means. If the canvas *is* allowed to approximate with
browser metrics, NFR2 is violated by design. And NFR5's "no separate preview renderer to drift"
is contradicted by FR33's own admission that the canvas is approximate — that *is* a second
renderer, and it *will* drift.

**Do:** Pick one and say it: *"the design canvas uses `folio-go` measurement via WASM for text
extents; it does not paginate. It is not a preview and may differ from the PDF in pagination
only."* Then add a latency budget (e.g. p95 < 16ms per measurement batch) so "fast" is testable.

---

## F29 — LOW — C1 counts ten expression functions; FR17 specifies eight

**Location:** C1 vs FR17

FR17 lists: `sum`, `count`, `avg`, `formatDate`, `formatNumber`, `upper`, `lower`, `if`. That is
**eight**. C1 reads: *"Expression function count exceeding the **ten** specified."*

A counter-metric that miscounts the thing it counts grants two free functions before it trips, and
signals that §7.3 was written without re-reading §5.3.

**Do:** Change C1 to eight. Then re-derive the rest of §7.3 against the FRs — C2 ("palette
exceeding five") is the only counter-metric that currently checks out.

---

## F30 — LOW — C5 measures a user who does not exist

**Location:** C5, §7

C5: *"Time-to-first-PDF for a new integrator exceeding **a few minutes**."* §7 states that Folio is
"presently built to scratch its author's own itch" and that success is "capability and reliability,
not adoption." There are no integrators. There is one, and he wrote the library.

"A few minutes" is also not a number, and the metric measures typing speed and Go module download
latency as much as API design.

**Do:** Replace with something the author can actually run: *"the quickstart in the README, copied
verbatim into an empty module, produces a valid PDF — verified by a CI job that does exactly
that."* That catches the real failure (a quickstart that has silently gone stale) and is
falsifiable.

---

## F31 — LOW — §9's register omits its own biggest assumption, and §8 omits its biggest question

**Location:** §9, §8

The register collects ten `[ASSUMPTION]` tags. It does **not** contain:

- the assumption that cross-target byte-identical output is **achievable at all** (F2) — the single
  largest bet in the document, stated as a MUST in NFR1 and as an open question in Q1;
- the assumption that Thai and CJK are in scope at all, reversing source §14 (F3);
- the assumption that WASM is the preview mechanism, reversing source §11's Preview API (F13).

And §8, "ordered by how much they block downstream work," does not contain the layout-model
question (F1) — the question that blocks the most work.

**Why it matters:** the register's value is that it is *complete*. A register that catches "sample
JSON lives in a separate file" but not "the product's signature requirement may be impossible"
gives false confidence to every reviewer downstream, who will reasonably assume the big risks were
the ones that got tagged.

**Do:** Add all three assumptions and the layout question. Re-order §8 so Q1, the layout model, and
Q2 sit above the UX questions.

---

# The v0.1 That Can Actually Ship

The workflow to prove is **Design → Bind → Preview → Render**. Here is the smallest thing that
proves all four arrows, for one person.

**Cut:**

| Cut | Why it is safe | What breaks |
|---|---|---|
| **CJK** (keep Latin + Thai) | UJ-1 is a Thai bank. CJK adds Type0/CID fonts, 20k-glyph subsetting, UAX#14. | S4 halves. NFR3 narrows. Nothing in the golden report. |
| **Cross-target byte equality** (NFR1b) | Keep byte-identical on one target + layout-identical across platforms. The engine is unchanged; you can restore it in v0.2. | S2 dies. C3's absolutism softens. The marketing claim waits one release. |
| **WASM preview** → localhost helper process | Same engine, same bytes, still offline, still private (NFR8 intact). Removes the 30–60MB bundle (Q9) and the hardest half of NFR1. | FR34's "in the browser tab" phrasing. Nothing a user notices. |
| **`if()`, conditional visibility (FR19), `avg`, `upper`, `lower`** | `if()` drags in an operator set and a type system (F4). The rest are one-liners you can add any time. | FR19 goes. §4.2 gets cleaner, not messier. |
| **Custom page size (FR2)** — A4 portrait only | The golden report is A4 portrait. | One property. |
| **FR27 (alternating rows), FR42 (REST)** | Already optional. Deleting them stops architecture designing for them (F19, F20). | Nothing. |
| **Line + Rectangle** — *last to cut* | Cheap once the renderer exists; cut only if Phase 3 overruns. | C2's palette count. |

**Keep, and specify properly first:**

- The layout model (F1) — flow content region between absolute header/footer bands.
- Units and origin (F17) — points, top-left, Y-down.
- The expression grammar (F4) — paths, `params.`, `row.`, `sum()`, `count()`, `formatNumber()`,
  `formatDate()`. Written as EBNF before any code.
- Stable element IDs + a shipped JSON Schema (F16, F18).
- Overflow policy (F10) and aggregate semantics (F15).
- The error type (F16).
- Generated date as a parameter (F5).

**Resequence** (replacing addendum §H, and answering Q5/Q6 rather than filing them):

1. **Spike Q1** — deterministic PDF + subset font embedding, native only. Two weeks. Go/no-go on
   the entire premise.
2. Schema + parser + validator + path resolution + `formatNumber`/`formatDate`. Hand-written
   golden `.folio` renders one page.
3. Thai + Latin measurement, shaping, wrapping. Multi-page, header/footer bands, `Page X of Y`.
4. Tables: rows, columns, page breaks, repeated headers, grand-total footer. Golden report at
   1/5/20/50 pages.
5. Determinism harness in CI: byte equality across two runs and two platforms, one target.
6. **Designer** — canvas, five (or three) components, properties panel, **table column editor**,
   JSON-path binding UI, open/save. Authors the golden template from an empty canvas.
7. **Preview loop** — localhost helper, save-the-PDF-out, error surfacing from FR40/FR41 in the UI.

Steps 1–5 give the Go developer a shippable library. Steps 6–7 give the template author a real
tool and turn S5 into a measure with teeth. Both users are served, in that order, and the order is
now stated rather than denied.

**If you will not cut NFR1b** — then cut the designer from v0.1 entirely and ship `folio-go` with
hand-written templates and a CLI preview. That is a coherent product and an honest release. What
is not coherent is one person building a cross-target byte-reproducible typesetting engine *and* a
visual design tool in the same release while §3 insists both users are equal.

---

# Severity Counts

| Severity | Count | Findings |
|---|---|---|
| **Critical** | 7 | F1, F2, F3, F4, F5, F6, F7 |
| **High** | 11 | F8, F9, F10, F11, F12, F13, F14, F15, F16, F17, F18 |
| **Medium** | 10 | F19, F20, F21, F22, F23, F24, F25, F26, F27, F28 |
| **Low** | 3 | F29, F30, F31 |
| **Total** | **31** | |

**Blocking before epics and stories:** F1 (layout model), F2 (spike Q1), F3 (cut CJK / scope
shaping), F4 (write the grammar), F5 (generated date), F17 (units). Everything else can be fixed
in flight. These six cannot.
