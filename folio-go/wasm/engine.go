// Package wasm is the imperative browser shell around the pure Folio core.
// It intentionally owns mutable browser-session state here, never under
// internal/, and exposes only bytes plus a deliberately small UI projection.
package wasm

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	folio "github.com/panitw/folio/folio-go"
	"github.com/panitw/folio/folio-go/fonts"
)

// Snapshot is a paint-safe projection, not a .folio schema mirror.
type Snapshot struct {
	DocumentState string                  `json:"documentState"`
	Revision      uint64                  `json:"revision"`
	ByteLength    int                     `json:"byteLength"`
	Canvas        *folio.CanvasProjection `json:"canvas,omitempty"`
}

// RenderResult is a deliberately opaque production-render projection. It is
// never a browser document model: callers can only receive the PDF bytes,
// their producer-computed digest, the revision that supplied template bytes,
// and bounded diagnostics from the existing production renderer.
type RenderResult struct {
	PDFSHA256   string             `json:"pdfSha256"`
	Revision    uint64             `json:"revision"`
	Diagnostics []folio.Diagnostic `json:"diagnostics"`
}

// Engine owns one live template and its canonical bytes for one worker.
type Engine struct {
	template *folio.Template
	bytes    []byte
	revision uint64
	canvas   *folio.CanvasProjection
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
	e.revision++
	return e.Snapshot(), nil
}

func (e *Engine) Snapshot() Snapshot {
	if e.template == nil {
		return Snapshot{DocumentState: "empty", Revision: e.revision}
	}
	return Snapshot{DocumentState: "loaded", Revision: e.revision, ByteLength: len(e.bytes), Canvas: e.canvas}
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
	return pdf, RenderResult{PDFSHA256: fmt.Sprintf("%x", digest), Revision: e.revision, Diagnostics: append([]folio.Diagnostic(nil), result.Diagnostics...)}, nil
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
	e.template = installed
	e.bytes = append(e.bytes[:0], canonical...)
	e.canvas = &projection
	e.revision++
	return e.Snapshot(), nil
}
