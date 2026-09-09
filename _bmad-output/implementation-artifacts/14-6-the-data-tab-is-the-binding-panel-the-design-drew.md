---
title: 'The DATA tab is the binding panel the design drew'
type: 'feature'
created: '2026-09-09'
status: 'done'
baseline_commit: '3a85a023f07a48dfc916c12b380bebe2c40becae'
review_loop_iteration: 0
context: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative, and rewritten after delivery to describe what actually shipped. The frozen Intent
below is what governs the implementation.*

The DATA tab is where an author connects a piece of their document to a piece of their data. Today it
shows them the wrong thing: instead of the value, each row reads out a type name, a count and a
truncated preview run together into one string, so a customer's name appears as `string · "สมชาย
วงศ์ประเสริฐ"` rather than simply as their name. This story makes each row show the value, marks
objects and lists with `{ }` and `[]`, and puts the loaded file's name and size at the top.

It also stops the panel inviting choices it is going to reject. A bar above the tree now says what is
selected and what choosing a path would do — before the choice, rather than after the engine refuses
it. Lists carry a TABLE ONLY badge and say why a piece of text cannot hold one. Runtime parameters,
which the engine finds in the template itself, appear under `params` marked RUNTIME so the author can
see the namespace exists, while remaining unselectable because the engine will not bind them.

Two paths were being offered that should never have been. One the engine refuses outright. The other
is worse and is the interesting one: an **empty** list was being offered as if it were a single value,
the engine **accepted** it, and the failure only appeared later when the document was actually
rendered. The cause turned out to be an inconsistency in the code that reads the author's sample file
— it withholds the "selectable" marker from objects but not from lists — so the fix is at that source
rather than in the panel that displays it.

Choosing a path now binds it immediately; the separate "connect" button is gone, and the choice is
undoable like any other edit. Anything shown but unselectable is dimmed **and says why**, because the
design document forbids a plain grey-out without a reason. The status bar gains a count of how many
elements carry a binding, and the panel now states in the empty case what it never said anywhere
before: sample data is never written into the template.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The DATA tab shows an author type names instead of their data. Every scalar row reads
`label · path · kind · count · preview · candidate` as one concatenated string, so `customer.name`
announces itself as `string · "สมชาย วงศ์ประเสริฐ" · root scalar candidate` rather than showing the
value. The panel then invites a pick it will refuse — its unconditional note tells *every* author to
"Choose an offered root scalar path, then connect it", including a Line author who is told it was
pointless only after picking (DW-352), and its refusal is generic for a Table, which legally *does*
take a binding through the table editor (DW-353). Two paths are offered that must not be: `params.*`,
which Go refuses outright (DW-349), and an **empty collection**, which Go *accepts* and the render then
fails (DW-350). Nothing anywhere states how much of the document is bound, or that sample data is never
written into the template.

**Approach:** Rebuild the DATA tab as the binding panel `Binding.dc.html` and `DESIGN.md` describe: a
file header with its size, a selection context bar stating what a pick would bind *before* the pick, a
tree whose rows show values, object/collection markers, TABLE ONLY and RUNTIME badges with dimmed
unpickable rows carrying a stated reason, a pick that binds immediately, a bound-element count in the
status bar, and an empty-state note. The model already carries every value this needs; this is
presentation plus two data joins, and it adds no engine surface.

## Boundaries & Constraints

**Always:**
- **A pick binds immediately** — *owner ruling, 2026-09-09*. There is no explicit "Connect selected
  path" control; the mockup's omission of one is a design decision, not a gap. The bind is undoable
  under UX-DR20 like any other edit, and the engine still validates and can still refuse. The picked-row
  treatment — a bind-accent left bar — is adopted either way. Accepted cost, stated by the owner:
  clicking a path while exploring the tree will bind it, recoverable by undo.
- **`DESIGN.md` outranks the mockup.** Where
  `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` and `Binding.dc.html`
  disagree, DESIGN.md wins.
- **A disabled row carries a stated reason** (DESIGN.md:550, restated :585) — *"Disabled nodes drop to
  `0.42` opacity and must carry a stated reason, never a bare grey-out."* Dimming alone is a violation.
- **The 5px amber dot marks a *bindable* leaf** (DESIGN.md:549, :351-353). A row that cannot be picked
  must not carry one.
