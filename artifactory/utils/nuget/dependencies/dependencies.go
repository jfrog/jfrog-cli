package dependencies

import (
	"fmt"
	deptree "github.com/jfrog/jfrog-cli-go/artifactory/utils/dependenciestree"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

var extractors []Extractor

// Register dependency extractor
func register(dependencyType Extractor) {
	extractors = append(extractors, dependencyType)
}

// The extractor responsible to calculate the project dependencies.
type Extractor interface {
	// Check whether the extractor is compatible with the current dependency resolution method
	IsCompatible(projectName, projectRoot string) (bool, error)
	// Get all the dependencies for the project
	AllDependencies() (map[string]*buildinfo.Dependency, error)
	// Get all the root dependencies of the project
	DirectDependencies() ([]string, error)
	// Dependencies relations map
	ChildrenMap() (map[string][]string, error)

	new(projectName, projectRoot string) (Extractor, error)
}

func CreateCompatibleExtractor(projectName, projectRoot string) (Extractor, error) {
	extractor, err := getCompatibleExtractor(projectName, projectRoot)
	if err != nil {
		return nil, err
	}
	return extractor, nil
}

func CreateDependencyTree(extractor Extractor) (deptree.Root, error) {
	rootDependencies, err := extractor.DirectDependencies()
	if err != nil {
		return nil, err
	}
	allDependencies, err := extractor.AllDependencies()
	if err != nil {
		return nil, err
	}
	childrenMap, err := extractor.ChildrenMap()
	if err != nil {
		return nil, err
	}
	return deptree.CreateDependencyTree(rootDependencies, allDependencies, childrenMap), nil
}

// Find suitable registered dependencies extractor.
func getCompatibleExtractor(projectName, projectRoot string) (Extractor, error) {
	for _, extractor := range extractors {
		compatible, err := extractor.IsCompatible(projectName, projectRoot)
		if err != nil {
			return nil, err
		}
		if compatible {
			return extractor.new(projectName, projectRoot)
		}
	}
	log.Debug(fmt.Sprintf("Unsupported project dependencies for project: %s", projectName))
	return nil, nil
}
