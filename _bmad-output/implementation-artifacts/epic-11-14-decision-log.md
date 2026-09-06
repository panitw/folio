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


### Re-grounding refresh — 2026-09-05 (second session)

*Filed by the orchestrator from the lead's report. The first lead did not survive the session break; this
is a re-grounding from this section and the rulings below it, not a re-derivation from the spine, the ADRs
or the epics doc. Verified at HEAD `28cd225`, working tree clean, re-measured rather than inherited from
the session's opening `gitStatus` — which is how the previous grounding acquired a stale baseline.*

**Sources read, CLOSED:** this log in full (890 lines); `sprint-status.yaml` (676 lines);
`deferred-work.md` §DW-191–195 plus the 195-entry index; `git log --oneline -25`; the diffstats of the four
post-grounding code commits.

**State.** All four epics of this run (11: 3 stories, 12: 5, 13: 5, 14: 10) are `backlog`; nothing in them
is built. What has shipped since the run opened is Wave A's 17.6 plus the matrix repair.

**Four of the five forks named at grounding are settled, and settling them shrank the run** — D-A (Epic 11
at three families and nine faces; Noto Sans SC stays Regular, making it the permanent shipped instance of
"this family has no face at this weight"), D-B (weight and slope ride `FontChainEntry`; the public API does
not move; coverage first, style within the covering entry), D-12.4.1 (padding stays an engine property),
and D-C/D-000.6 (the AD-17 scan repaired by named-owner exception). **The open fork is D-D — Story 15.2a's
position and blast radius.**

**Verified independently rather than taken from this log:** 17.6's fourth named exception
`withoutApprovedPaintedBorderReadback` exists, is scoped to `e2e/e9-5-border-no-ink.spec.ts`, and asserts
its own reason is still present (`canvas-authority-contract.test.ts:575–586`, five-row matrix). The
`folio-designer-known-red` quarantine job is gone and the six previously-dark steps now run — **the
masking hazard over Epics 13 and 14 is closed.** The matrix repair passes on `darwin/arm64`
(`ok … 1.973s`) — **CLOSED for one leg only**; the other three are Docker/CI. `gofmt -l` empty in all
three Go modules, CLOSED and enumerated.

**Four new deferred items came out of 17.6**, one load-bearing for this run's largest epic: DW-192 (the
pointer-input carve-out still deletes a whole `App.tsx` function body, waiving every AD-17 prohibition
inside the file holding the canvas projection), DW-193 (the e2e suite is compiled and never executed in
CI, so the assertion *earning* the new carve-out never runs), DW-194 (nothing pins the un-quarantined step
names or forbids `continue-on-error`), DW-195 (`prohibited` is a denylist — the third instance of that
class this run).

**Story 17.5's provisional flags are resolvable.** It committed at `d2f2a1e`, so
`property-prose-height.test.ts` is committed fact — but re-derive by symbol at each Epic 14 gate, since it
is a source-text contract on the shared property row.

**Three tensions raised, carried forward rather than closed here:**

1. **D-000.11's wording under-runs its own gate.** Measured CLOSED, by matching a `//go:build …matrix`
   constraint in the first three lines of every constrained `.go` file under `folio-go/`: the tag is on
   **7 files and 12 test functions**, not one file and six. `matrix.yml` runs **2 of the 12** per leg,
   name-filtered (`-run TestTargetRenderHash`, `-run TestTargetProbeHex`); `ci.yml` only *compiles* the
   tag. So "the matrix suite" and "what CI runs" are different sets, and a gate that says the former while
   pasting the latter's command **re-creates the exact blind spot D-000.11 was written about.** Among the
   ten CI never runs is `TestShippedFacesReproduceFromUpstream`, which is the guard Epic 11's nine new
   faces must satisfy. **Amended below.**
2. **Two boundary gates are owed and unbooked.** Epics 16 and 17 have every story `done` and are still
   `in-progress`. Epic 16 closed without a gate, and that is *why* the matrix regression hid.
3. **15.2a collides with D-12.4.1.** `component-property-command.ts:7` routes ten fields through the
   unquoted `rawNumberLiteral`, four of them `paddingTop|Right|Bottom|Left` — the exact set D-12.4.1 makes
   the Go layer refuse. 15.2a runs before 12.4, so it must not entrench a designer-side padding write path
   that 12.4 removes. Routed to the lead as **D-D.1**.

**One correction absorbed:** D-B's supporting count — "the only three channels for caller-supplied font
bytes" is at least five (`CanvasWithTextPaint`, `PreviewIdentity`). The ruling is strengthened, not
weakened; recorded because it is the incident that forced D-000.7.

**Open record discrepancy, unchanged:** `sprint-status.yaml` carries
`15-0-a-catalogue-face-arrives-when-it-is-picked: backlog`, and `15.0` appears nowhere in `epics.md`
(grep exit 1; positive control: `15.2a` matches once). A tracked story with no epic text is either a lost
story or a phantom, and Epic 15 is the release-blocker set.

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

---

### D-000.8 — Epic 13's premises, closed: nine of ten hold, and the tenth is half-false in the expensive direction
**Orchestrator record** of the survey D-000.3 mandates, first half.

**Why this was run.** The lead's Epic 13 findings were subagent-sourced and it had verified only two of
them itself, flagging the rest **SAMPLED** under D-000.7. Five preview stories were about to be specced on
them.

**Verified and closed:** no PDF export path exists anywhere in the designer; the palette rail is a sibling
of the design/preview ternary and so renders in preview mode; the evidence surface is a single `<p>`; zoom
is clamped 0.5–2.0 in 0.1 steps with both readouts as `<output>`; the render effect depends on the whole
view-state object; `previewOversample` is a constant and the `devicePixelRatio` ban behind it is
**implementing code, not a comment** — `prohibited` carries the pattern and a red-proof mutation exercises
it.

**THE `@media` CLAIM IS TRUE AND IS MORE THAN A FACT — IT IS AN ENFORCED CONTRACT.** The lead flagged this
as its most load-bearing sampled claim. Independently verified: `App.css` holds exactly three at-rules —
one `@import`, one `@supports`, one `@media (prefers-reduced-motion: reduce)` — enumerated by matching the
bare character `@` rather than the expected answer, with 217 `color` matches as the positive control that
the file reads as text. The part nobody had: `canvas-authority-contract.test.ts` has a test named
*"allows only the non-document reduced-motion media rule"* that asserts the extracted media-query list
**equals** `['prefers-reduced-motion: reduce']`. So the single-`@media` state is not a stylistic accident;
**any responsive breakpoint added by 13.2 or 13.3 turns that test red and must amend the contract
deliberately.**

**THE ONE HALF-FALSE PREMISE, AND IT COSTS ORDERING.** Epic 13 states that with no height cap,
`scrollTop`/`scrollLeft` can never be non-zero and the restore effect is dead code. True vertically. False
horizontally: `overflow: auto` is both axes and the canvas carries `max-width: none`, so a zoomed page
already scrolls sideways and already writes `scrollLeft`. **The render effect's tear-down is therefore live
today, not latent behind the height fix** — filed as **DW-191** — and 13.2 must fix the dependency array
**before or with** the cap, not after it as the epic implies.

**Three more corrections that would have produced wrong specs.**
- **13.1 has two hardcodings, not one.** Both tiers force the name as well as the MIME
  (`input-download.ts` at the blob type *and* through `folioName(...)`, mirrored in
  `file-system-access.ts`). **A spec naming only the MIME ships a PDF saved as `.folio`.**
- **13.4's gate is not a file.** It is parsed `SampleData` state, seeded as readily from a prop as from an
  opened file, and `runPreview` additionally requires an engine and a snapshot. Entering preview mode is
  **not** gated at all — only the render is.
- **`preview-authority-contract.test.ts` forbids 13 tokens, not the 4 recorded**, and pins six positive
  App.tsx literals plus three negatives — including the freshness `<p>`'s exact attribute spelling and
  order. **Story 13.5 cannot rewrite that element without amending this contract in the same change.**

**Consequences.** Two unflagged blockers now have owners rather than surfacing at a plan gate: the
`@media` contract (13.2/13.3) and the freshness-element pin (13.5). DW-191 reorders 13.2 internally.

---

### D-000.9 — Epics 12 and 14, closed per acceptance criterion: ten ACs would be actively wrong if specced as written
**Orchestrator record** of the survey D-000.3 mandates, second half.

All six of the lead's carried findings **CONFIRMED** against committed code, each with its population
stated. The command surface is **25** (24 in `ApplyComponentCommand`'s dispatch plus
`ApplyPageSetupCommand`), and none of the 25 writes a band height, a locale, a UTC offset, a table
`headerHeight`, `altRowBackground` or `headerStyle`. Millimetres: **0** hits over 106 tracked files, with
`localeCompare` matching in the same invocation as the positive control.

**THE TEN THAT WOULD SHIP A WRONG STORY.** Recorded here so each spec inherits the correction rather than
rediscovering it:

1. **14.2 AC2** — "the border controls are not offered, because a filled bar has no edge set to draw."
   **The reason is false.** `elementBoxDeclaration` is kind-agnostic and `borderPaints` returns true for a
   present, non-null border with a non-empty edge set: **a border on a Line paints.** Hiding the control
   removes a working capability, which collides with 14.2's own final AC (I-5).
2. **14.3 AC2** — "a 1pt Line is 1.33 CSS pixels at 100%." `.canvas-component-line` already floors to
   `max(2px, …)`. The AC also forbids padding that "changes what is drawn" — **the existing floor already
   does.** Both halves need reconciling with shipped CSS.
3. **14.7 AC8** — the millimetre ruling, framed as a live product-wide two-unit conflict. The product has
   **zero** millimetres. It is a mockup-fidelity decision only, and a much smaller one.
4. **14.6 AC1** — "rather than a concatenated `kind · count · preview · candidate` string". The value
   **is already in that string**. This is a layout change, not the "the panel never renders them" the
   preamble claims.
5. **12.4 AC1** — offers the owner two rulings when **a third is already claimed** (see D-000.10).
6. **Epic 12 preamble** — "the command layer accepts all four keys on any component, and the panel can
   write it." First clause true; **second false** — no `FieldSpec` authors padding. Population: the 9
   `FieldSpec` declarations, enumerated.
7. **14.7 AC5** — asks for the sample's item count as read-only context. `TableColumns` carries no item
   count and `candidates` carry no counts. **The story silently requires a new projection field it does
   not name.**
8. **14.10 AC1** — "that column is the selection" assumes a selection model that can hold a column.
   `selected` is `ReadonlyArray<string>` of **element** ids and every consumer resolves them against
   `canvas.components`. **Unnamed prerequisite, and a selection-model change.**
9. **14.1 AC3** — "Align is three SVG icons." It is **four** for an all-text selection since justify
   landed. The sweep will trip on it.
10. **14.7 AC9** — reads as a preservation clause and is **the largest mechanical risk in the story**:
    `cellCount = 11` is load-bearing for `focusCell`, `moveFocus`, Home/End and the post-reprojection
    refocus. Reducing to six columns rewrites all four.

**FOUR ACs ARE ALREADY SATISFIED and must be specced as preservation, not construction:** 14.6's roving
tab stop and full arrow/Home/End/Enter navigation; 14.7's roving grid navigation and dialog Tab trap;
14.4's "hidden controls preserve the underlying value" (true by architecture); and 12.3's alt-row and
header-style cascade, which is entirely engine-side and needs no work at all.

**One AC is satisfied trivially and cannot fail**: 12.1 AC5 and 12.2 AC4 ("unedited bytes unchanged") hold
**because no writer exists**. They become real tests only once the writer does.

**A blocker with a new owner:** `property-prose-height.test.ts` and `design-contract.test.ts` both assert
over **raw source text**, and between them pin an exact allowlist of eight declarations on the prose
handle, the absence of any hex/rgb/hsl literal in `App.css`, exact-once counts for two type tokens, and
several verbatim CSS lines. **Any Epic 14 restyle fires them.**

---

### D-000.10 — The padding ruling exists only in a code comment, and 12.4 offers the owner the wrong two options
**Orchestrator finding.** Escalated to the owner rather than ruled.

**Situation.** Story 12.4's first AC asks the owner to rule padding one of two ways — *insets on every
component*, or *table-only*. But `App.tsx:1714` already records a **third**: *"(owner's call, 2026-08-30):
`style.padding` stays an engine property that a loaded document keeps and renders — the panel simply does
not author it."*

**Measured, because a comment is not a measurement.** Across all five decision logs, `padding` occurs 30
times — 19 in the MVP log alone, against a `fontFamily` positive control that hits in every log. Reading
them: every one concerns the **engine** side (table cell padding and row height, base64 padding, digit
zero-padding). Filtering those 30 for `panel|author|inset|table-only` returns **nothing**. **There is no
decision-log entry recording an owner ruling that the panel does not author padding.**

**Why this cannot be resolved by ruling.** Either the comment records a real decision the log lost — in
which case 12.4 as written re-opens a settled question — or it records something that was never an owner
ruling, in which case it has been silently governing the panel's behaviour for six days on no authority.
**Both are possible and the difference is the owner's memory, not the repository's.**

**Consequence if guessed.** 12.4 is the **first** Epic 12 story in the build order, chosen precisely
because it is ruling-first and settles what `padding` means before 14.8 writes a cell-padding control
whose meaning depends on the answer. Guessing here mis-specs two stories, not one.

---

### D-12.4.1 — Padding stays an engine property; the panel does not author it, and the command layer stops accepting it off a table
**Owner decision**, taken at the terminal 2026-09-05, confirming the ruling that existed only as a comment.

**Verdict.** The `App.tsx:1714` comment was right and is now recorded where it belongs. `style.padding`
remains an engine property: **a loaded document keeps it and renders it**, and **the panel never offers a
control for it**. In addition, the command layer stops accepting `padding*` on non-table kinds, which it
currently accepts on any component.

**In simple terms.** Padding already does a real job inside a table cell, and the engine will keep doing it
for any document that arrives carrying it. What changes is that nothing in the designer will ever *write*
one onto a text box or a line — and the command layer stops pretending it would honour it if you did.

**Why the confirmation mattered more than the answer.** The ruling was being enforced in the product on the
authority of a comment nobody could source. Measured across all five decision logs: `padding` occurs 30
times, every occurrence concerns the engine side, and filtering for `panel|author|inset|table-only` returns
nothing, against a `fontFamily` positive control hitting in every log. **A rule that governs behaviour and
exists only as prose is one refactor away from being deleted as a stale comment** — which is precisely how
a settled decision becomes an accidental regression.

**What this does to the two stories that depended on it.**
- **12.4 shrinks from a fork to a small, well-founded story:** record the ruling, and add the guard so the
  command layer refuses `paddingTop|Right|Bottom|Left` on any kind that is not a table. Today
  `applyPropertyChanges` accepts all four on **any** element, with no type guard — so the guard is a real
  narrowing and needs a red-first test.
- **14.8's cell-padding acceptance criterion is DROPPED.** It proposed a control whose basis this ruling
  removes. Note the AC's own analysis was correct on the engine side — `table_render.go` is the only
  consumer of `.Padding` on any render path — which is exactly why the control is unnecessary rather than
  merely unwanted.

**The narrowing has a cost, and it is accepted.** A hand-written document carrying padding on a text box
still **loads and renders**; what it can no longer do is receive a padding command from the designer. That
is the intended asymmetry: the engine honours what it is given, and the designer refuses to author what it
cannot mean.

**Consequences.** No `FieldSpec` may author a padding field. The guard belongs in Go's command layer, not
in the panel, because the panel not offering a control is not a guarantee — AD-15 says the engine owns the
document, so the refusal must live where the document is written.

**How we'd know it was wrong.** Authors hand-editing padding onto text boxes and asking why the designer
will not round-trip it. That would mean padding-as-inset had real demand, and the answer would be to
implement it properly in the engine's layout path rather than to relax the guard.

---

### D-000.11 — A build tag is a place where regressions go to hide, so the epic gate must name it

**Decision.** The epic-boundary heavy test **must run the `matrix`-tagged suite UNFILTERED** —
`go test -count=1 -tags=matrix ./...` — on all four `FOLIO_MATRIX_TARGET` legs, as a named command in the
gate. Not as a consequence of `go test ./...`, which cannot see the tag, and **not by pasting CI's
name-filtered commands**.

**AMENDED 2026-09-05, and the amendment is the load-bearing half.** As first written this ruling said
"the `matrix`-tagged suite", which is a **proxy**, and the lead was right that a proxy is what caused the
regression in the first place. Measured CLOSED, by matching `//go:build …matrix` in the first three lines
of every `.go` file under `folio-go/`: the tag is on **7 files and 12 test functions**. CI's `matrix.yml`
runs **2 of the 12** per leg, name-filtered (`-run TestTargetRenderHash`, `-run TestTargetProbeHex`);
`ci.yml` only *compiles* the tag (`go build -tags=matrix ./...`). So "the matrix suite" and "what CI runs"
are **different sets**, and a gate that says the former while pasting the latter's command re-creates the
exact blind spot this ruling was written about. The twelve, enumerated so the difference cannot be
paraphrased away:

| Run per-commit by CI | Run by **no gate at all** |
|---|---|
| `TestTargetRenderHash` | `TestAssertFixturesShareToolchainRedProof` |
| `TestTargetProbeHex` | `TestCrossTargetByteIdentity` |
| | `TestEmbeddedFontTransferredReadingHolds` |
| | `TestExpectedBreaksHumanSignOffIsRecorded` |
| | `TestFMAProbeDiverges` |
| | `TestHarnessExportsPins` |
| | `TestShapedTextThaiSemanticSignOffIsRecorded` |
| | `TestShippedFacesReproduceFromUpstream` |
| | `TestStatementSemanticSignOffIsRecorded` |
| | `TestThaiStackedMarksSemanticSignOffIsRecorded` |

**What the gate uniquely buys is that right-hand column.** If a leg genuinely cannot afford the full run,
the gate **names those ten explicitly** rather than falling back to the tag.

**What happened.** Story 16.8 (`4d2b27e`, 2026-09-04) shipped Roboto as a fourth face and grew
`shippedFaceSpecs` from 3 to 4. Two guards in `folio-go/matrix_test.go` —
`requireInstancedShippedFaces` and `requireShapedTextIsShaped` — were comparing their fixture's
embedded-program count against `len(shippedFaceSpecs)`. Both fixtures render Latin+Thai+CJK and embed
three programs. Both went red on all four targets the moment 16.8 landed, and the `Cross-target byte
identity` workflow has been failing ever since: on `d546e2a`, `render-darwin-arm64`, `render-linux-amd64`,
`render-linux-arm64` and `render-js-wasm` all **failure**, `compare-render-hashes` **skipped**.

**Why nobody saw it.** `matrix_test.go` carries `//go:build matrix`. Measured over the 20 Epic 16/17 story
artifacts: **3 name `-tags=matrix`** (16.0, 16.1, 16.1b) and 16.8 is not among them, against a positive
control of **13 of 20** that run `go test` at all. Across all 110 artifacts, 61 do name the tag — so this
is a well-known gate that Epic 16 simply stopped invoking. Epic 16 then closed **without a boundary gate**,
which is the run parameter ("heavy test at the end of epic") that would have caught it.

**The defect the guards actually had.** They used *the whole shipped set* as a **proxy** for *the faces this
document uses*. The two quantities were equal for three faces and the proxy was invisible; it broke the
instant a shipped face stopped being a used one. The guard's own comment said "one per shipped face
**actually used**" and "exactly **the three** shipped faces" — the prose was right and the code had drifted
out from under it. **A proxy that is currently accurate is still a proxy**, and nothing in the type system
records which of the two things it meant.

**The fix.** `threeScriptFixtureFaceKeys` names the three faces the two three-script fixtures exercise, and
`threeScriptFixtureFaceSpecs` resolves them against `shippedFaceSpecs` so the PostScript names stay
single-sourced. A key that no longer resolves is a hard failure, so renaming a face cannot silently shrink
the expectation. Both the count and the `/BaseFont` coverage witness now derive from that named subset.
Shipping a fifth unused face no longer breaks either guard; adding a face to a *fixture* without listing it
here still does.

**Red-proved, three mutations, each on darwin/arm64:** adding `"Roboto"` as a fourth key → RED ("must embed
exactly 4 … got 3"); dropping `"Noto Sans Thai"` → RED ("exactly 2 … got 3"); renaming to a bogus
`"Noto Sans TC"` → RED ("the fixture and the shipped set have drifted apart"). A first attempt at the
add-Roboto mutation reported GREEN; that was a bug in my own mutation harness (the shell split the
search string on `|`, so the edit was a no-op) and not a vacuous guard — **a mutation that does not change
the file proves nothing, so a mutation harness needs its own positive control**, which is why the applied
line is echoed back above each run.

**Consequences.** Every epic-boundary gate from here names the matrix suite and its four targets
explicitly. A story that edits `shippedFaceSpecs`, `fonts.Shipped()`, or any fixture template must run it
regardless of cadence — those are the inputs the matrix guards close over.

**How we'd know it was wrong.** If the matrix legs proved slow enough at every boundary that epics stopped
gating at all. The four legs together ran in well under a minute locally with Docker available, so that
cost is not real today.

---

### D-000.12 — A guard whose denominator is its own input can only ever say "N of N"

**The finding as first written was overstated, and the overstatement is recorded here rather than edited
away.** The orchestrator reported that "the shipped Roboto binary has no recorded provenance." **That is
false.** Roboto's provenance is recorded *and* machine-checked per-commit by a complete chain:
`folio-go/fonts/roboto/NOTICE.md` carries the full idiom — upstream `googlefonts/roboto-3-classic`
release `v3.016`, download URL, path inside the archive, fetch date, source sha256
`e688a215…9355`, shipped sha256 (the same value), size 355,956 bytes, and the relation
`copied unmodified, no derivation`; `font-catalogue.test.ts` asserts the committed binary's sha256 equals
its own NOTICE's recorded digest; and `fonts_test.go`'s `TestShippedRobotoMatchesDesignerCatalogue`
asserts the engine's embedded bytes are byte-identical to that verified catalogue file. Provenance
transfers by asserted equality anchored on a digest, which is the correct construction.

**CLOSED by execution, not by reading** (the lead flagged this as its own open assumption): mutating the
last byte of `folio-designer/public/fonts/roboto/Roboto-Regular.ttf` turns `font-catalogue.test.ts` from
6 passed to `1 failed | 5 passed` with *"roboto: the binary's sha256 is not the digest its own NOTICE.md
records"*. The AC1 loop does execute over `roboto`. Binary restored.

**A permanent record that overstates a defect makes the repair look bigger than it is and misdirects
whoever reads it next** — which is why the correction is the first paragraph of the entry and not a
footnote. The generalisation that caused it: reasoning from one vacuous guard to "the face is unverified"
without measuring the other routes that might cover it.

**What is actually wrong — two things, neither of them an unverified binary.**

**(A) `fontgen`'s "3 of 3" is vacuous.** Its denominator is its own manifest, so it can only ever report
`N of N`. Measured: `tools/fontgen/instance_faces.py` carries three entries; `fonts.Shipped()` returns
four; a case-insensitive search for `roboto` across `tools/` returns nothing. The cost today is **not** an
unverified Roboto — it is that **nothing forces a new shipped face into any accounting route at all.**

**(B) One engine-side NOTICE is unpinned, and it is Roboto's.** `font-binary-identity.test.ts` iterates
the **six wasm-declared browser families**; Roboto is not one of them, so
`folio-go/fonts/roboto/NOTICE.md`'s digest table is the one engine-side provenance record nothing
asserts. It is currently correct; it is a duplicate record free to drift.

**Decision (lead, 2026-09-05).**

1. **Reuse the existing idiom; do not invent one.** A static, non-derived face's provenance record *is*
   its `NOTICE.md` table. Roboto already has all of it. **What is missing is the assertion, not the
   record.** Do not add a second manifest, and do not bend `instance_faces.py`'s `src`/`out`/instancer
   shape around a face with no derivation — *a route that has to lie about having an instancer step is
   worse than a second named route.*
2. **The coverage guard is Go-side, in `folio-go/fonts`, UNTAGGED**, modelled on
   `TestShippedSpecCoversEverythingShipped` and failing in both directions. Two obligations, both TOTAL
   over `fonts.Shipped()`: (a) every shipped face has a `NOTICE.md` beside it whose recorded shipped
   digest equals the embedded bytes and whose recorded size equals their length — closing (B) by one rule
   rather than three-plus-an-exception; (b) every shipped face is accounted for by **exactly one named
   route**, `derived` (present in fontgen's manifest) or `static-upstream` (its NOTICE records
   `copied unmodified, no derivation`), **with no fallback bucket and no default arm** — unaccounted is a
   hard failure, and a face in *both* routes is also a failure.
   **Untagged is the load-bearing half.** The accounting is pure file reading: no fontTools, no venv, no
   network, so it runs per-commit. **Putting it behind the tag would re-create D-000.11 exactly** — a
   cheap guard hiding in an expensive suite. The *reproduction* stays `matrix`-tagged because it genuinely
   needs the pinned toolchain, and `TestShippedFacesReproduceFromUpstream` must take its denominator from
   `fonts.Shipped()` and report *"derived and compared 3 of 4 shipped faces; 1 accounted static-upstream"*,
   or refuse.
3. **Scope: Epic 16's, not Epic 11's and not a deferred entry.** Story 16.8 grew the shipped set and
   extended `shippedFaceSpecs` but not the accounting; that is 16.8's own defect, and Epic 16 is still
   `in-progress`. **A repair may join an epic that has not closed; an extension may not** — and this is a
   repair. One small story: the accounting guard, the fontgen denominator, the fourth NOTICE's digest pin,
   red-proved in both directions. **It must land before Epic 11's first face:** six of D-A's nine are
   derivable and three are static, so 11.1 is the first story to use the static route *as a route* rather
   than as a one-off, and the route must exist and be enforced before it gains three more members.
4. **The Epic 16 boundary gate cannot conclude "pass", and must not conclude "Roboto is unverified".**
   It records (A) and (B), and Epic 16 closes once the guard in (2) exists. **Closing Epic 16 over a guard
   that is vacuous about Epic 16's own headline face — the one it made the default typeface for every new
   document — would make the gate ceremonial**, which is what D-000.1(3) traded per-story testing away to
   avoid. The gate procedure must also record that `TestShippedFacesReproduceFromUpstream` needs Python
   3.12.13 + fontTools 4.63.0 via `FOLIO_FONTGEN_PYTHON`, **which CI does not have and cannot acquire** —
   a standing property of the gate, and a second argument for the accounting half being toolchain-free.

**Grounding.** AD-26 and I-7 (every redistributed asset travels with its terms, upstream release URL and
source sha256) are **already satisfied** for Roboto — which is why this is not an AD-21/AD-22 breach. AD-21
is untouched: the shipped bytes are correct and unchanged, and nothing here moves a golden.

**Guardrails.** No fallback or default arm — a face matching neither route must fail, not land in "other";
*a two-option fork whose arms both leave a population uncovered is a false fork*. Every message the guard
prints takes its denominator from `fonts.Shipped()` and names the unaccounted face rather than reporting a
ratio. Do not restate Roboto's digest as a literal in Go — `fonts_test.go` already records why (a
hardcoded digest proves only that someone typed the same string twice); add the NOTICE pin as a third
comparison, not a fourth copy of the number. **The red-proof must include the direction that was silent:**
add a fifth face to `Shipped()` with no NOTICE and no manifest entry and prove the guard goes red. A proof
that only *removes* a face re-tests the direction that already failed loudly.

**How we'd know it was wrong.** If the two named routes turn out not to partition the real world — a face
that is neither derived nor copied unmodified, say one hand-subset in the repo. Then the fix is a third
*named* route, never a fallback arm.

---

### D-000.13 — A workflow that has never been green cannot be told from a broken one

**Measured, CLOSED over every `ci.yml` run GitHub retains:** **12 of 12 runs concluded `failure`**, from
`ec15d36` (2026-08-26) to `28cd225` (2026-09-04). `Build, vet, and guardrails` — the repository's main
gate — **has never once passed.**

**The mechanism, and it was mine.** Earlier in this run I quarantined the sanctioned known-red
(`TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`) into its own job, `folio-go-known-red`, and
described that as making CI meaningful. The job ran the test and **let its exit code become the job's**.
The test fails by design, so the step failed, so the job failed, so the workflow failed —
**unconditionally, forever.** On `28cd225` this job was the **sole** failure; `folio-go`, `hashmatrix`,
`lint` and `folio-designer` all passed.

**What it cost.** A permanently-red workflow is indistinguishable from a broken one, so its red carries no
information. That is precisely how a real `gofmt` breakage in `lint/internal/rules/licencegraph_test.go`
sat in CI unnoticed, and why I first reported that job green from a misread local `&&` chain rather than
from the workflow that was already telling me otherwise.

**The fix: assert the expectation, do not inherit it.** The step now inverts — green while the known-red
is red, and **red the moment the known-red starts passing**, which is the signal that the quarantine
should be deleted and which the previous form could never have delivered. Red-proved locally in both
directions before pushing (direction 1: the real filter → exit 0; direction 2: a filter standing in for a
fixed known-red → exit 1).

**A second defect the inversion closes for free.** `go test -run` with a filter matching **nothing** exits
**0**. Under the old form a misspelled `KNOWN_RED_TEST` produced a green step and a **silently vacuous
quarantine**; under the new form it goes red. This is the same shape as D-000.12's finding — *a guard
whose denominator is its own input* — and the third distinct instance this run of a guard that could not
fail.

**Consequences.** No job may take an expected-red's exit code as its own verdict. A job that expects a
specific failure asserts it and says what a pass would mean.

**Scope.** This is a repair of my own quarantine, so it lands now rather than waiting. The broader
subject — *what CI's red is allowed to mean* — remains **Story 15.2 ("CI's red means something")**, still
`backlog` in the release-blocker epic, and this entry is evidence for it rather than a substitute.

**Verification is only possible on GitHub.** A workflow change cannot be proved locally beyond its shell
logic, so this one is pushed to be measured, and the result recorded against it.

---

### D-000.14 — STANDING RULE: a guard that reads its own input as its verdict cannot fail

