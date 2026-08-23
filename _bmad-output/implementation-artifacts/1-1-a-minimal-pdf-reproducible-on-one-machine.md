# Story 1.1: A minimal PDF, reproducible on one machine

Status: done

| | |
|---|---|
| **Story key** | `1-1-a-minimal-pdf-reproducible-on-one-machine` |
| **Baseline commit** | `f2aa8c0` — planning artifacts only; **the repository contains no code whatsoever** |
| **Epic** | 1 — A Go developer can render a deterministic PDF |
| **Source** | `_bmad-output/planning-artifacts/epics.md` §Story 1.1 (lines 415–450) |
| **Contract** | `_bmad-output/specs/spec-folio/SPEC.md` + `acceptance.md` (S1–S3, C3, C6, risk **R1** = Critical) |
| **Invariants** | AD-1, AD-2, AD-3, AD-6, AD-7, AD-22 (plus AD-21, AD-23, AD-26 in the neighbourhood) |
| **Local toolchain** | Go 1.26.0 darwin/arm64 (verified in this run), Node v24.16.0, npm 11.13.0, Docker 29.6.2 |

> **Stop here — do not commit, do not branch, do not set `done`.** The final task of this story is
> "story file, decision log, sprint status → `review`". Committing belongs to the finisher, after
> review.
>
> **There are no ADRs in this project.** The `AD-N` invariants in
> `ARCHITECTURE-SPINE.md` are the ADR equivalent and live in that one file. Cite `AD-N` plus a SPEC
> clause or a story AC — never an ADR path.
>
> **If something in this file looks like it contradicts a decision-log entry, stop and surface it.
> Do not resolve it by choosing.** The ruling governs and this file is wrong.

---

## In plain terms (read this first if you just want the gist)

Folio's whole promise is that the same report, rendered anywhere, produces a PDF file with the same
fingerprint. That promise is a property of how numbers are stored and written down, not a feature
bolted on afterwards. This story wrote the repository's first program: a small library that creates
the project's skeleton, pins the exact compiler version everything will be built with, and produces
a single-page PDF containing one filled rectangle.

It established two rules everything built later depends on. Every measurement is a whole number of
thousandths of a point, never a decimal fraction with a wobbly, platform-dependent text form. And
the numbers reaching the finished file are written by two clearly separated, deliberately narrow
pieces of code rather than scattered across the program — one for measurements, one for plain
counts such as byte lengths or which object a reference points to. Keeping those apart, and away
from every other line of code, is what makes the fingerprint promise enforceable rather than
aspirational. A review pass found a couple of numbers that had slipped past that discipline, and a
guard too narrow to catch every way a stray formatting call could sneak back in; both were closed
before this story was finished.

It then proved the result: the same document was produced twice, by two independently started
programs, and the files were compared byte for byte. A fingerprint was recorded alongside the
compiler version that made it, so a future change to that fingerprint must be explained, not
quietly accepted.

What this story deliberately does not do: it renders one hard-coded rectangle and nothing else. No
text, no fonts, no data, no images, no page layout, no template, no compression. It proves nothing
about other machines — only that one machine repeats itself. Cross-machine proof is the next story.

Done looks like: a PDF a reader opens without complaint, two runs producing identical bytes, and a
fingerprint the test suite checks on every future change.

One thing will look wrong later and is not: a PDF containing a single rectangle is the point, not
an unfinished result.

---

## Story

As an integrating Go developer,
I want `folio-go` to emit a minimal valid PDF whose bytes are identical every time I render the same
input,
so that hash comparison becomes my regression test before any feature exists to regress.

---

## Settled decisions — apply these, do NOT re-open or surface them

Logged as **D-000.1 … D-000.5** in `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`.

| Setting | Value | Source |
|---|---|---|
| Go module path | `github.com/panitw/folio/folio-go` | D-000.5; Story 1.1 AC |
| Toolchain pin | **exactly `go1.26.0`** — not a range, not `go1.26.x`. Go 1.27 is deliberately **not** adopted: `compress/flate` byte-stability is verified only through 1.26, and adopting 1.27 is a re-measurement exercise under AD-22, not an upgrade | D-000.5; spine §Stack |
| Directory layout | Exactly the spine's §Source tree. `folio-go/`, `folio-designer/`, `fixtures/` are siblings of the existing `_bmad/`, `_bmad-output/`, `docs/` | D-000.5 |
| `fixtures/` location | **Repo root**, read by relative path at test runtime — **never** `go:embed`ed. Deliberate (AD-21): every future SDK conforms against the same bytes | D-000.5 |
| `fonts/` location | `folio-go/fonts/` — *inside* the module, because `go:embed` cannot reach outside it. **Not this story's concern; do not contradict it and do not create it** | D-000.5; AD-8 |
| Licence | `LICENSE` (MIT) at repo root, `Copyright (c) 2026 Panit Wechasil` | D-000.5; AD-26 |
| cgo | `CGO_ENABLED=0` everywhere | D-000.5; SPEC constraint |
| CI provider | **GitHub Actions**, `.github/workflows/`. Repo is public, so `macos-14` and `ubuntu-24.04-arm` runners are free. **Story 1.2 writes the matrix — 1.1 must merely not preclude it, and must not write a workflow file** | D-000.5 |
| AD-1 allow-list | `Sqrt`, `Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod` describe **what the import lint tolerates — not an invitation to use `float64`**, which stays banned under `internal/` by AD-2 and AD-23. Font scaling is round-half-to-even on an exact **integer** quotient, never a float rounding call | D-000.5 |
| The import lint | Story **1.3**, not this one. Do not write it here. But do not write code this story that the 1.3 lint would reject | epics.md §Story 1.3 |

---

## Corrections and measured findings (verified in this run, at `f2aa8c0`)

### F-1 — MEASURED: `go mod tidy` **deletes** `toolchain go1.26.0` when the `go` directive is also `1.26.0`

This is load-bearing and would otherwise be discovered as a silently-unpinned build.

Probe (run in a scratch module on this machine, Go 1.26.0 darwin/arm64):

```
module example.com/x        module example.com/y
go 1.26.0                   go 1.25.0
toolchain go1.26.0          toolchain go1.26.0

$ go mod tidy               $ go mod tidy
--- after ---               --- after ---
module example.com/x        module example.com/y
                            go 1.25.0
go 1.26.0                   toolchain go1.26.0
   ^ toolchain line GONE       ^ toolchain line SURVIVES
```

**Why:** Go treats a `toolchain` line that is not strictly greater than the `go` line as redundant
and removes it. With only `go 1.26.0` and `GOTOOLCHAIN=auto` (this machine's setting, verified via
`go env GOTOOLCHAIN`), a developer who installs Go 1.27 gets Go 1.27 — the exact hazard AD-22 exists
to prevent, invisible except in the hash.

**This story's instruction:** write

```
module github.com/panitw/folio/folio-go

go 1.25.0

toolchain go1.26.0
```

and add a test that asserts, from `go.mod`'s literal bytes, that a `toolchain go1.26.0` line is
present and exact. The `go` directive at `1.25.0` is a *language/minimum* floor; the `toolchain`
line is the actual pin and is what `GOTOOLCHAIN=auto` honours. See **DECISION NEEDED D-1.1.a** —
implement this arm, and if the lead rules otherwise the change is one line plus the fixture.

### F-2 — AD-3's "one emitter" cannot literally cover *every* number in the file

Object numbers, generation numbers, `/Length`, `/Size`, and the ten-digit zero-padded cross-reference
offsets are structural integers, not `geom.Length` values, and AD-3's stated rule ("converts a
`geom.Length` to its decimal text") does not describe them. Reading AD-3 as "literally every byte
that is a digit" would make the AC either impossible or vacuous.

**This story's reading:** two unexported helpers in `internal/pdf`, and only two.

1. The **geometric emitter** — `geom.Length` → sign + integer part + up to three trimmed fractional
   digits, by integer arithmetic only. This is AD-3's function. Every geometric number in the file
   goes through it and nothing else does.
2. A **structural integer helper** — non-negative `int` → decimal, by integer arithmetic only, for
   object numbers, `/Length`, `/Size` and xref offsets.

Neither may use `strconv.FormatFloat`, `fmt.Sprintf("%g"/"%v"/"%f")`, or any float at any point.
`strconv.AppendInt` / `strconv.Itoa` are exact integer operations and are permitted inside these two
helpers *only*. See **DECISION NEEDED D-1.1.b**.

### F-3 — `internal/geom` "imports nothing": the spine and AD-2 disagree slightly

The spine's §Source tree comment says `imports nothing (AD-2)`; AD-2's rule says
*"imports nothing outside the standard library's exact-integer surface"*. **Implement the strict
form:** `internal/geom`'s non-test files have an **empty import set**. `_test.go` files may import
`testing`. If a 128-bit intermediate is ever needed for scaling, that is a later story's problem and
a later decision — do not import `math/bits` here.

### F-4 — no previous story, no prior code, no prior patterns

`story_num == 1`, so there is no previous story file to mine. `git log` at `f2aa8c0` contains only
`f2aa8c0 Add epics, sprint status, and folio format spec` and `6976a39 Initial commit: Folio
planning artifacts and BMAD workspace`. Every convention in this file comes from the spine, the SPEC
or the decision log — none from precedent.

---

## Acceptance Criteria

Faithfully carried from `epics.md` lines 424–450 and numbered for task reference.

**AC1 — the module exists and the toolchain is pinned exactly.**
A Go module `github.com/panitw/folio/folio-go` exists at `folio-go/`, its `go.mod` carries a literal
`toolchain go1.26.0` line (exact version, no range), and that line survives `go mod tidy`. A test
asserts the line's presence and exact value by reading `go.mod`'s bytes.

**AC2 — no third-party PDF writer.**
`go.mod` has an empty `require` block. In particular no `signintech/gopdf`, no
`boxesandglue/baseline-pdf`, no `jung-kurt/gofpdf`, no `pdfcpu`, no `unidoc`. A test asserts the
module graph (`go list -m all`) contains exactly the main module.

**AC3 — `internal/geom` owns the geometric scalar, and it is the only owner.**
`internal/geom` declares `Length` as `int64` millipoints (1/1000 of a PDF point; 1 pt = 1/72 inch).
Every position, advance and dimension in the module is a `geom.Length`. `internal/geom` is the
**only** package declaring a geometric scalar type, and its non-test files import nothing (F-3).

