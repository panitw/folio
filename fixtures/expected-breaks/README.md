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

## A break opportunity is not an etymology

The first draft of this fixture split `thai-003`…`thai-010` into their historical morphemes — *school*
as hall + study, *train* as vehicle + fire — and labelled each seam as a break. **That was wrong.**
Every one of those eight is a single dictionary **headword**, and a reader no more breaks inside them
than an English reader breaks *notebook* after *note*.

The labels were corrected to one word each, on evidence independent of the engine's output (headword
membership in the shipped wordlist), and those eight are now the fixture's **negative controls**.
`thai-001` and `thai-002` are the contrast: sequences that are *not* headwords though both of their
parts are, and which therefore do carry a break.

The correction is recorded rather than quietly applied, because "the fixture changed and then the test
passed" is exactly the shape this fixture exists to make impossible.

## Coverage

25 items — 9 expecting at least one break, 16 expecting none. Both polarities are asserted to be
non-empty, so neither an engine that never proposes a break nor one that proposes them everywhere can
pass. Thai personal names (`ชัยวัฒน์`, `ดอเลาะ`, `แนแซ`, `ฉั่วสมบูรณ์`) appear as zero-break labels:
they are AD-25's atomic-unknown-run absolute stated as a conformance expectation.

## Pending: the human hand-check

**The labels here were authored by an agent, so an agent confirming them would be marking its own
work.** `folio-go/expected_breaks_signoff_matrix_test.go` is a `//go:build matrix` test that **fails**
until `break-signoff.json` exists, naming this file's sha256. The Epic 2 boundary gate cannot pass
until a person has read the labels.

This is **separate** from `fixtures/shaped-text/thai-signoff.json`, which answers a different question
(*does this Thai render legibly*) and is bound to a different digest. Resolving one does not resolve
the other, and re-recording either invalidates exactly one of them (D-2.4.3).
