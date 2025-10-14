package tests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	ioutils "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-client-go/utils/tests"

	"github.com/urfave/cli"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	commandutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/tests"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	corelog "github.com/jfrog/jfrog-cli-core/v2/utils/log"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/summary"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

var (
	JfrogUrl                  *string
	JfrogUser                 *string
	JfrogPassword             *string
	JfrogSshKeyPath           *string
	JfrogSshPassphrase        *string
	JfrogAccessToken          *string
	JfrogTargetUrl            *string
	JfrogTargetAccessToken    *string
	JfrogHome                 *string
	TestArtifactoryProject    *bool
	TestArtifactory           *bool
	TestArtifactoryProxy      *bool
	TestDistribution          *bool
	TestDocker                *bool
	TestPodman                *bool
	TestDockerScan            *bool
	ContainerRegistry         *string
	TestGo                    *bool
	TestNpm                   *bool
	TestGradle                *bool
	TestMaven                 *bool
	TestNuget                 *bool
	TestPip                   *bool
	TestPipenv                *bool
	TestPoetry                *bool
	TestPlugins               *bool
	TestXray                  *bool
	TestAccess                *bool
	TestTransfer              *bool
	TestLifecycle             *bool
	HideUnitTestLog           *bool
	ciRunId                   *string
	InstallDataTransferPlugin *bool
	timestampAdded            bool
)

func init() {
	JfrogUrl = flag.String("jfrog.url", "http://localhost:8081/", "JFrog platform url")
	JfrogUser = flag.String("jfrog.user", "admin", "JFrog platform  username")
	JfrogPassword = flag.String("jfrog.password", "password", "JFrog platform password")
	JfrogSshKeyPath = flag.String("jfrog.sshKeyPath", "", "Ssh key file path")
	JfrogSshPassphrase = flag.String("jfrog.sshPassphrase", "", "Ssh key passphrase")
	JfrogAccessToken = flag.String("jfrog.adminToken", tests.GetLocalArtifactoryTokenIfNeeded(*JfrogUrl), "JFrog platform admin token")
	JfrogTargetUrl = flag.String("jfrog.targetUrl", "", "JFrog target platform url for transfer tests")
	JfrogTargetAccessToken = flag.String("jfrog.targetAdminToken", "", "JFrog target platform admin token for transfer tests")
	JfrogHome = flag.String("jfrog.home", "", "The JFrog home directory of the local Artifactory installation")
	TestArtifactory = flag.Bool("test.artifactory", false, "Test Artifactory")
	TestArtifactoryProject = flag.Bool("test.artifactoryProject", false, "Test Artifactory project")
	TestArtifactoryProxy = flag.Bool("test.artifactoryProxy", false, "Test Artifactory proxy")
	TestDistribution = flag.Bool("test.distribution", false, "Test distribution")
	TestDocker = flag.Bool("test.docker", false, "Test Docker build")
	TestDockerScan = flag.Bool("test.dockerScan", false, "Test Docker scan")
	TestPodman = flag.Bool("test.podman", false, "Test Podman build")
	TestGo = flag.Bool("test.go", false, "Test Go")
	TestNpm = flag.Bool("test.npm", false, "Test Npm")
	TestGradle = flag.Bool("test.gradle", false, "Test Gradle")
	TestMaven = flag.Bool("test.maven", false, "Test Maven")
	TestNuget = flag.Bool("test.nuget", false, "Test Nuget")
	TestPip = flag.Bool("test.pip", false, "Test Pip")
	TestPipenv = flag.Bool("test.pipenv", false, "Test Pipenv")
	TestPoetry = flag.Bool("test.poetry", false, "Test Poetry")
	TestPlugins = flag.Bool("test.plugins", false, "Test Plugins")
	TestXray = flag.Bool("test.xray", false, "Test Xray")
	TestAccess = flag.Bool("test.access", false, "Test Access")
	TestTransfer = flag.Bool("test.transfer", false, "Test files transfer")
	TestLifecycle = flag.Bool("test.lifecycle", false, "Test lifecycle")
	ContainerRegistry = flag.String("test.containerRegistry", "localhost:8082", "Container registry")
	HideUnitTestLog = flag.Bool("test.hideUnitTestLog", false, "Hide unit tests logs and print it in a file")
	InstallDataTransferPlugin = flag.Bool("test.installDataTransferPlugin", false, "Install data-transfer plugin on the source Artifactory server")
	ciRunId = flag.String("ci.runId", "", "A unique identifier used as a suffix to create repositories and builds in the tests")
}

