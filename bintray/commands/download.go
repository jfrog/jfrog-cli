package commands

import (
	"github.com/jfrog/jfrog-cli-go/bintray/utils"
	"github.com/jfrog/jfrog-client-go/bintray"
	"github.com/jfrog/jfrog-client-go/bintray/services"
)

func DownloadFile(config bintray.Config, params *services.DownloadFileParams) (totalDownloaded, totalFailed int, err error) {
	return utils.DownloadFileFromBintray(config, params)
}

func DownloadVersion(config bintray.Config, params *services.DownloadVersionParams) (totalDownloaded, totalFailed int, err error) {
	bt, err := bintray.New(config)
	if err != nil {
		return
	}
	totalDownloaded, totalFailed, err = bt.DownloadVersion(params)
	return
}
