package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/pip"
	piputils "github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
	"github.com/jfrog/jfrog-cli-go/inttestutils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type PipCmd struct {
	Command string
	Options []string
}

func TestPipInstall(t *testing.T) {
	// Init pip.
	initPipTest(t)

	// Init CLI without credential flags.
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", "")

	// Add virtual-environment path to 'PATH' for executing all pip and python commands inside the virtual-environment.
	pathValue := setPathEnvForPipInstall(t)
	if t.Failed() {
		t.FailNow()
	}
	defer os.Setenv("PATH", pathValue)

	// Check pip env is clean.
	validateEmptyPipEnv(t)

	// Populate cli config with 'default' server.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(cliutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	// Create test cases.
	allTests := []struct {
		name                 string
		project              string
		outputFolder         string
		moduleId             string
		args                 []string
		expectedDependencies int
		cleanAfterExecution  bool
	}{
		{"setuppy", "setuppyproject", "setuppy", "jfrog-python-example", []string{"pip-install", ".", "--no-cache-dir", "--force-reinstall", "--build-name=" + tests.PipBuildName}, 3, true},
		{"setuppy-verbose", "setuppyproject", "setuppy-verbose", "jfrog-python-example", []string{"pip-install", ".", "--no-cache-dir", "--force-reinstall", "-v", "--build-name=" + tests.PipBuildName}, 3, true},
		{"setuppy-with-module", "setuppyproject", "setuppy-with-module", "setuppy-with-module", []string{"pip-install", ".", "--no-cache-dir", "--force-reinstall", "--build-name=" + tests.PipBuildName, "--module=setuppy-with-module"}, 3, true},
		{"requirements", "requirementsproject", "requirements", tests.PipBuildName, []string{"pip-install", "-r", "requirements.txt", "--no-cache-dir", "--force-reinstall", "--build-name=" + tests.PipBuildName}, 5, true},
		{"requirements-verbose", "requirementsproject", "requirements-verbose", tests.PipBuildName, []string{"pip-install", "-r", "requirements.txt", "--no-cache-dir", "--force-reinstall", "-v", "--build-name=" + tests.PipBuildName}, 5, false},
		{"requirements-use-cache", "requirementsproject", "requirements-verbose", "requirements-verbose-use-cache", []string{"pip-install", "-r", "requirements.txt", "--module=requirements-verbose-use-cache", "--build-name=" + tests.PipBuildName}, 5, true},
	}

	// Run test cases.
	for buildNumber, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			testPipCmd(t, test.name, createPipProject(t, test.outputFolder, test.project), strconv.Itoa(buildNumber), test.moduleId, test.expectedDependencies, test.args)
			if test.cleanAfterExecution {
				// cleanup
				inttestutils.DeleteBuild(artifactoryDetails.Url, tests.PipBuildName, artHttpDetails)
				cleanPipTest(t, test.name)
			}
		})
	}
	cleanPipTest(t, "cleanup")
	tests.CleanFileSystem()
}

func testPipCmd(t *testing.T, outputFolder, projectPath, buildNumber, module string, expectedDependencies int, args []string) {
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	err = os.Chdir(projectPath)
	if err != nil {
		t.Error(err)
	}
	defer os.Chdir(wd)

	args = append(args, "--build-number="+buildNumber)

	err = artifactoryCli.Exec(args...)
	if err != nil {
		t.Errorf("Failed executing pip-install command: %s", err.Error())
		cleanPipTest(t, outputFolder)
		return
	}

	artifactoryCli.Exec("bp", tests.PipBuildName, buildNumber)

	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.PipBuildName, buildNumber, t, artHttpDetails)
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		t.Error("Pip build info was not generated correctly, no modules were created.")
	}

	if expectedDependencies != len(buildInfo.Modules[0].Dependencies) {
		t.Error("Incorrect number of artifacts found in the build-info, expected:", expectedDependencies, " Found:", len(buildInfo.Modules[0].Dependencies))
	}

	if module != buildInfo.Modules[0].Id {
		t.Error(fmt.Errorf("Expected module name %s, got %s", module, buildInfo.Modules[0].Id))
	}
}

