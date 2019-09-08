package commands

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-cli-go/utils/ioutils"
	"github.com/jfrog/jfrog-cli-go/utils/lock"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"reflect"
	"sync"
	"syscall"
)

// Internal golang locking for the same process.
var mutex sync.Mutex

type ConfigCommand struct {
	details        *config.ArtifactoryDetails
	defaultDetails *config.ArtifactoryDetails
	interactive    bool
	encPassword    bool
	serverId       string
}

func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{}
}

func (cc *ConfigCommand) SetServerId(serverId string) *ConfigCommand {
	cc.serverId = serverId
	return cc
}

func (cc *ConfigCommand) SetEncPassword(encPassword bool) *ConfigCommand {
	cc.encPassword = encPassword
	return cc
}

func (cc *ConfigCommand) SetInteractive(interactive bool) *ConfigCommand {
	cc.interactive = interactive
	return cc
}

func (cc *ConfigCommand) SetDefaultDetails(defaultDetails *config.ArtifactoryDetails) *ConfigCommand {
	cc.defaultDetails = defaultDetails
	return cc
}

func (cc *ConfigCommand) SetDetails(details *config.ArtifactoryDetails) *ConfigCommand {
	cc.details = details
	return cc
}

func (cc *ConfigCommand) Run() error {
	return cc.Config()
}

func (cc *ConfigCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	// If cc.details is not empty, then return it.
	if cc.details != nil && !reflect.DeepEqual(config.ArtifactoryDetails{}, *cc.details) {
		return cc.details, nil
	}
	// If cc.defaultDetails is not empty, then return it.
	if cc.defaultDetails != nil && !reflect.DeepEqual(config.ArtifactoryDetails{}, *cc.defaultDetails) {
		return cc.defaultDetails, nil
	}
	return nil, nil
}

func (cc *ConfigCommand) CommandName() string {
	return "rt_config"
}

func (cc *ConfigCommand) Config() error {
	mutex.Lock()
	lockFile, err := lock.CreateLock()
	defer mutex.Unlock()
	defer lockFile.Unlock()

	if err != nil {
		return err
	}

	configurations, err := cc.prepareConfigurationData()
	if err != nil {
		return err
	}
	if cc.interactive {
		err = cc.getConfigurationFromUser()
		if err != nil {
			return err
		}
	}

	if len(configurations) == 1 {
		cc.details.IsDefault = true
	}

	err = checkSingleAuthMethod(cc.details)
	if err != nil {
		return err
	}

	if cc.encPassword {
		err = cc.encryptPassword()
		if err != nil {
			return err
		}
	}
	err = config.SaveArtifactoryConf(configurations)
	return err
}

func (cc *ConfigCommand) prepareConfigurationData() ([]*config.ArtifactoryDetails, error) {
	// If details is nil, initialize a new one
	if cc.details == nil {
		cc.details = new(config.ArtifactoryDetails)
		if cc.defaultDetails != nil {
			cc.details.InsecureTls = cc.defaultDetails.InsecureTls
		}
	}

	// Get configurations list
	configurations, err := config.GetAllArtifactoryConfigs()
	if err != nil {
		return configurations, err
	}

	// Get default server details
	if cc.defaultDetails == nil {
		cc.defaultDetails, err = config.GetDefaultConfiguredArtifactoryConf(configurations)
		if err != nil {
			return configurations, errorutils.CheckError(err)
		}
	}

	// Get server id
	if cc.interactive && cc.serverId == "" {
		ioutils.ScanFromConsole("Artifactory server ID", &cc.serverId, cc.defaultDetails.ServerId)
	}
	cc.details.ServerId = cc.resolveServerId()

	// Remove and get the server details from the configurations list
	tempConfiguration, configurations := config.GetAndRemoveConfiguration(cc.details.ServerId, configurations)

	// Change default server details if the server was exist in the configurations list
	if tempConfiguration != nil {
		cc.defaultDetails = tempConfiguration
		cc.details.IsDefault = tempConfiguration.IsDefault
	}

	// Append the configuration to the configurations list
	configurations = append(configurations, cc.details)
	return configurations, err
}

/// Returning the first non empty value:
// 1. The serverId argument sent.
// 2. details.ServerId
// 3. defaultDetails.ServerId
// 4. config.DEFAULT_SERVER_ID
func (cc *ConfigCommand) resolveServerId() string {
	if cc.serverId != "" {
		return cc.serverId
	}
	if cc.details.ServerId != "" {
		return cc.details.ServerId
	}
	if cc.defaultDetails.ServerId != "" {
		return cc.defaultDetails.ServerId
	}
	return config.DefaultServerId
}

