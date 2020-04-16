package dotnet

import (
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"io"
	"os/exec"
)

type CmdType int

const (
	Nuget CmdType = iota
	Dotnet
)

var CmdTypes = []string{
	"nuget",
	"dotnet",
}

func (cmdType CmdType) String() string {
	return CmdTypes[cmdType]
}

func NewDotnetCmd(cmdType CmdType) (*Cmd, error) {
	execPath, err := exec.LookPath(cmdType.String())
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return &Cmd{execPath: execPath}, nil
}

func (config *Cmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, config.execPath)
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
	execPath     string
	Command      []string
	CommandFlags []string
	StrWriter    io.WriteCloser
	ErrWriter    io.WriteCloser
}
