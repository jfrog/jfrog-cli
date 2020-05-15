package dotnet

import (
	"errors"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	dotnetCmd "github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/commandargs"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/version"
	"strings"
)

const minDotnetSdkCore = "3.0.0"

type DotnetCoreCliCommand struct {
	*dotnetCmd.DotnetCommand
}

func NewDotnetCoreCliCommand() *DotnetCoreCliCommand {
	dotnetCoreCliCmd := DotnetCoreCliCommand{&dotnetCmd.DotnetCommand{}}
	dotnetCoreCliCmd.SetToolchainType(dotnet.DotnetCore)
	return &dotnetCoreCliCmd
}

func (dccc *DotnetCoreCliCommand) Run() error {
	return dccc.Exec()
}

func ValidateDotnetCoreSdkVersion() error {
	// Run dotnet --version
	localsCmd, err := dotnet.NewToolchainCmd(dotnet.DotnetCore)
	if err != nil {
		return err
	}
	localsCmd.CommandFlags = []string{"--version"}

	output, err := gofrogcmd.RunCmdOutput(localsCmd)
	if err != nil {
		return err
	}

	dotNetSdkCoreVersion := version.NewVersion(output)
	if !dotNetSdkCoreVersion.AtLeast(minDotnetSdkCore) {
		return errorutils.CheckError(errors.New("JFrog cli dotnet command requires .NET Core SDK version " + minDotnetSdkCore + " or higher, while version " + strings.TrimSpace(output) + " in use."))
	}
	return nil
}
