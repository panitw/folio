---
baseline_commit: 0f68ed97ee832cf41c3afb6badd6014990a6c73f
---

# Story 4.8: Alternating row styling

**Epic:** 4 — A Go developer can render the golden report (**the C4 gate**)
**Story key:** `4-8-alternating-row-styling`
**Status:** `done`
**Covers:** **FR28** · **AD-13**, **AD-21**, **AD-24**

**Standing delivery decisions:** numeric story order; continuous checkpoints; terminal decision
channel; integration/e2e cadence **per epic** (D-000.4). The owner has explicitly included this
optional story in the run, so its former “first to be cut” label is a scope-history note, not an open
question.

## In plain terms (read this first if you just want the gist)

Dense transaction tables are now easier to scan: when an author declares an alternate background,
the renderer gives every odd zero-based data-row collection index that colour. The first row retains
the ordinary treatment, and the stripe follows the original collection across page breaks rather than
restarting on each page. The shipped `alternating-rows` fixture demonstrates that output, and its
native matrix leg has recorded the resulting PDF hash.

This is deliberately a presentation feature only. It does not reorder, filter, hide, split, resize, or
otherwise change rows. It does not shade the column heading or the totals row, and it does not invent a
default colour. If the author also gives ordinary data rows a background, the alternate colour takes
priority only on the rows where the stripe applies; the ordinary background remains visible on the
others. The rule is fixed from the collection position rather than from page position because page
breaks can change as content wraps, while a transaction’s place in the data does not.

The retained checks prove the visible stripe, global parity through a multi-page boundary, and an
enabled-versus-absent relational invariant: page partition, text runs, rectangle cardinality, and every
non-fill rectangle field are identical, so fills are the only allowed difference. Documents without
the setting retain their recorded bytes. The complete cross-platform comparison remains deferred to
the Epic 4 boundary; this story ran the required native leg and compile-checked the other tagged
targets, but does not claim cross-target equality. No designer control was added, and no colour-by-data
rule was introduced.

## Story

**As a** template author,
**I want** alternating row shading,
**So that** a dense transaction table is easier to read across.

---

## Premise verification at `0f68ed97` (performed inline at creation)

### Confirmed — the schema and safety fence exist; the render consumer does not

`TableExt.AltRowBackground` is already parsed and serialized, the format document already describes
it, and Story 3.5 already rejects a `{{ }}` placeholder in it. A production-code search finds the
field only in the parser/serializer and the expression-safety validator; no rendering path reads it.
The feature is therefore representable and guarded but still entirely inert, exactly the obligation
D-3.5.4 assigned to Story 4.8.

This story adds **no schema field** and no data-expression capability. The existing ordinary colour
string remains the sole input.

### Confirmed — the direct row-index seam was preserved specifically for this story

Story 4.2 already carries each data row’s zero-based collection index on both its chrome source and
its text runs. Stories 4.3–4.6 kept header and footer row-type tags distinct from data-row tags even
when pagination temporarily changes grouping. `table_footer.go` and the source-struct comments name
Story 4.8 explicitly: stripe from the data-row tag and collection index, never from a pagination group
key and never from a page-local counter.

The implementation seam is consequently narrow: choose the data row’s effective fill while its cell
rectangles are built. Pagination receives the same rectangles, dimensions, ordering, row identity and
grouping as before. No pagination code needs to change.

### Confirmed — colour precedence is already ruled and documented

D-4.2.1 requires `table.altRowBackground` to win over `style.background` on rows where the alternate
colour applies. The format document already records this precedence as “not yet implemented.” This
story implements that ruling and removes only the stale qualifier. Do not re-open the ranking.

The complementary cases are forced by that ruling:

- a non-alternate data row keeps `style.background` when present;
- an alternate data row uses `altRowBackground` even when the base background differs;
- without a base background, non-alternate rows remain unfilled;
- headers remain governed by `headerStyle` → `style`, and footers remain governed by the existing body
  style; neither is a data row.

### Confirmed — Story 4.7 deliberately cannot serve as this feature’s golden

All four statement fixtures explicitly omit the alternating-background field so Story 4.8 cannot
move them. Their shared sign-off binds all four digests and would be invalidated in whole if their
template were changed. Reusing them would therefore turn a small optional visual feature into four
golden re-recordings plus a second human inspection.

