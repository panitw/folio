---
title: 'Story 16.5: Installing a face is not embedding it'
type: 'feature'
created: '2026-09-03'
status: 'done'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: 'e80f6073d9d2ab08f7c51bb41463aca97b379979'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-16-decision-log.md'
warnings: []
deferred:
  - summary: 'README.md:147 claims "CI compiles the suite, gates run it" for the e2e suite; false at HEAD — ci.yml runs test:e2e:compile only and `playwright test` appears 0 times across the 93 gate-shaped files searched.'
    evidence: 'Fourth instance in this epic of documentation asserting a gate that does not run. Registered, deliberately not fixed here (orchestrator ruling).'
  - summary: 'The nameID-13 licence-signature tie stays Go-only, so an installed face can still be refused at first use. Port or share it if a mechanism ever exists that does not create a competing authority over what enters a document.'
    evidence: 'Accepted knowingly under D-16.R.46 Q4 as amended: the residue is per-face and needs ot.ParseFont/ot.ParseName plus two regex tables. Disclosed to the author at the refusal instead.'
  - summary: 'MEASURED RED, PRE-EXISTING AND NOT FIXED HERE: `e2e/font-embed-boundary.spec.ts` "the designer offers the whole local face tier" FAILS IN A REAL BROWSER at its disclosure assertion (`options.find(text => text.includes("families you can add"))` is `undefined`). The sentence IS rendered; the spec reads it through `getByRole("option")` and the row carries `role="presentation"`, so the locator can never see it. The locator is wrong, not the product.'
    evidence: 'Reproduced at the baseline commit e80f607 in an isolated detached worktree verified clean, same assertion, same line — so this story did not cause it. It had never been observed because ci.yml runs `test:e2e:compile` only and this file had no recorded browser run. Registered rather than repaired: fixing an unrelated red inside 16.5 would blur what 16.5 changed, and it belongs with DW-171"s post-16.4 CI repair. The other test in the file — 31 of 31 EMBEDDED — PASSED in the same run. SAME DEFECT AS the 16.4 listbox finding (DW-141 / its pending amendment), one artifact over: a row that is not an option, addressed through getByRole("option"). Two independent sightings of one defect in two different files is what makes this a matter for 16.4 rather than a curiosity.'
  - summary: 'The family control offers all 1,304 families and, with the web arm kept as an install trigger, picking one from a dropdown can block for up to 30 s (AbortSignal.timeout in the default fetcher). GATE: Story 16.4.'
    evidence: 'Surfaced while writing 16.5s copy ruling. The remedy is a grouping change (narrowing the control to declared+installed and routing web browsing to the 16.3 modal), and grouping is 16.4s subject, not this storys. Registered so 16.4 inherits a measured concern instead of rediscovering it.'
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

This shipped. Choosing a typeface used to do three things in one action: download the font, keep a
copy on this machine, and write it into the template file — even for a font never used, and once
written it stayed. Those three have come apart. Browse freely, install what looks promising onto this
machine, and the font enters your template only when something in it is actually set in that face.
Story 8.6's title reverses, but its guarantee does not: the file still carries its fonts, so a
colleague's pages come out identical. Only the moment a font starts travelling has moved.

Two limits are deliberate and disclosed. One rare licence check could not move earlier without
creating a second authority over what enters a document, so a face can install and still be refused
on first use — and the refusal now says so, and points at the control that removes it from this
machine. A browser that will not keep anything falls back to the old behaviour, putting the font
straight into the document and saying why.

Three things are intentional, not oversights. A browser-run test here fails, and failed identically
before this story began: it looks for the right sentence in the wrong place, and its repair belongs
to the next story. One guard added late has no test of its own, because nothing an author can do
today reaches it. And two rows of this story's behaviour table are proven below the interface rather
than by driving it, so they are recorded as partial.

<intent-contract>

## Intent

**Problem:** A pick fetches, stores and embeds in one action, so a family the author merely tried
lands in the `.folio` and stays there — `assetKeyReferenced` scopes orphan collection to the single
key a command just repointed away from, never a document-wide sweep. The owner wants three separable
things: browse freely, install to this machine, and put in the file only what the template uses
(D-16.R.46).

**Approach:** Embed-on-use, with **zero engine change**. Split today's single path into an
**install** (fetch → classify → store, no engine command) and an **embed at first use**. Story 16.3
built one named seam for exactly this, `addFamilyToDocument`; this story swaps that function's body
and adds a third arm to `choose()` for *installed but not embedded*. An installed-but-unused face is
never in the file to be pruned.

## Boundaries & Constraints

**Always:**
- **Designer-only.** `embedFontFamily`, `fontset.RefuseContradictedLicence`, the licence tables,
  `assetKeyReferenced`, AD-8/CAP-2 and the `.folio` format are untouched.
- **Install runs every admission check that can run at install without creating a competing authority
  over what enters a document. A check that cannot stays where it is, this story names it, and the
  late refusal it implies is disclosed to the author rather than left to surprise them.**
  (D-16.R.46 clause as amended by the lead.)
- **The install-time `fvar` filter may only refuse, never admit.** Go's refusal still runs at embed
  and remains the only thing deciding what enters a document. Every drift outcome of a refuse-only
  filter is either today's behaviour or a loud failure at the moment the author acted; none of them
  writes a document.
- **Two commands, never one.** First use dispatches `embedFontFamily` then `updateComponentProperties`.
  The order is forced by the engine: `canvas.fontFamilies` is the closed set `style.fontFamily` may
  name, so the property command is refused unless the chain is already declared.
- Every Go invocation carries `-count=1`, recorded as invoked (DW-168).
- Stage only files this story wrote.

**Ask First:**
- Any pressure toward a **compound command, a new Go command kind, or any edit under `folio-go/`** —
  that is the epic's zero-engine-change clause and Story 8.6's refused fusion. HALT.
- Any urge to make the TypeScript `fvar` filter *admit* anything, or to port the nameID-13 tie.

**Never:**
- **Prune-on-save.** Ruled out: it is a mutation with no command, no history entry and no undo
  (AD-15), it breaks `open(save(d)) == d` against AD-9, and it requires exactly the document-wide
  sweep `assetKeyReferenced` refuses — a walk **known incomplete by its own comment**, with table
  cells explicitly out of scope, where under-reporting a reference deletes a live asset with no
  compile error to announce it.
- A copy of the `fvar` predicate sited **at or behind the embed command**, in either language.
- Reformatting `src/font-catalogue.test.ts`, `src/font-licence.ts` or `src/engine-protocol.ts`:
  **four Go tests read designer files as fixtures**, so "designer-only" is not structurally enforced
  and a reformat reddens Go without changing behaviour.
- Changing a test because it now fails. Every touched test is labelled **behaviour-changed** or
  **mechanical**; a test that cannot be so labelled is a finding.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Install a web family | Static TTF, admitted licence token | Fetched, classified, written to the store. **No engine command, no revision bump, no history entry.** | N/A |
