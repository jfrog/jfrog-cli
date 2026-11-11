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

func TestGetHelmSubcommand(t *testing.T) {
	tests := []struct {
		args     []string
		expected string
	}{
		{[]string{"package", "./chart", "-d", "out"}, "package"},
		{[]string{"-n", "prod", "upgrade", "release", "./chart"}, "upgrade"},
		{[]string{"--kube-context", "ctx", "--debug", "push", "chart.tgz", "oci://repo"}, "push"},
		{[]string{"--namespace=prod", "template", "release", "./chart"}, "template"},
		{[]string{"--debug"}, ""},
		{nil, ""},
	}
	for _, test := range tests {
		require.Equal(t, test.expected, getHelmSubcommand(test.args), "args: %v", test.args)
	}
}

func TestGetModeForToolRoutesOnlyHelmBuildInfoCommandsToJf(t *testing.T) {
	config := &Config{}
	for _, args := range [][]string{
		{"package", "./chart"},
		{"push", "chart.tgz", "oci://repo"},
		{"dependency", "update"},
		{"-n", "prod", "package", "./chart"},
	} {
		require.Equal(t, ModeJF, getModeForTool(config, "helm", args), "args: %v", args)
	}
	for _, args := range [][]string{
		{"registry", "login", "ghcr.io"},
		{"registry", "logout", "ghcr.io"},
		{"-n", "prod", "template", "release", "./chart"},
		{"env"},
		{"version"},
		{},
	} {
		require.Equal(t, ModePass, getModeForTool(config, "helm", args), "args: %v", args)
	}

	config.ToolModes = map[string]AliasMode{"helm": ModePass}
	require.Equal(t, ModePass, getModeForTool(config, "helm", []string{"package", "./chart"}),
		"a configured pass mode should still apply to build-info commands")
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
