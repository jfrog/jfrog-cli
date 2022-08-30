package main

import (
	buildInfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferfiles"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/repostate"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

var targetArtHttpDetails *httputils.HttpClientDetails
var targetServerDetails *config.ServerDetails
var targetArtifactoryCli *tests.JfrogCli

func InitTransferTests() {
	initArtifactoryCli()
	cleanUpOldRepositories()
	cleanUpOldBuilds()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
	inttestutils.InstallDataTransferPlugin()
	var creds string
	creds, targetServerDetails, targetArtHttpDetails = inttestutils.AuthenticateTarget()
	targetArtifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", creds)
	inttestutils.CreateTargetRepos(targetArtifactoryCli)
}

func CleanTransferTests() {
	deleteCreatedRepos()
	inttestutils.DeleteTargetRepos(targetArtifactoryCli)
	cleanTestsHomeEnv()
}

func initTransferTest(t *testing.T) func() {
	if !*tests.TestTransfer {
		t.Skip("Skipping transfer test. To run transfer test add the '-test.transfer=true' option.")
	}
	oldHomeDir, newHomeDir := prepareHomeDir(t)

	// Delete the target server if exist
	targetConfig, err := commands.GetConfig(inttestutils.TargetServerId, false)
	if err == nil && targetConfig.ServerId != "" {
		err = configCli.WithoutCredentials().Exec("rm", inttestutils.TargetServerId, "--quiet")
		assert.NoError(t, err)
	}
	*tests.JfrogUrl = utils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	err = tests.NewJfrogCli(execMain, "jfrog config", "--access-token="+*tests.JfrogTargetAccessToken).Exec("add", inttestutils.TargetServerId, "--interactive=false", "--artifactory-url="+*tests.JfrogTargetUrl+tests.ArtifactoryEndpoint)
	assert.NoError(t, err)

	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("curl", "-XPOST", "/api/plugins/reload"))
	return func() {
		cleanArtifactory()
		inttestutils.CleanTargetRepos(targetArtifactoryCli)
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
		tests.CleanFileSystem()
	}
}

func TestTransferTwoRepos(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Populate source Artifactory
	repo1Spec, repo2Spec := inttestutils.UploadTransferTestFilesAndAssert(artifactoryCli, serverDetails, t)

	// Execute transfer-files
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1+";"+tests.RtRepo2))

	// Verify again that the files exist the source Artifactory.
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, serverDetails, t)
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo2(), repo2Spec, serverDetails, t)

	// Verify files were transferred to the target Artifactory
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, targetServerDetails, t)
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo2(), repo2Spec, targetServerDetails, t)
}

func TestTransferExcludeRepo(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Populate source Artifactory
	repo1Spec, repo2Spec := inttestutils.UploadTransferTestFilesAndAssert(artifactoryCli, serverDetails, t)

	// Execute transfer-files
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1+";"+tests.RtRepo2, "--exclude-repos="+tests.RtRepo2))

	// Verify again that the files exist the source Artifactory.
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, serverDetails, t)
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo2(), repo2Spec, serverDetails, t)

	// Verify repo1 files were transferred to the target Artifactory
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, targetServerDetails, t)
	inttestutils.VerifyExistInArtifactory([]string{}, repo2Spec, targetServerDetails, t)
}

func TestTransferDiff(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Execute transfer-files on empty repo1
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1))

	// Populate source Artifactory
	repo1Spec, _ := inttestutils.UploadTransferTestFilesAndAssert(artifactoryCli, serverDetails, t)

	// Execute diff
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1))
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, targetServerDetails, t)
}

func TestTransferProperties(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Populate source Artifactory
	repo1Spec, _ := inttestutils.UploadTransferTestFilesAndAssert(artifactoryCli, serverDetails, t)
	artifactoryCli.Exec("sp", "key1=value1;key2=value2,value3", "--spec="+repo1Spec)

	// Execute transfer-files
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1))

	// Verify properties
	resultItems, err := inttestutils.SearchInArtifactory(repo1Spec, targetServerDetails, t)
	assert.NoError(t, err)
	assert.Len(t, resultItems, 9)
	for _, item := range resultItems {
		properties := item.Props
		assert.GreaterOrEqual(t, len(properties), 2)
		for k, v := range properties {
			switch k {
			case "key1":
				assert.ElementsMatch(t, []string{"value1"}, v)
			case "key2":
				assert.Len(t, v, 2)
				assert.ElementsMatch(t, []string{"value2", "value3"}, v)
			case "sha256":
				// Do nothing
			default:
				assert.Fail(t, "Unexpected property key "+k)
			}
		}
	}
}

