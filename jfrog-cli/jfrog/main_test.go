package main

import (
	"flag"
	"os"
	"testing"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/tests"
)

func TestMain(m *testing.M) {
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
	if *tests.TestBuildTools || *tests.TestVgo {
		InitBuildToolsTests()
	}
	if *tests.TestDocker {
		InitDockerTests()
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