- **"Bound" means the element carries a `binding` **or** a `tableBind`** — *owner ruling, 2026-09-09*.
  A table bound to a collection **is** bound. Known and deliberate exclusions, which read as surprising
  and are correct: `"Total: {{amount}}"`, `"{{amount}} THB"`, `"{{upper(customer.name)}}"` and
  `"{{page}}"` all count as **unbound**, because `directCanvasBinding` populates `binding` only for a
  whole-value, single, non-reserved path placeholder.
- The panel keeps judging only the **component kind** and the **path's root namespace and shape**. It
  never re-derives the engine's type gate — `SCALAR_BINDING_COMPONENT_TYPES` and its mirror are
  inherited, not re-spelled.

**Ask First:**
- Adding, removing or renaming any entry in `folio-designer/src/design-tokens.ts`.
- Any change to `folio-go/**` — this story is browser-side and adds no engine surface.
- Any new engine round-trip beyond the one lazy `parameter-references` fetch this spec authorises.
- Removing the "Load sample JSON" / "Replace sample JSON" button.

**Never:**
- Never write sample data, or any param value, into the template.
- Never make the panel pre-judge whether a *scalar* path is bindable by its sampled runtime kind.
  D-6.2.1 governs **command legality** and the command stays legal; withholding an observed
  **collection** is explicitly contemplated by that ruling's own test and is in scope. Growing that into
  "is this path bindable" for scalars is not.
- Never touch `fixtures/declared-variants/expected.pdf`, `input.folio`, or `signoff.json`.
- Never add a hex, `rgb()` or `hsl()` literal to `App.css`, and never add a second `@media` query.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Scalar row | `{"customer":{"name":"สมชาย วงศ์ประเสริฐ"}}` | Row shows label `name` and value `สมชาย วงศ์ประเสริฐ` right-aligned; amber dot; pickable | N/A |
| Object row | `{"customer":{...}}` | Marked `{ }` beside the name; no value; no dot; not pickable | N/A |
| Populated collection | `{"transactions":[{...},{...}]}` | Label `transactions[]`, TABLE ONLY badge, reason *"Collection · 2 items. Text cannot bind a collection."*; row **not pickable**; children dimmed at 0.42, unpickable, reason stated | N/A |
| **Empty collection (DW-350)** | `{"items":[]}` | Label `items[]`, TABLE ONLY badge, reason line, **not pickable** — no `onPick`, no bind command | Previously picked, was accepted by Go, failed only at render |
| Sample-JSON `params` (DW-349) | `{"params":{"reportDate":"…"}}` | Every node under it dimmed, no dot, **not pickable**, reason stated | Previously picked, refused by Go |
| Engine `params` (AC6) | `parameter-references` → `["reportDate","branchName"]` | A `params` section with RUNTIME badge; each name with its current value, or a muted "not set"; visible, never pickable | On `failed`, say the engine could not provide references; never guess an empty list |
| Selection context | Text selected | Context bar: *"Text selected · binding to string"* in the bind accent | N/A |
| Selection context, Line | Line selected | Bar states a Line cannot take a scalar binding **before** any pick | N/A |
| Selection context, Table (DW-353) | Table selected | Bar names the table editor as where a table's collection is bound | N/A |
| File header | 18,432-byte `sample-statement.json` | `sample-statement.json` + size, size in mono | N/A |
| Status bar | 7 elements with `binding` or `tableBind`, 10 elements total | *"7 of 10 elements bound"* | 0 elements → the span is not rendered at all |
| Empty state | No sample loaded | Says so, and states sample data is never written into the template | N/A |

</frozen-after-approval>

## Code Map

Investigated at `2d9a97a`. Anchors verified by execution or by reading the cited line, not by recall.

**The file that changes most**
- `folio-designer/src/DataPanel.tsx` (147 lines) — the whole panel. `:83-93` the five-arm `unavailable`
  ladder (DW-352/353 live here; 14.4's own comment at `:61-70` says this arm's *wording* is 14.6's to
  replace and only **the rule** is inherited). `:94-107` the panel body. `:130` the single tree-row JSX
  expression — one line, every row treatment is in it. `:143` `describe()`, the concatenation AC1
  replaces. `:109-131` `DataTree`, including the roving tab stop and arrow/Home/End/Enter navigation.
