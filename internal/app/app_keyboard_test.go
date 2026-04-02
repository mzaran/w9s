package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/mzaran/w9s/internal/config"
	"github.com/mzaran/w9s/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	client := mock.NewFastMockClient()
	cfg := newTestConfig()
	a, err := New(client, cfg)
	require.NoError(t, err)
	return a
}

func TestNumberKeysSwitchViews(t *testing.T) {
	a := newTestApp(t)

	names := a.viewMgr.ViewNames()
	for i := 0; i < len(names) && i < 7; i++ {
		r := rune('1' + i)
		event := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
		result := a.inputCapture(event)
		assert.Nil(t, result, "number key %c should be consumed", r)
		assert.Equal(t, names[i], a.viewMgr.CurrentViewName(), "key %c should switch to %s", r, names[i])
	}
}

func TestQKeyStopsApp(t *testing.T) {
	a := newTestApp(t)

	event := tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)
	result := a.inputCapture(event)
	assert.Nil(t, result, "'q' should be consumed")

	select {
	case <-a.ctx.Done():
		// OK - q called Stop which cancels context
	default:
		t.Fatal("context should be cancelled after 'q'")
	}
}

func TestQuestionMarkSwitchesToHelp(t *testing.T) {
	a := newTestApp(t)

	// Start on dashboard.
	a.switchToView("dashboard")
	assert.Equal(t, "dashboard", a.viewMgr.CurrentViewName())

	event := tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone)
	result := a.inputCapture(event)
	assert.Nil(t, result, "'?' should be consumed")
	assert.Equal(t, "help", a.viewMgr.CurrentViewName())
}

func TestUpperCDoesNotCrash(t *testing.T) {
	// With only 1 cluster, 'C' should show error but not crash.
	a := newTestApp(t)

	assert.NotPanics(t, func() {
		event := tcell.NewEventKey(tcell.KeyRune, 'C', tcell.ModNone)
		result := a.inputCapture(event)
		assert.Nil(t, result, "'C' should be consumed")
	})
}

func TestTabCyclesToNextView(t *testing.T) {
	a := newTestApp(t)

	names := a.viewMgr.ViewNames()
	// Set to first view.
	a.switchToView(names[0])

	event := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	result := a.inputCapture(event)
	assert.Nil(t, result, "Tab should be consumed")
	assert.Equal(t, names[1], a.viewMgr.CurrentViewName())
}

func TestShiftTabCyclesToPreviousView(t *testing.T) {
	a := newTestApp(t)

	names := a.viewMgr.ViewNames()
	// Set to second view.
	a.switchToView(names[1])

	event := tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
	result := a.inputCapture(event)
	assert.Nil(t, result, "Shift+Tab should be consumed")
	assert.Equal(t, names[0], a.viewMgr.CurrentViewName())
}

func TestCtrlCStopsApp(t *testing.T) {
	a := newTestApp(t)

	event := tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModCtrl)
	result := a.inputCapture(event)
	assert.Nil(t, result, "Ctrl+C should be consumed")

	select {
	case <-a.ctx.Done():
		// OK
	default:
		t.Fatal("context should be cancelled after Ctrl+C")
	}
}

func TestColonDoesNotCrash(t *testing.T) {
	a := newTestApp(t)

	assert.NotPanics(t, func() {
		event := tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone)
		result := a.inputCapture(event)
		assert.Nil(t, result, "':' should be consumed (returns nil)")
	})
}

func TestSlashPassesThrough(t *testing.T) {
	a := newTestApp(t)

	event := tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone)
	result := a.inputCapture(event)
	// '/' is delegated to the view; since the mock view returns the event, it comes back.
	assert.NotNil(t, result, "'/' should pass through to view")
}

func TestUnhandledKeyPassesThrough(t *testing.T) {
	a := newTestApp(t)

	event := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
	result := a.inputCapture(event)
	// 'x' is not a global shortcut; should pass through.
	assert.NotNil(t, result, "unhandled key 'x' should pass through")
}

func TestUpperCWithMultipleClusters(t *testing.T) {
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

	// With 2 clusters, 'C' should show the switcher without crashing.
	assert.NotPanics(t, func() {
		event := tcell.NewEventKey(tcell.KeyRune, 'C', tcell.ModNone)
		result := a.inputCapture(event)
		assert.Nil(t, result, "'C' should be consumed")
	})
}
