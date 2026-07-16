// Package validation interpreta e valida valores decimais fornecidos pelo
// usuário antes que eles entrem nas regras de domínio.
package validation

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Input contém os valores já validados para o domínio. Os consumos ficam em
// big.Rat para representar exatamente o decimal digitado, e TotalAmountCents
// fica em centavos inteiros para que dinheiro nunca dependa de ponto
// flutuante. A invariante produzida por ParseAndValidate é: consumos não
// negativos, soma dos consumos positiva e valor total não negativo.
type Input struct {
	Consumption1     *big.Rat
	Consumption2     *big.Rat
	TotalAmountCents int64
}

// ParseAndValidate interpreta todos os campos do terminal e valida tanto cada
// valor isolado quanto a relação entre eles. Os erros ganham o nome do campo
// para que a camada de interface consiga orientar o usuário sem conhecer os
// detalhes dos parsers.
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

// ParseConsumption interpreta um decimal simples não negativo sem usar
// float64. Depois de normalizar vírgula/ponto, SetString constrói um big.Rat
// exato; por exemplo, "0,1" vira 1/10, e não uma aproximação binária.
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

// ParseMoneyCents converte texto monetário em reais para centavos. São aceitas
// no máximo duas casas decimais e valores fora de int64 são recusados. Não há
// arredondamento aqui: entradas com precisão menor que um centavo são erros,
// para que a conversão seja explícita e sem perda de informação.
func ParseMoneyCents(text string) (int64, error) {
	normalized, negative, err := normalizeDecimal(text, 2)
	if err != nil {
		return 0, err
	}
	if negative {
		return 0, errors.New("não pode ser negativo")
	}

	// Completar a fração à direita transforma "12", "12.3" e "12.34" em,
	// respectivamente, "1200", "1230" e "1234" centavos. A concatenação é
	// segura porque normalizeDecimal já garantiu apenas dígitos e um separador.
	parts := strings.Split(normalized, ".")
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	fraction += strings.Repeat("0", 2-len(fraction))
	centsText := parts[0] + fraction
	// big.Int permite interpretar primeiro sem limite de tamanho; IsInt64 faz
	// a checagem de overflow antes da conversão usada pelo restante do sistema.
	cents, ok := new(big.Int).SetString(centsText, 10)
	if !ok || !cents.IsInt64() {
		return 0, errors.New("valor fora do intervalo permitido")
	}
	return cents.Int64(), nil
}

// normalizeDecimal aceita dígitos com no máximo uma vírgula ou um ponto e
// devolve uma forma canônica com ponto, compreendida por math/big. Um
// fractionalLimit negativo significa quantidade ilimitada de casas.
//
// O sinal é devolvido separadamente: esta função cuida da sintaxe, enquanto os
// parsers públicos decidem se valores negativos são permitidos pelo domínio.
// Separadores de milhares não são inferidos para evitar interpretações
// ambíguas como tratar "1.234" ora como milhar, ora como decimal.
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

	// A varredura byte a byte é suficiente porque a gramática aceita somente
	// caracteres ASCII: dígitos, vírgula, ponto e um eventual '-' inicial.
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
