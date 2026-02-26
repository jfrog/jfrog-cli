package packagealias

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"gopkg.in/yaml.v3"
)

const (
	configLockFileName  = ".config.lock"
	configLockTimeout   = 5 * time.Second
	configLockRetryWait = 50 * time.Millisecond

	configLockTimeoutEnv = "JFROG_CLI_PACKAGE_ALIAS_LOCK_TIMEOUT"
)

func newDefaultConfig() *Config {
	return &Config{
		Enabled:         true,
		ToolModes:       make(map[string]AliasMode, len(SupportedTools)),
		SubcommandModes: make(map[string]AliasMode),
	}
}

func getConfigPath(aliasDir string) string {
	return filepath.Join(aliasDir, configFile)
}

func loadConfig(aliasDir string) (*Config, error) {
	config := newDefaultConfig()
	configPath := getConfigPath(aliasDir)
	data, readErr := os.ReadFile(configPath)
	if readErr == nil {
		if unmarshalErr := yaml.Unmarshal(data, config); unmarshalErr != nil {
			return nil, fmt.Errorf("failed parsing %s: %w", configPath, unmarshalErr)
		}
		return normalizeConfig(config), nil
	}
	if !os.IsNotExist(readErr) {
		return nil, errorutils.CheckError(readErr)
	}
	return normalizeConfig(config), nil
}

func normalizeConfig(config *Config) *Config {
	if config == nil {
		return newDefaultConfig()
	}
	if config.ToolModes == nil {
		config.ToolModes = make(map[string]AliasMode, len(SupportedTools))
	}
	if config.SubcommandModes == nil {
		config.SubcommandModes = make(map[string]AliasMode)
	}
	return config
}

func writeConfig(aliasDir string, config *Config) error {
	config = normalizeConfig(config)
	configPath := getConfigPath(aliasDir)
	return writeYAMLAtomic(configPath, config)
}

func writeYAMLAtomic(path string, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return writeBytesAtomic(path, yamlData, ".tmp-config-*.yaml")
}

