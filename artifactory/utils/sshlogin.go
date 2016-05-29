package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
)

func SshAuthentication(details *config.ArtifactoryDetails) {
	_, host, port := parseUrl(details.Url)

	fmt.Println("Performing SSH authentication...")
	if details.SshKeyPath == "" {
		cliutils.Exit(cliutils.ExitCodeError, "Cannot invoke the SshAuthentication function with no SSH key path. ")
	}

	buffer, err := ioutil.ReadFile(details.SshKeyPath)
	cliutils.CheckError(err)
	key, err := ssh.ParsePrivateKey(buffer)
	cliutils.CheckError(err)
	sshConfig := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	hostAndPort := host + ":" + strconv.Itoa(port)
	connection, err := ssh.Dial("tcp", hostAndPort, sshConfig)
	cliutils.CheckError(err)
	defer connection.Close()

	session, err := connection.NewSession()
	cliutils.CheckError(err)
	defer session.Close()

	stdout, err := session.StdoutPipe()
	cliutils.CheckError(err)

	var buf bytes.Buffer
	go io.Copy(&buf, stdout)

	session.Run("jfrog-authenticate")

	var result SshAuthResult
	err = json.Unmarshal(buf.Bytes(), &result)
	cliutils.CheckError(err)
	details.Url = cliutils.AddTrailingSlashIfNeeded(result.Href)
	details.SshAuthHeaders = result.Headers
	fmt.Println("SSH authentication successful.")
}

func parseUrl(url string) (protocol, host string, port int) {
	pattern1 := "^(.+)://(.+):([0-9].+)/$"
	pattern2 := "^(.+)://(.+)$"

	r, err := regexp.Compile(pattern1)
	cliutils.CheckError(err)
	groups := r.FindStringSubmatch(url)
	if len(groups) == 4 {
		protocol = groups[1]
		host = groups[2]
		port, err = strconv.Atoi(groups[3])
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, "URL: "+url+" is invalid. Expecting ssh://<host>:<port> or http(s)://...")
		}
		return
	}

	r, err = regexp.Compile(pattern2)
	cliutils.CheckError(err)
	groups = r.FindStringSubmatch(url)
	if len(groups) == 3 {
		protocol = groups[1]
		host = groups[2]
		port = 80
		return
	}
	return
}

type SshAuthResult struct {
	Href    string
	Headers map[string]string
}
