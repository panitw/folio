package folio

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/template"
)

// ComponentCommandError is the stable, bounded diagnostic seam for component
// mutations. It deliberately names only a paint-safe id and command field.
type ComponentCommandError struct {
	error
	ElementID string
	DataPath  string
	Message   string
}

func componentFailure(id, path, message string) error {
	return &ComponentCommandError{error: fmt.Errorf("folio: %s", message), ElementID: id, DataPath: path, Message: message}
}

// ApplyComponentCommand applies Story 5.7's small, versioned authoring
// vocabulary. The command is intentionally decoded in Go: the browser sends
// opaque bytes and never receives the template or its canonical JSON shape.
func ApplyComponentCommand(t *Template, command []byte) (CanvasProjection, error) {
	if t == nil {
		return CanvasProjection{}, errNilTemplate
	}
	dec := json.NewDecoder(bytes.NewReader(command))
	dec.UseNumber()
	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil || dec.More() {
		return CanvasProjection{}, fmt.Errorf("folio: component command is malformed")
	}
	if !equalNumber(raw["version"], "1") {
		return CanvasProjection{}, fmt.Errorf("folio: unknown component command")
	}
	var kind string
	if json.Unmarshal(raw["kind"], &kind) != nil {
		return CanvasProjection{}, fmt.Errorf("folio: unknown component command")
	}
	switch kind {
	case "createComponent":
		return createComponent(t, raw)
	case "dropComponent":
		return dropComponent(t, raw)
	case "moveComponent":
		return moveComponent(t, raw)
	case "resizeComponent":
		return resizeComponent(t, raw)
	case "deleteComponent":
		return deleteComponent(t, raw)
	default:
		return CanvasProjection{}, fmt.Errorf("folio: unknown component command")
	}
}

func componentFields(raw map[string]json.RawMessage, want int) error {
	if len(raw) != want {
		return fmt.Errorf("folio: component command has unknown or missing fields")
	}
	return nil
}

func commandString(raw map[string]json.RawMessage, name string) (string, error) {
	v, ok := raw[name]
	if !ok {
		return "", fmt.Errorf("folio: %s is required", name)
	}
	var out string
	if json.Unmarshal(v, &out) != nil || out == "" {
		return "", fmt.Errorf("folio: %s must be a non-empty string", name)
	}
	return out, nil
}

func commandBool(raw map[string]json.RawMessage, name string) (bool, error) {
	v, ok := raw[name]
	if !ok {
		return false, fmt.Errorf("folio: %s is required", name)
	}
	var out bool
	if json.Unmarshal(v, &out) != nil {
		return false, fmt.Errorf("folio: %s must be a boolean", name)
	}
	return out, nil
}

func componentLength(raw map[string]json.RawMessage, name string, snap bool) (geom.Length, error) {
	v, err := lengthField(raw, name)
	if err != nil {
		return 0, fmt.Errorf("folio: component.%s: %w", name, err)
	}
	if snap {
		var valid bool
		v, valid = SnapToGrid(v)
		if !valid {
			return 0, fmt.Errorf("folio: component.%s overflows grid snapping", name)
		}
	}
	return v, nil
}

func commandBand(raw map[string]json.RawMessage) (string, *template.Band, error) {
	name, err := commandString(raw, "band")
	if err != nil {
		return "", nil, err
	}
	return name, nil, nil
}

func bandByName(t *Template, name string) (*template.Band, CanvasBand, error) {
	projection, err := Canvas(t)
	if err != nil {
		return nil, CanvasBand{}, err
	}
	for _, projected := range projection.Bands {
		if projected.Name != name {
			continue
		}
		switch name {
		case "pageHeader":
			return &t.doc.Bands.PageHeader, projected, nil
		case "content":
			return &t.doc.Bands.Content, projected, nil
		case "pageFooter":
			return &t.doc.Bands.PageFooter, projected, nil
		}
	}
	return nil, CanvasBand{}, fmt.Errorf("folio: component.band must be pageHeader, content, or pageFooter")
}

