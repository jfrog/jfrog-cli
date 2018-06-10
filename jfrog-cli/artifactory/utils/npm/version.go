package npm

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"io"
	"io/ioutil"
)

func Version(executablePath string) ([]byte, error) {

	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()
	var npmError error

	configListCmdConfig := createVersionCmdConfig(executablePath, pipeWriter)
	go func() {
		npmError = utils.RunCmd(configListCmdConfig)
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

func createVersionCmdConfig(executablePath string, pipeWriter *io.PipeWriter) *NpmConfig {
	return &NpmConfig{
		Npm:       executablePath,
		Command:   []string{"-version"},
		StrWriter: pipeWriter,
		ErrWriter: nil,
	}
}
