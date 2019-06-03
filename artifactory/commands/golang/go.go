package golang

import (
	"github.com/jfrog/gocmd"
	"github.com/jfrog/gocmd/params"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/golang"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/golang/project"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services/go"
	"github.com/jfrog/jfrog-client-go/utils/version"
)

const GoCommandName = "rt_go"

type GoCommand struct {
	noRegistry         bool
	publishDeps        bool
	goArg              []string
	buildConfiguration *utils.BuildConfiguration
	deployerParams     *GoParamsCommand
	resolverParams     *GoParamsCommand
}

func NewGoCommand() *GoCommand {
	return &GoCommand{}
}

func (gc *GoCommand) SetResolverParams(resolverParams *GoParamsCommand) *GoCommand {
	gc.resolverParams = resolverParams
	return gc
}

func (gc *GoCommand) SetDeployerParams(deployerParams *GoParamsCommand) *GoCommand {
	gc.deployerParams = deployerParams
	return gc
}

func (gc *GoCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *GoCommand {
	gc.buildConfiguration = buildConfiguration
	return gc
}

func (gc *GoCommand) SetNoRegistry(noRegistry bool) *GoCommand {
	gc.noRegistry = noRegistry
	return gc
}

func (gc *GoCommand) SetPublishDeps(publishDeps bool) *GoCommand {
	gc.publishDeps = publishDeps
	return gc
}

func (gc *GoCommand) SetGoArg(goArg []string) *GoCommand {
	gc.goArg = goArg
	return gc
}

func (gc *GoCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	if gc.deployerParams != nil && !gc.deployerParams.isRtDetailsEmpty() {
		return gc.deployerParams.RtDetails()
	}
	return gc.resolverParams.RtDetails()
}

func (gc *GoCommand) CommandName() string {
	return GoCommandName
}

func (gc *GoCommand) Run() error {
	err := golang.LogGoVersion()
	if err != nil {
		return err
	}
	buildName := gc.buildConfiguration.BuildName
	buildNumber := gc.buildConfiguration.BuildNumber
	isCollectBuildInfo := len(buildName) > 0 && len(buildNumber) > 0
	if isCollectBuildInfo {
		err = utils.SaveBuildGeneralDetails(buildName, buildNumber)
		if err != nil {
			return err
		}
	}
	// The version is not necessary because we are collecting the dependencies only.
	goProject, err := project.Load("-")
	if err != nil {
		return err
	}

	resolverServiceManager, err := utils.CreateServiceManager(gc.resolverParams.rtDetails, false)
	if err != nil {
		return err
	}
	resolverParams := &params.Params{}
	resolverParams.SetRepo(gc.resolverParams.TargetRepo()).SetServiceManager(resolverServiceManager)
	goInfo := &params.ResolverDeployer{}
	goInfo.SetResolver(resolverParams)
	var deployerServiceManager *artifactory.ArtifactoryServicesManager
	if gc.publishDeps {
		deployerServiceManager, err = utils.CreateServiceManager(gc.deployerParams.rtDetails, false)
		if err != nil {
			return err
		}
		deployerParams := &params.Params{}
		deployerParams.SetRepo(gc.deployerParams.TargetRepo()).SetServiceManager(deployerServiceManager)
		goInfo.SetDeployer(deployerParams)
	}

	err = gocmd.RunWithFallbacksAndPublish(gc.goArg, gc.noRegistry, gc.publishDeps, goInfo)
	if err != nil {
		return err
	}
	if isCollectBuildInfo {
		includeInfoFiles, err := shouldIncludeInfoFiles(deployerServiceManager, resolverServiceManager)
		if err != nil {
			return err
		}
		err = goProject.LoadDependencies()
		if err != nil {
			return err
		}
		err = goProject.CreateBuildInfoDependencies(includeInfoFiles)
		if err != nil {
			return err
		}
		err = utils.SaveBuildInfo(buildName, buildNumber, goProject.BuildInfo(false))
	}

	return err
}

// Returns true/false if info files should be included in the build info.
func shouldIncludeInfoFiles(deployerServiceManager *artifactory.ArtifactoryServicesManager, resolverServiceManager *artifactory.ArtifactoryServicesManager) (bool, error) {
	var artifactoryVersion string
	var err error
	if deployerServiceManager != nil {
		artifactoryVersion, err = deployerServiceManager.GetConfig().GetArtDetails().GetVersion()
	} else {
		artifactoryVersion, err = resolverServiceManager.GetConfig().GetArtDetails().GetVersion()
	}
	if err != nil {
		return false, err
	}
	version := version.NewVersion(artifactoryVersion)
	includeInfoFiles := version.AtLeast(_go.ArtifactoryMinSupportedVersionForInfoFile)
	return includeInfoFiles, nil
}
