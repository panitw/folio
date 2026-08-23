package template

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// This file is the "points" numeric kind's exact decimal path (AC24,
// D-1.4.3): a json.Number literal (never a float64 — encoding/json's
// default decodes every number to float64, which is why UseNumber is
// mandatory, AC26) is converted to geom.Length millipoints by exact
// ×1000 decimal-string arithmetic (math/big.Int only — never
// strconv.ParseFloat, never float64). The on-disk spelling this
// produces (appendPoints) matches internal/pdf's appendLength byte for
// byte (AC25), by independent reimplementation rather than a shared
// symbol, because internal/template may not import internal/pdf
// (AD-1). folio-go/testdata/template/length-spelling-cases.txt is the
// shared value table (AC25) both packages' tests assert against, so a
// future divergence is a red test rather than a byte surprise.

// decodePoints converts a JSON number literal (as produced by
// json.Decoder.UseNumber) into millipoints. Non-canonical but valid
// spellings are accepted and normalised (AC27): "36.0000" and "1e3" are
// both legal. A value that is not an exact whole number of millipoints
// is a load error (AC28) — nothing is rounded. int64 millipoint
// overflow is a load error, never a wrap (AC27).
func decodePoints(literal string) (geom.Length, error) {
	sign, intPart, fracPart, exp, err := splitJSONNumber(literal)
	if err != nil {
		return 0, err
	}

	// digits*10^(exp - len(fracPart)) is the exact decimal value; we
	// want that value * 1000, i.e. digits * 10^(exp - len(fracPart) + 3).
	digitsStr := intPart + fracPart
	digits := new(big.Int)
	if _, ok := digits.SetString(digitsStr, 10); !ok {
		return 0, fmt.Errorf("template: invalid numeric literal %q", literal)
	}

	shift := exp - len(fracPart) + 3

	var millipoints *big.Int
	if shift >= 0 {
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(shift)), nil)
		millipoints = new(big.Int).Mul(digits, pow)
	} else {
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-shift)), nil)
		q, r := new(big.Int), new(big.Int)
		q.QuoRem(digits, pow, r)
		if r.Sign() != 0 {
			return 0, fmt.Errorf("template: value %q has more than three decimal places (not an exact whole number of millipoints)", literal)
		}
		millipoints = q
	}

	if sign < 0 {
		millipoints.Neg(millipoints)
	}

	if !millipoints.IsInt64() {
		return 0, fmt.Errorf("template: value %q overflows int64 millipoints", literal)
	}
	return geom.Length(millipoints.Int64()), nil
}

// splitJSONNumber splits a valid JSON number literal into its sign
// (1 or -1), integer digit string, fractional digit string (without the
// '.') and decimal exponent (0 if absent). It does not itself validate
// full JSON number grammar — literal always comes from
// encoding/json's own number scanner (via json.Number), which already
// guarantees that.
func splitJSONNumber(literal string) (sign int, intPart, fracPart string, exp int, err error) {
	s := literal
	sign = 1
	if strings.HasPrefix(s, "-") {
		sign = -1
		s = s[1:]
	}

	mantissa := s
	expPart := ""
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		mantissa = s[:i]
		expPart = s[i+1:]
	}

	if i := strings.IndexByte(mantissa, '.'); i >= 0 {
		intPart = mantissa[:i]
		fracPart = mantissa[i+1:]
	} else {
		intPart = mantissa
	}

	if expPart != "" {
		e, perr := parseDecimalExponent(expPart)
		if perr != nil {
			return 0, "", "", 0, fmt.Errorf("template: invalid exponent in numeric literal %q: %w", literal, perr)
		}
		exp = e
	}
	return sign, intPart, fracPart, exp, nil
}

func parseDecimalExponent(s string) (int, error) {
	neg := false
	if strings.HasPrefix(s, "+") {
		s = s[1:]
	} else if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	if s == "" {
		return 0, fmt.Errorf("empty exponent")
	}
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("non-digit %q in exponent", c)
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}

// appendPoints appends the canonical on-disk spelling of a geom.Length
// to dst: sign, integer part, and up to three fractional digits with
// trailing zeros (and a bare trailing point) trimmed. This is a
// deliberate byte-for-byte reimplementation of internal/pdf's
// appendLength (AC25) — see this file's header comment for why it is
// not shared, and length-spelling_test.go for the cross-check that
// keeps the two in step.
func appendPoints(dst []byte, v geom.Length) []byte {
	n := int64(v)
	negative := n < 0

	i := n / 1000
	f := n % 1000
	if i < 0 {
		i = -i
	}
	if f < 0 {
		f = -f
	}

	if negative && (i != 0 || f != 0) {
		dst = append(dst, '-')
	}

	dst = appendPlainInt(dst, i)

	if f != 0 {
		dst = append(dst, '.')
		digits := [3]byte{
			byte('0' + f/100),
			byte('0' + (f/10)%10),
			byte('0' + f%10),
		}
		end := 3
		for end > 0 && digits[end-1] == '0' {
			end--
		}
		dst = append(dst, digits[:end]...)
	}

	return dst
}

// appendPlainInt appends the plain decimal digits of a non-negative
// int64 to dst (used for both appendPoints' integer part and NextID).
func appendPlainInt(dst []byte, v int64) []byte {
	if v == 0 {
		return append(dst, '0')
	}
	var buf [20]byte
	pos := len(buf)
	for v != 0 {
		pos--
		buf[pos] = byte('0' + v%10)
		v /= 10
	}
	return append(dst, buf[pos:]...)
}
