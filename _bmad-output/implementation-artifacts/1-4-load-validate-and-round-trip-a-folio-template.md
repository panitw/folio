# Story 1.4: Load, validate, and round-trip a `.folio` template

Status: done

**Story key:** `1-4-load-validate-and-round-trip-a-folio-template`
**Epic:** 1 · **Baseline:** `af6b5d5`
**Covers:** FR11, FR12, FR13, FR37 · NFR6 · AD-9, AD-10
**Adjacent:** AD-1, AD-2, AD-3, AD-5, AD-14, AD-22, AD-23
**Rulings that govern:** D-1.4.1, D-1.4.2, D-1.4.3 **(extended)**, `D-1.4.4` **(reversed in part)**,
D-1.4.5, D-1.4.6, D-1.4.7, D-1.4.8 **(corrected)**, **D-1.4.9**, **D-1.4.10**, **D-1.4.11**, **D-1.4.12**, **D-1.4.13**,
**D-1.4.14**, **D-1.4.15**, **D-1.4.16**, D-000.9, D-000.4, D-000.5, D-000.6, D-000.10, D-000.11, D-1.1.b, D-1.1.c,
D-1.2.5, D-1.2.6, D-1.3.4

> **The three conflicts this story surfaced are all ruled.** Nothing is parked. See
> [§ Decisions resolved](#decisions-resolved) for what was ruled and why — the reasoning is carried
> because two of the three reversed a prior ruling and one is an owner override, and a developer who
> does not know why will re-derive the losing option. **Nothing is confirmation-pending** — AC7's
> `version` clause was confirmed as `D-1.4.13`, and `D-1.4.12` (owner) settled closed-set evolution.
>
> **Finisher's note (added at close-out).** Four further rulings landed in the decision log after
> this story was drafted and were never folded in — the code-reviewer's own review (this story's `##
> QA Results`) caught this as Finding 21 and built Findings 1, 2, 6, 7, 8 and 13 directly on three of
> them. **D-1.4.14** (the shared length-spelling table's location), **D-1.4.15** (AST extraction is
> necessary but not sufficient for the drift test) and **D-1.4.16** (two behavioural guards: presence
> serializes distinctly; `Render` ignores its argument today) are now added to the header above and
> disposed of in [§ Finding Resolutions](#finding-resolutions). **D-000.10** binds the *reviewer's*
> disposition table from Story 1.5 onward, not this story's developer — its dispositions already
> appear at the foot of this file; **D-000.11** (every gate runs `-count=1`) is a program-wide
> standing rule this story's own review discovered, applied here in `.github/workflows/ci.yml`.

---

## In plain terms (read this first if you just want the gist)

A Folio template is a file a person can open in a text editor, read, edit and hand to the engine.
This story taught the engine to read one, check it, and write it back out again — and it works: a
wrong template is rejected out loud, naming the field, the value and the item it belongs to; a
template from a much newer generation of the format is refused outright rather than guessed at; and
a template from a slightly newer generation loads, carrying forward everything the engine does not
personally recognise, unchanged, so that opening and re-saving a newer file never quietly throws
content away.

That last promise turned out to be only mostly true when the work was first reviewed. Content this
engine did not recognise was being carried forward almost everywhere — except at several specific
spots buried a level or two inside the document, where it was silently refused instead. A second,
related gap sat in the machinery meant to keep this file's own documentation honest: an ordinary code
change could add a new, working piece of the format without that change ever showing up anywhere the
checks would notice, because of exactly how the "code vs. documentation" comparison happened to be
written. Both were demonstrated failing, then fixed, then demonstrated working. One further spot that
looked like it should carry forward unrecognised content genuinely does not, on purpose — that one is
a fixed, named list the format itself declares closed, not a gap.

Writing back out still matters because the format has exactly one legal spelling: two people editing
the same template must produce the same bytes, or every save becomes a change nobody can review.
Item labels are still never renumbered, reused, or quietly repaired.

**Deliberately not in this story.** Working out which numbers a column total should add when the
author has not said; the stable error codes the product will publish later; and computing any total
at all. A column total with no stated source still simply loads today — a known, recorded gap, not an
oversight.

---

## Story

As an integrating Go developer,
I want to load a `.folio` file and have malformed or future-versioned templates rejected with a
located error,
So that a bad template is caught in CI rather than in production.

---

## Acceptance Criteria

Source ACs are `_bmad-output/planning-artifacts/epics.md` §Story 1.4 (lines 509–540), reproduced
verbatim in [§ Source acceptance criteria](#source-acceptance-criteria). **Where a ruling is quoted,
it is quoted verbatim with its ID — never paraphrased** (D-1.1.b: *"A story AC that paraphrases a
ruling is where the ruling gets silently widened."*).

### API surface and the `os` boundary

**AC1** — Package `folio` at the module root exports both `LoadTemplate(path string)` and
`ParseTemplate(b []byte)`, each returning `(*Template, error)`. Per **D-1.4.6**: *"So
`LoadTemplate(path string)` and `ParseTemplate(b []byte)` live in package `folio` at the module
root, where `os` is permitted; **`internal/template` never sees a path** — it takes `[]byte` and
returns a `*Template` or an error. AD-1 is untouched because no `internal/` package imports `os`."*
`LoadTemplate` delegates to `ParseTemplate` — *"this is honest factoring rather than scope growth,
and it does not violate the one-parameter-per-story rule, which governs `Render`."*

**AC2** — `Render`'s signature becomes exactly `func Render(t *Template) ([]byte, error)`. Per
**D-1.4.6**: *"**Verdict — signature after 1.4**, per D-1.1.c's one-parameter-per-story convergence:
`func Render(t *Template) ([]byte, error)`."* Story 1.1's `PROVISIONAL — Story 1.1 only` doc marker
is updated, not deleted — the signature is still provisional through 1.7 (**D-1.1.c**).

**AC3** — No package under `folio-go/internal/` imports `os`, `time`, `math/rand` or `net`; the
existing AD-1 lint reports zero findings with its own coverage witness non-zero.

### Loading a well-formed template

**AC4** — Given a well-formed `.folio` file carrying a `version`, page setup, and ordered band
content, `folio.LoadTemplate(path)` returns a `Template` with no error.

**AC5** — The closed sets `folio-format.md` states are enforced at load, each a load error naming
field and offending value:
- `locale` ∈ {`en`, `th`, `zh-Hans`, `ja`} — *"An unlisted tag is a load error (AD-12)"* (`:47`);
- element `type` ∈ {`text`, `image`, `table`, `line`, `rect`} — *"The set is closed (FR4); a sixth
  type is a load error"* (`:119`);
- `bands` has exactly `content`, `pageFooter`, `pageHeader`, with `height` on the latter two and
  **not** on `content` — *"Storing it would be a second source of truth"* (`:94–96`);
- `page.size` is `"A4"`, `"Letter"`, or an object carrying `height` and `width`; `page.orientation`
  ∈ {`portrait`, `landscape`};
- a `table` declares `x` and `y` and **neither `width` nor `height`** (`:149–151`, AD-13);
- `utcOffset` matches `±HH:MM`.

> **These closed sets bind the keys and values this library models.** They are **not** contradicted
> by AC8's unknown-key passthrough: an unrecognised *key* is carried through, while a recognised key
> carrying an unlisted *value* is still a load error. Keep the two paths visibly separate in the
> code — conflating them is how passthrough silently becomes "validation is optional".

### Version validation (FR13) and forward compatibility (D-1.4.9)

**AC6 — higher MAJOR is a load error.** Given a template declaring a MAJOR format version newer than
the library supports, loading fails with an error naming the declared version **and** the supported
version, and no render is attempted. `folio-format.md:46`: *"A higher `MAJOR` than the library
supports is a load error, never a best-effort render (FR13)."* `version` is a **string** —
**D-1.4.3**: *"so it carries no round-trip hazard."*

**AC7 — higher MINOR loads, and `version` is carried verbatim in both directions.** Per **D-1.4.9**:
**a higher MINOR loads.** Per **D-1.4.13**, the rule is symmetric, and the naive implementation
("write the library's own version") is wrong in **both** directions:

> **`version` is a property of the document, carried verbatim; raised only when content actually
> requiring a higher version is introduced; lowered never.**

- **Never lowered.** A `1.0` library saving a `1.1` file writes `"version": "1.1"` — the file still
  contains `1.1` content.
- **Never gratuitously raised.** A future `1.1` library loading an **unchanged** `1.0` file must
  write `"1.0"`. Otherwise `Serialize(Parse(b)) != b` for **every** older file, and the naive rule
  breaks the exact property passthrough was chosen to strengthen.

**At Story 1.4 the raise path is unreachable** — no `1.1` exists. **Do not build it; do not foreclose
it.** Isolate the decision behind a single well-named function so the raise path is one place to fill
in later.

**This reframe goes in that function's doc comment verbatim** — it is what stops the well-meaning
"fix":

> **`version` describes the document, not the writer.** A PDF 1.7 file edited by a tool that
> understands only 1.4 features does not become a PDF 1.4 file.

**Two further reasons to carry** (**D-1.4.13**): a downgrade would give the same document **two
canonical byte forms** depending on which library last wrote it — precisely what **AD-9's Prevents
line** exists to stop (*"the designer and a hand-editor producing different bytes for the same
document"*); and the spine plans `folio-node/` / `folio-java/` conforming against the **same fixture
corpus**, so a downgraded version propagates the lie to every future SDK.

**Two cheap tests pin the rule:** a `1.0` file round-trips as `1.0`; a synthetic higher-MINOR file
round-trips as `1.1` with unknown keys intact.

**AC8 — unknown keys pass through opaquely.** Per **D-1.4.9**: **unknown keys are carried opaquely
through parse, merged back in sorted order on serialize, and ignored when rendering. Nothing is
dropped, nothing is refused.** Merged *in sorted order* means an unknown key sorts among the known
keys by the same byte-order rule as AC13 — the serializer emits one sorted key sequence, not a known
block followed by an unknown block.

**AC9 — P1 holds *including* unknown keys.** The corpus carries a fixture with unknown keys at
**multiple nesting levels** — top level, inside a band, inside an element, inside a column, inside
`style` — not just the top. This fixture is a fixed point under AC11 exactly as any other canonical
file is.

**AC10 — the passthrough boundary is stated and enforced.** Two boundaries:
- **Against the drift test (AC42):** D-1.4.7's test governs keys **the serializer itself can emit**
  from its own struct tags. Unknown passthrough keys are explicitly **outside its scope**. State this
  in the drift test's doc comment, *"or a future maintainer will read a passthrough key as drift."*
- **Against the render path:** unknown keys **never reach `internal/layout` or `internal/pdf`**. AD-5
  and the determinism boundary are untouched. Since neither package exists yet, this is asserted
  structurally — the passthrough store lives on the template model and is not a field any render-path
  consumer reads.

**AC11 — the mandatory safety sentence ships in `folio-format.md` in this story's commit.** Verbatim,
in its **full three-clause form**: clauses 1–2 (additive-only, meaning) are **D-1.4.9**'s own
guardrail sentence; clause 3 (closed-set evolution is MAJOR) is **D-1.4.12**'s alone — D-1.4.9's
guardrail is a single clause, and attributing all three to both rulings over-attributes even though
the shipped text is correct (this story's finisher review, Finding 20):

> A MINOR increment may add **new optional keys** only. It may **not** change the meaning of an
> existing key, and it may **not** extend a closed set of legal values (element `type`, `locale`,
> `align`, `valign`, `columns[].footer`, `border.edges`, `page.orientation`, `page.size`). Extending
> a closed set is a **MAJOR** change, because every existing library validates those sets as load
> errors.

**Clause 2 — meaning.** Per **D-1.4.9** the chosen option is **unsafe without it**: *"Without it a
future MINOR could redefine a key this library already understands, and today's library would render
such a file **confidently and wrongly** — the one failure mode worse than refusing it."*

**Clause 3 — closed sets, and why passthrough does not save them (D-1.4.12, owner-confirmed).**
**Passthrough carries unknown *keys*; a new *value* in a known enum is a known key whose validation
rejects it.** Without this clause a `1.0` library would refuse a `1.1` file that merely added a sixth
element type — voiding the forward compatibility D-1.4.9 paid for, for exactly the additions most
likely to be made. This is the AC5 guard made normative: an unknown **key** passes, an unlisted
**value** on a known key is still a load error, and the two paths stay visibly separate in the code.

**Rejected alternatives, recorded because the first will be re-proposed:**
- *Older libraries skip unknown elements and report the skip* — rejected as conflicting with
  **FR13**'s *"never a best-effort render"*: on a bank statement a silently missing section is
  materially worse than a refused file.
- *Decide per set* — rejected because all eight genuinely cannot be faked by an older library (an
  unknown `locale` has no font or formatting rules; an unknown `align` would be silently wrong).

This is not optional polish; AC11 gates AC7 and AC8.

> **The accepted cost, stated plainly.** Two costs, both permanent.
> **(D-1.4.9)** This is the most machinery of the four options considered, and it lands in the one
> story whose job is to establish P1, P2, P3 and the drift test. The owner accepted that deliberately
> — see [§ Decisions resolved](#dn-3--d-149).
> **(D-1.4.12)** Evolution on those eight sets is now **permanently expensive**: a sixth element type
> or a fifth locale orphans every older library. The side effect is deliberate — it protects PRD
> counter-metrics **C1** (*"Expression function count exceeding eight"*) and **C2** (*"Component
> palette exceeding five"*), because *"just add one more element type"* now costs a MAJOR version
> rather than sliding in on a MINOR.

### The round-trip properties (the story's core)

**AC12 (P1)** — **D-1.4.3**: *"**P1 — round-trip through bytes:** for every `Document d` obtained by
parsing a valid file, `Parse(Serialize(d)) == d`."* Equality on the parsed value, not on bytes.

**AC13 (P2)** — **D-1.4.3**: *"**P2 — normalisation is idempotent:** for every valid input `b`,
`Serialize(Parse(b)) == Serialize(Parse(Serialize(Parse(b))))`."*

**AC14 (P3)** — **D-1.4.3**: *"**P3 — canonical is a fixed point:** `b` is canonical **iff**
`Serialize(Parse(b)) == b`."* And: *"The AC's `Serialize(Parse(b)) == b` then holds for canonical `b`
**by P3's definition**, and the substantive tests are P1, P2, and a corpus of canonical fixtures …"*
(continued, verbatim, in AC15 below).

**AC15 — the fixture corpus.** Per **D-1.4.3** it starts with *"`folio-format.md`'s own worked
example … which ships as a golden fixture under AD-21"* — now in its **amended, fully-expanded** form
per **D-1.4.10** (AC16). The corpus also carries the unknown-key fixture (AC9), the trap fixtures
(AC17–AC21), and the id fixtures (AC30–AC34).

**AC16 — the worked example is amended, and the amendment may not ratify the story's own serializer.**
Per **D-1.4.10** the worked example is rewritten to the fully-expanded canonical form (option a). The
guardrail is **mandatory**, and it exists because of this story's own D-000.9 exposure — if the story
generates canonical bytes with its own serializer and pastes them into the document, **the document
ratifies the story's guess and the first fixture proves nothing**:

1. The amended example is **cross-checked against an independent JSON pretty-printer** —
   `json.dumps(..., sort_keys=True, indent=2, ensure_ascii=False)` or equivalent — for **structure,
   indentation and key order**.
2. Known divergence classes are **enumerated and normalised explicitly**, and **each normalisation is
   named in the Delivery Log**. *"An unnamed normalisation is exactly where self-ratification would
   hide."*
   The classes are measured and listed in **M-9**; the developer confirms the list is complete for the
   amended file rather than assuming it.
3. **A test asserts the document's fenced example block is byte-identical to the shipped golden
   fixture**, closing the loop permanently. The test extracts the fence from the live document — it
   does not embed a copy (a copy is a third artifact, D-1.4.7's forbidden shape).

The example stays valid under **D-1.4.1** without adding `footerOf`: its `bind` is a single
`formatNumber(<bare row-scoped path>, <pattern>)`, which is derivation shape #2.

**AC17** — These tests ship **with** the format, not after it (AD-9, epic AC).

### The canonical byte form (AD-9)

**AC18** — Serialized bytes have: keys sorted, two-space indentation, LF line endings, no trailing
whitespace, trailing newline. Key sort is **byte-order over the UTF-8 key** — **D-1.4.3**: *"**Key
sort is byte-order over the UTF-8 key**, not locale-aware — all keys are ASCII lowerCamelCase so it
never bites, stated anyway because a locale-aware sort would be invisible until someone adds a key."*
Asserted by construction, not by coincidence over an all-ASCII corpus. **Unknown passthrough keys sort
in the same sequence** (AC8) and may be non-ASCII, so this AC is no longer only theoretical.

**AC19 — Trap 1, HTML escaping OFF.** **D-1.4.3**: *"**HTML escaping OFF** (`json.Marshal` escapes
`<`, `>`, `&` by default; left on, a template containing `<` round-trips to different bytes than a
hand-editor wrote)."* Retained fixture containing `<`, `>`, `&`. **Measured live** (M-2): the default
really does escape all three.

**AC20 — Trap 2, UTF-8 emitted literally.** **D-1.4.3**: *"**UTF-8 emitted literally**, no `\uXXXX`
above `0x1F` — the golden report is a **Thai** bank statement, and escaping Thai or CJK would destroy
the readability contract `folio-format.md` states as its purpose."* Retained Thai + CJK fixture; no
`\u` sequence in the bytes.

> **Narrowed to match the measurement (M-2), per the ruling on this AC.** Thai and CJK emit
> **literally under either escaping setting** — this is a real requirement but **not a default
> hazard**, and the story must not imply the default mangles Thai. It would become live the moment
> anyone hand-rolls string escaping, which is why the assertion still ships. The `<`/`>`/`&` half of
> the trap (AC19) **is** live against the default.

**AC21 — Trap 3, minimal escaping only.** **D-1.4.3**: *"**Minimal escaping only** — `"`, `\`, and
controls below `0x20`; `/` is never escaped."* Fixture covering all four classes.

**AC22 — Trap 4, `omitempty`, and it is two-sided.** Per **D-1.4.3**: *"**No `omitempty` on any field
whose zero value is legal authored content.** An `omitempty` on `assets`, `x`, `y` or `fontSize` turns
an authored `0` or `{}` into an absent key and breaks P1 silently for exactly the documents that look
most ordinary."*

Per **`D-1.4.3` (extended)**, the rule is now **two-sided**: *"no `omitempty` where the zero value is
legal content, **and** every map and slice field must be initialised to empty rather than left nil."*

**AC23 — Trap 5, the nil-container asymmetry.** *(Found by measurement, M-2; now ruled into
`D-1.4.3` (extended).)* A nil map or nil slice serializes to **`null`** — not `{}`/`[]`, not an absent
key — so removing `omitempty` is **necessary but not sufficient**, and the combination breaks P1
against the worked example's `"assets": {}`. Fixtures in both polarities.

A **structural** test covers AC22 and AC23 together, not a per-field one: reflect over every
serialized struct tag and fail on (a) any `omitempty` on a field whose kind admits a legal zero, and
(b) any map/slice field a constructor can leave nil. It reports **how many tags it inspected**, and
**zero inspected fails** (D-000.9).

### Numbers (D-1.4.3, D-1.4.5)

**AC24 — two kinds, no third.** **D-1.4.3**: *"Exactly two numeric kinds exist in a `.folio` file, and
there is no third. **Points** (`x`, `y`, `width`, `height`, `fontSize`, `headerHeight`, `padding.*`,
`margin.*`, `border.width`, `columns[].width`) are held as **`geom.Length`** (`int64` millipoints),
converted on load by **exact ×1000 decimal-string arithmetic, never through `float64`** … **Plain
integers** (`nextId`) are `int64`, plain digits."* And: *"Report-data numbers (AD-23's exact scaled
decimals) are **not** in the template — that type is Story 1.6's and 1.4 must not build it."*

**AC25 — on-disk spelling.** **D-1.4.3**: *"on disk they are points, up to three decimals, **trailing
zeros trimmed, no trailing `.`, no `-0`, no exponent** — the same spelling `appendLength` produces
(`36` not `36.0`; `72.5` not `72.500`)."* `appendLength` exists with exactly this contract but is
unexported under `internal/pdf`, which `internal/template` may not import. Produce the **same
spelling**, not a shared symbol; a cross-check test over a shared value table makes a future
divergence a red test rather than a byte surprise.

**AC26 — decoding.** **D-1.4.3**: *"`json.Decoder` with **`UseNumber()`** — not by analogy with AD-23
but because the alternative is unavailable: `encoding/json`'s default decodes every number to
`float64`."*

**AC27 — input tolerance.** **D-1.4.3**: *"Non-canonical but valid input is **accepted and
normalised**, not rejected: `36.0000` → `36`, `1e3` → `1000` … **Legality is a property of the value,
not the spelling** … `int64` millipoint overflow is a load error, never a wrap."* Both named examples
plus overflow, retained.

**AC28 — three decimals.** **D-1.4.5**: *"A geometry value that is not an exact whole number of
millipoints is a **load error**. Nothing is rounded at load."* And: *"`36.0000` is legal (exactly
36000 millipoints) and normalises to `36`; `72.5001` is a load error."* Both polarities. No second
rounding site: *"There remains exactly **one** rounding mode in the program, in exactly one function
(`geom.ScaleRound`), used only where a value is *computed* into thousandths."*

**AC29 — the classification inventory and its coverage witness.** **D-1.4.3**: *"the developer must
enumerate every numeric field in `folio-format.md` and classify each into kind 1 or kind 2, with a
test asserting **no field escaped classification**. That enumeration is a coverage witness in
D-000.9's sense."* The test reports **how many numeric fields it considered** and **fails on zero**
(*"**Zero candidates is a failure, not a pass.**"*), red-proved **by construction** with a retained
fixture carrying an unclassified numeric field.

**Kept despite measuring exhaustive.** M-3 measures the lead's inventory complete — 11 distinct keys,
zero escapees — which upgrades its stated **medium** confidence. **The AC still ships:** the document
will grow, and a test that was true once is not a coverage witness.

**AC30 — no `float64` under `internal/`.** AD-2 and AD-23. The existing `no-float64` guard stays green
with a non-zero witness over the new package.

### Ids and `nextId` — base 36 (`D-1.4.4` reversed in part, AD-10)

> **`D-1.4.4`'s "not base-36" clause is REVERSED.** `folio-format.md:118` stands as written: *"`e` +
> the counter in **lowercase base 36** — `e1`, `ea`, `e1z`."* **Every other clause of D-1.4.4 stands
> unchanged** — one document-wide counter, `nextId` absent or ≤ highest is a load error, never repair,
> never renumber, and the load-delete-serialize guardrail. Only the *rendering* changes. See
> [§ DN-2](#dn-2--d-144-reversed-in-part) for why, and [§ The transferable
> rule](#the-transferable-rule) for the lesson that must not be lost.

**AC31 — id rendering.** Ids are `"e"` + the counter in **lowercase base 36**:

| counter | id | | counter | id |
|---|---|---|---|---|
| 1 | `e1` | | 35 | `ez` |
| 10 | `ea` | | 36 | `e10` |
| | | | 71 | `e1z` |

**AC32 — `nextId` is a plain decimal JSON integer.** Not base 36. This asymmetry is deliberate and is
the deliverable of AC33.

**AC33 — the asymmetry is the deliverable, and gets its own table test.** *"It is the single most
likely place for a silent off-by-representation bug."* Allocation is:

> read `nextId` (**decimal**) → render `"e" + base36(nextId)` → increment → write back **decimal**

Base-36 **decoding** is needed **only** to validate `nextId` against the highest id present — nowhere
else. A table test covers the round of counter→id→counter across the boundaries that matter (1, 9,
10, 35, 36, 37, 71, 72, and a large value), asserting the decimal/base-36 split holds in both
directions. It reports how many cases it exercised and fails on zero.

**AC34 — canonical id spelling is enforced at load, never normalised.** Ids must match
`^e[0-9a-z]+$`, decode to **≥ 1**, with **no leading zeros and no uppercase**. `e01` is a **load
error, not a value to normalise** — normalising would re-spell an id, and AD-10 says ids are *"never
renumbered on save"*. Rejecting is what makes ids round-trip **verbatim**, so P1 and P3 hold trivially
for the id field. Retained fixtures for: leading zero, uppercase, `e0`, `e` alone, non-alphanumeric.

**AC35 — a duplicate id anywhere in the document is a load error.** Across bands, elements and
columns alike — the counter is document-wide (AC36).

**AC36 — one document-wide counter.** **D-1.4.4**: *"`folio-format.md`'s `columns[]` row says each
column carries 'its own `id` (**same counter as elements**, so a diagnostic can name a column)', and
AD-10 requires every diagnostic concerning a template element to carry its id, which demands
document-wide uniqueness."* Not per element type.

**AC37 — `nextId` validation.** **D-1.4.4**: *"**`nextId` absent → load error. `nextId` ≤ the highest
id present → load error. Never repair, never renumber.**"* All polarities retained. The rationale goes
in the code comment verbatim: *"Delete the highest-id element and the derived counter **reuses that id
on the next allocation** — violating AD-10's 'never reused' silently and permanently."*

**AC38 — the guardrail test.** **D-1.4.4**: *"'Unchanged by a save' is **not** proved by a
load→serialize round-trip; that is P1. It is proved by: **load, delete an element, serialize** — then
assert every surviving id is byte-identical **and `nextId` is unchanged, not decremented**. A
serializer that renumbers, or recomputes `nextId` from the surviving maximum, **passes P1 and fails
this**. Retained fixture, both polarities."*

### Absent vs `null` vs present (`D-1.4.8` corrected, for AD-14 / Story 1.6)

**AC39 — the presence-flag route is forced.** The invariant is unchanged, per **D-1.4.8**:
*"**Distinguish absent from `null` from present-with-value** in the parsed representation … A loader
that collapses `null` into absent destroys the distinction AD-14's second case depends on (explicit
`null` = empty and **not** an error), and the same decode helpers will be reached for in 1.6."*

Per **`D-1.4.8` (corrected)**, following M-2: **pointer fields cannot satisfy it** — `*int` is nil for
both `{}` and `{"a":null}`. Use a **presence flag** (a wrapper carrying `set bool`) or
`json.RawMessage`, **with custom `UnmarshalJSON`, since only the unmarshaller knows it was called**.
A three-polarity fixture (absent / `null` / present) asserts all three stay distinguishable after
parse.

**AC40 — never coerce.** **D-1.4.8**: *"**Never coerce on load.** A string where a number belongs is
an error naming field, element id and value — never a parse-and-convert. This is AD-14's third case
arriving early, and it is also what makes P1 hold: **coercion is lossy, so a coerced value cannot
round-trip.**"*

**AC41 — errors are plain Go errors; mint no codes.** **D-1.4.2**: *"**1.4 must not mint them early**
— its load failures are plain Go errors naming field, element id and offending value."* A test
enumerates the story's load-error constructors and asserts each message carries field, element id
(where applicable) and value, reporting **how many it enumerated** and failing on zero.

> **Recorded, not a conflict.** The spine's Consistency Conventions say *"Go errors from the public
> API wrap `diag.Diagnostic`."* `internal/diag` does not exist yet; AD-14 lands with Story 1.6 and the
> codes with Story 3.6. D-1.4.2 rules the sequencing explicitly. AC48's tripwire converges it.

### The footer schema (D-1.4.1, D-1.4.2)

**AC42 — the two new fields.** `columns[]` gains optional `footerOf` and `footerFormat` beside
`footer`, per **D-1.4.1** (`footerOf`: *"a bare root-relative dotted value path
(`"transactions.amount"`). No `{{ }}`, no function call, no `[]`. The numeric source"*;
`footerFormat`: *"a `formatNumber` pattern applied to the computed value. Legal with all three
operations"*). Key order per **D-1.4.1**: *"**Serializer key order** (AD-9, byte-visible): `align`,
`bind`, `footer`, `footerFormat`, `footerOf`, `id`, `label`, `width`. `footerFormat` sorts *before*
`footerOf`."* **Measured** (M-4) — byte-order sorting produces exactly this.

**AC43 — the three decidable checks, and only those three.** Load errors in 1.4, per **D-1.4.2**:
- *"`footerOf` present with `footer: "count"` → load error"* — *"pure field presence"*;
- *"`footerOf` or `footerFormat` present with no `footer` → load error"* — *"pure field presence"*;
- *"`footerOf` prefixed by the table's collection path + `.`"* — *"a **string** prefix test against
  `bind` with `[]` stripped — no parser"*.

The two rows **D-1.4.2** marks **No** must not be implemented: *"`footerOf` omitted → derive from
`bind`, else load error"* (*"needs the parsed expression tree"*) and *"Footer uses the *same*
aggregate evaluation as `{{sum(...)}}`"* (*"nothing renders a table until 4.5"*).

**AC44 — the permissiveness guardrail.** **D-1.4.2**: *"The derivation deferral must not silently
become permanent permissiveness: until 3.2, a `footer` with no `footerOf` **loads**. That is a *known*
gap and needs its own fixture asserting today's behaviour, or Story 3.2's author will read the absence
of a check as evidence none was intended."* Retained fixture, comment naming Story 3.2 and D-1.4.2.

### The drift test (D-1.4.7)

**AC45** — **D-1.4.7**: *"`internal/template`'s parser and serializer are **normative**.
`folio-format.md` is documentation."* And: *"**But a readability contract that lies is worse than
none**, so drift is caught mechanically, in both directions: extract every JSON key the serializer can
emit (reflection over struct tags — exact on the Go side) and assert each appears as a backticked
token in `folio-format.md`; scan the document's backticked lowerCamelCase tokens and assert each is a
real key."*

**AC46 — the corpus shape, per the measurement (M-6).** The naive doc→Go scan is **not viable** — 67
backticked tokens, roughly half not keys. Use the measured **union**: field-table first cells
(comma-split — 34 tokens, **zero** false positives) **plus** code-fence keys (20 more).

**Exclude by category, never by name.** M-6's one leftover, `body`, is **a value, not a schema key**
— an author-chosen key inside `"fonts"`, whose keys are user-defined by design, exactly as `assets`'
SHA-256 keys are. So the rule is: **skip keys nested directly under `fonts` and `assets`.** That is a
statement about the schema's shape, so it will not rot when a second font stack appears in an
example. **A name-based exemption in a drift test is the same erosion pattern as a `hashmatrix`
exemption in a lint** — and it is the shape D-1.4.7 already forbids one step further on (*"a
checked-in list of key names"*).

**D-1.4.7**'s fallback stays named in the code comment in case the union proves brittle: *"narrow to
the Go→doc direction only … and record the doc→Go direction as a deferral with an owner."* And either
way: *"**Do not** fall back to a checked-in list of key names — that is a third artifact and
reintroduces exactly the drift this closes."*

**AC47 — the drift test's coverage witness, on both sides.** **D-1.4.7**: *"**This test is itself
exposed to D-000.9** … It must report **how many keys it compared**, and zero compared must fail."*

**And the extractors themselves need witnesses, not just the comparison.** The doc-side extractor
reports **how many tokens it extracted**, and **zero extracted fails**; the Go side reports a
**non-zero count of reflected struct tags**. A regex that silently stops matching — after the
document is reformatted, or a struct moves — yields zero tokens, zero comparisons and a green test.
**Without this it is the eighth D-000.9 instance, inside the test whose entire job is catching
drift.** Red-proved by construction against a retained fixture document, never by mutating the real
one. Its scope boundary against unknown passthrough keys is stated (AC10).

### The deferrals and their tripwires (D-1.4.2, D-1.3.4, D-1.4.11)

**AC48** — Three entries appended to `deferred-work.md`, in DW-2/3/4's shape:

| # | Deferred | Owner | Tripwire |
|---|---|---|---|
| DW-5 | Derivation validation of `footerOf` from `bind` | **Story 3.2**, backstop **3.7** (`folio.Validate` must include it) | asserts `folio-go/internal/expr` **absent** |
| DW-6 | Diagnostic codes `TABLE_FOOTER_SOURCE_UNRESOLVED` / `TABLE_FOOTER_SOURCE_FORBIDDEN` | **Story 3.6** | asserts `folio-go/internal/diag` **absent** |
| DW-7 | Footer evaluation sameness with `{{sum(...)}}` / `{{avg(...)}}` / `{{count(...)}}` | **Story 4.5** by name | **none possible** |

DW-7 records **D-1.4.2**'s own honesty: *"**no package tripwire exists**, so `deferred-work.md` is the
only trigger — flagged as the weakest of the three."*

**AC49** — The two tripwires join `absenceChecks` in `lint/internal/rules/absences.go`, **keyed on the
directory**, per DW-2's recorded correction (exact filenames were a Major finding). **D-1.4.3**'s
mechanism, from **D-1.3.4**: *"Story 1.3 ships an assertion that `folio-designer/`'s lockfile is
**absent**. The day Epic 5 creates it, that assertion goes **red** and forces the JS half to be wired
before the build can pass again."*

**AC50 — `ScanAbsences` gains a coverage witness, and this happens BEFORE AC49.** Per **D-1.4.11**:
`ScanAbsences` returns **how many absence checks it evaluated**, and **zero evaluated is a failure**.
The production caller fails on zero; a red-proof exercises the witness.

**The sequencing is ruled, not stylistic.** Closing this stays **task T2, before** the two tripwires
join the list: *"Adding entries to a list whose emptiness passes is precisely how a tripwire becomes
decorative: the tripwires would be 'present', and the check would pass **identically if they were
absent**."*

> **This is the seventh D-000.9 instance, and it is in the enforcement code shipped one story ago to
> police this very class.** D-000.9's own falsification test reads: *"**How we'd know it was wrong.** A
> seventh instance reaching `done`."* It has not reached `done` — it was caught at story creation, by
> the diagnostic question rather than by reading the guard. That is the mechanism working; it is also
> evidence the class is not yet closed.

Follow the sibling checkers' shape (`ForbiddenImportsStats`, `MapRangeStats`) and their stated reason,
verbatim from that code: *"reports what `ScanForbiddenImports` actually examined, **from the scanner's
own execution** … see `MapRangeStats`' doc comment for why a second, independently-derived walk cannot
be trusted as a vacuity guard."*

### The document amendments (D-000.6)

**AC51** — `_bmad-output/specs/spec-folio/folio-format.md` is amended in **this story's own commit**.
**Four** amendments now ship together (this story's finisher review, Finding 17, Minor: amendment D
below shipped in the developer's diff but was never enumerated here — the substance was fine, the
bookkeeping was not):

| # | Amendment | Ruling |
|---|---|---|
| A | Extend the `columns[].footer` row; add `footerOf` and `footerFormat` rows, stating the derivation, the `count` prohibition, and the two diagnostic codes | D-1.4.1 via **D-1.4.2** |
| B | Add the MINOR-additive safety sentence (AC11) | **D-1.4.9** |
| C | Rewrite the worked example to fully-expanded canonical form, cross-checked per AC16 | **D-1.4.10** |
| D | Append a provenance note after the worked example, naming the serializer, the independent cross-check and the golden fixture it must stay byte-identical to | **D-1.4.10**, AC16 obligation 3 |

**No amendment to the `id` row** — `folio-format.md:118` was right and stands (DN-2).

All three obey **D-000.6**: *"(1) change only the clause proven wrong, (2) ship in the implementing
story's commit, never a separate tidy-up, (3) be quoted verbatim in this log so the before/after is
readable without `git log`, and (4) leave the document's stated *intent* … untouched."* Obligation (3)
is the finisher's: the before/after of all three goes into the decision log.

Amendment A's derivation text, verbatim from **D-1.4.1**:

> When `footer` is present and `footerOf` is omitted, it is **derived** from the column's own `bind`,
> but only when `bind` is one of exactly two syntactic shapes:
> 1. a bare row-scoped path `{{<alias>.<rest>}}` → `footerOf` = `<collection>.<rest>`;
> 2. a single `formatNumber(<bare row-scoped path>, <pattern literal>)` call → `footerOf` =
>    `<collection>.<rest>` from the first argument, **and** `footerFormat` defaults to `<pattern>`.
>
> `<collection>` is the table's own `bind` with `[]` stripped. **Any other `bind` shape is a load
> error** — never a guess.

and the `count` prohibition: *"`footerOf` present alongside `footer: "count"` is a **load error**,
because storing it would be a second source of truth against `bind` — exactly what AD-13 forbids."*

**Note the split.** Amendment A documents a derivation 1.4 **does not implement** (AC43, AC48). That
is deliberate — **D-1.4.1**: *"Until that lands, every agent reading the contract between 1.4 and 4.5
infers the aggregation source itself, differently each time."* The amended prose must be unambiguous
about what is documented-but-not-yet-enforced, or it becomes the readability contract that lies.

### Gates

**AC52** — On the default gate (`go build ./...`, `go vet ./...`, `gofmt -l` empty, `go test ./...`
across `folio-go`, `hashmatrix`, `lint`, plus `go build -tags=matrix` / `go vet -tags=matrix`):
everything green, every new guard's witness non-zero, and **every red-proof running in that gate**.
**D-000.9** obligation 2: *"The red is produced by construction, in a gate that actually runs."*

DW-1's recorded history is the precedent to avoid, verbatim: *"the developer's first pass placed [the
red-proof] in `folio-go/matrix_test.go`, which carries the `//go:build matrix` tag — so the red-proof
executed in **zero** gates … DW-1 was marked DONE on that strength anyway, which was the defect."*

**AC53** — The cross-target matrix is **not** run. **D-000.4**: it runs *"once per epic, after that
epic's last story is committed"*, and 1.4 is not on the override list. Story 1.4 produces no PDF
bytes. The Delivery Log *"name[s] the suites it actually measured **and** name[s] the unrun ones
explicitly."*

**AC54 — no unwarranted citations.** **D-1.2.5**: *"When code, a comment, or a workflow cites `F-N` or
a Delivery Log measurement as its justification, the finisher **verifies that `F-N` actually claims to
have been measured**."* Every measurement cited must be reproducible by the command named beside it.

---

## Decisions resolved

All three surfaced conflicts are ruled. The reasoning is carried because **two reversed a prior
ruling and one is an owner override** — a developer who does not know why will re-derive the losing
option.

### DN-2 → `D-1.4.4` (reversed in part)

**Ruled: base 36.** `folio-format.md:118` stands verbatim: *"`id` | `e` + the counter in **lowercase
base 36** — `e1`, `ea`, `e1z`."* D-1.4.4's *"not base-36"* clause is reversed; **all its other clauses
stand.**

**How the error happened, recorded so it does not recur.** The lead had skipped `folio-format.md`
lines 60–150 — exactly where the normative field table sits — and overturned its own correct note
using the worked example; the orchestrator then "verified" that using the same ambiguous evidence.
**M-7** is what exposed it: the worked example's ids `e1`…`e5` read identically under decimal and base
36, because the two alphabets first diverge at counter 10 and no id in the example reaches 10.

### The transferable rule

> **If the other hypothesis predicts the same observation, you have learned nothing.**

And its spec-reading corollary, which is what actually went wrong here:

> **In a spec, the field table is normative; the worked example is frequently ambiguous by
> construction** — examples are written small, and small examples are exactly where competing
> encodings agree.

Apply it to every *"verified in <document>"* claim: find the **discriminating case** and check it is
present in the cited sample. This is D-000.9's mechanism — *"the success signal is indistinguishable
from the not-run signal"* — applied to a citation rather than to a guard.

### DN-1 → D-1.4.10

**Ruled: amend the worked example (option a). Option (b) was never available.**

AD-9's Rule fixes *"two-space indent"*; **an inline nested container has no indentation at all**, so a
serializer emitting them implements a *different rule*. D-000.6 classifies changing what an invariant
**is** as a direction change requiring the owner — so (b) was not a cheaper alternative to 41 lines of
markdown, it was a different decision needing different approval. (a) sits squarely inside D-000.6: it
changes only the clause proven wrong (**whitespace**) and leaves the intent intact. The document
supports this itself (`:55–56`): *"Top-level keys appear sorted, as does every object in the file —
that is the **serializer's job** (AD-9), not something an author maintains by hand."*

The self-ratification guardrail this creates is **AC16**, and it is mandatory.

### DN-3 → D-1.4.9

**Owner decision: full passthrough (option 3), overriding the lead's recommended option 4.**

A higher MINOR loads; unknown keys are carried opaquely, merged back sorted, ignored at render;
nothing dropped, nothing refused. Higher MAJOR is still a load error.

**The reasoning, carried because it explains the cost.** The lead recommended
load-read-only-refuse-to-save as the cheaper honest minimum. The owner overrode it: **option 4's cost
lands on the designer user in Epic 5** — the person least equipped to understand why a file that
visibly opens cannot be saved. *Forward compatibility that works until you touch it is a promise users
discover is hollow at the worst moment.* Option 3 pays its cost **once, in engine code, now**, while
the schema is being frozen and the round-trip properties are being built. It is also the only option
under which the round-trip guarantee is **strictly stronger**: a canonical `1.1` file processed by a
`1.0` library still satisfies `Serialize(Parse(b)) == b`.

**Accepted cost, stated plainly:** the most machinery of the four options, landing in the one story
whose job is to establish P1, P2, P3 and the drift test.

**Unsafe without AC11's sentence** — see AC11.

### The M-2 corrections

Both accepted, logged as `D-1.4.8` (corrected) and `D-1.4.3` (extended): the presence-flag route is
forced (AC39); the `omitempty` rule becomes two-sided (AC22, AC23); AC20's Thai/CJK claim is narrowed
to match what was measured.

### M-5 → D-1.4.11

`ScanAbsences` gains a coverage witness; the T2-before-tripwires sequencing is confirmed correct.
See AC50.

---

## Tasks / Subtasks

Ordered so each measurement and each guard exists before the code depending on it. **No commit
task** — this story ends at status → `review`.

- [x] **T1 — Read and confirm the ground.** `folio-format.md` in full (**including lines 60–150 — the
normative field table; DN-2 is what skipping it cost**), AD-9/AD-10 and the adjacent ADs, and
D-1.4.1–D-1.4.11 + D-000.9. *(all)*

- [x] **T2 — Close the absence guard's vacuity hole. Before anything else touches `absenceChecks`.**
Witness from the checker's own execution; production caller fails on zero; red-proof. *(AC50)*

- [x] **T3 — `internal/template`: the model.** Struct types for document, page, fonts, bands, elements
(five types), table, columns, style, assets. `geom.Length` for points; `int64` for `nextId`; `string`
for `version`. No `omitempty` where a zero is legal; **every map and slice initialised empty, never
nil**. Presence-flag wrappers for absent/`null`/present — **not bare pointers** (M-2). A passthrough
store for unknown keys at every level. *(AC22, AC23, AC24, AC39, AC8)*

- [x] **T4 — The decimal path.** Exact ×1000 decimal-string arithmetic `json.Number` → `geom.Length` and
back to the `appendLength` spelling. Never `float64`, never `ParseFloat`. Overflow is an error; >3
decimals is an error. *(AC24–AC28, AC30)*

- [x] **T5 — Parser.** `json.Decoder` + `UseNumber()`. Closed-set validation (AC5), kept visibly separate
from the passthrough path. No coercion. Load errors carry field, element id, value. *(AC5, AC26,
AC40, AC41)*

- [x] **T6 — Version handling.** MAJOR gate; MINOR loads; `version` carried verbatim — **never lowered,
never gratuitously raised** — behind a single well-named function carrying AC7's reframe in its doc
comment. The raise path is unreachable today: do not build it, do not foreclose it. Two tests pin the
rule (`1.0`→`1.0`; synthetic `1.1`→`1.1` with unknown keys intact). *(AC6, AC7)*

- [x] **T7 — Unknown-key passthrough.** Opaque capture at every nesting level; merged back in **sorted
order** with known keys; never read by any render-path consumer. *(AC8, AC9, AC10)*

- [x] **T8 — Ids: the base-36/decimal asymmetry, with its own table test.** Render base 36 from a decimal
`nextId`; decode base 36 **only** to validate `nextId` against the highest id. Canonical spelling
enforced, never normalised; duplicates rejected. *(AC31–AC35)*

- [x] **T9 — `nextId` and the counter.** Document-wide; columns share it; absent and too-low are errors;
never repair, never renumber. *(AC36, AC37)*

- [x] **T10 — Serializer.** Sorted keys (byte-order, asserted as such, including passthrough keys),
two-space indent, LF, no trailing whitespace, trailing newline, HTML escaping off, literal UTF-8,
minimal escaping, `/` unescaped. *(AC18–AC21)*

- [x] **T11 — Shell API.** `LoadTemplate` and `ParseTemplate` in package `folio`; `LoadTemplate` delegates.
`Render(t *Template)`. Provisional marker updated. *(AC1, AC2, AC3)*

- [x] **T12 — Footer schema fields and the three decidable checks**, plus the permissiveness fixture.
*(AC42, AC43, AC44)*

- [x] **T13 — Amend `folio-format.md` (three amendments), with the cross-check.** A and B are prose. **C is
the guarded one:** rewrite the worked example expanded, cross-check against an independent
pretty-printer for structure/indentation/key order, enumerate and normalise every divergence class,
**name each normalisation in the Delivery Log**, and ship the byte-identity test between the fenced
block and the golden fixture. *(AC16, AC51, AC11)*

- [x] **T14 — The property tests and the fixture corpus.** P1, P2, P3. Delete-an-element guardrail, both
polarities. Unknown keys at multiple nesting levels. Every trap fixture. All retained, all in the
default gate. *(AC12–AC17, AC38, AC52)*

- [x] **T15 — Numeric classification test**, count witness, red-proof fixture. *(AC29)*

- [x] **T16 — Drift test**, union corpus per M-6, count witness, red-proof fixture document, passthrough
scope boundary stated. If the fallback is taken, record it in `deferred-work.md` with an owner.
*(AC45, AC46, AC47)*

- [x] **T17 — Tripwires and deferrals.** Two directory-keyed absence checks (after T2); three
`deferred-work.md` entries with DW-7's weakness stated. *(AC48, AC49)*

- [x] **T18 — Run the gates and write the Delivery Log.** Name every suite measured and every suite not
run, with reasons. **Name every normalisation applied in T13.** No carried-forward unit counts.
*(AC52, AC53, AC54)*

- [x] **T19 — Set status to `review`.** Do not commit, do not branch, do not set `done`.

---

## Dev Notes

### Measurements

Every figure was produced on this machine at baseline `af6b5d5` (Go 1.26.0, darwin/arm64) and is
reproducible by the command named. Per **D-1.2.5**, nothing here is inherited from a finding that
disclaims measurement.

**M-1 — the worked example is not a fixed point (→ D-1.4.10).** `folio-format.md:233–295` decoded with
`UseNumber`, re-encoded with `SetEscapeHTML(false)` + `SetIndent("", "  ")`: **1648 B / 63 lines →
2096 B / 104 lines**, not byte-identical, first divergence at **byte 181**. **7 nested containers**
are kept inline by the document (`:242`, `:243`, `:247`, `:259`, `:274`, `:285`, `:289`) which the
encoder expands. Keys were already correctly sorted — bytes 0–180 match exactly; the sort was never
the problem, the inline style was. (`"assets": {}` at `:234` is **not** a divergence: an empty object
stays `{}` inline in both forms.)

**M-2 — `encoding/json` behaviours the traps depend on (→ `D-1.4.3` extended, `D-1.4.8` corrected).**

| Probe | Result |
|---|---|
| `json.Marshal` default, on `a<b>c&d/e "q" ท 中` | **`<`, `>`, `&` escaped** — AC19's trap is live |
| Thai / CJK, **either** setting | emitted **literally**; no `\uXXXX`. AC20 is a real requirement but **not a default hazard** |
| `/` | never escaped, either setting |
| `Encoder.Encode` | appends `\n` — AD-9's trailing newline comes from the encoder, not manual concatenation |
| `omitempty` on empty map + zero int | both **dropped**: `{}`. D-1.4.3's trap, exactly |
| **nil** map, `omitempty` removed | emits **`null`** — not `{}`, not absent. The fifth trap; makes the rule two-sided |
| `*int`, inputs `{}` vs `{"a":null}` | **both nil** — indistinguishable. Falsifies "pointer fields" for D-1.4.8(1) |
| `UseNumber` on `36.0000`, `1e3`, `72.5001`, `9223372036854775807` | literals preserved verbatim — normalisation is the story's own job |
| `json.Number("9223372036854775808").Int64()` | `value out of range` — overflow detectable (AC27) |
| `{"n":"5"}` into `int64` | `cannot unmarshal string into … type int64` — the decoder does not coerce; the hazard is in the story's own `json.Number` conversion |

**M-3 — the numeric inventory measures exhaustive.** Every JSON key given a numeric literal anywhere
in `folio-format.md` — fences and prose — is **11 distinct keys**: `bottom`, `fontSize`,
`headerHeight`, `height`, `left`, `nextId`, `right`, `top`, `width`, `x`, `y`. D-1.4.3's inventory
covers all eleven with **zero escapees**, upgrading its stated **medium** confidence. **AC29 still
ships** — the document will grow.

**M-4 —** D-1.4.1's stated column key order is exactly byte-order; `footerFormat` precedes `footerOf`
because `F`(0x46) < `O`(0x4F).

**M-5 — a live D-000.9 exposure at baseline (→ D-1.4.11).** Of the three checkers in
`lint/internal/rules`, `ScanForbiddenImports` returns `(findings, ForbiddenImportsStats, error)` and
its production test fails on `stats.FilesSeen == 0`; the map-range checker has `MapRangeStats`
likewise. `ScanAbsences` returns `(findings, error)` only, and `TestAbsencesProductionScan` asserts
nothing but `len(findings) == 0`. **Empty the `absenceChecks` slice and the test passes.** D-000.9's
question — *"What would this check have printed if it had been unable to run at all?"* — answers *"zero
findings"*, identical to its healthy output. **Seventh instance of the class, inside the enforcement
code shipped one story ago to police it.**

**M-6 — the doc→Go drift corpus.** Whole-document backticked lowerCamelCase tokens: **67**, roughly
half not keys (`sum`, `count`, `avg`, `center`, `middle`, `en`, `th`, `ja`, `text`, `image`, `table`,
`line`, `rect`, `row`, `false`, `float64`, `e`, `e1`, `ea`, `e1z`, `addendum.md`, …). Field-table first
cells, comma-split: **34**, **zero** false positives. Code-fence keys: **20** more (`body`, `bottom`,
`color`, `columns`, `content`, `data`, `edges`, `elements`, `footer`, `label`, `left`, `margin`,
`mediaType`, `orientation`, `pageFooter`, `pageHeader`, `right`, `size`, `top`, `value`), of which
`body` is **a value, not a key** — an author-chosen key inside `"fonts"`. It is excluded **by
category** (skip keys nested directly under `fonts` and `assets`, both user-keyed by design), never
by name — see AC46. Union is viable; the naive whole-file scan is not.

**M-7 — the catch that reversed D-1.4.4.** `e1`…`e5` read as decimal → `[1,2,3,4,5]`; read as base 36
→ `[1,2,3,4,5]`. The alphabets first differ at counter 10; no id in the worked example reaches 10. The
cited verification could not discriminate, and the document's field table said base 36 in words. See
[§ The transferable rule](#the-transferable-rule).

**M-8 — working-tree state, re-measured at revision 0.2.** `folio-mvp-decision-log.md` is
**uncommitted** at `af6b5d5`, now **+459 lines**: D-000.9 and D-1.4.2–D-1.4.8 (present at revision
0.1, +250 lines), plus the rulings that resolved this story's three blocks — `D-1.4.9` (:2102),
`D-1.4.4 (reversed in part)` (:2177), `D-1.4.10` (:2216), `D-1.4.8 (corrected)` /
`D-1.4.3 (extended)` (:2264), `D-1.4.11` (:2285). **Every ruling this story quotes lives in the
working tree, not in `HEAD`** — re-read them there, and expect them to arrive in a commit that is not
this story's. No other file differs.

**M-9 — the AC16 cross-check is viable, and here are its divergence classes.** The worked example was
round-tripped through **two independent implementations** — Go's `encoding/json` with
`SetIndent("", "  ")`, and Python's `json.dumps(sort_keys=True, indent=2, ensure_ascii=False)`. With
scalar values normalised out, **structure, indentation and key order are byte-identical: 104 lines
each, zero differing lines.** The independent pretty-printer therefore confirms the expanded shape
rather than merely echoing Go.

Divergence classes to enumerate and normalise per AC16 step 2 — the developer confirms this list is
complete for the amended file rather than assuming it:

| # | Class | Detail |
|---|---|---|
| 1 | **Number spelling** | the only substantive one. Python re-renders through its own float/int handling; Go through its. **Neither is authoritative** — the story's own exact-decimal path (AC25) is. Compare structure with scalars normalised out, then verify numbers separately against the `appendLength` spelling |
| 2 | **Trailing newline** | `json.dumps` adds none; `Encoder.Encode` adds one. AD-9 requires one |
| 3 | **Escaping** | Python needs `ensure_ascii=False` **explicitly** — its default escapes Thai to `ท`, which would silently "confirm" a wrong file. Go needs `SetEscapeHTML(false)` |
| 4 | **Empty containers** | both emit `{}` / `[]` inline. No divergence; noted so it is not mistaken for one |

**The trap I hit and discarded, recorded because it is the guardrail's own failure mode.** My first
cross-check ran the Python side with `parse_float=str, parse_int=str`, which renders every number as
a quoted string. That **conflated number spelling with structure** and produced **25 differing
lines** that looked like real structural disagreement. Left unnoticed, the natural next move is to
"normalise" those 25 differences until the two sides agree — which is **exactly the self-ratification
AC16 exists to prevent**, arriving through the check rather than through the document. The correct
run normalises scalars out of *both* sides and compares structure alone; it reported **zero**
differing lines. **Rule: a normalisation that makes a disagreement disappear must be justified by
what it removes (a variable), never by that it worked.**

### The transferable rule

> **If the other hypothesis predicts the same observation, you have learned nothing.**
>
> **In a spec, the field table is normative; the worked example is frequently ambiguous by
> construction.**

This is why DN-2 was wrong twice before it was caught, and it generalises past ids: any *"verified in
<document>"* claim needs its **discriminating case** located in the cited sample. It is D-000.9's
mechanism — *"the success signal is indistinguishable from the not-run signal"* — applied to citations
rather than guards.

### Where things go

| Thing | Where | Why |
|---|---|---|
| `LoadTemplate`, `ParseTemplate`, `Render` | `folio-go/folio.go` (package `folio`, module root) | the shell — D-1.4.6; the spine annotates `folio.go` as *"LoadTemplate · Validate · Render · RenderTo — the shell"* |
| parser, serializer, validation, `Template` internals, passthrough store | `folio-go/internal/template/` | AD-9: *"`internal/template` owns both the parser and the serializer"* |
| round-trip, trap and id fixtures | `folio-go/testdata/template/` | D-1.3.3's fixture convention; AD-21's repo-root `fixtures/` is for **golden render** fixtures |
| the amended worked example as a golden fixture | per AD-21's convention, with AC16's byte-identity test binding it to the document | D-1.4.10 |
| absence tripwires | `lint/internal/rules/absences.go` | D-1.3.4's mechanism, already built |
| deferrals | `_bmad-output/implementation-artifacts/deferred-work.md` | append-only, DW-2/3/4 shape |
| the three amendments | `_bmad-output/specs/spec-folio/folio-format.md` | D-000.6, in this story's commit |

### Things that look wrong later but are not

- **A `footer` with no `footerOf` loads.** Deliberate (AC44, D-1.4.2), fixture-pinned, comment naming
  Story 3.2.
- **Load errors are plain Go errors, not `Diagnostic` codes.** Deliberate (AC41, D-1.4.2); the
  `internal/diag` tripwire converges it at Story 3.6.
- **`nextId` is decimal while ids are base 36.** Deliberate and ruled (AC32); the asymmetry has its own
  table test precisely because it looks like a bug.
- **`e01` is rejected rather than normalised to `e1`.** Deliberate (AC34): normalising re-spells an id,
  and AD-10 forbids renumbering on save.
- **A `1.1` file loads in a `1.0` library and saves as `1.1`.** Deliberate and owner-ruled (AC7,
  D-1.4.9). The mirror is equally deliberate: a future `1.1` library must save an unchanged `1.0`
  file as `1.0`. *"`version` describes the document, not the writer."*
- **Adding a sixth element type or a fifth locale requires a MAJOR bump.** Deliberate and
  owner-confirmed (AC11, D-1.4.12) — passthrough carries unknown *keys*, not new values in known
  enums. It is permanently expensive on purpose.
- **`appendLength`'s logic appears twice, unshared.** Deliberate (AC25): `internal/template` may not
  import `internal/pdf`, and sharing would put a template concern into the PDF emitter. The
  cross-check test is the seam.
- **No table renders, no aggregate is computed.** Nothing renders a table until Story 4.5.
- **The cross-target matrix is not run.** D-000.4, per-epic cadence; 1.4 is not on the override list.

### Things that must not happen

- Pasting the story's own serializer output into `folio-format.md` without the independent cross-check
  (**AC16** — this is the story's own self-ratification hazard).
- Applying a normalisation in T13 without naming it in the Delivery Log (**AC16**).
- Adding tripwires to `absenceChecks` before T2 closes the witness (**AC50**, **D-1.4.11**).
- Paraphrasing a ruling into an AC and widening it (**D-1.1.b**).
- A red-proof behind a build tag, or applied-and-reverted (**D-000.9** obligation 2; DW-1's defect).
- A guard reporting zero findings without reporting how many candidates it considered (**D-000.9**).
- Dropping, refusing, or rewriting an unknown key (**D-1.4.9**).
- Building AD-23's exact scaled-decimal type (**D-1.4.3**: *"that type is Story 1.6's and 1.4 must not
  build it"*).
- Minting `TABLE_FOOTER_SOURCE_UNRESOLVED` or `TABLE_FOOTER_SOURCE_FORBIDDEN` (**D-1.4.2**).
- Falling back to a checked-in list of key names for the drift test (**D-1.4.7**).
- Committing, branching, or setting `done`.

### Source acceptance criteria

`_bmad-output/planning-artifacts/epics.md` §Story 1.4, verbatim:

> **Given** a well-formed `.folio` file carrying a `version`, page setup, and ordered band content
> **When** `folio.LoadTemplate(path)` is called
> **Then** a `Template` is returned with no error
>
> **Given** a template declaring a format version newer than the library supports
> **When** it is loaded
> **Then** loading fails with an error naming the declared version and the supported version
> **And** no render is attempted
>
> **Given** any parsed template `d` and any canonical byte sequence `b`
> **When** round-trip properties are tested
> **Then** `Parse(Serialize(d)) == d` and `Serialize(Parse(b)) == b` hold
> **And** these tests ship with the format, not after it
>
> **Given** a template serialized by the library
> **When** the bytes are inspected
> **Then** object keys are sorted, indentation is two spaces, line endings are LF, there is no
> trailing whitespace, and the file ends with a newline
>
> **Given** any element in a loaded template
> **When** its identity is inspected
> **Then** it carries an opaque short id from a monotonic counter persisted in the document
> **And** ids are not UUIDs, are never reused, and are unchanged by a save

> **Note on the second epic AC.** *"a format version newer than the library supports"* is now split by
> **D-1.4.9**: a higher **MAJOR** satisfies it (AC6); a higher **MINOR** is explicitly **not** a
> failure case (AC7). The epic text predates that ruling and is not contradicted by it — MINOR
> versions are supported, so a higher MINOR is not "newer than the library supports".

---

## Dev Agent Record

### Delivery Log

**Architecture actually built.** `internal/template`'s parser and serializer do NOT use
`encoding/json` struct tags at all — the model (`model.go`) is decoded by hand from
`map[string]json.RawMessage` (`parse.go`, `parse_bands.go`) and serialized by hand as a sorted
`(key, value-writer)` sequence at every object level (`serialize.go`, `jsonstring.go`). This was a
deliberate choice, not a deviation: it is the only way to satisfy AC8's "one sorted key sequence, not
a known block followed by an unknown block" — `encoding/json`'s struct-tag `Marshal` cannot merge a
struct's own fields with a dynamic passthrough map in one sorted pass. Two consequences worth
recording because they change where later ACs' mechanisms live: (1) AC22's "no omitempty" guard is
satisfied by construction (there are no struct tags to carry `omitempty` in the first place) —
`TestNoOmitemptyStructTags` asserts that invariant by reflection so a future edit cannot silently
reintroduce tag-driven serialization; (2) AC45's drift test extracts the Go side by parsing
`serialize.go`'s own AST for the string keys passed to `kv{...}` literals, not by reflecting over
struct tags (D-1.4.7's originally-envisioned mechanism) — functionally equivalent ("a key that stops
being written is a key that stops appearing in source"), documented in `drift_test.go`'s header
comment.

**Gates run, all commands reproducible as given, raw exit codes.**

- `cd folio-go && go build ./...` — exit 0
- `cd folio-go && go vet ./...` — exit 0
- `cd folio-go && gofmt -l .` — exit 0, empty output
- `cd folio-go && go test ./... -v` — exit 0. Counts (re-measured on this run, not carried
  forward): **69 top-level tests, 137 total including subtests, 5 packages**
  (`folio`, `internal`, `internal/geom`, `internal/pdf`, `internal/template` — `internal/template` is
  new this story). The story's stated reference baseline ("26 top-level / 51 with subtests / 4
  packages") was the pre-1.4 count; it does not describe this story's own suite.
- `cd folio-go && go build -tags=matrix ./...` — exit 0
- `cd folio-go && go vet -tags=matrix ./...` — exit 0
- `cd lint && go build ./...` — exit 0
- `cd lint && go vet ./...` — exit 0
- `cd lint && gofmt -l .` — exit 0, empty output
- `cd lint && GOPROXY=off go test ./...` — exit 0 (15 top-level tests across
  `licence`/`manifest`/`rules`/`genmanifest`, zero failures)
- `cd hashmatrix && go vet ./...` — exit 0
- `cd hashmatrix && go build -o "$TMP/probe" ./probe` — exit 0
- `cd hashmatrix && go build ./...` (plain, NOT the correct gate) — exit 1, confirmed as expected by
  design (pre-existing "build output probe already exists and is a directory" — unrelated to this
  story, not a regression; the correct gate for this module is the two commands above it)

**Suites named but NOT run, per D-000.4 (this is not an override story):** the cross-target
byte-identity matrix (`folio-go`'s `matrix_test.go` cross-process/cross-target legs beyond the
default `go test ./...` run — i.e. `TestTargetRenderHash` and friends, which need CI's per-target
build environment). AC53 confirms 1.4 is not on the per-epic override list and this story produces no
PDF bytes of its own; the matrix runs once, at Epic 1's close.

**Real bugs the tests caught during this story (worth recording — none were pre-existing).**
1. `decodeNumberRaw` initially unmarshalled into `json.Number` directly; measured live that
   `json.Unmarshal(`\"0\"`, &n)` succeeds with **no error** for a `json.Number` target — `json.Number`
   is `type Number string` under the hood, so `encoding/json`'s ordinary string decoding silently
   accepts a quoted JSON string. This is exactly the AC40 coercion hazard ("never a parse-and-convert")
   and was caught by `TestNeverCoerceOnLoad`, not by inspection. Fixed by rejecting a leading `"` byte
   before the value ever reaches `json.Number`'s unmarshaller (`decodehelpers.go`).
2. The drift test's first AST pass (`extractGoKeys`) only matched explicitly-typed `kv{...}` composite
   literals; most of this file's `kv` literals are elided-type elements inside a `[]kv{ {...}, {...} }`
   slice literal (Go elides the repeated type name). The first version silently extracted a small
   subset of the real key set and `TestDriftGoToDoc` passed vacuously (zero missing, because almost
   nothing was compared) — caught by adding the tokens-considered/keys-compared coverage witnesses
   (D-000.9 shape) and noticing they were suspiciously low, then confirmed by `TestDriftDocToGo` failing
   loudly once the extraction was widened. Fixed by also matching the elided-type two-unkeyed-element
   shape.
3. `lint`'s pre-existing `ScanMapRange` guard (D-1.3.5/NFR1.d, shipped in Story 1.3) caught **nine**
   real map-range violations across this story's new code on the very first `go test ./...` run in
   `lint/` — every one was a `for k := range someMap` or `for k, v := range someMap` written before I
   had internalised the repo's "ranging a map anywhere under `internal/` is a hard, total ban"
   convention. All nine fixed to the required
   `for _, k := range slices.Sorted(maps.Keys(m)) { v := m[k]; … }` idiom
   (`decodehelpers.go`, `parse.go` ×3, `parse_bands.go` ×2, `rawvalue.go`, `serialize.go` ×2). This is
   exactly what that guard exists to catch, and it worked.

**AC16's independent cross-check, reproduced live (not merely cited from M-9).** The amended worked
example (`folio-format.md`'s fenced block, byte-identical to
`folio-go/testdata/template/golden/worked-example.json`) was produced by this story's OWN
`ParseDocument`/`SerializeDocument` round-tripping the pre-amendment worked example, THEN
cross-checked against Python's independent pretty-printer:
`python3 -c "import json; d=json.load(open('worked_example_raw.json')); print(json.dumps(d,
sort_keys=True, indent=2, ensure_ascii=False))"` — `diff` against the Go-produced canonical bytes
reports **zero differing lines** (both 104 lines, 2096 bytes). This matches M-9's recorded
measurement exactly and was independently reproduced during this story rather than trusted from the
story text alone (D-1.2.5).

Divergence classes named (per AC16 step 2 — all four of M-9's classes were checked against the
amended file and confirmed complete, none new found): (1) number spelling — not applicable to this
particular worked example, since every number in it is already an integral point value under both
Go's and this story's own exact-decimal path; (2) trailing newline — `Encoder.Encode`/this story's own
serializer both append one, Python's `json.dumps` does not (accounted for separately, not part of the
diffed comparison); (3) escaping — Python's run used `ensure_ascii=False` explicitly, defensively;
**finisher's correction (Finding 10, Major):** the amended worked example contains **zero characters
above U+007F** — its `locale` is `"th"` but no Thai script actually appears in it — so
`ensure_ascii=True` and `ensure_ascii=False` produce **identical output** for this file, and this
class removed no variable in practice, contrary to what was originally claimed here. The escaping
dimension this class intended to cover is instead genuinely exercised by `utf8TrapFixture` and
`htmlEscapeTrapFixture` under P3 (AC19, AC20), which do carry Thai/CJK and `<`/`>`/`&` respectively;
Go's serializer has HTML-escaping off and literal UTF-8 by construction regardless. (4) empty
containers — both sides emit `{}` inline for `"assets": {}`, no divergence.

**Three `folio-format.md` amendments, before/after (D-000.6 obligation 3).**

- **Amendment A** (`columns[].footer` extended; `footerOf`/`footerFormat` rows added). Before: the
  table listed only `columns[].footer` (`sum`/`count`/`avg`, no derivation text) and
  `altRowBackground`. After: two new rows for `footerOf` and `footerFormat` were inserted between them,
  carrying D-1.4.1's derivation rule verbatim, the `count` prohibition, and a note naming
  `TABLE_FOOTER_SOURCE_UNRESOLVED`/`TABLE_FOOTER_SOURCE_FORBIDDEN` as Story 3.6's, not yet minted.
- **Amendment B** (the MINOR-additive safety sentence, AC11). Before: the Document table's "Top-level
  keys appear sorted…" sentence stood alone. After: AC11's full three-clause blockquote was added
  immediately beneath it, verbatim.
- **Amendment C** (worked example rewritten to fully-expanded canonical form). Before: 1648 bytes / 63
  lines, with 7 nested containers kept inline (M-1). After: 2096 bytes / 104 lines, matching this
  story's own serializer output and the independent Python cross-check byte-for-byte, shipped as
  `folio-go/testdata/template/golden/worked-example.json` with `TestWorkedExampleMatchesGoldenFixture`
  and Amendment D's provenance note (see [§ Finding Resolutions](#finding-resolutions), Finding 17).

**Finisher's correction to Amendment A's own description above (Finding 9, Major).** "Extended"
overstated what shipped: the `columns[].footer` row itself was byte-identical to `HEAD` — only the
two new rows (`footerOf`, `footerFormat`) were inserted, so the row AC51 and D-000.6 scheduled for
amendment was in fact unchanged, though the *substance* D-1.4.1 wanted did ship, inside the new rows.
Fixed in this pass: `folio-format.md`'s `columns[].footer` row now reads *"**Unchanged — names the
operation only** (D-1.4.1); the numeric source is `columns[].footerOf`, below."* — the specific
sentence D-1.4.1's verdict table names (*"`footer` | enum `"sum"` | `"count"` | `"avg"` | **unchanged**
— names the *operation only*. Absent = no footer cell"*), appended to the row rather than replacing it
(D-000.6 obligation 4: leave the row's stated intent untouched).
  binding the two together.

No `id` row amendment — `folio-format.md:118`'s base-36 wording stands unchanged (DN-2).

**Ruling conflicts encountered during implementation:** none. All thirteen rulings quoted in the story
applied directly; no new decision or ruling conflict arose that required parking (D-1.2.6).

**Deferred, per the story's own scope boundary (not new deferrals — recording what was and was not
built):** `columns[].footerOf` derivation from `bind` (DW-5, Story 3.2), the two footer diagnostic
codes (DW-6, Story 3.6), footer/`{{sum(...)}}` evaluation sameness (DW-7, Story 4.5, no tripwire
possible) — all three appended to `deferred-work.md` in DW-2/3/4's shape, with DW-5/DW-6's tripwires
(`folio-go/internal/expr` and `folio-go/internal/diag` asserted absent) added to
`lint/internal/rules/absences.go`'s `absenceChecks` AFTER `ScanAbsences`'s own coverage-witness gap
was closed (T2, `AbsencesStats.ChecksEvaluated`), per AC50/D-1.4.11's load-bearing ordering.

### Finisher validation (this pass, D-000.11: every gate below runs `-count=1`)

Every command below was run with raw exit codes checked directly (`echo $?`), never through a
wrapper that could print a friendly summary over a non-zero exit, per this pass's own instruction.

- `cd folio-go && go build ./...` — exit 0
- `cd folio-go && go vet ./...` — exit 0
- `cd folio-go && gofmt -l .` — exit 0, empty output
- `cd folio-go && go test -count=1 ./...` — exit 0. **75 top-level tests, 161 total including
  subtests, 5 packages, 0 SKIPs** (re-measured after this pass's fixes; the reviewer's own
  independently re-measured baseline was 69/137/5 — the +6 top-level / +24-with-subtests delta is
  exactly this pass's new tests: `TestDriftASTMatchesRuntimeEmission`,
  `TestExtractGoKeysMissesAKeyedLiteral`, `TestPassthroughAtEveryObjectLevel` (11 subtests),
  `TestOmitemptyTypeListIsComplete`, `TestNextIDAbsentIsLoadError`, `TestRenderIgnoresItsArgumentToday`,
  plus the `serializes-differently` subtest added to `TestPresenceThreePolarity` and the `maximal`/
  `null-field` subtests added to the round-trip property tests).
- `cd folio-go && go build -tags=matrix ./...` — exit 0
- `cd folio-go && go vet -tags=matrix ./...` — exit 0
- `cd lint && go build ./...` — exit 0
- `cd lint && go vet ./...` — exit 0
- `cd lint && gofmt -l .` — exit 0, empty output
- `cd lint && GOPROXY=off go test -count=1 ./...` — exit 0. **16 top-level tests, 37 total including
  subtests** (reviewer's baseline: 15/36; +1 top-level is
  `TestAbsencesChecksIncludeAllFourEntries`).
- `cd hashmatrix && go vet ./...` — exit 0
- `cd hashmatrix && go build -o "$tmp/probe" ./probe` — exit 0
- `cd hashmatrix && gofmt -l .` — exit 0, empty output
- `cd hashmatrix && go build ./...` (plain, NOT the correct gate) — exit 1, confirmed as expected by
  design (pre-existing "build output probe already exists and is a directory," unrelated to this
  story — same as the developer's own measurement; the two commands above it are the correct gate)

**Cross-target matrix: NOT run**, named explicitly (AC53, D-000.4) — it runs once per epic after that
epic's last story is committed, and 1.4 is not on the override list.

**Both blockers' red-proofs, reproduced live in this pass** (method and results in
[§ Finding Resolutions](#finding-resolutions) above): Finding 1's keyed-literal injection into
`serialize.go`, reverted; Finding 2's `decodePage` refusal reverted temporarily, then restored. Both
diffed clean against a pre-mutation backup before being trusted, and the full suite was re-run green
after each restoration.

### Debug Log References

None — no debugger or ad hoc scripts survived into the delivered state; the AC16 cross-check's Python
one-liner is recorded above rather than left as a stray script.

### Completion Notes

All 54 ACs and all 19 tasks (T1–T19) implemented and verified; see the Delivery Log above for gate
results, the AC16 cross-check reproduction, and the three real bugs the test suite caught. `internal/
template` is a new package (parser + serializer for `.folio`, AD-9) with no `encoding/json` struct-tag
dependency by design (see Delivery Log's "Architecture actually built"). `folio.go`'s `LoadTemplate`,
`ParseTemplate` and `Render(t *Template)` are the new module-root shell (AC1–AC3); `Render` accepts but
does not yet consume its `*Template` argument (documented as PROVISIONAL, unchanged behaviour —
Story 1.1's fixed rectangle still ships). Every existing `Render()` call site
(`render_test.go`, `fixture_test.go`) was updated to `Render(nil)` — a mechanical signature-only
change, since `Render` does not read its argument yet. `lint/internal/rules/absences.go` gained a
coverage witness (`AbsencesStats`) and two new directory-keyed tripwires. `folio-format.md` carries its
three D-000.6 amendments in this commit. `deferred-work.md` gained DW-5/DW-6/DW-7.

Note on the epic's second source AC (quoted in `§ Source acceptance criteria`): as flagged in the story
text, "a format version newer than the library supports" is now split by D-1.4.9 into a MAJOR failure
case (AC6) and an explicitly-supported MINOR case (AC7) — this is expected, not a contradiction, and
both halves are tested.

### File List

**Created:**
- `folio-go/internal/template/model.go`
- `folio-go/internal/template/presence.go`
- `folio-go/internal/template/rawvalue.go`
- `folio-go/internal/template/jsonstring.go`
- `folio-go/internal/template/decimal.go`
- `folio-go/internal/template/ids.go`
- `folio-go/internal/template/version.go`
- `folio-go/internal/template/errors.go`
- `folio-go/internal/template/closedsets.go`
- `folio-go/internal/template/decodehelpers.go`
- `folio-go/internal/template/parse.go`
- `folio-go/internal/template/parse_bands.go`
- `folio-go/internal/template/serialize.go`
- `folio-go/internal/template/testutil_test.go`
- `folio-go/internal/template/roundtrip_test.go`
- `folio-go/internal/template/fixtures_test.go`
- `folio-go/internal/template/ids_test.go`
- `folio-go/internal/template/version_test.go`
- `folio-go/internal/template/footer_test.go`
- `folio-go/internal/template/presence_test.go`
- `folio-go/internal/template/coercion_test.go`
- `folio-go/internal/template/errors_test.go`
- `folio-go/internal/template/omitempty_test.go`
- `folio-go/internal/template/numeric_classification_test.go`
- `folio-go/internal/template/decimal_test.go`
- `folio-go/internal/template/length_spelling_test.go`
- `folio-go/internal/template/drift_test.go`
- `folio-go/internal/template/goldenfixture_test.go`
- `folio-go/internal/pdf/length_spelling_shared_test.go`
- `folio-go/testdata/template/golden/worked-example.json`
- ~~`folio-go/testdata/template/length-spelling-cases.txt`~~ **moved by the finisher** to
  `folio-go/testdata/shared/length-spelling-cases.txt` (D-1.4.14, Finding 13)
- `folio-go/template_test.go`

**Modified:**
- `folio-go/folio.go`
- `folio-go/render_test.go`
- `folio-go/fixture_test.go`
- `lint/internal/rules/absences.go`
- `lint/internal/rules/absences_test.go`
- `_bmad-output/specs/spec-folio/folio-format.md`
- `_bmad-output/implementation-artifacts/deferred-work.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (status value only)
- `_bmad-output/implementation-artifacts/1-4-load-validate-and-round-trip-a-folio-template.md` (this
  file)

**Finisher pass — additional files created:**
- `folio-go/internal/template/passthrough_test.go`
- `folio-go/testdata/shared/length-spelling-cases.txt` (moved from `testdata/template/`, see above)

**Finisher pass — additional files modified** (all findings-driven; see
[§ Finding Resolutions](#finding-resolutions) for which finding drove which change):
- `folio-go/internal/template/model.go`, `parse.go`, `parse_bands.go`, `serialize.go` (Finding 2)
- `folio-go/internal/template/drift_test.go` (Findings 1, 12)
- `folio-go/internal/template/fixtures_test.go`, `roundtrip_test.go` (Findings 6, 8, 11)
- `folio-go/internal/template/numeric_classification_test.go` (Finding 3)
- `folio-go/internal/template/presence_test.go` (Finding 6)
- `folio-go/internal/template/omitempty_test.go` (Finding 22)
- `folio-go/internal/template/ids_test.go`, `errors_test.go` (Finding 19)
- `folio-go/internal/template/length_spelling_test.go` (Finding 13)
- `folio-go/internal/pdf/length_spelling_shared_test.go` (Finding 13)
- `folio-go/folio.go` (Finding 15), `folio-go/template_test.go` (Findings 7, 19)
- `lint/internal/rules/absences_test.go` (Finding 5)
- `.github/workflows/ci.yml` (Finding 4 / D-000.11)
- `_bmad-output/specs/spec-folio/folio-format.md` (Findings 9, 16)
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` (Finding 18)

---

## Change Log

| Date | Version | Description | Author |
|---|---|---|---|
| 2026-08-23 | 0.1 | Story drafted from epic ACs + D-1.4.1–D-1.4.8, D-000.9. Three `DECISION NEEDED` blocks opened (worked-example canonicity, id alphabet, MINOR/unknown keys). Eight measurements recorded. | bmad-story-creator |
| 2026-08-23 | 0.2 | All three decisions ruled and folded in. **DN-2 → `D-1.4.4` reversed in part: ids are base 36** (M-7); asymmetry, canonical spelling and duplicate rules added as AC31–AC38 with their own task. **DN-1 → D-1.4.10:** worked example amended, with a mandatory independent cross-check (AC16) and a doc↔fixture byte-identity test. **DN-3 → D-1.4.9 (owner):** full unknown-key passthrough, higher MINOR loads, `version` verbatim (confirmation-pending), mandatory MINOR-additive sentence — AC6–AC11. M-2 corrections ruled into `D-1.4.8` (corrected) and `D-1.4.3` (extended); AC20 narrowed. M-5 → **D-1.4.11**, seventh D-000.9 instance, T2-first sequencing confirmed. M-6 union corpus adopted. M-9 added. ACs 43 → 54; tasks 16 → 19. | bmad-story-creator |
| 2026-08-23 | 0.3 | **D-1.4.13:** `version` verbatim confirmed — flag removed, and the rule made **symmetric** (never lowered, never gratuitously raised; raise path unreachable today, not foreclosed), with the *"describes the document, not the writer"* reframe required in the doc comment and two pinning tests. **D-1.4.12 (owner):** AC11's sentence extended to its full three-clause form — extending any of the eight closed sets is a **MAJOR** change; rejected alternatives and the C1/C2 side effect recorded. **AC46:** `body` exclusion changed from a name to a **category** (skip keys directly under `fonts`/`assets`); **AC47** gains extractor witnesses on both sides (zero tokens extracted fails) — otherwise the eighth D-000.9 instance. M-9 records the `parse_float=str` conflated-variables trap. | bmad-story-creator |
| 2026-08-23 | 0.4 | Implemented. All 54 ACs, all 19 tasks. New `internal/template` package (hand-written parser/serializer, no `encoding/json` struct tags — see Delivery Log); `folio.go` gains `LoadTemplate`/`ParseTemplate`/`Render(t *Template)`; `lint/internal/rules/absences.go` gains `AbsencesStats` and two directory-keyed tripwires (T2 before T17, per D-1.4.11); three `folio-format.md` amendments shipped with before/after in the Delivery Log; `deferred-work.md` gained DW-5/DW-6/DW-7. AC16's independent Python cross-check reproduced live (zero differing lines). Three real bugs caught by the test suite during implementation, recorded in the Delivery Log (a `json.Number` coercion hole, a vacuous drift-test AST match, nine pre-existing-guard map-range violations). All gates green with raw exit codes; cross-target matrix named unrun (AC53, D-000.4). Status → `review`. | bmad-story-developer |
| 2026-08-23 | 0.5 | Finished. Both blockers fixed and red-proved live: Finding 1 (drift test's AST-only extraction, D-1.4.15) closed with a second, behavioural set-equality extraction plus a new `maximalFixture`; Finding 2 (D-1.4.9 OWNER passthrough refused at 6 of 11 object levels) closed at 5 of those 6 (`page`, `page.margin`, `style.padding`, `style.border`, `assets[entry]`) — `bands` DISMISSed as a deliberate, AC5/FR6-backed exception, not a defect. All 8 majors and 11 minors FIXed except one documentation-only correction (Finding 10); 2 nits DISMISSed (D-1.4.16 rules Finding 23 explicitly not a defect; Finding 24 is a no-op style nit the reviewer offered no resolution for); 1 nit (25) superseded by this pass's own opener rewrite. Four late-landing rulings (D-1.4.14, D-1.4.15, D-1.4.16, D-000.10) folded into the header and disposed of; D-000.11 (`-count=1` in every gate) applied to `.github/workflows/ci.yml`. Zero findings DEFERred. All gates re-run with `-count=1`: 75/161/5 (`folio-go`), 16/37 (`lint`), zero SKIPs, zero regressions. Status → `done`. | bmad-story-finisher |

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-23
- **Baseline:** `af6b5d5`, all changes uncommitted (working-tree review)
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 2 · **Majors:** 8 · **Minors:** 13 · **Nits:** 4

**Gates re-measured independently (raw exit codes, no wrapper).** `folio-go`: `go build ./...` 0 · `go vet ./...` 0 · `gofmt -l .` 0/empty · `go test -count=1 ./...` **0, 69 top-level / 137 with subtests / 5 packages** — matches the Delivery Log exactly · `go build -tags=matrix ./...` 0 · `go vet -tags=matrix ./...` 0. `lint`: build 0, vet 0, gofmt empty, `GOPROXY=off go test -count=1 ./...` 0 (**15 top-level / 36 with subtests** — matches). `hashmatrix`: `go vet ./...` 0, `go build -o "$tmp/probe" ./probe` 0, plain `go build ./...` 1 (by design, as recorded). Zero test SKIPs. **No count discrepancy.** Cross-target matrix correctly unrun (AC53, D-000.4).

**Review method.** Every mechanism below was probed by mutation, not by reading. All mutations were applied to a `cp`-backed copy and reverted by copying back; final `diff -r` against the pre-review backup of `folio-go/`, `lint/` and `_bmad-output/specs/` reports **0 differing lines**, `git status` is identical to session start, and the full suite is green. **Nothing was fixed.**

**What is genuinely strong, stated so the finder list is not read as the whole picture.** Closed-set enforcement is complete and correct across all eight sets plus every structural rule (table `x`/`y`-only, `bands` three-key shape, `height` forbidden on `content`), each error naming field, element id and value. The numeric path is exact: `72.5001`, `1e300` and int64 overflow are load errors; `36.0000`, `1e3`, `72.501` normalise; base-36 id overflow is caught. The AC40 coercion fix is **complete, not patched at one site** — I swept nine field/type combinations and every one is refused. AC38's delete-an-element guardrail **genuinely reddens** against a renumbering serializer (I built one; both polarities failed). AC16's Python cross-check **reproduces exactly** (104 lines each side, 2096 bytes, zero differing lines). The doc-side drift extractor is precise (51 keys, `body` excluded structurally, not by name).

---

### Finding 1: The drift test can miss an emittable key by construction — neither direction catches it

- **Severity**: **Blocker**
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/internal/template/drift_test.go:34-79` (`extractGoKeys`), `:30` (`goKeyRegistrySource`)
- **Related AC**: AC45, AC47 · **Rulings**: D-1.4.7, **D-1.4.15**, D-000.9

**Observation.** `extractGoKeys` recognises exactly two source shapes: an explicitly-typed `kv{"lit", ...}` and an elided-type two-element composite literal whose first element is a `*ast.BasicLit`. Any other route to an emitted key is invisible. Proven by mutation, three ways:

- **Control (sanity).** Adding `{"zzUndocumentedNatural", writeString("x")}` to `writeDocument` in the natural shape → `TestDriftGoToDoc` **FAILS** as designed. The test works for the shape it models.
- **Probe A — keyed composite literal.** `kv{key: "zzUndocumentedKeyed", write: writeString("x")}` (`Elts[0]` is an `*ast.KeyValueExpr`, not a `BasicLit`) → **both drift tests stay GREEN**. The key *is* emitted — `TestP3FixturesAreCanonical` failed, proving emission.
- **Probe B — named constant.** Refactoring `{"nextId", ...}` to `{keyNextID, ...}` with `const keyNextID = "nextId"` → `nextId` vanishes from the Go-side set entirely. (`TestDriftDocToGo` caught this one *only* because the doc still documents `nextId`.)
- **Probe C — the full blocker.** A keyed literal inside a branch no fixture exercises (`if st.Italic.Set && st.Bold.Set && st.FontSize.Set`) → **the entire suite is GREEN with an emittable, undocumented key present in the serializer.** Neither drift direction, nor P1/P2/P3, notices.

Note the developer's stated equivalence argument in the file header — *"a key that stops being written is a key that stops appearing in serialize.go's source, full stop"* — is true for **removal** and false for **addition by another route**. That asymmetry is the whole defect.

**Impact.** The mechanism whose only job is preventing `folio-format.md` from silently drifting can be bypassed by three ordinary Go spellings, one of which (hoisting key literals to named constants) is a routine refactor a future maintainer would make without a second thought. D-000.9's question — *"what would this have printed if it had been unable to run at all?"* — answers "green", identically. This extractor has **already been vacuous once** in this story (Delivery Log bug #2); this is the same class, not yet closed.

**This is now independently ruled.** **D-1.4.15** (decision log `:2446`, orchestrator decision, landed in the same uncommitted diff and **not referenced by the story**) states the finding in advance and mandates the fix: *"AST extraction … is exact **iff every emitted key passes through that convention.** A key emitted by any other route … is invisible to it. **The failure is asymmetric in the bad direction:** if such a key is *also* undocumented, **neither** direction of the drift test notices."*

**Suggested Resolution.** Implement D-1.4.15's binding requirement: a **second, behavioural extraction** — serialize a document exercising every known field, parse the output bytes, collect every key actually present, and **assert `astKeys == runtimeKeys` in both directions**, printing the difference by name. Keep coverage witnesses on both extractors. Do not fix this by widening the AST shape list — that chases spellings forever and leaves the next one open.

---

### Finding 2: Unknown-key passthrough is refused at six of eleven object levels, contradicting D-1.4.9 (OWNER)

- **Severity**: **Blocker**
- **Category**: AC Conformance / Correctness
- **Location**: `folio-go/internal/template/parse.go`, `folio-go/internal/template/parse_bands.go` (the `unrecognised field on …` / `must carry exactly …` / `no others` refusal sites)
- **Related AC**: AC8, AC9, AC11 · **Rulings**: **D-1.4.9 (OWNER)**, D-1.4.12 (OWNER)

**Observation.** I injected one unknown key `"zzNew"` at every object level in a single otherwise-valid document and checked accept-and-preserve:

| Level | Result |
|---|---|
| top level | accepted + preserved |
| band (`bands.content`) | accepted + preserved |
| element | accepted + preserved |
| column | accepted + preserved |
| `style` | accepted + preserved |
| **`bands`** | **REFUSED** — `field bands: must have exactly the three keys content, pageFooter, pageHeader — no others (FR6)` |
| **`page`** | **REFUSED** — `field page: unrecognised field on page` |
| **`page.margin`** | **REFUSED** — `must carry exactly top, right, bottom, left` |
| **`style.padding`** | **REFUSED** — `unrecognised field on padding` |
| **`style.border`** | **REFUSED** — `unrecognised field on border` |
| **`assets[entry]`** | **REFUSED** — `must carry exactly data and mediaType` |

The five levels that work are **exactly the five AC9 names as fixture sites** ("top level, inside a band, inside an element, inside a column, inside `style`"). The implementation was built to AC9's fixture list rather than to D-1.4.9's rule.

**Impact.** D-1.4.9 is an **owner decision** whose operative sentence is *"unknown keys are carried opaquely through parse, merged back in sorted order on serialize, and ignored when rendering. **Nothing is dropped, nothing is refused.**"* Six of eleven object levels refuse. AC11's shipped, owner-confirmed sentence declares *"A MINOR increment may add **new optional keys** only"* — so a legitimate `1.1` file adding, say, `page.bleed` is **valid by the format's own normative rule and refused by this library**. That is precisely the failure D-1.4.9 bought its "most machinery of the four options" to prevent, and it lands on the Epic 5 designer user the owner overrode the lead to protect.

Note this is **not** authorised by D-1.4.12: its eight closed sets are all sets of legal **values** (`type`, `locale`, `align`, `valign`, `columns[].footer`, `border.edges`, `page.orientation`, `page.size`). None is a **key** set. `page`, `margin`, `padding`, `border`, `assets`-entry and `bands` key-closure is an implementation choice with no ruling behind it. The story's own guard sentence — *"an unrecognised **key** is carried through, while a recognised key carrying an unlisted **value** is still a load error"* — is the correct rule and is not what shipped.

**Suggested Resolution.** Add a passthrough `Extra` store to `Page`, `Margin`, `Padding`, `Border` and `Asset`, and to the `bands` container object, mirroring the five levels that already work; keep the *value* validation on known keys exactly as it is. If any of these six is genuinely intended to be key-closed (`bands` has the strongest case via FR6), that needs a ruling and an AC, not a silent refusal — and AC9's fixture must then be extended to pin the intended boundary in both polarities.

---

### Finding 3: The AC29 numeric-classification inventory never reads `folio-format.md`, so it cannot see the document grow

- **Severity**: Major
- **Category**: Tests (D-000.9)
- **Location**: `folio-go/internal/template/numeric_classification_test.go:48-72`
- **Related AC**: AC29

**Observation.** `TestNumericFieldClassificationInventory` builds its `documented` list as a **hardcoded 11-element slice** and checks it against `numericFieldRegistry`, which contains exactly those same 11 keys. The assertion is implied by its own setup. The test never opens `folio-format.md` (the four grep hits in the file are comment prose). Proven: I added a twelfth backticked numeric key (`lineSpacing`, points, value `1.5`) to `folio-format.md`'s field table and the test **stayed green**.

**Impact.** AC29's stated purpose is exactly this case: *"the developer must enumerate every numeric field in `folio-format.md` … with a test asserting **no field escaped classification**"*, kept despite M-3 measuring the inventory exhaustive because *"the document will grow, and a test that was true once is not a coverage witness."* As shipped it **is** a test that was true once. The `considered`/zero-fails witness counts a list the test itself wrote, so it reports healthy while measuring nothing about the document.

**Suggested Resolution.** Derive `documented` from `folio-format.md` at test time — the drift test's `extractDocKeys` already extracts the document's key set structurally; intersect it with keys carrying a numeric literal, or reuse the M-3 extraction rule. Keep the count witness on the *extracted* list. The red-proof (`TestNumericFieldClassificationCatchesAnEscapee`) is fine and should stay.

---

### Finding 4: The default gate caches away every guard that reads a file outside its own module

- **Severity**: Major
- **Category**: Tests (D-000.9) / Convention
- **Location**: `folio-go/internal/template/drift_test.go`, `folio-go/internal/template/goldenfixture_test.go`, `lint/internal/rules/absences.go`
- **Related AC**: AC52, AC47, AC16, AC49

**Observation.** Go's test cache does not track files read from outside the module. Demonstrated end to end: prime the cache with a clean `go test ./internal/template/`; inject **real drift** into `folio-format.md` (a backticked `zzPhantomKey` the serializer cannot emit); re-run the *plain* command → **`ok … (cached)`**. The same run with `-count=1` → **FAIL**. The subagent verification reproduced the identical behaviour in `lint`: after creating `folio-go/internal/expr/`, the first `-count`-less run reported `ok (cached)` and the tripwire fired only under `-count=1`.

Affected guards: `TestDriftGoToDoc`, `TestDriftDocToGo`, `TestWorkedExampleMatchesGoldenFixture`, and both new absence tripwires. (`length-spelling-cases.txt` lives inside the module and caches correctly.)

**Impact.** AC52 names the gate as `go test ./...` — no `-count=1`. So on any change that touches only `folio-format.md`, or only the repo directory layout, the guards that exist specifically to police those two things report green **without executing**. That is D-000.9's canonical shape — the success signal is indistinguishable from the not-run signal — sitting on top of four of this story's guards at once, including the drift test AC47 already flags as *"itself exposed to D-000.9"*.

**Suggested Resolution.** Either add `-count=1` to the recorded default gate for the packages carrying cross-module guards, or make each such test declare its external input to the cache (e.g. read it through a path the test binary also stats via `os.ReadFile` under a `testdata` symlink is *not* sufficient — the reliable fix is `-count=1` in the gate). Record whichever is chosen in the story's gate list so it is reproducible.

---

### Finding 5: The two new absence tripwires are decorative in the shrunk-list sense D-1.4.11 warned about

- **Severity**: Major
- **Category**: Tests (D-000.9)
- **Location**: `lint/internal/rules/absences.go:54-70`, `lint/internal/rules/absences_test.go`
- **Related AC**: AC49, AC50 · **Ruling**: D-1.4.11

**Observation.** The witness itself is **correctly built** — `stats.ChecksEvaluated++` is the first statement inside `ScanAbsences`' own loop (`absences.go:108`), derived from the scanner's execution exactly as `ForbiddenImportsStats`/`MapRangeStats` do, and both vacuity probes redden: neutering `ScanAbsences` to `return nil, AbsencesStats{}, nil` → FAIL, and emptying `absenceChecks` → FAIL. Both tripwires also genuinely fire (creating `folio-go/internal/expr/` reddens `TestAbsencesProductionScan`, naming the directory), and both are **keyed on the directory**, not on filenames — DW-2's recorded correction is respected.

**But:** deleting *only* the two new entries and leaving the rest → **suite green, EXIT 0**. Nothing pins their presence. `ChecksEvaluated` would be 2, non-zero, so the witness passes identically.

**Impact.** D-1.4.11's ruling reads *"they'd be 'present' and the check would pass **identically if they were absent**"* — that condition is still true for these two specific entries. T2 closed the **empty-list** case; the **silently-shrunk-list** case is open, and it is the one that actually happens (someone deletes a tripwire while chasing an unrelated red). The story records this AC as satisfied and calls it the seventh D-000.9 instance closed; it is closed for one of its two shapes.

**Suggested Resolution.** Assert the specific rule ids (`absence-expr-package`, `absence-diag-package`) are present in `absenceChecks`, or assert `ChecksEvaluated` equals an expected count that a deletion would break. Also note: `TestAbsencesZeroWitnessIsCaught` mutates the package-level `absenceChecks` var with `t.Cleanup` restore — safe today (no `t.Parallel()` in the package) but a shared-mutable-state coupling worth a comment.

---

### Finding 6: No test proves absent and explicit `null` serialize differently — D-1.4.8's distinction is unpinned on the write side

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/internal/template/presence_test.go` (parse side only); no canonical fixture contains `null`
- **Related AC**: AC39 · **Rulings**: D-1.4.8 (corrected), **D-1.4.16**

**Observation.** `TestPresenceThreePolarity` asserts all three states are distinguishable **after parse** and does so correctly. Nothing asserts they survive **serialization**. Grepping the whole corpus, the only `: null` occurrences in the repository are inside that one parse test — **no canonical fixture carries an explicit `null`**, so P1/P2/P3 never exercise the null path either.

I verified the behaviour is currently **correct**: absent omits the key, `"style": null` re-emits `null`, present-with-value re-emits the object, all three are serializer fixed points, and all three differ. So this is an unpinned invariant, not a broken one.

**Impact.** D-1.4.16 (decision log `:2530`, orchestrator decision, same uncommitted diff, **not referenced by the story**) rules precisely this: *"a three-state API exists whose third state no production code reads. That is a D-000.9 shape, and the guard is behavioural: **a test must prove `{}` and `{"a":null}` produce different `Presence` values *and* serialize back differently.** … **That test is what makes D-1.4.8 real rather than declared.**"* Without it, a future edit collapsing `Null` into absent goes unnoticed and Story 1.6 inherits a seam that was never load-bearing.

**Suggested Resolution.** Extend `TestPresenceThreePolarity` with a serialization leg asserting the three states produce three distinct byte outputs, and add an explicit-`null` fixture to `canonicalFixtures` so P1/P2/P3 cover it as a fixed point.

---

### Finding 7: No test asserts `Render(validTemplate)` and `Render(nil)` currently produce identical bytes

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/folio.go:52-57`, `folio-go/template_test.go:96-108` (`TestRenderAcceptsATemplate`)
- **Related AC**: AC2 · **Rulings**: D-1.1.c, **D-1.4.16**

**Observation.** The `PROVISIONAL` doc comment is **present and correct** — it states the parameter is accepted and not yet consumed, why, and which story consumes it. That half of the brief's question is satisfied: it is documented, not merely true. But every call site passes `Render(nil)` (`render_test.go` ×4, `fixture_test.go` ×1), and the one test passing a real template asserts only `len(b) != 0`. Removing the parameter entirely would leave every assertion in the suite valid.

**Impact.** D-1.4.16 rules that documentation alone is insufficient here: *"**'documented as provisional' decays into prose nobody re-reads.** Do both … **and a named test asserts `Render(validTemplate)` and `Render(nil)` currently produce identical bytes.** The moment Story 1.5 or 1.6 makes them differ, that test fails and **forces its own deletion**."* As shipped, the day `Render` starts reading `t`, nothing announces that the provisional note is now stale.

**Suggested Resolution.** Add the named test asserting byte-equality of `Render(tpl)` and `Render(nil)`, with a comment stating it is designed to fail and be deleted when Story 1.5/1.6 lands.

---

### Finding 8: Twelve of fifty-one emittable keys are never exercised by the round-trip corpus

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/internal/template/roundtrip_test.go:15-60` (`canonicalFixtures`)
- **Related AC**: AC12, AC13, AC14, AC15 · **Ruling**: **D-1.4.15**

**Observation.** Running D-1.4.15's mandated "cheap path first" measurement — serialize every corpus fixture, re-parse the bytes, collect every key actually present, compare to the AST key set:

- AST keys the serializer can emit: **51** (from 68 `kv` literals)
- Distinct keys appearing anywhere in the serialized corpus: **46**
- **AST keys never exercised by the round-trip corpus: 12** — `altRowBackground`, `asset`, `background`, `color`, `data`, `fontFamily`, `footerFormat`, `footerOf`, `italic`, `mediaType`, `valign`, `visibleIf`

**Impact.** D-1.4.15 predicts this exact result and rules it a finding in its own right: *"Check whether the **union of the existing round-trip corpus** already covers every key in the AST set. … **If it does not, that gap is itself a finding** — the round-trip corpus would not be exercising every field, independent of the drift test."* P1, P2 and P3 — the story's stated core — currently hold over 39 of the serializer's 51 fields. Notably `assets` (`data`, `mediaType`) never round-trips at all despite `writeAssets` being a non-trivial nested emitter, and `visibleIf`/`background` are the two `Null`-capable fields, which ties directly to Finding 6.

**Suggested Resolution.** Add one maximal canonical fixture exercising every known field, and use it as the runtime side of Finding 1's set-equality assertion — one artifact closes both.

---

### Finding 9: Amendment A is half-shipped — the `columns[].footer` row was never extended

- **Severity**: Major
- **Category**: AC Conformance / Convention
- **Location**: `_bmad-output/specs/spec-folio/folio-format.md:166`
- **Related AC**: AC51 · **Rulings**: D-1.4.1 via D-1.4.2, D-000.6

**Observation.** The working-tree `columns[].footer` row is **byte-identical to `HEAD`** — it appears in the diff as unchanged context, not as a `-`/`+` pair. Only the two new rows (`footerOf`, `footerFormat`) were inserted. Three artifacts say otherwise: AC51's amendment table (*"**Extend** the `columns[].footer` row; add `footerOf` and `footerFormat` rows"*), the Delivery Log's own headline (*"Amendment A (`columns[].footer` extended; …)"* — which its following prose then contradicts by describing only the two inserted rows), and D-000.6 `:619`, which names *"**The `columns[].footer` field-table row** (`folio-format.md`) — ships with **Story 1.4**"*. D-1.4.1 `:481` specifies the extension's substance: `footer` names the *operation only*. That wording is nowhere in the shipped row.

**Impact.** The substance D-1.4.1 wanted (derivation rule, `count` prohibition, both diagnostic codes) *did* ship — inside the new `footerOf` row. So the document is not wrong, but the row D-000.6 explicitly scheduled for amendment is unchanged, and the Delivery Log records it as done. A finisher checking obligation (1) against the diff will find a claim with no hunk behind it.

**Suggested Resolution.** Either extend the `footer` row as D-1.4.1 specifies (`footer` names the operation only; the source is `footerOf`), or correct AC51 and the Delivery Log to state that Amendment A ships as two added rows with no change to `footer`.

---

### Finding 10: The AC16 cross-check's named escaping normalisation removes no variable — the worked example contains zero non-ASCII bytes

- **Severity**: Major
- **Category**: AC Conformance
- **Location**: story Delivery Log, "Divergence classes named", class (3); M-9 class 3
- **Related AC**: AC16 · **Ruling**: D-1.4.10

**Observation.** The Delivery Log states class 3 as: *"escaping — Python's run used `ensure_ascii=False` explicitly (**its default would have escaped the worked example's Thai text** to `\uXXXX`, which would have silently 'confirmed' a wrong file)"*. Measured: the golden worked example — byte-identical to the document's fence — contains **zero characters above U+007F**. `json.dumps(..., ensure_ascii=True)` and `ensure_ascii=False` produce **identical output** for this file. Its `locale` is `"th"` but no Thai script appears anywhere in it. (Confirmed separately: the file contains no `\u` sequence and no `<`, `>` or `&` either.)

**Impact.** AC16 step 2 requires each normalisation to be *"justified by **what variable it removes**, never by that it worked"* — the story's own stated rule, written up in M-9 immediately after the `parse_float=str` trap. This normalisation removes no variable in this file, and its justification names content that does not exist. The cross-check's structure/indentation/key-order result is **real and I reproduced it** (104 lines each, zero differing lines), so the guardrail did its actual job; but one of the four enumerated classes is inert and recorded as load-bearing, which is the reporting failure AC16 exists to catch, arriving through the check rather than the document.

**Suggested Resolution.** Correct class (3) to state that no non-ASCII appears in the amended worked example, so `ensure_ascii=False` was set defensively and was **not** exercised — and note that the escaping dimension is instead covered by `utf8TrapFixture` and `htmlEscapeTrapFixture` under P3. If the escaping dimension is meant to be cross-checked, the worked example needs non-ASCII content.

---

### Finding 11: AC18's non-ASCII passthrough key has no discriminating fixture

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/template/fixtures_test.go`; no test file references key sorting
- **Related AC**: AC18

**Observation.** No fixture anywhere in the repo contains a non-ASCII **key** (`unknownKeysFixture`'s two non-ASCII characters are inside a passthrough *value*), and no test file mentions byte-order or key sorting. Sorting *is* correct by construction (`sort.Slice` on Go `string` `<`, and `slices.Sorted(maps.Keys(...))` in `rawvalue.go` — both byte-order), and I verified behaviourally that a document with keys `"zz"` and `"ä"` round-trips byte-identically with `zz` first, which is byte-order and not a collation order.

**Impact.** AC18 says the requirement *"is no longer only theoretical"* precisely because passthrough keys may be non-ASCII, and asks for it to be asserted rather than to hold *"by coincidence over an all-ASCII corpus"*. The corpus is all-ASCII on the key side. This is the story's own transferable rule — locate the discriminating case in the cited sample — applied to its own fixture set.

**Suggested Resolution.** Add a non-ASCII passthrough key to `unknownKeysFixture` (`"zz"` before `"ä"` discriminates byte-order from collation in one pair).

---

### Finding 12: The drift test's passthrough scope boundary is not stated, though two ACs require it

- **Severity**: Minor
- **Category**: Convention / AC Conformance
- **Location**: `folio-go/internal/template/drift_test.go:13-27` (header comment)
- **Related AC**: AC10, AC47

**Observation.** AC10 requires: *"Unknown passthrough keys are explicitly **outside its scope**. State this in the drift test's doc comment, *'or a future maintainer reads a passthrough key as drift.'*"* AC47 repeats it. The header comment discusses AST-vs-struct-tags at length and `extractGoKeys`' comment covers the `fonts`/`assets` structural exclusion, but **no comment in the file mentions passthrough or unknown keys at all**.

**Impact.** Low today; the named hazard is real later — the moment `folio-format.md` documents a key this library models as passthrough, `TestDriftDocToGo` reddens with no explanation of why that is or is not drift.

**Suggested Resolution.** Add the sentence AC10 specifies to the file header.

---

### Finding 13: The shared length-spelling table lives under one consumer's directory

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/testdata/template/length-spelling-cases.txt`, read by `folio-go/internal/pdf/length_spelling_shared_test.go:24`
- **Ruling**: **D-1.4.14**

**Observation.** The closure is **real** — I verified both sides genuinely read the same file (no forked copy exists), both count cases and fail on zero, and the table covers trailing-zero trimming, negatives, the sub-1-point negative case, and the three-decimal boundary. But `internal/pdf`'s test reaches into `../../testdata/template/` — another package's directory.

**Impact.** D-1.4.14 (decision log `:2404`, **not referenced by the story**) calls this out and rules it: *"the table is *shared* but lives under a path owned by one consumer … **Move it to a neutral location (`folio-go/testdata/shared/`)**, or a future reader in `internal/pdf` will find their test reading another package's directory and 'tidy' it. Finisher-sized change."* D-1.4.14's own falsification test is *"a second copy of the length-spelling table"* — which is exactly what tidying produces.

**Suggested Resolution.** Move the table to `folio-go/testdata/shared/` and update both readers.

---

### Finding 14: The shared spelling table omits the int64 extremes

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/testdata/template/length-spelling-cases.txt`
- **Related AC**: AC25

**Observation.** Largest magnitude in the table is `1000000` (1000 pt). `appendPlainInt` uses a fixed `var buf [20]byte` and `appendPoints` negates the integer part — both are places where two independent reimplementations could diverge, and `MaxInt64`/`MinInt64` millipoints are reachable (a literal `9223372036854775.807` loads; `9223372036854775808` is correctly rejected).

**Suggested Resolution.** Add `9223372036854775807`, `-9223372036854775808` and a value with a zero integer part and non-zero fraction at large scale to the table.

---

### Finding 15: The `PROVISIONAL` marker's stated range contradicts AC2

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: `folio-go/folio.go:41`
- **Related AC**: AC2 · **Ruling**: D-1.1.c

**Observation.** The shipped marker reads `PROVISIONAL — Story 1.1 through Story 1.6.` AC2 states *"the signature is still provisional through 1.7 (**D-1.1.c**)"*, and the comment's own body correctly describes Story 1.7's `RenderTo` arriving without breaking the call.

**Suggested Resolution.** Change the marker range to `Story 1.1 through Story 1.7`.

---

### Finding 16: The shipped AC11 sentence carries an extra trailing sentence beyond the verbatim text

- **Severity**: Minor
- **Category**: Convention (D-000.6 / D-1.1.b)
- **Location**: `_bmad-output/specs/spec-folio/folio-format.md:58-63`
- **Related AC**: AC11 · **Ruling**: D-1.4.12 (OWNER)

**Observation.** The required three-clause sentence ships **byte-identical as a prefix**, and **all eight closed sets are named** in the correct order (`type`, `locale`, `align`, `valign`, `columns[].footer`, `border.edges`, `page.orientation`, `page.size`) with all three clauses intact. Appended after it is a sentence AC11 and D-1.4.12's verdict blockquote do not contain: *"A higher MINOR still loads (`version` is carried verbatim); a higher MAJOR is always a load error."*

**Impact.** The addition restates D-1.4.9/D-1.4.13 correctly and contradicts nothing, so this is a fidelity issue rather than a correctness one — but the text was specified verbatim by an owner ruling and shipped extended, into a normative document.

**Suggested Resolution.** Either drop the trailing sentence or record it in the decision log as a deliberate addition to D-1.4.12's text.

---

### Finding 17: A fourth, unenumerated edit ships in `folio-format.md`

- **Severity**: Minor
- **Category**: Convention
- **Location**: `_bmad-output/specs/spec-folio/folio-format.md:348`
- **Related AC**: AC51 · **Ruling**: D-000.6

**Observation.** A new italic provenance paragraph after the worked-example fence (describing the serializer, the Python cross-check and the golden fixture) is not covered by any of AC51's three amendment rows. AC51 declares *"**Three** amendments now ship together"*, and AC16's third obligation asks for a **test**, not a note in the document — the test does exist and is correct.

**Impact.** Benign in effect; D-000.6 obligation (1) is *"change only the clause proven wrong"*, and this changes no clause but adds one.

**Suggested Resolution.** Add it as amendment D in AC51's table with its ruling, or remove it.

---

### Finding 18: D-000.6 obligation 3 is not yet fulfilled

- **Severity**: Minor
- **Category**: Convention
- **Location**: `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`
- **Related AC**: AC51 · **Ruling**: D-000.6

**Observation.** The working-tree decision log carries `D-000.6 amendment (Story 1.2)` `:1151`, `D-1.2.7` `:1303` and `D-000.6 amendment (Story 1.3)` `:1753` — and **no Story 1.4 amendment record**. All `**Before:** / **After:**` blocks in the log belong to Stories 1.2 and 1.3. The before/after exists only in the story's Delivery Log, and as prose summary rather than the verbatim quotation D-000.6 demands.

**Impact.** The story itself states *"Obligation (3) is the finisher's"*, so this is correctly sequenced rather than missed — recorded here so the finisher has it enumerated, and so the prose-vs-verbatim gap is visible.

**Suggested Resolution.** Finisher adds the Story 1.4 amendment record with verbatim before/after for all three (or four, per Finding 17) amendments.

---

### Finding 19: Several ACs' "all polarities retained" is behaviourally true but has no retained test

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/template/ids_test.go:225` (comment), `folio-go/internal/template/errors_test.go:15-60`, `folio-go/template_test.go:85-93`
- **Related AC**: AC37, AC41, AC6

**Observation.** Three related gaps, each verified behaviourally correct by probe:

- **AC37 `nextId` absent.** `TestNextIDValidation`'s comment says it is *"covered … exercised elsewhere"*; no test removes `nextId`. I removed it and confirmed the error is correct and well-formed (`field nextId: missing required field (never repaired, never inferred, AD-10/AC37)`), but AC37 says *"All polarities retained"* and this polarity is not.
- **AC41.** `TestLoadErrorsCarryFieldAndValue` enumerates six *cases*, describing itself as *"a representative set"*, where AC41 asks for an enumeration of *"the story's load-error constructors"*. It asserts `Field != ""` and conditionally `ElementID != ""` but **never asserts `Value` is populated**, though AC41 names value explicitly. Version errors bypass the `*LoadError` shape check entirely while still counting toward the witness.
- **AC6 (module root).** `TestHigherMajorVersionIsRejected` asserts only `err != nil`, though its own comment claims the error names the declared and supported version. (The stronger assertion **does** exist in `internal/template/version_test.go:16-25`, so the requirement is met — the module-root test is just weaker than its comment.)

**Suggested Resolution.** Add a retained `nextId`-absent case to `TestNextIDValidation`; assert `le.Value != ""` in the error enumeration; either strengthen or re-comment the module-root version test.

---

### Finding 20: Six quotes presented as verbatim are reworded

- **Severity**: Minor
- **Category**: Convention (D-1.1.b)
- **Location**: story AC10, AC11, AC16, AC22, AC33, AC50, AC1
- **Ruling**: D-1.1.b

**Observation.** A full audit of all 98 quoted spans in the story against the working-tree decision log found **every substantive verdict quoted verbatim** — D-1.4.13's rule blockquote and reframe, D-1.4.12's three clauses, D-1.4.4's surviving clauses, D-1.4.3's P1/P2/P3 and traps, D-1.4.7's fallback, D-000.9 obligation 2, D-000.6's four obligations. Six are reworded, and **all six narrow or are cosmetic; none widens**: AC11 drops D-1.4.9's superlative *"the one failure mode"*; AC50 rewords D-1.4.11's *"precisely"*→*"exactly"* and *"the tripwires would be"*→*"they'd be"*; AC16 turns *"would hide"*→*"hides"*; AC22 drops *"must be"*; AC33 swaps a pronoun with an identical referent; AC10 changes *"will read"*→*"reads"*; AC1 truncates at a comma. Two truncations elide log text without an ellipsis (AC14, AC22) without changing meaning.

One attribution looseness: **AC11 attributes the full three-clause sentence to "D-1.4.9 **and** D-1.4.12"**, but D-1.4.9's own guardrail is a **single** clause (*"MINOR additions are strictly additive and never change the meaning of an existing key"*). Clause 3 comes only from D-1.4.12. The outcome is correct because D-1.4.12 supersedes, but reading AC11 alone over-attributes.

**Every ruling ID the story cites exists in the log; no dangling citations.** Bookkeeping: the "Rulings that govern" header omits **D-1.1.b** although the AC preamble quotes it, and cites **D-1.2.5** where the header list says D-1.2.5 but the Delivery Log's D-1.2.5 usage is correct.

**Suggested Resolution.** Restore the seven quotes to verbatim; split AC11's citation so clause 3 is attributed to D-1.4.12 alone; add D-1.1.b to the header list.

---

### Finding 21: The story does not reference four rulings in its own uncommitted decision log that govern it directly

- **Severity**: Minor
- **Category**: Convention
- **Location**: `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md:2404` (D-1.4.14), `:2446` (D-1.4.15), `:2490` (D-000.10), `:2530` (D-1.4.16)

**Observation.** Four rulings landed in the same uncommitted decision-log diff and are not in the story's "Rulings that govern" list or anywhere in its text. Three of them rule directly on this story's code and are the basis of Findings 1, 6, 7, 8 and 13. D-000.10 is scoped *"from Story 1.5"* so it does not bind 1.4's developer, but its reviewer clause does bind this review (dispositions supplied below).

**Suggested Resolution.** Add D-1.4.14, D-1.4.15 and D-1.4.16 to the story's governing-rulings list and fold their requirements into ACs, so the finisher works from the story rather than from the log.

---

### Finding 22: `TestNoOmitemptyStructTags` reflects over a hardcoded type list

- **Severity**: Nit
- **Category**: Tests
- **Location**: `folio-go/internal/template/omitempty_test.go:22-28`

**Observation.** The list names 14 types and `model.go` declares exactly 14 — so it is complete **today**. A new model struct, or `RawValue`/`Presence` (declared in other files and both serialized), would not be inspected. AC22 says *"reflect over **every** serialized struct tag"*.

**Suggested Resolution.** Enumerate types from a registry the model itself owns, or add a comment binding the list to `model.go`'s declaration count.

---

### Finding 23: `absent[T]()` is unused

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/internal/template/presence.go:34`

**Observation.** Confirmed unused (only its definition matches). Per D-1.4.16 this is **not** a defect: *"absence is the **zero value**, so it is never constructed."* Its sibling `presentNull` is genuinely wired at four call sites (`parse_bands.go:199, 212, 230, 493`) and `present` at 27. The real issue in this area is Finding 6, not this function. Recorded only so it is not re-raised.

**Suggested Resolution.** Keep it (its doc comment already explains it is a readability alias), or delete it — either is fine; do not treat it as evidence of an incomplete API.

---

### Finding 24: `ids_test.go` builds JSON by concatenating into `WriteString`

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/internal/template/ids_test.go:69-77`

**Observation.** `sb.WriteString(`...` + id + `...`)` allocates a joined string per element rather than using successive `WriteString` calls. Cosmetic; test-only; correct.

---

### Finding 25: Story opener sits 2 words under the ceiling

- **Severity**: Nit
- **Category**: Convention
- **Location**: story lines 21-51

**Observation.** Opener present at **348 words** (band is 150-350), scope fence present and explicit (*"Deliberately not in this story…"*), and past-tense-ready — the prose is timeless throughout; the single `will` refers to a genuine future product event, not to this story. Compliant, with 2 words of headroom.

---

## Finding Resolutions

Triaged by the story-finisher. Count note first, since it is itself a finding-shaped fact: **`##
Review Summary` states 2 blockers / 8 majors / 13 minors / 4 nits = 27; the findings actually written
below it are 2 blockers / 8 majors / 11 minors / 4 nits = 25** (`### Finding` headings 1–25, severities
counted directly). The minors/total in the summary header are wrong by two; every finding that exists
in the body is resolved below, nothing was invented to fill the gap, and nothing was dropped.

**Both blockers were reproduced red before fixing and confirmed green after**, live, on this machine,
each reverted cleanly afterward (`diff` against a pre-mutation backup reported zero differing lines in
both cases):

- **Finding 1.** Added an unconditional keyed `kv{key: "...", write: ...}` literal to
  `writeDocument` (`serialize.go`) naming an undocumented key. `TestDriftGoToDoc`/`TestDriftDocToGo`
  stayed **green** (reproducing the reviewer's exact claim); the new `TestDriftASTMatchesRuntimeEmission`
  went **red**, reporting the injected key as "actually emitted but invisible to AST extraction."
  Reverted; full drift suite green again.
- **Finding 2.** Reverted `decodePage`'s fix to its pre-finisher shape (refusing any unrecognised
  key). `TestPassthroughAtEveryObjectLevel/page` went **red**: *"expected level \"page\" to accept an
  unknown key opaquely … got: template: field page: unrecognised field on page (value: zzLevelProbe)."*
  Reverted; full suite green again, all 11 subtests passing.

| # | Sev | Decision | Rationale (1–3 sentences) |
|---|---|---|---|
| 1 | Blocker | **FIX** | D-1.4.15 mandates exactly this: a second, behavioural extraction (serialize `maximalFixture`, parse the output, collect actually-emitted keys) asserted equal to the AST set, both directions. Implemented in `drift_test.go` (`TestDriftASTMatchesRuntimeEmission`); the narrow AST shape list is deliberately left as-is per D-1.4.15's "do not fix this by widening the AST shape list" and pinned red by `TestExtractGoKeysMissesAKeyedLiteral`. |
| 2 | Blocker | **FIX** (10 of 11 levels) / **DISMISS** (1 of 11: `bands`) | D-1.4.9 (OWNER) is unambiguous that unknown keys are never refused. Added a passthrough `Extra` store to `Page`, `Margin`, `Padding`, `Border` and `Asset` (`model.go`, `parse.go`, `parse_bands.go`, `serialize.go`), mirroring the five levels that already worked. `bands` is the deliberate exception: AC5 and `folio-format.md` (:101, "Exactly these three keys (FR6)") state the band-name set itself as one of this story's own closed sets — a normative rule from the spec, not an unbacked implementation choice, and the finisher's brief explicitly protected "every structural rule enforced … (bands three-key shape …)" from re-litigation. `TestPassthroughAtEveryObjectLevel` proves both halves at once, per-level. |
| 3 | Major | **FIX** | `TestNumericFieldClassificationInventory`'s `documented` list is now derived from `folio-format.md` at test time (`extractNumericDocKeys`, scanning fenced `"key": NUMBER` pairs) instead of hand-copied. Red-proved live: injecting `"lineSpacing": 1.5` into the document's fenced example reddened the test with the exact message the reviewer predicted; reverted, green again. |
| 4 | Major | **FIX** | D-000.11 (found by this story's reviewer, ruled program-wide by the orchestrator) is applied here: `.github/workflows/ci.yml`'s two `go test ./...` steps (`folio-go`, `lint`) now run `-count=1`, with a comment explaining why. All gates in this Delivery Log's own validation were re-run with `-count=1`. |
| 5 | Major | **FIX** | Added `TestAbsencesChecksIncludeAllFourEntries` (`lint/internal/rules/absences_test.go`), asserting the four specific rule ids (not just a non-zero count). Red-proved live: deleting the `absence-diag-package` entry reddened it with "absenceChecks has 3 entries, want 4"; reverted, green again. |
| 6 | Major | **FIX** | Added a `serializes-differently` subtest to `TestPresenceThreePolarity` (`presence_test.go`) asserting absent/null/present-with-value produce three distinct serialized byte strings, and a `nullFieldFixture` canonical fixture (`fixtures_test.go`, wired into `canonicalFixtures`) so P1/P2/P3 exercise an explicit `null` as a fixed point — D-1.4.16's exact requirement. |
| 7 | Major | **FIX** | Added `TestRenderIgnoresItsArgumentToday` (`template_test.go`) asserting `Render(tpl) == Render(nil)` today, with a comment stating it is designed to fail and be deleted once Story 1.5/1.6 makes `Render` consume its argument — D-1.4.16's exact requirement. |
| 8 | Major | **FIX** | Built `maximalFixture` (`fixtures_test.go`), a canonical document exercising all 51 AST-emittable keys (verified: `assets`' `data`/`mediaType`, `footerOf`/`footerFormat`, `altRowBackground`, `visibleIf`, `italic`, `valign`, etc.), added to `canonicalFixtures` so P1/P2/P3 now cover it, and reused as the runtime side of Finding 1's fix — D-1.4.15's own "one artifact closes both." |
| 9 | Major | **FIX** | Extended the `columns[].footer` row in `folio-format.md` with D-1.4.1's own verdict-table sentence ("unchanged — names the operation only … numeric source is `footerOf`"), appended rather than replacing (D-000.6 obligation 4). The Delivery Log's Amendment A description is corrected in place with a finisher's note rather than silently rewritten. |
| 10 | Major | **FIX** (documentation only, no code) | Corrected the Delivery Log's divergence-class-3 description: the amended worked example carries zero non-ASCII bytes, so `ensure_ascii=False` was defensive, not load-bearing, for this file — the escaping dimension is genuinely covered by `utf8TrapFixture`/`htmlEscapeTrapFixture` instead. No fixture content was altered (DN-1/D-000.6: the worked example's whitespace was the only clause this story was authorised to change, not its content). |
| 11 | Minor | **FIX** | `unknownKeysFixture` now carries a top-level non-ASCII key (`äNonAsciiOrderMarker`) alongside an ASCII one (`zzAsciiOrderMarker`); regenerated canonically confirms byte-order sorts `zz…` before `ä…` (0x7A < 0xC3), discriminating byte-order from locale collation as AC18 requires. |
| 12 | Minor | **FIX** | Added the AC10/AC47-required sentence to `drift_test.go`'s header comment stating unknown passthrough keys are explicitly out of scope for both extraction directions. |
| 13 | Minor | **FIX** | Moved `length-spelling-cases.txt` from `folio-go/testdata/template/` to `folio-go/testdata/shared/` (D-1.4.14) and updated both readers (`internal/template/length_spelling_test.go`, `internal/pdf/length_spelling_shared_test.go`). |
| 14 | Minor | **FIX** | Added `MaxInt64`/`MinInt64` millipoint rows (and a `-999` symmetry row) to the shared spelling table; both new extremes passed `appendPoints`/`appendLength` on first measurement (`9223372036854775.807` / `-9223372036854775.808`). |
| 15 | Minor | **FIX** | `folio.go`'s `PROVISIONAL` marker now reads "Story 1.1 through Story 1.7," matching AC2 and D-1.1.c. |
| 16 | Minor | **FIX** | Dropped the trailing sentence from `folio-format.md`'s shipped MINOR-additive blockquote — D-1.4.12's verdict is exactly three sentences ending "… load errors," verified against the decision log; the fourth sentence was accurate but not part of the owner-ruled verbatim text. |
| 17 | Minor | **FIX** | Added the provenance paragraph as Amendment D in AC51's table (this story), with its ruling (D-1.4.10, AC16 obligation 3). The paragraph itself was already correct and stays unchanged in `folio-format.md`. |
| 18 | Minor | **FIX** | This is the obligation the story itself flagged as the finisher's — closed by [§ D-000.6 amendment (Story 1.4)](#d-0006-amendment-story-14) below, added to the decision log with all four amendments' verbatim before/after. |
| 19 | Minor | **FIX** (all three sub-items) | (a) Added `TestNextIDAbsentIsLoadError` (`ids_test.go`) via a new `minimalDocMissingNextID` helper — the "all polarities retained" claim for AC37 is now true. (b) `TestLoadErrorsCarryFieldAndValue` (`errors_test.go`) now asserts `le.Value != ""` alongside `le.Field`. (c) `TestHigherMajorVersionIsRejected` (`template_test.go`) now asserts the error names both the declared and supported version at the module-root API, not only in `internal/template`'s mirror test. |
| 20 | Minor | **FIX** | Restored seven quotes to verbatim against the decision log (AC1, AC10, AC11, AC14, AC16, AC22, AC33), split AC11's citation so clause 3 is attributed to D-1.4.12 alone, and added `D-1.1.b` to the header's "Rulings that govern" list. |
| 21 | Minor | **FIX** | Added `D-1.4.14`, `D-1.4.15`, `D-1.4.16` (and `D-000.10`, `D-000.11` for completeness) to the header's "Rulings that govern" list, with a finisher's note explaining why they were absent and what they govern. |
| 22 | Nit | **FIX** | `omitempty_test.go`'s reflected type list is now bound to an AST-derived count of struct declarations across `model.go`/`presence.go`/`rawvalue.go` (`TestOmitemptyTypeListIsComplete`), and the list itself now includes `RawValue` and one `Presence[T]` instantiation, which the original 14-type list silently excluded. |
| 23 | Nit | **DISMISS** | D-1.4.16 rules this explicitly: absence is `Presence[T]`'s zero value, so `absent[T]()` is never constructed — that is correct, not an incompleteness. No change. |
| 24 | Nit | **DISMISS** | The reviewer's own text: "Cosmetic; test-only; correct," with no suggested resolution offered. A style-only rewrite of a passing test with zero behavioural stakes is exactly the kind of change this role's minimal-diff discipline argues against. No change. |
| 25 | Nit | **DISMISS (superseded)** | The finding concerns the *pre-finisher* opener. Per this pass's own protocol, [§ In plain terms](#in-plain-terms-read-this-first-if-you-just-want-the-gist) is rewritten below to describe what actually shipped (including this pass's fixes), which supersedes the word-count finding rather than needing it separately re-verified. |

**Files changed by this pass, by finding:** `folio-go/internal/template/{model.go,parse.go,parse_bands.go,serialize.go,drift_test.go,fixtures_test.go,roundtrip_test.go,numeric_classification_test.go,presence_test.go,omitempty_test.go,ids_test.go,errors_test.go,length_spelling_test.go,passthrough_test.go}`, `folio-go/{folio.go,template_test.go}`, `folio-go/internal/pdf/length_spelling_shared_test.go`, `folio-go/testdata/shared/length-spelling-cases.txt` (moved), `lint/internal/rules/absences_test.go`, `.github/workflows/ci.yml`, `_bmad-output/specs/spec-folio/folio-format.md`, this story file, `folio-mvp-decision-log.md`.

---

## Ruling dispositions (reviewer side, per D-000.10)

| Ruling | Reviewer disposition |
|---|---|
| D-1.4.1 | `applied-as-stated` — footer schema fields and key order shipped; `footerFormat` sorts before `footerOf` (verified in bytes) |
| D-1.4.2 | `applied-as-stated` — exactly the three decidable checks, both "No" rows unimplemented, no codes minted, AC44 fixture present with the naming comment |
| D-1.4.3 / (extended) | `applied-as-stated` — exact ×1000 `big.Int` arithmetic, `UseNumber`, no `float64`, no `omitempty`, nil containers emit `{}`/`[]` |
| D-1.4.4 (reversed in part) | `applied-as-stated` — base-36 ids, decimal `nextId`, spelling enforced not normalised, duplicates rejected, delete-guardrail **red-proved against a renumbering serializer** |
| D-1.4.5 | `applied-as-stated` — `72.5001` rejected, `36.0000` normalised, one rounding site |
| D-1.4.6 | `applied-as-stated` — `os` confined to package `folio`, `internal/template` takes `[]byte` |
| D-1.4.7 | `applied-with-a-different-mechanism` — AST extraction substituted for struct-tag reflection (forced: no struct tags exist). **Not exact** — see Finding 1 |
| D-1.4.8 (corrected) | `applied-with-a-different-mechanism` — presence flags correct on the parse side; **serialization side unpinned**, Finding 6 |
| **D-1.4.9 (OWNER)** | **`applied-partially` — Finding 2.** Passthrough works at 5 of 11 object levels; 6 refuse |
| D-1.4.10 | `applied-as-stated` — worked example expanded, cross-check reproduced (0 differing lines); one named class inert, Finding 10 |
| D-1.4.11 | `applied-with-a-gap` — witness correct and red-proved; shrunk-list case open, Finding 5 |
| **D-1.4.12 (OWNER)** | `applied-as-stated` — all eight closed sets named and **all eight enforced at load** (verified by probe); one extra trailing sentence, Finding 16 |
| D-1.4.13 | `applied-as-stated` — reframe verbatim in `versionForSave`'s doc comment, both pinning tests present and genuine |
| D-1.4.14 | `not-reached-in-this-story` — table not moved, Finding 13. No route to an output byte bypasses `appendJSONString` (passthrough strings decode to Go strings and re-encode through it; only `RawNumber` re-emits a literal, documented and P3-safe) |
| D-1.4.15 | **`not-reached-in-this-story`** — behavioural extraction and `astKeys == runtimeKeys` not built, Findings 1 and 8 |
| D-1.4.16 | **`not-reached-in-this-story`** — neither behavioural guard built, Findings 6 and 7 |
| D-000.4 | `applied-as-stated` — matrix correctly unrun and named |
| D-000.6 | `applied-partially` — Findings 9, 17, 18 |
| D-000.9 | `applied-partially` — Findings 3, 4, 5 |
| D-1.1.b | `applied-partially` — Finding 20 (all deviations narrowing) |
| D-1.1.c | `applied-as-stated` — provisional marker updated not deleted; range wrong, Finding 15 |
| D-1.2.5 | `applied-as-stated` — every cited measurement reproduced on this machine |
| D-1.2.6 | `applied-as-stated` — no ruling conflict was parked |
| D-1.3.4 | `applied-as-stated` — tripwires directory-keyed and **verified to fire** |

**Finisher's post-fix status (this pass, not a rewrite of the table above — that table is the
reviewer's own record at review time and stands as written).** Every ruling the reviewer marked
`applied-partially`, `applied-with-a-gap` or `not-reached-in-this-story` above is now, after this
pass's fixes, `applied-as-stated`: **D-1.4.7** (Finding 1 fixed — behavioural extraction added),
**D-1.4.8 (corrected)** (Finding 6 fixed), **D-1.4.9 (OWNER)** (Finding 2 fixed at the 10 levels with
no closed-set backing; `bands` stays closed, correctly, per AC5/FR6), **D-1.4.10** (Finding 10's
Delivery Log text corrected), **D-1.4.11** (Finding 5 fixed), **D-1.4.12 (OWNER)** (Finding 16 fixed),
**D-1.4.14** (Finding 13 fixed — table moved), **D-1.4.15** (Findings 1 and 8 fixed), **D-1.4.16**
(Findings 6 and 7 fixed), **D-000.6** (Findings 9, 17, 18 fixed), **D-000.9** (Findings 3, 4, 5 fixed),
**D-1.1.b** (Finding 20 fixed). See [§ Finding Resolutions](#finding-resolutions) above for the
per-finding detail and file list.

