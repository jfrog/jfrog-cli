package main

import (
	"encoding/json"
	biutils "github.com/jfrog/build-info-go/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gocarina/gocsv"
	buildInfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferconfigmerge"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferfiles"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferfiles/state"
	rtUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/reposnapshot"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/access"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/artifactory"
	artifactoryServices "github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
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
	err = tests.NewJfrogCli(execMain, "jfrog config", "--access-token="+*tests.JfrogTargetAccessToken).Exec("add", inttestutils.TargetServerId, "--interactive=false", "--url="+*tests.JfrogTargetUrl)
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

func TestTransferDirProperties(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Create the search spec before we change the working directory
	repo1Spec, err := tests.CreateSpec(tests.SearchRepo1IncludeDirs)
	assert.NoError(t, err)

	// Create temp directory and change the working dir to it
	tmpDir, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, cwd, tmpDir)
	defer chdirCallback()

	// Create an empty folder under tempDir/empty/folder
	assert.NoError(t, os.MkdirAll(filepath.Join("empty", "folder"), 0700))

	// Upload "empty" and "empty/folder" and set properties in the source server
	assert.NoError(t, artifactoryCli.Exec("upload", "empty/*", tests.RtRepo1, "--include-dirs"))
	assert.NoError(t, artifactoryCli.Exec("sp", tests.RtRepo1+"/empty", "a=b", "--include-dirs"))
	assert.NoError(t, artifactoryCli.Exec("sp", tests.RtRepo1+"/empty/folder", "c=d", "--include-dirs"))

	// Execute transfer-files
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1))

	// Verify directories transferred to the target instance
	resultItems, err := inttestutils.SearchInArtifactory(repo1Spec, targetServerDetails, t)
	assert.NoError(t, err)
	assert.Len(t, resultItems, 3)

	// Verify properties
	for _, item := range resultItems {
		switch item.Path {
		case tests.RtRepo1 + "/":
			// Do nothing
		case tests.RtRepo1 + "/empty":
			assert.Equal(t, map[string][]string{"a": {"b"}}, item.Props)
		case tests.RtRepo1 + "/empty/folder":
			assert.Equal(t, map[string][]string{"c": {"d"}}, item.Props)
		default:
			assert.Fail(t, "Unexpected entry", item.Path)
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

func TestUnsupportedRunStatusVersion(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Create run status file with lower version.
	transferDir, err := coreutils.GetJfrogTransferDir()
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(transferDir, 0777))
	statusFilePath := filepath.Join(transferDir, coreutils.JfrogTransferRunStatusFileName)
	trs := state.TransferRunStatus{Version: 0}
	content, err := json.Marshal(trs)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(statusFilePath, content, 0600))

	err = artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1+";"+tests.RtRepo2)
	assert.Equal(t, err, state.GetOldTransferDirectoryStructureError())
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
	// 'a' - all files, because it should be re-explored, even though it was previously explored.
	// 'b' - all files, because it was unexplored.
	// 'c' - no files, because it was completed.
	inttestutils.VerifyExistInArtifactory(tests.GetTransferExpectedRepoSnapshot(), repo1Spec, targetServerDetails, t)
}

