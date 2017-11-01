package main

import (
	"testing"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/commands"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/tests"
	"fmt"
	"io/ioutil"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/buger/jsonparser"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"encoding/json"
	rtutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	cliproxy "github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/tests/proxy/server"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"net/url"
	"net"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"strconv"
	"bytes"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/tests/proxy/server/certificate"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/spec"
	"os/exec"
	"crypto/tls"
	"time"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/tests/xray"
)

var artifactoryCli *tests.JfrogCli
var artifactoryDetails *config.ArtifactoryDetails
var artAuth *auth.ArtifactoryDetails
var artHttpDetails httputils.HttpClientDetails

func InitArtifactoryTests() {
	if !*tests.TestArtifactory {
		return
	}
	os.Setenv("JFROG_CLI_OFFER_CONFIG", "false")
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(main, "jfrog rt", cred)
	if err := createReposIfNeeded(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	cleanArtifactoryTest()
}

func authenticate() string {
	artifactoryDetails = &config.ArtifactoryDetails{Url: cliutils.AddTrailingSlashIfNeeded(*tests.RtUrl), SshKeyPath: *tests.RtSshKeyPath, SshPassphrase: *tests.RtSshPassphrase}
	cred := "--url=" + *tests.RtUrl
	if !artifactoryDetails.IsSsh() {
		if *tests.RtApiKey != "" {
			artifactoryDetails.ApiKey = *tests.RtApiKey
		} else {
			artifactoryDetails.User = *tests.RtUser
			artifactoryDetails.Password = *tests.RtPassword
		}
	}
	cred += getArtifactoryTestCredentials()
	var err error
	if artAuth, err = artifactoryDetails.CreateArtAuthConfig(); err != nil {
		cliutils.Exit(cliutils.ExitCodeError, "Failed while attempting to authenticate with Artifactory: " + err.Error())
	}
	artifactoryDetails.SshAuthHeaders = artAuth.SshAuthHeaders
	artifactoryDetails.Url = artAuth.Url
	artHttpDetails = artAuth.CreateArtifactoryHttpClientDetails()
	return cred
}

func getArtifactoryTestCredentials() string {
	if artifactoryDetails.IsSsh() {
		return getSshCredentials()
	}
	if *tests.RtApiKey != "" {
		return " --apikey=" + *tests.RtApiKey
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
	specFile := tests.GetFilePath(tests.SimpleUploadSpec)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	isExistInArtifactory(tests.SimpleUploadExpectedRepo1, tests.GetFilePath(tests.Search), t)
	cleanArtifactoryTest()
}

func TestArtifactorySimpleUploadSpecUsingConfig(t *testing.T) {
	initArtifactoryTest(t)
	const rtServerId = "rtTestServerId"
	var passphrase string
	if *tests.RtSshPassphrase != "" {
		passphrase = "--ssh-passphrase=" + *tests.RtSshPassphrase
	}
	artifactoryCli.Exec("c", rtServerId, "--interactive=false")
	artifactoryCommandExecutor := tests.NewJfrogCli(main, "jfrog rt", "")
	specFile := tests.GetFilePath(tests.SimpleUploadSpec)
	artifactoryCommandExecutor.Exec("upload", "--spec="+specFile, "--server-id="+rtServerId, passphrase)
	isExistInArtifactory(tests.SimpleUploadExpectedRepo1, tests.GetFilePath(tests.Search), t)
	artifactoryCommandExecutor.Exec("c", "delete", rtServerId, "--interactive=false")
	cleanArtifactoryTest()
}

func TestArtifactoryUploadPathWithSpecialCharsAsNoRegex(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1)
	isExistInArtifactory(tests.SimpleUploadSpecialCharNoRegexExpectedRepo1, tests.GetFilePath(tests.Search), t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopySingleFileNonFlat(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/")
	artifactoryCli.Exec("cp", tests.Repo1+"/path/a1.in", tests.Repo2)
	isExistInArtifactory(tests.SingleFileCopy, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestAqlFindingItemOnRoot(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/")
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/*", tests.Repo2)
	isExistInArtifactory(tests.AnyItemCopy, tests.GetFilePath(tests.SearchRepo2), t)
	artifactoryCli.Exec("del", tests.Repo2+"/*", "--quiet=true")
	artifactoryCli.Exec("del", tests.Repo1+"/*", "--quiet=true")
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/")
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/*/", tests.Repo2)
	isExistInArtifactory(tests.SingleFileCopy, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopy(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/")
	artifactoryCli.Exec("cp", tests.Repo1+"/path/", tests.Repo2)
	isExistInArtifactory(tests.SingleFileCopy, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyUsingWildcard(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/")
	artifactoryCli.Exec("cp", tests.Repo1+"/*/", tests.Repo2)
	isExistInArtifactory(tests.SingleFileCopy, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyUsingWildcardFlat(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/", )
	artifactoryCli.Exec("cp", tests.Repo1+"/path/inner", tests.Repo2, "--flat=true")
	isExistInArtifactory(tests.SingleDirectoryCopyFlat, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyPathsTwice(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/", )

	t.Log("Copy Folder to root twice")
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2)
	isExistInArtifactory(tests.SingleFileCopyFullPath, tests.GetFilePath(tests.SearchRepo2), t)
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2)
	isExistInArtifactory(tests.SingleFileCopyFullPath, tests.GetFilePath(tests.SearchRepo2), t)
	artifactoryCli.Exec("del", tests.Repo2, "--quiet=true")

	t.Log("Copy to from repo1/path to repo2/path twice")
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2+"/path")
	isExistInArtifactory(tests.SingleFileCopyFullPath, tests.GetFilePath(tests.SearchRepo2), t)
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2+"/path")
	isExistInArtifactory(tests.FolderCopyTwice, tests.GetFilePath(tests.SearchRepo2), t)
	artifactoryCli.Exec("del", tests.Repo2, "--quiet=true")

	t.Log("Copy to from repo1/path/ to repo2/path/ twice")
	artifactoryCli.Exec("cp", tests.Repo1+"/path/", tests.Repo2+"/path/")
	isExistInArtifactory(tests.SingleInnerFileCopyFullPath, tests.GetFilePath(tests.SearchRepo2), t)
	artifactoryCli.Exec("cp", tests.Repo1+"/path/", tests.Repo2+"/path/")
	isExistInArtifactory(tests.SingleInnerFileCopyFullPath, tests.GetFilePath(tests.SearchRepo2), t)
	artifactoryCli.Exec("del", tests.Repo2, "--quiet=true")

	t.Log("Copy to from repo1/path/ to repo2/path/ twice")
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2+"/path/")
	isExistInArtifactory(tests.FolderCopyIntoFolder, tests.GetFilePath(tests.SearchRepo2), t)
	artifactoryCli.Exec("cp", tests.Repo1+"/path", tests.Repo2+"/path/")
	isExistInArtifactory(tests.FolderCopyIntoFolder, tests.GetFilePath(tests.SearchRepo2), t)
	artifactoryCli.Exec("del", tests.Repo2, "--quiet=true")

	cleanArtifactoryTest()
}

func TestArtifactoryDirectoryCopyPatternEndsWithSlash(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/", )
	artifactoryCli.Exec("cp", tests.Repo1+"/path/", tests.Repo2, "--flat=true")
	isExistInArtifactory(tests.AnyItemCopyUsingSpec, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemUsingWildcardFlat(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/", )
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/*", tests.Repo2)
	isExistInArtifactory(tests.AnyItemCopy, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemRecursive(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/a/b/", )
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/aFile", "--flat=true")
	artifactoryCli.Exec("cp", tests.Repo1+"/a*", tests.Repo2, "--recursive=true")
	isExistInArtifactory(tests.AnyItemCopyRecursive, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAndRenameFolder(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/", )
	artifactoryCli.Exec("cp", tests.Repo1+"/*", tests.Repo2+"/newPath")
	isExistInArtifactory(tests.CopyFolderRename, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyAnyItemUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a+a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a+a\\a*")
	}

	specFile := tests.GetFilePath(tests.CopyItemsSpec)
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/path/inner/", )
	artifactoryCli.Exec("upload", filePath, tests.Repo1+"/someFile", "--flat=true")
	artifactoryCli.Exec("cp", "--spec="+specFile)
	isExistInArtifactory(tests.AnyItemCopyUsingSpec, tests.GetFilePath(tests.SearchRepo2), t)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyExcludeByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileA)
	artifactoryCli.Exec("upload", "--spec=" + specFileB)

	// Copy by pattern
	artifactoryCli.Exec("cp", "jfrog-cli-tests-repo1/data/ jfrog-cli-tests-repo2/", "--exclude-patterns=*b*;*c*")

	// Validate files are moved by build number
	isExistInArtifactory(tests.BuildCopyExclude, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryCopyExcludeBySpec(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileA)
	artifactoryCli.Exec("upload", "--spec=" + specFileB)

	// Copy by spec
	specFile := tests.GetFilePath(tests.MoveCopySpecExclude)
	artifactoryCli.Exec("cp", "--spec=" + specFile)

	// Validate files are moved by build number
	isExistInArtifactory(tests.BuildCopyExclude, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadandExplode(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "../testsdata/a.zip", "jfrog-cli-tests-repo1", "--explode=true")
	isExistInArtifactory(tests.ExplodeUploadExpectedRepo1, tests.GetFilePath(tests.Search), t)
	cleanArtifactoryTest()
}

// Test self-signed certificates with Artifactory. For the test, we set up a reverse proxy server.
func TestArtifactorySelfSignedCert(t *testing.T) {
	initArtifactoryTest(t)
	path, err := ioutil.TempDir("", "jfrog.cli.test.")
	err = errorutils.CheckError(err)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(path)
	os.Setenv("JFROG_CLI_HOME", path)
	go cliproxy.StartLocalReverseHttpProxy(artifactoryDetails.Url)

	// The two certificate files are created by the reverse proxy on startup in the current directory.
	defer os.Remove(certificate.KEY_FILE)
	defer os.Remove(certificate.CERT_FILE)
	// Let's wait for the reverse proxy to start up.
	checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https")
	spec := spec.NewBuilder().Pattern("jfrog-cli-tests-repo1/*.zip").Recursive(true).BuildSpec()
	if err != nil {
		t.Error(err)
	}
	parsedUrl, err := url.Parse(artifactoryDetails.Url)
	artifactoryDetails.Url = "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()
	_, err = commands.Search(spec, artifactoryDetails)
	// The server is using self-sign certificate
	// Without loading the certificated we expect all actions to fail due to error: "x509: certificate signed by unknown authority"
	if _, ok := err.(*url.Error); !ok {
		t.Error("Expected a connection failure, since reverse proxy didn't load self-signed-certs. Connection however is successful", err)
	}

	securityDirPath, err := utils.GetJfrogSecurityDir()
	if err != nil {
		t.Error(err)
	}
	// We need to copy the server certificate to the Cli security dir.
	err = fileutils.CopyFile(securityDirPath, certificate.KEY_FILE)
	if err != nil {
		t.Error(err)
	}
	err = fileutils.CopyFile(securityDirPath, certificate.CERT_FILE)
	if err != nil {
		t.Error(err)
	}
	_, err = commands.Search(spec, artifactoryDetails)
	if err != nil {
		t.Error(err)
	}
	artifactoryDetails.Url = artAuth.Url
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
	return "", errors.New("are you connected to the network?")
}

// Due the fact that go read the HTTP_PROXY and the HTTPS_PROXY
// argument only once we can't set the env var for specific test.
// We need to start a new process with the env var set to the value we want.
// We decide which var to set by the rtUrl scheme.
func TestArtifactoryProxy(t *testing.T) {
	initArtifactoryTest(t)
	rtUrl, err := url.Parse(artifactoryDetails.Url)
	if err != nil {
		t.Error(err)
	}
	var proxyTestArgs []string
	var httpProxyEnv string
	testArgs := []string{"-test.artifactoryProxy=true", "-rt.url="+*tests.RtUrl, "-rt.password="+*tests.RtPassword, "-rt.apikey="+*tests.RtApiKey, "-rt.sshKeyPath="+*tests.RtSshKeyPath, "-rt.sshPassphrase="+*tests.RtSshPassphrase}
	if rtUrl.Scheme == "https" {
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
	command := exec.Command("go", proxyTestArgs...)
	command.Env = append(os.Environ(), httpProxyEnv)
	output, err := command.Output()
	if err != nil {
		t.Error(err)
	}
	cliutils.CliLogger.Info(string(output))
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
	if !*tests.TestArtifactoryProxy {
		t.SkipNow()
	}
	authenticate()
	proxyRtUrl := prepareArtifactoryUrlForProxyTest(t)
	spec := spec.NewBuilder().Pattern("jfrog-cli-tests-repo1/*.zip").Recursive(true).BuildSpec()
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
	checkIfServerIsUp(port, "http")
	_, err := commands.Search(spec, artifactoryDetails)
	if err != nil {
		t.Error(err)
	}
	artifactoryDetails.Url = artAuth.Url
}

func prepareArtifactoryUrlForProxyTest(t *testing.T) string {
	rtUrl, err := url.Parse(artifactoryDetails.Url)
	if err != nil {
		t.Error(err)
	}
	rtHost, port, err := net.SplitHostPort(rtUrl.Host)
	if rtHost == "localhost" || rtHost == "127.0.0.1" {
		externalIp, err := getExternalIP()
		if err != nil {
			t.Error(err)
		}
		rtUrl.Host = externalIp + ":" + port
	}
	return rtUrl.String()
}

func checkForErrDueToMissingProxy(spec *spec.SpecFiles, t *testing.T) {
	_, err := commands.Search(spec, artifactoryDetails)
	_, isUrlErr := err.(*url.Error)
	if err == nil || !isUrlErr {
		t.Error("Expected the request to fails, since the proxy is down.", err)
	}
}

func checkIfServerIsUp(port, proxyScheme string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	attempt := 0
	for attempt < 10 {
		cliutils.CliLogger.Info("Checking if proxy server is up and running.", strconv.Itoa(attempt +1), "attempt.", "URL:", proxyScheme + "://localhost:" + port)
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
	return fmt.Errorf("Failed while waiting for the proxy server to be accessible.")
}

func TestXrayScanBuild(t *testing.T) {
	initArtifactoryTest(t)
	xrayServerPort := xray.StartXrayMockServer()
	serverUrl := "--url=http://localhost:" + strconv.Itoa(xrayServerPort)
	artifactoryCommandExecutor := tests.NewJfrogCli(main, "jfrog rt", serverUrl+getArtifactoryTestCredentials())
	artifactoryCommandExecutor.Exec("build-scan", xray.CleanScanBuildName, "3")

	cleanArtifactoryTest()
}

func TestArtifactorySetProperties(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/a.in")
	artifactoryCli.Exec("sp", "jfrog-cli-tests-repo1/a.*", "prop=val")
	spec, flags := getSpecAndCommonFlags(tests.GetFilePath(tests.Search))
	flags.SetArtifactoryDetails(artAuth)
	var resultItems []rtutils.ResultItem
	for i := 0; i < len(spec.Files); i++ {
		params, err := spec.Get(i).ToArtifatorySetPropsParams()
		if err != nil {
			t.Error(err)
		}
		currentResultItems, err := rtutils.SearchBySpecFiles(&rtutils.SearchParamsImpl{ArtifactoryCommonParams: params}, flags)
		if err != nil {
			t.Error("Failed Searching files:", err)
		}
		resultItems = append(resultItems, currentResultItems...)
	}

	for _, item := range resultItems {
		properties := item.Properties
		if len(properties) < 1 {
			t.Error("Failed setting properties on item:", item.GetItemRelativePath())
		}
		for i, prop := range properties {
			if i > 0 {
				t.Error("Expected single property.")
			}
			if prop.Key != "prop" || prop.Value != "val" {
				t.Error("Wrong properties")
			}
		}
	}
	cleanArtifactoryTest()
}

func TestArtifactorySetPropertiesExcludeByCli(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "../testsdata/a/a*.in", "jfrog-cli-tests-repo1/")
	artifactoryCli.Exec("sp", "jfrog-cli-tests-repo1/*", "prop=val", "--exclude-patterns=*a1.in;*a2.in")
	spec, flags := getSpecAndCommonFlags(tests.GetFilePath(tests.Search))
	flags.SetArtifactoryDetails(artAuth)
	var resultItems []rtutils.ResultItem
	for i := 0; i < len(spec.Files); i++ {
		params, err := spec.Get(i).ToArtifatorySetPropsParams()
		if err != nil {
			t.Error(err)
		}
		currentResultItems, err := rtutils.SearchBySpecFiles(&rtutils.SearchParamsImpl{ArtifactoryCommonParams: params}, flags)
		if err != nil {
			t.Error("Failed Searching files:", err)
		}
		resultItems = append(resultItems, currentResultItems...)
	}

	for _, item := range resultItems {
		log.Info(item.Name)
		if item.Name != "a3.in" {
			continue
		}
		properties := item.Properties
		if len(properties) < 1 {
			t.Error("Failed setting properties on item:", item.GetItemRelativePath())
		}
		for i, prop := range properties {
			if i > 0 {
				t.Error("Expected single property.")
			}
			if prop.Key != "prop" || prop.Value != "val" {
				t.Error("Wrong properties")
			}
		}
	}
	cleanArtifactoryTest()
}

func TestArtifactoryUploadFromHomeDir(t *testing.T) {
	initArtifactoryTest(t)

	testFileRel := "~" + fileutils.GetFileSeparator() + "cliTestFile.txt"
	testFileAbs := fileutils.GetHomeDir() + "/cliTestFile.txt"

	d1 := []byte("test file")
	err := ioutil.WriteFile(testFileAbs, d1, 0644)
	if err != nil {
		t.Error("Couldn't create file:", err)
	}

	artifactoryCli.Exec("upload", testFileRel, tests.Repo1, "--recursive=false")
	isExistInArtifactory(tests.TxtUploadExpectedRepo1, tests.GetFilePath(tests.SearchTxt), t)

	os.Remove(testFileAbs)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli1Wildcard(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a/a*"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a\\a*")
	}

	// Upload files
	artifactoryCli.Exec("upload", filePath, tests.Repo1, "--exclude-patterns=*a2*;*a3.in")
	isExistInArtifactory(tests.SimpleUploadSpecialCharNoRegexExpectedRepo1, tests.GetFilePath(tests.Search), t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli1Regex(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/a/a(.*)"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\a\\a(.*)")
	}

	// Upload files
	artifactoryCli.Exec("upload", filePath, tests.Repo1, "--exclude-patterns=(.*)a2.*;.*a3.in", "--regexp=true")
	isExistInArtifactory(tests.SimpleUploadSpecialCharNoRegexExpectedRepo1, tests.GetFilePath(tests.Search), t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli2Wildcard(t *testing.T) {
	initArtifactoryTest(t)

	// Create temp dir
	absDirPath, err := ioutil.TempDir("", "cliTestDir")
	if err != nil {
		t.Error("Couldn't create dir:", err)
	}
	defer os.Remove(absDirPath)
	if runtime.GOOS == "windows" {
		absDirPath = tests.FixWinPath(absDirPath) + "\\\\"
	} else {
		absDirPath += "/"
	}

	// Create temp files
	d1 := []byte("test file")
	err = ioutil.WriteFile(absDirPath+"cliTestFile1.in", d1, 0644)
	if err != nil {
		t.Error("Couldn't create file:", err)
	}
	err = ioutil.WriteFile(absDirPath+"cliTestFile2.in", d1, 0644)
	if err != nil {
		t.Error("Couldn't create file:", err)
	}

	// Upload files
	artifactoryCli.Exec("upload", absDirPath+"*", tests.Repo1, "--exclude-patterns=*cliTestFile1*")

	// Check files exists in artifactory
	isExistInArtifactory([]string{tests.Repo1 + "/cliTestFile2.in"}, tests.GetFilePath(tests.Search), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeByCli2Regex(t *testing.T) {
	initArtifactoryTest(t)

	// Create temp dir
	absDirPath, err := ioutil.TempDir("", "cliTestDir")
	if err != nil {
		t.Error("Couldn't create dir:", err)
	}
	defer os.Remove(absDirPath)
	if runtime.GOOS == "windows" {
		absDirPath = tests.FixWinPath(absDirPath) + "\\\\"
	} else {
		absDirPath += "/"
	}

	// Create temp files
	d1 := []byte("test file")
	err = ioutil.WriteFile(absDirPath+"cliTestFile1.in", d1, 0644)
	if err != nil {
		t.Error("Couldn't create file:", err)
	}
	err = ioutil.WriteFile(absDirPath+"cliTestFile2.in", d1, 0644)
	if err != nil {
		t.Error("Couldn't create file:", err)
	}

	// Upload files
	artifactoryCli.Exec("upload", absDirPath+"(.*)", tests.Repo1, "--exclude-patterns=(.*c)liTestFile1.*", "--regexp=true")

	// Check files exists in artifactory
	isExistInArtifactory([]string{tests.Repo1 + "/cliTestFile2.in"}, tests.GetFilePath(tests.Search), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeBySpecWildcard(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFile := tests.GetFilePath(tests.UploadSpecExclude)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	// Validate files are moved by build number
	isExistInArtifactory(tests.UploadSpecExcludeRepo1, tests.GetFilePath(tests.Search), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryUploadExcludeBySpecRegex(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFile := tests.GetFilePath(tests.UploadSpecExcludeRegex)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	// Validate files are moved by build number
	isExistInArtifactory(tests.UploadSpecExcludeRepo1, tests.GetFilePath(tests.Search), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryCopySpec(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.MoveCopyDeleteSpec)
	artifactoryCli.Exec("copy", "--spec="+specFile)

	isExistInArtifactory(tests.MassiveMoveExpected, tests.GetFilePath(tests.SearchMoveDeleteRepoSpec), t)
	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded.
// validate the symlink content checksum.
func TestSimpleSymlinkHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a/", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("u", link+" "+tests.Repo1+" --symlinks=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/ --validate-symlinks=true")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink to Artifactory using wildcard pattern and the link content checksum
// Download the symlink which was uploaded.
// validate the symlink content checksum.
func TestSymlinkWildcardPathHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath()+"a/", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link*")
	artifactoryCli.Exec("u", link1+" "+tests.Repo1+" --symlinks=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/ --validate-symlinks=true")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("u", link+" "+tests.Repo1+" --symlinks=true --recursive=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirWilcardHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath()+"a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "lin*")
	artifactoryCli.Exec("u", link1+" "+tests.Repo1+" --symlinks=true --recursive=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1+"/link "+tests.GetTestResourcesPath()+"a/")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
// The test create circular links and the test suppose to prune the circular searching.
func TestSymlinkInsideSymlinkDirWithRecursionIssueUpload(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localDirPath := filepath.Join(tests.GetTestResourcesPath(), "a")
	link1 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link1")
	err := os.Symlink(localDirPath, link1)
	if err != nil {
		t.Error(err.Error())
	}
	localFilePath := filepath.Join(tests.GetTestResourcesPath()+"a/", "a1.in")
	link2 := filepath.Join(tests.GetTestResourcesPath()+"a/", "link2")
	err = os.Symlink(localFilePath, link2)
	if err != nil {
		t.Error(err.Error())
	}

	artifactoryCli.Exec("u", localDirPath+"/link* "+tests.Repo1+" --symlinks=true --recursive=true")
	err = os.Remove(link1)
	if err != nil {
		t.Error(err.Error())
	}

	err = os.Remove(link2)
	if err != nil {
		t.Error(err.Error())
	}

	artifactoryCli.Exec("dl", tests.Repo1+"/link* "+tests.GetTestResourcesPath()+"a/")
	validateSymLink(link1, localDirPath, t)
	os.Remove(link1)
	validateSymLink(link2, localFilePath, t)
	os.Remove(link2)
	cleanArtifactoryTest()
}

func validateSymLink(localLinkPath, localFilePath string, t *testing.T) {
	exists := fileutils.IsPathSymlink(localLinkPath)
	if !exists {
		t.Error(errors.New("Faild to download symlinks from artifactory"))
	}
	symlinks, err := filepath.EvalSymlinks(localLinkPath)
	if err != nil {
		t.Error(errors.New("Can't eval symlinks"))
	}
	if symlinks != localFilePath {
		t.Error(errors.New("Symlinks wasn't created as expected. expected:" + localFilePath + " actual: " + symlinks))
	}
}

func TestArtifactoryDelete(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.MoveCopyDeleteSpec)
	artifactoryCli.Exec("copy", "--spec="+specFile)
	artifactoryCli.Exec("delete", tests.Repo2+"/nonflat_recursive_target/nonflat_recursive_source/a/b/*", "--quiet=true")

	isExistInArtifactory(tests.Delete1, tests.GetFilePath(tests.SearchMoveDeleteRepoSpec), t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFolderWithWildcard(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.MoveCopyDeleteSpec)
	artifactoryCli.Exec("copy", "--spec="+specFile)

	resp, _, _, _ := httputils.SendGet(artifactoryDetails.Url+"api/storage/"+tests.Repo2+"/nonflat_recursive_target/nonflat_recursive_source/a/b/", true, artHttpDetails)
	if resp.StatusCode != 200 {
		t.Error("Missing folder in artifactory : " + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/")
	}

	artifactoryCli.Exec("delete", tests.Repo2+"/nonflat_recursive_target/nonflat_recursive_source/*/b", "--quiet=true")
	resp, _, _, _ = httputils.SendGet(artifactoryDetails.Url+"api/storage/"+tests.Repo2+"/nonflat_recursive_target/nonflat_recursive_source/a/b/", true, artHttpDetails)
	if resp.StatusCode != 404 {
		t.Error("Couldn't delete folder in artifactory : " + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/")
	}

	isExistInArtifactory(tests.Delete1, tests.GetFilePath(tests.SearchMoveDeleteRepoSpec), t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFolder(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	artifactoryCli.Exec("delete", tests.Repo1+"/downloadTestResources", "--quiet=true")
	resp, body, _, err := httputils.SendGet(artifactoryDetails.Url+"api/storage/"+tests.Repo1+"/downloadTestResources", true, artHttpDetails)
	if err != nil || resp.StatusCode != 404 {
		t.Error("Couldn't delete path: " + tests.Repo1 + "/downloadTestResources/ " + string(body))
	}

	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFolderContent(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	artifactoryCli.Exec("delete", tests.Repo1+"/downloadTestResources/", "--quiet=true")

	resp, body, _, err := httputils.SendGet(artifactoryDetails.Url+"api/storage/"+tests.Repo1+"/downloadTestResources", true, artHttpDetails)
	if err != nil || resp.StatusCode != 200 {
		t.Error("downloadTestResources shouldnn't be deleted: " + tests.Repo1 + "/downloadTestResources/ " + string(body))
	}
	folderContent, _, _, err := jsonparser.Get(body, "children")
	if err != nil {
		t.Error("Couldn't parse body:", string(body))
	}
	var folderChildren []struct{}
	err = json.Unmarshal(folderContent, &folderChildren)
	if err != nil {
		t.Error("Couldn't parse body:", string(body))
	}
	if len(folderChildren) != 0 {
		t.Error("downloadTestResources content wasn't deleted")
	}
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFoldersBySpec(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	artifactoryCli.Exec("delete", "--spec="+tests.GetFilePath(tests.DeleteSpec), "--quiet=true")

	resp, body, _, err := httputils.SendGet(artifactoryDetails.Url+"api/storage/"+tests.Repo1+"/downloadTestResources", true, artHttpDetails)
	if err != nil || resp.StatusCode != 404 {
		t.Error("Couldn't delete path: " + tests.Repo1 + "/downloadTestResources/ " + string(body))
	}
	resp, body, _, err = httputils.SendGet(artifactoryDetails.Url+"api/storage/"+tests.Repo2+"/downloadTestResources", true, artHttpDetails)
	if err != nil || resp.StatusCode != 404 {
		t.Error("Couldn't delete path: " + tests.Repo2 + "/downloadTestResources/ " + string(body))
	}

	cleanArtifactoryTest()
}

func TestArtifactoryDeleteExcludeByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Delete by pattern
	artifactoryCli.Exec("del", "jfrog-cli-tests-repo1/data/", "--quiet=true", "--exclude-patterns=*b1.in;*b2.in;*b3.in;*c1.in")

	// Validate files are deleted
	isExistInArtifactory(tests.BuildDeleteExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteExcludeBySpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile := tests.GetFilePath(tests.DelSpecExclude)

	//upload files
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileA)
	artifactoryCli.Exec("upload", "--spec="+specFileB)

	// Delete by pattern
	artifactoryCli.Exec("del", "--spec="+specFile, "--quiet=true")

	// Validate files are deleted
	isExistInArtifactory(tests.BuildDeleteExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDisplyedPathToDelete(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.DeleteComplexSpec)
	artifactsToDelete := getPathsToDelete(specFile)
	var displayedPaths []commands.SearchResult
	for _, v := range artifactsToDelete {
		displayedPaths = append(displayedPaths, commands.SearchResult{Path: v.GetItemRelativePath()})
	}

	tests.CompareExpectedVsActuals(tests.DeleteDisplyedFiles, displayedPaths, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteBySpec(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.DeleteComplexSpec)
	artifactoryCli.Exec("delete", "--spec="+specFile, "--quiet=true")

	artifactsToDelete := getPathsToDelete(specFile)
	if len(artifactsToDelete) != 0 {
		t.Error("Couldn't delete paths")
	}

	cleanArtifactoryTest()
}

func TestArtifactoryMassiveDownloadSpec(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	specFile := tests.GetFilePath(tests.DownloadSpec)
	artifactoryCli.Exec("download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.IsExistLocally(tests.MassiveDownload, paths, t)
	cleanArtifactoryTest()
}

func TestArtifactoryMassiveUploadSpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile := tests.GetFilePath(tests.UploadSpec)
	resultSpecFile := tests.GetFilePath(tests.Search)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	isExistInArtifactory(tests.MassiveUpload, resultSpecFile, t)
	isExistInArtifactoryByProps(tests.PropsExpected, tests.Repo1+"/*/properties/*.in", "searchMe=true", t)
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadRecursiveNonFlat(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	canonicalPath := tests.Out + dirInnerPath

	err := os.MkdirAll(canonicalPath, 0777)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()+"(*)"), tests.Repo1+"/{1}/", "--include-dirs=true", "--recursive=true", "--flat=false")
	err = os.RemoveAll(tests.Out)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--include-dirs=true", "--recursive=true")
	expectedPath := []string{tests.Out, "inner", "folder", "out", "inner", "folder"}
	if !fileutils.IsPathExists(strings.Join(expectedPath, fileutils.GetFileSeparator())) {
		t.Error("Failed to download folders from Artifatory")
	}
	//cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFlatFolderUpload(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	canonicalPath := tests.Out + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()+"(*)"), tests.Repo1+"/{1}/", "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(tests.Out)
	if err != nil {
		t.Error(err.Error())
	}
	// Non flat download
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--include-dirs=true", "--recursive=true")
	if !fileutils.IsPathExists(canonicalPath + fileutils.GetFileSeparator() + "folder") {
		t.Error("Failed to download folders from Artifatory")
	}
	//cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories  - are directories which do not include other directories which match the pattern
func TestArtifactoryIncludeDirFlatNonEmptyFolderUpload(t *testing.T) {
	initArtifactoryTest(t)
	// 'c' folder is defined as bottom chain directory therefor should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.GetTestResourcesPath()+"*"), tests.Repo1, "--include-dirs=true", "--flat=true")
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--include-dirs=true", "--recursive=true")
	if !fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "c") {
		t.Error("Failed to download folders from Artifatory")
	}
	//cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories  - are directories which do not include other directories which match the pattern
func TestArtifactoryDownloadNotIncludeDirs(t *testing.T) {
	initArtifactoryTest(t)
	// 'c' folder is defined as bottom chain directory therefor should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.GetTestResourcesPath()+"*"+fileutils.GetFileSeparator()+"c"), tests.Repo1, "--include-dirs=true", "--flat=true")
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--recursive=true")
	if fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "c") {
		t.Error("Failed to download folders from Artifatory")
	}
	//cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories  - are directories which do not include other directories which match the pattern
func TestArtifactoryDownloadFlatTrue(t *testing.T) {
	initArtifactoryTest(t)
	// 'c' folder is defined as bottom chain directory therefor should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.GetTestResourcesPath()+"(*)"+fileutils.GetFileSeparator()+"*"), tests.Repo1+"/{1}/", "--include-dirs=true", "--flat=true")
	// Download without include-dirs
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--recursive=true", "--flat=true")
	if fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "c") {
		t.Error("'c' folder shouldn't be exist.")
	}
	err := os.RemoveAll(tests.Out)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--include-dirs=true", "--recursive=true", "--flat=true")
	// Inner folder with files in it
	if !fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "c") {
		t.Error("'c' folder should exist.")
	}
	// Empty inner folder
	if !fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "folder") {
		t.Error("'folder' folder should exist.")
	}
	// Folder on root with files
	if !fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "a+a") {
		t.Error("'a+a' folder should be exist.")
	}
	// None bottom directory - shouldn't exist.
	if fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "a") {
		t.Error("'a' folder shouldn't be exist.")
	}
	// None bottom directory - shouldn't exist.
	if fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "b") {
		t.Error("'b' folder shouldn't be exist.")
	}
	//cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryIncludeDirFlatNonEmptyFolderUploadMatchingPattern(t *testing.T) {
	initArtifactoryTest(t)
	// 'c' folder is defined as bottom chain directory therefor should be uploaded when using flat=true even though 'c' is not empty
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.GetTestResourcesPath()+"*"+fileutils.GetFileSeparator()+"c"), tests.Repo1, "--include-dirs=true", "--flat=true")
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--include-dirs=true", "--recursive=true")
	if !fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "c") {
		t.Error("Failed to download folders from Artifatory")
	}
	//cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories  - are directories which do not include other directories which match the pattern
func TestArtifactoryUploadFlatFolderWithFileAndInnerEmptyMatchingPattern(t *testing.T) {
	initArtifactoryTest(t)
	path := tests.FixWinPath(tests.GetTestResourcesPath() + fileutils.GetFileSeparator() + "a" + fileutils.GetFileSeparator() + "b" + fileutils.GetFileSeparator() + "c" + fileutils.GetFileSeparator() + "d")
	err := os.MkdirAll(path, 0777)
	if err != nil {
		t.Error(err.Error())
	}
	// We created a empty child folder to 'c' therefor 'c' is not longer a bottom chain and new 'd' inner directory is indeed bottom chain directory.
	// 'd' should uploaded and 'c' shouldn't
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.GetTestResourcesPath()+"*"), tests.Repo1, "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(path)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--include-dirs=true", "--recursive=true")
	if fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "c") {
		t.Error("'c' folder shouldn't be exsit")
	}
	if !fileutils.IsPathExists(tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()) + "d") {
		t.Error("bottom chian directory, 'd', is missing")
	}
	//cleanup
	cleanArtifactoryTest()
}

// Test the definition of bottom chain directories  - are directories which do not include other directories which match the pattern
func TestArtifactoryUploadFlatFolderWithFileAndInnerEmptyMatchingPatternWithPlaceHolders(t *testing.T) {
	initArtifactoryTest(t)
	separator := fileutils.GetFileSeparator()
	relativePaths := separator + "a" + fileutils.GetFileSeparator() + "b" + fileutils.GetFileSeparator() + "c" + fileutils.GetFileSeparator() + "d"
	path := tests.FixWinPath(tests.GetTestResourcesPath() + relativePaths)
	err := os.MkdirAll(path, 0777)
	if err != nil {
		t.Error(err.Error())
	}
	// We created a empty child folder to 'c' therefor 'c' is not longer a bottom chain and new 'd' inner directory is indeed bottom chain directory.
	// 'd' should uploaded and 'c' shouldn't
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.GetTestResourcesPath()+"(*)"+fileutils.GetFileSeparator()+"*"), tests.Repo1+"/{1}/", "--include-dirs=true", "--flat=true")
	err = os.RemoveAll(path)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--include-dirs=true", "--recursive=true")
	if !fileutils.IsPathExists(tests.FixWinPath(tests.Out + relativePaths)) {
		t.Error("bottom chian directory, 'd', is missing")
	}

	//cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFlatFolderDownload1(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	canonicalPath := tests.Out + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	if err != nil {
		t.Error(err.Error())
	}
	// Flat true by default for upload, by using placeholder we indeed create folders hierarchy in Artifactory inner/folder/folder
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()+"(*)"), tests.Repo1+"/{1}/", "--include-dirs=true")
	err = os.RemoveAll(tests.Out)
	if err != nil {
		t.Error(err.Error())
	}
	// Only the inner folder should be downland e.g 'folder'
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--include-dirs=true", "--flat=true")
	if !fileutils.IsPathExists(tests.Out + fileutils.GetFileSeparator() + "folder") && fileutils.IsPathExists(tests.Out+fileutils.GetFileSeparator()+"inner") {
		t.Error("Failed to download folders from Artifatory")
	}
	//cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadRecursiveUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	dirInnerPath := "empty" + fileutils.GetFileSeparator() + "folder"
	canonicalPath := tests.GetTestResourcesPath() + dirInnerPath
	err := os.MkdirAll(canonicalPath, 0777)
	if err != nil {
		t.Error(err.Error())
	}
	specFile := tests.GetFilePath(tests.UploadEmptyDirs)
	artifactoryCli.Exec("upload", "--spec="+specFile)
	//err = os.RemoveAll(tests.GetTestResourcesPath() + "empty")
	if err != nil {
		t.Error(err.Error())
	}
	specFile = tests.GetFilePath(tests.DownloadEmptyDirs)
	artifactoryCli.Exec("download", "--spec="+specFile)
	if !fileutils.IsPathExists(tests.FixWinPath(tests.Out + fileutils.GetFileSeparator() + "folder")) {
		t.Error("Failed to download folders from Artifatory")
	}
	//cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderUploadNonRecursive(t *testing.T) {
	initArtifactoryTest(t)
	canonicalPath := tests.Out + fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	err := os.MkdirAll(canonicalPath, 0777)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), tests.Repo1, "--recursive=true", "--include-dirs=true")
	err = os.RemoveAll(tests.Out)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("download", tests.Repo1, tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), "--include-dirs=true")
	if !fileutils.IsPathExists(tests.FixWinPath(tests.Out + fileutils.GetFileSeparator() + "folder")) {
		t.Error("Failed to download folder from Artifatory")
	}
	if fileutils.IsPathExists(canonicalPath) {
		t.Error("Path should be flat ")
	}
	//cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryFolderDownloadNonRecursive(t *testing.T) {
	initArtifactoryTest(t)
	canonicalPath := tests.Out + fileutils.GetFileSeparator() + "inner" + fileutils.GetFileSeparator() + "folder"
	err := os.MkdirAll(canonicalPath, 0777)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("upload", tests.FixWinPath(tests.Out+fileutils.GetFileSeparator()), tests.Repo1, "--recursive=true", "--include-dirs=true", "--flat=false")
	err = os.RemoveAll(tests.Out)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("download", tests.Repo1+"/*", "--recursive=false", "--include-dirs=true")
	if !fileutils.IsPathExists(tests.Out) {
		t.Error("Failed to download folder from Artifatory")
	}
	if fileutils.IsPathExists(canonicalPath) {
		t.Error("Path should be flat. ")
	}
	//cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryChecksumDownload(t *testing.T) {
	initArtifactoryTest(t)

	var filePath = "../testsdata/a/a1.in"
	artifactoryCli.Exec("upload", filePath, tests.Repo1)
	testChecksumDownload(t, "/a1.in")
	//cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryChecksumDownloadRenameFileName(t *testing.T) {
	initArtifactoryTest(t)

	var filePath = "../testsdata/a/a1.in"
	artifactoryCli.Exec("upload", filePath, tests.Repo1)
	testChecksumDownload(t, "/a1.out")
	//cleanup
	cleanArtifactoryTest()
}

