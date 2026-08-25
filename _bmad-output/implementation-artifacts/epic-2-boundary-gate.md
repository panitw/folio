# Epic 2 — Boundary Gate (accumulating; not yet run)

**Status: NOT YET RUN.** This file is a **living, append-only accumulator** for findings that must
reach the owner at the Epic 2 boundary (D-000.3: "the per-epic report is the owner's main
observation point"), started at Story 2.2 because it is the first story in this epic to produce
owner-visible findings before the epic itself closes. It is **not** a verdict — unlike
`epic-1-boundary-gate.md` (run once, at that epic's close, over the whole epic), this file
collects items story by story so none is lost, and the actual gate (matrix run, deferred-work
sweep, spine-drift audit, exported-surface audit — mirroring Epic 1's five items) still happens
once, at Epic 2's close, over everything landed by then.

## Items carried forward from Story 2.2

### 1. The `ja` (Japanese) glyph-form limitation is a stated, accepted gap — not a coverage bug (AC10, D-2.0.1)

Noto Sans SC is **Pan-CJK**: it carries kana and the shared ideograph set, so Japanese text
**renders** — AC4's coverage-based missing-glyph diagnostic correctly does **not** fire for it, and
this is measured and tested (`folio-go/ac4_coverage_test.go`,
`TestJapaneseTextThroughPanCJKFaceDoesNotFireDiagnostic`). What differs is **glyph form**:
Simplified-Chinese shapes where JP and SC conventions diverge for a subset of shared ideographs.
This is a genuine, accepted limitation — not tofu, not a coverage gap — recorded in
`folio-go/README.md` beside the locale documentation (an integrator choosing `locale: "ja"` will
read the README, not a story file). **The cost of fixing it**: a dedicated Noto Sans JP face would
add several MB, against Story 5.4's itemised "CJK 7.4 MB" (measured: 4.82 MB) and its ~9 MB total
download budget —
material enough that the owner should see it stated plainly, once, here, rather than only inside
a story file.

### 2. The payload figures were measured, and the shipped set now fits (AC11, D-2.2.5)

Story 2.2 is the first story able to weigh the REAL shipped faces. **The figures below replace an
earlier version of this section that was wrong on three axes at once** — MiB relabelled as MB,
**gzip** used where `epics.md:1382` mandates **brotli**, and the ~1.5 MB engine omitted from the
measured side while both budgets include it. It reported a 3.5 MB overrun that was partly an
artefact of its own arithmetic.

It also measured the *variable* faces, which are no longer what ships. Under D-2.2.4 the shipped
set is now **static, Regular-only** instances derived ahead of the build.

Measured on the committed artifacts, `brotli -q 11` (the codec `epics.md:1382` mandates), decimal
MB throughout, coverage witness **4 of 4 inputs compared**:

| Asset | Raw | brotli -q 11 | Ratio |
|---|---|---|---|
| Noto Sans (Latin) | 646,160 B (0.65 MB) | 226,026 B (0.23 MB) | 2.86× |
| Noto Sans Thai | 47,788 B (0.05 MB) | 24,871 B (0.02 MB) | 1.92× |
| Noto Sans SC (CJK) | 10,595,932 B (10.60 MB) | 4,817,020 B (4.82 MB) | 2.20× |
| **Three faces** | **11,289,880 B (11.29 MB)** | **5,067,917 B (5.07 MB)** | — |
| Thai word-break trie | 2,481,373 B (2.48 MB) | 311,844 B (0.31 MB) | 7.96× |
| **Faces + trie** | **13,771,253 B (13.77 MB)** | **5,379,761 B (5.38 MB)** | — |

**Like-for-like against the budgets** (both of which INCLUDE the ~1.5 MB engine, so the measured
side must too):

| Source | Budget | Measured (fonts + trie + engine) | Verdict |
|---|---|---|---|
| NFR7 (`epics.md:121`) | ~9.00 MB | **6.88 MB** | **inside, 2.12 MB headroom** |
| Story 5.4 AC2 (`epics.md:1405`) | ~9.52 MB | **6.88 MB** | **inside, 2.64 MB headroom** |

**The static switch is what resolves this, not the arithmetic.** The variable build measured
**9.33 MB** for the faces alone at the same codec — genuinely over NFR7's own 7.9 MB font budget,
which is why NFR7's *"variable"* adjective was amended under D-000.6 (see D-2.2.5): the clause was
unimplementable on its own terms. The glyf-over-CFF contrast it exists to defend survives verbatim
and is **reinforced** — glyf static 4.82 MB < glyf variable 8.30 MB < CFF static 10.90 MB.

**Two per-item deltas remain, and are reported rather than filled** (D-000.17, D-2.2.5):

- **Latin**: 5.4 claims 0.4 MB, measured **0.23 MB** — better than claimed.
- **CJK**: 5.4 claims 7.4 MB, measured **4.82 MB** — better than claimed.
- **Thai dictionary**: NFR7 says ~0.1 MB and 5.4 says 0.12 MB; measured **0.31 MB** — roughly
  **2.6–3.1× larger** than either claims. This is the one item that is *worse* than published. It
  is comfortably absorbed by the headroom above, but it is the number that is wrong.

**On the "~24× trie ratio" this section previously asserted**: no planning document states 24×. It
was an inference from NFR7's 0.1 MB, and NFR7 never says that figure is compressed — *"compressed"*
attaches only to the first item in its list. The measured ratio is **7.96×**. Recorded so the
inference is not re-made.

**Nothing was arbitrated or filled**: Story 5.4's ACs were not edited, no reading was picked between
NFR7 and 5.4, and no face, build or document was altered to make a published number come true. The
one `epics.md` edit in this story's commit is D-2.2.5's ruled NFR7 adjective. `prd.md:455` and
`addendum.md:399` retain the superseded costing and are **deliberately not edited** — they sit
outside D-000.6's scope, the glyf-over-CFF reasoning in them is unaffected and reinforced, and
whether to correct them is the owner's call since no implementation depends on it.

**The consequence that makes this more than bookkeeping**: Story 5.4's AC2 shows these figures **to
users**, on the first-run load screen, and requires explaining *why CJK dominates* — a wrong
itemisation shown to a user is a product defect, not a documentation typo. CJK is still 90% of the
font payload (4.82 of 5.07 MB), so that explanation stands.

### 3. DW-9 retired (AC9)

Story 2.2 re-ran DW-9's ceremony re-test (`folio-go/example_test.go`'s executed `Example`,
rewritten README) — the claim held; no `DECISION NEEDED` was raised. See `deferred-work.md`'s DW-9
entry for the full account.

---

*Next story to touch this file: append its own owner-visible findings here, in a new dated
section, rather than overwriting the above. The actual Epic 2 gate (matrix run, deferred-work
sweep, spine-drift audit, exported-surface audit) runs once, at Epic 2's close, and should read
this file's accumulated items alongside its own measurements.*

---

## Story 2.3a — the vendor boundary, audited (appended, not overwriting)

**The project now has an enumerated, executable answer to "what does our one dependency tell us when
it does not know?"** — `folio-go/internal/fontset/vendor-boundary.md`, measured against
`textshape v0.0.15`, with its load-bearing rows derived from the vendor by
`folio-go/internal/fontset/vendorboundary_test.go` rather than narrated beside it.

**Two live defects were found and BOTH were fixed here, not deferred.** Both are the same class —
*folio reads a table that may be absent and receives a substituted default* — with a loud member and
a silent one:

- **`maxp` (loud, D-2.3a.1).** `folio.Render` **panicked** on caller-supplied font bytes whose `maxp`
  was missing or short. `folio.FontSet` is a public `map[string][]byte`, so that was untrusted input
  reaching a panic through the documented entry point, against a convention the spine states
  verbatim. Recorded against Story 1.5, which shipped the call site.
- **`OS/2` (silent, and worse).** A caller-supplied face without `OS/2` **rendered successfully** and
  declared `/CapHeight 928` where the intact face gives `711` — the ascender, not a cap height — with
  every guard in the repository green. Harmless for everything folio ships (all three faces are
  `OS/2` v4 with a real `sCapHeight`), live for the input surface `folio.FontSet` is.

Both are closed by one guard at face ingestion, validating every table folio actually reads, with a
located error naming the face and the table. `name` and `cmap` are deliberately excluded, each for a
ruled reason recorded in the enumeration.

**`/BaseFont` no longer substitutes silently (D-2.3a.2).** A program declaring no `name` record still
gets the FontSet key — `/BaseFont` is Required and the key is true of the face — but the PDF now says
so in a comment. At `431a6a5` a nameless render was byte-identical in length to an intact one, which
is precisely the indistinguishability D-2.2.6 was bought to remove.

**AD-23 is now enforced by type, not by spelling.** The syntactic guard promised "no float arithmetic
under `internal/`" and delivered "no float *identifiers*". A new `lint` rule
(`no-float-typed-value`) matches on the type `go/types` resolves; it reported four expressions in
`internal/fontset/fontset.go` that the syntactic guard could not see, and reports zero now that both
sites read the integer the `hmtx` table actually carries. **The existing guard is unmodified and
still runs: two guards, two mechanisms, one invariant.**

**Nothing was added to this gate.** No new `//go:build matrix` test; the three matrix-tagged files,
`.github/workflows/matrix.yml` and all of `fixtures/shaped-text/` are byte-unchanged from `431a6a5`;
`thai-signoff.json` is still absent. **The gate owes exactly what it owed before: the four-target
matrix legs, and D-2.3.5's Thai sign-off.** No fixture was re-recorded and no golden digest moved —
including `fixtures/shaped-text`'s `5964aad0…92e00f`, which the pending Thai sign-off is bound to.

---

## Story 2.5a — the vertical model corrected; five goldens re-recorded; **the sign-off is HELD** (appended, not overwriting)

### 1. THE ITEM THAT PROTECTS A HUMAN — read this one first

**`fixtures/shaped-text/expected.pdf`'s digest MOVED.** It was
`5964aad0e696010c6e3f34a48d0775af6ae527a6cbe2f5c6319158f43c92e00f` and is now
`6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370`.

D-2.3.5's Thai **reading** sign-off binds to those bytes. **It had not been requested, and this story
did not request it** — that is the whole reason this story exists as its own story rather than as a
paragraph inside 2.6 ([[D-000.41]], which names this story by key). `thai-signoff.json` is still
absent and `TestShapedTextThaiSemanticSignOffIsRecorded` is still deliberately red.

**THE CONDITION FOR RELEASING THE REQUEST, binding on whoever sends it ([[D-000.43]]):**

> The reading sign-off is requested only once **both** hold:
>
> 1. **No scheduled work is known to move `fixtures/shaped-text/expected.pdf` again** (D-000.41), and
> 2. **the gate's four matrix legs have RUN and AGREED on the new digest** above.
>
> Condition 2 exists because a reader must never be asked to certify bytes that a later leg proves
> target-dependent. It composes with condition 1 rather than duplicating it: D-000.41 protects
> against a *future story* moving the artifact, condition 2 against the artifact *not reproducing*.
> Together they protect the scarce human from both futures, at zero cost — which is why the D-000.4
> per-story matrix override was **declined** rather than taken. The concern was sequencing, not
> divergence, and the sequence is what was fixed.

**HAND-OFF, stated here so it is not lost between two story files:** **Story 2.6's creator carries
D-000.41's own obligation** — measure whether 2.6 moves any sign-off-pending golden **before** any
request is sent.

### 2. The Epic 2 gate owes the same FOUR things — no fifth. Each one's fate:

| obligation | what Story 2.5a did to it |
|---|---|
| **the four-target matrix legs** | **Count unchanged; CONTENT changed.** 5 of the 7 registered `matrixDocuments` now compare against new `expected.json` digests, which the harness picks up automatically via `fixtureRelPath`. Nothing to edit, and nothing was edited. |
| **D-2.3.5's Thai READING sign-off** | **Still outstanding, now bound to DIFFERENT BYTES** (item 1). Not discharged, not weakened, not re-scoped. |
| **D-2.4.3's Thai BREAK sign-off** | **Untouched, and MEASURED so.** No break-related test moved: `fixtures/expected-breaks/` and `fixtures/thai-break-corpus/` are byte-unchanged and every break test stayed green. D-000.26 (refined)'s claim that the break judgment binds to the break-opportunity vector, which a baseline shift does not touch, is now **confirmed rather than assumed**. |
| **`three-band-page`'s deferred matrix legs** | **Still deferred to the gate**, still declined under D-000.4's criterion. They will compare the **new** digest `746efcbcfb5be30a06caaaefae25e3eaba1962c3fa47a74da10af6d0885372bf`. |

**No fifth obligation was added.** `declaredEpic2GateObligations` is byte-unchanged and
`TestEpic2GateObligationsMatchTheDeclaredSet` passes with a 7-of-7 witness. The tests this story
added are ordinary tests, not `//go:build matrix` files, and no new document was registered in
`matrixDocuments`.

### 3. Heavy tests: the override was DECLINED, and the legs were NOT run

Stated plainly because the alternative is a reader assuming otherwise: **this story did not run the
four-target matrix legs.** D-000.4's criterion is *a new source of cross-target divergence* — float
arithmetic, a vendor call, a compressor, a new dependency — **not a moved golden**. The new
arithmetic is `geom.ScaleRound` on `int64`, the same integer function in the same package that
`lineAdvance` has used since Story 2.4; the vendor entry points touched are a **subset** of those
already touched; `go list -m all` is still exactly two modules; `no-float-typed-value` reports zero.
`matrix_test.go`'s `requireThreeBandPageUsesAllThreeBands` was **compile-checked** under
`go vet -tags matrix` and its band-boundary thresholds re-verified by hand against the new baselines
— but it was **not executed**.

### 4. A guard whose remedy forbade the correct action, corrected in the same commit

`byte_neutrality_test.go`'s digest tripwire told its reader, verbatim: *"Do not update this literal
to make the test pass."* Story 2.5a **had to** update three of those literals. So a **true** guard
was carrying a remedy that **forbade the correct action** — [[D-000.37]]: a tripwire's failure
message is executed by a human. The message now names both authorised movements (2.3a's
byte-neutrality premise, 2.5a's re-record) and keeps the prohibition, restated as the rule it always
was rather than as one story's premise.

### 5. A golden's digest lives at FOUR sites, and that is now a checked property

Found by measurement, not inspection: the PDF, `expected.json`, `byte_neutrality_test.go`'s
independent second literal, and **two fixture READMEs** (which also quote the byte COUNT — a fifth
thing that goes stale). Per [[D-000.47]] the list is now **declared once**
(`goldenDigestRecord`) and the check **reads it**, including a completeness half that scans for each
digest value and fails on any **undeclared** site. That guard caught a real gap on its first run.

**Owner-visible consequence:** the second-literal discipline now covers **all seven** fixtures, where
before it covered five — `three-band-page` and `wrapped-text` had **no** independent second copy of
their digest at all.

---

## Story 2.6 — pagination; the gate now owes **FIVE** things (appended, not overwriting)

*Dated 2026-08-24. Appended by Story 2.6. Nothing above this line is altered.*

### 1. THE ITEM THAT MATTERS MOST — a recorded golden passed its hash and its semantic acceptance step, and was not a PDF

Story 2.6 recorded `fixtures/multi-page/expected.pdf`. It **passed** its golden hash and it **passed**
the story's whole semantic acceptance step. The **owner opened it and it would not render.**

```
$ qpdf --check fixtures/multi-page/expected.pdf
ERROR: file does not contain any pages
qpdf: errors detected                                            exit=2
```

The cause was one missing byte in the page tree — `/Kids [8 0 R10 0 R]`, with no separator. A PDF
tokenizer reads `R10` as one unknown token, so **neither** kid resolves and the page tree is empty. A
viewer showed *"page 0 of 2"*.

**Every acceptance check was true of the broken file, and each for its own reason.** This is the part
the gate should read, not the one-character fix:

| the check | why it passed on an unrenderable file |
|---|---|
| `/Type /Page` objects == 2 | counted **object definitions**. Both pages were emitted correctly and both were in the xref. What was broken was the **array pointing at them** — a different object. |
| `/Count 2` | read the integer. The integer was right. **`/Count` is a claim about the kids, not a fact derived from them.** |
| per-page header/footer Y, the pinned line→page partition, no negative Y | parsed the **content streams** directly. Content streams do not care whether anything references them, so they read exactly as in a healthy file. Independently confirmed: a length-preserving hand patch of the recorded bytes rendered correctly with **every one of these properties already right**. |
| byte-identity across two processes | says the output is **deterministic**. A deterministically wrong file is byte-identical to itself. |

**The common shape: every check asserted a property of a PART, and the defect was in a REFERENCE
BETWEEN parts.** No quantity of additional per-part assertions would have found it.

**A hash certifies that bytes have not changed since they were recorded. It is silent on whether they
were right when they were recorded.** [[D-000.22]]'s semantic acceptance step is the only thing
standing in that gap, and here it was assembled entirely out of checks a broken file satisfies.

**Remedy shipped**: `folio-go/golden_structural_validity_test.go` — a **hermetic** structural oracle
(no shell-out, so it runs on every target including `js-wasm`) applied to **every** fixture declared in
`goldenDigestRecord`. It parses `/Kids`, groups tokens **in threes** — which is what exposes a
run-together pair, where a substring search for `"8 0 R"` and `"10 0 R"` finds both and reports success
— requires each reference to resolve to a defined `/Type /Page` object, and requires no page object to
be orphaned. Red-proved against the real broken bytes: it reddened for `multi-page` alone while the
other seven stayed green. `internal/pdf`'s `TestPagesTreeKidsAreSeparated` guards the emitter itself.

> **PROPOSED STANDING RULE, for the lead — raised, not assumed.** *No recorded PDF golden's hash may
> be trusted until the artifact has passed a structural validity oracle.* A recording is an act of
> acceptance ([[D-000.44]] makes that true of re-recordings too), and accepting bytes nobody has shown
> to be a **document** is exactly what happened here. Two notes for whoever rules it: the oracle must
> be **hermetic**, because a `qpdf`-dependent check silently degrades to nothing where qpdf is absent;
> and it belongs in the **acceptance step**, not only in the suite, so it gates the **first** recording.

### 2. The Epic 2 gate now owes **FIVE** things, and the fifth was ruled, not assumed

| obligation | fate under Story 2.6 |
|---|---|
| the four-target matrix legs | **unchanged.** The override was DECLINED again; the legs are the gate's. |
| [[D-2.3.5]]'s Thai **reading** sign-off | **untouched, and RELEASED to be requested.** `fixtures/shaped-text/expected.pdf` is byte-unchanged at `6c040ef7…c6c85370`, re-verified at close. |
| [[D-2.4.3]]'s Thai **break** sign-off | **untouched.** `expected_breaks.json` byte-unchanged at `a545e042…9324de`. Its subject is now **pinned in the ordinary suite** — see item 4. |
| `three-band-page`'s deferred matrix legs | **still deferred, still declined** under [[D-000.4]]. |
| **`matrix-document: multi-page` — NEW, the FIFTH** | **ADDED**, sanctioned by [[D-2.6.2]]: it is the **only cross-target artifact for a shipped FR30**, and refusing it would ship FR30 with none. A **one-line diff** to `declaredEpic2GateObligations`, exactly the shape [[D-2.5.1]] created it for — the mechanism paid off one story after it was ruled. Legs **deferred** to the gate; heavy-test override **declined** (pagination is integer comparison and subtraction on `geom.Length`: no float, no vendor call, no compressor, no new dependency). |

**Golden**: `fixtures/multi-page/expected.pdf` = `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b`, 66,525 bytes, 2 pages.

### 3. AD-4's forward guard closes a hole the rank table structurally cannot express

`internal/pdf → internal/layout` was **legal**: the rank rule is *"may import only lower ranks"* and
`layout`(7) is below `pdf`(8), so the edge is **downward**. **Measured, not argued**: with that import
present, lint's `stage-rank` stayed **fully green across all four of its tests** while the new guard
reddened. A rank order expresses *"no backward edges"* and **cannot** express *"no edge to this
particular lower package"* ([[D-000.16]] limitation) — `pdf` legitimately needs `pagemodel` (the
**value**) and must not touch `layout` (the **computation**), and both sit below it.

### 4. `expected_breaks.json` was pinned NOWHERE, not "only behind the matrix tag"

[[D-2.6.4]] ordered an ordinary-suite pin on the premise that the only existing pin was matrix-tagged.
**Measured: the digest literal occurs at ZERO sites in the repository.** The matrix-tagged test
**computes** the digest at run time and compares it against `fixtures/expected-breaks/break-signoff.json`
— **which does not exist**. So the file was unpinned entirely and `folio-go/expected_breaks_digest_test.go`
is its **first** digest pin, not its second. The ordered remedy is unchanged; the premise was weaker
than stated.

**Two different things are explained by that one absent file, and the gate should not conflate them**:
the missing `break-signoff.json` making the matrix test fail is **the pending-sign-off blocker working
as designed** — the gate refusing to pass. The **unpinned fixture** was the genuine hole.

**Discriminating red-proof**: mutating a `gloss` field — which no engine consults — was **invisible** to
the pre-existing ordinary suite while the new pin caught it. Mutating a field the engine *does* read
would have proven nothing, because `internal/text/s4_expected_test.go` would have caught that anyway.

### 5. Carried to Story 2.8, and it is larger than clip machinery

[[D-2.6.5]] withdrew D-2.6.1's *"clip at the window bottom"* disposition: an item that fits in **no**
window is a **located error naming the element**, for a line and an image alike. But `epics.md:938-950`
requires 2.8's box-overflow case to clip **and** diagnose **and** still return PDF bytes — *"clipping
degrades output but does not fail the render."* **That is not expressible on `Render`'s current
surface**, which returns `([]byte, error)`. **Story 2.8 owns a diagnostic-channel design decision on
the public API.**

### 6. Flagged, not fixed

- `sprint-status.yaml` still reads `epic-2: backlog` while `2-1`…`2-6` are done or in review. **The
  epic key is the gate's to flip, not a story's.** Unchanged.
- **DW-14 is now the closest it has ever been to firing.** `fixtures/multi-page/` produces a
  **45-entry** `beginbfchar` section against the spec's cap of 100 — the largest any fixture has
  recorded, from a document that is only **two pages and one face**. A document roughly twice as long
  in one face reaches the cap, and that is an ordinary report. Re-measured, **not** discharged: the fix
  moves the golden hash of every document over the cap and stays the gate's.
- **DW-11 stays open at 2.** Story 2.6 falls in its owner window and owes an answer: **none were found,
  and none were invented.** Pagination cannot reach `internal/text`.

---

## 7. Pre-gate environment readiness — measured, not assumed

Recorded at `ecd0056` by the orchestrator, because a gate that cannot run is discovered at the worst
possible moment. **Only `darwin/arm64` has ever actually been executed**; the other three legs have been
compiled and vetted and never run ([[D-000.55]] is explicit that this is not the same thing).

| leg | dispatch | this host | measured |
|---|---|---|---|
| `darwin/arm64` | native | Go 1.26.0 | **runs** — `TestTargetRenderHash` passes |
| `linux/arm64` | Docker | Docker 29.6.2 | **platform executes** — `docker run --platform linux/arm64 alpine:3 uname -m` → `aarch64` |
| `linux/amd64` | Docker, **emulated** on this host | Docker 29.6.2 | **platform executes** — → `x86_64` |
| `js/wasm` | Node | Node v24.16.0 | toolchain present |

`docker buildx` advertises `linux/arm64, linux/amd64, linux/amd64/v2, linux/riscv64, linux/ppc64le,
linux/s390x, linux/386`. **All four gate targets are runnable on this machine.**

**What this proves and what it does not.** It proves the **environment** cannot fail the gate for want of
a toolchain or a platform. It proves **nothing** about whether the legs agree — that is the gate's whole
job, and `matrix_test.go:236` already names a missing daemon, a missing `linux/arm64` platform and a
missing Node as its failure modes. Stated in both halves per [[D-000.54]]'s guardrail, which required
exactly this distinction for the native leg.

## 8. DW-14 re-measured at `ecd0056` — Story 2.7 did NOT move it closer

**Method matters here, and the first attempt got it wrong.** The cap is **per `beginbfchar` section**, not
per document. An ad-hoc line count over the whole file sums sections and overstates the largest by roughly
double. The authoritative instrument is the regex `wrapped_text_fixture_test.go:377` already uses —
`(\d+)\s+beginbfchar(.*?)endbfchar` — where **the count is the integer the section declares**.

| fixture | sections | sizes |
|---|---|---|
| `multi-page` | 1 | **[45]** ← largest single section |
| `wrapped-text` | 3 | [28, 18, 38] |
| `page-count-20` | 1 | **[32]** ← Story 2.7's new golden |
| `shaped-text` | 3 | [14, 7, 28] |
| `font-text` | 1 | [25] |
| `three-band-page` | 1 | [17] |
| `multi-script-fallback` | 3 | [4, 1, 1] |
| `minimal-rect`, `image-embed` | 0 | — |

**Section 6's statement stands**: `multi-page` at **45** is the largest section any fixture records, and
DW-14 is **not** triggered. **Story 2.7's fixtures did not move it** — `page-count-20` records 32 despite
forcing all ten digit CIDs into the subset, because entries are per unique `(GID, Unicode-context)` pair
and page count does not multiply them.

**A process note against myself, recorded because this run keeps producing the same class.** The first
measurement here was taken with a hand-rolled `grep` rather than the project's own instrument, and
reported `wrapped-text` at 86 as a near-cap emergency. It was **the sum of three sections against a
per-section cap** — a real number about the wrong subject, which is [[D-000.26]] again, and the same
error this run has now recorded against a developer, a creator and the orchestrator twice. **The
established instrument existed and was not reached for first.**

## Story 2.8 — FR44's clip and diagnostic channel; the gate still owes SIX, not seven

*Appended by Story 2.8. Nothing above this line is altered. Baseline for this section: `278520b`.*

### 1. Both halves shipped together (D-2.8.2)

OD-1 (the diagnostic seam) was ruled by the owner before this story left `ready-for-dev`: `Result`
(D-2.8.3), `RenderTo` returns `([]Diagnostic, error)` not `Result` (D-2.8.6). With the seam ruled,
OD-3's "ship both, or neither" resolved to **both**: FR44's clip (source AC1) and the diagnostic
channel (source AC2) landed in this one story. **FR44 now ships in Epic 2** — the earlier framing
("under OD-3 (i), FR44 does not ship in Epic 2") is superseded and does not apply.

### 2. `Render`/`RenderTo` signatures changed — a MEASURED, not a v0.1.0, break

```go
func Render(t *Template, d Data, p Params, f FontSet) (Result, error)
func RenderTo(w io.Writer, t *Template, d Data, p Params, f FontSet) ([]Diagnostic, error)
```

`v0.1.0` has not shipped (D-000.31): one tag (`pre-email-rewrite`), `Version = "0.0.0-dev"`, zero
external consumers by construction. The names `Render`/`RenderTo` are unchanged, so **AD-1's
forbidden-import scan — which locates its target files by matching top-level functions named EXACTLY
those two names — keeps finding them**, confirmed by running lint's own suite (`go test ./...` under
`lint/`, all green) rather than assumed.

### 3. The gate still owes exactly SIX obligations; a seventh remains DECLINED

D-2.8.4, applying D-2.6.2's criterion: `wrapped-text` is already a registered `matrixDocuments` entry
carrying the ONE element FR44's clip acts on (`e4`). Declining a seventh entry does not leave FR44
without a cross-target artifact — the criterion that would compel one. `declaredEpic2GateObligations`
stays **byte-unchanged**, confirmed (`TestEpic2GateObligationsMatchTheDeclaredSet` passes unmodified).

**But note for the gate: `wrapped-text`'s recorded digest MOVES even though the obligation set does
not.** This is a re-recording of an EXISTING registered document, not a new registration — a
different obligation from D-000.54's native-leg requirement (which attaches only to newly registered
documents and is not owed here). Its four legs were re-run against the new bytes in this story's own
development (`TestCrossTargetByteIdentity`, all four targets agree) — the Story 2.4 in-story-override
precedent, applied to a re-recording rather than a first recording.

