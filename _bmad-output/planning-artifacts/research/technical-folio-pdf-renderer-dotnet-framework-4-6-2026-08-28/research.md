---
title: 'Technical research: Folio PDF renderer .NET Framework 4.6 wrapper'
type: technical
topic: 'Folio PDF renderer .NET Framework 4.6 wrapper'
decision: 'Can the Go-based Folio PDF renderer be wrapped and distributed as a .NET library consumable by .NET Framework 4.6 applications?'
source: native-run
status: complete
preset: standard
validation: normal
claims_verified: 7
claims_unverified: 0
created: '2026-08-28'
updated: '2026-08-28'
---

# Technical research: Folio PDF renderer .NET Framework 4.6 wrapper

## Executive summary

**Yes.** Folio can be wrapped as a .NET library supporting .NET Framework 4.6. The recommended architecture is a thin `net46` managed facade over a versioned Windows native DLL built from a new Go `main` bridge with `-buildmode=c-shared`. Folio already accepts byte-oriented inputs and returns complete PDF bytes with structured diagnostics; its current dependency graph is pure Go and compiled successfully for Windows amd64 during this research. [1] [3] [4]

Start with Windows x64 and a stateless synchronous ABI. Pass bytes with explicit lengths and return native-owned PDF, diagnostic, and error buffers that .NET copies and releases through `folio_free`. Go pointers, strings, slices, maps, callbacks, and errors must remain behind the boundary. [5] [6]

The main risks are packaging and deployment: process bitness, legacy `packages.config`, secure DLL loading, and the Windows floor imposed by modern Go. Strict net46 also cannot consume a `netstandard2.0`-only facade, and .NET Framework 4.6 itself is retired. [7] [9] [10] [11] [12]

**Verdict:** proceed to a bounded Windows proof of concept. Production feasibility should be accepted only after a real NuGet package renders Folio fixtures from clean PackageReference and `packages.config` net46 applications.

## Folio fit

Folio's public flow is `ParseTemplate` followed by `Render`, taking template/data/parameter/font bytes and returning a complete PDF plus ordered diagnostics. `Validate` predicts failures without producing output. This is naturally adaptable to one interop request/response rather than exposing Go objects. [1]

Rendering deliberately buffers the entire PDF, so a pointer-and-length return fits better than streaming callbacks. The optional shipped fonts add about 11.3 MB raw; baking them into the first DLL simplifies the proof of concept, while custom-font transport can wait. [1] [2]

The module pins Go 1.26.0 and has one direct external dependency. A resolved 128-package inspection found zero cgo packages, and `GOOS=windows GOARCH=amd64 CGO_ENABLED=0` produced a 20 MB PE32+ Folio CLI. The renderer is therefore Windows-portable as pure Go; only the proposed export bridge introduces cgo. [3]

## Interop architecture

Go officially supports building one `main` package as a C shared library and exposes functions marked with cgo `//export`. [4] The ordinary Go API cannot cross that boundary directly; cgo's pointer and type rules require a small C ABI. [5]

Recommended v1 contract:

```c
int32_t folio_render_v1(
  const uint8_t* request, size_t request_len,
  uint8_t** pdf_out, size_t* pdf_len_out,
  uint8_t** diagnostics_out, size_t* diagnostics_len_out,
  uint8_t** error_out, size_t* error_len_out);
void folio_free(void* buffer);
uint32_t folio_abi_version(void);
```

The request is a versioned envelope containing template, data, params, and font policy. Every buffer has an explicit length. Each export catches recoverable panics, returns stable status codes, never exposes Go pointers, and documents allocator ownership. The managed layer uses classic `DllImport` with `CallingConvention.Cdecl`, exact export names, `IntPtr` outputs, and `finally`-based release. [5] [6]

Avoid callbacks and cached template handles in v1. Callbacks complicate managed delegate lifetime; handles add stale-access, destruction, and concurrency rules. Folio does not yet document that one parsed template is safe for concurrent renders.

