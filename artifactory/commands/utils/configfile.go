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

/*
	The logic between Deployer/Resolver:
	Commands that only have Resolver capability, must set the Resolver at the config file,
	otherwise both of Deployer Resolver are optinal, and as a result the user decide if to set those or not.
*/
func CreateBuildConfig(global, allowDeployment bool, confType utils.ProjectType) error {
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

	var vConfig *viper.Viper
	allowResovle := true
	configResult := &ConfigFile{}
	configResult.Version = prompt.BUILD_CONF_VERSION
	configResult.ConfigType = confType.String()

	if allowDeployment {
		vConfig, err = prompt.ReadArtifactoryServer("Deploy project dependencies to Artifactory (y/n) [${default}]? ")
		if err != nil {
			return err
		}
		if vConfig.GetBool(prompt.USE_ARTIFACTORY) {
			configResult.Deployer.ServerId = vConfig.GetString(utils.SERVER_ID)
			if confType == utils.Maven {
				configResult.Resolver.ReleaseRepo, err = prompt.ReadRepo("Set resolution repository for release dependencies (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL)
				if err != nil {
					return err
				}
				configResult.Resolver.SnapshotRepo, err = prompt.ReadRepo("Set resolution repository for snapshot dependencies (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL)
				if err != nil {
					return err
				}
			} else {
				configResult.Deployer.Repo, err = prompt.ReadRepo("Set repository for dependencies deployment (press Tab for options): ", vConfig, utils.LOCAL, utils.VIRTUAL)
				if err != nil {
					return err
				}
				if confType == utils.Gradle {
					err = readDescriptors(&configResult.Deployer)
					if err != nil {
						return err
					}
				}
			}
		}
		vConfig, err = prompt.ReadArtifactoryServer("Resolve dependencies from Artifactory (y/n) [${default}]? ")
		if err != nil {
			return err
		}
		allowResovle = vConfig.GetBool(prompt.USE_ARTIFACTORY)
	} else {
		configResult.Resolver.ServerId, vConfig, err = prompt.ReadServerId()
		if err != nil {
			return err
		}
	}

	if allowResovle {
		configResult.Resolver.ServerId = vConfig.GetString(utils.SERVER_ID)
		if confType == utils.Maven {
			configResult.Deployer.ReleaseRepo, err = prompt.ReadRepo("Set repository for release artifacts deployment (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL)
			if err != nil {
				return err
			}
			configResult.Deployer.SnapshotRepo, err = prompt.ReadRepo("Set repository for snapshot artifacts deployment (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL)
			if err != nil {
				return err
			}
		} else {
			configResult.Resolver.Repo, err = prompt.ReadRepo("Set repository for dependencies resolution (press Tab for options): ", vConfig, utils.REMOTE, utils.VIRTUAL)
			if err != nil {
				return err
			}
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
	log.Info(confType.String() + " build config successfully created.")
	return nil
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
