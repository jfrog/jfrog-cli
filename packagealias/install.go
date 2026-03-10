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

	jfPath, err := os.Executable()
	if err != nil {
		return errorutils.CheckError(fmt.Errorf("could not determine executable path: %w", err))
	}
	jfPath, err = filepath.EvalSymlinks(jfPath)
	if err != nil {
		return errorutils.CheckError(fmt.Errorf("could not resolve executable path: %w", err))
	}
	log.Debug(fmt.Sprintf("Using jf binary at: %s", jfPath))

	selectedTools, err := parsePackageList(ic.packagesArg)
	if err != nil {
		return err
	}

	jfHash, err := computeFileSHA256(jfPath)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed computing jf binary hash: %v", err))
	}

	var createdCount int

	// Hold the lock for the entire mutation: symlink/copy creation + config update.
	// This prevents two parallel installs from racing on the bin directory.
	if err = withConfigLock(aliasDir, func() error {
		selectedToolsSet := make(map[string]struct{}, len(selectedTools))
		for _, tool := range selectedTools {
			selectedToolsSet[tool] = struct{}{}
		}

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
				if copyErr := copyFile(jfPath, aliasPath); copyErr != nil {
					log.Warn(fmt.Sprintf("Failed to create alias for %s: %v", tool, copyErr))
					continue
				}
			} else {
				_ = os.Remove(aliasPath)
				if symlinkErr := os.Symlink(jfPath, aliasPath); symlinkErr != nil {
					log.Warn(fmt.Sprintf("Failed to create alias for %s: %v", tool, symlinkErr))
					continue
				}
			}
			createdCount++
			log.Debug(fmt.Sprintf("Created alias: %s -> %s", aliasPath, jfPath))
		}

		cfg, loadErr := loadConfig(aliasDir)
		if loadErr != nil {
			return loadErr
		}

		for _, tool := range selectedTools {
			if _, exists := cfg.ToolModes[tool]; !exists {
				cfg.ToolModes[tool] = ModeJF
			}
		}

		cfg.EnabledTools = append([]string(nil), selectedTools...)
		cfg.JfBinarySHA256 = jfHash
		cfg.Enabled = true
		return writeConfig(aliasDir, cfg)
	}); err != nil {
		return errorutils.CheckError(err)
	}

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
	defer func() {
		_ = srcFile.Close()
	}()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
