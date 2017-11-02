package main

import (
	"testing"
	"flag"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/tests"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	clientTests "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/tests"
)

const (
	JfrogTestsHome      = ".jfrogTest"
	CliIntegrationTests = "jfrog-cli-go/jfrog-cli/jfrog"
)

func TestMain(m *testing.M) {
	setJfrogHome()
	packages := clientTests.GetPackages()
	packages = clientTests.ExcludeTestsPackage(packages, CliIntegrationTests)
	clientTests.RunTests(packages)
	unsetJfrogHome()
	setupIntegrationTests()
	result := m.Run()
	tearDownIntegrationTests()
	os.Exit(result)
}

func setupIntegrationTests() {
	flag.Parse()
	if *tests.TestBintray {
		InitBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		InitArtifactoryTests()
	}
	if *tests.TestBuildTools {
		InitBuildToolsTests()
	}
}

func tearDownIntegrationTests() {
	if *tests.TestBintray {
		CleanBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		CleanArtifactoryTests()
	}
	if *tests.TestBuildTools {
		CleanBuildToolsTests()
	}
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