package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// PurpleTheme is a custom purple-based theme for admin:admin
type PurpleTheme struct{}

var _ fyne.Theme = (*PurpleTheme)(nil)

// Purple color palette
var (
	PurplePrimary    = color.NRGBA{R: 138, G: 43, B: 226, A: 255}  // Blue Violet
	PurpleSecondary  = color.NRGBA{R: 147, G: 112, B: 219, A: 255} // Medium Purple
	PurpleLight      = color.NRGBA{R: 216, G: 191, B: 216, A: 255} // Thistle
	PurpleDark       = color.NRGBA{R: 75, G: 0, B: 130, A: 255}    // Indigo
	PurpleAccent     = color.NRGBA{R: 186, G: 85, B: 211, A: 255}  // Medium Orchid
	PurpleBackground = color.NRGBA{R: 25, G: 20, B: 35, A: 255}    // Dark purple background
	PurpleSurface    = color.NRGBA{R: 40, G: 30, B: 55, A: 255}    // Slightly lighter surface
	PurpleText       = color.NRGBA{R: 240, G: 235, B: 245, A: 255} // Light purple-white text
	PurpleDisabled   = color.NRGBA{R: 100, G: 90, B: 110, A: 255}  // Muted purple
)

func (t *PurpleTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return PurplePrimary
	case theme.ColorNameBackground:
		return PurpleBackground
	case theme.ColorNameButton:
		return PurplePrimary
	case theme.ColorNameDisabled:
		return PurpleDisabled
	case theme.ColorNameDisabledButton:
		return PurpleDisabled
	case theme.ColorNameForeground:
		return PurpleText
	case theme.ColorNameHover:
		return PurpleSecondary
	case theme.ColorNameInputBackground:
		return PurpleSurface
	case theme.ColorNameInputBorder:
		return PurpleSecondary
	case theme.ColorNamePlaceHolder:
		return PurpleDisabled
	case theme.ColorNamePressed:
		return PurpleDark
	case theme.ColorNameScrollBar:
		return PurpleSecondary
	case theme.ColorNameSeparator:
		return PurpleSecondary
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameSelection:
		return PurpleAccent
	case theme.ColorNameFocus:
		return PurpleAccent
	case theme.ColorNameMenuBackground:
		return PurpleSurface
	case theme.ColorNameOverlayBackground:
		return PurpleSurface
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 76, G: 175, B: 80, A: 255}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 255, G: 193, B: 7, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 244, G: 67, B: 54, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *PurpleTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *PurpleTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *PurpleTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 24
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameText:
		return 14
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// NewPurpleTheme returns the purple theme instance
func NewPurpleTheme() fyne.Theme {
	return &PurpleTheme{}
}