func testChecksumDownload(t *testing.T, outFileName string) {
	artifactoryCli.Exec("download jfrog-cli-tests-repo1/a1.in", tests.Out+outFileName)

	exists, err := fileutils.IsFileExists(tests.Out + outFileName)
	if err != nil {
		t.Error(err.Error())
	}
	if !exists {
		t.Error("Failed to download file from Artifatory")
	}

	firstFileInfo, _ := os.Stat(tests.Out + outFileName)
	firstDownloadTime := firstFileInfo.ModTime()

	artifactoryCli.Exec("download jfrog-cli-tests-repo1/a1.in", tests.Out+outFileName)
	secondFileInfo, _ := os.Stat(tests.Out + outFileName)
	secondDownloadTime := secondFileInfo.ModTime()

	if !firstDownloadTime.Equal(secondDownloadTime) {
		t.Error("Checksum download failed, the file was downloaded twice")
	}
}

func TestArtifactoryPublishBuildInfo(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumber := "cli-test-build", "10"

	//upload files with buildName and buildNumber
	specFile := tests.GetFilePath(tests.SimpleUploadSpec)
	artifactoryCli.Exec("upload", "--spec="+specFile, "--build-name="+buildName, "--build-number="+buildNumber)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	//validate files are uploaded with the build info name and number
	props := fmt.Sprintf("build.name=%v;build.number=%v", buildName, buildNumber)
	isExistInArtifactoryByProps(tests.SimpleUploadExpectedRepo1, tests.Repo1+"/*", props, t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.BuildDownloadSpec)

	// Upload files with buildName and buildNumber
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberB)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.BuildDownload, paths, t)

	// Cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

