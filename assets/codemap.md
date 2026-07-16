# assets/

## Responsabilidade

Mantém os recursos visuais distribuídos dentro do executável gráfico, incluindo
o ícone do Rateio Luz.

## Projeto

O PNG é incorporado em tempo de compilação por `go:embed`. Os bytes resultantes
são expostos como `fyne.Resource` estático, evitando dependência de caminhos ou
arquivos externos em tempo de execução.

## Fluxo

Durante a compilação, `rateio-luz.png` é convertido em bytes no binário. Na
inicialização do pacote, esses bytes originam `Icon`, pronto para ser consumido
pela aplicação Fyne.

## Integrações

- O executável gráfico em `main.go` fornece `Icon` à aplicação Fyne.
- `embed` realiza a inclusão do arquivo durante o build.
- `fyne.io/fyne/v2` adapta o conteúdo incorporado à abstração de recurso usada
  pela janela e pela aplicação.
