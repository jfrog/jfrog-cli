package npm

type NpmLegacyInstallCommand struct {
	commandName string
	*NpmCommandArgs
}

func NewNpmLegacyInstallCommand() *NpmLegacyInstallCommand {
	return &NpmLegacyInstallCommand{NpmCommandArgs: NewNpmCommandArgs("install"), commandName: "rt_npm_legacy_install"}
}

func NewNpmLegacyCiCommand() *NpmLegacyInstallCommand {
	return &NpmLegacyInstallCommand{NpmCommandArgs: NewNpmCommandArgs("ci"), commandName: "rt_npm_ci"}
}

func (nlic *NpmLegacyInstallCommand) Run() error {
	return nlic.run()
}

func (nlic *NpmLegacyInstallCommand) CommandName() string {
	return nlic.commandName
}
