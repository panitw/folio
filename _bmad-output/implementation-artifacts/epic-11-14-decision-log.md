# Epics 11–14 — decision log

The audit trail for the run that delivers Epics 11, 12, 13 and 14. One entry per decision that outlives
the story that raised it. Written for someone who was not in the run.

**`## Lead Grounding` is appended below once the engineering lead reports.**

---

## Standing decisions (settled by the owner at the terminal, 2026-09-05)

### D-000.1 — The run's five standing parameters
**Owner decision.** Answered at the terminal at invocation, unprompted by the setup questions.

**Verdict.**
1. **Answer channel: this terminal.** Not Telegram, and this holds for every question for the whole run.
2. **Run continuously to the end.** No checkpoint pause after a story, and none at an epic boundary.
3. **Heavy tests at the end of each epic.** Integration and e2e execution moves to the epic boundary;
   unit tests, typecheck, lint and build still run every story.
4. **Spec approval is delegated** to the orchestrator and the engineering lead. The owner reads the
   plain-terms opener in the per-story report.
5. **Build order: to be settled below** — see D-000.3.

**Consequences.** Under (3) every story's `## Verification` carries unit, typecheck, lint and build, plus
a **compile-only** check of any new integration or e2e package under its build tag. That compile check is
not optional: a package that fails to compile under its tag **silently skips its tests**, so without it
the epic-boundary catch-up opens with a compile error from five stories ago. Each story's Delivery Log
must name the unrun suites explicitly rather than reporting a bare "green".

Under (2), a genuine fork still stops the run — continuous means no *scheduled* pauses, not that a
decision the direction cannot settle gets guessed at.

**How we'd know it was wrong.** An epic-boundary catch-up run that goes red against three or more stories
at once: that is the deferral's cost arriving, and it would argue for moving to `every-story` for the
remaining epics.

---

### D-000.2 — Epics 11–14 have no prior rulings, despite a log that appears to cover them
**Orchestrator decision** (a record-keeping finding, not a design choice).

**Verdict.** This run starts a NEW decision log. `epic-8-15-decision-log.md` does not cover epics 11–15.

**Situation.** Four decision logs already exist. The one named `epic-8-15-decision-log.md` reads as though
it spans Epic 8 to Epic 15, which would make this run a continuation with rulings to inherit. Measured
instead: of its **92** `### D-` entries, the prefixes are **73 × `D-8.`**, **13 × `D-000.`** and
**6 × `D-16.`** — and `D-11.` through `D-14.` occur **zero** times across both that log and
`folio-mvp-decision-log.md`. The count of 92 is the positive control that the query works.

**Why it matters.** Had the name been trusted, the lead would have been told to re-ground from prior
rulings that do not exist, and would have reported finding none — or worse, inferred Epic 8's rulings
applied here.

**Consequences.** The lead grounds from scratch for this program. The misleading log name is left alone:
renaming a 265 KB committed artifact mid-run is churn, and this entry is the pointer.

---

### D-000.3 — The build order must be established by measurement, not by number
**Orchestrator decision**, pending the lead's grounding report.

**Verdict.** Before any spec is written, establish **what is already true** of epics 11–14 in the shipped
codebase. Stories found already satisfied are closed as such with evidence; stories whose premise has been
invalidated are re-scoped at the plan gate.

**Situation.** This repository built epics **15, 16 and 17 before 11–14**. All 23 stories in 11–14 are
`backlog`, and their specs were written against a codebase that has since moved a long way. Three
collisions are visible without looking hard:

- **Epic 11 is font weight and slope** — "the shipped families gain a weighted and a sloped face", "the
  engine resolves a face from the declared weight and slope". **Epic 16 spent itself on the shipped face
  set** and added Roboto as a fourth face with a machine-checked byte-identity rule. 11.1's premise has to
  be re-measured against `Shipped()` as it stands.
- **11.3, "the canvas paints the weight the engine resolved"**, touches the Bold/Italic controls that
  changed on **2026-09-04**: pressing a pressed toggle now clears rather than writing `bold: false`.
- **Epic 14 is "the designer's controls read as one product"** — ten stories about the property panel, the
  data tab and the table editor, written before Epics 16 and 17 reworked the property rows, the typography
  panel, the font browser, and the number-field and prose-field behaviour.

**In simple terms.** These four epics are a set of instructions for a building that has since had three
more floors added. Some instructions now describe work already done; others describe work that would
knock out a wall something new is resting on. Reading the instructions first, and the building second, is
how a wall gets knocked out.

**Why this wins.** The alternative — writing specs in numeric order and letting each build discover the
drift — is exactly the failure this run has already paid for twice at single-story scale: Story 17.5's
Code Map went stale in under a day, and its Design Notes described a hazard that had been fixed hours
earlier. At four epics' scale that cost is multiplied by 23.

**Consequences.** The first work of this run is a survey, not a spec. Its findings are logged per story.

**How we'd know it was wrong.** If the survey finds nothing stale in any of the 23 stories, it was
over-cautious — a cheap outcome, and still worth knowing.

---

## Lead Grounding

*Filed verbatim from the engineering lead's grounding report, 2026-09-05, with one orchestrator correction
recorded immediately below it. The lead grounded once; it does not re-read the spine, the ADRs, the epics
doc or this log on later resumes.*

### ORCHESTRATOR CORRECTION TO THE REPORT'S STATED BASELINE — read this before the report

**The report says "grounded against HEAD `8a9e297`, working tree clean". Both halves are false, and the
error runs in the direction that flatters the report rather than the one that undermines it.**

