---
title: 'Story 8.4h: The licence gate closes'
type: 'bugfix'
created: '2026-09-02'
status: 'ready-for-dev'
baseline_revision: '7aa283b2da688ab556307cedb1f5543683d8575c'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized', 'multiple-goals']
deferred: []
---

# Story 8.4h: The licence gate closes

**Epic:** 8 — A template author can choose a font, and the file carries it
**Story key:** `8-4h-the-licence-gate-closes`
**Status:** `ready-for-dev`
**Covers:** no FR/NFR. **This story does not come from `epics.md`.** **It was created at Story 8.5's
plan gate on 2026-09-02 by the engineering lead's ruling D-8.5.13, on the precedent of D-8.5.1(a)** —
the same move that inserted Story 8.4f ahead of this same story, in the same `8.4x` insertion series.
`epics.md` is **not** amended by this story and must not be: there is no clause in it this story
implements, weakens or corrects — the clause it enforces is already written there, and the defect is
that the code does not honour it.
**Primary invariant:** **AD-26** (the licence boundary), read with **D-8.5.3**'s owner decision that
extends its enforcement posture — fail the build, never warn — to redistributed **font assets**.
**Adjacent invariants:** AD-21 (byte identity), AD-22 (pinned toolchain), AD-14 (located diagnostics).
**Governing rulings:** **D-8.5.13 (the charter)** · **D-8.5.3** (the owner's four-id allowlist) ·
**D-8.5.2** (unclassifiable fails; the extension-class guard goes repo-wide) · **D-8.5.4** (the
quantifier lesson) · **D-8.5.8(b)** (when a proof will not fail for the right reason, move the proof,
never weaken the claim) · **D-2.1.3** (a loud-miss allowlist is a fail-safe; a silent-miss list is a
rotting list) · **D-1.3.8** (a silent pass on an unidentifiable licence is the realistic failure mode) ·
**D-1.8.11** (never widen `fontExtensions` — its miss is silent) · **D-000.4** (per-epic heavy cadence).

## In plain terms (read this first if you just want the gist)

*Non-normative. This section is a plain-language orientation for a human reader; the intent contract
below is what governs the build.*

Folio ships a check that is supposed to make sure every typeface the project redistributes carries an
acceptable licence. It does not actually do that. It confirms the paperwork is present — a licence
file, a notice, a copyright line — then writes down whatever it believes the licence to be and passes.
If it cannot work the licence out, it writes down "see the notice" and passes anyway. If it *can* work
it out and the answer is one the project must refuse, it writes that down and still passes. So an
unacceptable typeface would ship today on a clean build, with nothing said.

This story makes not knowing count as failing, and knowing an unacceptable answer count as failing
too. A typeface must carry one of the four licences the owner named; anything else stops the build
with a message naming the folder and the reason. The proof that it stops is written against invented
examples in a new empty folder rather than against the typefaces already here, so it keeps holding
when new ones arrive.

The Thai dictionary the software also ships is a separate matter. It is legitimately under a different
kind of licence than a typeface, and it is checked against its own list rather than the typefaces'.
Closing it against the typefaces' rules would reject something perfectly good on day one. Same phrase,
two different rooms.

No new typeface ships here. The catalogue is the next story.

<intent-contract>

## Intent

**Problem:** The asset licence gate **records** a licence instead of **enforcing** one. At
`7aa283b`, both asset paths call the classifier, keep only its SPDX string, discard its family verdict,
and fall through to the label `"SEE NOTICE"` when nothing classifies — then return no error either way.
Measured directly: a GPL-3.0 licence text classifies as `(copyleft, "GPL-3.0")` and still produces a
clean manifest row and a clean build, and an unclassifiable text produces a `"SEE NOTICE"` row and a
clean build. Two distinct holes at the same two lines. Separately, the owner's four-id allowlist has a
member the code does not implement, and the guard that reports a font the gate cannot even *see* looks
at one directory while the gate itself walks the whole repository.

**Approach:** Close both fall-through sites so that neither an unclassifiable licence nor a
classifiable-but-unpermitted one can produce a row, each failing with a located error in the existing
voice. Close them **against different lists**, because they are two populations under two policies.
Add the allowlist's missing fourth member to the classifier loudly — map entry, marker branch, tests —
rather than by widening a list to make a population pass. Widen the extension-class guard's directory
reach to the tracked repository population, and state by measurement what that newly picks up.

## Boundaries & Constraints

**Always:**

- **Both fall-through sites close. Neither keeps `"SEE NOTICE"` as a pass**, and no row anywhere in a
  generated manifest may read `"SEE NOTICE"` after this story.
- **THE TWO SITES CLOSE AGAINST DIFFERENT LISTS. This is the story's shape and must not be collapsed.**
  The **font** site closes against the owner's four-id **asset allowlist** (`OFL-1.1`, `Apache-2.0`,
  `MIT`, and the Ubuntu Font Licence — see Design Note 3 for its identifier). The **wordlist** site
  closes against the **existing permissive SPDX set**, which already carries `CC0-1.0` deliberately.
  One string literal, two consumers, two policies: **validate by consumer, not by key location.** The
  build must **not** collapse them into one shared constant because the strings match.
- Both refusals fail with a **located** error in the established voice: the located subject, then a
  lowercase statement of what is wrong, then a parenthesised citation — matching the three existing
  refusals in each function. The font-site message names the **directory** and the **reason** (which
  licence was classified, or that none could be).
- The classifier's **family verdict is consulted, not discarded.** A text classifying as a copyleft
  family must be refused by name, not merely by absence from the allowlist.
- The Ubuntu Font Licence is added **loudly**: a permissive SPDX map entry, a marker branch placed so
  it cannot be shadowed, a dated story-numbered justification comment in the established house style,
  and tests. **Never by widening a list silently.**
- **THE RED-PROOF IS POPULATION-INDEPENDENT.** Required: a synthetic font in a **new,
  otherwise-empty directory** carrying an **unclassifiable** licence, **and** a second carrying a
  **classifiable-but-not-allowlisted** one (`GPL-3.0`). **Each must red on its OWN message**, asserted
  by a distinctive substring of that message, never on a neighbouring guard. Proved this way, any real
  face in any directory is covered by construction (D-8.5.4, D-8.5.8b).
