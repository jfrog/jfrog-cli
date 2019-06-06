package npm

import "github.com/jfrog/jfrog-client-go/utils/log"

type NpmCiCommand struct {
	*NpmCommandArgs
}

func NewNpmCiCommand() *NpmCiCommand {
	return &NpmCiCommand{NpmCommandArgs: NewNpmCommandArgs("ci")}
}

func (ncc *NpmCiCommand) Run() error {
	log.Info("Running npm ci.")
	return ncc.run()
}

func (ncc *NpmCiCommand) CommandName() string {
	return "rt_npm_ci"
}
