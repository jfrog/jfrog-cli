package main

import (
	"errors"
	"fmt"
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
	"github.com/stretchr/testify/assert"
)

const (
	xrayDomain = "xray/"
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
	xrayDetails = &config.ServerDetails{XrayUrl: clientutils.AddTrailingSlashIfNeeded(*tests.JfrogUrl) + xrayDomain}
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
	binariesPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "binaries", "*")
	runXr(t, "scan", "--licenses", binariesPath)
}

func initXrayTest(t *testing.T, minVersion string) {
	if !*tests.TestXray {
		t.Skip("Skipping Xray test. To run Xray test add the '-test.xray=true' option.")
	}
	xrayVersion, err := getXrayVersion()
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !xrayVersion.AtLeast(minVersion) {
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