## net46 and NuGet constraints

The facade must target `net46` directly. .NET Framework 4.6 implements .NET Standard only through 1.3; .NET Standard 2.0 compatibility starts at Framework 4.6.1. [7]

Proposed package shape:

```text
lib/net46/Folio.dll
runtimes/win-x64/native/folio_native.dll
runtimes/win-x86/native/folio_native.dll  # only if required
build/net46/Folio.targets
```

NuGet defines managed, native-runtime, and build assets for these roles. [8] Older `packages.config` projects do not reliably deploy P/Invoke dependencies through managed reference resolution, so conservative MSBuild copy targets—or secure preloading from an absolute controlled path—are required and must be tested. [9]

An AnyCPU facade still loads native code at the host process bitness. The native DLL must match or loading can fail with `BadImageFormatException`; x64-only is the simplest initial contract. [10] DLL resolution must avoid attacker-writable search locations. [14]

## Operational reality

Windows c-shared builds need pinned Go and C toolchains. Go 1.25+ Windows cgo requires DWARF 5 compiler support; GCC builds require sufficiently new binutils. Go 1.21+ also requires Windows 10 or Windows Server 2016, regardless of the older OS range historically associated with net46. [12]

.NET Framework 4.6 retired on 2022-04-26, so this target is a legacy compatibility concession rather than a supported baseline. [11] Meanwhile the repository deliberately pins Go 1.26.0 for PDF reproducibility even though Go 1.27 is now released; the package needs an explicit security-update and byte-requalification policy. [3] [15]

A sidecar around Folio's existing CLI is the fallback. .NET Framework 4.6 supports named pipes. [13] As an architectural inference, a separate process provides stronger crash isolation and a protocol boundary that can evolve independently, at the cost of process/protocol lifecycle. Benchmark it only if native deployment or in-process runtime risk fails the proof-of-concept gate.

## Cross-dimension insights

Folio's renderer is not the hard part: its deterministic byte API lowers ABI complexity, while strict net46 support raises installer and runtime complexity. The acceptance gate must therefore emphasize clean consumer installation, safe loading, and architecture selection—not merely producing one PDF.

Also, net46 support does not imply support for every OS that once ran net46. The pinned modern Go runtime creates a separate Windows 10/Server 2016 floor.

## Recommendations

1. Build an x64 `c-shared` proof of concept with only `folio_render_v1`, `folio_free`, and `folio_abi_version`. [3] [4]
2. Ship a direct `net46` facade using classic P/Invoke and map Folio's stable diagnostic fields into managed result/exception types. [6] [7]
3. Make PackageReference and `packages.config` installation tests release gates, including safe loading and wrong-bitness failures. [9] [10] [14]
4. Keep v1 stateless, synchronous, and shipped-font-only. Add custom fonts and cached handles only after measured demand. [1] [2]
5. Retain the sidecar as a fallback for crash isolation or restrictive deployment environments. [13]
6. Document the Windows floor, architectures, net46 retirement, pinned toolchains, and toolchain requalification policy. [11] [12] [15]

## Proof-of-concept gate

- Build the DLL/header on pinned Windows CI and inspect exports/dependencies.
- Render representative Folio fixtures through a net46 console app and byte-compare with native Go output.
- Preserve structured warnings and errors; test malformed inputs, repeated calls, concurrency, and memory growth.
- Consume the packed NuGet from clean PackageReference and `packages.config` samples.
- Verify deterministic, safe DLL selection and clear missing/wrong-architecture errors.
- Measure first-call latency, steady-state rendering, peak memory, and package size.

## Open questions

- Which Windows versions, process architectures, IIS settings, and project styles must real consumers support?
- Are shipped fonts sufficient for v1, or are arbitrary fonts mandatory?
- How will Folio, textshape, and embedded-font notices be propagated in NuGet?
- Is sharing a parsed template concurrently safe? A handle API waits on a race-tested contract.
- Does the pinned Windows c-shared build preserve Folio's byte identity across CI rebuilds?

