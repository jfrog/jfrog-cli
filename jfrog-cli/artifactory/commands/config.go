package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/ioutils"
	"strings"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
)

func Config(details *config.ArtifactoryDetails, defaultDetails *config.ArtifactoryDetails, interactive,
shouldEncPassword bool, serverId string) (*config.ArtifactoryDetails, error) {

	details, defaultDetails, configurations, err := prepareConfigurationData(serverId, details, defaultDetails)
	if err != nil {
		return nil, err
	}
	if interactive {
		err = getConfigurationFromUser(details, defaultDetails)
		if err != nil {
			return nil, err
		}
	}
	serverId = resolveServerId(serverId, details, defaultDetails)
	err = checkSingleAuthMethod(details)
	if err != nil {
		return nil, err
	}

	details.Url = cliutils.AddTrailingSlashIfNeeded(details.Url)
	if shouldEncPassword {
		details, err = EncryptPassword(details)
		if err != nil {
			return nil, err
		}
	}
	populateConfigDetails(serverId, details, defaultDetails)
	err = config.SaveArtifactoryConf(configurations)
	return details, err
}

func populateConfigDetails(serverId string, details, defaultDetails *config.ArtifactoryDetails) {
	if defaultDetails == nil {
		defaultDetails = new(config.ArtifactoryDetails)
	}
	isDefault := defaultDetails.IsDefault
	*defaultDetails = *details
	defaultDetails.IsDefault = isDefault
	defaultDetails.ServerId = serverId
}

func prepareConfigurationData(serverId string, details, defaultDetails *config.ArtifactoryDetails) (*config.ArtifactoryDetails, *config.ArtifactoryDetails, []*config.ArtifactoryDetails, error) {
	configurations, err := config.GetAllArtifactoryConfigs()
	if err != nil {
		return details, defaultDetails, configurations, err
	}
	if details == nil {
		details = new(config.ArtifactoryDetails)
	}
	if defaultDetails == nil {
		defaultDetails, configurations, details, err = handleEmptyDefaultDetails(serverId, configurations, details)
	} else {
		// We got here from offer config flow
		configurations = append(configurations, defaultDetails)
		defaultDetails.IsDefault = true
	}

	return details, defaultDetails, configurations, err
}

/// Returning the first non empty value:
// 1. The serverId argument sent.
// 2. details.ServerId
// 3. defaultDetails.ServerId
// 4. config.DEFAULT_SERVER_ID
func resolveServerId(serverId string, details *config.ArtifactoryDetails, defaultDetails *config.ArtifactoryDetails) string {
	if serverId != "" {
		return serverId
	}
	if details.ServerId != "" {
		return details.ServerId
	}
	if defaultDetails.ServerId != "" {
		return defaultDetails.ServerId
	}
	return config.DEFAULT_SERVER_ID
}

func handleEmptyDefaultDetails(serverId string, configurations []*config.ArtifactoryDetails, details *config.ArtifactoryDetails) (*config.ArtifactoryDetails, []*config.ArtifactoryDetails, *config.ArtifactoryDetails, error) {
	var defaultDetails *config.ArtifactoryDetails
	var err error
	// If we don't have serverId we need to search for the default server
	if serverId == "" {
		defaultDetails, err = config.GetDefaultArtifactoryConf(configurations)
		// No default was found
		if err == nil && defaultDetails.IsEmpty() {
			configurations = append(configurations, defaultDetails)
		}
	} else {
		defaultDetails = config.GetArtifactoryConfByServerId(serverId, configurations)
		// No server with serverId was found
		if defaultDetails.IsEmpty() {
			defaultDetails.IsDefault = len(configurations) == 0
			configurations = append(configurations, defaultDetails)
		}
	}
	return defaultDetails, configurations, details, err
}

func getConfigurationFromUser(details, defaultDetails *config.ArtifactoryDetails) error {
	if details.Url == "" {
		ioutils.ScanFromConsole("Artifactory URL", &details.Url, defaultDetails.Url)
	}
	if details.ServerId == "" {
		ioutils.ScanFromConsole("Artifactory server ID", &details.ServerId, defaultDetails.ServerId)
	}
	if strings.Index(details.Url, "ssh://") == 0 || strings.Index(details.Url, "SSH://") == 0 {
		err := readSshKeyPathFromConsole(details, defaultDetails)
		if err != nil {
			return err
		}
	} else {
		if details.ApiKey == "" && details.Password == "" {
			ioutils.ScanFromConsole("API key (leave empty for basic authentication)", &details.ApiKey, "")
		}
		if details.ApiKey == "" {
			ioutils.ReadCredentialsFromConsole(details, defaultDetails)
		}
	}
	return nil
}

