package main

import (
	"fmt"
	"io"
	"os"

	"rateio-luz/internal/calculator"
	"rateio-luz/internal/data"
	"rateio-luz/internal/presentation"
	"rateio-luz/internal/validation"
)

func main() {
	// Este é o ponto de entrada do executável de terminal. Os fluxos padrão do
	// processo ficam restritos a esta borda, enquanto run permanece reutilizável.
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// run recebe Reader e Writer por injeção para não depender diretamente do
// terminal; assim, a mesma interação pode usar arquivos, buffers ou dublês.
func run(reader io.Reader, writer io.Writer) error {
	fmt.Fprintln(writer, "Rateio Luz — Conversor kWh → BRL")
	// A CLI atua como ponte entre as camadas: coleta texto bruto, entrega-o à
	// validação, encaminha os valores tipados ao cálculo e formata a resposta.
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

	// Somente a apresentação conhece a representação textual adequada de kWh,
	// percentuais e moeda; o resultado do domínio continua independente da CLI.
	fmt.Fprintln(writer, "\nResultado")
	fmt.Fprintf(writer, "Consumo total: %s\n", presentation.KWh(result.TotalConsumption))
	fmt.Fprintf(writer, "Proporção consumidor 1: %s (%s%%)\n", result.Share1.RatString(), presentation.Percentage(result.Share1))
	fmt.Fprintf(writer, "Proporção consumidor 2: %s (%s%%)\n", result.Share2.RatString(), presentation.Percentage(result.Share2))
	fmt.Fprintf(writer, "Consumidor 1 paga: %s\n", presentation.BRL(result.Amount1Cents))
	fmt.Fprintf(writer, "Consumidor 2 paga: %s\n", presentation.BRL(result.Amount2Cents))
	return nil
}
