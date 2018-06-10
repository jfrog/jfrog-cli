package auth

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
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
