package solution

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/nuget/solution/project"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

type Solution interface {
	BuildInfo() (*buildinfo.BuildInfo, error)
	Marshal() ([]byte, error)
	GetProjects() []project.Project
}

func Load(solutionPath, slnFile string) (Solution, error) {
	solution := &solution{path: solutionPath, slnFile: slnFile}
	err := solution.loadProjects()
	return solution, err
}

type solution struct {
	path     string
	slnFile  string
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
	regExp, err := utils.GetRegExp(`Project\("(.*)\nEndProject`)
	if err != nil {
		return err
	}

	allProjects, err := solution.getProjectsFromSlns(regExp)
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

// Finds all the project section within the sln files. If sln file is not provided,
// searching the entire directory for all sln files.
// Returns slice of all the projects.
func (solution *solution) getProjectsFromSlns(regExp *regexp.Regexp) ([]string, error) {
	var allProjects []string
	if solution.slnFile == "" {
		slnFiles, err := fileutils.ListFilesWithSpecificExtension(solution.path, ".sln")
		if err != nil {
			return nil, err
		}
		for _, slnFile := range slnFiles {
			projects, err := parseSlnFile(slnFile, regExp)
			if err != nil {
				return nil, err
			}
			allProjects = append(allProjects, projects...)
		}
	} else {
		projects, err := parseSlnFile(filepath.Join(solution.path, solution.slnFile), regExp)
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, projects...)
	}
	return allProjects, nil
}

// Parsing the project line for the project name and path information.
// Returns the name and path to csproj
func parseProject(projectLine, path string) (projectName, csprojPath string, err error) {
	parsedLine := strings.Split(projectLine, "=")
	if len(parsedLine) <= 1 {
		return "", "", errors.New("Wrong number of arguments for project line:" + projectLine)
	}

	projectInfo := strings.Split(parsedLine[1], ",")
	if len(projectInfo) <= 2 {
		return "", "", errors.New("Wrong number of arguments for project information: " + parsedLine[1])
	}
	projectName = removeApostrophes(projectInfo[0])
	csprojPath = filepath.Join(path, removeApostrophes(projectInfo[1]))
	return
}

// Parse the sln file and returns all the founded lines by the regex
func parseSlnFile(slnFile string, regExp *regexp.Regexp) ([]string, error) {
	content, err := ioutil.ReadFile(slnFile)
	if err != nil {
		return nil, err
	}
	projects := regExp.FindAllString(string(content), -1)
	return projects, nil
}

func removeApostrophes(value string) string {
	return strings.Trim(strings.TrimSpace(value), "\"")
}
