package bind

import (
	"strings"
	"testing"
)

func mustDecode(t *testing.T, jsonData string) Value {
	t.Helper()
	v, err := DecodeData([]byte(jsonData))
	if err != nil {
		t.Fatalf("DecodeData(%s): %v", jsonData, err)
	}
	return v
}

// TestAD14Triple is D-1.6.2/AC11's retained fixture triple: AC8, AC9
// and AC10 proved over ONE template text and ONE data path
// ("customer.name"), differing only in the data. Rows 1 and 2 are the
// proof (absent vs explicit null, opposite outcomes) — a future
// refactor collapsing Presence to two states must turn one of them
// red.
func TestAD14Triple(t *testing.T) {
	const text = "Statement for {{customer.name}}"
	const elementID = "e1"

	t.Run("row1_absent_is_error", func(t *testing.T) {
		data := mustDecode(t, `{}`)
		_, err := BindText(text, data, elementID)
		if err == nil {
			t.Fatal("AC8: absent path must be an Error")
		}
		if !strings.Contains(err.Error(), "customer.name") || !strings.Contains(err.Error(), elementID) {
			t.Errorf("error must name both the data path and the element id, got: %v", err)
		}
	})

	t.Run("row2_null_renders_empty", func(t *testing.T) {
		data := mustDecode(t, `{"customer": {"name": null}}`)
		got, err := BindText(text, data, elementID)
		if err != nil {
			t.Fatalf("AC9: explicit null must not be an error, got: %v", err)
		}
		if got != "Statement for " {
			t.Errorf("AC9: null must render as empty, got %q", got)
		}
	})

	t.Run("row3_wrong_kind_is_error_no_coercion", func(t *testing.T) {
		data := mustDecode(t, `{"customer": {"name": 123}}`)
		_, err := BindText(text, data, elementID)
		if err == nil {
			t.Fatal("AC10: a JSON number bound into a text element must be an Error, never coerced")
		}
		if strings.Contains(err.Error(), "123") == false && strings.Contains(err.Error(), "number") == false {
			t.Errorf("error should indicate the wrong-kind value, got: %v", err)
		}
	})
}