- ⚠ **`:18-22` over-reaches and must be amended by this story.** It forbids the panel judging a path's
  runtime kind, citing D-6.2.1 — but `folio-go/component_commands_test.go:279-281` says in its own
  words: *"This is legal command grammar even though **a picker would withhold an observed
  collection**."* D-6.2.1 governs command legality. Withholding a collection is contemplated by it.

**The model — already carries everything AC1/AC2/AC3/AC5 need** (measured by executing
`acceptSampleData`, not by reading)
- `folio-designer/src/sample-data.ts:6-18` `SampleNode` — `kind`, `path`, `label`, `preview`, `count`,
  `children`, `truncated`, `segments`. `:19` `SampleData` carries `bytes: ArrayBuffer` → AC3's size is
  `bytes.byteLength`.
- `:98,:107` **already append `[]` to a collection's label** — AC2's `[]` marker ships today. Only `{ }`
  is new.
- ⚠ `:98,:107` also attach `segments` to a **collection**, while `:77-92` `object()` never does. That
  asymmetry is DW-350's whole mechanism. Measured: `{"items":[]}` → `{kind:'collection', count:0,
  children:0, segments:["items"]}`; `DataPanel.tsx:130`'s `children.length ? toggle : segments &&
  onPick` therefore **picks** it. A populated `transactions[]` also carries `segments` and escapes only
  because it has children — an accident, not a rule.
- ⚠ `:103` passes `rootScoped: false`, so **row-scope children already carry no `segments` and are
  already unpickable.** AC5's "unpickable" is preservation for them; only the dimming and the reason
  are new. The pickable thing is the collection's own row.

**Host wiring**
- `folio-designer/src/App.tsx:2569` — `<footer className="status-bar">`. AC8 goes here, beside the exact
  precedent `{canvas.fontFamilies.length} font(s) in template`. `canvas` already in scope.
- `App.tsx:274` `parameterReferenceState`; `:790-808` `loadParameterReferences()`.
  ⚠ Called at `:813` and refreshed at `:1943`, `:1986`, `:2000` — **all four gated on
  `modeRef.current === 'preview'`.** Without a design-mode fetch the namespace stays invisible until
  Preview, which is exactly what AC6 exists to prevent.
- `App.tsx:271-272` `previewParams` / `previewParamsDraft` — the only "current values" that exist;
  `'{}'` until the author types in Preview. `:2584-2588` `parameterValues` extracts them as raw JSON
  token slices.
- `App.tsx:2980` — **`.binding-chip` + `.binding-dot` already implement DESIGN.md:545.** Reuse for the
  context bar and the row dot; do not invent.

**Engine facts (read-only — no Go changes in this story)**
- `folio-go/component_commands.go:749-751` — `params is not a root data binding`, field
  `binding.segments`. `:775-777` — `only text components can receive a scalar binding`. `:782` installs
  `"{{" + path + "}}"`.
- `folio-go/page_setup.go:1767` + `directCanvasBinding:1810-1823` — `binding` is populated **only** for
  a text element whose entire value is one non-reserved single-path placeholder. `:1798-1803` —
  `tableBind` is a *different* field.
- `folio-go/page_setup.go:1747-1808` `canvasComponents` — the denominator: every element in
  `pageHeader` + `content` + `pageFooter`, flat, once each regardless of page count.
- `folio-go/parameter_references.go:70-77` — projects **only the first segment after `params`**
  (`params.statement.reportDate` → `statement`); `:36-38` deliberately excludes table columns and
  footers. `MaxParameterReferenceNameLength = 128` at `:16`.
- `folio-go/internal/bind/text.go:415-423` — where an empty-collection bind fails, at render only.

**Hard contracts this story must not red** (all pre-registered as firing on an Epic 14 restyle)
- ⚠ `folio-designer/src/engine-bounds-mirror.test.ts:768-769` pins **two exact source lines** in
  `DataPanel.tsx`: the import line
  `import { SCALAR_BINDING_COMPONENT_TYPES, type CanvasComponentType } from './engine-protocol'`
  (that exact spelling, on one line) and
  `const bindableKind = selectedComponentType !== undefined && SCALAR_BINDING_COMPONENT_TYPES.includes(selectedComponentType)`.
  `:774` additionally asserts `DataPanel.tsx` contains **no `=== 'text'` or `!== 'text'` anywhere**,
  comments and string literals included.
- ⚠ `engine-bounds-mirror.test.ts:715-718` — *"Nothing here may grow into 'is this path bindable'"*.
  The new collection rule is asserted in `DataPanel.test.tsx`, **not** in the mirror.
- `design-contract.test.ts` — exact per-group name-set equality against `design-tokens.ts` (so **adding**
  a token reds it) and no hex/rgb/hsl literal in `App.css`.
- ⚠ **The media-query guard is NOT in `design-contract.test.ts`.** `epic-11-14-decision-log.md` places it
  there; measured, it is `canvas-authority-contract.test.ts:344`, and it is stricter than "exactly one":
  it asserts the exact **condition list** of `App.css`'s `@media` rules equals
  `['prefers-reduced-motion: reduce']`. Also note `grep -c "@media" App.css` returns **3** — two hits are
  prose inside comments (`:84`, `:751`); only `:789` is a rule. Count rules, not raw text.
- `control-vocabulary-contract.test.tsx:309-353` — **six** sweep states, none of which opens the DATA
  tab; population is `button, [role="button"]` only. Neither DATA-tab button is in the V2 census.

**Existing CSS to restyle** (`folio-designer/src/App.css`) — `:364-378` and `:373-376`. Every colour is
already a token; the mockup needs no new ones.
⚠ `:371` `.data-tree-picked` uses `--color-select` (**cyan**) — a DESIGN.md:596 violation
(*"Don't use amber for selection or cyan for data"*). AC7's bind-accent bar corrects it.

**Test pins that must be rewritten with the wording they pin**
- `DataPanel.test.tsx` (369 lines, 14 `it()`): `:46,:204,:313,:327` "Binding unavailable…"; `:50`
  "Local sample:"; `:262,:296,:314,:328,:345,:358,:367` "Only text components…"; `:366` "Choose an
  offered root scalar path."; `:136,:163,:184,:264,:297,:315` the Connect button.
- ⚠ `App.test.tsx:5955` and `:9027` also click **"Connect selected path"** — removing the button
  reaches beyond `DataPanel.test.tsx`.

**Design authority**
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` — `:549-551` Tree node,
  `:545-547` Binding chip, `:376` `{colors.dim}` is *"the faintest usable outline (a badge)"*, `:331-334`
  mono for machine values **including a byte count**, `:433` uppercase only for section labels, `:585`
  state the reason next to anything disabled.
