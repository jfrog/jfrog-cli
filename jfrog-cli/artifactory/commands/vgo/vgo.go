package vgo

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/vgo/project"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"os/exec"
)

func PublishDependencies(targetRepo string, threads int, details *config.ArtifactoryDetails) (int, int, error) {
	err := validatePrerequisites()
	if err != nil {
		return 0, 0, err
	}

	log.Info("Publishing project dependencies...")
	// The version is not necessary because we are publishing only the dependencies.
	goProject, err := project.Load("-")
	if err != nil {
		return 0, 0, err
	}

	succeeded := 0
	dependencies := goProject.Dependencies()
	for _, dependency := range dependencies {
		err = dependency.Publish(targetRepo, details)
		if err != nil {
			log.Error(err)
			continue
		}
		succeeded++
	}

	failed := len(dependencies) - succeeded
	if failed > 0 {
		err = errors.New("Publishing project dependencies finished with errors. Please review the logs.")
	}
	return succeeded, failed, err
}

func Publish(targetRepo, version, buildName, buildNumber string, details *config.ArtifactoryDetails) error {
	err := validatePrerequisites()
	if err != nil {
		return err
	}

	log.Info("Publishing project...")
	goProject, err := project.Load(version)
	if err != nil {
		return err
	}

	isCollectBuildInfo := len(buildName) > 0 && len(buildNumber) > 0
	if isCollectBuildInfo {
		err := utils.SaveBuildGeneralDetails(buildName, buildNumber)
		if err != nil {
			return err
		}
	}

	err = goProject.Publish(targetRepo, buildName, buildNumber, details)
	if err != nil {
		log.Error(err)
		return errors.New("Publishing project finished with errors. Please review the logs.")
	}

	buildInfo := goProject.BuildInfo()
	if isCollectBuildInfo {
		return utils.SaveBuildInfo(buildName, buildNumber, buildInfo)
	}
	return nil
}

func validatePrerequisites() error {
	_, err := exec.LookPath("vgo")
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}
