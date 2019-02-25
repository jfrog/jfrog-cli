package tests

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/ioutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var RtUrl *string
var RtUser *string
var RtPassword *string
var RtApiKey *string
var RtSshKeyPath *string
var RtSshPassphrase *string
var RtAccessToken *string
var BtUser *string
var BtKey *string
var BtOrg *string
var TestArtifactory *bool
var TestBintray *bool
var TestArtifactoryProxy *bool
var TestBuildTools *bool
var TestDocker *bool
var TestGo *bool
var DockerRepoDomain *string
var DockerTargetRepo *string
var TestNuget *bool
var HideUnitTestLog *bool

func init() {
	RtUrl = flag.String("rt.url", "http://127.0.0.1:8081/artifactory/", "Artifactory url")
	RtUser = flag.String("rt.user", "admin", "Artifactory username")
	RtPassword = flag.String("rt.password", "password", "Artifactory password")
	RtApiKey = flag.String("rt.apikey", "", "Artifactory user API key")
	RtSshKeyPath = flag.String("rt.sshKeyPath", "", "Ssh key file path")
	RtSshPassphrase = flag.String("rt.sshPassphrase", "", "Ssh key passphrase")
	RtAccessToken = flag.String("rt.accessToken", "", "Artifactory access token")
	TestArtifactory = flag.Bool("test.artifactory", true, "Test Artifactory")
	TestArtifactoryProxy = flag.Bool("test.artifactoryProxy", false, "Test Artifactory proxy")
	TestBintray = flag.Bool("test.bintray", false, "Test Bintray")
	BtUser = flag.String("bt.user", "", "Bintray username")
	BtKey = flag.String("bt.key", "", "Bintray API Key")
	BtOrg = flag.String("bt.org", "", "Bintray organization")
	TestBuildTools = flag.Bool("test.buildTools", false, "Test Maven, Gradle and npm builds")
	TestDocker = flag.Bool("test.docker", false, "Test Docker build")
	TestGo = flag.Bool("test.go", false, "Test Go")
	DockerRepoDomain = flag.String("rt.dockerRepoDomain", "", "Docker repository domain")
	DockerTargetRepo = flag.String("rt.dockerTargetRepo", "", "Docker repository domain")
	TestNuget = flag.Bool("test.nuget", false, "Test Nuget")
	HideUnitTestLog = flag.Bool("test.hideUnitTestLog", false, "Hide unit tests logs and print it in a file")
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

func IsExistLocally(expected, actual []string, t *testing.T) {
	if len(actual) == 0 && len(expected) != 0 {
		t.Error("Couldn't find all expected files, expected: " + strconv.Itoa(len(expected)) + ", found: " + strconv.Itoa(len(actual)))
	}
	err := compare(expected, actual)
	if err != nil {
		t.Error(err.Error())
	}
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

func CompareExpectedVsActuals(expected []string, actual []generic.SearchResult, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error(fmt.Sprintf("Unexpected behavior, expected: %s, \n%s\nfound: %s \n%s", strconv.Itoa(len(expected)), expected, strconv.Itoa(len(actual)), actual))
	}
	for _, v := range expected {
		for i, r := range actual {
			if v == r.Path {
				break
			}
			if i == len(actual)-1 {
				t.Error("Missing file: " + v)
			}
		}
	}

	for _, r := range actual {
		found := false
		for _, v := range expected {
			if v == r.Path {
				found = true
				break
			}
		}
		if !found {
			t.Error("Unexpected file: " + r.Path)
		}
	}
}

func GetTestResourcesPath() string {
	dir, _ := os.Getwd()
	fileSeparator := fileutils.GetFileSeparator()
	index := strings.LastIndex(dir, fileSeparator)
	dir = dir[:index]
	return dir + fileSeparator + "testsdata" + fileSeparator
}

func GetPath(filename, path string, a ...string) string {
	for i := 0; i < len(a); i++ {
		path += a[i] + fileutils.GetFileSeparator()
	}
	if filename != "" {
		path += filename
	}
	return ioutils.FixWinPath(path)
}

func getFileByOs(fileName string) string {
	var currentOs string
	fileSeparator := fileutils.GetFileSeparator()
	if cliutils.IsWindows() {
		currentOs = "win"
	} else {
		currentOs = "unix"
	}
	return GetTestResourcesPath() + "specs" + fileSeparator + currentOs + fileSeparator + fileName
}

func GetFilePath(fileName string) string {
	filePath := GetTestResourcesPath() + "specs/common" + fileutils.GetFileSeparator() + fileName
	isExists, _ := fileutils.IsFileExists(filePath, false)
	if isExists {
		return filePath
	}
	return getFileByOs(fileName)
}

func GetTestsLogsDir() (string, error) {
	tempDirPath := filepath.Join(cliutils.GetTempDir(), "jfrog_tests_logs")
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

func (cli *JfrogCli) WithSuffix(suffix string) *JfrogCli {
	return &JfrogCli{cli.main, cli.prefix, suffix}
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
	if _, err := os.Stat(testPath); err == nil {
		//path exists need to delete.
		err = os.RemoveAll(testPath)
		if err != nil {
			t.Error("Cannot remove path: " + testPath + " due to: " + err.Error())
		}
	}
}

// Renaming from old path to new path.
func RenamePath(oldPath, newPath string, t *testing.T) {
	err := fileutils.CopyDir(oldPath, newPath, true)
	if err != nil {
		t.Error("Error copying directory: ", oldPath, "to", newPath, err.Error())
		t.FailNow()
	}
	RemovePath(oldPath, t)
}

func DeleteFiles(deleteSpec *spec.SpecFiles, flags *generic.DeleteConfiguration) (successCount, failCount int, err error) {
	pathsToDelete, err := generic.GetPathsToDelete(deleteSpec, flags)
	if err != nil {
		return 0, 0, err
	}

	return generic.DeleteFiles(pathsToDelete, flags)
}

func GetNonVirtualRepositories() map[string]string {
	nonVirtualRepos := map[string]string{
		Repo1:   SpecsTestRepositoryConfig,
		Repo2:   MoveRepositoryConfig,
		LfsRepo: GitLfsTestRepositoryConfig,
	}

	if *TestBuildTools {
		nonVirtualRepos[JcenterRemoteRepo] = JcenterRemoteRepositoryConfig
		nonVirtualRepos[NpmLocalRepo] = NpmLocalRepositoryConfig
		nonVirtualRepos[NpmRemoteRepo] = NpmRemoteRepositoryConfig
	}

	return nonVirtualRepos
}

func GetVirtualRepositories() map[string]string {
	return map[string]string{
		VirtualRepo: VirtualRepositoryConfig,
	}
}

func getRepositoriesNameMap() map[string]string {
	return map[string]string{
		"${REPO1}":               Repo1,
		"${REPO2}":               Repo2,
		"${VIRTUAL_REPO}":        VirtualRepo,
		"${LFS_REPO}":            LfsRepo,
		"${JCENTER_REMOTE_REPO}": JcenterRemoteRepo,
		"${NPM_LOCAL_REPO}":      NpmLocalRepo,
		"${NPM_REMOTE_REPO}":     NpmRemoteRepo,
		"${GO_REPO}":             GoLocalRepo,
		"${RT_URL}":              *RtUrl,
		"${RT_API_KEY}":          *RtApiKey,
		"${RT_USERNAME}":         *RtUser,
		"${RT_PASSWORD}":         *RtPassword,
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
	log.Debug("Creating spec file at:", specPath)
	err = ioutil.WriteFile(specPath, []byte(content), 0700)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	return specPath, nil
}

func CreateSpec(fileName string) (string, error) {
	searchFilePath := GetFilePath(fileName)
	searchFilePath, err := ReplaceTemplateVariables(searchFilePath, "")
	return searchFilePath, err
}
