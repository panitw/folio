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
