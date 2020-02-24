package dependencies

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	serviceutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"strings"
)

// Populate project's dependencies with checksums and file names.
// If the dependency was downloaded in this pip-install execution, checksum will be fetched from Artifactory.
// Otherwise, check if exists in cache.
// Return dependency-names of all dependencies which its information could not be obtained.
func AddDepsInfoAndReturnMissingDeps(dependenciesMap map[string]*buildinfo.Dependency, dependenciesCache *DependenciesCache, dependencyToFileMap map[string]string, servicesManager *artifactory.ArtifactoryServicesManager, repository string) ([]string, error) {
	var missingDeps []string
	// Iterate dependencies map to update info.
	for depName := range dependenciesMap {
		// Get dependency info.
		depFileName, depChecksum, err := getDependencyInfo(depName, repository, dependenciesCache, dependencyToFileMap, servicesManager)
		if err != nil {
			return nil, err
		}

		// Check if info not found.
		if depFileName == "" || depChecksum == nil {
			// Dependency either wasn't downloaded in this run nor stored in cache.
			missingDeps = append(missingDeps, depName)

			// dependenciesMapT should contain only dependencies with checksums.
			delete(dependenciesMap, depName)

			continue
		}
		fileType := ""
		// Update dependency info.
		dependenciesMap[depName].Id = depFileName
		if i := strings.LastIndex(depFileName, "."); i != -1 {
			fileType = depFileName[i+1:]
		}
		dependenciesMap[depName].Type = fileType
		dependenciesMap[depName].Checksum = depChecksum
	}

	return missingDeps, nil
}

// Get dependency information.
// If dependency was downloaded in this pip-install execution, fetch info from Artifactory.
// Otherwise, fetch info from cache.
func getDependencyInfo(depName, repository string, dependenciesCache *DependenciesCache, dependencyToFileMap map[string]string, servicesManager *artifactory.ArtifactoryServicesManager) (string, *buildinfo.Checksum, error) {
	// Check if this dependency was updated during this pip-install execution, and we have its file-name.
	// If updated - fetch checksum from Artifactory, regardless of what was previously stored in cache.
	depFileName, ok := dependencyToFileMap[depName]
	if ok && depFileName != "" {
		checksum, err := getDependencyChecksumFromArtifactory(servicesManager, repository, depFileName)
		return depFileName, checksum, err
	}

	// Check cache for dependency checksum.
	if dependenciesCache != nil {
		dep := dependenciesCache.GetDependency(depName)
		if dep != nil {
			// Checksum found in cache, return info
			return dep.Id, dep.Checksum, nil
		}
	}

	return "", nil, nil
}

// Fetch checksum for file from Artifactory.
// If the file isn't found, or md5 or sha1 are missing, return nil.
func getDependencyChecksumFromArtifactory(servicesManager *artifactory.ArtifactoryServicesManager, repository, dependencyFile string) (*buildinfo.Checksum, error) {
	log.Debug(fmt.Sprintf("Fetching checksums for: %s", dependencyFile))
	result, err := servicesManager.Aql(serviceutils.CreateAqlQueryForPypi(repository, dependencyFile))
	if err != nil {
		return nil, err
	}

	parsedResult := new(aqlResult)
	err = json.Unmarshal(result, parsedResult)
	if err = errorutils.CheckError(err); err != nil {
		return nil, err
	}
	if len(parsedResult.Results) == 0 {
		log.Debug(fmt.Sprintf("File: %s could not be found in repository: %s", dependencyFile, repository))
		return nil, nil
	}

	// Verify checksum exist.
	sha1 := parsedResult.Results[0].Actual_sha1
	md5 := parsedResult.Results[0].Actual_md5
	if sha1 == "" || md5 == "" {
		// Missing checksum.
		log.Debug(fmt.Sprintf("Missing checksums for file: %s, sha1: '%s', md5: '%s'", dependencyFile, sha1, md5))
		return nil, nil
	}

	// Update checksum.
	checksum := &buildinfo.Checksum{Sha1: sha1, Md5: md5}
	log.Debug(fmt.Sprintf("Found checksums for file: %s, sha1: '%s', md5: '%s'", dependencyFile, sha1, md5))

	return checksum, nil
}

type aqlResult struct {
	Results []*results `json:"results,omitempty"`
}

type results struct {
	Actual_md5  string `json:"actual_md5,omitempty"`
	Actual_sha1 string `json:"actual_sha1,omitempty"`
}