- `…/mockups/Binding.dc.html` — every colour in it maps to a declared token: `#C9A758` bind, `#E0C07A`
  bind-text, `#8A7440` bind-muted, `#241F14` bind-tint, `#3A3118` bind-edge-soft, `#363D46` dim,
  `#737C86` ink-low, `#5E666F` ink-faint, `#4E565F` ink-ghost, `#AAB2BB` ink, `#E6E9EC` ink-high.

## Tasks & Acceptance

**Operational constraints for the implementer** (ratified channel, D-14.4.1):
- **Never** run `git commit`, `git add`, `git stash`, `git checkout`, `git reset`, `git revert`,
  `git restore`, `git clean`, `git push`, or create a branch. Reading git state is fine.
- **Do not edit** `sprint-status.yaml`, `deferred-work.md`, `epics.md`, or `DESIGN.md`.
- Shell is **zsh**: `${PIPESTATUS[0]}` is empty. Use `cmd > log 2>&1; echo $?`; never `$?` after a pipe.
  Quote every glob and every variable. `App.tsx` contains **exactly 2 NUL bytes, the first at byte
  294361, line 3997** (measured, not recalled) — plain `diff` prints "Binary files differ" with zero
  changed lines. Use `diff -a` and `cmp`.
- Sanity-check the magnitude of every number before acting on it.
- **D-14.6.1 (recorded 2026-09-09, adopted run-wide): `find` is non-deterministic here — a `find` miss
  is NOT evidence of absence.** It is a shell function wrapping `bfs`; the identical command returned
  empty once and the file minutes later. **Use `git ls-files` for anything tracked**
  (`git ls-files | awk '/NAME/'`), and pair every negative search with a positive control. Never report
  a file or symbol as absent on a single `find`.
- ⚠ **`DESIGN.md` is machine-read, so it is an authority you cannot quietly diverge from.**
  `folio-designer/src/design-contract.test.ts:8` resolves
  `../../_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md` and asserts
  (`:20-23`) exact token-**name** set equality against `design-tokens.ts`, and (`:25-31`) that every
  colour **value** declared there appears in `tokens.css` as `--color-<name>: <value>`. Adding a token
  would require editing DESIGN.md, which is forbidden — hence "Ask First". Every colour this story
  needs already exists.

