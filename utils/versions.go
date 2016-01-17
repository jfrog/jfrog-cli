package utils

import (
 	"net/http"
)

func HeadVersion(versionDetails *VersionDetails, bintrayDetails *BintrayDetails) *http.Response {
    url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

    return SendHead(url, bintrayDetails.User, bintrayDetails.Key)
}

func CreateVersionJson(versionName string, flags *VersionFlags) string {
    json := "{" +
        "\"name\": \"" + versionName + "\""
        if flags.Desc != "" {
            json += "," + "\"desc\": \"" + flags.Desc + "\""
        }
        if flags.GithubReleaseNotesFile != "" {
            json += "," + "\"github_release_notes_file\": \"" + flags.GithubReleaseNotesFile + "\""
        }
        if flags.VcsTag != "" {
            json += "," + "\"vcs_tag\": \"" + flags.VcsTag + "\""
        }
        if flags.Released != "" {
            json += "," + "\"released\": \"" + flags.Released + "\""
        }
        if flags.GithubUseTagReleaseNotes != "" {
            json += "," + "\"github_use_tag_release_notes\": " + flags.GithubUseTagReleaseNotes
        }
        json +=
    "}"

    return json
}

type VersionFlags struct {
    BintrayDetails *BintrayDetails
    Desc string
    VcsTag string
    Released string
    GithubReleaseNotesFile string
    GithubUseTagReleaseNotes string
}