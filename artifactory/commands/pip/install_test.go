package pip

import (
	"bytes"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
	logUtils "github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"testing"
)

func TestRunStam(t *testing.T) {
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	pic := NewPipInstallCommand()
	pic.Run()
}

func TestRunCmd(t *testing.T) {
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)


	pipInstallCmd := &pip.PipCmd{
		Executable:  "sh",
		Command:     "-c",
		CommandArgs: []string{"env"},
		EnvVars:     nil,
		StrWriter:   nil,
		ErrWriter:   nil,
	}
	gofrogcmd.RunCmd(pipInstallCmd)
}
