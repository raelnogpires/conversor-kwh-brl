package main

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"conversor-kwh-brl/internal/calculator"
	"conversor-kwh-brl/internal/data"
	"conversor-kwh-brl/internal/validation"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

func run(reader io.Reader, writer io.Writer) error {
	fmt.Fprintln(writer, "Conversor kWh → BRL")
	raw1, raw2, rawAmount, err := data.ReadInput(reader, writer)
	if err != nil {
		return err
	}
	input, err := validation.ParseAndValidate(raw1, raw2, rawAmount)
	if err != nil {
		return err
	}
	result, err := calculator.Calculate(input.Consumption1, input.Consumption2, input.TotalAmountCents)
	if err != nil {
		return err
	}

	fmt.Fprintln(writer, "\nResultado")
	fmt.Fprintf(writer, "Consumo total: %s kWh\n", formatDecimal(result.TotalConsumption))
	fmt.Fprintf(writer, "Proporção consumidor 1: %s (%s%%)\n", result.Share1.RatString(), formatPercentage(result.Share1))
	fmt.Fprintf(writer, "Proporção consumidor 2: %s (%s%%)\n", result.Share2.RatString(), formatPercentage(result.Share2))
	fmt.Fprintf(writer, "Consumidor 1 paga: %s\n", formatBRL(result.Amount1Cents))
	fmt.Fprintf(writer, "Consumidor 2 paga: %s\n", formatBRL(result.Amount2Cents))
	return nil
}

func formatBRL(cents int64) string {
	return fmt.Sprintf("R$ %d,%02d", cents/100, cents%100)
}

func decimalComma(value string) string {
	return strings.Replace(value, ".", ",", 1)
}

func formatPercentage(share *big.Rat) string {
	percentage := new(big.Rat).Mul(new(big.Rat).Set(share), big.NewRat(100, 1))
	return decimalComma(percentage.FloatString(2))
}

// formatDecimal emits an exact finite decimal for values originating from
// plain decimal input. RatString remains a safe fallback for other rationals.
func formatDecimal(value *big.Rat) string {
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
