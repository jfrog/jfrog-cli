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

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-client-go/artifactory"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"

	"github.com/buger/jsonparser"
	gofrogio "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	cliproxy "github.com/jfrog/jfrog-cli/utils/tests/proxy/server"
	"github.com/jfrog/jfrog-cli/utils/tests/proxy/server/certificate"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	rtutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils/tests/xray"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JFrog CLI for Artifactory commands
var artifactoryCli *tests.JfrogCli

// JFrog CLI for config command only (doesn't pass the --ssh-passphrase flag)
var configCli *tests.JfrogCli

var serverDetails *config.ServerDetails
var artAuth auth.ServiceDetails
var artHttpDetails httputils.HttpClientDetails

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
	serverDetails = &config.ServerDetails{ArtifactoryUrl: *tests.JfrogUrl + tests.ArtifactoryEndpoint, SshKeyPath: *tests.JfrogSshKeyPath, SshPassphrase: *tests.JfrogSshPassphrase}
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
	initArtifactoryTest(t)
	specFile, err := tests.CreateSpec(tests.UploadFlatRecursive)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleUploadExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactorySimpleUploadWithWildcardSpec(t *testing.T) {
	initArtifactoryTest(t)
	// Init tmp dir
	specFile, err := tests.CreateSpec(tests.UploadTempWildcard)
	assert.NoError(t, err)
	err = fileutils.CopyDir(tests.GetTestResourcesPath()+"cache", filepath.Dir(specFile), true, nil)
	assert.NoError(t, err)
	// Upload
	artifactoryCli.Exec("upload", "--spec="+specFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleWildcardUploadExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

// This test is similar to TestArtifactorySimpleUploadSpec but using "--server-id" flag
func TestArtifactorySimpleUploadSpecUsingConfig(t *testing.T) {
	initArtifactoryTest(t)
	passphrase, err := createServerConfigAndReturnPassphrase()
	defer deleteServerConfig()
	assert.NoError(t, err)
	artifactoryCommandExecutor := tests.NewJfrogCli(execMain, "jfrog rt", "")
	specFile, err := tests.CreateSpec(tests.UploadFlatRecursive)
	assert.NoError(t, err)
	artifactoryCommandExecutor.Exec("upload", "--spec="+specFile, "--server-id="+tests.ServerId, passphrase)

	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleUploadExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadPathWithSpecialCharsAsNoRegex(t *testing.T) {
	initArtifactoryTest(t)
	filePath := getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1, "--flat")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryEmptyBuild(t *testing.T) {
	initArtifactoryTest(t)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	buildNumber := "5"

	// Try to upload with non existent pattern
	err := artifactoryCli.Exec("upload", "*.notExist", tests.RtRepo1, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)
	assert.NoError(t, err)

	// Try to download with non existent pattern
	err = artifactoryCli.Exec("download", tests.RtRepo1+"/*.notExist", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)
	assert.NoError(t, err)

	// Publish empty build info
	err = artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumber)
	assert.NoError(t, err)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryPublishBuildUsingBuildlFile(t *testing.T) {
	initArtifactoryTest(t)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	// Create temp folder.
	tmpDir, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	defer func() { assert.NoError(t, fileutils.RemoveTempDir(tmpDir)) }()
	// Create build config in temp folder
	_, err = tests.ReplaceTemplateVariables(filepath.Join("testdata", "buildspecs", "build.yaml"), filepath.Join(tmpDir, ".jfrog", "projects"))
	assert.NoError(t, err)

	// Run cd command to temp dir.
	wdCopy, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := tests.ChangeDirWithCallback(t, tmpDir)

	// Upload file to create build-info data using the build.yaml file.
	assert.NoError(t, artifactoryCli.Exec("upload", filepath.Join(wdCopy, "testdata", "a", "a1.in"), tests.RtRepo1+"/foo"))

	// Publish build-info using the build.yaml file.
	err = artifactoryCli.Exec("build-publish")
	assert.NoError(t, err)

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
	assert.NoError(t, artifactoryCli.Exec("upload", filepath.Join(wdCopy, "testdata", "a", "a1.in"), tests.RtRepo1+"/bla-bla"))

	// Publish the second build-info build.yaml file.
	err = artifactoryCli.Exec("build-publish")
	assert.NoError(t, err)

	// Search artifacts based on the second published build.
	searchSpecBuilder = spec.NewBuilder().Pattern(tests.RtRepo1).Build(tests.RtBuildName1 + "/2")
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())

	// Validate the search result.
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	searchResultLength, err = reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, 1, searchResultLength)

	chdirCallback()
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadFromVirtual(t *testing.T) {
	initArtifactoryTest(t)

	artifactoryCli.Exec("upload", "testdata/a/*", tests.RtRepo1, "--flat=false")
	artifactoryCli.Exec("dl", tests.RtVirtualRepo+"/testdata/(*)", tests.Out+"/"+"{1}", "--flat=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally(tests.GetVirtualDownloadExpected(), paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadPathWithSpecialChars(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", getSpecialCharFilePath(), tests.RtRepo1, "--flat=false")
	artifactoryCli.Exec("upload", "testdata/c#/a#1.in", tests.RtRepo1, "--flat=false")

	artifactoryCli.Exec("dl", tests.RtRepo1+"/testdata/a$+~&^a#/a*", tests.Out+fileutils.GetFileSeparator(), "--flat=true")
	artifactoryCli.Exec("dl", tests.RtRepo1+"/testdata/c#/a#1.in", tests.Out+fileutils.GetFileSeparator(), "--flat=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a1.in"), filepath.Join(tests.Out, "a#1.in")}, paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadPatternWithUnicodeChars(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "testdata/unicode/", tests.RtRepo1, "--flat=false")

	// Verify files exist
	specFile, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetDownloadUnicode(), specFile, t)

	artifactoryCli.Exec("dl", tests.RtRepo1+"/testdata/unicode/(dirλrectory)/", filepath.Join(tests.Out, "{1}")+fileutils.GetFileSeparator(), "--flat=true")

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
	initArtifactoryTest(t)
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)
	randFile, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "randFile"), fileSize)
	assert.NoError(t, err)
	localFileDetails, err := fileutils.GetFileDetails(randFile.Name(), true)
	assert.NoError(t, err)

	artifactoryCli.Exec("u", randFile.Name(), tests.RtRepo1+"/testdata/", "--flat=true")
	randFile.File.Close()
	os.RemoveAll(tests.Out)
	artifactoryCli.Exec("dl", tests.RtRepo1+"/testdata/", tests.Out+fileutils.GetFileSeparator(), "--flat=true")

	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "randFile")}, paths, t)
	tests.ValidateChecksums(filepath.Join(tests.Out, "randFile"), localFileDetails.Checksum, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadWildcardInRepo(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	// Upload a file to repo1 and another one to repo2
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/a1.in")
	artifactoryCli.Exec("upload", filePath, tests.RtRepo2+"/path/a2.in")

	specFile, err := tests.CreateSpec(tests.DownloadWildcardRepo)
	assert.NoError(t, err)

	// Verify the 2 files exist using `*` in the repository name
	verifyExistInArtifactory(tests.GetDownloadWildcardRepo(), specFile, t)

	// Download the 2 files with `*` in the repository name
	artifactoryCli.Exec("dl", "--spec="+specFile)
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a1.in"), filepath.Join(tests.Out, "a2.in")}, paths, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadPlaceholderInRepo(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	// Upload a file to repo1 and another one to repo2
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/a1.in")
	artifactoryCli.Exec("upload", filePath, tests.RtRepo2+"/path/a2.in")

	specFile, err := tests.CreateSpec(tests.DownloadWildcardRepo)
	assert.NoError(t, err)

	// Verify the 2 files exist
	verifyExistInArtifactory(tests.GetDownloadWildcardRepo(), specFile, t)

	// Download the 2 files with placeholders in the repository name
	artifactoryCli.Exec("dl", tests.RtRepo1And2Placeholder, tests.Out+"/a/{1}/", "--flat=true")
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a", "1", "a1.in"), filepath.Join(tests.Out, "a", "2", "a2.in")}, paths, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadPlaceholderFlat(t *testing.T) {
	initArtifactoryTest(t)
	for _, flatValue := range []string{"true", "false"} {
		artifactoryCli.Exec("upload", "testdata/a/b/(*)", tests.RtRepo1+"/path/{1}", "--flat="+flatValue)
		searchPath, err := tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolder(), searchPath, t)
		artifactoryCli.Exec("del", tests.RtRepo1+"/*")

		artifactoryCli.Exec("upload", "testdata/a/b/(*)", tests.RtRepo1+"/path/{1}/", "--flat="+flatValue)
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolderlashSlashSuffix(), searchPath, t)
		artifactoryCli.Exec("del", tests.RtRepo1+"/*")

		artifactoryCli.Exec("upload", "testdata/a/b/(*)/(*)", tests.RtRepo1+"/path/{1}/{2}", "--flat="+flatValue)
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetUploadedFileWithDownloadedDoublePlaceHolder(), searchPath, t)
		artifactoryCli.Exec("del", tests.RtRepo1+"/*")
	}
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadWithPlaceholderFlat(t *testing.T) {
	initArtifactoryTest(t)
	// Upload test data to Artifactory
	artifactoryCli.Exec("upload", "testdata/a/b/(*)", tests.RtRepo1+"/path/{1}")
	// Download the tests data using place holder with flate
	for _, flatValue := range []string{"true", "false"} {
		assert.NoError(t, artifactoryCli.Exec("download", tests.RtRepo1+"/path/(*)", tests.Out+"/mypath2/{1}", "--flat="+flatValue))
		paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
		assert.NoError(t, err)
		checkSyncedDirContent(tests.GetFileWithDownloadedPlaceHolder(), paths, t)
		os.RemoveAll(tests.Out)

		assert.NoError(t, artifactoryCli.Exec("download", tests.RtRepo1+"/path/(*)", tests.Out+"/mypath2/{1}/", "--flat="+flatValue))
		paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
		assert.NoError(t, err)
		checkSyncedDirContent(tests.GetFileWithDownloadedPlaceHolderSlashSuffix(), paths, t)
		os.RemoveAll(tests.Out)

		assert.NoError(t, artifactoryCli.Exec("download", tests.RtRepo1+"/path/(*)/(*)", tests.Out+"/mypath2/{1}/{2}", "--flat="+flatValue))
		paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
		assert.NoError(t, err)
		checkSyncedDirContent(tests.GetFileWithDownloadedDoublePlaceHolder(), paths, t)
		os.RemoveAll(tests.Out)
	}
	cleanArtifactoryTest()
}

func TestArtifactoryCopyWithPlaceholderFlat(t *testing.T) {
	initArtifactoryTest(t)
	// Upload test data to Artifactory
	artifactoryCli.Exec("upload", "testdata/a/b/(*)", tests.RtRepo2+"/mypath2/{1}")
	// Download the tests data using place holder with flate
	for _, flatValue := range []string{"true", "false"} {
		assert.NoError(t, artifactoryCli.Exec("copy", tests.RtRepo2+"/mypath2/(*)", tests.RtRepo1+"/path/{1}", "--flat="+flatValue))
		searchPath, err := tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolder(), searchPath, t)
		artifactoryCli.Exec("del", tests.RtRepo1+"/*")

		assert.NoError(t, artifactoryCli.Exec("copy", tests.RtRepo2+"/mypath2/(*)", tests.RtRepo1+"/path/{1}/", "--flat="+flatValue))
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolderlashSlashSuffix(), searchPath, t)
		artifactoryCli.Exec("del", tests.RtRepo1+"/*")

		assert.NoError(t, artifactoryCli.Exec("copy", tests.RtRepo2+"/mypath2/(*)/(*)", tests.RtRepo1+"/path/{1}/{2}", "--flat="+flatValue))
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetUploadedFileWithDownloadedDoublePlaceHolder(), searchPath, t)
		artifactoryCli.Exec("del", tests.RtRepo1+"/*")
	}
	cleanArtifactoryTest()
}

func TestArtifactoryMoveWithPlaceholderFlat(t *testing.T) {
	initArtifactoryTest(t)
	// Download the tests data using place holder with flate
	for _, flatValue := range []string{"true", "false"} {
		artifactoryCli.Exec("upload", "testdata/a/b/(*)", tests.RtRepo2+"/mypath2/{1}")
		assert.NoError(t, artifactoryCli.Exec("move", tests.RtRepo2+"/mypath2/(*)", tests.RtRepo1+"/path/{1}", "--flat="+flatValue))
		searchPath, err := tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolder(), searchPath, t)
		artifactoryCli.Exec("del", tests.RtRepo1+"/*")
		artifactoryCli.Exec("del", tests.RtRepo2+"/*")

		artifactoryCli.Exec("upload", "testdata/a/b/(*)", tests.RtRepo2+"/mypath2/{1}")
		assert.NoError(t, artifactoryCli.Exec("move", tests.RtRepo2+"/mypath2/(*)", tests.RtRepo1+"/path/{1}/", "--flat="+flatValue))
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetUploadedFileWithDownloadedPlaceHolderlashSlashSuffix(), searchPath, t)
		artifactoryCli.Exec("del", tests.RtRepo1+"/*")
		artifactoryCli.Exec("del", tests.RtRepo2+"/*")

		artifactoryCli.Exec("upload", "testdata/a/b/(*)", tests.RtRepo2+"/mypath2/{1}")
		assert.NoError(t, artifactoryCli.Exec("move", tests.RtRepo2+"/mypath2/(*)/(*)", tests.RtRepo1+"/path/{1}/{2}", "--flat="+flatValue))
		searchPath, err = tests.CreateSpec(tests.SearchAllRepo1)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetUploadedFileWithDownloadedDoublePlaceHolder(), searchPath, t)
		artifactoryCli.Exec("del", tests.RtRepo1+"/*")
		artifactoryCli.Exec("del", tests.RtRepo2+"/*")
	}
	cleanArtifactoryTest()
}

