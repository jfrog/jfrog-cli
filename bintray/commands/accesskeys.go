package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
	"strconv"
)

func ShowAccessKeys(bintrayDetails *cliutils.BintrayDetails, org string) {
	path := getAccessKeysPath(bintrayDetails, org)
	resp, body, _, _ := cliutils.SendGet(path, nil, true, bintrayDetails.User, bintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func ShowAccessKey(flags *AccessKeyFlags, org string) {
	url := getAccessKeysPath(flags.BintrayDetails, org)
	url += "/" + flags.Id
	resp, body, _, _ := cliutils.SendGet(url, nil, true, flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func CreateAccessKey(flags *AccessKeyFlags, org string) {
	data := buildAccessKeyJson(flags, true)
	url := getAccessKeysPath(flags.BintrayDetails, org)
	resp, body := cliutils.SendPost(url, nil, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 201 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func UpdateAccessKey(flags *AccessKeyFlags, org string) {
	data := buildAccessKeyJson(flags, false)
	url := getAccessKeysPath(flags.BintrayDetails, org)
	url += "/" + flags.Id
	resp, body := cliutils.SendPatch(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func buildAccessKeyJson(flags *AccessKeyFlags, create bool) string {
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
	url := getAccessKeysPath(flags.BintrayDetails, org)
	url += "/" + flags.Id
	resp, body := cliutils.SendDelete(url, flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
}

func getAccessKeysPath(bintrayDetails *cliutils.BintrayDetails, org string) string {
	if org == "" {
		return bintrayDetails.ApiUrl + "users/" + bintrayDetails.User + "/download_keys"
	}
	return bintrayDetails.ApiUrl + "orgs/" + org + "/download_keys"
}

type AccessKeyFlags struct {
	BintrayDetails      *cliutils.BintrayDetails
	Id                  string
	Password            string
	Expiry              string
	ExistenceCheckUrl   string
	ExistenceCheckCache int
	WhiteCidrs          string
	BlackCidrs          string
}
