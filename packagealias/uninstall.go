package packagealias

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type UninstallCommand struct {
}

func NewUninstallCommand() *UninstallCommand {
	return &UninstallCommand{}
}

func (uc *UninstallCommand) CommandName() string {
	return "package_alias_uninstall"
}

func (uc *UninstallCommand) Run() error {
	binDir, err := GetAliasBinDir()
	if err != nil {
		return err
	}

	// Check if alias directory exists
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		log.Info("Package aliases are not installed.")
		return nil
	}

	// Remove all aliases
	removedCount := 0
	for _, tool := range SupportedTools {
		aliasPath := filepath.Join(binDir, tool)
		if runtime.GOOS == "windows" {
			aliasPath += ".exe"
		}

		if err := os.Remove(aliasPath); err != nil {
			if !os.IsNotExist(err) {
				log.Debug(fmt.Sprintf("Failed to remove %s: %v", aliasPath, err))
			}
		} else {
			removedCount++
			log.Debug(fmt.Sprintf("Removed alias: %s", aliasPath))
		}
	}

	// Remove the entire package-alias directory
	aliasDir, err := GetAliasHomeDir()
	if err == nil {
		if err := os.RemoveAll(aliasDir); err != nil {
			log.Warn(fmt.Sprintf("Failed to remove alias directory: %v", err))
		}
	}

	log.Info(fmt.Sprintf("Removed %d aliases", removedCount))
	log.Info("\nTo complete uninstallation, remove this from your shell configuration:")

	if runtime.GOOS == "windows" {
		log.Info(fmt.Sprintf("  Remove '%s' from your PATH environment variable", binDir))
	} else {
		log.Info(fmt.Sprintf("  Remove 'export PATH=\"%s:$PATH\"' from your shell rc file", binDir))
		log.Info("\nThen run: hash -r")
	}

	return nil
}

func (uc *UninstallCommand) SetRepo(repo string) *UninstallCommand {
	return uc
}

func (uc *UninstallCommand) ServerDetails() (*config.ServerDetails, error) {
	return nil, nil
}
