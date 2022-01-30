package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/gofrog/version"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/xray/services"
	"github.com/stretchr/testify/assert"
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
	*tests.JfrogUrl = clientutils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	xrayDetails = &config.ServerDetails{XrayUrl: *tests.JfrogUrl + tests.XrayEndpoint}
	cred := fmt.Sprintf("--url=%s", xrayDetails.XrayUrl)
	if *tests.JfrogAccessToken != "" {
		xrayDetails.AccessToken = *tests.JfrogAccessToken
		cred += fmt.Sprintf(" --access-token=%s", xrayDetails.AccessToken)
	} else {
		xrayDetails.User = *tests.JfrogUser
		xrayDetails.Password = *tests.JfrogPassword
		cred += fmt.Sprintf(" --user=%s --password=%s", xrayDetails.User, xrayDetails.Password)
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
	xrayCli = tests.NewJfrogCli(execMain, "jfrog", cred)
}

// Tests basic binary scan by providing pattern (path to testdata binaries) and --licenses flag
// and asserts any error.
func TestXrayBinaryScan(t *testing.T) {
	initXrayTest(t, commands.GraphScanMinXrayVersion)
	binariesPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "binaries", "*")
	output := xrayCli.RunCliCmdWithOutput(t, "scan", binariesPath, "--licenses", "--format=json")
	verifyScanResults(t, output, 0, 1, 1)
}

// Tests npm audit by providing simple npm project and asserts any error.
func TestXrayAuditNpm(t *testing.T) {
	initXrayTest(t, commands.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	npmProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "npm")
	// Copy the npm project from the testdata to a temp dir
	assert.NoError(t, fileutils.CopyDir(npmProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Run npm install before executing jfrog xr npm-audit
	assert.NoError(t, exec.Command("npm", "install").Run())

	output := xrayCli.RunCliCmdWithOutput(t, "audit-npm", "--licenses", "--format=json")
	verifyScanResults(t, output, 0, 1, 1)
}

func TestXrayAuditGradle(t *testing.T) {
	initXrayTest(t, commands.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	gradleProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "gradle")
	// Copy the gradle project from the testdata to a temp dir
	assert.NoError(t, fileutils.CopyDir(gradleProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)

	output := xrayCli.RunCliCmdWithOutput(t, "audit-gradle", "--licenses", "--format=json")
	verifyScanResults(t, output, 0, 0, 0)
}

func TestXrayAuditMaven(t *testing.T) {
	initXrayTest(t, commands.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	mvnProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "maven")
	// Copy the maven project from the testdata to a temp dir
	assert.NoError(t, fileutils.CopyDir(mvnProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	output := xrayCli.RunCliCmdWithOutput(t, "audit-mvn", "--licenses", "--format=json")
	verifyScanResults(t, output, 0, 1, 1)
}

func initXrayTest(t *testing.T, minVersion string) {
	if !*tests.TestXray {
		t.Skip("Skipping Xray test. To run Xray test add the '-test.xray=true' option.")
	}
	validateXrayVersion(t, minVersion)
}

func validateXrayVersion(t *testing.T, minVersion string) {
	xrayVersion, err := getXrayVersion()
	if err != nil {
		assert.NoError(t, err)
		return
	}
	err = commands.ValidateXrayMinimumVersion(xrayVersion.GetVersion(), minVersion)
	if err != nil {
		t.Skip(err)
	}
}

func getXrayVersion() (version.Version, error) {
	xrayVersion, err := xrayAuth.GetVersion()
	return *version.NewVersion(xrayVersion), err
}

func verifyScanResults(t *testing.T, content string, minViolations, minVulnerabilities, minLicenses int) {
	var results []services.ScanResponse
	err := json.Unmarshal([]byte(content), &results)
	assert.NoError(t, err)
	assert.True(t, len(results[0].Violations) >= minViolations, fmt.Sprintf("Expected at least %d violations in scan results, but got %d violations.", minViolations, len(results[0].Violations)))
	assert.True(t, len(results[0].Vulnerabilities) >= minVulnerabilities, fmt.Sprintf("Expected at least %d vulnerabilities in scan results, but got %d vulnerabilities.", minVulnerabilities, len(results[0].Vulnerabilities)))
	assert.True(t, len(results[0].Licenses) >= minLicenses, fmt.Sprintf("Expected at least %d Licenses in scan results, but got %d Licenses.", minLicenses, len(results[0].Licenses)))
}