- **The current committed population must be demonstrated to classify cleanly IN THE SAME COMMIT that
  makes classification fatal.** A gate that lands red is not a gate. The population is **eleven** font
  directories plus the wordlist — see Design Note 2; the charter's "nine, all OFL-1.1" is understated
  and is corrected there by measurement.
- **AC7 widens the guard's DIRECTORY reach, not its ASSET-CLASS reach.** The widened population must
  **enumerate what it newly picks up** and confirm no non-font asset is newly subjected to the font
  allowlist. Do not assume the widening is inert — Design Note 5 measures it.
- A widened population that comes back **empty** must **throw**, never yield an empty set. A guard
  that cannot look must not read as all-clear (D-3.6.5's own ground, and this file's stated doctrine).
- **Byte identity.** The 23 golden PDF digests must be byte-identical to the pre-dispatch snapshot,
  `README.md`'s md5 unchanged, and `maximumCacheAssets` still 64.
- Commit only on `main`. **Never push, never create a branch, never `git add -A`.**

**Block If:**

- **Any asset in the current committed population fails the new gate.** Halt with the directory and
  the classified licence. Do **not** weaken the gate, do **not** add an exemption, do **not** add an
  identifier to the allowlist to make it pass — that is the DW-23/DW-86 standing-red shape,
  manufactured by the very story that exists to stop silent passes.
- **A required red-proof will not fail on its own message** — it keeps failing on a neighbouring
  guard. Halt. Move the proof; **never weaken the claim** (D-8.5.8b).
- Adding `CC0-1.0` (or any non-font identifier) to the four-id **asset allowlist** appears necessary.
  That would amend an owner decision to fix a scoping error. Halt instead.
- **Any golden PDF digest moves.** A moved golden is a HALT, not a re-record.
- The widened AC7 population newly subjects a **non-font** asset to the font allowlist, or newly
  reports a file that is not a font binary.
- The correct SPDX identifier for the fourth allowlist member cannot be established, or establishing
  it would require accepting a non-SPDX alias as a live map key.
- Any change to the `.folio` format, to the engine, or to `folio-go/fonts/` appears necessary.

**Never:**

- Never add `CC0-1.0` to the four-id asset allowlist.
- Never introduce a shared `"SEE NOTICE"` constant, a shared list, or a shared helper that makes the
  two sites enforce one policy.
- Never widen `fontExtensions` — that list's miss is **silent**, the opposite failure mode from the
  permissive allowlist's loud miss (D-1.8.11 against D-2.1.3; the two look alike in a diff).
- Never add a cardinality assertion over the permissive SPDX set. A count written next to the thing it
  counts stops being true the moment the thing grows (D-8.5.4); the loudness comes from the set's
  fail-safe miss semantics, not from counting it.
- Never ship a catalogue face, a `@font-face` rule, picking behaviour, a byte threshold, or anything
  bold, italic, oblique, variable or sloped (Epic 11 owns that, D-000.7).
- Never edit `epics.md` or `ARCHITECTURE-SPINE.md`.
- Never mark the story done on a lint suite run without `-count=1` — these rules walk directories and
  Go's test cache does not track `ReadDir`. A cached `ok` is not a measurement.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Permitted font | A tracked directory with a font binary, a licence text classifying to an allowlisted id, a notice and a copyright line | A manifest row carrying that SPDX id | No error expected |
| Unclassifiable font | Same, but the licence text classifies to nothing | **Refused**, no row | Located error naming the directory and that the licence text could not be classified |
| Forbidden-family font | Same, but the licence text classifies as `GPL-3.0` | **Refused**, no row | Located error naming the directory and the classified identifier |
| Permitted-but-off-list font | Same, but classifies to a permissive id outside the four (e.g. `ISC`) | **Refused**, no row | Located error naming the directory and the classified identifier |
| Wordlist, CC0 | The declared wordlist location with its CC0 legal code | A manifest row carrying `CC0-1.0` — **not** refused | No error expected |
| Wordlist, unclassifiable | Same location, licence text classifying to nothing | **Refused**, no row | Located error naming the wordlist location and the reason, in that function's own citation voice |
| Font-magic file, unrecognised extension, any tracked directory | A tracked file whose first bytes are a font magic and whose extension is outside the recognised set | **Reported** by the extension-class guard | Guard fails naming the file, repo-root-relative |
| Widened population unreadable | The tracked-file enumeration returns nothing | **Throws** | Explicit failure; an empty population must never read as all-clear |

</intent-contract>

## Code Map

**Every anchor measured at `7aa283b2da688ab556307cedb1f5543683d8575c`, this dispatch's baseline.**
Line anchors are navigation aids; assertions are named by what they assert.

### The two fall-through sites — the story's centre

- `lint/internal/manifest/manifest.go`
  - **`:298-301` — SITE A, the FONT path**, inside `ResolveAssets` (`:165`):
    `licenceLabel := "SEE NOTICE"` / `if _, spdx := licence.ClassifyLicenceText(licenceText); spdx != "" { licenceLabel = spdx }`.
    **The `Family` return is discarded (`_, spdx :=`).** Governed by the **owner's four-id asset
    allowlist**.
  - **`:384-387` — SITE B, the THAI WORDLIST path**, inside `resolveWordlistAssetRow` (`:364`).
    Byte-identical shape, **different policy**: governed by the **existing permissive SPDX set**.
  - `:287` / `:290` / `:295` — the three existing located refusals whose **voice the new errors must
    match**, verbatim: `"%s: contains a committed font binary but no LICENSE* file (AC25, AD-26)"`,
    `"%s: contains a committed font binary but no NOTICE* file naming its copyright line (AC25, AD-26)"`,
    `"%s: NOTICE file does not contain a line starting with \"Copyright\" (AC25, AD-26)"`. The
    convention: located subject, `": "`, lowercase statement, parenthesised citation, no trailing period.
  - `:372` / `:376` / `:381` — the same three in the wordlist voice, keyed on the constant
    `wordlistAssetDir` (`:~355`, value `folio-go/internal/text/wordlist`) and citing `(AC9, AD-26)` /
    `(AC25-equivalent for AC9, AD-26)`.
  - `:329` — the D-3.6.5 scan-error floor. **The closest precedent for a NEW class of error** in this
    function: `"scan error: "` prefix, em-dash clause, citation suffix.
  - `:110` — `var fontExtensions = []string{".ttf", ".otf", ".ttc"}`. **Read-only. Never widen** (D-1.8.11).
  - `:169-204` — the walk: whole repo, skipping `.git` and any `lint` dir whose parent is `testdata`.
  - `:242-251` — the **git-tracked filter**: `gitTrackedFileCount` (`:474`, `git -C <root> ls-files -- <dir>`),
    then `if tracked == 0 { continue }`. **This runs BEFORE `os.ReadDir` and before any classification** —
    an untracked-font-only directory contributes neither a row nor an error. **The wordlist path
    (`:364`) has no tracked filter at all**, gated only on `os.Stat` of `words_th.txt` (`:366`).

