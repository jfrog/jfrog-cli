package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"net/http"
	"net/url"
	"os"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"fmt"
	"strings"
)

func GetEncryptedPasswordFromArtifactory(artifactoryDetails *config.ArtifactoryDetails) (*http.Response, string) {
	PingArtifactory(artifactoryDetails)
	apiUrl := artifactoryDetails.Url + "api/security/encryptedPassword"
	httpClientsDetails := GetArtifactoryHttpClientDetails(artifactoryDetails)
	resp, body, _, _ := ioutils.SendGet(apiUrl, true, httpClientsDetails)
	return resp, string(body)
}

func UploadFile(f *os.File, url string, artifactoryDetails *config.ArtifactoryDetails,
details *ioutils.FileDetails, httpClientsDetails ioutils.HttpClientDetails) *http.Response {
	if details == nil {
		details = ioutils.GetFileDetails(f.Name())
	}
	headers := map[string]string{
		"X-Checksum-Sha1": details.Sha1,
		"X-Checksum-Md5":  details.Md5,
	}
	AddAuthHeaders(headers, artifactoryDetails)
	requestClientDetails := httpClientsDetails.Clone()
	cliutils.MergeMaps(headers, requestClientDetails.Headers)

	return ioutils.UploadFile(f, url, *requestClientDetails)
}

func AddAuthHeaders(headers map[string]string, artifactoryDetails *config.ArtifactoryDetails) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}
	if artifactoryDetails.SshAuthHeaders != nil {
		cliutils.MergeMaps(artifactoryDetails.SshAuthHeaders, headers)
	}
	return headers
}

func loadCertificates(caCertPool *x509.CertPool) {
	securityDir := getJfrogSecurityFolder()
	if !ioutils.IsPathExists(securityDir) {
		return
	}
	files, err := ioutil.ReadDir(securityDir)
	cliutils.CheckError(err)
	fmt.Println("Loading certificates...")
	for _, file := range files {
		caCert, err := ioutil.ReadFile(securityDir + file.Name())
		cliutils.CheckError(err)
		caCertPool.AppendCertsFromPEM(caCert)
	}
}

func getJfrogSecurityFolder() string {
	return config.GetJfrogHomeFolder() + "security/"
}

func initTransport() *http.Transport {
	caCertPool := x509.NewCertPool()
	loadCertificates(caCertPool)
	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
		ClientSessionCache: tls.NewLRUClientSessionCache(1)}
	tlsConfig.BuildNameToCertificate()
	tempTransport := &http.Transport{TLSClientConfig: tlsConfig}
	return tempTransport
}

func PreCommandSetup(flags *Flags) {
	if flags.ArtDetails.SshKeyPath != "" {
		SshAuthentication(flags.ArtDetails)
	}
	if !flags.DryRun {
		PingArtifactory(flags.ArtDetails)
	}
}

func PingArtifactory(artDetails *config.ArtifactoryDetails) {
	defer func() {
		if r := recover(); r != nil {
			artDetails.Transport = initTransport()
		}
	}()
	httpClientsDetails := GetArtifactoryHttpClientDetails(artDetails)
	fmt.Println("Pinging Artifactory...")
	ioutils.SendGet(artDetails.Url, true, httpClientsDetails)
}

func GetArtifactoryHttpClientDetails(artifactoryDetails *config.ArtifactoryDetails) ioutils.HttpClientDetails {
	return ioutils.HttpClientDetails{
		User:      artifactoryDetails.User,
		Password:  artifactoryDetails.Password,
		Headers:   artifactoryDetails.SshAuthHeaders,
		Transport: artifactoryDetails.Transport}
}

func BuildArtifactoryUrl(baseUrl, path string, params map[string]string) string {
	escapedUrl, err := url.Parse(baseUrl)
	cliutils.CheckError(err)
	escapedUrl.Path += path
	q := escapedUrl.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	escapedUrl.RawQuery = q.Encode()
	return escapedUrl.String()
}

func IsWildcardPattern(pattern string) bool {
	return strings.Contains(pattern, "*") || strings.HasSuffix(pattern, "/") || !strings.Contains(pattern, "/")
}

type Flags struct {
	ArtDetails   *config.ArtifactoryDetails
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
