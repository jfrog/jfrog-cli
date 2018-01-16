package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray/services"
)

func DownloadFile(config bintray.Config, params *services.DownloadFileParams) (totalDownloded, totalFailed int, err error) {
	bt, err := bintray.New(config)
	if err != nil {
		return
	}
	return bt.DownloadFile(params)
}

func DownloadVersion(config bintray.Config, params *services.DownloadVersionParams) (totalDownloded, totalFailed int, err error) {
	bt, err := bintray.New(config)
	if err != nil {
		return
	}
	totalDownloded, totalFailed, err = bt.DownloadVersion(params)
	return
}
