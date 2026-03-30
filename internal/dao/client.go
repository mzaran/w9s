package dao

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mzaran/w9s/internal/errs"
)

const defaultTimeout = 30 * time.Second

// ClientOption configures an HTTPClient.
type ClientOption func(*HTTPClient)

// WithEndpoint sets the base URL for the Warewulf API.
func WithEndpoint(rawURL string) ClientOption {
	return func(c *HTTPClient) {
		c.baseURL = strings.TrimRight(rawURL, "/")
	}
}

// WithBasicAuth sets HTTP Basic Authentication credentials.
func WithBasicAuth(username, password string) ClientOption {
	return func(c *HTTPClient) {
		c.username = username
		c.password = password
	}
}

// WithInsecure disables TLS certificate verification.
func WithInsecure(skipVerify bool) ClientOption {
	return func(c *HTTPClient) {
		c.insecure = skipVerify
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *HTTPClient) {
		c.timeout = d
	}
}

// HTTPClient is a real HTTP client that implements WarewulfClient
// by talking to the Warewulf REST API.
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
	insecure   bool
	timeout    time.Duration

	hasPower     bool
	hasPowerOnce sync.Once
	powerMgr     PowerManager
}

// NewHTTPClient creates a new HTTPClient with the given options.
func NewHTTPClient(opts ...ClientOption) (*HTTPClient, error) {
	c := &HTTPClient{
		baseURL: "https://localhost:9873",
		timeout: defaultTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.insecure, //nolint:gosec // user-controlled
		},
	}
	c.httpClient = &http.Client{
		Timeout:   c.timeout,
		Transport: transport,
	}

	return c, nil
}

// NewClientFromConfig creates an HTTPClient from typical configuration values.
// This is used by both the CLI and integration tests (DRY).
func NewClientFromConfig(endpoint, username, password string, insecure bool, timeout time.Duration) (*HTTPClient, error) {
	opts := []ClientOption{WithEndpoint(endpoint)}
	if username != "" {
		opts = append(opts, WithBasicAuth(username, password))
	}
	if insecure {
		opts = append(opts, WithInsecure(true))
	}
	if timeout > 0 {
		opts = append(opts, WithTimeout(timeout))
	}
	return NewHTTPClient(opts...)
}

// Nodes returns the node manager.
func (c *HTTPClient) Nodes() NodeManager { return &httpNodeManager{client: c} }

// Profiles returns the profile manager.
func (c *HTTPClient) Profiles() ProfileManager { return &httpProfileManager{client: c} }

// Images returns the image manager.
func (c *HTTPClient) Images() ImageManager { return &httpImageManager{client: c} }

// Overlays returns the overlay manager.
func (c *HTTPClient) Overlays() OverlayManager { return &httpOverlayManager{client: c} }

// HasPower reports whether the wwctl binary is available on PATH.
func (c *HTTPClient) HasPower() bool {
	c.hasPowerOnce.Do(func() {
		_, err := exec.LookPath("wwctl")
		c.hasPower = err == nil
	})
	return c.hasPower
}

// Power returns the power manager. If wwctl is not available, returns a no-op manager.
func (c *HTTPClient) Power() PowerManager {
	if !c.HasPower() {
		return &noopPowerManager{}
	}
	if c.powerMgr == nil {
		mgr, err := NewWwctlPowerManager()
		if err != nil {
			return &noopPowerManager{}
		}
		c.powerMgr = mgr
	}
	return c.powerMgr
}

// ServerInfo returns basic server information.
func (c *HTTPClient) ServerInfo() (*ServerInfo, error) {
	// Try the root API endpoint for version info.
	resp, err := c.doRequest(http.MethodGet, "/api/", nil)
	if err != nil {
		// Fall back to a placeholder if the root endpoint is not available.
		return &ServerInfo{Version: "unknown"}, nil
	}
	defer resp.Body.Close()

	var info ServerInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return &ServerInfo{Version: "unknown"}, nil
	}
	return &info, nil
}

// Close releases any resources held by the client.
func (c *HTTPClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// doRequest builds and executes an HTTP request, returning the response.
// It sets Basic Auth and Content-Type headers automatically.
func (c *HTTPClient) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	reqURL := c.baseURL + path

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.classifyError(err)
	}

	return resp, nil
}

