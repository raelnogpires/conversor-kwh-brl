package validation

import (
	"math"
	"math/big"
	"strings"
	"testing"
)

func TestParseConsumptionAcceptsCommaAndDot(t *testing.T) {
	for _, text := range []string{"12,345", "12.345", "0", " 10,5 "} {
		got, err := ParseConsumption(text)
		if err != nil {
			t.Errorf("ParseConsumption(%q): %v", text, err)
			continue
		}
		normalized := strings.ReplaceAll(strings.TrimSpace(text), ",", ".")
		want, _ := new(big.Rat).SetString(normalized)
		if got.Cmp(want) != 0 {
			t.Errorf("ParseConsumption(%q) = %s, want %s", text, got, want)
		}
	}
}

func TestParseConsumptionRejectsInvalidInput(t *testing.T) {
	for _, text := range []string{"", " ", "abc", "1e3", "+1", "-1", "NaN", "Inf", "-Inf", ".5", "5,", "1,234.5", "1.234,5", "1,2,3"} {
		t.Run(text, func(t *testing.T) {
			if _, err := ParseConsumption(text); err == nil {
				t.Fatalf("ParseConsumption(%q) error = nil", text)
			}
		})
	}
}

func TestParseMoneyCentsAcceptsCommaDotAndZero(t *testing.T) {
	tests := map[string]int64{
		"0":      0,
		"0,00":   0,
		"10":     1000,
		"10.5":   1050,
		"184,72": 18472,
	}
	for text, want := range tests {
		got, err := ParseMoneyCents(text)
		if err != nil {
			t.Errorf("ParseMoneyCents(%q): %v", text, err)
		} else if got != want {
			t.Errorf("ParseMoneyCents(%q) = %d, want %d", text, got, want)
		}
	}
}

func TestParseMoneyCentsRejectsInvalidInput(t *testing.T) {
	for _, text := range []string{"", "abc", "-0.01", "NaN", "Inf", "1,000.00", "1.000,00", "1,234", "1.234", ".50", "50,", "+1", "92233720368547758.08"} {
		t.Run(text, func(t *testing.T) {
			if _, err := ParseMoneyCents(text); err == nil {
				t.Fatalf("ParseMoneyCents(%q) error = nil", text)
			}
		})
	}
}

func TestParseAndValidateRelationship(t *testing.T) {
	if _, err := ParseAndValidate("0", "0", "10"); err == nil {
		t.Fatal("both zero consumptions accepted")
	}
	input, err := ParseAndValidate("0", "1", "0")
	if err != nil {
		t.Fatalf("valid zero bill rejected: %v", err)
	}
	if input.TotalAmountCents != 0 {
		t.Errorf("bill = %d", input.TotalAmountCents)
	}
}

func TestParseMoneyMaximumInt64(t *testing.T) {
	got, err := ParseMoneyCents("92233720368547758.07")
	if err != nil || got != math.MaxInt64 {
		t.Fatalf("got %d, %v; want MaxInt64", got, err)
	}
}
