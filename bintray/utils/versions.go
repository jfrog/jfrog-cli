package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"net/http"
)

func HeadVersion(versionDetails *VersionDetails, bintrayDetails *config.BintrayDetails) (*http.Response, error) {
	url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version
	httpClientsDetails := GetBintrayHttpClientDetails(bintrayDetails)
	return httputils.SendHead(url, httpClientsDetails)
}

func CreateVersionJson(versionName string, flags *VersionFlags) string {
	m := map[string]string{
		"name": versionName,
		"desc": flags.Desc,
		"github_release_notes_file":    flags.GithubReleaseNotesFile,
		"VcsTag":                       flags.VcsTag,
		"released":                     flags.Released,
		"github_use_tag_release_notes": flags.GithubUseTagReleaseNotes,
	}
	return cliutils.MapToJson(m)
}

type VersionFlags struct {
	BintrayDetails           *config.BintrayDetails
	Desc                     string
	VcsTag                   string
	Released                 string
	GithubReleaseNotesFile   string
	GithubUseTagReleaseNotes string
}
