package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"errors"
	"strconv"
)

func ShowAccessKeys(bintrayDetails *config.BintrayDetails, org string) (err error) {
	path := GetAccessKeysPath(bintrayDetails, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, _, _ := ioutils.SendGet(path, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". "+utils.ReadBintrayMessage(body)))
		if err != nil {
		    return
		}
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
	return
}

func ShowAccessKey(flags *AccessKeyFlags, org string) (err error) {
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, _, _ := ioutils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". "+utils.ReadBintrayMessage(body)))
		if err != nil {
		    return
		}
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
	return
}

func CreateAccessKey(flags *AccessKeyFlags, org string) (err error) {
	data := BuildAccessKeyJson(flags, true)
	url := GetAccessKeysPath(flags.BintrayDetails, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, err := ioutils.SendPost(url, []byte(data), httpClientsDetails)
	if err != nil {
	    return
	}
	if resp.StatusCode != 201 {
		err = cliutils.CheckError(errors.New(resp.Status + ". "+utils.ReadBintrayMessage(body)))
		if err != nil {
		    return
		}
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
	return
}

func UpdateAccessKey(flags *AccessKeyFlags, org string) error {
	data := BuildAccessKeyJson(flags, false)
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, err := ioutils.SendPatch(url, []byte(data), httpClientsDetails)
	if err != nil {
	    return err
	}
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". "+utils.ReadBintrayMessage(body)))
        if err != nil {
            return err
        }
	}
	fmt.Println("Bintray response: " + resp.Status)
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
	resp, body, err := ioutils.SendDelete(url, nil, httpClientsDetails)
	if err != nil {
	    return err
	}
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". "+utils.ReadBintrayMessage(body)))
        if err != nil {
            return err
        }
	}
	fmt.Println("Bintray response: " + resp.Status)
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
