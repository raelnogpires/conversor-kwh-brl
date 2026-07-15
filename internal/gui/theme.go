package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	// MarineBlue and OffWhite are the two main Rateio Luz brand colors.
	MarineBlue = color.NRGBA{R: 0x12, G: 0x3B, B: 0x5D, A: 0xFF}
	OffWhite   = color.NRGBA{R: 0xF8, G: 0xF5, B: 0xEC, A: 0xFF}
	accent     = color.NRGBA{R: 0x35, G: 0xB7, B: 0xC9, A: 0xFF}
)

type rateioTheme struct {
	base fyne.Theme
}

// NewTheme returns the light, high-contrast Rateio Luz theme.
func NewTheme() fyne.Theme {
	return &rateioTheme{base: theme.DefaultTheme()}
}

func (t *rateioTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return OffWhite
	case theme.ColorNamePrimary:
		return MarineBlue
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0x16, G: 0x2A, B: 0x3A, A: 0xFF}
	case theme.ColorNameForegroundOnPrimary:
		return OffWhite
	case theme.ColorNameButton:
		return color.NRGBA{R: 0xE5, G: 0xEC, B: 0xF0, A: 0xFF}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0x8C, G: 0xA0, B: 0xAE, A: 0xFF}
	case theme.ColorNameFocus, theme.ColorNameSelection:
		return accent
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x1F, G: 0x7A, B: 0x68, A: 0xFF}
	default:
		return t.base.Color(name, theme.VariantLight)
	}
}

func (t *rateioTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *rateioTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *rateioTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
