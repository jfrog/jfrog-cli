package utils

import (
	"io/ioutil"
	"path/filepath"

	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/prompt"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	promptreader "github.com/jfrog/jfrog-client-go/utils/prompt"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func CreateBuildConfig(global, allowDeployment bool, confType utils.ProjectType) (err error) {
	projectDir, err := utils.GetProjectDir(global)
	if err != nil {
		return err
	}
	err = fileutils.CreateDirIfNotExist(projectDir)
	if err != nil {
		return err
	}
	configFilePath := filepath.Join(projectDir, confType.String()+".yaml")
	if err := prompt.VerifyConfigFile(configFilePath); err != nil {
		return err
	}
	configResult := &ConfigFile{}
	configResult.Version = prompt.BUILD_CONF_VERSION
	configResult.ConfigType = confType.String()
	switch confType {
	case utils.Nuget:
		err = nugetConfigeFile(configResult)
	case utils.Npm:
		err = npmConfigeFile(configResult)
	case utils.Go:
		err = goConfigeFile(configResult)
	case utils.Maven:
		err = mavenConfigeFile(configResult)
	case utils.Gradle:
		err = gradleConfigeFile(configResult)
	}
	if err != nil {
		return errorutils.CheckError(err)
	}
	resBytes, err := yaml.Marshal(&configResult)
	if err != nil {
		return errorutils.CheckError(err)
	}
	err = ioutil.WriteFile(configFilePath, resBytes, 0644)
	if err != nil {
		return errorutils.CheckError(err)
	}
	log.Info(confType.String() + " build config successfully created.")
	return nil
}

func npmConfigeFile(configResult *ConfigFile) error {
	err := createResolverRepo(configResult)
	if err != nil {
		return err
	}
	return createDeployerRepo(configResult)
}

