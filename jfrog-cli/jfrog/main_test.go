package main

import (
	"flag"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetLogger(log.NewLogger(cliutils.GetCliLogLevel()))
	setupIntegrationTests()
	result := m.Run()
	tearDownIntegrationTests()
	os.Exit(result)
}

func setupIntegrationTests() {
	flag.Parse()
	*tests.RtUrl = utils.AddTrailingSlashIfNeeded(*tests.RtUrl)

	if *tests.TestBintray {
		InitBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		InitArtifactoryTests()
	}
	if *tests.TestBuildTools || *tests.TestGo || *tests.TestNuget {
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
	if *tests.TestBuildTools || *tests.TestGo || *tests.TestNuget {
		CleanBuildToolsTests()
	}
}
