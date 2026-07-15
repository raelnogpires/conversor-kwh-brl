package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	// MarineBlue and OffWhite are the two main Rateio Luz brand colors.
	MarineBlue  = color.NRGBA{R: 0x12, G: 0x3B, B: 0x5D, A: 0xFF}
	OffWhite    = color.NRGBA{R: 0xF8, G: 0xF5, B: 0xEC, A: 0xFF}
	accent      = color.NRGBA{R: 0x35, G: 0xB7, B: 0xC9, A: 0xFF}
	deepNavy    = color.NRGBA{R: 0x08, G: 0x25, B: 0x38, A: 0xFF}
	ink         = color.NRGBA{R: 0x13, G: 0x29, B: 0x38, A: 0xFF}
	mutedInk    = color.NRGBA{R: 0x58, G: 0x6B, B: 0x76, A: 0xFF}
	surface     = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFD, A: 0xFF}
	softBlue    = color.NRGBA{R: 0xE8, G: 0xF3, B: 0xF5, A: 0xFF}
	line        = color.NRGBA{R: 0xD6, G: 0xE0, B: 0xE3, A: 0xFF}
	errorWash   = color.NRGBA{R: 0xFF, G: 0xEE, B: 0xEC, A: 0xFF}
	successWash = color.NRGBA{R: 0xE8, G: 0xF5, B: 0xF0, A: 0xFF}
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
		return ink
	case theme.ColorNameForegroundOnPrimary:
		return OffWhite
	case theme.ColorNameButton:
		return color.NRGBA{R: 0xE3, G: 0xEC, B: 0xEF, A: 0xFF}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0xEA, G: 0xEE, B: 0xED, A: 0xFF}
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return mutedInk
	case theme.ColorNameInputBackground:
		return surface
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0x91, G: 0xA5, B: 0xAF, A: 0xFF}
	case theme.ColorNameFocus, theme.ColorNameSelection:
		return accent
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x35, G: 0xB7, B: 0xC9, A: 0x22}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0x12, G: 0x3B, B: 0x5D, A: 0x28}
	case theme.ColorNameHeaderBackground, theme.ColorNameMenuBackground:
		return surface
	case theme.ColorNameSeparator:
		return line
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x08, G: 0x25, B: 0x38, A: 0x24}
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
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 17
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameInnerPadding:
		return 7
	case theme.SizeNameInputBorder:
		return 1.5
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius:
		return 10
	case theme.SizeNameSeparatorThickness:
		return 1
	}
	return t.base.Size(name)
}
