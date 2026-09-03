# Token Fidelity Review — DESIGN.md vs. `Font Browser.dc.html`

Companion to [`review-token-fidelity.md`](./review-token-fidelity.md), in the same form and by the
same method: mechanical extraction from the mockup (hex, rgba, box-shadow, border-radius) diffed
against the frontmatter of `DESIGN.md`. Ground truth = the mockup. Every finding cites `file:line`
and the exact string. Raised at Story 16.3, which builds this screen.

Corpus: the modal region (`Font Browser.dc.html:290–418`), its `renderVals()` style objects
(`:478–700`), and the `Add fonts…` row that opens it (`:230–237`).
**24 distinct hex values, 4 distinct rgba values, 1 box-shadow, 1 border-radius.**

**16 of the 24 hex values map cleanly to a token** — `select`, `ink-high`, `ink`, `ink-low`,
`ink-ghost`, `line`, `line-subtle`, `line-strong`, `edge`, `edge-strong`, `panel`, `active`,
`ground`, `bind`, `bind-edge` — and are used in this build exactly as `DESIGN.md` names them. What
follows is the residue.

---

## 1. Unnamed colours

| # | Hex | Uses | Where | What it is | Severity | What Story 16.3 shipped instead |
|---|---|---|---|---|---|---|
| FB1 | `#1c2126` | 4 | `:294, 322, 382, 406` | The modal's header, rail, grid card and footer ground. **A sixth chrome surface step**, between `panel` (#1A1E23) and `raised` (#22272D). | **high** | Nothing — `DESIGN.md` refuses a sixth step by name and says what to use instead ("an interface that needs one is expressing hierarchy that a hairline should carry instead"). The four regions are one `--color-panel` surface separated by `--color-line` rules; the grid card, which is a raised object rather than a region, takes `--color-raised`. |
| FB2 | `#8fd0e6` | 3 | `:233, 483, 560` | Active chip text, staged-button text, the `Add fonts…` label. One digit off `select-bright` (#8FD0E4). | **medium** | `--color-select-bright`. Almost certainly meant to be it. |
| FB3 | `#6b7d86` | 2 | `:234, 560` | The `Add fonts…` sub-label, and the `In template` button's disabled text. Between `ink-low` (#737C86) and `ink-faint` (#5E666F). | **medium** | `--color-ink-low` for the sub-label, `--color-ink-ghost` for the disabled button — the ramp already distinguishes "quiet" from "inert" and this one value was doing both jobs. |
| FB4 | `#1d2227` | 1 | `:364` | The Row-view hover. One digit off `row-focus` (#1D2228), which `DESIGN.md` reserves in those words for the focused matrix row and calls "the single exception". | **medium** | No hover tint at all. Taking `row-focus` would spend the one reserved exception on a list this screen does not need it for; the row already separates on a hairline. |
| FB5 | `#202a30` / `#26343b` | 1 each | `:230` | The `Add fonts…` row's ground and its hover. Two more unnamed surfaces, both cyan-tinted. | **medium** | `--color-select-tint` on a `--color-select-edge` border. It is the design's own cyan wash and reads as the same object without inventing a step. |
| FB6 | `#6bb4d0` | 1 | `:414` | The confirm button's hover. Between `select` (#58A6C4) and `select-hover` (#7CC0DA). | **low** | **`--color-select-hover`.** Corrected during review: an earlier pass drew **no hover fill at all**, which removed an interaction state the design specifies. Declining an undeclared literal is not a reason to delete the element the literal was colouring — the declared neighbour exists and is what it is for. **This was raised as a token gap and is NOT one.** |
| FB7 | `#262c33` | 1 | `:330` | The Thai-toggle hover. The same unnamed value `review-token-fidelity.md` already raised as **U2** across `TableEditor.dc.html`. | **low** | Nothing. Recorded here only to note that U2 recurs on a sixth artboard — it is a real gap in `colors:`, not a one-file slip. |
| FB8 | `#232d33` | 1 | `:496` | The selected dropdown row. Unrelated to the modal; carried by the panel Story 16.4 owns. | **low** | Not built here. Flagged for 16.4. |

## 2. Unnamed washes

| # | Value | Where | What it is | Severity | What shipped |
|---|---|---|---|---|---|
| FB9 | `rgba(8,10,12,0.68)` | `:291` | The modal scrim. `tints.scrim` is `rgba(6,8,10,0.72)` — a different colour *and* a different alpha. | **high** | `--tint-scrim`. `DESIGN.md` says the washes "run one alpha ladder", so a second scrim is a second ladder. |
| FB10 | `rgba(88,166,196,0.13)` / `rgba(88,166,196,0.14)` | `:482, 559` | The active chip's fill and the staged button's fill — **two alphas, one intent**, and neither is on the ladder (`0.035`, `0.07`, `0.22`). | **medium** | `--tint-select-fill-soft` (0.07) for both. Two values a hair apart is the ladder being ignored twice rather than extended once. |

## 3. Shadow

| # | Value | Where | Severity | What shipped |
|---|---|---|---|---|
| FB11 | `box-shadow: 0 24px 70px rgba(0,0,0,0.6)` | `:292` | **high** | **`--shadow-sheet`, minted by this story.** Corrected during review: this was first raised as "not a story's to make", and that was wrong. `DESIGN.md:474-484` **already declares** the sheet elevation at `0 18px 60px rgba(0,0,0,0.6)` over `{tints.scrim}`, and `components.sheet.shadow` carries the same value in the frontmatter. Nothing was designed — a declared elevation was transcribed into the file that implements declarations, so the count stays three and *"no fourth may be added"* is untouched. Routed through `design-contract.test.ts`, which now asserts both implemented elevations **from `DESIGN.md`'s own text**. **The wider finding is registered as DW-178**: `tokens.css` carried one of three declared elevations and five surfaces shared it, two of them mis-elevated and one (`.property-options`, a dropdown) outside the taxonomy entirely. |

## 4. Radius

`borderRadius: "50%"` at `:677` is the footer's staged dot — exactly `rounded.dot`, which `DESIGN.md`
reserves for "the 4–5 px status dots". Correct as drawn; shipped as `--radius-dot`. No other radius
appears anywhere in the file, which agrees with `rounded.DEFAULT: 0`.

## 5. Content the tokens cannot judge, recorded here because the same discipline applies

| # | Where | What | Why it did not ship |
|---|---|---|---|
| FB12 | `:296` | `web font library · 1,946 families` | Two false claims: 1,946 is what the source **published** on the snapshot date, not what this designer can add, and "library" reads as live over a build-time snapshot. Replaced by `familyIndexDisclosure()`, which Story 16.1 already shipped and which derives from `addableFamilyCount`. |
| FB13 | `:411`, `:680` | `≈ N weights · subset latin+thai` | This product embeds exactly one upright Regular per family and subsets nothing. The footer exists to state what is going into the file; the mockup's version states something else. |
| FB14 | `:369`, `:633` | The `designer` column and the `Most styles` sort arm | Neither the generated module nor the raw snapshot carries a designer, and style count sorts on a difference this product erases (D-16.R.33 R3). |
| FB15 | `:656` | `Cyrillic` and `Greek` writing-system chips | The snapshot records `latin` and `thai` only. A chip that can only ever empty the list is a false affordance, so the chips are derived from the vocabulary the data actually carries — which also surfaces `Handwriting`, a category the mockup's four chips omit and which covers 337 families. |
| FB16 | `:236` | The `⌘G` hint glyph beside `Add fonts…` | `⌘G` is the browser's Find Next, and this application puts app-specific actions on Option (D-16.R.33 R2, owner-confirmed). No shortcut is bound in this epic, so no glyph is drawn — a label beside a key that does nothing is a false UI string. |

---

## The discriminator, stated once because it decided every row above

**The test is not severity. It is whether `DESIGN.md` already says it.** Transcribe a declared value;
refuse an undeclared one.

That single rule separates the two corrections this review took. **FB11's sheet shadow was declared** —
in the elevation table and again in `components.sheet` — so implementing it is transcription and the
story mints it. **FB6's confirm hover and FB9's scrim are undeclared drift** — the mockup reaching past
the ramp — so they are refused in favour of `--color-select-hover` and `--tint-scrim`, and the deltas go
on the owner's deviation list rather than into the token file.

The failure mode the rule prevents is symmetric, and this review made **both** mistakes before catching
them: refusing a declared value as "not ours to add" (FB11) leaves the design system unimplemented and
every surface borrowing the wrong token; and treating an undeclared value as a gap to be filled (FB6, as
first raised) mints tokens nobody designed. A third instinct — **delete the element rather than resolve
its value** — appeared three times in this story and is worse than either, because it silently drops
what the design specifies instead of recording a disagreement.

## Recommendation

**The token this review first asked for has been minted (FB11).** `--shadow-sheet` now carries
`components.sheet.shadow`'s declared value, `.font-browser` uses it, and `design-contract.test.ts`
reads both implemented elevations out of `DESIGN.md`'s own text so neither can drift from the
declaration. The first version of this paragraph said the mint "is not this story's to make"; that was
wrong, and it is left corrected rather than deleted — transcribing an elevation the design system
already declares is implementation, not design.

**What remains is the wider taxonomy, registered as DW-178 and not this story's to settle:**
`.table-editor` and `.pdf-preview-scroll canvas` still borrow `--shadow-page` where the sheet
elevation is the one they mean, and `.property-options` — a dropdown, which is neither a page nor a
sheet — sits outside the three declared elevations altogether. Three surfaces, one question: which
elevation does a transient overlay have, and does `DESIGN.md` need to declare a fourth or say that a
dropdown carries none.

**Two values worth naming in `colors:`** if a third screen wants them: FB2's `#8fd0e6` (almost
certainly a typo for `select-bright`, so probably a mockup fix rather than a token) and FB7's
`#262c33`, which `review-token-fidelity.md` already raised as U2 and which now appears on a sixth
artboard.

**One note on the guard itself, earned the hard way during this review.** `design-contract.test.ts`'s
colour-literal scan reads **raw** text and does not blank comments, so an explanation that *quotes* the
hex values it is explaining fails the check. That is correct behaviour and not a rough edge: **a literal
in a comment is still one somebody can copy.** The CSS comment recording this decision names no hex.

**Everything else is the mockup reaching past the ramp**, and `DESIGN.md` wins where the two
disagree — which is what its own header says it does.
