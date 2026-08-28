package folio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/template"
)

// GridIncrement is the fixed six-point grid used by the designer projection.
// SnapNearest's documented midpoint rule is away from zero.
const GridIncrement int64 = 6000

// MaxCanvasMillipoints keeps every document value emitted to the JSON/JS paint
// boundary within Number.MAX_SAFE_INTEGER. The page-setup command has the same
// bound, so a successful command can never strand the worker with an
// unrepresentable projection.
const MaxCanvasMillipoints int64 = 9007199254740991

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
}
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
			out = append(out, CanvasComponent{ID: string(element.ID), Type: string(element.Type), Band: projected.Name, X: int64(element.X), Y: int64(element.Y), Width: int64(width), Height: int64(height), Resizable: element.Type != template.ElementTable})
		}
	}
	return out, nil
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
