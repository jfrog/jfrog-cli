package golang

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/prompt"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

func CreateBuildConfig(global bool) error {
	projectDir, err := utils.GetProjectDir(global)
	if err != nil {
		return err
	}
	err = fileutils.CreateDirIfNotExist(projectDir)
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(projectDir, "go.yaml")
	if err := prompt.VerifyConfigFile(configFilePath); err != nil {
		return err
	}

	configResult := &GoBuildConfig{}
	configResult.Version = prompt.BUILD_CONF_VERSION
	configResult.ConfigType = utils.GO.String()
	vConfig, err := prompt.ReadServerId()
	if err != nil {
		return err
	}
	err = configResult.Resolver.Server.Set(vConfig)
	if err != nil {
		return err
	}
	availableRepos, err := prompt.GetRepositories(vConfig, utils.REMOTE, utils.VIRTUAL)
	if err != nil {
		// If there are no available repos pass empty array.
		availableRepos = []string{}
	}
	configResult.Resolver.Repo, err = prompt.ReadRepo("Set repository for dependencies resolution (press Tab for options): ", availableRepos)
	if err != nil {
		return err
	}

	vConfig, err = prompt.ReadArtifactoryServer("Deploy project dependencies to Artifactory (y/n) [${default}]? ")
	if err != nil {
		return err
	}
	if vConfig.GetBool(prompt.USE_ARTIFACTORY) {
		err = configResult.Deployer.Server.Set(vConfig)
		if err != nil {
			return err
		}
		availableRepos, err := prompt.GetRepositories(vConfig, utils.LOCAL, utils.VIRTUAL)
		if err != nil {
			// If there are no available repos pass empty array.
			availableRepos = []string{}
		}
		configResult.Deployer.Repo, err = prompt.ReadRepo("Set repository for dependencies deployment (press Tab for options): ", availableRepos)
		if err != nil {
			return err
		}
	}
	resBytes, err := yaml.Marshal(&configResult)
	if err != nil {
		return errorutils.CheckError(err)
	}
	err = ioutil.WriteFile(configFilePath, resBytes, 0644)
	if err != nil {
		return errorutils.CheckError(err)
	}

	log.Info("Go build config successfully created.")
	return nil

}

type GoBuildConfig struct {
	prompt.CommonConfig `yaml:"common,inline"`
	Resolver            GoRepo `yaml:"resolver,omitempty"`
	Deployer            GoRepo `yaml:"deployer,omitempty"`
}

type GoRepo struct {
	Repo   string              `yaml:"repo,omitempty"`
	Server prompt.ServerConfig `yaml:"server,inline"`
}
