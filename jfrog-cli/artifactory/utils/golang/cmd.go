package golang

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"io"
	"net/url"
	"os"
	"os/exec"
	"github.com/mattn/go-shellwords"
)

func NewCmd() (*Cmd, error) {
	execPath, err := exec.LookPath("vgo")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return &Cmd{Vgo: execPath}, nil
}

func (config *Cmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, config.Vgo)
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
	Vgo          string
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

	err = os.Setenv("GOPROXY", rtUrl.String())
	return err
}

func GetVgoVersion() ([]byte, error) {
	vgoCmd, err := NewCmd()
	if err != nil {
		return nil, err
	}
	vgoCmd.Command = []string{"version"}
	output, err := utils.RunCmdOutput(vgoCmd)
	return output, err
}

func RunVgo(vgoArg string) error {
	vgoCmd, err := NewCmd()
	if err != nil {
		return err
	}

	vgoCmd.Command, err = shellwords.Parse(vgoArg)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return utils.RunCmd(vgoCmd)
}