func cleanPipTest(t *testing.T, outFolder string) {
	// Clean pip environment from installed packages.
	pipFreezeCmd := &PipCmd{Command: "freeze", Options: []string{"--local"}}
	out, err := gofrogcmd.RunCmdOutput(pipFreezeCmd)
	if err != nil {
		t.Fatal(err)
	}

	// If no packages to uninstall, return.
	if out == "" {
		return
	}

	// Save freeze output to file.
	freezeTarget, err := fileutils.CreateFilePath(tests.Temp, outFolder+"-freeze.txt")
	if err != nil {
		t.Error(err)
	}
	file, err := os.Create(freezeTarget)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	_, err = file.Write([]byte(out))
	if err != nil {
		t.Error(err)
	}

	// Delete freezed packages.
	pipUninstallCmd := &PipCmd{Command: "uninstall", Options: []string{"-y", "-r", freezeTarget}}
	err = gofrogcmd.RunCmd(pipUninstallCmd)
	if err != nil {
		t.Fatal(err)
	}
}

func createPipProject(t *testing.T, outFolder, projectName string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "pip", projectName)
	projectTarget := filepath.Join(tests.Out, outFolder+"-"+projectName)
	err := fileutils.CreateDirIfNotExist(projectTarget)
	if err != nil {
		t.Error(err)
	}

	// Copy pip-installation file.
	err = fileutils.CopyDir(projectSrc, projectTarget, true)
	if err != nil {
		t.Error(err)
	}

	// Copy pip-config file.
	configSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "pip", "pip.yaml")
	configTarget := filepath.Join(projectTarget, ".jfrog", "projects")
	tests.ReplaceTemplateVariables(configSrc, configTarget)

	return projectTarget
}

func initPipTest(t *testing.T) {
	if !*tests.TestPip {
		t.Skip("Skipping Pip test. To run Pip test add the '-test.pip=true' option.")
	}

	if !isRepoExist(tests.PypiRemoteRepo) {
		t.Error("Pypi test remote repository doesn't exist.")
	}

	if !isRepoExist(tests.PypiVirtualRepo) {
		t.Error("Pypi test virtual repository doesn't exist.")
	}

	if t.Failed() {
		t.FailNow()
	}
}

func setPathEnvForPipInstall(t *testing.T) string {
	// Keep original value of 'PATH'.
	pathValue, exists := os.LookupEnv("PATH")
	if !exists {
		t.Fatal("Couldn't find PATH variable, failing pip tests.")
	}

	// Append the path.
	virtualEnvPath := *tests.PipVirtualEnv
	if virtualEnvPath != "" {
		var newPathValue string
		if cliutils.IsWindows() {
			newPathValue = fmt.Sprintf("%s;%s", virtualEnvPath, pathValue)
		} else {
			newPathValue = fmt.Sprintf("%s:%s", virtualEnvPath, pathValue)
		}
		err := os.Setenv("PATH", newPathValue)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Return original PATH value.
	return pathValue
}

// Ensure that the provided pip virtual-environment is empty from installed packages.
func validateEmptyPipEnv(t *testing.T) {
	//pipFreezeCmd := &PipFreezeCmd{Executable: "pip", Command: "freeze"}
	pipFreezeCmd := &PipCmd{Command: "freeze", Options: []string{"--local"}}
	out, err := gofrogcmd.RunCmdOutput(pipFreezeCmd)
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Fatalf("Provided pip virtual-environment contains installed packages: %s\n. Please provide a clean environment.", out)
	}
}

