package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jfrog/gofrog/io"
	rtLifecycle "github.com/jfrog/jfrog-cli-artifactory/lifecycle"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	configUtils "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	artifactoryclientUtils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/lifecycle"
	"github.com/jfrog/jfrog-client-go/lifecycle/services"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
)

const (
	artifactoryLifecycleMinVersion          = "7.68.3"
	signingKeyOptionalArtifactoryMinVersion = "7.104.1"
	promotionTypeFlagArtifactoryMinVersion  = "7.106.1"
	gpgKeyPairName                          = "lc-tests-key-pair"
	lcTestdataPath                          = "lifecycle"
	releaseBundlesSpec                      = "release-bundles-spec.json"
	buildsSpec12                            = "builds-spec-1-2.json"
	buildsSpec3                             = "builds-spec-3.json"
	prodEnvironment                         = "PROD"
	number1, number2, number3               = "111", "222", "333"
	withoutSigningKey                       = true
	artifactoryLifecycleSetTagMinVersion    = "7.111.0"
	rbManifestName                          = "release-bundle.json.evd"
	releaseBundlesV2                        = "release-bundles-v2"
	minMultiSourcesArtifactoryVersion       = "7.114.0"
)

var (
	lcDetails *configUtils.ServerDetails
	lcCli     *coreTests.JfrogCli
)

func TestBackwardCompatibleReleaseBundleCreation(t *testing.T) {
	cleanCallback := initLifecycleTest(t, artifactoryLifecycleMinVersion)
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

func TestReleaseBundleCreationFromMultiBuildsUsingCommandFlag(t *testing.T) {
	cleanCallback := initLifecycleTest(t, minMultiSourcesArtifactoryVersion)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	createRbFromMultiSourcesUsingCommandFlags(t, lcManager, createBuildsSource(), "", tests.LcRbName1, number1, "default", true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)
	assertStatusCompleted(t, lcManager, tests.LcRbName1, number1, "")
}

func TestReleaseBundleCreationFromMultiBundlesUsingCommandFlag(t *testing.T) {
	cleanCallback := initLifecycleTest(t, minMultiSourcesArtifactoryVersion)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName1, number1, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	createRbFromSpec(t, tests.LifecycleBuilds3, tests.LcRbName2, number2, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName2, number2)

	createRbFromMultiSourcesUsingCommandFlags(t, lcManager, "", createReleaseBundlesSource(), tests.LcRbName3, number3, "default", true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName3, number3)
	assertStatusCompleted(t, lcManager, tests.LcRbName3, number3, "")
}

func TestReleaseBundleCreationFromMultipleBuildsAndBundlesUsingCommandFlags(t *testing.T) {
	cleanCallback := initLifecycleTest(t, minMultiSourcesArtifactoryVersion)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName1, number1, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	createRbFromSpec(t, tests.LifecycleBuilds3, tests.LcRbName2, number2, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName2, number2)

	createRbFromMultiSourcesUsingCommandFlags(t, lcManager, createBuildsSource(), createReleaseBundlesSource(), tests.LcRbName3, number3, "default", true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName3, number3)
	assertStatusCompleted(t, lcManager, tests.LcRbName3, number3, "")
}

func TestReleaseBundleCreationFromMultipleSourcesUsingSpec(t *testing.T) {
	t.Skip("JGC-412 - Skipping TestReleaseBundleCreationFromMultipleSourcesUsingSpec")

	cleanCallback := initLifecycleTest(t, minMultiSourcesArtifactoryVersion)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName1, number1, true, true)

	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	createRbFromSpec(t, tests.LifecycleBuilds3, tests.LcRbName2, number2, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName2, number2)

	createRbFromSpec(t, tests.LifecycleMultipleSources, tests.LcRbName3, number3, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName3, number3)
	assertStatusCompleted(t, lcManager, tests.LcRbName3, number3, "")
}

func createReleaseBundlesSource() string {
	return fmt.Sprintf("name=%s, version=%s; name=%s, version=%s", tests.LcRbName1, number1, tests.LcRbName2, number2)
}

func createBuildsSource() string {
	return fmt.Sprintf("name=%s, id=%s, include-deps=%s; name=%s, id=%s", tests.LcBuildName1, number1, "true", tests.LcBuildName2, number2)
}

