// Package config handles w9s configuration loading and validation.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config is the top-level w9s configuration.
type Config struct {
	RefreshRate    string         `yaml:"refreshRate" mapstructure:"refreshRate"`
	DefaultCluster string         `yaml:"defaultCluster" mapstructure:"defaultCluster"`
	Clusters       []ClusterEntry `yaml:"clusters" mapstructure:"clusters"`
	UI             UIConfig       `yaml:"ui" mapstructure:"ui"`

	// ActiveCluster is the resolved cluster configuration (computed, not in YAML).
	ActiveCluster *ClusterConfig `yaml:"-" mapstructure:"-"`
}

// ClusterEntry pairs a name with its cluster configuration.
type ClusterEntry struct {
	Name    string        `yaml:"name" mapstructure:"name"`
	Cluster ClusterConfig `yaml:"cluster" mapstructure:"cluster"`
}

// ClusterConfig holds connection details for a single Warewulf server.
type ClusterConfig struct {
	Endpoint string `yaml:"endpoint" mapstructure:"endpoint"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	Insecure bool   `yaml:"insecure" mapstructure:"insecure"`
	Timeout  string `yaml:"timeout" mapstructure:"timeout"`
}

// UIConfig holds user-interface preferences.
type UIConfig struct {
	EnableMouse bool `yaml:"enableMouse" mapstructure:"enableMouse"`
}

// Load reads and validates w9s configuration from the given file path,
// standard locations, or environment variables.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	// Set defaults.
	v.SetDefault("refreshRate", "5s")
	v.SetDefault("ui.enableMouse", false)

	if cfgFile != "" {
		// Explicit config file.
		v.SetConfigFile(cfgFile)
	} else {
		// Search standard locations.
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(filepath.Join(home, ".config", "w9s"))
			v.AddConfigPath(filepath.Join(home, ".w9s"))
		}
		v.SetConfigName("config")
	}

	// Try reading config file; not an error if missing.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && cfgFile != "" {
			return nil, fmt.Errorf("reading config file %s: %w", cfgFile, err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// If no clusters configured, try auto-detecting from Warewulf server config.
	if len(cfg.Clusters) == 0 {
		if entry, err := autoDetectWarewulf(); err == nil && entry != nil {
			cfg.Clusters = append(cfg.Clusters, *entry)
		}
	}

	// Apply environment variable overrides.
	applyEnvOverrides(cfg)

	// Validate configuration.
	if err := validate(cfg); err != nil {
		return nil, err
	}

	// Resolve the active cluster.
	resolveActiveCluster(cfg)

	return cfg, nil
}

// autoDetectWarewulf attempts to read /etc/warewulf/warewulf.conf to find
// the API endpoint configuration.
func autoDetectWarewulf() (*ClusterEntry, error) {
	wwConfPath := "/etc/warewulf/warewulf.conf"
	wv := viper.New()
	wv.SetConfigFile(wwConfPath)
	wv.SetConfigType("yaml")

	if err := wv.ReadInConfig(); err != nil {
		return nil, err
	}

	// Warewulf stores its listen port and address in its config.
	ipaddr := wv.GetString("ipaddr")
	port := wv.GetInt("warewulf.port")
	if port == 0 {
		port = 9873
	}
	if ipaddr == "" {
		ipaddr = "localhost"
	}

	endpoint := fmt.Sprintf("https://%s:%d", ipaddr, port)

	return &ClusterEntry{
		Name: "local",
		Cluster: ClusterConfig{
			Endpoint: endpoint,
			Insecure: true,
			Timeout:  "10s",
		},
	}, nil
}

// applyEnvOverrides applies W9S_* environment variables over the loaded config.
func applyEnvOverrides(cfg *Config) {
	if ep := os.Getenv("W9S_ENDPOINT"); ep != "" {
		if len(cfg.Clusters) == 0 {
			cfg.Clusters = append(cfg.Clusters, ClusterEntry{
				Name:    "env",
				Cluster: ClusterConfig{},
			})
		}
		cfg.Clusters[0].Cluster.Endpoint = ep
	}
	if user := os.Getenv("W9S_USERNAME"); user != "" {
		if len(cfg.Clusters) > 0 {
			cfg.Clusters[0].Cluster.Username = user
		}
	}
	if pass := os.Getenv("W9S_PASSWORD"); pass != "" {
		if len(cfg.Clusters) > 0 {
			cfg.Clusters[0].Cluster.Password = pass
		}
	}
}

// validate checks that the configuration is usable.
func validate(cfg *Config) error {
	if len(cfg.Clusters) == 0 {
		return fmt.Errorf("no clusters configured; set W9S_ENDPOINT or create a config file")
	}
	for i, entry := range cfg.Clusters {
		if strings.TrimSpace(entry.Cluster.Endpoint) == "" {
			return fmt.Errorf("cluster %d (%q) has no endpoint", i, entry.Name)
		}
	}
	return nil
}

// resolveActiveCluster sets the ActiveCluster field based on DefaultCluster
// or falls back to the first entry.
func resolveActiveCluster(cfg *Config) {
	if cfg.DefaultCluster != "" {
		for i := range cfg.Clusters {
			if cfg.Clusters[i].Name == cfg.DefaultCluster {
				cfg.ActiveCluster = &cfg.Clusters[i].Cluster
				return
			}
		}
	}
	// Fall back to the first cluster.
	if len(cfg.Clusters) > 0 {
		cfg.ActiveCluster = &cfg.Clusters[0].Cluster
	}
}
