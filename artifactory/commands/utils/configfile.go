package utils

import (
	"io/ioutil"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/prompt"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"gopkg.in/yaml.v2"
)

type ConfigFile struct {
	prompt.CommonConfig    `yaml:"common,inline"`
	ResolveFromArtifactory *bool            `yaml:"-"`
	DeployToArtifactory    *bool            `yaml:"-"`
	Resolver               utils.Repository `yaml:"resolver,omitempty"`
	Deployer               utils.Repository `yaml:"deployer,omitempty"`
	UsePlugin              bool             `yaml:"usePlugin,omitempty"`
	UseWrapper             bool             `yaml:"useWrapper,omitempty"`
}

func NewConfigFile(confType utils.ProjectType, c *cli.Context) *ConfigFile {
	configFile := &ConfigFile{
		CommonConfig: prompt.CommonConfig{
			Version:    prompt.BUILD_CONF_VERSION,
			ConfigType: confType.String(),
		},
	}
	configFile.fillConfigFromFlags(c)
	return configFile
}

func CreateBuildConfig(c *cli.Context, confType utils.ProjectType) (err error) {
	global := c.Bool("global")
	projectDir, err := utils.GetProjectDir(global)
	if err != nil {
		return err
	}
	if err = fileutils.CreateDirIfNotExist(projectDir); err != nil {
		return err
	}
	configFilePath := filepath.Join(projectDir, confType.String()+".yaml")
	interactive := !c.IsSet("interactive") || c.BoolT("interactive")
	if err := prompt.VerifyConfigFile(configFilePath, interactive); err != nil {
		return err
	}
	configFile := NewConfigFile(confType, c)
	if interactive {
		switch confType {
		case utils.Go:
			err = configFile.configGo()
		case utils.Pip:
			err = configFile.configPip()
		case utils.Npm:
			err = configFile.configNpm()
		case utils.Nuget:
			err = configFile.configNuget()
		case utils.Maven:
			err = configFile.configMaven(c)
		case utils.Gradle:
			err = configFile.configGradle(c)
		}
		if err != nil {
			return errorutils.CheckError(err)
		}
	}
	resBytes, err := yaml.Marshal(&configFile)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if err = ioutil.WriteFile(configFilePath, resBytes, 0644); err != nil {
		return errorutils.CheckError(err)
	}
	log.Info(confType.String() + " build config successfully created.")
	return nil
}

// Fill configuration from cli flags
func (configFile *ConfigFile) fillConfigFromFlags(c *cli.Context) {
	// If resolveFromArtifactory isn't set, leave it nil
	if c.IsSet("resolveFromArtifactory") {
		configFile.ResolveFromArtifactory = new(bool)
		*configFile.ResolveFromArtifactory = c.BoolT("resolveFromArtifactory")
	}
	configFile.Resolver.ServerId = c.String("resolutionServerId")
	configFile.Resolver.Repo = c.String("resolutionRepo")

	// If deployToArtifactory isn't set, leave it nil
	if c.IsSet("deployToArtifactory") {
		configFile.DeployToArtifactory = new(bool)
		*configFile.DeployToArtifactory = c.BoolT("deployToArtifactory")
	}
	configFile.Deployer.ServerId = c.String("deploymentServerId")
	configFile.Deployer.Repo = c.String("deploymentRepo")
}

// Fill Maven related configuration from cli flags
func (configFile *ConfigFile) fillMavenConfigFromFlags(c *cli.Context) {
	configFile.Resolver.SnapshotRepo = c.String("resolutionSnapshotRepo")
	configFile.Resolver.ReleaseRepo = c.String("resolutionReleaseRepo")
	configFile.Deployer.SnapshotRepo = c.String("deploymentSnapshotRepo")
	configFile.Deployer.ReleaseRepo = c.String("deploymentReleaseRepo")
}

// Fill Gradle related configuration from cli flags
func (configFile *ConfigFile) fillGradleConfigFromFlags(c *cli.Context) {
	configFile.Deployer.DeployMavenDesc = c.BoolT("deployMavenDescriptors")
	configFile.Deployer.DeployIvyDesc = c.BoolT("deployIvyDescriptors")
	configFile.Deployer.IvyPattern = c.String("ivyPattern")
	configFile.Deployer.ArtifactsPattern = c.String("artifactPattern")
	configFile.UsePlugin = c.Bool("usePlugin")
	configFile.UseWrapper = c.Bool("useWrapper")
}

func (configFile *ConfigFile) configGo() error {
	return configFile.setDeployerResolver()
}

func (configFile *ConfigFile) configPip() error {
	return configFile.setResolver()
}

func (configFile *ConfigFile) configNpm() error {
	return configFile.setDeployerResolver()
}

func (configFile *ConfigFile) configNuget() error {
	return configFile.setResolver()
}

func (configFile *ConfigFile) configMaven(c *cli.Context) error {
	configFile.fillMavenConfigFromFlags(c)
	// Set resolution repositories
	if configFile.ResolveFromArtifactory == nil || *configFile.ResolveFromArtifactory {
		if err := configFile.setResolverId(); err != nil {
			return err
		}
		if configFile.Resolver.ServerId != "" {
			if err := configFile.setRepo(&configFile.Resolver.ReleaseRepo, "Set resolution repository for release dependencies", configFile.Resolver.ServerId); err != nil {
				return err
			}
			if err := configFile.setRepo(&configFile.Resolver.SnapshotRepo, "Set resolution repository for snapshot dependencies", configFile.Resolver.ServerId); err != nil {
				return err
			}
		}
	}
	// Set deployment repositories
	if configFile.DeployToArtifactory == nil || *configFile.DeployToArtifactory {
		if err := configFile.setDeployerId(); err != nil {
			return err
		}
		if configFile.Deployer.ServerId != "" {
			if err := configFile.setRepo(&configFile.Deployer.ReleaseRepo, "Set repository for release artifacts deployment", configFile.Deployer.ServerId); err != nil {
				return err
			}
			return configFile.setRepo(&configFile.Deployer.SnapshotRepo, "Set repository for snapshot artifacts deployment", configFile.Deployer.ServerId)
		}
	}
	return nil
}