// Creates a new repo snapshot file on the jfrog home, and overrides it with a custom one.
// The snapshot tree before persisting is as follows:
// root - testdata - a -> explored, 2 files remaining.
// ----------------- b -> not fully explored, 1 file found.
// ----------------- c -> completed.
// ----------------- deleted-folder -> folder that isn't longer exists in the source Artifactory.
// 'a' is marked as explored but not completed, we expect it to be re-explored and all its files to be uploaded.
// 'b' is marked as unexplored, we expect its directory to be re-explored and then uploaded.
// 'c' is marked completed, so we expect no action there.
// 'deleted-folder' is marked as unexplored however, should not be returned from the AQL and therefore we expect no action there.
func generateTestRepoSnapshotFile(t *testing.T, repoKey, repoSnapshotFilePath string) {
	snapshotManager := reposnapshot.CreateRepoSnapshotManager(repoKey, repoSnapshotFilePath)
	assert.NotNil(t, snapshotManager)
	root, err := snapshotManager.LookUpNode(".")
	assert.NoError(t, err)
	childTestdata := addChildWithFiles(t, root, "testdata", true, false, 0)
	childA := addChildWithFiles(t, childTestdata, "a", true, false, 2)
	childB := addChildWithFiles(t, childA, "b", false, false, 1)
	_ = addChildWithFiles(t, childB, "c", true, true, 0)
	_ = addChildWithFiles(t, childB, "deleted-folder", true, false, 4)

	assert.NoError(t, snapshotManager.PersistRepoSnapshot())
	exists, err := fileutils.IsFileExists(repoSnapshotFilePath, false)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func addChildWithFiles(t *testing.T, parent *reposnapshot.Node, dirName string, explored, checkCompleted bool, filesCount int) *reposnapshot.Node {
	childNode := reposnapshot.CreateNewNode(dirName, nil)
	for i := 0; i < filesCount; i++ {
		assert.NoError(t, childNode.IncrementFilesCount(uint64(i)))
	}

	assert.NoError(t, parent.AddChildNode(dirName, []*reposnapshot.Node{childNode}))

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
	assert.NoError(t, biutils.CopyFile(repoSnapshotDir, repoState))

	// Create snapshot file at snapshots directory.
	repoSnapshotFile, err := state.GetRepoSnapshotFilePath(tests.RtRepo1)
	assert.NoError(t, err)
	generateTestRepoSnapshotFile(t, tests.RtRepo1, repoSnapshotFile)
	return
}

func TestTransferConfigMerge(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	rtVersion, err := getArtifactoryVersion()
	assert.NoError(t, err)
	projectsSupported := false
	if rtVersion.AtLeast("7.0.0") {
		// The module type only exist in Artifactory 7
		projectsSupported = true
	}

	if projectsSupported {
		// Create project on Source server
		deleteProject := createTestProject(t)
		if deleteProject != nil {
			defer func() {
				assert.NoError(t, deleteProject())
			}()
		}
	}
	// The following tests uses DockerRemoteRepo as example repository to test the merge functionality
	// First we remove it from target repository to test that its being transferred from source to target
	targetServicesManager, err := rtUtils.CreateServiceManager(targetServerDetails, -1, 0, false)
	assert.NoError(t, err)
	assert.NoError(t, targetServicesManager.DeleteRepository(tests.DockerRemoteRepo))

	// Execute transfer-config-merge
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-config-merge",
		inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.DockerRemoteRepo, "--include-projects="+tests.ProjectKey))

	// Validate that repository transferred to target:
	targetAuth, err := targetServerDetails.CreateArtAuthConfig()
	assert.NoError(t, err)
	assert.NoError(t, rtUtils.ValidateRepoExists(tests.DockerRemoteRepo, targetAuth))

	// Validate that project transferred to target:
	targetAccessManager, err := rtUtils.CreateAccessServiceManager(targetServerDetails, false)
	assert.NoError(t, err)
	var projectDetails *accessServices.Project
	if projectsSupported {
		projectDetails, err = targetAccessManager.GetProject(tests.ProjectKey)
		if assert.NoError(t, err) && assert.NotNil(t, projectDetails) {
			defer func() {
				assert.NoError(t, targetAccessManager.DeleteProject(tests.ProjectKey))
			}()
		}
	}

	// Validate no conflicts between source and target repositories
	configMergeCmd := transferconfigmerge.NewTransferConfigMergeCommand(serverDetails, targetServerDetails).SetIncludeProjectsPatterns([]string{tests.ProjectKey})
	configMergeCmd.SetIncludeReposPatterns([]string{tests.DockerRemoteRepo})
	csvPath, err := configMergeCmd.Run()
	assert.NoError(t, err)
	assert.Empty(t, csvPath, "No Csv file should be created.")

	// Change repo params on target server
	updateDockerRepoParams(t, targetServicesManager)
	if projectsSupported {
		// Change project params on target server
		updateProjectParams(t, projectDetails, targetAccessManager)
	}
	// Run Config Merge command and expect conflicts
	csvPath, err = configMergeCmd.Run()
	assert.NoError(t, err)
	validateCsvConflicts(t, csvPath, projectsSupported)
}

