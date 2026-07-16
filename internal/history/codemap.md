# internal/history/

## Responsabilidade

Este pacote é a camada de persistência do histórico de rateios. Ele armazena e
recupera resultados em um arquivo CSV e permite excluir um item pela posição em
que aparece no histórico. O pacote não calcula valores nem decide sua
apresentação: recebe da interface os campos já formatados e os conserva como
texto.

## Design

- `Entry` é o modelo persistido. A data é tipada como `time.Time` e serializada
  em RFC3339; consumos, percentuais e valores monetários permanecem em `string`
  para preservar a representação exibida ao usuário.
- `Store` encapsula somente o caminho do CSV. `NewStore` permite que a aplicação
  escolha esse caminho e que testes usem arquivos isolados.
- O cabeçalho de nove colunas é o contrato do arquivo e segue a mesma ordem de
  `Entry.record`: Data, Consumo 1, Consumo 2, Valor total, Consumo total,
  Percentual 1, Percentual 2, Valor 1 e Valor 2.
- Leitura e exclusão rejeitam cabeçalhos diferentes, linhas com outra quantidade
  de campos e datas que não sejam RFC3339. Os erros adicionam contexto com `%w`,
  e índices inexistentes preservam `ErrInvalidIndex` para uso com `errors.Is`.
- A exclusão reescreve o conteúdo em um arquivo temporário no mesmo diretório:
  mantém as permissões originais, sincroniza e fecha o novo conteúdo antes de
  publicá-lo com `os.Rename`. As garantias de atomicidade dependem do sistema de
  arquivos.

## Fluxo

1. `Save` cria o diretório e abre o CSV para anexação. Se o arquivo estiver
   vazio, grava primeiro o cabeçalho; depois serializa e acrescenta o `Entry`.
2. `List` trata arquivo inexistente como histórico vazio, valida o cabeçalho,
   lê todas as linhas e converte cada uma para `Entry`, mantendo a ordem física.
3. `DeleteAt` valida o índice, lê e valida o arquivo inteiro, remove a linha em
   memória e grava cabeçalho e linhas restantes em um temporário. Após `Sync` e
   fechamento, `Rename` substitui o histórico; falhas anteriores deixam o
   arquivo original intacto e o temporário é removido.

## Integrações

- `internal/gui` constrói o `Store`, cria snapshots como `history.Entry` e usa
  `Save`, `List` e `DeleteAt` por meio de sua interface local de armazenamento.
- `encoding/csv` cuida de escaping e leitura das colunas; `os` e `path/filepath`
  cuidam de diretórios, arquivos, permissões e substituição; `time` define a
  codificação da data.
- O pacote não depende da GUI nem das regras de cálculo, mantendo a persistência
  reutilizável e testável separadamente.

## Limitações

- `Store` não aplica locking entre processos. Salvar ou excluir simultaneamente
  em duas instâncias pode perder alterações.
- A exclusão recebe uma posição da lista, não um identificador estável; uma
  alteração externa entre `List` e `DeleteAt` pode mudar o registro apontado.
- `Save` não valida um arquivo não vazio antes de anexar. Corrupção preexistente
  é detectada por `List` ou `DeleteAt`, não no momento da gravação.
