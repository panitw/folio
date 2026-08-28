// Package wasm is the imperative browser shell around the pure Folio core.
// It intentionally owns mutable browser-session state here, never under
// internal/, and exposes only bytes plus a deliberately small UI projection.
package wasm

import (
	"fmt"

	folio "github.com/panitw/folio/folio-go"
)

// Snapshot is a paint-safe projection, not a .folio schema mirror.
type Snapshot struct {
	DocumentState string                  `json:"documentState"`
	Revision      uint64                  `json:"revision"`
	ByteLength    int                     `json:"byteLength"`
	Canvas        *folio.CanvasProjection `json:"canvas,omitempty"`
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
	projection, err := folio.Canvas(tpl)
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

// Apply is reserved for opaque, engine-defined committed commands. Component
// mutation semantics intentionally arrive in later stories, so accepting an
// unknown payload is forbidden rather than inventing a TypeScript schema.
func (e *Engine) Apply(command []byte) (Snapshot, error) {
	if e.template == nil {
		return Snapshot{}, fmt.Errorf("folio wasm: no document is loaded")
	}
	if string(command) == "commit" {
		// "commit" is deliberately an opaque, Go-owned acknowledgement today.
		// It provides the real command-channel boundary without inventing the
		// later editor command/schema vocabulary in TypeScript.
		e.revision++
		return e.Snapshot(), nil
	}
	// Apply to a fresh canonical clone. This makes command validation,
	// serialization and projection one transaction rather than relying on a
	// rollback that can itself alter the revision.
	candidate, err := folio.ParseTemplate(e.bytes)
	if err != nil {
		return Snapshot{}, err
	}
	if _, err := folio.ApplyPageSetupCommand(candidate, command); err != nil {
		return Snapshot{}, err
	}
	canonical, err := folio.SerializeTemplate(candidate)
	if err != nil {
		return Snapshot{}, err
	}
	projection, err := folio.Canvas(candidate)
	if err != nil {
		return Snapshot{}, err
	}
	e.template = candidate
	e.bytes = append(e.bytes[:0], canonical...)
	e.canvas = &projection
	e.revision++
	return e.Snapshot(), nil
}