// doJSON performs a request and decodes the JSON response body into dest.
// It checks for HTTP error status codes and maps them to typed errors.
func (c *HTTPClient) doJSON(method, path string, body io.Reader, dest interface{}) error {
	resp, err := c.doRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkStatus(resp); err != nil {
		return err
	}

	if dest != nil {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// doAction performs a request that does not return a meaningful body.
func (c *HTTPClient) doAction(method, path string, body io.Reader) error {
	resp, err := c.doRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Drain the body so the connection can be reused.
	io.Copy(io.Discard, resp.Body) //nolint:errcheck
	return c.checkStatus(resp)
}

// checkStatus maps HTTP status codes to typed errors.
func (c *HTTPClient) checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Read body for error message.
	bodyBytes, _ := io.ReadAll(resp.Body)
	msg := strings.TrimSpace(string(bodyBytes))
	if msg == "" {
		msg = resp.Status
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return errs.NewAuthError(nil, fmt.Sprintf("authentication failed: %s", msg))
	case http.StatusForbidden:
		return errs.NewForbiddenError(nil, fmt.Sprintf("access denied: %s", msg))
	case http.StatusNotFound:
		return errs.NewNotFoundError(nil, fmt.Sprintf("not found: %s", msg))
	default:
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}
}

// classifyError maps network-level errors to typed errors.
func (c *HTTPClient) classifyError(err error) error {
	if err == nil {
		return nil
	}

	// Check for connection refused (API disabled).
	var opErr *net.OpError
	if ok := errorAs(err, &opErr); ok {
		if opErr.Op == "dial" {
			return errs.NewAPIDisabledError(err, "cannot connect to Warewulf API (connection refused)")
		}
	}

	// Check for timeout.
	if isTimeout(err) {
		return errs.NewTimeoutError(err, "request to Warewulf API timed out")
	}

	return errs.NewConnectionError(err, fmt.Sprintf("API request failed: %v", err))
}

// errorAs is a helper wrapping errors.As for net.OpError.
func errorAs(err error, target interface{}) bool {
	// Use type switch to avoid import cycle concerns.
	switch t := target.(type) {
	case **net.OpError:
		for e := err; e != nil; {
			if opErr, ok := e.(*net.OpError); ok {
				*t = opErr
				return true
			}
			if u, ok := e.(interface{ Unwrap() error }); ok {
				e = u.Unwrap()
			} else {
				return false
			}
		}
	}
	return false
}

// isTimeout checks whether an error is a timeout.
func isTimeout(err error) bool {
	for e := err; e != nil; {
		if t, ok := e.(interface{ Timeout() bool }); ok && t.Timeout() {
			return true
		}
		if u, ok := e.(interface{ Unwrap() error }); ok {
			e = u.Unwrap()
		} else {
			return false
		}
	}
	return false
}

// jsonBody encodes v as JSON and returns a reader.
func jsonBody(v interface{}) (io.Reader, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return nil, fmt.Errorf("encoding JSON body: %w", err)
	}
	return buf, nil
}

// ---------- Node Manager ----------

type httpNodeManager struct {
	client *HTTPClient
}

