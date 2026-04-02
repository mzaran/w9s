package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_Use(t *testing.T) {
	assert.Equal(t, "w9s", rootCmd.Use)
}

func TestRootCmd_Short(t *testing.T) {
	assert.Equal(t, "Terminal UI for Warewulf cluster management", rootCmd.Short)
}

func TestRootCmd_ConfigFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("config")
	require.NotNil(t, f, "expected 'config' persistent flag to exist")
	assert.Equal(t, "string", f.Value.Type())
	assert.Equal(t, "", f.DefValue)
}

func TestRootCmd_DebugFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("debug")
	require.NotNil(t, f, "expected 'debug' persistent flag to exist")
	assert.Equal(t, "bool", f.Value.Type())
	assert.Equal(t, "false", f.DefValue)
}

func TestRootCmd_MockFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("mock")
	require.NotNil(t, f, "expected 'mock' persistent flag to exist")
	assert.Equal(t, "bool", f.Value.Type())
	assert.Equal(t, "false", f.DefValue)
}

func TestRootCmd_HasVersionSubcommand(t *testing.T) {
	var found bool
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected rootCmd to have a 'version' subcommand")
}

func TestRootCmd_HasInfoSubcommand(t *testing.T) {
	var found bool
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "info" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected rootCmd to have an 'info' subcommand")
}

func TestVersionCmd_Use(t *testing.T) {
	assert.Equal(t, "version", versionCmd.Use)
	assert.Equal(t, "Print the version information", versionCmd.Short)
}

func TestExecute_VersionDoesNotPanic(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})

	// Reset args after test so other tests are not affected.
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
	})

	assert.NotPanics(t, func() {
		err := Execute()
		require.NoError(t, err)
	})
}
