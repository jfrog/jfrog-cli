package entitlements

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func ShowEntitlements(bintrayDetails *config.BintrayDetails, details *utils.VersionDetails) error {
	url := BuildEntitlementsUrl(bintrayDetails, details)
	if bintrayDetails.User == "" {
		bintrayDetails.User = details.Subject
	}
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	log.Info("Getting entitlements...")
	resp, body, _, err := ioutils.SendGet(url, true, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Entitlements details:", "\n" + cliutils.IndentJson(body))
	return nil
}

func ShowEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) error {
	url := BuildEntitlementUrl(flags.BintrayDetails, details, flags.Id)
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = details.Subject
	}
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Getting entitlement...")
	resp, body, _, err := ioutils.SendGet(url, true, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Entitlement details:", "\n" + cliutils.IndentJson(body))
	return nil
}

