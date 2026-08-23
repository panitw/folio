# Token Fidelity Review — DESIGN.md vs. the five mockups

Method: mechanical extraction from `mockups/*.html` (hex, rgba, font-size, font-family,
border-radius, box-shadow, gradient, padding/gap, component markup) diffed against the
frontmatter of `DESIGN.md`. Ground truth = the mockups. Every finding cites file:line and
the exact string.

Corpus: 61 distinct hex values, 11 distinct rgba values, 9 distinct font-sizes,
5 distinct font-family declarations, 20 `border-radius` sites, 6 `box-shadow` values,
3 `gradient` sites.

---

## 1. Unnamed colours

Hex values used in a mockup that map to **no token** in `colors:`. A developer has no name
to reach for and will invent one.

| # | Hex | Uses | Where | What it is | Severity |
|---|---|---|---|---|---|
| U1 | `#7cc0da` | 5 | all five files, line 13 | `a:hover { color: #7cc0da; }` — the link-hover accent. Present in **every** mockup and named nowhere. | **high** |
| U2 | `#262c33` | 4 | `TableEditor.dc.html:29,79,91,127` | `border: 1px solid #262c33` — the sheet's mode-switch frame and its "none" selects. Sits between `line` (#272c33) and `line-subtle` (#22272d) with no name. | **high** |
| U3 | `#1c2025` | 3 | `TableEditor.dc.html:79,91,127` | `background: #1c2025` on inactive/empty select fields — a sixth chrome surface step the five-step ramp does not admit. | **high** |
| U4 | `#171a1e` | 3 | `Binding.dc.html:222`, `TableEditor.dc.html:189,197` | `background: #171a1e` — sheet footer/context strip. Another unnamed surface between `base` (#15181c) and `panel` (#1a1e23). | **high** |
| U5 | `#edf0f3` | 5 | `Main.dc.html:221-225` | `border-bottom: 1px solid #edf0f3` on placeholder table rows. Distinct from both `page-rule` (#f0f2f5) and `page-guide` (#eef1f4) — three near-identical page hairlines, one unnamed. | **medium** |
| U6 | `#22262b` | 1 | `TableEditor.dc.html:30` | `background: #22262b` — inactive mode-switch segment. Off by one digit from `raised` (#22272d); almost certainly meant to be it. | **medium** |
| U7 | `#6b737d` | 1 | `TableEditor.dc.html:30` | `color: #6b737d` — inactive mode-switch label. Off by a hair from `ink-low` (#737c86). | **medium** |
| U8 | `#8b949e` | 1 | `TableEditor.dc.html:191` | `color: #8b949e` on the sheet's explanatory sentence. Between `ink` (#aab2bb) and `ink-low` (#737c86). | **medium** |
| U9 | `#33505c` | 2 | `TableEditor.dc.html:24` | `border: 1.5px solid #33505c` / `background: #33505c` — the *dimmed* Folio mark inside the sheet. A muted `select` with no token. | **medium** |
| U10 | `#d8bd7e` | 1 | `Binding.dc.html:178` | `color: #d8bd7e` — "Text selected · binding to". A lighter amber than `bind-text` (#e0c07a). | **medium** |
| U11 | `#1e1c16` | 1 | `Binding.dc.html:176` | `background: #1e1c16` — binding-panel header. Not `bind-tint` (#241f14) and not `bind-tint-warm` (#1e1b13); a third amber ground. | **medium** |
| U12 | `#23262a` | 1 | `TableEditor.dc.html:35` | `background: #23262a` — the dimmed page behind the scrim. | **low** |

`#1D2228` (TableEditor.dc.html:84, focused matrix row) is named — but only inside
`components.matrix-row.focusBackground`, never in the `colors:` block. It is the only raw hex
in the component specs. **medium** — promote it to a token or drop it.

### Unnamed rgba

`DESIGN.md` names three rgba values (band tint, sheet scrim, the shadows). Six more exist:

| # | Value | Uses | Where | Severity |
|---|---|---|---|---|
| U13 | `rgba(201,167,88,0.14)` | 5 | `Preview.dc.html:196-200` — the overflow-row tint on the page | **high** |
| U14 | `rgba(201,167,88,0.28)` | 1 | `Preview.dc.html:67` — thumbnail overflow marker | medium |
| U15 | `rgba(201,167,88,0.10)` | 1 | `Binding.dc.html:115` — selected bound element fill | medium |
| U16 | `rgba(88,166,196,0.22)` | 1 | `TableEditor.dc.html:87` — focused-row cell wash | medium |
| U17 | `rgba(88,166,196,0.07)` | 1 | `Main.dc.html:180` — selected element fill | medium |
| U18 | `rgba(0,0,0,0.5)` | 1 | `TableEditor.dc.html:35` — see R2 | medium |

There is an entire amber-tint and cyan-tint alpha family in the mockups (0.035 / 0.07 / 0.10 /
0.14 / 0.22 / 0.28) of which the token file names exactly one.

---

## 2. Phantom tokens

Declared in `DESIGN.md`, appear in **no** mockup.

| # | Token | Declared | Reality | Severity |
|---|---|---|---|---|
| P1 | `colors.danger` | `'#D1655C'` | Zero occurrences across all five files. The `error-card` component (`border: 1px solid {colors.danger}`) is likewise never built — `Preview.dc.html:272-283` shows only a *legend* row describing the error shape ("Square, solid — render failed") drawn in `#737c86`. The red is unverified by any mockup. | **high** |
| P2 | `rounded.full` | `'9999px'` | Never used. All 20 radius sites are `border-radius: 50%`. Functionally identical, but a developer reading `{rounded.full}` will emit `9999px`, which the mockups never do. | **low** |

Every other hex in the `colors:` block does appear at least once. `#4A3D1E` (`bind-edge`) appears
exactly once (`Main.dc.html:356`) and `#12222A` (`select-tint-deep`), `#0A0C0E` (`ground-preview`)
once each — thin, but real.

---

## 3. Value mismatches

### V1 — `components.segmented-control.border` matches no mockup. **high**
Declared `border: '1px solid {colors.edge}'` (= `#3A4048`). Actual:

- `Main.dc.html:57` — `border: 1px solid #363d46` (that is `ink-pin`, a *text* token)
- `Preview.dc.html:41` — `border: 1px solid #58a6c4` (that is `select`)
- `TableEditor.dc.html:29` — `border: 1px solid #262c33` (unnamed, U2)

Three mockups, three different borders, none of them `edge`. A developer building from the
token gets a fourth appearance.

### V2 — `components.segmented-control.activeBackground` is wrong in preview. **high**
Declared `{colors.active}` (`#2A3037`). True in `Main.dc.html:58`
(`background: #2a3037; color: #e6e9ec`). But `Preview.dc.html:43` is
`background: #1e3b46; color: #cfe8f2; font-weight: 600` — `select-fill` on `select-on-fill`.
The active mode chip changes colour *and* weight between modes; the token file says it does not.

### V3 — `components.band-tab.font` contradicts the prose and the mockups. **high**
Frontmatter says `font: '{typography.mono-sm}'` (10px). The prose says "mono 9 px". The
mockups say 9px: `Main.dc.html:139` —
`font-family: 'IBM Plex Mono', monospace; font-size: 9px; letter-spacing: 0.08em; color: #58a6c4; border: 1px solid #2f4f5c; background: #16232a; padding: 3px 7px;`
The frontmatter (the machine-readable half) is the half that is wrong. Also note `0.08em`
tracking, which no typography token carries.

### V4 — `typography.brand` does not describe the load screen. **medium**
Token: `fontSize: '11px'`, `letterSpacing: '0.04em'`. True in Main/Binding/Preview/TableEditor
(`Main.dc.html:27`). But `Load.dc.html:27` is
`font-size: 13px; font-weight: 500; letter-spacing: 0.06em`. The load screen scales the brand
mark (and its box: `22px` vs `18px`) and no token or note records it.

### V5 — `components.status-bar.height` is not constant. **medium**
Token: `height: '24px'`; the Layout section says "a 24 px status bar below … The frame is
constant". True in `Main.dc.html:369` and `Binding.dc.html:269`. But `Preview.dc.html:294` is
`height: 32px` (the page-nav bar). Preview's bottom bar is 33% taller and the token file
does not say so.

### V6 — `components.tree-node.indent` is not 18px. **medium**
Token: `indent: '18px'`. `Binding.dc.html` uses `padding: 4px 12px 4px 30px` (level 1) and
`padding: 3px 12px 3px 42px` (level 2) — a first step of 18px and a second step of **12px**.
The indent is not a constant.

### V7 — `components.binding-chip` is not honoured in the table editor. **medium**
Token: `background: {colors.bind-tint}` (#241F14), `border: 1px solid {colors.bind-edge}`
(#4A3D1E), `dotSize: '5px'`. Honoured at `Main.dc.html:356`
(`background: #241f14; border: 1px solid #4a3d1e`) with a 5px dot. But the five chips in
`TableEditor.dc.html:76,88,100,112,124` are
`background: #22272d; border: 1px solid #3a3118` with `width: 4px; height: 4px` dots — a
different ground, a different border (`bind-edge-soft`), a different dot size. "The canonical
way to display a binding anywhere in the product" is not canonical.

### V8 — `components.selection-handle.outline` does not exist. **medium**
Token declares both `border: '1px solid {colors.ground}'` **and**
`outline: '1px solid {colors.select}'`. The markup has only the border:
`Main.dc.html:184` — `width: 6px; height: 6px; background: #58a6c4; border: 1px solid #0d0f12;`.
No `outline` property appears in any mockup.

### V9 — `components.progress-bar.height` covers only one of the two bars. **low**
Token: `height: '5px'`, `trackBorder: '1px solid {colors.line-subtle}'`. True for the summary
bar (`Load.dc.html:96` — `height: 5px; background: #1a1e23; border: 1px solid #22272d`). The
inline per-item bar at `Load.dc.html:83` is `height: 3px; background: #1a1e23` with **no**
border. Two progress bars, one token.

### V10 — `components.palette-item` height. **low**
Prose says "32 px row". Actual: `Main.dc.html:72` — `padding: 7px 12px` around a 16px icon =
**30px**. All other palette-item values (`padding 7px/12px`, `background: #22272d`,
`border-left: 2px solid #58a6c4`, 16px icon, `#4e565f` shortcut) are correct.

### V11 — "Green appears once, on the hash match" is false. **low**
`#7fbf8a` appears 6 times: `Load.dc.html:41,49,57` (three manifest check glyphs),
`Preview.dc.html:38,235,236` (the "current" state word and the hash match). The prose in
*Colors* contradicts `components.manifest-row.doneIcon: '{colors.ok}'` in the same file.
The component spec is right; the sentence is wrong.

---

## 4. Rule violations

### R1 — 11px is **not** the default. **high**
Claim: "Eleven pixels is the default, not a floor to escape."
Measured across all five mockups:

| size | uses | token |
|---|---|---|
| 10px | **142** | `body-sm` / `mono-sm` |
| 8px | 94 | page ramp only |
| 9px | 51 | `label` |
| **11px** | **34** | `body` / `mono` |
| 7px | 5 | **none** |
| 12px | 4 | `title` |
| 13px | 4 | `page-title` |
| 22px | 1 | `numeric-lg` |
| 19px | 1 | `display` |

10px outnumbers 11px more than four to one. 11px survives mostly as the inherited root
(`font-size: 11px` on the outer frame div in each file) and on a handful of explicit spans.
A developer told "11px is the default" will build a visibly larger UI than the mockups.
**The `body-sm`/`mono-sm` 10px pair is the de-facto default and should be described as such.**

### R2 — A fourth shadow exists. **medium**
Claim: "shadows appear in exactly three places… No other element casts a shadow. Ever."
Four distinct non-inset values exist:

| value | site | status |
|---|---|---|
| `0 2px 24px rgba(0,0,0,0.55)` | `Main.dc.html:153`, `Binding.dc.html:97` | ✅ matches `page-surface.shadow` |
| `0 4px 32px rgba(0,0,0,0.7)` | `Preview.dc.html:89` | ✅ matches prose |
| `0 18px 60px rgba(0,0,0,0.6)` | `TableEditor.dc.html:43` | ✅ matches `sheet.shadow` |
| `0 2px 24px rgba(0,0,0,0.5)` | `TableEditor.dc.html:35` | ❌ **undeclared** |

`TableEditor.dc.html:35` —
`width: 496px; height: 701px; background: #23262a; box-shadow: 0 2px 24px rgba(0,0,0,0.5);`
This is the page seen dimmed behind the sheet scrim. It is the page shadow at `0.5` instead
of `0.55`. Either an intentional dimmed variant that needs naming, or a typo. Also present:
one `inset` shadow, `box-shadow: inset 2px 0 0 #58a6c4` (`TableEditor.dc.html:84`), which is
correctly declared as `matrix-row.focusMarker` — inset is a marker, not elevation, so it is
not a fourth shadow.

### R3 — Chrome amber printed on the white page. **high**
Claim: "`{colors.bind-on-page}` … the chrome amber fails contrast there" and
"Don't mix page colours into chrome or chrome colours onto the page."
`bind-on-page` (`#a8801f`) is used correctly in 21 places (e.g. `Main.dc.html:221-225`,
`Binding.dc.html:110`). But chrome amber `#c9a758` is set **directly on the page** at:

- `Main.dc.html:171` — `font-size: 8px; color: #c9a758; text-align: right;">{{params.branchName}}`
- `Main.dc.html:172` — same, `{{params.reportDate}}`
- `Binding.dc.html:115` — `border: 1px solid #c9a758; background: rgba(201,167,88,0.10)` on the page
- `Preview.dc.html:196-200` — `border-bottom: 1px dashed #c9a758` on page table rows

Two placeholders in the same page header use two different ambers for the same job. This is
the single most likely thing a developer gets wrong, because the rule and the mockup disagree
in the same file.

### R4 — Amber selection handles. **high**
Claim: "Selection handles — 6 px squares in `{colors.select}`" and "Don't use amber for
selection or cyan for data — it breaks the one rule users learn."
`Binding.dc.html:118-121` — four handles, `background: #c9a758; border: 1px solid #0d0f12`,
on the selected `{{customer.account}}` element, with an `8px #c9a758` "bound" readout at
`Binding.dc.html:122`. `Main.dc.html:184-191` uses cyan handles correctly. Either the mockup
breaks the accent grammar, or the grammar has a documented exception (selection *of a bound
element* takes the data colour) that the token file omits. Must be resolved before build.

### R5 — Uppercase tracking is not `0.1em`. **medium**
Claim: "Uppercase belongs only to section labels and mode switches, always with `0.1em`
tracking and 600 weight." `typography.label` = 9px / 600 / 0.1em. Actual 9px tracking:
`0.1em` ×12, `0.09em` ×9, `0.08em` ×5, `0.07em` ×1, `0.04em` ×2. The mode switches — explicitly
named in the claim — are `letter-spacing: 0.04em` at **10px** with weight 500 or unset
(`Main.dc.html:58`, `Preview.dc.html:43`, `TableEditor.dc.html:30`), not 9px/600/0.1em. Five
distinct tracking values exist against two tokens (`0.04em`, `0.1em`, plus `0.14em` for
`page-eyebrow`).

### R6 — 7px type exists and no role covers it. **medium**
Five uses, three files: `Binding.dc.html:103,151`, `Main.dc.html:165,240`,
`Preview.dc.html:207` — e.g.
`font-family: 'IBM Plex Sans Thai', sans-serif; font-size: 7px; color: #6b7480;` (the Thai
confidentiality footer) and the 7px mono `logo.png` placeholder label. The page ramp bottoms
out at 8px. A `page-fine` role is missing.

### R7 — 12px used outside "sheet titles". **medium**
`typography.title` 12px is scoped to "Sheet titles" and the Don'ts say "Don't raise type above
11 px to signal importance. Only the load screen has display type." `TableEditor.dc.html:48`
is a genuine sheet title (`font-size: 12px; font-weight: 500`, ✅). The other three are on the
load screen: `Load.dc.html:32` (12px sans body copy, "Cached after this…"), `Load.dc.html:91`
and `:93` (12px mono, "of 9.5 MB" and "51%"). These are a load-screen body and load-screen
numeric-small; neither is a title, and neither has a token.

### R8 — Spacing values off the declared scale. **medium**
Declared: 4 / 6 / 8 / 10 / 12 / 14 / 16, plus 22 / 28 / 34 for load. Measured frequency of px
values inside `padding`: `4px` ×227, `5px` ×84, `8px` ×76, `12px` ×69, **`3px` ×31**, `7px` ×21,
`6px` ×19, **`9px` ×17**, `14px` ×15, `10px` ×14, `30px` ×8, `16px` ×8, `42px` ×5, `11px` ×3,
`22px` ×1. Gaps add `2px` ×4, `5px`, `1px`, and `26px` ×1 (`Load.dc.html:112`).

- `5px` (84) and `9px` (17) are covered by the prose ("dense rows use 5 px, standard rows 7 to
  9 px") and by `spacing.row-y: 5px`, but `7px` and `9px` have no named entry — `palette-item`
  and `manifest-row` inline them as raw `'7px'` / `'9px'`.
- **`3px` (31 uses) is entirely undeclared** — `padding: 3px 7px` (band tab), `padding: 0 3px`
  (page placeholders), `padding: 3px 8px` / `3px 9px` / `3px 10px` (Preview page-nav buttons).
  It is the fifth most common spacing value in the product.
- `11px` (`TableEditor.dc.html:189`, `Preview.dc.html:287`), `26px` gap and `30px`
  (`Load.dc.html:106` `margin-top: 30px`) are one-offs off-scale. The load-screen exception is
  documented as 22/28/34 — `30px` is not among them.

### R9 — Gradients: two hits, **both defensible, neither a violation**. **low**
Claim: "Don't use a gradient anywhere." Three sites:

1. `Main.dc.html:153` and `Binding.dc.html:97` —
   `background-image: radial-gradient(#dfe3e8 0.6px, transparent 0.6px); background-size: 12px 12px`
   **Not a violation.** This is the page dot grid, and `DESIGN.md` itself declares it
   (`page-surface.gridDot: '{colors.page-dot}'`, `gridPitch: '12px'` — both match exactly:
   `#dfe3e8` is `page-dot`, the pitch is 12px). `radial-gradient` is the standard CSS idiom for
   a dot lattice; there is no colour ramp and nothing reads as a gradient. The rule means
   "no decorative colour transitions", and this is a repeating pattern.
2. `Main.dc.html:343` —
   `background: repeating-conic-gradient(#3a4048 0% 25%, #22272d 0% 50%) 50% / 6px 6px`
   **Not a violation.** This is the checkerboard transparency swatch in the fill property
   field — the universal convention for "no fill". `repeating-conic-gradient` has hard stops
   (`0% 25%`) and produces flat squares, not a blend. Both colours (`edge`, `raised`) are tokens.

Recommendation: keep the rule but reword it — *"No colour ramps. `linear-gradient` and blended
`radial-gradient` are forbidden; hard-stop gradients used as repeating patterns (the page dot
grid, the transparency checkerboard) are patterns, not gradients."* As written, the rule bans
two things the mockups deliberately do, and a literal-minded developer will remove them.

### R10 — Google Fonts links do not match usage in two files. **low**
`Load.dc.html:10` and `TableEditor.dc.html:10` request
`family=IBM+Plex+Sans+Thai:wght@400;500` (no 600); `Binding`, `Main`, `Preview` request
`400;500;600`. Neither Load nor TableEditor uses Thai at all (0 occurrences of
`Plex Sans Thai` in both), so nothing renders wrong — but the two files that *do* need Thai 600
for `page-title` (`font-weight: 600` at `Binding.dc.html:105`, `Main.dc.html:168`,
`Preview.dc.html:97`) correctly request it. Fix: drop the Thai request from Load and
TableEditor, or normalise all five links.

---

## 5. Verified correct

Everything below was checked mechanically and **matches**.

**Colours.** 47 of the 49 declared tokens appear in at least one mockup at the role
`DESIGN.md` describes. The five-step chrome ramp (`#0d0f12` / `#15181c` / `#1a1e23` /
`#22272d` / `#2a3037`), the four-step rule ramp (`#191d21` / `#22272d` / `#272c33` /
`#2f353d`), the seven-step ink ramp, the full page palette (12 tokens, all present), and both
accents are exact. `ground-preview` `#0a0c0e` appears once, in `Preview.dc.html` only, as
claimed. `select-tint-deep` `#12222a` appears once, in Preview only.

**Font families.** Exactly three faces. No fourth family anywhere. 174 `'IBM Plex Mono',
monospace`, 22 `'IBM Plex Sans Thai'`, 8 explicit `'IBM Plex Sans'` plus the body default
(`body { font-family: 'IBM Plex Sans', system-ui, sans-serif; }` at line 12 of all five files).
No Inter, Roboto, Arial, or system-only stack. The mono/sans split holds: machine values are
mono throughout (`{{customer.account}}`, `1.5 MB`, `196 × 20`, `folio-go v0.1`).

**Shape rule.** Zero violations. All 20 `border-radius` declarations are `50%`, and every one
is on a small dot: 15 × `width: 5px; height: 5px` (unsaved marker `Main.dc.html:34`,
bindable-node dots `Binding.dc.html:31,194,199,211,216,249,254`, disabled-node dots
`Binding.dc.html:233-237`, `Preview.dc.html:31`, `Main.dc.html:357`) and 5 × `width: 4px;
height: 4px` (TableEditor binding chips). No radius on any panel, field, button, card, or
sheet. The `0` default is honoured everywhere by omission.

**Display type claims.** `font-size: 19px` appears **exactly once**, `Load.dc.html:31`
(`font-size: 19px; font-weight: 500; color: #e6e9ec; letter-spacing: -0.01em`) — ✅ load screen
only, and the weight/tracking match `typography.display` exactly. `font-size: 22px` appears
**exactly once**, `Load.dc.html:90` (`font-size: 22px; font-weight: 500`) — ✅ the megabyte
count, matching `typography.numeric-lg`. Both claims hold.

**Layout.** Every frame dimension is exact: 180px palette rail (`Main:67`, `Binding:52`),
300px properties panel (`Main:257`, `Binding:161`), 132px thumbnail rail (`Preview:50`),
320px render panel (`Preview:215`), 560px load column (`Load:20`), 1440×900 artboards.

**Document bar.** `Main.dc.html:21` — `height: 40px; background: #1a1e23; border-bottom: 1px
solid #272c33; gap: 16px; padding: 0 12px` — matches `components.doc-bar` in all five values
(`panel`, `line`, `spacing.5`, `spacing.7`). Identical in Binding and Preview.

**Status bar.** `Main.dc.html:369` — `height: 24px; background: #1a1e23; border-top: 1px solid
#272c33; font-family: 'IBM Plex Mono'; font-size: 10px; color: #737c86` — matches
`components.status-bar` exactly (`panel`, `line`, `ink-low`, `mono-sm`). Same in Binding.
(Preview differs — see V5.)

**Property field.** `components.property-field` verified at `Main.dc.html:274,279,284,289` —
`background: #22272d; border: 1px solid #2f353d` = `raised` on `line-strong`, exact. Value
colour `#e6e9ec` (`ink-high`), affix `#4e565f` (`ink-ghost`), padding `4px 8px`
(`spacing.1`/`spacing.3`) — all correct. 20+ instances, no drift.

**Band boundary.** `Main.dc.html:149,150` and `Binding.dc.html:94,95` —
`left: -14px; right: -14px; border-top: 1px dashed #58a6c4` — the 14px overhang and dashed cyan
are exact.

**Sheet.** `TableEditor.dc.html:43` — `background: #1a1e23; border: 1px solid #3a4048;
box-shadow: 0 18px 60px rgba(0,0,0,0.6)` over `TableEditor.dc.html:40`
`background: rgba(6,8,10,0.72)`. All four values match `components.sheet` character for
character.

**Diagnostic card.** `Preview.dc.html:252` — `border: 1px dashed #6b5726; border-left: 3px
dashed #c9a758; background: #1e1b13; padding: 9px 10px` — matches `components.diagnostic-card`
exactly (`bind-edge-dash`, `bind`, `bind-tint-warm`), with the triangle icon at
`Preview.dc.html:274`. The shape-first rule is honoured: `Preview.dc.html:272-283` carries an
explicit legend distinguishing triangle/dashed from square/solid **in neutral grey**, so the
distinction survives without colour.

**Manifest row.** `Load.dc.html:40,48,56` — `padding: 9px 0; border-bottom: 1px solid #191d21`
= `line-faint`, exact. Three glyph states present and correctly coloured: check `#7fbf8a`
(`ok`), filled square in ring `#58a6c4` (`select`, `Load.dc.html:66-68`), dash `#454c55`
(`ink-disabled`, `Load.dc.html:79`).

**Band tab colours.** `background: #16232a` (`select-tint`), `border: 1px solid #2f4f5c`
(`select-edge`), `color: #58a6c4` (`select`), `padding: 3px 7px` — all four exact at
`Main.dc.html:139,142,145` and `Binding.dc.html:92`. Only the declared *font* is wrong (V3).

**Tree node.** `Binding.dc.html:198` — `background: #241f14; border-left: 2px solid #c9a758` =
`bind-tint` + 2px `bind` marker, exact. Disabled opacity `0.42` present once
(`components.tree-node.disabledOpacity`), and disabled nodes do carry a stated reason
(`Binding.dc.html:229`), honouring the "state the reason" Do.

**Matrix row.** `TableEditor.dc.html:84` — `background: #1d2228; box-shadow: inset 2px 0 0
#58a6c4` — matches `components.matrix-row.focusMarker` and `focusBackground` exactly. Header
and body share `grid-template-columns: 26px 30px 1fr 1.15fr 92px 104px 118px 30px`, so the
"grid template is shared" claim holds.

**Load-screen spacing exception.** `34px` (`Load.dc.html:23` `margin-bottom: 34px`), `28px`
(`Load.dc.html:37` `margin-top: 28px`), `22px` (`Load.dc.html:88` `margin-top: 22px`) — all
three present, on the load screen only, exactly as `spacing.8/9/10` claim. Confirmed absent
from the other four files.

**No emoji.** None found. Icons are stroked SVG on a 16 viewBox throughout. The only
non-ASCII glyphs are `◀`/`▶` (`Preview.dc.html:296,298`), geometric shapes in a nav control,
and `·` separators — not emoji.

**No forbidden affordances.** No share button, avatar, sync indicator, or autosave string in
any mockup. `Load.dc.html:114-121` states the model explicitly ("No account", "Renders in this
tab", "Files stay on this machine").

**Progress + number pairing.** Both progress bars are paired with numerals
(`Load.dc.html:90-93` `4.8` / `of 9.5 MB` / `51%`; `Load.dc.html:71` `2.8 / 7.4 MB`),
honouring "a bar without numbers is not acceptable".

---

## Summary

| Category | Count | High | Medium | Low |
|---|---|---|---|---|
| Unnamed colours (hex) | 12 (+1 half-named) | 4 | 7 | 1 |
| Unnamed colours (rgba) | 6 | 1 | 5 | 0 |
| Phantom tokens | 2 | 1 | 0 | 1 |
| Value mismatches | 11 | 3 | 5 | 3 |
| Rule violations | 10 | 2 | 5 | 3 |
| **Total findings** | **41** | **11** | **22** | **8** |
| Verified correct | 24 groups | — | — | — |

**Fix first, in order:**

1. **R1** — restate the default as 10px, or resize the mockups. Every screen depends on this.
2. **R3 / R4** — settle the accent grammar on the page: is a selected *bound* element cyan or
   amber, and which amber prints on white. Two mockups currently answer differently.
3. **V1 / V2** — give the mode switch one border and one active treatment, or document the
   per-mode variants as tokens.
4. **P1** — `colors.danger` and the whole `error-card` are unbuilt. Either mock the error
   state or mark the token provisional.
5. **U1-U4** — name the four surfaces/borders the sheet and the link-hover need, or replace
   them with existing tokens.

**Also worth doing:** reword the gradient Don't (R9) so it stops banning the dot grid and the
checkerboard, and add tokens for `3px` spacing, `7px` type, and the amber/cyan tint alphas.
