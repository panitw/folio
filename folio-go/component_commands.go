package folio

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/panitw/folio/folio-go/internal/expr"
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
	if err := dec.Decode(&raw); err != nil {
		return CanvasProjection{}, fmt.Errorf("folio: component command is malformed")
	}
	var surplus any
	if err := dec.Decode(&surplus); err != io.EOF {
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
	case "setComponentAsset":
		return setComponentAsset(t, raw)
	case "bindComponentScalar":
		return bindComponentScalar(t, raw)
	case "addTableColumn":
		return applyTableColumnCommand(t, raw, addTableColumn)
	case "removeTableColumn":
		return applyTableColumnCommand(t, raw, removeTableColumn)
	case "moveTableColumn":
		return applyTableColumnCommand(t, raw, moveTableColumn)
	case "updateTableColumn":
		return applyTableColumnCommand(t, raw, updateTableColumn)
	case "configureTableBinding":
		return applyTableColumnCommand(t, raw, configureTableBinding)
	case "updateTableColumnBinding":
		return applyTableColumnCommand(t, raw, updateTableColumnBinding)
	case "updateTableColumnFooter":
		return applyTableColumnCommand(t, raw, updateTableColumnFooter)
	default:
		return CanvasProjection{}, fmt.Errorf("folio: unknown component command")
	}
}

// applyTableColumnCommand keeps the public command seam just as atomic as
// wasm.Engine.Apply. The individual handlers may mutate their candidate while
// checking geometry, but the caller's template is installed only after that
// candidate serializes, reparses, and projects successfully.
func applyTableColumnCommand(t *Template, raw map[string]json.RawMessage, apply func(*Template, map[string]json.RawMessage) (CanvasProjection, error)) (CanvasProjection, error) {
	before, err := SerializeTemplate(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	working, err := ParseTemplate(before)
	if err != nil {
		return CanvasProjection{}, err
	}
	if _, err := apply(working, raw); err != nil {
		return CanvasProjection{}, err
	}
	canonical, err := SerializeTemplate(working)
	if err != nil {
		return CanvasProjection{}, err
	}
	installed, err := ParseTemplate(canonical)
	if err != nil {
		return CanvasProjection{}, componentFailure("", "table.columns", "table columns did not pass format validation")
	}
	projection, err := Canvas(installed)
	if err != nil {
		return CanvasProjection{}, err
	}
	t.doc, t.derivedFooters = installed.doc, installed.derivedFooters
	return projection, nil
}

// The table commands are a deliberately closed authoring vocabulary. Sample
// input never enters these commands: it only helps the UI discover candidates.
func addTableColumn(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 4); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, componentFailure("", "table.id", err.Error())
	}
	index, err := commandInt(raw, "index")
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "column.index", err.Error())
	}
	_, band, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "table.id", "table was not found")
	}
	if element.Type != template.ElementTable || !element.Table.Set || element.Table.Null {
		return CanvasProjection{}, componentFailure(id, "table.id", "component is not a table")
	}
	columns := element.Table.Value.Columns
	if len(columns) >= maxTableColumns {
		return CanvasProjection{}, componentFailure(id, "column.index", "table has too many columns")
	}
	if index < 0 || index > len(columns) {
		return CanvasProjection{}, componentFailure(id, "column.index", "column index is out of range")
	}
	if t.doc.NextID <= 0 || t.doc.NextID == 1<<63-1 {
		return CanvasProjection{}, componentFailure(id, "column.id", "nextId cannot allocate another column")
	}
	column := template.Column{ID: template.AllocateElementID(t.doc), Label: fmt.Sprintf("Column %d", len(columns)+1), Width: geom.Length(72000)}
	element.Table.Value.Columns = append(columns, template.Column{})
	copy(element.Table.Value.Columns[index+1:], element.Table.Value.Columns[index:])
	element.Table.Value.Columns[index] = column
	width, height := projectedSize(*element)
	if err := containComponent(band, element.X, element.Y, width, height); err != nil {
		return CanvasProjection{}, componentFailure(id, "column.width", err.Error())
	}
	t.doc.NextID++
	return Canvas(t)
}

