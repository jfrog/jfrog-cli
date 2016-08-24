package entitlements

import (
    "errors"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func DeleteEntitlement(flags *EntitlementFlags, details *utils.VersionDetails) error {
	url := BuildEntitlementUrl(flags.BintrayDetails, details, flags.Id)
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = details.Subject
	}

	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, err := ioutils.SendDelete(url, nil, httpClientsDetails)
	if err != nil {
	    return err
	}
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". " + utils.ReadBintrayMessage(body)))
        if err != nil {
            return err
        }
	}
	fmt.Println("Bintray response: " + resp.Status)
	return nil
}

