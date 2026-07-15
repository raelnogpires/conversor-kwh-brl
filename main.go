package main

import (
	"rateio-luz/assets"
	appgui "rateio-luz/internal/gui"

	"fyne.io/fyne/v2/app"
)

func main() {
	application := app.NewWithID(appgui.AppID)
	application.Settings().SetTheme(appgui.NewTheme())
	application.SetIcon(assets.Icon)
	appgui.NewWindow(application).ShowAndRun()
}
