package buildinfo

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"path/filepath"
	"strings"
	"testing"
)

const (
	withGit    = "git_test_.git_suffix"
	withoutGit = "git_test_no_.git_suffix"
	buildName  = "TestExtractGitUrl"
)

func TestExtractGitUrlWithDotGit(t *testing.T) {
	runTest(t, withGit)
}

func TestExtractGitUrlWithoutDotGit(t *testing.T) {
	runTest(t, withoutGit)
}

func runTest(t *testing.T, originalDir string) {
	baseDir, dotGitPath := tests.PrepareDotGitDir(t, originalDir, true)
	buildDir := getBuildDir(t)
	checkFailureAndClean(t, buildDir, dotGitPath)
	partials := getBuildInfoPartials(baseDir, t, buildName, "1")
	checkFailureAndClean(t, buildDir, dotGitPath)
	checkVCSUrl(partials, t)
	tests.RemovePath(buildDir, t)
	tests.RenamePath(dotGitPath, filepath.Join(tests.GetBaseDir(true), originalDir), t)
}

// Clean the environment if fails
func checkFailureAndClean(t *testing.T, buildDir string, oldPath string) {
	if t.Failed() {
		t.Log("Performing cleanup...")
		tests.RemovePath(buildDir, t)
		tests.RenamePath(oldPath, filepath.Join(tests.GetBaseDir(true), withGit), t)
		t.FailNow()
	}
}

func getBuildInfoPartials(baseDir string, t *testing.T, buildName string, buildNumber string) buildinfo.Partials {
	buildAddGitConfiguration := &BuildAddGitConfiguration{BuildName: buildName, BuildNumber: buildNumber, DotGitPath: baseDir}
	err := AddGit(buildAddGitConfiguration)
	if err != nil {
		t.Error("Cannot run build add git due to: " + err.Error())
		return nil
	}
	partials, err := utils.ReadPartialBuildInfoFiles(buildName, buildNumber)
	if err != nil {
		t.Error("Cannot read partial build info due to: " + err.Error())
		return nil
	}
	return partials
}

func getBuildDir(t *testing.T) string {
	buildDir, err := utils.GetBuildDir(buildName, "1")
	if err != nil {
		t.Error("Cannot create temp dir due to: " + err.Error())
		return ""
	}
	return buildDir
}

func checkVCSUrl(partials buildinfo.Partials, t *testing.T) {
	for _, partial := range partials {
		if partial.Vcs != nil {
			url := partial.Vcs.Url
			urlSplitted := strings.Split(url, ".git")
			if len(urlSplitted) != 2 {
				t.Error("Argumanets value is different then two: ", urlSplitted)
				break
			}
		} else {
			t.Error("VCS cannot be nil")
			break
		}
	}
}

func TestPopulateIssuesConfigurations(t *testing.T) {
	// Test success scenario
	expectedIssuesConfiguration := &IssuesConfiguration{
		ServerID: "local",
		TrackerName: "TESTING",
		TrackerUrl: "http://TESTING.com",
		Regexp: `([a-zA-Z]+-[0-9]*)\s-\s(.*)`,
		KeyGroupIndex: 1,
		SummaryGroupIndex: 2,
		Aggregate: true,
		AggregationStatus: "RELEASE",
		LogLimit: 100,
	}
	ic := new(IssuesConfiguration)
	// Build config from file
	err := ic.populateIssuesConfigurationsFromSpec("../testdata/buildissues/issuesconfig_success.yaml")
	// Check they are equal
	if err != nil {
		t.Error(fmt.Sprintf("Reading configurations file ended with error: %s", err.Error()))
		t.FailNow()
	}
	if *ic != *expectedIssuesConfiguration {
		t.Error(fmt.Sprintf("Failed reading configurations file. Expected: %+v Received: %+v", *expectedIssuesConfiguration, *ic))
		t.FailNow()
	}

	// Test failing scenarios
	failing := []string{
		"../testdata/buildissues/issuesconfig_fail_no_issues.yaml",
		"../testdata/buildissues/issuesconfig_fail_no_server.yaml",
		"../testdata/buildissues/issuesconfig_fail_invalid_groupindex.yaml",
		"../testdata/buildissues/issuesconfig_fail_invalid_aggregate.yaml",
	}

	for _, config := range failing {
		err = ic.populateIssuesConfigurationsFromSpec(config)
		if err == nil {
			t.Error(fmt.Sprintf("Reading configurations file was supposed to end with error: %s", config))
			t.FailNow()
		}
	}
}

func TestAddGitDoCollect(t *testing.T) {
	// Create git folder with files
	originalFolder := "issues_.git_suffix"
	baseDir, dotGitPath := tests.PrepareDotGitDir(t, originalFolder, true)

	// Create BuildAddGitConfiguration
	config := BuildAddGitConfiguration{
		IssuesConfig: &IssuesConfiguration{
			LogLimit: 100,
			Aggregate: false,
			SummaryGroupIndex: 2,
			KeyGroupIndex: 1,
			Regexp: `(.+-[0-9]+)\s-\s(.+)`,
			TrackerName: "test",
		},
		BuildNumber: "1",
		BuildName: "cli-test-build-issues",
		ConfigFilePath: "",
		DotGitPath: dotGitPath,
	}

	// Collect issues
	issues, err := config.DoCollect(config.IssuesConfig, "")
	if err != nil {
		// Error - should succeed
		t.Error(err)
	}
	if len(issues) != 2 {
		// Error - should be empty
		t.Errorf("Issues list expected to have 2 issues, instead found %d issues.", len(issues))
	}

	// Clean previous git path and set new
	tests.RenamePath(dotGitPath, filepath.Join(baseDir, originalFolder), t)
	originalFolder = "issues2_.git_suffix"
	baseDir, dotGitPath = tests.PrepareDotGitDir(t, originalFolder, true)

	// Collect issues
	issues, err = config.DoCollect(config.IssuesConfig, "6198a6294722fdc75a570aac505784d2ec0d1818")
	if err != nil {
		// Error - should succeed
		t.Error(err)
	}
	if len(issues) != 2 {
		// Error - should find 3 issues
		t.Errorf("Issues list expected to have 2 issues, instead found %d issues.", len(issues))
	}

	// Clean git path
	tests.RenamePath(dotGitPath, filepath.Join(baseDir, originalFolder), t)
}
