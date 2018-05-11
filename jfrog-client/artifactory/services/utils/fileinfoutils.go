package utils

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
)

type FileHashes struct {
	Sha256 string `json:"sha256,omitempty"`
	Sha1   string `json:"sha1,omitempty"`
	Md5    string `json:"md5,omitempty"`
}

type FileInfo struct {
	*FileHashes
	LocalPath       string `json:"localPath,omitempty"`
	ArtifactoryPath string `json:"artifactoryPath,omitempty"`
}

func (fileInfo *FileInfo) ToBuildArtifact() buildinfo.InternalArtifact {
	artifact := buildinfo.InternalArtifact{Checksum: &buildinfo.Checksum{}}
	artifact.Sha1 = fileInfo.Sha1
	artifact.Md5 = fileInfo.Md5
	artifact.Path = fileInfo.ArtifactoryPath
	return artifact
}
