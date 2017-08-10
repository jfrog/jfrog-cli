package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"net/http"
)

func HeadVersion(versionDetails *VersionDetails, bintrayDetails *config.BintrayDetails) (resp *http.Response, body []byte, err error) {
	url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version
	httpClientsDetails := GetBintrayHttpClientDetails(bintrayDetails)

	resp, body, err = httputils.SendHead(url, httpClientsDetails)
	return
}

func CreateVersionJson(versionName string, flags *VersionFlags) string {
	m := map[string]string{
		"name": versionName,
		"desc": flags.Desc,
		"github_release_notes_file":    flags.GithubReleaseNotesFile,
		"vcs_tag":                      flags.VcsTag,
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
