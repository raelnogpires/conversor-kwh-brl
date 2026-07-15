// Package validation parses and validates user-provided decimal values.
package validation

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Input contains validated domain values. TotalAmountCents is integer cents.
type Input struct {
	Consumption1     *big.Rat
	Consumption2     *big.Rat
	TotalAmountCents int64
}

// ParseAndValidate parses all terminal fields and validates their relationship.
func ParseAndValidate(consumption1, consumption2, totalAmount string) (Input, error) {
	c1, err := ParseConsumption(consumption1)
	if err != nil {
		return Input{}, fmt.Errorf("consumo do consumidor 1: %w", err)
	}
	c2, err := ParseConsumption(consumption2)
	if err != nil {
		return Input{}, fmt.Errorf("consumo do consumidor 2: %w", err)
	}
	amount, err := ParseMoneyCents(totalAmount)
	if err != nil {
		return Input{}, fmt.Errorf("valor total da conta: %w", err)
	}
	if c1.Sign() == 0 && c2.Sign() == 0 {
		return Input{}, errors.New("o consumo total deve ser maior que zero")
	}
	return Input{Consumption1: c1, Consumption2: c2, TotalAmountCents: amount}, nil
}

// ParseConsumption parses a non-negative plain decimal without using float64.
func ParseConsumption(text string) (*big.Rat, error) {
	normalized, negative, err := normalizeDecimal(text, -1)
	if err != nil {
		return nil, err
	}
	if negative {
		return nil, errors.New("não pode ser negativo")
	}
	value, ok := new(big.Rat).SetString(normalized)
	if !ok {
		return nil, errors.New("número inválido")
	}
	return value, nil
}

// ParseMoneyCents parses BRL decimal text into cents. At most two decimal
// places are accepted, and overflow outside int64 is rejected.
func ParseMoneyCents(text string) (int64, error) {
	normalized, negative, err := normalizeDecimal(text, 2)
	if err != nil {
		return 0, err
	}
	if negative {
		return 0, errors.New("não pode ser negativo")
	}

	parts := strings.Split(normalized, ".")
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	fraction += strings.Repeat("0", 2-len(fraction))
	centsText := parts[0] + fraction
	cents, ok := new(big.Int).SetString(centsText, 10)
	if !ok || !cents.IsInt64() {
		return 0, errors.New("valor fora do intervalo permitido")
	}
	return cents.Int64(), nil
}

// normalizeDecimal accepts digits with at most one comma or dot. The
// fractionalLimit is unlimited when negative.
func normalizeDecimal(text string, fractionalLimit int) (normalized string, negative bool, err error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", false, errors.New("entrada vazia")
	}
	if text[0] == '-' {
		negative = true
		text = text[1:]
	}
	if text == "" {
		return "", false, errors.New("número inválido")
	}

	separator := byte(0)
	separatorAt := -1
	for i := 0; i < len(text); i++ {
		character := text[i]
		switch {
		case character >= '0' && character <= '9':
		case character == '.' || character == ',':
			if separator != 0 {
				return "", false, errors.New("use apenas um separador decimal, sem separador de milhares")
			}
			separator = character
			separatorAt = i
		default:
			return "", false, errors.New("número inválido")
		}
	}
	if separatorAt == 0 || separatorAt == len(text)-1 {
		return "", false, errors.New("número inválido")
	}
	if fractionalLimit >= 0 && separatorAt >= 0 && len(text)-separatorAt-1 > fractionalLimit {
		return "", false, fmt.Errorf("use no máximo %d casas decimais", fractionalLimit)
	}
	if separator != 0 {
		text = text[:separatorAt] + "." + text[separatorAt+1:]
	}
	return text, negative, nil
}
