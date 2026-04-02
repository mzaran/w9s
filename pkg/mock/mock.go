// Package mock provides an in-memory mock implementation of dao.WarewulfClient
// for development and testing.
package mock

import (
	"fmt"
	"sync"
	"time"

	"github.com/mzaran/w9s/internal/dao"
)

// MockWarewulfClient is an in-memory implementation of dao.WarewulfClient.
type MockWarewulfClient struct {
	nodes    *mockNodeManager
	profiles *mockProfileManager
	images   *mockImageManager
	overlays *mockOverlayManager
	power    *mockPowerManager
	delay    time.Duration
}

// NewMockClient creates a new mock client populated with sample data.
// Simulates a small delay on operations for realism.
func NewMockClient() dao.WarewulfClient {
	c := &MockWarewulfClient{
		delay: 50 * time.Millisecond,
	}
	c.nodes = &mockNodeManager{data: make(map[string]*dao.WwNode), delay: c.delay}
	c.profiles = &mockProfileManager{data: make(map[string]*dao.WwProfile), delay: c.delay}
	c.images = &mockImageManager{data: make(map[string]*dao.WwImage), delay: c.delay}
	c.overlays = &mockOverlayManager{data: make(map[string]*dao.WwOverlay), delay: c.delay}
	c.power = &mockPowerManager{status: make(map[string]string), delay: c.delay}
	populateSampleData(c)
	return c
}

// NewFastMockClient creates a mock client with zero delay, suitable for tests.
func NewFastMockClient() dao.WarewulfClient {
	c := &MockWarewulfClient{
		delay: 0,
	}
	c.nodes = &mockNodeManager{data: make(map[string]*dao.WwNode), delay: 0}
	c.profiles = &mockProfileManager{data: make(map[string]*dao.WwProfile), delay: 0}
	c.images = &mockImageManager{data: make(map[string]*dao.WwImage), delay: 0}
	c.overlays = &mockOverlayManager{data: make(map[string]*dao.WwOverlay), delay: 0}
	c.power = &mockPowerManager{status: make(map[string]string), delay: 0}
	populateSampleData(c)
	return c
}

func (c *MockWarewulfClient) Nodes() dao.NodeManager       { return c.nodes }
func (c *MockWarewulfClient) Profiles() dao.ProfileManager { return c.profiles }
func (c *MockWarewulfClient) Images() dao.ImageManager     { return c.images }
func (c *MockWarewulfClient) Overlays() dao.OverlayManager { return c.overlays }
func (c *MockWarewulfClient) Power() dao.PowerManager      { return c.power }
func (c *MockWarewulfClient) HasPower() bool               { return true }
func (c *MockWarewulfClient) Close() error                 { return nil }

func (c *MockWarewulfClient) ServerInfo() (*dao.ServerInfo, error) {
	return &dao.ServerInfo{Version: "4.5.0-mock"}, nil
}

// ---------- Node Manager ----------

type mockNodeManager struct {
	mu    sync.RWMutex
	data  map[string]*dao.WwNode
	delay time.Duration
}

func (m *mockNodeManager) sleep() {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
}

func (m *mockNodeManager) List() (map[string]*dao.WwNode, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]*dao.WwNode, len(m.data))
	for k, v := range m.data {
		out[k] = v
	}
	return out, nil
}

func (m *mockNodeManager) Get(id string) (*dao.WwNode, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.data[id]
	if !ok {
		return nil, fmt.Errorf("node %q not found", id)
	}
	return n, nil
}

func (m *mockNodeManager) GetRaw(id string) (*dao.WwNode, error) {
	return m.Get(id)
}

func (m *mockNodeManager) GetFields(id string) ([]dao.NodeField, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.data[id]
	if !ok {
		return nil, fmt.Errorf("node %q not found", id)
	}
	fields := []dao.NodeField{
		{Field: "ImageName", Value: n.ImageName, Source: "node"},
		{Field: "ClusterName", Value: n.ClusterName, Source: "profile:default"},
	}
	return fields, nil
}

func (m *mockNodeManager) GetOverlayInfo(id string) (*dao.NodeOverlayInfo, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[id]
	if !ok {
		return nil, fmt.Errorf("node %q not found", id)
	}
	sysTime := time.Now().Add(-1 * time.Hour)
	rtTime := time.Now().Add(-30 * time.Minute)
	return &dao.NodeOverlayInfo{
		SystemOverlay: &dao.OverlayStatus{
			Overlays: []string{"wwinit"},
			MTime:    &sysTime,
		},
		RuntimeOverlay: &dao.OverlayStatus{
			Overlays: []string{"runtime"},
			MTime:    &rtTime,
		},
	}, nil
}

func (m *mockNodeManager) BuildOverlays(_ string) error {
	m.sleep()
	return nil
}

func (m *mockNodeManager) BuildAllOverlays() error {
	m.sleep()
	return nil
}

func (m *mockNodeManager) Add(id string, node *dao.WwNode) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = node
	return nil
}

func (m *mockNodeManager) Update(id string, node *dao.WwNode) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return fmt.Errorf("node %q not found", id)
	}
	m.data[id] = node
	return nil
}

