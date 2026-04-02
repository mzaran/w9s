// Package cli provides the Cobra command-line interface for w9s.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/mzaran/w9s/internal/app"
	"github.com/mzaran/w9s/internal/config"
	"github.com/mzaran/w9s/internal/dao"
	"github.com/mzaran/w9s/internal/version"
	"github.com/mzaran/w9s/pkg/mock"
)

var (
	cfgFile  string
	debug    bool
	useMock  bool
	readOnly bool
)

// rootCmd is the base command for w9s.
var rootCmd = &cobra.Command{
	Use:     "w9s",
	Short:   "Terminal UI for Warewulf cluster management",
	Version: version.Version,
	RunE:    runApp,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/w9s/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	rootCmd.PersistentFlags().BoolVar(&useMock, "mock", false, "use mock data (requires W9S_ENABLE_MOCK=1)")
	// Only show --mock when the env var is set.
	if os.Getenv("W9S_ENABLE_MOCK") != "1" {
		_ = rootCmd.PersistentFlags().MarkHidden("mock")
	}

	rootCmd.PersistentFlags().BoolVar(&readOnly, "readonly", false, "disable all destructive actions (add, delete, edit, build, import)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(infoCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// runApp is the main entry point that sets up the app and runs it.
func runApp(cmd *cobra.Command, _ []string) error {
	// Load configuration.
	var cfg *config.Config
	var err error

	// In mock mode, we don't require a real config.
	if useMock && os.Getenv("W9S_ENABLE_MOCK") == "1" {
		cfg = &config.Config{
			RefreshRate: "5s",
			ReadOnly:    readOnly,
			Clusters: []config.ClusterEntry{
				{
					Name: "mock-1",
					Cluster: config.ClusterConfig{
						Endpoint: "mock://cluster-1",
					},
				},
				{
					Name: "mock-2",
					Cluster: config.ClusterConfig{
						Endpoint: "mock://cluster-2",
					},
				},
			},
			UI: config.UIConfig{EnableMouse: true},
		}
		cfg.ActiveCluster = &cfg.Clusters[0].Cluster
	} else {
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// CLI flag overrides config file.
	if readOnly {
		cfg.ReadOnly = true
	}

	// Create the appropriate client.
	var a *app.App
	if useMock && os.Getenv("W9S_ENABLE_MOCK") == "1" {
		client := mock.NewMockClient()
		a, err = app.New(client, cfg)
	} else {
		cluster := cfg.ActiveCluster
		timeout := 30 * time.Second
		if cluster.Timeout != "" {
			if d, parseErr := time.ParseDuration(cluster.Timeout); parseErr == nil {
				timeout = d
			}
		}
		opts := []dao.ClientOption{
			dao.WithEndpoint(cluster.Endpoint),
			dao.WithInsecure(cluster.Insecure),
			dao.WithTimeout(timeout),
		}
		if cluster.Username != "" {
			opts = append(opts, dao.WithBasicAuth(cluster.Username, cluster.Password))
		}
		if cluster.PowerTimeout != "" {
			if pt, ptErr := time.ParseDuration(cluster.PowerTimeout); ptErr == nil {
				opts = append(opts, dao.WithPowerTimeout(pt))
			}
		}
		client, clientErr := dao.NewHTTPClient(opts...)
		if clientErr != nil {
			return fmt.Errorf("failed to create client: %w", clientErr)
		}
		defer client.Close()

		// Verify API is reachable before launching TUI.
		if _, connErr := client.Nodes().List(); connErr != nil {
			return fmt.Errorf("cannot connect to Warewulf API at %s: %w", cluster.Endpoint, connErr)
		}

		a, err = app.New(client, cfg)
	}
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	// Set up signal handling for graceful shutdown.
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigCh:
			a.Stop()
			cancel()
		case <-ctx.Done():
		}
	}()

	// Run the TUI (blocks until exit).
	if err := a.Run(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	return nil
}
