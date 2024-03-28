package main

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/gofrog/io"
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
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"path"
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

func TestBackwardCompatibleReleaseBundleCreation(t *testing.T) {
	cleanCallback := initLifecycleTest(t)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	// Upload builds to create release bundles from.
	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	// Create release bundle from builds synchronously.
	createRbBackwardCompatible(t, buildsSpec12, cliutils.Builds, tests.LcRbName1, number1, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	// Create release bundle from a build asynchronously and assert status.
	// This build has dependencies which are included in the release bundle.
	createRbBackwardCompatible(t, buildsSpec3, cliutils.Builds, tests.LcRbName2, number2, false)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName2, number2)
	assertStatusCompleted(t, lcManager, tests.LcRbName2, number2, "")

	// Create a combined release bundle from the two previous release bundle.
	createRbBackwardCompatible(t, releaseBundlesSpec, cliutils.ReleaseBundles, tests.LcRbName3, number3, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName3, number3)

	assertRbArtifacts(t, lcManager, tests.LcRbName3, number3, tests.GetExpectedBackwardCompatibleLifecycleArtifacts())
}

func assertRbArtifacts(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion string, expected []string) {
	specs, err := getReleaseBundleSpecification(lcManager, rbName, rbVersion)
	if !assert.NoError(t, err) {
		return
	}
	compareRbArtifacts(t, specs, expected)
}

func compareRbArtifacts(t *testing.T, actual services.ReleaseBundleSpecResponse, expected []string) {
	var actualArtifactsPaths []string
	for _, artifact := range actual.Artifacts {
		actualArtifactsPaths = append(actualArtifactsPaths, path.Join(artifact.SourceRepositoryKey, artifact.Path))
	}
	assert.ElementsMatch(t, actualArtifactsPaths, expected)
}

func TestReleaseBundleCreationFromAql(t *testing.T) {
	testReleaseBundleCreation(t, tests.UploadDevSpecA, tests.LifecycleAql, tests.GetExpectedLifecycleCreationByAql())
}

func TestReleaseBundleCreationFromArtifacts(t *testing.T) {
	testReleaseBundleCreation(t, tests.UploadDevSpec, tests.LifecycleArtifacts, tests.GetExpectedLifecycleCreationByArtifacts())
}

func testReleaseBundleCreation(t *testing.T, uploadSpec, creationSpec string, expected []string) {
	cleanCallback := initLifecycleTest(t)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	specFile, err := tests.CreateSpec(uploadSpec)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)

	createRbFromSpec(t, creationSpec, tests.LcRbName1, number1, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	assertRbArtifacts(t, lcManager, tests.LcRbName1, number1, expected)
}

func TestLifecycleFullFlow(t *testing.T) {
	cleanCallback := initLifecycleTest(t)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	// Upload builds to create release bundles from.
	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	// Create release bundle from builds synchronously.
	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName1, number1, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	// Create release bundle from a build asynchronously and assert status.
	// This build has dependencies which are included in the release bundle.
	createRbFromSpec(t, tests.LifecycleBuilds3, tests.LcRbName2, number2, false)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName2, number2)
	assertStatusCompleted(t, lcManager, tests.LcRbName2, number2, "")

	// Create a combined release bundle from the two previous release bundle.
	createRbFromSpec(t, tests.LifecycleReleaseBundles, tests.LcRbName3, number3, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName3, number3)

	// Promote the last release bundle to prod repo 1.
	promoteRb(t, lcManager, number3)

	// Assert the artifacts of both the initial release bundles made it to prod repo 1.
	assertExpectedArtifacts(t, tests.SearchAllProdRepo1, tests.GetExpectedLifecycleArtifacts())
	// Assert no artifacts were promoted to prod repo 2.
	assertExpectedArtifacts(t, tests.SearchAllProdRepo2, []string{})

	// Export release lifecycle bundle archive

	tempDir, cleanUp := coreTests.CreateTempDirWithCallbackAndAssert(t)
	defer cleanUp()

	exportRb(t, tests.LcRbName2, number2, tempDir)
	defer deleteExportedReleaseBundle(t, tests.LcRbName2)

	// TODO Temporarily disabling till distribution on testing suite is stable.
	/*
		distributeRb(t)
		defer remoteDeleteReleaseBundle(t, lcManager, tests.LcRbName3, number3)

		// Verify the artifacts were distributed correctly by the provided path mappings.
		assertExpectedArtifacts(t, tests.SearchAllDevRepo, tests.GetExpectedLifecycleDistributedArtifacts())
	*/

}

// Import bundles only work on onPerm platforms
func TestImportReleaseBundle(t *testing.T) {
	cleanCallback := initLifecycleTest(t)
	defer cleanCallback()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	testFilePath := filepath.Join(wd, "testdata", "lifecycle", "import", "rb-import-test.zip")
	// Verify not supported
	assert.Error(t, lcCli.Exec("rbi", testFilePath))
}

