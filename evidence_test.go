package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	evidenceTests "github.com/jfrog/jfrog-cli-evidence/tests/e2e/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
)

var (
	evidenceCli                *coreTests.JfrogCli
	evidenceServerDetails      *config.ServerDetails
	evidenceArtifactoryDetails *config.ServerDetails
	evidenceServicesManager    artifactory.ArtifactoryServicesManager
)

// TestEvidence runs all Evidence E2E tests using the main runner
func TestEvidence(t *testing.T) {
	initEvidenceTest(t)

	// Ensure initialization is complete
	if evidenceCli == nil || evidenceServicesManager == nil {
		t.Fatal("Evidence CLI or Artifactory services manager not initialized. Make sure InitEvidenceTests() was called.")
	}

	// Create the Evidence E2E test runner
	runner := evidenceTests.NewEvidenceE2ETestsRunner(evidenceCli, evidenceServicesManager)

	// Run all E2E tests using the main runner
	runner.RunEvidenceCliTests(t)
}

// initEvidenceTest initializes evidence tests
func initEvidenceTest(t *testing.T) {
	if !*tests.TestEvidence {
		t.Skip("Skipping evidence test. To run evidence test add the '-test.evidence=true' option.")
	}

	// Verify we have required tokens
	if *tests.JfrogAccessToken == "" {
		t.Fatal("Evidence tests require --jfrog.adminToken flag")
	}
	if *tests.EvidenceAccessToken == "" {
		t.Fatal("Evidence tests require --jfrog.evidenceToken flag")
	}
}

// InitEvidenceTests initializes the evidence CLI and server details
func InitEvidenceTests() {
	initArtifactoryCli()
	initEvidenceCli()
	configureEvidenceServer()
	configureEvidenceArtifactory()
	initEvidenceArtifactoryClient()
}

// initEvidenceCli initializes the Evidence CLI
func initEvidenceCli() {
	if evidenceCli != nil {
		return
	}
	evidenceCli = coreTests.NewJfrogCli(execMain, "jfrog evd", "")
}

// configureEvidenceServer sets up server details for Evidence service
func configureEvidenceServer() {
	if evidenceServerDetails != nil {
		return
	}

	evidenceServerDetails = &config.ServerDetails{
		Url:         *tests.JfrogUrl,
		AccessToken: *tests.EvidenceAccessToken,
	}

	// Set environment variables for Evidence CLI (it reads from env)
	// The Evidence CLI uses JFROG_CLI_URL and access token from standard JFrog CLI config
	if err := os.Setenv("JFROG_CLI_URL", *tests.JfrogUrl); err != nil {
		fmt.Printf("Error setting JFROG_CLI_URL: %v\n", err)
	}
	if err := os.Setenv("JFROG_CLI_ACCESS_TOKEN", *tests.EvidenceAccessToken); err != nil {
		fmt.Printf("Error setting JFROG_CLI_ACCESS_TOKEN: %v\n", err)
	}

	fmt.Printf("✓ Evidence service configured for: %s\n", *tests.JfrogUrl)
}

// configureEvidenceArtifactory sets up Artifactory details for Evidence tests
func configureEvidenceArtifactory() {
	if evidenceArtifactoryDetails != nil {
		return
	}

	artifactoryUrl := clientutils.AddTrailingSlashIfNeeded(*tests.JfrogUrl) + "artifactory/"

	evidenceArtifactoryDetails = &config.ServerDetails{
		Url:            *tests.JfrogUrl,
		ArtifactoryUrl: artifactoryUrl,
		AccessToken:    *tests.JfrogAccessToken,
	}

	fmt.Printf("✓ Artifactory configured for Evidence tests: %s\n", artifactoryUrl)
}

// initEvidenceArtifactoryClient initializes the Artifactory services manager for Evidence tests
func initEvidenceArtifactoryClient() {
	if evidenceServicesManager != nil {
		return
	}

	var err error
	evidenceServicesManager, err = utils.CreateServiceManager(evidenceArtifactoryDetails, -1, 0, false)
	if err != nil {
		fmt.Printf("Error creating Artifactory services manager: %v\n", err)
		return
	}

	fmt.Printf("✓ Artifactory services manager initialized for Evidence tests\n")
}

// CleanEvidenceTests cleans up after evidence tests
func CleanEvidenceTests() {
	// Evidence tests clean up their own resources (repositories, artifacts) via t.Cleanup()
	// No global cleanup needed
}

