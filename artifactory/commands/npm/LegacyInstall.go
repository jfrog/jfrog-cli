package npm

import "github.com/jfrog/jfrog-client-go/utils/log"

type NpmLegacyInstallCommand struct {
	*NpmCommandArgs
}

func NewNpmLegacyInstallCommand() *NpmLegacyInstallCommand {
	return &NpmLegacyInstallCommand{NpmCommandArgs: NewNpmCommandArgs("install")}
}

func (nlic *NpmLegacyInstallCommand) Run() error {
	log.Info("Running npm legacy Install.")
	return nlic.run()
}

func (nlic *NpmLegacyInstallCommand) CommandName() string {
	return "rt_npm_legacy_install"
}
