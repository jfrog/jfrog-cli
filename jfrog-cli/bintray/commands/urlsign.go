package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
)

func SignVersion(urlSigningDetails *utils.PathDetails, flags *UrlSigningFlags) error {
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = urlSigningDetails.Subject
	}
	path := urlSigningDetails.Subject + "/" + urlSigningDetails.Repo + "/" + urlSigningDetails.Path
	url := flags.BintrayDetails.ApiUrl + "signed_url/" + path
	data := builJson(flags)

	log.Info("Signing URL...")
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, err := httputils.SendPost(url, []byte(data), httpClientsDetails)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
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
