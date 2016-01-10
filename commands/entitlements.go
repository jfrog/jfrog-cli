package commands

import (
	"fmt"
    "strings"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func ShowEntitlements(bintrayDetails *utils.BintrayDetails, details *utils.VersionDetails) {
    url := bintrayDetails.ApiUrl + createBintrayPath(details) + "/entitlements"
    if bintrayDetails.User == "" {
        bintrayDetails.User = details.Subject
    }
    resp, body := utils.SendGet(url, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}

func ShowEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) {
    url := flags.BintrayDetails.ApiUrl + createBintrayPath(details) +
        "/entitlements/" + flags.Id
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = details.Subject
    }
    resp, body := utils.SendGet(url, nil, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}

func DeleteEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) {
    url := flags.BintrayDetails.ApiUrl + createBintrayPath(details) +
        "/entitlements/" + flags.Id
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = details.Subject
    }
    resp, body := utils.SendDelete(url, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
}

func CreateEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) {
    var path = flags.BintrayDetails.ApiUrl + createBintrayPath(details) + "/entitlements"
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = details.Subject
    }
    data := buildEntitlementJson(flags, true)
    resp, body := utils.SendPost(path, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 201 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}

func UpdateEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) {
    var path = flags.BintrayDetails.ApiUrl + createBintrayPath(details) +
        "/entitlements/" + flags.Id
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = details.Subject
    }
    data := buildEntitlementJson(flags, true)
    resp, body := utils.SendPatch(path, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
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
        "\"access\": \"" + flags.Access + "\""

    if flags.Keys != "" {
        data += ",\"download_keys\": " + utils.BuildListString(flags.Keys)
    }
    if flags.Path != "" {
        data += ",\"path\": \"" + flags.Path + "\""
    }
    data += "}"
    return data
}

func createBintrayPath(details *utils.VersionDetails) string {
    if details.Version == "" {
        if details.Package == "" {
            return "repos/" + details.Subject + "/" + details.Repo
        }
        return "packages/" + details.Subject + "/" + details.Repo + "/" + details.Package
    } else {
        return "packages/" + details.Subject + "/" + details.Repo + "/" + details.Package +
            "/versions/" + details.Version
    }
}

type EntitlementFlags struct {
    BintrayDetails *utils.BintrayDetails
    Id string
    Path string
    Access string
    Keys string
}
