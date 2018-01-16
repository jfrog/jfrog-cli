package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray/services/accesskeys"
)

func ShowAllAccessKeys(config bintray.Config, org string) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.ShowAllAccessKeys(org)
}

func ShowAccessKey(config bintray.Config, org, id string) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.ShowAccessKey(org, id)
}

func CreateAccessKey(config bintray.Config, params *accesskeys.Params) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.CreateAccessKey(params)
}

func UpdateAccessKey(config bintray.Config, params *accesskeys.Params) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.UpdateAccessKey(params)
}

func DeleteAccessKey(config bintray.Config, org, id string) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.DeleteAccessKey(org, id)
}
