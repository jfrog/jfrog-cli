package utils

import (
	"bytes"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"golang.org/x/crypto/ssh"
	"errors"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func SshAuthentication(details *config.ArtifactoryDetails) error {
	_, host, port, err := parseUrl(details.Url)
	if err != nil {
	    return err
	}

	logger.Logger.Info("Performing SSH authentication...")
	if details.SshKeyPath == "" {
		err := cliutils.CheckError(errors.New("Cannot invoke the SshAuthentication function with no SSH key path. "))
        if err != nil {
            return err
        }
	}

	buffer, err := ioutil.ReadFile(details.SshKeyPath)
	err = cliutils.CheckError(err)
	if err != nil {
	    return err
	}
	key, err := ssh.ParsePrivateKey(buffer)
	err = cliutils.CheckError(err)
	if err != nil {
	    return err
	}
	sshConfig := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
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
	logger.Logger.Info("SSH authentication successful.")
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
