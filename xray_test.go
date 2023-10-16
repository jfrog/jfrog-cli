package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/xray/scangraph"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jfrog/gofrog/version"
	coreContainer "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/container"
	artCmdUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	coreCmd "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	commontests "github.com/jfrog/jfrog-cli-core/v2/common/tests"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	coreCuration "github.com/jfrog/jfrog-cli-core/v2/xray/commands/curation"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/scan"
	"github.com/jfrog/jfrog-cli-core/v2/xray/formats"
	"github.com/jfrog/jfrog-cli-core/v2/xray/utils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/xray/services"
	xrayUtils "github.com/jfrog/jfrog-client-go/xray/services/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

var (
	xrayDetails *config.ServerDetails
	xrayAuth    auth.ServiceDetails
	// JFrog CLI for Xray commands
	xrayCli *tests.JfrogCli
)

func InitXrayTests() {
	initXrayCli()
}

func authenticateXray() string {
	*tests.JfrogUrl = clientUtils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
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
func TestXrayBinaryScanJson(t *testing.T) {
	output := testXrayBinaryScan(t, string(utils.Json))
	verifyJsonScanResults(t, output, 0, 1, 1)
}

func TestXrayBinaryScanSimpleJson(t *testing.T) {
	output := testXrayBinaryScan(t, string(utils.SimpleJson))
	verifySimpleJsonScanResults(t, output, 1, 1)
}

func TestXrayBinaryScanJsonWithProgress(t *testing.T) {
	callback := tests.MockProgressInitialization()
	defer callback()
	output := testXrayBinaryScan(t, string(utils.Json))
	verifyJsonScanResults(t, output, 0, 1, 1)
}

func TestXrayBinaryScanSimpleJsonWithProgress(t *testing.T) {
	callback := tests.MockProgressInitialization()
	defer callback()
	output := testXrayBinaryScan(t, string(utils.SimpleJson))
	verifySimpleJsonScanResults(t, output, 1, 1)
}

func testXrayBinaryScan(t *testing.T, format string) string {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	binariesPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "binaries", "*")
	return xrayCli.RunCliCmdWithOutput(t, "scan", binariesPath, "--licenses", "--format="+format)
}

func TestXrayBinaryScanWithBypassArchiveLimits(t *testing.T) {
	initXrayTest(t, scan.BypassArchiveLimitsMinXrayVersion)
	unsetEnv := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JF_INDEXER_COMPRESS_MAXENTITIES", "10")
	defer unsetEnv()
	binariesPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "binaries", "*")
	scanArgs := []string{"scan", binariesPath, "--format=json", "--licenses"}
	// Run without bypass flag and expect scan to fail
	err := xrayCli.Exec(scanArgs...)
	// Expect error
	assert.Error(t, err)

	// Run with bypass flag and expect it to find vulnerabilities
	scanArgs = append(scanArgs, "--bypass-archive-limits")
	output := xrayCli.RunCliCmdWithOutput(t, scanArgs...)
	verifyJsonScanResults(t, output, 0, 1, 1)
}

// Tests npm audit by providing simple npm project and asserts any error.
func TestXrayAuditNpmJson(t *testing.T) {
	output := testXrayAuditNpm(t, string(utils.Json))
	verifyJsonScanResults(t, output, 0, 1, 1)
}

func TestXrayAuditNpmSimpleJson(t *testing.T) {
	output := testXrayAuditNpm(t, string(utils.SimpleJson))
	verifySimpleJsonScanResults(t, output, 1, 1)
}

