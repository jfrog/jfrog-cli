package solution

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/nuget/solution/project"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"os"
	"path/filepath"
	"strings"
)

type Solution interface {
	BuildInfo() (*buildinfo.BuildInfo, error)
	Marshal() ([]byte, error)
	GetProjects() []project.Project
}

func Load(solutionPath string) (Solution, error) {
	solution := &solution{path: solutionPath}
	err := solution.loadProjects()
	return solution, err
}

type solution struct {
	path     string
	projects []project.Project
}

func (solution *solution) BuildInfo() (*buildinfo.BuildInfo, error) {
	buildInfo := &buildinfo.BuildInfo{}
	var modules []buildinfo.Module
	for _, project := range solution.projects {
		// Get All project dependencies
		dependencies, err := project.Extractor().AllDependencies()
		if err != nil {
			return nil, err
		}
		var projectDependencies []buildinfo.Dependency

		for _, dep := range dependencies {
			projectDependencies = append(projectDependencies, *dep)
		}

		module := buildinfo.Module{Id: project.Name(), Dependencies: projectDependencies}
		modules = append(modules, module)
	}
	buildInfo.Modules = modules
	return buildInfo, nil
}

func (solution *solution) Marshal() ([]byte, error) {
	return json.Marshal(&struct {
		Projects []project.Project `json:"projects,omitempty"`
	}{
		Projects: solution.projects,
	})
}

func (solution *solution) GetProjects() []project.Project {
	return solution.projects
}

func (solution *solution) loadProjects() error {
	return filepath.Walk(solution.path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return errorutils.CheckError(err)
		}
		if f.IsDir() {
			return nil
		}
		if filepath.Ext(f.Name()) == ".csproj" {
			projectName := strings.TrimSuffix(f.Name(), ".csproj")
			csprojPath, err := filepath.Rel(solution.path, path)
			if err != nil {
				return errorutils.CheckError(err)
			}
			proj, err := project.Load(projectName, filepath.Dir(csprojPath), csprojPath)
			if err != nil {
				return err
			}
			if proj.Extractor() != nil {
				solution.projects = append(solution.projects, proj)
			}
		}
		return nil
	})
}
