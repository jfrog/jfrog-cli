package inttestutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	buildInfoGoUtils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferfiles"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

const (
	SourceServerId  = "default"
	TargetServerId  = "transfer-test-target"
	dataTransferUrl = "https://releases.jfrog.io/artifactory/jfrog-releases/data-transfer/[RELEASE]/"
	groovyFileName  = "dataTransfer.groovy"
	jarFileName     = "data-transfer.jar"
)

// Create test repositories in the target Artifactory
// targetArtifactoryCli - Target Artifactory CLI
func CreateTargetRepos(targetArtifactoryCli *tests.JfrogCli) {
	log.Info("Creating repositories in target Artifactory...")
	for _, template := range tests.CreatedNonVirtualRepositories {
		repoTemplate := filepath.Join("testdata", template)
		templateVars := fmt.Sprintf("--vars=REPO1=%s;REPO2=%s;MAVEN_REPO1=%s;MAVEN_REMOTE_REPO=%s",
			tests.RtRepo1, tests.RtRepo2, tests.MvnRepo1, tests.MvnRemoteRepo)
		coreutils.ExitOnErr(targetArtifactoryCli.Exec("repo-create", repoTemplate, templateVars))
	}
}

// Delete test repositories in the target Artifactory
// targetArtifactoryCli - Target Artifactory CLI
func DeleteTargetRepos(targetArtifactoryCli *tests.JfrogCli) {
	for repoKey := range tests.CreatedNonVirtualRepositories {
		coreutils.ExitOnErr(targetArtifactoryCli.Exec("repo-delete", *repoKey))
	}
}

// Clean test repositories content in the target Artifactory
// targetArtifactoryCli - Target Artifactory CLI
func CleanTargetRepos(targetArtifactoryCli *tests.JfrogCli) {
	for repoKey := range tests.CreatedNonVirtualRepositories {
		coreutils.ExitOnErr(targetArtifactoryCli.Exec("del", *repoKey))
	}
}

// Install data-transfer Artifactory user plugin
func InstallDataTransferPlugin() {
	pluginsDir := filepath.Join(*tests.JfrogHome, "artifactory", "var", "etc", "artifactory", "plugins")
	groovyFile := filepath.Join(pluginsDir, groovyFileName)
	err := buildInfoGoUtils.DownloadFile(groovyFile, dataTransferUrl+groovyFileName)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	libDir := filepath.Join(pluginsDir, "lib")
	fileutils.CreateDirIfNotExist(libDir)
	jarFile := filepath.Join(libDir, jarFileName)
	err = buildInfoGoUtils.DownloadFile(jarFile, dataTransferUrl+"lib/"+jarFileName)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

// Authenticate target Artifactory using the input flags
func AuthenticateTarget() (string, *config.ServerDetails, *httputils.HttpClientDetails) {
	*tests.JfrogTargetUrl = clientutils.AddTrailingSlashIfNeeded(*tests.JfrogTargetUrl)
	serverDetails := &config.ServerDetails{
		Url:            *tests.JfrogTargetUrl,
		ArtifactoryUrl: *tests.JfrogTargetUrl + tests.ArtifactoryEndpoint,
		AccessToken:    *tests.JfrogTargetAccessToken,
	}
	cred := "--url=" + serverDetails.ArtifactoryUrl + " --access-token=" + serverDetails.AccessToken
	serviceDetails, err := serverDetails.CreateArtAuthConfig()
	if err != nil {
		coreutils.ExitOnErr(errors.New("Failed while attempting to authenticate with Artifactory: " + err.Error()))
	}

	artAuthDetails := serviceDetails.CreateHttpClientDetails()
	return cred, serverDetails, &artAuthDetails
}

// Upload testdata/a/* to repo1 and testdata/a/b/* to repo2.
// sourceArtifactoryCli - Source Artifactory CLI
// serverDetails - Source server details
// t - The testing object
func UploadTransferTestFilesAndAssert(sourceArtifactoryCli *tests.JfrogCli, serverDetails *config.ServerDetails, t *testing.T) (string, string) {
	// Upload files
	assert.NoError(t, sourceArtifactoryCli.Exec("upload", "testdata/a/*", tests.RtRepo1))
	assert.NoError(t, sourceArtifactoryCli.Exec("upload", "testdata/a/b/*", tests.RtRepo2))

	// Verify files were uploaded to the source Artifactory
	repo1Spec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	VerifyExistInArtifactory(tests.GetTransferExpectedRepo1(), repo1Spec, serverDetails, t)
	repo2Spec, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	VerifyExistInArtifactory(tests.GetTransferExpectedRepo2(), repo2Spec, serverDetails, t)

	return repo1Spec, repo2Spec
}

// Return the number of artifacts in the given pattern
// pattern - Search wildcard pattern
// serverDetails - The Artifactory server details
// t - The testing object
func CountArtifactsInPath(pattern string, serverDetails *config.ServerDetails, t *testing.T) int {
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(spec.NewBuilder().Pattern(pattern).BuildSpec())
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	defer assert.NoError(t, reader.Close())
	length, err := reader.Length()
	assert.NoError(t, err)
	return length
}

// Wait for a metadata file to automatically generated in Artifactory
// pattern - The search pattern
// serverDetails - The Artifactory server details
// t - The testing object
func WaitForCreationInArtifactory(pattern string, serverDetails *config.ServerDetails, t *testing.T) {
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(spec.NewBuilder().Pattern(pattern).BuildSpec())
	for i := 0; i < 20; i++ {
		reader, err := searchCmd.Search()
		assert.NoError(t, err)
		defer assert.NoError(t, reader.Close())
		if !reader.IsEmpty() {
			return
		}
		time.Sleep(5 * time.Second)
	}
	assert.Fail(t, "Couldn't find in target Artifactory: "+pattern)
}

// Asynchronously execute transfer-files
// artifactoryCli - Source Artifactory CLI
// wg - Wait group to update when done
// filesTransferFinished - Changes to true when the file transfer process finished
// t - The testing object
func AsyncExecTransferFiles(artifactoryCli *tests.JfrogCli, wg *sync.WaitGroup, filesTransferFinished *bool, t *testing.T) {
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			*filesTransferFinished = true
		}()
		// Execute transfer-files
		assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("transfer-files", "default", TargetServerId, "--include-repos="+tests.RtRepo1))
	}()
}

// Asynchronously update the number of threads in transfer-files command
// wg - Wait group to update when done
// filesTransferFinished - Changes to true when the file transfer process finished
// t - The testing object
func AsyncUpdateThreads(wg *sync.WaitGroup, filesTransferFinished *bool, t *testing.T) {
	wg.Add(1)
	go func() {
		threadsCount := 0
		defer wg.Done()

		// Wait for the number of threads to be updated to the non-zero default and increace the number of threads by 1
		for !*filesTransferFinished {
			threadsCount = transferfiles.GetThreads()
			if threadsCount == 0 {
				time.Sleep(time.Second)
				continue
			}
			conf := &utils.TransferSettings{ThreadsNumber: threadsCount + 1}
			assert.NoError(t, utils.SaveTransferSettings(conf))
			break
		}

		// Wait for the number of threads to be increase by 1
		for !*filesTransferFinished {
			if transferfiles.GetThreads() == threadsCount+1 {
				return
			}
			time.Sleep(time.Second)
		}

		// If false, the transfer-files process is finished before the threads changed
		assert.Failf(t, "Number of threads did not change as expected", "Actual: %d, Expected: %d", transferfiles.GetThreads(), threadsCount+1)
	}()
}
