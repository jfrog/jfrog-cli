package tests

import (
	"bytes"
	"errors"
	"flag"
	"github.com/jfrog/jfrog-cli/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/utils/config"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

var RtUrl *string
var RtUser *string
var RtPassword *string
var RtApiKey *string
var RtSshKeyPath *string
var RtSshPassphrase *string
var RtAccessToken *string
var RtDistributionUrl *string
var RtDistributionAccessToken *string
var BtUser *string
var BtKey *string
var BtOrg *string
var TestArtifactory *bool
var TestBintray *bool
var TestArtifactoryProxy *bool
var TestDistribution *bool
var TestDocker *bool
var TestGo *bool
var TestNpm *bool
var TestGradle *bool
var TestMaven *bool
var DockerRepoDomain *string
var DockerTargetRepo *string
var TestNuget *bool
var HideUnitTestLog *bool
var TestPip *bool
var PipVirtualEnv *string

func init() {
	RtUrl = flag.String("rt.url", "http://127.0.0.1:8081/artifactory/", "Artifactory url")
	RtUser = flag.String("rt.user", "admin", "Artifactory username")
	RtPassword = flag.String("rt.password", "password", "Artifactory password")
	RtApiKey = flag.String("rt.apikey", "", "Artifactory user API key")
	RtSshKeyPath = flag.String("rt.sshKeyPath", "", "Ssh key file path")
	RtSshPassphrase = flag.String("rt.sshPassphrase", "", "Ssh key passphrase")
	RtAccessToken = flag.String("rt.accessToken", "", "Artifactory access token")
	RtDistributionUrl = flag.String("rt.distUrl", "", "Distribution url")
	RtDistributionAccessToken = flag.String("rt.distAccessToken", "", "Distribution access token")
	TestArtifactory = flag.Bool("test.artifactory", false, "Test Artifactory")
	TestArtifactoryProxy = flag.Bool("test.artifactoryProxy", false, "Test Artifactory proxy")
	TestBintray = flag.Bool("test.bintray", false, "Test Bintray")
	BtUser = flag.String("bt.user", "", "Bintray username")
	BtKey = flag.String("bt.key", "", "Bintray API Key")
	BtOrg = flag.String("bt.org", "", "Bintray organization")
	TestDistribution = flag.Bool("test.distribution", false, "Test distribution")
	TestDocker = flag.Bool("test.docker", false, "Test Docker build")
	TestGo = flag.Bool("test.go", false, "Test Go")
	TestNpm = flag.Bool("test.npm", false, "Test Npm")
	TestGradle = flag.Bool("test.gradle", false, "Test Gradle")
	TestMaven = flag.Bool("test.maven", false, "Test Maven")
	DockerRepoDomain = flag.String("rt.dockerRepoDomain", "", "Docker repository domain")
	DockerTargetRepo = flag.String("rt.dockerTargetRepo", "", "Docker repository domain")
	TestNuget = flag.Bool("test.nuget", false, "Test Nuget")
	HideUnitTestLog = flag.Bool("test.hideUnitTestLog", false, "Hide unit tests logs and print it in a file")
	TestPip = flag.Bool("test.pip", false, "Test Pip")
	PipVirtualEnv = flag.String("rt.pipVirtualEnv", "", "Pip virtual-environment path")
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
			os.RemoveAll(dir)
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
		return errors.New("Unexpected behavior, expected: " + strconv.Itoa(len(expected)) + " files, found: " + strconv.Itoa(len(actual)))
	}
	err := compare(expected, actual)
	return err
}

