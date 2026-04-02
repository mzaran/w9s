package mock_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/dao"
	"github.com/mzaran/w9s/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Client Creation ----------

func TestNewMockClientNotNil(t *testing.T) {
	c := mock.NewFastMockClient()
	require.NotNil(t, c)
}

func TestNewMockClientImplementsInterface(t *testing.T) {
	var _ dao.WarewulfClient = mock.NewFastMockClient()
}

// ---------- Server Info ----------

func TestMockServerInfo(t *testing.T) {
	c := mock.NewFastMockClient()
	info, err := c.ServerInfo()
	require.NoError(t, err)
	assert.Contains(t, info.Version, "mock")
}

// ---------- HasPower ----------

func TestMockHasPower(t *testing.T) {
	c := mock.NewFastMockClient()
	assert.True(t, c.HasPower())
}

// ---------- Close ----------

func TestMockClose(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Close()
	assert.NoError(t, err)
}

// ---------- Node Manager ----------

func TestMockNodeList(t *testing.T) {
	c := mock.NewFastMockClient()
	nodes, err := c.Nodes().List()
	require.NoError(t, err)
	assert.NotEmpty(t, nodes, "sample data should populate nodes")

	// Verify we have compute and GPU nodes.
	assert.Contains(t, nodes, "compute-01")
	assert.Contains(t, nodes, "gpu-01")
}

func TestMockNodeGet(t *testing.T) {
	c := mock.NewFastMockClient()
	node, err := c.Nodes().Get("compute-01")
	require.NoError(t, err)
	assert.Equal(t, "rocky9", node.ImageName)
	assert.Contains(t, node.Profiles, "default")
}

func TestMockNodeGetNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	_, err := c.Nodes().Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMockNodeGetRaw(t *testing.T) {
	c := mock.NewFastMockClient()
	node, err := c.Nodes().GetRaw("compute-01")
	require.NoError(t, err)
	assert.NotNil(t, node)
}

func TestMockNodeGetFields(t *testing.T) {
	c := mock.NewFastMockClient()
	fields, err := c.Nodes().GetFields("compute-01")
	require.NoError(t, err)
	assert.NotEmpty(t, fields)
}

func TestMockNodeGetFieldsNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	_, err := c.Nodes().GetFields("nonexistent")
	require.Error(t, err)
}

func TestMockNodeGetOverlayInfo(t *testing.T) {
	c := mock.NewFastMockClient()
	info, err := c.Nodes().GetOverlayInfo("compute-01")
	require.NoError(t, err)
	assert.NotNil(t, info.SystemOverlay)
	assert.NotEmpty(t, info.SystemOverlay.Overlays)
	assert.NotNil(t, info.RuntimeOverlay)
	assert.NotEmpty(t, info.RuntimeOverlay.Overlays)
}

func TestMockNodeBuildOverlays(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Nodes().BuildOverlays("compute-01")
	assert.NoError(t, err)
}

func TestMockNodeBuildAllOverlays(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Nodes().BuildAllOverlays()
	assert.NoError(t, err)
}

func TestMockNodeAddThenList(t *testing.T) {
	c := mock.NewFastMockClient()

	before, err := c.Nodes().List()
	require.NoError(t, err)
	initialCount := len(before)

	newNode := &dao.WwNode{
		WwProfile: dao.WwProfile{
			ImageName:   "rocky9",
			Profiles:    []string{"default"},
			ClusterName: "test",
		},
	}
	err = c.Nodes().Add("test-node", newNode)
	require.NoError(t, err)

	after, err := c.Nodes().List()
	require.NoError(t, err)
	assert.Len(t, after, initialCount+1)
	assert.Contains(t, after, "test-node")

	got, err := c.Nodes().Get("test-node")
	require.NoError(t, err)
	assert.Equal(t, "rocky9", got.ImageName)
}

func TestMockNodeUpdate(t *testing.T) {
	c := mock.NewFastMockClient()

	updated := &dao.WwNode{
		WwProfile: dao.WwProfile{ImageName: "updated-image"},
	}
	err := c.Nodes().Update("compute-01", updated)
	require.NoError(t, err)

	got, err := c.Nodes().Get("compute-01")
	require.NoError(t, err)
	assert.Equal(t, "updated-image", got.ImageName)
}

func TestMockNodeUpdateNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Nodes().Update("nonexistent", &dao.WwNode{})
	require.Error(t, err)
}

