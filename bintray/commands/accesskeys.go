package commands

import (
	"github.com/jfrog/jfrog-client-go/bintray"
	"github.com/jfrog/jfrog-client-go/bintray/services/accesskeys"
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