func removeTableColumn(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 4); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, componentFailure("", "table.id", err.Error())
	}
	columnID, err := commandString(raw, "columnId")
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "column.id", err.Error())
	}
	_, _, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "table.id", "table was not found")
	}
	if element.Type != template.ElementTable || !element.Table.Set || element.Table.Null {
		return CanvasProjection{}, componentFailure(id, "table.id", "component is not a table")
	}
	index := tableColumnIndex(element, columnID)
	if index < 0 {
		return CanvasProjection{}, componentFailure(id, "column.id", "column was not found")
	}
	columns := element.Table.Value.Columns
	copy(columns[index:], columns[index+1:])
	element.Table.Value.Columns = columns[:len(columns)-1]
	return Canvas(t)
}

func moveTableColumn(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 5); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, componentFailure("", "table.id", err.Error())
	}
	columnID, err := commandString(raw, "columnId")
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "column.id", err.Error())
	}
	toIndex, err := commandInt(raw, "toIndex")
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "column.toIndex", err.Error())
	}
	_, _, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "table.id", "table was not found")
	}
	if element.Type != template.ElementTable || !element.Table.Set || element.Table.Null {
		return CanvasProjection{}, componentFailure(id, "table.id", "component is not a table")
	}
	fromIndex := tableColumnIndex(element, columnID)
	if fromIndex < 0 {
		return CanvasProjection{}, componentFailure(id, "column.id", "column was not found")
	}
	columns := element.Table.Value.Columns
	if toIndex < 0 || toIndex >= len(columns) {
		return CanvasProjection{}, componentFailure(id, "column.toIndex", "column index is out of range")
	}
	if fromIndex == toIndex {
		return Canvas(t)
	}
	column := columns[fromIndex]
	if fromIndex < toIndex {
		copy(columns[fromIndex:toIndex], columns[fromIndex+1:toIndex+1])
	} else {
		copy(columns[toIndex+1:fromIndex+1], columns[toIndex:fromIndex])
	}
	columns[toIndex] = column
	return Canvas(t)
}

func updateTableColumn(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 6); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, componentFailure("", "table.id", err.Error())
	}
	columnID, err := commandString(raw, "columnId")
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "column.id", err.Error())
	}
	field, err := commandString(raw, "field")
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "column.field", err.Error())
	}
	value, ok := raw["value"]
	if !ok {
		return CanvasProjection{}, componentFailure(id, "column.value", "column value is required")
	}
	_, band, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "table.id", "table was not found")
	}
	if element.Type != template.ElementTable || !element.Table.Set || element.Table.Null {
		return CanvasProjection{}, componentFailure(id, "table.id", "component is not a table")
	}
	index := tableColumnIndex(element, columnID)
	if index < 0 {
		return CanvasProjection{}, componentFailure(id, "column.id", "column was not found")
	}
	column := &element.Table.Value.Columns[index]
	switch field {
	case "header":
		label, err := commandString(map[string]json.RawMessage{"value": value}, "value")
		if err != nil || len(label) > 256 {
			return CanvasProjection{}, componentFailure(id, "column.header", "header must be a bounded string")
		}
		column.Label = label
	case "width":
		width, err := propertyLength(value, "width")
		if err != nil || width <= 0 {
			return CanvasProjection{}, componentFailure(id, "column.width", "width must be a positive length")
		}
		column.Width = width
	case "align":
		align, err := commandString(map[string]json.RawMessage{"value": value}, "value")
		if err != nil || (align != "left" && align != "center" && align != "right") {
			return CanvasProjection{}, componentFailure(id, "column.align", "alignment must be left, center, or right")
		}
		column.Align = template.Presence[string]{Set: true, Value: align}
	default:
		return CanvasProjection{}, componentFailure(id, "column.field", "column field is not editable")
	}
	width, height := projectedSize(*element)
	if err := containComponent(band, element.X, element.Y, width, height); err != nil {
		return CanvasProjection{}, componentFailure(id, "column.width", err.Error())
	}
	return Canvas(t)
}

var rootCollectionPath = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*\[\]$`)
var boundedIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var rootValuePath = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*$`)

