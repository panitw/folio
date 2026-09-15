# folio8 rendering library for Go

`folio8-go` turns a `.folio` template, JSON data and runtime parameters into a PDF 1.7 document. It
reads no clock, no environment, no network and no host fonts while rendering: everything a document
needs is passed in. The same inputs rendered with the same Go toolchain produce the same bytes.

This guide covers installing the module, rendering your first PDF, the inputs and failures you have to
handle, the template features that change what is drawn, and every exported API of the `folio8`,
`fonts` and `wasm` packages. Two companion references hold the rules this guide does not repeat:

- [The `.folio` format](folio-format.md) — every field of a template, version rules and load errors.
- [Expressions](expression-reference.md) — the syntax inside `{{ }}` and Visibility formulas.

Contents: [Install](#install) · [Your first PDF](#your-first-pdf) ·
[Writing to an `io.Writer`](#writing-to-an-iowriter) · [Inputs](#inputs) ·
[Warnings and errors](#warnings-and-errors) · [Load time versus render time](#load-time-versus-render-time) ·
[Reproducible output](#reproducible-output) · [Known limitations](#known-limitations) ·
[Template features that change the PDF](#template-features-that-change-the-pdf) ·
[API reference](#api-reference) · [Command-line tool](#command-line-tool)

## Install

folio8 has no tagged release yet, so you install a commit of the `main` branch. From your
application's directory:

```sh
go mod init example.com/folio8-demo
go get github.com/panitw/folio8/folio8-go@main
go mod tidy
```

On 2026-09-15 `@main` resolved to `github.com/panitw/folio8/folio8-go v0.0.0-20260914182357-563352e6f92a`
(commit `563352e`), and this guide's programs and example templates were verified against that
version from a fresh module. `go get` records the resolved pseudo-version in your `go.mod`, so your
build stays pinned to that commit until you run `go get github.com/panitw/folio8/folio8-go@main` again
to upgrade. The public API is not frozen before a release is tagged, so read the changes before you
upgrade. If a module proxy still serves an older commit for `@main`, fetch with `GOPROXY=direct`.

Import the module root as `folio8`. The shipped fonts are a separate, opt-in package:

```go
import (
	folio8 "github.com/panitw/folio8/folio8-go"
	"github.com/panitw/folio8/folio8-go/fonts"
)
```

**Toolchain.** `folio8-go/go.mod` declares `go 1.25.0` as the minimum language version and
`toolchain go1.26.0`. A dependency's `toolchain` line does not choose the compiler for your
application: your own `go.mod` `toolchain` line, or your `GOTOOLCHAIN` setting, does. If you record
PDF hashes in your own tests and expect them to hold, pin your own toolchain as well as your inputs.

**Binary size.** `fonts` embeds its faces with `go:embed`, about 14.8 MB of raw font data. Package
`folio8` never imports `fonts`, so the data is in your binary only if you import `fonts` yourself.

## Your first PDF

A template, a data file and a short program. Save these two files next to the program:

`first-pdf.folio`

```json
{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 400, "height": 20, "value": "Hello, {{customer.name}}!", "style": {"fontFamily": "body", "fontSize": 14}}
      ]
    },
    "pageFooter": {
      "elements": [],
      "height": 20
    },
    "pageHeader": {
      "elements": [],
      "height": 20
    }
  },
  "fonts": {
    "body": ["Noto Sans"]
  },
  "locale": "en",
  "nextId": 2,
  "page": {
    "margin": {
      "bottom": 36,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
```

`first-pdf.data.json`

```json
{"customer": {"name": "Ada Lovelace"}}
```

`main.go`

```go
// Command first-pdf renders docs/examples/first-pdf.folio to first-pdf.pdf.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	folio8 "github.com/panitw/folio8/folio8-go"
	"github.com/panitw/folio8/folio8-go/fonts"
)

func main() {
	tpl, err := folio8.LoadTemplate("first-pdf.folio")
	if err != nil {
		log.Fatal(describe("load", err))
	}

	data, err := os.ReadFile("first-pdf.data.json")
	if err != nil {
		log.Fatal(err)
	}

	res, err := folio8.Render(tpl, folio8.Data(data), nil, fonts.Shipped())
	if err != nil {
		log.Fatal(describe("render", err))
	}
	for _, d := range res.Diagnostics {
		// Warnings accompany a successful render. Dispatch on d.Code, never on d.Message.
		fmt.Fprintf(os.Stderr, "%s %s element=%q path=%q: %s\n", d.Severity, d.Code, d.ElementID, d.DataPath, d.Message)
	}

	if err := os.WriteFile("first-pdf.pdf", res.Bytes, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote first-pdf.pdf (%d bytes, %d warnings)\n", len(res.Bytes), len(res.Diagnostics))
}

// describe adds the stable diagnostic code when err carries one.
func describe(stage string, err error) string {
	var re *folio8.RenderError
	if errors.As(err, &re) {
		d := re.Diagnostic
		return fmt.Sprintf("%s failed: %s element=%q path=%q: %v", stage, d.Code, d.ElementID, d.DataPath, err)
	}
	return fmt.Sprintf("%s failed: %v", stage, err)
}
```

`go run .` prints `wrote first-pdf.pdf (53035 bytes, 0 warnings)` with the verified version and
toolchain; the byte count is a property of the inputs and the library version, so treat it as
illustrative. The page reads "Hello, Ada Lovelace!" in Noto Sans.

What each step does:

1. `folio8.LoadTemplate` reads a file and calls `folio8.ParseTemplate`. Use `ParseTemplate` directly
   when the template bytes come from memory, an embed or a database. Both return an opaque
   `*folio8.Template`; you cannot build one field by field, only by parsing.
2. `folio8.Data` is your report data as raw JSON bytes. `nil` params means no runtime values.
3. `fonts.Shipped()` supplies the face named by the template's `fonts` chain (`"Noto Sans"`).
4. `folio8.Render` returns a `folio8.Result`: `Bytes` is the complete PDF whenever the error is nil,
   and `Diagnostics` holds warnings that accompanied that successful render.
5. A failure is an ordinary Go `error`. When it concerns the template, the data or a render rule, it
   is a `*folio8.RenderError` whose `Diagnostic` carries a stable `Code`; `errors.As` finds it. Other
   failures, such as malformed JSON data or a file that cannot be read, are plain errors — keep the
   fallback branch.

`folio8.SerializeTemplate(tpl)` returns the template's canonical `.folio` bytes, which is how an
editor saves a template it loaded. Saving writes keys in canonical order, keeps the declared
`version` unless the content requires a higher one, and never lowers it.

## Writing to an `io.Writer`

`folio8.RenderTo` takes the same arguments as `Render` with a writer first, writes the PDF and
returns the warnings:

```go
diagnostics, err := folio8.RenderTo(writer, tpl, data, params, fontSet)
```

It builds the whole document in memory first — the PDF's cross-reference table and `/ID` depend on
the complete file — and then makes one `Write` call with exactly the bytes `Render` returns. It does
not stream. If the writer returns an error, or accepts fewer bytes than it was given, `RenderTo`
returns an error; the writer may already have accepted part of the document, so discard or remove
partial output. A render failure returns before anything is written.

This program renders the first-PDF template both ways, compares the bytes, shows `Validate`, and
shows a failing writer:

```go
// Command render-to renders the same document twice — once with Render and
// once with RenderTo into a file — and checks the bytes are identical.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"

	folio8 "github.com/panitw/folio8/folio8-go"
	"github.com/panitw/folio8/folio8-go/fonts"
)

func main() {
	templateBytes, err := os.ReadFile("first-pdf.folio")
	if err != nil {
		log.Fatal(err)
	}
	tpl, err := folio8.ParseTemplate(templateBytes)
	if err != nil {
		log.Fatal(err)
	}
	data := folio8.Data(`{"customer": {"name": "Ada Lovelace"}}`)
	params := folio8.Params(`{}`)
	fontSet := fonts.Shipped()

	// Validate runs the same checks as a render, with the same inputs, without
	// producing a PDF.
	warnings, err := folio8.Validate(templateBytes, data, params, fontSet)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("validate: %d warnings\n", len(warnings))

	res, err := folio8.Render(tpl, data, params, fontSet)
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create("render-to.pdf")
	if err != nil {
		log.Fatal(err)
	}
	diagnostics, err := folio8.RenderTo(out, tpl, data, params, fontSet)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		// A failed write may already have put some bytes into the file.
		_ = os.Remove("render-to.pdf")
		log.Fatal(err)
	}
	for _, d := range diagnostics {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", d.Severity, d.Code, d.Message)
	}

	written, err := os.ReadFile("render-to.pdf")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("bytes equal:", bytes.Equal(written, res.Bytes))

	// A writer that fails is reported as an error; the PDF was fully built first.
	_, err = folio8.RenderTo(failingWriter{}, tpl, data, params, fontSet)
	fmt.Println("failing writer:", err)
	var re *folio8.RenderError
	fmt.Println("is a RenderError:", errors.As(err, &re))
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) { return 0, errors.New("disk full") }
```

Verified output:

```text
validate: 0 warnings
bytes equal: true
failing writer: folio8: RenderTo: write failed after 0 of 53035 bytes: disk full
is a RenderError: false
```

A writer failure is a plain error, not a `*folio8.RenderError`: nothing is wrong with the document.

**In an HTTP handler**, a render error returned by `RenderTo` happens before any byte is written, so
you can still send an error status. A write error happens after the response has started — the
status line and part of the body may already be on the wire — so you can only log it. To choose the
status code with certainty, call `Render` first and write `res.Bytes` yourself:

```go
res, err := folio8.Render(tpl, data, params, fontSet)
if err != nil {
	http.Error(w, "could not render statement", http.StatusInternalServerError)
	return
}
w.Header().Set("Content-Type", "application/pdf")
if _, err := w.Write(res.Bytes); err != nil {
	log.Printf("statement write failed: %v", err) // headers are already sent
}
```

**`Validate`** takes template **bytes** (it parses them itself), plus the same data, params and fonts
you will render with. It runs the checks a render runs and returns the error the render would
return, plus the warnings found while preparing the document, without producing a PDF. It is a prediction for those inputs, not a structural lint: with
empty data it correctly reports that the paths your template binds to are absent.

## Inputs

### Data and params

`folio8.Data` and `folio8.Params` are distinct types over raw JSON bytes, so swapping them at a call
site does not compile.

- **Data** is the report: `{{customer.name}}` and `transactions[]` read from it. It must be valid JSON.
- **Params** are runtime values that are not report data, read only through the reserved
  `params.` prefix: `{{params.reportDate}}`. A path starting with `params.` never reads data, so a
  top-level `"params"` key in your data is legal and unreachable. `nil` or empty `Params` means no
  runtime values were supplied, not malformed JSON.
- A path absent from the data — or `{{params.x}}` with no such param — is an error
  (`BINDING_PATH_ABSENT`) naming the element and the path. It is never rendered as empty. A value that
  is present and JSON `null` renders empty.
- Numbers keep the precision written in the JSON. Both inputs are decoded as exact decimals, never as
  `float64`. Build the JSON from your own values (strings, `json.Number`, decimal types) rather than
  decoding into `float64` and re-encoding, which rounds before folio8 ever sees the value.
- `params.documentDate` is reserved: an RFC 3339 timestamp that, when present, is written as the PDF's
  creation and modification date. A present value that is not a valid timestamp fails with
  `DOCUMENT_DATE_INVALID`. Without it the PDF has no date at all.

### Fonts

A render uses only the fonts you pass. `folio8.FontSet` maps a face name to raw OpenType/TrueType bytes:

```go
type FontSet map[string][]byte
```

A template's `fonts` object declares named fallback chains, and each text element names a chain in
`style.fontFamily`. Each character is drawn from the first face in the chain that covers it. A chain
entry is either a face **name**, looked up in your `FontSet`, or an `{"asset": "<key>"}` entry, a
face the template carries in its own `assets` — resolved by asset key only, never by name, so a
`FontSet` entry can never replace an embedded face or the reverse. A chain entry may also declare
`bold`, `italic` and `boldItalic` faces; `style.bold`/`style.italic` pick those, and folio8 never
synthesises a weight or slant. See [`fonts`](folio-format.md#fonts) and
[`assets`](folio-format.md#assets) for the rules.

`fonts.Shipped()` returns a fresh `FontSet` with these face names:

| Family | Face names |
|---|---|
| Noto Sans | `Noto Sans`, `Noto Sans Bold`, `Noto Sans Italic`, `Noto Sans Bold Italic` |
| Noto Sans Thai | `Noto Sans Thai`, `Noto Sans Thai Bold` |
| Noto Sans SC (Simplified Chinese, also used for Japanese) | `Noto Sans SC` |
| Roboto | `Roboto`, `Roboto Bold`, `Roboto Italic`, `Roboto Bold Italic` |

The shipped faces are licensed under the SIL Open Font License 1.1; the licence text and notice ship
next to each face in the `fonts` package directory. Faces you supply yourself, or embed in a template,
are under their own licences: a template that embeds a font redistributes that font with the file.

To use your own faces, build the map yourself, optionally starting from the shipped set:

```go
fontSet := fonts.Shipped()
brand, err := os.ReadFile("BrandSans-Regular.ttf")
if err != nil {
	log.Fatal(err)
}
fontSet["Brand Sans"] = brand // referenced by a chain such as "body": ["Brand Sans", "Noto Sans Thai"]
```

A character that no face in its chain covers is omitted, never drawn as a box, and reported with
the warning `TEXT_MISSING_GLYPH`. A bold or italic request that the covering entry declares no face
for is drawn in that entry's regular face with the warning `TEXT_STYLE_FACE_UNDECLARED`. A chain face
missing from the `FontSet`, or bytes that are not a usable font, fail the render with an error naming the element and the chain;
that error is not a `*folio8.RenderError`, so handle it in your fallback branch.

## Warnings and errors

folio8 reports every problem as a `folio8.Diagnostic`:

| Field | Meaning |
|---|---|
| `Severity` | `folio8.SeverityWarning` or `folio8.SeverityError`. `Severity.String()` returns `"Warning"` or `"Error"`; the zero value prints `"Severity(unset)"` and is never produced by folio8. |
| `Code` | A stable string from a closed registry, such as `"TEXT_CLIPPED_WIDTH"`. Compare against the `folio8.DiagCode…` constants. A code's meaning never changes. |
| `ElementID` | The template element (`"e7"`) or table column the problem concerns, when there is one. |
| `DataPath` | The data path, or for load errors the template field path (`"pages[1].sectionBreak"`, `"table.width"`), when there is one. |
| `Message` | A human-readable sentence. Print it; never parse it. |

**Warnings accompany a successful render.** The PDF in `Result.Bytes` (or written by `RenderTo`) is
complete, and something could not be honoured exactly — text clipped at its box, a missing glyph, a
barcode that could not be drawn. `Result.Diagnostics` lists them in document order, and is `nil`
when there are none. `folio8 render -strict` turns warnings into a failure; in Go, decide yourself.

**Errors abort.** `LoadTemplate`, `ParseTemplate`, `Render`, `RenderTo` and `Validate` return a Go
`error`, and no PDF. When the error is a known condition it is a `*folio8.RenderError` whose
`Diagnostic` has `Severity` `SeverityError` and a code; `Unwrap` exposes the underlying error.

```go
res, err := folio8.Render(tpl, data, params, fontSet)
var re *folio8.RenderError
switch {
case errors.As(err, &re):
	switch re.Diagnostic.Code {
	case folio8.DiagCodeBindingPathAbsent:
		return fmt.Errorf("data is missing %s (element %s)", re.Diagnostic.DataPath, re.Diagnostic.ElementID)
	default:
		return fmt.Errorf("%s: %w", re.Diagnostic.Code, err)
	}
case err != nil:
	return err // not a document problem: invalid JSON data, a nil template, an I/O failure
}
for _, d := range res.Diagnostics {
	if d.Code == folio8.DiagCodeTextMissingGlyph {
		log.Printf("element %s lost a character: %s", d.ElementID, d.Message)
	}
}
```

Every code, with its value and meaning, is listed in [Diagnostic codes](#diagnostic-codes).

## Load time versus render time

Some problems are decidable from the template alone, and loading refuses them: `ParseTemplate` and
`LoadTemplate` fail and there is no `*Template` to render. Others depend on data, params or fonts
and surface only when rendering.

| Refused at load (`ParseTemplate`/`LoadTemplate`) | Reported at render (`Render`/`RenderTo`/`Validate`) |
|---|---|
| Malformed JSON or a MAJOR version above 4 — `TEMPLATE_MALFORMED` | An absent data or params path — `BINDING_PATH_ABSENT` (error) |
| A field value outside its rules: unknown closed-set member, missing required field, duplicate id, bad table width allocation — `TEMPLATE_FIELD_INVALID`, located by `DataPath` | An expression that fails on this data (wrong kind, division by zero) — `EXPRESSION_INVALID` (error) |
| An invalid expression (syntax, unknown function, provably wrong kind) — `EXPRESSION_INVALID` | A line or image taller than the content window — `CONTENT_UNLAYOUTABLE` (error) |
| A table `minHeight` taller than the content window — `TABLE_MIN_HEIGHT_UNPLACEABLE` | A malformed colour — `STYLE_COLOR_INVALID` (error) |
| A section break out of range, duplicated, on a header/footer band, or `sectionBreakAnchor` without `sectionBreak` or not a boolean — `SECTION_BREAK_INVALID`; an element across the break — `SECTION_BREAK_STRADDLED` | Clipped text, missing glyphs, undeclared faces, suppressed table headers or footers, clipped rows — warnings |
| An invalid `pages` array — `PAGES_INVALID` | Barcode/QR content that cannot be encoded or fitted — warnings |
| A `lineSpacing` outside its range — `STYLE_LINE_SPACING_INVALID` | An `avg` over an empty collection — `AGGREGATE_EMPTY_AVERAGE` (warning) |

`PAGES_INVALID` and `SECTION_BREAK_INVALID` locate the problem through `Diagnostic.DataPath` —
`pages`, `pages[1]`, `pages[1].sectionBreak`, `bands.content.sectionBreakAnchor` — not through
`ElementID`, because a band or a page is not an element. `SECTION_BREAK_STRADDLED` names the
element in `ElementID`. Table width allocation problems are `TEMPLATE_FIELD_INVALID` with `DataPath`
`table.width` or `column.<field>`.

## Reproducible output

- Output depends only on the template, data, params, `FontSet` and the Go toolchain that built your
  binary. Rendering never reads the clock, environment variables, the host locale, host fonts or the
  network.
- Locale and time are document properties: the template's `locale` (`en`, `th`, `zh-Hans` or `ja`)
  and fixed `utcOffset` drive `formatDate`, `formatNumber` and line breaking. Any other locale is a
  load error.
- No PDF date is written unless you pass `params.documentDate`. The `folio8` command-line tool fills
  that param from `SOURCE_DATE_EPOCH` when you have not supplied one; the library never reads that
  variable.
- Nothing in this guide is a statement about concurrent use; the library makes no documented
  concurrency guarantee for sharing one `*Template` or `FontSet` across goroutines while it is being
  mutated. Treat `FontSet` maps you pass as read-only for the duration of a call.

## Known limitations

- **Japanese glyph forms.** `ja` text renders with coverage from `Noto Sans SC`, which draws the
  Simplified Chinese shape for the ideographs whose Japanese form differs. Supply a Japanese face in
  your `FontSet` and chain if that matters.
- **Thai names.** Thai has no spaces between words and breaks from a dictionary, which cannot tell a
  name from the ordinary words it is made of. List data paths whose values must never break in the
  document's `unbreakableValues`; a name inside free-form text is still breakable. See
  [Line breaking](folio-format.md#line-breaking).
- **Latin breaking** is at spaces only: no hyphenation, no break after `-`, and it is not UAX #14.
- **CJK kinsoku** is not implemented: a line may begin with `，` or end with an opening bracket.
- Expression syntax, limits and division scale: see [Expressions](expression-reference.md).

## Template features that change the PDF

Each feature below has a minimal template. The docs tests render every one of them on every change
and assert the key outcomes — page counts, where the moved content lands, and diagnostic codes. A template must declare at least the version its features
require — saving through `SerializeTemplate` raises `version` for you. Versions up to `4.1` load; a
higher MINOR also loads, a higher MAJOR is refused. The examples use a small 200 × 150 pt page with
10 pt margins and 10 pt header and footer bands, so the content window is 110 pt tall.

| Feature | Requires | Diagnostics |
|---|---|---|
| Formulas: comparisons, arithmetic, ternaries, `true`/`false`/`null` | `2.0`; `3.3` when a text expression can return a number | `EXPRESSION_INVALID`, `BINDING_PATH_ABSENT`, `AGGREGATE_EMPTY_AVERAGE` |
| Proportional table widths | `3.0` | `TEMPLATE_FIELD_INVALID` at `table.width` or `column.<field>` |
| Table frame, `rules`, `minHeight` | `3.1` | `TABLE_MIN_HEIGHT_UNPLACEABLE` |
| Column `headerAlign` | `3.2` | — |
| `barcode` and `qrcode` elements | `4.0` | `BARCODE_UNENCODABLE`, `BARCODE_DOES_NOT_FIT`, `BARCODE_MODULE_TOO_SMALL`, `QRCODE_TOO_LONG`, `QRCODE_DOES_NOT_FIT`, `QRCODE_MODULE_TOO_SMALL` |
| Section break, anchored and unanchored | `4.1` | `SECTION_BREAK_INVALID`, `SECTION_BREAK_STRADDLED`, `SECTION_BREAK_SPLITS_KEEP_TOGETHER` |
| Designed pages, `pageBreak` true and false | `4.1` | `PAGES_INVALID` |

### Formulas

Visibility (`visibleIf`) takes a bare formula; text takes `{{ }}` expressions. Comparisons, exact
decimal arithmetic, nested `? :` and the literals `true`, `false` and `null` work in both, and in
`if()`. A number printed bare in text appears as its exact decimal (`{{1 / 3}}` prints `0.3333`); use
`formatNumber` for grouping. A hidden element leaves its siblings where they are. Syntax, precedence,
limits and division scale are in [Expressions](expression-reference.md#formulas-and-visibility); field rules in
[Expressions](folio-format.md#expressions).

`formula-visibility.folio`

```json
{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 180, "height": 12, "value": "Total {{formatNumber(loanAmount + fee, \"#,##0.00\")}}", "style": {"fontFamily": "body", "fontSize": 8}},
        {"id": "e2", "type": "text", "x": 0, "y": 16, "width": 180, "height": 12, "value": "Manual review required", "visibleIf": "loanAmount > 20000", "style": {"fontFamily": "body", "fontSize": 8}},
        {"id": "e3", "type": "text", "x": 0, "y": 32, "width": 180, "height": 12, "value": "{{loanAmount > 20000 ? \"Tier: large\" : \"Tier: standard\"}}", "style": {"fontFamily": "body", "fontSize": 8}}
      ]
    },
    "pageFooter": {"elements": [], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "2.0"
}
```

`formula-visibility.above.json`

```json
{"loanAmount": 25000, "fee": 150.50}
```

`formula-visibility.below.json`

```json
{"loanAmount": 20000, "fee": 150.50}
```

With the first data the page reads "Total 25,150.50", "Manual review required" and "Tier: large".
With the second it reads "Total 20,150.50" and "Tier: standard"; the review line is not drawn and the
tier line stays at y 32. There are no warnings. A missing `loanAmount` fails with
`BINDING_PATH_ABSENT`; `"visibleIf": "loanAmount"` with a number is a located error, because
conditions have no truthiness.

### Barcode and QR code

A `barcode` element encodes its `value` as Code 128; a `qrcode` element encodes it as a QR Code at
error correction `L`, `M` (the default), `Q` or `H`. `value` binds like text. Both draw black vector
modules of one whole-millipoint width, as large as fits the box with the quiet zone inside it, centred
and never distorted, with no human-readable text and no `style`. Store control characters as real
characters (`"\r"` in JSON). See [Elements](folio-format.md#elements).

`barcode-qrcode.folio`

```json
{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "barcode", "x": 0, "y": 0, "width": 260, "height": 50, "value": "INV-{{invoice}}"},
        {"id": "e2", "type": "qrcode", "x": 0, "y": 60, "width": 100, "height": 100, "value": "https://example.com/pay/{{invoice}}"},
        {"id": "e3", "type": "qrcode", "x": 120, "y": 60, "width": 100, "height": 100, "errorCorrection": "H", "value": "{{invoice}}"}
      ]
    },
    "pageFooter": {"elements": [], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 300, "height": 220}},
  "utcOffset": "+00:00",
  "version": "4.0"
}
```

`barcode-qrcode.data.json`

```json
{"invoice": "2026-000123"}
```

`barcode-qrcode.unencodable.json`

```json
{"invoice": "2026-ใบที่-7"}
```

With the first data the page carries three codes, which an independent decoder reads as Code 128
`INV-2026-000123`, QR (level M) `https://example.com/pay/2026-000123` and QR (level H) `2026-000123`,
with no warnings. With the second data the barcode is omitted with the warning
`BARCODE_UNENCODABLE` on `e1` — Code 128 encodes ASCII only — while both QR codes still draw, because
QR encodes the UTF-8 bytes. The render succeeds. Static text that cannot be encoded (a non-ASCII
character outside `{{ }}` in a barcode, or a QR value already too long) is refused at load instead. A
symbol that cannot fit at one millipoint per module is omitted with `BARCODE_DOES_NOT_FIT` or
`QRCODE_DOES_NOT_FIT`; modules narrower than 0.25 mm (barcode) or 0.5 mm (QR) still draw, with
`BARCODE_MODULE_TOO_SMALL` or `QRCODE_MODULE_TOO_SMALL`; QR data too long for version 40 is omitted
with `QRCODE_TOO_LONG`.

### Section break

`sectionBreak` on the content band is an offset in points. Elements declared at or below it form a
section that moves as one rigid block when the content above grows past the line. Anchored (the
default) moves the section by whole pages, keeping its declared position; with
`"sectionBreakAnchor": false` the section follows where the content above ends when it fits on that
page. When nothing crosses the line, the PDF is byte-identical to the same template without the key.
The break is never drawn, and a `keepTogether` group split by it is split with the warning
`SECTION_BREAK_SPLITS_KEEP_TOGETHER`. See [Pagination](folio-format.md#pagination) and
[`bands`](folio-format.md#bands).

`section-break.folio` (anchored)

```json
{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
          "style": {"fontFamily": "body", "fontSize": 8},
          "columns": [
            {"id": "e2", "label": "Item", "width": 100, "bind": "{{row.name}}"},
            {"id": "e3", "label": "Amount", "width": 80, "align": "right", "bind": "{{row.amount}}"}
          ]},
        {"id": "e5", "type": "text", "x": 0, "y": 80, "width": 180, "height": 12, "value": "Legend", "style": {"fontFamily": "body", "fontSize": 8}}
      ],
      "sectionBreak": 75
    },
    "pageFooter": {"elements": [{"id": "e4", "type": "text", "x": 0, "y": 0, "width": 180, "height": 8, "value": "Page {{page}} of {{pages}}", "style": {"fontFamily": "body", "fontSize": 6}}], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 6,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "4.1"
}
```

`section-break-unanchored.folio` differs only in the content band's keys:

```json
{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
          "style": {"fontFamily": "body", "fontSize": 8},
          "columns": [
            {"id": "e2", "label": "Item", "width": 100, "bind": "{{row.name}}"},
            {"id": "e3", "label": "Amount", "width": 80, "align": "right", "bind": "{{row.amount}}"}
          ]},
        {"id": "e5", "type": "text", "x": 0, "y": 80, "width": 180, "height": 12, "value": "Legend", "style": {"fontFamily": "body", "fontSize": 8}}
      ],
      "sectionBreak": 75, "sectionBreakAnchor": false
    },
    "pageFooter": {"elements": [{"id": "e4", "type": "text", "x": 0, "y": 0, "width": 180, "height": 8, "value": "Page {{page}} of {{pages}}", "style": {"fontFamily": "body", "fontSize": 6}}], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 6,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "4.1"
}
```

`section-break.5-rows.json`

```json
{"items": [{"name": "Item 1", "amount": "10.00"}, {"name": "Item 2", "amount": "20.00"}, {"name": "Item 3", "amount": "30.00"}, {"name": "Item 4", "amount": "40.00"}, {"name": "Item 5", "amount": "50.00"}]}
```

`section-break.7-rows.json`

```json
{"items": [{"name": "Item 1", "amount": "10.00"}, {"name": "Item 2", "amount": "20.00"}, {"name": "Item 3", "amount": "30.00"}, {"name": "Item 4", "amount": "40.00"}, {"name": "Item 5", "amount": "50.00"}, {"name": "Item 6", "amount": "60.00"}, {"name": "Item 7", "amount": "70.00"}]}
```

The header row is 10 pt and each row about 10.9 pt, so five rows end at 64.5 pt and seven at 86.3 pt.

| Template | Data | Result |
|---|---|---|
| either | 5 rows | One page; the legend at its declared y 80. The PDF is identical to the template without `sectionBreak`. |
| anchored | 7 rows | Two pages. Page 1 holds the table; the legend moves to page 2 at its declared y 80. Footers read `Page 1 of 2` and `Page 2 of 2`. |
| unanchored | 7 rows | One page. The legend moves down 11.272 pt — the distance the rows ran past the line — and sits directly under the table. |

No warnings in any case. A `sectionBreak` of 0 or at least the content height, a second break, a
break on `pageHeader`/`pageFooter`, or `sectionBreakAnchor` without `sectionBreak` fails at load
with `SECTION_BREAK_INVALID` located at the band's key; an element whose box lies on both sides of
the line fails with `SECTION_BREAK_STRADDLED` naming it.

### Designed pages

A template with a top-level `pages` array has two or more designed pages, rendered in order. They
share the page setup, page header and page footer, and `{{page}}`/`{{pages}}` count the output pages
of all of them. Each page's content overflows onto further output pages as usual, and each page may
have its own section break. With `bands.content` then empty, a page's `pageBreak` decides how it
follows the previous page: `true` (Page Break on, the default after page 1) starts a new output page
after the previous page and all its overflow; `false` continues directly where the previous page's
content ended, as one rigid block, if the previous page overflowed onto more than one output page
and the block fits in the room left — otherwise it starts a new output page. See
[Designed pages](folio-format.md#designed-pages).

`designed-pages.folio` (Page Break on)

```json
{
  "assets": {},
  "bands": {
    "content": {"elements": []},
    "pageFooter": {"elements": [{"id": "e4", "type": "text", "x": 0, "y": 0, "width": 180, "height": 8, "value": "Page {{page}} of {{pages}}", "style": {"fontFamily": "body", "fontSize": 6}}], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 6,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "pages": [
    {"elements": [
      {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
        "style": {"fontFamily": "body", "fontSize": 8},
        "columns": [
          {"id": "e2", "label": "Item", "width": 100, "bind": "{{row.name}}"},
          {"id": "e3", "label": "Amount", "width": 80, "align": "right", "bind": "{{row.amount}}"}
        ]}
    ]},
    {"elements": [
      {"id": "e5", "type": "text", "x": 0, "y": 0, "width": 180, "height": 12, "value": "Approved by", "style": {"fontFamily": "body", "fontSize": 8}}
    ], "pageBreak": true}
  ],
  "utcOffset": "+00:00",
  "version": "4.1"
}
```

`designed-pages-page-break-off.folio`

```json
{
  "assets": {},
  "bands": {
    "content": {"elements": []},
    "pageFooter": {"elements": [{"id": "e4", "type": "text", "x": 0, "y": 0, "width": 180, "height": 8, "value": "Page {{page}} of {{pages}}", "style": {"fontFamily": "body", "fontSize": 6}}], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 6,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "pages": [
    {"elements": [
      {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
        "style": {"fontFamily": "body", "fontSize": 8},
        "columns": [
          {"id": "e2", "label": "Item", "width": 100, "bind": "{{row.name}}"},
          {"id": "e3", "label": "Amount", "width": 80, "align": "right", "bind": "{{row.amount}}"}
        ]}
    ]},
    {"elements": [
      {"id": "e5", "type": "text", "x": 0, "y": 0, "width": 180, "height": 12, "value": "Approved by", "style": {"fontFamily": "body", "fontSize": 8}}
    ], "pageBreak": false}
  ],
  "utcOffset": "+00:00",
  "version": "4.1"
}
```

`designed-pages.data.json`

```json
{"items": [{"name": "Item 1", "amount": "10.00"}, {"name": "Item 2", "amount": "20.00"}, {"name": "Item 3", "amount": "30.00"}, {"name": "Item 4", "amount": "40.00"}, {"name": "Item 5", "amount": "50.00"}, {"name": "Item 6", "amount": "60.00"}, {"name": "Item 7", "amount": "70.00"}, {"name": "Item 8", "amount": "80.00"}, {"name": "Item 9", "amount": "90.00"}, {"name": "Item 10", "amount": "100.00"}, {"name": "Item 11", "amount": "110.00"}, {"name": "Item 12", "amount": "120.00"}]}
```

Designed page 1's twelve rows need two output pages: rows 1–9 on the first, a repeated header and
rows 10–12 on the second.

| Page Break | Result |
|---|---|
| on | Three output pages. "Approved by" is alone on output page 3 at its declared position. Footers read `Page 1 of 3` to `Page 3 of 3`. |
| off | Two output pages. "Approved by" is drawn on output page 2 directly under row 12. Footers read `Page 1 of 2` and `Page 2 of 2`. |

No warnings. A `pages` array with fewer than two entries, elements or a section break left in
`bands.content`, an unknown key on a page, a non-boolean `pageBreak`, or a `keepTogether` group on two
pages fails at load with `PAGES_INVALID`, located by `DataPath`.

### Table frame, rules and minHeight

A table's `style.border` and `style.background` draw one frame around each page's slice of the table,
not a border on every cell. `rules` draws interior lines between `columns` and/or `rows`, never on
the frame's edge. `minHeight` is a floor for each page's slice: the frame and column rules extend to
it, rows are never stretched, and following content starts below it. Column labels may wrap onto
several lines, and `headerHeight` is then the header's minimum height. See [`table`](folio-format.md#table).

`ruled-table.folio`

```json
{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
          "minHeight": 90,
          "rules": {"between": ["columns", "rows"], "width": 0.5, "color": "#808080"},
          "style": {"fontFamily": "body", "fontSize": 8, "border": {"width": 1, "color": "#000000"}},
          "columns": [
            {"id": "e2", "label": "Item", "width": 100, "bind": "{{row.name}}"},
            {"id": "e3", "label": "Amount", "width": 80, "align": "right", "bind": "{{row.amount}}"}
          ]}
      ]
    },
    "pageFooter": {"elements": [], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "3.1"
}
```

`ruled-table.data.json`

```json
{"items": [{"name": "Item 1", "amount": "10.00"}, {"name": "Item 2", "amount": "20.00"}, {"name": "Item 3", "amount": "30.00"}]}
```

The page shows a 180 × 90 pt frame although three rows fill only about 43 pt, one vertical rule
between the two columns running the frame's full height, and three horizontal rules: under the
header and between the rows, none under the last row. No warnings.

`ruled-table-unplaceable.folio` asks for a floor taller than the 110 pt content window:

```json
{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "table", "x": 0, "y": 0, "bind": "items[]", "headerHeight": 10,
          "minHeight": 200,
          "rules": {"between": ["columns", "rows"], "width": 0.5, "color": "#808080"},
          "style": {"fontFamily": "body", "fontSize": 8, "border": {"width": 1, "color": "#000000"}},
          "columns": [
            {"id": "e2", "label": "Item", "width": 100, "bind": "{{row.name}}"},
            {"id": "e3", "label": "Amount", "width": 80, "align": "right", "bind": "{{row.amount}}"}
          ]}
      ]
    },
    "pageFooter": {"elements": [], "height": 10},
    "pageHeader": {"elements": [], "height": 10}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 10, "left": 10, "right": 10, "top": 10}, "orientation": "portrait", "size": {"width": 200, "height": 150}},
  "utcOffset": "+00:00",
  "version": "3.1"
}
```

`ParseTemplate` refuses it with a `*folio8.RenderError` whose code is `TABLE_MIN_HEIGHT_UNPLACEABLE`
and `ElementID` is `e1`; there is no template to render. The same page with data or fonts cannot
change that outcome.

The frame meaning of `style.border` applies to every table, including templates written before
`3.1`: a template that relied on `style.border` drawing a grid now draws only the frame. Declare
`"rules": {"between": ["columns", "rows"]}` to draw the interior lines.

### Other version-gated table features

- **Proportional widths** (`3.0`): declare the table's total `width` and a `proportion` on each column
  instead of column widths. folio8 allocates exact millipoint widths that sum to the total. A zero-width
  allocation or a mix of both representations fails at load with `TEMPLATE_FIELD_INVALID`, `DataPath`
  `table.width` or `column.<field>`.
- **Column `headerAlign`** (`3.2`): aligns one column's header cell independently of its data.

## API reference

This section lists every exported identifier in the three packages of the `github.com/panitw/folio8/folio8-go` module:

| Import path | Package | Role |
|---|---|---|
| `github.com/panitw/folio8/folio8-go` | `folio8` | Parsing, rendering, validation, diagnostics, template helpers, and the canvas/authoring helpers used by folio8 Designer. |
| `github.com/panitw/folio8/folio8-go/fonts` | `fonts` | The shipped font faces, as a ready-made `folio8.FontSet`. Opt-in: package `folio8` never imports it. |
| `github.com/panitw/folio8/folio8-go/wasm` | `wasm` | The browser/editor session engine compiled into folio8 Designer's WebAssembly worker. |

**Stability.** The module has no release tag yet: `folio8.Version` is `"0.0.0-dev"`. The rendering and validation entry points, the font input and the diagnostic types are the library surface a server-side application uses. The *authoring and canvas* APIs and the whole `wasm` package are **editor-facing helpers**. Their JSON projections and command payloads track folio8 Designer, which lives in the same repository and changes in the same commit. They are not a frozen API; expect fields and command kinds to change without notice.

**Units.** Every `int64` length in a projection or command result is **millipoints**: 1/1000 of a PDF point, so 1 pt = 1000 and A4 is 595276 × 841890. Length values *inside* a command payload are written in **points** as JSON decimals with at most three decimal places and no exponent. The engine converts them to millipoints exactly, never through `float64`; for example `12.5` becomes 12500. Line spacing is a dimensionless ratio carried in **thousandths**, so 1000 means 1.0.

**Concurrency.** Nothing in these packages documents or tests concurrent use. `*folio8.Template` is mutated in place by the authoring commands, and `*wasm.Engine` holds unsynchronised session state. Do not share either across goroutines without your own locking.

### Rendering and validation

#### `ParseTemplate`

```go
func ParseTemplate(b []byte) (*Template, error)
```

Parses `b` as a `.folio` document and returns an opaque `*Template`. Beyond decoding, it performs every check that can be decided from the document alone:

- It parses and statically checks every `{{ }}` expression: syntax, arity, unknown function names and literal argument kinds. It does not evaluate them.
- It derives `footerOf`/`footerFormat` for `sum`/`avg` table footers that omit `footerOf`.
- It refuses a table `minHeight` taller than the content window.
- It refuses an invalid section break or an element straddling one.

Errors are `*RenderError` values carrying a load code:

- `DiagCodeTemplateMalformed` for bytes that are not a loadable document at all.
- Otherwise the code the loader attached, by default `DiagCodeTemplateFieldInvalid`, or a specific code such as `DiagCodeStyleLineSpacingInvalid`, `DiagCodeTableFooterSourceForbidden`, `DiagCodeTableFooterSourceUnresolved`, `DiagCodeExpressionInvalid`, `DiagCodeTableMinHeightUnplaceable`, `DiagCodeSectionBreakInvalid`, `DiagCodeSectionBreakStraddled` or `DiagCodePagesInvalid`.

For `DiagCodeSectionBreakInvalid` and `DiagCodePagesInvalid`, `Diagnostic.DataPath` carries the band or page field path, such as `pages` or `pages[1]`, because there is no element id.

`b` is not retained after the call.

#### `LoadTemplate`

```go
func LoadTemplate(path string) (*Template, error)
```

Reads `path` with `os.ReadFile` and delegates to `ParseTemplate`. A file-system error is returned unwrapped, not as a `*RenderError`. Parse failures are exactly those of `ParseTemplate`.

#### `Template`

```go
type Template struct {
	// Has unexported fields.
}
```

A parsed, canonicalised `.folio` document. It is opaque: there are no exported fields or accessors, and a composite literal cannot construct a usable one. Obtain it only from `ParseTemplate` or `LoadTemplate`. Read-only functions such as `Render`, `Canvas` and `TableColumns` do not modify it. `ApplyComponentCommand` and `ApplyPageSetupCommand` **do** modify it in place.

#### `SerializeTemplate`

```go
func SerializeTemplate(t *Template) ([]byte, error)
```

Returns the engine's canonical `.folio` bytes for `t`. This is the save path the designer uses. A nil `t` returns an error. The written format version is raised when the document's expressions need it:

- `"2.0"` for formula syntax or boolean/null literals.
- `"3.3"` for a text expression whose static kind includes a number.

The returned slice is newly allocated.

#### `Render`

```go
func Render(t *Template, d Data, p Params, f FontSet) (Result, error)
```

Produces a PDF 1.7 document. It resolves every placeholder against `d` (report data) and `p` (runtime parameters) and embeds a subset of every font face used from `f`.

- **Preconditions.**
  - `t` must be non-nil, or a plain error is returned.
  - `d` must be syntactically valid JSON. It is decoded once with number literals preserved exactly, and a decode failure is a plain error prefixed `folio8: Render:`.
  - A nil or empty `p` means "no runtime values": `{{params.x}}` is then absent.
  - `f` must contain every face the document's font chains actually need. The engine never looks for fonts on the host.
- **Returns.** On success, `Result.Bytes` is the complete PDF and `Result.Diagnostics` holds every Warning. A non-nil error means nothing was rendered; ignore `Result` in that case.
- **Errors.** Located failures are `*RenderError` with a `SeverityError` diagnostic: for example `DiagCodeBindingPathAbsent`, `DiagCodeExpressionInvalid`, `DiagCodeContentUnlayoutable`, `DiagCodeStyleColorInvalid` or `DiagCodeDocumentDateInvalid`. Missing fonts and invalid input JSON may arrive as ordinary errors, so always keep a non-`RenderError` fallback.
- **Ownership.** `t`, `d`, `p` and `f` are not modified.

#### `RenderTo`

```go
func RenderTo(w io.Writer, t *Template, d Data, p Params, f FontSet) ([]Diagnostic, error)
```

Renders exactly as `Render` does, then writes the finished bytes to `w` in **one** `Write` call. RenderTo cannot stream: the PDF `/ID` and cross-reference offsets require the whole document first. It returns the Warning diagnostics.

Errors:

- `w == nil` returns an error without rendering.
- Any `Render` error is returned unchanged.
- A `Write` error is wrapped, with the message stating how many of the total bytes were accepted.
- A short write (`n < len` with a nil error) is reported as an error.

After a write failure `w` may already hold a partial document.

#### `Validate`

```go
func Validate(b []byte, d Data, p Params, f FontSet) ([]Diagnostic, error)
```

A dry-run predictor of `Render` over **template bytes**. It parses `b` (the same errors as `ParseTemplate`), decodes `d` and `p`, checks the reserved `documentDate` parameter, then runs the same preparation `Render` uses without composing pages or producing PDF bytes.

It returns the Warning diagnostics that preparation produced, or the first located error `Render` would report for the **same inputs**. Pass the data you will actually render with. An empty `Data` produces `BINDING_PATH_ABSENT` errors, which are correct predictions for a render with empty data, not template defects. On error the returned slice is nil.

#### `Data`

```go
type Data []byte
```

The report data as raw JSON bytes. It is a distinct defined type, so swapping `Data` and `Params` in a call is a compile error. Pass bytes, not a decoded Go value; the library owns the decode so that decimal precision survives. A top-level `"params"` key inside `Data` is legal but unreachable by bindings.

#### `Params`

```go
type Params []byte
```

Runtime values (for example a statement date) as raw JSON bytes. They are reachable only as `{{params.…}}` and decoded the same way as `Data`. Nil or empty means no runtime values; it is not a decode error.

#### `Result`

```go
type Result struct {
	Bytes       []byte
	Diagnostics []Diagnostic
}
```

| Field | Type | Meaning |
|---|---|---|
| `Bytes` | `[]byte` | The complete PDF. Meaningful only when `Render` returned a nil error. |
| `Diagnostics` | `[]Diagnostic` | Every Warning from the render, in document order: page header, then content, then page footer, and within a band in element declaration order. It is `nil`, never an empty non-nil slice, when there are none. It is never needed to decide whether `Bytes` is valid. |

### Fonts

#### `FontSet`

```go
type FontSet map[string][]byte
```

The engine's only font input. It maps a **face name**, as written in a document's `fonts` fallback chains (for example `"Noto Sans"`), to that face's raw OpenType/TrueType bytes. Rendering resolves faces only from this map and never queries the host system. Faces a document embeds in its own `assets` are resolved from the document. The map and its byte slices are read, not modified.

#### `fonts.Shipped`

```go
func Shipped() folio8.FontSet
```

Returns a new `folio8.FontSet` map holding the eleven faces embedded in the `fonts` package. Each face ships with its OFL-1.1 licence text and NOTICE under `folio8-go/fonts/`. The keys are exactly:

| Family | Face names (map keys) |
|---|---|
| Noto Sans | `Noto Sans`, `Noto Sans Bold`, `Noto Sans Italic`, `Noto Sans Bold Italic` |
| Noto Sans Thai | `Noto Sans Thai`, `Noto Sans Thai Bold` (no italic cuts exist upstream) |
| Noto Sans SC | `Noto Sans SC` (Regular only) |
| Roboto | `Roboto`, `Roboto Bold`, `Roboto Italic`, `Roboto Bold Italic` |

A chain entry must name one of these keys verbatim. Weight and slope come from the variants a chain entry *declares* (for example an entry object with `bold: "Roboto Bold"`); the engine never derives `"Roboto Bold"` from `"Roboto"`.

Each call builds a fresh map, but the byte slices are shared package data: never modify them. Importing `fonts` adds roughly 14.8 MB of raw font bytes to a binary.

### Diagnostics and errors

#### `Diagnostic`

```go
type Diagnostic struct {
	Severity  Severity
	Code      string
	ElementID string
	DataPath  string
	Message   string
}
```

| Field | Type | Meaning |
|---|---|---|
| `Severity` | `Severity` | `SeverityWarning` on `Result.Diagnostics` or `RenderTo`/`Validate` returns; `SeverityError` inside a `*RenderError`. |
| `Code` | `string` | A stable code from the closed registry below. Dispatch on this, never on `Message`. |
| `ElementID` | `string` | The template element concerned, when there is one. |
| `DataPath` | `string` | The data path concerned (for example the absent binding path), or a document field path for band/page load errors (`pages[1]`, `bands.content`, `table.width`, `column.<field>`). Empty when not applicable. |
| `Message` | `string` | A human-readable sentence, safe to display. Not a parsing target. |

`Diagnostic` has no JSON struct tags, so encoding it directly produces the Go field names and an integer `Severity`.

#### `Severity`, `SeverityWarning`, `SeverityError`, `Severity.String`

```go
type Severity int

const (
	SeverityWarning Severity // = 1
	SeverityError            // = 2
)

func (s Severity) String() string
```

A Warning accompanies a successful render; an Error aborts it and travels as Go's error return. The zero value is deliberately **not** a valid severity: it is an unexported "unset" constant, so a `Diagnostic{}` that forgot its severity is not mistaken for a Warning. Compare against the named constants, not integers.

`String` returns `"Warning"`, `"Error"`, `"Severity(unset)"` for the zero value, and `"Severity(N)"` for any other integer.

#### `RenderError`, `RenderError.Error`, `RenderError.Unwrap`

```go
type RenderError struct {
	Diagnostic Diagnostic
	Err        error
}

func (e *RenderError) Error() string
func (e *RenderError) Unwrap() error
```

| Field | Type | Meaning |
|---|---|---|
| `Diagnostic` | `Diagnostic` | `Severity` is `SeverityError`. `Code`, `ElementID` and `DataPath` locate the failure; `Message` equals `Err.Error()`. |
| `Err` | `error` | The underlying error. |

`Error()` returns `Err.Error()` unchanged. `Unwrap()` returns `Err`, so `errors.Is`/`errors.As` reach the wrapped error. Match on the code:

```go
res, err := folio8.Render(tpl, data, params, fontSet)
if err != nil {
	var re *folio8.RenderError
	if errors.As(err, &re) {
		switch re.Diagnostic.Code {
		case folio8.DiagCodeBindingPathAbsent:
			log.Printf("missing data %s (element %s)", re.Diagnostic.DataPath, re.Diagnostic.ElementID)
		default:
			log.Printf("%s: %s", re.Diagnostic.Code, re.Diagnostic.Message)
		}
	} else {
		log.Printf("render failed: %v", err)
	}
	return
}
_ = res
```

#### `ComponentCommandError`

```go
type ComponentCommandError struct {
	ElementID string
	DataPath  string
	Message   string
	// Has unexported fields.
}
```

The located refusal returned by the authoring commands (`ApplyComponentCommand`, `PreviewComponentMove`, and `wasm.Engine.Apply`/`GroupMovePreview` for component commands). It embeds an unexported `error`, so `*ComponentCommandError` satisfies `error`; its `Error()` text is `"folio8: " + Message`. There are no other exported methods.

| Field | Type | Meaning |
|---|---|---|
| `ElementID` | `string` | The element the refusal concerns, or empty (for example for document-level commands). |
| `DataPath` | `string` | The command field or document location refused, such as `component.geometry`, `column.width`, `table.headerStyle.fontSize`, `fonts.body`, `bands.pageHeader.height`, `pages`, `locale`, or `command` for a command with a duplicated JSON key. |
| `Message` | `string` | A human-readable reason. |

Not every command failure is a `*ComponentCommandError`. Several arrive as plain errors, so use `errors.As` and keep a fallback:

- malformed JSON (`folio8: component command is malformed`)
- an unknown kind or wrong `version` (`folio8: unknown component command`)
- some arity and field errors (`folio8: component command has unknown or missing fields`)
- every `ApplyPageSetupCommand` refusal, whose messages start with `folio8: page`
- a candidate that fails `ParseTemplate` during the command transaction, which is a `*RenderError`

#### Diagnostic codes

Each constant is an untyped string constant whose value is the registry string. "Load" means refused by `ParseTemplate`/`LoadTemplate` (and therefore also by `Validate`), always as a `*RenderError`. "Render error" means `Render`/`RenderTo`/`Validate` return a `*RenderError`. "Render warning" means the render succeeds and the diagnostic is on `Result.Diagnostics`, the `RenderTo` return, or the `Validate` return.

| Go constant | String value | Disposition / when | Meaning |
|---|---|---|---|
| `DiagCodeTemplateMalformed` | `TEMPLATE_MALFORMED` | Load error | The bytes are not a loadable document: not a JSON object, an unreadable value, or an unsupported major version. |
| `DiagCodeTemplateFieldInvalid` | `TEMPLATE_FIELD_INVALID` | Load error | A well-formed document has an unacceptable field value: outside its closed set, missing, of the wrong JSON kind, or a duplicate/misspelled id. Also static barcode/QR content that cannot be encoded. The message names the field, element and value. |
| `DiagCodeStyleLineSpacingInvalid` | `STYLE_LINE_SPACING_INVALID` | Load error | `style.lineSpacing` or `headerStyle.lineSpacing` is outside [0.001, 1000] or has more than three decimal places. |
| `DiagCodeTableFooterSourceForbidden` | `TABLE_FOOTER_SOURCE_FORBIDDEN` | Load error | `footerOf` paired with `footer: "count"`, or a footer companion field without a `footer`. |
| `DiagCodeTableFooterSourceUnresolved` | `TABLE_FOOTER_SOURCE_UNRESOLVED` | Load error | A `sum`/`avg` footer omits `footerOf`, and its column bind is not a shape the source can be derived from. |
| `DiagCodeTableMinHeightUnplaceable` | `TABLE_MIN_HEIGHT_UNPLACEABLE` | Load error | A table's `minHeight` is taller than the content window. |
| `DiagCodeSectionBreakInvalid` | `SECTION_BREAK_INVALID` | Load error (`DataPath` = band) | A `sectionBreak` at or above the band top or at or below the content height, declared twice, or declared on the page header or footer. |
| `DiagCodeSectionBreakStraddled` | `SECTION_BREAK_STRADDLED` | Load error (`ElementID`) | An element's declared box lies on both sides of the section break. |
| `DiagCodePagesInvalid` | `PAGES_INVALID` | Load error (`DataPath` = `pages`, `pages[i]` or `bands.content`) | The `pages` array cannot be loaded: fewer than two entries, a non-object entry or unknown key, a non-boolean `pageBreak`, content in both `pages` and `bands.content`, or a keepTogether group spanning pages. |
| `DiagCodeExpressionInvalid` | `EXPRESSION_INVALID` | Load error (static check) and render error (evaluation) | An expression does not parse or check, or fails when evaluated. |
| `DiagCodeBindingPathAbsent` | `BINDING_PATH_ABSENT` | Render error (`DataPath` = the path) | A data or `params` path an expression needs is absent from the supplied JSON. |
| `DiagCodeContentUnlayoutable` | `CONTENT_UNLAYOUTABLE` | Render error | An ungrouped item (a text line or an image box) is taller than the content window. |
| `DiagCodeStyleColorInvalid` | `STYLE_COLOR_INVALID` | Render error | A style colour consumed at render (`background`, `border.color` or their header equivalents) is not `#RRGGBB`. |
| `DiagCodeDocumentDateInvalid` | `DOCUMENT_DATE_INVALID` | Render error (also `Validate`) | The reserved `documentDate` parameter is present but not a valid RFC 3339 timestamp. |
| `DiagCodeTextClippedWidth` | `TEXT_CLIPPED_WIDTH` | Render warning | A text element's widest line exceeds its declared width and is clipped at the box edge. |
| `DiagCodeTextMissingGlyph` | `TEXT_MISSING_GLYPH` | Render warning | No face in the element's declared chain covers a character. The character is omitted (no glyph, no advance); the message names the rune and the chain. |
| `DiagCodeTextStyleFaceUndeclared` | `TEXT_STYLE_FACE_UNDECLARED` | Render warning | Bold and/or italic was requested, but the chain entry covering the character declares no such face. It is drawn in that entry's base face, with no synthetic emboldening or slant. |
| `DiagCodeEmptyAverage` | `AGGREGATE_EMPTY_AVERAGE` | Render warning | `avg()` over a present but empty collection; the aggregate resolves to empty. |
| `DiagCodeTableHeaderRepeatSuppressed` | `TABLE_HEADER_REPEAT_SUPPRESSED` | Render warning | The repeated table header was dropped on one continuation page because the next row would not fit under it. |
| `DiagCodeTableFooterOrphanSuppressed` | `TABLE_FOOTER_ORPHAN_SUPPRESSED` | Render warning | The footer and its preceding row together exceed the window, so the footer was placed alone. |
| `DiagCodeTableRowClippedHeight` | `TABLE_ROW_CLIPPED_HEIGHT` | Render warning | A header, data or footer row is taller than the whole content window. It was placed alone and cut off at the page bottom, so content is lost. |
| `DiagCodeBarcodeUnencodable` | `BARCODE_UNENCODABLE` | Render warning | A data-bound barcode value contains a character Code 128 cannot encode; the barcode is omitted. |
| `DiagCodeBarcodeModuleTooSmall` | `BARCODE_MODULE_TOO_SMALL` | Render warning | The barcode fits only with modules narrower than 0.25 mm (709 mp). It is still drawn. |
| `DiagCodeBarcodeDoesNotFit` | `BARCODE_DOES_NOT_FIT` | Render warning | The symbol plus quiet zones cannot fit even at 1 mp per module; the barcode is omitted. |
| `DiagCodeQRCodeTooLong` | `QRCODE_TOO_LONG` | Render warning | A data-bound value exceeds a version-40 QR symbol at its error-correction level; the QR code is omitted. |
| `DiagCodeQRCodeModuleTooSmall` | `QRCODE_MODULE_TOO_SMALL` | Render warning | The QR code fits only with modules narrower than 0.5 mm (1418 mp). It is still drawn. |
| `DiagCodeQRCodeDoesNotFit` | `QRCODE_DOES_NOT_FIT` | Render warning | The symbol plus its 4-module quiet zone cannot fit the box's smaller side at 1 mp per module; the QR code is omitted. |
| `DiagCodeSectionBreakSplitsKeepTogether` | `SECTION_BREAK_SPLITS_KEEP_TOGETHER` | Render warning | A keepTogether group has members on both sides of a section break. The break wins and each side is kept together separately. |
| `DiagCodeInternalUnhandledCaveat` | `INTERNAL_UNHANDLED_CAVEAT` | Render warning | Safety net for an internal evaluation caveat with no mapping; not reachable with the current engine. |

Codes are additive: once shipped, a code's string and meaning do not change. The Designer worker additionally uses its own transport codes, such as `COMPONENT_INVALID` and `PAGE_SETUP_INVALID`. Those are not exported by these packages.

### Template and asset helpers

#### `AssetBytes`

```go
func AssetBytes(t *Template, key string) ([]byte, string, error)
```

Returns one asset's decoded raw bytes and its declared media type, looked up by content-addressed key. The key is the 64-character lowercase hex SHA-256 digest used in the document's `assets` map.

Errors:

- a nil `t`
- a key that is not 64 lowercase hex characters
- a key absent from the document
- asset data that fails to decode

It is read-only.

#### `ParameterReferences` and `MaxParameterReferenceNameLength`

```go
func ParameterReferences(tpl *Template) ([]string, error)

const MaxParameterReferenceNameLength = 128
```

Returns the sorted, de-duplicated top-level parameter names the template requests directly. For `{{params.statement.date}}` the name is `"statement"`. It scans every element's `visibleIf` and the `{{ }}` placeholders in text, barcode and qrcode `value`s. Table column bindings are not scanned, and data paths never appear.

Errors:

- a nil template
- an expression that does not parse
- a name longer than `MaxParameterReferenceNameLength` bytes (names are ASCII)
- more than 128 distinct names

#### `StandInData`, `StandInInstant`, `MaxStandInDataPaths`, `MaxStandInDataBytes`

```go
func StandInData(tpl *Template) ([]byte, error)

const StandInInstant = "2024-01-15T12:00:00Z"
const (
	MaxStandInDataPaths = 512
	MaxStandInDataBytes = 256 << 10 // 262144
)
```

Generates a byte-deterministic JSON data document for previewing a template before real sample data exists. The value at each referenced data path is chosen to satisfy the expression wrapping that path; for example a `formatDate` operand receives the fixed `StandInInstant`, never the current time. It does not change `Render`: rendering with absent data still fails with `BINDING_PATH_ABSENT`.

Errors:

- a nil template
- more than `MaxStandInDataPaths` distinct data paths
- a generated document larger than `MaxStandInDataBytes`
- a path whose expression contexts share no legal value (the refusal names the path)

#### `PreviewIdentity`

```go
func PreviewIdentity(template []byte, data Data, params Params, fontSet FontSet) string
```

Returns an opaque lowercase hex SHA-256 digest over the inputs that can affect a render: the template bytes, `data`, `params`, `Version` and every font face (name and bytes). Equal inputs give an equal digest, and changing any of them changes it. It never fails. Compare digests; do not try to recompute them, as the byte construction is an engine-internal protocol.

#### `LocaleTableVersion` and `Version`

```go
const LocaleTableVersion = expr.LocaleTableVersion // currently 1 (untyped integer)
const Version = "0.0.0-dev"
```

`LocaleTableVersion` identifies the built-in locale formatting table; its current value is `1`. `Version` is the library version string. It is included in `PreviewIdentity` and reported by `wasm.RenderResult.Version`. No release tag exists yet.

#### `TableColumns`, `TableColumnsProjection`, `TableColumnProjection`

```go
func TableColumns(t *Template, tableID string) (TableColumnsProjection, error)
```

An editor-facing, read-only projection of one table's configuration: committed values plus the values the renderer resolves.

Errors:

- a nil `t`
- an empty or longer-than-128-byte `tableID`
- a table that is not found, or a component that is not a table
- more than 128 columns
- a `bind` that is not a root collection path ending in `[]` (at most 256 bytes)
- an alias that is invalid, longer than 64, or reserved (`params`, `page`, `pages`)
- a column that cannot be projected, such as a label over 256 characters or a non-positive width
- width allocation failures, as a `*RenderError` with `DataPath` `table.width` or `column.<field>`

In `TableColumnsProjection`, "committed" members hold what the document declares, with `""`/`0`/`false` meaning absent. "Resolved" members hold what rendering uses after the header style cascade.

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Sizing` | `string` | `sizing` | `"points"` (fixed column widths) or `"proportion"` (the table declares a total `width`). |
| `TotalWidth` | `int64` | `totalWidth` | Table width, millipoints. |
| `TableID` | `string` | `tableId` | The requested table id. |
| `Collection` | `string` | `collection` | The table `bind`, e.g. `items[]`. |
| `Alias` | `string` | `alias` | Row alias; `"row"` when `as` is absent. |
| `HeaderHeight` | `int64` | `headerHeight` | Millipoints (required field, so committed = resolved). |
| `AltRowBackground` | `string` | `altRowBackground` | `#RRGGBB` or `""`. |
| `HeaderFontFamily` / `HeaderFontFamilyResolved` | `string` | `headerFontFamily` / `headerFontFamilyResolved` | Chain name. |
| `HeaderFontSize` / `HeaderFontSizeResolved` | `int64` | `headerFontSize` / `headerFontSizeResolved` | Millipoints. |
| `HeaderLineSpacing` / `HeaderLineSpacingResolved` | `int64` | `headerLineSpacing` / `headerLineSpacingResolved` | Thousandths. |
| `HeaderBackground` / `HeaderBackgroundResolved` | `string` | `headerBackground` / `headerBackgroundResolved` | Colour. |
| `HeaderColor` / `HeaderColorResolved` | `string` | `headerColor` / `headerColorResolved` | Text colour. |
| `HeaderValign` / `HeaderValignResolved` | `string` | `headerValign` / `headerValignResolved` | Vertical alignment. |
| `HeaderAlign` / `HeaderAlignResolved` | `string` | `headerAlign` / `headerAlignResolved` | Horizontal alignment. |
| `HeaderBold` / `HeaderBoldResolved` | `bool` | `headerBold` / `headerBoldResolved` | Committed `false` cannot be told apart from absent. |
| `HeaderItalic` / `HeaderItalicResolved` | `bool` | `headerItalic` / `headerItalicResolved` | As above. |
| `HeaderBorderWidth` / `HeaderBorderWidthResolved` | `string` | `headerBorder.width` / `headerBorder.widthResolved` | Millipoints written as a decimal string; `""` means absent and `"0"` means declared zero. |
| `HeaderBorderColor` / `HeaderBorderColorResolved` | `string` | `headerBorder.color` / `headerBorder.colorResolved` | Colour, or `""`. |
| `HeaderBorderEdges` / `HeaderBorderEdgesResolved` | `string` | `headerBorder.edges` / `headerBorder.edgesResolved` | Comma-joined edges in canonical order; a resolved `""` means no header border is painted. |
| `MinHeight` | `int64` | `minHeight` | Millipoints; `0` means absent. |
| `RulesWidth` / `RulesWidthResolved` | `string` | `rules.width` / `rules.widthResolved` | Millipoints as a string; `""` means absent. |
| `RulesColor` / `RulesColorResolved` | `string` | `rules.color` / `rules.colorResolved` | Colour. |
| `RulesBetween` | `string` | `rules.between` | Comma-joined subset of `columns`, `rows` in canonical order; `""` means none. |
| `PaddingLeft`, `PaddingRight` | `string` | `paddingLeft`, `paddingRight` | Committed `style.padding` sides, millipoints as strings (may be negative); `""` means absent. |
| `PaddingHeaderOverride` | `bool` | `paddingHeaderOverride` | `headerStyle.padding` exists, so the two values above do not reach the header row. |
| `Columns` | `[]TableColumnProjection` | `columns` | One entry per column, in order. |

The dotted JSON keys (for example `headerBorder.width`) are literal key names, not nested objects.

`TableColumnProjection`:

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Proportion` | `string` | `proportion` | The column's declared proportion as a decimal string, or `""`. |
| `ID` | `string` | `id` | Column id. |
| `Header` | `string` | `header` | Column label. |
| `Width` | `int64` | `width` | Allocated width, millipoints. |
| `Align` | `string` | `align` | Column `align`; `"left"` when absent. |
| `HeaderAlign` | `string` | `headerAlign` | Committed `headerAlign`, or `""`. |
| `HeaderAlignResolved` | `string` | `headerAlignResolved` | Alignment the header cell prints with; never `""`. |
| `Binding` | `string` | `binding` | The column `bind` text. |
| `RowField` | `string` | `rowField` | Row-relative field when the bind is a simple row path. |
| `RowFieldEditable` | `bool` | `rowFieldEditable` | The bind is empty or a simple row path the editor can edit as a field. |
| `Footer` | `string` | `footer` | `sum`, `avg`, `count` or `""`. |
| `FooterOf` | `string` | `footerOf` | Explicit footer source path, or `""`. |
| `FooterFormat` | `string` | `footerFormat` | Footer format string, or `""`. |

### Authoring and canvas (editor-facing)

These functions let an editor draw a template and change it without ever holding the template model. Command payloads and projection shapes follow folio8 Designer and are not a frozen API.

#### `Canvas`

```go
func Canvas(t *Template) (CanvasProjection, error)
```

Returns page geometry, bands, components, font chains and section-break/page data, all in millipoints. It does not shape text, so it has no `TextPaint`, image, barcode or QR paint and no pagination. `ContentWindowCount` is the number of designed pages (one window per page), every origin is `0`, and `ContentWindowCountIsExact` is `false`.

Errors: a nil `t`, an unsupported page size, page setup leaving no positive content region, geometry beyond `MaxCanvasMillipoints`, and projection bounds (for example an authored string over 512 bytes). It is read-only.

#### `CanvasWithTextPaint`

```go
func CanvasWithTextPaint(t *Template, fs FontSet) (CanvasProjection, error)
```

`Canvas` plus a production-parity paint plan: shaped, pre-broken text lines and fragments, image draw rectangles, barcode bars, QR module runs, header label lines, and content-window pagination computed with `fs`. It is read-only and does not change the template or its serialization. Wherever the canvas draws the document (for example `wasm.Engine`), use this projection.

#### `SnapToGrid` and `GridIncrement`

```go
func SnapToGrid(proposed geom.Length) (geom.Length, bool)

const GridIncrement int64 = 6000
```

Rounds a millipoint value to the nearest multiple of `GridIncrement` (6 pt), with a midpoint rounding away from zero. The boolean is `false` when rounding would overflow.

`geom.Length` is a defined `int64` type in the module's `internal/geom` package. Code outside the module cannot import or name it. It *can* pass an untyped constant, and convert the result with `int64(...)`; a variable of type `int64` cannot be passed without naming the type:

```go
snapped, ok := folio8.SnapToGrid(3500) // untyped constant: compiles
fmt.Println(int64(snapped), ok)       // 6000 true; SnapToGrid(-3000) gives -6000
```

The authoring commands' `"snap": true` applies this same rounding inside the engine, so most callers never call it.

#### `MaxCanvasMillipoints`

```go
const MaxCanvasMillipoints int64 = 9007199254740991
```

The largest magnitude any projected or commanded length may have: JavaScript's `Number.MAX_SAFE_INTEGER`, so projections are exact in the browser. Commands refuse values beyond it.

#### `ApplyComponentCommand`

```go
func ApplyComponentCommand(t *Template, command []byte, fonts ...FontSet) (CanvasProjection, error)
```

Decodes and applies one authoring command (JSON bytes, see the catalog below) to `t`.

- **Input.** `command` must be exactly one JSON object with a top-level `"version": 1` literal and a known `"kind"`. Trailing JSON, a repeated key at any depth, missing fields and unexpected fields are all refused. `fonts` is optional and used only by `moveComponents` with `constrainToWindow`, to paginate like `CanvasWithTextPaint`; pass the same `FontSet` the canvas uses.
- **Mutation.** On success `t` is modified in place, so later `SerializeTemplate(t)` reflects the change. The returned value is a `Canvas(t)` projection of the new document, without text paint; call `CanvasWithTextPaint` if you need paint.
- **Refusal.**
  - Most kinds apply to a canonical working copy and install it only after it serializes, reparses and projects. On any refusal `t` is untouched. This covers every table, font-chain, property and asset kind, `moveComponents`, `deleteComponents` and `duplicateComponents`.
  - A few single-element kinds finish all checks first and then write `t` directly: `moveComponent`, `resizeComponent`, `setComponentBounds`, `deleteComponent`, `duplicateComponent`, `createComponent`/`dropComponent`, the page, section-break, band-height and document-setting kinds.
  - Of these direct kinds, some restore the previous value if the final projection fails; `moveComponent`, `resizeComponent`, `setComponentBounds` and `deleteComponent` do not.
  - For an unconditional all-or-nothing guarantee, apply commands to a copy made with `SerializeTemplate` + `ParseTemplate`, as `wasm.Engine.Apply` does.
- **Errors.** Mostly `*ComponentCommandError`; see that type for the plain-error cases.

#### `ApplyPageSetupCommand`

```go
func ApplyPageSetupCommand(t *Template, command []byte) (CanvasProjection, error)
```

Applies the single `pageSetup` command (payload below) to `t` in place and returns `Canvas(t)`. It refuses without changing `t`:

- a repeated JSON key
- anything other than exactly seven top-level keys
- a wrong `kind` or `version`
- an invalid preset or orientation
- negative margins
- non-positive custom dimensions
- a change that would leave a table `minHeight` or a section break beyond the new content window
- a projection failure

A refusal after the page fields were written restores the previous canonical page. Errors are plain errors whose messages begin `folio8: page` (or `folio8: unknown page setup command` / `folio8: page setup command is malformed`), not `*ComponentCommandError`.

#### `PreviewComponentMove` and `ComponentMove`

```go
func PreviewComponentMove(t *Template, command []byte, fonts ...FontSet) (ComponentMove, error)

type ComponentMove struct {
	DX int64 `json:"dx"`
	DY int64 `json:"dy"`
}
```

Runs the `moveComponents` solver without changing `t` and returns the translation a real `moveComponents` would apply, after clamping to bands and windows and optional snapping.

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `DX` | `int64` | `dx` | Accepted horizontal displacement, millipoints. |
| `DY` | `int64` | `dy` | Accepted vertical displacement, millipoints. |

Only a `moveComponents` payload is accepted, including `expectedRevision`. The revision is checked here only for presence and range; `wasm.Engine` compares it with the session revision. Pass the canvas `FontSet` when using `constrainToWindow`. Refusals are `*ComponentCommandError`, or a plain error for malformed JSON.

#### `CanvasProjection`

Every array field below is always emitted as an array (never `null`). Only the four section-break fields use `omitempty`.

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Width`, `Height` | `int64` | `width`, `height` | Oriented page size, millipoints. |
| `Locale` | `string` | `locale` | Document locale: `en`, `th`, `zh-Hans` or `ja`. |
| `UTCOffset` | `string` | `utcOffset` | Document `±HH:MM` offset. |
| `Orientation` | `string` | `orientation` | `portrait` or `landscape`. |
| `Preset` | `string` | `preset` | `A4`, `Letter` or `custom`. |
| `MarginTop`, `MarginRight`, `MarginBottom`, `MarginLeft` | `int64` | `marginTop` … `marginLeft` | Millipoints. |
| `GridIncrement` | `int64` | `gridIncrement` | Always `GridIncrement` (6000). |
| `CommandWidth`, `CommandHeight` | `int64` | `commandWidth`, `commandHeight` | The unrotated size a `pageSetup` command would carry. |
| `Bands` | `[]CanvasBand` | `bands` | `pageHeader`, `content`, `pageFooter`, in that order. |
| `Components` | `[]CanvasComponent` | `components` | Every element. |
| `FontFamilies` | `[]string` | `fontFamilies` | Sorted names of declared non-empty chains (the values `fontFamily` may take). |
| `FontChains` | `[]CanvasFontChain` | `fontChains` | The same chains with their entries. |
| `DefaultFontSize` | `int64` | `defaultFontSize` | Size used when no `fontSize` is set, millipoints. |
| `DefaultLineSpacing` | `int64` | `defaultLineSpacing` | Ratio used when no `lineSpacing` is set, thousandths. |
| `ContentWindowHeight` | `int64` | `contentWindowHeight` | One page's content column height, millipoints. |
| `ContentWindowCount` | `int64` | `contentWindowCount` | Number of content windows (sheets) the canvas draws; at least 1. |
| `ContentWindowOrigins` | `[]int64` | `contentWindowOrigins` | Start of each window in its page's content-band frame (page-local), one per window. |
| `ContentWindowPages` | `[]int` | `contentWindowPages` | Designed page (0-based) of each window. |
| `PageBreaks` | `[]bool` | `pageBreaks` | Page Break setting per designed page; index 0 is always `true`. |
| `ContentWindowCountIsExact` | `bool` | `contentWindowCountIsExact` | `false` when the count may be wrong: a data-bound table, unshapeable text, an element taller than a window, a `visibleIf` element, or any projection from `Canvas`. |
| `SectionBreak` | `*int64` | `sectionBreak,omitempty` | One-page documents: break offset in millipoints; absent if none. |
| `SectionBreakAnchor` | `*bool` | `sectionBreakAnchor,omitempty` | One-page documents: present (and `false`) only for an unanchored break. |
| `SectionBreaks` | `[]*int64` | `sectionBreaks,omitempty` | Multi-page documents: one offset or `null` per page. |
| `SectionBreakAnchors` | `[]bool` | `sectionBreakAnchors,omitempty` | Multi-page documents: `false` only where that page's break is unanchored. |

#### `CanvasBand`

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Name` | `string` | `name` | `pageHeader`, `content` or `pageFooter`. |
| `X`, `Y`, `Width`, `Height` | `int64` | `x`, `y`, `width`, `height` | Band rectangle on the page, millipoints, top-left origin with Y down. |

#### `CanvasComponent`

Component `X`/`Y` are **band-relative** millipoints. Optional pointer fields are omitted when absent.

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Authored` | `*CanvasAuthoredProperties` | `authored,omitempty` | Authored-state evidence for the inspector. |
| `ID` | `string` | `id` | Element id. |
| `Type` | `string` | `type` | `text`, `image`, `table`, `line`, `rect`, `barcode` or `qrcode`. |
| `Band` | `string` | `band` | Band name. |
| `X`, `Y`, `Width`, `Height` | `int64` | `x`, `y`, `width`, `height` | Box, millipoints; a table's height is its header height. |
| `Resizable` | `bool` | `resizable` | Whether resize commands apply (tables are not resizable). |
| `Value` | `*string` | `value,omitempty` | Literal value text. |
| `Binding` | `*string` | `binding,omitempty` | Display label for a direct binding. |
| `VisibleIf` | `*string` | `visibleIf,omitempty` | Visibility condition text. |
| `FontFamily` | `*string` | `fontFamily,omitempty` | Chain name. |
| `FontSize` | `*int64` | `fontSize,omitempty` | Millipoints. |
| `LineSpacing` | `*int64` | `lineSpacing,omitempty` | Thousandths. |
| `Bold`, `Italic` | `*bool` | `bold,omitempty`, `italic,omitempty` | Style flags. |
| `Align`, `Valign` | `*string` | `align,omitempty`, `valign,omitempty` | Alignment. |
| `Background`, `Color` | `*string` | `background,omitempty`, `color,omitempty` | `#RRGGBB`. |
| `BorderWidth` | `*int64` | `borderWidth,omitempty` | Millipoints. |
| `BorderColor` | `*string` | `borderColor,omitempty` | `#RRGGBB`. |
| `BorderEdges` | `[]string` | `borderEdges,omitempty` | Subset of `top`, `right`, `bottom`, `left`. |
| `TableBind` | `*string` | `tableBind,omitempty` | Table collection path. |
| `PaddingTop`, `PaddingRight`, `PaddingBottom`, `PaddingLeft` | `*int64` | `paddingTop,omitempty` … | Millipoints. |
| `TextPaint` | `*CanvasTextPaint` | `textPaint,omitempty` | Text paint plan (only from `CanvasWithTextPaint`). |
| `Image` | `*CanvasImagePaint` | `image,omitempty` | Image paint, present only when the asset decodes. |
| `ImageUnavailable` | `*string` | `imageUnavailable,omitempty` | `"missing"` or `"undecodable"`, when an image has no paint. |
| `Barcode` | `*CanvasBarcodePaint` | `barcode,omitempty` | Barcode bars. |
| `BarcodeUnavailable` | `*string` | `barcodeUnavailable,omitempty` | `"unencodable"` or `"doesNotFit"`. |
| `QRCode` | `*CanvasQRCodePaint` | `qrcode,omitempty` | QR module runs. |
| `QRCodeUnavailable` | `*string` | `qrcodeUnavailable,omitempty` | `"tooLong"` or `"doesNotFit"`. |
| `Columns` | `[]CanvasTableColumn` | `columns,omitempty` | Table columns; absent for non-tables and for tables without columns. |
| `BelowSectionBreak` | `*bool` | `belowSectionBreak,omitempty` | For content components on a page with a break: at or below it (moves with the section). |
| `Page` | `int` | `page` | Designed page (0-based) of a content component; `0` for header/footer components. |

#### `CanvasAuthoredProperties` and `AuthoredProperty[T]`

```go
type AuthoredProperty[T any] struct {
	State string `json:"state"`
	Value *T     `json:"value,omitempty"`
}
```

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `State` | `string` | `state` | `"absent"` (key not in the document), `"null"` (explicit null, or its parent `style`/`border` is null) or `"value"`. |
| `Value` | `*T` | `value,omitempty` | The authored value when `State` is `"value"`. |

`CanvasAuthoredProperties` holds one `AuthoredProperty` per inspector field. For kinds other than `text` and `table`, the text fields (`FontFamily` through `Color`) are always `"absent"`.

| Field | Type | JSON key |
|---|---|---|
| `VisibleIf` | `AuthoredProperty[string]` | `visibleIf` |
| `FontFamily` | `AuthoredProperty[string]` | `fontFamily` |
| `FontSize` | `AuthoredProperty[geom.Length]` | `fontSize` |
| `LineSpacing` | `AuthoredProperty[int64]` (thousandths) | `lineSpacing` |
| `Bold` | `AuthoredProperty[bool]` | `bold` |
| `Italic` | `AuthoredProperty[bool]` | `italic` |
| `Align` | `AuthoredProperty[string]` | `align` |
| `Valign` | `AuthoredProperty[string]` | `valign` |
| `Color` | `AuthoredProperty[string]` | `color` |
| `Background` | `AuthoredProperty[string]` | `background` |
| `BorderWidth` | `AuthoredProperty[geom.Length]` | `borderWidth` |
| `BorderColor` | `AuthoredProperty[string]` | `borderColor` |
| `BorderEdges` | `AuthoredProperty[[]string]` | `borderEdges` |
| `ErrorCorrection` | `AuthoredProperty[string]` | `errorCorrection` (qrcode only: `L`/`M`/`Q`/`H`; absent means the default `M`) |

`FontSize` and `BorderWidth` use the internal `geom.Length` type, which external code cannot name. On the JSON wire they are plain integer millipoints, for example `{"state":"value","value":12000}`. In Go you can still read them through conversion without naming the type: `if v := p.FontSize.Value; v != nil { size := int64(*v) }`.

#### Paint and chain types

`CanvasTextPaint` holds the text paint plan for one text component:

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Overflow` | `bool` | `overflow` | Laid-out text overflows its box. |
| `Truncated` | `bool` | `truncated` | The plan is a prefix: the projection stopped at a painting bound. The PDF still renders all the text. |
| `Lines` | `[]CanvasTextLine` | `lines` | Pre-broken lines. |

`CanvasTextLine` holds one engine line. Coordinates are band-relative millipoints, top-left origin, Y down:

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Top` | `int64` | `top` | Line top. |
| `Baseline` | `int64` | `baseline` | Baseline position. |
| `Advance` | `int64` | `advance` | Vertical advance to the next line. |
| `Width` | `int64` | `width` | Line width. |
| `Fragments` | `[]CanvasTextFragment` | `fragments` | Positioned runs. |

`CanvasTextFragment` is one run drawn in a single face:

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Text` | `string` | `text` | Characters of the run. |
| `X` | `int64` | `x` | Band-relative paint origin, millipoints. |
| `AssetKey` | `string` | `assetKey,omitempty` | Set when the face is embedded in the document's assets. |
| `Face` | `string` | `face,omitempty` | Set when the face comes from the caller's `FontSet`. Exactly one of `AssetKey` and `Face` is set. |

`CanvasImagePaint` holds paint data for a placed image:

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `MediaType` | `string` | `mediaType` | Declared media type. |
| `AssetKey` | `string` | `assetKey` | Asset key; fetch the bytes with `AssetBytes`. |
| `Width`, `Height` | `int64` | `width`, `height` | Validated intrinsic size, **pixels**. |
| `DrawX`, `DrawY`, `DrawWidth`, `DrawHeight` | `int64` | `drawX`, `drawY`, `drawWidth`, `drawHeight` | Fit-and-centre draw rectangle, band-relative millipoints. |

`CanvasBarcodePaint` and `CanvasBarcodeBar` hold a barcode's bars:

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `CanvasBarcodePaint.ModuleWidth` | `int64` | `moduleWidth` | Module width, millipoints. |
| `CanvasBarcodePaint.Bars` | `[]CanvasBarcodeBar` | `bars` | Bars, each spanning the full component height. |
| `CanvasBarcodeBar.X` | `int64` | `x` | Offset from the component's left edge, millipoints. |
| `CanvasBarcodeBar.Width` | `int64` | `width` | Bar width, millipoints. |

`CanvasQRCodePaint` and `CanvasQRCodeRect` hold a QR code's module runs:

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `CanvasQRCodePaint.ModuleWidth` | `int64` | `moduleWidth` | Module size, millipoints. |
| `CanvasQRCodePaint.Rects` | `[]CanvasQRCodeRect` | `rects` | Horizontal runs of dark modules, rows top to bottom. |
| `CanvasQRCodeRect.X`, `.Y`, `.Width`, `.Height` | `int64` | `x`, `y`, `width`, `height` | Relative to the component's top-left corner, millipoints. |

`CanvasTableColumn` is one table column as the canvas paints it:

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `ID` | `string` | `id` | Column id. |
| `Label` | `string` | `label` | Declared label (may be `""`). |
| `LabelLines` | `[]string` | `labelLines` | Label broken into the lines the engine prints; always an array. |
| `Width` | `int64` | `width` | Millipoints, projected verbatim. |
| `HeaderAlign` | `string` | `headerAlign` | Resolved header-cell alignment. |
| `CellAlign` | `string` | `cellAlign` | Resolved data-cell alignment. |
| `Bind` | `string` | `bind` | Column bind (may be `""`). |

`CanvasFontChain` and `CanvasFontChainEntry` describe the declared font chains:

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `CanvasFontChain.Name` | `string` | `name` | Chain name. |
| `CanvasFontChain.Entries` | `[]CanvasFontChainEntry` | `entries` | Entries in authored order. |
| `CanvasFontChainEntry.Face` | `string` | `face` | `FontSet` face name for a named-face entry, else `""`. |
| `CanvasFontChainEntry.AssetKey` | `string` | `assetKey` | Asset key for an embedded-face entry, else `""`. Exactly one of `Face`/`AssetKey` is non-empty. |
| `CanvasFontChainEntry.Family` | `string` | `family` | Embedded entry: the asset's font family, or the asset key when none is declared. `""` for a named face. |
| `CanvasFontChainEntry.Style` | `string` | `style` | Embedded entry's style; `""` for a named face. |
| `CanvasFontChainEntry.Bold`, `.Italic`, `.BoldItalic` | `string` | `bold`, `italic`, `boldItalic` | Declared variant faces, copied verbatim (face names or asset keys, matching the entry kind); `""` when not declared. |

All seven `CanvasFontChainEntry` keys are always emitted.

### Authoring command catalog

`ApplyComponentCommand` (and `wasm.Engine.Apply`) accept the kinds below; `ApplyPageSetupCommand` accepts `pageSetup`. Common rules:

- **Envelope.** The command is one JSON object, and every command includes `"kind"` and `"version": 1`. The version must be the literal `1`: not `1.0`, not `"1"`.
- **Field count.** The field set must match exactly; unknown or missing keys are refused. Optional keys are listed with `?`.
- **Duplicate keys.** A key repeated anywhere in the command is refused.
- **Lengths.** Lengths are JSON numbers in **points**, with at most three decimals and no exponent, converted exactly to millipoints. Where noted, a table length may also be a decimal string.
- **Snapping.** `snap: true` rounds to the 6 pt grid in the engine, before containment checks.
- **Pages.** `page`, `after` and similar indexes are 0-based designed-page numbers. A target `page` applies only to the content band.
- **Refusals.** Errors are described under `ComponentCommandError`. The Designer worker reports them as `COMPONENT_INVALID`, and page-setup refusals as `PAGE_SETUP_INVALID`.

**Differences from the api-inventory baseline.** The source accepts three component kinds the inventory's list omits: `bindComponentScalar`, `bindTableCollection` and `configureTableBinding`. Every kind in the inventory is still accepted. In total there are 44 component kinds plus `pageSetup`.

#### Page setup (`ApplyPageSetupCommand`)

| Kind | Payload fields |
|---|---|
| `pageSetup` | `preset` (`"A4"`, `"Letter"` or `"custom"`), `orientation` (`"portrait"` or `"landscape"`), `width` and `height` (points; the keys are always required, but only read for `custom`), and `margin` (an object with exactly `top`, `right`, `bottom`, `left` in points, none negative). Exactly seven top-level keys. |

```json
{"kind":"pageSetup","version":1,"preset":"custom","orientation":"portrait","width":300,"height":400,"margin":{"top":10,"right":10,"bottom":10,"left":10}}
```

#### Document

| Kind | Payload fields |
|---|---|
| `setDocumentLocale` | `locale`: one of `en`, `th`, `zh-Hans`, `ja`. |
| `setDocumentUTCOffset` | `utcOffset`: `±HH:MM`. |
| `setBandHeight` | `band` (`"pageHeader"` or `"pageFooter"`; the content band's height is derived and refused), `height` (points, not negative), `snap` (bool). Refused when the band would clip its elements or leave no content window. |

```json
{"kind":"setBandHeight","version":1,"band":"pageHeader","height":80,"snap":false}
```

#### Pages

| Kind | Payload fields |
|---|---|
| `addPage` | `after?`: page index. Inserts an empty page with Page Break on, after that page, or at the end if omitted. |
| `deletePage` | `page`. Removes the page and its elements. The last remaining page cannot be deleted. |
| `setPageBreak` | `page` (≥ 1; page 0's Page Break does not apply), `pageBreak` (bool, not `null`). |

#### Section break

| Kind | Payload fields |
|---|---|
| `setSectionBreak` | `offset` (points, content-band-relative; must be inside the band and cross no element), `snap` (bool), `page?`. |
| `removeSectionBreak` | `page?`. Refused when there is no break. Also removes the anchor. |
| `setSectionBreakAnchor` | `anchor` (bool; `true` clears the key, `false` writes it), `page?`. Refused when there is no break. |

```json
{"kind":"setSectionBreak","version":1,"offset":40,"snap":true}
```

#### Components

`type` is one of `text`, `image`, `table`, `line`, `rect`, `barcode`, `qrcode`.

| Kind | Payload fields |
|---|---|
| `createComponent` | `type`, `band` (`pageHeader`, `content` or `pageFooter`), `x`, `y`, `width`, `height` (points, band-relative), `snap`, `page?`. Tables take the band's full width and a 24 pt header, and ignore the given x/width/height. |
| `dropComponent` | `type`, `x`, `y` (a **page** point; the band is hit-tested), `snap`, `page?`. Starter sizes: 72 × 24 pt (image 96 × 48, line height 1 pt, barcode 216 × 48, qrcode 72 × 72). |
| `moveComponent` | `id`, `x`, `y` (band-relative points), `snap`. |
| `moveComponents` | `ids` (non-empty, unique), `referenceId` (one of `ids`; the snap reference), `dx`, `dy` (points), `snap` (bool), `expectedRevision` (integer ≥ 0; `wasm.Engine` requires it to equal the session revision), `constrainToWindow?` (bool), `page?` (move content members, all from one page, to this page). Also the only kind `PreviewComponentMove` accepts. |
| `resizeComponent` | `id`, `width`, `height` (points), `snap`. Tables are refused. |
| `setComponentBounds` | `id`, `x`, `y`, `width`, `height` (points), `snap`. One rectangle change; tables are refused. |
| `deleteComponent` | `id`. |
| `deleteComponents` | `ids`. |
| `duplicateComponent` | `id`, `snap`. The copy is offset 6 pt when it fits and joins no keepTogether group. |
| `duplicateComponents` | `ids`, `snap`, `page?` (paste content copies onto that page). |
| `updateComponentProperties` | `ids` (non-empty, unique), `changes` (non-empty object). Each key maps to an operation object: `{"op":"set","value":…}`, `{"op":"clear"}` (remove the key) or `{"op":"null"}` (write explicit `null`, where the field supports it). The same changes apply to every id, atomically. |
| `setComponentAsset` | `id` (an image), `mediaType` (e.g. `image/png`), `data` (standard base64 of the file). The decoded file is limited to 6,288,384 bytes. The key is the file's SHA-256; the asset the element previously pointed to is removed if nothing else references it. |
| `bindComponentScalar` | `id` (text, barcode or qrcode), `segments` (1–32 non-empty JSON object keys, each ≤ 64 bytes, first segment not `params`). Sets `value` to `{{a.b.c}}`. |

Keys accepted in `changes`, and which element types allow them:

| Key | Allowed on | Operations and value |
|---|---|---|
| `x`, `y` | all | `set` only; points. |
| `width`, `height` | all except `table` | `set` only; positive points. |
| `value` | text, barcode, qrcode | `set` only; a literal string without `{{`/`}}` (use `bindComponentScalar` for bindings). Only for a single id. For barcode/qrcode, `\n`, `\r` and `\\` escapes are decoded. |
| `expression` | text, barcode, qrcode | `set` only; a string that must contain a placeholder. Written to `value`. |
| `visibleIf` | all | `set` (string), `clear` or `null`. |
| `fontFamily` | text, table | `set` (a declared non-empty chain name) or `clear`. |
| `fontSize` | text, table | `set` (positive points) or `clear`. |
| `lineSpacing` | text, table | `set` (a JSON number ratio, e.g. `1.5`, at most three decimals) or `clear`. |
| `bold`, `italic` | text, table | `set` (bool) or `clear`. |
| `align` | text, table | `set` (the element type's closed set; tables do not accept `justify`) or `clear`. |
| `valign` | text, table | `set` or `clear`. |
| `color` | text, table | `set` (`#RRGGBB`), `clear` or `null`. |
| `background` | text, image, table, line, rect | `set` (`#RRGGBB`), `clear` or `null`. |
| `borderWidth` | text, image, table, line, rect | `set` (positive points) or `clear`. |
| `borderColor` | text, image, table, line, rect | `set` (`#RRGGBB`) or `clear`. |
| `borderEdges` | text, image, table, line, rect | `set` (non-empty string array) or `clear`. |
| `paddingTop`, `paddingRight`, `paddingBottom`, `paddingLeft` | table | `set` (points) or `clear`. |
| `errorCorrection` | qrcode | `set` (`L`, `M`, `Q` or `H`) or `clear`. |

```json
{"kind":"createComponent","version":1,"type":"text","band":"content","x":12,"y":12,"width":72,"height":24,"snap":false}
{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"value":{"op":"set","value":"literal text"}}}
{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["customer","name"]}
{"kind":"moveComponents","version":1,"ids":["e5"],"referenceId":"e5","dx":0,"dy":200,"snap":false,"expectedRevision":1,"constrainToWindow":true,"page":1}
```

#### Tables

`id` is the table element id; `columnId` is a column id.

| Kind | Payload fields |
|---|---|
| `addTableColumn` | `id`, `index` (0 … column count). Adds a column labelled `Column N` (72 pt, or proportion 1 in proportional tables); at most 128 columns. |
| `removeTableColumn` | `id`, `columnId`. |
| `moveTableColumn` | `id`, `columnId`, `toIndex`. |
| `updateTableColumn` | `id`, `columnId`, `field`, `value`. Fields: `header` (string ≤ 256 characters), `width` (positive points, number or decimal string; point-width tables only), `proportion` (decimal, number or string; proportional tables only), `align` (`left`/`center`/`right`), `headerAlign` (`left`/`center`/`right`). |
| `setTableWidth` | `id`, `value` (positive points, number or decimal string). Proportional tables only. |
| `bindTableCollection` | `id`, `segments` (identifier keys). Sets `bind` to `a.b[]` and re-roots explicit `footerOf` paths. |
| `configureTableBinding` | `id`, `collection` (root path ending `[]`, ≤ 256 bytes, not under `params`), `alias` (identifier ≤ 64, or `""` for the default `row`). Rewrites row bindings to a new alias. |
| `updateTableColumnBinding` | `id`, `columnId`, `field` (row field path such as `amount`, or `""` to clear). Written as `{{<alias>.<field>}}`. |
| `updateTableColumnExpression` | `id`, `columnId`, `binding` (complete bind text, ≤ 256 bytes). |
| `updateTableColumnFooter` | `id`, `columnId`, `footer` (`sum`, `avg`, `count` or `""`), `footerOf` (root path inside the table collection, or `""`; must be `""` for `count`), `footerFormat` (string ≤ 256, or `""`). Companion fields require a footer. |
| `setTableHeaderHeight` | `id`, `height` (positive points). Has no `op`: the field cannot be cleared. |
| `setTableMinHeight` | `id`, `op` (`set` or `clear`), `value` (positive points, with `set` only; not taller than the content window). |
| `setTableAltRowBackground` | `id`, `op` (`set` or `clear`), `value` (`#RRGGBB`, with `set` only). |
| `updateTableHeaderStyle` | `id`, `field`, `op` (`set` or `clear`), `value` (with `set` only). Fields: `fontFamily`, `fontSize` (points), `lineSpacing`, `background`, `color`, `valign`, `align`, `bold`, `italic`, `border.width`, `border.color`, `border.edges`. Clearing the last field removes `headerStyle`. |
| `updateTableRules` | `id`, `field` (`width`, `color` or `between`), `op` (`set` or `clear`), `value` (with `set` only). `width` is points ≥ 0; `color` is `#RRGGBB`; `between` is an array of `columns`/`rows`, stored in canonical order. Clearing `between` removes the whole `rules` block. |

```json
{"kind":"updateTableColumnFooter","version":1,"id":"e3","columnId":"e4","footer":"sum","footerOf":"","footerFormat":""}
{"kind":"updateTableHeaderStyle","version":1,"id":"e3","field":"fontSize","op":"set","value":14}
{"kind":"setTableMinHeight","version":1,"id":"e3","op":"clear"}
```

#### Font chains

A chain *entry* in a command is either a face-name string or an object `{"face": "<name>"}` that may also carry `bold`, `italic` and `boldItalic` face names. Commands never accept an `asset` entry.

| Kind | Payload fields |
|---|---|
| `addFontChain` | `name` (≤ 512 bytes, not already declared), `entries` (non-empty array of entries). |
| `renameFontChain` | `name`, `to`. Every `style.fontFamily` and `headerStyle.fontFamily` naming the chain is updated. |
| `deleteFontChain` | `name`. Refused while any element names the chain. |
| `addFontChainEntry` | `name`, `index` (0 … entry count), `face` (string). At most 64 entries. |
| `moveFontChainEntry` | `name`, `from`, `to`. |
| `removeFontChainEntry` | `name`, `index`. Refused if it would empty the chain. |
| `embedFontFamily` | `name`, `family`, `style`, `licence`, `licenceText`, `copyright`, `source` (all non-blank strings), `mediaType` (e.g. `font/ttf`), `data` (base64 face bytes, same size limit as `setComponentAsset`), `tail` (array of fallback entries, `[]` allowed). Declares chain `name` as the embedded face followed by `tail`. The face must be a single, non-variable face this build can read. |

```json
{"kind":"addFontChain","version":1,"name":"caption","entries":["Noto Sans","Noto Sans Thai"]}
{"kind":"moveFontChainEntry","version":1,"name":"body","from":0,"to":2}
```

### wasm integration

Package `wasm` is the stateful session engine that folio8 Designer compiles into its WebAssembly worker. The worker entry point is the `main` package in `folio8-go/wasm/cmd/engine`, which is not an importable API. It receives base64 request fields, bounded to 8 MiB each, and maps them onto the methods below. Everything here is an editor-facing helper.

An `Engine` always renders and projects with `fonts.Shipped()`. It holds one template, that template's canonical bytes, a revision counter, the last projection, and undo/redo stacks of at most 100 canonical byte snapshots each. Its fields are not synchronised.

#### `Engine` and `NewEngine`

```go
type Engine struct {
	// Has unexported fields.
}

func NewEngine() *Engine
```

`NewEngine` returns an empty engine: no document, revision 0.

#### Loading

```go
func (e *Engine) Initialize(input []byte) (Snapshot, error)
func (e *Engine) Load(input []byte) (Snapshot, error)
```

Both behave identically. `input` is `.folio` document bytes.

The engine parses them with `folio8.ParseTemplate`, re-serializes them to canonical bytes, and projects with `folio8.CanvasWithTextPaint`. Only after all three steps succeed does it install the document: it clears both history stacks, increments the revision and returns the snapshot. On error nothing changes.

The input slice is copied, not retained.

#### `Engine.Snapshot`

```go
func (e *Engine) Snapshot() Snapshot
```

Returns the current `Snapshot`. It never fails.

#### `Engine.Serialize`

```go
func (e *Engine) Serialize() ([]byte, Snapshot, error)
```

Returns a copy of the canonical `.folio` bytes and the snapshot. It errors when no document is loaded.

#### `Engine.Apply`

```go
func (e *Engine) Apply(command []byte) (Snapshot, error)
```

`command` is an authoring command in the JSON shape described in the catalog.

Preconditions and refusals:

- It errors when no document is loaded.
- A `moveComponents` command must carry `expectedRevision` equal to the current revision, or it is refused as outdated.
- Malformed JSON is refused.

Processing:

1. The command is applied to a fresh parse of the canonical bytes. `kind: "pageSetup"` goes to `folio8.ApplyPageSetupCommand`; any other kind goes to `folio8.ApplyComponentCommand` with the shipped fonts.
2. If the resulting canonical bytes equal the current bytes, the current snapshot is returned. The revision and history are unchanged.
3. Otherwise the new bytes are reparsed and projected with text paint. The previous bytes are pushed to undo, redo is cleared, and the revision is incremented.

Any refusal leaves the engine unchanged.

#### `Engine.Undo` and `Engine.Redo`

```go
func (e *Engine) Undo() (Snapshot, error)
func (e *Engine) Redo() (Snapshot, error)
```

These restore the previous or next canonical bytes and move the current bytes to the opposite stack. The revision **increases** on every successful undo or redo, even though the bytes return to an earlier state. With an empty stack they return the current snapshot together with `ErrNoUndo` or `ErrNoRedo`.

#### `Engine.Validate`

```go
func (e *Engine) Validate() (Snapshot, error)
```

Reparses the engine's canonical bytes with `folio8.ParseTemplate` and returns the snapshot. It does not call `folio8.Validate` and does not render.

#### `Engine.Render`

```go
func (e *Engine) Render(template, data, params []byte) ([]byte, RenderResult, error)
```

Preconditions:

- A document must be loaded.
- All three inputs must be non-empty. Send `{}` for no params.
- `template` must be byte-for-byte equal to the engine's current canonical bytes; it refuses stale template bytes.

The engine parses `template` and calls `folio8.Render` with `data` as `folio8.Data`, `params` as `folio8.Params`, and the shipped fonts. It returns a copy of the PDF bytes and a `RenderResult`. Render errors, such as a `*folio8.RenderError`, are returned unchanged.

#### `Engine.PreviewIdentity`

```go
func (e *Engine) PreviewIdentity(data, params []byte) (string, uint64, error)
```

Returns `folio8.PreviewIdentity` over the canonical bytes, `data`, `params` and the shipped fonts, plus the current revision. Both inputs must be non-empty.

#### `Engine.AssetBytes`

```go
func (e *Engine) AssetBytes(key string) ([]byte, Snapshot, error)
```

Returns `folio8.AssetBytes` raw bytes for `key` (the media type is not returned) and the snapshot. Read-only.

#### `Engine.ParameterReferences`

```go
func (e *Engine) ParameterReferences() ([]string, uint64, error)
```

Returns `folio8.ParameterReferences` as a non-nil slice, plus the current revision.

#### `Engine.StandInData`

```go
func (e *Engine) StandInData() ([]byte, error)
```

Returns `folio8.StandInData` bytes. It carries no revision; correlate with `Snapshot().Revision`.

#### `Engine.TableColumns`

```go
func (e *Engine) TableColumns(tableID string) (TableColumnsResult, error)
```

Returns `folio8.TableColumns` for `tableID`, tagged with the current revision.

#### `Engine.GroupMovePreview`

```go
func (e *Engine) GroupMovePreview(command []byte) (GroupMoveResult, error)
```

`command` is a `moveComponents` payload, and its `expectedRevision` must equal the current revision. The engine calls `folio8.PreviewComponentMove` with the shipped fonts and returns the accepted translation. It does not change bytes, revision or history.

Every read-only query above returns an error when no document is loaded.

#### `ErrNoUndo` and `ErrNoRedo`

```go
var ErrNoUndo = errors.New("folio8 wasm: no undo history")
var ErrNoRedo = errors.New("folio8 wasm: no redo history")
```

Sentinel errors from `Engine.Undo` and `Engine.Redo`. Test them with `errors.Is`.

#### `Snapshot`

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `DocumentState` | `string` | `documentState` | `"empty"` or `"loaded"`. |
| `Revision` | `uint64` | `revision` | Monotonic session revision; incremented by every successful load, committed apply, undo and redo. |
| `ByteLength` | `int` | `byteLength` | Length of the canonical bytes (0 when empty). |
| `CanUndo` | `bool` | `canUndo` | Undo history is non-empty. |
| `CanRedo` | `bool` | `canRedo` | Redo history is non-empty. |
| `Canvas` | `*folio8.CanvasProjection` | `canvas,omitempty` | The current `CanvasWithTextPaint` projection; nil when empty. The pointer is shared with the engine, so treat it as read-only. |

#### `RenderResult`

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `PDFSHA256` | `string` | `pdfSha256` | Lowercase hex SHA-256 of the returned PDF. |
| `Identity` | `string` | `identity` | `PreviewIdentity` digest for this render's inputs. |
| `Revision` | `uint64` | `revision` | Revision whose canonical bytes were rendered. |
| `Diagnostics` | `[]folio8.Diagnostic` | `diagnostics` | Copy of the render's Warning diagnostics. Not `omitempty`; the elements have no JSON tags, so they encode with Go field names. |
| `ElapsedMs` | `int64` | `elapsedMs` | Wall-clock milliseconds spent in `folio8.Render` alone. It is measured with `time`, varies between runs, and never affects PDF bytes. |
| `Version` | `string` | `version` | `folio8.Version`. |

#### `TableColumnsResult`

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Revision` | `uint64` | `revision` | Revision the projection was taken at. |
| `Table` | `folio8.TableColumnsProjection` | `table` | The table projection. |

#### `GroupMoveResult`

| Field | Type | JSON key | Meaning |
|---|---|---|---|
| `Revision` | `uint64` | `revision` | Current revision. |
| `DX` | `int64` | `dx` | Accepted horizontal displacement, millipoints. |
| `DY` | `int64` | `dy` | Accepted vertical displacement, millipoints. |

## Command-line tool

The module also contains a `folio8` command with `validate` and `render` subcommands
(`go run github.com/panitw/folio8/folio8-go/cmd/folio8@main render -data data.json -o out.pdf template.folio`).
It is a thin wrapper over `Validate` and `Render`; its flags are described in the
[repository README](../README.md#render-from-the-command-line).
