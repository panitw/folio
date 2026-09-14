---
title: 'Section Break — unanchored section follows content onto a later page'
type: 'bugfix'
created: '2026-09-14'
status: 'done'
route: 'oneshot'
context:
  - '{project-root}/_bmad-output/specs/spec-section-break/SPEC.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Take an unanchored break whose above-line rows fill page 1 and end on page 2 higher than the line's declared offset. The section lands on page 2 at its declared y, not directly after the rows, because `paginateWithSectionBreak` sends any "ends at or above the line on its last page" case down the anchored path (owner report after 7b8a845).

**Approach:** When the break is unanchored and the content above reaches a page after page 1, the line sits at E, where that content ends on its last page, whether E is above or below the declared line. The same fit rule applies: if the section fits, it is shifted by D = E − line on that page, keeping its offsets from the line; otherwise it moves to a new page with the line at the window top. Page 1 is unchanged: never pulled up, and CAP-3 byte identity holds. Anchored breaks are unchanged. `folio-format.md` Pagination rules 3 and 5 and the unanchored tests follow the new rule, and a named case pins rows that end above the line on page 2.

</frozen-after-approval>

## Implementation Notes

- `folio-go/section_break.go`: an unanchored break now takes the move path when `!shared || len(planA.Pages) > 1`. `shared` is reset to false at the start of that path. Without the reset, a later-page section with no room would still be merged onto E's page instead of moving to a new page. D = E − line may be negative. Members sit at or below the line, so every moved item stays at or below E, which is at or below the window top.
- Page 1 content ending at or above the line still takes the CAP-3 fast path, unchanged. Anchored breaks never enter the move path.
- Tests (`folio-go/section_break_test.go`):
  - The oracle's "not crossed" branch now applies to page 1 only. Later-page rows ending above the line must land at `declaredY + (E − line)`, and a `pulled` witness proves that case is reached.
  - New named case: 12 rows end at 42.7pt on page 2, with the line at 75pt.
- Row ends were measured with a throwaway test, since deleted.
- `folio-format.md`: rule 3 limits no-pull-up to the first page. Rule 5 now covers content that reaches a later page, where the section is moved by E − the break in either direction.
- Unaffected: the `section-break-unanchored` golden (one page) and the designer (no projection change).
- Verification: `gofmt`, `go vet` and `go test ./...` are clean, apart from the known baseline failure `TestCorpusMeetsP6ExerciseFloors`.

## Review Triage Log

Blind hunter:

| # | Finding | Verdict | Evidence | Route |
|---|---|---|---|---|
| 1 | A tall section whose rows end above the line on page 2 now goes to page 3 at the window top | false | This is the frozen fit rule: a section needing more than one window never fits, so it moves to a new page with the line at the top. It is the same outcome as a crossing on page 1 (`TestSectionBreakUnanchoredTallSectionStartsAtTheWindowTop`) | reject |
| 2 | `folio-format.md` rule 2 still says "same page, anchored or not" and conflicts with rule 5 | low | Confirmed; direct wording correction | patch |
| 3 | The 12-row named case does not assert the legend's y or that the rows end above the line | low | Confirmed; the oracle covers it exactly, but the named case alone passes on old code | patch |
| 4 | The branch condition is always true after the fast path, and `shared` carries two meanings | low | Confirmed: the fast path returns on `shared && pages == 1`, so the condition reduces to `sb.unanchored`, now simplified. Splitting the variable is a naming preference with no named caller at risk | patch (condition) |
| 5 | Stale "pushed" wording in oracle messages and in the named-case `pushed` field | low | Confirmed; messages reworded and the field meaning commented | patch |
| 6 | The older build spec and memlog assumption still describe never-pulled-up | low | The build record and memlog are append-only history. The new memlog decision supersedes the assumption, and SPEC.md is current | reject |
| 7 | New artifact still `in-progress` with empty Implementation Notes | false | The review ran while the notes were being written; status is set at finalize | reject |
| 8 | No golden fixture covers the pulled-up path | low | Real, but the path uses the same integer translation as the pushed golden, which is already in the hash matrix. A new fixture is more than a direct correction | reject |
| 9 | CAP-3 contradicts the change; CAP-7 intent says only "pushed down" | low | The CAP-3 part is false: rows that fill page 1 crossed the line there. The CAP-7 intent wording lag is real, but it is a spec edit | defer |
| 10 | Rule 3 wording is ambiguous and gives no reason for the page-1 asymmetry | low | Confirmed; reworded with the CAP-3 reason | patch (with 2) |
| 11 | No anchored regression test for rows ending above the line on page 2 | false | `TestSectionBreakLandingMatchesTheOracleForEveryRowCount` (anchored) sweeps 0–20 rows, including rows 10–14 and 19–20 that end above the line on later pages, and passed | reject |