| Install a variable face | Fetched bytes carry an `fvar` table | **Refused at install**, named, nothing stored | Refusal sentence in the caller's existing surface |
| Install, unmapped licence | METADATA.pb token not in the closed table | Refused at install, as today | Existing `classifyLicenceToken` refusal |
| First use of an installed family | Author sets a component's family to an installed one | `embedFontFamily` **then** `updateComponentProperties`; two history entries | Property command not sent if the embed is refused |
| First use, licence contradiction | Go's nameID-13 tie refuses at embed | Refusal states the face **is installed on this machine and cannot be embedded**, and points at the removal control | Property command not sent; face stays installed and removable |
| Family already in the template | Author picks a declared name | One `updateComponentProperties`. Unchanged. | Existing |
| Store write fails at install | Origin quota refuses the `put` | **The install fails and says so** — inverting 16.2's degradation, which has nothing left to degrade from | Refusal, not a silent success |
| Unparsable bytes | 200 that is not a font | TS refuses in `requireStaticTrueTypeTables` **before** the `fvar` lookup; Go's `RefuseVariableFace` returns `nil` for the same bytes | **Deliberate divergence — asserted on purpose**, not reconciled |

</intent-contract>

## Code Map

All anchors verified at `e80f607`. `git diff --name-only 4306362..e80f607` touches **0 files** under
`folio-go/`, `folio-designer/`, `lint/` or `scripts/`, so anchors measured during planning still hold.

**The seam and the fork (`folio-designer/src/App.tsx`)**
- `:864-873` `addFamilyToDocument` — **the named seam whose body this story swaps.** Engine guard,
  re-entry guard, `holdFontChain`, delegates to `resolveAndEmbedFamily`. Returns
  `Promise<string | undefined>`; **a string IS the refusal sentence** — it does not throw for anything
  it anticipates. Exactly **2 call sites**, both in `App.tsx` (`:877`, `:1283`); **0 occurrences
  anywhere else across 1,601 tracked files, and no test names it.**
- `:880-971` `resolveAndEmbedFamily` — the body: stored tier `:913`, local tier `:925`, web tier
  `:937`, tail proposal `:952`, **embed dispatch `:953`**, **store write `:970`**.
- `:875-878` `pickCatalogueFamily`; `:1283` the `FontBrowser onAddFamily` wiring.
- `:1777-1787` **`choose()` and the `THE FORK` comment.** Two arms today, discriminated on
  `match.source`: a declared name commits `fontFamily`; a catalogue family embeds **and deliberately
  does not also set `fontFamily`**. This story adds the third arm.
- `:1718` `FontFamilyProperty`; `:1521` its render site; `:1847` the panel refusal surface.
- `:360-365` `removeStoredFace`; `:1559-1572` `MachineFontStore`, whose per-face control carries
  a per-face control whose accessible name reads "Remove <family> (<style>) from this machine", at
  `:1570` — **this is the
  affordance the late-refusal disclosure points at. It is already shipped; do not build a second one.**

**The admission surface**
- `src/font-licence.ts:100-112` `classifyLicenceToken`, table `:73-82` — pure, synchronous,
  engine-free. Portable to install for free.
- `src/font-name-table.ts:39-49` `sfntTableDirectory` (returns `Readonly<Record<string, SfntTable>>`
  keyed by four-char tag), `:57-66` `requireStaticTrueTypeTables` (the sfnt-version guard),
  `:94` `nameTableString`, `:151-156` `faceCopyright`. **The new `fvar` predicate belongs here,
  beside `faceCopyright` and shaped exactly like it.**
- `src/font-source.ts:399-516` `fetchWebFamily`; `:490` the `faceCopyright` call.
- `folio-go/internal/fontset/variableface.go:33-75` — `variableFaceError` / `RefuseVariableFace`,
  **the sole authority, unchanged by this story.** Its comment at `:20-26` is amended (Task 6).
- `folio-go/internal/fontset/licencesignature.go:298-360` `RefuseContradictedLicence` — Go-only,
  deliberately browser-sited-refused, **stays**. Outcomes: contradiction refuses, confirmation admits,
  **no evidence admits**.

**The tie fixtures — the same files both languages read**
- `folio-go/testdata/fonts/notosansthai-variable-testonly/NotoSansThai-VF.ttf` (213.5 KB, carries
  `fvar`) and `folio-go/testdata/fonts/Roboto-Regular.ttf` (167.7 KB, static).
- `folio-go/testfont_embed_test.go:22,34` — Go reaches those **same paths** by `//go:embed`.
- `folio-go/component_commands_test.go:2139` `TestVariableFaceIsRefusedAtTheCommandAndAtIngestion`
  `OverTheSameBytes` and `:2191` `TestStaticFaceIsStillEmbeddedAtBothDoors` — **the construction to
  mirror, including its over-broadness control.**
- `src/engine-bounds-mirror.test.ts:44` — the established idiom for a designer test resolving
  `../../folio-go/...`; **13 designer test files already reference `folio-go`.**

**Tier state**
- `src/font-index.ts:69-75` `FamilySource` (`local | stored | web`), `:123` `addableFromTheWeb`,
  `:135-137` *"A HIDDEN ROW IS A PRESENTATION CHOICE, NEVER A GUARD… The authority stays Go."*
  `:257-271` `familySourceNote` with `const unhandled: never` at `:263`.
- `src/font-browser-model.ts:142-152` `rowTierNote` (`never` at `:147`), `:246-251` `rowState`
  (`'in-template' | 'staged' | 'addable'` — **no `never` guards this one**).
- `src/engine-protocol.ts:157-160` — `canvas.fontFamilies` is the closed set, from Go.

## Tasks & Acceptance

**Execution:**
- [x] `folio-designer/src/font-name-table.ts` -- add an exported refuse-only predicate over the face
      bytes (`'fvar' in requireStaticTrueTypeTables(fontView(bytes))`), shaped like `faceCopyright`
      and documented as **a filter that may only refuse**, naming Go as the authority. **Measured
      correction to the cost estimate:** `faceCopyright` builds the view and directory *internally*
      and returns only a string (`:151-156`), so `font-source.ts:490` does **not** hold the directory
      — this costs one extra table-directory walk per fetched face (header plus 12 bytes per record,
      no glyph parsing). The lead's conclusions stand unchanged: no new dependency, no new failure
      mode, and the container guard has already run.
- [x] `folio-designer/src/font-source.ts` -- run the new predicate on the fetched bytes, beside the
      existing `faceCopyright` call, refusing with a sentence that names the family and the cause.
- [x] `folio-designer/src/App.tsx` -- add `installFamily`: fetch → classify → store, **no engine
      command**. Swap `addFamilyToDocument`'s body onto it. Keep the busy hold, the generation
      guard, and both refusal surfaces (`announce: 'panel' | 'caller'`) exactly as they are.
- [x] `folio-designer/src/App.tsx` -- **invert the store-write ordering ruling and write the new one
      down where the old one lives (`:954-969`).** With no embed to follow, the store write is the
      only act, so a quota refusal now *does* decide whether the author gets the font. 16.2's
      "degradation, not a failed pick" no longer applies and its sentence (`font-store.ts:453-462`)
      must stop claiming the family "was added to this document".
