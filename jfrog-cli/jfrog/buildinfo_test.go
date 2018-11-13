package main

import (
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/git"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/jfrog/inttestutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/ioutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestBuildAddDependenciesFromHomeDir(t *testing.T) {
	initArtifactoryTest(t)
	// Clean old build tests if exists
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.BuildAddDepsBuildName, artHttpDetails)

	fileName := "cliTestFile.txt"
	testFileRelPath, testFileAbs := createFileInHomeDir(t, fileName)

	test := buildAddDepsBuildInfoTestParams{description: "'rt bad' from home dir", commandArgs: []string{testFileRelPath, "--recursive=false"}, expectedDependencies: []string{fileName}, buildNumber: "1"}
	collectDepsAndPublishBuild(test, t)
	validateBuildAddDepsBuildInfo(t, test)

	os.Remove(testFileAbs)
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.BuildAddDepsBuildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestBuildAddDependenciesDryRun(t *testing.T) {
	initArtifactoryTest(t)
	// Clean old build tests if exists
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.BuildAddDepsBuildName, artHttpDetails)
	err := utils.RemoveBuildDir(tests.BuildAddDepsBuildName, "1")
	if err != nil {
		t.Error(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	defer os.Chdir(wd)
	err = os.Chdir("../testsdata")
	if err != nil {
		t.Error(err)
	}

	noCredsCli := tests.NewJfrogCli(main, "jfrog rt", "")
	// Execute tha bad command
	noCredsCli.Exec("bad", tests.BuildAddDepsBuildName, "1", ioutils.PrepareFilePathForWindows("a/*"), "--dry-run=true")
	buildDir, err := utils.GetBuildDir(tests.BuildAddDepsBuildName, "1")
	if err != nil {
		t.Error(err)
	}

	files, _ := ioutil.ReadDir(buildDir)
	if len(files) > 0 {
		t.Error(errors.New("'rt bad' command with dry-run failed. The dry-run option has no effect."))
	}

	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.BuildAddDepsBuildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestBuildAddDependencies(t *testing.T) {
	initArtifactoryTest(t)
	// Clean old build tests if exists
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.BuildAddDepsBuildName, artHttpDetails)
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	defer os.Chdir(wd)
	err = os.Chdir("../testsdata")
	if err != nil {
		t.Error(err)
	}

	buildNumbers := []string{}
	allFiles := []string{"a1.in", "a2.in", "a3.in", "b1.in", "b2.in", "b3.in", "c1.in", "c2.in", "c3.in"}

	var badTests = []buildAddDepsBuildInfoTestParams{
		{description: "'rt bad' simple cli", commandArgs: []string{ioutils.PrepareFilePathForWindows("a/*")}, expectedDependencies: allFiles},
		{description: "'rt bad' single file", commandArgs: []string{ioutils.PrepareFilePathForWindows("a/a1.in")}, expectedDependencies: []string{"a1.in"}},
		{description: "'rt bad' none recursive", commandArgs: []string{ioutils.PrepareFilePathForWindows("a/*"), "--recursive=false"}, expectedDependencies: []string{"a1.in", "a2.in", "a3.in"}},
		{description: "'rt bad' special chars recursive", commandArgs: []string{getSpecialCharFilePath()}, expectedDependencies: []string{"a1.in"}},
		{description: "'rt bad' exclude command line wildcards", commandArgs: []string{ioutils.PrepareFilePathForWindows("../testsdata/a/*"), "--exclude-patterns=*a2*;*a3.in"}, expectedDependencies: []string{"a1.in", "b1.in", "b2.in", "b3.in", "c1.in", "c2.in", "c3.in"}},
		{description: "'rt bad' spec", commandArgs: []string{"--spec=" + tests.GetFilePath(tests.BuildAddDepsSpec)}, expectedDependencies: allFiles},
		{description: "'rt bad' two specFiles", commandArgs: []string{"--spec=" + tests.GetFilePath(tests.BuildAddDepsDoubleSpec)}, expectedDependencies: []string{"a1.in", "a2.in", "a3.in", "b1.in", "b2.in", "b3.in"}},
		{description: "'rt bad' exclude command line regexp", commandArgs: []string{ioutils.PrepareFilePathForWindows("a/a(.*)"), "--exclude-patterns=(.*)a2.*;.*a3.in", "--regexp=true"}, expectedDependencies: []string{"a1.in"}},
	}

	for i, badTest := range badTests {
		badTest.buildNumber = strconv.Itoa(i + 1)
		buildNumbers = append(buildNumbers, badTest.buildNumber)
		collectDepsAndPublishBuild(badTest, t)
		validateBuildAddDepsBuildInfo(t, badTest)
	}

	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.BuildAddDepsBuildName, artHttpDetails)
	clearTempBuildFiles(tests.BuildAddDepsBuildName, buildNumbers)
}

