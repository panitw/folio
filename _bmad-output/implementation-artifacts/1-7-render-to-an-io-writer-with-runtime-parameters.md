---
baseline_commit: 34631191cf2ee1761a4e08d00a08362c2376ebc5
---

# Story 1.7: Render to an `io.Writer` with runtime parameters

Status: done

**Story key:** `1-7-render-to-an-io-writer-with-runtime-parameters`
**Epic:** 1 · **Baseline:** `3463119`
**Covers:** FR39, FR40 · AD-7, AD-8, AD-14, AD-22, AD-23
**Adjacent:** AD-1, AD-4, AD-5, §Source tree, §Deferred
**Rulings that govern:** **D-1.7.1**, **D-1.7.2**, **D-1.7.3**, **D-1.7.4**, **D-1.7.5**,
**D-1.7.6**, **D-1.7.7**, D-000.4, D-000.9 *(and its extension)*, D-000.10, D-000.11, D-000.12,
**D-000.13**, **D-000.14**, D-1.1.a *(addendum)*, D-1.1.c *(and its addendum)*, D-1.2.6, D-1.3.1,
D-1.3.4, D-1.4.11, D-1.4.14, D-1.5.5, D-1.5.9, D-1.6.1, D-1.6.3, D-1.6.5, D-1.6.6
**Deferred work this story's rulings register:** **DW-9** (owner: Story 2.2), **DW-10** (owner:
Story 3.7)
**Counter-metric under test:** **C5** — *"Time-to-first-PDF for a new integrator exceeding a few
minutes → the 'extremely simple API' goal has failed."*

---

> ### Where these rulings live — read before diffing
>
> **D-1.7.1 – D-1.7.7 and D-000.14 are UNCOMMITTED additions in the working-tree copy of
> `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`. They are NOT in `HEAD`, and
> they are NOT in baseline `3463119`.** The same is true of **DW-9** and **DW-10** in
> `_bmad-output/implementation-artifacts/deferred-work.md`. **A reviewer who diffs against
> `3463119` will not find any of them** and must read the working-tree files instead.
>
> **Verified at the time of writing, not inherited from the brief.** Measured:
>
> | Check | Command | Result |
> |---|---|---|
> | `HEAD` is the stated baseline | `git rev-parse HEAD` | `34631191cf2ee1761a4e08d00a08362c2376ebc5` |
> | The rulings are absent from `HEAD` | `git show 3463119:…/folio-mvp-decision-log.md \| grep -c "D-1.7"` | **0** |
> | The rulings are present in the working tree | `grep -c "D-1.7" …/folio-mvp-decision-log.md` | **7 matches** (D-1.7.1 … D-1.7.6 + one cross-reference) |
> | DW-9 absent from `HEAD` | `git show 3463119:…/deferred-work.md \| grep -c "DW-9"` | **0** |
> | DW-9 present in the working tree | `grep -c "DW-9" …/deferred-work.md` | **1** |
> | Exactly which files differ | `git diff --stat 3463119` | `deferred-work.md` **+22**, `folio-mvp-decision-log.md` **+145**. **No Go file differs from `HEAD`.** |
>
> **The counts above were measured BEFORE D-1.7.7, D-000.14 and DW-10 were ruled** — those three
> were logged during this story's creation, in response to the `DECISION NEEDED` this file raised,
> and are **also** working-tree-only. So the true D-1.7.x count is now higher than the `7` recorded
> above; the row is left as measured rather than silently updated, and this note carries the delta.
> **Re-measure rather than trusting either number.**
>
> These rulings will land in a commit that is **not this story's**. Story 1.6's file carried the
> same banner and its first line counts were wrong; the counts above were measured, once, from
> `git diff --stat` output, not copied.

