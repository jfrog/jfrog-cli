package commands

import (
	"github.com/jfrog/jfrog-client-go/bintray"
	"github.com/jfrog/jfrog-client-go/bintray/services/versions"
)

func LogsList(config bintray.Config, versionPath *versions.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.LogsList(versionPath)
}

func DownloadLog(config bintray.Config, versionPath *versions.Path, logName string) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.DownloadLog(versionPath, logName)
}
