package commands

import (
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func ShowDownloadKeys(bintrayDetails *utils.BintrayDetails, org string) {
    path := getDownloadKeysPath(bintrayDetails, org)
    resp, body := utils.SendGet(path, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    println(string(body))
}

func CreateDownloadKeys(flags *DownloadKeyFlags, org string) {
    data := "{" +
        "\"id\": \"" + flags.Id + "\"," +
       "\"expiry\": \"" + flags.Expiry + "\"" +
    "}";

    url := getDownloadKeysPath(flags.BintrayDetails, org)
    resp, body := utils.SendPost(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }

    println(body)
}

func getDownloadKeysPath(bintrayDetails *utils.BintrayDetails, org string) string {
    if org == "" {
        return bintrayDetails.ApiUrl + "users/" + bintrayDetails.User + "/download_keys"
    }
    return bintrayDetails.ApiUrl + "orgs/" + org + "/download_keys"
}

type DownloadKeyFlags struct {
    BintrayDetails *utils.BintrayDetails
    Id string
    Expiry string
}