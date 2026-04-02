package views

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testItem is a simple item type for testing ResourceView filtering.
type testItem struct {
	Role   string
	Status string
}

func newTestResourceView(items map[string]*testItem, filter func(string, *testItem, string) bool) *ResourceView[*testItem] {
	cfg := ResourceViewConfig[*testItem]{
		Fetch: func() (map[string]*testItem, error) { return items, nil },
		Columns: []Column[*testItem]{
			{Name: "Name", Width: 20, Extract: func(name string, _ *testItem) string { return name }},
			{Name: "Role", Width: 15, Extract: func(_ string, item *testItem) string { return item.Role }},
			{Name: "Status", Width: 10, Extract: func(_ string, item *testItem) string { return item.Status }},
		},
		Filter: filter,
	}

	app := tview.NewApplication()
	rv := NewResourceView[*testItem]("test", "Test", app, cfg)
	// Manually set items and sortedKeys (skip Init/Refresh to avoid tview setup).
	rv.items = items
	rv.sortedKeys = sortKeys(items)
	return rv
}

func TestFuzzyFilter_MatchPartial(t *testing.T) {
	items := map[string]*testItem{
		"gpu-node-01": {Role: "compute", Status: "ready"},
		"cpu-node-02": {Role: "compute", Status: "ready"},
		"login-01":    {Role: "login", Status: "ready"},
	}
	rv := newTestResourceView(items, nil)
	rv.filterQuery = "gp"

	keys := rv.filteredKeys()
	assert.Contains(t, keys, "gpu-node-01", "fuzzy match should find 'gpu-node-01' for query 'gp'")
}

func TestFuzzyFilter_CaseInsensitive(t *testing.T) {
	items := map[string]*testItem{
		"GPU-Node-01": {Role: "Compute", Status: "Ready"},
		"cpu-node-02": {Role: "compute", Status: "ready"},
	}
	rv := newTestResourceView(items, nil)
	rv.filterQuery = "gpu"

	keys := rv.filteredKeys()
	assert.Contains(t, keys, "GPU-Node-01", "fuzzy match should be case insensitive")
}

func TestFuzzyFilter_EmptyQuery(t *testing.T) {
	items := map[string]*testItem{
		"node-01": {Role: "compute", Status: "ready"},
		"node-02": {Role: "login", Status: "ready"},
	}
	rv := newTestResourceView(items, nil)
	rv.filterQuery = ""

	keys := rv.filteredKeys()
	assert.Len(t, keys, 2, "empty query should return all items")
}

func TestFuzzyFilter_NoMatches(t *testing.T) {
	items := map[string]*testItem{
		"node-01": {Role: "compute", Status: "ready"},
		"node-02": {Role: "login", Status: "ready"},
	}
	rv := newTestResourceView(items, nil)
	rv.filterQuery = "zzzzxxx"

	keys := rv.filteredKeys()
	assert.Empty(t, keys, "nonsense query should return no matches")
}

func TestFuzzyFilter_CustomFilterTakesPrecedence(t *testing.T) {
	items := map[string]*testItem{
		"node-01": {Role: "compute", Status: "ready"},
		"node-02": {Role: "login", Status: "ready"},
		"node-03": {Role: "storage", Status: "down"},
	}

	// Custom filter: only match items where Role contains query.
	customFilter := func(name string, item *testItem, query string) bool {
		return strings.Contains(strings.ToLower(item.Role), query)
	}
	rv := newTestResourceView(items, customFilter)
	rv.filterQuery = "login"

	keys := rv.filteredKeys()
	require.Len(t, keys, 1, "custom filter should find exactly 1 match")
	assert.Equal(t, "node-02", keys[0])
}

func TestFuzzyFilter_MatchesColumnValues(t *testing.T) {
	items := map[string]*testItem{
		"node-01": {Role: "compute", Status: "ready"},
		"node-02": {Role: "login", Status: "down"},
	}
	rv := newTestResourceView(items, nil)
	// Match on Status column value.
	rv.filterQuery = "down"

	keys := rv.filteredKeys()
	assert.Contains(t, keys, "node-02", "fuzzy match should search across column values")
}

// Ensure unused import of tcell doesn't cause issues — it's used for Column.Color type.
var _ tcell.Color
