package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mzaran/w9s/internal/version"
)

// versionCmd prints the w9s version information.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Info())
	},
}