**The rule.** Before trusting any guard, ask **where its expected value comes from**. If the expectation is
derived from the same artifact the guard is checking, the guard can only ever return the answer it started
with. State the *independent* source of the expectation, or the guard is decorative.

**This is not a theory. It is the single most productive defect class found in this run — four instances,
all real, all shipped, none caught by review:**

| # | Guard | Its verdict came from | What it could never see |
|---|---|---|---|
| 1 | `requireInstancedShippedFaces` / `requireShapedTextIsShaped` | `len(shippedFaceSpecs)` — the *shipped set*, standing in for *the faces this fixture uses* | The two quantities diverging. They did, at 16.8. Loudly red for four days. |
| 2 | `TestShippedFacesReproduceFromUpstream` | its own manifest — reports `3 of 3` while `fonts.Shipped()` returns 4 | A shipped face nobody added to the manifest. **Silently green.** |
| 3 | `folio-go-known-red` CI job | the expected-red test's own exit code | Anything. The job was red unconditionally, 12 of 12 runs, so the whole gate's signal was worthless. |
| 4 | a recursive `grep` naming its own answers | the pattern enumerated what it expected to find | Whatever nobody thought to enumerate — the denylist form of the same error. |

**The two directions are not equally dangerous.** #1 failed *loudly* and was found in four days. #2 failed
*silently* and nothing was ever going to find it. **A guard sized too large announces itself; a guard
sized too small congratulates you.** When choosing which way to be wrong, prefer the loud one.

**The tell.** Every one of these reads naturally and passes review. "Compare against the shipped set",
"compare against the manifest", "run the known-red test", "grep for the forbidden patterns" are all
sentences that sound like verification. **The defect is never visible in the guard's own text — only in
the relationship between its expectation and its subject**, which is why it survives reading and dies to
one question: *what would have to change for this to fail?*

**Consequences.** A guard's expectation is stated independently of its subject, or the guard says out
loud that it is a tautology and why that is acceptable. A count declares its denominator's origin
(D-000.7's population rule, sharpened: it is not enough to name the population — name where the
*expectation* came from). Every new guard is red-proved by mutating the **subject**, not the expectation;
mutating the expectation only proves the guard reads its own input, which is precisely what is in
question.

**How we'd know it was wrong.** If applying it turned routine assertions into ceremony. It should not:
the question is one sentence, and in four for four cases here it would have found the defect immediately.

---

### D-000.15 — `git grep` has its own false zero, and I gave the wrong advice three times

**Correction to my own standing instruction.** Earlier in this run I told the builder, the lead and two implementers to use `git grep` **instead of** recursive `grep`, because recursive `grep` returns false zeros in this repository. That advice was incomplete in a way that matters: **`git grep` searches tracked files only.**

**Found by Story 15.2a's implementer, on its own work.** `git grep -ln '"folio-designer"'` returned three files and missed a fourth — the implementer's own, because it was untracked. It needed `--untracked`.

**This is worse than the trap it replaced, in one specific way.** Recursive `grep`'s false zero is arbitrary. `git grep`'s is *systematic and adversarially aimed*: it hides exactly the files a story has just created, which are the files most likely to be the answer to a question that story is asking. A survey of "does anything else do X?" run during implementation will reliably fail to see the thing the implementer wrote ten minutes earlier.

**The rule.** Use `git grep --untracked` whenever new files could be in scope — which is always, during implementation — and **state which form was used beside the count**. A bare `git grep` zero is a claim about the tracked tree, not about the working tree, and those differ precisely when it matters.

**The general lesson, which outlives both tools.** "Use tool X instead of tool Y" is never a complete answer to a measurement problem. **Every search tool has a population it cannot see**, and the population is the thing to state — not the tool. This is the third distinct false-zero mechanism found in this run (recursive `grep`'s, `git grep`'s tracked-only default, and a `-run` filter matching nothing and exiting 0), and they share no implementation, only the shape: *a zero that means "I did not look there" rendered identically to a zero that means "it is not there."*

---

### D-12.A — Epic 12's five stories, surveyed per acceptance criterion at `71627a5`

**Populations CLOSED and named:** P-CMD (the command surface) = **25** — 24 `case` arms in
`ApplyComponentCommand` plus `ApplyPageSetupCommand`, every arm read. P-PAD-GO (`.Padding` occurrences
under `folio-go/`) = **25 across 5 files**. P-FS (`FieldSpec` declarations in `App.tsx`) = **9**. P-FOLIO
(`.folio` files outside `node_modules`) = **31**. Positive controls fired: `borderWidth` hits 7 designer
files in the same sweep where `paddingTop` hits 5 and **not** `App.tsx`, so the padding zero in `App.tsx`
is real.

