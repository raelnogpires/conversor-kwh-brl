# Conversor kWh-BRL

Aplicação para dividir proporcionalmente o valor de uma conta de energia elétrica entre dois consumidores, com base no consumo individual registrado em kWh.

## Problema

A residência possui dois espaços independentes, mas ambos compartilham a mesma conta de energia elétrica.

Cada espaço possui um medidor que informa seu consumo individual em kWh. Entretanto, a concessionária emite apenas uma conta com o valor total em reais.

O programa recebe:

* consumo do consumidor 1 em kWh;
* consumo do consumidor 2 em kWh;
* valor total da conta em reais.

Como resultado, informa quanto cada consumidor deve pagar proporcionalmente ao seu consumo.

## Regra de cálculo

Primeiro, o programa calcula o consumo total:

```text
consumo total = consumo 1 + consumo 2
```

Depois, determina a proporção de consumo de cada consumidor:

```text
proporção 1 = consumo 1 / consumo total
proporção 2 = consumo 2 / consumo total
```

Por fim, aplica essas proporções ao valor total da conta:

```text
valor 1 = valor total da conta × proporção 1
valor 2 = valor total da conta × proporção 2
```

A soma dos dois valores deve corresponder ao valor total da conta:

```text
valor 1 + valor 2 = valor total da conta
```

## Exemplo

Dados de entrada:

```text
Consumidor 1: 105,5 kWh
Consumidor 2: 67,2 kWh
Conta total: R$ 184,72
```

Consumo total:

```text
105,5 + 67,2 = 172,7 kWh
```

Resultado aproximado:

```text
Consumidor 1: R$ 112,85
Consumidor 2: R$ 71,87
```

## Validações esperadas

O programa deve rejeitar:

* consumos negativos;
* valor da conta negativo;
* consumo total igual a zero;
* entradas vazias;
* textos que não possam ser convertidos em números.

Consumo individual igual a zero é permitido, desde que o outro consumidor tenha consumo maior que zero.

## Estrutura planejada

```text
conversor-kwh-brl/
├── cmd/
│   └── app/
├── internal/
│   ├── calculator/
│   └── validation/
├── assets/
├── go.mod
├── README.md
└── LICENSE.md
```

Responsabilidades:

* `internal/calculator`: regra de divisão proporcional da conta;
* `internal/validation`: validação dos valores recebidos;
* `cmd/app`: inicialização do programa e montagem da interface;
* `assets`: ícones e outros recursos visuais.

Durante o primeiro estágio, o projeto pode utilizar uma estrutura reduzida:

```text
conversor-kwh-brl/
├── main.go
├── go.mod
├── README.md
└── LICENSE.md
```

A separação em pacotes será feita depois que a regra principal estiver implementada e testada.

## Requisitos

* Go 1.24 ou superior.

## Executando o projeto

Na raiz do projeto:

```bash
go run .
```

Para gerar um executável:

```bash
go build -o conversor-kwh-brl .
```

No Windows:

```powershell
go build -o conversor-kwh-brl.exe .
```

## Escopo inicial

A primeira versão será executada no terminal e terá como objetivo validar a regra de negócio.

A interface gráfica será adicionada posteriormente, sem alterar a lógica central do cálculo.

## Possíveis evoluções

* interface gráfica para Linux e Windows;
* suporte a mais de dois consumidores;
* histórico de cálculos;
* exportação de comprovantes;
* configuração de arredondamento;
* testes automatizados;
* geração de executáveis para diferentes sistemas operacionais.

## Licença

Este projeto está licenciado sob a licença MIT. Consulte o arquivo [LICENSE.md](LICENSE.md) para mais informações.
