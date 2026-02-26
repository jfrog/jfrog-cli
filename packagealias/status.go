package packagealias

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type StatusCommand struct {
}

func NewStatusCommand() *StatusCommand {
	return &StatusCommand{}
}

func (sc *StatusCommand) CommandName() string {
	return "package_alias_status"
}

func (sc *StatusCommand) Run() error {
	log.Info("Package Alias Status")
	log.Info("===================")

	// Check if installed
	binDir, err := GetAliasBinDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		log.Info("Status: NOT INSTALLED")
		log.Info("\nRun 'jf package-alias install' to set up package aliasing")
		return nil
	}

	log.Info("Status: INSTALLED")
	log.Info(fmt.Sprintf("Location: %s", binDir))

	// Check if enabled
	enabled := isEnabled()
	if enabled {
		log.Info("State: ENABLED")
	} else {
		log.Info("State: DISABLED")
	}

	// Check if in PATH
	inPath := checkIfInPath(binDir)
	if inPath {
		log.Info("PATH: Configured ✓")
	} else {
		log.Warn("PATH: Not configured")
		if runtime.GOOS == "windows" {
			log.Info(fmt.Sprintf("\nAdd to PATH: set PATH=%s;%%PATH%%", binDir))
		} else {
			log.Info(fmt.Sprintf("\nAdd to PATH: export PATH=\"%s:$PATH\"", binDir))
		}
	}

	// Load and display configuration
	log.Info("\nTool Configuration:")
	aliasDir, _ := GetAliasHomeDir()
	cfg, cfgErr := loadConfig(aliasDir)
	if cfgErr != nil {
		log.Warn(fmt.Sprintf("Failed loading config for status: %v", cfgErr))
		cfg = newDefaultConfig()
	}
	for _, tool := range getConfiguredTools(cfg) {
		mode := getModeForTool(cfg, tool, nil)

		// Check if alias exists
		aliasPath := filepath.Join(binDir, addExecutableSuffix(tool))
		aliasExists := "✓"
		if _, err := os.Stat(aliasPath); os.IsNotExist(err) {
			aliasExists = "✗"
		}

		// Check if real tool exists
		realExists := "✓"
		if _, err := findRealToolPath(tool, binDir); err != nil {
			realExists = "✗"
		}

		log.Info(fmt.Sprintf("  %-10s mode=%-5s alias=%s real=%s", tool, mode, aliasExists, realExists))
		for _, detail := range getStatusModeDetails(cfg, tool) {
			log.Info(fmt.Sprintf("    %s", detail))
		}
	}

	if runtime.GOOS == "windows" {
		showWindowsStalenessWarning(cfg)
	}

	// Show example usage
	if enabled && inPath {
		log.Info("\nPackage aliasing is active. You can now run:")
		log.Info("  mvn install")
		log.Info("  npm install")
		log.Info("  go build")
		log.Info("...and they will be intercepted by JFrog CLI")
	}

	return nil
}

func (sc *StatusCommand) SetRepo(repo string) *StatusCommand {
	return sc
}

func (sc *StatusCommand) ServerDetails() (*config.ServerDetails, error) {
	return nil, nil
}

func findRealToolPath(tool, aliasBinDir string) (string, error) {
	filteredPath := FilterOutDirFromPATH(os.Getenv("PATH"), aliasBinDir)
	return lookPathInPathEnv(tool, filteredPath)
}

func showWindowsStalenessWarning(cfg *Config) {
	cfg = normalizeConfig(cfg)
	if cfg.JfBinarySHA256 == "" {
		return
	}
	jfPath, err := os.Executable()
	if err != nil {
		return
	}
	jfPath, err = filepath.EvalSymlinks(jfPath)
	if err != nil {
		return
	}
	currentHash, err := computeFileSHA256(jfPath)
	if err != nil {
		return
	}
	if currentHash == cfg.JfBinarySHA256 {
		return
	}
	log.Warn("Windows alias copies may be stale compared to current jf binary.")
	log.Warn("Run 'jf package-alias install' to refresh alias executables.")
}

// checkIfInPath checks if a directory is in PATH
func checkIfInPath(dir string) bool {
	pathEnv := os.Getenv("PATH")
	paths := filepath.SplitList(pathEnv)

	dir = filepath.Clean(dir)
	for _, p := range paths {
		if filepath.Clean(p) == dir {
			return true
		}
	}

	return false
}

func lookPathInPathEnv(fileName, pathEnv string) (string, error) {
	if strings.ContainsRune(fileName, os.PathSeparator) {
		if isExecutableFile(fileName) {
			return fileName, nil
		}
		return "", exec.ErrNotFound
	}
	fileExtensions := []string{""}
	if runtime.GOOS == "windows" && filepath.Ext(fileName) == "" {
		pathext := os.Getenv("PATHEXT")
		fileExtensions = parseWindowsPathExtensions(pathext)
	}
	for _, pathDir := range filepath.SplitList(pathEnv) {
		if pathDir == "" {
			continue
		}
		for _, extension := range fileExtensions {
			candidatePath := filepath.Join(pathDir, fileName+extension)
			if isExecutableFile(candidatePath) {
				return candidatePath, nil
			}
		}
	}
	return "", exec.ErrNotFound
}

func parseWindowsPathExtensions(pathExtValue string) []string {
	if strings.TrimSpace(pathExtValue) == "" {
		pathExtValue = ".COM;.EXE;.BAT;.CMD"
	}
	extensions := strings.Split(strings.ToLower(pathExtValue), ";")
	normalizedExtensions := make([]string, 0, len(extensions))
	for _, extension := range extensions {
		trimmedExtension := strings.TrimSpace(extension)
		if trimmedExtension == "" {
			continue
		}
		if !strings.HasPrefix(trimmedExtension, ".") {
			trimmedExtension = "." + trimmedExtension
		}
		normalizedExtensions = append(normalizedExtensions, trimmedExtension)
	}
	if len(normalizedExtensions) == 0 {
		return []string{".com", ".exe", ".bat", ".cmd"}
	}
	return normalizedExtensions
}

func getStatusModeDetails(config *Config, tool string) []string {
	if tool != "go" {
		return nil
	}
	config = normalizeConfig(config)
	policyKeys := make([]string, 0, len(config.SubcommandModes))
	for policyKey := range config.SubcommandModes {
		if strings.HasPrefix(policyKey, "go.") {
			policyKeys = append(policyKeys, policyKey)
		}
	}
	if len(policyKeys) == 0 {
		return nil
	}
	sort.Strings(policyKeys)
	modeDetails := make([]string, 0, len(policyKeys))
	for _, policyKey := range policyKeys {
		mode := getModeForTool(config, "go", strings.Split(strings.TrimPrefix(policyKey, "go."), "."))
		modeDetails = append(modeDetails, fmt.Sprintf("%s mode=%s", policyKey, mode))
	}
	return modeDetails
}

func isExecutableFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil || fileInfo.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return fileInfo.Mode()&0111 != 0
}
