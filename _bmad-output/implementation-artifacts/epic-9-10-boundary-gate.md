# Epics 9 and 10 — Boundary Gate

**Run 2026-09-02 at `cb98151`, branch `main`, tree clean before and after. VERDICT: PASS.**

Go 1.26.0 darwin/arm64, node v24.16.0, `@playwright/test` 1.58.2. Nothing was pushed or branched. The
gate fixed no code; the one file it added is named below and stated as a deviation, not smuggled in.

Epics 9 and 10 were reconstructed, reviewed for the first time and repaired (eight edits plus three
follow-ups) before this gate. This run defers nothing further: it is the gate that decides whether
they close, and **Story 15.3 cuts `folio-go/v0.1.0` over exactly this code.**

## The byte-identity spine — RAN, and green

Not run since Epic 8's gate. Every invocation below is from `folio-go/`.

| Leg | Command | Result |
|---|---|---|
| `darwin/arm64` | `FOLIO_MATRIX_TARGET=darwin/arm64 go test -count=1 -tags=matrix -run TestTargetRenderHash -v .` | **PASS 1.31s** |
| `linux/amd64` | same, `FOLIO_MATRIX_TARGET=linux/amd64` | **PASS 9.48s** |
| `linux/arm64` | same, `FOLIO_MATRIX_TARGET=linux/arm64` | **PASS 6.19s** |
| `js/wasm` | same, `FOLIO_MATRIX_TARGET=js/wasm` | **PASS 12.04s** |
| Cross-target | `go test -count=1 -tags=matrix -run TestCrossTargetByteIdentity -v .` | **PASS 27.88s** |

That is AD-1/AD-2/AD-3/AD-21/AD-22/AD-23. **The cross-target check passed.** It reported **24 distinct
documents**, each with an identical sha256 and byte length on all four targets — for instance
`embedded-font` `f533b04b7a4ccb20587f096c9e3173a48fbc870b8c718a73fecf869c6d851832`, 3225 bytes on
darwin/arm64, linux/amd64, linux/arm64 and js/wasm alike; `keep-together`
`6ed495b4c22d7473d82c536c40dce8ca6f2a2fa4bf38efff44b2207929137640`, 82825 bytes, likewise on all four.

**The 24 reconcile exactly against the goldens, by set difference and not by arithmetic.** Comparing
the matrix document names against `ls fixtures/*/expected.pdf`: every one of the **23** golden PDFs is
a matrix document, and the single matrix document with no golden PDF is `hidden-image`, which pins an
`expected.json` rather than a PDF. Nothing is in the goldens that the matrix does not cover.

## Goldens — and what the digest is NOT evidence about

`shasum -a 256 fixtures/*/expected.pdf`, run from the **repo root**, returned **23 digests**, and the
rolled digest of that listing is
`892a1505e5e7fff0184310d5f70eb7bfcfa10d18cda9af4e2aecf262a0630ce9`. So: **all 23 goldens hold at
`892a1505…30ce9` — and that is not evidence about colour or element-box strokes, because no fixture
declares either (DW-147).** Both halves of that sentence are deliberately in one paragraph, because
the failure this line exists to prevent is a rolled digest being read later as coverage it never was.
No golden moved. No digest was re-recorded. **This is what Story 15.3 inherits: a determinism proof of
23 documents, none of which exercises `style.color`, `style.background` or `style.border`.**

## The E9-2 defeater — DISCHARGED, with its evidence

The defeater was that a negative `style.border.width` would force an owner question — clamp it, or
refuse it. Re-run at this gate from the repo root:

```
grep -rnE '"width"[[:space:]]*:[[:space:]]*-' --include='*.json' --include='*.folio' --include='*.go' --include='*.ts' --include='*.tsx' .
```

**Two hits, both in `folio-go/element_box_test.go`** (lines 465 and 503), and **zero hits in any
`.json`, `.folio`, `.ts` or `.tsx` file anywhere in the repository** — that narrower grep, run
separately over `--include='*.json' --include='*.folio'` and over `--include='*.ts' --include='*.tsx'`,
returned empty with exit 1 in both cases. A broader form (`[Ww]idth"?\s*[:=]\s*-[0-9.]`) found the same
two and nothing else.

