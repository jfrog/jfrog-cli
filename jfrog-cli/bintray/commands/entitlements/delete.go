package entitlements

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	clientuitls "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
)

func DeleteEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) error {
	url := BuildEntitlementUrl(flags.BintrayDetails, details, flags.Id)
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = details.Subject
	}

	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Deleting entitlement...")
	resp, body, err := httputils.SendDelete(url, nil, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientuitls.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientuitls.IndentJson(body))
	return nil
}

