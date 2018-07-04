package nuget

import (
	"os/exec"
	"io"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
)

func NewNugetCmd() (*Cmd, error) {
	execPath, err := exec.LookPath("nuget")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return &Cmd{Nuget: execPath}, nil
}

func (config *Cmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, config.Nuget)
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
	Nuget        string
	Command      []string
	CommandFlags []string
	StrWriter    io.WriteCloser
	ErrWriter    io.WriteCloser
}
