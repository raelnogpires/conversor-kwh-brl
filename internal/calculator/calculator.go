// Package calculator implementa a regra de domínio que divide uma conta
// proporcionalmente ao consumo de duas pessoas.
package calculator

import (
	"errors"
	"math/big"
)

// Result reúne os valores derivados do cálculo proporcional.
//
// Consumos e participações usam big.Rat para preservar exatamente as frações:
// não há a imprecisão binária que ocorreria com float64. Cada *big.Rat do
// resultado é uma nova instância e não compartilha memória mutável com os
// argumentos recebidos. Os valores monetários já estão em centavos inteiros;
// a invariante de reconciliação é Amount1Cents + Amount2Cents igual ao total
// informado a Calculate.
type Result struct {
	TotalConsumption *big.Rat
	Share1           *big.Rat
	Share2           *big.Rat
	Amount1Cents     int64
	Amount2Cents     int64
}

// Calculate divide totalAmountCents na proporção entre os dois consumos.
// O contrato exige ponteiros de consumo não nil, valores não negativos, conta
// não negativa e consumo total maior que zero. Um dos consumos pode ser zero.
// O valor da pessoa 1 é arredondado ao centavo
// mais próximo pelo critério half-up: meio centavo exato sobe. A pessoa 2
// recebe o restante, em vez de ser arredondada separadamente, para que os dois
// valores sempre reconciliem exatamente com a conta original.
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

	// big.Rat é mutável: Set cria cópias para que Add, Quo e Mul não alterem os
	// valores pertencentes ao chamador.
	c1 := new(big.Rat).Set(consumption1)
	c2 := new(big.Rat).Set(consumption2)
	total := new(big.Rat).Add(c1, c2)
	if total.Sign() == 0 {
		return Result{}, errors.New("total consumption must be greater than zero")
	}

	// As participações são frações exatas consumo/total. O valor monetário
	// também vira um racional antes da multiplicação, adiando o único
	// arredondamento necessário até a conversão final para centavos inteiros.
	share1 := new(big.Rat).Quo(new(big.Rat).Set(c1), total)
	share2 := new(big.Rat).Quo(new(big.Rat).Set(c2), total)
	exactAmount1 := new(big.Rat).Mul(new(big.Rat).SetInt64(totalAmountCents), share1)
	amount1 := roundHalfUp(exactAmount1)

	return Result{
		TotalConsumption: new(big.Rat).Set(total),
		Share1:           new(big.Rat).Set(share1),
		Share2:           new(big.Rat).Set(share2),
		Amount1Cents:     amount1,
		// Calcular por diferença é a etapa de reconciliação: evita que dois
		// arredondamentos independentes criem ou eliminem um centavo.
		Amount2Cents: totalAmountCents - amount1,
	}, nil
}

// roundHalfUp converte um racional não negativo para o inteiro mais próximo.
// QuoRem separa numerador/denominador em parte inteira e resto usando big.Int,
// portanto a comparação continua exata e não sofre erros de ponto flutuante.
func roundHalfUp(value *big.Rat) int64 {
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(value.Num(), value.Denom(), remainder)
	// Comparar 2*resto com o denominador equivale a perguntar se a parte
	// fracionária é pelo menos 1/2. A igualdade implementa o "half-up".
	if new(big.Int).Lsh(remainder, 1).Cmp(value.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.Int64()
}
