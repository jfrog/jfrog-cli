package solution

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/dependencies"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/solution/project"
	"github.com/jfrog/jfrog-cli/utils/ioutils"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"os"
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
	err := solution.getDependenciesSources()
	if err != nil {
		return solution, err
	}
	err = solution.loadProjects()
	return solution, err
}

type solution struct {
	path string
	// If there are more then one sln files in the directory,
	// the user must specify as arguments the sln file that should be used.
	slnFile             string
	projects            []project.Project
	dependenciesSources []string
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
		module := buildinfo.Module{Id: getModuleId(module, project.Name()), Dependencies: projectDependencies}
		modules = append(modules, module)
	}
	buildInfo.Modules = modules
	return buildInfo, nil
}

func getModuleId(customModuleID, projectName string) string {
	if customModuleID != "" {
		return customModuleID
	}
	return projectName
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
	slnProjects, err := solution.getProjectsFromSlns()
	if err != nil {
		return err
	}
	if slnProjects != nil {
		return solution.loadProjectsFromSolutionFile(slnProjects)
	}

	return solution.loadSingleProjectFromDir()
}

func (solution *solution) loadProjectsFromSolutionFile(slnProjects []string) error {
	for _, projectLine := range slnProjects {
		projectName, csprojPath, err := parseProject(projectLine, solution.path)
		if err != nil {
			log.Error(err)
			continue
		}
		if !strings.HasSuffix(csprojPath, ".csproj") {
			log.Debug(fmt.Sprintf("Skipping a project \"%s\", since it doesn't have a csproj file path.", projectName))
			continue
		}
		solution.loadSingleProject(projectName, csprojPath)
	}
	return nil
}

func (solution *solution) loadSingleProjectFromDir() error {
	csprojFiles, err := fileutils.ListFilesWithExtension(solution.path, ".csproj")
	if err != nil {
		return err
	}
	if len(csprojFiles) == 1 {
		projectName := strings.TrimSuffix(filepath.Base(csprojFiles[0]), ".csproj")
		solution.loadSingleProject(projectName, csprojFiles[0])
	}
	return nil
}

func (solution *solution) loadSingleProject(projectName, csprojPath string) {
	// First we wil find the project's dependencies source.
	// It can be located in the project's root directory or in a directory with the project name under the solution root.
	projectRootPath := filepath.Dir(csprojPath)
	projectPathPattern := projectRootPath + string(filepath.Separator)
	projectNamePattern := string(filepath.Separator) + projectName + string(filepath.Separator)
	var dependeciesSource string
	for _, source := range solution.dependenciesSources {
		if strings.Contains(source, projectPathPattern) || strings.Contains(source, projectNamePattern) {
			dependeciesSource = source
			break
		}
	}
	// If no dependencies source was found, we will skip the current project
	if len(dependeciesSource) == 0 {
		log.Debug(fmt.Sprintf("Project dependencies was not found for project: %s", projectName))
		return
	}
	proj, err := project.Load(projectName, projectRootPath, dependeciesSource)
	if err != nil {
		log.Error(err)
		return
	}
	if proj.Extractor() != nil {
		solution.projects = append(solution.projects, proj)
	}
	return
}

// Finds all the projects by reading the content of the sln files.
// Returns a slice with all the projects in the solution.
func (solution *solution) getProjectsFromSlns() ([]string, error) {
	var allProjects []string
	slnFiles, err := solution.getSlnFiles()
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
	return allProjects, nil
}

// If sln file is not provided, finds all sln files in the directory.
func (solution *solution) getSlnFiles() (slnFiles []string, err error) {
	if solution.slnFile != "" {
		slnFiles = append(slnFiles, filepath.Join(solution.path, solution.slnFile))
	} else {
		slnFiles, err = fileutils.ListFilesWithExtension(solution.path, ".sln")
	}
	return
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
	// In case we are running on a non-Windows OS, the solution root path and the relative path to csproj file might used different path separators.
	// We want to make sure we will get a valid path after we join both parts, so we will replace the csproj separators.
	if utils.IsWindows() {
		projectInfo[1] = ioutils.UnixToWinPathSeparator(projectInfo[1])
	} else {
		projectInfo[1] = ioutils.WinToUnixPathSeparator(projectInfo[1])
	}
	csprojPath = filepath.Join(path, filepath.FromSlash(removeQuotes(projectInfo[1])))
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

// We'll walk through the file system to find all potential dependencies sources: packages.config and project.assets.json files
func (solution *solution) getDependenciesSources() error {
	err := fileutils.Walk(solution.path, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, dependencies.PackagesFileName) || strings.HasSuffix(path, dependencies.AssetFileName) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			solution.dependenciesSources = append(solution.dependenciesSources, absPath)
		}
		return nil
	}, true)

	return errorutils.CheckError(err)
}
