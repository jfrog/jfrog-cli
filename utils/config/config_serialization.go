package config

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
)

const artifactoryDetailsSerializationVersion = 1

type ArtifactoryDetailsSerialization struct {
	Version        int               `json:"version,omitempty"`
	Url            string            `json:"url,omitempty"`
	SshUrl         string            `json:"sshUrl,omitempty"`
	User           string            `json:"user,omitempty"`
	Password       string            `json:"password,omitempty"`
	SshKeyPath     string            `json:"sshKeyPath,omitempty"`
	SshPassphrase  string            `json:"sshPassphrase,omitempty"`
	SshAuthHeaders map[string]string `json:"sshAuthHeaders,omitempty"`
	AccessToken    string            `json:"accessToken,omitempty"`
	ServerId       string            `json:"serverId,omitempty"`
	InsecureTls    bool              `json:"insecureTls,omitempty"`
	ApiKey         string            `json:"apiKey,omitempty"`
}

func fromArtifactoryDetails(details *ArtifactoryDetails) *ArtifactoryDetailsSerialization {
	return &ArtifactoryDetailsSerialization{
		Version:        artifactoryDetailsSerializationVersion,
		Url:            details.Url,
		SshUrl:         details.SshUrl,
		User:           details.User,
		Password:       details.Password,
		SshKeyPath:     details.SshKeyPath,
		SshPassphrase:  details.SshPassphrase,
		SshAuthHeaders: details.SshAuthHeaders,
		AccessToken:    details.AccessToken,
		ServerId:       details.ServerId,
		InsecureTls:    details.InsecureTls,
		ApiKey:         details.ApiKey,
	}
}

func toArtifactoryDetails(detailsSerialization *ArtifactoryDetailsSerialization) *ArtifactoryDetails {
	return &ArtifactoryDetails{
		Url:            detailsSerialization.Url,
		SshUrl:         detailsSerialization.SshUrl,
		User:           detailsSerialization.User,
		Password:       detailsSerialization.Password,
		SshKeyPath:     detailsSerialization.SshKeyPath,
		SshPassphrase:  detailsSerialization.SshPassphrase,
		SshAuthHeaders: detailsSerialization.SshAuthHeaders,
		AccessToken:    detailsSerialization.AccessToken,
		ServerId:       detailsSerialization.ServerId,
		InsecureTls:    detailsSerialization.InsecureTls,
		ApiKey:         detailsSerialization.ApiKey,
	}
}

func Export(details *ArtifactoryDetails) (string, error) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(fromArtifactoryDetails(details)); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

func Import(serverToken string) (*ArtifactoryDetails, error) {
	decoded, err := base64.StdEncoding.DecodeString(serverToken)
	if err != nil {
		return nil, err
	}
	buffer := bytes.Buffer{}
	buffer.Write(decoded)
	decoder := gob.NewDecoder(&buffer)
	artifactoryDetails := &ArtifactoryDetailsSerialization{}
	if err = decoder.Decode(artifactoryDetails); err != nil {
		return nil, err
	}
	return toArtifactoryDetails(artifactoryDetails), nil
}
