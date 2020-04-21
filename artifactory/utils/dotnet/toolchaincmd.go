package dotnet

import (
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"io"
	"os/exec"
)

type ToolchainType int

const (
	Nuget ToolchainType = iota
	Dotnet
)

var CmdTypes = []string{
	"nuget",
	"dotnet",
}

func (toolchainType ToolchainType) String() string {
	return CmdTypes[toolchainType]
}

var CmdFlagPrefixes = []string{
	"-",
	"--",
}

func (toolchainType ToolchainType) GetTypeFlagPrefix() string {
	return CmdFlagPrefixes[toolchainType]
}

var AddSourceArgs = [][]string{
	{"sources", "add"},
	{"nuget", "add", "source"},
}

func (toolchainType ToolchainType) GetAddSourceArgs() []string {
	return AddSourceArgs[toolchainType]
}

func NewToolchainCmd(cmdType ToolchainType) (*Cmd, error) {
	execPath, err := exec.LookPath(cmdType.String())
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return &Cmd{toolchain: cmdType, execPath: execPath}, nil
}

func CreateDotnetAddSourceCmd(cmdType ToolchainType, sourceUrl string) (*Cmd, error) {
	execPath, err := exec.LookPath(cmdType.String())
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	addSourceCmd := Cmd{toolchain: cmdType, execPath: execPath, Command: cmdType.GetAddSourceArgs()}
	switch cmdType {
	case Nuget:
		addSourceCmd.CommandFlags = append(addSourceCmd.CommandFlags, "-source", sourceUrl)
	case Dotnet:
		addSourceCmd.Command = append(addSourceCmd.Command, sourceUrl)
	}
	return &addSourceCmd, nil
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

func (config *Cmd) Toolchain() ToolchainType {
	return config.toolchain
}

type Cmd struct {
	toolchain    ToolchainType
	execPath     string
	Command      []string
	CommandFlags []string
	StrWriter    io.WriteCloser
	ErrWriter    io.WriteCloser
}