func CleanFileSystem() {
	removeDirs(Out, Temp)
}

func removeDirs(dirs ...string) {
	for _, dir := range dirs {
		isExist, err := fileutils.IsDirExists(dir, false)
		if err != nil {
			log.Error(err)
		}
		if isExist {
			err = fileutils.RemoveTempDir(dir)
			if err != nil {
				log.Error(errors.New("Cannot remove path: " + dir + " due to: " + err.Error()))
			}
		}
	}
}

func VerifyExistLocally(expected, actual []string, t *testing.T) {
	if len(actual) == 0 && len(expected) != 0 {
		t.Error("Couldn't find all expected files, expected: " + strconv.Itoa(len(expected)) + ", found: " + strconv.Itoa(len(actual)))
	}
	err := compare(expected, actual)
	assert.NoError(t, err)
}

func ValidateListsIdentical(expected, actual []string) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("unexpected behavior, \nexpected: [%s], \nfound:    [%s]", strings.Join(expected, ", "), strings.Join(actual, ", "))
	}
	err := compare(expected, actual)
	return err
}

func ValidateChecksums(filePath string, expectedChecksum buildinfo.Checksum, t *testing.T) {
	localFileDetails, err := fileutils.GetFileDetails(filePath, true)
	if err != nil {
		t.Error("Couldn't calculate sha1, " + err.Error())
	}
	if localFileDetails.Checksum.Sha1 != expectedChecksum.Sha1 {
		t.Error("sha1 mismatch for "+filePath+", expected: "+expectedChecksum.Sha1, "found: "+localFileDetails.Checksum.Sha1)
	}
	if localFileDetails.Checksum.Md5 != expectedChecksum.Md5 {
		t.Error("md5 mismatch for "+filePath+", expected: "+expectedChecksum.Md5, "found: "+localFileDetails.Checksum.Sha1)
	}
	if localFileDetails.Checksum.Sha256 != expectedChecksum.Sha256 {
		t.Error("sha256 mismatch for "+filePath+", expected: "+expectedChecksum.Sha256, "found: "+localFileDetails.Checksum.Sha1)
	}
}

func compare(expected, actual []string) error {
	for _, v := range expected {
		for i, r := range actual {
			if v == r {
				break
			}
			if i == len(actual)-1 {
				return errors.New("Missing file : " + v)
			}
		}
	}
	return nil
}

func getPathsFromSearchResults(searchResults []artUtils.SearchResult) []string {
	var paths []string
	for _, result := range searchResults {
		paths = append(paths, result.Path)
	}
	return paths
}

func CompareExpectedVsActual(expected []string, actual []artUtils.SearchResult, t *testing.T) {
	actualPaths := getPathsFromSearchResults(actual)
	assert.ElementsMatch(t, expected, actualPaths, fmt.Sprintf("Expected: %v \nActual: %v", expected, actualPaths))
}

func GetTestResourcesPath() string {
	dir, _ := os.Getwd()
	return filepath.ToSlash(dir + "/testdata/")
}

func GetFilePathForArtifactory(fileName string) string {
	return GetTestResourcesPath() + "filespecs/" + fileName
}

func GetTestsLogsDir() (string, error) {
	tempDirPath := filepath.Join(coreutils.GetCliPersistentTempDirPath(), "jfrog_tests_logs")
	return tempDirPath, fileutils.CreateDirIfNotExist(tempDirPath)
}

type PackageSearchResultItem struct {
	Name      string
	Path      string
	Package   string
	Version   string
	Repo      string
	Owner     string
	Created   string
	Size      int64
	Sha1      string
	Published bool
}

