package utils

import (
    "strconv"
)

func CreatePackageJson(packageName string, flags *PackageFlags) string {
    json := "{" +
        "\"name\": \"" + packageName + "\"," +
        "\"desc\": \"" + flags.Desc + "\"," +
        "\"labels\": " + BuildListString(flags.Labels) + "," +
        "\"licenses\": " + BuildListString(flags.Licenses) + "," +
        "\"custom_licenses\": \"" + BuildListString(flags.CustomLicenses) + "\"," +
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
    PublicDownloadNumbers bool
    PublicStats bool
}