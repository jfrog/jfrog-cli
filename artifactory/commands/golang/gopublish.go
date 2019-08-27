package golang

import (
	"errors"
	commandutils "github.com/jfrog/jfrog-cli-go/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/golang"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/golang/project"
	"github.com/jfrog/jfrog-client-go/artifactory/services/go"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/version"
	"os/exec"
	"strings"
)

const minSupportedArtifactoryVersion = "6.2.0"

type GoPublishCommand struct {
	publishPackage     bool
	buildConfiguration *utils.BuildConfiguration
	dependencies       string
	version            string
	result             *commandutils.Result
	utils.RepositoryConfig
}

func NewGoPublishCommand() *GoPublishCommand {
	return &GoPublishCommand{result: new(commandutils.Result)}
}

func (gpc *GoPublishCommand) Result() *commandutils.Result {
	return gpc.result
}

func (gpc *GoPublishCommand) SetVersion(version string) *GoPublishCommand {
	gpc.version = version
	return gpc
}

func (gpc *GoPublishCommand) SetDependencies(dependencies string) *GoPublishCommand {
	gpc.dependencies = dependencies
	return gpc
}

func (gpc *GoPublishCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *GoPublishCommand {
	gpc.buildConfiguration = buildConfiguration
	return gpc
}

func (gpc *GoPublishCommand) SetPublishPackage(publishPackage bool) *GoPublishCommand {
	gpc.publishPackage = publishPackage
	return gpc
}

func (gpc *GoPublishCommand) Run() error {
	err := validatePrerequisites()
	if err != nil {
		return err
	}

	err = golang.LogGoVersion()
	if err != nil {
		return err
	}

	rtDetails, err := gpc.RtDetails()
	if errorutils.CheckError(err) != nil {
		return err
	}
	serviceManager, err := utils.CreateServiceManager(rtDetails, false)
	if err != nil {
		return err
	}
	artifactoryVersion, err := serviceManager.GetConfig().GetArtDetails().GetVersion()
	if err != nil {
		return err
	}

	version := version.NewVersion(artifactoryVersion)
	if !version.AtLeast(minSupportedArtifactoryVersion) {
		return errorutils.CheckError(errors.New("This operation requires Artifactory version 6.2.0 or higher."))
	}

	buildName := gpc.buildConfiguration.BuildName
	buildNumber := gpc.buildConfiguration.BuildNumber
	isCollectBuildInfo := len(buildName) > 0 && len(buildNumber) > 0
	if isCollectBuildInfo {
		err = utils.SaveBuildGeneralDetails(buildName, buildNumber)
		if err != nil {
			return err
		}
	}

	goProject, err := project.Load(gpc.version)
	if err != nil {
		return err
	}

	// Publish the package to Artifactory
	if gpc.publishPackage {
		err = goProject.PublishPackage(gpc.TargetRepo(), buildName, buildNumber, serviceManager)
		if err != nil {
			return err
		}
	}

	result := gpc.Result()
	if gpc.dependencies != "" {
		// Publish the package dependencies to Artifactory
		depsList := strings.Split(gpc.dependencies, ",")
		err = goProject.LoadDependencies()
		if err != nil {
			return err
		}
		succeeded, failed, err := goProject.PublishDependencies(gpc.TargetRepo(), serviceManager, depsList)
		result.SetSuccessCount(succeeded)
		result.SetFailCount(failed)
		if err != nil {
			return err
		}
	}
	if gpc.publishPackage {
		result.SetSuccessCount(result.SuccessCount() + 1)
	}

	// Publish the build-info to Artifactory
	if isCollectBuildInfo {
		if len(goProject.Dependencies()) == 0 {
			// No dependencies were published but those dependencies need to be loaded for the build info.
			goProject.LoadDependencies()
		}
		err = goProject.CreateBuildInfoDependencies(version.AtLeast(_go.ArtifactoryMinSupportedVersionForInfoFile))
		if err != nil {
			return err
		}
		err = utils.SaveBuildInfo(buildName, buildNumber, goProject.BuildInfo(true, gpc.buildConfiguration.Module))
	}

	return err
}

func (gpc *GoPublishCommand) CommandName() string {
	return "rt_go_publish"
}

func validatePrerequisites() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}
