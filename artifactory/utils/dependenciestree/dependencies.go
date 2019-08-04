package dependenciestree

import (
	"encoding/json"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
)

// Dependency tree
type Tree interface {
	MarshalJSON() ([]byte, error)
}

type Root []*DependenciesTree

type DependenciesTree struct {
	Dependency         *buildinfo.Dependency `json:"dependencies,omitempty"`
	DirectDependencies []*DependenciesTree
	Id                 string
}

func (r Root) MarshalJSON() ([]byte, error) {
	type Alias Root
	return json.Marshal(Alias(r))
}

func (t DependenciesTree) MarshalJSON() ([]byte, error) {
	type Alias []*DependenciesTree
	return json.Marshal(&struct {
		*buildinfo.Dependency
		Alias `json:"dependencies,omitempty"`
	}{
		Dependency: t.Dependency,
		Alias:      t.DirectDependencies,
	})
}

// Add children nodes for a dependency
func (t *DependenciesTree) AddChildren(allDependencies map[string]*buildinfo.Dependency, children map[string][]string) {
	for _, child := range children[t.Id] {
		if _, ok := allDependencies[child]; !ok {
			//No such child, skip...
			continue
		}
		childTree := &DependenciesTree{Id: child, Dependency: allDependencies[child]}
		childTree.AddChildren(allDependencies, children)
		t.DirectDependencies = append(t.DirectDependencies, childTree)
	}
}

// Create dependency tree using the data received from the extractors.
func CreateDependencyTree(rootDependencies []string, allDependencies map[string]*buildinfo.Dependency, childrenMap map[string][]string) Root {
	var rootTree Root
	for _, root := range rootDependencies {
		if _, ok := allDependencies[root]; !ok {
			//No such root, skip...
			continue
		}
		subTree := &DependenciesTree{Id: root, Dependency: allDependencies[root]}
		subTree.AddChildren(allDependencies, childrenMap)
		rootTree = append(rootTree, subTree)
	}
	return rootTree
}
