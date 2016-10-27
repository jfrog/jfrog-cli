package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"net/http"
	"net/url"
	"os"
	"crypto/x509"
	"io/ioutil"
	"errors"
	"strings"
	"crypto/tls"
)

func GetEncryptedPasswordFromArtifactory(artifactoryDetails *config.ArtifactoryDetails) (*http.Response, string, error) {
	err := initTransport(artifactoryDetails)
	if(err != nil){
		return nil, "", err
	}
	apiUrl := artifactoryDetails.Url + "api/security/encryptedPassword"
	httpClientsDetails := GetArtifactoryHttpClientDetails(artifactoryDetails)
	resp, body, _, err := ioutils.SendGet(apiUrl, true, httpClientsDetails)
	return resp, string(body), err
}

func UploadFile(f *os.File, url string, artifactoryDetails *config.ArtifactoryDetails,
details *ioutils.FileDetails, httpClientsDetails ioutils.HttpClientDetails) (*http.Response, error) {
	var err error
	if details == nil {
		details, err = ioutils.GetFileDetails(f.Name())
	}
	if err != nil {
	    return nil, err
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

func loadCertificates(caCertPool *x509.CertPool) error {
	securityDir, err := getJfrogSecurityDir()
	if err != nil {
	    return err
	}
	if !ioutils.IsPathExists(securityDir) {
		return nil
	}
	files, err := ioutil.ReadDir(securityDir)
	err = cliutils.CheckError(err)
	if err != nil {
	    return err
	}
	for _, file := range files {
		caCert, err := ioutil.ReadFile(securityDir + file.Name())
		err = cliutils.CheckError(err)
        if err != nil {
            return err
        }
		caCertPool.AppendCertsFromPEM(caCert)
	}
	return nil
}

func getJfrogSecurityDir() (string, error) {
	confPath, err := config.GetJfrogHomeDir()
    if err != nil {
        return "", err
    }
	return confPath + "security/", nil
}

func initTransport(artDetails *config.ArtifactoryDetails) error {
	caCertPool, err := x509.SystemCertPool()
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}
	err = loadCertificates(caCertPool)
	if err != nil {
	    return err
	}
	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
		ClientSessionCache: tls.NewLRUClientSessionCache(1)}
	tlsConfig.BuildNameToCertificate()
	artDetails.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	return nil
}

func PreCommandSetup(flags CommonFlag) (err error) {
	if flags.GetArtifactoryDetails().SshKeyPath != "" {
		SshAuthentication(flags.GetArtifactoryDetails())
	}
	if !flags.IsDryRun() {
		err = initTransport(flags.GetArtifactoryDetails())
		if(err != nil){
			return
		}
	}
	return
}

func GetArtifactoryHttpClientDetails(artifactoryDetails *config.ArtifactoryDetails) ioutils.HttpClientDetails {
	return ioutils.HttpClientDetails{
		User:      artifactoryDetails.User,
		Password:  artifactoryDetails.Password,
		ApiKey:    artifactoryDetails.ApiKey,
		Headers:   artifactoryDetails.SshAuthHeaders,
		Transport: artifactoryDetails.Transport}
}

func BuildArtifactoryUrl(baseUrl, path string, params map[string]string) (string, error) {
	escapedUrl, err := url.Parse(baseUrl + path)
	err = cliutils.CheckError(err)
	if err != nil {
	    return "", nil
	}
	q := escapedUrl.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	escapedUrl.RawQuery = q.Encode()
	return escapedUrl.String(), nil
}

func IsWildcardPattern(pattern string) bool {
	return strings.Contains(pattern, "*") || strings.HasSuffix(pattern, "/") || !strings.Contains(pattern, "/")
}

func EncodeParams(props string) (string, error) {
	propList := strings.Split(props, ";")
	result := []string{}
	for _, prop := range propList {
		key, value, err := SplitProp(prop)
		if err != nil {
			return "", err
		}
		result = append(result, url.QueryEscape(key) + "=" + url.QueryEscape(value))
	}

	return strings.Join(result, ";"), nil
}

// Simple directory path - dir path without wildcards.
func IsSimpleDirectoryPath(path string) bool {
	return IsDirectoryPath(path) && !strings.Contains(path, "*")
}

func IsDirectoryPath(path string) bool {
	return path != "" && strings.HasSuffix(path, "/")
}

type CommonFlag interface {
	GetArtifactoryDetails() *config.ArtifactoryDetails
	IsDryRun() bool
}

func SplitProp(prop string) (string, string, error) {
	splitIndex := strings.Index(prop, "=")
	if splitIndex < 1 || len(prop[splitIndex + 1:]) < 1 {
		err := cliutils.CheckError(errors.New("Invalid property: " + prop))
		return "", "", err
	}
	return prop[:splitIndex], prop[splitIndex + 1:], nil

}