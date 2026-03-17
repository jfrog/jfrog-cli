package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	buildutils "github.com/jfrog/build-info-go/build/utils"
	buildinfo "github.com/jfrog/build-info-go/entities"
	biutils "github.com/jfrog/build-info-go/utils"
	npmCmdUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/ioutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	artServices "github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/lifecycle/services"
	"github.com/jfrog/jfrog-client-go/utils/log"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
)

type pnpmTestParams struct {
	testName       string
	command        string
	repo           string
	pnpmArgs       string
	wd             string
	buildNumber    string
	moduleName     string
	validationFunc func(*testing.T, pnpmTestParams)
}

func cleanPnpmTest(t *testing.T) {
	clientTestUtils.UnSetEnvAndAssert(t, coreutils.HomeDir)
	deleteSpec := spec.NewBuilder().Pattern(tests.NpmRepo).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
	assert.NoError(t, err)
	tests.CleanFileSystem()
}

func TestPnpmInstall(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	pnpmProjectPath, pnpmScopedProjectPath := initPnpmFilesTest(t)
	var pnpmTests = []pnpmTestParams{
		{testName: "pnpm i", command: "install", repo: tests.NpmRemoteRepo, wd: pnpmProjectPath, validationFunc: validatePnpmInstall},
		{testName: "pnpm i with module", command: "install", repo: tests.NpmRemoteRepo, wd: pnpmProjectPath, moduleName: ModuleNameJFrogTest, validationFunc: validatePnpmInstall},
		{testName: "pnpm i with scoped project", command: "install", repo: tests.NpmRemoteRepo, wd: pnpmScopedProjectPath, validationFunc: validatePnpmInstall},
	}

	for i, pt := range pnpmTests {
		t.Run(pt.testName, func(t *testing.T) {
			buildNumber := strconv.Itoa(i + 200)
			clientTestUtils.ChangeDirAndAssert(t, filepath.Dir(pt.wd))

			args := []string{"pnpm", pt.command, "--store-dir=" + tempCacheDirPath,
				"--build-name=" + tests.PnpmBuildName, "--build-number=" + buildNumber}
			if pt.pnpmArgs != "" {
				args = append(args, strings.Split(pt.pnpmArgs, " ")...)
			}
			if pt.moduleName != "" {
				args = append(args, "--module="+pt.moduleName)
			} else {
				pt.moduleName = readPnpmModuleId(t, pt.wd)
			}

			runJfrogCli(t, args...)
			validatePnpmLocalBuildInfo(t, tests.PnpmBuildName, buildNumber, pt.moduleName)
			assert.NoError(t, artifactoryCli.Exec("bp", tests.PnpmBuildName, buildNumber))
			pt.buildNumber = buildNumber
			pt.validationFunc(t, pt)
		})
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

func TestPnpmPublish(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	pnpmProjectPath, pnpmScopedProjectPath := initPnpmFilesTest(t)
	var pnpmTests = []pnpmTestParams{
		{testName: "pnpm publish", command: "publish", repo: tests.NpmRepo, wd: pnpmProjectPath, validationFunc: validatePnpmPublish},
		{testName: "pnpm p with module", command: "p", repo: tests.NpmScopedRepo, wd: pnpmScopedProjectPath, moduleName: ModuleNameJFrogTest, validationFunc: validatePnpmScopedPublish},
	}

	for i, pt := range pnpmTests {
		t.Run(pt.testName, func(t *testing.T) {
			buildNumber := strconv.Itoa(i + 300)
			projectDir := filepath.Dir(pt.wd)
			clientTestUtils.ChangeDirAndAssert(t, projectDir)

			cleanupAuth := setupPnpmPublishAuth(t, pt.repo)
			defer cleanupAuth()

			args := []string{"pnpm", pt.command, "--no-git-checks",
				"--build-name=" + tests.PnpmBuildName, "--build-number=" + buildNumber}
			if pt.moduleName != "" {
				args = append(args, "--module="+pt.moduleName)
			} else {
				pt.moduleName = readPnpmModuleId(t, pt.wd)
			}

			runJfrogCli(t, args...)
			validatePnpmLocalBuildInfo(t, tests.PnpmBuildName, buildNumber, pt.moduleName)
			assert.NoError(t, artifactoryCli.Exec("bp", tests.PnpmBuildName, buildNumber))
			pt.buildNumber = buildNumber
			pt.validationFunc(t, pt)
		})
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

func TestPnpmPublishWorkspace(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	pnpmWorkspacePath := initPnpmWorkspaceTest(t)
	buildNumber := "400"
	clientTestUtils.ChangeDirAndAssert(t, filepath.Dir(pnpmWorkspacePath))

	cleanupAuth := setupPnpmPublishAuth(t, tests.NpmRepo)
	defer cleanupAuth()

	runJfrogCli(t, "pnpm", "publish", "-r", "--no-git-checks",
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)

	buildInfoService := build.CreateBuildInfoService()
	pnpmBuild, err := buildInfoService.GetOrCreateBuildWithProject(tests.PnpmBuildName, buildNumber, "")
	assert.NoError(t, err)
	bi, err := pnpmBuild.ToBuildInfo()
	assert.NoError(t, err)
	assert.NotEmpty(t, bi.Started)
	expectedWorkspaceCount := 2 // nested1 and nested2 (root is private)
	assert.Len(t, bi.Modules, expectedWorkspaceCount,
		"module count should equal number of published workspaces (nested1, nested2), got %d", len(bi.Modules))

	assert.NoError(t, artifactoryCli.Exec("bp", tests.PnpmBuildName, buildNumber))

	workspaceArtifacts := []string{
		tests.NpmRepo + "/nested1/-/nested1-1.0.0.tgz",
		tests.NpmRepo + "/nested2/-/nested2-1.0.0.tgz",
	}
	verifyExistInArtifactoryByProps(workspaceArtifacts,
		tests.NpmRepo+"/*",
		fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", tests.PnpmBuildName, buildNumber), t)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

func TestPnpmInstallAndPublishNormalProject(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	buildNumber := "500"
	pnpmProjectPath := createPnpmProject(t, "pnpmproject")
	projectDir := filepath.Dir(pnpmProjectPath)
	prepareArtifactoryForPnpmBuild(t, projectDir)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)
	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)

	cleanupAuth := setupPnpmPublishAuth(t, tests.NpmRepo)
	defer cleanupAuth()
	runJfrogCli(t, "pnpm", "publish", "--no-git-checks",
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.PnpmBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PnpmBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	bi := publishedBuildInfo.BuildInfo
	assert.Len(t, bi.Modules, 1)
	assert.NotEmpty(t, bi.Modules[0].Dependencies, "module should have dependencies")
	assert.NotEmpty(t, bi.Modules[0].Artifacts, "module should have artifacts")

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

func TestPnpmInstallAndPublishWorkspace(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	buildNumber := "501"
	pnpmWorkspacePath := initPnpmWorkspaceTest(t)
	clientTestUtils.ChangeDirAndAssert(t, pnpmWorkspacePath)

	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)

	cleanupAuth := setupPnpmPublishAuth(t, tests.NpmRepo)
	defer cleanupAuth()
	runJfrogCli(t, "pnpm", "publish", "-r", "--no-git-checks",
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.PnpmBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PnpmBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	bi := publishedBuildInfo.BuildInfo
	packagesPublished := 2 // nested1 and nested2
	modulesWithDepsAndArtifacts := 0
	for _, mod := range bi.Modules {
		hasDeps := len(mod.Dependencies) > 0
		hasArtifacts := len(mod.Artifacts) > 0
		assert.True(t, hasDeps, "module %s should have dependencies", mod.Id)
		if hasArtifacts {
			modulesWithDepsAndArtifacts++
			assert.True(t, hasDeps, "module %s has artifacts so must have dependencies", mod.Id)
		}
	}
	assert.Equal(t, packagesPublished, modulesWithDepsAndArtifacts,
		"number of modules with both dependencies and artifacts should equal packages published (nested1, nested2), got %d", modulesWithDepsAndArtifacts)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

// TestPnpmInstallWithPreviousBuildCache verifies that a second install with the same build name
// can use the previous published build's dependencies for checksum cache (GetDependenciesFromLatestBuild).
func TestPnpmInstallWithPreviousBuildCache(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	buildName := tests.PnpmBuildName
	buildNum1, buildNum2 := "602", "603"
	pnpmProjectPath := createPnpmProject(t, "pnpmproject")
	projectDir := filepath.Dir(pnpmProjectPath)

	// Write .npmrc so pnpm resolves through Artifactory (required for AQL checksum resolution).
	registry := npmCmdUtils.GetNpmRepositoryUrl(tests.NpmRemoteRepo, serverDetails.GetArtifactoryUrl())
	registryWithSlash := strings.TrimSuffix(registry, "/") + "/"
	authKey, authValue := npmCmdUtils.GetNpmAuthKeyValue(serverDetails, registryWithSlash)
	npmrcContent := fmt.Sprintf("registry=%s\n%s=%s\n", registryWithSlash, authKey, authValue)
	err = os.WriteFile(filepath.Join(projectDir, ".npmrc"), []byte(npmrcContent), 0644)
	assert.NoError(t, err)

	// Clear pnpm metadata cache for the Artifactory host to avoid stale tarball URLs
	// from repos created by previous test runs (repo names include a unique timestamp suffix).
	artHost := strings.TrimPrefix(strings.TrimPrefix(serverDetails.GetArtifactoryUrl(), "https://"), "http://")
	artHost = strings.SplitN(artHost, "/", 2)[0]
	if homeDir, hErr := os.UserHomeDir(); hErr == nil {
		_ = os.RemoveAll(filepath.Join(homeDir, "Library", "Caches", "pnpm", "metadata-v1.3", artHost))
	}

	prepareArtifactoryForPnpmBuild(t, projectDir)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)
	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--build-name="+buildName, "--build-number="+buildNum1)
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNum1))
	time.Sleep(3 * time.Second)

	// Second install: redirect log to capture "Checksum resolution: N cached, ..." from fetchChecksums
	_, logBuffer, previousLog := coretests.RedirectLogOutputToBuffer()
	defer log.SetLogger(previousLog)
	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--build-name="+buildName, "--build-number="+buildNum2)
	logOut := logBuffer.String()
	assert.Regexp(t, regexp.MustCompile(`Checksum resolution: ([1-9]\d*) cached`), logOut,
		"second install must use previous build cache for at least one dependency; log output: %s", logOut)
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNum2))

	for _, buildNumber := range []string{buildNum1, buildNum2} {
		publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
		assert.NoError(t, err)
		assert.True(t, found, "build %s/%s should be found", buildName, buildNumber)
		assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules, "build should have modules")
		assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies, "module should have dependencies")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// setupPnpmPublishAuth writes Artifactory registry and auth to ~/.npmrc
// so that pnpm publish (which delegates to npm from a temp dir) can authenticate.
// The registry URL must end with "/" for npm's nerfDart URL matching to work.
// Returns a cleanup function that restores the original ~/.npmrc.
func setupPnpmPublishAuth(t *testing.T, repo string) func() {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)
	npmrcPath := filepath.Join(homeDir, ".npmrc")

	origContent, origErr := os.ReadFile(npmrcPath)

	registry := npmCmdUtils.GetNpmRepositoryUrl(repo, serverDetails.GetArtifactoryUrl())
	registryWithSlash := strings.TrimSuffix(registry, "/") + "/"
	authKey, authValue := npmCmdUtils.GetNpmAuthKeyValue(serverDetails, registryWithSlash)
	assert.NotEmpty(t, authKey, "npm auth key must not be empty")

	npmrcContent := fmt.Sprintf("registry=%s\n%s=%s\n", registryWithSlash, authKey, authValue)
	err = os.WriteFile(npmrcPath, []byte(npmrcContent), 0644)
	assert.NoError(t, err)

	return func() {
		if origErr == nil {
			_ = os.WriteFile(npmrcPath, origContent, 0644)
		} else {
			_ = os.Remove(npmrcPath)
		}
	}
}

func readPnpmModuleId(t *testing.T, wd string) string {
	packageInfo, err := buildutils.ReadPackageInfoFromPackageJsonIfExists(filepath.Dir(wd), nil)
	assert.NoError(t, err)
	packageInfo.Version = strings.TrimPrefix(packageInfo.Version, "v")
	packageInfo.Version = strings.TrimPrefix(packageInfo.Version, "=")
	return packageInfo.BuildInfoModuleId()
}

func validatePnpmLocalBuildInfo(t *testing.T, buildName, buildNumber, moduleName string) {
	buildInfoService := build.CreateBuildInfoService()
	pnpmBuild, err := buildInfoService.GetOrCreateBuildWithProject(buildName, buildNumber, "")
	assert.NoError(t, err)
	bi, err := pnpmBuild.ToBuildInfo()
	assert.NoError(t, err)
	assert.NotEmpty(t, bi.Started)
	if assert.Len(t, bi.Modules, 1) {
		assert.Equal(t, moduleName, bi.Modules[0].Id)
		assert.Equal(t, buildinfo.Npm, bi.Modules[0].Type)
	}
}

func validatePnpmInstall(t *testing.T, pt pnpmTestParams) {
	expectedDependencies := []expectedDependency{{id: "xml:1.0.1", scopes: []string{"prod"}}}
	if !strings.Contains(pt.pnpmArgs, "-prod") && !strings.Contains(pt.pnpmArgs, "--prod") {
		expectedDependencies = append(expectedDependencies, expectedDependency{id: "json:9.0.6", scopes: []string{"dev"}})
	}
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PnpmBuildName, pt.buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	if buildInfo.Modules == nil {
		assert.NotNil(t, buildInfo.Modules)
		return
	}
	assert.NotEmpty(t, buildInfo.Modules)
	equalDependenciesSlices(t, expectedDependencies, buildInfo.Modules[0].Dependencies)
}

func validatePnpmPublish(t *testing.T, pt pnpmTestParams) {
	// pnpm publish normalizes versions (strips "v" prefix), so use isNpm7=false for Artifactory path expectations
	verifyExistInArtifactoryByProps(tests.GetNpmDeployedArtifacts(false),
		tests.NpmRepo+"/*",
		fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", tests.PnpmBuildName, pt.buildNumber), t)
	// pnpm pack preserves the "v" prefix in the local tarball filename used for build info
	validatePnpmCommonPublish(t, pt, tests.GetNpmArtifactName(true, false))
}

func validatePnpmScopedPublish(t *testing.T, pt pnpmTestParams) {
	// pnpm publish normalizes versions (strips "=" prefix), so use isNpm7=false for Artifactory path expectations
	verifyExistInArtifactoryByProps(tests.GetNpmDeployedScopedArtifacts(pt.repo, false),
		pt.repo+"/*",
		fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", tests.PnpmBuildName, pt.buildNumber), t)
	// pnpm pack includes the scope in the tarball filename (e.g., jscope-jfrog-cli-tests-=1.0.0.tgz)
	validatePnpmCommonPublish(t, pt, "jscope-jfrog-cli-tests-=1.0.0.tgz")
	// E2E: assert artifact Path in build info matches Artifactory layout for scoped packages (@scope/name/-/@scope/name-version.tgz)
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PnpmBuildName, pt.buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 && len(publishedBuildInfo.BuildInfo.Modules[0].Artifacts) > 0 {
		path := publishedBuildInfo.BuildInfo.Modules[0].Artifacts[0].Path
		assert.Equal(t, "@jscope/jfrog-cli-tests/-/@jscope/jfrog-cli-tests-1.0.0.tgz", path,
			"scoped artifact path in build info should match Artifactory layout")
	}
}

func validatePnpmCommonPublish(t *testing.T, pt pnpmTestParams, expectedArtifactName string) {
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PnpmBuildName, pt.buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	if len(buildInfo.Modules) == 0 {
		assert.Fail(t, "pnpm publish test failed", "params: \n%v \nexpected to have module with artifact: \n%v \nbut has no modules", pt, expectedArtifactName)
		return
	}
	assert.Len(t, buildInfo.Modules[0].Artifacts, 1, "pnpm publish test with params: \n%v \nexpected artifact: \n%v \nbut has: \n%v", pt, expectedArtifactName, buildInfo.Modules[0].Artifacts)
	assert.Equal(t, pt.moduleName, buildInfo.Modules[0].Id)
	assert.Equal(t, expectedArtifactName, buildInfo.Modules[0].Artifacts[0].Name)
}

func initPnpmFilesTest(t *testing.T) (pnpmProjectPath, pnpmScopedProjectPath string) {
	pnpmProjectPath = createPnpmProject(t, "pnpmproject")
	pnpmScopedProjectPath = createPnpmProject(t, "pnpmscopedproject")
	prepareArtifactoryForPnpmBuild(t, filepath.Dir(pnpmProjectPath))
	return
}

func initPnpmWorkspaceTest(t *testing.T) (pnpmWorkspacePath string) {
	pnpmWorkspacePath = filepath.Dir(createPnpmProject(t, "pnpmworkspace"))
	testFolder := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "pnpm", "pnpmworkspace")
	err := biutils.CopyDir(testFolder, pnpmWorkspacePath, true, []string{})
	assert.NoError(t, err)
	prepareArtifactoryForPnpmBuild(t, pnpmWorkspacePath)
	return
}

func prepareArtifactoryForPnpmBuild(t *testing.T, workingDirectory string) {
	clientTestUtils.ChangeDirAndAssert(t, workingDirectory)
	caches := ioutils.DoubleWinPathSeparator(filepath.Join(workingDirectory, "caches"))
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("pnpm", "install", "-store-dir="+caches))
	clientTestUtils.RemoveAllAndAssert(t, filepath.Join(workingDirectory, "node_modules"))
	clientTestUtils.RemoveAllAndAssert(t, caches)
}

func createPnpmProject(t *testing.T, dir string) string {
	srcPackageJson := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "pnpm", dir, "package.json")
	targetPackageJson := filepath.Join(tests.Out, dir)
	packageJson, err := tests.ReplaceTemplateVariables(srcPackageJson, targetPackageJson)
	assert.NoError(t, err)
	packageJson, err = filepath.Abs(packageJson)
	assert.NoError(t, err)
	return packageJson
}