**Execution:**
- [x] `folio-designer/src/sample-data.ts` -- **leave `segments` on collection nodes exactly as they are.**
      They are correct data with a second reader: `tableSampleCandidates` (`App.tsx:62`) is gated on
      `kind === 'collection' && segments?.length` and feeds the Table Editor's datalists. Record why in a
      comment so the next reader does not re-derive the removal.
- [x] `folio-designer/src/DataPanel.tsx` -- replace `describe()` and the `:130` row expression with the
      designed row: label, value right-aligned, `{ }`/`[]` marker, a 5px bind dot **only on a bindable
      leaf**, TABLE ONLY and RUNTIME badges as plain `<span>`s inside the treeitem button -- so a badge
      is announced as part of its row and never becomes a separate control.
- [x] `folio-designer/src/DataPanel.tsx` -- make a node unpickable when its kind is `collection` **or**
      its first segment is `params`, from any source; dim it to `0.42` and give it a stated reason --
      DESIGN.md:550 forbids a bare grey-out. Preserve `bindableKind` and the import line verbatim; add
      no `=== 'text'` comparison anywhere in the file, comments included.
- [x] `folio-designer/src/DataPanel.tsx` -- replace the `unavailable` ladder's wording with the file
      header, the selection context bar and the empty-state note; name the table editor for a Table
      (DW-353) and state a Line's refusal before any pick (DW-352). Amend the `:18-22` comment, which
      currently forbids what D-6.2.1's own test contemplates.
- [x] `folio-designer/src/DataPanel.tsx` -- remove the Connect button; a pick calls `onConnect`
      directly (owner ruling). Keep the Load/Replace sample JSON button.
- [x] `folio-designer/src/App.tsx` -- render the engine-discovered `params` section from
      `parameterReferenceState` + `previewParams`, and fetch references **lazily and idempotently** in
      design mode: on document load or first DATA-tab render only, never per keystroke, selection change
      or re-render.
- [x] `folio-designer/src/App.tsx:2569` -- add the bound-element count to the status bar; suppress the
      span entirely when the document has zero elements.
- [x] `folio-designer/src/App.css` -- restyle `:364-378` to the design; correct `.data-tree-picked` from
      cyan to the bind accent. Tokens only — no hex literal, no new `@media`.
- [x] `folio-designer/src/DataPanel.test.tsx`, `folio-designer/src/App.test.tsx` -- rewrite the pins
      listed in the Code Map, and add the fences below.

**Acceptance Criteria:**
- Given a scalar path, when the tree is shown, then its value appears beside the label, right-aligned,
  and the row no longer renders a `kind · count · preview · candidate` string.
- Given an object node, when it is shown, then it is marked `{ }`; given a collection, `[]`.
- Given a loaded file, when the header is shown, then it names the file and its size, size in mono.
- Given a selected component, when the DATA tab is shown, then a context bar states what is selected and
  what a pick would bind, before any pick — a Line's refusal is stated up front (DW-352) and a Table is
  told where its collection is edited (DW-353).
- Given a collection, when it is shown, then it carries a TABLE ONLY badge and a stated reason, **its own
  row is not pickable**, and its row-scope children are dimmed with a stated reason.
- **Given an empty collection `{"items":[]}`, when its row is clicked or Enter'd, then no bind command is
  sent** — DW-350, the one case the engine accepts and the render then fails.
- Given the engine's discovered parameters, when the tree is shown, then they appear under `params` with
  a RUNTIME badge and their current values or a muted "not set", visible and never pickable — and given a
  `params` key in the author's own sample JSON, its nodes are equally unpickable (DW-349). *These are two
  different sources; neither may be satisfied by the other.*
- Given a pick, when the author clicks a bindable leaf, then the bind is dispatched immediately with no
  intermediate control, and the picked row carries the bind-accent left bar.
- Given the document's components, when the status bar is shown, then it states how many carry a
  `binding` or a `tableBind`, out of every element in all three bands — and renders nothing at zero.
- Given no sample data, when the panel is shown, then it says so and states that sample data is never
  written into the template.
