package bundle

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/ioutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/lock"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/pkg/errors"
	"sync"
)

// Internal golang locking for the same process.
var mutex sync.Mutex

func Config(details *config.BundleDetails, interactive bool, bundleConfigId string) error {
	// Lock scope
	if err := guard(); err != nil {
		return err
	}

	if details == nil {
		details = new(config.BundleDetails)
	}
	details, defaultDetails, configurations, err := prepareConfigurationData(bundleConfigId, details, interactive)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if interactive {
		getConfigurationFromUser(details, defaultDetails)
	}

	if len(configurations) == 1 {
		details.IsDefault = true
	}

	return config.SaveBundleConfigs(configurations)
}

func guard() error {
	mutex.Lock()
	bundleLockFile, err := lock.CreateLock()
	defer unlockBundles(bundleLockFile)
	return err
}

func unlockBundles(lockFile lock.Lock) {
	mutex.Unlock()
	if err := lockFile.Unlock(); err != nil {
		log.Error("An error occurred while trying to unlock bundle config file mutex", err)
	}
}

func prepareConfigurationData(bundleConfigId string, details *config.BundleDetails, interactive bool) (*config.BundleDetails, *config.BundleDetails, []*config.BundleDetails, error) {
	// Get configurations list
	configurations, err := config.GetAllBundleConfigs()
	if err != nil {
		return details, nil, configurations, err
	}

	// Get default bundle config details
	defaultDetails, err := config.GetDefaultBundleConf(configurations)
	if err != nil {
		return details, defaultDetails, configurations, err
	}

	// Get bundle config id
	if interactive && bundleConfigId == "" {
		ioutils.ScanFromConsole("Bundle config ID", &bundleConfigId, defaultDetails.BundleConfigId)
	}
	details.BundleConfigId = resolveBundleConfigId(bundleConfigId, details, defaultDetails)

	// Remove and get the bundle config details from the configurations list
	tempConfiguration, configurations := config.GetAndRemoveBundleConfiguration(details.BundleConfigId, configurations)

	// Change default server details if the server was exist in the configurations list
	if tempConfiguration != nil {
		defaultDetails = tempConfiguration
		details.IsDefault = tempConfiguration.IsDefault
	}

	// Append the configuration to the configurations list
	configurations = append(configurations, details)
	return details, defaultDetails, configurations, err
}

/// Returning the first non empty value:
// 1. The bundleConfigId argument sent.
// 2. details.bundleConfigId
// 3. defaultDetails.bundleConfigId
// 4. config.DefaultBundleConfigId
func resolveBundleConfigId(bundleConfigId string, details *config.BundleDetails, defaultDetails *config.BundleDetails) string {
	if bundleConfigId != "" {
		return bundleConfigId
	}
	if details.BundleConfigId != "" {
		return details.BundleConfigId
	}
	if defaultDetails.BundleConfigId != "" {
		return defaultDetails.BundleConfigId
	}
	return config.DefaultBundleConfigId
}

func getConfigurationFromUser(details, defaultDetails *config.BundleDetails) {
	for details.ServerId == "" {
		ioutils.ScanFromConsole("Server ID", &details.ServerId, defaultDetails.ServerId)
	}
	for details.Name == "" {
		ioutils.ScanFromConsole("Bundle name", &details.Name, defaultDetails.Name)
	}
	if details.CurrentVersion == "" {
		ioutils.ScanFromConsole("Current bundle version (optional)", &details.CurrentVersion, defaultDetails.CurrentVersion)
	}
	for details.VersionConstraints == "" {
		ioutils.ScanFromConsole("Bundle version constraints", &details.VersionConstraints, defaultDetails.VersionConstraints)
	}
	for details.ScriptPath == "" {
		ioutils.ScanFromConsole("Installation script path", &details.ScriptPath, defaultDetails.ScriptPath)
	}
}