func TestReleaseBundleCreationFromAql(t *testing.T) {
	testReleaseBundleCreation(t, tests.UploadDevSpecA, tests.LifecycleAql, tests.GetExpectedLifecycleCreationByAql(), false)
}

func TestReleaseBundleCreationFromArtifacts(t *testing.T) {
	testReleaseBundleCreation(t, tests.UploadDevSpec, tests.LifecycleArtifacts, tests.GetExpectedLifecycleCreationByArtifacts(), false)
}

func TestReleaseBundleCreationFromArtifactsWithoutSigningKey(t *testing.T) {
	testReleaseBundleCreation(t, tests.UploadDevSpec, tests.LifecycleArtifacts, tests.GetExpectedLifecycleCreationByArtifacts(), withoutSigningKey)
}

func createRbFromMultiSourcesUsingCommandFlags(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, buildsSourcesOption, bundlesSourcesOption,
	rbName, rbVersion, project string, sync bool,
) {
	var sources []services.RbSource
	sources = buildMultiSources(sources, buildsSourcesOption, bundlesSourcesOption, project)

	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}
	queryParams := services.CommonOptionalQueryParams{
		Async:      !sync,
		ProjectKey: project,
	}

	_, err := lcManager.CreateReleaseBundlesFromMultipleSources(rbDetails, queryParams, gpgKeyPairName, sources)
	assert.NoError(t, err)
}

func buildMultiSources(sources []services.RbSource, buildsSourcesStr, bundlesSourcesStr, projectKey string) []services.RbSource {
	// Process Builds
	if buildsSourcesStr != "" {
		sources = buildMultiBuildSources(sources, buildsSourcesStr)
	}

	// Process Release Bundles
	if bundlesSourcesStr != "" {
		sources = buildMultiBundleSources(sources, bundlesSourcesStr, projectKey)
	}

	return sources
}

func buildMultiBundleSources(sources []services.RbSource, bundlesSourcesStr, projectKey string) []services.RbSource {
	var releaseBundleSources []services.ReleaseBundleSource
	bundleEntries := strings.Split(bundlesSourcesStr, ";")
	for _, entry := range bundleEntries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// Assuming the format "name=xxx, version=xxx"
		components := strings.Split(entry, ",")
		if len(components) != 2 {
			continue
		}
		name := strings.TrimSpace(strings.Split(components[0], "=")[1])
		version := strings.TrimSpace(strings.Split(components[1], "=")[1])

		releaseBundleSources = append(releaseBundleSources, services.ReleaseBundleSource{
			ProjectKey:           projectKey,
			ReleaseBundleName:    name,
			ReleaseBundleVersion: version,
		})
	}
	if len(releaseBundleSources) > 0 {
		sources = append(sources, services.RbSource{
			SourceType:     "release_bundles",
			ReleaseBundles: releaseBundleSources,
		})
	}
	return sources
}

func buildMultiBuildSources(sources []services.RbSource, sourcesStr string) []services.RbSource {
	var buildSources []services.BuildSource
	buildEntries := strings.Split(sourcesStr, ";")
	for _, entry := range buildEntries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// Assuming the format "name=xxx, number=xxx, include-dep=true"
		components := strings.Split(entry, ",")
		if len(components) < 2 {
			continue
		}

		name := strings.TrimSpace(strings.Split(components[0], "=")[1])
		number := strings.TrimSpace(strings.Split(components[1], "=")[1])

		includeDepStr := "false"
		if len(components) >= 3 {
			parts := strings.Split(components[2], "=")
			if len(parts) > 1 {
				includeDepStr = strings.TrimSpace(parts[1])
			}
		}

		includeDep, _ := strconv.ParseBool(includeDepStr)

		buildSources = append(buildSources, services.BuildSource{
			BuildRepository:     getBuildInfoRepositoryByProject("default"),
			BuildName:           name,
			BuildNumber:         number,
			IncludeDependencies: includeDep,
		})
	}
	if len(buildSources) > 0 {
		sources = append(sources, services.RbSource{
			SourceType: "builds",
			Builds:     buildSources,
		})
	}
	return sources
}

func getBuildInfoRepositoryByProject(projectKey string) string {
	buildRepo := "artifactory"
	if projectKey != "" && projectKey != "default" {
		buildRepo = projectKey
	}
	return buildRepo + "-build-info"
}