### The classifier

- `lint/internal/licence/classify.go`
  - `:96-106` `permissiveSPDX` — **unexported**, so the manifest package cannot read it directly; a
    predicate must be exported rather than a second copy of the list made. Current members: MIT,
    Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, 0BSD, Unlicense, CC0-1.0, MIT-0, BlueOak-1.0.0,
    CC-BY-4.0, OFL-1.1. **No Ubuntu Font Licence entry** — confirmed by probe.
  - `:81-83` — the comment recording **CC0 joining at Story 2.1 / D-2.1.3 as the first CC0 asset.**
    This is the evidence that the wordlist's licence is deliberate, not incidental.
  - `:85-95` — D-2.1.3's binding general rule: **a loud-miss allowlist is a fail-safe; a silent-miss
    list is a rotting list.** This is the authority for adding an entry as ordinary maintenance, and
    the authority against widening `fontExtensions` instead.
  - `:121-187` `ClassifyLicenceText` — the ordered marker `switch`. `:129-136` the four copyleft
    branches (`AGPL-3.0`, `SSPL-1.0`, `LGPL-3.0`, `GPL-3.0`). `:137-162` the OFL-1.1 branch, placed
    **above** MIT precisely because of the shared grant clause, with the required-version conjunct
    argued in full at `:148-161` — **the exact template the new branch must follow.**
    `:163` the MIT branch, matching `"PERMISSION IS HEREBY GRANTED, FREE OF CHARGE"`.
  - `:189-199` `classifyBySPDX` — exact map lookup; the copyleft prefix list is `:108`.
- `lint/internal/licence/classify_test.go`
  - `:12` `TestClassifyLicenceText` — broad table, `{name, text, wantFamily}`, **family only**.
  - `:62` `TestClassifyOFL` — the **per-licence** table, `{name, text, wantFamily, wantID}`, asserting
    the id as well. **This is the template for the new licence's table**, including `:92-101`'s
    over-match negative ("OFL 1.0 must NOT be labelled OFL-1.1").
  - `:128` `TestCommittedOFLTextClassifiesAsOFL11` — reads the real committed artifact with an
    explicit emptiness guard (`"is empty — this assertion would be vacuous"`). **No analogue is
    possible for the new licence: this story ships no asset under it** (Design Note 4).

### The tests that will move, and the one that will red

- `lint/internal/manifest/manifest_test.go`
  - `:12` `repoRootFromTest`, `:119` `initGitRepo` (`git init -q` + user config), `:134` `gitAdd`.
    **The scratch-repo recipe the new red-proofs must follow**: `t.TempDir()` → `initGitRepo` →
    write a fake `.ttf` (`[]byte("not a real font, just bytes")` — the walk keys on extension only) →
    write `LICENSE.txt` and `NOTICE` (`"Copyright 2026 Test\n"`) → `gitAdd`. **`git add` only, no
    commit** — `git ls-files` reads the index.
  - **`:183` — `os.WriteFile(licensePath, []byte("a licence"), 0o644)`. MEASURED: `"a licence"`
    classifies to `(unknown, "")`. `TestResolveAssetsExcludesUntrackedDirectoryWithoutError` (`:161`)
    WILL GO RED** the moment unclassifiable becomes fatal, because its *tracked* fixture carries it.
    Its fixture licence text must be upgraded to a classifiable, allowlisted one in the same change.
    Its subject is the untracked-directory exclusion, not licence classification, so this is
    repairing a fixture, not weakening a claim.
  - `:226` `TestResolveAssetsStillReportsATrackedViolation`, `:272` and `:309` the two scan-error
    tests — all three return before classification and are **unaffected**. Confirm, do not assume.
  - `:41` `TestManifestUpToDate` — byte-compares the committed `lint/MANIFEST.md`. Reds if any label
    changes. `:93` `TestResolveAssetsIncludesWordlist` — asserts the wordlist row's `Licence == "CC0-1.0"`;
    **this is the existing tie proving Site B must stay permissive.**
- `lint/internal/rules/fontsassets_test.go`
  - `:283` `TestFontsAssetsNoticeRemovalRedProof`, `:322` `TestFontsAssetsLicenceRemovalRedProof` —
    mutate the **real** repo file and restore via `t.Cleanup` + `os.WriteFile` (explicitly never
    `git checkout`), with a `len(original) == 0` sanity guard at `:330`. **Read-only reference for
    how this module red-proves against the live tree; the new proofs must NOT use this shape** — they
    must be synthetic and population-independent.
  - `:102` `minimalSfnt()` — 12 bytes with the `00 01 00 00` sfnt magic. Reusable.

### The extension-class guard (AC7)

