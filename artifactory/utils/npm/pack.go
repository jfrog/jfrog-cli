package npm

import (
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

func Pack(npmFlags []string, executablePath string) error {

	configListCmdConfig := createPackCmdConfig(executablePath, npmFlags)
	if err := gofrogcmd.RunCmd(configListCmdConfig); err != nil {
		return errorutils.CheckError(err)
	}

	return nil
}

func createPackCmdConfig(executablePath string, splitFlags []string) *NpmConfig {
	return &NpmConfig{
		Npm:          executablePath,
		Command:      []string{"pack"},
		CommandFlags: append(splitFlags),
		StrWriter:    nil,
		ErrWriter:    nil,
	}
}
