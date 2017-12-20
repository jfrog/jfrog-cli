package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"encoding/json"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
)

func ShowAccessKeys(bintrayDetails *config.BintrayDetails, org string) error {
	path := GetAccessKeysPath(bintrayDetails, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	log.Info("Getting access keys...")
	resp, body, _, _ := httputils.SendGet(path, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func ShowAccessKey(flags *AccessKeyFlags, org string) error {
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Getting access key...")
	resp, body, _, _ := httputils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func CreateAccessKey(flags *AccessKeyFlags, org string) error {
	data, err := BuildAccessKeyJson(flags)
	if err != nil {
		return err
	}
	url := GetAccessKeysPath(flags.BintrayDetails, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Creating access key...")
	resp, body, err := httputils.SendPost(url, []byte(data), httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func UpdateAccessKey(flags *AccessKeyFlags, org string) error {
	data, err := BuildAccessKeyJson(flags)
	if err != nil {
		return err
	}
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Updating access key...")
	resp, body, err := httputils.SendPatch(url, []byte(data), httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func BuildAccessKeyJson(flags *AccessKeyFlags) (string, error) {
	apiOnly, err := cliutils.StringToBool(flags.ApiOnly, true)
	if err != nil {
		return "", err
	}

	var whiteCidrs []string
	if flags.WhiteCidrs != "" {
		whiteCidrs = strings.Split(flags.WhiteCidrs, ",")
	}

	var blackCidrs []string
	if flags.BlackCidrs != "" {
		blackCidrs = strings.Split(flags.BlackCidrs, ",")
	}

	data := AccessKeyConfig{
		Id:         flags.Id,
		Expiry:     flags.Expiry,
		WhiteCidrs: whiteCidrs,
		BlackCidrs: blackCidrs,
		ApiOnly:    apiOnly,
		ExistenceCheck: ExistenceCheckConfig{
			Url:             flags.ExistenceCheckUrl,
			Cache_for_secs:  flags.ExistenceCheckCache},
	}
	requestContent, err := json.Marshal(data)
	if err != nil {
		return "", errorutils.CheckError(errors.New("Failed to execute request. " + cliutils.GetDocumentationMessage()))
	}
	return string(requestContent), nil
}

func DeleteAccessKey(flags *AccessKeyFlags, org string) error {
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Deleting access key...")
	resp, body, err := httputils.SendDelete(url, nil, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Deleted access key.")
	return nil
}

func GetAccessKeyPath(bintrayDetails *config.BintrayDetails, id, org string) string {
	return GetAccessKeysPath(bintrayDetails, org) + "/" + id
}

func GetAccessKeysPath(bintrayDetails *config.BintrayDetails, org string) string {
	if org == "" {
		return bintrayDetails.ApiUrl + "users/" + bintrayDetails.User + "/download_keys"
	}
	return bintrayDetails.ApiUrl + "orgs/" + org + "/download_keys"
}

type AccessKeyFlags struct {
	BintrayDetails      *config.BintrayDetails
	Id                  string
	Password            string
	Expiry              int64
	ExistenceCheckUrl   string
	ExistenceCheckCache int
	WhiteCidrs          string
	BlackCidrs          string
	ApiOnly             string
}

type AccessKeyConfig struct {
	Id             string                `json:"id,omitempty"`
	Expiry         int64                 `json:"expiry,omitempty"`
	ExistenceCheck ExistenceCheckConfig  `json:"existence_check,omitempty"`
	WhiteCidrs     []string              `json:"white_cidrs,omitempty"`
	BlackCidrs     []string              `json:"black_cidrs,omitempty"`
	ApiOnly        bool                  `json:"api_only"`
}

type ExistenceCheckConfig struct {
	Url            string  `json:"url,omitempty"`
	Cache_for_secs int     `json:"cache_for_secs,omitempty"`
}