func writeBytesAtomic(path string, content []byte, tempPattern string) error {
	dirPath := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dirPath, tempPattern)
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()

	defer func() {
		if _, statErr := os.Stat(tempPath); statErr == nil {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err = tempFile.Write(content); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err = tempFile.Chmod(0644); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err = tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err = tempFile.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func withConfigLock(aliasDir string, action func() error) error {
	lockPath := filepath.Join(aliasDir, configLockFileName)
	lockTimeout := getConfigLockTimeout()
	deadline := time.Now().Add(lockTimeout)
	for {
		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			_ = lockFile.Close()
			defer func() {
				_ = os.Remove(lockPath)
			}()
			return action()
		}
		if !os.IsExist(err) {
			return errorutils.CheckError(err)
		}
		if time.Now().After(deadline) {
			return errorutils.CheckError(fmt.Errorf(
				"timed out waiting for config lock: %s. If this is from a crashed process in CI, remove it and retry. You can tune timeout with %s",
				lockPath,
				configLockTimeoutEnv,
			))
		}
		time.Sleep(configLockRetryWait)
	}
}

func getConfigLockTimeout() time.Duration {
	return getDurationFromEnv(configLockTimeoutEnv, configLockTimeout)
}

func getDurationFromEnv(envVarName string, defaultValue time.Duration) time.Duration {
	rawValue := strings.TrimSpace(os.Getenv(envVarName))
	if rawValue == "" {
		return defaultValue
	}
	parsedValue, err := time.ParseDuration(rawValue)
	if err != nil || parsedValue <= 0 {
		log.Warn(fmt.Sprintf("Invalid %s value '%s'. Falling back to default %s.", envVarName, rawValue, defaultValue))
		return defaultValue
	}
	return parsedValue
}

func parsePackageList(packageList string) ([]string, error) {
	if strings.TrimSpace(packageList) == "" {
		return append([]string(nil), SupportedTools...), nil
	}

	uniquePackages := make(map[string]struct{})
	selectedPackages := make([]string, 0)
	for _, rawPackage := range strings.Split(packageList, ",") {
		normalizedPackage := strings.ToLower(strings.TrimSpace(rawPackage))
		if normalizedPackage == "" {
			continue
		}
		if !isSupportedTool(normalizedPackage) {
			return nil, errorutils.CheckError(fmt.Errorf("unsupported package manager: %s. Supported package managers: %s", normalizedPackage, strings.Join(SupportedTools, ", ")))
		}
		if _, exists := uniquePackages[normalizedPackage]; exists {
			continue
		}
		uniquePackages[normalizedPackage] = struct{}{}
		selectedPackages = append(selectedPackages, normalizedPackage)
	}
	if len(selectedPackages) == 0 {
		return nil, errorutils.CheckError(fmt.Errorf("no valid packages provided for --packages"))
	}
	return selectedPackages, nil
}

func isSupportedTool(tool string) bool {
	for _, supportedTool := range SupportedTools {
		if tool == supportedTool {
			return true
		}
	}
	return false
}

func getConfiguredTools(config *Config) []string {
	config = normalizeConfig(config)
	if len(config.EnabledTools) == 0 {
		return append([]string(nil), SupportedTools...)
	}
	configured := make([]string, 0, len(config.EnabledTools))
	for _, tool := range config.EnabledTools {
		normalizedTool := strings.ToLower(strings.TrimSpace(tool))
		if normalizedTool == "" {
			continue
		}
		if !isSupportedTool(normalizedTool) {
			continue
		}
		configured = append(configured, normalizedTool)
	}
	if len(configured) == 0 {
		return append([]string(nil), SupportedTools...)
	}
	return configured
}

func isConfiguredTool(config *Config, tool string) bool {
	for _, configuredTool := range getConfiguredTools(config) {
		if configuredTool == tool {
			return true
		}
	}
	return false
}

func validateAliasMode(mode AliasMode) bool {
	switch mode {
	case ModeJF, ModeEnv, ModePass:
		return true
	default:
		return false
	}
}

func getGoSubcommandPolicyKeys(args []string) []string {
	if len(args) == 0 {
		return []string{"go"}
	}
	subcommandParts := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		subcommandParts = append(subcommandParts, strings.ToLower(arg))
	}
	if len(subcommandParts) == 0 {
		return []string{"go"}
	}
	keys := make([]string, 0, len(subcommandParts)+1)
	for index := len(subcommandParts); index >= 1; index-- {
		keys = append(keys, "go."+strings.Join(subcommandParts[:index], "."))
	}
	keys = append(keys, "go")
	return keys
}

func getModeForTool(config *Config, tool string, args []string) AliasMode {
	config = normalizeConfig(config)
	if tool == "go" {
		keys := getGoSubcommandPolicyKeys(args)
		for _, key := range keys {
			mode, found := config.SubcommandModes[key]
			if found {
				if validateAliasMode(mode) {
					return mode
				}
				log.Warn(fmt.Sprintf("Invalid subcommand mode '%s' for key '%s'. Falling back to defaults.", mode, key))
			}
			mode, found = config.ToolModes[key]
			if found {
				if validateAliasMode(mode) {
					return mode
				}
				log.Warn(fmt.Sprintf("Invalid tool mode '%s' for key '%s'. Falling back to defaults.", mode, key))
			}
		}
		return ModeJF
	}

	mode, found := config.ToolModes[tool]
	if !found {
		return ModeJF
	}
	if !validateAliasMode(mode) {
		log.Warn(fmt.Sprintf("Invalid mode '%s' for tool '%s'. Falling back to default.", mode, tool))
		return ModeJF
	}
	return mode
}

func getEnabledState(aliasDir string) bool {
	config, err := loadConfig(aliasDir)
	if err != nil {
		return true
	}
	return config.Enabled
}

func computeFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func addExecutableSuffix(tool string) string {
	if runtime.GOOS == "windows" {
		return tool + ".exe"
	}
	return tool
}