- [x] `folio-designer/src/App.tsx` -- add `choose()`'s **third arm**: an installed family embeds
      **then** commits `fontFamily` — two commands, two undo entries. Do not fuse them. Preserve the
      existing two arms byte for byte, and extend the `THE FORK` comment to say why the third arm
      commits the property while the second still does not.
- [x] `folio-go/internal/fontset/variableface.go` -- **comment only, no code.** Three sentences:
      the prohibition binds copies whose **pass admits** (anything sited at or behind the command);
      an install-time filter that can **only refuse** is permitted, with the drift asymmetry named as
      the reason; and the name of the test that ties them. **A carve-out not written into the document
      it deviates from gets re-litigated or silently generalised.**
- [x] `folio-designer/src/font-browser-model.ts` and `folio-designer/src/font-index.ts` -- surface the
      third state: `rowState` (`:246-251`) gains an `installed` value, and `rowTierNote` (`:142-152`)
      and `familySourceNote` (`font-index.ts:257-271`) gain their arm. **Reinterpreting `'stored'` in place produces
      zero compile errors anywhere** — the discriminant string does not move — so add the state
      rather than redefining one, and let the two `const unhandled: never` bindings do their work.
- [x] `folio-designer/src/App.tsx` -- the **late-refusal disclosure**: when the embed at first use is
      refused, say the face is installed on this machine and cannot be embedded, and point at the
      existing removal control (`:1570`). One message, reusing shipped machinery.
- [x] `folio-designer/src/font-name-table.test.ts` (or a new tie test) -- **the behavioural tie.**
      **Two facts to state in the test's own comment, because they are what make it a tie rather than
      a lookalike:** (i) the fixtures are **the very files Go `//go:embed`s** — same paths, same
      bytes, not a copy (`folio-go/testfont_embed_test.go:22,34`); and (ii) this **mirrors an existing
      named construction** rather than inventing one — `TestVariableFaceIsRefusedAtTheCommandAndAt`
      `IngestionOverTheSameBytes` (`component_commands_test.go:2139`) with
      `TestStaticFaceIsStillEmbeddedAtBothDoors` (`:2191`) already standing as its **named
      over-broadness control**. Four requirements: **both arms** (the variable fixture must
      be refused, the static fixture must install — a refuse-everything filter passes a refusal-only
      test); **assert the deliberate divergence on purpose** (Go returns `nil` for an unparsable face,
      TS throws earlier in `requireStaticTrueTypeTables`) so a later reader does not "fix" it by
      widening one side; **fail, never skip, on a missing fixture**; and red-proof by **deleting the
      predicate** in its own worktree, serially, tree verified clean before and after, **asserting the
      mutation applied** (D-16.R.55).
- [x] `folio-designer/src/App.test.tsx`, `src/App.font-store.test.tsx`, `src/FontBrowser.test.tsx`,
      `src/font-browser-model.test.ts` -- rewrite the **9 trigger-dependent** cases and the copy-dependent cases per the ruling
      at CHECKPOINT 1. **Label every touched test behaviour-changed or mechanical.**
- [x] `folio-designer/src/App.test.tsx`, `folio-designer/src/App.font-store.test.tsx` -- **neutralise
      the three vacuous survivors**, each of which asserts "no command was
      dispatched" and becomes trivially true when nothing dispatches at pick:
      `App.test.tsx:1755`, `App.font-store.test.tsx:229`, `App.font-store.test.tsx:357`. Give each a
      claim that can still fail, and **prove it per test by making the new claim actually fail** —
      observing that the suite is green proves nothing here, because a tautology is green too. This
      is the failure this story is most likely to manufacture more of, since the behaviour it changes
      IS "a command is dispatched at pick" and these three assert exactly that absence. Re-anchor the two implicit ordering barriers
      (`App.font-store.test.tsx:302,347`), which used the store row's appearance as a proof the
      command had already run.
- [x] `folio-designer/e2e/` -- update the two specs that encode the old behaviour
      (`font-embed-boundary.spec.ts:127` reads a revision advance as `EMBEDDED` and demands 21/21;
      `font-browser.spec.ts:102`). **Both are compile-only and cannot redden in CI** — `ci.yml:218`
      runs `test:e2e:compile` and `playwright test` appears **0 times across the 93 gate-shaped
      files** searched. State plainly whether a real browser run was performed.

- [x] `folio-designer/src/App.tsx` -- **REVERT the stored-read-miss change (orchestrator ruling, 2026-09-03).**
      A stored read that misses or fails to decode must **fall back to a refetch**, as it did before this
      story. Story 16.2's matrix is explicit and is a shipped contract this story may not amend:
      *"Stored bytes fail to decode later | Corrupt entry | Entry treated as absent and dropped; refetch
      on next pick | Self-healing, logged honestly."* Removing the fallback converts a self-healing path
      into a permanent local failure clearable only by a manual removal, and does it silently.
- [x] `folio-designer/src/App.tsx` -- **A STORE THAT CANNOT BE OPENED DEGRADES TO TODAY'S BEHAVIOUR, not
      to a refusal (orchestrator ruling, 2026-09-03).** When the store is unavailable, the pick **embeds
      directly**, exactly as it does before this story. This is the only option that satisfies both
      standing contracts at once: 16.2's *"Given storage that cannot be opened or written, when the
      designer runs, then it still works and says what is degraded"*, and D-16.R.46 Q4's refusal of a
      dead end. **Say so in the degradation message** — the author is told the font went straight into
      the document because this browser will not keep it. IndexedDB stays a convenience, never a
      dependency.
- [x] `folio-designer/src/` -- **the Row 5 test must prove the message reaches the AUTHOR on the refusal
      path.** A test that merely finds the string in a module is the "guard that observes the declaration
      and not the behaviour" defect from 16.3 in a third medium. Drive the refusal; assert what renders.

- [x] `folio-designer/src/App.font-store.test.tsx` -- **TEST THE DEGRADED ARM (lead guardrail, 2026-09-03).**
      After this story, embed-at-pick stops being *"what the product does"* and becomes **a retained arm —
      and a retained arm is an unmeasured claim, the half nobody changed and nobody exercises.** Assert
      that with the store unavailable **a pick embeds, the document changes, and the message says why**.
      **Red-prove it by REMOVING the fallback**, in its own worktree, with the mutation asserted as
      applied. Without this, 16.5 ships a path that executes only on machines nobody tests on, which is
      how a fallback rots into a crash.

**Acceptance Criteria:**
- Given a family the author installs, when the install succeeds, then the face is on this machine and
  **the document's revision, history and asset map are unchanged**.
- Given fetched bytes carrying an `fvar` table, when they are installed, then the install is refused
  at the moment the author acted, and nothing is stored.
- Given an installed family, when a component is first set to it, then the face is embedded and the
  property committed as **two commands with two undo entries**, in that order.
- Given a face Go refuses at embed on a licence contradiction, when first use is attempted, then the
  refusal states the face is installed and unembeddable and points at the removal control, and the
  face remains installed and removable.
