package folio

import (
	"strings"
	"testing"
)

// This file pins `projectFontChainEntry` (page_setup.go) — the EMBEDDED
// branch, which had no Go test at all when Story 8.3 first landed.
//
// WHY THAT WAS DANGEROUS, MEASURED. Deleting `out.Family = entry.AssetKey`
// left the whole Go suite green. With it gone, a font asset carrying no
// `font.family` projects `family: ""`, and the designer's own guard rejects
// exactly that (`isFontChainEntry`: `if (assetKey.length > 0) return
// family.length > 0`) — so `isCanvas` returns false, `parseInbound` returns
// undefined, `engine-client` calls `worker.terminate()`, and the canvas is
// permanently blank with no element id and nothing to attribute it to. That is
// precisely the failure mode D-8.2.8 and DW-82 exist to prevent, shipping
// green.
//
// The TypeScript side of this contract is tested in engine-protocol.test.ts
// and the key sets are tied together by canvas_projection_wire_test.go. What
// neither of those can see is whether GO EVER PRODUCES the shape the guard
// requires — a hand-written TypeScript fixture proves the guard accepts a
// well-formed entry, not that the engine emits one. This file is that half.

// canvasChainDoc is a one-text-element document carrying one font asset whose
// `font` record is `record` verbatim (or absent when empty), and whose `body`
// chain names the shipped face and then the asset.
func canvasChainDoc(t *testing.T, record string) string {
	t.Helper()
	source := embeddedFontTemplateJSON()
	// Replace the generated document's own `font` record, which is bracketed
	// in canonical (sorted) key order by `data` before it and `mediaType`
	// after it. Sorted order is a property AD-9 guarantees, so this is a
	// bracket rather than a brace count over ~47 KB of base64.
	const open, next = "      \"font\": {", "      \"mediaType\":"
	start := strings.Index(source, open)
	end := strings.Index(source, next)
	if start < 0 || end < 0 || end < start {
		t.Fatalf("fixture assumption violated: the generated document's font record is not bracketed by %q and %q", open, next)
	}
	replacement := ""
	if record != "" {
		replacement = "      \"font\": " + record + ",\n"
	}
	return source[:start] + replacement + source[end:]
}

// projectedChainEntries runs a document through the REAL projection entry
// point — the one that reaches the browser — and returns the `body` chain's
// entries.
func projectedChainEntries(t *testing.T, source string) []CanvasFontChainEntry {
	t.Helper()
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	projection, err := Canvas(tpl)
	if err != nil {
		t.Fatalf("Canvas: %v", err)
	}
	for _, chain := range projection.FontChains {
		if chain.Name == "body" {
			return chain.Entries
		}
	}
	t.Fatalf("the projection carries no chain named body: %+v", projection.FontChains)
	return nil
}

