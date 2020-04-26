package nuget

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	dotnetCmd "github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/commandargs"
)

type NugetLegacyCommand struct {
	*dotnetCmd.DotnetCommandArgs
}

func NewLegacyNugetCommand() *NugetLegacyCommand {
	nugetLegacyCmd := NugetLegacyCommand{&dotnetCmd.DotnetCommandArgs{}}
	nugetLegacyCmd.SetToolchainType(dotnet.Nuget)
	return &nugetLegacyCmd
}

func (nlic *NugetLegacyCommand) Run() error {
	return nlic.Exec()
}
