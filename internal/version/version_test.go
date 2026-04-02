package version_test

import (
	"testing"

	"github.com/mzaran/w9s/internal/version"
	"github.com/stretchr/testify/assert"
)

func TestInfoDefaults(t *testing.T) {
	// Save originals and restore after test.
	origVersion := version.Version
	origCommit := version.Commit
	origDate := version.Date
	origBuiltBy := version.BuiltBy
	t.Cleanup(func() {
		version.Version = origVersion
		version.Commit = origCommit
		version.Date = origDate
		version.BuiltBy = origBuiltBy
	})

	version.Version = "dev"
	version.Commit = "none"
	version.Date = "unknown"
	version.BuiltBy = "unknown"

	info := version.Info()
	assert.Contains(t, info, "w9s")
	assert.Contains(t, info, "dev")
	assert.Contains(t, info, "none")
	assert.Contains(t, info, "unknown")
}

func TestInfoReflectsSetValues(t *testing.T) {
	origVersion := version.Version
	origCommit := version.Commit
	origDate := version.Date
	origBuiltBy := version.BuiltBy
	t.Cleanup(func() {
		version.Version = origVersion
		version.Commit = origCommit
		version.Date = origDate
		version.BuiltBy = origBuiltBy
	})

	version.Version = "1.2.3"
	version.Commit = "abc1234"
	version.Date = "2025-01-15"
	version.BuiltBy = "goreleaser"

	info := version.Info()
	assert.Equal(t, "w9s 1.2.3 (commit: abc1234, built: 2025-01-15 by goreleaser)", info)
}

func TestInfoFormat(t *testing.T) {
	origVersion := version.Version
	origCommit := version.Commit
	origDate := version.Date
	origBuiltBy := version.BuiltBy
	t.Cleanup(func() {
		version.Version = origVersion
		version.Commit = origCommit
		version.Date = origDate
		version.BuiltBy = origBuiltBy
	})

	version.Version = "v0.1.0"
	version.Commit = "deadbeef"
	version.Date = "2025-06-01"
	version.BuiltBy = "ci"

	expected := "w9s v0.1.0 (commit: deadbeef, built: 2025-06-01 by ci)"
	assert.Equal(t, expected, version.Info())
}
