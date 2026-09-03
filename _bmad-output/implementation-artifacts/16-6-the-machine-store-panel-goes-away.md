---
title: 'Story 16.6: The machine store panel goes away'
type: 'refactor'
created: '2026-09-03'
status: 'in-progress'
review_loop_iteration: 0
baseline_commit: 'b09c296'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/16-2-a-fetched-face-stays-on-this-machine.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The typography panel carries a section listing the typefaces this designer has downloaded. Looking at it
running, the owner judged it unnecessary: when nothing has been downloaded it is three paragraphs of
explanation above an empty list, in a panel where every other control is one line. The design never drew
it — it appears in none of the six mockups — so removing it moves the screen toward the drawing.

The list it draws is one the font menu already draws. Since Story 16.4, every typeface downloaded to this
machine appears in the dropdown under `AVAILABLE LOCALLY`, beside the ones that ship with the designer —
so the panel has been a second, narrower view of a set the author can already see in the place they go to
choose a font. Nothing about finding an installed typeface is lost by removing it.

It goes entirely, and two things go with it, both chosen deliberately. There is no longer any way to
remove a downloaded typeface from this machine. And the designer no longer says anything when it cannot
keep typefaces at all — in a private window, or with storage blocked, fonts still work, each one going
straight into the document, but nothing on screen explains that any more.

<intent-contract>

## Intent

**Problem:** `MachineFontStore` occupies a section of the typography panel to list downloaded faces, and
when nothing is stored it renders a heading, an empty-state paragraph and an unconditional explanatory
paragraph — no controls at all. The design draws no such panel in any of its six mockups; it was invented
by Story 16.2.

**Approach:** Delete the component, its call site, its per-face removal control and the store status line
it hosted, along with the state and handlers that fed them. Rewrite the two refusal sentences that name
its removal control so they stop offering a remedy that no longer exists.

## Boundaries & Constraints

**Always:**
- **The whole section goes, in every state.** Empty, populated, or degraded, the typography panel renders
  nothing where `MachineFontStore` was.
- **The two refusal sentences stop offering a remedy rather than pointing at a vanished control.**
  `lateEmbedRefusal` no longer distinguishes a removable face from a bundled one — with nothing able to
  remove anything, that parameter and both its branches collapse into one sentence.
- **The store still stores.** Fetching, keeping, reading back and embedding from the store are untouched.
  Only the panel, the removal affordance and the status line go.
- **`FontStore`'s own API is NOT changed.** `remove(key)` stays on the interface and keeps its own tests;
  the designer simply stops calling it. **The blast radius is the UI layer, and it stops there.**
- **Deleted code is deleted, not commented out or left unreachable.** State, handlers, helper functions
  and CSS rules that no longer have a caller go with it.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Any change to `FontStore`'s interface or to how faces are written, read or keyed.
- Any new surface that re-introduces a store status message somewhere else.

**Never:** a second removal affordance elsewhere · re-homing the status line · touching the font browser,
the family control's three groups, or the install/embed fork · changing what is stored or how.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Panel with nothing stored | Empty store | Typography panel renders no store section at all | — |
| Panel with faces stored | 3 stored faces | Still no store section; the faces remain offered under `AVAILABLE LOCALLY` in the family control | — |
| Store cannot be opened | Private window | Designer works; picks embed straight into the document; **nothing is said on screen** | Deliberate silence, ratified |
| Store write fails | Quota refuses | The refusal names the failure and offers **no remedy** | Refusal, not a silent success |
| Late embed refusal, stored face | Engine refuses an installed face | One sentence: refused, nothing written, face still on this machine — **no removal instruction** | — |
| Late embed refusal, bundled face | Engine refuses a local-tier face | The **same** sentence as the stored case | — |

</intent-contract>

## Code Map

All anchors measured at `b09c296`. **Re-verify before editing** — this file moves under its own change.

**`folio-designer/src/App.tsx`**
- `:1766-1791` **`MachineFontStore`** — the component. `:1789` is
  `{note && <p className="honest-note" role="status">{note}</p>}`, **the only render of `storeNote`
  anywhere in the file.** `:1786` is the per-face remove button.
