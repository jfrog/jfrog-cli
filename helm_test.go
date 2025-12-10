package main

import (
	"testing"

	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

var (
	helmCli *coreTests.JfrogCli
)

func InitHelmTests() {
	initArtifactoryCli()
	initHelmCli()
	cleanUpOldBuilds()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
}

func CleanHelmTests() {
	deleteCreatedRepos()
}

func initHelmCli() {
	if helmCli != nil {
		return
	}
	helmCli = coreTests.NewJfrogCli(execMain, "jfrog", authenticate(false))
}

func initHelmTest(t *testing.T) {
	if !*tests.TestHelm {
		t.Skip("Skipping Helm test. To run Helm test add the '-test.helm=true' option.")
	}
}

func TestHelmExample(t *testing.T) {
	initHelmTest(t)
}
