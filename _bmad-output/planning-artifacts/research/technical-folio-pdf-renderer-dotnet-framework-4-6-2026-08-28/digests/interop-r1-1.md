# Go/.NET interop digest — round 1

## Findings

- Go can build one `main` package as a C-compatible shared library with `go build -buildmode=c-shared`; functions carrying cgo `//export` declarations become callable and the build generates a C header. Source: https://go.dev/cmd/go/ (Go project; accessed 2026-08-28). Confidence: high. Class: version/compatibility.
- The wrapper needs a purpose-designed C ABI. Ordinary Go structs, slices, maps, interfaces, errors, and pointers cannot safely or directly represent the exported contract. C-facing data should use fixed-width scalars and pointer/length pairs. Source: https://pkg.go.dev/cmd/cgo (Go project; accessed 2026-08-28). Confidence: high. Class: version/compatibility.
- `runtime/cgo.Handle` can represent Go values as integer tokens across foreign code, but handles must be explicitly deleted and invalid access panics. Source: https://pkg.go.dev/runtime/cgo (Go project; accessed 2026-08-28). Confidence: high. Class: architecture.
- .NET Framework can invoke unmanaged exports with `DllImport`. Managed declarations must match native types and calling convention; this wrapper should specify `CallingConvention.Cdecl`, exact spelling, and byte buffers rather than ambiguous strings. Source: https://learn.microsoft.com/en-us/dotnet/framework/interop/creating-prototypes-in-managed-code (Microsoft; accessed 2026-08-28). Confidence: high. Class: version/compatibility.
- Native output should cross as `IntPtr` plus length, be copied into managed memory, and be released only by a matching native `folio_free`; automatic string marshalling obscures allocator ownership. Same Microsoft source. Confidence: high. Class: architecture.
- Managed callbacks are possible but add delegate-rooting and reentrancy hazards; the initial ABI should use synchronous return values and avoid callbacks. Source: https://learn.microsoft.com/en-us/dotnet/framework/interop/marshalling-a-delegate-as-a-callback-method (Microsoft; accessed 2026-08-28). Confidence: high. Class: architecture.
- Process and native DLL architectures must match; mismatches can cause `BadImageFormatException`. Distribution therefore needs architecture-specific assets or an explicitly x64-only target. Source: https://learn.microsoft.com/en-us/dotnet/api/system.badimageformatexception (Microsoft; accessed 2026-08-28). Confidence: high. Class: version/compatibility.
- Windows `c-shared` builds require cgo and a C compiler; cross-compilation needs explicit configuration. Release builds must pin the Go and C toolchains. Source: https://pkg.go.dev/cmd/cgo (Go project; accessed 2026-08-28). Confidence: high. Class: implementation.
- A sidecar remains viable on .NET Framework 4.6 via redirected process streams or named pipes. It improves fault isolation and ABI evolution but adds process/protocol lifecycle and IPC cost. Sources: https://learn.microsoft.com/en-us/dotnet/api/system.diagnostics.processstartinfo.redirectstandardinput?view=netframework-4.8.1 and https://learn.microsoft.com/en-us/dotnet/api/system.io.pipes.namedpipeserverstream (Microsoft; accessed 2026-08-28). Confidence: high for capability, medium for tradeoff. Class: architecture.

## Recommended experimental ABI

Use a synchronous `folio_render(request_ptr, request_len, pdf_out, pdf_len, error_out, error_len)` plus `folio_free` and `folio_version`. UTF-8/opaque bytes always carry explicit lengths. Convert panics at every export boundary, use explicit status codes, document concurrency, and begin with Windows x64.

## Contrary evidence and gaps

`c-shared` is officially supported, so a sidecar is not inherently required. However, official documentation does not guarantee that fatal Go-runtime failures can be isolated inside the managed process. Build the DLL on the actual Windows CI image, inspect exports, test dependency portability, and benchmark DLL versus sidecar on Folio fixtures.
