package views_test

import (
	"context"
	"testing"

	"github.com/mzaran/w9s/internal/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseViewNameAndTitle(t *testing.T) {
	bv := views.NewBaseView("nodes", "Nodes")
	assert.Equal(t, "nodes", bv.Name())
	assert.Equal(t, "Nodes", bv.Title())
}

func TestBaseViewDefaultMethodsDoNotPanic(t *testing.T) {
	bv := views.NewBaseView("test", "Test")

	assert.NotPanics(t, func() {
		_ = bv.Init(context.Background())
	})
	assert.NotPanics(t, func() {
		_ = bv.Refresh()
	})
	assert.NotPanics(t, func() {
		_ = bv.OnKey(nil)
	})
	assert.NotPanics(t, func() {
		_ = bv.OnFocus()
	})
	assert.NotPanics(t, func() {
		_ = bv.OnLoseFocus()
	})
	assert.NotPanics(t, func() {
		_ = bv.Stop()
	})
}

func TestBaseViewHintsDefault(t *testing.T) {
	bv := views.NewBaseView("test", "Test")
	assert.Nil(t, bv.Hints())
}

func TestBaseViewRenderDefault(t *testing.T) {
	bv := views.NewBaseView("test", "Test")
	assert.Nil(t, bv.Render())
}

func TestBaseViewSetSwitchViewFnAndSwitchToView(t *testing.T) {
	bv := views.NewBaseView("test", "Test")

	var switched string
	bv.SetSwitchViewFn(func(name string) {
		switched = name
	})
	bv.SwitchToView("other")
	assert.Equal(t, "other", switched)
}

func TestBaseViewSwitchToViewWithoutFn(t *testing.T) {
	bv := views.NewBaseView("test", "Test")
	// Should not panic when switchViewFn is nil.
	assert.NotPanics(t, func() {
		bv.SwitchToView("other")
	})
}

func TestBaseViewFocusTracking(t *testing.T) {
	bv := views.NewBaseView("test", "Test")
	assert.False(t, bv.IsFocused())

	require.NoError(t, bv.OnFocus())
	assert.True(t, bv.IsFocused())

	require.NoError(t, bv.OnLoseFocus())
	assert.False(t, bv.IsFocused())
}

func TestBaseViewRefreshingState(t *testing.T) {
	bv := views.NewBaseView("test", "Test")
	assert.False(t, bv.IsRefreshing())

	bv.SetRefreshing(true)
	assert.True(t, bv.IsRefreshing())

	bv.SetRefreshing(false)
	assert.False(t, bv.IsRefreshing())
}

func TestBaseViewLastError(t *testing.T) {
	bv := views.NewBaseView("test", "Test")
	assert.Nil(t, bv.LastError())

	bv.SetLastError(assert.AnError)
	assert.Equal(t, assert.AnError, bv.LastError())
}

// -- ViewManager tests --

func TestViewManagerRegister(t *testing.T) {
	mgr := views.NewViewManager()
	v1 := &stubView{BaseView: views.NewBaseView("nodes", "Nodes")}
	v2 := &stubView{BaseView: views.NewBaseView("images", "Images")}

	mgr.Register(v1)
	mgr.Register(v2)

	assert.NotNil(t, mgr.GetView("nodes"))
	assert.NotNil(t, mgr.GetView("images"))
	assert.Nil(t, mgr.GetView("nonexistent"))
}

func TestViewManagerFirstRegisteredIsDefault(t *testing.T) {
	mgr := views.NewViewManager()
	v1 := &stubView{BaseView: views.NewBaseView("first", "First")}
	v2 := &stubView{BaseView: views.NewBaseView("second", "Second")}

	mgr.Register(v1)
	mgr.Register(v2)

	assert.Equal(t, "first", mgr.CurrentViewName())
}

func TestViewManagerSetCurrentView(t *testing.T) {
	mgr := views.NewViewManager()
	v1 := &stubView{BaseView: views.NewBaseView("nodes", "Nodes")}
	v2 := &stubView{BaseView: views.NewBaseView("images", "Images")}

	mgr.Register(v1)
	mgr.Register(v2)

	err := mgr.SetCurrentView("images")
	require.NoError(t, err)
	assert.Equal(t, "images", mgr.CurrentViewName())
}

func TestViewManagerSetCurrentViewUnknown(t *testing.T) {
	mgr := views.NewViewManager()
	v1 := &stubView{BaseView: views.NewBaseView("nodes", "Nodes")}
	mgr.Register(v1)

	// Setting an unknown view returns nil (no error) but doesn't change current.
	err := mgr.SetCurrentView("nonexistent")
	require.NoError(t, err)
	assert.Equal(t, "nodes", mgr.CurrentViewName())
}

func TestViewManagerNextViewCyclesCircularly(t *testing.T) {
	mgr := views.NewViewManager()
	mgr.Register(&stubView{BaseView: views.NewBaseView("a", "A")})
	mgr.Register(&stubView{BaseView: views.NewBaseView("b", "B")})
	mgr.Register(&stubView{BaseView: views.NewBaseView("c", "C")})

	// Current is "a".
	assert.Equal(t, "b", mgr.NextView())

	_ = mgr.SetCurrentView("c")
	// Next from "c" wraps to "a".
	assert.Equal(t, "a", mgr.NextView())
}

func TestViewManagerPreviousViewCyclesCircularly(t *testing.T) {
	mgr := views.NewViewManager()
	mgr.Register(&stubView{BaseView: views.NewBaseView("a", "A")})
	mgr.Register(&stubView{BaseView: views.NewBaseView("b", "B")})
	mgr.Register(&stubView{BaseView: views.NewBaseView("c", "C")})

	// Current is "a", previous wraps to "c".
	assert.Equal(t, "c", mgr.PreviousView())

	_ = mgr.SetCurrentView("b")
	assert.Equal(t, "a", mgr.PreviousView())
}

func TestViewManagerNextViewEmpty(t *testing.T) {
	mgr := views.NewViewManager()
	assert.Equal(t, "", mgr.NextView())
}

func TestViewManagerPreviousViewEmpty(t *testing.T) {
	mgr := views.NewViewManager()
	assert.Equal(t, "", mgr.PreviousView())
}

func TestViewManagerViewNames(t *testing.T) {
	mgr := views.NewViewManager()
	mgr.Register(&stubView{BaseView: views.NewBaseView("nodes", "Nodes")})
	mgr.Register(&stubView{BaseView: views.NewBaseView("images", "Images")})
	mgr.Register(&stubView{BaseView: views.NewBaseView("overlays", "Overlays")})

	names := mgr.ViewNames()
	assert.Equal(t, []string{"nodes", "images", "overlays"}, names)
}

func TestViewManagerViews(t *testing.T) {
	mgr := views.NewViewManager()
	v1 := &stubView{BaseView: views.NewBaseView("a", "A")}
	v2 := &stubView{BaseView: views.NewBaseView("b", "B")}
	mgr.Register(v1)
	mgr.Register(v2)

	vv := mgr.Views()
	require.Len(t, vv, 2)
	assert.Equal(t, "a", vv[0].Name())
	assert.Equal(t, "b", vv[1].Name())
}

func TestViewManagerCurrentView(t *testing.T) {
	mgr := views.NewViewManager()
	v1 := &stubView{BaseView: views.NewBaseView("nodes", "Nodes")}
	mgr.Register(v1)

	cur := mgr.CurrentView()
	require.NotNil(t, cur)
	assert.Equal(t, "nodes", cur.Name())
}

// stubView implements the View interface for testing.
type stubView struct {
	views.BaseView
}
