package main

import (
	buildInfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferfiles"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferfiles/state"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/reposnapshot"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
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
	if *tests.InstallDataTransferPlugin {
		if *tests.JfrogHome != "" {
			coreutils.ExitOnErr(artifactoryCli.WithoutCredentials().Exec("transfer-plugin-install", inttestutils.SourceServerId, "--home-dir="+*tests.JfrogHome))
		} else {
			coreutils.ExitOnErr(artifactoryCli.WithoutCredentials().Exec("transfer-plugin-install", inttestutils.SourceServerId))
		}
	}

	// Delete the target server if exist
	targetConfig, err := commands.GetConfig(inttestutils.TargetServerId, false)
	if err == nil && targetConfig.ServerId != "" {
		err = configCli.WithoutCredentials().Exec("rm", inttestutils.TargetServerId, "--quiet")
		assert.NoError(t, err)
	}
	*tests.JfrogUrl = utils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	err = tests.NewJfrogCli(execMain, "jfrog config", "--access-token="+*tests.JfrogTargetAccessToken).Exec("add", inttestutils.TargetServerId, "--interactive=false", "--artifactory-url="+*tests.JfrogTargetUrl+tests.ArtifactoryEndpoint)
	assert.NoError(t, err)

	if *tests.InstallDataTransferPlugin {
		assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("curl", "-XPOST", "/api/plugins/reload"))
	}
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

	// Verify again that that files are exist the source Artifactory
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

	// Verify again that that files are exist the source Artifactory
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
	assert.NoError(t, artifactoryCli.Exec("sp", "key1=value1;key2=value2,value3", "--spec="+repo1Spec))

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
	rtVersion, err := getArtifactoryVersion()
	assert.NoError(t, err)
	var moduleType buildInfo.ModuleType
	if rtVersion.AtLeast("7.0.0") {
		// The module type only exist in Artifactory 7
		moduleType = buildInfo.Maven
	}
	validateSpecificModule(publishedBuildInfo.BuildInfo, t, 2, 2, 0, "org.jfrog:cli-test:1.0", moduleType)
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

func TestUnsupportedTransferDirectory(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Mimic the old unsupported transfer directory structure with a joint state.json file.
	transferDir, err := coreutils.GetJfrogTransferDir()
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(transferDir, 0777))
	_, err = os.Create(filepath.Join(transferDir, coreutils.JfrogTransferStateFileName))
	assert.NoError(t, err)

	err = artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1+";"+tests.RtRepo2)
	assert.ErrorContains(t, err, transferfiles.OldTransferDirectoryStructureErrorMsg)
}

func TestTransferWithRepoSnapshot(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	repoSnapshotDir := generateSnapshotFiles(t)

	// Populate source Artifactory.
	repo1Spec, _ := inttestutils.UploadTransferTestFilesAndAssert(artifactoryCli, serverDetails, t)

	// Execute transfer-files.
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1))

	// Assert repo snapshot files were removed after the full transfer was completed.
	empty, err := fileutils.IsDirEmpty(repoSnapshotDir)
	assert.NoError(t, err)
	assert.True(t, empty)

	// Verify again that the files exist the source Artifactory.
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, serverDetails, t)

	// Verify only the expected files were transferred to the target Artifactory:
	// 'a' - only files included in snapshot, because it was explored.
	// 'b' - all files, because it was unexplored.
	// 'c' - no files, because it was completed.
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepoSnapshot(), repo1Spec, targetServerDetails, t)
}

// Creates a new repo snapshot file on the jfrog home, and overrides it with a custom one:
// root - testdata - a -> a1.in
// ---------------------> a3.in
// ----------------- b -> b1.in
// ----------------- c
// 'a' is marked as explored but not completed, we expect its files to be uploaded.
// 'b' is marked as unexplored, we expect its directory to be re-explored and then uploaded.
// 'c' is marked completed, so we expect no action there.
func generateTestRepoSnapshotFile(t *testing.T, repoKey, repoSnapshotFilePath string) {
	snapshotManager := reposnapshot.CreateRepoSnapshotManager(repoKey, repoSnapshotFilePath)
	assert.NotNil(t, snapshotManager)
	root, err := snapshotManager.LookUpNode(".")
	assert.NoError(t, err)
	childTestdata := addChildWithFiles(t, root, "testdata", true, false)
	childA := addChildWithFiles(t, childTestdata, "a", true, false, "a1.in", "a3.in")
	childB := addChildWithFiles(t, childA, "b", false, false, "b1.in")
	_ = addChildWithFiles(t, childB, "c", true, true)

	assert.NoError(t, snapshotManager.PersistRepoSnapshot())
	exists, err := fileutils.IsFileExists(repoSnapshotFilePath, false)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func addChildWithFiles(t *testing.T, parent *reposnapshot.Node, dirName string, explored, checkCompleted bool, files ...string) *reposnapshot.Node {
	childNode := reposnapshot.CreateNewNode(dirName, nil)
	for _, file := range files {
		assert.NoError(t, childNode.AddFile(file, 1))
	}
	childrenPool := make(map[string]*reposnapshot.Node)
	childrenPool[dirName] = childNode
	assert.NoError(t, parent.AddChildNode(dirName, childrenPool))

	if explored {
		assert.NoError(t, childNode.MarkDoneExploring())
	}

	if checkCompleted {
		assert.NoError(t, childNode.CheckCompleted())
	}

	return childNode
}

func generateSnapshotFiles(t *testing.T) (repoSnapshotDir string) {
	// Create new state manager and save.
	stateManager, err := state.NewTransferStateManager(false)
	assert.NoError(t, err)
	assert.NoError(t, stateManager.SetRepoState(tests.RtRepo1, 9, 9, false, true))
	// Set starting time to 10 minutes from now, so that the files diffs phase will not upload the files marked as uploaded by the snapshot.
	assert.NoError(t, stateManager.SetRepoFullTransferStarted(time.Now().Add(10*time.Minute)))
	assert.NoError(t, stateManager.SaveStateAndSnapshots())

	// Copy state file to snapshots directory.
	repoSnapshotDir, err = state.GetJfrogTransferRepoSnapshotDir(tests.RtRepo1)
	assert.NoError(t, err)
	repoState, err := state.GetRepoStateFilepath(tests.RtRepo1, false)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CopyFile(repoSnapshotDir, repoState))

	// Create snapshot file at snapshots directory.
	repoSnapshotFile, err := state.GetRepoSnapshotFilePath(tests.RtRepo1)
	assert.NoError(t, err)
	generateTestRepoSnapshotFile(t, tests.RtRepo1, repoSnapshotFile)
	return
}
