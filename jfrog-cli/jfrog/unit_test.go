package main

import (
	"testing"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	clientTests "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/tests"
	"path/filepath"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

const (
	JfrogTestsHome      = ".jfrogTest"
	CliIntegrationTests = "jfrog-cli-go/jfrog-cli/jfrog"
)

func TestUnitTests(t *testing.T) {
	homePath, err := filepath.Abs(JfrogTestsHome)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	setJfrogHome(homePath)
	packages := clientTests.GetPackages()
	packages = clientTests.ExcludeTestsPackage(packages, CliIntegrationTests)
	clientTests.RunTests(packages)
	cleanUnitTestsJfrogHome(homePath)
}

func setJfrogHome(homePath string) {
	if err := os.Setenv(config.JfrogHomeEnv, homePath); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func cleanUnitTestsJfrogHome(homePath string) {
	os.RemoveAll(homePath)
	if err := os.Unsetenv(config.JfrogHomeEnv); err != nil {
		os.Exit(1)
	}
}
