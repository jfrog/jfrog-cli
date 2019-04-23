package commands

import (
	"github.com/jfrog/jfrog-client-go/bintray"
	"github.com/jfrog/jfrog-client-go/bintray/services/url"
)

func SignVersion(config bintray.Config, params *url.Params) (err error) {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.SignUrl(params)
}
