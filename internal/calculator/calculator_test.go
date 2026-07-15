package calculator

import (
	"math"
	"math/big"
	"strings"
	"testing"
)

func rat(text string) *big.Rat {
	value, ok := new(big.Rat).SetString(text)
	if !ok {
		panic("invalid test rational: " + text)
	}
	return value
}

func TestCalculateExpectedExample(t *testing.T) {
	result, err := Calculate(rat("105.5"), rat("67.2"), 18472)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalConsumption.Cmp(rat("172.7")) != 0 {
		t.Errorf("total = %s, want 172.7", result.TotalConsumption)
	}
	if result.Share1.Cmp(rat("1055/1727")) != 0 || result.Share2.Cmp(rat("672/1727")) != 0 {
		t.Errorf("shares = %s and %s", result.Share1, result.Share2)
	}
	if result.Amount1Cents != 11284 || result.Amount2Cents != 7188 {
		t.Errorf("amounts = %d and %d, want 11284 and 7188", result.Amount1Cents, result.Amount2Cents)
	}
}

func TestCalculateConsumptionScenarios(t *testing.T) {
	tests := []struct {
		name             string
		first            string
		second           string
		billCents        int64
		wantTotal        string
		wantShare1       string
		wantShare2       string
		wantAmount1Cents int64
		wantAmount2Cents int64
	}{
		{
			name:             "equal consumption",
			first:            "100",
			second:           "100",
			billCents:        10001,
			wantTotal:        "200",
			wantShare1:       "1/2",
			wantShare2:       "1/2",
			wantAmount1Cents: 5001,
			wantAmount2Cents: 5000,
		},
		{
			name:             "different consumption",
			first:            "30",
			second:           "70",
			billCents:        10000,
			wantTotal:        "100",
			wantShare1:       "3/10",
			wantShare2:       "7/10",
			wantAmount1Cents: 3000,
			wantAmount2Cents: 7000,
		},
		{
			name:             "decimal consumption",
			first:            "0.125",
			second:           "0.375",
			billCents:        12345,
			wantTotal:        "0.5",
			wantShare1:       "1/4",
			wantShare2:       "3/4",
			wantAmount1Cents: 3086,
			wantAmount2Cents: 9259,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Calculate(rat(test.first), rat(test.second), test.billCents)
			if err != nil {
				t.Fatalf("Calculate() error = %v", err)
			}

			if result.TotalConsumption.Cmp(rat(test.wantTotal)) != 0 {
				t.Errorf("TotalConsumption = %s, want %s", result.TotalConsumption, test.wantTotal)
			}
			if result.Share1.Cmp(rat(test.wantShare1)) != 0 {
				t.Errorf("Share1 = %s, want %s", result.Share1, test.wantShare1)
			}
			if result.Share2.Cmp(rat(test.wantShare2)) != 0 {
				t.Errorf("Share2 = %s, want %s", result.Share2, test.wantShare2)
			}
			if result.Amount1Cents != test.wantAmount1Cents || result.Amount2Cents != test.wantAmount2Cents {
				t.Errorf(
					"amounts = %d and %d, want %d and %d",
					result.Amount1Cents,
					result.Amount2Cents,
					test.wantAmount1Cents,
					test.wantAmount2Cents,
				)
			}
		})
	}
}

func TestCalculateRoundsBelowAtAndAboveHalfCent(t *testing.T) {
	tests := []struct {
		name             string
		first            string
		second           string
		wantAmount1Cents int64
	}{
		{name: "below half cent", first: "149", second: "151", wantAmount1Cents: 1},
		{name: "exactly half cent", first: "1", second: "1", wantAmount1Cents: 2},
		{name: "above half cent", first: "151", second: "149", wantAmount1Cents: 2},
	}

	const billCents int64 = 3
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Calculate(rat(test.first), rat(test.second), billCents)
			if err != nil {
				t.Fatalf("Calculate() error = %v", err)
			}
			if result.Amount1Cents != test.wantAmount1Cents {
				t.Errorf("Amount1Cents = %d, want %d", result.Amount1Cents, test.wantAmount1Cents)
			}
			if result.Amount1Cents+result.Amount2Cents != billCents {
				t.Errorf("amount sum = %d, want %d", result.Amount1Cents+result.Amount2Cents, billCents)
			}
		})
	}
}

