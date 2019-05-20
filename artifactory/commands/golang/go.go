package golang

import (
	"github.com/jfrog/gocmd"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/golang"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/golang/project"
	"github.com/jfrog/jfrog-cli-go/utils/config"
)

type GoCommand struct {
	noRegistry         bool
	publishDeps        bool
	goArg              []string
	buildConfiguration *utils.BuildConfiguration
	GoParamsCommand
}

func NewGoCommand() *GoCommand {
	return &GoCommand{}
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

func (gc *GoCommand) SetTargetRepo(targetRepo string) *GoCommand {
	gc.targetRepo = targetRepo
	return gc
}

func (gc *GoCommand) SetArtifactoryDetails(details *config.ArtifactoryDetails) *GoCommand {
	gc.rtDetails = details
	return gc
}

func (gc *GoCommand) CommandName() string {
	return "rt_go"
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

	serviceManager, err := utils.CreateServiceManager(gc.rtDetails, false)
	if err != nil {
		return err
	}

	err = gocmd.RunWithFallbacksAndPublish(gc.goArg, gc.targetRepo, gc.noRegistry, gc.publishDeps, serviceManager)
	if err != nil {
		return err
	}
	if isCollectBuildInfo {
		err = goProject.LoadDependencies()
		if err != nil {
			return err
		}
		err = utils.SaveBuildInfo(buildName, buildNumber, goProject.BuildInfo(false))
	}

	return err
}
