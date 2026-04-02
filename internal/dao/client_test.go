package dao_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mzaran/w9s/internal/dao"
	"github.com/mzaran/w9s/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cannedNodesJSON is realistic Warewulf API output with space-in-key field names.
const cannedNodesJSON = `{
  "node1": {
    "discoverable": "false",
    "asset key": "",
    "profiles": ["default"],
    "cluster name": "prod",
    "image name": "rocky9",
    "runtime overlay": ["runtime"],
    "system overlay": ["wwinit", "generic"],
    "network devices": {
      "eth0": {
        "ipaddr": "10.0.0.1",
        "netmask": "255.255.255.0",
        "hwaddr": "00:11:22:33:44:01"
      }
    }
  },
  "node2": {
    "discoverable": "true",
    "profiles": ["default", "gpu"],
    "cluster name": "prod",
    "image name": "rocky9-cuda",
    "runtime overlay": ["runtime"],
    "system overlay": ["wwinit"],
    "network devices": {
      "eth0": {
        "ipaddr": "10.0.0.2",
        "netmask": "255.255.255.0",
        "hwaddr": "00:11:22:33:44:02"
      }
    }
  }
}`

// GET /api/nodes/{id} returns a single node object (not wrapped in a map).
const cannedSingleNodeJSON = `{
  "discoverable": "false",
  "profiles": ["default"],
  "cluster name": "prod",
  "image name": "rocky9",
  "runtime overlay": ["runtime"],
  "system overlay": ["wwinit"],
  "network devices": {
    "eth0": {
      "ipaddr": "10.0.0.1",
      "netmask": "255.255.255.0",
      "hwaddr": "00:11:22:33:44:01"
    }
  }
}`

const cannedProfilesJSON = `{
  "default": {
    "comment": "Default profile",
    "cluster name": "mycluster",
    "image name": "rocky9",
    "runtime overlay": ["generic", "runtime"],
    "system overlay": ["wwinit"]
  }
}`

const cannedImagesJSON = `{
  "rocky9": {
    "kernels": ["5.14.0-362.el9.x86_64"],
    "size": 2254857830,
    "buildtime": 1700000000,
    "writable": false
  }
}`

// GET /api/profiles/{id} returns a single profile object.
const cannedSingleProfileJSON = `{
  "comment": "Default profile",
  "cluster name": "mycluster",
  "image name": "rocky9",
  "runtime overlay": ["generic", "runtime"],
  "system overlay": ["wwinit"]
}`

// GET /api/images/{name} returns a single image object.
const cannedSingleImageJSON = `{
  "kernels": ["5.14.0-362.el9.x86_64"],
  "size": 2254857830,
  "buildtime": 1700000000,
  "writable": false
}`

const cannedOverlaysJSON = `{
  "wwinit": {
    "files": ["/etc/hostname", "/etc/hosts"],
    "site": false
  }
}`

const cannedNodeFieldsJSON = `[
  {"field": "ImageName", "value": "rocky9", "source": "node"},
  {"field": "ClusterName", "value": "prod", "source": "profile:default"}
]`

const cannedNodeOverlayInfoJSON = `{
  "system overlay": {"overlays": ["wwinit"], "mtime": "2023-11-14T22:13:20Z"},
  "runtime overlay": {"overlays": ["runtime"], "mtime": "2023-11-14T22:30:00Z"}
}`

const cannedOverlayFileJSON = `{
  "overlay": "wwinit",
  "path": "/etc/hostname",
  "contents": "node1.local\n",
  "perms": 420,
  "uid": 0,
  "gid": 0
}`