func deleteExportedReleaseBundle(t *testing.T, rbName string) {
	assert.NoError(t, os.RemoveAll(rbName))
}

func assertExpectedArtifacts(t *testing.T, specFileName string, expected []string) {
	searchProdSpec, err := tests.CreateSpec(specFileName)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(expected, searchProdSpec, serverDetails, t)
}

func uploadBuilds(t *testing.T) func() {
	uploadBuildWithArtifacts(t, tests.UploadDevSpecA, tests.LcBuildName1, number1)
	uploadBuildWithArtifacts(t, tests.UploadDevSpecB, tests.LcBuildName2, number2)
	uploadBuildWithDeps(t, tests.LcBuildName3, number3)
	return func() {
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.LcBuildName1, artHttpDetails)
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.LcBuildName2, artHttpDetails)
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.LcBuildName3, artHttpDetails)
	}
}

func createRbBackwardCompatible(t *testing.T, specName, sourceOption, rbName, rbVersion string, sync bool) {
	specFile, err := getSpecFile(specName)
	assert.NoError(t, err)
	createRb(t, specFile, sourceOption, rbName, rbVersion, sync)
}

func createRbFromSpec(t *testing.T, specName, rbName, rbVersion string, sync bool) {
	specFile, err := tests.CreateSpec(specName)
	assert.NoError(t, err)
	createRb(t, specFile, "spec", rbName, rbVersion, sync)
}

func createRb(t *testing.T, specFilePath, sourceOption, rbName, rbVersion string, sync bool) {
	argsAndOptions := []string{
		"rbc",
		rbName,
		rbVersion,
		getOption(sourceOption, specFilePath),
		getOption(cliutils.SigningKey, gpgKeyPairName),
	}
	// Add the --sync option only if requested, to test the default value.
	if sync {
		argsAndOptions = append(argsAndOptions, getOption(cliutils.Sync, "true"))
	}
	assert.NoError(t, lcCli.Exec(argsAndOptions...))
}

func exportRb(t *testing.T, rbName, rbVersion, targetPath string) {
	lcCli.RunCliCmdWithOutput(t, "rbe", rbName, rbVersion, targetPath+"/")
	exists, err := fileutils.IsDirExists(path.Join(targetPath, rbName), false)
	assert.NoError(t, err)
	assert.Equal(t, true, exists)
}

/*
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
*/

func getOption(option, value string) string {
	return fmt.Sprintf("--%s=%s", option, value)
}

func promoteRb(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbVersion string) {
	output := lcCli.RunCliCmdWithOutput(t, "rbp", tests.LcRbName3, rbVersion, prodEnvironment,
		getOption(cliutils.SigningKey, gpgKeyPairName),
		getOption(cliutils.IncludeRepos, tests.RtProdRepo1),
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

func getReleaseBundleSpecification(lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion string) (services.ReleaseBundleSpecResponse, error) {
	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}

	return lcManager.GetReleaseBundleSpecification(rbDetails)
}

func deleteReleaseBundle(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion string) {
	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}

	assert.NoError(t, lcManager.DeleteReleaseBundleVersion(rbDetails, services.CommonOptionalQueryParams{Async: false}))
	// Wait after remote deleting. Can be removed once remote deleting supports sync.
	time.Sleep(5 * time.Second)
}

/*
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
*/

func uploadBuildWithArtifacts(t *testing.T, specFileName, buildName, buildNumber string) {
	specFile, err := tests.CreateSpec(specFileName)
	assert.NoError(t, err)

	runRt(t, "upload", "--spec="+specFile, "--build-name="+buildName, "--build-number="+buildNumber)
	runRt(t, "build-publish", buildName, buildNumber)
}

func uploadBuildWithDeps(t *testing.T, buildName, buildNumber string) {
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	randFile, err := io.CreateRandFile(filepath.Join(tests.Out, "dep-file"), 1000)
	assert.NoError(t, err)

	runRt(t, "upload", randFile.Name(), tests.RtDevRepo, "--flat")
	assert.NoError(t, lcCli.WithoutCredentials().Exec("rt", "bad", buildName, buildNumber, tests.RtDevRepo+"/dep-file", "--from-rt"))

	runRt(t, "build-publish", buildName, buildNumber)
}

func initLifecycleTest(t *testing.T) (cleanCallback func()) {
	if !*tests.TestLifecycle {
		t.Skip("Skipping lifecycle test. To run release bundle test add the '-test.lc=true' option.")
	}
	validateArtifactoryVersion(t, artifactoryLifecycleMinVersion)

	if !isLifecycleSupported(t) {
		t.Skip("Skipping lifecycle test because the functionality is not enabled on the provided JPD.")
	}

	// Populate cli config with 'default' server.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	return func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
		cleanLifecycleTests(t)
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
	deleteFilesFromRepo(t, tests.RtProdRepo1)
	deleteFilesFromRepo(t, tests.RtProdRepo2)
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
