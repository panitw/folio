// Package wasm is the imperative browser shell around the pure Folio core.
// It intentionally owns mutable browser-session state here, never under
// internal/, and exposes only bytes plus a deliberately small UI projection.
package wasm

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	folio "github.com/panitw/folio/folio-go"
	"github.com/panitw/folio/folio-go/fonts"
)

const historyLimit = 100

var ErrNoUndo = errors.New("folio wasm: no undo history")
var ErrNoRedo = errors.New("folio wasm: no redo history")

// Snapshot is a paint-safe projection, not a .folio schema mirror.
type Snapshot struct {
	DocumentState string                  `json:"documentState"`
	Revision      uint64                  `json:"revision"`
	ByteLength    int                     `json:"byteLength"`
	CanUndo       bool                    `json:"canUndo"`
	CanRedo       bool                    `json:"canRedo"`
	Canvas        *folio.CanvasProjection `json:"canvas,omitempty"`
}

type TableColumnsResult struct {
	Revision uint64                       `json:"revision"`
	Table    folio.TableColumnsProjection `json:"table"`
}

// RenderResult is a deliberately opaque production-render projection. It is
// never a browser document model: callers can only receive the PDF bytes,
// their producer-computed digest, the revision that supplied template bytes,
// and bounded diagnostics from the existing production renderer.
type RenderResult struct {
	PDFSHA256   string             `json:"pdfSha256"`
	Identity    string             `json:"identity"`
	Revision    uint64             `json:"revision"`
	Diagnostics []folio.Diagnostic `json:"diagnostics"`
}

// ParameterReferences exposes only engine-derived display metadata for the
// transient Preview parameter editor. It is neither document state nor a
// template/schema projection.
func (e *Engine) ParameterReferences() ([]string, uint64, error) {
	if e.template == nil {
		return nil, 0, fmt.Errorf("folio wasm: no document is loaded")
	}
	references, err := folio.ParameterReferences(e.template)
	if err != nil {
		return nil, 0, err
	}
	// The browser protocol's closed parameterReferences shape requires an
	// array. A nil Go slice marshals as null and would invalidate an otherwise
	// healthy worker response when the document has no params references.
	out := make([]string, len(references))
	copy(out, references)
	return out, e.revision, nil
}

// TableColumns exposes one revision-correlated selected-table projection.
// It is intentionally a query, not a browser-side document model.
func (e *Engine) TableColumns(tableID string) (TableColumnsResult, error) {
	if e.template == nil {
		return TableColumnsResult{}, fmt.Errorf("folio wasm: no document is loaded")
	}
	table, err := folio.TableColumns(e.template, tableID)
	if err != nil {
		return TableColumnsResult{}, err
	}
	return TableColumnsResult{Revision: e.revision, Table: table}, nil
}

// PreviewIdentity obtains evidence from the current engine-owned canonical
// template and the two raw JSON channels without exposing or parsing the
// template in the browser.
func (e *Engine) PreviewIdentity(data, params []byte) (string, uint64, error) {
	if e.template == nil {
		return "", 0, fmt.Errorf("folio wasm: no document is loaded")
	}
	if len(data) == 0 || len(params) == 0 {
		return "", 0, fmt.Errorf("folio wasm: identity inputs must be non-empty")
	}
	return folio.PreviewIdentity(e.bytes, folio.Data(data), folio.Params(params), fonts.Shipped()), e.revision, nil
}

// Engine owns one live template and its canonical bytes for one worker.
type Engine struct {
	template *folio.Template
	bytes    []byte
	revision uint64
	canvas   *folio.CanvasProjection
	undo     [][]byte
	redo     [][]byte
}

func NewEngine() *Engine { return &Engine{} }

// Initialize and Load parse through the public engine boundary before storing
// a canonical copy. A caller's byte slice is never retained or aliased.
func (e *Engine) Initialize(input []byte) (Snapshot, error) { return e.load(input) }
func (e *Engine) Load(input []byte) (Snapshot, error)       { return e.load(input) }

