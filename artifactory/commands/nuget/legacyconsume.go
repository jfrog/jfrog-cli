package nuget

import "github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"

type NugetLegacyCommand struct {
	*dotnetCommandArgs
}

func NewLegacyNugetCommand() *NugetLegacyCommand {
	return &NugetLegacyCommand{&dotnetCommandArgs{cmdType: dotnet.Nuget}}
}

func (nlic *NugetLegacyCommand) Run() error {
	return nlic.run()
}
