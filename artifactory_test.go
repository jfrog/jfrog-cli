package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/buger/jsonparser"
	gofrogio "github.com/jfrog/gofrog/io"
	"github.com/jfrog/gofrog/version"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	cliproxy "github.com/jfrog/jfrog-cli/utils/tests/proxy/server"
	"github.com/jfrog/jfrog-cli/utils/tests/proxy/server/certificate"
	"github.com/jfrog/jfrog-client-go/access"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	rtutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils/tests/xray"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Access does not support creating an admin token without UI. Skipping projects tests till this functionality will be implemented.
// https://jira.jfrog.org/browse/JA-2620
const projectsTokenMinArtifactoryVersion = "7.41.0"

// Minimum Artifactory version with Terraform support
const terraformMinArtifactoryVersion = "7.38.4"

// JFrog CLI for Artifactory sub-commands (jfrog rt ...)
var artifactoryCli *tests.JfrogCli

// JFrog CLI for Platform commands (jfrog ...)
var platformCli *tests.JfrogCli

// JFrog CLI for config command only (doesn't pass the --ssh-passphrase flag)
var configCli *tests.JfrogCli

var serverDetails *config.ServerDetails
var artAuth auth.ServiceDetails
var artHttpDetails httputils.HttpClientDetails

// Run `jfrog rt` command
func runRt(t *testing.T, args ...string) {
	assert.NoError(t, artifactoryCli.Exec(args...))
}

func InitArtifactoryTests() {
	initArtifactoryCli()
	cleanUpOldBuilds()
	cleanUpOldRepositories()
	cleanUpOldUsers()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
	cleanArtifactoryTest()
}

func authenticate(configCli bool) string {
	*tests.JfrogUrl = clientutils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	serverDetails = &config.ServerDetails{Url: *tests.JfrogUrl, ArtifactoryUrl: *tests.JfrogUrl + tests.ArtifactoryEndpoint, SshKeyPath: *tests.JfrogSshKeyPath, SshPassphrase: *tests.JfrogSshPassphrase}
	var cred string
	if configCli {
		cred += "--artifactory-url=" + serverDetails.ArtifactoryUrl
	} else {
		cred += "--url=" + serverDetails.ArtifactoryUrl
	}
	if !fileutils.IsSshUrl(serverDetails.ArtifactoryUrl) {
		if *tests.JfrogAccessToken != "" {
			serverDetails.AccessToken = *tests.JfrogAccessToken
		} else {
			serverDetails.User = *tests.JfrogUser
			serverDetails.Password = *tests.JfrogPassword
		}
	}
	cred += getArtifactoryTestCredentials()
	var err error
	if artAuth, err = serverDetails.CreateArtAuthConfig(); err != nil {
		coreutils.ExitOnErr(errors.New("Failed while attempting to authenticate with Artifactory: " + err.Error()))
	}
	serverDetails.ArtifactoryUrl = artAuth.GetUrl()
	serverDetails.SshUrl = artAuth.GetSshUrl()
	serverDetails.AccessUrl = clientutils.AddTrailingSlashIfNeeded(*tests.JfrogUrl) + tests.AccessEndpoint
	artHttpDetails = artAuth.CreateHttpClientDetails()
	return cred
}

// A Jfrog CLI to be used to execute a config task.
// Removed the ssh-passphrase flag that cannot be passed to with a config command
func createConfigJfrogCLI(cred string) *tests.JfrogCli {
	if strings.Contains(cred, " --ssh-passphrase=") {
		cred = strings.Replace(cred, " --ssh-passphrase="+*tests.JfrogSshPassphrase, "", -1)
	}
	return tests.NewJfrogCli(execMain, "jfrog config", cred)
}

func getArtifactoryTestCredentials() string {
	if fileutils.IsSshUrl(serverDetails.ArtifactoryUrl) {
		return getSshCredentials()
	}
	if *tests.JfrogAccessToken != "" {
		return " --access-token=" + *tests.JfrogAccessToken
	}
	return " --user=" + *tests.JfrogUser + " --password=" + *tests.JfrogPassword
}

func getSshCredentials() string {
	cred := ""
	if *tests.JfrogSshKeyPath != "" {
		cred += " --ssh-key-path=" + *tests.JfrogSshKeyPath
	}
	if *tests.JfrogSshPassphrase != "" {
		cred += " --ssh-passphrase=" + *tests.JfrogSshPassphrase
	}
	return cred
}

func TestArtifactorySimpleUploadSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	specFile, err := tests.CreateSpec(tests.UploadFlatRecursive)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadExpectedRepo1(), searchFilePath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactorySimpleUploadWithWildcardSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	// Init tmp dir
	specFile, err := tests.CreateSpec(tests.UploadTempWildcard)
	assert.NoError(t, err)
	err = fileutils.CopyDir(tests.GetTestResourcesPath()+"cache", filepath.Dir(specFile), true, nil)
	assert.NoError(t, err)
	// Upload
	runRt(t, "upload", "--spec="+specFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSimpleWildcardUploadExpectedRepo1(), searchFilePath, serverDetails, t)
	cleanArtifactoryTest()
}

// This test is similar to TestArtifactorySimpleUploadSpec but using "--server-id" flag
func TestArtifactorySimpleUploadSpecUsingConfig(t *testing.T) {
	initArtifactoryTest(t, "")
	passphrase, err := createServerConfigAndReturnPassphrase(t)
	defer deleteServerConfig(t)
	assert.NoError(t, err)
	artifactoryCommandExecutor := tests.NewJfrogCli(execMain, "jfrog rt", "")
	specFile, err := tests.CreateSpec(tests.UploadFlatRecursive)
	assert.NoError(t, err)
	assert.NoError(t, artifactoryCommandExecutor.Exec("upload", "--spec="+specFile, "--server-id="+tests.ServerId, passphrase))

	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadExpectedRepo1(), searchFilePath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadPathWithSpecialCharsAsNoRegex(t *testing.T) {
	initArtifactoryTest(t, "")
	filePath := getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1, "--flat")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryIgnoreChecksumOnFilteredResource(t *testing.T) {
	initArtifactoryTest(t, "")
	filePath := "testdata/filtered/a.txt"

	// Upload a filtered resource
	runRt(t, "upload", filePath, tests.RtRepo1, "--flat", "--target-props", "artifactory.filtered=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadFilteredRepo1(), searchFilePath, serverDetails, t)

	// Discard output logging to prevent negative logs
	previousLogger := tests.RedirectLogOutputToNil()
	defer log.SetLogger(previousLogger)

	// Attempt to download the filtered resource (this will fail the checksum check)
	err = artifactoryCli.Exec([]string{"download", tests.RtRepo1 + "/a.txt", tests.Out + "/mypath2/filtered.txt"}...)
	assert.Error(t, err)

	// Attempt to download the filtered resource but skip the checksum
	runRt(t, "download", tests.RtRepo1+"/a.txt", tests.Out+"/mypath2/filtered.txt", "--skip-checksum")

	cleanArtifactoryTest()
}

func TestArtifactoryEmptyBuild(t *testing.T) {
	initArtifactoryTest(t, "")
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	buildNumber := "5"

	// Try to upload with non-existent pattern
	runRt(t, "upload", "*.notExist", tests.RtRepo1, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)

	// Try to download with non-existent pattern
	runRt(t, "download", tests.RtRepo1+"/*.notExist", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)

	// Publish empty build info
	runRt(t, "build-publish", tests.RtBuildName1, buildNumber)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryPublishBuildUsingBuildFile(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	// Create temp folder.
	tmpDir, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	// Create build config in temp folder
	_, err := tests.ReplaceTemplateVariables(filepath.Join("testdata", "buildspecs", "build.yaml"), filepath.Join(tmpDir, ".jfrog", "projects"))
	assert.NoError(t, err)

	// Run cd command to temp dir.
	wdCopy, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wdCopy, tmpDir)
	defer chdirCallback()
	// Upload file to create build-info data using the build.yaml file.
	runRt(t, "upload", filepath.Join(wdCopy, "testdata", "a", "a1.in"), tests.RtRepo1+"/foo")

	// Publish build-info using the build.yaml file.
	runRt(t, "build-publish")

	// Search artifacts based on the published build.
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).Build(tests.RtBuildName1 + "/1")
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())

	// Validate the search result.
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	searchResultLength, err := reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, 1, searchResultLength)

	// Upload file to create a second build-info data.
	runRt(t, "upload", filepath.Join(wdCopy, "testdata", "a", "a1.in"), tests.RtRepo1+"/bla-bla")

	// Publish the second build-info build.yaml file.
	runRt(t, "build-publish")

	// Search artifacts based on the second published build.
	searchSpecBuilder = spec.NewBuilder().Pattern(tests.RtRepo1).Build(tests.RtBuildName1 + "/2")
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())

	// Validate the search result.
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	searchResultLength, err = reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, 1, searchResultLength)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
}

func TestArtifactoryDownloadFromVirtual(t *testing.T) {
	initArtifactoryTest(t, "")

	runRt(t, "upload", "testdata/a/*", tests.RtRepo1, "--flat=false")
	runRt(t, "dl", tests.RtVirtualRepo+"/testdata/(*)", tests.Out+"/"+"{1}", "--flat=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally(tests.GetVirtualDownloadExpected(), paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndUploadWithProgressBar(t *testing.T) {
	initArtifactoryTest(t, "")

	callback := tests.MockProgressInitialization()
	defer callback()

	runRt(t, "upload", "testdata/a/*", tests.RtRepo1, "--flat=false")
	runRt(t, "dl", tests.RtVirtualRepo+"/testdata/(*)", tests.Out+"/"+"{1}", "--flat=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally(tests.GetVirtualDownloadExpected(), paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadPathWithSpecialChars(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", getSpecialCharFilePath(), tests.RtRepo1, "--flat=false")
	runRt(t, "upload", "testdata/c#/a#1.in", tests.RtRepo1, "--flat=false")

	runRt(t, "dl", tests.RtRepo1+"/testdata/a$+~&^a#/a*", tests.Out+fileutils.GetFileSeparator(), "--flat=true")
	runRt(t, "dl", tests.RtRepo1+"/testdata/c#/a#1.in", tests.Out+fileutils.GetFileSeparator(), "--flat=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a1.in"), filepath.Join(tests.Out, "a#1.in")}, paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadPatternWithUnicodeChars(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", "testdata/unicode/", tests.RtRepo1, "--flat=false")

	// Verify files exist
	specFile, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetDownloadUnicode(), specFile, serverDetails, t)

	runRt(t, "dl", tests.RtRepo1+"/testdata/unicode/(dirλrectory)/", filepath.Join(tests.Out, "{1}")+fileutils.GetFileSeparator(), "--flat=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical([]string{
		tests.Out,
		filepath.Join(tests.Out, "dirλrectory"),
		filepath.Join(tests.Out, "dirλrectory", "文件.in"),
		filepath.Join(tests.Out, "dirλrectory", "aȩ.ȥ1"),
	}, paths)
	assert.NoError(t, err)

	cleanArtifactoryTest()
}

func TestArtifactoryConcurrentDownload(t *testing.T) {
	testArtifactoryDownload(cliutils.DownloadMinSplitKb*1000, t)
}

func TestArtifactoryBulkDownload(t *testing.T) {
	testArtifactoryDownload(cliutils.DownloadMinSplitKb*1000-1, t)
}

func testArtifactoryDownload(fileSize int, t *testing.T) {
	initArtifactoryTest(t, "")
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)
	randFile, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "randFile"), fileSize)
	assert.NoError(t, err)
	localFileDetails, err := fileutils.GetFileDetails(randFile.Name(), true)
	assert.NoError(t, err)

	runRt(t, "u", randFile.Name(), tests.RtRepo1+"/testdata/", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	runRt(t, "dl", tests.RtRepo1+"/testdata/", tests.Out+fileutils.GetFileSeparator(), "--flat=true")

	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "randFile")}, paths, t)
	tests.ValidateChecksums(filepath.Join(tests.Out, "randFile"), localFileDetails.Checksum, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadWildcardInRepo(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	// Upload a file to repo1 and another one to repo2
	runRt(t, "upload", filePath, tests.RtRepo1+"/path/a1.in")
	runRt(t, "upload", filePath, tests.RtRepo2+"/path/a2.in")

	specFile, err := tests.CreateSpec(tests.DownloadWildcardRepo)
	assert.NoError(t, err)

	// Verify the 2 files exist using `*` in the repository name
	inttestutils.VerifyExistInArtifactory(tests.GetDownloadWildcardRepo(), specFile, serverDetails, t)

	// Download the 2 files with `*` in the repository name
	runRt(t, "dl", "--spec="+specFile)
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a1.in"), filepath.Join(tests.Out, "a2.in")}, paths, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadPlaceholderInRepo(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	// Upload a file to repo1 and another one to repo2
	runRt(t, "upload", filePath, tests.RtRepo1+"/path/a1.in")
	runRt(t, "upload", filePath, tests.RtRepo2+"/path/a2.in")

	specFile, err := tests.CreateSpec(tests.DownloadWildcardRepo)
	assert.NoError(t, err)

	// Verify the 2 files exist
	inttestutils.VerifyExistInArtifactory(tests.GetDownloadWildcardRepo(), specFile, serverDetails, t)

	// Download the 2 files with placeholders in the repository name
	runRt(t, "dl", tests.RtRepo1And2Placeholder, tests.Out+"/a/{1}/", "--flat=true")
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a", "1", "a1.in"), filepath.Join(tests.Out, "a", "2", "a2.in")}, paths, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadPlaceholderFlat(t *testing.T) {
	initArtifactoryTest(t, "")
	for _, flatValue := range []string{"true", "false"} {
		runRt(t, "upload", "testdata/a/b/(*)", tests.RtRepo1+"/path/{1}", "--flat="+flatValue)
		searchPath, err := tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolder(), searchPath, serverDetails, t)
		runRt(t, "del", tests.RtRepo1+"/*")

		runRt(t, "upload", "testdata/a/b/(*)", tests.RtRepo1+"/path/{1}/", "--flat="+flatValue)
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolderlashSlashSuffix(), searchPath, serverDetails, t)
		runRt(t, "del", tests.RtRepo1+"/*")

		runRt(t, "upload", "testdata/a/b/(*)/(*)", tests.RtRepo1+"/path/{1}/{2}", "--flat="+flatValue)
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetUploadedFileWithDownloadedDoublePlaceHolder(), searchPath, serverDetails, t)
		runRt(t, "del", tests.RtRepo1+"/*")
	}
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadWithPlaceholderFlat(t *testing.T) {
	initArtifactoryTest(t, "")
	// Upload test data to Artifactory
	runRt(t, "upload", "testdata/a/b/(*)", tests.RtRepo1+"/path/{1}")
	// Download the tests data using placeholder with flat
	for _, flatValue := range []string{"true", "false"} {
		runRt(t, "download", tests.RtRepo1+"/path/(*)", tests.Out+"/mypath2/{1}", "--flat="+flatValue)
		paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
		assert.NoError(t, err)
		checkSyncedDirContent(tests.GetFileWithDownloadedPlaceHolder(), paths, t)
		clientTestUtils.RemoveAllAndAssert(t, tests.Out)

		runRt(t, "download", tests.RtRepo1+"/path/(*)", tests.Out+"/mypath2/{1}/", "--flat="+flatValue)
		paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
		assert.NoError(t, err)
		checkSyncedDirContent(tests.GetFileWithDownloadedPlaceHolderSlashSuffix(), paths, t)
		clientTestUtils.RemoveAllAndAssert(t, tests.Out)

		runRt(t, "download", tests.RtRepo1+"/path/(*)/(*)", tests.Out+"/mypath2/{1}/{2}", "--flat="+flatValue)
		paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
		assert.NoError(t, err)
		checkSyncedDirContent(tests.GetFileWithDownloadedDoublePlaceHolder(), paths, t)
		clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	}
	cleanArtifactoryTest()
}

func TestArtifactoryCopyWithPlaceholderFlat(t *testing.T) {
	initArtifactoryTest(t, "")
	// Upload test data to Artifactory
	runRt(t, "upload", "testdata/a/b/(*)", tests.RtRepo2+"/mypath2/{1}")
	// Download the tests data using placeholder with flat
	for _, flatValue := range []string{"true", "false"} {
		runRt(t, "copy", tests.RtRepo2+"/mypath2/(*)", tests.RtRepo1+"/path/{1}", "--flat="+flatValue)
		searchPath, err := tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolder(), searchPath, serverDetails, t)
		runRt(t, "del", tests.RtRepo1+"/*")

		runRt(t, "copy", tests.RtRepo2+"/mypath2/(*)", tests.RtRepo1+"/path/{1}/", "--flat="+flatValue)
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolderlashSlashSuffix(), searchPath, serverDetails, t)
		runRt(t, "del", tests.RtRepo1+"/*")

		runRt(t, "copy", tests.RtRepo2+"/mypath2/(*)/(*)", tests.RtRepo1+"/path/{1}/{2}", "--flat="+flatValue)
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetUploadedFileWithDownloadedDoublePlaceHolder(), searchPath, serverDetails, t)
		runRt(t, "del", tests.RtRepo1+"/*")
	}
	cleanArtifactoryTest()
}

func TestArtifactoryMoveWithPlaceholderFlat(t *testing.T) {
	initArtifactoryTest(t, "")
	// Download the tests data using placeholder with flat
	for _, flatValue := range []string{"true", "false"} {
		runRt(t, "upload", "testdata/a/b/(*)", tests.RtRepo2+"/mypath2/{1}")
		runRt(t, "move", tests.RtRepo2+"/mypath2/(*)", tests.RtRepo1+"/path/{1}", "--flat="+flatValue)
		searchPath, err := tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolder(), searchPath, serverDetails, t)
		runRt(t, "del", tests.RtRepo1+"/*")
		runRt(t, "del", tests.RtRepo2+"/*")

		runRt(t, "upload", "testdata/a/b/(*)", tests.RtRepo2+"/mypath2/{1}")
		runRt(t, "move", tests.RtRepo2+"/mypath2/(*)", tests.RtRepo1+"/path/{1}/", "--flat="+flatValue)
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolderlashSlashSuffix(), searchPath, serverDetails, t)
		runRt(t, "del", tests.RtRepo1+"/*")
		runRt(t, "del", tests.RtRepo2+"/*")

		runRt(t, "upload", "testdata/a/b/(*)", tests.RtRepo2+"/mypath2/{1}")
		runRt(t, "move", tests.RtRepo2+"/mypath2/(*)/(*)", tests.RtRepo1+"/path/{1}/{2}", "--flat="+flatValue)
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetUploadedFileWithDownloadedDoublePlaceHolder(), searchPath, serverDetails, t)
		runRt(t, "del", tests.RtRepo1+"/*")
		runRt(t, "del", tests.RtRepo2+"/*")
	}
	cleanArtifactoryTest()
}

func TestArtifactoryCopySingleFileNonFlat(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1+"/path/", "--flat")
	runRt(t, "cp", tests.RtRepo1+"/path/a1.in", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyPrefixFilesFlat(t *testing.T) {
	initArtifactoryTest(t, "")

	runRt(t, "upload", "testdata/prefix/(*)", tests.RtRepo1+"/prefix/prefix-{1}")
	runRt(t, "cp", tests.RtRepo1+"/prefix/*", tests.RtRepo2, "--flat")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetPrefixFilesCopy(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestAqlFindingItemOnRoot(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1+"/path/inner/", "--flat=true")
	runRt(t, "upload", filePath, tests.RtRepo1+"/someFile", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/*", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetAnyItemCopy(), searchPath, serverDetails, t)
	runRt(t, "del", tests.RtRepo2+"/*")
	runRt(t, "del", tests.RtRepo1+"/*")
	runRt(t, "upload", filePath, tests.RtRepo1+"/path/", "--flat=true")
	runRt(t, "upload", filePath, tests.RtRepo1+"/someFile", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/*/", tests.RtRepo2)
	inttestutils.VerifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestExitCode(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload dummy file in order to test move and copy commands
	runRt(t, "upload", path.Join("testdata", "a", "a1.in"), tests.RtRepo1)

	// Discard output logging to prevent negative logs
	previousLogger := tests.RedirectLogOutputToNil()
	defer log.SetLogger(previousLogger)
	defer clientTestUtils.UnSetEnvAndAssert(t, coreutils.FailNoOp)

	// Test upload commands
	err := artifactoryCli.Exec("upload", "DummyText", tests.RtRepo1, "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeError, err)
	err = artifactoryCli.Exec("upload", path.Join("testdata", "a", "a1.in"), "DummyTargetPath")
	checkExitCode(t, coreutils.ExitCodeError, err)
	err = artifactoryCli.Exec("upload", "testdata/a/(*.dummyExt)", tests.RtRepo1+"/{1}.in", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Setenv(coreutils.FailNoOp, "true"))
	err = artifactoryCli.Exec("upload", "testdata/a/(*.dummyExt)", tests.RtRepo1+"/{1}.in")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Unsetenv(coreutils.FailNoOp))

	// Test download command
	err = artifactoryCli.Exec("dl", "DummyFolder", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Setenv(coreutils.FailNoOp, "true"))
	err = artifactoryCli.Exec("dl", "DummyFolder")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Unsetenv(coreutils.FailNoOp))

	// Test move commands
	err = artifactoryCli.Exec("move", tests.RtRepo1, "DummyTargetPath")
	checkExitCode(t, coreutils.ExitCodeError, err)
	err = artifactoryCli.Exec("move", "DummyText", tests.RtRepo1, "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Setenv(coreutils.FailNoOp, "true"))
	err = artifactoryCli.Exec("move", "DummyText", tests.RtRepo1)
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Unsetenv(coreutils.FailNoOp))

	// Test copy commands
	err = artifactoryCli.Exec("copy", tests.RtRepo1, "DummyTargetPath")
	checkExitCode(t, coreutils.ExitCodeError, err)
	err = artifactoryCli.Exec("copy", "DummyText", tests.RtRepo1, "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Setenv(coreutils.FailNoOp, "true"))
	err = artifactoryCli.Exec("copy", "DummyText", tests.RtRepo1)
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Unsetenv(coreutils.FailNoOp))

	// Test delete command
	err = artifactoryCli.Exec("delete", "DummyText", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Setenv(coreutils.FailNoOp, "true"))
	err = artifactoryCli.Exec("delete", "DummyText")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Unsetenv(coreutils.FailNoOp))

	// Test search command
	err = artifactoryCli.Exec("s", "DummyText", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Setenv(coreutils.FailNoOp, "true"))
	err = artifactoryCli.Exec("s", "DummyText")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Unsetenv(coreutils.FailNoOp))

	// Test props commands
	err = artifactoryCli.Exec("sp", "DummyText", "prop=val;key=value", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Setenv(coreutils.FailNoOp, "true"))
	err = artifactoryCli.Exec("sp", "DummyText", "prop=val;key=value")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	err = artifactoryCli.Exec("delp", "DummyText", "prop=val;key=valuos.Unsetenv(coreutils.FailNoOp)e", "--fail-no-op=true")

	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Setenv(coreutils.FailNoOp, "true"))
	err = artifactoryCli.Exec("delp", "DummyText", "prop=val;key=value")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	assert.NoError(t, os.Unsetenv(coreutils.FailNoOp))

	cleanArtifactoryTest()
}