## Source appendix

| Ref | Finding | Publisher | Pub date | Accessed | Confidence |
|---|---|---|---|---|---|
| [1] | Folio API, buffering, diagnostics | [Render](/Users/panitw/Projects/folio/folio-go/render_entry.go), [templates](/Users/panitw/Projects/folio/folio-go/folio.go), [validation](/Users/panitw/Projects/folio/folio-go/validate.go) | 2026-08-28 inspection | 2026-08-28 | High |
| [2] | Embedded font design and size | [Folio repository](/Users/panitw/Projects/folio/folio-go/fonts/fonts.go) | 2026-08-28 inspection | 2026-08-28 | High |
| [3] | Dependency graph, Windows build, and toolchain pin | [Run evidence](/Users/panitw/Projects/folio/_bmad-output/planning-artifacts/research/technical-folio-pdf-renderer-dotnet-framework-4-6-2026-08-28/digests/windows-build-evidence.md), [go.mod](/Users/panitw/Projects/folio/folio-go/go.mod) | 2026-08-28 | 2026-08-28 | High |
| [4] | c-shared contract | [Go project](https://go.dev/cmd/go/) | maintained | 2026-08-28 | High |
| [5] | cgo export and pointer rules | [Go project](https://pkg.go.dev/cmd/cgo) | maintained | 2026-08-28 | High |
| [6] | P/Invoke prototypes/marshalling | [Microsoft](https://learn.microsoft.com/en-us/dotnet/framework/interop/creating-prototypes-in-managed-code) | maintained | 2026-08-28 | High |
| [7] | .NET Standard compatibility | [Microsoft](https://learn.microsoft.com/en-us/dotnet/standard/net-standard) | 2025-11 | 2026-08-28 | High |
| [8] | NuGet asset layout | [Microsoft/NuGet](https://learn.microsoft.com/en-us/nuget/create-packages/creating-a-package) | 2026-04 | 2026-08-28 | High |
| [9] | Native asset handling | [Microsoft/NuGet](https://learn.microsoft.com/en-us/nuget/create-packages/select-assemblies-referenced-by-projects) | 2022-04 | 2026-08-28 | High |
| [10] | Process architecture and mismatch failure | [PlatformTarget](https://learn.microsoft.com/en-us/dotnet/csharp/language-reference/compiler-options/output#platformtarget), [BadImageFormatException](https://learn.microsoft.com/en-us/dotnet/api/system.badimageformatexception) | maintained | 2026-08-28 | High |
| [11] | Framework 4.6 retirement | [Microsoft](https://learn.microsoft.com/en-us/lifecycle/products/microsoft-net-framework) | current | 2026-08-28 | High |
| [12] | Go Windows/cgo requirements | [Go project](https://go.dev/wiki/MinimumRequirements) | maintained | 2026-08-28 | High |
| [13] | net46 named-pipe capability | [Microsoft](https://learn.microsoft.com/en-us/dotnet/api/system.io.pipes.namedpipeserverstream) | maintained | 2026-08-28 | High |
| [14] | Secure Windows DLL loading | [Microsoft](https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-search-order) | 2023-02-09 | 2026-08-28 | High |
| [15] | Go release history | [Go project](https://go.dev/doc/devel/release) | maintained | 2026-08-28 | High |

## Staleness map

| Claim class | Re-check by | Status on 2026-08-28 |
|---|---:|---|
| Framework 4.6 lifecycle | 2023-04-26 | Stale by date; rechecked this run and retirement remains current |
| Versions/compatibility | 2026-09-28 | Current |
| Packaging | 2027-08-28 | Current |
| Implementation | 2027-08-28 | Current |
| Architecture | 2028-08-28 | Current |

The earliest forward re-check is **2026-09-28** for compatibility and platform requirements. Use Refresh after toolchain or target changes; use Deepen after the Windows proof of concept.
