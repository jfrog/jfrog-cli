package npm

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"io"
	"io/ioutil"
	"strings"
)

func Version(executablePath string) (string, error) {

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
		return "", errorutils.CheckError(err)
	}

	if npmError != nil {
		return "", errorutils.CheckError(npmError)
	}

	npmVersion := strings.TrimSpace(string(data))
	return npmVersion, nil
}

func createVersionCmdConfig(executablePath string, pipeWriter *io.PipeWriter) *NpmConfig {
	return &NpmConfig{
		Npm:       executablePath,
		Command:   []string{"-version"},
		StrWriter: pipeWriter,
		ErrWriter: nil,
	}
}
