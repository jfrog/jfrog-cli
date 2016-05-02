package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"strconv"
)

func ShowAccessKeys(bintrayDetails *config.BintrayDetails, org string) {
	path := GetAccessKeysPath(bintrayDetails, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, _, _ := ioutils.SendGet(path, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func ShowAccessKey(flags *AccessKeyFlags, org string) {
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, _, _ := ioutils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func CreateAccessKey(flags *AccessKeyFlags, org string) {
	data := BuildAccessKeyJson(flags, true)
	url := GetAccessKeysPath(flags.BintrayDetails, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body := ioutils.SendPost(url, []byte(data), httpClientsDetails)
	if resp.StatusCode != 201 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func UpdateAccessKey(flags *AccessKeyFlags, org string) {
	data := BuildAccessKeyJson(flags, false)
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body := ioutils.SendPatch(url, []byte(data), httpClientsDetails)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
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

func DeleteAccessKey(flags *AccessKeyFlags, org string) {
	url := GetAccessKeyPath(flags.BintrayDetails, flags.Id, org)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body := ioutils.SendDelete(url, nil, httpClientsDetails)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
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
