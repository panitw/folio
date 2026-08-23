// This root module contains NO third-party code (D-1.3.8). It exists
// only to give Story 1.3's licence checker a graph to walk: one sibling
// stub module with NO LICENSE file at all, resolved entirely through a
// local `replace` directive (AC30) — never fetched from a proxy — so the
// checker's unresolvable-licence path is proven to fail (AC29, V11).
module example.test/unknown-root

go 1.25.0

require example.test/mystery-lib v0.0.0-00010101000000-000000000000

replace example.test/mystery-lib => ./example.test/mystery-lib
