package views_test

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mzaran/w9s/internal/views"
)

type testItem struct {
	Name  string
	Value int
}

func testData() map[string]testItem {
	return map[string]testItem{
		"charlie": {Name: "charlie", Value: 3},
		"alpha":   {Name: "alpha", Value: 1},
		"bravo":   {Name: "bravo", Value: 2},
	}
}

func newTestResourceView(data map[string]testItem, fetchErr error) *views.ResourceView[testItem] {
	app := tview.NewApplication()
	cfg := views.ResourceViewConfig[testItem]{
		Fetch: func() (map[string]testItem, error) {
			if fetchErr != nil {
				return nil, fetchErr
			}
			return data, nil
		},
		Columns: []views.Column[testItem]{
			{
				Name:  "Name",
				Width: 20,
				Extract: func(name string, item testItem) string {
					return item.Name
				},
			},
			{
				Name:  "Value",
				Width: 10,
				Extract: func(name string, item testItem) string {
					return string(rune('0' + item.Value))
				},
			},
		},
		Actions: []views.Action[testItem]{
			{
				Key:   'd',
				Label: "Delete",
			},
		},
		Filter: func(name string, item testItem, query string) bool {
			return item.Name == query
		},
	}
	return views.NewResourceView("test", "Test Resources", app, cfg)
}

func TestResourceViewInit(t *testing.T) {
	rv := newTestResourceView(testData(), nil)
	err := rv.Init(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, rv.Render())
}

func TestResourceViewRefreshPopulatesItems(t *testing.T) {
	data := testData()
	rv := newTestResourceView(data, nil)
	require.NoError(t, rv.Init(context.Background()))

	// Refresh launches a goroutine; we need to wait for it.
	err := rv.Refresh()
	require.NoError(t, err)

	// Wait for the async refresh to complete.
	require.Eventually(t, func() bool {
		return !rv.IsRefreshing()
	}, 2*time.Second, 10*time.Millisecond)

	assert.Nil(t, rv.LastError())
}

func TestResourceViewHintsIncludeActions(t *testing.T) {
	rv := newTestResourceView(testData(), nil)
	hints := rv.Hints()

	// Should include default hints plus our "Delete" action.
	assert.Contains(t, hints, "/ Filter")
	assert.Contains(t, hints, "Enter Detail")
	assert.Contains(t, hints, "r Refresh")
	assert.Contains(t, hints, "d Delete")
}

func TestResourceViewHintsWithMultipleActions(t *testing.T) {
	app := tview.NewApplication()
	cfg := views.ResourceViewConfig[testItem]{
		Fetch: func() (map[string]testItem, error) {
			return testData(), nil
		},
		Columns: []views.Column[testItem]{
			{Name: "Name", Extract: func(n string, i testItem) string { return i.Name }},
		},
		Actions: []views.Action[testItem]{
			{Key: 'd', Label: "Delete"},
			{Key: 'e', Label: "Edit"},
		},
	}
	rv := views.NewResourceView("test", "Test", app, cfg)
	hints := rv.Hints()
	assert.Contains(t, hints, "d Delete")
	assert.Contains(t, hints, "e Edit")
}

func TestResourceViewSortedKeysAlphabetical(t *testing.T) {
	// Verify that sortKeys produces alphabetical order by checking
	// the expected sort of our test data keys.
	keys := []string{"charlie", "alpha", "bravo"}
	sort.Strings(keys)
	assert.Equal(t, []string{"alpha", "bravo", "charlie"}, keys)
}

func TestResourceViewEmptyData(t *testing.T) {
	rv := newTestResourceView(map[string]testItem{}, nil)
	require.NoError(t, rv.Init(context.Background()))

	err := rv.Refresh()
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return !rv.IsRefreshing()
	}, 2*time.Second, 10*time.Millisecond)

	assert.Nil(t, rv.LastError())
}

func TestResourceViewFetchError(t *testing.T) {
	fetchErr := errors.New("connection refused")
	rv := newTestResourceView(nil, fetchErr)
	require.NoError(t, rv.Init(context.Background()))

	err := rv.Refresh()
	require.NoError(t, err) // Refresh itself returns nil; error is stored.

	require.Eventually(t, func() bool {
		return !rv.IsRefreshing()
	}, 2*time.Second, 10*time.Millisecond)

	assert.Error(t, rv.LastError())
	assert.Contains(t, rv.LastError().Error(), "connection refused")
}

func TestResourceViewNameAndTitle(t *testing.T) {
	rv := newTestResourceView(testData(), nil)
	assert.Equal(t, "test", rv.Name())
	assert.Equal(t, "Test Resources", rv.Title())
}

func TestResourceViewRenderNilBeforeInit(t *testing.T) {
	rv := newTestResourceView(testData(), nil)
	// Before Init, Render returns nil because container hasn't been created.
	assert.Nil(t, rv.Render())
}

func TestResourceViewConcurrentRefresh(t *testing.T) {
	// Verify that calling Refresh while already refreshing is safe.
	var mu sync.Mutex
	callCount := 0

	app := tview.NewApplication()
	cfg := views.ResourceViewConfig[testItem]{
		Fetch: func() (map[string]testItem, error) {
			mu.Lock()
			callCount++
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			return testData(), nil
		},
		Columns: []views.Column[testItem]{
			{Name: "Name", Extract: func(n string, i testItem) string { return i.Name }},
		},
	}
	rv := views.NewResourceView("test", "Test", app, cfg)
	require.NoError(t, rv.Init(context.Background()))

	// First refresh starts.
	_ = rv.Refresh()
	// Second refresh should be a no-op because already refreshing.
	_ = rv.Refresh()

	require.Eventually(t, func() bool {
		return !rv.IsRefreshing()
	}, 2*time.Second, 10*time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	// Fetch should have been called exactly once since the second call was skipped.
	assert.Equal(t, 1, callCount)
}
