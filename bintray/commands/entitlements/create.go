package entitlements

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"fmt"
)

func CreateEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) (err error) {
	var path = BuildEntitlementsUrl(flags.BintrayDetails, details)

	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = details.Subject
	}
	data := buildEntitlementJson(flags, true)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	log.Info("Creating entitlement...")
	resp, body, err := ioutils.SendPost(path, []byte(data), httpClientsDetails)
	if err != nil {
		return
	}
	if resp.StatusCode != 201 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Created entitlement, details:")
	fmt.Println(cliutils.IndentJson(body))
	return
}

