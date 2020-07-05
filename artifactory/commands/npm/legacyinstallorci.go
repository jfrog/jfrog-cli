package npm

type NpmLegacyInstallOrCiCommand struct {
	internalCommandName string
	*NpmCommandArgs
}

func NewNpmLegacyInstallCommand() *NpmLegacyInstallOrCiCommand {
	return &NpmLegacyInstallOrCiCommand{NpmCommandArgs: NewNpmCommandArgs("install"), internalCommandName: "rt_npm_legacy_install"}
}

func NewNpmLegacyCiCommand() *NpmLegacyInstallOrCiCommand {
	return &NpmLegacyInstallOrCiCommand{NpmCommandArgs: NewNpmCommandArgs("ci"), internalCommandName: "rt_npm_legacy_ci"}
}

func (nlic *NpmLegacyInstallOrCiCommand) Run() error {
	return nlic.run()
}

func (nlic *NpmLegacyInstallOrCiCommand) CommandName() string {
	return nlic.internalCommandName
}
