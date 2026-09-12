package folio

import (
	"fmt"
	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/template"
)

// AuthoredProperty is bounded inspector evidence, separate from paint. It
// preserves authored absence/null and values even when a border paints no ink.
type AuthoredProperty[T any] struct {
	State string `json:"state"`
	Value *T     `json:"value,omitempty"`
}

type CanvasAuthoredProperties struct {
	VisibleIf   AuthoredProperty[string]      `json:"visibleIf"`
	FontFamily  AuthoredProperty[string]      `json:"fontFamily"`
	FontSize    AuthoredProperty[geom.Length] `json:"fontSize"`
	LineSpacing AuthoredProperty[int64]       `json:"lineSpacing"`
	Bold        AuthoredProperty[bool]        `json:"bold"`
	Italic      AuthoredProperty[bool]        `json:"italic"`
	Align       AuthoredProperty[string]      `json:"align"`
	Valign      AuthoredProperty[string]      `json:"valign"`
	Color       AuthoredProperty[string]      `json:"color"`
	Background  AuthoredProperty[string]      `json:"background"`
	BorderWidth AuthoredProperty[geom.Length] `json:"borderWidth"`
	BorderColor AuthoredProperty[string]      `json:"borderColor"`
	BorderEdges AuthoredProperty[[]string]    `json:"borderEdges"`
}

func authoredProperty[T any](value template.Presence[T], parentNull bool) AuthoredProperty[T] {
	if parentNull || value.Set && value.Null {
		return AuthoredProperty[T]{State: "null"}
	}
	if !value.Set {
		return AuthoredProperty[T]{State: "absent"}
	}
	return AuthoredProperty[T]{State: "value", Value: &value.Value}
}

func canvasAuthoredProperties(element template.Element) (*CanvasAuthoredProperties, error) {
	style := element.Style.Value
	border := style.Border.Value
	styleNull := element.Style.Set && element.Style.Null
	borderNull := styleNull || style.Border.Set && style.Border.Null
	// Empty edges are a authored value, not a nil array on the wire.
	if border.Edges.Set && !border.Edges.Null && border.Edges.Value == nil {
		border.Edges.Value = []string{}
	}
	result := &CanvasAuthoredProperties{
		VisibleIf:   authoredProperty(element.VisibleIf, false),
		FontFamily:  authoredProperty(style.FontFamily, styleNull),
		FontSize:    authoredProperty(style.FontSize, styleNull),
		LineSpacing: authoredProperty(style.LineSpacing, styleNull),
		Bold:        authoredProperty(style.Bold, styleNull),
		Italic:      authoredProperty(style.Italic, styleNull),
		Align:       authoredProperty(style.Align, styleNull),
		Valign:      authoredProperty(style.Valign, styleNull),
		Color:       authoredProperty(style.Color, styleNull),
		Background:  authoredProperty(style.Background, styleNull),
		BorderWidth: authoredProperty(border.Width, borderNull),
		BorderColor: authoredProperty(border.Color, borderNull),
		BorderEdges: authoredProperty(border.Edges, borderNull),
	}
	if element.Type != template.ElementText && element.Type != template.ElementTable {
		result.FontFamily = AuthoredProperty[string]{State: "absent"}
		result.FontSize = AuthoredProperty[geom.Length]{State: "absent"}
		result.LineSpacing = AuthoredProperty[int64]{State: "absent"}
		result.Bold = AuthoredProperty[bool]{State: "absent"}
		result.Italic = AuthoredProperty[bool]{State: "absent"}
		result.Align = AuthoredProperty[string]{State: "absent"}
		result.Valign = AuthoredProperty[string]{State: "absent"}
		result.Color = AuthoredProperty[string]{State: "absent"}
	}
	for _, field := range []AuthoredProperty[string]{result.VisibleIf, result.FontFamily, result.Align, result.Valign, result.Color, result.Background, result.BorderColor} {
		if field.Value != nil && len(*field.Value) > maxCanvasPropertyString {
			return nil, fmt.Errorf("folio: authored property exceeds projection bound")
		}
	}
	for _, field := range []AuthoredProperty[geom.Length]{result.FontSize, result.BorderWidth} {
		if field.Value != nil && (*field.Value < -geom.Length(MaxCanvasMillipoints) || *field.Value > geom.Length(MaxCanvasMillipoints)) {
			return nil, fmt.Errorf("folio: authored property exceeds safe geometry")
		}
	}
	return result, nil
}