func testXrayAuditNpm(t *testing.T, format string) string {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	npmProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "npm")
	// Copy the npm project from the testdata to a temp dir
	assert.NoError(t, biutils.CopyDir(npmProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Run npm install before executing jfrog xr npm-audit
	assert.NoError(t, exec.Command("npm", "install").Run())
	// Add dummy descriptor file to check that we run only specific audit
	addDummyPackageDescriptor(t, true)
	return xrayCli.RunCliCmdWithOutput(t, "audit", "--npm", "--licenses", "--format="+format)
}

func TestXrayAuditYarnV2Json(t *testing.T) {
	testXrayAuditYarn(t, "yarn-v2", func() {
		output := runXrayAuditYarnWithOutput(t, string(utils.Json))
		verifyJsonScanResults(t, output, 0, 1, 1)
	})
}

func TestXrayAuditYarnV2SimpleJson(t *testing.T) {
	testXrayAuditYarn(t, "yarn-v2", func() {
		output := runXrayAuditYarnWithOutput(t, string(utils.SimpleJson))
		verifySimpleJsonScanResults(t, output, 1, 1)
	})
}

func TestXrayAuditYarnV1Json(t *testing.T) {
	testXrayAuditYarn(t, "yarn-v1", func() {
		output := runXrayAuditYarnWithOutput(t, string(utils.Json))
		verifyJsonScanResults(t, output, 0, 1, 1)
	})
}
func TestXrayAuditYarnV1SimpleJson(t *testing.T) {
	testXrayAuditYarn(t, "yarn-v1", func() {
		output := runXrayAuditYarnWithOutput(t, string(utils.SimpleJson))
		verifySimpleJsonScanResults(t, output, 1, 1)
	})
}

func testXrayAuditYarn(t *testing.T, projectDirName string, yarnCmd func()) {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	yarnProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", projectDirName)
	// Copy the Yarn project from the testdata to a temp directory
	assert.NoError(t, biutils.CopyDir(yarnProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Run yarn install before executing jf audit --yarn. Return error to assert according to test.
	assert.NoError(t, exec.Command("yarn").Run())
	// Add dummy descriptor file to check that we run only specific audit
	addDummyPackageDescriptor(t, true)
	yarnCmd()
}

func runXrayAuditYarnWithOutput(t *testing.T, format string) string {
	return xrayCli.RunCliCmdWithOutput(t, "audit", "--yarn", "--licenses", "--format="+format)
}

// Tests NuGet audit by providing simple NuGet project + multi-project NuGet project and asserts any error.
func TestXrayAuditNugetJson(t *testing.T) {
	var testdata = []struct {
		projectName        string
		format             string
		restoreTech        string
		minVulnerabilities int
		minLicences        int
	}{
		{
			projectName:        "single4.0",
			format:             string(utils.Json),
			restoreTech:        "nuget",
			minVulnerabilities: 2,
			minLicences:        0,
		},
		{
			projectName:        "single5.0",
			format:             string(utils.Json),
			restoreTech:        "dotnet",
			minVulnerabilities: 3,
			minLicences:        2,
		},
		{
			projectName:        "single5.0",
			format:             string(utils.Json),
			restoreTech:        "",
			minVulnerabilities: 3,
			minLicences:        2,
		},
		{
			projectName:        "multi",
			format:             string(utils.Json),
			restoreTech:        "dotnet",
			minVulnerabilities: 5,
			minLicences:        3,
		},
		{
			projectName:        "multi",
			format:             string(utils.Json),
			restoreTech:        "",
			minVulnerabilities: 5,
			minLicences:        3,
		},
	}
	for _, test := range testdata {
		runInstallCommand := test.restoreTech != ""
		t.Run(fmt.Sprintf("projectName:%s,runInstallCommand:%t", test.projectName, runInstallCommand),
			func(t *testing.T) {
				output := testXrayAuditNuget(t, test.projectName, test.format, test.restoreTech)
				verifyJsonScanResults(t, output, 0, test.minVulnerabilities, test.minLicences)
			})
	}
}

func TestXrayAuditNugetSimpleJson(t *testing.T) {
	var testdata = []struct {
		projectName        string
		format             string
		restoreTech        string
		minVulnerabilities int
		minLicences        int
	}{
		{
			projectName:        "single4.0",
			format:             string(utils.SimpleJson),
			restoreTech:        "nuget",
			minVulnerabilities: 2,
			minLicences:        0,
		},
		{
			projectName:        "single5.0",
			format:             string(utils.SimpleJson),
			restoreTech:        "dotnet",
			minVulnerabilities: 3,
			minLicences:        2,
		},
		{
			projectName:        "single5.0",
			format:             string(utils.SimpleJson),
			restoreTech:        "",
			minVulnerabilities: 3,
			minLicences:        2,
		},
	}
	for _, test := range testdata {
		runInstallCommand := test.restoreTech != ""
		t.Run(fmt.Sprintf("projectName:%s,runInstallCommand:%t", test.projectName, runInstallCommand),
			func(t *testing.T) {
				output := testXrayAuditNuget(t, test.projectName, test.format, test.restoreTech)
				verifySimpleJsonScanResults(t, output, test.minVulnerabilities, test.minLicences)
			})
	}
}

func testXrayAuditNuget(t *testing.T, projectName, format string, restoreTech string) string {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	projectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "nuget", projectName)

	assert.NoError(t, biutils.CopyDir(projectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Add dummy descriptor file to check that we run only specific audit
	addDummyPackageDescriptor(t, false)
	// Run NuGet/Dotnet restore before executing jfrog xr audit (NuGet)
	if restoreTech != "" {
		_, err := exec.Command(restoreTech, "restore").CombinedOutput()
		assert.NoError(t, err)
	}
	return xrayCli.RunCliCmdWithOutput(t, "audit", "--nuget", "--format="+format, "--licenses")
}

func TestXrayAuditGradleJson(t *testing.T) {
	output := testXrayAuditGradle(t, string(utils.Json))
	verifyJsonScanResults(t, output, 0, 3, 3)
}

func TestXrayAuditGradleSimpleJson(t *testing.T) {
	output := testXrayAuditGradle(t, string(utils.SimpleJson))
	verifySimpleJsonScanResults(t, output, 3, 3)
}

func testXrayAuditGradle(t *testing.T, format string) string {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	gradleProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "gradle")
	// Copy the gradle project from the testdata to a temp dir
	assert.NoError(t, biutils.CopyDir(gradleProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Add dummy descriptor file to check that we run only specific audit
	addDummyPackageDescriptor(t, false)
	return xrayCli.RunCliCmdWithOutput(t, "audit", "--gradle", "--licenses", "--format="+format)
}

func TestXrayAuditMavenJson(t *testing.T) {
	output := testXrayAuditMaven(t, string(utils.Json))
	verifyJsonScanResults(t, output, 0, 1, 1)
}

func TestXrayAuditMavenSimpleJson(t *testing.T) {
	output := testXrayAuditMaven(t, string(utils.SimpleJson))
	verifySimpleJsonScanResults(t, output, 1, 1)
}

func testXrayAuditMaven(t *testing.T, format string) string {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	mvnProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "maven")
	// Copy the maven project from the testdata to a temp dir
	assert.NoError(t, biutils.CopyDir(mvnProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Add dummy descriptor file to check that we run only specific audit
	addDummyPackageDescriptor(t, false)
	return xrayCli.RunCliCmdWithOutput(t, "audit", "--mvn", "--licenses", "--format="+format)
}

func TestXrayAuditNoTech(t *testing.T) {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Run audit on empty folder, expect an error
	err := xrayCli.Exec("audit")
	assert.NoError(t, err)
}

func TestXrayAuditDetectTech(t *testing.T) {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	mvnProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "maven")
	// Copy the maven project from the testdata to a temp dir
	assert.NoError(t, biutils.CopyDir(mvnProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Run generic audit on mvn project with a vulnerable dependency
	output := xrayCli.RunCliCmdWithOutput(t, "audit", "--licenses", "--format="+string(utils.SimpleJson))
	var results formats.SimpleJsonResults
	err := json.Unmarshal([]byte(output), &results)
	assert.NoError(t, err)
	// Expects the ImpactedPackageType of the known vulnerability to be maven
	assert.Equal(t, strings.ToLower(results.Vulnerabilities[0].ImpactedDependencyType), "maven")
}

func TestXrayAuditMultiProjects(t *testing.T) {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	multiProject := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray")
	// Copy the multi project from the testdata to a temp dir
	assert.NoError(t, biutils.CopyDir(multiProject, tempDirPath, true, nil))
	workingDirsFlag := fmt.Sprintf("--working-dirs=%s, %s ,%s, %s",
		filepath.Join(tempDirPath, "maven"), filepath.Join(tempDirPath, "nuget", "single4.0"),
		filepath.Join(tempDirPath, "python", "pip"), filepath.Join(tempDirPath, "jas-test"))
	// Configure a new server named "default"
	createJfrogHomeConfig(t, true)
	defer cleanTestsHomeEnv()
	output := xrayCli.WithoutCredentials().RunCliCmdWithOutput(t, "audit", "--format="+string(utils.SimpleJson), workingDirsFlag)
	verifySimpleJsonScanResults(t, output, 35, 0)
	verifySimpleJsonJasResults(t, output, 3, 9, 7, 3)
}

func TestXrayAuditPipJson(t *testing.T) {
	output := testXrayAuditPip(t, string(utils.Json), "")
	verifyJsonScanResults(t, output, 0, 3, 1)
}

func TestXrayAuditPipSimpleJson(t *testing.T) {
	output := testXrayAuditPip(t, string(utils.SimpleJson), "")
	verifySimpleJsonScanResults(t, output, 3, 1)
}

func TestXrayAuditPipJsonWithRequirementsFile(t *testing.T) {
	output := testXrayAuditPip(t, string(utils.Json), "requirements.txt")
	verifyJsonScanResults(t, output, 0, 2, 0)
}

func TestXrayAuditPipSimpleJsonWithRequirementsFile(t *testing.T) {
	output := testXrayAuditPip(t, string(utils.SimpleJson), "requirements.txt")
	verifySimpleJsonScanResults(t, output, 2, 0)
}

func testXrayAuditPip(t *testing.T, format, requirementsFile string) string {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	pipProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "python", "pip")
	// Copy the pip project from the testdata to a temp dir
	assert.NoError(t, biutils.CopyDir(pipProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Add dummy descriptor file to check that we run only specific audit
	addDummyPackageDescriptor(t, false)
	args := []string{"audit", "--pip", "--licenses", "--format=" + format}
	if requirementsFile != "" {
		args = append(args, "--requirements-file="+requirementsFile)

	}
	return xrayCli.RunCliCmdWithOutput(t, args...)
}

func TestXrayAuditPipenvJson(t *testing.T) {
	output := testXrayAuditPipenv(t, string(utils.Json))
	verifyJsonScanResults(t, output, 0, 3, 1)
}

func TestXrayAuditPipenvSimpleJson(t *testing.T) {
	output := testXrayAuditPipenv(t, string(utils.SimpleJson))
	verifySimpleJsonScanResults(t, output, 3, 1)
}

func testXrayAuditPipenv(t *testing.T, format string) string {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	pipenvProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "python", "pipenv")
	// Copy the pipenv project from the testdata to a temp dir
	assert.NoError(t, biutils.CopyDir(pipenvProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Add dummy descriptor file to check that we run only specific audit
	addDummyPackageDescriptor(t, false)
	return xrayCli.RunCliCmdWithOutput(t, "audit", "--pipenv", "--licenses", "--format="+format)
}

func TestDownloadAnalyzerManagerIfNeeded(t *testing.T) {
	initXrayTest(t, "")

	// Configure a new JFrog CLI home dir.
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.HomeDir, tempDirPath)
	defer setEnvCallBack()

	// Download
	err := artUtils.DownloadAnalyzerManagerIfNeeded()
	assert.NoError(t, err)

	// Validate Analyzer manager app & checksum.sh2 file exist
	path, err := utils.GetAnalyzerManagerDirAbsolutePath()
	assert.NoError(t, err)
	amPath := filepath.Join(path, utils.GetAnalyzerManagerExecutableName())
	exists, err := fileutils.IsFileExists(amPath, false)
	assert.NoError(t, err)
	assert.True(t, exists)
	checksumPath := filepath.Join(path, artUtils.ChecksumFileName)
	exists, err = fileutils.IsFileExists(checksumPath, false)
	assert.NoError(t, err)
	assert.True(t, exists)
	checksumFileStat, err := os.Stat(checksumPath)
	assert.NoError(t, err)
	assert.True(t, checksumFileStat.Size() > 0)

	// Validate no second download occurred
	firstFileStat, err := os.Stat(amPath)
	assert.NoError(t, err)
	err = artUtils.DownloadAnalyzerManagerIfNeeded()
	assert.NoError(t, err)
	secondFileStat, err := os.Stat(amPath)
	assert.NoError(t, err)
	assert.Equal(t, firstFileStat.ModTime(), secondFileStat.ModTime())
}

func TestXrayAuditPoetryJson(t *testing.T) {
	output := testXrayAuditPoetry(t, string(utils.Json))
	verifyJsonScanResults(t, output, 0, 3, 1)
}

func TestXrayAuditPoetrySimpleJson(t *testing.T) {
	output := testXrayAuditPoetry(t, string(utils.SimpleJson))
	verifySimpleJsonScanResults(t, output, 3, 1)
}

func testXrayAuditPoetry(t *testing.T, format string) string {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	poetryProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray", "python", "poetry")
	// Copy the poetry project from the testdata to a temp dir
	assert.NoError(t, biutils.CopyDir(poetryProjectPath, tempDirPath, true, nil))
	prevWd := changeWD(t, tempDirPath)
	defer clientTestUtils.ChangeDirAndAssert(t, prevWd)
	// Add dummy descriptor file to check that we run only specific audit
	addDummyPackageDescriptor(t, false)
	return xrayCli.RunCliCmdWithOutput(t, "audit", "--poetry", "--licenses", "--format="+format)
}

func addDummyPackageDescriptor(t *testing.T, hasPackageJson bool) {
	descriptor := "package.json"
	if hasPackageJson {
		descriptor = "pom.xml"
	}
	dummyFile, err := os.Create(descriptor)
	assert.NoError(t, err)
	assert.NoError(t, dummyFile.Close())
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
	err = clientUtils.ValidateMinimumVersion(clientUtils.Xray, xrayVersion.GetVersion(), minVersion)
	if err != nil {
		t.Skip(err)
	}
}

func getXrayVersion() (version.Version, error) {
	xrayVersion, err := xrayAuth.GetVersion()
	return *version.NewVersion(xrayVersion), err
}

func verifyJsonScanResults(t *testing.T, content string, minViolations, minVulnerabilities, minLicenses int) {
	var results []services.ScanResponse
	err := json.Unmarshal([]byte(content), &results)
	if assert.NoError(t, err) {
		var violations []services.Violation
		var vulnerabilities []services.Vulnerability
		var licenses []services.License
		for _, result := range results {
			violations = append(violations, result.Violations...)
			vulnerabilities = append(vulnerabilities, result.Vulnerabilities...)
			licenses = append(licenses, result.Licenses...)
		}
		assert.True(t, len(violations) >= minViolations, fmt.Sprintf("Expected at least %d violations in scan results, but got %d violations.", minViolations, len(violations)))
		assert.True(t, len(vulnerabilities) >= minVulnerabilities, fmt.Sprintf("Expected at least %d vulnerabilities in scan results, but got %d vulnerabilities.", minVulnerabilities, len(vulnerabilities)))
		assert.True(t, len(licenses) >= minLicenses, fmt.Sprintf("Expected at least %d Licenses in scan results, but got %d Licenses.", minLicenses, len(licenses)))
	}
}

func verifySimpleJsonScanResults(t *testing.T, content string, minVulnerabilities, minLicenses int) {
	var results formats.SimpleJsonResults
	err := json.Unmarshal([]byte(content), &results)
	if assert.NoError(t, err) {
		assert.GreaterOrEqual(t, len(results.Vulnerabilities), minVulnerabilities)
		assert.GreaterOrEqual(t, len(results.Licenses), minLicenses)
	}
}

func TestXrayCurl(t *testing.T) {
	initXrayTest(t, "")
	// Configure a new server named "default".
	createJfrogHomeConfig(t, true)
	defer cleanTestsHomeEnv()
	// Check curl command with the default configured server.
	err := xrayCli.WithoutCredentials().Exec("xr", "curl", "-XGET", "/api/v1/system/version")
	assert.NoError(t, err)
	// Check curl command with '--server-id' flag
	err = xrayCli.WithoutCredentials().Exec("xr", "curl", "-XGET", "/api/system/version", "--server-id=default")
	assert.NoError(t, err)
	// Check curl command with invalid server id - should get an error.
	err = xrayCli.WithoutCredentials().Exec("xr", "curl", "-XGET", "/api/system/version", "--server-id=not_configured_name")
	assert.EqualError(t, err, "Server ID 'not_configured_name' does not exist.")
}

func initNativeDockerWithXrayTest(t *testing.T) func() {
	if !*tests.TestDockerScan || !*tests.TestXray {
		t.Skip("Skipping Docker scan test. To run Xray Docker test add the '-test.xrayScan=true' and '-test.xray=true' options.")
	}
	oldHomeDir := os.Getenv(coreutils.HomeDir)
	initXrayCli()
	validateXrayVersion(t, scan.DockerScanMinXrayVersion)
	// Create server config to use with the command.
	createJfrogHomeConfig(t, true)
	return func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
	}
}

func TestDockerScan(t *testing.T) {
	cleanup := initNativeDockerWithXrayTest(t)
	defer cleanup()

	watchName, deleteWatch := createTestWatch(t)
	defer deleteWatch()

	imagesToScan := []string{
		// Simple image with vulnerabilities
		"bitnami/minio:2022",

		// Image with RPM with vulnerabilities
		"redhat/ubi8-micro:8.4",
	}
	for _, imageName := range imagesToScan {
		runDockerScan(t, imageName, watchName, 3, 3, 3)
	}

	// On Xray 3.40.3 there is a bug whereby xray fails to scan docker image with 0 vulnerabilities,
	// So we skip it for now till the next version will be released
	validateXrayVersion(t, "3.41.0")

	// Image with 0 vulnerabilities
	runDockerScan(t, "busybox:1.35", "", 0, 0, 0)
}

func TestDockerScanWithProgressBar(t *testing.T) {
	callback := tests.MockProgressInitialization()
	defer callback()
	TestDockerScan(t)
}

func runDockerScan(t *testing.T, imageName, watchName string, minViolations, minVulnerabilities, minLicenses int) {
	// Pull image from docker repo
	imageTag := path.Join(*tests.ContainerRegistry, tests.DockerVirtualRepo, imageName)
	dockerPullCommand := coreContainer.NewPullCommand(container.DockerClient)
	dockerPullCommand.SetCmdParams([]string{"pull", imageTag}).SetImageTag(imageTag).SetRepo(tests.DockerVirtualRepo).SetServerDetails(serverDetails).SetBuildConfiguration(new(artUtils.BuildConfiguration))
	if assert.NoError(t, dockerPullCommand.Run()) {
		defer inttestutils.DeleteTestImage(t, imageTag, container.DockerClient)

		args := []string{"docker", "scan", imageTag, "--server-id=default", "--licenses", "--format=json", "--fail=false", "--min-severity=low", "--fixable-only"}

		// Run docker scan on image
		output := xrayCli.WithoutCredentials().RunCliCmdWithOutput(t, args...)
		if assert.NotEmpty(t, output) {
			verifyJsonScanResults(t, output, 0, minVulnerabilities, minLicenses)
		}

		// Run docker scan on image with watch
		if watchName != "" {
			args = append(args, "--watches="+watchName)
			output = xrayCli.WithoutCredentials().RunCliCmdWithOutput(t, args...)
			if assert.NotEmpty(t, output) {
				verifyJsonScanResults(t, output, minViolations, 0, 0)
			}
		}
	}
}

func createTestWatch(t *testing.T) (string, func()) {
	xrayManager, err := utils.CreateXrayServiceManager(xrayDetails)
	assert.NoError(t, err)
	// Create new default policy.
	policyParams := xrayUtils.PolicyParams{
		Name: fmt.Sprintf("%s-%s", "docker-policy", strconv.FormatInt(time.Now().Unix(), 10)),
		Type: xrayUtils.Security,
		Rules: []xrayUtils.PolicyRule{{
			Name:     "sec_rule",
			Criteria: *xrayUtils.CreateSeverityPolicyCriteria(xrayUtils.Low),
			Priority: 1,
			Actions: &xrayUtils.PolicyAction{
				FailBuild: clientUtils.Pointer(true),
			},
		}},
	}
	if !assert.NoError(t, xrayManager.CreatePolicy(policyParams)) {
		return "", func() {}
	}
	// Create new default watch.
	watchParams := xrayUtils.NewWatchParams()
	watchParams.Name = fmt.Sprintf("%s-%s", "docker-watch", strconv.FormatInt(time.Now().Unix(), 10))
	watchParams.Active = true
	watchParams.Builds.Type = xrayUtils.WatchBuildAll
	watchParams.Policies = []xrayUtils.AssignedPolicy{
		{
			Name: policyParams.Name,
			Type: "security",
		},
	}
	assert.NoError(t, xrayManager.CreateWatch(watchParams))
	return watchParams.Name, func() {
		assert.NoError(t, xrayManager.DeleteWatch(watchParams.Name))
		assert.NoError(t, xrayManager.DeletePolicy(policyParams.Name))
	}
}

func TestXrayOfflineDBSyncV3(t *testing.T) {
	initXrayTest(t, "")

	// Validate license-id
	err := xrayCli.WithoutCredentials().Exec("xr", "ou")
	assert.EqualError(t, err, "the --license-id option is mandatory")
	// Periodic valid only with stream
	err = xrayCli.WithoutCredentials().Exec("xr", "ou", "--license-id=123", "--periodic")
	assert.EqualError(t, err, fmt.Sprintf("the %s option is only valid with %s", cliutils.Periodic, cliutils.Stream))
	err = xrayCli.WithoutCredentials().Exec("xr", "ou", "--license-id=123", "--stream=", "--periodic")
	assert.EqualError(t, err, fmt.Sprintf("the %s option is only valid with %s", cliutils.Periodic, cliutils.Stream))
	// Invalid stream
	err = xrayCli.WithoutCredentials().Exec("xr", "ou", "--license-id=123", "--stream=bad_name")
	assert.ErrorContains(t, err, "Invalid stream type")
}

func TestXrayAuditJasSimpleJson(t *testing.T) {
	output := testXrayAuditJas(t, string(utils.SimpleJson), "jas-test")
	verifySimpleJsonJasResults(t, output, 3, 9, 7, 2)
}

func TestXrayAuditJasSimpleJsonWithConfig(t *testing.T) {
	output := testXrayAuditJas(t, string(utils.SimpleJson), "jas-config")
	verifySimpleJsonJasResults(t, output, 0, 0, 1, 2)
}

func TestXrayAuditJasNoViolationsSimpleJson(t *testing.T) {
	output := testXrayAuditJas(t, string(utils.SimpleJson), "npm")
	verifySimpleJsonScanResults(t, output, 2, 0)
	verifySimpleJsonJasResults(t, output, 0, 0, 0, 0)
}

func testXrayAuditJas(t *testing.T, format string, project string) string {
	initXrayTest(t, scangraph.GraphScanMinXrayVersion)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	projectDir := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), filepath.Join("xray", project))
	// Copy the multi project from the testdata to a temp dir
	assert.NoError(t, biutils.CopyDir(projectDir, tempDirPath, true, nil))
	// Configure a new server named "default"
	createJfrogHomeConfig(t, true)
	defer cleanTestsHomeEnv()
	baseWd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, baseWd, tempDirPath)
	defer chdirCallback()
	return xrayCli.WithoutCredentials().RunCliCmdWithOutput(t, "audit", "--format="+format)
}

func verifySimpleJsonJasResults(t *testing.T, content string, minSastViolations, minIacViolations, minSecrets, minApplicable int) {
	var results formats.SimpleJsonResults
	err := json.Unmarshal([]byte(content), &results)
	if assert.NoError(t, err) {
		assert.GreaterOrEqual(t, len(results.Sast), minSastViolations, "Found less sast then expected")
		assert.GreaterOrEqual(t, len(results.Secrets), minSecrets, "Found less secrets then expected")
		assert.GreaterOrEqual(t, len(results.Iacs), minIacViolations, "Found less IaC then expected")
		var applicableResults, notApplicableResults int
		for _, vuln := range results.Vulnerabilities {
			if vuln.Applicable == string(utils.NotApplicable) {
				notApplicableResults++
			} else if vuln.Applicable == string(utils.Applicable) {
				applicableResults++
			}
		}
		assert.GreaterOrEqual(t, applicableResults, minApplicable, "Found less applicableResults then expected")
		assert.GreaterOrEqual(t, notApplicableResults, 1, "Found less notApplicableResults then expected")
	}
}

func TestCurationAudit(t *testing.T) {
	initXrayTest(t, "")
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	multiProject := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "xray")
	assert.NoError(t, biutils.CopyDir(multiProject, tempDirPath, true, nil))
	rootDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Chdir(rootDir))
	}()
	require.NoError(t, os.Chdir(filepath.Join(tempDirPath, "npm")))
	expectedRequest := map[string]bool{
		"/api/npm/npms/json/-/json-9.0.6.tgz": false,
		"/api/npm/npms/xml/-/xml-1.0.1.tgz":   false,
	}
	requestToFail := map[string]bool{
		"/api/npm/npms/xml/-/xml-1.0.1.tgz": false,
	}
	serverMock, config := curationServer(t, expectedRequest, requestToFail)

	cleanUpJfrogHome, err := coretests.SetJfrogHome()
	assert.NoError(t, err)
	defer cleanUpJfrogHome()

	config.User = "admin"
	config.Password = "password"
	config.ServerId = "test"
	configCmd := coreCmd.NewConfigCommand(coreCmd.AddOrEdit, "test").SetDetails(config).SetUseBasicAuthOnly(true).SetInteractive(false)
	assert.NoError(t, configCmd.Run())

	defer serverMock.Close()
	// Create build config
	resolutionServerId := "server-id-resolve"
	deploymentServerId := "server-id-deploy"
	resolutionRepo := "repo-resolve"
	deploymentRepo := "repo-deploy"
	context := createContext(t, resolutionServerId+"="+config.ServerId, resolutionRepo+"=npms", deploymentServerId+"="+config.ServerId, deploymentRepo+"=npm-local", "global=false")
	err = artCmdUtils.CreateBuildConfig(context, artUtils.Npm)
	assert.NoError(t, err)

	localXrayCli := xrayCli.WithoutCredentials()
	workingDirsFlag := fmt.Sprintf("--working-dirs=%s", filepath.Join(tempDirPath, "npm"))
	output := localXrayCli.RunCliCmdWithOutput(t, "curation-audit", "--format="+string(utils.Json), workingDirsFlag)
	expectedResp := getCurationExpectedResponse(config)
	var got []coreCuration.PackageStatus
	bracketIndex := strings.Index(output, "[")
	require.Less(t, 0, bracketIndex, "Unexpected Curation output with missing '['")
	err = json.Unmarshal([]byte(output[bracketIndex:]), &got)
	assert.NoError(t, err)
	assert.Equal(t, expectedResp, got)
	for k, v := range expectedRequest {
		assert.Truef(t, v, "didn't receive expected GET request for package url %s", k)
	}
}

func getCurationExpectedResponse(config *config.ServerDetails) []coreCuration.PackageStatus {
	expectedResp := []coreCuration.PackageStatus{
		{
			Action:            "blocked",
			PackageName:       "xml",
			PackageVersion:    "1.0.1",
			BlockedPackageUrl: config.ArtifactoryUrl + "api/npm/npms/xml/-/xml-1.0.1.tgz",
			BlockingReason:    coreCuration.BlockingReasonPolicy,
			ParentName:        "xml",
			ParentVersion:     "1.0.1",
			DepRelation:       "direct",
			PkgType:           "npm",
			Policy: []coreCuration.Policy{
				{
					Policy:         "pol1",
					Condition:      "cond1",
					Explanation:    "explanation",
					Recommendation: "recommendation",
				},
				{
					Policy:         "pol2",
					Condition:      "cond2",
					Explanation:    "explanation2",
					Recommendation: "recommendation2",
				},
			},
		},
	}
	return expectedResp
}

func curationServer(t *testing.T, expectedRequest map[string]bool, requestToFail map[string]bool) (*httptest.Server, *config.ServerDetails) {
	mapLockReadWrite := sync.Mutex{}
	serverMock, config, _ := commontests.CreateRtRestsMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			mapLockReadWrite.Lock()
			if _, exist := expectedRequest[r.RequestURI]; exist {
				expectedRequest[r.RequestURI] = true
			}
			mapLockReadWrite.Unlock()
			if _, exist := requestToFail[r.RequestURI]; exist {
				w.WriteHeader(http.StatusForbidden)
			}
		}
		if r.Method == http.MethodGet {
			if _, exist := requestToFail[r.RequestURI]; exist {
				w.WriteHeader(http.StatusForbidden)
				_, err := w.Write([]byte("{\n    \"errors\": [\n        {\n            \"status\": 403,\n            " +
					"\"message\": \"Package download was blocked by JFrog Packages " +
					"Curation service due to the following policies violated {pol1, cond1, explanation, recommendation}, {pol2, cond2, explanation2, recommendation2}\"\n        }\n    ]\n}"))
				require.NoError(t, err)
			}
		}
	})
	return serverMock, config
}

func createContext(t *testing.T, stringFlags ...string) *cli.Context {
	flagSet := flag.NewFlagSet("TestFlagSet", flag.ContinueOnError)
	flags := setStringFlags(flagSet, stringFlags...)
	assert.NoError(t, flagSet.Parse(flags))
	return cli.NewContext(nil, flagSet, nil)
}

func setStringFlags(flagSet *flag.FlagSet, flags ...string) []string {
	cmdFlags := []string{}
	for _, stringFlag := range flags {
		flagSet.String(strings.Split(stringFlag, "=")[0], "", "")
		cmdFlags = append(cmdFlags, "--"+stringFlag)
	}
	return cmdFlags
}
