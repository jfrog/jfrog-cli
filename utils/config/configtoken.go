package config

import (
	"encoding/base64"
	"encoding/json"
)

const tokenVersion = 1

type configToken struct {
	Version                  int    `json:"version,omitempty"`
	Url                      string `json:"url,omitempty"`
	User                     string `json:"user,omitempty"`
	Password                 string `json:"password,omitempty"`
	SshKeyPath               string `json:"sshKeyPath,omitempty"`
	SshPassphrase            string `json:"sshPassphrase,omitempty"`
	AccessToken              string `json:"accessToken,omitempty"`
	ClientCertificatePath    string `json:"clientCertificatePath,omitempty"`
	ClientCertificateKeyPath string `json:"clientCertificateKeyPath,omitempty"`
	ServerId                 string `json:"serverId,omitempty"`
	ApiKey                   string `json:"apiKey,omitempty"`
}

func fromArtifactoryDetails(details *ArtifactoryDetails) *configToken {
	return &configToken{
		Version:                  tokenVersion,
		Url:                      details.Url,
		User:                     details.User,
		Password:                 details.Password,
		SshKeyPath:               details.SshKeyPath,
		SshPassphrase:            details.SshPassphrase,
		AccessToken:              details.AccessToken,
		ClientCertificatePath:    details.ClientCertificatePath,
		ClientCertificateKeyPath: details.ClientCertificateKeyPath,
		ServerId:                 details.ServerId,
		ApiKey:                   details.ApiKey,
	}
}

func toArtifactoryDetails(detailsSerialization *configToken) *ArtifactoryDetails {
	return &ArtifactoryDetails{
		Url:                      detailsSerialization.Url,
		User:                     detailsSerialization.User,
		Password:                 detailsSerialization.Password,
		SshKeyPath:               detailsSerialization.SshKeyPath,
		SshPassphrase:            detailsSerialization.SshPassphrase,
		AccessToken:              detailsSerialization.AccessToken,
		ClientCertificatePath:    detailsSerialization.ClientCertificatePath,
		ClientCertificateKeyPath: detailsSerialization.ClientCertificateKeyPath,
		ServerId:                 detailsSerialization.ServerId,
		ApiKey:                   detailsSerialization.ApiKey,
	}
}

func Export(details *ArtifactoryDetails) (string, error) {
	buffer, err := json.Marshal(fromArtifactoryDetails(details))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buffer), nil
}

func Import(serverToken string) (*ArtifactoryDetails, error) {
	decoded, err := base64.StdEncoding.DecodeString(serverToken)
	if err != nil {
		return nil, err
	}
	token := &configToken{}
	if err = json.Unmarshal(decoded, token); err != nil {
		return nil, err
	}
	return toArtifactoryDetails(token), nil
}