func (pfc *PipCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "pip")
	cmd = append(cmd, pfc.Command)
	cmd = append(cmd, pfc.Options...)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (pfc *PipCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (pfc *PipCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (pfc *PipCmd) GetErrWriter() io.WriteCloser {
	return nil
}

func TestPipDepsTree(t *testing.T) {
	initPipTest(t)

	// Add virtual-environment path to 'PATH' for executing all pip and python commands inside the virtual-environment.
	pathValue := setPathEnvForPipInstall(t)
	if t.Failed() {
		t.FailNow()
	}
	defer os.Setenv("PATH", pathValue)

	// Check pip env is clean.
	validateEmptyPipEnv(t)

	// Populate cli config with 'default' server.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(cliutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	// Create test cases.
	allTests := []struct {
		name                 string
		project              string
		outputFolder         string
		moduleId             string
		args                 []string
		expectedDependencies int
		cleanAfterExecution  bool
	}{
		{"setuppy", "setuppyproject", "setuppy", "jfrog-python-example", []string{".", "--no-cache-dir", "--force-reinstall"}, 3, true},
		{"setuppy-verbose", "setuppyproject", "setuppy-verbose", "jfrog-python-example", []string{".", "--no-cache-dir", "--force-reinstall", "-v"}, 3, true},
		{"setuppy-with-module", "setuppyproject", "setuppy-with-module", "setuppy-with-module", []string{".", "--no-cache-dir", "--force-reinstall"}, 3, true},
		{"requirements", "requirementsproject", "requirements", tests.PipBuildName, []string{"-r", "requirements.txt", "--no-cache-dir", "--force-reinstall"}, 5, true},
		{"requirements-verbose", "requirementsproject", "requirements-verbose", tests.PipBuildName, []string{"-r", "requirements.txt", "--no-cache-dir", "--force-reinstall", "-v"}, 5, false},
		{"requirements-use-cache", "requirementsproject", "requirements-verbose", "requirements-verbose-use-cache", []string{"-r", "requirements.txt"}, 5, true},
	}

	// Run test cases.
	for _, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			testPipDepsTreeCmd(t, createPipProject(t, test.outputFolder, test.project), test.expectedDependencies, test.args)
			if test.cleanAfterExecution {
				// cleanup
				cleanPipTest(t, test.name)
			}
		})
	}
	cleanPipTest(t, "cleanup")
	tests.CleanFileSystem()
}

func testPipDepsTreeCmd(t *testing.T, projectPath string, expectedElements int, args []string) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	// Get pip configuration.
	pipConfig, err := piputils.GetPipConfiguration()
	if err != nil {
		t.Fatalf("Error occurred while attempting to read pip-configuration file: %s\n"+
			"Please run 'jfrog rt pip-config' command prior to running 'jfrog rt pip-deps-tree'.", err.Error())
	}
	// Set arg values.
	rtDetails, err := pipConfig.RtDetails()
	if err != nil {
		t.Fatal(err)
	}

	// Create command.
	pipDepsTreeCmd := pip.NewPipDepTreeCommand()
	pipDepsTreeCmd.SetRtDetails(rtDetails).SetRepo(pipConfig.TargetRepo()).SetArgs(args)

	err = pipDepsTreeCmd.Run()
	if err != nil {
		t.Fatalf("Failed while executing pip-deps-tree command: %s", err)
	}

	// Check result elements.
	treeJsonData, err := pipDepsTreeCmd.DepsTreeRoot.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed parsing tree json: %s", err)
	}

	// Count dependencies.
	var depsTreeTest []DependenciesTreeTest
	err = json.Unmarshal(treeJsonData, &depsTreeTest)
	if err != nil {
		t.Error(err)
	}
	depsCount := countDependencies(depsTreeTest)

	if expectedElements != depsCount {
		t.Errorf("Incorrect number of dependencies found, expected: %d, found: %d", expectedElements, depsCount)
	}
}

type DependenciesTreeTest struct {
	Id                 string                 `json:"id,omitempty"`
	DirectDependencies []DependenciesTreeTest `json:"dependencies,omitempty"`
}

func countDependencies(allDeps []DependenciesTreeTest) int {
	depsMap := make(map[string]int)
	// Iterate over dependencies, resolve and discover more dependencies.
	index := -1
	var currentDep string
	for {
		index++
		// Check if should stop.
		if len(allDeps) < index+1 {
			break
		}
		currentDep = allDeps[index].Id
		// Check if current dependency already resolved.
		if _, ok := depsMap[currentDep]; ok {
			// Already resolved.
			continue
		}
		// Add currentDep dependencies for cound.
		allDeps = append(allDeps, allDeps[index].DirectDependencies...)
	}
	return index
}
