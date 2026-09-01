---
title: 'Story 8.4h: The licence gate closes'
type: 'bugfix'
created: '2026-09-02'
status: 'done'
baseline_revision: '80f46a0d19d4ae092e54658bad62cdb01f33e5a2'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: ['oversized', 'multiple-goals']
deferred:
  - summary: >-
      An MIT licence text that merely mentions "Ubuntu Font Licence" AND "version 1.0" anywhere in
      its body classifies as Ubuntu-font-1.0, because the new marker branch is an unanchored
      substring conjunction placed above the MIT case.
    evidence: |-
      Verified pre-existing as a class, not introduced here: the OFL branch at
      lint/internal/licence/classify.go:181 has the identical unanchored shape
      (`Contains("SIL OPEN FONT LICENSE") && Contains("VERSION 1.1")`) and Story 8.4h's spec
      named that branch as "the exact template the new branch must follow". The bundled-notice
      negative in classify_test.go omits "version 1.0" from its sentence, so it does not reach
      this case. Fixing it means anchoring both conjuncts to one title line, which changes the
      OFL branch too and is therefore its own story.
    location: >-
      lint/internal/licence/classify.go:207
    severity: medium
  - summary: >-
      An asset directory holding more than one LICENSE* file lets the last file in os.ReadDir
      order decide the enforced verdict and the published attribution.
    evidence: |-
      The ReadDir loop overwrites licenceText on each LICENSE* match and pre-dates this story.
      What changed is the consequence: the arbitrarily chosen file's classification is now
      fail-closed policy rather than a "SEE NOTICE" label. Verified via git ls-files that every
      committed asset directory carries exactly one LICENSE* today, so nothing is wrong now and
      no test covers the two-file case.
    location: >-
      lint/internal/manifest/manifest.go:300-315
    severity: medium
  - summary: >-
      insideTheLicenceGateWalk hand-copies the licence gate's two SkipDir rules into TypeScript
      with nothing tying it back to manifest.go, unlike the extension list beside it.
    evidence: |-
      folio-designer/src/font-binary-identity.test.ts parses `var fontExtensions` out of
      manifest.go source precisely so the guard agrees with the gate rather than with a copy of
      the gate, and its header states that doctrine. The walk exclusions (skip .git, skip any
      lint dir whose parent is testdata) are restated as literals instead. The silently harmful
      drift direction is narrow — the Go side narrowing its */testdata/lint skip while the TS
      copy keeps skipping — but nothing compares the two.
    location: >-
      folio-designer/src/font-binary-identity.test.ts:325-329
    severity: low
  - summary: >-
      Nothing pins the membership of the owner's four-id fontAssetLicenceAllowlist, so a fifth
      identifier could be added without any test reddening.
    evidence: |-
      Every test reads INTO the list (fontAssetLicenceAllowed["X"]); none asserts the list's
      exact contents. D-8.5.13 forbids silent list-widening, and this list encodes a fixed owner
      decision (D-8.5.3) rather than a population that grows by design, so D-8.5.4's
      "no cardinality assertion" argument does not cover it. An exact []string equality naming
      the decision would close it.
    location: >-
      lint/internal/manifest/manifest.go:140
    severity: medium
  - summary: >-
      The AC7 throws-proof depends on os.tmpdir() not being inside a git repository; under a
      TMPDIR set within a work tree it could enumerate an ancestor repo instead of throwing.
    evidence: |-
      The assertion is expect(() => licenceGateTrackedFiles(notARepository)).toThrow(...). git
      searches upward for a repository, so on a machine whose TMPDIR sits under a checkout the
      call succeeds and the arm goes vacuous or asserts the wrong thing. Passing
      GIT_CEILING_DIRECTORIES, or clearing GIT_DIR/GIT_WORK_TREE, would make it machine-independent.
    location: >-
      folio-designer/src/font-binary-identity.test.ts
    severity: low
  - summary: >-
      The guard's per-file `git ls-files` intersection is narrower than the gate's per-directory
      tracked filter, so an untracked font in a directory that has any tracked file is seen by
      the gate but not by the guard.
    evidence: |-
      manifest.go:242-251 skips a directory only when gitTrackedFileCount is zero; the guard
      intersects file by file. The comment claims the two exclude untracked files "exactly" the
      same way. All 18 files under folio-designer/public/fonts are tracked, so the widening
      loses nothing today, but the stated equivalence is not accurate.
    location: >-
      folio-designer/src/font-binary-identity.test.ts
    severity: low
  - summary: >-
      looksLikeAFontBinary ignores the readSync return count, so a tracked file shorter than four
      bytes would match the zero-filled 00 01 00 00 sfnt magic and be reported as a font.
    evidence: |-
      Newly reachable because the population went repo-wide; under public/fonts no such file
      existed. Verified that no tracked file in the repository is currently smaller than four
      bytes, so the false positive is latent rather than live. Checking the readSync return
      before comparing would close it.
    location: >-
      folio-designer/src/font-binary-identity.test.ts:208-214
    severity: low