func readSshKeyPathFromConsole(details, savedDetails *config.ArtifactoryDetails) error {
	if details.SshKeyPath == "" {
		ioutils.ScanFromConsole("SSH key file path", &details.SshKeyPath, savedDetails.SshKeyPath)
	}

	details.SshKeyPath = clientutils.ReplaceTildeWithUserHome(details.SshKeyPath)
	exists, err := fileutils.IsFileExists(details.SshKeyPath)
	if err != nil {
	    return err
	}
	if !exists {
		cliutils.CliLogger.Warn("Could not find SSH key file at:", details.SshKeyPath)
	}
	return nil
}

func ShowConfig(serverName string) error {
	var configuration []*config.ArtifactoryDetails
	if serverName != "" {
		singleConfig, err := config.GetArtifactorySpecificConfig(serverName)
		if err != nil {
			return err
		}
		configuration = []*config.ArtifactoryDetails{singleConfig}
	} else {
		var err error
		configuration, err = config.GetAllArtifactoryConfigs()
		if err != nil {
			return err
		}
	}
	printConfigs(configuration)
	return nil
}

func printConfigs(configuration []*config.ArtifactoryDetails) {
	for _, details := range configuration {
		if details.ServerId != "" {
			fmt.Println("Server ID: " + details.ServerId)
		}
		if details.Url != "" {
			fmt.Println("Url: " + details.Url)
		}
		if details.ApiKey != "" {
			fmt.Println("API key: " + details.ApiKey)
		}
		if details.User != "" {
			fmt.Println("User: " + details.User)
		}
		if details.Password != "" {
			fmt.Println("Password: ***")
		}
		if details.SshKeyPath != "" {
			fmt.Println("SSH key file path: " + details.SshKeyPath)
		}
		fmt.Println("Default: ", details.IsDefault)
		fmt.Println()
	}
}

func DeleteConfig(serverName string) error {
	configurations, err := config.GetAllArtifactoryConfigs()
	if err != nil {
		return err
	}
	var isDefault, isFoundName bool
	for i, config := range configurations {
		if config.ServerId == serverName {
			isDefault = config.IsDefault
			configurations = append(configurations[:i], configurations[i+1:]...)
			isFoundName = true
			break
		}

	}
	if isDefault && len(configurations) > 0 {
		configurations[0].IsDefault = true
	}
	if isFoundName {
		return config.SaveArtifactoryConf(configurations)
	}
	cliutils.CliLogger.Info("\"" + serverName + "\" configuration could not be found.\n")
	return nil
}

// Set the default configuration
func Use(serverName string) error {
	configurations, err := config.GetAllArtifactoryConfigs()
	if err != nil {
		return err
	}
	var isFoundName bool
	for _, config := range configurations {
		if config.ServerId == serverName {
			// In case the serverId is already default we can return, no more changes needed
			if config.IsDefault {
				return nil
			}
			config.IsDefault = true
			isFoundName = true
		} else {
			config.IsDefault = false
		}
	}
	// Need to save only if we found a server with the serverId
	if isFoundName {
		return config.SaveArtifactoryConf(configurations)
	} else {
		cliutils.CliLogger.Info("Couldn't find matching server, no changes were made.")
	}
	return nil
}

func ClearConfig(interactive bool) {
	if interactive {
		confirmed := cliutils.InteractiveConfirm("Are you sure you want to delete all the configurations?")
		if !confirmed {
			return
		}
	}
	config.SaveArtifactoryConf(make([]*config.ArtifactoryDetails, 0))
}

func GetConfig(serverId string) (*config.ArtifactoryDetails, error) {
	return config.GetArtifactorySpecificConfig(serverId)
}

func EncryptPassword(details *config.ArtifactoryDetails) (*config.ArtifactoryDetails, error) {
	if details.Password == "" {
		return details, nil
	}
	cliutils.CliLogger.Info("Encrypting password...")
	artAuth := details.CreateArtAuthConfig()
	encPassword, err := utils.GetEncryptedPasswordFromArtifactory(artAuth)
	if err != nil {
		return nil, err
	}
	details.Password = encPassword
	return details, err
}

func checkSingleAuthMethod(details *config.ArtifactoryDetails) (err error) {
	boolArr := []bool{details.User != "" && details.Password != "", details.ApiKey != "", details.SshKeyPath != ""}
	if cliutils.SumTrueValues(boolArr) > 1 {
		err = errorutils.CheckError(errors.New("Only one authentication method is allowd: Username/Password, API key or RSA tokens."))
	}
	return
}

type ConfigFlags struct {
	ArtDetails  *config.ArtifactoryDetails
	Interactive bool
	EncPassword bool
}