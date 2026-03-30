package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mzaran/w9s/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfigFile(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

const fullYAML = `
refreshRate: "10s"
defaultCluster: "prod"
clusters:
  - name: "prod"
    cluster:
      endpoint: "https://ww-prod:9873"
      username: "admin"
      password: "secret"
      insecure: true
      timeout: "30s"
  - name: "dev"
    cluster:
      endpoint: "https://ww-dev:9873"
      username: "devuser"
      password: "devpass"
      insecure: false
      timeout: "15s"
ui:
  enableMouse: true
`

func TestLoadFullConfig(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, fullYAML)

	cfg, err := config.Load(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "10s", cfg.RefreshRate)
	assert.Equal(t, "prod", cfg.DefaultCluster)
	assert.True(t, cfg.UI.EnableMouse)

	require.Len(t, cfg.Clusters, 2)
	assert.Equal(t, "prod", cfg.Clusters[0].Name)
	assert.Equal(t, "https://ww-prod:9873", cfg.Clusters[0].Cluster.Endpoint)
	assert.Equal(t, "admin", cfg.Clusters[0].Cluster.Username)
	assert.Equal(t, "secret", cfg.Clusters[0].Cluster.Password)
	assert.True(t, cfg.Clusters[0].Cluster.Insecure)
	assert.Equal(t, "30s", cfg.Clusters[0].Cluster.Timeout)

	assert.Equal(t, "dev", cfg.Clusters[1].Name)
	assert.Equal(t, "https://ww-dev:9873", cfg.Clusters[1].Cluster.Endpoint)
}

func TestLoadMissingExplicitFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading config file")
}

func TestLoadNoFileNoEnv(t *testing.T) {
	// With no config file found and no env vars, Load searches standard
	// locations. On a clean temp HOME it should fail validation because
	// no clusters are configured (unless /etc/warewulf/warewulf.conf exists).
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Clear any W9S env vars that might be set.
	t.Setenv("W9S_ENDPOINT", "")
	t.Setenv("W9S_USERNAME", "")
	t.Setenv("W9S_PASSWORD", "")

	_, err := config.Load("")
	// Either succeeds (if warewulf.conf auto-detect works) or fails with
	// "no clusters configured".
	if err != nil {
		assert.Contains(t, err.Error(), "no clusters configured")
	}
}

func TestEnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, fullYAML)

	t.Setenv("W9S_ENDPOINT", "https://env-override:9999")
	t.Setenv("W9S_USERNAME", "envuser")
	t.Setenv("W9S_PASSWORD", "envpass")

	cfg, err := config.Load(path)
	require.NoError(t, err)

	// Environment overrides apply to the first cluster.
	assert.Equal(t, "https://env-override:9999", cfg.Clusters[0].Cluster.Endpoint)
	assert.Equal(t, "envuser", cfg.Clusters[0].Cluster.Username)
	assert.Equal(t, "envpass", cfg.Clusters[0].Cluster.Password)
}

func TestEnvOnlyNoFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("W9S_ENDPOINT", "https://env-only:9873")
	t.Setenv("W9S_USERNAME", "admin")
	t.Setenv("W9S_PASSWORD", "pass")

	cfg, err := config.Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.True(t, len(cfg.Clusters) > 0)
	assert.Equal(t, "https://env-only:9873", cfg.Clusters[0].Cluster.Endpoint)
	assert.Equal(t, "admin", cfg.Clusters[0].Cluster.Username)
}

func TestValidateMissingEndpoint(t *testing.T) {
	dir := t.TempDir()
	yaml := `
clusters:
  - name: "bad"
    cluster:
      endpoint: ""
      username: "u"
`
	path := writeConfigFile(t, dir, yaml)

	t.Setenv("W9S_ENDPOINT", "")
	t.Setenv("W9S_USERNAME", "")
	t.Setenv("W9S_PASSWORD", "")

	_, err := config.Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no endpoint")
}