// newTestServer creates a httptest.Server that routes requests to the appropriate
// canned JSON responses, similar to the Warewulf REST API.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Nodes
	mux.HandleFunc("/api/nodes/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
		path = strings.TrimSuffix(path, "/")

		switch {
		case path == "" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedNodesJSON))

		case path == "overlays/build" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)

		case strings.HasSuffix(path, "/overlays/build") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)

		case strings.HasSuffix(path, "/fields") && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedNodeFieldsJSON))

		case strings.HasSuffix(path, "/overlays") && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedNodeOverlayInfoJSON))

		case strings.HasSuffix(path, "/raw") && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedSingleNodeJSON))

		case path != "" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedSingleNodeJSON))

		case path != "" && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)

		case path != "" && r.Method == http.MethodPatch:
			w.WriteHeader(http.StatusOK)

		case path != "" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Profiles
	mux.HandleFunc("/api/profiles/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
		path = strings.TrimSuffix(path, "/")

		switch {
		case path == "" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedProfilesJSON))

		case path != "" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedSingleProfileJSON))

		case path != "" && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)

		case path != "" && r.Method == http.MethodPatch:
			w.WriteHeader(http.StatusOK)

		case path != "" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Images
	mux.HandleFunc("/api/images/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/images/")
		path = strings.TrimSuffix(path, "/")

		switch {
		case path == "" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedImagesJSON))

		case strings.HasSuffix(path, "/import") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)

		case strings.HasSuffix(path, "/build") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)

		case path != "" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedSingleImageJSON))

		case path != "" && r.Method == http.MethodPatch:
			w.WriteHeader(http.StatusOK)

		case path != "" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Overlays
	mux.HandleFunc("/api/overlays/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/overlays/")
		path = strings.TrimSuffix(path, "/")

		switch {
		case path == "" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedOverlaysJSON))

		case strings.HasSuffix(path, "/file") && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedOverlayFileJSON))

		case strings.HasSuffix(path, "/file") && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)

		case strings.HasSuffix(path, "/file") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)

		case path != "" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cannedOverlaysJSON))

		case path != "" && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)

		case path != "" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Server info
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version": "4.5.8"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	return httptest.NewServer(mux)
}

func newClient(t *testing.T, serverURL string) *dao.HTTPClient {
	t.Helper()
	c, err := dao.NewHTTPClient(
		dao.WithEndpoint(serverURL),
		dao.WithBasicAuth("admin", "secret"),
		dao.WithTimeout(5*time.Second),
	)
	require.NoError(t, err)
	t.Cleanup(func() { c.Close() })
	return c
}

// ---------- Node Tests ----------

func TestNodeList(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	nodes, err := c.Nodes().List()
	require.NoError(t, err)
	assert.Len(t, nodes, 2)

	node1 := nodes["node1"]
	require.NotNil(t, node1)
	assert.Equal(t, dao.WWBoolFalse, node1.Discoverable)
	assert.Equal(t, "prod", node1.ClusterName)
	assert.Equal(t, "rocky9", node1.ImageName)
	assert.Equal(t, []string{"default"}, node1.Profiles)
	assert.Equal(t, []string{"runtime"}, node1.RuntimeOverlay)
	assert.Equal(t, []string{"wwinit", "generic"}, node1.SystemOverlay)

	// Verify network devices with space-in-key JSON field.
	require.NotNil(t, node1.NetDevs)
	eth0 := node1.NetDevs["eth0"]
	require.NotNil(t, eth0)
	assert.Equal(t, "10.0.0.1", eth0.Ipaddr)
	assert.Equal(t, "255.255.255.0", eth0.Netmask)
	assert.Equal(t, "00:11:22:33:44:01", eth0.Hwaddr)

	node2 := nodes["node2"]
	require.NotNil(t, node2)
	assert.Equal(t, dao.WWBoolTrue, node2.Discoverable)
}

func TestNodeGet(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	node, err := c.Nodes().Get("node1")
	require.NoError(t, err)
	assert.Equal(t, "prod", node.ClusterName)
	assert.Equal(t, "rocky9", node.ImageName)
}

func TestNodeGetRaw(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	node, err := c.Nodes().GetRaw("node1")
	require.NoError(t, err)
	assert.Equal(t, "rocky9", node.ImageName)
}