// Test publish build info without --build-url
func TestArtifactoryPublishBuildInfo(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumber := "cli-test-build", "10"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)

	body := uploadFilesAndGetBuildInfo(t, buildName, buildNumber, "")

	// Validate no build url
	_, _, _, err := jsonparser.Get(body, "buildInfo", "url")
	if err == nil {
		t.Error("Build url is expected to be empty")
	}

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

// Test publish build info with --build-url
func TestArtifactoryPublishBuildInfoBuildUrl(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumber := "cli-test-build", "11"
	buildUrl := "http://example.ci.com"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)

	body := uploadFilesAndGetBuildInfo(t, buildName, buildNumber, buildUrl)

	// Validate correctness of build url
	actualBuildUrl, err := jsonparser.GetString(body, "buildInfo", "url")
	if err != nil {
		t.Error(err)
	}
	if buildUrl != actualBuildUrl {
		t.Errorf("Build url expected %v, got %v", buildUrl, actualBuildUrl)
	}

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryCleanBuildInfo(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumber := "cli-test-build", "11"
	buildNameNotToPromote := "cli-test-build-not-to-promote"

	//upload files with buildName and buildNumber
	specFile, err := tests.PreparePath(tests.UploadSpec)
	if err != nil {
		t.Error(err)
	}
	artifactoryCli.Exec("upload", "--spec="+specFile, "--build-name="+buildNameNotToPromote, "--build-number="+buildNumber)

	//cleanup buildInfo
	artifactoryCli.WithSuffix("").Exec("build-clean", buildName, buildNumber)

	//upload files with buildName and buildNumber
	specFile, err = tests.PreparePath(tests.SimpleUploadSpec)
	if err != nil {
		t.Error(err)
	}
	artifactoryCli.Exec("upload", "--spec="+specFile, "--build-name="+buildName, "--build-number="+buildNumber)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	//promote buildInfo
	artifactoryCli.Exec("build-promote", buildName, buildNumber, tests.Repo2)

	//validate files are uploaded with the build info name and number
	props := fmt.Sprintf("build.name=%v;build.number=%v", buildName, buildNumber)
	isExistInArtifactoryByProps(tests.GetSimpleUploadExpectedRepo2(), tests.Repo2+"/*", props, t)

	//cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.BuildAddDepsBuildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestCollectGitBuildInfo(t *testing.T) {
	initArtifactoryTest(t)
	gitCollectCliRunner := tests.NewJfrogCli(main, "jfrog rt", "")
	buildName, buildNumber := "cli-test-build", "13"
	dotGitPath := ioutils.FixWinPath(getCliDotGitPath(t))
	gitCollectCliRunner.Exec("build-add-git", buildName, buildNumber, dotGitPath)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	client := httpclient.NewDefaultHttpClient()
	_, body, _, err := client.SendGet(artifactoryDetails.Url+"api/build/"+buildName+"/"+buildNumber, false, artHttpDetails)
	if err != nil {
		t.Error(err)
	}
	buildInfoVcsRevision, err := jsonparser.GetString(body, "buildInfo", "vcsRevision")
	if err != nil {
		t.Error(err)
	}
	buildInfoVcsUrl, err := jsonparser.GetString(body, "buildInfo", "vcsUrl")
	if err != nil {
		t.Error(err)
	}
	if buildInfoVcsRevision == "" {
		t.Error("Failed to get git revision.")
	}

	if buildInfoVcsUrl == "" {
		t.Error("Failed to get git remote url.")
	}

	gitManager := git.NewManager(dotGitPath)
	if err = gitManager.ReadConfig(); err != nil {
		t.Error("Failed to read .git config file.")
	}
	if gitManager.GetRevision() != buildInfoVcsRevision {
		t.Error("Wrong revision", "expected: "+gitManager.GetRevision(), "Got: "+buildInfoVcsRevision)
	}

	gitConfigUrl := gitManager.GetUrl()
	if gitConfigUrl != buildInfoVcsUrl {
		t.Error("Wrong url", "expected: "+gitConfigUrl, "Got: "+buildInfoVcsUrl)
	}

	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.BuildAddDepsBuildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestReadGitConfig(t *testing.T) {
	dotGitPath := getCliDotGitPath(t)
	gitManager := git.NewManager(dotGitPath)
	err := gitManager.ReadConfig()
	if err != nil {
		t.Error("Failed to read .git config file.")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get current dir.")
	}
	gitExecutor := tests.GitExecutor(workingDir)
	revision, _, err := gitExecutor.GetRevision()
	if err != nil {
		t.Error(err)
		return
	}

	if gitManager.GetRevision() != revision {
		t.Error("Wrong revision", "expected: "+revision, "Got: "+gitManager.GetRevision())
	}

	url, _, err := gitExecutor.GetUrl()
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.HasSuffix(url, ".git") {
		url += ".git"
	}

	if gitManager.GetUrl() != url {
		t.Error("Wrong url", "expected: "+url, "Got: "+gitManager.GetUrl())
	}
}

func uploadFilesAndGetBuildInfo(t *testing.T, buildName, buildNumber, buildUrl string) []byte {
	//upload files with buildName and buildNumber
	specFile, err := tests.PreparePath(tests.SimpleUploadSpec)
	if err != nil {
		t.Error(err)
	}
	artifactoryCli.Exec("upload", "--spec="+specFile, "--build-name="+buildName, "--build-number="+buildNumber)

	//publish buildInfo
	publishBuildInfoArgs := []string{"build-publish", buildName, buildNumber}
	if buildUrl != "" {
		publishBuildInfoArgs = append(publishBuildInfoArgs, "--build-url="+buildUrl)
	}
	artifactoryCli.Exec(publishBuildInfoArgs...)

	//validate files are uploaded with the build info name and number
	props := fmt.Sprintf("build.name=%v;build.number=%v", buildName, buildNumber)
	isExistInArtifactoryByProps(tests.GetSimpleUploadExpectedRepo1(), tests.Repo1+"/*", props, t)

	//download build info
	buildInfoUrl := fmt.Sprintf("%vapi/build/%v/%v", artifactoryDetails.Url, buildName, buildNumber)
	client := httpclient.NewDefaultHttpClient()
	_, body, _, err := client.SendGet(buildInfoUrl, false, artHttpDetails)
	if err != nil {
		t.Error(err)
	}
	return body
}

func collectDepsAndPublishBuild(badTest buildAddDepsBuildInfoTestParams, t *testing.T) {
	noCredsCli := tests.NewJfrogCli(main, "jfrog rt", "")
	// Remove old tests data from fs if exists
	err := utils.RemoveBuildDir(tests.BuildAddDepsBuildName, badTest.buildNumber)
	if err != nil {
		t.Error(err)
	}

	command := []string{"bad", tests.BuildAddDepsBuildName, badTest.buildNumber}
	// Execute tha bad command
	noCredsCli.Exec(append(command, badTest.commandArgs...)...)
	artifactoryCli.Exec("bp", tests.BuildAddDepsBuildName, badTest.buildNumber)
}

func validateBuildAddDepsBuildInfo(t *testing.T, buildInfoTestParams buildAddDepsBuildInfoTestParams) {
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.BuildAddDepsBuildName, buildInfoTestParams.buildNumber, t, artHttpDetails)
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		// Case no module was not created
		t.Errorf("%s test with the command: \nrt bad %s \nexpected to have module with the following dependencies: \n%s \nbut has no modules: \n%s",
			buildInfoTestParams.description, buildInfoTestParams.commandArgs, buildInfoTestParams.expectedDependencies, buildInfo)
	}
	if len(buildInfoTestParams.expectedDependencies) != len(buildInfo.Modules[0].Dependencies) {
		// The checksums are ignored when comparing the actual and the expected
		t.Errorf("%s test with the command: \nrt bad %s  \nexpected to have the following dependencies: \n%s \nbut has: \n%s",
			buildInfoTestParams.description, buildInfoTestParams.commandArgs, buildInfoTestParams.expectedDependencies, dependenciesToPrintableArray(buildInfo.Modules[0].Dependencies))
	}

	for _, expectedDependency := range buildInfoTestParams.expectedDependencies {
		found := false
		for _, actualDependency := range buildInfo.Modules[0].Dependencies {
			if actualDependency.Id == expectedDependency {
				found = true
				break
			}
		}
		if !found {
			// The checksums are ignored when comparing the actual and the expected
			t.Errorf("%s test with the command: \nrt bad %s \nexpected to have the following dependencies: \n%s \nbut has: \n%s",
				buildInfoTestParams.description, buildInfoTestParams.commandArgs, buildInfoTestParams.expectedDependencies, dependenciesToPrintableArray(buildInfo.Modules[0].Dependencies))
		}
	}
}

func clearTempBuildFiles(buildName string, buildNumbers []string) {
	for _, buildNumber := range buildNumbers {
		utils.RemoveBuildDir(buildName, buildNumber)
	}

}

func dependenciesToPrintableArray(dependencies []buildinfo.Dependency) []string {
	ids := []string{}
	for _, dependency := range dependencies {
		ids = append(ids, dependency.Id)
	}
	return ids
}

type buildAddDepsBuildInfoTestParams struct {
	description          string
	commandArgs          []string
	expectedDependencies []string
	buildNumber          string
	validationFunc       func(*testing.T, buildAddDepsBuildInfoTestParams)
}
