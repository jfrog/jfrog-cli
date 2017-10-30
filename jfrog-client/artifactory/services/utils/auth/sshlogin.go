package auth

import (
	"bytes"
	"encoding/json"
	"github.com/xanzy/ssh-agent"
	"errors"
	"io"
	"regexp"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"golang.org/x/crypto/ssh"
)

func sshAuthentication(url string, sshKey, sshPassphrase []byte) (map[string]string, string, error) {
	_, host, port, err := parseUrl(url)
	if err != nil {
	    return nil, "", err
	}

	log.Info("Performing SSH authentication...")
	var sshAuth ssh.AuthMethod
	if len(sshKey) == 0 {
		sshAuth, err = sshAuthAgent()
	} else {
		sshAuth, err = sshAuthPublicKey(sshKey, sshPassphrase)
	}
	if err != nil {
		return nil, "", err
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
	if errorutils.CheckError(err) != nil {
		return nil, "", err
	}
	defer connection.Close()
	session, err := connection.NewSession()
	if errorutils.CheckError(err) != nil {
		return nil, "", err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if errorutils.CheckError(err) != nil {
		return nil, "", err
	}

	if err = session.Run("jfrog-authenticate"); err != nil && err != io.EOF {
		return nil, "", errorutils.CheckError(err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, stdout)

	var result SshAuthResult
	if err = json.Unmarshal(buf.Bytes(), &result); errorutils.CheckError(err) != nil {
		return nil, "", err
	}
	url = cliutils.AddTrailingSlashIfNeeded(result.Href)
	sshAuthHeaders := result.Headers
	log.Info("SSH authentication successful.")
	return sshAuthHeaders, url, nil
}

func parseUrl(url string) (protocol, host string, port int, err error) {
	pattern1 := "^(.+)://(.+):([0-9].+)/$"
	pattern2 := "^(.+)://(.+)$"

    var r *regexp.Regexp
	r, err = regexp.Compile(pattern1)
	if errorutils.CheckError(err) != nil {
	    return
	}
	groups := r.FindStringSubmatch(url)
	if len(groups) == 4 {
		protocol = groups[1]
		host = groups[2]
		port, err = strconv.Atoi(groups[3])
		if err != nil {
			err = errorutils.CheckError(errors.New("URL: " + url + " is invalid. Expecting ssh://<host>:<port> or http(s)://..."))
		}
		return
	}

	r, err = regexp.Compile(pattern2)
	err = errorutils.CheckError(err)
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

func sshAuthPublicKey(sshKey, sshPassphrase []byte) (ssh.AuthMethod, error) {
	key, err := ssh.ParsePrivateKeyWithPassphrase(sshKey, sshPassphrase)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func sshAuthAgent() (ssh.AuthMethod, error) {
	log.Info("Authenticating Using SSH agent")
	sshAgent, _, err := sshagent.New()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	cbk := sshAgent.Signers
	authMethod := ssh.PublicKeysCallback(cbk)
	return authMethod, nil
}

type SshAuthResult struct {
	Href    string
	Headers map[string]string
}
