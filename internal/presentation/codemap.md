# internal/presentation/

## Responsabilidade

Converter valores tipados do domínio em textos destinados ao usuário: reais,
porcentagens, decimais com convenção brasileira e consumos acompanhados de
`kWh`.

## Design

- Dinheiro chega em centavos inteiros e é separado com divisão e resto, sem
  usar ponto flutuante.
- Porcentagens e decimais operam sobre `*big.Rat`; arredondamento, quando
  necessário para exibição, fica restrito a esta camada.
- `Decimal` detecta representações finitas fatorando o denominador em 2 e 5.
  Se houver outro fator, usa `RatString` para não perder exatidão.
- `decimalComma` concentra a adaptação do ponto produzido por `math/big` para a
  vírgula esperada na apresentação brasileira.

## Fluxo

- `BRL` separa reais e centavos e produz sempre duas casas monetárias.
- `Percentage` multiplica a participação por 100, limita a saída a duas casas e
  troca o separador decimal. Duas participações formatadas independentemente
  podem somar visualmente 99,99% ou 100,01% por causa do arredondamento.
- `Decimal` analisa o denominador, calcula as casas necessárias, remove zeros
  finais sem valor significativo e localiza o separador.
- `KWh` reutiliza `Decimal` e acrescenta a unidade.

## Integrações

É consumido pelas interfaces de saída para exibir valores vindos de
`internal/calculator` e, quando necessário, de `internal/validation`. Não faz
validação de entrada nem altera regras do cálculo; somente representa valores
já tipados.
