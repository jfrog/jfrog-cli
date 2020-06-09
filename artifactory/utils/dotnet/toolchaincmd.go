package dotnet

import (
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"io"
	"os/exec"
)

type ToolchainType int

const (
	Nuget ToolchainType = iota
	DotnetCore
)

type toolchainInfo struct {
	name          string
	flagPrefix    string
	addSourceArgs []string
}

var toolchainsMap = map[ToolchainType]toolchainInfo{
	Nuget: {
		name:          "nuget",
		flagPrefix:    "-",
		addSourceArgs: []string{"sources", "add"},
	},
	DotnetCore: {
		name:          "dotnet",
		flagPrefix:    "--",
		addSourceArgs: []string{"nuget", "add", "source"},
	},
}

func (toolchainType ToolchainType) String() string {
	return toolchainsMap[toolchainType].name
}

func (toolchainType ToolchainType) GetTypeFlagPrefix() string {
	return toolchainsMap[toolchainType].flagPrefix
}

func (toolchainType ToolchainType) GetAddSourceArgs() []string {
	return toolchainsMap[toolchainType].addSourceArgs
}

func NewToolchainCmd(cmdType ToolchainType) (*Cmd, error) {
	// On Linux OS, NuGet can run only using mono.
	if cmdType == Nuget && cliutils.IsLinux() {
		return NewMonoCmd()
	}
	execPath, err := exec.LookPath(cmdType.String())
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return &Cmd{toolchain: cmdType, execPath: execPath}, nil
}

// Mono's first argument exec is nuget.exe's path, so we look for both mono and nuget.exe in PATH.
func NewMonoCmd() (*Cmd, error) {
	monoPath, err := exec.LookPath("mono")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	nugetExePath, err := exec.LookPath("nuget.exe")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return &Cmd{toolchain: Nuget, execPath: monoPath, Command: []string{nugetExePath}}, nil
}

func CreateDotnetAddSourceCmd(cmdType ToolchainType, sourceUrl string) (*Cmd, error) {
	addSourceCmd, err := NewToolchainCmd(cmdType)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	addSourceCmd.Command = append(addSourceCmd.Command, cmdType.GetAddSourceArgs()...)
	switch cmdType {
	case Nuget:
		addSourceCmd.CommandFlags = append(addSourceCmd.CommandFlags, "-source", sourceUrl)
	case DotnetCore:
		addSourceCmd.Command = append(addSourceCmd.Command, sourceUrl)
		// DotnetCore cli does not support password encryption on non-Windows OS, so we will write the raw password.
		if !cliutils.IsWindows() {
			addSourceCmd.CommandFlags = append(addSourceCmd.CommandFlags, "--store-password-in-clear-text")
		}
	}
	return addSourceCmd, nil
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

func (config *Cmd) GetToolchain() ToolchainType {
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
