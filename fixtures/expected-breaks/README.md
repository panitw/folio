# `fixtures/expected-breaks/` — S4's frozen expected-break fixture (Story 2.4)

`acceptance.md:65` and `epics.md:836` both require an expected-break fixture for Thai and CJK line
breaking, **hand-checked once and then never regenerated to make a test pass**. This is it.

## This is the oracle. `thai-break-corpus/computed_breaks.json` is not.

The two files look similar and answer opposite questions:

| | `expected-breaks/` | `thai-break-corpus/computed_breaks.json` |
|---|---|---|
| Answers | *is the answer right?* | *does this target agree with the others?* |
| Written by | a person, by hand | `cmd/genbreaks`, from the engine |
| On a mismatch | the **engine** is wrong (or the label is) | a **target** disagrees |
| Regenerating it | never | expected, on an intended change |

That distinction is in `thai-break-corpus/README.md`'s own words — it is *"a cross-target REGRESSION
ANCHOR ONLY, never a correctness oracle"*. **There is deliberately no generator for this directory.**
Nothing in the repository can write `expected_breaks.json`.

## How to check it

Read the `words` field of each item — that is the label. `expectedBreaks` is only its arithmetic (the
cumulative rune lengths of every word but the last), and
`TestS4ExpectedBreaksAreLabelsNotEngineOutput` enforces that the numbers follow the words, so you
cannot quietly move a number without moving a word boundary a human can see.

For each item ask one question: **would a reader accept a line ending at each of those seams, and not
inside any of the words?**

## A break opportunity is not an etymology — and there is no derivable rule for these labels

The first draft of this fixture split `thai-003`…`thai-010` into their historical morphemes — *school*
as hall + study, *train* as vehicle + fire — and labelled each seam as a break. **That was wrong.**
Every one of those eight is a single dictionary **headword**, and the first correction re-labelled all
eight as one word each, on a claimed principle: *headword membership in the shipped wordlist*.

**That principle was itself falsified by the owner's 2026-08-24 hand-check.** `หนังสือพิมพ์`
(newspaper), `วันเกิด` (birthday) and `ที่อยู่` (address) are headwords — and so are their split
constituents (`หนังสือ`+`พิมพ์`, `วัน`+`เกิด`, `ที่`+`อยู่`) — and so is every one of the five items
kept whole (`thai-003`/`004`/`005`/`006`/`010`) and their constituents too. **Headword membership does
not, and cannot, distinguish the three the owner split from the five they kept.** The owner corrected
`thai-007`/`008`/`009` to carry a break anyway.

**These labels are per-item native-speaker judgments. No rule derivable from the shipped wordlist
predicts them** — this is AD-25's stated impossibility (*"the engine never infers membership; it
cannot"*, the Surname Act argument), observed here from the **acceptability** side rather than the
surname side. **The fixture is the oracle BECAUSE no rule can replace it.**

Three states, recorded separately in `expected_breaks.json`'s own `_README` — silence is not
ratification:
- **explicitly corrected**: `thai-007`, `thai-008`, `thai-009` (split), and `cjk-004`, `cjk-005`,
  `cjk-007` (declared unbreakable — see below);
- **explicitly declined**: none remain — `cjk-005` was re-asked with its actual break positions shown
  and the owner changed the answer;
- **unremarked, therefore not ratified**: `thai-003`/`004`/`005`/`006`/`010`.

`thai-001` and `thai-002` remain the contrast for the compounds kept whole: sequences that are *not*
headwords though both of their parts are, and which therefore do carry a break.

AC14's exact-equality assertion carries a small, enumerated, **fail-closed-only** exception for
`thai-007`/`008`/`009` — see `s4ExpectedDivergences` in
`folio-go/internal/text/s4_expected_test.go`. The engine cannot be made to propose these breaks without
inventing a heuristic (forbidden by AD-25) or editing the shipped wordlist to make a real Thai word
disappear (forbidden, D-000.32) — the divergence is named and bounded rather than hidden inside a
relaxed assertion.

`cjk-004`, `cjk-005` and `cjk-007` take the opposite correction: the owner wants **fewer** breaks, which
*is* expressible. Each now carries `"declaredAtomic": true`, D-2.1.6's declaration channel reaching this
fixture for the first time — the whole value is passed to the engine as a declared, never-split unit
(AD-25's third mechanism), so the engine reproduces "no interior break" because it was told, not because
anything was inferred.

The correction is recorded rather than quietly applied, because "the fixture changed and then the test
passed" is exactly the shape this fixture exists to make impossible.

## Coverage

25 items — 9 expecting at least one break, 16 expecting none. Both polarities are asserted to be
non-empty, so neither an engine that never proposes a break nor one that proposes them everywhere can
pass. Thai personal names (`ชัยวัฒน์`, `ดอเลาะ`, `แนแซ`, `ฉั่วสมบูรณ์`) appear as zero-break labels:
they are AD-25's atomic-unknown-run absolute stated as a conformance expectation.

## Capability limit, stated where the narrowing already lives

`folio-format.md`'s line-breaking section narrows its UAX #14 claim by name (no hyphenation, no break at
`-`, no contextual pair rules). It also now states, in the same register: **no break inside a dictionary
headword, including lexicalised compounds a native reader would accept breaking.** This is a stated
capability limit, not a hidden one, and it is fail-closed — the compound moves to the next line whole,
never renders wrongly.

## Pending: the human sign-off

**The labels here were authored by an agent, so an agent confirming them would be marking its own
work.** `folio-go/expected_breaks_signoff_matrix_test.go` is a `//go:build matrix` test that **fails**
until `break-signoff.json` exists, naming this file's (now corrected) sha256. The Epic 2 boundary gate
cannot pass until a person has signed off on the corrected labels — requested only now that this
correction has landed (D-000.41, D-000.43).

This is **separate** from `fixtures/shaped-text/thai-signoff.json`, which answers a different question
(*does this Thai render legibly*) and is bound to a different digest. Resolving one does not resolve
the other, and re-recording either invalidates exactly one of them (D-2.4.3).