func TestNodeGetFields(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	fields, err := c.Nodes().GetFields("node1")
	require.NoError(t, err)
	require.Len(t, fields, 2)
	assert.Equal(t, "ImageName", fields[0].Field)
	assert.Equal(t, "rocky9", fields[0].Value)
	assert.Equal(t, "node", fields[0].Source)
}

func TestNodeGetOverlayInfo(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	info, err := c.Nodes().GetOverlayInfo("node1")
	require.NoError(t, err)
	require.NotNil(t, info.SystemOverlay)
	require.Len(t, info.SystemOverlay.Overlays, 1)
	assert.Equal(t, "wwinit", info.SystemOverlay.Overlays[0])
	require.NotNil(t, info.RuntimeOverlay)
	require.Len(t, info.RuntimeOverlay.Overlays, 1)
	assert.Equal(t, "runtime", info.RuntimeOverlay.Overlays[0])
}

func TestNodeBuildOverlays(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Nodes().BuildOverlays("node1")
	assert.NoError(t, err)
}

func TestNodeBuildAllOverlays(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Nodes().BuildAllOverlays()
	assert.NoError(t, err)
}

func TestNodeAdd(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	node := &dao.WwNode{
		WwProfile: dao.WwProfile{
			ImageName: "rocky9",
			Profiles:  []string{"default"},
		},
	}
	err := c.Nodes().Add("newnode", node)
	assert.NoError(t, err)
}

func TestNodeUpdate(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	node := &dao.WwNode{
		WwProfile: dao.WwProfile{ImageName: "rocky9-updated"},
	}
	err := c.Nodes().Update("node1", node)
	assert.NoError(t, err)
}

func TestNodeDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Nodes().Delete("node1")
	assert.NoError(t, err)
}

// ---------- Profile Tests ----------

func TestProfileList(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	profiles, err := c.Profiles().List()
	require.NoError(t, err)
	assert.Len(t, profiles, 1)

	def := profiles["default"]
	require.NotNil(t, def)
	assert.Equal(t, "Default profile", def.Comment)
	assert.Equal(t, "mycluster", def.ClusterName)
	assert.Equal(t, "rocky9", def.ImageName)
}

func TestProfileGet(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	profile, err := c.Profiles().Get("default")
	require.NoError(t, err)
	assert.Equal(t, "mycluster", profile.ClusterName)
}

func TestProfileAdd(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Profiles().Add("test", &dao.WwProfile{Comment: "test"})
	assert.NoError(t, err)
}

func TestProfileUpdate(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Profiles().Update("default", &dao.WwProfile{Comment: "updated"})
	assert.NoError(t, err)
}

func TestProfileDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Profiles().Delete("default")
	assert.NoError(t, err)
}

// ---------- Image Tests ----------

func TestImageList(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	images, err := c.Images().List()
	require.NoError(t, err)
	assert.Len(t, images, 1)

	rocky := images["rocky9"]
	require.NotNil(t, rocky)
	assert.Equal(t, int64(2254857830), rocky.Size)
	assert.Equal(t, int64(1700000000), rocky.BuildTime)
	assert.Equal(t, []string{"5.14.0-362.el9.x86_64"}, rocky.Kernels)
	assert.False(t, rocky.Writable)
}

func TestImageGet(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	img, err := c.Images().Get("rocky9")
	require.NoError(t, err)
	assert.Equal(t, int64(2254857830), img.Size)
}

func TestImageBuild(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Images().Build("rocky9")
	assert.NoError(t, err)
}

func TestImageImport(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Images().Import("newimg", "docker://rockylinux:9")
	assert.NoError(t, err)
}

func TestImageUpdate(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Images().Update("rocky9", "rocky9-renamed")
	assert.NoError(t, err)
}

func TestImageDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Images().Delete("rocky9")
	assert.NoError(t, err)
}

// ---------- Overlay Tests ----------