There are **13** committed `expected.pdf` artifacts at creation. AD-21 requires this visual feature to
ship a golden, so Story 4.8 adds one small, independently registered fixture rather than changing any
of those 13.

---

## Settled row parity

Collection index **0** retains the base background and the alternate background applies to **odd
zero-based indexes** (the second, fourth, sixth visible rows). The field is named the *alternate*
background, so the first row establishes the base treatment and the next alternates from it. This
also preserves the ordinary `style.background` visibly on the first row when both colours are
declared.

This is the story's fixed behavioural convention, not an implementation choice to reopen. The
collection-global index already carried by each data row is authoritative, so pagination never resets
the pattern.

---

## Governing decisions and invariants

- **D-4.2.1 — precedence.** The alternate background wins over the table body’s ordinary background
  on applicable rows.
- **D-4.2.2 — direct row identity.** Collection membership and index are carried, never reconstructed
  from geometry, ordering, element ids, or pagination groups.
- **D-3.5.4 / D-3.5.5 — inert style fields and the schema-derived safety population.** This story must
  turn the existing field into a real output consumer without weakening the existing no-placeholder
  check or replacing its relational completeness guard with a hand list.
- **AD-13.** Column widths and row geometry remain derived. Shading changes no width, height, padding,
  wrapping, or page partition.
- **AD-24.** Boxes are absolute and do not negotiate. The stripe paints the cell rectangles already
  derived for the row.
- **AD-21 / AD-22.** A visual feature ships a recorded fixture and enters the four-target matrix; an
  intended golden addition is reviewed and recorded, never regenerated merely to obtain green.
- **D-000.22 / D-000.53.** First recording needs a semantic assertion read from the produced PDF and
  structural validation by an independent reader, with tool version, invocation and output retained
  in provenance.
- **D-000.54.** A new matrix document runs its native host leg when registered. This proves only that
  the leg executes and hashes on one target; cross-target equality remains the Epic 4 boundary gate.
- **D-000.4.** Heavy integration/e2e and the complete four-target matrix run at the Epic 4 boundary.
  This feature adds no new source of cross-target divergence—only an integer parity choice and an
  existing exact colour path—so it does not qualify for the determinism override.
- **D-000.68 / D-000.79.** Tests anchor expectations in literals they own, and their discriminating
  mutations were named before implementation. If the final shape makes a named mutation impossible,
  replace it explicitly and explain why.

---

## Design constraints

1. **Resolve from template state only.** Alternation is not data-driven. Do not evaluate row fields,
   conditions, or expressions to choose a colour; Story 3.5’s placeholder rejection remains intact.
2. **Use the collection index already carried by the row.** Do not derive parity from the page, the
   row’s y-coordinate, output order, `layout.ItemGroup.Key`, or how many headers/footers were emitted.
3. **Paint only data-row cell rectangles.** Header, repeated-header and footer rectangles retain their
   existing style cascades. Text runs need no new styling state.
4. **Do not modify row geometry.** The same number of rects and runs, with identical x/y/width/height,
   must be produced with the feature on or off; only fill presence/colour may differ.
5. **Reuse the existing colour path.** `#RRGGBB` parsing, `pagemodel.Color`, the exact integer PDF
   channel conversion, and `STYLE_COLOR_INVALID` already exist. Add no colour parser, float, new
   diagnostic code, or PDF operator.
6. **Keep the change local.** The production seam is the existing data-row rectangle construction in
   `table_render.go`. No new internal package, schema field, pagination primitive, public API, or
   generated-file edit is justified.
7. **The statement goldens are immutable in this story.** Do not add the field to
   `statement-{1,5,20,50}`, move their bytes, or invalidate `statement-signoff.json`.

---

## Acceptance criteria

### AC1 — alternate data rows receive the configured background

**Given** a table with an alternating-row background and at least five data rows
**When** it renders
**Then** every column cell in each applicable row carries exactly the configured fill
**And** every non-applicable row retains the base body background, or remains unfilled when no base
background is declared.

**Instrument.** Add a focused `table_alternating_row_test.go` table-driven test over both base cases:
base `#112233` and no base fill. Use at least two columns and five rows, inspect the page model’s cell
rectangles, and compare a test-owned ordered sequence of row colours. Assert every cell in one row has
the same result; checking one cell would allow half-striped rows to pass.

