package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mzaran/w9s/internal/config"
	"github.com/mzaran/w9s/internal/version"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print w9s configuration info",
	Run: func(cmd *cobra.Command, args []string) {
		printLogo()
		const fmat = "%-18s %s\n"
		fmt.Printf(fmat, "Version:", version.Info())
		fmt.Printf(fmat, "Repo:", version.GitRepo)

		// Config file location.
		configPath := cfgFile
		if configPath == "" {
			if home, err := os.UserHomeDir(); err == nil {
				xdg := filepath.Join(home, ".config", "w9s", "config.yaml")
				dot := filepath.Join(home, ".w9s", "config.yaml")
				if _, err := os.Stat(xdg); err == nil {
					configPath = xdg
				} else if _, err := os.Stat(dot); err == nil {
					configPath = dot
				} else {
					configPath = xdg + " (not found)"
				}
			}
		}
		fmt.Printf(fmat, "Config:", configPath)

		// Skins directory.
		if home, err := os.UserHomeDir(); err == nil {
			fmt.Printf(fmat, "Skins:", filepath.Join(home, ".config", "w9s", "skins")+"/")
		}

		// Try loading config for active cluster info.
		cfg, err := config.Load(cfgFile)
		if err == nil && cfg.ActiveCluster != nil {
			for _, c := range cfg.Clusters {
				if c.Cluster.Endpoint == cfg.ActiveCluster.Endpoint {
					fmt.Printf(fmat, "Active Cluster:", fmt.Sprintf("%s (%s)", c.Name, c.Cluster.Endpoint))
					break
				}
			}
			if cfg.UI.Skin != "" {
				fmt.Printf(fmat, "Skin:", cfg.UI.Skin)
			}
			fmt.Printf(fmat, "Read-Only:", fmt.Sprintf("%t", cfg.ReadOnly))
		}
	},
}