func TestArtifactoryCopySingleFileNonFlat(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/", "--flat")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path/a1.in", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyPrefixFilesFlat(t *testing.T) {
	initArtifactoryTest(t)

	artifactoryCli.Exec("upload", "testdata/prefix/(*)", tests.RtRepo1+"/prefix/prefix-{1}")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/prefix/*", tests.RtRepo2, "--flat")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetPrefixFilesCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestAqlFindingItemOnRoot(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/inner/", "--flat=true")
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/*", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetAnyItemCopy(), searchPath, t)
	artifactoryCli.Exec("del", tests.RtRepo2+"/*")
	artifactoryCli.Exec("del", tests.RtRepo1+"/*")
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/", "--flat=true")
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/*/", tests.RtRepo2)
	verifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestExitCode(t *testing.T) {
	initArtifactoryTest(t)

	// Upload dummy file in order to test move and copy commands
	artifactoryCli.Exec("upload", path.Join("testdata", "a", "a1.in"), tests.RtRepo1)

	// Discard output logging to prevent negative logs
	previousLogger := tests.RedirectLogOutputToNil()
	defer log.SetLogger(previousLogger)

	// Test upload commands
	err := artifactoryCli.Exec("upload", "DummyText", tests.RtRepo1, "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeError, err)
	err = artifactoryCli.Exec("upload", path.Join("testdata", "a", "a1.in"), "DummyTargetPath")
	checkExitCode(t, coreutils.ExitCodeError, err)
	err = artifactoryCli.Exec("upload", "testdata/a/(*.dummyExt)", tests.RtRepo1+"/{1}.in", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)

	// Test download command
	err = artifactoryCli.Exec("dl", "DummyFolder", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)

	// Test move commands
	err = artifactoryCli.Exec("move", tests.RtRepo1, "DummyTargetPath")
	checkExitCode(t, coreutils.ExitCodeError, err)
	err = artifactoryCli.Exec("move", "DummyText", tests.RtRepo1, "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)

	// Test copy commands
	err = artifactoryCli.Exec("copy", tests.RtRepo1, "DummyTargetPath")
	checkExitCode(t, coreutils.ExitCodeError, err)
	err = artifactoryCli.Exec("copy", "DummyText", tests.RtRepo1, "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)

	// Test delete command
	err = artifactoryCli.Exec("delete", "DummyText", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)

	// Test search command
	err = artifactoryCli.Exec("s", "DummyText", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)

	// Test props commands
	err = artifactoryCli.Exec("sp", "DummyText", "prop=val;key=value", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)
	err = artifactoryCli.Exec("delp", "DummyText", "prop=val;key=value", "--fail-no-op=true")
	checkExitCode(t, coreutils.ExitCodeFailNoOp, err)

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
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path/", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyUsingWildcard(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/*/", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyFilesNameWithParentheses(t *testing.T) {
	initArtifactoryTest(t)

	artifactoryCli.Exec("upload", "testdata/b/*", tests.RtRepo1, "--flat=false")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/(/(.in", tests.RtRepo2)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/(b/(b.in", tests.RtRepo2)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/b(/b(.in", tests.RtRepo2)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/b)/b).in", tests.RtRepo2)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/(b)/(b).in", tests.RtRepo2)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/)b/)b.in", tests.RtRepo2)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/)b)/)b).in", tests.RtRepo2)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/(b/(b.in", tests.RtRepo2+"/()/", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/(b)/(b).in", tests.RtRepo2+"/()/")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/b(/b(.in", tests.RtRepo2+"/(/", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/(/(*.in)", tests.RtRepo2+"/c/{1}.zip", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/(/(*.in)", tests.RtRepo2+"/(/{1}.zip")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/b(/(b*.in)", tests.RtRepo2+"/(/{1}-up", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/testdata/b/b(/(*).(*)", tests.RtRepo2+"/(/{2}-{1}", "--flat=true")

	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetCopyFileNameWithParentheses(), searchPath, t)

	cleanArtifactoryTest()
}

func TestArtifactoryCreateUsers(t *testing.T) {
	initArtifactoryTest(t)
	usersCSVPath := "testdata/usersmanagement/users.csv"
	randomUsersCSVPath, err := tests.ReplaceTemplateVariables(usersCSVPath, "")
	assert.NoError(t, err)
	err = artifactoryCli.Exec("users-create", "--csv="+randomUsersCSVPath)
	// Clean up
	defer func() {
		err = artifactoryCli.Exec("users-delete", "--csv="+randomUsersCSVPath)
		assert.NoError(t, err)
		cleanArtifactoryTest()
	}()
	assert.NoError(t, err)

	verifyUsersExistInArtifactory(randomUsersCSVPath, t)
}

func verifyUsersExistInArtifactory(csvFilePath string, t *testing.T) {
	// Parse input CSV
	content, err := os.Open(csvFilePath)
	assert.NoError(t, err)
	csvReader := csv.NewReader(content)
	// Ignore the header
	csvReader.Read()
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
	initArtifactoryTest(t)

	specFile, err := tests.CreateSpec(tests.UploadFileWithParenthesesSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	searchPath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadFileNameWithParentheses(), searchPath, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadFilesNameWithParenthesis(t *testing.T) {
	initArtifactoryTest(t)

	artifactoryCli.Exec("upload", "testdata/b/*", tests.RtRepo1, "--flat=false")
	artifactoryCli.Exec("download", path.Join(tests.RtRepo1), tests.Out+"/")

	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(tests.GetFileWithParenthesesDownload(), paths, t)

	cleanArtifactoryTest()
}
func TestArtifactoryDownloadDotAsTarget(t *testing.T) {
	initArtifactoryTest(t)
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))
	randFile, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "DownloadDotAsTarget"), 100000)
	randFile.File.Close()
	assert.NoError(t, artifactoryCli.Exec("upload", tests.Out+"/*", tests.RtRepo1+"/p-modules/", "--flat=true"))
	assert.NoError(t, os.RemoveAll(tests.Out))
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))

	chdirCallback := tests.ChangeDirWithCallback(t, tests.Out)
	defer chdirCallback()

	assert.NoError(t, artifactoryCli.Exec("download", tests.RtRepo1+"/p-modules/DownloadDotAsTarget", "."))
	chdirCallback()

	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally([]string{tests.Out, filepath.Join(tests.Out, "p-modules"), filepath.Join(tests.Out, "p-modules", "DownloadDotAsTarget")}, paths, t)
	tests.RemoveTempDirAndAssert(t, tests.Out)
	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyUsingWildcardFlat(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/inner/", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path/inner", tests.RtRepo2, "--flat=true")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetSingleDirectoryCopyFlat(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyPathsTwice(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/inner/", "--flat")

	log.Info("Copy Folder to root twice")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetSingleFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path", tests.RtRepo2)
	verifyExistInArtifactory(tests.GetSingleFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("del", tests.RtRepo2)

	log.Info("Copy to from repo1/path to repo2/path twice")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path", tests.RtRepo2+"/path")
	verifyExistInArtifactory(tests.GetSingleFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path", tests.RtRepo2+"/path")
	verifyExistInArtifactory(tests.GetFolderCopyTwice(), searchPath, t)
	artifactoryCli.Exec("del", tests.RtRepo2)

	log.Info("Copy to from repo1/path/ to repo2/path/ twice")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path/", tests.RtRepo2+"/path/")
	verifyExistInArtifactory(tests.GetSingleInnerFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path/", tests.RtRepo2+"/path/")
	verifyExistInArtifactory(tests.GetSingleInnerFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("del", tests.RtRepo2)

	log.Info("Copy to from repo1/path/ to repo2/path/ twice")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path", tests.RtRepo2+"/path/")
	verifyExistInArtifactory(tests.GetFolderCopyIntoFolder(), searchPath, t)
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path", tests.RtRepo2+"/path/")
	verifyExistInArtifactory(tests.GetFolderCopyIntoFolder(), searchPath, t)
	artifactoryCli.Exec("del", tests.RtRepo2)

	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyPatternEndsWithSlash(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/inner/")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path/", tests.RtRepo2, "--flat=true")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetAnyItemCopyUsingSpec(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemUsingWildcardFlat(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/inner/", "--flat")
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/*", tests.RtRepo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetAnyItemCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemRecursive(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/a/b/", "--flat")
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/aFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/a*", tests.RtRepo2, "--recursive=true")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetAnyItemCopyRecursive(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAndRenameFolder(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/inner/", "--flat")
	artifactoryCli.Exec("cp", tests.RtRepo1+"/path/(*)", tests.RtRepo2+"/newPath/{1}")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetCopyFolderRename(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	specFile, err := tests.CreateSpec(tests.CopyItemsSpec)
	assert.NoError(t, err)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/path/inner/")
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", "--spec="+specFile)
	verifyExistInArtifactory(tests.GetAnyItemCopyUsingSpec(), searchPath, t)
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
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Copy by pattern
	artifactoryCli.Exec("cp", tests.RtRepo1+"/data/", tests.RtRepo2+"/", "--exclusions=*/*b*;*/*c*")

	// Validate files are moved by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildCopyExclude(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadDebian(t *testing.T) {
	initArtifactoryTest(t)
	specFile, err := tests.CreateSpec(tests.DebianUploadSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile, "--deb=bionic/main/i386")
	verifyExistInArtifactoryByProps(tests.GetUploadDebianExpected(), tests.RtDebianRepo+"/*", "deb.distribution=bionic;deb.component=main;deb.architecture=i386", t)
	artifactoryCli.Exec("upload", "--spec="+specFile, "--deb=cosmic/main\\/18.10/amd64")
	verifyExistInArtifactoryByProps(tests.GetUploadDebianExpected(), tests.RtDebianRepo+"/*", "deb.distribution=cosmic;deb.component=main/18.10;deb.architecture=amd64", t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadAndExplode(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", filepath.Join("testdata", "archives", "a.zip"), tests.RtRepo1, "--explode=true", "--flat")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetExplodeUploadExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadAndSyncDelete(t *testing.T) {
	initArtifactoryTest(t)
	// Upload all testdata/a/
	artifactoryCli.Exec("upload", path.Join("testdata", "a", "*"), tests.RtRepo1+"/syncDir/", "--flat=false")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep1(), searchFilePath, t)
	// Upload testdata/a/b/*1.in and sync syncDir/testdata/a/b/
	artifactoryCli.Exec("upload", path.Join("testdata", "a", "b", "*1.in"), tests.RtRepo1+"/syncDir/", "--sync-deletes="+tests.RtRepo1+"/syncDir/testdata/a/b/", "--flat=false")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep2(), searchFilePath, t)
	// Upload testdata/archives/* and sync syncDir/
	artifactoryCli.Exec("upload", path.Join("testdata", "archives", "*"), tests.RtRepo1+"/syncDir/", "--sync-deletes="+tests.RtRepo1+"/syncDir/", "--flat=true")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep3(), searchFilePath, t)
	// Upload testdata/b/ and sync syncDir/testdata/b/b
	// Noticed that testdata/c/ includes sub folders with special chars like '-' and '#'
	artifactoryCli.Exec("upload", path.Join("testdata", "c", "*"), tests.RtRepo1+"/syncDir/", "--sync-deletes="+tests.RtRepo1+"/syncDir/", "--flat=false")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep4(), searchFilePath, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplode(t *testing.T) {
	initArtifactoryTest(t)
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	randFile, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "randFile"), 100000)
	assert.NoError(t, err)

	err = archiver.TarGz.Make(filepath.Join(tests.Out, "concurrent.tar.gz"), []string{randFile.Name()})
	assert.NoError(t, err)
	err = archiver.Tar.Make(filepath.Join(tests.Out, "bulk.tar"), []string{randFile.Name()})
	assert.NoError(t, err)
	err = archiver.Zip.Make(filepath.Join(tests.Out, "zipFile.zip"), []string{randFile.Name()})
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/*", tests.RtRepo1, "--flat=true")
	randFile.File.Close()
	os.RemoveAll(tests.Out)
	// Download 'concurrent.tar.gz' as 'concurrent' file name and explode it.
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/concurrent", "--explode=true"))
	// Download 'concurrent.tar.gz' and explode it.
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/", "--explode=true"))
	// Download 'concurrent.tar.gz' without explode it.
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/", "--explode=false"))
	// Try to explode the archive that already been downloaded.
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/", "--explode=true"))
	os.RemoveAll(tests.Out)
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "randFile"), tests.Out+"/", "--explode=true"))
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "concurrent.tar.gz"), tests.Out+"/", "--explode=false"))
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "bulk.tar"), tests.Out+"/", "--explode=true"))
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "zipFile.zip"), tests.Out+"/", "--explode=true"))
	verifyExistAndCleanDir(t, tests.GetExtractedDownload)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplodeCurDirAsTarget(t *testing.T) {
	initArtifactoryTest(t)
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	randFile, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "DownloadAndExplodeCurDirTarget"), 100000)
	assert.NoError(t, err)

	err = archiver.TarGz.Make(filepath.Join(tests.Out, "curDir.tar.gz"), []string{randFile.Name()})
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", tests.Out+"/*", tests.RtRepo1, "--flat=true")
	assert.NoError(t, artifactoryCli.Exec("upload", tests.Out+"/*", tests.RtRepo1+"/p-modules/", "--flat=true"))
	randFile.File.Close()
	tests.RemoveTempDirAndAssert(t, tests.Out)

	// Change working dir to tests temp "out" dir
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))
	chdirCallback := tests.ChangeDirWithCallback(t, tests.Out)
	defer chdirCallback()

	// Dot as target
	assert.NoError(t, artifactoryCli.Exec("download", tests.RtRepo1+"/p-modules/curDir.tar.gz", ".", "--explode=true"))
	// Changing current working dir to "out" dir
	chdirCallback()
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadCurDir)
	assert.NoError(t, os.Chdir(tests.Out))

	// No target
	assert.NoError(t, artifactoryCli.Exec("download", tests.RtRepo1+"/p-modules/curDir.tar.gz", "--explode=true"))
	// Changing working dir for testing
	chdirCallback()
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadCurDir)
	assert.NoError(t, os.Chdir(tests.Out))

	chdirCallback()
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplodeFlat(t *testing.T) {
	initArtifactoryTest(t)
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	file1, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "file1"), 100000)
	assert.NoError(t, err)

	err = archiver.Tar.Make(filepath.Join(tests.Out, "flat.tar"), []string{file1.Name()})
	assert.NoError(t, err)
	err = archiver.Zip.Make(filepath.Join(tests.Out, "tarZipFile.zip"), []string{tests.Out + "/flat.tar"})
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", tests.Out+"/*", tests.RtRepo1+"/checkFlat/dir/", "--flat=true")
	file1.File.Close()
	os.RemoveAll(tests.Out)
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))

	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "checkFlat", "dir", "flat.tar"), tests.Out+"/checkFlat/", "--explode=true", "--flat=true", "--min-split=50"))
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "checkFlat", "dir", "tarZipFile.zip"), tests.Out+"/", "--explode=true", "--flat=false"))
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadFlatFalse)
	// Explode 'flat.tar' while the file exists in the file system using --flat
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "checkFlat", "dir", "tarZipFile.zip"), tests.Out+"/", "--explode=true", "--flat=false"))
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "checkFlat", "dir", "flat.tar"), tests.Out+"/checkFlat/dir/", "--explode=true", "--flat", "--min-split=50"))
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadTarFileFlatFalse)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplodeConcurrent(t *testing.T) {
	initArtifactoryTest(t)
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", path.Join("testdata", "archives", "a.zip"), tests.RtRepo1, "--flat=true")
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "a.zip"), tests.Out+"/", "--explode=true", "--min-split=50"))
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadConcurrent)
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "a.zip"), tests.Out+"/", "--explode=false", "--min-split=50"))
	verifyExistAndCleanDir(t, tests.GetArchiveConcurrent)
	assert.NoError(t, artifactoryCli.Exec("download", path.Join(tests.RtRepo1, "a.zip"), tests.Out+"/", "--explode=true", "--split-count=15", "--min-split=50"))
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadConcurrent)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplodeSpecialChars(t *testing.T) {
	initArtifactoryTest(t)
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)
	file1, err := gofrogio.CreateRandFile(filepath.Join(tests.Out, "file $+~&^a#1"), 1000)
	assert.NoError(t, err)
	err = archiver.Tar.Make(filepath.Join(tests.Out, "a$+~&^a#.tar"), []string{file1.Name()})
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", tests.Out+"/*", tests.RtRepo1+"/dir/", "--flat=true")
	os.RemoveAll(tests.Out)
	err = fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.RtRepo1+"/dir/a$+~&^a#.tar", tests.Out+"/dir $+~&^a# test/", "--flat=true", "--explode")
	artifactoryCli.Exec("dl", tests.RtRepo1+"/dir/a$+~&^a#.tar", tests.Out+"/dir $+~&^a# test/", "--flat=false", "--explode")
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadTarFileSpecialChars)
	// Concurrently download
	artifactoryCli.Exec("dl", tests.RtRepo1+"/dir/a$+~&^a#.tar", tests.Out+"/dir $+~&^a# test/", "--flat=true", "--explode", "--min-split=50")
	artifactoryCli.Exec("dl", tests.RtRepo1+"/dir/a$+~&^a#.tar", tests.Out+"/dir $+~&^a# test/", "--flat=false", "--explode", "--min-split=50")
	verifyExistAndCleanDir(t, tests.GetExtractedDownloadTarFileSpecialChars)
	cleanArtifactoryTest()
}

func verifyExistAndCleanDir(t *testing.T, GetExtractedDownload func() []string) {
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(GetExtractedDownload(), paths, t)
	os.RemoveAll(tests.Out)
	assert.NoError(t, fileutils.CreateDirIfNotExist(tests.Out))
}

func TestArtifactoryUploadAsArchive(t *testing.T) {
	initArtifactoryTest(t)

	uploadSpecFile, err := tests.CreateSpec(tests.UploadAsArchive)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+uploadSpecFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	verifyExistInArtifactory(tests.GetUploadAsArchive(), searchFilePath, t)

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
	artifactoryCli.Exec("download", "--spec="+downloadSpecFile)
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(tests.GetDownloadArchiveAndExplode(), paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryUploadAsArchiveWithExplodeAndSymlinks(t *testing.T) {
	initArtifactoryTest(t)

	uploadSpecFile, err := tests.CreateSpec(tests.UploadAsArchive)
	assert.NoError(t, err)
	err = artifactoryCli.Exec("upload", "--spec="+uploadSpecFile, "--symlinks", "--explode")
	assert.Error(t, err)

	cleanArtifactoryTest()
}

func TestArtifactoryUploadAsArchiveToDir(t *testing.T) {
	initArtifactoryTest(t)

	uploadSpecFile, err := tests.CreateSpec(tests.UploadAsArchiveToDir)
	assert.NoError(t, err)
	err = artifactoryCli.Exec("upload", "--spec="+uploadSpecFile)
	assert.Error(t, err)

	cleanArtifactoryTest()
}

func TestArtifactoryUploadAsArchiveWithIncludeDirs(t *testing.T) {
	initArtifactoryTest(t)
	assert.NoError(t, createEmptyTestDir())
	uploadSpecFile, err := tests.CreateSpec(tests.UploadAsArchiveEmptyDirs)
	assert.NoError(t, err)
	err = artifactoryCli.Exec("upload", "--spec="+uploadSpecFile)
	assert.NoError(t, err)

	// Check the empty directories inside the archive by downloading and exploding it.
	downloadSpecFile, err := tests.CreateSpec(tests.DownloadAndExplodeArchives)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+downloadSpecFile)
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	downloadedEmptyDirs := tests.GetDownloadArchiveAndExplodeWithIncludeDirs()
	// Verify dirs exists.
	tests.VerifyExistLocally(downloadedEmptyDirs, paths, t)
	// Verify empty dirs.
	verifyEmptyDirs(t, downloadedEmptyDirs)

	// Check the empty directories inside the archive by downloading without exploding it, using os "unzip" command.
	tests.RemoveTempDirAndAssert(t, tests.Out)
	assert.NoError(t, os.MkdirAll(tests.Out, 0777))
	downloadSpecFile, err = tests.CreateSpec(tests.DownloadWithoutExplodeArchives)
	artifactoryCli.Exec("download", "--spec="+downloadSpecFile)
	// Change working directory to the zip file's location and unzip it.
	chdirCallback := tests.ChangeDirWithCallback(t, path.Join(tests.Out, "archive", "archive"))
	defer chdirCallback()
	cmd := exec.Command("unzip", "archive.zip")
	assert.NoError(t, errorutils.CheckError(cmd.Run()))
	chdirCallback()
	verifyEmptyDirs(t, downloadedEmptyDirs)
	cleanArtifactoryTest()
}

func verifyEmptyDirs(t *testing.T, dirPaths []string) {
	for _, path := range dirPaths {
		empty, err := fileutils.IsDirEmpty(path)
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
	initArtifactoryTest(t)

	outDirPath := tests.Out + string(os.PathSeparator)
	// Upload all testdata/a/ to repo1/syncDir/
	artifactoryCli.Exec("upload", path.Join("testdata", "a", "*"), tests.RtRepo1+"/syncDir/", "--flat=false")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep1(), searchFilePath, t)

	// Download repo1/syncDir/ to out/
	artifactoryCli.Exec("download", tests.RtRepo1+"/syncDir/", tests.Out+"/")
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(tests.GetExpectedSyncDeletesDownloadStep2(), paths, t)

	// Download repo1/syncDir/ to out/ with flat=true and sync out/
	artifactoryCli.Exec("download", tests.RtRepo1+"/syncDir/", outDirPath, "--flat=true", "--sync-deletes="+outDirPath)
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetExpectedSyncDeletesDownloadStep3(), paths, t)

	// Download all files ended with 2.in from repo1/syncDir/ to out/ and sync out/
	artifactoryCli.Exec("download", tests.RtRepo1+"/syncDir/*2.in", outDirPath, "--flat=true", "--sync-deletes="+outDirPath)
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetExpectedSyncDeletesDownloadStep4(), paths, t)

	// Download repo1/syncDir/ to out/, exclude the pattern "*c*.in" and sync out/
	artifactoryCli.Exec("download", tests.RtRepo1+"/syncDir/", outDirPath, "--sync-deletes="+outDirPath+"syncDir"+string(os.PathSeparator), "--exclusions=*/syncDir/testdata/*c*in")
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out+string(os.PathSeparator)+"syncDir"+string(os.PathSeparator), false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetSyncExpectedDeletesDownloadStep5(), paths, t)

	// Delete all files from repo1/syncDir/
	artifactoryCli.Exec("delete", tests.RtRepo1+"/syncDir/")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyDoesntExistInArtifactory(searchFilePath, t)

	// Upload all testdata/archives/ to repo1/syncDir/
	artifactoryCli.Exec("upload", path.Join("testdata", "archives", "*"), tests.RtRepo1+"/syncDir/", "--flat=false")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSyncExpectedDeletesDownloadStep6(), searchFilePath, t)

	// Download repo1/syncDir/ to out/ and sync out/
	artifactoryCli.Exec("download", tests.RtRepo1+"/syncDir/", outDirPath, "--sync-deletes="+outDirPath+"syncDir"+string(os.PathSeparator))
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out+string(os.PathSeparator)+"syncDir"+string(os.PathSeparator), false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetSyncExpectedDeletesDownloadStep7(), paths, t)

	cleanArtifactoryTest()
}

// After syncDeletes we must make sure that the content of the synced directory contains the last operation result only.
// Therefore we verify that there are no other files in the synced directory, other than the list of the expected files.
func checkSyncedDirContent(expected, actual []string, t *testing.T) {
	// Check if all expected files are actually exist
	tests.VerifyExistLocally(expected, actual, t)
	// Check if all the existing files were expected
	err := isExclusivelyExistLocally(expected, actual)
	assert.NoError(t, err)
}

// Check if only the files we were expect, exist locally, i.e return an error if there is a local file we didn't expect.
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
	initArtifactoryTest(t)
	tempDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	assert.NoError(t, os.Setenv(coreutils.HomeDir, tempDirPath))
	assert.NoError(t, os.Setenv(tests.HttpsProxyEnvVar, "1024"))
	defer func() {
		tests.RemoveTempDirAndAssert(t, tempDirPath)
		assert.NoError(t, os.Unsetenv(coreutils.HomeDir))
		assert.NoError(t, os.Unsetenv(tests.HttpsProxyEnvVar))
	}()
	go cliproxy.StartLocalReverseHttpProxy(serverDetails.ArtifactoryUrl, false)

	// The two certificate files are created by the reverse proxy on startup in the current directory.
	defer tests.RemoveAndAssert(t, certificate.KEY_FILE)
	defer tests.RemoveAndAssert(t, certificate.CERT_FILE)
	// Let's wait for the reverse proxy to start up.
	err = checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", false)
	assert.NoError(t, err)

	fileSpec := spec.NewBuilder().Pattern(tests.RtRepo1 + "/*.zip").Recursive(true).BuildSpec()
	assert.NoError(t, err)
	parsedUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	serverDetails.ArtifactoryUrl = "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()

	// The server is using self-signed certificates
	// Without loading the certificates (or skipping verification) we expect all actions to fail due to error: "x509: certificate signed by unknown authority"
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err := searchCmd.Search()
	if reader != nil {
		assert.NoError(t, reader.Close())
	}
	_, isUrlErr := err.(*url.Error)
	assert.True(t, isUrlErr, "Expected a connection failure, since reverse proxy didn't load self-signed-certs. Connection however is successful", err)

	// Set insecureTls to true and run again. We expect the command to succeed.
	serverDetails.InsecureTls = true
	searchCmd = generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	assert.NoError(t, reader.Close())

	// Set insecureTls back to false.
	// Copy the server certificates to the CLI security dir and run again. We expect the command to succeed.
	serverDetails.InsecureTls = false
	certsPath, err := coreutils.GetJfrogCertsDir()
	assert.NoError(t, err)
	err = fileutils.CopyFile(certsPath, certificate.KEY_FILE)
	assert.NoError(t, err)
	err = fileutils.CopyFile(certsPath, certificate.CERT_FILE)
	assert.NoError(t, err)
	searchCmd = generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	assert.NoError(t, reader.Close())

	serverDetails.ArtifactoryUrl = artAuth.GetUrl()
	cleanArtifactoryTest()
}

// Test client certificates with Artifactory. For the test, we set up a reverse proxy server.
func TestArtifactoryClientCert(t *testing.T) {
	initArtifactoryTest(t)
	tempDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	assert.NoError(t, os.Setenv(coreutils.HomeDir, tempDirPath))
	assert.NoError(t, os.Setenv(tests.HttpsProxyEnvVar, "1025"))
	defer func() {
		tests.RemoveTempDirAndAssert(t, tempDirPath)
		assert.NoError(t, os.Unsetenv(coreutils.HomeDir))
		assert.NoError(t, os.Unsetenv(tests.HttpsProxyEnvVar))
	}()
	go cliproxy.StartLocalReverseHttpProxy(serverDetails.ArtifactoryUrl, true)

	// The two certificate files are created by the reverse proxy on startup in the current directory.
	defer tests.RemoveAndAssert(t, certificate.KEY_FILE)
	defer tests.RemoveAndAssert(t, certificate.CERT_FILE)
	// Let's wait for the reverse proxy to start up.
	err = checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", true)
	assert.NoError(t, err)

	fileSpec := spec.NewBuilder().Pattern(tests.RtRepo1 + "/*.zip").Recursive(true).BuildSpec()
	assert.NoError(t, err)
	parsedUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	serverDetails.ArtifactoryUrl = "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()
	serverDetails.InsecureTls = true

	// The server is requiring client certificates
	// Without loading a valid client certificate, we expect all actions to fail due to error: "tls: bad certificate"
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err := searchCmd.Search()
	if reader != nil {
		assert.NoError(t, reader.Close())
	}
	_, isUrlErr := err.(*url.Error)
	assert.True(t, isUrlErr, "Expected a connection failure, since client did not provide a client certificate. Connection however is successful")

	// Inject client certificates, we expect the search to succeed
	serverDetails.ClientCertPath = certificate.CERT_FILE
	serverDetails.ClientCertKeyPath = certificate.KEY_FILE

	searchCmd = generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(fileSpec)
	reader, err = searchCmd.Search()
	if reader != nil {
		assert.NoError(t, reader.Close())
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

// Due the fact that go read the HTTP_PROXY and the HTTPS_PROXY
// argument only once we can't set the env var for specific test.
// We need to start a new process with the env var set to the value we want.
// We decide which var to set by the rtUrl scheme.
func TestArtifactoryProxy(t *testing.T) {
	initArtifactoryTest(t)
	rtUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	assert.NoError(t, err)
	var proxyTestArgs []string
	var httpProxyEnv string
	testArgs := []string{"-test.artifactoryProxy=true", "-jfrog.url=" + *tests.JfrogUrl, "-jfrog.user=" + *tests.JfrogUser, "-jfrog.password=" + *tests.JfrogPassword, "-jfrog.sshKeyPath=" + *tests.JfrogSshKeyPath, "-jfrog.sshPassphrase=" + *tests.JfrogSshPassphrase, "-jfrog.adminToken=" + *tests.JfrogAccessToken}
	if rtUrl.Scheme == "https" {
		assert.NoError(t, os.Setenv(tests.HttpsProxyEnvVar, "1026"))
		defer func() {
			assert.NoError(t, os.Unsetenv(tests.HttpsProxyEnvVar))
		}()
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
		assert.NoError(t, reader.Close())
	}
	serverDetails.ArtifactoryUrl = artAuth.GetUrl()
}

func prepareArtifactoryUrlForProxyTest(t *testing.T) string {
	rtUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	assert.NoError(t, err)
	rtHost, port, err := net.SplitHostPort(rtUrl.Host)
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
		assert.NoError(t, reader.Close())
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
			if _, err := os.Stat(certificate.CERT_FILE); os.IsNotExist(err) {
				log.Info("Waiting for certificate to appear...")
				time.Sleep(time.Second)
				continue
			}

			if _, err := os.Stat(certificate.KEY_FILE); os.IsNotExist(err) {
				log.Info("Waiting for key to appear...")
				time.Sleep(time.Second)
				continue
			}

			break
		}

		cert, err := tls.LoadX509KeyPair(certificate.CERT_FILE, certificate.KEY_FILE)
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
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			time.Sleep(time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed while waiting for the proxy server to be accessible")
}

func TestXrayScanBuild(t *testing.T) {
	initArtifactoryTest(t)
	xrayServerPort := xray.StartXrayMockServer()
	serverUrl := "--url=http://localhost:" + strconv.Itoa(xrayServerPort)
	artifactoryCommandExecutor := tests.NewJfrogCli(execMain, "jfrog rt", serverUrl+getArtifactoryTestCredentials())
	artifactoryCommandExecutor.Exec("build-scan", xray.CleanScanBuildName, "3")

	cleanArtifactoryTest()
}

func TestArtifactorySetProperties(t *testing.T) {
	initArtifactoryTest(t)
	// Upload a file.
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/a.in")
	// Set the 'prop=red' property to the file.
	artifactoryCli.Exec("sp", tests.RtRepo1+"/a.*", "prop=red")
	// Now let's change the property value, by searching for the 'prop=red'.
	specFile, err := tests.CreateSpec(tests.SetDeletePropsSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("sp", "prop=green", "--spec="+specFile)

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
	initArtifactoryTest(t)
	targetPath := path.Join(tests.RtRepo1, "a$+~&^a#")
	// Upload a file with special chars.
	artifactoryCli.Exec("upload", "testdata/a/a1.in", targetPath)
	// Set the 'prop=red' property to the file.
	artifactoryCli.Exec("sp", targetPath, "prop=red")

	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	resultItems, err := searchInArtifactory(searchSpec, t)
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
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "testdata/a/a*.in", tests.RtRepo1+"/")
	artifactoryCli.Exec("sp", tests.RtRepo1+"/*", "prop=val", "--exclusions=*/*a1.in;*/*a2.in")
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
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "testdata/a/a*.in", tests.RtRepo1+"/")
	artifactoryCli.Exec("sp", tests.RtRepo1+"/*", "prop=val", "--exclusions=*/*a1.in;*/*a2.in")
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
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "testdata/a/a*.in", tests.RtRepo1+"/a/")
	artifactoryCli.Exec("sp", tests.RtRepo1+"/a/*", "color=yellow;prop=red;status=ok")
	// Delete the 'color' property.
	artifactoryCli.Exec("delp", tests.RtRepo1+"/a/*", "color")
	// Delete the 'status' property, by a spec which filters files by 'prop=red'.
	specFile, err := tests.CreateSpec(tests.SetDeletePropsSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("delp", "status", "--spec="+specFile)

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
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "testdata/a/a*.in", tests.RtRepo1+"/")
	artifactoryCli.Exec("sp", tests.RtRepo1+"/*", "prop=val")

	artifactoryCli.Exec("delp", tests.RtRepo1+"/*", "prop", "--exclusions=*/*a1.in;*/*a2.in")
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
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "testdata/a/a*.in", tests.RtRepo1+"/")
	artifactoryCli.Exec("sp", tests.RtRepo1+"/*", "prop=val")

	artifactoryCli.Exec("delp", tests.RtRepo1+"/*", "prop", "--exclusions=*/*a1.in;*/*a2.in")
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

func TestArtifactoryUploadFromHomeDir(t *testing.T) {
	initArtifactoryTest(t)
	testFileRel, testFileAbs := createFileInHomeDir(t, "cliTestFile.txt")
	artifactoryCli.Exec("upload", testFileRel, tests.RtRepo1, "--recursive=false", "--flat=true")
	searchTxtPath, err := tests.CreateSpec(tests.SearchTxt)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetTxtUploadExpectedRepo1(), searchTxtPath, t)
	os.Remove(testFileAbs)
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
	initArtifactoryTest(t)
	// Upload files
	artifactoryCli.Exec("upload", "testdata/a/a*", tests.RtRepo1, "--exclusions=*a2*;*a3.in", "--flat=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli1Regex(t *testing.T) {
	initArtifactoryTest(t)
	// Upload files
	artifactoryCli.Exec("upload", "testdata/a/a(.*)", tests.RtRepo1, "--exclusions=(.*)a2.*;.*a3.in", "--regexp=true", "--flat=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli2Wildcard(t *testing.T) {
	initArtifactoryTest(t)

	// Create temp dir
	absDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err, "Couldn't create dir")
	defer tests.RemoveTempDirAndAssert(t, absDirPath)

	// Create temp files
	d1 := []byte("test file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile1.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile2.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")

	// Upload files
	artifactoryCli.Exec("upload", filepath.ToSlash(absDirPath)+"/*", tests.RtRepo1, "--exclusions=*cliTestFile1*", "--flat=true")

	// Check files exists in artifactory
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory([]string{tests.RtRepo1 + "/cliTestFile2.in"}, searchFilePath, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli2Regex(t *testing.T) {
	initArtifactoryTest(t)

	// Create temp dir
	absDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err, "Couldn't create dir")
	defer tests.RemoveTempDirAndAssert(t, absDirPath)

	// Create temp files
	d1 := []byte("test file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile1.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile2.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")

	// Upload files
	artifactoryCli.Exec("upload", filepath.ToSlash(absDirPath)+"(.*)", tests.RtRepo1, "--exclusions=(.*c)liTestFile1.*", "--regexp=true", "--flat=true")

	// Check files exists in artifactory
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory([]string{tests.RtRepo1 + "/cliTestFile2.in"}, searchFilePath, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeBySpecWildcard(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadSpecExclude)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	// Validate files are moved by build number
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetUploadSpecExcludeRepo1(), searchFilePath, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeBySpecRegex(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadSpecExcludeRegex)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	// Validate files are moved by build number
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetUploadSpecExcludeRepo1(), searchFilePath, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadWithRegexEscaping(t *testing.T) {
	initArtifactoryTest(t)
	// Upload files
	artifactoryCli.Exec("upload", "testdata/regexp"+"(.*)"+"\\."+".*", tests.RtRepo1, "--regexp=true", "--flat=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)

	verifyExistInArtifactory([]string{tests.RtRepo1 + "/has.dot"}, searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopySpec(t *testing.T) {
	testMoveCopySpec("copy", t)
}

func TestArtifactoryMoveSpec(t *testing.T) {
	testMoveCopySpec("move", t)
}

func testMoveCopySpec(command string, t *testing.T) {
	initArtifactoryTest(t)
	preUploadBasicTestResources()
	specFile, err := tests.CreateSpec(tests.CopyMoveSimpleSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec(command, "--spec="+specFile)

	// Verify files exist in target location successfully
	searchMovedCopiedSpec, err := tests.CreateSpec(tests.SearchTargetInRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetMoveCopySpecExpected(), searchMovedCopiedSpec, t)

	searchOriginalSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)

	if command == "copy" {
		// Verify original files still exist
		verifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchOriginalSpec, t)
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
	initArtifactoryTest(t)
	// Path to local file
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a", "a1.in")
	// Path to valid symLink
	validLink := filepath.Join(tests.GetTestResourcesPath()+"a", "link")

	// Link valid symLink to local file
	err := os.Symlink(localFile, validLink)
	assert.NoError(t, err)

	// Upload symlink to artifactory
	artifactoryCli.Exec("u", validLink, tests.RtRepo1, "--symlinks=true", "--flat=true")

	// Delete the local symlink
	err = os.Remove(validLink)
	assert.NoError(t, err)

	// Download symlink from artifactory
	artifactoryCli.Exec("dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true")

	// Should be valid if successful
	validateSymLink(validLink, localFile, t)

	// Delete symlink and clean
	os.Remove(validLink)
	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Unlink and delete the pointed file.
// Download the symlink which was uploaded with validation. The command should failed.
func TestValidateBrokenSymlink(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	// Create temporary file in resourcesPath/a/
	tmpFile, err := ioutil.TempFile(tests.GetTestResourcesPath()+"a/", "a.in.")
	if assert.NoError(t, err) {
		tmpFile.Close()
	}
	localFile := tmpFile.Name()

	// Path to the symLink
	linkPath := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	// Link to the temporary file
	err = os.Symlink(localFile, linkPath)
	assert.NoError(t, err)

	// Upload symlink to artifactory
	artifactoryCli.Exec("u", linkPath, tests.RtRepo1, "--symlinks=true", "--flat=true")

	// Delete the local symlink and the temporary file
	err = os.Remove(linkPath)
	assert.NoError(t, err)
	err = os.Remove(localFile)
	assert.NoError(t, err)

	// Try downloading symlink from artifactory. Since the link should be broken, it shouldn't be downloaded
	err = artifactoryCli.Exec("dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true")
	if !assert.Error(t, err, "A broken symLink was downloaded although validate-symlinks flag was set to true") {
		os.Remove(linkPath)
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
	initArtifactoryTest(t)

	// Creating broken symlink
	os.Mkdir(tests.Out, 0777)
	linkToNonExistingPath := filepath.Join(tests.Out, "link_to_non_existing_path")
	err := os.Symlink("non_existing_path", linkToNonExistingPath)
	assert.NoError(t, err)

	// This command should succeed because all artifacts are excluded.
	artifactoryCli.Exec("u", filepath.Join(tests.Out, "*"), tests.RtRepo1, "--symlinks=true", "--exclusions=*")
	cleanArtifactoryTest()
}

// Upload symlink to Artifactory using wildcard pattern and the link content checksum
// Download the symlink which was uploaded.
// validate the symlink content checksum.
func TestSymlinkWildcardPathHandling(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a/", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link*")
	artifactoryCli.Exec("u", link1, tests.RtRepo1, "--symlinks=true", "--flat=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

func TestUploadWithArchiveAndSymlink(t *testing.T) {
	initArtifactoryTest(t)
	// Path to local file with a different name from symlinkTarget
	testFile := filepath.Join(tests.GetTestResourcesPath(), "a", "a1.in")
	tmpDir, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.RemoveAll(tmpDir)) }()
	err = fileutils.CopyFile(tmpDir, testFile)
	assert.NoError(t, err)
	// Link valid symLink to local file
	symlinkTarget := filepath.Join(tmpDir, "a1.in")
	err = os.Symlink(symlinkTarget, filepath.Join(tmpDir, "symlink"))
	assert.NoError(t, err)
	// Upload symlink and local file to artifactory
	assert.NoError(t, artifactoryCli.Exec("u", tmpDir+"/*", tests.RtRepo1+"/test-archive.zip", "--archive=zip", "--symlinks=true", "--flat=true"))
	assert.NoError(t, os.RemoveAll(tmpDir))
	assert.NoError(t, os.Mkdir(tmpDir, 0777))
	assert.NoError(t, artifactoryCli.Exec("download", tests.RtRepo1+"/test-archive.zip", tmpDir+"/", "--explode=true"))
	// Validate
	assert.True(t, fileutils.IsPathExists(filepath.Join(tmpDir, "a1.in"), false), "Failed to download file from Artifactory")
	validateSymLink(filepath.Join(tmpDir, "symlink"), symlinkTarget, t)

	cleanArtifactoryTest()
}

func TestUploadWithArchiveAndSymlinkZipSlip(t *testing.T) {
	initArtifactoryTest(t)
	symlinkTarget := filepath.Join(tests.GetTestResourcesPath(), "a", "a2.in")
	tmpDir, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	defer func() { assert.NoError(t, os.RemoveAll(tmpDir)) }()
	// Link symLink to local file, outside of the extraction directory
	err = os.Symlink(symlinkTarget, filepath.Join(tmpDir, "symlink"))
	assert.NoError(t, err)

	// Upload symlink and local file to artifactory
	assert.NoError(t, artifactoryCli.Exec("u", tmpDir+"/*", tests.RtRepo1+"/test-archive.zip", "--archive=zip", "--symlinks=true", "--flat=true"))
	assert.NoError(t, os.RemoveAll(tmpDir))
	assert.NoError(t, os.Mkdir(tmpDir, 0777))

	// Discard output logging to prevent negative logs
	previousLogger := tests.RedirectLogOutputToNil()
	defer log.SetLogger(previousLogger)

	// Make sure download failed
	assert.Error(t, artifactoryCli.Exec("download", tests.RtRepo1+"/test-archive.zip", tmpDir+"/", "--explode=true"))
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirHandling(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", link, tests.RtRepo1, "--symlinks=true", "--recursive=true", "--flat=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirWildcardHandling(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "lin*")
	artifactoryCli.Exec("u", link1, tests.RtRepo1, "--symlinks=true", "--recursive=true", "--flat=true", "--flat=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
// The test create circular links and the test suppose to prune the circular searching.
func TestSymlinkInsideSymlinkDirWithRecursionIssueUpload(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localDirPath := filepath.Join(tests.GetTestResourcesPath(), "a")
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link1")
	err := os.Symlink(localDirPath, link1)
	assert.NoError(t, err)
	localFilePath := filepath.Join(tests.GetTestResourcesPath()+"a/", "a1.in")
	link2 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link2")
	err = os.Symlink(localFilePath, link2)
	assert.NoError(t, err)

	artifactoryCli.Exec("u", localDirPath+"/link*", tests.RtRepo1, "--symlinks=true", "--recursive=true", "--flat=true")
	err = os.Remove(link1)
	assert.NoError(t, err)

	err = os.Remove(link2)
	assert.NoError(t, err)

	artifactoryCli.Exec("dl", tests.RtRepo1+"/link*", tests.GetTestResourcesPath()+"a/")
	validateSymLink(link1, localDirPath, t)
	os.Remove(link1)
	validateSymLink(link2, localFilePath, t)
	os.Remove(link2)
	cleanArtifactoryTest()
}

func validateSymLink(localLinkPath, localFilePath string, t *testing.T) {
	// In macOS, localFilePath may lead to /var/folders/dn instead of /private/var/folders/dn.
	// EvalSymlinks in localLinkPath should fix it.
	localFilePath, err := filepath.EvalSymlinks(localLinkPath)
	assert.NoError(t, err)

	exists := fileutils.IsPathSymlink(localLinkPath)
	assert.True(t, exists, "failed to download symlinks from artifactory")
	symlinks, err := filepath.EvalSymlinks(localLinkPath)
	assert.NoError(t, err, "can't eval symlinks")
	assert.Equal(t, localFilePath, symlinks, "Symlinks wasn't created as expected")
}

func TestArtifactoryDeleteNoSpec(t *testing.T) {
	initArtifactoryTest(t)
	testArtifactorySimpleDelete(t, "")
}

func TestArtifactoryDeleteBySpec(t *testing.T) {
	initArtifactoryTest(t)
	deleteSpecPath, err := tests.CreateSpec(tests.DeleteSimpleSpec)
	assert.NoError(t, err)
	testArtifactorySimpleDelete(t, deleteSpecPath)
}

func testArtifactorySimpleDelete(t *testing.T, deleteSpecPath string) {
	preUploadBasicTestResources()

	// Verify exists before deleting
	searchSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchSpec, t)

	if deleteSpecPath != "" {
		artifactoryCli.Exec("delete", "--spec="+deleteSpecPath)
	} else {
		artifactoryCli.Exec("delete", tests.RtRepo1+"/test_resources/b/*")
	}

	verifyExistInArtifactory(tests.GetSimpleDelete(), searchSpec, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFolderWithWildcard(t *testing.T) {
	initArtifactoryTest(t)
	preUploadBasicTestResources()

	// Verify exists before deleting
	searchSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchSpec, t)

	artifactoryCli.Exec("delete", tests.RtRepo1+"/test_resources/*/c")

	verifyExistInArtifactory(tests.GetDeleteFolderWithWildcard(), searchSpec, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFolderCompletelyNoSpec(t *testing.T) {
	testArtifactoryDeleteFoldersNoSpec(t, false)
}

func TestArtifactoryDeleteFolderContentNoSpec(t *testing.T) {
	testArtifactoryDeleteFoldersNoSpec(t, true)
}

func testArtifactoryDeleteFoldersNoSpec(t *testing.T, contentOnly bool) {
	initArtifactoryTest(t)
	preUploadBasicTestResources()

	// Verify exists before deleting
	searchSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchSpec, t)

	// Delete folder
	deletePath := tests.RtRepo1 + "/test_resources"
	// End with separator if content only
	if contentOnly {
		deletePath += "/"
	}
	artifactoryCli.Exec("delete", deletePath)

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
	initArtifactoryTest(t)
	preUploadBasicTestResources()

	// Verify exists before deleting
	searchSpec, err := tests.CreateSpec(tests.SearchRepo1TestResources)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetRepo1TestResourcesExpected(), searchSpec, t)

	deleteSpecPath, err := tests.CreateSpec(specPath)
	assert.NoError(t, err)
	artifactoryCli.Exec("delete", "--spec="+deleteSpecPath)

	completeSearchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyDoesntExistInArtifactory(completeSearchSpec, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteExcludeByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Delete by pattern
	artifactoryCli.Exec("del", tests.RtRepo1+"/data/", "--exclusions=*/*b1.in;*/*b2.in;*/*b3.in;*/*c1.in")

	// Validate files are deleted
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteExclusionsByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Delete by pattern
	artifactoryCli.Exec("del", tests.RtRepo1+"/data/", "--exclusions=*/*b1.in;*/*b2.in;*/*b3.in;*/*c1.in")

	// Validate files are deleted
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteExclusionsBySpec(t *testing.T) {
	initArtifactoryTest(t)
	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	specFile, err := tests.CreateSpec(tests.DelSpecExclusions)
	assert.NoError(t, err)

	// Delete by pattern
	artifactoryCli.Exec("del", "--spec="+specFile)

	// Validate files are deleted
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

// Deleting files when one file name is a prefix to another in the same dir
func TestArtifactoryDeletePrefixFiles(t *testing.T) {
	initArtifactoryTest(t)

	// Prepare search command
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadPrefixFiles)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	// Delete by pattern
	artifactoryCli.Exec("delete", tests.RtRepo1+"/*")

	// Validate files are deleted
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	length, err := reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, 0, length)
	assert.NoError(t, reader.Close())

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteByProps(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadWithPropsSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	// Set properties to the directories as well (and their content)
	artifactoryCli.Exec("sp", tests.RtRepo1+"/a/b/", "D=5", "--include-dirs")
	artifactoryCli.Exec("sp", tests.RtRepo1+"/a/b/c/", "D=2", "--include-dirs")

	//  Set the property D=5 to c1.in, which is a different value then its directory c/
	artifactoryCli.Exec("sp", tests.RtRepo1+"/a/b/c/c1.in", "D=5")

	// Prepare search command
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())

	// Delete all artifacts with D=5 but without c=3
	artifactoryCli.Exec("delete", tests.RtRepo1+"/*", "--props=D=5", "--exclude-props=c=3")

	// Search all artifacts in repo1
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	assert.NoError(t, reader.GetError())
	reader.Reset()

	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchResultAfterDeleteByPropsStep1())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Delete all artifacts with c=3 but without a=1
	artifactoryCli.Exec("delete", tests.RtRepo1+"/*", "--props=c=3", "--exclude-props=a=1")

	// Search all artifacts in repo1
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	assert.NoError(t, reader.GetError())
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchResultAfterDeleteByPropsStep2())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Delete all artifacts with a=1 but without b=3&c=3
	artifactoryCli.Exec("delete", tests.RtRepo1+"/*", "--props=a=1", "--exclude-props=b=3;c=3")

	// Search all artifacts in repo1
	reader, err = searchCmd.Search()
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	assert.NoError(t, reader.GetError())
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchResultAfterDeleteByPropsStep3())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryMultipleFileSpecsUpload(t *testing.T) {
	initArtifactoryTest(t)
	specFile, err := tests.CreateSpec(tests.UploadMultipleFileSpecs)
	assert.NoError(t, err)
	resultSpecFile, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	verifyExistInArtifactory(tests.GetMultipleFileSpecs(), resultSpecFile, t)
	verifyExistInArtifactoryByProps([]string{tests.RtRepo1 + "/multiple/properties/testdata/a/b/b2.in"}, tests.RtRepo1+"/*/properties/*.in", "searchMe=true", t)
	cleanArtifactoryTest()
}

func TestArtifactorySimplePlaceHolders(t *testing.T) {
	initArtifactoryTest(t)
	specFile, err := tests.CreateSpec(tests.UploadSimplePlaceholders)
	assert.NoError(t, err)

	resultSpecFile, err := tests.CreateSpec(tests.SearchSimplePlaceholders)
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", "--spec="+specFile)

	verifyExistInArtifactory(tests.GetSimplePlaceholders(), resultSpecFile, t)
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadRecursiveNonFlat(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := fileutils.GetFileSeparator() + filepath.Join("inner", "folder")
	canonicalPath := tests.Out + dirInnerPath

	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/(*)", tests.RtRepo1+"/{1}/", "--include-dirs=true", "--recursive=true", "--flat=false")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(filepath.Join(tests.Out, "inner", "folder", "folder"), false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFlatFolderUpload(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := fileutils.GetFileSeparator() + filepath.Join("inner", "folder")
	canonicalPath := tests.Out + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/(*)", tests.RtRepo1+"/{1}/", "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	// Non flat download
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(filepath.Join(canonicalPath, "folder"), false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryIncludeDirFlatNonEmptyFolderUpload(t *testing.T) {
	initArtifactoryTest(t)
	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"a/b/*", tests.RtRepo1, "--include-dirs=true", "--flat=true")
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/c", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryDownloadNotIncludeDirs(t *testing.T) {
	initArtifactoryTest(t)
	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"*/c", tests.RtRepo1, "--include-dirs=true", "--flat=true")
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--recursive=true")
	assert.False(t, fileutils.IsPathExists(tests.Out+"/c", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryDownloadFlatTrue(t *testing.T) {
	initArtifactoryTest(t)
	canonicalPath := tests.GetTestResourcesPath() + path.Join("an", "empty", "folder")
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)

	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"(a*)/*", tests.RtRepo1+"/{1}/", "--include-dirs=true", "--flat=true")
	// Download without include-dirs
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--recursive=true", "--flat=true")
	assert.False(t, fileutils.IsPathExists(tests.Out+"/c", false), "'c' folder shouldn't exist.")

	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true", "--flat=true")
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
	initArtifactoryTest(t)
	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"*/c", tests.RtRepo1, "--include-dirs=true", "--flat=true")
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/c", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryUploadFlatFolderWithFileAndInnerEmptyMatchingPattern(t *testing.T) {
	initArtifactoryTest(t)
	newFolderPath := tests.GetTestResourcesPath() + "a/b/c/d"
	err := os.MkdirAll(newFolderPath, 0777)
	assert.NoError(t, err)
	// We created an empty child folder to 'c' therefore 'c' is not longer a bottom chain and new 'd' inner directory is indeed bottom chain directory.
	// 'd' should uploaded and 'c' shouldn't
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"a/b/*", tests.RtRepo1, "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(newFolderPath)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.False(t, fileutils.IsPathExists(tests.Out+"/c", false), "'c' folder shouldn't exist")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/d", false), "bottom chain directory, 'd', is missing")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryUploadFlatFolderWithFileAndInnerEmptyMatchingPatternWithPlaceHolders(t *testing.T) {
	initArtifactoryTest(t)
	relativePath := "/b/c/d"
	fullPath := tests.GetTestResourcesPath() + "a/" + relativePath
	err := os.MkdirAll(fullPath, 0777)
	assert.NoError(t, err)
	// We created a empty child folder to 'c' therefore 'c' is not longer a bottom chain and new 'd' inner directory is indeed bottom chain directory.
	// 'd' should uploaded and 'c' shouldn't
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"a/(*)/*", tests.RtRepo1+"/{1}/", "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(fullPath)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+relativePath, false), "bottom chain directory, 'd', is missing")

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFlatFolderDownload1(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := fileutils.GetFileSeparator() + filepath.Join("inner", "folder")
	canonicalPath := tests.Out + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	// Flat true by default for upload, by using placeholder we indeed create folders hierarchy in Artifactory inner/folder/folder
	artifactoryCli.Exec("upload", tests.Out+"/(*)", tests.RtRepo1+"/{1}/", "--include-dirs=true")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	// Only the inner folder should be downland e.g 'folder'
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true", "--flat=true")
	assert.False(t, !fileutils.IsPathExists(filepath.Join(tests.Out, "folder"), false) &&
		fileutils.IsPathExists(filepath.Join(tests.Out, "inner"), false),
		"Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadRecursiveUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	assert.NoError(t, createEmptyTestDir())
	specFile, err := tests.CreateSpec(tests.UploadEmptyDirs)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	specFile, err = tests.CreateSpec(tests.DownloadEmptyDirs)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+specFile)
	assert.True(t, fileutils.IsPathExists(tests.Out+"/folder", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadNonRecursive(t *testing.T) {
	initArtifactoryTest(t)
	canonicalPath := filepath.Join(tests.Out, "inner", "folder")
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/", tests.RtRepo1, "--recursive=true", "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.RtRepo1, tests.Out+"/", "--include-dirs=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/folder", false), "Failed to download folder from Artifactory")
	assert.False(t, fileutils.IsPathExists(canonicalPath, false), "Path should be flat ")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderDownloadNonRecursive(t *testing.T) {
	initArtifactoryTest(t)
	canonicalPath := filepath.Join(tests.Out, "inner", "folder")
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/", tests.RtRepo1, "--recursive=true", "--include-dirs=true", "--flat=false")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.RtRepo1+"/*", "--recursive=false", "--include-dirs=true")
	assert.True(t, fileutils.IsPathExists(tests.Out, false), "Failed to download folder from Artifactory")
	assert.False(t, fileutils.IsPathExists(canonicalPath, false), "Path should be flat. ")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryChecksumDownload(t *testing.T) {
	initArtifactoryTest(t)

	var filePath = "testdata/a/a1.in"
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1, "--flat=true")
	testChecksumDownload(t, "/a1.in")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryChecksumDownloadRenameFileName(t *testing.T) {
	initArtifactoryTest(t)

	var filePath = "testdata/a/a1.in"
	artifactoryCli.Exec("upload", filePath, tests.RtRepo1, "--flat=true")
	testChecksumDownload(t, "/a1.out")
	// Cleanup
	cleanArtifactoryTest()
}

func testChecksumDownload(t *testing.T, outFileName string) {
	artifactoryCli.Exec("download", tests.RtRepo1+"/a1.in", tests.Out+outFileName)

	exists, err := fileutils.IsFileExists(tests.Out+outFileName, false)
	assert.NoError(t, err)
	assert.True(t, exists, "Failed to download file from Artifactory")

	firstFileInfo, _ := os.Stat(tests.Out + outFileName)
	firstDownloadTime := firstFileInfo.ModTime()

	artifactoryCli.Exec("download", tests.RtRepo1+"/a1.in", tests.Out+outFileName)
	secondFileInfo, _ := os.Stat(tests.Out + outFileName)
	secondDownloadTime := secondFileInfo.ModTime()

	assert.Equal(t, firstDownloadTime, secondDownloadTime, "Checksum download failed, the file was downloaded twice")
}

func TestArtifactoryDownloadByPatternAndBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryGenericBuildnameAndNumberFromEnv(t *testing.T) {
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	assert.NoError(t, os.Setenv(coreutils.BuildName, tests.RtBuildName1))
	assert.NoError(t, os.Setenv(coreutils.BuildNumber, buildNumberA))
	defer func() {
		assert.NoError(t, os.Unsetenv(coreutils.BuildName))
		assert.NoError(t, os.Unsetenv(coreutils.BuildNumber))
	}()
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	assert.NoError(t, os.Setenv(coreutils.BuildNumber, "11"))
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Publish buildInfo
	assert.NoError(t, os.Setenv(coreutils.BuildNumber, buildNumberA))
	artifactoryCli.Exec("build-publish")
	assert.NoError(t, os.Setenv(coreutils.BuildNumber, buildNumberB))
	artifactoryCli.Exec("build-publish")

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildNoPatternUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoPattern)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

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
	initArtifactoryTest(t)
	buildNumber := "1337"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	// Add build artifacts.
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Add build dependencies.
	artifactoryCliNoCreds := tests.NewJfrogCli(execMain, "jfrog rt", "")
	artifactoryCliNoCreds.Exec("bad", "--spec="+specFileB, tests.RtBuildName1, buildNumber)

	// Publish build.
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumber)
}

func TestArtifactoryDownloadByBuildWithDependenciesSpecNoPattern(t *testing.T) {
	prepareDownloadByBuildWithDependenciesTests(t)

	// Download with exclude-artifacts.
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecExcludeArtifacts)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+specFile)
	// Validate.
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadDoesntExist(), paths)
	assert.NoError(t, err)

	// Download deps-only.
	specFile, err = tests.CreateSpec(tests.BuildDownloadSpecDepsOnly)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+specFile)
	// Validate.
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetDownloadByBuildOnlyDeps(), paths)
	assert.NoError(t, err)

	// Download artifacts and deps.
	specFile, err = tests.CreateSpec(tests.BuildDownloadSpecIncludeDeps)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+specFile)
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
	artifactoryCli.Exec("download", tests.RtRepo1, "out/download/download_build_with_dependencies/", "--build="+tests.RtBuildName1, "--exclude-artifacts=true", "--flat=true")
	// Validate.
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err := tests.ValidateListsIdentical(tests.GetBuildDownloadDoesntExist(), paths)
	assert.NoError(t, err)

	// Download deps-only.
	artifactoryCli.Exec("download", tests.RtRepo1, "out/download/download_build_only_dependencies/", "--build="+tests.RtBuildName1, "--exclude-artifacts=true", "--include-deps=true", "--flat=true")
	// Validate.
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetDownloadByBuildOnlyDeps(), paths)
	assert.NoError(t, err)

	// Download artifacts and deps.
	artifactoryCli.Exec("download", tests.RtRepo1, "out/download/download_build_with_dependencies/", "--build="+tests.RtBuildName1, "--include-deps=true", "--flat=true")
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
	initArtifactoryTest(t)
	buildNumber := "10"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumber)
	assert.NoError(t, err)
	// Upload a file
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a10.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumber)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumber)

	// Download from different build number
	artifactoryCli.Exec("download", "--spec="+specFile)

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
	initArtifactoryTest(t)
	buildNumberA, buildNumberB, buildNumberC := "10", "11", "12"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumber)
	assert.NoError(t, err)

	// Upload 3 similar files to 3 different builds
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a10.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a11.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a12.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberC)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName2, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberC)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

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
	initArtifactoryTest(t)
	buildNumberA, buildNumberB, buildNumberC := "10", "11", "12"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumber)
	assert.NoError(t, err)

	// Upload 3 similar files to 2 different builds
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a10.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a11.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberB)
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a12.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberC)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberC)
	artifactoryCli.Exec("build-publish", tests.RtBuildName2, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName2, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadByShaAndBuildName(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildUsingSimpleDownload(t *testing.T) {
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "b", "a"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)

	// Download by build number, a1 should not be downloaded, b1 should
	artifactoryCli.Exec("download", tests.RtRepo1+"/data/a1.in", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(), "--build="+tests.RtBuildName1)
	artifactoryCli.Exec("download", tests.RtRepo1+"/data/b1.in", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(), "--build="+tests.RtBuildName1)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildUsingSimpleDownloadWithProject(t *testing.T) {
	initArtifactoryProjectTest(t)
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	assert.NoError(t, err)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)[:3]
	projectKey := "prj" + timestamp[len(timestamp)-3:]
	// Delete the project if already exists
	accessManager.DeleteProject(projectKey)

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
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA, "--project="+projectKey)

	// Publish buildInfo with project flag
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA, "--project="+projectKey)

	// Download by project, b1 should be downloaded
	artifactoryCli.Exec("download", tests.RtRepo1+"/data/b1.in", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(),
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
	initArtifactoryProjectTest(t)
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	assert.NoError(t, err)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)[:3]
	projectKey := "prj" + timestamp[len(timestamp)-3:]
	// Delete the project if already exists
	accessManager.DeleteProject(projectKey)

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
	assert.NoError(t, os.Setenv(coreutils.BuildName, tests.RtBuildName1))
	assert.NoError(t, os.Setenv(coreutils.BuildNumber, buildNumberA))
	assert.NoError(t, os.Setenv(coreutils.Project, projectKey))
	defer func() {
		assert.NoError(t, os.Unsetenv(coreutils.BuildName))
		assert.NoError(t, os.Unsetenv(coreutils.BuildNumber))
		assert.NoError(t, os.Unsetenv(coreutils.Project))
	}()
	// Upload files with buildName, buildNumber and project flags
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Publish buildInfo with project flag
	artifactoryCli.Exec("build-publish")

	// Download by project, b1 should be downloaded
	artifactoryCli.Exec("download", tests.RtRepo1+"/data/b1.in", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(),
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
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "b", "a"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)

	// Download by build number, a1 should not be downloaded, b1 should
	artifactoryCli.Exec("download", "*", filepath.Join(tests.Out, "download", "simple_by_build")+fileutils.GetFileSeparator(), "--build="+tests.RtBuildName1+"/"+buildNumberA)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownloadNoPattern(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByArchiveEntriesCli(t *testing.T) {
	initArtifactoryTest(t)
	uploadSpecFile, err := tests.CreateSpec(tests.ArchiveEntriesUpload)
	assert.NoError(t, err)

	// Upload archives
	artifactoryCli.Exec("upload", "--spec="+uploadSpecFile)

	// Create executor for running with retries
	retryExecutor := createRetryExecutorForArchiveEntries(tests.GetBuildArchiveEntriesDownloadCli(),
		[]string{"dl", tests.RtRepo1, "out/", "--archive-entries=(*)c1.in", "--flat=true"})

	// Perform download by archive-entries only the archives containing c1.in, and validate results
	assert.NoError(t, retryExecutor.Execute())

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByArchiveEntriesSpecificPathCli(t *testing.T) {
	initArtifactoryTest(t)
	uploadSpecFile, err := tests.CreateSpec(tests.ArchiveEntriesUpload)
	assert.NoError(t, err)

	// Upload archives
	artifactoryCli.Exec("upload", "--spec="+uploadSpecFile)

	// Create executor for running with retries
	retryExecutor := createRetryExecutorForArchiveEntries(tests.GetBuildArchiveEntriesSpecificPathDownload(),
		[]string{"dl", tests.RtRepo1, "out/", "--archive-entries=b/c/c1.in", "--flat=true"})

	// Perform download by archive-entries only the archives containing c1.in, and validate results
	assert.NoError(t, retryExecutor.Execute())

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByArchiveEntriesSpec(t *testing.T) {
	initArtifactoryTest(t)
	uploadSpecFile, err := tests.CreateSpec(tests.ArchiveEntriesUpload)
	assert.NoError(t, err)
	downloadSpecFile, err := tests.CreateSpec(tests.ArchiveEntriesDownload)
	assert.NoError(t, err)

	// Upload archives
	artifactoryCli.Exec("upload", "--spec="+uploadSpecFile)

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
		MaxRetries:      120,
		RetriesInterval: 1,
		ErrorMessage:    "Waiting for Artifactory to index archives...",
		ExecutionHandler: func() (bool, error) {
			err := validateDownloadByArchiveEntries(expected, args)
			if err != nil {
				return true, err
			}

			return false, nil
		},
	}
}

func validateDownloadByArchiveEntries(expected []string, args []string) error {
	// Execute the requested cli command
	artifactoryCli.Exec(args...)

	// Validate files are downloaded as expected
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	return tests.ValidateListsIdentical(expected, paths)
}

func TestArtifactoryDownloadExcludeByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--recursive=true")
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--recursive=true")

	// Download by pattern
	artifactoryCli.Exec("download", tests.RtRepo1, "out/download/aql_by_artifacts/", "--exclusions=*/*/a1.in;*/*a2.*;*/data/c2.in")

	// Validate files are excluded
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildExcludeDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExclusionsByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--recursive=true")
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--recursive=true")

	// Download by pattern
	artifactoryCli.Exec("download", tests.RtRepo1, "out/download/aql_by_artifacts/", "--exclusions=*/*/a1.in;*/*a2.*;*/data/c2.in")

	// Validate files are excluded
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildExcludeDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExclusionsBySpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile, err := tests.CreateSpec(tests.DownloadSpecExclusions)
	assert.NoError(t, err)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--recursive=true", "--flat=false")
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--recursive=true", "--flat=false")

	// Download by spec
	artifactoryCli.Exec("download", "--spec="+specFile)

	// Validate files are excluded
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildExcludeDownloadBySpec(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExclusionsBySpecOverride(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--recursive=true", "--flat=false")
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--recursive=true", "--flat=false")

	// Download by spec
	specFile, err := tests.CreateSpec(tests.DownloadSpecExclusions)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+specFile, "--exclusions=*a1.in;*a2.in;*c2.in")

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
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", link, tests.RtRepo1, "--symlinks=true", "--flat=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true", "--limit=1")
	validateSortLimitWithSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded with limit param.
// validate the symlink content checksum.
func TestArtifactorySortWithSymlink(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", link, tests.RtRepo1, "--symlinks=true", "--flat=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.RtRepo1+"/link", tests.GetTestResourcesPath()+"a/", "--validate-symlinks=true", "--sort-by=created")
	validateSortLimitWithSymLink(link, localFile, t)
	os.Remove(link)
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
	initArtifactoryTest(t)
	buildNumberA, buildNumberB, buildNumberC := "10", "11", "12"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumberWithSort)
	assert.NoError(t, err)
	// Upload 3 similar files to 2 different builds
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a10.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a11.in", "--build-name="+tests.RtBuildName2, "--build-number="+buildNumberB)
	artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/a12.in", "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberC)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberC)
	artifactoryCli.Exec("build-publish", tests.RtBuildName2, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName2, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--sort-by=created", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(filepath.Join(tests.Out, "download", "sort_limit_by_build"), false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadByShaAndBuildNameWithSort(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName2, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber: a* uploaded with build number "10", b* uploaded with build number "11"
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)

	// Copy by build name "cli-test-build" and build number "10" from spec, a* should be copied
	artifactoryCli.Exec("copy", "--spec="+specFile)

	// Validate files are Copied by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildPatternAllUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildPatternAllSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber: a* uploaded with build number "10", b* uploaded with build number "11"
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)

	// Copy by build name "cli-test-build" and build number "10" from spec, a* should be copied
	artifactoryCli.Exec("copy", "--spec="+specFile)

	// Validate files are Copied by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactorySortAndLimit(t *testing.T) {
	initArtifactoryTest(t)

	// Upload all testdata/a/ files
	artifactoryCli.Exec("upload", "testdata/a/(*)", tests.RtRepo1+"/data/{1}")

	// Download 1 sorted by name asc
	artifactoryCli.Exec("download", tests.RtRepo1+"/data/", "out/download/sort_limit/", "--sort-by=name", "--limit=1")

	// Download 3 sorted by depth desc
	artifactoryCli.Exec("download", tests.RtRepo1+"/data/", "out/download/sort_limit/", "--sort-by=depth", "--limit=3", "--sort-order=desc")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err := tests.ValidateListsIdentical(tests.GetSortAndLimit(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactorySortByCreated(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files separately so we can sort by created.
	artifactoryCli.Exec("upload", "testdata/created/or", tests.RtRepo1, `--target-props=k1=v1`, "--flat=true")
	artifactoryCli.Exec("upload", "testdata/created/o", tests.RtRepo1, "--flat=true")
	artifactoryCli.Exec("upload", "testdata/created/org", tests.RtRepo1, "--flat=true")

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
	// Verify the sort by checking if the item results are ordereds by asc.
	assert.True(t, reflect.DeepEqual(resultItems[0], tests.GetFirstSearchResultSortedByAsc()))
	assert.True(t, reflect.DeepEqual(resultItems[1], tests.GetSecondSearchResultSortedByAsc()))
	assert.True(t, reflect.DeepEqual(resultItems[2], tests.GetThirdSearchResultSortedByAsc()))

	assert.NoError(t, reader.Close())
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
	assert.NoError(t, reader.Close())

	// Cleanup
	cleanArtifactoryTest()
}
func TestArtifactoryOffset(t *testing.T) {
	initArtifactoryTest(t)

	// Upload all testdata/a/ files
	artifactoryCli.Exec("upload", "testdata/a/*", path.Join(tests.RtRepo1, "offset_test")+"/", "--flat=true")

	// Downloading files one by one, to check that the offset is working as expected.
	// Download only the first file, expecting to download a1.in
	artifactoryCli.Exec("download", tests.RtRepo1+"/offset_test/", tests.Out+"/", "--flat=true", "--sort-by=name", "--limit=1", "--offset=0")
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a1.in")}, paths, t)

	// Download the second file, expecting to download a2.in
	artifactoryCli.Exec("download", tests.RtRepo1+"/offset_test/", tests.Out+"/", "--flat=true", "--sort-by=name", "--limit=1", "--offset=1")
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a2.in")}, paths, t)

	// Download the third file, expecting to download a3.in
	artifactoryCli.Exec("download", tests.RtRepo1+"/offset_test/", tests.Out+"/", "--flat=true", "--sort-by=name", "--limit=1", "--offset=2")
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{filepath.Join(tests.Out, "a3.in")}, paths, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildOverridingByInlineFlag(t *testing.T) {
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber: b* uploaded with build number "10", a* uploaded with build number "11"
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)

	// Copy by build number: using override of build by flag from inline (no number set so LATEST build should be copied), a* should be copied
	artifactoryCli.Exec("copy", "--build="+tests.RtBuildName1, "--spec="+specFile)

	// Validate files are Copied by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryMoveByBuildUsingFlags(t *testing.T) {
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)

	// Move by build name and number
	artifactoryCli.Exec("move", "--build="+tests.RtBuildName1+"/11", "--spec="+specFile)

	// Validate files are moved by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryMoveNoSpec(t *testing.T) {
	testCopyMoveNoSpec("mv", tests.GetBuildBeforeMoveExpected(), tests.GetBuildMoveExpected(), t)
}

func TestArtifactoryMoveExcludeByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Move by pattern
	artifactoryCli.Exec("move", tests.RtRepo1+"/data/", tests.RtRepo2+"/", "--exclusions=*/*b*;*/*c*")

	// Validate excluded files didn't move
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryMoveExclusionsByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Move by pattern
	artifactoryCli.Exec("move", tests.RtRepo1+"/data/", tests.RtRepo2+"/", "--exclusions=*/*b*;*/*c*")

	// Validate excluded files didn't move
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryMoveExclusionsBySpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile, err := tests.CreateSpec(tests.MoveCopySpecExclusions)
	assert.NoError(t, err)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Move by spec
	artifactoryCli.Exec("move", "--spec="+specFile)

	// Validate excluded files didn't move
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteByLatestBuild(t *testing.T) {
	initArtifactoryTest(t)
	buildNumberA, buildNumberB := "10", "11"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+tests.RtBuildName1, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberA)
	artifactoryCli.Exec("build-publish", tests.RtBuildName1, buildNumberB)

	// Delete by build name and LATEST
	artifactoryCli.Exec("delete", "--build="+tests.RtBuildName1+"/LATEST", "--spec="+specFile)

	// Validate files are deleted by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

func TestGitLfsCleanup(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "testdata/gitlfs/(4b)(*)"
	artifactoryCli.Exec("upload", filePath, tests.RtLfsRepo+"/objects/4b/f4/{2}{1}", "--flat=true")
	artifactoryCli.Exec("upload", filePath, tests.RtLfsRepo+"/objects/4b/f4/", "--flat=true")
	refs := filepath.Join("refs", "heads", "*")
	dotGitPath := getCliDotGitPath(t)
	artifactoryCli.Exec("glc", dotGitPath, "--repo="+tests.RtLfsRepo, "--refs=HEAD,"+refs)
	gitlfsSpecFile, err := tests.CreateSpec(tests.GitLfsAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetGitLfsExpected(), gitlfsSpecFile, t)
	cleanArtifactoryTest()
}

func TestPing(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("ping")
	cleanArtifactoryTest()
}

type summaryExpected struct {
	errors  bool
	status  string
	success int64
	failure int64
}

func TestSummaryReport(t *testing.T) {
	initArtifactoryTest(t)

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
	initArtifactoryTest(t)
	testFailNoOpSummaryReport(t, true)
}

func TestSummaryReportFailNoOpFalse(t *testing.T) {
	initArtifactoryTest(t)
	testFailNoOpSummaryReport(t, false)
}

// Test summary after commands that do no actions, with/without failNoOp flag.
func testFailNoOpSummaryReport(t *testing.T, failNoOp bool) {
	initArtifactoryTest(t)

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
	buffer, previousLog := tests.RedirectLogOutputToBuffer()
	// Restore previous logger when the function returns
	defer log.SetLogger(previousLog)

	for _, cmd := range []string{"upload", "move", "copy", "delete", "set-props", "delete-props", "download"} {
		// Execute the cmd with it's args.
		err := artifactoryCli.Exec(append([]string{cmd}, argsMap[cmd]...)...)
		verifySummary(t, buffer, previousLog, err, expected)
	}
	cleanArtifactoryTest()
}

func TestUploadDetailedSummary(t *testing.T) {
	initArtifactoryTest(t)
	uploadCmd := generic.NewUploadCommand()
	fileSpec := spec.NewBuilder().Pattern(filepath.Join("testdata", "a", "a*.in")).Target(tests.RtRepo1).BuildSpec()
	uploadCmd.SetUploadConfiguration(createUploadConfiguration()).SetSpec(fileSpec).SetServerDetails(serverDetails).SetDetailedSummary(true)
	commands.Exec(uploadCmd)
	result := uploadCmd.Result()
	reader := result.Reader()
	assert.NoError(t, reader.GetError())
	defer reader.Close()
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
	initArtifactoryTest(t)
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber
	for i := 1; i <= 5; i++ {
		artifactoryCli.Exec("upload", "testdata/a/a1.in", tests.RtRepo1+"/data/", "--build-name="+tests.RtBuildName1, "--build-number="+strconv.Itoa(i))
		artifactoryCli.Exec("build-publish", tests.RtBuildName1, strconv.Itoa(i))
	}

	// Test discard by max-builds
	artifactoryCli.Exec("build-discard", tests.RtBuildName1, "--max-builds=3")
	jsonResponse := getAllBuildsByBuildName(client, tests.RtBuildName1, t, http.StatusOK)
	assert.Len(t, jsonResponse.Builds, 3, "Incorrect operation of build-discard by max-builds.")

	// Test discard with exclusion
	artifactoryCli.Exec("build-discard", tests.RtBuildName1, "--max-days=-1", "--exclude-builds=3,5")
	jsonResponse = getAllBuildsByBuildName(client, tests.RtBuildName1, t, http.StatusOK)
	assert.Len(t, jsonResponse.Builds, 2, "Incorrect operation of build-discard with exclusion.")

	// Test discard by max-days
	artifactoryCli.Exec("build-discard", tests.RtBuildName1, "--max-days=-1")
	jsonResponse = getAllBuildsByBuildName(client, tests.RtBuildName1, t, http.StatusNotFound)
	assert.Zero(t, jsonResponse, "Incorrect operation of build-discard by max-days.")

	//Cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	cleanArtifactoryTest()
}

// Tests compatibility to file paths with windows separators.
// Verifies the upload and download commands work as expected for inputs of both arguments and spec files.
func TestArtifactoryWinBackwardsCompatibility(t *testing.T) {
	initArtifactoryTest(t)
	if !coreutils.IsWindows() {
		t.Skip("Not running on Windows, skipping...")
	}
	uploadSpecFile, err := tests.CreateSpec(tests.WinSimpleUploadSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+uploadSpecFile)
	artifactoryCli.Exec("upload", "testdata\\\\a\\\\b\\\\*", tests.RtRepo1+"/compatibility_arguments/", "--exclusions=*b2.in;*c*")

	downloadSpecFile, err := tests.CreateSpec(tests.WinSimpleDownloadSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+downloadSpecFile)
	artifactoryCli.Exec("download", tests.RtRepo1+"/*arguments*", "out\\\\win\\\\", "--flat=true")

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetWinCompatibility(), paths)
	assert.NoError(t, err)
	cleanArtifactoryTest()
}

func TestArtifactorySearchIncludeDir(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--recursive", "--flat=false")

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
	assert.NoError(t, reader.GetError())
	reader.Reset()

	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchNotIncludeDirsFiles())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Search with IncludeDirs
	searchCmd.SetSpec(searchSpecBuilder.IncludeDirs(true).BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	assert.NoError(t, reader.GetError())
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchIncludeDirsFiles())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactorySearchProps(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadWithPropsSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile, "--recursive")

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
	assert.NoError(t, reader.GetError())
	reader.Reset()

	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep1())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Search artifacts without c=3
	searchCmd.SetSpec(searchSpecBuilder.Props("").ExcludeProps("c=3").BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	assert.NoError(t, reader.GetError())
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep2())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Search artifacts without a=1&b=2
	searchCmd.SetSpec(searchSpecBuilder.Props("").ExcludeProps("a=1;b=2").BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	assert.NoError(t, reader.GetError())
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep3())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Search artifacts without a=1&b=2 and with c=3
	searchCmd.SetSpec(searchSpecBuilder.Props("c=3").ExcludeProps("a=1;b=2").BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	assert.NoError(t, reader.GetError())
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep4())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Search artifacts without a=1 and with c=5
	searchCmd.SetSpec(searchSpecBuilder.Props("c=5").ExcludeProps("a=1").BuildSpec())
	reader, err = searchCmd.Search()
	assert.NoError(t, err)

	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.NoError(t, assertDateInSearchResult(*resultItem))
	}
	assert.NoError(t, reader.GetError())
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep5())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

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
	assert.NoError(t, reader.GetError())
	reader.Reset()

	resultItems = []utils.SearchResult{}
	readerNoDate, err = utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for resultItem := new(utils.SearchResult); readerNoDate.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		resultItems = append(resultItems, *resultItem)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchPropsStep6())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Cleanup
	cleanArtifactoryTest()
}

// Remove not to be deleted dirs from delete command from path to delete.
func TestArtifactoryDeleteExcludeProps(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadWithPropsSpecdeleteExcludeProps)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile, "--recursive")

	// Prepare search command
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.RtRepo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)

	// Delete all artifacts without c=1 but keep dirs that has at least one artifact with c=1 props
	artifactoryCli.Exec("delete", tests.RtRepo1+"/*", "--exclude-props=c=1")

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
	assert.NoError(t, readerNoDate.GetError())
	assert.ElementsMatch(t, resultItems, tests.GetSearchAfterDeleteWithExcludeProps())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())

	// Cleanup
	cleanArtifactoryTest()
}

