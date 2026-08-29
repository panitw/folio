# Folio API digest — round 1

- Public flow: `ParseTemplate([]byte)` creates an opaque template; `Render` accepts template, Data JSON, Params JSON, and FontSet, returning complete PDF bytes plus diagnostics. `Validate` predicts failures without producing a PDF. Sources: `/Users/panitw/Projects/folio/folio-go/folio.go`, `render_entry.go`, `diagnostic.go`, `validate.go` (inspected 2026-08-28). Confidence: high. Class: current implementation.
- Fonts are caller-provided named byte arrays; the optional fonts package embeds three Noto faces and adds about 11.3 MB raw. Sources: `fontset.go`, `fonts/fonts.go`, `README.md`. Confidence: high. Class: packaging.
- The module pins Go 1.26.0, declares a Go 1.25 floor, and has one direct dependency (`github.com/boxesandglue/textshape v0.0.15`). No cgo declaration or C/C++ source exists in the Folio subtree. Source: `go.mod` and exhaustive repository search. Confidence: high for Folio; transitive dependency status required a separate check. Class: version/compatibility.
- The core embeds a Thai dictionary and initializes it with `sync.Once`; rendering is otherwise designed to avoid host filesystem, network, time, and locale state. Source: `internal/text/data.go` and `render_entry.go`. Confidence: high. Class: architecture.
- Render intentionally buffers the whole PDF, so a pointer/length return buffer matches the engine better than a streaming ABI. Source: `render_entry.go`. Confidence: high. Class: architecture.
- Errors and warnings already have a stable structured form: code, severity, element ID, data path, and message; successful PDFs can carry ordered diagnostics. Sources: `render_error.go`, `diagnostic.go`. Confidence: high. Class: API contract.
- No C ABI exists. A new package-main bridge must translate Go-native values into C primitives. A one-shot versioned render call is simpler than cached handles; handles additionally require lifetime and concurrency rules. Confidence: high on required adaptation, medium on preferred design. Class: inference.
- Shared-template concurrent safety is not explicitly documented. The initial wrapper should avoid shared handles or serialize them until a race-tested contract exists. Confidence: medium. Class: risk.
- The existing CLI already supplies a viable sidecar seam using byte/file inputs, PDF output, diagnostics, and exit status. Source: `cmd/folio/main.go`. Confidence: high. Class: alternative architecture.

Open issues: actual Windows c-shared build, concurrency/race testing, dependency portability, and licensing review.

