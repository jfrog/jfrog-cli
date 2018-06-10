package commands

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/packages"
)

func CreatePackage(config bintray.Config, params *packages.Params) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.CreatePackage(params)
}

func UpdatePackage(config bintray.Config, params *packages.Params) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.UpdatePackage(params)
}

func ShowPackage(config bintray.Config, params *packages.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.ShowPackage(params)
}

func DeletePackage(config bintray.Config, params *packages.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.DeletePackage(params)
}
