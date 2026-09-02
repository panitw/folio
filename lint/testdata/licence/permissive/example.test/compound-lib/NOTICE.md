This stub module contains NO third-party code and no font. It exists only to be
classified by Story 1.3's licence checker fixture graph (D-1.3.8): a module we
authored ourselves, whose LICENSE file is fixture data we wrote in order to be
classified, creates no licensee relationship and transfers no copyrighted work.
It is not a real software licence grant.

Added at Story 8.5 (Design Note 3) to satisfy D-8.4j.2: the census must pin at
least one COMPOUND SPDX expression, so the compound-line fix Story 8.4j made
(`spdxLineRE` capturing the whole line, resolved per term) stays fixed.

WHY A FIXTURE AND NOT A FONT. The compound case cannot come from a real
catalogue face, and the domain was measured before the decision rather than
after (D-000.12): NO procurable family carries a compound expression admissible
under the owner's four-id allowlist. The two compound font licences found —
Hack (MIT + Bitstream Vera) and Public Sans (OFL-1.1 + CC0-1.0) — must both
FAIL the asset gate, correctly, because per-term admission refuses a term that
is not one of the four. So a committed font could only ever pin a compound
expression that is REFUSED, which is not the case D-8.4j.2 asks for.

`MIT OR Apache-2.0` is chosen because both terms are on the owner's four-id
font allowlist AND on `permissiveSPDX`, so it resolves permissive through every
consumer of `ClassifySPDXExpressionTerms` rather than only through one.

WHAT IT DOES NOT DO. It changes no gate, no allowlist and no classifier table —
this story may not touch the licence gate at all (D-000.11). It is a committed
file with a pinned verdict, and nothing more.