func getAllBuildsByBuildName(client *httpclient.HttpClient, buildName string, t *testing.T, expectedHttpStatusCode int) buildsApiResponseStruct {
	resp, body, _, _ := client.SendGet(serverDetails.ArtifactoryUrl+"api/build/"+buildName, true, artHttpDetails, "")
	assert.Equal(t, expectedHttpStatusCode, resp.StatusCode, "Failed retrieving build information from artifactory.")

	buildsApiResponse := &buildsApiResponseStruct{}
	err := json.Unmarshal(body, buildsApiResponse)
	assert.NoError(t, err, "Unmarshaling failed with an error")
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

	content := buffer.Bytes()
	buffer.Reset()
	logger.Output(string(content))

	status, err := jsonparser.GetString(content, "status")
	assert.NoError(t, err)
	assert.Equal(t, expected.status, status, "Summary validation failed")

	resultSuccess, err := jsonparser.GetInt(content, "totals", "success")
	assert.NoError(t, err)

	resultFailure, err := jsonparser.GetInt(content, "totals", "failure")
	assert.NoError(t, err)

	assert.Equal(t, expected.success, resultSuccess, "Summary validation failed")
	assert.Equal(t, expected.failure, resultFailure, "Summary validation failed")
}

func CleanArtifactoryTests() {
	cleanArtifactoryTest()
	deleteCreatedRepos()
}

