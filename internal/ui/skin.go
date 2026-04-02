// Package ui skin.go implements YAML-based skin/theme loading for w9s.
package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"gopkg.in/yaml.v3"
)

// Skin defines a color palette loadable from YAML.
type Skin struct {
	Name   string       `yaml:"name"`
	Header SkinColors   `yaml:"header"`
	Table  TableColors  `yaml:"table"`
	Status StatusColors `yaml:"status"`
	Border string       `yaml:"border"`
	Accent string       `yaml:"accent"`
}

// SkinColors holds foreground/background pairs.
type SkinColors struct {
	Foreground string `yaml:"foreground"`
	Background string `yaml:"background"`
}

// TableColors holds table-specific colors.
type TableColors struct {
	Header  string `yaml:"header"`
	RowEven string `yaml:"row_even"`
	RowOdd  string `yaml:"row_odd"`
}

// StatusColors holds status indicator colors.
type StatusColors struct {
	Success string `yaml:"success"`
	Error   string `yaml:"error"`
	Info    string `yaml:"info"`
}

// LoadSkin loads a skin by name. Checks built-in skins first, then ~/.config/w9s/skins/.
func LoadSkin(name string) (*Skin, error) {
	if skin, ok := builtinSkins[name]; ok {
		s := skin // copy
		return &s, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("skin %q not found: %w", name, err)
	}
	path := filepath.Join(configDir, "w9s", "skins", name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("skin %q not found: %w", name, err)
	}

	var skin Skin
	if err := yaml.Unmarshal(data, &skin); err != nil {
		return nil, fmt.Errorf("invalid skin %q: %w", name, err)
	}
	return &skin, nil
}

// SkinToTheme converts a Skin to a Theme for use by UI components.
// Missing fields fall back to DefaultTheme values.
func SkinToTheme(s *Skin) *Theme {
	theme := DefaultTheme // copy defaults

	if s.Header.Foreground != "" {
		theme.HeaderFg = parseColor(s.Header.Foreground)
		theme.MetricValue = parseColor(s.Header.Foreground)
		theme.TextFg = parseColor(s.Header.Foreground)
	}
	if s.Header.Background != "" {
		theme.HeaderBg = parseColor(s.Header.Background)
		theme.BgColor = parseColor(s.Header.Background)
	}
	if s.Table.Header != "" {
		theme.TableHeaderFg = parseColor(s.Table.Header)
		theme.TabActiveBg = parseColor(s.Table.Header)
		theme.TabNumber = parseColor(s.Table.Header)
	}
	if s.Table.RowEven != "" {
		theme.SelectionBg = parseColor(s.Table.RowEven)
	}
	if s.Table.RowOdd != "" {
		theme.TableRowFg = parseColor(s.Table.RowOdd)
	}
	if s.Status.Success != "" {
		theme.StatusReady = parseColor(s.Status.Success)
		theme.ConnectedFg = parseColor(s.Status.Success)
	}
	if s.Status.Error != "" {
		theme.StatusDown = parseColor(s.Status.Error)
	}
	if s.Status.Info != "" {
		theme.LogoColor = parseColor(s.Status.Info)
		theme.TableSepColor = parseColor(s.Status.Info)
	}
	if s.Border != "" {
		theme.BorderColor = parseColor(s.Border)
		theme.DimFg = parseColor(s.Border)
		theme.HintFg = parseColor(s.Border)
		theme.MetricLabel = parseColor(s.Border)
		theme.TabName = parseColor(s.Border)
		theme.CountFg = parseColor(s.Border)
		theme.BrandColor = parseColor(s.Border)
	}
	if s.Accent != "" {
		theme.AccentColor = parseColor(s.Accent)
		theme.HotkeyFg = parseColor(s.Accent)
		theme.StatusPending = parseColor(s.Accent)
	}
	return &theme
}

// parseColor converts a color string (hex like "#ff5555" or named like "red")
// to a tcell.Color. Returns tcell.ColorDefault for empty strings.
func parseColor(s string) tcell.Color {
	if s == "" {
		return tcell.ColorDefault
	}
	return tcell.GetColor(s)
}

// builtinSkins contains the built-in skin definitions.
var builtinSkins = map[string]Skin{
	"default": {
		Name: "default",
		Header: SkinColors{
			Foreground: "#ffffff",
			Background: "",
		},
		Table: TableColors{
			Header:  "#4682b4",
			RowEven: "#1a2332",
			RowOdd:  "#c8c8c8",
		},
		Status: StatusColors{
			Success: "#4caf50",
			Error:   "#f44336",
			Info:    "#00bcd4",
		},
		Border: "#888888",
		Accent: "#ffc107",
	},
	"dracula": {
		Name: "dracula",
		Header: SkinColors{
			Foreground: "#f8f8f2",
			Background: "#282a36",
		},
		Table: TableColors{
			Header:  "#bd93f9",
			RowEven: "#44475a",
			RowOdd:  "#f8f8f2",
		},
		Status: StatusColors{
			Success: "#50fa7b",
			Error:   "#ff5555",
			Info:    "#8be9fd",
		},
		Border: "#6272a4",
		Accent: "#ff79c6",
	},
	"gruvbox": {
		Name: "gruvbox",
		Header: SkinColors{
			Foreground: "#ebdbb2",
			Background: "#282828",
		},
		Table: TableColors{
			Header:  "#fabd2f",
			RowEven: "#3c3836",
			RowOdd:  "#ebdbb2",
		},
		Status: StatusColors{
			Success: "#b8bb26",
			Error:   "#fb4934",
			Info:    "#83a598",
		},
		Border: "#928374",
		Accent: "#fe8019",
	},
	"nord": {
		Name: "nord",
		Header: SkinColors{
			Foreground: "#eceff4",
			Background: "#2e3440",
		},
		Table: TableColors{
			Header:  "#88c0d0",
			RowEven: "#3b4252",
			RowOdd:  "#d8dee9",
		},
		Status: StatusColors{
			Success: "#a3be8c",
			Error:   "#bf616a",
			Info:    "#81a1c1",
		},
		Border: "#4c566a",
		Accent: "#ebcb8b",
	},
	"solarized": {
		Name: "solarized",
		Header: SkinColors{
			Foreground: "#839496",
			Background: "#002b36",
		},
		Table: TableColors{
			Header:  "#268bd2",
			RowEven: "#073642",
			RowOdd:  "#839496",
		},
		Status: StatusColors{
			Success: "#859900",
			Error:   "#dc322f",
			Info:    "#2aa198",
		},
		Border: "#586e75",
		Accent: "#b58900",
	},
}