func checkExitCode(t *testing.T, expected coreutils.ExitCode, er error) {
	switch underlyingType := er.(type) {
	case coreutils.CliError:
		assert.Equal(t, expected, underlyingType.ExitCode)
	default:
		assert.Fail(t, "Exit code expected error code %v, got %v", expected.Code, er)
	}
}
func TestArtifactoryDirectoryCopy(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1+"/path/", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/path/", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyUsingWildcard(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1+"/path/", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/*/", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyFilesNameWithParentheses(t *testing.T) {
	initArtifactoryTest(t, "")

	runRt(t, "upload", "testdata/b/*", tests.RtRepo1, "--flat=false")
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/(/(.in", tests.RtRepo2)
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/(b/(b.in", tests.RtRepo2)
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/b(/b(.in", tests.RtRepo2)
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/b)/b).in", tests.RtRepo2)
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/(b)/(b).in", tests.RtRepo2)
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/)b/)b.in", tests.RtRepo2)
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/)b)/)b).in", tests.RtRepo2)
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/(b/(b.in", tests.RtRepo2+"/()/", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/(b)/(b).in", tests.RtRepo2+"/()/")
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/b(/b(.in", tests.RtRepo2+"/(/", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/(/(*.in)", tests.RtRepo2+"/c/{1}.zip", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/(/(*.in)", tests.RtRepo2+"/(/{1}.zip")
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/b(/(b*.in)", tests.RtRepo2+"/(/{1}-up", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/testdata/b/b(/(*).(*)", tests.RtRepo2+"/(/{2}-{1}", "--flat=true")

	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetCopyFileNameWithParentheses(), searchPath, serverDetails, t)

	cleanArtifactoryTest()
}

func TestArtifactoryCreateUsers(t *testing.T) {
	initArtifactoryTest(t, "")
	usersCSVPath := "testdata/usersmanagement/users.csv"
	randomUsersCSVPath, err := tests.ReplaceTemplateVariables(usersCSVPath, "")
	assert.NoError(t, err)
	runRt(t, "users-create", "--csv="+randomUsersCSVPath)
	// Clean up
	defer func() {
		runRt(t, "users-delete", "--csv="+randomUsersCSVPath)
		cleanArtifactoryTest()
	}()

	verifyUsersExistInArtifactory(randomUsersCSVPath, t)
}

func verifyUsersExistInArtifactory(csvFilePath string, t *testing.T) {
	// Parse input CSV
	output, err := os.Open(csvFilePath)
	assert.NoError(t, err)
	csvReader := csv.NewReader(output)
	// Ignore the header
	_, err = csvReader.Read()
	assert.NoError(t, err)
	for {
		// Read each record from csv
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		user, password := record[0], record[1]
		err = tests.NewJfrogCli(execMain, "jfrog rt", "--url="+serverDetails.ArtifactoryUrl+" --user="+user+" --password="+password).Exec("ping")
		assert.NoError(t, err)
	}

}

func TestArtifactoryUploadFilesNameWithParenthesis(t *testing.T) {
	initArtifactoryTest(t, "")

	specFile, err := tests.CreateSpec(tests.UploadFileWithParenthesesSpec)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)

	searchPath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetUploadFileNameWithParentheses(), searchPath, serverDetails, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadFilesNameWithParenthesis(t *testing.T) {
	initArtifactoryTest(t, "")

	runRt(t, "upload", "testdata/b/*", tests.RtRepo1, "--flat=false")
	runRt(t, "download", path.Join(tests.RtRepo1), tests.Out+"/")

	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(tests.GetFileWithParenthesesDownload(), paths, t)

	cleanArtifactoryTest()
}
func TestArtifactoryDownloadDotAsTarget(t *testing.T) {
	initArtifactoryTest(t, "")
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))
	_, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "DownloadDotAsTarget"), 100000)
	assert.NoError(t, err)
	runRt(t, "upload", tests.Out+"/*", tests.RtRepo1+"/p-modules/", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tests.Out)
	defer chdirCallback()

	runRt(t, "download", tests.RtRepo1+"/p-modules/DownloadDotAsTarget", ".")
	chdirCallback()

	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally([]string{tests.Out, filepath.Join(tests.Out, "p-modules"), filepath.Join(tests.Out, "p-modules", "DownloadDotAsTarget")}, paths, t)
	assert.NoError(t, fileutils.RemoveTempDir(tests.Out), "Couldn't remove temp dir")
	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyUsingWildcardFlat(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1+"/path/inner/", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/path/inner", tests.RtRepo2, "--flat=true")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetSingleDirectoryCopyFlat(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyPathsTwice(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()
	runRt(t, "upload", filePath, tests.RtRepo1+"/path/inner/", "--flat")

	log.Info("Copy Folder to root twice")
	runRt(t, "cp", tests.RtRepo1+"/path", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetSingleFileCopyFullPath(), searchPath, serverDetails, t)
	runRt(t, "cp", tests.RtRepo1+"/path", tests.RtRepo2)
	inttestutils.VerifyExistInArtifactory(tests.GetSingleFileCopyFullPath(), searchPath, serverDetails, t)
	runRt(t, "del", tests.RtRepo2)

	log.Info("Copy to from repo1/path to repo2/path twice")
	runRt(t, "cp", tests.RtRepo1+"/path", tests.RtRepo2+"/path")
	inttestutils.VerifyExistInArtifactory(tests.GetSingleFileCopyFullPath(), searchPath, serverDetails, t)
	runRt(t, "cp", tests.RtRepo1+"/path", tests.RtRepo2+"/path")
	inttestutils.VerifyExistInArtifactory(tests.GetFolderCopyTwice(), searchPath, serverDetails, t)
	runRt(t, "del", tests.RtRepo2)

	log.Info("Copy to from repo1/path/ to repo2/path/ twice")
	runRt(t, "cp", tests.RtRepo1+"/path/", tests.RtRepo2+"/path/")
	inttestutils.VerifyExistInArtifactory(tests.GetSingleInnerFileCopyFullPath(), searchPath, serverDetails, t)
	runRt(t, "cp", tests.RtRepo1+"/path/", tests.RtRepo2+"/path/")
	inttestutils.VerifyExistInArtifactory(tests.GetSingleInnerFileCopyFullPath(), searchPath, serverDetails, t)
	runRt(t, "del", tests.RtRepo2)

	log.Info("Copy to from repo1/path/ to repo2/path/ twice")
	runRt(t, "cp", tests.RtRepo1+"/path", tests.RtRepo2+"/path/")
	inttestutils.VerifyExistInArtifactory(tests.GetFolderCopyIntoFolder(), searchPath, serverDetails, t)
	runRt(t, "cp", tests.RtRepo1+"/path", tests.RtRepo2+"/path/")
	inttestutils.VerifyExistInArtifactory(tests.GetFolderCopyIntoFolder(), searchPath, serverDetails, t)
	runRt(t, "del", tests.RtRepo2)

	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyPatternEndsWithSlash(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1+"/path/inner/")
	runRt(t, "cp", tests.RtRepo1+"/path/", tests.RtRepo2, "--flat=true")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetAnyItemCopyUsingSpec(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemUsingWildcardFlat(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1+"/path/inner/", "--flat")
	runRt(t, "upload", filePath, tests.RtRepo1+"/someFile", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/*", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetAnyItemCopy(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemRecursive(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1+"/a/b/", "--flat")
	runRt(t, "upload", filePath, tests.RtRepo1+"/aFile", "--flat=true")
	runRt(t, "cp", tests.RtRepo1+"/a*", tests.RtRepo2, "--recursive=true")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetAnyItemCopyRecursive(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAndRenameFolder(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	runRt(t, "upload", filePath, tests.RtRepo1+"/path/inner/", "--flat")
	runRt(t, "cp", tests.RtRepo1+"/path/(*)", tests.RtRepo2+"/newPath/{1}")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetCopyFolderRename(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemUsingSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = getSpecialCharFilePath()

	specFile, err := tests.CreateSpec(tests.CopyItemsSpec)
	assert.NoError(t, err)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	runRt(t, "upload", filePath, tests.RtRepo1+"/path/inner/")
	runRt(t, "upload", filePath, tests.RtRepo1+"/someFile", "--flat=true")
	runRt(t, "cp", "--spec="+specFile)
	inttestutils.VerifyExistInArtifactory(tests.GetAnyItemCopyUsingSpec(), searchPath, serverDetails, t)
	cleanArtifactoryTest()
}

func getSpecialCharFilePath() string {
	return "testdata/a$+~&^a#/a*"
}

func getAntPatternFilePath() string {
	return "testdata/cache/downloa?/**/*.in"
}

func TestArtifactoryCopyNoSpec(t *testing.T) {
	testCopyMoveNoSpec("cp", tests.GetBuildBeforeCopyExpected(), tests.GetBuildCopyExpected(), t)
}

func TestArtifactoryCopyExcludeByCli(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA)
	runRt(t, "upload", "--spec="+specFileB)

	// Copy by pattern
	runRt(t, "cp", tests.RtRepo1+"/data/", tests.RtRepo2+"/", "--exclusions=*/*b*;*/*c*")

	// Validate files are moved by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetBuildCopyExclude(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadDebian(t *testing.T) {
	initArtifactoryTest(t, "")
	specFile, err := tests.CreateSpec(tests.DebianUploadSpec)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile, "--deb=bionic/main/i386")
	verifyExistInArtifactoryByProps(tests.GetUploadDebianExpected(), tests.RtDebianRepo+"/*", "deb.distribution=bionic;deb.component=main;deb.architecture=i386", t)
	runRt(t, "upload", "--spec="+specFile, "--deb=cosmic/main\\/18.10/amd64")
	verifyExistInArtifactoryByProps(tests.GetUploadDebianExpected(), tests.RtDebianRepo+"/*", "deb.distribution=cosmic;deb.component=main/18.10;deb.architecture=amd64", t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadAndExplode(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", filepath.Join("testdata", "archives", "a.zip"), tests.RtRepo1, "--explode=true", "--flat")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetExplodeUploadExpectedRepo1(), searchFilePath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestUploadAndSyncDeleteCaseSensitivity(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", "--sync-deletes", tests.RtRepo1+"/", path.Join("testdata", "syncdeletes", "*"), tests.RtRepo1+"/")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	resultItems, err := inttestutils.SearchInArtifactory(searchFilePath, serverDetails, t)
	assert.NoError(t, err)
	assert.Len(t, resultItems, 5)
	// Make sure files are being deleted for the second sync-deletes upload
	runRt(t, "upload", "--sync-deletes", tests.RtRepo1+"/", path.Join("testdata", "syncdeletes", "industrial-os-7.4.3", "TestReport", "*"), tests.RtRepo1+"/")
	resultItems, err = inttestutils.SearchInArtifactory(searchFilePath, serverDetails, t)
	assert.NoError(t, err)
	assert.Len(t, resultItems, 1)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadAndSyncDelete(t *testing.T) {
	initArtifactoryTest(t, "")
	// Upload all testdata/a/
	runRt(t, "upload", path.Join("testdata", "a", "*"), tests.RtRepo1+"/syncDir/", "--flat=false")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep1(), searchFilePath, serverDetails, t)
	// Upload testdata/a/b/*1.in and sync syncDir/testdata/a/b/
	runRt(t, "upload", path.Join("testdata", "a", "b", "*1.in"), tests.RtRepo1+"/syncDir/", "--sync-deletes="+tests.RtRepo1+"/syncDir/testdata/a/b/", "--flat=false")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep2(), searchFilePath, serverDetails, t)
	// Upload testdata/archives/* and sync syncDir/
	runRt(t, "upload", path.Join("testdata", "archives", "*"), tests.RtRepo1+"/syncDir/", "--sync-deletes="+tests.RtRepo1+"/syncDir/", "--flat=true")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep3(), searchFilePath, serverDetails, t)
	// Upload testdata/b/ and sync syncDir/testdata/b/b
	// Noticed that testdata/c/ includes sub folders with special chars like '-' and '#'
	runRt(t, "upload", path.Join("testdata", "c", "*"), tests.RtRepo1+"/syncDir/", "--sync-deletes="+tests.RtRepo1+"/syncDir/", "--flat=false")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep4(), searchFilePath, serverDetails, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplode(t *testing.T) {
	initArtifactoryTest(t, "")
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	randFile, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "randFile"), 100000)
	assert.NoError(t, err)

	err = archiver.Archive([]string{randFile.Name()}, filepath.Join(tests.Out, "concurrent.tar.gz"))
	assert.NoError(t, err)
	err = archiver.Archive([]string{randFile.Name()}, filepath.Join(tests.Out, "bulk.tar"))
	assert.NoError(t, err)
	err = archiver.Archive([]string{randFile.Name()}, filepath.Join(tests.Out, "zipFile.zip"))
	assert.NoError(t, err)
	runRt(t, "upload", tests.Out+"/*", tests.RtRepo1, "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	// Download 'concurrent.tar.gz' as 'concurrent' file name and explode it.
	runRt(t, "download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/concurrent", "--explode=true")
	// Download 'concurrent.tar.gz' and explode it.
	runRt(t, "download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/", "--explode=true")
	// Download 'concurrent.tar.gz' without explode it.
	runRt(t, "download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/", "--explode=false")
	// Try to explode the archive that already been downloaded.
	runRt(t, "download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/", "--explode=true")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	runRt(t, "download", path.Join(tests.RtRepo1, "randFile"), tests.Out+"/", "--explode=true")
	runRt(t, "download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/", "--explode=false")
	runRt(t, "download", path.Join(tests.RtRepo1, "bulk.tar"), tests.Out+"/", "--explode=true")
	runRt(t, "download", path.Join(tests.RtRepo1, "zipFile.zip"), tests.Out+"/", "--explode=true")
	verifyExistAndCleanDir(t, tests.GetExtractedDownload)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplodeCurDirAsTarget(t *testing.T) {
	initArtifactoryTest(t, "")
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	randFile, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "DownloadAndExplodeCurDirTarget"), 100000)
	assert.NoError(t, err)

	err = archiver.Archive([]string{randFile.Name()}, filepath.Join(tests.Out, "curDir.tar.gz"))
	assert.NoError(t, err)

	runRt(t, "upload", tests.Out+"/*", tests.RtRepo1, "--flat=true")
	runRt(t, "upload", tests.Out+"/*", tests.RtRepo1+"/p-modules/", "--flat=true")
	assert.NoError(t, fileutils.RemoveTempDir(tests.Out), "Couldn't remove temp dir")

	// Change working dir to tests temp "out" dir
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tests.Out)
	defer chdirCallback()

	// Dot as target
	runRt(t, "download", tests.RtRepo1+"/p-modules/curDir.tar.gz", ".", "--explode=true")
	// Changing current working dir to "out" dir
	chdirCallback()
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadCurDir)
	clientTestUtils.ChangeDirAndAssert(t, tests.Out)

	// No target
	runRt(t, "download", tests.RtRepo1+"/p-modules/curDir.tar.gz", "--explode=true")
	// Changing working dir for testing
	chdirCallback()
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadCurDir)
	clientTestUtils.ChangeDirAndAssert(t, tests.Out)

	chdirCallback()
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplodeFlat(t *testing.T) {
	initArtifactoryTest(t, "")
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	file1, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "file1"), 100000)
	assert.NoError(t, err)

	err = archiver.Archive([]string{file1.Name()}, filepath.Join(tests.Out, "flat.tar"))
	assert.NoError(t, err)
	err = archiver.Archive([]string{tests.Out + "/flat.tar"}, filepath.Join(tests.Out, "tarZipFile.zip"))
	assert.NoError(t, err)

	runRt(t, "upload", tests.Out+"/*", tests.RtRepo1+"/checkFlat/dir/", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))

	runRt(t, "download", path.Join(tests.RtRepo1, "checkFlat", "dir", "flat.tar"), tests.Out+"/checkFlat/", "--explode=true", "--flat=true", "--min-split=50")
	runRt(t, "download", path.Join(tests.RtRepo1, "checkFlat", "dir", "tarZipFile.zip"), tests.Out+"/", "--explode=true", "--flat=false")
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadFlatFalse)
	// Explode 'flat.tar' while the file exists in the file system using --flat
	runRt(t, "download", path.Join(tests.RtRepo1, "checkFlat", "dir", "tarZipFile.zip"), tests.Out+"/", "--explode=true", "--flat=false")
	runRt(t, "download", path.Join(tests.RtRepo1, "checkFlat", "dir", "flat.tar"), tests.Out+"/checkFlat/dir/", "--explode=true", "--flat", "--min-split=50")
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadTarFileFlatFalse)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplodeConcurrent(t *testing.T) {
	initArtifactoryTest(t, "")
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	runRt(t, "upload", path.Join("testdata", "archives", "a.zip"), tests.RtRepo1, "--flat=true")
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))
	runRt(t, "download", path.Join(tests.RtRepo1, "a.zip"), tests.Out+"/", "--explode=true", "--min-split=50")
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadConcurrent)
	runRt(t, "download", path.Join(tests.RtRepo1, "a.zip"), tests.Out+"/", "--explode=false", "--min-split=50")
	verifyExistAndCleanDir(t, tests.GetArchiveConcurrent)
	runRt(t, "download", path.Join(tests.RtRepo1, "a.zip"), tests.Out+"/", "--explode=true", "--split-count=15", "--min-split=50")
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadConcurrent)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplodeSpecialChars(t *testing.T) {
	initArtifactoryTest(t, "")
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)
	file1, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "file $+~&^a#1"), 1000)
	assert.NoError(t, err)
	err = archiver.Archive([]string{file1.Name()}, filepath.Join(tests.Out, "a$+~&^a#.tar"))
	assert.NoError(t, err)

	runRt(t, "upload", tests.Out+"/*", tests.RtRepo1+"/dir/", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	err = fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)
	runRt(t, "dl", tests.RtRepo1+"/dir/a$+~&^a#.tar", tests.Out+"/dir $+~&^a# test/", "--flat=true", "--explode")
	runRt(t, "dl", tests.RtRepo1+"/dir/a$+~&^a#.tar", tests.Out+"/dir $+~&^a# test/", "--flat=false", "--explode")
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadTarFileSpecialChars)
	// Concurrently download
	runRt(t, "dl", tests.RtRepo1+"/dir/a$+~&^a#.tar", tests.Out+"/dir $+~&^a# test/", "--flat=true", "--explode", "--min-split=50")
	runRt(t, "dl", tests.RtRepo1+"/dir/a$+~&^a#.tar", tests.Out+"/dir $+~&^a# test/", "--flat=false", "--explode", "--min-split=50")
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadTarFileSpecialChars)
	cleanArtifactoryTest()
}

func verifyExistAndCleanDir(t *testing.T, GetExtractedDownload func() []string) {
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(GetExtractedDownload(), paths, t)
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))
}

