package dependencies

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	golangutil "github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/global"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services/go"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Load() ([]Dependency, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return nil, err
	}
	return loadDependencies(cachePath)
}

func GetCachePath() (string, error) {
	goPath, err := getGOPATH()
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return filepath.Join(goPath, "pkg", "mod", "cache", "download"), nil
}

// Represent go dependency project.
// Includes publishing capabilities and build info dependencies.
type Dependency struct {
	buildInfoDependencies  []buildinfo.Dependency
	id                     string
	modContent             []byte
	zipPath                string
	version                string
	transitiveDependencies []Dependency
}

func (dependency *Dependency) GetId() string {
	return dependency.id
}

func (dependency *Dependency) GetModContent() []byte {
	return dependency.modContent
}

func (dependency *Dependency) SetModContent(modContent []byte) {
	dependency.modContent = modContent
}

func (dependency *Dependency) GetZipPath() string {
	return dependency.zipPath
}

func (dependency *Dependency) GetTransitiveDependency() []Dependency {
	return dependency.transitiveDependencies
}

func (dependency *Dependency) PopulateDependenciesModAndPublish(cachePath, tempDir, targetRepo string, depsTidy, shouldRunGoModCommand bool, details *config.ArtifactoryDetails, regExp *RegExp) error {
	var path string
	var err error
	output := map[string]bool{}
	if depsTidy {
		global := global.GetGlobalVariables()
		dependenciesMap := global.GetGlobalMap()
		published, _ := dependenciesMap[dependency.GetId()]
		if published {
			log.Debug("Overwriting the mod file in the cache from the one from Artifactory", dependency.GetId())
			moduleAndVersion := strings.Split(dependency.GetId(), ":")
			path = overwriteModFileWithinCache(cachePath, targetRepo, moduleAndVersion[0], moduleAndVersion[1], details, httpclient.NewDefaultHttpClient())
			err = dependency.UpdateModContent(depsTidy, path)
			logErrorIfOccurred(err)
			shouldRunGoModCommand = false
		}
		log.Debug("Entering dependency", dependency.GetId())
		path, output, err = dependency.CreateDependencyAndRunGo(tempDir, shouldRunGoModCommand)
		if err != nil {
			log.Debug("Received and error:", err)
		}
	}
	return dependency.publishDependencyAndPopulateTransitive(path, cachePath, tempDir, targetRepo, output, depsTidy, details, regExp)
}

func (dependency *Dependency) publishDependencyAndPopulateTransitive(path, cachePath, tempDir, targetRepo string, graphDependencies map[string]bool, depsTidy bool, details *config.ArtifactoryDetails, regExp *RegExp) error {
	// Update the mod content
	err := dependency.UpdateModContent(depsTidy, path)
	if err != nil {
		return err
	}

	// If the mod is not empty, populate transitive dependencies
	if depsTidy && dependency.PatternMatched(regExp.GetNotEmptyModRegex()) {
		dependency.setTransitiveDependencies(targetRepo, graphDependencies, details)
	}

	published, _ := global.GetGlobalVariables().GetGlobalMap()[dependency.GetId()]
	if !published {
		err = dependency.writeModToGoCache(cachePath)
		logErrorIfOccurred(err)
	}

	// Populate and publish the transitive dependencies.
	if depsTidy && dependency.GetTransitiveDependency() != nil {
		dependency.populateTransitive(cachePath, tempDir, targetRepo, details, regExp)
	}

	// Update the mod file within the cache and publish to Artifactory the dependency if needed.
	err = dependency.updateCacheAndPublishDependency(cachePath, path, targetRepo, depsTidy, regExp, details)
	if err != nil {
		return err
	}

	if depsTidy {
		// Remove from temp folder the dependency.
		err = os.RemoveAll(filepath.Dir(path))
		if errorutils.CheckError(err) != nil {
			log.Debug("Received and error:", err.Error())
		}
	}
	return nil
}

// Updates the mod in the cache and publish the package if needed.
func (dependency *Dependency) updateCacheAndPublishDependency(cachePath, path, targetRepo string, depsTidy bool, regExp *RegExp, details *config.ArtifactoryDetails) error {
	global := global.GetGlobalVariables()
	dependenciesMap := global.GetGlobalMap()
	published, _ := dependenciesMap[dependency.GetId()]
	failed := false
	if !published {
		if depsTidy {
			// Now we need to check if there are some indirect dependencies in the go.mod file:
			dependency.updateModWithoutIndirect(cachePath, path, depsTidy, regExp)
		}
		log.Debug("Writing the new mod content to cache of the dependency", dependency.GetId())
		err := dependency.writeModToGoCache(cachePath)
		logErrorIfOccurred(err)

		serviceManager, err := utils.CreateServiceManager(details, false)
		if err != nil {
			global.IncreaseFailures()
			return err
		}
		totalOutOf := fmt.Sprintf("%d/%d", global.GetSuccess() + 1, global.GetTotal())
		dependenciesMap[dependency.GetId()] = true
		// Publish the dependency
		err = dependency.Publish(totalOutOf, targetRepo, serviceManager)
		if err != nil {
			global.IncreaseFailures()
			failed = true
			if depsTidy {
				log.Debug("Received and error:", err.Error())
			} else {
				return err
			}
		}
	}
	if !failed {
		global.IncreaseSuccess()
	}
	return nil
}

