package pip

import (
	"io"
	"os/exec"
)

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

type PipCmd struct {
	Executable  string
	Command     string
	CommandArgs []string
	EnvVars     map[string]string
	StrWriter   io.WriteCloser
	ErrWriter   io.WriteCloser
}

