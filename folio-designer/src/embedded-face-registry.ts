import { embeddedFaceFamily } from './embedded-face-family'

// THE ONE SEAM IN THIS BUILD THAT REGISTERS A FONT FACE AT RUNTIME
// (Story 8.4a). Everything else the browser can rasterize with is declared at
// build time by scripts/build-wasm.mjs and reaches the page as a stylesheet.
// A face a DOCUMENT carries cannot be: it exists only in that document's
// `assets` map, so its bytes have to travel over the engine's own `asset`
// operation and be added to the page's font set while the document is open.
// `canvas-font-stack.test.ts` asserts this is the only such site, and
// `canvas-authority-contract.test.ts` carves this file — by name, and only
// while it still spells this function — out of an otherwise flat prohibition
// on `new FontFace` and on touching `document.fonts`.
//
// AD-17 IS UNTOUCHED. Nothing here measures anything, nothing here is awaited
// by layout, and no paint waits for it. The canvas draws every fragment at the
// engine's own x the moment the projection arrives; a face that registers
// later changes which glyphs are drawn there and nothing else. That is also
// why registration failing is not an error state: the fragment keeps the
// stylesheet's declared stack, the canvas keeps painting, and the session is
// untouched.
//
// THE LIFETIME IS THE DOCUMENT'S, WHICH IS THE WHOLE REASON THIS IS NOT
// SHAPED LIKE ImagePaint. An image's bytes become an object URL — a scoped
// handle owned by the one mounted component that made it, and ImagePaint is
// mounted once per component AND once per repeated sheet, so N components
// across M sheets make N x M independent requests and N x M handles. Copying
// that here would be a defect rather than a duplication: `document.fonts` is a
// GLOBAL, NAME-KEYED registry, so the same face would be added many times
// under one family, one unmounting instance would delete a face another is
// still painting with, and an unmount arriving after a remount would delete
// the live one. Registration therefore happens ONCE PER DOCUMENT, above every
// component, and is released when the document it belongs to is replaced.

/** Reads one asset's bytes out of the open document, or resolves undefined. */
export type CarriedFaceBytes = (assetKey: string) => Promise<ArrayBuffer | undefined>

// The two operations this seam performs on the page's font set, named so the
// handle can be read out of `document` without depending on which revision of
// the DOM typings the build happens to ship — `FontFaceSet` gained its
// set-like members late, and a page that predates them still has them.
type PageFontSet = Readonly<{ add: (face: FontFace) => unknown; delete: (face: FontFace) => unknown }>

// registerCarriedFaces registers one face per asset key and returns the
// release for all of them. The caller re-invokes it when the DOCUMENT changes
// and when the set of carried entries changes; the returned function removes
// exactly the faces this call added and nothing else, so a superseded document
// cannot leave a face behind for the next one to paint with.
//
// `onRegistered` is called with the keys that have actually reached the font
// set, growing as the fetches land. The canvas asks for the derived family for
// exactly those keys — never for a key whose bytes did not arrive — so a
// failed fetch degrades to the stylesheet's declared stack instead of asking
// for a family nothing declares.
//
// STORY 16.3 — THE DERIVATION IS A PARAMETER, AND ITS DEFAULT IS THE DOCUMENT'S.
//
// The font browser sets each specimen in its own family, which means a face has
// to reach `document.fonts` for a family NOTHING OWNS YET — not carried by the
// open document, not shipped by the build. That is a third lifetime, and
// `preview-face-family.ts` argues it.
//
// IT COULD NOT BE DONE BY PASSING PREVIEW KEYS TO THE OLD SIGNATURE, and the
// reason is the whole hazard this module was written around: the family was
// derived HERE, from `embeddedFaceFamily`, so a preview registration would have
// landed under the very names the canvas asks for and a modal's scroll position
// would have decided what the page paints. Handing the derivation in is what
// keeps the two populations in disjoint namespaces while leaving ONE seam that
// may touch the page's font set — the property `canvas-font-stack.test.ts`
// asserts by exact list, and which a second registration module would have
// broken.
//
// A DERIVATION MAY DECLINE. `previewFaceFamily` returns `undefined` for a name
// it will not encode, and that key is then simply not registered — the same
// degrade a fetch returning no bytes already takes, for the same reason: a face
// the browser cannot name is not a session fault.
export function registerCarriedFaces(assetKeys: ReadonlyArray<string>, readFaceBytes: CarriedFaceBytes, onRegistered: (registered: ReadonlySet<string>) => void, familyFor: (assetKey: string) => string | undefined = embeddedFaceFamily): () => void {
  let active = true
  const added: FontFace[] = []
  const registered = new Set<string>()
  // A build with no font-set API at all — a test environment, an old browser —
  // registers nothing and says so by never calling back. It is not a failure:
  // the canvas paints with the declared stack, which is what it did before
  // this seam existed.
  const fontSet = typeof document === 'undefined' ? undefined : (document.fonts as unknown as PageFontSet | undefined)
  if (typeof FontFace === 'function' && fontSet !== undefined && typeof fontSet.add === 'function') {
    for (const assetKey of new Set(assetKeys)) {
      void readFaceBytes(assetKey)
        .then((bytes) => {
          // No bytes is the engine declining, not an exception to swallow: the
          // asset may be absent, unreadable, or not a font at all, and every
          // one of those is a fragment that keeps the declared stack.
          if (!active || bytes === undefined || bytes.byteLength === 0) return undefined
          const family = familyFor(assetKey)
          if (family === undefined) return undefined
          const face = new FontFace(family, bytes)
          return face.load().then((loaded) => {
            // The document may have been replaced while the bytes were in
            // flight. Adding here would put a superseded document's face into
            // the registry with nothing left to remove it.
            if (!active) return
            fontSet.add(loaded)
            added.push(loaded)
            registered.add(assetKey)
            onRegistered(new Set(registered))
          })
        })
        .catch(() => {
          // A face that will not parse is a document fact, not a session
          // fault. It must never reach the engine's failure channel: the
          // canvas keeps painting and the worker keeps running.
          return undefined
        })
    }
  }
  return () => {
    active = false
    for (const face of added) {
      try {
        fontSet?.delete(face)
      } catch (error) {
        // A face already gone from the set is the outcome this asked for.
        void error
      }
    }
  }
}