func DeleteFiles(deleteSpec *spec.SpecFiles, serverDetails *config.ServerDetails) (successCount, failCount int, err error) {
	deleteCommand := generic.NewDeleteCommand()
	deleteCommand.SetThreads(3).SetSpec(deleteSpec).SetServerDetails(serverDetails).SetDryRun(false)
	reader, err := deleteCommand.GetPathsToDelete()
	if err != nil {
		return 0, 0, err
	}
	defer ioutils.Close(reader, &err)
	return deleteCommand.DeleteFiles(reader)
}

// This function makes no assertion, caller is responsible to assert as needed.
func GetBuildInfo(serverDetails *config.ServerDetails, buildName, buildNumber string) (pbi *buildinfo.PublishedBuildInfo, found bool, err error) {
	servicesManager, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return nil, false, err
	}
	params := services.NewBuildInfoParams()
	params.BuildName = buildName
	params.BuildNumber = buildNumber
	return servicesManager.GetBuildInfo(params)
}

func GetBuildRuns(serverDetails *config.ServerDetails, buildName string) (pbi *buildinfo.BuildRuns, found bool, err error) {
	servicesManager, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return nil, false, err
	}
	params := services.NewBuildInfoParams()
	params.BuildName = buildName
	return servicesManager.GetBuildRuns(params)
}

var reposConfigMap = map[*string]string{
	&DistRepo1:                      DistributionRepoConfig1,
	&DistRepo2:                      DistributionRepoConfig2,
	&GoRepo:                         GoLocalRepositoryConfig,
	&GoRemoteRepo:                   GoRemoteRepositoryConfig,
	&GoVirtualRepo:                  GoVirtualRepositoryConfig,
	&GradleRepo:                     GradleRepositoryConfig,
	&MvnRepo1:                       MavenRepositoryConfig1,
	&MvnRepo2:                       MavenRepositoryConfig2,
	&MvnRemoteRepo:                  MavenRemoteRepositoryConfig,
	&GradleRemoteRepo:               GradleRemoteRepositoryConfig,
	&NpmRepo:                        NpmLocalRepositoryConfig,
	&NpmScopedRepo:                  NpmLocalScopedRespositoryConfig,
	&NpmRemoteRepo:                  NpmRemoteRepositoryConfig,
	&NugetRemoteRepo:                NugetRemoteRepositoryConfig,
	&YarnRemoteRepo:                 YarnRemoteRepositoryConfig,
	&PypiLocalRepo:                  PypiLocalRepositoryConfig,
	&PypiRemoteRepo:                 PypiRemoteRepositoryConfig,
	&PypiVirtualRepo:                PypiVirtualRepositoryConfig,
	&PipenvRemoteRepo:               PipenvRemoteRepositoryConfig,
	&PipenvVirtualRepo:              PipenvVirtualRepositoryConfig,
	&PoetryLocalRepo:                PoetryLocalRepositoryConfig,
	&PoetryRemoteRepo:               PoetryRemoteRepositoryConfig,
	&PoetryVirtualRepo:              PoetryVirtualRepositoryConfig,
	&RtDebianRepo:                   DebianTestRepositoryConfig,
	&RtLfsRepo:                      GitLfsTestRepositoryConfig,
	&RtRepo1:                        Repo1RepositoryConfig,
	&RtRepo2:                        Repo2RepositoryConfig,
	&RtVirtualRepo:                  VirtualRepositoryConfig,
	&TerraformRepo:                  TerraformLocalRepositoryConfig,
	&DockerLocalRepo:                DockerLocalRepositoryConfig,
	&DockerLocalPromoteRepo:         DockerLocalPromoteRepositoryConfig,
	&DockerRemoteRepo:               DockerRemoteRepositoryConfig,
	&DockerVirtualRepo:              DockerVirtualRepositoryConfig,
	&RtDevRepo:                      DevRepoRepositoryConfig,
	&RtProdRepo1:                    ProdRepo1RepositoryConfig,
	&RtProdRepo2:                    ProdRepo2RepositoryConfig,
	&ReleaseLifecycleDependencyRepo: ReleaseLifecycleImportDependencySpec,
}

