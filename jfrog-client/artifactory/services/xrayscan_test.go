package services

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils/tests/xray"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	"strconv"
	"strings"
	"testing"
)

var testsXrayScanService *XrayScanService

func TestNewXrayScanService(t *testing.T) {
	xrayServerPort := xray.StartXrayMockServer()
	testsXrayScanService = NewXrayScanService(httpclient.NewDefaultHttpClient())
	testsXrayScanService.ArtDetails = getArtDetails()
	testsXrayScanService.ArtDetails.SetUrl("http://localhost:" + strconv.Itoa(xrayServerPort) + "/")

	// Run tests
	tests := []struct {
		name        string
		buildName   string
		buildNumber string
		expected    string
	}{
		{name: "scanCleanBuild", buildName: xray.CleanScanBuildName, buildNumber: "3", expected: xray.CleanXrayScanResponse},
		{name: "scanVulnerableBuild", buildName: xray.VulnerableBuildName, buildNumber: "3", expected: xray.VulnerableXrayScanResponse},
		{name: "scanFatalBuild", buildName: xray.FatalScanBuildName, buildNumber: "3", expected: xray.FatalErrorXrayScanResponse},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanBuild(t, test.buildName, test.buildNumber, test.expected)
		})
	}
}

func scanBuild(t *testing.T, buildName, buildNumber, expected string) {
	params := new(XrayScanParamsImpl)
	params.BuildName = buildName
	params.BuildNumber = buildNumber
	result, err := testsXrayScanService.ScanBuild(params)
	if err != nil {
		t.Error(err)
	}

	expected = strings.Replace(expected, "\n", "", -1)
	if string(result) != expected {
		t.Error("Expected:", string(result), "Got: ", expected)
	}
}
