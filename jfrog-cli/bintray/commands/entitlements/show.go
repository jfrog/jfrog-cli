package entitlements

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	clientuitls "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
)

func ShowEntitlements(bintrayDetails *config.BintrayDetails, details *utils.VersionDetails) error {
	url := BuildEntitlementsUrl(bintrayDetails, details)
	if bintrayDetails.User == "" {
		bintrayDetails.User = details.Subject
	}
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	log.Info("Getting entitlements...")
	resp, body, _, err := httputils.SendGet(url, true, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientuitls.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Entitlements details:")
	fmt.Println(clientuitls.IndentJson(body))
	return nil
}

func ShowEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) error {
	url := BuildEntitlementUrl(flags.BintrayDetails, details, flags.Id)
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = details.Subject
	}
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Getting entitlement...")
	resp, body, _, err := httputils.SendGet(url, true, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientuitls.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Entitlement details:")
	fmt.Println(clientuitls.IndentJson(body))
	return nil
}

