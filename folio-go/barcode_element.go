// This file holds the barcode element's load-time checks, the designer's
// escape projection for its value, and its canvas paint
// (spec-barcode-qr-elements CAP-4, CAP-5).
//
// ESCAPES LIVE HERE, NOT IN THE ENGINE. A `.folio` file stores the real
// characters (a carriage return is the JSON escape "\r"), and the engine
// never interprets a backslash. The designer's single-line content field
// cannot hold a control character, so the canvas projects the value with
// `\r`, `\n` and `\\` spelled out, and the property command turns them back
// into the characters they name. Both conversions touch only the text
// OUTSIDE {{ }}: expression string literals stay escape-free.
package folio

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/panitw/folio/folio-go/internal/barcode"
	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/template"
)

// checkBarcodeValue is a barcode value's load-time check: every expression
// parses and checks as a text expression does, the reserved page tokens are
// refused (a barcode's content must be known before layout), and the static
// text outside {{ }} must be encodable — a non-ASCII character there could
// never draw for any record, so it is a load Error rather than a Warning.
func checkBarcodeValue(value string, id template.ElementID) error {
	literal, placeholders, trailing, serr := expr.ScanPlaceholders(value)
	if serr != nil {
		return newRenderError(DiagCodeExpressionInvalid, string(id), "value", fmt.Errorf("folio: ParseTemplate: element %s: %s", id, serr))
	}
	for _, ph := range placeholders {
		if ph.Reserved {
			return newRenderError(DiagCodeTemplateFieldInvalid, string(id), "value", fmt.Errorf(
				"folio: ParseTemplate: element %s: a barcode value cannot use {{%s}} — page numbers are resolved after layout, and a barcode's content must be known before it",
				id, strings.TrimSpace(ph.Inner)))
		}
	}
	if err := checkTextExpressions(value, id); err != nil {
		return newRenderError(DiagCodeExpressionInvalid, string(id), "value", err)
	}
	parts := make([]string, 0, len(literal)+1)
	parts = append(parts, literal...)
	parts = append(parts, trailing)
	for _, part := range parts {
		if off, bad := barcode.FirstUnencodable(part); bad {
			r, _ := utf8.DecodeRuneInString(part[off:])
			return newRenderError(DiagCodeTemplateFieldInvalid, string(id), "value", fmt.Errorf(
				"folio: ParseTemplate: element %s: barcode value contains %q (U+%04X), which Code 128 cannot encode — only ASCII 0-127 may appear outside {{ }}",
				id, r, r))
		}
	}
	return nil
}

// encodeBarcodeEscapes is the designer's view of a barcode value: outside
// {{ }}, a backslash becomes `\\`, a carriage return `\r` and a line feed
// `\n`. Placeholders are copied verbatim.
func encodeBarcodeEscapes(value string) string {
	literal, placeholders, trailing, err := expr.ScanPlaceholders(value)
	if err != nil {
		// Unreachable for a loaded document (the loader refuses an
		// unterminated placeholder); escape the whole text rather than guess.
		return escapeBarcodeLiteral(value)
	}
	var b strings.Builder
	for i, ph := range placeholders {
		b.WriteString(escapeBarcodeLiteral(literal[i]))
		b.WriteString("{{")
		b.WriteString(ph.Inner)
		b.WriteString("}}")
	}
	b.WriteString(escapeBarcodeLiteral(trailing))
	return b.String()
}

func escapeBarcodeLiteral(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			b.WriteString(`\\`)
		case '\r':
			b.WriteString(`\r`)
		case '\n':
			b.WriteString(`\n`)
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// decodeBarcodeEscapes is encodeBarcodeEscapes' inverse, applied by the
// property command. Any other backslash sequence, or a trailing backslash, is
// refused: an escape encodes exactly the byte it names, never a guess.
func decodeBarcodeEscapes(text string) (string, error) {
	literal, placeholders, trailing, err := expr.ScanPlaceholders(text)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for i, ph := range placeholders {
		part, err := unescapeBarcodeLiteral(literal[i])
		if err != nil {
			return "", err
		}
		b.WriteString(part)
		b.WriteString("{{")
		b.WriteString(ph.Inner)
		b.WriteString("}}")
	}
	part, err := unescapeBarcodeLiteral(trailing)
	if err != nil {
		return "", err
	}
	b.WriteString(part)
	return b.String(), nil
}

func unescapeBarcodeLiteral(s string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		if i+1 == len(s) {
			return "", fmt.Errorf(`a trailing backslash escapes nothing; write \\ for a backslash`)
		}
		i++
		switch s[i] {
		case '\\':
			b.WriteByte('\\')
		case 'r':
			b.WriteByte('\r')
		case 'n':
			b.WriteByte('\n')
		default:
			return "", fmt.Errorf(`\%c is not a supported escape; use \r, \n or \\`, s[i])
		}
	}
	return b.String(), nil
}

