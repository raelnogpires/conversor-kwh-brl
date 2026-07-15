// Package data handles raw terminal input.
package data

import (
	"bufio"
	"fmt"
	"io"
)

// ReadInput prompts for and returns the three input lines without parsing them.
func ReadInput(reader io.Reader, writer io.Writer) (consumption1, consumption2, totalAmount string, err error) {
	scanner := bufio.NewScanner(reader)
	values := []*string{&consumption1, &consumption2, &totalAmount}
	prompts := []string{
		"Consumo do consumidor 1 (kWh): ",
		"Consumo do consumidor 2 (kWh): ",
		"Valor total da conta (R$): ",
	}

	for i, prompt := range prompts {
		if _, err = fmt.Fprint(writer, prompt); err != nil {
			return "", "", "", fmt.Errorf("escrever pergunta: %w", err)
		}
		if !scanner.Scan() {
			if scanner.Err() != nil {
				return "", "", "", fmt.Errorf("ler entrada: %w", scanner.Err())
			}
			return "", "", "", io.ErrUnexpectedEOF
		}
		*values[i] = scanner.Text()
	}

	return consumption1, consumption2, totalAmount, nil
}
