package commands

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/versions"
)

func CreateVersion(config bintray.Config, params *versions.Params) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.CreateVersion(params)
}

func UpdateVersion(config bintray.Config, params *versions.Params) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.UpdateVersion(params)
}

func PublishVersion(config bintray.Config, params *versions.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.PublishVersion(params)
}

func ShowVersion(config bintray.Config, params *versions.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.ShowVersion(params)
}

func DeleteVersion(config bintray.Config, params *versions.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.DeleteVersion(params)
}
