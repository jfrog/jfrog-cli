package utils

import (
	"bytes"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func SshAuthentication(details *config.ArtifactoryDetails) error {
	_, host, port, err := parseUrl(details.Url)
	if err != nil {
	    return err
	}

	log.Info("Performing SSH authentication...")

	var sshAuth ssh.AuthMethod
	if details.SshKeyPath != "" {
		sshAuth, err = PublicKeyFile(details.SshKeyPath)
	} else {
		sshAuth, err = SSHAgent()
	}
	if err != nil {
		return err
	}

	sshConfig := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			sshAuth,
		},
		HostKeyCallback : ssh.InsecureIgnoreHostKey(),
	}

	hostAndPort := host + ":" + strconv.Itoa(port)
	connection, err := ssh.Dial("tcp", hostAndPort, sshConfig)
	err = cliutils.CheckError(err)
	if err != nil {
	    return err
	}
	defer connection.Close()

	session, err := connection.NewSession()
	err = cliutils.CheckError(err)
	if err != nil {
	    return err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	err = cliutils.CheckError(err)
	if err != nil {
	    return err
	}

	var buf bytes.Buffer
	go io.Copy(&buf, stdout)

	session.Run("jfrog-authenticate")

	var result SshAuthResult
	err = json.Unmarshal(buf.Bytes(), &result)
	err = cliutils.CheckError(err)
	if err != nil {
	    return err
	}
	details.Url = cliutils.AddTrailingSlashIfNeeded(result.Href)
	details.SshAuthHeaders = result.Headers
	log.Info("SSH authentication successful.")
	return nil
}

func parseUrl(url string) (protocol, host string, port int, err error) {
	pattern1 := "^(.+)://(.+):([0-9].+)/$"
	pattern2 := "^(.+)://(.+)$"

    var r *regexp.Regexp
	r, err = regexp.Compile(pattern1)
	err = cliutils.CheckError(err)
	if err != nil {
	    return
	}
	groups := r.FindStringSubmatch(url)
	if len(groups) == 4 {
		protocol = groups[1]
		host = groups[2]
		port, err = strconv.Atoi(groups[3])
		if err != nil {
			err = cliutils.CheckError(errors.New("URL: " + url + " is invalid. Expecting ssh://<host>:<port> or http(s)://..."))
		}
		return
	}

	r, err = regexp.Compile(pattern2)
	err = cliutils.CheckError(err)
	if err != nil {
	    return
	}
	groups = r.FindStringSubmatch(url)
	if len(groups) == 3 {
		protocol = groups[1]
		host = groups[2]
		port = 80
	}
	return
}

type SshAuthResult struct {
	Href    string
	Headers map[string]string
}

func PublicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	err = cliutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	key, err := ssh.ParsePrivateKey(buffer)
	err = cliutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func SSHAgent() (ssh.AuthMethod, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers), nil
}