**The two hits are the discharge, not a breach of it.** They are the negative literals inside
`TestANegativeBorderWidthIsRefusedAtLoad`, which asserts that `"border": {"width": -5}` fails
`ParseTemplate` with a `*RenderError` carrying `DiagCodeTemplateFieldInvalid` and naming both the
element and the field `style.border.width` — and that the same refusal fires on a table's own
`headerStyle.border.width`, the path the value has been reachable through since Epic 4. Zero stays
valid: it is the thinnest line PDF can draw, not an absent border.

So *"clamp-versus-refuse never became an owner question"* is a **measured** statement in the precise
sense intended: the code settles it (**refuse, at load**), and **no authored artifact in the
repository — no fixture, no template, no designer source — ever declares a negative width for the
question to arise from.** A defeater that did not fire is evidence; this is that evidence.

## Playwright — EXECUTED AND GREEN, one assertion, with its deviation stated

**It ran. It is not owed.** A real browser opened and the assertion passed.

```
cd folio-designer
PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH="$HOME/Library/Caches/ms-playwright/chromium-1228/chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing" \
  npx playwright test e2e/e9-5-border-no-ink.spec.ts --reporter=list
→ 1 passed (1.9m); the test itself 698ms
```

**The claim, and why it and nothing else.** *A component whose border paints no ink must project no
border fields, and the canvas must paint no border.* That is the E9-5 repair: `border: {"edges": [],
"color": "#ff0000"}` previously projected `borderWidth` and `borderColor` while `borderEdges` vanished
under `omitempty`, so App.tsx's `component.borderEdges ?? boxEdges` fell back to all four and **the
canvas painted a full red border for a document that prints none.** It was chosen because it has **no
other witness**: **DW-74 is exactly blind to it** — the Go/TS wire test records the projection's
top-level and `CanvasFontChain` key lists and **not `CanvasComponent`'s**, so which per-component
fields are emitted is precisely its blind spot. Before this run the claim was asserted only by
`TestABorderPaintingNoInkProjectsNoBorderFields`, Go-side.

**One assertion, not a suite**, per the standing scope prohibition. The document carries both arms so
the single assertion is non-vacuous by construction: `e1` declares a border that paints nothing, `e2`
declares one that paints, and the expectation is the **exact** list
`['text e2 :: bottom=rgb(0, 255, 0)']` over every `.canvas-box` the canvas drew.

