package folio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/panitw/folio/folio-go/internal/geom"
	"github.com/panitw/folio/folio-go/internal/template"
)

// ComponentMove is engine-owned geometry evidence, in document millipoints.
type ComponentMove struct {
	DX int64 `json:"dx"`
	DY int64 `json:"dy"`
}

type groupMember struct {
	band         CanvasBand
	element      *template.Element
	windowOrigin geom.Length
	windowHeight geom.Length
}

// Project once per operation; looking up every member through findComponent
// would repeat pagination and text projection for each selected ID.
func groupMemberIndex(t *Template, fonts ...FontSet) (map[string]groupMember, error) {
	var projection CanvasProjection
	var err error
	if len(fonts) > 0 {
		projection, err = CanvasWithTextPaint(t, fonts[0])
	} else {
		projection, err = Canvas(t)
	}
	if err != nil {
		return nil, err
	}
	bands := map[string]*template.Band{bandPageHeader: &t.doc.Bands.PageHeader, bandContent: &t.doc.Bands.Content, bandPageFooter: &t.doc.Bands.PageFooter}
	members := make(map[string]groupMember, len(projection.Components))
	for _, band := range projection.Bands {
		for index := range bands[band.Name].Elements {
			element := &bands[band.Name].Elements[index]
			origin := geom.Length(0)
			windowHeight := geom.Length(projection.ContentWindowHeight)
			if band.Name == bandContent {
				for index, candidate := range projection.ContentWindowOrigins {
					if geom.Length(candidate) > element.Y {
						break
					}
					origin = geom.Length(candidate)
					windowHeight = geom.Length(projection.ContentWindowHeight)
					if index+1 < len(projection.ContentWindowOrigins) {
						windowHeight = min(windowHeight, geom.Length(projection.ContentWindowOrigins[index+1])-origin)
					}
				}
			}
			members[string(element.ID)] = groupMember{band: band, element: element, windowOrigin: origin, windowHeight: windowHeight}
		}
	}
	return members, nil
}

// PreviewComponentMove accepts exactly the atomic movement command vocabulary.
// It is read-only and shares its solver with the public mutation door.
// Supply the same FontSet used by CanvasWithTextPaint when constraining to
// displayed windows; without fonts the canvas only has its fallback window.
func PreviewComponentMove(t *Template, command []byte, fonts ...FontSet) (ComponentMove, error) {
	if t == nil {
		return ComponentMove{}, errNilTemplate
	}
	if err := refuseDuplicateCommandKeys(command, componentCommandPath); err != nil {
		return ComponentMove{}, err
	}
	var raw map[string]json.RawMessage
	d := json.NewDecoder(bytes.NewReader(command))
	var trailing any
	if d.Decode(&raw) != nil || d.Decode(&trailing) != io.EOF {
		return ComponentMove{}, fmt.Errorf("folio: group move is malformed")
	}
	_, move, err := solveComponentMove(t, raw, fonts...)
	return move, err
}

