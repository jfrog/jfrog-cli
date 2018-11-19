package golang

import (
	"errors"
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

func ExecuteGo(depsTidy, noRegistry bool, goArg, targetRepo, buildName, buildNumber string, details *config.ArtifactoryDetails) error {
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
		if shouldDownloadDependencyDirectly(err, noRegistry) {
			log.Info("Received", err.Error(), "from Artifactory. Trying download the dependencies from the VCS and upload them to Artifactory...")
			err = retrieveAndPublish(targetRepo, goArg, depsTidy, details)
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

func retrieveAndPublish(targetRepo, goArg string, depsTidy bool, details *config.ArtifactoryDetails) error {
	// Need to run Go without Artifactory to resolve all dependencies.
	cache := goutils.GetStaticCache()
	wd, rootDir, err := collectDependenciesPopulateAndPublish(targetRepo, depsTidy, &cache, details)
	if err != nil {
		if depsTidy {
			log.Debug("Received and error:", err)
		} else {
			return err
		}
	}
	// Lets run the same command again now that all the dependencies were downloaded.
	// Need to run only if the command is not go mod download or go mod tidy since this was run by the CLI to download and publish to Artifactory
	log.Info("Finished with following stats:", cache.GetSuccess(), "/", cache.GetTotal())
	if !strings.Contains(goArg, "mod download") || !strings.Contains(goArg, "mod tidy") {
		if depsTidy {
			// Remove first the go.sum file that contains previous information from the go mod tidy run without Artifactory
			err = removeGoSumFile(wd, rootDir)
			if err != nil {
				log.Debug("Received and error:", err)
			}
		}
		err = goutils.RunGo(goArg)
	}
	return err
}

// Returns true if needed to download and publish to Artifactory.
func shouldDownloadDependencyDirectly(err error, noRegistry bool) bool {
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
func collectDependenciesPopulateAndPublish(targetRepo string, depsTidy bool, cache *goutils.DynamicCache, details *config.ArtifactoryDetails) (wd, rootProjectDir string, err error) {
	err = os.Unsetenv(goutils.GOPROXY)
	if err != nil {
		return
	}
	dependenciesToPublish, err := projectDep.CollectProjectNeededDependencies(targetRepo, cache, details)
	if err != nil {
		return
	}
	if len(dependenciesToPublish) > 0 {
		var dependency projectDep.GoPackage
		if depsTidy {
			dependency = &projectDep.PackageWithDeps{}
			err = dependency.Init()
			defer fileutils.RemoveTempDir()
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

		err = runPopulateAndPublishDependencies(targetRepo, depsTidy, dependency, dependenciesToPublish, cache, details)
		if err != nil {
			return
		}
	}
	return
}

func runPopulateAndPublishDependencies(targetRepo string, depsTidy bool, dependenciesInterface projectDep.GoPackage, dependenciesToPublish map[string]bool, cache *goutils.DynamicCache, details *config.ArtifactoryDetails) error {
	cachePath, err := projectDep.GetCachePath()
	if err != nil {
		return err
	}

	dependencies, err := projectDep.GetDependencies(cachePath, dependenciesToPublish)
	if err != nil {
		return err
	}

	cache.IncreaseTotal(len(dependencies))
	for _, dep := range dependencies {
		dependenciesInterface = dependenciesInterface.New(cachePath, dep)
		err := dependenciesInterface.PopulateModIfNeededAndPublish(targetRepo, cache, details)
		if err != nil {
			if depsTidy {
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

	goSumFile := filepath.Join(strings.TrimSpace(rootDir), "go.sum")
	exists, err := fileutils.IsFileExists(goSumFile, false)
	if err != nil {
		return err
	}
	if exists {
		return errorutils.CheckError(os.Remove(goSumFile))
	}
	return nil
}
