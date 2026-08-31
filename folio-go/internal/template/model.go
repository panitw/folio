// Package template owns the `.folio` document model, its parser and its
// serializer (AD-9: "internal/template owns both the parser and the
// serializer"). It never imports "os" (D-1.4.6: LoadTemplate's path
// argument is handled entirely in package folio at the module root) and
// never imports internal/pdf (AC25: the two packages' numeric-spelling
// functions are deliberately unshared).
//
// Every exported type here is Story 1.4's document model. Report data
// (AD-23's exact scaled decimals) is NOT modelled here — that type
// belongs to Story 1.6 (D-1.4.3: "1.4 must not build it").
package template

import "github.com/panitw/folio/folio-go/internal/geom"

// Document is the parsed, canonicalised form of a `.folio` file. Every
// slice and map field is initialised to non-nil empty by the parser and
// by any constructor (D-1.4.3 extended: a nil map/slice with omitempty
// removed serializes to "null", not "{}"/"[]" — the fifth trap, AC23).
type Document struct {
	// Version is carried verbatim (D-1.4.13): never lowered, never
	// gratuitously raised. It is a plain string ("MAJOR.MINOR") and so
	// carries no round-trip hazard of its own (D-1.4.3).
	Version string

	// Locale is one of the closed set en/th/zh-Hans/ja (AD-12).
	Locale string

	// UTCOffset matches ±HH:MM.
	UTCOffset string

	Page  Page
	Fonts Fonts

	// Bands has exactly the three keys content/pageFooter/pageHeader.
	Bands Bands

	// Assets is keyed by lowercase hex SHA-256 of the raw bytes.
	Assets map[string]Asset

	// UnbreakableValues is the document's declaration of which bound
	// values must never be split across a line break (Story 2.4;
	// D-2.1.6 OWNER, D-2.4.1). Each entry is a BARE ROOT-RELATIVE
	// DOTTED DATA PATH, spelled exactly as `footerOf` is (D-1.4.1: "a
	// bare root-relative dotted value path… No `{{ }}`, no function
	// call, no `[]`") — one path convention in the format, not two.
	// Row-scoped paths are written root-relative under the same
	// convention.
	//
	// DOCUMENT-LEVEL, NOT ELEMENT-LEVEL, AND THE FORMAT'S OWN EXAMPLE
	// IS WHY (D-2.4.1). folio-format.md defines a text element's
	// `value` as a string "which may contain {{ }} bindings", and both
	// canonical examples MIX literal text with bindings — "Statement
	// for {{customer.name}}". An element-level flag would forbid
	// breaking between "Statement" and "for", breaking wrapping for
	// exactly the shape the specification demonstrates. The property
	// belongs to the DATA, not to a box: if customer.name holds a name
	// in the header it holds one in the footer, so it is declared once.
	//
	// The engine NEVER infers membership of this list. That is the
	// whole point of D-2.1.6: Thai surnames are coined by law out of
	// ordinary dictionary words, so no dictionary-coverage rule can
	// tell a proper noun from its parts. See internal/text's package
	// doc for the mechanism and its disclosed limitation.
	//
	// Optional and ADDITIVE: absent from a document that does not use
	// it, and round-tripped as an absent key. A new optional key is a
	// MINOR addition under D-1.4.12 — it is a list, not an extension of
	// a closed set of values — and D-1.4.9's passthrough already
	// guarantees an older library loads a file carrying it.
	//
	// Authored order is preserved (order carries no meaning, and
	// preserving it keeps load/save a fixed point without reordering an
	// author's file). Duplicates are a load error.
	UnbreakableValues []string

	// NextID is the next element-id counter value, decimal (AC32).
	NextID int64

	// Extra carries unknown top-level keys opaquely (AC8), sorted by
	// byte-order key (AC18).
	Extra []Field
}

// Field is one opaque, unknown key/value pair captured by the
// passthrough store at some nesting level (AC8, AC9).
type Field struct {
	Key   string
	Value RawValue
}

