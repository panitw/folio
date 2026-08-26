// Package diag is AD-14's closed code registry (Story 3.6, D-3.6.4,
// ruled arm A): a Code defined type, the code constants, and the
// REGISTRY VALUE itself — a constructed value the package builds, not
// a bare const block (D-1.4.2's decision-log row `:9118`: "assert
// membership in the registry as constructed, not the existence of two
// string constants — a constant nothing registers is not a code").
//
// It holds ONLY the registry. `folio.Diagnostic`, `folio.Severity` and
// `folio.Result` stay at the module root (D-2.8.3, an owner decision;
// see ARCHITECTURE-SPINE.md:613 as amended by this story's AC11). The
// caveat-kind -> code MAPPING (diagnosticFromCaveat) also stays out of
// this package and in folio-go/render.go — it switches on
// internal/expr.CaveatKind and constructs a folio.Diagnostic, and this
// package may import neither (see below).
//
// This package imports NO first-party package, not even
// internal/geom. That is not an accident of what a registry happens to
// need — it is the REQUIREMENT R1 names: AC4's five FR41 conditions
// arise in internal/template (stage rank 2), internal/expr (rank 3),
// internal/bind (rank 4), internal/fontset (rank 6) and internal/layout
// (rank 7) — five packages spanning the whole pipeline. lint's
// stage-rank guard permits a package to import only a STRICTLY LOWER
// rank, so only a rank-0/1 leaf is importable by every one of those
// five. This package is ranked 1 (folio-go/internal/, stagerank.go)
// precisely so every stage can attach a registry code to a Diagnostic
// it constructs — the first first-party import added here would
// foreclose whichever of those five packages sits at or above this
// one's rank. internal/arch's TestDiagPackageHasZeroFirstPartyImports
// asserts this property directly, rather than leaving it to be merely
// observed (AC1).
package diag

// Code is a stable, closed-registry string naming one failure mode a
// caller can programmatically dispatch on (AD-14: "a stable string
// code from a closed registry"). A defined type, never a bare string
// (R4, D-000.68's compiler anchor): a naked string literal cannot be
// registered by accident, and a mistyped code does not silently become
// a new one. folio.Diagnostic.Code stays `string` (a public API
// surface this story does not widen); DiagCodeTextClippedWidth and
// DiagCodeEmptyAverage bridge as `= string(diag.CodeX)`, keeping the
// two spellings in permanent lockstep.
type Code string