// configureTableBinding changes the two document-owned row-scope settings as
// one candidate. An empty alias deliberately means the schema's absent `as`
// form; render resolution supplies the established default alias, `row`.
func configureTableBinding(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 5); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, componentFailure("", "table.id", err.Error())
	}
	collection, err := commandString(raw, "collection")
	if err != nil || len(collection) > 256 || !rootCollectionPath.MatchString(collection) || strings.HasPrefix(collection, "params.") || collection == "params[]" {
		return CanvasProjection{}, componentFailure(id, "table.collection", "collection must be a bounded root collection path ending in []")
	}
	aliasRaw, ok := raw["alias"]
	if !ok {
		return CanvasProjection{}, componentFailure(id, "table.alias", "alias is required")
	}
	var alias string
	if json.Unmarshal(aliasRaw, &alias) != nil || len(alias) > 64 || (alias != "" && (!boundedIdentifier.MatchString(alias) || reservedRowAlias(alias))) {
		return CanvasProjection{}, componentFailure(id, "table.alias", "alias must be a bounded identifier or empty for row")
	}
	_, _, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "table.id", "table was not found")
	}
	if element.Type != template.ElementTable || !element.Table.Set || element.Table.Null {
		return CanvasProjection{}, componentFailure(id, "table.id", "component is not a table")
	}
	oldAlias := resolvedTableAlias(element.Table.Value.As)
	newAlias := alias
	if newAlias == "" {
		newAlias = "row"
	}
	if oldAlias != newAlias {
		for i := range element.Table.Value.Columns {
			next, migrated, used, migrationErr := expr.RewriteRowBinding(element.Table.Value.Columns[i].Bind, oldAlias, newAlias)
			if migrationErr != nil || (used && !migrated) {
				return CanvasProjection{}, componentFailure(id, "table.alias", "alias change cannot migrate a row-scoped column binding")
			}
			if migrated {
				element.Table.Value.Columns[i].Bind = next
			}
		}
	}
	element.Table.Value.Bind = collection
	if alias == "" {
		element.Table.Value.As = template.Presence[string]{}
	} else {
		element.Table.Value.As = template.Presence[string]{Set: true, Value: alias}
	}
	return Canvas(t)
}

// updateTableColumnBinding accepts only a canonical single row-relative field
// expression. The UI passes a discovered field path, while this Go boundary
// constructs and owns the actual expression spelling.
func updateTableColumnBinding(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 5); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, componentFailure("", "table.id", err.Error())
	}
	columnID, err := commandString(raw, "columnId")
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "column.id", err.Error())
	}
	field, err := commandString(raw, "field")
	if err != nil || len(field) > 192 || !rootValuePath.MatchString(field) {
		return CanvasProjection{}, componentFailure(id, "column.bind", "field must be a bounded row field path")
	}
	_, _, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "table.id", "table was not found")
	}
	if element.Type != template.ElementTable || !element.Table.Set || element.Table.Null {
		return CanvasProjection{}, componentFailure(id, "table.id", "component is not a table")
	}
	index := tableColumnIndex(element, columnID)
	if index < 0 {
		return CanvasProjection{}, componentFailure(id, "column.id", "column was not found")
	}
	alias := "row"
	if element.Table.Value.As.Set && !element.Table.Value.As.Null {
		alias = element.Table.Value.As.Value
	}
	element.Table.Value.Columns[index].Bind = "{{" + alias + "." + field + "}}"
	return Canvas(t)
}

func reservedRowAlias(alias string) bool {
	return alias == "params" || alias == "page" || alias == "pages"
}

func resolvedTableAlias(value template.Presence[string]) string {
	if value.Set && !value.Null && value.Value != "" {
		return value.Value
	}
	return "row"
}

