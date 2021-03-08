package utils

import (
	"io"
	"os/exec"
)

// Command used to execute plugin commands.
type PluginExecCmd struct {
	ExecPath string
	Command  []string
}

func (pluginCmd *PluginExecCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, pluginCmd.ExecPath)
	cmd = append(cmd, pluginCmd.Command...)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (pluginCmd *PluginExecCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (pluginCmd *PluginExecCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (pluginCmd *PluginExecCmd) GetErrWriter() io.WriteCloser {
	return nil
}
