# Windows portability evidence

Executed from `/Users/panitw/Projects/folio/folio-go` on 2026-08-28 with the repository-pinned Go toolchain:

```sh
go list -deps -json ./... | jq -s '{packages:length,cgo_packages:[.[] | select(((.CgoFiles // [])|length)>0) | .ImportPath]}'
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o "$temporary_directory/folio.exe" ./cmd/folio
file "$temporary_directory/folio.exe"
ls -lh "$temporary_directory/folio.exe"
```

Observed output:

```text
{"packages":128,"cgo_packages":[]}
folio.exe: PE32+ executable (console) x86-64, for MS Windows
folio.exe: approximately 20 MB
```

This proves that the current Folio CLI and its resolved dependency graph compile for Windows amd64 without cgo. It does not prove the proposed c-shared bridge, which requires a Windows-compatible C compiler and is deliberately part of the proof-of-concept gate.