func mavenConfigeFile(configResult *ConfigFile) error {
	vConfig, err := setArrtifactoryResolver()
	if err != nil {
		return err
	}
	if vConfig.GetBool(prompt.USE_ARTIFACTORY) {
		configResult.Resolver.ServerId = vConfig.GetString(utils.SERVER_ID)
		configResult.Resolver.ReleaseRepo, err = prompt.ReadRepo("Set resolution repository for release dependencies (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL)
		if err != nil {
			return err
		}
		configResult.Resolver.SnapshotRepo, err = prompt.ReadRepo("Set resolution repository for snapshot dependencies (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL)
		if err != nil {
			return err
		}
	}
	vConfig, err = setArrtifactoryDeployer()
	if err != nil {
		return err
	}
	if vConfig.GetBool(prompt.USE_ARTIFACTORY) {
		configResult.Deployer.ServerId = vConfig.GetString(utils.SERVER_ID)
		configResult.Deployer.ReleaseRepo, err = prompt.ReadRepo("Set repository for release artifacts deployment (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL)
		if err != nil {
			return err
		}
		configResult.Deployer.SnapshotRepo, err = prompt.ReadRepo("Set repository for snapshot artifacts deployment (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL)
		if err != nil {
			return err
		}
	}
	return nil
}

func gradleConfigeFile(configResult *ConfigFile) error {
	err := createDeployerRepo(configResult)
	if err != nil {
		return err
	}
	if &configResult.Deployer != nil {
		err = readDescriptors(&configResult.Deployer)
		if err != nil {
			return err
		}
	}
	err = readGradleGlobalConfig(configResult)
	if err != nil {
		return err
	}
	err = createResolverRepo(configResult)
	if err != nil {
		return err
	}
	return nil
}

func nugetConfigeFile(configResult *ConfigFile) error {
	return createResolverRepo(configResult)
}

func goConfigeFile(configResult *ConfigFile) error {
	return npmConfigeFile(configResult)
}

func readGradleGlobalConfig(configResult *ConfigFile) error {
	globalOptions := &promptreader.Array{
		Prompts: []promptreader.Prompt{
			&promptreader.YesNo{
				Msg:     "Is the Gradle Artifactory Plugin already applied in the build script (y/n) [${default}]? ",
				Default: "n",
				Label:   utils.USE_GRADLE_PLUGIN,
			},
			&promptreader.YesNo{
				Msg:     "Use Gradle wrapper (y/n) [${default}]? ",
				Default: "n",
				Label:   utils.USE_GRADLE_WRAPPER,
			},
		},
	}
	err := globalOptions.Read()
	if err != nil {
		return errorutils.CheckError(err)
	}
	vConfig := globalOptions.GetResults()
	configResult.UsePlugin = vConfig.GetBool(utils.USE_GRADLE_PLUGIN)
	configResult.UseWrapper = vConfig.GetBool(utils.USE_GRADLE_WRAPPER)
	return nil
}

func createDeployerRepo(configResult *ConfigFile) error {
	vConfig, err := setArrtifactoryDeployer()
	if err != nil {
		return err
	}
	if vConfig.GetBool(prompt.USE_ARTIFACTORY) {
		configResult.Deployer.ServerId = vConfig.GetString(utils.SERVER_ID)
		configResult.Deployer.Repo, err = prompt.ReadRepo("Set repository for dependencies deployment (press Tab for options): ", vConfig, utils.LOCAL, utils.VIRTUAL)
		if err != nil {
			return err
		}
	}
	return nil
}

func createResolverRepo(configResult *ConfigFile) error {
	vConfig, err := setArrtifactoryResolver()
	if err != nil {
		return err
	}
	if vConfig.GetBool(prompt.USE_ARTIFACTORY) {
		configResult.Resolver.ServerId = vConfig.GetString(utils.SERVER_ID)
		configResult.Resolver.Repo, err = prompt.ReadRepo("Set repository for dependencies resolution (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL, utils.LOCAL)
		if err != nil {
			return err
		}
	}
	return nil
}

func setArrtifactoryResolver() (*viper.Viper, error) {
	return prompt.ReadArtifactoryServer("Resolve dependencies from Artifactory (y/n) [${default}]? ")
}
func setArrtifactoryDeployer() (*viper.Viper, error) {
	return prompt.ReadArtifactoryServer("Deploy project dependencies to Artifactory (y/n) [${default}]? ")
}
func readDescriptors(deployer *utils.Repository) error {
	descriptors := &promptreader.Array{
		Prompts: []promptreader.Prompt{
			&promptreader.YesNo{
				Msg:     "Deploy Maven descriptor (y/n) [${default}]? ",
				Default: "n",
				Label:   utils.MAVEN_DESCRIPTOR,
			},
			&promptreader.YesNo{
				Msg:     "Deploy Ivy descriptor (y/n) [${default}]? ",
				Default: "n",
				Label:   utils.IVY_DESCRIPTOR,
				Yes: &promptreader.Array{
					Prompts: []promptreader.Prompt{
						&promptreader.Simple{
							Msg:   "Set ivy pattern, [organization]/[module]/ivy-[revision].xml: ",
							Label: utils.IVY_PATTERN,
						},
						&promptreader.Simple{
							Msg:   "Set ivy artifact pattern, [organization]/[module]/[revision]/[artifact]-[revision](-[classifier]).[ext]: ",
							Label: utils.ARTIFACT_PATTERN,
						},
					},
				},
			},
		},
	}
	err := descriptors.Read()
	if err != nil {
		return errorutils.CheckError(err)
	}
	vConfig := descriptors.GetResults()
	deployer.DeployMavenDesc = vConfig.GetBool(utils.MAVEN_DESCRIPTOR)
	deployer.DeployIvyDesc = vConfig.GetBool(utils.IVY_DESCRIPTOR)
	if deployer.DeployIvyDesc {
		deployer.IvyPattern = vConfig.GetString(utils.IVY_PATTERN)
		deployer.ArtifactsPattern = vConfig.GetString(utils.ARTIFACT_PATTERN)
	}

	return nil
}

type ConfigFile struct {
	prompt.CommonConfig `yaml:"common,inline"`
	Resolver            utils.Repository `yaml:"resolver,omitempty"`
	Deployer            utils.Repository `yaml:"deployer,omitempty"`
	UsePlugin           bool             `yaml:"usePlugin,omitempty"`
	UseWrapper          bool             `yaml:"useWrapper,omitempty"`
}