Measured: HEAD at grounding was **`c366deb`**. `8a9e297` is dated **2026-09-03** — 48 commits and two days
stale, and it is the HEAD this session STARTED at, which is almost certainly where the value came from.
The tree was **not clean**: six files were modified.

**The report's code observations are drawn from the live working tree and are therefore MORE current than
its stated baseline, not less.** It cites `property-prose-height.test.ts` asserting `resize: none` and a
6–8px handle hit height. Neither exists at `8a9e297`, neither exists at `c366deb`, and neither is
committed anywhere: both are **Story 17.5's uncommitted, still-being-built work**. The lead read a
construction site and reported it as the building.

**Every claim resting on 17.5 is therefore PROVISIONAL** until that story lands and is reviewed — at
minimum the `property-prose-height.test.ts` source-text contract and the CONTENT row's drag grip. The lead
has been asked to name the rest.

This does not discredit the report; its substance is unusually well measured, and it is filed in full
below. It does mean **anyone re-deriving from the stated baseline would get different answers**, which is
exactly the staleness D-000.3 exists to prevent, arriving inside the document written to prevent it.

### The report, verbatim

**Sources read.** ARCHITECTURE-SPINE.md (728 lines, AD-1..AD-26 in full); epics.md §Epic 11–14 (lines
3887–5045); epic-11-14-decision-log.md (D-000.1/.2/.3); sprint-status.yaml; the lead's own agent memory for
folio. Plus the code cited below, read directly. Two parallel surveys of `folio-designer/` (preview,
inspector) whose load-bearing claims were re-measured before use; three of their corrections to the epics
doc were verified personally and are marked as such.

**One naming trap before anything else.** "CAP-2" names two different things in this repo.
`_bmad-output/specs/spec-folio/SPEC.md:43` — *data binding from sample data*.
`_bmad-output/specs/spec-fonts/SPEC.md:40` — *embedded font assets, "a `.folio` carries its faces"*. The
run brief means the second. Every ruling says `CAP-2(fonts)` to avoid inheriting the wrong one.

#### 1. The invariants this run must not break

**I-1. No story may make a rendered byte a function of the build environment or the host.** (AD-1, AD-2,
AD-3, AD-8, AD-22.) A new shipped face must be committed output, not generated at build time; a new number
reaching the PDF takes one of AD-3's exactly two representations; no `float64` under `internal/`; no
`time`/`os`/`math/rand`/transcendental import enters `internal/`. Epic 11 is the only one of the four that
can reach a render byte at all, and 11.1 is the only story in the run that adds a binary asset.

**I-2. A face is a face, or it does not exist.** Synthetic bold and synthetic oblique are forbidden at emit
time *and* on the canvas. `internal/text/shape.go:90` records that the vendor exposes `SetSyntheticBold`
and `SetSyntheticSlant` and that **none of them is called**. Epic 11 must not change that, and 11.3 must
remove the browser's own faking rather than tidy it.

**I-3. The browser never measures — and the ban is machine-enforced repo-wide, not by convention.**
(AD-17.) `canvas-authority-contract.test.ts:18–70` scans every `.ts`/`.tsx`/`.css` under `src/` plus `e2e/`
and forbids `measureText`, `getBoundingClientRect`/`getClientRects`, `offset*`/`client*`/`scroll*`,
`offsetX/Y`, `ResizeObserver`, `getComputedStyle`, `document.fonts`, `new FontFace`, `devicePixelRatio`,
`Range`/`getSelection`, and a set of CSS wrapping property/value pairs. Exactly four exceptions exist, each
scoped to a named owner (`:355–370`, `:340–356`): `src/preview/**` may spell `scroll(Width|Height|Left|Top)`
**and nothing else**; `App.tsx` may spell a pointer coordinate inside `placementPoint` only;
`embedded-face-registry.ts` may register faces inside `registerCarriedFaces`; `canvas-font-stack.test.ts`
may name two spellings as fixtures. **`clientWidth` sits in the same regex line as `scroll*` and is NOT
covered by the preview exception.** Any Epic 13 or 14 story that wants a container's pixel size is blocked
here, by design.