func initArtifactoryTest(t *testing.T) {
	if !*tests.TestArtifactory {
		t.Skip("Skipping artifactory test. To run artifactory test add the '-test.artifactory=true' option.")
	}
}

func initArtifactoryProjectTest(t *testing.T) {
	if !*tests.TestArtifactoryProject {
		t.Skip("Skipping artifactory project test. To run artifactory test add the '-test.artifactoryProject=true' option.")
	}
}

func cleanArtifactoryTest() {
	if !*tests.TestArtifactory {
		return
	}
	log.Info("Cleaning test data...")
	cleanArtifactory()
	tests.CleanFileSystem()
}

func preUploadBasicTestResources() {
	uploadPath := tests.GetTestResourcesPath() + "a/(.*)"
	targetPath := tests.RtRepo1 + "/test_resources/{1}"
	artifactoryCli.Exec("upload", uploadPath, targetPath,
		"--threads=10", "--regexp=true", "--target-props=searchMe=true", "--flat=false")
}

func execDeleteRepo(repoName string) {
	err := artifactoryCli.Exec("repo-delete", repoName, "--quiet")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func execDeleteUser(username string) {
	err := artifactoryCli.Exec("users-delete", username, "--quiet")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func getAllRepos() (repositoryKeys []string, err error) {
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, false)
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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body))
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
	content, err := ioutil.ReadFile(repoConfig)
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
	resp, body, err := client.SendPut(serverDetails.ArtifactoryUrl+"api/repositories/"+repoName, content, artHttpDetails, "")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Error(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
		os.Exit(1)
	}
	log.Info("Repository", repoName, "created.")
}