// The registered codes.
//
// R7 (D-3.6.6), ratified as the criterion for what earns a code here,
// stated in this doc comment because a principle beside the thing it
// governs is read by whoever touches this file next, not only by
// whoever reads the story that minted it:
//
//	A code exists for a failure mode a caller can act on — FR41's
//	five, plus conditions a template author can cause and fix. An
//	internal-invariant violation ("this should never happen") stays
//	a plain error.
//
// Codes are ADDITIVE ONLY (AD-14, verbatim): once shipped, a code's
// string and its meaning are permanent. Never repurpose one; never
// change its string. AC5's registry test pins every shipped code's
// exact string against a literal the test owns, independent of this
// file, so a change here that alters a byte is caught as the breaking
// change it is.
const (
	// CodeTextClippedWidth names a text element's widest packed line
	// exceeding its declared width (FR44, D-2.8.1). Story 2.8's
	// original code, absorbed here unchanged (AC3): the public
	// constant is folio.DiagCodeTextClippedWidth.
	CodeTextClippedWidth Code = "TEXT_CLIPPED_WIDTH"

	// CodeEmptyAverage names avg() evaluated over a present-but-empty
	// collection (Story 3.3, DECISION-5). Absorbed here unchanged
	// (AC3): the public constant is folio.DiagCodeEmptyAverage.
	CodeEmptyAverage Code = "AGGREGATE_EMPTY_AVERAGE"

	// CodeTableFooterSourceUnresolved names a table column with
	// footer: "sum"/"avg", footerOf omitted, and a bind that is not one
	// of D-1.4.1's two derivable shapes (DW-6, R8; site:
	// folio-go/folio_expr_validate.go:316-320).
	CodeTableFooterSourceUnresolved Code = "TABLE_FOOTER_SOURCE_UNRESOLVED"

	// CodeTableFooterSourceForbidden names footerOf paired with
	// footer: "count", or a footer field present with no footer at all
	// (DW-6, R8, D-1.4.2's parenthetical; sites:
	// internal/template/parse_bands.go:408-410, :419-421 — one code,
	// two sites, because the code names the CONDITION, not the line).
	CodeTableFooterSourceForbidden Code = "TABLE_FOOTER_SOURCE_FORBIDDEN"

	// CodeTemplateMalformed names FR41's "malformed template" mode: a
	// `.folio` document that fails to load (internal/template's
	// load-time validation, surfaced through folio.LoadTemplate/
	// ParseTemplate). AC4/AC8.
	CodeTemplateMalformed Code = "TEMPLATE_MALFORMED"

	// CodeBindingPathAbsent names FR41's "unresolvable binding" mode:
	// a data or params path an element's binding names is absent from
	// the document supplied at render time (AD-14's own "absent" data
	// case: "an absent path is an Error carrying the path"). AC4/AC8.
	CodeBindingPathAbsent Code = "BINDING_PATH_ABSENT"

	// CodeExpressionInvalid names FR41's "invalid expression" mode: a
	// "{{ }}" expression that does not parse or does not check (syntax,
	// arity, unknown function name, wrong-kind literal argument). AC4/
	// AC8.
	CodeExpressionInvalid Code = "EXPRESSION_INVALID"

	// CodeContentUnlayoutable names FR41's "unlayoutable content" mode:
	// an element (a text line or an image) taller than the content
	// window it must fit inside (internal/layout.OverflowError, FR44/
	// D-2.6.1). AC4/AC8.
	CodeContentUnlayoutable Code = "CONTENT_UNLAYOUTABLE"

	// CodeTextMissingGlyph names FR41's fifth mode, and the one
	// Warning among the five (AD-8, divergence 6, ruled OPEN-1): a rune
	// covered by no face in its element's declared font chain. Per
	// OPEN-1's ruling, the render OMITS the rune — no glyph, no
	// advance — rather than drawing `.notdef` or substituting a
	// visible replacement; this Warning is the SOLE record that the
	// rune was dropped, so its message names the element id, the rune
	// (both as U+XXXX and its literal form), and the exact chain that
	// was searched (D-000.37: a diagnostic must be actionable, not just
	// present). AC4.
	CodeTextMissingGlyph Code = "TEXT_MISSING_GLYPH"

	// CodeInternalUnhandledCaveat names an internal/expr.Caveat whose
	// Kind has no matching arm in diagnosticFromCaveat (render.go) —
	// unreachable given expr.CaveatKind's current single member, but a
	// live, returnable construction site whose output must never carry
	// an empty Code (AD-14; R12/D-3.6.7, AC7). This is NOT a case of
	// R7's criterion being relaxed for an internal condition: the arm
	// already produces a Diagnostic that is already returned to a
	// caller, so the only choice is between a coded one and a codeless
	// one, never between a coded one and a plain error.
	CodeInternalUnhandledCaveat Code = "INTERNAL_UNHANDLED_CAVEAT"

	// CodeDocumentDateInvalid names Story 3.7's reserved params key
	// (D-3.7.2) carrying a value that is present but not a valid RFC
	// 3339 timestamp — a template author/caller-actionable condition
	// (R7's own criterion) caught by both Render and Validate (D-3.7.1)
	// before it ever reaches internal/pdf's date assembly. AC10.
	CodeDocumentDateInvalid Code = "DOCUMENT_DATE_INVALID"

	// CodeStyleColorInvalid names Story 4.1's own new condition: a
	// style colour string reaching render that is not `#RRGGBB`
	// (folio-format.md: "Colours are #RRGGBB") — unvalidated at load
	// (folio_expr_validate.go's own scope fence: "hex colours...
	// remain entirely unvalidated"), and unreachable before this story
	// because no colour was ever CONSUMED before it (D1: "no fill or
	// stroke anywhere in the PDF writer"). Minted here, at the point
	// the condition first ships (R7/D-000.65), rather than in advance.
	CodeStyleColorInvalid Code = "STYLE_COLOR_INVALID"
)

// allCodes is the registry's own enumeration, in the order the codes
// above are declared. It is the ONE place that lists every shipped
// code; buildRegistry constructs the lookup value from it, and nothing
// else in this package or its callers may bypass it to test a bare
// string literal for validity.
var allCodes = []Code{
	CodeTextClippedWidth,
	CodeEmptyAverage,
	CodeTableFooterSourceUnresolved,
	CodeTableFooterSourceForbidden,
	CodeTemplateMalformed,
	CodeBindingPathAbsent,
	CodeExpressionInvalid,
	CodeContentUnlayoutable,
	CodeTextMissingGlyph,
	CodeInternalUnhandledCaveat,
	CodeDocumentDateInvalid,
	CodeStyleColorInvalid,
}

// registry is the CONSTRUCTED value R2 requires (D-1.4.2 `:9118`): a
// map the package builds from allCodes, not the constants themselves.
// Registered queries this value, never the const block directly.
var registry = buildRegistry()

func buildRegistry() map[Code]struct{} {
	m := make(map[Code]struct{}, len(allCodes))
	for _, c := range allCodes {
		m[c] = struct{}{}
	}
	return m
}

// Registered reports whether c is a member of the registry AS
// CONSTRUCTED — the lookup R2 requires, never a check against the
// const block or against allCodes directly.
func Registered(c Code) bool {
	_, ok := registry[c]
	return ok
}

// All returns every registered code, in the registry's own declared
// order — a copy, so a caller cannot mutate the package's set. Tests
// that must enumerate the registry as constructed (AC5, AC7) call this
// rather than re-deriving a list of their own.
func All() []Code {
	out := make([]Code, len(allCodes))
	copy(out, allCodes)
	return out
}
