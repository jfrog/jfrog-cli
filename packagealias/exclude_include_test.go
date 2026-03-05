package packagealias

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetToolModeRejectsToolNotInConfiguredList(t *testing.T) {
	testHomeDir := t.TempDir()
	t.Setenv("JFROG_CLI_HOME_DIR", testHomeDir)

	aliasDir := filepath.Join(testHomeDir, "package-alias")
	binDir := filepath.Join(aliasDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	config := &Config{
		ToolModes: map[string]AliasMode{
			"mvn": ModeJF,
		},
		EnabledTools: []string{"mvn"},
	}
	require.NoError(t, writeConfig(aliasDir, config))

	err := setToolMode("npm", ModePass)
	require.Error(t, err)
}

func TestSetToolModeConcurrentUpdates(t *testing.T) {
	testHomeDir := t.TempDir()
	t.Setenv("JFROG_CLI_HOME_DIR", testHomeDir)

	aliasDir := filepath.Join(testHomeDir, "package-alias")
	binDir := filepath.Join(aliasDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	config := &Config{
		ToolModes: map[string]AliasMode{
			"mvn": ModeJF,
			"npm": ModeJF,
		},
		EnabledTools: []string{"mvn", "npm"},
	}
	require.NoError(t, writeConfig(aliasDir, config))

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		require.NoError(t, setToolMode("mvn", ModePass))
	}()
	go func() {
		defer waitGroup.Done()
		require.NoError(t, setToolMode("npm", ModePass))
	}()
	waitGroup.Wait()

	updatedConfig, err := loadConfig(aliasDir)
	require.NoError(t, err)
	require.Equal(t, ModePass, updatedConfig.ToolModes["mvn"])
	require.Equal(t, ModePass, updatedConfig.ToolModes["npm"])
}

func TestSetToolModeStressConcurrentUpdates(t *testing.T) {
	testHomeDir := t.TempDir()
	t.Setenv("JFROG_CLI_HOME_DIR", testHomeDir)
	t.Setenv(configLockTimeoutEnv, "30s")

	aliasDir := filepath.Join(testHomeDir, "package-alias")
	binDir := filepath.Join(aliasDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	config := &Config{
		ToolModes: map[string]AliasMode{
			"mvn": ModeJF,
			"npm": ModeJF,
		},
		EnabledTools: []string{"mvn", "npm"},
	}
	require.NoError(t, writeConfig(aliasDir, config))

	var waitGroup sync.WaitGroup
	var errorCount atomic.Int32
	workers := 12
	iterations := 24
	for workerIndex := 0; workerIndex < workers; workerIndex++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			for iterationIndex := 0; iterationIndex < iterations; iterationIndex++ {
				targetTool := "mvn"
				if (index+iterationIndex)%2 == 0 {
					targetTool = "npm"
				}
				targetMode := ModeJF
				if (index+iterationIndex)%3 == 0 {
					targetMode = ModePass
				}
				if err := setToolMode(targetTool, targetMode); err != nil {
					errorCount.Add(1)
				}
			}
		}(workerIndex)
	}
	waitGroup.Wait()
	require.Equal(t, int32(0), errorCount.Load())

	updatedConfig, err := loadConfig(aliasDir)
	require.NoError(t, err)
	require.True(t, validateAliasMode(updatedConfig.ToolModes["mvn"]))
	require.True(t, validateAliasMode(updatedConfig.ToolModes["npm"]))
}
