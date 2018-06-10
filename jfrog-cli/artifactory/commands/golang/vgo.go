package golang

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang/project"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	cliutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"os/exec"
	"strings"
)

func PublishDependencies(targetRepo string, details *config.ArtifactoryDetails, includeDepSlice []string) (int, int, error) {
	err := validatePrerequisites()
	if err != nil {
		return 0, 0, err
	}

	log.Info("Publishing project dependencies...")
	includeDep := cliutils.GetMapFromStringSlice(includeDepSlice, ":")
	// The version is not necessary because we are publishing only the dependencies.
	goProject, err := project.Load("-")
	if err != nil {
		return 0, 0, err
	}

	succeeded := 0
	skip := 0
	_, includeAll := includeDep["ALL"]
	dependencies := goProject.Dependencies()
	for _, dependency := range dependencies {
		depToBeIncluded := false
		id := strings.Split(dependency.GetId(), ":")
		if includedVersion, included := includeDep[id[0]]; included && strings.EqualFold(includedVersion, id[1]) {
			depToBeIncluded = true
		}
		if includeAll || depToBeIncluded {
			err = dependency.Publish(targetRepo, details)
			if err != nil {
				err = errors.New("Failed to publish " + dependency.GetId() + " due to: " + err.Error())
				log.Error("Failed to publish", dependency.GetId(), ":", err)
			} else {
				succeeded++
			}
			continue
		}
		skip++
	}

	failed := len(dependencies) - succeeded - skip
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
