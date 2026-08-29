package folio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/template"
	"github.com/panitw/folio/folio-go/internal/text"
)

// GridIncrement is the fixed six-point grid used by the designer projection.
// SnapNearest's documented midpoint rule is away from zero.
const GridIncrement int64 = 6000

// MaxCanvasMillipoints keeps every document value emitted to the JSON/JS paint
// boundary within Number.MAX_SAFE_INTEGER. The page-setup command has the same
// bound, so a successful command can never strand the worker with an
// unrepresentable projection.
const MaxCanvasMillipoints int64 = 9007199254740991
const maxCanvasPropertyString = 512
const maxCanvasTextLines = 256
const maxCanvasTextFragments = 512

// CanvasTextFragment is a shaped, positioned paint fragment. It is not a
// document text node: x is the engine-owned, band-relative paint origin.
type CanvasTextFragment struct {
	Text string `json:"text"`
	X    int64  `json:"x"`
}

// CanvasTextLine is one pre-broken engine line. All coordinates are
// band-relative, top-left/Y-down millipoints. Advance is retained so the
// browser never derives a following line's origin from CSS metrics.
type CanvasTextLine struct {
	Top       int64                `json:"top"`
	Baseline  int64                `json:"baseline"`
	Advance   int64                `json:"advance"`
	Width     int64                `json:"width"`
	Fragments []CanvasTextFragment `json:"fragments"`
}

// CanvasTextPaint is the closed browser paint plan for one text component.
// It deliberately carries no CSS, browser metric, or document-schema input.
type CanvasTextPaint struct {
	Overflow bool             `json:"overflow"`
	Lines    []CanvasTextLine `json:"lines"`
}

// SnapToGrid is the reusable core-command seam for Story 5.7 placement.
// It uses the fixed six-point grid and half-away-from-zero rule; callers pass
// millipoints and never browser pixels.
func SnapToGrid(proposed geom.Length) (geom.Length, bool) {
	return proposed.SnapNearest(geom.Length(GridIncrement))
}