func createComponent(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 9); err != nil {
		return CanvasProjection{}, err
	}
	kind, err := commandString(raw, "type")
	if err != nil {
		return CanvasProjection{}, err
	}
	elementType := template.ElementType(kind)
	if elementType != template.ElementText && elementType != template.ElementImage && elementType != template.ElementTable && elementType != template.ElementLine && elementType != template.ElementRect {
		return CanvasProjection{}, fmt.Errorf("folio: component.type must be text, image, table, line, or rect")
	}
	bandName, _, err := commandBand(raw)
	if err != nil {
		return CanvasProjection{}, err
	}
	snap, err := commandBool(raw, "snap")
	if err != nil {
		return CanvasProjection{}, err
	}
	x, err := componentLength(raw, "x", snap)
	if err != nil {
		return CanvasProjection{}, err
	}
	y, err := componentLength(raw, "y", snap)
	if err != nil {
		return CanvasProjection{}, err
	}
	width, err := componentLength(raw, "width", snap)
	if err != nil {
		return CanvasProjection{}, err
	}
	height, err := componentLength(raw, "height", snap)
	if err != nil {
		return CanvasProjection{}, err
	}
	return createComponentInBand(t, elementType, bandName, x, y, width, height)
}

// dropComponent resolves a document point in Go. Band rectangles use the
// half-open convention [x, x+width) × [y, y+height): a shared boundary belongs
// to the next band, and a page edge outside the last band is rejected.
func dropComponent(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 6); err != nil {
		return CanvasProjection{}, err
	}
	kind, err := commandString(raw, "type")
	if err != nil {
		return CanvasProjection{}, err
	}
	elementType := template.ElementType(kind)
	if elementType != template.ElementText && elementType != template.ElementImage && elementType != template.ElementTable && elementType != template.ElementLine && elementType != template.ElementRect {
		return CanvasProjection{}, componentFailure("", "component.type", "component type must be text, image, table, line, or rect")
	}
	snap, err := commandBool(raw, "snap")
	if err != nil {
		return CanvasProjection{}, err
	}
	pageX, err := componentLength(raw, "x", false)
	if err != nil {
		return CanvasProjection{}, err
	}
	pageY, err := componentLength(raw, "y", false)
	if err != nil {
		return CanvasProjection{}, err
	}
	_, projected, err := hitTestBand(t, pageX, pageY)
	if err != nil {
		return CanvasProjection{}, err
	}
	x, y := pageX-geom.Length(projected.X), pageY-geom.Length(projected.Y)
	if snap {
		var valid bool
		x, valid = SnapToGrid(x)
		if !valid {
			return CanvasProjection{}, componentFailure("", "component.x", "component x overflows grid snapping")
		}
		y, valid = SnapToGrid(y)
		if !valid {
			return CanvasProjection{}, componentFailure("", "component.y", "component y overflows grid snapping")
		}
	}
	return createComponentInBand(t, elementType, projected.Name, x, y, 72000, 24000)
}

func createComponentInBand(t *Template, elementType template.ElementType, bandName string, x, y, width, height geom.Length) (CanvasProjection, error) {
	band, projected, err := bandByName(t, bandName)
	if err != nil {
		return CanvasProjection{}, err
	}
	if t.doc.NextID <= 0 || t.doc.NextID == 1<<63-1 {
		return CanvasProjection{}, fmt.Errorf("folio: nextId cannot allocate another component")
	}
	element := template.Element{ID: template.AllocateElementID(t.doc), Type: elementType, X: x, Y: y}
	if elementType == template.ElementTable {
		element.Table = template.Presence[template.TableExt]{Set: true, Value: template.TableExt{Bind: "items[]", Columns: []template.Column{}, HeaderHeight: 12000}}
		// Tables intentionally ignore free-box dimensions. Their paint size is
		// derived from the newly-created table state, never stored.
		width, height = 0, 12000
	} else {
		if width <= 0 || height <= 0 {
			return CanvasProjection{}, fmt.Errorf("folio: component.width and component.height must be positive")
		}
		element.Width = template.Presence[geom.Length]{Set: true, Value: width}
		element.Height = template.Presence[geom.Length]{Set: true, Value: height}
		if elementType == template.ElementText {
			element.Value = template.Presence[string]{Set: true, Value: "Text"}
		}
		if elementType == template.ElementImage {
			element.Asset = template.Presence[string]{Set: true, Value: ""}
		}
	}
	if err := containComponent(projected, x, y, width, height); err != nil {
		return CanvasProjection{}, componentFailure("", "component.geometry", err.Error())
	}
	band.Elements = append(band.Elements, element)
	t.doc.NextID++
	return Canvas(t)
}