- Given every value, badge and dimmed row this story adds, when read by assistive technology or operated
  by keyboard, then the roving tab stop and arrow/Home/End/Enter navigation are **preserved** (they
  already exist — this is a preservation claim, not construction), a badge is announced as part of its
  row rather than as a separate control, and an unpickable row reports that it is unpickable.

**Fence obligations — this story is dense with absence claims and they are where this epic's defects
have been.** Every claim below is an *absence* or *preservation* claim. Per **D-14.4.3**, a red proof
that only REVERTS cannot falsify one — reverting removes the thing and the claim passes more easily. Per
**D-14.5.1**, the ADD mutation must be applied **at every position the forbidden thing could occupy** —
on the container, on the subject, and on any wrapper between — and the fence shown to red at each.
- "never pickable" (`params`, from both sources) — ADD `segments` to a params node; ADD an `onPick` call.
- "the collection row is not pickable" — ADD `segments` back in `sample-data.ts`; separately ADD a pick
  handler on the collection row.
- "no connect control" — ADD a button to the panel, to the tree, and to the row.
- "sample data is never written into the template" — ADD a write and show the fence reds.
- "a badge is announced as part of its row rather than as a separate control" — ADD `role="button"` and
  ADD an `aria-label` to the badge span, **and to its wrapper**. ⚠ Testing Library role queries exclude
  `aria-hidden` subtrees by default: assert **both** the default spelling and `{ hidden: true }`, and
  sweep `[root, ...root.querySelectorAll('*')]` so the container itself is included — a `within(x)`
  sweep cannot see a violation **on** `x`.
- "the roving tab stop is preserved" — mutate the tabIndex logic and show the fence reds.
- Do not pin behaviour by reading source text where you can exercise it (DW-362).
- State explicitly, for each fence, which assertion redded under which mutation. A fence can red for the
  wrong reason, or a neighbouring assertion can carry it while the one you named is inert.

## Spec Change Log

*Append-only. Populated by step-04 during review loops.*

### 2026-09-10 — amendment in place of a `bad_spec` loopback (coordinator's ruling)