- `folio-designer/src/font-binary-identity.test.ts` (**1144 lines**)
  - `:67` `const designerFontsDir = path.join(designerRoot, 'public', 'fonts')` — **the population
    constant AC7 replaces.** Its intent is stated as a population claim at `:64-66`.
  - `:216-222` `filesUnder(directory)` — recursive `fs.readdirSync`. `:208-214` `looksLikeAFontBinary`;
    magics at `:206` (`\x00\x01\x00\x00`, `true`, `ttcf`, `OTTO`, `wOFF`, `wOF2`).
  - `:225-233` `committedFontsTheLicenceGateCannotSee(directory, recognised)` — **the guard proper**:
    files whose extension is *not* recognised **and** which carry font magic. `:232` renders paths
    with `path.relative(designerRoot, file)`; repo-wide input needs a **repo-root-relative** rendering.
  - `:616` the `it(...)` carrying **two independent population claims**; `:623` the
    `['.ttf','.otf','.ttc']` mirror — **read out of `manifest.go` source** by `licenceGateFontExtensions`
    (`:382-386`, which throws if the declaration is absent), never restated. `:625-632` the generator-slot
    claim — **not touched by AC7**. `:635` the sweep call — **the line AC7 changes.**
  - `:646-682` the **discrimination proof**. It drives the sweep **through a directory argument**
    (`:668`, `:673`), writing a `wOF2`-headed `.woff2` into a scratch dir and asserting it is reported,
    then renaming to `.ttf` and asserting it is not. **If the widened sweep stops taking a directory,
    this proof goes vacuous and must be rewritten, not deleted** (D-8.5.8b).
  - `:100-124` `withoutComments` — the file's own comment-stripping reader. `:44-50` its header
    doctrine: read tracked sources, not build outputs, "because a missing file is the classic way a
    guard goes quietly vacuous."
  - **Latent defect to handle when widening:** `:184` `extensionOf` is
    `file.slice(file.lastIndexOf('.')).toLowerCase()`; for an **extensionless** file `lastIndexOf` is
    `-1`, so it returns the file's **last character**. Under `public/fonts` no extensionless file
    exists; repo-wide, many do. It cannot produce a false positive (the magic check still gates) but
    it would print nonsense in a failure message.
  - Prose that becomes wrong on widening: `:64-66`, `:227`, `:400-405`, `:611-615`.
- **Precedent for git-tracked enumeration:** none in the designer test suite. Nearest idioms:
  `lint/internal/manifest/manifest.go:476` (`exec.Command("git", "-C", repoRoot, "ls-files", ...)`),
  and `folio-designer/scripts/verify-offline-release.mjs:187` (`execFileSync('git', ['status','--porcelain'], { cwd: join(root,'..') })`)
  — precedent that a designer-side Node process may shell to git with the repo root as cwd and must
  fail loudly if git errors. Vitest runs under `jsdom` with cwd `folio-designer/`, so any git call
  **must** pass `-C <repoRoot>` or `cwd`.

### Read-only evidence

- `lint/MANIFEST.md:290-301` — the 12 committed asset rows. **Every one already carries a real SPDX
  id** (`OFL-1.1` ×10, `Apache-2.0` ×1, `CC0-1.0` ×1). `"SEE NOTICE"` appears in **no** committed
  manifest row, **no** test, **no** golden — only at the two production sites. **Closing the
  fall-through therefore changes no committed manifest bytes.**
- `lint/cmd/genmanifest/main.go:31` — the **sole** production caller of `ResolveAssets`; regenerate
  with `cd lint && go run ./cmd/genmanifest`.
- `.github/workflows/ci.yml:180-184` — CI runs `cd lint && go test -count=1 ./...`; `:199-209` runs
  the designer's `test`, `typecheck`, `lint`, `build`. **Both gates this story touches genuinely execute.**
- `lint/testdata/licence/{permissive,copyleft,unknown}/` — 26 files; every fixture `LICENSE` is a
  **single SPDX marker line**, never full text (AC30). Consumed only by
  `lint/internal/rules/licencegraph_test.go:124`. Its permissive subtest asserts **zero findings**
  (`:147-149`), so adding a stub module there is safe and cannot silently pass.