// Updating the new mod content
func (dependency *Dependency) UpdateModContent(depsTidy bool, path string) error {
	if depsTidy && path != "" {
		modContent, err := ioutil.ReadFile(path)
		if err != nil {
			global.GetGlobalVariables().IncreaseFailures()
			return errorutils.CheckError(err)
		}
		dependency.SetModContent(modContent)
	}
	return nil
}

// Runs over the transitive dependencies, populate them and publish to Artifactory.
func (dependency *Dependency) populateTransitive(cachePath, tempDir, targetRepo string, details *config.ArtifactoryDetails, regExp *RegExp) {
	global.GetGlobalVariables().IncreaseTotal(len(dependency.GetTransitiveDependency()))
	for _, transitiveDep := range dependency.GetTransitiveDependency() {
		published, _ := global.GetGlobalVariables().GetGlobalMap()[transitiveDep.GetId()]
		if !published {
			log.Debug("Starting to work on transitive dependency:", transitiveDep.GetId())
			err := transitiveDep.PopulateDependenciesModAndPublish(cachePath, tempDir, targetRepo, true, !transitiveDep.PatternMatched(regExp.GetNotEmptyModRegex()), details, regExp)
			logErrorIfOccurred(err)
		} else {
			global.GetGlobalVariables().IncreaseSuccess()
			log.Debug("The dependency", transitiveDep.GetId(), "was already handled")
		}
	}
}

func (dependency *Dependency) CreateDependencyAndRunGo(tempDir string, shouldRunGoModCommand bool) (path string, output map[string]bool, err error) {
	err = os.Unsetenv(golangutil.GOPROXY)
	if err != nil {
		return
	}
	path, err = createDependencyWithMod(tempDir, *dependency)
	if err != nil {
		return
	}
	output, err = populateModAndGetDependenciesGraph(path, shouldRunGoModCommand, true)
	return
}

func (dependency *Dependency) setTransitiveDependencies(targetRepo string, graphDependencies map[string]bool, details *config.ArtifactoryDetails) {
	var dependencies []Dependency
	for transitiveDependency := range graphDependencies {
		module := strings.Split(transitiveDependency, "@")
		if len(module) == 2 {
			dependenciesMap := global.GetGlobalVariables().GetGlobalMap()
			name := getDependencyName(module[0])
			_, exists := dependenciesMap[name+":"+module[1]]
			if !exists {
				// Check if the dependency in the cache
				cachePath, err := GetCachePath()
				dep, err := createDependency(cachePath, name, module[1])
				logErrorIfOccurred(err)
				if err != nil {
					continue
				}
				// Check if this dependency exists in Artifactory.
				client := httpclient.NewDefaultHttpClient()
				downloadedFromArtifactory, err := shouldDownloadFromArtifactory(module[0], module[1], targetRepo, details, client)
				logErrorIfOccurred(err)
				if err != nil {
					continue
				}
				if dep == nil {
					// Dependency is missing within the cache!! need to download it...
					dep, err = downloadAndCreateDependency(cachePath, name, module[1], transitiveDependency, targetRepo, downloadedFromArtifactory, details)
					logErrorIfOccurred(err)
					if err != nil {
						continue
					}
				}

				if dep != nil {
					log.Debug(fmt.Sprintf("Dependency %s has transitive dependency %s", dependency.GetId(), dep.GetId()))
					dependencies = append(dependencies, *dep)
					dependenciesMap[name+":"+module[1]] = downloadedFromArtifactory
				}
			} else {
				log.Debug("Dependency", transitiveDependency, "was add previously.")
			}
		}
	}
	dependency.transitiveDependencies = dependencies
}

func (dependency *Dependency) updateModWithoutIndirect(cachePath, path string, depsTidy bool, regExp *RegExp) {
	if dependency.PatternMatched(regExp.GetIndirectRegex()) {
		// Now run again go mod tidy.
		log.Debug(fmt.Sprintf("Dependency %s has indirect dependencies. Updating mod.", path))
		_, err := populateModAndGetDependenciesGraph(path, true, false)
		logErrorIfOccurred(err)
		err = dependency.UpdateModContent(depsTidy, path)
		logErrorIfOccurred(err)
	}
}

func (dependency *Dependency) Publish(summary string, targetRepo string, servicesManager *artifactory.ArtifactoryServicesManager) error {
	message := fmt.Sprintf("Publishing: %s to %s", dependency.id, targetRepo)
	if summary != "" {
		message += ":" + summary
	}
	log.Info(message)
	params := &_go.GoParamsImpl{}
	params.ZipPath = dependency.zipPath
	params.ModContent = dependency.modContent
	params.Version = dependency.version
	params.TargetRepo = targetRepo
	params.ModuleId = dependency.id

	return servicesManager.PublishGoProject(params)
}

func (dependency *Dependency) Dependencies() []buildinfo.Dependency {
	return dependency.buildInfoDependencies
}

// Returns true if regex found a match otherwise false.
func (dependency *Dependency) PatternMatched(regExp *regexp.Regexp) bool {
	lines := strings.Split(string(dependency.modContent), "\n")
	for _, line := range lines {
		if regExp.FindString(line) != "" {
			return true
		}
	}
	return false
}

func (dependency *Dependency) writeModToGoCache(cachePath string) error {
	moduleAndVersion := strings.Split(dependency.GetId(), ":")
	pathToModule := strings.Split(moduleAndVersion[0], "/")

	path := filepath.Join(cachePath, strings.Join(pathToModule, fileutils.GetFileSeparator()), "@v", moduleAndVersion[1]+".mod")
	err := ioutil.WriteFile(path, dependency.GetModContent(), 0700)
	return errorutils.CheckError(err)
}
