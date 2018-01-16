package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray/services/versions"
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