| digest | before (`278520b`) | after (this story) |
|---|---|---|
| `fixtures/wrapped-text/expected.pdf` | `277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5` (72,738 bytes) | `07c38cf765a39d86376c1a3c78bfb6f0a96f089f19792c9bfeeaa1dc754269d6` (72,790 bytes) |
| `fixtures/shaped-text/expected.pdf` | `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370` | **unchanged** |
| `fixtures/expected-breaks/expected_breaks.json` | `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de` | **unchanged** |
| `fixtures/multi-page/expected.pdf` | `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b` | **unchanged** |
| `page-count-20` | (recorded, Story 2.7) | **unchanged** |

Both pending sign-offs (`shaped-text` Thai reading, `expected-breaks` Thai break) are **untouched** —
`wrapped-text` carries neither, and OD-2's fence (the vertical axis stays out of scope, D-2.8.1)
means the seven `shaped-text` elements that overflow vertically are neither clipped nor diagnosed.
Measured (`TestGoldenDigestAgreesAtEveryDeclaredSite`, all sites agree except the one deliberate
`wrapped-text` movement, which the test names and explains rather than silently accepting).

### 4. `epics.md`'s FR44 coverage claim is now TRUE, where the earlier framing expected it false

The prior boundary-gate note (Story 2.6, item 5) anticipated a diagnostic-channel design decision
still open at the gate. It was ruled (D-2.8.3) and both halves of FR44 shipped in this story. No
`epics.md` amendment is owed for FR44 — the coverage table's claim now holds.

