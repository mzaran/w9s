// Package dao provides data access objects and interfaces for Warewulf cluster management.
package dao

// WarewulfClient is the main interface for interacting with a Warewulf server.
type WarewulfClient interface {
	// Nodes returns the node manager.
	Nodes() NodeManager

	// Profiles returns the profile manager.
	Profiles() ProfileManager

	// Images returns the image manager.
	Images() ImageManager

	// Overlays returns the overlay manager.
	Overlays() OverlayManager

	// Power returns the power manager for IPMI operations.
	Power() PowerManager

	// ServerInfo returns basic server information.
	ServerInfo() (*ServerInfo, error)

	// HasPower reports whether power management is available.
	HasPower() bool

	// Close releases any resources held by the client.
	Close() error
}

// NodeManager provides operations for managing Warewulf nodes.
type NodeManager interface {
	// List returns all nodes keyed by node name.
	List() (map[string]*WwNode, error)

	// Get returns the merged/effective configuration for a node.
	Get(id string) (*WwNode, error)

	// GetRaw returns the node configuration without profile merging.
	GetRaw(id string) (*WwNode, error)

	// GetFields returns the individual fields and their sources for a node.
	GetFields(id string) ([]NodeField, error)

	// GetOverlayInfo returns overlay build information for a node.
	GetOverlayInfo(id string) (*NodeOverlayInfo, error)

	// BuildOverlays triggers an overlay rebuild for the specified node.
	BuildOverlays(id string) error

	// BuildAllOverlays triggers an overlay rebuild for all nodes.
	BuildAllOverlays() error

	// Add creates a new node with the given configuration.
	Add(id string, node *WwNode) error

	// Update modifies an existing node.
	Update(id string, node *WwNode) error

	// Delete removes a node.
	Delete(id string) error
}

// ProfileManager provides operations for managing Warewulf profiles.
type ProfileManager interface {
	// List returns all profiles keyed by profile name.
	List() (map[string]*WwProfile, error)

	// Get returns a specific profile.
	Get(id string) (*WwProfile, error)

	// Add creates a new profile.
	Add(id string, profile *WwProfile) error

	// Update modifies an existing profile.
	Update(id string, profile *WwProfile) error

	// Delete removes a profile.
	Delete(id string) error
}

// ImageManager provides operations for managing Warewulf container images.
type ImageManager interface {
	// List returns all images keyed by image name.
	List() (map[string]*WwImage, error)

	// Get returns a specific image.
	Get(name string) (*WwImage, error)

	// Build triggers an image rebuild.
	Build(name string) error

	// Import imports a new image from the given source (URI or path).
	Import(name string, source string) error

	// Update renames an image.
	Update(name string, newName string) error

	// Delete removes an image.
	Delete(name string) error
}

// OverlayManager provides operations for managing Warewulf overlays.
type OverlayManager interface {
	// List returns all overlays keyed by overlay name.
	List() (map[string]*WwOverlay, error)

	// Get returns a specific overlay.
	Get(name string) (*WwOverlay, error)

	// GetFile returns the contents of a file within an overlay,
	// optionally rendered for a specific node.
	GetFile(name, path, renderNode string) (*OverlayFile, error)

	// Create creates a new empty overlay.
	Create(name string) error

	// AddFile adds or replaces a file in an overlay.
	AddFile(name, path, content string) error

	// Delete removes an overlay. If force is true, removes even if in use.
	Delete(name string, force bool) error

	// DeleteFile removes a single file from an overlay.
	DeleteFile(name, path string) error
}

// PowerManager provides IPMI power management operations.
type PowerManager interface {
	// Status returns the current power status string for a node.
	Status(nodeID string) (string, error)

	// On powers on a node.
	On(nodeID string) error

	// Off powers off a node.
	Off(nodeID string) error

	// Cycle power-cycles a node (off then on).
	Cycle(nodeID string) error

	// Reset sends a hardware reset to a node.
	Reset(nodeID string) error
}