func TestArtifactoryUploadAsArchive(t *testing.T) {
	initArtifactoryTest(t, "")

	uploadSpecFile, err := tests.CreateSpec(tests.UploadAsArchive)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+uploadSpecFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetUploadAsArchive(), searchFilePath, serverDetails, t)

	// Verify the properties are valid
	resultItems := searchItemsInArtifactory(t, tests.SearchAllRepo1)
	assert.NotZero(t, len(resultItems))
	for _, item := range resultItems {
		if item.Name != "a.zip" {
			assert.Zero(t, len(item.Properties))
			continue
		}
		properties := item.Properties
		assert.Equal(t, 3, len(properties))

		// Sort the properties alphabetically by key and value to make the comparison easier
		sort.Slice(properties, func(i, j int) bool {
			if properties[i].Key == properties[j].Key {
				return properties[i].Value < properties[j].Value
			}
			return properties[i].Key < properties[j].Key
		})
		assert.Contains(t, properties, rtutils.Property{Key: "k1", Value: "v11"})
		assert.Contains(t, properties, rtutils.Property{Key: "k1", Value: "v12"})
		assert.Contains(t, properties, rtutils.Property{Key: "k2", Value: "v2"})
	}

	// Check the files inside the archives by downloading and exploding them
	downloadSpecFile, err := tests.CreateSpec(tests.DownloadAndExplodeArchives)
	assert.NoError(t, err)
	runRt(t, "download", "--spec="+downloadSpecFile)
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(tests.GetDownloadArchiveAndExplode(), paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryUploadAsArchiveWithExplodeAndSymlinks(t *testing.T) {
	initArtifactoryTest(t, "")

	uploadSpecFile, err := tests.CreateSpec(tests.UploadAsArchive)
	assert.NoError(t, err)
	err = artifactoryCli.Exec("upload", "--spec="+uploadSpecFile, "--symlinks", "--explode")
	assert.Error(t, err)

	cleanArtifactoryTest()
}

func TestArtifactoryUploadAsArchiveToDir(t *testing.T) {
	initArtifactoryTest(t, "")

	uploadSpecFile, err := tests.CreateSpec(tests.UploadAsArchiveToDir)
	assert.NoError(t, err)
	err = artifactoryCli.Exec("upload", "--spec="+uploadSpecFile)
	assert.Error(t, err)

	cleanArtifactoryTest()
}

func TestArtifactoryUploadAsArchiveWithIncludeDirs(t *testing.T) {
	initArtifactoryTest(t, "")
	assert.NoError(t, createEmptyTestDir())
	uploadSpecFile, err := tests.CreateSpec(tests.UploadAsArchiveEmptyDirs)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+uploadSpecFile)

	// Check the empty directories inside the archive by downloading and exploding it.
	downloadSpecFile, err := tests.CreateSpec(tests.DownloadAndExplodeArchives)
	assert.NoError(t, err)
	runRt(t, "download", "--spec="+downloadSpecFile)
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	downloadedEmptyDirs := tests.GetDownloadArchiveAndExplodeWithIncludeDirs()
	// Verify dirs exists.
	tests.VerifyExistLocally(downloadedEmptyDirs, paths, t)
	// Verify empty dirs.
	verifyEmptyDirs(t, downloadedEmptyDirs)

	// Check the empty directories inside the archive by downloading without exploding it, using os "unzip" command.
	assert.NoError(t, fileutils.RemoveTempDir(tests.Out), "Couldn't remove temp dir")
	assert.NoError(t, os.MkdirAll(tests.Out, 0777))
	downloadSpecFile, err = tests.CreateSpec(tests.DownloadWithoutExplodeArchives)
	assert.NoError(t, err)
	runRt(t, "download", "--spec="+downloadSpecFile)
	// Change working directory to the zip file's location and unzip it.
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, path.Join(tests.Out, "archive", "archive"))
	defer chdirCallback()
	cmd := exec.Command("unzip", "archive.zip")
	assert.NoError(t, errorutils.CheckError(cmd.Run()))
	chdirCallback()
	verifyEmptyDirs(t, downloadedEmptyDirs)
	cleanArtifactoryTest()
}

func verifyEmptyDirs(t *testing.T, dirPaths []string) {
	for _, dirPath := range dirPaths {
		empty, err := fileutils.IsDirEmpty(dirPath)
		assert.NoError(t, err)
		assert.True(t, empty)
	}
}

func createEmptyTestDir() error {
	dirInnerPath := filepath.Join("empty", "folder")
	canonicalPath := tests.GetTestResourcesPath() + dirInnerPath
	return os.MkdirAll(canonicalPath, 0777)
}

func TestArtifactoryDownloadAndSyncDeletes(t *testing.T) {
	initArtifactoryTest(t, "")

	outDirPath := tests.Out + string(os.PathSeparator)
	// Upload all testdata/a/ to repo1/syncDir/
	runRt(t, "upload", path.Join("testdata", "a", "*"), tests.RtRepo1+"/syncDir/", "--flat=false")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep1(), searchFilePath, serverDetails, t)

	// Download repo1/syncDir/ to out/
	runRt(t, "download", tests.RtRepo1+"/syncDir/", tests.Out+"/")
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(tests.GetExpectedSyncDeletesDownloadStep2(), paths, t)

	// Download repo1/syncDir/ to out/ with flat=true and sync out/
	runRt(t, "download", tests.RtRepo1+"/syncDir/", outDirPath, "--flat=true", "--sync-deletes="+outDirPath)
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetExpectedSyncDeletesDownloadStep3(), paths, t)

	// Download all files ended with 2.in from repo1/syncDir/ to out/ and sync out/
	runRt(t, "download", tests.RtRepo1+"/syncDir/*2.in", outDirPath, "--flat=true", "--sync-deletes="+outDirPath)
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetExpectedSyncDeletesDownloadStep4(), paths, t)

	// Download repo1/syncDir/ to out/, exclude the pattern "*c*.in" and sync out/
	runRt(t, "download", tests.RtRepo1+"/syncDir/", outDirPath, "--sync-deletes="+outDirPath+"syncDir"+string(os.PathSeparator), "--exclusions=*/syncDir/testdata/*c*in")
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out+string(os.PathSeparator)+"syncDir"+string(os.PathSeparator), false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetSyncExpectedDeletesDownloadStep5(), paths, t)

	// Delete all files from repo1/syncDir/
	runRt(t, "delete", tests.RtRepo1+"/syncDir/")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyDoesntExistInArtifactory(searchFilePath, t)

	// Upload all testdata/archives/ to repo1/syncDir/
	runRt(t, "upload", path.Join("testdata", "archives", "*"), tests.RtRepo1+"/syncDir/", "--flat=false")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSyncExpectedDeletesDownloadStep6(), searchFilePath, serverDetails, t)

	// Download repo1/syncDir/ to out/ and sync out/
	runRt(t, "download", tests.RtRepo1+"/syncDir/", outDirPath, "--sync-deletes="+outDirPath+"syncDir"+string(os.PathSeparator))
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out+string(os.PathSeparator)+"syncDir"+string(os.PathSeparator), false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetSyncExpectedDeletesDownloadStep7(), paths, t)

	cleanArtifactoryTest()
}

// After syncDeletes we must make sure that the content of the synced directory contains the last operation result only.
// Therefore, we verify that there are no other files in the synced directory, other than the list of the expected files.
func checkSyncedDirContent(expected, actual []string, t *testing.T) {
	// Check if all expected files are actually exist
	tests.VerifyExistLocally(expected, actual, t)
	// Check if all the existing files were expected
	err := isExclusivelyExistLocally(expected, actual)
	assert.NoError(t, err)
}

// Check if only the files we were expecting, exist locally, i.e return an error if there is a local file we didn't expect.
// Since the "actual" list contains paths of both directories and files, for each element in the "actual" list:
// Check if the path equals to an existing file (for a file) OR
// if the path is a prefix of some path of an existing file (for a dir).
func isExclusivelyExistLocally(expected, actual []string) error {
	expectedLastIndex := len(expected) - 1
	for _, v := range actual {
		for i, r := range expected {
			if strings.HasPrefix(r, v) || v == r {
				break
			}
			if i == expectedLastIndex {
				return errors.New("Should not have : " + v)
			}
		}
	}
	return nil
}

// Test self-signed certificates with Artifactory. For the test, we set up a reverse proxy server.
func TestArtifactorySelfSignedCert(t *testing.T) {
	initArtifactoryTest(t, "")
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.HomeDir, tempDirPath)
	defer setEnvCallBack()
	setEnvCallBack = clientTestUtils.SetEnvWithCallbackAndAssert(t, tests.HttpsProxyEnvVar, "1024")
	defer setEnvCallBack()
	go cliproxy.StartLocalReverseHttpProxy(serverDetails.ArtifactoryUrl, false)

	// The two certificate files are created by the reverse proxy on startup in the current directory.
	defer clientTestUtils.RemoveAndAssert(t, certificate.KeyFile)
	defer clientTestUtils.RemoveAndAssert(t, certificate.CertFile)
	// Let's wait for the reverse proxy to start up.
	err := checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", false)
	assert.NoError(t, err)

	fileSpec := spec.NewBuilder().Pattern(tests.RtRepo1 + "/*.zip").Recursive(true).BuildSpec()
	assert.NoError(t, err)
	parsedUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	assert.NoError(t, err)
	serverDetails.ArtifactoryUrl = "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()

	// The server is using self-signed certificates
	// Without loading the certificates (or skipping verification) we expect all actions to fail due to error: "x509: certificate signed by unknown authority"
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err := searchCmd.Search()
	if reader != nil {
		readerCloseAndAssert(t, reader)
	}
	_, isUrlErr := err.(*url.Error)
	assert.True(t, isUrlErr, "Expected a connection failure, since reverse proxy didn't load self-signed-certs. Connection however is successful", err)

	// Set insecureTls to true and run again. We expect the command to succeed.
	serverDetails.InsecureTls = true
	searchCmd = generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	readerCloseAndAssert(t, reader)

	// Set insecureTls back to false.
	// Copy the server certificates to the CLI security dir and run again. We expect the command to succeed.
	serverDetails.InsecureTls = false
	certsPath, err := coreutils.GetJfrogCertsDir()
	assert.NoError(t, err)
	err = fileutils.CopyFile(certsPath, certificate.KeyFile)
	assert.NoError(t, err)
	err = fileutils.CopyFile(certsPath, certificate.CertFile)
	assert.NoError(t, err)
	searchCmd = generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	readerCloseAndAssert(t, reader)

	serverDetails.ArtifactoryUrl = artAuth.GetUrl()
	cleanArtifactoryTest()
}

// Test client certificates with Artifactory. For the test, we set up a reverse proxy server.
func TestArtifactoryClientCert(t *testing.T) {
	initArtifactoryTest(t, "")
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.HomeDir, tempDirPath)
	defer setEnvCallBack()
	setEnvCallBack = clientTestUtils.SetEnvWithCallbackAndAssert(t, tests.HttpsProxyEnvVar, "1025")
	defer setEnvCallBack()
	go cliproxy.StartLocalReverseHttpProxy(serverDetails.ArtifactoryUrl, true)

	// The two certificate files are created by the reverse proxy on startup in the current directory.
	defer clientTestUtils.RemoveAndAssert(t, certificate.KeyFile)
	defer clientTestUtils.RemoveAndAssert(t, certificate.CertFile)
	// Let's wait for the reverse proxy to start up.
	err := checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", true)
	assert.NoError(t, err)

	fileSpec := spec.NewBuilder().Pattern(tests.RtRepo1 + "/*.zip").Recursive(true).BuildSpec()
	assert.NoError(t, err)
	parsedUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	assert.NoError(t, err)
	serverDetails.ArtifactoryUrl = "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()
	serverDetails.InsecureTls = true

	// The server is requiring client certificates
	// Without loading a valid client certificate, we expect all actions to fail due to error: "tls: bad certificate"
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err := searchCmd.Search()
	if reader != nil {
		readerCloseAndAssert(t, reader)
	}
	_, isUrlErr := err.(*url.Error)
	assert.True(t, isUrlErr, "Expected a connection failure, since client did not provide a client certificate. Connection however is successful")

	// Inject client certificates, we expect the search to succeed
	serverDetails.ClientCertPath = certificate.CertFile
	serverDetails.ClientCertKeyPath = certificate.KeyFile

	searchCmd = generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err = searchCmd.Search()
	if reader != nil {
		readerCloseAndAssert(t, reader)
	}
	assert.NoError(t, err)

	serverDetails.ArtifactoryUrl = artAuth.GetUrl()
	serverDetails.InsecureTls = false
	serverDetails.ClientCertPath = ""
	serverDetails.ClientCertKeyPath = ""
	cleanArtifactoryTest()
}

func getExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("check connection to the network")
}

// Due to the fact that go reads the HTTP_PROXY and the HTTPS_PROXY
// argument only once we can't set the env var for specific test.
// We need to start a new process with the env var set to the value we want.
// We decide which var to set by the rtUrl scheme.
func TestArtifactoryProxy(t *testing.T) {
	initArtifactoryTest(t, "")
	rtUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	assert.NoError(t, err)
	var proxyTestArgs []string
	var httpProxyEnv string
	testArgs := []string{"-test.artifactoryProxy=true", "-jfrog.url=" + *tests.JfrogUrl, "-jfrog.user=" + *tests.JfrogUser, "-jfrog.password=" + *tests.JfrogPassword, "-jfrog.sshKeyPath=" + *tests.JfrogSshKeyPath, "-jfrog.sshPassphrase=" + *tests.JfrogSshPassphrase, "-jfrog.adminToken=" + *tests.JfrogAccessToken}
	if rtUrl.Scheme == "https" {
		setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, tests.HttpsProxyEnvVar, "1026")
		defer setEnvCallBack()
		proxyTestArgs = append([]string{"test", "-run=TestArtifactoryHttpsProxyEnvironmentVariableDelegator"}, testArgs...)
		httpProxyEnv = "HTTPS_PROXY=localhost:" + cliproxy.GetProxyHttpsPort()
	} else {
		proxyTestArgs = append([]string{"test", "-run=TestArtifactoryHttpProxyEnvironmentVariableDelegator"}, testArgs...)
		httpProxyEnv = "HTTP_PROXY=localhost:" + cliproxy.GetProxyHttpPort()
	}
	runProxyTest(t, proxyTestArgs, httpProxyEnv)
	cleanArtifactoryTest()
}

func runProxyTest(t *testing.T, proxyTestArgs []string, httpProxyEnv string) {
	cmd := exec.Command("go", proxyTestArgs...)
	cmd.Env = append(os.Environ(), httpProxyEnv)

	tempDirPath, err := tests.GetTestsLogsDir()
	assert.NoError(t, err)
	f, err := os.Create(filepath.Join(tempDirPath, "artifactory_proxy_tests.log"))
	assert.NoError(t, err)

	cmd.Stdout, cmd.Stderr = f, f
	assert.NoError(t, cmd.Run(), "Artifactory proxy tests failed, full report available at the following path:", f.Name())
	log.Info("Full Artifactory proxy testing report available at the following path:", f.Name())
}

// Should be run only by @TestArtifactoryProxy test.
func TestArtifactoryHttpProxyEnvironmentVariableDelegator(t *testing.T) {
	testArtifactoryProxy(t, false)
}

// Should be run only by @TestArtifactoryProxy test.
func TestArtifactoryHttpsProxyEnvironmentVariableDelegator(t *testing.T) {
	testArtifactoryProxy(t, true)
}

func testArtifactoryProxy(t *testing.T, isHttps bool) {
	// Value is set to 'true' via testArgs @TestArtifactoryProxy
	if !*tests.TestArtifactoryProxy {
		t.SkipNow()
	}
	authenticate(true)
	proxyRtUrl := prepareArtifactoryUrlForProxyTest(t)
	spec := spec.NewBuilder().Pattern(tests.RtRepo1 + "/*.zip").Recursive(true).BuildSpec()
	serverDetails.ArtifactoryUrl = proxyRtUrl
	checkForErrDueToMissingProxy(spec, t)
	var port string
	if isHttps {
		go cliproxy.StartHttpsProxy()
		port = cliproxy.GetProxyHttpsPort()
	} else {
		go cliproxy.StartHttpProxy()
		port = cliproxy.GetProxyHttpPort()
	}
	// Let's wait for the reverse proxy to start up.
	err := checkIfServerIsUp(port, "http", false)
	assert.NoError(t, err)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(spec)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	if reader != nil {
		readerCloseAndAssert(t, reader)
	}
	serverDetails.ArtifactoryUrl = artAuth.GetUrl()
}

func prepareArtifactoryUrlForProxyTest(t *testing.T) string {
	rtUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	assert.NoError(t, err)
	rtHost, port, err := net.SplitHostPort(rtUrl.Host)
	assert.NoError(t, err)
	if rtHost == "localhost" || rtHost == "127.0.0.1" {
		externalIp, err := getExternalIP()
		assert.NoError(t, err)
		rtUrl.Host = externalIp + ":" + port
	}
	return rtUrl.String()
}

func checkForErrDueToMissingProxy(spec *spec.SpecFiles, t *testing.T) {
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(spec)
	reader, err := searchCmd.Search()
	if reader != nil {
		readerCloseAndAssert(t, reader)
	}
	_, isUrlErr := err.(*url.Error)
	assert.True(t, isUrlErr, "Expected the request to fails, since the proxy is down.", err)
}