func testReleaseBundleCreation(t *testing.T, uploadSpec, creationSpec string, expected []string, withoutSigningKey bool) {
	if withoutSigningKey {
		cleanCallback := initLifecycleTest(t, signingKeyOptionalArtifactoryMinVersion)
		defer cleanCallback()
	} else {
		cleanCallback := initLifecycleTest(t, artifactoryLifecycleMinVersion)
		defer cleanCallback()
	}

	lcManager := getLcServiceManager(t)
	specFile, err := tests.CreateSpec(uploadSpec)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)

	createRbFromSpec(t, creationSpec, tests.LcRbName1, number1, true, withoutSigningKey)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	assertRbArtifacts(t, lcManager, tests.LcRbName1, number1, expected)
}

func TestLifecycleFullFlow(t *testing.T) {
	cleanCallback := initLifecycleTest(t, signingKeyOptionalArtifactoryMinVersion)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	// Upload builds to create release bundles from.
	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	// Create release bundle from builds synchronously.
	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName1, number1, true, false)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	// Create release bundle from a build asynchronously and assert status.
	// This build has dependencies which are included in the release bundle.
	createRbFromSpec(t, tests.LifecycleBuilds3, tests.LcRbName2, number2, false, false)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName2, number2)
	assertStatusCompleted(t, lcManager, tests.LcRbName2, number2, "")

	// Create a combined release bundle from the two previous release bundle.
	createRbFromSpec(t, tests.LifecycleReleaseBundles, tests.LcRbName3, number3, true, false)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName3, number3)

	// Promote the last release bundle to prod repo 1.
	promoteRb(t, lcManager, tests.LcRbName3, number3, tests.RtProdRepo1, "")
	// Assert the artifacts of both the initial release bundles made it to prod repo 1.
	assertExpectedArtifacts(t, tests.SearchAllProdRepo1, tests.GetExpectedLifecycleArtifacts())
	// Assert no artifacts were promoted to prod repo 2.
	assertExpectedArtifacts(t, tests.SearchAllProdRepo2, []string{})

	// Export release lifecycle bundle archive

	_, cleanUp := coreTests.CreateTempDirWithCallbackAndAssert(t)
	defer cleanUp()

	// TODO Temporarily disabling till export on testing suite is stable.
	/*exportRb(t, tests.LcRbName2, number2, tempDir)
	defer deleteExportedReleaseBundle(t, tests.LcRbName2)*/

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
	cleanCallback := initLifecycleTest(t, artifactoryLifecycleMinVersion)
	defer cleanCallback()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	testFilePath := filepath.Join(wd, "testdata", "lifecycle", "import", "cli-tests-2.zip")
	// Verify not supported
	err = lcCli.Exec("rbi", testFilePath)
	assert.Error(t, err)
}

func TestPromoteReleaseBundleWithPromotionTypeFlag(t *testing.T) {
	cleanCallback := initLifecycleTest(t, promotionTypeFlagArtifactoryMinVersion)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName1, number1, true, false)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)
	assertStatusCompleted(t, lcManager, tests.LcRbName1, number1, "")

	promoteRb(t, lcManager, tests.LcRbName1, number1, tests.RtProdRepo1, "move")
	assertStatusCompleted(t, lcManager, tests.LcRbName1, number1, "")
}

/*func deleteExportedReleaseBundle(t *testing.T, rbName string) {
	assert.NoError(t, os.RemoveAll(rbName))
}*/

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

func uploadBuildsWithProject(t *testing.T) func() {
	uploadBuildWithArtifactsAndProject(t, tests.UploadDevSpecA, tests.LcBuildName1, number1, tests.ProjectKey)
	uploadBuildWithArtifactsAndProject(t, tests.UploadDevSpecB, tests.LcBuildName2, number2, tests.ProjectKey)
	uploadBuildWithDepsAndProject(t, tests.LcBuildName3, number3, tests.ProjectKey)
	return func() {
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.LcBuildName1, artHttpDetails)
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.LcBuildName2, artHttpDetails)
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.LcBuildName3, artHttpDetails)
	}
}

func createRbBackwardCompatible(t *testing.T, specName, sourceOption, rbName, rbVersion string, sync bool) {
	specFile, err := getSpecFile(specName)
	assert.NoError(t, err)
	createRbWithFlags(t, specFile, sourceOption, "", "", rbName, rbVersion, "", sync, false)
}

