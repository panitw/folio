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