func TestMockNodeDeleteThenVerifyGone(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Nodes().Delete("compute-01")
	require.NoError(t, err)

	_, err = c.Nodes().Get("compute-01")
	require.Error(t, err)

	nodes, err := c.Nodes().List()
	require.NoError(t, err)
	assert.NotContains(t, nodes, "compute-01")
}

func TestMockNodeDeleteNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Nodes().Delete("nonexistent")
	require.Error(t, err)
}

// ---------- Profile Manager ----------

func TestMockProfileList(t *testing.T) {
	c := mock.NewFastMockClient()
	profiles, err := c.Profiles().List()
	require.NoError(t, err)
	assert.NotEmpty(t, profiles)
	assert.Contains(t, profiles, "default")
	assert.Contains(t, profiles, "gpu")
}

func TestMockProfileGet(t *testing.T) {
	c := mock.NewFastMockClient()
	profile, err := c.Profiles().Get("default")
	require.NoError(t, err)
	assert.Equal(t, "mycluster", profile.ClusterName)
	assert.Equal(t, "rocky9", profile.ImageName)
}

func TestMockProfileGetNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	_, err := c.Profiles().Get("nonexistent")
	require.Error(t, err)
}

func TestMockProfileAddThenList(t *testing.T) {
	c := mock.NewFastMockClient()

	before, err := c.Profiles().List()
	require.NoError(t, err)
	initialCount := len(before)

	err = c.Profiles().Add("test-profile", &dao.WwProfile{
		Comment:   "test profile",
		ImageName: "ubuntu22",
	})
	require.NoError(t, err)

	after, err := c.Profiles().List()
	require.NoError(t, err)
	assert.Len(t, after, initialCount+1)
	assert.Contains(t, after, "test-profile")
}

func TestMockProfileUpdate(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Profiles().Update("default", &dao.WwProfile{Comment: "updated"})
	require.NoError(t, err)

	got, err := c.Profiles().Get("default")
	require.NoError(t, err)
	assert.Equal(t, "updated", got.Comment)
}

func TestMockProfileUpdateNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Profiles().Update("nonexistent", &dao.WwProfile{})
	require.Error(t, err)
}

func TestMockProfileDeleteThenVerifyGone(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Profiles().Delete("gpu")
	require.NoError(t, err)

	_, err = c.Profiles().Get("gpu")
	require.Error(t, err)
}

func TestMockProfileDeleteNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Profiles().Delete("nonexistent")
	require.Error(t, err)
}

// ---------- Image Manager ----------

func TestMockImageList(t *testing.T) {
	c := mock.NewFastMockClient()
	images, err := c.Images().List()
	require.NoError(t, err)
	assert.NotEmpty(t, images)
	assert.Contains(t, images, "rocky9")
	assert.Contains(t, images, "rocky9-cuda")
}

func TestMockImageGet(t *testing.T) {
	c := mock.NewFastMockClient()
	img, err := c.Images().Get("rocky9")
	require.NoError(t, err)
	assert.True(t, img.Size > 0)
	assert.NotEmpty(t, img.Kernels)
}

func TestMockImageGetNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	_, err := c.Images().Get("nonexistent")
	require.Error(t, err)
}

func TestMockImageBuild(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Images().Build("rocky9")
	assert.NoError(t, err)
}

func TestMockImageImport(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Images().Import("newimg", "docker://ubuntu:22.04")
	require.NoError(t, err)

	img, err := c.Images().Get("newimg")
	require.NoError(t, err)
	assert.True(t, img.Size > 0)
}

func TestMockImageUpdate(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Images().Update("rocky9", "rocky9-renamed")
	require.NoError(t, err)

	_, err = c.Images().Get("rocky9")
	require.Error(t, err, "old name should be gone")

	img, err := c.Images().Get("rocky9-renamed")
	require.NoError(t, err)
	assert.True(t, img.Size > 0)
}

func TestMockImageUpdateNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Images().Update("nonexistent", "newname")
	require.Error(t, err)
}

func TestMockImageDeleteThenVerifyGone(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Images().Delete("rocky9-cuda")
	require.NoError(t, err)

	_, err = c.Images().Get("rocky9-cuda")
	require.Error(t, err)

	images, err := c.Images().List()
	require.NoError(t, err)
	assert.NotContains(t, images, "rocky9-cuda")
}

func TestMockImageDeleteNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Images().Delete("nonexistent")
	require.Error(t, err)
}

// ---------- Overlay Manager ----------

