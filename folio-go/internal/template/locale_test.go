package template

import "testing"

// TestClosedLocalesMatchesLocaleTags is AC4's own-package half: the
// exported ordered LocaleTags and the unexported closedLocales lookup
// used by parse.go must name exactly the same set. Both are built from
// the same four constants (locale.go), so this is a construction
// check, not a coincidence — but it stays as an explicit assertion
// because "built from the same constants" is a property of the source
// text, not something the compiler enforces for two independently
// maintained declarations.
func TestClosedLocalesMatchesLocaleTags(t *testing.T) {
	if len(LocaleTags) == 0 {
		t.Fatal("presence precondition (D-000.9): LocaleTags is empty")
	}
	seen := map[string]bool{}
	for _, tag := range LocaleTags {
		seen[tag] = true
		if !closedLocales[tag] {
			t.Errorf("LocaleTags contains %q, but closedLocales does not", tag)
		}
	}
	for tag := range closedLocales {
		if !seen[tag] {
			t.Errorf("closedLocales contains %q, but LocaleTags does not", tag)
		}
	}
}

// TestLocaleTagsExactOrder is AC4a: the exported order is asserted as
// an exact literal sequence, not "some deterministic order" — pinning
// the sequence itself so a later switch to (e.g.) sorted order is a
// visible, deliberate edit to this test rather than a silent reorder.
func TestLocaleTagsExactOrder(t *testing.T) {
	want := []string{LocaleEN, LocaleTH, LocaleZhHans, LocaleJA}
	if len(LocaleTags) != len(want) {
		t.Fatalf("LocaleTags has %d entries, want %d: %v", len(LocaleTags), len(want), LocaleTags)
	}
	for i, tag := range want {
		if LocaleTags[i] != tag {
			t.Errorf("LocaleTags[%d] = %q, want %q (exact order pinned, AC4a)", i, LocaleTags[i], tag)
		}
	}
}
