package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const tokenVersion = 1

type configToken struct {
	Version              int    `json:"version,omitempty"`
	Url                  string `json:"url,omitempty"`
	DistributionUrl      string `json:"distributionUrl,omitempty"`
	User                 string `json:"user,omitempty"`
	Password             string `json:"password,omitempty"`
	SshKeyPath           string `json:"sshKeyPath,omitempty"`
	SshPassphrase        string `json:"sshPassphrase,omitempty"`
	AccessToken          string `json:"accessToken,omitempty"`
	RefreshToken         string `json:"refreshToken,omitempty"`
	TokenRefreshInterval int    `json:"tokenRefreshInterval,omitempty"`
	ClientCertPath       string `json:"clientCertPath,omitempty"`
	ClientCertKeyPath    string `json:"clientCertKeyPath,omitempty"`
	ServerId             string `json:"serverId,omitempty"`
	ApiKey               string `json:"apiKey,omitempty"`
}

func fromArtifactoryDetails(details *ArtifactoryDetails) *configToken {
	return &configToken{
		Version:              tokenVersion,
		Url:                  details.Url,
		DistributionUrl:      details.DistributionUrl,
		User:                 details.User,
		Password:             details.Password,
		SshKeyPath:           details.SshKeyPath,
		SshPassphrase:        details.SshPassphrase,
		AccessToken:          details.AccessToken,
		RefreshToken:         details.RefreshToken,
		TokenRefreshInterval: details.TokenRefreshInterval,
		ClientCertPath:       details.ClientCertPath,
		ClientCertKeyPath:    details.ClientCertKeyPath,
		ServerId:             details.ServerId,
		ApiKey:               details.ApiKey,
	}
}

func toArtifactoryDetails(detailsSerialization *configToken) *ArtifactoryDetails {
	return &ArtifactoryDetails{
		Url:                  detailsSerialization.Url,
		DistributionUrl:      detailsSerialization.DistributionUrl,
		User:                 detailsSerialization.User,
		Password:             detailsSerialization.Password,
		SshKeyPath:           detailsSerialization.SshKeyPath,
		SshPassphrase:        detailsSerialization.SshPassphrase,
		AccessToken:          detailsSerialization.AccessToken,
		RefreshToken:         detailsSerialization.RefreshToken,
		TokenRefreshInterval: detailsSerialization.TokenRefreshInterval,
		ClientCertPath:       detailsSerialization.ClientCertPath,
		ClientCertKeyPath:    detailsSerialization.ClientCertKeyPath,
		ServerId:             detailsSerialization.ServerId,
		ApiKey:               detailsSerialization.ApiKey,
	}
}

func Export(details *ArtifactoryDetails) (string, error) {
	conf, err := readConf()
	if err != nil {
		return "", err
	}
	// If config is encrypted, ask for master key.
	if conf.Enc {
		masterKeyFromFile, _, err := getMasterKeyFromSecurityFile()
		if err != nil {
			return "", err
		}
		masterKeyFromConsole, err := readMasterKeyFromConsole()
		if err != nil {
			return "", err
		}
		if masterKeyFromConsole != masterKeyFromFile {
			return "", errorutils.CheckError(errors.New("could not generate config token: config is encrypted, and wrong master key was provided"))
		}
	}
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