func createRbFromSpec(t *testing.T, specName, rbName, rbVersion string, sync bool, withoutSigningKey bool) {
	specFile, err := tests.CreateSpec(specName)
	assert.NoError(t, err)
	createRbWithFlags(t, specFile, "spec", "", "", rbName, rbVersion, "", sync, withoutSigningKey)
}

func TestCreateBundleWithoutSpec(t *testing.T) {
	cleanCallback := initLifecycleTest(t, signingKeyOptionalArtifactoryMinVersion)
	defer cleanCallback()

	lcManager := getLcServiceManager(t)

	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	createRbWithFlags(t, "", "", tests.LcBuildName1, number1, tests.LcRbName1, number1, "default", false, false)
	assertStatusCompleted(t, lcManager, tests.LcRbName1, number1, "")
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	createRbWithFlags(t, "", "", tests.LcBuildName2, number2, tests.LcRbName2, number2, "default", false, true)
	assertStatusCompleted(t, lcManager, tests.LcRbName2, number2, "")
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName2, number2)
}

func TestCreateBundleWithoutSpecAndWithProject(t *testing.T) {
	cleanCallback := initLifecycleTest(t, signingKeyOptionalArtifactoryMinVersion)
	defer cleanCallback()
	deleteProject := createTestProject(t)
	if deleteProject != nil {
		defer func() {
			if err := deleteProject(); err != nil {
				t.Error(err)
			}
		}()
	}
	lcManager := getLcServiceManager(t)
	deleteBuilds := uploadBuildsWithProject(t)
	defer deleteBuilds()

	createRbWithFlags(t, "", "", tests.LcBuildName1, number1, tests.LcRbName1, number1, tests.ProjectKey, false, false)
	assertStatusCompletedWithProject(t, lcManager, tests.LcRbName1, number1, "", tests.ProjectKey)
	defer deleteReleaseBundleWithProject(t, lcManager, tests.LcRbName1, number1, tests.ProjectKey)
}

func createRbWithFlags(t *testing.T, specFilePath, sourceOption, buildName, buildNumber, rbName, rbVersion, project string,
	sync, withoutSigningKey bool,
) {
	argsAndOptions := []string{
		"rbc",
		rbName,
		rbVersion,
	}

	if specFilePath != "" {
		argsAndOptions = append(argsAndOptions, getOption(sourceOption, specFilePath))
	}

	if buildName != "" && buildNumber != "" {
		argsAndOptions = append(argsAndOptions, getOption(cliutils.BuildName, buildName))
		argsAndOptions = append(argsAndOptions, getOption(cliutils.BuildNumber, buildNumber))
	}

	if !withoutSigningKey {
		argsAndOptions = append(argsAndOptions, getOption(cliutils.SigningKey, gpgKeyPairName))
	}

	if sync {
		argsAndOptions = append(argsAndOptions, getOption(cliutils.Sync, "true"))
	}

	if project != "" {
		argsAndOptions = append(argsAndOptions, getOption(cliutils.Project, project))
	}

	assert.NoError(t, lcCli.Exec(argsAndOptions...))
}

/*func exportRb(t *testing.T, rbName, rbVersion, targetPath string) {
	lcCli.RunCliCmdWithOutput(t, "rbe", rbName, rbVersion, targetPath+"/")
	exists, err := fileutils.IsDirExists(path.Join(targetPath, rbName), false)
	assert.NoError(t, err)
	assert.Equal(t, true, exists)
}*/

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

