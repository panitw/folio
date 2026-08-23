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

// noParams is the "no runtime values supplied" params root most
// existing BindText tests use — they exercise the data root only, and
// AC16's absent-params behaviour is pinned separately.
var noParams = Value{Kind: KindObject}

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
		_, err := BindText(text, data, noParams, elementID)
		if err == nil {
			t.Fatal("AC8: absent path must be an Error")
		}
		if !strings.Contains(err.Error(), "customer.name") || !strings.Contains(err.Error(), elementID) {
			t.Errorf("error must name both the data path and the element id, got: %v", err)
		}
	})

	t.Run("row2_null_renders_empty", func(t *testing.T) {
		data := mustDecode(t, `{"customer": {"name": null}}`)
		got, err := BindText(text, data, noParams, elementID)
		if err != nil {
			t.Fatalf("AC9: explicit null must not be an error, got: %v", err)
		}
		if got != "Statement for " {
			t.Errorf("AC9: null must render as empty, got %q", got)
		}
	})

	t.Run("row3_wrong_kind_is_error_no_coercion", func(t *testing.T) {
		data := mustDecode(t, `{"customer": {"name": 123}}`)
		_, err := BindText(text, data, noParams, elementID)
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
	got, err := BindText("Statement for {{customer.name}}", data, noParams, "e1")
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
		_, err := BindText(c, data, noParams, "e1")
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
	got, err := BindText("Page {{page}} of {{pages}}", data, noParams, "e1")
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

	got1, err1 := BindText("{{page}}", withPage, noParams, "e1")
	if err1 != nil {
		t.Fatalf("unexpected error: %v", err1)
	}
	got2, err2 := BindText("{{page}}", withOtherPage, noParams, "e1")
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
	_, err := BindText("Amount: {{transaction.amount}}", data, noParams, "e7")
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
	_, err := BindText("Amount: {{transaction.amount}}", data, noParams, "e7")
	if err == nil {
		t.Fatal("AC10: a JSON number bound into a text element must be an Error, never coerced, even when it is a well-formed Decimal")
	}
	if !strings.Contains(err.Error(), "not a string") {
		t.Errorf("error should still read as a wrong-kind rejection, got: %v", err)
	}
}

// TestBindTextParamsRootTakesPrecedenceOverData is AC13/AC14
// (D-1.7.4), copying TestBindTextPageReservationIsRedProofable's
// corrected shape (QA Nit 15) verbatim: two DATA documents, both
// carrying a top-level "params" key with DIFFERENT values, plus one
// supplied params document. Equality alone (got1 == got2) is the
// weaker half — an implementation that simply MERGES params into the
// data root would still pass equality if both data documents were
// identical, so both checks are required, and the value check is what
// actually names the hijack were the reservation to fail.
func TestBindTextParamsRootTakesPrecedenceOverData(t *testing.T) {
	dataA := mustDecode(t, `{"params": {"reportDate": "SHADOWED-FROM-DATA-A"}}`)
	dataB := mustDecode(t, `{"params": {"reportDate": "SHADOWED-FROM-DATA-B"}}`)
	params := mustDecode(t, `{"reportDate": "2026-08-23"}`)

	got1, err1 := BindText("{{params.reportDate}}", dataA, params, "e1")
	if err1 != nil {
		t.Fatalf("unexpected error: %v", err1)
	}
	got2, err2 := BindText("{{params.reportDate}}", dataB, params, "e1")
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if got1 != got2 {
		t.Fatalf("a top-level \"params\" data key changed {{params.reportDate}}'s rendering: %q vs %q", got1, got2)
	}
	if got1 != "2026-08-23" {
		t.Fatalf("AC13: {{params.reportDate}} must resolve against the SUPPLIED PARAMS document, got %q, want the params value \"2026-08-23\"", got1)
	}

	// The whitespace-tolerant spelling shadows identically (M-6): the
	// fix belongs at path-segment level, not on the trimmed literal.
	got3, err3 := BindText("{{ params.reportDate }}", dataA, params, "e1")
	if err3 != nil {
		t.Fatalf("unexpected error: %v", err3)
	}
	if got3 != "2026-08-23" {
		t.Fatalf("AC13: the spaced spelling must resolve identically, got %q", got3)
	}
}

// TestBindTextParamsMissDoesNotFallBackToData is AC13/AC14's
// complementary case, and RP-3's own target (D-1.7.4): a SUPPLIED
// params document that is missing the queried key must still produce
// AC16's absent-params error, never silently fall back to a same-named
// key sitting in report data. Two data documents, both carrying a
// top-level "params.reportDate" decoy with DIFFERENT values, plus one
// supplied params document that does NOT have "reportDate" — both
// calls must produce the located absent-params error, and neither may
// equal the decoy value (which would prove a fallback happened, naming
// the hijack).
func TestBindTextParamsMissDoesNotFallBackToData(t *testing.T) {
	dataA := mustDecode(t, `{"params": {"reportDate": "SHADOWED-FROM-DATA-A"}}`)
	dataB := mustDecode(t, `{"params": {"reportDate": "SHADOWED-FROM-DATA-B"}}`)
	params := mustDecode(t, `{"otherField": "present-but-no-reportDate"}`)

	got1, err1 := BindText("{{params.reportDate}}", dataA, params, "e1")
	if err1 == nil {
		t.Fatalf("RP-3: a params miss must not fall back to report data, got no error, value %q (report data's decoy)", got1)
	}
	if !strings.Contains(err1.Error(), "params.reportDate") || !strings.Contains(err1.Error(), "the supplied params") {
		t.Errorf("error must name the params path and the params root, got: %v", err1)
	}

	got2, err2 := BindText("{{params.reportDate}}", dataB, params, "e1")
	if err2 == nil {
		t.Fatalf("RP-3: a params miss must not fall back to report data, got no error, value %q (report data's decoy)", got2)
	}
}

// TestBindTextParamsAbsentIsLocatedError is AC16 (D-1.7.4, AD-14):
// "{{params.x}}" with no params supplied is an Error naming the path
// AND the element id — and the message must name the PARAMS root, not
// report data (M-6: "absent from the report data" would be actively
// misleading here, since the value was never sought in report data at
// all).
func TestBindTextParamsAbsentIsLocatedError(t *testing.T) {
	data := mustDecode(t, `{}`)
	_, err := BindText("{{params.reportDate}}", data, noParams, "e9")
	if err == nil {
		t.Fatal("AC16: an absent params path must be a located Error")
	}
	if !strings.Contains(err.Error(), "e9") {
		t.Errorf("error must name the element id, got: %v", err)
	}
	if !strings.Contains(err.Error(), "params.reportDate") {
		t.Errorf("error must name the params path, got: %v", err)
	}
	if !strings.Contains(err.Error(), "params") || strings.Contains(err.Error(), "report data") {
		t.Errorf("error must name the PARAMS root, not report data, got: %v", err)
	}
}

// TestBindTextParamsBareIsLocatedError is AC17: "{{params}}" bare, no
// dot, is a located error — params is a namespace, not a value.
func TestBindTextParamsBareIsLocatedError(t *testing.T) {
	data := mustDecode(t, `{}`)
	_, err := BindText("{{params}}", data, noParams, "e10")
	if err == nil {
		t.Fatal("AC17: a bare {{params}} must be a located Error")
	}
	if !strings.Contains(err.Error(), "e10") {
		t.Errorf("error must name the element id, got: %v", err)
	}
	if !strings.Contains(err.Error(), "namespace") {
		t.Errorf("error should say params is a namespace, not a value, got: %v", err)
	}
}

// TestBindTextTopLevelParamsKeyInDataIsNotAnError is AC15 (D-1.7.4,
// verbatim): "rejecting it would refuse legitimate caller data over a
// name collision the caller may not control." A top-level "params" key
// in report data is legal, ordinary, unreachable JSON — decoding and
// binding an UNRELATED placeholder against that data must not error.
func TestBindTextTopLevelParamsKeyInDataIsNotAnError(t *testing.T) {
	data := mustDecode(t, `{"params": {"reportDate": "SHADOWED"}, "customer": {"name": "Ada"}}`)
	got, err := BindText("Statement for {{customer.name}}", data, noParams, "e11")
	if err != nil {
		t.Fatalf("a top-level \"params\" key in report data must not be an error for an unrelated binding, got: %v", err)
	}
	if got != "Statement for Ada" {
		t.Errorf("got %q", got)
	}
}

// TestBindTextPageAndPagesUnaffectedByParamsRoot is AC17a (AD-4): the
// existing {{page}}/{{pages}} reservation stays byte-for-byte
// unchanged even in a document that ALSO supplies a params document —
// a params implementation that quietly reroutes the reservation branch
// reddens here.
func TestBindTextPageAndPagesUnaffectedByParamsRoot(t *testing.T) {
	data := mustDecode(t, `{"page": "HIJACKED"}`)
	params := mustDecode(t, `{"reportDate": "2026-08-23"}`)
	got, err := BindText("Page {{page}} of {{pages}}, dated {{params.reportDate}}", data, params, "e12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Page {{page}} of {{pages}}, dated 2026-08-23" {
		t.Errorf("got %q", got)
	}
}
