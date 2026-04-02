// Package version holds build-time version information injected via ldflags.
package version

const GitRepo = "github.com/mzaran/w9s"

// These variables are set at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// LogoSmall is the W9S ASCII art logo (Graffiti font + wolf face).
var LogoSmall = []string{
	`__      __  ________  _________`,
	`/  \    /  \/   __   \/   _____/   /\_/\`,
	`\   \/\/   /\____    /\_____  \   ( o.o )`,
	` \        /    /    / /        \   > ^ <`,
	`  \__/\  /    /____/ /_______  /`,
	`       \/                    \/`,
}

// Info returns the version string.
func Info() string {
	return Version
}

// Short returns just the semver tag, stripping git metadata like
// "-8-gb3da363-dirty" that makes the string too long for TUI headers.
func Short() string {
	v := Version
	// Strip "-dirty" suffix.
	if i := len(v) - 6; i > 0 && v[i:] == "-dirty" {
		v = v[:i]
	}
	// Strip "-N-gHASH" git describe suffix.
	// Format: vX.Y.Z-N-gHASH where N is commit count and gHASH is the hash.
	for i := len(v) - 1; i >= 0; i-- {
		if v[i] == '-' {
			// Check if this looks like "-gHASH"
			if i+1 < len(v) && v[i+1] == 'g' {
				// Find the previous dash for the commit count.
				for j := i - 1; j >= 0; j-- {
					if v[j] == '-' {
						return v[:j]
					}
				}
			}
			break
		}
	}
	return v
}