// Page is the document's page setup.
type Page struct {
	// Margin is required in this story's model — folio-format.md's
	// worked example always carries all four edges and no default is
	// documented for an absent margin (unlike style.padding, whose
	// default of 0 IS documented). A future story may relax this.
	Margin Margin

	// Orientation is "portrait" or "landscape".
	Orientation string

	// Size is either a named page size ("A4", "Letter") or a custom
	// {height,width} object. Exactly one of SizeName/SizeCustom is set.
	SizeName   string
	SizeCustom PageSize
	SizeIsName bool

	// Extra carries unknown keys on the page object opaquely (AC8,
	// D-1.4.9 OWNER — this story's finisher review, Finding 2: unlike
	// `bands`, folio-format.md states no closed key set for `page`, so
	// there is no ruling authorising a refusal here).
	Extra []Field
}

// PageSize is a custom page size in points.
type PageSize struct {
	Height geom.Length
	Width  geom.Length
}

// Margin holds the four page-margin edges, in points.
type Margin struct {
	Top, Right, Bottom, Left geom.Length

	// Extra carries unknown keys on page.margin opaquely (AC8, D-1.4.9
	// OWNER; this story's finisher review, Finding 2).
	Extra []Field
}

// Padding holds the four style-padding edges, in points. Each edge is
// individually optional (the worked example's `"padding": {"left": 3,
// "right": 3}` omits top/bottom) — an omitted edge means "use the
// documented default, 0" for layout purposes, but the key itself stays
// absent on serialize (P3: canonical is a fixed point, so an omitted
// edge must round-trip as omitted, not reappear as an authored "0").
type Padding struct {
	Top, Right, Bottom, Left Presence[geom.Length]

	// Extra carries unknown keys on style.padding opaquely (AC8, D-1.4.9
	// OWNER; this story's finisher review, Finding 2).
	Extra []Field
}

// FontChainEntry is ONE entry of a fallback chain, and it has exactly
// two shapes (Story 8.3, FR53/FR56): a face NAME the renderer is handed
// at render time, or a reference to a face carried INSIDE the document
// as an `assets` entry, spelled `{"asset": "<key>"}` in the file.
//
// The two are discriminated by which field is non-empty, and exactly one
// of them ever is — decodeFonts refuses an empty asset key and an empty
// face alike, so `Face != ""` and `AssetKey != ""` partition the type
// rather than merely overlapping it. Embedded() is THE predicate; a
// caller that writes `e.AssetKey != ""` itself is writing the same test
// a second time.
//
// It is a struct rather than an interface or a `any` because it crosses
// the parse/serialize/project boundary three times and every crossing
// wants the discriminant checkable at compile time. A one-key object is
// the file's whole shape here, so the struct has no Extra: an entry
// object carrying any key besides `asset` is a located load error, not
// passthrough — the object IS the discriminant, and an unknown key in it
// would be an entry of an unknown kind, not a known entry with an
// unknown decoration.
type FontChainEntry struct {
	// Face is the name of a face the FontSet supplies. Non-empty exactly
	// when this entry is a plain JSON string in the file.
	Face string
	// AssetKey is the `assets` key of a face the document carries.
	// Non-empty exactly when this entry is a `{"asset": …}` object.
	AssetKey string
}

// Embedded reports whether this entry names a face the document carries
// rather than one the renderer is given. It is the ONE place the
// discriminant is spelled.
func (e FontChainEntry) Embedded() bool { return e.AssetKey != "" }

// FaceEntry and AssetEntry build the two shapes. They exist so a caller
// never writes a bare composite literal whose field choice IS the
// discriminant — `FontChainEntry{Face: name}` and
// `FontChainEntry{AssetKey: name}` differ by one word and mean opposite
// things.
func FaceEntry(face string) FontChainEntry { return FontChainEntry{Face: face} }

// AssetEntry builds the embedded shape; see FaceEntry.
func AssetEntry(key string) FontChainEntry { return FontChainEntry{AssetKey: key} }

// Fonts maps a fallback-chain name to its ordered list of entries.
// The chain's own array order is authored and preserved verbatim; only
// the map's keys are sorted at serialize time (AC18).
//
// The element type was []string until Story 8.3. It is []FontChainEntry
// now because a chain may name a face the document itself carries, and
// a string could only ever name a face the renderer already ships — a
// document that wanted any other typeface was an install instruction,
// not a contract.
type Fonts map[string][]FontChainEntry