// TestPnpmInstallWithoutBuildInfo verifies install succeeds when build-name/number are missing (RTECO-907).
func TestPnpmInstallWithoutBuildInfo(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	pnpmProjectPath := createPnpmProject(t, "pnpmproject")
	projectDir := filepath.Dir(pnpmProjectPath)
	prepareArtifactoryForPnpmBuild(t, projectDir)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)
	// No --build-name or --build-number: install should succeed, build info not collected
	err = runJfrogCliWithoutAssertion("pnpm", "install", "--store-dir="+tempCacheDirPath)
	assert.NoError(t, err)
}

// TestPnpmInstallOnlyAllow verifies install succeeds when the project contains an 'only-allow pnpm' preinstall script (RTECO-920).
func TestPnpmInstallOnlyAllow(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	pnpmProjectPath := createPnpmProject(t, "pnpmprojectonlyallow")
	projectDir := filepath.Dir(pnpmProjectPath)
	prepareArtifactoryForPnpmBuild(t, projectDir)

	buildNumber := "600"
	clientTestUtils.ChangeDirAndAssert(t, projectDir)
	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)

	moduleName := readPnpmModuleId(t, pnpmProjectPath)
	validatePnpmLocalBuildInfo(t, tests.PnpmBuildName, buildNumber, moduleName)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.PnpmBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PnpmBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