func TestOverlayList(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	overlays, err := c.Overlays().List()
	require.NoError(t, err)
	assert.Len(t, overlays, 1)

	wwinit := overlays["wwinit"]
	require.NotNil(t, wwinit)
	assert.Equal(t, []string{"/etc/hostname", "/etc/hosts"}, wwinit.Files)
	assert.False(t, wwinit.Site)
}

func TestOverlayGet(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	overlay, err := c.Overlays().Get("wwinit")
	require.NoError(t, err)
	assert.Equal(t, []string{"/etc/hostname", "/etc/hosts"}, overlay.Files)
}

func TestOverlayGetFile(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	file, err := c.Overlays().GetFile("wwinit", "/etc/hostname", "node1")
	require.NoError(t, err)
	assert.Equal(t, "wwinit", file.Overlay)
	assert.Equal(t, "/etc/hostname", file.Path)
	assert.Equal(t, "node1.local\n", file.Contents)
	assert.Equal(t, uint32(420), file.Perms)
}

func TestOverlayCreate(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Overlays().Create("newoverlay")
	assert.NoError(t, err)
}

func TestOverlayAddFile(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Overlays().AddFile("wwinit", "/etc/motd", "Welcome!\n")
	assert.NoError(t, err)
}

func TestOverlayDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Overlays().Delete("wwinit", false)
	assert.NoError(t, err)
}

func TestOverlayDeleteForce(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Overlays().Delete("wwinit", true)
	assert.NoError(t, err)
}

func TestOverlayDeleteFile(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Overlays().DeleteFile("wwinit", "/etc/hostname")
	assert.NoError(t, err)
}

// ---------- Server Info ----------

func TestServerInfo(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	info, err := c.ServerInfo()
	require.NoError(t, err)
	assert.Equal(t, "4.5.8", info.Version)
}

// ---------- Error Handling ----------

func TestHTTPError401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("bad credentials"))
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	_, err := c.Nodes().List()
	require.Error(t, err)
	var w9sErr *errs.W9sError
	require.True(t, errors.As(err, &w9sErr))
	assert.Equal(t, errs.ErrAuth, w9sErr.Code)
	assert.Contains(t, w9sErr.Message, "authentication failed")
}

func TestHTTPError403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("insufficient permissions"))
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	_, err := c.Nodes().List()
	require.Error(t, err)
	var w9sErr *errs.W9sError
	require.True(t, errors.As(err, &w9sErr))
	assert.Equal(t, errs.ErrForbidden, w9sErr.Code)
}

func TestHTTPError404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("no such node"))
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	_, err := c.Nodes().Get("nonexistent")
	require.Error(t, err)
	var w9sErr *errs.W9sError
	require.True(t, errors.As(err, &w9sErr))
	assert.Equal(t, errs.ErrNotFound, w9sErr.Code)
}

func TestHTTPError500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	_, err := c.Nodes().List()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestConnectionRefused(t *testing.T) {
	// Point at a port that nothing is listening on.
	c, err := dao.NewHTTPClient(
		dao.WithEndpoint("http://127.0.0.1:1"),
		dao.WithTimeout(2*time.Second),
	)
	require.NoError(t, err)
	defer c.Close()

	_, err = c.Nodes().List()
	require.Error(t, err)
	var w9sErr *errs.W9sError
	if errors.As(err, &w9sErr) {
		assert.True(t, w9sErr.Code == errs.ErrAPIDisabled || w9sErr.Code == errs.ErrConnection,
			"expected ErrAPIDisabled or ErrConnection, got %d", w9sErr.Code)
	}
}

func TestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := dao.NewHTTPClient(
		dao.WithEndpoint(srv.URL),
		dao.WithTimeout(100*time.Millisecond),
	)
	require.NoError(t, err)
	defer c.Close()

	_, err = c.Nodes().List()
	require.Error(t, err)
	var w9sErr *errs.W9sError
	if errors.As(err, &w9sErr) {
		assert.Equal(t, errs.ErrTimeout, w9sErr.Code)
	}
}