**It was red-proved, so "passed" is a measurement.** Flipping `e1`'s border to `{edges: ['top'], color:
'#ff0000'}` and re-running made the assertion **fail**, with the diff naming exactly
`+ "text e1 :: top=rgb(255, 0, 0)"`. The spec was restored immediately afterwards. The assertion
detects the defect it was written for.

**THE DEVIATION, stated rather than smoothed — unchanged from Epic 8 and re-measured here.** The
pinned browser is chromium revision **1208 / Chrome 145.0.7632.6**. `~/Library/Caches/ms-playwright/
chromium-1208` is still the **428 KB stub** (a 52 KB launcher and nothing behind it), and the download
host still refuses the asset: `cdn.playwright.dev/.../chromium/1208/chromium-mac-arm64.zip`
307-redirects to `playwright.download.prss.microsoft.com`, which returns **HTTP 400**. The suite
therefore ran, through `playwright.config.ts`'s official
`launchOptions.executablePath` hatch, against **Chrome for Testing 149.0.7827.55 (revision 1228, 344
MB)**. **The pinned-revision leg remains undischarged, and behaviour under Chrome 145 is a COULD NOT
LOOK.**

## The full set, for the record — every command with its working directory

| Command | Working dir | Result |
|---|---|---|
| `go test -count=1 ./...` | `folio-go` | **1892 pass / 2 fail / 5 skip** (see carried failure) |
| `go test -count=1 ./...` | `lint` | **GREEN** — 4 packages ok (`cmd/genmanifest`, `internal/licence`, `internal/manifest`, `internal/rules`) |
| `gofmt -l folio-go lint` | **repo root** | exactly `lint/internal/rules/licencegraph_test.go` |
| `npm run typecheck` | `folio-designer` | **exit 0** |
| `npm run lint` | `folio-designer` | **exactly 4** `react(only-export-components)` warnings — `src/preview/pdf-viewer.tsx:16,17`, `src/App.tsx:1403,1410` |
| `npm test` | `folio-designer` | **42 files / 432 tests, all passed** (17.46s) |
| `npm run build` | `folio-designer` | exit 0 |
| `npm run verify:offline` | `folio-designer` | exit 0 |
| `npm run verify:offline:red` | `folio-designer` | exit 0 |
| `npm run verify:offline:wasm` | `folio-designer` | exit 0 |
| `npx tsc -p tsconfig.e2e.json --noEmit` | `folio-designer` | exit 0 |
| `shasum -a 256 fixtures/*/expected.pdf` | **repo root** | 23 digests, rolled `892a1505…30ce9` |
| `go run ./cmd/genmanifest` | `lint` | wrote `lint/MANIFEST.md` |
| `git ls-files --error-unmatch lint/MANIFEST.md` | **repo root** | resolves — pathspec proved |
| `git diff --exit-code` | **repo root** | **exit 0** — the manifest is a function of the source |

**`lint`'s green is load-bearing and is reported as green.** Its cross-package race was fixed at
`664a822` and verified 10/10 bare; a red there would have been a REAL red, and it is not red.

**The three commands that fail in a way reading as clean were all run the way that makes them mean
something.** `gofmt -l folio-go lint` from the **repo root** (from `folio-go/` it prints `lstat
folio-go: no such file or directory` and exits clean, proving nothing). `genmanifest` from inside
`lint/`, paired with a **root-level** `git diff --exit-code`, with the pathspec proved by
`git ls-files --error-unmatch` **run from root, not from `lint/`** — the short-circuit that hid this
at Epic 8. And no conclusion here rests on `git diff <a> <b> -- <path>` patch text, which returns
empty in this environment.

## The carried failure — CARRIED, not this run's

`folio-go` **1892 pass / 2 fail / 5 skip**. The failure set is **exactly**:

```
FAIL github.com/panitw/folio/folio-go/internal/text  TestCorpusMeetsP6ExerciseFloors
FAIL github.com/panitw/folio/folio-go/internal/text  TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)
    corpus_test.go:196: P6g (opaque names) floor not met: got 7, need >=20
    corpus_test.go:200: P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}
```

**Pre-existing and carried**, mandated-unmet by D-000.17 / D-2.1.14, standing at Epic 8's gate in the
identical shape. **Not this gate's to fix, and not rediscovered here as new.** Zero new failures: the
identity check is over the whole set, so a real regression could not have hidden behind it.

## The stale roundtrip manifest — did NOT drift at this gate, and why

`evidence/story-6.7-roundtrip-manifest.json` is **unchanged**, and that is not an all-clear. Epic 8's
gate proved it cannot have come from a real Playwright run, because Playwright had never executed
here; a real run of `browser-native-roundtrip.spec.ts` drifts it by roughly **+182/−102**. **That spec
was deliberately not run at this gate** — this run is one assertion, not a suite — so the manifest was
never touched. **The Epic 8 finding stands exactly as recorded; this gate neither re-confirmed nor
weakened it.** Tree state was verified clean after the Playwright run: the only entry `git status
--porcelain` reported was the new spec file.

## Gate deviation — one file this gate added

`folio-designer/e2e/e9-5-border-no-ink.spec.ts` is committed with this record. **This is a stated
deviation** from "the gate record plus any status flip": the instruction to point Playwright at this
claim requires an instrument, and deleting that instrument after proving it works would leave this
document asserting a browser result nothing can reproduce — and would reopen DW-74's blind spot the
moment this run ended. It is **one test with one assertion**, carrying a comment that says so. It is
not the start of browser coverage, and it reverts in one commit if the owner wants it gone.

## COULD NOT LOOK

- **Chrome 145 (revision 1208) behaviour** — the download host returns HTTP 400 for the archive.
- **Real linux/amd64, linux/arm64 and js/wasm hardware** — the legs are cross-compiled and
  hash-compared from a darwin/arm64 host, as designed.
- **Colour and element-box stroke determinism** — no fixture declares `style.color`,
  `style.background` or `style.border`, so the 23 goldens say nothing about them (DW-147). Gating
  input to Story 15.3.
- **Whether the 6.7 roundtrip manifest still drifts by +182/−102** — the spec that writes it was out
  of this gate's scope.

## VERDICT

**PASS.** All four matrix legs green, cross-target byte identity green across 24 documents, all 23
goldens holding at `892a1505…30ce9` with no digest moved and none re-recorded, `lint` green,
designer typecheck/lint/test/build/offline all at their expected identities, the E9-2 defeater
discharged with evidence, and Playwright **executed and green** on the one claim it was pointed at,
red-proved. The only failure is the carried `P6g` floor. **Epics 9 and 10 close.**