// Upload a file to build A.
// Verify that it doesn't exist in B.
func TestArtifactoryDownloadArtifactDoesntExistInBuild(t *testing.T) {
	initArtifactoryTest(t)
	buildNameA, buildNumberA := "cli-test-build1", "10"
	specFile := tests.GetFilePath(tests.BuildDownloadSpecNoBuildNumber)

	// Upload a file
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a10.in", "--build-name="+buildNameA, "--build-number="+buildNumberA)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildNameA, buildNumberA)

	// Download from different build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.BuildDownloadDoesntExist, paths, t)

	// Cleanup
	deleteBuild(buildNameA)
	cleanArtifactoryTest()
}

// Upload a file to 2 different builds.
// Verify that we don't download files with same sha and different build name and build number.
func TestArtifactoryDownloadByShaAndBuild(t *testing.T) {
	initArtifactoryTest(t)
	buildNameA, buildNameB, buildNumberA, buildNumberB, buildNumberC := "cli-test-build1", "cli-test-build2", "10", "11", "12"
	specFile := tests.GetFilePath(tests.BuildDownloadSpecNoBuildNumber)

	// Upload 3 similar files to 3 different builds
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a10.in", "--build-name="+buildNameB, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a11.in", "--build-name="+buildNameA, "--build-number="+buildNumberB)
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a12.in", "--build-name="+buildNameA, "--build-number="+buildNumberC)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberA)
	artifactoryCli.Exec("build-publish", buildNameA, buildNumberB)
	artifactoryCli.Exec("build-publish", buildNameA, buildNumberC)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.BuildDownloadByShaAndBuild, paths, t)

	// Cleanup
	deleteBuild(buildNameA)
	deleteBuild(buildNameB)
	cleanArtifactoryTest()
}