### 5. `folio-format.md` amended under D-000.6

The `x, y, width, height` row (`_bmad-output/specs/spec-folio/folio-format.md`) now states, as an
outcome: a text element's declared `width` is a clip bound (FR44), its declared `height` is not, and
an image's `height` is honoured for AD-24's fit-and-centre and reserved for a future `valign`.

### 6. `internal/diag` was NOT created by this story

D-2.8.3/D-2.8.6 place `Result` and `Diagnostic` directly in package `folio` (module root), not in
`internal/diag` — the owner's ruling names them unqualified. DW-6's tripwire
(`lint/internal/rules/absences.go`'s `absence-diag-package` rule, corrected by D-2.8.4) therefore
stays **green and untouched**; confirmed by running `lint`'s suite and by confirming
`folio-go/internal/diag` does not exist on disk. `deferred-work.md`'s DW-6 entry was corrected (wrong
noun: "a test" → the actual lint rule) but its obligation is not this story's.

### 7. Flagged, not fixed

- **DW-17 (new)**: surfacing a returned `Diagnostic` to a human is a presented-interface obligation
  owed by Stories 3.7, 5.12 and 6.6 (D-2.8.5) — not a Go call-graph property, and no AST guard was
  built (declined on D-000.15/D-000.50: the call-site population is overwhelmingly tests that
  legitimately do not care about diagnostics).
