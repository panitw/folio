# Story 1.5: Embed and subset a font, byte-stably

Status: done

**Story key:** `1-5-embed-and-subset-a-font-byte-stably`
**Epic:** 1 · **Baseline:** `94b453a`
**Covers:** FR32 · NFR1.e, NFR7 · AD-8
**Adjacent:** AD-1, AD-2, AD-3, AD-6, AD-7, AD-12, AD-21, AD-22, AD-24, AD-26, §Consistency
Conventions, §Stack, §Source tree
**Rulings that govern:** **D-1.5.1**, **D-1.5.2**, **D-1.5.3**, **D-1.5.4** *(corrected)*,
**D-1.5.5**, **D-1.5.6**, **D-1.5.7**, **D-1.5.8**, **D-1.5.9**, **D-1.5.10**, D-000.4, D-000.9,
D-000.10, D-000.11, D-000.5, D-000.6, D-1.1.a, D-1.1.b, D-1.1.c, D-1.2.6, D-1.3.1, D-1.3.4,
D-1.3.6, D-1.3.9, D-1.4.15

> **Mechanism tagging (D-1.4.15's convention, first applied at Story 1.5).** Every named mechanism
> below carries `(mechanism: binding)` or `(mechanism: illustrative)`. **An illustrative mechanism
> may be improved on; a binding one may not.** Where this story quotes a ruling, the tag travels
> with the quote.
>
> **Nothing is parked. Both conflicts this story surfaced are ruled**, as D-1.5.7 (the subsetting
> input's ordering) and D-1.5.8 (which glyph-id set the tag hashes); the golden-fixture collision
> is ruled as D-1.5.9; three further confirmations are D-1.5.10. See
> [§ Decisions resolved](#decisions-resolved).
>
> **Three `(mechanism: binding)` clauses were REVERSED by measurement taken while this file was
> being written** — D-1.5.4's sorted-input clause (withdrawn as inert, D-1.5.7), D-1.5.4's
> "or zero" latitude and its "never near the current time" half (both withdrawn, D-1.5.10), and
> D-1.5.2's guardrail restated in the only form that can bind (D-1.5.10). **This is D-1.4.15's
> mechanism tagging doing exactly what it was introduced for, one story after being adopted:** the
> tags made it possible to see that the *property* survived while the *mechanism* did not.
>
> **Where these rulings live.** D-1.5.1 – D-1.5.10 are in the **working tree** copy of
> `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`, **not in `HEAD`**. They will
> land in a commit that is **not this story's**. A reviewer diffing against baseline `94b453a`
> will not find them there. (The story brief stated the log was committed and the tree clean; that
> was verified wrong — 187 uncommitted lines at the time of writing, more since.)

---

## In plain terms (read this first if you just want the gist)

The engine used to write a PDF with no real text in it, only a fixed rectangle. It now reads a
document, works out which font each piece of text needs, builds a small private copy of that font
containing only the glyphs actually used, and packs that copy inside the PDF — so the file looks
right on a machine that has never heard of the font. That mechanism shipped and was proven byte-for-
byte identical across four different operating systems and processor types, twice: once when it was
first built, and again after everything below was fixed, with no change to the bytes either time.

The hard part was never embedding, it was sameness — the same document must produce the same bytes
every time, everywhere, and font tools have a habit of quietly stamping the moment they ran into
their output. That was checked and closed off before review even started. What review found instead
was risk sitting on top of a sound mechanism: several tests built to guard sameness didn't actually
guard it. One test meant to fail if a dangerous shortcut crept in never called the code it was
checking, so it could never fail. A safety check had a quiet escape hatch that would let it pass
without checking anything, given the right future change. A build guard missed a font-loading
pattern the very next piece of work is expected to use, and the pipeline only tested one of this
story's two output documents, not both. All of that, plus a dozen smaller gaps, is now closed, each
one demonstrated failing before the fix and passing after.

Nothing about what glyphs end up in the file, or what the file's bytes look like, changed. What
changed is confidence that the guards protecting those bytes will actually catch it if something
breaks them later. As before: the real, shippable set of fonts is a later story's job — this one
proves the mechanism against a single small test font only.

---

## Story

As an integrating Go developer,
I want every font the document uses embedded as a subset that is identical on every run,
So that the PDF renders correctly on a machine with no fonts installed without breaking hash
equality.

---

## Acceptance Criteria

Source ACs are `_bmad-output/planning-artifacts/epics.md` §Story 1.5 (lines 542–578). They are
carried below unchanged in substance and split so each is separately red-proofable, with the
ruling-mandated additions folded in.

### The font seam (AD-8, D-1.5.6)

**AC1.** `Render` takes a `FontSet` value as an explicit parameter, in its final target position:
`func Render(t *Template, f FontSet) ([]byte, error)` *(mechanism: binding, D-1.5.5)*. Font
resolution inside a render is a pure lookup against that supplied value.

**AC2.** The `FontSet` **type** is declared in package `folio` at the module root (`fontset.go`,
spine §Source tree: *"`fontset.go` — FontSet as a public input"*). `internal/fontset` performs
resolution and subsetting against the value it is handed, and **never** embeds font data and
**never** queries the host.

**AC3.** No package under `internal/` embeds font data (AD-8's Rule, verbatim: *"no package under
`internal/` embeds font data"*). Asserted by a guard, not by inspection: no `go:embed` directive
naming a font file exists anywhere under `folio-go/internal/`. The guard reports a **coverage
witness** — how many files it parsed — and **zero candidate files is a failure** (D-000.9).

**AC4.** No host font query occurs on the render path. Asserted through the existing AD-1 import
guard surface (`lint/internal/rules/forbiddenimports.go`): no package under `internal/` reaches a
font-discovery or filesystem-enumeration route.

### Embedding and subsetting (FR32, AD-7)

**AC5.** A document using a subset of a Latin face is embedded as a `Type0` / `Identity-H`
composite font carrying a `FontFile2` stream and a `ToUnicode` CMap.

> **Two sites, opposite answers. The story must carry both, and confusing them breaks a proof.**
> **Subset input → do NOT sort.** **Tag derivation → MUST sort.** (D-1.5.7)

**AC6 — the tag derivation site: sort.** The six-letter subset tag is `A`–`Z` only, exactly six
characters, derived from a hash of the **sorted** glyph-id set *(mechanism: binding, D-1.5.4
retained at this site; AD-7: "Font subset tags are the six letters `A`–`Z` derived from a hash of
the sorted glyph-id set, which ISO 32000-1 §9.6.4 permits")*. **Sorting happens before hashing.**
Here the ordering is **ours**, it **reaches an output byte**, and it comes from ranging a map — so
**Story 1.3's `ScanMapRange` guard is load-bearing at this site and must not be worked around**
*(mechanism: binding, D-1.5.7)*.

**AC7 — which set.** The tag hashes **`Plan.GlyphSet()` — the closure set, in SOURCE-font glyph
numbering, sorted ascending** *(mechanism: binding, D-1.5.8)*. Verified: that method's doc comment
reads *"returns the set of **old** glyph IDs to retain"* (`subset/plan.go:368`). *(mechanism:
illustrative)* for the encoding — a digest over the sorted big-endian GID sequence, folded into six
`A`–`Z` letters.

**AC7a — the two rejected readings, recorded with their reasons** so neither is re-derived:

- **Not the requested set** (the ids our cmap lookup resolved). The tag exists so a consumer can
  tell two subsets of one base font apart (ISO 32000-1 §9.6.4). A function of what we *asked for*
  means two different embedded programs — same request, different closure after a dependency
  upgrade — carry **the same tag while differing in content**. AD-22 already makes a subsetting
  change a versioned breaking change, so **a tag that moves on upgrade is a correct signal**, not a
  defect.
- **Not the output numbering — and this is the most natural-sounding reading, which is why it is
  written down.** Verified in source: `createCompactMapping` sorts the retained old GIDs and then
  does `for newGID, oldGID := range gids`, assigning `newGID = index`. **The output glyph-id set is
  therefore always `{0, 1, … n-1}`**, so a tag hashed over it is **a function of glyph count
  alone** — two entirely different 8-glyph subsets collide on the same tag. Catastrophic, and
  silently so.

**AC7b — the retained fixture that kills the output-numbering reading permanently** *(mechanism:
binding, D-1.5.8)*: **two documents whose glyph sets differ but are the same size produce different
tags.** Red under the output-numbering reading, green under the ruled one, and **not vacuous**.
Retained, per Story 1.3's standing rule.

**AC8 — the subset input site: do NOT sort. The sorted-input mechanism is WITHDRAWN as inert**
*(D-1.5.7, withdrawing a D-1.5.4 `(mechanism: binding)` clause)*. `Input.glyphs` is a
`map[GlyphID]bool` and `AddGlyphs` writes straight into it, so **input order dies at the door**.
Measured: 6860 bytes both orders; 2684 bytes both orders (Dev Notes F-2).

**Two proofs replace it, and both are buildable** *(mechanism: binding, D-1.5.7)*:

1. **Repeat-invariance.** Call the subsetting path **N ≥ 16 times in one process** on the same
   glyph set; every output byte-identical to the first. The lead measured that ranging an
   8-element map 200 times in one Go process yields **8 distinct iteration orders** — **Go
   randomises per `range` statement, not once per process.** So if a future `textshape` drops its
   internal sort, this reddens immediately (coincidence ≈ 1/8!, driven to nothing by N). **No
   vendoring, no pinning, survives upgrades.**
2. **Permutation-invariance.** Distinct input orders produce identical bytes. **Only meaningful if
   we do NOT sort first** — which is the whole reason AC8's original mechanism had to go.

**AC8a — the general rule this produced, for the vacuity-guard list** *(mechanism: binding,
D-1.5.7)*: ***a defensive normalisation upstream of a check can delete that check's discriminating
power.*** Sorting our input made the permutation test vacuous, because permuted-then-sorted inputs
**are the same input**. This is a **new** entry, distinct from the existing D-000.9 instances: the
defect is not a check that cannot observe its subject, but a *safety measure* that removes the
subject before the check runs.

**AC9.** **One subset per font per document**, over the union of glyphs the document uses, with the
subsetting call made **once after all glyphs are collected** — never per page, never per element
*(mechanism: binding, D-1.5.4; spine §Consistency Conventions: "One subset per font per
**document**, over the union of glyphs the whole document uses. Never per page.")*. Asserted
**behaviourally**: a **two-page** document using one face embeds **exactly one** `FontFile2`.

### Byte stability (NFR1.e, AD-21)

**AC10.** The same document rendered in **two separate OS processes** produces byte-identical
embedded font programs, and byte-identical PDFs.

**AC10a — the subprocess seam renders BOTH documents** *(mechanism: binding, D-1.5.9)*.
`FOLIO_SUBPROCESS_RENDER=1` **keeps the fontless path** (so `fixtures/minimal-rect` does not lose
its four-target verification — the property that makes it the repo's most valuable artifact), and
a **second selector** covers the template+font document. **The matrix runs both.**

**AC10b — vacuity guard, binding** *(D-1.5.10, endorsing this story's own finding)*: the child
render on the font path must be asserted to **actually contain a `FontFile2`** before any byte
comparison runs. Without it, **AC10 *and* AC15's four-target matrix pass on two identical fontless
rectangles** (D-000.9: *"What would this have printed if it had been unable to run at all?"* — same
output ⇒ instance). **The D-000.4 matrix override for this story is only meaningful if the
subprocess path actually embeds a font.**

**AC11.** **The two-run timestamp comparison is rejected** *(mechanism: binding — the rejection,
D-1.5.4)*.

D-1.5.4's reason, verbatim: *"OpenType `head.created` and `head.modified` are **LONGDATETIME:
seconds since 1904**. **Two renders in the same second produce identical bytes even if the
subsetter stamps the clock on every call.** A same-second comparison *cannot fail* — so it would
report 'no timestamp' for a subsetter that timestamps every single time."*

**AC11a — what replaces it, tightened to equality alone** *(mechanism: binding, D-1.5.10)*. In the
embedded `FontFile2` stream, the head table's `created` (offset 20) and `modified` (offset 28)
**equal the source face's values exactly, byte-for-byte**. *(mechanism: illustrative for the
offsets.)*

**Two clauses of D-1.5.4 are withdrawn here, both of them `(mechanism: binding)`:**

- **"or zero" is withdrawn.** Measured (Dev Notes F-5), `subsetHead` copies the whole 54-byte head
  table and patches only offset 50, so **equality is what actually happens**. Latitude to accept
  zero would mask a future change rather than catch it.
- **"never a value near the current time" is withdrawn — on the face-age argument alone**
  *(corrected, Finding 10, this story's QA review; D-1.5.10 (corrected) in the decision log)*.
  The originally-stated reason — that this test "must live" under `internal/`, where D-1.3.1 bans
  `time` in `_test.go` files — is **false**: the shipped test,
  `TestEmbeddedHeadTimestampsEqualSourceExactly`, ships at `folio-go/render_test.go`, package
  `folio`, the **module root**, where D-1.3.1's ban does not apply. The withdrawal's **conclusion**
  still stands on the argument that was always independent and sufficient on its own: the face-age
  trap this story measured (Roboto `modified` ≈ 3.57e9 vs ≈ 3.87e9 for 2026-08-23) means "not near
  the current time" is not merely weak, it is **permanently, meaninglessly true** — a face this old
  can never be near "now" again, so the clause could never discriminate a subsetter that
  timestamps every call from one that never does. The shipped equality-only assertion is strictly
  stronger either way.

**AC12.** The measured provenance — `textshape` imports no `time` in non-test code, and
`ot/metrics.go:27–28,59–60` only **reads** those fields — is recorded in the Delivery Log **as
evidence, not as an assertion** *(mechanism: binding)*. D-1.5.4: *"**Do not turn it into an
assertion over the module cache** — that would bind the test to `GOMODCACHE` layout, whereas the
behavioural assertion survives an upgrade and a source scan of a pinned version does not."*

**AC13.** The resulting PDF renders its text correctly when opened on a machine with none of its
fonts installed. Verified by an **independent reader** and recorded in the Delivery Log, not wired
into the module — `folio-go` intentionally has no PDF-reading dependency (AD-6, Story 1.1 AC2).

**AC14.** This story ships a **new** golden fixture, **beside** `fixtures/minimal-rect/`, for the
template-driven font document (AD-21: *"No change may land without a fixture covering it"*). Same
shape as the existing one: SHA-256 of the rendered bytes, the `folio-go` version, and the exact Go
toolchain version, with `sha256` a JSON string of exactly 64 lower-case hex characters, never a
per-target map (Story 1.2 AC16, D-1.2.2).

**AC14a — `fixtures/minimal-rect/` is RETAINED UNCHANGED and RECLASSIFIED** *(mechanism: binding,
D-1.5.9)*. Its hash, its bytes, its README and its *"do not regenerate"* instruction **all stand**.
What changes is only what it is understood to pin: **the fontless emission path in `internal/pdf`**
— AD-3 number emission, xref arithmetic, content-derived `/ID` — and so **its test calls
`pdf.Serialize()` directly**, not the public `Render`. See Dev Notes F-8 for the proof that no
`.folio` template can reproduce its bytes.

**AC14b — consequence: `Render` may require a non-nil template** *(D-1.5.9)*, returning a **located
error** on nil. That is better than a public entry point documented as ignoring its argument.

**AC15.** **D-000.4 override — the full four-target matrix runs in THIS story, not at epic close.**
Font subsetting byte-stability is precisely the property AD-21's matrix exists to falsify, and it
is the first property in the project whose determinism depends on a third-party module's internal
ordering. `darwin/arm64`, `linux/amd64`, `linux/arm64`, and `js/wasm` under Node, hashes compared
across all four, **for both documents (AC10a)**. Say so in the Delivery Log with all the measured
hashes. **This override is only meaningful if the subprocess font path actually embeds a font —
see AC10b.**

### `unitsPerEm` at the trust boundary (D-1.5.2, AD-2)

**AC16.** `internal/fontset` validates `unitsPerEm` **before** any value derived from a
caller-supplied font reaches `geom.ScaleRound`. Valid range **16–16384**; anything outside —
**including `0`** — is a **located load error naming the font and the value**. At Story 1.5 that is
a plain Go error; `internal/diag` is Story 3.6 *(mechanism: binding, D-1.5.2)*.

**AC17.** **The by-construction guardrail** *(mechanism: binding, D-1.5.2, verbatim)*: *"**the
parsed-font type must not expose a raw, unvalidated `unitsPerEm`** that a caller can hand to
`ScaleRound`. Validation happens in the constructor and the validated value is the **only** one
reachable. A field that can be read before it is checked will eventually be read before it is
checked."*

**AC17a — the binding form of that guardrail** *(mechanism: binding, D-1.5.10, adopting this
story's F-4)*: **do not leak `*ot.Font`, `*ot.Head` or `*subset.Plan` out of `internal/fontset`.**
The third-party parsed type **does** expose a raw, unvalidated `unitsPerEm` (`ot.Head.UnitsPerEm`,
an exported `uint16` with no range check) and we do not control it — so **a vendor type we cannot
constrain must not become part of a seam we rely on being constrained.** Without this, AC17 is
satisfied in letter and defeated in fact. See Dev Notes F-4.

**AC18.** **`ScaleRound` gets NO second variant** *(mechanism: binding, D-1.5.2)*. AD-2: *"Font
scaling is one exported function with one documented rounding mode."* `internal/geom` gains no new
exported scaling entry point in this story; asserted by a guard over `internal/geom`'s exported
surface. D-1.5.2: *"the fix is **layering, not a second door**."*

**AC19.** **Retained fixture, both polarities, plus the upper bound** *(mechanism: binding)*: a
font with `unitsPerEm = 0` produces a located error and **not a panic**; a font with
`unitsPerEm > 16384` produces a located error; a valid font scales correctly. All three retained
(Story 1.3's standing rule: *"a deliberately violating fixture proves it fires, and the fixture is
retained"*).

### The module graph (D-1.5.1, AD-6, AD-26)

**AC20.** **The one-module assertion in `folio-go/gomod_test.go` breaks by design and is replaced,
not deleted.** Primary assertion: `go list -m all` must equal **exactly**
`{github.com/panitw/folio/folio-go, github.com/boxesandglue/textshape}`, **asserted by name**, with
the difference printed **both ways** on failure *(mechanism: binding, D-1.5.1)*.

**AC21.** Independent backstop: **no known Go PDF writer at any depth** — `signintech/gopdf`,
`boxesandglue/baseline-pdf`, `jung-kurt/gofpdf`, `pdfcpu/pdfcpu`, `unidoc/unipdf`,
`phpdave11/gofpdi` *(mechanism: binding for the property; illustrative for the list, which must be
extendable)*. D-1.5.1's reason, verbatim: *"**A denylist alone is weaker than the count it
replaces:** a PDF writer nobody listed passes silently, and lists rot."* The allowlist and the
denylist are **two independent checks with different failure modes**; neither substitutes for the
other.

**AC22.** **The split is preserved.** The graph assertion **stays in `gomod_test.go`** — an AD-6
property, cheap, must fail in the fast per-story gate. The **AD-26 licence half stays in `lint/`**.
D-1.5.1: *"Different invariants, different homes; **do not merge them**."*

**AC23.** The `lint/` licence check runs over `folio-go`'s now-**non-empty** module graph, and the
manifest records `textshape` as the **first shipped** dependency under D-1.3.9's
shipped-vs-build-time labelling *(mechanism: binding, D-1.5.3)*. D-1.5.3: *"**A check that has only
ever run against an empty graph has never been exercised** — this is the story that proves it
works, and the proof belongs in the Delivery Log with the actual manifest output."* Paste the
regenerated `lint/MANIFEST.md` table into the Delivery Log.

**AC24.** **Pre-empt the false alarm** *(mechanism: binding, D-1.5.3)*: `textshape`'s own `go.mod`
declares `go 1.23.0` and **`toolchain go1.24.0`**. Per D-1.1.a's addendum a `toolchain` directive
binds **only the main module**, and `go 1.23.0` is below our `1.25.0` floor. **It is not a
toolchain-pin violation.** Record this in the Dev Agent Record so it is not re-litigated at review.

### The redistributed asset (AD-26, D-1.5.6)

**AC25.** Any committed font binary — fixture or shipped — travels with its **licence text and
copyright line** *(mechanism: binding, D-1.5.6)*. AD-26, verbatim: *"Redistributed non-code assets
keep their own terms and their notices."* **This is the repo's first non-code redistributed asset
and it sets the pattern for Story 2.2.** The `lint/` manifest accounts for it.

**AC26.** **Scope fence** *(mechanism: binding, D-1.5.6)*: **Story 1.5 does NOT ship the production
faces.** Noto Sans / Noto Sans Thai / Noto Sans SC are **Story 2.2**, which is also where the
unresolved **AD-12 `ja` coverage question** comes due — potentially an **owner** call, since a new
face breaks Story 5.4's fixed CJK payload numbers and the ~9 MB budget. **1.5 builds the `FontSet`
type, the resolution seam and the subsetting path, and tests against one small Latin test face.**

### Self-retiring assertions (D-1.5.5)

**AC27.** **The Story 1.4 pinning test dies here.** `folio-go/template_test.go:129–150`
(`Render(tpl)` vs `Render(nil)` byte-identity) **is deleted, not weakened** *(mechanism: binding,
D-1.5.5)*. Its own comment names this story. D-1.5.5: *"The test was built in Story 1.4 to fail and
force its own deletion (D-1.4.16); **this is the story that does it, and the finisher deletes it
rather than weakening it.** First self-retiring assertion in this run to reach its expiry as
designed."*

**AC28.** **A new provisional needing the same treatment** *(mechanism: illustrative, D-1.5.5)*.
AD-24 makes element `x`/`y` **band-relative**, and bands are placed by `internal/layout`, which does
not exist until Story 2.5. So 1.5 adopts a **provisional origin convention** to place text at all.
It is documented as provisional **and pinned by an assertion that fails when band composition
arrives** — *"otherwise the provisional convention quietly becomes the real one."* The mechanism is
illustrative: **a better one may be substituted, the property may not.**

### Standing rules

**AC29.** **D-000.11 — every gate in this story runs `-count=1`.** Go's test cache does not track
out-of-module reads; a cached `ok` is not evidence. **This story reads `fixtures/` (out of module)
and a test font, so it is directly exposed.**

**AC30.** **D-000.9 — every check returns findings *and* a coverage witness; zero candidates is a
failure; errors are assessed separately and first.** For every new check in this story, answer in
the Dev Agent Record: *"What would this have printed if it had been unable to run at all?"* Same
output ⇒ instance. **Ten instances so far**, one of them (D-1.5.4's rejected timestamp comparison)
caught inside this very ruling bundle.

**AC31.** **D-000.10 — the developer dispositions every ruling ID this story enumerates, one line
each.** The complete enumeration is [§ Ruling disposition table](#ruling-disposition-table-d-00010)
below; fill it in, do not rebuild it.

---

## Decisions surfaced by this story, and their rulings

> Both conflicts this story raised under **D-1.2.6** were ruled before it left drafting, together
> with the golden-fixture collision and three confirmations. **Nothing is parked.** The reasoning
> is carried in full because **three `(mechanism: binding)` clauses were reversed**, and a
> developer who does not know why will re-derive a losing option.

### D-1.5.7 — the sorted-input mechanism is WITHDRAWN as inert; two proofs replace it

**Raised as this story's DECISION NEEDED 1. Ruled: withdrawn.** D-1.5.4's clause *"Pass a sorted,
deduplicated slice into `subset.Subset` / `CreatePlan`, so determinism is ours by construction"*
**(mechanism: binding)** is retired.

**Worse than vacuous.** `Input.glyphs` is a `map[GlyphID]bool` and `AddGlyphs` writes into it, so
**input order dies at the door.** This story's measurements are the evidence: 6860 bytes for both
rune orders; 2684 bytes for both glyph-id orders (Dev Notes F-2). Determinism comes from
`createCompactMapping`'s `sort.Slice` at `plan.go:334–350`, **inside the dependency**.

**The sharpest consequence, and a named rule** *(mechanism: binding)*: sorting our input **made the
permutation test vacuous**, because permuted-then-sorted inputs **are the same input**. The
defensive normalisation **deleted the check's ability to fail**.

> ***A defensive normalisation upstream of a check can delete that check's discriminating power.***

That is a **new entry for the vacuity-guard list**, distinct from the existing D-000.9 instances —
the defect is not a check blind to its subject, but a *safety measure* that removes the subject
before the check runs.

**Replacements** — see AC8. Repeat-invariance (N ≥ 16 calls in one process) is buildable because
**Go randomises map iteration per `range` statement, not once per process**: the lead measured an
8-element map ranged 200 times in one process yielding **8 distinct iteration orders**.
Permutation-invariance is buildable **only because we no longer sort**.

**Two sites, opposite answers** *(mechanism: binding)*: **subset input → do NOT sort**;
**tag derivation → MUST sort**, because there the ordering is ours, reaches an output byte, and
comes from a map range — **`ScanMapRange` is load-bearing at that site.**

### D-1.5.8 — the tag hashes `Plan.GlyphSet()`, in source-font numbering, sorted ascending

**Raised as this story's DECISION NEEDED 2. Ruled.** Verified: the method's doc comment reads
*"returns the set of **old** glyph IDs to retain"* (`subset/plan.go:368`). Both rejected readings,
with their reasons, are in **AC7a**; the fixture that permanently kills the catastrophic one is
**AC7b**.

The reasoning worth carrying: the tag exists so a consumer can tell two subsets of one base font
apart (ISO 32000-1 §9.6.4). Hashing the *requested* set lets two different embedded programs carry
the same tag. Hashing the *output* numbering is worse — `createCompactMapping` makes that set
always `{0, 1, … n-1}`, so the tag becomes **a function of glyph count alone**. And **a tag that
moves on a dependency upgrade is a correct signal**, since AD-22 already classes a subsetting
change as a versioned breaking change.

### D-1.5.9 — `fixtures/minimal-rect/` is retained unchanged and reclassified; option (a) is impossible

**Raised as this story's F-8. Ruled (b).** The verdict and its consequences are **AC14a** and
**AC14b**; the seam consequence is **AC10a**.

**Option (a) — "author a `.folio` that reproduces the 547 bytes" — is out on the merits, provably.**
The golden's content stream is exactly `72.5 100.25 200.125 50 re` — **no colour operator**. Story
1.1 omitted it deliberately, relying on PDF's default black fill, because *"adding `0 g` would be a
number that is not a `geom.Length`"*. A template-authored rectangle draws from `style.background`,
and `#000000` goes through **D-1.1.b**'s colour rule and **emits a colour operator**. No `.folio`
can reproduce those bytes without a "black background → emit nothing" special case, **which nobody
should own forever.**

**Why the bind dissolved rather than being traded away:** it existed *only* from binding the fixture
to the public API, which **D-1.1.c designed to change every story**. **The fixture's subject was
never the API** — it was the emission path. Reclassifying it costs nothing and AD-21's *"do not
regenerate"* is fully honoured.

### D-1.5.10 — three confirmations, one of which withdraws two more binding clauses

1. **This story's F-4 seam is adopted as D-1.5.2's binding form** — see **AC17a**. *"A vendor type
   we cannot constrain must not become part of a seam we rely on being constrained."*
2. **The face-age trap killed the weaker timestamp half, on its own** *(corrected, Finding 10, this
   story's QA review — see D-1.5.10 (corrected) in the decision log)*. See **AC11a**. The
   originally-stated reason — "not near now" is unbuildable because D-1.3.1 bans `time` in
   `_test.go` under `internal/`, which is where this test must live — is **false**: the shipped
   test lives at `folio-go/render_test.go`, the module root, outside D-1.3.1's scope entirely. The
   withdrawal still holds on the argument that was always sufficient by itself: Roboto's committed
   face is permanently ≈3 decades old relative to "now" (measured `modified` ≈ 3.57e9 vs ≈ 3.87e9
   for the present), so "not near the current time" can never discriminate a timestamping
   subsetter from a clean one — a meaningless assertion, not merely an illegal one. And D-1.5.4's
   *"or zero"* latitude is tightened away, because this story measured that `subsetHead` copies the
   whole table — **equality is what happens**, and latitude would mask a future change.
3. **This story's subprocess `FontFile2` guard is endorsed as binding** — see **AC10b**.

---

## Decisions resolved

Quoted verbatim with the ruling ID and the mechanism tag. **Never paraphrase these into code
comments** — quote them.

### D-1.5.1 — allowlist + independent denylist

> *"Two assertions in `gomod_test.go`, not one. **Allowlist (primary).** `go list -m all` must
> equal **exactly** the expected set — after this story
> `{github.com/panitw/folio/folio-go, github.com/boxesandglue/textshape}` — asserted by name, with
> the difference printed both ways on failure. *(mechanism: binding)* **PDF-writer denylist
> (independent backstop).** No module path matching a known Go PDF writer … at any depth.
> *(mechanism: binding for the property; illustrative for the list, which must be extendable)*"*

> *"**A denylist alone is weaker than the count it replaces:** a PDF writer nobody listed passes
> silently, and lists rot. The count assertion's *spirit* was right — anything new must be a
> conscious act — and only its *form* was brittle."*

> *"The graph assertion **stays in `gomod_test.go`** — it is an **AD-6** property … and must fail
> in the fast per-story gate. The **AD-26 licence** half stays in `lint/`. Different invariants,
> different homes; do not merge them."*

### D-1.5.2 — validate at the boundary; no second `ScaleRound`

> *"`internal/fontset` validates `unitsPerEm` **before** any value derived from a caller-supplied
> font reaches `geom.ScaleRound`. Valid range **16–16384** …; anything outside — including `0` — is
> a located load error naming the font and the value. At Story 1.5 that is a plain Go error
> (`internal/diag` is Story 3.6 …). *(mechanism: binding)*"*

> *"The panics are correct as programmer-error guards; the fix is **layering, not a second door**.
> Untrusted bytes are validated at the boundary, after which the internal invariant genuinely holds
> — so the spine's *'nothing in `internal/` panics on malformed input'* is satisfied, because by
> the time bytes reach `ScaleRound` they are no longer malformed."*

### D-1.5.3 — `textshape` clears AD-26

> *"Cleared. Measured at `v0.0.15`, twice, by both the lead and the orchestrator: Licence **MIT**;
> `go list -m all` in a scratch module — **the module alone**, zero transitive dependencies;
> `\"time\"` imported in non-test code — **none**."*

**Re-measured a third time by this story's creator, independently — all three confirmed.** See Dev
Notes F-1.

### D-1.5.4 — reject the vacuous comparison *(corrected in three places)*

Quoted in AC6, AC7, AC11, AC12 above. **What survives:** the rejection of the two-run timestamp
comparison *(mechanism: binding)*; "hash the resolved glyph ids, not the input runes"; "one subset
per font per document"; the offsets *(mechanism: illustrative)*.

**What was corrected, all three by measurement taken in this story:**

| Clause | Original tag | Outcome | Ruling |
|---|---|---|---|
| "pass a sorted, deduplicated slice … determinism by construction" | binding | **withdrawn as inert** | D-1.5.7 |
| "equal to the source's values, **or zero**" | binding | **tightened to equality alone** | D-1.5.10 |
| "never a value near the current time" | binding | **withdrawn — on the face-age argument alone** *(the D-1.3.1 rationale first given for this was itself corrected — Finding 10, D-1.5.10 (corrected))* | D-1.5.10 |

The ambiguity in *"the resolved glyph ids"* is settled by **D-1.5.8**: the closure set,
`Plan.GlyphSet()`, in source-font numbering.

### D-1.5.5 — signature, insertion order, and the dying test

> *"Parameters are always in the **final target order**, and each story **inserts** its parameter
> into its final position rather than appending. Target is `Render(t, d, p, f)`, so: **1.5 →
> `Render(t, f)`; 1.6 → `Render(t, d, f)`; 1.7 → `Render(t, d, p, f)`.** No call site is ever
> reordered, only extended. *(mechanism: binding)*"*

> *"Story 1.5's ACs require *'a document using a subset of a Latin face'* and *'the text renders
> correctly'*. Text has to come from somewhere, and **the template is the only source of text in
> the system** — Story 1.6 adds *data binding*, not text. So `t` starts mattering here."*

### D-1.5.6 — `folio-go/fonts/` is a package as well as a directory

> *"AD-8's Rule answers it directly: *'no package under `internal/` embeds font data… The
> repository holds font binaries in exactly one place, `folio-go/fonts/`; **the `folio/fonts`
> package** wraps them with `go:embed` for native callers.'* So `folio-go/fonts/` is simultaneously
> the binary directory **and** a Go package, sitting at the **module root, not under `internal/`**.
> Both rules hold at once: the embed lives outside `internal/`, and `go:embed` reaches it because
> it is inside the module."*

**The apparent AD-8 conflict is not one.** Do not re-raise it.

---

## Tasks / Subtasks

1. **Add the dependency and re-prove the graph guards.** Add `github.com/boxesandglue/textshape
   v0.0.15` to `folio-go/go.mod` (spine §Stack pins v0.0.15). Replace
   `TestModuleGraphHasNoThirdPartyDependencies` with the AC20 allowlist and the AC21 denylist as
   **two separate test functions**. Red-proof both: the allowlist against a scratch graph missing
   and containing an extra module; the denylist against a scratch graph containing one of the six
   names. (AC20, AC21, AC22)
2. **Run the licence check over the now-non-empty graph** and regenerate `lint/MANIFEST.md`. Verify
   `textshape` appears with `Serves: folio-go` / `Shipped`. Paste the table into the Delivery Log.
   Record the AC24 toolchain-directive non-issue. (AC23, AC24)
3. **Commit the Latin test face with its licence text and copyright line**, and account for it in
   the manifest. Check Story 1.3's absence tripwires for `folio-go/fonts/` (DW-2) — landing a font
   binary may fire one, and that is the tripwire working, not a defect. (AC25)
4. **Declare `FontSet` in `folio/fontset.go`** and build `internal/fontset` with a constructor that
   validates `unitsPerEm` and exposes only the validated value. **Do not leak `*ot.Font`,
   `*ot.Head` or `*subset.Plan` out of `internal/fontset`** — binding, D-1.5.10, Dev Notes F-4.
   (AC1, AC2, AC16, AC17, AC17a)
5. **Red-proof the `unitsPerEm` boundary** with three retained fixtures: `0`, `> 16384`, and a
   valid face. The `0` case must produce a **located error, not a panic** — assert the error text
   names the font and the value, not just that an error occurred. (AC19)
6. **Assert `internal/geom` gained no second exported scaling function.** (AC18)
7. **Change `Render`'s signature to `Render(t *Template, f FontSet)`**, and make it **require a
   non-nil template**, returning a located error on nil (AC14b). Update every call site — the
   enumeration is Dev Notes F-7. (AC1)
8. **Delete `folio-go/template_test.go:129–150`** — the `Render(tpl)` vs `Render(nil)` pinning
   test. Do not weaken it. (AC27)
9. **Repoint `TestRenderMatchesGoldenFixture` at `pdf.Serialize()` directly**, leaving
   `fixtures/minimal-rect/` — hash, bytes and README — **entirely untouched**. (AC14a)
10. **Build the text render path** — template-driven text, provisional band origin, `Type0` /
    `Identity-H` / `FontFile2` / `ToUnicode` in `internal/pdf`. Document the provisional origin and
    pin it with an assertion that goes red when `internal/layout` arrives. (AC5, AC28)
11. **Derive the six-letter tag** from `Plan.GlyphSet()` — the closure set, source-font numbering,
    **sorted before hashing**, iterated `ScanMapRange`-compliantly. (AC6, AC7)
12. **Ship AC7b's retained fixture**: two documents whose glyph sets differ but are the same size
    produce **different** tags. Red-proof it against the output-numbering derivation, which is the
    reading it exists to kill. (AC7b)
13. **Do NOT sort the subset input.** Wire the two replacement proofs: **repeat-invariance**
    (N ≥ 16 calls in one process, all byte-identical to the first) and **permutation-invariance**
    (distinct input orders ⇒ identical bytes). Both are void if a sort is reintroduced upstream.
    (AC8, AC8a)
14. **Assert the one-subset-per-document property behaviourally** with a two-page document using
    one face: exactly one `FontFile2` in the output. (AC9)
15. **Give the subprocess seam a second selector** so it renders **both** the fontless document and
    the template+font document, and **run the matrix over both**. Wire the binding vacuity guard:
    the font child's output must contain a `FontFile2` before any byte comparison. (AC10, AC10a,
    AC10b)
16. **Wire the single-run head-timestamp assertion** at offsets 20 and 28 of the embedded
    `FontFile2`'s head table: **equality with the source face, exactly**. No "or zero" latitude, and
    **no clock read** — D-1.3.1 bans `time` in `_test.go` under `internal/`. (AC11a)
17. **Record the measured provenance in the Delivery Log** — no assertion over `GOMODCACHE`.
    (AC12)
18. **Validate with an independent reader** on a machine with none of the document's fonts
    installed; record in the Delivery Log. (AC13)
19. **Record this story's new golden fixture, beside `minimal-rect/`.** (AC14)
20. **Run the full four-target matrix in this story**, for both documents, and record every hash.
    (AC15)
21. **Ensure every gate runs `-count=1`**, including the new font-reading tests. (AC29)
22. **Answer D-000.9's question for every new check** in the Dev Agent Record, and fill in the
    ruling disposition table. (AC30, AC31)

---

## Dev Notes

Everything in this section was **measured in this run**, at baseline `94b453a`, against
`textshape v0.0.15` in the local module cache. Nothing here is inferred from documentation.

### F-1 — D-1.5.3's three checks, independently re-measured (third measurement)

| Check | Method | Result |
|---|---|---|
| Licence | `head -5 $GOMODCACHE/github.com/boxesandglue/textshape@v0.0.15/LICENSE` | `MIT License` — **MIT** |
| Transitive deps | scratch module, `go get textshape@latest`, `go list -m all` | `github.com/boxesandglue/textshape v0.0.15` **alone** |
| `"time"` in non-test code | recursive grep, `--include='*.go'`, excluding `_test.go` | **no matches** |
| Its own `go.mod` | `cat` | `go 1.23.0` + `toolchain go1.24.0` — the AC24 non-issue, confirmed present |

**`lint/MANIFEST.md` at baseline contains eight rows, every one of them `Serves: lint` /
`build-time-only`.** `folio-go` contributes **zero** rows. D-1.5.3's claim that the licence check
*"has only ever run against an empty graph"* is **confirmed by direct inspection of the shipped
manifest**. `textshape` will be the first `Serves: folio-go` / `Shipped` row in the repo's history.

### F-2 — The subsetting API's real shape, and the sorted-input measurement

Measured signatures (module cache, `v0.0.15`):

- `subset.Subset(font *ot.Font, codepoints []rune) ([]byte, error)` — `subset/execute.go:552`.
  **D-1.5.4's API-shape trap is real and confirmed: this convenience function takes codepoints.**
- `subset.CreatePlan(font *ot.Font, input *Input) (*Plan, error)` — `subset/plan.go:53`.
- `(*subset.Plan).Execute() ([]byte, error)` — `subset/execute.go:17`.
- **There is a glyph-id door**: `(*Input).AddGlyph(gid ot.GlyphID)` and
  `(*Input).AddGlyphs(gids ...ot.GlyphID)` — `subset/input.go:98,103`. So the trap is avoidable by
  driving `CreatePlan` with an `Input` rather than calling `Subset`.
- `(*Input).Glyphs() map[ot.GlyphID]bool`, `(*Plan).GlyphSet() map[ot.GlyphID]bool`,
  `(*Plan).GlyphMap() map[ot.GlyphID]ot.GlyphID` — **all three return maps.** Any of them reaching
  the tag derivation puts `ScanMapRange` directly in the path, exactly as D-1.5.4 says.
- `Flags` includes **`FlagDropLayoutTables`**, whose own doc comment reads *"Use this for PDF
  embedding where shaping is already done."* It also **suppresses GSUB closure**
  (`plan.go:174–179`), which shrinks the embedded glyph set and narrows the determinism surface.
  Worth considering; not mandated by any ruling.
- **Avoid the variable-font surface.** `(*Input).PinAxisLocation(axisTag ot.Tag, value float32)`
  takes a `float32`. `internal/arch_test.go`'s no-float guard flags the **bare identifier**
  `float32`/`float64` anywhere in the AST under `internal/`, so a call site that names the type
  would go red. Nothing in this story needs axis pinning.

**The sorted-input measurement** (scratch module, `Roboto-Regular.ttf` from `textshape/testdata`):

```
Subset ascending vs reversed rune order byte-identical: true (len 6860/6860)
Input.AddGlyphs unsorted+dup vs sorted byte-identical: true (len 2684/2684)
```

**This measurement is the evidence that withdrew a binding clause — see D-1.5.7 and AC8.** Root
cause: `Input.glyphs` is `map[ot.GlyphID]bool` (`input.go:12`); `AddGlyphs` writes into it and
`computeGlyphClosure` ranges it (`plan.go:167`); **input order dies at the door**, before anything
downstream sees it. The library's own determinism comes from `createCompactMapping`'s `sort.Slice`
at `plan.go:334–350`, **inside the dependency**, where `ScanMapRange` cannot reach.

**Do not re-introduce the sort at this site.** Sorting here would make AC8's permutation-invariance
proof vacuous, because permuted-then-sorted inputs are the same input (D-1.5.7's named rule:
*a defensive normalisation upstream of a check can delete that check's discriminating power*).

### F-3 — The requested glyph set is not the embedded glyph set

```
Plan.GlyphSet() size for requested {5,40,70}: 8 (requested 3)
Plan.NumOutputGlyphs=8
```

`computeGlyphClosure` (`plan.go:152–180`) unconditionally sets `p.glyphSet[0] = true` (`.notdef`),
then adds composite components and GSUB-reachable glyphs. **Ruled as D-1.5.8: the tag hashes this
closure set, in source-font numbering, sorted ascending** — see AC7, and AC7a for the two rejected
readings.

**The catastrophic near-miss, verified in source.** `createCompactMapping` (`plan.go:334–350`)
sorts the retained old GIDs and then does `for newGID, oldGID := range gids`, assigning
`newGID = index`. **The output glyph-id set is therefore always `{0, 1, … n-1}`** — a tag hashed
over it would be **a function of glyph count alone**, colliding across two entirely different
8-glyph subsets. AC7b's retained fixture (glyph sets differing but the same size ⇒ different tags)
exists to keep that reading dead.

### F-4 — `unitsPerEm` is exported and unvalidated on the third-party type

`ot.Head` (`ot/metrics.go:25–44`) declares `UnitsPerEm uint16` as an **exported field**, populated
by `ot.ParseHead` from offset 18 with **no range check**. `(*Plan).Source() *ot.Font`
(`plan.go:374`) also hands the raw font straight back.

D-1.5.2's binding guardrail — *"the parsed-font type must not expose a raw, unvalidated
`unitsPerEm`"* — therefore **cannot be satisfied by the third-party type**, which we do not
control. It binds **our** type: `internal/fontset`'s parsed-font value must **wrap** the
third-party font rather than embedding or re-exporting it, and must not surface `*ot.Font`,
`*ot.Head`, or a `*subset.Plan` to any caller that could read `UnitsPerEm` before validation. This
is a constraint on the seam's **shape**, not just on a check's placement — get it wrong and AC17
is satisfied in letter and defeated in fact.

**Adopted as D-1.5.2's binding form by D-1.5.10 — see AC17a.** *"A vendor type we cannot constrain
must not become part of a seam we rely on being constrained."*

Also note `uint16`: the range `16–16384` sits inside `0..65535`, so **both** negative polarities in
AC19 are constructible, and `0` is representable. A signed-vs-unsigned mistake would make the
`> 16384` case unreachable — check it goes red.

### F-5 — The head table is copied verbatim, so equality is the assertion

`(*Plan).subsetHead` (`subset/execute.go:85–101`) reads the source `head` table, does
`newData := make(...); copy(newData, data)`, and patches **only** `indexToLocFormat` at offset 50.
`created` (offset 20) and `modified` (offset 28) are therefore **byte-copies of the source**.
Measured end to end on `Roboto-Regular.ttf`:

```
SOURCE head.created=3304067374 head.modified=3573633780
SUBSET head.created=3304067374 head.modified=3573633780  equal-to-source=true
SOURCE unitsPerEm=2048
```

**Equality is what actually happens**, by a `copy` that is legible in the dependency's source. That
is why **D-1.5.10 tightened D-1.5.4's *"equal to the source's values, or zero"* to equality alone**
(AC11a): the "or zero" latitude would mask a future change rather than catch it.

**And why the other half went away entirely.** *"Never a value near the current time"* would only
discriminate if the committed test face's own head times were far from the present — Roboto's
`modified` is `3.57e9` against roughly `3.87e9` for 2026-08-23, so a recently built face would have
made it a false red or a meaningless green (D-000.9). **But the clause could never have shipped
regardless: asserting "not near now" requires reading the clock, and D-1.3.1 bans `time` in
`_test.go` under `internal/`, which is where this test lives.** Withdrawn as unbuildable, not
merely weak (D-1.5.10). Still record the committed face's measured `created`/`modified` in the
Delivery Log — it documents what the equality assertion is pinned to.

### F-6 — The subprocess seam renders no font at all, and both documents must keep a seam

`folio-go/render_test.go`'s `TestMain` routes `FOLIO_SUBPROCESS_RENDER=1` to `Render(nil)`, which
today emits Story 1.1's fixed rectangle — **no font, no `FontFile2`.** Story 1.2's matrix harness
(`folio-go/matrix_test.go`, `//go:build matrix`) drives the same seam.

**Left as is, AC10 passes on two identical fontless rectangles and certifies nothing** — and so
does AC15's four-target matrix, in the story whose central claim is font byte-stability.
**Endorsed as binding by D-1.5.10** (AC10b): assert the child output contains a `FontFile2` before
comparing. **The D-000.4 matrix override is only meaningful if the subprocess path actually embeds
a font.**

**But do not simply repoint the seam** *(D-1.5.9, AC10a)*. If `FOLIO_SUBPROCESS_RENDER=1` switches
to the font document, **`minimal-rect` loses its four-target verification** — the property that
makes it the repo's most valuable artifact. **The seam renders both:** the existing selector keeps
the fontless path, a **second selector** covers template+font, and **the matrix runs both.**

### F-7 — The `Render(nil)` call sites, enumerated

| File | Site | What changes |
|---|---|---|
| `folio-go/render_test.go` | `TestMain` subprocess seam | keeps the fontless path; **gains a second selector** for template+font (AC10a) |
| `folio-go/render_test.go` | `TestRenderProducesAValidPDF` | signature |
| `folio-go/render_test.go` | `TestRenderHasNoCreationOrModDate` | signature; **its forbidden-key list includes `/Filter`** — if this story compresses the `FontFile2`, that assertion goes red |
| `folio-go/render_test.go` | `TestRenderIDEntriesAreIdentical` | signature |
| `folio-go/fixture_test.go:273` | `TestRenderMatchesGoldenFixture` | **calls `pdf.Serialize()` directly**, not `Render` (AC14a, F-8) |
| `folio-go/template_test.go:115` | `TestRenderAcceptsAParsedTemplate` | signature |
| `folio-go/template_test.go:129–150` | the pinning test | **deleted** (AC27) |

### F-8 — Why `fixtures/minimal-rect` cannot be reproduced from a template — RULED (D-1.5.9)

`TestRenderMatchesGoldenFixture` (`folio-go/fixture_test.go:227+`) renders `Render(nil)` and
compares SHA-256 against `fixtures/minimal-rect/expected.json`
(`0f925e1b…4f7c`, `goToolchain: go1.26.0`). **The fixture directory contains no `.folio` template
at all** — only `expected.json`, `expected.pdf` and `README.md`.

**The lead verified why authoring one is impossible, not merely awkward.** The golden's content
stream is exactly `72.5 100.25 200.125 50 re` — **no colour operator**. Story 1.1 omitted it
deliberately, relying on PDF's default black fill, because *"adding `0 g` would be a number that is
not a `geom.Length`"*. A template-authored rectangle draws from `style.background`, and `#000000`
goes through **D-1.1.b**'s colour rule and **emits a colour operator**. So no `.folio` reproduces
those 547 bytes without a *"black background → emit nothing"* special case — **which nobody should
own forever.**

**Verdict (AC14a, AC14b).** The fixture is **retained unchanged** — hash, bytes, README and its
*"do not regenerate"* instruction all stand — and **reclassified** as pinning the **fontless
emission path in `internal/pdf`** (AD-3 number emission, xref arithmetic, content-derived `/ID`).
Its test calls **`pdf.Serialize()` directly**. `Render` may then **require a non-nil template**,
returning a located error on nil — better than a public entry point documented as ignoring its
argument. Story 1.5 adds a **new** template-driven fixture beside it.

**Why the bind dissolved rather than being traded away:** it existed *only* from binding the fixture
to the public API, which **D-1.1.c designed to change every story**. **The fixture's subject was
never the API.**

### F-9 — Where the governing rulings actually live

The story brief stated the decision log was committed at `94b453a` with a clean tree. **Verified
wrong, and the coordinator has confirmed the correction.** D-1.5.1 – D-1.5.10 live in the
**working tree** copy of `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`, **not
in `HEAD`** — 187 uncommitted lines at the time of first measurement, more since. They will land in
a commit that is **not this story's**, so **a reviewer diffing against `94b453a` will not find
them there.** The content quoted throughout this file was read from the working tree and is
complete.

---

## Ruling disposition table (D-000.10)

One line each, filled in by the developer. **Do not rebuild this list** — it is the complete
enumeration D-000.10 requires. Twenty-five entries *(corrected from "Twenty-four" — Finding 7, this
story's QA review: the original table omitted `D-1.1.c`, which the header enumerates and which is
load-bearing in D-1.5.9's own argument)*; the four ruled during this story's drafting are marked.

| Ruling | Disposition |
|---|---|
| D-1.5.1 | applied-as-stated — `gomod_test.go` carries `TestModuleGraphAllowlist` (exact-set allowlist, two-way diff printed on failure) and `TestModuleGraphDenylistsKnownPDFWriters` (independent backstop) as two separate test functions, each with its own red-proof (`TestDiffModuleSetsCatchesExtraAndMissing`, `TestModuleGraphDenylistRedProof`). Graph assertion stays in `gomod_test.go`; licence half stays in `lint/`. |
| D-1.5.2 | applied-as-stated — `internal/fontset.New` validates `unitsPerEm` (16–16384, `0` included) before any font-derived value can reach `geom.ScaleRound`, returning a plain Go error naming the font and the value; no panic. `internal/geom` gained no second scaling function (`TestExactlyOneExportedScalingFunction`). |
| D-1.5.3 | applied-as-stated — independently re-measured a fourth time this story: MIT licence, zero transitive deps (`go list -m all` after `go get`), no `"time"` import in non-test code, `go 1.23.0`/`toolchain go1.24.0` confirmed present in `textshape`'s own `go.mod` (matches F-1 exactly). |
| D-1.5.4 *(corrected in three places)* | applied-as-stated — implemented in its corrected form: single-run equality-only head-timestamp assertion (no "or zero", no clock read), tag hashes the resolved glyph ids (not input runes), one subset per font per document. The three corrections are separately dispositioned under D-1.5.7/D-1.5.10 below. |
| D-1.5.5 | applied-as-stated — `func Render(t *Template, f FontSet) ([]byte, error)`, `f` inserted in final target position. `template_test.go:129–150`'s `TestRenderIgnoresItsArgumentToday` deleted, not weakened (AC27). |
| D-1.5.6 | not-reached-in-this-story for the `folio-go/fonts/`-as-package mechanism itself — that is Story 2.2's shipped-face directory, and this story's one test face deliberately lives at `folio-go/testdata/fonts/` instead (Task 3: landing it under `folio-go/fonts/` would trip the DW-2 absence tripwire reserved for 2.2). AC26's scope fence, the same ruling bundle's other half, **is** applied as stated: no production faces shipped, one small Latin test face only. |
| **D-1.5.7** *(new — sorted input withdrawn as inert)* | applied-as-stated — `Font.Subset` never sorts its glyph-id input; it ranges only the caller's `[]rune` slice (never a map) to preserve the caller's own ordering while still deduplicating. `TestRepeatInvarianceWithinOneProcess` (N=32) and `TestPermutationInvariance` both pass, in `internal/fontset`. See note below on the ScanMapRange interaction this ruling's mechanism raised. |
| **D-1.5.8** *(new — the tag hashes `Plan.GlyphSet()`)* | applied-as-stated — `deriveTag` hashes `plan.GlyphSet()` (source-font numbering), sorted ascending via `slices.Sorted(maps.Keys(...))` before hashing. `TestDeriveTagUsesClosureSetNotOutputNumbering` (AC7b, "ABC" vs "DEF", both `NumGlyphs=8`, different tags) and its red-proof against the output-numbering reading both pass. |
| **D-1.5.9** *(new — `minimal-rect` retained and reclassified)* | applied-as-stated — `fixtures/minimal-rect/` bytes, hash and README confirmed byte-for-byte unchanged (`git diff` empty; matrix run reproduced the identical recorded hash `0f925e1b…` on all four targets). `TestRenderMatchesGoldenFixture` repointed to `pdf.Serialize()` directly. `Render(nil, …)` returns a located error (AC14b). New fixture `fixtures/font-text/` added beside it. |
| **D-1.5.10** *(new — three confirmations, two withdrawals)* | applied-as-stated — `internal/fontset.Font` wraps `*ot.Face` privately and exposes only primitive-typed accessors (`UnitsPerEm() uint16`, `HeadTimes() (int64, int64)`, `Metrics()`, `Subset(...)`); no `*ot.Font`/`*ot.Head`/`*subset.Plan` crosses the package boundary. `TestEmbeddedHeadTimestampsEqualSourceExactly` asserts equality alone, no clock read. `AC10b`'s `FontFile2` guard is wired into the local two-process test, the golden-fixture test, and the matrix. |
| D-000.4 | applied-as-stated — the full four-target matrix (`TestCrossTargetByteIdentity`) ran in this story, for both documents, via Docker (darwin host + linux/amd64 + linux/arm64 containers) and Node (js/wasm); see the pasted table below. |
| D-000.5 | not-reached-in-this-story — this ruling concerns the licence check's `folio-designer/`-vs-`designer/` naming, a Story 5.1 concern this story's redistributed-asset accounting (AC25) never touches. |
| D-000.6 | not-reached-in-this-story — this story never found `ARCHITECTURE-SPINE.md`, `folio-format.md` or another SPEC companion stating something false or unimplementable; no document amendment was needed. |
| D-000.9 | applied-as-stated — every new check in this story answers "what would this print if unable to run at all?" with a coverage witness distinct from "zero findings": `internal/fontset`'s subset tests assert real glyph/tag output (not silent zero-value structs); `lint`'s new `ScanEmbedFont` reports `FilesParsed` and fails the build if it parsed zero files, exactly like `ScanMapRange`/`ScanAbsences`/`scanNoFloat64`'s existing pattern; `TestSerializeTextDocumentEmbedsOneFontFileForTwoPages` asserts `>=2` page objects and the exact program-byte count, not just "no error". |
| D-000.10 | applied-as-stated — this is that disposition table. |
| D-000.11 | applied-as-stated — every gate in the Delivery Log below was run with `GOFLAGS=-count=1` (or `go test -count=1` directly), including the new font-reading tests, which read `fixtures/`, `testdata/fonts/` and the go:embed'd test font — all out-of-module-cache-tracked reads. |
| D-1.1.a | applied-as-stated — the `go`/`toolchain` directive split in `folio-go/go.mod` was untouched by this story's `go get` (`TestGoModPinsToolchainExactly` still passes unmodified), and AC24 directly invokes the addendum's recorded position ("a `toolchain` directive binds only the **main** module") to establish that `textshape`'s own `go 1.23.0`/`toolchain go1.24.0` is not a pin violation. |
| D-1.1.b | applied-as-stated *(corrected from "applied-with-a-different-mechanism" — Finding 4, this story's QA review)*. `internal/pdf`'s existing two representations (`appendLength`, `appendInt`/`appendIntPadded`, confined to `numbers.go`) are untouched and still the only routes for geometric/count values. The Identity-H CID hex strings and the six-letter subset tag are D-1.1.b's own explicitly-named exemption ("Glyph ids under Identity-H take neither route ... Same for the six-letter subset tag") — not a departure requiring its own ruling, so the original "different-mechanism" flag was an over-report. The ruling's own next sentence — *"This must be said in a code comment or a later agent will 'unify' it"* — was the actual gap: the byte-encoding statement was missing from all three sites (`appendHex4`, `appendHexCIDString`, `deriveTag`). Fixed: the ruling is now quoted verbatim in a code comment at all three. |
| **D-1.1.c** *(added — Finding 7, this story's QA review; originally omitted)* | reached-and-applied — `Render`'s signature changed exactly as D-1.1.c anticipates ("designed to change every story": `Render(t, f)` at 1.5, inserting `f` into its final target position per D-1.5.5). Cited by name in `folio.go`'s `Render` doc comment and load-bearing in D-1.5.9's own argument (*"it existed only from binding the fixture to the public API, which D-1.1.c designed to change every story"*) — it required a disposition line and did not have one. |
| D-1.2.6 | applied-as-stated — one apparent conflict surfaced during implementation (D-1.3.5's `ScanMapRange` total ban on map-ranging under `internal/`, vs D-1.5.7's requirement that the subset-input site NOT be sorted, i.e. not go through `ScanMapRange`'s sorted escape hatch). Resolved by construction, not by choosing one ruling over the other: `Font.Subset` ranges only the caller's slice (never a map) at every site that must stay unsorted, using map lookups (not map ranges) for deduplication — both rulings hold simultaneously, verified by `lint`'s `TestMapRangeProductionScan` passing AND `TestPermutationInvariance` passing. Surfaced here per D-1.2.6's spirit rather than resolved silently; flagged for reviewer double-check since it is exactly the shape of conflict this ruling exists to police. |
| D-1.3.1 | applied-as-stated — no `_test.go` file added or touched under `internal/` in this story imports `"time"` (confirmed by grep); `TestEmbeddedHeadTimestampsEqualSourceExactly` reads two already-parsed head tables' `created`/`modified` fields and asserts equality, with no clock read anywhere. |
| D-1.3.4 | applied-as-stated — this story's new redistributed-asset accounting (AC25, `manifest.ResolveAssets`) follows the same "assert absence until the dependent half is wired" shape D-1.3.4 established; it extends the Go-half-adjacent manifest, not a JS half. |
| D-1.3.6 | applied-as-stated — the new `ScanEmbedFont` guard (AC3) lives in `lint/internal/rules`, the same standalone repo-root module every other guard lives in. |
| D-1.3.9 | applied-as-stated — regenerated `lint/MANIFEST.md` labels `github.com/boxesandglue/textshape` as `Serves: folio-go` / `Shipped` — the first non-`lint` shipped row in the repo's history (AC23). |
| D-1.4.15 | applied-as-stated — every named mechanism in this story's own code comments carries a `(mechanism: binding)`/`(mechanism: illustrative)` tag matching the story file's tags; behavioural tests (repeat/permutation-invariance, AC7b's same-size-different-tag fixture, AC9's two-page/one-`FontFile2` fixture) were added alongside structural ones, extending D-1.4.15's "AST extraction is necessary but not sufficient" pattern to font subsetting. |

---

## Dev Agent Record

### Delivery Log

- [x] `lint/MANIFEST.md` regenerated, table pasted, `textshape` row shown as `folio-go` / shipped (AC23) — see table below.
- [x] Measured `textshape` provenance recorded as evidence, not asserted (AC12) — see below.
- [x] Committed test face's `head.created` / `head.modified` recorded — what the equality assertion is pinned to (AC11a, F-5) — `created=3304067374`, `modified=3573633780`, matches F-5 exactly, asserted in `TestEmbeddedHeadTimestampsEqualSourceExactly` (`folio-go/render_test.go`).
- [x] Independent-reader validation on a fontless machine (AC13) — `qpdf --check` on `fixtures/font-text/expected.pdf`: *"No syntax or stream encoding errors found."* PyMuPDF (Python, a fully independent PDF implementation with no relationship to `folio-go`) `page.get_text()` on the same PDF returns exactly `"Hello, World!\nPage footer 0123456789\n"` — confirming the embedded subset's glyph shapes, positions and `ToUnicode` CMap all round-trip correctly, not merely that the file is well-formed. Rendered to a 100dpi PNG for visual confirmation as well; header/content/footer bands are correctly separated (no overlap) after the provisional band-origin math (AC28) accounts for header/footer band heights.
- [x] Four-target matrix hashes, **all four legs for both documents** (AC15, AC10a, D-000.4 override) — see table below. Ran via Docker (linux/amd64, linux/arm64 containers) + Node (js/wasm) + native (darwin/arm64), `go test -tags=matrix . -run TestCrossTargetByteIdentity`.
- [x] Evidence that the font subprocess child actually emits a `FontFile2` — without it the matrix override proves nothing (AC10b) — wired as a hard `t.Fatalf` gate in `TestCrossTargetByteIdentity` (checked per leg, before any hash is computed) and separately in `TestRenderMatchesFontTextGoldenFixture` and `TestRenderWithFontIsByteIdenticalAcrossTwoProcesses`; the matrix run above did not trip it on any of the 4 legs — every leg's font-document render genuinely contains `/FontFile2`.
- [x] Repeat-invariance run recorded with its N (AC8) — `TestRepeatInvarianceWithinOneProcess`, N=32 (>= 16), `internal/fontset/fontset_test.go`; all 32 calls byte-identical to the first.
- [x] `fixtures/minimal-rect/expected.json` and `expected.pdf` confirmed **byte-unchanged** at close-out (AC14a) — `git diff` / `git status --porcelain` against `fixtures/minimal-rect/` is empty; the matrix run above independently reproduced the exact recorded hash `0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c` on all four targets.
- [x] D-000.9 answered for every new check: *"what would this print if it could not run at all?"* (AC30) — see the ten instances below.
- [x] AC24's toolchain non-issue recorded so it is not re-litigated (AC24) — `textshape@v0.0.15`'s own `go.mod` declares `go 1.23.0` and `toolchain go1.24.0`. Per D-1.1.a's addendum, a `toolchain` directive binds only the **main module** (`folio-go`, pinned to `go1.26.0`) — `textshape`'s toolchain line has no effect on the build. `go 1.23.0` is below `folio-go`'s `go 1.25.0` floor, so no minimum-version conflict either. Not a toolchain-pin violation.

#### D-000.9 — ten new-check instances, "what would this print if unable to run at all?"

1. **`TestModuleGraphAllowlist`/`TestModuleGraphDenylistsKnownPDFWriters`** — an `exec.Command` failure or a nil-slice bug would report `len(got)==0`, which the allowlist's `extra`/`missing` diff distinguishes from "matches exactly" (missing would list every expected module); red-proofed directly (`TestDiffModuleSetsCatchesExtraAndMissing`, `TestModuleGraphDenylistRedProof`) without touching `go list`.
2. **`internal/fontset.New`'s `unitsPerEm` boundary** — a scanner that silently accepted any value would pass `TestNewAcceptsAValidFace` but also `TestNewRejectsZeroUnitsPerEm`/`...AboveMax`, which assert an **error**, not just "no crash"; a `New` that always errors would fail the valid-face case. Both polarities plus the valid case are retained fixtures (AC19).
3. **`TestExactlyOneExportedScalingFunction`** — a scanner that visited zero files would report `filesScanned==0` and fail before the emptiness of `exportedFuncs` is even considered (explicit vacuity guard in the test).
4. **`deriveTag`'s AC7b fixture** — a tag function that always returned the same string (e.g. a stubbed `"AAAAAA"`) would still make `TestDeriveTagUsesClosureSetNotOutputNumbering` fail (`subA.Tag == subB.Tag`), not pass; a tag function that panicked would fail loudly, not report false success.
5. **`TestRepeatInvarianceWithinOneProcess`/`TestPermutationInvariance`** — a `Subset` that returned `nil, nil` for every call would make `bytes.Equal(first.Program, got.Program)` trivially true (vacuity risk) *(corrected, Finding 18, this story's QA review: originally closed by relying on `TestSubsetContainsRequestedGlyphs` running first in the same package — a scheduling accident, not a guaranteed dependency; `-run` filtering or a file split breaks it)* — now closed by each test asserting `len(Program) != 0` directly on its own first call, independent of test ordering.
6. **`lint`'s `ScanEmbedFont`** — follows `ScanMapRange`/`ScanAbsences`/`scanNoFloat64`'s exact pattern: `EmbedFontStats.FilesParsed` is asserted `> 0` before `len(findings)==0` is trusted (`TestEmbedFontProductionScan`); a scanner that walked nothing reports `FilesParsed==0` and fails distinctly.
7. **`manifest.ResolveAssets`** — a scan that silently skipped the `folio-go/testdata/fonts/` directory would report zero asset rows, which `TestManifestUpToDate` would catch as a diff against the committed `MANIFEST.md` (which now has a non-empty assets table) — the check is anchored to a committed artifact, not to its own success.
8. **`TestSerializeTextDocumentEmbedsOneFontFileForTwoPages`** — asserts the exact count (`1`), the exact program-byte occurrence count (`1`), AND a minimum page-object count (`>=2`) — a builder that silently dropped all font objects would report `count==0`, not `1`, and is distinguished from "correctly embedded once."
9. **`TestCrossTargetByteIdentity`'s AC10b guard** — checked explicitly, per leg, per document, BEFORE any hash comparison — a subprocess that silently fell back to the fontless path would be caught here, not laundered into "all four targets agree" (which a fontless-vs-fontless comparison would also satisfy).
10. *(carried from the story's own drafting, D-1.5.4's rejected timestamp comparison)* — a same-second two-run comparison cannot fail even if the subsetter timestamps every call; withdrawn rather than shipped (AC11).

#### `lint/MANIFEST.md` (regenerated, AC23, AC25)

```
| Module | Version | Licence | Serves | Shipped / build-time-only |
|---|---|---|---|---|
| github.com/boxesandglue/textshape | v0.0.15 | MIT | folio-go | shipped |
| github.com/google/go-cmp | v0.6.0 | BSD-3-Clause | lint | build-time-only |
| github.com/yuin/goldmark | v1.4.13 | MIT | lint | build-time-only |
| golang.org/x/mod | v0.39.0 | BSD-3-Clause | lint | build-time-only |
| golang.org/x/net | v0.58.0 | BSD-3-Clause | lint | build-time-only |
| golang.org/x/sync | v0.22.0 | BSD-3-Clause | lint | build-time-only |
| golang.org/x/sys | v0.47.0 | BSD-3-Clause | lint | build-time-only |
| golang.org/x/telemetry | v0.0.0-20260811182544-a038080d80e5 | BSD-3-Clause | lint | build-time-only |
| golang.org/x/tools | v0.49.0 | BSD-3-Clause | lint | build-time-only |

## Redistributed non-code assets
| Path | Licence | Copyright | Serves |
|---|---|---|---|
| folio-go/testdata/fonts/Roboto-Regular.ttf | Apache-2.0 | Copyright 2011 Google Inc. All Rights Reserved. | folio-go test fixture |
```

`textshape` is the **first `Serves: folio-go` / `Shipped`** row in the repo's history (D-1.5.3 confirmed by direct inspection: the baseline manifest had eight rows, all `Serves: lint` / `build-time-only`).

#### `textshape` provenance (AC12, evidence not assertion)

- Licence: `MIT License` (`$GOMODCACHE/github.com/boxesandglue/textshape@v0.0.15/LICENSE`).
- `go list -m all` in a scratch module after `go get textshape@latest`: `github.com/boxesandglue/textshape v0.0.15` **alone** — zero transitive dependencies.
- `grep -rn '"time"' textshape@v0.0.15` (all `.go` files, test and non-test): **zero matches** anywhere in the module.
- `ot/metrics.go`'s `Head` struct declares `Created`/`Modified int64`; `ParseHead` only **reads** them from `data[20:]`/`data[28:]` via `binary.BigEndian.Uint64` — no write, no clock.
- `textshape`'s own `go.mod`: `go 1.23.0` / `toolchain go1.24.0` — the AC24 non-issue (see above).

None of this is wired into a module-cache-path assertion (D-1.5.4: "do not turn it into an assertion over the module cache") — it is recorded here as evidence, and the actual guardrails are `TestModuleGraphAllowlist`/`...Denylist` (module identity) and the licence manifest (licence family).

#### Four-target matrix — both documents (AC15, AC10a, D-000.4 override)

```
cross-target render matrix (minimal-rect (fontless)):
darwin/arm64   0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c  547 bytes
linux/amd64    0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c  547 bytes
linux/arm64    0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c  547 bytes
js/wasm        0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c  547 bytes

cross-target render matrix (font-text (template+font)):
darwin/arm64   dcd453a1f593c1135c44b52d732683d19419c24743344d3843826fccbafacbdf  22299 bytes
linux/amd64    dcd453a1f593c1135c44b52d732683d19419c24743344d3843826fccbafacbdf  22299 bytes
linux/arm64    dcd453a1f593c1135c44b52d732683d19419c24743344d3843826fccbafacbdf  22299 bytes
js/wasm        dcd453a1f593c1135c44b52d732683d19419c24743344d3843826fccbafacbdf  22299 bytes

--- PASS: TestCrossTargetByteIdentity (6.43s)
```

All four targets agree with each other AND with the recorded golden, for BOTH documents. The fontless leg reproduces `fixtures/minimal-rect/`'s pre-existing, untouched hash exactly (AC14a). The font-text leg is the first cross-target proof that font subsetting — the property whose determinism depends on a third-party module's internal glyph-map ordering (`textshape`'s `createCompactMapping`) — is genuinely byte-identical across `darwin/arm64`, `linux/amd64`, `linux/arm64` (Docker) and `js/wasm` (Node), not merely within one process on one machine.

**Finisher re-run, after all 23 findings' fixes landed** (`rtk proxy git show` verification per D-000.12, never a raw pipe): `TestCrossTargetByteIdentity` re-run in full, both documents, all four targets. Reproduced **byte-identical**: `minimal-rect` → `0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c` / 547 bytes ×4; `font-text` → `dcd453a1f593c1135c44b52d732683d19419c24743344d3843826fccbafacbdf` / 22299 bytes ×4 — exit 0, `--- PASS: TestCrossTargetByteIdentity (4.51s)`. The fixes touched proof quality and guard reach, never the emission bytes; this confirms it.

#### Finisher gates (all `-count=1`, D-000.11; raw exit codes via `rtk proxy`, never piped, D-000.12)

| Gate | Command | Result |
|---|---|---|
| `folio-go` build | `go build ./...` | exit 0 |
| `folio-go` vet | `go vet ./...` | exit 0 |
| `folio-go` gofmt | `gofmt -l .` | exit 0, no files listed |
| `folio-go` full suite | `go test ./... -count=1` | exit 0 — **99 top-level / 185 with subtests / 6 packages, zero SKIPs** |
| `hashmatrix` vet | `go vet ./...` | exit 0 |
| `hashmatrix` gofmt | `gofmt -l .` | exit 0, no files listed |
| `hashmatrix` plain build | `go build ./...` | exit 1, **by design, unchanged from Story 1.3/1.4** — `go: build output "probe" already exists and is a directory` (the module's own README names this; `probe/` is a directory, not a build target) |
| `lint` build | `go build ./...` | exit 0 |
| `lint` vet | `go vet ./...` | exit 0 |
| `lint` gofmt | `gofmt -l .` | exit 0, no files listed |
| `lint` full suite, `GOPROXY=off` | `GOPROXY=off go test ./... -count=1` | exit 0 — **18 top-level / 39 with subtests / 4 packages, zero SKIPs** |
| `folio-go` matrix build | `go build -tags=matrix ./...` | exit 0 |
| `folio-go` matrix vet | `go vet -tags=matrix ./...` | exit 0 |
| Full four-target matrix, both documents | `go test -tags=matrix . -run TestCrossTargetByteIdentity -v -count=1` | exit 0 — see table above |
| Licence manifest regeneration | `cd lint && go run ./cmd/genmanifest` | exit 0, **no diff** against the committed `lint/MANIFEST.md` (Finding 9's repo-wide walk produces byte-identical output to the old fixed-list scan for both currently-committed assets) |

### Completion Notes

All 22 tasks implemented and verified; every AC (1–31) satisfied — see the AC-by-AC mapping folded into this Delivery Log and the ruling disposition table above. Key implementation decisions:

- **`FontSet = map[string][]byte`** (`folio-go/fontset.go`), keyed by face name as referenced in a `.folio` document's `fonts` fallback chains. `internal/fontset.New(name, data)` validates and wraps each face; `internal/fontset.(*Font).Subset(runes)` does the one-call-per-font-per-document subsetting.
- **`internal/pdf` gained a second document builder** (`builder.go`, `textdoc.go`) alongside Story 1.1's untouched `document.go`/`Serialize()`: a small multi-object PDF writer (`builder`) generalising `document.go`'s fixed four-object xref to an arbitrary object count, plus `SerializeTextDocument` which emits `Type0`/`Identity-H`/`CIDFontType2`/`FontFile2`/`ToUnicode` PDF structure from plain `EmbeddedFace`/`TextPage`/`TextRun` data (no vendor font types touch `internal/pdf`).
- **`folio.Render(t, f)`** (`folio-go/render.go`) walks the template's three bands, resolves each text element's face via `style.fontFamily` → `fonts` chain → first entry present in `f`, collects the union of runes per face across the WHOLE document, subsets each distinct face exactly once, and hands the result to `internal/pdf.SerializeTextDocument`.
- **Provisional band-origin convention** (AC28): pageHeader at y=0, content below the header band's declared height, pageFooter flush against the bottom margin — good enough to keep this story's fixtures non-overlapping; pinned by `TestProvisionalBandOriginIsPinned`, which fails the day `internal/layout` exists (Story 2.5).
- **The subprocess seam gained a second selector** (`FOLIO_SUBPROCESS_RENDER_FONT=1`) alongside the untouched fontless one, both handled in the same `TestMain` and the same compiled test binary — required for the matrix's Docker/wasm legs, which run only a single pre-built binary with no filesystem access to `testdata/`, hence the test font is `go:embed`'d directly into the test binary (`folio-go/testfont_embed_test.go`, a root-level `_test.go` file, outside AC3's `internal/`-scoped guard).
- **A new `lint` guard, `ScanEmbedFont`** (AC3), was added since none previously existed to assert "no `go:embed` directive names a font file under `folio-go/internal/`" — the property existed only by inspection before this story.
- **`lint/internal/manifest` gained a redistributed-non-code-asset table** (AC25) alongside its existing Go-module table, scanning `folio-go/testdata/fonts/` (and, once it exists, `folio-go/fonts/`) for committed font binaries and requiring a sibling `LICENSE*`/`NOTICE*` pair, failing the build if either is missing.

### File List

**New:**
- `folio-go/fontset.go` — public `FontSet` type
- `folio-go/render.go` — document assembly (`renderDocument`, face resolution, provisional band origin)
- `folio-go/testfont_embed_test.go` — `go:embed` of the test font into the test binary
- `folio-go/internal/fontset/fontset.go` — font seam: validation, subsetting, tag derivation
- `folio-go/internal/fontset/fontset_test.go` — unitsPerEm boundary, repeat/permutation-invariance, AC7b fixture
- `folio-go/internal/pdf/builder.go` — generalised multi-object PDF writer
- `folio-go/internal/pdf/textdoc.go` — `Type0`/`Identity-H` document serializer, content stream, ToUnicode CMap
- `folio-go/internal/pdf/textdoc_test.go` — AC9 two-page/one-`FontFile2` fixture, missing-face/missing-glyph red-proofs
- `folio-go/internal/geom/scale_surface_test.go` — AC18 "no second `ScaleRound`" guard
- `folio-go/testdata/fonts/Roboto-Regular.ttf`, `LICENSE-Roboto.txt`, `NOTICE.md` — the committed test face (AC25, AC26)
- `folio-go/testdata/lint/embed-font/violating.go`, `compliant.go` — `ScanEmbedFont` fixtures
- `folio-go/testdata/lint/embed-font/dirembed/`, `globembed/`, `allembed/`, `quoted.go` — Finding 5's four evading-shape fixtures (finisher pass): a bare directory embed, a glob, the `all:` prefix, and a quoted argument with an internal space
- `lint/internal/rules/embedfont.go`, `embedfont_test.go` — the new AC3 guard
- `fixtures/font-text/expected.json`, `expected.pdf`, `input.folio`, `README.md` — the new golden fixture (AC14)

**Modified:**
- `folio-go/go.mod`, `folio-go/go.sum` — added `github.com/boxesandglue/textshape v0.0.15`
- `folio-go/folio.go` — `Render` signature change, AC14b nil-template error
- `folio-go/gomod_test.go` — AC20/AC21/AC22 allowlist + denylist, replacing the old single-module count
- `folio-go/render_test.go` — second subprocess selector, new font-path tests, `TestMain` update
- `folio-go/fixture_test.go` — `TestRenderMatchesGoldenFixture` repointed to `pdf.Serialize()`; new `TestRenderMatchesFontTextGoldenFixture`
- `folio-go/template_test.go` — `Render` call signature; `TestRenderIgnoresItsArgumentToday` deleted (AC27)
- `folio-go/matrix_test.go` — two-document matrix (`matrixDocument`, `captureFontRender`, `checkCrossTargetResults`)
- `lint/cmd/genmanifest/main.go`, `lint/internal/manifest/manifest.go`, `manifest_test.go` — redistributed-asset table (AC25); finisher pass widened `ResolveAssets` to a repo-wide walk (Finding 9) and corrected the regeneration command (Finding 16)
- `lint/MANIFEST.md` — regenerated (AC23, AC25); re-verified byte-identical after Finding 9's rewrite

**Finisher pass, additionally modified (all 23 findings — see § Finding Resolutions):**
- `folio-go/internal/fontset/fontset_test.go` — Findings 1, 2, 18, 19 (red-proof rewrite, skip→Fatalf, vacuity closures, set-equality)
- `folio-go/internal/fontset/fontset.go` — Findings 4, 21 (D-1.1.b comment, dead `ItalicRad` field)
- `folio-go/gomod_test.go` — Finding 3 (shared `deniedHits` matcher)
- `folio-go/go.mod` — Finding 11 (`// indirect` dropped)
- `folio-go/internal/pdf/textdoc.go` — Findings 4, 20, 22, 23 (D-1.1.b comments ×2, dead `sanitizePostScriptName`, `/ItalicAngle` via `writeInt`, resource-name collision guard)
- `folio-go/internal/pdf/textdoc_test.go` — Finding 6 (comment correction pointing at the new behavioural test)
- `folio-go/render.go` — Finding 17 (`pageDimensions` returns a located error)
- `folio-go/render_test.go` — Findings 6, 12, 13, 17 (new behavioural AC9 test, table-driven metadata sweep, two-site pinning message, `letterSizeTemplateJSON` fixture)
- `folio-go/fixture_test.go` — Finding 14 (`input.folio` sync assertion)
- `folio-go/matrix_test.go` — Findings 8, 15 (`TestTargetRenderHash` widened to both documents, `assertFixturesShareToolchain`)
- `lint/internal/rules/embedfont.go`, `embedfont_test.go` — Finding 5 (directory/glob/quote-aware detection)
- `.github/workflows/matrix.yml` — Finding 8 (compare-render-hashes widened to both documents)
- `fixtures/font-text/README.md` — Finding 14 (wording correction)
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — Finding 10 (`D-1.5.10 (corrected)` appended), D-000.12 recorded

### Change Log

| Date | Change |
|---|---|
| 2026-08-23 | Story created, `ready-for-dev`, baseline `94b453a`. Two conflicts surfaced as `DECISION NEEDED` under D-1.2.6. |
| 2026-08-23 | **Revised after ruling.** Both decisions ruled (D-1.5.7, D-1.5.8); the golden-fixture collision ruled (D-1.5.9); three confirmations logged (D-1.5.10). **Three `(mechanism: binding)` clauses reversed** by this story's compiled probes. Governing-ruling location corrected: working tree, not `HEAD`. |
| 2026-08-23 | **Implemented.** All 22 tasks complete, all 31 ACs verified. `internal/fontset` (validation, subsetting, tag derivation), `internal/pdf`'s `Type0`/`Identity-H` text document builder, `folio.Render(t, f)`, the `lint` `ScanEmbedFont` guard, the redistributed-asset manifest table, and the new `fixtures/font-text/` golden fixture all added. Full four-target matrix run for both documents (all agree, both match golden). `fixtures/minimal-rect/` confirmed byte-unchanged. Status → `review`. |
| 2026-08-23 | **Reviewed.** bmad-code-reviewer logged 23 findings (0 blockers, 9 majors, 10 minors, 4 nits) plus an independent 25-row disposition table disagreeing with the developer on 7 rows. Verdict: substance sound, defects concentrated in proof quality and guard reach. Status stayed `review`. |
| 2026-08-23 | **Finished.** All 23 findings triaged FIX (none DISMISS/DEFER — all in-scope proof-quality/correctness gaps in this story's own artifacts) and fixed, each red-proofed live where a red-proof applies (mutate → observe FAIL → revert → confirm clean). All seven disposition disagreements resolved: `D-1.1.c` row added and preamble corrected (Finding 7), `D-1.1.b`'s disposition corrected to `applied-as-stated` (Finding 4), the tautological red-proof and silent-skip fixture fixed (Findings 1, 2), the denylist red-proof's shared-matcher gap fixed (Finding 3), D-000.9 instances 4/5 closed (Findings 1, 18), and D-1.5.10's false withdrawal rationale corrected in both the story and the decision log (Finding 10, `D-1.5.10 (corrected)` appended; D-000.12 recorded as a new program-wide standing rule). Full four-target matrix re-run for both documents after all fixes: byte-identical to the developer's recorded hashes on all four targets (minimal-rect and font-text unchanged). Status → `done`. |

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer (fresh context, no knowledge of the developer's reasoning)
- **Date:** 2026-08-23
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 0
- **Majors:** 9
- **Minors:** 10
- **Nits:** 4

**What was independently re-measured, not taken on report.** Every headline claim in the Delivery
Log was reproduced from scratch:

| Claim | Independent result |
|---|---|
| `fixtures/minimal-rect/` byte-unchanged (D-1.5.9, priority 1) | **CONFIRMED.** `git cat-file -p 94b453a:…` vs working tree: `cmp` identical for `expected.pdf` (547 B, `0f925e1b…4f7c`), `expected.json` and `README.md`. `git diff 94b453a -- fixtures/` empty. |
| Four-target matrix, both documents (AC15, D-000.4 override) | **CONFIRMED, run by the reviewer.** Docker buildx `linux/amd64` + `linux/arm64`, Node v24.16.0 `js/wasm`, native `darwin/arm64`. `minimal-rect` → `0f925e1b…4f7c` / 547 B ×4; `font-text` → `dcd453a1…facbdf` / 22299 B ×4. Exit 0. Matches the reported hashes exactly. |
| `folio-go` gate, `-count=1` | **CONFIRMED.** True exit 0, **6 packages** (incl. new `internal/fontset`), **zero SKIPs** on `darwin/arm64`. |
| `lint` gate, `-count=1` | **CONFIRMED.** Exit 0, 4 packages. |
| AC10b vacuity guard actually fires | **CONFIRMED BY CONSTRUCTION.** Repointed the font subprocess selector at `pdf.Serialize()`: local two-process test **RED** (`render_test.go:637`) and matrix leg **RED** (`matrix_test.go:507`) before any hash was compared. Reverted by hand. |
| AC7b kills the output-numbering reading | **CONFIRMED BY CONSTRUCTION.** Replaced `deriveTag(plan.GlyphSet())` with a `{0…n-1}` set: `TestDeriveTagUsesClosureSetNotOutputNumbering` **RED** with the collided tag `"BQJKPS"`. Reverted by hand. |
| AC27 pinning test deleted, not weakened | **CONFIRMED.** `TestRenderIgnoresItsArgumentToday`'s body is gone from `template_test.go`, replaced by a tombstone comment; zero references remain repo-wide. |

**All reviewer mutations were reverted by hand** (`cp` from byte-exact scratch copies; `git checkout --`
never used, since the story's work is uncommitted). `cmp` clean on `fontset.go`, `render_test.go`,
`render.go`; subagent-side reverts verified against their own backups and against `git status --porcelain`.

**Verdict in one line.** The story's *substance* is sound — every central property genuinely holds and
I proved the load-bearing ones by construction. The defects are concentrated in **proof quality and
guard reach**: two shipped artifacts that are *named* as red-proofs cannot fail, one binding fixture
has a silent-skip path, a binding D-1.1.b code-comment mandate is unimplemented, and the new
`ScanEmbedFont` guard misses the exact evasion shape Story 2.2 will present. Nothing here blocks
merge on correctness; all nine Majors are about assertions that do not assert what they claim.

---

## Ruling disposition table — reviewer's INDEPENDENT reading (D-000.10)

Built from the code first, then compared. **7 disagreements**, marked ⚠️. The story header enumerates
**25** ruling IDs; the developer's table has **24** rows.

| # | Ruling | Developer | Reviewer (independent) | Agree? |
|---|---|---|---|---|
| 1 | D-1.5.1 | applied-as-stated | **applied-with-a-gap** — allowlist is exact-set, by name, both directions printed, and red-proves via the *shared* `diffModuleSets`. The denylist exists, but its named red-proof duplicates the matcher inline; neutering the production matcher leaves **both** tests green. | ⚠️ **DISAGREE** (Major 3) |
| 2 | D-1.5.2 | applied-as-stated | applied-as-stated — 16–16384 enforced in the constructor; `0` and `20000` both located errors naming font *and* value, no panic. `readUnitsPerEm` deliberately bypasses `(*ot.Face).Upem()`'s silent 1000-substitution, which is what makes the `0` fixture observable at all — a genuinely good catch. | ✅ |
| 3 | D-1.5.3 | applied-as-stated | applied-as-stated — MIT; `go list -m all` returns **exactly 2** modules; no `"time"` anywhere in the module. | ✅ |
| 4 | D-1.5.4 *(corrected)* | applied-as-stated | applied-as-stated — equality-only head assertion, no clock read, one subset per font per document. (The *reason* given for one withdrawal is falsified — see row 10.) | ✅ |
| 5 | D-1.5.5 | applied-as-stated | applied-as-stated — `Render(t *Template, f FontSet)`, `f` in final target position; Story 1.4's pinning test **deleted with a tombstone**, verified in the diff. | ✅ |
| 6 | D-1.5.6 | not-reached (package half) / AC26 applied | same reading — no production faces; test face at `testdata/fonts/`; `absences.go`'s DW-2 tripwire on `folio-go/fonts/` still armed. | ✅ |
| 7 | **D-1.5.7** | applied-as-stated | applied-as-stated — **verified by reading the whole path: there is no sort anywhere upstream of the permutation test.** `Subset` ranges the caller's `[]rune` twice; `seen`/`oldGIDForRune` are indexed, never ranged; `AddGlyphs(gids...)` receives caller order. Repeat-invariance is N=32 in one process and would redden if the dependency's internal sort vanished. | ✅ |
| 8 | **D-1.5.8** | applied-as-stated | **applied-with-a-gap** — `deriveTag(plan.GlyphSet())`, sorted ascending, source numbering: correct, and AC7b's fixture **genuinely reddens** (proved). But that fixture carries a conditional `t.Skipf`, and the shipped `TestDeriveTagRedProofAgainstOutputNumbering` **stayed green under the same mutation**. | ⚠️ **DISAGREE** (Majors 1, 2) |
| 9 | **D-1.5.9** | applied-as-stated | applied-as-stated — fixture byte-identical on all three files; matrix reproduced the recorded hash on all four targets; `Render(nil, …)` located error; golden test repointed to `pdf.Serialize()`. | ✅ |
| 10 | **D-1.5.10** | applied-as-stated | **applied-with-a-gap** — no `*ot.Font`/`*ot.Head`/`*subset.Plan` crosses the boundary (only primitive accessors); equality-only; FontFile2 guard live on both surfaces. **But the withdrawal rationale is false:** the test ships at `folio-go/render_test.go` (module root), where D-1.3.1's `time` ban does not apply. | ⚠️ **DISAGREE** (Minor 1) |
| 11 | D-000.4 | applied-as-stated | applied-as-stated — **re-run by the reviewer**, both documents, all four legs, exact hash match. | ✅ |
| 12 | D-000.5 | not-reached | not-reached. | ✅ |
| 13 | D-000.6 | not-reached | not-reached (the falsified premise at row 10 sits in the decision log, not a SPEC companion). | ✅ |
| 14 | D-000.9 | applied-as-stated, 10 instances | **applied-with-a-gap** — instance **4** is itself a D-000.9 failure: it prints the same thing whether production is right or wrong. Instance **5**'s stated closure relies on intra-package test *ordering*, which Go does not guarantee as a dependency. | ⚠️ **DISAGREE** (Majors 1, Minor 9) |
| 15 | D-000.10 | applied-as-stated ("this is that table") | **not-fully-applied** — 24 rows for **25** enumerated rulings; **`D-1.1.c` has no row at all**, and the preamble mis-states the count as "Twenty-four entries". | ⚠️ **DISAGREE** (Major 7) |
| 16 | D-000.11 | applied-as-stated | applied-as-stated — every gate re-run with `-count=1` by the reviewer; exit 0. | ✅ |
| 17 | D-1.1.a | applied-as-stated | applied-as-stated — `TestGoModPinsToolchainExactly` untouched and passing; AC24's non-issue correctly reasoned. | ✅ |
| 18 | **D-1.1.b** | applied-with-a-different-mechanism | **applied-with-a-gap.** The over-report is **not** a departure — D-1.1.b already exempts Identity-H glyph ids and the subset tag *verbatim*, so flagging it was the right instinct but the wrong conclusion. The real gap is the next sentence: *"This must be said in a code comment or a later agent will 'unify' it."* **No such comment exists at any of the three sites.** `numbers.go` confinement itself is intact (still exactly `appendLength`/`appendInt`/`appendIntPadded`). | ⚠️ **DISAGREE** (Major 4) |
| 19 | D-1.2.6 | applied-as-stated, "resolved by construction" | **AGREE — verified true, not a silent choice.** `ScanMapRange` flags only an `*ast.RangeStmt` whose subject types to a map. `Subset` ranges only slices at both unsorted sites; `deriveTag` uses `slices.Sorted(maps.Keys(…))`, the escape-hatch idiom the guard names verbatim; `fnv64aOverGIDs` ranges a slice. Both rulings hold simultaneously with neither bent. | ✅ |
| 20 | D-1.3.1 | applied-as-stated | applied-as-stated — no `"time"` import in any `_test.go` under `internal/`; no clock read anywhere in the head-timestamp test. (The ban was never actually in play — row 10.) | ✅ |
| 21 | D-1.3.4 | applied-as-stated | applied-as-stated. | ✅ |
| 22 | D-1.3.6 | applied-as-stated | applied-as-stated — `ScanEmbedFont` lives in `lint/internal/rules` beside every other guard. | ✅ |
| 23 | D-1.3.9 | applied-as-stated | applied-as-stated — `MANIFEST.md` regenerates byte-identically; `textshape` is the only `shipped`/`Serves: folio-go` row; the asset table is genuinely **scanned** (copyright extracted from `NOTICE.md`, red-proved twice). Reach caveat logged as Major 9. | ✅ |
| 24 | D-1.4.15 | applied-as-stated | applied-as-stated — mechanism tags present in code comments; behavioural tests added alongside structural ones. (Two *named* proofs are structural duplicates rather than behavioural — Majors 1 and 3 — which cuts against the ruling's spirit without breaching its letter.) | ✅ |
| 25 | **D-1.1.c** | **— NO ROW —** | **reached-and-applied.** `Render`'s signature changed exactly as D-1.1.c anticipates ("designed to change every story"); it is cited in `folio.go`'s doc comment and is load-bearing in D-1.5.9's own reasoning. It required a disposition line. | ⚠️ **DISAGREE — omitted** (Major 7) |

### All seven disagreements, resolved (finisher pass)

The developer's table above (**§ Ruling disposition table (D-000.10)**) has been corrected in
place and now carries all 25 rows. Each disagreement the reviewer raised is closed as follows —
none arbitrated by picking a side; each resolved by fixing the artifact or the record:

| # | Ruling | Reviewer's disagreement | Resolution |
|---|---|---|---|
| 1 | D-1.5.1 | Denylist red-proof exercised a copy of the matcher, not production (Major 3) | **FIX.** `deniedHits` extracted as the one matcher both the production check and its red-proof call. Re-red-proofed live: neutering `deniedHits` reddens the red-proof while the production check stays green (today's graph has no denied module) — exactly the blind spot the reviewer proved, now closed. |
| 8 | D-1.5.8 | AC7b's fixture has a silent `t.Skipf`; its named red-proof is a tautology (Majors 1, 2) | **FIX (both).** The skip is now `t.Fatalf` (a broken precondition is a human decision, never a pass). The red-proof (`TestDeriveTagRedProofAgainstOutputNumbering`) now exercises real production subsets and reproduces the reviewer's own collision (`"BQJKPS"`) under the output-numbering mutation, red-proofed live. |
| 10 | D-1.5.10 | The "never near the current time" withdrawal's stated reason (D-1.3.1's `internal/` scope) is false — the test ships at the module root | **FIX (record only, no behaviour change).** `D-1.5.10 (corrected)` appended to the decision log; this story's AC11a and § Decisions surfaced text corrected to rest the withdrawal on the face-age argument alone. The shipped equality-only assertion was already correct and needed no code change. |
| 14 | D-000.9 | Instance 4 prints the same thing whether production is right or wrong; instance 5's closure depends on unguaranteed test ordering (Major 1, Minor 9) | **FIX (both, same code as rows 1/8 above and Finding 18).** Instance 4 is closed by the same red-proof rewrite as row 8. Instance 5 is closed by adding `len(Program) == 0` vacuity checks directly inside `TestRepeatInvarianceWithinOneProcess` and `TestPermutationInvariance`, so each test's vacuity closure no longer depends on `TestSubsetContainsRequestedGlyphs` running first. |
| 15 | D-000.10 | 24 rows for 25 enumerated rulings; `D-1.1.c` has no row; preamble mis-states the count | **FIX.** Same fix as row 25 below — the `D-1.1.c` row now exists, and the preamble reads "Twenty-five entries." |
| 18 | D-1.1.b | The developer's `applied-with-a-different-mechanism` is an over-report; the real gap is the missing code comment | **FIX + table correction.** The three sites (`appendHex4`, `appendHexCIDString`, `deriveTag`) now quote D-1.1.b's mandated sentence verbatim. The table row is corrected to `applied-as-stated`, matching the reviewer's own reasoning. |
| 25 | D-1.1.c | No disposition row at all | **FIX.** Row added (see the corrected table above), disposition `reached-and-applied`. |

---

## Findings

### Finding 1: `TestDeriveTagRedProofAgainstOutputNumbering` is a tautology — Task 12's mandated red-proof cannot fail
- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/internal/fontset/fontset_test.go:280–311`
- **Observation**: The test defines a local `rejectedTag(numGlyphs int)` and then asserts only
  `rejectedTag(8) != rejectedTag(8)` (twice, under two variable names). Both comparisons pass the
  **same argument** to a pure function, so both are unfailable. The test **never calls `deriveTag`**,
  never calls `Subset`, and never compares two *different* glyph sets. Proved by construction: I
  replaced `deriveTag(plan.GlyphSet())` with the rejected `{0…n-1}` derivation in
  `fontset.go:258`; `TestDeriveTagUsesClosureSetNotOutputNumbering` went **RED** while this test
  **stayed PASS**.
- **Impact**: Task 12 requires "Red-proof it against the output-numbering derivation, which is the
  reading it exists to kill", and D-000.9 sweep instance 4 cites this test as a coverage witness.
  The artifact shipped under that mandate has zero coupling to production code — it would stay green
  if `deriveTag` were replaced by `return "AAAAAA", nil`. Its own doc comment claims it "confirms it
  collides for same-size sets, so the fixture above is demonstrably discriminating", which is not
  what the code does. **The property itself is fine** — Finding-1 is about the proof, not the tag.
- **Suggested Resolution**: Make it compute the rejected derivation over **two real subsets of
  different content but equal `NumGlyphs`** (reuse `setA`/`setB`) and assert the rejected tags are
  **equal** while the real `subA.Tag`/`subB.Tag` **differ** — i.e. exercise the same `Subset` output
  the production assertion uses. Alternatively delete it and rely on the retained fixture, which is
  genuinely discriminating.
- **Related AC**: AC7b, AC30 (D-000.9 instance 4); Task 12

### Finding 2: AC7b's retained binding fixture has a silent-skip path
- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/internal/fontset/fontset_test.go:262–264`
- **Observation**: `if subA.NumGlyphs != subB.NumGlyphs { t.Skipf(...) }` — if the two closures ever
  differ in size, the fixture **skips** rather than fails. It does **not** skip today (measured: the
  test runs and passes, no SKIP in `-v` output), so the fixture is non-vacuous right now.
- **Impact**: AC7b is `(mechanism: binding, D-1.5.8)` and requires a *retained* fixture that is "not
  vacuous". A skip is a green build. A `textshape` upgrade that changes GSUB/composite closure for
  `ABC` or `DEF` would silently retire the only permanent guard against the catastrophic
  output-numbering reading — the precise failure mode D-1.5.8 exists to prevent, arriving through
  the door the story left open.
- **Suggested Resolution**: Turn the skip into `t.Fatalf` — the fixture's precondition failing means
  the *fixture* is broken and must be re-chosen, which is a human decision, not a pass. Better still,
  pick the two sets so equal size is guaranteed, or assert over `Plan.GlyphSet()` sizes directly.
- **Related AC**: AC7b

### Finding 3: `TestModuleGraphDenylistRedProof` red-proves a copy of the matcher, not the matcher
- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/gomod_test.go` (`TestModuleGraphDenylistRedProof`, and the production
  matcher in `TestModuleGraphDenylistsKnownPDFWriters`)
- **Observation**: The red-proof inlines a duplicate of the nested match loop over a local `scratch`
  slice; it shares only the `knownPDFWriterModulePrefixes` variable with production. Proved by
  construction: neutering the production matcher to `if false && (...)` — a no-op denylist — left
  **both tests green, exit 0**. (Contrast `TestDiffModuleSetsCatchesExtraAndMissing`, which calls the
  *shared* `diffModuleSets` and correctly went red under the equivalent mutation.)
- **Impact**: AC21 makes the denylist a `(mechanism: binding)` independent backstop. Its red-proof
  cannot observe the thing it certifies, so any divergence in the production matcher ships silently.
  Note also the denylist derives from the same `goListModuleNames` observation as the allowlist, so
  on today's exact-2-module graph it can never fire — its real value is catching *allowlist widening*,
  which makes the integrity of its own matcher the only thing standing behind it.
- **Suggested Resolution**: Extract `matchesDeniedModule(path string) bool` (or
  `deniedHits([]string) []string`) and have both the production test and the red-proof call it —
  exactly the shape `diffModuleSets` already models correctly in the same file.
- **Related AC**: AC21

### Finding 4: D-1.1.b's mandated "byte encoding, not a number" code comment is absent at all three sites
- **Severity**: Major
- **Category**: Convention
- **Location**: `folio-go/internal/pdf/textdoc.go:304` (`appendHex4`), `:379–386`
  (`appendHexCIDString`), and `folio-go/internal/fontset/fontset.go:299–351`
  (`deriveTag` / `fnv64aOverGIDs`)
- **Observation**: D-1.1.b reads, verbatim: *"Glyph ids under `Identity-H` take neither route — they
  are a big-endian hex pair inside a string literal, i.e. a byte encoding like `/ID`, not a number.
  Same for the six-letter subset tag. **This must be said in a code comment or a later agent will
  'unify' it.**"* The comment at `textdoc.go:379–386` explains the CID lookup and the missing-glyph
  error but **never says this is a byte encoding rather than a number**, nor that it must not be
  unified with `numbers.go`. `appendHex4` carries **no comment at all**. `deriveTag`'s comment
  discusses sorting and the rejected readings but never makes the byte-encoding statement for the
  **six-letter subset tag**, which D-1.1.b names in the same breath.
- **Impact**: This is the one part of D-1.1.b that is a *mechanism*, not a classification, and it is
  the part that was unimplemented. The ruling names the exact consequence: a later agent sees three
  number-ish emitters, two confined to `numbers.go` and one not, and "unifies" them — routing 16-bit
  CIDs through `appendInt` and corrupting every content stream. The developer's disposition line
  correctly quotes the first half of the ruling and stops one sentence short of the obligation.
  (`numbers.go` confinement itself is verified intact: still exactly `appendLength`, `appendInt`,
  `appendIntPadded`, no `strconv`/`fmt` in the hex path.)
- **Suggested Resolution**: Add the ruling's own sentence, quoted (D-1.5.x convention: *"Never
  paraphrase these into code comments — quote them"*), to all three sites.
- **Related AC**: D-1.1.b (dispositioned row 18); AD-3

### Finding 5: `ScanEmbedFont` misses directory and glob embeds — the exact shape Story 2.2 will present
- **Severity**: Major
- **Category**: Security / Convention (guard reach)
- **Location**: `lint/internal/rules/embedfont.go:100–120` (line/suffix matching)
- **Observation**: The guard is textual: it suffix-matches a font extension against the directive
  line, never resolving the directive's arguments against the filesystem and never stripping quoting.
  Measured, one variant at a time against the fixture:

  | directive | detected |
  |---|---|
  | `//go:embed shipped-face.ttf` | yes |
  | `//go:embed all:fonts/shipped-face.ttf` | yes |
  | `//go:embed fonts/*` | **NO** |
  | `//go:embed fonts` | **NO** |
  | `//go:embed all:fonts` | **NO** |
  | `//go:embed "shipped face.ttf"` (legal Go) | **NO** |

- **Impact**: AD-8's Rule is *"no package under `internal/` embeds font data"*. A directory embed
  embeds font data and is invisible to this guard. Story 2.2 ships production faces into a `fonts/`
  directory — `//go:embed fonts` / `//go:embed all:fonts` is precisely the idiom that arrives next,
  and it is the one shape the guard cannot see. AC3's literal wording ("naming a font file") is
  arguably satisfied; AD-8's Rule is not. The coverage witness (`FilesParsed > 0`) is real and
  red-proves correctly — this finding is about *what* it detects, not whether it runs.
- **Suggested Resolution**: Resolve each directive argument relative to the file's directory and walk
  it (a directory argument matches if it *contains* a font by extension); strip surrounding quotes and
  the `all:` prefix before matching. Add the four evading variants above as retained fixtures.
- **Related AC**: AC3; AD-8

### Finding 6: AC9's "one subsetting call" half is asserted by code review, not by a test
- **Severity**: Major
- **Category**: AC Conformance
- **Location**: `folio-go/internal/pdf/textdoc_test.go:47–80`
  (`TestSerializeTextDocumentEmbedsOneFontFileForTwoPages`); production site
  `folio-go/render.go:150–200` (`renderDocument`)
- **Observation**: AC9 requires the property asserted **behaviourally**: "a two-page document using
  one face embeds exactly one `FontFile2`". The shipped test builds a synthetic `fakeFace` and two
  hand-constructed `TextPage` values and calls `SerializeTextDocument` directly — it never reaches
  `renderDocument`, never calls `Font.Subset`, and never exercises the union-collection loop. The
  test's own doc comment concedes this: the "one subsetting call" half is *"asserted at folio.go's
  `renderDocument` **via code review**"*. Separately, `renderDocument` can only ever emit **one**
  page (`return pdf.SerializeTextDocument([]pdf.TextPage{page}, embedded)`), and
  `assertWellFormedPDF` hard-fails on `pageCount != 1`, so a two-page document is not constructible
  through the public path at all today.
- **Impact**: The property *does* hold by construction (the loop over `faceNames` calls
  `font.Subset` once per face over a document-wide rune union), so this is not a live defect. But
  "via code review" is not an assertion, and AC9 is `(mechanism: binding, D-1.5.4)`. A future refactor
  that moves subsetting inside the per-run loop — the exact regression AC9 names — reddens nothing.
- **Suggested Resolution**: Count `Font.Subset` invocations directly, e.g. a package-level test in
  `folio` that renders a document whose bands reference the same face from several elements and
  asserts the resulting PDF contains the face program exactly once **and** that the subset tag is the
  document-wide one (a per-element subset would yield a different, smaller closure). Record honestly
  in the AC that the two-page half is deferred until pagination exists.
- **Related AC**: AC9

### Finding 7: The D-000.10 disposition table omits `D-1.1.c` — 24 rows for 25 enumerated rulings
- **Severity**: Major
- **Category**: AC Conformance
- **Location**: story file, § *Ruling disposition table (D-000.10)*; enumeration at the story header
  (*"Rulings that govern"*)
- **Observation**: The header enumerates **25** distinct ruling IDs; the table has **24** rows and its
  preamble states *"Twenty-four entries"*. `D-1.1.c` is absent. It is not an idle citation: it governs
  the public API changing shape every story, it is quoted in `folio.go`'s `Render` doc comment
  (*"Story 1.6 gives `Render` its data parameter and Story 1.7 its `io.Writer` form (D-1.1.c)"*), and
  it is load-bearing in D-1.5.9's own argument (*"it existed only from binding the fixture to the
  public API, which **D-1.1.c** designed to change every story"*).
- **Impact**: AC31 requires the developer to disposition **every** ruling ID the story enumerates, and
  D-000.10's stated purpose is that an unexamined ruling is the failure mode. A missing row is
  strictly worse than an unread `applied-as-stated`: nothing signals that the ruling was even
  considered. The table also instructs *"Do not rebuild this list — it is the complete enumeration"*,
  so the undercount propagates to anyone who trusts it.
- **Suggested Resolution**: Add the `D-1.1.c` row (reviewer's reading: **reached-and-applied** — the
  signature changed exactly as the ruling anticipates) and correct the preamble to "Twenty-five
  entries".
- **Related AC**: AC31, AC30 (D-000.10)

### Finding 8: CI has no per-target cross-target coverage of the font document
- **Severity**: Major
- **Category**: Tests
- **Location**: `folio-go/matrix_test.go:630–671` (`TestTargetRenderHash`);
  `.github/workflows/matrix.yml:108` and the four per-target jobs
- **Observation**: CI's per-target jobs invoke `go test -tags=matrix -run TestTargetRenderHash`.
  That test captures **`captureRender`** (the fontless selector) only and compares against
  `fixtures/minimal-rect/expected.json`. It never calls `captureFontRender` and never loads
  `fixtures/font-text/expected.json`. Both documents are covered **only** by
  `TestCrossTargetByteIdentity`, which is `//go:build matrix` and run locally/on demand.
  The developer disclosed this deferral.
- **Impact**: AC15 and AC10a are satisfied **for this story** — I re-ran the four-target matrix over
  both documents myself. But the durable consequence is that from Story 1.6 onward, CI regression-tests
  cross-target byte-identity for the *rectangle* and not for the *font*. Font subsetting is, in the
  story's own words, "the first property in the project whose determinism depends on a third-party
  module's internal ordering" — a `textshape` upgrade that broke cross-target identity would pass CI.
  The property with the most fragile determinism has the weakest standing guard.
- **Suggested Resolution**: Either parameterise `TestTargetRenderHash` over `matrixDocuments` (publishing
  one hash artifact per target **per document**, with the compare job diffing both sets), or record the
  gap explicitly as a Story 1.6 obligation so it is not silently inherited. Note the compare jobs'
  existing "fail on missing or malformed artifact" logic would need the same per-document widening.
- **Related AC**: AC15, AC10a; D-000.4

### Finding 9: Redistributed-asset accounting is a flat scan of two hard-coded directories
- **Severity**: Major
- **Category**: Security / Convention
- **Location**: `lint/internal/manifest/manifest.go:94–100` (`assetDirs`), `:124–201` (`ResolveAssets`)
- **Observation**: `assetDirs` is a fixed two-entry list (`folio-go/testdata/fonts`,
  `folio-go/fonts`); `ResolveAssets` `os.Stat`s each and uses a **non-recursive** `os.ReadDir` that
  skips subdirectories. Proved by construction: a font copied to `folio-go/assets/Stray.ttf` with no
  licence beside it left `cd lint && go test -count=1 ./...` **green, exit 0**; a font at
  `folio-go/testdata/fonts/extra/Second.ttf` likewise **green**. Control: the same file at
  `folio-go/testdata/fonts/Second.ttf` correctly went **red**.
- **Impact**: AC25 reads *"**Any** committed font binary — fixture or shipped — travels with its
  licence text and copyright line"*, and this story is explicitly the one that "sets the pattern for
  Story 2.2". The accounting cannot see a font outside two flat directories, so the AD-26 obligation
  is enforced by directory convention rather than by the guard. (`absences.go`'s DW-2 tripwire does
  independently cover the `folio-go/fonts/` path until 2.2 lands, which narrows but does not close
  the gap.) The parts that *are* covered are genuinely good: the copyright line is really extracted
  from `NOTICE.md` (red-proved by rewriting and by deleting it), the missing-`LICENSE*` case
  fails closed, and `MANIFEST.md` regenerates byte-identically.
- **Suggested Resolution**: Replace the fixed list with a repo-wide `filepath.WalkDir` filtered by
  font extension (excluding `.git`), keeping the per-directory `LICENSE*`/`NOTICE*` requirement. Add
  the two evading placements above as retained fixtures.
- **Related AC**: AC25; AD-26

### Finding 10: AC11a's withdrawal rationale is falsified by where the test actually ships
- **Severity**: Minor
- **Category**: Convention (record accuracy)
- **Location**: `folio-go/render_test.go:680` (`TestEmbeddedHeadTimestampsEqualSourceExactly`); story
  AC11a and D-1.5.10 item 2
- **Observation**: D-1.5.10 withdrew D-1.5.4's *"never a value near the current time"* clause as
  **unbuildable**, reasoning: *"asserting 'not near now' requires reading the clock, and **D-1.3.1
  bans `time` in `_test.go` under `internal/`, which is where this test must live**."* The test does
  **not** live under `internal/` — it ships at `folio-go/render_test.go`, package `folio` at the
  module root, where D-1.3.1's ban does not apply. The clause was therefore buildable.
- **Impact**: A `(mechanism: binding)` clause was withdrawn on a premise the implementation
  contradicts. The **conclusion still stands** on the independent, correct argument (the face-age
  trap: Roboto's `modified` ≈ 3.57e9 vs ≈ 3.87e9 for the present would make "not near now" a
  meaningless green), and the shipped equality assertion is strictly stronger than the withdrawn
  clause. But the record as written will be re-derived wrongly: a future story that needs a
  clock-reading assertion at the module root will believe it is banned. Verified separately that
  the *implemented* constraint holds — no `"time"` import in any `_test.go` under `internal/`.
- **Suggested Resolution**: Amend D-1.5.10's item 2 to rest solely on the face-age argument, and
  correct "which is where this test must live" — the test lives at the module root by choice, and
  that choice is fine.
- **Related AC**: AC11a; D-1.5.10, D-1.3.1

### Finding 11: `go.mod` marks `textshape` `// indirect` though it is a direct dependency
- **Severity**: Minor
- **Category**: Convention
- **Location**: `folio-go/go.mod:20`
- **Observation**: `require github.com/boxesandglue/textshape v0.0.15 // indirect`, while
  `folio-go/internal/fontset/fontset.go:23–24` imports `…/textshape/ot` and `…/textshape/subset`
  directly. Confirmed on a throwaway copy: `go mod tidy` (exit 0) rewrites the line, stripping
  `// indirect`.
- **Impact**: Cosmetic in resolution terms — `go list -m all` and the allowlist are unaffected — but
  the file is not tidy, and `gomod_test.go` asserts over this file. Any future `go mod tidy -diff`
  or dirty-tree check in CI flags it, and the marker misdescribes the module graph the story just
  spent two ACs pinning.
- **Suggested Resolution**: Drop the `// indirect` comment (or run `go mod tidy`).
- **Related AC**: AC20, AC23

### Finding 12: The forbidden-metadata sweep never runs on the font document
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/render_test.go:512–527` (`TestRenderHasNoCreationOrModDate`)
- **Observation**: The test renders `minimalTemplateJSON` with an **empty** `FontSet` — a fontless
  document — and checks `/CreationDate`, `/ModDate`, `/Info`, `/Producer`, `/Creator`, `/Metadata`,
  `/Filter`. The font document is never swept. I measured the font document directly: **all seven
  keys are absent today**, so this is a coverage gap, not a live defect.
- **Impact**: F-7 explicitly flagged `/Filter` as the compression risk on the font path
  (*"if this story compresses the `FontFile2`, that assertion goes red"*) — but as wired, it cannot
  go red, because it never sees that path. The document with the large binary stream, the one where
  a future `/Filter` or `/Producer` would actually be tempting, has no metadata guard.
- **Suggested Resolution**: Run the same forbidden-key loop over `Render(parseFontTestTemplate(t),
  testFontSet())`, or table-drive it over both documents.
- **Related AC**: AC5, AC10

### Finding 13: `TestProvisionalBandOriginIsPinned` points at the wrong file
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/render_test.go:616–625`; convention implemented at
  `folio-go/render.go:78–92` (`collectTextRuns`)
- **Observation**: The pinning test's failure message names *"the PROVISIONAL origin convention in
  `internal/pdf/textdoc.go` (`buildTextContentStream`)"*. `textdoc.go:317–328` does hold a provisional
  convention (the Y-flip into PDF user space), but AC28's actual subject — the **band** origin
  (pageHeader at 0, content below the header band, pageFooter at `usableHeight - footerHeight`) — is
  implemented in `folio-go/render.go:78–92`, which the message never mentions.
- **Impact**: The test correctly fires the day `folio-go/internal/layout` exists (verified: it stats
  that path relative to the package directory). But the human it summons is pointed at one of the two
  files that must change, so the band-origin half can be missed — the "provisional quietly becomes
  the real convention" outcome AC28 exists to prevent, arriving through a stale pointer.
- **Suggested Resolution**: Name both sites in the failure message.
- **Related AC**: AC28; AD-24

### Finding 14: `fixtures/font-text/input.folio` is hand-synced with no assertion
- **Severity**: Minor
- **Category**: Tests
- **Location**: `fixtures/font-text/input.folio`; `folio-go/render_test.go:46–83`
  (`fontTestTemplateJSON`); `fixtures/font-text/README.md`
- **Observation**: The README states the fixture input is *"identical to `folio-go/fontTestTemplateJSON`,
  **kept in sync by hand** — the fixture is not read at test time"*. I verified they match **byte-for-byte
  today** (839 bytes each, exact match). Nothing asserts it.
- **Impact**: AD-21 makes the fixture the normative record of what produced the hash. If the constant
  is edited and `input.folio` is not, the fixture documents a document that does not produce its own
  recorded hash — a silently lying artifact, and the sort of drift that is only discovered when
  someone tries to reproduce the golden by hand.
- **Suggested Resolution**: Add a one-line test asserting
  `os.ReadFile("../fixtures/font-text/input.folio") == fontTestTemplateJSON` (it already reads
  out-of-module files under `-count=1`, so this costs nothing new), or have the fixture test parse
  `input.folio` rather than the constant.
- **Related AC**: AC14; AD-21

### Finding 15: The font fixture's recorded toolchain is never verified in the matrix
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/matrix_test.go:494`
- **Observation**: `assertToolchainWitness(t, target, binPath, fixtures[matrixDocuments[0].label].GoToolchain)`
  — the gate always uses the **first** document's fixture (`minimal-rect`). `fixtures/font-text/expected.json`'s
  `goToolchain` field is loaded but never compared against anything.
- **Impact**: Both files currently record `go1.26.0`, so nothing diverges. But the font fixture's
  toolchain field is decorative: if it were re-recorded under a different toolchain, the matrix would
  compare its hash while silently accepting a mismatched provenance claim — the exact AD-22/C6
  property the toolchain gate exists to enforce, unenforced for one of the two documents.
- **Suggested Resolution**: Assert every document's fixture records the same `goToolchain`, or run the
  witness check per document.
- **Related AC**: AC14, AC15

### Finding 16: The documented manifest-regeneration command does not work
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `lint/cmd/genmanifest/main.go:3`; `lint/internal/manifest/manifest_test.go:35,52,56`
- **Observation**: All four sites instruct the reader to run `go run ./lint/cmd/genmanifest` **from the
  repo root**. That fails: `go: cannot find main module, but found .git/config in
  /Users/panitw/Projects/folio` (exit 1) — the repo root has no `go.mod`. The working command is
  `cd lint && go run ./cmd/genmanifest`.
- **Impact**: The instruction is printed in `TestManifestUpToDate`'s **failure message**, so the one
  moment a developer needs it is the moment it misleads them.
- **Suggested Resolution**: Correct the command in all four places.
- **Related AC**: AC23, AC25

### Finding 17: `pageDimensions` silently falls back to A4 for any unrecognised named size
- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-go/render.go:32–48`
- **Observation**: The `else` branch returns A4 for **any** named size that is not `"A4"`, with only a
  comment noting the choice. A document declaring `"size": "Letter"` renders as A4 with no error.
- **Impact**: Silent wrong output on a valid-looking input, in a module whose entire premise is
  predictable bytes. Out of this story's declared scope, but it is a new code path this story
  introduced, and the failure is invisible rather than loud.
- **Suggested Resolution**: Return a located error for an unrecognised named size (consistent with how
  `resolveFace` treats an unresolvable chain), or restrict the template schema to the sizes actually
  supported.
- **Related AC**: AC1 (new code introduced by this story)

### Finding 18: D-000.9 instance 5's stated closure depends on test ordering
- **Severity**: Minor
- **Category**: Tests
- **Location**: story § *D-000.9 — ten new-check instances*, item 5;
  `folio-go/internal/fontset/fontset_test.go:121–228`
- **Observation**: Item 5 acknowledges that a `Subset` returning `nil, nil` would make the invariance
  tests' `bytes.Equal(first.Program, got.Program)` trivially true, and states the risk is *"closed by
  `TestSubsetContainsRequestedGlyphs` running first in the same package"*. Go does not guarantee this
  as a dependency — it is declaration order today, and `-run` filtering, `-shuffle`, or a future file
  split breaks it. Neither invariance test asserts non-empty output itself.
- **Impact**: The vacuity closure D-000.9 requires is real in practice but is not a property of either
  test; it is a property of how they happen to be scheduled. Running
  `go test -run TestPermutationInvariance` alone removes it entirely.
- **Suggested Resolution**: Add `if len(base.Program) == 0 { t.Fatal(...) }` (and the same for
  `first.Program`) inside each invariance test, so each closes its own vacuity.
- **Related AC**: AC8, AC30

### Finding 19: `TestPermutationInvariance` checks the length of `shuffled`, not that it is the same set
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/fontset/fontset_test.go:201–204`
- **Observation**: `shuffled` is a hand-written 26-rune literal guarded only by
  `len(shuffled) != len(ascending)`. I verified it **is** a genuine permutation of A–Z today (all 26
  letters, each once).
- **Impact**: If someone edits the literal and introduces a duplicate plus an omission, the length
  check still passes and the test silently compares subsets of **different** rune sets — which would
  then legitimately produce different bytes and read as a determinism failure, or (worse) coincidentally
  match and read as a pass. The test would be measuring something other than permutation invariance.
- **Suggested Resolution**: Assert set equality (`slices.Sorted` of both, compared) rather than length.
- **Related AC**: AC8

### Finding 20: `sanitizePostScriptName` is an identity function
- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/internal/pdf/textdoc.go:256–258`
- **Observation**: `func sanitizePostScriptName(name string) string { return name }` — it does nothing,
  and its only call site already wraps it in `pdfNameEscape`.
- **Impact**: A named no-op reads as a implemented safeguard. A later reader assumes PostScript-name
  sanitisation exists.
- **Suggested Resolution**: Delete it, or implement it and drop the redundant `pdfNameEscape` wrap.
- **Related AC**: AC5

### Finding 21: `Metrics.ItalicRad` is declared but never set or read
- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/internal/fontset/fontset.go:152`
- **Observation**: The field carries the comment *"always 0 in this story; no italic faces are tested"*
  and is never assigned in `Metrics()` nor read by `internal/pdf`.
- **Impact**: Dead surface on a struct that crosses the package's trust boundary.
- **Suggested Resolution**: Remove it until an italic face exists.
- **Related AC**: AC17a

### Finding 22: `/ItalicAngle` is emitted as a bare literal rather than via the ruled route
- **Severity**: Nit
- **Category**: Convention
- **Location**: `folio-go/internal/pdf/textdoc.go:202`
- **Observation**: `b.write([]byte("] /ItalicAngle 0 /Ascent "))` embeds the value in a string literal.
  D-1.1.b's "neither" case names `/ItalicAngle` explicitly: *"convert to thousandths, decimal route."*
- **Impact**: None today — the value is a constant `0` and no float is involved. But `/ItalicAngle` is
  called out by name in the ruling, so the first italic face will meet a site that does not follow it.
- **Suggested Resolution**: Route it through `b.writeInt(0)` now (or `appendLength` on thousandths when
  it becomes non-zero), so the shape is already correct when the value stops being constant.
- **Related AC**: D-1.1.b; AD-3

### Finding 23: `pdfNameEscape` can collapse two distinct face names into one resource name
- **Severity**: Nit
- **Category**: Correctness
- **Location**: `folio-go/internal/pdf/textdoc.go:241–254`
- **Observation**: The function **drops** every character outside `[A-Za-z0-9_-]`. `"Roboto Regular"`
  and `"RobotoRegular"` both become `RobotoRegular`, and any name reducing to empty becomes `"F"`.
  Resource-dictionary keys and content-stream `Tf` operands are both built from it.
- **Impact**: Two faces in one document whose names differ only by dropped characters would collide on
  a single PDF resource name, silently rendering one face's runs in the other's glyphs. Unreachable in
  this story (one face), but the collision is silent when it arrives.
- **Suggested Resolution**: Use proper PDF `#xx` name escaping, or detect a collision across
  `faceNames` and return a located error.
- **Related AC**: AC5

---

## Finding Resolutions (finisher pass)

All 23 findings are **FIX**. None are DISMISS or DEFER: every one is a narrow, in-scope
correctness or proof-quality gap in an artifact this story itself created, matching the reviewer's
own verdict ("all nine Majors are about assertions that do not assert what they claim") — none
imply an architecture change, none reach outside this story's files, and D-000.9's standing rule
("a proof that cannot fail is indistinguishable from a proof that was never run") means a vacuous
test finding is close to a default-FIX category, not a judgment call. Every red-proofable fix below
was red-proofed live (mutate → observe FAIL → revert → confirm PASS/clean diff); the four that are
pure documentation corrections (7, 10, 16 partially) are marked as such.

| # | Title | Decision | Rationale | Files/functions changed |
|---|---|---|---|---|
| 1 | `TestDeriveTagRedProofAgainstOutputNumbering` tautology | **FIX** | Task 12 mandated a red-proof against the output-numbering reading; the shipped test never called `deriveTag`. Rewritten to exercise real production subsets; red-proofed live (collided on `"BQJKPS"`, matching the reviewer's own reproduction). | `folio-go/internal/fontset/fontset_test.go` |
| 2 | AC7b fixture's silent `t.Skipf` | **FIX** | A skip is a green build; it would silently retire the only permanent guard against the catastrophic reading. Changed to `t.Fatalf`. | `folio-go/internal/fontset/fontset_test.go` |
| 3 | Denylist red-proof exercises a copy of the matcher | **FIX** | The red-proof must observe the thing it certifies. Extracted `deniedHits` as the one matcher both production and its red-proof call; red-proofed live (neutering `deniedHits` reddens the red-proof, leaves production green on today's clean graph — exactly the blind spot proved). | `folio-go/gomod_test.go` |
| 4 | D-1.1.b's mandated code comment absent at three sites | **FIX** | The ruling's own next sentence makes this a mechanism, not a classification: *"This must be said in a code comment or a later agent will 'unify' it."* Quoted verbatim at all three sites. Also corrects the developer's over-report of D-1.1.b's disposition (see § All seven disagreements, resolved). | `folio-go/internal/pdf/textdoc.go` (`appendHex4`, `appendHexCIDString`), `folio-go/internal/fontset/fontset.go` (`deriveTag`) |
| 5 | `ScanEmbedFont` misses directory/glob embeds | **FIX** | AD-8's Rule is "no package under `internal/` embeds font data," not "no directive naming a font file" — a directory or glob embeds font data and was invisible. Story 2.2 ships exactly this shape. Rewrote to resolve directory/glob arguments against the filesystem, strip quotes and the `all:` prefix; added four evading fixtures; red-proofed live against the old suffix-only scanner (all four missed, confirmed). | `lint/internal/rules/embedfont.go`, `embedfont_test.go`, four new fixtures under `folio-go/testdata/lint/embed-font/` |
| 6 | AC9's "one subsetting call" asserted via code review | **FIX** | "Via code review" is not an assertion; the test never reached `renderDocument`. Added a real behavioural test through the public `Render` path on a two-element, two-band document, asserting exactly one `/FontFile2` AND that the embedded tag is the document-wide union tag, not either element's own. Red-proofed live against a per-element-subset mutation. | `folio-go/render_test.go`, `folio-go/internal/pdf/textdoc_test.go` (comment correction) |
| 7 | Disposition table omits `D-1.1.c`; miscounted | **FIX** | AC31 requires every enumerated ruling dispositioned; a missing row is worse than an unexamined `applied-as-stated`. Added the row, corrected the preamble to "Twenty-five entries." | story file: § Ruling disposition table (D-000.10) |
| 8 | CI has no per-target font-document coverage | **FIX** | The property with the most fragile determinism (a third-party module's internal ordering) had the weakest standing guard — CI's per-target jobs covered only the fontless document. `TestTargetRenderHash` now loops over both `matrixDocuments`, publishing one hash per target per document; `.github/workflows/matrix.yml`'s compare job widened to check both. Verified the shell logic locally against fabricated hash files (pass and mismatch cases). | `folio-go/matrix_test.go`, `.github/workflows/matrix.yml` |
| 9 | Redistributed-asset accounting is a flat two-directory scan | **FIX** | AC25 reads "**any** committed font binary," and this story explicitly sets the pattern for Story 2.2. Replaced the fixed list with a repo-wide `filepath.WalkDir` (excluding `.git` and lint's own guard-fixture tree), keeping the per-directory LICENSE\*/NOTICE\* requirement. Red-proofed live: a stray font at `folio-go/assets/Stray.ttf` and at `folio-go/testdata/fonts/extra/Second.ttf` both now fail the build (previously silent). | `lint/internal/manifest/manifest.go` |
| 10 | AC11a's withdrawal rationale falsified by where the test ships | **FIX (record only)** | A `(mechanism: binding)` clause was withdrawn on a premise the implementation contradicts (the test ships at the module root, not under `internal/`). The conclusion still holds on the independent face-age argument. No code change — story AC11a/§ Decisions surfaced text corrected, `D-1.5.10 (corrected)` appended to the decision log. | story file (AC11a, § D-1.5.10), `folio-mvp-decision-log.md` (new appended entry) |
| 11 | `go.mod` marks `textshape` `// indirect` though direct | **FIX** | Cosmetic in resolution terms but the file is not tidy and `gomod_test.go` asserts over it. Dropped the comment; confirmed `go mod tidy` now makes no further change. | `folio-go/go.mod` |
| 12 | Forbidden-metadata sweep never runs on the font document | **FIX** | F-7 flagged `/Filter` as the compression risk specifically on the font path, but the test swept only the fontless document. Made table-driven over both documents. Red-proofed live: injecting `/Filter` into the `FontFile2` object stream is now caught (previously would have passed). | `folio-go/render_test.go` |
| 13 | `TestProvisionalBandOriginIsPinned` points at the wrong file | **FIX** | The failure message named only `internal/pdf/textdoc.go`, missing AC28's actual subject (`render.go`'s band-origin math) — a human summoned by this test could miss the more important half. Message now names both sites. | `folio-go/render_test.go` |
| 14 | `fixtures/font-text/input.folio` hand-synced, unasserted | **FIX** | AD-21 makes the fixture the normative record of what produced the hash; nothing asserted the hand-sync claim. Added a byte-equality assertion against the JSON constant; red-proofed live (appending a byte to `input.folio` reddens the test). | `folio-go/fixture_test.go`, `fixtures/font-text/README.md` (wording correction) |
| 15 | Font fixture's recorded toolchain never verified in the matrix | **FIX** | The gate always compared against `matrixDocuments[0]`'s (minimal-rect's) fixture; font-text's `goToolchain` was loaded but never compared. Added `assertFixturesShareToolchain`, called before the existing per-binary witness check in both `TestCrossTargetByteIdentity` and `TestTargetRenderHash`; red-proofed via a dedicated unit test using a fake reporter (a real `t.Run` subtest cannot be used here — a failing subtest always marks its parent failed too). | `folio-go/matrix_test.go` |
| 16 | Documented manifest-regeneration command does not work | **FIX** | The command is printed in the test's own failure message — the one moment a developer needs it, it misleads. Corrected to `cd lint && go run ./cmd/genmanifest` at all three code sites; verified the corrected command actually runs (exit 0, no diff). | `lint/cmd/genmanifest/main.go`, `lint/internal/manifest/manifest_test.go` |
| 17 | `pageDimensions` silently falls back to A4 | **FIX** | Silent wrong output on a valid-looking, schema-legal input (`"size": "Letter"` is in `internal/template`'s own closed set) in a module whose entire premise is predictable bytes. Now returns a located error naming the unrecognised size. Added a retained fixture; red-proofed live against the old silent-fallback behaviour. | `folio-go/render.go`, `folio-go/render_test.go` |
| 18 | D-000.9 instance 5's closure depends on test ordering | **FIX** | The stated vacuity closure ("closed by `TestSubsetContainsRequestedGlyphs` running first") is a scheduling accident, not a guaranteed dependency — `-run` filtering breaks it. Added `len(Program) == 0` checks directly inside both invariance tests. | `folio-go/internal/fontset/fontset_test.go` |
| 19 | `TestPermutationInvariance` checks length, not set equality | **FIX** | A length-only check would still pass if the literal were edited to a duplicate-plus-omission, silently comparing different rune sets. Changed to a sorted-slice set-equality assertion. | `folio-go/internal/fontset/fontset_test.go` |
| 20 | `sanitizePostScriptName` is an identity function | **FIX** | A named no-op reads as an implemented safeguard to a later reader. Deleted; its one call site already wraps the result in `pdfNameEscape`. | `folio-go/internal/pdf/textdoc.go` |
| 21 | `Metrics.ItalicRad` declared but never set or read | **FIX** | Dead surface on a struct that crosses the package's trust boundary. Removed. | `folio-go/internal/fontset/fontset.go` |
| 22 | `/ItalicAngle` emitted as a bare literal | **FIX** | D-1.1.b names `/ItalicAngle` explicitly as a "convert to thousandths, decimal route" value; a bare string literal is the shape the ruling exists to prevent, even though the value is a harmless constant `0` today. Routed through `writeInt`. | `folio-go/internal/pdf/textdoc.go` |
| 23 | `pdfNameEscape` can collapse two distinct face names | **FIX** | Two faces whose names differ only by dropped characters would silently collide on one PDF resource name. Added a collision check across the document's whole face set, failing loudly before any object is written. | `folio-go/internal/pdf/textdoc.go` |

---

## AC-by-AC disposition

Every AC was explicitly considered.

| AC | Verdict |
|---|---|
| AC1 | satisfied — `Render(t *Template, f FontSet) ([]byte, error)`, final target position |
| AC2 | satisfied — `FontSet` in `folio-go/fontset.go`; `internal/fontset` neither embeds nor queries the host |
| AC3 | satisfied in letter; guard reach deficient — **Finding 5** |
| AC4 | satisfied — no host-font/FS-enumeration route on the render path; `resolveFace` fails closed |
| AC5 | satisfied — `Type0`/`Identity-H`/`CIDFontType2`/`FontFile2`/`ToUnicode` all asserted present |
| AC6 | satisfied — six chars, A–Z asserted, sorted before hashing |
| AC7 | satisfied — hashes `Plan.GlyphSet()`, source numbering, sorted ascending |
| AC7a | satisfied — both rejected readings recorded |
| AC7b | property satisfied and reviewer-red-proved; fixture has a silent-skip path — **Finding 2**; its named red-proof is a tautology — **Finding 1** |
| AC8 | satisfied — no sort upstream; N=32 in one process; genuine permutation. Vacuity closure is ordering-dependent — **Finding 18**; set-equality unasserted — **Finding 19** |
| AC8a | satisfied — recorded, and honoured by construction |
| AC9 | property holds by construction; behavioural assertion is at the wrong layer — **Finding 6** |
| AC10 | satisfied — two-process identity, guard-gated |
| AC10a | satisfied — both selectors, matrix runs both (reviewer-reproduced) |
| AC10b | satisfied — **red-proved live by the reviewer on both surfaces** |
| AC11 | satisfied — two-run comparison correctly absent |
| AC11a | satisfied — equality alone, no clock, no "or zero". Withdrawal rationale falsified — **Finding 10** |
| AC12 | satisfied — recorded as evidence, no `GOMODCACHE` assertion |
| AC13 | satisfied — `qpdf --check` + PyMuPDF text round-trip recorded in the Delivery Log |
| AC14 | satisfied — `fixtures/font-text/` matches Story 1.2's shape; `sha256` a 64-char lower-case string |
| AC14a | satisfied — **byte-identity independently confirmed**; test repointed to `pdf.Serialize()` |
| AC14b | satisfied — `Render(nil, …)` returns a located error |
| AC15 | satisfied for this story (reviewer re-ran all four legs, both documents); CI coverage gap — **Finding 8** |
| AC16 | satisfied — 16–16384, `0` included, located error naming font and value |
| AC17 | satisfied — validation in the constructor, validated value the only reachable one |
| AC17a | satisfied — no `*ot.Font`/`*ot.Head`/`*subset.Plan` crosses the boundary |
| AC18 | satisfied — exported set is exactly `{ScaleRound}`, with a real vacuity guard |
| AC19 | satisfied — all three fixtures retained; `0` case returns an error, not a panic |
| AC20 | satisfied — exact-set allowlist by name, both directions printed, red-proved |
| AC21 | present but its red-proof does not guard production — **Finding 3** |
| AC22 | satisfied — graph assertion in `gomod_test.go`, licence half in `lint/` |
| AC23 | satisfied — `textshape` is the first `Serves: folio-go`/`shipped` row; manifest regenerates byte-identically |
| AC24 | satisfied — recorded, and `TestGoModPinsToolchainExactly` untouched |
| AC25 | satisfied for the shipped asset; accounting reach deficient — **Finding 9** |
| AC26 | satisfied — no production faces; one Latin test face only |
| AC27 | satisfied — **deleted with a tombstone, verified in the diff** |
| AC28 | satisfied — pinned by a test that genuinely fires; pointer stale — **Finding 13** |
| AC29 | satisfied — every gate re-run by the reviewer with `-count=1` |
| AC30 | mostly satisfied — instance 4 is itself a D-000.9 failure (**Finding 1**), instance 5 ordering-dependent (**Finding 18**) |
| AC31 | **not satisfied** — `D-1.1.c` has no disposition row — **Finding 7** |

**Finisher update.** The table above is the reviewer's original record and is left unedited. Every
AC it flagged as deficient is now fixed (see § Finding Resolutions above for the file-by-file
detail): AC3 (Finding 5), AC7b (Findings 1, 2), AC8 (Findings 18, 19), AC9 (Finding 6), AC11a
(Finding 10, record-only), AC15 (Finding 8), AC21 (Finding 3), AC25 (Finding 9), AC28 (Finding 13),
AC30 (Findings 1, 18), AC31 (Finding 7 — now satisfied, the disposition table carries all 25 rows).

### Process checks

- **Scope fence (AC26)** — held. No production faces; `folio-go/fonts/` does not exist and
  `absences.go`'s DW-2 tripwire remains armed for Story 2.2.
- **`sprint-status.yaml`** — only the status value changed (`backlog` → `review`). Clean.
- **Story opener** — 346 words (within 150–350), past-tense-ready, scope fence stated.
- **D-1.2.6 near-conflict** — the developer's "resolved by construction, not by choosing a side"
  claim is **verified true**, not a silent choice. See disposition row 19.
