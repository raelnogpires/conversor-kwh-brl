// Package presentation formata valores do domínio para exibição nas
// interfaces com o usuário.
package presentation

import (
	"fmt"
	"math/big"
	"strings"
)

// BRL formata um valor representado em centavos como reais brasileiros. O
// contrato do fluxo atual fornece centavos não negativos; divisão e resto por
// 100 separam, sem ponto flutuante, a parte inteira das duas casas decimais.
func BRL(cents int64) string {
	return fmt.Sprintf("R$ %d,%02d", cents/100, cents%100)
}

// Percentage formata uma participação racional exata como porcentagem com
// duas casas decimais. A multiplicação continua em big.Rat e FloatString faz o
// arredondamento somente na fronteira de apresentação, sem contaminar o
// cálculo de domínio. Como cada participação é formatada separadamente, os dois
// textos arredondados podem somar 99,99% ou 100,01%, embora as frações originais
// continuem somando exatamente 1.
func Percentage(share *big.Rat) string {
	percentage := new(big.Rat).Mul(new(big.Rat).Set(share), big.NewRat(100, 1))
	return decimalComma(percentage.FloatString(2))
}

// Decimal exibe como decimal exato os valores vindos de entradas decimais
// simples. Uma fração reduzida possui representação decimal finita somente se
// seu denominador tiver 2 e 5 como únicos fatores primos. Para qualquer outro
// racional, RatString é a alternativa exata e evita inventar uma precisão ou
// um arredondamento para uma dízima periódica.
func Decimal(value *big.Rat) string {
	// big.Int é mutável; copiamos o denominador antes de retirar seus fatores
	// para não alterar o big.Rat recebido pelo chamador.
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

	// Após reduzir a fração, max(expoente de 2, expoente de 5) é a quantidade
	// de casas necessária para escrever seu decimal finito por completo.
	digits := twos
	if fives > digits {
		digits = fives
	}
	formatted := value.FloatString(digits)
	// Zeros finais são apenas artefatos da largura necessária a FloatString;
	// removê-los preserva o valor e mantém a saída amigável.
	if strings.Contains(formatted, ".") {
		formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
	}
	return decimalComma(formatted)
}

// KWh formata um consumo exato e acrescenta sua unidade. A conversão numérica
// permanece centralizada em Decimal para manter o mesmo separador e precisão.
func KWh(value *big.Rat) string {
	return Decimal(value) + " kWh"
}

func decimalComma(value string) string {
	// As rotinas de math/big usam ponto; a apresentação brasileira usa vírgula.
	return strings.Replace(value, ".", ",", 1)
}
