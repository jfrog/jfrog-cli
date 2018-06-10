package project

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/nuget/dependencies"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
)

type Project interface {
	Name() string
	Dependencies() []buildinfo.Dependency
	MarshalJSON() ([]byte, error)
}

func Load(name, rootPath, csprojPath string) (Project, error) {
	project := &project{name: name, rootPath: rootPath, csprojPath: csprojPath}
	err := project.extractDependencies()
	return project, err
}

func (project *project) extractDependencies() error {
	var err error
	project.dependencyTree, err = dependencies.CreateDependencyTree(project.name, project.rootPath)
	return err
}

type project struct {
	name           string
	rootPath       string
	csprojPath     string
	dependencyTree dependencies.Tree
}

func (project *project) Name() string {
	return project.name
}

func (project *project) Dependencies() []buildinfo.Dependency {
	return project.dependencyTree.AllDependencies()
}

func (project *project) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name         string            `json:"name,omitempty"`
		Dependencies dependencies.Tree `json:"dependencies,omitempty"`
	}{
		Name:         project.name,
		Dependencies: project.dependencyTree,
	})
}