func hitTestBand(t *Template, x, y geom.Length) (*template.Band, CanvasBand, error) {
	projection, err := Canvas(t)
	if err != nil {
		return nil, CanvasBand{}, err
	}
	for _, band := range projection.Bands {
		left, top := geom.Length(band.X), geom.Length(band.Y)
		if x < left || x >= left+geom.Length(band.Width) || y < top || y >= top+geom.Length(band.Height) {
			continue
		}
		switch band.Name {
		case "pageHeader":
			return &t.doc.Bands.PageHeader, band, nil
		case "content":
			return &t.doc.Bands.Content, band, nil
		case "pageFooter":
			return &t.doc.Bands.PageFooter, band, nil
		}
	}
	return nil, CanvasBand{}, componentFailure("", "component.drop", "drop point is outside a page band")
}

func findComponent(t *Template, id string) (*template.Band, CanvasBand, int, *template.Element, error) {
	for _, name := range []string{"pageHeader", "content", "pageFooter"} {
		band, projected, err := bandByName(t, name)
		if err != nil {
			return nil, CanvasBand{}, 0, nil, err
		}
		for index := range band.Elements {
			if string(band.Elements[index].ID) == id {
				return band, projected, index, &band.Elements[index], nil
			}
		}
	}
	return nil, CanvasBand{}, 0, nil, fmt.Errorf("folio: component %q was not found", id)
}

func moveComponent(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 6); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, err
	}
	snap, err := commandBool(raw, "snap")
	if err != nil {
		return CanvasProjection{}, err
	}
	x, err := componentLength(raw, "x", snap)
	if err != nil {
		return CanvasProjection{}, err
	}
	y, err := componentLength(raw, "y", snap)
	if err != nil {
		return CanvasProjection{}, err
	}
	_, projected, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "component.id", "component was not found")
	}
	width, height := projectedSize(*element)
	if err := containComponent(projected, x, y, width, height); err != nil {
		return CanvasProjection{}, componentFailure(id, "component.geometry", err.Error())
	}
	element.X, element.Y = x, y
	return Canvas(t)
}

func resizeComponent(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 6); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, err
	}
	snap, err := commandBool(raw, "snap")
	if err != nil {
		return CanvasProjection{}, err
	}
	width, err := componentLength(raw, "width", snap)
	if err != nil {
		return CanvasProjection{}, err
	}
	height, err := componentLength(raw, "height", snap)
	if err != nil {
		return CanvasProjection{}, err
	}
	_, projected, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "component.id", "component was not found")
	}
	if element.Type == template.ElementTable {
		return CanvasProjection{}, componentFailure(id, "component.geometry", "table has derived geometry and cannot be resized")
	}
	if width <= 0 || height <= 0 {
		return CanvasProjection{}, componentFailure(id, "component.geometry", "component width and height must be positive")
	}
	if err := containComponent(projected, element.X, element.Y, width, height); err != nil {
		return CanvasProjection{}, componentFailure(id, "component.geometry", err.Error())
	}
	element.Width = template.Presence[geom.Length]{Set: true, Value: width}
	element.Height = template.Presence[geom.Length]{Set: true, Value: height}
	return Canvas(t)
}

func deleteComponent(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 3); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, err
	}
	band, _, index, _, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "component.id", "component was not found")
	}
	band.Elements = append(band.Elements[:index:index], band.Elements[index+1:]...)
	return Canvas(t)
}

func projectedSize(element template.Element) (geom.Length, geom.Length) {
	if element.Type != template.ElementTable {
		return element.Width.Value, element.Height.Value
	}
	var width geom.Length
	for _, column := range element.Table.Value.Columns {
		width += column.Width
	}
	return width, element.Table.Value.HeaderHeight
}

func containComponent(band CanvasBand, x, y, width, height geom.Length) error {
	if x < 0 || y < 0 || width < 0 || height < 0 || x > geom.Length(band.Width) || y > geom.Length(band.Height) || width > geom.Length(band.Width)-x || height > geom.Length(band.Height)-y {
		return fmt.Errorf("folio: component geometry must stay within %s", band.Name)
	}
	return nil
}
