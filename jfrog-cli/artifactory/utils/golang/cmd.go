package golang

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/mattn/go-shellwords"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const GOPROXY = "GOPROXY"

func NewCmd() (*Cmd, error) {
	execPath, err := exec.LookPath("go")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return &Cmd{Go: execPath}, nil
}

func (config *Cmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, config.Go)
	cmd = append(cmd, config.Command...)
	cmd = append(cmd, config.CommandFlags...)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (config *Cmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (config *Cmd) GetStdWriter() io.WriteCloser {
	return config.StrWriter
}

func (config *Cmd) GetErrWriter() io.WriteCloser {
	return config.ErrWriter
}

type Cmd struct {
	Go           string
	Command      []string
	CommandFlags []string
	StrWriter    io.WriteCloser
	ErrWriter    io.WriteCloser
}

func SetGoProxyEnvVar(artifactoryDetails *config.ArtifactoryDetails, repoName string) error {
	rtUrl, err := url.Parse(artifactoryDetails.Url)
	if err != nil {
		return err
	}
	rtUrl.User = url.UserPassword(artifactoryDetails.User, artifactoryDetails.Password)
	rtUrl.Path += "api/go/" + repoName

	err = os.Setenv(GOPROXY, rtUrl.String())
	return err
}

func GetGoVersion() ([]byte, error) {
	goCmd, err := NewCmd()
	if err != nil {
		return nil, err
	}
	goCmd.Command = []string{"version"}
	output, err := utils.RunCmdOutput(goCmd)
	return output, err
}

func RunGo(goArg string) error {
	goCmd, err := NewCmd()
	if err != nil {
		return err
	}

	goCmd.Command, err = shellwords.Parse(goArg)
	if err != nil {
		return errorutils.CheckError(err)
	}

	regExp, err := utils.GetRegExp(`((http|https):\/\/\w.*?:\w.*?@)`)
	if err != nil {
		return err
	}

	protocolRegExp := utils.CmdOutputPattern{
		RegExp: regExp,
	}
	protocolRegExp.ExecFunc = protocolRegExp.MaskCredentials

	regExp, err = utils.GetRegExp("(404 Not Found)")
	if err != nil {
		return err
	}

	notFoundRegExp := utils.CmdOutputPattern{
		RegExp: regExp,
	}
	notFoundRegExp.ExecFunc = notFoundRegExp.ErrorOnNotFound

	return utils.RunCmdWithOutputParser(goCmd, &protocolRegExp, &notFoundRegExp)
}

// Using go mod download {dependency} command to download the dependency
func DownloadDependency(dependencyName string) error {
	goCmd, err := NewCmd()
	if err != nil {
		return err
	}

	goCmd.Command = []string{"mod", "download", "-json", dependencyName}
	return utils.RunCmd(goCmd)
}

// Runs go mod graph command and returns slice of the dependencies
func GetDependenciesGraph() (map[string]bool, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Read and store the content of the go.mod file,
	// because it may change by the "go mod graph" command.
	var modFilePath string
	var modFileContent []byte
	var modFileStat os.FileInfo
	modFilePath, modFileContent, modFileStat, err = getModFileDetails()
	if err != nil {
		return nil, err
	}

	log.Info("Running 'go mod graph' in", pwd)
	goCmd, err := NewCmd()
	if err != nil {
		return nil, err
	}
	goCmd.Command = []string{"mod", "graph"}

	// Restore the content of the go.mod file, to make sure it stays the same as before
	// running the "go mod graph" command.
	err = ioutil.WriteFile(modFilePath, modFileContent, modFileStat.Mode())
	if err != nil {
		return nil, err
	}

	output, err := utils.RunCmdOutput(goCmd)
	if len(output) != 0 {
		log.Debug(string(output))
	}
	return outputToMap(string(output)), errorutils.CheckError(err)
}

func getModFileDetails() (modFilePath string, modFileContent []byte, modFileStat os.FileInfo, err error) {
	rootDir, err := GetRootDir()
	if err != nil {
		return
	}
	modFilePath = filepath.Join(rootDir, "go.mod")
	modFileStat, err = os.Stat(modFilePath)
	if err != nil {
		return
	}
	modFileContent, err = ioutil.ReadFile(modFilePath)
	return
}

func outputToMap(output string) map[string]bool {
	lineOutput := strings.Split(output, "\n")
	var result []string
	mapOfDeps := map[string]bool{}
	for _, line := range lineOutput {
		splitLine := strings.Split(line, " ")
		if len(splitLine) == 2 {
			mapOfDeps[splitLine[1]] = true
			result = append(result, splitLine[1])
		}
	}
	return mapOfDeps
}

// Using go mod download command to download all the dependencies before publishing to Artifactory
func RunGoModTidy() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	log.Info("Running 'go mod tidy' in", pwd)
	goCmd, err := NewCmd()
	if err != nil {
		return err
	}

	goCmd.Command = []string{"mod", "tidy"}
	_, err = utils.RunCmdOutput(goCmd)
	if err != nil {
		return err
	}

	err = signModFile()
	return err
}

func signModFile() error {
	rootDir, err := GetRootDir()
	if err != nil {
		return err
	}
	modFilePath := filepath.Join(rootDir, "go.mod")
	stat, err := os.Stat(modFilePath)
	if err != nil {
		return errorutils.CheckError(err)
	}
	modFileContent, err := ioutil.ReadFile(modFilePath)
	newContent := append([]byte("// Edited by JFrog CLI\n\n"), modFileContent...)
	err = ioutil.WriteFile(modFilePath, newContent, stat.Mode())
	return errorutils.CheckError(err)
}

// Returns the root dir where the go.mod located.
func GetRootDir() (string, error) {
	goCmd, err := NewCmd()
	if err != nil {
		return "", err
	}

	goCmd.Command = []string{"list"}
	goCmd.CommandFlags = []string{"-m", "-f={{.Dir}}"}
	output, err := utils.RunCmdOutput(goCmd)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return strings.TrimSpace(string(output)), nil
}
