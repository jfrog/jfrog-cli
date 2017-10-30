package auth

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
)

type ArtifactoryDetails struct {
	Url            string            `json:"-"`
	User           string            `json:"-"`
	Password       string            `json:"-"`
	ApiKey         string            `json:"-"`
	SshAuthHeaders map[string]string `json:"-"`
}

func (rt *ArtifactoryDetails) SshAuthentication(sshKey, sshPassphrase []byte) (map[string]string, string, error) {
	return sshAuthentication(rt.Url, sshKey, sshPassphrase)
}

func (rt *ArtifactoryDetails) GetSshAuthHeaders() map[string]string {
	return rt.SshAuthHeaders
}

func (rt *ArtifactoryDetails) GetUser() string {
	return rt.User
}

func (rt *ArtifactoryDetails) CreateArtifactoryHttpClientDetails() httputils.HttpClientDetails {
	return httputils.HttpClientDetails{
		User:     rt.User,
		Password: rt.Password,
		ApiKey:   rt.ApiKey,
		Headers:  cliutils.CopyMap(rt.SshAuthHeaders)}
}
