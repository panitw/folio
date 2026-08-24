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
