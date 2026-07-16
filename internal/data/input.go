// Package data coleta a entrada textual usada pela interface de terminal.
package data

import (
	"bufio"
	"fmt"
	"io"
)

// ReadInput exibe as perguntas e devolve as três linhas sem interpretá-las;
// conversão e validação pertencem às camadas seguintes do fluxo.
func ReadInput(reader io.Reader, writer io.Writer) (consumption1, consumption2, totalAmount string, err error) {
	// O Scanner percorre o Reader uma linha por vez, o que desacopla a coleta de
	// os.Stdin e preserva exatamente o texto que será validado posteriormente.
	scanner := bufio.NewScanner(reader)
	// Ponteiros alinham cada linha lida ao retorno correspondente e evitam
	// duplicar o mesmo ciclo de pergunta, leitura e tratamento de falhas.
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
			// Uma falha do Reader é distinta do encerramento normal antes de todas
			// as respostas, por isso cada situação produz um erro próprio.
			if scanner.Err() != nil {
				return "", "", "", fmt.Errorf("ler entrada: %w", scanner.Err())
			}
			return "", "", "", io.ErrUnexpectedEOF
		}
		*values[i] = scanner.Text()
	}

	return consumption1, consumption2, totalAmount, nil
}
