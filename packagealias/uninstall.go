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

	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		log.Info("Package aliases are not installed.")
		return nil
	}

	aliasDir, err := GetAliasHomeDir()
	if err != nil {
		return err
	}

	var removedCount int

	// Hold the lock while removing aliases and the directory so a
	// concurrent install doesn't recreate files mid-removal.
	lockErr := withConfigLock(aliasDir, func() error {
		for _, tool := range SupportedTools {
			aliasPath := filepath.Join(binDir, tool)
			if runtime.GOOS == "windows" {
				aliasPath += ".exe"
			}

			if removeErr := os.Remove(aliasPath); removeErr != nil {
				if !os.IsNotExist(removeErr) {
					log.Debug(fmt.Sprintf("Failed to remove %s: %v", aliasPath, removeErr))
				}
			} else {
				removedCount++
				log.Debug(fmt.Sprintf("Removed alias: %s", aliasPath))
			}
		}
		return nil
	})

	// Remove the entire directory tree after releasing the lock (the lock
	// file itself lives inside aliasDir, so we can't delete it while held).
	if removeErr := os.RemoveAll(aliasDir); removeErr != nil {
		log.Warn(fmt.Sprintf("Failed to remove alias directory: %v", removeErr))
	}

	if lockErr != nil {
		return lockErr
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
