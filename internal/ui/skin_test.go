package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSkinBuiltinDefault(t *testing.T) {
	skin, err := LoadSkin("default")
	require.NoError(t, err)
	assert.NotNil(t, skin)
	assert.Equal(t, "default", skin.Name)
}

func TestLoadSkinBuiltinDracula(t *testing.T) {
	skin, err := LoadSkin("dracula")
	require.NoError(t, err)
	assert.NotNil(t, skin)
	assert.Equal(t, "dracula", skin.Name)
	assert.Equal(t, "#282a36", skin.Header.Background)
	assert.Equal(t, "#f8f8f2", skin.Header.Foreground)
}

func TestLoadSkinNonexistent(t *testing.T) {
	_, err := LoadSkin("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestAllBuiltinSkinsExist(t *testing.T) {
	expected := []string{"default", "dracula", "gruvbox", "nord", "solarized"}
	for _, name := range expected {
		skin, err := LoadSkin(name)
		require.NoError(t, err, "builtin skin %q should load", name)
		assert.Equal(t, name, skin.Name)
	}
}

func TestSkinToThemeConvertsColors(t *testing.T) {
	skin := &Skin{
		Name: "test",
		Header: SkinColors{
			Foreground: "#ff0000",
			Background: "#00ff00",
		},
		Table: TableColors{
			Header:  "#0000ff",
			RowEven: "#111111",
			RowOdd:  "#222222",
		},
		Status: StatusColors{
			Success: "#00ff00",
			Error:   "#ff0000",
			Info:    "#0000ff",
		},
		Border: "#333333",
		Accent: "#ffff00",
	}

	theme := SkinToTheme(skin)
	assert.NotNil(t, theme)

	// Header foreground mapped to HeaderFg.
	assert.Equal(t, tcell.GetColor("#ff0000"), theme.HeaderFg)
	// Header background mapped to HeaderBg.
	assert.Equal(t, tcell.GetColor("#00ff00"), theme.HeaderBg)
	// Table header mapped to TableHeaderFg.
	assert.Equal(t, tcell.GetColor("#0000ff"), theme.TableHeaderFg)
	// Status success mapped to StatusReady.
	assert.Equal(t, tcell.GetColor("#00ff00"), theme.StatusReady)
	// Status error mapped to StatusDown.
	assert.Equal(t, tcell.GetColor("#ff0000"), theme.StatusDown)
	// Border mapped to BorderColor.
	assert.Equal(t, tcell.GetColor("#333333"), theme.BorderColor)
	// Accent mapped to AccentColor.
	assert.Equal(t, tcell.GetColor("#ffff00"), theme.AccentColor)
}

func TestSkinToThemeFallsBackToDefaults(t *testing.T) {
	skin := &Skin{Name: "empty"}
	theme := SkinToTheme(skin)

	// Should preserve DefaultTheme values when skin fields are empty.
	assert.Equal(t, DefaultTheme.HeaderFg, theme.HeaderFg)
	assert.Equal(t, DefaultTheme.StatusReady, theme.StatusReady)
	assert.Equal(t, DefaultTheme.BorderColor, theme.BorderColor)
}

func TestParseColorHex(t *testing.T) {
	c := parseColor("#ff5555")
	r, g, b := c.RGB()
	assert.Equal(t, int32(255), r)
	assert.Equal(t, int32(85), g)
	assert.Equal(t, int32(85), b)
}

func TestParseColorEmpty(t *testing.T) {
	c := parseColor("")
	assert.Equal(t, tcell.ColorDefault, c)
}

func TestParseColorNamed(t *testing.T) {
	c := parseColor("red")
	assert.NotEqual(t, tcell.ColorDefault, c)
}

func TestLoadSkinFromYAMLFile(t *testing.T) {
	// Create a temporary skins directory.
	tmpDir := t.TempDir()
	skinsDir := filepath.Join(tmpDir, "w9s", "skins")
	require.NoError(t, os.MkdirAll(skinsDir, 0o755))

	skinYAML := `name: custom
header:
  foreground: "#aabbcc"
  background: "#112233"
table:
  header: "#445566"
  row_even: "#778899"
  row_odd: "#aabbcc"
status:
  success: "#00ff00"
  error: "#ff0000"
  info: "#0000ff"
border: "#999999"
accent: "#ff00ff"
`
	require.NoError(t, os.WriteFile(filepath.Join(skinsDir, "custom.yaml"), []byte(skinYAML), 0o644))

	// Override XDG config dir for test.
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	skin, err := LoadSkin("custom")
	require.NoError(t, err)
	assert.Equal(t, "custom", skin.Name)
	assert.Equal(t, "#aabbcc", skin.Header.Foreground)
	assert.Equal(t, "#999999", skin.Border)
}

func TestBuiltinSkinReturnsCopy(t *testing.T) {
	skin1, _ := LoadSkin("dracula")
	skin2, _ := LoadSkin("dracula")
	skin1.Name = "modified"
	assert.Equal(t, "dracula", skin2.Name, "modifying one skin should not affect another")
}