**Two acceptance criteria would have shipped wrong, and they are one defect twice.** 12.1 AC4 and 12.5 AC7
both promise a stranded component is *"accepted and clipped with FR44's existing diagnostic"*. Falsified
two ways: `element_box.go` states verbatim *"the declared WIDTH is FR44's only clip bound"*, and the diag
registry's **19** Code declarations contain no band-overflow code — **the AC cites an engine answer the
engine does not have**; and `containComponent` (**10** call sites, *"the ONE band-extent validation in the
designer command path"*) already refuses that state for the capping bands, so accepting it yields a
document where every later command on the stranded element is refused and the TS mirror drops the whole
snapshot.

**RULING, Fork 1 — the correction is layered, and it is the pattern already shipped.** Not "refuse" versus
"clamp" but both, one at each layer: **the engine refuses** (`containComponent`, unchanged, no new
diagnostic, so AD-14's permanent-surface cost is never paid), and **the panel's control cannot propose
it** — the band-height field's floor is the maximum of (y + height) over that band's components, derived
from the projection the browser already holds. This is exactly what `POSITIVE_LENGTH_FIELDS` exists for,
in its own words: *"a keypress can never propose a value the command path will refuse."* Band height is
the same bound one axis over.

**This dissolved the escalation.** The objection to refusing was that an author could not shorten a band
without first moving things. Under the layered shape they can shorten it *to the lowest occupied edge,
with the control stopping there visibly*. The epic's promise survives, so nothing is traded and there is
nothing to put to the owner.

Bounded deliberately: **only the vertically capping bands** (pageHeader, pageFooter) — the content band
paginates and gets no floor. **12.1 and 12.5 share ONE derivation of that floor, written once and consumed
twice** — identical text in two ACs drifts; one function called twice cannot. Story 17.4's Spec Change Log
records `containComponent`'s upper bounds as an open question; **12.1 settles it for the band-height axis
only** and is not licence to complete the mirror. The D-7.4.5 red-proof is required: move the Go bound
without the TS floor and prove a keypress can propose a refusal.

**RULING, Fork 2 — fold the projection widening into 12.2 and 12.3; do not split.** A story that widens
`CanvasProjection` without its consumer has no observable behaviour and no red-proof — zero call sites is
unbuilt — and D-7.4.5 requires the projection and its `hasOnly` mirror to move in one commit regardless,
so the split would be a split of one commit. **The red-proof is the omission**: add the field without the
`hasOnly` entry and assert the canvas blanks. A happy-path test cannot see the failure mode that makes
this worth writing down — a mismatch does not degrade, it drops the snapshot, terminates the worker and
blanks the canvas.

**New run rule, from three instances (12.2 AC1, 12.3 AC1, 14.7 AC5):** *any AC of the form "the panel
shows the engine's current X" is a projection-widening claim unless X is already projected.* Every such AC
must say which, and name the field. One grep at spec time against a failure that blanks the canvas at
runtime.

**RULING, Fork 3 — dissolved. 12.5 AC6 is FOCUS, not selection.** Verified by reading `App.tsx`: the band
renders as a `section` carrying `tabIndex={0}` and **its own `onKeyDown`** — already focusable, already
handling keys. The arrow-key nudge is a *document-level* handler gated on `selectedRef.current.length ===
1`; AC6 was only entangled with the selection model because it would have reused it. Extending the
section's existing handler needs no selection model, and the event is handled at the focused element
before reaching the document listener. **So this is not a second instance of 14.10's prerequisite — it is
one story reaching for a mechanism it does not need.** 12.5 stays in Wave B. Conditions: the handler must
`preventDefault`/`stopPropagation` on the arrows it consumes, and the discriminating test is *component
selected, boundary focused, arrows move the boundary and the component does not move.* Without that
assertion the two handlers are one regression from both firing.

**Sequencing correction.** DW-143 names 12.5's drag as the trigger for the first story that can strand a
component. It is **12.1** — the band-height *writer* strands it; the drag is a second door to the same
act. Re-price at 12.1's gate.

**Settled without escalation:** 12.4 is a **branch split, not a narrowing** — the four padding keys sit in
one `allowed` branch naming all **5** members of the closed element-type set, and the located refusal
already exists, with the four keys already in the canonical property order so the DataPath is right for
free. Byte-neutral: of 31 `.folio` files, exactly **4** contain padding and all four have it on element
`e8`, of type table. 12.2 AC2 is already satisfied engine-side (2 non-test format-context sites, its test
file untagged). 12.3 AC2/AC3 already satisfied and genuinely gated (7 tests across two untagged files,
checked by `head -3`).

**One correction to an AC about to be specced.** A 12.3 red-first test asserting *"nothing writes
`headerStyle`"* would be **false at this tree** — the font-chain rename cascade rewrites
`HeaderStyle.FontFamily` at 2 sites. Scope it to *no command **authors** `headerStyle`*, and **name the
rename cascade as the known writer inside the test**, so the next person to widen it trips on a named
exception rather than a mystery. **An assertion that is false at the tree it ships on is not a red-first
test; it is a red test.**

**A gap in Epic 12's planning text, not a note.** `property-prose-height.test.ts` and
`design-contract.test.ts` assert over **raw source text** of `App.tsx`/`App.css` — exact allowlists, no
hex/rgb/hsl literal permitted in `App.css`, exact-once counts on two type tokens. D-000.9 filed that
blocker against Epic 14; **12.1's panel rows and 12.5's drag CSS fire it just as hard** and Epic 12's text
does not mention it. Re-derive what those files assert at each plan gate rather than inheriting today's
reading — they are the least stable contracts in the repository.

---

### D-000.17 — An instruction to a subagent can be lost in delivery, and only the subagent can notice

**What happened.** I sent Story 16.11's builder a correction withdrawing a settled decision. The tool
confirmed it queued for delivery. **It never appeared in the builder's transcript.** The builder then
wrote a spec, referred in passing to "your correction's arm-2 assertion", and I replied *accepting that
description* — agreeing with a paraphrase of my own message rather than with the message. The builder
caught it, said plainly *"there is no such correction message anywhere in my transcript … I should not
have spoken as though I had read it"*, and marked the affected task as its own reconstruction rather than
as instructed work.

**Why it did not cost anything this time, and why that is luck rather than process.** The builder had
independently derived the same shape from measurement, and its reasoning reached the sharper half I had
not written: priming the element to another family can only pick an `AVAILABLE LOCALLY` row, every one of
which carries a `source` and therefore *embeds* — so priming both breaks the fixture's no-confound rule
and records `EMBEDDED` for a family that took the plain-commit branch. Two independent derivations agreed.
**Had they disagreed, I would have approved a spec believing it carried a ruling it did not.**

**The failure mode is specific and nasty: a paraphrase reads exactly like a receipt.** "Your correction's
arm-2 assertion" is what an agent writes whether it read the correction or reconstructed it, and the
orchestrator cannot tell the difference from the outside. Confirmation of *queuing* is not confirmation of
*delivery*, and delivery is not readership.

**Consequences.**
- **A ruling that withdraws or reverses an earlier instruction must be echoed back before it is acted
  on.** Not paraphrased — quoted, or at least summarised in the agent's own words *with the instruction
  named*, so a reconstruction cannot pass for a receipt.
- **When an agent refers to an instruction, check it against what was actually sent** rather than against
  what it plausibly was. I did not, and the builder rescued it.
- A subagent that flags a discrepancy in its own inputs is doing the highest-value thing available to it.
  This one did it twice in one story — here, and on the false "issues no command" premise.

**How we'd know it was wrong.** If echo-backs became ceremony on routine instructions. Scope it to
withdrawals and reversals, which are the ones where acting on the stale version is silently wrong.

---

### D-000.18 — `$?` is clobbered by ANY intervening command, including `echo`

**A sharpening of D-000.15's family, found by 16.11's builder on its own work, in the story about false
zeros.** The standing rule so far was *read exit codes from `$?` directly, never through a pipe.* That is
necessary and not sufficient. `$?` holds the status of **the immediately preceding command** — so this
reports the status of the `echo`, not the search:

```
git grep ... :!_bmad-output
echo "--- results above ---"
rc=$?          # <- this is echo's 0, not git grep's 128
```

The builder's `git grep` had **exited 128** — `:!` is "unimplemented pathspec magic" in this git — and the
harness printed `rc=0`. **A search that failed to run reported success with no output**, which is
indistinguishable from a search that ran and found nothing. It re-ran with `:(exclude)` and the exit code
read immediately, and only then had a measurement.

**This is the same defect as the three false-zero mechanisms already recorded, arriving by a fourth
route.** Recursive `grep`'s arbitrary misses, `git grep`'s tracked-only default, a `-run` filter matching
nothing and exiting 0, and now **a status silently overwritten between the command and the read**. None
share an implementation; all four render *"I did not look"* identically to *"it is not there."*

**The rule.** Capture the status on the same line or the very next one, before anything else runs —
`cmd; rc=$?` — and prefer `if cmd; then … else … fi`, which cannot be clobbered at all. When a search
returns zero, **the zero is only a measurement if the command's own status was 0**; a non-zero status
means the search did not answer the question, whatever it printed.

**Why it is worth its own entry rather than a footnote.** The agent nearly reported an unverified premise
as verified, and said so unprompted. The near-miss is the useful artifact: the wrong reading was
*plausible and quiet*, and nothing downstream would have questioned it — the premise was one I had
supplied and was expecting to see confirmed. **A measurement that confirms what the orchestrator already
believes gets the least scrutiny and therefore needs the most.**

---

### D-12.A.1 — Correction to D-12.A: `design-contract.test.ts` does not read `App.tsx`

**The claim was false and I propagated it.** D-12.A's closing paragraph states that
`property-prose-height.test.ts` **and `design-contract.test.ts`** assert over the raw source text of
`App.tsx` and `App.css`. The second half is wrong. Caught by Story 12.4's builder at its own plan gate,
and verified here before correcting:

`design-contract.test.ts` contains **zero** occurrences of `App.tsx` (grep rc=1). Its six handles are
`DESIGN.md`, `tokens.css`, `App.css`, `package.json`, `package-lock.json` and `tsconfig.app.json` —
enumerated, CLOSED. Positive control on the same file and the same query shape: `App.css` hits **3**
times. So the zero is a measurement, not a dead grep.

**The real constraint is narrower and sharper than the one I recorded**, and 12.4 would have been specced
against the wrong hazard. The genuine `App.tsx` source-text readers are:

- **`property-prose-height.test.ts`** — no assertion in it can fire on a comment rewrite.
- **`canvas-authority-contract.test.ts`** — scans raw text **without stripping comments** against a fixed
  prohibited-token list. **So a comment can break that scan.** 12.4 replaces a comment in `App.tsx`, and
  the replacement must not spell `getComputedStyle(`, `measureText`, `offsetWidth`, `ResizeObserver` or
  any other prohibited token — in prose that is *about* what the panel does not do, which is exactly the
  prose most likely to name them.

**How the error was made, because the mechanism matters more than the fact.** The survey classified two
test files together because they occupy the same *category* — "source-text contracts over the designer" —
and I carried the pairing forward without checking that both members had the property the category
implied. **A category is not a measurement.** Neither is a plausible pairing: the two files really are the
same kind of test, and that is precisely why the wrong member was never questioned.

**Consequence.** D-12.A's final paragraph is superseded by this entry. Its instruction — re-derive what
those files assert at each plan gate rather than inheriting today's reading — stands and is, if anything,
vindicated: 12.4's builder did exactly that and found the error.

---

### D-12.A.2 — Correction to D-12.A.1: the comment scanner DOES strip comments

**My correction was itself wrong, in the opposite direction.** D-12.A.1 replaced one false claim about the
designer's source-text contracts with another. It said `canvas-authority-contract.test.ts` *"scans raw
text **without stripping comments**"*, and warned that 12.4's replacement comment could break it by naming
a prohibited token.

**False.** Verified: `withoutComments` is defined in that file and applied at **both** scan sites — the
per-file scan and the composed exception chain. It is quote-aware. **A comment cannot trip the prohibited
token list.** Found by Story 12.1's builder, which was told to re-derive rather than inherit and did.

**Two further attributions in D-12.A.1 and the 12.1 dispatch were also wrong**, both verified now:
- The **hex/rgb/hsl literal rule and the exact-once type-token counts live in `design-contract.test.ts`**,
  not `property-prose-height.test.ts`. Which is doubly awkward, since D-12.A.1 exists to correct a claim
  about `design-contract.test.ts` — I moved the file out of the picture and then attributed its rules to
  its neighbour.
- The `borderWidth` positive control is **6** designer files, not 7. The seventh was a compiled
  `src/generated/runtime/*.wasm` blob matching under `grep -a`. The zero it controls for is still real and
  CLOSED over 109 files. **A binary artifact inflated a control, which is the quieter half of the
  false-zero family: a control that fires for the wrong reason licenses a zero it never tested.**

**Also corrected, from the same report:** `internal/diag/diag.go` has **17** `Code =` declarations, not the
19 recorded in D-12.A. Corroborated four ways. **No band-overflow code either way, so every conclusion
drawn from it stands** — only the number was wrong.

**What this says about the survey, and it is the useful part.** Three separate facts about the designer's
source-text contracts have now been wrong in the record, and each was corrected by the *next* story that
had to act on it. The pattern is not carelessness about any one file; it is that **a survey classifies by
category and the category is where the error hides**. "Source-text contracts over the designer" is a true
grouping whose members do materially different things, and every claim I made about the group was
inherited by its members without being checked against any of them.

**Consequence, sharper than D-12.A's original instruction.** Re-deriving at each plan gate is necessary
and has now paid three times. But the rule it implies is stronger: **a claim about a group is not evidence
about a member.** When a survey says "these files assert X", the spec that acts on it names the file and
the line, and the builder confirms that file does X — not that the group does.

---

### D-12.B — Story 12.1: both halves of D-12.A's Fork 1 ruling were false, and the story got smaller

**Falsified by 12.1's builder at its own plan gate, verified independently by the lead and by me.**

**The engine half.** D-12.A ruled *"the engine refuses — `containComponent` already does this and needs no
change."* Measured: all **11** call sites are in `component_commands.go`, every one a component command.
**`Canvas` never calls it.** A band-height writer is not on that path. Proved by a throwaway module
outside the repo (exported API only, deleted, tree verified clean): a `pageHeader` of 20pt holding `e1` at
`y=50 h=30` gives `ParseTemplate` nil, `Canvas` nil, a well-formed projection, and the TS guard **admits
it**. Nothing refuses the strand.

**What is true is worse, and it strengthens the conclusion while destroying its mechanism.** On that
stranded element: `moveComponent` refused; `resizeComponent` refused **even shrinking to `h=1`** (`y=50 >
20` fails before the new size is considered); and `updateComponentProperties` setting only **`color`**
refused with *"component geometry must stay within pageHeader"*, where the identical command succeeds on
the unshortened document. **A strand poisons edits that carry no geometry at all.** Only
`setComponentBounds` and `deleteComponent` still work.

**The panel half.** D-12.A's floor was justified by an engine refusal that did not exist — and is
independently forbidden. Story 17.4's item 5 was cited as an open question; **item 5 was amended in place
the same day and answered by item 9**, which rules *do not clamp*. Its distinction is **not** *mirrored
bound vs. only copy* — that was the orchestrator's misreading — but:

> *"**The floor and the ceiling are not the same kind of bound, which is why one is mirrored and the other
> must not be.** Zero and negative are values these fields can NEVER legally hold — whatever document is
> open… That fact is a property of the FIELD… A band-edge ceiling is a property of the LAYOUT: it depends
> on the component's position, its band's height, and the page."*

A floor at `max(y + height)` is **layout-dependent by that definition's own terms.** Item 9 explicitly
assumes the engine refuses and *still* forbids the clamp, because the objection is the browser computing
layout geometry at all. **So the panel does not floor: it sends, the engine refuses, the existing
`role="alert"` path renders it.** Consistency with typing is the property asserted.

**RULING.** (1) Add the check to the **band-height writer**, reusing `containComponent` as the predicate
over a candidate band — one new call site, the predicate itself unchanged, no new diag code. **The
refusal must name the act and the stranded element, not "component geometry must stay within pageHeader";
the author moved no component.** Reuse the predicate, not its sentence. (2) **A new `setBandHeight` arm on
`ApplyComponentCommand`**, not a widened page setup — because `Canvas`'s invariant refusals begin `"folio:
page setup "` (space) while the command door's own validations begin `"folio: page."` (dot), and
`pageSetupDiagnostic` discards the message for anything not carrying `PAGE_SETUP_INVALID`. **The command
door's validations are located; the projection's invariant checks are not.** Plus page setup's strict
`len(raw) != 7` arity, which a new arm does not disturb. (3) The new command validates the content-window
invariant and refuses located; `Canvas`'s bare refusal stays as the **hand-edit backstop** — two guards,
two audiences, authoring versus loading. **Condition: both call ONE predicate.** Two implementations of an
invariant is the drifting copy 17.4 warns about; two call sites of one function is not. (4) **The panel
floor is deleted**, and with it D-12.A's D-7.4.5 mirror obligation and its unbuildable red-proof.

**Two rules from the lead's own errors, each worth more than the correction.**

**A predicate that guards N existing doors says nothing about a door that does not exist yet.** The
forcing relation was asserted, never checked — and the tell was inside the story: **12.1 is the story that
adds the writer**, so no existing call site could have been on its path. When a ruling says "X already
handles this", the check is whether X is reachable from the *new* code, not whether X is thorough.

**A code comment is a claim about a decision, never the decision.** The lead read
`component-property-command.ts`'s comment calling 17.4's item 5 an open question and treated it as the
record. That is **D-000.10's failure mode arriving inside a ruling that cites D-000.10.** When a comment
says "open question", read the story's Spec Change Log before building on it.

**Registered, not built:** `pageSetupDiagnostic` discards the engine's message — every `"folio: page setup
"` refusal reaches the author as a fixed sentence about size and margins, the real reason thrown away.
Independent of this story, and the reason the door choice was forced. And the strand-poisons-edits
finding, **with its escape named** (`setComponentBounds` and `deleteComponent` still work), so it reads as
a disclosed shortfall rather than a trap.

**DW-143 does not fire here.** Its discriminator is off-*page* geometry; a component stranded in a
shortened header renders **on** the page, overlapping content. Different defect. D-12.A's re-pricing
instruction is struck, owner unchanged.

**Implementation condition:** the candidate-band check must evaluate **every** component in the band, not
a selected one. A check that examines one is the vacuity this run keeps finding.

---

### D-000.19 — Epic 16's third e2e failure is a HANG, not a slow test. Measured, and the gate can now conclude.

**The gate's last open item is closed.** `e2e/browser-native-roundtrip.spec.ts` — *"fresh authored
sessions close exactly through admitted Preview and native Folio"* — was failing at its own
`test.setTimeout(300_000)`, and nothing distinguished *genuinely deadlocked* from *merely slower than its
budget*. That distinction decides whether it is a defect or a configuration line, so it was worth one
measurement rather than a story (the lead's ruling, and it was right).

**Measured by raising the test's own timeout — the CLI cannot override `test.setTimeout`, so the file was
edited, run, and restored from an absolute-path backup; `git status` confirms it clean.**

| budget | result |
|---|---|
| 300 s (shipped) | FAIL, timeout |
| **1500 s (5× the budget)** | **FAIL, timeout** |

**And the resource profile names the mechanism.** Over 27 minutes of wall clock the run consumed **152 s
user + 5 s system — 9% CPU**. A test that is merely slow burns CPU; this one **waits**. It is blocked on
something that never arrives, not grinding through work it will eventually finish.

The captured page snapshot puts it in DESIGN mode with *"Unsaved local changes"* — mid-authoring, before
the save/preview/native-render comparison the test exists to perform.

**Why this matters beyond one test.** It is the heaviest end-to-end assertion in the repository — it
builds the Go CLI, authors two full documents through the real UI, saves both, renders each in the browser
*and* natively, compares the PDFs byte-for-byte, and runs a Go witness test. **It is the only thing that
proves the browser and the native binary agree on a document a human actually authored.** While it hangs,
that guarantee is unverified, and DW-193 (the e2e suite is compiled in CI and never executed) means
nothing was ever going to tell us.

**What is NOT yet known, stated so the next reader does not over-read this:** where it blocks. The
snapshot places it mid-authoring but the trace was not decomposed to the failing action, and no bisection
of the authoring helpers was run. **"It hangs" is measured; "it hangs at X" is not.**

**Consequence for the Epic 16 boundary gate.** The gate can now conclude, and it concludes with three
findings rather than two: the accounting guard was vacuous (fixed, 16.11), two assertions had gone stale
(fixed, 16.11), and **the deepest cross-boundary test in the repository is hung** — pre-existing, not
Epic 16's doing, and not repaired here. The gate does **not** pass. Filing the entry and naming an owner
is Epic 16's obligation; fixing it is not.

---

### D-12.B.1 — Correction to D-12.B: the TypeScript guard DOES drop the snapshot, and the consequence is a dead editor

**Third correction in this family, and this one reverses a correction I accepted against the original
record.** D-12.B recorded, on 12.1's builder's measurement, that a stranded component *"gives a
well-formed projection and the TS guard **admits it** — the snapshot is not dropped."* **False.**
D-12.A's original claim was right, and I was talked out of it.

Verified at `engine-protocol.ts:345`:

```
if (BANDS_CAPPING_VERTICALLY.includes(component.band as string) && !(box.y + box.height <= band.height)) return false
```

For the exact strand measured (`y=50000, h=30000, band.height=20000`): `80000 <= 20000` is false, so
`isCanvas` returns false, `parseInbound` returns undefined, and `EngineClient.#fail('PROTOCOL_INVALID')`
**terminates the worker.**

**The builder's account of its own error is the transferable part:** *"I based my claim on the
numeric-range check fourteen lines earlier, as reported by a surveying agent, and never read the vertical
clause — a survey passed off as a measurement."* That is **D-12.A.2's rule biting the agent who reported
it**: a claim about a group is not evidence about a member, and a surveyor's summary of a guard is not the
guard. The clause was fourteen lines away.

**What it breaks, and what it strengthens.** A frozen Always clause justified the send-only-if-changed
rule by *"an already-stranded hand-edited document can still accept a margin change."* **That document
cannot be opened in the designer at all.** The AC asserting it was green only because `App.test.tsx` mocks
the engine and never runs the real guard — measured: **0** occurrences of `isCanvas`/`parseInbound` in
that file, against **305** engine references as the positive control.

But the correction makes the story's case far stronger. Without 12.1's engine check, a band-height command
creates a strand and **the very next projection terminates the worker mid-session** — the editor bricks
itself with no attributable error. *"Later commands are refused"* was an inconvenience; **this is data
loss shaped like a crash**, and it is now the true reason the check must exist.

**Ruled:** strike the false justification and its AC; **keep** the send-only-if-changed rule on the honest
ground that it avoids needless commands and history entries — *a rule can be right for a reason other than
the one first given*; replace the engine check's stated reason with worker termination. One narrow frozen
edit, no revert: the root cause is a justification, not an implementation, and the 309 implemented lines
are untouched by it.

**And the same review found the sixth instance of the run's dominant defect.** Two layers independently
found the `pageFooter` arm is never exercised in Go. Mutation run and confirmed: **deleting the
band-pointer swap, so a footer command writes the header, leaves the entire Go suite green** except the
mandated red. Setting the page-footer height would silently resize the page **header**. Every accepting,
strand, boundary and content-window test uses `pageHeader`; the single `pageFooter` row returns above the
swap. **This is Story 12.4's key-to-edge rotation, one story later, in a story explicitly warned about
it.**

**The standing lesson, now paid for three times in one family:** every correction in this thread was made
by the story that had to *act* on the claim, never by the story that recorded it. Records are checked by
use, not by review — which is an argument for specs citing symbol and line so the next actor can fail to
find them.

### Re-grounding refresh — 2026-09-06 (third session)

*Filed by the orchestrator from the lead's report. Re-grounded from this section, its 2026-09-05 refresh,
and the rulings below — not re-derived from the spine, the ADRs or the epics doc. Verified at HEAD
`429cb1a`, working tree **clean**, re-measured with `git rev-parse HEAD` and `git status --porcelain`
rather than inherited from the session's opening `gitStatus`, which showed `ffec48c` and 21 modified + 14
untracked paths. **That opening snapshot was two commits and one story stale** — anyone reading it as
current would have concluded 11.1 was mid-flight. Third session in a row the opening snapshot has been the
grounding hazard; treat it as never authoritative.*

**State.** Epic 11 `in-progress` — 11.1 `done`, 11.2/11.3 `backlog`, 11.4 added this day by owner
decision. Epic 12 `done` (5/5). Epic 13 `backlog` (5). Epic 14 `backlog` (10). Epic 15 `in-progress` —
15.1/15.2a/15.2b `done`; 15.0, 15.2, 15.3 `backlog`, **plus 8.4d and 8.4k**.

**Invariants carried, not re-derived.** I-1 no rendered byte is a function of the build environment; I-2 a
weight is a face or it does not exist, no synthetic bold/oblique at emit **or** on the canvas; I-3 AD-17,
the browser never measures, machine-enforced with five named-owner exceptions; I-4 AD-15, the engine owns
the document; I-5 a control the panel declines to author must not delete what the document carries; I-6
AD-21; I-7 AD-26; I-8 AD-14, diagnostics are one type on one channel from an additive registry; I-9
`Render`'s error contract is not negotiable from the designer.

**Three things verified in code rather than inherited:**
1. **`fontChain` already serves both the render path and the canvas projection.** `render.go:1148` has
   three non-test callers: `render.go:786`, `table_render.go:866`, and **`page_setup.go:1347`, the canvas
   text-paint projection.** So 11.3 needs no new resolution path — it inherits 11.2's.
2. **`chainFaceNames` is the single boundary and is style-blind by construction** (`render.go:1216`, three
   non-test call sites). Its comment and `embedded_face.go`'s opening block record that ten functions
   consume the chain and none can reach a `*Template` — which is why Story 8.4 put name→bytes behind the
   `fontCache` instead of widening six signatures. **That argument transfers verbatim to "what is a bold
   entry".**
3. **The long-carried "15.0 is tracked but has no epic text" discrepancy is CLOSED** — `epics.md:5133`
   now carries it in full, and `:5139` records the sequence 15.0 → 8.4d → 15.3 (D-8.4d.2).

**One record discrepancy the lead found, and it was a release-gate risk.** `sprint-status.yaml` routes
`8-4k` and `8-4d` to Epic 15 (`deferred-to-epic-15`, `moved-to-epic-15`) while **Epic 15's own story rows
enumerate neither**, and both stories' text lives in Epic 8's section. So Epic 15's list under-counted its
own work by two — and 8.4d is explicitly sequenced **before 15.3 cuts the tag**. Closed the same day by
adding pointer comments to Epic 15's rows; the authoritative rows stay under Epic 8 so no story carries two
status values.

**Ordering risks across 11/13/14/15, as the lead framed them:**
- **11.2 is the run's only hard deadline dependency.** It spends format freedom — the chain entry gains
  keys — and that freedom expires when 15.3 cuts `folio-go/v0.1.0` (D-000.14 **in
  `epic-8-15-decision-log.md`**, not this log's D-000.14; per D-000.23 a bare citation names two rules).
  11.2 and 11.3 must both close before 15.3, and today that is satisfied by build order rather than by
  anything that checks it.
- **11.3 before Epic 14.** 11.3 touches the TYPOGRAPHY B/I pair that Epics 16 and 17 rebuilt and Epic 14
  rebuilds again. Running it first means Epic 14 inherits a third state on the B control as a *behaviour to
  preserve*, which is cheaper than adding it afterwards.
- **Epic 13 before Epic 14, and 13.4 first within Epic 13**, against numeric order — `App.tsx`'s
  sample-data gate blocks every preview without a JSON file.
- **DW-192/DW-193 remain the live masking hazard over Epics 13 and 14** — the pointer carve-out waives
  every AD-17 prohibition inside `App.tsx`, and the e2e suite is compiled and never executed. Fifteen
  designer stories' problem, not Epic 11's.

**Confidence, stated with its limit.** High on Epic 11 — the lead read the code, not only the log. **Medium
on Epics 13 and 14**: the 2026-09-05 survey's findings were re-read but **not re-measured**, and D-000.4
requires line anchors from that survey to be re-derived by symbol at each plan gate. **Every Epic 13/14
anchor in the original grounding report is SAMPLED until its own plan gate re-derives it.**

---

---

## D-12.C — The loose regexp is a breach of a comment's promise, not a disagreement between peers

**Story 12.2.** The builder held three tasks at `[HELD-B]` on one question: `setDocumentUTCOffset`
accepts what the loader accepts, but the loader's `utcOffsetPattern` admits `+99:99` — a value the
evaluator's `parseUTCOffsetMinutes` then refuses at render. Three options: **(a)** the command reuses
the loader's predicate as-is and ships a control that can author an unrenderable document; **(b)** the
command range-checks on top, becoming stricter than the loader; **(c)** repair the loader's pattern.

**Ruled: (c) then (a).** The fork *dissolves* rather than being decided. Repair the predicate, then have
the command reuse the repaired predicate — one spelling of the document rule, in one place, with command
and loader agreeing by construction. `IsUTCOffset` is the single exported predicate both doors use.

**The builder argued (a) on the ground that (c) narrows what the format accepts and is therefore the
owner's call. It is not, and the measurement is unusually clean:**

| Claim | Measured |
|---|---|
| A compatibility clause in `folio-format.md` | **It exists, at `:67-72`** — see the correction below. My absence query returned **0**, rc 1, and was wrong. |
| A compatibility promise anywhere else | `SPEC.md:248` lists it as an *open question*; `epics.md:148` NFR6 states *"Forward/backward compatibility rules undefined and need a policy before v1."* Those describe *template-vs-library* rules, not the format's own increment rule. |
| The format's own rule | `folio-format.md:49` — *"`utcOffset` \| Fixed offset, `±HH:MM`."* |
| Real documents excluded by (c) | **0** of 28 `.folio` files — 21× `+00:00`, 7× `+07:00` |

`+99:99` is not a fixed offset. **So (c) does not narrow the format — it implements it.** The regexp was
a loose rendering of a spec that was never loose. And there is no promise to break: the repo records
compatibility rules as *explicitly undefined pending a policy*, which is a stronger result than the
absence of a clause — the absence is deliberate and written down.

**The real defect is a comment that made the gap invisible.** `parseUTCOffsetMinutes`'s doc comment says
the document field is *"already syntax-checked at load."* It is not. So `+99:99` reaching the evaluator
is a **breach of a documented contract**, not three peers disagreeing about strictness. The comment
becomes true only after this story, and must be updated in it — **[[D-000.10]] applies to
accurate-*sounding* comments as much as to stale rulings.**

**`Z` is not a divergence, and I was wrong to frame it as one.** `parseUTCOffsetMinutes` serves **two
populations with one function** — the document's `utcOffset` and offsets embedded in RFC 3339 timestamps
arriving in report *data*. Verified: three call sites, two on the RFC 3339 path (`formatdate.go:196`,
`:237`) and one on the document field (`:313`). `Z` is RFC 3339's canonical UTC spelling and the data
caller *requires* it. The sets do not overlap imperfectly; one function correctly admits the **union of
what its two callers need**. Removing it breaks `Z`-terminated timestamps in bound data; adding it to the
loader is worse — **two byte-spellings of one semantic offset is a canonicalization hazard in a
byte-identity product**, and round-trip identity would depend on which the serializer picked.

**Carried into the story:** the red-proof is the **reachability**, not the regexp — `+99:99` must move
from *"loads, then every date fails at render"* to *"refused at load with a located message"*; the
anti-divergence guard must **name `Z`** as the one expected asymmetry with its reason stated in the
assertion, or it hides the fact this decision uncovered; any existing test pinning the old looseness is
corrected and reported, since such a test would mean the looseness was once observed and accepted; and
the evaluator's document-path check at `:313` **stays** — [[D-12.1]] established `Canvas(t)` never reads
`UTCOffset`, so there is no backstop, and **what this run punishes is redundant checks across an
authority boundary, not within one package**.

**A charter that constrains the fix rather than the outcome is a charter being read wrong.** The builder
invoked 12.2's own charter — *"every value it makes authorable is one the loader already accepts"* — to
justify (a). Under (a) alone that charter is satisfied in **letter** while the control ships a value that
destroys the document's rendering; it was doing work only because the loader was doing too little. Under
(c)+(a) it is true in **substance**. That is the generalizable point: a constraint phrased against the
mechanism can be met by a fix that misses the purpose.

### D-12.C.4 — the canonicalization argument against `Z` was false, and `-00:00` falsifies it

**Caught by 12.2's review layers.** D-12.C argued that admitting `Z` at the loader would be worse than
excluding it because *"two byte-spellings of one semantic offset is a canonicalization hazard in a
byte-identity product — the serializer would have to pick one, and round-trip identity would depend on
which."*

**That hazard already exists and is already accepted.** Measured against the repaired pattern
`^[+-](?:[01][0-9]|2[0-3]):[0-5][0-9]$`:

| Value | Admitted |
|---|---|
| `+00:00` | **yes** |
| `-00:00` | **yes** |
| `+07:00` | yes |
| `+99:99` / `+24:00` / `Z` | no |

**`+00:00` and `-00:00` are two byte-spellings of one semantic offset, and the loader takes both** — both
before this story and after it. My argument therefore proves too much: taken seriously it would condemn
the pattern D-12.C just ruled correct.

**And the reason the hazard never materialised is the reason it was never a hazard.** This format's values
travel **verbatim** — the serializer does not normalise an offset, so a document declaring `-00:00`
round-trips as `-00:00` and byte identity is preserved without anything having to choose. **A
canonicalization hazard requires a canonicalizer.** There isn't one, which is precisely why two spellings
have coexisted unremarked.

**The conclusion survives on the grounds it should have rested on all along:** `folio-format.md:49` says
*"Fixed offset, `±HH:MM`"*, and `Z` is not that — so excluding it from the loader **implements the
format**, exactly as excluding `+99:99` does. `Z` stays in the evaluator because RFC 3339 data requires
it. Nothing about the outcome changes; a surplus argument that was wrong has been removed.

### The pattern worth naming: four corrections to one ruling

D-12.C has now been corrected four times — [[D-12.C.1]] (the compatibility clause exists), [[D-12.C.2]]
(the version precedent exists, twice, and states the principle), [[D-12.C.3]] (the corpus was 31, not 28),
and this one. **Every correction came from the party that had to act on the claim, and none from review of
the claim itself.** The ruling's *outcome* has survived all four intact, which is the reassuring half. The
worrying half is that **each wrong reason was independently sufficient-sounding**, and three of the four
were mine.

The common shape: I reached for an argument from a general principle — compatibility clauses, narration
precedent, canonicalization — **without measuring whether the principle's precondition held in this
codebase.** A canonicalization hazard needs a canonicalizer. A missing precedent needs the file read to
the end. An absent clause needs a query in the document's own vocabulary. **Each time the general
principle was sound and its applicability was assumed** — which is the same defect as
[[D-12.C.2]]'s citation-without-its-boundary, arriving one level up: not a claim about a line, but a claim
about a rule.

### D-12.C.3 — the corpus number, corrected by the story that used it

**D-12.C records 28 `.folio` files, 21× `+00:00`, 7× `+07:00`.** 12.2's implementer re-measured over
every `.folio` that `git ls-files --others --cached` reports and got **31 files, 24× `+00:00`, 7×
`+07:00`, zero out of range** — the extra three are the designer's built copy, the generated runtime, and
the Go example, and two malformed testdata files sit inside the 31.

**Same conclusion, larger population, and the direction is the reassuring one:** the widened count adds
only in-range values, so "nothing real is excluded" is now measured over more of the tree rather than
less. The builder's note carries its own measurement and cross-references D-12.C's narrower count instead
of overwriting it, which is right — **the narrower number was not wrong, it was a smaller population**,
and that is the distinction [[D-15.2.1]] exists to preserve.

Worth naming that the original 28 came from a tracked-only enumeration and the 31 from one that includes
untracked files. **This is [[D-15.2.1]]'s bias caught in the act, in the same story that provoked the
rule**, and it moved a number this run had already cited three times.

### D-12.C.1 — the clause is real; the ruling holds because it is inapplicable, not absent

**Corrected the same day, by the builder whose citation I had just called imaginary.** I told it *"the
clause you cited is not in the document you named it in."* It is, at **`folio-format.md:67-72`**, and I
verified it myself rather than take the correction on trust:

> *"A MINOR increment may add **new optional keys** only. It may **not** change the meaning of an existing
> key, and it may **not** extend a closed set of legal values (element `type`, `locale`, `style.align`,
> `columns[].align`, `valign`, `columns[].footer`, `border.edges`, `page.orientation`, `page.size`).
> Extending a closed set is a **MAJOR** change, because every existing library validates those sets as
> load errors."*

It is **stronger than the builder quoted** — the next paragraph records a worked precedent: *"`style.align`
gained `justify` (Story 7.3, FR47), and the format is therefore at `2.0`"*, with the additive-boolean
alternative rejected because an older reader *"would have ignored the unknown key and drawn the paragraph
ragged while believing it had rendered the document correctly."*

**How I produced the false zero: a vocabulary-based absence query.** I searched
`compatib|backward|will always load|never reject`. The clause is written in **versioning** vocabulary —
`MINOR`, `MAJOR`, `increment`, 8 hits in that file — and contains not one word from my phrase list. **My
positive control did not protect me**: `utcOffset` hit 3 times, proving only that the file was being read,
never that my *query* could see the thing I was looking for. This is [[D-000.20]]'s lesson one turn later
and one level deeper: **a positive control validates the channel, not the vocabulary.** An absence query
needs a control on a *synonym of the thing sought*, not on any token in the file.

**The ruling is unchanged, and now rests on a better reason.** The clause governs **extending** a closed
set. This story **narrows a loose regexp** — the opposite direction — and `utcOffset` is not among the nine
enumerated sets. More decisively: `folio-format.md:49` already specifies *"Fixed offset, `±HH:MM`"*, so the
format is not changing at all and no increment is in question. **(c) is right because the clause is
inapplicable, not because it is absent.** Filing it as absent was the error.

**Why the correction could not be left to stand as a footnote.** As first written, D-12.C told the next
reader that `folio-format.md` carries no compatibility rule. The next person deciding whether to add a
fifth `locale` tag or a sixth `border.edges` value would run the same vocabulary, get the same zero, and
conclude they were free. **They are not** — that change is MAJOR, and the format has already paid the cost
once. The false premise licensed **precisely the change the clause forbids**, in the opposite direction
from the one this story travels. A wrong reason that reaches the right answer is not harmless when the
reason is what gets reused.

**And the builder's own citation was wrong in the mirror-image way.** It cited a real line and asserted it
said something it does not say — it read the clause as protecting `+99:99`'s admission, when the clause
never speaks to narrowing. So the same two lines were misread twice, in opposite directions, by both
parties to the decision. **The record now names the lines and states what they do and do not govern**,
which is the only form of citation that survives reuse.

**One reading confirmed for reuse.** The builder was right that *"the command must not be stricter than
the loader"* is **not** this project's rule — [[D-12.4.1]] set the opposite two stories ago. It does not
decide this one, because under (c)+(a) the command is not stricter. But it is the answer the next time
that argument is reached for.

---

## D-000.20 — The shell's `grep` prints no matches for `App.tsx` while counting nine of them

**Not a story decision — a measurement defect in the orchestrator's own tooling, found while checking
[[D-12.C]]'s neighbourhood.** Recorded because it invalidates a class of evidence used across this run.

I greped `folio-designer/src/App.tsx` for `pageSetup` and got **nothing**, and briefly concluded the page
setup panel did not exist in the product at all — a conclusion that would have made Story 12.2's first
acceptance criterion ("the page setup panel, which is where document-level settings already live") false
and sent the story back for rescoping **mid-implementation**. It is not false. The panel is at
`App.tsx:1682`, `applyPageSetup` at `:780-823`, `pageSetupDiagnostic` at `:3000` — exactly where 12.2's
Code Map already said they were. **The story's record was right and my measurement was wrong.**

**Measured, on one file, in one shell, seconds apart:**

| Command | Result |
|---|---|
| `grep -c "pageSetup" App.tsx` | **9** |
| `grep "pageSetup" App.tsx` | **no output**, rc 1 |
| `grep -n "pageSetup" App.tsx` | **no output**, rc 1 |
| `grep -in "pageSetup" App.tsx` | **no output**, rc 1 |
| `command grep -n "pageSetup" App.tsx` | **9 lines** |
| same shimmed `grep -n` on `page-setup-command.ts` | **2 lines**, correct |

`grep` in this shell is **not the binary** — it is a shell function from the session snapshot that execs
`ugrep` with `--ignore-files --hidden -I`. On this file it **counts correctly and prints nothing**. The
file is not gitignored (`git check-ignore` rc 1), and the matching lines are short (40–394 chars),
so neither ignore rules nor per-match length explains it. **I could not identify the mechanism and am
not asserting one** — the file does contain a 2196-char line, which is a candidate and nothing more.
Recording an unexplained reproducible defect beats recording a tidy wrong cause; I have overstated a
finding twice in this run already.

**Why this is the worst false-zero mechanism of the five.** The four before it announced themselves in
principle — recursive `grep`'s arbitrary misses, `git grep`'s tracked-only default, a `-run` filter
matching nothing, `$?` clobbered by an intervening command. Each is a rule you can learn and then apply.
**This one is file-specific and silent**: the same command, same flags, same shell, is correct on one
file and empty on another, so no amount of care with flags protects you, and a positive control run
against a *different* file passes while the measurement that matters returns zero. It is the run's
dominant defect class — a guard that cannot distinguish a correct outcome from a plausible wrong one —
now degenerate: **the guard reports the one answer that always looks like a legitimate finding.**
"Zero occurrences" reads as evidence. It was noise.

**Standing rule, effective now:** every negative grep result over the designer sources is worthless
unless it carries a **positive control on the same file** — `-c` on a token known to be present. Prefer
the `Grep` tool or `command grep` over the Bash builtin for any measurement a decision rests on. This
is the direct generalisation of the rule this run keeps re-learning: **a zero is a claim, and a claim
needs a control.**

**One thing the false zero got right by accident, and it stands.** The `epics.md:4107` **Design pointer
for 12.2 is genuinely wrong** — measured over the mockups directory, which the shimmed `grep` handled
correctly (positive control: `Font Browser.dc.html` returns 1): **no mockup contains "page setup",
"locale" or "UTC offset"**, and `Main.dc.html`'s inspector draws PROPERTIES / POSITION / TYPOGRAPHY /
CONTENT / BANDS / COMPONENTS and no PAGE SETUP. 12.2's builder had already found this independently and
filed it in the spec at line 454. **The pointer is stale; the acceptance criterion it appears to support
is not** — the panel exists in the product, just not in the drawing. Correct the pointer, leave the AC.

---

## D-15.0 — Epic 15 is five stories, not two, and 15.0's spec predates three epics

**Ruled by the lead, verified by me against the tracker and the decision logs before acting.** Story
15.0 (`a catalogue face arrives when it is picked`) has a tracker slot, an OWNER DECISION behind it
(D-8.4d.1, 2026-09-02), a recorded position in the build sequence, and **no text in `epics.md`** — no
user story, no Covers line, no acceptance criteria. Five record files already depend on it landing.

### The tail was short by three, and the tracker says so plainly

I had been carrying Epic 15 as **15.2 and 15.3**. Re-derived from `sprint-status.yaml` rather than from
my own program list:

| Key | Status |
|---|---|
| `15-0-a-catalogue-face-arrives-when-it-is-picked` | `backlog` |
| `15-2-ci-s-red-means-something` | `backlog` |
| `15-3-folio-go-v0-1-0-is-cut` | `backlog` |
| `8-4d-the-size-budget-is-a-number-something-checks` | **`moved-to-epic-15`** |
| `8-4k-a-licence-exception-is-read-and-checked` | **`deferred-to-epic-15`** |

**Five.** The two I was missing hid behind *two different non-`backlog` statuses*, neither of which
carries the string `15-` in its key — so every list I had built by looking for Epic 15's stories was
structurally incapable of finding them. The tracker's own definitions are unambiguous:
`deferred-to-epic-<N>` is *"Terminal HERE, open THERE. The work is still owed, by a later epic."*
**A status that reads as closed in one place and open in another is only visible from the second place,
and nobody was standing there.** Re-derive the open set from the tracker, never from a program list,
and do it once more before 15.3 — an epic that has absorbed two stories from another epic is exactly
where a third hides.

**The sequence binds, and for a stated reason:** 15.0 → 8.4d → 15.3. 8.4d's subject *is* the size
budget, and 15.0 removes ~2.2 MB from the payload; setting the threshold first would set it against a
number known to be about to change. D-8.4d.1 says so outright — removing the catalogue still leaves
~13.5 MB and *"a threshold still has to be set, and the `~9 MB` figure still has to be superseded."*
**8.4d is not merely after 15.0; it is unanswerable before it.** 15.3 cuts the tag and is genuinely last.

### Do not draft 15.0 from the decision that created it

D-8.4d.1's surviving sentence — *"no Google Fonts API and no arbitrary URL — the catalogue remains the
only source, and nothing is fetched from a third party"* — is **explicitly struck** by a later reversal
in the same log, which permits three named hosts and says in terms that the sentence *"does not survive
this decision and must be amended in the same edit."* Drafting from D-8.4d.1 would **reinstate a struck
clause**. Worse, Epic 16 has since shipped a fetch tier for web faces, so a draft written from the
2026-09-02 record would re-specify work that already exists.

**So 15.0's spec must open with a boundary statement** — what the catalogue tier does *today* (bundled,
precached, 20-slot margin), what Epic 16's web tier *already* does, and what 15.0 *changes*. Without it
the builder cannot tell which half is built. **This is the stale-anchor failure at its worst case: a
story whose specification predates three epics.** Drafting is mine, gating is the lead's; the owner is
not needed, because the policy is determined and only the text is missing.

### The standing rule this produces

**A decision whose correctness depends on something NOT having happened must carry an executable
assertion of that absence, not a prose caveat.** `epic-16-decision-log.md:1427` already warns that
someone will later reverse a decision *"believing it was a budget call."* That warning is prose, and
**prose discharges only if a future reader happens to find it** — which is precisely what failed today,
twice, in [[D-000.20]] and [[D-12.C.1]]: a clause that was true when written, a query that could not see
it, and nothing obliging either party to notice.

Before 15.0 is dispatched it must carry a **tripwire**: a test asserting that no path fetches a
catalogue face on pick, **keyed on the capability rather than on a proxy** like a sprint-status value or
a filename, whose failure message names the two Epic 16 decisions it invalidates (`:789`'s precache
margin, `:1427`'s not-a-budget-call warning) and states what must be re-decided. It passes today because
the capability does not exist. **The moment 15.0 adds it, the tripwire goes red and cannot be merged
around silently.**

**15.0 discharges it by replacement, never by deletion** — its acceptance includes re-taking the
precache margin with the new payload in hand and amending `:1427` with the reason the decision was
actually made. *A tripwire deleted rather than replaced converts a caught event into an uncaught one,
and the deletion will look like cleanup to whoever does it.* The discharge list goes into 15.0's spec
**before** dispatch, enumerated from the five dependent records, so the builder discharges a list rather
than discovering one.

**Absence is a testable state right up until it stops being one, and that transition is exactly the
moment you need to be told about.**

### One reservation, carried deliberately

The lead ruled the owner's intent determined **by D-8.4d.1 as amended by the second reversal**. If the
draft reaches a point where those two records *disagree* about what 15.0 should do — rather than merely
layering — that specific conflict is escalated to the owner, not reconciled. **Two reversals of one
Non-goal in a single day is where an unrecorded third intent would hide.**

---

## D-12.C.2 — Three of us argued a version question without reading the document that governs versions

**Closing [[D-12.C]] / [[D-12.C.1]].** I ruled that `folio-format.md` narrates version *events* but has
no precedent for a version *non-event*, and asked 12.2 to write the first one. **That was wrong too.**
12.2's builder measured it, and I verified every line before accepting:

- **`:90-93` narrates a non-event** — *"**The trigger is the ENTRY, not the asset.** A document that
  carries a font asset no chain references still declares whatever its other content requires … so
  raising it would orphan a document from readers that can in fact read it."*
- **`:573-591` narrates another**, headed *"**THE VERSION TRIGGER AND `SupportedMajor` DO NOT MOVE, and
  the derivation is here rather than asserted.**"*
- **`:575` states the governing test**, which I had reconstructed from first principles: *"would a
  pre-`2.0` reader refuse this file, or render it wrong?"*
- **`:585-587` states my own principle in the document's own voice**, this case with the nouns swapped:
  *"Making the reader stricter about `2.0` documents is not a version trigger: a document has no way to
  declare 'I am missing licence text', so this is **reader strictness**, and **version describes the
  document, never the writer**."*

**So the note 12.2 writes applies a recorded principle instead of deriving a new one**, cross-referencing
`:585-587` rather than restating it. Its genuinely new contribution is narrow and should be stated as
such: the principle extends to a **pattern-constrained string**, not only to reader-strictness about
absent metadata, and this particular tightening excluded nothing real (28 files, 21× `+00:00`,
7× `+07:00`).

**A fourth ground none of us used, sitting at `:589-591`:** a bump *"would also make every document
declare `3.0`, including the twenty-two fixtures that make no font choice at all, moving their bytes and
their goldens for a reason unrelated to fonts."* **This one bites 12.2 directly** — its acceptance
already requires that no golden hash moves, so **an increment would have failed the story on its own
terms**, by a mechanism nobody in the thread had connected to the version question.

**And a standing owner ruling that is in neither decision log**, recorded only inside the format file at
`:585-591`: *"Folio is unreleased; the format may be broken (owner ruling, 2026-09-02), and breaking is
free now and expensive after `folio-go/v0.1.0`."* That is [[D-7.8.3]]'s rule — *narrowing what is
accepted is free exactly once* — **stated independently, by the owner, in the document that governs the
thing being narrowed.** Two records of one ruling that had never been cross-referenced, which is why the
whole argument was possible.

### The finding

**Three parties — the builder, the lead, and me — argued a format-version question through the decision
logs, and `folio-format.md` held the answer, the test, two precedents, a fourth argument, and the owner's
standing ruling the entire time.** Every error in this thread has the same shape: the builder cited the
file and misread it; I searched the file with a vocabulary it does not use and declared the clause absent;
the lead ran the same query and reported the sample as an enumeration; and I then asserted a precedent was
missing from a file I still had not read end to end. **Four failures, one cause — the file was treated as
something to query rather than something to read.**

**Standing rule: a question about what a document permits is answered by reading that document, and the
decision logs are the secondary source.** A log records why a choice was made; the spec records what is
true. When they are in tension the spec wins, and when the log is silent the spec is usually not. The
practical form: **before escalating a question about a governing document, read the governing document's
relevant section in full** — not a grep of it, and not the log entries that cite it. Grep finds what you
already know to name, which is exactly the wrong instrument for discovering that your question has already
been answered.

---

## D-000.21 — A gate that states its own hole honestly, and an orchestrator rule that holds it permanently open

**Found by 12.2's builder while assembling its commit-boundary report.** `npm run scan:font-hosts`
returns rc 0 and reports *"0 occurrence(s) in 625 **tracked** source files under .. (floor 400)"*. The
population comes from `git ls-files -z` at `forbidden-font-hosts.mjs:253`. **Story 12.2 creates four new
files, and my standing prohibition on subagents staging anything means all four are untracked for the
entire life of the story.** The scan is green over a tree containing none of the files the story wrote.

**This is not a dishonest guard, and that distinction matters — it is the first of the run's false-green
findings where the tool is blameless.** `forbidden-font-hosts.mjs:8-13` states the hole before anyone
hits it:

> *"It proves that no forbidden font host appears in the SCANNED POPULATION, and nothing more. It is NOT
> a proof that 'no request leaves the machine' — a source scan cannot see a host assembled at runtime,
> **a host in an untracked file**, a host in a dependency, or a request made by something this repository
> did not write. Every message below is worded to claim the bounded thing."*

The guard names untracked files explicitly, bounds its own claim, and even says its messages are worded
to keep that bound. **Every one of the nine prior instances of this run's dominant defect was a guard
that could not distinguish a correct outcome from a plausible wrong one. This one can, and says so.**
The defect would have been entirely in the **consumer** — a report saying "font-host scan clean" about a
tree missing the only files the story wrote. *A bounded claim becomes a false claim at the moment
somebody quotes it without its bound.*

**And the cause is mine, not the tool's.** The prohibition on subagents staging exists so the
orchestrator makes every commit, and it is worth keeping. But it converts a hole the script describes as
occasional — *a host in an untracked file* — into a **structural certainty for every new file in every
story**. A rule that is right on its own terms can hold a second rule's exception permanently open, and
neither rule is wrong where it is written.

**Handled correctly by the builder without being asked:** the caveat is written into the spec's
Verification section *beside the command*, and the implementer runs an explicit compensating grep over
the four new files by absolute path — for `fonts.googleapis.com`, `fonts.gstatic.com` and the
`DECLARED_ONLY_FONT_HOSTS` entries at `:41-44` — with the file list stated and a positive control that
fires, **reported as a separate measurement rather than folded into the scan's number**. Two
measurements that cover different populations must be reported as two, or the narrower one inherits the
broader one's authority.

**This is the second time `git`'s tracked-only default has produced a false zero in this run**, after
`git grep`'s. Both times the default was documented and neither time was it front of mind.

### The builder's sharpening, which is better than the rule it refines

[[D-12.C.2]] concluded *"a question about what a document permits is answered by reading that document."*
The builder corrected the diagnosis of its own error:

> *"My error wasn't failing to read `folio-format.md` — it was grepping it, finding line 68, and citing
> it **without reading the sentence to its end**. Grep is worse than useless when it half-answers,
> because a real line number makes a wrong claim look measured."*

**That is the sharper statement and it supersedes mine.** A grep that finds nothing at least announces
its own failure. **A grep that finds the right line and stops mid-sentence produces a citation with a
verifiable address and a false payload** — and the address is what makes reviewers trust it. It is the
same failure as citing a symbol without checking what it does, and it was the form all three of this
story's corrections took. **A citation must carry the boundary of the claim, not just its location.**

---

## D-15.2.1 — Widen the scan's population, and re-mark every `git grep` absence this run has leaned on

**Story 15.2b created**, `epics.md` and tracker updated. The fix to [[D-000.21]]'s hole is **(b)** — widen
`scan:font-hosts` to `git ls-files --others --exclude-standard` alongside the tracked listing — dispatched
as **its own small story alongside 12.3**, not inside 12.2 and not deferred to 15.2.

**Why not the compensating grep as the permanent answer.** *"Option (a) is not a workaround that happens
to be manual; it is a **prose caveat in procedural form.** It discharges only if whoever dispatches the
next story remembers, and this run has established what that is worth."* The interim check stays in force
**time-boxed to 15.2b landing**, so it has a named end rather than becoming permanent by habit.

**And the reason for (b) is not population size.** Today the scan reports *"clean over 625 files"* and
**cannot distinguish "no forbidden host found" from "I did not look at the four files you just wrote."*"
Widening does not merely raise the number — **it collapses all-clear into couldn't-look**, because there
is then nothing in the non-ignored tree the scan did not read. That also disposes of (c), a CI-only gate:
CI sees only what was committed, so it has the identical hole.

**Placement.** Not 12.2 — its subject is locale and UTC offset, and moving a shipped guard's population
as a side effect of an unrelated story is the scope creep this run polices. Not 15.2 — that is tail, which
means roughly twenty more stories carrying the hole, and 15.2's subject is *CI's red means something*,
which is signal integrity, a different property from a local gate's population.

### The lead's blocker did not materialize, and the check found a worse one

The lead flagged one assumption to measure first: if `src/generated/` is not gitignored, the emitted
stylesheet and the font-index snapshot — which **legitimately carries host-shaped strings** — enter the
widened population, and the story is bigger than one line. Measured:

- **`git check-ignore folio-designer/src/generated` → rc 1. The directory is NOT ignored.**
- **But every generated artifact in it is ignored individually** — `.gitignore:68-75` lists
  `src/generated/runtime/`, `offline-assets.ts`, `runtime-fonts.css`, `font-catalogue.ts`, `font-index.ts`.
- One file in that directory, `pdfjs-assets.ts`, is **tracked on purpose**.

**So the blocker is cleared and the story stays small.** The condition is discharged by measurement, and
the measurement is recorded in the story rather than the conclusion alone.

**The worse thing is the pattern that clears it.** The ignores are **per file, not per directory**, and
that is deliberate — it is what lets `pdfjs-assets.ts` stay tracked in a directory of generated output.
Under a tracked-only scan a newly emitted artifact was invisible either way, so the pattern cost nothing.
**Under the widened population, an unignored new artifact enters the scan the moment it is emitted** — and
the person who added it will get a red they cannot explain, in a guard they did not touch. 15.2b must
leave a comment beside the ignore block saying so. **Widening a population converts a dormant convention
into a live constraint**, which is the kind of consequence that belongs in the story that causes it.

### The standing rule, and it reaches backwards

**`git grep` and `git ls-files` are tracked-only, and this run has used them for CLOSED counts throughout
— mine, the lead's, and the builders'.** Combined with my prohibition on subagents staging, **every file a
story creates is invisible to every git-based query for that story's entire life.** The lead probed it:
`git grep -l` returns only the tracked file where `/usr/bin/grep -rl` returns both, and the repository
holds **6** untracked-not-ignored files right now, four of them 12.2's new sources.

**The bias is not occasional, it is structural, and it always points the same way — toward "absent" or
"clean."** That is the direction that produces false confidence rather than false alarm.

Under [[D-000.7]]:

1. **A population built with `git grep` or `git ls-files` is SAMPLED, not CLOSED**, unless untracked files
   are separately accounted for in the same measurement.
2. **An absence claim must use a tool that sees untracked files** — `/usr/bin/grep -r`, or
   `git grep --untracked` — or state the exclusion inside the claim. *Absence claims are where this bias
   changes a conclusion; a presence count biased low is merely incomplete.*
3. **Re-mark, do not re-run everything.** The ones to revisit are the exit-1 / zero-hit findings this run
   has leaned on. The lead named two of its own by construction — the `containComponent` call-site
   enumeration and the padding-caller absence in `App.tsx` — and did not pretend otherwise. Neither
   conclusion moves: `App.tsx` is tracked, and the padding finding was about a tracked file's contents.
   **But they were marked CLOSED on a tool that cannot see part of the tree, and the label was wrong even
   where the answer was right.**

**The sentence worth carrying out of this whole thread:** *a correct answer reached by a method that could
not have detected the alternative is still a sample.* It is the grep-that-named-its-own-four-functions
defect arriving through a different door — and this time the door was the version-control tool everyone
reaches for first.

---

## D-12.3.0 — 12.3's AC2 and AC3 are prohibitions, and I was about to dispatch them as finished work

**Recorded before dispatch, which is the only reason it is cheap.** I had 12.3 down as the easy next
story on the strength of [[D-000.9]]: *"12.3's alt-row and header-style cascade, which is entirely
engine-side and needs no work at all."* **That sentence is true of the rendering and false of the story.**

Read at the source rather than through the summary:

- **AC2's operative clause** — *"this story adds no rendering rule and **no second implementation of the
  cascade**"*
- **AC3's operative clause** — *"which is the cascade **Story 4.1 already defines**"*

**Neither is a satisfied feature. Both are constraints on what 12.3 may build.** Filing them as "already
satisfied, spec as preservation" inverts their function: a builder told they are discharged writes a
preservation test over the engine cascade, ticks them, and **loses the constraint at the exact moment it
binds**.

**And it binds on AC1.** AC1 is the construction work — panel, command arm, projection — and the
tempting way to show a header-style field's fallback is to **compute the effective value in TypeScript**.
That is a second implementation of the cascade, in the browser, violating AC2 and AC3 and also
[[AD-15]] and [[AD-17]]. **AC2 and AC3 are what shape AC1's implementation**, and they must be spec'd as
live constraints on the panel rather than as engine-side preservation.

**This is [[D-12.C]]'s charter shape almost exactly, one story later.** There, a constraint phrased
against the mechanism was satisfiable in letter by a fix that missed the purpose. Here, a constraint
phrased as a description of the engine reads as a finished fact, and its force as a prohibition is
invisible unless you notice it is written in the negative. **Both times the danger was a clause that
looked like a statement about the world and was actually a rule about the story.**

### The fork no acceptance criterion resolves, and it decides the projection's shape

AC1 says each field shows *"the engine's committed value"* and is *"clearable back to absent."* AC3 says
an absent field **falls back through the cascade**. So: **when the field is absent, does the control show
empty, or the resolved value the document will actually use?**

Those are different controls with different projections — **one field per property versus two** — and
under [[D-7.4.5]] a projection and its `hasOnly` mirror move in **one commit**, so choosing wrong means a
second projection change later. **Story 17.3 is the shipped precedent and it points away from AC1's
wording:** it made size and leading *show the value they use*. Cite 17.3 at the plan gate so the builder
settles this deliberately rather than re-litigating a settled question or silently contradicting it.

**Also carried from D-12.A:** a red-first test phrased *"nothing writes `headerStyle`"* is **false at this
tree** — the font-chain rename cascade rewrites `HeaderStyle.FontFamily` at two sites. Scope it to *no
command **authors** it*, and **name the rename cascade inside the test as a known writer**.

### 15.2b's red-proof must be hermetic, and sequencing is the weaker fix

The two stories cannot share a tree. After 15.2b widens the population, **an untracked file containing a
forbidden host is exactly what the scan is designed to catch — and 15.2b's red-proof plants one.** A
concurrent 12.3 gate would go red for a reason unconnected to 12.3, and an interrupted prover leaves the
file behind to poison every later scan. **A mutating prover must not share a tree with another story's
gates**; contamination of this kind arrives as a plausible intermittent, the most expensive way to find
anything.

**But the fix is the prover's construction, not the schedule:** build a throwaway repository in a temp
directory — `git init`, commit one file, leave a second untracked carrying the host — and invoke the scan
against that root. That removes the shared-tree hazard, the interruption hazard and the concurrency
question together. A fixture directory inside this repository **cannot** serve, because the one property
under test is *being untracked*, which a committed fixture cannot have.

**Sequence 15.2b first regardless**, for an independent reason: 12.3's own new files then sit inside the
scanned population, and the compensating grep **stands down for 12.3** instead of being carried through
it.

**And the symmetry is worth naming.** 15.2b's own text warns a future contributor about *"a red they
cannot explain, in a guard they never touched."* Run concurrently, that hazard arrives **immediately,
between two live stories**, rather than months later. The story now carries that observation, because a
hazard a story documents and then commits is not a hazard it understood.

### The prose-into-guard move, for the third time

15.2b's per-file ignore hazard was going to ship as a comment beside the ignore block. **Prose discharges
only if someone finds it.** `build-wasm.mjs` knows exactly what it emits and `.gitignore` states exactly
what is ignored, so the relationship is **assertable**: the emitted set must be a subset of the ignored
set, failing with a message that names the new artifact and tells the author to ignore it or justify
tracking it, with `pdfjs-assets.ts` **named in the assertion as the deliberate tracked exception** rather
than merely absent from the emission list. Now an acceptance criterion. The comment stays as the
explanation; the assertion is what closes it.

**That is the third time this run has converted a warning into a guard**, after [[D-15.0]]'s tripwire and
[[D-000.21]]'s compensating check. The pattern is stable enough to state as a rule: **if a document knows
a fact and a program can read that document, the fact belongs in an assertion and the prose belongs
beside it as the reason.**

---

## D-000.22 — An entire epic shipped without ever being written in the document builders read

**Found while pre-measuring [[D-12.3.0]]'s projection fork.** The lead cited *"Story 17.3 is the shipped
precedent"* for how a panel field shows a value the document has not set. I went to read it in
`epics.md` and found nothing — then checked with a control rather than trusting the zero, per
[[D-000.20]]:

| Query | Result |
|---|---|
| `^### Story 17\.` in `epics.md` | **0**, rc 1 |
| `^### Story ` in `epics.md` (control) | **114** |
| `Epic 17` anywhere in `epics.md` | **0** |
| `^  17-` in `sprint-status.yaml` | **6 stories, all `done`** |
| Spec files `17-*.md` in implementation-artifacts | **6, all present and full** |

**Epic 17 is six shipped stories with complete specs and no entry whatsoever in the planning document
that builders resolve epic context from.** This is [[D-15.0]]'s defect **inverted**: there, a story had a
tracker slot and no text and had not been built; here, six stories were built, reviewed, closed and
cited as precedent while the planning record never learned they existed. **The two failures share a
cause — the tracker and the epics doc are updated by different actors at different moments, and nothing
reconciles them.**

**Why it was not harmless.** Epic 17 being `done` means no dispatch depends on it, which is why it went
unnoticed. But **two of its stories are live constraints on work still ahead**:

- **17.3 decides the exact question [[D-12.3.0]] flagged as unresolved** — it replaced placeholder
  defaults with the effective value and writes it *only on commit*, under *"opening a document may never
  mutate it."* Story 12.3's AC1 (*"clearable back to absent"*) has to be reconciled against that, and a
  builder resolving epic context from `epics.md` **would not have found the precedent at all**.
- **17.4 carries the human ruling on bounds** — a floor and a ceiling are not the same kind of bound —
  which forbids browser-side clamping even where the engine refuses, and binds every later control.

**Written as a pointer, not reconstructed.** `epics.md` now carries an Epic 17 section listing the six
stories, their spec files and their status, saying plainly that the specs are the primary record and
this section is not. **Re-deriving acceptance criteria from shipped code would have manufactured a
second, weaker source** that disagrees with the first in ways nobody would notice — which is the failure
this entry exists to record, not one to repeat while fixing it.

**Record hygiene, flagged not fixed:** `17-1`, `17-2` and `17-3` carry `status: in-review` /
`in-progress` in their own frontmatter while the tracker says `done`. Same inconsistency Story 12.1's
closer reconciled downward. Left for whoever next touches each file, with the tracker named as the
authority.

**The generalisation, and it is the one this run keeps arriving at from new directions:** *a record is
checked by use, never by review.* Epic 17's absence survived six closes, an epic boundary gate and
repeated citation of its own stories, because **nobody needed to look it up until somebody needed to
look it up.** The only reason it surfaced now is that a decision required reading a precedent rather
than citing it.

---

## D-12.3.1 — 17.3 answers 12.3's fork, and the answer does not transfer unchanged

**Pre-measured before dispatch, so [[D-12.3.0]]'s unresolved fork does not cost a CHECKPOINT 1 round
trip.** The question was: when a header-style field is absent, does the control show empty or the value
the document will use? The lead named Story 17.3 as the precedent. **It is, and reading it changes the
answer.**

**What 17.3 actually did — and it is not "show the resolved value".** It put the **engine's default** on
the canvas projection beside the committed value — `defaultFontSize`, then `defaultLineSpacing` added by
that story — and had the panel render the effective value by reading the default **off the projection**
when the committed value is absent. Its Always clause is explicit that the default *"comes from the
projection at the render site — never from a literal in `App.tsx`"*, and its red-proof mutates the Go
constant to `999` to prove the panel is not re-minting it. **The browser composes; it does not compute.**

**Why that does not transfer to 12.3 unchanged, and this is the operative point.** 17.3's fallback is
**one level deep** — a field is set, or the engine's default applies — so shipping the default as a
projection member leaves the browser doing nothing but choosing between two values it was handed. **12.3's
fallback is a two-level cascade**: a header-style field falls back to *the table's own style* and then to
*that field's documented default* ([[AC3]], the cascade Story 4.1 defines). Handing the browser both
ingredients and letting it pick **is** implementing the cascade in TypeScript — the precise thing
[[D-12.3.0]] identified AC2 and AC3 as prohibiting, and a violation of AD-15 and AD-17 besides.

**So 12.3 projects the RESOLVED value, not the ingredients.** The engine already owns the cascade and
already runs it; the projection carries its answer. That satisfies 17.3's principle — *the panel shows
the value the document will use, sourced from the engine* — by the only route that does not put a second
cascade in the browser. **The precedent's principle transfers; its mechanism does not, because the
mechanism was shaped by a one-level fallback.**

**This also resolves the one-field-versus-two question [[D-12.3.0]] raised.** It is **two projection
members per property** — the committed value (so the control can tell set from unset, and so *clearable
back to absent* remains meaningful) and the resolved value (so the control can show what will actually be
used). One member cannot serve both: a single resolved value makes *absent* unrepresentable, and a single
committed value forces the browser to resolve. Under [[D-7.4.5]] the projection and its `hasOnly` mirror
move in one commit, so this is settled at the plan gate or paid for twice.

**The protocol-site checklist 12.3 inherits, recorded by symbol rather than line number** — per
[[D-12.C.4]]'s stale-citation finding, where the cited lines were moved by the comment block citing them.
17.3 found that adding one projection member touches **five** sites and that its own Code Map had omitted
one of them:

1. the Go `CanvasProjection` struct;
2. the `CanvasProjection{…}` construction in `page_setup.go`;
3. the TypeScript projection type in `engine-protocol.ts`;
4. **both** `hasOnly` *and* the typed-key validator in `engine-protocol.ts` — these are two edits at one
   site and are what the omission was;
5. the recorded sorted wire-key set in `canvas_projection_wire_test.go`.

**12.3 adds two members per property across three properties, so this checklist runs six times.** A story
whose Code Map omits site 4 will pass `tsc -b` and fail the wire-key record — which is the good outcome,
and only because that record exists.

---

## D-000.23 — `D-000.x` is numbered per log, so a bare cross-log citation names two rules

**Found by 15.2b's builder at its plan gate**, reported as a numbering collision. It is not one, and the
correct diagnosis changes the fix from "renumber" to "cite differently."

**Measured across every decision log:** `D-000.x` is a **per-log namespace**. Each run-scoped log numbers
its own series from 1, and **27 numbers are defined in more than one log, deliberately**:

| Number | `folio-mvp-decision-log.md` | `epic-8-15-decision-log.md` | `epic-11-14-decision-log.md` |
|---|---|---|---|
| D-000.1 | build order is numeric | scope: all of Epics 8–15 | the run's five standing parameters |
| D-000.4 | heavy tests at each epic boundary | heavy-test cadence: per-epic | every Code Map cites a symbol and a count |
| D-000.21 | **assert on the produced thing** | — | a gate that states its own hole honestly |

**So nothing here is misnumbered and nothing gets renumbered.** Renumbering would break the convention
and orphan every intra-log citation — the fix would cost more than the defect and introduce a worse one.

**What is genuinely defective is the bare cross-log citation.** Inside a log, `D-000.21` is unambiguous.
From a story spec in `implementation-artifacts/`, or from `epics.md`, it names **two different rules** —
and the MVP log's is *"Assert on the produced thing, never on the thing you asked for"*, one of the
most-cited rules in the project, with a sharpened successor and twenty-odd references. A reader grepping
the bare number lands on the wrong one, **and lands on a real, plausible, load-bearing rule** rather than
on nothing. **That is what makes it dangerous: the wrong answer looks exactly like the right one.**

**Rule, effective now: a citation that leaves its log names the log.** `epics.md`'s Story 15.2b entry is
corrected accordingly. Intra-log citations stay bare — the convention is fine where it was designed to
work, and the failure is entirely at the boundary.

**Not swept.** Every existing bare cross-log `D-000.x` citation is left as it stands. A global rewrite
across shipped specs and closed stories is a large edit whose only benefit is to readers who have not yet
been confused, and this run has learned that a large mechanical edit to records is how a correct record
becomes a wrong one. **Filed as a convention going forward, with a named trigger:** correct any bare
cross-log citation *in a file you are already editing for another reason*.

**The shape, one more time.** [[D-12.C.2]] recorded that a citation must carry the **boundary** of its
claim. This is the same rule at the level of the identifier: **a citation must carry enough of its
namespace to resolve.** A number that resolves in the room where it was written and misresolves
everywhere else is a citation that works exactly until someone takes it seriously enough to look it up.

---

## D-12.3.2 — Story 12.3's four forks, two corrections to my own rulings, and a guard that had become a layout requirement

**12.3 halted at step-02 rather than guessing, and the investigation was worth the stop.** Four forks
ruled below; the fifth question — whether Story 14.8 survives — is **escalated to the owner** and is
**not 12.3's blocker**, because the answer changes not one line of the story.

### Q1 — [A]. 12.3 ships command, projection AND working controls

*"This story presents Story 12.3's capability and adds no second way to **store** it"* (14.8's AC1) is
about **storage ownership, not about who first renders a control.** 12.3's own AC1 is a UI criterion in
its own Given/When/Then, FR60 says *"from the table editor"*, and the traceability table puts FR60 in
Epic 12. The storage-only reading makes 12.3's AC1 untestable — and **a story whose acceptance criterion
cannot be run is not a smaller story, it is a broken one.**

### Q2 — [A]. Exempt `headerHeight`, amend the criterion, and mark the format Required

`headerHeight` **cannot** be cleared back to absent, and AC1 said every field could be. Measured: the
loader hard-errors *"missing required field for a table"* — one of only three, with `bind` and `columns`;
the serializer writes it unconditionally, unlike its two neighbours; `folio-format.md` marks
`as`/`headerStyle`/`altRowBackground` Optional and gives this one **no optional marker and no default**;
all 31 corpus files match every table 1:1 with one. **Nothing in five decision logs, the planning
artifacts, the format spec or the deferral register addressed the collision.**

**Making the key optional is not merely a format change — it is a MAJOR one.** New documents could omit
it and an existing reader would refuse the file, failing the same pre-reader test that `justify` and the
`{"asset":…}` chain entry both failed, and both of those took the 2.0 increment. **A table-editor story
does not take a MAJOR format increment.**

So the criterion was wrong, not the product: `x`, `y`, `width`, `height` and `value` are **already**
non-clearable. AC1 now reads *clearable back to absent **where the format permits absence***.
**And `folio-format.md` marks the field Required** — today it is required **by silence**, a rule enforced
by code and documented by omission. That is not a format change; it is the document describing what
ships. Same move as 12.2's `utcOffset` row.

### Q3 — seven style fields, and a forward finding nobody had

`template.Style` has **11** presence fields; `resolveHeaderStyle` has arms for **9** — **Bold and Italic
have no arm at all.** AC1's own wording is the bound: *"the fields the engine already resolves."* Minus
`padding` (forced by [[D-12.4.1]] — projecting it would offer a value nothing can write) and `border` (a
nested block the cascade treats block-granularly, deferred **with a named trigger**) leaves **seven
scalars**: fontFamily, fontSize, lineSpacing, background, color, valign, align. **14 projection members,
plus `altRowBackground` and `headerHeight`.**

**My D-12.3.1 priced this as "two members per property across three properties." It treated `headerStyle`
as one property.** At nine fields it would have been eighteen members. The estimate was off by the size
of the story.

**The forward finding, and it is the most valuable thing the investigation produced: `resolveHeaderStyle`
has no Bold and no Italic arm.** Epic 11 is about to make bold and italic real on the render path. When
it lands, **a table's header row still will not be able to be bold**, because the cascade has nowhere to
resolve it from. Nobody had this on any list. Registered with a trigger keyed on Epic 11's resolution
story, so 11.2 meets it as a **named consequence rather than a surprise**.

The spec also carries an **11-row `Style` mapping** — every field marked in-scope, forbidden, deferred, or
*no arm exists* — which discharges `epic-7-8-decision-log.md`'s D-8.1.2 (*any story walking a `style.X`
must state whether it walks `headerStyle.X`, and why*) permanently for this area. **12.3 is that rule's
archetype and its inverse.**

### Q4 — my own ruling's mechanism was wrong, and the correction upgrades the story

[[D-12.3.1]]'s five-site checklist named the **document-level** `CanvasProjection`, inherited from 17.3.
But these three properties are **per-table**, nothing restricts a document to one table, and **the table
editor does not read that projection at all.** It reads `TableColumnsProjection`, requested by element id,
which already carries table-level non-column data. **I picked the surface before anyone measured which
projection the table editor reads.** The principle — project the resolved value, two members per property,
projection and mirror in one commit — survived intact; only the sites were wrong.

Three consequences that make this more than a surface swap:

- **`hasExactKeys` is stricter in both directions than `hasOnly`, so the red-proof gains an arm.** Prove
  RED for a projected member missing from the guard **and** for a guard key with no projected member.
  Under `hasOnly` only the first was possible.
- **The surface is pinned by nothing** — `isTableColumns` in any `.go` = **0** against a positive control
  of `isCanvas` = 12; `TableColumnsProjection` in any `_test.go` = **0** against `CanvasProjection` = 38.
  **12.3 is the first story to put load-bearing data on an unguarded surface, so adding the pin is part of
  the story**, not an adjacent improvement.
- **`engine-client.ts:#settle` hand-enumerates the table object's members**, so a new table-LEVEL member
  passes the guard and is then **silently dropped** before `App.tsx` sees it. Columns survive only because
  they ride a spread. This is the most dangerous thing in the section: it fails **quietly**, where a
  `hasOnly` mismatch at least blanks the canvas loudly. All nine properties ride that path.

### Two corrections to my rulings

**[[D-12.3.0]] said the rename cascade writes `HeaderStyle.FontFamily` at two sites. It is one.** The
only non-test writer is `renameFontChain`; the second reference is `fontChainReferences`, a **read** used
by `deleteFontChain`. Checked at the ruling's own tree with `git show` — identical there, so **it was an
over-count when written, not drift**, and that distinction matters because drift would have implied
something changed. The test names it singular; **two writers would be false in the other direction, which
is how a correction over-corrects.**

**AC3's "documented default" walks into a live format-vs-code disagreement.** `folio-format.md` documents
`fontSize` **10**; `defaultFontSizePt` is **12000**, `resolveHeaderStyle` seeds from it, and
`CanvasProjection.DefaultFontSize` **already ships 12 to the browser today**. Filed, not fixed — and the
entry must say why, because it looks inconsistent with [[D-12.C]]. **There, the code was loose and the fix
moved no rendered bytes. Here, `defaultFontSizePt` seeds the render path, so "fixing" it to 10 moves every
document that does not declare a size — an AD-21 event relocating the whole golden corpus.** The entry
names both sources and **states which one ships (12)**, or a future reader finds "the format says 10" and
treats the code as the bug.

### A guard that had quietly become a layout requirement

The builder found that `trapDialog` builds its focus list in document order, so placing the new controls
**after** the matrix — where the design draws them — breaks `App.test.tsx`'s Tab-wrap assertion, while
placing them **before** keeps it green. It proposed placing them before.

**Overruled, and the lead's reasoning is the general form: choosing a layout to keep a test green is
letting a test dictate the product.** Place the controls where the design says and **amend the assertion
deliberately, stating the new focus order as the decision.** That it breaks *silently* is an argument for
making the break visible and deciding it, not for routing around it. **A guard that has quietly become a
requirement has stopped being a guard.**

Also re-keyed: **DW-199's trigger moves from "Epic 14's first designer story" to its PURPOSE** — the first
story that adds a designer UI section. 12.3 dispatches first, so the epic-numbered trigger was already
stale; a trigger keyed on a number goes stale every time dispatch order moves.

---

## D-14.8.1 — OWNER DECISION: 14.8 becomes a restyle, and the mockup is corrected rather than built (2026-09-05)

**Owner decision**, taken at the terminal with the measurement in hand. Story 14.8 was written to add
HEADER, CELLS and BORDERS sections to the table editor, following `TableEditor.dc.html`. **Four things
happened after it was written, and together they emptied it.**

| The mockup draws | Status |
|---|---|
| `Padding` | Ruled out of the panel by [[D-12.4.1]] |
| `Row height` | A fact the engine **derives** — a control would restate it |
| `Repeat on continuation pages` | Same |
| `Show header row` | **No field exists in the format.** 14.8's own AC4 forbids inventing one |
| borders preset | **No field exists in the format.** Same |

And the three things Story 12.3 is making authorable — header height, alternating-row colour, header
text style — **are not drawn in that mockup at all** (measured: alt-row/zebra 0, "Header height" 0,
colour control 0, against positive controls of 108 hex literals and "Row height" 1 on the same file).
**14.8's drawing and 14.8's available substance had almost no overlap.**

**Ruled: option (1) — restyle, and correct the drawing.** 14.8's remaining substance is presentation:
take the controls 12.3 ships in the existing editor idiom and group them into the drawn sections so the
editor reads as one designed thing, **which is Epic 14's actual subject**. `Show header row` and the
borders preset are **not coming**, and the mockup is edited to stop promising them. No format change, no
version increment. `epics.md`'s 14.8 entry carries the decision and its acceptance criteria are to be
rewritten against it before dispatch.

**The principle the owner accepted, and it is worth keeping:** *a mockup that draws capabilities the
product does not have is a drawing to correct, not a specification to implement.* The cost, stated
plainly rather than buried: **two drawn controls are being withdrawn**, and the owner decided those
promises were aspirational. Option (3) — growing the format to back them — was available and was
declined; it would have taken a version increment and new render behaviour **near the release tag**, and
would have stopped Epic 14 being a presentation epic.

**Why this reached the owner at all.** The lead ruled 12.3's half and escalated only this, correctly:
every way of answering it is defensible engineering, nothing in the architecture or the epics decides
it, and it is a question about **what the product should be able to do**. It was also **not 12.3's
blocker** — the answer changes not one line of that story — so it was framed and put to the owner
without stopping the run. **That separation is the thing to repeat: escalate the part that is genuinely
the owner's, and keep shipping the part that is not.**

---

## D-000.24 — A gate that fires on every instance is a toll, not a gate

**12.3's builder declined to argue past the token gate and said so plainly:** *"the 1600 threshold is
unreachable in this project and I will not claim otherwise."*

Measured across this project's specs: 12.1 ~10,849 · 12.2 ~17,528 · 12.4 ~9,434 · 15.2b ~14,605 · 17.3
~6,033 · **12.3 ~11,677**. The threshold is **1,600**. **Not one spec in the project has ever been under
it, and none plausibly could be** — the bulk of every spec is the Code Map, which is the investigation
the spec exists to preserve so the implementer never re-derives it.

**So the gate discriminates nothing.** It halts every story, the answer is `[K]` every time, and the only
thing it produces is a round trip and a paragraph of justification. **This run has already condemned
exactly this shape once** — [[D-8.4d.2]] rejected an interim size ratchet because *"a gate that gets
routinely bumped is not a gate; it is a chore that teaches people to bump gates."* A gate that is
routinely waived teaches the same reflex, and worse: it trains a builder to treat a halt as a formality,
which is precisely the instinct that must survive for the halts that matter.

**Handled at the orchestration level, not by each story.** The threshold belongs in
`_bmad/custom/bmad-build.toml` set to a figure this project's specs could actually cross — or the gate
should be dropped and the split judgement left where it already lives, in the multi-goal check at
step-01, which **is** discriminating (12.3's builder judged single-goal and said why). **Fixing it in
config is a one-line change; leaving it costs a round trip per story for the rest of the run.**

**The generalisation:** a threshold's value is in the cases it separates. **One that admits everything
and one that rejects everything are the same instrument, and neither is a measurement.** Before shipping
a gate, measure the population it will judge — a floor nothing reaches and a ceiling everything exceeds
are both decoration.

### D-000.24.1 — the remedy, corrected before it was applied

D-000.24 says the threshold *"belongs in `_bmad/custom/bmad-build.toml`"*. **I went to make that change
and did not.** No such file exists; `_bmad/custom/` holds only `config.toml` and a gitignored
`config.user.toml`, both empty of anything but comments. **And I do not know the key name.**

**Guessing it would produce a config file that silently does nothing** — a file that looks like the
problem is handled, changes no behaviour, and reports no error. That is this run's dominant defect class
in a new location: **a remedy that cannot be distinguished from a working one.** Worse than the toll it
was fixing, because the toll at least announces itself every story.

**So the remedy is the one I fully control: pre-answer the gate in the dispatch.** Every dispatch from
here says the token gate is answered `[K]` in advance, with the population measurement, and instructs the
builder not to halt on it. Zero risk, no guessed key, and the reasoning travels to the agent that would
otherwise have to re-derive it. If the real key surfaces later — from BMAD's own docs or a workflow
source, **read, not inferred** — the config change becomes the better fix and this note is its trigger.

**The general form, and it is the sharper half of D-000.24:** *when the obvious remedy requires a fact
you do not have, the remedy that requires no facts is better than a plausible guess at the fact.*

---

## D-000.25 — A refusal asserted only by its DataPath is not asserted

**Story 12.3's highest-value finding, produced by mutation at step-03 before any reviewer ran.** The
builder mutated six guards; five reddened and one did not.

**`setTableHeaderHeight`'s "headerHeight cannot be cleared" refusal could be deleted outright with the
entire suite still green.** With the check gone the command fell through to *"headerHeight must be a
positive length"* — **a different refusal, for a different reason, located at the same
`table.headerHeight`** — and the shared test helper compared only DataPaths. Both refusals landed on the
same path, so both satisfied the assertion. The test proved the command refused *something about that
field*, never that it refused **this** thing for **this** reason.

The fix is `refusalSaysWhy`, and it was verified the only way that counts: **the identical mutation, run
again. Green before, red after.**

**Why this is worse than an ordinary weak assertion.** A refusal has two payloads — *that* it refused,
and *why*. The DataPath carries the first. Asserting only the path means **every refusal at a given field
is interchangeable with every other**, so a story can delete a rule and inherit its neighbour's error
message as camouflage. **The wrong behaviour and the right one produce identical evidence** — this run's
dominant defect, now found in the shape of the assertion rather than the shape of the guard.

**And it generalises past refusals.** Any assertion keyed on *where* something happened rather than
*what* happened admits every sibling outcome at that location. The rule: **an assertion about an error
must pin the reason, not only the address.** A located message has two halves and a test that reads one
of them is testing half a guard.

**The method deserves as much credit as the finding.** Nothing in three review layers found this — the
builder found it by mutating its own guards before submitting them, and the one mutation that stayed
green is the only one that mattered. **Five green mutations were the cost of finding the sixth**, which
is the correct price and worth paying every time.

### Two more from the same story, kept because each is a class

**A browser guard stricter than the file door kills the worker.** A hand-authored **negative**
`headerHeight` loads fine — `decodePoints` negates on `sign < 0` and nothing bounds it — and 12.3's new
`>= 0` browser guard would have **silently terminated the worker on first table-editor open**. The guard
now admits exactly what the file door admits. **A browser bound tighter than the loader's is not defence
in depth; it is a crash reachable by a legal document**, and it echoes Story 17.4's ruling that a floor
and a ceiling are not the same kind of bound.

**A test mock that returns one frozen value for every request hides state bugs by construction.** The
colour text box was uncontrolled beside a controlled swatch, so a picked colour was reverted by the next
blur — invisible because the mock answered every projection request with the same object. **A mock that
cannot represent change cannot fail a test about change.**

### One refused patch, recorded because the refusal was right

I ordered a synchronous in-flight ref for the colour picker, on the premise that a drag would emit
several commands. **Instrumented: the handler is entered three times and exactly one command goes out** —
React flushes discrete events before the next. The builder refused the patch and kept a non-vacuous
property test instead. **My premise was wrong, and evidence beat compliance.** A builder that implements
a wrong instruction correctly has produced a defect with an alibi.

---

## D-000.26 — A story sat in `review` for four stories, and the census that named it was reported to me twice

**Orchestration defect, mine.** Story 12.4 shipped in `fd4da07`. I never dispatched a closer for it: I
went from 12.1's close straight into 12.2's dispatch and it fell in the gap. Measured today — spec
frontmatter says `status: 'done'`, the tracker says `review`, and the story has **no `## Delivery Log` at
all** (0 occurrences against a control of 9 `##` headings). Four stories shipped over the top of it.

**The part worth recording is not that I missed it. It is that I was told, twice, in a format I had asked
for.** 12.2's closer reported a tracker value census ending *"…3 in-progress, **2 review**, 1
moved-to-epic-15, 1 deferred-to-epic-15"* — the two being 12.2 itself and 12.4. After 12.2 closed, the
count went to **1 review**, and 12.4 was the one. **I read both numbers as totals and neither as a claim
about a specific story.**

**An aggregate is a claim about its members, and a total that changes tells you which member moved.** A
census exists precisely so that an unexpected count is noticeable — and *"1 review"* while no story was in
review was exactly that signal. I had asked for the census, received it, and used it only to confirm that
nothing had been mangled. **A number checked for damage is not a number read for meaning.**

**Why the usual defences did not catch it.** The tracker was never wrong — `12-4: review` was accurate the
whole time. No gate failed, because a story that is built and committed breaks nothing by lacking a
Delivery Log. **The record was internally consistent and incomplete**, which is the state no assertion in
this run can detect: every guard here checks that a claim matches reality, and none checks that a claim
exists. **This is the absence-needs-a-tripwire rule ([[D-15.0]]) turned on the orchestration layer, where
I had not applied it.**

**Standing correction, effective now: before dispatching any story, read the tracker's `review` set by
NAME, not by count.** A story in `review` is a story whose code shipped and whose record did not, and it
is invisible to every gate in the system. The check is one grep and it belongs in the same breath as
confirming the tree is clean.

**And a note for the close itself:** 12.4's closer is running under an explicit instruction that gates run
today measure **today's** tree, not 12.4's, because 12.2, 15.2b and 12.3 have shipped over it. A late
close is the one moment someone reads an old record against a new tree — **the temptation is to write a
Delivery Log that implies a verification nobody performed**, and that would be a worse defect than the
missing log it replaces. It has been told to attribute every figure to its era and to report, not repair,
any drift it finds in 12.4's record. **D-12.4.1 is cited by both 12.2 and 12.3, so what that story
actually shipped needs to be legible from its own file.**

---

## D-000.27 — The guard census, and the one measured case where a line break made two guards both say PASS

**Filed as `source-text-guard-inventory.md`.** An exhaustive census of every test that reads a source file
and matches its **text**, classified by the two properties that decide whether such a guard can be
trusted: does it strip comments, and does it break when a line wraps.

**The finding that justifies the whole exercise is already written in this repository**, in
`matrixdocs_source_test.go`: two guards shared one regex `(?m)^\s*slug:\s*"([^"]+)"`, and *"a seventh
entry written as a single-line composite literal … is gofmt-clean, compiles under `-tags matrix`, and
**BOTH guards report PASS**."* **That is this run's dominant defect produced by formatting alone** — no
bug, no bad edit, just a legal way of writing the same data that the parser could not see. And the fix
was the right one: **remove the parsing assumption by moving to AST, rather than widen the regex.**

### Comment stripping is inconsistent, and not all of it is deliberate

Three files carry **independent copies of the same quote-aware comment scanner**, duplicated on purpose so
no suite depends on another's helper. Eight more strip via a shared `blankComments`. Every AST guard in
`folio-go` and `lint` is comment-blind by construction — **9 of lint's 11 scanners**, and the larger half
of folio-go's.

But a long list **does not strip**, including **all 24 tests in `engine-bounds-mirror.test.ts`** — the
Go↔TS mirror that pins eight numeric bounds and four list identities across the language boundary — where
a mention in a comment counts as a live copy.

**Three places record what not stripping already cost, in the repo's own words:**
- `font-binary-identity.test.ts`: with three rules commented out, *"every test in this file stayed green
  while the emitted stylesheet dropped to three rules, which is the exact state that shipped the reported
  Thai overlap."*
- `canvas-font-stack.test.ts`: a comment spelling `new FontFace` made a raw scan report the seam as
  registering **after the registration had been removed.**
- `file-access-contract.test.ts`: an earlier whole-file skip also dropped `localStorage` — the very
  prohibition D-16.2 names at that module.

**And one unreconciled disagreement worth a ruling: `font-catalogue.test.ts` does NOT strip while
`font-binary-identity.test.ts` DOES, over the same generator file.** Two guards, one subject, opposite
policies, and the stripping one is the one that measured a real defect.

**Two scanners read raw text deliberately and are tested for it** — `forbidden-font-hosts.mjs` and
`host-font-access.mjs` treat a host or an API call in a comment as a finding, and blank comments *only*
for the declaration exemption, **so nobody can buy an exemption by writing a comment.** That is the model:
not "strip" or "don't", but **choose, state the reason, and assert the behaviour in both directions.**

### Line-break fragility is concentrated — and, to the repo's credit, almost always fails loudly

The heaviest concentration is `engine-bounds-mirror.test.ts`, whose every extraction is `^…$` against an
exact one-line gofmt or prettier spelling. **But in nearly every fragile case a broken extraction produces
a RED** — a non-vacuity floor, a `t.Fatal`, or a `throw`. The `*RedProof` tests that locate an injection
point by exact literal all fail with *"this red-proof's injection point is stale."*
`offline-release-contract.mjs` treats 0 **or 2** matches as a throw, never first-match-wins, and proves it
on both a commented-out and an indented occurrence.

**That is the difference between fragile and dangerous.** A fragile guard that fails loudly costs a
maintainer ten minutes. A fragile guard that fails *open* is the matrix-slug case: two green checks over a
tree nobody had checked. **The property to preserve is not robustness — it is that the failure mode of
not-understanding is red.**

Only **two** places neutralise wraps before matching, and both say why. `statement_golden_fixture_test.go`
puts it best: *"a guard that missed it would be checking the paragraph's line breaks rather than what it
says."*

### Standing rules

1. **Prefer AST over regex for any new source guard.** The repo already knows this and wrote down why.
2. **Declare the comment policy and prove it**, in whichever direction — both directions asserted.
3. **A wrap-fragile extraction must fail loudly**, never first-match-wins.
4. **Before editing a file, check the inventory.** `App.tsx`, `App.css`, `engine-protocol.ts`,
   `build-wasm.mjs` and `component_commands.go` are each read by several guards that do not strip
   comments and are line-anchored.

---

## D-12.5.1 — A ruling whose verdict was right and whose ground was too wide, corrected by the story it would have blocked

**Story 12.5's plan gate, and the most consequential correction of the run** — because unlike the others,
this one would have prevented correct work rather than merely misrecording why correct work was correct.

### The correction

I dispatched 12.5 asking whether its AC5 drag limit was the browser-side clamping Story 17.4 forbids. The
builder read 17.4 in full rather than my summary and found three things:

1. **17.4's ruling is about the INSPECTOR** — its closing clause says so, and the property it asserts is
   *consistency with typing*. **A canvas boundary has no typing to be consistent with.**
2. **A browser-side LAYOUT clamp on a pointer gesture already ships**, and I verified it:
   `resize-anchor.ts` computes `limitHeight` from the band for the capping bands and clamps both `y` and
   `bottom` against it. Its comment states AC5's rationale almost verbatim — *"a pointer that leaves the
   band should leave the component against that edge … not hand Go a rectangle it can only reject — a
   rejected drag lands the component back where it started, which reads as the drag having been thrown
   away."* **17.4 itself preserved this split**: its arrow keys send and are refused, while the drag
   clamps.
3. **The quiet-second-copy objection has a named answer.** DW-36 permitted that clamp on the condition
   that it *read* the engine's declaration rather than re-spell it, and `engine-bounds-mirror.test.ts`
   declares `dragClampPath` at that file. **A copy inside the mirror census is not a quiet one.**

**So [[D-12.B]]'s sentence — "the objection is the browser computing layout geometry at all" — was wrong.**
It flattened 17.4's two load-bearing qualifiers, *"in the inspector"* and *"quietly-drifting"*. The lead
owned the error rather than assigning it. **D-12.B's verdict survives untouched** — a typed band-height
panel field is still given no floor — **but its ground did not**, and a ruling whose conclusion is right
and whose reason is too wide is the most durable kind of wrong record: nothing ever fails to make you
re-examine it.

### The rule that replaces it, which is sharper than the one it corrects

**Clamp a gesture at bounds that carry no information; send, and let the engine refuse, where the refusal
names something the author needs.**

That subsumes gesture-versus-field and explains why AC5's three bounds rule differently:

| Bound | Kind | Ruling |
|---|---|---|
| Floor at 0 | property of the FIELD | **Clamp.** 17.4 authorises it directly. |
| Ceiling at no content window | property of the LAYOUT | **Clamp**, via one new mirrored pair tying Go's `− 1` margin to the TS side with a one-sided-edit red proof. |
| Floor at the lowest occupied edge (the strand) | property of the LAYOUT | **Do NOT build.** 12.1's Q4 already ruled it deliberately unbuilt. |

**The strand floor is the one that makes the rule legible.** Its refusal is the only one carrying an
**ElementID** — it names *which component is in the way*. A clamp would silently say "no further" and
destroy exactly the information the author needs. **Two bounds of the same kind, in the same criterion,
decided oppositely, and the discriminator is what the refusal would have told you.**

Retiring 12.1's AC4 is approved **as a narrowing, not a deletion**: scope the comment to the panel and
**state why the panel keeps no bound while the gesture gets one**, or the next reader finds a bare
asymmetry and "fixes" it in whichever direction they prefer. That sentence is the whole value of the edit.

### Two more approvals, and a sweep that found its own boundary

**AC7 is discharged by 12.1's refusal.** The engine refuses rather than clipping — probe-measured, four
shipped tests pin it — and FR44 has no vertical axis: 17 diagnostic codes, two clip codes, and
`render.go` stating that D-2.8.1 *"fences the vertical axis out entirely."*

**The lead asked for a sweep because this was the second story ambushed by one false premise. Measured:
exactly two instances in `epics.md`, both vertical, 12.1's and 12.5's, and no third in Epics 13–14.** Both
struck with the measurement beside them. **The useful half is what the sweep did NOT strike:** Story 7.3's
clip-and-warn is about a justified line exceeding its declared **width**, which is exactly the axis where
`TEXT_CLIPPED_WIDTH` exists. **A sweep that struck every mention of FR44 clipping would have destroyed a
true criterion alongside the two false ones** — the premise was wrong only where it was applied vertically.

**AC4's snap gets a fifth command field.** The alternative puts the **first** coordinate rounding in the
browser, which is the actual AD-17 breach, and the precedent costs far more than an arity change on an
internal, pre-tag command. Two conditions: the arity gate and all 22 payload literals move in **one**
commit, and *"byte-preserved"* must say **document** bytes — the payload itself gains `"snap":false`, and
an ambiguous claim there gets filed as a defect against a story that did exactly what it said.

### Three warnings I gave that did not fire, all three corrected by measurement

- **`trapDialog` focus order** — both copies scope to `dialog.current`; no existing test breaks.
- **`property-prose-height.test.ts` "asserts the drag affordance is unique"** — it does not. Its three
  uniqueness assertions are keyed to **class names**, and a new handle is invisible to all of them. **I
  generalised "a guard about the resize handle" into "a guard about resize handles" without reading it**
  — a real guard, correctly cited, described one scope too wide. The same move as [[D-12.C.1]] and
  [[D-12.C.2]].
- **"A drag handler is exactly the code that reaches for banned measurement APIs"** — a *delta* gesture
  needs `clientY` only, and the ban covers `client(Width|Height|Left|Top)`. Both shipped drags already
  work this way; `moveProseResize`'s comment says it outright. **The `offsetX/offsetY` seam exists for
  absolute placement, not delta**, so this story needs no seam at all.

**Three warnings, three misses, and every one found by checking rather than by routing around.** That is
the behaviour asked for after the opposite instinct was overruled in 12.3, and it has now paid twice in
two stories. **The general lesson is about my own warnings: a hazard asserted from memory of a guard is a
citation without its boundary**, and it costs a builder real time to disprove — but a builder who builds
around an unverified warning pays forever.

### Standing expectation

**A story citing a mockup states what it found there, including absence, with the positive control.**
12.5 measured `ns-resize` and `grip` at **zero** across every mockup, and no cursor vocabulary in
`DESIGN.md` or `EXPERIENCE.md` (rc 1 against a positive control of 47) — for a capability
`EXPERIENCE.md` says exists. With [[D-14.8.1]], **the design record is now unreliable in both
directions**: over-promising in one place, silent in another. Neither is discoverable without a stated
measurement.

---

## D-000.28 — I ran the matrix gate with a `-run` filter and got a green from two tests that assert nothing

**Epic 12's boundary gate, first attempt, and the defect is mine.** I invoked
`go test -tags=matrix -count=1 -run 'TestTargetRenderHash|TestTargetProbeHex' -v .` and got **`PASS`,
`ok`, rc 0.** Both named tests ran. Both passed. **Neither asserted anything.**

The tests say so themselves, at length, in the output I received:

> *"FOLIO_MATRIX_TARGET not set: this test asserts NOTHING and is a deliberate no-op — it is CI's
> single-target entry point (AC1, AC13), not one of the four counted legs AC11 governs. …
> **`TestCrossTargetByteIdentity` exercises all four targets from one process and is what the D-000.4
> local gate runs.** CI itself never reaches this no-op path, because every render-* job in matrix.yml
> sets `FOLIO_MATRIX_TARGET` explicitly."*

**I chose the filter by reading `matrix.yml` and copying the test names CI invokes.** That was the wrong
source: CI's per-leg jobs each set `FOLIO_MATRIX_TARGET` and are *entry points for one target*, whereas
the local gate is a different pair of functions that drive all four targets in one process. **The names
in the workflow are the right names for the workflow and the wrong names for a laptop**, and nothing but
the test's own runtime message says so.

**This is the `-run`-filter false zero — the third mechanism in this run's own catalogue — executed by
the person who catalogued it.** A `-run` filter that matches nothing exits 0; a `-run` filter that
matches only no-ops exits 0 more convincingly, because two tests genuinely ran and genuinely passed.

**Three things make it worse than a slip, and all three are the interesting part:**

1. **The guard was maximally cooperative and I still got it wrong.** It did not fail silently — it
   printed a paragraph naming itself a no-op, naming the environment variable, naming the correct test,
   and naming why CI never hits the path. **The most honest guard in the repository, and a filtered
   invocation still produced a green I could have pasted into a boundary gate.** A message only helps if
   the reader is looking for a reason to doubt.
2. **It would have gone into a gate artifact as a matrix pass**, on an epic where — as 12.4's close
   established — *no run of the matrix or Playwright has ever covered anything*. The first-ever coverage
   of five stories would have been a vacuous one.
3. **The correct invocation is the unfiltered one.** The whole tagged package is the gate; narrowing it
   is what created the hole. **A filter is a claim that you know which tests matter, and at a boundary
   gate that claim is exactly the thing you do not yet have.**

**Standing rule: a boundary gate runs its suite unfiltered.** No `-run`, no `-short`, no narrowing of any
kind — the gate's purpose is to discover what a story-level run could not, and every filter is a
pre-judgement of the answer. If a suite is too slow to run whole, that is a fact to record in the gate,
not a reason to select from it.

Re-run unfiltered. **The corrected result, and not the filtered one, is what Epic 12's gate will record —
along with this entry, because a gate that hides its own false start is worth less than one that shows
it.**

---

## D-000.29 — A second epic was sitting closed-but-open, and the tracker said so the whole time

**Found while deriving the next target from the tracker rather than from my own program list** — which is
the practice [[D-000.26]] imposed after Story 12.4 sat in `review` for four stories. It caught a second
instance immediately, one level up.

**`epic-17` was `in-progress` with all six of its stories `done`.** Epic 17 is the epic that
[[D-000.22]] already found had shipped without ever being written in `epics.md`; its tracker state was
the other half of the same neglect, and neither half was visible to any gate.

**Closed, and the evidence backing the close is stated because it is weaker than every other epic's.**
Measured: **zero** boundary-gate deferrals across all six specs, against a positive control of **6** in
12.5's — so no heavy suite is owed and no gate artifact is missing in the sense that matters. But Epic 17
**never had a gate at all**, where ten other epics do. What covers its code is incidental: Epic 12's gate
ran the unfiltered matrix and the full Playwright suite at a HEAD that contains every Epic 17 commit.
**That is real coverage and it is not the same thing as having been gated**, and the tracker comment says
so rather than letting a later reader assume parity.

**The pattern, now twice: a `done`-in-substance item stays open in the tracker because nothing ever asks
it to close.** A story lands, its code ships, gates pass — and the one remaining transition is a
bookkeeping act with no owner and no failing check behind it. **Every guard in this project verifies that
a claim matches reality; none verifies that a claim was made.**

**The standing check is now two lines, not one.** Before dispatching any story: read the tracker's
`review` set **by name**, and read the epic-level states **by name** for any epic whose stories are all
`done`. Both are one grep, both belong beside confirming the tree is clean, and both have now caught
something the first time they were run.

---

### D-11.1.1 — Procurement is not a new capability; it is the path this repo has always used

**Context.** Story 11.1 needs upstream font sources for the new cuts. 11.1's builder framed this as a
Block-If: *"nothing in the repo fetches a font binary"* — reading the offline-first constraint as
forbidding acquisition.

**Ruling (orchestrator, accepted by the builder).** That was right about *automated* fetching and the
wrong frame for the decision. `.font-sources/` is gitignored (`.gitignore:123`) and
`tools/fontgen/instance_faces.py` states the convention itself: the sources *"are NOT committed (20 MB
of inputs for 11 MB of outputs); each entry records the release URL and the sha256 to fetch them by."*
Every `UPSTREAM` entry already carries a `src_url` and a `src_sha256`, and the derivation hashes each
source before instancing and requires it to match. The three variable fonts already on this machine got
there by exactly this path.

So the offline constraint governs the **product**, not the **workshop**. The shipped artifact stays
fully offline and no build step gains a network call. We are not introducing downloading into an
offline-first repo; we are using the existing, documented, hash-pinned procurement path for releases
the repo already names.

**Why it matters beyond 11.1.** This is the second time in the run that a constraint stated about the
product has been read as a constraint on the process. Record the distinction: *a rule about what the
artifact may do at runtime is not a rule about how its inputs were obtained.*

### D-11.1.2 — Two digests, two purposes; and I read the wrong row

**What happened.** I was about to verify `Roboto_v3.016.zip` against
`e688a215e0841b6e4edb1207c93f88f4c609f82e870884349a7257e449eb9355`, having taken that value from
`folio-go/fonts/roboto/NOTICE.md`. 11.1's builder stopped me: that is the sha256 of the **extracted
`Roboto-Regular.ttf`**, which the NOTICE labels *"sha256 of the SOURCE (upstream) file"* — the file
*inside* the archive. That NOTICE records **no archive digest at all**.

The archive digest is real and lives in a **different NOTICE**:
`folio-designer/public/fonts/roboto/NOTICE.md:50` records
`1653dbe12f248da8fb0b9920db7b9496cd677ed3981154f6f15285c8bd4e334f` (29,162,959 bytes).

**Verified.** The downloaded archive hashes to `1653dbe1…` — match. Independently, its extracted
`android/static/Roboto-Regular.ttf` hashes to `e688a215…`, byte-identical to the committed shipped
file. Two digests, two purposes: one proves the archive, one proves it is the release this repo already
shipped from.

**The error class, which is one this run has seen before.** I queried for a fact using the vocabulary I
expected it to be written in, found a value of the right *shape* in the first file I looked at, and
stopped. Same shape as D-12.C.1 (compatibility vocabulary against a versioning clause). **The lesson is
narrower than "check twice": a digest is only meaningful with its subject attached.** A hex string is
not self-describing, and two hex strings in two files can both be correct and neither be the one you
need. Quote the label, not just the value.

### D-11.1.3 — The Noto Sans archive can only be verified transitively, and the spec must say so

There is **no recorded archive digest for `NotoSans-v2.015.zip` anywhere in the repo** — neither NOTICE
carries a release-archive row for it. The downloaded archive is 117,491,253 bytes, sha256
`0c34df072a3fa7efbb7cbf34950e1f971a4447cffe365d3a359e2d4089b958f5`; neither figure was recorded
anywhere beforehand.

**Verified transitively instead:** the extracted
`NotoSans/googlefonts/variable-ttf/NotoSans[wdth,wght].ttf` hashes to `bfb7bb69…`, exactly the
`src_sha256` already pinned in `UPSTREAM`. Same file, therefore the right archive.

**Ruling.** That is a genuinely weaker guarantee than Roboto's, and the spec states it in those terms
rather than implying the two were verified alike. **Record `0c34df07…` as an archive row in the new
NOTICEs while we have the file** — the whole cost of never repeating this is one line, and the next
story inherits the stronger check.

### D-11.1.4 — Both determinations came back YES: seven faces, zero new upstream releases

Two open determinations were resolved by opening the archives rather than reasoning about them.

**The Noto italic VF is already in the archive the repo pins.**
`NotoSans/googlefonts/variable-ttf/NotoSans-Italic[wdth,wght].ttf`, 2,322,640 B, sha256
`58e6e0eb…` — the sibling path the builder predicted from the notofonts release layout. Placed at
`.font-sources/NotoSans-Italic-VF.ttf`, gitignored. The italic half therefore needs **no new upstream
release**: it is one new `src` filename against a `src_url` that differs only after the `->`.

**All three Roboto cuts ship as statics.** `android/static/Roboto-{Bold,Italic,BoldItalic}.ttf` are all
present. So they are **static-upstream**: copied unmodified, no instancing, no `UPSTREAM` entry, no
fontgen involvement — Roboto-Regular's existing route. The accounting test's SOURCE-digest ==
SHIPPED-digest identity is satisfied by construction, because it is literally the same bytes.
`make fonts` / `make fonts-verify` never touch Roboto.

**Consequence.** The owner's [C] ruling — *"fetch all five, ship everything possible"* — was priced at
five upstream releases. It costs **zero**. Every source the seven faces need is now on disk and every
one traces to an archive the repo already names; exactly one new source *file* was added.

| Face | Route | Source |
|---|---|---|
| Noto Sans Bold | derived | `NotoSans-VF.ttf` (on disk, pinned) |
| Noto Sans Italic | derived | `NotoSans-Italic-VF.ttf` (new file, same archive) |
| Noto Sans Bold Italic | derived | `NotoSans-Italic-VF.ttf` |
| Noto Sans Thai Bold | derived | `NotoSansThai-VF.ttf` (on disk, pinned) |
| Roboto Bold / Italic / Bold Italic | static | `Roboto_v3.016.zip` (verified archive) |

**Two absences are rulings, not oversights, and the spec states both with their reasons attached.**
Noto Sans Thai publishes nine styles and **zero** italic variants (verified independently), so there is
no Thai italic to ship. Noto Sans SC gets no cut: its Regular alone is 10.6 MB and a CJK bold would
dominate the payload. A reader who counts families will otherwise read either gap as a bug.

### D-11.1.5 — Each cut is its own family name, and that forces the uncatalogued route

**The ruling.** The browser declares one static Regular per family with no `font-weight` and no
`font-style`, and that convention is machine-asserted. The canvas must paint bold under AD-17, so a bold
cut needs a browser declaration — but a `font-weight: 700` rule under the same family would break the
convention and put a weight axis into CSS that the format deliberately excludes. **So each cut carries
its own family name**, making `"Noto Sans Bold"` one string on three surfaces at once: the Go `FontSet`
key, the `@font-face` family, and the family the canvas paints with. The key is READABLE; **nothing may
PARSE it.** This also dissolves the "no multi-cut precedent" objection — a cut with its own family name
is not a second cut of an existing family.

**Measurement confirmed the route is forced, and it is the clean one.** `font-catalogue.test.ts`
iterates the catalogue's 31 entries and asserts per face `subfamily === 'Regular'`,
`usWeightClass === 400`, `macStyle === 0`, `italicAngle === 0`. **A bold cut fails all four, so it
cannot be a catalogue entry.** The precedent already exists: only Roboto is catalogued; the three Notos
reach the browser uncatalogued, through hardcoded `assets` slots plus hand-written `@font-face` rules in
`shippedRules`. So each cut goes to `public/fonts/<cut-dir>/` with its own LICENSE + NOTICE, a hardcoded
slot, a hand-written rule under its own family, and a `shippedFamilies` entry — **and no catalogue
entry**. Two build-time throws move together from 6 to 6+N. `font-catalogue.test.ts` needs no edit, and
the catalogue genuinely stays Regular-only.

**Accepted asymmetry:** Roboto Regular is catalogued while Roboto Bold will not be. Forced by the
Regular-only assertion, and it is the safe direction.

### D-11.1.6 — AD-26 measured: neither branch applied, and the guard I asked about was the wrong guard

I asked whether `lint`'s `fonts-asset-unaccounted` **enumerates the tree** or **reads the catalogue**,
treating those as the two branches. Measurement says the question had no answer as posed.

`lint/internal/rules/fontsassets.go` pins `const fontsAssetLocation = "folio-go/fonts"` and walks only
that. Its expected set comes from the `//go:embed` directives parsed out of `fonts.go`, not from any
catalogue. **For a designer-side directory it is simply not the guard.**

The guard that does cover the designer side is **`manifest.ResolveAssets`, and it enumerates the whole
repo.** So an uncatalogued designer-side directory is **visible and must be accounted**: it demands
`LICENSE*` + `NOTICE*` + a Copyright line, classifies the licence against the four-id font allowlist,
and adds a `MANIFEST.md` row that `TestManifestUpToDate` byte-compares — with
`licencecensus_test.go:pinnedCensus` pinning every tracked LICENSE path on top. **No AD-26 breach and no
invisibility: the build fails until the directory is accounted.** That holds without any change from
this story.

**Recorded because the next story will otherwise re-derive it**, and because my own
`source-text-guard-inventory.md` **structurally cannot see directory-listing guards** — this is the
second finding it was blind to by construction.

### D-11.1.7 — The shipped-slot faces are the one population nothing verifies; 11.1 closes it in-story

The builder found, while measuring D-11.1.6, that the six hardcoded-slot shipped faces are covered by
**no** metadata assertion at all — the Regular-only checks live exclusively on the catalogue population.
It asked whether closing that should be its own item.

**Ruling: in-story, not a separate item.** This story is what makes the gap load-bearing. It is the
first time a **non-Regular** face enters that population, and it enters the one population where nothing
would catch a mislabelled cut. The assertion is per-face over the shipped slots, asserting each face's
**intended** subfamily and weight — not a blanket Regular, which would be false the moment the cuts
land. Splitting it out would ship seven cuts through an unguarded surface and leave the guard orphaned
in a story with no reason to be written.

**This is instance ~13 of the run's dominant defect class** — a guard that cannot distinguish a correct
outcome from a plausible wrong one — and the first found by asking *which population is this assertion
actually quantified over?* rather than by finding a false zero. Add that question to the census axes.

### D-11.1.8 — Q2's non-additivity is real by three mechanisms and asserted by none

The builder verified rather than assumed, and the distinction produced the AC.

**`thai-dictionary` is non-additive today, three independent ways:** `parseS1Payload` rejects
`cached-bytes-mismatch` unless `cachedBytes` equals Σ `cacheAssets.bytes`, so rows contribute nothing
structurally (already red-proved by `release-payload.test.ts`'s `staleArithmetic`); `LoadScreen.tsx`
computes `verified` from `payload.cacheAssets` filtered by verified URLs, **not from `rows`**, which is
the mechanism that makes N embedded face rows safe and the one that would have been got wrong by
assuming; and `generate-offline-release.mjs` reduces `rows` into `s1VisibleBytes` **before** appending
the dictionary row.

**What is not asserted is the AC.** Nothing pins that the *rendered* total ignores `rows`. Rewrite
`LoadScreen`'s total from `payload.cachedBytes` to a row sum and every existing test still passes —
`parseS1Payload` never sees the UI. So AC3 is a LoadScreen-level test that a payload with embedded rows
displays Σ `cacheAssets`, not Σ `rows`, red-provable by making the total a row sum. That is falsifiable,
it is about the surface an author actually reads, and it is what stops the "you download 696 KB twice"
failure.

**The pattern worth keeping:** three correct mechanisms, zero assertions at the surface that matters.
Being right by construction and being guarded are different properties, and only one of them survives a
refactor.

### D-11.1.9 — DW-12 for 11.1, stated in the form that cannot be discharged by a green subset

11.1 runs the **FULL `-tags=matrix` suite unfiltered, on four targets, in-story** — never a `-run`
filter, never CI's name-filtered subset. Two facts make this non-negotiable here rather than ritual:
`TestShippedFacesReproduceFromUpstream` **is among the tests CI never runs**, and this story changes
that guard's own subject. A green from anything narrower discharges nothing.

**And it must run with `FOLIO_FONTGEN_PYTHON` set.** Measured: with it, 2125 pass / 2 fail / 5 skip and
that test passes; without it, 2124 / 3 / 5 — the third being that same test as a **could-not-execute**.
A degraded form that cannot run is not a form that can pass. (Both against the same two known failures,
`TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)`.) `TestCrossTargetByteIdentity` really executes
all four legs in 24.16s, so Docker and Node are present and the gate is not silently skipping — itself a
check worth keeping, since a four-target gate that quietly ran one leg is the same defect class again.

### D-11.1.10 — Arm A: the row type must name how the bytes actually reach the reader

**The fork, surfaced by the builder rather than absorbed.** My Q2 ruling (rows are `embedded-in-engine`)
and my D-11.1.5 three-surface ruling (each cut declared to the browser under its own family) imply
**different row types**, because these faces are delivered twice. Q2 was ruled on an analysis that turned
out to be incomplete. The builder caught that choosing either arm silently would overturn a ruling made
an hour earlier, and stopped. That was correct and is the behaviour to keep.

**My slot hypothesis was refuted, cleanly.** I had assumed engine-embedded cuts might not consume
precache slots. The three Notos refute it by existing: they are embedded in the wasm **and** hold three
of the six hardcoded slots, because a CSS `@font-face` needs a URL and the wasm's copy has none. The
builder decomposed `s1.assetCount` = 54 exactly — 31 catalogue faces, 9 Vite chunks, 6 hardcoded font
slots, 4 pdfjs fonts, 4 engine/starter/index — every one traceable to a `fingerprint()` call. **The rule
is exact: one `fingerprint()` slot = one dist asset = one cache slot.** My predicted 61 rested on a false
premise and was withdrawn.

**RULING: Arm A** — 11.1 commits the cuts engine-side; 11.3 bundles them.

1. **Q2 stays true because on Arm A it *is* true.** Not preserved by wording: on Arm A the bytes really
   are only in the wasm, so `embedded-in-engine` **describes** the delivery. Arm B would have a row say
   *"embedded in engine; no second request"* about a face the browser separately downloads — a false row
   on a pinned surface, which no rewording rescues. **A row type must name how the bytes actually reach
   the reader.**
2. **Blast radius follows acceptance.** 11.1's acceptance is that the engine lays out and paints the
   seven cuts; nothing in it requires the browser to hold the bytes. 11.3's acceptance does. Each story's
   payload cost then attributes to the thing that made it necessary, which is what makes the size record
   readable later.
3. **Margin timing — recorded as the WEAKEST of the three, deliberately.** Both arms end at margin 3;
   only the timing differs. Recorded as weak so nobody later reconstructs this ruling as having been
   about the margin. It was about row truth.

**Cost accepted knowingly:** 11.3 rewrites seven rows from `embedded-in-engine` to `cached-asset`. That
churn lands in a story already rewriting those rows, and it forces 11.3 to re-attest each row against the
delivery it then actually has — a feature. Arm B's pay-once saving is real and smaller than a false row.

**The falsifier, issued as part of the ruling so it costs no second round trip.** Arm A splits a mirrored
surface across two stories, and D-7.4.5 says a mirrored invariant moves in one commit. **If a guard exists
that reads both the Go-side face set and the browser-side `shippedFamilies` and compares them, Arm A is
not churn-heavy but impossible** — 11.1 would leave the mirror half-populated and red the suite. In that
case the builder takes Arm B without asking, records that Q2 is superseded by the mirror constraint and
by whose measurement, and states the margin as 10 → 3. **The falsifier IS the ruling.** Issued as a
precondition rather than an assertion because I have not measured it and the census is structurally blind
to it — this would be the third such blind spot.

**The obligation that is Arm A's price, and is not optional.** Deferring the slots defers the check, and
*a deferred check with no named owner is exactly how DW-162's figure aged 41 → 20 → 10 while three
stories walked past it.* So: the spec hands DW-162 forward to **Story 11.3 by name**, with the predicted
54 → 61 and 10 → 3 marked as predictions to be measured; and **the approach-warning threshold ships in
11.1 on either arm** — 11.1 is the story that knows the margin is about to hit 3, 11.3 is the story that
will be busy hitting it.

### D-11.1.11 — I wrote a stale mechanism into a standing instruction, and the correction improves the instruction

**What happened.** In re-keying Story 8.4d's trigger I justified "measure arm-to-arm in one environment"
by DW-100's finding that `s1VisibleBytes` drifts every commit — the engine row being 58% of the total and
`build-wasm.mjs` leaving `-buildvcs` at its default. The builder measured it: **that is no longer true.**
`wasm-vcs-stamp.mjs` declares `export const ENGINE_BUILD_FLAGS = ['-buildvcs=false']`, `buildEngineWasm`
uses it, and `assertNoVCSStamp` actively verifies that none of `vcs.revision`, `vcs.time`,
`vcs.modified` or `build\tvcs=` survives in the emitted binary, throwing if one does. The tree-state
input was deliberately closed. Register amended the same day, before the text was acted on.

**The instruction survives and is stronger, which is the point worth keeping.** The conclusion was right
for a reason that had expired. With the drift closed, a **same-arm repeat is available as a control**:
build the same arm twice, and only then quote the cross-arm delta against it. If the repeat is
byte-stable the delta is a measurement rather than the difference of two drifting numbers; if it is
**not** stable, that is a finding in its own right, because something is supposed to have closed that
input and evidently has not. **Never quote a cross-arm delta without stating the same-arm repeat behind
it.**

**The error class.** Not a wrong fact — a fact that *was* right, cited without checking whether it still
was. Distinct from D-11.1.2 (right shape, wrong subject) and D-12.C.1 (wrong vocabulary). This one is
**a finding quoted past its own repair**, and a decision log is precisely the instrument that makes it
easy: the entry stays readable forever and says nothing about having been fixed. Whenever a DW entry is
cited as a live hazard, check its status before leaning on it.

### D-11.1.12 — the italic PostScript names are upstream output, and the warning belongs at the assertion

The two italic derivations emit `name[6]` values that read as typos and are not:
`NotoSans-Italic.ttf` → **`NotoSansItalic-Italic`**, `NotoSans-BoldItalic.ttf` →
**`NotoSansItalic-BoldItalic`**. fontTools composes name[6] from the *source* VF's
variations-PostScript prefix, which for the italic VF is `NotoSansItalic`; the bold cut escapes it
because the roman VF's prefix is `NotoSans`.

**Ruling: record the measured value, never hand-correct it.** A shipped face is committed output of a
replayable script; editing its name table by hand puts a manual edit inside the byte-identity regime,
which is the one thing that discipline exists to forbid.

**And the warning goes in a comment at `shippedFaceSpecs` itself, not only in the spec.** `shippedFaceSpecs`
asserts name[6] exactly, so a reader who "corrects" these two rows reds the suite. **The spec is read
once; the assertion is read by whoever is about to change it** — and that is the person who needs the
warning. The comment states that the value is upstream-determined output and that correcting it reds the
suite.

### D-11.1.13 — the payload delta is measured, and D-A's estimate held

Exact, from the extracted and derived files: derived **2,025,112 B** over four cuts (Noto Sans Bold
648,284; Italic 663,520; Bold Italic 665,508; Thai Bold 47,800); static **1,111,656 B** over three
(Roboto Bold 358,188; Italic 375,320; Bold Italic 378,148). **Total 3,136,768 B = 2.99 MiB**, taking the
shipped set 11,645,836 → 14,782,604 raw — **+26.9%**.

That lands almost exactly on D-A's "~3 MB" estimate, so the owner's arithmetic held. **Worth one line in
the Delivery Log**: an estimate that is checked and found good is how the next estimate earns its trust,
and this run has spent far more words on estimates that did not.

### D-11.1.14 — the falsifier fired, and the mechanism is now standing practice

**Arm A is impossible, as D-11.1.10 said it would be if the guard existed.** It exists. Verified
independently at `folio-designer/src/font-binary-identity.test.ts:982`:

```js
familiesWithNoRule(generator, [...chromeFamilies,
  ...shippedFaceNames(fontsGo).filter((face) => !catalogueFamilySet.has(face))]).toEqual([])
```

`shippedFaceNames` parses the `Shipped()` map keys straight out of `folio-go/fonts/fonts.go`; every key
that is not a catalogue family must have an `@font-face` rule, and `familySourcePaths` resolves each rule
through a real `assets` slot, with `isSentinel` catching a rule whose slot resolves to nothing. A bold cut
cannot be a catalogue family, so all seven land in the must-have-a-rule set. **Arm A would leave the
mirror half-populated and red the suite.** So: **Arm B.** Rows are `cached-asset`, `s1.assetCount`
54 → 61, margin **10 → 3 in this story**, and the spec records that Q2's `embedded-in-engine` ruling is
superseded by the mirror constraint, naming the measurement and the falsifier it was taken under.

**The mechanism is the durable part.** Attaching an explicit, measurable falsifier to a ruling — *"if X
exists, this ruling is void and you take the other arm without asking"* — converted a wrong ruling into a
right one **with no round trip**. The cost of being wrong was one paragraph rather than one dispatch, and
the builder never had to choose between obeying a ruling and obeying the code. **Standing practice: when
a ruling rests on a fact the ruler has not measured, name the fact, name what its presence implies, and
hand over the authority to flip.** This is now the preferred shape for any ruling issued under
uncertainty, and it is strictly better than either guessing or blocking.

**The load-screen non-additivity test survives the flip, for a better reason than it was written under.**
The seven new rows are now correctly additive, but `thai-dictionary` becomes the **only**
`embedded-in-engine` row; its non-additivity is still asserted nowhere at the surface an author reads;
and at twelve rows instead of five, a row-sum total is **worse**, not moot.

### D-11.1.15 — twelve rows: itemise, because `rows` is a record before it is a screen

**Question.** AC3 itemises per face, so the load screen goes from 5 rows to 12. Aggregate them?

**Ruling: itemise.** Three reasons, the third general.

1. **`rows` is a manifest surface before it is a screen.** `verify-offline-release.mjs` checks `s1Ids` and
   `semanticLabels` by ordered exact join. An aggregate does not satisfy "the manifest states the per-face
   cost" — it destroys per-face attribution at the moment the payload grew 26.9%. The one question a
   reader will have later is *which face cost what*, and an aggregate is the single shape that cannot
   answer it.
2. **The CJK precedent is one row per face**, and `cjk-font` is already read by id. A second convention
   for the cuts gives one surface two rules, and the next story must learn which applies.
3. **Never fix a presentation problem by making the record less specific.** If twelve rows reads badly,
   that is a rendering problem with a rendering fix — grouping, a disclosure, a heading — in the
   component, not the manifest. Same instinct as DW-162, where a figure kept in aggregate prose aged 41 →
   20 → 10 unnoticed. If the screen genuinely reads badly, the Delivery Log records what was seen and a
   presentation story takes it; the data is not thinned pre-emptively.

### D-11.1.16 — count-independence, and an index-keyed RED PROOF is the defect wearing its own uniform

The builder specced the positional `s1.rows[4]` read as *"make it count-independent"* rather than *"change
4 to 11"*, because the index will move again in 11.3. Correct, and the scope is larger than it looked.

- **`release-payload.ts:57` carries three couplings on one line:**
  `cached.length !== 4 || rows[4].delivery !== 'embedded-in-engine' || rows[4].assetUrl !== rows[0].assetUrl`.
  The count becomes 11, the index becomes 11, **and `rows[0]` is itself positional** — it means "the
  engine wasm row" and says so nowhere. All three key by id.
- `verify-offline-release.mjs:95` — `const dictionaryRow = s1.rows[4]`.
- **`verify-offline-release.mjs:295` and `:325` mutate `rows[4]` and `rows[0]` BY INDEX inside `redProof`
  harnesses.** This is the part that matters. Those are the falsifiers proving the guards can go red. An
  index-keyed falsifier that lands on the wrong row after an insertion either **fails to go red**, or goes
  **red for the wrong reason** — and a red proof passing for the wrong reason is this run's dominant
  defect class wearing the costume of the thing meant to catch it. **A falsifier must target its subject
  by identity, never by position.** New rule, and it generalises past this story.
- **The precedent is in the same file**: `verify-offline-release.mjs:98` already does
  `s1.rows.find(row => row.id === 'cjk-font')`. Cited so the change reads as adopting the file's own
  better convention rather than as invention.

### D-11.1.17 — the fragment fallback stack does not change, and the reason outranks scope

`canvas-font-stack.test.ts`'s counts move; **its fragment fallback stack must not.** The builder flagged
this as the one place where the obvious edit is the wrong one, and was right for a reason stronger than
scope: **widening the stack changes what every unattributed fragment in every existing document falls back
to** — a silent rendering change to documents nobody edited, under a byte-determinism regime whose whole
premise is that output moves only when input does. That is this project's worst failure mode, and it would
arrive disguised as a tidy-up.

Written into the spec as a **`Never`**, not only as a table note: the table is read by the implementer,
the `Never` is read by the reviewer wondering why the obvious edit was skipped.

### D-11.1.18 — token gate: [K], and the reason is that there is no seam

~7,142 tokens against the workflow's 1,600 threshold. Kept whole **deliberately**, not by inheriting the
run's earlier `[K]`.

**The story cannot be split along the seam that matters.** The Go/browser split *is* Arm A, and D-11.1.14
just proved the mirror must move in one commit under D-7.4.5. Splitting by face multiplies the full
four-target matrix — the run's most expensive gate — by the number of splits. Splitting the accounting
from the faces ships binaries through unaccounted guards. **There is no seam, so `[K]`.**

**Mitigation pre-authorized:** if the implementer starts missing rows, hand it the 20-row guard table
separately, as its own artifact, and do **not** trim it. A complete table delivered twice is cheap; a
table trimmed to fit is how a directory-listing guard gets skipped — and three of this story's four
hardest blockers are that kind.

**The guard table is the best artifact this story has produced.** It names the three files my census
missed (`accounting_test.go`, `folio-go/fonts/fonts.go`, `NOTICE.md`) and separates dir-listing guards
from text-matchers — the axis my census structurally could not see. It is the **tree-shape axis done
properly**, and the closer is pointed at it as the reference implementation for the amended census.

### D-11.1.19 — one more assertion quantified over a population that just changed

`shipped_faces_test.go:assertShippedFaceMatchesSpec` contains a bare literal:
`!strings.HasSuffix(names[6], "-Regular")`. It reds on every new cut. **Parameterise it from the spec
row; do not delete it**, and record *why* beside it.

It is a **correct** assertion quantified over a population that just changed — the same shape as
D-11.1.7's shipped-slot metadata gap, which this story is also closing. Deleting it rather than widening
it would be the third instance of the class in a single story, and the only one we introduced ourselves.
**When a story changes a population, every assertion quantified over that population is either widened or
consciously exempted — never quietly dropped because it went red.**

### D-11.1.20 — a builder forbidden to touch git cannot green a guard that reads `git ls-files`

**How it surfaced.** 11.1's builder was cut off mid-implementation with no completion record. Assessing
the tree rather than assuming, the work had survived and was substantially green — but `lint` failed two
tests:

```
TestLicenceSignalCensus: census walked up only 59 committed licence files,
  but pinnedCensus records 73 of them — the walk itself looks broken
TestManifestUpToDate: lint/MANIFEST.md is out of date
```

**Neither was a defect.** Both guards are scoped to **tracked** files — `manifest.go:704` shells out to
`git -C <root> ls-files -- <dir>` — and this story's fourteen new directories were untracked. The
builder's fourteen census rows and its regenerated `MANIFEST.md` were correct the entire time; the guards
were measuring a repository that did not contain the story's files. Staging the fourteen directories **by
explicit path** (never `git add -A`) turned all four `lint` packages green with no code change.

**The structural finding, which is about the process and not the code.** The standing rule forbids the
builder to commit, add, stash, checkout, reset, revert or restore. Two of this story's guards read
`git ls-files`. Therefore **a story that adds files is structurally unable to verify its own work on
exactly the surfaces AD-26 cares most about.** That is not a bug in either rule; it is an interaction
between them that nobody had hit, because no previous story in this run added a tracked directory.

**Standing gate order for any file-adding story, from here:**
`implement → the ORCHESTRATOR stages by explicit path → licence and manifest gates → review.`
Staging is not committing, it is within the orchestrator's authority, and it is now a named step rather
than something rediscovered under a red test.

**Why this is the safe direction, and worth saying explicitly.** This is the run's catalogued tracked-only
mechanism — the one that produced false zeros in the font-host scanners until their populations were
widened to `git ls-files --others --exclude-standard`. Here the same mechanism produced a **false
FAILURE** rather than a false pass. A guard that cannot see a file and therefore *fails* is behaving
correctly under uncertainty; a guard that cannot see a file and therefore *passes* is the defect class
this run keeps finding. **Same mechanism, opposite sign, and only one of the two signs is dangerous** —
which is a useful thing to be able to say about a red test before spending an hour on it.

**Also recorded: the builder's work was fully recoverable.** Fourteen directories each with binary +
`LICENSE-OFL.txt` + `NOTICE.md`, eleven `//go:embed` directives, eleven `Shipped()` keys spelled as ruled,
twenty-one tracked files modified, all four CHECKPOINT 1 edits applied. `folio-go` green apart from the
two known baseline failures; designer **63 files / 947 tests all passing** against a 938 baseline, `tsc`
clean, `build:wasm` succeeding — which is itself the proof that `shippedFamilies` moved 6 → 13 and the
rule-count throw is satisfied. An interrupted builder is resumed and re-oriented against a measured tree,
never restarted.

### D-11.1.20a — AMENDMENT: the loud failure was a property of the guard, not of the mistake

11.1's builder improved D-11.1.20 and the improvement is the more useful half. **The tracked-only
mechanism has a silent direction, and we happened to hit the loud one.**

**The silent direction, reasoned through the code.** `TestManifestUpToDate` compares the **committed
`MANIFEST.md`** against a **live walk**, and `manifest.go:704` scopes that walk with `git ls-files`.
Regenerate the manifest *before* staging and **both sides are blind in the same way**: the generated file
omits the fourteen rows, the walk omits them too, the comparison agrees, and the test goes **GREEN while
`MANIFEST.md` silently fails to account for fourteen redistributed font binaries.** That is an AD-26
breach that passes its own gate. The correct order — stage, then regenerate — is the difference between a
red test and a shipped licence hole.

**Why we got the loud failure instead, which is the part to keep.** The licence census failed loudly
**because it is hand-pinned**: the builder wrote 73 rows from the *intended* set, so a hand-written
population disagreed with a walked one and the disagreement was visible. The manifest would have failed
silently **because both of its sides derive from the same walk**.

> **A record compared only against itself cannot detect a blind spot it shares with its own source.**

That is DW-230's one-direction finding in general form. It also means the loud failure was a property of
**how that particular guard is built**, not a property of the mistake — so nothing about this episode
licenses the inference that the tracked-only hazard announces itself. Registered separately as a live
hazard for every future file-adding story.

**And the inverted near-miss, worth recording because it runs the other way to every other instance.**
The builder's first `grep -c "go:embed" fonts.go` returned **12**. The twelfth is line 22's prose,
`// estimated) — go:embed`. The real count is eleven embeds and eleven `Shipped()` keys. It resolves
clean because `lint`'s `expectedShippedFaces` trims the line and prefix-matches `"//go:embed "`, which
that prose does not satisfy. **The repo's matcher was the strict one; the ad-hoc grep was the loose one.**
Every other instance in this run has gone the other way, and it is a concrete reminder that
strips-comments is a property of the *reader*, not of the file — which is exactly why the census tracks
it per guard rather than per file.

### D-11.1.21 — the builder disclosed that it did not author the designer half, and that is the behaviour to keep

On resume, 11.1's builder reported that its visible transcript contains step-01, step-02, the spec, the
CHECKPOINT 1 edits, the font binaries, all fourteen provenance directories, the four `UPSTREAM` entries,
`make fonts` / `make fonts-verify`, and **two** implementation subagents — Go and `lint`. It **never
dispatched a designer-side agent**; that chunk is absent from its transcript. (A first implementation
agent died on an API error after 44 tool calls having written only prose.)

**Confirmed from outside its transcript:** the git snapshot in the orchestrator's own system prompt, taken
when this session resumed, already listed every designer file as modified before the orchestrator's first
tool call. So the designer chunk landed in the part of the session neither party can see.

**The disposition is correct and is now standing practice.** The builder is reviewing that diff as
**third-party code under full adversarial scrutiny** rather than as work it knows to be right. It cost
nothing — it was going to review it anyway — and it removes the one assumption that would have made the
review shallow. **Code being good is not evidence that it was reviewed**, and an agent that cannot
remember authoring something must not treat its quality as a substitute for having checked it.

Worth noting what the builder did *not* do: absorb the gap, reconstruct a plausible account, or let the
work's evident quality stand in for provenance. It reported a discontinuity in its own memory as a fact
about the review's reliability. That is the correct failure mode for an agent that has been interrupted.

### D-11.1.22 — the release is measured, and every prediction held

```
assetCount 61   rows 12   cachedBytes 53,939,356
approach warning: 61 cache assets against a declared maximum of 64 — the margin is 3.
  The warning threshold is `warnCacheAssets` = 56 in src/release-payload.ts;
  nothing fails until the maximum is exceeded.
```

`s1.assetCount` **54 → 61**; margin **10 → 3**; twelve rows; seven new `cached-asset` rows itemised per
face per D-11.1.15; `thai-dictionary` the sole `embedded-in-engine` row per D-11.1.14. **DW-162's
discharge condition is met** — the message names the threshold, its file, and states that nothing fails
until the maximum is exceeded, which is precisely the over-versus-approach distinction the entry existed
to draw.

**AC3's non-additivity is now empirical rather than argued.** Σ `rows` = **17,238,271**;
Σ `cacheAssets` = **53,939,356** = `cachedBytes` exactly. They differ by a factor of three, so a row-sum
total could not masquerade as correct — and the LoadScreen test pins the one surface where it would
otherwise go unnoticed.

**The builder exceeded the instruction on `warnCacheAssets`, and the excess is the valuable part.** The
ruling asked for a threshold read through `readDeclaredConstant`. The builder also wrote the negative
cases — a threshold **above the ceiling** and one **below the floor** — under the comment *"than left as
a warning that can never fire."* **The obvious failure of an approach warning is not that it is absent
but that it is set somewhere it can never trip**, and nothing in the ruling would have produced that
check. Recorded because it is the defect class caught *prospectively* for the first time in this run.

### D-11.1.23 — the same-arm control exists now, it passed, and its limit is part of the measurement

D-11.1.11 replaced a stale justification with a better instruction: build the same arm twice and quote
the cross-arm delta against that repeat. **This is the first story in the run to actually have the
control, and it passed.**

**Two consecutive clean `npm run build` runs, compared on `s1.assetCount`, `s1.cachedBytes`, all twelve
rows, and every asset URL with its bytes: IDENTICAL.**

**The limit, stated because it is part of the measurement.** Both builds ran at HEAD `ffec48c` — the
manifest was written 10:03:21 and commit `4a03678` landed 10:03:56. So this is **reproducibility at a
fixed tree**, which is what D-11.1.11 asked for, and **not** a cross-commit demonstration. It must not be
read as one.

**So the vcs question was settled directly on the artifact rather than inferred from the timing.**
`strings` over `dist/assets/folio-engine.0230bd76b83f1e373a3e-DJc1Bw29.wasm` matching
`vcs.revision|vcs.time|vcs.modified|build\tvcs=` returns **zero markers**. `-buildvcs=false` is in force
**in the emitted binary**, not merely declared in `wasm-vcs-stamp.mjs`. DW-100's drift mechanism is
confirmed closed by inspection of the artifact — a better basis than reading the flag, and a better basis
than my timestamp reasoning, which I had begun to lean on before checking.

**The general lesson, and it is one this run keeps re-learning in new costumes.** I nearly reported the
repeat as evidence that the commit-hash input was closed, on the strength of a plausible interleaving I
had not pinned. The interleaving turned out to be the *other* way round — both builds preceded the
commit — so the inference would have been false even though the conclusion happens to be true. **A true
conclusion reached through an unchecked premise is not a measurement**, and it is worth exactly as much
as the premise. Checking the artifact cost one command.

**Recorded figures for the Delivery Log**, in order of durability:
- **3,136,768 B of new font binary**, measured directly from the files — needs no arm comparison and will
  still mean something in a year.
- `s1.assetCount` 54 → 61, margin 10 → 3, twelve rows, `cachedBytes` 53,939,356, all at this arm.
- The control: two consecutive clean builds byte-identical at a fixed tree; emitted wasm carries zero vcs
  markers.

### D-11.1.24 — two acceptance criteria the correct code could not satisfy, and the rule for what to do about that

11.1's builder finished step-04 with 13 patches applied, 5 deferred, and **two findings it deliberately
refused to resolve** because the honest fix was to edit an acceptance criterion — *"the one move I should
never make unilaterally."* That refusal was correct and is the behaviour to keep: a builder that edits its
own marking scheme has stopped being marked.

**(A) An AC instructed production code to read a test-only symbol.** The AC read: *"when any production
code needs the family, then it reads `shippedFaceSpecs.Family` (sfnt name ID 1) and never parses the
key."* `shippedFaceSpecs` lives in `folio-go/shipped_faces_test.go`; no non-test package can import it.
Five comments had been written pointing production code at it.

**Ruling: reword the AC and the five comments. No frozen renegotiation, because the frozen block is
already correct.** Boundaries & Constraints reads *"The machine-readable family is
`shippedFaceSpecs.Family`, **which is sfnt name ID 1** and is asserted against the binary"* — which names
name ID 1 as the authority and the symbol as the thing that asserts it. **The AC was a lossy paraphrase of
frozen text that was already right**, collapsing "the test that asserts name ID 1" into "the symbol
production code reads." No code change; the code already honours the intent's single reading. That the
never-parse prohibition is enforced by nothing is the real gap and is correctly deferred — a comment
pointing at an unreachable symbol was never going to enforce anything.

**(B) AC2 required an archive sha256 that existed nowhere for Noto Sans Thai Bold.** Options were: fetch
it, narrow the AC, or accept and defer.

**Ruling: fetched. The AC becomes TRUE as written rather than edited to fit.**
`NotoSansThai-v2.002.zip`, **4,720,990 bytes**, sha256
`af889cc673fc714060ce5e4e088fbad32aa4c0571a19958efeaff128a22da485`. The extracted
`NotoSansThai/googlefonts/variable/NotoSansThai[wdth,wght].ttf` is 218,652 B and hashes to
`5a1c559b…` — **exactly the pinned `src_sha256`** — so the archive is verified transitively, the same
standard as Noto Sans and explicitly not parity with Roboto's directly-verified archive.

Three reasons over narrowing: **D-11.1.3 already ruled it** for Noto Sans (record the digest while we have
the file), and applying that to two faces in a story but not the third would be arbitrary; **"narrow the
AC to archives this story fetched" is self-referential** and would be satisfied by fetching nothing — an
AC that can be met by doing less is not a criterion; and the tree already has the precedent, since
`notosansthailooped/NOTICE.md` records one. Plain Thai was the outlier.

**THE GENERAL RULE, because this will recur.** `bad_spec` exists to stop code being derived from a wrong
spec. When the defect is in prose **no code was derived from** — an AC that describes correct code
inaccurately — reverting re-derives byte-identical code and buys nothing but the risk of the re-derivation
differing. **The test is: would re-deriving from the corrected spec produce different code?** If provably
not, correct the spec text in place and record why.

**Three conditions, and the rule is void without all three:** the intent has exactly one reading; the code
already honours it; and **the call is the orchestrator's, never the builder's.** Without the third the
rule is just a licence to move the goalposts.

### D-11.1.25 — what step-04 actually caught, and the one that was a shipped bug

Recorded because the review layers earned it, and because three of the patches are the run's dominant
defect class found in guards that this very story had introduced.

**Guards that existed and could not fail:**
- **`TestShippedRobotoMatchesDesignerCatalogue`'s coverage witness could not fire** — `checked++` sat
  outside the `t.Run` closure, counting loop turns rather than completed subtests, so `checked !=
  len(cuts)` was unreachable **by construction**. Fixed, and the agent then **watched it fail**
  (`checked 3 of 4`) before restoring. *A witness nobody watched fail is not a witness.*
- **The cache-asset approach warning — the sole realization of an AC — was executed by no test at all**,
  and raising `warnCacheAssets` to 64, legal under the reader's own bounds, made it **silently stop firing
  with everything green**. This is the guard I ruled into existence at D-11.1.22 and praised for its
  negative cases; the negative cases checked the *threshold's* legality and nothing checked that the
  warning *fires*. Now a six-test file, red-proved twice.
- **The warning's latch was consumable by a red-proof fixture**, so under `--red-only` the single warning
  line described a deliberately mutated release rather than the real one. A falsifier eating the signal it
  was meant to prove — the same family as D-11.1.16's index-keyed red proofs.

**And the one that was a real bug, not a guard gap: `scriptFallbacks` was validated against all thirteen
families, so a script fallback could name a bold cut and render an entire script bold in every author's
document.** Re-pointed at the six upright Regulars, with a duplicate check both existing throws were blind
to. That is a shipped rendering defect closed before it existed — **in code neither the builder nor I can
remember authoring**, which is the strongest possible argument for the third-party-scrutiny disposition of
D-11.1.21.

**Also worth keeping:** `font-catalogue.test.ts` carried a six-family population against `build-wasm.mjs`'s
thirteen and is now **derived from the same slot table**, so the duplicate cannot drift again — the fix
that removes the class rather than the instance. And the OBLIQUE bit turned out to be genuinely uneven
across the shipped italics (Roboto Italic `0x0201`, Roboto Bold Italic `0x0221` set it; the Noto italics
do not), so it is now a per-face intended column **with a non-vacuity check that the column carries both
values** — otherwise it would be a blanket rule wearing a table.

**On the builder's three self-reports.** `epic-11-context.md` was promised to reviewers and omitted:
compiled context, untracked, not shipped, derived from artifacts the reviewers had — **no re-review**. But
a blind reviewer caught that the package's own description did not match its contents, which is a reviewer
checking the **frame** rather than the picture, and that is what catches a planted premise. The builder
also caught its own mis-measurement **because 227 tests across 4 packages did not match the shape it
expected**, not because anything failed. *Cross-checking the shape of a number against what it should be
is the single habit that would have caught most of this run's defects, the orchestrator's included.*

---

### D-11.0.1 — OWNER DECISION: Epic 11 gains a fourth story, so bold is reachable in a document that already exists

**The question put to the owner.** After 11.2 and 11.3 as specified, pressing B in a real document still
bolds nothing. D-11.2.1 rules that the engine resolves a weight only to a face the document **explicitly
names** and will never infer one. No document names them: `folio-designer/public/templates/starter.folio`
declares `"Roboto": ["Roboto", "Noto Sans Thai", "Noto Sans SC"]` — three regular faces — and **no
acceptance criterion in 11.2 or 11.3 touches that file.** Both stories can pass in full while a new user
pressing B is told *"Roboto has no bold face"* with `Roboto-Bold.ttf` sitting inside the product at 358 KB.
Epic 11's own opening line is *"Ploy bolds a heading."*

The two halves were separated in the question so they could be decided independently: **the format change
is deadline-bound** (it dies when 15.3 cuts the tag) while **filling in the names is just writing a
document** and carries no deadline at all.

**RULING: option four — fix the starter AND add the fourth story now.** The owner took the more expensive
option over the lead's recommended triggered deferral. Both halves land:
- **11.3 updates `starter.folio`**, naming Roboto's and the Notos' available cuts. Placed in 11.3, not
  11.2, because 11.3 is the story whose B control would otherwise be telling the truth about a document we
  could have fixed — and it keeps 11.2 a pure engine story.
- **Story 11.4, *Picking a family declares the cuts it has*,** is added to `epics.md` and the tracker. The
  family control writes a family's available variants as it declares the family, through the existing
  command family.

**What the owner bought, and what they knowingly paid.** Option 1 alone would leave every *existing*
document unable to bold until its author hand-edited JSON — the capability 3.1 MB of font binary was spent
on, reachable only outside the product. Option 4 (defer 11.4 to a trigger) was the lead's recommendation
and the owner declined it; the register has entries that aged past their triggers, and this one would have
been keyed to a file Epic 14 opens. **The price is real and was stated: 11.4 touches the family control
Epic 16 rebuilt two weeks ago and Epic 14 will rebuild again**, so it is work in a contested file, and it
makes a three-story epic four.

**Recorded because the shape recurs:** the epic was *correct* at three stories and *undelivered*. Every
acceptance criterion passed and the user-facing sentence in the epic's own opening paragraph remained
false. **An epic's ACs are not a substitute for its first sentence** — check the narrative against the
deliverable before declaring the story set complete, not after.

### D-11.2.1 — RULING: the weight mapping is declared on the chain entry and derived from nothing

**The question.** Where does `(family, weight, slope) → face` live? The orchestrator staged it before
11.2's dispatch, offering a metadata index — read sfnt name ID 1, `OS/2.usWeightClass` and the italic bits
from each `FontSet` face and group by family — as the candidate that seemed to fit 11.1's direction.

**RULING: declared on the `FontChainEntry`, exactly as D-B already ruled, and derived from NOTHING.** Not
name ID 1, not `usWeightClass`, not `fsSelection`, not `macStyle`, not `post.italicAngle`.

1. A `FontChainEntry` gains optional style-variant siblings — a **FontSet face name** on the shipped arm,
   an **asset key** on the embedded arm — using the existing `Presence` idiom, absent by default.
   `folio.FontSet` does not change shape.
2. **Coverage decides the entry on the BASE chain; style is resolved WITHIN the chosen entry.** Do not
   substitute variant names into the chain before coverage runs — *a variant whose `cmap` is narrower than
   its base would push a rune to the next entry and silently change the TYPEFACE to keep the WEIGHT*,
   which is the substitution D-B forbids by name. This trap is the lead's, and it is the kind that ships.
3. **A face name is never constructed, and never parsed.** `entry.Face + " Bold"` is the foreclosed
   naming-convention weight carrier written in the other direction. The prohibition as written in five
   places says *"never parse"*; **it binds construction identically, and 11.2's spec must say so** — a
   literal reading could otherwise honour the words while reinstating the mechanism.
4. **One answer site.** `embedded_face.go`'s rejected alternative — widening the six
   `(chain, FontSet, *fontCache)` signatures — is rejected again for the same reason. A second answer site
   is the signal the shape is wrong.
5. **Absence is a first-class result.** An entry with no declared variant resolves to **its own base
   face** and emits AC3's diagnostic. *"Nearest available face"* means **this entry's** base face — never a
   walk down the chain hunting for something bold.

**Two grounds that kill the index independently of D-B, and the orchestrator's own third item was the
fatal one.**
- **AD-8 / D-8.4.1.** `embedded_face.go` derives an embedded face's name from the **asset key** and never
  from `font.family`, precisely so a document's family and a caller's cannot substitute for one another.
  An index keyed on name ID 1 **regroups exactly those two populations**. `fontCache`'s precedence rule
  protects *direct lookup*; weight resolution asks a different question — *"which face is the bold sibling
  of this family?"* — and that question has no reserved namespace to protect it. A document embedding a
  face whose name table says `Roboto` would acquire the caller's `Roboto Bold`. Fatal, not qualifying.
- **AD-21, against a criterion 11.1 already shipped on.** 11.1 shipped on *"a face nothing names is a face
  nothing embeds."* Under inference, a face **nothing in the document names** would be shaped, subset and
  embedded because it happened to sit in the caller's `FontSet` — so two `Render` calls with the same
  document and `FontSet`s differing only by an unrelated extra face would produce different bytes.

**Cost is dissolved rather than weighed:** nothing is parsed eagerly, `fontCache` stays lazy and per-name,
and the 10.6 MB Noto Sans SC is parsed only if a rune lands on it. There is no third way to find because
the second way is not needed.

**On the evidence I offered for the index, honestly weighed by the lead.** Name ID 1 is load-bearing for
**identity and provenance** — it is what `shippedFaceSpecs` asserts the binary against, and D-11.1.24
corrected an AC that had read it as something production code consults. The `usWeightClass` and OBLIQUE
columns are a **guard table proving the binaries are what we claim**, which is the opposite of a resolution
authority. Measured, the bits *would* work (all four shipped italics set `fsSelection` bit 0; only the two
Roboto italics set bit 9) — the index is refused because it is already ruled out and breaks AD-8, not
because it could not be made to function. **That distinction matters: a design can be implementable and
still be wrong.**

**The enforcement, priced at one test AC3 needed anyway.** A chain entry `"Roboto"` with **no declared
variant**, in a `FontSet` that **does** contain `"Roboto Bold"`, plus an element declaring bold, must
render Roboto **Regular** and emit AC3's diagnostic. An implementation that constructs the name reds that
test. **DW-233's prohibition finally acquires a behavioural tripwire**, inside AC3's own subject rather
than as new scope. The repo-wide `lint` rule stays deferred.

**DW-234 does not bite in 11.2 under this ruling** — the resolver reads no name IDs at all, so the
nameID 16/17-vs-1/2 divergence stays a guard-table concern for whoever ships the first non-RIBBI weight.

**The assumption to check at 11.2's plan gate rather than at implementation** (it is D-B's own carried
assumption, and D-11.0.1 makes it sharper): **the designer must be able to write chain variants through
the existing font-chain command family without a new command kind.** If it cannot, 11.2/11.3/11.4 acquire
a command-surface change and land behind the hardening 15.2a already shipped.

### D-11.0.2 — DW routing settled before 11.2 dispatches

- **DW-233** (the never-parse prohibition enforced by nothing) → **Story 11.2**, confirmed. D-11.2.1's
  guardrail discharges it behaviourally at the cost of one test the story already owed AC3. It widens 11.2
  slightly and that widening is authorised.
- **DW-234** (two "what the binary calls itself" authorities read different sfnt name IDs) → stays
  **unowned and open**. It does not bite in 11.2 because nothing in the resolver reads a name ID.
- **DW-235** (the load screen's twelve rows mix two naming conventions) → **Story 15.0**, routed. The
  closer could only reach *"a presentation story"* and flagged it unrouted. 15.0 is the right owner
  because fetch-on-first-pick **changes which rows exist** — catalogue faces stop consuming precache
  slots — so it is the next story that must open the load screen's row rendering for its own reasons. A
  presentation fix keyed to a story that has to be there anyway is the kind that gets done.
- **DW-236** (`make fonts-verify` covers only the engine copy, never the designer mirror) → left as the
  closer filed it. Hand-mirrored pairs went 4 → 11 in 11.1, so the exposure grew; it is not Epic 11's to
  close.

**And the finding behind all four is worth more than any of them.** Of 11.1's five review deferrals, only
one had been written up; the other four were appended as **raw, unnumbered blocks lodged inside DW-231's
body** — deferred in substance, **invisible to any `### DW-` census**, one of them a duplicate of its own
host. **67 more such blocks remain from eleven earlier stories, ten of them from Story 12.3 alone.** The
register has been silently under-reporting itself for most of this run. That is a sweep story, and it is
the same defect class as everything else this run keeps finding: *a record whose own index cannot see part
of its contents.*

### D-11.2.2 — RULING: the object form, and the version that would have lied

**D-11.2.1 §1's mechanism was falsified by measurement, and the lead corrected its own ruling.** §1 said
the chain entry gains variant siblings *"using the existing `Presence` idiom"*. **There is no such idiom.**
`FontChainEntry` has no `Presence` field and no `Extra`; the shipped arm is a **bare JSON string** with
nowhere to hang a key; the embedded arm enforces **exactly one key** (`parse.go:416`). All three are
deliberate and documented. The lead had carried the phrase verbatim out of D-B without checking it —
**D-11.1.11's error class again** (a supporting mechanism quoted past its own repair), and the second time
D-B has had a supporting fact corrected.

**The verdict is unaffected and the correction strengthens it.** D-B's actual reason for "not a new format
axis" was never the idiom — it was that `serialize.go:writeFontChain` **already** emits an entry as either
a bare string or an object. Still true, and exactly what the object form uses.

**RULING — ratified:** shipped arm gains an object form `{"face":"Roboto","bold":"Roboto Bold"}`; embedded
arm becomes `{"asset":"…","bold":"…"}`; the discriminant is **exactly-one-of `face`|`asset`**; the one-key
rule widens to a **closed set of three — `bold`, `italic`, `boldItalic` — never to passthrough**. Bare
strings stay bare, so AC5 holds by construction.

**Why this is a mechanism change in service of the invariant rather than a breach of it.** `model.go:157-163`
and `parse.go:411-418` state the property in words: *"the object IS the entry's discriminant, so an
unrecognised key in it is an entry of an unknown kind, not a known entry with an unknown decoration."* The
**one-key rule is the mechanism; the property is that an unknown key cannot ride along disguised as a
decoration.** An exactly-one-of discriminant over a closed set preserves the property exactly. Generalise:
*before treating a documented invariant as a wall, separate the rule from the property it protects — they
are often not the same size.*

**It stays a lead ruling and not the owner's**, on the test that kept D-B off the owner's desk: no public
API moves, no shipped bytes move, and — checked rather than assumed — **no version rank is added**.

═══ THE DEFECT THE PROPOSAL WOULD HAVE SHIPPED: A VERSION THAT LIES ═══

**Neither the orchestrator nor the builder had this. Verified independently in code before forwarding.**

```go
// internal/template/version.go:358
func fontsRequireMajor(f Fonts) bool { … if entry.Embedded() { return true } … }
// internal/template/serialize.go:215
        if entry.Embedded() { dst = writeObject(dst, …, []kv{{"asset", …}}) }
```

**Two copies of one predicate, agreeing only because object-form and embedded are currently the same
set.** The object form separates them: `{"face":"Roboto","bold":"Roboto Bold"}` has `AssetKey == ""`, so
`Embedded()` is false, so nothing raises the version, so `versionForSave` stamps **`1.0`** on a document
carrying an object entry **no 1.x reader can decode**. `version.go`'s own comment gives the real trigger —
*"a 1.x reader decodes a chain entry as a string and never coerces, so it refuses the file outright"* — and
`folio-format.md:86` names the failure: **"a version that lies: it would claim a reader sufficient for
content that reader cannot load."** It would ship **green**, because nothing asserts the negative direction
for a non-embedded object entry. There has never been one.

**Fix, in scope as a forced consequence of AC1:** the predicate moves from *"the entry is embedded"* to
**"the entry serialises as an object"** — the property a 1.x reader actually chokes on, and the property
the comment already claims to be about. **One predicate on `FontChainEntry`, consumed by BOTH
`writeFontChain` and `fontsRequireMajor`**, so they cannot disagree. A D-7.4.5 mirrored invariant: one
commit.

**Red proof must be behavioural:** for each entry shape in a literal table, assert *(serialised entry
begins with `{`) ⟺ (versionForSave raises to 2.0)*. **Never pin the predicate against itself** — pin the
serializer's real output against the version's real output.

**NO MAJOR BUMP, ruled rather than defaulted.** The doc's own test — *would a pre-2.0 reader refuse this
file or render it wrong?* — answers "refuses, on the entry shape", so Story 8.3's existing 2.0 trigger
already covers it. 3.0 was considered and rejected on `folio-format.md:617-634`'s recorded grounds: Folio
is unreleased, and a bump would make **every** document declare 3.0 including the twenty-two fixtures that
make no font choice at all, **moving their bytes and goldens for a reason unrelated to fonts.** We widen
what 2.0 means; every reader of 2.0 that has ever existed lives in this repo and moves in the same commit.
**This is the pre-tag free window being spent correctly.**

**The general lesson, and it is the sharpest one this run has produced about duplicated logic.** Two copies
of a predicate that have *always* agreed are not evidence that they mean the same thing — they may be
evidence that no input has yet distinguished them. **A duplicated predicate is a hypothesis that the two
call sites ask the same question, and it is only tested by an input that could separate them.** This story
was that input, and the separation was silent.

═══ THREE MORE THE SAME CHANGE MUST CARRY ═══

- **A sibling must not cross namespaces.** `{"asset":"myRoboto","bold":"Roboto Bold"}` — an embedded
  regular whose bold is a **shipped** face — is the AD-8 substitution **smuggled inside a single entry
  where no precedence rule can see it**. A `face` entry's siblings are FontSet face names; an `asset`
  entry's are asset keys. Cross-namespace = located load error, asserted both directions.
- **`requireEmbeddedFaceLicence` must run on a variant asset key** (`parse.go:433`), or a document can
  carry an **unlicensed embedded bold face** — AD-26 / I-7. `folio-format.md:598`'s wording must widen with
  it, or the requirement does not reach a sibling.
- **The sibling walk order is DECLARED, never map order.** `newEmbeddedFaceIndex` pins determinism on
  chains in sorted name order and entries in authored order; siblings are a **third axis**. Fixed order in
  the walk and in `writeFontChain`'s key emission, or round-trip bytes move (AD-22).

═══ THE CLOSED SET IS CLOSED AT BOTH ENDS, WITH ONE AUTHORITY AND A LITERAL FACING IT ═══

A closed set enforced only in the parser is a rule with no declaration; one declared in prose and enforced
by a `len()` is the split the guard census was about. So: **the parser holds the set once as a named
enumeration** (the authority); **the refusal message is DERIVED from it** — three hand-written messages go
stale otherwise (`parse.go:407`, the `default` branch, and the "carries no other key" sentence), and Story
8.3's own comment says why: *"unpinned wording in a refusal is wording that goes stale silently and sends
the author to fix the one thing that was not wrong"*; **the test holds a LITERAL list** and asserts it
equals the parser's set and that the format doc carries a row per key. **Literal-vs-derived, never
derived-vs-derived** — a doc assertion reading the same enumeration the parser reads moves with it and is
vacuous.

**The set is three keys, closed, and the doc states the closure AND its price.** An open extension point
(`{"face":"X","variants":{…}}`) was rejected: it surrenders the unknown-key refusal that is the whole
property this change exists to preserve, and it is speculative generality D-A already foreclosed. A future
weight or width axis costing a version bump post-tag **is what versioning is for.**

**And the doc row must state the type difference:** `style.bold` is a boolean; a chain entry's `bold` is a
face name or asset key. One token, two meanings, one document.

### D-11.2.3 — the format doc moves in four places, and two existing tests invert

**Four sites in `folio-format.md`, all in one commit under D-7.4.5:** `:192` *"A chain entry has exactly two
legal shapes"* (a normative claim, not a caption — the builder caught this); `:47` the 2.0 trigger, from
"an embedded-face entry" to "an object-form entry"; `:598` the licence requirement, widened to reach a
variant asset key; and the chain-entry row itself gaining the three keys with their type difference stated.
**Fork B's assertion checks only the fourth** — widening it to police the other three is a different and
much larger test.

**Two existing tests, and the one that matters was not the one I asked about.**
- `fonts_embedded_test.go:621` — **subject unmoved, ground moved**, from "exactly one key" to "not in the
  closed set". `requireLoadError` asserts **the field only**, so the test **cannot see** the change and
  would keep passing with a false message. Add the message assertion; the name stays.
- **`fonts_embedded_test.go:620` is a behaviour REVERSAL**: `{"face": "Noto Sans"}` is pinned **today** as
  a load error and becomes **valid**. Replace the row, do not re-explain it; its replacement is the genuine
  no-discriminant case (`{"bold": "Roboto Bold"}`). **And add a both-discriminants row** or "exactly one
  of" is asserted in neither direction.

*I asked about the row whose reason changed and missed the row whose answer inverted, one line above it.
Asking "what does this test now mean?" found less than asking "which currently-red case does this make
green?"*

### D-11.2.4 — an absence in a subagent's return is a lead, not a result

11.2's builder reported two housekeeping findings — that `Main.dc.html` does not exist, and that FR57 has
no definition. **I checked both before acting** (I had already told the builder I was "handling the
design-file gap") and **both were wrong**: the mockups directory holds six tracked files including
`Main.dc.html` at exactly the cited path, and FR57 is defined at `epics.md:121` with a coverage row at
`:350`.

**The builder's own diagnosis is better than my correction and is adopted as the rule.** I said *report the
search that produced the absence*. It found the sharper thing: it had applied that discipline **twice in
the same session to its own greps** — catching an `AD-8:` search that false-zeroed on em-dashes, and a
`git ls-files` glob whose positive control came back empty — and then relayed a **subagent's** absences
untested. **Text arriving from a subagent feels like a finding; it is raw input with the same failure modes
as one's own shell, minus the ability to see the command.**

> **An absence in a subagent's return is a lead, not a result. Verify it, and report the search rather than
> the conclusion.**

It generalises past housekeeping: **a reviewer's "nothing enforces this" is a claim about a population and
is worth exactly what that population is worth.** This is now a standing rule for every agent in the run.

**FR57 earned its citation on the way.** Its own wording — *"resolved per rune through the **declared**
chain"* — is independent textual support for D-11.2.1: the ruling is not layered on the requirement, it is
what the requirement already said. Now cited in the spec's Design Notes so a later reader has it before
re-litigating.

### D-11.2.5 — an acceptance criterion that knows how it would be faked

AC2 (tables take the weight) is written as its own criterion **with its failure mode inside it**: the table
header arm calls `chainFaceNames` **directly at `table_render.go:707`, bypassing `fontChain`**, so an
implementation resolving style in `fontChain` alone **passes every text AC and silently fails this one**.
Stating the mechanism in the criterion means it cannot be satisfied by a table that happens not to
exercise the header path.

Recorded as a pattern worth repeating: **an AC that names how it could be faked is harder to fake than one
that names only what it wants.** It is the same instinct as a red proof, applied to prose.

Also settled here: 11.2's builder held the spec at `draft` rather than freezing `<frozen-after-approval>`
while Fork A was open — three I/O matrix rows turned on the answer, and only a human can reopen a frozen
block. **Freezing to look like progress and reopening later is strictly worse than waiting.** Correct, and
recorded so the pattern is available to later builders.

### D-11.2.6 — `patch` over `bad_spec` when the spec was right and the code was not

11.2's builder routed all 16 review findings as `patch` where step-04 says to prefer `bad_spec` when in
doubt, and flagged it as the call it most wanted second-guessed. **Ratified**, and the rule already covers
it from the other side.

**D-11.1.24 says `bad_spec` exists to stop code being derived from a WRONG SPEC**, and its test is *would
re-deriving from the corrected spec produce different code?* Here the spec was **correct** — the frozen
Boundaries, AC4 and the Design Notes all specified the behaviour — and the **code failed to honour it**.
That is a bug. Re-deriving ~1,600 lines would have reproduced the same code at high cost **and discarded
the review that found the defect**, which is the part that makes the mechanical route actively worse than
useless here.

*"Prefer `bad_spec` when in doubt"* is a tie-breaker for doubt. The builder did not have doubt; it had a
measurement. Recording the invariant in the Spec Change Log is exactly what a `bad_spec` amendment would
have written, so nothing was lost.

**Two related calls, ratified as one rule.** The implementer changed production code after being told to
escalate instead, **and said so plainly**; the builder let it stand because the spec was unambiguous and
the code disagreed with it. Same rule: *a disagreement between clear spec and code is a bug, not an intent
gap.* **The failure mode to police is the silent change, not the reported one** — an implementer that
deviates and announces it is doing what you want.

**And the implementer pushed back on two patch instructions and was right twice.** P9 as framed was
**vacuous** — `versionForSave` never lowers, so it answers 2.0 regardless — and was restated over
`versionRequiredByContent` with both directions pinned. On P5 it fixed its own warning duplication,
measured that the shipped `TEXT_MISSING_GLYPH` twin has the identical defect, and **deliberately did not
widen the fix**, citing the Ask First fence. *Declining to fix a bug you can see, because fixing it is out
of scope, and registering it instead, is the discipline that keeps stories closeable.*

### D-11.2.7 — the verification-gap review layer runs ALONE. Standing rule.

11.2's builder reported: *"A reviewer mutated the working tree while the other two were reading it… I
verified the tree byte-for-byte afterwards and it was correctly restored, but that is luck, not design."*

**It is worse than a race on files.** The verification-gap layer's whole method is to **revert code and
watch a test stay green**. So while it runs, the other layers are reading a tree that may not be the tree
under review, and a finding they report may be about a mutation rather than about the code.

**The asymmetry is what makes it a standing rule rather than a preference.** A false *finding* is
recoverable — someone checks it and it dissolves. A **false CLEAN is not**: a reviewer that reads reverted
code and sees no defect **reports nothing**, and an absence of findings is indistinguishable from an
absence of defects. It is D-11.2.4's lesson arriving through a different door — an unexamined absence,
this time manufactured by a peer.

**RULE: the verification-gap layer runs alone**, before or after the read-only layers, never beside them.
Binding on every remaining story in this run. That layer **proved three of 11.2's findings**, so the
answer is to sequence it, never to drop it.

### D-11.2.8 — an assertion whose two sides could be equal is not yet an assertion

The new leading test for the metrics-chain defect **would have passed either way**: Noto Sans Thai and Noto
Sans SC both give 13920, so the assertion could not distinguish the fixed code from the broken code. The
implementer caught it, swapped the fixture's faces to **13920 vs 12732**, and — the durable half — **added
a guard pinning that the two values genuinely differ.**

Without that second half a later face substitution silently restores the vacuity and nobody learns. With
it, the vacuity becomes a failure.

> **An assertion whose two sides could be equal is not an assertion until something pins that they are
> not.**

This is the run's dominant defect class in its purest form — a guard that cannot distinguish a correct
outcome from a plausible wrong one — and the first time it has been caught **inside a test being written
for that very class.** Add "could both sides of this comparison be equal?" to the census questions
alongside *which population is this quantified over?* and *is this record enforced in both directions?*

### D-11.2.9 — two things about 11.2 worth keeping, and one number to watch

**The Matrix Test Audit is not a formality.** It found the metrics-chain defect: one uncovered matrix row →
a missing test → a production bug. That is the complete chain from coverage gap to shipped defect, and it
is the clearest evidence this run has produced for keeping the audit.

**AC2 caught nothing, and that is the criterion working.** D-11.2.5 wrote the failure mode *into* the
criterion — that the table-header arm reaches `chainFaceNames` directly and bypasses `fontChain`. The
implementer read it and routed style into `chainFaceNames` from the start. **A guard that changes
behaviour before it can fire has already paid for itself**, and it will not appear in any findings count.

**The number to watch: the AC3 Warning is per (element, distinct rune).** `worked-example.json` emits 16.
**For CJK body text that is thousands per render.** Per spec, and not a defect — but it is the kind of
figure that becomes a defect at a scale nobody tested. Registered for 11.3/11.4 to revisit.

### D-11.2.10 — the data-loss path is latent, and split across two owners rather than raced between them

11.2 created a path where a chain edit destroys declared variants — `setFontChain`/`addFontChain` rebuild
entries from a `[]string` — and its own Ask First fences both halves out. The builder registered it as
*"whichever story lands first must close it."*

**Re-routed, after measuring reachability rather than accepting it.** No live path destroys variants today:
the font-chain command builders have **no production caller** (`FontChainEntry` and `fontChainCommand`
appear nowhere in `App.tsx`; Story 16.9 deleted that UI), and the one live command **cannot rebuild an
existing chain** — `embedFontFamily` refuses at `component_commands.go:3356`, *"a font chain named %q
already exists"*. It creates; it does not overwrite. **Severity is LOW today and HIGH the moment 11.4
builds the surface.**

**"Whichever lands first must close it" names two owners, and that is how an item gets dropped by both.**
Split instead: the **destruction** half is **Story 11.4's hard precondition**; the **projection** half —
`CanvasFontChainEntry` cannot see variants, so the designer can author a weight it cannot read back — is
**Story 11.3's**. Neither is optional and **Epic 11 does not close with either open.**

*Generalisable: a deferred item whose trigger is a race between two stories has no owner. Name one story
per half, or accept that neither will do it.*

### D-11.2.11 — RULING (DW-241): a variant that names its own base is a load error; a sibling collision is legal

**Verdict.** A self-referential variant — `{"face":"Roboto","bold":"Roboto"}` — is a **located load error**
at `fonts.<chain>[<i>].bold`, and the same on the embedded arm (`{"asset":"k","bold":"k"}`). One predicate
per arm, measured against **that arm's own discriminant value**; an asymmetry would be arbitrary and would
become a "which arm am I on?" trap. **A cross-variant collision — `{"face":"Roboto","bold":"X","italic":"X"}`
— is LEGAL and silent.**

**AC3's trigger does not move.** It stays a *declaration* property ("no variant declared"), because under
this verdict the self-referential case never reaches the resolver at all. **DW-233's tripwire is untouched
— option 1 dissolves that worry rather than managing it**, which was the deciding practical difference
against option 2.

**The structure this reveals, and it is exactly one axis: the base is privileged.** Every variant is
measured against `face`/`asset`; **no variant is ever measured against a sibling.** One axis, not a
lattice. The reason is that only the base collision is something the format can honestly call meaningless:
`bold: "Roboto"` on `face: "Roboto"` **reduces to no declaration at all** — byte-identical to declaring
nothing, which is precisely the state AC3 exists to announce, and this door bypasses the announcement. By
contrast `bold: "X"` and `italic: "X"` declares a real face distinct from the base; it may be an odd choice,
but it is a **choice**, it is legible in the file, and bold text visibly does not come out as regular. The
format models no face semantics — D-11.2.1 forbids the engine to read a name table — so it has no ground on
which to call one variant assignment more wrong than another. It has exact ground for the one that
collapses to the base.

**The precedent is one line above the rows 11.2 was editing.** `parse.go:400` already refuses an empty face
name — *"a font chain entry must name a face — an empty string names none"* — and D-8.3.2's note records
why: before Story 8.3, `["Noto Sans", ""]` *"LOADED AND RENDERED — resolveRuneFace silently skipped the
empty entry and drew with Noto Sans, which is the silent substitution AD-8 forbids by name."* **A
self-referential variant is that same defect with a value in it instead of an empty string.**

**Option 3 refused on a category ground, not a product one.** "Deliberate suppression" is coherent, and it
is a **warning-suppression channel smuggled into a font-resolution field**. Its effect is not "this author
accepted the shortfall" — it is that the document renders bold-as-regular **with no signal to anyone**,
including a colleague who receives the file and never saw the warning. If authors should be able to
acknowledge and silence a diagnostic, that is a diagnostics feature under AD-14, designed once, applying to
every warning, with the acknowledgement legible as one.

═══ MY DEADLINE REASONING WAS INVERTED, AND THE CORRECTION IS THE MOST USEFUL PART ═══

I wrote that option 1 *"forecloses option 3 permanently"* and that options 2 and 3 *"can wait"*. **It runs
the other way:**

- **Refusing now is the REVERSIBLE direction.** Accepting later what you refuse today is a **widening**, and
  widenings are free after the tag, always.
- **Accepting now is the IRREVERSIBLE one.** Refusing later what you accept today is a **narrowing**, and
  D-7.8.3 makes narrowing free exactly once, before 15.3.

**So option 1 preserves the freedom to adopt option 3; options 2 and 3 spend it.** And **no document in
existence carries a self-referential variant** — the object form shipped at `d0ded7e` and `starter.folio`
does not gain variants until 11.3 — so the narrowing costs zero documents now and would cost real ones
later.

**Generalise it, because I will make this error again otherwise:** *the reversible direction is the strict
one.* Refusing is undoable by widening; accepting is undoable only by narrowing, and narrowing is the thing
with a deadline. When a permission decision is uncertain and a tag is coming, **strict is the option that
keeps the choice open**, which is the opposite of how "let's not foreclose it" instinctively reads.

**Guardrails.**
- **Disclose what the check does NOT catch**, in the code and in the format doc row. It is string equality
  against the entry's own discriminant. It cannot see `{"face":"Roboto","bold":"Roboto Copy"}` where two
  FontSet keys hold identical bytes — that renders bold-as-regular silently and nothing can detect it
  without reading binaries, which D-11.2.1 forbids. Registered separately, because **a check whose limit is
  unstated ages into a false reassurance.**
- The refusal message is **derived from the closed set**, never hand-written — the three keys must not be
  spelled a second time in a sentence.
- **Both directions asserted:** a self-referential variant reds; a cross-variant collision **loads and
  renders**, so the narrowing cannot quietly become "no two variants may agree."

**OWNER: Story 11.4**, assigned by the orchestrator (the lead ruled the verdict; the placement was mine).
It is a **narrowing and must land before Story 15.3 cuts the tag.** If 11.4 were ever to fall after 15.3
this becomes an escalation rather than a deferral — a before-the-tag narrowing with no named owner is the
shape D-7.8.3's window exists to prevent.

### D-11.2.12 — the 11.3 seam: split the golden out, and this is the INVERSE of D-11.1.18

**Of everything assigned to 11.3, exactly one item moves a rendered byte: DW-237's bold golden.** The
projection (DW-239), the canvas paint, the B control's third state, the no-synthetic contract test and
DW-240 are all designer-side; `starter.folio` is a document edit that reds nothing on its own.

**RULING: the bold golden leaves 11.3 and becomes its own story. Everything else stays together.**
Consequence: **11.3 runs no `-tags=matrix` suite at all.**

**And D-11.1.18 does not apply here, though it will be quoted.** There, splitting would have **multiplied**
the four-target matrix because both halves needed it, so `[K]` was the saving. Here the matrix is forced by
**one artifact**, so splitting it out **removes** the matrix from the larger half. *Same gate, opposite
arithmetic, opposite answer.* Recorded together so the two rulings are not read as inconsistent.

**What remains must not split further — it is one mirrored invariant (D-7.4.5).** The canvas is told the
**resolved outcome**, not the requested flag, so the B control's third state is *derived from* the
projection. A projection shipped without its consumer is unconsumed; a control shipped without its
projection is guessing.

**Sequencing that falls out of the seam:** `starter.folio` must be written **before** the paint can be
demonstrated, not left as a closing tidy-up — 11.3's acceptance needs a document that declares variants,
and the starter is now inside 11.3. *A story that leaves its own fixture until last discovers at the end
that it had nothing to test against.*

**DW-240 placed on 11.3's side** — a designer-side read-back that moves no rendered byte.

### D-11.2.13 — DW-243 stays unassigned, with one condition

Leave it unassigned, but **re-price it once at Epic 11's boundary gate rather than never.** The open
question is narrow and explicitly unverified: `forChain`'s stated purpose in `table_render.go` is that a
located capability error names the chain the label draws through, so **as filed this is an error-address
defect and LOW is right.** What has not been measured is whether **11.2 gave a chain-scoped cache a second
job — variant resolution** — in which case a footer row shaping through the unscoped cache would resolve
against the **wrong chain's variants**, which is a rendering defect and not a message defect.

One read at the gate settles it. **If it comes back address-only, leave it unassigned with that measurement
attached**, so the next person does not re-ask. *An entry that has been deliberately left alone is only
distinguishable from one nobody looked at if the looking is recorded.*

### D-11.3.1 — Story 11.2 shipped a canvas regression, and it is the orchestrator's to own

**Verified in code before ruling on anything else.** `App.tsx:3394` sets
`'--text-font-weight': component.bold ? 700 : 400` **from the requested flag**; `App.css:152` feeds it to
`font-weight`; and `font-catalogue.test.ts:536` asserts the generated `@font-face` rules carry **no**
`font-weight` descriptor — *"a font-weight descriptor would declare a weight matrix this story does not
ship."* The fragment separately receives `fontFamily: shippedFaceFamily(fragment.face)`.

**So since `d0ded7e`, a real `Roboto Bold` fragment has been painted in the bold face and then
synthetically emboldened on top of it. The canvas has been double-bolding.**

**Nobody violated anything.** 11.2's ACs did not reach the canvas and neither did its review; the browser
faking predates it and was *correct* while every shipped face was Regular. What 11.2 changed is the other
side: once a real bold face reaches the fragment, the faking stops being a substitute and becomes a
duplicate.

**But the honest description is that 11.2 made the canvas worse in exchange for making the engine right**,
and it went unnoticed until 11.3's builder traced the paint path end to end. **`d0ded7e` is the
orchestrator's commit, so this is mine to own**, and 11.3's plain-terms opener says so rather than claiming
"the canvas now paints the real face" — for the interval between the two stories it painted a doubly-bold
one, and **I-2 forbids synthetic emboldening, which means we were doing the forbidden thing.**

**The generalisable shape, and it is not "the review missed it":** *a correct behaviour on one side of a
boundary can be turned into a defect by a change on the other side, without either side being wrong when
it was written.* The faking was right; the engine change was right; the pair became wrong. **A story that
makes one half of a mirrored pair truthful should ask what the other half was compensating for** — the
compensation is invisible while it is load-bearing and only becomes visible once it is redundant.

### D-11.3.2 — the B control's third state, and quantifying "the family has no bold face"

**Q1 — RULED (c): a genuine third state, neither on nor off, always operable.**

The deciding argument is the AC's own words: *"states that this family has no bold face **rather than
appearing to be on**"*. A disabled control still renders as on-or-off, so that sentence rules out both
disabled options. **(a) is refused outright** — a state where `bold: true` cannot be cleared is **a control
that has taken the document hostage**, and I-5's posture is that the panel must never leave the author
unable to reach what the document carries. (b) fixes the trapdoor and still lies about the
absent-and-unset case.

**The builder applied D-11.2.11's reversibility asymmetry back at me, correctly** — (c) is the strict
option because it refuses to render a reachable-looking ON state, and relaxing later is a widening. Noted
that the asymmetry is *weaker* here: this is a UI state, not a format narrowing, so nothing expires at
15.3. It still points the same way.

**Required: the state must be reached by the route that reaches it in life** — bold a Roboto element, then
switch its family to a CJK-only chain — not by constructing a projection by hand. *A state only reachable
through a fabricated fixture is a state nobody has shown is reachable.*

**Q2 — RATIFIED: "no bold face" quantifies over EVERY entry in the chain, not the first.**

The counter-example settles it: first-entry-only reports *"no bold face"* for `["Noto Sans SC","Roboto"]`
while Latin bolds perfectly well. And `App.tsx:2792`'s `declaredChainEntry` returns exactly `entries[0]`
and sits three functions from the call site — **so the spec names it as the wrong function to reuse here**,
rather than only stating the right rule. A correct rule beside an available wrong helper is a rule waiting
to be violated.

**The builder's vacuity catch is the better half**, and it is D-11.2.8 applied *prospectively* for the
first time: `starter.folio`'s first entry is Roboto, which declares bold, so **both rules agree on it** and
a test written against the starter alone cannot distinguish them. A distinguishing fixture is required.
Second time this run a fixture had to change to make an assertion capable of failing.

### D-11.3.3 — Q3: the right answer for a reason narrower than the one offered

**RULED: follow the struct's convention** — `TableColumnsProjection` spells absence as the zero value, so a
bool collapses committed-absent with committed-`false`, and the limit is disclosed exactly as
`headerStyle.fontSize: 0` already discloses it. **Consistency inside one struct beats inventing a second
idiom for one field**, the same reasoning that settled plain-`string` over `Presence` in 11.2.

**The builder's justification was doing more work than it could bear.** It argued the collapse is safe
because *no command can write `false`*. But 11.2 shipped `tableHeaderStyleFields` with nine fields, so the
Go command layer **does** accept header bold, and `"bold": null` decoding to `present(false)` means
committed-false is representable. Unwritability is not the ground.

**The actual ground: `CanvasProjection` is not the file format.** It is engine↔browser, both in this
repo, moving in one commit — **so it is not tag-bound, and a tri-state can be added at any time, before or
after 15.3, for free.** D-11.2.11's asymmetry does not apply here at all. That makes the cheap choice both
right *and* reversible, which is a stronger position than the one argued from.

**One conditional, and the builder was told to check it rather than inherit it:** the third state must be
derivable **without** distinguishing committed-absent from committed-`false` — it needs
*declared-true-but-no-face*, which survives the collapse. **If a case is found where the control genuinely
needs absent-vs-false, the collapse is lossy where it matters and Q3 flips to tri-state.**

### D-11.3.4 — a Code Map is measured before the story edits the files it anchors

11.3's builder found that **every Go anchor in 11.2's Code Map had moved** — `render.go:1148` is
`lookupFontChain`, not `fontChain` (now `:1194`) — because **11.2's Code Map was measured at `102e1fc`,
before its own implementation commit `d0ded7e`.**

**This is structural, not carelessness.** A Code Map is measured at the plan gate; the story then edits
precisely the files it anchored. **Every line number in a Code Map is stale by the time the story it
belongs to closes**, and it is stale by exactly the amount the story changed.

**So D-000.4's "cite by symbol" is not stylistic advice — it is the only part of a Code Map that survives
the story's own diff.** Recorded because the register and two dispatch chains have now propagated stale
anchors from that map, including mine.

**Also corrected:** `folio-format.md`'s worked-example fence is at **846-957**, not `:876` — that line is
mid-table.

### D-11.3.5 — three findings that change 11.3's work, one of which would have shipped green

**A guard that cannot see the defect it is written for.** `canvas-authority-contract.test.ts:46` bans
`property: literal-value` pairs — which works for `white-space: normal` because no legitimate use exists.
But the value 11.3 must ban is **`var(--text-font-weight)`**, so a literal-value pattern **matches neither
of the two lines this story deletes** and the new AC2 guard would have **shipped green over the defect it
was written to prevent.** Scoped to the painted-document surface and written against both spellings, with
the chrome's seven legitimate `font-weight: 500` rules and its deliberate `.property-fx` italic noted as
outside it.

**A deletion with no witness.** The two custom properties are asserted by nothing, so removing them reds
nothing. The positive control is what makes this actionable rather than a suspicion:
`--text-line-baseline` **is** asserted (`App.test.tsx:1433`) and `--text-ink` is (`:2293`) — same file,
same idiom, so the absence is a hole rather than a convention. The deletion gets its own positive
assertion, or *"we removed the faking"* is a claim with no witness.

**A stale comment 11.2 falsified.** `table-style-command.ts:34-39` still says `bold`/`italic` *"have no arm
in the engine's header cascade to resolve from"*. 11.2 gave them one. **Retire it by editing, not
deleting**, so the history reads. DW-240 has **five** mirror sites, not the two the register named.

**And two starter details that would have produced a first-save diff:** `{"face":"Noto Sans SC"}`
canonicalises back to a bare string (a variant-free object is not `SerialisesAsObject()`), so **SC goes in
as a bare string**; and `"version": "2.0"` is **mandatory**, not optional.

**No release build for 11.3** — the starter is fingerprinted only into gitignored regenerated output, and
`verify-offline-release.mjs:113` class-checks `.folio` rather than pinning a digest.

### D-000.30 — the e2e suite is NOT dead, and DW-193's wording would have made me plan as if it were

**Measured before Epic 13 dispatches, because fifteen designer stories were about to depend on the
answer.** DW-193 says *"the e2e suite is compiled but never executed in CI"*, and the lead's grounding
carried it forward as *"the e2e suite is compiled and never executed"* — the load-bearing four words having
quietly lost *"in CI"*. I had accepted that framing and written it into 11.3's dispatch.

**What is true:**
- **CI runs only the typecheck.** `.github/workflows/ci.yml:249` runs `npm run test:e2e:compile`, which is
  `tsc -p tsconfig.e2e.json --noEmit`. `git log -S'npm run test:e2e"'` over that workflow returns
  **nothing** — CI has **never** run the real suite, going back to `05c8c70` when the compile step was
  introduced.
- **42 test cases across 16 spec files** are covered by that typecheck and by no execution in CI.
- **Playwright 1.63.0 is installed with chromium browsers present**, so the suite is runnable on this
  machine — it is not blocked, merely unscheduled.

**What is FALSE, and this is the part that changes the plan:** the suite is **not** unexecuted. The most
recent change to `folio-designer/e2e/` is commit **`0c0f3e9` (2026-09-06)** — *"Fix the two e2e failures
Epic 12's boundary gate found, one of them a regression"* — touching `band-boundary-drag.spec.ts` and
`image-asset.spec.ts`. **Epic 12's boundary gate ran the suite, it failed, and one of the two failures was
a genuine regression.** The suite works, it catches real defects, and somebody runs it — at epic
boundaries, by hand, not per story and not in CI.

**Why the distinction is worth a decision entry rather than a footnote.** *"Never executed"* invites the
conclusion that the suite has rotted and that turning it on is an unbounded excavation. **The measurement
says the opposite: it was green enough at Epic 12's gate that exactly two failures surfaced, and both were
fixed the same day.** Those are very different inputs to the question *should Epics 13 and 14 proceed
before e2e is in CI?* — and I would have answered it wrongly from the register's wording alone.

**Consequences for the remaining program, recorded now rather than discovered at Epic 13:**
1. **Per-story verification for Epics 13 and 14 does not include e2e**, and every designer story's
   Verification section must say so — a green `test:e2e:compile` is a **typecheck**, and reading it as
   coverage is the false-clean shape (D-11.2.7).
2. **The boundary gate is the net**, and it is the only net. That makes each epic's boundary gate
   load-bearing in a way it is not for engine epics, and it must run the suite **unfiltered** — the same
   rule D-000.28 established after I opened Epic 12's gate with a filtered matrix run.
3. **Story 15.2 — "CI's red means something" — is the right owner** for putting the suite into CI, and
   this measurement makes its scope knowable rather than open-ended.
4. **Fifteen designer stories will land between now and then.** Each one that a boundary gate later finds
   broken is a story that closed green. That is a real cost and it is now a priced one rather than an
   invisible one.

**The register entry itself is wrong as written and must be amended** — DW-193's *"never executed"* becomes
*"never executed IN CI; executed by epic boundary gates, most recently at Epic 12's, which found two
failures including a regression (`0c0f3e9`)"*. Assigned to 11.3's closer, since the tree is mid-story.

**The general lesson, and it is the third time this run:** *a register entry ages into a stronger claim
than it was written to make.* DW-193 was accurate on the day it was filed and the four words that bounded
it were dropped in one hand-off. **D-11.1.11 said to check a DW entry's status before leaning on it; this
says to check its scope too** — not just *is it still true?* but *is it still saying only what it said?*

### D-11.3.6 — F1: the combined cut. A generalisation that fixes the state and breaks the statement.

**The gap.** `boldItalic` is projected across the whole new seam — Go emits it, `engine-protocol.ts` types
and guards it — and **read by nothing**: `chainDeclaresCut` is typed `field: 'bold' | 'italic'`. So an
element with **both** flags set, on a chain declaring `bold` and `italic` but **not** `boldItalic`,
resolves to no declared variant, falls back to the base face, warns — and **both controls read plainly
on.** That is exactly the state AC3 exists to prevent, arriving through the one combination the spec's I/O
matrix never enumerated. The builder owned the omission as its own.

**RULING: `patch`, not an intent_gap or a `bad_spec` loopback.** AC3's principle — *the panel never shows a
state the document cannot reach* — determines the rule even though the matrix missed the row. Under
D-11.1.24 the test is *would re-deriving from a corrected spec produce different code?* It would produce
**this** code plus one generalised predicate. Reverting 17 files and ~965 lines to re-derive a predicate is
the wrong trade, and D-11.2.12 already fenced the story as indivisible. **The missing matrix row is a spec
defect; the repair is a patch plus a Spec Change Log entry.**

**But the builder's proposed generalisation was HALF a rule, and applied as written it makes the panel say
something FALSE.** Its predicate — *"the cut the element's resulting (bold, italic) combination requires
has no declared face"* — is correct. Trace it: both flags set, `boldItalic` absent, so B's required cut is
`boldItalic`, so **B enters the unavailable state and says "No bold face in this family."** There IS a bold
face; the chain declares one. **The generalisation fixes the state and breaks the statement, and a panel
that lies precisely is not better than one that lies vaguely.**

**The whole rule, three parts, all required:**
1. **Predicate as proposed** — `cutAbsent` is computed against the cut the element's resulting combination
   requires, not against the control's own axis.
2. **The sentence names the MISSING CUT, never the control.** `boldItalic` absent → *"No bold italic face
   in this family."* Derived from the cut, so there is **one sentence per cut rather than one per
   control**, and it cannot go false the way the current phrasing does.
3. **When the missing cut is the combined one, the reason is stated ONCE for the pair.** Both controls
   enter the unavailable state — both are implicated and marking only one would imply the other is fine —
   but the sentence appears once. **This also disposes of P9's double-announcement bug** (folded into
   `aria-label` *and* a visible non-`aria-hidden` `<p>`, phrased three ways); the combined case would have
   made it a quadruple.

**And the part that makes this state acceptable at all: both controls stay operable, and here that is the
escape route rather than merely Q1's rule.** Turning off *either* B or I lands the element on a combination
the chain **does** declare. **The combined-absence state is self-resolving through the very controls the
author is looking at** — which is the difference between *"this family cannot do what you asked"* and
*"this family cannot do those two at once."* The second tells the author what to do next, and it is the
accurate one.

### D-11.3.7 — the guard written for the defect could not see the defect. Twice, in one story.

**P1/P2.** I flagged this hazard at 11.3's plan gate (D-11.3.5), the rules were written against it, and
they **still** shipped two holes. The builder found them by **running the regexes rather than reading
them** — the same method that has now caught the same class three times in this run.

```
RULE 16   RED    style={{ fontWeight: 700, ...(a ? {f} : {}) }}    <- weight FIRST
          GREEN  style={{ ...(a ? {f} : {}), fontWeight: 700 }}    <- weight AFTER the spread
```

`[^}]*` stops at the first `}`, and **the real canvas-text style objects in `App.tsx` already contain a
conditional-spread `{}`** — so **the most natural reintroduction point, in the very file the rule guards,
is invisible to it.**

Rule 15's `font:` shorthand hole is worse than hypothetical: **this same commit's
`.property-toggle-unavailable` rule uses `font:` shorthand precisely to override a weight.** *The idiom the
guard cannot see is already live in the stylesheet the guard protects.*

**Standing addition to the census questions:** for any guard expressed as a pattern, **run it against the
defect it forbids AND against the nearest legitimate spelling of the same thing**, in the file it guards.
Reading a regex is not testing a regex, and a guard's author is the worst-placed person to imagine the
spelling they did not think of.

### D-11.3.8 — removing a compensation without supplying what it compensated for. Twice, on one axis.

**P3, reclassified: this is a regression the story introduces, not a finding.** `carriedFaceKeys`
(`App.tsx:293`) collects only `entry.assetKey` and never the **variant** asset keys this change added. So
for a document declaring `{"asset": K1, "bold": K2}`, the engine resolves and emits
`fragment.assetKey = K2`, `K2` was never fetched, `carriedFaces.has(K2)` is false, and the fragment gets
**no `fontFamily` at all** — falling to the default stack. **Before this story it at least got
`font-weight: 700`.** Removing the synthetic weight without registering the variant keys makes embedded
bold **strictly worse**.

**Same shape as D-11.3.1, on the same axis, two stories running.** 11.2 made the engine truthful and left
the canvas double-bolding; 11.3 removes the faking and leaves the embedded arm with no face at all. **A
compensation is invisible while it is load-bearing.** The rule D-11.3.1 stated — *a story that makes one
half of a mirrored pair truthful should ask what the other half was compensating for* — now has its second
instance, and both were found by review rather than by the story that caused them.

### D-11.3.9 — nothing that runs parses the shipped starter

**P7, and it is the sharpest finding of the story.** `starter.folio`'s only real-bytes reader is
`folio-go/wasm/cmd/engine/main_test.go:229`, which is `//go:build js && wasm` — **so it is not in
`go list ./...`, `go test ./...` never compiles it, and CI never runs a js/wasm test.**

**Ship the starter with `"version": "1.0"` and object entries, or typo `"Roboto-Bold"`, and every gate
stays green.** The shipped template — the first document every new user opens, and the one artifact
D-11.0.1 added to this story to make bold reachable — is verified by nothing.

**A Go test that reads the real file is now mandatory rather than preferred.** This was my item C, and the
measurement gave it teeth I did not have when I asked for it: I wanted the round-trip promoted from a
measurement to a test; the actual position is that there was no test to promote it *from*, and no gate
would have noticed.

### D-11.3.10 — the rest of 11.3's triage, and what the audit caught this time

**Twelve patches approved.** Beyond P1/P2, P3 and P7 above: **P4** — the headline `bolded` fixture asserts
`face: 'Roboto Bold'` while its chain declares `boldItalic`, so **the story's central test asserts a
resolution the engine cannot produce**. **P5 — the Matrix Test Audit failed again**, second consecutive
story: the row *"bold element, no variant declared → painted in Regular"* has **no test asserting the
fragment**; every absent-cut test sets `bold: false` and inspects only the control. **P6** — a deliberate
`every` is unpinned; swapping it to `some` reds nothing. **P8** — `folio-go/fonts/roboto/NOTICE.md:29` and
`fonts.go:138-140` still quote the pre-11.3 starter chain; fixed in-story, because *the reason they were
left is precisely the reason they would not otherwise be found.* **P10** — `field[:1]` panics on a json tag
spelled exactly `header`, and an unsorted-vs-sorted comparison makes a second failure order-dependent.
**P11/P12** — D-000.30's correction pushed to the source, and the `.property-toggle-unavailable` CSS
recorded as a **stated limit** rather than hand-computed specificity in a report nobody re-reads, plus a
layout check the hand-computation could not cover.

**Five rejections, all correct** — including the starter's non-canonical `fonts` block, rejected with the
`build-wasm.mjs:71-75` citation that says the whole file is deliberately non-canonical. *A rejection that
cites why saves the next reviewer from re-raising it.*

**One deferral:** `TableColumnsProjection`'s four new members have no consumer — DW-240's read-back half
landing ahead of its control. Registered explicitly **so it does not read as a contradiction of the spec's
own D-7.4.5 note**, with the split recorded as deliberate and an owner named.

**And two process results worth keeping.** The verification-gap layer ran **alone** per D-11.2.7, and the
builder **proved the tree survived it** — snapshotting all 17 source files before and `cmp`-ing every one
after, same manifest hash. That is the rule working and being *shown* to work. Q3's open check is
**discharged by measurement**: the third state derives from the chain's variant strings and never from the
committed boolean, so the zero-value collapse is not lossy where it matters. **Q3 does not flip.**

### D-000.31 — 21% of the register is invisible to its own index, and two remaining stories are owed work by it

**Measured while 11.3's implementer worked, because this has now been deferred twice and it bears directly
on finishing Epics 13–15.**

```
### DW-        numbered entries visible to a census   247
- source_spec: raw blocks with NO number                67    (21% of the register)
```

**Sixty-seven blocks, from fifteen-plus stories across Epics 12, 15, 16 and 17** — Story 12.3 alone
contributed ten. They are in the file; what they lack is a **number**, which means nothing can cite them,
no `### DW-` census counts them, and a builder grepping for work owed to its own story does not find them.
This is the same defect the 11.1 and 11.2 closers each caught locally, seen at register scale: **the
register has been under-reporting itself by more than a fifth for most of this run.**

**The concrete risk, which is why this stopped being a tidiness item.** Extracting owners from all 67:

| Orphan | Owner named inside it | Severity |
|---|---|---|
| `deferred-work.md:8483` — `preview/pdf-viewer.tsx:89`'s effect re-runs on every view-state write, disposing and re-rasterizing the `PDFDocumentProxy` | **Story 13.2** | MEDIUM |
| `deferred-work.md:8778` — `TestShippedFacesReproduceFromUpstream` is reachable from no CI gate; its sole invocation is `Makefile:39,45` | **Story 15.2** | MEDIUM |
| `deferred-work.md:8691` — `parseMillipoints` tests the overflow bound *before* the multiply | *"whoever next changes `parseMillipoints`' arithmetic"* | LOW |

**Two of the three are owed to stories still in front of me**, and neither is findable by any search a
builder would run.

**And reading them corrected one of my own attributions.** My extraction pattern matched `Story 15.3` in
the `:8691` block — but the sentence is *"**Not Story 15.3** — it is not an exported-surface question and
it needs no decision before the tag."* **The pattern matched the story that was explicitly excluded.** That
is D-11.1.2's error class in a new costume: matching on shape without reading the subject, and here the
shape was a *negation*. A grep that finds "Story 15.3" cannot tell an assignment from a disclaimer.

**The `:8483` entry also falsifies a premise Epic 13 states as fact**, which makes it worth more than its
severity: Epic 13's prose says `.pdf-preview-scroll` has no height cap so `scrollTop`/`scrollLeft` *"can
never be non-zero"* and the restore effect is dead code. **True only on the vertical axis** — `App.css:368`
sets `overflow: auto` on both, and at a zoom above fit the canvas is wider than its box. **Epic 13's
dispatch must carry that correction**, or 13.2 will be planned against a false premise the epic asserts.

**Actions, and deliberately not taken now.** I did **not** edit the register: 11.3's builder writes its own
deferrals into that file at step-05, and racing it is how two writers produce one lost entry. Instead:
1. **11.3's closer numbers all three orphans**, alongside 11.3's own deferrals — the job both previous
   closers did well, and they are the agent already holding that context.
2. **The remaining 64 are a sweep**, and a sweep is a story, not a chore squeezed into a close. Registered
   as such rather than attempted piecemeal.
3. **Epic 13's dispatch carries the `:8483` correction and the `:8778` pointer**, from this entry, so
   neither depends on the sweep landing first.

**The rule this yields, and it is the fourth register-integrity finding of the run:** *a register is only
as good as the search a reader will actually run.* Content that is present but unindexed is, for every
practical purpose, absent — and it is worse than absent, because its presence in the file makes the
register look complete to anyone who scrolls it.

### D-000.31a — CORRECTION: D-000.31's central table is wrong, and my own search was the defective one

**D-000.31 is retracted on its central claim.** It reported *"67 raw blocks with NO number — 21% of the
register invisible to its own index"* and named three orphans owed to Stories 13.2, 15.2 and a
`parseMillipoints` owner. **All three are numbered entries that already exist, with those exact owners
already written in.** Caught by 11.3's closer; **verified by me before accepting it**, which is the
standard I have been holding every agent in this run to.

| I cited | What is actually there |
|---|---|
| `:8483` pdf-viewer effect churn | **`### DW-191`**, heading two lines above — Owner Story 13.2, MEDIUM, OPEN |
| `:8691` `parseMillipoints` overflow bound | **`### DW-197`** — LOW, and it already says *"Not Story 15.3"* |
| `:8778` `TestShippedFacesReproduceFromUpstream` | **`### DW-200`** — Owner Story 15.2, MEDIUM |

**The mechanism of my error.** `- source_spec:` is the **provenance metadata field of a numbered entry**,
sitting one or two lines under its own `### DW-` heading. I read it as the marker of an *unnumbered* block.
Measured properly — nearest `### DW-` heading above each occurrence:

```
metadata lines (within 3 lines of a DW heading):  58
genuine standalone raw blocks:                    10
  all ten inside DW-189 (7: lines 8413-8437) and DW-190 (3: 8467-8475)
```

**So the register holds 258 entries and ~4% of it is unindexed, not 21%.** Two of the three "risks to
remaining stories" never existed: **DW-191 already carries the Epic 13 correction prominently**, under its
own heading *"Why it fires today, which is the part the epic gets wrong."* My conclusion — *Epic 13 will
otherwise be planned against a false premise* — was exactly backwards: **the register had already caught
it and said so.**

**The lesson, and it is the one I have been demanding of everyone else since D-11.2.4.** *My grep had no
positive control.* One check — *does `- source_spec:` also appear inside a known-good numbered entry?* —
would have collapsed the whole finding in seconds. I ran an absence-shaped search, got a large number, and
reported the number instead of testing what it was counting.

**And the irony is load-bearing rather than decorative.** D-000.31's closing rule reads: *"a register is
only as good as the search a reader will actually run."* **The defective search was mine**, in the entry
that coined the rule. A rule about searches, established by a bad search, and neither the writing of it nor
committing it caught the problem — **a closer re-deriving the citation did.** Records are checked by use,
never by review; this run has now paid for that lesson four times.

**What survives D-000.31, and it is not nothing:**
1. **Ten genuine raw blocks remain**, inside DW-189 and DW-190, from the Epic 16 adversarial round and
   Story 17.5. They are real and still unindexed.
2. **A sweep is no longer a story.** Ten blocks under two known headings is a close-sized task, not an
   epic-sized one. **Folded into the next close rather than scheduled.**
3. **Epic 13's dispatch still carries the pdf-viewer correction** — but as a **pointer to DW-191**, which
   any builder finds by grepping its own story id. That is the register working as designed.
4. **The census-invisible `- source_spec:` block form is still a real defect** — it is how 11.1 lost four
   deferrals and how 11.3's would have gone unnumbered without its closer. That finding stands; only its
   scale was wrong.

**DW-246 is created and needs an owner** — `TableColumnsProjection`'s four new members have no consumer,
and **no story in `epics.md` adds a table-header weight control.** Not routable to an existing story;
carried to Epic 11's boundary gate for placement.
