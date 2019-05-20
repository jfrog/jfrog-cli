package main

import (
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/log"
	clientTests "github.com/jfrog/jfrog-client-go/utils/tests"
	"os"
	"path/filepath"
	"testing"
)

const (
	JfrogTestsHome      = ".jfrogTest"
	CliIntegrationTests = "github.com/jfrog/jfrog-cli-go"
)

func TestUnitTests(t *testing.T) {
	homePath, err := filepath.Abs(JfrogTestsHome)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	setJfrogHome(homePath)
	packages := clientTests.GetTestPackages("./...")
	packages = clientTests.ExcludeTestsPackage(packages, CliIntegrationTests)
	clientTests.RunTests(packages, *tests.HideUnitTestLog)
	cleanUnitTestsJfrogHome(homePath)
}

func setJfrogHome(homePath string) {
	if err := os.Setenv(cliutils.JfrogHomeDirEnv, homePath); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func cleanUnitTestsJfrogHome(homePath string) {
	os.RemoveAll(homePath)
	if err := os.Unsetenv(cliutils.JfrogHomeDirEnv); err != nil {
		os.Exit(1)
	}
}