func TestMockOverlayList(t *testing.T) {
	c := mock.NewFastMockClient()
	overlays, err := c.Overlays().List()
	require.NoError(t, err)
	assert.NotEmpty(t, overlays)
	assert.Contains(t, overlays, "wwinit")
	assert.Contains(t, overlays, "generic")
	assert.Contains(t, overlays, "runtime")
}

func TestMockOverlayGet(t *testing.T) {
	c := mock.NewFastMockClient()
	overlay, err := c.Overlays().Get("wwinit")
	require.NoError(t, err)
	assert.NotEmpty(t, overlay.Files)
}

func TestMockOverlayGetNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	_, err := c.Overlays().Get("nonexistent")
	require.Error(t, err)
}

func TestMockOverlayGetFile(t *testing.T) {
	c := mock.NewFastMockClient()
	file, err := c.Overlays().GetFile("wwinit", "/etc/hostname", "node1")
	require.NoError(t, err)
	assert.Equal(t, "wwinit", file.Overlay)
	assert.Equal(t, "/etc/hostname", file.Path)
	assert.NotEmpty(t, file.Contents)
}

func TestMockOverlayGetFileNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	_, err := c.Overlays().GetFile("nonexistent", "/etc/hostname", "")
	require.Error(t, err)
}

func TestMockOverlayCreate(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Overlays().Create("custom")
	require.NoError(t, err)

	overlay, err := c.Overlays().Get("custom")
	require.NoError(t, err)
	assert.NotNil(t, overlay)
}

func TestMockOverlayAddFile(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Overlays().AddFile("wwinit", "/etc/motd", "Welcome\n")
	require.NoError(t, err)

	overlay, err := c.Overlays().Get("wwinit")
	require.NoError(t, err)
	assert.Contains(t, overlay.Files, "/etc/motd")
}

func TestMockOverlayAddFileNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Overlays().AddFile("nonexistent", "/etc/motd", "test")
	require.Error(t, err)
}

func TestMockOverlayDeleteThenVerifyGone(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Overlays().Delete("chrony", false)
	require.NoError(t, err)

	_, err = c.Overlays().Get("chrony")
	require.Error(t, err)

	overlays, err := c.Overlays().List()
	require.NoError(t, err)
	assert.NotContains(t, overlays, "chrony")
}

func TestMockOverlayDeleteNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Overlays().Delete("nonexistent", false)
	require.Error(t, err)
}

func TestMockOverlayDeleteFile(t *testing.T) {
	c := mock.NewFastMockClient()

	overlay, err := c.Overlays().Get("wwinit")
	require.NoError(t, err)
	initialCount := len(overlay.Files)
	require.True(t, initialCount > 0)

	err = c.Overlays().DeleteFile("wwinit", overlay.Files[0])
	require.NoError(t, err)

	overlay, err = c.Overlays().Get("wwinit")
	require.NoError(t, err)
	assert.Len(t, overlay.Files, initialCount-1)
}

func TestMockOverlayDeleteFileNotFound(t *testing.T) {
	c := mock.NewFastMockClient()
	err := c.Overlays().DeleteFile("nonexistent", "/etc/foo")
	require.Error(t, err)
}

// ---------- Power Manager ----------

func TestMockPowerStatus(t *testing.T) {
	c := mock.NewFastMockClient()

	status, err := c.Power().Status("compute-01")
	require.NoError(t, err)
	assert.Equal(t, "on", status)

	status, err = c.Power().Status("compute-07")
	require.NoError(t, err)
	assert.Equal(t, "off", status)
}

func TestMockPowerStatusDefault(t *testing.T) {
	c := mock.NewFastMockClient()

	// Unknown node should default to "on".
	status, err := c.Power().Status("unknown-node")
	require.NoError(t, err)
	assert.Equal(t, "on", status)
}

func TestMockPowerOn(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Power().On("compute-07")
	require.NoError(t, err)

	status, err := c.Power().Status("compute-07")
	require.NoError(t, err)
	assert.Equal(t, "on", status)
}

func TestMockPowerOff(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Power().Off("compute-01")
	require.NoError(t, err)

	status, err := c.Power().Status("compute-01")
	require.NoError(t, err)
	assert.Equal(t, "off", status)
}

func TestMockPowerCycle(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Power().Cycle("compute-01")
	require.NoError(t, err)

	status, err := c.Power().Status("compute-01")
	require.NoError(t, err)
	assert.Equal(t, "on", status)
}

