package dotnet

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	dotnetCmd "github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/commandargs"
)

type DotnetCoreCliCommand struct {
	*dotnetCmd.DotnetCommand
}

func NewDotnetCoreCliCommand() *DotnetCoreCliCommand {
	dotnetCoreCliCmd := DotnetCoreCliCommand{&dotnetCmd.DotnetCommand{}}
	dotnetCoreCliCmd.SetToolchainType(dotnet.Dotnet)
	return &dotnetCoreCliCmd
}

func (dccc *DotnetCoreCliCommand) Run() error {
	return dccc.Exec()
}