**AC4 — one font-scaling function, round-half-to-even on the exact integer quotient.**
`internal/geom` exports exactly one scaling function. It computes on `int64` throughout and rounds
half-to-even (banker's rounding) on the exact integer quotient. There is no float anywhere in it and
no call to `math.Round`. It is not open-coded at any call site.

**AC5 — a valid PDF 1.7 document with one filled rectangle.**
`folio.Render` produces a single-page PDF 1.7 document containing a catalog, a page tree, a content
stream and a classic cross-reference table, with exactly one filled rectangle on the page. The
output is non-empty, has page count 1, and is accepted by an independent PDF reader.

**AC6 — every geometric number was written by the single unexported emitter.**
Every geometric number in the output — MediaBox entries and the rectangle's four operands — was
produced by one unexported emitter as sign + integer part + up to three fractional digits with
trailing zeros trimmed, computed by integer arithmetic. No other code writes a geometric number into
a content stream or object body. `strconv.FormatFloat`, `fmt.Sprintf` with `%g`/`%f`/`%v`, and
`fmt.Fprint*` of a numeric value appear nowhere under `folio-go/`.

**AC7 — no `float64` in any signature under `internal/`.**
No function, method, field, constant, variable or type parameter under `folio-go/internal/` has
`float64` (or `float32`) in its signature or declaration. A test enforces this by parsing the
packages with `go/ast`, not by grep.

**AC8 — byte-identity across two separate OS processes.**
The same input rendered by **two independently started OS processes** produces byte-identical
output. Not two calls in one process — a same-process comparison would pass on shared memoised
state. The comparison also asserts both outputs are non-empty, are valid PDFs, and have page
count 1, so equality cannot pass on two identical failures.

**AC9 — `/CreationDate` and `/ModDate` are absent.**
Neither key appears anywhere in the output bytes. No `/Info` dictionary is emitted at all in this
story.

**AC10 — `/ID` is content-derived, and both entries are identical.**
`/ID` is the first 16 bytes of a SHA-256 taken over the serialized body up to the point `/ID` is
written, rendered as a hex string, with both array entries byte-identical to each other.

**AC11 — the golden fixture matches, and records both versions.**
A golden fixture at repo-root `fixtures/` records the SHA-256 of the render, the `folio-go` version,
**and** the exact Go toolchain version. The test compares the live render's hash to the recorded one
and fails on mismatch. If `runtime.Version()` differs from the recorded toolchain version, the test
fails with a message stating that a toolchain bump is a versioned breaking change under AD-22 and
that the hash must be re-measured deliberately, **not** regenerated (counter-metric C6).

---

## Vacuity guards — no AC may be satisfied by asserting on these

1. **AC8 must not pass on two identical failures.** Two empty files, two zero-byte writes, or two
   identical error paths are byte-identical. AC8's assertion set must include non-empty, parseable,
   and page-count-1 before it compares.
2. **AC7 must not pass on an empty `internal/`.** The AST walk must assert it actually visited at
   least the `geom` and `pdf` packages and a non-zero number of declarations; a walk over zero files
   is trivially float-free.
3. **AC6 must not pass because nothing calls the emitter.** The assertion is on the *emitted bytes*
   — the content stream must literally contain the trimmed forms `72.5`, `100.25`, `200.125` and
   `50` (see Dev Notes) — not on the emitter's unit tests alone. A unit test of the emitter in
   isolation is necessary and not sufficient.
4. **AC11 must not be satisfied by regenerating the golden.** The fixture is written **once**, from
   the first passing render, and thereafter is an input. If it goes red, that is a defect until
   proven to be an intended, versioned change (C6, AD-21).
5. **AC1 must not be satisfied by a `toolchain` line that `go mod tidy` will delete** — see F-1. The
   test must run after a `go mod tidy` has been run, or the CI equivalent will erase it.
6. **AC2 must not be satisfied by a `require` block that is empty only because nothing imports
   anything yet.** State in the Delivery Log that the assertion is on the resolved module graph.

---

## Red-proofs — every new assertion names the mutation that reddens it

The developer must actually run each of these, observe the failure, revert, and record the failure
message in the Delivery Log. A red-proof that was reasoned about but not executed does not count.

| # | Assertion | Mutation that must redden it | Expected symptom |
|---|---|---|---|
| RP-1 | AC6 emitter format | Stop trimming trailing zeros — always emit three fractional digits | Content stream shows `50.000` / MediaBox shows `841.890`; golden hash mismatch |
| RP-2 | AC6 emitter format | Replace the emitter body with `strconv.FormatFloat(float64(l)/1000, 'f', -1, 64)` | AC6's "no `FormatFloat` under `folio-go/`" test reddens; golden may or may not move — **the source assertion is the one that must fire** |
| RP-3 | AC4 rounding | Change round-half-to-even to round-half-away-from-zero | The `2.5 → 2` and `-2.5 → -2` table rows fail (half-up gives 3 / -3); `7.5 → 8` does **not** move, so the table must contain a case where the two modes disagree |
| RP-4 | AC8 two-process identity | Derive `/ID` from `crypto/rand` instead of the content hash | The two subprocess outputs differ; the test names the byte offset of first divergence |
| RP-5 | AC8 two-process identity | Emit a `/CreationDate` from `time.Now()` | Both AC9 (key present) and AC8 (outputs differ) fire |
| RP-6 | AC7 no-float64 | Add `func scaleF(l Length, f float64) Length` to `internal/geom` | AST walk names the file, the declaration and the offending type |
| RP-7 | AC2 no third-party writer | `go get github.com/signintech/gopdf` | Module-graph assertion names the added module |
| RP-8 | AC11 toolchain recording | Delete the toolchain field from the fixture JSON | Fixture-shape assertion fails before the hash comparison runs |
| RP-9 | AC1 toolchain pin | Change `go 1.25.0` to `go 1.26.0` and run `go mod tidy` | The `toolchain` line vanishes and the go.mod assertion fails — this is F-1, restated as a guard |
| RP-10 | AC10 `/ID` identity | Make the second `/ID` array entry a different 16 bytes | The both-entries-identical assertion fires |

---

## Tasks / Subtasks

- [x] **Task 1 — Repository skeleton and licence** (AC: 1)
  - [x] Create `LICENSE` at repo root: MIT, `Copyright (c) 2026 Panit Wechasil` (AD-26).
  - [x] Create the directories this story actually needs, and **only** those:
        `folio-go/`, `folio-go/internal/geom/`, `folio-go/internal/pdf/`, `fixtures/`.
        **Do not** create empty placeholder packages for `template`, `bind`, `layout`, `text`,
        `fontset`, `pagemodel`, `diag`, `expr`, `cmd/folio`, `wasm/`, `folio-go/fonts/`, or
        `folio-designer/`. Empty packages are noise and would make Story 1.3's lint vacuous. Every
        directory that *is* created must sit exactly where the spine's §Source tree puts it.
  - [x] **Do not** write anything under `.github/` — that is Story 1.2.
  - [x] Confirm `.gitignore` does not exclude `fixtures/` or `folio-go/`. Note that `go.work` is
        gitignored, so this must remain a single module with no workspace file.

- [x] **Task 2 — Module and the toolchain pin** (AC: 1, 2; red-proofs RP-7, RP-9)
  - [x] `go mod init github.com/panitw/folio/folio-go` in `folio-go/`.
  - [x] Set `go 1.25.0` and `toolchain go1.26.0` per **F-1**. Run `go mod tidy` and confirm the
        `toolchain` line survives.
  - [x] Add `folio-go/version.go`: `package folio; const Version = "0.0.0-dev"`. This is the value
        AC11's fixture records; a real tag is `folio-go/v0.1.0` (directory-prefixed, AD-22) and is
        not this story's business.
  - [x] Test: read `go.mod` bytes, assert a line exactly `toolchain go1.26.0` exists.
  - [x] Test: `go list -m all` returns exactly the main module (AC2). Run RP-7, observe, revert.
  - [x] Run RP-9, observe the toolchain line disappear, revert.

- [x] **Task 3 — `internal/geom`** (AC: 3, 4, 7; red-proofs RP-3, RP-6)
  - [x] `type Length int64` with a doc comment stating: millipoints, 1/1000 pt, 1 pt = 1/72 inch,
        and that this is the module's **only** geometric scalar type (AD-2).
  - [x] Add `Rect` (four `Length` fields) — needed by the rectangle and by AD-2's "one owner".
  - [x] One exported scaling function. Suggested shape:
        `func ScaleRound(v Length, num, den int64) Length` — computes `v*num/den` on `int64`
        throughout, rounding **half-to-even** on the exact integer quotient, correct for negative
        `v`, panicking on `den == 0`. Document the rounding mode on the function.
  - [x] Table test for `ScaleRound` including cases where half-to-even and half-away-from-zero
        **disagree**: `(5,1,2) → 2`, `(-5,1,2) → -2`, `(15,1,2) → 8`, `(-15,1,2) → -8`, plus exact
        cases and a `den == 0` panic case. Run RP-3, observe, revert.
  - [x] Confirm the package's non-test files have an **empty import set** (F-3).
  - [x] Test (in `internal/geom` or a dedicated `internal/arch_test.go`): parse every package under
        `folio-go/internal/` with `go/ast` and fail on any `float64`/`float32` in a declaration or
        signature. Guard against vacuity per guard 2 — assert the walk visited ≥ 2 packages and a
        non-zero declaration count. Run RP-6, observe the named file/decl, revert.

- [x] **Task 4 — the number emitter** (AC: 6; red-proofs RP-1, RP-2)
  - [x] In `internal/pdf`, one **unexported** emitter converting `geom.Length` → decimal text by
        integer arithmetic only: sign, integer part, up to three fractional digits, trailing zeros
        trimmed, no trailing `.`, and never the string `-0`.
  - [x] Table test covering: `0 → "0"`, `50000 → "50"`, `72500 → "72.5"`, `100250 → "100.25"`,
        `200125 → "200.125"`, `595276 → "595.276"`, `841890 → "841.89"`, `-500 → "-0.5"`,
        `-1 → "-0.001"`, `1 → "0.001"`, and `math.MinInt64`-adjacent bounds. **Beware:** Go integer
        division truncates toward zero, so the negative cases must take the absolute value first —
        this is the most likely defect in the whole task.
  - [x] Second unexported helper for structural integers per **F-2**, with its own test.
  - [x] Source-level test: assert `strconv.FormatFloat`, `fmt.Sprintf`, `fmt.Fprintf`, `%g`, `%f`
        and `%v` appear nowhere under `folio-go/` outside `_test.go` files. Run RP-2, observe,
        revert.
  - [x] Run RP-1, observe the golden mismatch, revert.

- [x] **Task 5 — `internal/pdf` serializer** (AC: 5, 9, 10; red-proofs RP-5, RP-10)
  - [x] Emit exactly the document specified in **Dev Notes → The document, byte by byte**. No
        compression, no `/Info`, no `/Filter`, LF line endings only.
  - [x] `/ID`: SHA-256 over every byte emitted from `%PDF-1.7` up to (not including) the `/ID` key in
        the trailer; take the first 16 bytes; write them as `<` + 32 upper-case hex + `>`, twice,
        identically. Note the non-circularity argument in Dev Notes and restate it in a code comment.
  - [x] Test: `/CreationDate` and `/ModDate` byte-substrings absent (AC9). Run RP-5, observe both
        AC9 and AC8 fire, revert.
  - [x] Test: the two `/ID` array entries are byte-identical and 16 bytes each. Run RP-10, observe,
        revert.

- [x] **Task 6 — `folio.Render`** (AC: 5)
  - [x] `folio-go/folio.go`: `func Render() ([]byte, error)`, doc-commented
        **`// PROVISIONAL — Story 1.1 only.`** with a note that Story 1.4 gives it a template, 1.6
        data and params, and 1.7 the `RenderTo(w io.Writer, …)` writer form (which must arrive
        without an API break, per the spine's Deferred row). See **DECISION NEEDED D-1.1.c**.
  - [x] The hard-coded page and rectangle live behind this call. `folio` may import
        `internal/geom` and `internal/pdf`; `internal/pdf` may import `internal/geom`; nothing else.
  - [x] Validate the output with an **independent** reader — `qpdf --check` or `mutool info` if
        available, otherwise a `pdfjs-dist` or Python `pypdf` one-liner. Record the tool, its
        version and its exact output in the Delivery Log. Do **not** add a PDF library to `go.mod`
        for this (AC2); use an external CLI or a throwaway script outside the module.

- [x] **Task 7 — two-process byte-identity harness** (AC: 8; red-proofs RP-4, RP-5)
  - [x] Implement as a self-re-executing test binary, **not** two calls in one process:
        a `TestMain` (or an early check in the test) that, when `FOLIO_SUBPROCESS_RENDER=1` is set,
        calls `folio.Render()`, writes the bytes to stdout and exits 0. The determinism test then
        runs `exec.Command(os.Args[0], "-test.run=…")` **twice**, each with that env var set.
  - [x] Run child A with `GOMAXPROCS=1` and child B with `GOMAXPROCS=4`, so the comparison also
        catches order-dependence introduced by concurrency.
  - [x] Before comparing: assert each output is non-empty, starts with `%PDF-1.7`, ends with
        `%%EOF`, and has exactly one `/Type /Page` object (vacuity guard 1).
  - [x] On mismatch, report the byte offset of first divergence and a short hex window around it —
        every future determinism failure in this project will be diagnosed with this message.
  - [x] Run RP-4, observe, revert.

- [x] **Task 8 — the golden fixture** (AC: 11; red-proof RP-8)
  - [x] Create `fixtures/minimal-rect/` containing:
        `expected.json` — `{"folioGoVersion": "0.0.0-dev", "goToolchain": "go1.26.0",
        "sha256": "<64 hex>"}`, keys sorted, two-space indent, LF, trailing newline;
        `expected.pdf` — the recorded bytes, for human diffing only (the hash is normative);
        `README.md` — what the fixture is, and that a hash change is investigated as a defect until
        proven to be an intended versioned change (C6, AD-21).
  - [x] Read the fixture from the test by **relative path**, never `go:embed` (D-000.5, AD-21). Find
        the repo root by walking up from the test file's directory until a directory containing both
        `folio-go/` and `fixtures/` is found. Put this helper in a `_test.go` file — a non-test
        helper importing `os` under `internal/` would violate AD-1 and Story 1.3's lint.
  - [x] The test must (a) assert the fixture carries **both** version fields before comparing, and
        (b) fail with the AD-22 / C6 message if `runtime.Version()` != the recorded toolchain.
  - [x] Record the fixture **once**, from the first passing render. Run RP-8, observe, revert.

- [x] **Task 9 — full verification sweep**
  - [x] `CGO_ENABLED=0 go build ./...`, `CGO_ENABLED=0 go vet ./...`, `gofmt -l .` (empty),
        `CGO_ENABLED=0 go test ./... -count=1`. Record every command verbatim with its output in the
        Delivery Log.
  - [x] Per **D-000.4**, story 1.1 is *not* a hash-matrix override story. Do **not** run the
        four-target Docker/Node matrix here; state explicitly in the Delivery Log:
        *"cross-target matrix not run — per-epic cadence, due at Epic 1 close (Story 1.2 builds it)."*
  - [x] Confirm `git status --porcelain` contains only the files this story intends to add, and that
        every red-proof mutation has been reverted.
  - [x] Record every red-proof's observed failure message in the Delivery Log.

- [x] **Task 10 — hand off**
  - [x] Fill in `## Delivery Log` below.
  - [x] Append any decisions taken during the run to
        `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` as `D-1.1.n` entries.
  - [x] Set `sprint-status.yaml` → `1-1-a-minimal-pdf-reproducible-on-one-machine: review` and
        `epic-1: in-progress`.
  - [x] **Stop here — do not commit, do not branch, do not set `done`.**

---

## Dev Notes

### The document, byte by byte

Fix these exactly. They are a throwaway fixture for this story, not a page-size system — real page
setup arrives in Story 1.4.

- **Line endings:** `\n` only, everywhere. Never `\r\n`.
- **No compression.** No `/Filter`, no `FlateDecode`. This removes `compress/flate` from the variable
  set entirely for Story 1.1. Compression arrives later and is a hash-changing, versioned event
  under AD-22 (and is the subject of carried risk **R4**).
- **Header:** `%PDF-1.7\n` followed by a binary comment line of four bytes ≥ 128, then `\n`.
- **Objects**, in this order, each `N 0 obj\n… \nendobj\n`:
  1. Catalog — `<< /Type /Catalog /Pages 2 0 R >>`
  2. Pages — `<< /Type /Pages /Kids [3 0 R] /Count 1 >>`
  3. Page — `<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595.276 841.89] /Resources << >> /Contents 4 0 R >>`
  4. Content stream — `<< /Length N >>\nstream\n…\nendstream`
- **MediaBox** is A4 as the literal millipoint constants `0, 0, 595276, 841890`. These are
  fixture constants declared once as `geom.Length`; **there is no runtime mm→pt conversion in this
  story** — a conversion would be a rounding decision this story has no mandate to make.
- **Content stream**, exactly:
  ```
  72.5 100.25 200.125 50 re
  f
  ```
  from `geom.Length` values `72500, 100250, 200125, 50000`. This rectangle is chosen deliberately:
  it exercises three, two, one and zero fractional digits in one line, so RP-1's untrimmed mutation
  reddens visibly. **Emit no colour operator** — the PDF default fill colour is black, so the only
  numbers in the content stream are the four rectangle operands. Adding `0 g` would be a number that
  is not a `geom.Length`, i.e. an AD-3 carve-out for no benefit.
- **Cross-reference table:** classic (not a stream), `xref\n0 5\n`, free entry
  `0000000000 65535 f \n`, then four in-use entries `%010d 00000 n \n`. Note the trailing space
  before the newline — each entry is exactly 20 bytes.
- **Trailer:** `trailer\n<< /Size 5 /Root 1 0 R /ID [<…><…>] >>\nstartxref\n<offset>\n%%EOF\n`.

### Why `/ID` is not circular

The trailer is written *after* the cross-reference table, so `startxref`'s offset is already fixed
before `/ID` is computed. `/ID`'s value is a fixed 16 bytes → a fixed-width 32-hex string, so
writing it cannot shift any offset recorded earlier in the file. The digest input is every byte from
`%PDF-1.7` up to but not including the literal `/ID`. Put this paragraph in a code comment; it is
the kind of reasoning that gets lost and then "fixed" wrongly.

### The emitter, stated precisely

For `v geom.Length`, let `n = |v|`, `i = n / 1000`, `f = n % 1000` (both exact integer ops).
Emit `-` if `v < 0` **and** the result is not zero; emit `i`; if `f != 0`, emit `.` then `f`
zero-padded to three digits with trailing zeros stripped. Never emit `-0`, never a bare trailing `.`.
Worked: `841890 → "841.89"`, `595276 → "595.276"`, `50000 → "50"`, `-500 → "-0.5"`, `0 → "0"`.

### Things this story must not do

No text, no fonts, no `go:embed`, no images, no template parsing, no data binding, no diagnostics
registry, no expression language, no CLI, no wasm entry point, no CI workflow, no designer, no
`internal/layout`. No `time`, `os`, `math/rand`, `net` or any `math` transcendental in any non-test
file under `internal/` (AD-1) — and no package-level mutable state, and no `range` over a map whose
order could reach an output byte (NFR1.d). The AD-1 *lint* is Story 1.3, but the *rule* binds now.

The one arrow that matters most in this architecture is the one that does not exist:
**`internal/layout` must never import `internal/pdf`** (AD-5). Neither package exists yet — do not
create the habit that breaks it.

### Testing standards

Table-driven Go tests (spine §Consistency Conventions). Every feature ships its golden fixture
(AD-21). `_test.go` files are exempt from AD-1's `os` ban — they must be, or repo-root `fixtures/`
becomes unreadable and AD-21 unimplementable (D-000.5 §Consequences). Nothing under `internal/`
panics on malformed input, except where a programmer error is unambiguous (`den == 0` in Task 3).

### Project Structure Notes

Created by this story, all at the spine's §Source tree positions:

```
folio/
  LICENSE                                   # NEW — MIT, AD-26
  folio-go/                                 # NEW — module github.com/panitw/folio/folio-go
    go.mod                                  #   go 1.25.0 + toolchain go1.26.0  (F-1)
    folio.go                                #   Render() — PROVISIONAL, Story 1.1 only
    version.go                              #   const Version = "0.0.0-dev"
    internal/
      geom/                                 #   Length (int64 millipoints), Rect, ScaleRound — imports nothing
      pdf/                                  #   object serializer, xref, the two number helpers
  fixtures/                                 # NEW — repo root, read by relative path, never embedded
    minimal-rect/                           #   expected.json · expected.pdf · README.md
```

Deliberately **not** created: `folio-go/fonts/`, `folio-go/cmd/`, `folio-go/wasm/`,
`folio-designer/`, `.github/`, and the eight other `internal/` packages. Each belongs to a later
story and an empty stub would make Story 1.3's lint vacuous.

**Known variance:** the spine and D-000.5 say "exactly the spine's §Source tree". This story creates
that tree's *shape* but only its *needed* nodes. Flagged here rather than silently.

### References

- Story text and ACs — `_bmad-output/planning-artifacts/epics.md#L411-450` (Epic 1 preamble at 411,
  Story 1.1 at 415)
- AD-1 determinism boundary — `…/ARCHITECTURE-SPINE.md#L92`
- AD-2 one fixed-point unit — `…/ARCHITECTURE-SPINE.md#L105`
- AD-3 one number emitter — `…/ARCHITECTURE-SPINE.md#L119`
- AD-5 layout must not import pdf — `…/ARCHITECTURE-SPINE.md#L146`
- AD-6 Folio emits its own PDF — `…/ARCHITECTURE-SPINE.md#L156`
- AD-7 PDF profile pinned, `/ID`, no dates — `…/ARCHITECTURE-SPINE.md#L174`
- AD-21 every feature ships its golden fixture — `…/ARCHITECTURE-SPINE.md#L363`
- AD-22 toolchain pinned, bump is a release event — `…/ARCHITECTURE-SPINE.md#L375`
- AD-23 no `float64` for data — `…/ARCHITECTURE-SPINE.md#L388`
- AD-26 MIT, nothing copyleft — `…/ARCHITECTURE-SPINE.md#L445`
- Source tree — `…/ARCHITECTURE-SPINE.md#L568`; Stack (Go 1.26.x, Go 1.27 not adopted) — `#L481`
- NFR1.a–NFR1.g — `_bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md#L328-378`
- FR29, FR43 — `…/prd.md#L269`, `…/prd.md#L309`
- S1–S3, C3, C6, risk R1 — `_bmad-output/specs/spec-folio/acceptance.md`
- CAP-13 byte-reproducible rendering — `_bmad-output/specs/spec-folio/SPEC.md`
- D-000.1 … D-000.5 — `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`

---

## DECISION NEEDED

Three items, all with a recommended arm already implemented in the tasks above. **Implement the
recommendation; if the lead rules otherwise, each is a small, localized change.** Do not treat these
as licence to choose differently on your own.

### D-1.1.a — the `go` directive must sit below `1.26.0` for the toolchain pin to survive

**Measured, F-1 above.** `go mod tidy` deletes `toolchain go1.26.0` when `go` is also `1.26.0`, and
`GOTOOLCHAIN=auto` then silently uses a newer local toolchain — exactly the AD-22 hazard.

- **(a) Recommended, implemented:** `go 1.25.0` + `toolchain go1.26.0`. The toolchain line survives
  tidy and is honoured by `GOTOOLCHAIN=auto`. Cost: the `go` directive understates the requirement,
  and the module claims a 1.25 language floor it does not need.
- **(b)** `go 1.26.0` alone, pin via `GOTOOLCHAIN=go1.26.0` in CI and a `.go-version` file. Cost: the
  pin lives outside `go.mod`, contradicting AD-22's letter ("`go.mod` pins an exact `toolchain`
  directive"), and a developer building locally is unpinned.
- **(c)** Both — (a) plus `GOTOOLCHAIN=go1.26.0` exported in CI as belt-and-braces. Cost: none
  material; Story 1.2 would carry the CI half.

Recommendation: **(a) now, and (c) once Story 1.2 exists.**

### D-1.1.b — AD-3's "exactly one function" versus structural integers

**F-2 above.** `/Length`, `/Size`, object numbers and ten-digit xref offsets are not `geom.Length`
values. Reading AD-3 literally makes the AC impossible; ignoring the question invites a third
formatter later.

- **(a) Recommended, implemented:** exactly two unexported helpers — the geometric emitter (AD-3's)
  and a structural-integer helper — both integer-only, both float-free, and an explicit
  source-level assertion that no third numeric formatter exists.
- **(b)** Force structural integers through the `geom.Length` emitter by scaling them by 1000. Cost:
  meaningless, and an xref offset is not a length.

### D-1.1.c — the shape of `Render` for Story 1.1

The AC names `Render`, but the real signature arrives across Stories 1.4, 1.6 and 1.7.

- **(a) Recommended, implemented:** exported `func Render() ([]byte, error)`, doc-commented
  `PROVISIONAL — Story 1.1 only`. Cost: the public API changes shape three times before v0.1.0 — all
  at `0.0.0-dev`, so no downstream user is affected.
- **(b)** Keep it unexported for now and expose only a test hook. Cost: AC5 says "`Render` is
  called"; an unexported one arguably does not satisfy it, and C5 (time-to-first-PDF) wants the name
  to exist early.

---

## Delivery Log

### Decisions applied

D-1.1.a, D-1.1.a (addendum), D-1.1.b, D-1.1.c and D-1.1.c (addendum) were already settled in
`folio-mvp-decision-log.md` before this run started and are applied verbatim — recommendation (a)
for each, exactly as written in the story's own `## DECISION NEEDED` section. No new decision was
required during this run: every implementation choice not covered by a settled decision (the exact
four bytes of the header's binary comment; `appendIntPadded`'s zero-fill applying to the digit run
rather than to sign+digits) was a mechanical detail with one obviously-correct shape and no
downstream consequence, not a genuine design/scope fork, so nothing was escalated.

### Spine amendment (D-000.6)

AD-3's `- **Rule:**` paragraph in `ARCHITECTURE-SPINE.md` was replaced with the two-representation
text specified for this story (decimal emitter + integer writer). Only the Rule paragraph changed;
Binds and Prevents are untouched.

**Finisher addendum (QA review Nit 28):** D-000.6 consequence (3) requires the before/after to be
quoted verbatim in this log; only the resulting text was quoted in the original Delivery Log entry.
The superseded paragraph, quoted here for the record:

> - **Rule:** `internal/pdf` contains exactly one unexported function that converts a `geom.Length`
>   to PDF decimal text — sign, integer part, up to three fractional digits, trailing zeros
>   trimmed — by integer arithmetic only. No other code in the module writes a number into a
>   content stream or object body.

**Finisher addendum (QA review Major 12):** the section heading directly above this Rule
paragraph — `### AD-3 — Numbers reach the PDF through exactly one function` — was left unchanged
by the original amendment and contradicted the replacement Rule beneath it (which specifies two
representations across three functions). The finisher amended the heading in this commit, as part
of the same D-000.6 amendment, to `### AD-3 — Numbers reach the PDF through exactly one file, in
exactly two representations`.

Full resulting AD-3 text (Rule paragraph and heading, current as of this commit):

> - **Rule:** `internal/pdf` emits numbers in exactly **two** representations, and no number
>   reaches an output byte by any other route.
>   - **Decimal.** One unexported emitter converts a thousandths-scaled `int64` — every
>     `geom.Length`, and every other value defined in thousandths — to decimal text by integer
>     arithmetic: sign, integer part, exactly three fractional digits, trailing zeros and a bare
>     trailing point trimmed. Exact by construction, since a millipoint is exactly three decimal
>     places of a point. No rounding occurs here; rounding occurs only where a value is converted
>     *into* thousandths, and always round-half-to-even (AD-2).
>   - **Integer.** One unexported writer converts an `int64` to plain decimal digits with no
>     scaling, optionally zero-filled to a fixed width. It carries every count, byte length, file
>     offset, object and generation number, pixel dimension, and glyph metric in 1000-unit em
>     space — values whose text form has no platform variance and for which the decimal emitter
>     would be actively wrong (an xref offset of 12345 bytes is not 12.345).
>
>   A value that is neither is converted into one of the two by integer arithmetic before emission
>   and never formatted on its own terms: an 8-bit colour component becomes thousandths
>   (`c*1000/255`, round-half-to-even) and takes the decimal route. A glyph id under `Identity-H`
>   is a big-endian hex pair inside a string literal — a byte encoding, not a number, and it takes
>   neither route.
>
>   Both live in one file in `internal/pdf`. Because no package outside `internal/pdf` writes an
>   output byte at all (AD-5), the restriction needs to police only that package: CI fails any
>   reference to `strconv.Format*`, `strconv.Itoa`, `strconv.Append*`, or a `fmt` formatting call
>   in any other file of `internal/pdf`. Number formatting inside a diagnostic message is not an
>   output byte and is not covered.

Routing implemented in `internal/pdf/numbers.go`: `appendLength` for **all four** MediaBox entries
and the four rectangle operands; `appendInt` for `/Length`, `/Size`, `/Count`, and every object and
generation number (including inside indirect references like `2 0 R`); `appendIntPadded` for the
10-digit xref offsets and 5-digit generation field, including the free entry. `appendIntPadded`
delegates to `appendInt` and left-pads — exactly one place produces integer digits.

**Finisher addendum (QA review Blocker 1, Major 8):** the original Delivery Log's "Routing
implemented" line above was itself inaccurate on two counts, both now corrected. First, only the
rectangle's four operands and the MediaBox's upper-right corner were routed through `appendLength`
in the shipped code — the MediaBox's lower-left corner (`llx`, `lly`) was written as literal `"0 0
"` bytes inside a dict string constant, never touched by any emitter; the output was correct only
because `appendLength(0)` also happens to produce `"0"`. Second, `/Size`, `/Count`, and every
object/generation number inside an indirect reference (`/Pages 2 0 R`, `/Kids [3 0 R]`, `/Parent 2
0 R`, `/Contents 4 0 R`, `/Root 1 0 R`, and each object header's own `N 0 obj`) were hard-coded
literal text, contradicting D-1.1.b's routing table despite the Delivery Log's claim that the
table was "still exercise[d]... in all three functions." `internal/pdf/document.go` now routes
every one of these through `appendLength`/`appendInt`/`appendIntPadded` as named above, verified
by re-running the full suite and re-confirming the golden hash is byte-identical
(`0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c`) before and after the change —
the refactor changed which code produces the bytes, not the bytes themselves.

### Acceptance criteria — how each was verified

- **AC1** — `folio-go/go.mod` carries literal `go 1.25.0` and `toolchain go1.26.0` lines (with an
  explanatory comment recorded in the file itself). `TestGoModPinsToolchainExactly` reads the raw
  bytes and asserts both lines. `go mod tidy` was run after writing the lines; the toolchain line
  survived (verified again for real via RP-9 below).
- **AC2** — `go.mod`'s `require` block is empty. `TestModuleGraphHasNoThirdPartyDependencies` runs
  `go list -m all` and asserts exactly one line (the main module). Verified against a real added
  dependency via RP-7 below — the assertion is on the resolved module graph, not on nothing
  importing anything yet (vacuity guard 6).
- **AC3** — `internal/geom.Length` is `int64` millipoints; `internal/geom.Rect` groups four
  `Length`s; no other package under `folio-go/` declares a geometric scalar. `internal/geom`'s
  non-test files (`geom.go`, `scale.go`) import nothing — confirmed by reading the files (empty
  import blocks) and by `go vet`/`go build` succeeding with no import lines present.
- **AC4** — `geom.ScaleRound` computes on `int64` only, rounds half-to-even on the exact integer
  quotient, panics on `den == 0`. Table test in `scale_test.go` includes the four
  half-to-even/half-away-from-zero-disagreement cases from the story
  (`(5,1,2)→2`, `(-5,1,2)→-2`, `(15,1,2)→8`, `(-15,1,2)→-8`) plus exact cases, ordinary rounding
  cases and a second even/odd tie pair. RP-3 (below) confirms the disagreement cases actually
  distinguish the rounding mode.
- **AC5** — `folio.Render()` produces a 547-byte PDF 1.7 document: catalog (obj 1), page tree
  (obj 2, `/Count 1`), one page (obj 3), one content stream (obj 4) with `re`/`f`, and a classic
  (non-stream) xref table. `TestRenderProducesAValidPDF` checks non-empty, `%PDF-1.7` header,
  `%%EOF` trailer, exactly one `/Type /Page` object. Independently validated with `qpdf --check`
  (installed via Homebrew for this run, `qpdf version 12.4.0`) — see command output below.
- **AC6** — every geometric number is produced by `appendLength`. The rendered content stream
  literally contains `72.5 100.25 200.125 50 re` (verified by inspecting the raw bytes — see below)
  and the page's `MediaBox` shows `595.276 841.89`. `TestNoOtherNumberFormattingUnderPDF` asserts
  `strconv.FormatFloat`/`fmt.Sprintf`/`fmt.Fprintf`/`%g`/`%f`/`%v` appear nowhere under `folio-go/`
  outside `_test.go` files (comments are stripped via `go/parser`+`go/format` before matching, so
  the emitter's own doc comment — which names these forbidden calls to explain why they're
  forbidden — does not trip its own guard). RP-1 and RP-2 (below) both fired as specified.
- **AC7** — `internal/arch_test.go` (`package arch`, the only file in `folio-go/internal/`) walks
  every `.go` file under `internal/` with `go/ast` and fails on any `float64`/`float32` in a
  declaration or signature. Vacuity guard: it asserts it visited ≥ 2 package directories (`geom`,
  `pdf`) and a non-zero declaration count before concluding "no float". RP-6 (below) confirms it
  actually reddens on a real float64 parameter.
- **AC8** — `folio-go/render_test.go`'s `TestMain` re-executes the test binary as a subprocess when
  `FOLIO_SUBPROCESS_RENDER=1` is set;
  `TestRenderIsByteIdenticalAcrossTwoProcesses` runs `exec.Command(os.Args[0], "-test.run=^TestMain$")`
  twice — genuinely two separate OS processes, not two in-process calls — child A with
  `GOMAXPROCS=1`, child B with `GOMAXPROCS=4`. Each output is asserted non-empty/well-formed/
  page-count-1 *before* the byte comparison (vacuity guard 1). RP-4 and RP-5 (below) both fired.
- **AC9** — `TestRenderHasNoCreationOrModDate` asserts `/CreationDate`, `/ModDate` and `/Info` are
  all absent from the output bytes. RP-5 (below) confirms this fires on a real `time.Now()`-derived
  value.
- **AC10** — `/ID` is computed in `internal/pdf/document.go`'s `computeID`: SHA-256 over every byte
  written from `%PDF-1.7` up to (not including) the trailer's `/ID` key, first 16 bytes, upper-case
  hex, written twice identically. `TestRenderIDEntriesAreIdentical` parses both `<...>` entries out
  of the rendered bytes and asserts both are 32 hex characters and byte-identical. RP-10 (below)
  confirms this fires on a real divergence.
- **AC11** — `fixtures/minimal-rect/expected.json` records `folioGoVersion: "0.0.0-dev"`,
  `goToolchain: "go1.26.0"`, and the SHA-256 of the first passing render
  (`0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c`), recorded once and never
  regenerated to force a pass. `fixtures/minimal-rect/expected.pdf` holds the same bytes for human
  diffing (hash independently reconfirmed against it with a standalone `sha256` check). `folio-go`'s
  `TestRenderMatchesGoldenFixture` reads the fixture by relative path (never `go:embed`), asserts
  both version fields are present *before* comparing hashes (vacuity guard), fails with the AD-22/C6
  message on a `runtime.Version()` mismatch, and only then compares the hash. RP-8 (below) confirms
  the toolchain-field check fires before any hash comparison runs.

### Red-proofs — actually run, observed, reverted

Every mutation below was applied to the real file, run for real, observed to fail with the message
shown, then reverted with `cp` from a pre-mutation backup (not `git checkout`, since these files
are untracked) and re-confirmed green.

- **RP-1** (stop trimming trailing zeros in `appendLength`): `go test ./internal/pdf/... -run
  TestAppendLength -count=1` →
  `appendLength(50000) = "50.000", want "50"`, `appendLength(841890) = "841.890", want "841.89"`
  (plus 6 more rows). `go test . -run TestRenderMatchesGoldenFixture -count=1` also failed:
  `golden fixture mismatch: got sha256 7b269745...a3d529, want 0f925e1b...cb4f7c`. Reverted; full
  suite green again.
- **RP-2** (replace `appendLength`'s body with `strconv.FormatFloat(float64(l)/1000, 'f', -1,
  64)`): `go test ./internal/pdf/... -run TestNoOtherNumberFormattingUnderPDF -v` →
  `forbidden float/format calls found outside _test.go files (AC6): folio-go/internal/pdf/
  numbers.go: matched strconv\.FormatFloat`. Reverted; green.
- **RP-3** (round-half-away-from-zero instead of half-to-even): `go test ./internal/geom/... -run
  TestScaleRound -v` → `ScaleRound(5, 1, 2) = 3, want 2`, `ScaleRound(-5, 1, 2) = -3, want -2`, and
  — as the story specifies the table must contain a disagreeing case that does *not* move —
  `ScaleRound(9, 1, 2) = 5, want 4` also failed while the `(15,1,2)→8` row stayed green
  (half-away-from-zero also gives 8 there). Reverted; green.
- **RP-4** (derive `/ID` from `crypto/rand` instead of the content hash): `go test . -run
  TestRenderIsByteIdenticalAcrossTwoProcesses -v` →
  `subprocess outputs differ (len a=547, len b=547); first divergence at byte offset 455;
  a=52202f4944205b3c3342413844373346 b=52202f4944205b3c3732383237433944` — offset 455 is inside the
  `/ID` value, as expected. Reverted; green.
- **RP-5** (emit a `time.Now()`-derived marker after `%%EOF`): `go test . -run
  "TestRenderHasNoCreationOrModDate|TestRenderIsByteIdenticalAcrossTwoProcesses" -v` → AC9 fired
  (`output contains /CreationDate`) and AC8 fired (`child A (GOMAXPROCS=1): output does not end
  with %%EOF`) — both, as the row specifies. Reverted; green.
- **RP-6** (add `func scaleF(l Length, f float64) Length` to `internal/geom`): `go test
  ./internal/... -run TestNoFloat64UnderInternal -v` → `float64/float32 found under internal/
  (forbidden by AD-23): geom/scale.go:50: func scaleF parameter in scaleF uses float64` — names the
  file, the declaration and the offending type, as specified. Reverted; green.
- **RP-7** (`go get github.com/signintech/gopdf`): `go test . -run
  TestModuleGraphHasNoThirdPartyDependencies -v` → `expected exactly one module in the graph (the
  main module), got 4: [github.com/panitw/folio/folio-go github.com/phpdave11/gofpdi
  v1.0.14-0.20211212211723-1f10f9844311 github.com/pkg/errors v0.8.1
  github.com/signintech/gopdf v0.38.0]`. Reverted `go.mod`, deleted the generated `go.sum`, ran
  `go mod tidy`; green, no `go.sum` present (matches the pre-mutation state — this module has no
  dependencies at all).
- **RP-8** (delete `goToolchain` from `fixtures/minimal-rect/expected.json`): `go test . -run
  TestRenderMatchesGoldenFixture -count=1 -v` → `fixture is missing goToolchain (RP-8: this must
  fail before the hash comparison runs)` — fired before any hash comparison, as required. Reverted;
  green.
- **RP-9** (change `go 1.25.0` to `go 1.26.0` in `go.mod`, run `go mod tidy`): the `toolchain
  go1.26.0` line (and its explanatory comment) vanished from `go.mod` after `tidy`, exactly as F-1
  predicted. Reverted from a pre-mutation backup, ran `go mod tidy` again; `toolchain go1.26.0`
  survived.
- **RP-10** (make the second `/ID` array entry differ from the first): `go test . -run
  TestRenderIDEntriesAreIdentical -v` → `/ID entries differ: "C6086F83998B551A85031BCAB15DACC6" !=
  "C6086F83998B551A85031BCAB15DACC0"`. Reverted; green.

### Independent PDF-reader validation (AC5)

`qpdf` was not present on this machine and was installed via `brew install qpdf` for this run
(`qpdf version 12.4.0`, no PDF library added to `go.mod` — AC2 unaffected, this is an external
CLI). Output of `qpdf --check` against the live render:

```
$ qpdf --check /tmp/out.pdf
checking /tmp/out.pdf
PDF Version: 1.7
File is not encrypted
File is not linearized
No syntax or stream encoding errors found; the file may still contain
errors that qpdf cannot detect
```

Raw bytes of the render were also inspected directly (Python `repr()` of the file), confirming the
document byte-for-byte against the Dev Notes spec: `%PDF-1.7\n%\xe2\xe3\xcf\xd3\n`, four objects in
order, content stream `72.5 100.25 200.125 50 re\nf\n` with `/Length 28` (28 is correct: the
content stream's exact byte count), classic xref with 20-byte entries, and a trailer whose two
`/ID` entries were identical 32-hex-character strings.

### Full verification sweep (Task 9)

**Finisher addendum (QA review Nit 23):** the original entry below claimed "30 individual test
functions/subtests." The reviewer independently counted 16 top-level functions + 13 subtests = 29
against the pre-finisher code. Both numbers are now superseded: the finisher's fixes added tests
(the ScaleRound boundary/overflow rows and panic tests, the AC10 provenance tests, and the merged
D-1.1.b file-scoped guardrail replacing the old regex-based one), and the true count as of this
commit — actually run, not estimated — is recorded below.

All run from `folio-go/`, real commands, real output, re-verified after every finisher change:

```
$ CGO_ENABLED=0 GOWORK=off go build ./...
(no output — success)

$ CGO_ENABLED=0 GOWORK=off go vet ./...
(no output — success)

$ gofmt -l .
(no output — clean)

$ CGO_ENABLED=0 GOWORK=off go mod tidy && cat go.mod
(toolchain go1.26.0 line survives, as F-1/D-1.1.a require)

$ CGO_ENABLED=0 GOWORK=off go test ./... -count=1 -v
--- 23 top-level test functions + 19 subtests = 42, all PASS, across 4 packages:
    github.com/panitw/folio/folio-go               ok  0.476s
    github.com/panitw/folio/folio-go/internal      ok  0.327s
    github.com/panitw/folio/folio-go/internal/geom ok  0.191s
    github.com/panitw/folio/folio-go/internal/pdf  ok  0.582s

$ shasum -a 256 fixtures/minimal-rect/expected.pdf
0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c  (unchanged; matches expected.json)
```

**Cross-target matrix not run** — per-epic cadence (D-000.4), due at Epic 1 close; Story 1.2 builds
the matrix. This run only exercised darwin/arm64, Go 1.26.0, `CGO_ENABLED=0`.

`git status --porcelain` at completion shows exactly the files this story adds/modifies: `LICENSE`
(new), `fixtures/` (new), `folio-go/` (new), this story file, the decision log (both already
untracked from an earlier run this session), `sprint-status.yaml` (status transitions) and
`ARCHITECTURE-SPINE.md` (D-000.6 amendment). No red-proof mutation was left in place — every one
was reverted and the full suite re-confirmed green afterward.

---

## Dev Agent Record

### Agent Model Used

Claude (bmad-story-developer agent), driving the `/bmad-dev-story` skill.

### Debug Log References

No debugger was used. All verification was via `go build`, `go vet`, `gofmt`, `go test -v`, direct
mutation-and-revert red-proof runs (see Delivery Log), and `qpdf --check` for independent PDF
structural validation.

### Completion Notes List

- Repository skeleton, `LICENSE`, and the exact directory set the story calls for were created;
  no placeholder packages for later stories were added.
- `folio-go/go.mod` pins `go 1.25.0` / `toolchain go1.26.0` per F-1/D-1.1.a, with the reasoning
  recorded as an in-file comment; `go mod tidy` was re-run after every mutation touching `go.mod`
  and the toolchain line's survival was reconfirmed each time.
- `internal/geom` (`Length`, `Rect`, `ScaleRound`) has an empty non-test import set and one
  round-half-to-even scaling function, table-tested including the story's four disagreement cases.
- `internal/pdf/numbers.go` implements exactly the three functions named in D-1.1.b
  (`appendLength`, `appendInt`, `appendIntPadded`), with `appendIntPadded` delegating to
  `appendInt`. `appendLength` avoids the `math.MinInt64` negation-overflow trap flagged in the
  story by dividing before negating, verified with MinInt64-adjacent and MinInt64-exact test cases.
  AD-3's spine text was amended (D-000.6) to match this two-representation shape.
  `internal/arch_test.go` provides the whole-of-`internal/` `float64`/`float32` AST guard the story
  asked to place either in `geom` or a dedicated file — placed as a dedicated file with its own
  package, since it needs to see both `geom` and `pdf`.
- `internal/pdf/document.go` serializes the fixed document exactly as specified in Dev Notes
  (byte-for-byte, verified against `qpdf --check` and direct byte inspection), with the `/ID`
  non-circularity argument restated as a code comment on `computeID`.
- `folio.Render()` is exported, doc-marked `PROVISIONAL — Story 1.1 only` (D-1.1.c), and imports
  only `internal/pdf` (which imports only `internal/geom`) — the story's allowed import chain.
- The two-process determinism harness (`folio-go/render_test.go`) genuinely re-executes the test
  binary via `exec.Command(os.Args[0], ...)` gated on `FOLIO_SUBPROCESS_RENDER=1`, with
  `GOMAXPROCS=1` / `GOMAXPROCS=4` children, and reports the first-divergence byte offset plus a hex
  window on mismatch.
- The golden fixture (`fixtures/minimal-rect/`) was recorded once from the first passing render and
  never regenerated afterward, including through every red-proof mutation/revert cycle.
- All ten red-proofs (RP-1 through RP-10) were actually applied, observed red, and reverted — see
  Delivery Log for every observed failure message.
- Per D-000.4/Task 9, the cross-target Docker/Node matrix was explicitly **not** run in this story.

### File List

**Finisher addendum:** the list below is corrected and extended from the pre-review version to
reflect what this commit actually contains, including the finisher's own new files (`.gitattributes`,
`internal/pdf/document_test.go`) and the amendment to `epics.md` the finisher made under D-000.6 (see
QA Results → Finding Resolutions, Finding 14).

New files:
- `LICENSE`
- `.gitattributes` (finisher: QA review Major 10)
- `folio-go/go.mod`
- `folio-go/version.go`
- `folio-go/folio.go`
- `folio-go/gomod_test.go`
- `folio-go/render_test.go`
- `folio-go/fixture_test.go`
- `folio-go/internal/arch_test.go`
- `folio-go/internal/geom/geom.go`
- `folio-go/internal/geom/scale.go`
- `folio-go/internal/geom/scale_test.go`
- `folio-go/internal/pdf/document.go`
- `folio-go/internal/pdf/document_test.go` (finisher: QA review Major 13, AC10 provenance tests)
- `folio-go/internal/pdf/numbers.go`
- `folio-go/internal/pdf/numbers_test.go`
- `folio-go/internal/pdf/emit_source_test.go`
- `folio-go/internal/pdf/testutil_test.go`
- `fixtures/minimal-rect/expected.json`
- `fixtures/minimal-rect/expected.pdf`
- `fixtures/minimal-rect/README.md`

Modified files:
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md`
  (D-000.6: AD-3 Rule paragraph amendment; finisher: heading amendment, QA review Major 12)
- `_bmad-output/planning-artifacts/epics.md` (finisher: `epics.md:147` and `epics.md:439` amended
  under D-000.6 to match the amended AD-3 — QA review Minor 14)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (status transitions, see below)
- `_bmad-output/implementation-artifacts/1-1-a-minimal-pdf-reproducible-on-one-machine.md` (this
  file: task checkboxes, Delivery Log, Dev Agent Record, Status, QA Results → Finding Resolutions)

---

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer (adversarial, no prior context on the implementation)
- **Date:** 2026-08-23
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 3
- **Majors:** 9
- **Minors:** 9
- **Nits:** 6

**What was independently verified (not taken from the Delivery Log).** Full sweep re-run from
`folio-go/`: `CGO_ENABLED=0 go build ./...` clean, `go vet ./...` clean, `gofmt -l .` empty (bare
`$(go env GOROOT)/bin/gofmt`, not a wrapper), `CGO_ENABLED=0 go test ./... -count=1 -v` **all
green — 16 top-level test functions + 13 subtests across 4 packages**. Golden hash independently
re-derived from a fresh out-of-band render: `0f925e1b…cb4f7c`, 547 bytes, byte-identical to both
`expected.json` and `expected.pdf`. `/ID` independently re-derived in Python: SHA-256 over the
449-byte pre-`/ID` body, first 16 bytes, upper hex = `C6086F83998B551A85031BCAB15DACC6`, matching
both array entries. `qpdf --check` (12.4.0) clean. xref verified by hand: five entries, **each
exactly 20 bytes**, every declared offset lands on `N 0 obj`, `startxref 309` lands on `xref`,
`/Length 28` equals the real stream length. The PDF was rasterised (`qlmanage`) and **genuinely
renders one black filled rectangle** — hash equality is not being satisfied by two identical broken
files. The two-process harness was instrumented and confirmed to fork **three distinct PIDs**
(parent 96724, child A 96725, child B 96726, each child self-reporting its own pid) — it is
genuinely two OS processes, not a same-process comparison. Red-proofs RP-1, RP-4, RP-9 and RP-10
were applied by hand and each reddened **with the exact symptom and, for RP-4 and RP-10, the exact
message text recorded in the Delivery Log** (RP-4's "byte offset 455", RP-10's
`…B15DACC6 != …B15DACC0`). The Delivery Log's red-proof transcripts are genuine, not reconstructed.
`go mod tidy` was run against the shipped `go.mod` and the `toolchain go1.26.0` pin survived.

**What is genuinely good.** Scope discipline is exemplary: no `.github/`, no `fonts/`, no `cmd/`,
no `wasm/`, no `folio-designer/`, none of the eight deferred `internal/` packages, no `go.work`, no
`go.sum`, no `vendor/`, no stubs, no premature abstraction beyond one item (Minor 21). `internal/geom`'s
non-test files have a genuinely empty import set (F-3, strict form). No `time`, `os`, `math/rand`,
`net` or `math` import in any non-test file under `internal/`; no goroutines; no map iteration; no
`go:embed`. `appendLength` is correct on every boundary probed, including `math.MinInt64`
(`-9223372036854775.808`), `MaxInt64`, and the leading-zero-fraction-plus-trim combination the
tests do not cover (`1010 → "1.01"`, `10 → "0.01"`); I could not construct a mutation to
`numbers.go` that `numbers_test.go` misses. D-1.1.b's three function names, signatures and the
`appendIntPadded → appendInt` delegation are implemented exactly as ruled, and the delegation is
real (verified by reading, and `appendIntPadded`'s overlapping `copy` is memmove-correct under
reallocation). `go.mod` carries both literal lines and a test asserts **both**, per D-1.1.a
(addendum).

**The shape of the problem.** The shipped artifact is correct. The bytes are right, the PDF is
valid and non-blank, the hash reproduces, the two processes agree. What is not right is the
**assertion array** — which is this story's actual deliverable. Three of the guards this story
exists to establish do not catch the hazards they name, and I demonstrated each by mutation with
the whole suite staying green. Blocker 2 in particular ships AD-3's own **Prevents** clause verbatim
into the output with every guard, the golden hash, `vet` and `gofmt` all silent.

---

### Finding 1: AC6 violated — two of the four MediaBox entries never pass through the emitter

- **Severity**: Blocker
- **Category**: AC Conformance
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/pdf/document.go:139`
- **Observation**: The MediaBox is written as
  `dst = append(dst, " 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 "...)` — the array's first
  two entries (`llx`, `lly`) are **literal bytes inside a string constant**. Only `urx`/`ury` go
  through `appendLength` (lines 140–142). Proven by mutation: I made `appendLength` return the
  marker `Z` for a zero-valued `geom.Length`, re-rendered, and the page object still read
  `/MediaBox [0 0 595.276 841.89]` — the two leading zeros did not move, because they are not
  produced by the emitter. There are no `mediaBoxX`/`mediaBoxY` constants.
- **Impact**: AC6 says, without qualification, "Every geometric number in the output — **MediaBox
  entries** and the rectangle's four operands — was produced by one unexported emitter… **No other
  code writes a geometric number into a content stream or object body**." MediaBox is
  `[llx lly urx ury]`; all four are user-space coordinates. Two of them are written by other code.
  The bytes are correct today only because `appendLength(0)` would also emit `0` — the invariant is
  held by coincidence, not by construction. This is precisely the "third route to an output byte"
  that AD-3 (and its amended Rule: "no number reaches an output byte by any other route") exists to
  forbid, and it is in the story that establishes AD-3. Vacuity guard 3 does not catch it: it
  asserts the *content stream* contains the trimmed forms, and says nothing about MediaBox
  provenance.
- **Suggested Resolution**: Declare `mediaBoxX`/`mediaBoxY` (or a `geom.Rect`, see Minor 21) as
  `geom.Length` constants of `0` and emit all four entries via `appendLength`, exactly as
  `urx`/`ury` already are. Then add the provenance assertion that would have caught it — see
  Blocker 2's resolution.
- **Related AC**: AC6 (also AD-3 as amended, and D-1.1.b's routing table)

### Finding 2: AC6/AC7's guards do not catch AD-3's own named hazard — a float64 conversion plus float formatting ships with the entire suite green and an identical golden hash

- **Severity**: Blocker
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/pdf/emit_source_test.go:18-29` and
  `/Users/panitw/Projects/folio/folio-go/internal/arch_test.go:26-35`; hazard demonstrated at
  `/Users/panitw/Projects/folio/folio-go/internal/pdf/document.go:140-142`
- **Observation**: I replaced the MediaBox emission with
  `dst = strconv.AppendFloat(dst, float64(mediaBoxWidth)/1000, 'f', -1, 64)` (and likewise for
  height) — a literal `float64` conversion followed by float formatting, inside `internal/`,
  reaching an output byte. Result: `go build` clean, `go vet` clean, `gofmt -l` empty,
  `TestNoFloat64UnderInternal` **PASS**, `TestNoOtherNumberFormattingUnderPDF` **PASS**,
  `TestRenderIsByteIdenticalAcrossTwoProcesses` **PASS**, `TestRenderMatchesGoldenFixture`
  **PASS with a byte-identical hash** (`0f925e1b…cb4f7c`), full suite **all four packages green**.
  Two independent holes combine: (a) `forbiddenNumberFormatting` lists `strconv.FormatFloat` but
  not `strconv.AppendFloat` (nor `ParseFloat`, nor the `strconv.*Float` family); (b) `isFloatIdent`
  type-asserts `expr.(*ast.Ident)`, so `float64(x)` — an `*ast.CallExpr` in expression position —
  is invisible to the AST walk. The golden hash stays silent because `595.276` and `841.89` happen
  to round-trip through float64 shortest-representation to the same text.
- **Impact**: AD-3's **Prevents** line names exactly this: "`strconv.FormatFloat` or
  `fmt.Sprintf("%g")` reaching an output stream … which reintroduces platform-visible float
  formatting." The story's entire deliverable is the assertion array that closes this class, and
  the class is open. RP-2 was supposed to establish this guard, but it only ever tested a single
  spelling of the hazard — it is a red-proof of one regex, not of the property. On a value whose
  shortest float representation differs from its exact decimal (or on a platform where it does),
  byte-identity is lost with every guard still green — which is the failure mode that is
  undetectable until it has already shipped.
- **Suggested Resolution**: Make both guards AST-based on resolved identifiers rather than on
  literal text and bare type positions. (a) In `arch_test.go`, flag **any** `*ast.Ident` named
  `float64`/`float32` anywhere in the file (which subsumes conversions, composites, nested structs,
  interfaces and type params in one change — see Major 7), plus `token.FLOAT` basic literals. (b) In
  `emit_source_test.go`, match the `strconv.{Format,Append,Parse}Float` family and `math.*`, resolve
  the import alias rather than matching `strconv\.` literally, and match float verbs with
  flags/width/precision (`%[-+ #0]*[0-9.]*[eEfFgG]`) rather than the two-character strings `%f`/`%g`.
  (c) Add the provenance assertion AC6 actually needs: an `internal/pdf` test that mutates the
  emitter and asserts every geometric byte in `Serialize()`'s output moves — that single test would
  have caught Blocker 1 and Blocker 2 together.
- **Related AC**: AC6, AC7 (and RP-2, RP-6, which are weaker than they appear)

### Finding 3: AC4 violated — `ScaleRound` returns the wrong sign for `math.MinInt64`

- **Severity**: Blocker
- **Category**: Correctness
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/geom/scale.go:22-46`
- **Observation**: Measured, by direct call:
  `ScaleRound(math.MinInt64, 1, 2)` returns **`+4611686018427387904`**; the exact half-to-even
  result is `-4611686018427387904`. The sign is inverted. Cause: line 25–27 does `if n < 0 { n = -n }`,
  and `-math.MinInt64 == math.MinInt64` in two's complement, so `n` stays negative; `q = n/d` is then
  already negative, and line 44–45's `if negative { q = -q }` flips it positive. Second case:
  `ScaleRound(1000, 1, math.MinInt64)` returns **`-1`** where the exact answer is `0` — `d = -den`
  is likewise a no-op, so `twiceR > d` compares `2000 > -9223372036854775808`, which is trivially
  true, and the quotient is incremented. Neither case is covered: every row in `scale_test.go` uses
  `num == 1` and `den ∈ {1, 2, 4}`, all positive; I also confirmed the negative-denominator
  normalisation is dead code as far as the tests are concerned by neutering `if d < 0 { d = -d }`
  and watching `go test ./internal/geom/` stay green.
- **Impact**: AC4 states the function "computes on `int64` throughout and rounds half-to-even
  (banker's rounding) **on the exact integer quotient**". For legal `int64` inputs it does neither.
  This is the exact negate-first trap the story's own Dev Notes flagged as "the most likely defect
  in the whole task" — `appendLength` was written to avoid it (divide before negating) and is
  explicitly tested at `MinInt64`; `ScaleRound`, twenty lines away, walks straight into it and has
  no boundary case at all. Blast radius today is zero — `grep` confirms **no non-test file calls
  `ScaleRound`** — which is why this is a Blocker on AC conformance rather than on shipped output.
  It becomes a live defect the moment Story 1.4/1.5 wires up font scaling, and it would then bake
  into a re-recorded golden.
- **Suggested Resolution**: Normalise sign without negating in the `int64` domain — mirror
  `appendLength`'s divide-then-negate approach, or promote to `math/bits`-based 128-bit
  intermediates (note F-3 defers that deliberately, so the former is in scope and the latter is
  not). Extend the table with `den < 0`, `num < 0`, both negative, and `MinInt64`/`MaxInt64` rows,
  mirroring `numbers_test.go:30-32`.
- **Related AC**: AC4

### Finding 4: AC5's validity test and AC8's vacuity guard 1 both accept a 118-byte stub that is not a PDF

- **Severity**: Major
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/render_test.go:36-51` (`assertWellFormedPDF`),
  used at `:133-134` and `:148-154`
- **Observation**: `assertWellFormedPDF` checks only four things: non-empty, prefix `%PDF-1.7`,
  suffix `%%EOF\n`, and `bytes.Count(b, "/Type /Page ") == 1`. I replaced the whole body of
  `pdf.Serialize` with a single 118-byte literal containing only those four tokens plus a
  `/ID [<A×32><A×32>]` — **no catalog, no page tree, no content stream, no rectangle, no xref
  table, no trailer, no `startxref`**. `TestRenderProducesAValidPDF`,
  `TestRenderIsByteIdenticalAcrossTwoProcesses`, `TestRenderHasNoCreationOrModDate` and
  `TestRenderIDEntriesAreIdentical` **all passed**. Only the golden hash noticed.
- **Impact**: `assertWellFormedPDF` is explicitly documented as vacuity guard 1 — the thing that
  stops AC8 passing on "two identical failures". It admits a document with no object structure at
  all, so two identical wrecks would satisfy it in both children and `bytes.Equal` would then
  compare them successfully. And `TestRenderProducesAValidPDF` is the only test named for AC5's
  validity; it does not check validity. `qpdf --check` is human-only, recorded in the Delivery Log
  (`render_test.go:145-147`), so nothing automated replaces it.
- **Suggested Resolution**: AC2 forbids a PDF-*reading dependency*, not structural assertions on a
  byte slice. Assert what the document describes about itself: `xref`/`trailer`/`startxref` present
  and in order; the integer after `startxref` indexes to the `x` of `xref`; a `/Type /Catalog`
  object exists; the xref entry for object *N* points at the byte offset where `N 0 obj` begins;
  `/Length` equals the real stream body length; the content stream carries `re` and `f` with the
  expected operands. All are integer/byte work, all toolchain-independent (see Major 5), and all
  are things I verified by hand today.
- **Related AC**: AC5, AC8 (vacuity guard 1)

### Finding 5: the golden hash is the module's only byte-observing assertion, and it is gated behind an exact-toolchain equality that fatals first

- **Severity**: Major
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/fixture_test.go:86-95` (the `runtime.Version()`
  gate) preceding `:101` (`Render()`) and `:106-117` (the hash comparison)
- **Observation**: The toolchain gate is a `t.Fatalf` and it runs **before** `Render()` is ever
  called. Separately I confirmed by mutation that the golden hash is the *only* assertion that
  observes output bytes at all: xref offsets `+1`, `/Length` `+7`, `computeID` returning a constant,
  and the 118-byte stub of Major 4 each reddened exactly one test — `TestRenderMatchesGoldenFixture`
  — with `internal`, `internal/geom` and `internal/pdf` all staying green.
- **Impact**: On any machine not running exactly `go1.26.0` — i.e. every contributor who is not on
  the recording machine, and every CI leg Story 1.2 is about to add — the fatal fires and the hash
  comparison never executes. At that moment **nothing in the module observes the output bytes**. The
  gate itself is correct and is exactly what AC11 mandates; the defect is that the project has a
  single point of verification for its entire product and that point is conditional. Compounding it,
  the failure message on that path instructs the reader that "the golden hash must be re-measured" —
  and re-measuring under a renderer that has silently regressed bakes the regression into the
  fixture, which is precisely the C6 / vacuity-guard-4 hazard.
- **Suggested Resolution**: Keep the AC11 gate exactly as it is, and add the toolchain-independent
  structural assertions from Major 4 in `internal/pdf`, so that a machine on a different toolchain
  still observes the document's self-consistency even when the hash is legitimately incomparable.
- **Related AC**: AC11 (mechanism correct), AC5/AC8/AC10 (all of which depend on it)

### Finding 6: `ScaleRound` overflows silently while its doc comment promises exactness

- **Severity**: Major
- **Category**: Correctness
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/geom/scale.go:22` (`int64(v) * num`),
  and `:39` (`r * 2`)
- **Observation**: Measured: `ScaleRound(math.MaxInt64/2, 4, 2)` returns **`-2`** (exact answer
  ≈ `MaxInt64`); `ScaleRound(1000000, math.MaxInt64/1000, 1)` returns **`-808000`**. No panic, no
  error, no guard, no documented domain. `r * 2` at line 39 has the same exposure for `d > 2^62`,
  where a negative `twiceR` inverts the rounding decision.
- **Impact**: The doc comment at lines 7–13 asserts "the exact numerator `v*num` and the exact
  remainder against `den` are integer values, so 'exact integer quotient' above is literal". That is
  a true statement about floats and a false statement about exactness. The failure is deterministic
  (Go's wraparound is defined), so it survives the two-process test and would reproduce identically
  into a golden — a wrong number that is perfectly reproducible is the worst outcome for this
  project, because the hash gate will bless it.
- **Suggested Resolution**: Either document the valid domain and panic outside it (consistent with
  the existing `den == 0` panic and with the story's "panics only where a programmer error is
  unambiguous"), or check the multiplication for overflow before performing it. Add table rows at
  the boundary either way.
- **Related AC**: AC4

### Finding 7: AC7's AST guard misses most float-bearing declaration shapes

- **Severity**: Major
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/arch_test.go:26-35` (`isFloatIdent`),
  reached only from `:44`, `:61`, `:76`
- **Observation**: `isFloatIdent` type-asserts `expr.(*ast.Ident)` and returns false for every other
  node, and the three `ast.Inspect` cases only ever hand it `f.Type` / `decl.Type` directly. I
  compiled a probe for each shape into `internal/geom` and ran `TestNoFloat64UnderInternal`.
  **Caught (4):** `func f(x float64)` (RP-6's shape), a method parameter, a named result, `type A = float64`.
  **Missed (9+):** `[]float64`, `*float64`, `map[string]float64`, `...float64`, `[3]float64` as a
  struct field, a nested anonymous struct field, an interface method signature, `var Dpi = 72.0`
  (untyped), `const Inch = 25.4` (untyped) — plus `float64(x)` conversions and `func(float64) float64`
  field types.
- **Impact**: The function's own docstring claims it finds float "appearing **anywhere** in a type
  expression (func signatures, struct fields, var/const declarations, type declarations, type
  parameters)". In practice it covers only the un-nested single-identifier form of each. AC7 says
  "No function, method, field, constant, variable or type parameter under `folio-go/internal/` has
  `float64` (or `float32`) in its signature or declaration" — `func f(xs []float64)` unambiguously
  does, and passes. RP-6 tested the one shape the guard handles. This is the same root cause as
  Blocker 2(b) and one change fixes both.
- **Suggested Resolution**: Replace the type-position walk with a flag on **any** `*ast.Ident` whose
  `Name` is `float64`/`float32` anywhere in the file, plus `token.FLOAT` basic literals for the
  untyped-constant case. Keep the existing per-declaration reporting for the diagnostic message.
- **Related AC**: AC7 (and RP-6)

### Finding 8: D-1.1.b's routing table is violated — `/Size`, `/Count`, object references and the xref free entry are hard-coded numeric text

- **Severity**: Major
- **Category**: Convention
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/pdf/document.go:57`, `:60`, `:76`,
  `:139`, `:143`, `:166`, `:167`
- **Observation**: D-1.1.b fixes a routing table "that must not be re-derived per story", assigning
  `/Length`, `/Size`, `/Count`, `/N`, "object and generation numbers" to `appendInt`, and "xref
  offsets (10-digit), xref generation (5-digit)" to `appendIntPadded`. In the implementation,
  `/Count 1`, `/Size 5`, `/Pages 2 0 R`, `/Kids [3 0 R]`, `/Parent 2 0 R`, `/Contents 4 0 R`, the
  xref subsection header `0 5`, and the entire free entry `0000000000 65535 f ` are literal text
  inside string constants. Only the leading object number of each object, `/Length`, `startxref`
  and the four in-use xref entries actually route through the helpers.
- **Impact**: D-1.1.b's Consequences explicitly claim "Story 1.1 still exercises all three
  functions: `appendLength` for `MediaBox` and the rectangle, `appendInt` for `/Length` `/Count`
  `/Size` and object numbers, `appendIntPadded` for the xref" — and the Delivery Log repeats it. The
  `/Count` and `/Size` half is not true. This is the precedent-setting story for a table eight later
  stories will follow; a page-tree `/Count` is explicitly listed as arriving at Story 2.6, and it
  will be copied from here. Note this is a *convention* finding, not an output defect: unlike
  Blocker 1, none of these are geometric numbers, so AC6 is not violated.
- **Suggested Resolution**: Route the structural integers through `appendInt`/`appendIntPadded` as
  the ruling specifies, or — if the finisher judges hard-coded object graph literals acceptable for
  a fixed four-object document — record that carve-out as a `D-1.1.x` decision-log entry rather than
  leaving the ruling silently contradicted, and correct the Delivery Log's claim.
- **Related AC**: AC6 (adjacent); D-1.1.b

### Finding 9: package-level mutable state, which the story explicitly forbids

- **Severity**: Major
- **Category**: Convention
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/pdf/document.go:34`
- **Observation**: `var binaryCommentBytes = [4]byte{0xE2, 0xE3, 0xCF, 0xD3}`. A repo-wide scan of
  non-test files finds exactly one package-level `var` — this one. Every other fixed value in the
  file is a `const`.
- **Impact**: Dev Notes → "Things this story must not do" states: "no package-level mutable state"
  (NFR1.d, AD-1's rule binding now even though the lint is Story 1.3). Any later file in
  `package pdf` can reassign this array, silently changing **every** rendered byte from the header
  onward — in the one package where that matters most. It also sets the precedent the ban exists to
  prevent, in the story that states the ban.
- **Suggested Resolution**: `const binaryComment = "\xE2\xE3\xCF\xD3"` and append it directly. This
  is byte-identical output and removes the hazard entirely.
- **Related AC**: Dev Notes scope fence / NFR1.d / AD-1

### Finding 10: the golden `expected.pdf` is stored as a git *text* file with no `.gitattributes` — a `core.autocrlf` checkout corrupts it

- **Severity**: Major
- **Category**: Correctness
- **Location**: `/Users/panitw/Projects/folio/fixtures/minimal-rect/expected.pdf`; missing
  `/Users/panitw/Projects/folio/.gitattributes`
- **Observation**: The file contains **no NUL byte** (verified: the only bytes ≥ 0x80 are
  `0xE2 0xE3 0xCF 0xD3`), so git's binary heuristic classifies it as text — confirmed with
  `git diff --numstat`, which reports `30 0` (line counts) rather than `- -`. There is no
  `.gitattributes` anywhere in the repo. The file contains 30 LF bytes.
- **Impact**: On a Windows checkout with `core.autocrlf=true` (the Git-for-Windows default), the
  547-byte normative golden becomes 577 bytes and its SHA-256 changes — and a contributor who
  re-records from that checkout commits a CRLF fixture that breaks everyone else. The same applies
  to every `.go` file (`gofmt -l` would then flag the whole module) and to `expected.json`. For a
  project whose sole value proposition is byte-identity, and with Story 1.2's cross-target matrix
  next in the queue, this is a live hazard rather than a theoretical one.
- **Suggested Resolution**: Add `.gitattributes` at repo root: `fixtures/**/*.pdf binary`,
  `* text=auto eol=lf` (or at minimum `*.go text eol=lf` and `*.json text eol=lf`). Cheap, and it
  belongs with the fixture that needs it rather than with Story 1.2.
- **Related AC**: AC11 (the fixture's integrity), AD-21

### Finding 11: the `folioGoVersion == folio.Version` hard-fail manufactures exactly the golden-fixture edit that vacuity guard 4 forbids

- **Severity**: Major
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/fixture_test.go:97-99`
- **Observation**: Measured. I changed `folio.Version` from `"0.0.0-dev"` to `"0.1.0"` and ran the
  suite: `TestRenderMatchesGoldenFixture` failed with
  `fixture folioGoVersion "0.0.0-dev" does not match folio.Version "0.1.0"`, while
  `TestRenderProducesAValidPDF` and `TestRenderIsByteIdenticalAcrossTwoProcesses` stayed green —
  i.e. **the rendered bytes were provably identical**.
- **Impact**: AC11 requires the fixture to *record* the `folio-go` version; it does not require the
  recorded value to equal the current one. As written, the first legitimate version bump (Epic 4's
  `folio-go/v0.1.0` tag, per `version.go`'s own doc comment) reddens the golden test with unchanged
  output, and the only way to clear it is to hand-edit `expected.json`. That is the precise reflex
  the fixture README and vacuity guard 4 exist to suppress — "Do not regenerate this fixture to make
  a failing test pass" — and it trains it on a false positive, which is the most effective way to
  erode the habit before it is needed for a true positive.
- **Suggested Resolution**: Assert the field is present and well-formed (already done at `:73-75`),
  not that it equals `folio.Version`. If a coupling is wanted, make it a `t.Logf` or a separate
  non-fatal check, and keep the fatal path reserved for the hash and the toolchain.
- **Related AC**: AC11, vacuity guard 4, C6

### Finding 12: AD-3's section heading now contradicts the rule beneath it

- **Severity**: Major
- **Category**: Convention
- **Location**: `/Users/panitw/Projects/folio/_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md:119`
- **Observation**: The heading still reads **"### AD-3 — Numbers reach the PDF through exactly one
  function"**. The Rule beneath it, replaced under D-000.6 by this story, now specifies "exactly
  **two** representations" implemented (per D-1.1.b) across **three** functions. Confirmed against
  the git diff: only the `- **Rule:**` paragraph changed; `Binds` and `Prevents` are correctly
  untouched, per D-000.6 consequence (4).
- **Impact**: D-000.6 exists precisely so that "the documents would [not] keep stating things that
  are false"; its rationale notes that otherwise "every future agent then re-derives the fix
  independently and the next one may not flag it". A canonical invariant whose title contradicts its
  own rule is the exact failure mode. Story 1.3's lint author and the eight later stories D-1.1.b
  anticipates will read the heading first.
- **Suggested Resolution**: Amend the heading in the same commit — e.g. "AD-3 — Numbers reach the
  PDF through exactly one file, in exactly two representations". Log the heading change alongside
  the Rule change under D-000.6, since it is the same amendment.
- **Related AC**: D-000.6, D-1.1.b

### Finding 13: AC10's content-derivation clause has no assertion — only the golden hash notices, and the golden legitimately moves

- **Severity**: Major
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/render_test.go:177-207`
  (`TestRenderIDEntriesAreIdentical`); mechanism at
  `/Users/panitw/Projects/folio/folio-go/internal/pdf/document.go:99-102`
- **Observation**: Two probes. (a) `computeID` returning a **constant** 16 zero bytes: every
  behavioural test passed — only the golden hash reddened. (b) `computeID` digesting a
  *different range* (`append(soFar, 'X')`): same result. `TestRenderIDEntriesAreIdentical` cannot
  distinguish a content-derived digest from a constant, because both entries are written from a
  single `idHex` variable (`document.go:80-86`) — the equality it asserts is true by construction.
  RP-10 probes only the identity half.
- **Impact**: AC10 has two clauses — "`/ID` is the first 16 bytes of a SHA-256 taken over the
  serialized body up to the point `/ID` is written" **and** "both array entries identical". The
  second is pinned; the first is pinned by nothing but a whole-document fingerprint that will be
  deliberately re-recorded the moment Story 1.4 gives `Render` a template. At that re-recording, a
  wrong derivation is silently blessed. I did independently confirm the derivation is correct
  *today* (Python re-derivation over the 449-byte prefix matches both entries exactly), and the
  non-circularity reasoning holds: the trailer follows the xref, `startxref`'s value is already
  fixed before `computeID` runs, and `/ID`'s 32-hex width is constant, so writing it cannot shift
  any recorded offset. The reasoning is sound and the code implements it — it is just not asserted.
- **Suggested Resolution**: An `internal/pdf` test that calls `computeID` directly on two different
  inputs and asserts the results differ, plus a test that re-derives the digest from `Serialize()`'s
  own output prefix and compares it to the emitted `/ID` — the same check I ran by hand, in Go.
- **Related AC**: AC10 (derivation half), RP-10

### Finding 14: `epics.md` and the decision log still restate AD-3's superseded "one emitter" rule

- **Severity**: Minor
- **Category**: Convention
- **Location**: `/Users/panitw/Projects/folio/_bmad-output/planning-artifacts/epics.md:147` and
  `:439`; `/Users/panitw/Projects/folio/_bmad-output/implementation-artifacts/folio-mvp-decision-log.md:43`
- **Observation**: Sibling instances of the claim Finding 12 flags. `epics.md:147` states "One
  unexported emitter converts numbers to PDF decimal text; no other code writes a number to output
  (AD-3)" — false under the amended AD-3. `epics.md:439` (Story 1.1's own source AC) says "**every**
  number in the output was written by the single unexported emitter"; the story file narrowed this
  to "every **geometric** number" (AC6) on D-1.1.b's authority, so the epic text is now stale
  against the ruling that governs it. `folio-mvp-decision-log.md:43` carries the same "one number
  emitter" shorthand. (`epics.md:818`'s "exactly one function" is about AD-24's Y-flip and is
  unaffected.)
- **Impact**: D-000.6's stated purpose is one canonical statement per rule. Three documents now
  state the superseded form. `epics.md` is the source every later story's ACs are carried from,
  so the stale line will propagate.
- **Suggested Resolution**: The finisher should decide whether `epics.md` falls inside D-000.6's
  amendment scope ("`ARCHITECTURE-SPINE.md`, `folio-format.md`, and the other files in the SPEC's
  `companions:` list" — `epics.md` may not be listed). If in scope, amend with this commit; if not,
  raise it as a `D-1.1.x` note so the next story does not re-derive the conflict.
- **Related AC**: D-000.6, D-1.1.b

### Finding 15: `numbers.go`'s header comment states an enforcement and an exemption that both do not exist

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/pdf/numbers.go:16-19`
- **Observation**: The comment claims this is "the only file under `internal/pdf` **allowed** to
  reference `strconv.Format*`, `strconv.Itoa`, `strconv.Append*`, or a `fmt` formatting call
  (enforced by `TestNoOtherNumberFormattingUnderPDF` in `emit_source_test.go`)". Both halves are
  false. (a) That test grants no exemption to any file — RP-2's own transcript shows it firing on
  `numbers.go` itself. (b) It does not check `strconv.Itoa` or `strconv.Append*` at all: I placed
  each in a probe file inside `internal/pdf` and the test passed. Per D-1.1.b, the file-scoped rule
  is **Story 1.3's** lint, not a Story 1.1 test.
- **Impact**: A future author reading this comment will believe `numbers.go` has a carve-out it does
  not have, and will believe `strconv.Itoa` elsewhere in the package is already blocked when it is
  not. The comment attributes Story 1.3's job to a Story 1.1 test.
- **Suggested Resolution**: Reword to say what is true now — that this file is where numeric
  formatting belongs by convention, and that the file-scoped ban arrives with Story 1.3's lint per
  D-1.1.b — and note that the Story 1.1 test is a blanket ban with no exemptions.
- **Related AC**: AC6; D-1.1.b

### Finding 16: the AC7 vacuity guard counts the test's own directory, so it does not require the packages it names

- **Severity**: Minor
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/arch_test.go:167-172`, populated at
  `:148-158`
- **Observation**: The walk parses `arch_test.go` itself, so `pkgDir == "."` is unconditionally
  inserted into `packagesVisited` and the test file's own declarations are unconditionally added to
  `declCount`. Measured: with `internal/pdf` moved aside the guard passed
  (`{".", "geom"}`); with `internal/geom` moved aside it passed (`{".", "pdf"}`); only with **both**
  gone did it fire. So `len(packagesVisited) < 2` requires exactly one real package, and
  `declCount == 0` is unreachable while `arch_test.go` exists.
- **Impact**: The error message says "expected >= 2 (geom and pdf)" but the assertion never checks
  those names, and the Delivery Log repeats the stronger claim ("it asserts it visited ≥ 2 package
  directories (`geom`, `pdf`)"). Deleting the entire package that writes output bytes leaves the
  guard green. This is a weaker form of vacuity guard 2 than the story specifies.
- **Suggested Resolution**: Assert `packagesVisited` contains `"geom"` and `"pdf"` by name, and
  count declarations only from files where `pkgDir != "."`.
- **Related AC**: AC7, vacuity guard 2

### Finding 17: an xref offset of ten or more digits silently produces a 21-byte entry

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/pdf/numbers.go:123-125`; caller at
  `/Users/panitw/Projects/folio/folio-go/internal/pdf/document.go:169`
- **Observation**: `appendIntPadded` returns early when `digits >= width`, so it widens rather than
  truncating — correct for a general helper, wrong for a fixed-width field. Measured: an entry built
  for offset `12345678901` is `"12345678901 00000 n \n"` — **21 bytes**. There is no guard at the
  xref call site. (The shipped document's entries are all exactly 20 bytes; I counted them.)
- **Impact**: A classic xref table is byte-addressed; a 21-byte entry desynchronises every
  subsequent entry and produces a file some readers accept and others reject — the exact failure
  mode the story's Dev Notes warn about. It requires an output ≥ 10 GB, so it is not reachable in
  MVP, but it is a silent-corruption path in the structure the whole format depends on.
- **Suggested Resolution**: Panic in `appendXref` (or in `appendIntPadded` behind an explicit
  fixed-width variant) when the value exceeds the field width — consistent with `ScaleRound`'s
  "unambiguous programmer error" panic precedent.
- **Related AC**: AC5

### Finding 18: no test asserts the PDF's internal self-consistency

- **Severity**: Minor
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/pdf/` (no test file calls
  `Serialize`, `appendXref`, `appendPageObject`, `appendContentStreamObject`, `buildContentStream`
  or `computeID`)
- **Observation**: `internal/pdf` has two test files; `numbers_test.go` covers only the three number
  helpers and `emit_source_test.go` executes no production code. Nothing asserts that the xref entry
  for object *N* points at `N 0 obj`, that `/Length` equals the stream body length, or that
  `startxref` points at `xref`. I verified all three by hand today and they are correct.
- **Impact**: Had any been wrong on day one, the golden fixture would have enshrined the wrong bytes
  and the whole suite would still be green — vacuity guard 4 makes the fixture an input thereafter,
  so the error would have become permanent. This is the same gap as Major 4/5 viewed from the
  package side, and one set of assertions closes all three.
- **Suggested Resolution**: See Major 4.
- **Related AC**: AC5

### Finding 19: `expected.pdf` is never read by any test and can silently drift from the normative hash

- **Severity**: Minor
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/fixtures/minimal-rect/expected.pdf`;
  `/Users/panitw/Projects/folio/folio-go/fixture_test.go:57-117`
- **Observation**: `grep` confirms no test references `expected.pdf`. Its SHA-256 currently matches
  `expected.json` exactly (I verified: `0f925e1b…cb4f7c`).
- **Impact**: The README directs humans to use `expected.pdf` for diffing while declaring
  `expected.json` normative. Nothing keeps them in agreement, so the file a human diffs against can
  become the one artifact in the fixture that is wrong.
- **Suggested Resolution**: One line in `TestRenderMatchesGoldenFixture`: read `expected.pdf`, assert
  its SHA-256 equals `fixture.SHA256`. As a bonus this lets the mismatch message diff against real
  bytes.
- **Related AC**: AC11, AD-21

### Finding 20: no assertion on the "no compression / no metadata" half of the format contract

- **Severity**: Minor
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/render_test.go:158-172`
- **Observation**: `TestRenderHasNoCreationOrModDate` covers `/CreationDate`, `/ModDate` and
  `/Info`. Nothing asserts `/Filter` is absent, though Dev Notes, the package doc
  (`document.go:1-6`) and the fixture README all promise uncompressed output; nor `/Producer`,
  `/Creator` or `/Metadata`.
- **Impact**: `/Producer` and `/Metadata` (XMP) are the next places a timestamp, a machine name or a
  toolchain string leaks into the bytes — the same AD-7 hazard AC9 exists to close, one key over.
  `/Filter` is the R4 compression risk the story explicitly removes from the variable set.
- **Suggested Resolution**: Extend the existing test's substring list. Three lines.
- **Related AC**: AC9, AD-7

### Finding 21: `geom.Rect` is dead code, and the story's stated reason for adding it is not met

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/geom/geom.go:22-24`
- **Observation**: `grep -rn "geom.Rect\|Rect{" --include='*.go'` returns nothing outside the
  declaration — zero references anywhere, tests included. Task 3 justifies it as "needed by the
  rectangle and by AD-2's 'one owner'", but `document.go:25-28` declares the rectangle as four loose
  `geom.Length` constants and never constructs a `Rect`. (`ScaleRound` is likewise uncalled, but
  AC4 mandates it, so that is expected.)
- **Impact**: The story's own scope fence bans premature abstraction and empty scaffolding on the
  grounds that it "would make Story 1.3's lint vacuous". `Rect` is the one item that slipped
  through — a type added for a use it does not serve. Everything else in the scope fence is
  honoured exactly.
- **Suggested Resolution**: Either use it (`rect geom.Rect = geom.Rect{X: 72500, …}` in
  `document.go`, which would also give Blocker 1's fix a natural home) or drop it until a caller
  exists. Using it is the better option — it makes Task 3's stated rationale true.
- **Related AC**: AC3; Dev Notes scope fence

### Finding 22: the plain-terms opener describes the plan, not the shipped result, and restates a superseded rule

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: this file, `## In plain terms` (lines 28–60)
- **Observation**: Assessed as requested. **Format is good**: no file paths, no function names, no
  code, plain language throughout. **The scope fence is honest and accurate** — "renders one
  hard-coded rectangle and nothing else… proves nothing about other machines — only that one machine
  repeats itself" is exactly what shipped, and "one thing will look wrong later and is not" is a fair
  warning. **But the tense is planning tense**: "Today the repository holds only planning documents
  — there is no program at all" is now false; "This story writes the first one"; "Done looks like…".
  And one substantive claim is wrong: "every number that lands in the finished file is written out
  by one single piece of code" — D-1.1.b overturned that (two representations, three functions), and
  Blocker 1 shows some numbers are not written by any of them.
- **Impact**: The opener is the section a non-technical reader trusts. It currently promises an
  invariant the project deliberately replaced.
- **Suggested Resolution**: The finisher should rewrite it in past tense describing what was built,
  and replace the "one single piece of code" sentence with the two-representations shape from
  D-1.1.b — e.g. that measurements and structural counts are written by two different, deliberately
  separated pieces of code that both work in whole numbers. Keep the scope fence verbatim; it is the
  best part of the section.
- **Related AC**: n/a (story hygiene)

### Finding 23: the Delivery Log's test count and one red-proof detail do not match a real run

- **Severity**: Nit
- **Category**: Tests
- **Location**: this file, `## Delivery Log` → "Full verification sweep (Task 9)"
- **Observation**: The log records "30 individual test functions/subtests". A real
  `go test ./... -count=1 -v` yields **16 top-level test functions + 13 subtests = 29**. Everything
  else in the sweep reproduces exactly (build, vet, gofmt clean; four packages green). Separately,
  RP-1's recorded transcript is accurate only for the story's *literal* mutation ("always emit three
  fractional digits", which I reproduced and which yields exactly the 8 rows logged); the weaker
  "stop trimming" reading yields 4 rows and leaves `50000 → "50"` green.
- **Impact**: Trivial, but this project's whole culture is that reported numbers are measured. Worth
  correcting so the log stays trustworthy.
- **Suggested Resolution**: Correct to 29.
- **Related AC**: Task 9

### Finding 24: `-test.run=^TestMain$` matches no test function

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `/Users/panitw/Projects/folio/folio-go/render_test.go:59`
- **Observation**: `TestMain` is not a test function, so the regex selects nothing; the
  `FOLIO_SUBPROCESS_RENDER` short-circuit at `:20-28` is what does the work. Harmless today.
- **Impact**: If the env short-circuit were ever refactored, the child would exit 0 having written
  nothing and the failure would surface as `assertWellFormedPDF`'s "output is empty" rather than the
  real cause.
- **Suggested Resolution**: Use `-test.run=^$` (explicitly select nothing) and add a comment saying
  the env var is what routes the child.
- **Related AC**: AC8

### Finding 25: `os.Stdout.Write` error and short-write are discarded in the subprocess path

- **Severity**: Nit
- **Category**: Correctness
- **Location**: `/Users/panitw/Projects/folio/folio-go/render_test.go:26`
- **Observation**: `os.Stdout.Write(b)` — return values ignored, then `os.Exit(0)`.
- **Impact**: This is the one write in the harness whose truncation would be indistinguishable from
  a determinism failure: a short write produces a mismatch that the diagnostic would report as a
  byte-offset divergence, sending the reader after the renderer instead of the pipe.
- **Suggested Resolution**: Check `n, err`; exit non-zero with a distinct message on short write.
- **Related AC**: AC8

### Finding 26: the page-count check depends on an incidental trailing space

- **Severity**: Nit
- **Category**: Tests
- **Location**: `/Users/panitw/Projects/folio/folio-go/render_test.go:47`
- **Observation**: `bytes.Count(b, []byte("/Type /Page "))` — the trailing space is the only thing
  distinguishing it from `/Type /Pages`, and the scan covers the whole file including content
  streams. Correct today (I verified the count is 1).
- **Impact**: A dictionary written as `/Type/Page` or `/Type /Page>>` would read as zero pages.
- **Suggested Resolution**: Match `/Type /Page` followed by a PDF delimiter, or read `/Count` from
  the Pages node.
- **Related AC**: AC5, AC8

### Finding 27: `appendIntPadded`'s doc comment misdescribes its own padding for negative values

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `/Users/panitw/Projects/folio/folio-go/internal/pdf/numbers.go:108-111`
- **Observation**: The comment says "left zero-filled to at least `width` characters". It actually
  pads the **digit run**, not sign+digits: `appendIntPadded(nil, -1, 5)` returns `"-00001"` — six
  characters. The test at `numbers_test.go:92` pins this behaviour and its inline comment states it
  correctly; only the function's own doc is wrong. (The Delivery Log records this choice as a
  deliberate mechanical detail, so the behaviour is intended.)
- **Impact**: Doc/behaviour drift in the file that is the module's numeric contract.
- **Suggested Resolution**: Say "left zero-filled so that at least `width` **digits** are emitted,
  excluding any sign".
- **Related AC**: D-1.1.b

### Finding 28: D-000.6 asks for before/after; only the "after" was recorded

- **Severity**: Nit
- **Category**: Convention
- **Location**: this file, `## Delivery Log` → "Spine amendment (D-000.6)";
  `/Users/panitw/Projects/folio/_bmad-output/implementation-artifacts/folio-mvp-decision-log.md:494-497`
- **Observation**: D-000.6 consequence (3) requires each amendment to "be quoted verbatim in this
  log so the **before/after** is readable without `git log`". D-1.1.b delegates the quote to the
  Story 1.1 Delivery Log, which quotes the **resulting** AD-3 text in full but not the superseded
  paragraph.
- **Impact**: Minor governance drift on the very decision that exists to prevent governance drift.
  The `before` is one paragraph and is available in the diff I read.
- **Suggested Resolution**: Add the superseded Rule paragraph above the new one in the Delivery Log's
  "Spine amendment" section.
- **Related AC**: D-000.6

---

### Notes for the finisher

- **Mutation hygiene**: every mutation in this review was applied to a hand-taken snapshot copy and
  reverted from it (never `git checkout`, since these files are untracked). After the final revert,
  `diff -r` against the snapshot reports the `folio-go/` and `fixtures/` trees **byte-identical**,
  and `go test ./... -count=1` is green across all four packages. No red-proof residue was left.
- **Cheapest high-value fix**: one AST-based rewrite of `isFloatIdent` (flag any `*ast.Ident` named
  `float64`/`float32`) closes Blocker 2(b) and Major 7 together. One provenance test on
  `Serialize()`'s geometric bytes closes Blocker 1 and Blocker 2(c). One set of structural
  assertions in `internal/pdf` closes Major 4, Major 5 and Minor 18.
- **What I could not fault**: `numbers.go`'s three functions, the `/ID` non-circularity reasoning and
  its implementation, the xref byte arithmetic, the two-process harness's process separation, the
  toolchain pin and its rationale, the scope fence, and the honesty of the Delivery Log's red-proof
  transcripts. Those are solid.

---

## Finding Resolutions (finisher)

Every finding above was triaged FIX / DISMISS / DEFER. All three Blockers were fixed (per the
orchestrator's ruling, they were not dismissible, and the orchestrator had independently reproduced
Blockers 1 and 2). 25 of 28 findings were FIXed, 3 were DISMISSed with rationale below, none were
deferred as a follow-up ticket — everything raised was judged in-scope for this story and cheap
enough to close now rather than carry forward.

### Blockers

**Finding 1 (Blocker) — MediaBox's two leading-zero entries bypassed the emitter.** **FIX.** All
four MediaBox entries now route through `appendLength` via a `geom.Rect` (`internal/pdf/document.go`,
`appendMediaBox`). Golden hash re-verified unchanged before and after
(`0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c`).

**Finding 2 (Blocker) — AC6/AC7 guards missed AD-3's own named hazard (`strconv.AppendFloat` +
`float64` conversion).** **FIX.** `internal/arch_test.go`'s float guard now flags any `*ast.Ident`
named `float64`/`float32` anywhere in the AST (subsuming every declaration shape) plus any untyped
`token.FLOAT` literal, rather than type-asserting only at specific declaration positions.
`internal/pdf/emit_source_test.go` was rewritten as an AST-based, resolved-import-alias check
(`TestNumberFormattingIsConfinedToNumbersGo`) implementing D-1.1.b's file-scoped guardrail brought
forward from Story 1.3 into this story, per the orchestrator's explicit instruction: within
`internal/pdf`, no file but `numbers.go` may reference `strconv.Format*`/`Itoa`/`Append*` or any
`fmt` formatting call — matched by resolved selector, not by an enumerated regex list, so
`strconv.AppendFloat` is caught as *any* `strconv.Append*` call rather than needing its own list
entry. **Re-mutation proof, run for real as instructed:** reapplied the exact
`strconv.AppendFloat(dst, float64(mediaBoxWidth)/1000, 'f', -1, 64)` mutation to
`appendMediaBox`. Result: both `TestNoFloat64UnderInternal` and
`TestNumberFormattingIsConfinedToNumbersGo` failed, independently, each naming the file and line
(`document.go:178`, `:180`, `:182`, `:184`) and the exact hazard (`identifier float64` /
`strconv.AppendFloat`) — the suite is genuinely RED, not merely "expected to be." Reverted via a
pre-mutation backup; `diff` confirmed byte-identical restoration; full suite re-confirmed green.
Additionally confirmed by throwaway probe files (compiled, tested, then deleted) that the rewritten
AST guard catches a `float64` struct field, a `float64` named return, and a `type X = float64`
alias — the three shapes the review specifically doubted — each reddening
`TestNoFloat64UnderInternal` with the correct file/line.

**Finding 3 (Blocker) — `ScaleRound` returns the wrong sign at `math.MinInt64`.** **FIX.**
`internal/geom/scale.go` rewritten to never negate a value that could be `math.MinInt64` (mirroring
`appendLength`'s divide-before-negate technique, but going further: it never negates the dividend or
quotient at all, only the always-safely-bounded remainder). `ScaleRound(math.MinInt64, 1, 2)` now
returns `-4611686018427387904`, the exact half-to-even answer. Boundary rows added to
`scale_test.go`: `MinInt64`, `MaxInt64`, `-1`, `999`, `1000`, `-1000` (all with `num=1, den=2`, as
requested), all passing.

### Majors

**Finding 4 — `assertWellFormedPDF` (and AC5's validity test) accepted a 118-byte stub.** **FIX.**
`render_test.go`'s `assertWellFormedPDF` now re-derives and checks: `/Type /Catalog` presence,
`startxref` genuinely pointing at the `xref` keyword, every xref in-use entry's offset genuinely
landing on its own `N 0 obj`, and every `/Length` declaration matching its stream body's actual
byte count — all toolchain-independent, no PDF-reading dependency (AC2 preserved).

**Finding 5 — golden hash is the only byte-observing assertion, gated behind a toolchain fatal
that runs before `Render()`.** **FIX.** `fixture_test.go`'s `TestRenderMatchesGoldenFixture`
reordered: presence checks, then `Render()` + `assertWellFormedPDF` (unconditional), then the
`expected.pdf`-drift check, then the toolchain gate, then the hash comparison. A machine on a
different toolchain now observes real structural self-consistency instead of nothing.

**Finding 6 — `ScaleRound` overflows silently on large inputs.** **FIX**, folded into Finding 3's
fix. `v*num` overflow, `den == math.MinInt64`, and the `den == -1 && v*num == math.MinInt64`
division-overflow case are all now explicit panics (consistent with the existing `den == 0` panic
precedent) rather than silent wraparound; `2|r|` is never computed (avoiding that overflow site
entirely) via the algebraically equivalent `|r| > |den|-|r|` comparison. Both of the review's
measured examples now panic instead of returning a wrong answer; new tests pin this.

**Finding 7 — AC7's AST guard misses most float-bearing declaration shapes.** **FIX**, same change
as Finding 2's `arch_test.go` fix (both were the identical root cause). Empirically confirmed to
catch all nine previously-missed shapes the review measured, by construction (the new walk flags
the identifier's occurrence anywhere in the AST, not its position) and spot-checked with three
throwaway probes (struct field, named return, type alias — see Finding 2).

**Finding 8 — D-1.1.b's routing table violated (`/Size`, `/Count`, object references, xref free
entry hard-coded).** **FIX.** `internal/pdf/document.go` rewritten: every object/generation number
inside every reference (`appendRef`), every object header's own number (`appendObjHeader`), `/Size`
and `/Count`, and the xref free entry's offset/generation fields all now route through
`appendInt`/`appendIntPadded`. Golden hash re-verified unchanged.

**Finding 9 — package-level mutable state (`var binaryCommentBytes`).** **FIX.** Changed to
`const binaryComment = "\xE2\xE3\xCF\xD3"` (a string, not a `[4]byte` var) — byte-identical output,
no package-level `var` remains in `internal/pdf`.

**Finding 10 — `expected.pdf` has no `.gitattributes`, git classifies it as text.** **FIX.** Added
`/.gitattributes` at repo root: `* text=auto eol=lf`, `fixtures/**/*.pdf binary`,
`fixtures/**/*.json text eol=lf`, `*.go text eol=lf`.

**Finding 11 — `folioGoVersion == folio.Version` hard-fail manufactures a false-positive golden
break.** **FIX.** Removed the equality check; the fixture's version field is now only checked for
presence (which AC11 actually requires), not equality with the current module version.

**Finding 12 — AD-3's heading contradicts the amended Rule.** **FIX.** Heading changed to
`### AD-3 — Numbers reach the PDF through exactly one file, in exactly two representations`.
Cross-references at spine lines 634/637 checked — both cite `AD-3` bare, without restating the
superseded wording, so neither needed amendment.

**Finding 13 — AC10's content-derivation clause has no assertion.** **FIX.** Added
`internal/pdf/document_test.go`: `TestComputeIDIsContentDerived` (two different inputs to
`computeID` must differ) and `TestSerializeIDMatchesReDerivedDigest` (re-derives the digest from
`Serialize()`'s own pre-`/ID` prefix and compares it to the emitted entry).

### Minors

**Finding 14 — `epics.md` and the decision log restate AD-3's superseded rule.** **FIX** for
`epics.md:147` and `epics.md:439` (explicit orchestrator instruction; both amended under D-000.6 to
the two-representation form, with the `epics.md:439` AC narrowed to "geometric" per D-1.1.b, as
instructed). **DISMISS** for `folio-mvp-decision-log.md:43`: that line sits inside the "Lead
Grounding" section, which is explicitly stated to be filed at the run's start and the log's own
header states entries are "append-only: a reversal is appended, never a rewrite." D-000.6's
amendment mechanism is scoped to `ARCHITECTURE-SPINE.md`, `folio-format.md` and the SPEC's
companions list — not the decision log itself, which has its own, incompatible convention (append,
don't edit). Editing a historical grounding note would violate that convention for no benefit: the
ruling that supersedes it (D-1.1.b) already exists later in the same file as the authoritative,
appended correction.

**Finding 15 — `numbers.go`'s header comment describes an enforcement/exemption that don't exist.**
**FIX**, and now true rather than aspirational: reworded to describe the merged
`TestNumberFormattingIsConfinedToNumbersGo` guard (Finding 2's fix), which does now enforce exactly
what the comment claims, file-scoped, no exemptions.

**Finding 16 — AC7's vacuity guard counts the test's own directory.** **FIX.** `arch_test.go`'s
guard now requires `packagesVisited["geom"]` and `packagesVisited["pdf"]` **by name**, and counts
declarations only from files where `pkgDir != "."`.

**Finding 17 — a ≥10-digit xref offset silently produces a 21-byte entry.** **FIX.** `appendXref`
now panics if an offset exceeds `9999999999` (the 10-digit field width), before calling
`appendIntPadded` — `appendIntPadded` itself keeps its existing general (widen, don't truncate)
contract, since only the xref call site actually depends on the fixed-width invariant.

**Finding 18 — no test asserts the PDF's internal self-consistency.** **FIX**, same change as
Finding 4 (`assertWellFormedPDF`'s xref/offset/`/Length` checks).

**Finding 19 — `expected.pdf` is never read by any test.** **FIX.** `TestRenderMatchesGoldenFixture`
now reads `expected.pdf`, hashes it, and asserts the hash equals `expected.json`'s recorded
`sha256`.

**Finding 20 — no assertion on the no-compression/no-metadata half of the contract.** **FIX.**
`TestRenderHasNoCreationOrModDate` extended to also assert `/Producer`, `/Creator`, `/Metadata` and
`/Filter` are absent.

**Finding 21 — `geom.Rect` is dead code.** **FIX.** Both the MediaBox (via `appendMediaBox(dst,
geom.Rect{...})`) and the fill rectangle (`buildContentStream`'s `fillRect := geom.Rect{...}`) now
construct and use a real `geom.Rect` value, giving Task 3's stated rationale ("needed by the
rectangle") a caller — and giving Finding 1's fix a natural home, as the review suggested.

**Finding 22 — the plain-terms opener describes the plan, not the shipped result.** **FIX.**
Rewritten past-tense, describing the shipped two-representation number-writing shape (replacing the
overturned "one single piece of code" claim) and naming, honestly, that review found and closed a
couple of provenance gaps before the story was finished. Scope fence kept verbatim, as instructed
and as the review recommended. 349 words, no file paths/function names/code.

### Nits

**Finding 23 — Delivery Log's test count (30) doesn't match a real run (29 measured by review).**
**FIX.** Corrected in the Delivery Log to the actual current count, re-measured after all finisher
changes: 23 top-level test functions + 19 subtests = 42, across 4 packages, all passing.

**Finding 24 — `-test.run=^TestMain$` matches no test function.** **FIX.** Changed to
`-test.run=^$` with a comment explaining the env-var short-circuit is what actually routes the
child process.

**Finding 25 — `os.Stdout.Write`'s error/short-write is discarded in the subprocess path.** **FIX.**
`TestMain`'s subprocess branch now checks both the write error and the byte count, exiting non-zero
with a distinguishing message on either failure.

**Finding 26 — the page-count check depends on an incidental trailing space.** **FIX**, folded into
Finding 4's `assertWellFormedPDF` rewrite: `countPageObjects` now explicitly excludes
`/Type /Pages` by checking the byte immediately following the `/Type /Page` match, rather than
relying on a literal trailing space.

**Finding 27 — `appendIntPadded`'s doc comment misdescribes its own padding for negative values.**
**FIX.** Doc comment reworded to state the actual (and already-tested, already-intended) behaviour:
width counts digits, excluding any sign.

**Finding 28 — D-000.6's before/after wasn't fully quoted.** **FIX.** The superseded AD-3 Rule
paragraph is now quoted in the Delivery Log's "Spine amendment (D-000.6)" section, alongside the
resulting text that was already there.

---

## Superseded

**By Story 1.3, ruling D-1.3.2.** This file's AC6 text, and the two places that restate it
(lines **202**, **327** and **678**), describe the numeric-formatting guard's scope as "nowhere
under `folio-go/`" — module-wide. Story 1.3 measured (F-2) that this over-broad scope blocks
Story 1.4's `internal/template` error path (`fmt.Errorf` naming the declared and supported
versions) before that story can even start, and narrowed the guard's scope to exactly
`folio-go/internal/pdf/`, deleting the module-wide half outright rather than retaining it with an
exemption list — "AD-5 is why deletion is safe: nothing outside `internal/pdf` writes an output
byte, so the wide half protects nothing while costing AD-3's diagnostic carve-out" (D-1.3.2).

Per this project's append-only convention for completed story files, **lines 202, 327 and 678 are
left intact above, unedited, and this file's AC6 text is not rewritten** — rewriting it would claim
this story verified something it did not. The guard's current, narrowed behaviour is `folio-go`'s
`internal/pdf/emit_source_test.go` (`TestNumberFormattingIsConfinedToNumbersGo`), scoped to exactly
`folio-go/internal/pdf/`, `_test.go` files included, `numbers.go` exempted, any `testdata/`
directory excluded. `epics.md` was verified twice (D-1.3.2) to already state the narrowed scope
correctly, so no `D-000.6` spine amendment applies to this narrowing.
