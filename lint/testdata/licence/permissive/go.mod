// This root module contains NO third-party code (D-1.3.8). It exists
// only to give Story 1.3's licence checker a graph to walk: three
// sibling stub modules declaring MIT, Apache-2.0 and BSD-3-Clause via a
// short SPDX marker, resolved entirely through local `replace`
// directives (AC30) — never fetched from a proxy.
module example.test/permissive-root

go 1.25.0

require (
	example.test/mit-lib v0.0.0-00010101000000-000000000000
	example.test/apache-lib v0.0.0-00010101000000-000000000000
	example.test/bsd-lib v0.0.0-00010101000000-000000000000
)

replace (
	example.test/mit-lib => ./example.test/mit-lib
	example.test/apache-lib => ./example.test/apache-lib
	example.test/bsd-lib => ./example.test/bsd-lib
)
