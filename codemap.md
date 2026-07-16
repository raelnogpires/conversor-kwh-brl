# Mapa do Repositório: Rateio Luz

## Responsabilidade do projeto

O Rateio Luz divide uma conta de energia entre dois moradores na proporção do
consumo de cada um. O mesmo núcleo atende dois executáveis: uma aplicação
desktop feita com Fyne e uma CLI para terminal. A GUI também pode guardar os
resultados escolhidos pelo usuário em um arquivo CSV local.

O projeto prioriza exatidão: consumos e proporções usam `math/big.Rat`, enquanto
dinheiro usa `int64` em centavos. Nenhuma regra monetária depende de `float32` ou
`float64`.

## Pontos de entrada

- `main.go`: cria a aplicação Fyne, instala tema e ícone e abre a janela GUI.
- `cmd/main.go`: executa o fluxo terminal; sua função `run` recebe `io.Reader` e
  `io.Writer` para que o fluxo seja testado sem depender do terminal real.
- `go.mod`: declara o módulo `rateio-luz`, a versão mínima do Go e o Fyne.
- `FyneApp.toml`: contém nome, ID, versão, ícone e metadados de empacotamento.
- `README.md`: documenta uso, desenvolvimento, precisão e distribuição.

## Arquitetura

```text
                          texto digitado
                                |
              +-----------------+-----------------+
              |                                   |
       cmd + internal/data                 internal/gui
              |                                   |
              +--------> internal/validation <----+
                                |
                                v
                      internal/calculator
                                |
                                v
                     internal/presentation
                                |
              +-----------------+-----------------+
              |                                   |
          saída da CLI                      widgets da GUI
                                                    |
                                          internal/history
                                               CSV local
```

As dependências apontam das interfaces para os pacotes internos. O cálculo não
importa GUI, terminal, arquivos nem formatação, o que mantém a regra de domínio
isolada e simples de testar.

## Fluxo de um rateio

1. A CLI lê três linhas por `data.ReadInput`; a GUI lê o texto dos seus campos.
2. `validation.ParseAndValidate` normaliza os decimais, rejeita dados inválidos,
   cria dois `*big.Rat` e converte a conta para centavos inteiros.
3. `calculator.Calculate` soma os consumos, encontra as duas proporções e
   calcula o pagamento do morador 1 com arredondamento *half-up*.
4. O pagamento do morador 2 é o total menos o primeiro pagamento. Essa operação
   reconcilia os dois resultados exatamente com a conta original.
5. `presentation` converte os valores exatos para textos em BRL, percentual e
   kWh, usando vírgula decimal.
6. A interface exibe os textos. Na GUI, um `history.Entry` é preparado e só é
   gravado quando o usuário aciona **Salvar no histórico**.

## Invariantes importantes

- os consumos individuais não podem ser negativos;
- pelo menos um consumo deve ser maior que zero;
- a conta não pode ser negativa e aceita no máximo duas casas decimais;
- `Amount1Cents + Amount2Cents` deve ser igual a `TotalAmountCents`;
- o cabeçalho e a quantidade de campos do CSV devem coincidir com o esquema de
  `history.Entry`;
- editar qualquer campo da GUI invalida o resultado e o retrato anteriores.

## Mapa de diretórios

| Diretório | Responsabilidade | Mapa detalhado |
| --- | --- | --- |
| `assets/` | Incorpora o ícone ao executável. | [`assets/codemap.md`](assets/codemap.md) |
| `cmd/` | Ponto de entrada e orquestração da CLI. | [`cmd/codemap.md`](cmd/codemap.md) |
| `internal/` | Pacotes privados compartilhados pelas interfaces. | [`internal/codemap.md`](internal/codemap.md) |
| `internal/calculator/` | Regra proporcional, arredondamento e reconciliação. | [`internal/calculator/codemap.md`](internal/calculator/codemap.md) |
| `internal/data/` | Leitura de entradas brutas do terminal. | [`internal/data/codemap.md`](internal/data/codemap.md) |
| `internal/gui/` | Janela, widgets, eventos, tema e feedback da GUI. | [`internal/gui/codemap.md`](internal/gui/codemap.md) |
| `internal/history/` | Persistência, leitura e exclusão no CSV local. | [`internal/history/codemap.md`](internal/history/codemap.md) |
| `internal/presentation/` | Formatação de valores para o usuário. | [`internal/presentation/codemap.md`](internal/presentation/codemap.md) |
| `internal/validation/` | Parsing e validação da entrada textual. | [`internal/validation/codemap.md`](internal/validation/codemap.md) |

Diretórios `dist/` e `fyne-cross/` contêm artefatos gerados e não fazem parte do
código-fonte que deve ser estudado ou editado.

## Estratégia de testes

Os arquivos `*_test.go` ficam ao lado do pacote testado. Testes de domínio
cobrem limites, arredondamento e entradas inválidas; testes de persistência usam
diretórios temporários; testes da CLI injetam buffers; e testes da GUI usam o
driver de teste do Fyne. A suíte completa deve ser executada com:

```bash
go test -tags ci ./...
```

## Ordem de leitura sugerida

Comece por `cmd/main.go` para enxergar o fluxo completo em poucas linhas. Siga
por `validation`, `calculator` e `presentation`; depois leia `history`. Por fim,
leia `main.go`, `gui/app.go` e `gui/theme.go`, onde a maior parte do volume vem
da composição visual e dos callbacks do Fyne.