// TestPnpmAddCommand verifies that 'pnpm add' does not return an error with or without build info flags (RTECO-918).
func TestPnpmAddCommand(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	pnpmProjectPath := createPnpmProject(t, "pnpmproject")
	projectDir := filepath.Dir(pnpmProjectPath)
	prepareArtifactoryForPnpmBuild(t, projectDir)

	t.Run("without build flags", func(t *testing.T) {
		clientTestUtils.ChangeDirAndAssert(t, projectDir)
		err := runJfrogCliWithoutAssertion("pnpm", "add", "xml@1.0.1", "--store-dir="+tempCacheDirPath)
		assert.NoError(t, err, "pnpm add without build flags should not return an error")
	})

	t.Run("with build flags", func(t *testing.T) {
		clientTestUtils.ChangeDirAndAssert(t, projectDir)
		err := runJfrogCliWithoutAssertion("pnpm", "add", "xml@1.0.1", "--store-dir="+tempCacheDirPath,
			"--build-name="+tests.PnpmBuildName, "--build-number=603")
		assert.NoError(t, err, "pnpm add with build flags should not return an error")
	})
}

// TestPnpmInstallWithServerID verifies that --server-id selects the correct server configuration
// for build info collection. It creates a dummy default config (invalid creds) and a valid
// secondary config, then runs pnpm install with --server-id pointing to the valid config.
// If --server-id were ignored, build info collection would fail due to the broken default.
func TestPnpmInstallWithServerID(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	// Set up pnpm project and warm the Artifactory cache (uses valid "default" config).
	pnpmProjectPath := createPnpmProject(t, "pnpmproject")
	projectDir := filepath.Dir(pnpmProjectPath)
	prepareArtifactoryForPnpmBuild(t, projectDir)

	// Create a second server config ("pnpm-valid-server") with valid credentials.
	const validServerID = "pnpm-valid-server"
	configCli := coretests.NewJfrogCli(execMain, "jfrog config", "")
	var validCreds string
	if *tests.JfrogAccessToken != "" {
		validCreds = "--access-token=" + *tests.JfrogAccessToken
	} else {
		validCreds = "--user=" + *tests.JfrogUser + " --password=" + *tests.JfrogPassword
	}
	err = coretests.NewJfrogCli(execMain, "jfrog config", validCreds).Exec(
		"add", validServerID, "--interactive=false", "--url="+*tests.JfrogUrl, "--enc-password=false")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, configCli.Exec("rm", validServerID, "--quiet"))
	}()

	// Remove the "default" config, then re-add with dummy/invalid credentials so it cannot reach Artifactory.
	err = coretests.NewJfrogCli(execMain, "jfrog config", "").Exec("rm", "default", "--quiet")
	assert.NoError(t, err)
	err = coretests.NewJfrogCli(execMain, "jfrog config", "--access-token=invalid-token").Exec(
		"add", "default", "--interactive=false", "--url="+*tests.JfrogUrl, "--enc-password=false")
	assert.NoError(t, err)

	// Run pnpm install with --server-id pointing to the valid config.
	buildNumber := "700"
	clientTestUtils.ChangeDirAndAssert(t, projectDir)
	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--server-id="+validServerID,
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)

	// Validate local build info was collected.
	moduleName := readPnpmModuleId(t, pnpmProjectPath)
	validatePnpmLocalBuildInfo(t, tests.PnpmBuildName, buildNumber, moduleName)

	// Publish build info using the valid server and verify it was published with dependencies.
	assert.NoError(t, coretests.NewJfrogCli(execMain, "jfrog rt", validCreds+" --url="+serverDetails.ArtifactoryUrl).
		Exec("bp", tests.PnpmBuildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PnpmBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info should be found after publish")
	if assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules, "build info should contain modules") {
		assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies, "module should contain dependencies")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

// TestPnpmReleaseBundleCreation verifies successful creation of a Release Bundle using pnpm build info (RTECO-910).
func TestPnpmReleaseBundleCreation(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	pnpmProjectPath := createPnpmProject(t, "pnpmproject")
	projectDir := filepath.Dir(pnpmProjectPath)
	prepareArtifactoryForPnpmBuild(t, projectDir)

	buildNumber := "615"
	clientTestUtils.ChangeDirAndAssert(t, projectDir)
	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.PnpmBuildName, buildNumber))

	// Verify the published build info has dependencies before creating a release bundle
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PnpmBuildName, buildNumber)
	assert.NoError(t, err)
	if !assert.True(t, found, "build info should be found after publish") {
		return
	}
	assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules, "build info should have modules")
	assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies, "build info module should have dependencies")

	// Create a release bundle from the pnpm build info
	rbName := "pnpm-rb-creation-test"
	rbVersion := "1.0.0"
	err = runJfrogCliWithoutAssertion("rbc", rbName, rbVersion, "--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)
	if err != nil {
		// Release bundle operations require Distribution or lifecycle service - skip if unavailable
		t.Skipf("Skipping release bundle test: %v", err)
	}

	// Verify the release bundle was created by checking its status
	lcManager, err := utils.CreateLifecycleServiceManager(serverDetails, false)
	if assert.NoError(t, err) {
		rbDetails := services.ReleaseBundleDetails{
			ReleaseBundleName:    rbName,
			ReleaseBundleVersion: rbVersion,
		}
		resp, err := lcManager.GetReleaseBundleCreationStatus(rbDetails, "", true)
		if assert.NoError(t, err) {
			assert.Equal(t, services.Completed, resp.Status, "release bundle creation status should be COMPLETED")
		}
		// Clean up the release bundle
		_ = lcManager.DeleteReleaseBundleVersion(rbDetails, services.CommonOptionalQueryParams{Async: false})
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

// TestPnpmInstallTransitiveDependencies verifies that transitive dependencies are correctly resolved and included in the build info (RTECO-904).
func TestPnpmInstallTransitiveDependencies(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	pnpmProjectPath := createPnpmProject(t, "pnpmtransitive")
	projectDir := filepath.Dir(pnpmProjectPath)
	prepareArtifactoryForPnpmBuild(t, projectDir)

	buildNumber := "617"
	clientTestUtils.ChangeDirAndAssert(t, projectDir)
	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)

	assert.NoError(t, artifactoryCli.Exec("bp", tests.PnpmBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PnpmBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules) {
		deps := publishedBuildInfo.BuildInfo.Modules[0].Dependencies
		// chalk@2.4.2 has transitive deps: ansi-styles, escape-string-regexp, supports-color, color-convert, has-flag, color-name
		assert.Greater(t, len(deps), 1, "should have transitive dependencies beyond the direct dependency (chalk)")

		// Verify that at least one known transitive dependency is present
		hasTransitiveDep := false
		for _, dep := range deps {
			if strings.HasPrefix(dep.Id, "ansi-styles:") || strings.HasPrefix(dep.Id, "supports-color:") || strings.HasPrefix(dep.Id, "color-convert:") {
				hasTransitiveDep = true
				break
			}
		}
		assert.True(t, hasTransitiveDep, "build info should include transitive dependencies of chalk")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

// TestPnpmInstallEmptyLockfile verifies that the CLI correctly handles an empty pnpm-lock.yaml (RTECO-903).
func TestPnpmInstallEmptyLockfile(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	pnpmProjectPath := createPnpmProject(t, "pnpmemptylockfile")
	projectDir := filepath.Dir(pnpmProjectPath)

	// Also copy the empty lockfile to the target directory
	srcLockfile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "pnpm", "pnpmemptylockfile", "pnpm-lock.yaml")
	targetLockfile := filepath.Join(projectDir, "pnpm-lock.yaml")
	lockfileContent, err := os.ReadFile(srcLockfile)
	assert.NoError(t, err)
	err = os.WriteFile(targetLockfile, lockfileContent, 0644)
	assert.NoError(t, err)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)
	// Install should succeed even with an empty lockfile (pnpm regenerates it)
	buildNumber := "618"
	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--build-name="+tests.PnpmBuildName, "--build-number="+buildNumber)

	moduleName := readPnpmModuleId(t, pnpmProjectPath)
	validatePnpmLocalBuildInfo(t, tests.PnpmBuildName, buildNumber, moduleName)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PnpmBuildName, artHttpDetails)
}

// TestPnpmBuildPublishWithCIVcsProps verifies that CI VCS properties (vcs.provider, vcs.org, vcs.repo)
// are set on artifacts published via pnpm after build-publish (RTECO-923).
func TestPnpmBuildPublishWithCIVcsProps(t *testing.T) {
	initPnpmTest(t)
	defer cleanPnpmTest(t)

	buildName := "pnpm-civcs-test"
	buildNumber := "1"

	// Setup GitHub Actions environment (uses real env vars on CI, mock values locally)
	cleanupEnv, actualOrg, actualRepo := tests.SetupGitHubActionsEnv(t)
	defer cleanupEnv()

	// Clean old build
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	// Setup pnpm project
	pnpmProjectPath := createPnpmProject(t, "pnpmproject")
	projectDir := filepath.Dir(pnpmProjectPath)
	prepareArtifactoryForPnpmBuild(t, projectDir)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Run pnpm publish with build info collection
	cleanupAuth := setupPnpmPublishAuth(t, tests.NpmRepo)
	defer cleanupAuth()
	runJfrogCli(t, "pnpm", "publish", "--no-git-checks",
		"--build-name="+buildName, "--build-number="+buildNumber)

	// Publish build info - should set CI VCS props on artifacts
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Restore working directory before getting build info
	clientTestUtils.ChangeDirAndAssert(t, wd)

	// Get the published build info to find artifact paths
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "Build info was not found")

	// Create service manager for getting artifact properties
	serviceManager, err := utils.CreateServiceManager(serverDetails, 3, 1000, false)
	assert.NoError(t, err)

	// Verify VCS properties on each artifact from build info
	artifactCount := 0
	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		for _, artifact := range module.Artifacts {
			fullPath := artifact.OriginalDeploymentRepo + "/" + artifact.Path
			if artifact.OriginalDeploymentRepo == "" || artifact.Path == "" {
				t.Logf("Artifact %s missing OriginalDeploymentRepo or Path, skipping", artifact.Name)
				continue
			}

			props, err := serviceManager.GetItemProps(fullPath)
			assert.NoError(t, err, "Failed to get properties for artifact: %s", fullPath)
			if props == nil {
				assert.Fail(t, "Properties are nil for artifact: %s", fullPath)
				continue
			}

			// Validate VCS properties
			assert.Contains(t, props.Properties, "vcs.provider", "Missing vcs.provider on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.provider"], "github", "Wrong vcs.provider on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.org", "Missing vcs.org on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.org"], actualOrg, "Wrong vcs.org on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.repo", "Missing vcs.repo on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.repo"], actualRepo, "Wrong vcs.repo on %s", artifact.Name)

			artifactCount++
		}
	}
	assert.Greater(t, artifactCount, 0, "No artifacts in build info")
}

