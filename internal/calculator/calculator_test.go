package calculator

import (
	"math"
	"math/big"
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
		name   string
		first  *big.Rat
		second *big.Rat
		bill   int64
	}{
		{"both zero", rat("0"), rat("0"), 100},
		{"negative first", rat("-1"), rat("2"), 100},
		{"negative second", rat("1"), rat("-2"), 100},
		{"negative bill", rat("1"), rat("2"), -1},
		{"nil first", nil, rat("2"), 100},
		{"nil second", rat("1"), nil, 100},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Calculate(test.first, test.second, test.bill); err == nil {
				t.Fatal("Calculate() error = nil")
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
