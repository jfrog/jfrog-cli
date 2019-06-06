package solution

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/nuget/solution/project"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

type Solution interface {
	BuildInfo(module string) (*buildinfo.BuildInfo, error)
	Marshal() ([]byte, error)
	GetProjects() []project.Project
}

var projectRegExp *regexp.Regexp

func Load(solutionPath, slnFile string) (Solution, error) {
	solution := &solution{path: solutionPath, slnFile: slnFile}
	err := solution.loadProjects()
	return solution, err
}

type solution struct {
	path string
	// If there are more then one sln files in the directory,
	// the user must specify as arguments the sln file that should be used.
	slnFile  string
	projects []project.Project
}

func (solution *solution) BuildInfo(module string) (*buildinfo.BuildInfo, error) {
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
		if module == "" {
			module = project.Name()
		}
		module := buildinfo.Module{Id: module, Dependencies: projectDependencies}
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
	allProjects, err := solution.getProjectsFromSlns()
	if err != nil {
		return err
	}

	for _, projectLine := range allProjects {
		projectName, csprojPath, err := parseProject(projectLine, solution.path)
		if err != nil {
			log.Error(err)
			continue
		}
		proj, err := project.Load(projectName, filepath.Dir(csprojPath), csprojPath)
		if err != nil {
			log.Error(err)
			continue
		}
		if proj.Extractor() != nil {
			solution.projects = append(solution.projects, proj)
		}
	}
	return nil
}

// Finds all the projects by reading the content of the the sln files. If sln file is not provided,
// finds all sln files in the directory.
// Returns a slice with all the projects in the solution.
func (solution *solution) getProjectsFromSlns() ([]string, error) {
	var allProjects []string
	if solution.slnFile == "" {
		slnFiles, err := fileutils.ListFilesWithExtension(solution.path, ".sln")
		if err != nil {
			return nil, err
		}
		for _, slnFile := range slnFiles {
			projects, err := parseSlnFile(slnFile)
			if err != nil {
				return nil, err
			}
			allProjects = append(allProjects, projects...)
		}
	} else {
		projects, err := parseSlnFile(filepath.Join(solution.path, solution.slnFile))
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, projects...)
	}
	return allProjects, nil
}

// Parses the project line for the project name and path information.
// Returns the name and path to csproj
func parseProject(projectLine, path string) (projectName, csprojPath string, err error) {
	parsedLine := strings.Split(projectLine, "=")
	if len(parsedLine) <= 1 {
		return "", "", errors.New("Unexpected project line format: " + projectLine)
	}

	projectInfo := strings.Split(parsedLine[1], ",")
	if len(projectInfo) <= 2 {
		return "", "", errors.New("Unexpected project information format: " + parsedLine[1])
	}
	projectName = removeQuotes(projectInfo[0])
	csprojPath = filepath.Join(path, removeQuotes(projectInfo[1]))
	return
}

// Parse the sln file according to project regular expression and returns all the founded lines by the regex
func parseSlnFile(slnFile string) ([]string, error) {
	var err error
	if projectRegExp == nil {
		projectRegExp, err = utils.GetRegExp(`Project\("(.*)\nEndProject`)
		if err != nil {
			return nil, err
		}
	}

	content, err := ioutil.ReadFile(slnFile)
	if err != nil {
		return nil, err
	}
	projects := projectRegExp.FindAllString(string(content), -1)
	return projects, nil
}

func removeQuotes(value string) string {
	return strings.Trim(strings.TrimSpace(value), "\"")
}
