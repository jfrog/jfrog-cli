package entitlements

import (
    "errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
)

func BuildEntitlementsUrl(bintrayDetails *config.BintrayDetails, details *utils.VersionDetails) string {
	return bintrayDetails.ApiUrl + createBintrayPath(details) + "/entitlements"
}

func BuildEntitlementUrl(bintrayDetails *config.BintrayDetails, details *utils.VersionDetails, entId string) string {
	return BuildEntitlementsUrl(bintrayDetails, details) + "/" + entId
}

func CreateVersionDetails(versionStr string) (*utils.VersionDetails, error) {
	parts := strings.Split(versionStr, "/")
	if len(parts) == 1 {
		err := errorutils.CheckError(errors.New("Argument format should be subject/repository or subject/repository/package or subject/repository/package/version. Got " + versionStr))
        if err != nil {
            return nil, err
        }
	}
	return utils.CreateVersionDetails(versionStr)
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

func buildEntitlementJson(flags *EntitlementFlags, create bool) string {
	m := map[string]string{
		"access":        flags.Access,
		"download_keys": cliutils.BuildListString(flags.Keys),
		"path":          flags.Path,
	}
	return cliutils.MapToJson(m)
}

type EntitlementFlags struct {
	BintrayDetails *config.BintrayDetails
	Id             string
	Path           string
	Access         string
	Keys           string
}
