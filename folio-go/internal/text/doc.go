// Package text implements Story 2.1's spike deliverable: a Thai
// break-opportunity engine built against AD-25 ("Thai break
// opportunities fail toward not breaking") and NFR3 / CAP-14 (multi-
// script text support).
//
// This package does not build the line breaker (Story 2.4) and does
// not shape text (Story 2.3, GPOS). It answers one question
// mechanically: "at which rune positions may a line legally break?" —
// for Thai, that requires a dictionary, because Thai is written without
// interword spaces.
//
// AD-25's two absolutes, both implemented here and nowhere overridden:
//
//   - No break inside a Thai character cluster — determined mechanically
//     from Unicode combining classes and the Thai block's own structure
//     (see cluster.go). This needs no dictionary and does not depend on
//     shaping or a font (Trap 4).
//   - A run of Thai text the dictionary cannot cover yields no interior
//     break opportunities — it is treated as atomic (see break.go).
//
// The dictionary itself is PyThaiNLP's words_th wordlist (CC0-1.0,
// wordlist/), compiled into a compact BytesTrie (trie.go) and embedded
// into the binary via go:embed (data.go) — never read from disk at
// runtime, so folio-go behaves identically under js/wasm and on a
// server (AC1, AC2, D-000.5, Trap 3).
package text