func checkIfServerIsUp(port, proxyScheme string, useClientCerts bool) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if useClientCerts {
		for attempt := 0; attempt < 10; attempt++ {
			if _, err := os.Stat(certificate.CertFile); os.IsNotExist(err) {
				log.Info("Waiting for certificate to appear...")
				time.Sleep(time.Second)
				continue
			}

			if _, err := os.Stat(certificate.KeyFile); os.IsNotExist(err) {
				log.Info("Waiting for key to appear...")
				time.Sleep(time.Second)
				continue
			}

			break
		}

		cert, err := tls.LoadX509KeyPair(certificate.CertFile, certificate.KeyFile)
		if err != nil {
			return fmt.Errorf("failed loading client certificate")
		}
		tr.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}
	client := &http.Client{Transport: tr}

	for attempt := 0; attempt < 20; attempt++ {
		log.Info("Checking if proxy server is up and running.", strconv.Itoa(attempt+1), "attempt.", "URL:", proxyScheme+"://localhost:"+port)
		resp, err := client.Get(proxyScheme + "://localhost:" + port)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			log.Error(fmt.Sprintf("Couldn't close response body. Error: %s", err.Error()))
		}
		if resp.StatusCode != http.StatusOK {
			time.Sleep(time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed while waiting for the proxy server to be accessible")
}

func TestXrayScanBuild(t *testing.T) {
	initArtifactoryTest(t, "")
	xrayServerPort := xray.StartXrayMockServer()
	serverUrl := "--url=http://localhost:" + strconv.Itoa(xrayServerPort)
	artifactoryCommandExecutor := tests.NewJfrogCli(execMain, "jfrog rt", serverUrl+getArtifactoryTestCredentials())
	assert.NoError(t, artifactoryCommandExecutor.Exec("build-scan", xray.CleanScanBuildName, "3"))

	cleanArtifactoryTest()
}

func TestArtifactorySetProperties(t *testing.T) {
	initArtifactoryTest(t, "")
	// Upload a file.
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/a.in")
	// Set the 'prop=red' property to the file.
	runRt(t, "sp", tests.RtRepo1+"/a.*", "prop=red")
	// Now let's change the property value, by searching for the 'prop=red'.
	specFile, err := tests.CreateSpec(tests.SetDeletePropsSpec)
	assert.NoError(t, err)
	runRt(t, "sp", "prop=green", "--spec="+specFile)

	resultItems := searchItemsInArtifactory(t, tests.SearchRepo1ByInSuffix)
	assert.NotZero(t, len(resultItems), "No artifacts were found.")
	for _, item := range resultItems {
		properties := item.Properties
		assert.GreaterOrEqual(t, len(properties), 1, "Failed setting properties on item:", item.GetItemRelativePath())
		for i, prop := range properties {
			assert.Zero(t, i, "Expected a single property.")
			assert.Equal(t, "prop", prop.Key, "Wrong property key")
			assert.Equal(t, "green", prop.Value, "Wrong property value")
		}
	}
	cleanArtifactoryTest()
}

func TestArtifactorySetPropertiesOnSpecialCharsArtifact(t *testing.T) {
	initArtifactoryTest(t, "")
	targetPath := path.Join(tests.RtRepo1, "a$+~&^a#")
	// Upload a file with special chars.
	runRt(t, "upload", "testdata/a/a1.in", targetPath)
	// Set the 'prop=red' property to the file.
	runRt(t, "sp", targetPath, "prop=red")

	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	resultItems, err := inttestutils.SearchInArtifactory(searchSpec, serverDetails, t)
	assert.NoError(t, err)

	assert.Equal(t, len(resultItems), 1)
	for _, item := range resultItems {
		properties := item.Props
		assert.Equal(t, len(properties), 1)
		for k, v := range properties {
			assert.Equal(t, "prop", k, "Wrong property key")
			assert.Len(t, v, 1)
			assert.Equal(t, "red", v[0], "Wrong property value")
		}
	}
	cleanArtifactoryTest()
}

func TestArtifactorySetPropertiesExcludeByCli(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", "testdata/a/a*.in", tests.RtRepo1+"/")
	runRt(t, "sp", tests.RtRepo1+"/*", "prop=val", "--exclusions=*/*a1.in;*/*a2.in")
	resultItems := searchItemsInArtifactory(t, tests.SearchRepo1ByInSuffix)
	assert.NotZero(t, len(resultItems), "No artifacts were found.")
	for _, item := range resultItems {
		if item.Name != "a3.in" {
			continue
		}
		properties := item.Properties
		assert.GreaterOrEqual(t, len(properties), 1, "Failed setting properties on item:", item.GetItemRelativePath())
		for i, prop := range properties {
			assert.Zero(t, i, "Expected single property.")
			assert.Equal(t, "prop", prop.Key, "Wrong property key")
			assert.Equal(t, "val", prop.Value, "Wrong property value")
		}
	}
	cleanArtifactoryTest()
}

func TestArtifactorySetPropertiesExclusionsByCli(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", "testdata/a/a*.in", tests.RtRepo1+"/")
	runRt(t, "sp", tests.RtRepo1+"/*", "prop=val", "--exclusions=*/*a1.in;*/*a2.in")
	resultItems := searchItemsInArtifactory(t, tests.SearchRepo1ByInSuffix)
	assert.NotZero(t, len(resultItems), "No artifacts were found.")
	for _, item := range resultItems {
		if item.Name != "a3.in" {
			continue
		}
		properties := item.Properties
		assert.GreaterOrEqual(t, len(properties), 1, "Failed setting properties on item:", item.GetItemRelativePath())
		for i, prop := range properties {
			assert.Zero(t, i, "Expected single property.")
			assert.Equal(t, "prop", prop.Key, "Wrong property key")
			assert.Equal(t, "val", prop.Value, "Wrong property value")
		}
	}
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteProperties(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", "testdata/a/a*.in", tests.RtRepo1+"/a/")
	runRt(t, "sp", tests.RtRepo1+"/a/*", "color=yellow;prop=red;status=ok")
	// Delete the 'color' property.
	runRt(t, "delp", tests.RtRepo1+"/a/*", "color")
	// Delete the 'status' property, by a spec which filters files by 'prop=red'.
	specFile, err := tests.CreateSpec(tests.SetDeletePropsSpec)
	assert.NoError(t, err)
	runRt(t, "delp", "status", "--spec="+specFile)

	resultItems := searchItemsInArtifactory(t, tests.SearchRepo1ByInSuffix)
	assert.NotZero(t, len(resultItems), "No artifacts were found.")

	for _, item := range resultItems {
		properties := item.Properties
		for _, prop := range properties {
			assert.False(t, prop.Key == "color" || prop.Key == "status", "Properties 'color' and/or 'status' were not deleted from artifact", item.Name)
		}
	}
	cleanArtifactoryTest()
}

func TestArtifactoryDeletePropertiesWithExclude(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", "testdata/a/a*.in", tests.RtRepo1+"/")
	runRt(t, "sp", tests.RtRepo1+"/*", "prop=val")

	runRt(t, "delp", tests.RtRepo1+"/*", "prop", "--exclusions=*/*a1.in;*/*a2.in")
	resultItems := searchItemsInArtifactory(t, tests.SearchRepo1ByInSuffix)
	assert.NotZero(t, len(resultItems), "No artifacts were found.")

	for _, item := range resultItems {
		properties := item.Properties
		for _, prop := range properties {
			if item.Name == "a1.in" || item.Name == "a2.in" {
				// Check that properties were not removed.
				assert.Equal(t, "prop", prop.Key, "Wrong property key")
				assert.Equal(t, "val", prop.Value, "Wrong property value")
			}
		}
	}
	cleanArtifactoryTest()
}

func TestArtifactoryDeletePropertiesWithExclusions(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", "testdata/a/a*.in", tests.RtRepo1+"/")
	runRt(t, "sp", tests.RtRepo1+"/*", "prop=val")

	runRt(t, "delp", tests.RtRepo1+"/*", "prop", "--exclusions=*/*a1.in;*/*a2.in")
	resultItems := searchItemsInArtifactory(t, tests.SearchRepo1ByInSuffix)
	assert.NotZero(t, len(resultItems), "No artifacts were found.")

	for _, item := range resultItems {
		properties := item.Properties
		for _, prop := range properties {
			if item.Name == "a1.in" || item.Name == "a2.in" {
				// Check that properties were not removed.
				assert.False(t, prop.Key != "prop" || prop.Value != "val", "Wrong properties")
			}
		}
	}
	cleanArtifactoryTest()
}

func TestArtifactoryUploadOneArtifactToMultipleLocation(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumber := "333"
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/root/", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)
	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumber)
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.RtBuildName1, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	assert.Equal(t, 2, len(publishedBuildInfo.BuildInfo.Modules[0].Artifacts))
	cleanArtifactoryTest()
}

func TestArtifactoryUploadFromHomeDir(t *testing.T) {
	initArtifactoryTest(t, "")
	testFileRel, testFileAbs := createFileInHomeDir(t, "cliTestFile.txt")
	runRt(t, "upload", testFileRel, tests.RtRepo1, "--recursive=false", "--flat=true")
	searchTxtPath, err := tests.CreateSpec(tests.SearchTxt)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetTxtUploadExpectedRepo1(), searchTxtPath, serverDetails, t)
	clientTestUtils.RemoveAndAssert(t, testFileAbs)

	cleanArtifactoryTest()
}

func createFileInHomeDir(t *testing.T, fileName string) (testFileRelPath string, testFileAbsPath string) {
	testFileRelPath = filepath.Join("~", fileName)
	testFileAbsPath = filepath.Join(fileutils.GetHomeDir(), fileName)
	d1 := []byte("test file")
	err := ioutil.WriteFile(testFileAbsPath, d1, 0644)
	assert.NoError(t, err, "Couldn't create file")
	return
}