func TestMultipleClustersActiveClusterResolution(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, fullYAML)

	t.Setenv("W9S_ENDPOINT", "")
	t.Setenv("W9S_USERNAME", "")
	t.Setenv("W9S_PASSWORD", "")

	cfg, err := config.Load(path)
	require.NoError(t, err)

	// defaultCluster is "prod", so ActiveCluster should point to the prod cluster.
	require.NotNil(t, cfg.ActiveCluster)
	assert.Equal(t, "https://ww-prod:9873", cfg.ActiveCluster.Endpoint)
	assert.Equal(t, "admin", cfg.ActiveCluster.Username)
}

func TestActiveClusterFallsBackToFirst(t *testing.T) {
	dir := t.TempDir()
	yaml := `
clusters:
  - name: "alpha"
    cluster:
      endpoint: "https://alpha:9873"
  - name: "beta"
    cluster:
      endpoint: "https://beta:9873"
`
	path := writeConfigFile(t, dir, yaml)

	t.Setenv("W9S_ENDPOINT", "")
	t.Setenv("W9S_USERNAME", "")
	t.Setenv("W9S_PASSWORD", "")

	cfg, err := config.Load(path)
	require.NoError(t, err)

	// No defaultCluster set, so ActiveCluster should be the first entry.
	require.NotNil(t, cfg.ActiveCluster)
	assert.Equal(t, "https://alpha:9873", cfg.ActiveCluster.Endpoint)
}

func TestTimeoutFieldParsed(t *testing.T) {
	dir := t.TempDir()
	yaml := `
clusters:
  - name: "test"
    cluster:
      endpoint: "https://test:9873"
      timeout: "45s"
`
	path := writeConfigFile(t, dir, yaml)

	t.Setenv("W9S_ENDPOINT", "")
	t.Setenv("W9S_USERNAME", "")
	t.Setenv("W9S_PASSWORD", "")

	cfg, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "45s", cfg.Clusters[0].Cluster.Timeout)
}

func TestDefaultRefreshRate(t *testing.T) {
	dir := t.TempDir()
	yaml := `
clusters:
  - name: "test"
    cluster:
      endpoint: "https://test:9873"
`
	path := writeConfigFile(t, dir, yaml)

	t.Setenv("W9S_ENDPOINT", "")
	t.Setenv("W9S_USERNAME", "")
	t.Setenv("W9S_PASSWORD", "")

	cfg, err := config.Load(path)
	require.NoError(t, err)
	// Default refreshRate should be "5s" per the code.
	assert.Equal(t, "5s", cfg.RefreshRate)
}

func TestXDGPathPreference(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("W9S_ENDPOINT", "")
	t.Setenv("W9S_USERNAME", "")
	t.Setenv("W9S_PASSWORD", "")

	// Create config in both locations with different endpoints.
	xdgDir := filepath.Join(home, ".config", "w9s")
	dotDir := filepath.Join(home, ".w9s")
	require.NoError(t, os.MkdirAll(xdgDir, 0o755))
	require.NoError(t, os.MkdirAll(dotDir, 0o755))

	xdgYAML := `
clusters:
  - name: "xdg"
    cluster:
      endpoint: "https://xdg-endpoint:9873"
`
	dotYAML := `
clusters:
  - name: "dot"
    cluster:
      endpoint: "https://dot-endpoint:9873"
`
	require.NoError(t, os.WriteFile(filepath.Join(xdgDir, "config.yaml"), []byte(xdgYAML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dotDir, "config.yaml"), []byte(dotYAML), 0o644))

	cfg, err := config.Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// XDG path (~/.config/w9s) is added first, so it should be preferred.
	assert.Equal(t, "xdg", cfg.Clusters[0].Name)
	assert.Equal(t, "https://xdg-endpoint:9873", cfg.Clusters[0].Cluster.Endpoint)
}