func promoteRb(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion, promoteRepo, promotionType string) {
	cmdArgs := []string{
		"rbp", rbName, rbVersion, prodEnvironment,
		getOption(cliutils.SigningKey, gpgKeyPairName),
		getOption(cliutils.IncludeRepos, promoteRepo),
		"--project=default",
	}

	// Include promotion type if specified
	if promotionType != "" {
		cmdArgs = append(cmdArgs, getOption(cliutils.PromotionType, promotionType))
	}

	output := lcCli.RunCliCmdWithOutput(t, cmdArgs...)

	var promotionResp services.RbPromotionResp
	if !assert.NoError(t, json.Unmarshal([]byte(output), &promotionResp)) {
		return
	}
	assertStatusCompleted(t, lcManager, rbName, rbVersion, promotionResp.CreatedMillis.String())
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

// If createdMillis is provided, assert status for promotion. If blank, assert for creation.
func assertStatusCompletedWithProject(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion, createdMillis, projectKey string) {
	resp, err := getStatusWithProject(lcManager, rbName, rbVersion, createdMillis, projectKey)
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
		Url: *tests.JfrogUrl,
	}
	rtLifecycle.PlatformToLifecycleUrls(lcDetails)

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

func getStatusWithProject(lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion, createdMillis, projectKey string) (services.ReleaseBundleStatusResponse, error) {
	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}

	if createdMillis == "" {
		return lcManager.GetReleaseBundleCreationStatus(rbDetails, projectKey, true)
	}
	return lcManager.GetReleaseBundlePromotionStatus(rbDetails, projectKey, createdMillis, true)
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

func deleteReleaseBundleWithProject(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion, projectKey string) {
	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}

	assert.NoError(t, lcManager.DeleteReleaseBundleVersion(rbDetails, services.CommonOptionalQueryParams{Async: false, ProjectKey: projectKey}))
	// Wait after remote deleting. Can be removed once remote deleting supports sync.
	time.Sleep(5 * time.Second)
}

func TestSetReleaseBundleTag(t *testing.T) {
	cleanCallback := initLifecycleTest(t, artifactoryLifecycleSetTagMinVersion)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	// set tag
	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName1, number1, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)

	assertStatusCompleted(t, lcManager, tests.LcRbName1, number1, "")
	setReleaseBundleTag(t, lcManager, tests.LcRbName1, number1, "", "bundle-tag")

	// unset tag
	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName2, number1, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName2, number1)
	assertStatusCompleted(t, lcManager, tests.LcRbName2, number1, "")
	setReleaseBundleTag(t, lcManager, tests.LcRbName1, number1, "", "bundle-tag")
	unsetReleaseBundleTag(t, lcManager, tests.LcRbName1, number1)
}

func unsetReleaseBundleTag(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, version string) {
	setReleaseBundleTag(t, lcManager, rbName, version, "", "")
}

func TestReleaseBundleCreateOrUpdateProperties(t *testing.T) {
	cleanCallback := initLifecycleTest(t, artifactoryLifecycleSetTagMinVersion)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	// set properties
	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName1, number1, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)
	assertStatusCompleted(t, lcManager, tests.LcRbName1, number1, "")
	setReleaseBundleProperties(t, lcManager, tests.LcRbName1, number1, "default",
		"key1=value1;key2=value2")
	setReleaseBundleProperties(t, lcManager, tests.LcRbName1, number1, "default",
		"key1=value1;key2=''")
}

func TestReleaseBundleDeleteProperties(t *testing.T) {
	cleanCallback := initLifecycleTest(t, artifactoryLifecycleSetTagMinVersion)
	defer cleanCallback()
	lcManager := getLcServiceManager(t)

	deleteBuilds := uploadBuilds(t)
	defer deleteBuilds()

	// set and delete properties
	createRbFromSpec(t, tests.LifecycleBuilds12, tests.LcRbName1, number1, true, true)
	defer deleteReleaseBundle(t, lcManager, tests.LcRbName1, number1)
	assertStatusCompleted(t, lcManager, tests.LcRbName1, number1, "")
	setReleaseBundleProperties(t, lcManager, tests.LcRbName1, number1, "default",
		"key1=value1;key2=value2")
	deleteReleaseBundleProperties(t, lcManager, tests.LcRbName1, number1, "default", "key1,key2")
}

func deleteReleaseBundleProperties(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName,
	rbVersion, projectKey, delProps string,
) {
	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}
	queryParams := services.CommonOptionalQueryParams{
		Async:      false,
		ProjectKey: projectKey,
	}

	annotateParams := buildAnnotateParams("", "", delProps, false, false,
		true, rbDetails, queryParams)
	assert.NoError(t, lcManager.AnnotateReleaseBundle(annotateParams))
	// Wait after remote deleting. Can be removed once remote deleting supports sync.
	time.Sleep(5 * time.Second)
}

