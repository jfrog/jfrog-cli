package dependencies

import (
	"fmt"
	golangutil "github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
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

// Represents go dependency when running with deps-tidy set to true.
type PackageWithDeps struct {
	Dependency             *Package
	transitiveDependencies []PackageWithDeps
	regExp                 *RegExp
	runGoModCommand        bool
	cachePath              string
}

// Creates a new dependency
func (pwd *PackageWithDeps) New(cachePath string, dependency Package) GoPackage {
	pwd.Dependency = &dependency
	pwd.cachePath = cachePath
	pwd.transitiveDependencies = nil
	return pwd
}

// Populate the mod file and publish the dependency and it's transitive dependencies to Artifactory
func (pwd *PackageWithDeps) PopulateModAndPublish(targetRepo string, cache *golangutil.DependenciesCache, details *config.ArtifactoryDetails) error {
	var path string
	var err error
	log.Debug("Starting to work on", pwd.Dependency.GetId())
	dependenciesMap := cache.GetMap()
	published, _ := dependenciesMap[pwd.Dependency.GetId()]
	if published {
		log.Debug("Overwriting the mod file in the cache from the one from Artifactory", pwd.Dependency.GetId())
		moduleAndVersion := strings.Split(pwd.Dependency.GetId(), ":")
		path = downloadModFileFromArtifactoryToLocalCache(pwd.cachePath, targetRepo, moduleAndVersion[0], moduleAndVersion[1], details, httpclient.NewDefaultHttpClient())
		err = pwd.updateModContent(path, cache)
		logError(err)
		pwd.runGoModCommand = false
	} else {
		// Checks if mod is empty, need to run go mod tidy command to populate the mod file.
		pwd.runGoModCommand = !pwd.patternMatched(pwd.regExp.GetNotEmptyModRegex())
		log.Debug(fmt.Sprintf("Dependency %s mod file is empty: %t", pwd.Dependency.GetId(), pwd.runGoModCommand))
	}

	// Creates the dependency in the temp folder and runs go commands: go mod tidy and go mod graph.
	// Returns the path to the project in the temp and the a map with the project dependencies
	path, output, err := pwd.createDependencyAndRunGo()
	logError(err)
	return pwd.publishDependencyAndPopulateTransitive(path, targetRepo, output, cache, details)
}

// Updating the new mod content
func (pwd *PackageWithDeps) updateModContent(path string, cache *golangutil.DependenciesCache) error {
	if path != "" {
		modContent, err := ioutil.ReadFile(path)
		if err != nil {
			cache.IncrementFailures()
			return errorutils.CheckError(err)
		}
		pwd.Dependency.SetModContent(modContent)
	}
	return nil
}

// Init the dependency information if needed.
func (pwd *PackageWithDeps) Init() error {
	var err error
	pwd.regExp, err = GetRegex()
	if err != nil {
		return err
	}
	return nil
}

// Returns true if regex found a match otherwise false.
func (pwd *PackageWithDeps) patternMatched(regExp *regexp.Regexp) bool {
	lines := strings.Split(string(pwd.Dependency.modContent), "\n")
	for _, line := range lines {
		if regExp.FindString(line) != "" {
			return true
		}
	}
	return false
}

// Creates the dependency in the temp folder and runs go mod tidy and go mod graph
// Returns the path to the project in the temp and the a map with the project dependencies
func (pwd *PackageWithDeps) createDependencyAndRunGo() (path string, output map[string]bool, err error) {
	err = os.Unsetenv(golangutil.GOPROXY)
	if err != nil {
		return
	}
	path, err = createDependencyWithMod(*pwd.Dependency)
	if err != nil {
		return
	}
	output, err = populateModAndGetDependenciesGraph(path, pwd.runGoModCommand, true)
	return
}

func (pwd *PackageWithDeps) publishDependencyAndPopulateTransitive(path, targetRepo string, graphDependencies map[string]bool, cache *golangutil.DependenciesCache, details *config.ArtifactoryDetails) error {
	// Update the mod content
	err := pwd.updateModContent(path, cache)
	logError(err)

	// If the mod is not empty, populate transitive dependencies
	if pwd.patternMatched(pwd.regExp.GetNotEmptyModRegex()) {
		pwd.setTransitiveDependencies(targetRepo, graphDependencies, cache, details)
	}

	published, _ := cache.GetMap()[pwd.Dependency.GetId()]
	if !published && pwd.patternMatched(pwd.regExp.GetNotEmptyModRegex()) {
		err = pwd.writeModContentToGoCache()
		logError(err)
	}

	// Populate and publish the transitive dependencies.
	if pwd.transitiveDependencies != nil {
		pwd.populateTransitive(targetRepo, cache, details)
	}

	// Update the mod file within the cache and publish to Artifactory the dependency if needed.
	err = pwd.updateCacheAndPublishDependency(path, targetRepo, cache, details)
	logError(err)

	// Remove from temp folder the dependency.
	err = os.RemoveAll(filepath.Dir(path))
	if errorutils.CheckError(err) != nil {
		log.Error("Received an error:", err.Error())
	}

	return nil
}

// Prepare for publishing and publish the dependency to Artifactory
// Mark this dependency as published
func (pwd *PackageWithDeps) prepareAndPublish(targetRepo string, cache *golangutil.DependenciesCache, details *config.ArtifactoryDetails) error {
	err := pwd.Dependency.prepareAndPublish(targetRepo, cache, details)
	cache.GetMap()[pwd.Dependency.GetId()] = true
	return err
}

// Updates the mod in the cache and publish the package if needed.
func (pwd *PackageWithDeps) updateCacheAndPublishDependency(path, targetRepo string, cache *golangutil.DependenciesCache, details *config.ArtifactoryDetails) error {
	dependenciesMap := cache.GetMap()
	published, _ := dependenciesMap[pwd.Dependency.GetId()]
	if !published {
		// Now we need to check if there are some indirect dependencies in the go.mod file:
		pwd.updateModWithoutIndirect(path, cache)
		log.Debug("Writing the new mod content to cache of the dependency", pwd.Dependency.GetId())
		err := pwd.writeModContentToGoCache()
		logError(err)
		return pwd.prepareAndPublish(targetRepo, cache, details)
	}
	return nil
}

func (pwd *PackageWithDeps) updateModWithoutIndirect(path string, cache *golangutil.DependenciesCache) {
	if pwd.patternMatched(pwd.regExp.GetIndirectRegex()) {
		// Now run again go mod tidy.
		log.Debug(fmt.Sprintf("Dependency %s has indirect dependencies. Updating mod.", path))
		_, err := populateModAndGetDependenciesGraph(path, true, false)
		logError(err)
		err = pwd.updateModContent(path, cache)
		logError(err)
	}
}

func (pwd *PackageWithDeps) setTransitiveDependencies(targetRepo string, graphDependencies map[string]bool, cache *golangutil.DependenciesCache, details *config.ArtifactoryDetails) {
	var dependencies []PackageWithDeps
	for transitiveDependency := range graphDependencies {
		module := strings.Split(transitiveDependency, "@")
		if len(module) == 2 {
			dependenciesMap := cache.GetMap()
			name := getDependencyName(module[0])
			_, exists := dependenciesMap[name+":"+module[1]]
			if !exists {
				// Check if the dependency in the cache
				dep, err := createDependency(pwd.cachePath, name, module[1])
				logError(err)
				if err != nil {
					continue
				}
				// Check if this dependency exists in Artifactory.
				client := httpclient.NewDefaultHttpClient()
				downloadedFromArtifactory, err := shouldDownloadFromArtifactory(module[0], module[1], targetRepo, details, client)
				logError(err)
				if err != nil {
					continue
				}
				if dep == nil {
					// Dependency is missing within the cache. Need to download it...
					dep, err = downloadAndCreateDependency(pwd.cachePath, name, module[1], transitiveDependency, targetRepo, downloadedFromArtifactory, details)
					logError(err)
					if err != nil {
						continue
					}
				}

				if dep != nil {
					log.Debug(fmt.Sprintf("Dependency %s has transitive dependency %s", pwd.Dependency.GetId(), dep.GetId()))
					depsWithTrans := &PackageWithDeps{Dependency: dep,
						regExp:    pwd.regExp,
						cachePath: pwd.cachePath}
					dependencies = append(dependencies, *depsWithTrans)
					dependenciesMap[name+":"+module[1]] = downloadedFromArtifactory
				}
			} else {
				log.Debug("Dependency", transitiveDependency, "was add previously.")
			}
		}
	}
	pwd.transitiveDependencies = dependencies
}

func (pwd *PackageWithDeps) writeModContentToGoCache() error {
	moduleAndVersion := strings.Split(pwd.Dependency.GetId(), ":")
	pathToModule := strings.Split(moduleAndVersion[0], "/")
	path := filepath.Join(pwd.cachePath, strings.Join(pathToModule, fileutils.GetFileSeparator()), "@v", moduleAndVersion[1]+".mod")
	err := ioutil.WriteFile(path, pwd.Dependency.GetModContent(), 0700)
	return errorutils.CheckError(err)
}

// Runs over the transitive dependencies, populate the mod files of those transitive dependencies
func (pwd *PackageWithDeps) populateTransitive(targetRepo string, cache *golangutil.DependenciesCache, details *config.ArtifactoryDetails) {
	cache.IncrementTotal(len(pwd.transitiveDependencies))
	for _, transitiveDep := range pwd.transitiveDependencies {
		published, _ := cache.GetMap()[transitiveDep.Dependency.GetId()]
		if !published {
			log.Debug("Starting to work on transitive dependency:", transitiveDep.Dependency.GetId())
			err := transitiveDep.PopulateModAndPublish(targetRepo, cache, details)
			logError(err)
		} else {
			cache.IncrementSuccess()
			log.Debug("The dependency", transitiveDep.Dependency.GetId(), "was already handled")
		}
	}
}