func (configFile *ConfigFile) configGradle(c *cli.Context) error {
	configFile.fillGradleConfigFromFlags(c)
	if err := configFile.setDeployerResolver(); err != nil {
		return err
	}
	if configFile.Deployer.ServerId != "" {
		if err := configFile.setMavenIvyDescriptors(c); err != nil {
			return err
		}
	}
	return configFile.readGradleGlobalConfig(c)
}

func (configFile *ConfigFile) readGradleGlobalConfig(c *cli.Context) error {
	var err error
	if !c.IsSet("usePlugin") {
		configFile.UsePlugin, err = prompt.AskYesNo("Is the Gradle Artifactory Plugin already applied in the build script (y/n) [${default}]? ", "n", utils.USE_GRADLE_PLUGIN)
		if err != nil {
			return err
		}
	}
	if !c.IsSet("useWrapper") {
		configFile.UseWrapper, err = prompt.AskYesNo("Use Gradle wrapper (y/n) [${default}]? ", "n", utils.USE_GRADLE_WRAPPER)
	}
	return err
}

func (configFile *ConfigFile) setDeployer() error {
	// Check if the user explicitly set deployToArtifactory=false
	if configFile.DeployToArtifactory != nil && !*configFile.DeployToArtifactory {
		return nil
	}

	// Set deployer id
	if err := configFile.setDeployerId(); err != nil {
		return err
	}

	// Set deployment repository
	if configFile.Deployer.ServerId != "" {
		return configFile.setRepo(&configFile.Deployer.Repo, "Set repository for artifacts deployment", configFile.Deployer.ServerId)
	}
	return nil
}

func (configFile *ConfigFile) setResolver() error {
	// Check if the user explicitly set resolveFromArtifactory=false
	if configFile.ResolveFromArtifactory != nil && !*configFile.ResolveFromArtifactory {
		return nil
	}

	// Set resolver id
	if err := configFile.setResolverId(); err != nil {
		return err
	}

	// Set resolution repository
	if configFile.Resolver.ServerId != "" {
		return configFile.setRepo(&configFile.Resolver.Repo, "Set repository for dependencies resolution", configFile.Resolver.ServerId)
	}
	return nil
}

func (configFile *ConfigFile) setDeployerResolver() error {
	if err := configFile.setResolver(); err != nil {
		return err
	}
	return configFile.setDeployer()
}

func (configFile *ConfigFile) setResolverId() error {
	return configFile.setServerId(&configFile.Resolver.ServerId, configFile.ResolveFromArtifactory, "Resolve dependencies from Artifactory")
}

func (configFile *ConfigFile) setDeployerId() error {
	return configFile.setServerId(&configFile.Deployer.ServerId, configFile.DeployToArtifactory, "Deploy project artifacts to Artifactory")
}

func (configFile *ConfigFile) setServerId(serverId *string, useArtifactory *bool, useArtifactoryQuestion string) error {
	if *serverId != "" {
		return nil
	}
	var err error
	if useArtifactory == nil {
		// Ask whether to use artifactory and ask for the serverId
		*serverId, err = prompt.ReadArtifactoryServer(useArtifactoryQuestion + " (y/n) [${default}]? ")
	} else if *useArtifactory {
		// Ask for the serverId only
		*serverId, err = prompt.ReadArtifactoryServer("")
	}
	return err
}

func (configFile *ConfigFile) setRepo(repo *string, message string, serverId string) error {
	var err error
	if *repo == "" {
		*repo, err = prompt.ReadRepo(message+" (press Tab for options): ", serverId, utils.REMOTE, utils.VIRTUAL)
	}
	return err
}

func (configFile *ConfigFile) setMavenIvyDescriptors(c *cli.Context) error {
	var err error
	if !c.IsSet("deployMavenDescriptors") {
		configFile.Deployer.DeployMavenDesc, err = prompt.AskYesNo("Deploy Maven descriptors (y/n) [${default}]? ", "n", utils.MAVEN_DESCRIPTOR)
		if err != nil {
			return err
		}
	}

	if !c.IsSet("deployIvyDescriptors") {
		configFile.Deployer.DeployIvyDesc, err = prompt.AskYesNo("Deploy Ivy descriptors (y/n) [${default}]? ", "n", utils.IVY_DESCRIPTOR)
		if err != nil {
			return err
		}
	}

	if configFile.Deployer.DeployIvyDesc {
		if !c.IsSet("ivyPattern") {
			configFile.Deployer.IvyPattern, err = prompt.AskString("Set Ivy pattern [${default}]:", "[organization]/[module]/ivy-[revision].xml", utils.IVY_PATTERN)
			if err != nil {
				return err
			}
		}

		if !c.IsSet("artifactPattern") {
			configFile.Deployer.ArtifactsPattern, err = prompt.AskString("Set Ivy artifact pattern [${default}]:", "[organization]/[module]/[revision]/[artifact]-[revision](-[classifier]).[ext]", utils.ARTIFACT_PATTERN)
		}
	}
	return err
}