func (e *Engine) load(input []byte) (Snapshot, error) {
	tpl, err := folio.ParseTemplate(input)
	if err != nil {
		return Snapshot{}, err
	}
	canonical, err := folio.SerializeTemplate(tpl)
	if err != nil {
		return Snapshot{}, err
	}
	projection, err := folio.CanvasWithTextPaint(tpl, fonts.Shipped())
	if err != nil {
		return Snapshot{}, err
	}
	// Install only after every fallible projection step has completed. Failed
	// Open/Initialize must leave the old template, bytes, snapshot and revision
	// untouched so a later Save cannot serialize rejected input.
	e.template = tpl
	e.bytes = append(e.bytes[:0], canonical...)
	e.canvas = &projection
	e.undo = nil
	e.redo = nil
	e.revision++
	return e.Snapshot(), nil
}

func (e *Engine) Snapshot() Snapshot {
	if e.template == nil {
		return Snapshot{DocumentState: "empty", Revision: e.revision, CanUndo: len(e.undo) > 0, CanRedo: len(e.redo) > 0}
	}
	return Snapshot{DocumentState: "loaded", Revision: e.revision, ByteLength: len(e.bytes), CanUndo: len(e.undo) > 0, CanRedo: len(e.redo) > 0, Canvas: e.canvas}
}

// Serialize returns a copy of canonical bytes. Bytes are the authority across
// the worker boundary; no live document handle is exposed.
func (e *Engine) Serialize() ([]byte, Snapshot, error) {
	if e.template == nil {
		return nil, Snapshot{}, fmt.Errorf("folio wasm: no document is loaded")
	}
	return append([]byte(nil), e.bytes...), e.Snapshot(), nil
}

// Render reparses the caller-provided canonical template bytes through the
// production boundary and invokes the one public renderer. The three byte
// channels intentionally remain distinct: data and params must preserve the
// exact-decimal JSON semantics owned by folio.Render.
func (e *Engine) Render(template, data, params []byte) ([]byte, RenderResult, error) {
	if e.template == nil {
		return nil, RenderResult{}, fmt.Errorf("folio wasm: no document is loaded")
	}
	if len(template) == 0 || len(data) == 0 || len(params) == 0 {
		return nil, RenderResult{}, fmt.Errorf("folio wasm: render inputs must be non-empty")
	}
	if !bytes.Equal(template, e.bytes) {
		return nil, RenderResult{}, fmt.Errorf("folio wasm: render template is not the current canonical revision")
	}
	tpl, err := folio.ParseTemplate(template)
	if err != nil {
		return nil, RenderResult{}, err
	}
	result, err := folio.Render(tpl, folio.Data(data), folio.Params(params), fonts.Shipped())
	if err != nil {
		return nil, RenderResult{}, err
	}
	pdf := append([]byte(nil), result.Bytes...)
	digest := sha256.Sum256(pdf)
	identity, revision, err := e.PreviewIdentity(data, params)
	if err != nil {
		return nil, RenderResult{}, err
	}
	return pdf, RenderResult{PDFSHA256: fmt.Sprintf("%x", digest), Identity: identity, Revision: revision, Diagnostics: append([]folio.Diagnostic(nil), result.Diagnostics...)}, nil
}

// AssetBytes is Story 5.13's per-key paintable-bytes query (D-5.13.2's
// "Producer" clause). It is read-only: it never advances revision or
// touches undo/redo history, and it never reproduces asset lookup/decoding
// rules here — folio.AssetBytes owns those.
func (e *Engine) AssetBytes(key string) ([]byte, Snapshot, error) {
	if e.template == nil {
		return nil, Snapshot{}, fmt.Errorf("folio wasm: no document is loaded")
	}
	raw, _, err := folio.AssetBytes(e.template, key)
	if err != nil {
		return nil, Snapshot{}, err
	}
	return raw, e.Snapshot(), nil
}

// Validate reparses the engine-owned canonical bytes. It deliberately does
// not reproduce validation rules in the transport layer.
func (e *Engine) Validate() (Snapshot, error) {
	if e.template == nil {
		return Snapshot{}, fmt.Errorf("folio wasm: no document is loaded")
	}
	if _, err := folio.ParseTemplate(e.bytes); err != nil {
		return Snapshot{}, err
	}
	return e.Snapshot(), nil
}