- Given the shared fixtures, when the tie test runs, then the variable face is refused and the static
  face is admitted, and deleting the predicate reddens it.
- Given the whole suite, when it runs, then no test asserting "no command was dispatched" passes
  vacuously.
- Given a browser whose store cannot be opened, when the author confirms staged families, then the
  confirm affordance **names the count** — *"this browser will not keep fonts, so confirming adds N
  families to this document"* — because in degraded mode confirming five staged families embeds five
  faces, which is the outcome the owner's reversal was made to prevent. Acceptable as a stated
  degradation of a rare path; not acceptable unstated.

## Spec Change Log

**The CHECKPOINT 1 copy ruling was not on disk, and the builder ruled the copy against the contract
instead.** Two tasks refer to *"the ruling at CHECKPOINT 1"* and the third deferral refers to *"16.5's
copy ruling"*, but no such ruling exists in `epic-16-decision-log.md` or anywhere in
`_bmad-output/`. The copy below was therefore derived from this spec's own I/O matrix and acceptance
criteria rather than invented freely, and every string is named here so a later reader can overrule one
without re-reading the diff:

| where | was | is |
|---|---|---|
| `confirmLabel` | `Add N to template` | `Install N on this machine` |
| `pendingLine` | `N families ready to embed` | `N families ready to install` |
| `buttonLabel('addable')` | `+ Add` | `+ Install` |
| `buttonLabel('installed')` | — (new state) | `On this machine` |
| `buttonName('addable')` | `Add X to this template` | `Install X on this machine` |
| `buttonName('staged')` | `Remove X from the families to add` | `Remove X from the families to install` |
| `buttonName('installed')` | — (new state) | `X is already on this machine` |
| `rowTierNote('web')` | `downloaded when you add it` | `downloaded when you install it` |
| `familySourceNote('local')` | ` — add to document, already on this machine` | ` — use it, already on this machine` |
| `familySourceNote('stored')` | ` — add to document, already downloaded to this machine` | ` — use it, already downloaded to this machine` |
| `familySourceNote('web')` | ` — add to document` | ` — install on this machine` |
| `storeWriteDegradation` | `X was added to this document, but…` | `storeWriteRefusal`: `X was not installed on this machine (…). Nothing was kept and no document was changed.` |

