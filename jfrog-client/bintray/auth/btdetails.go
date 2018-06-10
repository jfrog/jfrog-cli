package auth

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
)

const BINTRAY_API_URL = "https://bintray.com/api/v1/"
const BINTRAY_DOWNLOAD_SERVER_URL = "https://dl.bintray.com/"

func NewBintrayDetails() BintrayDetails {
	return &bintrayDetails{
		ApiUrl:            BINTRAY_API_URL,
		DownloadServerUrl: BINTRAY_DOWNLOAD_SERVER_URL,
	}
}

type BintrayDetails interface {
	GetApiUrl() string
	GetDownloadServerUrl() string
	GetUser() string
	GetKey() string
	GetDefPackageLicense() string

	SetApiUrl(apiUrl string)
	SetDownloadServerUrl(downloadUrl string)
	SetUser(user string)
	SetKey(key string)
	SetDefPackageLicense(license string)

	CreateHttpClientDetails() httputils.HttpClientDetails

	Marshal() ([]byte, error)
}

type bintrayDetails struct {
	ApiUrl            string `json:"-"`
	DownloadServerUrl string `json:"-"`
	User              string `json:"user,omitempty"`
	Key               string `json:"key,omitempty"`
	DefPackageLicense string `json:"defPackageLicense,omitempty"`
}

func (bt *bintrayDetails) GetApiUrl() string {
	return bt.ApiUrl
}

func (bt *bintrayDetails) GetDownloadServerUrl() string {
	return bt.DownloadServerUrl
}

func (bt *bintrayDetails) GetUser() string {
	return bt.User
}

func (bt *bintrayDetails) GetKey() string {
	return bt.Key
}

func (bt *bintrayDetails) GetDefPackageLicense() string {
	return bt.DefPackageLicense
}

func (bt *bintrayDetails) SetApiUrl(apiUrl string) {
	bt.ApiUrl = apiUrl
}

func (bt *bintrayDetails) SetDownloadServerUrl(downloadUrl string) {
	bt.DownloadServerUrl = downloadUrl
}

func (bt *bintrayDetails) SetUser(user string) {
	bt.User = user
}

func (bt *bintrayDetails) SetKey(key string) {
	bt.Key = key
}

func (bt *bintrayDetails) SetDefPackageLicense(license string) {
	bt.DefPackageLicense = license
}

func (bt *bintrayDetails) CreateHttpClientDetails() httputils.HttpClientDetails {
	return httputils.HttpClientDetails{
		User:     bt.GetUser(),
		Password: bt.GetKey()}
}

func (bt *bintrayDetails) Marshal() ([]byte, error) {
	return json.Marshal(bt)
}
