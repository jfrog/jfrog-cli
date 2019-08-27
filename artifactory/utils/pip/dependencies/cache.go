package dependencies

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"io/ioutil"
	"os"
	"path/filepath"
)

type DependenciesCache map[string]*buildinfo.Dependency

// Return project's dependencies cache.
// If cache not exist -> return nil, nil.
// If error occurred, return error.
func GetProjectDependenciesCache() (*DependenciesCache, error) {
	cache := new(DependenciesCache)
	cacheFilePath, exists, err := getCacheFilePath()
	if errorutils.CheckError(err) != nil || !exists {
		return nil, err
	}
	jsonFile, err := os.Open(cacheFilePath)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	err = json.Unmarshal(byteValue, cache)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	return cache, nil
}

// Receives map of project dependencies.
// Write new project's dependencies cache with current dependencies.
func UpdateDependenciesCache(updatedDependencyMap DependenciesCache) error {
	content, err := json.Marshal(&updatedDependencyMap)
	if err != nil {
		return errorutils.CheckError(err)
	}
	cacheFilePath, _, err := getCacheFilePath()
	if err != nil {
		return errorutils.CheckError(err)
	}

	cacheFile, err := os.Create(cacheFilePath)
	if err != nil {
		return errorutils.CheckError(err)
	}
	defer cacheFile.Close()

	_, err = cacheFile.Write(content)
	if err != nil {
		return errorutils.CheckError(err)
	}

	return nil
}

// Return required dependency from cache.
// If dependency does not exist, return nil.
// dependencyName - Name of dependency (lowercase package name).
func (cache DependenciesCache) GetDependency(dependencyName string) *buildinfo.Dependency {
	dependency, ok := cache[dependencyName]
	if !ok {
		return nil
	}

	return dependency
}

func getCacheFilePath() (cacheFilePath string, exists bool, err error) {
	// Cache file should be in the same path of the pip configuration file.
	confFilePath, _, err := utils.GetProjectConfFilePath(utils.Pip)
	if errorutils.CheckError(err) != nil {
		return "", false, err
	}
	// Get the parent dir of the configuration
	cacheFilePath = filepath.Join(filepath.Dir(confFilePath), "cache.json")
	exists, err = fileutils.IsFileExists(cacheFilePath, false)
	return

}
