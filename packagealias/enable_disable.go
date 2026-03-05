package packagealias

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// EnableCommand enables package aliasing
type EnableCommand struct {
}

func NewEnableCommand() *EnableCommand {
	return &EnableCommand{}
}

func (ec *EnableCommand) CommandName() string {
	return "package_alias_enable"
}

func (ec *EnableCommand) Run() error {
	return setEnabledState(true)
}

func (ec *EnableCommand) SetRepo(repo string) *EnableCommand {
	return ec
}

func (ec *EnableCommand) ServerDetails() (*config.ServerDetails, error) {
	return nil, nil
}

// DisableCommand disables package aliasing
type DisableCommand struct {
}

func NewDisableCommand() *DisableCommand {
	return &DisableCommand{}
}

func (dc *DisableCommand) CommandName() string {
	return "package_alias_disable"
}

func (dc *DisableCommand) Run() error {
	return setEnabledState(false)
}

func (dc *DisableCommand) SetRepo(repo string) *DisableCommand {
	return dc
}

func (dc *DisableCommand) ServerDetails() (*config.ServerDetails, error) {
	return nil, nil
}

// setEnabledState updates the enabled state
func setEnabledState(enabled bool) error {
	aliasDir, err := GetAliasHomeDir()
	if err != nil {
		return err
	}

	// Check if aliases are installed
	binDir := filepath.Join(aliasDir, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return errorutils.CheckError(fmt.Errorf("package aliases are not installed. Run 'jf package-alias install' first"))
	}

	// Update config
	if err := withConfigLock(aliasDir, func() error {
		config, loadErr := loadConfig(aliasDir)
		if loadErr != nil {
			return loadErr
		}
		config.Enabled = enabled
		return writeConfig(aliasDir, config)
	}); err != nil {
		return errorutils.CheckError(err)
	}

	if enabled {
		log.Info("Package aliasing is now ENABLED")
		log.Info("All supported package manager commands will be intercepted by JFrog CLI")
	} else {
		log.Info("Package aliasing is now DISABLED")
		log.Info("Package manager commands will run natively without JFrog CLI interception")
	}

	return nil
}
