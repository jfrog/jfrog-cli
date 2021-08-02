package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	xrutils "github.com/jfrog/jfrog-cli-core/v2/xray/utils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
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
	// Close the temp stdout in order to read it.
	stdWriter.Close()
	verifyScanResults(t, newStdout, 0, 22, 5, previousStdout)
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

func verifyScanResults(t *testing.T, pipe *os.File, violations, vulnerabilities, licenses int, stdout *os.File) {
	content, err := ioutil.ReadAll(pipe)
	assert.NoError(t, err)
	// Prints the redirected output to the standart output as well.
	stdout.Write(content)

	var results []services.ScanResponse
	err = json.Unmarshal(content, &results)
	assert.NoError(t, err)
	assert.Equal(t, len(results[0].Violations), violations, fmt.Sprintf("Expected %d violations in scan results, but got %d violations.", violations, len(results[0].Violations)))
	assert.Equal(t, len(results[0].Vulnerabilities), vulnerabilities, fmt.Sprintf("Expected %d vulnerabilities in scan results, but got %d vulnerabilities.", vulnerabilities, len(results[0].Vulnerabilities)))
	assert.Equal(t, len(results[0].Licenses), licenses, fmt.Sprintf("Expected %d Licenses in scan results, but got %d Licenses.", licenses, len(results[0].Licenses)))
}
