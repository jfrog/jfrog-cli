package npm

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/mattn/go-shellwords"
	"io"
	"io/ioutil"
	"sync"
)

func RunList(flags, executablePath string) (stdResult, errResult []byte, err error) {
	log.Debug("Running npm list command.")
	splitFlags, err := shellwords.Parse(flags)
	if err != nil {
		return nil, nil, errorutils.CheckError(err)
	}

	listCmd := createListCommand(executablePath, splitFlags)
	outData, errData, err := listCmd.exec()
	log.Debug("npm list standard output is:\n" + string(outData))
	log.Debug("npm list error output is:\n" + string(errData))
	return outData, errData, err
}

func (listCmd *listCommand) exec() (outData, errData []byte, err error) {
	var wg sync.WaitGroup
	cmdErrors := make([]error, 3)
	wg.Add(3)
	go func() {
		defer wg.Done()
		cmdErrors[0] = utils.RunCmd(listCmd.cmdConfig)
	}()

	go func() {
		defer wg.Done()
		defer listCmd.outPipeReader.Close()
		data, err := ioutil.ReadAll(listCmd.outPipeReader)
		cmdErrors[1] = err
		outData = data
	}()

	go func() {
		defer wg.Done()
		defer listCmd.errPipeReader.Close()
		data, err := ioutil.ReadAll(listCmd.errPipeReader)
		cmdErrors[2] = err
		errData = data
	}()

	wg.Wait()
	for _, err := range cmdErrors {
		if err != nil {
			return outData, errData, errorutils.CheckError(err)
		}
	}
	return outData, errData, nil
}

func createListCommand(executablePath string, splitFlags []string) *listCommand {
	outPipeReader, outPipeWriter := io.Pipe()
	errPipeReader, errPipeWriter := io.Pipe()
	configListCmdConfig := createListCmdConfig(executablePath, splitFlags, outPipeWriter, errPipeWriter)
	return &listCommand{cmdConfig: configListCmdConfig,
		outPipeReader: outPipeReader,
		errPipeReader: errPipeReader,
	}
}

func createListCmdConfig(executablePath string, splitFlags []string, outPipeWriter *io.PipeWriter, errPipeWriter *io.PipeWriter) *NpmConfig {
	return &NpmConfig{
		Npm:          executablePath,
		Command:      []string{"list"},
		CommandFlags: append(splitFlags, "-json=true"),
		StrWriter:    outPipeWriter,
		ErrWriter:    errPipeWriter,
	}
}

type listCommand struct {
	cmdConfig     *NpmConfig
	outPipeReader *io.PipeReader
	errPipeReader *io.PipeReader
}