func updateDockerRepoParams(t *testing.T, targetServicesManager artifactory.ArtifactoryServicesManager) {
	params := artifactoryServices.DockerRemoteRepositoryParams{}
	assert.NoError(t, targetServicesManager.GetRepository(tests.DockerRemoteRepo, &params))
	// Exactly 22 field changes:
	params.BlackedOut = inverseBooleanPointer(params.BlackedOut)
	params.XrayIndex = inverseBooleanPointer(params.XrayIndex)
	params.PriorityResolution = inverseBooleanPointer(params.PriorityResolution)
	params.ExternalDependenciesEnabled = inverseBooleanPointer(params.ExternalDependenciesEnabled)
	params.EnableTokenAuthentication = inverseBooleanPointer(params.EnableTokenAuthentication)
	params.BlockPushingSchema1 = inverseBooleanPointer(params.BlockPushingSchema1)
	params.HardFail = inverseBooleanPointer(params.HardFail)
	params.Offline = inverseBooleanPointer(params.Offline)
	params.ShareConfiguration = inverseBooleanPointer(params.ShareConfiguration)
	params.SynchronizeProperties = inverseBooleanPointer(params.SynchronizeProperties)
	params.BlockMismatchingMimeTypes = inverseBooleanPointer(params.BlockMismatchingMimeTypes)
	params.AllowAnyHostAuth = inverseBooleanPointer(params.AllowAnyHostAuth)
	params.EnableCookieManagement = inverseBooleanPointer(params.EnableCookieManagement)
	params.BypassHeadRequests = inverseBooleanPointer(params.BypassHeadRequests)
	*params.SocketTimeoutMillis += 100
	*params.RetrievalCachePeriodSecs += 100
	*params.MetadataRetrievalTimeoutSecs += 100
	*params.MissedRetrievalCachePeriodSecs += 100
	*params.UnusedArtifactsCleanupPeriodHours += 100
	*params.AssumedOfflinePeriodSecs += 100
	params.Username = "test123"
	params.ContentSynchronisation.Enabled = inverseBooleanPointer(params.ContentSynchronisation.Enabled)

	assert.NoError(t, targetServicesManager.UpdateRemoteRepository().Docker(params))
}

func inverseBooleanPointer(boolPtr *bool) *bool {
	boolValue := true
	if boolPtr != nil && *boolPtr {
		boolValue = false
	}
	return &boolValue
}

func validateCsvConflicts(t *testing.T, csvPath string, projectsSupported bool) {
	if assert.NotEmpty(t, csvPath) {
		createdFile, err := os.Open(csvPath)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, createdFile.Close())
		}()
		conflicts := new([]transferconfigmerge.Conflict)
		assert.NoError(t, gocsv.UnmarshalFile(createdFile, conflicts))

		if projectsSupported {
			// Verify project conflict
			projectConflict := (*conflicts)[0]
			assert.Equal(t, projectConflict.Type, transferconfigmerge.Project)
			assert.Len(t, strings.Split(projectConflict.DifferentProperties, ";"), 4)
		}

		// Verify repo conflict
		repoConflict := (*conflicts)[len(*conflicts)-1]
		assert.Equal(t, repoConflict.Type, transferconfigmerge.Repository)
		assert.Equal(t, repoConflict.SourceName, tests.DockerRemoteRepo)
		assert.Equal(t, repoConflict.TargetName, tests.DockerRemoteRepo)
		assert.Len(t, strings.Split(repoConflict.DifferentProperties, ";"), 22)
	}
}

func createTestProject(t *testing.T) func() error {
	accessManager, err := rtUtils.CreateAccessServiceManager(serverDetails, false)
	assert.NoError(t, err)
	// Delete the project if already exists
	deleteProjectIfExists(t, accessManager, tests.ProjectKey)

	// Create new project
	adminPrivileges := accessServices.AdminPrivileges{
		ManageMembers:   utils.Pointer(false),
		ManageResources: utils.Pointer(false),
		IndexResources:  utils.Pointer(false),
	}
	projectDetails := accessServices.Project{
		DisplayName:       tests.ProjectKey + "MyProject",
		Description:       "My Test Project",
		AdminPrivileges:   &adminPrivileges,
		SoftLimit:         utils.Pointer(false),
		StorageQuotaBytes: 1073741825,
		ProjectKey:        tests.ProjectKey,
	}

	if assert.NoError(t, accessManager.CreateProject(accessServices.ProjectParams{ProjectDetails: projectDetails})) {
		return func() error {
			return accessManager.DeleteProject(tests.ProjectKey)
		}
	}
	return nil
}

func updateProjectParams(t *testing.T, projectParams *accessServices.Project, targetAccessManager *access.AccessServicesManager) {
	projectParams.Description = "123123123123"
	projectParams.AdminPrivileges.IndexResources = utils.Pointer(true)
	projectParams.SoftLimit = utils.Pointer(true)
	projectParams.StorageQuotaBytes += 1
	assert.NoError(t, targetAccessManager.UpdateProject(accessServices.ProjectParams{ProjectDetails: *projectParams}))
}
