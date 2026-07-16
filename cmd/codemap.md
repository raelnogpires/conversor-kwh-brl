# cmd/

## Responsabilidade

Contém o executável de linha de comando do Rateio Luz. Sua função é coordenar
a conversa com o usuário, sem incorporar regras de validação ou de rateio.

## Projeto

`main` conecta o processo a `os.Stdin`, `os.Stdout` e `os.Stderr`. A função
`run` recebe `io.Reader` e `io.Writer`, mantendo a orquestração independente de
um terminal concreto e permitindo reutilização com outros fluxos de E/S.

## Fluxo

1. Exibe o título e solicita três valores textuais por meio de `internal/data`.
2. Envia as entradas brutas para conversão e validação.
3. Encaminha os valores tipados ao cálculo do rateio.
4. Formata consumo, proporções e valores monetários para exibição.
5. Propaga falhas até `main`, que as escreve em stderr e encerra com erro.

## Integrações

- `internal/data`: interação textual e leitura das linhas.
- `internal/validation`: interpretação e validação da entrada bruta.
- `internal/calculator`: regra de cálculo do rateio.
- `internal/presentation`: representação de kWh, percentual e BRL.
- Biblioteca padrão: abstrações de E/S e fluxos do processo.
