package mvn

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/prompt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func CreateBuildConfig(configFilePath string) error {
	if err := prompt.VerifyConfigFile(configFilePath); err != nil {
		return err
	}

	configResult := &MavenBuildConfig{}
	configResult.Version = prompt.BUILD_CONF_VERSION
	configResult.ConfigType = utils.MAVEN.String()

	vConfig, err := prompt.ReadArtifactoryServer("Resolve dependencies from Artifactory (y/n) [${default}]? ")
	if err != nil {
		return err
	}
	if vConfig.GetBool(prompt.USE_ARTIFACTORY) {
		err = configResult.Resolver.Server.Set(vConfig)
		if err != nil {
			return err
		}
		availableRepos, err := prompt.GetRepositories(vConfig, utils.REMOTE, utils.VIRTUAL)
		if err != nil {
			// If there are no available repos pass empty array.
			availableRepos = []string{}
		}
		configResult.Resolver.ReleaseRepo, err = prompt.ReadRepo("Set resolution repository for release dependencies (press Tab for options): ", availableRepos)
		if err != nil {
			return err
		}
		configResult.Resolver.SnapshotRepo, err = prompt.ReadRepo("Set resolution repository for snapshot dependencies (press Tab for options): ", availableRepos)
		if err != nil {
			return err
		}
	}

	vConfig, err = prompt.ReadArtifactoryServer("Deploy artifacts to Artifactory (y/n) [${default}]? ")
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
		configResult.Deployer.ReleaseRepo, err = prompt.ReadRepo("Set repository for release artifacts deployment (press Tab for options): ", availableRepos)
		if err != nil {
			return err
		}
		configResult.Deployer.SnapshotRepo, err = prompt.ReadRepo("Set repository for snapshot artifacts deployment (press Tab for options): ", availableRepos)
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
		return err
	}

	log.Info("Maven build config successfully created.")
	return nil
}

type MavenBuildConfig struct {
	prompt.CommonConfig `yaml:"common,inline"`
	Resolver            MavenRepos `yaml:"resolver,omitempty"`
	Deployer            MavenRepos `yaml:"deployer,omitempty"`
}

type MavenRepos struct {
	SnapshotRepo string              `yaml:"snapshotRepo,omitempty"`
	ReleaseRepo  string              `yaml:"releaseRepo,omitempty"`
	Server       prompt.ServerConfig `yaml:"server,inline"`
}
