package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
	"strings"
)

func BuildEntitlementsUrl(bintrayDetails *cliutils.BintrayDetails, details *utils.VersionDetails) string {
    return bintrayDetails.ApiUrl + createBintrayPath(details) + "/entitlements"
}

func BuildEntitlementUrl(bintrayDetails *cliutils.BintrayDetails, details *utils.VersionDetails, entId string) string {
	return BuildEntitlementsUrl(bintrayDetails, details) + "/" + entId
}

func ShowEntitlements(bintrayDetails *cliutils.BintrayDetails, details *utils.VersionDetails) {
	url := BuildEntitlementsUrl(bintrayDetails, details)
	if bintrayDetails.User == "" {
		bintrayDetails.User = details.Subject
	}
	resp, body, _, _ := cliutils.SendGet(url, nil, true, bintrayDetails.User, bintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func ShowEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) {
	url := BuildEntitlementUrl(flags.BintrayDetails, details, flags.Id)
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = details.Subject
	}
	resp, body, _, _ := cliutils.SendGet(url, nil, true, flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func DeleteEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) {
    url := BuildEntitlementUrl(flags.BintrayDetails, details, flags.Id)
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = details.Subject
	}
	resp, body := cliutils.SendDelete(url, flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
}

func CreateEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) {
	var path = BuildEntitlementsUrl(flags.BintrayDetails, details)

	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = details.Subject
	}
	data := buildEntitlementJson(flags, true)
	resp, body := cliutils.SendPost(path, nil, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 201 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func UpdateEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) {
    path := BuildEntitlementUrl(flags.BintrayDetails, details, flags.Id)
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = details.Subject
	}
	data := buildEntitlementJson(flags, true)
	resp, body := cliutils.SendPatch(path, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func CreateVersionDetailsForEntitlements(versionStr string) *utils.VersionDetails {
	parts := strings.Split(versionStr, "/")
	if len(parts) == 1 {
		cliutils.Exit(cliutils.ExitCodeError, "Argument format should be subject/repository or subject/repository/package or subject/repository/package/version. Got "+versionStr)
	}
	return utils.CreateVersionDetails(versionStr)
}

func buildEntitlementJson(flags *EntitlementFlags, create bool) string {
	m := map[string]string{
		"access":        flags.Access,
		"download_keys": cliutils.BuildListString(flags.Keys),
		"path":          flags.Path,
	}
	return cliutils.MapToJson(m)
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
	BintrayDetails *cliutils.BintrayDetails
	Id             string
	Path           string
	Access         string
	Keys           string
}