// ---------- Basic Auth Verification ----------

func TestBasicAuthHeaderSent(t *testing.T) {
	var receivedUser, receivedPass string
	var authPresent bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUser, receivedPass, authPresent = r.BasicAuth()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := dao.NewHTTPClient(
		dao.WithEndpoint(srv.URL),
		dao.WithBasicAuth("myuser", "mypass"),
	)
	require.NoError(t, err)
	defer c.Close()

	_, _ = c.Nodes().List()

	assert.True(t, authPresent, "expected Basic Auth header to be present")
	assert.Equal(t, "myuser", receivedUser)
	assert.Equal(t, "mypass", receivedPass)
}

func TestNoAuthHeaderWhenNotConfigured(t *testing.T) {
	var authPresent bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, authPresent = r.BasicAuth()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := dao.NewHTTPClient(dao.WithEndpoint(srv.URL))
	require.NoError(t, err)
	defer c.Close()

	_, _ = c.Nodes().List()

	assert.False(t, authPresent, "expected no Basic Auth header")
}

// ---------- Request Verification ----------

func TestImageImportSendsSourceInBody(t *testing.T) {
	var receivedBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Images().Import("testimg", "docker://rockylinux:9")
	require.NoError(t, err)
	assert.Equal(t, "docker://rockylinux:9", receivedBody["source"])
}

func TestOverlayDeleteForceQueryParam(t *testing.T) {
	var receivedURL string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	err := c.Overlays().Delete("myoverlay", true)
	require.NoError(t, err)
	assert.Contains(t, receivedURL, "force=true")
}

func TestOverlayGetFileQueryParams(t *testing.T) {
	var receivedPath, receivedRender string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Query().Get("path")
		receivedRender = r.URL.Query().Get("render")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cannedOverlayFileJSON))
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	_, err := c.Overlays().GetFile("wwinit", "/etc/hostname", "node1")
	require.NoError(t, err)
	assert.Equal(t, "/etc/hostname", receivedPath)
	assert.Equal(t, "node1", receivedRender)
}

// ---------- HTTP Method Verification ----------

func TestCorrectHTTPMethods(t *testing.T) {
	tests := []struct {
		name     string
		action   func(c *dao.HTTPClient) error
		wantPath string
		wantMethod string
	}{
		{
			name:       "NodeAdd uses PUT",
			action:     func(c *dao.HTTPClient) error { return c.Nodes().Add("n1", &dao.WwNode{}) },
			wantPath:   "/api/nodes/n1",
			wantMethod: http.MethodPut,
		},
		{
			name:       "NodeUpdate uses PATCH",
			action:     func(c *dao.HTTPClient) error { return c.Nodes().Update("n1", &dao.WwNode{}) },
			wantPath:   "/api/nodes/n1",
			wantMethod: http.MethodPatch,
		},
		{
			name:       "NodeDelete uses DELETE",
			action:     func(c *dao.HTTPClient) error { return c.Nodes().Delete("n1") },
			wantPath:   "/api/nodes/n1",
			wantMethod: http.MethodDelete,
		},
		{
			name:       "ProfileAdd uses PUT",
			action:     func(c *dao.HTTPClient) error { return c.Profiles().Add("p1", &dao.WwProfile{}) },
			wantPath:   "/api/profiles/p1",
			wantMethod: http.MethodPut,
		},
		{
			name:       "ProfileDelete uses DELETE",
			action:     func(c *dao.HTTPClient) error { return c.Profiles().Delete("p1") },
			wantPath:   "/api/profiles/p1",
			wantMethod: http.MethodDelete,
		},
		{
			name:       "ImageDelete uses DELETE",
			action:     func(c *dao.HTTPClient) error { return c.Images().Delete("img1") },
			wantPath:   "/api/images/img1",
			wantMethod: http.MethodDelete,
		},
		{
			name:       "OverlayCreate uses PUT",
			action:     func(c *dao.HTTPClient) error { return c.Overlays().Create("ov1") },
			wantPath:   "/api/overlays/ov1",
			wantMethod: http.MethodPut,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotMethod, gotPath string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()
			c := newClient(t, srv.URL)

			err := tc.action(c)
			require.NoError(t, err)
			assert.Equal(t, tc.wantMethod, gotMethod)
			assert.Equal(t, tc.wantPath, gotPath)
		})
	}
}

