package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/ioutils"
	"strings"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
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
	err = checkSingleAuthMethod(details)
	if err != nil {
		return nil, err
	}

	details.Url = cliutils.AddTrailingSlashIfNeeded(details.Url)
	if shouldEncPassword {
		details, err = encryptPassword(details)
		if err != nil {
			return nil, err
		}
	}
	copyDetails(details, defaultDetails)
	err = config.SaveArtifactoryConf(configurations)
	return details, err
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
		if serverId == "" {
			defaultDetails = config.GetDefaultArtifactoryConf(configurations)
		} else {
			defaultDetails = config.GetArtifactoryConfByServerId(serverId, configurations)
			if len(configurations) == 0 {
				defaultDetails.IsDefault = true
			}
			if defaultDetails.Url == "" {
				configurations = append(configurations, defaultDetails)
				details.ServerId = serverId
			}
		}
	}
	return details, defaultDetails, configurations, err
}

func getConfigurationFromUser(details, defaultDetails *config.ArtifactoryDetails) error {
	if details.Url == "" {
		ioutils.ScanFromConsole("Artifactory URL", &details.Url, defaultDetails.Url)
	}
	if details.ServerId == "" {
		ioutils.ScanFromConsole("Artifactory server name", &details.ServerId, defaultDetails.ServerId)
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

func copyDetails(src, dst *config.ArtifactoryDetails) {
	if dst == nil {
		dst = new(config.ArtifactoryDetails)
	}
	isDefault := dst.IsDefault
	*dst = *src
	dst.IsDefault = isDefault
}

func readSshKeyPathFromConsole(details, savedDetails *config.ArtifactoryDetails) error {
	if details.SshKeyPath == "" {
		ioutils.ScanFromConsole("SSH key file path", &details.SshKeyPath, savedDetails.SshKeyPath)
	}

	details.SshKeyPath = cliutils.ReplaceTildeWithUserHome(details.SshKeyPath)
	exists, err := fileutils.IsFileExists(details.SshKeyPath)
	if err != nil {
	    return err
	}
	if !exists {
		log.Warn("Could not find SSH key file at:", details.SshKeyPath)
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
		if details.ServerId != "" {
			fmt.Println("Server Name: " + details.ServerId)
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
		log.Info("Couldn't find matching server, no changes were made.")
	}
	return nil
}

func ClearConfig() {
	config.SaveArtifactoryConf(make([]*config.ArtifactoryDetails, 0))
}

func GetConfig(serverId string) (*config.ArtifactoryDetails, error) {
	return config.GetArtifactorySpecificConfig(serverId)
}

func encryptPassword(details *config.ArtifactoryDetails) (*config.ArtifactoryDetails, error) {
	if details.Password == "" {
		return details, nil
	}
	log.Info("Encrypting password...")
	response, encPassword, err := utils.GetEncryptedPasswordFromArtifactory(details)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
        case 409:
            message := "\nYour Artifactory server is not configured to encrypt passwords.\n" +
                "You may use \"art config --enc-password=false\""
            err = cliutils.CheckError(errors.New(message))
        case 200:
            details.Password = encPassword
            log.Info("Done encrypting password.")
        default:
            err = cliutils.CheckError(errors.New("\nArtifactory response: " + response.Status))
	}
	return details, err
}

func checkSingleAuthMethod(details *config.ArtifactoryDetails) (err error) {
	boolArr := []bool{details.User != "" && details.Password != "", details.ApiKey != "", details.SshKeyPath != ""}
	if (cliutils.SumTrueValues(boolArr) > 1) {
		err = cliutils.CheckError(errors.New("Only one authentication method is allowd: Username/Password, API key or RSA tokens."))
	}
	return
}

type ConfigFlags struct {
	ArtDetails  *config.ArtifactoryDetails
	Interactive bool
	EncPassword bool
}