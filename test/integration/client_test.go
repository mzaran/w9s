package integration_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mzaran/w9s/internal/dao"
)

// testEndpoint returns the warewulfd endpoint from env, or skips.
func testEndpoint(t *testing.T) string {
	t.Helper()
	if os.Getenv("W9S_INTEGRATION_TESTS") != "1" {
		t.Skip("set W9S_INTEGRATION_TESTS=1 to run integration tests")
	}
	ep := os.Getenv("W9S_TEST_ENDPOINT")
	if ep == "" {
		ep = "http://192.168.1.10:9873"
	}
	return ep
}

func testClient(t *testing.T) *dao.HTTPClient {
	t.Helper()
	ep := testEndpoint(t)
	user := os.Getenv("W9S_TEST_USERNAME")
	if user == "" {
		user = "admin"
	}
	pass := os.Getenv("W9S_TEST_PASSWORD")
	if pass == "" {
		pass = "w9stest"
	}
	c, err := dao.NewClientFromConfig(ep, user, pass, true, 10*time.Second)
	require.NoError(t, err)
	return c
}

func TestListProfiles(t *testing.T) {
	c := testClient(t)
	defer c.Close()

	profiles, err := c.Profiles().List()
	require.NoError(t, err)
	assert.Contains(t, profiles, "default", "should have default profile")

	def := profiles["default"]
	assert.NotEmpty(t, def.SystemOverlay, "default profile should have system overlays")
	assert.NotEmpty(t, def.RuntimeOverlay, "default profile should have runtime overlays")
}

func TestListOverlays(t *testing.T) {
	c := testClient(t)
	defer c.Close()

	overlays, err := c.Overlays().List()
	require.NoError(t, err)
	assert.Greater(t, len(overlays), 5, "should have multiple stock overlays")

	// Check a known overlay.
	wwinit, ok := overlays["wwinit"]
	assert.True(t, ok, "should have wwinit overlay")
	if ok {
		assert.NotEmpty(t, wwinit.Files, "wwinit should have files")
	}
}

func TestListNodes(t *testing.T) {
	c := testClient(t)
	defer c.Close()

	nodes, err := c.Nodes().List()
	require.NoError(t, err)

	// Check nodes exist (testnode-01, testnode-02 created in setup).
	if len(nodes) == 0 {
		t.Log("no nodes found — create test nodes with: wwctl node add testnode-01 --ipaddr 192.168.1.101 --hwaddr 00:11:22:33:44:01")
		return
	}

	// Verify JSON deserialization with space-in-key fields.
	for name, node := range nodes {
		t.Logf("node %s: profiles=%v, image=%q, cluster=%q", name, node.Profiles, node.ImageName, node.ClusterName)
		assert.NotEmpty(t, node.Profiles, "node %s should have profiles", name)
	}
}

func TestGetNode(t *testing.T) {
	c := testClient(t)
	defer c.Close()

	nodes, err := c.Nodes().List()
	require.NoError(t, err)
	if len(nodes) == 0 {
		t.Skip("no nodes to test Get on")
	}

	// Get first node.
	var firstName string
	for name := range nodes {
		firstName = name
		break
	}

	node, err := c.Nodes().Get(firstName)
	require.NoError(t, err)
	assert.NotNil(t, node)
	assert.NotEmpty(t, node.Profiles)
}

func TestNodeCRUDRoundtrip(t *testing.T) {
	c := testClient(t)
	defer c.Close()

	testName := "w9s-integration-test-node"

	// Cleanup in case of previous failed test.
	_ = c.Nodes().Delete(testName)

	// Add.
	err := c.Nodes().Add(testName, &dao.WwNode{})
	require.NoError(t, err, "add node should succeed")

	// Verify present.
	nodes, err := c.Nodes().List()
	require.NoError(t, err)
	assert.Contains(t, nodes, testName, "added node should appear in list")

	// Delete.
	err = c.Nodes().Delete(testName)
	require.NoError(t, err, "delete node should succeed")

	// Verify gone.
	nodes, err = c.Nodes().List()
	require.NoError(t, err)
	assert.NotContains(t, nodes, testName, "deleted node should not appear in list")
}

func TestOverlayFiles(t *testing.T) {
	c := testClient(t)
	defer c.Close()

	overlays, err := c.Overlays().List()
	require.NoError(t, err)

	// Find an overlay with files.
	for name, ov := range overlays {
		if len(ov.Files) > 0 {
			t.Logf("overlay %s has %d files", name, len(ov.Files))
			// Try getting the first file.
			file, err := c.Overlays().GetFile(name, ov.Files[0], "")
			if err != nil {
				t.Logf("GetFile(%s, %s): %v (may be expected for binary files)", name, ov.Files[0], err)
			} else {
				assert.NotNil(t, file)
				t.Logf("file contents length: %d", len(file.Contents))
			}
			return
		}
	}
}

func TestConnectionError(t *testing.T) {
	if os.Getenv("W9S_INTEGRATION_TESTS") != "1" {
		t.Skip("set W9S_INTEGRATION_TESTS=1")
	}

	// Wrong endpoint should fail.
	c, err := dao.NewClientFromConfig("http://192.168.1.10:19999", "", "", true, 3*time.Second)
	require.NoError(t, err)
	defer c.Close()

	_, err = c.Nodes().List()
	assert.Error(t, err, "wrong port should cause connection error")
}
