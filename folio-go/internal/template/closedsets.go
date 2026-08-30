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

// THE ALIGNMENT VOCABULARY IS THREE SETS, PARTITIONED BY CONSUMER
// (Story 7.3 D-7.3.1, re-partitioned at Story 7.8).
//
// Before Story 7.3 one shared map served every align field, so
// extending it with `justify` would have silently legalised justified
// TABLE CELLS — a scope decision nobody made, arriving as a side effect
// of a map edit. Story 7.3 split it in two.
//
// IT SPLIT IT ALONG THE WRONG SEAM, and that is the reusable lesson.
// The split was BY JSON KEY LOCATION — `style`/`headerStyle` on one
// side, `columns[]` on the other — but a table's `style.align` and its
// `columns[].align` are read into the SAME variable
// (table_render.go's r.alignFallback) and consumed at the SAME four
// switches. They are one consumer wearing two key paths, so the
// guardrail meant to make justified table cells impossible let the
// value straight in through the other door: a table declaring
// `style.align: "justify"` loaded, paid a MAJOR version bump, and drew
// every cell at the start edge with no diagnostic (DW-29).
//
//	WHEN SPLITTING A CLOSED SET, PARTITION IT BY THE CODE THAT
//	CONSUMES THE VALUE, NOT BY WHERE THE VALUE IS WRITTEN IN THE
//	DOCUMENT.
//
// So the sets are keyed on the CONSUMER, which here means the ELEMENT
// TYPE that owns the style block:
//
//   - StyleAlignTokens — a NON-TABLE element's own `style.align`. The
//     paragraph justifier consumes it, and `justify` (FR47) means
//     something there.
//   - TableStyleAlignTokens — a TABLE's `style.align` and
//     `headerStyle.align`. The table cell renderer consumes it, through
//     exactly the same alignFallback as the column set, so it admits
//     exactly what that consumer can draw.
//   - ColumnAlignTokens — `columns[].align`. Same consumer, same three
//     values; kept as its own declaration because it is a separately
//     documented closed set in the format.
//
// The three are SEPARATE DECLARATIONS and none is derived from another
// by subtraction at a call site.
//
// The tokens are named constants, and the ordered slices are built from
// THEM rather than from a second set of string literals — locale.go's
// LocaleTags/closedLocales precedent (D-1.3.5). The slices exist because
// ranging a map anywhere under internal/ is a BUILD FAILURE, and because
// the rejection MESSAGE is derived from the slice rather than restated
// as a literal: after Story 7.3 no message may claim a set of legal
// values that differs from the set actually enforced. That invariant is
// what forced a third SET rather than a type-guard bolted above the
// existing two — under a bare guard a table's `align: "middle"` would
// still be rejected by the style map and would report `left, center,
// right, justify`, naming as legal a value that element can never carry.
const (
	AlignLeft    = "left"
	AlignCenter  = "center"
	AlignRight   = "right"
	AlignJustify = "justify"
)

// StyleAlignTokens is the closed set a NON-TABLE element's own
// `style.align` admits, in the exact order the load error lists them.
// `justify` (FR47) is in THIS set and no other — extending a closed set
// is a MAJOR version change under D-1.4.12, which is why the format
// moved to 2.0 in the same story, and it is why declaring it is the one
// thing in the format that raises a document to 2.0.
//
// It is NOT what `headerStyle.align` admits, and not what a table's own
// `style.align` admits: those two attachment points belong to a table,
// whose consumer is TableStyleAlignTokens below (Story 7.8). The
// sentence this doc comment carried until then — "the closed set
// `style.align` and `headerStyle.align` admit" — described the key
// path, which is exactly the partition that shipped the defect.
var StyleAlignTokens = []string{AlignLeft, AlignCenter, AlignRight, AlignJustify}

// TableStyleAlignTokens is the closed set a TABLE's `style.align` and
// `headerStyle.align` admit (Story 7.8). It is the column triple, and
// it is the column triple for a reason that is not coincidence: those
// two attachment points and `columns[].align` are read into one
// alignFallback and consumed at one set of switches, so they must admit
// exactly what that consumer can draw. Justified table cells remain a
// separate scope decision, and this is the set that actually makes them
// impossible.
//
// A separate declaration from ColumnAlignTokens rather than an alias of
// it: they are separately documented closed sets in folio-format.md,
// and either could move without the other. TestAlignSetsAreThreeSets…
// pins them against their own maps and against each other.
var TableStyleAlignTokens = []string{AlignLeft, AlignCenter, AlignRight}

// ColumnAlignTokens is the closed set `columns[].align` admits. It keeps
// exactly the three values it has always had: justified table cells are
// a separate scope decision.
var ColumnAlignTokens = []string{AlignLeft, AlignCenter, AlignRight}

var closedStyleAligns = map[string]bool{
	AlignLeft: true, AlignCenter: true, AlignRight: true, AlignJustify: true,
}

var closedTableStyleAligns = map[string]bool{
	AlignLeft: true, AlignCenter: true, AlignRight: true,
}

var closedColumnAligns = map[string]bool{
	AlignLeft: true, AlignCenter: true, AlignRight: true,
}

// IsStyleAlign reports whether s is a member of a NON-TABLE element's
// style alignment set. Exported for the property-command path
// (component_commands.go), which sets style.align from a command and
// must validate it against the same single source the loader does
// rather than against a second literal.
func IsStyleAlign(s string) bool { return closedStyleAligns[s] }

// IsTableStyleAlign reports whether s is a member of a TABLE's style
// alignment set — its own `style.align` and its `headerStyle.align`.
// Exported for the same reason and under the same obligation as
// IsStyleAlign: the property-command path selects between the two on
// the element's type, exactly as decodeStyle does, so the command door
// and the file door cannot admit different values (Story 7.8). Without
// it the designer's engine could still stamp a table document 2.0 with
// a value the loader would then refuse to reopen.
func IsTableStyleAlign(s string) bool { return closedTableStyleAligns[s] }

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