---

# Story 8.4h: The licence gate closes

**Epic:** 8 — A template author can choose a font, and the file carries it
**Story key:** `8-4h-the-licence-gate-closes`
**Status:** `done`
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

Folio's check that every typeface it redistributes carries an acceptable licence did not do that. It
confirmed the paperwork was there, recorded whatever it thought the licence was, and passed either
way. The defect had two halves. If the check could not work a licence out, it wrote "see the notice"
and passed. Worse, and far less obvious: if it *could* work it out and the answer was one the project
must refuse, it recorded that and passed anyway. A forbidden typeface would have shipped on a
clean build, correctly identified, with nothing said.

Both halves are closed: not knowing now fails, and so does knowing an unacceptable answer. A typeface
must carry one of the four licences the owner named; anything else stops the build with a message
naming the folder and the reason. The proof uses invented examples in new empty folders, so it
keeps holding as real typefaces arrive. Nothing already here had to change to pass.

The Thai dictionary the software also ships is a separate matter. It is legitimately under a different
kind of licence than a typeface, and is checked against its own list. Closing it against the typefaces'
rules would reject something perfectly good on day one. Same phrase, two different rooms.

No new typeface ships here; the catalogue is next. Two gaps found at close are recorded, not fixed: one
licence spelled the American way is still quietly accepted under the wrong name, and the owner's
four-name list can still be widened with no test noticing.

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