func TestArtifactoryUploadExcludeByCli1Wildcard(t *testing.T) {
	initArtifactoryTest(t, "")
	// Upload files
	runRt(t, "upload", "testdata/a/a*", tests.RtRepo1, "--exclusions=*a2*;*a3.in", "--flat=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli1Regex(t *testing.T) {
	initArtifactoryTest(t, "")
	// Upload files
	runRt(t, "upload", "testdata/a/a(.*)", tests.RtRepo1, "--exclusions=(.*)a2.*;.*a3.in", "--regexp=true", "--flat=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli2Wildcard(t *testing.T) {
	initArtifactoryTest(t, "")

	// Create temp dir
	absDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	// Create temp files
	d1 := []byte("test file")
	err := ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile1.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile2.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")

	// Upload files
	runRt(t, "upload", filepath.ToSlash(absDirPath)+"/*", tests.RtRepo1, "--exclusions=*cliTestFile1*", "--flat=true")

	// Check files exists in artifactory
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory([]string{tests.RtRepo1 + "/cliTestFile2.in"}, searchFilePath, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli2Regex(t *testing.T) {
	initArtifactoryTest(t, "")

	// Create temp dir
	absDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	// Create temp files
	d1 := []byte("test file")
	err := ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile1.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile2.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")

	// Upload files
	runRt(t, "upload", filepath.ToSlash(absDirPath)+"(.*)", tests.RtRepo1, "--exclusions=(.*c)liTestFile1.*", "--regexp=true", "--flat=true")

	// Check files exists in artifactory
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory([]string{tests.RtRepo1 + "/cliTestFile2.in"}, searchFilePath, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeBySpecWildcard(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadSpecExclude)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)

	// Validate files are moved by build number
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetUploadSpecExcludeRepo1(), searchFilePath, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeBySpecRegex(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadSpecExcludeRegex)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)

	// Validate files are moved by build number
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetUploadSpecExcludeRepo1(), searchFilePath, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadWithRegexEscaping(t *testing.T) {
	initArtifactoryTest(t, "")
	// Upload files
	runRt(t, "upload", "testdata/regexp"+"(.*)"+"\\."+".*", tests.RtRepo1, "--regexp=true", "--flat=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory([]string{tests.RtRepo1 + "/has.dot"}, searchFilePath, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopySpec(t *testing.T) {
	testMoveCopySpec("copy", t)
}

func TestArtifactoryMoveSpec(t *testing.T) {
	testMoveCopySpec("move", t)
}

func testMoveCopySpec(command string, t *testing.T) {
	initArtifactoryTest(t, "")
	preUploadBasicTestResources(t)
	specFile, err := tests.CreateSpec(tests.CopyMoveSimpleSpec)
	assert.NoError(t, err)
	runRt(t, command, "--spec="+specFile)

	// Verify files exist in target location successfully
	searchMovedCopiedSpec, err := tests.CreateSpec(tests.SearchTargetInRepo2)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetMoveCopySpecExpected(), searchMovedCopiedSpec, serverDetails, t)

	searchOriginalSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)

	if command == "copy" {
		// Verify original files still exist
		inttestutils.VerifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchOriginalSpec, serverDetails, t)
	} else {
		// Verify original files were moved
		verifyDoesntExistInArtifactory(searchOriginalSpec, t)
	}

	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded.
// validate the symlink content checksum
func TestValidateValidSymlink(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	// Path to local file
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a", "a1.in")
	// Path to valid symLink
	validLink := filepath.Join(tests.GetTestResourcesPath()+"a", "link")

	// Link valid symLink to local file
	err := os.Symlink(localFile, validLink)
	assert.NoError(t, err)

	// Upload symlink to artifactory
	runRt(t, "u", validLink, tests.RtRepo1, "--symlinks=true", "--flat=true")

	// Delete the local symlink
	clientTestUtils.RemoveAndAssert(t, validLink)

	// Download symlink from artifactory
	runRt(t, "dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true")

	// Should be valid if successful
	validateSymLink(validLink, localFile, t)

	// Delete symlink and clean
	clientTestUtils.RemoveAndAssert(t, validLink)

	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Unlink and delete the pointed file.
// Download the symlink which was uploaded with validation. The command should fail.
func TestValidateBrokenSymlink(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	// Create temporary file in resourcesPath/a/
	tmpFile, err := ioutil.TempFile(tests.GetTestResourcesPath()+"a/", "a.in.")
	if assert.NoError(t, err) {
		assert.NoError(t, tmpFile.Close())
	}
	localFile := tmpFile.Name()

	// Path to the symLink
	linkPath := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	// Link to the temporary file
	err = os.Symlink(localFile, linkPath)
	assert.NoError(t, err)

	// Upload symlink to artifactory
	runRt(t, "u", linkPath, tests.RtRepo1, "--symlinks=true", "--flat=true")

	// Delete the local symlink and the temporary file
	clientTestUtils.RemoveAndAssert(t, linkPath)
	clientTestUtils.RemoveAndAssert(t, localFile)

	// Try downloading symlink from artifactory. Since the link should be broken, it shouldn't be downloaded
	err = artifactoryCli.Exec("dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true")
	if !assert.Error(t, err, "A broken symLink was downloaded although validate-symlinks flag was set to true") {
		clientTestUtils.RemoveAndAssert(t, linkPath)
	}

	// Clean
	cleanArtifactoryTest()
}

// Testing exclude pattern with symlinks.
// This test should not upload any files.
func TestExcludeBrokenSymlink(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")

	// Creating broken symlink
	assert.NoError(t, os.Mkdir(tests.Out, 0777))
	linkToNonExistingPath := filepath.Join(tests.Out, "link_to_non_existing_path")
	err := os.Symlink("non_existing_path", linkToNonExistingPath)
	assert.NoError(t, err)

	// This command should succeed because all artifacts are excluded.
	runRt(t, "u", filepath.Join(tests.Out, "*"), tests.RtRepo1, "--symlinks=true", "--exclusions=*")
	cleanArtifactoryTest()
}

// Upload symlink to Artifactory using wildcard pattern and the link content checksum
// Download the symlink which was uploaded.
// validate the symlink content checksum.
func TestSymlinkWildcardPathHandling(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a/", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link*")
	runRt(t, "u", link1, tests.RtRepo1, "--symlinks=true", "--flat=true")
	clientTestUtils.RemoveAndAssert(t, link)
	runRt(t, "dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true")
	validateSymLink(link, localFile, t)
	clientTestUtils.RemoveAndAssert(t, link)
	cleanArtifactoryTest()
}

func TestUploadWithArchiveAndSymlink(t *testing.T) {
	initArtifactoryTest(t, "")
	// Path to local file with a different name from symlinkTarget
	testFile := filepath.Join(tests.GetTestResourcesPath(), "a", "a1.in")
	tmpDir, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	err := fileutils.CopyFile(tmpDir, testFile)
	assert.NoError(t, err)
	// Link valid symLink to local file
	symlinkTarget := filepath.Join(tmpDir, "a1.in")
	err = os.Symlink(symlinkTarget, filepath.Join(tmpDir, "symlink"))
	assert.NoError(t, err)
	// Upload symlink and local file to artifactory
	runRt(t, "u", tmpDir+"/*", tests.RtRepo1+"/test-archive.zip", "--archive=zip", "--symlinks=true", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, tmpDir)
	assert.NoError(t, os.Mkdir(tmpDir, 0777))
	runRt(t, "download", tests.RtRepo1+"/test-archive.zip", tmpDir+"/", "--explode=true")
	// Validate
	assert.True(t, fileutils.IsPathExists(filepath.Join(tmpDir, "a1.in"), false), "Failed to download file from Artifactory")
	validateSymLink(filepath.Join(tmpDir, "symlink"), symlinkTarget, t)

	cleanArtifactoryTest()
}

func TestUploadWithArchiveAndSymlinkZipSlip(t *testing.T) {
	initArtifactoryTest(t, "")
	symlinkTarget := filepath.Join(tests.GetTestResourcesPath(), "a", "a2.in")
	tmpDir, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	// Link symLink to local file, outside the extraction directory
	err := os.Symlink(symlinkTarget, filepath.Join(tmpDir, "symlink"))
	assert.NoError(t, err)

	// Upload symlink and local file to artifactory
	runRt(t, "u", tmpDir+"/*", tests.RtRepo1+"/test-archive.zip", "--archive=zip", "--symlinks=true", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, tmpDir)
	assert.NoError(t, os.Mkdir(tmpDir, 0777))

	// Discard output logging to prevent negative logs
	previousLogger := tests.RedirectLogOutputToNil()
	defer log.SetLogger(previousLogger)

	// Make sure download failed
	err = artifactoryCli.Exec("download", tests.RtRepo1+"/test-archive.zip", tmpDir+"/", "--explode=true")
	assert.Error(t, err)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirHandling(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	runRt(t, "u", link, tests.RtRepo1, "--symlinks=true", "--recursive=true", "--flat=true")
	clientTestUtils.RemoveAndAssert(t, link)
	runRt(t, "dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/")
	validateSymLink(link, localFile, t)
	clientTestUtils.RemoveAndAssert(t, link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirWildcardHandling(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "lin*")
	runRt(t, "u", link1, tests.RtRepo1, "--symlinks=true", "--recursive=true", "--flat=true", "--flat=true")
	clientTestUtils.RemoveAndAssert(t, link)
	runRt(t, "dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/")
	validateSymLink(link, localFile, t)
	clientTestUtils.RemoveAndAssert(t, link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
// The test create circular links and the test suppose to prune the circular searching.
func TestSymlinkInsideSymlinkDirWithRecursionIssueUpload(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	localDirPath := filepath.Join(tests.GetTestResourcesPath(), "a")
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link1")
	err := os.Symlink(localDirPath, link1)
	assert.NoError(t, err)
	localFilePath := filepath.Join(tests.GetTestResourcesPath()+"a/", "a1.in")
	link2 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link2")
	err = os.Symlink(localFilePath, link2)
	assert.NoError(t, err)

	runRt(t, "u", localDirPath+"/link*", tests.RtRepo1, "--symlinks=true", "--recursive=true", "--flat=true")
	clientTestUtils.RemoveAndAssert(t, link1)

	clientTestUtils.RemoveAndAssert(t, link2)

	runRt(t, "dl", tests.RtRepo1+"/link*", tests.GetTestResourcesPath()+"a/")
	validateSymLink(link1, localDirPath, t)
	clientTestUtils.RemoveAndAssert(t, link1)
	validateSymLink(link2, localFilePath, t)
	clientTestUtils.RemoveAndAssert(t, link2)
	cleanArtifactoryTest()
}

func validateSymLink(localLinkPath, localFilePath string, t *testing.T) {
	// In macOS, localFilePath may lead to /var/folders/dn instead of /private/var/folders/dn.
	// EvalSymlinks on localFilePath will fix it a head of the comparison at the end of this function.
	localFilePath, err := filepath.EvalSymlinks(localFilePath)
	assert.NoError(t, err)

	exists := fileutils.IsPathSymlink(localLinkPath)
	assert.True(t, exists, "failed to download symlinks from artifactory")
	symlinks, err := filepath.EvalSymlinks(localLinkPath)
	assert.NoError(t, err, "can't eval symlinks")
	assert.Equal(t, localFilePath, symlinks, "Symlinks wasn't created as expected")
}

func TestArtifactoryDeleteNoSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	testArtifactorySimpleDelete(t, "")
}

func TestArtifactoryDeleteBySpec(t *testing.T) {
	initArtifactoryTest(t, "")
	deleteSpecPath, err := tests.CreateSpec(tests.DeleteSimpleSpec)
	assert.NoError(t, err)
	testArtifactorySimpleDelete(t, deleteSpecPath)
}

func testArtifactorySimpleDelete(t *testing.T, deleteSpecPath string) {
	preUploadBasicTestResources(t)

	// Verify exists before deleting
	searchSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchSpec, serverDetails, t)

	if deleteSpecPath != "" {
		runRt(t, "delete", "--spec="+deleteSpecPath)
	} else {
		runRt(t, "delete", tests.RtRepo1+"/test_resources/b/*")
	}

	inttestutils.VerifyExistInArtifactory(tests.GetSimpleDelete(), searchSpec, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFolderWithWildcard(t *testing.T) {
	initArtifactoryTest(t, "")
	preUploadBasicTestResources(t)

	// Verify exists before deleting
	searchSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchSpec, serverDetails, t)

	runRt(t, "delete", tests.RtRepo1+"/test_resources/*/c")

	inttestutils.VerifyExistInArtifactory(tests.GetDeleteFolderWithWildcard(), searchSpec, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFolderCompletelyNoSpec(t *testing.T) {
	testArtifactoryDeleteFoldersNoSpec(t, false)
}

func TestArtifactoryDeleteFolderContentNoSpec(t *testing.T) {
	testArtifactoryDeleteFoldersNoSpec(t, true)
}

func testArtifactoryDeleteFoldersNoSpec(t *testing.T, contentOnly bool) {
	initArtifactoryTest(t, "")
	preUploadBasicTestResources(t)

	// Verify exists before deleting
	searchSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchSpec, serverDetails, t)

	// Delete folder
	deletePath := tests.RtRepo1 + "/test_resources"
	// End with separator if content only
	if contentOnly {
		deletePath += "/"
	}
	runRt(t, "delete", deletePath)

	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	// Verify folder exists only if content only
	var expectedStatusCode int
	if contentOnly {
		expectedStatusCode = http.StatusOK
	} else {
		expectedStatusCode = http.StatusNotFound
	}
	resp, body, _, err := client.SendGet(serverDetails.ArtifactoryUrl+"api/storage/"+tests.RtRepo1+"/test_resources", true, artHttpDetails, "")
	assert.NoError(t, err)
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "test_resources shouldn't be deleted: "+tests.RtRepo1+"/test_resources/ "+string(body))

	// Verify no content exists
	verifyDoesntExistInArtifactory(searchSpec, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFoldersBySpecAllRepo(t *testing.T) {
	testArtifactoryDeleteFoldersBySpec(t, tests.DeleteSpec)
}

func TestArtifactoryDeleteFoldersBySpecWildcardInRepo(t *testing.T) {
	testArtifactoryDeleteFoldersBySpec(t, tests.DeleteSpecWildcardInRepo)
}

func testArtifactoryDeleteFoldersBySpec(t *testing.T, specPath string) {
	initArtifactoryTest(t, "")
	preUploadBasicTestResources(t)

	// Verify exists before deleting
	searchSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchSpec, serverDetails, t)

	deleteSpecPath, err := tests.CreateSpec(specPath)
	assert.NoError(t, err)
	runRt(t, "delete", "--spec="+deleteSpecPath)

	completeSearchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyDoesntExistInArtifactory(completeSearchSpec, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteExcludeByCli(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	runRt(t, "upload", "--spec="+specFileA)
	runRt(t, "upload", "--spec="+specFileB)

	// Delete by pattern
	runRt(t, "del", tests.RtRepo1+"/data/", "--exclusions=*/*b1.in;*/*b2.in;*/*b3.in;*/*c1.in")

	// Validate files are deleted
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteExclusionsByCli(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	runRt(t, "upload", "--spec="+specFileA)
	runRt(t, "upload", "--spec="+specFileB)

	// Delete by pattern
	runRt(t, "del", tests.RtRepo1+"/data/", "--exclusions=*/*b1.in;*/*b2.in;*/*b3.in;*/*c1.in")

	// Validate files are deleted
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteExclusionsBySpec(t *testing.T) {
	initArtifactoryTest(t, "")
	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA)
	runRt(t, "upload", "--spec="+specFileB)

	specFile, err := tests.CreateSpec(tests.DelSpecExclusions)
	assert.NoError(t, err)

	// Delete by pattern
	runRt(t, "del", "--spec="+specFile)

	// Validate files are deleted
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

// Deleting files when one file name is a prefix to another in the same dir
func TestArtifactoryDeletePrefixFiles(t *testing.T) {
	initArtifactoryTest(t, "")

	// Prepare search command
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadPrefixFiles)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)

	// Delete by pattern
	runRt(t, "delete", tests.RtRepo1+"/*")

	// Validate files are deleted
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	length, err := reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, 0, length)
	readerCloseAndAssert(t, reader)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteByProps(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadWithPropsSpec)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)

	// Set properties to the directories as well (and their content)
	runRt(t, "sp", tests.RtRepo1+"/a/b/", "D=5", "--include-dirs")
	runRt(t, "sp", tests.RtRepo1+"/a/b/c/", "D=2", "--include-dirs")

	//  Set the property D=5 to c1.in, which is a different value then its directory c/
	runRt(t, "sp", tests.RtRepo1+"/a/b/c/c1.in", "D=5")

	// Prepare search command
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())

	// Delete all artifacts with D=5 but without c=3
	runRt(t, "delete", tests.RtRepo1+"/*", "--props=D=5", "--exclude-props=c=3")

	// Search all artifacts in repo1
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchResultAfterDeleteByPropsStep1())
	readerCloseAndAssert(t, readerNoDate)

	// Delete all artifacts with c=3 but without a=1
	runRt(t, "delete", tests.RtRepo1+"/*", "--props=c=3", "--exclude-props=a=1")

	// Search all artifacts in repo1
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchResultAfterDeleteByPropsStep2())
	readerCloseAndAssert(t, readerNoDate)

	// Delete all artifacts with a=1 but without b=3&c=3
	runRt(t, "delete", tests.RtRepo1+"/*", "--props=a=1", "--exclude-props=b=3;c=3")

	// Search all artifacts in repo1
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchResultAfterDeleteByPropsStep3())
	readerCloseAndAssert(t, readerNoDate)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryMultipleFileSpecsUpload(t *testing.T) {
	initArtifactoryTest(t, "")
	specFile, err := tests.CreateSpec(tests.UploadMultipleFileSpecs)
	assert.NoError(t, err)
	resultSpecFile, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)

	inttestutils.VerifyExistInArtifactory(tests.GetMultipleFileSpecs(), resultSpecFile, serverDetails, t)
	verifyExistInArtifactoryByProps([]string{tests.RtRepo1 + "/multiple/properties/testdata/a/b/b2.in"}, tests.RtRepo1+"/*/properties/*.in", "searchMe=true", t)
	cleanArtifactoryTest()
}

func TestArtifactorySimplePlaceHolders(t *testing.T) {
	initArtifactoryTest(t, "")
	specFile, err := tests.CreateSpec(tests.UploadSimplePlaceholders)
	assert.NoError(t, err)

	resultSpecFile, err := tests.CreateSpec(tests.SearchSimplePlaceholders)
	assert.NoError(t, err)

	runRt(t, "upload", "--spec="+specFile)

	inttestutils.VerifyExistInArtifactory(tests.GetSimplePlaceholders(), resultSpecFile, serverDetails, t)
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadRecursiveNonFlat(t *testing.T) {
	initArtifactoryTest(t, "")
	dirInnerPath := fileutils.GetFileSeparator() + filepath.Join("inner", "folder")
	canonicalPath := tests.Out + dirInnerPath

	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	runRt(t, "upload", tests.Out+"/(*)", tests.RtRepo1+"/{1}/", "--include-dirs=true", "--recursive=true", "--flat=false")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(filepath.Join(tests.Out, "inner", "folder", "folder"), false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFlatFolderUpload(t *testing.T) {
	initArtifactoryTest(t, "")
	dirInnerPath := fileutils.GetFileSeparator() + filepath.Join("inner", "folder")
	canonicalPath := tests.Out + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	runRt(t, "upload", tests.Out+"/(*)", tests.RtRepo1+"/{1}/", "--include-dirs=true", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	// Non-flat download
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(filepath.Join(canonicalPath, "folder"), false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryIncludeDirFlatNonEmptyFolderUpload(t *testing.T) {
	initArtifactoryTest(t, "")
	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	runRt(t, "upload", tests.GetTestResourcesPath()+"a/b/*", tests.RtRepo1, "--include-dirs=true", "--flat=true")
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/c", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryDownloadNotIncludeDirs(t *testing.T) {
	initArtifactoryTest(t, "")
	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	runRt(t, "upload", tests.GetTestResourcesPath()+"*/c", tests.RtRepo1, "--include-dirs=true", "--flat=true")
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--recursive=true")
	assert.False(t, fileutils.IsPathExists(tests.Out+"/c", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryDownloadFlatTrue(t *testing.T) {
	initArtifactoryTest(t, "")
	canonicalPath := tests.GetTestResourcesPath() + path.Join("an", "empty", "folder")
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)

	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	runRt(t, "upload", tests.GetTestResourcesPath()+"(a*)/*", tests.RtRepo1+"/{1}/", "--include-dirs=true", "--flat=true")
	// Download without include-dirs
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--recursive=true", "--flat=true")
	assert.False(t, fileutils.IsPathExists(tests.Out+"/c", false), "'c' folder shouldn't exist.")

	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true", "--flat=true")
	// Inner folder with files in it
	assert.True(t, fileutils.IsPathExists(tests.Out+"/c", false), "'c' folder should exist.")
	// Empty inner folder
	assert.True(t, fileutils.IsPathExists(tests.Out+"/folder", false), "'folder' folder should exist.")
	// Folder on root with files
	assert.True(t, fileutils.IsPathExists(tests.Out+"/a$+~&^a#", false), "'a$+~&^a#' folder should exist.")
	// None bottom directory - shouldn't exist.
	assert.False(t, fileutils.IsPathExists(tests.Out+"/a", false), "'a' folder shouldn't exist.")
	// None bottom directory - shouldn't exist.
	assert.False(t, fileutils.IsPathExists(tests.Out+"/b", false), "'b' folder shouldn't exist.")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryIncludeDirFlatNonEmptyFolderUploadMatchingPattern(t *testing.T) {
	initArtifactoryTest(t, "")
	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	runRt(t, "upload", tests.GetTestResourcesPath()+"*/c", tests.RtRepo1, "--include-dirs=true", "--flat=true")
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/c", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryUploadFlatFolderWithFileAndInnerEmptyMatchingPattern(t *testing.T) {
	initArtifactoryTest(t, "")
	newFolderPath := tests.GetTestResourcesPath() + "a/b/c/d"
	err := os.MkdirAll(newFolderPath, 0777)
	assert.NoError(t, err)
	// We created an empty child folder to 'c' therefore 'c' is not longer a bottom chain and new 'd' inner directory is indeed bottom chain directory.
	// 'd' should uploaded and 'c' shouldn't
	runRt(t, "upload", tests.GetTestResourcesPath()+"a/b/*", tests.RtRepo1, "--include-dirs=true", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, newFolderPath)
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.False(t, fileutils.IsPathExists(tests.Out+"/c", false), "'c' folder shouldn't exist")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/d", false), "bottom chain directory, 'd', is missing")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryUploadFlatFolderWithFileAndInnerEmptyMatchingPatternWithPlaceHolders(t *testing.T) {
	initArtifactoryTest(t, "")
	relativePath := "/b/c/d"
	fullPath := tests.GetTestResourcesPath() + "a/" + relativePath
	err := os.MkdirAll(fullPath, 0777)
	assert.NoError(t, err)
	// We created an empty child folder to 'c' therefore 'c' is not longer a bottom chain and new 'd' inner directory is indeed bottom chain directory.
	// 'd' should uploaded and 'c' shouldn't
	runRt(t, "upload", tests.GetTestResourcesPath()+"a/(*)/*", tests.RtRepo1+"/{1}/", "--include-dirs=true", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, fullPath)
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+relativePath, false), "bottom chain directory, 'd', is missing")

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFlatFolderDownload1(t *testing.T) {
	initArtifactoryTest(t, "")
	dirInnerPath := fileutils.GetFileSeparator() + filepath.Join("inner", "folder")
	canonicalPath := tests.Out + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	// Flat true by default for upload, by using placeholder we indeed create folders hierarchy in Artifactory inner/folder/folder
	runRt(t, "upload", tests.Out+"/(*)", tests.RtRepo1+"/{1}/", "--include-dirs=true")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	// Only the inner folder should be downland e.g 'folder'
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--flat=true")
	assert.False(t, !fileutils.IsPathExists(filepath.Join(tests.Out, "folder"), false) &&
		fileutils.IsPathExists(filepath.Join(tests.Out, "inner"), false),
		"Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadRecursiveUsingSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	assert.NoError(t, createEmptyTestDir())
	specFile, err := tests.CreateSpec(tests.UploadEmptyDirs)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile)

	specFile, err = tests.CreateSpec(tests.DownloadEmptyDirs)
	assert.NoError(t, err)
	runRt(t, "download", "--spec="+specFile)
	assert.True(t, fileutils.IsPathExists(tests.Out+"/folder", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadNonRecursive(t *testing.T) {
	initArtifactoryTest(t, "")
	canonicalPath := filepath.Join(tests.Out, "inner", "folder")
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	runRt(t, "upload", tests.Out+"/", tests.RtRepo1, "--recursive=true", "--include-dirs=true", "--flat=true")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	runRt(t, "download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/folder", false), "Failed to download folder from Artifactory")
	assert.False(t, fileutils.IsPathExists(canonicalPath, false), "Path should be flat ")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderDownloadNonRecursive(t *testing.T) {
	initArtifactoryTest(t, "")
	canonicalPath := filepath.Join(tests.Out, "inner", "folder")
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	runRt(t, "upload", tests.Out+"/", tests.RtRepo1, "--recursive=true", "--include-dirs=true", "--flat=false")
	clientTestUtils.RemoveAllAndAssert(t, tests.Out)
	runRt(t, "download", tests.RtRepo1+"/*", "--recursive=false", "--include-dirs=true")
	assert.True(t, fileutils.IsPathExists(tests.Out, false), "Failed to download folder from Artifactory")
	assert.False(t, fileutils.IsPathExists(canonicalPath, false), "Path should be flat. ")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryChecksumDownload(t *testing.T) {
	initArtifactoryTest(t, "")

	var filePath = "testdata/a/a1.in"
	runRt(t, "upload", filePath, tests.RtRepo1, "--flat=true")
	testChecksumDownload(t, "/a1.in")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryChecksumDownloadRenameFileName(t *testing.T) {
	initArtifactoryTest(t, "")

	var filePath = "testdata/a/a1.in"
	runRt(t, "upload", filePath, tests.RtRepo1, "--flat=true")
	testChecksumDownload(t, "/a1.out")
	// Cleanup
	cleanArtifactoryTest()
}

func testChecksumDownload(t *testing.T, outFileName string) {
	runRt(t, "download", tests.RtRepo1+"/a1.in", tests.Out+outFileName)

	exists, err := fileutils.IsFileExists(tests.Out+outFileName, false)
	assert.NoError(t, err)
	assert.True(t, exists, "Failed to download file from Artifactory")

	firstFileInfo, _ := os.Stat(tests.Out + outFileName)
	firstDownloadTime := firstFileInfo.ModTime()

	runRt(t, "download", tests.RtRepo1+"/a1.in", tests.Out+outFileName)
	secondFileInfo, _ := os.Stat(tests.Out + outFileName)
	secondDownloadTime := secondFileInfo.ModTime()

	assert.Equal(t, firstDownloadTime, secondDownloadTime, "Checksum download failed, the file was downloaded twice")
}

func TestArtifactoryDownloadByPatternAndBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)

	// Download by build number
	runRt(t, "download", "--spec="+specFile)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryGenericBuildNameAndNumberFromEnv(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildName, tests.RtBuildName1)
	defer setEnvCallBack()
	setEnvCallBack = clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildNumber, buildNumberA)
	defer setEnvCallBack()
	runRt(t, "upload", "--spec="+specFileA)
	clientTestUtils.SetEnvAndAssert(t, coreutils.BuildNumber, "11")
	runRt(t, "upload", "--spec="+specFileB)

	// Publish buildInfo
	clientTestUtils.SetEnvAndAssert(t, coreutils.BuildNumber, buildNumberA)
	runRt(t, "build-publish")
	clientTestUtils.SetEnvAndAssert(t, coreutils.BuildNumber, buildNumberB)
	runRt(t, "build-publish")

	// Download by build number
	runRt(t, "download", "--spec="+specFile)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildNoPatternUsingSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoPattern)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)

	// Download by build number
	runRt(t, "download", "--spec="+specFile)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func prepareDownloadByBuildWithDependenciesTests(t *testing.T) {
	// Init
	initArtifactoryTest(t, "")
	buildNumber := "1337"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	// Add build artifacts.
	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)
	runRt(t, "upload", "--spec="+specFileB)

	// Add build dependencies.
	artifactoryCliNoCreds := tests.NewJfrogCli(execMain, "jfrog rt", "")
	assert.NoError(t, artifactoryCliNoCreds.Exec("bad", "--spec="+specFileB, tests.RtBuildName1, buildNumber))

	// Publish build.
	runRt(t, "build-publish", tests.RtBuildName1, buildNumber)
}

func TestArtifactoryDownloadByBuildWithDependenciesSpecNoPattern(t *testing.T) {
	prepareDownloadByBuildWithDependenciesTests(t)

	// Download with exclude-artifacts.
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecExcludeArtifacts)
	assert.NoError(t, err)
	runRt(t, "download", "--spec="+specFile)
	// Validate.
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadDoesntExist(), paths)
	assert.NoError(t, err)

	// Download deps-only.
	specFile, err = tests.CreateSpec(tests.BuildDownloadSpecDepsOnly)
	assert.NoError(t, err)
	runRt(t, "download", "--spec="+specFile)
	// Validate.
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetDownloadByBuildOnlyDeps(), paths)
	assert.NoError(t, err)

	// Download artifacts and deps.
	specFile, err = tests.CreateSpec(tests.BuildDownloadSpecIncludeDeps)
	assert.NoError(t, err)
	runRt(t, "download", "--spec="+specFile)
	// Validate.
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out+string(os.PathSeparator)+"download"+string(os.PathSeparator)+"download_build_with_dependencies", false)
	err = tests.ValidateListsIdentical(tests.GetDownloadByBuildIncludeDeps(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildWithDependencies(t *testing.T) {
	prepareDownloadByBuildWithDependenciesTests(t)

	// Download with exclude-artifacts.
	runRt(t, "download", tests.RtRepo1, "out/download/download_build_with_dependencies/", "--build="+tests.RtBuildName1, "--exclude-artifacts=true", "--flat=true")
	// Validate.
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err := tests.ValidateListsIdentical(tests.GetBuildDownloadDoesntExist(), paths)
	assert.NoError(t, err)

	// Download deps-only.
	runRt(t, "download", tests.RtRepo1, "out/download/download_build_only_dependencies/", "--build="+tests.RtBuildName1, "--exclude-artifacts=true", "--include-deps=true", "--flat=true")
	// Validate.
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetDownloadByBuildOnlyDeps(), paths)
	assert.NoError(t, err)

	// Download artifacts and deps.
	runRt(t, "download", tests.RtRepo1, "out/download/download_build_with_dependencies/", "--build="+tests.RtBuildName1, "--include-deps=true", "--flat=true")
	// Validate.
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out+string(os.PathSeparator)+"download"+string(os.PathSeparator)+"download_build_with_dependencies", false)
	err = tests.ValidateListsIdentical(tests.GetDownloadByBuildIncludeDeps(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

// Upload a file to build A.
// Verify that it doesn't exist in B.
func TestArtifactoryDownloadArtifactDoesntExistInBuild(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumber := "10"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumber)
	assert.NoError(t, err)
	// Upload a file
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a10.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumber)

	// Download from different build number
	runRt(t, "download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadDoesntExist(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

// Upload a file to 2 different builds.
// Verify that we don't download files with same sha and different build name and build number.
func TestArtifactoryDownloadByShaAndBuild(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB, buildNumberC := "10", "11", "12"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumber)
	assert.NoError(t, err)

	// Upload 3 similar files to 3 different builds
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a10.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberA)
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a11.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a12.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberC)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName2, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberC)

	// Download by build number
	runRt(t, "download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadByShaAndBuild(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	cleanArtifactoryTest()
}

// Upload a file to 2 different builds.
// Verify that we don't download files with same sha and build name and different build number.
func TestArtifactoryDownloadByShaAndBuildName(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB, buildNumberC := "10", "11", "12"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumber)
	assert.NoError(t, err)

	// Upload 3 similar files to 2 different builds
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a10.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberA)
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a11.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberB)
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a12.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberC)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberC)
	runRt(t, "build-publish", tests.RtBuildName2, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName2, buildNumberB)

	// Download by build number
	runRt(t, "download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadByShaAndBuildName(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildUsingSimpleDownload(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "b", "a"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)

	// Download by build number, a1 should not be downloaded, b1 should
	runRt(t, "download", tests.RtRepo1+"/data/a1.in", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(), "--build="+tests.RtBuildName1)
	runRt(t, "download", tests.RtRepo1+"/data/b1.in", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(), "--build="+tests.RtBuildName1)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildUsingSimpleDownloadWithProject(t *testing.T) {
	initArtifactoryProjectTest(t, projectsTokenMinArtifactoryVersion)
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	assert.NoError(t, err)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	projectKey := "prj" + timestamp[len(timestamp)-3:]
	// Delete the project if already exists
	deleteProjectIfExists(t, accessManager, projectKey)

	// Create new project
	projectParams := accessServices.ProjectParams{
		ProjectDetails: accessServices.Project{
			DisplayName: "testProject " + projectKey,
			ProjectKey:  projectKey,
		},
	}
	err = accessManager.CreateProject(projectParams)
	assert.NoError(t, err)
	// Assign the repository to the project
	err = accessManager.AssignRepoToProject(tests.RtRepo1, projectKey, true)
	assert.NoError(t, err)

	// Delete the build if exists
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	buildNumberA := "123"

	// Upload files with buildName, buildNumber and project flags
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA, "--project="+projectKey)

	// Publish buildInfo with project flag
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA, "--project="+projectKey)

	// Download by project, b1 should be downloaded
	runRt(t, "download", tests.RtRepo1+"/data/b1.in", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(),
		"--build="+tests.RtBuildName1, "--project="+projectKey)

	// Validate files are downloaded by build number
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	err = accessManager.UnassignRepoFromProject(tests.RtRepo1)
	assert.NoError(t, err)
	err = accessManager.DeleteProject(projectKey)
	assert.NoError(t, err)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadWithEnvProject(t *testing.T) {
	initArtifactoryProjectTest(t, projectsTokenMinArtifactoryVersion)
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	assert.NoError(t, err)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	projectKey := "prj" + timestamp[len(timestamp)-3:]
	// Delete the project if already exists
	deleteProjectIfExists(t, accessManager, projectKey)

	// Create new project
	projectParams := accessServices.ProjectParams{
		ProjectDetails: accessServices.Project{
			DisplayName: "testProject " + projectKey,
			ProjectKey:  projectKey,
		},
	}
	err = accessManager.CreateProject(projectParams)
	assert.NoError(t, err)
	// Assign the repository to the project
	err = accessManager.AssignRepoToProject(tests.RtRepo1, projectKey, true)
	assert.NoError(t, err)

	// Delete the build if exists
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	buildNumberA := "123"
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildName, tests.RtBuildName1)
	defer setEnvCallBack()
	setEnvCallBack = clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildNumber, buildNumberA)
	defer setEnvCallBack()
	setEnvCallBack = clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.Project, projectKey)
	defer setEnvCallBack()
	// Upload files with buildName, buildNumber and project flags
	runRt(t, "upload", "--spec="+specFileB)

	// Publish buildInfo with project flag
	runRt(t, "build-publish")

	// Download by project, b1 should be downloaded
	runRt(t, "download", tests.RtRepo1+"/data/b1.in", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(),
		"--build="+tests.RtBuildName1, "--project="+projectKey)

	// Validate files are downloaded by build number
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	err = accessManager.UnassignRepoFromProject(tests.RtRepo1)
	assert.NoError(t, err)
	err = accessManager.DeleteProject(projectKey)
	assert.NoError(t, err)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildNoPatternUsingSimpleDownload(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "b", "a"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)

	// Download by build number, a1 should not be downloaded, b1 should
	runRt(t, "download", "*", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(), "--build="+tests.RtBuildName1+"/"+buildNumberA)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownloadNoPattern(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByArchiveEntriesCli(t *testing.T) {
	initArtifactoryTest(t, "")
	uploadSpecFile, err := tests.CreateSpec(tests.ArchiveEntriesUpload)
	assert.NoError(t, err)

	// Upload archives
	runRt(t, "upload", "--spec="+uploadSpecFile)

	// Trigger archive indexing on the repo.
	triggerArchiveIndexing(t)

	// Create executor for running with retries
	retryExecutor := createRetryExecutorForArchiveEntries(tests.GetBuildArchiveEntriesDownloadCli(),
		[]string{"dl", tests.RtRepo1, "out/", "--archive-entries=(*)c1.in", "--flat=true"})

	// Perform download by archive-entries only the archives containing c1.in, and validate results
	assert.NoError(t, retryExecutor.Execute())

	// Cleanup
	cleanArtifactoryTest()
}

func triggerArchiveIndexing(t *testing.T) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)
	resp, _, err := client.SendPost(serverDetails.ArtifactoryUrl+"api/archiveIndex/"+tests.RtRepo1, []byte{}, artHttpDetails, "")
	if err != nil {
		assert.NoError(t, err, "archive indexing failed")
		return
	}
	assert.Equal(t, http.StatusAccepted, resp.StatusCode, "archive indexing failed")
	// Indexing buffer
	time.Sleep(3 * time.Second)
}

func TestArtifactoryDownloadByArchiveEntriesSpecificPathCli(t *testing.T) {
	initArtifactoryTest(t, "")
	uploadSpecFile, err := tests.CreateSpec(tests.ArchiveEntriesUpload)
	assert.NoError(t, err)

	// Upload archives
	runRt(t, "upload", "--spec="+uploadSpecFile)

	// Trigger archive indexing on the repo.
	triggerArchiveIndexing(t)

	// Create executor for running with retries
	retryExecutor := createRetryExecutorForArchiveEntries(tests.GetBuildArchiveEntriesSpecificPathDownload(),
		[]string{"dl", tests.RtRepo1, "out/", "--archive-entries=b/c/c1.in", "--flat=true"})

	// Perform download by archive-entries only the archives containing c1.in, and validate results
	assert.NoError(t, retryExecutor.Execute())

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByArchiveEntriesSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	uploadSpecFile, err := tests.CreateSpec(tests.ArchiveEntriesUpload)
	assert.NoError(t, err)
	downloadSpecFile, err := tests.CreateSpec(tests.ArchiveEntriesDownload)
	assert.NoError(t, err)

	// Upload archives
	runRt(t, "upload", "--spec="+uploadSpecFile)

	// Trigger archive indexing on the repo.
	triggerArchiveIndexing(t)

	// Create executor for running with retries
	retryExecutor := createRetryExecutorForArchiveEntries(tests.GetBuildArchiveEntriesDownloadSpec(),
		[]string{"dl", "--spec=" + downloadSpecFile})

	// Perform download by archive-entries only the archives containing d1.in, and validate results
	assert.NoError(t, retryExecutor.Execute())

	// Cleanup
	cleanArtifactoryTest()
}

func createRetryExecutorForArchiveEntries(expected []string, args []string) *clientutils.RetryExecutor {
	return &clientutils.RetryExecutor{
		MaxRetries: 120,
		// RetriesIntervalMilliSecs in milliseconds
		RetriesIntervalMilliSecs: 1 * 1000,
		ErrorMessage:             "Waiting for Artifactory to index archives...",
		ExecutionHandler: func() (bool, error) {
			// Execute the requested cli command
			err := artifactoryCli.Exec(args...)
			if err != nil {
				return true, err
			}
			err = validateDownloadByArchiveEntries(expected)
			if err != nil {
				return false, err
			}
			return false, nil
		},
	}
}

func validateDownloadByArchiveEntries(expected []string) error {
	// Validate files are downloaded as expected
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	return tests.ValidateListsIdentical(expected, paths)
}

func TestArtifactoryDownloadExcludeByCli(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--recursive=true")
	runRt(t, "upload", "--spec="+specFileB, "--recursive=true")

	// Download by pattern
	runRt(t, "download", tests.RtRepo1, "out/download/aql_by_artifacts/", "--exclusions=*/*/a1.in;*/*a2.*;*/data/c2.in")

	// Validate files are excluded
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildExcludeDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExclusionsByCli(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--recursive=true")
	runRt(t, "upload", "--spec="+specFileB, "--recursive=true")

	// Download by pattern
	runRt(t, "download", tests.RtRepo1, "out/download/aql_by_artifacts/", "--exclusions=*/*/a1.in;*/*a2.*;*/data/c2.in")

	// Validate files are excluded
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildExcludeDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExclusionsBySpec(t *testing.T) {
	initArtifactoryTest(t, "")
	specFile, err := tests.CreateSpec(tests.DownloadSpecExclusions)
	assert.NoError(t, err)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--recursive=true", "--flat=false")
	runRt(t, "upload", "--spec="+specFileB, "--recursive=true", "--flat=false")

	// Download by spec
	runRt(t, "download", "--spec="+specFile)

	// Validate files are excluded
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildExcludeDownloadBySpec(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExclusionsBySpecOverride(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--recursive=true", "--flat=false")
	runRt(t, "upload", "--spec="+specFileB, "--recursive=true", "--flat=false")

	// Download by spec
	specFile, err := tests.CreateSpec(tests.DownloadSpecExclusions)
	assert.NoError(t, err)
	runRt(t, "download", "--spec="+specFile, "--exclusions=*a1.in;*a2.in;*c2.in")

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildExcludeDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

// Sort and limit changes the way properties are used so this should be tested with symlinks and search by build

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded with limit param.
// validate the symlink content checksum.
func TestArtifactoryLimitWithSymlink(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	runRt(t, "u", link, tests.RtRepo1, "--symlinks=true", "--flat=true")
	clientTestUtils.RemoveAndAssert(t, link)
	runRt(t, "dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true", "--limit=1")
	validateSortLimitWithSymLink(link, localFile, t)
	clientTestUtils.RemoveAndAssert(t, link)
	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded with limit param.
// validate the symlink content checksum.
func TestArtifactorySortWithSymlink(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t, "")
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	runRt(t, "u", link, tests.RtRepo1, "--symlinks=true", "--flat=true")
	clientTestUtils.RemoveAndAssert(t, link)
	runRt(t, "dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true", "--sort-by=created")
	validateSortLimitWithSymLink(link, localFile, t)
	clientTestUtils.RemoveAndAssert(t, link)
	cleanArtifactoryTest()
}

func validateSortLimitWithSymLink(localLinkPath, localFilePath string, t *testing.T) {
	exists := fileutils.IsPathSymlink(localLinkPath)
	assert.True(t, exists, "failed to download symlinks from artifactory with Sort/Limit flag")
	symlinks, err := filepath.EvalSymlinks(localLinkPath)
	assert.NoError(t, err, "can't eval symlinks with Sort/Limit flag")
	assert.Equal(t, localFilePath, symlinks, "Symlinks wasn't created as expected with Sort/Limit flag")
}

// Upload a file to 2 different builds.
// Verify that we don't download files with same sha and build name and different build number when sort is configured.
func TestArtifactoryDownloadByShaAndBuildNameWithSort(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB, buildNumberC := "10", "11", "12"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumberWithSort)
	assert.NoError(t, err)
	// Upload 3 similar files to 2 different builds
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a10.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberA)
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a11.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberB)
	runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a12.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberC)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberC)
	runRt(t, "build-publish", tests.RtBuildName2, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName2, buildNumberB)

	// Download by build number
	runRt(t, "download", "--sort-by=created", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(filepath.Join(tests.Out, "download", "sort_limit_by_build"), false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadByShaAndBuildNameWithSort(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber: a* uploaded with build number "10", b* uploaded with build number "11"
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)

	// Copy by build name "cli-test-build" and build number "10" from spec, a* should be copied
	runRt(t, "copy", "--spec="+specFile)

	// Validate files are Copied by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildPatternAllUsingSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildPatternAllSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber: a* uploaded with build number "10", b* uploaded with build number "11"
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)

	// Copy by build name "cli-test-build" and build number "10" from spec, a* should be copied
	runRt(t, "copy", "--spec="+specFile)

	// Validate files are Copied by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactorySortAndLimit(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload all testdata/a/ files
	runRt(t, "upload", "testdata/a/(*)", tests.RtRepo1+"/data/{1}")

	// Download 1 sorted by name asc
	runRt(t, "download", tests.RtRepo1+"/data/", "out/download/sort_limit/", "--sort-by=name", "--limit=1")

	// Download 3 sorted by depth desc
	runRt(t, "download", tests.RtRepo1+"/data/", "out/download/sort_limit/", "--sort-by=depth", "--limit=3", "--sort-order=desc")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err := tests.ValidateListsIdentical(tests.GetSortAndLimit(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactorySortByCreated(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files separately so we can sort by created.
	runRt(t, "upload", "testdata/created/or", tests.RtRepo1, `--target-props=k1=v1`, "--flat=true")
	runRt(t, "upload", "testdata/created/o", tests.RtRepo1, "--flat=true")
	runRt(t, "upload", "testdata/created/org", tests.RtRepo1, "--flat=true")

	// Prepare search command
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).SortBy([]string{"created"}).SortOrder("asc").Limit(3)
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	reader, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)

	var resultItems []utils.SearchResult
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)

	}
	assert.Len(t, resultItems, 3)
	// Verify the sort by checking if the item results are ordered by asc.
	assert.True(t, reflect.DeepEqual(resultItems[0], tests.GetFirstSearchResultSortedByAsc()))
	assert.True(t, reflect.DeepEqual(resultItems[1], tests.GetSecondSearchResultSortedByAsc()))
	assert.True(t, reflect.DeepEqual(resultItems[2], tests.GetThirdSearchResultSortedByAsc()))

	readerCloseAndAssert(t, reader)
	searchCmd.SetSpec(searchSpecBuilder.SortOrder("desc").BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	reader, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	resultItems = nil
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.Len(t, resultItems, 3)
	// Verify the sort by checking if the item results are ordered by desc.
	assert.True(t, reflect.DeepEqual(resultItems[2], tests.GetFirstSearchResultSortedByAsc()))
	assert.True(t, reflect.DeepEqual(resultItems[1], tests.GetSecondSearchResultSortedByAsc()))
	assert.True(t, reflect.DeepEqual(resultItems[0], tests.GetThirdSearchResultSortedByAsc()))
	readerCloseAndAssert(t, reader)

	// Cleanup
	cleanArtifactoryTest()
}
func TestArtifactoryOffset(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload all testdata/a/ files
	runRt(t, "upload", "testdata/a/*", path.Join(tests.RtRepo1, "offset_test")+"/", "--flat=true")

	// Downloading files one by one, to check that the offset is working as expected.
	// Download only the first file, expecting to download a1.in
	runRt(t, "download", tests.RtRepo1+"/offset_test/", tests.Out+"/", "--flat=true", "--sort-by=name", "--limit=1", "--offset=0")
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a1.in")}, paths, t)

	// Download the second file, expecting to download a2.in
	runRt(t, "download", tests.RtRepo1+"/offset_test/", tests.Out+"/", "--flat=true", "--sort-by=name", "--limit=1", "--offset=1")
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a2.in")}, paths, t)

	// Download the third file, expecting to download a3.in
	runRt(t, "download", tests.RtRepo1+"/offset_test/", tests.Out+"/", "--flat=true", "--sort-by=name", "--limit=1", "--offset=2")
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a3.in")}, paths, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildOverridingByInlineFlag(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber: b* uploaded with build number "10", a* uploaded with build number "11"
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)

	// Copy by build number: using override of build by flag from inline (no number set so LATEST build should be copied), a* should be copied
	runRt(t, "copy", "--build="+tests.RtBuildName1, "--spec="+specFile)

	// Validate files are Copied by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryMoveByBuildUsingFlags(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)

	// Move by build name and number
	runRt(t, "move", "--build="+tests.RtBuildName1+"/11", "--spec="+specFile)

	// Validate files are moved by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryMoveNoSpec(t *testing.T) {
	testCopyMoveNoSpec("mv", tests.GetBuildBeforeMoveExpected(), tests.GetBuildMoveExpected(), t)
}

func TestArtifactoryMoveExcludeByCli(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA)
	runRt(t, "upload", "--spec="+specFileB)

	// Move by pattern
	runRt(t, "move", tests.RtRepo1+"/data/", tests.RtRepo2+"/", "--exclusions=*/*b*;*/*c*")

	// Validate excluded files didn't move
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryMoveExclusionsByCli(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA)
	runRt(t, "upload", "--spec="+specFileB)

	// Move by pattern
	runRt(t, "move", tests.RtRepo1+"/data/", tests.RtRepo2+"/", "--exclusions=*/*b*;*/*c*")

	// Validate excluded files didn't move
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryMoveExclusionsBySpec(t *testing.T) {
	initArtifactoryTest(t, "")
	specFile, err := tests.CreateSpec(tests.MoveCopySpecExclusions)
	assert.NoError(t, err)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA)
	runRt(t, "upload", "--spec="+specFileB)

	// Move by spec
	runRt(t, "move", "--spec="+specFile)

	// Validate excluded files didn't move
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteByLatestBuild(t *testing.T) {
	initArtifactoryTest(t, "")
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	runRt(t, "upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberA)
	runRt(t, "build-publish", tests.RtBuildName1, buildNumberB)

	// Delete by build name and LATEST
	runRt(t, "delete", "--build="+tests.RtBuildName1+"/LATEST", "--spec="+specFile)

	// Validate files are deleted by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	inttestutils.VerifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, serverDetails, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestGitLfsCleanup(t *testing.T) {
	initArtifactoryTest(t, "")
	var filePath = "testdata/gitlfs/(4b)(*)"
	runRt(t, "upload", filePath, tests.RtLfsRepo+"/objects/4b/f4/{2}{1}", "--flat=true")
	runRt(t, "upload", filePath, tests.RtLfsRepo+"/objects/4b/f4/", "--flat=true")
	refs := path.Join("refs", "remotes", "*")
	dotGitPath := getCliDotGitPath(t)
	runRt(t, "glc", dotGitPath, "--repo="+tests.RtLfsRepo, "--refs=HEAD,"+refs)
	gitlfsSpecFile, err := tests.CreateSpec(tests.GitLfsAssertSpec)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetGitLfsExpected(), gitlfsSpecFile, serverDetails, t)
	cleanArtifactoryTest()
}

func TestPing(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "ping")
	cleanArtifactoryTest()
}

