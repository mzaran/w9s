package dao

import (
	"net/http"
	"net/url"
)

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
