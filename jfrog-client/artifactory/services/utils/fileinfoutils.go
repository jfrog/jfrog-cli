package utils

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
