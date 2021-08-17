package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	xrutils "github.com/jfrog/jfrog-cli-core/v2/xray/utils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/version"
	"github.com/jfrog/jfrog-client-go/xray/services"
	"github.com/stretchr/testify/assert"
)

const (
	xrayEndpoint = "xray/"
)

var (
	xrayDetails     *config.ServerDetails
	xrayAuth        auth.ServiceDetails
	xrayHttpDetails httputils.HttpClientDetails
	// JFrog CLI for Xray commands
	xrayCli *tests.JfrogCli
)

func InitXrayTests() {
	initXrayCli()
}

func authenticateXray() string {
	xrayDetails = &config.ServerDetails{XrayUrl: clientutils.AddTrailingSlashIfNeeded(*tests.JfrogUrl) + xrayEndpoint}
	cred := "--url=" + xrayDetails.XrayUrl
	if *tests.JfrogAccessToken != "" {
		xrayDetails.AccessToken = *tests.JfrogAccessToken
		cred += fmt.Sprintf(" --access-token=%q", xrayDetails.AccessToken)
	} else {
		xrayDetails.User = *tests.JfrogUser
		xrayDetails.Password = *tests.JfrogPassword
		cred += fmt.Sprintf(" --user=%q --password=%q", xrayDetails.User, xrayDetails.Password)
	}

	var err error
	if xrayAuth, err = xrayDetails.CreateXrayAuthConfig(); err != nil {
		coreutils.ExitOnErr(errors.New("Failed while attempting to authenticate with Xray: " + err.Error()))
	}
	xrayDetails.XrayUrl = xrayAuth.GetUrl()
	xrayHttpDetails = xrayAuth.CreateHttpClientDetails()
	return cred
}

func initXrayCli() {
	if xrayCli != nil {
		return
	}
	cred := authenticateXray()
	xrayCli = tests.NewJfrogCli(execMain, "jfrog xr", cred)
}

// Tests basic binary scan by providing pattern (path to testdata binaries) and --licenses flag
// and asserts any error.
func TestXrayBinaryScan(t *testing.T) {
	initXrayTest(t, xrutils.GraphScanMinVersion)
	newStdout, stdWriter, previousStdout := tests.RedirectStdOutToPipe()
	// Restore previous stdout when the function returns
	defer func() {
		os.Stdout = previousStdout
	}()

	binariesPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "binaries", "*")
	runXr(t, "scan", "--licenses", "--format=json", binariesPath)
	// Closing the temp stdout in order to be able to read it's content.
	stdWriter.Close()
	verifyScanResults(t, newStdout, 0, 1, 1, previousStdout)
}

// Tests npm audit by providing simple npm project and asserts any error.
func TestXrayAuditNpm(t *testing.T) {
	initXrayTest(t, xrutils.GraphScanMinVersion)
	newStdout, stdWriter, previousStdout := tests.RedirectStdOutToPipe()
	// Restore previous stdout when the function returns
	defer func() {
		os.Stdout = previousStdout
	}()
	tempDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	defer tests.RemoveTempDirAndAssert(t, tempDirPath)
	npmProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "npm")
	// Copy the npm project from the testdata to a temp dir
	assert.NoError(t, fileutils.CopyDir(npmProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer tests.ChangeDirAndAssert(t, prevWd)
	// Run npm install before executing jfrog xr npm-audit
	assert.NoError(t, exec.Command("npm", "install").Run())
	runXr(t, "audit-npm", "--licenses", "--format=json")
	// Closing the temp stdout in order to be able to read it's content.
	stdWriter.Close()
	verifyScanResults(t, newStdout, 0, 1, 1, previousStdout)
}

func TestXrayAuditGradle(t *testing.T) {
	initXrayTest(t, xrutils.GraphScanMinVersion)
	newStdout, stdWriter, previousStdout := tests.RedirectStdOutToPipe()
	// Restore previous stdout when the function returns
	defer func() {
		os.Stdout = previousStdout
	}()
	tempDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	defer tests.RemoveTempDirAndAssert(t, tempDirPath)
	gradleProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "gradle")
	// Copy the gradle project from the testdata to a temp dir
	assert.NoError(t, fileutils.CopyDir(gradleProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer tests.ChangeDirAndAssert(t, prevWd)
	runXr(t, "audit-gradle", "--licenses", "--format=json")
	// Closing the temp stdout in order to be able to read it's content.
	stdWriter.Close()
	verifyScanResults(t, newStdout, 0, 0, 0, previousStdout)
}

func TestXrayAuditMaven(t *testing.T) {
	initXrayTest(t, xrutils.GraphScanMinVersion)
	newStdout, stdWriter, previousStdout := tests.RedirectStdOutToPipe()
	// Restore previous stdout when the function returns
	defer func() {
		os.Stdout = previousStdout
	}()
	tempDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	defer tests.RemoveTempDirAndAssert(t, tempDirPath)
	mvnProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "maven")
	// Copy the maven project from the testdata to a temp dir
	assert.NoError(t, fileutils.CopyDir(mvnProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer tests.ChangeDirAndAssert(t, prevWd)
	runXr(t, "audit-mvn", "--licenses", "--format=json")
	// Closing the temp stdout in order to be able to read it's content.
	stdWriter.Close()
	verifyScanResults(t, newStdout, 0, 1, 1, previousStdout)
}

func initXrayTest(t *testing.T, minVersion ...string) {
	if !*tests.TestXray {
		t.Skip("Skipping Xray test. To run Xray test add the '-test.xray=true' option.")
	}
	xrayVersion, err := getXrayVersion()
	if err != nil {
		assert.NoError(t, err)
		return
	}
	// If minimal version was supplied, make sure the Xray version fulfil the minimum version requirement
	if len(minVersion) > 0 && !xrayVersion.AtLeast(minVersion[0]) {
		t.Skip(fmt.Sprintf("Skipping Xray test. You are using Xray %s, while  this test requires Xray version %s or higher.", xrayVersion, minVersion))
	}
}

func getXrayVersion() (version.Version, error) {
	xrayVersion, err := xrayAuth.GetVersion()
	return *version.NewVersion(xrayVersion), err
}

// Run `jfrog xr` command
func runXr(t *testing.T, args ...string) {
	err := xrayCli.Exec(args...)
	assert.NoError(t, err)
}

func verifyScanResults(t *testing.T, pipe *os.File, minViolations, minVulnerabilities, minLicenses int, stdout *os.File) {
	content, err := ioutil.ReadAll(pipe)
	assert.NoError(t, err)
	// Prints the redirected output to the standart output as well.
	stdout.Write(content)

	var results []services.ScanResponse
	err = json.Unmarshal(content, &results)
	assert.NoError(t, err)
	assert.True(t, len(results[0].Violations) >= minViolations, fmt.Sprintf("Expected at least %d violations in scan results, but got %d violations.", minViolations, len(results[0].Violations)))
	assert.True(t, len(results[0].Vulnerabilities) >= minVulnerabilities, fmt.Sprintf("Expected at least %d vulnerabilities in scan results, but got %d vulnerabilities.", minVulnerabilities, len(results[0].Vulnerabilities)))
	assert.True(t, len(results[0].Licenses) >= minLicenses, fmt.Sprintf("Expected at least %d Licenses in scan results, but got %d Licenses.", minLicenses, len(results[0].Licenses)))
}
