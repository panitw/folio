# Research brief

## Decision

Can the Go-based Folio PDF renderer be wrapped and distributed as a .NET library consumable by .NET Framework 4.6 applications?

## Approved scope

- Inspect Folio's exported Go API, dependency graph, binary assets, state, concurrency, and platform assumptions.
- Compare native shared-library/C ABI plus P/Invoke, sidecar executable, local service, and managed port approaches.
- Establish .NET Framework 4.6, NuGet, native loading, architecture, memory ownership, deployment, and build constraints.
- Recommend an architecture and a bounded proof of concept, with risks and verification criteria.

## Research method

Technical research; breadth-first topology; standard preset; normal verification. Repository files are explicitly admitted as primary evidence for Folio-specific claims. External compatibility claims require sources retrieved during this run.