// Upload a file to 2 different builds.
// Verify that we don't download files with same sha and build name and different build number.
func TestArtifactoryDownloadByShaAndBuildName(t *testing.T) {
	initArtifactoryTest(t)
	buildNameA, buildNameB, buildNumberA, buildNumberB, buildNumberC := "cli-test-build1", "cli-test-build2", "10", "11", "12"
	specFile := tests.GetFilePath(tests.BuildDownloadSpecNoBuildNumber)

	// Upload 3 similar files to 2 different builds
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a10.in", "--build-name="+buildNameB, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a11.in", "--build-name="+buildNameB, "--build-number="+buildNumberB)
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a12.in", "--build-name="+buildNameA, "--build-number="+buildNumberC)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildNameA, buildNumberC)
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberA)
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--spec="+specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.BuildDownloadByShaAndBuildName, paths, t)

	// Cleanup
	deleteBuild(buildNameA)
	deleteBuild(buildNameB)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildUsingSimpleDownload(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"

	//upload files with buildName and buildNumber
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Download by build number, a1 should not be downloaded, b1 should
	artifactoryCli.Exec("download jfrog-cli-tests-repo1/data/a1.in "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--build="+buildName)
	artifactoryCli.Exec("download jfrog-cli-tests-repo1/data/b1.in "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--build="+buildName)

	//validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.BuildSimpleDownload, paths, t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExcludeByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--recursive=true")
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--recursive=true")

	// Download by pattern
	artifactoryCli.Exec("download", "jfrog-cli-tests-repo1 out/download/aql_by_artifacts/", "--exclude-patterns=*/a1.in;*a2.*;data/c2.in")

	// Validate files are excluded
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.BuildExcludeDownload, paths, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExcludeBySpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile := tests.GetFilePath(tests.DownloadSpecExclude)

	// Upload files
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--recursive=true", "--flat=false")
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--recursive=true", "--flat=false")

	// Download by spec
	artifactoryCli.Exec("download", "--spec="+specFile)

	// Validate files are excluded
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.BuildExcludeDownloadBySpec, paths, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadExcludeBySpecOverride(t *testing.T) {
	initArtifactoryTest(t)

	////upload files with buildName and buildNumber
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--recursive=true", "--flat=false")
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--recursive=true", "--flat=false")

	// Download by spec
	specFile := tests.GetFilePath(tests.DownloadSpecExclude)
	artifactoryCli.Exec("download", "--spec="+specFile, "--exclude-patterns=*a1.in;*a2.in;*c2.in")

	// Validate files are downloaded by build number
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.BuildExcludeDownload, paths, t)

	// Cleanup
	cleanArtifactoryTest()
}

//Sort and limit changes the way properties are used so this should be tested with symlinks and search by build

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded with limit param.
// validate the symlink content checksum.
func TestArtifactoryLimitWithSimlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath() + "a/", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath() + "a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("u", link + " " + tests.Repo1 + " --symlinks=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1 + "/link " + tests.GetTestResourcesPath() + "a/ --validate-symlinks=true --limit=1")
	validateSortLimitWithSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded with limit param.
// validate the symlink content checksum.
func TestArtifactorySortWithSimlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath() + "a/", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath() + "a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("u", link + " " + tests.Repo1 + " --symlinks=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1+"/link " + tests.GetTestResourcesPath() + "a/ --validate-symlinks=true --sort-by=created")
	validateSortLimitWithSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

