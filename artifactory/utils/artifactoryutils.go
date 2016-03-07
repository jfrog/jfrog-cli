package utils

import (
	"github.com/JFrogDev/jfrog-cli-go/cliutils"
	"net/http"
	"os"
)

func GetEncryptedPasswordFromArtifactory(artifactoryDetails *cliutils.ArtifactoryDetails) (*http.Response, string) {
	apiUrl := artifactoryDetails.Url + "api/security/encryptedPassword"
	resp, body, _, _ := cliutils.SendGet(apiUrl, nil, true, artifactoryDetails.User, artifactoryDetails.Password)
	return resp, string(body)
}

func UploadFile(f *os.File, url string, artifactoryDetails *cliutils.ArtifactoryDetails,
	details *cliutils.FileDetails) *http.Response {
	if details == nil {
		details = cliutils.GetFileDetails(f.Name())
	}
	headers := map[string]string{
		"X-Checksum-Sha1": details.Sha1,
		"X-Checksum-Md5":  details.Md5,
	}
	AddAuthHeaders(headers, artifactoryDetails)

	return cliutils.UploadFile(f, url, artifactoryDetails.User, artifactoryDetails.Password, headers)
}

func AddAuthHeaders(headers map[string]string, artifactoryDetails *cliutils.ArtifactoryDetails) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}
	if artifactoryDetails.SshAuthHeaders != nil {
		for name := range artifactoryDetails.SshAuthHeaders {
			headers[name] = artifactoryDetails.SshAuthHeaders[name]
		}
	}
	return headers
}

type Flags struct {
	ArtDetails   *cliutils.ArtifactoryDetails
	DryRun       bool
	Props        string
	Deb          string
	Recursive    bool
	Flat         bool
	UseRegExp    bool
	Threads      int
	MinSplitSize int64
	SplitCount   int
	Interactive  bool
	EncPassword  bool
}