// ---------- Space-in-Key JSON Deserialization ----------

func TestSpaceInKeyDeserialization(t *testing.T) {
	// Directly test that JSON with space-in-key fields deserializes correctly.
	jsonData := `{
		"discoverable": "true",
		"asset key": "ABC123",
		"profiles": ["default"],
		"cluster name": "test-cluster",
		"image name": "ubuntu22",
		"runtime overlay": ["rt1", "rt2"],
		"system overlay": ["sys1"],
		"network devices": {
			"ib0": {
				"ipaddr": "192.168.1.1",
				"netmask": "255.255.0.0",
				"hwaddr": "AA:BB:CC:DD:EE:FF"
			}
		}
	}`

	var node dao.WwNode
	err := json.Unmarshal([]byte(jsonData), &node)
	require.NoError(t, err)

	assert.Equal(t, dao.WWBoolTrue, node.Discoverable)
	assert.Equal(t, "ABC123", node.AssetKey)
	assert.Equal(t, "test-cluster", node.ClusterName)
	assert.Equal(t, "ubuntu22", node.ImageName)
	assert.Equal(t, []string{"rt1", "rt2"}, node.RuntimeOverlay)
	assert.Equal(t, []string{"sys1"}, node.SystemOverlay)

	require.Contains(t, node.NetDevs, "ib0")
	assert.Equal(t, "192.168.1.1", node.NetDevs["ib0"].Ipaddr)
}

// ---------- Client Options ----------

func TestWithInsecure(t *testing.T) {
	c, err := dao.NewHTTPClient(
		dao.WithEndpoint("https://localhost:9873"),
		dao.WithInsecure(true),
	)
	require.NoError(t, err)
	defer c.Close()
	// Just verify it doesn't panic or error; TLS config is internal.
}

func TestWithEndpointTrailingSlash(t *testing.T) {
	// Ensure trailing slashes are stripped so paths don't get doubled.
	var receivedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := dao.NewHTTPClient(dao.WithEndpoint(srv.URL + "/"))
	require.NoError(t, err)
	defer c.Close()

	_, _ = c.Nodes().List()
	assert.Equal(t, "/api/nodes/", receivedPath)
	assert.NotContains(t, receivedPath, "//")
}

// ---------- Close ----------

func TestClose(t *testing.T) {
	c, err := dao.NewHTTPClient(dao.WithEndpoint("http://localhost:1"))
	require.NoError(t, err)
	err = c.Close()
	assert.NoError(t, err)
}

// ---------- ContentType Header ----------

func TestContentTypeHeader(t *testing.T) {
	var gotContentType, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	_, _ = c.Nodes().List()

	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "application/json", gotAccept)
}

// ---------- ServerInfo Fallback ----------

func TestServerInfoFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	info, err := c.ServerInfo()
	require.NoError(t, err)
	assert.Equal(t, "unknown", info.Version)
}

// ---------- Noop Power Manager ----------

