package folio

import (
	"fmt"
	"maps"
	"slices"

	"github.com/panitw/folio/folio-go/internal/template"
)

// This file is Story 8.4's SEAM: the one place the document's own font
// assets are turned into something the render path's face-name world can
// consume, and the one place an embedded entry's bytes are decoded.
//
// THE CHOICE, AND THE ALTERNATIVE THAT WAS REJECTED (Task 1 states both
// so a later reader can weigh the trade rather than rediscover it).
//
// The problem: SIX non-test functions take (chain []string, fs FontSet,
// cache *fontCache) — resolveRuneFace, shapeSegments, digitTableRun,
// chainLineMetrics, chainVerticalModel, lineAdvance — and FOUR more take
// the chain for messages and vertical arithmetic only (formatFontChain,
// missingGlyphMessage, verticalModel, scaleAdvanceByLineSpacing). NONE of
// the ten can reach a *Template, so none can reach Assets.
//
// REJECTED: widening the six to carry the document (or a richer chain
// element type). That spreads "what is an embedded entry" across six
// answer sites, which is precisely what chainFaceNames' own doc comment
// says the single boundary exists to prevent — and it leaves the other
// four taking a different chain type from their siblings for no reason
// they can state.
//
// CHOSEN: materialise a per-render NAME -> BYTES view upstream of the
// []string narrowing. chainFaceNames keeps emitting []string; an embedded
// entry simply emits a RESERVED face name (embeddedFaceName below)
// instead of being dropped, and the fontCache — which every one of the
// six already holds — is the single site that knows such a name resolves
// from the document's assets rather than from the caller's FontSet. All
// ten signatures are untouched; there is exactly one answer site; and the
// decode stays LAZY, at the point of use, which is what keeps "a chain
// entry naming a non-font asset that nothing draws from renders clean"
// true rather than aspirational.
//
// THE COST, STATED: the view must be built at BOTH fontCache construction
// sites (predictDocument, render.go; addCanvasTextPaint, page_setup.go),
// or the canvas silently keeps the pre-8.4 behaviour. Both use
// newDocumentFontCache below, and TestCanvasMeasuresWithTheEmbeddedFace
// is what reddens if a third site ever calls newFontCache() with a
// document in hand.

// embeddedFaceNamePrefix is the reserved head of every face name this
// package mints for a face the DOCUMENT carries.
//
// WHY A DERIVED NAME AT ALL. fontCache, pagemodel.TextRun.Face and
// pdf.EmbeddedFace.Name are all documented as keying on "the caller's
// FontSet key", so an embedded face needs a key in that same namespace or
// it cannot be cached, cannot be named on a run, and cannot be embedded.
//
// WHY IT IS DERIVED FROM THE ASSET KEY AND NEVER FROM font.family
// (D-8.4.1). AD-8: an embedded face and a shipped face that share a
// family name must never substitute for one another — the ASSET KEY
// decides. font.family/font.style are DISPLAY identity. Deriving this
// name from font.family would let a document's "Inter" and the caller's
// "Inter" collide in exactly the registry AD-8 keeps them out of.
//
// THE PRECEDENCE RULE, so a collision cannot be exploited either way: a
// name in this namespace is resolved from the DOCUMENT'S OWN ASSETS,
// always, and the supplied FontSet is not consulted for it even if the
// caller happens to have filed an entry under the same string
// (fontCache.get checks the embedded index FIRST). A caller therefore
// cannot shadow a document's carried face, and a document cannot capture
// a caller's.
//
// The prefix is letters and a colon: internal/pdf's pdfNameEscape keeps
// [A-Za-z0-9_-] and drops the rest, so this reaches a PDF resource
// dictionary as "asset" + the 64 hex characters of the key — distinct per
// asset, and unmistakable in a produced file.
const embeddedFaceNamePrefix = "asset:"

// embeddedFaceName is the ONE construction site for that name. A caller
// that spells the prefix itself is writing the derivation a second time.
func embeddedFaceName(assetKey string) string { return embeddedFaceNamePrefix + assetKey }

