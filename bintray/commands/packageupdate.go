package commands

import (
	"fmt"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func UpdatePackage(packageDetails *utils.VersionDetails, flags *utils.PackageFlags) error {
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = packageDetails.Subject
	}
	data := utils.CreatePackageJson(packageDetails.Package, flags)
	url := flags.BintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package

	log.Info("Updating package:", packageDetails.Package)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, err := ioutils.SendPatch(url, []byte(data), httpClientsDetails)
	if err != nil {
	    return err
	}
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". "+utils.ReadBintrayMessage(body)))
        if err != nil {
            return err
        }
	}
	log.Info("Bintray response:", resp.Status)
	fmt.Println(cliutils.IndentJson(body))
	return nil
}
