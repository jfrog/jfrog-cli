package dependencies

import (
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Dependencies extractor for setup.py
type setupExtractor struct {
	allDependencies  map[string]*buildinfo.Dependency
	childrenMap      map[string][]string
	rootDependencies []string

	setuppyFilePath      string
	pythonExecutablePath string
}

func NewSetupExtractor(fileName, projectRoot, pythonExecutablePath string) Extractor {
	// Create new extractor.
	return &setupExtractor{setuppyFilePath: filepath.Join(projectRoot, fileName), pythonExecutablePath: pythonExecutablePath}
}

func (extractor *setupExtractor) Extract() error {
	// Parse setup.py, add to rootDependencies.
	dependencies, err := extractor.getRootDependencies()
	if errorutils.CheckError(err) != nil {
		return err
	}
	extractor.rootDependencies = dependencies

	// Get installed packages tree.
	environmentPackages, err := BuildPipDependencyMap(extractor.pythonExecutablePath, nil)
	if err != nil {
		return nil
	}

	// Extract all project dependencies.
	allDeps, childMap, err := extractDependencies(extractor.rootDependencies, environmentPackages)
	if err != nil {
		return err
	}

	// Update extracted dependencies.
	extractor.allDependencies = allDeps
	extractor.childrenMap = childMap

	return nil
}

// Get dependencies from setup.py
func (extractor *setupExtractor) getRootDependencies() ([]string, error) {
	// Create temp-dir.
	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		return nil, err
	}
	defer fileutils.RemoveTempDir(tempDirPath)

	// Execute egg_info command and return requires.txt content.
	content, err := extractor.getEgginfoRequiresContent(tempDirPath)
	if err != nil {
		return nil, err
	}

	// Extract dependencies from file content.
	rootDeps, err := extractor.getRootDependenciesFromFileContent(content)
	if err != nil {
		return nil, err
	}

	// return the root dependencies
	return rootDeps, nil
}

func (extractor *setupExtractor) getRootDependenciesFromFileContent(content []byte) ([]string, error) {
	// Parse dependencies.
	depsRegexp, err := utils.GetRegExp(`(?m)^\w[\w-\.]+`)
	if err != nil {
		return nil, err
	}

	return depsRegexp.FindAllString(string(content), -1), nil
}

// Run egg-info command on setup.py, the command generates metadata files.
// Return the content of the 'requires.txt' file.
func (extractor *setupExtractor) getEgginfoRequiresContent(tempPath string) ([]byte, error) {
	// Change work-dir to temp, preserve current work-dir when method ends.
	wd, err := os.Getwd()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	defer os.Chdir(wd)
	err = os.Chdir(tempPath)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	// Run python egg_info command.
	egginfoOutput, err := extractor.executeEgginfoCommandWithOutput()
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	// Parse egg_info execution output to find requires.txt path.
	requirestxtPath, err := extractor.extractRequirestxtPathFromCommandOutput(egginfoOutput)
	if err != nil {
		return nil, err
	}

	// Read requires.txt file.
	requiresFileExists, err := fileutils.IsFileExists(requirestxtPath, false)
	if !requiresFileExists {
		return nil, errorutils.CheckError(errors.New(fmt.Sprintf("File 'requires.txt' couldn't be found in its designated location: %s", requirestxtPath)))
	}

	return ioutil.ReadFile(requirestxtPath)
}

func (extractor *setupExtractor) extractRequirestxtPathFromCommandOutput(egginfoOutput string) (string, error) {
	//(?m) means a multiline match, matching line-by-line.
	requiresRegexp, err := utils.GetRegExp(`(?m)^writing\srequirements\sto\s(\w[\w-\.]+\.egg\-info[\\\/](requires\.txt)$)`)
	if err != nil {
		return "", err
	}

	matchedOutputLines := requiresRegexp.FindAllString(egginfoOutput, -1)
	if len(matchedOutputLines) != 1 {
		return "", errorutils.CheckError(errors.New("Failed parsing egg_info command, couldn't find requires.txt location."))
	}

	// Extract path from matched line.
	matchedResults := requiresRegexp.FindStringSubmatch(matchedOutputLines[0])
	return matchedResults[1], nil
}

// Execute egg_info command for setup.py, return command's output.
func (extractor *setupExtractor) executeEgginfoCommandWithOutput() (string, error) {
	pythonEggInfoCmd := &pip.PipCmd{
		Executable:  extractor.pythonExecutablePath,
		Command:     extractor.setuppyFilePath,
		CommandArgs: []string{"egg_info"},
		EnvVars:     nil,
		StrWriter:   nil,
		ErrWriter:   nil,
	}
	return gofrogcmd.RunCmdOutput(pythonEggInfoCmd)
}

func (extractor *setupExtractor) AllDependencies() map[string]*buildinfo.Dependency {
	return extractor.allDependencies
}

func (extractor *setupExtractor) DirectDependencies() []string {
	return extractor.rootDependencies
}

func (extractor *setupExtractor) ChildrenMap() map[string][]string {
	return extractor.childrenMap
}