// Apply accepts only opaque, engine-defined committed commands.
func (e *Engine) Apply(command []byte) (Snapshot, error) {
	if e.template == nil {
		return Snapshot{}, fmt.Errorf("folio wasm: no document is loaded")
	}
	// Apply to a fresh canonical clone. This makes command validation,
	// serialization and projection one transaction rather than relying on a
	// rollback that can itself alter the revision.
	candidate, err := folio.ParseTemplate(e.bytes)
	if err != nil {
		return Snapshot{}, err
	}
	var commandKind struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(command, &commandKind); err != nil {
		return Snapshot{}, fmt.Errorf("folio wasm: command is malformed")
	}
	var projection folio.CanvasProjection
	if commandKind.Kind == "pageSetup" {
		projection, err = folio.ApplyPageSetupCommand(candidate, command)
	} else {
		projection, err = folio.ApplyComponentCommand(candidate, command)
	}
	if err != nil {
		return Snapshot{}, err
	}
	canonical, err := folio.SerializeTemplate(candidate)
	if err != nil {
		return Snapshot{}, err
	}
	// Some closed commands are valid but leave canonical bytes unchanged (for
	// example, applying the page setup already in force). They are not committed
	// mutations: preserve revision, dirty state, preview authority, and both
	// history branches exactly as they were.
	if bytes.Equal(canonical, e.bytes) {
		return e.Snapshot(), nil
	}
	// Reparse the candidate's canonical bytes before installation. Command
	// factories therefore cannot bypass the normal format validator, and the
	// persisted bytes, live template, and paint projection all describe the
	// same accepted document.
	installed, err := folio.ParseTemplate(canonical)
	if err != nil {
		return Snapshot{}, err
	}
	projection, err = folio.CanvasWithTextPaint(installed, fonts.Shipped())
	if err != nil {
		return Snapshot{}, err
	}
	e.pushUndo(e.bytes)
	e.redo = nil
	e.install(installed, canonical, projection)
	return e.Snapshot(), nil
}

// Undo and Redo replay canonical engine bytes from this live wasm session.
// They never serialize history into .folio bytes or ask TypeScript to retain
// a mirror/inverse command. Revisions stay monotonic even when document bytes
// return to an earlier state, which keeps preview authority correlation sound.
func (e *Engine) Undo() (Snapshot, error) {
	if len(e.undo) == 0 {
		return e.Snapshot(), ErrNoUndo
	}
	prior := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]
	e.pushRedo(e.bytes)
	return e.restore(prior)
}

func (e *Engine) Redo() (Snapshot, error) {
	if len(e.redo) == 0 {
		return e.Snapshot(), ErrNoRedo
	}
	next := e.redo[len(e.redo)-1]
	e.redo = e.redo[:len(e.redo)-1]
	e.pushUndo(e.bytes)
	return e.restore(next)
}

func (e *Engine) restore(canonical []byte) (Snapshot, error) {
	tpl, err := folio.ParseTemplate(canonical)
	if err != nil {
		return Snapshot{}, err
	}
	projection, err := folio.CanvasWithTextPaint(tpl, fonts.Shipped())
	if err != nil {
		return Snapshot{}, err
	}
	e.install(tpl, canonical, projection)
	return e.Snapshot(), nil
}

func (e *Engine) install(tpl *folio.Template, canonical []byte, projection folio.CanvasProjection) {
	e.template = tpl
	e.bytes = append(e.bytes[:0], canonical...)
	e.canvas = &projection
	e.revision++
}

func (e *Engine) pushUndo(value []byte) { e.undo = appendBounded(e.undo, value) }
func (e *Engine) pushRedo(value []byte) { e.redo = appendBounded(e.redo, value) }
func appendBounded(history [][]byte, value []byte) [][]byte {
	copyValue := append([]byte(nil), value...)
	if len(history) == historyLimit {
		copy(history, history[1:])
		history[len(history)-1] = copyValue
		return history
	}
	return append(history, copyValue)
}