var (
	CreatedNonVirtualRepositories map[*string]string
	CreatedVirtualRepositories    map[*string]string
)

func getNeededRepositories(reposMap map[*bool][]*string) map[*string]string {
	reposToCreate := map[*string]string{}
	for needed, testRepos := range reposMap {
		if *needed {
			for _, repo := range testRepos {
				reposToCreate[repo] = reposConfigMap[repo]
			}
		}
	}
	return reposToCreate
}

func getNeededBuildNames(buildNamesMap map[*bool][]*string) []string {
	var neededBuildNames []string
	for needed, buildNames := range buildNamesMap {
		if *needed {
			for _, buildName := range buildNames {
				neededBuildNames = append(neededBuildNames, *buildName)
			}
		}
	}
	return neededBuildNames
}

// Return local and remote repositories for the test suites, respectfully
func GetNonVirtualRepositories() map[*string]string {
	nonVirtualReposMap := map[*bool][]*string{
		TestArtifactory:        {&RtRepo1, &RtRepo2, &RtLfsRepo, &RtDebianRepo, &TerraformRepo, &ReleaseLifecycleDependencyRepo},
		TestArtifactoryProject: {&RtRepo1, &RtRepo2, &RtLfsRepo, &RtDebianRepo},
		TestDistribution:       {&DistRepo1, &DistRepo2},
		TestDocker:             {&DockerLocalRepo, &DockerLocalPromoteRepo, &DockerRemoteRepo},
		TestDockerScan:         {&DockerLocalRepo, &DockerLocalPromoteRepo, &DockerRemoteRepo},
		TestPodman:             {&DockerLocalRepo, &DockerLocalPromoteRepo, &DockerRemoteRepo},
		TestGo:                 {&GoRepo, &GoRemoteRepo},
		TestGradle:             {&GradleRepo, &GradleRemoteRepo},
		TestMaven:              {&MvnRepo1, &MvnRepo2, &MvnRemoteRepo},
		TestNpm:                {&NpmRepo, &NpmScopedRepo, &NpmRemoteRepo},
		TestNuget:              {&NugetRemoteRepo},
		TestPip:                {&PypiLocalRepo, &PypiRemoteRepo},
		TestPipenv:             {&PipenvRemoteRepo},
		TestPoetry:             {&PoetryLocalRepo, &PoetryRemoteRepo},
		TestPlugins:            {&RtRepo1},
		TestXray:               {&NpmRemoteRepo, &NugetRemoteRepo, &YarnRemoteRepo, &GradleRemoteRepo, &MvnRemoteRepo, &GoRepo, &GoRemoteRepo, &PypiRemoteRepo},
		TestAccess:             {&RtRepo1},
		TestTransfer:           {&RtRepo1, &RtRepo2, &MvnRepo1, &MvnRemoteRepo, &DockerRemoteRepo},
		TestLifecycle:          {&RtDevRepo, &RtProdRepo1, &RtProdRepo2},
	}
	return getNeededRepositories(nonVirtualReposMap)
}

// Return virtual repositories for the test suites, respectfully
func GetVirtualRepositories() map[*string]string {
	virtualReposMap := map[*bool][]*string{
		TestArtifactory:  {&RtVirtualRepo},
		TestDistribution: {},
		TestDocker:       {&DockerVirtualRepo},
		TestDockerScan:   {&DockerVirtualRepo},
		TestPodman:       {&DockerVirtualRepo},
		TestGo:           {&GoVirtualRepo},
		TestGradle:       {},
		TestMaven:        {},
		TestNpm:          {},
		TestNuget:        {},
		TestPip:          {&PypiVirtualRepo},
		TestPipenv:       {&PipenvVirtualRepo},
		TestPoetry:       {&PoetryVirtualRepo},
		TestPlugins:      {},
		TestXray:         {&GoVirtualRepo},
		TestAccess:       {},
	}
	return getNeededRepositories(virtualReposMap)
}

func GetAllRepositoriesNames() []string {
	var baseRepoNames []string
	for repoName := range GetNonVirtualRepositories() {
		baseRepoNames = append(baseRepoNames, *repoName)
	}
	for repoName := range GetVirtualRepositories() {
		baseRepoNames = append(baseRepoNames, *repoName)
	}
	return baseRepoNames
}

