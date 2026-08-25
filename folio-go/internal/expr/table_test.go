package expr

import "testing"

// TestLegalFunctionNamesIsExactlyEight is AC5, checked at the package's
// own public surface rather than by reaching into functionTable
// (AC7/table.go's own AST guard, internal/expr_arch_test.go, is what
// asserts the LITERAL is exactly eight — this test additionally pins
// the exported accessor built on top of it).
func TestLegalFunctionNamesIsExactlyEight(t *testing.T) {
	want := map[string]bool{
		"sum": true, "count": true, "avg": true,
		"formatDate": true, "formatNumber": true,
		"upper": true, "lower": true, "if": true,
	}
	got := LegalFunctionNames()
	if len(got) != 8 {
		t.Fatalf("LegalFunctionNames() has %d entries, want 8: %v", len(got), got)
	}
	for _, name := range got {
		if !want[name] {
			t.Errorf("unexpected function name %q", name)
		}
		delete(want, name)
	}
	if len(want) != 0 {
		t.Errorf("missing function names: %v", want)
	}
}

// TestThreeImplementedFunctions and TestFiveUnimplementedFunctions pin
// which of the eight are implemented at this story (AC12-AC18) —
// exactly three, never more, never fewer.
func TestThreeImplementedFunctions(t *testing.T) {
	for _, name := range []string{"upper", "lower", "if"} {
		e, ok := lookupFunc(name)
		if !ok || !e.implemented {
			t.Errorf("%s: expected implemented=true", name)
		}
	}
}

func TestFiveUnimplementedFunctions(t *testing.T) {
	for _, name := range []string{"sum", "count", "avg", "formatDate", "formatNumber"} {
		e, ok := lookupFunc(name)
		if !ok || e.implemented {
			t.Errorf("%s: expected implemented=false", name)
		}
		if e.owningStory == "" {
			t.Errorf("%s: unimplemented entry must name its owning story", name)
		}
	}
}

// TestAggregationEntriesDeclareDecimalReturn is AC9: sum/count/avg's
// table entries declare a Decimal-typed signature (a compile-time
// property — this test merely observes it, since a stringly-typed
// table would fail to compile this assertion at all).
func TestAggregationEntriesDeclareDecimalReturn(t *testing.T) {
	for _, name := range []string{"sum", "count", "avg"} {
		e, _ := lookupFunc(name)
		if _, ok := e.ret.(returnDecimal); !ok {
			t.Errorf("%s: expected a Decimal-typed return declaration, got %T", name, e.ret)
		}
	}
}
