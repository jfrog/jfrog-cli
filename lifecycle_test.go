package main

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	configUtils "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	lifecycleCli "github.com/jfrog/jfrog-cli/lifecycle"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/lifecycle"
	"github.com/jfrog/jfrog-client-go/lifecycle/services"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/distribution"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	artifactoryLifecycleMinVersion = "7.68.3"
	gpgKeyPairName                 = "lc-tests-key-pair"
	lcTestdataPath                 = "lifecycle"
	releaseBundlesSpec             = "release-bundles-spec.json"
	buildsSpec12                   = "builds-spec-1-2.json"
	buildsSpec3                    = "builds-spec-3.json"
	prodEnvironment                = "PROD"
	number1, number2, number3      = "111", "222", "333"
)

var (
	lcDetails *configUtils.ServerDetails
	lcCli     *coreTests.JfrogCli
)

func TestLifecycle(t *testing.T) {
	initLifecycleTest(t)
	defer cleanLifecycleTests(t)
	lcManager := getLcServiceManager(t)

	// Upload builds to create release bundles from.
	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	// Create release bundle from builds synchronously.
	createRb(t, buildsSpec12, cliutils.Builds, tests.LcRbName1, number1, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	// Create release bundle from builds asynchronously and assert status.
	createRb(t, buildsSpec3, cliutils.Builds, tests.LcRbName2, number2, false)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName2, number2)
	assertStatusCompleted(t, lcManager, tests.LcRbName2, number2, "")

	// Create a combined release bundle from the two previous release bundle.
	createRb(t, releaseBundlesSpec, cliutils.ReleaseBundles, tests.LcRbName3, number3, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName3, number3)

	// Promote the last release bundle.
	promoteRb(t, lcManager, number3)

	// Verify the artifacts of both the initial release bundles made it to the prod repo.
	searchProdSpec, err := tests.CreateSpec(tests.SearchAllProdRepo)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetExpectedLifecycleArtifacts(), searchProdSpec, serverDetails, t)

	distributeRb(t)
	defer remoteDeleteReleaseBundle(t, lcManager, tests.LcRbName3, number3)

	// Verify the artifacts were distributed correctly by the provided path mappings.
	searchDevSpec, err := tests.CreateSpec(tests.SearchAllDevRepo)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetExpectedLifecycleDistributedArtifacts(), searchDevSpec, serverDetails, t)
}

func uploadBuilds(t *testing.T) func() {
	uploadBuild(t, tests.UploadDevSpecA, tests.LcBuildName1, number1)
	uploadBuild(t, tests.UploadDevSpecB, tests.LcBuildName2, number2)
	uploadBuild(t, tests.UploadDevSpecC, tests.LcBuildName3, number3)
	return func() {
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.LcBuildName1, artHttpDetails)
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.LcBuildName2, artHttpDetails)
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.LcBuildName3, artHttpDetails)
	}
}

func createRb(t *testing.T, specName, sourceOption, rbName, rbVersion string, sync bool) {
	specFile, err := getSpecFile(specName)
	assert.NoError(t, err)
	argsAndOptions := []string{
		"rbc",
		rbName,
		rbVersion,
		getOption(sourceOption, specFile),
		getOption(cliutils.SigningKey, gpgKeyPairName),
	}
	// Add the --sync option only if requested, to test the default value.
	if sync {
		argsAndOptions = append(argsAndOptions, getOption(cliutils.Sync, "true"))
	}
	assert.NoError(t, lcCli.Exec(argsAndOptions...))
}

func distributeRb(t *testing.T) {
	distributionRulesPath := filepath.Join(tests.GetTestResourcesPath(), "distribution", tests.DistributionRules)
	assert.NoError(t, lcCli.Exec(
		"rbd", tests.LcRbName3, number3,
		getOption(cliutils.DistRules, distributionRulesPath),
		getOption(cliutils.PathMappingPattern, "(*)/(*)"),
		getOption(cliutils.PathMappingTarget, "{1}/target/{2}"),
	))
	// Wait after distribution before asserting. Can be removed once distribute supports sync.
	time.Sleep(5 * time.Second)
}

func getOption(option, value string) string {
	return fmt.Sprintf("--%s=%s", option, value)
}

func promoteRb(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbVersion string) {
	output := lcCli.RunCliCmdWithOutput(t, "rbp", tests.LcRbName3, rbVersion, prodEnvironment,
		getOption(cliutils.SigningKey, gpgKeyPairName),
		getOption(cliutils.Overwrite, "true"),
		"--project=default")
	var promotionResp services.RbPromotionResp
	if !assert.NoError(t, json.Unmarshal([]byte(output), &promotionResp)) {
		return
	}
	assertStatusCompleted(t, lcManager, tests.LcRbName3, rbVersion, promotionResp.CreatedMillis.String())
}

func getSpecFile(fileName string) (string, error) {
	source := filepath.Join(tests.GetTestResourcesPath(), lcTestdataPath, fileName)
	return tests.ReplaceTemplateVariables(source, "")
}

// If createdMillis is provided, assert status for promotion. If blank, assert for creation.
func assertStatusCompleted(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion, createdMillis string) {
	resp, err := getStatus(lcManager, rbName, rbVersion, createdMillis)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, services.Completed, resp.Status)
}

func getLcServiceManager(t *testing.T) *lifecycle.LifecycleServicesManager {
	lcManager, err := utils.CreateLifecycleServiceManager(lcDetails, false)
	assert.NoError(t, err)
	return lcManager
}