### 2026-09-02 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 6: (high 0, medium 1, low 5)
- defer: 7: (high 0, medium 3, low 4)
- reject: 7
- addressed_findings:
  - `[medium]` `[patch]` **The wordlist site's third refusal arm shipped unexercised.**
    `case !licence.IsPermissiveSPDX(wordlistSPDX)` is the only arm reachable when
    `ClassifyLicenceText` returns a known id with `FamilyUnknown` — an SPDX line naming an
    identifier that is neither copyleft nor in `permissiveSPDX`. **Confirmed by mutation, not
    asserted:** deleting the whole `case` left all four lint packages green. Added a fourth
    subtest to `TestWordlistSiteEnforcesThePermissiveSetNotTheFontAllowlist` using an
    `SPDX-License-Identifier: CC-BY-SA-4.0` wordlist fixture, asserting that arm's own substring
    (`does not recognise as a permissive licence`) and asserting the other two arms' substrings
    are absent. Re-proved after the fix: deleting the arm now reds that named subtest and
    nothing else.
  - `[low]` `[patch]` **DW-114's measurement was stale by the commit's own fixtures.** The table
    was measured before this story added `lint/testdata/licence/permissive/example.test/ufl-lib/`
    (three files). Re-measured at the committed tree and corrected: `git ls-files` 1435 → **1438**,
    widened population 1371 → **1374**, newly covered 1353 → **1356**, tracked extensionless
    18 → **19** (by `extensionOf`'s own rule, under which a dotfile counts as extensionless).
    `64` excluded by the `*/testdata/lint` skip and `11, all .ttf` were both re-verified correct
    and stand. A record whose stated purpose is "measured, not assumed" must be measured after
    its own fixtures land.
  - `[low]` `[patch]` **The flagship reach test hardcoded the extension list it is supposed to
    mirror.** `it('reports a font the licence gate cannot see from anywhere it walks...')` set
    `const recognised = ['.ttf', '.otf', '.ttc']` literally, while every other site in that file
    derives it via `licenceGateFontExtensions(...)` read out of `manifest.go` source — the file's
    own stated doctrine that the guard must agree with the gate, not with a copy of the gate.
    Now derived the same way.
  - `[low]` `[patch]` **`extensionOf`'s doc comment cited `go.sum` as an extensionless file.** It
    has an extension. Replaced with real examples.
  - `[low]` `[patch]` **The `ufl-lib` fixture overclaimed what it proves.** Both its `NOTICE.md`
    and `lint/testdata/licence/permissive/go.mod` said a wrong identifier "reds loudly".
    `ScanLicenceGraph` switches on licence *family* and discards the id, so any permissive id
    produces the same green. Corrected to state what the fixture actually proves — that the
    `permissiveSPDX` entry is live — and to point at `classify_test.go` as where the id itself
    is pinned.
  - `[low]` `[patch]` The spec body's `**Status:**` line still read `ready-for-dev` after the
    frontmatter advanced. Synchronised.

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
- `gofmt -l folio-go lint` — **run from the REPO ROOT**. After a `cd` into a module it prints an
  `lstat` error that reads like success. **Expected: exactly ONE line,
  `lint/internal/rules/licencegraph_test.go` — a STANDING RED BY IDENTITY, not this story's.**
  The plan-gate line here originally read *"expected: no output"*; that was a mis-measurement,
  corrected by **D-8.4h.5** rather than by reformatting a file outside this story's scope.
  Verified independently at close: the file's content extracted from `7aa283b` is unformatted there
  too, and `git diff --name-only 80f46a0..HEAD` does not list it — the story never touched it.
  **Never "fix" it here; a SECOND file appearing in this output is a real failure.** Registered as
  **DW-116**, owned by whichever story next touches `lint/internal/rules/`.
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

Status: done
Blocking condition: none.

Baseline revision: `80f46a0d19d4ae092e54658bad62cdb01f33e5a2` (branch `main`, tree clean).
**Note on the baseline move.** The spec was planned at `7aa283b`, and every Code Map anchor is
measured there. Two commits landed between: `git diff --name-only 7aa283b..80f46a0` returns three
`_bmad-output/` files and no code, so every anchor in the Code Map still resolves and the difference
is not drift.

### Summary of the implemented change

The asset licence gate **recorded** a licence; it now **enforces** one. Both fall-through sites close
fail-closed, and — the story's shape — **against different lists**, because they are two populations
under two policies. The classifier's `Family` return, discarded at both sites, is now consulted, so a
copyleft text is refused **by name** rather than merely by absence from a list. The owner's four-id
allowlist gained its missing fourth member under its canonical SPDX identifier.

### Files changed

- `lint/internal/licence/classify.go` — `Ubuntu-font-1.0` added to `permissiveSPDX` **and** as a
  marker branch placed strictly above the MIT case (the UFL text carries MIT's grant clause
  verbatim); `IsPermissiveSPDX` exported as a predicate over that same map, so no second copy of the
  list exists.
- `lint/internal/licence/classify_test.go` — the per-licence table, the MIT-collision proof, the
  version-lookalike negative, the bundled-notice negative, the SPDX-line path, and a test pinning the
  exported predicate to the same list.
- `lint/internal/manifest/manifest.go` — **SITE A** (font) closes against a new, font-path-local
  `fontAssetLicenceAllowlist` of exactly the owner's four ids; **SITE B** (Thai wordlist) closes
  against the existing permissive SPDX set via the exported predicate, with a comment at the site
  stating why the policy deliberately differs so the two cannot be "tidied" together.
- `lint/internal/manifest/manifest_test.go` — the population-independent red-proofs in scratch git
  repositories, the two-sites-do-not-share-a-policy tie, the committed-population witness, and the
  repaired `"a licence"` fixture in the untracked-directory test.
- `folio-designer/src/font-binary-identity.test.ts` — AC7: the extension-class guard's population
  widened from `public/fonts` to the git-tracked repository the gate itself walks, throwing rather
  than yielding an empty set; `extensionOf` fixed for extensionless files; the discrimination proof
  preserved and extended to prove the new directory reach.
- `lint/testdata/licence/permissive/example.test/ufl-lib/{go.mod,LICENSE,NOTICE.md}` + that tree's
  `go.mod` — the SPDX-line fixture exercising the identifier end-to-end through the graph scan.
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-114 (AC8's measured enumeration) and
  DW-115 (the allowlist member with no asset in the tree under it).

### Review findings breakdown

Patches applied: **6** (1 medium, 5 low) — see the Review Triage Log above.
Items deferred: **7** (3 medium, 4 low) — recorded in frontmatter `deferred`.
Items rejected: **7**. Notable rejections, with reasons: the duplicated case in
`TestClassifyUbuntuFontLicence` (redundant but correct — removing it is churn); the absence of a test
for the American spelling `"UBUNTU FONT LICENSE"` (its miss is **loud** — `FamilyUnknown` is a build
failure at the gate — and D-2.1.3 makes a loud miss fail-safe by design, which the comment records);
and the reading that AC7 belongs in the Go lint suite rather than the designer's (the spec's Code Map
names `folio-designer/src/font-binary-identity.test.ts` as the artifact, settling it).

Follow-up review recommendation: **true**. Patched counts by severity: high 0, medium 1, low 5.
Score = (3 × 1) + (1 × 5) = **8**, which is ≥ 5.

### Verification performed

Every figure below was measured by this run, not taken from the implementation's report.

- `cd lint && go test -count=1 ./...` — **four `ok`, zero FAIL**, exit 0.
- `cd lint && go vet ./...` — no output, exit 0.
- `cd lint && go run ./cmd/genmanifest` then `git diff --stat -- lint/MANIFEST.md` — **no diff**.
  Closing the fall-through moved zero manifest bytes, as predicted.
- `cd folio-go && go test -count=1 -v ./...` — **1815 pass / 2 fail / 5 skip**, byte-for-byte the
  spec's baseline. The two failures are the one standing red **by identity**:
  `TestCorpusMeetsP6ExerciseFloors` and its subtest `P6g_(opaque_names)`. No second distinct red.
- `cd folio-go && go vet ./...` — no output, exit 0.
- `gofmt -l folio-go lint` **from the repo root** — prints `lint/internal/rules/licencegraph_test.go`.
  **Proven pre-existing, not asserted:** that file is untouched by this change
  (`git diff --name-only 80f46a0..HEAD` does not list it), and running `gofmt -l` against the file's
  content extracted from the baseline commit reports it unformatted there too. The spec's baseline
  line claiming "no output" was therefore already inaccurate at `7aa283b`.
- `cd folio-designer && npm test` — **40 files / 411 tests, all passing** (baseline 409; the two new
  ones are AC7's reach and throws proofs).
- `npm run typecheck` — clean. `npm run lint` — **exactly 4** `only-export-components` warnings.
  `npm run build` — succeeds. `npm run test:e2e:compile` — `tsc --noEmit` passes (compile-only; not a run).
- `shasum -a 256 fixtures/*/expected.pdf` from the repo root — **23 lines, byte-identical** to the
  pre-dispatch snapshot, diffed rather than eyeballed. `md5 -q README.md` — `078d7d80d518d54af2fc04fb270d46b8`,
  unchanged. `maximumCacheAssets` — still **64**.
- No `.folio` file, no engine source, nothing under `folio-go/fonts/`, and neither `epics.md` nor
  `ARCHITECTURE-SPINE.md` appears in the change set.

**Matrix test audit.** All eight I/O & Edge-Case Matrix rows are covered by a test that **ran and
passed**, confirmed by name in verbose output rather than inferred from a green package.

**Manual checks — the red-proofs, discharged by mutation.** Each refusal was made unreachable in turn
and the suite re-run; the mutation was then reverted from an absolute-path copy and the tree
re-checked clean each time:

| mutation | tests that redden |
|---|---|
| font site, unclassifiable arm | `TestResolveAssetsRefusesAnUnclassifiableFontLicence` — **and nothing else** |
| font site, copyleft arm | `TestResolveAssetsRefusesACopyleftFontLicence` — **and nothing else** |
| font site, not-allowlisted arm | `TestResolveAssetsRefusesAPermissiveButOffAllowlistFontLicence` — **and nothing else** |
| wordlist, unclassifiable arm | that test's own named `unclassifiable` subtest |
| wordlist, copyleft arm | that test's own named `copyleft` subtest |
| wordlist, non-permissive arm | that test's own named subtest (**added this pass** — see triage log) |
| Site B repointed at the **font** allowlist (the forbidden collapse) | `TestWordlistSiteEnforcesThePermissiveSetNotTheFontAllowlist/CC0-1.0_passes…`, plus `TestResolveAssetsIncludesWordlist`, `TestManifestUpToDate`, `TestCommittedAssetPopulationClassifiesCleanly` |
| classifier: marker branch unreachable | `TestClassifyUbuntuFontLicence` (incl. the MIT-collision subtest) and `TestResolveAssetsAcceptsEveryAllowlistedFontLicence/Ubuntu-font-1.0_by_marker` |
| classifier: `permissiveSPDX` entry removed | a **disjoint** set — `TestUbuntuFontLicenceSPDXLineIsPermissive`, `TestIsPermissiveSPDXReadsTheSameList`, `TestLicenceGraphFixtureScan/permissive` |

**No proof reddened on a neighbouring guard.** Two mutations, two distinct reds, as AC5 requires.

**AC8, re-measured independently of the implementation.** 1438 tracked files; 64 excluded by the
`*/testdata/lint` skip; **1374** in the widened population against 18 before it; 14 `.ttf` tracked in
total, **11** after the skip, and **zero** tracked files carrying any other font-plausible extension.
The widened guard newly reports **nothing**. Confirmed that AC7 widened the **guard**, not the gate —
`ResolveAssets` has walked repo-wide since Story 3.6 and already filters by extension — so **no
non-font asset is newly subjected to the font allowlist**.

**AC6, the gate lands green.** `go run ./cmd/genmanifest` regenerates `lint/MANIFEST.md`
byte-identically, and `TestCommittedAssetPopulationClassifiesCleanly` re-checks on every run that the
real population resolves without error, that no row reads `SEE NOTICE`, and that every font row
carries one of the four ids. The population is the **eleven** font directories plus the wordlist:
six under `folio-designer/public/fonts/` (OFL-1.1), three under `folio-go/fonts/` (OFL-1.1),
`folio-go/testdata/fonts` (**Apache-2.0**), `folio-go/testdata/fonts/notosansthai-variable-testonly`
(OFL-1.1), and `folio-go/internal/text/wordlist/words_th.txt` (**CC0-1.0**).

### Residual risks

- The largest is deferred item 1: the new marker branch is an unanchored substring conjunction, so a
  licence text merely *mentioning* the Ubuntu Font Licence *and* "version 1.0" classifies as
  `Ubuntu-font-1.0`. It is a faithful copy of the OFL branch's shape, which the spec named as the
  template, so correcting it means correcting both branches — its own story.
- `Ubuntu-font-1.0` is on the allowlist with **no asset in the tree under it**. Its classification is
  proved only against synthetic inputs and the SPDX-line fixture; no real-population witness exists
  until Story 8.5 ships a face. Recorded as DW-115.
- The repository's widest directory-walking check now runs in the **designer's** vitest suite rather
  than the Go lint suite that owns the gate it mirrors. Both execute in CI, so nothing is unguarded,
  but the two halves of one invariant now live in two suites.

### Process note — an unauthorized commit at step-03

The implementation subagent created commit `ad6258b` itself, during step-03. This is the **fourth**
such breach in this pipeline (D-8.5.9 records three). It was audited rather than trusted before being
kept, per the Finalize rule that commits already created during the run are retained: its content
matches the spec's task list, it touches no forbidden path, it does not modify `<intent-contract>`,
and it carries the required trailers. The review patches were then applied **without** a further
subagent commit, and finalization committed them itself.

## Delivery Log

### 2026-09-02 — done

Baseline `80f46a0`. Two commits: **`ad6258b`** (implementation) and **`1dee3f5`** (review patches),
both on `main`, both audited, **nothing pushed** — `origin/main` still sits at `c985b9c`.

**What actually shipped.** The asset licence gate stopped recording and started refusing, at both
fall-through sites, **against two different lists** — which is the story's shape and the thing most at
risk of being tidied away later. The half of the defect that mattered more was not the one D-8.5.13
described: a `GPL-3.0` font never reached the `"SEE NOTICE"` fall-through at all — it classified
perfectly well and passed because both sites discarded the classifier's family verdict. **D-8.4h.1**
corrected the ruling's account of its own defect without changing the story; **D-8.4h.2** settled that
the owner's `"UFL"` denotes `Ubuntu-font-1.0` and that the bare alias must not become a live map key;
**D-8.4h.3** accepted the `oversized` and `multiple-goals` warnings with reasons rather than splitting
again. The story exists at all because of **D-8.5.13** (the charter, which lifted the licence gate out
of Story 8.5), sitting alongside **D-8.5.12** in the same plan-gate session.

**Triage: patch 6 / defer 7 / reject 7, with 0 `intent_gap` and 0 `bad_spec`.** No `high` at any point.
`followup_review_recommended: true` was raised mechanically (score 8 ≥ 5) and was **discharged without
a second review dispatch, per D-8.4h.4** — the reasoning being that the review's one material finding
was caught by *deletion* mutation after *substitution* mutation had wrongly reported the arm covered,
which is the stronger instrument, so a second pass would re-ask an answered question. The scrutiny was
routed to this close instead, as that decision requires.

**Gates, measured at close on the committed tree — none carried forward.** `lint`: four `ok`, zero
FAIL, `go vet` clean. `genmanifest` run **twice** for idempotency, `git diff lint/MANIFEST.md` empty
both times, and `SEE NOTICE` appears **zero** times in the manifest. `folio-go`: **1815 pass / 2 fail /
5 skip**, the two being `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest — the
one standing red by identity (D-000.17 / D-2.1.14), with no second distinct red; `go vet` clean.
Designer: **40 files / 411 tests** all passing, `typecheck` clean, `lint` at **exactly 4**
`only-export-components` warnings, `build` succeeded through `build:offline` and `verify:offline`,
`test:e2e:compile` (`tsc --noEmit`) clean. The **23** golden PDF digests were **diffed** against a
worktree checked out at `80f46a0` and are identical; `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`
unchanged; `maximumCacheAssets` still **64**.

**`gofmt -l folio-go lint` prints `lint/internal/rules/licencegraph_test.go`.** Verified pre-existing
independently at close and recorded as a **named standing red** — the spec's `## Verification` line
that claimed "no output" was amended here per **D-8.4h.5**, and the file was deliberately **not**
reformatted, because that would have put an unrelated file in this story's diff. Registered as DW-116.

**Heavy tests: per-epic cadence (D-000.4).** The four `FOLIO_MATRIX_TARGET` legs,
`TestCrossTargetByteIdentity` and the Playwright suite are **written and compiling but NOT RUN here** —
confirmed compiling by `go vet ./...` and `go test` over `folio-go` (which builds `matrix_test.go`,
`fixture_test.go` and `byte_neutrality_test.go`) and by `test:e2e:compile` passing. They come due at
**Epic 8's boundary gate**. This story adds no integration or e2e surface of its own.

**Provenance.** `ad6258b` was created by the **step-03 implementation subagent**, not by the dispatch's
finalization — **instance four of D-8.5.9**. It is **kept and audited** per **D-8.4h.6**, which
reclassified the no-commit rule as unenforceable by prompt and replaced it with this per-story audit.
Both commits were checked at close: contents limited to this story's files, required
`Co-Authored-By:` and `Claude-Session:` trailers present, author and committer consistent with the
rest of the branch, branch `main`, nothing pushed.

**Deferred work — all seven registered with owners, plus two the close found.** DW-117 through DW-123
carry the review's seven deferrals out of the spec frontmatter and into the standing register.
**DW-117** (the unanchored marker branch) is explicitly **its own story**, owned by the engineering
lead, because correcting it touches the pre-existing OFL branch that classifies ten of the eleven
committed font directories. **DW-120** is the one most worth promoting: measured at close, appending a
fifth identifier to the owner's four-id allowlist reddens **nothing**.

**Two findings the close raises, reported rather than patched.** **DW-124** overturns one of the
review's seven rejections: an American-spelled Ubuntu Font Licence does **not** fail loudly as the
rejection claimed — it falls through to the MIT branch and passes under the wrong label, because any
real UFL text carries MIT's grant clause. **DW-125** was named by no review layer: the classifier
returns on the **first** `SPDX-License-Identifier:` line anywhere in the text, before any marker branch
runs, so a full GPL text carrying a stray permissive SPDX line passes the now-fail-closed gate. Both
are pre-existing mechanisms whose **consequence** this story created, both are owned by the engineering
lead, and neither is live in the committed tree.