func GetTestUsersNames() []string {
	return []string{UserName1, UserName2}
}

func GetBuildNames() []string {
	buildNamesMap := map[*bool][]*string{
		TestArtifactory:  {&RtBuildName1, &RtBuildName2, &RtBuildNameWithSpecialChars},
		TestDistribution: {},
		TestDocker:       {&DockerBuildName},
		TestPodman:       {&DockerBuildName},
		TestGo:           {&GoBuildName},
		TestGradle:       {&GradleBuildName},
		TestMaven:        {&MvnBuildName},
		TestNpm:          {&NpmBuildName, &YarnBuildName},
		TestNuget:        {&NuGetBuildName},
		TestPip:          {&PipBuildName},
		TestPipenv:       {&PipenvBuildName},
		TestPoetry:       {&PoetryBuildName},
		TestPlugins:      {},
		TestXray:         {},
		TestAccess:       {},
		TestTransfer:     {&MvnBuildName},
		TestLifecycle:    {&LcBuildName1, &LcBuildName2, &LcBuildName3},
	}
	return getNeededBuildNames(buildNamesMap)
}

// Builds and repositories names to replace in the test files.
// We use substitution map to set repositories and builds with timestamp.
func getSubstitutionMap() map[string]string {
	return map[string]string{
		"${REPO1}":                     RtRepo1,
		"${REPO2}":                     RtRepo2,
		"${REPO_1_AND_2}":              RtRepo1And2,
		"${VIRTUAL_REPO}":              RtVirtualRepo,
		"${LFS_REPO}":                  RtLfsRepo,
		"${DEBIAN_REPO}":               RtDebianRepo,
		"${DOCKER_REPO}":               DockerLocalRepo,
		"${DOCKER_PROMOTE_REPO}":       DockerLocalPromoteRepo,
		"${DOCKER_REMOTE_REPO}":        DockerRemoteRepo,
		"${DOCKER_VIRTUAL_REPO}":       DockerVirtualRepo,
		"${DOCKER_IMAGE_NAME}":         DockerImageName,
		"${CONTAINER_REGISTRY_DOMAIN}": RtContainerHostName,
		"${MAVEN_REPO1}":               MvnRepo1,
		"${MAVEN_REPO2}":               MvnRepo2,
		"${MAVEN_REMOTE_REPO}":         MvnRemoteRepo,
		"${GRADLE_REMOTE_REPO}":        GradleRemoteRepo,
		"${GRADLE_REPO}":               GradleRepo,
		"${NPM_REPO}":                  NpmRepo,
		"${NPM_SCOPED_REPO}":           NpmScopedRepo,
		"${NPM_REMOTE_REPO}":           NpmRemoteRepo,
		"${NUGET_REMOTE_REPO}":         NugetRemoteRepo,
		"${YARN_REMOTE_REPO}":          YarnRemoteRepo,
		"${GO_REPO}":                   GoRepo,
		"${GO_REMOTE_REPO}":            GoRemoteRepo,
		"${GO_VIRTUAL_REPO}":           GoVirtualRepo,
		"${TERRAFORM_REPO}":            TerraformRepo,
		"${SERVER_ID}":                 ServerId,
		"${URL}":                       *JfrogUrl,
		"${USERNAME}":                  *JfrogUser,
		"${PASSWORD}":                  *JfrogPassword,
		"${RT_CREDENTIALS_BASIC_AUTH}": base64.StdEncoding.EncodeToString([]byte(*JfrogUser + ":" + *JfrogPassword)),
		"${ACCESS_TOKEN}":              *JfrogAccessToken,
		"${PYPI_LOCAL_REPO}":           PypiLocalRepo,
		"${PYPI_REMOTE_REPO}":          PypiRemoteRepo,
		"${PYPI_VIRTUAL_REPO}":         PypiVirtualRepo,
		"${PIPENV_REMOTE_REPO}":        PipenvRemoteRepo,
		"${PIPENV_VIRTUAL_REPO}":       PipenvVirtualRepo,
		"${POETRY_LOCAL_REPO}":         PoetryLocalRepo,
		"${POETRY_REMOTE_REPO}":        PoetryRemoteRepo,
		"${POETRY_VIRTUAL_REPO}":       PoetryVirtualRepo,
		"${BUILD_NAME1}":               RtBuildName1,
		"${BUILD_NAME2}":               RtBuildName2,
		"${BUNDLE_NAME}":               BundleName,
		"${DIST_REPO1}":                DistRepo1,
		"${DIST_REPO2}":                DistRepo2,
		"{USER_NAME_1}":                UserName1,
		"{PASSWORD_1}":                 Password1,
		"{USER_NAME_2}":                UserName2,
		"{PASSWORD_2}":                 Password2,
		"${LC_BUILD_NAME1}":            LcBuildName1,
		"${LC_BUILD_NAME2}":            LcBuildName2,
		"${LC_BUILD_NAME3}":            LcBuildName3,
		"${RB_NAME1}":                  LcRbName1,
		"${RB_NAME2}":                  LcRbName2,
		"${DEV_REPO}":                  RtDevRepo,
		"${PROD_REPO1}":                RtProdRepo1,
		"${PROD_REPO2}":                RtProdRepo2,
	}
}