func ShowConfig(bundleConfigId string) error {
	var configuration []*config.BundleDetails
	if bundleConfigId != "" {
		singleConfig, err := config.GetBundleConf(bundleConfigId)
		if err != nil {
			return err
		}
		configuration = []*config.BundleDetails{singleConfig}
	} else {
		var err error
		if configuration, err = config.GetAllBundleConfigs(); err != nil {
			return err
		}
	}
	printConfigs(configuration)
	return nil
}

func printConfigs(configuration []*config.BundleDetails) {
	for _, details := range configuration {
		if details.BundleConfigId != "" {
			log.Output("Bundle config ID:", details.BundleConfigId)
		}
		if details.ServerId != "" {
			log.Output("Server ID:", details.ServerId)
		}
		if details.Name != "" {
			log.Output("Name:", details.Name)
		}
		if details.CurrentVersion != "" {
			log.Output("Current bundle version:", details.CurrentVersion)
		}
		if details.VersionConstraints != "" {
			log.Output("Bundle version constraints:", details.VersionConstraints)
		}
		if details.ScriptPath != "" {
			log.Output("Installation script path:", details.ScriptPath)
		}
		log.Output("Default:", details.IsDefault)
		log.Output()
	}
}

func DeleteConfig(bundleConfigId string) error {
	// Lock scope
	if err := guard(); err != nil {
		return err
	}
	bundleConfigs, err := config.GetAllBundleConfigs()
	if err != nil {
		return err
	}
	var isDefault, isFoundName bool
	for i, bundleConfig := range bundleConfigs {
		if bundleConfig.BundleConfigId == bundleConfigId {
			isDefault = bundleConfig.IsDefault
			bundleConfigs = append(bundleConfigs[:i], bundleConfigs[i+1:]...)
			isFoundName = true
			break
		}

	}
	if isDefault && len(bundleConfigs) > 0 {
		bundleConfigs[0].IsDefault = true
	}
	if isFoundName {
		return config.SaveBundleConfigs(bundleConfigs)
	}
	log.Info("\"" + bundleConfigId + "\" configuration could not be found.\n")
	return nil
}

// Set the default configuration
func Use(bundleConfigId string) error {
	// Lock scope
	if err := guard(); err != nil {
		return err
	}
	configurations, err := config.GetAllBundleConfigs()
	if err != nil {
		return err
	}
	var bundleConfigFound *config.BundleDetails
	newDefaultBundleConfig := true
	for _, bundleConfig := range configurations {
		if bundleConfig.BundleConfigId == bundleConfigId {
			bundleConfigFound = bundleConfig
			if bundleConfig.IsDefault {
				newDefaultBundleConfig = false
				break
			}
			bundleConfig.IsDefault = true
		} else {
			bundleConfig.IsDefault = false
		}
	}
	// Need to save only if we found a bundle configuration with the bundleConfigId
	if bundleConfigFound != nil {
		if newDefaultBundleConfig {
			err = config.SaveBundleConfigs(configurations)
			if err != nil {
				return err
			}
		}
		log.Info(fmt.Sprintf("Using bundle config ID '%s' (%s/%s).", bundleConfigFound.BundleConfigId, bundleConfigFound.Name, bundleConfigFound.VersionConstraints))
		return nil
	}
	return errorutils.CheckError(errors.New(fmt.Sprintf("Could not find a bundle config with ID '%s'.", bundleConfigId)))
}

func ClearConfig(interactive bool) error {
	// Lock scope
	if err := guard(); err != nil {
		return err
	}
	if interactive {
		confirmed := cliutils.InteractiveConfirm("Are you sure you want to delete all the configurations?")
		if !confirmed {
			return nil
		}
	}
	return config.SaveBundleConfigs(make([]*config.BundleDetails, 0))
}

func GetConfig(serverId string) (*config.ArtifactoryDetails, error) {
	return config.GetArtifactorySpecificConfig(serverId)
}

type ConfigCommandConfiguration struct {
	BundleDetails *config.BundleDetails
	Interactive   bool
}
