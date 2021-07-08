package main

import (
	"os"
	"testing"

	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/log"
	clientTests "github.com/jfrog/jfrog-client-go/utils/tests"
)

const (
	JfrogTestsHome      = ".jfrogTest"
	CliIntegrationTests = "github.com/jfrog/jfrog-cli"
)

func TestUnitTests(t *testing.T) {
	// Create temp jfrog home
	err, cleanUpJfrogHome := coreTests.SetJfrogHome()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	// Clean from previous tests.
	defer cleanUpJfrogHome()

	packages := clientTests.GetTestPackages("./...")
	packages = clientTests.ExcludeTestsPackage(packages, CliIntegrationTests)
	clientTests.RunTests(packages, *tests.HideUnitTestLog)
}
