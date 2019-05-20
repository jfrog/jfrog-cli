package npm

import (
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/mattn/go-shellwords"
	"io"
	"io/ioutil"
)

// This method runs "npm config list --json" command and returns the json object that contains the current configurations of npm
// Fore more info see https://docs.npmjs.com/cli/config
func GetConfigList(npmFlags, executablePath string) ([]byte, error) {
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	splitFlags, err := shellwords.Parse(npmFlags)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	configListCmdConfig := createConfigListCmdConfig(executablePath, splitFlags, pipeWriter)
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
