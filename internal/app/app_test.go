package app

import (
	"testing"

	"github.com/mzaran/w9s/internal/config"
	"github.com/mzaran/w9s/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig() *config.Config {
	return &config.Config{
		RefreshRate: "5s",
		Clusters: []config.ClusterEntry{
			{
				Name: "test",
				Cluster: config.ClusterConfig{
					Endpoint: "mock://test",
				},
			},
		},
		ActiveCluster: &config.ClusterConfig{
			Endpoint: "mock://test",
		},
	}
}

func TestNewCreatesAppWithNonNilFields(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.NotNil(t, a.client)
	assert.NotNil(t, a.tviewApp)
	assert.NotNil(t, a.viewMgr)
	assert.NotNil(t, a.pages)
	assert.NotNil(t, a.header)
	assert.NotNil(t, a.statusBar)
	assert.NotNil(t, a.config)
	assert.NotNil(t, a.ctx)
	assert.NotNil(t, a.cancel)
}

func TestNewWithMouseEnabled(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()
	cfg.UI.EnableMouse = true

	a, err := New(client, cfg)
	require.NoError(t, err)
	require.NotNil(t, a)
	// If EnableMouse doesn't panic and app is created, that's sufficient.
	assert.NotNil(t, a.tviewApp)
}

func TestClientReturnsCorrectValue(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	assert.Equal(t, client, a.Client())
}

func TestConfigReturnsCorrectValue(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	assert.Equal(t, cfg, a.Config())
}

func TestTviewAppReturnsCorrectValue(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	assert.NotNil(t, a.TviewApp())
}

func TestViewMgrReturnsCorrectValue(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	assert.NotNil(t, a.ViewMgr())
}

func TestContextReturnsNonNilContext(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	ctx := a.Context()
	assert.NotNil(t, ctx)
	// Context should not be cancelled yet.
	select {
	case <-ctx.Done():
		t.Fatal("context should not be done yet")
	default:
		// OK
	}
}

func TestStopCancelsContext(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	ctx := a.Context()
	a.Stop()

	select {
	case <-ctx.Done():
		// OK - context was cancelled
	default:
		t.Fatal("context should be done after Stop()")
	}
}

func TestSwitchToViewChangesCurrentView(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	names := a.viewMgr.ViewNames()
	require.True(t, len(names) >= 2, "expected at least 2 views registered")

	// Switch to second view.
	a.switchToView(names[1])
	assert.Equal(t, names[1], a.viewMgr.CurrentViewName())

	// Switch back to first view.
	a.switchToView(names[0])
	assert.Equal(t, names[0], a.viewMgr.CurrentViewName())
}

func TestSwitchToViewInvalidNameNoOp(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	current := a.viewMgr.CurrentViewName()
	a.switchToView("nonexistent-view")
	// Should remain on current view.
	assert.Equal(t, current, a.viewMgr.CurrentViewName())
}

func TestViewsRegisteredInCorrectOrder(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	names := a.viewMgr.ViewNames()
	// Mock client has power, so expect all 7 views.
	expected := []string{"dashboard", "nodes", "profiles", "images", "overlays", "power", "help"}
	assert.Equal(t, expected, names)
}

func TestInitViewRegistersAndAddsPage(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := newTestConfig()

	a, err := New(client, cfg)
	require.NoError(t, err)

	// All views should be accessible by name.
	for _, name := range a.viewMgr.ViewNames() {
		v := a.viewMgr.GetView(name)
		assert.NotNil(t, v, "view %q should be registered", name)
	}
}
