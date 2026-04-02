package ui

import "github.com/gdamore/tcell/v2"

// Theme defines the color palette for the w9s TUI.
// Single source of truth — swap themes by changing one struct.
type Theme struct {
	LogoColor     tcell.Color
	HeaderBg      tcell.Color
	HeaderFg      tcell.Color
	BrandColor    tcell.Color
	ConnectedFg   tcell.Color
	MetricLabel   tcell.Color
	MetricValue   tcell.Color
	TabActiveBg   tcell.Color
	TabActiveFg   tcell.Color
	TabNumber     tcell.Color
	TabName       tcell.Color
	TableHeaderFg tcell.Color
	TableSepColor tcell.Color
	TableRowFg    tcell.Color
	SelectionBg   tcell.Color
	SelectionFg   tcell.Color
	StatusReady   tcell.Color
	StatusDown    tcell.Color
	StatusPending tcell.Color
	HotkeyFg     tcell.Color
	HintFg       tcell.Color
	CountFg      tcell.Color
	TextFg       tcell.Color
	DimFg        tcell.Color
	BgColor      tcell.Color
	BorderColor  tcell.Color
	AccentColor  tcell.Color
}

// DefaultTheme is the Design C color palette.
var DefaultTheme = Theme{
	LogoColor:     tcell.NewRGBColor(0, 188, 212),
	HeaderBg:      tcell.ColorDefault,
	HeaderFg:      tcell.ColorWhite,
	BrandColor:    tcell.NewRGBColor(100, 100, 100),
	ConnectedFg:   tcell.NewRGBColor(76, 175, 80),
	MetricLabel:   tcell.NewRGBColor(136, 136, 136),
	MetricValue:   tcell.ColorWhite,
	TabActiveBg:   tcell.NewRGBColor(70, 130, 180),
	TabActiveFg:   tcell.ColorWhite,
	TabNumber:     tcell.NewRGBColor(70, 130, 180),
	TabName:       tcell.NewRGBColor(136, 136, 136),
	TableHeaderFg: tcell.ColorWhite,
	TableSepColor: tcell.NewRGBColor(0, 188, 212),
	TableRowFg:    tcell.NewRGBColor(200, 200, 200),
	SelectionBg:   tcell.NewRGBColor(26, 35, 50),
	SelectionFg:   tcell.ColorWhite,
	StatusReady:   tcell.NewRGBColor(76, 175, 80),
	StatusDown:    tcell.NewRGBColor(244, 67, 54),
	StatusPending: tcell.NewRGBColor(255, 193, 7),
	HotkeyFg:     tcell.NewRGBColor(255, 193, 7),
	HintFg:       tcell.NewRGBColor(136, 136, 136),
	CountFg:      tcell.NewRGBColor(100, 100, 100),
	TextFg:       tcell.ColorWhite,
	DimFg:        tcell.NewRGBColor(136, 136, 136),
	BgColor:      tcell.ColorDefault,
	BorderColor:  tcell.NewRGBColor(136, 136, 136),
	AccentColor:  tcell.NewRGBColor(255, 193, 7),
}

// ColorToHex converts a tcell.Color to a 24-bit hex integer for tview markup.
func ColorToHex(c tcell.Color) int32 {
	r, g, b := c.RGB()
	return int32(r)<<16 | int32(g)<<8 | int32(b)
}