- `:1728` the call site, inside the `TYPOGRAPHY` `PropertySection`.
- `:185` `const [storeNote, setStoreNote] = useState<string>()`, and its **seven** setters at `:297`,
  `:306`, `:355`, `:369`, `:370`, `:1008`, `:1130`.
- `:365-372` `removeStoredFace` — the only caller of `store.remove`.
- `:1975-1984` **`lateEmbedRefusal`**, whose `removable` branch names the panel; called at `:1123`.

**`folio-designer/src/font-store.ts`**
- `:469-470` **`storeWriteRefusal`** — its tail offers *"Free some space — the …"* and names the panel.
- `:160` and `:442` `remove(key)` on the `FontStore` interface and its implementation — **read-only here.**
- `storeUnavailableEmbedNote` and `storeWriteDegradation` lose their only consumers; check for others
  before deleting either.

**`folio-designer/src/App.css`** — `:276-283`, **8 rules** under `.machine-font-store`,
`.machine-font-list`, `.machine-font-row`, `.machine-font-name`, `.machine-font-meta`.

**Tests that assert the panel or removal — retire deliberately, and say so**
- `src/App.font-store.test.tsx:26` `storePanelHeading`, and the tests at `:459`, `:723-737`, `:851`,
  `:940-955`, `:1017`.
- `src/font-store.test.ts:286-296` — reads `App.tsx`'s source for the panel heading and asserts
  `storeWriteRefusal` contains it.
- **Measured: `TYPEFACES THIS DESIGNER HAS DOWNLOADED` appears in 0 e2e specs** (population 15 e2e files;
  positive control `Add fonts` hits `e2e/font-browser.spec.ts`). No Playwright spec addresses this panel.

## Tasks & Acceptance

**Execution:**
- [x] `folio-designer/src/App.tsx` — delete `MachineFontStore` and its call site — the section goes.
- [x] `folio-designer/src/App.tsx` — delete `storeNote` state, all seven setters, and `removeStoredFace` —
      with no render site, the state and every write to it are dead.
- [x] `folio-designer/src/App.tsx` — collapse `lateEmbedRefusal` to one sentence and drop its `removable`
      parameter — the distinction it drew no longer exists.
- [x] `folio-designer/src/font-store.ts` — cut `storeWriteRefusal`'s remedy clause; delete
      `storeUnavailableEmbedNote` / `storeWriteDegradation` **only if** a measured sweep shows no other
      consumer.
- [x] `folio-designer/src/App.css` — delete the 8 `.machine-font-*` rules.
- [x] Tests — retire the panel and removal assertions listed in the Code Map, **each with a one-line
      reason in the Delivery Log**; keep every assertion about storing, reading back and embedding.
- [x] Add a test asserting the typography panel renders **no** store section with faces present — the
      deletion's own guard, or nothing proves the section is gone.

**Acceptance Criteria:**
- Given three stored faces, when the typography panel renders, then no store section appears and the three
  faces are still offered under `AVAILABLE LOCALLY`.
- Given storage that cannot be opened, when the designer runs, then it still works, picks still embed, and
  **no message about storage appears** — the deliberate reversal of 16.2's stated-degradation clause.
- Given the engine refuses an installed face, when the refusal is shown, then it is the same sentence for a
  stored face and a bundled one, and names no removal control.

## Spec Change Log

- **2026-09-03, planning — OWNER, two decisions.** (1) The section is unnecessary and is deleted, with
  per-face removal going with it, chosen over hiding it when empty and over moving `×` onto the dropdown
  rows (D-16.R.82). (2) Planning then measured that `App.tsx:1789` is the **only** render of `storeNote`,
  so deleting the panel also deletes the store's status line — five messages, only two of them about
  removal. Offered the choice of keeping that one line, the owner chose **delete everything and accept
  silence**, knowing it reverses 16.2's ratified *"Storage failure is a degradation, stated."*
  **Recorded as a deliberate reversal, not a consequence nobody noticed.**
- **2026-09-03, planning — OWNER, the redundancy stated.** *"The font installed on the machine could be
  found in the dropdown. No need a separate panel."* This also settles a clarification raised at the gate:
  **the font store itself stays** — faces are still fetched, kept and read back, which is what makes them
  findable in the dropdown. Only the panel goes. Removing the store would unwind Story 16.5's
  install-vs-embed separation (D-16.R.46) and is explicitly not this story.
