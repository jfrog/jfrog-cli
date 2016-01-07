package commands

import (
    "fmt"
    "strconv"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func CreatePackage(packageDetails *utils.VersionDetails, flags *PackageFlags) {
    data := createPackageJson(packageDetails.Package, flags)
    url := flags.BintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
        packageDetails.Repo

println(data)

    fmt.Println("Creating package: " + packageDetails.Package)
    resp, body := utils.SendPost(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 201 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}

func createPackageJson(packageName string, flags *PackageFlags) string {
    json := "{" +
        "\"name\": \"" + packageName + "\"," +
        "\"desc\": \"" + flags.Desc + "\"," +
        "\"labels\": " + utils.BuildListString(flags.Labels) + "," +
        "\"licenses\": " + utils.BuildListString(flags.Licenses) + "," +
        "\"custom_licenses\": \"" + utils.BuildListString(flags.CustomLicenses) + "\"," +
        "\"vcs_url\": \"" + flags.VcsUrl + "\"," +
        "\"website_url\": \"" + flags.WebsiteUrl + "\"," +
        "\"issue_tracker_url\": \"" + flags.IssueTrackerUrl + "\"," +
        "\"github_repo\": \"" + flags.GithubRepo + "\"," +
        "\"github_release_notes_file\": \"" + flags.GithubReleaseNotesFile + "\"," +
        "\"public_download_numbers\": " + strconv.FormatBool(flags.PublicDownloadNumbers) + "," +
        "\"public_stats\": " + strconv.FormatBool(flags.PublicStats) +
    "}"
    return json
}

type PackageFlags struct {
    BintrayDetails *utils.BintrayDetails
    Desc string
    Labels string
    Licenses string
    CustomLicenses string
    VcsUrl string
    WebsiteUrl string
    IssueTrackerUrl string
    GithubRepo string
    GithubReleaseNotesFile string
    PublicDownloadNumbers bool
    PublicStats bool
}