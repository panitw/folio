// This file is the barcode element's render obligation
// (spec-barcode-qr-elements CAP-1, CAP-3, CAP-4): its bindable value is
// resolved through the same bind path a text value uses, encoded as Code 128
// by internal/barcode, fitted into its box in whole millipoints, and emitted
// as black filled page-model rectangles.
//
// The bars travel as a tableRectSource — the carrier element_box.go already
// uses — so pagination, keepTogether, visibility and header/footer repetition
// apply through the element id with no new machinery. A document without a
// barcode contributes nothing here and its bytes are unchanged.
package folio

import (
	"errors"
	"fmt"

	"github.com/panitw/folio/folio-go/internal/barcode"
	"github.com/panitw/folio/folio-go/internal/bind"
	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/pagemodel"
	"github.com/panitw/folio/folio-go/internal/template"
)

// barcodeLayout is one barcode's resolved geometry, box-relative.
type barcodeLayout struct {
	bars        []barcode.Bar
	moduleWidth geom.Length
}

// layoutBarcode is THE ONE derivation of a barcode's bars from its resolved
// content and box, shared by the render path and the canvas paint so the two
// cannot disagree (AD-17).
//
// It returns the layout (empty when nothing is drawn) and at most one
// Warning. Empty content draws nothing and says nothing: a null or empty
// binding is the ordinary "no code for this record" case.
func layoutBarcode(elementID, content string, boxW, boxH geom.Length) (barcodeLayout, *Diagnostic) {
	if content == "" {
		return barcodeLayout{}, nil
	}
	values, err := barcode.Encode(content)
	if err != nil {
		var ue *barcode.UnencodableError
		if errors.As(err, &ue) {
			return barcodeLayout{}, &Diagnostic{
				Severity:  SeverityWarning,
				Code:      DiagCodeBarcodeUnencodable,
				ElementID: elementID,
				Message:   fmt.Sprintf("element %s: barcode not drawn: its value %s", elementID, ue.Error()),
			}
		}
		// Unreachable for non-empty content: UnencodableError is Encode's
		// only failure. Reported under the same code rather than dropped.
		return barcodeLayout{}, &Diagnostic{
			Severity:  SeverityWarning,
			Code:      DiagCodeBarcodeUnencodable,
			ElementID: elementID,
			Message:   fmt.Sprintf("element %s: barcode not drawn: %v", elementID, err),
		}
	}
	runs := barcode.Runs(values)
	fit, ok := barcode.FitWidth(runs, boxW)
	if !ok || boxH <= 0 {
		return barcodeLayout{}, &Diagnostic{
			Severity:  SeverityWarning,
			Code:      DiagCodeBarcodeDoesNotFit,
			ElementID: elementID,
			Message: fmt.Sprintf("element %s: barcode not drawn: %d modules plus %d quiet-zone modules on each side need at least %d mp of width and a positive height, and the box is %d x %d mp — widen the box or shorten the value",
				elementID, barcode.Modules(runs), barcode.QuietZoneModules, barcode.Modules(runs)+2*barcode.QuietZoneModules, boxW, boxH),
		}
	}
	out := barcodeLayout{bars: barcode.Bars(runs, fit), moduleWidth: fit.ModuleWidth}
	if fit.ModuleWidth < barcode.MinModuleWidth {
		return out, &Diagnostic{
			Severity:  SeverityWarning,
			Code:      DiagCodeBarcodeModuleTooSmall,
			ElementID: elementID,
			Message: fmt.Sprintf("element %s: barcode modules are %d mp wide, below the 0.25 mm (%d mp) scanners need — widen the box or shorten the value",
				elementID, fit.ModuleWidth, barcode.MinModuleWidth),
		}
	}
	return out, nil
}

// collectBarcodeRects walks every band's barcode elements in document order.
// It returns one tableRectSource per visible barcode that draws, and each
// band's barcode Warnings, indexed like bands.
//
// Binding runs for EVERY barcode, visible or not (render_visibility.go's R2):
// an absent path is a located Error whichever report the template is handed.
// Only a hidden barcode's output and Warnings are suppressed.
func collectBarcodeRects(doc *Template, bands []bandWithOrigin, data, params bind.Value, visible visibilityVerdicts) ([]tableRectSource, [][]Diagnostic, error) {
	fc := expr.NewFormatContext(doc.doc.Locale, doc.doc.UTCOffset)
	var sources []tableRectSource
	diags := make([][]Diagnostic, len(bands))
	for bandIndex, b := range bands {
		for _, el := range b.band.Elements {
			if el.Type != template.ElementBarcode {
				continue
			}
			if !el.Value.Set || el.Value.Null || el.Value.Value == "" {
				continue
			}
			content, _, _, berr := bind.BindTextSpans(el.Value.Value, data, params, fc, string(el.ID))
			if berr != nil {
				return nil, nil, expressionRuntimeError(string(el.ID), "value", fmt.Errorf("folio: Render: %w", berr))
			}
			if !isVisible(visible, el.ID) {
				continue
			}
			w, h := el.Width.Value, el.Height.Value
			if !el.Width.Set || el.Width.Null || !el.Height.Set || el.Height.Null {
				w, h = 0, 0
			}
			lay, warning := layoutBarcode(string(el.ID), content, w, h)
			if warning != nil {
				diags[bandIndex] = append(diags[bandIndex], *warning)
			}
			if len(lay.bars) == 0 {
				continue
			}
			top := layout.PlaceInBand(b.origin, el.Y)
			rects := make([]pagemodel.Rect, 0, len(lay.bars))
			for _, bar := range lay.bars {
				rects = append(rects, pagemodel.Rect{
					X: el.X + bar.X, Y: top, W: bar.W, H: h,
					HasFill: true, Fill: pagemodel.Color{R: 0, G: 0, B: 0},
				})
			}
			sources = append(sources, tableRectSource{
				band:      bandIndex,
				elementID: string(el.ID),
				top:       top,
				bottom:    top + h,
				rects:     rects,
			})
		}
	}
	return sources, diags, nil
}