**Anchor.** Literal input colours and the lead-ruled parity, stated independently in the test.

**Red-proof.** Remove the alternate selection from the data-row rectangle call: only the test for
applicable rows reds; header/body chrome tests remain green. Also force only column zero to receive the
alternate colour: the every-cell assertion reds.

### AC2 — the alternate colour wins over `style.background`, and only where it applies

**Given** `style.background` and `table.altRowBackground` are both declared with different colours
**When** data rows render
**Then** applicable rows use the alternate colour
**And** the intervening rows use the ordinary body colour.

This is D-4.2.1 becoming executable. The test must use visibly different test-owned RGB literals and
assert both directions. A test that checks only the alternate rows can pass if the implementation
accidentally overwrites every row.

**Red-proof.** Reverse the precedence so base wins: the alternate-row half reds. Apply alternate to
all rows: the base-row half reds.

### AC3 — alternation follows collection index across a page break

**Given** a table whose data rows span at least three pages
**When** pagination places the rows
**Then** each row’s fill agrees with its original collection index
**And** the first data row on a continuation page is not treated as a new index zero.

**Instrument.** Build a fixture whose measured page partition puts a parity-discriminating row first
on at least one continuation page. Read each page’s `ContentRects` back to `tableRectSource`, assert
the presence of at least three pages and a continuation whose first collection index would receive a
different colour if parity restarted, then compare that source’s actual rect fills to the original
index. The presence precondition is mandatory; “several pages” without a parity-discriminating
boundary is vacuous.

Also render through the public API and assert page count and structural validity, so the private
pagination observation is tied to the shipped route.

**Red-proof.** Compute parity from a page-local row counter. The chosen continuation boundary must red
while the one-page AC1 test stays green.

### AC4 — headers, repeated headers, and footers never become striped rows

**Given** a paginated table with a styled header, an ordinary body background, an alternate
background, and a configured footer aggregate
**When** it renders
**Then** the original and repeated headers retain the header cascade
**And** the footer retains the existing body-style background
**And** neither uses the alternate background regardless of the adjacent data row’s index.

**Instrument.** Reuse the explicit `isHeaderRow` / `isFooterRow` / `isDataRow` distinctions and inspect
both source groups and final page-model rectangles. Use three different literal colours so identity is
unambiguous.

**Red-proof.** Key the stripe from group index alone: the footer-orphan path can temporarily share a
data-row key and must red this test. Key it from “not header”: the footer half reds.

### AC5 — malformed configured colour uses the existing located error

**Given** at least two data rows and a malformed ordinary string in the alternating-background field
**When** rendering reaches an applicable row
**Then** rendering fails with the existing `STYLE_COLOR_INVALID` code
**And** the diagnostic names the table element and the alternating-background field in its message
**And** no PDF bytes are returned as a successful result.

Do not mint an alternating-row-specific code. Reuse the same parser/error channel as body/header
background colours, but make the field name truthful rather than reporting only `style.background`.

**Red-proof.** Bypass colour parsing only for the alternate branch: this test reds while the existing
malformed body/header colour tests remain green.

### AC6 — the field is a real output consumer and remains non-data-driven

**Given** two otherwise byte-identical templates differing only in the alternating-background value
**When** each renders the same non-empty data
**Then** their page-model fills and PDF bytes differ
**And** changing unrelated report data does not change which rows receive which colour.

Retain Story 3.5’s load-time tests for rejecting placeholders and its schema-derived completeness
guard. Add a positive render assertion so D-3.5.4’s previously inert field is now observably live.

**Red-proof.** Hard-code the first test colour in production: the page-model literal comparison reds.
Choose parity from a row value: the unrelated-data comparison reds.

### AC7 — a new, independently accepted golden fixture ships and joins the matrix

Add one small one-page fixture, recommended slug **`alternating-rows`**, with five single-line data rows
and one column. Keep the ordinary body background absent so the produced PDF has exactly the ruled
number of alternate-colour row fills; give the header a different colour to keep it distinguishable.
The fixture carries the normal `input.folio`, `data.json`, `expected.pdf`, `expected.json` and
`README.md` provenance set.

