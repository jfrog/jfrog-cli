package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"net/http"
)

func HeadPackage(packageDetails *VersionDetails, bintrayDetails *config.BintrayDetails) (resp *http.Response, body []byte, err error) {
	url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package
	httpClientsDetails := GetBintrayHttpClientDetails(bintrayDetails)

	resp, body, err = httputils.SendHead(url, httpClientsDetails)
	return
}

func HeadRepo(packageDetails *VersionDetails, bintrayDetails *config.BintrayDetails) (resp *http.Response, body []byte, err error) {
	url := bintrayDetails.ApiUrl + "repos/" + packageDetails.Subject + "/" +
		packageDetails.Repo
	httpClientsDetails := GetBintrayHttpClientDetails(bintrayDetails)

	resp, body, err = httputils.SendHead(url, httpClientsDetails)
	return
}

func CreatePackageJson(packageName string, flags *PackageFlags) string {
	m := map[string]string{
		"name":                      packageName,
		"desc":                      flags.Desc,
		"labels":                    cliutils.BuildListString(flags.Labels),
		"licenses":                  cliutils.BuildListString(flags.Licenses),
		"custom_licenses":           cliutils.BuildListString(flags.CustomLicenses),
		"vcs_url":                   flags.VcsUrl,
		"website_url":               flags.WebsiteUrl,
		"issue_tracker_url":         flags.IssueTrackerUrl,
		"github_repo":               flags.GithubRepo,
		"github_release_notes_file": flags.GithubReleaseNotesFile,
		"public_download_numbers":   flags.PublicDownloadNumbers,
		"public_stats":              flags.PublicStats,
	}

	return cliutils.MapToJson(m)
}

type PackageFlags struct {
	BintrayDetails         *config.BintrayDetails
	Desc                   string
	Labels                 string
	Licenses               string
	CustomLicenses         string
	VcsUrl                 string
	WebsiteUrl             string
	IssueTrackerUrl        string
	GithubRepo             string
	GithubReleaseNotesFile string
	PublicDownloadNumbers  string
	PublicStats            string
}
