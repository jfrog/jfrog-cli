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
	"strings"
	"sync"
)

// Dependencies extractor for setup.py
type setupExtractor struct {
	allDependencies      map[string]*buildinfo.Dependency
	childrenMap          map[string][]string
	rootDependencies     []string
	setuppyFilePath      string
	pythonExecutablePath string
	Pkg                  string
	once                 sync.Once
}

func NewSetupExtractor(fileName, projectRoot, pythonExecutablePath string) Extractor {
	// Create new extractor.
	return &setupExtractor{setuppyFilePath: filepath.Join(projectRoot, fileName), pythonExecutablePath: pythonExecutablePath}
}

func (extractor *setupExtractor) Extract() error {
	// Get installed packages tree.
	environmentPackages, err := BuildPipDependencyMap(extractor.pythonExecutablePath, nil)
	if err != nil {
		return nil
	}

	// Populate rootDependencies.
	if err := extractor.extractRootDependencies(environmentPackages); err != nil {
		return err
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

func (extractor *setupExtractor) PackageName() (string, error) {
	var err error
	extractor.once.Do(func() {
		extractor.Pkg, err = getProjectName(extractor.pythonExecutablePath, extractor.setuppyFilePath)
	})
	return extractor.Pkg, err
}

func (extractor *setupExtractor) extractRootDependencies(envDeps map[string]pipDependencyPackage) error {
	// Get package name.
	pkgName, err := extractor.PackageName()
	if err != nil {
		return err
	}

	// Get installed package from environment-dependencies map.
	pipDepPkg, ok := envDeps[strings.ToLower(pkgName)]
	if !ok {
		// Package not installed.
		return errorutils.CheckError(errors.New(fmt.Sprintf("Failed receiving root-dependencies for installed package: %s", pkgName)))
	}

	// Extract package's root-dependencies.
	extractor.rootDependencies = pipDepPkg.getDependencies()

	return nil
}

// Get dependencies from setup.py
func getProjectName(pythonExecutablePath, setuppyFilePath string) (string, error) {
	// Create temp-dir.
	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		return "", err
	}
	defer fileutils.RemoveTempDir(tempDirPath)

	// Execute egg_info command and return PKG-INFO content.
	content, err := getEgginfoPkginfoContent(tempDirPath, pythonExecutablePath, setuppyFilePath)
	if err != nil {
		return "", err
	}

	// Extract project name from file content.
	return getProjectNameFromFileContent(content)
}

// Get package-name from PKG-INFO file content.
// If pattern of package-name not found, return an error.
func getProjectNameFromFileContent(content []byte) (string, error) {
	// Create package-name regexp.
	packageNameRegexp, err := utils.GetRegExp(`(?m)^Name\:\s(\w[\w-\.]+)`)
	if err != nil {
		return "", err
	}

	// Find first match of packageNameRegexp.
	match := packageNameRegexp.FindStringSubmatch(string(content))
	if len(match) < 2 {
		return "", errorutils.CheckError(errors.New("Failed extracting package name from content."))
	}

	return match[1], nil
}

// Run egg-info command on setup.py, the command generates metadata files.
// Return the content of the 'PKG-INFO' file.
func getEgginfoPkginfoContent(tempPath, pythonExecutablePath, setuppyFilePath string) ([]byte, error) {
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
	egginfoOutput, err := executeEgginfoCommandWithOutput(pythonExecutablePath, setuppyFilePath)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	// Parse egg_info execution output to find PKG-INFO path.
	pkginfoPath, err := extractPkginfoPathFromCommandOutput(egginfoOutput)
	if err != nil {
		return nil, err
	}

	// Read PKG-INFO file.
	pkginfoFileExists, err := fileutils.IsFileExists(pkginfoPath, false)
	if !pkginfoFileExists {
		return nil, errorutils.CheckError(errors.New(fmt.Sprintf("File 'PKG-INFO' couldn't be found in its designated location: %s", pkginfoPath)))
	}

	return ioutil.ReadFile(pkginfoPath)
}

func extractPkginfoPathFromCommandOutput(egginfoOutput string) (string, error) {
	//(?m) means a multiline match, matching line-by-line.
	pkginfoRegexp, err := utils.GetRegExp(`(?m)^writing\s(\w[\w-\.]+\.egg\-info[\\\/](PKG-INFO)$)`)
	if err != nil {
		return "", err
	}

	matchedOutputLines := pkginfoRegexp.FindAllString(egginfoOutput, -1)
	if len(matchedOutputLines) != 1 {
		return "", errorutils.CheckError(errors.New("Failed parsing egg_info command, couldn't find PKG-INFO location."))
	}

	// Extract path from matched line.
	matchedResults := pkginfoRegexp.FindStringSubmatch(matchedOutputLines[0])
	return matchedResults[1], nil
}

// Execute egg_info command for setup.py, return command's output.
func executeEgginfoCommandWithOutput(pythonExecutablePath, setuppyFilePath string) (string, error) {
	pythonEggInfoCmd := &pip.PipCmd{
		Executable:  pythonExecutablePath,
		Command:     setuppyFilePath,
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