// TestProjectedEmbeddedEntryCarriesTheDiscriminantAndTheRecord asserts the
// whole projected entry, not one field of it. Asserting the struct is what
// keeps a family or style LEAKING onto a named face visible: the browser
// rejects that too (`a named face carries no family and no style`), and it
// blanks the canvas the same way an absent family does.
func TestProjectedEmbeddedEntryCarriesTheDiscriminantAndTheRecord(t *testing.T) {
	key := embeddedFontAssetKey()
	for _, tc := range []struct {
		name   string
		record string
		want   CanvasFontChainEntry
	}{
		{
			name:   "family and style from the record",
			record: `{"family": "Noto Sans Thai", "style": "Regular"}`,
			want:   CanvasFontChainEntry{AssetKey: key, Family: "Noto Sans Thai", Style: "Regular"},
		},
		{
			// The style is genuinely optional, and an absent one projects
			// empty — which the browser accepts, unlike an empty family.
			name:   "family only",
			record: `{"family": "Noto Sans Thai"}`,
			want:   CanvasFontChainEntry{AssetKey: key, Family: "Noto Sans Thai"},
		},
		{
			// THE FALLBACK, and the whole reason this file exists. No
			// `font.family`, so the ASSET KEY is projected as the family. The
			// engine decides what the panel shows; the browser is never handed
			// an empty name it would have to invent a rule for.
			name:   "no family — the asset key is the fallback",
			record: `{"style": "Regular"}`,
			want:   CanvasFontChainEntry{AssetKey: key, Family: key, Style: "Regular"},
		},
		{
			name:   "no font record at all",
			record: "",
			want:   CanvasFontChainEntry{AssetKey: key, Family: key},
		},
		{
			name:   "an empty record",
			record: `{}`,
			want:   CanvasFontChainEntry{AssetKey: key, Family: key},
		},
		{
			// An explicit null is a legal, round-trippable spelling in the
			// FILE, and it means nothing to DISPLAY — so it falls back the
			// same way absence does. The distinction survives in the document
			// (Presence round-trips it); it simply does not reach the panel.
			name:   "an explicitly null family",
			record: `{"family": null, "style": "Regular"}`,
			want:   CanvasFontChainEntry{AssetKey: key, Family: key, Style: "Regular"},
		},
		{
			// An empty-STRING family is a name the panel cannot draw, so it
			// falls back too. Without this the browser's `family.length > 0`
			// rejects the snapshot.
			name:   "an empty-string family",
			record: `{"family": "", "style": "Regular"}`,
			want:   CanvasFontChainEntry{AssetKey: key, Family: key, Style: "Regular"},
		},
		{
			name:   "an explicitly null style",
			record: `{"family": "Noto Sans Thai", "style": null}`,
			want:   CanvasFontChainEntry{AssetKey: key, Family: "Noto Sans Thai"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entries := projectedChainEntries(t, canvasChainDoc(t, tc.record))
			if len(entries) != 2 {
				t.Fatalf("projected %d entries, want 2: %+v", len(entries), entries)
			}
			// The NAMED face first — asserted whole, so a display string
			// leaking onto it is caught here rather than in the browser.
			if want := (CanvasFontChainEntry{Face: "Noto Sans"}); entries[0] != want {
				t.Errorf("named-face entry = %+v, want %+v — a named face's name IS its identity and it carries no family or style", entries[0], want)
			}
			if entries[1] != tc.want {
				t.Errorf("embedded entry = %+v, want %+v", entries[1], tc.want)
			}
			// The invariant the designer's guard turns on, stated directly so
			// its violation is named rather than inferred from a struct diff.
			if entries[1].Family == "" {
				t.Error("an embedded entry projected an EMPTY family — the designer's guard rejects that, parseInbound returns undefined, the worker is terminated and the canvas goes permanently blank (D-8.2.8, DW-82)")
			}
			if (entries[1].Face == "") == (entries[1].AssetKey == "") {
				t.Errorf("exactly one of Face and AssetKey must be non-empty, got %+v", entries[1])
			}
		})
	}
}

// TestProjectedEntryStringsAreBounded closes the other half of the bound. Only
// the FACE arm was exercised before (canvas_body_text_bounds_test.go's "fonts
// chain entry" row); Family and Style are equally on the wire, the browser
// bounds all four, and a bound applied to one field of four is a bound on
// nothing.
//
// The AssetKey arm is deliberately absent and that is not an oversight: a key
// is 64 hex characters by the format's own load rule (isSHA256HexKey), so no
// loadable document can carry one over the bound — a test would have to build
// a Document the loader refuses, which asserts something about a state the
// engine cannot be in.
func TestProjectedEntryStringsAreBounded(t *testing.T) {
	long := strings.Repeat("f", maxCanvasPropertyString+1)
	atLimit := strings.Repeat("f", maxCanvasPropertyString)

	for _, tc := range []struct {
		name    string
		record  string
		refused bool
	}{
		{"family at the limit", `{"family": "` + atLimit + `"}`, false},
		{"family over the limit", `{"family": "` + long + `"}`, true},
		{"style at the limit", `{"family": "Noto Sans Thai", "style": "` + atLimit + `"}`, false},
		{"style over the limit", `{"family": "Noto Sans Thai", "style": "` + long + `"}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := ParseTemplate([]byte(canvasChainDoc(t, tc.record)))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			_, cerr := Canvas(tpl)
			if tc.refused {
				if cerr == nil {
					t.Fatal("a projected string over the bound must be REFUSED with a stated reason, never silently cut")
				}
				if !strings.Contains(cerr.Error(), "font chain entry exceeds the projection bound") {
					t.Errorf("Canvas refused with %q, want the projection-bound reason", cerr)
				}
				return
			}
			if cerr != nil {
				t.Fatalf("a projected string AT the bound must be accepted, got: %v", cerr)
			}
		})
	}
}
