package template

import "regexp"

// The closed sets folio-format.md states, enforced at load (AC5). Kept
// visibly separate from passthrough (AC10, AC11 clause 3): an unknown
// KEY passes through opaquely; a recognised key carrying an unlisted
// VALUE from one of these sets is still a load error. Conflating the
// two paths is how passthrough silently becomes "validation is
// optional" — every check against one of these sets lives here, and
// nowhere near the passthrough code in parse.go.

var closedLocales = map[string]bool{
	"en": true, "th": true, "zh-Hans": true, "ja": true,
}

var closedElementTypes = map[string]bool{
	"text": true, "image": true, "table": true, "line": true, "rect": true,
}

var closedPageOrientations = map[string]bool{
	"portrait": true, "landscape": true,
}

var closedPageSizeNames = map[string]bool{
	"A4": true, "Letter": true,
}

var closedAligns = map[string]bool{
	"left": true, "center": true, "right": true,
}

var closedValigns = map[string]bool{
	"top": true, "middle": true, "bottom": true,
}

var closedBorderEdges = map[string]bool{
	"top": true, "right": true, "bottom": true, "left": true,
}

var closedFooterKinds = map[string]bool{
	"sum": true, "count": true, "avg": true,
}

// utcOffsetPattern is ±HH:MM (AC5).
var utcOffsetPattern = regexp.MustCompile(`^[+-][0-9]{2}:[0-9]{2}$`)
