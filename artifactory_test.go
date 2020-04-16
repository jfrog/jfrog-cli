package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/buger/jsonparser"
	"github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	logUtils "github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-cli/utils/tests"
	cliproxy "github.com/jfrog/jfrog-cli/utils/tests/proxy/server"
	"github.com/jfrog/jfrog-cli/utils/tests/proxy/server/certificate"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	rtutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils/tests/xray"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
)

// JFrog CLI for Artifactory commands
var artifactoryCli *tests.JfrogCli

// JFrog CLI for config command only (doesn't pass the --ssh-passphrase flag)
var configArtifactoryCli *tests.JfrogCli

var artifactoryDetails *config.ArtifactoryDetails
var artAuth auth.ServiceDetails
var artHttpDetails httputils.HttpClientDetails

func InitArtifactoryTests() {
	initArtifactoryCli()
	createReposIfNeeded()
	cleanArtifactoryTest()
}

func authenticate() string {
	artifactoryDetails = &config.ArtifactoryDetails{Url: clientutils.AddTrailingSlashIfNeeded(*tests.RtUrl), SshKeyPath: *tests.RtSshKeyPath, SshPassphrase: *tests.RtSshPassphrase, AccessToken: *tests.RtAccessToken}
	cred := "--url=" + *tests.RtUrl
	if *tests.RtDistributionUrl != "" {
		cred += " --dist-url=" + *tests.RtDistributionUrl
	}
	if !fileutils.IsSshUrl(artifactoryDetails.Url) {
		if *tests.RtApiKey != "" {
			artifactoryDetails.ApiKey = *tests.RtApiKey
		} else if *tests.RtAccessToken != "" {
			artifactoryDetails.AccessToken = *tests.RtAccessToken
		} else {
			artifactoryDetails.User = *tests.RtUser
			artifactoryDetails.Password = *tests.RtPassword
		}
	}
	cred += getArtifactoryTestCredentials()
	var err error
	if artAuth, err = artifactoryDetails.CreateArtAuthConfig(); err != nil {
		cliutils.ExitOnErr(errors.New("Failed while attempting to authenticate with Artifactory: " + err.Error()))
	}
	artifactoryDetails.SshAuthHeaders = artAuth.GetSshAuthHeaders()
	artifactoryDetails.Url = artAuth.GetUrl()
	artifactoryDetails.SshUrl = artAuth.GetSshUrl()
	artHttpDetails = artAuth.CreateHttpClientDetails()
	return cred
}

// A Jfrog CLI to be used to execute a config task.
// Removed the ssh-passphrase flag that cannot be passed to with a config command
func createConfigJfrogCLI(cred string) *tests.JfrogCli {
	if strings.Contains(cred, " --ssh-passphrase=") {
		cred = strings.Replace(cred, " --ssh-passphrase="+*tests.RtSshPassphrase, "", -1)
	}
	return tests.NewJfrogCli(execMain, "jfrog rt", cred)
}

func getArtifactoryTestCredentials() string {
	if fileutils.IsSshUrl(artifactoryDetails.Url) {
		return getSshCredentials()
	}
	if *tests.RtApiKey != "" {
		return " --apikey=" + *tests.RtApiKey
	}
	if *tests.RtAccessToken != "" {
		return " --access-token=" + *tests.RtAccessToken
	}
	return " --user=" + *tests.RtUser + " --password=" + *tests.RtPassword
}

func getSshCredentials() string {
	cred := ""
	if *tests.RtSshKeyPath != "" {
		cred += " --ssh-key-path=" + *tests.RtSshKeyPath
	}
	if *tests.RtSshPassphrase != "" {
		cred += " --ssh-passphrase=" + *tests.RtSshPassphrase
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
	err = fileutils.CopyDir(tests.GetTestResourcesPath()+"cache", filepath.Dir(specFile), true)
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
	passphrase := createServerConfigAndReturnPassphrase()
	artifactoryCommandExecutor := tests.NewJfrogCli(execMain, "jfrog rt", "")
	specFile, err := tests.CreateSpec(tests.UploadFlatRecursive)
	assert.NoError(t, err)
	artifactoryCommandExecutor.Exec("upload", "--spec="+specFile, "--server-id="+tests.RtServerId, passphrase)

	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleUploadExpectedRepo1(), searchFilePath, t)
	deleteServerConfig()
	cleanArtifactoryTest()
}

