package packagealias

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetToolModeFromConfig(t *testing.T) {
	testHomeDir := t.TempDir()
	t.Setenv("JFROG_CLI_HOME_DIR", testHomeDir)

	aliasDir := filepath.Join(testHomeDir, "package-alias")
	require.NoError(t, os.MkdirAll(aliasDir, 0755))

	config := &Config{
		ToolModes: map[string]AliasMode{
			"npm": ModePass,
			"go":  ModeJF,
		},
		SubcommandModes: map[string]AliasMode{
			"go.mod.tidy": ModePass,
			"go.mod":      ModeJF,
		},
		EnabledTools: []string{"npm", "go"},
	}
	require.NoError(t, writeConfig(aliasDir, config))

	require.Equal(t, ModePass, getToolMode("npm", []string{"install"}))
	require.Equal(t, ModePass, getToolMode("go", []string{"mod", "tidy"}))
	require.Equal(t, ModeJF, getToolMode("go", []string{"mod", "download"}))
}

func TestGetToolModeInvalidFallsBackToDefault(t *testing.T) {
	testHomeDir := t.TempDir()
	t.Setenv("JFROG_CLI_HOME_DIR", testHomeDir)

	aliasDir := filepath.Join(testHomeDir, "package-alias")
	require.NoError(t, os.MkdirAll(aliasDir, 0755))

	config := &Config{
		ToolModes: map[string]AliasMode{
			"npm": AliasMode("invalid"),
		},
	}
	require.NoError(t, writeConfig(aliasDir, config))

	require.Equal(t, ModeJF, getToolMode("npm", []string{"install"}))
}