func TestNoopPowerManager(t *testing.T) {
	// The HTTPClient's HasPower depends on wwctl being on PATH.
	// We can at least verify the noop power manager returns appropriate errors
	// by testing a client that won't have wwctl.

	// This test is a bit environment-dependent; we test the concept
	// by checking that Power() returns a manager (noop or real).
	c, err := dao.NewHTTPClient(dao.WithEndpoint("http://localhost:1"))
	require.NoError(t, err)
	defer c.Close()

	pm := c.Power()
	require.NotNil(t, pm)

	// If wwctl is not on PATH, all operations should return errors.
	if !c.HasPower() {
		_, err := pm.Status("node1")
		require.Error(t, err)

		err = pm.On("node1")
		require.Error(t, err)

		err = pm.Off("node1")
		require.Error(t, err)

		err = pm.Cycle("node1")
		require.Error(t, err)

		err = pm.Reset("node1")
		require.Error(t, err)

		var w9sErr *errs.W9sError
		if errors.As(err, &w9sErr) {
			assert.Equal(t, errs.ErrPowerUnavailable, w9sErr.Code)
		}
	}
}

// ---------- Empty Response Body on Action ----------

func TestActionEndpointsIgnoreResponseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	assert.NoError(t, c.Nodes().Delete("n1"))
	assert.NoError(t, c.Profiles().Delete("p1"))
	assert.NoError(t, c.Images().Delete("i1"))
	assert.NoError(t, c.Overlays().Delete("o1", false))
}

// ---------- JSON with Various WWBool Values in API Response ----------

func TestNodeListWithVariousWWBoolValues(t *testing.T) {
	cannedJSON := `{
		"n1": {"discoverable": "yes"},
		"n2": {"discoverable": "1"},
		"n3": {"discoverable": "no"},
		"n4": {"discoverable": "0"},
		"n5": {"discoverable": "UNDEF"},
		"n6": {"discoverable": ""}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cannedJSON))
	}))
	defer srv.Close()
	c := newClient(t, srv.URL)

	nodes, err := c.Nodes().List()
	require.NoError(t, err)

	assert.Equal(t, dao.WWBoolTrue, nodes["n1"].Discoverable)
	assert.Equal(t, dao.WWBoolTrue, nodes["n2"].Discoverable)
	assert.Equal(t, dao.WWBoolFalse, nodes["n3"].Discoverable)
	assert.Equal(t, dao.WWBoolFalse, nodes["n4"].Discoverable)
	assert.Equal(t, dao.WWBoolUndef, nodes["n5"].Discoverable)
	assert.Equal(t, dao.WWBoolFalse, nodes["n6"].Discoverable)
}

// ---------- Concurrent Request Safety ----------

func TestConcurrentRequests(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	c := newClient(t, srv.URL)

	const goroutines = 20
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := c.Nodes().List()
			errs <- err
		}()
	}

	for i := 0; i < goroutines; i++ {
		assert.NoError(t, <-errs)
	}
}

// ---------- Multiple Status Codes Table Test ----------

func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantCode   *errs.ErrorCode
		wantInErr  string
	}{
		{"200 OK", 200, nil, ""},
		{"401 Unauthorized", 401, ptr(errs.ErrAuth), "authentication failed"},
		{"403 Forbidden", 403, ptr(errs.ErrForbidden), "access denied"},
		{"404 Not Found", 404, ptr(errs.ErrNotFound), "not found"},
		{"500 Internal Server Error", 500, nil, "500"},
		{"502 Bad Gateway", 502, nil, "502"},
		{"503 Service Unavailable", 503, nil, "503"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tc.statusCode >= 200 && tc.statusCode < 300 {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{}`))
				} else {
					w.WriteHeader(tc.statusCode)
					_, _ = w.Write([]byte("error body"))
				}
			}))
			defer srv.Close()
			c := newClient(t, srv.URL)

			_, err := c.Nodes().List()
			if tc.wantCode != nil {
				require.Error(t, err)
				var w9sErr *errs.W9sError
				require.True(t, errors.As(err, &w9sErr))
				assert.Equal(t, *tc.wantCode, w9sErr.Code)
			}
			if tc.wantInErr != "" {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.wantInErr))
			}
			if tc.wantCode == nil && tc.wantInErr == "" {
				assert.NoError(t, err)
			}
		})
	}
}

func ptr(c errs.ErrorCode) *errs.ErrorCode { return &c }
