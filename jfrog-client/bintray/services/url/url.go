package url

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"path"
)

func NewService(client *httpclient.HttpClient) *UrlService {
	us := &UrlService{client: client}
	return us
}

func NewURLParams() *Params {
	return &Params{PathDetails: &utils.PathDetails{}}
}

type UrlService struct {
	client         *httpclient.HttpClient
	BintrayDetails auth.BintrayDetails
}

type Params struct {
	*utils.PathDetails
	Expiry         int64
	ValidFor       int
	CallbackId     string
	CallbackEmail  string
	CallbackUrl    string
	CallbackMethod string
}

func (us *UrlService) SignVersion(params *Params) error {
	if us.BintrayDetails.GetUser() == "" {
		us.BintrayDetails.SetUser(params.Subject)
	}
	content, err := createSignVersionContent(params)
	if err != nil {
		return err
	}
	signVersionUrl := us.BintrayDetails.GetApiUrl() + path.Join("signed_url", params.Subject, params.Repo, params.Path)

	log.Info("Signing URL...")
	httpClientsDetails := us.BintrayDetails.CreateHttpClientDetails()
	resp, body, err := httputils.SendPost(signVersionUrl, content, httpClientsDetails)
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

func createSignVersionContent(params *Params) ([]byte, error) {
	Config := contentConfig{
		Expiry:         params.Expiry,
		ValidFor:       params.ValidFor,
		CallbackId:     params.CallbackId,
		CallbackEmail:  params.CallbackEmail,
		CallbackUrl:    params.CallbackUrl,
		CallbackMethod: params.CallbackMethod,
	}
	requestContent, err := json.Marshal(Config)
	if err != nil {
		return nil, errorutils.CheckError(errors.New("Failed to execute request."))
	}
	return requestContent, nil
}

type contentConfig struct {
	Expiry         int64  `json:"expiry,omitempty"`
	ValidFor       int    `json:"valid_for_secs,omitempty"`
	CallbackId     string `json:"callback_id,omitempty"`
	CallbackEmail  string `json:"callback_email,omitempty"`
	CallbackUrl    string `json:"callback_url,omitempty"`
	CallbackMethod string `json:"callback_method,omitempty"`
}
