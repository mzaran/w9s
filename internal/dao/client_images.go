package dao

import (
	"net/http"
	"net/url"
)

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