// Add timestamp to builds and repositories names
func AddTimestampToGlobalVars() {
	// Make sure the global timestamp is added only once even in case of multiple tests flags
	if timestampAdded {
		return
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	uniqueSuffix := "-" + timestamp
	if *ciRunId != "" {
		uniqueSuffix = "-" + *ciRunId + uniqueSuffix
	}
	// Artifactory accepts only lowercase repository names
	uniqueSuffix = strings.ToLower(uniqueSuffix)

	// Repositories
	DistRepo1 += uniqueSuffix
	DistRepo2 += uniqueSuffix
	GoRepo += uniqueSuffix
	GoRemoteRepo += uniqueSuffix
	GoVirtualRepo += uniqueSuffix
	DockerLocalRepo += uniqueSuffix
	DockerLocalPromoteRepo += uniqueSuffix
	DockerRemoteRepo += uniqueSuffix
	DockerVirtualRepo += uniqueSuffix
	TerraformRepo += uniqueSuffix
	GradleRemoteRepo += uniqueSuffix
	GradleRepo += uniqueSuffix
	MvnRemoteRepo += uniqueSuffix
	MvnRepo1 += uniqueSuffix
	MvnRepo2 += uniqueSuffix
	NpmRepo += uniqueSuffix
	NpmScopedRepo += uniqueSuffix
	NpmRemoteRepo += uniqueSuffix
	NugetRemoteRepo += uniqueSuffix
	YarnRemoteRepo += uniqueSuffix
	PypiLocalRepo += uniqueSuffix
	PypiRemoteRepo += uniqueSuffix
	PypiVirtualRepo += uniqueSuffix
	PipenvRemoteRepo += uniqueSuffix
	PipenvVirtualRepo += uniqueSuffix
	PoetryLocalRepo += uniqueSuffix
	PoetryRemoteRepo += uniqueSuffix
	PoetryVirtualRepo += uniqueSuffix
	RtDebianRepo += uniqueSuffix
	RtLfsRepo += uniqueSuffix
	RtRepo1 += uniqueSuffix
	RtRepo1And2 += uniqueSuffix
	RtRepo1And2Placeholder += uniqueSuffix
	RtRepo2 += uniqueSuffix
	RtVirtualRepo += uniqueSuffix
	RtDevRepo += uniqueSuffix
	RtProdRepo1 += uniqueSuffix
	RtProdRepo2 += uniqueSuffix

	// Builds/bundles/images
	BundleName += uniqueSuffix
	DockerBuildName += uniqueSuffix
	DockerImageName += uniqueSuffix
	DotnetBuildName += uniqueSuffix
	GoBuildName += uniqueSuffix
	GradleBuildName += uniqueSuffix
	NpmBuildName += uniqueSuffix
	YarnBuildName += uniqueSuffix
	MvnBuildName += uniqueSuffix
	NuGetBuildName += uniqueSuffix
	PipBuildName += uniqueSuffix
	PipenvBuildName += uniqueSuffix
	PoetryBuildName += uniqueSuffix
	RtBuildName1 += uniqueSuffix
	RtBuildName2 += uniqueSuffix
	RtBuildNameWithSpecialChars += uniqueSuffix
	RtPermissionTargetName += uniqueSuffix
	LcBuildName1 += uniqueSuffix
	LcBuildName2 += uniqueSuffix
	LcBuildName3 += uniqueSuffix
	LcRbName1 += uniqueSuffix
	LcRbName2 += uniqueSuffix
	LcRbName3 += uniqueSuffix

	// Users
	UserName1 += uniqueSuffix
	UserName2 += uniqueSuffix

	randomSequence := rand.New(rand.NewSource(time.Now().Unix()))
	Password1 += uniqueSuffix + strconv.FormatFloat(randomSequence.Float64(), 'f', 2, 32)
	Password2 += uniqueSuffix + strconv.FormatFloat(randomSequence.Float64(), 'f', 2, 32)

	// Projects
	ProjectKey += timestamp[len(timestamp)-7:]

	timestampAdded = true
}

// Replace all variables in the form of ${VARIABLE} in the input file, according to the substitution map (see getSubstitutionMap()).
// path - Path to the input file.
// destPath - Path to the output file. If empty, the output file will be under ${CWD}/tmp/.
func ReplaceTemplateVariables(path, destPath string) (string, error) {
	return commonCliUtils.ReplaceTemplateVariables(path, destPath, getSubstitutionMap())
}

func CreateSpec(fileName string) (string, error) {
	searchFilePath := GetFilePathForArtifactory(fileName)
	searchFilePath, err := ReplaceTemplateVariables(searchFilePath, "")
	return searchFilePath, err
}

func ConvertSliceToMap(props []utils.Property) map[string][]string {
	propsMap := make(map[string][]string)
	for _, item := range props {
		propsMap[item.Key] = append(propsMap[item.Key], item.Value)
	}
	return propsMap
}

// Set user and password from access token.
// Return the original user and password to allow restoring them in the end of the test.
func SetBasicAuthFromAccessToken() (string, string) {
	origUser := *JfrogUser
	origPassword := *JfrogPassword
	*JfrogUser = auth.ExtractUsernameFromAccessToken(*JfrogAccessToken)
	*JfrogPassword = *JfrogAccessToken
	return origUser, origPassword
}

// Clean items with timestamp older than 24 hours. Used to delete old repositories, builds, release bundles and Docker images.
// baseItemNames - The items to delete without timestamp, i.e. [cli-rt1, cli-rt2, ...]
// getActualItems - Function that returns all actual items in the remote server, i.e. [cli-rt1-1592990748, cli-rt2-1592990748, ...]
// deleteItem - Function that deletes the item by name
func CleanUpOldItems(baseItemNames []string, getActualItems func() ([]string, error), deleteItem func(string)) {
	actualItems, err := getActualItems()
	if err != nil {
		log.Warn("Couldn't retrieve items", err)
		return
	}
	now := time.Now()
	for _, baseItemName := range baseItemNames {
		itemPattern := regexp.MustCompile(`^` + baseItemName + `[\w-]*-(\d*)$`)
		for _, item := range actualItems {
			regexGroups := itemPattern.FindStringSubmatch(item)
			if regexGroups == nil {
				// Item does not match
				continue
			}

			itemTimestamp, err := strconv.ParseInt(regexGroups[len(regexGroups)-1], 10, 64)
			if err != nil {
				log.Warn("Error while parsing timestamp of", item, err)
				continue
			}

			itemTime := time.Unix(itemTimestamp, 0)
			if now.Sub(itemTime).Hours() > 24 {
				deleteItem(item)
			}
		}
	}
}

// Set new logger with output redirection to a null logger. This is useful for negative tests.
// Caller is responsible to set the old log back.
func RedirectLogOutputToNil() (previousLog log.Log) {
	previousLog = log.Logger
	newLog := log.NewLogger(corelog.GetCliLogLevel(), nil)
	newLog.SetOutputWriter(io.Discard)
	newLog.SetLogsWriter(io.Discard, 0)
	log.SetLogger(newLog)
	return previousLog
}

// Redirect output to a file, execute the command and read output.
// The reason for redirecting to a file and not to a buffer is the limited
// size of the buffer while using os.Pipe.
func GetCmdOutput(t *testing.T, jfrogCli *coreTests.JfrogCli, cmd ...string) ([]byte, []byte, error) {
	oldStdout := os.Stdout
	oldStdErr := os.Stderr
	temp, err := os.CreateTemp("", "output")
	assert.NoError(t, err)
	tempErr, err := os.CreateTemp("", "output")
	assert.NoError(t, err)
	os.Stdout = temp
	os.Stderr = tempErr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStdErr
		assert.NoError(t, temp.Close())
		assert.NoError(t, os.Remove(temp.Name()))
	}()
	err = jfrogCli.Exec(cmd...)
	assert.NoError(t, err)
	content, err := os.ReadFile(temp.Name())
	assert.NoError(t, err)
	errContent, err := os.ReadFile(tempErr.Name())
	return content, errContent, err
}