**I-4. The engine owns the document; the UI holds a snapshot and sends commands.** (AD-15.) No TypeScript
model of a `.folio`. No local uncommitted buffer standing in for the document. Transient interaction state
lives in the UI and never enters the document. Binds 12.5 (band drag), 14.7 (any Cancel/Apply), 14.9 (the
canvas must paint tables from the engine's projection, never a browser-side model of the columns).

**I-5. A control the panel declines to author must not delete the value the document already carries.**
Explicit in 14.2 and 14.4's last ACs, and it generalises: the panel narrowing its vocabulary is never a
document migration. No story in this run may normalise a document on selection.

**I-6. A document that declares none of the new things must hash identically.** (AD-21.) Checked: across
the 29-fixture golden corpus, `"bold"` and `"italic"` occur **zero** times (`grep -arln` over `fixtures/`,
exit 1; positive control, `"fontFamily"` occurs in **23** fixture files, exit 0). `"padding"` occurs in
**four** fixtures — `statement-1/5/20/50/input.folio` — and in every one of them on the element
`"id": "e8", "type": "table"` at line 18–19, i.e. **no corpus document declares padding on a non-table
element**. So 11.1's, 11.2's and 12.4's identical-hash ACs are all satisfiable against the corpus as it
stands, and the corpus is a real witness rather than a formality.

**I-7. The licence boundary is CI-enforced and every redistributed asset travels with its terms.** (AD-26.)
Any face Epic 11 adds needs its own `NOTICE.md` with upstream release URL and source sha256 plus OFL text,
and must be accounted for by lint's `fonts-asset-unaccounted`/`fonts-asset-missing` guard.

**I-8. Diagnostics are one type on one channel, from a closed additive registry.** (AD-14.) 11.2's "the
fallback is stated, not silent" and 12.1's refusals mint or reuse codes; codes are additive only.

**I-9. `Render`'s error contract is not negotiable from the designer.** 13.4 changes what a *preview with
no data supplied* does; it must not change `BINDING_PATH_ABSENT`'s severity in the engine, and `folio-go`
rendering that template with absent data must still fail exactly as today.

#### 2. Per epic: what the stories assume, and what is stale

Every one of the 23 was measured. **None is wholly done.** But the epics doc is wrong at *AC* granularity in
at least nine places, so the survey D-000.3 calls for must be run **per acceptance criterion, not per
story**.

**Epic 11 — bold and italic.**
*Still true.* `Style.Bold`/`Style.Italic` are stored, projected and consumed by no producer.
`git grep "\.Bold\b" -- '*.go'` returns hits in exactly four non-test files: `internal/template/parse_bands.go:667`
(parse), `internal/template/serialize.go:442` (serialize), `page_setup.go:1785–1789` (canvas projection),
`component_commands.go:1225–1238` (the edit command). **`render.go` and `table_render.go` are not among
them.** Positive control on the same population: `FontFamily` returns `render.go:3` and `table_render.go:8`,
so the render path does read style — it just never reads weight. The canvas fake is live and exact:
`App.tsx:3005` sets `'--text-font-weight': component.bold ? 700 : 400, '--text-font-style': component.italic
? 'italic' : 'normal'`. 11.3's premise stands verbatim.

*Stale — 11.1's arithmetic.* The epic says "three families × three new instances". `folio-go/fonts/fonts.go:80–87`
`Shipped()` now returns **four**: `"Noto Sans"`, `"Noto Sans Thai"`, `"Noto Sans SC"`, `"Roboto"` — all
Regular. The cost is 4 × 3, not 3 × 3.

*Invalidated — 11.1's mechanism, for one quarter of the set.* `tools/fontgen/instance_faces.py:117`
`UPSTREAM` is a hardcoded **three-entry** list. Roboto is not in it and `fonts.go:56–67` says why: upstream
publishes a static TTF directly, so the shipped Roboto is a byte-for-byte copy of the designer catalogue's
face, not a derivation. `fonts_test.go:35` `TestShippedRobotoMatchesDesignerCatalogue` makes "there is
exactly one Roboto" machine-checked. A Roboto Bold cannot ride 11.1's stated mechanism.

*Satisfied, and better than 11.1 expects.* `folio-go/shipped_faces_test.go` is a spec-table guard written
*in anticipation of this story*: one `shippedFaceSpec` row per shipped face asserting name[1], name[2],
name[6], and `OS/2.usWeightClass`, with `TestShippedSpecCoversEverythingShipped` failing in **both**
directions. `:37` says outright the design exists "so the future Bold story adds faces and inherits" the
checks. 11.1 must add a spec row per face and must not weaken the `Subfamily: "Regular"` assertions into a
denylist (`:74–80` warns against exactly that).

*The load-bearing gap 11.2 does not know about.* There is no weight axis anywhere in the format or the
public API. `folio-go/fontset.go:20` — `type FontSet map[string][]byte`, face **name** to bytes.
`internal/template/model.go:197` — `type Fonts map[string][]FontChainEntry`, and `:164–186` an entry is
**either** a shipped face name **or** an embedded asset key. `render.go:1148` `fontChain` returns
`[]string`. So "resolve to the face carrying that weight and slope" has no expression today, and the
embedded arm has no *name* to derive one from.

*Two more collisions Epic 16 created.* (a) `font-browser-model.ts:379` renders the string
`"… faces · one upright Regular each, no bold or italic"` to the author. (b) `starter.folio:9` — every new
document opens on `"Roboto"`, so the first family an author will bold is the one whose bold is hardest to
obtain.

*Note for 11.3's third AC.* `component_commands.go:1223–1229`, a `clear` on `bold`/`italic` writes
`template.Presence[bool]{}` (absent), not `{Set: true, Value: false}`, and `cleanupEmptyStyle` (`:1287–1295`)
then drops an emptied `style` block entirely. "Off" and "never set" are the same document state.

**Epic 12 — the inspector reaches the engine that is already there.**
*Wholly true.* The complete command registry is `component_commands.go:60–105`, 24 kinds. **None writes a
band height, a locale, a UTC offset, a table `headerHeight`, `altRowBackground`, or `headerStyle`.** The
values are all real on the render path: `Band.Height` at `render.go:305–309` and `internal/layout/band.go:76`;
`HeaderHeight`/`AltRowBackground`/`HeaderStyle` at `table_render.go:667–740`, cascaded by
`resolveHeaderStyle` (`:311–324`); `utcOffset` by `internal/expr/formatcontext.go:23`. The designer's only
`locale` matches are four `localeCompare` calls — a false positive, checked.

*A collision 12.1 must resolve and does not name.* 12.1's fourth AC says a band height reduced below its
content "is accepted and the overflow is clipped". But `component_commands.go:1911–1919` `containComponent`
**refuses** any component whose `y > band.Height || height > band.Height - y`, and it is called from eleven
sites.

*A shape constraint 12.1 must decide.* `page_setup.go:1913` gates on `len(raw) != 7 || …` — a strict-arity
check. Extending the pageSetup command changes that arity; a new command does not. The anchor for both
panels is `App.tsx:1607`.

*12.4's premise verified, including the doc claim.* `folio-format.md:457` — `| padding | 0 on all four
edges |`. The command layer accepts all four keys on any component (`component_commands.go:1040–1043`), the
projection carries all four (`page_setup.go:271–274`, `:1848–1876`), and the only render-path consumer is
the table cell cascade (`table_render.go:380–383`).

*12.5's anchor partly exists.* `App.tsx:1442`: each band is a `<section className="page-band…" tabIndex={0}
aria-label=…>` whose first child is `<span>{bandName(band.name)}</span>`. The label is there and the band is
already focusable. But there is no draggable control: `/usr/bin/grep -a "band-tab\|bandTab"` over App.tsx
returns nothing, against a positive control of 27 occurrences of `band` in the same file.

**Epic 13 — the preview.**
*Every one of the ten premises in the epic's opening paragraphs is still true.* `src/preview/` was untouched
by Epics 15, 16 and 17 (5 commits, newest 2026-08-30). No PDF export path anywhere in `folio-designer/src`
(`savePdf|export pdf|download pdf` and `\.pdf['"]` both return nothing; positive control, `.folio` matches
24 files) — the download tier at `file/input-download.ts:49` hardcodes `type: 'application/json'`. The
palette rail at `App.tsx:1454` is a **sibling** of the design/preview ternary, so it renders in preview mode.
The whole evidence surface is one `<p className="preview-evidence">` at `App.tsx:1475`. Zoom is
`Math.max(0.5, scale - 0.1)` / `Math.min(2, scale + 0.1)` at `preview/pdf-viewer.tsx:102–104`, both readouts
`<output>`. `App.css:358` `.pdf-preview-scroll` has no height cap, so `onScroll` at `pdf-viewer.tsx:106`
cannot fire and the restore effect at `:91–95` is dead. The sample-data gate is `App.tsx:471` —
`if (!sampleDataRef.current) { setPreviewStatus('idle'); return }`.

*Three things 13.x will hit that the epic does not mention.* **13.2's fit-width is blocked at the contract,
not merely unimplemented** — every route to a container's width is prohibited, and `src/preview/**`'s
exception covers `scroll*` only. The existing precedent is `pdf-viewer.tsx:59–82`: a **constant**
`previewOversample = 2`, chosen precisely because `devicePixelRatio` is banned. **13.2 must fix
`pdf-viewer.tsx:89` in the same story** — the render effect lists the whole `state` object in its dependency
array; harmless only because the scroll handler is dead, and fixing the CSS height cap alone turns a dead
path pathological. **13.3 and 14.6 both hit hard markup contracts** —
`preview-authority-contract.test.ts:17–24` pins literal App.tsx substrings and forbids
`createObjectURL`/`revokeObjectURL`/`<iframe`/`<embed`/`fetch(` inside `pdf-viewer.tsx`, and
`design-contract.test.ts` asserts `App.css` contains **exactly one** `@media` query.

**Epic 14 — the designer's controls.**
*True, and verified directly.* `bindComponentScalar` refuses non-text at `component_commands.go:585–587`
while the BINDING section at `App.tsx:1773` is rendered with no type predicate, unlike CONTENT/TYPOGRAPHY/
IMAGE/TABLE beside it — 14.4 stands. Placement does not select: `App.tsx:710–714` `place` and `:724–729`
`placeInBand` both call `clearInteraction()` and commit without the optional `after` callback, and no
`setSelected` appears in either path — 14.3 stands, and `duplicateSelection` (`:733`) has the same defect,
which 14.3 does not name. The canvas projection carries only `tableBind?: string` (`engine-protocol.ts:179`,
guarded at `:376`) and both canvas paths end in `component.type === 'table' ? 'Table' : ''` (`App.tsx:2893`,
`:3067`) — 14.9 stands. `updateTableColumnBinding` is dispatched at `component_commands.go:90` and its sole
call site is the modal's `onBlur` — 14.10 stands.

*Four premises are wrong, two of which dissolve a ruling the epic demands.* **14.7's millimetres do not
exist in the product** — `git grep "(mm)\|millimet\| mm\b"` over `folio-designer/src` returns nothing, exit
1; positive control, `(pt)` matches 5 times in App.tsx and once in TableEditor.tsx; the mockup does carry
`mm` (4 hits in `TableEditor.dc.html`). **14.7's Cancel/Apply is already decided by the shipped code** —
`TableEditor.tsx` has no Apply and no Cancel (`grep -ac "Apply\|Cancel"` → 0, exit 1; positive control
`onClose` → 4); every cell commits on blur (`:65`, `:68`). AD-15 already won. **14.2's border controls do not
"mean nothing" on a Line** — `folio-go/element_box.go:124–157` `elementBoxDeclaration` is kind-agnostic and
`borderPaints` returns true for a present, non-null border with a non-empty edge set. The defect is
incoherence, not inertness. **14.6 is a smaller build than it reads** — the `[]` marker exists
(`sample-data.ts:98`), a selection status line exists (`DataPanel.tsx:34`), and the roving tab stop with full
arrow/Home/End/Enter navigation is **already built and pinned** (`DataPanel.tsx:48–62`, asserted by
`DataPanel.test.tsx:54` and `:72`).

*What Epics 16 and 17 put in Epic 14's path.* The family control now names **two** source groups
(`App.tsx:2596–2599`), pinned by seven named cases in `App.test.tsx:1671–1856`. Number fields step by arrow
key in integer thousandths (`App.tsx:2155`, 20 cases at `App.test.tsx:2383`). Size and leading show the
engine's projected defaults (`App.tsx:1633`, 13 cases at `:2918`). The CONTENT row is a single debounced
prose textarea with a drag grip, guarded by `property-prose-height.test.ts` — **PROVISIONAL, see the
orchestrator correction above**. `design-contract.test.ts` pins the literal
`.tree-item:focus-visible { outline: 2px solid var(--color-select); outline-offset: -2px; }`;
`--color-select` is `tokens.css:7`, `#58A6C4`, matching 14.5's expectation.

*Two record discrepancies.* `sprint-status.yaml` carries
`15-0-a-catalogue-face-arrives-when-it-is-picked: backlog`, but `15.0` does not appear anywhere in
`epics.md` (grep exit 1; positive control, `15.2a` matches). And Epics 15, 16 and 17 are all `in-progress`.

#### 3. Recommended build order

**Gate, before any spec is written.** (a) The owner rules Epic 11's branch — nothing in Epic 11 can be
specced without it, *and it is also an Epic 14 input*: if bold and italic are retired, Epic 11 collapses
into a panel change belonging inside 14.4. (b) The AD-17 corpus scan's masking is settled. (c) 15.2a's
position is settled.

**Wave A** — `15.2a` (pulled forward from Epic 15) and the AD-17 scan repair. Both cheap now and expensive
after twenty stories of new controls.

**Wave B — Epic 12.** `12.4` first: ruling-first, cheap, touches no panel, and settles what `padding` means
before 14.8 writes a cell-padding control whose meaning depends on the answer. Then `12.1` → `12.2` (same
`PAGE SETUP` panel; one rework rather than two). Then `12.3`, a **hard prerequisite of 14.8**. Then `12.5`,
which commits through 12.1's command.

**Wave C — Epic 13.** `13.4` first, against numeric order: the gate at `App.tsx:471` blocks *every* preview
without a JSON file. Then `13.1` (Save PDF). Then `13.2` (needs the fit-width ruling; must fix
`pdf-viewer.tsx:89` in the same commit). Then `13.3`, the large restructure. Then `13.5`.

**Wave D — Epic 14, and not before the open 17.x stories land.** Within it: `14.1` first (the rule is the
deliverable). `14.5` next. Then `14.2` and `14.4` — 14.4 before 14.6. Then `14.3` and `14.9` once 17.x
clears. Then `14.6`. Then `14.7` (after 12.3, and after 17.3/17.4). Then `14.8` (hard-depends on 12.3). Then
`14.10` last.

**Epic 11 slots by its branch.** If *realize*: `11.1` can run parallel to Wave B or C; `11.2` follows;
`11.3` waits for 17.x. Strict chain. If *retire*: there is no Epic 11; there is one AC inside 14.4.

**Nine acceptance criteria are already satisfied or falsified** — survey these per AC before speccing:
14.6's roving tab stop (**built and pinned**); 14.6's `[]` marker (**present**); 14.6's status line
(**present**); 14.7's Cancel/Apply (**already decided by shipped code**); 14.7's millimetre unit (**does not
exist**); 14.2's inert border (**false — it paints**); 12.5's focusable band boundary (**already
`tabIndex={0}`**); 13.3's diagnostics count (**partly present**); 11.1's guard apparatus (**inherited
free**).

#### 4. The decisions this run must make

**D-A — Epic 11: realize bold and italic, or retire them.** Owner's. Four families not three; the fourth
cannot ride the stated derivation mechanism; a shipped user-facing string already tells authors there is no
bold; every new document opens on that fourth family; and *retire* folds into 14.4 and changes what Epic 14
is.

**D-B — if realized, how weight and slope reach a face.** No weight axis in the public API
(`FontSet map[string][]byte`), the format (`Fonts map[string][]FontChainEntry`), or the resolution seam
(`fontChain` → `[]string`). Three routes: a naming convention over FontSet keys; structure on the chain
entry (a format change under AD-9, breaking under AD-22); or a shape change to `FontSet` (a public API
break). The embedded arm has no name to derive from at all.

**D-C — the AD-17 corpus scan is standing red, and Epics 13 and 14 are the lane it is masking.** The single
violation is `e2e/e9-5-border-no-ink.spec.ts:66` — one e2e spec, not production code. While it is red **no
new AD-17 violation anywhere in `src/` or `e2e/` can make anything go redder**, and this run is fifteen
stories of designer UI.

**D-D — Story 15.2a versus every new numeric command in Epics 12 and 14.**
`component-property-command.ts:58–62` — `rawNumberLiteral` returns the author's string **unquoted and
unvalidated** into command JSON, and `pointFields` at `:7` routes ten fields through it. Epic 12 adds band
heights and a header height; Epic 14 rebuilds the width column and adds cell padding.

**A fifth cluster, smaller.** 12.4's padding fork; 13.3's "Matches native render" wording; 14.6's
pick-commits-immediately versus an explicit Connect button; 14.7's transaction model and unit.

**Contradicted documents, each verified by locating a definition or by a negative paired with a positive
control:** 11.1's face count and Roboto's derivation mechanism; 14.7's millimetres; 14.7's Cancel/Apply;
14.2's inert border; 14.6's three already-built ACs; and the epics doc's claim of "exactly three readers" of
`Style.Bold` (there are four non-test files, and the fourth is the parser).

---

### D-A — Epic 11 realizes bold and italic for the Latin and Thai families only
**Owner decision**, taken at the terminal 2026-09-05, on the payload arithmetic.

**Verdict.** Noto Sans, Noto Sans Thai and Roboto gain **bold, italic and bold-italic**. **Noto Sans SC
stays Regular.** Epic 11 exists and is smaller than written: **three families, nine new faces**, not four
families and twelve.

**The situation.** Bold and italic are stored (`Style.Bold`/`Style.Italic`), shown in the panel, and faked
by the browser at `App.tsx` (`--text-font-weight: 700`) — and read by nothing on the render path. A bold
document prints unbold. `render.go` and `table_render.go` are not among the four non-test files that read
`.Bold`; the positive control on the same population is `FontFamily`, which *is* read by both. So the
question was never "fix a bug"; it was "make this real, or stop offering it".

**In simple terms.** The B button today is a light switch wired to nothing. You can press it, it lights up,
the panel remembers you pressed it — and the printed page is identical either way. There were only ever two
honest ends to that: run the wire, or take the switch off the wall.

**What decided it: the arithmetic, not the principle.** The shipped faces total **11.1 MB**, and **Noto
Sans SC alone is 10.35 MB** of that (Noto Sans 631 KB, Roboto 348 KB, Noto Sans Thai 47 KB — measured, not
estimated). Three instances of the CJK face is roughly **31 MB**, taking the payload every user downloads,
offline, from 11 MB to about **45 MB**. The Latin-and-Thai set adds about **3 MB**. Epic 16 spent an entire
story arguing over a single 348 KB face; this was two orders of magnitude larger.

**Options considered.** *All four families* — complete and consistent, and ~31 MB of that consistency buys
bold for a script whose upstream generally publishes no italic at all. *Retire both* — folds Epic 11 into
one acceptance criterion in 14.4, adds nothing to the payload, and makes an already-shipped user-facing
string true; rejected because a document designer without bold is a hard sell. *Bold only, no italic* —
halves the weight-resolution design since slope never needs an expression; rejected as too small a
capability for the work either way.

**Why this one wins.** It buys the capability where it is nearly free and declines it where it is
ruinous, and — this is the part that makes it coherent rather than merely cheap — **Epic 11 already has to
build the machinery for "this family has no face at this weight."** 11.2's requirement that the fallback be
*stated, not silent* was written for edge cases. Under this ruling Noto Sans SC becomes a permanent,
shipped instance of exactly that case, so the CJK path exercises the honest-fallback machinery as its
ordinary behaviour rather than as a corner nobody hits.

**Consequences.**
- **I-2 is untouched and load-bearing**: no synthetic bold, on the canvas or at emit. `internal/text/shape.go:90`
  records that the vendor exposes `SetSyntheticBold`/`SetSyntheticSlant` and that neither is called. Bold on
  CJK text must *say so*, never fake it.
- **11.3's B control needs three states over two document states.** A `clear` on `bold` writes
  `Presence[bool]{}` — absent, not `false` — and `cleanupEmptyStyle` then drops an emptied style block, so
  "off" and "never set" are the same bytes.
- **11.1's mechanism fork inverts.** Roboto was the family that could not ride `instance_faces.py`'s
  three-entry `UPSTREAM` derivation. Roboto is now *in* scope and Noto Sans SC is *out*, so the derivable
  set and the in-scope set no longer coincide. 11.1 must either carry
  `TestShippedRobotoMatchesDesignerCatalogue`'s one-cut discipline into the bold and italic cuts, or rule
  explicitly that the one-Roboto rule is scoped to Regular. **In the spec, not at implementation.**
- **A shipped string becomes false.** `font-browser-model.ts:379` tells authors
  *"one upright Regular each, no bold or italic"*. Realizing bold makes that a lie the product tells; it
  must change in the same epic.
- **The starter template's default family is Roboto**, so the first family an author bolds is in scope. Good.

**How we'd know it was wrong.** Authors setting Chinese text and reporting that bold "does nothing" — which
would mean the stated-fallback surface is not carrying its weight, and the honest answer would be a clearer
statement rather than the 31 MB.

---

### D-000.4 — Every Code Map in this run cites a symbol and a count, and re-derives the line at its plan gate
**Orchestrator decision**, forced by the lead's own grounding report.

**Verdict.** Line numbers are not addresses in this repository. A Code Map anchor must name a **symbol** and
carry a **count** (occurrences, with a positive control); the line number is re-derived at the story's own
plan gate and never inherited from an earlier document.

**Situation.** The lead's grounding report cited `App.tsx` line numbers that were wrong for **both** the
committed tree and the working tree — because `App.tsx` grew about forty lines *while the grounding was
being written*. `function bandName` was filed at 2850; it is 2790 at HEAD and 2890 in the tree. The content
claims behind those numbers all survived re-verification by symbol; the addresses did not.

**In simple terms.** Citing a line number in a file under active edit is like giving directions by counting
houses from the corner while someone is building a new one halfway down the street. The house is still
there; the count is not.

**Why this wins.** This run has now paid for line-number staleness **three times in two days** — Story
17.5's Code Map, its Design Notes describing a hazard fixed hours earlier, and now the grounding document
written expressly to prevent both. The cost of a symbol-plus-count anchor is one extra grep at the plan
gate. The cost of a stale address is a re-planning cycle.

**Consequences.** Applies to every spec in this run, and to the survey D-000.3 mandates — which is run
**per acceptance criterion and by symbol**, not per story and by line.

---

### D-000.5 — "Check the author" does not discriminate in this repository
**Orchestrator decision**, and a correction to guidance I issued myself.

**Verdict.** A commit's author field cannot tell you whether the orchestrator, an agent, or the owner made
it. Establish authorship by the **timestamp band** against commits already known to be yours, by reading
the message as a deliberate act, and by **asking** — never by author alone.

**Situation.** After Story 17.5's builder wrongly reported that a subagent had committed against
instruction, I told it the rule was "check the author and the message before concluding anything". It
checked, and came back with the flaw: `1a56007`, `c366deb`, `f0fa858` and `d2f2a1e` all carry the same
author *and* committer, because that is the machine's single configured `user.email`. Every commit made
here — mine, an agent's, the owner's — is identical in that field.

**In simple terms.** I told it to identify the driver by reading the licence plate, on a fleet where every
car has the same plate.

**Why the corrected rule works.** `00:18:59 → 00:33:54 → 00:35:00` is one continuous band of orchestrator
commits, and `d2f2a1e` sits inside it. That is evidence; the author field is not. The version leaning on
author would have failed **silently**, which is worse than the original error — it would have returned a
confident wrong answer instead of an obviously unusable one.

**Consequences.** Any agent told to watch for unexpected commits gets the timestamp-band rule, not the
author rule. And the sequence on seeing HEAD move is: confirm nothing was pushed, confirm the file list,
**re-run verification against the committed tree because HEAD is now what ships**, then report the move as
an *observation* and ask — never offer a `git reset` before the diagnosis comes back.

**What the builder did right, and it is the reason this was caught at all.** It re-measured
`git rev-parse HEAD` rather than trusting `git status`, which read clean and would have said nothing. The
instinct was correct; only the inference from it was wrong.

---

### D-000.6 — The AD-17 scan's `e2e/**` population is a default, not a recorded ruling
**Orchestrator decision**, settling the one condition the lead flagged as able to flip D-C.

**Verdict.** The lead's **preferred** branch stands: repair D-C with a **named-owner exception**, not by
rewriting the assertion. Its stated flip condition does not hold.

**Situation.** D-C's ruling prefers narrowing the AD-17 scan over changing
`e2e/e9-5-border-no-ink.spec.ts`, but flagged one assumption that would reverse it: *if* the scan's
inclusion of `e2e/**` was a deliberate, recorded ruling, then narrowing it reopens someone's decision.

**Measured.** The population is built inside the test file itself —
`canvas-authority-contract.test.ts:11-13`, a `readdirSync` over `e2e/` with a `.ts|.tsx` filter, sitting
beside the identical constructions for `production` and `tests`. It is a construction choice in code, with
no comment justifying the e2e arm specifically. Three decision logs mention `canvas-authority-contract`;
reading the surrounding context in each, **none discusses why `e2e/` is scanned**. Positive control that
the query works: those same greps returned substantive material about the scan — its comment-stripping
requirement, and that it was red-proved in both directions — so the silence about `e2e/` is real silence
and not a failed search.

**Consequences.** The Wave A story may narrow the scan by a named-owner exception without reopening a
prior decision. **The four existing exceptions are the shape to copy**: scoped to one named block in one
named file, asserted present by the test itself so the carve-out cannot outlive its reason, rewriting only
the offending spelling — never a directory-wide pass.

**How we'd know it was wrong.** A decision surfacing later that recorded the e2e population deliberately;
this entry is the pointer for whoever finds it.

---

### D-B — Weight and slope are carried by the font chain entry, on both arms
**Engineering lead's ruling**, accepted with one correction to its supporting count (below). **No owner
decision required, and that is the main reason this route was taken.**

**Verdict.** The `(family, weight, slope) → face` mapping is declared on the **`FontChainEntry`** and
nowhere else. An entry gains optional style-variant siblings — a face-name variant on the shipped arm, an
asset-key variant on the embedded arm — using the existing `Presence` idiom, absent by default.
`folio.FontSet` keeps its shape. **The public API does not change.**

**Why this is not a new format axis.** `internal/template/serialize.go:204` `writeFontChain` **already**
emits a chain entry as either a bare JSON string or an object — verified by reading the function, not its
comment: an embedded entry goes through `writeObject`, a face entry through `appendJSONString`. So a
document with no variants serialises byte-for-byte as it does today, and **AD-21's corpus-identity
requirement is satisfied by construction rather than by care**: no corpus document declares bold, every
corpus chain entry stays a bare string, and bare strings are emitted unchanged.

**Resolution order — the thing the epic never settled and implementation would have guessed.** Per-rune
**coverage decides the entry; style is resolved within that entry only.** A covering entry with no face at
the requested weight renders that rune in **its own Regular** with a stated diagnostic — never in a
*different* entry's bold. Falling down the chain to find a bold face would change the **typeface** to keep
the **weight**, which is the worse substitution, and would make a Thai run silently change family because a
CJK run asked for bold.

**In simple terms.** You ask for bold. The Latin words find Roboto Bold. The Chinese characters are only
carried by a family that has no bold, so they print in that family's Regular and the product *says so*.
The alternative — hunting down the chain until something bold turns up — would print the Chinese in
whatever face happened to have a bold, which is a different typeface entirely. Losing the weight is a
smaller lie than changing the typeface.

**This is what makes D-A's exclusion of Noto Sans SC safe.** A mixed line under bold prints Roboto Bold and
Noto Sans SC Regular, with the shortfall named. Absence is a **first-class result** the caller must handle,
not a nil-shaped silence that defaults to Regular — and because CJK is now a permanent shipped instance of
it, the test proving that arm belongs in 11.2's main body, not an edge-case file.

**Foreclosed explicitly, so no spec reopens them.** *Naming convention over `FontSet` keys* — dead: it
would silently reinterpret a caller's already-legal key `"X Bold"`, cannot express absence (a missing key
is indistinguishable from a typo), and cannot serve the embedded arm at all, whose keys are hex SHA-256
with no name to append to. *A shape change to `FontSet`* — not this epic: it is an irreversible change to
the product's primary public input taken under release-schedule pressure, and it is free only if Epic 11
lands before Story 15.3 cuts the tag, which would make a font epic a release blocker. *Synthetic bold or
oblique* — dead, per I-2.