func ValidateChecksums(filePath string, expectedChecksum fileutils.ChecksumDetails, t *testing.T) {
	localFileDetails, err := fileutils.GetFileDetails(filePath)
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

func getPathsFromSearchResults(searchResults []generic.SearchResult) []string {
	var paths []string
	for _, result := range searchResults {
		paths = append(paths, result.Path)
	}
	return paths
}

func CompareExpectedVsActual(expected []string, actual []generic.SearchResult, t *testing.T) {
	assert.ElementsMatch(t, expected, getPathsFromSearchResults(actual))
}

func GetTestResourcesPath() string {
	dir, _ := os.Getwd()
	return filepath.ToSlash(dir + "/testsdata/")
}

func GetFilePathForBintray(filename, path string, a ...string) string {
	for i := 0; i < len(a); i++ {
		path += a[i] + "/"
	}
	if filename != "" {
		path += filename
	}
	return path
}

func GetFilePathForArtifactory(fileName string) string {
	return GetTestResourcesPath() + "specs/" + fileName
}

func GetTestsLogsDir() (string, error) {
	tempDirPath := filepath.Join(cliutils.GetCliPersistentTempDirPath(), "jfrog_tests_logs")
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

type JfrogCli struct {
	main        func() error
	prefix      string
	credentials string
}

func NewJfrogCli(mainFunc func() error, prefix, credentials string) *JfrogCli {
	return &JfrogCli{mainFunc, prefix, credentials}
}

func (cli *JfrogCli) Exec(args ...string) error {
	spaceSplit := " "
	os.Args = strings.Split(cli.prefix, spaceSplit)
	output := strings.Split(cli.prefix, spaceSplit)
	for _, v := range args {
		if v == "" {
			continue
		}
		args := strings.Split(v, spaceSplit)
		os.Args = append(os.Args, args...)
		output = append(output, args...)
	}
	if cli.credentials != "" {
		args := strings.Split(cli.credentials, spaceSplit)
		os.Args = append(os.Args, args...)
	}

	log.Info("[Command]", strings.Join(output, " "))
	return cli.main()
}

func (cli *JfrogCli) LegacyBuildToolExec(args ...string) error {
	spaceSplit := " "
	os.Args = strings.Split(cli.prefix, spaceSplit)
	output := strings.Split(cli.prefix, spaceSplit)
	for _, v := range args {
		if v == "" {
			continue
		}
		os.Args = append(os.Args, v)
		output = append(output, v)
	}
	if cli.credentials != "" {
		args := strings.Split(cli.credentials, spaceSplit)
		os.Args = append(os.Args, args...)
	}

	log.Info("[Command]", strings.Join(output, " "))
	return cli.main()
}

func (cli *JfrogCli) WithoutCredentials() *JfrogCli {
	return &JfrogCli{cli.main, cli.prefix, ""}
}

type gitManager struct {
	dotGitPath string
}

func GitExecutor(dotGitPath string) *gitManager {
	return &gitManager{dotGitPath: dotGitPath}
}

func (m *gitManager) GetUrl() (string, string, error) {
	return m.execGit("config", "--get", "remote.origin.url")
}

func (m *gitManager) GetRevision() (string, string, error) {
	return m.execGit("show", "-s", "--format=%H", "HEAD")
}

func (m *gitManager) execGit(args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Dir = m.dotGitPath
	cmd.Stdin = nil
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	errorutils.CheckError(err)
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

// Prepare the .git environment for the test. Takes an existing folder and making it .git dir.
// sourceDirPath - Relative path to the source dir to change to .git
// targetDirPath - Relative path to the target created .git dir, usually 'testdata' under the parent dir.
func PrepareDotGitDir(t *testing.T, sourceDirPath, targetDirPath string) (string, string) {
	// Get path to create .git folder in
	baseDir, _ := os.Getwd()
	baseDir = filepath.Join(baseDir, targetDirPath)
	// Create .git path and make sure it is clean
	dotGitPath := filepath.Join(baseDir, ".git")
	RemovePath(dotGitPath, t)
	// Get the path of the .git candidate path
	dotGitPathTest := filepath.Join(baseDir, sourceDirPath)
	// Rename the .git candidate
	RenamePath(dotGitPathTest, dotGitPath, t)
	return baseDir, dotGitPath
}

// Removing the provided path from the filesystem
func RemovePath(testPath string, t *testing.T) {
	err := fileutils.RemovePath(testPath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

// Renaming from old path to new path.
func RenamePath(oldPath, newPath string, t *testing.T) {
	err := fileutils.RenamePath(oldPath, newPath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func DeleteFiles(deleteSpec *spec.SpecFiles, artifactoryDetails *config.ArtifactoryDetails) (successCount, failCount int, err error) {
	deleteCommand := generic.NewDeleteCommand()
	deleteCommand.SetThreads(3).SetSpec(deleteSpec).SetRtDetails(artifactoryDetails).SetDryRun(false)
	err = deleteCommand.GetPathsToDelete()
	if err != nil {
		return 0, 0, err
	}

	return deleteCommand.DeleteFiles()
}

var reposConfigMap = map[*string]string{
	&Repo1:             Repo1RepositoryConfig,
	&Repo2:             Repo2RepositoryConfig,
	&VirtualRepo:       VirtualRepositoryConfig,
	&JcenterRemoteRepo: JcenterRemoteRepositoryConfig,
	&LfsRepo:           GitLfsTestRepositoryConfig,
	&DebianRepo:        DebianTestRepositoryConfig,
	&NpmLocalRepo:      NpmLocalRepositoryConfig,
	&NpmRemoteRepo:     NpmRemoteRepositoryConfig,
	&GoLocalRepo:       GoLocalRepositoryConfig,
	&PypiRemoteRepo:    PypiRemoteRepositoryConfig,
	&PypiVirtualRepo:   PypiVirtualRepositoryConfig,
}

var CreatedNonVirtualRepositories map[*string]string
var CreatedVirtualRepositories map[*string]string

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

func GetNonVirtualRepositories() map[*string]string {
	NonVirtualReposMap := map[*bool][]*string{
		TestArtifactory:  {&Repo1, &Repo2, &LfsRepo, &DebianRepo},
		TestDistribution: {&Repo1, &Repo2},
		TestDocker:       {},
		TestGo:           {},
		TestGradle:       {&Repo1},
		TestMaven:        {&Repo1, &Repo2, &JcenterRemoteRepo},
		TestNpm:          {&NpmLocalRepo, &NpmRemoteRepo},
		TestNuget:        {},
		TestPip:          {&PypiRemoteRepo},
	}
	return getNeededRepositories(NonVirtualReposMap)
}

func GetVirtualRepositories() map[*string]string {
	VirtualReposMap := map[*bool][]*string{
		TestArtifactory:  {&VirtualRepo},
		TestDistribution: {},
		TestDocker:       {},
		TestGo:           {},
		TestGradle:       {},
		TestMaven:        {},
		TestNpm:          {},
		TestNuget:        {},
		TestPip:          {&PypiVirtualRepo},
	}
	return getNeededRepositories(VirtualReposMap)
}

func getRepositoriesNameMap() map[string]string {
	return map[string]string{
		"${REPO1}":               Repo1,
		"${REPO2}":               Repo2,
		"${REPO_1_AND_2}":        Repo1And2,
		"${VIRTUAL_REPO}":        VirtualRepo,
		"${LFS_REPO}":            LfsRepo,
		"${DEBIAN_REPO}":         DebianRepo,
		"${JCENTER_REMOTE_REPO}": JcenterRemoteRepo,
		"${NPM_LOCAL_REPO}":      NpmLocalRepo,
		"${NPM_REMOTE_REPO}":     NpmRemoteRepo,
		"${GO_REPO}":             GoLocalRepo,
		"${RT_SERVER_ID}":        RtServerId,
		"${RT_URL}":              *RtUrl,
		"${RT_API_KEY}":          *RtApiKey,
		"${RT_USERNAME}":         *RtUser,
		"${RT_PASSWORD}":         *RtPassword,
		"${RT_ACCESS_TOKEN}":     *RtAccessToken,
		"${PYPI_REMOTE_REPO}":    PypiRemoteRepo,
		"${PYPI_VIRTUAL_REPO}":   PypiVirtualRepo,
	}
}

func ReplaceTemplateVariables(path, destPath string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	repos := getRepositoriesNameMap()
	for repoName, repoValue := range repos {
		content = bytes.Replace(content, []byte(repoName), []byte(repoValue), -1)
	}
	if destPath == "" {
		destPath, err = os.Getwd()
		if err != nil {
			return "", errorutils.CheckError(err)
		}
		destPath = filepath.Join(destPath, Temp)
	}
	err = os.MkdirAll(destPath, 0700)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	specPath := filepath.Join(destPath, filepath.Base(path))
	log.Info("Creating spec file at:", specPath)
	err = ioutil.WriteFile(specPath, []byte(content), 0700)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	return specPath, nil
}

func CreateSpec(fileName string) (string, error) {
	searchFilePath := GetFilePathForArtifactory(fileName)
	searchFilePath, err := ReplaceTemplateVariables(searchFilePath, "")
	return searchFilePath, err
}

func ConvertSliceToMap(props []utils.Property) map[string]string {
	propsMap := make(map[string]string)
	for _, item := range props {
		propsMap[item.Key] = item.Value
	}
	return propsMap
}
