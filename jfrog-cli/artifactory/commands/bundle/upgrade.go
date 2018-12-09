package bundle

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services/bundle"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/mcuadros/go-version"
	"io"
	"os/exec"
)

func UpgradeBundle(artDetails *config.ArtifactoryDetails, bundleConfigDetails *config.BundleDetails, dryRun bool) error {
	// Lock scope and unlock when finished
	if err := guard(); err != nil {
		return err
	}

	// Create Service Manager:
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return err
	}

	// Fetch all versions of the bundle from the edge node
	versions, err := servicesManager.GetBundleVersions(bundleConfigDetails.Name)
	if err != nil {
		return err
	}
	log.Debug("Versions found", versions)

	// Calculate the latest matched version
	latestMatchedVersion := getLatestMatchedVersion(bundleConfigDetails.VersionConstraints, versions)
	if latestMatchedVersion == "" {
		log.Debug(fmt.Printf("Bundle name '%s' with constraints '%s' didn't match with current versions: %v", bundleConfigDetails.Name, bundleConfigDetails.VersionConstraints, versions))
		return nil
	}

	// Quit if the latest version installed
	if isLatestVersion(bundleConfigDetails.CurrentVersion, latestMatchedVersion) {
		log.Debug(fmt.Printf("Latest version '%s' already installed", latestMatchedVersion))
		return nil
	}
	log.Output(fmt.Printf("New bundle version detected: '%s'", latestMatchedVersion))

	// Quit if dry run
	if dryRun {
		log.Output("[Dry run] Running", bundleConfigDetails.ScriptPath)
		return nil
	}

	// Run installation script
	if err := installNewVersion(bundleConfigDetails); err != nil {
		return err
	}

	// Updates bundle config version in 'jfrog-bundle.conf'
	return updateBundleConfigVersion(bundleConfigDetails, latestMatchedVersion)
}

// Return latest version in 'COMPLETE' status that matched the bundle config constraints.
func getLatestMatchedVersion(versionConstraints string, versions []bundle.Version) string {
	var versionStrings []string
	constraintGroup := version.NewConstrainGroupFromString(versionConstraints)
	for _, ver := range versions {
		if ver.Status == bundle.Complete && constraintGroup.Match(ver.Version) {
			versionStrings = append(versionStrings, ver.Version)
		}
	}
	if len(versionStrings) == 0 {
		return ""
	}
	version.Sort(versionStrings)
	return versionStrings[len(versionStrings)-1]
}

// Return true if currentVersion >= latestMatchedVersion
func isLatestVersion(currentVersion, latestMatchedVersion string) bool {
	if currentVersion == "" {
		return false
	}
	return version.Compare(currentVersion, latestMatchedVersion, ">=")
}

func installNewVersion(bundleConfigDetails *config.BundleDetails) error {
	log.Output("Running", bundleConfigDetails.ScriptPath)
	commandConfig := &bundleInstallationScript{bundleConfigDetails.ScriptPath}
	if err := utils.RunCmd(commandConfig); err != nil {
		return err
	}
	log.Output("Upgrade succeeded.")
	return nil
}

func updateBundleConfigVersion(bundleConfigDetails *config.BundleDetails, latestMatchedVersion string) error {
	bundleConfigDetails.CurrentVersion = latestMatchedVersion
	return config.EditBundleConfig(bundleConfigDetails)
}

func (config *bundleInstallationScript) GetCmd() *exec.Cmd {
	return exec.Command(config.scriptPath)
}

func (config *bundleInstallationScript) GetEnv() map[string]string {
	return map[string]string{}
}

func (config *bundleInstallationScript) GetStdWriter() io.WriteCloser {
	return nil
}

func (config *bundleInstallationScript) GetErrWriter() io.WriteCloser {
	return nil
}

type bundleInstallationScript struct {
	scriptPath string
}
