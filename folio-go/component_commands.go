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
	case "duplicateComponent":
		return duplicateComponent(t, raw)
	case "updateComponentProperties":
		return updateComponentProperties(t, raw)
	default:
		return CanvasProjection{}, fmt.Errorf("folio: unknown component command")
	}
}

// updateComponentProperties is deliberately a small closed mutation language.
// It applies the supplied changes to every named component as one candidate;
// the engine's serialize/reparse transaction makes the update atomic.
func updateComponentProperties(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	before, err := SerializeTemplate(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	working, err := ParseTemplate(before)
	if err != nil {
		return CanvasProjection{}, err
	}
	projection, err := updateComponentPropertiesInPlace(working, raw)
	if err != nil {
		return CanvasProjection{}, err
	}
	// Keep the public helper transactional too. wasm.Apply uses a fresh clone,
	// but direct callers must receive the same no-partial-mutation guarantee.
	canonical, err := SerializeTemplate(working)
	if err != nil {
		return CanvasProjection{}, err
	}
	installed, err := ParseTemplate(canonical)
	if err != nil {
		return CanvasProjection{}, componentFailure("", "component.changes", "component properties did not pass format validation")
	}
	t.doc, t.derivedFooters = installed.doc, installed.derivedFooters
	return projection, nil
}

func updateComponentPropertiesInPlace(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 4); err != nil {
		return CanvasProjection{}, err
	}
	idsRaw, ok := raw["ids"]
	if !ok {
		return CanvasProjection{}, componentFailure("", "component.ids", "component ids are required")
	}
	var ids []string
	if json.Unmarshal(idsRaw, &ids) != nil || len(ids) == 0 {
		return CanvasProjection{}, componentFailure("", "component.ids", "component ids must be a non-empty string array")
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			return CanvasProjection{}, componentFailure(id, "component.ids", "component ids must be unique non-empty strings")
		}
		seen[id] = true
	}
	changesRaw, ok := raw["changes"]
	if !ok {
		return CanvasProjection{}, componentFailure("", "component.changes", "component changes are required")
	}
	var changes map[string]json.RawMessage
	if json.Unmarshal(changesRaw, &changes) != nil || len(changes) == 0 {
		return CanvasProjection{}, componentFailure("", "component.changes", "component changes must be a non-empty object")
	}
	if len(ids) > 1 {
		if _, ok := changes["value"]; ok {
			return CanvasProjection{}, componentFailure("", "component.value", "text value cannot be edited across a selection")
		}
	}
	for _, id := range ids {
		_, band, _, element, err := findComponent(t, id)
		if err != nil {
			return CanvasProjection{}, componentFailure(id, "component.id", "component was not found")
		}
		if err := applyPropertyChanges(t, element, changes); err != nil {
			return CanvasProjection{}, componentFailure(id, "component."+propertyPath(changes), err.Error())
		}
		width, height := projectedSize(*element)
		if err := containComponent(band, element.X, element.Y, width, height); err != nil {
			return CanvasProjection{}, componentFailure(id, "component.geometry", err.Error())
		}
	}
	return Canvas(t)
}

func propertyPath(changes map[string]json.RawMessage) string {
	// This is a fixed command vocabulary, so use its canonical order rather
	// than ranging a map (diagnostic location must be repeatable too).
	for _, key := range []string{"x", "y", "width", "height", "value", "visibleIf", "fontFamily", "fontSize", "bold", "italic", "align", "valign", "background", "borderWidth", "borderColor", "borderEdges", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft"} {
		if _, ok := changes[key]; ok {
			return key
		}
	}
	return "changes"
}

func propertyChange(raw json.RawMessage) (string, json.RawMessage, error) {
	var value map[string]json.RawMessage
	if json.Unmarshal(raw, &value) != nil || (len(value) != 1 && len(value) != 2) {
		return "", nil, fmt.Errorf("property change must be an operation object")
	}
	op, ok := value["op"]
	var operation string
	if !ok || json.Unmarshal(op, &operation) != nil || (operation != "set" && operation != "clear" && operation != "null") {
		return "", nil, fmt.Errorf("property operation must be set, clear, or null")
	}
	if operation == "clear" || operation == "null" {
		if len(value) != 1 {
			return "", nil, fmt.Errorf("clear property operation cannot carry a value")
		}
		return operation, nil, nil
	}
	v, ok := value["value"]
	if !ok || len(value) != 2 {
		return "", nil, fmt.Errorf("set property operation requires exactly one value")
	}
	return operation, v, nil
}

func propertyString(raw json.RawMessage) (string, error) {
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return "", fmt.Errorf("property value must be a string")
	}
	return value, nil
}
func propertyBool(raw json.RawMessage) (bool, error) {
	var value bool
	if json.Unmarshal(raw, &value) != nil {
		return false, fmt.Errorf("property value must be a boolean")
	}
	return value, nil
}
func propertyLength(raw json.RawMessage, key string) (geom.Length, error) {
	return lengthField(map[string]json.RawMessage{key: raw}, key)
}
func styleFor(element *template.Element) *template.Style {
	if !element.Style.Set || element.Style.Null {
		element.Style = template.Presence[template.Style]{Set: true}
	}
	return &element.Style.Value
}

