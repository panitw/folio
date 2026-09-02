---
title: 'Story 16.1a: The local face tier covers the head of the library'
type: 'feature'
created: '2026-09-03'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: '4aa610a'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/specs/spec-fonts/font-catalogue.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-16-decision-log.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the intent contract below governs implementation.*

The font browser can offer an author almost two thousand typeface families, but not the ones they are
most likely to want. Roboto, Open Sans, Inter, Montserrat — thirty-seven of the fifty most popular —
are published on Google's mirror in a form this product deliberately does not accept, because it holds
every weight in one file and picking a weight out of it at the last moment would make the same
template print differently on different machines.

There is a cheaper answer than it first appeared. Those families publish perfectly ordinary
single-weight files from their *own* project releases; only the mirror lacks them. This product already
carries twenty-one typefaces exactly that way, each with its licence, its copyright and a record of the
precise file it came from. This story adds a batch of the most-wanted families the same way, so the
browser can offer them from the machine rather than refusing them.

What it deliberately does not do: it does not convert anything, it does not add a general mechanism for
getting statics, and it does not promise the whole library. It is a curated batch, and the story names
its size, its membership rule and who re-runs it — because a batch nobody owns is the failure this work
exists to prevent.

Done looks like: the families an author is most likely to type are offered and embed with no download
at all, each carrying its own terms.

<intent-contract>

## Intent

**Problem:** D-16.R.2 (owner) funded a bounded batch so the font browser does not refuse the head of the
popularity distribution — measured, **37 of the top 50 families are variable-only** on the `google/fonts`
mirror and therefore unaddable. D-16.R.2a then corrected the mechanism: the batch is **not** a
`tools/fontgen` derivation queue. Measured, `instance_faces.py` drives a hardcoded three-entry
`UPSTREAM` list of **engine** faces, and **none of the 21 designer catalogue faces is derived** — every
`NOTICE.md` says *"NO DERIVATION APPLIES."* This repository already ships **Roboto** and **Inter** as
byte-for-byte upstream statics from `googlefonts/roboto-classic` and `rsms/inter` v4.1.

**Approach:** Grow the **local face tier** by the mechanism this repository has executed 21 times: take
each family's static Regular from **that family's own upstream release**, commit it with its unmodified
`LICENSE*`, write its `NOTICE.md` recording both the upstream archive digest and the committed digest,
and add its `font-catalogue.json` row with `scripts` verified against the binary's own `cmap`.
Derivation stays available for a family whose own upstream publishes no static — now the exception, not
the plan.

## Boundaries & Constraints

**Always:**
- **The provenance regime is copied exactly, not approximated.** Per face: the unmodified upstream
  `LICENSE*`; a `NOTICE.md` recording the upstream project, the release, the **source** SHA-256 and the
  **committed** SHA-256; the `font-catalogue.json` row; and the copyright from the binary's own nameID 0.
  A face that cannot supply all of them is not added.
- **Each face passes the existing gates before it is proposed** — `font-catalogue.test.ts`'s nameID 13
  tie against its declared SPDX id, and its `scripts` declaration checked against its own `cmap` **in
  both directions** (a claimed script it cannot draw renders tofu; an unclaimed one it can draw staples
  a redundant fallback onto every chain that picks it).
- **The licence is one of D-8.5.3's four**, admitted by identifier, and an unclassifiable licence is a
  **build failure**, never a warning (D-8.5.2).
- **The batch's membership rule, size and OWNER are written down in this story.** D-16.5 left the batch
  unbounded and unowned and that is the defect D-16.R.2 exists to close. A batch nobody re-runs is a
  queue, and a queue was the thing the owner was told they were not funding.
- **From the family's own upstream, never from the `google/fonts` mirror.** The mirror is what lacks the
  statics; taking bytes from it is the failure this story routes around.
- **No derivation unless a family genuinely publishes no static**, and then it is `instance_faces.py`'s
  regime — committed output, replayable, hashes asserted in both directions — never a build-time or
  runtime instancing.
- Every recorded measurement carries **command, commit, tree state and working directory** (D-8.4j.8).
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **A face's licence cannot be classified against the four identifiers**, or its `LICENSE*` cannot be
  fetched from the same release as the binary.
- **A face's `cmap` disagrees with its declared `scripts`** in either direction.
- **The batch would be unbounded**, or would land without a named owner and a stated re-run trigger.
- **`maximumCacheAssets` (64) would be exceeded**, or the offline release contract would move. Measure
  the slot count before proposing a size; 44 of 64 were in use at Story 8.6.
- **Any of the 23 golden digests moves.** This story adds assets; it renders nothing differently.

