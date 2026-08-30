package folio

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/diag"
	"github.com/panitw/folio/folio-go/internal/expr"
)

// TestDiagnosticRegistryErrorCensus derives its worklist from the constructed
// registry. A new registered Error therefore cannot pass merely because a
// hand-maintained subset forgot it: it needs a disposition and a real
// production trigger before the suite is green.
func TestDiagnosticRegistryErrorCensus(t *testing.T) {
	for _, code := range diag.All() {
		if disposition, ok := diag.Classified(code); !ok || (disposition != diag.DispositionError && disposition != diag.DispositionWarning) {
			t.Fatalf("registered code %q has no valid registry disposition", code)
		}
	}

	triggers := map[diag.Code]func(*testing.T) error{
		diag.CodeTemplateMalformed: func(t *testing.T) error {
			_, err := ParseTemplate([]byte(malformedTemplateJSON))
			return err
		},
		diag.CodeExpressionInvalid: func(t *testing.T) error {
			_, err := ParseTemplate([]byte(invalidExpressionTemplateJSON))
			return err
		},
		diag.CodeBindingPathAbsent: func(t *testing.T) error {
			tpl, err := ParseTemplate([]byte(unresolvableBindingTemplateJSON))
			if err != nil {
				return err
			}
			_, err = Render(tpl, Data(`{}`), Params(`{}`), testShippedFontSet())
			return err
		},
		diag.CodeContentUnlayoutable: func(t *testing.T) error {
			tpl, err := ParseTemplate([]byte(overflowLineTemplate))
			if err != nil {
				return err
			}
			_, err = Render(tpl, Data(`{}`), Params(`{}`), testShippedFontSet())
			return err
		},
		diag.CodeDocumentDateInvalid: func(t *testing.T) error {
			tpl, err := ParseTemplate([]byte(unresolvableBindingTemplateJSON))
			if err != nil {
				return err
			}
			_, err = Render(tpl, Data(`{"customer":{"name":"Ada"}}`), Params(`{"documentDate":"not-rfc3339"}`), testShippedFontSet())
			return err
		},
		diag.CodeStyleColorInvalid: func(t *testing.T) error {
			source := strings.Replace(roundTripGoldenSource(t), `"border": {`, "\"background\": \"not-a-colour\",\n            \"border\": {", 1)
			tpl, err := ParseTemplate([]byte(source))
			if err != nil {
				return err
			}
			_, err = Render(tpl, Data(`{"customer":{"name":"Ada"},"transactions":[{"date":"2026-08-29","amount":1}]}`), Params(`{}`), testShippedFontSet())
			return err
		},
		diag.CodeStyleLineSpacingInvalid: func(t *testing.T) error {
			// Story 7.2. A LOAD-time trigger, unlike STYLE_COLOR_INVALID's
			// render-time one: lineSpacing's domain is checked at the
			// trust boundary by the one function both the file path and
			// the property-command path call, so ParseTemplate is where
			// the production condition actually occurs. 1000.001 is one
			// thousandth past the stated sanity ceiling.
			source := strings.Replace(roundTripGoldenSource(t), `"fontSize": 12`, `"fontSize": 12,
            "lineSpacing": 1000.001`, 1)
			_, err := ParseTemplate([]byte(source))
			return err
		},
		diag.CodeTableFooterSourceUnresolved: func(t *testing.T) error {
			source := roundTripGoldenSource(t)
			source = strings.Replace(source, `{{formatNumber(transaction.amount, \"#,##0.00\")}}`, `{{transaction}}`, 1)
			_, err := ParseTemplate([]byte(source))
			return err
		},
		diag.CodeTableFooterSourceForbidden: func(t *testing.T) error {
			source := roundTripGoldenSource(t)
			source = strings.Replace(source, `"footer": "sum",`, "\"footer\": \"count\",\n              \"footerOf\": \"transactions.amount\",", 1)
			_, err := ParseTemplate([]byte(source))
			return err
		},
	}

	expected := diag.ErrorCodes()
	if len(expected) == 0 {
		t.Fatal("registry declares no Error codes; census would be vacuous")
	}
	if len(triggers) != len(expected) {
		t.Fatalf("error trigger map has %d entries for %d registry Error codes", len(triggers), len(expected))
	}
	for _, code := range expected {
		trigger, ok := triggers[code]
		if !ok {
			t.Fatalf("registered Error code %q has no production trigger", code)
		}
		t.Run(string(code), func(t *testing.T) {
			err := trigger(t)
			if err == nil {
				t.Fatalf("production trigger for %q succeeded", code)
			}
			var renderErr *RenderError
			if !errors.As(err, &renderErr) {
				t.Fatalf("%q error is not transported as RenderError: %T %v", code, err, err)
			}
			d := renderErr.Diagnostic
			if d.Severity != SeverityError {
				t.Errorf("%q severity = %s, want Error", code, d.Severity)
			}
			if d.Code != string(code) {
				t.Errorf("code = %q, want registry code %q; production error = %v", d.Code, code, err)
			}
			if strings.TrimSpace(d.Message) == "" {
				t.Errorf("%q has blank actionable message", code)
			}
			if code != diag.CodeTemplateMalformed && code != diag.CodeDocumentDateInvalid && d.ElementID == "" && d.DataPath == "" {
				t.Errorf("%q lacks an applicable element or data-path location: %+v", code, d)
			}
		})
	}

	// Warnings must be exercised just as dynamically as errors.  The only
	// exception to a public Render result is INTERNAL_UNHANDLED_CAVEAT: its
	// production path is deliberately unreachable while expr.CaveatKind is a
	// closed set, so its real return-site is driven with a future-kind value.
	warnings := map[diag.Code]func(*testing.T) Result{
		diag.CodeTextClippedWidth: func(t *testing.T) Result { return renderClipTemplate(t, clipNarrowTemplate) },
		diag.CodeEmptyAverage: func(t *testing.T) Result {
			tpl, err := ParseTemplate([]byte(emptyAverageTemplateJSON("")))
			if err != nil {
				t.Fatal(err)
			}
			result, err := Render(tpl, Data(`{"t":[]}`), nil, testShippedFontSet())
			if err != nil {
				t.Fatal(err)
			}
			return result
		},
		diag.CodeTextMissingGlyph: func(t *testing.T) Result {
			tpl, err := ParseTemplate([]byte(missingGlyphTemplateJSON))
			if err != nil {
				t.Fatal(err)
			}
			result, err := Render(tpl, Data(`{"name":"ก"}`), Params(`{}`), testShippedFontSet())
			if err != nil {
				t.Fatal(err)
			}
			return result
		},
		diag.CodeInternalUnhandledCaveat: func(t *testing.T) Result {
			return Result{Bytes: []byte("mapped"), Diagnostics: []Diagnostic{diagnosticFromCaveat("e1", expr.Caveat{Kind: expr.CaveatKind(255), Path: "future.path"})}}
		},
		diag.CodeTableHeaderRepeatSuppressed: func(t *testing.T) Result {
			tpl, err := ParseTemplate([]byte(tallRowRepeatDoc()))
			if err != nil {
				t.Fatal(err)
			}
			result, err := Render(tpl, Data(multiRowTableData(3, -1)), nil, testShippedFontSet())
			if err != nil {
				t.Fatal(err)
			}
			return result
		},
		diag.CodeTableFooterOrphanSuppressed: func(t *testing.T) Result {
			return renderFooterFixture(t, footerFixtureDocUnsatisfiableTie(), footerFixtureDataUnsatisfiableTie())
		},
		diag.CodeTableRowClippedHeight: func(t *testing.T) Result {
			tpl, err := ParseTemplate([]byte(overTallRowDoc()))
			if err != nil {
				t.Fatal(err)
			}
			result, err := Render(tpl, Data(overTallRowFixtureData()), nil, testShippedFontSet())
			if err != nil {
				t.Fatal(err)
			}
			return result
		},
	}
	warningCount := 0
	for _, code := range diag.All() {
		disposition, _ := diag.Classified(code)
		if disposition != diag.DispositionWarning {
			continue
		}
		warningCount++
		trigger, ok := warnings[code]
		if !ok {
			t.Fatalf("registered Warning code %q has no actionable witness", code)
		}
		t.Run(string(code), func(t *testing.T) {
			result := trigger(t)
			if len(result.Bytes) == 0 {
				t.Fatalf("Warning %q discarded successful render bytes", code)
			}
			for _, d := range result.Diagnostics {
				if d.Code != string(code) {
					continue
				}
				if d.Severity != SeverityWarning {
					t.Fatalf("Warning %q severity = %s", code, d.Severity)
				}
				if d.ElementID == "" && d.DataPath == "" {
					t.Fatalf("Warning %q lacks a location: %+v", code, d)
				}
				if strings.TrimSpace(d.Message) == "" {
					t.Fatalf("Warning %q has no actionable message", code)
				}
				return
			}
			t.Fatalf("successful production witness did not return Warning %q: %+v", code, result.Diagnostics)
		})
	}
	if len(warnings) != warningCount {
		t.Fatalf("warning trigger map has %d entries for %d registry Warning codes", len(warnings), warningCount)
	}
}

func roundTripGoldenSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/template/golden/worked-example.json")
	if err != nil {
		t.Fatalf("read golden source: %v", err)
	}
	return string(b)
}
