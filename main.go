package main

import (
	"rateio-luz/assets"
	appgui "rateio-luz/internal/gui"

	"fyne.io/fyne/v2/app"
)

func main() {
	// Este é o ponto de entrada do executável gráfico. O ID estável permite ao
	// Fyne associar preferências e demais dados persistentes à aplicação.
	application := app.NewWithID(appgui.AppID)
	// Tema e ícone são definidos antes da janela para que ela já seja criada
	// com a identidade visual correta em todas as plataformas suportadas.
	application.Settings().SetTheme(appgui.NewTheme())
	application.SetIcon(assets.Icon)
	// A camada GUI monta a janela e assume o loop de eventos até seu fechamento.
	appgui.NewWindow(application).ShowAndRun()
}
