# internal/calculator/

## Responsabilidade

Implementar a regra de domínio para dividir uma conta entre dois consumidores
na proporção de seus consumos. O módulo também garante a reconciliação: as duas
parcelas em centavos sempre somam exatamente o total recebido.

## Design

- `Result` separa medidas exatas (`*big.Rat`) de dinheiro já quantizado em
  centavos (`int64`).
- Os racionais recebidos são copiados porque `big.Rat` é mutável; o cálculo e o
  resultado não alteram nem compartilham os objetos do chamador.
- Apenas a primeira parcela é arredondada, pelo critério half-up exato com
  `big.Int`. A segunda é o total menos a primeira, evitando divergência de um
  centavo causada por dois arredondamentos independentes.
- O contrato rejeita consumos nulos ou negativos, conta negativa e soma de
  consumos igual a zero.

## Fluxo

1. `Calculate` valida as pré-condições.
2. Copia e soma os consumos, então calcula cada participação como
   `consumo / consumo total`.
3. Multiplica a participação 1 pelo total em centavos sem perder precisão.
4. `roundHalfUp` converte esse racional em centavos inteiros.
5. Reconcilia a participação 2 por diferença e devolve todos os derivados em
   `Result`.

## Integrações

Recebe normalmente os consumos e centavos produzidos por
`internal/validation`. Seu `Result` é consumido pelas interfaces e pode ser
convertido em textos localizados por `internal/presentation`. O pacote não
conhece detalhes de terminal nem regras de formatação.
