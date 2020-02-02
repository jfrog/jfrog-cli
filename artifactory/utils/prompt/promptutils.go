package prompt

import (
	"errors"
	"os"

	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/prompt"
	"github.com/spf13/viper"
)

const BUILD_CONF_VERSION = 1
const USE_ARTIFACTORY = "useArtifactory"

type CommonConfig struct {
	Version    int    `yaml:"version,omitempty"`
	ConfigType string `yaml:"type,omitempty"`
}

type ServerConfig struct {
	ServerId string `yaml:"serverID,omitempty"`
	User     string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Url      string `yaml:"url,omitempty"`
}

func (server *ServerConfig) Set(config *viper.Viper) error {
	if config.GetBool(USE_ARTIFACTORY) {
		server.ServerId = config.GetString(utils.SERVER_ID)
	}
	return nil
}

func VerifyConfigFile(configFilePath string, interactive bool) error {
	exists, err := fileutils.IsFileExists(configFilePath, false)
	if err != nil {
		return err
	}
	if exists {
		if !interactive {
			return nil
		}
		override, err := AskYesNo("Configuration file already exists at "+configFilePath+". Override it (y/n) [${default}]? ", "n", "override")
		if err != nil {
			return err
		}
		if !override {
			return errorutils.CheckError(errors.New("Operation canceled."))
		}
		return nil
	}

	// Create config file to make sure the path is valid
	f, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if errorutils.CheckError(err) != nil {
		return err
	}
	f.Close()
	// The file will be written at the end of successful configuration command.
	return errorutils.CheckError(os.Remove(configFilePath))
}

// Get Artifactory serverId from the user. If useArtifactoryQuestion is not empty, ask first whether to use artifactory.
func ReadArtifactoryServer(useArtifactoryQuestion string) (string, error) {
	// Get all Artifactory servers
	serversIds, defaultServer, err := getServersIdAndDefault()
	if err != nil {
		return "", err
	}
	if len(serversIds) == 0 {
		return "", errorutils.CheckError(errors.New("Artifactory server configuration is missing, use 'jfrog rt c' command to set server details."))
	}

	// Ask whether to use artifactory
	if useArtifactoryQuestion != "" {
		useArtifactory, err := AskYesNo(useArtifactoryQuestion, "y", USE_ARTIFACTORY)
		if err != nil || !useArtifactory {
			return "", err
		}
	}

	return AskAutocomplete("Set Artifactory server ID (press Tab for options) [${default}]: ", "Server does not exist. Please set a valid server ID.", serversIds, defaultServer, utils.SERVER_ID)
}

func ReadRepo(msg string, serverId string, repoTypes ...utils.RepoType) (string, error) {
	availableRepos, err := GetRepositories(serverId, repoTypes...)
	if err != nil {
		// If there are no available repos pass empty array.
		availableRepos = []string{}
	}
	repo := &prompt.Autocomplete{
		Msg:     msg,
		Options: availableRepos,
		Label:   utils.ProjectConfigRepo,
	}
	if len(availableRepos) > 0 {
		repo.ConfirmationMsg = "No such repository, continue anyway (y/n) [${default}]? "
		repo.ConfirmationDefault = "n"
	} else {
		repo.ErrMsg = "Repository name cannot be empty."
	}
	err = repo.Read()
	if err != nil {
		return "", err
	}
	return repo.GetResults().GetString(utils.ProjectConfigRepo), nil
}

func getServersIdAndDefault() ([]string, string, error) {
	allConfigs, err := config.GetAllArtifactoryConfigs()
	if err != nil {
		return nil, "", err
	}
	var defaultVal string
	var serversId []string
	for _, v := range allConfigs {
		if v.IsDefault {
			defaultVal = v.ServerId
		}
		serversId = append(serversId, v.ServerId)
	}
	return serversId, defaultVal, nil
}

func GetRepositories(serverId string, repoTypes ...utils.RepoType) ([]string, error) {
	artDetails, err := config.GetArtifactoryConf(serverId)
	if err != nil {
		return nil, err
	}

	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}

	return utils.GetRepositories(artAuth, repoTypes...)
}

func AskYesNo(message string, defaultStr string, label string) (bool, error) {
	question := &prompt.YesNo{
		Msg:     message,
		Default: defaultStr,
		Label:   label,
	}
	if err := question.Read(); err != nil {
		return false, errorutils.CheckError(err)
	}
	return question.Result.GetBool(label), nil
}

func AskString(message string, defaultStr string, label string) (string, error) {
	question := &prompt.Simple{
		Msg:     message,
		Default: defaultStr,
		Label:   label,
	}
	if err := question.Read(); err != nil {
		return "", errorutils.CheckError(err)
	}
	return question.Result.GetString(label), nil
}

func AskAutocomplete(msg string, errMsg string, options []string, defaultStr string, label string) (string, error) {
	question := &prompt.Autocomplete{
		Msg:     msg,
		ErrMsg:  errMsg,
		Options: options,
		Default: defaultStr,
		Label:   label,
	}
	if err := question.Read(); err != nil {
		return "", errorutils.CheckError(err)
	}
	return question.Result.GetString(label), nil
}