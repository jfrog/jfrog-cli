package dependencies

import (
	"errors"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
)

// Receives map of project dependencies.
// Update project's dependencies cache with new dependencies.
// Remove cached dependencies which are no longer required.
func UpdateDependenciesCache(updatedDependencyMap map[string]*buildinfo.Dependency) error {
	panic(errors.New("Implement me!"))
}

// Return required dependency's checksum from cache.
// If checksum does not exist, return nil.
// dependencyName - Name of dependency (lowercase package name).
// dependencyFileName - File name as stored in Artifactory.
func GetDependencyChecksum(dependencyName, dependencyFileName string) (*buildinfo.Checksum, error) {
	panic(errors.New("Implement me!"))
}