- `_bmad-output/specs/spec-fonts/SPEC.md:130-145` — the settled catalogue Open Question naming the
  four-id allowlist. **Read-only; this story does not amend it** (D-000.10 already propagated it).
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md:504`
  — AD-26, whose **heading is not its Rule**. **Read-only.**

## Tasks & Acceptance

**Execution:**

1. `lint/internal/licence/classify.go` — add the Ubuntu Font Licence 1.0 as `Ubuntu-font-1.0`:
   the `permissiveSPDX` entry **and** a marker branch, with a dated, story-numbered "measured"
   justification comment in the established CC0/OFL house style. **The branch must be placed after the
   copyleft cases and strictly ABOVE the MIT case at `:163`** — the UFL text carries the identical
   `"Permission is hereby granted, free of charge"` grant clause (Design Note 4), so below MIT it is
   shadowed and every UFL face is silently mislabelled `MIT`. Use a required-version conjunct on the
   licence's own title line (`"UBUNTU FONT LICENCE"` and `"VERSION 1.0"`), on the OFL precedent's
   stated reasoning, and **state in the comment** the deliberate decision about the British spelling
   `LICENCE` (a mis-spelled variant misses the marker and therefore fails **loudly**, which is
   fail-safe — record the choice rather than loosening it later). Both halves are required and neither
   substitutes for the other: the marker branch returns its id directly and never consults the map, so
   a map entry alone leaves full text misclassifying as MIT, and a branch alone leaves
   `SPDX-License-Identifier: Ubuntu-font-1.0` classifying as unknown.
2. `lint/internal/licence/classify.go` — export a predicate over the **same** `permissiveSPDX` map
   (e.g. `IsPermissiveSPDX(id string) bool`) so the manifest package can consult it **without a second
   copy of the list**. One declaration, one list; a duplicated list is a guard the code can move.
3. `lint/internal/licence/classify_test.go` — add a per-licence table on `TestClassifyOFL`'s shape
   (`{name, text, wantFamily, wantID}`) covering, at minimum: the full UFL text classifying to
   `(FamilyPermissive, "Ubuntu-font-1.0")`; **the MIT-collision proof** — the same text asserted NOT to
   classify as `MIT`, which is the whole reason the branch exists; a version-lookalike negative
   (`"UBUNTU FONT LICENCE\nVersion 1.1\n"` → `FamilyUnknown, ""`); a bundled-notice negative (MIT text
   merely *mentioning* the Ubuntu Font Licence still classifies `MIT`); and the SPDX-line path
   (`SPDX-License-Identifier: Ubuntu-font-1.0` → permissive). **Red-prove:** removing the marker branch
   must redden a **named** test on its own message, and removing the map entry must redden a different
   named test. Two independent mutations, two distinct reds. **Add no cardinality assertion.**
4. `lint/internal/licence/` (fixture) — add the SPDX-line fixture module
   `lint/testdata/licence/permissive/example.test/ufl-lib/{go.mod,LICENSE,NOTICE.md}` with `LICENSE`
   exactly `SPDX-License-Identifier: Ubuntu-font-1.0`, plus the `require`/`replace` edits to
   `lint/testdata/licence/permissive/go.mod`, so the identifier is exercised end-to-end through the
   graph scan and not only through a table. The permissive subtest asserts **zero** findings, so a
   wrong identifier reds loudly. **Rationale:** "add it loudly" means the id is proved live, not
   asserted in one place.
5. **`lint/internal/manifest/manifest.go:298-301` — SITE A, the FONT path. This is its own task with
   its own red-proof and it must not be merged with task 6.** Replace the fall-through with a
   fail-closed check against the **four-id asset allowlist** — `OFL-1.1`, `Apache-2.0`, `MIT`,
   `Ubuntu-font-1.0` — declared as a named, commented constant **local to the font path** and
   documented as the owner's decision (D-8.5.3). **Consult the classifier's `Family` return, which is
   discarded today**, so a copyleft text is refused by name. Three outcomes, all located, in the
   `:287`/`:290`/`:295` voice, naming the directory and the reason: unclassifiable → refuse;
   classified but outside the allowlist → refuse, **naming the classified identifier**; allowlisted →
   the row, as today. The three refusal messages must be **distinguishable by substring** so a
   red-proof can assert its own.
6. **`lint/internal/manifest/manifest.go:384-387` — SITE B, the THAI WORDLIST path. Its own task, its
   own red-proof, its own list.** Replace the fall-through with a fail-closed check against the
   **existing permissive SPDX set** via task 2's predicate — **not** the four-id allowlist. The
   located error uses that function's own citation voice (`(AC9, AD-26)`) and names the wordlist
   location. **Add a comment at the site stating, in one sentence, that this site's policy is
   deliberately different from the font site's and why** (the owner ruled about *fonts*; `CC0-1.0` is
   not one of the four ids and the dictionary is not a font), so the next reader cannot "tidy" the two
   into one constant. **Do not add `CC0-1.0` to the four-id list. Do not share a constant between the
   two sites.**
7. `lint/internal/manifest/manifest_test.go` — **the population-independent red-proofs (D-8.5.13's
   hard condition).** Using the `initGitRepo`/`gitAdd` scratch recipe, add tests each building a
   **new, otherwise-empty** directory in a fresh temp repo:
   (a) a font with an **unclassifiable** licence text → error, asserted on the *unclassifiable*
   message's own substring;
   (b) a font with a **`GPL-3.0`** licence text → error, asserted on the *not-allowlisted / copyleft*
   message's own substring, **not** on (a)'s;
   (c) a font with a permissive-but-**off-allowlist** licence (e.g. `ISC`) → error — this is the arm
   that proves the check is against the four ids and not merely against "permissive";
   (d) a font with an allowlisted licence → **no** error, and a row bearing that id;
   (e) the **wordlist** site, separately: a scratch wordlist location with a `CC0-1.0` text →
   **no** error and a `CC0-1.0` row; the same location with an unclassifiable text → error.
   **(e) is the assertion that discharges the two-populations trap: it fails if a future change
   collapses Site B onto the font allowlist.** Each test must be verified to fail on **its own**
   message; if any of them reds on a neighbouring guard instead, **HALT** — move the proof, never
   weaken the claim.
8. `lint/internal/manifest/manifest_test.go:183` — upgrade
   `TestResolveAssetsExcludesUntrackedDirectoryWithoutError`'s **tracked** fixture licence from
   `"a licence"` (measured: classifies to nothing) to a classifiable, allowlisted marker, so the test
   keeps testing the untracked-directory exclusion rather than the new refusal. **Confirm by running
   that the other three synthetic tests (`:226`, `:272`, `:309`) return before classification and are
   genuinely unaffected — do not assume it.**
9. `lint/internal/manifest/manifest_test.go` — add the **committed-population witness**: assert that
   `ResolveAssets` over the **real** repo root returns **no error** and that **no** returned row's
   licence is `"SEE NOTICE"`, and that every font row's licence is one of the four ids. This is the
   assertion that makes "the gate does not land red" a fact the build re-checks rather than a claim
   made once at implementation time.
10. `folio-designer/src/font-binary-identity.test.ts` — **AC7.** Widen the extension-class guard's
    population from `:67`'s `public/fonts` to the **repo-wide tracked** population: the licence gate's
    own walk (repo root, skipping `.git` and any `*/testdata/lint` directory) **intersected with
    `git ls-files`**, using the `-C <repoRoot>` idiom. **The enumeration must throw if git fails or
    returns an empty list** — an empty population must never read as all-clear. Keep `:623`'s exact
    `['.ttf','.otf','.ttc']` mirror (still **read from `manifest.go` source**, not restated) and keep
    `:625-632`'s generator-slot claim untouched. Render reported paths **repo-root-relative**. Fix
    `:184` `extensionOf` for extensionless files. Update the now-wrong prose at `:64-66`, `:227`,
    `:400-405`, `:611-615` to state the new population and its exclusions.
11. `folio-designer/src/font-binary-identity.test.ts:646-682` — **preserve the discrimination proof
    non-vacuously.** It drives the sweep through a *directory* argument; if the widened sweep no
    longer takes one, the proof must be **rewritten to exercise the widened population** (e.g. a
    tracked scratch path, or a seam that lets the tracked list be injected), **extended to prove the
    new directory reach is real** — a font-magic file in a directory outside `public/fonts` is now
    reported, and was not before. **Deleting or narrowing this proof is a HALT.**
12. `_bmad-output/implementation-artifacts/deferred-work.md` — record the **measured enumeration** AC7
    requires: what the widened walk newly covers, and that it newly reports **nothing** today
    (Design Note 5), so a later reader can tell an inert widening from an unmeasured one. Record also
    that this story adds an allowlist member with **no asset in the tree under it** and that its first
    real-population witness arrives with Story 8.5.

**Acceptance Criteria:**

**Numbering note, so a reviewer routing on AC numbers is not misled.** These are **this story's**
criteria and they renumber what arrived from Story 8.5. The mapping: 8.5's **AC2** (licence,
fail-closed) becomes **AC1 + AC2 + AC3 + AC4** here — split because Design Note 1 measures it as two
distinct holes over two populations, and a single criterion could be satisfied by closing one of them.
8.5's **AC7** (gate visibility) is **AC7** here, deliberately keeping its number. **AC5, AC6, AC8 and
AC9 are new to this story** and have no 8.5 antecedent. Nothing here is 8.5's AC1, AC3, AC4, AC5, AC6
or AC8; those stay in 8.5 (D-8.5.13).

- **AC1 (font site, fail-closed).** Given a committed, tracked font directory whose licence text
  classifies outside `{OFL-1.1, Apache-2.0, MIT, Ubuntu-font-1.0}` **or** cannot be classified at all,
  when the lint suite runs, then it **fails** with a located error naming the directory and the
  reason, and produces no row for it.
- **AC2 (population independence).** Given a **new, otherwise-empty** directory in a scratch
  repository carrying a synthetic font, when its licence is unclassifiable, and separately when it is
  `GPL-3.0`, then **each** case fails on **its own** distinct message — verified by asserting a
  substring unique to that message, not by observing that something failed.
- **AC3 (wordlist site, different list).** Given the Thai wordlist under its declared location with
  its `CC0-1.0` legal code, when the lint suite runs, then it **passes** and produces a `CC0-1.0` row;
  and given the same location with an unclassifiable licence, then it **fails** with a located error
  in that function's own voice. **`CC0-1.0` is not added to the four-id allowlist, and the two sites
  share no list, constant or helper carrying policy.**
- **AC4 (no silent pass survives).** Given any generated manifest after this story, when it is
  inspected, then **no** row reads `SEE NOTICE`, and the string is unreachable as a passing outcome
  at both sites.
- **AC5 (the allowlist's fourth member exists).** Given a licence text that is the Ubuntu Font Licence
  1.0, when it is classified, then it returns `(permissive, "Ubuntu-font-1.0")` and **not** `MIT`;
  given `SPDX-License-Identifier: Ubuntu-font-1.0`, then it classifies permissive; and given the
  marker branch removed, or the map entry removed, then a **named** test reds in each case — two
  mutations, two distinct reds.
- **AC6 (the gate lands green).** Given the current committed population — **eleven** tracked font
  directories plus the wordlist — when the fatal classification lands, then **every one** of them
  resolves without error **in the same commit**, demonstrated by a run and by an assertion that
  re-checks it, not by inspection alone.
- **AC7 (gate visibility, repo-wide).** Given a file with font magic bytes and an unrecognised
  extension committed in **any** tracked directory the licence gate walks, when the extension-class
  guard runs, then it is reported — proved by a fixture outside `folio-designer/public/fonts`; and
  given the tracked enumeration cannot be obtained, then the guard **throws** rather than reporting
  a clean population.
- **AC8 (the widening is measured, not assumed).** Given the widened population, when the story is
  complete, then what it newly covers is **enumerated and recorded**, and it is confirmed that no
  non-font asset is newly subjected to the font allowlist.
- **AC9 (no drift elsewhere).** Given the whole change, when the gates run, then the **23** golden PDF
  digests are byte-identical to the pre-dispatch snapshot, `README.md`'s md5 is unchanged,
  `maximumCacheAssets` is still 64, and no `.folio`, engine or `folio-go/fonts/` file is touched.

## Spec Change Log

*(empty — no `bad_spec` loopback has occurred; this spec is at its first gate.)*

## Review Triage Log

*(empty — no review pass has run.)*

## Design Notes

**1. The hole is TWO holes, and the charter's stated mechanism is the smaller one. Measured, not
inferred.** D-8.5.13 describes the defect as *"a font whose licence text does not classify gets
`licenceLabel = "SEE NOTICE"` and `ResolveAssets` returns nil."* That is true and it is half the story.
Probed directly at `7aa283b` by calling `ClassifyLicenceText` on five inputs:

| input | family | spdx | today's outcome at both asset sites |
|---|---|---|---|
| `GNU GENERAL PUBLIC LICENSE\nVersion 3, 29 June 2007` | **copyleft** | `GPL-3.0` | a clean row labelled `GPL-3.0`, no error |
| `SPDX-License-Identifier: GPL-3.0` | **copyleft** | `GPL-3.0` | a clean row labelled `GPL-3.0`, no error |
| `a licence` | unknown | `""` | a clean row labelled `SEE NOTICE`, no error |
| `SPDX-License-Identifier: UFL-1.0` | unknown | `UFL-1.0` | a clean row labelled `UFL-1.0`, no error |
| `SPDX-License-Identifier: Ufont-1.0` | unknown | `Ufont-1.0` | a clean row labelled `Ufont-1.0`, no error |

**A GPL font never reaches the `"SEE NOTICE"` fall-through at all** — it classifies perfectly well and
passes because **both sites discard the `Family` return** (`_, spdx :=`) and nothing compares the id to
any list. So the charter's headline claim (*"a GPL font ships with a clean build today"*) is **exactly
right**, while the mechanism it attributes it to is the *other* hole. **This changes nothing about the
verdict and nothing about scope** — 8.5's AC2, which D-8.5.13 moves here verbatim, already says *"outside
`{...}` **or** unclassifiable"*, covering both. It changes the **shape of the proof**: the two required
red-proofs are not two flavours of one arm, they are one arm each, and neither substitutes for the
other. Recorded because a reviewer reading only the charter would expect one mechanism and find two.

**2. The committed population is ELEVEN font directories, not nine — and not all OFL-1.1.** D-8.5.13
states *"nine directories, all OFL-1.1"* (three under `folio-go/fonts/`, six under
`folio-designer/public/fonts/`). Measured from `git ls-files` and `lint/MANIFEST.md:290-301`, the
population `ResolveAssets` actually resolves is **twelve rows across eleven font directories plus the
wordlist**:

- 6 × `folio-designer/public/fonts/{ibmplexmono,ibmplexsans,ibmplexsansthai,notosans,notosanssc,notosansthai}` — **OFL-1.1**
- 3 × `folio-go/fonts/{notosans,notosanssc,notosansthai}` — **OFL-1.1**
- `folio-go/testdata/fonts` (`Roboto-Regular.ttf`) — **Apache-2.0**
- `folio-go/testdata/fonts/notosansthai-variable-testonly` (`NotoSansThai-VF.ttf`) — **OFL-1.1**
- `folio-go/internal/text/wordlist/words_th.txt` — **CC0-1.0**

The two extra directories are under `folio-go/testdata/`, which the walk does **not** skip — only
`*/testdata/lint` is skipped. **The verdict is unaffected: `Apache-2.0` is on the owner's four-id
list, so the gate still lands green.** But AC6's obligation is over **eleven**, and a story that
enumerated nine would have left two real directories unwitnessed. This is the same defect class the
run keeps catching: the observation was true, the quantifier attached to it was not (D-8.5.4).
`folio-go/testdata/lint/embed-font/*/fonts/shipped-face.ttf` is correctly out of scope — skipped by
the walk, and measured to be a text stub (`not `), not a font.

**3. "UFL" is not an SPDX identifier. `Ubuntu-font-1.0` is — verified by fetch, not from memory.**
`https://spdx.org/licenses/Ubuntu-font-1.0.html` returns **HTTP 200**, titled *"Ubuntu Font Licence
v1.0"*; `Ufont-1.0` and `UFL-1.0` both return **404**. The owner's allowlist, as recorded in
`epics.md`, `SPEC.md` and D-8.5.3, names its fourth member **"UFL"** — the community abbreviation —
while its other three members (`OFL-1.1`, `Apache-2.0`, `MIT`) are exact SPDX identifiers.

**Ruled here rather than escalated, on the project's own convention.** Every id in `permissiveSPDX`
today is a canonical SPDX identifier, and `classifyBySPDX` is an **exact map lookup**, so a literal
`"UFL"` key could only ever be produced by a licence file writing `SPDX-License-Identifier: UFL` —
which is not a valid SPDX line and which no real font ships. Such an entry would be **dead code that
looks live**: precisely the "advertises detection that does not exist" failure the classifier's own
doc comment was written to prevent. So "UFL" in the owner's list denotes *the Ubuntu Font Licence*,
and the code's obligation is to use its canonical identifier. **The bare alias is deliberately NOT
accepted**: tolerating it would be exactly the silent list-widening D-8.5.13 forbids, and it would
admit a string no licence authority issues. **Flagged for the plan gate** — this resolves a wording
discrepancy in an owner decision, and if the lead reads it as the owner's call rather than the
convention's, it is cheap to reverse here and expensive later.

**4. The new marker branch collides with MIT exactly as OFL did — confirmed against the canonical
text, not assumed.** The SPDX-published UFL 1.0 text opens `UBUNTU FONT LICENCE Version 1.0` and, at
line 48 of that file, carries `Permission is hereby granted, free of charge, to any person obtaining a`
— the **exact substring** the MIT branch at `classify.go:163` matches on. Placed below MIT, every
UFL-licensed face would classify as `(permissive, "MIT")`: right family, **wrong label**, and the
label is what lands in `lint/MANIFEST.md` and attributes the asset. This is character-for-character
the defect documented at `classify.go:138-147` for OFL, which is why task 1 states the placement as a
requirement rather than a preference. Audited against every other branch: no collision with the four
copyleft cases, with OFL (the UFL text contains neither *"SIL OPEN FONT LICENSE"* nor *"VERSION 1.1"*),
with Apache, BSD or CC0.

**5. AC7's widening is INERT TODAY — and that is a measurement, not the assumption the charter warns
against.** D-8.5.13 requires the widening's new coverage to be enumerated, *"because assuming it is
inert is precisely what produced the wordlist trap."* So it was enumerated. Against 1434 tracked
files: **zero** tracked files carry a font-plausible extension other than `.ttf` (no `.woff`,
`.woff2`, `.eot`, `.otc`, `.pfb`, `.dfont`); and reading the **first four bytes of every one of the
~1420 tracked non-`.ttf/.otf/.ttc` files** against the guard's own magic set yields **zero** hits. So
the widened guard newly asserts over ~1420 files and newly reports none. **Two consequences that are
not inert.** First, the `git ls-files` intersection is load-bearing rather than cosmetic: a disk-only
repo-root walk would read the three real variable TTFs in the gitignored `.font-sources/` (≈20 MB) and
everything under `node_modules/`, `dist/` and `src/generated/`. Second, **AC7 widens the guard, not the
licence gate** — the gate's own walk was already repo-wide (`manifest.go:169`) and already filters by
extension, so **no non-font asset is newly subjected to the font allowlist by this story.** The
widening closes the gap D-8.5.2 names: a file the gate cannot see *because of its extension*, in a
directory the guard could not see *because of its path*.

**6. Why the two sites cannot share a constant, stated as a test rather than as a principle.** The
tell that the collapse has happened is not a shared string — it is that Site B starts refusing
`CC0-1.0`. Task 7(e) asserts the wordlist passes with `CC0-1.0` **and** fails when unclassifiable, so
a future change that points Site B at the font allowlist reds on a test whose name says why, rather
than reddening the whole build on a shipped, legitimate, owner-unobjected asset. `manifest_test.go:93`
already asserts the `CC0-1.0` row today; task 7(e) is the arm that makes that assertion *load-bearing
against this specific mistake*.

**7. `permissiveSPDX` is unexported, and the fix must not be a second copy.** The manifest package
cannot read it. Task 2 exports a predicate over the same map rather than restating the list, because a
duplicated list is a list the code can move — the failure D-8.5.8(c) named on the Node/TypeScript
constant. Note the two lists are genuinely independent in scope, not in mechanism: the four-id
**asset allowlist** is a new, font-path-local constant expressing an owner decision; the **permissive
SPDX set** is the existing dependency-side list. They must not be merged, and neither may be derived
from the other.

**8. What this story does NOT prove.** It adds an allowlist member with **no asset in the tree under
it**: nothing in this repository is Ubuntu-licensed, so `Ubuntu-font-1.0`'s classification is proved
only against synthetic inputs and the SPDX-line fixture. There is no analogue of
`TestCommittedOFLTextClassifiesAsOFL11` to write. That is expected — D-8.5.13 scopes the member here
precisely so it exists before 8.5 needs it — but the story must not claim a real-population witness it
does not have. Recorded in task 12 so the first such witness is attributable to 8.5.

## Verification

**Baseline MEASURED at `7aa283b2da688ab556307cedb1f5543683d8575c` before any edit, with the working
tree carrying only the workflow's own `M _bmad-output/implementation-artifacts/epic-8-context.md`
(step-01's regenerated cache — expected, not drift).**

**Commands:**

- `cd lint && go test -count=1 ./...` — **baseline measured: four `ok`, zero FAIL**
  (`cmd/genmanifest`, `internal/licence`, `internal/manifest`, `internal/rules`).
  **`-count=1` always**: these rules walk directories and Go's cache does not track `ReadDir`, so a
  cached `ok` is no measurement at all. **This is the story's primary suite.**
- `cd lint && go vet ./...` — expected: no output.
- `cd lint && go run ./cmd/genmanifest && git diff --stat lint/MANIFEST.md` — expected: **no diff**.
  Every committed row already carries a real SPDX id, so closing the fall-through must change **zero**
  manifest bytes. A diff here means a label moved and is a finding, not a re-record.
- `cd folio-go && go test -count=1 ./...` — **baseline measured: 1815 pass / 2 fail / 5 skip.**
  The two are `TestCorpusMeetsP6ExerciseFloors` and its subtest `P6g_(opaque_names)`
  (`got 7, need >=20`). **The one genuine standing red; never "fix" it. Exactly ONE distinct red by
  identity — any second is a real failure.** No `folio-go` file is touched by this story, so the
  post-change figure must be identical.
- `cd folio-go && go vet ./...` — expected: no output.
- `gofmt -l folio-go lint` — **run from the REPO ROOT**, expected: no output. After a `cd` into a
  module it prints an `lstat` error that reads like success.
- `cd folio-designer && npm test` — **baseline measured: 40 files / 409 tests, all passing.**
  This is the suite that runs the AC7 guard. `npm test` runs `build:wasm` first, so the generated
  assets exist when the guard runs.
- `cd folio-designer && npm run typecheck` — expected: clean.
- `cd folio-designer && npm run lint` — expected **exactly 4** `only-export-components` warnings; the
  count and the rule name are the invariant, not the line numbers.
- `cd folio-designer && npm run build` — expected: succeeds. This is the **build** leg of the cadence.
- `cd folio-designer && npm run test:e2e:compile` — the compile-only check D-000.4 requires. It is
  `tsc --noEmit`; **do not report it as a run.** This story adds no integration/e2e package, so it is
  a no-new-surface confirmation.
- `shasum -a 256 fixtures/*/expected.pdf` — **run from the REPO ROOT** (the directory is `fixtures/`
  at the root, not under `folio-go/`). Expected **23** lines, **byte-identical** to the pre-dispatch
  snapshot; **diff it, do not eyeball.** **A moved golden is a HALT, not a re-record.**
- `md5 -q README.md` — expected `078d7d80d518d54af2fc04fb270d46b8`, unchanged (measured at baseline).
- `grep -rn "maximumCacheAssets" folio-designer/src` — expected: still `64`, untouched.

**NOT in this story's `## Verification`, by D-000.4's per-epic cadence — these run at Epic 8's
boundary gate:** the four `FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity`, and Playwright.
Naming them here is a record of what was deliberately not run, not an omission.

