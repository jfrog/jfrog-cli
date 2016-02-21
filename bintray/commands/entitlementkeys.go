package commands

import (
	"fmt"
    "strconv"
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
    "github.com/jfrogdev/jfrog-cli-go/bintray/utils"
)

func ShowDownloadKeys(bintrayDetails *cliutils.BintrayDetails, org string) {
    path := getDownloadKeysPath(bintrayDetails, org)
    resp, body, _, _ := cliutils.SendGet(path, nil, true, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(cliutils.IndentJson(body))
}

func ShowDownloadKey(flags *DownloadKeyFlags, org string) {
    url := getDownloadKeysPath(flags.BintrayDetails, org)
    url += "/" + flags.Id
    resp, body, _, _ := cliutils.SendGet(url, nil, true, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(cliutils.IndentJson(body))
}

func CreateDownloadKey(flags *DownloadKeyFlags, org string) {
    data := buildDownloadKeyJson(flags, true)
    url := getDownloadKeysPath(flags.BintrayDetails, org)
    resp, body := cliutils.SendPost(url, nil, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 201 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(cliutils.IndentJson(body))
}

func UpdateDownloadKey(flags *DownloadKeyFlags, org string) {
    data := buildDownloadKeyJson(flags, false)
    url := getDownloadKeysPath(flags.BintrayDetails, org)
    url += "/" + flags.Id
    resp, body := cliutils.SendPatch(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(cliutils.IndentJson(body))
}

func buildDownloadKeyJson(flags *DownloadKeyFlags, create bool) string {
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

func DeleteDownloadKey(flags *DownloadKeyFlags, org string) {
    url := getDownloadKeysPath(flags.BintrayDetails, org)
    url += "/" + flags.Id
    resp, body := cliutils.SendDelete(url, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
}

func getDownloadKeysPath(bintrayDetails *cliutils.BintrayDetails, org string) string {
    if org == "" {
        return bintrayDetails.ApiUrl + "users/" + bintrayDetails.User + "/download_keys"
    }
    return bintrayDetails.ApiUrl + "orgs/" + org + "/download_keys"
}

type DownloadKeyFlags struct {
    BintrayDetails *cliutils.BintrayDetails
    Id string
    Expiry string
    ExistenceCheckUrl string
    ExistenceCheckCache int
    WhiteCidrs string
    BlackCidrs string
}