func getAllUsernames() (usersnames []string, err error) {
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, false)
	if err != nil {
		return nil, err
	}
	users, err := servicesManager.GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		usersnames = append(usersnames, user.Name)
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
		log.Info("Build", buildName, "deleted.")
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
	// Important - Virtual repositories most be deleted first
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
	fmt.Println(deleteSpecFile)
	deleteSpecFile, err := tests.ReplaceTemplateVariables(deleteSpecFile, "")
	fmt.Println(deleteSpecFile)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	deleteSpec, _ := spec.CreateSpecFromFile(deleteSpecFile, nil)
	tests.DeleteFiles(deleteSpec, serverDetails)
}

func searchInArtifactory(specFile string, t *testing.T) ([]utils.SearchResult, error) {
	searchSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for searchResult := new(utils.SearchResult); readerNoDate.NextRecord(searchResult) == nil; searchResult = new(utils.SearchResult) {
		resultItems = append(resultItems, *searchResult)
	}
	assert.NoError(t, readerNoDate.GetError())
	assert.NoError(t, readerNoDate.Close())
	assert.NoError(t, reader.Close())
	return resultItems, err
}

func getSpecAndCommonFlags(specFile string) (*spec.SpecFiles, rtutils.CommonConf) {
	searchFlags, _ := rtutils.NewCommonConfImpl(artAuth)
	searchSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	return searchSpec, searchFlags
}

