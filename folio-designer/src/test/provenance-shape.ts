// THE `source` PROVENANCE SHAPE, SHARED BY THE TWO SUITES THAT ASSERT IT
// (D-16.R.13, DW-160). Test-only, and it lives beside `sfnt-fixture.ts` for
// that reason.
//
// It is shared rather than duplicated because the defect this tripwire guards
// is **the two tiers drifting apart**. Two copies of the predicate can drift
// exactly as the two writers did, and then each tier would be checked against
// its own idea of the rule — which is the failure one level up from the one
// being prevented.
//
//   `src/font-provenance.test.ts`  — the committed tier, over every generated
//                                    catalogue row, plus `webFaceSource`.
//   `src/font-source.test.ts`      — the fetched tier's REAL write path, on an
//                                    outcome from `fetchWebFamily` with no
//                                    `today` argument, so the default date
//                                    expression is observed rather than
//                                    bypassed.

/**
 * A HOST, RECOGNISED BY ITS TLD RATHER THAN BY A LIST OF KNOWN FONT HOSTS.
 *
 * A list of the hosts this product does reach would pass the moment somebody
 * reached for a new one, which is the whole failure shape. The TLD set is
 * deliberately narrow so that a FILENAME does not read as a host: `.ttf`,
 * `.pb`, `.md` and a version like `4.005_Desktop` are not in it.
 */
export const hostShaped = /\b[a-z0-9][a-z0-9-]*\.(?:com|org|net|io|dev|co|app|sh|xyz)\b/i
export const schemeShaped = /[a-z][a-z0-9+.-]*:\/\//i
export const digestShaped = /\b[0-9a-f]{64}\b/i

/**
 * A MOVING REF OF ANY NAME, not just the one that was actually there.
 *
 * The defect was `main`. Pinning the tripwire to that word alone would let the
 * identical defect back in spelled `develop`, `trunk` or `Main` — a guard
 * fitted to the instance rather than to the class. Case-insensitive for the
 * same reason.
 */
export const branchShaped = /(?:^|[/@\s])(?:main|master|trunk|develop|dev|head|latest|default)(?:[/\s]|$)/i

/**
 * THE PROJECT HALF, SHAPED — `<owner>/<repo>` with an optional `@<release>`.
 *
 * Asserting merely that it is non-empty is very nearly vacuous: `split(' — ')`
 * returns the WHOLE string when the separator is absent, so a `source` with no
 * separator at all would pass a non-empty check while carrying no path and no
 * project. The separator's presence is therefore asserted first, and the half
 * before it is held to a shape.
 */
export const projectShaped = /^[A-Za-z0-9][A-Za-z0-9._-]*\/[A-Za-z0-9][A-Za-z0-9._-]*(?:@[A-Za-z0-9][A-Za-z0-9._-]*)?$/

/**
 * A TYPE-ONLY reference to vitest's `expect`, so this module carries no runtime
 * import of the test runner while still being fully typed at the call sites.
 */
type Expect = (typeof import('vitest'))['expect']

/**
 * `source` NAMES PROVENANCE, NOT A RETRIEVAL PATH. Three prohibitions and two
 * positive requirements, applied identically to both tiers.
 *
 * `expect` is passed in rather than imported so this module stays free of a
 * test-runner import while being usable from either suite.
 */
export function assertProvenanceShape(expect: Expect, tier: string, subject: string, value: string): void {
  expect(value, `${tier}: ${subject} publishes an empty source`).not.toBe('')
  expect(value, `${tier}: ${subject} carries a URL scheme in \`source\`. A resolvable-looking string is a promise of fetchability, and a promise that decays reads as broken provenance (D-16.R.13).`).not.toMatch(schemeShaped)
  expect(value, `${tier}: ${subject} carries a HOST in \`source\`. \`source\` names provenance, not a retrieval path (D-16.R.13).`).not.toMatch(hostShaped)
  expect(value, `${tier}: ${subject} carries a moving ref in \`source\`. A branch does not identify the bytes the field claims to describe — that was the defect (D-16.R.13).`).not.toMatch(branchShaped)
  expect(value, `${tier}: ${subject} restates a SHA-256 in \`source\`. The digest is already the asset key, and duplicating it puts two authorities on one fact (D-16.R.13).`).not.toMatch(digestShaped)

  // AND IT IS STILL PROVENANCE, not merely a string that avoids three
  // prohibitions: a project, a path within it, and a fetch date.
  expect(value, `${tier}: ${subject} does not name a fetch date`).toMatch(/, fetched \d{4}-\d{2}-\d{2}$/)
  expect(value, `${tier}: ${subject} carries no ' — ' separator, so it names no project and no path within one`).toMatch(/ — /)
  expect(value.split(' — ')[0], `${tier}: ${subject}'s project half is not shaped <owner>/<repo>[@<release>]`).toMatch(projectShaped)
}
