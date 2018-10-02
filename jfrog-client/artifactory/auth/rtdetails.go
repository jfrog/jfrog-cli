package auth

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"strings"
)

func NewArtifactoryDetails() ArtifactoryDetails {
	return &artifactoryDetails{}
}

type ArtifactoryDetails interface {
	GetUrl() string
	GetUser() string
	GetPassword() string
	GetApiKey() string
	GetSshAuthHeaders() map[string]string
	GetVersion() (string, error)
	SetUrl(url string)
	SetUser(user string)
	SetPassword(password string)
	SetApiKey(apiKey string)
	SetSshAuthHeaders(sshAuthHeaders map[string]string)

	AuthenticateSsh(sshKey, sshPassphrase []byte) error

	CreateHttpClientDetails() httputils.HttpClientDetails
}

type artifactoryDetails struct {
	Url            string            `json:"-"`
	User           string            `json:"-"`
	Password       string            `json:"-"`
	ApiKey         string            `json:"-"`
	version        string            `json:"-"`
	SshAuthHeaders map[string]string `json:"-"`
}

func (rt *artifactoryDetails) GetUrl() string {
	return rt.Url
}

func (rt *artifactoryDetails) GetUser() string {
	return rt.User
}

func (rt *artifactoryDetails) GetPassword() string {
	return rt.Password
}

func (rt *artifactoryDetails) GetApiKey() string {
	return rt.ApiKey
}

func (rt *artifactoryDetails) GetSshAuthHeaders() map[string]string {
	return rt.SshAuthHeaders
}

func (rt *artifactoryDetails) SetUrl(url string) {
	rt.Url = url
}

func (rt *artifactoryDetails) SetUser(user string) {
	rt.User = user
}

func (rt *artifactoryDetails) SetPassword(password string) {
	rt.Password = password
}

func (rt *artifactoryDetails) SetApiKey(apiKey string) {
	rt.ApiKey = apiKey
}

func (rt *artifactoryDetails) SetSshAuthHeaders(sshAuthHeaders map[string]string) {
	rt.SshAuthHeaders = sshAuthHeaders
}

func (rt *artifactoryDetails) AuthenticateSsh(sshKey, sshPassphrase []byte) error {
	sshHeaders, baseUrl, err := sshAuthentication(rt.Url, sshKey, sshPassphrase)
	if err != nil {
		return err
	}
	rt.SshAuthHeaders = sshHeaders
	rt.Url = baseUrl
	return nil
}

func (rt *artifactoryDetails) CreateHttpClientDetails() httputils.HttpClientDetails {
	return httputils.HttpClientDetails{
		User:     rt.User,
		Password: rt.Password,
		ApiKey:   rt.ApiKey,
		Headers:  utils.CopyMap(rt.SshAuthHeaders)}
}

func (rt *artifactoryDetails) GetVersion() (string, error) {
	var err error
	if rt.version == "" {
		rt.version, err = rt.getArtifactoryVersion()
		if err != nil {
			return "", err
		}
		log.Debug("The Artifactory version is:", rt.version)
	}
	return rt.version, nil
}

func (rt *artifactoryDetails) getArtifactoryVersion() (string, error) {
	client := httpclient.NewDefaultHttpClient()
	resp, body, _, err := client.SendGet(rt.GetUrl()+"api/system/version", true, rt.CreateHttpClientDetails())
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + utils.IndentJson(body)))
	}

	serverValues := strings.Split(resp.Header.Get("Server"), "/")
	if len(serverValues) != 2 {
		err = errors.New("Cannot parse Artifactory version from the server header.")
	}
	return strings.TrimSpace(serverValues[1]), errorutils.CheckError(err)
}
