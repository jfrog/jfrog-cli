package commands

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"strings"
)

func TestExtractGitUrlWithDotGit(t *testing.T) {
	baseDir, dotGitPath := preparation(t)
	buildDir := getBuildDir(t)
	checkFailureAndClean(t, buildDir, dotGitPath)

	createConfigFile(true, dotGitPath, t)
	createHeadFile(dotGitPath, t)
	checkFailureAndClean(t, buildDir, dotGitPath)

	checkAndCleanPrevious(buildDir, t)
	checkFailureAndClean(t, buildDir, dotGitPath)

	partials := getBuildInfoPartials(baseDir, t, "TestExtractGitUrlWithDotGit", "1")
	checkFailureAndClean(t, buildDir, dotGitPath)
	checkVCSUrl(partials, t)
	checkAndCleanPrevious(buildDir, t)
	checkAndCleanPrevious(dotGitPath, t)
}

func TestExtractGitUrlWithoutDotGit(t *testing.T) {
	baseDir, dotGitPath := preparation(t)
	buildDir := getBuildDir(t)
	checkFailureAndClean(t, buildDir, dotGitPath)

	createConfigFile(false, dotGitPath, t)
	createHeadFile(dotGitPath, t)
	checkFailureAndClean(t, buildDir, dotGitPath)

	checkAndCleanPrevious(buildDir, t)
	checkFailureAndClean(t, buildDir, dotGitPath)

	partials := getBuildInfoPartials(baseDir, t, "TestExtractGitUrlWithoutDotGit", "1")
	checkFailureAndClean(t, buildDir, dotGitPath)
	checkVCSUrl(partials, t)
	checkAndCleanPrevious(buildDir, t)
	checkAndCleanPrevious(dotGitPath, t)
}

// Clean the environment if fails
func checkFailureAndClean(t *testing.T, buildDir string, dotGitPath string) {
	if t.Failed() {
		t.Log("Performing cleanup...")
		checkAndCleanPrevious(buildDir, t)
		checkAndCleanPrevious(dotGitPath, t)
		t.FailNow()
	}
}

func getBuildInfoPartials(baseDir string, t *testing.T, buildName string, buildNumber string) buildinfo.Partials {
	err := BuildAddGit(buildName, buildNumber, baseDir)
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
	buildDir, err := utils.GetBuildDir("TestExtractGitUrlWithDotGit", "1")
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
			if len(urlSplitted) > 2 {
				t.Error("Too many argumanets: ", urlSplitted)
				break
			}
		} else {
			t.Error("VCS cannot be nil")
			break
		}
	}
}
// Need to prepare the environment such as .git directory and the config, head files.
func preparation(t *testing.T) (string, string) {
	pwd, _ := os.Getwd()
	baseDir := filepath.Join(pwd, "testdata")
	dotGitPath := filepath.Join(baseDir, ".git")
	checkAndCleanPrevious(dotGitPath, t)
	createDirs(dotGitPath, t)
	return baseDir, dotGitPath
}

// This method is needed to pass the revision check on the ReadConfig
func createHeadFile(path string, t *testing.T) {
	path = filepath.Join(path, "HEAD")
	f, err := os.Create(path)
	if err != nil {
		t.Error("Cannot create file: " + path + " due to: " + err.Error())
		return
	}
	defer f.Close()
	config := "rev-info"
	_, err = f.WriteString(config)
	if err != nil {
		t.Error("Cannot write to file: " + path + " due to: " + err.Error())
		return
	}
	f.Sync()
}

func createConfigFile(withGit bool, path string, t *testing.T) {
	path = filepath.Join(path, "config")

	f, err := os.Create(path)
	if err != nil {
		t.Error("Cannot create file: " + path + " due to: " + err.Error())
		return
	}
	defer f.Close()
	config := "[remote \"origin\"]\n" +
		"url = https://github.com/JFrogDev/jfrog-cli-go"
	if withGit {
		config += ".git"
	}
	_, err = f.WriteString(config)
	if err != nil {
		t.Error("Cannot write to file: " + path + " due to: " + err.Error())
		return
	}

	f.Sync()
}

func checkAndCleanPrevious(testPath string, t *testing.T) {
	if _, err := os.Stat(testPath); err == nil {
		//path exists need to delete.
		err = os.RemoveAll(testPath)
		if err != nil {
			t.Error("Cannot remove path: " + testPath + " due to: " + err.Error())
		}
	}
}

func createDirs(path string, t *testing.T) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		t.Error("Cannot create path: " + path + " due to: " + err.Error())

	}
}