type summaryExpected struct {
	errors  bool
	status  string
	success int64
	failure int64
}

func TestSummaryReport(t *testing.T) {
	initArtifactoryTest(t, "")

	uploadSpecFile, err := tests.CreateSpec(tests.UploadFlatNonRecursive)
	assert.NoError(t, err)
	downloadSpecFile, err := tests.CreateSpec(tests.DownloadAllRepo1TestResources)
	assert.NoError(t, err)

	argsMap := map[string][]string{
		"upload":       {"--spec=" + uploadSpecFile},
		"move":         {path.Join(tests.RtRepo1, "*.in"), tests.RtRepo2 + "/"},
		"copy":         {path.Join(tests.RtRepo2, "*.in"), tests.RtRepo1 + "/"},
		"delete":       {path.Join(tests.RtRepo2, "*.in")},
		"set-props":    {path.Join(tests.RtRepo1, "*.in"), "prop=val"},
		"delete-props": {path.Join(tests.RtRepo1, "*.in"), "prop"},
		"download":     {"--spec=" + downloadSpecFile},
	}
	expected := summaryExpected{
		false,
		"success",
		3,
		0,
	}
	testSummaryReport(t, argsMap, expected)
}

func TestSummaryReportFailNoOpTrue(t *testing.T) {
	initArtifactoryTest(t, "")
	testFailNoOpSummaryReport(t, true)
}

func TestSummaryReportFailNoOpFalse(t *testing.T) {
	initArtifactoryTest(t, "")
	testFailNoOpSummaryReport(t, false)
}

// Test summary after commands that do no actions, with/without failNoOp flag.
func testFailNoOpSummaryReport(t *testing.T, failNoOp bool) {
	initArtifactoryTest(t, "")

	nonExistingSource := "./*non/existing/source/*"
	nonExistingDest := "non/existing/dest/"
	failNoOpFlag := "--fail-no-op=" + strconv.FormatBool(failNoOp)

	argsMap := map[string][]string{
		"upload":       {nonExistingSource, nonExistingDest, failNoOpFlag},
		"move":         {nonExistingSource, nonExistingDest, failNoOpFlag},
		"copy":         {nonExistingSource, nonExistingDest, failNoOpFlag},
		"delete":       {nonExistingSource, failNoOpFlag},
		"set-props":    {nonExistingSource, "prop=val", failNoOpFlag},
		"delete-props": {nonExistingSource, "prop", failNoOpFlag},
		"download":     {nonExistingSource, nonExistingDest, failNoOpFlag},
	}
	expected := summaryExpected{
		failNoOp,
		"success",
		0,
		0,
	}
	if failNoOp {
		expected.status = "failure"
	}
	testSummaryReport(t, argsMap, expected)
}

func testSummaryReport(t *testing.T, argsMap map[string][]string, expected summaryExpected) {
	buffer, _, previousLog := tests.RedirectLogOutputToBuffer()
	// Restore previous logger when the function returns
	defer log.SetLogger(previousLog)

	for _, cmd := range []string{"upload", "move", "copy", "delete", "set-props", "delete-props", "download"} {
		// Execute the cmd with it's args.
		err := artifactoryCli.Exec(append([]string{cmd}, argsMap[cmd]...)...)
		verifySummary(t, buffer, previousLog, err, expected)
	}
	cleanArtifactoryTest()
}

func TestUploadDeploymentView(t *testing.T) {
	initArtifactoryTest(t, "")
	assertPrintedDeploymentViewFunc, cleanupFunc := initDeploymentViewTest(t)
	defer cleanupFunc()
	uploadCmd := generic.NewUploadCommand()
	fileSpec := spec.NewBuilder().Pattern(filepath.Join("testdata", "a", "a*.in")).Target(tests.RtRepo1).BuildSpec()
	uploadCmd.SetUploadConfiguration(createUploadConfiguration()).SetSpec(fileSpec).SetServerDetails(serverDetails)
	assert.NoError(t, artifactoryCli.Exec("upload", filepath.Join("testdata", "a", "a*.in"), tests.RtRepo1))
	assertPrintedDeploymentViewFunc()
	cleanArtifactoryTest()
}

func TestUploadDeploymentViewWithArchive(t *testing.T) {
	initArtifactoryTest(t, "")
	assertPrintedDeploymentViewFunc, cleanupFunc := initDeploymentViewTest(t)
	defer cleanupFunc()
	assert.NoError(t, artifactoryCli.Exec("upload", filepath.Join("testdata", "a", "a*.in"), path.Join(tests.RtRepo1, "z.zip"), "--archive", "zip"))
	assertPrintedDeploymentViewFunc()
	cleanArtifactoryTest()
}

func TestUploadDetailedSummary(t *testing.T) {
	initArtifactoryTest(t, "")
	uploadCmd := generic.NewUploadCommand()
	fileSpec := spec.NewBuilder().Pattern(filepath.Join("testdata", "a", "a*.in")).Target(tests.RtRepo1).BuildSpec()
	uploadCmd.SetUploadConfiguration(createUploadConfiguration()).SetSpec(fileSpec).SetServerDetails(serverDetails).SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(uploadCmd))
	result := uploadCmd.Result()
	reader := result.Reader()
	readerGetErrorAndAssert(t, reader)
	defer readerCloseAndAssert(t, reader)
	var files []clientutils.FileTransferDetails
	for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
		files = append(files, *transferDetails)
	}
	assert.ElementsMatch(t, files, tests.GetExpectedUploadSummaryDetails(*tests.JfrogUrl+tests.ArtifactoryEndpoint))
	cleanArtifactoryTest()
}

func createUploadConfiguration() *utils.UploadConfiguration {
	uploadConfiguration := new(utils.UploadConfiguration)
	uploadConfiguration.Threads = cliutils.Threads
	return uploadConfiguration
}

func TestArtifactoryBuildDiscard(t *testing.T) {
	// Initialize
	initArtifactoryTest(t, "")
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber
	for i := 1; i <= 5; i++ {
		runRt(t, "upload", "testdata/a/a1.in", tests.RtRepo1+"/data/", "--build-name="+tests.RtBuildName1, "--build-number="+strconv.Itoa(i))
		runRt(t, "build-publish", tests.RtBuildName1, strconv.Itoa(i))
	}

	// Test discard by max-builds
	runRt(t, "build-discard", tests.RtBuildName1, "--max-builds=3")
	jsonResponse := getAllBuildsByBuildName(client, tests.RtBuildName1, t, http.StatusOK)
	assert.Len(t, jsonResponse.Builds, 3, "Incorrect operation of build-discard by max-builds.")

	// Test discard with exclusion
	runRt(t, "build-discard", tests.RtBuildName1, "--max-days=-1", "--exclude-builds=3,5")
	jsonResponse = getAllBuildsByBuildName(client, tests.RtBuildName1, t, http.StatusOK)
	assert.Len(t, jsonResponse.Builds, 2, "Incorrect operation of build-discard with exclusion.")

	// Test discard by max-days
	runRt(t, "build-discard", tests.RtBuildName1, "--max-days=-1")
	jsonResponse = getAllBuildsByBuildName(client, tests.RtBuildName1, t, http.StatusNotFound)
	assert.Zero(t, jsonResponse, "Incorrect operation of build-discard by max-days.")

	//Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

// Tests compatibility to file paths with windows separators.
// Verifies the upload and download commands work as expected for inputs of both arguments and spec files.
func TestArtifactoryWinBackwardsCompatibility(t *testing.T) {
	initArtifactoryTest(t, "")
	if !coreutils.IsWindows() {
		t.Skip("Not running on Windows, skipping...")
	}
	uploadSpecFile, err := tests.CreateSpec(tests.WinSimpleUploadSpec)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+uploadSpecFile)
	runRt(t, "upload", "testdata\\\\a\\\\b\\\\*", tests.RtRepo1+"/compatibility_arguments/", "--exclusions=*b2.in;*c*")

	downloadSpecFile, err := tests.CreateSpec(tests.WinSimpleDownloadSpec)
	assert.NoError(t, err)
	runRt(t, "download", "--spec="+downloadSpecFile)
	runRt(t, "download", tests.RtRepo1+"/*arguments*", "out\\\\win\\\\", "--flat=true")

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetWinCompatibility(), paths)
	assert.NoError(t, err)
	cleanArtifactoryTest()
}

