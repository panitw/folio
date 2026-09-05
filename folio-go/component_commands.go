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
	"unicode/utf8"

	"github.com/panitw/folio/folio-go/internal/expr"
	"github.com/panitw/folio/folio-go/internal/fontset"
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

// The two exported command doors both name a DOCUMENT rather than an element
// when they refuse ambiguous bytes, because a duplicate key means the id the
// command names is exactly the thing that cannot be trusted.
const (
	componentCommandPath = "command"
	pageSetupCommandPath = "page.setup"
)

// EACH DOOR'S REFUSAL CARRIES ITS OWN DIAGNOSTIC CODE, and the two are reached
// by different machinery at the host.
//
// wasm/cmd/engine/main.go's engineFailure matches *ComponentCommandError FIRST,
// before the page-setup fallback below it, and answers COMPONENT_INVALID. That
// is right for the component door and WRONG for page setup: the designer's only
// code-branching page-setup consumer keys on PAGE_SETUP_INVALID, so a page-setup
// refusal wearing COMPONENT_INVALID is a refusal that surface never sees.
//
// So the page-setup door returns a plain error whose message opens with the
// `folio: page.` prefix the host's own fallback tests for. That path is not the
// unlocated ENGINE_REJECTED the "use componentFailure" rule exists to avoid — it
// sets DiagnosticCode PAGE_SETUP_INVALID, carries the message, and defaults
// DataPath to page.setup, with ElementID left empty exactly as required.
const pageSetupFailurePrefix = "folio: " + pageSetupCommandPath

// maxCommandKeyScanDepth bounds the duplicate-key walk's recursion. It is not a
// rule about commands — the deepest one this vocabulary has is three objects —
// it is a bound on hostile bytes, because Decoder.Token() streams to ANY depth
// (measured: 20 000 nested arrays tokenise cleanly) while this walk is
// recursive.
//
// The number is encoding/json's own nesting limit, and that is the whole point:
// measured, json.Unmarshal refuses input nested deeper than 10 000 with
// "exceeded max depth", so a value this scan declines to walk is a value the
// caller's ordinary decode rejects on the very next line. The scan never has to
// be the thing that reports depth, and nothing decodable escapes it.
const maxCommandKeyScanDepth = 10000

// refuseDuplicateCommandKeys is the SOLE duplicate-key guard for both exported
// command doors, and it exists because every arity and version check on either
// door is duplicate-BLIND by construction. Both decode into
// map[string]json.RawMessage, which resolves a repeated key silently by
// last-wins BEFORE anything counts the map — so componentFields(raw, want) and
// len(raw) != 7 both count the deduplicated map, equalNumber(raw["version"],
// "1") reads the LAST version, and dispatch routes on the LAST kind.
//
// Executed at the baseline this replaces: a command that NAMED e1 in the page
// header MUTATED e5 in the page footer and returned a nil error. A version:0
// command was admitted by appending a second version:1. A deleteComponent was
// escalated into deleteFontChain because the two happen to have the same
// arity, and only the second handler's own field names stopped it — arity
// coincidence, not a check.
//
// encoding/json cannot report duplicates through any decode: by the time a map
// or a RawMessage is in hand, the duplicate is already gone. Token streaming is
// the only stdlib mechanism that sees them, and it is the idiom this module
// already uses (decoder.More at :519, internal/bind/value.go:161-183).
func refuseDuplicateCommandKeys(command []byte, door string) error {
	dec := json.NewDecoder(bytes.NewReader(command))
	dec.UseNumber()
	first, err := dec.Token()
	if err != nil {
		// Not this guard's refusal to make. Bytes that do not tokenise are
		// malformed, and the caller's own decode says so in its own words.
		return nil
	}
	at, key, found := scanForDuplicateKey(dec, first, "$", 0)
	if !found {
		return nil
	}
	message := fmt.Sprintf("the command declares the key %q twice at %s, so it names one thing and could change another", key, at)
	if door == pageSetupCommandPath {
		return fmt.Errorf("%s: %s", pageSetupFailurePrefix, message)
	}
	return componentFailure("", door, message)
}

// scanForDuplicateKey consumes the remainder of the value whose opening token
// is tok, and reports the path of the first object that declares a key twice.
// Arrays are walked too: an object nested inside one is not a special case
// anywhere else and must not be one here.
func scanForDuplicateKey(dec *json.Decoder, tok json.Token, path string, depth int) (string, string, bool) {
	delim, ok := tok.(json.Delim)
	if !ok {
		// A scalar is one token and it has already been read.
		return "", "", false
	}
	if depth > maxCommandKeyScanDepth {
		// DRAIN, NEVER JUST RETURN. Returning here without consuming the value
		// would leave the decoder positioned INSIDE it, so the parent object's
		// loop would read this value's nested keys as its own — reporting a
		// duplicate at a path that does not exist, or at no duplicate at all.
		// That would be a correctness bug in the very scanner this guard is.
		drainValue(dec, delim)
		return "", "", false
	}
	switch delim {
	case '{':
		// Lookup and insert only. lint/internal/rules/maprange.go bans ranging
		// a map in non-test Go, and nothing here needs to.
		seen := map[string]bool{}
		for dec.More() {
			keyToken, err := dec.Token()
			if err != nil {
				return "", "", false
			}
			key, ok := keyToken.(string)
			if !ok {
				return "", "", false
			}
			if seen[key] {
				return path, key, true
			}
			seen[key] = true
			valueToken, err := dec.Token()
			if err != nil {
				return "", "", false
			}
			if at, duplicate, found := scanForDuplicateKey(dec, valueToken, path+"."+key, depth+1); found {
				return at, duplicate, true
			}
		}
	case '[':
		for index := 0; dec.More(); index++ {
			valueToken, err := dec.Token()
			if err != nil {
				return "", "", false
			}
			if at, duplicate, found := scanForDuplicateKey(dec, valueToken, fmt.Sprintf("%s[%d]", path, index), depth+1); found {
				return at, duplicate, true
			}
		}
	}
	// The closing delimiter. Consuming it here is what leaves the decoder
	// positioned on the parent's next token.
	if _, err := dec.Token(); err != nil {
		return "", "", false
	}
	return "", "", false
}

// drainValue consumes the remainder of a composite value whose opening
// delimiter has already been read, leaving the decoder on the parent's next
// token. It is iterative on purpose: it exists to handle input too deeply
// nested to recurse over, so recursing over it would defeat itself.
func drainValue(dec *json.Decoder, opening json.Delim) {
	depth := 1
	if opening != '{' && opening != '[' {
		return
	}
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		if delim, ok := tok.(json.Delim); ok {
			if delim == '{' || delim == '[' {
				depth++
			} else {
				depth--
			}
		}
	}
}