The untagged fixture test must re-render and byte-compare the committed PDF, assert it is non-empty and
one page, and read the **produced PDF bytes** to assert the alternate-colour fill count and vertical
ordering from test-owned literals. Counting the template field or inspecting pre-PDF rectangles alone
does not discharge D-000.22.

Record independent structural validation in the README: reader name/version, verbatim invocation,
page/object result, date, and digest (D-000.53). Register the digest in every required
`goldenDigestRecord` site, add the document to `matrixDocuments`, add its slug and all four artifact
paths to `.github/workflows/matrix.yml`, and keep the registration guard green.

Run the new document’s native host matrix leg before `review` exactly as D-000.54 requires. Record both
truths: it executed and hashed on one target; it did **not** prove cross-target equality. The complete
four-target run remains deferred to the Epic 4 boundary.

**Red-proofs.** (a) Shade the opposite parity: the PDF semantic count/order assertion reds. (b) Remove
the matrix entry or workflow slug: the untagged registration guard reds. (c) Change the committed PDF
without its independent digest site: the digest-completeness guard reds.

### AC8 — absent configuration is byte-neutral for every existing fixture

**Given** a table without an alternating-background field
**When** it renders
**Then** its page-model rectangles and PDF bytes are unchanged from the baseline.

All **13 existing** golden PDFs must reproduce byte-for-byte and retain their current digests. In
particular, all four statement fixtures and their sign-off remain untouched. Measure and record the
existing-fixture result as a set, not as “goldens green”; if any existing digest moves, stop and treat it
as a defect until a lead ruling explains why.

**Red-proof.** Force a default stripe when the field is absent: at least the existing table goldens
must red. If no existing golden exercises the mutation, add a direct absent-field byte comparison and
state honestly which layer catches it.

### AC9 — quality gates and cadence are reported without overclaiming

Before status `review`, run:

1. `folio-go`: ordinary unit suite with the named corpus-floor red skipped, plus the full suite to
   confirm that the only red remains `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`;
2. `lint`: full tests;
3. `hashmatrix`: full tests;
4. `go build ./...` in all three modules, plus the project’s normal formatting/vet checks;
5. all new discriminating mutations, each applied, observed red for the intended reason, and restored;
6. the `alternating-rows` native host matrix leg required by D-000.54.

Integration/e2e tests are **written but not run** in-story under the per-epic cadence. The four-target
matrix, including the new document, runs at the Epic 4 boundary before `epic-4` can become `done`.
This story is not complete on unit tests alone, and the Delivery Log must name the deferred suites
explicitly.

---

## Implementation tasks

- [x] Bake the settled odd-zero-based-index parity into tests before recording the golden.
- [x] Add a single template-only resolution of the alternate colour and apply it only while building
      data-row cell rectangles from the carried collection index.
- [x] Preserve header, repeated-header, footer, row geometry, wrapping, grouping, pagination and
      diagnostics other than the existing colour-error path.
- [x] Add focused table-driven unit tests for AC1–AC6, including page-boundary and footer-orphan cases
      with explicit non-vacuity preconditions.
- [x] Add and independently validate the `alternating-rows` golden; add its byte comparison, semantic
      PDF assertion, digest registry entry, matrix document, CI slug and four artifact paths.
- [x] Update `folio-format.md` only to state the implemented parity and remove “not yet implemented”;
      do not change the schema or permit data-driven colours.
- [x] Run the required native matrix leg, unit/lint/build gates and every named mutation; report exact
      counts and exact deferrals in the Delivery Log.
- [x] Set this story to `review` only after all in-story gates pass. Do not mark `epic-4` done; the
      orchestrator owns the Epic 4 heavy boundary gate after this story’s commit.

---

## Files expected to change

**Production:** `folio-go/table_render.go` only, unless a measured need proves otherwise.

**Focused tests:** a new `folio-go/table_alternating_row_test.go`, plus existing golden/matrix registry
tests where the new fixture must be declared.

**Fixture and registration:** a new `fixtures/alternating-rows/` directory;
`folio-go/byte_neutrality_test.go`; `folio-go/matrix_test.go`; `.github/workflows/matrix.yml`; and the
small capture/fixture test file chosen consistently with existing fixture conventions.

