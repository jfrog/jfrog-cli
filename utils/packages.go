package utils

import (
 	"net/http"
 )

func HeadPackage(packageDetails *VersionDetails, bintrayDetails *BintrayDetails) *http.Response {
    url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
        packageDetails.Repo + "/" + packageDetails.Package

    return SendHead(url, bintrayDetails.User, bintrayDetails.Key)
}

func CreatePackageJson(packageName string, flags *PackageFlags) string {
    json := "{" +
        "\"name\": \"" + packageName + "\""
        if flags.Desc != "" {
            json += "," + "\"desc\": \"" + flags.Desc + "\""
        }
        if flags.Labels != "" {
            json += "," + "\"labels\": " + BuildListString(flags.Labels)
        }
        if flags.Licenses != "" {
            json += "," + "\"licenses\": " + BuildListString(flags.Licenses)
        }
        if flags.CustomLicenses != "" {
            json += "," + "\"custom_licenses\": \"" + BuildListString(flags.CustomLicenses)
        }
        if flags.VcsUrl != "" {
            json += "," + "\"vcs_url\": \"" + flags.VcsUrl + "\""
        }
        if flags.WebsiteUrl != "" {
            json += "," + "\"website_url\": \"" + flags.WebsiteUrl + "\""
        }
        if flags.IssueTrackerUrl != "" {
            json += "," + "\"issue_tracker_url\": \"" + flags.IssueTrackerUrl + "\""
        }
        if flags.GithubRepo != "" {
            json += "," + "\"github_repo\": \"" + flags.GithubRepo + "\""
        }
        if flags.GithubReleaseNotesFile != "" {
            json += "," + "\"github_release_notes_file\": \"" + flags.GithubReleaseNotesFile + "\""
        }
        if flags.PublicDownloadNumbers != "" {
            json += "," + "\"public_download_numbers\": " + flags.PublicDownloadNumbers
        }
        if flags.PublicStats != "" {
            json += "," + "\"public_stats\": " + flags.PublicStats
        }
        json += "}"

    return json
}

type PackageFlags struct {
    BintrayDetails *BintrayDetails
    Desc string
    Labels string
    Licenses string
    CustomLicenses string
    VcsUrl string
    WebsiteUrl string
    IssueTrackerUrl string
    GithubRepo string
    GithubReleaseNotesFile string
    PublicDownloadNumbers string
    PublicStats string
}