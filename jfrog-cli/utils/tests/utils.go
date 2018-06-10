package tests

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/ioutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
var BtUser *string
var BtKey *string
var BtOrg *string
var TestArtifactory *bool
var TestBintray *bool
var TestArtifactoryProxy *bool
var TestBuildTools *bool
var TestDocker *bool
var TestVgo *bool
var DockerRepoDomain *string
var DockerTargetRepo *string

func init() {
	RtUrl = flag.String("rt.url", "http://127.0.0.1:8081/artifactory/", "Artifactory url")
	RtUser = flag.String("rt.user", "admin", "Artifactory username")
	RtPassword = flag.String("rt.password", "password", "Artifactory password")
	RtApiKey = flag.String("rt.apikey", "", "Artifactory user API key")
	RtSshKeyPath = flag.String("rt.sshKeyPath", "", "Ssh key file path")
	RtSshPassphrase = flag.String("rt.sshPassphrase", "", "Ssh key passphrase")
	TestArtifactory = flag.Bool("test.artifactory", true, "Test Artifactory")
	TestArtifactoryProxy = flag.Bool("test.artifactoryProxy", false, "Test Artifactory proxy")
	TestBintray = flag.Bool("test.bintray", false, "Test Bintray")
	BtUser = flag.String("bt.user", "", "Bintray username")
	BtKey = flag.String("bt.key", "", "Bintray API Key")
	BtOrg = flag.String("bt.org", "", "Bintray organization")
	TestBuildTools = flag.Bool("test.buildTools", false, "Test Maven, Gradle and npm builds")
	TestDocker = flag.Bool("test.docker", false, "Test Docker build")
	TestVgo = flag.Bool("test.vgo", false, "Test Vgo")
	DockerRepoDomain = flag.String("rt.dockerRepoDomain", "", "Docker repository domain")
	DockerTargetRepo = flag.String("rt.dockerTargetRepo", "", "Docker repository domain")
}

func CleanFileSystem() {
	isExist, err := fileutils.IsDirExists(Out)
	if err != nil {
		log.Error(err)
	}
	if isExist {
		os.RemoveAll(Out)
	}
}

func IsExistLocally(expected, actual []string, t *testing.T) {
	if len(actual) == 0 && len(expected) != 0 {
		t.Error("Couldn't find all expected files, expected: " + strconv.Itoa(len(expected)) + ", found: " + strconv.Itoa(len(actual)))
	}
	compare(expected, actual, t)
}

func ValidateListsIdentical(expected, actual []string, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error("Unexpected behavior, expected: " + strconv.Itoa(len(expected)) + " files, found: " + strconv.Itoa(len(actual)))
	}
	compare(expected, actual, t)
}

func ValidateChecksums(filePath string, expectedChecksum fileutils.ChecksumDetails, t *testing.T) {
	localFileDetails, err := fileutils.GetFileDetails(filePath)
	if err != nil {
		t.Error("Couldn't calculate sha1, " + err.Error())
	}
	if localFileDetails.Checksum.Sha1 != expectedChecksum.Sha1 {
		t.Error("sha1 mismatch for " + filePath + ", expected: " + expectedChecksum.Sha1, "found: " + localFileDetails.Checksum.Sha1)
	}
	if localFileDetails.Checksum.Md5 != expectedChecksum.Md5 {
		t.Error("md5 mismatch for " + filePath + ", expected: " + expectedChecksum.Md5, "found: " + localFileDetails.Checksum.Sha1)
	}
	if localFileDetails.Checksum.Sha256 != expectedChecksum.Sha256 {
		t.Error("sha256 mismatch for " + filePath + ", expected: " + expectedChecksum.Sha256, "found: " + localFileDetails.Checksum.Sha1)
	}
}

func compare(expected, actual []string, t *testing.T) {
	for _, v := range expected {
		for i, r := range actual {
			if v == r {
				break
			}
			if i == len(actual)-1 {
				t.Error("Missing file : " + v)
			}
		}
	}
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
	if runtime.GOOS == "windows" {
		currentOs = "win"
	} else {
		currentOs = "unix"
	}
	return GetTestResourcesPath() + "specs" + fileSeparator + currentOs + fileSeparator + fileName
}

func GetFilePath(fileName string) string {
	filePath := GetTestResourcesPath() + "specs/common" + fileutils.GetFileSeparator() + fileName
	isExists, _ := fileutils.IsFileExists(filePath)
	if isExists {
		return filePath
	}
	return getFileByOs(fileName)
}

func GetTestsLogsDir() (string, error) {
	tempDirPath := filepath.Join(os.TempDir(), "jfrog_tests_logs")
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
	main   func()
	prefix string
	suffix string
}

func NewJfrogCli(mainFunc func(), prefix, suffix string) *JfrogCli {
	return &JfrogCli{mainFunc, prefix, suffix}
}

func (cli *JfrogCli) Exec(args ...string) {
	spaceSplit := " "
	os.Args = strings.Split(cli.prefix, spaceSplit)
	for _, v := range args {
		if v == "" {
			continue
		}
		os.Args = append(os.Args, strings.Split(v, spaceSplit)...)
	}
	if cli.suffix != "" {
		os.Args = append(os.Args, strings.Split(cli.suffix, spaceSplit)...)
	}

	log.Info("[Command]", strings.Join(os.Args, " "))
	cli.main()
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