func applyPropertyChanges(t *Template, element *template.Element, changes map[string]json.RawMessage) error {
	allowed := map[string]bool{"x": true, "y": true, "visibleIf": true}
	propertyOrder := []string{"x", "y", "width", "height", "value", "visibleIf", "fontFamily", "fontSize", "bold", "italic", "align", "valign", "background", "borderWidth", "borderColor", "borderEdges", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft"}
	if element.Type != template.ElementTable {
		allowed["width"], allowed["height"] = true, true
	}
	if element.Type == template.ElementText {
		allowed["value"] = true
	}
	if element.Type == template.ElementText || element.Type == template.ElementImage || element.Type == template.ElementTable || element.Type == template.ElementLine || element.Type == template.ElementRect {
		for _, key := range []string{"background", "borderWidth", "borderColor", "borderEdges", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft"} {
			allowed[key] = true
		}
	}
	if element.Type == template.ElementText || element.Type == template.ElementTable {
		for _, key := range []string{"fontFamily", "fontSize", "bold", "italic", "align", "valign"} {
			allowed[key] = true
		}
	}
	known := 0
	for _, key := range propertyOrder {
		if _, ok := changes[key]; ok {
			known++
		}
	}
	if known != len(changes) {
		return fmt.Errorf("property is not editable")
	}
	for _, key := range propertyOrder {
		change, present := changes[key]
		if !present {
			continue
		}
		if !allowed[key] {
			return fmt.Errorf("property %s is not editable for %s", key, element.Type)
		}
		op, value, err := propertyChange(change)
		if err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
		clear := op == "clear"
		setNull := op == "null"
		switch key {
		case "x", "y", "width", "height", "fontSize", "borderWidth", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft":
			if setNull {
				return fmt.Errorf("%s does not support null", key)
			}
			if clear && (key == "x" || key == "y" || key == "width" || key == "height") {
				return fmt.Errorf("%s cannot be cleared", key)
			}
			var length geom.Length
			if !clear {
				length, err = propertyLength(value, key)
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				if (key == "width" || key == "height" || key == "fontSize" || key == "borderWidth") && length <= 0 {
					return fmt.Errorf("%s must be positive", key)
				}
				if stringsContainsPlaceholder(string(value)) {
					return fmt.Errorf("%s must not contain a placeholder", key)
				}
			}
			switch key {
			case "x":
				element.X = length
			case "y":
				element.Y = length
			case "width":
				element.Width = template.Presence[geom.Length]{Set: true, Value: length}
			case "height":
				element.Height = template.Presence[geom.Length]{Set: true, Value: length}
			case "fontSize":
				st := styleFor(element)
				if clear {
					st.FontSize = template.Presence[geom.Length]{}
				} else {
					st.FontSize = template.Presence[geom.Length]{Set: true, Value: length}
				}
			case "borderWidth":
				st := styleFor(element)
				if !st.Border.Set || st.Border.Null {
					st.Border = template.Presence[template.Border]{Set: true}
				}
				if clear {
					st.Border.Value.Width = template.Presence[geom.Length]{}
				} else {
					st.Border.Value.Width = template.Presence[geom.Length]{Set: true, Value: length}
				}
			default:
				st := styleFor(element)
				if !st.Padding.Set || st.Padding.Null {
					st.Padding = template.Presence[template.Padding]{Set: true}
				}
				target := map[string]*template.Presence[geom.Length]{"paddingTop": &st.Padding.Value.Top, "paddingRight": &st.Padding.Value.Right, "paddingBottom": &st.Padding.Value.Bottom, "paddingLeft": &st.Padding.Value.Left}[key]
				if clear {
					*target = template.Presence[geom.Length]{}
				} else {
					*target = template.Presence[geom.Length]{Set: true, Value: length}
				}
			}
		case "value", "visibleIf", "fontFamily", "align", "valign", "background", "borderColor":
			var text string
			if !clear && !setNull {
				text, err = propertyString(value)
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				if key != "value" && stringsContainsPlaceholder(text) {
					return fmt.Errorf("%s must not contain a placeholder", key)
				}
			}
			switch key {
			case "value":
				if clear || setNull {
					return fmt.Errorf("value cannot be cleared")
				}
				element.Value = template.Presence[string]{Set: true, Value: text}
			case "visibleIf":
				if clear {
					element.VisibleIf = template.Presence[string]{}
				} else if setNull {
					element.VisibleIf = template.Presence[string]{Set: true, Null: true}
				} else {
					element.VisibleIf = template.Presence[string]{Set: true, Value: text}
				}
			case "fontFamily":
				if setNull {
					return fmt.Errorf("fontFamily does not support null")
				}
				if !clear && !knownFontFamily(t, text) {
					return fmt.Errorf("fontFamily must name a declared non-empty font chain")
				}
				st := styleFor(element)
				if clear {
					st.FontFamily = template.Presence[string]{}
				} else {
					st.FontFamily = template.Presence[string]{Set: true, Value: text}
				}
			case "align":
				if setNull {
					return fmt.Errorf("align does not support null")
				}
				st := styleFor(element)
				if clear {
					st.Align = template.Presence[string]{}
				} else {
					st.Align = template.Presence[string]{Set: true, Value: text}
				}
			case "valign":
				if setNull {
					return fmt.Errorf("valign does not support null")
				}
				st := styleFor(element)
				if clear {
					st.Valign = template.Presence[string]{}
				} else {
					st.Valign = template.Presence[string]{Set: true, Value: text}
				}
			case "background":
				st := styleFor(element)
				if clear {
					st.Background = template.Presence[string]{}
				} else if setNull {
					st.Background = template.Presence[string]{Set: true, Null: true}
				} else if !validPropertyColor(text) {
					return fmt.Errorf("background must be a #RRGGBB colour")
				} else {
					st.Background = template.Presence[string]{Set: true, Value: text}
				}
			case "borderColor":
				if setNull {
					return fmt.Errorf("borderColor does not support null")
				}
				if !clear && !validPropertyColor(text) {
					return fmt.Errorf("borderColor must be a #RRGGBB colour")
				}
				st := styleFor(element)
				if !st.Border.Set || st.Border.Null {
					st.Border = template.Presence[template.Border]{Set: true}
				}
				if clear {
					st.Border.Value.Color = template.Presence[string]{}
				} else {
					st.Border.Value.Color = template.Presence[string]{Set: true, Value: text}
				}
			}
		case "bold", "italic":
			if setNull {
				return fmt.Errorf("%s does not support null", key)
			}
			if clear {
				if key == "bold" {
					styleFor(element).Bold = template.Presence[bool]{}
				} else {
					styleFor(element).Italic = template.Presence[bool]{}
				}
				continue
			}
			flag, err := propertyBool(value)
			if err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
			if key == "bold" {
				styleFor(element).Bold = template.Presence[bool]{Set: true, Value: flag}
			} else {
				styleFor(element).Italic = template.Presence[bool]{Set: true, Value: flag}
			}
		case "borderEdges":
			if setNull {
				return fmt.Errorf("borderEdges does not support null")
			}
			if clear {
				st := styleFor(element)
				if st.Border.Set && !st.Border.Null {
					st.Border.Value.Edges = template.Presence[[]string]{}
				}
				continue
			}
			var edges []string
			if json.Unmarshal(value, &edges) != nil || len(edges) == 0 {
				return fmt.Errorf("borderEdges must be a non-empty string array")
			}
			st := styleFor(element)
			if !st.Border.Set || st.Border.Null {
				st.Border = template.Presence[template.Border]{Set: true}
			}
			st.Border.Value.Edges = template.Presence[[]string]{Set: true, Value: edges}
		}
	}
	cleanupEmptyStyle(element)
	return nil
}

func validPropertyColor(value string) bool {
	_, ok := parseHexColor(value)
	return ok
}

func knownFontFamily(t *Template, value string) bool {
	chain, ok := t.doc.Fonts[value]
	return ok && len(chain) > 0
}

func cleanupEmptyStyle(element *template.Element) {
	if !element.Style.Set || element.Style.Null {
		return
	}
	style := &element.Style.Value
	if style.Border.Set && !style.Border.Null {
		border := style.Border.Value
		if !border.Color.Set && !border.Width.Set && !border.Edges.Set && len(border.Extra) == 0 {
			style.Border = template.Presence[template.Border]{}
		}
	}
	if style.Padding.Set && !style.Padding.Null {
		padding := style.Padding.Value
		if !padding.Top.Set && !padding.Right.Set && !padding.Bottom.Set && !padding.Left.Set && len(padding.Extra) == 0 {
			style.Padding = template.Presence[template.Padding]{}
		}
	}
	if !style.Align.Set && !style.Background.Set && !style.Bold.Set && !style.Italic.Set && !style.Border.Set && !style.FontFamily.Set && !style.FontSize.Set && !style.Padding.Set && !style.Valign.Set && len(style.Extra) == 0 {
		element.Style = template.Presence[template.Style]{}
	}
}

func stringsContainsPlaceholder(value string) bool {
	return bytes.Contains([]byte(value), []byte("{{")) || bytes.Contains([]byte(value), []byte("}}"))
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

func duplicateComponent(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 4); err != nil {
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
	band, projected, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "component.id", "component was not found")
	}
	if t.doc.NextID <= 0 || t.doc.NextID == 1<<63-1 {
		return CanvasProjection{}, fmt.Errorf("folio: nextId cannot allocate another component")
	}
	clone := *element
	clone.ID = template.AllocateElementID(t.doc)
	width, height := projectedSize(clone)
	x, y := clone.X+6000, clone.Y+6000
	if snap {
		x, _ = SnapToGrid(x)
		y, _ = SnapToGrid(y)
	}
	if containComponent(projected, x, y, width, height) != nil {
		x, y = clone.X, clone.Y
	}
	clone.X, clone.Y = x, y
	band.Elements = append(band.Elements, clone)
	t.doc.NextID++
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
