# Conversor kWh-BRL

Aplicação de terminal em Go que divide o valor de uma conta de energia entre dois consumidores, proporcionalmente ao consumo individual em kWh. Ela resolve o caso de dois espaços com medidores próprios que recebem uma única conta da concessionária.

## Entradas e saídas

O programa solicita:

1. consumo do consumidor 1 em kWh;
2. consumo do consumidor 2 em kWh;
3. valor total da conta em reais (BRL).

Ele apresenta o consumo total, a proporção exata e a porcentagem formatada de cada consumidor e os dois valores a pagar. Exemplo:

```text
Consumo do consumidor 1 (kWh): 105,5
Consumo do consumidor 2 (kWh): 67,2
Valor total da conta (R$): 184,72

Resultado
Consumo total: 172,7 kWh
Proporção consumidor 1: 1055/1727 (61,09%)
Proporção consumidor 2: 672/1727 (38,91%)
Consumidor 1 paga: R$ 112,84
Consumidor 2 paga: R$ 71,88
```

## Regra de cálculo

Os cálculos usam estas fórmulas:

```text
totalConsumption = consumption1 + consumption2
share1 = consumption1 / totalConsumption
share2 = consumption2 / totalConsumption
amount1 = arredondar(totalAmountCents × share1)
amount2 = totalAmountCents - amount1
```

As proporções são obtidas por divisão e permanecem como valores racionais exatos. Por isso, `share1 + share2` é exatamente 1. O valor do consumidor 1 é arredondado para o centavo mais próximo; um resultado exatamente no meio do centavo é arredondado para cima (*half-up*). O consumidor 2 recebe o restante. Assim, os pagamentos sempre somam exatamente o total da conta.

Essa política é determinística e simples, mas atribui ao consumidor 2 qualquer diferença causada pelo arredondamento. Outras políticas poderiam alternar quem recebe o restante ou distribuí-lo por outro critério; esta versão prioriza a conservação exata do total e uma regra previsível.

## Representação numérica

Dinheiro é interpretado sem `float64` e armazenado como `int64` em centavos. Isso evita resultados binários aproximados como `184.719999...` e torna exata a soma dos pagamentos. O limite é `math.MaxInt64` centavos.

Uma biblioteca decimal de ponto fixo também poderia representar dinheiro com segurança e oferecer mais escalas e operações, mas exigiria uma dependência externa ou uma implementação mais complexa. Centavos inteiros são suficientes para BRL nesta aplicação e mantêm o projeto apenas com a biblioteca padrão.

Consumo não é dinheiro e pode ter mais casas decimais. Ele usa `math/big.Rat`, que conserva exatamente os decimais informados e permite proporções exatas, sem as aproximações de `float64`.

## Validação e limitações da entrada

- vírgula ou ponto é aceito como único separador decimal;
- não são aceitos separadores de milhares, formatos mistos, notação científica, `NaN` ou infinito;
- cada parte ao redor do separador deve conter ao menos um dígito (`.5` e `5,` não são aceitos);
- valores monetários aceitam no máximo duas casas decimais;
- consumos e conta negativos são rejeitados;
- zero é válido para um consumidor e para o valor da conta;
- os dois consumos iguais a zero são inválidos, pois a proporção ficaria indefinida;
- espaços externos são ignorados, mas entradas vazias e textos malformados são rejeitados.

Um único ponto ou uma única vírgula sempre significa separador decimal. Para evitar ambiguidade, não digite agrupamento de milhares: use `1234,56`, e não `1.234,56`.

## Executar, compilar e testar

É necessário Go 1.25.1 ou versão compatível com o `go.mod`. Na raiz do projeto:

```bash
go run ./cmd
go build -o conversor-kwh-brl ./cmd
go test ./...
```

No Windows, o build pode ser gerado com:

```powershell
go build -o conversor-kwh-brl.exe ./cmd
```

## Arquitetura

```text
cmd/main.go                        orquestra entrada, validação, cálculo e apresentação
internal/data/input.go             lê somente as três strings do terminal
internal/validation/validation.go  converte e valida consumo e dinheiro
internal/calculator/calculator.go  contém a regra de domínio, sem terminal ou formatação
```

As validações de domínio também existem no calculador, de modo que ele continua seguro se for reutilizado por outra interface. A apresentação em português e a formatação em BRL ficam fora da regra de cálculo.

## Evoluções possíveis

- interface gráfica para Linux e Windows reutilizando os pacotes internos;
- divisão entre mais de dois consumidores;
- política configurável para distribuição do restante;
- histórico e exportação de comprovantes.

## Licença

Distribuído sob a licença MIT. Consulte [LICENSE.md](LICENSE.md).
Ela permite usar, modificar, redistribuir, sublicenciar e explorar
comercialmente o projeto, desde que o aviso de copyright e o texto da licença
sejam preservados nas cópias ou partes substanciais do software.