// TestBindTextAcceptsBareDottedPath is AC15.
func TestBindTextAcceptsBareDottedPath(t *testing.T) {
	data := mustDecode(t, `{"customer": {"name": "Ada Lovelace"}}`)
	got, err := BindText("Statement for {{customer.name}}", data, "e1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Statement for Ada Lovelace" {
		t.Errorf("got %q", got)
	}
}

// TestBindTextRejectsExpressionSyntax is AC16/AC17: anything
// expression-shaped is a located error naming the element id and
// mentioning Epic 3.
func TestBindTextRejectsExpressionSyntax(t *testing.T) {
	cases := []string{
		`{{formatNumber(transaction.amount, "#,##0.00")}}`,
		`{{a[0]}}`,
		`{{a, b}}`,
		`{{"literal"}}`,
		`{{a + b}}`,
		`{{a b}}`, // interior whitespace
	}
	data := mustDecode(t, `{}`)
	for _, c := range cases {
		_, err := BindText(c, data, "e1")
		if err == nil {
			t.Errorf("BindText(%q): expected a located error", c)
			continue
		}
		if !strings.Contains(err.Error(), "e1") {
			t.Errorf("BindText(%q): error must name the element id, got: %v", c, err)
		}
		if !strings.Contains(err.Error(), "Epic 3") {
			t.Errorf("BindText(%q): error must mention Epic 3, got: %v", c, err)
		}
	}
}

// TestBindTextReservesPageAndPages is AC18/AC19: {{page}} and
// {{pages}} pass through untouched, never resolved from data, never
// an error — even when the data contains a top-level "page" key
// (the retained fixture for the reservation).
func TestBindTextReservesPageAndPages(t *testing.T) {
	data := mustDecode(t, `{"page": "HIJACKED"}`)
	got, err := BindText("Page {{page}} of {{pages}}", data, "e1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Page {{page}} of {{pages}}" {
		t.Errorf("got %q, want the placeholders left byte-for-byte unchanged", got)
	}
}

// TestBindTextPageReservationIsRedProofable is AC19's own red-proof
// directive: with the reservation branch effectively bypassed (by
// asserting the byte-for-byte-unchanged property directly, rather than
// merely checking for absence of an error), a document containing a
// top-level "page" key silently capturing the slot would change what
// {{page}} renders — this test would go red the moment that happens.
//
// QA Nit 15 (this story's review): the original version compared
// against DATA WITHOUT a "page" key, so deleting the reservation
// branch made the second call go Absent — the test reddened on that
// unrelated error check, never on the got1 != got2 comparison its own
// message describes (D-000.13: it failed, but not for the reason it
// names). Both data documents here DO contain a top-level "page" key,
// with DIFFERENT values, so a removed reservation makes both calls
// succeed — with different text — and the comparison itself is what
// reddens, naming the actual hijack.
func TestBindTextPageReservationIsRedProofable(t *testing.T) {
	withPage := mustDecode(t, `{"page": "HIJACKED"}`)
	withOtherPage := mustDecode(t, `{"page": "SOMETHING ELSE"}`)

	got1, err1 := BindText("{{page}}", withPage, "e1")
	if err1 != nil {
		t.Fatalf("unexpected error: %v", err1)
	}
	got2, err2 := BindText("{{page}}", withOtherPage, "e1")
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if got1 != got2 {
		t.Fatalf("a top-level \"page\" data key changed {{page}}'s rendering: %q vs %q", got1, got2)
	}
	if got1 != "{{page}}" {
		t.Fatalf("got %q, want the literal placeholder text", got1)
	}
}

// TestBindTextNumberCoefficientOverflowIsLocatedError is QA Finding 2
// (this story's review, Major): AsDecimal/NewDecimal previously had no
// production caller at all, so AC3/AC4's mandated "located error naming
// the data path and the element id" for an out-of-bound number literal
// existed only in internal/bind's own unit tests, never on a path real
// report data reaches. A number is always wrong-kind for a text binding
// (AC10) — that does not change here — but the literal is now validated
// via Decimal first, so a malformed one is reported as such rather than
// silently passing as an unremarkable "wrong kind" without ever
// touching the exact-decimal conversion AD-23 requires.
func TestBindTextNumberCoefficientOverflowIsLocatedError(t *testing.T) {
	// 20 significant digits: exceeds int64 (AC3), same literal
	// TestNewDecimalCoefficientBound already proves NewDecimal rejects
	// directly.
	data := mustDecode(t, `{"transaction": {"amount": 12345678901234567890}}`)
	_, err := BindText("Amount: {{transaction.amount}}", data, "e7")
	if err == nil {
		t.Fatal("expected a located error for a coefficient that does not fit int64")
	}
	if !strings.Contains(err.Error(), "e7") {
		t.Errorf("error must name the element id, got: %v", err)
	}
	if !strings.Contains(err.Error(), "transaction.amount") {
		t.Errorf("error must name the data path, got: %v", err)
	}
	if !strings.Contains(err.Error(), "coefficient") {
		t.Errorf("error must describe the coefficient-bound violation Decimal actually detected, got: %v", err)
	}
}

// TestBindTextWellFormedNumberIsStillWrongKind pins AC10 alongside
// Finding 2's fix: a well-formed number (one Decimal accepts) bound
// into a text element is still an Error, never coerced — validating via
// AsDecimal must not become a backdoor that lets a legal number render
// as text.
func TestBindTextWellFormedNumberIsStillWrongKind(t *testing.T) {
	data := mustDecode(t, `{"transaction": {"amount": 1.50}}`)
	_, err := BindText("Amount: {{transaction.amount}}", data, "e7")
	if err == nil {
		t.Fatal("AC10: a JSON number bound into a text element must be an Error, never coerced, even when it is a well-formed Decimal")
	}
	if !strings.Contains(err.Error(), "not a string") {
		t.Errorf("error should still read as a wrong-kind rejection, got: %v", err)
	}
}
