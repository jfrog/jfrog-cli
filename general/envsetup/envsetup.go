package envsetup

import (
	"github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
)

const (
	registrationPageURL = "https://jfrog.com/start-free/cli/"
)

func RunEnvSetupCmd(outputFormat envsetup.OutputFormat) error {
	setupCmd := envsetup.NewEnvSetupCommand(registrationPageURL).SetOutputFormat(outputFormat)
	return progressbar.ExecWithProgress(setupCmd, false)
}
