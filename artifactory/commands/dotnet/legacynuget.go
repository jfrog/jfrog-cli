package dotnet

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
)

type NugetLegacyCommand struct {
	*DotnetCommand
}

func NewLegacyNugetCommand() *NugetLegacyCommand {
	nugetLegacyCmd := NugetLegacyCommand{&DotnetCommand{}}
	nugetLegacyCmd.SetToolchainType(dotnet.Nuget)
	return &nugetLegacyCmd
}

func (nlc *NugetLegacyCommand) Run() error {
	nlc.useNugetAddSource = true
	return nlc.Exec()
}
