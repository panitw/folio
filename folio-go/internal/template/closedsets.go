package template

import (
	"regexp"
	"strings"
)

// The closed sets folio-format.md states, enforced at load (AC5). Kept
// visibly separate from passthrough (AC10, AC11 clause 3): an unknown
// KEY passes through opaquely; a recognised key carrying an unlisted
// VALUE from one of these sets is still a load error. Conflating the
// two paths is how passthrough silently becomes "validation is
// optional" — every check against one of these sets lives here, and
// nowhere near the passthrough code in parse.go.

// closedLocales moved to locale.go (Story 3.4, AC4): the closed set of
// valid `locale` tags is now built from named constants exported for
// internal/expr's locale table to key off, rather than a second copy
// of the tag literals.

var closedElementTypes = map[string]bool{
	"text": true, "image": true, "table": true, "line": true, "rect": true,
}

var closedPageOrientations = map[string]bool{
	"portrait": true, "landscape": true,
}

var closedPageSizeNames = map[string]bool{
	"A4": true, "Letter": true,
}

// THE ALIGNMENT VOCABULARY IS TWO SETS, NOT ONE (Story 7.3, D-7.3.1).
// Until this story both `style.align`/`headerStyle.align` and
// `columns[].align` validated against ONE shared map, so extending it
// with `justify` would have silently legalised justified TABLE CELLS —
// a scope decision nobody made, arriving as a side effect of a map
// edit. The two sets are therefore SEPARATE DECLARATIONS, and neither
// is derived from the other by subtraction at a call site.
//
// The tokens are named constants, and the ordered slices are built from
// THEM rather than from a second set of string literals — locale.go's
// LocaleTags/closedLocales precedent (D-1.3.5). The slices exist because
// ranging a map anywhere under internal/ is a BUILD FAILURE, and because
// the rejection MESSAGE is derived from the slice rather than restated
// as a literal: after Story 7.3 no message may claim a set of legal
// values that differs from the set actually enforced.
const (
	AlignLeft    = "left"
	AlignCenter  = "center"
	AlignRight   = "right"
	AlignJustify = "justify"
)

// StyleAlignTokens is the closed set `style.align` and
// `headerStyle.align` admit, in the exact order the load error lists
// them. `justify` (FR47) joins THIS set and only this set — extending a
// closed set is a MAJOR version change under D-1.4.12, which is why the
// format moved to 2.0 in the same story.
var StyleAlignTokens = []string{AlignLeft, AlignCenter, AlignRight, AlignJustify}

// ColumnAlignTokens is the closed set `columns[].align` admits. It keeps
// exactly the three values it has always had: justified table cells are
// a separate scope decision.
var ColumnAlignTokens = []string{AlignLeft, AlignCenter, AlignRight}

var closedStyleAligns = map[string]bool{
	AlignLeft: true, AlignCenter: true, AlignRight: true, AlignJustify: true,
}

var closedColumnAligns = map[string]bool{
	AlignLeft: true, AlignCenter: true, AlignRight: true,
}

// IsStyleAlign reports whether s is a member of the STYLE alignment set.
// Exported for the property-command path (component_commands.go), which
// sets style.align from a command and must validate it against the same
// single source the loader does rather than against a second literal.
func IsStyleAlign(s string) bool { return closedStyleAligns[s] }

// closedSetMessage renders an ordered token slice as the tail of a
// closed-set load error. Derived from the slice, never hand-written, so
// a set that gains or loses a member cannot ship a message that lies
// about what is legal (internal/expr/locale.go:lookupLocale's precedent,
// which formats template.LocaleTags into its own message the same way).
func closedSetMessage(tokens []string) string {
	return "not one of the closed set " + strings.Join(tokens, ", ")
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