> ### Mechanism tagging (D-1.4.15's convention)
>
> Every named mechanism below carries `(binding)` or `(illustrative)`. **An illustrative mechanism
> may be improved on; a binding one may not be departed from without surfacing it as
> `DECISION NEEDED`.** Where this story quotes a ruling, the tag travels with the quote.
>
> **Nothing is parked.** This story's creator surfaced one `DECISION NEEDED` — AD-7's
> `/CreationDate` clause, which becomes satisfiable for the first time here — and it was **ruled
> while this file was being written, as D-1.7.7: it dissolves, because Story 3.7 already owns it
> with acceptance criteria written.** Per D-1.2.6 the creator surfaced it and did not arbitrate it;
> the ruling is carried at [§ Decisions resolved](#decisions-resolved), and 1.7's two mechanical
> obligations are at [§ Section F](#f--handing-creationdate-to-story-37-mechanically-d-177--ad-7-nfr1f--dw-10).

---

## In plain terms (read this first if you just want the gist)

Today the library turns a document plus report data into a finished PDF and hands back a block of
bytes. This story adds the two things an integrator needs before the library is usable from a web
server, and both shipped.

The first is a way to write the finished document straight into a destination the caller supplies —
a web response, a file already opened — instead of always receiving bytes to copy. How the document
is produced did not change: it still has to be complete before a single byte can leave, since its
fingerprint is computed over the whole thing. A destination that fails partway, accepts less than it
was given, or is missing entirely is now reported as a located error rather than a silent partial
write or a crash.

The second is a separate channel for values supplied at render time — a statement date, say — kept
apart from the report data, with a missing value failing loudly and by name. Review found this
separation had partly been built as a second, parallel copy of the report-data checks rather than
genuinely sharing them, so a couple of checks looked identical but were unproven for the new
channel; that was consolidated into one shared implementation, and a follow-up check confirmed the
previously-unproven cases now fail when they should. Two structural guards meant to prove "only one
place produces documents" and "only one place decodes JSON" were also narrower than claimed, and
were broadened to match.

The story also wrote the project's first integrator-facing guide. Review found small things wrong
with it — a placement promise pointing at content that did not exist yet, an example comment that no
longer matched the code beside it, and a missing caveat about the library's own "no temporary file"
claim — all now corrected.

Deliberately not here: the ready-made typeface bundle, page numbering, and hidden-property dates —
the guide's example still supplies its own font bytes, and a tripwire keeps the date question from
being skipped later unnoticed. The cross-machine byte-matching checks were not run this story, as
planned.

---

## Story

As an integrating Go developer,
I want to write a PDF straight to my HTTP response and pass runtime values separately from report
data,
So that I can serve a generated document without a temporary file and without smuggling dates into
the data.

---

## Acceptance Criteria

Source ACs are `_bmad-output/planning-artifacts/epics.md` §Story 1.7 (lines 609–636). They are
carried below unchanged in substance and split so each is separately red-proofable, with the
ruling-mandated additions folded in. Every AC below is stated so that the three diagnostic
questions (AC26) have an answer other than *"the same thing"*.

**The source AC set is four items. It expands to 31 below** because AC2 as written is *vacuous*
(D-1.7.2), AC1's "no temporary file" is a capability claim that needs a guard *and* a stated
residual gap (D-1.7.3), AC3 splits into precedence, absence, and the bare-namespace case (D-1.7.4),
and AC4 pulls in the type guardrail (D-1.7.5) and the toolchain caveat (D-1.7.6). **Section F's two
ACs are not derived from a source AC at all** — they are D-1.7.7's mechanical handoff to Story 3.7.

### A — The writer entry point (D-1.7.1, D-1.7.2 · AD-7, AD-8 · FR39 · source AC1, AC2)

**AC1 — the two entry points, five and four positional arguments** *(binding, D-1.7.1)*. Package
`folio` declares, verbatim:

```go
func Render(t *Template, d Data, p Params, f FontSet) ([]byte, error)
func RenderTo(w io.Writer, t *Template, d Data, p Params, f FontSet) error
```

**No options struct** *(binding)*. D-1.7.1's decisive argument, quoted:

> *"An options struct makes `FontSet` **omittable at compile time**: `folio.Request{Template: t,
> Data: d}` would compile and carry a zero `FontSet`, turning an **AD-8 violation from a compile
> error into a runtime one**."*

AD-8's Rule makes this load-bearing (verified, `ARCHITECTURE-SPINE.md:220`): *"the same template
with a different chain is a different render, not a silent substitution."* **D-1.1.c's
medium-confidence flag on argument packaging is resolved to high and marked closed** — nobody
reopens it at the `folio-go/v0.1.0` tag.

**AC1a — the source AC's shorthand, recorded not treated as a contradiction.** The source AC writes
`folio.RenderTo(w, template, data, params)`, omitting the `FontSet` AD-8 requires. **The spine
governs; the AC is shorthand** (D-1.1.c addendum, verbatim: *"The 1.7 creator carries the
five-argument form and notes the AC's shorthand rather than treating it as a contradiction."*)

**Finisher's note (Finding 9, this story's review, Minor, D-1.2.6):** D-1.7.1's own Verdict states
`RenderTo`'s signature **verbatim** with five parameters, then four paragraphs later says
*"Carry the **six-argument** form"* when explaining the `FontSet` shorthand. This is an arithmetic
slip inside D-1.7.1 against its own verbatim verdict, not a second, different instruction — the
shipped code (five/four positional arguments, matching the verbatim Verdict) is correct, and the
verdict governs over the later miscount. Surfaced per D-1.2.6 rather than arbitrated; no code
change. The decision log itself is append-only and is not edited.

**AC2 — `p Params` is INSERTED, never appended** *(binding, D-1.5.5, verbatim: "No call site is
ever reordered, only extended")*. Target order is `Render(t, d, p, f)`; `p` goes **between** `d` and
`f`. **Measured blast radius (M-1): 19 real `Render(` call expressions, all in test files, zero
production callers.** See [§ Dev Notes → M-1](#m-1--the-signature-blast-radius-is-19-call-sites-all-in-tests).

**AC3 — `RenderTo` builds the whole document, then writes it once, and says why in a code comment**
*(binding, D-1.7.2)*. **Streaming is not available and must not be attempted.** AD-7's Rule
(verified, `ARCHITECTURE-SPINE.md:199–201`) makes `/ID` *"the first 16 bytes of a SHA-256 over the
serialized body up to the point `/ID` is written"*, and the classic xref table carries **byte
offsets** — both need the complete document before the trailer can be written. D-1.7.2, verbatim:

> *"'No buffering of the whole document' is a property we **cannot have**, and asserting it would
> assert something false. **Put that reasoning in a code comment** — someone will eventually try to
> 'optimise' `RenderTo` into a stream, and it would break AD-7 silently."*

The comment lives on `RenderTo` itself, names AD-7 and D-1.7.2, and states the xref-offset half as
well as the `/ID` half.

**AC4 — the delegation is asserted STRUCTURALLY, and the vacuous behavioural test is not written**
*(binding for the property; illustrative for the AST form, D-1.7.2)*. A test parses package
`folio`'s non-test files with `go/ast` and asserts: **exactly one function produces the document
bytes**, and **both `Render` and `RenderTo` route through it**. Byte-identity then holds as a
**theorem**, not as a test that cannot fail.

> **D-1.7.2, verbatim** *(binding)*: *"**Delete the vacuous behavioural comparison rather than
> keeping it for comfort — a green test nobody can fail reads as coverage.**"*

**Source AC2 ("rendered through `Render` and through `RenderTo`, the bytes are identical") is
therefore NOT implemented as a behavioural byte comparison.** A shared core is the *correct* design
(two emitters would be D-1.4.14's drift hazard), which is exactly why such a comparison cannot fail.
**If a developer writes one anyway, the reviewer removes it.** The one core today is
`renderDocument` in `folio-go/render.go` (measured, M-2).

**AC4a — the AST assertion's own vacuity guard** *(binding, D-000.9)*. Zero producing functions
found, or zero routing entry points found, is a **hard failure**, never "nothing to check, pass".
Its red-proof (RP-2) adds a second, independent byte-producing path inside `RenderTo` and asserts
the finding **names both functions**, not merely that the test failed.

**AC5 — a writer that errors mid-write surfaces the error** *(binding, D-1.7.2)*. `RenderTo`
returns that error, **never `nil`**, and the returned error identifies the write failure rather
than being swallowed or replaced by a render error.

**AC6 — a short write is detected and reported** *(binding, D-1.7.2)*. A writer that returns
`n < len(p)` with `err == nil` must produce a non-nil error from `RenderTo`. **Story 1.1's `TestMain`
learned this the hard way (Nit 25)** — the reference is deliberate, not decorative.

**AC7 — the exact byte count written equals the document length** *(binding, D-1.7.2)*. Asserted
with a **counting writer**: total bytes observed by the writer `== len(b)` where `b` is what
`Render` returns for the same inputs. **This is the honest half of source AC2** — it compares a
number derived from the writer path against a number derived from the byte path, and it *can* fail.

**AC8 — nothing extra is written** *(binding, D-1.7.2)*. No trailing newline, no second `Write`
call carrying a tail. Asserted by recording every `Write` call's payload and comparing the
concatenation to `Render`'s output — and by asserting the **call count** the implementation
actually makes, so an added flush/tail write reddens.

### B — The world-reading fence survived the move (D-1.7.3 · AD-1 · FR38 · source AC1)

**AC9 — the guard locates the render entry file BY AST, never by filename** *(binding, D-1.7.3)*.
D-1.6.3's rule must locate **by AST the file(s) declaring `Render` AND `RenderTo`** — by
declaration, never by path. **Measured (M-3): Story 1.6's finisher already rewrote the guard this
way** — `findRenderDeclaringFiles` in `lint/internal/rules/forbiddenimports_test.go` parses every
non-test `.go` file directly under `folio-go/` and collects those declaring a top-level `Render` or
`RenderTo`. **D-1.7.3 states this as a requirement rather than assuming it carried, and 1.7's job
is to PROVE it survived**, because:

> **D-1.7.3, verbatim** *(binding)*: *"Story 1.6's blocker was the same family one level down: the
> lead ruled the *property* ('the file declaring `Render`') and the implementation asserted a
> *name*. The reviewer moved `Render` into the `os`-importing file, left the empty file in place,
> and the guard **passed**. **1.7 is the very story that moves the entry point.**"*

**AC10 — the red-proof is exactly the Story 1.6 reviewer's mutation, and it asserts on the FINDING,
not on exit status** *(binding, D-1.7.3 + D-000.13)*. Move `RenderTo` into `folio.go` (which imports
`os`), rebuild — the mutation **must compile** (valid syntax, forbidden semantics) — and the guard
must go red **with a named finding** identifying the file and the offending import.

**Pre-measured at baseline `3463119` (M-4). This red-proof has already been run and it fires.**
Both the control (build succeeds) and the finding text are recorded in
[§ Dev Notes → M-4](#m-4--the-d-173-red-proof-fires-at-baseline-measured-not-predicted). The
developer re-runs it in-story; it is recorded here so a green run cannot be mistaken for "nothing
to prove".

**AC11 — the residual gap is RECORDED, not papered over** *(illustrative, D-1.7.3)*. A file-scoped
import rule does not stop a **deliberate cross-file route**: `RenderTo` in the clean file calling a
helper in `folio.go`, which legitimately imports `os`. **That is accepted** — the AC's intent is a
**capability** claim (*"you can serve a PDF without touching disk"*), not a security boundary, and
AC7/AC8's counting-writer assertions give the behavioural half.

**Measured, not predicted (M-5): the gap is real and green.** A probe in which `RenderTo` stayed in
the clean file and routed through an `os.WriteFile` helper declared in `folio.go` **compiled and
the guard PASSED**. The note in the code and in the story must say the guard is a **capability
fence, not a proof**. **Do not build a filesystem-snapshot check** *(binding)* — disproportionate,
and it would test the OS rather than the library.

### C — `params` is a resolution ROOT, not a reserved token (D-1.7.4 · AD-4, AD-14 · FR40 · source AC3)

**AC12 — `params` is a second resolution root, and it is NOT added to the reservation list**
*(binding, D-1.7.4)*.

> **D-1.7.4, verbatim** *(binding)*: *"`page` / `pages` are **reserved whole tokens** … `params` is
> a **namespace** — a second resolution root resolved at bind time, alongside the data root. **Do
> not implement `params` by extending the `page`/`pages` reservation list.** They are different
> things, and conflating them is how `page` eventually acquires a namespace — which AD-4 forbids
> forever."*

Concretely: `reservedPlaceholders` in `folio-go/internal/bind/text.go` gains **no new entry**. The
binder gains a second root, threaded through the same `BindText` signature family. **A developer who
adds `"params": true` to that map has failed this AC even if every behavioural test passes.**

**AC13 — precedence: the first segment `params` ALWAYS selects the params document** *(binding,
D-1.7.4)*, never report data, **regardless of what data contains**.

**AC14 — the retained fixture, copying D-1.6.5's `page` fixture verbatim in shape** *(binding,
D-1.7.4)*: **data containing a top-level `params` key must not change what `{{params.reportDate}}`
resolves to.** The fixture copies the *shape* of `TestBindTextPageReservationIsRedProofable`
(`folio-go/internal/bind/text_test.go:132`), which QA Nit 15 already sharpened once:

- **Two data documents, both carrying a top-level `params` key, with DIFFERENT values**, plus one
  supplied params document.
- Assert **both** that the two renders are equal **and** that each equals the **params** value.
  Equality alone is the weaker half; the value assertion is what names the hijack.

**Measured pre-fix (M-6). At baseline this AC is genuinely red, and the shadowing is real:** with
data `{"params":{"reportDate":"SHADOWED-FROM-DATA"}}`, `{{params.reportDate}}` today renders
**`"SHADOWED-FROM-DATA"`, with no error** — and the whitespace-tolerant spelling
`{{ params.reportDate }}` shadows identically, so the fix belongs at **path-segment** level, not on
the trimmed literal string.

**AC15 — a top-level `params` key in report data is NOT an error** *(binding, D-1.7.4)*. It becomes
unreachable by any binding, **and that is documented** — in `Render`'s doc comment and in the
README.

> **D-1.7.4, verbatim**: *"rejecting it would refuse legitimate caller data over a name collision
> the caller may not control. **The caller's JSON shape is their business; only the binding
> namespace is ours.**"*

**AC16 — `{{params.x}}` with no params supplied is an Error naming the path AND the element id**
*(binding, D-1.7.4)*. AD-14's absent-path case (verified, `ARCHITECTURE-SPINE.md:298–300`: *"an
**absent** path is an `Error` carrying the path"*) applied to the params root, through the **same
code path** as absent data.

> **D-1.7.4, verbatim**: *"Rendering empty would **silently produce a statement missing its report
> date** — precisely the failure AD-14's first case exists to make loud."*

**The error must name the PARAMS root, not report data.** Measured (M-6): today the message reads
`data path "params.reportDate" is absent from the report data` — which after this story would be
actively misleading, since the value was never sought in report data at all.

**AC17 — `{{params}}` bare, no dot, is a located error** *(illustrative for the wording, D-1.7.4)*:
*"`params` is a namespace, not a value."* Measured (M-6): today it produces
`data path "params" is absent from the report data` — wrong on both counts.

**AC17a — `{{page}}` and `{{pages}}` are unaffected** *(binding, AD-4)*. The existing reservation
tests stay green **and are re-run in the same gate** as the params tests, so a params implementation
that quietly reroutes the reservation branch reddens. Measured (M-6): `{{page}}` still passes
through byte-for-byte at baseline. **AD-4's "no `page` namespace exists, ever" is not weakened by
this story.**

### D — `Params` is a defined type on one decode path (D-1.7.5 · AD-23 · source AC3, AC4)

**AC18 — `type Params []byte` is a DEFINED type, never an alias** *(binding, D-1.7.5)*. **Verified:
`type Data []byte` at `folio-go/render_entry.go:52` is defined, not aliased.** `Params` matches it.

> **D-1.7.5, verbatim** *(binding)*: *"**An alias (`type Params = []byte`) would make the two
> mutually assignable and destroy the entire point**"* — D-1.1.c: *"in a product whose acceptance
> fixture is a bank statement, that swap must be a compile error, not a support ticket."* *"**At 1.7
> they become adjacent same-typed-looking arguments, which is the exact moment that reasoning was
> written for.**"*

**AC19 — a COMPILE-TIME proof that the swap is rejected — an assertion, not a comment**
*(binding, D-1.7.5)*. A `//go:build ignore` fixture compiled by a test, or a `go/types`-based
assertion, proves `Render(t, params, data, f)` **does not type-check**.

**Measured (M-7): the property holds for defined types and is destroyed by an alias**, and the
probe's control compiles, so the assertion is not vacuously red. Exact compiler diagnostic recorded
in [§ Dev Notes → M-7](#m-7--the-swap-proof-and-what-it-does-and-does-not-cover).

**AC19a — the guardrail covers `Data` as well as `Params`, and this is the FIRST such assertion.**
Measured (M-8): **Story 1.6 shipped no assertion at all that `Data` is a defined type** — AC23 of
that story is carried only by a doc comment in `render_entry.go`. A comment claiming a safety
property is what D-1.7.5 explicitly refuses. The new assertion covers both types in one place.

**AC19b — the residual gap is stated** *(illustrative)*. An **explicit conversion**
(`Render(t, Data(p), Params(d), f)`) still compiles with defined types. The guardrail proves the
**accidental** swap is a compile error; it does not stop a deliberate cast. Same shape as AC11 —
record it so a later reader does not mistake the guard for a proof.

**AC20 — ONE decode path, shared with report data** *(binding, D-1.7.5)*. Params are decoded with
the **same `UseNumber` discipline**, the **same single literal splitter (D-1.6.1)**, and the **same
`Decimal`** as report data. **No second decoder.**

> **D-1.7.5, verbatim**: *"A second decoder would be the drift hazard refused three times already —
> and worse here, since **params carry dates and amounts, where a divergence would be least visible
> and most expensive.**"*

Concretely: params go through `internal/bind.DecodeData`, inheriting its trailing-garbage rejection
(QA Finding 4 of Story 1.6) and its map-range escape hatch **for free**. **A guard asserts exactly
one `json.NewDecoder`/`UseNumber` site exists under `internal/bind`**, so a future second decoder
reddens rather than drifting silently. Zero sites found is a hard failure (AC26).

### E — The README (D-1.7.1, D-1.7.6 · AD-22 · C5 · source AC4)

**AC21 — a README exists, and producing a first PDF takes a load call and a render call**
*(binding for the two-call shape; illustrative for prose)*. **Verified: there is no README anywhere
in the repository — this story writes the first** (M-9). This is the artifact counter-metric **C5**
measures: *"A new integrator produces their first PDF in minutes from the README alone."*

**AC22 — the AC4 ceremony judgement is DEFERRED to Story 2.2 (DW-9)** *(binding, D-1.7.1)*.
**Verified: `folio-go/fonts/` does not exist** (M-9) — and AD-8's Rule requires it to
(`ARCHITECTURE-SPINE.md:214`: *"The repository holds font binaries in exactly one place,
`folio-go/fonts/`"*). The shipped faces arrive at Story 2.2. So the first-PDF example **must show a
caller assembling a `FontSet` from their own bytes**, which *is* ceremony —

> **D-1.7.1, verbatim** *(binding)*: *"**ceremony that 2.2 removes, not ceremony the signature
> causes.**"*

**AC22a — the escape hatch, stated exactly** *(binding, D-1.7.1)*: *"If the 1.7 README author finds
the example reads as ceremony **for reasons other than the missing font set**, that is a
`DECISION NEEDED`, not a unilateral change."* **The developer may not swap the signature. DW-9's
owner is Story 2.2, and it must be raised before the `folio-go/v0.1.0` tag.**

**AC23 — the toolchain caveat, near the golden-hash discussion** *(binding, D-1.7.6)*. The README
states that a **`toolchain` directive binds only the main module** — so an integrator compiling
`folio-go` into their own binary builds with **their** toolchain and may get different bytes.

> **D-1.7.6, verbatim** *(binding)*: *"That is AD-22's accepted position, **but it is invisible
> until it bites**, and **counter-metric C5's 'first PDF in minutes' integrator is exactly the
> person who hits it.** It belongs **near the golden-hash discussion, not in a footnote**: someone
> recording Folio's hashes in their own test suite needs to know what pins them."*

**Measured, not asserted from documentation (M-10).** The claim was proved by construction with a
control that fires. Numbers in [§ Dev Notes → M-10](#m-10--the-toolchain-caveat-is-empirically-true-with-a-control-that-fires).
The README's wording must tell the integrator **what actually pins their bytes** (their own
`toolchain` line, or `GOTOOLCHAIN`), not merely that Folio pins one.

**AC23a — D-1.1.a's `go`-below-`toolchain` reason is not restated wrongly.** `folio-go/go.mod`
carries `go 1.25.0` **below** `toolchain go1.26.0` deliberately (D-1.1.a: Go strips a non-greater
`toolchain` line on `go mod tidy`). If the README explains the pin, it must not present the version
gap as an error to fix.

### F — Handing `/CreationDate` to Story 3.7, mechanically (D-1.7.7 · AD-7, NFR1.f · DW-10)

**Context.** This story's creator surfaced AD-7's `/CreationDate` clause as `DECISION NEEDED`. It is
**ruled as D-1.7.7 and it dissolves — Story 3.7 already owns it, with acceptance criteria written.**
See [§ Decisions resolved](#decisions-resolved) for the full ruling. The two ACs below are 1.7's
entire share of it: **1.7 emits no PDF date and builds no date type.** What it must not do is leave
the handoff resting on nobody noticing.

**AC24 — re-scope `TestRenderHasNoCreationOrModDate`; do not leave it accidentally right**
*(binding, D-1.7.7)*. `folio-go/render_test.go:598` asserts `/CreationDate`, `/ModDate`, `/Info`,
`/Producer`, `/Creator`, `/Metadata` and `/Filter` are absent **unconditionally**. That assertion is
true today **solely because nothing can supply a date**. Now that `Params` exists it becomes a
statement about **a render whose params carry no date** — which is precisely Story 3.7's fourth AC,
verbatim from `epics.md:1098–1100`:

> **Given** no date supplied by any route · **When** the PDF is produced · **Then**
> `/CreationDate` and `/ModDate` are omitted

**It is testable today, so assert it deliberately.** Rename and re-scope the test so its name and
its message say *"a render whose params supply no date"*, add at least one case whose params are
**non-empty but carry no date key**, and cite Story 3.7 as the owner of the positive case.

> **Why this is not cosmetic** *(binding rationale, D-1.7.7)*: **left unconditional, 3.7's developer
> meets a red test whose cheapest resolution is to weaken it** — and weakening a metadata-absence
> assertion is exactly the erosion this run has spent six stories preventing. A test that is right
> for a reason it does not state is a test that will be edited for the wrong reason.

**AC25 — add an absence tripwire on `folio-go/cmd/`** *(binding, D-1.7.7 · D-1.3.4's mechanism)*.
One row in `absenceChecks` (`lint/internal/rules/absences.go:43`), alongside the existing
`folio-designer`, `folio-go/fonts`, `folio-go/internal/expr` and `folio-go/internal/diag` rows:

| Field | Value |
|---|---|
| `relPath` | `folio-go/cmd` |
| `rule` | `absence-cmd-dir` |
| `desc` | `folio-go/cmd/ must be absent until Story 3.7 builds the CLI and settles AD-7's params-date wiring (DW-10)` |

**Verified: `folio-go/cmd` does not exist** (M-13). Keyed on the **directory**, per DW-2's recorded
correction — an exact filename guess is the false-pass hazard a directory-level check does not have.
**`cmd/folio` cannot be created without tripping this**, so its creation is the mechanical trigger
that forces the wiring to be settled rather than remembered.

**AC25a — the D-1.4.11 sequencing precondition, verified before the row is added** *(binding)*.

> *"Adding entries to a list whose emptiness passes is exactly how a tripwire becomes decorative."*
> — `absences.go:56-58`, Story 1.4's own note

**Verified live (M-12):** `ScanAbsences` returns `AbsencesStats.ChecksEvaluated`,
`TestAbsencesProductionScan` **fails hard** on `ChecksEvaluated == 0`, and
`TestAbsencesZeroWitnessIsCaught` is its red-proof by construction (a swapped-in empty list, not a
reverted production mutation). **The witness is live, so adding a row is safe now.** The developer
re-confirms this before touching the table and records the check — **do not add the row on this
story's say-so.**

**AC25b — the row's own red-proof.** Create `folio-go/cmd/` (with a file in it — an empty directory
is not tracked by git), run the absence gate `-count=1`, and assert the finding carries **rule id
`absence-cmd-dir`** and names the path (D-000.13: rule id and message, never exit status). Remove
the directory. Control: with the directory absent, the gate is green **and `ChecksEvaluated` equals
the row count**.

### Standing rules

**AC26 — D-000.9, its extension, and D-000.13: three questions of every guard and every red-proof**
*(binding)*:

1. *What would this check have printed if it had been unable to run at all?* (D-000.9)
2. *What would this red-proof have printed if the mutation had never been applied?* (D-000.9 extended)
3. ***Did it fail for the reason it names?*** (D-000.13)

**A red-proof asserts on the rule id and the finding message, never on exit status or mere failure.
A control mutation must be VALID SYNTAX with FORBIDDEN SEMANTICS** *(binding, D-000.13)*. Every
red-proof in this story's table records the **control result** alongside the mutation result.

**AC27 — D-000.11.** Every gate runs `-count=1`. **A cached `ok` is not evidence.**

**AC28 — D-000.12 and D-000.14: measurement discipline.** **Never verify bytes, hashes or panic
stacks through a shell pipe** (D-000.12) — use `rtk proxy` writing to a file. **And when a count is
load-bearing, count by AST** (D-000.14, promoted from this story's M-1, binding):

> ***"A text-matched count measures the regex; a filtered count measures the filter."***

**Both bit during story creation:** piping `rtk`'s output into `head` produced a Rust broken-pipe
panic mid-listing, and the same quantity read **19** (AST), **36** (raw text) and **8** (piped) —
see M-1. Every measurement in this file that mattered was re-taken with `rtk proxy` into a file, or
with a Go/Python program that parses.

**AC29 — D-000.4, stated explicitly so nobody reads silence as green.** Heavy-test cadence is
**per-epic**. **Story 1.7 is NOT one of D-000.4's override stories** (those are 1.2, 1.5, 1.8, 2.4,
4.7 — a hash-shaped deliverable). **The cross-target hash matrix (`linux/amd64`, `linux/arm64`
under Docker, `js/wasm` under Node, against local `darwin/arm64`) is DEFERRED to the Epic 1 boundary
and is NOT run in this story.** Unit tests, `go vet`/lint and the build run on this story
regardless, and this story does not reach `done` on a red unit suite. **Do not imply the matrix is
green; it is not run.**

**AC30 — D-1.3.1 / AD-1.** Any new or modified non-test file under `folio-go/` inherits AD-1's full
ban (`time`, `os`, `math/rand`, `net`, and `math` transcendentals by selector) and D-1.3.1's exact
`_test.go` exemption (`os`, `testing`, `path/filepath`, `embed` — **and nothing else**). **`io` is
NOT banned** — verified against `bannedImportPaths` in `lint/internal/rules/forbiddenimports.go:31`,
which holds exactly `time`, `os`, `math/rand`, `net` — so `RenderTo(w io.Writer, …)` may live in the
render entry file without touching the fence.

**AC31 — D-000.10.** The developer dispositions **every ruling ID this story enumerates**, one line
each, in the table at [§ Ruling disposition table](#ruling-disposition-table-d-00010); the reviewer
mirrors it independently. **The enumeration is explicit and complete in that table.**

---

## Decisions resolved

### D-1.7.7 — AD-7's `/CreationDate` clause is neither a gap nor a contradiction: Story 3.7 owns it, with acceptance criteria already written

**Orchestrator decision**, on the lead's ruling, resolving the `DECISION NEEDED` this story's
creator surfaced. **The item is not parked.**

**What the creator surfaced.** AD-7 (`ARCHITECTURE-SPINE.md:201`, binding) says `/CreationDate` and
`/ModDate` are *"omitted **unless a date arrives through `params`**"*. Until this story `params` did
not exist, so the clause's condition was unreachable; 1.7 creates the trigger, and the shipped
`TestRenderHasNoCreationOrModDate` asserts absence **unconditionally**.

**The resolution: it is scheduled, owned and specified.** Verified in `epics.md:1074–1100` —
**Story 3.7, "Validate a template and render it from the command line"**:

- Its user story reads *"…and to **pin document dates reproducibly**"*.
- Its **Covers** line names **AD-7** explicitly (`Covers: FR42 · AD-7`).
- It carries these two ACs verbatim:

  > **Given** `SOURCE_DATE_EPOCH` set in the environment · **When** `cmd/folio render` runs ·
  > **Then** the CLI reads it and passes it in as a parameter · **And** the library core still reads
  > no environment variable
  >
  > **Given** **no date supplied by any route** · **When** the PDF is produced · **Then**
  > `/CreationDate` and `/ModDate` are omitted

**The lead's tell, worth carrying** *(illustrative)*: ***"no date supplied by any route" is the
negative case of a positive case.*** **Nobody writes that sentence unless a route exists.** 3.7 is
also the story that builds `cmd/folio` — the only component NFR1.f names as reading
`SOURCE_DATE_EPOCH`. **So the clause was never orphaned; the creator found the trigger before the
owner, which is the right order.**

**The footer-vs-metadata distinction is real, but it is NOT the disposal — recorded so nobody
reuses it as one** *(binding)*. The two surfaces genuinely differ: `acceptance.md:16`'s generated
date is **footer text**, which 1.7's `params` binding delivers; AD-7's clause governs the
**metadata dictionary**. **But NFR1.f is a stated requirement** (`prd.md:370`, verified: *"Creation
and modification dates and the document ID are either omitted or derived from content… The
`SOURCE_DATE_EPOCH` convention is honoured instead by callers and by Folio's own command-line
tooling, which read it and pass the date in as a parameter (FR21)"*), and `SOURCE_DATE_EPOCH`
exists for exactly one purpose. **A route that reads it and delivers it nowhere would leave NFR1.f
honoured in letter and void in effect.** Both surfaces are real; **neither substitutes for the
other.**

**Why 3.7 rather than 1.7** *(illustrative)*. A PDF date is `D:YYYYMMDDHHmmSSOHH'mm'`, and date
formatting matures at **Story 3.4**. A PDF date emitter at 1.7 — **before any date type, before
`internal/diag`, before the CLI that supplies the value** — would be premature in three directions
at once.

**1.7's entire share is two mechanical items**, both binding, at
[§ Section F](#f--handing-creationdate-to-story-37-mechanically-d-177--ad-7-nfr1f--dw-10):
**AC24** re-scopes the absence test so it asserts 3.7's negative case *deliberately*, and **AC25**
adds a `folio-go/cmd/` absence tripwire so the CLI cannot be created without the wiring being
settled. **1.7 emits no PDF date and builds no date type.**

**De-risking note for Story 3.7** *(illustrative, recorded because it changes how 3.7 is scoped)*:
**the creator's "every golden hash moves" fear is smaller than it looks.** Only fixtures that
**supply a date** would move. A params-carrying render with **no** date is **byte-identical to
today**. So 3.7's blast radius is **new fixtures plus any existing fixture that opts in — not the
corpus.**

**Registered as DW-10**, owner **Story 3.7**. *(DW-10, like the D-1.7.x rulings, is an uncommitted
working-tree addition — see the banner at the top of this file.)*

### D-000.14 — When a count is load-bearing, count by AST

**Orchestrator decision**, promoting this story's M-1 to a **program-wide standing rule**
*(binding)*.

> **Verdict:** a count that any AC, ruling or task depends on is derived from a **parsed AST**,
> never from text matching, and never from a piped shell filter. ***"A text-matched count measures
> the regex; a filtered count measures the filter."***

**The measurement that produced it** (M-1): the same quantity read **19** (AST), **36** (raw text —
doc comments and `t.Fatalf("Render() error: …")` format strings both match), and **8** (piped
through the `rtk` wrapper). **Story 1.6's equivalent claim was wrong twice in the same way** — its
M-7 said "thirteen call sites" and its own review corrected it to 11. **Two occurrences make it a
pattern, not an incident.**

**Interaction with D-000.12**, stated because they compound: D-000.12 already bans verifying bytes
and hashes through a shell pipe. D-000.14 extends the same reasoning to **counts** — and the `8`
above is a case where **both** failure modes fired at once.

---

## Acknowledgements carried forward

Recorded at the coordinator's direction, because each one is a reusable lesson rather than a
compliment.

- **M-5 — an accepted gap that is measured beats one that is implied.** D-1.7.3's prose said a
  file-scoped import rule *"does not stop a deliberate cross-file route"*. This story **built one
  and watched it pass**, rather than accepting the sentence. **A reviewer now meets the gap in the
  story instead of discovering it.** Keep it recorded as a **capability fence, not a proof**
  (AC11).
- **M-8 — the lead's own miss, owned plainly.** D-1.7.5 demanded a compile-time proof for `Params`
  while Story 1.6 shipped `Data`'s defined-type property carried by **a doc comment only** — exactly
  what the ruling refuses one story later. **It asked what 1.6 *declared*, not what it *asserted*.**
  **AC19a covering both types is right**, and the general lesson is: *a property a story claims and
  a property a story asserts are different facts; check the second.*
- **M-1 — promoted to D-000.14 above.** The counting spread is now program-wide measurement
  discipline, not a story note.

## Tasks / Subtasks

- [x] **Task 1 — `Params`, its decode path, and the swap proof** (AC: 18, 19, 19a, 19b, 20)
  - [x] Declare `type Params []byte` in `folio-go/render_entry.go`, **defined, not aliased**, with a
        doc comment mirroring `Data`'s and citing D-1.7.5.
  - [x] Decode params through `internal/bind.DecodeData` — **no new decoder**. Confirm by
        inspection that no `json.NewDecoder` appears outside it.
  - [x] Add the guard asserting exactly one `UseNumber` site under `internal/bind`; **zero sites is
        a hard failure**, not a pass (AC26 Q1).
  - [x] Add the compile-time swap assertion covering **both** `Data` and `Params` (AC19a — measure
        first that no such assertion exists today; M-8 says it does not).
  - [x] Red-proof **RP-1**: change `Params` to `type Params = []byte`; the swap assertion must go
        red **naming the assignability**, and the control (unchanged types) must be green. Revert.

- [x] **Task 2 — the `params` resolution root** (AC: 12, 13, 14, 15, 16, 17, 17a)
  - [x] Thread a second root into `internal/bind`'s text binder. **Do not touch
        `reservedPlaceholders`** (AC12) — a reviewer checks this by reading the map, not the tests.
  - [x] First segment `params` selects the params document unconditionally (AC13).
  - [x] Retained fixture in D-1.6.5's shape: **two data documents both carrying a top-level
        `params` key with different values**, one supplied params document; assert equality **and**
        the params value (AC14).
  - [x] Absent-params error names the **params root** and the element id (AC16) — not *"absent from
        the report data"*.
  - [x] Bare `{{params}}` → located error (AC17).
  - [x] Re-run the existing `{{page}}`/`{{pages}}` reservation tests **in the same gate** (AC17a).
  - [x] Red-proof **RP-3**: make the params root fall back to report data on miss; AC14's fixture
        must redden **on the value comparison**, naming the hijacked value — not on an unrelated
        `err != nil` check (D-000.13; this is exactly QA Nit 15's trap).
  - [x] Confirm at baseline, before the fix, that AC14 and AC16 are **red** (M-6 records the
        pre-fix behaviour; re-measure rather than trusting it).

- [x] **Task 3 — `Render` gains `p`, `RenderTo` is born** (AC: 1, 1a, 2, 3)
  - [x] Insert `p Params` between `d` and `f` in `Render`; update all **19** call sites (M-1).
        **Never reorder** (D-1.5.5).
  - [x] Declare `RenderTo` in the **same file as `Render`** — `folio-go/render_entry.go` — with the
        five-argument form. `io` is not a banned import (AC30).
  - [x] `RenderTo` builds the complete document via the one core, then writes it once.
  - [x] Write the anti-streaming comment (AC3): AD-7's `/ID` over the serialized body **and** the
        classic xref's byte offsets, citing D-1.7.2 by id.
  - [x] `go vet` and build clean.

- [x] **Task 4 — the writer-path assertions that can actually fail** (AC: 4, 4a, 5, 6, 7, 8)
  - [x] AST test: exactly one document-byte producer; both entry points route through it. Vacuity
        guard: zero producers or zero routing entry points is a **hard failure** (AC4a).
  - [x] **Do not write a behavioural `Render`-vs-`RenderTo` byte comparison.** If one exists from a
        draft, delete it (AC4, D-1.7.2 binding).
  - [x] Error-mid-write test (AC5), short-write test (AC6), counting-writer byte-count test (AC7),
        payload-concatenation + call-count test (AC8).
  - [x] Red-proof **RP-2**: give `RenderTo` its own second byte-producing path (valid syntax,
        forbidden semantics); the AST test must redden **naming both producing functions**.
  - [x] Red-proof **RP-4**: make `RenderTo` ignore `w.Write`'s returned `n` and `err`; AC5 and AC6
        must **both** redden, and AC7 must redden on the **count**, naming the shortfall.
  - [x] Red-proof **RP-5**: append a trailing `\n` after the document; AC8 must redden naming the
        extra byte **and** the extra call.

- [x] **Task 5 — the world-reading fence, re-proved** (AC: 9, 10, 11)
  - [x] Confirm by reading `findRenderDeclaringFiles` that the guard still locates **by AST**
        (M-3) and that it now collects **both** `Render` and `RenderTo`.
  - [x] Red-proof **RP-6** (the D-1.7.3 mutation): move `RenderTo` into `folio.go`. **Assert the
        build succeeds** (control: valid syntax) **and** that the guard reddens with a finding
        naming `folio.go` and `"os"`. Revert. M-4 records the exact expected finding text.
  - [x] Record the residual cross-file gap (AC11) in a code comment on the guard **and** in the
        README's determinism section — as a **capability fence, not a proof**. M-5 measured it
        green.

- [x] **Task 6 — the README** (AC: 21, 22, 22a, 23, 23a)
  - [x] Write `README.md` — **the repository's first** (M-9). Load call + render call + writer call.
  - [x] The first-PDF example assembles a `FontSet` from the caller's own bytes and **says in one
        line that Story 2.2 removes this step** (AC22, DW-9).
  - [x] Golden-hash section carrying the **toolchain caveat** (AC23), with M-10's measurement stated
        plainly enough that an integrator can act on it.
  - [x] Document that a top-level `params` key in report data is legal but unreachable (AC15).
  - [x] Do not restate D-1.1.a's `go`-below-`toolchain` gap as a defect (AC23a).
  - [x] **If the example reads as ceremony for any reason other than the missing font set, raise
        `DECISION NEEDED` — do not change the signature** (AC22a, binding).

- [x] **Task 7 — hand `/CreationDate` to Story 3.7, mechanically** (AC: 24, 25, 25a, 25b)
  - [x] Rename and re-scope `TestRenderHasNoCreationOrModDate` so its name and failure message say
        **"a render whose params supply no date"**, citing Story 3.7 as the owner of the positive
        case (AC24). Add at least one case whose params are **non-empty but carry no date key**.
  - [x] **Before touching `absenceChecks`:** re-confirm `ScanAbsences`' coverage witness is live —
        `AbsencesStats.ChecksEvaluated`, `TestAbsencesProductionScan`'s hard failure on `== 0`, and
        `TestAbsencesZeroWitnessIsCaught`. **Record the check** (AC25a, D-1.4.11). M-12 measured it
        live; verify, do not inherit.
  - [x] Add the `folio-go/cmd` row: rule id `absence-cmd-dir`, keyed on the **directory** (AC25).
  - [x] Red-proof **RP-8**: create `folio-go/cmd/` with a file in it, run the absence gate
        `-count=1`, assert the finding carries **rule id `absence-cmd-dir`** and names the path.
        Remove it. Control: gate green **and `ChecksEvaluated` equals the row count** (AC25b).
  - [x] **Emit no PDF date and build no date type.** If the work starts to look like a date
        emitter, stop — that is Story 3.7's, and D-1.7.7 says why.

- [x] **Task 8 — gates, dispositions and handoff** (AC: 26, 27, 28, 29, 31)
  - [x] `go test ./... -count=1` green in `folio-go`, `lint` and `hashmatrix`; `go vet` clean; build
        clean. **`-count=1` on every gate** (AC27).
  - [x] Every byte/hash/panic verification through `rtk proxy` into a file (AC28).
  - [x] Fill the [§ Ruling disposition table](#ruling-disposition-table-d-00010) — **one line per
        enumerated ruling id, complete** (AC31).
  - [x] State in the Delivery Log that the **cross-target matrix was not run** and why (AC29).
  - [x] Set status → `review`.

---

## Dev Notes

Every figure below was measured at baseline `3463119` during story creation. **Re-measure rather
than trusting these** — they are recorded so a green run cannot be mistaken for "nothing to prove".

### M-1 — the signature blast radius is 19 call sites, all in tests

Counted with a `go/ast` walk over `folio-go/` (excluding `testdata/` and `.matrix-build/`), matching
`*ast.CallExpr` whose `Fun` is an `*ast.Ident` named `Render` or `RenderTo`:

| File | Call expressions | Lines |
|---|---|---|
| `folio-go/render_test.go` | 9 | 124, 214, 573, 610, 641, 732, 746, 757, 857 |
| `folio-go/render_bind_test.go` | 8 | 70, 87, 104, 117, 131, 163, 214, 226 |
| `folio-go/fixture_test.go` | 1 | 396 |
| `folio-go/template_test.go` | 1 | 182 |
| **Total** | **19** | — |

**Zero production callers.** `folio-go/matrix_test.go` mentions `Render(` in text but contains **no
call expression** — a grep-based count would have over-reported it.

**Why the AST count, and AC28's point in miniature.** A naive
`grep -rn 'Render(' … | grep -v 'func Render' | wc -l` returned **8** under the `rtk` wrapper and
**36** under raw text matching (doc comments and `t.Fatalf("Render() error: …")` format strings both
match). **19 is the AST-exact figure.** Story 1.6's M-7 recorded "thirteen call sites" and its own
review corrected it to 11 — the same class of error, twice. Counted by parser here.

### M-2 — the one core is `renderDocument`

`folio-go/render.go` declares exactly one function returning document bytes:

```go
func renderDocument(t *Template, data bind.Value, fs FontSet) ([]byte, error)
```

`Render` (`render_entry.go:92`) is its only caller today. AC4's AST assertion targets this shape.
Note `internal/pdf` exports two byte-producing entry points (`SerializeTextDocument`, and
`Serialize` on Story 1.1's throwaway fixture document) — **the AST assertion must be scoped to
package `folio`'s own entry points**, or it will find three producers and redden for the wrong
reason (AC26 Q3).

### M-3 — Story 1.6's finisher already rewrote the guard to locate by AST

`lint/internal/rules/forbiddenimports_test.go:209` — `TestRenderEntryFileHasNoForbiddenImports` —
now calls `findRenderDeclaringFiles`, which parses every non-test `.go` file directly under
`folio-go/` and collects those declaring a top-level `Render` or `RenderTo`. Its own comment records
the blocker:

> *"the original version hard-coded `folio-go/render_entry.go` as the target. Moving `func Render`
> back into `folio.go` (which imports `"os"`), leaving `render_entry.go` in place holding only
> `type Data` and a doc comment, left this test **PASSING**."*

**So D-1.7.3's requirement is already met at baseline.** This story's obligation is to **prove it
survived the move it performs**, and to extend the guard's coverage to the case where `Render` and
`RenderTo` land in **two** files (the helper already handles this — it returns a sorted slice and
scans each).

### M-4 — the D-1.7.3 red-proof fires at baseline, measured not predicted

Applied the exact D-1.7.3 mutation: appended a `RenderTo` declaration to `folio-go/folio.go` (the
`os`-importing file), delegating to `Render` and writing to a writer.

**Control (D-000.13: valid syntax, forbidden semantics):**

```
cd folio-go && go build ./...   →  exit 0
```

**The guard, `-count=1` (D-000.11):**

```
=== RUN   TestRenderEntryFileHasNoForbiddenImports
    forbiddenimports_test.go:226: AC12: the file declaring Render must import none of
    os/time/net/math/rand, found:
        folio.go:6: forbidden import "os" (AD-1's allow-listed numeric surface: + - * /,
        comparison, and Sqrt, Floor, Ceil, Round, Trunc, Abs, Mod)
--- FAIL: TestRenderEntryFileHasNoForbiddenImports (0.00s)
```

**All three AC26 questions answered.** It ran (the control built). The unmutated baseline is green.
And it failed **for the reason it names** — the finding identifies `folio.go:6` and the offending
import, not merely a non-zero exit. Mutation reverted; `git status` confirmed clean.

### M-5 — the residual cross-file gap is real and GREEN

Second probe: `RenderTo` left **in `render_entry.go`** (the clean file), routing through a helper
declared in `folio.go`:

```go
func sinkHelper(b []byte) error { return os.WriteFile("/tmp/folio-gap-probe.pdf", b, 0o600) }
```

```
go build ./...                                    →  exit 0
TestRenderEntryFileHasNoForbiddenImports (-count=1) →  PASS
```

**A `RenderTo` that writes a temporary file to disk passes the guard.** This is D-1.7.3's accepted
gap, now measured rather than predicted. AC11 requires it recorded so a later reader does not
mistake the guard for a proof of source AC1's "no temporary file created". Probe reverted, probe
file removed.

### M-6 — params shadowing at baseline: what is red today, and why AC14's fixture must collide

Ran through `internal/bind` directly (temporary test file, removed afterwards; `git status`
confirmed clean):

| Input | Baseline result |
|---|---|
| `{{params.reportDate}}`, data `{"params":{"reportDate":"SHADOWED-FROM-DATA"}}` | `"SHADOWED-FROM-DATA"`, **err = nil** |
| `{{ params.reportDate }}` (spaced), same data | `"SHADOWED-FROM-DATA"`, **err = nil** |
| `{{params.reportDate}}`, data `{}` | err = `bind: element el-2: data path "params.reportDate" is absent from the report data` |
| `{{params}}`, data `{}` | err = `bind: element el-3: data path "params" is absent from the report data` |
| `{{page}}`, data with a top-level `params` key | `"{{page}}"` — reservation intact |

**Three things follow.**

1. **AC13/AC14 are genuinely red at baseline** — report data shadows the namespace completely, with
   no error. The story has real work to do.
2. **The whitespace-tolerant spelling shadows identically**, so the fix belongs at path-segment
   level (`path[0] == "params"`), not on the trimmed literal.
3. **AC14's fixture MUST carry a colliding top-level `params` key in data.** Without the collision,
   an implementation that simply *merges* params into the data root would pass — and the test would
   assert nothing about precedence. This is exactly the trap QA Nit 15 caught in D-1.6.5's `page`
   fixture, which compared against data with **no** `page` key and so reddened on an unrelated
   `Absent` error. **Copy the corrected shape, not the original.**

Also note AC16's message: after this story, *"absent from the report data"* is **wrong** for a
params path — the value was never sought in report data.

### M-7 — the swap proof, and what it does and does not cover

Probed in a scratch module with two defined types and one alias.

**Control (must compile):** defined `Data`/`Params`, correct argument order → `go vet` exit 0.

**The swap:**

```
vet: ./defined.go:26:9: cannot use p (variable of slice type Params) as Data value
                        in argument to Render
```

**The alias half — the thing D-1.7.5 forbids:** with `type ParamsAlias = []byte`, a plain `[]byte`
is accepted where `ParamsAlias` is expected **and vice versa**, both directions compiling cleanly.
**The alias destroys the property exactly as D-1.7.5 says.**

**The gap (AC19b):** `Render(t, Data(p), Params(d), f)` — an explicit conversion — **still
compiles** with defined types. The assertion proves the *accidental* swap is a compile error, not
that a deliberate cast is impossible.

### M-8 — no assertion that `Data` is a defined type exists today

Searched `folio-go/render_test.go`, `gomod_test.go` and `internal/arch_test.go` for a `go/types`,
`reflect`, or AST-based alias check on `Data`. **None exists.** Story 1.6's AC23 is carried **only**
by prose in `render_entry.go:42-52`:

> *"A named, defined type — never `[]byte` and never a type alias (AC23)"*

A comment claiming a safety property is precisely what D-1.7.5 refuses (*"An assertion that the
safety property holds, rather than a comment claiming it does"*). **AC19a's assertion is the first
real one, and it must cover `Data` too** — otherwise 1.7 protects the newer type and leaves the
older one on a comment.

### M-9 — verified: no README, and no `folio-go/fonts/`

Both premises D-1.7.1 and DW-9 rest on, re-verified rather than inherited:

- `ls README* folio-go/README*` → **no matches anywhere in the repository.** This story writes the
  first.
- `folio-go/fonts/` → **No such file or directory.** The only `fonts` directories under the repo are
  test fixtures: `folio-go/testdata/fonts/`, three under `folio-go/testdata/lint/embed-font/`, and
  `folio-go/testdata/lint/absences/violating/folio-go/fonts` — **a lint fixture that mimics the path,
  not the real directory.** Anyone grepping for `folio-go/fonts` will hit that decoy; it is not
  evidence the shipped set exists.

AD-8's Rule requires `folio-go/fonts/` to be the one place font binaries live
(`ARCHITECTURE-SPINE.md:214`), and it arrives at **Story 2.2**. **DW-9's premise holds.**

### M-10 — the toolchain caveat is empirically true, with a control that fires

D-1.7.6's claim was proved by construction rather than read out of documentation. Local toolchain:
`go1.26.0 darwin/arm64`, `GOTOOLCHAIN=auto`.

**Probe** — a dependency module carrying an impossible pin, `toolchain go1.99.0`, consumed by a main
module with no `toolchain` line:

```
go build ./...   →  exit 0        (the dependency's toolchain line was IGNORED entirely)
```

**Control** — the *same* `toolchain go1.99.0` line moved into the **main** module:

```
go: downloading go1.99.0 (darwin/arm64)
go: download go1.99.0 for darwin/arm64: toolchain not available
go build ./...   →  exit 1
```

**The probe is not vacuous** (AC26 Q2): an identical directive binds in the main module and is
ignored in a dependency. **`folio-go`'s `toolchain go1.26.0` therefore does nothing for an
integrator who compiles it into their own binary** — their own main module's `toolchain` line, or
`GOTOOLCHAIN`, decides. That is what the README must say, next to the golden hashes, so an
integrator recording Folio's hashes knows what actually pins them.

`folio-go/go.mod` for reference (`go` deliberately **below** `toolchain`, D-1.1.a):

```
go 1.25.0
toolchain go1.26.0
```

### M-11 — `io` is not on the banned list

`lint/internal/rules/forbiddenimports.go:31` — `bannedImportPaths` holds exactly `time`, `os`,
`math/rand`, `net`, and `isBannedImportPath` also matches subpackages (`net/http`, `os/exec`,
`math/rand/v2`). **`io` matches nothing.** `RenderTo(w io.Writer, …)` may live in the render entry
file without weakening AD-1's fence. Stated because a developer might otherwise move `RenderTo` out
of the fenced file "to be safe" — which is the exact move D-1.7.3's red-proof exists to catch.

### M-12 — `ScanAbsences`' coverage witness is LIVE, so AC25's row is not decorative

D-1.4.11's sequencing precondition, checked before proposing the new row rather than after
(`lint/internal/rules/absences.go`, `absences_test.go`):

| Element | State at baseline |
|---|---|
| `ScanAbsences` returns `AbsencesStats` | Yes — `ChecksEvaluated`, incremented once per row **inside** the loop body (`absences.go:108`) |
| `TestAbsencesProductionScan` fails on an empty scan | Yes — hard `t.Fatalf` on `stats.ChecksEvaluated == 0`, **before** the findings check |
| The witness has its own red-proof | Yes — `TestAbsencesZeroWitnessIsCaught` swaps in an empty `absenceChecks` via `t.Cleanup`, **not** a reverted production mutation |
| Existing rows | 4 — `folio-designer`, `folio-go/fonts`, `folio-go/internal/expr`, `folio-go/internal/diag` |

**Verdict: the witness is live; adding a row is safe now.** Story 1.4's own note is the reason to
check rather than assume (`absences.go:56-58`): *"adding entries to a list whose emptiness passes is
exactly how a tripwire becomes decorative."*

**One sharpening, offered not mandated** *(illustrative)*: the production test asserts
`ChecksEvaluated != 0`, **not** `ChecksEvaluated == len(absenceChecks)`. Today those are equivalent
because the loop body is unconditional — but a future `continue` guard would skip a row silently and
the witness would still read healthy. If AC25b's control is written as *"`ChecksEvaluated` equals
the row count"* (as specified), the stronger form comes for free and the gap closes.

### M-13 — verified: `folio-go/cmd/` does not exist

`ls -la folio-go/cmd` → **No such file or directory.** AC25's tripwire therefore starts green, and
`cmd/folio` — which Story 3.7 builds, and which NFR1.f names as the component that reads
`SOURCE_DATE_EPOCH` — **cannot be created without tripping it.** That is the mechanical trigger
D-1.7.7 relies on: the wiring gets settled because the build stops, not because someone remembers.

---

## Ruling disposition table (D-000.10)

**One line per ruling id this story enumerates, complete.** The developer fills the disposition
column; the reviewer mirrors it **independently** in their own copy of this table.

| Ruling | What it binds here | Developer disposition |
|---|---|---|
| **D-1.7.1** | Five/four positional args, no options struct; AC4 ceremony deferred to 2.2 | Applied as stated. `Render(t, d, p, f)`/`RenderTo(w, t, d, p, f)` declared verbatim, no options struct. README's first-PDF example assembles a `FontSet` from caller bytes and names Story 2.2 as removing that step (DW-9). |
| **D-1.7.2** | Streaming impossible + comment; structural delegation assertion; delete the vacuous comparison; writer-path tests | Applied as stated. Anti-streaming reasoning (AD-7 `/ID`, xref offsets) is a doc comment on `RenderTo` citing D-1.7.2. No behavioural byte comparison written. `render_arch_test.go`'s AST test asserts the structural property; RP-2 red-proofed. `renderto_test.go` covers AC5-AC8, RP-4/RP-5 red-proofed. |
| **D-1.7.3** | Guard locates by AST; red-proof is the move mutation; residual gap recorded | Applied as stated. `findRenderDeclaringFiles` (Story 1.6) confirmed unchanged and now covers both `Render`/`RenderTo`. RP-6 re-run: mutation builds, guard reddens naming `folio.go:8` and `"os"`, matching M-4. AC11's residual gap re-measured (M-5 probe re-run) and recorded in code comments (`render_entry.go`, `forbiddenimports_test.go`) and the README's "Reproducible bytes" note is not the right place for it — recorded instead as a capability-fence note next to `RenderTo`'s doc comment. |
| **D-1.7.4** | `params` is a root not a reserved token; precedence; retained fixture; absent/bare errors | Applied as stated. `reservedPlaceholders` untouched (still exactly `page`/`pages`); `params` handled as a distinct first-segment branch in `BindText`. Retained fixture `TestBindTextParamsRootTakesPrecedenceOverData` copies the corrected page-reservation shape (two colliding data docs + one params doc, asserts value not just equality). Absent/bare errors located and root-named (AC16/AC17). |
| **D-1.7.5** | `Params` defined not aliased; compile-time swap proof; one decode path | Applied as stated. `type Params []byte` (defined). `TestParamsDataSwapDoesNotTypeCheck` builds `testdata/swapproof/{bad,good,convert}` fixtures via `go build`/exec.Command, proving the swap fails to compile (RP-1 red-proofed with an alias mutation) and covers both `Data` and `Params` (AC19a). One decode path confirmed by `TestExactlyOneUseNumberSiteUnderBind` (AC20). |
| **D-1.7.6** | README toolchain caveat near the golden hashes | Applied as stated. "Reproducible bytes — the toolchain caveat" section in `folio-go/README.md`, adjacent to the golden-hash discussion, states the main-module-only binding of `toolchain` and that an integrator's own toolchain/`GOTOOLCHAIN` decides, per M-10. |
| **D-1.7.7** | `/CreationDate` belongs to Story 3.7; 1.7 re-scopes the absence test and adds the `cmd/` tripwire; **emits no PDF date** | Applied as stated. `TestRenderHasNoCreationOrModDate` renamed to `TestRenderWithNoDateInParamsOmitsCreationAndModDate`, re-scoped with a "params present, no date key" case, citing Story 3.7 as owner of the positive case (AC24). `absence-cmd-dir` row added (AC25), red-proofed (RP-8). No PDF date type or emitter was built. |
| **D-000.14** | Load-bearing counts derived by AST, never by text match or shell filter | Applied as stated. All 19 `Render(` call sites named in M-1's AST-derived table were updated; `go vet ./...` (a full type-check, strictly stronger than an AST count) confirmed no call site was missed or miscounted — a leftover unmigrated call site would have failed to compile. `TestExactlyOneUseNumberSiteUnderBind` and `render_arch_test.go`'s producer/routing counts are themselves AST-derived guards, not text-matched. |
| **D-000.4** | Cross-target matrix DEFERRED to the Epic 1 boundary; not run here | Applied as stated. Not run — see Delivery Log. |
| **D-000.9** *(+ extension)* | Every guard and every red-proof answers Q1 and Q2 | Applied as stated. Every red-proof in this story (RP-1 through RP-8) was run against BOTH a construction/control (build succeeds / baseline green) and the mutation (redden), and every AST guard added (`TestExactlyOneUseNumberSiteUnderBind`, `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt`, `findRenderDeclaringFiles`) has an explicit zero-result hard failure. |
| **D-000.10** | This table, complete; reviewer mirrors independently | Applied as stated. All 29 rows dispositioned below. |
| **D-000.11** | `-count=1` on every gate | Applied as stated. Every `go test` invocation in this story's verification (recorded in the Delivery Log) carries `-count=1`. |
| **D-000.12** | No bytes/hashes/panics through a shell pipe; `rtk proxy` into a file | Applied as stated. All red-proof verifications routed through `rtk proxy … > /tmp/*.out 2>&1` then read via a file-reading tool, never piped through `head`/`grep -c` on raw byte output. |
| **D-000.13** | Red-proofs assert rule id + message; controls are valid syntax, forbidden semantics | Applied as stated. Every red-proof (RP-1 through RP-8) asserts on message text (assignability, rule id, byte counts, function names) — none on exit status alone — and every mutation compiled before being run. |
| **D-1.1.a** *(addendum)* | `go` sits below `toolchain`; don't restate the gap as a defect | Applied as stated. README's toolchain section explicitly states the low `go` line is deliberate, not a mistake to fix (AC23a). |
| **D-1.1.c** *(+ addendum)* | Target API shape; `FontSet` shorthand in the source AC; medium-confidence flag now CLOSED | Applied as stated. Five/four-argument shape shipped verbatim; AC1a records the source AC's `FontSet`-omitting shorthand as recorded, not a contradiction. |
| **D-1.2.6** | Conflicts between rulings surfaced, never arbitrated in-story | Applied as stated. No new conflict between two rulings was encountered during implementation; none surfaced. |
| **D-1.3.1** | AD-1 ban in non-test files; exact `_test.go` exemption | Applied as stated. `render_entry.go`/`folio.go` unchanged import posture (verified via RP-6); no new non-test file under `folio-go/` imports a banned path; `testdata/swapproof/*/main.go` files are `package main`, not `_test.go`, and import only `folio` itself. |
| **D-1.3.4** | Absence tripwires, never a conditional "check it if present" | Applied as stated. The new `folio-go/cmd` row follows the existing unconditional-check shape in `ScanAbsences`; no "skip if missing" branch added. |
| **D-1.4.11** | Verify `ScanAbsences`' coverage witness is live **before** adding a row | Applied as stated. Verified live before editing `absenceChecks`: `AbsencesStats.ChecksEvaluated` present, `TestAbsencesProductionScan` hard-fails on `ChecksEvaluated == 0`, `TestAbsencesZeroWitnessIsCaught` red-proofs it by construction (all read directly from `absences.go`/`absences_test.go` before the row was added). |
| **D-1.4.14** | Two emitters would be the drift hazard — one core is correct | Applied as stated. `renderDocument` remains the single core; `render_arch_test.go` proves both `Render` and `RenderTo` route through exactly one producer. |
| **D-1.5.5** | Parameters inserted, never reordered | Applied as stated. `p Params` inserted between `d Data` and `f FontSet` at all 19 call sites; none reordered. |
| **D-1.5.9** | A public entry point documented as ignoring its argument is worse than a located error | Applied as stated (inherited, unmodified): `errNilTemplate` behaviour on `Render`/`RenderTo` unchanged by this story. |
| **D-1.6.1** | The single literal splitter; `Decimal` reused for params | Applied as stated. Params values route through `internal/bind.DecodeData` → the same `Value`/`Decimal` machinery `AsDecimal` already uses; no second splitter introduced. |
| **D-1.6.3** | The world-reading fence this story's move re-tests | Applied as stated. RP-6 re-proves it survived `RenderTo`'s addition to the same file. |
| **D-1.6.5** | The `page`/`pages` reservation params must not be folded into; the fixture shape to copy | Applied as stated. `params` implemented as a separate first-path-segment branch, not added to `reservedPlaceholders`; `TestBindTextPageAndPagesUnaffectedByParamsRoot` confirms no interaction. |
| **D-1.6.6** | Bound checks apply to constructed values — inherited free via the shared decode path | Applied as stated (inherited, unmodified): params values are `bind.Value`s produced by the same `DecodeData`/`AsDecimal` path data already uses, so existing bound-value checks apply without new code. |
| **DW-9** | Registered by D-1.7.1; owner Story 2.2; raise before the `v0.1.0` tag | Applied as stated. README's first-PDF example names Story 2.2 as removing the FontSet-assembly step; example does not read as ceremony for any OTHER reason, so no `DECISION NEEDED` was raised (AC22a's escape hatch not triggered). |
| **DW-10** | Registered by D-1.7.7; owner Story 3.7; the `cmd/` tripwire is its trigger | Applied as stated. `absence-cmd-dir` row registered, red-proofed (RP-8); `folio-go/cmd` remains absent at story end. |

---

## Red-proof register

Each row records the **mutation**, the **control**, and **what the finding must say** (D-000.13).

| # | AC | Mutation | Control (must be green / must build) | Required finding |
|---|---|---|---|---|
| RP-1 | 19, 19a | `type Params = []byte` | Defined types compile and assert green | Names the assignability, not just "failed" |
| RP-2 | 4, 4a | Second byte-producing path inside `RenderTo` | Mutation compiles | Names **both** producing functions |
| RP-3 | 14 | Params root falls back to report data on miss | Mutation compiles; baseline fixture green | Reddens on the **value** comparison, naming the hijacked value |
| RP-4 | 5, 6, 7 | `RenderTo` ignores `w.Write`'s `n` and `err` | Mutation compiles | AC5 + AC6 redden; AC7 names the byte shortfall |
| RP-5 | 8 | Trailing `\n` appended after the document | Mutation compiles | Names the extra byte **and** the extra `Write` call |
| RP-6 | 9, 10 | `RenderTo` moved into `folio.go` | **Measured: builds, exit 0** (M-4) | **Measured:** `folio.go:6: forbidden import "os"` (M-4) |
| RP-7 | 20 | A second `json.NewDecoder`/`UseNumber` site under `internal/bind` | Mutation compiles | Names both decoder sites; **zero sites found is a hard failure**, not a pass |
| RP-8 | 25, 25b | Create `folio-go/cmd/` with a file in it | Gate green with it absent, **and `ChecksEvaluated` == row count** | Rule id **`absence-cmd-dir`** + the path — never exit status |

---

## Dev Agent Record

### Context Reference

This story file itself (rulings D-1.7.1–D-1.7.7, D-000.14, DW-9, DW-10 read from the working-tree
copies of `folio-mvp-decision-log.md` and `deferred-work.md`, per the banner at the top of this
file); the existing codebase at baseline `3463119` (`render_entry.go`, `render.go`,
`internal/bind/text.go`, `internal/bind/value.go`, `lint/internal/rules/forbiddenimports.go`,
`lint/internal/rules/absences.go`).

### Agent Model Used

Claude Opus 5 (1M context), via the `bmad-dev-story` workflow.

### Debug Log References

No separate debug log file was kept; all gate output was captured via `rtk proxy … > /tmp/*.out
2>&1` (D-000.12) and inspected with a file-reading tool. Every red-proof's mutation, control, and
revert is recorded in the Ruling disposition table and in the completion notes below; none of the
scratch `/tmp/*.bak` copies used to construct/revert mutations were committed.

### Completion Notes

- **Task 1 (`Params`, decode path, swap proof).** `type Params []byte` declared in
  `render_entry.go`, defined not aliased, doc comment citing D-1.7.5. `decodeParams` decodes via
  `internal/bind.DecodeData` with a zero-length short-circuit to an empty object (`nil`/empty
  Params = "no runtime values supplied", not a decode error). `TestExactlyOneUseNumberSiteUnderBind`
  (AST-based) confirms one decode site. `TestParamsDataSwapDoesNotTypeCheck` builds three fixtures
  (`testdata/swapproof/{good,bad,convert}`) via `go build` in a subprocess, proving the accidental
  swap fails to compile while an explicit conversion still does (AC19b). RP-1 (mutate `Params` to an
  alias) reddened the swap test naming the assignability, then was reverted; control (unmutated)
  verified green both before and after.
- **Task 2 (`params` resolution root).** `internal/bind/text.go`'s `BindText` gained a `params
  Value` parameter; `reservedPlaceholders` is untouched. The first path segment `"params"` is
  checked at the segment level (catching the whitespace-tolerant spelling identically) before any
  data lookup, so it can never be shadowed by data. Six new fixture tests added (precedence, miss,
  absent, bare, top-level-key-in-data-is-legal, page/pages unaffected). RP-3 (params falls back to
  report data on a miss) reddened `TestBindTextParamsMissDoesNotFallBackToData`, naming the
  hijacked decoy value from report data — not merely `err != nil` — then was reverted.
- **Task 3 (signature changes).** `Render` gained `p Params` inserted between `d` and `f`; all 19
  AST-identified call sites (M-1) updated, none reordered. `RenderTo` declared in the same file as
  `Render` (`render_entry.go`), with the anti-streaming doc comment citing AD-7 and D-1.7.2 (`/ID`'s
  content hash, the xref table's byte offsets).
- **Task 4 (writer-path assertions).** `render_arch_test.go` adds an AST-based structural guard
  (`TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt`) with its own vacuity guards
  for zero producers / zero routing entry points; no behavioural `Render`-vs-`RenderTo` byte
  comparison exists. `renderto_test.go` adds AC5-AC8 tests using purpose-built writer doubles
  (`errWriter`, `shortWriter`, `countingWriter`). RP-2 (a second byte-producing path inside
  `RenderTo`) reddened the AST test, naming both `renderDocument` and the mutation function. RP-4
  (ignore `w.Write`'s `n`/`err`) reddened AC5 and AC6, verified in two mutation variants (ignore
  both; check err but not n) to isolate each assertion. RP-5 (append a trailing `\n`) reddened AC8,
  naming both the extra byte and the extra `Write` call. All mutations reverted; every restoration
  verified against the pre-mutation backup and a full green re-run.
- **Task 5 (world-reading fence).** `findRenderDeclaringFiles` (Story 1.6) confirmed unchanged and
  already covers both `Render`/`RenderTo` by AST. RP-6 (move `RenderTo` into `folio.go`) re-run:
  build succeeded (control) and the guard reddened naming `folio.go:8: forbidden import "os"`,
  matching M-4's baseline measurement. AC11's residual gap re-measured (M-5's probe re-run: a
  cross-file `os.WriteFile` helper called from the clean file still builds and still passes the
  guard) and recorded as a code comment on `RenderTo` and in `forbiddenimports_test.go`'s doc
  comment — a capability fence, not a proof. No filesystem-snapshot check was built.
- **Task 6 (README).** `folio-go/README.md` written (the repository's first, confirmed via `ls`
  before writing). Three-call example (load, assemble fonts, render), a `RenderTo`/`http.Handler`
  example, a "Data vs. Params" section documenting AC15's legal-but-unreachable top-level `params`
  key, and a "Reproducible bytes" section carrying the toolchain caveat (D-1.7.6) next to the
  golden-hash discussion, plus D-1.1.a's addendum that the low `go` line is deliberate. The
  first-PDF example's only ceremony is the FontSet-assembly step DW-9 documents; AC22a's escape
  hatch (raise `DECISION NEEDED` if it reads as ceremony for any OTHER reason) was not triggered.
- **Task 7 (`/CreationDate` handoff).** `TestRenderHasNoCreationOrModDate` renamed to
  `TestRenderWithNoDateInParamsOmitsCreationAndModDate`, re-scoped with a case whose params are
  non-empty but carry no date key, citing Story 3.7 as owner of the positive case. Before touching
  `absenceChecks`, `ScanAbsences`'s coverage witness was re-confirmed live by reading
  `absences.go`/`absences_test.go` directly (`ChecksEvaluated`, `TestAbsencesProductionScan`'s hard
  failure, `TestAbsencesZeroWitnessIsCaught`'s construction-based red-proof) — matching M-12. The
  `folio-go/cmd` row (`absence-cmd-dir`) was then added; `TestAbsencesChecksIncludeAllFiveEntries`
  renamed/extended; `TestAbsencesProductionScan` strengthened to assert `ChecksEvaluated ==
  len(absenceChecks)` exactly (M-12's offered sharpening, taken per AC25a). RP-8 (create
  `folio-go/cmd/` with a file in it) reddened `TestAbsencesProductionScan`, naming rule id
  `absence-cmd-dir` and the path; removed, control (absent, `ChecksEvaluated == len(absenceChecks)`)
  re-confirmed green. No PDF date type or emitter was built anywhere in this story.
- **Task 8 (gates, dispositions, handoff).** All gates run and recorded below; the 29-row ruling
  disposition table filled; the cross-target matrix explicitly not run (see Delivery Log).

### File List

**Developer's original list, plus the finisher's additions (marked ⟵ finisher below) resolving
Findings 1-11.**

- `folio-go/render_entry.go` — modified: `Params` type, `decodeParams`, `Render` signature, new
  `RenderTo`. ⟵ finisher: `errNilWriter` + `w == nil` check (Finding 4); `decodeParams` now calls
  `bind.DecodeParams` (Finding 6); doc comment corrected.
- `folio-go/render.go` — modified: `collectTextRuns`/`renderDocument` thread `params bind.Value`.
- `folio-go/internal/bind/text.go` — modified: `BindText` gains `params Value`; `lookupBound`
  helper; params-root resolution branch. ⟵ finisher: the data branch now ALSO routes through
  `lookupBound` instead of a duplicated inline switch (Finding 3) — this is the story's most
  significant post-review change.
- `folio-go/internal/bind/value.go` — ⟵ finisher (new): `DecodeData` split into a shared
  `decodeJSON(d, rootLabel)` plus a new exported `DecodeParams`, so a decode error on the params
  root is reported as "params", never "report data" (Finding 6). Still exactly one
  `json.NewDecoder`/`UseNumber` call site (AC20 unaffected).
- `folio-go/internal/bind/text_test.go` — modified: existing `BindText` call sites updated with a
  `noParams` value; six new fixture tests for AC13-AC17a.
- `folio-go/internal/bind/decodeguard_test.go` — added: AC20's one-decode-path AST guard. ⟵
  finisher: now counts `json.NewDecoder` and `json.Unmarshal` sites alongside `UseNumber` and
  asserts all three, closing the guard's "duplicate that omits UseNumber" blind spot (Finding 2).
- `folio-go/render_arch_test.go` — added: AC4/AC4a's structural producer/routing AST guard. ⟵
  finisher: producer detection now matches ANY call through the `internal/pdf` alias, not only the
  `SerializeTextDocument` selector, so a second producer calling `pdf.Serialize()` is caught
  (Finding 1).
- `folio-go/renderto_test.go` — added: AC5-AC8 writer-path tests and doubles. ⟵ finisher:
  `partialThenErrWriter` + `TestRenderToSurfacesPartialThenErrWrite` (Finding 8);
  `TestRenderToRejectsNilWriter` (Finding 4).
- `folio-go/render_bind_test.go` — ⟵ finisher: `TestRenderMalformedParamsIsReportedAsParams`,
  `TestRenderTrailingGarbageInParamsIsReportedAsParams` (Finding 6).
- `folio-go/paramsswap_test.go` — added: AC19/AC19b compile-time swap proof.
- `folio-go/testdata/swapproof/good/main.go` — added: AC19 control fixture.
- `folio-go/testdata/swapproof/bad/main.go` — added: AC19 swap fixture.
- `folio-go/testdata/swapproof/convert/main.go` — added: AC19b explicit-conversion fixture.
- `folio-go/render_test.go` — modified: 19 `Render(` call sites updated with `p Params`;
  `TestRenderHasNoCreationOrModDate` renamed/re-scoped to
  `TestRenderWithNoDateInParamsOmitsCreationAndModDate` (AC24).
- `folio-go/fixture_test.go`, `folio-go/template_test.go` — modified: `Render(` call sites updated
  with `nil` for `p Params`.
- `folio-go/README.md` — added: the repository's first README (AC21-AC23a). ⟵ finisher: capability
  fence paragraph in the `RenderTo` section (Finding 5); golden-hash location paragraph in
  "Reproducible bytes" (Finding 7); first-PDF example comment corrected to match the code shown
  (Finding 11).
- `folio-go/testdata/lint/absences/violating/folio-go/cmd/placeholder.go` — ⟵ finisher (new):
  makes `absence-cmd-dir`'s red-proof permanent instead of one-shot-only (Finding 10).
- `lint/internal/rules/absences.go` — modified: new `folio-go/cmd` row (`absence-cmd-dir`, AC25).
- `lint/internal/rules/absences_test.go` — modified: `TestAbsencesChecksIncludeAllFourEntries`
  renamed to `TestAbsencesChecksIncludeAllFiveEntries` and extended;
  `TestAbsencesProductionScan` strengthened to assert the exact `ChecksEvaluated` count (AC25a/b).
  ⟵ finisher: `TestAbsencesFixtureScan`'s `want` extended with the new fixture's finding
  (Finding 10).
- `lint/internal/rules/forbiddenimports_test.go` — modified: doc comment recording AC9-AC11's
  coverage and the residual cross-file gap (M-5).

### Delivery Log

**The cross-target byte-identity matrix (`linux/amd64`, `linux/arm64` under Docker, `js/wasm` under
Node, against local `darwin/arm64`) was NOT run in this story, per AC29/D-000.4.** Story 1.7 is not
one of D-000.4's override stories (1.2, 1.5, 1.8, 2.4, 4.7); heavy-test cadence is per-epic and this
matrix is deferred to the Epic 1 boundary. This is stated explicitly so silence is not read as
green.

Gates run, all `-count=1`, all via `rtk proxy` (raw exit codes, D-000.12):

| Gate | Command | Result |
|---|---|---|
| `folio-go` gofmt | `gofmt -l folio-go lint hashmatrix` | clean (no output) |
| `folio-go` build | `go build ./...` | exit 0 |
| `folio-go` vet | `go vet ./...` | exit 0 |
| `folio-go` build (matrix tag) | `go build -tags=matrix ./...` | exit 0 |
| `folio-go` vet (matrix tag) | `go vet -tags=matrix ./...` | exit 0 |
| `folio-go` test | `go test ./... -count=1` | exit 0 — 7 packages ok; **143 top-level tests, 234 with subtests** (re-measured; story's reference figure of 129/220 predates this story's 14 new top-level tests) |
| `lint` build | `go build ./...` | exit 0 |
| `lint` vet | `go vet ./...` | exit 0 |
| `lint` test (`GOPROXY=off`) | `GOPROXY=off go test ./... -count=1` | exit 0 — 4 packages ok, 19 top-level tests |
| `lint` manifest | `go test ./internal/manifest/... -run TestManifestUpToDate -count=1` | exit 0 (no new module dependency; manifest unaffected) |
| `hashmatrix` build | `go build ./...` | exit 1 — pre-existing, unrelated to this story (`go: build output "probe" already exists and is a directory`; `hashmatrix/` has no changes in `git status`, confirming this predates the story) |
| `hashmatrix` vet/test | `go vet ./...` / `go test ./... -count=1` | exit 0 / exit 0 |

Red-proofs constructed and observed red this story (all reverted after observation, `git status`
confirmed clean before/after each):

- **RP-1** (AC19/D-1.7.5): `type Params = []byte` → `TestParamsDataSwapDoesNotTypeCheck` reddens,
  naming the accepted swap; control (unmutated) green.
- **RP-2** (AC4/AC4a/D-1.7.2): a second `pdf.SerializeTextDocument` call inside `RenderTo` →
  `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt` reddens, naming
  `renderDocument` and the mutation function; mutation builds (control).
- **RP-3** (AC14/D-1.7.4): params root falls back to report data on a miss →
  `TestBindTextParamsMissDoesNotFallBackToData` reddens naming the hijacked decoy value; baseline
  fixture green pre-mutation.
- **RP-4** (AC5-AC7/D-1.7.2): `RenderTo` ignores `w.Write`'s `n`/`err` → `TestRenderToSurfacesWriterError`
  and `TestRenderToDetectsShortWrite` redden (two mutation variants used to isolate each check).
- **RP-5** (AC8/D-1.7.2): trailing `\n` appended after the document →
  `TestRenderToWritesNothingExtra` reddens, naming both the extra byte and the extra `Write` call.
- **RP-6** (AC9-AC11/D-1.7.3): `RenderTo` moved into `folio.go` → build succeeds (control);
  `TestRenderEntryFileHasNoForbiddenImports` reddens naming `folio.go:8` and `"os"`, matching M-4.
- **RP-8** (AC25/AC25b/D-1.7.7): `folio-go/cmd/` created with a file in it →
  `TestAbsencesProductionScan` reddens naming rule id `absence-cmd-dir` and the path; removed,
  control green with `ChecksEvaluated == len(absenceChecks)`.
- **M-5** (AC11, re-measured): a deliberate cross-file route (`RenderTo` calling an `os.WriteFile`
  helper declared in `folio.go`) still builds and the guard still passes — confirming the residual
  gap is real, not merely predicted; recorded, not closed (a filesystem-snapshot check was
  deliberately not built).

- **RP-7** (AC20/D-1.7.5): a second `json.NewDecoder`/`UseNumber` site (`rp7MutationProbe`) appended
  to `internal/bind/value.go` → `TestExactlyOneUseNumberSiteUnderBind` reddens, naming both sites by
  file:line (`value.go:133 value.go:195`) — not merely a count — after the guard was strengthened to
  report locations instead of a bare count. Mutation reverted; control (unmutated) re-confirmed
  green, alongside the full `folio-go` suite.

---

### Finisher's Delivery Log addendum

Everything below was run by the story finisher AFTER applying fixes for all eleven review findings
(see [§ Finisher's Finding Resolutions](#finishers-finding-resolutions)). All commands used the
bare `go`/`gofmt` binaries directly (never through a shell pipe for bytes/hashes/panics, D-000.12),
with `GOWORK=off` (no `go.work` at the repo root, verified — this is belt-and-braces, not a fix for
a live problem) and `-count=1` on every test invocation (D-000.11). Raw exit codes are recorded, not
inferred from output text.

| Gate | Command | Result |
|---|---|---|
| `folio-go` gofmt | `gofmt -l folio-go lint hashmatrix` (from repo root) | clean, exit 0 |
| `folio-go` build | `go build ./...` | exit 0 |
| `folio-go` vet | `go vet ./...` | exit 0 |
| `folio-go` build (matrix tag) | `go build -tags=matrix ./...` | exit 0 |
| `folio-go` vet (matrix tag) | `go vet -tags=matrix ./...` | exit 0 |
| `folio-go` test | `go test ./... -count=1` | exit 0 — 7 packages ok |
| `lint` build | `go build ./...` | exit 0 |
| `lint` vet | `go vet ./...` | exit 0 |
| `lint` test | `GOPROXY=off go test ./... -count=1` | exit 0 — 4 packages ok |
| `lint` manifest | `GOPROXY=off go test ./internal/manifest/... -run TestManifestUpToDate -count=1` | exit 0 |
| `hashmatrix` build | `go build ./...` | exit 1 — **re-confirmed pre-existing**: `build output "probe" already exists and is a directory`; `git status --porcelain hashmatrix/` is empty, both before and after this story's fixes |
| `hashmatrix` vet | `go vet ./...` | exit 0 |
| `hashmatrix` test | `go test ./... -count=1` | exit 0 |
| `go.mod`/`go.sum` | `git diff --stat 3463119 -- folio-go/go.mod folio-go/go.sum lint/go.mod lint/go.sum` | empty — unchanged |

**Test counts, re-derived, not carried from the developer's or reviewer's figures**
(D-000.14: AST/binary-derived, never `grep -c` on raw output — this story's finisher additionally
avoided piping through `rtk`'s stdout-shaping for build/test/gofmt, using the bare toolchain
binaries instead, since `rtk` has been observed to reformat these tools' exit signals):
**147 top-level / 238 with subtests**, 0 skips, 0 failures (`go test ./... -list '.*'` → 147
`Test*` names; `go test ./... -count=1 -v` → 238 `=== RUN` / 238 `--- PASS` lines, exactly). This is
the pre-fix figure of **143 top-level / 234 with subtests** plus exactly **4** new top-level tests
this finisher added (`TestRenderMalformedParamsIsReportedAsParams`,
`TestRenderTrailingGarbageInParamsIsReportedAsParams`, `TestRenderToSurfacesPartialThenErrWrite`,
`TestRenderToRejectsNilWriter`) and **0** new subtests — the arithmetic is exact (143+4=147,
234+4=238) and was checked, not assumed.

**The cross-target byte-identity matrix stays explicitly unrun**, per D-000.4 (Story 1.7 is not one
of the override stories 1.2/1.5/1.8/2.4/4.7): only `go build -tags=matrix ./...` and
`go vet -tags=matrix ./...` were run (both exit 0, above); `hashmatrix/`'s actual cross-platform
comparison is deferred to the Epic 1 boundary, unchanged from the developer's original disclosure.

**Every red-proof this finisher's fixes touch was re-run and confirmed to fail for the reason it
names (D-000.13):**

- **F1's fix** (`render_arch_test.go`): re-ran the reviewer's exact escape route — a throwaway
  function calling `pdf.Serialize()` (not `SerializeTextDocument`) — added to package `folio`.
  `go build ./...` exit 0 (valid mutation); `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt`
  now reddens: `AC4: expected EXACTLY ONE document-byte producer, found 2: [renderDocument
  mutantProducerUsesSerialize]`. Removed; `go build ./...` exit 0 again; tree confirmed
  byte-identical via `/usr/bin/diff` before/after.
- **F2's fix** (`decodeguard_test.go`): re-ran the reviewer's exact omitting mutation — a second
  `json.NewDecoder(bytes.NewReader(d))` **without** `.UseNumber()` — appended to
  `internal/bind/value.go`. Builds; `TestExactlyOneUseNumberSiteUnderBind` now reddens: `AC20:
  expected exactly ONE json.NewDecoder call site under internal/bind, found 2: [value.go:154
  value.go:184]`. Reverted; `/usr/bin/diff` confirmed byte-identical; full `internal/bind` suite
  green again.
- **F3's fix** (`text.go`): re-ran **all three** of the reviewer's `lookupBound` mutations from
  Finding 3's table, each against the FULL suite (`go test ./... -count=1`), exactly as the finding
  measured them pre-fix:
  - **wrong-kind arm coerces** (`return "PROBE-B-COERCED", nil` instead of erroring) — pre-fix: full
    suite exit 0. **Post-fix: full suite exit 1**, failing `TestRenderWrongKindIsLocatedError`,
    `TestAD14Triple/row3_wrong_kind_is_error_no_coercion`, and
    `TestBindTextWellFormedNumberIsStillWrongKind` — this is the exact re-run the orchestrator asked
    for by name: **the wrong-kind mutation now reddens, closing the coverage hole.**
  - **`Null` arm returns a hard error** — pre-fix: full suite exit 0. Post-fix: full suite exit 1,
    failing `TestRenderNullPathRendersEmpty`, `TestRenderNullBoundTextStillValidatesFontChain`, and
    `TestAD14Triple/row2_null_renders_empty`.
  - **`AsDecimal` bound check deleted** — pre-fix: full suite exit 0. Post-fix: exit 1, failing
    `TestBindTextNumberCoefficientOverflowIsLocatedError` (now surfaces the more generic "not a
    string" wrong-kind message instead of the coefficient-bound one it must name).

  All three mutations were applied one at a time, run, and reverted (`/usr/bin/diff` against a
  pre-mutation backup confirmed byte-identical each time); the full suite was re-confirmed green
  after each revert.
- **RP-6** (D-1.7.3, unaffected in mechanism but `render_entry.go` and `folio.go` both changed
  content this story): re-ran by moving `RenderTo` (with `errNilWriter`) out of `render_entry.go`
  and into `folio.go`, patching imports so the mutation compiles (D-000.13's "valid syntax,
  forbidden semantics" requirement). `go build ./...` exit 0; in the `lint` module,
  `TestRenderEntryFileHasNoForbiddenImports` reddens: `folio.go:8: forbidden import "os"` —
  matching M-4's original measurement exactly. Reverted; both files confirmed byte-identical via
  `/usr/bin/diff`.
- **RP-8** (D-1.7.7/AC25b, re-run because this finisher's F10 fix added a permanent fixture in the
  same area): created `folio-go/cmd/placeholder.go` at the REAL path (not the fixture root) →
  `TestAbsencesProductionScan` reddens naming `absence-cmd-dir` and the path. Removed → green again.
  The new permanent fixture (`testdata/lint/absences/violating/folio-go/cmd/`) was exercised by the
  ordinary `go test ./... -count=1` run above, not by this transient probe.

**A discrepancy in the disposition table's own arithmetic, found while correcting it.** The
developer's Ruling disposition table claims **29 of 29** rows "Applied as stated". The reviewer's
independent mirror disagrees and says so explicitly — but this finisher counted the reviewer's own
29-row table by its `Agrees?` column (✅/❌/⚠️) rather than trusting the reviewer's prose summary,
and the two do not match: **the reviewer's summary text says "Agreement: 18 of 29 rows… 6
substantively… 5 partially" (11 total disagreements); counting the ❌/⚠️ marks in the reviewer's own
table gives 17 agreements, 8 rows marked ❌ (substantive), and 4 marked ⚠️ (partial) — 12 total
disagreements, not 11.** Both the developer's 29/29 claim and the reviewer's own 18/29 summary
undercount the actual gap by this table's own evidence — exactly the failure mode D-000.10 exists to
catch, one level deeper than the review already found it. See
[§ Finisher's corrected ruling disposition table](#finishers-corrected-ruling-disposition-table)
for the row-by-row count and the post-fix disposition of all 29 rows.

---

## QA Results

> **Finisher's note:** All eleven findings below (0 blockers, 5 majors, 4 minors, 2 nits) are
> FIX-resolved — see
> [§ Finisher's Finding Resolutions](#finishers-finding-resolutions) and the corrected disposition
> table at the end of this file for triage, fixes, and re-run red-proofs. This section is left
> exactly as the reviewer wrote it.

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-23
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 0
- **Majors:** 5
- **Minors:** 4
- **Nits:** 2

**Every figure below was re-measured in this review, not read from the story.** All gates re-run
`-count=1` (D-000.11); every count derived by AST or by the test binary's own `-list`, never by text
match or pipe (D-000.14); every mutation applied to a backed-up copy, reverted by hand, and the
whole tree verified byte-identical to a pre-review snapshot afterwards (`git diff 3463119` diffed
against the snapshot; all seven untracked files `diff`'d individually — **identical**).

### What was re-measured and holds

| Claim | My independent measurement |
|---|---|
| `folio-go` suite | `go test ./... -count=1` → **exit 0, 7 packages** |
| Test counts | **143 top-level / 234 with subtests**, 0 skips (via `-list '.*'` and `=== RUN` on `-v`). Baseline `3463119` in a throwaway worktree: **129 / 220**. The Delivery Log's four numbers are all exact. |
| `lint` suite | `GOPROXY=off go test ./... -count=1` → **exit 0, 4 packages** |
| `hashmatrix` build | **exit 1**, `build output "probe" already exists and is a directory` — reproduced, and `hashmatrix/` is unmodified in `git status`, so pre-existing and correctly disclosed |
| Matrix tag (AC29/D-000.4) | `go build -tags=matrix ./...` → **exit 0**; `go vet -tags=matrix ./...` → **exit 0**. Matrix itself correctly not run; the Delivery Log says so explicitly. |
| `gofmt -l folio-go lint hashmatrix` | clean (bare binary, not through `rtk`) |
| Module graph | `go.mod`/`go.sum` unchanged vs `3463119`; `go list -m all` → 2 lines, unchanged |
| M-1's 19 call sites | AST walk over `folio-go/` (excluding `testdata/`, `.matrix-build/`): the 19 pre-existing sites match M-1's per-file table **exactly** (render_test 9, render_bind_test 8, fixture_test 1, template_test 1). Post-story total **28** = 19 + 8 new in `renderto_test.go` + 1 new production caller (`render_entry.go:200`). None reordered (D-1.5.5). |
| RP-6 (D-1.7.3) | Reproduced. `RenderTo` moved into `folio.go`: `go build ./...` **exit 0** and `go vet` **exit 0** (valid control), guard reddens with `folio.go:8: forbidden import "os"` — matching the Delivery Log verbatim. |
| M-5 residual gap | Reproduced. `RenderTo` left in the clean file calling an `os.WriteFile` helper in `folio.go`: **builds, guard PASSES.** Correctly recorded as a capability fence, not a proof. **Not logged as a defect.** |
| RP-8 (AC25/AC25b) | Reproduced. `folio-go/cmd/` created → `TestAbsencesProductionScan` reddens naming rule id `absence-cmd-dir` and the path; removed → green with `ChecksEvaluated == len(absenceChecks)`. `folio-go/cmd` verified absent at review end. |
| RP-3 (AC14/D-1.7.4) | Reproduced. Params falling back to data on a miss reddens `TestBindTextParamsMissDoesNotFallBackToData` **naming the hijacked decoy** `"SHADOWED-FROM-DATA-A"` — on the value, not on a bare `err != nil`. |
| AC13/14/16/17/17a | Red-proofed by disabling the `path[0] == "params"` branch: **five** params tests redden, each on its own property. Precedence is decided at path-segment level, so `{{ params.reportDate }}` (spaced) is covered identically. `reservedPlaceholders` read directly — still exactly `page`/`pages` (AC12 holds). |
| RP-4 / RP-5 (AC5-AC8) | Reproduced. Ignoring `n`/`err` reddens AC5 + AC6; a trailing `\n` reddens AC7 (`531 vs 530`) **and** AC8 naming both the extra byte and the second `Write` call. I additionally proved AC7's shortfall-naming assertion is reachable (stripping the numbers from the short-write message reddens it specifically) — it is not dead behind the earlier `t.Fatal`. |
| AC19 / AC19a (D-1.7.5) | The swap proof is **real, not decorative**. `testdata/swapproof/good` and `convert` build (exit 0); `bad` fails with two diagnostics naming both directions. I red-proofed **both** halves: `type Params = []byte` **and** `type Data = []byte` each make `bad` compile and redden the test. AC19a genuinely covers `Data` — Story 1.6's comment-only property is now asserted. |
| AC4's routing half | Red-proofed: making `RenderTo` stop calling `Render` reddens with `AC4: [RenderTo] does not route through the single document-byte producer "renderDocument"`. |
| AC25a (M-12's sharpening) | **Taken.** `TestAbsencesProductionScan` now asserts `ChecksEvaluated == len(absenceChecks)`, not `!= 0`. The gap a future `continue` would have opened is closed. Credit where due — this was offered, not mandated, and the developer took it. |
| Section F scope | Grepped every non-test `.go` file under `folio-go/`: **no PDF date emitter and no date type anywhere**; the only `/CreationDate` hits are three pre-existing comments in `internal/pdf`. 1.7 stayed inside D-1.7.7's fence. |
| AC30 / D-1.3.1 | `bannedImportPaths` holds exactly `time`, `os`, `math/rand`, `net`; `io` matches nothing. No new non-test file under `folio-go/` imports a banned path. |
| `sprint-status.yaml` | Status value only (`backlog` → `review`); no other line touched. |

---

### Finding 1: AC4's "exactly one document-byte producer" is a one-selector whitelist — the second producer the story itself names walks straight through it

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/render_arch_test.go:102-103` (`if fn.Sel.Name == "SerializeTextDocument" { info.callsProducerPkg = true }`)
- **Observation**: `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt` defines "producer" as *a function whose body calls `pdf.SerializeTextDocument`*. **Measured:** I added a second byte-producing function inside `RenderTo` that calls `pdf.Serialize()` instead — `go build ./...` **exit 0**, and the guard **PASSED** (`--- PASS`, exit 0). The RP-2 mutation the developer ran used `pdf.SerializeTextDocument`, i.e. the exact selector the whitelist matches; it reddens correctly (I reproduced it: `AC4: expected EXACTLY ONE document-byte producer, found 2: [mutantProducerA renderDocument]`). So RP-2 proves the guard fires on the one shape it was written for, not that the property holds.
- **Impact**: AC4's property is *"exactly one function produces the document bytes"* (binding, D-1.7.2). The guard asserts a strictly narrower thing. This is not a hypothetical escape route: **M-2 of this very story names `Serialize` as internal/pdf's other exported byte-producing entry point** and warns the assertion must be scoped so it does not find three producers — the scoping was done by narrowing to one selector rather than by scoping the *package*, which discarded the other producer along with the false positives. Byte-identity is claimed to hold "as a theorem" of this structural shape (AC4, and the D-1.7.2 disposition); a theorem resting on a premise the guard cannot check is not a theorem. Because AC4 explicitly replaces the behavioural comparison, this guard is the **only** thing standing between `Render` and `RenderTo` and D-1.4.14's drift hazard.
- **Suggested Resolution**: Match the producer by *any* call into `internal/pdf` that returns document bytes — at minimum both `SerializeTextDocument` and `Serialize` — or invert it to "any same-package function whose body calls into the resolved `internal/pdf` alias at all". Then re-run RP-2 in **both** spellings; a red-proof that only exercises the matched selector cannot distinguish a whitelist from a property.
- **Related AC**: AC4, AC4a, AC26 Q3 (D-1.7.2, D-1.4.14)

### Finding 2: AC20's decode guard implements half its own text — `json.NewDecoder` is not counted, and the uncounted half is the dangerous one

- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/internal/bind/decodeguard_test.go:57` (`if sel.Sel.Name == "UseNumber"`), test at `decodeguard_test.go:87`
- **Observation**: AC20 states, binding: *"A guard asserts exactly one `json.NewDecoder`/`UseNumber` site exists under `internal/bind`."* The shipped guard counts **`UseNumber` only**. **Measured, two mutations, both compiling:**
  - a second `json.NewDecoder(bytes.NewReader(d))` **with** `.UseNumber()` appended to `internal/bind/value.go` → guard **reddens**, naming both sites by location: `found 2: [value.go:133 value.go:192]`. (RP-7 reproduces; the location-reporting sharpening the Delivery Log describes is real and correct.)
  - a second `json.NewDecoder(bytes.NewReader(d))` that **omits** `.UseNumber()` → `go build ./...` exit 0 and the guard **PASSES**, exit 0.
- **Impact**: The guard catches the *harmless* duplicate (a second decoder that preserves precision) and misses the *harmful* one (a second decoder that silently narrows every number literal through `float64`). D-1.7.5's own words for why this ruling exists: *"params carry dates and amounts, where a divergence would be least visible and most expensive."* A `json.NewDecoder` without `UseNumber` — or a `json.Unmarshal` — is exactly that divergence, and it is the shape a future developer is most likely to write by accident. Secondary: `countUseNumberSites` uses `os.ReadDir` and `continue`s on `e.IsDir()`, so "under `internal/bind`" is really "directly under `internal/bind`"; a future subpackage is invisible to it.
- **Suggested Resolution**: Count `json.NewDecoder` **and** `json.Unmarshal` sites alongside `UseNumber`, and assert the pairing (exactly one `NewDecoder`, exactly one `UseNumber`, in the same function). Red-proof with the *omitting* mutation, not only the duplicating one. Recurse into subdirectories, or state in the guard's own comment that it is scoped to the top level and why.
- **Related AC**: AC20, AC26 Q1 (D-1.7.5, D-1.6.1)

### Finding 3: D-1.7.4's binding "the same code path as absent data" shipped as a SECOND duplicated path, and three of that path's four arms are unasserted

- **Severity**: Major
- **Category**: AC Conformance
- **Location**: `folio-go/internal/bind/text.go:155-191` (`lookupBound`) versus the inline data branch at `folio-go/internal/bind/text.go:106-137`
- **Observation**: D-1.7.4, binding, verbatim: *"`{{params.x}}` with no params supplied → Error naming the path and element id. This is AD-14's absent-path case applied to the params root, **and the same code path as absent data**."* AC16 carries the same phrase. The implementation instead **duplicates** AD-14's three-case switch into a new `lookupBound` helper used only by the params root; the data branch was left untouched. The two copies are currently output-equivalent — `lookupBound(data, path, path, elementID, "data", "the report data")` would produce byte-identical messages — so the shared-path refactor was available and free, and was not taken.

  The consequence is measurable. `TestAD14Triple` pins all three AD-14 cases, **but only for the data root**. Three mutations to `lookupBound`, each compiling, each run against the **full suite** `go test ./... -count=1` (7 packages):

  | Mutation to `lookupBound` | AD-14 case broken | Full-suite result |
  |---|---|---|
  | `Null` arm returns a hard error | case 2 (null renders empty, not an error) | **exit 0 — all 7 packages ok** |
  | wrong-kind arm returns `"PROBE-B-COERCED"` instead of an error | case 3 (never coerced) | **exit 0 — all 7 packages ok** |
  | `AsDecimal` bound check deleted | D-1.6.6's bound-value check on the params root | **exit 0 — all 7 packages ok** |

  Only the `Absent` arm is pinned, by `TestBindTextParamsAbsentIsLocatedError`.
- **Impact**: Two things at once. (a) A binding clause of D-1.7.4 was departed from without the departure being surfaced — the disposition row reads "Applied as stated". (b) The departure has a real cost: AD-14's cases 2 and 3 and D-1.6.6's bound check now hold on the params root **by inspection only**. The wrong-kind mutation is the sharp one — it silently renders a JSON number or object into a bank statement as text, which is precisely what AD-14 case 3 exists to forbid, and nothing in the repository notices. This also makes the D-1.6.6 disposition (*"existing bound-value checks apply without new code"*) true of the code and false of the coverage.
- **Suggested Resolution**: Route the data branch through `lookupBound` as well (`rootName: "data"`, `rootDesc: "the report data"`), which restores D-1.7.4's stated single code path and lets `TestAD14Triple`'s existing coverage carry both roots. If the duplication is kept deliberately, that is a `DECISION NEEDED` under D-1.2.6, and the params root needs its own AD-14 triple plus a params bound-value case — red-proofed with the three mutations above.
- **Related AC**: AC16, AC20's spirit; D-1.7.4 (binding), D-1.6.6, D-1.4.14

### Finding 4: `RenderTo` panics on a nil writer, and its doc comment claims it cannot

- **Severity**: Major
- **Category**: Correctness
- **Location**: `folio-go/render_entry.go:199-215` (`RenderTo`), doc comment at `render_entry.go:172`
- **Observation**: **Measured:** `RenderTo(nil, tpl, Data("{}"), nil, FontSet{})` → `runtime error: invalid memory address or nil pointer dereference`. The sibling case, `Render(nil, ...)`, returns `errNilTemplate` — a located error, and `TestRenderToPropagatesRenderError` pins it. `w` is a **new public argument introduced by this story**, and it is the only one of `RenderTo`'s five with no located-error treatment. The doc comment at `render_entry.go:172-173` states that `RenderTo` *"turns **every** writer failure into a located, non-nil error (AC5-AC8)"* — that sentence is false for the nil case.
- **Impact**: D-1.5.9 is enumerated as governing this story: *"A public entry point documented as ignoring its argument is worse than a located error."* The same reasoning applied to `Render(nil template)` at Story 1.5 applies to `RenderTo(nil writer)` here, and more sharply, because C5's "first PDF in minutes" integrator is wiring this into an HTTP handler. A comment asserting a safety property the code does not have is exactly what D-1.7.5 refuses one section earlier in this same file (M-8's lesson). The story's D-1.5.9 disposition reads *"Applied as stated (inherited, unmodified)"*, which cannot be true of an argument that did not exist at baseline.
- **Suggested Resolution**: Either add `if w == nil { return errNilWriter }` with a located message in the shape of `errNilTemplate`, and a test alongside `TestRenderToPropagatesRenderError`; or, if a panic is the deliberate choice, correct the doc comment so it does not claim every writer failure is returned, and record the decision. Do not leave the comment as-is either way.
- **Related AC**: AC5, AC1; D-1.5.9

### Finding 5: Task 5's README obligation was deliberately not done, the subtask is checked `[x]`, and the row is still dispositioned "Applied as stated"

- **Severity**: Major
- **Category**: AC Conformance
- **Location**: Story `§ Tasks / Subtasks → Task 5`, third subtask; `§ Ruling disposition table`, row **D-1.7.3**; `folio-go/README.md` (whole file)
- **Observation**: Task 5's third subtask, marked `[x]`, reads: *"Record the residual cross-file gap (AC11) in a code comment on the guard **and** in the README's determinism section — as a **capability fence, not a proof**."* **Measured:** `grep -in "capabilit|fence|temporary file|touching disk|no temporary" folio-go/README.md` → **zero matches**. The gap is recorded in `render_entry.go` and `forbiddenimports_test.go` (both verified present and accurate) but nowhere in the README. The D-1.7.3 disposition cell states the reason in its own prose — *"the README's 'Reproducible bytes' note is not the right place for it — recorded instead as a capability-fence note next to `RenderTo`'s doc comment"* — while the disposition itself still reads **"Applied as stated."**
- **Impact**: This is the precise failure mode D-000.10's mirrored table exists to catch. A reviewer scanning the disposition column sees 29/29 clean; the departure is legible only to someone who reads the full cell. Whether or not the judgement is right (it may well be — the README's determinism section is about toolchains, not import fences), the *disposition* is wrong, and a checked task box asserts work that was not done. Separately, the README is now the artifact C5 measures, and it tells an integrator that folio-go *"never reads the filesystem outside the calls you make"* (`README.md:6`) without the fence's stated limit anywhere near it.
- **Suggested Resolution**: Change the D-1.7.3 disposition to a **Departed** row naming what was substituted and why, and either un-check the subtask or amend it to match what shipped. If the README genuinely is the wrong home, say so in the story's AC11 text rather than only in a disposition cell.
- **Related AC**: AC11, AC31; D-1.7.3, D-000.10

### Finding 6: a malformed `Params` document is reported as invalid *report data*, and no test asserts the params decode error at all

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-go/render_entry.go:94-103` (`decodeParams`)
- **Observation**: **Measured** through the public API:
  - `Render(tpl, Data("{}"), Params("{oops"), FontSet{})` → `folio: Render: params: bind: invalid JSON **report data**: invalid character 'o' looking for beginning of object key string`
  - `Render(tpl, Data("{}"), Params("{\"a\":1} junk"), FontSet{})` → `folio: Render: params: bind: trailing data after the JSON **report data**'s single top-level value: …`

  And `grep` over every `_test.go` in the module finds **no test** asserting any `decodeParams` error — the wrap at `render_entry.go:100` is entirely unexercised.
- **Impact**: M-6 and AC16's whole rationale is that a message naming the wrong root is *"actively misleading, since the value was never sought in report data at all."* The story fixed that at the binding layer and left the identical defect one layer up, where an integrator meets it first. The `folio: Render: params:` prefix keeps the message from being outright wrong, but it makes it self-contradictory. The absence of any test also means AC20's "params inherit `DecodeData`'s trailing-garbage rejection for free" is claimed but never asserted.
- **Suggested Resolution**: Give `bind.DecodeData` a root label (or add a params-specific wrap that replaces rather than nests the phrase), and add two tests: malformed params JSON, and trailing garbage after params — both asserting the message names *params*.
- **Related AC**: AC16, AC20 (D-1.7.4, D-1.7.5)

### Finding 7: AC23's "near the golden-hash discussion" is satisfied vacuously — the README has no golden-hash discussion

- **Severity**: Minor
- **Category**: Convention
- **Location**: `folio-go/README.md:131-151`
- **Observation**: The substance of D-1.7.6 **is** delivered, and I want that on the record: `README.md:137-141` states that the `toolchain` directive *"binds the **main** module of whatever `go build` you run — not every module in the dependency graph"*, and that *"your own `go.mod`'s `toolchain` line (or your `GOTOOLCHAIN` setting) decides which Go compiles folio-go"*. That is exactly the binds-only-the-main-module claim, and it tells the integrator what actually pins their bytes. What is missing is the anchor: D-1.7.6 requires it *"near the golden-hash discussion, not in a footnote"*, and this README contains no golden-hash discussion at all — golden hashes appear only in passing at `README.md:142` ("An integrator who records folio-go's golden hashes…"). The placement clause is met only because the thing it names does not exist.
- **Impact**: The reader D-1.7.6 targets — *"someone recording Folio's hashes in their own test suite"* — is told what pins the hashes but never told where the hashes are or how the project publishes them. C5's integrator cannot act on the advice.
- **Suggested Resolution**: Add two or three sentences to the same section on where folio-go's golden fixtures/hashes live (`folio-go/testdata/`, `hashmatrix/`) and what the project itself pins them against, so the caveat has the subject it was written for. Not a rewrite — the caveat's own wording is fine.
- **Related AC**: AC23 (D-1.7.6)

### Finding 8: AC5's "errors mid-write" is only tested as "errors on the first and only write"

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/renderto_test.go:42-44` (`errWriter` returns `(0, w.err)` unconditionally)
- **Observation**: `RenderTo` issues exactly one `Write`, so a writer that fails on a *later* call is unreachable — that is correct behaviour and I am **not** logging it as a gap. What is reachable and untested is the `(n > 0, err != nil)` shape: a writer that accepts part of the buffer and then reports an error. `render_entry.go:205-206` handles it (`write failed after %d of %d bytes`), and the `n` it reports on that path has no test. `shortWriter` covers `(n < len(p), err == nil)`; nothing covers the partial-then-error combination.
- **Impact**: AC5 is worded as "a writer that errors **mid-write**"; the shipped double never writes anything before erroring, so the branch that actually reports how far the write got is unasserted. Low practical risk — the code is correct — but the assertion does not cover the wording.
- **Suggested Resolution**: Add a `partialThenErrWriter` returning `(len(p)/2, err)` and assert both `errors.Is` and that the message names the partial count.
- **Related AC**: AC5 (D-1.7.2)

### Finding 9: D-1.7.1's own text contradicts its verdict ("six-argument form" vs the five-argument `RenderTo` it states verbatim), and the contradiction was never surfaced

- **Severity**: Minor
- **Category**: AC Conformance
- **Location**: working-tree `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`, **D-1.7.1**, paragraph *"The `FontSet` shorthand — D-1.1.c still governs"*; story `§ AC1a`; disposition row **D-1.2.6**
- **Observation**: D-1.7.1's Verdict states, verbatim, `RenderTo(w io.Writer, t *Template, d Data, p Params, f FontSet) error` — **five** parameters. Four paragraphs later the same ruling says: *"The AC writes `RenderTo(w, template, data, params)` and omits `FontSet`; the spine governs, the AC is shorthand. **Carry the six-argument form.**"* The implementation follows the verbatim signature (five/four) and is **correct**; the story's AC1 titles it "five and four positional arguments". But AC1a records only the `FontSet` omission, not the argument-count contradiction, and the D-1.2.6 disposition asserts *"No new conflict … was encountered during implementation; none surfaced."*
- **Impact**: D-1.2.6 exists so that ruling defects are surfaced rather than silently arbitrated. A later reader auditing the shipped signature against D-1.7.1 will find a binding sentence that says six and shipped code that says five, with nothing in either artifact explaining the discrepancy.
- **Suggested Resolution**: One line in AC1a (or the decision log) recording that D-1.7.1's "six-argument form" sentence is an arithmetic slip against its own verbatim verdict, and that the verdict governs. No code change.
- **Related AC**: AC1, AC1a, AC31; D-1.7.1, D-1.2.6

### Finding 10: `absence-cmd-dir` has no permanent fixture — only the one-shot RP-8 ever proved it fires

- **Severity**: Nit
- **Category**: Tests
- **Location**: `folio-go/testdata/lint/absences/violating/` (contains only `folio-designer/` and `folio-go/fonts/`); `lint/internal/rules/absences_test.go:107`
- **Observation**: `TestAbsencesFixtureScan`'s `violating/` root exercises two of the five rows. The three added since (`absence-expr-package`, `absence-diag-package`, `absence-cmd-dir`) are proved only by transient mutations whose evidence does not survive the story. I re-ran RP-8 myself and it fires correctly, so the row is not decorative today — but the permanent suite cannot re-prove it, and `TestAbsencesChecksIncludeAllFiveEntries` only pins that the *row exists*.
- **Impact**: Low, and **inherited** — the two Story 1.4 rows have the same shape, and the firing mechanism (`os.Stat` on a joined `relPath`) is shared and fixture-proved for two rows. Logged so it is a recorded state rather than an unnoticed one.
- **Suggested Resolution**: Optional — add `violating/folio-go/cmd/placeholder.go` and extend `want` to three findings. If declined, note in `absences.go` that fixture coverage is deliberately representative rather than per-row.
- **Related AC**: AC25b

### Finding 11: the README's first-PDF example contradicts its own comment

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/README.md:45-49`
- **Observation**: The comment reads *"here, an **empty data document** and a params document supplying one runtime value"*; the next two lines are `data := folio.Data(`{"customer": {"name": "Ada Lovelace"}}`)` — not empty — and `params := folio.Params(`{"reportDate": "2026-08-23"}`)`.
- **Impact**: Cosmetic, but this is the repository's first integrator-facing artifact and the direct subject of counter-metric C5; the very next section teaches Data-versus-Params, and the example's own annotation misdescribes which is which.
- **Suggested Resolution**: Change "an empty data document" to "a small report-data document".
- **Related AC**: AC21 (C5)

---

## Reviewer's independent ruling-disposition mirror (D-000.10 / AC31)

Built from the code before reading the developer's column, then compared. **The developer's table is
29 of 29 "Applied as stated". Mine disagrees on 11 rows** — 6 substantively (marked **DEPARTED**),
5 partially. The enumeration itself is correct and complete: 27 ruling ids + DW-9 + DW-10 = 29.

| # | Ruling | Reviewer's independent disposition | Agrees? |
|---|---|---|---|
| 1 | **D-1.7.1** | **Applied.** Signatures verbatim (`render_entry.go:136`, `:199`); no options struct; README example assembles a `FontSet` from caller bytes and names Story 2.2 (`README.md:33-38`). | ✅ (see F-9 for the ruling's own defect) |
| 2 | **D-1.7.2** | **DEPARTED (partial).** Behavioural comparison correctly absent ✅; anti-streaming comment present, names AD-7, D-1.7.2, `/ID` **and** xref offsets ✅; RP-2/RP-4/RP-5 all reproduce ✅. But the structural assertion is a one-selector whitelist — **F-1** — and AC5's mid-write shape is partly untested — **F-8**. | ❌ |
| 3 | **D-1.7.3** | **DEPARTED.** Guard locates by AST ✅; RP-6 reproduces with the exact recorded finding ✅; M-5's gap reproduces green and is correctly framed as a capability fence ✅. But Task 5's README half was not done while the row reads "Applied as stated" — **F-5**. | ❌ |
| 4 | **D-1.7.4** | **DEPARTED.** `reservedPlaceholders` untouched ✅; precedence at path-segment level, spaced spelling covered ✅; AC14's fixture asserts the **value**, not just equality ✅; RP-3 reddens on the value naming the decoy ✅. But "the same code path as absent data" shipped as a second duplicated path with three unasserted arms — **F-3**. | ❌ |
| 5 | **D-1.7.5** | **Applied for the type; DEPARTED for the decode guard.** `type Params []byte` defined ✅; the swap proof genuinely fails to compile and I red-proofed **both** the `Params` alias and the `Data` alias (AC19a real) ✅. But AC20's guard implements only the `UseNumber` half — **F-2**. | ⚠️ partial |
| 6 | **D-1.7.6** | **Applied in substance; placement vacuous.** The binds-only-the-main-module claim is genuinely stated ✅. The "near the golden-hash discussion" clause is met only because no such discussion exists — **F-7**. | ⚠️ partial |
| 7 | **D-1.7.7** | **Applied.** No date type, no date emitter, no `/CreationDate` outside three pre-existing comments (grepped every non-test `.go`) ✅; test renamed and re-scoped with a non-empty-no-date-key case ✅; `absence-cmd-dir` added, directory-keyed, RP-8 reproduces ✅; **M-12's offered sharpening actually taken** ✅. Fixture gap is F-10 (Nit, inherited). | ✅ |
| 8 | **D-000.4** | **Applied.** Matrix not run and said so; `go build`/`go vet -tags=matrix ./...` both exit 0 (re-measured). | ✅ |
| 9 | **D-000.9** *(+ext)* | **Partial.** Every new guard has an explicit zero-result hard failure ✅. But three guards run, pass, and cannot see the mutation that matters (F-1, F-2, F-3) — Q1 is answered, Q3 is not. | ⚠️ partial |
| 10 | **D-000.10** | **DEPARTED.** Enumeration complete ✅, but 29/29 "Applied as stated" is contradicted by the D-1.7.3 cell's own prose (F-5) and by F-1/F-2/F-3. This mirror disagrees on 11 rows. | ❌ |
| 11 | **D-000.11** | **Applied.** Every gate in the Delivery Log carries `-count=1`; I re-ran all of them with `-count=1` and got the same results. | ✅ |
| 12 | **D-000.12** | **Applied.** No shipped test verifies bytes/hashes through a pipe; all my own byte, count and diagnostic checks went to files via `rtk proxy` or the bare binary. | ✅ |
| 13 | **D-000.13** | **Applied.** Every red-proof I reproduced asserts on message text and every control compiled. Soft spot: under RP-1 the swap test reddens on *"must NOT type-check, but it built successfully"*, not on the assignability arm — the assignability text is asserted in the other branch. Not logged separately. | ✅ |
| 14 | **D-000.14** | **Applied.** AST-counted independently: M-1's 19 sites match per file and per line-neighbourhood; 28 total post-story. `go vet ./...` is indeed strictly stronger for the migration claim. | ✅ |
| 15 | **D-1.1.a** *(addendum)* | **Applied.** `README.md:147-151` presents the low `go` line as deliberate, not a defect. | ✅ |
| 16 | **D-1.1.c** *(+addendum)* | **Applied.** Five/four shape shipped; AC1a records the AC's `FontSet` shorthand. | ✅ |
| 17 | **D-1.2.6** | **DEPARTED (minor).** An intra-ruling contradiction inside D-1.7.1 was resolved silently — **F-9**. | ❌ |
| 18 | **D-1.3.1** | **Applied.** No new non-test file under `folio-go/` imports a banned path; `io` confirmed absent from `bannedImportPaths`. Note for the record: `ScanForbiddenImports` skips any `testdata` directory, so `testdata/swapproof/*/main.go` sit outside the fence — no live violation (they import only `folio`). | ✅ |
| 19 | **D-1.3.4** | **Applied.** The new row is unconditional; `ScanAbsences` still has no skip-if-missing branch. | ✅ |
| 20 | **D-1.4.11** | **Applied, and improved.** Witness re-verified live; `TestAbsencesProductionScan` strengthened to `ChecksEvaluated == len(absenceChecks)`. This is the strongest single piece of work in the story. | ✅ |
| 21 | **D-1.4.14** | **DEPARTED.** One core in fact ✅, but the guard proving it is F-1's whitelist, and F-3's duplicated lookup path is a **new instance** of the very drift hazard this ruling refuses. | ❌ |
| 22 | **D-1.5.5** | **Applied.** AST-verified: `p` inserted between `d` and `f` at all 19 sites, none reordered. | ✅ |
| 23 | **D-1.5.9** | **DEPARTED.** The new public `w` argument gets a panic rather than a located error, and the doc comment claims otherwise — **F-4**. "Inherited, unmodified" cannot describe an argument that did not exist at baseline. | ❌ |
| 24 | **D-1.6.1** | **Applied in fact, unguarded in part.** Params route through `DecodeData` → the same `Value`/`Decimal` machinery ✅; the guard meant to keep it that way is F-2. | ⚠️ partial |
| 25 | **D-1.6.3** | **Applied.** RP-6 reproduced end to end, control and finding. | ✅ |
| 26 | **D-1.6.5** | **Applied.** `params` is a separate first-segment branch, checked *after* the reservation; `{{page}}`/`{{pages}}` byte-identical and asserted in the same gate (`TestBindTextPageAndPagesUnaffectedByParamsRoot`). AD-4 not weakened. | ✅ |
| 27 | **D-1.6.6** | **DEPARTED (unasserted).** The `AsDecimal` bound check exists on the params root, but deleting it leaves the entire suite green — the disposition's "apply without new code" is true of the code and false of the coverage (**F-3**). | ❌ |
| 28 | **DW-9** | **Applied.** README names Story 2.2 as removing the FontSet step; I agree AC22a was not triggered — the only ceremony in the example is the font assembly. | ✅ |
| 29 | **DW-10** | **Applied.** Tripwire live and red-proofed; `folio-go/cmd` verified absent at review end. | ✅ |

**Agreement: 18 of 29 rows. Substantive departures: 6 (D-1.7.2, D-1.7.3, D-1.7.4, D-000.10,
D-1.2.6, D-1.5.9, D-1.4.14, D-1.6.6 — counting the last two as consequences of F-3). Partial: 5.**

### Things the story got right that a lazier review would have logged as defects

Recorded so a later reader does not re-open them:

- **The absent `Render`-vs-`RenderTo` byte comparison is correct**, not an omission. D-1.7.2 deletes it deliberately, and I confirmed the structural replacement exists and reddens on both halves (producer count and routing).
- **M-5's cross-file gap is real, green, and accepted.** I reproduced it. It is correctly described as a capability fence, not a proof, in two code comments. **Not a defect.**
- **`AC8`'s payload-concatenation check is not the vacuous comparison** — it reddened under RP-5 naming both the extra byte and the extra call.
- **The matrix's absence is not a defect** (D-000.4); the `-tags=matrix` build and vet both pass.
- **`hashmatrix`'s build failure is pre-existing** and correctly disclosed.
- **AC22a's escape hatch was correctly left untriggered.**

### Tree integrity

Every mutation in this review was applied to a hand-restored copy. After the last one, `git diff
3463119` was diffed byte-for-byte against a snapshot taken before the review (**identical**), all
seven untracked files were `diff`'d individually (**identical**), `folio-go` and `lint` suites were
re-run `-count=1` (**both exit 0**), and `gofmt -l` is clean. `git status --porcelain` shows the same
21 entries it showed at the start. The throwaway worktree used for the baseline count was removed.

---

## Finisher's Finding Resolutions

Every finding the reviewer logged was reproduced independently before being triaged (see the
[Delivery Log addendum](#finishers-delivery-log-addendum) above for the reproduction of Findings
1-3 specifically). **All eleven findings are FIX** — every one was a real, in-scope gap with a
minimal available fix; none required expanding scope beyond what this story's own rulings and ACs
already call for, so none were DISMISSED or DEFERRED.

| # | Severity | Decision | Rationale | Files changed |
|---|---|---|---|---|
| F1 | Major | **FIX** | Reproduced live: a second producer calling `pdf.Serialize()` (M-2's own named "other" entry point) built and passed the guard undetected. Broadened producer detection to any call resolved through the `internal/pdf` import alias — today that still matches exactly one function (`renderDocument`), so this is a narrowing of what future code can get away with, not a broadening of what passes now. | `render_arch_test.go` |
| F2 | Major | **FIX** | Reproduced live: a second `json.NewDecoder` **omitting** `.UseNumber()` built and passed. The guard now counts `json.NewDecoder`, `UseNumber`, and `json.Unmarshal` sites and asserts all three (exactly one of the first two, zero of the third), so a second decoder of any shape reddens, not only one that copies the first one's discipline. | `internal/bind/decodeguard_test.go` |
| F3 | Major | **FIX** (the finding that matters most) | D-1.7.4's binding text is *"the same code path as absent data"* — verified `lookupBound` had exactly one call site (params) and `TestAD14Triple` did not reach it. Routed the data branch through the same `lookupBound` helper (byte-identical messages, confirmed) rather than adding a parallel params-only test, exactly as directed. Re-ran all three of the reviewer's mutations against the full suite — all three now redden where they previously exited 0. | `internal/bind/text.go` |
| F4 | Major | **FIX** | `w` is a new public argument this story introduces; `RenderTo(nil, ...)` panicked while the doc comment claimed every writer failure becomes a located error. D-1.5.9 requires a located error, not a panic, for a public entry point's argument. Added `errNilWriter`, a nil check, and a test. | `render_entry.go`, `renderto_test.go` |
| F5 | Major | **FIX** | Task 5's third subtask (record the capability-fence limit in the README) was genuinely not done while both the subtask checkbox and the D-1.7.3 disposition row read as complete/"Applied as stated". Added the capability-fence paragraph to the README's `RenderTo` section; the disposition table below is corrected to match. | `README.md`, disposition table |
| F6 | Minor | **FIX** | A malformed or trailing-garbage Params document was reported as invalid *report data* — self-contradictory (M-6/AC16's own rationale, one layer up), and entirely untested. Split `DecodeData`/`DecodeParams` from one shared `decodeJSON(d, rootLabel)` implementation (AC20's one-decode-path guard is unaffected — still one `NewDecoder`/`UseNumber` call site) and added two tests. | `internal/bind/value.go`, `render_entry.go`, `render_bind_test.go` |
| F7 | Minor | **FIX** | AC23/D-1.7.6's "near the golden-hash discussion" placement clause was met only because the README had no golden-hash discussion to be near. Added a short paragraph naming where folio-go's golden fixtures/hashes actually live (`folio-go/testdata/`, `hashmatrix/`). | `README.md` |
| F8 | Minor | **FIX** | AC5 is worded as a writer that errors **mid**-write; the shipped `errWriter` double never accepts any bytes before failing, leaving the `(n > 0, err != nil)` branch (and the partial count it reports) unasserted. Added `partialThenErrWriter` and a test naming the partial count. | `renderto_test.go` |
| F9 | Minor | **FIX** (doc-only) | D-1.7.1's own ruling text contradicts its verbatim Verdict ("carry the six-argument form" vs. the five-parameter signature stated verbatim earlier in the same ruling) — an arithmetic slip against its own verdict, not a second instruction. Surfaced in AC1a per D-1.2.6 (surface, don't arbitrate); the decision log is append-only and was not edited; no code change. | story file (`§ AC1a`) |
| F10 | Nit | **FIX** | `absence-cmd-dir`'s red-proof only ever existed as the one-shot RP-8 mutation; the permanent fixture suite could not re-prove it. Added a fixture directory (git cannot track an empty one) and extended `TestAbsencesFixtureScan`'s `want`. | `testdata/lint/absences/violating/folio-go/cmd/placeholder.go`, `lint/internal/rules/absences_test.go` |
| F11 | Nit | **FIX** | The first-PDF example's comment said "an empty data document" while the two lines immediately below show a populated one. Corrected the comment to match the code. | `README.md` |

---

## Finisher's corrected ruling disposition table

**The developer's table claimed 29 of 29 "Applied as stated". It was not.** The reviewer's
independent mirror found real departures and said so — but even the reviewer's own summary
undercounted its own table (see the [Delivery Log addendum](#finishers-delivery-log-addendum)
above: the reviewer's table itself shows **17** agreements / **8** substantive (❌) / **4** partial
(⚠️) — 12 disagreements — against a prose summary claiming 18/29 and 11 disagreements). **Both
"29/29" and "18/29" undercount what the developer's table actually got wrong.**

Below is the post-fix disposition of all 29 rows, now that Findings 1-11 are FIX-resolved. Every
row that the reviewer marked ❌ or ⚠️ is genuinely **Applied** now — not because the disagreement is
waved away, but because the underlying code gap each one pointed at (F-1 through F-4, F-9) has been
closed, and this table says so explicitly rather than silently reverting to "Applied as stated."
Rows the reviewer marked ✅ are carried forward unchanged.

| # | Ruling | Original developer disposition | Reviewer's mirror | Post-fix disposition (finisher) |
|---|---|---|---|---|
| 1 | **D-1.7.1** | Applied as stated | ✅ (ruling's own text defect noted separately, F-9) | **Applied as stated.** Unchanged — F-9 is about the ruling's own prose, not the implementation; surfaced in AC1a, no code change. |
| 2 | **D-1.7.2** | Applied as stated | ❌ DEPARTED (partial) — F-1, F-8 | **Applied as stated.** F-1 (whitelist) and F-8 (mid-write undertested) are both fixed; the structural guard now proves the property it claims, and AC5's mid-write shape is asserted. |
| 3 | **D-1.7.3** | Applied as stated | ❌ DEPARTED — F-5 | **Applied as stated.** F-5 fixed: the README now carries the capability-fence note Task 5's subtask always claimed to have added. |
| 4 | **D-1.7.4** | Applied as stated | ❌ DEPARTED — F-3 | **Applied as stated.** F-3 fixed: the data branch now genuinely routes through the same `lookupBound` code path as the params branch — the ruling's "same code path as absent data" is now literally true, not narrated. |
| 5 | **D-1.7.5** | Applied as stated | ⚠️ partial — F-2 | **Applied as stated.** F-2 fixed: the decode guard now counts `NewDecoder`/`Unmarshal` alongside `UseNumber`, closing the "omits UseNumber" blind spot. |
| 6 | **D-1.7.6** | Applied as stated | ⚠️ partial — F-7 | **Applied as stated.** F-7 fixed: the README now names where the golden hashes live, giving the placement clause a real subject to be "near". |
| 7 | **D-1.7.7** | Applied as stated | ✅ (F-10 nit noted, inherited pattern) | **Applied as stated.** Unchanged in substance; F-10 additionally fixed (permanent fixture), strengthening but not changing the disposition. |
| 8 | **D-000.4** | Applied as stated | ✅ | **Applied as stated.** Unchanged; matrix explicitly re-confirmed unrun in the Delivery Log addendum. |
| 9 | **D-000.9** *(+ext)* | Applied as stated | ⚠️ partial — "Q1 answered, Q3 not" (F-1/F-2/F-3) | **Applied as stated.** With F-1/F-2/F-3 fixed, the three guards the reviewer named now genuinely cannot pass while the property they assert is false — Q3 is now answered too. |
| 10 | **D-000.10** | Applied as stated. All 29 rows dispositioned below. | ❌ DEPARTED — table doesn't match reality; reviewer's own count is off too | **Corrected, and the discrepancy stated in both directions** (developer's false 29/29, reviewer's own miscounted 18/29 summary) — see the Delivery Log addendum and this table. |
| 11 | **D-000.11** | Applied as stated | ✅ | **Applied as stated.** Unchanged; every finisher gate also carries `-count=1`. |
| 12 | **D-000.12** | Applied as stated | ✅ | **Applied as stated.** Unchanged; finisher used the bare `go`/`gofmt` binaries directly (see Delivery Log addendum) — no bytes/hashes/panics verified through a shell pipe. |
| 13 | **D-000.13** | Applied as stated | ✅ | **Applied as stated.** Unchanged; every red-proof this finisher touched was re-run and asserts on message text (see Delivery Log addendum). |
| 14 | **D-000.14** | Applied as stated | ✅ | **Applied as stated.** Unchanged; the finisher's own test counts were re-derived via `-list`/`-v` `RUN`/`PASS` counting, not `grep -c` on raw output. |
| 15 | **D-1.1.a** *(addendum)* | Applied as stated | ✅ | **Applied as stated.** Unchanged. |
| 16 | **D-1.1.c** *(+addendum)* | Applied as stated | ✅ | **Applied as stated.** Unchanged. |
| 17 | **D-1.2.6** | Applied as stated. No new conflict … was encountered; none surfaced. | ❌ DEPARTED (minor) — F-9 | **Corrected.** The original disposition was about NEW conflicts arising during implementation, which is true — but it did not surface the PRE-EXISTING arithmetic slip inside D-1.7.1 itself. Now surfaced in AC1a, per D-1.2.6 (surface, do not arbitrate); the decision log remains append-only and unedited. |
| 18 | **D-1.3.1** | Applied as stated | ✅ | **Applied as stated.** Unchanged. |
| 19 | **D-1.3.4** | Applied as stated | ✅ | **Applied as stated.** Unchanged. |
| 20 | **D-1.4.11** | Applied as stated | ✅ (reviewer: "strongest single piece of work in the story") | **Applied as stated.** Unchanged. |
| 21 | **D-1.4.14** | Applied as stated | ❌ DEPARTED — F-1's whitelist, F-3's duplicated path | **Applied as stated.** F-1 and F-3 both fixed: the guard now actually proves the one-producer property, and the duplicated lookup path is eliminated (one `lookupBound`, not two) — the drift hazard this ruling refuses no longer has a live instance. |
| 22 | **D-1.5.5** | Applied as stated | ✅ | **Applied as stated.** Unchanged. |
| 23 | **D-1.5.9** | Applied as stated (inherited, unmodified) | ❌ DEPARTED — F-4; "inherited, unmodified" cannot describe an argument new to this story | **Corrected, then fixed.** The original "inherited, unmodified" framing was wrong on its face (`w` did not exist before this story). F-4 fixed: `w == nil` now returns `errNilWriter`, a located error, matching D-1.5.9's own reasoning rather than merely citing it. |
| 24 | **D-1.6.1** | Applied as stated | ⚠️ partial — F-2 | **Applied as stated.** F-2 fixed; see row 5. |
| 25 | **D-1.6.3** | Applied as stated | ✅ | **Applied as stated.** Unchanged. |
| 26 | **D-1.6.5** | Applied as stated | ✅ | **Applied as stated.** Unchanged. |
| 27 | **D-1.6.6** | Applied as stated (inherited, unmodified) | ❌ DEPARTED (unasserted) — F-3 consequence | **Applied — and now genuinely covered**, not merely true "of the code and false of the coverage" as the reviewer put it. F-3's fix means the `AsDecimal` bound check is exercised identically on both roots through the one shared `lookupBound`, and the mutation that deletes it now reddens the full suite (confirmed in the Delivery Log addendum). |
| 28 | **DW-9** | Applied as stated | ✅ | **Applied as stated.** Unchanged. |
| 29 | **DW-10** | Applied as stated | ✅ | **Applied as stated.** Unchanged; F-10 additionally makes the tripwire's own proof permanent. |

**Post-fix: 29 of 29 rows are Applied as stated — this time verified by re-running the mutation each
disagreement was based on, not asserted by a table cell.** Rows 10, 17, and 23 needed a correction
to their FRAMING (not just a code fix) and say so explicitly above, rather than reverting silently
to "Applied as stated" the way the original table did.