**Never:** bytes from the `google/fonts` mirror · runtime or build-time instancing · a CJK family
(D-8.5.3's scope) · a variable face committed to the catalogue · a hand-copied licence text.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Family publishes a static Regular | `Roboto` at `googlefonts/roboto-classic` | Added: binary, `LICENSE*`, `NOTICE.md` with both digests, catalogue row | — |
| Family publishes VF only, everywhere | no static in any upstream release | **Out of the batch**, recorded with the evidence — not derived silently | Excluded, stated |
| Upstream ships several statics | full weight range published | **Regular only.** Bold and italic are out of scope by a standing Non-goal | — |
| Licence outside the four | e.g. a ShareAlike release | **Build failure**, never a warning | Fails the build |
| `cmap` disagrees with `scripts` | claims `thai`, cannot draw it | **Block** — the declaration is corrected or the face is dropped | Fails the build |
| Family already in the local tier | `Roboto`, `Inter` | No-op; not added twice | — |
| Slot budget | additions approach `maximumCacheAssets` | Measured before proposal; the batch is sized to fit | Halt rather than raise the cap |

</intent-contract>

## Code Map

- `folio-designer/font-catalogue.json` — the 21 rows and their shape (`id`, `directory`, `file`,
  `family`, `licence`, `scripts`).
- `folio-designer/public/fonts/<id>/` — the committed pattern per face: the `.ttf`, `LICENSE-*.txt`,
  `NOTICE.md`. **Read `roboto/NOTICE.md` first** — it is the template, and its *"NO DERIVATION
  APPLIES"* clause is the exact statement most of this batch will make.
- `folio-designer/src/font-catalogue.test.ts` — the gates every added face must pass, including the
  nameID 13 tie at `:355-366` and the `cmap` cross-check.
- `folio-designer/scripts/build-wasm.mjs` — validates and fingerprints the catalogue into
  `src/generated/`; a new row flows through here.
- `folio-designer/src/release-payload.ts:33` — `maximumCacheAssets = 64`, line-anchored and regex-read.
- `tools/fontgen/instance_faces.py` — the derivation regime, for the exception case only. Its
  `UPSTREAM` list is three engine faces; **read its header before extending it.**
- `_bmad-output/specs/spec-fonts/font-catalogue.md` — the selection criteria this batch is admitted by.

## Tasks & Acceptance

**Execution:**
- **Measure and propose the batch, and get it agreed before fetching anything.** Rank the index by
  `popularity`, intersect with families the browser would otherwise filter out, and for each candidate
  establish whether its **own** upstream publishes a static Regular. Report: the proposed list, each
  family's upstream and licence, the total committed bytes, and the resulting `maximumCacheAssets` slot
  count. **The six Thai variable-only families are in the candidate set by name.**
- Per admitted family: fetch the static Regular and its `LICENSE*` from the same upstream release; write
  `NOTICE.md` to the existing template with both digests; add the `font-catalogue.json` row with
  `scripts` derived from the binary's `cmap`; confirm nameID 0 and nameID 13.
- `folio-designer/src/font-catalogue.test.ts` — the population floor moves with the batch, deliberately
  and in one place, so the count is a decision rather than a drift.
- `_bmad-output/specs/spec-fonts/font-catalogue.md` — record the batch, its membership rule, its owner
  and its re-run trigger.
- `deferred-work.md` — register the standing curation obligation with its trigger, since popularity
  moves and this batch is a snapshot of it.

**Acceptance Criteria:**
- Given the batch, when it is proposed, then its membership rule, its size, its owner and its re-run
  trigger are written down, and the slot count is measured rather than assumed.
- Given each added family, when it is committed, then it carries its unmodified upstream licence text, a
  NOTICE recording both the upstream and committed digests, and a copyright read from its own binary.
- Given each added face, when the build runs, then its declared `scripts` agree with its own `cmap` in
  both directions and its declared SPDX id agrees with its nameID 13.
- Given a family whose own upstream publishes no static, when the batch is assembled, then it is
  excluded with the evidence recorded — never derived silently to make the number look better.
- Given the font browser after this story, when an author types `Roboto`, `Open Sans` or `Inter`, then
  the family is offered and embeds with **no download at all**.
- Given the 23 golden digests, when this story lands, then none has moved.

## Design Notes

**Why this is not a fontgen story, and why that correction matters.** D-16.5 asserted that the remedy
for variable-only families is derivation, and D-16.R.2 was taken by the owner on that framing. Both were
wrong about the mechanism: measured, this repository **cannot derive a catalogue face at all**, and the
two families D-16.5 cited as "already done" are one derivation and one upstream static. The cheap
remedy was in the repository the whole time, 21 times over. Recorded here because a later reader will
otherwise find a fontgen story that never happened.

**Why Regular only.** Bold and italic are out by a standing Non-goal (*"no synthetic bold or oblique,
and no variable-font axes. A weight is a face or it does not exist"*) and by Epic 11's placement. A
batch that quietly added weights would triple its own size and pre-empt an epic.

**The batch is a snapshot of popularity, and popularity moves.** That is the standing cost the owner
accepted, and it is the reason the owner and trigger are acceptance criteria rather than notes.

## Verification

- `cd folio-designer && npm run scan:font-hosts && npm run test && npm run build`
- `cd folio-go && go test ./...`
- The 23 golden digests, unmoved; `maximumCacheAssets` still 64 with the measured slot count reported.
- The e2e specs compile (`npm run test:e2e:compile`); heavy suites run at the end-of-run catch-up per
  D-16.R.1.
