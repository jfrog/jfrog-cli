package prompt

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/prompt"
	"github.com/spf13/viper"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"errors"
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

func VerifyConfigOverride(configFilePath string) bool {
	exists, err := fileutils.IsFileExists(configFilePath)
	if err != nil {
		return false
	}
	if !exists {
		return true
	}
	yesNoPrompt := &prompt.YesNo{
		Msg:     "Configuration file already exists at " + configFilePath + ". Override it (y/n) [${default}]? ",
		Default: "n",
		Label:   "answer",
	}
	if yesNoPrompt.Read() != nil {
		return false
	}
	return yesNoPrompt.Result.GetBool("answer")
}

func ReadArtifactoryServer(msg string) (*viper.Viper, error) {
	serversId, defaultServer, err := getServersIdAndDefault()
	if err != nil {
		return nil, err
	}

	if len(serversId) == 0 {
		return nil, errorutils.CheckError(errors.New("Artifactory server configuration is missing, use 'jfrog rt c' command to set server details."))
	}

	server := &prompt.YesNo{
		Msg:     msg,
		Default: "y",
		Label:   USE_ARTIFACTORY,
		Yes: &prompt.Autocomplete{
			Msg:     "Set Artifactory server ID (press Tab for options) [${default}]: ",
			ErrMsg:  "Server does not exist. Please set a valid server ID.",
			Options: serversId,
			Default: defaultServer,
			Label:   utils.SERVER_ID,
		},
	}

	err = server.Read()
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return server.GetResults(), nil
}

func ReadRepo(msg string, repos []string) (string, error) {
	repo := &prompt.Autocomplete{
		Msg:     msg,
		Options: repos,
		Label:   utils.REPO,
	}
	if len(repos) > 0 {
		repo.ConfirmationMsg = "No such repository, continue anyway (y/n) [${default}]? "
		repo.ConfirmationDefault = "n"
	}
	err := repo.Read()
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return repo.GetResults().GetString(utils.REPO), nil
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

func GetRepositories(resolveRes *viper.Viper, repoTypes ...utils.RepoType) ([]string, error) {
	var artDetails *config.ArtifactoryDetails
	serversId, err := config.GetAllArtifactoryConfigs()
	if err != nil {
		return []string{}, err
	}
	artDetails = config.GetArtifactoryConfByServerId(resolveRes.GetString(utils.SERVER_ID), serversId)
	return utils.GetRepositories(artDetails.CreateArtAuthConfig(), repoTypes...)
}