func VerifySha256DetailedSummaryFromBuffer(t *testing.T, buffer *bytes.Buffer, logger log.Log) {
	content := buffer.Bytes()
	buffer.Reset()
	logger.Output(string(content))

	var result summary.BuildInfoSummary
	err := json.Unmarshal(content, &result)
	assert.NoError(t, err)

	assert.Equal(t, summary.Success, result.Status)
	assert.True(t, result.Totals.Success > 0)
	assert.Equal(t, 0, result.Totals.Failure)
	// Verify a sha256 was returned
	assert.NotEmpty(t, result.Sha256Array, "Summary validation failed - no sha256 has returned from Artifactory.")
	for _, sha256 := range result.Sha256Array {
		// Verify sha256 is valid (a string size 256 characters) and not an empty string.
		assert.Equal(t, 64, len(sha256.Sha256Str), "Summary validation failed - invalid sha256 has returned from artifactory")
	}
}

func VerifySha256DetailedSummaryFromResult(t *testing.T, result *commandutils.Result) {
	if assert.NotNil(t, result) {
		reader := result.Reader()
		defer func() {
			assert.NoError(t, reader.Close())
		}()
		assert.NoError(t, reader.GetError())
		for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
			assert.Equal(t, 64, len(transferDetails.Sha256), "Summary validation failed - invalid sha256 has returned from artifactory")
		}
	}
}

