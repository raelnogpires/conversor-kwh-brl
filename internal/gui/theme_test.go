package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestRateioThemeKeepsBrandColorsInAnySystemVariant(t *testing.T) {
	themeUnderTest := NewTheme()
	for _, variant := range []struct {
		name  string
		value fyne.ThemeVariant
	}{
		{name: "light", value: theme.VariantLight},
		{name: "dark", value: theme.VariantDark},
	} {
		t.Run(variant.name, func(t *testing.T) {
			actualBackground := themeUnderTest.Color(theme.ColorNameBackground, variant.value)
			if actualBackground != OffWhite {
				t.Errorf("background = %v, want %v", actualBackground, OffWhite)
			}
			actualPrimary := themeUnderTest.Color(theme.ColorNamePrimary, variant.value)
			if actualPrimary != MarineBlue {
				t.Errorf("primary = %v, want %v", actualPrimary, MarineBlue)
			}
		})
	}
}
