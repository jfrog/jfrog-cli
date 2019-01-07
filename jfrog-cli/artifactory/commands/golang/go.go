package golang

import (
	"errors"
	"github.com/jfrog/gocmd"
	"github.com/jfrog/gocmd/golang"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang/project"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/version"
	"os/exec"
	"strings"
)

func Publish(publishPackage bool, dependencies, targetRepo, version, buildName, buildNumber string, details *config.ArtifactoryDetails) (succeeded, failed int, err error) {
	err = validatePrerequisites()
	if err != nil {
		return
	}

	serviceManager, err := utils.CreateServiceManager(details, false)
	if err != nil {
		return 0, 0, err
	}
	artifactoryVersion, err := serviceManager.GetConfig().GetArtDetails().GetVersion()
	if err != nil {
		return
	}

	if !isMinSupportedVersion(artifactoryVersion) {
		return 0, 0, errorutils.CheckError(errors.New("This operation requires Artifactory version 6.2.0 or higher."))
	}

	isCollectBuildInfo := len(buildName) > 0 && len(buildNumber) > 0
	if isCollectBuildInfo {
		err = utils.SaveBuildGeneralDetails(buildName, buildNumber)
		if err != nil {
			return
		}
	}

	goProject, err := project.Load(version)
	if err != nil {
		return
	}

	// Publish the package to Artifactory
	if publishPackage {
		err = goProject.PublishPackage(targetRepo, buildName, buildNumber, serviceManager)
		if err != nil {
			return
		}
	}

	if dependencies != "" {
		// Publish the package dependencies to Artifactory
		depsList := strings.Split(dependencies, ",")
		err = goProject.LoadDependencies()
		if err != nil {
			return
		}
		succeeded, failed, err = goProject.PublishDependencies(targetRepo, serviceManager, depsList)
		if err != nil {
			return
		}
	}
	if publishPackage {
		succeeded++
	}

	// Publish the build-info to Artifactory
	if isCollectBuildInfo {
		if len(goProject.Dependencies()) == 0 {
			// No dependencies were published but those dependencies need to be loaded for the build info.
			goProject.LoadDependencies()
		}
		err = utils.SaveBuildInfo(buildName, buildNumber, goProject.BuildInfo(true))
	}

	return
}

func isMinSupportedVersion(artifactoryVersion string) bool {
	minSupportedArtifactoryVersion := "6.2.0"
	if version.Compare(artifactoryVersion, minSupportedArtifactoryVersion) < 0 && artifactoryVersion != "development" {
		return false
	}
	return true
}

func ExecuteGo(tidyEnum golang.TidyEnum, noRegistry bool, goArg, targetRepo, buildName, buildNumber string, details *config.ArtifactoryDetails) error {
	isCollectBuildInfo := len(buildName) > 0 && len(buildNumber) > 0
	if isCollectBuildInfo {
		err := utils.SaveBuildGeneralDetails(buildName, buildNumber)
		if err != nil {
			return err
		}
	}
	// The version is not necessary because we are collecting the dependencies only.
	goProject, err := project.Load("-")
	if err != nil {
		return err
	}

	serviceManager, err := utils.CreateServiceManager(details, false)
	if err != nil {
		return err
	}

	err = gocmd.ExecuteGo(goArg, targetRepo, tidyEnum, noRegistry, serviceManager)

	if isCollectBuildInfo {
		err = goProject.LoadDependencies()
		if err != nil {
			return err
		}
		err = utils.SaveBuildInfo(buildName, buildNumber, goProject.BuildInfo(false))
	}

	return err
}

func validatePrerequisites() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}