func verifyExistInArtifactory(expected []string, specFile string, t *testing.T) {
	results, _ := searchInArtifactory(specFile, t)
	tests.CompareExpectedVsActual(expected, results, t)
}

func verifyDoesntExistInArtifactory(specFile string, t *testing.T) {
	verifyExistInArtifactory([]string{}, specFile, t)
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
	assert.NoError(t, readerNoDate.GetError())
	tests.CompareExpectedVsActual(expected, resultItems, t)
	assert.NoError(t, reader.Close())
	assert.NoError(t, readerNoDate.Close())
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
	assert.NoError(t, err, "Failed to get current dir.")
	dotGitExists, err := fileutils.IsDirExists(filepath.Join(dotGitPath, ".git"), false)
	assert.NoError(t, err)
	assert.True(t, dotGitExists, "Can't find .git")
	return dotGitPath
}

func deleteServerConfig() {
	configCli.WithoutCredentials().Exec("rm", tests.ServerId, "--quiet")
}

// This function will create server config and return the entire passphrase flag if it needed.
// For example if passphrase is needed it will return "--ssh-passphrase=${theConfiguredPassphrase}" or empty string.
func createServerConfigAndReturnPassphrase() (passphrase string, err error) {
	deleteServerConfig()
	if *tests.JfrogSshPassphrase != "" {
		passphrase = "--ssh-passphrase=" + *tests.JfrogSshPassphrase
	}
	return passphrase, configCli.Exec("add", tests.ServerId)
}

