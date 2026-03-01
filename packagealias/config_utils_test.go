package packagealias

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigReadsConfigFile(t *testing.T) {
	aliasDir := t.TempDir()
	configPath := filepath.Join(aliasDir, configFile)
	configContent := `enabled: false
tool_modes:
  npm: pass
subcommand_modes:
  go.mod.tidy: pass
enabled_tools:
  - npm
  - go
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0600))

	config, err := loadConfig(aliasDir)
	require.NoError(t, err)
	require.False(t, config.Enabled)
	require.Equal(t, ModePass, config.ToolModes["npm"])
	require.Equal(t, ModePass, config.SubcommandModes["go.mod.tidy"])
	require.ElementsMatch(t, []string{"npm", "go"}, config.EnabledTools)
}

func TestParsePackageList(t *testing.T) {
	packages, err := parsePackageList("mvn, npm, mvn,go")
	require.NoError(t, err)
	require.Equal(t, []string{"mvn", "npm", "go"}, packages)
}

func TestGetGoSubcommandPolicyKeys(t *testing.T) {
	keys := getGoSubcommandPolicyKeys([]string{"mod", "tidy", "-v"})
	require.Equal(t, []string{"go.mod.tidy", "go.mod", "go"}, keys)
}

func TestGetModeForToolUsesMostSpecificGoPolicy(t *testing.T) {
	config := &Config{
		ToolModes: map[string]AliasMode{
			"go": ModeJF,
		},
		SubcommandModes: map[string]AliasMode{
			"go.mod":      ModeJF,
			"go.mod.tidy": ModePass,
		},
	}

	mode := getModeForTool(config, "go", []string{"mod", "tidy"})
	require.Equal(t, ModePass, mode)

	mode = getModeForTool(config, "go", []string{"mod", "download"})
	require.Equal(t, ModeJF, mode)
}

func TestGetModeForToolFallsBackForInvalidModes(t *testing.T) {
	config := &Config{
		ToolModes: map[string]AliasMode{
			"npm": AliasMode("invalid"),
		},
	}
	mode := getModeForTool(config, "npm", []string{"install"})
	require.Equal(t, ModeJF, mode)
}

func TestWriteConfigCreatesYamlConfig(t *testing.T) {
	aliasDir := t.TempDir()
	config := &Config{
		ToolModes: map[string]AliasMode{
			"npm": ModePass,
		},
	}

	require.NoError(t, writeConfig(aliasDir, config))
	_, err := os.Stat(filepath.Join(aliasDir, configFile))
	require.NoError(t, err)
}

func TestGetDurationFromEnv(t *testing.T) {
	require.Equal(t, configLockTimeout, getDurationFromEnv("NON_EXISTENT_ENV"))

	t.Setenv(configLockTimeoutEnv, "2s")
	require.Equal(t, 2*time.Second, getDurationFromEnv(configLockTimeoutEnv))

	t.Setenv(configLockTimeoutEnv, "bad-value")
	require.Equal(t, configLockTimeout, getDurationFromEnv(configLockTimeoutEnv))
}

func TestWithConfigLockTimeoutWhenLockExists(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), configLockFileName)
	require.NoError(t, os.WriteFile(lockPath, []byte("locked"), 0600))
	t.Setenv(configLockTimeoutEnv, "100ms")

	err := withConfigLock(filepath.Dir(lockPath), func() error {
		return nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), configLockTimeoutEnv)
	require.Contains(t, err.Error(), "remove it and retry")
}

func TestWithConfigLockReleasesLockFile(t *testing.T) {
	aliasDir := t.TempDir()
	require.NoError(t, withConfigLock(aliasDir, func() error {
		_, err := os.Stat(filepath.Join(aliasDir, configLockFileName))
		require.NoError(t, err)
		return nil
	}))
	_, err := os.Stat(filepath.Join(aliasDir, configLockFileName))
	require.True(t, os.IsNotExist(err))
}

func TestGetEnabledStateDefaultWhenConfigMissing(t *testing.T) {
	aliasDir := t.TempDir()
	require.True(t, getEnabledState(aliasDir))
}

func TestGetEnabledStateFalseWhenDisabledInConfig(t *testing.T) {
	aliasDir := t.TempDir()
	require.NoError(t, writeConfig(aliasDir, &Config{
		Enabled: false,
	}))
	require.False(t, getEnabledState(aliasDir))
}
