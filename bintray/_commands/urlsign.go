package commands

import (
    "fmt"
    "github.com/jFrogdev/jfrog-cli-go/cliutils"
    "github.com/jFrogdev/jfrog-cli-go/bintray/utils"
)

func SignVersion(urlSigningDetails *utils.PathDetails, flags *UrlSigningFlags) {
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = urlSigningDetails.Subject
    }
    path := urlSigningDetails.Subject + "/" + urlSigningDetails.Repo + "/" + urlSigningDetails.Path
    url := flags.BintrayDetails.ApiUrl + "signed_url/" + path
    data := builJson(flags)

    fmt.Println("Signing URL for: " + path)
    resp, body := cliutils.SendPost(url, nil, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(cliutils.IndentJson(body))
}

func builJson(flags *UrlSigningFlags) string {
    m := map[string]string {
       "expiry": flags.Expiry,
       "valid_for_secs": flags.ValidFor,
       "callback_id": flags.CallbackId,
       "callback_email": flags.CallbackEmail,
       "callback_url": flags.CallbackUrl,
       "callback_method": flags.CallbackMethod,
    }
    return cliutils.MapToJson(m)
}

type UrlSigningFlags struct {
    BintrayDetails *utils.BintrayDetails
    Expiry string
    ValidFor string
    CallbackId string
    CallbackEmail string
    CallbackUrl string
    CallbackMethod string
}