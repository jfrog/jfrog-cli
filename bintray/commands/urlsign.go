package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func SignVersion(urlSigningDetails *utils.PathDetails, flags *UrlSigningFlags) error {
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = urlSigningDetails.Subject
	}
	path := urlSigningDetails.Subject + "/" + urlSigningDetails.Repo + "/" + urlSigningDetails.Path
	url := flags.BintrayDetails.ApiUrl + "signed_url/" + path
	data := builJson(flags)

	logger.Logger.Info("Signing URL for: " + path)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, err := ioutils.SendPost(url, []byte(data), httpClientsDetails)
	if err != nil {
	    return err
	}
	logger.Logger.Info("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
	return nil
}

func builJson(flags *UrlSigningFlags) string {
	m := map[string]string{
		"expiry":          flags.Expiry,
		"valid_for_secs":  flags.ValidFor,
		"callback_id":     flags.CallbackId,
		"callback_email":  flags.CallbackEmail,
		"callback_url":    flags.CallbackUrl,
		"callback_method": flags.CallbackMethod,
	}
	return cliutils.MapToJson(m)
}

type UrlSigningFlags struct {
	BintrayDetails *config.BintrayDetails
	Expiry         string
	ValidFor       string
	CallbackId     string
	CallbackEmail  string
	CallbackUrl    string
	CallbackMethod string
}
