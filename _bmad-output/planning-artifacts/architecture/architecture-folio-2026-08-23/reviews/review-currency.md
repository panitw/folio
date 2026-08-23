# Reviewer lens — technology currency and reality-check

Verdict: **PASS.** Every pinned version was checked against a live source on 2026-08-23; nothing rests on training-data recall.

| Claim in spine | How verified | Result |
|---|---|---|
| Go 1.26.x pinned, 1.27 deliberately not adopted | go.dev release notes / blog | Go 1.27 released 2026-08-19. Deliberate non-adoption is correct: addendum §K measured determinism on 1.26 and `compress/flate` stability is verified only through 1.26. 1.27 also ships a rewritten JSON engine, which touches the `.folio` parse path. |
| `boxesandglue/textshape` v0.0.15 | `proxy.golang.org/@latest` | v0.0.15, tagged 2026-08-19. Current, four days old. |
| `go-text/typesetting` (cross-validation) | `proxy.golang.org/@latest` | v0.3.4, 2026-02-25. Spine said "current" — **fixed to the pinned version**. |
| gopdf cannot be used | `proxy.golang.org/@latest` + source read | v0.38.0 (2026-07-26), MIT. `CreateEmbeddedFontSubsetName` in `strhelper.go` at current master is `strings.Replace(name," ","+")` + `strings.Replace(name,"/","+")` — it emits **no** six-letter subset tag. Addendum §L's finding confirmed against master, not stale. Independently, gopdf owns font embedding via `SubsetFontObj` and exposes no glyph-id draw API, so it cannot consume a textshape subset. |
| `baseline-pdf` is the right layer but pre-1.0 | GitHub README + `go.mod` + proxy | v1.1.20 (2026-08-19). README: "Not yet used in production. Expect API changes." `go.mod` pulls `boxesandglue/gofpdi` + `speedata/pdfdisassembler` (PDF-import surface Folio does not need). Both reasons for AD-6 hold. |
| React 19.2.x / Vite 7.3.x | react.dev/versions, vite.dev | React 19.2.8 (2026-07-21) latest stable; no React 20. Vite 7.3.0; Node floor 20.19+/22.12+. |
| `pdfjs-dist` 6.2.x | npm | 6.2.108. Worker version must match package version exactly. |
| File System Access API | MDN + Chrome docs | Still Chromium-only in 2026; Firefox and Safari ship OPFS only. AD-20's two-tier model is required, not defensive. |
| Font licensing | Fontsource / Google Fonts / Noto docs | Noto Sans, Noto Sans Thai, Noto Sans SC (variable, `wght`) all SIL OFL 1.1. IBM Plex OFL 1.1. |

**Applied:** pinned `go-text/typesetting` to v0.3.4; recorded `baseline-pdf` v1.1.20 in the AD-6 rationale.
