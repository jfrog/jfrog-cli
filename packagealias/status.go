package packagealias

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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
	configPath := filepath.Join(aliasDir, configFile)

	if data, err := os.ReadFile(configPath); err == nil {
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err == nil {
			for _, tool := range SupportedTools {
				mode := cfg.ToolModes[tool]
				if mode == "" {
					mode = ModeJF
				}

				// Check if alias exists
				aliasPath := filepath.Join(binDir, tool)
				if runtime.GOOS == "windows" {
					aliasPath += ".exe"
				}

				aliasExists := "✓"
				if _, err := os.Stat(aliasPath); os.IsNotExist(err) {
					aliasExists = "✗"
				}

				// Check if real tool exists
				realExists := "✓"
				if _, err := exec.LookPath(tool); err != nil {
					realExists = "✗"
				}

				log.Info(fmt.Sprintf("  %-10s mode=%-5s alias=%s real=%s", tool, mode, aliasExists, realExists))
			}
		}
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