**"`rowTierNote` and `familySourceNote` gain their arm" was read as REVISED ARMS, not a fourth tier.** A
`FamilySource` tier says where a face's BYTES come from, and install/embed separation added no byte
source — it split what a pick DOES. So `RowState` gained `installed` (the task's own first clause) while
`FamilySource` did not gain a tier, and the two `const unhandled: never` bindings are untouched and still
standing. The distinction the story needs is exported once, as `familyIsInstalled` in `font-index.ts`,
and both the dropdown's fork and the browser's row state read it — so the two surfaces cannot come to
different conclusions about the same face.

**`installed` = `local` ∪ `stored`, and the local tier therefore takes the THIRD arm.** The line is *"can
these bytes be had with no network"*. Routing the local tier to the second arm would have made picking
one of the 31 committed faces do literally nothing — `keepOnThisMachine` is guarded by `if (fetched)` and
a bundled face is never fetched — which is the dead end D-16.R.46 Q4 forbids. It also keeps
`font-embed-boundary.spec.ts`'s 31-of-31 EMBEDDED measurement true and meaningful.

**A stored-read MISS at first use is now refused rather than falling through to a fetch.** Story 16.2's
fall-through was correct while a pick fetched anyway. Under the split, the store IS the installed set, so
a face that has left it is not installed; answering a property set with up to six cross-origin
round-trips would put the install trigger straight back on the path this story removed it from, and it
would resurrect the second store-write sentence the `font-store.ts` task exists to delete. Stated in
`embedInstalledFamily`'s own doc.

**A store that cannot be OPENED now refuses the install too.** 16.2 called that *"a cache and a source,
never a precondition"* and let the pick embed regardless. There is no embed behind an install, so an
install into a browser that will not keep anything has produced nothing, and reporting it as a success
with a footnote is the exact inversion this story's matrix row forbids. `App.font-store.test.tsx`'s
*"still designs when the store cannot be opened at all"* is labelled behaviour-changed and asserts the
refusal.

### 2026-09-03 (later) — SUPERSEDING NOTE on this log's first, undated entry

**The log is append-only, so the paragraphs below are left standing and this is the forward pointer a
reader landing on them needs.** Two claims in the FIRST entry of this Spec Change Log — the undated one
headed *"The CHECKPOINT 1 copy ruling was not on disk…"* — describe behaviour that the entry immediately
below this one, **`2026-09-03 — orchestrator rulings applied after the step-03 matrix audit`**, reversed
within hours:

| paragraph in the undated entry | what it asserted | superseded by |
|---|---|---|
| *"A stored-read MISS at first use is now refused rather than falling through to a fetch."* | the fall-through was removed | **2026-09-03 rulings, item 4** — reverted; Story 16.2's self-healing refetch stands, and the face is now written BACK after the embed. It also cites a test name this change renamed. |
| *"A store that cannot be OPENED now refuses the install too."* | an unopenable store refused | **2026-09-03 rulings, item 3** — refused; an unopenable store degrades to the pre-16.5 model and EMBEDS directly, saying so. |

**Both reversals were the gate working, and both are now covered by tests that did not exist when the
undated entry was written** — `refetches and heals when a stored entry has gone missing` and `still
designs and still adds fonts when the store cannot be opened at all`. The rest of the undated entry —
the copy table, the fourth-tier reading, and `installed = local ∪ stored` — **stands unchanged**.

**One row is ADDED to that entry's copy table rather than edited into it** (patch 10):

| where | was | is |
|---|---|---|
| `font-source.ts` — every refusal it produces | `${family} cannot be added: …`, `You cannot add a family without a network connection` | `${family} cannot be installed: …`, `You cannot install a family without a network connection; … using one of those needs no network at all` |

That module now runs on exactly one path — the install — so *"cannot be added"* described a step that no
longer happens there, which is the false-UI-string class this epic has ruled against four times.

### 2026-09-03 — orchestrator rulings applied after the step-03 matrix audit

**1. THE COPY RULING WAS WRONG AND THE BUILDER CORRECTED IT. Recorded as a correction, not as the
ruling.** CHECKPOINT 1 ruled *"install language for web/stored"*. **A `stored` face is already
installed**, so picking one is a *use*, not an install; `stored` takes use language. The
implementing agent did not diverge from the ruling — it fixed it, and the record says so rather than
quietly shipping the better answer under the old ruling's name.

**2. THE DEEPER FINDING, AND IT IS A STANDING RULE NOW.** That ruling reached the implementer only
through the orchestration channel, and two tasks pointed at *"the ruling at CHECKPOINT 1"* — **a
pointer to nothing.** The spec is the implementer's sole source of truth by design, so **any ruling
the implementer must act on is written INTO the spec, verbatim, before dispatch. A ruling that exists
only in the orchestration channel does not exist.** This entry is the first application of that rule.

**3. STORE-OPEN FAILURE: the builder's refusal is REFUSED, and neither obvious answer was right.**
Refusing the install contradicts Story 16.2's locked contract (*"Storage failure is a degradation,
stated"*). But keeping 16.2's degradation *literally* also fails, and the builder's instinct was
sound about why: **under the old model the store was a convenience because the pick embedded
immediately; under embed-on-use the store is load-bearing**, so an install into an unopenable store
stores nothing and nothing can ever be used — the dead end Q4 forbids. **The third option: degrade to
the pre-16.5 model and embed directly.** It is also the honest description of what that mode is —
with nowhere to put an installed face, embedding at once is the only way to use a web font.

**4. STORED-READ MISS: reverted.** See the task above; 16.2's self-healing refetch stands.

**5. WHY THE MATRIX BLOCKER WAS THE GATE WORKING.** Row 5's disclosure existed in `App.tsx` and was
asserted by **zero tests across 22 designer test files and 15 e2e specs**. Built and not proven is
exactly the failure the disclosure ruling was written to prevent.

## Design Notes

**Why a refuse-only install filter is not the second authority D-16.6 forbids.** An authority is a
check whose **pass decides admission**. Test both candidates by what drift does:

| | drift → permissive | drift → strict |
|---|---|---|
| command-side copy | a variable face **admitted** into a `.folio` that fails at render — D-16.6's original defect reproduced | a legitimate face refused at the command |
| install-time filter | face installs, Go refuses at first use — **exactly today's behaviour** | a legitimate face fails to install — **loud, at the moment the author acted** |

Every drift outcome of the install filter is either the status quo or a loud failure; **none writes a
document.** The repository already relies on this distinction: `font-index.ts:135-137` ships a
designer-side variable-face filter and says of itself that it is a presentation choice and that the
authority stays Go. This story **upgrades that filter's evidence from a build-time snapshot field to
the bytes in hand.** It improves an existing filter; it does not add a door. Measured: **537 of 1,811
index rows** are filtered out by the snapshot flag today.

**Why the tie is behavioural and not source-text.** `TestGoLicenceTableSubsumesTheDesignerTable`
reads TypeScript as text because the designer file exports nothing and because the thing compared is
**a table of declared ids**. The `fvar` predicate has no table; matching source text would assert
that the string `'fvar'` appears in a file — spelling, not behaviour.

**Why the byte-snapshot short-circuit is relied upon rather than observed.** History is whole
canonical `.folio` byte snapshots and `Apply` short-circuits when the bytes do not move
(`folio-go/wasm/engine.go:244-246`), so an action touching no document bytes yields no revision bump,
no history entry and no undo **by construction**. That is what makes "no engine command" free, and it
is why this is a story rather than an epic.

**Why the compound command is closed in both directions.** It is **unavailable** — no batch,
transaction or history-grouping mechanism exists anywhere (0 hits for eight candidate names across
1,601 tracked files), the wire refuses trailing content
(`folio-go/component_commands.go:48-51`), and the kind vocabulary is a closed 25-case switch. It is
also **unnecessary** — the order is forced by validation, so two commands give two unambiguous undos.
Story 8.6's refused *fusion* is not reopened.

**Why nameID-13 stays in Go, and why that does not collide with D-16.R.46 Q4.** Q4 was written about
**disk fonts**, a class every member of which is unembeddable and knowable in advance at zero cost —
*"a face that can **never** be embedded"*. The nameID-13 residue is **per-face**, discoverable only by
parsing a name table and running two regex tables, and only an active **contradiction** refuses
(no evidence admits). If install can cheaply know a face is unembeddable it must not install it; if
knowing requires machinery that would create a competing authority, the late refusal is accepted —
**and disclosed**. A dead end the author can see and clear is a stated limit; a dead end that simply
fails is what Q4 forbids.

**The in-memory `FontStore` was considered and rejected — recorded so nobody re-derives it.** The
shipped seam already takes an injectable factory
(`openFontStore(factory: IDBFactory | undefined = globalThis.indexedDB)`), so a session-scoped
in-memory store behind the same interface is ~20 lines with **zero downstream branches**, and on the
maintenance-surface axis it is genuinely one path with two backends. **It loses on Story 16.2's locked
matrix row:** *"Private window / storage blocked → Group empty; picks still fetch; message states
it."* An in-memory store makes that group **non-empty**, contradicting the row directly, and amending
a locked contract is not a story's to do. **Degrading to the pre-16.5 model preserves both letter and
intent:** nothing is stored so the group stays empty, the pick still fetches, and it now also embeds —
which the row never forbade, because before 16.5 that is exactly what a pick did.

## Verification

**Commands** (heavy tests are `end-of-run`, D-16.R.1; `npm run test` and `npm run build` are separate
because `&&` short-circuits on the standing DW-152 red):
- `cd folio-designer && npm run typecheck` -- expected: clean.
- `cd folio-designer && npm run lint` -- expected: exactly 4 pre-existing `only-export-components`
  warnings; re-measure their line numbers rather than relaying them.
- `cd folio-designer && npm run test` -- expected: full pass. Record file/test counts before and
  after and **diff the name sets**, not the totals.
- `cd folio-designer && npm run build` -- expected: succeeds. Requires node `v24.16.0`.
- `cd folio-designer && npm run test:e2e:compile` -- expected: clean. **This is a compile, not a run.**
- `cd folio-go && go test -count=1 ./...` -- expected: unchanged from the measured baseline. Take the
  baseline in a detached worktree at `e80f607`; **do not assume "exactly one red"** — that figure has
  moved between stories.
- `cd lint && go test -count=1 ./...` -- expected: unchanged. `-count=1` is mandatory here: the rules
  package walks a directory and Go's test cache does not track `ReadDir`.
- `cd folio-designer && npx vitest run <tie test>` after **deleting** the predicate -- expected: red,
  with the mutation asserted as applied, in a separate worktree from any observing agent.

**A CITED IDENTIFIER IS VERIFIED BY LOCATING ITS DEFINITION, NEVER BY RE-READING THE CITATION.**
`DW-180` was cited three times — including a decision-log line reading *"Registered as DW-180"* — while
no such entry existed in the register. **An identifier is a carrier whose text is one line long, which
is precisely why it gets checked least: the citation names an id, and an id looks like evidence.**
Closing sweep, reported PER IDENTIFIER and never as a verdict: every `DW-\d+` cited in this epic's
artifacts must resolve to a definition line in `deferred-work.md`.
**Measured at 16.5's close: 37 distinct identifiers cited across the 11 epic-16 artifacts, 37 resolved,
0 unresolved; second population 503 source files under `folio-designer/src` and `folio-go`, 45 distinct
identifiers, 0 unresolved.** Register held 182 definition lines.

**A MATRIX AUDIT REPORTS PER ROW, NOT PER VERDICT.** N rows, N results. Row 7 was missed by an audit
that reported "clean" while having checked only the two rows already flagged — *"matrix audit clean"*
reads as ALL rows and meant THE FLAGGED ROWS, which is this epic's signature defect committed in the
check itself. **And any row whose only assertion calls the production function directly, rather than
driving the path a user takes, is FAILED, not passed** — the Row 5 ruling applied to every row rather
than only to the row it was written about.

**Reading a status (the fifth false green of this epic):** **a status that passed through a wrapper is
unverified until the wrapper's own capture is read.** Take `rc=$?` immediately after the command and
assert on it. **The tell is a wrapper whose last statement is not the command** — an echo, a summary
line or a cleanup step each silently become the status. **Record the invocation, not only the
verdict**, so a reader can see whose status is being quoted.

**Manual checks:**
- **Nothing in this epic may be described as CI-verified**: the designer CI job halts at step 2 until
  the post-16.4 repair.
- State explicitly whether a real browser run was performed for the two e2e specs, or that none was.

## Measured Gates (builder, on the working tree over `e80f607`)

Every figure below was run locally by the builder and read from the command's own output, not through a
pipe. Nothing here is CI-verified; the designer CI job still halts at step 2 (DW-171).

| gate | baseline at `e80f607` | after |
|---|---|---|
| `npm run test` | **659 passed / 1 failed of 660**, 54 files | **672 passed / 1 failed of 673**, 55 files (`rc=1`, read from `$?` immediately after the command, never through a pipe) |
| the one red | `canvas-authority-contract` — `e9-5-border-no-ink.spec.ts` `getComputedStyle` (standing DW-152) | **the same test, unchanged** |
| `npm run typecheck` | — | clean |
| `npm run lint` | — | `rc=0`, exactly **4** `only-export-components` warnings: `preview/pdf-viewer.tsx:16,17` and `App.tsx:2141,2148` (re-measured after the final edit, not relayed) |
| `npm run build` | — | succeeds, node v24.16.0 |
| `npm run test:e2e:compile` | — | clean. **This is a compile, not a run.** |
| `folio-go` `go test -count=1 ./...` | exit 1, **only** `TestCorpusMeetsP6ExerciseFloors/P6g` | exit 1, **failing-test NAME SET diffed and identical** |
| `lint` `go test -count=1 ./...` | — | exit 0, 4 packages |

**The suite delta is a NAME-SET diff, not a total.** **15 names left, 22 arrived.** Every departure is
a paired rename of a test this story re-anchored; the seven net additions are the tie test's 3 cases and
four new ones — the two audited matrix rows, the restored self-healing refetch, and the degraded
confirm affordance. **No test disappeared.**

**A REAL BROWSER RUN WAS PERFORMED**, chromium-1228 (Chrome for Testing, arm64) against the built
bundle, via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` — the pinned `chromium_headless_shell-1208` Playwright
1.58.2 asks for is not installed on this machine, which is DW-180's browser pin biting again.

- `e2e/font-browser.spec.ts` — **6 of 6 passed.**
- `e2e/font-embed-boundary.spec.ts` — **31 of 31 families EMBEDDED**, that test passed. Its sibling
  *"the designer offers the whole local face tier"* FAILED, and **the failure was reproduced at
  `e80f607` in an isolated clean worktree**: pre-existing, registered as a deferral, not repaired here.

**Red-proofs, each with its mutation asserted applied by diffing the file (D-16.R.55), run serially.**

1. **The tie test**, in its own detached worktree at `e80f607` carrying only this story's files (`git
   status --porcelain` verified before and after to list exactly those files and nothing else).
   `faceIsVariable`'s body replaced by `return false` — diff printed, non-empty. `font-variable-face-tie.test.ts`
   went **1 failed / 2 passed**: the variable arm reddened, the static arm and the divergence arm stayed
   green, which is what says the arms measure different things. Predicate restored, 3/3 green again.
2. **`App.test.tsx`'s vacuous survivor** (`expect(request).not.toHaveBeenCalled()`): the declared-name
   property commit in `choose()` replaced by a no-op. The OLD line stayed green — which is the whole
   point — and the new *"the observer is alive"* claim reddened (`expected [] to have a length of 1`).
3. **`App.font-store.test.tsx`'s browser-refusal survivor** (`expect(embedPayloads).toEqual([])`): a
   well-formed record written to the store BEFORE the fetch outcome is checked. The old line stayed
   green; the new *"a refused install must keep nothing"* claim reddened.
4. **`App.font-store.test.tsx`'s removal survivor** (`toHaveLength(embedsBefore)`): the `fontFamily`
   commit dropped from the third arm. The new two-command ordering claim reddened
   (`['embedFontFamily'] != ['embedFontFamily','updateComponentProperties']`), which is what makes
   `embedsBefore` a non-zero number worth guarding.

Each mutation was reverted and the file diffed against its pristine copy to prove the revert.

### The step-03 matrix audit, and the rulings that followed it

**TWO MATRIX ROWS SHIPPED UNASSERTED AND THE AUDIT CAUGHT BOTH.** Row 2's `fvar` filter was proven at
the PREDICATE and never on the INSTALL PATH; Row 5's late-refusal disclosure existed in `App.tsx` and
was asserted by **zero tests across 22 designer test files and 15 e2e specs**. Both are now covered by
tests that drive the real path in a mounted designer and assert what RENDERS — never a module string.

| new test | what it owns | red-proved by |
|---|---|---|
| `refuses a variable face at install, names it, and keeps nothing on this machine` | Row 2, over **the file Go `//go:embed`s**, with the store read back **past the designer** through `openFontStore` | (a) disabling the `font-source.ts` filter — reds the message half; (b) writing a sound record before the fetch outcome is checked — reds the *nothing stored* half **while every message assertion stays green**, which is the half most easily left out |
| `says an installed face cannot be embedded, points at the removal control, and leaves it installed` | Row 5. The pointer's own words are handed to `getByRole` as an **accessible name**, so prose naming a control that is not there reds it; the test then **presses that control** and proves the dead end clears | (a) dropping `lateEmbedRefusal` — the author gets the engine's bare sentence; (b) committing the property despite the refusal — reds the two-command ordering |
| `refetches and heals when a stored entry has gone missing, rather than refusing first use` | Story 16.2's self-healing contract, which this story briefly broke **with the suite staying green, because nothing covered it** | restoring the refusal — reds both the embed and the heal |
| `still designs and still adds fonts when the store cannot be opened at all, by embedding directly and saying so` | the **retained** pre-16.5 arm — a retained arm nobody exercises is an unmeasured claim | **REMOVING the fallback** (not mutating around it), in its own worktree |
| `names the count in the confirm control and says so in the footer, and stops when the store works` | the degraded confirm affordance, with an **over-broadness control**: re-rendered with a working store, none of the warning may appear | making `confirmName` ignore the degraded flag |

All five mutations were applied in a detached worktree at `e80f607` carrying only this story's files
(`git status --porcelain` read before and after), run **serially**, each **diffed to prove it applied**
before the run, each reverted and re-diffed after. The worktree was removed and the tree re-verified.

**The in-memory `FontStore` was not implemented.** See Design Notes: it contradicts Story 16.2's locked
*"Private window / storage blocked → Group empty"* row, and amending a locked contract is not this
story's to do.

### The review's ten patches, and the three alibis they closed

**Ten PATCH findings, all applied.** Three were serious and all three had the same shape: **a clause
that exists, a test that asserts it, and an alibi that still holds** — so each was red-proved by the
mutation that exposes its alibi rather than by one that merely breaks the feature.

| # | patch | mutation run | observed |
|---|---|---|---|
| 1 | `commitFirstUse` had no generation/selection re-check between the embed and the property commit | **remove the guard**, then replace the document mid-embed | `updateComponentProperties` **lands against the replaced document** carrying the previous one's element ids. It matters because `applyProperties` **sends first and guards after** — a stale commit is not a dropped response, it is a command reaching the engine. Guard sited in `embedInstalledFamily`, where `documentGeneration`/`selectedRef` are live; the family control's own copies are the render's and would compare to themselves. |
| 2 | matrix row 7 (quota refusal at install) covered only by a module-level string check | **`if (kept !== undefined) return undefined`** | the author is told **nothing** — `Unable to find role="alert"` — while **274 tests across six other files stay green**. That is the silent success the row forbids, and nothing could see it. |
| 3 | `storeKeepsFaces` asserted only where the prop is hand-passed | **hardcode `storeKeepsFaces={true}` at the `App.tsx` call site** | the new App-level test reds; **`FontBrowser.test.tsx` stays 25/25 green** — the alibi, demonstrated and then closed. |
| 4 | WCAG 2.5.3 Label in Name: accessible name shared no words with the visible label | return the warning without the label | reds. The name now **contains** `Install N on this machine`, so a speech-input author can say what they see. |
| 5 | `storeWriteDegradation` had zero test references after its wording changed | delete the `setStoreNote` on a failed heal | the failed heal goes **completely silent** — `Unable to find role="status"`. Its three sentences are now pinned against each other. |
| 6 | `embedInstalledFamily`'s no-engine and busy early returns were silent | — | routed through `refuseFontChain`. **No behavioural test: the path is not reachable from the UI today** — the combobox is disabled whenever `fontChainBusy \|\| fileBusy` — so the fix is defence against the ref/state lag that D-16.R.15 named, and a test asserting the guard's presence would be the spelling test this story forbids. **Stated rather than faked.** |
| 7 | `lateEmbedRefusal`'s `removable === false` arm was unreached | **hoist `removable` to `true`** | the local-tier test reds while the stored-tier Row 5 test **stays green** — a bundled face would have pointed at a `Remove …` control that is not on screen, the same failure Row 5 exists to catch, one tier over. |
| 8 | `rowState`'s `installed` parameter defaulted to `[]` | make it optional again | `tsc` fails: **`Unused '@ts-expect-error' directive`**. The `never` bindings guard the arms; this guards the input. |
| 9 | the log asserted two reversed behaviours as current | — | **appended** a superseding note naming both by date; the originals are untouched, because the log is append-only. |
| 10 | `font-source.ts` still said "cannot be added" / "add a family" | — | migrated to install language; the word "add" survives nowhere in the sentences it produces. A row is added to the copy table in the superseding note. |

Every mutation was **diffed to prove it applied** before its run, run **serially**, restored by absolute
path and re-diffed, in a detached worktree at `e80f607` carrying only this story's files. The worktree
was removed and `typecheck` re-run clean afterwards.

**A second real browser run** was performed after the confirm control changed:
`e2e/font-browser.spec.ts` **6 of 6, `rc=0`** (chromium-1228, against the built bundle). In a real
browser IndexedDB exists, so the dialog is in its ordinary mode and the confirm control keeps its
`Install N on this machine` name — which is why that spec needed no edit for the degraded copy.

## Suggested Review Order

**The split itself — start here**

- The one named seam whose body this story swapped; everything else follows from it.
  [`App.tsx:880`](../../folio-designer/src/App.tsx#L880)

- Install: fetch, classify, store. Sends no engine command — the whole point.
  [`App.tsx:964`](../../folio-designer/src/App.tsx#L964)

- First use: the embed, and where the staleness guard is sited.
  [`App.tsx:1066`](../../folio-designer/src/App.tsx#L1066)

- The fork's third arm; the two existing arms are byte-for-byte unchanged.
  [`App.tsx:2041`](../../folio-designer/src/App.tsx#L2041)

- Two commands, two undos, order forced by the engine — never fused.
  [`App.tsx:2011`](../../folio-designer/src/App.tsx#L2011)

**Admission moved earlier, without a second authority**

- The refuse-only filter: one key off a directory the fetch path already builds.
  [`font-name-table.ts:198`](../../folio-designer/src/font-name-table.ts#L198)

- Called beside the existing copyright read, inside the same container guard.
  [`font-source.ts:513`](../../folio-designer/src/font-source.ts#L513)

- The carve-out written where the prohibition lives; comment only, no code.
  [`variableface.go:43`](../../folio-go/internal/fontset/variableface.go#L43)

**Degradation, which is where the rulings landed**

- Store unavailable degrades to the pre-16.5 model and says so.
  [`App.tsx:22`](../../folio-designer/src/App.tsx#L22)

- A failed store write refuses the install rather than reporting success.
  [`font-store.ts:469`](../../folio-designer/src/font-store.ts#L469)

- Late refusal names the installed face and points at the removal control.
  [`App.tsx:1123`](../../folio-designer/src/App.tsx#L1123)

**The third state, and what the author reads**

- "Installed" is tier-based and never consults the store.
  [`font-index.ts:264`](../../folio-designer/src/font-index.ts#L264)

- The new row state; its `installed` argument is required, not defaulted.
  [`font-browser-model.ts:283`](../../folio-designer/src/font-browser-model.ts#L283)

- Degraded confirm names the count, and contains its own visible label.
  [`font-browser-model.ts:405`](../../folio-designer/src/font-browser-model.ts#L405)

- Where that reaches the author.
  [`FontBrowser.tsx:347`](../../folio-designer/src/FontBrowser.tsx#L347)

**Peripherals**

- The cross-language tie: both arms plus the deliberate divergence, over the files Go embeds.
  [`font-variable-face-tie.test.ts:1`](../../folio-designer/src/font-variable-face-tie.test.ts#L1)

- Row 7's covering test — the one the first audit missed.
  [`App.font-store.test.tsx:651`](../../folio-designer/src/App.font-store.test.tsx#L651)

## Delivery Log

### 2026-09-03 — done

Baseline `e80f607`. Shipped as specified: the single pick action is split into an **install** (fetch →
classify → store, dispatching no engine command) and an **embed at first use** that sends
`embedFontFamily` then `updateComponentProperties` as two commands with two undo entries. Zero engine
change — the only edit under `folio-go/` is the 24-line carve-out comment in `variableface.go`, which
records why a refuse-only install-time filter is not the second authority D-16.6 forbids. Story 8.6's
title reverses; its guarantee does not.

**Decisions applied by ID.** D-16.R.46 (and its Q4 amendment) is the owner decision the story exists to
discharge. D-16.R.59 — a ruling that reached the implementer only through the orchestration channel is
a pointer to nothing — became the standing rule that any ruling the implementer must act on is written
into the spec verbatim before dispatch; this story is its first application, and the `2026-09-03 —
orchestrator rulings applied after the step-03 matrix audit` entry is that rule discharged. D-16.R.55's
mutation-asserted-applied discipline governed every red-proof. D-16.R.63's rule — mutate to expose the
*alibi*, not to break the feature — governed the three HIGH patches. D-16.R.69 records the patch round;
D-16.R.70 records why the two shared record files were committed separately.

**Findings triaged: 10 patched / 0 review-deferred / 0 rejected.** All ten review findings were applied
and red-proved; no finding was deferred or rejected in triage. Separately, the story registers **four
deferrals** in its own frontmatter (the README e2e-gate claim, the Go-only nameID-13 residue, the
pre-existing browser red in `font-embed-boundary.spec.ts`, and the 30-second dropdown-block hazard gated
to Story 16.4). No rejection tally appears anywhere in the record, because there were no rejections —
stated here so a later reader does not read the absence as an omission.

**What the review caught.** Three review layers with disjoint yields and one corroborated defect
(D-16.R.66). The three HIGH items shared one shape — a clause that exists, a test that asserts it, and
an alibi that still holds. The staleness guard was sited in `embedInstalledFamily` rather than
`commitFirstUse` where the finding pointed, because the family control's `documentGeneration` and `ids`
are the render's, frozen in the closure, so a check there compares them to themselves; the red-proof
also established that `applyProperties` **sends first and guards after**, making a stale commit a
command that reaches the engine rather than a dropped response. Row 7's quota refusal was covered only
by a module-level string check: replacing the refusal with `return undefined` left the author told
nothing while 274 tests across six other files stayed green.

**PATCH 6 HAS NO BEHAVIOURAL TEST, DELIBERATELY, AND THE REASONING IS THE POINT.** The busy/no-engine
early returns in `embedInstalledFamily` were silent and are now routed through `refuseFontChain`. That
path is **unreachable from the UI today**: the family combobox is `disabled={pending || pickBusy}`, and
`pickBusy` is fed from `fontChainBusy || fileBusy` at the `ComponentProperties` call site, so an author
cannot open the window the guard closes. The guard defends against the **ref/state lag** D-16.R.15
named — it reads `fontChainBusyRef.current`, which updates synchronously, while the combobox's disabled
state re-renders a tick later. So it is defence against a lag, not a fix for a reachable bug. A test
could therefore only assert that the guard is *present*, which is exactly the spelling test this story
forbids and which 16.3 committed in a different medium. Writing one would have converted an honest
defence into a false measurement. Verified at close: the disabled condition and the ref-vs-state split
are both still as described.

**THE MATRIX AUDIT FAILED ITS OWN FIRST EXECUTION, AND THAT FAILURE IS WHY THE RULE EXISTS.** The
original step-03 audit reported the matrix "clean" while having checked **only the two rows already
flagged** — it never looked at row 7, which was the row that was actually broken. *"Matrix audit clean"*
read as ALL rows and meant THE FLAGGED ROWS: this epic's signature defect committed inside the check
written to catch it (D-16.R.62). The correction is the standing rule now named in `## Verification` —
**a matrix audit reports N rows and N results, never a single verdict** — together with its companion,
that any row whose only assertion calls the production function directly rather than driving the path a
user takes is FAILED, not passed.

**ROWS 3 AND 8 ARE MODULE-LEVEL PARTIALS, NOT PASSES.** Rows 1, 2, 4, 5, 6 and 7 are driven in a mounted
designer and pass. Row 3 (unmapped licence token) is exercised only by `font-licence.test.ts` calling
`classifyLicenceToken` directly; no mounted test drives that refusal along the install path. Row 8 (bytes
that are not a font) is exercised only by the tie test calling the predicate directly. Both are recorded
as partial and must not be upgraded to passes by a later reader counting eight green rows.

**The Spec Change Log's superseding note.** Two claims in the log's first entry assert behaviours that
were reversed within hours. The log is append-only, so the originals stand untouched and a superseding
note was appended above them naming both reversals and pointing each at the dated `2026-09-03 —
orchestrator rulings applied after the step-03 matrix audit` entry, items 4 and 3 respectively. Note
for precision: the two reversed claims are two *paragraphs of one entry*, not two entries, and that
first entry carries no date of its own — the note says so explicitly and identifies it by its heading
text. Both superseding rulings are dated. Both reversals are now covered by tests that did not exist
when the original entry was written.

**Measured gates, re-run at close on `b0f9a74`,** every exit code taken from `$?` immediately after the
command with no pipe or wrapper between:

| gate | result |
|---|---|
| `npm run typecheck` | `rc=0`, clean |
| `npm run lint` | `rc=0`, exactly **4** `only-export-components` warnings, re-measured at `preview/pdf-viewer.tsx:16,17` and `App.tsx:2141,2148` |
| `npm run test` | `rc=1` — **672 passed / 1 failed of 673, 55 files**. Failing **name set** is the single standing `canvas-authority-contract` red (DW-152), matched by name and not by count |
| `npm run build` | `rc=0`, node v24.16.0 |
| `npm run test:e2e:compile` | `rc=0`. **A compile, not a run** |
| `folio-go` `go test -count=1 ./...` | `rc=1` — failing leaf **name set** exactly `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`, of 13 ok packages, 1 failing package (`internal/text`), 4 with no test files |
| `lint` `go test -count=1 ./...` | `rc=0`, 4 packages |

**Standing-rule sweeps re-measured at close, not cited.** Identifier sweep, reported per identifier:
**37 distinct `DW-\d+` cited across the 11 epic-16 artifacts, 37 resolved, 0 unresolved**; second
population **547 tracked files** under `folio-designer/src` and `folio-go` (the build's 503 was
`.go`+`.ts`+`.tsx`+`.json`; the wider set was swept), **45 distinct cited, 45 resolved, 0 unresolved**.
Union across both populations: 76 distinct, 0 unresolved. Register holds **182 definition lines, 182
distinct**. No citation was verified by re-reading a citation.

**Not CI-verified — nothing in this epic may be described as such.** The designer CI job still halts at
step 2 until the post-16.4 repair (DW-171). No browser run was performed at close; the two real browser
runs recorded above are the builder's, and are labelled as the builder's.

**Deferred, with owners.** DW-181 and DW-182 are this story's output; DW-180 was filed late by the
orchestrator against its own dangling citation. The `font-embed-boundary.spec.ts` browser red and the
README e2e-gate claim both belong with **DW-171's post-16.4 CI repair**. The 30-second dropdown-block
hazard and the listbox `role="presentation"` defect (a second sighting of DW-141's) are **gated to Story
16.4**, which is deliberately held at `status: draft` because its middle group asserts what this story
made false.

Story commit `f4d401c` (22 files, 2,659 insertions, 384 deletions). Shared-register commit `b0f9a74`.
Closed by this entry.
