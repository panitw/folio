//go:build js && wasm

package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"

	folio "github.com/panitw/folio/folio-go"
	"github.com/panitw/folio/folio-go/internal/text"
	"github.com/panitw/folio/folio-go/wasm"
)

// request is intentionally byte-oriented. The JavaScript boundary never
// reads document values as JS numbers or recreates document fields.
type request struct {
	Operation     string `json:"operation"`
	PayloadBase64 string `json:"payloadBase64,omitempty"`
}

type response struct {
	OK               bool          `json:"ok"`
	Snapshot         wasm.Snapshot `json:"snapshot,omitempty"`
	BytesBase64      string        `json:"bytesBase64,omitempty"`
	DiagnosticCode   string        `json:"diagnosticCode,omitempty"`
	Message          string        `json:"message,omitempty"`
	ElementID        string        `json:"elementId,omitempty"`
	DataPath         string        `json:"dataPath,omitempty"`
	DictionarySHA256 string        `json:"dictionarySha256,omitempty"`
}

func main() {
	engine := wasm.NewEngine()
	handle := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return marshal(response{DiagnosticCode: "WASM_PROTOCOL_INVALID", Message: "expected one JSON request string"})
		}
		var in request
		if err := json.Unmarshal([]byte(args[0].String()), &in); err != nil {
			return marshal(response{DiagnosticCode: "WASM_PROTOCOL_INVALID", Message: "malformed request"})
		}
		out := dispatch(engine, in)
		return marshal(out)
	})
	js.Global().Set("FolioWasmHost", js.ValueOf(map[string]any{"handle": handle}))
	select {}
}

func dispatch(engine *wasm.Engine, in request) response {
	decode := func() ([]byte, error) {
		if len(in.PayloadBase64) > 8<<20 {
			return nil, fmt.Errorf("payload exceeds 8 MiB")
		}
		return base64.StdEncoding.DecodeString(in.PayloadBase64)
	}
	switch in.Operation {
	case "offline-audit":
		return response{OK: true, DictionarySHA256: text.DictionarySHA256()}
	case "initialize", "load":
		payload, err := decode()
		if err != nil {
			return failure("WASM_INPUT_INVALID", err)
		}
		var snapshot wasm.Snapshot
		if in.Operation == "initialize" {
			snapshot, err = engine.Initialize(payload)
		} else {
			snapshot, err = engine.Load(payload)
		}
		if err != nil {
			return engineFailure(err)
		}
		return response{OK: true, Snapshot: snapshot}
	case "snapshot":
		return response{OK: true, Snapshot: engine.Snapshot()}
	case "validate":
		snapshot, err := engine.Validate()
		if err != nil {
			return engineFailure(err)
		}
		return response{OK: true, Snapshot: snapshot}
	case "serialize":
		bytes, snapshot, err := engine.Serialize()
		if err != nil {
			return engineFailure(err)
		}
		return response{OK: true, Snapshot: snapshot, BytesBase64: base64.StdEncoding.EncodeToString(bytes)}
	case "command":
		payload, err := decode()
		if err != nil {
			return failure("WASM_INPUT_INVALID", err)
		}
		snapshot, err := engine.Apply(payload)
		if err != nil {
			return engineFailure(err)
		}
		return response{OK: true, Snapshot: snapshot}
	default:
		return response{DiagnosticCode: "WASM_OPERATION_UNKNOWN", Message: "unknown operation"}
	}
}

func failure(code string, err error) response {
	return response{DiagnosticCode: code, Message: "The engine request was invalid"}
}

func engineFailure(err error) response {
	var renderErr *folio.RenderError
	if errors.As(err, &renderErr) {
		diagnostic := renderErr.Diagnostic
		return response{
			DiagnosticCode: diagnostic.Code,
			Message:        "The template could not be processed",
			ElementID:      bounded(diagnostic.ElementID, 128),
			DataPath:       bounded(diagnostic.DataPath, 256),
		}
	}
	return response{DiagnosticCode: "ENGINE_REJECTED", Message: "The engine rejected the request"}
}

func bounded(value string, max int) string {
	if len(value) > max {
		return value[:max]
	}
	return value
}
func marshal(out response) string { bytes, _ := json.Marshal(out); return string(bytes) }