func validateSortLimitWithSymLink(localLinkPath, localFilePath string, t *testing.T) {
	exists := fileutils.IsPathSymlink(localLinkPath)
	if !exists {
		t.Error(errors.New("Faild to download symlinks from artifactory with Sort/Limit flag"))
	}
	symlinks, err := filepath.EvalSymlinks(localLinkPath)
	if err != nil {
		t.Error(errors.New("Can't eval symlinks with Sort/Limit flag"))
	}
	if symlinks != localFilePath {
		t.Error(errors.New("Symlinks wasn't created as expected with Sort/Limit flag. expected:" + localFilePath + " actual: " + symlinks))
	}
}

// Upload a file to 2 different builds.
// Verify that we don't download files with same sha and build name and different build number when sort is configured.
func TestArtifactoryDownloadByShaAndBuildNameWithSort(t *testing.T) {
	initArtifactoryTest(t)
	buildNameA, buildNameB, buildNumberA, buildNumberB, buildNumberC := "cli-test-build1", "cli-test-build2", "10", "11", "12"
	specFile := tests.GetFilePath(tests.BuildDownloadSpecNoBuildNumberWithSort)

	// Upload 3 similar files to 2 different builds
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a10.in", "--build-name=" + buildNameB, "--build-number=" + buildNumberA)
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a11.in", "--build-name=" + buildNameB, "--build-number=" + buildNumberB)
	artifactoryCli.Exec("upload", "../testsdata/a/a1.in", "jfrog-cli-tests-repo1/data/a12.in", "--build-name=" + buildNameA, "--build-number=" + buildNumberC)

	// Publish buildInfo
	artifactoryCli.Exec("build-publish", buildNameA, buildNumberC)
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberA)
	artifactoryCli.Exec("build-publish", buildNameB, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--sort-by=created --spec=" + specFile)

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.BuildDownloadByShaAndBuildNameWithSort, paths, t)

	// Cleanup
	deleteBuild(buildNameA)
	deleteBuild(buildNameB)
	cleanArtifactoryTest()
}

