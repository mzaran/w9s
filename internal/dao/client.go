package dao

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
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

// WithPowerTimeout sets the timeout for power (IPMI) operations.
func WithPowerTimeout(d time.Duration) ClientOption {
	return func(c *HTTPClient) {
		c.powerTimeout = d
	}
}

// HTTPClient is a real HTTP client that implements WarewulfClient
// by talking to the Warewulf REST API.
type HTTPClient struct {
	baseURL      string
	httpClient   *http.Client
	username     string
	password     string
	insecure     bool
	timeout      time.Duration
	powerTimeout time.Duration

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
		timeout := c.powerTimeout
		if timeout <= 0 {
			timeout = DefaultPowerTimeout
		}
		mgr, err := NewWwctlPowerManagerWithTimeout(timeout)
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
