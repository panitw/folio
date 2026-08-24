# Epic 1 — Boundary Gate

**Run:** 2026-08-24 · **HEAD:** `614bc99` (Story 1.8) · **Run by:** the orchestrator, directly
**Cadence:** D-000.4 (per-epic). Stories **1.3, 1.4, 1.6, 1.7** deferred their cross-target matrix
runs to this gate; **1.2, 1.5, 1.8** ran theirs in-story under the override.

## Verdict — **PASS**

All five gate items pass. Two findings are recorded, **neither blocking**: a public-API type alias
(ruled, fix owed before anything else touches `render.go`) and a latent licence-manifest coverage
gap with **zero current instances**. Three deferrals rest on prose triggers rather than mechanical
ones and are named below. `epic-1: done` may be written.

*Method note, stated because it nearly produced a false finding:* two early attempts at item 1
reported **"DID NOT FIRE"** for a tripwire that in fact fires — a false negative from grepping
**wrapper-truncated output**. Writing to a file and reading the file back gave the truth. This is
the third false signal from that wrapper in this run (after a `Go build: Success` printed over a
non-zero exit, and a wrong hash through a pipe). **Per D-000.12, file-then-read is now the default
for gate measurement, not the fallback after a surprising result** — a surprising result is exactly
when the instrument is least likely to be questioned.

---

## Item 1 — Every absence tripwire fires · **PASS**

The lead's top-priority item. `ScanAbsences` was the **seventh** instance of the dominant defect
class (it passed with an empty check list), and Epic 1 added four more rows on the same pattern.
**Proved by construction, not inspection:** each named path was created, the red observed, the path
removed.

| Path created | Exit | Rule named |
|---|---|---|
| `folio-designer` | 1 | `absence-designer-project` |
| `folio-go/fonts` | 1 | `absence-fonts-dir` |
| `folio-go/internal/expr` | 1 | `absence-expr-package` |
| `folio-go/internal/diag` | 1 | `absence-diag-package` |
| `folio-go/cmd` | 1 | `absence-cmd-dir` |

Baseline with no paths present: **green**. Tree restored clean after every cycle. The coverage
witness asserts `ChecksEvaluated == len(absenceChecks)` (not merely non-zero), closing the gap a
future `continue` would otherwise leave open.

## Item 2 — Deferred-work sweep · **PASS, with three prose-only triggers named**

Ten entries, **all with named owners that exist as planned stories**. One closed (DW-1, by Story
1.3). **Six carry mechanical tripwires — every one of them proved firing in item 1:**

| Entry | Owner | Trigger |
|---|---|---|
| DW-2 | Story 5.1 | `folio-designer` tripwire |
| DW-5 | Story 3.2 | `internal/expr` tripwire |
| DW-6 | Story 3.6 | `internal/diag` tripwire |
| DW-8 | Story 3.2 | `internal/expr` tripwire (shared) |
| DW-9 | Story 2.2 | `folio-go/fonts` tripwire |
| DW-10 | Story 3.7 | `folio-go/cmd` tripwire |

**Three rest on prose alone — the fragile ones:**

- **DW-3** — publishing the licence manifest as a release artifact. Owner: **Epic 4 close**.
- **DW-4** — **nobody owns cutting `folio-go/v0.1.0`. Owner: the project owner, at Epic 4 planning.**
  This is the only entry whose owner is a person rather than a story, and DW-3 depends on it.
- **DW-7** — footer evaluation sameness with `{{sum(...)}}`. Owner: **Story 4.5, by name**. The lead
  flagged this at ruling time as *"the weakest of the three"*; no package tripwire is possible.

No entry was found silently satisfied, and no trigger has passed unnoticed.

## Item 3 — Licence manifest, first shipped dependency · **PASS, with a latent gap**

`boxesandglue/textshape v0.0.15 | MIT | Serves: folio-go | **shipped**` — correct. This is the
first time the shipped-vs-build-time labelling means anything; before Story 1.5 every row was
`Serves: lint` / build-time-only. Regenerating produces **zero diff**. `Roboto-Regular.ttf` is
listed and travels with `LICENSE-Roboto.txt` and `NOTICE.md` beside it (AD-26 satisfied).

**Finding — the coverage claim is broader than what is enforced.** The manifest states *"Every
committed font binary — fixture or shipped — appears"*. Three committed `.ttf` files do **not**
appear: `folio-go/testdata/lint/embed-font/{allembed,dirembed,globembed}/fonts/shipped-face.ttf`.
**Measured: 49 bytes each, beginning with the ASCII bytes `not `** — text placeholders, not font
binaries, carrying no third-party content. **So AD-26 has nothing to protect and there is no
licensing exposure.** What is wrong is the *claim*, which is the narrated-vs-artifact drift pattern
again.

**And the known image gap is latent, not live.** The scanner matches font extensions only, so a
committed image would be invisible to AD-26 enforcement. **Measured: there are zero committed image
binaries** — Story 1.8's fixture carries its PNG as base64 inside `input.folio`. The gap opens the
moment a standalone image is committed. **Owner: D-1.8.11**, which already rules the fix must
**invert** the check (every committed file in the declared asset and fixture locations must be
*accounted for*; anything uncovered fails) rather than adding `.png` to a list — the rotting-list
pattern. Both findings above are the same fix.

