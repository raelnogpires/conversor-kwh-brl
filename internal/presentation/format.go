// Package presentation formats domain values for user interfaces.
package presentation

import (
	"fmt"
	"math/big"
	"strings"
)

// BRL formats an amount represented in cents as Brazilian reais.
func BRL(cents int64) string {
	return fmt.Sprintf("R$ %d,%02d", cents/100, cents%100)
}

// Percentage formats an exact share as a percentage with two decimal places.
func Percentage(share *big.Rat) string {
	percentage := new(big.Rat).Mul(new(big.Rat).Set(share), big.NewRat(100, 1))
	return decimalComma(percentage.FloatString(2))
}

// Decimal emits an exact finite decimal for values originating from plain
// decimal input. RatString remains a safe fallback for other rationals.
func Decimal(value *big.Rat) string {
	denominator := new(big.Int).Set(value.Denom())
	twos, fives := 0, 0
	for new(big.Int).Mod(denominator, big.NewInt(2)).Sign() == 0 {
		denominator.Quo(denominator, big.NewInt(2))
		twos++
	}
	for new(big.Int).Mod(denominator, big.NewInt(5)).Sign() == 0 {
		denominator.Quo(denominator, big.NewInt(5))
		fives++
	}
	if denominator.Cmp(big.NewInt(1)) != 0 {
		return value.RatString()
	}

	digits := twos
	if fives > digits {
		digits = fives
	}
	formatted := value.FloatString(digits)
	if strings.Contains(formatted, ".") {
		formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
	}
	return decimalComma(formatted)
}

// KWh formats an exact consumption with its unit.
func KWh(value *big.Rat) string {
	return Decimal(value) + " kWh"
}

func decimalComma(value string) string {
	return strings.Replace(value, ".", ",", 1)
}