- **2026-09-03, planning — the frozen block is marked `<intent-contract>`, not the template's
  `<frozen-after-approval>`.** All eight sibling specs in this epic use the former and the epic's byte
  checks key on it; a template-matching marker here would be a zero-population match.

## Design Notes

**Two clauses this story reverses, both owner-ratified, both named so a later reader does not mistake them
for drift.** 16.2's *"Storage failure is a degradation, stated"* becomes silence. And **DW-175** — *bytes
can outlive the face record that names them, with nothing able to free them* — is owned by *"the first
story that adds store maintenance"*; there will now be none, so **it stops being deferred work and becomes
shipped behaviour.** Close it against this story rather than leaving it waiting for a story that will never
come.

**THE PANEL IS REDUNDANT, NOT MERELY UNTIDY, AND THIS IS THE OWNER'S OWN REASONING.** *"The font installed
on the machine could be found in the dropdown. No need a separate panel."* Story 16.4 made that true: the
family control's `AVAILABLE LOCALLY` group is `familyIsInstalled(source)` — the `local` and `stored` arms
together — so every downloaded face already appears where the author goes to choose a font, carrying its
own *"already downloaded to this machine"* note. **The panel listed a subset of a list the dropdown
already shows, in a place the author does not go to pick fonts.** That is why deleting it costs no
discoverability: the capability it duplicated stays, and only the duplicate goes. It also explains the
sequencing — before 16.4 this panel was the only place a downloaded face was visible at all.

**Why the design supports this rather than merely permitting it.** Measured: `AVAILABLE LOCALLY` occurs
exactly once across the six mockup files, as the family control's group heading, and **no mockup draws a
machine-store panel at all** (population 6; positive control `TYPOGRAPHY` hits 2 files). Story 16.4 gave
that borrowed label back; this story removes the borrower.

## Verification

**Nothing in Epic 16 is CI-verified** — the designer job halts at step 2 (DW-171), so every gate below is a
local measurement with no machine watching. Say what you did not run.

**Commands** — one per line; take every exit code from `$?` **immediately**, never through a pipe, a
wrapper or a trailing `echo`. The shell is zsh: `${PIPESTATUS[0]}` is wrong here.
- `cd folio-designer && npm test` — expect the count to **fall** as panel tests retire. The one permanent
  red is `canvas-authority-contract.test.ts:190` (DW-152); **match it by NAME, not count.**
- `cd folio-designer && npm run typecheck` — `rc=0`. **This is the gate that catches dead state:** an
  unused `storeNote` or an orphaned import fails here.
- `cd folio-designer && npx oxlint` — `rc=0`, exactly **4** `only-export-components` warnings.
  **Read lint from the tool directly, never `npm run lint`** — an exit code is not portable between
  agents, and that runner has already reported a false `rc=1` in this epic.
- `cd folio-designer && npm run build` — `rc=0`. `cd folio-designer && npm run test:e2e:compile` — `rc=0`.
- `cd lint && go test -count=1 ./...` and `cd folio-go && go test -count=1 ./...` — this story touches no
  Go; the `folio-go` failing leaf set must stay exactly `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`.

**Standing rules — re-run, never cite:**
- **The per-row matrix audit reports N rows, N results.** This matrix is **6 rows**. A row whose only
  assertion calls a production function directly rather than driving the path is **FAILED**; module-level
  coverage is a flagged **PARTIAL**, never upgraded.
- **State the population beside every zero**, read the matches rather than the count, and pair every
  absence claim with a positive control that must produce hits. An unquoted `grep --include=*.ts` dies in
  zsh printing nothing, which reads exactly like a clean result.
- **A COMMENT IS NOT A MEASUREMENT.** Orderings, invariants and bounds are established by a run.
- **A deletion needs a guard.** Proving a section is gone by observing a green suite proves nothing; assert
  its absence with faces present, and red-prove that assertion by restoring the render.

**Manual check:** open the designer with a populated store — the typography panel shows the family control
and no store section, and the stored faces still appear under `AVAILABLE LOCALLY`.
