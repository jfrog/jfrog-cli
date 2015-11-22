package commands

import (
    "strings"
    "strconv"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func ShowEntitlements(bintrayDetails *utils.BintrayDetails, details *utils.VersionDetails) {
    var path = bintrayDetails.ApiUrl + utils.CreateBintrayPath(details)
    if bintrayDetails.User == "" {
        bintrayDetails.User = details.Subject
    }
    resp, body := utils.SendGet(path, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    println("Bintray response: " + resp.Status)
    println(utils.IndentJson(body))
}

func CreateEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) {
    var path = flags.BintrayDetails.ApiUrl + utils.CreateBintrayPath(details)
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = details.Subject
    }
    data := buildEntitlementJson(flags, true)
    resp, body := utils.SendPost(path, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 201 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    println("Bintray response: " + resp.Status)
    println(utils.IndentJson(body))
}

func ShowDownloadKeys(bintrayDetails *utils.BintrayDetails, org string) {
    path := getDownloadKeysPath(bintrayDetails, org)
    resp, body := utils.SendGet(path, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    println("Bintray response: " + resp.Status)
    println(utils.IndentJson(body))
}

func ShowDownloadKey(flags *DownloadKeyFlags, org string) {
    url := getDownloadKeysPath(flags.BintrayDetails, org)
    url += "/" + flags.Id
    resp, body := utils.SendGet(url, nil, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    println("Bintray response: " + resp.Status)
    println(utils.IndentJson(body))
}

func CreateDownloadKey(flags *DownloadKeyFlags, org string) {
    data := buildDownloadKeyJson(flags, true)
    url := getDownloadKeysPath(flags.BintrayDetails, org)
    resp, body := utils.SendPost(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 201 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    println("Bintray response: " + resp.Status)
    println(utils.IndentJson(body))
}

func UpdateDownloadKey(flags *DownloadKeyFlags, org string) {
    data := buildDownloadKeyJson(flags, false)
    url := getDownloadKeysPath(flags.BintrayDetails, org)
    url += "/" + flags.Id
    resp, body := utils.SendPatch(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    println("Bintray response: " + resp.Status)
    println(utils.IndentJson(body))
}

func buildEntitlementJson(flags *EntitlementFlags, create bool) string {
    data := "{" +
        "\"access\": \"" + flags.Access + "\"," +
        "\"download_keys\": " + fixArgList(flags.Keys) + "," +
        "\"path\": \"" + flags.Path + "\"" +
        "}"

    return data
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
        whiteCidrs = "\"white_cidrs\": " + fixArgList(flags.WhiteCidrs)
    }
    if flags.BlackCidrs != "" {
        blackCidrs = "\"black_cidrs\": " + fixArgList(flags.BlackCidrs)
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
    resp, body := utils.SendDelete(url, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    println("Bintray response: " + resp.Status)
}

func fixArgList(cidr string) string {
    split := strings.Split(cidr, ",")
    size := len(split)
    str := "[\""
    for i := 0; i < size; i++ {
        str += split[i]
        if i+1 < size {
            str += "\",\""
        }
    }
    str += "\"]"
    return str
}

func getDownloadKeysPath(bintrayDetails *utils.BintrayDetails, org string) string {
    if org == "" {
        return bintrayDetails.ApiUrl + "users/" + bintrayDetails.User + "/download_keys"
    }
    return bintrayDetails.ApiUrl + "orgs/" + org + "/download_keys"
}

type EntitlementFlags struct {
    BintrayDetails *utils.BintrayDetails
    Id string
    Path string
    Access string
    Keys string
}

type DownloadKeyFlags struct {
    BintrayDetails *utils.BintrayDetails
    Id string
    Expiry string
    ExistenceCheckUrl string
    ExistenceCheckCache int
    WhiteCidrs string
    BlackCidrs string
}