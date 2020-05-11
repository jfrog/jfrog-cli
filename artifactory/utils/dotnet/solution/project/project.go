package project

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dependenciestree"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/dependencies"
)

type Project interface {
	Name() string
	MarshalJSON() ([]byte, error)
	Extractor() dependencies.Extractor
	CreateDependencyTree() error
}

func Load(name, rootPath, dependeciesSource string) (Project, error) {
	var err error
	project := &project{name: name, rootPath: rootPath, dependenciesSource: dependeciesSource}
	project.extractor, err = project.getCompatibleExtractor()
	return project, err
}

func (project *project) getCompatibleExtractor() (dependencies.Extractor, error) {
	extractor, err := dependencies.CreateCompatibleExtractor(project.name, project.dependenciesSource)
	return extractor, err
}

func (project *project) CreateDependencyTree() error {
	var err error
	project.dependencyTree, err = dependencies.CreateDependencyTree(project.extractor)
	return err
}

type project struct {
	name               string
	rootPath           string
	dependenciesSource string
	dependencyTree     dependenciestree.Tree
	extractor          dependencies.Extractor
}

func (project *project) Name() string {
	return project.name
}

func (project *project) Extractor() dependencies.Extractor {
	return project.extractor
}

func (project *project) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name         string                `json:"name,omitempty"`
		Dependencies dependenciestree.Tree `json:"dependencies,omitempty"`
	}{
		Name:         project.name,
		Dependencies: project.dependencyTree,
	})
}
