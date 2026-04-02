package dao

import (
	"net/http"
	"net/url"
)

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