// TestPnpmInstallAndPublishWithProject verifies that pnpm install and publish work correctly
// when targeting a non-default Artifactory project (RTECO-924).
// The test uses --project flag with install, publish, and build-publish to verify that
// build info is correctly stored and published under the specified project key.
func TestPnpmInstallAndPublishWithProject(t *testing.T) {
	initPnpmTest(t)

	// Create Access service manager and project before deferring cleanPnpmTest,
	// so that t.Skipf doesn't trigger cleanup asserts that override the skip status.
	accessManager, err := utils.CreateAccessServiceManager(serverDetails, false)
	if err != nil {
		t.Skipf("Skipping project test - cannot create access manager: %v", err)
	}

	// Try creating project first to verify access works before deferring any cleanup
	projectParams := accessServices.ProjectParams{
		ProjectDetails: accessServices.Project{
			DisplayName: "pnpm-project-test " + tests.ProjectKey,
			ProjectKey:  tests.ProjectKey,
		},
	}
	// First delete if exists, ignoring errors since access might not support it
	_ = accessManager.DeleteProject(tests.ProjectKey)
	if err = accessManager.CreateProject(projectParams); err != nil {
		t.Skipf("Skipping project test - cannot create project: %v", err)
	}

	defer cleanPnpmTest(t)
	defer func() {
		_ = accessManager.UnassignRepoFromProject(tests.NpmRepo)
		_ = accessManager.UnassignRepoFromProject(tests.NpmRemoteRepo)
		_ = accessManager.DeleteProject(tests.ProjectKey)
	}()

	// Assign npm repos to the project
	err = accessManager.AssignRepoToProject(tests.NpmRepo, tests.ProjectKey, true)
	assert.NoError(t, err)
	err = accessManager.AssignRepoToProject(tests.NpmRemoteRepo, tests.ProjectKey, true)
	assert.NoError(t, err)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	buildName := tests.PnpmBuildName + "-project"
	buildNumber := "800"

	// Clean old build
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	// Setup pnpm project
	pnpmProjectPath := createPnpmProject(t, "pnpmproject")
	projectDir := filepath.Dir(pnpmProjectPath)
	prepareArtifactoryForPnpmBuild(t, projectDir)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Run pnpm install with --project flag
	runJfrogCli(t, "pnpm", "install", "--store-dir="+tempCacheDirPath,
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--project="+tests.ProjectKey)

	// Run pnpm publish with --project flag
	cleanupAuth := setupPnpmPublishAuth(t, tests.NpmRepo)
	defer cleanupAuth()
	runJfrogCli(t, "pnpm", "publish", "--no-git-checks",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--project="+tests.ProjectKey)

	// Publish build info with --project flag
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber, "--project="+tests.ProjectKey))

	// Restore working directory
	clientTestUtils.ChangeDirAndAssert(t, wd)

	// Get the published build info with project key
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	assert.NoError(t, err)
	params := artServices.NewBuildInfoParams()
	params.BuildName = buildName
	params.BuildNumber = buildNumber
	params.ProjectKey = tests.ProjectKey
	publishedBuildInfo, found, err := servicesManager.GetBuildInfo(params)
	assert.NoError(t, err)
	assert.True(t, found, "Build info was not found for project %s", tests.ProjectKey)

	bi := publishedBuildInfo.BuildInfo
	// pnpm install + publish on the same build should produce 1 module with both deps and artifacts
	if assert.NotEmpty(t, bi.Modules, "Build info should contain modules") {
		assert.NotEmpty(t, bi.Modules[0].Dependencies, "Module should have dependencies from pnpm install")
		assert.NotEmpty(t, bi.Modules[0].Artifacts, "Module should have artifacts from pnpm publish")
	}
}

func initPnpmTest(t *testing.T) {
	if !*tests.TestPnpm {
		t.Skip("Skipping Pnpm test. To run Pnpm test add the '-test.pnpm=true' option.")
	}
	_ = os.Unsetenv("JFROG_RUN_NATIVE")
	createJfrogHomeConfig(t, true)
}
