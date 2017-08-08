package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"net/http"
	"net/url"
	"os"
	"crypto/x509"
	"io/ioutil"
	"errors"
	"strings"
	"crypto/tls"
)

const ARTIFACTORY_SYMLINK = "symlink.dest"
const SYMLINK_SHA1 = "symlink.destsha1"
func GetEncryptedPasswordFromArtifactory(artifactoryDetails *config.ArtifactoryDetails) (*http.Response, string, error) {
	err := initTransport(artifactoryDetails)
	if err != nil {
		return nil, "", err
	}
	apiUrl := artifactoryDetails.Url + "api/security/encryptedPassword"
	httpClientsDetails := GetArtifactoryHttpClientDetails(artifactoryDetails)
	resp, body, _, err := httputils.SendGet(apiUrl, true, httpClientsDetails)
	return resp, string(body), err
}

func UploadFile(f *os.File, url string, artifactoryDetails *config.ArtifactoryDetails, details *fileutils.FileDetails,
	httpClientsDetails httputils.HttpClientDetails) (*http.Response, []byte, error) {
	var err error
	if details == nil {
		details, err = fileutils.GetFileDetails(f.Name())
	}
	if err != nil {
		return nil, nil, err
	}
	headers := map[string]string{
		"X-Checksum-Sha1": details.Checksum.Sha1,
		"X-Checksum-Md5":  details.Checksum.Md5,
	}
	AddAuthHeaders(headers, artifactoryDetails)
	requestClientDetails := httpClientsDetails.Clone()
	cliutils.MergeMaps(headers, requestClientDetails.Headers)

	return httputils.UploadFile(f, url, *requestClientDetails)
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
	if !fileutils.IsPathExists(securityDir) {
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

func initTransport(artDetails *config.ArtifactoryDetails) (error) {
	// Remove once SystemCertPool supports windows
	caCertPool, err := LoadSystemRoots()

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

func PreCommandSetup(flags CommonFlags) (err error) {
	if flags.GetArtifactoryDetails().Ssh {
		err = SshAuthentication(flags.GetArtifactoryDetails())
		if err != nil {
			return
		}
	}
	if !flags.IsDryRun() {
		err = initTransport(flags.GetArtifactoryDetails())
		if err != nil {
			return
		}
	}
	return
}

func GetArtifactoryHttpClientDetails(artifactoryDetails *config.ArtifactoryDetails) httputils.HttpClientDetails {
	return httputils.HttpClientDetails{
		User:      artifactoryDetails.User,
		Password:  artifactoryDetails.Password,
		ApiKey:    artifactoryDetails.ApiKey,
		Headers:   cliutils.CopyMap(artifactoryDetails.SshAuthHeaders),
		Transport: artifactoryDetails.Transport}
}

func SetContentType(contentType string, headers *map[string]string) {
	AddHeader("Content-Type", contentType, headers)
}

func AddHeader(headerName, headerValue string, headers *map[string]string) {
	if *headers == nil {
		*headers = make(map[string]string)
	}
	(*headers)[headerName] = headerValue
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
		if prop == "" {
			continue
		}
		key, value, err := SplitProp(prop)
		if err != nil {
			return "", err
		}
		result = append(result, url.QueryEscape(key) + "=" + url.QueryEscape(value))
	}

	return strings.Join(result, ";"), nil
}

func SplitProp(prop string) (string, string, error) {
	splitIndex := strings.Index(prop, "=")
	if splitIndex < 1 || len(prop[splitIndex + 1:]) < 1 {
		err := cliutils.CheckError(errors.New("Invalid property: " + prop))
		return "", "", err
	}
	return prop[:splitIndex], prop[splitIndex + 1:], nil

}

// @paths - sorted array
// @index - index of the current path which we want to check if it a prefix of any of the other previous paths
// @separator - file separator
// returns true paths[index] is a prefix of any of the paths[i] where i<index , otherwise returns false
func IsSubPath(paths []string, index int, separator string) bool {
	currentPath := paths[index]
	if !strings.HasSuffix(currentPath, separator) {
		currentPath += separator
	}
	for i := index - 1; i >= 0; i-- {
		if strings.HasPrefix(paths[i], currentPath) {
			return true
		}
	}
	return false
}

type CommonFlags interface {
	GetArtifactoryDetails() *config.ArtifactoryDetails
	IsDryRun() bool
}

type CommonFlagsImpl struct {
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

func (flags *CommonFlagsImpl) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *CommonFlagsImpl) IsDryRun() bool {
	return flags.DryRun
}