// Chain is THE authority for "is this a chain style.fontFamily may name":
// it returns the chain only when the key is PRESENT and the chain is
// NON-EMPTY, because a chain with no entries resolves to no face and so is
// not a family anything may name. A caller that needs the weaker question —
// "is this key declared at all", which a chain-editing command needs so an
// empty chain stays deletable — must index the map itself, deliberately.
//
// It replaces the five open-coded copies of that same two-part test measured
// across the module at b2fdaa1, and exists so a sixth is never written:
//   - folio.knownFontFamily      (component_commands.go) — the fontFamily property command
//   - folio.defaultFontFamily    (component_commands.go) — the chain a new text element adopts
//   - folio.canvasFontChains     (page_setup.go)         — the projected chain list
//   - folio.fontChain            (render.go)             — a text element's chain at render
//   - the table header-style resolver (table_render.go)  — headerStyle.fontFamily at render
//
// Each caller keeps its own message text; only the predicate is shared.
// All five were typed on []string until Story 8.3 and are typed on
// []FontChainEntry now; the list is re-verified rather than inherited,
// and it is still exactly five (measured over the module at f51dd5e:
// `grep -rn "Fonts.Chain" --include='*.go'` outside _test.go names these
// five call sites and no sixth).
func (f Fonts) Chain(name string) ([]FontChainEntry, bool) {
	chain, ok := f[name]
	if !ok || len(chain) == 0 {
		return nil, false
	}
	return chain, true
}

// Bands holds exactly the three band keys (AC5). Unlike Page, Margin,
// Padding, Border and Asset, Bands deliberately carries NO Extra field:
// AC5 and folio-format.md (:101, "Exactly these three keys (FR6)") make
// the band-name set itself one of the closed sets this story enforces —
// D-1.4.9's "nothing is refused" governs unknown KEYS inside an object,
// not a structural rule the format's own field table states as closed.
// This story's finisher review (Finding 2) confirmed the other five
// object levels had no such backing and fixed those; bands' closure is
// not the same defect and stays as shipped.
type Bands struct {
	Content    Band
	PageFooter Band
	PageHeader Band
}

// Band is one of the three template bands. Height is a presence flag:
// content bears none (AC5 — "not on content"), pageHeader/pageFooter
// carry one.
type Band struct {
	Elements []Element
	Height   Presence[geom.Length]
	Extra    []Field
}

// ElementType is the closed set of element kinds (FR4).
type ElementType string

const (
	ElementText  ElementType = "text"
	ElementImage ElementType = "image"
	ElementTable ElementType = "table"
	ElementLine  ElementType = "line"
	ElementRect  ElementType = "rect"
)

// ElementID is the canonical spelling of an element/column id: "e" plus
// a lowercase base-36 counter, e.g. "e1", "ea", "e1z" (AD-10).
type ElementID string

// Element is one of the five element kinds, common fields plus the
// kind-specific extension. A table's Width/Height are never set — a
// table declares only X and Y (AC5, AD-13).
type Element struct {
	ID   ElementID
	Type ElementType

	X, Y          geom.Length
	Width, Height Presence[geom.Length] // absent for a table (AC5)

	VisibleIf Presence[string]
	Style     Presence[Style]

	// KeepTogether is Story 7.7's author-declared keep-together tag
	// (FR51): elements in the CONTENT band sharing one non-empty tag
	// paginate as ONE indivisible unit — the whole set stays in the
	// window it started in, or the whole set moves to the next.
	//
	// It is an ELEMENT-level key rather than a document-level list of
	// id lists (D-7.7 Ruling B): a document-level list would be a
	// second place element ids appear, and something would have to
	// prune it when a component is deleted. A tag is deleted with its
	// own element and can never dangle.
	//
	// Absent (or explicitly null) is "not grouped", and a document
	// declaring no tag renders byte-identically to one written before
	// this key existed. parse_bands.go refuses a tag on a
	// page-header/page-footer element (FR51 scopes the feature to the
	// content band) and on a `table` element (a table's items already
	// carry a row key, and honouring both would be a second grouping
	// model).
	KeepTogether Presence[string]

	// text
	Value Presence[string]

	// image
	Asset Presence[string]

	// table
	Table Presence[TableExt]

	Extra []Field
}