func TestMockPowerReset(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Power().Reset("compute-01")
	require.NoError(t, err)

	status, err := c.Power().Status("compute-01")
	require.NoError(t, err)
	assert.Equal(t, "on", status)
}

// ---------- Sample Data Verification ----------

func TestSampleDataNodeProfiles(t *testing.T) {
	c := mock.NewFastMockClient()

	node, err := c.Nodes().Get("gpu-01")
	require.NoError(t, err)
	assert.Contains(t, node.Profiles, "gpu")
	assert.Equal(t, "rocky9-cuda", node.ImageName)
}

func TestSampleDataNodeNetworkDevices(t *testing.T) {
	c := mock.NewFastMockClient()

	node, err := c.Nodes().Get("compute-01")
	require.NoError(t, err)
	require.Contains(t, node.NetDevs, "eth0")

	eth0 := node.NetDevs["eth0"]
	assert.Equal(t, "10.0.0.1", eth0.Ipaddr)
	assert.Equal(t, "255.255.255.0", eth0.Netmask)
	assert.NotEmpty(t, eth0.Hwaddr)
	assert.True(t, eth0.OnBoot.IsTrue())
}

func TestSampleDataNodeIPMI(t *testing.T) {
	c := mock.NewFastMockClient()

	node, err := c.Nodes().Get("compute-01")
	require.NoError(t, err)
	require.NotNil(t, node.Ipmi)
	assert.Equal(t, "10.0.1.1", node.Ipmi.Ipaddr)
	assert.Equal(t, "admin", node.Ipmi.UserName)
}

func TestSampleDataProfileKernel(t *testing.T) {
	c := mock.NewFastMockClient()

	profile, err := c.Profiles().Get("default")
	require.NoError(t, err)
	require.NotNil(t, profile.Kernel)
	assert.NotEmpty(t, profile.Kernel.Version)
}

func TestSampleDataImageKernels(t *testing.T) {
	c := mock.NewFastMockClient()

	img, err := c.Images().Get("rocky9")
	require.NoError(t, err)
	assert.NotEmpty(t, img.Kernels)
	assert.True(t, img.Size > 0)
}

func TestSampleDataOverlayFiles(t *testing.T) {
	c := mock.NewFastMockClient()

	overlay, err := c.Overlays().Get("wwinit")
	require.NoError(t, err)
	assert.Contains(t, overlay.Files, "/etc/hostname")
	assert.Contains(t, overlay.Files, "/etc/hosts")
}

// ---------- Full CRUD Cycle ----------

func TestFullNodeCRUDCycle(t *testing.T) {
	c := mock.NewFastMockClient()

	// Create.
	newNode := &dao.WwNode{
		Discoverable: dao.WWBoolFalse,
		WwProfile: dao.WwProfile{
			ImageName:   "rocky9",
			Profiles:    []string{"default"},
			ClusterName: "test",
		},
	}
	err := c.Nodes().Add("crud-test", newNode)
	require.NoError(t, err)

	// Read.
	got, err := c.Nodes().Get("crud-test")
	require.NoError(t, err)
	assert.Equal(t, "rocky9", got.ImageName)

	// Update.
	got.ImageName = "ubuntu22"
	err = c.Nodes().Update("crud-test", got)
	require.NoError(t, err)

	updated, err := c.Nodes().Get("crud-test")
	require.NoError(t, err)
	assert.Equal(t, "ubuntu22", updated.ImageName)

	// Delete.
	err = c.Nodes().Delete("crud-test")
	require.NoError(t, err)

	_, err = c.Nodes().Get("crud-test")
	require.Error(t, err)
}

func TestFullProfileCRUDCycle(t *testing.T) {
	c := mock.NewFastMockClient()

	err := c.Profiles().Add("crud-test", &dao.WwProfile{
		Comment:   "test",
		ImageName: "rocky9",
	})
	require.NoError(t, err)

	got, err := c.Profiles().Get("crud-test")
	require.NoError(t, err)
	assert.Equal(t, "test", got.Comment)

	err = c.Profiles().Update("crud-test", &dao.WwProfile{Comment: "updated"})
	require.NoError(t, err)

	got, err = c.Profiles().Get("crud-test")
	require.NoError(t, err)
	assert.Equal(t, "updated", got.Comment)

	err = c.Profiles().Delete("crud-test")
	require.NoError(t, err)

	_, err = c.Profiles().Get("crud-test")
	require.Error(t, err)
}
