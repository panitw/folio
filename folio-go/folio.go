// Package folio is the public entry point of the folio-go rendering
// library.
package folio

import "github.com/panitw/folio/folio-go/internal/pdf"

// Render produces the fixed, single-page PDF 1.7 document described by
// Story 1.1: one filled rectangle on an A4 page, with no compression, no
// /Info dictionary, and no /CreationDate or /ModDate.
//
// PROVISIONAL — Story 1.1 only. This signature exists solely so a
// time-to-first-PDF caller has a name to call and so the two-process
// determinism harness has something to invoke; it takes no template and
// no data. Story 1.4 gives it a template, Story 1.6 data and runtime
// parameters, and Story 1.7 the RenderTo(w io.Writer, …) writer form —
// which arrives without breaking this call, per the architecture spine's
// Deferred row. Do not build against this signature as if it were final.
func Render() ([]byte, error) {
	return pdf.Serialize(), nil
}
