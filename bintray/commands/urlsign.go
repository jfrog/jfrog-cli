package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"errors"
	"fmt"
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
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Signed URL", path + ", details:")
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
