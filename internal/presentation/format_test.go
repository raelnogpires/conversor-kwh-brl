package presentation

import (
	"math/big"
	"testing"
)

func testRat(t *testing.T, text string) *big.Rat {
	t.Helper()
	value, ok := new(big.Rat).SetString(text)
	if !ok {
		t.Fatalf("invalid test rational %q", text)
	}
	return value
}

func TestBRL(t *testing.T) {
	tests := map[int64]string{
		0:     "R$ 0,00",
		1:     "R$ 0,01",
		18472: "R$ 184,72",
	}
	for cents, want := range tests {
		if got := BRL(cents); got != want {
			t.Errorf("BRL(%d) = %q, want %q", cents, got, want)
		}
	}
}

func TestDecimalAndKWh(t *testing.T) {
	tests := map[string]string{
		"172.700": "172,7",
		"1/3":     "1/3",
		"0.125":   "0,125",
		"20":      "20",
	}
	for input, want := range tests {
		value := testRat(t, input)
		if got := Decimal(value); got != want {
			t.Errorf("Decimal(%s) = %q, want %q", input, got, want)
		}
	}
	if got := KWh(testRat(t, "10.5")); got != "10,5 kWh" {
		t.Errorf("KWh() = %q", got)
	}
}

func TestPercentageRoundsToTwoDecimalPlaces(t *testing.T) {
	if got := Percentage(testRat(t, "1055/1727")); got != "61,09" {
		t.Errorf("Percentage() = %q, want 61,09", got)
	}
}
