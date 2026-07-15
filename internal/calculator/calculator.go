// Package calculator implements the proportional bill split domain rule.
package calculator

import (
	"errors"
	"math/big"
)

// Result is an exact calculation. Rationals are newly allocated and do not
// alias the inputs.
type Result struct {
	TotalConsumption *big.Rat
	Share1           *big.Rat
	Share2           *big.Rat
	Amount1Cents     int64
	Amount2Cents     int64
}

// Calculate splits totalAmountCents in proportion to the two consumptions.
// Amount 1 is rounded to the nearest cent, with an exact half-cent rounded up.
// Amount 2 receives the remainder, guaranteeing that both amounts add up to
// the original bill.
func Calculate(consumption1, consumption2 *big.Rat, totalAmountCents int64) (Result, error) {
	if consumption1 == nil || consumption2 == nil {
		return Result{}, errors.New("consumption cannot be nil")
	}
	if consumption1.Sign() < 0 || consumption2.Sign() < 0 {
		return Result{}, errors.New("consumption cannot be negative")
	}
	if totalAmountCents < 0 {
		return Result{}, errors.New("bill cannot be negative")
	}

	c1 := new(big.Rat).Set(consumption1)
	c2 := new(big.Rat).Set(consumption2)
	total := new(big.Rat).Add(c1, c2)
	if total.Sign() == 0 {
		return Result{}, errors.New("total consumption must be greater than zero")
	}

	share1 := new(big.Rat).Quo(new(big.Rat).Set(c1), total)
	share2 := new(big.Rat).Quo(new(big.Rat).Set(c2), total)
	exactAmount1 := new(big.Rat).Mul(new(big.Rat).SetInt64(totalAmountCents), share1)
	amount1 := roundHalfUp(exactAmount1)

	return Result{
		TotalConsumption: new(big.Rat).Set(total),
		Share1:           new(big.Rat).Set(share1),
		Share2:           new(big.Rat).Set(share2),
		Amount1Cents:     amount1,
		Amount2Cents:     totalAmountCents - amount1,
	}, nil
}

func roundHalfUp(value *big.Rat) int64 {
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(value.Num(), value.Denom(), remainder)
	if new(big.Int).Lsh(remainder, 1).Cmp(value.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.Int64()
}
