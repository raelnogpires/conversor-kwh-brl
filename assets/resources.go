// Package assets contains resources embedded in the Rateio Luz executable.
package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed rateio-luz.png
var iconPNG []byte

// Icon is the application and window icon bundled into the executable.
var Icon = fyne.NewStaticResource("rateio-luz.png", iconPNG)
