package packagealias

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type InstallCommand struct {
	packagesArg string
}

func NewInstallCommand(packagesArg string) *InstallCommand {
	return &InstallCommand{packagesArg: packagesArg}
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

	selectedTools, err := parsePackageList(ic.packagesArg)
	if err != nil {
		return err
	}

	// 3. Create symlinks/copies for selected tools and remove unselected aliases
	selectedToolsSet := make(map[string]struct{}, len(selectedTools))
	for _, tool := range selectedTools {
		selectedToolsSet[tool] = struct{}{}
	}

	createdCount := 0
	for _, tool := range SupportedTools {
		aliasPath := filepath.Join(binDir, tool)
		if runtime.GOOS == "windows" {
			aliasPath += ".exe"
		}

		if _, shouldInstall := selectedToolsSet[tool]; !shouldInstall {
			if removeErr := os.Remove(aliasPath); removeErr != nil && !os.IsNotExist(removeErr) {
				log.Warn(fmt.Sprintf("Failed to remove alias for %s: %v", tool, removeErr))
			}
			continue
		}

		if runtime.GOOS == "windows" {
			// On Windows, we need to copy the binary
			if copyErr := copyFile(jfPath, aliasPath); copyErr != nil {
				log.Warn(fmt.Sprintf("Failed to create alias for %s: %v", tool, copyErr))
				continue
			}
		} else {
			// On Unix, create symlink
			_ = os.Remove(aliasPath)
			if symlinkErr := os.Symlink(jfPath, aliasPath); symlinkErr != nil {
				log.Warn(fmt.Sprintf("Failed to create alias for %s: %v", tool, symlinkErr))
				continue
			}
		}
		createdCount++
		log.Debug(fmt.Sprintf("Created alias: %s -> %s", aliasPath, jfPath))
	}

	jfHash, err := computeFileSHA256(jfPath)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed computing jf binary hash: %v", err))
	}

	// 4. Load and update config under lock
	if err = withConfigLock(aliasDir, func() error {
		config, loadErr := loadConfig(aliasDir)
		if loadErr != nil {
			return loadErr
		}

		for _, tool := range selectedTools {
			if _, exists := config.ToolModes[tool]; !exists {
				config.ToolModes[tool] = ModeJF
			}
		}

		config.EnabledTools = append([]string(nil), selectedTools...)
		config.JfBinarySHA256 = jfHash
		config.Enabled = true
		return writeConfig(aliasDir, config)
	}); err != nil {
		return errorutils.CheckError(err)
	}

	// Success message
	log.Info(fmt.Sprintf("Created %d aliases in %s", createdCount, binDir))
	log.Info(fmt.Sprintf("Configured packages: %s", strings.Join(selectedTools, ", ")))
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
