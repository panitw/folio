# Epic 6 Context: A template author can bind a report to data and build the golden report

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Enable a template author to discover supplied report data, bind it safely to report elements, configure a table, and prove that the browser preview and native renderer produce the same report bytes. The workflow stays local and offline: the author supplies a sample JSON document and Folio never connects to a database or server.

## Stories

- Story 6.1: Load sample data and browse its paths
- Story 6.2: Bind a component by picking a path
- Story 6.3: Supply parameters for preview
- Story 6.4: Configure table columns as a matrix
- Story 6.5: Bind columns in row scope and configure footer aggregates
- Story 6.6: Present a failed render honestly
- Story 6.7: The round trip closes

## Requirements & Constraints

Sample JSON is supplied from the local machine as a companion to a template and powers path discovery and exact Preview. Binding is selected from discovered paths, with scalar and collection behavior made explicit; no database, cloud transport, or account is involved. Preview remains a mode switch that presents the engine-produced PDF, not a browser approximation. The final authoring round trip must match a native render byte-for-byte.

## Technical Decisions

The Go/wasm engine exclusively parses, holds, mutates, validates, and serializes `.folio`; TypeScript paints immutable snapshots and sends committed document commands only. Sample loading is transient UI state rather than an engine command or undoable document state. One dedicated worker owns one wasm module/instance. Preview uses serialized template bytes plus raw sample data and parameter-document inputs, and its identity includes those inputs, engine version, and FontSet identity. File access uses one capability-selected local-file abstraction with a File System Access API tier and input/download fallback. Report data numeric literals must be preserved exactly by the engine and must never be silently narrowed to floating point.

## UX & Interaction Patterns

The binding panel is docked on the right of the workspace, never a separate destination. Its no-sample state says that binding is unavailable, offers a local load action, and does not block normal canvas authoring. The populated panel is a keyboard-operable navigable data tree. Interactive controls have visible cyan focus and accessible names; amber denotes data only. Preview rendering may block only the Preview surface, while the canvas stays usable. The UI must remain terse and technical and never imply server rendering or a cloud round trip.

## Cross-Story Dependencies

Story 6.1 establishes local sample discovery and Preview data input. Story 6.2 adds path-to-component binding; Story 6.3 supplies parameters; Stories 6.4 and 6.5 add table structure and row-scope binding. Story 6.6 adds failed-render presentation, and Story 6.7 validates the complete browser-to-native round trip.