func testCopyMoveNoSpec(command string, beforeCommandExpected, afterCommandExpected []string, t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Run command with dry-run
	artifactoryCli.Exec(command, tests.RtRepo1+"/data/*a*", tests.RtRepo2+"/", "--dry-run")

	// Validate files weren't affected
	cpMvSpecFilePath, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(beforeCommandExpected, cpMvSpecFilePath, t)

	// Run command
	artifactoryCli.Exec(command, tests.RtRepo1+"/data/*a*", tests.RtRepo2+"/")

	// Validate files were affected
	verifyExistInArtifactory(afterCommandExpected, cpMvSpecFilePath, t)

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
		assert.NoError(t, reader.GetError())
		assert.NoError(t, reader.Close())
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
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "testdata/a/../a/a1.*", tests.RtRepo1, "--flat=true")
	artifactoryCli.Exec("upload", "testdata/./a/a1.*", tests.RtRepo1, "--flat=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, t)

	artifactoryCli.Exec("upload", "testdata/./a/../a/././././a2.*", tests.RtRepo1, "--flat=true")
	searchFilePath, err = tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpected2filesRepo1(), searchFilePath, t)
	if coreutils.IsWindows() {
		artifactoryCli.Exec("upload", `testdata\\a\\..\\a\\a1.*`, tests.RtRepo2, "--flat=true")
		artifactoryCli.Exec("upload", `testdata\\.\\\a\a1.*`, tests.RtRepo2, "--flat=true")
		searchFilePath, err = tests.CreateSpec(tests.SearchRepo2)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo2(), searchFilePath, t)

		artifactoryCli.Exec("upload", `testdata\\.\\a\\..\\a\\.\\.\\.\\.\\a2.*`, tests.RtRepo2, "--flat=true")
		searchFilePath, err = tests.CreateSpec(tests.SearchRepo2)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpected2filesRepo2(), searchFilePath, t)
	}
	cleanArtifactoryTest()
}

