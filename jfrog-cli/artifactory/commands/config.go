package commands

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/ioutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/lock"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/prompt"
	"sync"
)

// Internal golang locking for the same process.
var mutux sync.Mutex

func Config(details *config.ArtifactoryDetails, defaultDetails *config.ArtifactoryDetails, interactive,
	shouldEncPassword bool, serverId string) (*config.ArtifactoryDetails, error) {
	mutux.Lock()
	lockFile, err := lock.CreateLock()
	defer mutux.Unlock()
	defer lockFile.Unlock()

	if err != nil {
		return nil, err
	}

	if details == nil {
		details = new(config.ArtifactoryDetails)
	}
	details, defaultDetails, configurations, err := prepareConfigurationData(serverId, details, defaultDetails, interactive)
	if err != nil {
		return nil, err
	}
	if interactive {
		err = getConfigurationFromUser(details, defaultDetails)
		if err != nil {
			return nil, err
		}
	}

	if len(configurations) == 1 {
		details.IsDefault = true
	}

	err = checkSingleAuthMethod(details)
	if err != nil {
		return nil, err
	}

	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	if shouldEncPassword {
		details, err = EncryptPassword(details)
		if err != nil {
			return nil, err
		}
	}
	err = config.SaveArtifactoryConf(configurations)
	return details, err
}

func prepareConfigurationData(serverId string, details, defaultDetails *config.ArtifactoryDetails, interactive bool) (*config.ArtifactoryDetails, *config.ArtifactoryDetails, []*config.ArtifactoryDetails, error) {
	// Get configurations list
	configurations, err := config.GetAllArtifactoryConfigs()
	if err != nil {
		return details, defaultDetails, configurations, err
	}

	// Get default server details
	if defaultDetails == nil {
		defaultDetails, err = config.GetDefaultArtifactoryConf(configurations)
		if err != nil {
			return details, defaultDetails, configurations, err
		}
	}

	// Get server id
	if interactive && serverId == "" {
		ioutils.ScanFromConsole("Artifactory server ID", &serverId, defaultDetails.ServerId)
	}
	details.ServerId = resolveServerId(serverId, details, defaultDetails)

	// Remove and get the server details from the configurations list
	tempConfiguration, configurations := config.GetAndRemoveConfiguration(details.ServerId, configurations)

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
	return config.DefaultServerId
}

func getConfigurationFromUser(details, defaultDetails *config.ArtifactoryDetails) error {
	allowUsingSavedPassword := true
	if details.Url == "" {
		ioutils.ScanFromConsole("Artifactory URL", &details.Url, defaultDetails.Url)
		allowUsingSavedPassword = false
	}
	if fileutils.IsSshUrl(details.Url) {
		useAgentPrompt := &prompt.YesNo{
			Msg:     "Would you like to use SSH agent (y/n) [${default}]? ",
			Label:   "useSshAgent",
			Default: "n",
		}
		if err := useAgentPrompt.Read(); err != nil {
			return err
		}
		if !useAgentPrompt.GetResults().GetBool("useSshAgent") {
			if err := readSshKeyPathFromConsole(details, defaultDetails); err != nil {
				return err
			}
		}
	} else {
		if details.ApiKey == "" && details.Password == "" {
			ioutils.ReadCredentialsFromConsole(details, defaultDetails, allowUsingSavedPassword)
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
		if details.ServerId != "" {
			log.Output("Server ID: " + details.ServerId)
		}
		if details.Url != "" {
			log.Output("Url: " + details.Url)
		}
		if details.ApiKey != "" {
			log.Output("API key: " + details.ApiKey)
		}
		if details.User != "" {
			log.Output("User: " + details.User)
		}
		if details.Password != "" {
			log.Output("Password: ***")
		}
		if details.SshKeyPath != "" {
			log.Output("SSH key file path: " + details.SshKeyPath)
		}
		log.Output("Default: ", details.IsDefault)
		log.Output()
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
	log.Info("\"" + serverName + "\" configuration could not be found.\n")
	return nil
}

// Set the default configuration
func Use(serverId string) error {
	configurations, err := config.GetAllArtifactoryConfigs()
	if err != nil {
		return err
	}
	var serverFound *config.ArtifactoryDetails
	newDefaultServer := true
	for _, config := range configurations {
		if config.ServerId == serverId {
			serverFound = config
			if config.IsDefault {
				newDefaultServer = false
				break
			}
			config.IsDefault = true
		} else {
			config.IsDefault = false
		}
	}
	// Need to save only if we found a server with the serverId
	if serverFound != nil {
		if newDefaultServer {
			err = config.SaveArtifactoryConf(configurations)
			if err != nil {
				return err
			}
		}
		log.Info(fmt.Sprintf("Using server ID '%s' (%s).", serverFound.ServerId, serverFound.Url))
		return nil
	}
	return errorutils.CheckError(errors.New(fmt.Sprintf("Could not find a server with ID '%s'.", serverId)))
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
	log.Info("Encrypting password...")
	artAuth, err := details.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	encPassword, err := utils.GetEncryptedPasswordFromArtifactory(artAuth)
	if err != nil {
		return nil, err
	}
	details.Password = encPassword
	return details, err
}

func checkSingleAuthMethod(details *config.ArtifactoryDetails) error {
	boolArr := []bool{details.User != "" && details.Password != "", details.ApiKey != "", fileutils.IsSshUrl(details.Url)}
	if cliutils.SumTrueValues(boolArr) > 1 {
		return errorutils.CheckError(errors.New("Only one authentication method is allowd: Username/Password, API key or RSA tokens."))
	}
	return nil
}

type ConfigCommandConfiguration struct {
	ArtDetails  *config.ArtifactoryDetails
	Interactive bool
	EncPassword bool
}