**Documentation/records:** `_bmad-output/specs/spec-folio/folio-format.md`, this story file, status-only
`sprint-status.yaml`, and the append-only decision log for the parity ruling.

**Must not change:** any existing `fixtures/*/expected.pdf` or recorded digest; any
`fixtures/statement-*` input/data/params/README; `fixtures/statement-signoff.json`; pagination
semantics under `internal/layout`; the template schema; generated files; or the public API.

---

## Delivery Log

### 2026-08-27 — created (`ready-for-dev`)

Baseline `0f68ed97ee832cf41c3afb6badd6014990a6c73f`, driven explicitly as Story 4.8. The owner retained
the run’s existing settings and confirmed the optional story is included. Creation verified the
actual seam rather than assuming it: the field is parsed, serialized, documented and protected from
data placeholders but has no render consumer; data rows already carry their collection index; header
and footer row types are deliberately distinct; D-4.2.1 already settles colour precedence; and the
four signed statement fixtures deliberately omit this field. Thirteen committed golden PDFs exist, so
the story requires one new isolated fixture and zero changes to the existing set.

Decisions applied: D-4.2.1 (alternate wins over base), D-4.2.2 (direct carried row identity),
D-3.5.4/D-3.5.5 (make the inert field real without weakening the safety population), D-000.4
(per-epic heavy cadence), D-000.22/D-000.53 (semantic and independent first-recording checks), and
D-000.54 (native matrix leg on registration). Creation also fixed the row convention explicitly:
collection index zero keeps the base treatment and odd zero-based indexes receive the alternate
colour, so no parity choice is left to implementation accident.

Measured baseline gates, all from a clean story baseline apart from unrelated pre-existing BMAD
configuration/untracked agent files which were not touched:

