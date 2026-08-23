// This file declares Render (and, at Story 1.7, RenderTo) — the
// world-reading fence (D-1.6.3, AC12): it imports NONE of "os",
// "time", "net", "math/rand". The shell entry points that DO need
// those (LoadTemplate, which reads a path from disk) live in the
// different file folio.go, which may (D-1.4.6 draws the "os" boundary
// at the package, not the file, but this file narrows it further:
// AC12's property is decidable and exact — "the file declaring Render
// imports none of the four" — and fails the moment someone adds a
// convenience os.Getenv to the render entry point, the same shape
// D-1.1.b used for internal/pdf/numbers.go).
//
// AD-12 locale clarification (AC14, D-1.6.3): the source AC's "read no
// … locale" means the HOST locale — the machine Render happens to run
// on. Discovering the machine's locale is banned. Reading the
// document's OWN declared locale is not merely allowed but REQUIRED
// (AD-12: "the document declares one locale tag and one fixed UTC
// offset … an unlisted tag is a load error, not a fallback") —
// internal/template's parser already enforces the closed locale set at
// load time (Story 1.4); this file's job is never to substitute the
// host's locale for it.
//
// QA Finding 8 (this story's review, Minor): the blank line above is
// deliberate. Without it, Go treats this block as a SECOND package
// doc comment (there is no blank line requirement violated by having
// one — the issue was this file's comment sitting directly above
// `package folio` with folio.go's own package comment already present
// elsewhere), and `go doc .` published every ruling id and story
// number here into the library's public documentation, concatenated
// after folio.go's real package comment. A blank line demotes this to
// an ordinary file comment; AC14's AD-12 locale paragraph stays here,
// where it documents this file specifically.

package folio

import (
	"errors"
	"fmt"

	"github.com/panitw/folio/folio-go/internal/bind"
)

// Data is the caller's report data as raw JSON bytes (AC22-AC24,
// D-1.6.4). A named, defined type — never []byte and never a type
// alias (AC23): Data and Params become adjacent same-typed arguments
// at Story 1.7, and in a product whose acceptance fixture is a bank
// statement, that swap must be a compile error, not a support ticket
// (D-1.1.c). Bytes, not a decoded value (AC24): AD-23 requires the
// library to own the UseNumber-preserving decode
// (internal/bind.DecodeData) — a caller-decoded `any` or
// `map[string]any` would arrive with its number literals already
// destroyed by encoding/json's default float64 conversion.
type Data []byte

// errNilTemplate is AC14b's located error for Render(nil, d, f).
var errNilTemplate = errors.New("folio: Render: template is nil (a document must be loaded via LoadTemplate/ParseTemplate first)")

// Render produces a PDF 1.7 document from t, resolving every
// "{{…}}" placeholder in every text element's value against d
// (AC1, AC15-AC21, D-1.6.5) and embedding every font the document's
// text elements use as a subset (AC1, D-1.5.5): the f parameter is
// inserted into its final target position, never appended — target is
// Render(t, d, p, f), so 1.6 adds d before f, 1.7 adds p before f
// (D-1.5.5, verbatim: "No call site is ever reordered, only
// extended").
//
// t must be non-nil (AC14b): Story 1.1's fixture document
// (fixtures/minimal-rect/) predates the public API accepting a
// template at all and is pinned via internal/pdf.Serialize() directly
// (AC14a) — a public Render documented as silently ignoring its
// argument is worse than a located error on nil (D-1.5.9).
//
// d must be syntactically valid JSON (AC24); it is decoded once, via
// internal/bind.DecodeData (json.Decoder.UseNumber under the hood), so
// every number literal a bound text placeholder resolves to keeps its
// own author-written precision (AD-23) rather than being narrowed
// through float64.
//
// PROVISIONAL: text is placed using a provisional band-relative origin
// convention (AC28) that stands in for AD-24's real placement until
// internal/layout exists (Story 2.5); TestProvisionalBandOriginIsPinned
// (render_test.go) fails the day that package exists, forcing this to
// be revisited. Story 1.6 gives Render its data parameter and Story 1.7
// its io.Writer form (D-1.1.c).
func Render(t *Template, d Data, f FontSet) ([]byte, error) {
	if t == nil {
		return nil, errNilTemplate
	}
	data, err := bind.DecodeData(d)
	if err != nil {
		return nil, fmt.Errorf("folio: Render: %w", err)
	}
	return renderDocument(t, data, f)
}
