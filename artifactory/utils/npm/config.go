package npm

import (
	"io"
	"io/ioutil"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

// This method runs "npm config list --json" command and returns the json object that contains the current configurations of npm
// For more info see https://docs.npmjs.com/cli/config
func GetConfigList(npmFlags []string, executablePath string) ([]byte, error) {
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	configListCmdConfig := createConfigListCmdConfig(executablePath, npmFlags, pipeWriter)
	var npmError error
	go func() {
		npmError = gofrogcmd.RunCmd(configListCmdConfig)
	}()

	data, err := ioutil.ReadAll(pipeReader)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	if npmError != nil {
		return nil, errorutils.CheckError(npmError)
	}
	return data, nil
}

func createConfigListCmdConfig(executablePath string, splitFlags []string, pipeWriter *io.PipeWriter) *NpmConfig {
	return &NpmConfig{
		Npm:          executablePath,
		Command:      []string{"c", "ls"},
		CommandFlags: append(splitFlags, "-json=true"),
		StrWriter:    pipeWriter,
		ErrWriter:    nil,
	}
}
