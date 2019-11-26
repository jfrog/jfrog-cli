package nuget

type NugetLegacyCommand struct {
	*NugetCommandArgs
}

func NewLegacyNugetCommand() *NugetLegacyCommand {
	return &NugetLegacyCommand{&NugetCommandArgs{}}
}

func (nlic *NugetLegacyCommand) Run() error {
	return nlic.run()
}