// CanvasBarcodeBar is one bar, in millipoints, its X relative to the
// component's own left edge. Every bar spans the component's full height.
type CanvasBarcodeBar struct {
	X     int64 `json:"x"`
	Width int64 `json:"width"`
}

// CanvasBarcodePaint is the Go-computed geometry the canvas draws for a
// barcode: the same bars layoutBarcode gives the render path.
type CanvasBarcodePaint struct {
	ModuleWidth int64              `json:"moduleWidth"`
	Bars        []CanvasBarcodeBar `json:"bars"`
}

// BarcodeUnavailable's bounded values, set only when Barcode is absent for a
// barcode with a non-empty value.
const (
	barcodeUnavailableUnencodable = "unencodable"
	barcodeUnavailableDoesNotFit  = "doesNotFit"
)

// illustrativeBarcodePlaceholder stands in for every {{ }} placeholder on the
// canvas, which never sees data (owner decision, review pass 1).
const illustrativeBarcodePlaceholder = "0123456789"

// illustrativeBarcodeContent is what the canvas encodes: the value with each
// placeholder replaced by illustrativeBarcodePlaceholder. A static value is
// returned unchanged, so its canvas bars are the PDF's bars.
func illustrativeBarcodeContent(value string) string {
	literal, placeholders, trailing, err := expr.ScanPlaceholders(value)
	if err != nil || len(placeholders) == 0 {
		return value
	}
	var b strings.Builder
	for i := range placeholders {
		b.WriteString(literal[i])
		b.WriteString(illustrativeBarcodePlaceholder)
	}
	b.WriteString(trailing)
	return b.String()
}

// addCanvasBarcodePaint is the barcode's paint producer, beside
// addCanvasImagePaint. A static value is encoded as written; a bound value
// draws illustrative bars (illustrativeBarcodeContent) — the preview shows the
// real code. A barcode that cannot be painted degrades to a bounded reason and
// never fails the projection.
func addCanvasBarcodePaint(t *Template, projection *CanvasProjection) error {
	components := make(map[string]*CanvasComponent, len(projection.Components))
	for i := range projection.Components {
		component := &projection.Components[i]
		components[component.ID] = component
	}
	for _, band := range []struct {
		name     string
		elements []template.Element
	}{
		{bandPageHeader, t.doc.Bands.PageHeader.Elements},
		{bandContent, t.doc.Bands.Content.Elements},
		{bandPageFooter, t.doc.Bands.PageFooter.Elements},
	} {
		for _, element := range band.elements {
			if element.Type != template.ElementBarcode {
				continue
			}
			component := components[string(element.ID)]
			if component == nil || component.Band != band.name {
				return fmt.Errorf("folio: canvas barcode component %q is missing from geometry projection", element.ID)
			}
			if !element.Value.Set || element.Value.Null || element.Value.Value == "" {
				continue
			}
			content := illustrativeBarcodeContent(element.Value.Value)
			lay, warning := layoutBarcode(string(element.ID), content, element.Width.Value, element.Height.Value)
			if len(lay.bars) == 0 {
				if warning != nil {
					reason := barcodeUnavailableDoesNotFit
					if warning.Code == DiagCodeBarcodeUnencodable {
						reason = barcodeUnavailableUnencodable
					}
					component.BarcodeUnavailable = &reason
				}
				continue
			}
			paint := &CanvasBarcodePaint{ModuleWidth: int64(lay.moduleWidth), Bars: make([]CanvasBarcodeBar, 0, len(lay.bars))}
			for _, bar := range lay.bars {
				paint.Bars = append(paint.Bars, CanvasBarcodeBar{X: int64(bar.X), Width: int64(bar.W)})
			}
			component.Barcode = paint
		}
	}
	return nil
}

// canvasBarcodeIsPlaced answers canvasElementIsPlaced's question for a
// barcode: the render path places a column item exactly when bars are drawn.
// A static value is decided here; a bound one is assumed placed, and
// canvasContentBandHasBoundBarcode registers that as a cause of inexactness.
func canvasBarcodeIsPlaced(element template.Element) bool {
	if !element.Value.Set || element.Value.Null || element.Value.Value == "" {
		return false
	}
	if stringsContainsPlaceholder(element.Value.Value) {
		return true
	}
	lay, _ := layoutBarcode(string(element.ID), element.Value.Value, element.Width.Value, element.Height.Value)
	return len(lay.bars) > 0
}

// canvasContentBandHasBoundBarcode is a cause of window-count inexactness:
// whether a bound barcode draws depends on data the canvas does not have.
func canvasContentBandHasBoundBarcode(t *Template) bool {
	for _, element := range t.doc.Bands.Content.Elements {
		if element.Type == template.ElementBarcode && element.Value.Set && !element.Value.Null && stringsContainsPlaceholder(element.Value.Value) {
			return true
		}
	}
	return false
}
