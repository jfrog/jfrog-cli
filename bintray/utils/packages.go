package utils

import (
 	"net/http"
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
)

func HeadPackage(packageDetails *VersionDetails, bintrayDetails *BintrayDetails) *http.Response {
    url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
        packageDetails.Repo + "/" + packageDetails.Package

    return cliutils.SendHead(url, bintrayDetails.User, bintrayDetails.Key)
}

func CreatePackageJson(packageName string, flags *PackageFlags) string {
    m := map[string]string {
       "name": packageName,
       "desc": flags.Desc,
       "labels": cliutils.BuildListString(flags.Labels),
       "licenses": cliutils.BuildListString(flags.Licenses),
       "custom_licenses": cliutils.BuildListString(flags.CustomLicenses),
       "vcs_url": flags.VcsUrl,
       "website_url": flags.WebsiteUrl,
       "issue_tracker_url": flags.IssueTrackerUrl,
       "github_repo": flags.GithubRepo,
       "github_release_notes_file": flags.GithubReleaseNotesFile,
       "public_download_numbers": flags.PublicDownloadNumbers,
       "public_stats": flags.PublicStats,
    }

    return cliutils.MapToJson(m)
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