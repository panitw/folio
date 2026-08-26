package expr

import (
	"fmt"
	"math/big"
	"strings"
)

// This file is formatNumber (AC11-AC14): the first of the two
// functions this story flips to implemented (AC16).

// scaleToFractionDigits rescales d to exactly fracDigits fractional
// digits, using ONLY tenPowLookup's integer table (AC12) — never
// math.Pow, never a float64 anywhere on this path. When d already has
// MORE fractional digits than requested it rounds half-to-even, ONCE,
// at the target scale (AC13); when it has fewer, it pads with zeros
// (exact, no rounding needed). Returns the resulting coefficient as a
// *big.Int (never narrowed to int64 — formatNumber's output is a
// display string, not a Decimal, so there is no reason to reintroduce
// int64's narrower bound here).
func scaleToFractionDigits(d Decimal, fracDigits int) (*big.Int, error) {
	targetExponent := -fracDigits
	coeff := big.NewInt(d.Coefficient)

	if d.Exponent == targetExponent {
		return coeff, nil
	}
	if d.Exponent > targetExponent {
		// Source is coarser than the target scale: pad with zeros —
		// exact, no rounding.
		shift := d.Exponent - targetExponent
		p, err := tenPowLookup(shift)
		if err != nil {
			return nil, err
		}
		return coeff.Mul(coeff, p), nil
	}

	// d.Exponent < targetExponent: source is finer than the target
	// scale — round half-to-even at the target scale (AC13).
	shift := targetExponent - d.Exponent
	divisor, err := tenPowLookup(shift)
	if err != nil {
		return nil, err
	}
	remainder := new(big.Int)
	quotient, _ := new(big.Int).QuoRem(coeff, divisor, remainder)

	absRemainderTwice := new(big.Int).Abs(remainder)
	absRemainderTwice.Lsh(absRemainderTwice, 1)

	switch absRemainderTwice.Cmp(divisor) {
	case 1:
		quotient = roundQuotientAwayFromZero(quotient)
	case 0:
		absQuotient := new(big.Int).Abs(quotient)
		if absQuotient.Bit(0) == 1 { // quotient is odd: round to even
			quotient = roundQuotientAwayFromZero(quotient)
		}
	}
	return quotient, nil
}

// groupDigits inserts sep every groupSize digits, counted from the
// right (the ordinary grouping convention every locale in this table
// shares) — applied to the unsigned integer-part digit string only.
func groupDigits(digits, sep string, groupSize int) string {
	if groupSize <= 0 || len(digits) <= groupSize {
		return digits
	}
	var parts []string
	for len(digits) > groupSize {
		cut := len(digits) - groupSize
		parts = append([]string{digits[cut:]}, parts...)
		digits = digits[:cut]
	}
	parts = append([]string{digits}, parts...)
	return strings.Join(parts, sep)
}

// translateDigits maps each ASCII digit '0'-'9' in s to entry's own
// digit set (AC11a: an explicit named field, Western/ASCII for every
// locale today) — kept as a real translation step, not a no-op
// assumption, so a future table edit to a non-Western digit set works
// with zero change to this function.
func translateDigits(s, digitSet string) string {
	if digitSet == digitsWestern {
		return s
	}
	digits := []rune(digitSet)
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(digits[r-'0'])
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// renderNumberPattern is AC11/AC13/AC14: renders coeff (already scaled
// to spec.fracDigits, AC12/AC13) as a locale-formatted string —
// grouping (AC11), decimal separator (AC11), zero-padding the integer
// part to spec.minIntDigits, and the locale's own digit set (AC11a).
func renderNumberPattern(coeff *big.Int, spec numberPatternSpec, entry localeEntry) string {
	neg := coeff.Sign() < 0
	digits := new(big.Int).Abs(coeff).String()

	fracDigits := spec.fracDigits
	for len(digits) < fracDigits+1 {
		digits = "0" + digits
	}

	intDigits := digits[:len(digits)-fracDigits]
	fracPart := digits[len(digits)-fracDigits:]

	for len(intDigits) < spec.minIntDigits {
		intDigits = "0" + intDigits
	}

	if spec.grouping {
		intDigits = groupDigits(intDigits, entry.groupSep, entry.groupSize)
	}

	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	b.WriteString(intDigits)
	if fracDigits > 0 {
		b.WriteString(entry.decimalSep)
		b.WriteString(fracPart)
	}
	return translateDigits(b.String(), entry.digits)
}

// evalFormatNumber is AC11-AC14's evaluator.
func evalFormatNumber(call *CallExpr, resolver Resolver, fc FormatContext, elementID string) (Value, []Caveat, error) {
	entry, lerr := lookupLocale(fc.Locale)
	if lerr != nil {
		return Value{}, nil, fmt.Errorf("expr: element %s: formatNumber(): %w", elementID, lerr)
	}

	operandVal, caveats, err := Eval(call.Args[0], resolver, fc, elementID)
	if err != nil {
		return Value{}, nil, err
	}
	if operandVal.Kind != KindNumber {
		return Value{}, nil, fmt.Errorf(
			"expr: element %s: formatNumber(): operand must be a number, got %s (never coerced): %s",
			elementID, operandVal.Kind, call.Raw,
		)
	}

	patternLit, ok := call.Args[1].(*StringLit)
	if !ok {
		return Value{}, nil, fmt.Errorf("expr: element %s: formatNumber(): pattern argument must be a string literal: %s", elementID, call.Raw)
	}
	spec, perr := validateNumberPattern(patternLit.Value)
	if perr != nil {
		return Value{}, nil, fmt.Errorf("expr: element %s: formatNumber(): internal: %w", elementID, perr)
	}

	scaled, serr := scaleToFractionDigits(operandVal.Num, spec.fracDigits)
	if serr != nil {
		return Value{}, nil, fmt.Errorf("expr: element %s: formatNumber(): %w: %s", elementID, serr, call.Raw)
	}

	return Value{Kind: KindString, Str: renderNumberPattern(scaled, spec, entry)}, caveats, nil
}
