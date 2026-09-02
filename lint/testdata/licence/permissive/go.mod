// This root module contains NO third-party code (D-1.3.8). It exists
// only to give Story 1.3's licence checker a graph to walk: four
// sibling stub modules declaring MIT, Apache-2.0, BSD-3-Clause and
// Ubuntu-font-1.0 via a short SPDX marker, resolved entirely through
// local `replace` directives (AC30) — never fetched from a proxy.
//
// compound-lib joined at Story 8.5 (Design Note 3, D-8.4j.2): the census must
// pin at least one COMPOUND SPDX expression so Story 8.4j's whole-line fix
// stays fixed, and no procurable font carries a compound expression the
// four-id allowlist admits. Both of its terms are permissive, so the subtest
// below still expects zero findings.
//
// ufl-lib joined at Story 8.4h (D-8.5.3): the fourth member of the
// owner's asset allowlist, exercised end to end through the graph scan
// rather than only through a classifier table. The scan switches on
// licence FAMILY and discards the id, so what this proves is that the
// permissiveSPDX entry is LIVE — remove it and the id classifies as
// unknown and this subtest reds. The id itself is pinned in
// lint/internal/licence/classify_test.go, not here.
module example.test/permissive-root

go 1.25.0

require (
	example.test/mit-lib v0.0.0-00010101000000-000000000000
	example.test/apache-lib v0.0.0-00010101000000-000000000000
	example.test/bsd-lib v0.0.0-00010101000000-000000000000
	example.test/ufl-lib v0.0.0-00010101000000-000000000000
	example.test/compound-lib v0.0.0-00010101000000-000000000000
)

replace (
	example.test/mit-lib => ./example.test/mit-lib
	example.test/apache-lib => ./example.test/apache-lib
	example.test/bsd-lib => ./example.test/bsd-lib
	example.test/ufl-lib => ./example.test/ufl-lib
	example.test/compound-lib => ./example.test/compound-lib
)
