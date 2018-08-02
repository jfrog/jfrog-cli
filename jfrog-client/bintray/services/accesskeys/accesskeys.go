package accesskeys

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"path"
	"strings"
)

func NewService(client *httpclient.HttpClient) *AccessKeysService {
	us := &AccessKeysService{client: client}
	return us
}

func NewAccessKeysParams() *Params {
	return &Params{}
}

type AccessKeysService struct {
	client         *httpclient.HttpClient
	BintrayDetails auth.BintrayDetails
}

type Params struct {
	Id                  string
	Password            string
	Org                 string
	Expiry              int64
	ExistenceCheckUrl   string
	ExistenceCheckCache int
	WhiteCidrs          string
	BlackCidrs          string
	ApiOnly             bool
}

func (aks *AccessKeysService) ShowAll(org string) error {
	path := getAccessKeysPath(aks.BintrayDetails, org)
	httpClientsDetails := aks.BintrayDetails.CreateHttpClientDetails()
	log.Info("Getting access keys...")
	client := httpclient.NewDefaultHttpClient()
	resp, body, _, _ := client.SendGet(path, true, httpClientsDetails)
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func (aks *AccessKeysService) Show(org, id string) error {
	url := getAccessKeyPath(aks.BintrayDetails, id, org)
	httpClientsDetails := aks.BintrayDetails.CreateHttpClientDetails()
	log.Info("Getting access key...")
	client := httpclient.NewDefaultHttpClient()
	resp, body, _, _ := client.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func (aks *AccessKeysService) Create(params *Params) error {
	data, err := buildAccessKeyJson(params)
	if err != nil {
		return err
	}
	url := getAccessKeysPath(aks.BintrayDetails, params.Org)
	httpClientsDetails := aks.BintrayDetails.CreateHttpClientDetails()
	log.Info("Creating access key...")
	client := httpclient.NewDefaultHttpClient()
	resp, body, err := client.SendPost(url, []byte(data), httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func (aks *AccessKeysService) Update(params *Params) error {
	data, err := buildAccessKeyJson(params)
	if err != nil {
		return err
	}
	url := getAccessKeyPath(aks.BintrayDetails, params.Id, params.Org)
	httpClientsDetails := aks.BintrayDetails.CreateHttpClientDetails()
	log.Info("Updating access key...")
	client := httpclient.NewDefaultHttpClient()
	resp, body, err := client.SendPatch(url, []byte(data), httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func (aks *AccessKeysService) Delete(org, id string) error {
	url := getAccessKeyPath(aks.BintrayDetails, id, org)
	httpClientsDetails := aks.BintrayDetails.CreateHttpClientDetails()
	log.Info("Deleting access key...")
	client := httpclient.NewDefaultHttpClient()
	resp, body, err := client.SendDelete(url, nil, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Deleted access key.")
	return nil
}

func buildAccessKeyJson(params *Params) ([]byte, error) {
	var whiteCidrs []string
	if params.WhiteCidrs != "" {
		whiteCidrs = strings.Split(params.WhiteCidrs, ",")
	}

	var blackCidrs []string
	if params.BlackCidrs != "" {
		blackCidrs = strings.Split(params.BlackCidrs, ",")
	}

	data := AccessKeyContent{
		Id:         params.Id,
		Expiry:     params.Expiry,
		WhiteCidrs: whiteCidrs,
		BlackCidrs: blackCidrs,
		ApiOnly:    params.ApiOnly,
		ExistenceCheck: ExistenceCheckContent{
			Url:            params.ExistenceCheckUrl,
			Cache_for_secs: params.ExistenceCheckCache},
	}
	content, err := json.Marshal(data)
	if err != nil {
		return []byte{}, errorutils.CheckError(errors.New("Failed to execute request."))
	}
	return content, nil
}

func getAccessKeyPath(bintrayDetails auth.BintrayDetails, id, org string) string {
	return getAccessKeysPath(bintrayDetails, org) + "/" + id
}

func getAccessKeysPath(bintrayDetails auth.BintrayDetails, org string) string {
	if org == "" {
		return bintrayDetails.GetApiUrl() + path.Join("users", bintrayDetails.GetUser(), "access_keys")
	}
	return bintrayDetails.GetApiUrl() + path.Join("orgs", org, "access_keys")
}

type AccessKeyContent struct {
	Id             string                `json:"id,omitempty"`
	Expiry         int64                 `json:"expiry,omitempty"`
	ExistenceCheck ExistenceCheckContent `json:"existence_check,omitempty"`
	WhiteCidrs     []string              `json:"white_cidrs,omitempty"`
	BlackCidrs     []string              `json:"black_cidrs,omitempty"`
	ApiOnly        bool                  `json:"api_only"`
}

type ExistenceCheckContent struct {
	Url            string `json:"url,omitempty"`
	Cache_for_secs int    `json:"cache_for_secs,omitempty"`
}
