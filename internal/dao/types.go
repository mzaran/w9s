package dao

import (
	"encoding/json"
	"strings"
	"time"
)

// WWBool represents a Warewulf boolean that can be true, false, or unset (UNDEF).
// Warewulf transmits these as strings: "true", "yes", "1", "false", "no", "0", "", "UNDEF".
type WWBool string

const (
	WWBoolTrue  WWBool = "true"
	WWBoolFalse WWBool = "false"
	WWBoolUndef WWBool = "UNDEF"
)

// UnmarshalJSON handles the various string representations of Warewulf booleans.
func (b *WWBool) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		// Try as a native JSON boolean.
		var boolVal bool
		if err2 := json.Unmarshal(data, &boolVal); err2 != nil {
			return err
		}
		if boolVal {
			*b = WWBoolTrue
		} else {
			*b = WWBoolFalse
		}
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "yes", "1":
		*b = WWBoolTrue
	case "false", "no", "0", "":
		*b = WWBoolFalse
	case "undef":
		*b = WWBoolUndef
	default:
		*b = WWBool(raw)
	}
	return nil
}

// MarshalJSON encodes the WWBool as a JSON string.
func (b WWBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

// IsTrue reports whether the value is explicitly true.
func (b WWBool) IsTrue() bool { return b == WWBoolTrue }

// IsFalse reports whether the value is explicitly false.
func (b WWBool) IsFalse() bool { return b == WWBoolFalse }

// IsUnset reports whether the value is unset (UNDEF or empty).
func (b WWBool) IsUnset() bool { return b == WWBoolUndef || b == "" }

// String returns the string representation.
func (b WWBool) String() string { return string(b) }

// WwNode represents a Warewulf compute node.
type WwNode struct {
	Discoverable WWBool `json:"discoverable,omitempty"`
	AssetKey     string `json:"asset key,omitempty"`
	WwProfile
}

// WwProfile represents a Warewulf configuration profile.
// Profiles provide layered defaults that nodes inherit.
type WwProfile struct {
	Profiles       []string               `json:"profiles,omitempty"`
	Comment        string                 `json:"comment,omitempty"`
	ClusterName    string                 `json:"cluster name,omitempty"`
	ImageName      string                 `json:"image name,omitempty"`
	Ipxe           string                 `json:"ipxe template,omitempty"`
	RuntimeOverlay []string               `json:"runtime overlay,omitempty"`
	SystemOverlay  []string               `json:"system overlay,omitempty"`
	Kernel         *KernelConf            `json:"kernel,omitempty"`
	Ipmi           *IpmiConf              `json:"ipmi,omitempty"`
	Init           string                 `json:"init,omitempty"`
	Root           string                 `json:"root,omitempty"`
	NetDevs        map[string]*NetDev     `json:"network devices,omitempty"`
	Tags           map[string]string      `json:"tags,omitempty"`
	PrimaryNetDev  string                 `json:"primary network,omitempty"`
	Disks          map[string]*Disk       `json:"disks,omitempty"`
	FileSystems    map[string]*FileSystem `json:"filesystems,omitempty"`
}

// IpmiConf holds IPMI/BMC configuration for a node.
type IpmiConf struct {
	UserName   string            `json:"username,omitempty"`
	Password   string            `json:"password,omitempty"`
	Ipaddr     string            `json:"ipaddr,omitempty"`
	Gateway    string            `json:"gateway,omitempty"`
	Netmask    string            `json:"netmask,omitempty"`
	Port       string            `json:"port,omitempty"`
	Interface  string            `json:"interface,omitempty"`
	EscapeChar string            `json:"escapechar,omitempty"`
	Write      WWBool            `json:"write,omitempty"`
	Template   string            `json:"template,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// KernelConf holds kernel boot configuration.
type KernelConf struct {
	Version string   `json:"version,omitempty"`
	Args    []string `json:"args,omitempty"`
}

// NetDev represents a network device configuration.
type NetDev struct {
	Type       string            `json:"type,omitempty"`
	Device     string            `json:"device,omitempty"`
	Hwaddr     string            `json:"hwaddr,omitempty"`
	Ipaddr     string            `json:"ipaddr,omitempty"`
	Netmask    string            `json:"netmask,omitempty"`
	Gateway    string            `json:"gateway,omitempty"`
	Ipaddr6    string            `json:"ipaddr6,omitempty"`
	PrefixLen6 string            `json:"prefixlen6,omitempty"`
	Gateway6   string            `json:"gateway6,omitempty"`
	MTU        string            `json:"mtu,omitempty"`
	OnBoot     WWBool            `json:"onbot,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// Disk represents a disk configuration for Ignition-style provisioning.
type Disk struct {
	WipeTable  *bool                 `json:"wipe_table,omitempty"`
	Partitions map[string]*Partition `json:"partitions,omitempty"`
}

// Partition represents a disk partition.
type Partition struct {
	Number             string `json:"number,omitempty"`
	SizeMiB            string `json:"size_mib,omitempty"`
	StartMiB           string `json:"start_mib,omitempty"`
	TypeGUID           string `json:"type_guid,omitempty"`
	GUID               string `json:"guid,omitempty"`
	WipePartitionEntry *bool  `json:"wipe_partition_entry,omitempty"`
	ShouldExist        *bool  `json:"should_exist,omitempty"`
	Resize             *bool  `json:"resize,omitempty"`
}

// FileSystem represents a filesystem configuration.
type FileSystem struct {
	Format         string   `json:"format,omitempty"`
	Path           string   `json:"path,omitempty"`
	Label          string   `json:"label,omitempty"`
	UUID           string   `json:"uuid,omitempty"`
	MountOptions   string   `json:"mount_options,omitempty"`
	WipeFileSystem *bool    `json:"wipe_filesystem,omitempty"`
	Options        []string `json:"options,omitempty"`
}

// WwImage represents a Warewulf container image.
type WwImage struct {
	Kernels   []string `json:"kernels,omitempty"`
	Size      int64    `json:"size,omitempty"`
	BuildTime int64    `json:"buildtime,omitempty"`
	Writable  bool     `json:"writable,omitempty"`
}

// WwOverlay represents a Warewulf overlay.
type WwOverlay struct {
	Files []string `json:"files,omitempty"`
	Site  bool     `json:"site,omitempty"`
}

// OverlayFile represents a single file within an overlay.
type OverlayFile struct {
	Overlay  string `json:"overlay"`
	Path     string `json:"path"`
	Contents string `json:"contents"`
	Perms    uint32 `json:"perms"`
	UID      uint32 `json:"uid"`
	Gid      uint32 `json:"gid"`
}

// NodeField represents a single configuration field for a node,
// including where the value was inherited from.
type NodeField struct {
	Field  string `json:"field"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

// NodeOverlayInfo contains overlay build status for a node.
type NodeOverlayInfo struct {
	SystemOverlay  *OverlayStatus `json:"system overlay,omitempty"`
	RuntimeOverlay *OverlayStatus `json:"runtime overlay,omitempty"`
}

// OverlayStatus describes the build state of an overlay set.
type OverlayStatus struct {
	Overlays []string   `json:"overlays,omitempty"`
	MTime    *time.Time `json:"mtime,omitempty"`
}

// ServerInfo contains basic Warewulf server information.
type ServerInfo struct {
	Version string `json:"version"`
}
