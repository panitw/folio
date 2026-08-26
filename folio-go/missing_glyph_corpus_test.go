package folio

import "testing"

// TestCorpusFixturesProduceNoMissingGlyphWarnings is the obligation the
// engineering lead attached to OPEN-1's ruling (Story 3.6): rendering
// any COMMITTED fixture must produce ZERO missing-glyph Warnings. This
// is what stops the corpus quietly normalising a dropped character —
// D-000.22's own warning, "a wrong first recording is not a bug that
// gets caught later, it is a bug that gets RATIFIED later" — applied to
// a Warning whose only in-band effect is a silently shorter string.
//
// It renders every fixture this module commits to its acceptance
// corpus (the same five documents first_baseline_acceptance_test.go's
// baselineAcceptanceFixtures walks) and asserts the COUNT of
// DiagCodeTextMissingGlyph diagnostics is zero — DERIVED by counting
// res.Diagnostics, never a literal len(res.Diagnostics) == 0, because
// at least one of these fixtures (wrapped-text, e4) legitimately
// carries an UNRELATED Warning (DiagCodeTextClippedWidth) that this
// test must not treat as a violation.
//
// This is deliberately over the COMMITTED corpus only. The genuinely
// uncoverable-rune subject this story needs for AC4's own coverage
// (TestMissingGlyphDiagnosticFiresOnUncoveredRune, ac4_coverage_test.go)
// is a SYNTHETIC, test-only template built inline in that test, never
// added here or to any retained fixture directory — the two obligations
// would look like they conflict without that said explicitly, so it is
// said here as well as in that test's own doc comment.
func TestCorpusFixturesProduceNoMissingGlyphWarnings(t *testing.T) {
	type corpusFixture struct {
		name string
		tpl  func(t *testing.T) *Template
		data Data
		fs   func(t *testing.T) FontSet
	}

	fixtures := []corpusFixture{
		{
			name: "font-text",
			tpl:  func(t *testing.T) *Template { return parseFontTestTemplate(t) },
			data: Data("{}"),
			fs:   func(t *testing.T) FontSet { return testFontSet() },
		},
		{
			name: "multi-script-fallback",
			tpl: func(t *testing.T) *Template {
				tpl, err := ParseTemplate([]byte(multiScriptTestTemplateJSON))
				if err != nil {
					t.Fatalf("parse multi-script template: %v", err)
				}
				return tpl
			},
			data: Data("{}"),
			fs:   func(t *testing.T) FontSet { return testShippedFontSet() },
		},
		{
			name: "shaped-text",
			tpl: func(t *testing.T) *Template {
				tpl, err := ParseTemplate([]byte(shapedTextTemplateJSON))
				if err != nil {
					t.Fatalf("parse shaped-text template: %v", err)
				}
				return tpl
			},
			data: Data("{}"),
			fs:   func(t *testing.T) FontSet { return testShippedFontSet() },
		},
		{
			name: "three-band-page",
			tpl:  func(t *testing.T) *Template { return parseThreeBandPageTemplate(t) },
			data: Data("{}"),
			fs:   func(t *testing.T) FontSet { return testShippedFontSet() },
		},
		{
			name: "wrapped-text",
			tpl: func(t *testing.T) *Template {
				tpl, err := ParseTemplate([]byte(wrappedTextTemplateJSON))
				if err != nil {
					t.Fatalf("parse wrapped-text template: %v", err)
				}
				return tpl
			},
			data: Data(wrappedTextDataJSON),
			fs:   func(t *testing.T) FontSet { return testShippedFontSet() },
		},
	}

	// Finding 11 (QA review, Minor): a length check alone cannot see a
	// fixture SWAP (count unchanged, membership changed) — this test's
	// own doc comment promises "the two must name the same committed
	// corpus," which a count does not verify. Compared here as sets of
	// NAMES, both directions, so a swap in either table reddens.
	wantNames := make(map[string]bool, len(baselineAcceptanceFixtures))
	for _, bf := range baselineAcceptanceFixtures {
		wantNames[bf.name] = true
	}
	gotNames := make(map[string]bool, len(fixtures))
	for _, f := range fixtures {
		gotNames[f.name] = true
	}
	for name := range wantNames {
		if !gotNames[name] {
			t.Errorf("presence precondition: baselineAcceptanceFixtures names %q, which this test's corpus table does not — the two must name the same committed corpus", name)
		}
	}
	for name := range gotNames {
		if !wantNames[name] {
			t.Errorf("presence precondition: this test's corpus table names %q, which baselineAcceptanceFixtures does not — the two must name the same committed corpus", name)
		}
	}
	if len(fixtures) != len(baselineAcceptanceFixtures) {
		t.Fatalf("presence precondition: this test's corpus table has %d entries, baselineAcceptanceFixtures has %d — the two must name the same committed corpus", len(fixtures), len(baselineAcceptanceFixtures))
	}

	total := 0
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			res, err := Render(f.tpl(t), f.data, nil, f.fs(t))
			if err != nil {
				t.Fatalf("Render(%s): %v", f.name, err)
			}
			if len(res.Bytes) == 0 {
				t.Fatalf("Render(%s) produced no bytes", f.name)
			}
			count := 0
			for _, d := range res.Diagnostics {
				if d.Code == DiagCodeTextMissingGlyph {
					count++
				}
			}
			if count != 0 {
				t.Errorf("%s: rendering produced %d missing-glyph Warning(s), want 0 — a committed fixture must never exercise this path: %+v", f.name, count, res.Diagnostics)
			}
			total += count
		})
	}
	if total != 0 {
		t.Errorf("corpus total missing-glyph Warnings across %d fixture(s): %d, want 0", len(fixtures), total)
	}
}