// ApplyComponentCommand applies Story 5.7's small, versioned authoring
// vocabulary. The command is intentionally decoded in Go: the browser sends
// opaque bytes and never receives the template or its canonical JSON shape.
func ApplyComponentCommand(t *Template, command []byte) (CanvasProjection, error) {
	if t == nil {
		return CanvasProjection{}, errNilTemplate
	}
	if err := refuseDuplicateCommandKeys(command, componentCommandPath); err != nil {
		return CanvasProjection{}, err
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
	case "setComponentBounds":
		return setComponentBounds(t, raw)
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
	case "addFontChain":
		return applyFontChainCommand(t, raw, addFontChain)
	case "renameFontChain":
		return applyFontChainCommand(t, raw, renameFontChain)
	case "deleteFontChain":
		return applyFontChainCommand(t, raw, deleteFontChain)
	case "addFontChainEntry":
		return applyFontChainCommand(t, raw, addFontChainEntry)
	case "moveFontChainEntry":
		return applyFontChainCommand(t, raw, moveFontChainEntry)
	case "removeFontChainEntry":
		return applyFontChainCommand(t, raw, removeFontChainEntry)
	case "embedFontFamily":
		return applyFontChainCommand(t, raw, embedFontFamily)
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

// assetKeyReferenced reports whether anything in the document still names key —
// an image ELEMENT, across every band, or a FONT CHAIN entry. D-5.13.3: orphan
// collection is scoped to exactly the one
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
//
// DW-80 IS FIXED HERE (Story 8.6), AND IT WAS A REAL HOLE, NOT A LATENT ONE.
// Until this story the only `true` arm required `el.Type == ElementImage`, and
// the walk never read `t.doc.Fonts` at all — so this function answered FALSE
// for every font asset in every document, however many chains named it. Nothing
// called it for a font, which is why it was survivable; this story's orphan
// drop is the first caller that does, and shipping that drop over the old walk
// would have deleted a face a live chain was still drawing with. The DELETE is
// the dangerous direction: under-reporting a reference here removes a live
// asset, and no compile error announces it.
//
// THE THREE-WALK WARNING ABOVE DOES NOT EXTEND TO THIS ARM, and that is a
// measurement rather than an omission. `findComponent` and `addCanvasImagePaint`
// enumerate ELEMENTS, because an image is placed by one; a font asset is named
// by an entry of a document-level map that holds no elements at all, so there
// is no third location for those two to gain. What DOES pair with this arm is
// fontChainReferences, which walks the same map for the same safety reason from
// the other direction (which elements name a chain).
func assetKeyReferenced(t *Template, key string) bool {
	for _, elements := range [][]template.Element{t.doc.Bands.PageHeader.Elements, t.doc.Bands.Content.Elements, t.doc.Bands.PageFooter.Elements} {
		for _, el := range elements {
			if el.Type == template.ElementImage && el.Asset.Set && !el.Asset.Null && el.Asset.Value == key {
				return true
			}
		}
	}
	// SORTED KEYS, not a bare map range (AD-1, NFR1.d). The ANSWER here does
	// not depend on the order — it is an existence question — but the rule is
	// module-wide and absolute for a reason: an iteration order that happens
	// not to matter today is the one that quietly starts mattering when
	// somebody returns the first match instead of a bool.
	for _, name := range slices.Sorted(maps.Keys(t.doc.Fonts)) {
		for _, entry := range t.doc.Fonts[name] {
			// Embedded() is THE discriminant (template.FontChainEntry); a
			// bare `entry.AssetKey == key` would be the same test written a
			// second time, and a face entry can never carry an asset key
			// anyway — the partition is the type's own invariant.
			if entry.Embedded() && entry.AssetKey == key {
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
	for _, key := range []string{"x", "y", "width", "height", "value", "expression", "visibleIf", "fontFamily", "fontSize", "lineSpacing", "bold", "italic", "align", "valign", "color", "background", "borderWidth", "borderColor", "borderEdges", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft"} {
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
	propertyOrder := []string{"x", "y", "width", "height", "value", "expression", "visibleIf", "fontFamily", "fontSize", "lineSpacing", "bold", "italic", "align", "valign", "color", "background", "borderWidth", "borderColor", "borderEdges", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft"}
	if element.Type != template.ElementTable {
		allowed["width"], allowed["height"] = true, true
	}
	if element.Type == template.ElementText {
		allowed["value"] = true
		allowed["expression"] = true
	}
	if element.Type == template.ElementText || element.Type == template.ElementImage || element.Type == template.ElementTable || element.Type == template.ElementLine || element.Type == template.ElementRect {
		for _, key := range []string{"background", "borderWidth", "borderColor", "borderEdges"} {
			allowed[key] = true
		}
	}
	if element.Type == template.ElementTable {
		// D-12.4.1: `padding` is consumed by a table's cell chrome and by
		// nothing else on any render path, so a table is the only kind that
		// may be commanded to change it. The asymmetry is deliberate: a
		// loaded document KEEPS and round-trips padding on any kind, and
		// RENDERS it where a table consumes it — the engine honours what it
		// is given — while the designer refuses to author, on the four kinds
		// that would never paint it, a value nothing would ever read.
		for _, key := range []string{"paddingTop", "paddingRight", "paddingBottom", "paddingLeft"} {
			allowed[key] = true
		}
	}
	if element.Type == template.ElementText || element.Type == template.ElementTable {
		// Story 10.1: `color` is the ink text prints in, so it is offered
		// exactly where text is — never on a rect, line or image, which
		// carry no glyphs for it to colour.
		for _, key := range []string{"fontFamily", "fontSize", "lineSpacing", "bold", "italic", "align", "valign", "color"} {
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
		case "value", "expression", "visibleIf", "fontFamily", "align", "valign", "color", "background", "borderColor":
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
					// Story 7.3. This arm set style.align to WHATEVER
					// STRING ARRIVED — pre-existing, and harmless only
					// while one closed set served both vocabularies.
					// With more than one live it is the one remaining
					// place they could be conflated, so it validates
					// through the closed sets' own exported predicates,
					// and names the legal values from the matching
					// ordered slice rather than from a literal
					// restating it. The COLUMN arm (updateTableColumn,
					// above) keeps its own triple and still refuses
					// "justify".
					//
					// Story 7.8: SELECTED BY ELEMENT TYPE, the same way
					// decodeStyle selects. IsStyleAlign's own doc
					// comment requires this path to validate against
					// the same single source the loader does.
					//
					// MEASURED, not assumed: updateComponentProperties
					// serializes and RE-PARSES before installing, so
					// once the loader refuses a table's `justify` the
					// round trip already stops the document reaching
					// 2.0 through this door — with the generic
					// "component properties did not pass format
					// validation". What this arm adds is the REFUSAL
					// THE AUTHOR CAN ACT ON: the field named, and the
					// legal values for a table rather than for a
					// paragraph. Without it the inspector would report
					// a whole-command failure for one bad value and
					// name neither. It is also the layer that does not
					// depend on the round trip continuing to exist.
					admits, tokens := template.IsStyleAlign, template.StyleAlignTokens
					if element.Type == template.ElementTable {
						admits, tokens = template.IsTableStyleAlign, template.TableStyleAlignTokens
					}
					if !admits(text) {
						return fmt.Errorf("align must be one of %s", strings.Join(tokens, ", "))
					}
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
			case "color":
				st := styleFor(element)
				if clear {
					st.Color = template.Presence[string]{}
				} else if setNull {
					st.Color = template.Presence[string]{Set: true, Null: true}
				} else if !validPropertyColor(text) {
					return fmt.Errorf("color must be a #RRGGBB colour")
				} else {
					st.Color = template.Presence[string]{Set: true, Value: text}
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
		case "lineSpacing":
			// NOT propertyLength. That decoder reads POINTS and bounds
			// them by MaxCanvasMillipoints; lineSpacing is a
			// dimensionless ratio with its own domain, and borrowing a
			// length decoder would give the inspector a different notion
			// of a legal value from the one the file path enforces.
			//
			// D-7.2.3's "a value refused in a file is refused in the
			// inspector for the SAME reason" is satisfied by calling the
			// SAME function the loader calls — template.DecodeLineSpacing
			// — not by mirroring its bounds here.
			if setNull {
				return fmt.Errorf("%s does not support null", key)
			}
			st := styleFor(element)
			if clear {
				st.LineSpacing = template.Presence[int64]{}
				continue
			}
			thousandths, err := template.DecodeLineSpacingRaw(value)
			if err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
			st.LineSpacing = template.Presence[int64]{Set: true, Value: thousandths}
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
	_, ok := t.doc.Fonts.Chain(value)
	return ok
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
	if !style.Align.Set && !style.Background.Set && !style.Bold.Set && !style.Color.Set && !style.Italic.Set && !style.Border.Set && !style.FontFamily.Set && !style.FontSize.Set && !style.LineSpacing.Set && !style.Padding.Set && !style.Valign.Set && len(style.Extra) == 0 {
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
		return snapField(name, v)
	}
	return v, nil
}

func snapField(name string, value geom.Length) (geom.Length, error) {
	snapped, valid := SnapToGrid(value)
	if !valid {
		return 0, fmt.Errorf("folio: component.%s overflows grid snapping", name)
	}
	return snapped, nil
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
		case bandPageHeader:
			return &t.doc.Bands.PageHeader, projected, nil
		case bandContent:
			return &t.doc.Bands.Content, projected, nil
		case bandPageFooter:
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
// The size a dropped component starts at, before any property edit. An image
// starts larger than the rest: until a file is chosen its box carries the
// designer's empty-state placeholder — an icon above a label — which a
// 72x24 box cuts in half. Both sizes sit on the 6pt grid, so a snapped drop
// keeps them exactly.
const dropWidth, dropHeight geom.Length = 72000, 24000
const imageDropWidth, imageDropHeight geom.Length = 96000, 48000

// Story 9.2: a line's declared HEIGHT is its thickness — element_box.go
// paints a line as a filled bar of its declared box — so a line drops as a
// 1pt rule rather than as a 72x24 slab. Off the 6pt grid on purpose: a
// rule's thickness is not a position, and snapping applies to x/y alone.
const lineDropHeight geom.Length = 1000

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
	width, height := dropWidth, dropHeight
	if elementType == template.ElementImage {
		width, height = imageDropWidth, imageDropHeight
	}
	if elementType == template.ElementLine {
		height = lineDropHeight
	}
	x, y := pageX-geom.Length(projected.X), pageY-geom.Length(projected.Y)
	unsnappedX, unsnappedY := x, y
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
		if containComponent(projected, unsnappedX, unsnappedY, width, height) == nil {
			x = containEdge(x, geom.Length(projected.Width)-width)
			y = containEdgeY(projected, y, geom.Length(projected.Height)-height)
		}
	}
	return createComponentInBand(t, elementType, projected.Name, x, y, width, height)
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
		// Story 9.2: a line and a rect ARE their box — they carry no text
		// and no asset — so a placed one with no style would render, and
		// paint on the canvas, as nothing at all. Each starts with the one
		// declaration that makes it the shape its palette entry names: a
		// line is a filled rule, a rect is an outlined box. Both are
		// ordinary style values the author edits or clears like any other.
		if elementType == template.ElementLine {
			styleFor(&element).Background = template.Presence[string]{Set: true, Value: "#000000"}
		}
		if elementType == template.ElementRect {
			styleFor(&element).Border = template.Presence[template.Border]{Set: true, Value: template.Border{
				Color: template.Presence[string]{Set: true, Value: "#000000"},
				Width: template.Presence[geom.Length]{Set: true, Value: 1000},
			}}
		}
		if elementType == template.ElementImage {
			// A placed image starts empty: the author positions and sizes the
			// box first and chooses the file through the inspector, which is
			// the state the design draws as a dashed placeholder. The asset
			// field is present and null rather than absent, so the document
			// still declares the box as an image and Render draws nothing for
			// it until a file is set.
			element.Asset = template.Presence[string]{Set: true, Null: true}
		}
	}
	if err := containComponent(projected, x, y, width, height); err != nil {
		return CanvasProjection{}, componentFailure("", "component.geometry", err.Error())
	}
	band.Elements = append(band.Elements, element)
	t.doc.NextID++
	return Canvas(t)
}

// defaultFontFamily names the chain a newly created text element adopts: the
// first declared non-empty chain in sorted key order. Sorted rather than
// ranged (ScanMapRange), so a document's declared fonts pick the same chain on
// every run. An empty result means the document declares no usable chain and
// there is nothing to adopt; fontFamily stays absent exactly as before.
func defaultFontFamily(t *Template) string {
	for _, name := range slices.Sorted(maps.Keys(t.doc.Fonts)) {
		if _, ok := t.doc.Fonts.Chain(name); ok {
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
		case bandPageHeader:
			return &t.doc.Bands.PageHeader, band, nil
		case bandContent:
			return &t.doc.Bands.Content, band, nil
		case bandPageFooter:
			return &t.doc.Bands.PageFooter, band, nil
		}
	}
	return nil, CanvasBand{}, componentFailure("", "component.drop", "drop point is outside a page band")
}

func findComponent(t *Template, id string) (*template.Band, CanvasBand, int, *template.Element, error) {
	for _, name := range []string{bandPageHeader, bandContent, bandPageFooter} {
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
	unsnappedX, err := componentLength(raw, "x", false)
	if err != nil {
		return CanvasProjection{}, err
	}
	unsnappedY, err := componentLength(raw, "y", false)
	if err != nil {
		return CanvasProjection{}, err
	}
	x, y := unsnappedX, unsnappedY
	if snap {
		if x, err = snapField("x", unsnappedX); err != nil {
			return CanvasProjection{}, err
		}
		if y, err = snapField("y", unsnappedY); err != nil {
			return CanvasProjection{}, err
		}
	}
	_, projected, _, element, err := findComponent(t, id)
	if err != nil {
		return CanvasProjection{}, componentFailure(id, "component.id", "component was not found")
	}
	width, height := projectedSize(*element)
	if snap && containComponent(projected, unsnappedX, unsnappedY, width, height) == nil {
		x = containEdge(x, geom.Length(projected.Width)-width)
		y = containEdgeY(projected, y, geom.Length(projected.Height)-height)
	}
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

// setComponentBounds is one rectangle, not a move followed by a resize. A
// resize anchored at any edge or corner other than the bottom-right moves the
// origin and the size together; sending moveComponent and resizeComponent in
// sequence would put two entries in history for one drag and would test
// containment against an intermediate rectangle the caller never asked for.
func setComponentBounds(t *Template, raw map[string]json.RawMessage) (CanvasProjection, error) {
	if err := componentFields(raw, 8); err != nil {
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
	unsnappedX, err := componentLength(raw, "x", false)
	if err != nil {
		return CanvasProjection{}, err
	}
	unsnappedY, err := componentLength(raw, "y", false)
	if err != nil {
		return CanvasProjection{}, err
	}
	unsnappedWidth, err := componentLength(raw, "width", false)
	if err != nil {
		return CanvasProjection{}, err
	}
	unsnappedHeight, err := componentLength(raw, "height", false)
	if err != nil {
		return CanvasProjection{}, err
	}
	x, y, width, height := unsnappedX, unsnappedY, unsnappedWidth, unsnappedHeight
	if snap {
		for _, field := range [4]struct {
			name  string
			value *geom.Length
		}{{"x", &x}, {"y", &y}, {"width", &width}, {"height", &height}} {
			if *field.value, err = snapField(field.name, *field.value); err != nil {
				return CanvasProjection{}, err
			}
		}
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
	if snap && containComponent(projected, unsnappedX, unsnappedY, unsnappedWidth, unsnappedHeight) == nil {
		width = containEdge(width, geom.Length(projected.Width)-x)
		height = containEdgeY(projected, height, geom.Length(projected.Height)-y)
		x = containEdge(x, geom.Length(projected.Width)-width)
		y = containEdgeY(projected, y, geom.Length(projected.Height)-height)
	}
	if err := containComponent(projected, x, y, width, height); err != nil {
		return CanvasProjection{}, componentFailure(id, "component.geometry", err.Error())
	}
	element.X, element.Y = x, y
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
	// Story 7.9 / D-7.7.10: a duplicate joins NO keep-together group.
	//
	// `clone := *element` above is a whole-struct copy, so without this line
	// the copy silently inherits the original's tag — and Epic 7 ships no way
	// anywhere in the designer to see a tag, set one or clear one (file-only
	// authoring is the stated scope boundary; FR51 asks only that a group can
	// be DECLARED). Duplicating a signature block would therefore enlarge a
	// keep-together set the author cannot reach, moving a page break for a
	// reason the product never shows them. The project refuses document state
	// the author cannot undo, and this is that rule at the copy path.
	//
	// It is cleared to the ZERO Presence, never to an explicit null: `Set:
	// true, Null: true` serializes back as `"keepTogether": null`, which is
	// still the key appearing in the file and still raises the document's
	// required format version. "No tag" is the field's absence. This is the
	// same spelling every other optional field's clear site uses
	// (`element.VisibleIf = template.Presence[string]{}`).
	clone.KeepTogether = template.Presence[string]{}
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

// Grid snapping is a convenience applied to the caller's number, not a second
// constraint placed on it: a rectangle that fitted its band before the grid
// rounded it must still fit afterwards. Callers below pull an edge back to the
// last grid line that fits, and only ever when the unsnapped rectangle already
// fitted — so snapping can never turn a legal drag into a refusal, and
// geometry the caller placed outside the band is still refused unchanged.
func containEdge(value, limit geom.Length) geom.Length {
	if value <= limit {
		return value
	}
	return floorToGrid(limit)
}

func floorToGrid(value geom.Length) geom.Length {
	if value <= 0 {
		return 0
	}
	return geom.Length(int64(value) / GridIncrement * GridIncrement)
}

// The three band identities, as page_setup.go mints them and as every command
// path compares them. Story 7.5 made band identity LOAD-BEARING for the first
// time — the content band's vertical cap lifts and the two repeating bands'
// does not — and a fifth inline spelling of a bare string is exactly how a
// distinction like that diverges without anything going red.
const (
	bandPageHeader = "pageHeader"
	bandContent    = "content"
	bandPageFooter = "pageFooter"
)

// bandsCappingVertically names the bands whose DECLARED HEIGHT bounds a
// component's vertical extent.
//
// The content band is absent by MEANING, not by omission. A page header and a
// page footer REPEAT on every page, so each is exactly one page tall and a
// component that left it would have nowhere to be. The content band is a
// COLUMN that pagination slices into page-height windows (internal/layout's
// Paginate), so it has no single height to be inside of: a component below
// the foot of page one is on page two, not outside the document.
//
// THE MIRROR. folio-designer/src/engine-protocol.ts declares this same list
// under this same name, and engine-bounds-mirror.test.ts reads BOTH files and
// asserts they agree. D-7.4.5, as widened by Story 7.5: any invariant
// duplicated across the Go/TypeScript boundary moves in ONE commit, with a
// test that reads both sides. Lifting this here alone would ship a story
// invisible in the running app — the browser's copy of this gate drops the
// whole snapshot, terminates the worker and blanks the canvas, with no
// element id and no attributable error.
var bandsCappingVertically = []string{bandPageHeader, bandPageFooter}

// containComponent is the ONE band-extent validation in the designer command
// path. It enforces two DIFFERENT KINDS of constraint, which used to share a
// single eight-disjunct expression and are separated here by what they MEAN:
//
//   - REPRESENTATIONAL, and therefore universal: a negative coordinate or
//     size is not geometry at all, in any band. This is also the only place
//     negativity is caught — lengthField admits values down to
//     -MaxCanvasMillipoints — so these terms are load-bearing.
//   - HORIZONTAL, and therefore universal: the column is unbounded
//     vertically, never horizontally. A band is as wide as the printable
//     page and nothing may hang off its side.
//   - BAND CAPACITY, and therefore only where a band HAS a capacity: see
//     bandsCappingVertically.
//
// Every surviving refusal keeps the same message, to the byte, because from
// the author's side they are one complaint: this component is not inside that
// band.
//
// The split keys on band.Name INSIDE this function and never at a call site.
// findComponent, bandByName and hitTestBand each range over all three names,
// so every one of the eleven callers can receive any of the three bands.
func containComponent(band CanvasBand, x, y, width, height geom.Length) error {
	outside := x < 0 || y < 0 || width < 0 || height < 0 || x > geom.Length(band.Width) || width > geom.Length(band.Width)-x
	if !outside && slices.Contains(bandsCappingVertically, band.Name) {
		outside = y > geom.Length(band.Height) || height > geom.Length(band.Height)-y
	}
	if outside {
		return fmt.Errorf("folio: component geometry must stay within %s", band.Name)
	}
	return nil
}

// containEdgeY is containEdge on the VERTICAL axis, and after Story 7.5 it is
// a pull-back only in the bands that cap vertically.
//
// It exists because the pre-clamps at its call sites are GATED on the
// unsnapped rectangle already fitting, so lifting the content band's cap does
// not neutralise them — it WIDENS the gate. A drag that is refused outright
// today starts passing the probe, and a containEdge left in place would then
// quietly pull its Y back to the foot of page one: "this component may live
// on page four" would become "this component snapped to the bottom of page
// one", with no refusal and no explanation.
func containEdgeY(band CanvasBand, value, limit geom.Length) geom.Length {
	if !slices.Contains(bandsCappingVertically, band.Name) {
		return value
	}
	return containEdge(value, limit)
}

// ---------------------------------------------------------------------------
// STORY 8.1: THE DOCUMENT'S FONT CHAINS, AS COMMANDS.
//
// Six kinds that write the one document-level map nothing could write before
// (template.Document.Fonts). They are modelled on setComponentAsset, the
// module's other command that writes a document-level map and repoints the
// elements naming it (D-5.13.1), and they inherit its guarantee the same way:
// applyFontChainCommand serializes, reparses and projects a CLONE, and the
// caller's template is installed only if all three succeed. A rename that
// rewrites a map key and four element references is therefore ONE mutation,
// which is what lets wasm.Engine.Apply push exactly one undo entry for it.

// maxCanvasFontChainEntries bounds ONE chain's entry list the way
// maxCanvasFontFamilies bounds the chain list itself: the projection now
// carries the entries, so an unbounded chain is an unbounded projected array.
// A document declaring a longer chain is refused a projection with a stated
// reason, never silently cut, and the commands refuse to build one.
const maxCanvasFontChainEntries = 64

// maxComponentFailureMessageBytes and maxComponentDataPathBytes are the widths
// the wasm host cuts a ComponentCommandError's Message and DataPath to
// (wasm/cmd/engine/main.go's bounded(componentErr.Message, 512) and
// bounded(componentErr.DataPath, 256)). They are read here so a long id list
// and a long chain name are trimmed HERE — on a whole-id and a whole-rune
// boundary — instead of arriving at the author cut through the middle of an
// element id or a multi-byte character.
//
// They are HAND-COPIED, which is the one-sided-constant defect in pure form:
// wasm/cmd/engine is //go:build js && wasm, so `go test ./...` never compiles
// it and nothing links these two numbers to the host's literals. So they are
// tied by a source-reading tripwire instead —
// TestComponentFailureBoundsMatchTheHostsOwnLiterals reads main.go the way
// canvas_projection_wire_test.go reads engine-protocol.ts. Change one, change
// the other, in the same commit.
const (
	maxComponentFailureMessageBytes = 512
	maxComponentDataPathBytes       = 256
)

// fontChainPath is the DataPath every font-chain refusal carries. A chain
// command is not addressed to an element, so ElementID stays empty and the
// path names the map entry — the shape page-setup refusals already use, and
// the only one available: ComponentCommandError.ElementID is single-valued and
// the orphaning-delete refusal names a LIST of ids, which is why that list
// lives in Message.
// It is bounded HERE at maxComponentDataPathBytes because the two bounds
// disagree: a chain name is legal up to maxCanvasPropertyString (512), so
// "fonts." + name overruns the host's 256-byte DataPath cut, and the host's
// bounded() slices by BYTES — which on a multi-byte name splits a UTF-8 rune.
// The over-long-name refusal is the one case where locating the name is the
// entire point, so it is the one case that must not arrive mangled.
func fontChainPath(name string) string {
	if name == "" {
		return "fonts"
	}
	return truncateAtRuneBoundary("fonts."+name, maxComponentDataPathBytes)
}

// truncateAtRuneBoundary cuts value to at most limit BYTES without splitting a
// UTF-8 rune. The host's own bounded() does not do this, so anything this
// module hands it that could exceed a wire bound is cut here first.
func truncateAtRuneBoundary(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(value[cut]) {
		cut--
	}
	return value[:cut]
}

// applyFontChainCommand is the transaction wrapper, identical in shape to
// applyTableColumnCommand: a handler may mutate its candidate freely, and the
// caller's document is replaced only after that candidate serializes,
// reparses and projects.
func applyFontChainCommand(t *Template, raw map[string]json.RawMessage, apply func(*Template, map[string]json.RawMessage) error) (CanvasProjection, error) {
	before, err := SerializeTemplate(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	working, err := ParseTemplate(before)
	if err != nil {
		return CanvasProjection{}, err
	}
	if err := apply(working, raw); err != nil {
		return CanvasProjection{}, err
	}
	canonical, err := SerializeTemplate(working)
	if err != nil {
		return CanvasProjection{}, err
	}
	installed, err := ParseTemplate(canonical)
	if err != nil {
		return CanvasProjection{}, componentFailure("", "fonts", "font chains did not pass format validation")
	}
	projection, err := Canvas(installed)
	if err != nil {
		return CanvasProjection{}, err
	}
	t.doc, t.derivedFooters = installed.doc, installed.derivedFooters
	return projection, nil
}

// fontChainName reads the author's chain name and applies the two rules every
// chain command shares: it is a non-empty string (commandString's own refusal,
// reused rather than restated) and it fits the projection's identifier bound,
// maxCanvasPropertyString. The length is refused HERE, located at fonts.<name>,
// so the author sees which name is too long instead of canvasFontChains'
// unlocated bare error firing later in the same transaction.
func fontChainName(raw map[string]json.RawMessage, field string) (string, error) {
	name, err := commandString(raw, field)
	if err != nil {
		return "", componentFailure("", "fonts", err.Error())
	}
	if len(name) > maxCanvasPropertyString {
		return "", componentFailure("", fontChainPath(name), "font chain name exceeds the projection bound")
	}
	return name, nil
}

// declaredFontChain resolves a chain a command is about to edit. It asks
// whether the KEY is declared, deliberately NOT template.Fonts.Chain: a chain
// with no entries is not one style.fontFamily may name, but decodeFonts
// accepts one at load and it must stay deletable and fillable rather than
// become unreachable to every command at once.
func declaredFontChain(t *Template, raw map[string]json.RawMessage) (string, []template.FontChainEntry, error) {
	name, err := fontChainName(raw, "name")
	if err != nil {
		return "", nil, err
	}
	chain, ok := t.doc.Fonts[name]
	if !ok {
		return "", nil, componentFailure("", fontChainPath(name), fmt.Sprintf("no font chain named %q is declared", name))
	}
	return name, chain, nil
}

// fontChainFace reads one face name: non-empty, and bounded by the same
// identifier bound the chain name is, because the projection now carries the
// entries too. A face this build's FontSet does not ship is ACCEPTED — the
// format's standing tolerance (render.go's resolveRuneFace skips a chain
// member absent from the set rather than failing), and a chain naming a face
// an embedding story will supply later is a legal chain today.
func fontChainFace(raw map[string]json.RawMessage, name, field string) (string, error) {
	face, err := commandString(raw, field)
	if err != nil {
		return "", componentFailure("", fontChainPath(name), err.Error())
	}
	if len(face) > maxCanvasPropertyString {
		return "", componentFailure("", fontChainPath(name), "font chain entry exceeds the projection bound")
	}
	return face, nil
}

func fontChainIndex(raw map[string]json.RawMessage, name, field string, limit int) (int, error) {
	index, err := commandInt(raw, field)
	if err != nil {
		return 0, componentFailure("", fontChainPath(name), err.Error())
	}
	if index < 0 || index > limit {
		return 0, componentFailure("", fontChainPath(name), "entry index is out of range")
	}
	return index, nil
}

// addFontChain declares a new chain. The duplicate-name refusal is this
// story's, not Story 8.2's: 8.2's panel only reports what the engine answers.
func addFontChain(t *Template, raw map[string]json.RawMessage) error {
	if err := componentFields(raw, 4); err != nil {
		return err
	}
	name, err := fontChainName(raw, "name")
	if err != nil {
		return err
	}
	if _, exists := t.doc.Fonts[name]; exists {
		return componentFailure("", fontChainPath(name), fmt.Sprintf("a font chain named %q already exists", name))
	}
	entriesRaw, ok := raw["entries"]
	if !ok {
		return componentFailure("", fontChainPath(name), "font chain entries are required")
	}
	var entries []string
	if json.Unmarshal(entriesRaw, &entries) != nil {
		return componentFailure("", fontChainPath(name), "font chain entries must be a string array")
	}
	if len(entries) == 0 {
		return componentFailure("", fontChainPath(name), "a font chain must declare at least one entry")
	}
	if len(entries) > maxCanvasFontChainEntries {
		return componentFailure("", fontChainPath(name), "a font chain declares more entries than the projection bound")
	}
	for _, face := range entries {
		if face == "" {
			return componentFailure("", fontChainPath(name), "a font chain entry must be a non-empty string")
		}
		if len(face) > maxCanvasPropertyString {
			return componentFailure("", fontChainPath(name), "font chain entry exceeds the projection bound")
		}
	}
	if len(t.doc.Fonts)+1 > maxCanvasFontFamilies {
		return componentFailure("", fontChainPath(name), "document declares more font chains than the projection bound")
	}
	if t.doc.Fonts == nil {
		t.doc.Fonts = template.Fonts{}
	}
	// Every entry this command can express is a FACE NAME. Story 8.6's
	// pick-and-embed command is what will express an embedded entry;
	// `entries` arrives here as a []string and there is deliberately no
	// spelling in the command vocabulary that produces anything else.
	chainEntries := make([]template.FontChainEntry, 0, len(entries))
	for _, face := range entries {
		chainEntries = append(chainEntries, template.FaceEntry(face))
	}
	t.doc.Fonts[name] = chainEntries
	return nil
}

// renameFontChain moves the key AND carries every element that names it, in
// this one handler. fontFamily has exactly two attachment points in the model
// (see fontChainReferences), and a rename that moved only the key would leave
// them naming a chain that no longer exists — a document that loads and then
// fails at render. Because both halves happen inside one applyFontChainCommand
// transaction, wasm.Engine.Apply's single pushUndo covers the map and the
// elements together, and one undo restores all of them.
func renameFontChain(t *Template, raw map[string]json.RawMessage) error {
	if err := componentFields(raw, 4); err != nil {
		return err
	}
	name, chain, err := declaredFontChain(t, raw)
	if err != nil {
		return err
	}
	to, err := fontChainName(raw, "to")
	if err != nil {
		return err
	}
	if _, exists := t.doc.Fonts[to]; exists {
		// The destination is never silently destroyed, and renaming a chain
		// onto its own name is the same refusal: the key is taken.
		return componentFailure("", fontChainPath(to), fmt.Sprintf("a font chain named %q already exists", to))
	}
	delete(t.doc.Fonts, name)
	t.doc.Fonts[to] = chain
	for _, elements := range fontChainBands(t) {
		for i := range elements {
			element := &elements[i]
			if fontChainNamedBy(element.Style, name) {
				element.Style.Value.FontFamily.Value = to
			}
			if element.Table.Set && !element.Table.Null && fontChainNamedBy(element.Table.Value.HeaderStyle, name) {
				element.Table.Value.HeaderStyle.Value.FontFamily.Value = to
			}
		}
	}
	return nil
}

// deleteFontChain removes a chain nothing names. AC2's principle — a chain is
// never deleted with the orphaned elements left to fail at render — reaches
// headerStyle.fontFamily as squarely as style.fontFamily, so the refusal is
// measured over both.
func deleteFontChain(t *Template, raw map[string]json.RawMessage) error {
	if err := componentFields(raw, 3); err != nil {
		return err
	}
	name, chain, err := declaredFontChain(t, raw)
	if err != nil {
		return err
	}
	if referees := fontChainReferences(t, name); len(referees) > 0 {
		return componentFailure("", fontChainPath(name), fontChainOrphanMessage(name, referees))
	}
	delete(t.doc.Fonts, name)
	// Deleting a chain un-names every entry in it at once, which is the same
	// event removeFontChainEntry produces one at a time — so it collects the
	// same way, scoped to exactly those entries (AC5).
	dropUnnamedFontAssets(t, chain...)
	return nil
}

func addFontChainEntry(t *Template, raw map[string]json.RawMessage) error {
	if err := componentFields(raw, 5); err != nil {
		return err
	}
	name, chain, err := declaredFontChain(t, raw)
	if err != nil {
		return err
	}
	index, err := fontChainIndex(raw, name, "index", len(chain))
	if err != nil {
		return err
	}
	face, err := fontChainFace(raw, name, "face")
	if err != nil {
		return err
	}
	if len(chain)+1 > maxCanvasFontChainEntries {
		return componentFailure("", fontChainPath(name), "a font chain declares more entries than the projection bound")
	}
	t.doc.Fonts[name] = slices.Insert(slices.Clone(chain), index, template.FaceEntry(face))
	return nil
}

func moveFontChainEntry(t *Template, raw map[string]json.RawMessage) error {
	if err := componentFields(raw, 5); err != nil {
		return err
	}
	name, chain, err := declaredFontChain(t, raw)
	if err != nil {
		return err
	}
	from, err := fontChainIndex(raw, name, "from", len(chain)-1)
	if err != nil {
		return err
	}
	to, err := fontChainIndex(raw, name, "to", len(chain)-1)
	if err != nil {
		return err
	}
	moved := slices.Clone(chain)
	face := moved[from]
	moved = slices.Insert(slices.Delete(moved, from, from+1), to, face)
	t.doc.Fonts[name] = moved
	return nil
}

// removeFontChainEntry refuses to empty a chain. A chain with no entries is
// not one style.fontFamily may name (template.Fonts.Chain), so emptying one
// through this command would orphan every element naming it just as surely as
// deleting it would — an ADDITIONAL guard at the command path, not a
// relocation of the render-time rule, which stays where it is.
func removeFontChainEntry(t *Template, raw map[string]json.RawMessage) error {
	if err := componentFields(raw, 4); err != nil {
		return err
	}
	name, chain, err := declaredFontChain(t, raw)
	if err != nil {
		return err
	}
	index, err := fontChainIndex(raw, name, "index", len(chain)-1)
	if err != nil {
		return err
	}
	if len(chain) == 1 {
		return componentFailure("", fontChainPath(name), fmt.Sprintf("removing that entry would leave font chain %q with no entries", name))
	}
	removed := chain[index]
	t.doc.Fonts[name] = slices.Delete(slices.Clone(chain), index, index+1)
	dropUnnamedFontAssets(t, removed)
	return nil
}

// dropUnnamedFontAssets is AC5: A FACE NOTHING NAMES ANY LONGER IS DROPPED BY
// THE COMMAND THAT UN-NAMED IT.
//
// WHY IT IS HERE AND NOT IN THE SERIALIZER. writeAssets (serialize.go) says so
// itself, and it is not a preference: AD-9 / D-1.4.3's P1 is
// `Parse(Serialize(d)) == d`, which FORCES the serializer to preserve an
// orphan unconditionally — a save that collected one would not be a fixed
// point, and a document may legally carry an asset nothing references
// (D-1.4.13, RP-11's positive control). Collecting orphans is a DESIGNER
// FEATURE: the author's own action removes the face, so the author's own
// command removes the bytes, and nothing happens behind their back on save.
// serialize.go is not touched by this story.
//
// SCOPED TO THE KEYS JUST UN-NAMED, NEVER A DOCUMENT-WIDE SWEEP. This is
// setComponentAsset's rule (D-5.13.3) applied to the other asset kind: the
// candidates are exactly the entries this command removed, and each is dropped
// only if assetKeyReferenced now answers false for it. An asset a SECOND chain
// still names is retained — that is the arm DW-80 would have got wrong, since
// the old walk answered false for every font asset in every document.
//
// It takes the removed ENTRIES rather than keys so a caller cannot pass a key
// it derived some other way; a face entry contributes no candidate at all,
// which Embedded() decides rather than the caller.
func dropUnnamedFontAssets(t *Template, removed ...template.FontChainEntry) {
	for _, entry := range removed {
		if !entry.Embedded() {
			continue
		}
		if assetKeyReferenced(t, entry.AssetKey) {
			continue
		}
		delete(t.doc.Assets, entry.AssetKey)
	}
}

// ---------------------------------------------------------------------------
// STORY 8.6: PICKING A FAMILY PUTS IT IN THE FILE.
//
// embedFontFamily is the command Story 8.5 left missing. Twenty-one catalogue
// typefaces reached the designer and NOTHING could author an embedded-face
// chain entry or write a font asset, so picking a family changed nothing about
// the document. addFontChainEntry only ever builds a FaceEntry, and `entries`
// arrives at addFontChain as a []string; there was no spelling in the command
// vocabulary that produced an AssetEntry.
//
// ONE COMMAND, ONE HISTORY ENTRY, ONE UNDO (AC1). Embedding the face and
// declaring the chain are one applyFontChainCommand transaction, which is what
// makes wasm.Engine.Apply's single pushUndo cover both: one undo takes the
// asset and the chain away together, and there is never a document that
// carries the bytes with nothing naming them or names a key the document does
// not carry.
//
// IT IS SHAPED ON setComponentAsset (D-5.13.1/D-5.13.3), the module's other
// asset-authoring command: a CLOSED payload the {op,value} property grammar
// cannot express, whose bytes GO alone decodes, bounds, hashes and admits —
// the browser hashes nothing and decides no legality. The one shape it adds is
// that this command also writes the RECORD, and refuses to embed a face whose
// caller cannot supply one.
//
// THE WRITER CAN NEVER PRODUCE A DOCUMENT ITS OWN PARSER WOULD REJECT. Story
// 8.6 made licence, licenceText and copyright REQUIRED of an asset a chain
// names (parse.go's requireEmbeddedFaceLicence). This command declares exactly
// such a chain, so it refuses the pick up front when any of the three is
// missing — with a located reason naming the field. Without that guard the
// refusal would still fire, but at the transaction's reparse, as the
// unlocated "font chains did not pass format validation": correct, and useless
// to whoever has to fix it.
func embedFontFamily(t *Template, raw map[string]json.RawMessage) error {
	if err := componentFields(raw, 12); err != nil {
		return err
	}
	name, err := fontChainName(raw, "name")
	if err != nil {
		return err
	}
	record, err := embeddedFontRecord(raw, name)
	if err != nil {
		return err
	}
	mediaType, err := commandString(raw, "mediaType")
	if err != nil {
		return componentFailure("", fontChainPath(name), err.Error())
	}
	decoded, err := embeddedFaceBytes(raw, name)
	if err != nil {
		return err
	}
	tail, err := embeddedFontTail(raw, name)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%x", sha256.Sum256(decoded))
	// RECOGNITION AND STRUCTURE AT THE COMMAND, never relying on the loader as
	// the catcher — setComponentAsset's rule (AC1 there) and the same reason:
	// bytes this build cannot read are refused before anything is written to
	// t.doc.Assets. An unrecognised container is refused HERE even though the
	// FORMAT accepts one (D-1.8.1 as amended, mediaType is an open set): a
	// hand-written `.folio` may carry a face this build cannot draw, but a
	// pick the designer makes must never produce one.
	if ferr := template.DecodeFontForRender(mediaType, decoded, template.FontChainSite{AssetKey: key, ChainName: name}); ferr != nil {
		return componentFailure("", fontChainPath(name), ferr.Error())
	}
	// AND THE ONE CLASS THAT GATE STRUCTURALLY CANNOT SEE (Story 16.0,
	// D-16.6). DecodeFontForRender's fence is "can this build read these bytes
	// as a single face, and nothing more", and checkSfnt beneath it never
	// inspects a table TAG — so a VARIABLE face is readable as a single face
	// and sailed straight through, into a `.folio` that saved cleanly and then
	// failed at render, where fontset.New refuses it. About a quarter of what
	// Google publishes is a variable build, so this was reachable by an
	// ordinary pick.
	//
	// The refusal is THE RENDERER'S OWN, not a second sentence written here:
	// fontset.RefuseVariableFace is the single function fontset.New also calls,
	// and its message already names the fonttools varLib.instancer remedy the
	// author needs. The renderer's guard is KEPT — a hand-written `.folio`
	// bypasses this command entirely — so this is an addition, never a move.
	//
	// It sits beside DecodeFontForRender and BEFORE anything reaches
	// t.doc.Assets, which is setComponentAsset's stated rule applied to the one
	// class it currently misses: "bytes this build cannot read are refused
	// before anything is written to t.doc.Assets."
	if verr := fontset.RefuseVariableFace(name, decoded); verr != nil {
		return componentFailure("", fontChainPath(name), verr.Error())
	}
	// AND THE CLASS NEITHER OF THOSE TWO CAN SEE: bytes that CONTRADICT the
	// licence being claimed for them (Story 16.1b, D-16.R.5/D-16.R.7).
	//
	// Until Epic 16 the tie between a face's declared SPDX id and its own
	// `name` table was a BUILD-TIME test over the 21 reviewed catalogue faces
	// (font-catalogue.test.ts:355-366). Epic 16 lets a face arrive from the
	// published library at the moment of a pick, so that gate stops covering
	// the population — and a runtime check that did not carry the tie would
	// make Epic 16 STRICTLY WEAKER than what it replaces, on the exact axis
	// D-8.6.5 already cost this project once (17 of 21 faces under another
	// project's terms, green until review).
	//
	// IT IS SITED HERE AND NOT IN THE BROWSER because the browser is not the
	// only door: embedFontFamily is reachable from wasm.Engine.Apply and from
	// a hand-authored command, which is the same gap fontset.go's `fvar`
	// refusal exists to cover. One implementation also covers the local tier.
	//
	// AND IT IS THE `fontset` PACKAGE'S OWN JUDGEMENT, not a second one
	// written here — this file stays a CALLER, never a checker, for the same
	// single-authority reason the `fvar` refusal above is one function. The
	// declared id comes from the record about to be written beside these
	// bytes, and the guard re-reads the bytes rather than trusting any other
	// field of the same payload (D-16.R.7: both wire fields move together, so
	// a check over them proves nothing).
	//
	// Beside the two gates above and BEFORE anything reaches t.doc.Assets.
	if lerr := fontset.RefuseContradictedLicence(name, record.Licence.Value, decoded); lerr != nil {
		return componentFailure("", fontChainPath(name), lerr.Error())
	}

	// DEDUPE BY CONTENT HASH (AC2). If ANY chain already names this key the
	// pick is already in the document: no second asset, no second chain, and
	// — because the canonical bytes then do not move — no second history
	// entry either (wasm.Engine.Apply's no-op short-circuit). The existing
	// chain is what the author is offered.
	if assetKeyReferenced(t, key) {
		return nil
	}
	if _, exists := t.doc.Fonts[name]; exists {
		return componentFailure("", fontChainPath(name), fmt.Sprintf("a font chain named %q already exists", name))
	}
	if len(t.doc.Fonts)+1 > maxCanvasFontFamilies {
		return componentFailure("", fontChainPath(name), "document declares more font chains than the projection bound")
	}
	if len(tail)+1 > maxCanvasFontChainEntries {
		return componentFailure("", fontChainPath(name), "a font chain declares more entries than the projection bound")
	}

	if t.doc.Assets == nil {
		t.doc.Assets = map[string]template.Asset{}
	}
	// INSERTED ONLY IF ABSENT. Re-picking a family whose chain was deleted
	// re-declares the chain over the asset the document already carries
	// rather than storing a second copy of it — the key IS the content, so
	// "already there" and "identical bytes" are the same question.
	if _, exists := t.doc.Assets[key]; !exists {
		// Re-wrapped canonically at 76 columns (AD-9) by writeAssets on the
		// way out, whatever shape it is held in here.
		t.doc.Assets[key] = template.Asset{
			MediaType: mediaType,
			Data:      []string{base64.StdEncoding.EncodeToString(decoded)},
			Font:      template.Presence[template.FontRecord]{Set: true, Value: record},
		}
	}
	if t.doc.Fonts == nil {
		t.doc.Fonts = template.Fonts{}
	}
	// THE PICKED FACE FIRST, THE PROPOSED TAIL BEHIND IT (AC3). The tail is
	// the shipped faces for the scripts the picked face does not cover; the
	// author edits it with the chain commands 8.1 already shipped, which is
	// why nothing here is privileged or locked.
	entries := make([]template.FontChainEntry, 0, len(tail)+1)
	entries = append(entries, template.AssetEntry(key))
	for _, face := range tail {
		entries = append(entries, template.FaceEntry(face))
	}
	t.doc.Fonts[name] = entries
	return nil
}

// embeddedFontRecord reads the six keys the document will record about the
// face. Three of them — licence, licenceText, copyright — are what parse.go
// REQUIRES of an asset a chain names; family, style and source are display and
// provenance and are required HERE for a different reason: this command is the
// designer's only door into the assets map, and a catalogue row that cannot
// say what the face is or where it came from is a row that should not ship a
// face into anybody's document.
func embeddedFontRecord(raw map[string]json.RawMessage, name string) (template.FontRecord, error) {
	var record template.FontRecord
	for _, field := range []struct {
		key string
		dst *template.Presence[string]
	}{
		{"family", &record.Family},
		{"style", &record.Style},
		{"licence", &record.Licence},
		{"licenceText", &record.LicenceText},
		{"copyright", &record.Copyright},
		{"source", &record.Source},
	} {
		value, err := commandString(raw, field.key)
		if err != nil {
			return template.FontRecord{}, componentFailure("", fontChainPath(name), err.Error())
		}
		// BLANK IS EMPTY HERE TOO, and it has to be: parse.go's
		// requireEmbeddedFaceLicence refuses whitespace-only terms, and a
		// command that admitted them would hand the transaction's reparse a
		// document its own parser rejects — a correct refusal, arriving as the
		// unlocated "font chains did not pass format validation". commandString
		// refuses "" and stops there, so the trim is this function's.
		if strings.TrimSpace(value) == "" {
			return template.FontRecord{}, componentFailure("", fontChainPath(name), "folio: "+field.key+" must be a non-empty string")
		}
		// The licence TEXT is the one field with no small legal value, so it
		// is bounded by the payload rather than by the projection string
		// bound the others share — a real OFL is ~4 KB and the projection
		// bound is 512.
		if field.key != "licenceText" && len(value) > maxCanvasPropertyString {
			return template.FontRecord{}, componentFailure("", fontChainPath(name), "embedded face "+field.key+" exceeds the projection bound")
		}
		*field.dst = template.Presence[string]{Set: true, Value: value}
	}
	return record, nil
}

// embeddedFaceBytes decodes and bounds the face, exactly as setComponentAsset
// does for a picture and against the same host-memory bound (D-5.13.4): the
// payload arrives through one engine protocol and one 8 MiB ceiling, so one
// derived byte bound covers both asset kinds. The largest catalogue face is
// under half a megabyte, so the bound is nowhere near the pick; it is here
// because a command that writes arbitrary bytes into a document needs one.
func embeddedFaceBytes(raw map[string]json.RawMessage, name string) ([]byte, error) {
	dataRaw, ok := raw["data"]
	if !ok {
		return nil, componentFailure("", fontChainPath(name), "face data is required")
	}
	var dataB64 string
	if json.Unmarshal(dataRaw, &dataB64) != nil || dataB64 == "" {
		return nil, componentFailure("", fontChainPath(name), "face data must be a non-empty base64 string")
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(dataB64)
	if err != nil {
		return nil, componentFailure("", fontChainPath(name), "face data must be valid base64")
	}
	if len(decoded) == 0 {
		return nil, componentFailure("", fontChainPath(name), "face data cannot be empty")
	}
	if len(decoded) > maxComponentAssetBytes {
		return nil, componentFailure("", fontChainPath(name), fmt.Sprintf("face exceeds the %d-byte supported size", maxComponentAssetBytes))
	}
	return decoded, nil
}

// embeddedFontTail reads the proposed fallback tail: the shipped face names
// that follow the picked face in the chain. It is a []string on the wire for
// the same reason addFontChain's `entries` is — every entry it can express is
// a FACE NAME, and the ONE entry that names an asset is the one this command
// builds itself from the bytes it just hashed. A caller cannot put a second
// asset entry in a chain by writing one down.
//
// AN EMPTY TAIL IS LEGAL: a face that covers every script the document renders
// needs no fallback behind it, and a chain of one embedded entry is a chain
// (TestLoadNeitherResolvesNorRefusesAnEmbeddedEntry).
func embeddedFontTail(raw map[string]json.RawMessage, name string) ([]string, error) {
	tailRaw, ok := raw["tail"]
	if !ok {
		return nil, componentFailure("", fontChainPath(name), "the proposed fallback tail is required — write [] for a face that needs none")
	}
	var tail []string
	if json.Unmarshal(tailRaw, &tail) != nil {
		return nil, componentFailure("", fontChainPath(name), "the fallback tail must be a string array")
	}
	for _, face := range tail {
		if face == "" {
			return nil, componentFailure("", fontChainPath(name), "a font chain entry must be a non-empty string")
		}
		if len(face) > maxCanvasPropertyString {
			return nil, componentFailure("", fontChainPath(name), "font chain entry exceeds the projection bound")
		}
	}
	return tail, nil
}

// fontChainBands is the three top-level band element lists in DOCUMENT ORDER
// (pageHeader, content, pageFooter) — the order the orphaning-delete refusal
// names its ids in.
func fontChainBands(t *Template) [][]template.Element {
	return [][]template.Element{t.doc.Bands.PageHeader.Elements, t.doc.Bands.Content.Elements, t.doc.Bands.PageFooter.Elements}
}

func fontChainNamedBy(style template.Presence[template.Style], name string) bool {
	return style.Set && !style.Null && style.Value.FontFamily.Set && !style.Value.FontFamily.Null && style.Value.FontFamily.Value == name
}

// fontChainReferences reports the ids of every element naming name, in
// document order. It is the SAFETY half of a delete, like assetKeyReferenced:
// under-reporting a reference here deletes a chain something still names, with
// no compile error to announce it.
//
// MEASURED, NOT ASSUMED: fontFamily has exactly TWO attachment points in
// template's model — Element.Style.FontFamily and
// Element.Table.HeaderStyle.FontFamily (model.go's Style is reached from
// nowhere else), and both are live at render (render.go's fontChain resolves
// the first, table_render.go's header-style resolver the second). Columns,
// footers and assets carry no Style. If a later story attaches a Style
// anywhere else, this walk needs that location added: like assetKeyReferenced,
// findComponent and addCanvasImagePaint, it enumerates the three band lists by
// hand because the module has no shared element enumerator.
func fontChainReferences(t *Template, name string) []string {
	var ids []string
	for _, elements := range fontChainBands(t) {
		for _, element := range elements {
			if fontChainNamedBy(element.Style, name) || (element.Table.Set && !element.Table.Null && fontChainNamedBy(element.Table.Value.HeaderStyle, name)) {
				ids = append(ids, string(element.ID))
			}
		}
	}
	return ids
}

// fontChainOrphanListReserve is the room fontChainOrphanMessage keeps, past
// the chain name, for what the message still has to say. The shortest of those
// endings is the "%d elements" fallback, and 32 bytes covers that for any id
// count a document can hold — so reserving it makes the budget below positive
// for EVERY name fontChainName accepts, which is the whole point.
const fontChainOrphanListReserve = 32

// fontChainOrphanMessage names the blocking elements in the refusal itself,
// which is the only place a LIST of ids can go: ComponentCommandError carries
// one ElementID. The host cuts the message at maxComponentFailureMessageBytes,
// so a list that would not fit is trimmed here on a whole-id boundary and
// closed with " and N more" rather than left to be cut mid-id.
func fontChainOrphanMessage(name string, ids []string) string {
	// THE NAME IS TRIMMED FIRST, and without this the guarantee above is not
	// merely weakened but inverted: fontChainName admits a name up to
	// maxCanvasPropertyString (512), which alone is the whole message width,
	// so a long-named chain drives the budget below NEGATIVE, every branch
	// overruns, and the host cuts the message wherever it lands — through the
	// middle of the name, or of an id, or of a rune. The trim is measured
	// against the FORMATTED prefix rather than against len(name) because %q
	// escapes, so a name of quotes or backslashes widens by more than its own
	// length; it cuts on a rune boundary and marks the cut.
	label, elision := name, ""
	prefix := fmt.Sprintf("font chain %q is still named by ", label)
	for len(prefix) > maxComponentFailureMessageBytes-fontChainOrphanListReserve && label != "" {
		label, elision = truncateAtRuneBoundary(label, len(label)-1), "…"
		prefix = fmt.Sprintf("font chain %q is still named by ", label+elision)
	}
	budget := maxComponentFailureMessageBytes - len(prefix)
	more := func(remaining int) string { return fmt.Sprintf(" and %d more", remaining) }
	list := ""
	for i, id := range ids {
		candidate := id
		if i > 0 {
			candidate = list + ", " + id
		}
		// A trimmed list must fit WITH its suffix; a complete one needs none.
		need := len(candidate)
		if i+1 < len(ids) {
			need += len(more(len(ids) - i - 1))
		}
		if need > budget {
			if i == 0 {
				break
			}
			return prefix + list + more(len(ids)-i)
		}
		list = candidate
	}
	if list == "" {
		// Not even one id fits beside a name this long: the count is all that
		// is left to say, and it is still true.
		return prefix + fmt.Sprintf("%d elements", len(ids))
	}
	return prefix + list
}
