// Package version holds build-time version information injected via ldflags.
package version

import "fmt"

// These variables are set at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)

// Info returns a formatted version string.
func Info() string {
	return fmt.Sprintf("w9s %s (commit: %s, built: %s by %s)", Version, Commit, Date, BuiltBy)
}
