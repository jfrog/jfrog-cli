package packagealias

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type InstallCommand struct {
}

func NewInstallCommand() *InstallCommand {
	return &InstallCommand{}
}

func (ic *InstallCommand) CommandName() string {
	return "package_alias_install"
}

func (ic *InstallCommand) Run() error {
	// 1. Create alias directories
	aliasDir, err := GetAliasHomeDir()
	if err != nil {
		return err
	}
	binDir, err := GetAliasBinDir()
	if err != nil {
		return err
	}

	log.Info("Creating package alias directories...")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return errorutils.CheckError(err)
	}

	// 2. Get the path of the current executable
	jfPath, err := os.Executable()
	if err != nil {
		return errorutils.CheckError(fmt.Errorf("could not determine executable path: %w", err))
	}
	// Resolve any symlinks to get the real path
	jfPath, err = filepath.EvalSymlinks(jfPath)
	if err != nil {
		return errorutils.CheckError(fmt.Errorf("could not resolve executable path: %w", err))
	}
	log.Debug(fmt.Sprintf("Using jf binary at: %s", jfPath))

	// 3. Create symlinks/copies for each supported tool
	createdCount := 0
	for _, tool := range SupportedTools {
		// Create alias
		aliasPath := filepath.Join(binDir, tool)
		if runtime.GOOS == "windows" {
			aliasPath += ".exe"
			// On Windows, we need to copy the binary
			if err := copyFile(jfPath, aliasPath); err != nil {
				log.Warn(fmt.Sprintf("Failed to create alias for %s: %v", tool, err))
				continue
			}
		} else {
			// On Unix, create symlink
			// Remove existing symlink if any
			os.Remove(aliasPath)
			if err := os.Symlink(jfPath, aliasPath); err != nil {
				log.Warn(fmt.Sprintf("Failed to create alias for %s: %v", tool, err))
				continue
			}
		}
		createdCount++
		log.Debug(fmt.Sprintf("Created alias: %s -> %s", aliasPath, jfPath))
	}

	// 4. Create default config
	config := &Config{
		Enabled:   true,
		ToolModes: make(map[string]AliasMode),
	}
	// Set default modes
	for _, tool := range SupportedTools {
		config.ToolModes[tool] = ModeJF
	}
	configPath := filepath.Join(aliasDir, configFile)
	if err := saveJSON(configPath, config); err != nil {
		return errorutils.CheckError(err)
	}

	// 5. Create enabled state
	state := &State{Enabled: true}
	statePath := filepath.Join(aliasDir, stateFile)
	if err := saveJSON(statePath, state); err != nil {
		return errorutils.CheckError(err)
	}

	// Success message
	log.Info(fmt.Sprintf("Created %d aliases in %s", createdCount, binDir))
	log.Info("\nTo enable package aliasing, add this to your shell configuration:")

	if runtime.GOOS == "windows" {
		log.Info(fmt.Sprintf("  set PATH=%s;%%PATH%%", binDir))
	} else {
		log.Info(fmt.Sprintf("  export PATH=\"%s:$PATH\"", binDir))
		log.Info("\nThen run: hash -r")
	}
	log.Info("\nPackage aliasing is now installed. Run 'jf package-alias status' to verify.")

	return nil
}

func (ic *InstallCommand) SetRepo(repo string) *InstallCommand {
	return ic
}

func (ic *InstallCommand) ServerDetails() (*config.ServerDetails, error) {
	return nil, nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func saveJSON(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}
