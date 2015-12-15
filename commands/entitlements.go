package commands

import (
    "strings"
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

func CreateVersionDetailsForEntitlements(versionStr string) *utils.VersionDetails {
    parts := strings.Split(versionStr, "/")
    if len(parts) == 1 {
        utils.Exit("Argument format should be subject/repository or subject/repository/package or subject/repository/package/version. Got " + versionStr)
    }
    return utils.CreateVersionDetails(versionStr)
}

func buildEntitlementJson(flags *EntitlementFlags, create bool) string {
    data := "{" +
        "\"access\": \"" + flags.Access + "\"," +
        "\"download_keys\": " + fixArgList(flags.Keys) + "," +
        "\"path\": \"" + flags.Path + "\"" +
        "}"

    return data
}

type EntitlementFlags struct {
    BintrayDetails *utils.BintrayDetails
    Id string
    Path string
    Access string
    Keys string
}
