# internal/data/

## Responsabilidade

Implementa a coleta de dados brutos da interface de terminal. Exibe as
perguntas na ordem esperada e devolve as respostas ainda como texto.

## Projeto

A API opera sobre `io.Reader` e `io.Writer`, não sobre os fluxos globais do
processo. Um `bufio.Scanner` separa a entrada por linhas, enquanto uma tabela de
perguntas e destinos mantém uniforme o tratamento dos três campos.

## Fluxo

Para cada campo, escreve a pergunta, lê a próxima linha e armazena seu conteúdo
sem conversão. Erros de escrita ou leitura são contextualizados; o fim da
entrada antes da terceira resposta é informado como `io.ErrUnexpectedEOF`.

## Integrações

- É chamada pelo executável em `cmd/`, que fornece os fluxos de entrada e saída.
- Entrega strings à camada de validação indiretamente, por meio do orquestrador
  da CLI; não conhece regras de domínio nem formatos numéricos.
- Usa `bufio`, `fmt` e `io` da biblioteca padrão para leitura e escrita.