func TestTransferMaven(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	assert.NoError(t, runMaven(t, createSimpleMavenProject, tests.MavenConfig, "install", "--build-name="+tests.MvnBuildName, "--build-number=1"))

	// Verify files were uploaded to the source Artifactory
	mvnRepoSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), mvnRepoSpec, serverDetails, t)

	// Publish build info
	runRt(t, "build-publish", tests.MvnBuildName, "1")
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.MvnBuildName, artHttpDetails)
	defer inttestutils.DeleteBuild(targetServerDetails.ArtifactoryUrl, tests.MvnBuildName, *targetArtHttpDetails)

	// Execute transfer-files
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.MvnRepo1+";artifactory-build-info"))

	// Verify maven files were transferred to the target Artifactory
	inttestutils.VerifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), mvnRepoSpec, targetServerDetails, t)

	// Wait for creation of maven-metadata.xml in the target Artifactory
	inttestutils.WaitForCreationInArtifactory(tests.MvnRepo1+"/org/jfrog/cli-test/maven-metadata.xml", targetServerDetails, t)

	// Verify build exist in the target Artifactory
	publishedBuildInfo, found, err := tests.GetBuildInfo(targetServerDetails, tests.MvnBuildName, "1")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	validateSpecificModule(publishedBuildInfo.BuildInfo, t, 2, 2, 0, "org.jfrog:cli-test:1.0", buildInfo.Maven)
}

func TestTransferPaginationAndThreads(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Decrease the AQL pagination limit to 10
	transferfiles.AqlPaginationLimit = 10
	defer func() {
		transferfiles.AqlPaginationLimit = transferfiles.DefaultAqlPaginationLimit
	}()

	// Upload 101 files to the source Artifactory
	for i := 0; i < 101; i++ {
		fileIndexStr := strconv.FormatInt(int64(i), 10)
		assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("curl", "-s", "-XPUT", tests.RtRepo1+"/"+fileIndexStr, "-d "+fileIndexStr))
	}

	// Asynchronously exec transfer-files and increase the threads count by 1
	var wg sync.WaitGroup
	done := false
	inttestutils.AsyncExecTransferFiles(artifactoryCli, &wg, &done, t)
	inttestutils.AsyncUpdateThreads(&wg, &done, t)
	wg.Wait()

	// Verify 101 files were uploaded to the target
	assert.Equal(t, 101, inttestutils.CountArtifactsInPath(tests.RtRepo1, targetServerDetails, t))
}

func TestTransferWithRepoState(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Creates a new repo state file on the jfrog home, and override it with a custom one.
	repoStateFile, err := transferfiles.GetRepoStateFilePath(tests.RtRepo1)
	assert.NoError(t, err)
	generateTestRepoStateFile(t, tests.RtRepo1, repoStateFile)

	// Populate source Artifactory.
	repo1Spec, _ := inttestutils.UploadTransferTestFilesAndAssert(artifactoryCli, serverDetails, t)

	// Execute transfer-files.
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1))

	// Assert repo state file was removed after the full transfer was completed.
	exists, err := fileutils.IsFileExists(repoStateFile, false)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Verify again that the files exist the source Artifactory.
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, serverDetails, t)

	// Verify only the expected files were transferred to the target Artifactory:
	// 'a' - only files included in state, because it was explored.
	// 'b' - all files, because it was unexplored.
	// 'c' - no files, because it was completed.
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepoState(), repo1Spec, targetServerDetails, t)
}

// Creates a new repo state file on the jfrog home, and overrides it with a custom one:
// root - testdata - a -> a1.in
//					   -> a3.in
//					   - b -> b1.in
//					   	   - c
// 'a' is marked as explored but not completed, we expect its files to be uploaded.
// 'b' is marked as unexplored, we expect its files names to be re-explored and then uploaded.
// 'c' is marked completed, so we expect no action there.
func generateTestRepoStateFile(t *testing.T, repoKey, repoStatePath string) {
	stateManager, created, err := repostate.LoadOrCreateRepoStateManager(repoKey, repoStatePath)
	assert.NoError(t, err)
	assert.True(t, created)

	childTestdata := addChildWithFiles(stateManager.Root, "testdata", true, false)
	childA := addChildWithFiles(childTestdata, "a", true, false, "a1.in", "a3.in")
	childB := addChildWithFiles(childA, "b", false, false, "b1.in")
	_ = addChildWithFiles(childB, "c", false, true)

	assert.NoError(t, stateManager.SaveToFile())
	exists, err := fileutils.IsFileExists(repoStatePath, false)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func addChildWithFiles(parent *repostate.Node, dirName string, explored, completed bool, files ...string) *repostate.Node {
	parent.AddChildNode(dirName, nil)
	child := parent.Children[dirName]
	child.DoneExploring = explored
	child.Completed = completed
	for _, file := range files {
		child.AddFileName(file)
	}
	return child
}
