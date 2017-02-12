package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"errors"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"fmt"
)

func ShowAccessKeys(bintrayDetails *config.BintrayDetails, org string) error {
	path := GetAccessKeysPath(bintrayDetails, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	log.Info("Getting access keys...")
	resp, body, _, _ := httputils.SendGet(path, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Access keys details:")
	fmt.Println(cliutils.IndentJson(body))
	return nil
}

func ShowAccessKey(flags *AccessKeyFlags, org string) (err error) {
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Getting access key...")
	resp, body, _, _ := httputils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Access keys details:")
	fmt.Println(cliutils.IndentJson(body))
	return
}

func CreateAccessKey(flags *AccessKeyFlags, org string) (err error) {
	data := BuildAccessKeyJson(flags, true)
	url := GetAccessKeysPath(flags.BintrayDetails, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Creating access key...")
	resp, body, err := httputils.SendPost(url, []byte(data), httpClientsDetails)
	if err != nil {
		return
	}
	if resp.StatusCode != 201 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Created access key, details:")
	fmt.Println(cliutils.IndentJson(body))
	return
}

func UpdateAccessKey(flags *AccessKeyFlags, org string) error {
	data := BuildAccessKeyJson(flags, false)
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Updating access key...")
	resp, body, err := httputils.SendPatch(url, []byte(data), httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Updated access key, details:")
	fmt.Println(cliutils.IndentJson(body))
	return nil
}

func BuildAccessKeyJson(flags *AccessKeyFlags, create bool) string {
	var existenceCheck string
	var whiteCidrs string
	var blackCidrs string
	if flags.ExistenceCheckUrl != "" {
		existenceCheck = "\"existence_check\": {" +
				"\"url\": \"" + flags.ExistenceCheckUrl + "\"," +
				"\"cache_for_secs\": \"" + strconv.Itoa(flags.ExistenceCheckCache) + "\"" +
				"}"
	}
	if flags.WhiteCidrs != "" {
		whiteCidrs = "\"white_cidrs\": " + cliutils.BuildListString(flags.WhiteCidrs)
	}
	if flags.BlackCidrs != "" {
		blackCidrs = "\"black_cidrs\": " + cliutils.BuildListString(flags.BlackCidrs)
	}
	data := "{"
	if create {
		data += "\"id\": \"" + flags.Id + "\","
	}
	if flags.Password != "" {
		data += "\"password\": \"" + flags.Password + "\","
	}
	data += "\"expiry\": \"" + flags.Expiry + "\""

	if existenceCheck != "" {
		data += "," + existenceCheck
	}
	if whiteCidrs != "" {
		data += "," + whiteCidrs
	}
	if blackCidrs != "" {
		data += "," + blackCidrs
	}
	data += "}"

	return data
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
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
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
	Expiry              string
	ExistenceCheckUrl   string
	ExistenceCheckCache int
	WhiteCidrs          string
	BlackCidrs          string
}
