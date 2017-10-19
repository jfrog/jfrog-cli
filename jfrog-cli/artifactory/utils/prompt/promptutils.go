package prompt

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/prompt"
	"github.com/spf13/viper"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"errors"
	"os"
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

func VerifyConfigFile(configFilePath string) error {
	exists, err := fileutils.IsFileExists(configFilePath)
	if err != nil {
		return err
	}
	if exists {
		yesNoPrompt := &prompt.YesNo{
			Msg:     "Configuration file already exists at " + configFilePath + ". Override it (y/n) [${default}]? ",
			Default: "n",
			Label:   "override",
		}
		err = yesNoPrompt.Read()
		if err != nil {
			return err
		}

		if !yesNoPrompt.Result.GetBool("override") {
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
	} else {
		repo.ErrMsg = "Repository name cannot be empty."
	}
	err := repo.Read()
	if err != nil {
		return "", err
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
