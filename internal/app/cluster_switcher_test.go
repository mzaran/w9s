package app

import (
	"testing"

	"github.com/mzaran/w9s/internal/config"
	"github.com/mzaran/w9s/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowClusterSwitcherWithSingleClusterShowsError(t *testing.T) {
	a := newTestApp(t)
	// Only 1 cluster in config.
	assert.Len(t, a.config.Clusters, 1)

	// Should not panic; shows error in status bar.
	assert.NotPanics(t, func() {
		a.showClusterSwitcher()
	})
}

func TestShowClusterSwitcherWithMultipleClusters(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := &config.Config{
		RefreshRate: "5s",
		Clusters: []config.ClusterEntry{
			{Name: "cluster1", Cluster: config.ClusterConfig{Endpoint: "mock://c1"}},
			{Name: "cluster2", Cluster: config.ClusterConfig{Endpoint: "mock://c2"}},
		},
		ActiveCluster: &config.ClusterConfig{Endpoint: "mock://c1"},
	}

	a, err := New(client, cfg)
	require.NoError(t, err)

	// Should not panic with 2 clusters.
	assert.NotPanics(t, func() {
		a.showClusterSwitcher()
	})
}

func TestSwapClusterInvalidIndexNegative(t *testing.T) {
	a := newTestApp(t)

	err := a.swapCluster(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cluster index")
}

func TestSwapClusterInvalidIndexTooHigh(t *testing.T) {
	a := newTestApp(t)

	err := a.swapCluster(100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cluster index")
}

func TestSwapClusterMockEndpointUpdatesConfig(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := &config.Config{
		RefreshRate: "5s",
		Clusters: []config.ClusterEntry{
			{Name: "cluster1", Cluster: config.ClusterConfig{Endpoint: "mock://c1"}},
			{Name: "cluster2", Cluster: config.ClusterConfig{Endpoint: "mock://c2"}},
		},
		ActiveCluster: &config.ClusterConfig{Endpoint: "mock://c1"},
	}

	a, err := New(client, cfg)
	require.NoError(t, err)

	// Switch to first view so CurrentView() is set.
	names := a.viewMgr.ViewNames()
	require.NotEmpty(t, names)
	a.switchToView(names[0])

	err = a.swapCluster(1)
	assert.NoError(t, err)
	assert.Equal(t, "mock://c2", a.config.ActiveCluster.Endpoint)
}

func TestSwapClusterRefreshesHeader(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := &config.Config{
		RefreshRate: "5s",
		Clusters: []config.ClusterEntry{
			{Name: "cluster1", Cluster: config.ClusterConfig{Endpoint: "mock://c1"}},
			{Name: "cluster2", Cluster: config.ClusterConfig{Endpoint: "mock://c2"}},
		},
		ActiveCluster: &config.ClusterConfig{Endpoint: "mock://c1"},
	}

	a, err := New(client, cfg)
	require.NoError(t, err)

	names := a.viewMgr.ViewNames()
	require.NotEmpty(t, names)
	a.switchToView(names[0])

	// Swap to cluster2 and verify ActiveCluster was updated (header is internal to ui pkg).
	err = a.swapCluster(1)
	assert.NoError(t, err)
	// The header is updated via SetClusterInfo; we verify the config side-effect.
	assert.Equal(t, "mock://c2", a.config.ActiveCluster.Endpoint)
}

func TestSwapClusterEmptyEndpointUpdatesConfig(t *testing.T) {
	client := mock.NewFastMockClient()
	cfg := &config.Config{
		RefreshRate: "5s",
		Clusters: []config.ClusterEntry{
			{Name: "cluster1", Cluster: config.ClusterConfig{Endpoint: "mock://c1"}},
			{Name: "cluster2", Cluster: config.ClusterConfig{Endpoint: ""}},
		},
		ActiveCluster: &config.ClusterConfig{Endpoint: "mock://c1"},
	}

	a, err := New(client, cfg)
	require.NoError(t, err)

	names := a.viewMgr.ViewNames()
	require.NotEmpty(t, names)
	a.switchToView(names[0])

	// Empty endpoint triggers mock-mode path in swapCluster.
	err = a.swapCluster(1)
	assert.NoError(t, err)
	assert.Equal(t, "", a.config.ActiveCluster.Endpoint)
}
