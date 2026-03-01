package packagealias

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallPreservesExistingModesAndSelectsTools(t *testing.T) {
	testHomeDir := t.TempDir()
	t.Setenv("JFROG_CLI_HOME_DIR", testHomeDir)

	aliasDir := filepath.Join(testHomeDir, "package-alias")
	binDir := filepath.Join(aliasDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	initialConfig := `tool_modes:
  npm: pass
enabled_tools:
  - npm
  - mvn
`
	require.NoError(t, os.WriteFile(filepath.Join(aliasDir, configFile), []byte(initialConfig), 0600))

	command := NewInstallCommand("mvn,npm")
	require.NoError(t, command.Run())

	config, err := loadConfig(aliasDir)
	require.NoError(t, err)
	require.Equal(t, ModePass, config.ToolModes["npm"])
	require.Equal(t, ModeJF, config.ToolModes["mvn"])
	require.ElementsMatch(t, []string{"mvn", "npm"}, config.EnabledTools)
	require.True(t, config.Enabled)
}

func TestInstallRejectsUnsupportedTools(t *testing.T) {
	t.Setenv("JFROG_CLI_HOME_DIR", t.TempDir())
	command := NewInstallCommand("mvn,not-a-tool")
	err := command.Run()
	require.Error(t, err)
}

func TestFindRealToolPathFiltersAliasDirectory(t *testing.T) {
	aliasDir := t.TempDir()
	realDir := t.TempDir()
	toolName := "fake-tool"

	aliasToolPath := filepath.Join(aliasDir, toolName)
	realToolPath := filepath.Join(realDir, toolName)
	require.NoError(t, os.WriteFile(aliasToolPath, []byte("#!/bin/sh\necho alias\n"), 0755))
	require.NoError(t, os.WriteFile(realToolPath, []byte("#!/bin/sh\necho real\n"), 0755))

	t.Setenv("PATH", aliasDir+string(os.PathListSeparator)+realDir)

	foundPath, err := findRealToolPath(toolName, aliasDir)
	require.NoError(t, err)
	require.Equal(t, realToolPath, foundPath)
}

func TestParseWindowsPathExtensionsNormalizesValues(t *testing.T) {
	extensions := parseWindowsPathExtensions("EXE; .BAT;cmd;;")
	require.Equal(t, []string{".exe", ".bat", ".cmd"}, extensions)
}

func TestGetStatusModeDetailsShowsGoPolicies(t *testing.T) {
	config := &Config{
		ToolModes: map[string]AliasMode{
			"go": ModeJF,
		},
		SubcommandModes: map[string]AliasMode{
			"go.mod.tidy": ModePass,
			"go.mod":      ModeJF,
		},
	}
	details := getStatusModeDetails(config, "go")
	require.Equal(t, []string{
		"go.mod mode=jf",
		"go.mod.tidy mode=pass",
	}, details)
}
