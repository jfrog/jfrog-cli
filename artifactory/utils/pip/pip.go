package pip

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// Get executable path.
// If run inside a virtual-env, this should return the path for the correct executable.
func GetExecutablePath(executableName string) (string, error) {
	executablePath, err := exec.LookPath(executableName)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	if executablePath == "" {
		return "", errorutils.CheckError(errors.New(fmt.Sprintf("Could not find '%s' executable", executableName)))
	}

	log.Debug(fmt.Sprintf("Found %s executable at: %s", executableName, executablePath))
	return executablePath, nil
}

func (pc *PipCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, pc.Executable)
	cmd = append(cmd, pc.Command)
	cmd = append(cmd, pc.CommandArgs...)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (pc *PipCmd) GetEnv() map[string]string {
	return pc.EnvVars
}

func (pc *PipCmd) GetStdWriter() io.WriteCloser {
	return pc.StrWriter
}

func (pc *PipCmd) GetErrWriter() io.WriteCloser {
	return pc.ErrWriter
}

func GetPipDepTreeScriptPath() (string, error) {
	pipDependenciesPath, err := config.GetJfrogDependenciesPath()
	if err != nil {
		return "", err
	}
	pipDependenciesPath = filepath.Join(pipDependenciesPath, "pip", pipDepTreeVersion)

	return writePipDepTreeScriptIfNeeded(pipDependenciesPath)
}

func writePipDepTreeScriptIfNeeded(targetPath string) (string, error) {
	scriptPath := path.Join(targetPath, "pipdeptree.py")
	exists, err := fileutils.IsFileExists(scriptPath, false)
	if err != nil {
		return "", err
	}
	if !exists {
		err = os.MkdirAll(targetPath, os.ModeDir|os.ModePerm)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(scriptPath, pipDepTreeContent, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return scriptPath, nil
}

type PipCmd struct {
	Executable  string
	Command     string
	CommandArgs []string
	EnvVars     map[string]string
	StrWriter   io.WriteCloser
	ErrWriter   io.WriteCloser
}