**ORCHESTRATOR CORRECTION TO THE RULING'S SUPPORTING COUNT.** The lead wrote that `Render`, `RenderTo` and
`Validate` are "the **only three** channels by which caller-supplied font bytes reach the engine".
Measured: exported functions in `folio-go` taking a `FontSet` also include `CanvasWithTextPaint`
(`page_setup.go:784`) and `PreviewIdentity` (`preview_identity.go:14`) — **at least five, not three.**
The ruling is unaffected and in fact **strengthened**: the argument was that no channel has room for a
second font-shaped argument, and more channels means more break sites for the foreclosed `FontSet` route.
Recorded because a count stated as exhaustive is load-bearing for whoever re-derives this later.

**Guardrails carried into the specs.**
- **The no-synthetic contract test is an ALLOWLIST of permitted paint properties, not a denylist.** A
  denylist of `font-weight`/`font-style` — which 11.3's AC as written implies — cannot see
  `font-synthesis`, `-webkit-text-stroke`, `text-shadow`, `paint-order`, or `transform: skewX()`, and a
  skew is exactly how an oblique gets faked. **Second independent instance of this defect class in this
  run** (17.5's review found a "paints nothing" denylist permitting `border-top` and `box-shadow`), which
  is what makes it a standing rule: *a guard that enumerates what is forbidden cannot see what nobody
  thought to forbid.*
- **11.1 has two mechanisms, not one.** Re-derived: `tools/fontgen/instance_faces.py`'s `UPSTREAM` holds
  exactly three keys — the three Noto families. So of D-A's in-scope set, **Noto Sans and Noto Sans Thai
  are derivable and Roboto is not**, and the one derivable entry now out of scope is the CJK face.
  `TestShippedRobotoMatchesDesignerCatalogue`'s one-cut discipline **extends to the new cuts** — do not
  scope it to Regular, or two cuts of Roboto Bold under one name reproduce, at the weight axis, exactly the
  defect that rule exists to prevent.
- **No document is migrated on open.** A pre-epic document correctly reports "this family has no bold
  face"; the remedy is an explicit author action, never a rewrite on load or selection (I-5).
- **The canvas is told the resolved outcome, not the requested flag** — a projection field, not a format
  field. The browser cannot discover that a family has no bold face without measuring (AD-17) or holding a
  second model of the font set (AD-15).
- **The B control is three states over two document states**, the third derived from the projection and
  never stored. It must not write `bold: false` to mean "off" — that would reverse the 2026-09-04 toggle
  change and dirty every document whose author touches it.

**The assumption that would qualify this.** The lead assumes the designer can write chain variants through
the existing font-chain command family without a new command kind. If not, 11.2 and 11.3 acquire a
command-surface change and fall behind Story 15.2a in Wave A. **Check at 11.2's plan gate.**

---

### D-000.7 — Every count is marked CLOSED or SAMPLED, with its population named
**Orchestrator decision**, forced by two errors of the same shape in one day.

**Verdict.** Any count stated in this run — by the lead, a builder, or me — says whether it was **CLOSED**
(the population was enumerated, with the boundary of the enumeration stated) or **SAMPLED** (do not build
on it without re-deriving). A number whose provenance cannot be told is a number the next reader must
re-measure.

**Situation.** The lead's D-B ruling said `Render`, `RenderTo` and `Validate` were "the **only three**
channels by which caller-supplied font bytes reach the engine". Measured: five, plus a producer. Asked how
it got there, it gave the mechanism rather than an apology, and the mechanism is the valuable part: it ran
a grep that **named the four functions it expected to find**, then reported the result as a sweep of the
population. **The query could not have found the two it missed no matter how many existed.**

**In simple terms.** It is the difference between "I counted everyone in the room" and "I called out four
names and four people answered". Both produce a number. Only one of them is a count. The tell was sitting
in the pattern: it described the answers, not the population.

**Why this is worth a standing rule rather than a note.** This is the second instance in two days of a
measurement that was *shaped like* evidence and was not — the first being the grounding report's baseline,
which reproduced the session's opening `gitStatus` block without re-measuring. Both were confidently
stated, both survived their own author's review, and both were caught only because someone downstream
re-ran them. A rule that makes provenance explicit at the point of statement is cheaper than a verification
pass that assumes every number is wrong.

**Consequences.** The lead now marks its counts, and has produced a ledger of the run so far. Three entries
in it are **SAMPLED and load-bearing**, and must be re-derived before the stories that rest on them:
- *"`App.css` contains exactly one `@media` query"* — never verified by the lead, subagent-sourced, and
  **load-bearing for 13.3**, which cannot introduce a rail with a breakpoint if it is true. Story 17.5 has
  since touched `App.css`, so it must be derived at 13.3's gate, not inherited.
- The Epic 13 preview findings generally — eight of ten premises are uncorroborated by the lead itself.
- *"37 families under `folio-designer/public/fonts`"* — an `ls | wc -l` of directories, with no check that
  each is a family.

One count was also **stated more tightly than the truth**: "the complete command registry is 24 kinds" is
closed for `ApplyComponentCommand`'s dispatch, but `ApplyPageSetupCommand` is a separate exported entry
point, so the command **surface** is 25. Nothing downstream is wrong — page setup was discussed separately
— but anyone counting new commands against 24 would be off by one.

**Verified independently before filing:** `prohibited` holds **14** regex entries, of which three police
arithmetic rather than measurement; there are **5** exception sites, not four; and `:318` rewrites
`document.fonts.ready` **repo-wide, scoped to no owner**. That last one is the strongest argument available
that Story 17.6's named-owner exception is a *narrowing* — it is a smaller instance of a carve-out this
guard already carries — and it was relayed to 17.6's builder mid-flight.

**How we'd know it was wrong.** A run where marking counts becomes ceremony — every number tagged CLOSED
without a boundary actually being stated. The boundary is the substance; the label without it is worse than
nothing, because it launders a sample as an enumeration.