func TestGetExtractorsRemoteDetails(t *testing.T) {
	initArtifactoryTest(t)
	_, err := createServerConfigAndReturnPassphrase()
	assert.NoError(t, err)
	defer deleteServerConfig()

	// Make sure extractor1.jar downloaded from oss.jfrog.org.
	downloadPath := "org/jfrog/buildinfo/build-info-extractor/extractor1.jar"
	expectedRemotePath := path.Join("oss-release-local", downloadPath)
	validateExtractorRemoteDetails(t, downloadPath, expectedRemotePath)

	// Make sure extractor2.jar also downloaded from oss.jfrog.org.
	downloadPath = "org/jfrog/buildinfo/build-info-extractor/extractor2.jar"
	expectedRemotePath = path.Join("oss-release-local", downloadPath)
	validateExtractorRemoteDetails(t, downloadPath, expectedRemotePath)

	// Set 'JFROG_CLI_EXTRACTORS_REMOTE' and make sure extractor3.jar downloaded from a remote repo 'test-remote-repo' in ServerId.
	testRemoteRepo := "test-remote-repo"
	assert.NoError(t, os.Setenv(utils.ExtractorsRemoteEnv, tests.ServerId+"/"+testRemoteRepo))
	defer func() { assert.NoError(t, os.Unsetenv(utils.ExtractorsRemoteEnv)) }()
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
	initArtifactoryTest(t)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RtBuildName1, artHttpDetails)
	testDir := initVcsTestDir(t)
	artifactoryCli.Exec("upload", filepath.Join(testDir, "*"), tests.RtRepo1, "--flat=false", "--build-name="+tests.RtBuildName1, "--build-number=2020")
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
	path, err := filepath.Abs(tests.Temp)
	assert.NoError(t, err)
	return path
}

func TestConfigAddOverwrite(t *testing.T) {
	initArtifactoryTest(t)
	// Add a new instance.
	err := tests.NewJfrogCli(execMain, "jfrog config", "").Exec("add", tests.ServerId, "--artifactory-url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint, "--user=admin", "--password=password", "--enc-password=false")
	// Remove the instance at the end of the test.
	defer tests.NewJfrogCli(execMain, "jfrog config", "").Exec("rm", tests.ServerId, "--quiet")
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
	initArtifactoryTest(t)
	// Configure server with dummy credentials
	err := tests.NewJfrogCli(execMain, "jfrog config", "").Exec("add", tests.ServerId, "--artifactory-url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint, "--user=admin", "--password=password", "--enc-password=false")
	defer deleteServerConfig()
	assert.NoError(t, err)

	// Init tmp dir
	specFile, err := tests.CreateSpec(tests.ReplicationTempCreate)
	assert.NoError(t, err)

	// Create push replication
	err = artifactoryCli.Exec("rplc", specFile)
	assert.NoError(t, err)

	// Validate create replication
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, false)
	assert.NoError(t, err)
	result, err := servicesManager.GetReplication(tests.RtRepo1)
	assert.NoError(t, err)
	// The Replicator may encrypt the password internally, therefore we should only check that the password is not empty
	assert.NotEmpty(t, result[0].Password)
	result[0].Password = ""
	assert.ElementsMatch(t, result, tests.GetReplicationConfig())

	// Delete replication
	err = artifactoryCli.Exec("rpldel", tests.RtRepo1)
	assert.NoError(t, err)

	// Validate delete replication
	result, err = servicesManager.GetReplication(tests.RtRepo1)
	assert.Error(t, err)
	// Cleanup
	cleanArtifactoryTest()
}

func TestAccessTokenCreate(t *testing.T) {
	initArtifactoryTest(t)

	buffer, previousLog := tests.RedirectLogOutputToBuffer()
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
		err := artifactoryCli.Exec("atc")
		assert.NoError(t, err)
	}

	// Check access token
	checkAccessToken(t, buffer)

	// Create access token for current user, explicitly
	err := artifactoryCli.Exec("atc", *tests.JfrogUser)
	assert.NoError(t, err)

	// Check access token
	checkAccessToken(t, buffer)

	// Cleanup
	cleanArtifactoryTest()
}

func checkAccessToken(t *testing.T, buffer *bytes.Buffer) {
	// Write the command output to the origin
	content := buffer.Bytes()
	buffer.Reset()

	// Extract the the token from the output
	token, err := jsonparser.GetString(content, "access_token")
	assert.NoError(t, err)

	// Try ping with the new token
	err = tests.NewJfrogCli(execMain, "jfrog rt", "--url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint+" --access-token="+token).Exec("ping")
	assert.NoError(t, err)
}

func TestRefreshableTokens(t *testing.T) {
	initArtifactoryTest(t)

	if *tests.JfrogAccessToken != "" {
		t.Skip("Test only with username and password , skipping...")
	}

	// Create server with initialized refreshable tokens.
	_, err := createServerConfigAndReturnPassphrase()
	defer deleteServerConfig()
	assert.NoError(t, err)

	// Upload a file and assert the refreshable tokens were generated.
	artifactoryCommandExecutor := tests.NewJfrogCli(execMain, "jfrog rt", "")
	uploadedFiles := 1
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, tests.ServerId, "testdata/a/a1.in", uploadedFiles)
	if err != nil {
		return
	}
	curAccessToken, curRefreshToken, err := getTokensFromConfig(t, tests.ServerId)
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
	err = setRefreshTokenInConfig(t, tests.ServerId, "invalid-token")
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
	newAccessToken, newRefreshToken, err := getTokensFromConfig(t, tests.ServerId)
	if err != nil {
		return
	}
	assert.Equal(t, curAccessToken, newAccessToken)
	assert.Equal(t, curRefreshToken, newRefreshToken)

	// Cleanup
	cleanArtifactoryTest()
}

func setRefreshTokenInConfig(t *testing.T, serverId, token string) error {
	details, err := config.GetAllServersConfigs()
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	for _, server := range details {
		if server.ServerId == serverId {
			server.SetRefreshToken(token)
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

func getTokensFromConfig(t *testing.T, serverId string) (accessToken, refreshToken string, err error) {
	details, err := config.GetSpecificConfig(serverId, false, false)
	if err != nil {
		assert.NoError(t, err)
		return "", "", err
	}
	return details.AccessToken, details.RefreshToken, nil
}

func assertTokensChanged(t *testing.T, serverId, curAccessToken, curRefreshToken string) (newAccessToken, newRefreshToken string, err error) {
	newAccessToken, newRefreshToken, err = getTokensFromConfig(t, serverId)
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
	initArtifactoryTest(t)

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
	err := artifactoryCli.Exec("upload", filePath, tests.RtRepo1, "--ant", "--regexp=false", "--flat=true")
	assert.NoError(t, err)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleAntPatternUploadExpectedRepo1(), searchFilePath, t)
}

func simpleUploadWithAntPatternSpec(t *testing.T) {
	// Init tmp dir
	specFile, err := tests.CreateSpec(tests.UploadAntPattern)
	assert.NoError(t, err)
	err = fileutils.CopyDir(tests.GetTestResourcesPath()+"cache", filepath.Dir(specFile), true, nil)
	assert.NoError(t, err)
	// Upload
	artifactoryCli.Exec("upload", "--spec="+specFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleAntPatternUploadExpectedRepo1(), searchFilePath, t)
	searchFilePath, err = tests.CreateSpec(tests.SearchRepo1NonExistFile)
	verifyDoesntExistInArtifactory(searchFilePath, t)
}

func uploadUsingAntAIncludeDirsAndFlat(t *testing.T) {
	filePath := "testdata/*/empt?/**"
	err := artifactoryCli.Exec("upload", filePath, tests.RtRepo1, "--ant", "--include-dirs=true", "--flat=true")
	assert.NoError(t, err)
	err = artifactoryCli.Exec("upload", filePath, tests.RtRepo1, "--ant", "--include-dirs=true", "--flat=false")
	assert.NoError(t, err)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1IncludeDirs)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetAntPatternUploadWithIncludeDirsExpectedRepo1(), searchFilePath, t)
}

func TestUploadWithAntPatternAndExclusionsSpec(t *testing.T) {
	initArtifactoryTest(t)
	// Init tmp dir
	specFile, err := tests.CreateSpec(tests.UploadAntPatternExclusions)
	assert.NoError(t, err)
	err = fileutils.CopyDir(tests.GetTestResourcesPath(), filepath.Dir(specFile), true, nil)
	assert.NoError(t, err)
	// Upload
	artifactoryCli.Exec("upload", "--spec="+specFile)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetAntPatternUploadWithExclusionsExpectedRepo1(), searchFilePath, t)
	searchFilePath, err = tests.CreateSpec(tests.SearchRepo1NonExistFileAntExclusions)
	verifyDoesntExistInArtifactory(searchFilePath, t)
	cleanArtifactoryTest()
}

func TestPermissionTargets(t *testing.T) {
	initArtifactoryTest(t)
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, false)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	templatePath := filepath.Join(tests.GetTestResourcesPath(), "permissiontarget", "template")

	// Create permission target on specific repo.
	assert.NoError(t, artifactoryCli.Exec("ptc", templatePath, createPermissionTargetsTemplateVars(tests.RtRepo1)))
	assertPermissionTarget(t, servicesManager, tests.RtRepo1)

	// Update permission target to ANY repo.
	any := "ANY"
	assert.NoError(t, artifactoryCli.Exec("ptu", templatePath, createPermissionTargetsTemplateVars(any)))
	assertPermissionTarget(t, servicesManager, any)

	// Delete permission target.
	assert.NoError(t, artifactoryCli.Exec("ptdel", tests.RtPermissionTargetName))
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
	expected := tests.GetExpectedPermissionTarget(repoValue)
	assert.EqualValues(t, expected, *actual)
}

func assertPermissionTargetDeleted(t *testing.T, manager artifactory.ArtifactoryServicesManager) {
	permission, err := manager.GetPermissionTarget(tests.RtPermissionTargetName)
	assert.NoError(t, err)
	assert.Nil(t, permission)
}

func cleanPermissionTarget() {
	_ = artifactoryCli.Exec("ptdel", tests.RtPermissionTargetName)
}

func TestArtifactoryCurl(t *testing.T) {
	initArtifactoryTest(t)
	_, err := createServerConfigAndReturnPassphrase()
	defer deleteServerConfig()
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
