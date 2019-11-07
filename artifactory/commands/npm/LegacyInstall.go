package npm

import "github.com/jfrog/jfrog-client-go/utils/log"

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
	log.Info("Running npm legacy " + nlic.CommandName() + ".")
	return nlic.run()
}

func (nlic *NpmLegacyInstallCommand) CommandName() string {
	return nlic.commandName
}