| Gate | Exact command/scope | Measured result |
|---|---|---|
| `folio-go` green unit gate | `CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -json ./...` | **1177 pass · 0 fail · 4 skip** (including subtests) |
| `folio-go` full disclosure | same command without `-skip` | **1183 pass · 2 fail entries · 4 skip**; only `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest are red by design |
| `lint` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -json ./...` | **115 pass · 0 fail · 0 skip** |
| `hashmatrix` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -json ./...` | **3 pass · 0 fail · 0 skip** |
| build | `CGO_ENABLED=0 GOWORK=off go build ./...` in `folio-go`, `lint`, and `hashmatrix` | **all three clean** |

No integration/e2e or cross-target matrix suite ran at creation. They are explicitly deferred by the
per-epic cadence and are due at the Epic 4 boundary. Story implementation must still run the newly
registered document’s single native host matrix leg under D-000.54; that is a registration sequencing
check, not the deferred cross-target gate and not a cadence override.

### 2026-08-27 — development recovery complete (`review`)

Recovered and verified the pre-existing Story 4.8 work against baseline
`0f68ed97ee832cf41c3afb6badd6014990a6c73f`; no existing golden was re-recorded. Production consumes
`TableExt.AltRowBackground` only when data row `rowIdx` is odd, while building the row's existing cell
rectangles. Index 0 is therefore base, carried collection identity remains global across page breaks,
and the header/repeated-header/footer paths remain separate. The invalid-colour branch now truthfully
names `table.altRowBackground` while retaining `STYLE_COLOR_INVALID`. AC4 was strengthened during
recovery to count the final composed-page footer cells as base-colour cells as well as checking the
pre-pagination row-type sources.

The recovery also appended ratification provenance to D-4.8.1: its original text was proposed and
logged before lead consultation; the warm engineering lead subsequently reviewed and ratified the
unchanged ruling (index 0 base; odd zero-based collection indexes alternate globally; headers,
repeated headers and footer aggregates excluded; appearance-only under AD-13/AD-24).

| Gate | Exact command/scope | Measured result |
|---|---|---|
| focused alternating behaviour + fixture tests | `rtk go test -count=1 -run 'TestAlternating(Row|Rows)' .` in `folio-go` | **10 pass · 0 fail** |
| focused registration/digest/structure/table-suite guards | `rtk go test -count=1 -run 'Test(MatrixDocumentSlugsAreRegisteredInCI|GoldenDigestAgreesAtEveryDeclaredSite|TableBehaviourSuiteIsNotSupersededByTheGolden|EveryGoldenPDFResolvesItsPageTree)' .` | **18 pass · 0 fail**; digest guard re-hashed **14** artifacts (the 13 inherited PDFs plus `alternating-rows`) and all **42** recording sites |
| ordinary `folio-go` green gate | `rtk env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -json ./...` | **1188 pass · 0 fail · 4 skip** (test actions) |
| ordinary `folio-go` full disclosure | same command without `-skip` | **1194 pass · 2 fail events · 4 skip**; the two events are solely parent `TestCorpusMeetsP6ExerciseFloors` and `P6g_(opaque_names)`, still **7 < 20** |
| `lint` | `rtk env CGO_ENABLED=0 GOWORK=off go test -count=1 -json ./...` | **115 pass · 0 fail · 0 skip** |
| `hashmatrix` | `rtk env CGO_ENABLED=0 GOWORK=off go test -count=1 -json ./...` | **3 pass · 0 fail · 0 skip** |
| format/vet/build | `rtk gofmt -l $(rtk rg --files -g '*.go')`, then `rtk env CGO_ENABLED=0 GOWORK=off go vet ./...` and `rtk env CGO_ENABLED=0 GOWORK=off go build ./...` in `folio-go`, `lint`, and `hashmatrix` | **0** unformatted Go files; all six vet/build commands clean |
| D-000.54 native matrix leg | `rtk env CGO_ENABLED=0 GOWORK=off FOLIO_MATRIX_TARGET=darwin/arm64 go test -count=1 -v -tags=matrix -run '^TestTargetRenderHash$' .` | all **15** registered documents executed, structurally checked and hashed on native `darwin/arm64`; `alternating-rows` = `e491d628ecd1dae9ad2d396341c014fb9dc5ce1e55c535a2f88bdae15b0e8bbd`, **55,734 bytes** |

Discriminating mutations were each applied, observed red, and restored: (1) disable alternate
selection — applicable rows fail; (2) apply it to every row — base rows fail; (3) apply it only to
column zero — the other cell in each alternate row fails; (4) hard-code the alternate colour — the
template-only colour consumer test fails; (5) bypass alternate-colour parsing — the located-error test
fails; (6) invert parity — the three-page continuation/source/final-composition test fails; (7)
recolour the footer — that test fails both source-footer and final-page-colour counts; (8) force a
default stripe when absent — all four immutable statement goldens fail byte comparison; (9) remove
the workflow slug — the untagged matrix-registration guard fails; and (10) change only the fixture's
recorded digest — the digest-completeness guard fails. A literal page-local counter cannot be injected
at this seam because fills are deliberately fixed before pagination; the inversion mutation is the
executable replacement and the three-page test's odd-index continuation precondition makes the
non-global result observable.

Deferred and explicitly **unrun** under D-000.4: the full four-target matrix/cross-target comparison
(`TestCrossTargetByteIdentity`, including `alternating-rows` on linux/amd64, linux/arm64 and js/wasm),
integration tests, and e2e tests. The native result proves execution and hash on one target only; it
does not claim cross-target equality. These remain Epic 4 boundary work. No blocker or new decision
arose. Story remains `review`; `epic-4` remains `in-progress`.

---

## Review Findings — 2026-08-27

### Finding 1 — The permanent suite does not directly prove geometry and pagination neutrality

- **Severity:** Major
- **Category:** Verification gap / architecture invariant
- **Location:** `folio-go/table_alternating_row_test.go:91-293`; `folio-go/alternating_rows_fixture_test.go:128-177`
- **Observation:** The focused tests assert fill presence, colour, collection parity, page-boundary
  continuation, header/footer exclusion, page count, PDF structure and the enabled fixture's recorded
  bytes. They never render the same multi-page table once with `altRowBackground` present and once
  absent, normalize only fill state, and compare the resulting rectangle count and non-fill geometry,
  text runs, and page partition. The golden pins the enabled artifact after first recording; it does
  not independently establish the required enabled-versus-absent equality. A reviewer-only temporary
  comparator over the 20-row, three-page footer fixture found the current implementation neutral and
  passed, then was removed; the runtime code is not presently exhibiting drift.
- **Impact:** Design constraint 4 and AD-13/AD-24 are binding: shading must not change width, height,
  padding, wrapping, row/run cardinality, placement or pagination. Without a permanent relational
  witness, a future implementation or first re-recording can ratify an appearance-triggered geometry
  or text-run change while all colour-specific assertions remain satisfied.
- **Suggested resolution:** Add a permanent focused test that builds the same parity-discriminating
  multi-page document with the field present and absent; assert identical page count/partition,
  rectangle and run cardinality, every rectangle field except `HasFill`/`Fill`, and all text runs.
  Keep the existing PDF golden and semantic fill assertions as the artifact-level companion.
- **Related AC:** AC3, AC6, AC8; Design constraint 4; AD-13; AD-24; D-000.68; D-000.79.

## Review Summary — 2026-08-27

**Verdict:** Changes requested. Story and sprint tracker remain `review` as required.

| Severity | Count |
|---|---:|
| Blocker | 0 |
| Major | 1 |
| Minor | 0 |
| Nit | 0 |

### Measured review gates

| Gate | Reviewer measurement |
|---|---|
| Focused alternating behaviour + fixture | **10 pass · 0 fail · 0 skip** |
| Registration/digest/structure/table-suite guards | **18 pass · 0 fail · 0 skip**; **14** PDFs re-hashed and **42** declared digest sites checked |
| `folio-go` ordinary gate | **1188 pass · 0 fail · 4 skip** with `TestCorpusMeetsP6ExerciseFloors` skipped |
| `folio-go` full disclosure | **1194 pass · 2 fail events · 4 skip**; only `TestCorpusMeetsP6ExerciseFloors` and `P6g_(opaque_names)` fail at **7 < 20** |
| Baseline attribution | The same two corpus-floor fail events and **7 < 20** result reproduce at baseline `0f68ed97ee832cf41c3afb6badd6014990a6c73f` |
| `lint` | **115 pass · 0 fail · 0 skip** |
| `hashmatrix` | **3 pass · 0 fail · 0 skip** |
| Format / diff / vet / build | no unformatted Go files; `git diff --check` clean; vet and build clean in `folio-go`, `lint`, and `hashmatrix` |
| Deferred tagged coverage | matrix-tagged tests compile for `darwin/arm64`, `linux/amd64`, `linux/arm64`, and `js/wasm`; no cross-target legs were executed by this compile check |
| D-000.54 native leg | all **15** registered documents executed and hashed on `darwin/arm64`; `alternating-rows` = `e491d628ecd1dae9ad2d396341c014fb9dc5ce1e55c535a2f88bdae15b0e8bbd`, **55,734 bytes** |
| Independent artifact/provenance check | qpdf **12.4.0** reports one well-formed PDF; SHA-256 matches; the README's recorded CLI invocation independently reproduces byte-identical output |
| Reviewer-only geometry comparator | enabled versus absent, 20-row three-page footer fixture: identical page count, rectangle/run cardinality, all non-fill rectangle state and all text runs; **pass**, temporary test removed |

### Mutation evidence

Each mutation was applied separately, failed for the intended assertion, and was restored before the
final green runs:

1. opposite collection parity — focused row, continuation and golden assertions red;
2. simulated page-local reset at the measured nine-row continuation boundary — page 2's original
   collection indexes and final fill counts red;
3. alternate colour applied to the original/repeated header — source and all three composed-page
   header assertions red;
4. alternate colour applied to the footer — source-footer and final composed-page counts red;
5. alternate-colour parsing bypassed — the located `STYLE_COLOR_INVALID` assertion red;
6. default stripe forced when the field is absent — all four immutable statement golden comparisons red;
7. `alternating-rows` removed from the workflow comparison slug list — the untagged registration guard red;
8. only `expected.json`'s recorded digest changed — the independent digest-site/completeness guard red.

### Deferrals and decision routing

Deferred under D-000.4 and deliberately **not run** in this story review: the full four-target
cross-target comparison, integration, and e2e suites. They remain the Epic 4 boundary gate; this
review verified tagged compilation only. The native matrix evidence proves one-host execution and
hashing, not cross-target equality.

D-4.8.1's record preserves the material provenance correction: the developer proposed the parity
text before consultation, and the engineering lead later ratified the unchanged convention. No new
lead/owner decision is needed from this review.

---

## Finisher resolution and final Delivery Log — 2026-08-27 (`done`)

**Review triage:** 0 Blocker, 1 Major, 0 Minor, 0 Nit. Finding 1 is **FIXED**; no finding was
dismissed or deferred. The old review verdict remains historical evidence of the requested change;
the measured result below is the completion decision.

**Finding 1 — permanent AD-13/AD-24 relational witness (FIXED).**
`TestAlternatingRowBackgroundIsGeometryAndPaginationNeutral` now renders the same 20-row,
footer-bearing, parity-discriminating table with `altRowBackground` enabled and absent. It first
requires at least three pages, then compares each final page: identical page count/partition,
page geometry, rectangle and run cardinality; every text run; and every rectangle field after only
`HasFill`/`Fill` are normalized. It also requires exactly 30 permitted fill differences (ten odd rows
times three columns), preventing a vacuous neutral result. This is retained beside the enabled golden
and semantic PDF checks, so neither artifact recording nor a colour-only assertion can stand in for
geometry neutrality.

**Red proof for the resolution.** A temporary production mutation added one millipoint to an odd
row's height only when `altRowBackground` was configured. The new test failed before restoration,
reporting changed page-0 text runs and changed non-fill rectangle fields. The mutation was removed,
`gofmt` was rerun, and the restored witness passed.

| Gate | Exact command/scope | Measured result |
|---|---|---|
| focused alternating suite | `rtk go test -count=1 -run 'TestAlternating(Row|Rows)' .` in `folio-go` | **11 pass · 0 fail** |
| registration/digest/structure/table-suite guards | `rtk go test -count=1 -run 'Test(MatrixDocumentSlugsAreRegisteredInCI\|GoldenDigestAgreesAtEveryDeclaredSite\|TableBehaviourSuiteIsNotSupersededByTheGolden\|EveryGoldenPDFResolvesItsPageTree)' .` | **18 pass · 0 fail** |
| ordinary `folio-go` gate | `rtk env CGO_ENABLED=0 GOWORK=off go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' -json ./...` | **1189 pass · 0 fail · 4 test skips** |
| full disclosure | same command without `-skip` | **1195 pass · 2 expected fail events · 4 test skips**; only `TestCorpusMeetsP6ExerciseFloors` and `P6g_(opaque_names)` remain at **7 < 20** |
| `lint` / `hashmatrix` | `rtk env CGO_ENABLED=0 GOWORK=off go test -count=1 -json ./...` in each module | **115 pass · 0 fail · 0 skip** / **3 pass · 0 fail · 0 skip** |
| format / diff / normal vet-build | `rtk gofmt -l $(rtk rg --files -g '*.go')`; `rtk git diff --check`; `rtk env CGO_ENABLED=0 GOWORK=off go vet ./...` and `go build ./...` in `folio-go`, `lint`, `hashmatrix` | **clean**: zero unformatted files, no diff errors, all six vet/build gates pass |
| native D-000.54 matrix leg | `rtk env CGO_ENABLED=0 GOWORK=off FOLIO_MATRIX_TARGET=darwin/arm64 go test -count=1 -v -tags=matrix -run '^TestTargetRenderHash$' .` | **15 documents passed**; `alternating-rows` SHA-256 `e491d628ecd1dae9ad2d396341c014fb9dc5ce1e55c535a2f88bdae15b0e8bbd`, **55,734 bytes** |
| tagged compile coverage | `rtk env CGO_ENABLED=0 GOWORK=off go build -tags=matrix ./...`; `go vet -tags=matrix ./...`; `GOOS/GOARCH go test -c -tags=matrix .` | native build/vet clean; matrix test package compiles for **darwin/arm64, linux/amd64, linux/arm64, js/wasm** |

**Explicitly unrun and still deferred to the Epic 4 boundary under D-000.4:** the complete four-target
matrix and cross-target comparison (`TestCrossTargetByteIdentity`, including `alternating-rows` on
linux/amd64, linux/arm64, and js/wasm), integration tests, and e2e tests. Tagged compilation and the
native result do not prove equality across targets. `epic-4` remains `in-progress`; this completion
does not start Epic 5 or mark the epic done.

**Commit:** the scoped local completion commit was created with no push; its immutable SHA is reported
in the finisher handoff rather than self-referenced inside the commit it identifies.
