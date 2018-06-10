package commands

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/versions"
)

func GpgSignFile(config bintray.Config, pathDetails *utils.PathDetails, passphrase string) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.GpgSignFile(pathDetails, passphrase)
}

func GpgSignVersion(config bintray.Config, versionPath *versions.Path, passphrase string) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.GpgSignVersion(versionPath, passphrase)
}