func solveComponentMove(t *Template, raw map[string]json.RawMessage, fonts ...FontSet) ([]string, ComponentMove, error) {
	fail := func(id, path, message string) ([]string, ComponentMove, error) {
		return nil, ComponentMove{}, componentFailure(id, path, message)
	}
	fields := 8
	constrain := false
	if value, present := raw["constrainToWindow"]; present {
		fields++
		if bytes.Equal(value, []byte("null")) || json.Unmarshal(value, &constrain) != nil {
			return fail("", "component.constrainToWindow", "constrainToWindow must be a boolean")
		}
	}
	if componentFields(raw, fields) != nil || !equalNumber(raw["version"], "1") || string(raw["kind"]) != `"moveComponents"` {
		return fail("", "component.move", "group move has unknown or missing fields")
	}
	var revision *uint64
	if json.Unmarshal(raw["expectedRevision"], &revision) != nil || revision == nil || *revision > uint64(MaxCanvasMillipoints) {
		return fail("", "component.move", "group move requires a safe revision")
	}
	var ids []string
	if json.Unmarshal(raw["ids"], &ids) != nil || len(ids) == 0 {
		return fail("", "component.ids", "group move requires component ids")
	}
	ref, err := commandString(raw, "referenceId")
	if err != nil || !slices.Contains(ids, ref) {
		return fail(ref, "component.referenceId", "the movement reference must belong to the selection")
	}
	snap, err := commandBool(raw, "snap")
	if err != nil || bytes.Equal(raw["snap"], []byte("null")) {
		return fail("", "component.snap", "snap must be a boolean")
	}
	dx, err := componentLength(raw, "dx", false)
	if err != nil {
		return fail("", "component.dx", err.Error())
	}
	dy, err := componentLength(raw, "dy", false)
	if err != nil {
		return fail("", "component.dy", err.Error())
	}
	bound := geom.Length(MaxCanvasMillipoints)
	minX, maxX, minY, maxY := -bound, bound, -bound, bound
	var refX, refY geom.Length
	seen := make(map[string]bool, len(ids))
	members, err := groupMemberIndex(t, fonts...)
	if err != nil {
		return nil, ComponentMove{}, err
	}
	for _, id := range ids {
		if seen[id] {
			return fail(id, "component.ids", "group move ids must be unique")
		}
		seen[id] = true
		member, found := members[id]
		if !found {
			return fail(id, "component.id", "component was not found")
		}
		band, element := member.band, member.element
		width, height := projectedSize(*element)
		if err := containComponent(band, element.X, element.Y, width, height); err != nil {
			return fail(id, "component.geometry", err.Error())
		}
		// Bound both origins and derived far edges before adding any delta.
		if element.X > bound-width || element.Y > bound-height {
			return fail(id, "component.geometry", "component exceeds safe geometry")
		}
		minX = max(minX, -element.X)
		maxX = min(maxX, geom.Length(band.Width)-width-element.X)
		minY = max(minY, -element.Y)
		maxY = min(maxY, bound-height-element.Y)
		if slices.Contains(bandsCappingVertically, band.Name) {
			maxY = min(maxY, geom.Length(band.Height)-height-element.Y)
		}
		if constrain && band.Name == bandContent {
			// Preserve an existing overflow rather than normalizing authored
			// geometry. Every range contains zero, including oversized/orphan
			// components, so a group always has a feasible common displacement.
			low := min(member.windowOrigin-element.Y, 0)
			high := max(member.windowOrigin+member.windowHeight-height-element.Y, 0)
			minY = max(minY, low)
			maxY = min(maxY, high)
		}
		if id == ref {
			refX, refY = element.X, element.Y
		}
	}
	// A returned-to-start gesture is a no-op even for an off-grid reference.
	if dx == 0 && dy == 0 {
		return ids, ComponentMove{}, nil
	}
	accept := func(proposed, low, high, origin geom.Length) geom.Length {
		legal := min(max(proposed, low), high)
		if !snap {
			return legal
		}
		grid := geom.Length(GridIncrement)
		// The reference's feasible origins are nonnegative. Find the first and
		// last grid points; clamp the nearest grid point into that interval.
		first := ((origin + low + grid - 1) / grid) * grid
		last := ((origin + high) / grid) * grid
		if first > last {
			return legal
		}
		nearest, _ := SnapToGrid(origin + legal)
		return min(max(nearest, first), last) - origin
	}
	return ids, ComponentMove{DX: int64(accept(dx, minX, maxX, refX)), DY: int64(accept(dy, minY, maxY, refY))}, nil
}

func moveComponents(t *Template, raw map[string]json.RawMessage, fonts ...FontSet) (CanvasProjection, error) {
	ids, move, err := solveComponentMove(t, raw, fonts...)
	if err != nil {
		return CanvasProjection{}, err
	}
	before, err := SerializeTemplate(t)
	if err != nil {
		return CanvasProjection{}, err
	}
	working, err := ParseTemplate(before)
	if err != nil {
		return CanvasProjection{}, err
	}
	members, err := groupMemberIndex(working)
	if err != nil {
		return CanvasProjection{}, err
	}
	for _, id := range ids {
		member := members[id]
		band, element := member.band, member.element
		x, y := element.X+geom.Length(move.DX), element.Y+geom.Length(move.DY)
		width, height := projectedSize(*element)
		if err := containComponent(band, x, y, width, height); err != nil {
			return CanvasProjection{}, componentFailure(id, "component.geometry", err.Error())
		}
		element.X, element.Y = x, y
	}
	canonical, err := SerializeTemplate(working)
	if err != nil {
		return CanvasProjection{}, err
	}
	installed, err := ParseTemplate(canonical)
	if err != nil {
		return CanvasProjection{}, err
	}
	projection, err := Canvas(installed)
	if err != nil {
		return CanvasProjection{}, err
	}
	t.doc, t.derivedFooters = installed.doc, installed.derivedFooters
	return projection, nil
}