- **Call-site count discrepancy, still not fully reconciled.** The creator measured 72 `Render` call
  expressions (25 files, 5 non-test) at `f651409`, text-stripping comments/strings. The orchestrator
  re-derived 55 at `7f97ef7` with a different stripping method. This developer re-derived a THIRD
  number at this story's own pre-commit working tree (atop `278520b`) using a `go/ast`-based
  `CallExpr` scan (no text stripping needed — the parser already excludes comments/strings and
  declarations structurally): **75** `Render` call expressions across **26** files (4 non-test:
  `render_entry.go` + the three `testdata/swapproof/*/main.go` fixtures — `example_test.go` is
  counted as an ordinary test file by this method, unlike the creator's manual carve-out), and **9**
  `RenderTo` call expressions across **2** files (0 non-test). The three numbers are NOT reconciled
  to each other and are not expected to be: each was taken at a different commit with a different
  method, and this story's own diff legitimately added new call sites (its own new test file). Named
  here per D-2.6.9 so the next reader does not inherit any of the three uncritically.

## Standalone correction (references Story 2.4; not a reopening) — the owner's break hand-check applied

**Executed** (D-000.55): `fixtures/expected-breaks/expected_breaks.json` was corrected against the
owner's 2026-08-24 hand-check (recorded as D-2.4.8..D-2.4.11), and every declared digest site was
updated to match, on target `folio-go` (`go build ./...`, `go test ./...`, and `go vet -tags matrix
./...` / `go test -tags matrix ./...`, all run at this correction's own tree).

### 1. Six corrections, two opposite mechanisms, applied per D-2.4.9's direction rule

`cjk-004` (北京), `cjk-005` (东京都) and `cjk-007` (こんにちは) now carry `"declaredAtomic": true` —
D-2.1.6's declaration channel (`internal/text/opportunity.go:94`'s `atomic` parameter) reaching this
fixture for the first time. `thai-007` (หนังสือพิมพ์), `thai-008` (วันเกิด) and `thai-009` (ที่อยู่)
were re-labelled with a break each (`words`, `expectedBreaks` AND `gloss`, all three sites per
D-000.48). Both forbidden levers stayed forbidden: the shipped wordlist was not touched, and no
heuristic was invented.

### 2. The fixture's founding principle was falsified and is not replaced

Measured against the shipped 62,107-entry `words_th.txt`: หนังสือพิมพ์, วันเกิด and ที่อยู่ are ALL
headwords, so are their split constituents, and so are all five controls the owner kept whole
(thai-003/004/005/006/010). Headword membership never predicted this. The `_README` and
`fixtures/expected-breaks/README.md` were rewritten to state there is no derivable rule — AD-25's
stated impossibility observed from the acceptability side.

### 3. AC14 gained an enumerated, fail-closed-only divergence list

`TestS4ExpectedBreaksMatchTheEngine` (`folio-go/internal/text/s4_expected_test.go`) now asserts set
equality against `s4ExpectedDivergences` (D-2.5.1's shape) for thai-007/008/009, where the engine's
conservative "no break" cannot be closed without an invented heuristic or a wordlist edit. The
direction rule is enforced in code, not just documented: a divergence whose engine-proposed positions
are not a subset of the human label's is rejected as FAIL-OPEN and fails the test. Both this rejection
and the declaration-removal reversion (cjk-004 reverting to `[1]` once `declaredAtomic` is stripped)
were red-proved live and reverted; the working tree was confirmed byte-identical afterward.

### 4. Three states recorded, silence is not ratification

Explicitly corrected: thai-007/008/009, cjk-004/005/007. Explicitly declined: none — cjk-005 was
re-asked with its actual break positions shown and the owner changed the answer. Unremarked, NOT
ratified: thai-003/004/005/006/010.

### 5. `folio-format.md` gained a stated, fail-closed capability limit (D-000.6)

The Line breaking section's existing UAX #14 narrowing now also states, as an outcome: no break falls
inside a dictionary headword, including a lexicalised compound a native reader would accept breaking.

### 6. The digest moved at its one declared, non-matrix-tagged site

| digest | before | after |
|---|---|---|
| `fixtures/expected-breaks/expected_breaks.json` | `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de` | `40ba08f6da1bfadb4178d6f8d420454bee2f4f61ce7a1b3be584b84e7a1cf26c` |
| `fixtures/shaped-text/expected.pdf` | `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370` | **unchanged, confirmed** |
| `fixtures/multi-page/expected.pdf` | `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b` | **unchanged, confirmed** |
| `fixtures/wrapped-text/expected.pdf` | `07c38cf765a39d86376c1a3c78bfb6f0a96f089f19792c9bfeeaa1dc754269d6` | **unchanged, confirmed** |

Updated at its only declared site, `folio-go/expected_breaks_digest_test.go`'s
`expectedBreaksDigest` literal (this artifact is deliberately not folded into `goldenDigestRecord` —
it is not a PDF golden — so `TestGoldenDigestAgreesAtEveryDeclaredSite`'s D-000.47 scan does not cover
it; confirmed by re-running that test, which still passes). `byte_neutrality_test.go`'s
`goldenDigestRemedy` "do not update this literal" message was checked (D-000.37) and still tells the
truth for this change — it governs `goldenDigestRecord`'s PDF-golden literals, none of which moved
here. `TestS4ExpectedBreaksAreLabelsNotEngineOutput` was confirmed, not assumed, to survive unchanged
against the corrected fixture (still passes: 9 of 25 items partitioned).

### 7. D-000.44 discharged by construction

The re-recording owed a semantic acceptance step. The owner's hand-check IS that step — performed
before this correction, on the prior digest, and it is what this correction applies. The sign-off
itself (`break-signoff.json`) is REQUESTED ONLY NOW, against the new digest above, per D-000.41/43 —
it does not exist and `TestExpectedBreaksHumanSignOffIsRecorded` (`-tags matrix`) confirmed still red,
naming the new digest.

### 8. Gate obligation count is unchanged: SIX matrix legs plus TWO sign-offs

This correction touches only one of the two pending sign-offs' subject (the break labels), not the
obligation count. `epic-2: backlog` stays. The all-occurrences baseline (`go test ./... -v`) is
unchanged at 600 PASS / 1 FAIL (`TestCorpusMeetsP6ExerciseFloors`, stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`, identical to the pre-correction baseline);
368/1 at top level. Under `-tags matrix`, the three expected red gates are unchanged in kind: the
break sign-off (now naming the new digest), the pre-existing shaped-text Thai reading sign-off
(untouched, digest confirmed unmoved), and the pre-existing missing-font-sources reproduction gate
(environmental, unrelated to this correction).
