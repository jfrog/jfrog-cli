package main

import (
	"strconv"
	"sync"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferfiles"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
)

var targetArtifactoryCli *tests.JfrogCli
var targetServerDetails *config.ServerDetails

func InitTransferTests() {
	initArtifactoryCli()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
	inttestutils.InstallDataTransferPlugin()
	var creds string
	creds, targetServerDetails = inttestutils.AuthenticateTarget()
	targetArtifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", creds)
	inttestutils.CreateTargetRepos(targetArtifactoryCli)
	inttestutils.RefreshStorageInfoAndWait(serverDetails)
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
	config, err := commands.GetConfig(inttestutils.TargetServerId, false)
	if err == nil && config.ServerId != "" {
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

	// Upload files
	runRt(t, "upload", "testdata/a/*", tests.RtRepo1)
	runRt(t, "upload", "testdata/a/b/*", tests.RtRepo2)

	// Verify files were uploaded to the source Artifactory
	repo1Spec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, t)
	repo2Spec, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetTransferExpectedRepo2(), repo2Spec, t)

	// Execute transfer-files
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1+";"+tests.RtRepo2))

	// Verify again that that files are exist the source Artifactory
	verifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, t)
	verifyExistInArtifactory(tests.GetTransferExpectedRepo2(), repo2Spec, t)

	// Verify files were transferred to the target Artifactory
	inttestutils.VerifyExistInTargetArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, targetServerDetails, t)
	inttestutils.VerifyExistInTargetArtifactory(tests.GetTransferExpectedRepo2(), repo2Spec, targetServerDetails, t)
}

func TestTransferExcludeRepo(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	// Upload files
	runRt(t, "upload", "testdata/a/*", tests.RtRepo1)
	runRt(t, "upload", "testdata/a/b/*", tests.RtRepo2)

	// Verify files were uploaded to the source Artifactory
	repo1Spec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, t)
	repo2Spec, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetTransferExpectedRepo2(), repo2Spec, t)

	// Execute transfer-files
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.RtRepo1+";"+tests.RtRepo2, "--exclude-repos="+tests.RtRepo2))

	// Verify repo1 files were transferred to the target Artifactory
	inttestutils.VerifyExistInTargetArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, targetServerDetails, t)
	inttestutils.VerifyExistInTargetArtifactory([]string{}, repo2Spec, targetServerDetails, t)
}

func TestTransferMaven(t *testing.T) {
	cleanUp := initTransferTest(t)
	defer cleanUp()

	assert.NoError(t, runMaven(t, createSimpleMavenProject, tests.MavenConfig, "install"))

	// Verify files were uploaded to the source Artifactory
	mvnRepoSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), mvnRepoSpec, t)

	// Execute transfer-files
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", inttestutils.SourceServerId, inttestutils.TargetServerId, "--include-repos="+tests.MvnRepo1))

	// Verify maven files were transferred to the target Artifactory
	inttestutils.VerifyExistInTargetArtifactory(tests.GetMavenDeployedArtifacts(), mvnRepoSpec, targetServerDetails, t)

	// Wait for creation of maven-metadata.xml in the target Artifactory
	inttestutils.WaitForCreationInArtifactory(tests.MvnRepo1+"/org/jfrog/cli-test/maven-metadata.xml", targetServerDetails, t)
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
