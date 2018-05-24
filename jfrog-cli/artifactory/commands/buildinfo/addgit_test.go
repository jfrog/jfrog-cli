package buildinfo

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
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

func runTest(t *testing.T, originalFolder string) {
	baseDir, dotGitPath := preparation(t, originalFolder)
	buildDir := getBuildDir(t)
	checkFailureAndClean(t, buildDir, dotGitPath, originalFolder)
	partials := getBuildInfoPartials(baseDir, t, buildName, "1")
	checkFailureAndClean(t, buildDir, dotGitPath, originalFolder)
	checkVCSUrl(partials, t)
	removePath(buildDir, t)
	renamePath(dotGitPath, filepath.Join(getBaseDir(), originalFolder), t)
}

// Clean the environment if fails
func checkFailureAndClean(t *testing.T, buildDir string, oldPath, newPath string) {
	if t.Failed() {
		t.Log("Performing cleanup...")
		removePath(buildDir, t)
		renamePath(oldPath, filepath.Join(getBaseDir(), withGit), t)
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

// Need to prepare the environment such as .git directory and the config, head files.
// Renaming the already prepared folders.
func preparation(t *testing.T, path string) (string, string) {
	baseDir := getBaseDir()
	dotGitPath := filepath.Join(baseDir, ".git")
	removePath(dotGitPath, t)
	dotGitPathTest := filepath.Join(baseDir, path)
	renamePath(dotGitPathTest, dotGitPath, t)
	return baseDir, dotGitPath
}

func removePath(testPath string, t *testing.T) {
	if _, err := os.Stat(testPath); err == nil {
		//path exists need to delete.
		err = os.RemoveAll(testPath)
		if err != nil {
			t.Error("Cannot remove path: " + testPath + " due to: " + err.Error())
		}
	}
}

func renamePath(oldPath, newPath string, t *testing.T) {
	err := fileutils.CopyDir(oldPath, newPath, true)
	if err != nil {
		t.Error("Error copying directory: ", oldPath, "to", newPath, err.Error())
		t.FailNow()
	}
	removePath(oldPath, t)
}

func getBaseDir() (baseDir string) {
	pwd, _ := os.Getwd()
	pwd = filepath.Dir(pwd)
	baseDir = filepath.Join(pwd, "testdata")
	return
}
