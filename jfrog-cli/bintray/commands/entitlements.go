package commands

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/entitlements"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/versions"
)

func ShowAllEntitlements(config bintray.Config, path *versions.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.ShowAllEntitlements(path)
}

func ShowEntitlement(config bintray.Config, id string, path *versions.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.ShowEntitlement(id, path)
}

func CreateEntitlement(config bintray.Config, params *entitlements.Params) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.CreateEntitlement(params)
}

func UpdateEntitlement(config bintray.Config, params *entitlements.Params) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.UpdateEntitlement(params)
}

func DeleteEntitlement(config bintray.Config, id string, path *versions.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.DeleteEntitlement(id, path)
}
