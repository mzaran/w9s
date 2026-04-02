package version_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/version"
	"github.com/stretchr/testify/assert"
)

func TestInfoDefaults(t *testing.T) {
	origVersion := version.Version
	t.Cleanup(func() {
		version.Version = origVersion
	})

	version.Version = "dev"
	assert.Equal(t, "dev", version.Info())
}

func TestInfoReflectsSetValues(t *testing.T) {
	origVersion := version.Version
	t.Cleanup(func() {
		version.Version = origVersion
	})

	version.Version = "1.2.3"
	assert.Equal(t, "1.2.3", version.Info())
}

func TestInfoFormat(t *testing.T) {
	origVersion := version.Version
	t.Cleanup(func() {
		version.Version = origVersion
	})

	version.Version = "v0.1.0"
	assert.Equal(t, "v0.1.0", version.Info())
}

func TestLogoSmall(t *testing.T) {
	assert.Len(t, version.LogoSmall, 6, "LogoSmall should have 6 lines")
	// Verify the logo contains expected characters.
	joined := ""
	for _, line := range version.LogoSmall {
		joined += line
	}
	assert.Contains(t, joined, "___", "LogoSmall should contain underscores from ASCII art")
}

func TestGitRepo(t *testing.T) {
	assert.Equal(t, "github.com/mzaran/w9s", version.GitRepo)
}