// TableExt is the table-specific extension (AD-13).
type TableExt struct {
	Bind             string
	As               Presence[string]
	Columns          []Column
	HeaderHeight     geom.Length
	AltRowBackground Presence[string]

	// HeaderStyle is an OPTIONAL Style block governing the header row
	// ONLY, never a data row — Story 4.1, the owner's ruling: the
	// author controls how a header looks, reusing the existing Style
	// vocabulary rather than a bespoke header-only schema. A field it
	// leaves absent falls back to the table's own Style (above), then
	// to that field's documented default. It is a deliberate, RULED
	// extension of R5's otherwise-permanent TableExt field set (Story
	// 4.1's Delivery Log records the ruling by name).
	HeaderStyle Presence[Style]
}

// Column is one table column.
type Column struct {
	ID    ElementID
	Label string
	Width geom.Length
	Align Presence[string]
	Bind  string

	Footer       Presence[string]
	FooterOf     Presence[string]
	FooterFormat Presence[string]

	Extra []Field
}

// Style is the optional per-element style block. Every field is
// optional; an absent field means "inherit the documented default"
// (folio-format.md's Style table).
type Style struct {
	Align      Presence[string]
	Background Presence[string]
	// Color is the INK a text-bearing element prints in (Story 10.1).
	// Background is the box behind it; this is the glyphs themselves.
	// Absent means the PDF's own initial fill colour, black, and emits
	// nothing — which is what leaves every document that declares no
	// colour byte-identical.
	Color      Presence[string]
	Bold       Presence[bool]
	Italic     Presence[bool]
	Border     Presence[Border]
	FontFamily Presence[string]
	FontSize   Presence[geom.Length]
	// LineSpacing is Story 7.2's author-set leading ratio, carried as a
	// WHOLE NUMBER OF THOUSANDTHS (the authored 1.5 is 1500, and the
	// absent default is exactly LineSpacingUnit). It is an int64 count
	// and deliberately NOT a geom.Length: it is dimensionless, and
	// spelling it as a length would invite it into the millipoint
	// arithmetic that AD-2 keeps to one unit.
	//
	// It scales the vertical model's Advance and NOTHING else, so a
	// multi-line element's first baseline — hence its top edge — does
	// not move when it changes (D-2.5a/DW-15's two-model split). Absent
	// emits nothing, which is what leaves every document that declares
	// no spacing byte-identical.
	LineSpacing Presence[int64]
	Padding     Presence[Padding]
	Valign      Presence[string]

	Extra []Field
}

// Border is style.border.
type Border struct {
	Color Presence[string]
	Edges Presence[[]string]
	Width Presence[geom.Length]

	// Extra carries unknown keys on style.border opaquely (AC8, D-1.4.9
	// OWNER; this story's finisher review, Finding 2).
	Extra []Field
}

// FontRecord is the optional `font` object a FONT asset may carry
// (Story 8.3): a record ABOUT the face, for the people reading and
// reusing the document, never something the engine derives from the
// bytes and never something resolution consults — a chain entry names a
// face by ASSET KEY, so nothing here can cause a substitution.
//
// Every field is a Presence because absence and an explicit JSON null
// are different things in this format (presence.go), and a refusal
// written only in the non-null branch would let `"family": null` past
// every guard.
type FontRecord struct {
	Family  Presence[string]
	Style   Presence[string]
	Licence Presence[string]
	Source  Presence[string]

	// Extra carries unknown keys on assets[k].font opaquely (AC8,
	// D-1.4.9 OWNER) — the same passthrough every other object level in
	// this model has. NOT the same as FontChainEntry, which deliberately
	// has none: there the object IS the discriminant.
	Extra []Field
}

// Asset is one embedded binary asset.
type Asset struct {
	Data      []string
	MediaType string

	// Font is the optional `font` record a font asset carries (Story
	// 8.3). Presence, not a bare pointer: `"font": null` is a legal,
	// round-trippable spelling that is NOT the same as the key being
	// absent, and an image asset must keep serializing without the key
	// at all — the six shipped fixtures with a non-empty assets map are
	// the population that proves absence costs no bytes.
	Font Presence[FontRecord]

	// Extra carries unknown keys on one assets[entry] object opaquely
	// (AC8, D-1.4.9 OWNER; this story's finisher review, Finding 2).
	Extra []Field
}
