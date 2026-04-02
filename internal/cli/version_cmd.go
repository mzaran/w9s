package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mzaran/w9s/internal/version"
)

var versionShort bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		if versionShort {
			fmt.Println(version.Info())
			return
		}
		printLogo()
		fmt.Printf("%-12s %s\n", "Version:", version.Info())
		fmt.Printf("%-12s %s\n", "Repo:", version.GitRepo)
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "Print version in short format")
}

// printLogo prints the ASCII logo in green (warewulf branding).
func printLogo() {
	for _, line := range version.LogoSmall {
		fmt.Printf("\033[32m%s\033[0m\n", line)
	}
	fmt.Println()
}
