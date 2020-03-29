package utils

import (
	"github.com/jfrog/jfrog-client-go/bintray"
	"github.com/jfrog/jfrog-client-go/bintray/services"
)

func DownloadFileFromBintray(config bintray.Config, params *services.DownloadFileParams) (totalDownloaded, totalFailed int, err error) {
	bt, err := bintray.New(config)
	if err != nil {
		return
	}
	return bt.DownloadFile(params)
}
