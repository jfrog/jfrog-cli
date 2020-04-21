package dotnet

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	dotnetCmd "github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/commandArgs"
)

type DotnetCommand struct {
	*dotnetCmd.DotnetCommandArgs
}

func NewDotnetCommand() *DotnetCommand {
	dotnetCmd := DotnetCommand{&dotnetCmd.DotnetCommandArgs{}}
	dotnetCmd.SetToolchainType(dotnet.Dotnet)
	return &dotnetCmd
}

func (dc *DotnetCommand) Run() error {
	return dc.Exec()
}