func SkipKnownFailingTest(t *testing.T) {
	skipDate := time.Date(2023, time.November, 1, 0, 0, 0, 0, time.UTC)
	if time.Now().Before(skipDate) {
		t.Skip("Skipping a known failing test, will resume testing after ", skipDate.String())
	} else {
		t.Error("Not skipping test. Please fix the test or delay the skipMonth")
	}
}

func CreateContext(t *testing.T, testFlags, testArgs []string) (*cli.Context, *bytes.Buffer) {
	flagSet := createFlagSet(t, testFlags, testArgs)
	app := cli.NewApp()
	app.Writer = &bytes.Buffer{}
	return cli.NewContext(app, flagSet, nil), &bytes.Buffer{}
}

// Create flag set with input flags and arguments.
func createFlagSet(t *testing.T, flags []string, args []string) *flag.FlagSet {
	flagSet := flag.NewFlagSet("TestFlagSet", flag.ContinueOnError)
	flags = append(flags, "url=http://127.0.0.1:8081/artifactory")
	var cmdFlags []string
	for _, curFlag := range flags {
		flagSet.String(strings.Split(curFlag, "=")[0], "", "")
		cmdFlags = append(cmdFlags, "--"+curFlag)
	}
	cmdFlags = append(cmdFlags, args...)
	assert.NoError(t, flagSet.Parse(cmdFlags))
	return flagSet
}

func SkipTest(reason string) {
	log.Info(reason)
	os.Exit(0)
}
