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
| FB6 | `#6bb4d0` | 1 | `:414` | The confirm button's hover. Between `select` (#58A6C4) and `select-hover` (#7CC0DA). | **low** | `--color-select-hover` is the named neighbour; this build draws no hover fill on the confirm button at all, so the value is simply unused. |
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
| FB11 | `box-shadow: 0 24px 70px rgba(0,0,0,0.6)` | `:292` | **high** | `--shadow-page`. `DESIGN.md`'s *Elevation & Depth* lists **three** shadows and says "no fourth may be added"; the nearest listed one is the sheet's `0 18px 60px rgba(0,0,0,0.6)`, which is not in `tokens.css` as a token at all. **This is a genuine gap to close**: the modal is a floating surface and the product's only shadow token is the page's. Raised rather than resolved — minting `--shadow-sheet` is a `DESIGN.md`/`tokens.css` change, not a story's to make on its own. |

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

## Recommendation

**One token to add, and it is not this story's to add:** a `sheet` shadow in `tokens.css` matching
`components.sheet.shadow`, so a floating surface has a shadow of its own rather than borrowing the
page's (FB11). Both the table editor and this modal currently borrow it.

**Two values worth naming in `colors:`** if a third screen wants them: FB2's `#8fd0e6` (almost
certainly a typo for `select-bright`, so probably a mockup fix rather than a token) and FB7's
`#262c33`, which `review-token-fidelity.md` already raised as U2 and which now appears on a sixth
artboard.

**Everything else is the mockup reaching past the ramp**, and `DESIGN.md` wins where the two
disagree — which is what its own header says it does.
