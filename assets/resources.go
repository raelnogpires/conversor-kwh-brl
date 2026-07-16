// Package assets disponibiliza recursos incorporados ao executável Rateio Luz.
package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

// A diretiva go:embed inclui o PNG na compilação, dispensando a distribuição
// de um arquivo de imagem separado ao lado do programa.
//
//go:embed rateio-luz.png
var iconPNG []byte

// Icon adapta os bytes incorporados para o recurso estático esperado pelo
// Fyne; o nome também identifica o conteúdo para consumidores da biblioteca.
var Icon = fyne.NewStaticResource("rateio-luz.png", iconPNG)
