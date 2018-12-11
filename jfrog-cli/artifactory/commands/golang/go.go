package golang

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	goutils "github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang/project"
	projectDep "github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang/project/dependencies"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/version"
	"os"
	"os/exec"
	"path/filepath"
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

func ExecuteGo(recursiveTidy, recursiveTidyOverwrite, noRegistry bool, goArg, targetRepo, buildName, buildNumber string, details *config.ArtifactoryDetails) error {
	isCollectBuildInfo := len(buildName) > 0 && len(buildNumber) > 0
	if isCollectBuildInfo {
		err := utils.SaveBuildGeneralDetails(buildName, buildNumber)
		if err != nil {
			return err
		}
	}

	if !noRegistry {
		goutils.SetGoProxyEnvVar(details, targetRepo)
	}
	err := goutils.RunGo(goArg)
	if err != nil {
		if dependencyNotFoundInArtifactory(err, noRegistry) {
			log.Info("Received", err.Error(), "from Artifactory. Trying download the dependencies from the VCS and upload them to Artifactory...")
			err = downloadFromVcsAndPublish(targetRepo, goArg, recursiveTidy, recursiveTidyOverwrite, details)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if isCollectBuildInfo {
		// The version is not necessary because we are collecting the dependencies only.
		goProject, err := project.Load("-")
		if err != nil {
			return err
		}
		err = goProject.LoadDependencies()
		if err != nil {
			return err
		}
		err = utils.SaveBuildInfo(buildName, buildNumber, goProject.BuildInfo(false))
	}

	return err
}

// Downloads all dependencies from VCS and publish them to Artifactory.
func downloadFromVcsAndPublish(targetRepo, goArg string, recursiveTidy, recursiveTidyOverwrite bool, details *config.ArtifactoryDetails) error {
	// Need to run Go without Artifactory to resolve all dependencies.
	cache := goutils.DependenciesCache{}
	wd, rootDir, err := collectDependenciesPopulateAndPublish(targetRepo, recursiveTidy, recursiveTidyOverwrite, &cache, details)
	if err != nil {
		if !recursiveTidy {
			return err
		}
		log.Error("Received an error:", err)
	}
	// Lets run the same command again now that all the dependencies were downloaded.
	// Need to run only if the command is not go mod download or go mod tidy since this was run by the CLI to download and publish to Artifactory
	log.Info(fmt.Sprintf("Done building and publishing %d go dependencies to Artifactory out of a total of %d dependencies.", cache.GetSuccesses(), cache.GetTotal()))
	if !strings.Contains(goArg, "mod download") || !strings.Contains(goArg, "mod tidy") {
		if recursiveTidy {
			// Remove the go.sum file, since it includes information which is not up to date (it was created by the "go mod tidy" command executed without Artifactory
			err = removeGoSumFile(wd, rootDir)
			if err != nil {
				log.Error("Received an error:", err)
			}
		}
		err = goutils.RunGo(goArg)
	}
	return err
}

// Returns true if a dependency was not found Artifactory.
func dependencyNotFoundInArtifactory(err error, noRegistry bool) bool {
	if !noRegistry && strings.EqualFold(err.Error(), "404 Not Found") {
		return true
	}
	return false
}

func validatePrerequisites() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

// Download the dependencies from VCS and publish them to Artifactory.
func collectDependenciesPopulateAndPublish(targetRepo string, recursiveTidy, recursiveTidyOverwrite bool, cache *goutils.DependenciesCache, details *config.ArtifactoryDetails) (wd, rootProjectDir string, err error) {
	err = os.Unsetenv(goutils.GOPROXY)
	if err != nil {
		return
	}
	dependenciesToPublish, err := projectDep.CollectProjectDependencies(targetRepo, cache, details)
	if err != nil || len(dependenciesToPublish) == 0 {
		return
	}

	var dependency projectDep.GoPackage
	if recursiveTidy {
		err = fileutils.CreateTempDirPath()
		if err != nil {
			return "", "", err
		}
		defer fileutils.RemoveTempDir()

		dependency = &projectDep.PackageWithDeps{}
		err = dependency.Init()
		if err != nil {
			return
		}
		wd, err = os.Getwd()
		if err != nil {
			return "", "", errorutils.CheckError(err)
		}

		// Preparations before starting
		rootProjectDir, err = goutils.GetRootDir()
		if err != nil {
			return
		}
	} else {
		dependency = &projectDep.Package{}
	}

	err = runPopulateAndPublishDependencies(targetRepo, recursiveTidy, recursiveTidyOverwrite, dependency, dependenciesToPublish, cache, details)
	return
}

func runPopulateAndPublishDependencies(targetRepo string, recursiveTidy, recursiveTidyOverwrite bool, dependenciesInterface projectDep.GoPackage, dependenciesToPublish map[string]bool, cache *goutils.DependenciesCache, details *config.ArtifactoryDetails) error {
	cachePath, err := projectDep.GetCachePath()
	if err != nil {
		return err
	}

	dependencies, err := projectDep.GetDependencies(cachePath, dependenciesToPublish)
	if err != nil {
		return err
	}

	cache.IncrementTotal(len(dependencies))
	for _, dep := range dependencies {
		dependenciesInterface = dependenciesInterface.New(cachePath, dep, recursiveTidyOverwrite)
		err := dependenciesInterface.PopulateModAndPublish(targetRepo, cache, details)
		if err != nil {
			if recursiveTidy {
				log.Warn(err)
				continue
			}
			return err
		}
	}
	return nil
}

func removeGoSumFile(wd, rootDir string) error {
	log.Debug("Changing back to the working directory")
	err := os.Chdir(wd)
	if err != nil {
		return errorutils.CheckError(err)
	}

	goSumFile := filepath.Join(rootDir, "go.sum")
	exists, err := fileutils.IsFileExists(goSumFile, false)
	if err != nil {
		return err
	}
	if exists {
		return errorutils.CheckError(os.Remove(goSumFile))
	}
	return nil
}