type CanvasBand struct {
	Name   string `json:"name"`
	X      int64  `json:"x"`
	Y      int64  `json:"y"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}
type CanvasComponent struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Band      string `json:"band"`
	X         int64  `json:"x"`
	Y         int64  `json:"y"`
	Width     int64  `json:"width"`
	Height    int64  `json:"height"`
	Resizable bool   `json:"resizable"`
	// The following explicitly named optional values are the minimum committed
	// property-panel projection. This is not a generic style or document bag.
	Value *string `json:"value,omitempty"`
	// Binding is a bounded, Go-derived paint label for a direct text binding.
	// It is not a general expression/template projection and cannot be used to
	// reconstruct canonical document bytes in the browser.
	Binding       *string           `json:"binding,omitempty"`
	VisibleIf     *string           `json:"visibleIf,omitempty"`
	FontFamily    *string           `json:"fontFamily,omitempty"`
	FontSize      *int64            `json:"fontSize,omitempty"`
	Bold          *bool             `json:"bold,omitempty"`
	Italic        *bool             `json:"italic,omitempty"`
	Align         *string           `json:"align,omitempty"`
	Valign        *string           `json:"valign,omitempty"`
	Background    *string           `json:"background,omitempty"`
	BorderWidth   *int64            `json:"borderWidth,omitempty"`
	BorderColor   *string           `json:"borderColor,omitempty"`
	BorderEdges   []string          `json:"borderEdges,omitempty"`
	TableBind     *string           `json:"tableBind,omitempty"`
	PaddingTop    *int64            `json:"paddingTop,omitempty"`
	PaddingRight  *int64            `json:"paddingRight,omitempty"`
	PaddingBottom *int64            `json:"paddingBottom,omitempty"`
	PaddingLeft   *int64            `json:"paddingLeft,omitempty"`
	TextPaint     *CanvasTextPaint  `json:"textPaint,omitempty"`
	Image         *CanvasImagePaint `json:"image,omitempty"`
	// ImageUnavailable is a small, bounded discriminant set ONLY when this
	// is an image element and Image is absent (Finding 9, review of
	// 2026-08-29): "missing" when the element's own asset key is not in
	// the document's assets map, or "undecodable" when the key resolves
	// but the bytes fail to decode or the media type is one this library
	// version cannot render. D-5.13.2's "one Go-side signal drives both"
	// governed the media-type case only; collapsing a dangling asset
	// reference into that same undecodable text was a defect — the media
	// type there is fine, the asset is simply gone. This does not widen
	// the projection's authority: it is still Go stating which of two
	// bounded, enumerated reasons applies, never bytes or a path.
	ImageUnavailable *string `json:"imageUnavailable,omitempty"`
}

// imageUnavailableMissing / imageUnavailableUndecodable are
// ImageUnavailable's only two values (Finding 9). Kept as named constants,
// not inline literals, so the Go producer and any future consumer cannot
// drift on spelling.
const (
	imageUnavailableMissing     = "missing"
	imageUnavailableUndecodable = "undecodable"
)

// CanvasImagePaint is Story 5.13's read-only, paint-only projection of one
// placed image element's Go-owned display data: the declared media type,
// the asset's content-addressed key, VALIDATED intrinsic pixel dimensions,
// and the fit-and-centre draw rectangle already computed for the PDF
// (resolveImagePlacement), in BAND-RELATIVE millipoints (matching this
// component's own X/Y, D-5.13.2's "Frame" clause). It carries no asset
// BYTES (AD-17: a paint-only projection must not carry anything that
// reconstructs canonical bytes or the assets map): AssetKey is only a
// LOOKUP TOKEN for the separate, explicit, per-key bytes request
// (AssetBytes/wasm.Engine.AssetBytes) the canvas uses to obtain what it
// paints — a key alone cannot reconstruct the assets map or canonical
// bytes any more than a table id (already sent on every projection) can.
// The inspector abbreviates this same key for DISPLAY (a formatting choice
// over a value Go already supplied, same as it formats millipoints as
// "12.5pt"); Go does not truncate it on the wire, because the canvas needs
// the real key to ask for bytes.
//
// The whole field is present only when the referenced asset decodes
// successfully through the recognised-image path — D-5.13.2's "Absence,
// not zero": DecodedImage.Width()/Height() are reachable only through
// decodeRecognisedImage, so a legally-loaded asset of an unrecognised media
// type (or one whose bytes fail to decode) has no dimensions and no
// computable rectangle. Rather than carry two independently-absent signals
// (a known media type but a missing rectangle), the ENTIRE paint is absent
// together — ONE Go-side signal drives both AC2's inspector failure text
// and AC3's canvas placeholder, never two.
type CanvasImagePaint struct {
	MediaType  string `json:"mediaType"`
	AssetKey   string `json:"assetKey"`
	Width      int64  `json:"width"`
	Height     int64  `json:"height"`
	DrawX      int64  `json:"drawX"`
	DrawY      int64  `json:"drawY"`
	DrawWidth  int64  `json:"drawWidth"`
	DrawHeight int64  `json:"drawHeight"`
}

const maxCanvasBindingString = 256

type CanvasProjection struct {
	Width         int64             `json:"width"`
	Height        int64             `json:"height"`
	Orientation   string            `json:"orientation"`
	Preset        string            `json:"preset"`
	MarginTop     int64             `json:"marginTop"`
	MarginRight   int64             `json:"marginRight"`
	MarginBottom  int64             `json:"marginBottom"`
	MarginLeft    int64             `json:"marginLeft"`
	GridIncrement int64             `json:"gridIncrement"`
	CommandWidth  int64             `json:"commandWidth"`
	CommandHeight int64             `json:"commandHeight"`
	Bands         []CanvasBand      `json:"bands"`
	Components    []CanvasComponent `json:"components"`
}

// Canvas returns immutable paint geometry. It intentionally exposes neither
// template fields nor elements, canonical bytes, or browser measurements.
func Canvas(t *Template) (CanvasProjection, error) {
	if t == nil {
		return CanvasProjection{}, errNilTemplate
	}
	w, h, err := canvasDimensions(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	m := t.doc.Page.Margin
	header, footer := t.doc.Bands.PageHeader.Height.Value, t.doc.Bands.PageFooter.Height.Value
	if w <= 0 || h <= 0 || m.Left < 0 || m.Right < 0 || m.Top < 0 || m.Bottom < 0 || m.Left >= w-m.Right || m.Top >= h-m.Bottom {
		return CanvasProjection{}, fmt.Errorf("folio: page setup leaves no positive content region")
	}
	for _, v := range []geom.Length{w, h, m.Top, m.Right, m.Bottom, m.Left, header, footer} {
		if v < 0 || v > geom.Length(MaxCanvasMillipoints) {
			return CanvasProjection{}, fmt.Errorf("folio: page setup exceeds the JavaScript-safe geometry bound")
		}
	}
	innerW, innerH := w-m.Left-m.Right, h-m.Top-m.Bottom
	if header < 0 || footer < 0 || header >= innerH-footer {
		return CanvasProjection{}, fmt.Errorf("folio: page setup leaves no positive content region")
	}
	preset := "custom"
	if t.doc.Page.SizeIsName {
		preset = t.doc.Page.SizeName
	}
	commandW, commandH := w, h
	if !t.doc.Page.SizeIsName {
		commandW, commandH = t.doc.Page.SizeCustom.Width, t.doc.Page.SizeCustom.Height
	}
	bands := []CanvasBand{
		{Name: "pageHeader", X: int64(m.Left), Y: int64(m.Top), Width: int64(innerW), Height: int64(header)},
		{Name: "content", X: int64(m.Left), Y: int64(m.Top + header), Width: int64(innerW), Height: int64(innerH - header - footer)},
		{Name: "pageFooter", X: int64(m.Left), Y: int64(h - m.Bottom - footer), Width: int64(innerW), Height: int64(footer)},
	}
	components, err := canvasComponents(t, bands)
	if err != nil {
		return CanvasProjection{}, err
	}
	return CanvasProjection{Width: int64(w), Height: int64(h), Orientation: t.doc.Page.Orientation, Preset: preset, MarginTop: int64(m.Top), MarginRight: int64(m.Right), MarginBottom: int64(m.Bottom), MarginLeft: int64(m.Left), GridIncrement: GridIncrement, CommandWidth: int64(commandW), CommandHeight: int64(commandH), Bands: bands, Components: components}, nil
}

// CanvasWithTextPaint returns Canvas geometry augmented with a read-only,
// production-parity text paint plan. It is session output only: it never
// mutates the template or its canonical serialization.
func CanvasWithTextPaint(t *Template, fs FontSet) (CanvasProjection, error) {
	projection, err := Canvas(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	if err := addCanvasTextPaint(t, &projection, fs); err != nil {
		return CanvasProjection{}, err
	}
	if err := addCanvasImagePaint(t, &projection); err != nil {
		return CanvasProjection{}, err
	}
	return projection, nil
}

// addCanvasImagePaint is addCanvasTextPaint's sibling for image components
// (D-5.13.2's "Producer" clause): a paint PRODUCER invoked from
// CanvasWithTextPaint, never computed inside setComponentAsset or any other
// command — every mutating command's own Canvas(t) is discarded and
// recomputed by wasm/engine.go, so the paint must be derivable from
// template state alone, exactly like text paint.
//
// It builds each run in the BAND frame (element.X/Y untranslated by band
// origin, matching this component's own X/Y) and calls the same
// resolveImagePlacement collectImageRuns/renderDocument use for the PDF —
// never a second fit computation. canvas_image_paint_test.go asserts that
// this band-relative rectangle is exactly a translation of the page-absolute
// one collectImageRuns/resolveImagePlacement produce, rather than assuming
// it from the two call sites merely sharing a function.
//
// A missing asset key or a decode failure (unrecognised media type or
// malformed bytes) leaves this component's Image field absent and does NOT
// fail the whole projection — Render (render.go) is the located, fatal
// diagnostic for a genuinely broken document; this paint-only projection
// must stay paintable (AC3: "not a crash") even for a document a save
// cannot yet produce a clean render from.
func addCanvasImagePaint(t *Template, projection *CanvasProjection) error {
	components := make(map[string]*CanvasComponent, len(projection.Components))
	for i := range projection.Components {
		component := &projection.Components[i]
		components[component.ID] = component
	}
	for _, band := range []struct {
		name     string
		elements []template.Element
	}{
		{"pageHeader", t.doc.Bands.PageHeader.Elements},
		{"content", t.doc.Bands.Content.Elements},
		{"pageFooter", t.doc.Bands.PageFooter.Elements},
	} {
		for _, element := range band.elements {
			if element.Type != template.ElementImage {
				continue
			}
			component := components[string(element.ID)]
			if component == nil || component.Band != band.name {
				return fmt.Errorf("folio: canvas image component %q is missing from geometry projection", element.ID)
			}
			if !element.Width.Set || !element.Height.Set || !element.Asset.Set || element.Asset.Null {
				// Load-time validation (parse_bands.go) already makes these
				// required for a successfully parsed document — handled
				// rather than assumed, never reached in practice.
				continue
			}
			assetKey := element.Asset.Value
			asset, ok := t.doc.Assets[assetKey]
			if !ok {
				missing := imageUnavailableMissing
				component.ImageUnavailable = &missing
				continue
			}
			raw, err := template.DecodeAssetBytes(asset)
			if err != nil {
				undecodable := imageUnavailableUndecodable
				component.ImageUnavailable = &undecodable
				continue
			}
			img, err := template.DecodeImageForRender(asset.MediaType, raw, assetKey, string(element.ID))
			if err != nil {
				undecodable := imageUnavailableUndecodable
				component.ImageUnavailable = &undecodable
				continue
			}
			run := imageRunSource{elementID: string(element.ID), assetKey: assetKey, x: element.X, y: element.Y, boxW: element.Width.Value, boxH: element.Height.Value}
			drawX, drawY, drawW, drawH := resolveImagePlacement(run, img)
			width, err := canvasDerived("image intrinsic width", geom.Length(img.Width()))
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			height, err := canvasDerived("image intrinsic height", geom.Length(img.Height()))
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			dx, err := canvasDerived("image draw x", drawX)
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			dy, err := canvasDerived("image draw y", drawY)
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			dw, err := canvasDerived("image draw width", drawW)
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			dh, err := canvasDerived("image draw height", drawH)
			if err != nil {
				return fmt.Errorf("folio: canvas image element %s: %w", element.ID, err)
			}
			component.Image = &CanvasImagePaint{
				MediaType:  asset.MediaType,
				AssetKey:   assetKey,
				Width:      int64(width),
				Height:     int64(height),
				DrawX:      int64(dx),
				DrawY:      int64(dy),
				DrawWidth:  int64(dw),
				DrawHeight: int64(dh),
			}
		}
	}
	return nil
}

func addCanvasTextPaint(t *Template, projection *CanvasProjection, fs FontSet) error {
	components := make(map[string]*CanvasComponent, len(projection.Components))
	for i := range projection.Components {
		component := &projection.Components[i]
		components[component.ID] = component
	}
	cache := newFontCache()
	for _, band := range []struct {
		name     string
		elements []template.Element
	}{
		{"pageHeader", t.doc.Bands.PageHeader.Elements},
		{"content", t.doc.Bands.Content.Elements},
		{"pageFooter", t.doc.Bands.PageFooter.Elements},
	} {
		for _, element := range band.elements {
			if element.Type != template.ElementText {
				continue
			}
			component := components[string(element.ID)]
			if component == nil || component.Band != band.name {
				return fmt.Errorf("folio: canvas text component %q is missing from geometry projection", element.ID)
			}
			chain, err := fontChain(t, element)
			if err != nil {
				// Existing designer documents can be structurally valid while
				// incomplete for production rendering (for example, a text
				// component without a chosen font chain). They remain loadable;
				// there is simply no honest measured paint to display yet.
				component.TextPaint = &CanvasTextPaint{Lines: []CanvasTextLine{}}
				continue
			}
			paint := &CanvasTextPaint{Lines: []CanvasTextLine{}}
			if !element.Value.Set || element.Value.Null || element.Value.Value == "" {
				component.TextPaint = paint
				continue
			}
			fontSize := defaultFontSizePt
			if element.Style.Set && !element.Style.Null && element.Style.Value.FontSize.Set && !element.Style.Value.FontSize.Null {
				fontSize = element.Style.Value.FontSize.Value
			}
			segs, _, err := shapeSegments(string(element.ID), chain, element.Value.Value, fs, cache)
			if err != nil {
				return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
			}
			atomic := atomicSpansFor(t.doc.UnbreakableValues, nil)
			ops := text.Opportunities(text.Dictionary(), element.Value.Value, atomic)
			boxWidth := geom.Length(0)
			if element.Width.Set && !element.Width.Null {
				boxWidth = element.Width.Value
			}
			lines := packLines(segs, ops, len([]rune(element.Value.Value)), fontSize, boxWidth)
			if len(lines) > maxCanvasTextLines {
				return fmt.Errorf("folio: canvas text element %s exceeds the line projection bound", element.ID)
			}
			_, paint.Overflow = detectWidthOverflow(string(element.ID), lines, boxWidth)
			vm, err := chainVerticalModel(chain, fontSize, fs, cache)
			if err != nil {
				return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
			}
			for i, line := range lines {
				top, err := canvasLineTop(element.Y, i, vm.Advance)
				if err != nil {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				placed, err := positionSegments(segs, line.from, line.to, element.X, top, fontSize, vm.FirstBaseline, nil)
				if err != nil {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				baseline, err := canvasDerivedSum(top, vm.FirstBaseline)
				if err != nil {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				advance, err := canvasDerived("line advance", vm.Advance)
				if err != nil {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				width, err := canvasDerived("line width", line.width)
				if err != nil {
					return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
				}
				paintLine := CanvasTextLine{Top: int64(top), Baseline: int64(baseline), Advance: int64(advance), Width: int64(width), Fragments: []CanvasTextFragment{}}
				for _, fragment := range placed {
					if len(fragment.text) > maxCanvasPropertyString || len(paintLine.Fragments) >= maxCanvasTextFragments {
						return fmt.Errorf("folio: canvas text element %s exceeds the fragment projection bound", element.ID)
					}
					x, err := canvasDerived("fragment x", fragment.x)
					if err != nil {
						return fmt.Errorf("folio: canvas text element %s: %w", element.ID, err)
					}
					paintLine.Fragments = append(paintLine.Fragments, CanvasTextFragment{Text: fragment.text, X: int64(x)})
				}
				paint.Lines = append(paint.Lines, paintLine)
			}
			component.TextPaint = paint
		}
	}
	return nil
}

func canvasDerived(name string, value geom.Length) (geom.Length, error) {
	if value < 0 || value > geom.Length(MaxCanvasMillipoints) {
		return 0, fmt.Errorf("%s exceeds the JavaScript-safe projection bound", name)
	}
	return value, nil
}

func canvasDerivedSum(left, right geom.Length) (geom.Length, error) {
	if left < 0 || right < 0 || left > geom.Length(MaxCanvasMillipoints)-right {
		return 0, fmt.Errorf("derived canvas coordinate exceeds the JavaScript-safe projection bound")
	}
	return left + right, nil
}

func canvasLineTop(elementY geom.Length, index int, advance geom.Length) (geom.Length, error) {
	if index < 0 || advance < 0 || elementY < 0 || advance > 0 && geom.Length(index) > (geom.Length(MaxCanvasMillipoints)-elementY)/advance {
		return 0, fmt.Errorf("derived canvas line origin exceeds the JavaScript-safe projection bound")
	}
	return canvasDerivedSum(elementY, geom.Length(index)*advance)
}

func canvasComponents(t *Template, bands []CanvasBand) ([]CanvasComponent, error) {
	out := make([]CanvasComponent, 0)
	for _, projected := range bands {
		var elements []template.Element
		switch projected.Name {
		case "pageHeader":
			elements = t.doc.Bands.PageHeader.Elements
		case "content":
			elements = t.doc.Bands.Content.Elements
		case "pageFooter":
			elements = t.doc.Bands.PageFooter.Elements
		}
		for _, element := range elements {
			width, height := projectedSize(element)
			for _, value := range []geom.Length{element.X, element.Y, width, height} {
				if value > geom.Length(MaxCanvasMillipoints) {
					return nil, fmt.Errorf("folio: component exceeds the JavaScript-safe geometry bound")
				}
			}
			component := CanvasComponent{ID: string(element.ID), Type: string(element.Type), Band: projected.Name, X: int64(element.X), Y: int64(element.Y), Width: int64(width), Height: int64(height), Resizable: element.Type != template.ElementTable}
			if element.Type == template.ElementText && element.Value.Set && !element.Value.Null {
				if len(element.Value.Value) > maxCanvasPropertyString {
					return nil, fmt.Errorf("folio: component value exceeds the projection bound")
				}
				component.Value = stringPointer(element.Value.Value)
				if binding := directCanvasBinding(element.Value.Value); binding != "" {
					component.Binding = stringPointer(binding)
				}
			}
			if element.VisibleIf.Set && !element.VisibleIf.Null {
				if len(element.VisibleIf.Value) > maxCanvasPropertyString {
					return nil, fmt.Errorf("folio: component visibleIf exceeds the projection bound")
				}
				component.VisibleIf = stringPointer(element.VisibleIf.Value)
			}
			if element.Type == template.ElementTable && element.Table.Set && !element.Table.Null {
				if len(element.Table.Value.Bind) > maxCanvasPropertyString {
					return nil, fmt.Errorf("folio: component table bind exceeds the projection bound")
				}
				component.TableBind = stringPointer(element.Table.Value.Bind)
			}
			if element.Style.Set && !element.Style.Null {
				if err := applyCanvasStyle(&component, element.Type, element.Style.Value); err != nil {
					return nil, err
				}
			}
			out = append(out, component)
		}
	}
	return out, nil
}

func directCanvasBinding(value string) string {
	literal, placeholders, trailing, err := expr.ScanPlaceholders(value)
	if err != nil || len(literal) != 1 || literal[0] != "" || len(placeholders) != 1 || trailing != "" || placeholders[0].Reserved {
		return ""
	}
	parsed, err := expr.Parse(placeholders[0].Inner)
	if err != nil {
		return ""
	}
	path, ok := parsed.(*expr.PathExpr)
	if !ok || path.Raw == "" || len(path.Raw) > maxCanvasBindingString {
		return ""
	}
	return path.Raw
}

func stringPointer(value string) *string     { return &value }
func boolPointer(value bool) *bool           { return &value }
func lengthPointer(value geom.Length) *int64 { rendered := int64(value); return &rendered }
func canvasPropertyLength(name string, value geom.Length) (*int64, error) {
	if value < 0 || value > geom.Length(MaxCanvasMillipoints) {
		return nil, fmt.Errorf("folio: component %s exceeds the projection bound", name)
	}
	return lengthPointer(value), nil
}

func applyCanvasStyle(component *CanvasComponent, elementType template.ElementType, style template.Style) error {
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.FontFamily.Set && !style.FontFamily.Null {
		if len(style.FontFamily.Value) > maxCanvasPropertyString {
			return fmt.Errorf("folio: component fontFamily exceeds the projection bound")
		}
		component.FontFamily = stringPointer(style.FontFamily.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.FontSize.Set && !style.FontSize.Null {
		value, err := canvasPropertyLength("fontSize", style.FontSize.Value)
		if err != nil {
			return err
		}
		component.FontSize = value
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.Bold.Set && !style.Bold.Null {
		component.Bold = boolPointer(style.Bold.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.Italic.Set && !style.Italic.Null {
		component.Italic = boolPointer(style.Italic.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.Align.Set && !style.Align.Null {
		component.Align = stringPointer(style.Align.Value)
	}
	if (elementType == template.ElementText || elementType == template.ElementTable) && style.Valign.Set && !style.Valign.Null {
		component.Valign = stringPointer(style.Valign.Value)
	}
	if style.Background.Set && !style.Background.Null {
		if len(style.Background.Value) > maxCanvasPropertyString {
			return fmt.Errorf("folio: component background exceeds the projection bound")
		}
		component.Background = stringPointer(style.Background.Value)
	}
	if style.Border.Set && !style.Border.Null {
		border := style.Border.Value
		if border.Width.Set && !border.Width.Null {
			value, err := canvasPropertyLength("borderWidth", border.Width.Value)
			if err != nil {
				return err
			}
			component.BorderWidth = value
		}
		if border.Color.Set && !border.Color.Null {
			if len(border.Color.Value) > maxCanvasPropertyString {
				return fmt.Errorf("folio: component borderColor exceeds the projection bound")
			}
			component.BorderColor = stringPointer(border.Color.Value)
		}
		if border.Edges.Set && !border.Edges.Null {
			component.BorderEdges = append([]string(nil), border.Edges.Value...)
		}
	}
	if style.Padding.Set && !style.Padding.Null {
		padding := style.Padding.Value
		if padding.Top.Set && !padding.Top.Null {
			value, err := canvasPropertyLength("paddingTop", padding.Top.Value)
			if err != nil {
				return err
			}
			component.PaddingTop = value
		}
		if padding.Right.Set && !padding.Right.Null {
			value, err := canvasPropertyLength("paddingRight", padding.Right.Value)
			if err != nil {
				return err
			}
			component.PaddingRight = value
		}
		if padding.Bottom.Set && !padding.Bottom.Null {
			value, err := canvasPropertyLength("paddingBottom", padding.Bottom.Value)
			if err != nil {
				return err
			}
			component.PaddingBottom = value
		}
		if padding.Left.Set && !padding.Left.Null {
			value, err := canvasPropertyLength("paddingLeft", padding.Left.Value)
			if err != nil {
				return err
			}
			component.PaddingLeft = value
		}
	}
	return nil
}

func canvasDimensions(t *Template) (geom.Length, geom.Length, error) {
	var width, height geom.Length
	switch {
	case t.doc.Page.SizeIsName && t.doc.Page.SizeName == "A4":
		width, height = 595276, 841890
	case t.doc.Page.SizeIsName && t.doc.Page.SizeName == "Letter":
		width, height = 612000, 792000
	case !t.doc.Page.SizeIsName:
		width, height = t.doc.Page.SizeCustom.Width, t.doc.Page.SizeCustom.Height
	default:
		return 0, 0, fmt.Errorf("folio: unsupported page size")
	}
	if t.doc.Page.Orientation == "landscape" {
		width, height = height, width
	}
	return width, height, nil
}

// ApplyPageSetupCommand decodes the one versioned, Go-defined opaque command.
// Numeric input stays a JSON literal until exact millipoint conversion; it is
// never decoded through float64.
func ApplyPageSetupCommand(t *Template, command []byte) (CanvasProjection, error) {
	if t == nil {
		return CanvasProjection{}, errNilTemplate
	}
	dec := json.NewDecoder(bytes.NewReader(command))
	dec.UseNumber()
	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil || dec.More() {
		return CanvasProjection{}, fmt.Errorf("folio: page setup command is malformed")
	}
	if len(raw) != 7 || !equalString(raw["kind"], "pageSetup") || !equalNumber(raw["version"], "1") {
		return CanvasProjection{}, fmt.Errorf("folio: unknown page setup command")
	}
	preset, orientation := stringField(raw, "preset"), stringField(raw, "orientation")
	if orientation != "portrait" && orientation != "landscape" {
		return CanvasProjection{}, fmt.Errorf("folio: page.orientation must be portrait or landscape")
	}
	if preset != "A4" && preset != "Letter" && preset != "custom" {
		return CanvasProjection{}, fmt.Errorf("folio: page.size must be A4, Letter, or custom")
	}
	var width, height geom.Length
	if preset == "custom" {
		var err error
		width, err = lengthField(raw, "width")
		if err != nil {
			return CanvasProjection{}, fmt.Errorf("folio: page.width: %w", err)
		}
		height, err = lengthField(raw, "height")
		if err != nil {
			return CanvasProjection{}, fmt.Errorf("folio: page.height: %w", err)
		}
	}
	marginRaw, ok := raw["margin"]
	if !ok {
		return CanvasProjection{}, fmt.Errorf("folio: page.margin is required")
	}
	var margins map[string]json.RawMessage
	if json.Unmarshal(marginRaw, &margins) != nil || len(margins) != 4 {
		return CanvasProjection{}, fmt.Errorf("folio: page.margin must contain top, right, bottom, left")
	}
	readMargin := func(name string) (geom.Length, error) {
		v, err := lengthField(margins, name)
		if err != nil {
			return 0, fmt.Errorf("folio: page.margin.%s: %w", name, err)
		}
		if v < 0 {
			return 0, fmt.Errorf("folio: page.margin.%s must not be negative", name)
		}
		return v, nil
	}
	top, err := readMargin("top")
	if err != nil {
		return CanvasProjection{}, err
	}
	right, err := readMargin("right")
	if err != nil {
		return CanvasProjection{}, err
	}
	bottom, err := readMargin("bottom")
	if err != nil {
		return CanvasProjection{}, err
	}
	left, err := readMargin("left")
	if err != nil {
		return CanvasProjection{}, err
	}
	if preset == "A4" {
		width, height = 595276, 841890
	} else if preset == "Letter" {
		width, height = 612000, 792000
	}
	if width <= 0 || height <= 0 {
		return CanvasProjection{}, fmt.Errorf("folio: page.size width and height must be positive")
	}
	if preset == "custom" && (width <= 0 || height <= 0) {
		return CanvasProjection{}, fmt.Errorf("folio: custom page size is required")
	}
	before, err := SerializeTemplate(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	page := &t.doc.Page
	page.Orientation = orientation
	page.Margin = template.Margin{Top: top, Right: right, Bottom: bottom, Left: left}
	if preset == "custom" {
		page.SizeIsName = false
		page.SizeName = ""
		page.SizeCustom = template.PageSize{Width: width, Height: height}
	} else {
		page.SizeIsName = true
		page.SizeName = preset
		page.SizeCustom = template.PageSize{}
	}
	projection, err := Canvas(t)
	if err != nil {
		restorePage(t, before)
		return CanvasProjection{}, err
	}
	return projection, nil
}

func restorePage(t *Template, canonical []byte) {
	restored, err := ParseTemplate(canonical)
	if err == nil {
		t.doc = restored.doc
		t.derivedFooters = restored.derivedFooters
	}
}
func equalString(raw json.RawMessage, want string) bool {
	var got string
	return json.Unmarshal(raw, &got) == nil && got == want
}
func equalNumber(raw json.RawMessage, want string) bool { return string(raw) == want }
func stringField(raw map[string]json.RawMessage, key string) string {
	var value string
	_ = json.Unmarshal(raw[key], &value)
	return value
}
func lengthField(raw map[string]json.RawMessage, key string) (geom.Length, error) {
	v, ok := raw[key]
	if !ok {
		return 0, fmt.Errorf("%s is required", key)
	}
	literal := string(v)
	if strings.ContainsAny(literal, "eE") {
		return 0, fmt.Errorf("%s must be a decimal with at most three places", key)
	}
	if dot := strings.IndexByte(literal, '.'); dot >= 0 && len(literal)-dot-1 > 3 {
		return 0, fmt.Errorf("%s must have at most three decimal places", key)
	}
	var n json.Number
	if json.Unmarshal(v, &n) != nil {
		return 0, fmt.Errorf("%s must be a finite number", key)
	}
	value, err := parseMillipoints(literal, key)
	if err != nil {
		return 0, err
	}
	if value > geom.Length(MaxCanvasMillipoints) || value < -geom.Length(MaxCanvasMillipoints) {
		return 0, fmt.Errorf("%s exceeds the JavaScript-safe geometry bound", key)
	}
	return value, nil
}
func parseMillipoints(literal, key string) (geom.Length, error) {
	negative := strings.HasPrefix(literal, "-")
	if negative {
		literal = literal[1:]
	}
	parts := strings.Split(literal, ".")
	if len(parts) > 2 || len(parts[0]) == 0 {
		return 0, fmt.Errorf("%s must be a number", key)
	}
	whole := int64(0)
	for _, c := range parts[0] {
		if c < '0' || c > '9' || whole > (1<<63-1)/10 {
			return 0, fmt.Errorf("%s overflows millipoints", key)
		}
		whole = whole*10 + int64(c-'0')
	}
	frac := int64(0)
	if len(parts) == 2 {
		for _, c := range parts[1] {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("%s must be a number", key)
			}
			frac = frac*10 + int64(c-'0')
		}
		for len(parts[1]) < 3 {
			parts[1] += "0"
			frac *= 10
		}
	}
	if whole > (1<<63-1)/1000 || (whole == (1<<63-1)/1000 && frac > (1<<63-1)%1000) {
		return 0, fmt.Errorf("%s overflows millipoints", key)
	}
	value := whole*1000 + frac
	if negative {
		value = -value
	}
	return geom.Length(value), nil
}