func TestArtifactorySortAndLimit(t *testing.T) {
	initArtifactoryTest(t)

	// Upload all testdata/a/ files
	artifactoryCli.Exec("upload", "../testsdata/a/(*)", "jfrog-cli-tests-repo1/data/{1}")

	// Download 1 sorted by name asc
	artifactoryCli.Exec("download", "jfrog-cli-tests-repo1/data/ out/download/sort_limit/", "--sort-by=name", "--limit=1")

	// Download 3 sorted by depth desc
	artifactoryCli.Exec("download", "jfrog-cli-tests-repo1/data/ out/download/sort_limit/", "--sort-by=depth", "--limit=3", "--sort-order=desc")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.AreListsIdentical(tests.SortAndLimit, paths, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.CopyByBuildSpec)

	// Upload files with buildName and buildNumber: a* uploaded with build number "10", b* uploaded with build number "11"
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Copy by build build number "10" from spec, a* should be copied
	artifactoryCli.Exec("copy", "--spec="+specFile)

	//validate files are Copied by build number
	isExistInArtifactory(tests.BuildCopyExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildOverridingByInlineFlag(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.CopyByBuildSpec)

	// Upload files with buildName and buildNumber: b* uploaded with build number "10", a* uploaded with build number "11"
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Copy by build number: using override of build by flag from inline (no number set so LATEST build should be copied), a* should be copied
	artifactoryCli.Exec("copy", "--build="+buildName+" --spec="+specFile)

	//validate files are Copied by build number
	isExistInArtifactory(tests.BuildCopyExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryMoveByBuildUsingFlags(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.CopyByBuildSpec)

	//upload files with buildName and buildNumber
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec="+specFileB, "--build-name="+buildName, "--build-number="+buildNumberA)
	artifactoryCli.Exec("upload", "--spec="+specFileA, "--build-name="+buildName, "--build-number="+buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Move by build name and number
	artifactoryCli.Exec("move", "--build="+buildName+"/11 --spec="+specFile)

	//validate files are moved by build number
	isExistInArtifactory(tests.BuildMoveExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryMoveExcludeByCli(t *testing.T) {
	initArtifactoryTest(t)

	// Upload files
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileA)
	artifactoryCli.Exec("upload", "--spec=" + specFileB)

	// Move by pattern
	artifactoryCli.Exec("move", "jfrog-cli-tests-repo1/data/ jfrog-cli-tests-repo2/", "--exclude-patterns=*b*;*c*")

	// Validate excluded files didn't move
	isExistInArtifactory(tests.BuildMoveExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryMoveExcludeBySpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile := tests.GetFilePath(tests.MoveCopySpecExclude)

	// Upload files
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileA)
	artifactoryCli.Exec("upload", "--spec=" + specFileB)

	// Move by spec
	artifactoryCli.Exec("move", "--spec=" + specFile)

	// Validate excluded files didn't move
	isExistInArtifactory(tests.BuildMoveExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteByLatestBuild(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.CopyByBuildSpec)

	//upload files with buildName and buildNumber
	specFileA := tests.GetFilePath(tests.SplitUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplitUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileB, "--build-name=" + buildName, "--build-number=" + buildNumberA)
	artifactoryCli.Exec("upload", "--spec=" + specFileA, "--build-name=" + buildName, "--build-number=" + buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Delete by build name and LATEST
	artifactoryCli.Exec("delete", "--build="+buildName+"/LATEST --quiet=true --spec="+specFile)

	//validate files are deleted by build number
	isExistInArtifactory(tests.BuildDeleteExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestGitLfsCleanup(t *testing.T) {
	initArtifactoryTest(t)
	var filePath = "../testsdata/gitlfs/(4b)(*)"
	if runtime.GOOS == "windows" {
		filePath = tests.FixWinPath("..\\testsdata\\gitlfs\\(4b)(*)")
	}
	artifactoryCli.Exec("upload", filePath, tests.Lfs_Repo+"/objects/4b/f4/{2}{1}")
	artifactoryCli.Exec("upload", filePath, tests.Lfs_Repo+"/objects/4b/f4/")
	separator := "/"
	if runtime.GOOS == "windows" {
		separator = "\\"
	}
	refs := strings.Join([]string{"refs", "heads", "*"}, separator)
	dotGitPath := getCliDotGitPath(t)
	artifactoryCli.Exec("glc", dotGitPath, "--repo="+tests.Lfs_Repo, "--refs="+refs, "--quiet=true")
	isExistInArtifactory(tests.GitLfsExpected, tests.GetFilePath(tests.GitLfsAssertSpec), t)
	cleanArtifactoryTest()
}

func TestArtifactoryCleanBuildInfo(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumber := "cli-test-build", "11"

	//upload files with buildName and buildNumber
	specFile := tests.GetFilePath(tests.UploadSpec)
	artifactoryCli.Exec("upload", "--spec="+specFile, "--build-name="+buildName, "--build-number="+buildNumber)

	//cleanup buildInfo
	artifactoryCli.WithSuffix("").Exec("build-clean", buildName, buildNumber)

	//upload files with buildName and buildNumber
	specFile = tests.GetFilePath(tests.SimpleUploadSpec)
	artifactoryCli.Exec("upload", "--spec="+specFile, "--build-name="+buildName, "--build-number="+buildNumber)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	//promote buildInfo
	artifactoryCli.Exec("build-promote", buildName, buildNumber, tests.Repo2)

	//validate files are uploaded with the build info name and number
	props := fmt.Sprintf("build.name=%v;build.number=%v", buildName, buildNumber)
	isExistInArtifactoryByProps(tests.SimpleUploadExpectedRepo2, tests.Repo2+"/*", props, t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestCollectGitBuildInfo(t *testing.T) {
	initArtifactoryTest(t)
	gitCollectCliRunner := tests.NewJfrogCli(main, "jfrog rt", "")
	buildName, buildNumber := "cli-test-build", "13"
	dotGitPath := tests.FixWinPath(getCliDotGitPath(t))
	gitCollectCliRunner.Exec("build-add-git", buildName, buildNumber, dotGitPath)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	_, body, _, err := httputils.SendGet(artifactoryDetails.Url+"api/build/"+buildName+"/"+buildNumber, false, artHttpDetails)
	if err != nil {
		t.Error(err)
	}
	buildInfoVcsRevision, err := jsonparser.GetString(body, "buildInfo", "vcsRevision")
	if err != nil {
		t.Error(err)
	}
	buildInfoVcsUrl, err := jsonparser.GetString(body, "buildInfo", "vcsUrl")
	if err != nil {
		t.Error(err)
	}
	if buildInfoVcsRevision == "" {
		t.Error("Failed to get git revision.")
	}

	if buildInfoVcsUrl == "" {
		t.Error("Failed to get git remote url.")
	}

	gitManager := utils.NewGitManager(dotGitPath)
	if err = gitManager.ReadGitConfig(); err != nil {
		t.Error("Failed to read .git config file.")
	}
	if gitManager.GetRevision() != buildInfoVcsRevision {
		t.Error("Wrong revision", "expected: "+gitManager.GetRevision(), "Got: "+buildInfoVcsRevision)
	}

	gitConfigUrl := gitManager.GetUrl() + ".git"
	if gitConfigUrl != buildInfoVcsUrl {
		t.Error("Wrong url", "expected: "+gitConfigUrl, "Got: "+buildInfoVcsUrl)
	}

	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestReadGitConfig(t *testing.T) {
	dotGitPath := getCliDotGitPath(t)
	gitManager := utils.NewGitManager(dotGitPath)
	err := gitManager.ReadGitConfig()
	if err != nil {
		t.Error("Failed to read .git config file.")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get current dir.")
	}
	gitExecutor := tests.GitExecutor(workingDir)
	revision, _, err := gitExecutor.GetRevision()
	if err != nil {
		t.Error(err)
		return
	}

	if gitManager.GetRevision() != revision {
		t.Error("Wrong revision", "expected: "+revision, "Got: "+gitManager.GetRevision())
	}

	url, _, err := gitExecutor.GetUrl()
	if err != nil {
		t.Error(err)
		return
	}

	if gitManager.GetUrl() != url {
		t.Error("Wrong revision", "expected: "+url, "Got: "+gitManager.GetUrl())
	}
}

func createJfrogHomeConfig(t *testing.T) {
	templateConfigPath := filepath.Join(tests.GetTestResourcesPath(), "configtemplate", config.JFROG_CONFIG_FILE)

	err := os.Setenv(config.JFROG_HOME_ENV, filepath.Join(tests.Out, "jfroghome"))
	jfrogHomePath, err := config.GetJfrogHomeDir()
	if err != nil {
		t.Error(err)
	}
	_, err = copyTemplateFile(templateConfigPath, jfrogHomePath, config.JFROG_CONFIG_FILE, true)
	if err != nil {
		t.Error(err)
	}
	if err != nil {
		t.Error(err)
	}
}

func CleanArtifactoryTests() {
	cleanArtifactoryTest()
	if err := deleteRepos(); err != nil {
		log.Error(err)
	}
}

func initArtifactoryTest(t *testing.T) {
	if !*tests.TestArtifactory {
		t.Skip("Artifactory is not being tested, skipping...")
	}
}

func cleanArtifactoryTest() {
	if !*tests.TestArtifactory {
		return
	}
	os.Unsetenv(config.JFROG_HOME_ENV)
	cleanArtifactory()
	tests.CleanFileSystem()
}

func copyTemplateFile(srcFile, destPath, destFileName string, replaceCredentials bool) (string, error) {
	content, err := fileutils.ReadFile(srcFile)
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(destPath, 0777)
	if err != nil {
		return "", err
	}

	if replaceCredentials {
		content = bytes.Replace(content, []byte("${RT_URL}"), []byte(*tests.RtUrl), -1)
		content = bytes.Replace(content, []byte("${RT_API_KEY}"), []byte(*tests.RtApiKey), -1)
		content = bytes.Replace(content, []byte("${RT_USERNAME}"), []byte(*tests.RtUser), -1)
		content = bytes.Replace(content, []byte("${RT_PASSWORD}"), []byte(*tests.RtPassword), -1)
	}

	destFile := filepath.Join(destPath, destFileName)
	err = ioutil.WriteFile(destFile, content, 0644)
	if err != nil {
		return "", err
	}
	return destFile, nil
}

func prepUploadFiles() {
	uploadPath := tests.FixWinPath(tests.GetTestResourcesPath()) + "(.*)"
	targetPath := tests.Repo1 + "/downloadTestResources/{1}"
	flags := "--threads=10 --regexp=true --props=searchMe=true --flat=false"
	artifactoryCli.Exec("upload", uploadPath, targetPath, flags)
}

func prepCopyFiles() {
	specFile := tests.GetFilePath(tests.PrepareCopy)
	artifactoryCli.Exec("copy", "--spec="+specFile)
}

func getPathsToDelete(specFile string) []rtutils.ResultItem {
	deleteSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	artifactsToDelete, _ := commands.GetPathsToDelete(deleteSpec, &commands.DeleteConfiguration{ArtDetails: artifactoryDetails})
	return artifactsToDelete
}

func deleteBuild(buildName string) {
	resp, body, err := httputils.SendDelete(artifactoryDetails.Url+"api/build/"+buildName+"?deleteAll=1", nil, artHttpDetails)
	if err != nil {
		log.Error(err)
	}
	if resp.StatusCode != 200 {
		log.Error(resp.Status)
		log.Error(string(body))
	}
}

func execDeleteRepoRest(repoName string) error {
	resp, body, err := httputils.SendDelete(artifactoryDetails.Url+"api/repositories/"+repoName, nil, artHttpDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body))
	}
	log.Info("Repository", repoName, "was deleted.")
	return nil
}

func execCreateRepoRest(repoConfig, repoName string) error {
	content, err := ioutil.ReadFile(repoConfig)
	if err != nil {
		return err
	}
	rtutils.AddHeader("Content-Type", "application/json", &artHttpDetails.Headers)
	resp, body, err := httputils.SendPut(artifactoryDetails.Url+"api/repositories/"+repoName, content, artHttpDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body))
	}
	log.Info("Repository", repoName, "was created.")
	return nil
}

func createReposIfNeeded() error {
	repos := map[string]string{
		tests.Repo1:             tests.SpecsTestRepositoryConfig,
		tests.Repo2:             tests.MoveRepositoryConfig,
		tests.Lfs_Repo:          tests.GitLfsTestRepositoryConfig,
		tests.JcenterRemoteRepo: tests.JcenterRemoteRepositoryConfig,
	}
	for repoName, configFile := range repos {
		if !isRepoExist(repoName) {
			repoConfig := tests.GetTestResourcesPath() + configFile
			err := execCreateRepoRest(repoConfig, repoName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteRepos() error {
	repos := []string{
		tests.Repo1,
		tests.Repo2,
		tests.Lfs_Repo,
		tests.JcenterRemoteRepo,
	}
	for _, repoName := range repos {
		if isRepoExist(repoName) {
			err := execDeleteRepoRest(repoName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func cleanArtifactory() {
	deleteFlags := new(commands.DeleteConfiguration)
	deleteSpec, _ := spec.CreateSpecFromFile(tests.GetFilePath(tests.DeleteSpec), nil)
	deleteFlags.ArtDetails = artifactoryDetails
	commands.Delete(deleteSpec, deleteFlags)
}

func searchInArtifactory(specFile string) (result []commands.SearchResult, err error) {
	searchSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	result, err = commands.Search(searchSpec, artifactoryDetails)
	return
}

func getSpecAndCommonFlags(specFile string) (*spec.SpecFiles, rtutils.CommonConf) {
	searchFlags := new(rtutils.CommonConfImpl)
	searchSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	return searchSpec, searchFlags
}

func isExistInArtifactory(expected []string, specFile string, t *testing.T) {
	results, _ := searchInArtifactory(specFile)
	if *tests.PrintSearchResult {
		for _, v := range results {
			fmt.Print("\"")
			fmt.Print(v.Path)
			fmt.Print("\"")
			fmt.Print(",")
			fmt.Println("")
		}
	}
	tests.CompareExpectedVsActuals(expected, results, t)
}

func isExistInArtifactoryByProps(expected []string, pattern, props string, t *testing.T) {
	searchSpec := spec.NewBuilder().Pattern(pattern).Props(props).Recursive(true).BuildSpec()
	results, err := commands.Search(searchSpec, artifactoryDetails)
	if err != nil {
		t.Error(err)
	}
	tests.CompareExpectedVsActuals(expected, results, t)
}

func isRepoExist(repoName string) bool {
	resp, _, _, err := httputils.SendGet(artifactoryDetails.Url+tests.RepoDetailsUrl+repoName, true, artHttpDetails)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if resp.StatusCode != 400 {
		return true
	}
	return false
}

func getCliDotGitPath(t *testing.T) string {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get current dir.")
	}
	dotGitPath := filepath.Join(workingDir, "..", "..")
	dotGitExists, err := fileutils.IsDirExists(filepath.Join(dotGitPath, ".git"))
	if err != nil {
		t.Error(err)
	}
	if !dotGitExists {
		t.Error("Can't find .git")
	}
	return dotGitPath
}