func TestArtifactorySearchIncludeDir(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFileA, "--recursive", "--flat=false")

	// Prepare search command
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)
	searchCmd.SetSpec(searchSpecBuilder.IncludeDirs(false).BuildSpec())
	reader, err := searchCmd.Search()
	assert.NoError(t, err)

	// Search without IncludeDirs
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchNotIncludeDirsFiles())
	readerCloseAndAssert(t, readerNoDate)

	// Search with IncludeDirs
	searchCmd.SetSpec(searchSpecBuilder.IncludeDirs(true).BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchIncludeDirsFiles())
	readerCloseAndAssert(t, readerNoDate)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactorySearchProps(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadWithPropsSpec)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile, "--recursive")

	// Prepare search command
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)

	// Search artifacts with c=3
	searchCmd.SetSpec(searchSpecBuilder.Props("c=3").BuildSpec())
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep1())
	readerCloseAndAssert(t, readerNoDate)

	// Search artifacts without c=3
	searchCmd.SetSpec(searchSpecBuilder.Props("").ExcludeProps("c=3").BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep2())
	readerCloseAndAssert(t, readerNoDate)

	// Search artifacts without a=1&b=2
	searchCmd.SetSpec(searchSpecBuilder.Props("").ExcludeProps("a=1;b=2").BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep3())
	readerCloseAndAssert(t, readerNoDate)

	// Search artifacts without a=1&b=2 and with c=3
	searchCmd.SetSpec(searchSpecBuilder.Props("c=3").ExcludeProps("a=1;b=2").BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep4())
	readerCloseAndAssert(t, readerNoDate)

	// Search artifacts without a=1 and with c=5
	searchCmd.SetSpec(searchSpecBuilder.Props("c=5").ExcludeProps("a=1").BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep5())
	readerCloseAndAssert(t, readerNoDate)

	// Search artifacts by pattern "*b*", exclude pattern "*3*", with "b=1" and without "c=3"
	pattern := tests.RtRepo1 + "/*b*"
	exclusions := []string{tests.RtRepo1 + "/*3*"}
	searchSpecBuilder = spec.NewBuilder().Pattern(pattern).Recursive(true).Exclusions(exclusions).Props("b=1").ExcludeProps("c=3")
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	readerGetErrorAndAssert(t, reader)
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep6())
	readerCloseAndAssert(t, readerNoDate)

	// Cleanup
	cleanArtifactoryTest()
}

// Remove not to be deleted dirs from delete command from path to delete.
func TestArtifactoryDeleteExcludeProps(t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadWithPropsSpecdeleteExcludeProps)
	assert.NoError(t, err)
	runRt(t, "upload", "--spec="+specFile, "--recursive")

	// Prepare search command
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)

	// Delete all artifacts without c=1 but keep dirs that has at least one artifact with c=1 props
	runRt(t, "delete", tests.RtRepo1+"/*", "--exclude-props=c=1")

	// Search artifacts with c=1
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	resultItems := []utils.SearchResult{}
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	assert.ElementsMatch(t, resultItems, tests.GetSearchAfterDeleteWithExcludeProps())
	readerCloseAndAssert(t, readerNoDate)

	// Cleanup
	cleanArtifactoryTest()
}

func getAllBuildsByBuildName(client *httpclient.HttpClient, buildName string, t *testing.T, expectedHttpStatusCode int) buildsApiResponseStruct {
	resp, body, _, _ := client.SendGet(serverDetails.ArtifactoryUrl+"api/build/"+buildName, true, artHttpDetails, "")
	assert.Equal(t, expectedHttpStatusCode, resp.StatusCode, "Failed retrieving build information from artifactory.")

	buildsApiResponse := &buildsApiResponseStruct{}
	err := json.Unmarshal(body, buildsApiResponse)
	assert.NoError(t, err, "Unmarshalling failed with an error")
	return *buildsApiResponse
}

type buildsApiInnerBuildsStruct struct {
	Uri     string `json:"uri,omitempty"`
	Started string `json:"started,omitempty"`
}

type buildsApiResponseStruct struct {
	Uri    string                       `json:"uri,omitempty"`
	Builds []buildsApiInnerBuildsStruct `json:"buildsNumbers,omitempty"`
}

func verifySummary(t *testing.T, buffer *bytes.Buffer, logger log.Log, cmdError error, expected summaryExpected) {
	if expected.errors {
		assert.Error(t, cmdError)
	} else {
		assert.NoError(t, cmdError)
	}

	output := buffer.Bytes()
	buffer.Reset()
	logger.Output(string(output))

	status, err := jsonparser.GetString(output, "status")
	assert.NoError(t, err)
	assert.Equal(t, expected.status, status, "Summary validation failed")

	resultSuccess, err := jsonparser.GetInt(output, "totals", "success")
	assert.NoError(t, err)

	resultFailure, err := jsonparser.GetInt(output, "totals", "failure")
	assert.NoError(t, err)

	assert.Equal(t, expected.success, resultSuccess, "Summary validation failed")
	assert.Equal(t, expected.failure, resultFailure, "Summary validation failed")
}

func CleanArtifactoryTests() {
	cleanArtifactoryTest()
	deleteCreatedRepos()
}

func initArtifactoryTest(t *testing.T, minVersion string) {
	if !*tests.TestArtifactory {
		t.Skip("Skipping artifactory test. To run artifactory test add the '-test.artifactory=true' option.")
	}
	if minVersion != "" {
		validateArtifactoryVersion(t, minVersion)
	}
}

func initArtifactoryProjectTest(t *testing.T, minVersion string) {
	if !*tests.TestArtifactoryProject {
		t.Skip("Skipping artifactory project test. To run artifactory test add the '-test.artifactoryProject=true' option.")
	}
	if minVersion != "" {
		validateArtifactoryVersion(t, minVersion)
	}
}

func validateArtifactoryVersion(t *testing.T, minVersion string) {
	rtVersion, err := getArtifactoryVersion()
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !rtVersion.AtLeast(minVersion) {
		t.Skip("Skipping artifactory project test. Artifactory version not supported.")
	}
}

func getArtifactoryVersion() (version.Version, error) {
	rtVersion, err := artAuth.GetVersion()
	return *version.NewVersion(rtVersion), err
}

func cleanArtifactoryTest() {
	if !*tests.TestArtifactory {
		return
	}
	log.Info("Cleaning test data...")
	cleanArtifactory()
	tests.CleanFileSystem()
}

func preUploadBasicTestResources(t *testing.T) {
	uploadPath := tests.GetTestResourcesPath() + "a/(.*)"
	targetPath := tests.RtRepo1 + "/test_resources/{1}"
	runRt(t, "upload", uploadPath, targetPath,
		"--threads=10", "--regexp=true", "--target-props=searchMe=true", "--flat=false")
}

func execDeleteRepo(repoName string) {
	err := artifactoryCli.Exec("repo-delete", repoName, "--quiet")
	if err != nil {
		log.Error("Couldn't delete repository", repoName, ":", err.Error())
	}
}

func execDeleteUser(username string) {
	err := artifactoryCli.Exec("users-delete", username, "--quiet")
	if err != nil {
		log.Error("Couldn't delete user", username, ":", err.Error())
	}
}

func getAllRepos() (repositoryKeys []string, err error) {
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return nil, err
	}
	repos, err := servicesManager.GetAllRepositories()
	if err != nil {
		return nil, err
	}
	for _, repo := range *repos {
		repositoryKeys = append(repositoryKeys, repo.Key)
	}
	return
}

func execListBuildNamesRest() ([]string, error) {
	var buildNames []string

	// Build http client
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return nil, err
	}

	// Send get request
	resp, body, _, err := client.SendGet(serverDetails.ArtifactoryUrl+"api/build", true, artHttpDetails, "")
	if err != nil {
		return nil, err
	}

	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK, http.StatusCreated); err != nil {
		return nil, err
	}

	builds, _, _, err := jsonparser.Get(body, "builds")
	if err != nil {
		return nil, err
	}

	// Extract repository keys from the json response
	var keyError error
	_, err = jsonparser.ArrayEach(builds, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil || keyError != nil {
			return
		}
		buildName, err := jsonparser.GetString(value, "uri")
		if err != nil {
			keyError = err
			return
		}
		buildNames = append(buildNames, strings.TrimPrefix(buildName, "/"))
	})
	if keyError != nil {
		return nil, err
	}

	return buildNames, err
}

func execCreateRepoRest(repoConfig, repoName string) {
	output, err := ioutil.ReadFile(repoConfig)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	rtutils.AddHeader("Content-Type", "application/json", &artHttpDetails.Headers)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	resp, body, err := client.SendPut(serverDetails.ArtifactoryUrl+"api/repositories/"+repoName, output, artHttpDetails, "")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK, http.StatusCreated); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.Info("Repository", repoName, "created.")
}

func getAllUsernames() (usernames []string, err error) {
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return nil, err
	}
	users, err := servicesManager.GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		usernames = append(usernames, user.Name)
	}
	return
}

func createRequiredRepos() {
	tests.CreatedNonVirtualRepositories = tests.GetNonVirtualRepositories()
	createRepos(tests.CreatedNonVirtualRepositories)
	tests.CreatedVirtualRepositories = tests.GetVirtualRepositories()
	createRepos(tests.CreatedVirtualRepositories)
}

func cleanUpOldBuilds() {
	tests.CleanUpOldItems(tests.GetBuildNames(), execListBuildNamesRest, func(buildName string) {
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	})
}

func cleanUpOldRepositories() {
	tests.CleanUpOldItems(tests.GetAllRepositoriesNames(), getAllRepos, execDeleteRepo)
}

func cleanUpOldUsers() {
	tests.CleanUpOldItems(tests.GetTestUsersNames(), getAllUsernames, execDeleteUser)
}

func createRepos(repos map[*string]string) {
	for repoName, configFile := range repos {
		if !isRepoExist(*repoName) {
			repoConfig := tests.GetTestResourcesPath() + configFile
			repoConfig, err := tests.ReplaceTemplateVariables(repoConfig, "")
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			execCreateRepoRest(repoConfig, *repoName)
		}
	}
}

func deleteCreatedRepos() {
	// Important - Virtual repositories must be deleted first
	deleteRepos(tests.CreatedVirtualRepositories)
	deleteRepos(tests.CreatedNonVirtualRepositories)
}

func deleteRepos(repos map[*string]string) {
	for repoName := range repos {
		if isRepoExist(*repoName) {
			execDeleteRepo(*repoName)
		}
	}
}

func cleanArtifactory() {
	deleteSpecFile := tests.GetFilePathForArtifactory(tests.DeleteSpec)
	log.Output(deleteSpecFile)
	deleteSpecFile, err := tests.ReplaceTemplateVariables(deleteSpecFile, "")
	log.Output(deleteSpecFile)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	deleteSpec, _ := spec.CreateSpecFromFile(deleteSpecFile, nil)
	_, _, err = tests.DeleteFiles(deleteSpec, serverDetails)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func getSpecAndCommonFlags(specFile string) (*spec.SpecFiles, rtutils.CommonConf) {
	searchFlags, _ := rtutils.NewCommonConfImpl(artAuth)
	searchSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	return searchSpec, searchFlags
}

func verifyDoesntExistInArtifactory(specFile string, t *testing.T) {
	inttestutils.VerifyExistInArtifactory([]string{}, specFile, serverDetails, t)
}

func verifyExistInArtifactoryByProps(expected []string, pattern, props string, t *testing.T) {
	searchSpec := spec.NewBuilder().Pattern(pattern).Props(props).Recursive(true).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	readerGetErrorAndAssert(t, readerNoDate)
	tests.CompareExpectedVsActual(expected, resultItems, t)
	readerCloseAndAssert(t, readerNoDate)
}

func isRepoExist(repoName string) bool {
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	resp, _, _, err := client.SendGet(serverDetails.ArtifactoryUrl+tests.RepoDetailsUrl+repoName, true, artHttpDetails, "")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusBadRequest {
		return true
	}
	return false
}

func getCliDotGitPath(t *testing.T) string {
	dotGitPath, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	dotGitExists, err := fileutils.IsDirExists(filepath.Join(dotGitPath, ".git"), false)
	assert.NoError(t, err)
	assert.True(t, dotGitExists, "Can't find .git")
	return dotGitPath
}

func deleteServerConfig(t *testing.T) {
	assert.NoError(t, configCli.WithoutCredentials().Exec("rm", tests.ServerId, "--quiet"))
}

// This function will create server config and return the entire passphrase flag if it needed.
// For example if passphrase is needed it will return "--ssh-passphrase=${theConfiguredPassphrase}" or empty string.
func createServerConfigAndReturnPassphrase(t *testing.T) (passphrase string, err error) {
	deleteServerConfig(t)
	if *tests.JfrogSshPassphrase != "" {
		passphrase = "--ssh-passphrase=" + *tests.JfrogSshPassphrase
	}
	return passphrase, configCli.Exec("add", tests.ServerId)
}

func testCopyMoveNoSpec(command string, beforeCommandExpected, afterCommandExpected []string, t *testing.T) {
	initArtifactoryTest(t, "")

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	runRt(t, "upload", "--spec="+specFileA)
	runRt(t, "upload", "--spec="+specFileB)

	// Run command with dry-run
	runRt(t, command, tests.RtRepo1+"/data/*a*", tests.RtRepo2+"/", "--dry-run")

	// Validate files weren't affected
	cpMvSpecFilePath, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(beforeCommandExpected, cpMvSpecFilePath, serverDetails, t)

	// Run command
	runRt(t, command, tests.RtRepo1+"/data/*a*", tests.RtRepo2+"/")

	// Validate files were affected
	inttestutils.VerifyExistInArtifactory(afterCommandExpected, cpMvSpecFilePath, serverDetails, t)

	// Cleanup
	cleanArtifactoryTest()
}

func searchItemsInArtifactory(t *testing.T, specSource string) []rtutils.ResultItem {
	fileSpec, err := tests.CreateSpec(specSource)
	assert.NoError(t, err)
	spec, flags := getSpecAndCommonFlags(fileSpec)
	var resultItems []rtutils.ResultItem
	for i := 0; i < len(spec.Files); i++ {
		searchParams, err := utils.GetSearchParams(spec.Get(i))
		assert.NoError(t, err, "Failed Searching files")
		reader, err := services.SearchBySpecFiles(searchParams, flags, rtutils.ALL)
		assert.NoError(t, err, "Failed Searching files")
		for resultItem := new(rtutils.ResultItem); reader.NextRecord(resultItem) == nil; resultItem = new(rtutils.ResultItem) {
			resultItems = append(resultItems, *resultItem)
		}
		readerGetErrorAndAssert(t, reader)
		readerCloseAndAssert(t, reader)
	}
	return resultItems
}

func assertDateInSearchResult(searchResult utils.SearchResult) error {
	if searchResult.Created == "" || searchResult.Modified == "" {
		message, err := json.Marshal(&searchResult)
		if err != nil {
			return errors.New("failed to process search result to assert it includes date: " + err.Error())
		}
		return errors.New("search result does not include date: " + string(message))
	}
	return nil
}

func TestArtifactoryUploadInflatedPath(t *testing.T) {
	initArtifactoryTest(t, "")
	runRt(t, "upload", "testdata/a/../a/a1.*", tests.RtRepo1, "--flat=true")
	runRt(t, "upload", "testdata/./a/a1.*", tests.RtRepo1, "--flat=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, serverDetails, t)

	runRt(t, "upload", "testdata/./a/../a/././././a2.*", tests.RtRepo1, "--flat=true")
	searchFilePath, err = tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpected2filesRepo1(), searchFilePath, serverDetails, t)
	if coreutils.IsWindows() {
		runRt(t, "upload", `testdata\\a\\..\\a\\a1.*`, tests.RtRepo2, "--flat=true")
		runRt(t, "upload", `testdata\\.\\\a\a1.*`, tests.RtRepo2, "--flat=true")
		searchFilePath, err = tests.CreateSpec(tests.SearchRepo2)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo2(), searchFilePath, serverDetails, t)

		runRt(t, "upload", `testdata\\.\\a\\..\\a\\.\\.\\.\\.\\a2.*`, tests.RtRepo2, "--flat=true")
		searchFilePath, err = tests.CreateSpec(tests.SearchRepo2)
		assert.NoError(t, err)
		inttestutils.VerifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpected2filesRepo2(), searchFilePath, serverDetails, t)
	}
	cleanArtifactoryTest()
}

func TestGetExtractorsRemoteDetails(t *testing.T) {
	initArtifactoryTest(t, "")
	_, err := createServerConfigAndReturnPassphrase(t)
	assert.NoError(t, err)
	defer deleteServerConfig(t)

	// Make sure extractor1.jar downloaded from releases.jfrog.io.
	downloadPath := "org/jfrog/buildinfo/build-info-extractor/extractor1.jar"
	expectedRemotePath := path.Join("oss-release-local", downloadPath)
	validateExtractorRemoteDetails(t, downloadPath, expectedRemotePath)

	// Make sure extractor2.jar also downloaded from releases.jfrog.io.
	downloadPath = "org/jfrog/buildinfo/build-info-extractor/extractor2.jar"
	expectedRemotePath = path.Join("oss-release-local", downloadPath)
	validateExtractorRemoteDetails(t, downloadPath, expectedRemotePath)

	// Set 'JFROG_CLI_EXTRACTORS_REMOTE' and make sure extractor3.jar downloaded from a remote repo 'test-remote-repo' in ServerId.
	testRemoteRepo := "test-remote-repo"
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, utils.ExtractorsRemoteEnv, tests.ServerId+"/"+testRemoteRepo)
	defer setEnvCallBack()
	downloadPath = "org/jfrog/buildinfo/build-info-extractor/extractor3.jar"
	expectedRemotePath = path.Join(testRemoteRepo, downloadPath)
	validateExtractorRemoteDetails(t, downloadPath, expectedRemotePath)

	cleanArtifactoryTest()
}

func validateExtractorRemoteDetails(t *testing.T, downloadPath, expectedRemotePath string) {
	serverDetails, remotePath, err := utils.GetExtractorsRemoteDetails(downloadPath)
	assert.NoError(t, err)
	assert.Equal(t, expectedRemotePath, remotePath)
	assert.False(t, os.Getenv(utils.ExtractorsRemoteEnv) != "" && serverDetails == nil, "Expected a server to be returned")
}

func TestVcsProps(t *testing.T) {
	initArtifactoryTest(t, "")
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	testDir := initVcsTestDir(t)
	runRt(t, "upload", filepath.Join(testDir, "*"), tests.RtRepo1, "--flat=false", "--build-name="+tests.RtBuildName1, "--build-number=2020")
	resultItems := searchItemsInArtifactory(t, tests.SearchRepo1ByInSuffix)
	assert.NotZero(t, len(resultItems), "No artifacts were found.")
	for _, item := range resultItems {
		properties := item.Properties
		foundUrl, foundRevision := false, false
		for _, prop := range properties {
			if item.Name == "a1.in" || item.Name == "a2.in" {
				// Check that properties were not removed.
				if prop.Key == "vcs.url" && prop.Value == "https://github.com/jfrog/jfrog-cli.git" {
					assert.False(t, foundUrl, "Found duplicate VCS property(url) in artifact")
					foundUrl = true
				}
				if prop.Key == "vcs.revision" && prop.Value == "d63c5957ad6819f4c02a817abe757f210d35ff92" {
					assert.False(t, foundRevision, "Found duplicate VCS property(revision) in artifact")
					foundRevision = true
				}
			}
			if item.Name == "b1.in" || item.Name == "b2.in" {
				if prop.Key == "vcs.url" && prop.Value == "https://github.com/jfrog/jfrog-client-go.git" {
					assert.False(t, foundUrl, "Found duplicate VCS property(url) in artifact")
					foundUrl = true
				}
				if prop.Key == "vcs.revision" && prop.Value == "ad99b6c068283878fde4d49423728f0bdc00544a" {
					assert.False(t, foundRevision, "Found duplicate VCS property(revision) in artifact")
					foundRevision = true
				}
			}
		}
		assert.True(t, foundUrl && foundRevision, "VCS property was not found on artifact: "+item.Name)
	}
	cleanArtifactoryTest()
}

func initVcsTestDir(t *testing.T) string {
	testdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "vcs")
	testdataTarget := tests.Temp
	err := fileutils.CopyDir(testdataSrc, testdataTarget, true, nil)
	assert.NoError(t, err)
	if found, err := fileutils.IsDirExists(filepath.Join(testdataTarget, "gitdata"), false); found {
		assert.NoError(t, err)
		coretests.RenamePath(filepath.Join(testdataTarget, "gitdata"), filepath.Join(testdataTarget, ".git"), t)
	}
	if found, err := fileutils.IsDirExists(filepath.Join(testdataTarget, "OtherGit", "gitdata"), false); found {
		assert.NoError(t, err)
		coretests.RenamePath(filepath.Join(testdataTarget, "OtherGit", "gitdata"), filepath.Join(testdataTarget, "OtherGit", ".git"), t)
	}
	dirPath, err := filepath.Abs(tests.Temp)
	assert.NoError(t, err)
	return dirPath
}

