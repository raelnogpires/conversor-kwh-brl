# internal/validation/

## Responsabilidade

Transformar os textos digitados pelo usuário em tipos seguros para o domínio e
rejeitar entradas inválidas ou relações impossíveis. Centraliza as regras de
sintaxe decimal, consumo, dinheiro e consumo total positivo.

## Design

- `Input` é o contrato validado: consumos exatos em `*big.Rat` e conta em
  centavos `int64`.
- Decimais aceitam vírgula ou ponto, mas não separador de milhares. A
  normalização produz ponto decimal antes de usar `math/big`.
- Consumo não passa por `float64`; `big.Rat` preserva exatamente todos os
  dígitos informados.
- Dinheiro aceita até duas casas e é convertido por concatenação de dígitos.
  `big.Int` permite detectar overflow antes da redução para `int64`.
- Os parsers específicos aplicam regras de domínio; `normalizeDecimal` cuida
  somente da gramática e informa o sinal separadamente.

## Fluxo

1. `ParseAndValidate` encaminha os dois consumos a `ParseConsumption` e o total
   a `ParseMoneyCents`, acrescentando contexto aos erros.
2. Cada parser chama `normalizeDecimal`, que remove espaços externos, valida os
   caracteres e padroniza o separador.
3. Consumos viram racionais; o valor monetário vira uma sequência inteira de
   centavos e é checado contra o intervalo de `int64`.
4. A validação conjunta exige que pelo menos um consumo seja positivo.
5. O módulo devolve `Input` somente quando todas as invariantes estão atendidas.

## Integrações

É a fronteira entre textos das interfaces de entrada e os tipos esperados por
`internal/calculator`. Devolve mensagens em português prontas para serem
propagadas pela interface, mas não realiza cálculos proporcionais nem formata
resultados.
