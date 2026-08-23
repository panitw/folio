// This root module contains NO third-party code (D-1.3.8). It exists
// only to give Story 1.3's licence checker a graph to walk: four
// sibling stub modules, each declaring a different forbidden copyleft
// licence via a short SPDX marker, resolved entirely through local
// `replace` directives (AC30) — never fetched from a proxy.
module example.test/copyleft-root

go 1.25.0

require (
	example.test/gpl-lib v0.0.0-00010101000000-000000000000
	example.test/lgpl-lib v0.0.0-00010101000000-000000000000
	example.test/agpl-lib v0.0.0-00010101000000-000000000000
	example.test/sspl-lib v0.0.0-00010101000000-000000000000
)

replace (
	example.test/gpl-lib => ./example.test/gpl-lib
	example.test/lgpl-lib => ./example.test/lgpl-lib
	example.test/agpl-lib => ./example.test/agpl-lib
	example.test/sspl-lib => ./example.test/sspl-lib
)
