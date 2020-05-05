package dotnet

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	dotnetCmd "github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/commandargs"
)

type NugetLegacyCommand struct {
	*dotnetCmd.DotnetCommand
}

func NewLegacyNugetCommand() *NugetLegacyCommand {
	nugetLegacyCmd := NugetLegacyCommand{&dotnetCmd.DotnetCommand{}}
	nugetLegacyCmd.SetToolchainType(dotnet.Nuget)
	return &nugetLegacyCmd
}

func (nlc *NugetLegacyCommand) Run() error {
	return nlc.Exec()
}