func TestArtifactoryUploadPathWithSpecialCharsAsNoRegex(t *testing.T) {
	initArtifactoryTest(t)
	filePath := getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadFromVirtual(t *testing.T) {
	initArtifactoryTest(t)

	artifactoryCli.Exec("upload", "testsdata/a/*", tests.Repo1, "--flat=false")
	artifactoryCli.Exec("dl", tests.VirtualRepo+"/testsdata/(*)", tests.Out+"/"+"{1}", "--flat=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally(tests.GetVirtualDownloadExpected(), paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadPathWithSpecialChars(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", getSpecialCharFilePath(), tests.Repo1, "--flat=false")
	artifactoryCli.Exec("upload", "testsdata/c#/a#1.in", tests.Repo1, "--flat=false")

	artifactoryCli.Exec("dl", tests.Repo1+"/testsdata/a$+~&^a#/a*", tests.Out+fileutils.GetFileSeparator(), "--flat=true")
	artifactoryCli.Exec("dl", tests.Repo1+"/testsdata/c#/a#1.in", tests.Out+fileutils.GetFileSeparator(), "--flat=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{tests.Out + fileutils.GetFileSeparator() + "a1.in", tests.Out + fileutils.GetFileSeparator() + "a#1.in"}, paths, t)

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
	randFile, err := io.CreateRandFile(filepath.Join(tests.Out, "randFile"), fileSize)
	assert.NoError(t, err)
	localFileDetails, err := fileutils.GetFileDetails(randFile.Name())
	assert.NoError(t, err)

	artifactoryCli.Exec("u", randFile.Name(), tests.Repo1+"/testsdata/", "--flat=true")
	randFile.File.Close()
	os.RemoveAll(tests.Out)
	artifactoryCli.Exec("dl", tests.Repo1+"/testsdata/", tests.Out+fileutils.GetFileSeparator(), "--flat=true")

	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally([]string{tests.Out + fileutils.GetFileSeparator() + "randFile"}, paths, t)
	tests.ValidateChecksums(tests.Out+fileutils.GetFileSeparator()+"randFile", localFileDetails.Checksum, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadWildcardInRepo(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	// Upload a file to repo1 and another one to repo2
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/a1.in")
	artifactoryCli.Exec("upload", filePath, tests.Repo2+"/path/a2.in")

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

func TestArtifactoryCopySingleFileNonFlat(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/")
	artifactoryCli.Exec("cp", tests.Repo1+"/path/a1.in", tests.Repo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestAqlFindingItemOnRoot(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/")
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/*", tests.Repo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetAnyItemCopy(), searchPath, t)
	artifactoryCli.Exec("del", tests.Repo2+"/*")
	artifactoryCli.Exec("del", tests.Repo1+"/*")
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/")
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/*/", tests.Repo2)
	verifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestExitCode(t *testing.T) {
	initArtifactoryTest(t)

	err := artifactoryCli.Exec("upload", "DummyText", tests.Repo1, "--fail-no-op=true")
	checkExitCode(t, cliutils.ExitCodeError, err)
	err = artifactoryCli.Exec("upload", path.Join("testsdata", "a", "a1.in"), "tests.Repo1")
	checkExitCode(t, cliutils.ExitCodeError, err)
	err = artifactoryCli.Exec("upload", "testsdata/a/(*.dummyExt)", tests.Repo1+"/{1}.in", "--fail-no-op=true")
	checkExitCode(t, cliutils.ExitCodeFailNoOp, err)

	err = artifactoryCli.Exec("dl", "DummyFolder", "--fail-no-op=true")
	checkExitCode(t, cliutils.ExitCodeFailNoOp, err)

	//upload dummy file inorder to test move & copy
	artifactoryCli.Exec("upload", path.Join("testsdata", "a", "a1.in"), tests.Repo1)
	err = artifactoryCli.Exec("move", tests.Repo1, "DummyTargetPath")
	checkExitCode(t, cliutils.ExitCodeError, err)
	err = artifactoryCli.Exec("move", "DummyText", tests.Repo1, "--fail-no-op=true")
	checkExitCode(t, cliutils.ExitCodeFailNoOp, err)

	err = artifactoryCli.Exec("copy", tests.Repo1, "DummyTargetPath")
	checkExitCode(t, cliutils.ExitCodeError, err)
	err = artifactoryCli.Exec("copy", "DummyText", tests.Repo1, "--fail-no-op=true")
	checkExitCode(t, cliutils.ExitCodeFailNoOp, err)

	err = artifactoryCli.Exec("delete", "DummyText", "--fail-no-op=true")
	checkExitCode(t, cliutils.ExitCodeFailNoOp, err)

	err = artifactoryCli.Exec("s", "DummyText", "--fail-no-op=true")
	checkExitCode(t, cliutils.ExitCodeFailNoOp, err)

	err = artifactoryCli.Exec("sp", "DummyText", "prop=val;key=value", "--fail-no-op=true")
	checkExitCode(t, cliutils.ExitCodeFailNoOp, err)

	err = artifactoryCli.Exec("delp", "DummyText", "prop=val;key=value", "--fail-no-op=true")
	checkExitCode(t, cliutils.ExitCodeFailNoOp, err)

	cleanArtifactoryTest()
}

func checkExitCode(t *testing.T, expected cliutils.ExitCode, er error) {
	switch underlyingType := er.(type) {
	case cliutils.CliError:
		assert.Equal(t, expected, underlyingType.ExitCode)
	default:
		assert.Fail(t, "Exit code expected error code %v, got %v", expected.Code, er)
	}
}
func TestArtifactoryDirectoryCopy(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/")
	artifactoryCli.Exec("cp", tests.Repo1+"/path/", tests.Repo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyUsingWildcard(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/")
	artifactoryCli.Exec("cp", tests.Repo1+"/*/", tests.Repo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSingleFileCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyFilesNameWithParentheses(t *testing.T) {
	initArtifactoryTest(t)

	artifactoryCli.Exec("upload", "testsdata/b/*", tests.Repo1, "--flat=false")
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/(/(.in", tests.Repo2)
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/(b/(b.in", tests.Repo2)
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/b(/b(.in", tests.Repo2)
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/b)/b).in", tests.Repo2)
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/(b)/(b).in", tests.Repo2)
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/)b/)b.in", tests.Repo2)
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/)b)/)b).in", tests.Repo2)
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/(b/(b.in", tests.Repo2+"/()/", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/(b)/(b).in", tests.Repo2+"/()/")
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/b(/b(.in", tests.Repo2+"/(/", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/(/(*.in)", tests.Repo2+"/c/{1}.zip", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/(/(*.in)", tests.Repo2+"/(/{1}.zip")
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/b(/(b*.in)", tests.Repo2+"/(/{1}-up", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/testsdata/b/b(/(*).(*)", tests.Repo2+"/(/{2}-{1}", "--flat=true")

	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetCopyFileNameWithParentheses(), searchPath, t)

	cleanArtifactoryTest()
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

	artifactoryCli.Exec("upload", "testsdata/b/*", tests.Repo1, "--flat=false")
	artifactoryCli.Exec("download", path.Join(tests.Repo1), tests.Out+"/")

	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(tests.GetFileWithParenthesesDownload(), paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyUsingWildcardFlat(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/")
	artifactoryCli.Exec("cp", tests.Repo1+"/path/inner", tests.Repo2, "--flat=true")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetSingleDirectoryCopyFlat(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyPathsTwice(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/")

	log.Info("Copy Folder to root twice")
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetSingleFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2)
	verifyExistInArtifactory(tests.GetSingleFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("del", tests.Repo2)

	log.Info("Copy to from repo1/path to repo2/path twice")
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2+"/path")
	verifyExistInArtifactory(tests.GetSingleFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2+"/path")
	verifyExistInArtifactory(tests.GetFolderCopyTwice(), searchPath, t)
	artifactoryCli.Exec("del", tests.Repo2)

	log.Info("Copy to from repo1/path/ to repo2/path/ twice")
	artifactoryCli.Exec("cp", tests.Repo1+"/path/", tests.Repo2+"/path/")
	verifyExistInArtifactory(tests.GetSingleInnerFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("cp", tests.Repo1+"/path/", tests.Repo2+"/path/")
	verifyExistInArtifactory(tests.GetSingleInnerFileCopyFullPath(), searchPath, t)
	artifactoryCli.Exec("del", tests.Repo2)

	log.Info("Copy to from repo1/path/ to repo2/path/ twice")
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2+"/path/")
	verifyExistInArtifactory(tests.GetFolderCopyIntoFolder(), searchPath, t)
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2+"/path/")
	verifyExistInArtifactory(tests.GetFolderCopyIntoFolder(), searchPath, t)
	artifactoryCli.Exec("del", tests.Repo2)

	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyPatternEndsWithSlash(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/")
	artifactoryCli.Exec("cp", tests.Repo1+"/path/", tests.Repo2, "--flat=true")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetAnyItemCopyUsingSpec(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemUsingWildcardFlat(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/")
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/*", tests.Repo2)
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetAnyItemCopy(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemRecursive(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/a/b/")
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/aFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/a*", tests.Repo2, "--recursive=true")
	searchPath, err := tests.CreateSpec(tests.SearchRepo2)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetAnyItemCopyRecursive(), searchPath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAndRenameFolder(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = getSpecialCharFilePath()

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/")
	artifactoryCli.Exec("cp", tests.Repo1+"/*", tests.Repo2+"/newPath")
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
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/")
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", "--spec="+specFile)
	verifyExistInArtifactory(tests.GetAnyItemCopyUsingSpec(), searchPath, t)
	cleanArtifactoryTest()
}

func getSpecialCharFilePath() string {
	return "testsdata/a$+~&^a#/a*"
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
	artifactoryCli.Exec("cp", tests.Repo1+"/data/ "+tests.Repo2+"/", "--exclude-patterns=*b*;*c*")

	// Validate files are moved by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildCopyExclude(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryCopyExcludeBySpec(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Copy by spec
	specFile, err := tests.CreateSpec(tests.MoveCopySpecExclude)
	assert.NoError(t, err)
	artifactoryCli.Exec("cp", "--spec="+specFile)
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	// Validate files are moved by build number
	verifyExistInArtifactory(tests.GetBuildCopyExclude(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadDebian(t *testing.T) {
	initArtifactoryTest(t)
	specFile, err := tests.CreateSpec(tests.DebianUploadSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile, "--deb=bionic/main/i386")
	verifyExistInArtifactoryByProps(tests.GetUploadDebianExpected(), tests.DebianRepo+"/*", "deb.distribution=bionic;deb.component=main;deb.architecture=i386", t)
	artifactoryCli.Exec("upload", "--spec="+specFile, "--deb=cosmic/main\\/18.10/amd64")
	verifyExistInArtifactoryByProps(tests.GetUploadDebianExpected(), tests.DebianRepo+"/*", "deb.distribution=cosmic;deb.component=main/18.10;deb.architecture=amd64", t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadAndExplode(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", filepath.Join("testsdata", "archives", "a.zip"), tests.Repo1, "--explode=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetExplodeUploadExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadAndSyncDelete(t *testing.T) {
	initArtifactoryTest(t)
	// Upload all testdata/a/
	artifactoryCli.Exec("upload", path.Join("testsdata", "a", "*"), tests.Repo1+"/syncDir/", "--flat=false")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep1(), searchFilePath, t)
	// Upload testdata/a/b/*1.in and sync syncDir/testdata/a/b/
	artifactoryCli.Exec("upload", path.Join("testsdata", "a", "b", "*1.in"), tests.Repo1+"/syncDir/", "--sync-deletes="+tests.Repo1+"/syncDir/testsdata/a/b/", "--flat=false")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep2(), searchFilePath, t)
	// Upload testdata/archives/* and sync syncDir/
	artifactoryCli.Exec("upload", path.Join("testsdata", "archives", "*"), tests.Repo1+"/syncDir/", "--sync-deletes="+tests.Repo1+"/syncDir/")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep3(), searchFilePath, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndExplode(t *testing.T) {
	initArtifactoryTest(t)
	err := fileutils.CreateDirIfNotExist(tests.Out)
	assert.NoError(t, err)
	randFile, err := io.CreateRandFile(filepath.Join(tests.Out, "randFile"), 100000)
	assert.NoError(t, err)

	err = archiver.TarGz.Make(filepath.Join(tests.Out, "concurrent.tar.gz"), []string{randFile.Name()})
	assert.NoError(t, err)
	err = archiver.Tar.Make(filepath.Join(tests.Out, "bulk.tar"), []string{randFile.Name()})
	assert.NoError(t, err)
	err = archiver.Zip.Make(filepath.Join(tests.Out, "zipFile.zip"), []string{randFile.Name()})
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/*", tests.Repo1, "--flat=true")
	randFile.File.Close()
	os.RemoveAll(tests.Out)
	artifactoryCli.Exec("download", path.Join(tests.Repo1, "randFile"), tests.Out+"/", "--explode=true")
	artifactoryCli.Exec("download", path.Join(tests.Repo1, "concurrent.tar.gz"), tests.Out+"/", "--explode=false", "--min-split=50")
	artifactoryCli.Exec("download", path.Join(tests.Repo1, "bulk.tar"), tests.Out+"/", "--explode=true")
	artifactoryCli.Exec("download", path.Join(tests.Repo1, "zipFile.zip"), tests.Out+"/", "--explode=true", "--min-split=50")
	artifactoryCli.Exec("download", path.Join(tests.Repo1, "zipFile.zip"), tests.Out+"/", "--explode=true")

	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(tests.GetExtractedDownload(), paths, t)

	cleanArtifactoryTest()
}

func TestArtifactoryDownloadAndSyncDeletes(t *testing.T) {
	initArtifactoryTest(t)

	outDirPath := tests.Out + string(os.PathSeparator)
	// Upload all testdata/a/ to repo1/syncDir/
	artifactoryCli.Exec("upload", path.Join("testsdata", "a", "*"), tests.Repo1+"/syncDir/", "--flat=false")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetUploadExpectedRepo1SyncDeleteStep1(), searchFilePath, t)

	// Download repo1/syncDir/ to out/
	artifactoryCli.Exec("download", tests.Repo1+"/syncDir/", tests.Out+"/")
	paths, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	tests.VerifyExistLocally(tests.GetExpectedSyncDeletesDownloadStep2(), paths, t)

	// Download repo1/syncDir/ to out/ with flat=true and sync out/
	artifactoryCli.Exec("download", tests.Repo1+"/syncDir/", outDirPath, "--flat=true", "--sync-deletes="+outDirPath)
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetExpectedSyncDeletesDownloadStep3(), paths, t)

	// Download all files ended with 2.in from repo1/syncDir/ to out/ and sync out/
	artifactoryCli.Exec("download", tests.Repo1+"/syncDir/*2.in", outDirPath, "--flat=true", "--sync-deletes="+outDirPath)
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetExpectedSyncDeletesDownloadStep4(), paths, t)

	// Download repo1/syncDir/ to out/, exclude the pattern "*c*.in" and sync out/
	artifactoryCli.Exec("download", tests.Repo1+"/syncDir/", outDirPath, "--sync-deletes="+outDirPath+"syncDir"+string(os.PathSeparator), "--exclude-patterns=syncDir/testsdata/*c*in")
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	assert.NoError(t, err)
	checkSyncedDirContent(tests.GetSyncExpectedDeletesDownloadStep5(), paths, t)

	// Delete all files from repo1/syncDir/
	artifactoryCli.Exec("delete", tests.Repo1+"/syncDir/")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyDoesntExistInArtifactory(searchFilePath, t)

	// Upload all testdata/archives/ to repo1/syncDir/
	artifactoryCli.Exec("upload", path.Join("testsdata", "archives", "*"), tests.Repo1+"/syncDir/", "--flat=false")
	searchFilePath, err = tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSyncExpectedDeletesDownloadStep6(), searchFilePath, t)

	// Download repo1/syncDir/ to out/ and sync out/
	artifactoryCli.Exec("download", tests.Repo1+"/syncDir/", outDirPath, "--sync-deletes="+outDirPath+"syncDir"+string(os.PathSeparator))
	paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
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
	for _, v := range actual {
		for i, r := range expected {
			if strings.HasPrefix(r, v) || v == r {
				break
			}
			if i == len(actual)-1 {
				return errors.New("Should not have : " + v)
			}
		}
	}
	return nil
}

// Test self-signed certificates with Artifactory. For the test, we set up a reverse proxy server.
func TestArtifactorySelfSignedCert(t *testing.T) {
	initArtifactoryTest(t)
	tempDirPath, err := ioutil.TempDir("", "jfrog.cli.test.")
	err = errorutils.CheckError(err)
	assert.NoError(t, err)
	defer os.RemoveAll(tempDirPath)
	os.Setenv(cliutils.HomeDir, tempDirPath)
	os.Setenv(tests.HttpsProxyEnvVar, "1024")
	go cliproxy.StartLocalReverseHttpProxy(artifactoryDetails.Url, false)

	// The two certificate files are created by the reverse proxy on startup in the current directory.
	defer os.Remove(certificate.KEY_FILE)
	defer os.Remove(certificate.CERT_FILE)
	// Let's wait for the reverse proxy to start up.
	err = checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", false)
	assert.NoError(t, err)

	fileSpec := spec.NewBuilder().Pattern(tests.Repo1 + "/*.zip").Recursive(true).BuildSpec()
	assert.NoError(t, err)
	parsedUrl, err := url.Parse(artifactoryDetails.Url)
	artifactoryDetails.Url = "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()

	// The server is using self-signed certificates
	// Without loading the certificates (or skipping verification) we expect all actions to fail due to error: "x509: certificate signed by unknown authority"
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(fileSpec)
	err = searchCmd.Search()
	_, isUrlErr := err.(*url.Error)
	assert.True(t, isUrlErr, "Expected a connection failure, since reverse proxy didn't load self-signed-certs. Connection however is successful", err)

	// Set insecureTls to true and run again. We expect the command to succeed.
	artifactoryDetails.InsecureTls = true
	searchCmd = generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(fileSpec)
	err = searchCmd.Search()
	assert.NoError(t, err)

	// Set insecureTls back to false.
	// Copy the server certificates to the CLI security dir and run again. We expect the command to succeed.
	artifactoryDetails.InsecureTls = false
	securityDirPath, err := cliutils.GetJfrogSecurityDir()
	assert.NoError(t, err)
	err = fileutils.CopyFile(securityDirPath, certificate.KEY_FILE)
	assert.NoError(t, err)
	err = fileutils.CopyFile(securityDirPath, certificate.CERT_FILE)
	assert.NoError(t, err)
	searchCmd = generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(fileSpec)
	err = searchCmd.Search()
	assert.NoError(t, err)

	artifactoryDetails.Url = artAuth.GetUrl()
	cleanArtifactoryTest()
}

// Test client certificates with Artifactory. For the test, we set up a reverse proxy server.
func TestArtifactoryClientCert(t *testing.T) {
	initArtifactoryTest(t)
	tempDirPath, err := ioutil.TempDir("", "jfrog.cli.test.")
	err = errorutils.CheckError(err)
	assert.NoError(t, err)
	defer os.RemoveAll(tempDirPath)
	os.Setenv(cliutils.HomeDir, tempDirPath)
	os.Setenv(tests.HttpsProxyEnvVar, "1025")
	go cliproxy.StartLocalReverseHttpProxy(artifactoryDetails.Url, true)

	// The two certificate files are created by the reverse proxy on startup in the current directory.
	defer os.Remove(certificate.KEY_FILE)
	defer os.Remove(certificate.CERT_FILE)
	// Let's wait for the reverse proxy to start up.
	err = checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", true)
	assert.NoError(t, err)

	fileSpec := spec.NewBuilder().Pattern(tests.Repo1 + "/*.zip").Recursive(true).BuildSpec()
	assert.NoError(t, err)
	parsedUrl, err := url.Parse(artifactoryDetails.Url)
	artifactoryDetails.Url = "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()
	artifactoryDetails.InsecureTls = true

	// The server is requiring client certificates
	// Without loading a valid client certificate, we expect all actions to fail due to error: "tls: bad certificate"
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(fileSpec)
	err = searchCmd.Search()
	_, isUrlErr := err.(*url.Error)
	assert.True(t, isUrlErr, "Expected a connection failure, since client did not provide a client certificate. Connection however is successful")

	// Inject client certificates, we expect the search to succeed
	artifactoryDetails.ClientCertPath = certificate.CERT_FILE
	artifactoryDetails.ClientCertKeyPath = certificate.KEY_FILE

	searchCmd = generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(fileSpec)
	err = searchCmd.Search()
	assert.NoError(t, err)

	artifactoryDetails.Url = artAuth.GetUrl()
	artifactoryDetails.InsecureTls = false
	artifactoryDetails.ClientCertPath = ""
	artifactoryDetails.ClientCertKeyPath = ""
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
	rtUrl, err := url.Parse(artifactoryDetails.Url)
	assert.NoError(t, err)
	var proxyTestArgs []string
	var httpProxyEnv string
	testArgs := []string{"-test.artifactoryProxy=true", "-rt.url=" + *tests.RtUrl, "-rt.user=" + *tests.RtUser, "-rt.password=" + *tests.RtPassword, "-rt.apikey=" + *tests.RtApiKey, "-rt.sshKeyPath=" + *tests.RtSshKeyPath, "-rt.sshPassphrase=" + *tests.RtSshPassphrase}
	if rtUrl.Scheme == "https" {
		os.Setenv(tests.HttpsProxyEnvVar, "1026")
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
	authenticate()
	proxyRtUrl := prepareArtifactoryUrlForProxyTest(t)
	spec := spec.NewBuilder().Pattern(tests.Repo1 + "/*.zip").Recursive(true).BuildSpec()
	artifactoryDetails.Url = proxyRtUrl
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
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(spec)
	err = searchCmd.Search()
	assert.NoError(t, err)
	artifactoryDetails.Url = artAuth.GetUrl()
}

func prepareArtifactoryUrlForProxyTest(t *testing.T) string {
	rtUrl, err := url.Parse(artifactoryDetails.Url)
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
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(spec)
	err := searchCmd.Search()
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

	for attempt := 0; attempt < 10; attempt++ {
		log.Info("Checking if proxy server is up and running.", strconv.Itoa(attempt+1), "attempt.", "URL:", proxyScheme+"://localhost:"+port)
		resp, err := client.Get(proxyScheme + "://localhost:" + port)
		if err != nil {
			attempt++
			time.Sleep(time.Second)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
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
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/a.in")
	// Set the 'prop=red' property to the file.
	artifactoryCli.Exec("sp", tests.Repo1+"/a.*", "prop=red")
	// Now let's change the property value, by searching for the 'prop=red'.
	specFile, err := tests.CreateSpec(tests.SetDeletePropsSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("sp", "prop=green", "--spec="+specFile)

	resultItems := searchItemsInArtifactory(t)
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

func TestArtifactorySetPropertiesExcludeByCli(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "testsdata/a/a*.in", tests.Repo1+"/")
	artifactoryCli.Exec("sp", tests.Repo1+"/*", "prop=val", "--exclude-patterns=*a1.in;*a2.in")
	resultItems := searchItemsInArtifactory(t)
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
	artifactoryCli.Exec("upload", "testsdata/a/a*.in", tests.Repo1+"/")
	artifactoryCli.Exec("sp", tests.Repo1+"/*", "prop=val", "--exclusions=*/*a1.in;*/*a2.in")
	resultItems := searchItemsInArtifactory(t)
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
	artifactoryCli.Exec("upload", "testsdata/a/a*.in", tests.Repo1+"/a/")
	artifactoryCli.Exec("sp", tests.Repo1+"/a/*", "color=yellow;prop=red;status=ok")
	// Delete the 'color' property.
	artifactoryCli.Exec("delp", tests.Repo1+"/a/*", "color")
	// Delete the 'status' property, by a spec which filters files by 'prop=red'.
	specFile, err := tests.CreateSpec(tests.SetDeletePropsSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("delp", "status", "--spec="+specFile)

	resultItems := searchItemsInArtifactory(t)
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
	artifactoryCli.Exec("upload", "testsdata/a/a*.in", tests.Repo1+"/")
	artifactoryCli.Exec("sp", tests.Repo1+"/*", "prop=val")

	artifactoryCli.Exec("delp", tests.Repo1+"/*", "prop", "--exclude-patterns=*a1.in;*a2.in")
	resultItems := searchItemsInArtifactory(t)
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
	artifactoryCli.Exec("upload", "testsdata/a/a*.in", tests.Repo1+"/")
	artifactoryCli.Exec("sp", tests.Repo1+"/*", "prop=val")

	artifactoryCli.Exec("delp", tests.Repo1+"/*", "prop", "--exclusions=*/*a1.in;*/*a2.in")
	resultItems := searchItemsInArtifactory(t)
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
	artifactoryCli.Exec("upload", testFileRel, tests.Repo1, "--recursive=false")
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
	artifactoryCli.Exec("upload", "testsdata/a/a*", tests.Repo1, "--exclusions=*a2*;*a3.in")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli1Regex(t *testing.T) {
	initArtifactoryTest(t)
	// Upload files
	artifactoryCli.Exec("upload", "testsdata/a/a(.*)", tests.Repo1, "--exclusions=(.*)a2.*;.*a3.in", "--regexp=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli2Wildcard(t *testing.T) {
	initArtifactoryTest(t)

	// Create temp dir
	absDirPath, err := ioutil.TempDir("", "cliTestDir")
	assert.NoError(t, err, "Couldn't create dir")
	defer os.Remove(absDirPath)

	// Create temp files
	d1 := []byte("test file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile1.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile2.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")

	// Upload files
	artifactoryCli.Exec("upload", filepath.ToSlash(absDirPath)+"/*", tests.Repo1, "--exclusions=*cliTestFile1*")

	// Check files exists in artifactory
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory([]string{tests.Repo1 + "/cliTestFile2.in"}, searchFilePath, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli2Regex(t *testing.T) {
	initArtifactoryTest(t)

	// Create temp dir
	absDirPath, err := ioutil.TempDir("", "cliTestDir")
	assert.NoError(t, err, "Couldn't create dir")
	defer os.Remove(absDirPath)

	// Create temp files
	d1 := []byte("test file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile1.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")
	err = ioutil.WriteFile(filepath.Join(absDirPath, "cliTestFile2.in"), d1, 0644)
	assert.NoError(t, err, "Couldn't create file")

	// Upload files
	artifactoryCli.Exec("upload", filepath.ToSlash(absDirPath)+"(.*)", tests.Repo1, "--exclusions=(.*c)liTestFile1.*", "--regexp=true")

	// Check files exists in artifactory
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)

	verifyExistInArtifactory([]string{tests.Repo1 + "/cliTestFile2.in"}, searchFilePath, t)

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
	artifactoryCli.Exec("upload", "testsdata/regexp"+"(.*)"+"\\."+".*", tests.Repo1, "--regexp=true")
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)

	verifyExistInArtifactory([]string{tests.Repo1 + "/has.dot"}, searchFilePath, t)
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
	if cliutils.IsWindows() {
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
	artifactoryCli.Exec("u", validLink+" "+tests.Repo1+" --symlinks=true")

	// Delete the local symlink
	err = os.Remove(validLink)
	assert.NoError(t, err)

	// Download symlink from artifactory
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/ --validate-symlinks=true")

	// Should be valid if successful
	validateSymLink(validLink, localFile, t)

	// Delete symlink and clean
	os.Remove(validLink)
	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded.
// validate the symlink content checksum.
func TestValidateBrokenSymlink(t *testing.T) {
	if cliutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)

	// Path to broken symLink
	brokenLink := filepath.Join(tests.GetTestResourcesPath()+"a/", "brokenLink")

	// Link broken symLink to non_existing_path
	err := os.Symlink("non-non_existing_path-path", brokenLink)
	assert.NoError(t, err)

	// Upload symlink to artifactory
	artifactoryCli.Exec("u", brokenLink+" "+tests.Repo1+" --symlinks=true")

	// Delete the local symlink
	err = os.Remove(brokenLink)
	assert.NoError(t, err)

	// Try downloading symlink from artifactory. Since the link is broken, it shouldn't be downloaded
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/ --validate-symlinks=true")
	if fileutils.IsPathExists(brokenLink, true) {
		os.Remove(brokenLink)
		assert.Fail(t, "A broken symLink was downloaded although validate-symlinks flag was set to true")
	}

	// Clean
	cleanArtifactoryTest()
}

// Testing exclude pattern with symlinks.
// This test should not upload any files.
func TestExcludeBrokenSymlink(t *testing.T) {
	if cliutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)

	// Creating broken symlink
	os.Mkdir(tests.Out, 0777)
	linkToNonExistingPath := filepath.Join(tests.Out, "link_to_non_existing_path")
	err := os.Symlink("non_existing_path", linkToNonExistingPath)
	assert.NoError(t, err)

	// This command should succeed because all artifacts are excluded.
	artifactoryCli.Exec("u", filepath.Join(tests.Out, "*"), tests.Repo1, "--symlinks=true", "--exclusions=*")
	cleanArtifactoryTest()
}

// Upload symlink to Artifactory using wildcard pattern and the link content checksum
// Download the symlink which was uploaded.
// validate the symlink content checksum.
func TestSymlinkWildcardPathHandling(t *testing.T) {
	if cliutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a/", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link*")
	artifactoryCli.Exec("u", link1+" "+tests.Repo1+" --symlinks=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/ --validate-symlinks=true")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirHandling(t *testing.T) {
	if cliutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", link+" "+tests.Repo1+" --symlinks=true --recursive=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirWildcardHandling(t *testing.T) {
	if cliutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "lin*")
	artifactoryCli.Exec("u", link1+" "+tests.Repo1+" --symlinks=true --recursive=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
// The test create circular links and the test suppose to prune the circular searching.
func TestSymlinkInsideSymlinkDirWithRecursionIssueUpload(t *testing.T) {
	if cliutils.IsWindows() {
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

	artifactoryCli.Exec("u", localDirPath+"/link* "+tests.Repo1+" --symlinks=true --recursive=true")
	err = os.Remove(link1)
	assert.NoError(t, err)

	err = os.Remove(link2)
	assert.NoError(t, err)

	artifactoryCli.Exec("dl", tests.Repo1+"/link* "+tests.GetTestResourcesPath()+"a/")
	validateSymLink(link1, localDirPath, t)
	os.Remove(link1)
	validateSymLink(link2, localFilePath, t)
	os.Remove(link2)
	cleanArtifactoryTest()
}

func validateSymLink(localLinkPath, localFilePath string, t *testing.T) {
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
		artifactoryCli.Exec("delete", tests.Repo1+"/test_resources/b/*")
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

	artifactoryCli.Exec("delete", tests.Repo1+"/test_resources/*/c")

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
	deletePath := tests.Repo1 + "/test_resources"
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
	resp, body, _, err := client.SendGet(artifactoryDetails.Url+"api/storage/"+tests.Repo1+"/test_resources", true, artHttpDetails)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "test_resources shouldn't be deleted: "+tests.Repo1+"/test_resources/ "+string(body))

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
	artifactoryCli.Exec("del", tests.Repo1+"/data/", "--exclude-patterns=*b1.in;*b2.in;*b3.in;*c1.in")

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
	artifactoryCli.Exec("del", tests.Repo1+"/data/", "--exclusions=*/*b1.in;*/*b2.in;*/*b3.in;*/*c1.in")

	// Validate files are deleted
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteExcludeBySpec(t *testing.T) {
	initArtifactoryTest(t)
	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	specFile, err := tests.CreateSpec(tests.DelSpecExclude)
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

func TestArtifactoryDeleteByProps(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.UploadWithPropsSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)
	// Set properties to the directories as well (and their content)
	artifactoryCli.Exec("sp", tests.Repo1+"/a/b", "D=5", "--include-dirs")
	artifactoryCli.Exec("sp", tests.Repo1+"/a/b/c", "D=2", "--include-dirs")
	//  Set the property D=5 to c1.in, which is a different value then its directory c/
	artifactoryCli.Exec("sp", tests.Repo1+"/a/b/c/c1.in", "D=5")

	// Prepare search command
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.Repo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails)
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())

	// Delete all artifacts with D=5 but without c=3
	artifactoryCli.Exec("delete", tests.Repo1+"/*", "--props=D=5", "--exclude-props=c=3")
	// Search all artifacts in repo1
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchResultAfterDeleteByPropsStep1())

	// Delete all artifacts with c=3 but without a=1
	artifactoryCli.Exec("delete", tests.Repo1+"/*", "--props=c=3", "--exclude-props=a=1")
	// Search all artifacts in repo1
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchResultAfterDeleteByPropsStep2())

	// Delete all artifacts with a=1 but without b=3&c=3
	artifactoryCli.Exec("delete", tests.Repo1+"/*", "--props=a=1", "--exclude-props=b=3;c=3")
	// Search all artifacts in repo1
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchResultAfterDeleteByPropsStep3())

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
	verifyExistInArtifactoryByProps([]string{tests.Repo1 + "/multiple/properties/testsdata/a/b/b2.in"}, tests.Repo1+"/*/properties/*.in", "searchMe=true", t)
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
	dirInnerPath := fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	canonicalPath := tests.Out + dirInnerPath

	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/(*)", tests.Repo1+"/{1}/", "--include-dirs=true", "--recursive=true", "--flat=false")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	expectedPath := []string{tests.Out, "inner", "folder", "out", "inner", "folder"}
	assert.True(t, fileutils.IsPathExists(strings.Join(expectedPath, fileutils.GetFileSeparator()), false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFlatFolderUpload(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	canonicalPath := tests.Out + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/(*)", tests.Repo1+"/{1}/", "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	// Non flat download
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(canonicalPath+fileutils.GetFileSeparator()+"folder", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryIncludeDirFlatNonEmptyFolderUpload(t *testing.T) {
	initArtifactoryTest(t)
	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"a/b/*", tests.Repo1, "--include-dirs=true", "--flat=true")
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/c", false), "Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories - Directories which do not include other directories that match the pattern
func TestArtifactoryDownloadNotIncludeDirs(t *testing.T) {
	initArtifactoryTest(t)
	// 'c' folder is defined as bottom chain directory therefore should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"*/c", tests.Repo1, "--include-dirs=true", "--flat=true")
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--recursive=true")
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
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"(a*)/*", tests.Repo1+"/{1}/", "--include-dirs=true", "--flat=true")
	// Download without include-dirs
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--recursive=true", "--flat=true")
	assert.False(t, fileutils.IsPathExists(tests.Out+"/c", false), "'c' folder shouldn't exist.")

	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--include-dirs=true", "--recursive=true", "--flat=true")
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
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"*/c", tests.Repo1, "--include-dirs=true", "--flat=true")
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
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
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"a/b/*", tests.Repo1, "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(newFolderPath)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
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
	artifactoryCli.Exec("upload", tests.GetTestResourcesPath()+"a/(*)/*", tests.Repo1+"/{1}/", "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(fullPath)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--include-dirs=true", "--recursive=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+relativePath, false), "bottom chain directory, 'd', is missing")

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFlatFolderDownload1(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	canonicalPath := tests.Out + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	// Flat true by default for upload, by using placeholder we indeed create folders hierarchy in Artifactory inner/folder/folder
	artifactoryCli.Exec("upload", tests.Out+"/(*)", tests.Repo1+"/{1}/", "--include-dirs=true")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	// Only the inner folder should be downland e.g 'folder'
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--include-dirs=true", "--flat=true")
	assert.False(t, !fileutils.IsPathExists(tests.Out+fileutils.GetFileSeparator()+"folder", false) &&
		fileutils.IsPathExists(tests.Out+fileutils.GetFileSeparator()+"inner", false),
		"Failed to download folders from Artifactory")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadRecursiveUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := "empty" + fileutils.GetFileSeparator() + "folder"
	canonicalPath := tests.GetTestResourcesPath() + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
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
	canonicalPath := tests.Out + fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/", tests.Repo1, "--recursive=true", "--include-dirs=true")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.Repo1, tests.Out+"/", "--include-dirs=true")
	assert.True(t, fileutils.IsPathExists(tests.Out+"/folder", false), "Failed to download folder from Artifactory")
	assert.False(t, fileutils.IsPathExists(canonicalPath, false), "Path should be flat ")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderDownloadNonRecursive(t *testing.T) {
	initArtifactoryTest(t)
	canonicalPath := tests.Out + fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	err := os.MkdirAll(canonicalPath, 0777)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", tests.Out+"/", tests.Repo1, "--recursive=true", "--include-dirs=true", "--flat=false")
	err = os.RemoveAll(tests.Out)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", tests.Repo1+"/*", "--recursive=false", "--include-dirs=true")
	assert.True(t, fileutils.IsPathExists(tests.Out, false), "Failed to download folder from Artifactory")
	assert.False(t, fileutils.IsPathExists(canonicalPath, false), "Path should be flat. ")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryChecksumDownload(t *testing.T) {
	initArtifactoryTest(t)

	var filePath = "testsdata/a/a1.in"
	artifactoryCli.Exec("upload", filePath, tests.Repo1)
	testChecksumDownload(t, "/a1.in")
	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryChecksumDownloadRenameFileName(t *testing.T) {
	initArtifactoryTest(t)

	var filePath = "testsdata/a/a1.in"
	artifactoryCli.Exec("upload", filePath, tests.Repo1)
	testChecksumDownload(t, "/a1.out")
	// Cleanup
	cleanArtifactoryTest()
}

func testChecksumDownload(t *testing.T, outFileName string) {
	artifactoryCli.Exec("download "+tests.Repo1+"/a1.in", tests.Out+outFileName)

	exists, err := fileutils.IsFileExists(tests.Out+outFileName, false)
	assert.NoError(t, err)
	assert.True(t, exists, "Failed to download file from Artifactory")

	firstFileInfo, _ := os.Stat(tests.Out + outFileName)
	firstDownloadTime := firstFileInfo.ModTime()

	artifactoryCli.Exec("download "+tests.Repo1+"/a1.in", tests.Out+outFileName)
	secondFileInfo, _ := os.Stat(tests.Out + outFileName)
	secondDownloadTime := secondFileInfo.ModTime()

	assert.Equal(t, firstDownloadTime, secondDownloadTime, "Checksum download failed, the file was downloaded twice")
}

func TestArtifactoryDownloadByPatternAndBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryGenericBuildnameAndNumberFromEnv(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	os.Setenv(cliutils.BuildName, buildName)
	os.Setenv(cliutils.BuildNumber, buildNumberA)
	defer os.Unsetenv(cliutils.BuildName)
	defer os.Unsetenv(cliutils.BuildNumber)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	os.Setenv(cliutils.BuildNumber, "11")
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Publish buildInfo
	os.Setenv(cliutils.BuildNumber, buildNumberA)
	artifactoryCli.Exec("build-publish")
	os.Setenv(cliutils.BuildNumber, buildNumberB)
	artifactoryCli.Exec("build-publish")

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildNoPatternUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoPattern)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

// Upload a file to build A.
// Verify that it doesn't exist in B.
func TestArtifactoryDownloadArtifactDoesntExistInBuild(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumber := "cli-test-build1", "10"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumber)
	assert.NoError(t, err)
	// Upload a file
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a10.in", "--build-name="+buildName, "--build-number="+buildNumber)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	// Download from different build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadDoesntExist(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

// Upload a file to 2 different builds.
// Verify that we don't download files with same sha and different build name and build number.
func TestArtifactoryDownloadByShaAndBuild(t *testing.T) {
	initArtifactoryTest(t)
	buildNameA, buildNameB, buildNumberA, buildNumberB, buildNumberC := "cli-test-build1", "cli-test-build2", "10", "11", "12"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameA, artHttpDetails)
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameB, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumber)
	assert.NoError(t, err)

	// Upload 3 similar files to 3 different builds
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a10.in", "--build-name="+buildNameB, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a11.in", "--build-name="+buildNameA, "--build-number="+buildNumberB)
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a12.in", "--build-name="+buildNameA, "--build-number="+buildNumberC)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberA)
	artifactoryCli.Exec("build-publish", buildNameA, buildNumberB)
	artifactoryCli.Exec("build-publish", buildNameA, buildNumberC)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadByShaAndBuild(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameA, artHttpDetails)
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameB, artHttpDetails)
	cleanArtifactoryTest()
}

// Upload a file to 2 different builds.
// Verify that we don't download files with same sha and build name and different build number.
func TestArtifactoryDownloadByShaAndBuildName(t *testing.T) {
	initArtifactoryTest(t)
	buildNameA, buildNameB, buildNumberA, buildNumberB, buildNumberC := "cli-test-build1", "cli-test-build2", "10", "11", "12"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameA, artHttpDetails)
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameB, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumber)
	assert.NoError(t, err)

	// Upload 3 similar files to 2 different builds
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a10.in", "--build-name="+buildNameB, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a11.in", "--build-name="+buildNameB, "--build-number="+buildNumberB)
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a12.in", "--build-name="+buildNameA, "--build-number="+buildNumberC)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildNameA, buildNumberC)
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberA)
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadByShaAndBuildName(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameA, artHttpDetails)
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameB, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildUsingSimpleDownload(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Download by build number, a1 should not be downloaded, b1 should
	artifactoryCli.Exec("download "+tests.Repo1+"/data/a1.in "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--build="+buildName)
	artifactoryCli.Exec("download "+tests.Repo1+"/data/b1.in "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--build="+buildName)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildNoPatternUsingSimpleDownload(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)

	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Download by build number, a1 should not be downloaded, b1 should
	artifactoryCli.Exec("download * "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--build="+buildName+"/"+buildNumberA)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownloadNoPattern(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
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
		[]string{"dl", tests.Repo1, "out/", "--archive-entries=(*)c1.in", "--flat=true"})

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
		[]string{"dl", tests.Repo1, "out/", "--archive-entries=b/c/c1.in", "--flat=true"})

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
	artifactoryCli.Exec("download", tests.Repo1+" out/download/aql_by_artifacts/", "--exclude-patterns=*/a1.in;*a2.*;data/c2.in")

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
	artifactoryCli.Exec("download", tests.Repo1+" out/download/aql_by_artifacts/", "--exclusions=*/*/a1.in;*/*a2.*;*/data/c2.in")

	// Validate files are excluded
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildExcludeDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExcludeBySpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile, err := tests.CreateSpec(tests.DownloadSpecExclude)
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

func TestArtifactoryDownloadExcludeBySpecOverride(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--recursive=true", "--flat=false")
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--recursive=true", "--flat=false")

	// Download by spec
	specFile, err := tests.CreateSpec(tests.DownloadSpecExclude)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+specFile, "--exclude-patterns=*a1.in;*a2.in;*c2.in")

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildExcludeDownload(), paths)
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
	if cliutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", link+" "+tests.Repo1+" --symlinks=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/ --validate-symlinks=true --limit=1")
	validateSortLimitWithSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded with limit param.
// validate the symlink content checksum.
func TestArtifactorySortWithSymlink(t *testing.T) {
	if cliutils.IsWindows() {
		t.Skip("Running on windows, skipping...")
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a", "link")
	err := os.Symlink(localFile, link)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", link+" "+tests.Repo1+" --symlinks=true")
	err = os.Remove(link)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/ --validate-symlinks=true --sort-by=created")
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
	buildNameA, buildNameB, buildNumberA, buildNumberB, buildNumberC := "cli-test-build1", "cli-test-build2", "10", "11", "12"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameA, artHttpDetails)
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameB, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.BuildDownloadSpecNoBuildNumberWithSort)
	assert.NoError(t, err)
	// Upload 3 similar files to 2 different builds
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a10.in", "--build-name="+buildNameB, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a11.in", "--build-name="+buildNameB, "--build-number="+buildNumberB)
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/a12.in", "--build-name="+buildNameA, "--build-number="+buildNumberC)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildNameA, buildNumberC)
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberA)
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--sort-by=created --spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(filepath.Join(tests.Out, "download", "sort_limit_by_build"), false)
	err = tests.ValidateListsIdentical(tests.GetBuildDownloadByShaAndBuildNameWithSort(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameA, artHttpDetails)
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildNameB, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber: a* uploaded with build number "10", b* uploaded with build number "11"
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Copy by build name "cli-test-build" and build number "10" from spec, a* should be copied
	artifactoryCli.Exec("copy", "--spec="+specFile)

	// Validate files are Copied by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildPatternAllUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildPatternAllSpec)
	assert.NoError(t, err)
	// Upload files with buildName and buildNumber: a* uploaded with build number "10", b* uploaded with build number "11"
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Copy by build name "cli-test-build" and build number "10" from spec, a* should be copied
	artifactoryCli.Exec("copy", "--spec="+specFile)

	// Validate files are Copied by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactorySortAndLimit(t *testing.T) {
	initArtifactoryTest(t)

	// Upload all testdata/a/ files
	artifactoryCli.Exec("upload", "testsdata/a/(*)", tests.Repo1+"/data/{1}")

	// Download 1 sorted by name asc
	artifactoryCli.Exec("download", tests.Repo1+"/data/ out/download/sort_limit/", "--sort-by=name", "--limit=1")

	// Download 3 sorted by depth desc
	artifactoryCli.Exec("download", tests.Repo1+"/data/ out/download/sort_limit/", "--sort-by=depth", "--limit=3", "--sort-order=desc")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err := tests.ValidateListsIdentical(tests.GetSortAndLimit(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryOffset(t *testing.T) {
	initArtifactoryTest(t)

	// Upload all testdata/a/ files
	artifactoryCli.Exec("upload", "testsdata/a/*", path.Join(tests.Repo1, "offset_test")+"/", "--flat=true")

	// Downloading files one by one, to check that the offset is working as expected.
	// Download only the first file, expecting to download a1.in
	artifactoryCli.Exec("download", tests.Repo1+"/offset_test/", tests.Out+"/", "--flat=true", "--sort-by=name", "--limit=1", "--offset=0")
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{tests.Out + fileutils.GetFileSeparator() + "a1.in"}, paths, t)

	// Download the second file, expecting to download a2.in
	artifactoryCli.Exec("download", tests.Repo1+"/offset_test/", tests.Out+"/", "--flat=true", "--sort-by=name", "--limit=1", "--offset=1")
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{tests.Out + fileutils.GetFileSeparator() + "a2.in"}, paths, t)

	// Download the third file, expecting to download a3.in
	artifactoryCli.Exec("download", tests.Repo1+"/offset_test/", tests.Out+"/", "--flat=true", "--sort-by=name", "--limit=1", "--offset=2")
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.VerifyExistLocally([]string{tests.Out + fileutils.GetFileSeparator() + "a3.in"}, paths, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildOverridingByInlineFlag(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber: b* uploaded with build number "10", a* uploaded with build number "11"
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Copy by build number: using override of build by flag from inline (no number set so LATEST build should be copied), a* should be copied
	artifactoryCli.Exec("copy", "--build="+buildName+" --spec="+specFile)

	// Validate files are Copied by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestArtifactoryMoveByBuildUsingFlags(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Move by build name and number
	artifactoryCli.Exec("move", "--build="+buildName+"/11 --spec="+specFile)

	// Validate files are moved by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
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
	artifactoryCli.Exec("move", tests.Repo1+"/data/ "+tests.Repo2+"/", "--exclude-patterns=*b*;*c*")

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
	artifactoryCli.Exec("move", tests.Repo1+"/data/ "+tests.Repo2+"/", "--exclusions=*/*b*;*/*c*")

	// Validate excluded files didn't move
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryMoveExcludeBySpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile, err := tests.CreateSpec(tests.MoveCopySpecExclude)
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
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	specFile, err := tests.CreateSpec(tests.CopyByBuildSpec)
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Delete by build name and LATEST
	artifactoryCli.Exec("delete", "--build="+buildName+"/LATEST --spec="+specFile)

	// Validate files are deleted by build number
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

func TestGitLfsCleanup(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "testsdata/gitlfs/(4b)(*)"
	artifactoryCli.Exec("upload", filePath, tests.LfsRepo+"/objects/4b/f4/{2}{1}")
	artifactoryCli.Exec("upload", filePath, tests.LfsRepo+"/objects/4b/f4/")
	refs := strings.Join([]string{"refs", "heads", "*"}, fileutils.GetFileSeparator())
	dotGitPath := getCliDotGitPath(t)
	artifactoryCli.Exec("glc", dotGitPath, "--repo="+tests.LfsRepo, "--refs=HEAD,"+refs)
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

func TestSummaryReport(t *testing.T) {
	initArtifactoryTest(t)

	previousLog := log.Logger
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)

	// Set new logger with output redirection to buffer
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	specFile, err := tests.CreateSpec(tests.UploadFlatNonRecursive)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)
	verifySummary(t, buffer, 3, 0, previousLog)

	artifactoryCli.Exec("move", path.Join(tests.Repo1, "*.in"), tests.Repo2+"/")
	verifySummary(t, buffer, 3, 0, previousLog)

	artifactoryCli.Exec("copy", path.Join(tests.Repo2, "*.in"), tests.Repo1+"/")
	verifySummary(t, buffer, 3, 0, previousLog)

	artifactoryCli.Exec("delete", path.Join(tests.Repo2, "*.in"))
	verifySummary(t, buffer, 3, 0, previousLog)

	artifactoryCli.Exec("set-props", path.Join(tests.Repo1, "*.in"), "prop=val")
	verifySummary(t, buffer, 3, 0, previousLog)

	specFile, err = tests.CreateSpec(tests.DownloadAllRepo1TestResources)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+specFile)
	verifySummary(t, buffer, 3, 0, previousLog)

	// Restore previous logger
	log.SetLogger(previousLog)
	cleanArtifactoryTest()
}

func TestArtifactoryBuildDiscard(t *testing.T) {
	// Initialize
	initArtifactoryTest(t)
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	// Upload files with buildName and buildNumber
	buildName := "jfrog-cli-discard-builds-test"
	for i := 1; i <= 5; i++ {
		artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/data/", "--build-name="+buildName, "--build-number="+strconv.Itoa(i))
		artifactoryCli.Exec("build-publish", buildName, strconv.Itoa(i))
	}

	// Test discard by max-builds
	artifactoryCli.Exec("build-discard", buildName, "--max-builds=3")
	jsonResponse := getAllBuildsByBuildName(client, buildName, t, http.StatusOK)
	assert.Len(t, jsonResponse.Builds, 3, "Incorrect operation of build-discard by max-builds.")

	// Test discard with exclusion
	artifactoryCli.Exec("build-discard", buildName, "--max-days=-1", "--exclude-builds=3,5")
	jsonResponse = getAllBuildsByBuildName(client, buildName, t, http.StatusOK)
	assert.Len(t, jsonResponse.Builds, 2, "Incorrect operation of build-discard with exclusion.")

	// Test discard by max-days
	artifactoryCli.Exec("build-discard", buildName, "--max-days=-1")
	jsonResponse = getAllBuildsByBuildName(client, buildName, t, http.StatusNotFound)
	assert.Zero(t, jsonResponse, "Incorrect operation of build-discard by max-days.")

	//Cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanArtifactoryTest()
}

// Tests compatibility to file paths with windows separators.
// Verifies the upload and download commands work as expected for inputs of both arguments and spec files.
func TestArtifactoryWinBackwardsCompatibility(t *testing.T) {
	initArtifactoryTest(t)
	if !cliutils.IsWindows() {
		t.Skip("Not running on Windows, skipping...")
	}
	uploadSpecFile, err := tests.CreateSpec(tests.WinSimpleUploadSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+uploadSpecFile)
	artifactoryCli.Exec("upload", "testsdata\\\\a\\\\b\\\\*", tests.Repo1+"/compatibility_arguments/", "--exclusions=*b2.in;*c*")

	downloadSpecFile, err := tests.CreateSpec(tests.WinSimpleDownloadSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+downloadSpecFile)
	artifactoryCli.Exec("download", tests.Repo1+"/*arguments*", "out\\\\win\\\\", "--flat=true")

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
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.Repo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails)

	// Search without IncludeDirs
	searchCmd.SetSpec(searchSpecBuilder.IncludeDirs(false).BuildSpec())
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchNotIncludeDirsFiles())

	// Search with IncludeDirs
	searchCmd.SetSpec(searchSpecBuilder.IncludeDirs(true).BuildSpec())
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchIncludeDirsFiles())

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
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.Repo1).Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails)

	// Search artifacts with c=3
	searchCmd.SetSpec(searchSpecBuilder.Props("c=3").BuildSpec())
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchPropsStep1())

	// Search artifacts without c=3
	searchCmd.SetSpec(searchSpecBuilder.Props("").ExcludeProps("c=3").BuildSpec())
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchPropsStep2())

	// Search artifacts without a=1&b=2
	searchCmd.SetSpec(searchSpecBuilder.Props("").ExcludeProps("a=1;b=2").BuildSpec())
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchPropsStep3())

	// Search artifacts without a=1&b=2 and with c=3
	searchCmd.SetSpec(searchSpecBuilder.Props("c=3").ExcludeProps("a=1;b=2").BuildSpec())
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchPropsStep4())

	// Search artifacts without a=1 and with c=5
	searchCmd.SetSpec(searchSpecBuilder.Props("c=5").ExcludeProps("a=1").BuildSpec())
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchPropsStep5())

	// Search artifacts by pattern "*b*", exclude pattern "*3*", with "b=1" and without "c=3"
	pattern := tests.Repo1 + "/*b*"
	exclusions := []string{tests.Repo1 + "/*3*"}
	searchSpecBuilder = spec.NewBuilder().Pattern(pattern).Recursive(true).Exclusions(exclusions).Props("b=1").ExcludeProps("c=3")
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())
	assert.NoError(t, searchCmd.Search())
	assert.NoError(t, assertDateInSearchResult(searchCmd.SearchResult()))
	assert.ElementsMatch(t, searchCmd.SearchResultNoDate(), tests.GetSearchPropsStep6())

	// Cleanup
	cleanArtifactoryTest()
}

func getAllBuildsByBuildName(client *httpclient.HttpClient, buildName string, t *testing.T, expectedHttpStatusCode int) buildsApiResponseStruct {
	resp, body, _, _ := client.SendGet(artifactoryDetails.Url+"api/build/"+buildName, true, artHttpDetails)
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

func verifySummary(t *testing.T, buffer *bytes.Buffer, success, failure int64, logger log.Log) {
	content := buffer.Bytes()
	buffer.Reset()
	logger.Output(string(content))

	status, err := jsonparser.GetString(content, "status")
	assert.NoError(t, err)
	assert.Equal(t, "success", status, "Summary validation failed")

	resultSuccess, err := jsonparser.GetInt(content, "totals", "success")
	assert.NoError(t, err)

	resultFailure, err := jsonparser.GetInt(content, "totals", "failure")
	assert.NoError(t, err)

	assert.Equal(t, success, resultSuccess, "Summary validation failed")
	assert.Equal(t, failure, resultFailure, "Summary validation failed")
}

func CleanArtifactoryTests() {
	cleanArtifactoryTest()
	deleteRepos()
}

func initArtifactoryTest(t *testing.T) {
	if !*tests.TestArtifactory {
		t.Skip("Artifactory is not being tested, skipping...")
	}
}

func cleanArtifactoryTest() {
	if !*tests.TestArtifactory && !*tests.TestDistribution {
		return
	}
	os.Unsetenv(cliutils.HomeDir)
	log.Info("Cleaning test data...")
	cleanArtifactory()
	tests.CleanFileSystem()
}

func preUploadBasicTestResources() {
	uploadPath := tests.GetTestResourcesPath() + "a/(.*)"
	targetPath := tests.Repo1 + "/test_resources/{1}"
	flags := "--threads=10 --regexp=true --props=searchMe=true --flat=false"
	artifactoryCli.Exec("upload", uploadPath, targetPath, flags)
}

func execDeleteRepoRest(repoName string) {
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	resp, body, err := client.SendDelete(artifactoryDetails.Url+"api/repositories/"+repoName, nil, artHttpDetails)
	if err != nil {
		log.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Error(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
		return
	}
	log.Info("Repository", repoName, "deleted.")
}

func execListRepoRest() ([]string, error) {
	var repositoryKeys []string

	// Build http client
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return nil, err
	}

	// Send get request
	resp, body, _, err := client.SendGet(artifactoryDetails.Url+"api/repositories", true, artHttpDetails)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body))
	}

	// Extract repository keys from the json response
	var keyError error
	_, err = jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil || keyError != nil {
			return
		}
		repoKey, err := jsonparser.GetString(value, "key")
		if err != nil {
			keyError = err
			return
		}
		repositoryKeys = append(repositoryKeys, repoKey)
	})
	if keyError != nil {
		return nil, err
	}

	return repositoryKeys, err
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
	resp, body, err := client.SendPut(artifactoryDetails.Url+"api/repositories/"+repoName, content, artHttpDetails)
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

func createReposIfNeeded() {
	cleanUpOldRepositories()
	createRandomReposName()
	nonVirtualRepos := tests.GetNonVirtualRepositories()
	createRepos(nonVirtualRepos)
	virtualRepos := tests.GetVirtualRepositories()
	createRepos(virtualRepos)
}

func cleanUpOldRepositories() {
	repositoryKeys, err := execListRepoRest()
	if err != nil {
		log.Warn("Couldn't retrieve repository list from Artifactory", err)
		return
	}

	now := time.Now()
	repoPattern := regexp.MustCompile(`^jfrog-cli-tests(-\w*)+-(\d*)$`)
	for _, repoKey := range repositoryKeys {
		regexGroups := repoPattern.FindStringSubmatch(repoKey)
		if regexGroups == nil {
			// Repository is not "jfrog-cli-tests-..."
			continue
		}

		repoTimestamp, err := strconv.ParseInt(regexGroups[len(regexGroups)-1], 10, 64)
		if err != nil {
			log.Warn("Error while parsing repository timestamp of repository ", repoKey, err)
			continue
		}

		repoTime := time.Unix(repoTimestamp, 0)
		if now.Sub(repoTime).Hours() > 24.0 {
			log.Info("Deleting old repository", repoKey)
			execDeleteRepoRest(repoKey)
		}
	}
}

func createRepos(repos map[string]string) {
	for repoName, configFile := range repos {
		if !isRepoExist(repoName) {
			repoConfig := tests.GetTestResourcesPath() + configFile
			repoConfig, err := tests.ReplaceTemplateVariables(repoConfig, "")
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			execCreateRepoRest(repoConfig, repoName)
		}
	}
}

func createRandomReposName() {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	tests.Repo1 += "-" + timestamp
	tests.Repo2 += "-" + timestamp
	tests.Repo1And2 += "-" + timestamp
	tests.VirtualRepo += "-" + timestamp
	tests.LfsRepo += "-" + timestamp
	tests.DebianRepo += "-" + timestamp
	tests.JcenterRemoteRepo += "-" + timestamp
	if *tests.TestNpm {
		tests.NpmLocalRepo += "-" + timestamp
		tests.NpmRemoteRepo += "-" + timestamp
	}
	if *tests.TestGo {
		tests.GoLocalRepo += "-" + timestamp
	}
	if *tests.TestPip {
		tests.PypiRemoteRepo += "-" + timestamp
		tests.PypiVirtualRepo += "-" + timestamp
	}
}

// Important - Virtual repositories most be deleted first
func deleteRepos() {
	repos := []string{
		tests.VirtualRepo,
		tests.Repo1,
		tests.Repo2,
		tests.LfsRepo,
		tests.DebianRepo,
		tests.JcenterRemoteRepo,
	}

	if *tests.TestNpm {
		repos = append(repos, tests.NpmLocalRepo, tests.NpmRemoteRepo)
	}
	if *tests.TestGo {
		repos = append(repos, tests.GoLocalRepo)
	}
	if *tests.TestPip {
		repos = append(repos, tests.PypiVirtualRepo, tests.PypiRemoteRepo)
	}

	for _, repoName := range repos {
		if isRepoExist(repoName) {
			execDeleteRepoRest(repoName)
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
	tests.DeleteFiles(deleteSpec, artifactoryDetails)
}

func searchInArtifactory(specFile string) ([]generic.SearchResult, error) {
	searchSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(searchSpec)
	err := searchCmd.Search()
	return searchCmd.SearchResult(), err
}

func getSpecAndCommonFlags(specFile string) (*spec.SpecFiles, rtutils.CommonConf) {
	searchFlags := new(rtutils.CommonConfImpl)
	searchSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	return searchSpec, searchFlags
}

func verifyExistInArtifactory(expected []string, specFile string, t *testing.T) {
	results, _ := searchInArtifactory(specFile)
	tests.CompareExpectedVsActual(expected, results, t)
}

func verifyDoesntExistInArtifactory(specFile string, t *testing.T) {
	verifyExistInArtifactory([]string{}, specFile, t)
}

func verifyExistInArtifactoryByProps(expected []string, pattern, props string, t *testing.T) {
	searchSpec := spec.NewBuilder().Pattern(pattern).Props(props).Recursive(true).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(searchSpec)
	err := searchCmd.Search()
	assert.NoError(t, err)
	tests.CompareExpectedVsActual(expected, searchCmd.SearchResult(), t)
}

func isRepoExist(repoName string) bool {
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	resp, _, _, err := client.SendGet(artifactoryDetails.Url+tests.RepoDetailsUrl+repoName, true, artHttpDetails)
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
	configArtifactoryCli.Exec("c", "delete", tests.RtServerId, "--interactive=false")
}

// This function will create server config and return the entire passphrase flag if it needed.
// For example if passphrase is needed it will return "--ssh-passphrase=${theConfiguredPassphrase}" or empty string.
func createServerConfigAndReturnPassphrase() (passphrase string) {
	if *tests.RtSshPassphrase != "" {
		passphrase = "--ssh-passphrase=" + *tests.RtSshPassphrase
	}
	configArtifactoryCli.Exec("c", tests.RtServerId, "--interactive=false")
	return passphrase
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
	artifactoryCli.Exec(command, tests.Repo1+"/data/*a* "+tests.Repo2+"/", "--dry-run")

	// Validate files weren't affected
	cpMvSpecFilePath, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(beforeCommandExpected, cpMvSpecFilePath, t)

	// Run command
	artifactoryCli.Exec(command, tests.Repo1+"/data/*a* "+tests.Repo2+"/")

	// Validate files were affected
	verifyExistInArtifactory(afterCommandExpected, cpMvSpecFilePath, t)

	// Cleanup
	cleanArtifactoryTest()
}

func searchItemsInArtifactory(t *testing.T) []rtutils.ResultItem {
	fileSpec, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	spec, flags := getSpecAndCommonFlags(fileSpec)
	flags.SetArtifactoryDetails(artAuth)
	var resultItems []rtutils.ResultItem
	for i := 0; i < len(spec.Files); i++ {

		searchParams, err := generic.GetSearchParams(spec.Get(i))
		assert.NoError(t, err, "Failed Searching files")

		currentResultItems, err := services.SearchBySpecFiles(searchParams, flags, rtutils.ALL)
		assert.NoError(t, err, "Failed Searching files")
		resultItems = append(resultItems, currentResultItems...)
	}
	return resultItems
}

func assertDateInSearchResult(searchResult []generic.SearchResult) error {
	for _, v := range searchResult {
		if v.Created == "" || v.Modified == "" {
			message, err := json.Marshal(&v)
			if err != nil {
				return errors.New("failed to process search result to assert it includes date: " + err.Error())
			}
			return errors.New("search result does not include date: " + string(message))
		}
	}
	return nil
}

func TestArtifactoryUploadInflatedPath(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "testsdata/a/../a/a1.*", tests.Repo1)
	artifactoryCli.Exec("upload", "testsdata/./a/a1.*", tests.Repo1)
	searchFilePath, err := tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo1(), searchFilePath, t)

	artifactoryCli.Exec("upload", "testsdata/./a/../a/././././a2.*", tests.Repo1)
	searchFilePath, err = tests.CreateSpec(tests.SearchRepo1ByInSuffix)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpected2filesRepo1(), searchFilePath, t)
	if cliutils.IsWindows() {
		artifactoryCli.Exec("upload", `testsdata\\a\\..\\a\\a1.*`, tests.Repo2)
		artifactoryCli.Exec("upload", `testsdata\\.\\\a\a1.*`, tests.Repo2)
		searchFilePath, err = tests.CreateSpec(tests.SearchRepo2)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpectedRepo2(), searchFilePath, t)

		artifactoryCli.Exec("upload", `testsdata\\.\\a\\..\\a\\.\\.\\.\\.\\a2.*`, tests.Repo2)
		searchFilePath, err = tests.CreateSpec(tests.SearchRepo2)
		assert.NoError(t, err)
		verifyExistInArtifactory(tests.GetSimpleUploadSpecialCharNoRegexExpected2filesRepo2(), searchFilePath, t)
	}
	cleanArtifactoryTest()
}

func TestGetJcenterRemoteDetails(t *testing.T) {
	initArtifactoryTest(t)
	createServerConfigAndReturnPassphrase()

	unsetEnvVars := func() {
		err := os.Unsetenv(utils.JCenterRemoteServerEnv)
		assert.NoError(t, err)
		err = os.Unsetenv(utils.JCenterRemoteRepoEnv)
		assert.NoError(t, err)
	}
	unsetEnvVars()
	defer unsetEnvVars()

	// The utils.JCenterRemoteServerEnv env var is not set, so extractor1.jar should be downloaded from jcenter.
	downloadPath := "org/jfrog/buildinfo/build-info-extractor/extractor1.jar"
	expectedRemotePath := path.Join("bintray/jcenter", downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	// Still, the utils.JCenterRemoteServerEnv env var is not set, so the download should be from jcenter.
	// Expecting a different download path this time.
	downloadPath = "org/jfrog/buildinfo/build-info-extractor/extractor2.jar"
	expectedRemotePath = path.Join("bintray/jcenter", downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	// Setting the utils.JCenterRemoteServerEnv env var now,
	// Expecting therefore the download to be from the the server ID configured by this env var.
	err := os.Setenv(utils.JCenterRemoteServerEnv, tests.RtServerId)
	assert.NoError(t, err)
	downloadPath = "org/jfrog/buildinfo/build-info-extractor/extractor3.jar"
	expectedRemotePath = path.Join("jcenter", downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	// Still expecting the download to be from the same server ID, but this time, not through a remote repo named
	// jcenter, but through test-remote-repo.
	testRemoteRepo := "test-remote-repo"
	err = os.Setenv(utils.JCenterRemoteRepoEnv, testRemoteRepo)
	assert.NoError(t, err)
	downloadPath = "1org/jfrog/buildinfo/build-info-extractor/extractor4.jar"
	expectedRemotePath = path.Join(testRemoteRepo, downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)
	cleanArtifactoryTest()
}

func validateJcenterRemoteDetails(t *testing.T, downloadPath, expectedRemotePath string) {
	artDetails, remotePath, err := utils.GetJcenterRemoteDetails(downloadPath)
	assert.NoError(t, err)
	assert.Equal(t, expectedRemotePath, remotePath)
	assert.False(t, os.Getenv(utils.JCenterRemoteServerEnv) != "" && artDetails == nil, "Expected a server to be returned")
}

func TestVcsProps(t *testing.T) {
	initArtifactoryTest(t)
	testDir := initVcsTestDir(t)
	artifactoryCli.Exec("upload", filepath.Join(testDir, "*"), tests.Repo1, "--flat=false", "--build-name=jfrog-cli-test-vcs-props", "--build-number=2020")
	resultItems := searchItemsInArtifactory(t)
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
	testsdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "vcs")
	testsdataTarget := tests.Temp
	err := fileutils.CopyDir(testsdataSrc, testsdataTarget, true)
	assert.NoError(t, err)
	if found, err := fileutils.IsDirExists(filepath.Join(testsdataTarget, "gitdata"), false); found {
		assert.NoError(t, err)
		tests.RenamePath(filepath.Join(testsdataTarget, "gitdata"), filepath.Join(testsdataTarget, ".git"), t)
	}
	if found, err := fileutils.IsDirExists(filepath.Join(testsdataTarget, "OtherGit", "gitdata"), false); found {
		assert.NoError(t, err)
		tests.RenamePath(filepath.Join(testsdataTarget, "OtherGit", "gitdata"), filepath.Join(testsdataTarget, "OtherGit", ".git"), t)
	}
	path, err := filepath.Abs(tests.Temp)
	assert.NoError(t, err)
	return path
}

func TestArtifactoryReplicationCreate(t *testing.T) {
	initArtifactoryTest(t)
	configArtifactoryCli.Exec("c", tests.RtServerId, "--url="+*tests.RtUrl, "--user="+*tests.RtUser, "--password="+*tests.RtPassword, "--apikey="+*tests.RtApiKey, "--access-token="+*tests.RtAccessToken, "--interactive=false")
	defer deleteServerConfig()

	// Init tmp dir
	specFile, err := tests.CreateSpec(tests.ReplicationTempCreate)
	assert.NoError(t, err)

	// Create push replication
	err = artifactoryCli.Exec("rplc", specFile)
	assert.NoError(t, err)

	// Validate create replication
	servicesManager, err := utils.CreateServiceManager(artifactoryDetails, false)
	assert.NoError(t, err)
	result, err := servicesManager.GetReplication(tests.Repo1)
	assert.NoError(t, err)
	result[0].Password = ""
	assert.ElementsMatch(t, result, tests.GetReplicationConfig())

	// Delete replication
	err = artifactoryCli.Exec("rpldel", tests.Repo1)
	assert.NoError(t, err)

	// Validate delete replication
	result, err = servicesManager.GetReplication(tests.Repo1)
	assert.Error(t, err)
	// Cleanup
	cleanArtifactoryTest()
}