func (cc *ConfigCommand) getConfigurationFromUser() error {
	allowUsingSavedPassword := true
	if cc.details.Url == "" {
		ioutils.ScanFromConsole("Artifactory URL", &cc.details.Url, cc.defaultDetails.Url)
		allowUsingSavedPassword = false
	}
	// Ssh-Key
	if fileutils.IsSshUrl(cc.details.Url) {
		return getSshKeyPath(cc.details)
	}
	cc.details.Url = clientutils.AddTrailingSlashIfNeeded(cc.details.Url)
	// Api-Key/Password/Access-Token
	if cc.details.ApiKey == "" && cc.details.Password == "" && cc.details.AccessToken == "" {
		err := readAccessTokenFromConsole(cc.details)
		if err != nil {
			return err
		}
		if len(cc.details.GetAccessToken()) == 0 {
			return ioutils.ReadCredentialsFromConsole(cc.details, cc.defaultDetails, allowUsingSavedPassword)
		}
	}
	return nil
}

func readAccessTokenFromConsole(details *config.ArtifactoryDetails) error {
	print("Access token (Leave blank for username and password/API key): ")
	byteToken, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return errorutils.CheckError(err)
	}
	// New-line required after the access token input:
	fmt.Println()
	if len(byteToken) > 0 {
		details.SetAccessToken(string(byteToken))
		_, err := new(generic.PingCommand).SetRtDetails(details).Ping()
		return err
	}
	return nil
}

func getSshKeyPath(details *config.ArtifactoryDetails) error {
	// If path not provided as a key, read from console:
	if details.SshKeyPath == "" {
		ioutils.ScanFromConsole("SSH key file path (optional)", &details.SshKeyPath, "")
	}

	// If path still not provided, return and warn about relying on agent.
	if details.SshKeyPath == "" {
		log.Info("SSH Key path not provided. You can also specify a key path using the --ssh-key-path command option. If no key will be specified, you will rely on ssh-agent only.")
		return nil
	}

	// If SSH key path provided, check if exists:
	details.SshKeyPath = clientutils.ReplaceTildeWithUserHome(details.SshKeyPath)
	exists, err := fileutils.IsFileExists(details.SshKeyPath, false)
	if err != nil {
		return err
	}

	if exists {
		sshKeyBytes, err := ioutil.ReadFile(details.SshKeyPath)
		if err != nil {
			return nil
		}
		encryptedKey, err := auth.IsEncrypted(sshKeyBytes)
		// If exists and not encrypted (or error occurred), return without asking for passphrase
		if err != nil || !encryptedKey {
			return err
		}
		log.Info("The key file at the specified path is encrypted, you may pass the passphrase as an option with every command (but config).")

	} else {
		log.Info("Could not find key in provided path. You may place the key file there later. If you choose to use an encrypted key, you may pass the passphrase as an option with every command.")
	}

	return err
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

func Import(serverToken string) error {
	artifactoryDetails, err := config.Import(serverToken)
	if err != nil {
		return err
	}
	log.Output("Importing server ID", "'"+artifactoryDetails.ServerId+"'")
	configCommand := &ConfigCommand{
		details:  artifactoryDetails,
		serverId: artifactoryDetails.ServerId,
	}
	return configCommand.Config()
}

func Export(serverName string) error {
	artifactoryDetails, err := config.GetArtifactorySpecificConfig(serverName)
	if err != nil {
		return err
	}
	serverToken, err := config.Export(artifactoryDetails)
	if err != nil {
		return err
	}
	log.Output(serverToken)
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
		if details.AccessToken != "" {
			log.Output("Access token: ***")
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

func (cc *ConfigCommand) encryptPassword() error {
	if cc.details.Password == "" {
		return nil
	}

	// New-line required after the password input:
	fmt.Println()

	log.Info("Encrypting password...")

	artAuth, err := cc.details.CreateArtAuthConfig()
	if err != nil {
		return err
	}
	encPassword, err := utils.GetEncryptedPasswordFromArtifactory(artAuth, cc.details.InsecureTls)
	if err != nil {
		return err
	}
	cc.details.Password = encPassword
	return err
}

func checkSingleAuthMethod(details *config.ArtifactoryDetails) error {
	boolArr := []bool{details.User != "" && details.Password != "", details.ApiKey != "", fileutils.IsSshUrl(details.Url), details.AccessToken != ""}
	if cliutils.SumTrueValues(boolArr) > 1 {
		return errorutils.CheckError(errors.New("Only one authentication method is allowed: Username + Password/API key, RSA Token (SSH) or Access Token."))
	}
	return nil
}

type ConfigCommandConfiguration struct {
	ArtDetails  *config.ArtifactoryDetails
	Interactive bool
	EncPassword bool
}

func GetAllArtifactoryServerIds() []string {
	var serverIds []string
	configuration, err := config.GetAllArtifactoryConfigs()
	if err != nil {
		return serverIds
	}
	for _, serverConfig := range configuration {
		serverIds = append(serverIds, serverConfig.ServerId)
	}
	return serverIds
}
