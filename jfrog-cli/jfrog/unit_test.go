package main

import (
	"testing"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	clientTests "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/tests"
)

const (
	JfrogTestsHome      = ".jfrogTest"
	CliIntegrationTests = "jfrog-cli-go/jfrog-cli/jfrog"
)

func TestUnitTests(t *testing.T) {
	setJfrogHome()
	packages := clientTests.GetPackages()
	packages = clientTests.ExcludeTestsPackage(packages, CliIntegrationTests)
	clientTests.RunTests(packages)
	unsetJfrogHome()
}

func setJfrogHome() {
	if err := os.Setenv(config.JfrogHomeEnv, JfrogTestsHome); err != nil {
		os.Exit(1)
	}
}

func unsetJfrogHome() {
	if err := os.Unsetenv(config.JfrogHomeEnv); err != nil {
		os.Exit(1)
	}
}
