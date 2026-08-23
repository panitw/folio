package template

import (
	"testing"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// TestDecodePointsNormalisesNonCanonicalInput is AC27: "36.0000" -> 36000
// millipoints, "1e3" -> 1000000 millipoints (1000 points); legality is a
// property of the value, not the spelling.
func TestDecodePointsNormalisesNonCanonicalInput(t *testing.T) {
	cases := []struct {
		literal string
		want    int64 // millipoints
	}{
		{"36", 36000},
		{"36.0000", 36000},
		{"1e3", 1000000},
		{"1E3", 1000000},
		{"72.5", 72500},
		{"-5.5", -5500},
		{"0", 0},
	}
	for _, c := range cases {
		got, err := decodePoints(c.literal)
		if err != nil {
			t.Errorf("decodePoints(%q): unexpected error: %v", c.literal, err)
			continue
		}
		if int64(got) != c.want {
			t.Errorf("decodePoints(%q) = %d, want %d", c.literal, int64(got), c.want)
		}
	}
}

// TestDecodePointsRejectsExtraDecimals is AC28: a value that is not an
// exact whole number of millipoints is a load error — nothing is
// rounded. Both polarities: 36.0000 legal, 72.5001 a load error.
func TestDecodePointsRejectsExtraDecimals(t *testing.T) {
	if _, err := decodePoints("36.0000"); err != nil {
		t.Errorf("36.0000 must be legal (exactly 36000 millipoints): %v", err)
	}
	if _, err := decodePoints("72.5001"); err == nil {
		t.Error("72.5001 must be a load error (more than three decimal places)")
	}
}

// TestDecodePointsOverflow is AC27: int64 millipoint overflow is a load
// error, never a wrap.
func TestDecodePointsOverflow(t *testing.T) {
	if _, err := decodePoints("99999999999999999999999999"); err == nil {
		t.Error("expected an overflow load error")
	}
}

// TestAppendPointsSpelling is AC25's on-disk spelling: trailing zeros
// trimmed, no trailing '.', no '-0', no exponent.
func TestAppendPointsSpelling(t *testing.T) {
	cases := []struct {
		millipoints int64
		want        string
	}{
		{36000, "36"},
		{72500, "72.5"},
		{0, "0"},
		{-36000, "-36"},
		{1, "0.001"},
		{-1, "-0.001"},
	}
	for _, c := range cases {
		got := string(appendPoints(nil, geom.Length(c.millipoints)))
		if got != c.want {
			t.Errorf("appendPoints(%d) = %q, want %q", c.millipoints, got, c.want)
		}
	}
}
