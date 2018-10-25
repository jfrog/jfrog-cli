package buildinfo

import (
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
	err := AddGit(buildName, buildNumber, baseDir)
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