func (m *httpNodeManager) List() (map[string]*WwNode, error) {
	var result map[string]*WwNode
	if err := m.client.doJSON(http.MethodGet, "/api/nodes/", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *httpNodeManager) Get(id string) (*WwNode, error) {
	var result WwNode
	if err := m.client.doJSON(http.MethodGet, "/api/nodes/"+url.PathEscape(id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *httpNodeManager) GetRaw(id string) (*WwNode, error) {
	var result WwNode
	if err := m.client.doJSON(http.MethodGet, "/api/nodes/"+url.PathEscape(id)+"/raw", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *httpNodeManager) GetFields(id string) ([]NodeField, error) {
	var result []NodeField
	if err := m.client.doJSON(http.MethodGet, "/api/nodes/"+url.PathEscape(id)+"/fields", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *httpNodeManager) GetOverlayInfo(id string) (*NodeOverlayInfo, error) {
	var result NodeOverlayInfo
	if err := m.client.doJSON(http.MethodGet, "/api/nodes/"+url.PathEscape(id)+"/overlays", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *httpNodeManager) BuildOverlays(id string) error {
	return m.client.doAction(http.MethodPost, "/api/nodes/"+url.PathEscape(id)+"/overlays/build", nil)
}

func (m *httpNodeManager) BuildAllOverlays() error {
	return m.client.doAction(http.MethodPost, "/api/nodes/overlays/build", nil)
}

func (m *httpNodeManager) Add(id string, node *WwNode) error {
	body, err := jsonBody(struct {
		Node *WwNode `json:"node"`
	}{Node: node})
	if err != nil {
		return err
	}
	return m.client.doAction(http.MethodPut, "/api/nodes/"+url.PathEscape(id), body)
}

func (m *httpNodeManager) Update(id string, node *WwNode) error {
	body, err := jsonBody(struct {
		Node *WwNode `json:"node"`
	}{Node: node})
	if err != nil {
		return err
	}
	return m.client.doAction(http.MethodPatch, "/api/nodes/"+url.PathEscape(id), body)
}

func (m *httpNodeManager) Delete(id string) error {
	return m.client.doAction(http.MethodDelete, "/api/nodes/"+url.PathEscape(id), nil)
}

// ---------- Profile Manager ----------

type httpProfileManager struct {
	client *HTTPClient
}

func (m *httpProfileManager) List() (map[string]*WwProfile, error) {
	var result map[string]*WwProfile
	if err := m.client.doJSON(http.MethodGet, "/api/profiles/", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *httpProfileManager) Get(id string) (*WwProfile, error) {
	var result WwProfile
	if err := m.client.doJSON(http.MethodGet, "/api/profiles/"+url.PathEscape(id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *httpProfileManager) Add(id string, profile *WwProfile) error {
	body, err := jsonBody(struct {
		Profile *WwProfile `json:"profile"`
	}{Profile: profile})
	if err != nil {
		return err
	}
	return m.client.doAction(http.MethodPut, "/api/profiles/"+url.PathEscape(id), body)
}

func (m *httpProfileManager) Update(id string, profile *WwProfile) error {
	body, err := jsonBody(struct {
		Profile *WwProfile `json:"profile"`
	}{Profile: profile})
	if err != nil {
		return err
	}
	return m.client.doAction(http.MethodPatch, "/api/profiles/"+url.PathEscape(id), body)
}

func (m *httpProfileManager) Delete(id string) error {
	return m.client.doAction(http.MethodDelete, "/api/profiles/"+url.PathEscape(id), nil)
}

// ---------- Image Manager ----------

type httpImageManager struct {
	client *HTTPClient
}

func (m *httpImageManager) List() (map[string]*WwImage, error) {
	var result map[string]*WwImage
	if err := m.client.doJSON(http.MethodGet, "/api/images/", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *httpImageManager) Get(name string) (*WwImage, error) {
	var result WwImage
	if err := m.client.doJSON(http.MethodGet, "/api/images/"+url.PathEscape(name), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *httpImageManager) Build(name string) error {
	return m.client.doAction(http.MethodPost, "/api/images/"+url.PathEscape(name)+"/build", nil)
}

func (m *httpImageManager) Import(name string, source string) error {
	body, err := jsonBody(map[string]string{"source": source})
	if err != nil {
		return err
	}
	return m.client.doAction(http.MethodPost, "/api/images/"+url.PathEscape(name)+"/import", body)
}

func (m *httpImageManager) Update(name string, newName string) error {
	body, err := jsonBody(map[string]string{"name": newName})
	if err != nil {
		return err
	}
	return m.client.doAction(http.MethodPatch, "/api/images/"+url.PathEscape(name), body)
}

func (m *httpImageManager) Delete(name string) error {
	return m.client.doAction(http.MethodDelete, "/api/images/"+url.PathEscape(name), nil)
}

// ---------- Overlay Manager ----------

type httpOverlayManager struct {
	client *HTTPClient
}

func (m *httpOverlayManager) List() (map[string]*WwOverlay, error) {
	var result map[string]*WwOverlay
	if err := m.client.doJSON(http.MethodGet, "/api/overlays/", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *httpOverlayManager) Get(name string) (*WwOverlay, error) {
	var result map[string]*WwOverlay
	if err := m.client.doJSON(http.MethodGet, "/api/overlays/"+url.PathEscape(name), nil, &result); err != nil {
		return nil, err
	}
	overlay, ok := result[name]
	if !ok {
		for _, v := range result {
			return v, nil
		}
		return nil, errs.NewNotFoundError(nil, fmt.Sprintf("overlay %q not in response", name))
	}
	return overlay, nil
}

func (m *httpOverlayManager) GetFile(name, path, renderNode string) (*OverlayFile, error) {
	params := url.Values{}
	params.Set("path", path)
	if renderNode != "" {
		params.Set("render", renderNode)
	}
	var result OverlayFile
	if err := m.client.doJSON(http.MethodGet, "/api/overlays/"+url.PathEscape(name)+"/file?"+params.Encode(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *httpOverlayManager) Create(name string) error {
	return m.client.doAction(http.MethodPut, "/api/overlays/"+url.PathEscape(name), nil)
}

func (m *httpOverlayManager) AddFile(name, path, content string) error {
	params := url.Values{}
	params.Set("path", path)
	body := strings.NewReader(content)
	return m.client.doAction(http.MethodPut, "/api/overlays/"+url.PathEscape(name)+"/file?"+params.Encode(), body)
}

func (m *httpOverlayManager) Delete(name string, force bool) error {
	path := "/api/overlays/" + url.PathEscape(name)
	if force {
		path += "?force=true"
	}
	return m.client.doAction(http.MethodDelete, path, nil)
}

func (m *httpOverlayManager) DeleteFile(name, path string) error {
	params := url.Values{}
	params.Set("path", path)
	return m.client.doAction(http.MethodDelete, "/api/overlays/"+url.PathEscape(name)+"/file?"+params.Encode(), nil)
}

// ---------- No-op Power Manager ----------

type noopPowerManager struct{}

func (n *noopPowerManager) Status(_ string) (string, error) {
	return "", errs.NewPowerUnavailableError(nil, "power management is not available (wwctl not found)")
}
func (n *noopPowerManager) On(_ string) error {
	return errs.NewPowerUnavailableError(nil, "power management is not available (wwctl not found)")
}
func (n *noopPowerManager) Off(_ string) error {
	return errs.NewPowerUnavailableError(nil, "power management is not available (wwctl not found)")
}
func (n *noopPowerManager) Cycle(_ string) error {
	return errs.NewPowerUnavailableError(nil, "power management is not available (wwctl not found)")
}
func (n *noopPowerManager) Reset(_ string) error {
	return errs.NewPowerUnavailableError(nil, "power management is not available (wwctl not found)")
}
