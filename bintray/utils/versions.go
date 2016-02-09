package utils

import (
 	"net/http"
 	"github.com/jFrogdev/jfrog-cli-go/cliutils"
)

func HeadVersion(versionDetails *VersionDetails, bintrayDetails *BintrayDetails) *http.Response {
    url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

    return cliutils.SendHead(url, bintrayDetails.User, bintrayDetails.Key)
}

func CreateVersionJson(versionName string, flags *VersionFlags) string {
    m := map[string]string {
       "name": versionName,
       "desc": flags.Desc,
       "github_release_notes_file": flags.GithubReleaseNotesFile,
       "VcsTag": flags.VcsTag,
       "released": flags.Released,
       "github_use_tag_release_notes": flags.GithubUseTagReleaseNotes,
    }
    return cliutils.MapToJson(m)
}

type VersionFlags struct {
    BintrayDetails *BintrayDetails
    Desc string
    VcsTag string
    Released string
    GithubReleaseNotesFile string
    GithubUseTagReleaseNotes string
}