## Item 4 — Spine drift, and D-000.6's constraint (4) · **PASS — first audit ever run**

**No invariant's `Binds` or `Prevents` line was changed by any Epic 1 amendment.** Measured across
every commit touching the spine:

| Commit | Invariants | Binds/Prevents lines |
|---|---|---|
| `6976a39` initial | 24 | 48 |
| `f2aa8c0` planning | 26 | 52 |
| `048999b` Story 1.1 | 26 | 52 |
| `f9c27b3` Story 1.2 | 26 | 52 |
| `af6b5d5` Story 1.3 | 26 | 52 |
| `614bc99` HEAD | 26 | 52 |

The two additional invariants (24 → 26) landed at **`f2aa8c0`, the planning commit, before Story
1.1** — outside D-000.6's scope entirely. **From the first story commit through HEAD the invariant
set and every binding line are byte-identical.** The Epic 1 amendments touched only Rules, headings
and the Source tree, exactly as D-000.6 constrains.

## Item 5 — Exported-surface audit · **PASS, with one finding ruled**

The public surface is **nine declarations**: `Version`, `LoadTemplate`, `ParseTemplate`, `Render`,
`RenderTo`, `Data`, `Params`, `FontSet`, `Template`. Eight are exactly as D-1.1.c fixed them.
`FontSet` was checked specifically as the surface's least-examined member and is **clean** — a
defined type over a builtin map (`map[string][]byte`), no internal type reachable.

**Finding — `type Template = template.Document` is a type ALIAS** (`folio.go:16`). Note the `=`:
`folio.Template` **is** the internal document type, so every exported field of that struct and **32
exported declarations in `internal/template`** become reachable public API.

Two consequences: the surface is not four functions and four types but four functions plus a struct
graph; and **a caller can construct a `Template` field by field, bypassing `ParseTemplate`
entirely** — so asset-key validation, the version rules, the eight closed sets, the exact-decimal
discipline and `nextId` monotonicity are all sidesteppable. **Four stories of making the load
boundary unfalsifiable have an open side door at the type boundary**, which quietly undermines the
spine's *"nothing in `internal/` panics on malformed input"* convention: malformed input now arrives
by a route that never passed validation.

**Ruled by the lead: option (c), an opaque handle** — `type Template struct { doc *template.Document }`,
construction only through `LoadTemplate`/`ParseTemplate`, no exported fields, **no accessors shipped
pre-emptively**. The lead corrected the option I proposed: a *defined* type without the `=` still
permits composite-literal construction, so it closes the hole **not at all** while reading in review
as though it did.

**Why it needed no owner escalation:** opening the surface later is a **MINOR addition**; closing it
later is a **MAJOR break** (D-1.4.12). The opaque form **preserves** the owner's ability to decide
deliberately; the alias **forecloses** it by accident at the `v0.1.0` tag.

**Migration measured: 10 references, all inside package `folio`** (`render.go` ×8, `template_test.go`
×2), all able to reach an unexported field from the same package. **Zero external callers.**
**Owner: land it before anything else touches `render.go`.**

---

## Heavy-test catch-up — the four-target matrix · **PASS**

Run at HEAD over all three documents. **Every target byte-identical; every document matching its
recorded golden.**

| Document | Hash | Bytes | darwin/arm64 | linux/amd64 | linux/arm64 | js/wasm |
|---|---|---|---|---|---|---|
| minimal-rect (fontless) | `0f925e1b…cb4f7c` | 547 | ✓ | ✓ | ✓ | ✓ |
| font-text (template+font) | `dcd453a1…afacbdf` | 22299 | ✓ | ✓ | ✓ | ✓ |
| image-embed (template+image) | `e5778eb8…be689fc` | 995 | ✓ | ✓ | ✓ | ✓ |

This discharges the deferred matrix debt from Stories **1.3, 1.4, 1.6 and 1.7** in one pass, and
proves byte-identity survives **font subsetting and image embedding** — the two features most likely
to break it. `image-embed` at **995 bytes** against `font-text`'s 22299 is a quiet confirmation that
D-1.8.1's PNG passthrough did what it intended: no compressor invoked, no bloat.

---

## What must be resolved before `epic-1: done`

**Nothing.** The gate passes. Carried forward, with owners:

1. **`Template` opaque-handle migration** — ruled, ~10 in-package edits, **before anything else
   touches `render.go`**.
2. **Licence-manifest inversion (D-1.8.11)** — latent, zero current instances; closes both the
   image-invisibility gap and the overbroad coverage claim.
3. **DW-3, DW-4, DW-7** — prose-only triggers. **DW-4 is an owner decision at Epic 4 planning** and
   DW-3 depends on it.
4. **Alpha/transparent PNGs are rejected** (D-1.8.1) — a scope call, purely additive later, but a
   transparent PNG is the commonest form of a company logo and Epic 4's golden report carries one.
   **For the owner, while Epic 4 is still being designed.**