func (m *mockNodeManager) Delete(id string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return fmt.Errorf("node %q not found", id)
	}
	delete(m.data, id)
	return nil
}

// ---------- Profile Manager ----------

type mockProfileManager struct {
	mu    sync.RWMutex
	data  map[string]*dao.WwProfile
	delay time.Duration
}

func (m *mockProfileManager) sleep() {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
}

func (m *mockProfileManager) List() (map[string]*dao.WwProfile, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]*dao.WwProfile, len(m.data))
	for k, v := range m.data {
		out[k] = v
	}
	return out, nil
}

func (m *mockProfileManager) Get(id string) (*dao.WwProfile, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data[id]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", id)
	}
	return p, nil
}

func (m *mockProfileManager) Add(id string, profile *dao.WwProfile) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = profile
	return nil
}

func (m *mockProfileManager) Update(id string, profile *dao.WwProfile) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return fmt.Errorf("profile %q not found", id)
	}
	m.data[id] = profile
	return nil
}

func (m *mockProfileManager) Delete(id string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return fmt.Errorf("profile %q not found", id)
	}
	delete(m.data, id)
	return nil
}

// ---------- Image Manager ----------

type mockImageManager struct {
	mu    sync.RWMutex
	data  map[string]*dao.WwImage
	delay time.Duration
}

func (m *mockImageManager) sleep() {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
}

func (m *mockImageManager) List() (map[string]*dao.WwImage, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]*dao.WwImage, len(m.data))
	for k, v := range m.data {
		out[k] = v
	}
	return out, nil
}

func (m *mockImageManager) Get(name string) (*dao.WwImage, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	img, ok := m.data[name]
	if !ok {
		return nil, fmt.Errorf("image %q not found", name)
	}
	return img, nil
}

func (m *mockImageManager) Build(_ string) error {
	m.sleep()
	return nil
}

func (m *mockImageManager) Import(name, _ string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[name] = &dao.WwImage{
		Size:      1073741824,
		BuildTime: time.Now().Unix(),
	}
	return nil
}

func (m *mockImageManager) Update(name, newName string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	img, ok := m.data[name]
	if !ok {
		return fmt.Errorf("image %q not found", name)
	}
	delete(m.data, name)
	m.data[newName] = img
	return nil
}

func (m *mockImageManager) Delete(name string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[name]; !ok {
		return fmt.Errorf("image %q not found", name)
	}
	delete(m.data, name)
	return nil
}

// ---------- Overlay Manager ----------

type mockOverlayManager struct {
	mu    sync.RWMutex
	data  map[string]*dao.WwOverlay
	delay time.Duration
}

func (m *mockOverlayManager) sleep() {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
}

func (m *mockOverlayManager) List() (map[string]*dao.WwOverlay, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]*dao.WwOverlay, len(m.data))
	for k, v := range m.data {
		out[k] = v
	}
	return out, nil
}

func (m *mockOverlayManager) Get(name string) (*dao.WwOverlay, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.data[name]
	if !ok {
		return nil, fmt.Errorf("overlay %q not found", name)
	}
	return o, nil
}

func (m *mockOverlayManager) GetFile(name, path, _ string) (*dao.OverlayFile, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[name]
	if !ok {
		return nil, fmt.Errorf("overlay %q not found", name)
	}
	return &dao.OverlayFile{
		Overlay:  name,
		Path:     path,
		Contents: "# Mock file contents for " + path + "\n",
		Perms:    0644,
	}, nil
}

func (m *mockOverlayManager) Create(name string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[name] = &dao.WwOverlay{}
	return nil
}

func (m *mockOverlayManager) AddFile(name, path, _ string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.data[name]
	if !ok {
		return fmt.Errorf("overlay %q not found", name)
	}
	o.Files = append(o.Files, path)
	return nil
}

func (m *mockOverlayManager) Delete(name string, _ bool) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[name]; !ok {
		return fmt.Errorf("overlay %q not found", name)
	}
	delete(m.data, name)
	return nil
}

func (m *mockOverlayManager) DeleteFile(name, path string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.data[name]
	if !ok {
		return fmt.Errorf("overlay %q not found", name)
	}
	for i, f := range o.Files {
		if f == path {
			o.Files = append(o.Files[:i], o.Files[i+1:]...)
			break
		}
	}
	return nil
}

// ---------- Power Manager ----------

type mockPowerManager struct {
	mu     sync.RWMutex
	status map[string]string
	delay  time.Duration
}

func (m *mockPowerManager) sleep() {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
}

func (m *mockPowerManager) Status(nodeID string) (string, error) {
	m.sleep()
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.status[nodeID]
	if !ok {
		return "on", nil
	}
	return s, nil
}

func (m *mockPowerManager) On(nodeID string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status[nodeID] = "on"
	return nil
}

func (m *mockPowerManager) Off(nodeID string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status[nodeID] = "off"
	return nil
}

func (m *mockPowerManager) Cycle(nodeID string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status[nodeID] = "on"
	return nil
}

func (m *mockPowerManager) Reset(nodeID string) error {
	m.sleep()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status[nodeID] = "on"
	return nil
}