// embeddedFaceSource is one carried face's resolution record: the asset
// to decode, and the address to put in the error if it cannot be.
type embeddedFaceSource struct {
	assetKey string
	// asset is the document's own asset record, captured at index time so
	// the decode needs no *Template. present is false when the chain names
	// a key the assets map does not hold — structurally unreachable today
	// (decodeFontChainEntry refuses an absent key at load) and reported
	// rather than panicked on, exactly as the image loop reports the same
	// condition.
	asset   template.Asset
	present bool
	// site is where the render error points the reader: the chain and the
	// entry index, which is what makes "asset K is not a font" actionable
	// when one chain is shared by many elements. ElementID is left EMPTY
	// on purpose — see template.FontChainSite's own doc comment.
	site template.FontChainSite
}

// embeddedFaceIndex maps a minted face name to the asset behind it. It is
// LOOKED UP BY KEY ONLY and never ranged (ScanMapRange, D-2.2.3): its
// only iteration is at construction, over a SORTED chain-name slice.
type embeddedFaceIndex map[string]embeddedFaceSource

// newEmbeddedFaceIndex builds the per-render view from the document's
// own fonts map. It DECODES NOTHING — it costs one map walk and no
// base64, so building it unconditionally at both cache sites is free.
//
// DETERMINISM. One asset key may appear in several chains, and the face
// name is derived from the asset key ALONE (AD-8: the key decides), so
// exactly one record survives per key and which one must not depend on
// map order. Chains are visited in SORTED NAME order and entries in the
// document's authored order; the first occurrence wins and is the address
// the error names.
func newEmbeddedFaceIndex(t *Template) embeddedFaceIndex {
	if t == nil || t.doc == nil {
		return embeddedFaceIndex{}
	}
	index := embeddedFaceIndex{}
	chainNames := slices.Sorted(maps.Keys(t.doc.Fonts)) // sorted: deterministic first-occurrence.
	for _, chainName := range chainNames {
		for i, entry := range t.doc.Fonts[chainName] {
			if !entry.Embedded() {
				continue
			}
			name := embeddedFaceName(entry.AssetKey)
			if _, seen := index[name]; seen {
				continue
			}
			asset, present := t.doc.Assets[entry.AssetKey]
			index[name] = embeddedFaceSource{
				assetKey: entry.AssetKey,
				asset:    asset,
				present:  present,
				site: template.FontChainSite{
					AssetKey:   entry.AssetKey,
					ChainName:  chainName,
					EntryIndex: i,
				},
			}
		}
	}
	return index
}

// decode is the font analogue of predictDocument's image loop
// (render.go): DecodeAssetBytes -> template.DecodeFontForRender, in that
// order, with prose location and NO diagnostic code — D-7.8.1, and the
// image precedent (render.go's `fmt.Errorf("folio: Render: %w", derr)`)
// takes the same shape for the same reason: no consumer branches on it.
//
// THE ERRORS ARE RETURNED BARE, without a "folio: Render:" prefix,
// because every caller on this path already applies its own located
// wrapper — collectBandTextRuns wraps everything out of shapeSegments
// with "folio: Render: element %s:", naming the element that actually
// needed the face. That is why FontChainSite leaves ElementID empty
// here: prefixing again would name the element twice in one sentence.
//
// It is called from fontCache.parseEmbedded and nowhere else, so a face
// is decoded at most once per render even though this function memoizes
// nothing itself.
func (s embeddedFaceSource) decode() ([]byte, error) {
	if !s.present {
		return nil, fmt.Errorf("%s: names an asset that is not present in the document's assets map", s.site)
	}
	raw, derr := template.DecodeAssetBytes(s.asset)
	if derr != nil {
		return nil, fmt.Errorf("asset %q: %w", s.assetKey, derr)
	}
	if derr := template.DecodeFontForRender(s.asset.MediaType, raw, s.site); derr != nil {
		return nil, derr
	}
	return raw, nil
}
