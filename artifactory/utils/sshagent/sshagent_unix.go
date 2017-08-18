// +build !windows

package sshagent

import (
	"net"
	"os"
	"golang.org/x/crypto/ssh/agent"
)

func SshAgent() (agent.Agent, error) {
	conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	return agent.NewClient(conn), nil
}