func TestCalculateAlwaysReconcilesToOriginalBill(t *testing.T) {
	tests := []struct {
		name      string
		first     string
		second    string
		billCents int64
	}{
		{name: "equal split of one cent", first: "1", second: "1", billCents: 1},
		{name: "different consumption", first: "1", second: "2", billCents: 100},
		{name: "decimal consumption", first: "0.1", second: "0.2", billCents: 999},
		{name: "first consumer zero", first: "0", second: "7.75", billCents: 12345},
		{name: "second consumer zero", first: "7.75", second: "0", billCents: 12345},
		{name: "zero bill", first: "3.2", second: "4.8", billCents: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Calculate(rat(test.first), rat(test.second), test.billCents)
			if err != nil {
				t.Fatalf("Calculate() error = %v", err)
			}
			if got := result.Amount1Cents + result.Amount2Cents; got != test.billCents {
				t.Errorf("amount sum = %d, want original bill %d", got, test.billCents)
			}
		})
	}
}

func TestCalculateZeroConsumer(t *testing.T) {
	tests := []struct {
		name       string
		first      string
		second     string
		wantFirst  int64
		wantSecond int64
	}{
		{"consumer 1 zero", "0", "20", 0, 999},
		{"consumer 2 zero", "20", "0", 999, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Calculate(rat(test.first), rat(test.second), 999)
			if err != nil {
				t.Fatal(err)
			}
			if result.Amount1Cents != test.wantFirst || result.Amount2Cents != test.wantSecond {
				t.Errorf("amounts = %d and %d", result.Amount1Cents, result.Amount2Cents)
			}
		})
	}
}

func TestCalculateRejectsInvalidDomain(t *testing.T) {
	tests := []struct {
		name      string
		first     *big.Rat
		second    *big.Rat
		bill      int64
		wantError string
	}{
		{"both zero", rat("0"), rat("0"), 100, "total consumption"},
		{"negative first", rat("-1"), rat("2"), 100, "consumption cannot be negative"},
		{"negative second", rat("1"), rat("-2"), 100, "consumption cannot be negative"},
		{"negative bill", rat("1"), rat("2"), -1, "bill cannot be negative"},
		{"nil first", nil, rat("2"), 100, "consumption cannot be nil"},
		{"nil second", rat("1"), nil, 100, "consumption cannot be nil"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Calculate(test.first, test.second, test.bill)
			if err == nil {
				t.Fatal("Calculate() error = nil")
			}
			if !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("Calculate() error = %q, want it to contain %q", err, test.wantError)
			}
		})
	}
}

func TestCalculateZeroBillAndDecimalConsumption(t *testing.T) {
	result, err := Calculate(rat("0.125"), rat("0.375"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalConsumption.Cmp(rat("0.5")) != 0 || result.Amount1Cents != 0 || result.Amount2Cents != 0 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestCalculateHalfUpAndRemainder(t *testing.T) {
	result, err := Calculate(rat("1"), rat("1"), 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Amount1Cents != 1 || result.Amount2Cents != 0 {
		t.Errorf("amounts = %d and %d, want half-up 1 and remainder 0", result.Amount1Cents, result.Amount2Cents)
	}
	if result.Amount1Cents+result.Amount2Cents != 1 {
		t.Fatal("payable sum differs from bill")
	}
}

func TestCalculateSharesSumExactlyAndApproximately(t *testing.T) {
	result, err := Calculate(rat("1"), rat("2"), 100)
	if err != nil {
		t.Fatal(err)
	}
	sum := new(big.Rat).Add(result.Share1, result.Share2)
	if sum.Cmp(big.NewRat(1, 1)) != 0 {
		t.Errorf("exact share sum = %s", sum)
	}
	first, _ := result.Share1.Float64()
	second, _ := result.Share2.Float64()
	if math.Abs(first+second-1) > 1e-15 {
		t.Errorf("approximate share sum = %.17g", first+second)
	}
}

func TestCalculateVeryLargeValuesAndDoesNotMutateInputs(t *testing.T) {
	first := rat("999999999999999999999999999999.123456789")
	second := rat("888888888888888888888888888888.987654321")
	firstBefore := new(big.Rat).Set(first)
	secondBefore := new(big.Rat).Set(second)
	result, err := Calculate(first, second, math.MaxInt64)
	if err != nil {
		t.Fatal(err)
	}
	if first.Cmp(firstBefore) != 0 || second.Cmp(secondBefore) != 0 {
		t.Fatal("Calculate mutated an input")
	}
	if result.Amount1Cents+result.Amount2Cents != math.MaxInt64 {
		t.Fatal("large payable amounts do not add to bill")
	}
}