**Triggering finding.** Step-04's verification-gap layer found that this spec's own Execution task —
"stop attaching `segments` to a collection node" — emptied `tableSampleCandidates` (`App.tsx:62`) for
every template, killing the Table Editor's "Root collection" and "Row field" datalists. Confirmed by
executing the parser, not by reading it. **The same story's new context bar had just started directing
table authors into that editor** (DW-353's discharge), so the change increased traffic to the surface it
broke. The implementation also *pinned* the regression with `expect(...segments).toBeUndefined()`, and
the whole suite was green because nothing anywhere read a `<datalist>` option.

**Root cause, and whose.** The builder proposed the parser fix and **the coordinator endorsed it**, on
the argument that "fixing the source of the wrong data beats fixing each reader of it" — a good rule
whose precondition is that you know who the readers are. Neither party enumerated them. The
coordinator's *original* ruling, a panel-side clause, was correct and was talked out of. Recorded as the
coordinator's error at the coordinator's instruction.

**What was amended.** The Execution task above (parser change → leave `segments` alone, refuse by kind
in `DataPanel.tsx`'s `rowFor`) and the Design Note explaining it. **The `<frozen-after-approval>` block
is untouched and was correct throughout**; the wrong text was three lines in non-frozen sections.

**Why an amendment rather than the workflow's revert-and-re-derive.** `bad_spec`'s remedy exists to stop
code being derived from wrong intent. The intent was right; the fix restores a ruling already given.
Reverting ~680 lines would have discarded six executed mutation proofs to re-derive otherwise sound code
and re-run the risk of new defects. The coordinator is the human that loopback would have escalated to,
and ruled to amend. `review_loop_iteration` stays **0**.

**Known-bad state avoided.** A silently dead discovery surface behind a control the same story started
advertising, protected by an assertion that pinned the deadness.

**KEEP — must survive any later re-derivation.**
- `rowFor` refuses `collection` and `inCollection` **by kind**, never by the absence of `segments`.
- The six fences I red-proved by executed mutation, namely: the collection row refusing *with* a `segments` marker
  present; the table-editor datalist fence (red-proved by re-removing `segments`); the AC6 end-to-end
  assertion that renders the real `<App>` and finds `reportDate` in the DOM (red-proved by dropping the
  prop, while `tsc` stays green); and M4 redding on the scan rather than its own witness.
- `aria-disabled` marks only rows that can be neither picked **nor expanded**.
- The bind accent is applied only when a command is actually dispatched.

## Design Notes

**The mockup loses twice, to a normative source, and that is now a pattern worth naming.** DESIGN.md
outranks `Binding.dc.html`, and this is the second time in Epic 14 (Story 14.5's AC3 was the first).
1. The mockup draws `params` children with the **bind dot** and no dimming — i.e. as pickable. DESIGN.md
   :549 reserves that dot for a *bindable* leaf, and :603 forbids drawing "an affordance the product
   cannot honour". Go refuses `params` at `component_commands.go:749-751`. AC6 wins; reuse the row-scope
   dimming.
2. The mockup dims `transactions[]`'s children with **no stated reason**. DESIGN.md:550 and :585 both
   forbid a bare grey-out. Every dimmed row gets a reason.
3. The mockup's DATA tab shows **no Load button**. Its file row is a header, not a replacement for the
   control — removing the button would strand the author with no way to load data, and no AC asks for it.
   Recorded as mockup incompleteness.
4. The mockup renders the file size in sans; DESIGN.md:332 names "a byte count" among the values set in
   **mono**. DESIGN.md wins.

**AC7 fixes a defect; it does not adopt a preference.** `App.css:371` `.data-tree-picked` currently
uses `--color-select` — **cyan** — for the picked row in a *data* tree. DESIGN.md:596 is explicit:
*"Don't use amber for selection or cyan for data — it breaks the one rule users learn."* A picked path
is data. So the shipped picked-row treatment is a **live DESIGN.md violation**, and AC7's bind-accent
left bar corrects it. Recorded here and to be repeated in the Delivery Log, because a later reader will
otherwise see this as a gratuitous restyle.

**The mockup loses a THIRD time, and this one is deliberate separation.** The mockup badges the
`params` root RUNTIME. Here only the **engine-discovered** section carries that badge; a `params` key in
the author's own sample JSON is dimmed with a stated reason and no badge. The two sources are different
data (`parameter-references` names vs. a parsed JSON subtree) and the owner required that neither be
satisfiable by the other — a shared badge would let one test cover both and hide the loss of either.

**Rows are deliberately NOT dimmed for a selection-kind refusal.** When the selected component cannot
take a scalar binding, the context bar states the refusal and the tree stays at full contrast. Dimming
every row whenever nothing bindable is selected would grey out the entire panel on load — hostile, and
it would also overload the 0.42 treatment, which means "this row can never be picked" rather than "not
right now". The refusal is stated once, where DESIGN.md:585 asks for it.

**The file size is AC3's format, in DESIGN.md's face, using the product's existing formatter.** An
earlier spelling printed an exact `18,432 bytes`, citing DESIGN.md as outranking the mockup. It does —
but `:331-334` rules on the **typeface** ("is set in mono"), not on rounding, and is silent on format;
`:418` calls the load screen's size display "the megabyte count", so where the design does speak about a
size it speaks in rounded units. AC3 names the rendering outright (`18 KB`) and the mockup agrees, so
there was nothing for DESIGN.md to overrule. It now reuses `formatRenderSize`
(`preview/evidence-rail-facts.ts`), which picks the unit **after** rounding — the documented reason it
never prints `1024 KB` — keeping one rounding rule in the product rather than two that can disagree.

**Why DW-350 is closed in the panel and NOT in the parser.** `object()` withholds `segments` and
`array()` does not, and that asymmetry looks like the root defect. It is not. A collection's `segments`
are **correct data with a second reader** — `tableSampleCandidates` (`App.tsx:62`) is gated on
`kind === 'collection' && segments?.length` and feeds the Table Editor's "Root collection" and "Row
field" datalists. Only the *panel's* reading of "has segments" as "is a scalar candidate" was wrong, so
`rowFor` refuses by **kind**. That is a presentation choice about what to offer; command legality stays
exactly where D-6.2.1 put it, and `component_commands_test.go:279-281` anticipates a picker that
withholds an observed collection.

**Why the bound count needed a ruling.** `binding` is far narrower than "has a placeholder". Measured:
the golden `worked-example.json` would read *"0 of 3 elements bound"* and the shipped `starter.folio`
*"0 of 0"*. Counting `tableBind` too is what makes the number mean what an author reads it to mean.

## Verification

**Commands** (the per-story cadence, D-000.33 — these four and no others):
- `cd folio-designer && npx vitest run` -- expected: **≥ 1263 tests, 0 failures**. Baseline at `2d9a97a`
  is 75 files / 1263 tests / 0 failures. **Diff test names as a MULTISET**, not by total.
  *(HEAD moved to `3a85a02` during planning. The two intervening commits touch only `deferred-work.md`
  and `epic-11-14-decision-log.md` — no source file — so this baseline still holds at `3a85a02`.)*
- `cd folio-designer && npx tsc -b --force` -- expected: exit 0. `--force` is mandatory.
- `cd folio-designer && npx oxlint` -- expected: exit 0, **exactly 4** `only-export-components`
  warnings, 0 errors. Re-measure; do not quote the baseline.
- `cd folio-designer && npm run test:e2e:compile` -- expected: exit 0.

**Per D-000.32, name the suites that did not run, in those words:** the browser suite, the Go suites,
the matrix legs, `npm run build` as a gate, the `verify:offline*` chain, and the font-host scans **did
not run** — this story changes no Go, no bundled asset and no font surface.

**Manual checks:**
- Confirm no hex/`rgb()`/`hsl()` literal was added to `App.css`, and that `App.css`'s `@media`
  **rule** condition list is still exactly `['prefers-reduced-motion: reduce']` (count rules, not raw
  `@media` text — two mentions live in comments).
- Confirm `design-tokens.ts` is byte-identical to its state at `2d9a97a`.
- Confirm `DataPanel.tsx` still contains the two exact lines pinned by
  `engine-bounds-mirror.test.ts:768-769`, and contains no `=== 'text'` / `!== 'text'`.

## Suggested Review Order

**The gesture and what the panel will offer** — start here; every other stop follows from this rule.

- The one place DW-350 is closed: refuse by KIND, never by a missing `segments` marker.
  [`DataPanel.tsx:173`](../../folio-designer/src/DataPanel.tsx#L173)

- Why the parser keeps `segments` on a collection, and who the second reader is.
  [`sample-data.ts:18`](../../folio-designer/src/sample-data.ts#L18)

- The second reader itself: gated on `kind === 'collection' && segments?.length`.
  [`App.tsx:62`](../../folio-designer/src/App.tsx#L62)

- A pick binds immediately, and the accent marks a dispatch rather than a gesture.
  [`DataPanel.tsx:77`](../../folio-designer/src/DataPanel.tsx#L77)

**What the panel says before the pick**

- The refusal ladder rewritten as a context bar; DW-352 and DW-353 both discharge here.
  [`DataPanel.tsx:62`](../../folio-designer/src/DataPanel.tsx#L62)

- The row: dot only when bindable, badges as plain spans, `aria-disabled` only when inoperable.
  [`DataPanel.tsx:159`](../../folio-designer/src/DataPanel.tsx#L159)

- AC6's namespace: display only, no control of any kind, never pickable.
  [`DataPanel.tsx:114`](../../folio-designer/src/DataPanel.tsx#L114)

- AC3's units via the product's existing formatter, so one rounding rule exists.
  [`DataPanel.tsx:96`](../../folio-designer/src/DataPanel.tsx#L96)

**Host wiring**

- The one design-mode engine round-trip: lazy, idempotent, re-armed on failure.
  [`App.tsx:825`](../../folio-designer/src/App.tsx#L825)

- AC8's count — `binding` OR `tableBind`, suppressed entirely at zero elements.
  [`App.tsx:2611`](../../folio-designer/src/App.tsx#L2611)

- Must stay below `.binding-chip`: equal specificity, source order decides.
  [`App.css:399`](../../folio-designer/src/App.css#L399)

**Fences worth reading before the code they guard**

- The regression fence that was missing; red-proved by re-removing `segments`.
  [`App.test.tsx:266`](../../folio-designer/src/App.test.tsx#L266)

- AC6 end-to-end through the real `<App>`; red-proved by dropping the prop while `tsc` stayed green.
  [`DataPanel.test.tsx:737`](../../folio-designer/src/DataPanel.test.tsx#L737)

- DW-350 itself: no command for an empty collection, by click or by Enter.
  [`DataPanel.test.tsx:454`](../../folio-designer/src/DataPanel.test.tsx#L454)

- Sample data never reaches the template; reds on the scan, not on its own witness.
  [`DataPanel.test.tsx:697`](../../folio-designer/src/DataPanel.test.tsx#L697)