func TestConfigAddOverwrite(t *testing.T) {
	initArtifactoryTest(t, "")
	// Add a new instance.
	err := tests.NewJfrogCli(execMain, "jfrog config", "").Exec("add", tests.ServerId, "--artifactory-url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint, "--user=admin", "--password=password", "--enc-password=false")
	// Remove the instance at the end of the test.
	defer func() {
		assert.NoError(t, tests.NewJfrogCli(execMain, "jfrog config", "").Exec("rm", tests.ServerId, "--quiet"))
	}()
	// Expect no error, because the instance we created has a unique ID.
	assert.NoError(t, err)
	// Try creating an instance with the same ID, and expect to fail, because an instance with the
	// same ID already exists.
	err = tests.NewJfrogCli(execMain, "jfrog config", "").Exec("add", tests.ServerId, "--artifactory-url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint, "--user=admin", "--password=password", "--enc-password=false")
	assert.Error(t, err)
	// Now create it again, this time with the --overwrite option and expect no error.
	err = tests.NewJfrogCli(execMain, "jfrog config", "").Exec("add", tests.ServerId, "--overwrite", "--artifactory-url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint, "--user=admin2", "--password=password", "--enc-password=false")
	assert.NoError(t, err)
}

func TestArtifactoryReplicationCreate(t *testing.T) {
	initArtifactoryTest(t, "")
	// Configure server with dummy credentials
	err := tests.NewJfrogCli(execMain, "jfrog config", "").Exec("add", tests.ServerId, "--artifactory-url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint, "--user=admin", "--password=password", "--enc-password=false")
	defer deleteServerConfig(t)
	assert.NoError(t, err)

	// Init tmp dir
	specFile, err := tests.CreateSpec(tests.ReplicationTempCreate)
	assert.NoError(t, err)

	// Create push replication
	runRt(t, "rplc", specFile)

	// Validate create replication
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	assert.NoError(t, err)
	result, err := servicesManager.GetReplication(tests.RtRepo1)
	assert.NoError(t, err)
	// The Replicator may encrypt the password internally, therefore we should only check that the password is not empty
	assert.NotEmpty(t, result[0].Password)
	result[0].Password = ""
	assert.ElementsMatch(t, result, tests.GetReplicationConfig())

	// Delete replication
	runRt(t, "rpldel", tests.RtRepo1)

	// Validate delete replication
	_, err = servicesManager.GetReplication(tests.RtRepo1)
	assert.Error(t, err)
	// Cleanup
	cleanArtifactoryTest()
}

func TestAccessTokenCreate(t *testing.T) {
	initArtifactoryTest(t, "")

	buffer, _, previousLog := tests.RedirectLogOutputToBuffer()
	// Restore previous logger when the function returns
	defer log.SetLogger(previousLog)

	// Create access token for current user, implicitly
	if *tests.JfrogAccessToken != "" {
		// Use Artifactory CLI with basic auth to allow running `jfrog rt atc` without arguments
		origAccessToken := *tests.JfrogAccessToken
		origUsername, origPassword := tests.SetBasicAuthFromAccessToken(t)
		defer func() {
			*tests.JfrogUser = origUsername
			*tests.JfrogPassword = origPassword
			*tests.JfrogAccessToken = origAccessToken
		}()
		*tests.JfrogAccessToken = ""
		err := tests.NewJfrogCli(execMain, "jfrog rt", authenticate(false)).Exec("atc")
		assert.NoError(t, err)
	} else {
		runRt(t, "atc")
	}

	// Check access token
	checkAccessToken(t, buffer)

	// Create access token for current user, explicitly
	runRt(t, "atc", *tests.JfrogUser)

	// Check access token
	checkAccessToken(t, buffer)

	// Cleanup
	cleanArtifactoryTest()
}

func checkAccessToken(t *testing.T, buffer *bytes.Buffer) {
	// Write the command output to the origin
	output := buffer.Bytes()
	buffer.Reset()

	// Extract the token from the output
	token, err := jsonparser.GetString(output, "access_token")
	assert.NoError(t, err)

	// Try ping with the new token
	err = tests.NewJfrogCli(execMain, "jfrog rt", "--url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint+" --access-token="+token).Exec("ping")
	assert.NoError(t, err)
}

func TestRefreshableArtifactoryTokens(t *testing.T) {
	initArtifactoryTest(t, "")

	if *tests.JfrogAccessToken != "" {
		t.Skip("Test only with username and password , skipping...")
	}

	// Create server with initialized refreshable tokens.
	_, err := createServerConfigAndReturnPassphrase(t)
	defer deleteServerConfig(t)
	assert.NoError(t, err)

	// Upload a file and assert the refreshable tokens were generated.
	artifactoryCommandExecutor := tests.NewJfrogCli(execMain, "jfrog rt", "")
	uploadedFiles := 1
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, tests.ServerId, "testdata/a/a1.in", uploadedFiles)
	if err != nil {
		return
	}
	curAccessToken, curRefreshToken, err := getArtifactoryTokensFromConfig(t, tests.ServerId)
	if err != nil {
		return
	}
	assert.NotEmpty(t, curAccessToken)
	assert.NotEmpty(t, curRefreshToken)

	// Make the token always refresh.
	auth.RefreshBeforeExpiryMinutes = 60

	// Upload a file and assert tokens were refreshed.
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, tests.ServerId, "testdata/a/a2.in", uploadedFiles)
	if err != nil {
		return
	}
	curAccessToken, curRefreshToken, err = assertTokensChanged(t, tests.ServerId, curAccessToken, curRefreshToken)
	if err != nil {
		return
	}

	// Make refresh token invalid. Refreshing using tokens should fail, so new tokens should be generated using credentials.
	err = setArtifactoryRefreshTokenInConfig(t, tests.ServerId, "invalid-token")
	if err != nil {
		return
	}
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, tests.ServerId, "testdata/a/a3.in", uploadedFiles)
	if err != nil {
		return
	}
	curAccessToken, curRefreshToken, err = assertTokensChanged(t, tests.ServerId, curAccessToken, curRefreshToken)
	if err != nil {
		return
	}

	// Make password invalid. Refreshing should succeed, and new token should be obtained.
	err = setPasswordInConfig(t, tests.ServerId, "invalid-pass")
	if err != nil {
		return
	}
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, tests.ServerId, "testdata/a/b/b1.in", uploadedFiles)
	if err != nil {
		return
	}
	curAccessToken, curRefreshToken, err = assertTokensChanged(t, tests.ServerId, curAccessToken, curRefreshToken)
	if err != nil {
		return
	}

	// Make the token not refresh. Verify Tokens did not refresh.
	auth.RefreshBeforeExpiryMinutes = 0
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, tests.ServerId, "testdata/a/b/b2.in", uploadedFiles)
	if err != nil {
		return
	}
	newAccessToken, newRefreshToken, err := getArtifactoryTokensFromConfig(t, tests.ServerId)
	if err != nil {
		return
	}
	assert.Equal(t, curAccessToken, newAccessToken)
	assert.Equal(t, curRefreshToken, newRefreshToken)

	// Cleanup
	cleanArtifactoryTest()
}

func setArtifactoryRefreshTokenInConfig(t *testing.T, serverId, token string) error {
	details, err := config.GetAllServersConfigs()
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	for _, server := range details {
		if server.ServerId == serverId {
			server.SetArtifactoryRefreshToken(token)
		}
	}
	assert.NoError(t, config.SaveServersConf(details))
	return nil
}

func setPasswordInConfig(t *testing.T, serverId, password string) error {
	details, err := config.GetAllServersConfigs()
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	for _, server := range details {
		if server.ServerId == serverId {
			server.SetPassword(password)
		}
	}
	assert.NoError(t, config.SaveServersConf(details))
	return nil
}

func getArtifactoryTokensFromConfig(t *testing.T, serverId string) (accessToken, refreshToken string, err error) {
	details, err := config.GetSpecificConfig(serverId, false, false)
	if err != nil {
		assert.NoError(t, err)
		return "", "", err
	}
	return details.AccessToken, details.ArtifactoryRefreshToken, nil
}

func assertTokensChanged(t *testing.T, serverId, curAccessToken, curRefreshToken string) (newAccessToken, newRefreshToken string, err error) {
	newAccessToken, newRefreshToken, err = getArtifactoryTokensFromConfig(t, serverId)
	if err != nil {
		assert.NoError(t, err)
		return "", "", err
	}
	assert.NotEqual(t, curAccessToken, newAccessToken)
	assert.NotEqual(t, curRefreshToken, newRefreshToken)
	return newAccessToken, newRefreshToken, nil
}

func uploadWithSpecificServerAndVerify(t *testing.T, cli *tests.JfrogCli, serverId string, source string, expectedResults int) error {
	err := cli.Exec("upload", source, tests.RtRepo1, "--server-id="+serverId)
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	assert.Len(t, searchItemsInArtifactory(t, tests.SearchRepo1ByInSuffix), expectedResults)
	return nil
}

func TestArtifactorySimpleUploadAntPattern(t *testing.T) {
	initArtifactoryTest(t, "")

	// --ant and --regexp together: should get an error
	uploadUsingAntAndRegexpTogether(t)
	// Upload empty dir
	uploadUsingAntAIncludeDirsAndFlat(t)
	// Simple uploads
	simpleUploadAntIsTrueRegexpIsFalse(t)
	simpleUploadWithAntPatternSpec(t)

	cleanArtifactoryTest()
}

func uploadUsingAntAndRegexpTogether(t *testing.T) {
	filePath := getAntPatternFilePath()
	err := artifactoryCli.Exec("upload", filePath, tests.RtRepo1, "--regexp", "--ant", "--flat=true")
	assert.Error(t, err)
}

func simpleUploadAntIsTrueRegexpIsFalse(t *testing.T) {
	filePath := getAntPatternFilePath()
	runRt(t, "upload", filePath, tests.RtRepo1, "--ant", "--regexp=false", "--flat=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSimpleAntPatternUploadExpectedRepo1(), searchFilePath, serverDetails, t)
}

func simpleUploadWithAntPatternSpec(t *testing.T) {
	// Init tmp dir
	specFile, err := tests.CreateSpec(tests.UploadAntPattern)
	assert.NoError(t, err)
	err = fileutils.CopyDir(tests.GetTestResourcesPath()+"cache", filepath.Dir(specFile), true, nil)
	assert.NoError(t, err)
	// Upload
	runRt(t, "upload", "--spec="+specFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetSimpleAntPatternUploadExpectedRepo1(), searchFilePath, serverDetails, t)
	searchFilePath, err = tests.CreateSpec(tests.SearchRepo1NonExistFile)
	assert.NoError(t, err)
	verifyDoesntExistInArtifactory(searchFilePath, t)
}

func uploadUsingAntAIncludeDirsAndFlat(t *testing.T) {
	filePath := "testdata/*/empt?/**"
	runRt(t, "upload", filePath, tests.RtRepo1, "--ant", "--include-dirs=true", "--flat=true")
	runRt(t, "upload", filePath, tests.RtRepo1, "--ant", "--include-dirs=true", "--flat=false")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1IncludeDirs)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetAntPatternUploadWithIncludeDirsExpectedRepo1(), searchFilePath, serverDetails, t)
}

func TestUploadWithAntPatternAndExclusionsSpec(t *testing.T) {
	initArtifactoryTest(t, "")
	// Init tmp dir
	specFile, err := tests.CreateSpec(tests.UploadAntPatternExclusions)
	assert.NoError(t, err)
	err = fileutils.CopyDir(tests.GetTestResourcesPath(), filepath.Dir(specFile), true, nil)
	assert.NoError(t, err)
	// Upload
	runRt(t, "upload", "--spec="+specFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetAntPatternUploadWithExclusionsExpectedRepo1(), searchFilePath, serverDetails, t)
	searchFilePath, err = tests.CreateSpec(tests.SearchRepo1NonExistFileAntExclusions)
	assert.NoError(t, err)
	verifyDoesntExistInArtifactory(searchFilePath, t)
	cleanArtifactoryTest()
}

func TestPermissionTargets(t *testing.T) {
	initArtifactoryTest(t, "")
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	templatePath := filepath.Join(tests.GetTestResourcesPath(), "permissiontarget", "template")

	// Create permission target on specific repo.
	runRt(t, "ptc", templatePath, createPermissionTargetsTemplateVars(tests.RtRepo1))
	assertPermissionTarget(t, servicesManager, tests.RtRepo1)

	// Update permission target to ANY repo.
	any := "ANY"
	runRt(t, "ptu", templatePath, createPermissionTargetsTemplateVars(any))
	assertPermissionTarget(t, servicesManager, any)

	// Delete permission target.
	runRt(t, "ptdel", tests.RtPermissionTargetName)
	assertPermissionTargetDeleted(t, servicesManager)

	cleanArtifactoryTest()
}

func createPermissionTargetsTemplateVars(reposValue string) string {
	ptNameVarKey := "pt_name"
	reposVarKey := "repos_var"
	return fmt.Sprintf("--vars=%s=%s;%s=%s", ptNameVarKey, tests.RtPermissionTargetName, reposVarKey, reposValue)
}

func assertPermissionTarget(t *testing.T, manager artifactory.ArtifactoryServicesManager, repoValue string) {
	actual, err := manager.GetPermissionTarget(tests.RtPermissionTargetName)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if actual == nil {
		assert.NotNil(t, actual)
		return
	}
	expected := tests.GetExpectedPermissionTarget(repoValue)
	assert.EqualValues(t, expected, *actual)
}

func assertPermissionTargetDeleted(t *testing.T, manager artifactory.ArtifactoryServicesManager) {
	permission, err := manager.GetPermissionTarget(tests.RtPermissionTargetName)
	assert.NoError(t, err)
	assert.Nil(t, permission)
}

func TestArtifactoryCurl(t *testing.T) {
	initArtifactoryTest(t, "")
	_, err := createServerConfigAndReturnPassphrase(t)
	defer deleteServerConfig(t)
	assert.NoError(t, err)
	// Check curl command with config default server
	err = artifactoryCli.WithoutCredentials().Exec("curl", "-XGET", "/api/system/version")
	assert.NoError(t, err)
	// Check curl command with '--server-id' flag
	err = artifactoryCli.WithoutCredentials().Exec("curl", "-XGET", "/api/system/version", "--server-id="+tests.ServerId)
	assert.NoError(t, err)
	// Check curl command with invalid server id - should get an error.
	err = artifactoryCli.WithoutCredentials().Exec("curl", "-XGET", "/api/system/version", "--server-id=not_configured_name_"+tests.ServerId)
	assert.Error(t, err)

	cleanArtifactoryTest()
}

func deleteProjectIfExists(t *testing.T, accessManager *access.AccessServicesManager, projectKey string) {
	err := accessManager.DeleteProject(projectKey)
	if err != nil {
		if !strings.Contains(err.Error(), "Could not find project") {
			t.Error(t, err)
		}
	}
}

func readerCloseAndAssert(t *testing.T, reader *content.ContentReader) {
	assert.NoError(t, reader.Close(), "Couldn't close reader")
}

func readerGetErrorAndAssert(t *testing.T, reader *content.ContentReader) {
	assert.NoError(t, reader.GetError(), "Couldn't get reader error")
}

func TestProjectInitMaven(t *testing.T) {
	testProjectInit(t, "multiproject", coreutils.Maven)
}

func TestProjectInitGradle(t *testing.T) {
	testProjectInit(t, "gradleproject", coreutils.Gradle)
}

func TestProjectInitNpm(t *testing.T) {
	testProjectInit(t, "npmproject", coreutils.Npm)
}

func TestProjectInitGo(t *testing.T) {
	testProjectInit(t, "dependency", coreutils.Go)
}

func TestProjectInitPip(t *testing.T) {
	testProjectInit(t, "requirementsproject", coreutils.Pip)
}

func TestProjectInitNuget(t *testing.T) {
	testProjectInit(t, "multipackagesconfig", coreutils.Nuget)
}

func testProjectInit(t *testing.T, projectExampleName string, technology coreutils.Technology) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()
	// Create temp JFrog home dir
	tmpHomeDir, deleteHomeDir := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer deleteHomeDir()
	clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, tmpHomeDir)
	_, err := createServerConfigAndReturnPassphrase(t)
	assert.NoError(t, err)

	// Copy a simple project in a temp work dir
	tmpWorkDir, deleteWorkDir := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer deleteWorkDir()
	testdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), technology.ToString(), projectExampleName)
	err = fileutils.CopyDir(testdataSrc, tmpWorkDir, true, nil)
	assert.NoError(t, err)
	if technology == coreutils.Go {
		goModeOriginalPath := filepath.Join(tmpWorkDir, "createGoProject_go.mod_suffix")
		goModeTargetPath := filepath.Join(tmpWorkDir, "go.mod")
		assert.NoError(t, os.Rename(goModeOriginalPath, goModeTargetPath))
	}

	// Run cd command to temp dir.
	currentWd, err := os.Getwd()
	assert.NoError(t, err)
	changeDirBack := clientTestUtils.ChangeDirWithCallback(t, currentWd, tmpWorkDir)
	defer changeDirBack()
	// Run JFrog project init
	err = platformCli.WithoutCredentials().Exec("project", "init", "--path", tmpWorkDir, "--server-id="+tests.ServerId)
	assert.NoError(t, err)
	// Validate correctness of .jfrog/projects/$technology.yml
	validateProjectYamlFile(t, tmpWorkDir, technology.ToString())
	// Validate correctness of .jfrog/projects/build.yml
	validateBuildYamlFile(t, tmpWorkDir)
}

func validateProjectYamlFile(t *testing.T, projectDir, technology string) {
	techConfig, err := utils.ReadConfigFile(filepath.Join(projectDir, ".jfrog", "projects", technology+".yaml"), utils.YAML)
	if assert.NoError(t, err) {
		assert.Equal(t, technology, techConfig.GetString("type"))
		assert.Equal(t, tests.ServerId, techConfig.GetString("resolver.serverId"))
		assert.Equal(t, tests.ServerId, techConfig.GetString("deployer.serverId"))
	}
}

func validateBuildYamlFile(t *testing.T, projectDir string) {
	techConfig, err := utils.ReadConfigFile(filepath.Join(projectDir, ".jfrog", "projects", "build.yaml"), utils.YAML)
	assert.NoError(t, err)
	assert.Equal(t, "build", techConfig.GetString("type"))
	assert.Equal(t, filepath.Base(projectDir+"/"), techConfig.GetString("name"))
}

func TestTerraformPublish(t *testing.T) {
	initArtifactoryTest(t, terraformMinArtifactoryVersion)
	defer cleanArtifactoryTest()
	createJfrogHomeConfig(t, true)
	projectPath := prepareTerraformProject("terraformproject", t, true)
	// Change working directory to be the project's local root.
	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, filepath.Join(projectPath, "aws"))
	defer chdirCallback()
	artifactoryCli.SetPrefix("jf")

	// Terraform publish
	err = artifactoryCli.Exec("terraform", "publish", "--namespace=namespace", "--provider=provider", "--tag=tag", "--exclusions=*test*")
	assert.NoError(t, err)
	artifactoryCli.SetPrefix("jf rt")

	// Download modules to 'result' directory.
	chdirCallback()
	assert.NoError(t, os.MkdirAll(tests.Out+"/results/", 0777))
	// Verify terraform modules have been uploaded to artifactory correctly.
	verifyModuleInArtifactoryWithRetry(t)
}

func prepareTerraformProject(projectName string, t *testing.T, copyDirs bool) string {
	projectPath := filepath.Join(tests.GetTestResourcesPath(), "terraform", projectName)
	testdataTarget := filepath.Join(tests.Out, "terraformProject")
	assert.NoError(t, os.MkdirAll(testdataTarget+string(os.PathSeparator), 0777))
	// Copy terraform tests to test environment, so we can change project's config file.
	assert.NoError(t, fileutils.CopyDir(projectPath, testdataTarget, copyDirs, nil))
	configFileDir := filepath.Join(filepath.FromSlash(testdataTarget), ".jfrog", "projects")
	_, err := tests.ReplaceTemplateVariables(filepath.Join(configFileDir, "terraform.yaml"), configFileDir)
	assert.NoError(t, err)
	return testdataTarget
}

func verifyModuleInArtifactoryWithRetry(t *testing.T) {
	retryExecutor := &clientutils.RetryExecutor{
		MaxRetries: 5,
		// RetriesIntervalMilliSecs in milliseconds
		RetriesIntervalMilliSecs: 1000,
		ErrorMessage:             "Waiting for Artifactory to create \"module.json\" files for terraform modules....",
		ExecutionHandler:         downloadModuleAndVerify(),
	}
	err := retryExecutor.Execute()
	assert.NoError(t, err)
}

func downloadModuleAndVerify() clientutils.ExecutionHandlerFunc {
	return func() (shouldRetry bool, err error) {
		err = artifactoryCli.Exec("download", tests.TerraformRepo+"/namespace/*", tests.Out+"/results/", "--explode=true")
		if err != nil {
			return false, err
		}
		// Validate
		paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(filepath.Join(tests.Out, "results"), false)
		if err != nil {
			return false, err
		}
		// After uploading terraform module to Artifactory the indexing is async.
		// It could take some time for "module.json" files to be created by artifactory - that's why we should try downloading again in case comparison has failed.
		err = tests.ValidateListsIdentical(tests.GetTerraformModulesFilesDownload(), paths)
		if err != nil {
			return true, err
		}
		return false, nil
	}
}