// updateTableColumnFooter is intentionally a complete footer configuration,
// not three independent mutations. Empty companion strings mean absent schema
// fields, making an accepted command one revision/history step.
func updateTableColumnFooter(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 7); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, componentFailure("", "table.id", err.Error())
	}
	columnID, err := commandString(raw, "columnId")
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "column.id", err.Error())
	}
	footer, ok := optionalCommandString(raw, "footer", 16)
	if !ok || (footer != "" && footer != "sum" && footer != "avg" && footer != "count") {
		return CanvasProjection{}, componentFailure(id, "column.footer", "footer must be sum, avg, count, or empty")
	}
	footerOf, ok := optionalCommandString(raw, "footerOf", 256)
	if !ok || (footerOf != "" && !rootValuePath.MatchString(footerOf)) {
		return CanvasProjection{}, componentFailure(id, "column.footerOf", "footerOf must be a bounded root data path")
	}
	footerFormat, ok := optionalCommandString(raw, "footerFormat", 256)
	if !ok {
		return CanvasProjection{}, componentFailure(id, "column.footerFormat", "footerFormat must be a bounded string")
	}
	if footer == "" && (footerOf != "" || footerFormat != "") {
		return CanvasProjection{}, componentFailure(id, "column.footer", "footer companions require a footer")
	}
	if footer == "count" && footerOf != "" {
		return CanvasProjection{}, componentFailure(id, "column.footerOf", "count uses the table collection and forbids footerOf")
	}
	_, _, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "table.id", "table was not found")
	}
	if element.Type != template.ElementTable || !element.Table.Set || element.Table.Null {
		return CanvasProjection{}, componentFailure(id, "table.id", "component is not a table")
	}
	index := tableColumnIndex(element, columnID)
	if index < 0 {
		return CanvasProjection{}, componentFailure(id, "column.id", "column was not found")
	}
	collection := strings.TrimSuffix(element.Table.Value.Bind, "[]")
	if footerOf != "" && !strings.HasPrefix(footerOf, collection+".") {
		return CanvasProjection{}, componentFailure(id, "column.footerOf", "footerOf must stay within the table collection")
	}
	column := &element.Table.Value.Columns[index]
	column.Footer, column.FooterOf, column.FooterFormat = template.Presence[string]{}, template.Presence[string]{}, template.Presence[string]{}
	if footer != "" {
		column.Footer = template.Presence[string]{Set: true, Value: footer}
	}
	if footerOf != "" {
		column.FooterOf = template.Presence[string]{Set: true, Value: footerOf}
	}
	if footerFormat != "" {
		column.FooterFormat = template.Presence[string]{Set: true, Value: footerFormat}
	}
	return Canvas(t)
}

func optionalCommandString(raw map[string]json.RawMessage, name string, max int) (string, bool) {
	v, ok := raw[name]
	if !ok {
		return "", false
	}
	var out string
	if json.Unmarshal(v, &out) != nil || len(out) > max {
		return "", false
	}
	return out, true
}

func tableColumnIndex(element *template.Element, columnID string) int {
	for index, column := range element.Table.Value.Columns {
		if string(column.ID) == columnID {
			return index
		}
	}
	return -1
}

func commandInt(raw map[string]json.RawMessage, name string) (int, error) {
	value, ok := raw[name]
	if !ok {
		return 0, fmt.Errorf("%s is required", name)
	}
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err != nil || decoder.More() {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	integer, err := number.Int64()
	if err != nil || int64(int(integer)) != integer {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return int(integer), nil
}

// bindComponentScalar is the sole Story 6.2 mutation for a picked root data
// path. The caller transports JSON object-key segments, not an expression or a
// browser-side validity judgment. This command owns the conversion to Folio's
// established expression grammar and rejects every non-root/reserved form
// before touching the target element.
func bindComponentScalar(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 4); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, componentFailure("", "component.id", err.Error())
	}
	segmentsRaw, ok := raw["segments"]
	if !ok {
		return CanvasProjection{}, componentFailure(id, "binding.segments", "binding segments are required")
	}
	var segments []string
	if json.Unmarshal(segmentsRaw, &segments) != nil || len(segments) == 0 || len(segments) > 32 {
		return CanvasProjection{}, componentFailure(id, "binding.segments", "binding segments must be a non-empty bounded string array")
	}
	for _, segment := range segments {
		if segment == "" || len(segment) > 64 {
			return CanvasProjection{}, componentFailure(id, "binding.segments", "binding segments must be bounded non-empty strings")
		}
	}
	if segments[0] == "params" {
		return CanvasProjection{}, componentFailure(id, "binding.segments", "params is not a root data binding")
	}
	path := strings.Join(segments, ".")
	if len(path) > maxCanvasBindingString {
		return CanvasProjection{}, componentFailure(id, "binding.segments", "binding path exceeds the projection bound")
	}
	parsed, err := expr.Parse(path)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "binding.segments", "binding path is not a valid Folio expression")
	}
	pathExpr, ok := parsed.(*expr.PathExpr)
	// Joining is only an intermediate representation for Folio's established
	// identifier grammar. It must never reinterpret a decoded JSON key such as
	// "a.b" as two keys. Keys that Folio cannot represent are rejected before
	// mutation, rather than silently binding a different path.
	if !ok || !sameSegments(pathExpr.Segments, segments) || expr.IsReserved(path) {
		return CanvasProjection{}, componentFailure(id, "binding.segments", "binding path must be a non-reserved root data path")
	}
	if err := expr.Check(parsed); err != nil {
		return CanvasProjection{}, componentFailure(id, "binding.segments", "binding path is not a valid Folio expression")
	}
	_, _, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "component.id", "component was not found")
	}
	if element.Type != template.ElementText {
		return CanvasProjection{}, componentFailure(id, "component.id", "only text components can receive a scalar binding")
	}
	// The generated expression is canonical and then independently reparsed by
	// wasm.Engine before installation. No sample bytes or local tree metadata
	// enter the template.
	element.Value = template.Presence[string]{Set: true, Value: "{{" + path + "}}"}
	return Canvas(t)
}

