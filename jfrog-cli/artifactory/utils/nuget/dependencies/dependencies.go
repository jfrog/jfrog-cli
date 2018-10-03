package dependencies

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
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

// Dependency tree
type Tree interface {
	MarshalJSON() ([]byte, error)
}

func CreateCompatibleExtractor(projectName, projectRoot string) (Extractor, error) {
	extractor, err := getCompatibleExtractor(projectName, projectRoot)
	if err != nil {
		return nil, err
	}
	return extractor, nil
}

func CreateDependencyTree(extractor Extractor) (root, error) {
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
	return createDependencyTree(rootDependencies, allDependencies, childrenMap), nil
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
	return nil, errorutils.CheckError(fmt.Errorf("Unsupported project dependencies for project: %s", projectName))
}

type root []*tree

type tree struct {
	Dependency         *buildinfo.Dependency `json:"dependencies,omitempty"`
	DirectDependencies []*tree
	id                 string
}

func (r root) MarshalJSON() ([]byte, error) {
	type Alias root
	return json.Marshal(Alias(r))
}

func (t tree) MarshalJSON() ([]byte, error) {
	type Alias []*tree
	return json.Marshal(&struct {
		*buildinfo.Dependency
		Alias `json:"dependencies,omitempty"`
	}{
		Dependency: t.Dependency,
		Alias:      t.DirectDependencies,
	})
}

// Create dependency tree using the data received from the extractors.
func createDependencyTree(rootDependencies []string, allDependencies map[string]*buildinfo.Dependency, childrenMap map[string][]string) root {
	var rootTree root
	for _, root := range rootDependencies {
		if _, ok := allDependencies[root]; !ok {
			//No such root, skip...
			continue
		}
		subTree := &tree{id: root, Dependency: allDependencies[root]}
		subTree.addChildren(allDependencies, childrenMap)
		rootTree = append(rootTree, subTree)
	}
	return rootTree
}

// Add children nodes for a dependency
func (t *tree) addChildren(allDependencies map[string]*buildinfo.Dependency, children map[string][]string) {
	for _, child := range children[t.id] {
		if _, ok := allDependencies[child]; !ok {
			//No such child, skip...
			continue
		}
		childTree := &tree{id: child, Dependency: allDependencies[child]}
		childTree.addChildren(allDependencies, children)
		t.DirectDependencies = append(t.DirectDependencies, childTree)
	}
}