func setReleaseBundleTag(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion,
	projectKey, tag string,
) {
	log.Info(fmt.Sprintf("Setting release bundle tag=%s to: %s/%s", tag, rbName, rbVersion))
	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}
	queryParams := services.CommonOptionalQueryParams{
		Async:      false,
		ProjectKey: projectKey,
	}

	annotateParams := buildAnnotateParams(tag, "", "", true, false, false,
		rbDetails, queryParams)
	assert.NoError(t, lcManager.AnnotateReleaseBundle(annotateParams))
	// Wait after remote deleting. Can be removed once remote deleting supports sync.
	time.Sleep(5 * time.Second)
}

func setReleaseBundleProperties(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion,
	projectKey, properties string,
) {
	rbDetails := services.ReleaseBundleDetails{
		ReleaseBundleName:    rbName,
		ReleaseBundleVersion: rbVersion,
	}
	queryParams := services.CommonOptionalQueryParams{
		Async:      false,
		ProjectKey: projectKey,
	}

	annotateParams := buildAnnotateParams("", properties, "", false, true,
		false, rbDetails, queryParams)
	assert.NoError(t, lcManager.AnnotateReleaseBundle(annotateParams))
	// Wait after remote deleting. Can be removed once remote deleting supports sync.
	time.Sleep(5 * time.Second)
}

func buildAnnotateParams(tag, properties, keysToDelete string, tagExists, propsExist, delExist bool, rbDetails services.ReleaseBundleDetails,
	queryParams services.CommonOptionalQueryParams,
) services.AnnotateOperationParams {
	return services.AnnotateOperationParams{
		RbTag: services.RbAnnotationTag{
			Tag:   tag,
			Exist: tagExists,
		},
		RbProps: services.RbAnnotationProps{
			Properties: resolveProps(properties),
			Exist:      propsExist,
		},
		RbDelProps: services.RbDelProps{
			Keys:  keysToDelete,
			Exist: delExist,
		},
		RbDetails:   rbDetails,
		QueryParams: queryParams,
		PropertyParams: services.CommonPropParams{
			Path:      buildManifestPath(queryParams.ProjectKey, rbDetails.ReleaseBundleName, rbDetails.ReleaseBundleVersion),
			Recursive: false,
		},
		ArtifactoryUrl: services.ArtCommonParams{
			Url: serverDetails.ArtifactoryUrl,
		},
	}
}

func buildManifestPath(projectKey, bundleName, bundleVersion string) string {
	return fmt.Sprintf("%s/%s/%s/%s", buildRepoKey(projectKey), bundleName, bundleVersion, rbManifestName)
}

func buildRepoKey(project string) string {
	if project == "" || project == "default" {
		return releaseBundlesV2
	}
	return fmt.Sprintf("%s-%s", project, releaseBundlesV2)
}

func resolveProps(properties string) map[string][]string {
	if properties == "" {
		return make(map[string][]string)
	}

	props, err := artifactoryclientUtils.ParseProperties(properties)
	if err != nil {
		return make(map[string][]string)
	}

	return props.ToMap()
}

/*func remoteDeleteReleaseBundle(t *testing.T, lcManager *lifecycle.LifecycleServicesManager, rbName, rbVersion string) {
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
}*/

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

func uploadBuildWithArtifactsAndProject(t *testing.T, specFileName, buildName, buildNumber, projectKey string) {
	specFile, err := tests.CreateSpec(specFileName)
	assert.NoError(t, err)

	runRt(t, "upload", "--spec="+specFile, "--build-name="+buildName, "--build-number="+buildNumber, "--project="+projectKey)
	runRt(t, "build-publish", buildName, buildNumber, "--project="+projectKey)
}

func uploadBuildWithDepsAndProject(t *testing.T, buildName, buildNumber, projectKey string) {
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	randFile, err := io.CreateRandFile(filepath.Join(tests.Out, "dep-file"), 1000)
	assert.NoError(t, err)

	runRt(t, "upload", randFile.Name(), tests.RtDevRepo, "--flat", "--project="+projectKey)
	assert.NoError(t, lcCli.WithoutCredentials().Exec("rt", "bad", buildName, buildNumber, tests.RtDevRepo+"/dep-file", "--from-rt"))

	runRt(t, "build-publish", buildName, buildNumber, "--project="+projectKey)
}

func initLifecycleTest(t *testing.T, minVersion string) (cleanCallback func()) {
	if !*tests.TestLifecycle {
		t.Skip("Skipping lifecycle test. To run release bundle test add the '-test.lc=true' option.")
	}

	validateArtifactoryVersion(t, minVersion)

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