func sameSegments(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
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

// engineProtocolMaxPayloadBytes mirrors MAX_ENGINE_PAYLOAD_BYTES
// (folio-designer/src/engine-protocol.ts) — the one number the transport
// enforces on every worker request payload, asset commands included.
const engineProtocolMaxPayloadBytes = 8 * 1024 * 1024

// maxComponentAssetPayloadOverheadBytes reserves room, inside the envelope
// above, for setComponentAssetCommand's OWN JSON skeleton around its
// base64 "data" field — the "kind"/"version"/"id"/"mediaType" keys and
// values (component-asset-command.ts). id is bounded by
// MAX_ENGINE_ELEMENT_ID_LENGTH (128, engine-protocol.ts) and mediaType is
// a handful of ASCII bytes in practice ("image/png", "image/jpeg"); even a
// pathological worst case with every id/mediaType byte JSON-escaped to
// \uXXXX (6 bytes each) stays under 2 KiB. 4 KiB leaves comfortable
// headroom without materially shrinking the budget below.
const maxComponentAssetPayloadOverheadBytes = 4 * 1024

// maxComponentAssetBytes is D-5.13.4's host-memory bound, DERIVED from the
// protocol envelope rather than reused verbatim (Finding 6, review of
// 2026-08-29). The command travels as JSON containing BASE64 — a 4/3
// expansion — so applying the raw 8 MiB envelope ceiling directly to the
// DECODED byte count (as this constant did before the fix) let a file
// between roughly 6 and 8 MiB pass Go's own check while the base64-inflated
// envelope had already rejected it at the TRANSPORT, before the command
// diagnostic AC2 requires could ever be produced — the protocol threshold
// and the author-facing diagnostic disagreed, which D-5.13.4 explicitly
// forbids ("the two must not disagree about the threshold"). This is
// instead the largest DECODED size whose base64-encoded command payload,
// plus the skeleton overhead above, is still guaranteed to fit inside
// engineProtocolMaxPayloadBytes — so a file Go is willing to accept can
// always actually arrive. It remains, as before, a memory judgement rather
// than an arithmetic proof like maxImagePixelDimension's (int64 overflow
// in geom.ScaleRound) — only its GROUND changed, not its honesty about
// what kind of number it is.
const maxComponentAssetBytes = (engineProtocolMaxPayloadBytes - maxComponentAssetPayloadOverheadBytes) * 3 / 4

// setComponentAsset is AC1/AC4's asset-authoring command (D-5.13.1): a
// closed, two-value payload (raw bytes plus declared media type) that
// propertyChange's {op,value} grammar cannot express, and AC4's rule that an
// image element is never legally asset-less means clear/null must stay
// inexpressible for it. It is therefore its own top-level command kind, not
// a key threaded through applyPropertyChanges/propertyPath/allowed.
//
// Go alone hashes the decoded bytes, recognises the media type (reusing
// image.go's DecodeImageForRender rather than a second capability check),
// inserts the asset only if its key is absent, repoints the target element,
// and collects the previous asset key it just orphaned — scoped to that one
// key, never a document-wide sweep (D-5.13.3). The whole thing runs inside
// one serialize/reparse/project transaction, matching every other component
// command's no-partial-mutation guarantee.
func setComponentAsset(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	before, err := SerializeTemplate(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	working, err := ParseTemplate(before)
	if err != nil {
		return CanvasProjection{}, err
	}
	projection, err := setComponentAssetInPlace(working, raw)
	if err != nil {
		return CanvasProjection{}, err
	}
	canonical, err := SerializeTemplate(working)
	if err != nil {
		return CanvasProjection{}, err
	}
	installed, err := ParseTemplate(canonical)
	if err != nil {
		return CanvasProjection{}, componentFailure("", "component.asset", "component asset did not pass format validation")
	}
	t.doc, t.derivedFooters = installed.doc, installed.derivedFooters
	return projection, nil
}

func setComponentAssetInPlace(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 5); err != nil {
		return CanvasProjection{}, err
	}
	id, err := commandString(raw, "id")
	if err != nil {
		return CanvasProjection{}, err
	}
	_, _, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "component.id", "component was not found")
	}
	if element.Type != template.ElementImage {
		return CanvasProjection{}, componentFailure(id, "component.id", "only an image component can receive an asset")
	}
	mediaType, err := commandString(raw, "mediaType")
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "component.mediaType", err.Error())
	}
	dataRaw, ok := raw["data"]
	if !ok {
		return CanvasProjection{}, componentFailure(id, "component.data", "asset data is required")
	}
	var dataB64 string
	if json.Unmarshal(dataRaw, &dataB64) != nil || dataB64 == "" {
		return CanvasProjection{}, componentFailure(id, "component.data", "asset data must be a non-empty base64 string")
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(dataB64)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "component.data", "asset data must be valid base64")
	}
	if len(decoded) == 0 {
		return CanvasProjection{}, componentFailure(id, "component.data", "asset data cannot be empty")
	}
	if len(decoded) > maxComponentAssetBytes {
		return CanvasProjection{}, componentFailure(id, "component.data", fmt.Sprintf("asset exceeds the %d-byte supported size", maxComponentAssetBytes))
	}
	digest := sha256.Sum256(decoded)
	key := fmt.Sprintf("%x", digest)
	// AC1: media-type recognition and decode validation happen at the
	// COMMAND, never relying on decodeAssets (parse.go) as the catcher — a
	// file this library version cannot decode is refused here, before
	// anything is written to t.doc.Assets.
	if _, err := template.DecodeImageForRender(mediaType, decoded, key, id); err != nil {
		return CanvasProjection{}, componentFailure(id, "component.mediaType", err.Error())
	}
	previousKey := ""
	if element.Asset.Set && !element.Asset.Null {
		previousKey = element.Asset.Value
	}
	if t.doc.Assets == nil {
		t.doc.Assets = map[string]template.Asset{}
	}
	if _, exists := t.doc.Assets[key]; !exists {
		// Re-wrapped canonically (76 columns, AD-9) by writeAssets at
		// serialize time regardless of how it is stored here; a single
		// element is sufficient in memory.
		t.doc.Assets[key] = template.Asset{MediaType: mediaType, Data: []string{base64.StdEncoding.EncodeToString(decoded)}}
	}
	element.Asset = template.Presence[string]{Set: true, Value: key}
	if previousKey != "" && previousKey != key && !assetKeyReferenced(t, previousKey) {
		delete(t.doc.Assets, previousKey)
	}
	return Canvas(t)
}

