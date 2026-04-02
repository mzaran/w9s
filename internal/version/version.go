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
