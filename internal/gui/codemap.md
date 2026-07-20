# internal/gui/

## Responsabilidade

Este pacote implementa a camada de apresentação desktop do Rateio Luz com
Fyne. Ele monta a janela, mantém o estado dos widgets, coordena as interações do
usuário e transforma resultados e falhas dos demais pacotes em feedback visual.
Também oferece o tema da aplicação, mas não contém as regras de validação,
cálculo, formatação monetária ou persistência do histórico.

## Design

- `screen` reúne as referências dos widgets e o estado transitório de uma
  janela: entradas, resultado atual, feedback, abas e itens carregados do
  histórico. Seus métodos funcionam como callbacks e coordenadores da tela.
- `historyStore` é uma interface mínima injetável (`Save`, `List` e `DeleteAt`).
  A GUI pode usar o store real em produção e substitutos controlados nos testes,
  sem conhecer o formato CSV.
- `now func() time.Time` aplica a mesma ideia ao relógio usado no snapshot.
- Os métodos `build...` constroem grupos maiores de widgets; helpers como
  `newSurface`, `fieldGroup`, `resultTile` e `historyResidentTile` mantêm padrões
  visuais sem duplicar layout.
- `maxWidthLayout` centraliza e limita a largura do conteúdo. A rolagem continua
  responsável por acomodar janelas com pouca altura.
- Resultado e mensagens são criados uma vez e alternados com `Show`/`Hide`.
  Qualquer edição invalida o snapshot, evitando salvar dados que já não
  correspondem às entradas visíveis.
- `rateioTheme` personaliza apenas cores e tamanhos da marca. Fontes, ícones e
  tokens desconhecidos são delegados ao tema padrão claro do Fyne.

## Fluxo de dados e controle

1. `NewWindow` cria uma `screen`; em produção, `newScreen` resolve o caminho no
   diretório de configuração do usuário e injeta `history.NewStore`.
2. `newScreenWithStore` cria entradas e feedback, monta os cartões e abas,
   registra callbacks e instala o conteúdo final na janela.
3. Ao editar um campo, a tela remove erros e invalida resultado/snapshot antigos.
   Enter avança o foco; no último campo, chama `calculate`.
4. `calculate` envia os textos a `validation.ParseAndValidate`, passa os valores
   normalizados a `calculator.Calculate` e confere se as parcelas recompõem o
   total. Erros são traduzidos e associados ao campo pertinente.
5. Em sucesso, `presentation` formata kWh, percentuais e reais. Esses textos
   atualizam os labels e formam um `history.Entry` transitório; o usuário ainda
   precisa escolher **Salvar no histórico**.
6. `saveSnapshot` copia o snapshot e delega a gravação ao store em background.
   Enquanto isso, as ações de histórico ficam desabilitadas e uma atividade é
   exibida; o resultado volta à GUI por `fyne.Do`. Falhas mantêm a tentativa
   disponível; sucesso desabilita o botão para evitar duplicação imediata.
7. Ao selecionar ou atualizar a aba Histórico, `loadHistory` relê o store em
   background e publica o resultado na GUI por `fyne.Do`. A lista visual é
   paginada em blocos de 25 cartões. Cada botão **Excluir** captura o índice da
   entrada na ordem carregada, chama `DeleteAt` em background e, após a remoção,
   atualiza apenas a página corrente em memória.
8. Se uma carga for solicitada enquanto outra operação de histórico está ativa,
   a tela registra a solicitação e inicia `List` assim que a operação atual
   termina. Falhas de leitura preservam o cartão de erro em vez de apresentá-las
   como histórico vazio.

As operações do store continuam síncronas dentro do próprio contrato de
persistência, mas são executadas fora do callback da GUI. A tela serializa as
operações por estado ocupado, exibe atividade e limita a quantidade de cartões
materializados por página. O CSV ainda é lido integralmente por `List`; a
paginação reduz o custo de construção e atualização visual, não o custo de
leitura do arquivo.

## Integrações

- **Fyne (`fyne.io/fyne/v2`)**: janela, canvas, layouts, widgets, recursos,
  navegação por abas e contrato de tema.
- **`assets`**: ícone exibido no cabeçalho.
- **`internal/validation`**: parsing dos textos e regras de validade da entrada.
- **`internal/calculator`**: cálculo proporcional e reconciliação em centavos.
- **`internal/presentation`**: representação localizada de consumo, percentual e BRL.
- **`internal/history`**: modelo `Entry` e store persistente usado pela aplicação.
- **Sistema operacional**: diretório de configuração do usuário, com fallback
  para `~/.config` e, em último caso, o diretório temporário.