// assetKeyReferenced reports whether any image element, across every band,
// still names key. D-5.13.3: orphan collection is scoped to exactly the one
// key this command just repointed away from, never a document-wide sweep —
// a document may legally carry an asset no element references (RP-11's
// positive control, render_image_test.go), and this command must never
// silently remove one it did not just orphan.
//
// This is the SAFETY half of a delete: under-reporting a reference here
// deletes a live asset with no compile error to announce it. It walks the
// same three top-level band element lists (pageHeader/content/pageFooter)
// findComponent (component_commands.go) and addCanvasImagePaint
// (page_setup.go) enumerate — correct for today's model, where images in
// table cells are explicitly out of scope (AC4's exclusions, Finding 17,
// review of 2026-08-29). If a later story places an image anywhere else
// (a table cell, most likely), this walk, findComponent's and
// addCanvasImagePaint's ALL need the new location added together — there
// is no single shared element-enumeration helper today, so update all
// three by hand rather than assuming one covers the others.
func assetKeyReferenced(t *Template, key string) bool {
	for _, elements := range [][]template.Element{t.doc.Bands.PageHeader.Elements, t.doc.Bands.Content.Elements, t.doc.Bands.PageFooter.Elements} {
		for _, el := range elements {
			if el.Type == template.ElementImage && el.Asset.Set && !el.Asset.Null && el.Asset.Value == key {
				return true
			}
		}
	}
	return false
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
	for _, key := range []string{"x", "y", "width", "height", "value", "expression", "visibleIf", "fontFamily", "fontSize", "bold", "italic", "align", "valign", "background", "borderWidth", "borderColor", "borderEdges", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft"} {
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
	propertyOrder := []string{"x", "y", "width", "height", "value", "expression", "visibleIf", "fontFamily", "fontSize", "bold", "italic", "align", "valign", "background", "borderWidth", "borderColor", "borderEdges", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft"}
	if element.Type != template.ElementTable {
		allowed["width"], allowed["height"] = true, true
	}
	if element.Type == template.ElementText {
		allowed["value"] = true
		allowed["expression"] = true
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
		case "value", "expression", "visibleIf", "fontFamily", "align", "valign", "background", "borderColor":
			var text string
			if !clear && !setNull {
				text, err = propertyString(value)
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				if key != "value" && key != "expression" && stringsContainsPlaceholder(text) {
					return fmt.Errorf("%s must not contain a placeholder", key)
				}
			}
			switch key {
			case "value":
				if clear || setNull {
					return fmt.Errorf("value cannot be cleared")
				}
				if stringsContainsPlaceholder(text) {
					// Direct bindings are authored only by bindComponentScalar. This
					// command deliberately remains useful for literal text, but cannot
					// become a second, typed expression route.
					return fmt.Errorf("value must not contain a placeholder; choose a path in the Data panel")
				}
				element.Value = template.Presence[string]{Set: true, Value: text}
			case "expression":
				if clear || setNull || !stringsContainsPlaceholder(text) {
					return fmt.Errorf("expression must contain a template placeholder")
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
			// The palette's text control is usable immediately, for the same
			// reason the image control below embeds a default asset: Render
			// resolves a face through style.fontFamily and refuses text without
			// one, so a placed element that named no chain could never render.
			if chain := defaultFontFamily(t); chain != "" {
				styleFor(&element).FontFamily = template.Presence[string]{Set: true, Value: chain}
			}
		}
		if elementType == template.ElementImage {
			// The palette's image control is usable immediately: it inserts the
			// shipped Folio mark as an embedded, canonical document asset rather
			// than an empty image reference that could never render. Authors can
			// position/size it through the normal selected-component controls.
			if t.doc.Assets == nil {
				t.doc.Assets = map[string]template.Asset{}
			}
			if _, exists := t.doc.Assets[defaultAuthoringLogoKey]; !exists {
				t.doc.Assets[defaultAuthoringLogoKey] = template.Asset{MediaType: "image/png", Data: []string{
					"iVBORw0KGgoAAAANSUhEUgAAADAAAAAgCAIAAADbtmxLAAAAR0lEQVR42u3XQQ0AIAwEsLnACW7w",
					"LwMkwAPCIF1OQB+3JYtSW6oE0GegfmCAgIBWQOnWHug50K4KAwEBuUNAQEC3QT5XoGkGGMiUzlsM",
					"R5wAAAAASUVORK5CYII=",
				}}
			}
			element.Asset = template.Presence[string]{Set: true, Value: defaultAuthoringLogoKey}
		}
	}
	if err := containComponent(projected, x, y, width, height); err != nil {
		return CanvasProjection{}, componentFailure("", "component.geometry", err.Error())
	}
	band.Elements = append(band.Elements, element)
	t.doc.NextID++
	return Canvas(t)
}

const defaultAuthoringLogoKey = "f4a37bba5652865abc8e24be5e1aad4d5ad42ce5727715f6d19b93861d23f6a4"

// defaultFontFamily names the chain a newly created text element adopts: the
// first declared non-empty chain in sorted key order. Sorted rather than
// ranged (ScanMapRange), so a document's declared fonts pick the same chain on
// every run. An empty result means the document declares no usable chain and
// there is nothing to adopt; fontFamily stays absent exactly as before.
func defaultFontFamily(t *Template) string {
	for _, name := range slices.Sorted(maps.Keys(t.doc.Fonts)) {
		if len(t.doc.Fonts[name]) > 0 {
			return name
		}
	}
	return ""
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