func authenticateLifecycle() string {
	*tests.JfrogUrl = clientUtils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	lcDetails = &configUtils.ServerDetails{
		Url: *tests.JfrogUrl}
	lifecycleCli.PlatformToLifecycleUrls(lcDetails)

	cred := fmt.Sprintf("--url=%s", *tests.JfrogUrl)
	if *tests.JfrogAccessToken != "" {
		lcDetails.AccessToken = *tests.JfrogAccessToken
		cred += fmt.Sprintf(" --access-token=%s", lcDetails.AccessToken)
	} else {
		lcDetails.User = *tests.JfrogUser
		lcDetails.Password = *tests.JfrogPassword
		cred += fmt.Sprintf(" --user=%s --password=%s", lcDetails.User, lcDetails.Password)
	}
	return cred
}

func getStatus(lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion, createdMillis string) (services.ReleaseBundleStatusResponse, error) {
	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}

	if createdMillis == "" {
		return lcManager.GetReleaseBundleCreationStatus(rbDetails, "", true)
	}
	return lcManager.GetReleaseBundlePromotionStatus(rbDetails, "", createdMillis, true)
}

func deleteReleaseBundle(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion string) {
	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}

	assert.NoError(t, lcManager.DeleteReleaseBundle(rbDetails, services.ReleaseBundleQueryParams{Async: false}))
	// Wait after remote deleting. Can be removed once remote deleting supports sync.
	time.Sleep(5 * time.Second)
}

func remoteDeleteReleaseBundle(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion string) {
	params := distribution.NewDistributeReleaseBundleParams(rbName, rbVersion)
	rules := &distribution.DistributionCommonParams{
		SiteName:     "*",
		CityName:     "*",
		CountryCodes: []string{"*"},
	}
	params.DistributionRules = append(params.DistributionRules, rules)

	assert.NoError(t, lcManager.RemoteDeleteReleaseBundle(params, false))
	// Wait after remote deleting. Can be removed once remote deleting supports sync.
	time.Sleep(5 * time.Second)
}

func uploadBuild(t *testing.T, specFileName, buildName, buildNumber string) {
	specFile, err := tests.CreateSpec(specFileName)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile, "--build-name="+buildName, "--build-number="+buildNumber)
	runRt(t, "build-publish", buildName, buildNumber)
}

func initLifecycleTest(t *testing.T) {
	if !*tests.TestLifecycle {
		t.Skip("Skipping lifecycle test. To run release bundle test add the '-test.lc=true' option.")
	}
	validateArtifactoryVersion(t, artifactoryLifecycleMinVersion)

	if !isLifecycleSupported(t) {
		t.Skip("Skipping lifecycle test because the functionality is not enabled on the provided JPD.")
	}
}

func isLifecycleSupported(t *testing.T) (skip bool) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	resp, _, _, err := client.SendGet(serverDetails.ArtifactoryUrl+"api/release_bundles/records/non-existing-rb", true, artHttpDetails, "")
	if !assert.NoError(t, err) {
		return
	}
	return resp.StatusCode != http.StatusNotImplemented
}

func InitLifecycleTests() {
	initArtifactoryCli()
	initLifecycleCli()
	cleanUpOldBuilds()
	cleanUpOldRepositories()
	cleanUpOldUsers()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
	sendGpgKeyPair()
}

func initLifecycleCli() {
	if lcCli != nil {
		return
	}
	lcCli = coreTests.NewJfrogCli(execMain, "jfrog", authenticateLifecycle())
}

func CleanLifecycleTests() {
	deleteCreatedRepos()
}

func cleanLifecycleTests(t *testing.T) {
	deleteFilesFromRepo(t, tests.RtDevRepo)
	deleteFilesFromRepo(t, tests.RtProdRepo)
	tests.CleanFileSystem()
}

func sendGpgKeyPair() {
	// Create http client
	client, err := httpclient.ClientBuilder().Build()
	coreutils.ExitOnErr(err)

	// Check if one already exists
	resp, body, _, err := client.SendGet(*tests.JfrogUrl+"artifactory/api/security/keypair/"+gpgKeyPairName, true, artHttpDetails, "")
	coreutils.ExitOnErr(err)
	if resp.StatusCode == http.StatusOK {
		return
	}
	coreutils.ExitOnErr(errorutils.CheckResponseStatusWithBody(resp, body, http.StatusNotFound))

	// Read gpg public and private keys
	keysDir := filepath.Join(tests.GetTestResourcesPath(), lcTestdataPath, "keys")
	publicKey, err := os.ReadFile(filepath.Join(keysDir, "public.txt"))
	coreutils.ExitOnErr(err)
	privateKey, err := os.ReadFile(filepath.Join(keysDir, "private.txt"))
	coreutils.ExitOnErr(err)

	// Send keys to Artifactory
	payload := KeyPairPayload{
		PairName:   gpgKeyPairName,
		PairType:   "GPG",
		Alias:      gpgKeyPairName + "-alias",
		Passphrase: "password",
		PublicKey:  string(publicKey),
		PrivateKey: string(privateKey),
	}
	content, err := json.Marshal(payload)
	coreutils.ExitOnErr(err)
	resp, body, err = client.SendPost(*tests.JfrogUrl+"artifactory/api/security/keypair", content, artHttpDetails, "")
	coreutils.ExitOnErr(err)
	coreutils.ExitOnErr(errorutils.CheckResponseStatusWithBody(resp, body, http.StatusCreated))
}

type KeyPairPayload struct {
	PairName   string `json:"pairName,omitempty"`
	PairType   string `json:"pairType,omitempty"`
	Alias      string `json:"alias,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
}