**Manual checks:**

- **Red-prove each new refusal on its OWN message.** For each of the three font-site outcomes and the
  wordlist-site refusal, confirm the test asserts a substring unique to that message. Then confirm by
  mutation: make each refusal's condition unreachable in turn and check that the *matching* named test
  reds and the others stay green. **If any proof reds on a neighbouring guard instead, HALT** —
  move the proof, never weaken the claim (D-8.5.8b).
- **Red-prove the classifier addition twice.** Remove the marker branch → a named test reds. Restore.
  Remove the `permissiveSPDX` entry → a *different* named test reds. Restore with an **absolute**
  path; a `cd` earlier in the same compound command breaks a relative restore and silently leaves the
  mutation in place. Then `git status --porcelain` before continuing.
- **Prove the widened guard is not vacuous.** Confirm the discrimination proof still drives the
  *widened* population and reports a font-magic file placed in a directory outside
  `folio-designer/public/fonts` — and that the same file renamed to `.ttf` is not reported. Then
  confirm the enumeration **throws** when the tracked list cannot be obtained.
- **Enumerate the widening's new coverage and record it** (AC8): what the widened population now
  covers, and what it newly reports. Design Note 5 records the baseline measurement; re-run it after
  the change and report both figures.
- **Confirm the committed population classifies cleanly in the same commit** by running
  `cd lint && go run ./cmd/genmanifest` and the lint suite after the change is complete, and by
  reporting all **eleven** font directories plus the wordlist with their classified identifiers.
- Confirm no `.folio` schema file, no engine source, and nothing under `folio-go/fonts/` appears in
  `git diff --name-only`.

## Auto Run Result

Status: ready-for-dev
Blocking condition: none — `Halt after planning.` was directed, so the run stops at the
ready-for-development gate with the spec written and **no code derived from it**.

Baseline revision: `7aa283b2da688ab556307cedb1f5543683d8575c` (branch `main`).
Commits created this dispatch: none.
Working tree at halt: the spec itself (new, untracked) plus `epic-8-context.md`, which step-01
regenerated because `epics.md` was newer than the cached copy. That modification is the workflow's own
step-01 output, not a change this story made.
