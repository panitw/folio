# folio-go

`folio-go` turns a `.folio` template plus report data into a PDF 1.7 document —
byte-identically, every time, on a given build toolchain (see
[Reproducible bytes](#reproducible-bytes--the-toolchain-caveat) below). It
never reads the clock, the filesystem outside the calls you make, the
network, or the host's locale: everything the document needs — data, runtime
parameters, and fonts — is handed in explicitly.

## Your first PDF

Three calls: load a template, assemble the fonts it needs, render.

```go
package main

import (
	"log"
	"os"

	"github.com/panitw/folio/folio-go"
)

func main() {
	// 1. Load a `.folio` template (the document's layout, bands and
	// elements — see _bmad-output/specs/spec-folio/folio-format.md).
	tpl, err := folio.LoadTemplate("statement.folio")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Assemble the fonts the template's text elements need, from
	// your own bytes. This step is ceremony Story 2.2 REMOVES: once
	// folio-go ships a bundled font set (folio-go/fonts/), an
	// integrator with no special typography needs will not have to
	// read a font file at all. Until then, every caller supplies the
	// bytes themselves — the engine never reaches onto the host
	// looking for a font (AD-8).
	fontBytes, err := os.ReadFile("Roboto-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}
	fonts := folio.FontSet{"Roboto-Regular": fontBytes}

	// 3. Render. Report data and runtime parameters are two SEPARATE
	// JSON documents (see "Data vs. Params" below) — here, a small
	// report-data document and a params document supplying one
	// runtime value.
	data := folio.Data(`{"customer": {"name": "Ada Lovelace"}}`)
	params := folio.Params(`{"reportDate": "2026-08-23"}`)

	pdfBytes, err := folio.Render(tpl, data, params, fonts)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("statement.pdf", pdfBytes, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

## Writing straight to a destination: `RenderTo`

`Render` returns a `[]byte`. `RenderTo` writes the same bytes to an
`io.Writer` you already have open — an HTTP response, a file, anything —
instead of handing you a byte slice to copy yourself:

```go
func statementHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/pdf")
	if err := folio.RenderTo(w, tpl, data, params, fonts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

`RenderTo` takes the exact same arguments as `Render`, in the exact same
order, with the writer added at the front. There is no options struct — an
options struct would let the font set be *omitted* at compile time, and a
render missing its fonts is exactly the mistake AD-8 exists to catch at
build time rather than at run time.

**`RenderTo` cannot stream the document out as it is produced, and this is
deliberate, not a missed optimisation.** The PDF's `/ID` is a hash taken over
the entire serialized body, and its cross-reference table records the byte
offset of every object in the finished file — both require the whole
document to exist before the first byte can be written. `RenderTo` builds
the complete document once (through the exact same code `Render` uses, so
the two are never able to drift apart), then performs one write of the
result — never a partial write left unchecked, and never a trailing write
carrying anything extra. If the writer you pass fails partway, returns
fewer bytes accepted than it was given, or otherwise misbehaves, `RenderTo`
reports it as an error rather than looking like it succeeded.

**"Never touches disk" is a capability, not a proof.** `RenderTo` writes
straight to the `io.Writer` you give it, and folio-go's own code never opens
a temporary file to get there — that's what makes serving a PDF from an HTTP
handler with no scratch file possible. What folio-go does not do is *forbid*
a caller from wiring a filesystem-backed `io.Writer` in front of `RenderTo`
themselves; that has always been the caller's own choice to make, the same
way it is for any `io.Writer`-based API. Read this as "you *can* serve a PDF
without touching disk", not as a sandboxed guarantee that nothing in the
process ever will.

## Data vs. Params

`Render` and `RenderTo` take **two** separate JSON inputs:

- **`Data`** — your report's data: the rows, totals and customer details a
  template's placeholders bind against, e.g. `{{customer.name}}`.
- **`Params`** — runtime values that are not part of the report itself, e.g.
  the date the statement was generated: `{{params.reportDate}}`.

They are kept apart on purpose. A placeholder whose first path segment is
`params` — `{{params.reportDate}}` — **always** resolves against the params
document, never against report data, no matter what report data happens to
contain. This means:

- **A top-level `"params"` key in your report data is legal, ordinary JSON,
  and simply unreachable by any binding.** It is your data; folio-go does
  not reject it, and it cannot be used to smuggle a value into
  `{{params.…}}`.
- **`{{params.x}}` with no params document supplied (or one missing that
  key) is an error**, naming the path and the element — the same way an
  absent report-data path is an error. A statement silently missing its
  report date is a worse failure than a render that refuses to produce one.
- Passing `nil` (or an empty `Params`) means "no runtime values supplied" —
  it is not treated as malformed JSON.

Both `Data` and `Params` are decoded through the exact same path, so a
number in either one keeps its author-written precision (no float64
rounding) — there is no separate, second decoder for params that could
someday drift from how report data is read.

`Data` and `Params` are distinct Go types with the same underlying
representation (`[]byte`), and that distinction is enforced by the
compiler, not by convention: swapping them at a call site —
`Render(t, params, data, fonts)` — fails to build. In a library whose
acceptance fixture is a bank statement, that has to be a compile error, not
a support ticket filed after the fact.

## Reproducible bytes — the toolchain caveat

folio-go's whole reason to exist is that the same template and the same
inputs produce the same bytes, every time. `folio-go/go.mod` pins an exact
Go toolchain (`toolchain go1.26.0`) for that reason.

**Where the golden hashes actually live.** folio-go's own regression fixtures
(`folio-go/testdata/`) and the cross-platform byte-identity matrix
(`hashmatrix/`, verified at each epic boundary — see
[What's not here yet](#whats-not-here-yet)) are what pin folio-go's output
against itself. Anyone recording those hashes — or their own — in an
external test suite is the reader the next paragraph is for.

**That pin only binds folio-go's OWN build.** Go's `toolchain` directive
binds the *main* module of whatever `go build` you run — not every module in
the dependency graph. If you `go get` folio-go into your own application,
your application is the main module, and **your own `go.mod`'s `toolchain`
line (or your `GOTOOLCHAIN` setting) decides which Go compiles folio-go**,
not folio-go's own pin. An integrator who records folio-go's golden hashes
in their own test suite, expecting them to hold indefinitely, needs to pin
their *own* toolchain the same way — folio-go's pin alone does not do it for
them.

(If you look at `folio-go/go.mod`, you'll notice the `go` directive sits
*below* the `toolchain` directive, and at a lower version. That is
deliberate, not a stray typo to "fix": Go silently strips a `toolchain` line
that is not strictly greater than the `go` line on `go mod tidy`, which
would quietly undo the pin this section is about.)

## What's not here yet

- **A bundled set of ready-made fonts.** Story 2.2 ships one; until then,
  every caller supplies font bytes directly, as shown above.
- **Page numbering.** `{{page}}` and `{{pages}}` are reserved placeholder
  names — they parse and pass through a template untouched — but nothing
  resolves them yet.
- **A `/CreationDate` or `/ModDate` in the output PDF.** No PDF metadata
  date is emitted by this version of the library at all: rendering with no
  date supplied by any route produces a document with no `/CreationDate`,
  `/ModDate`, or `/Info` dictionary, by design (AD-7). Wiring
  `SOURCE_DATE_EPOCH` through the command-line tool arrives with `cmd/folio`
  in a later story.
- **The cross-platform byte-identity matrix** (Linux amd64/arm64, WASM under
  Node, compared against local `darwin/arm64`) is verified at the Epic 1
  boundary, not on every change — see `hashmatrix/`.
