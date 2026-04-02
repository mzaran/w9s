package dao

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mzaran/w9s/internal/errs"
)

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
