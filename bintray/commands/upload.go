package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/bintray"
	"github.com/jfrog/jfrog-client-go/bintray/services"
	"github.com/jfrog/jfrog-client-go/bintray/services/packages"
	"github.com/jfrog/jfrog-client-go/bintray/services/repositories"
	"github.com/jfrog/jfrog-client-go/bintray/services/versions"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func Upload(config bintray.Config, uploadDetails *services.UploadParams) (uploaded int, failed int, err error) {
	var sm *bintray.ServicesManager
	sm, err = bintray.New(config)
	if err != nil {
		return
	}

	if !config.IsDryRun() {
		var exists bool
		log.Info("Verifying repository", uploadDetails.Repo, "exists...")
		exists, err = sm.IsRepoExists(&repositories.Path{Subject: uploadDetails.Subject, Repo: uploadDetails.Repo})
		if err != nil {
			return
		}
		if !exists {
			promptRepoNotExist(uploadDetails.Params)
		}

		log.Info("Verifying package", uploadDetails.Package, "exists...")
		packagePath := &packages.Path{Subject: uploadDetails.Subject, Repo: uploadDetails.Repo, Package: uploadDetails.Repo}
		exists, err = sm.IsPackageExists(packagePath)
		if err != nil {
			return
		}
		if !exists {
			promptPackageNotExist(uploadDetails.Path)
		}

		exists, err = sm.IsVersionExists(uploadDetails.Path)
		if err != nil {
			return
		}
		if !exists {
			if err = sm.CreateVersion(uploadDetails.Params); err != nil {
				return
			}
		}
	}

	return sm.UploadFiles(uploadDetails)
}

func promptRepoNotExist(versionDetails *versions.Params) error {
	msg := "It looks like repository '" + versionDetails.Repo + "' does not exist.\n"
	return errorutils.CheckError(errors.New(msg))
}

func promptPackageNotExist(versionDetails *versions.Path) error {
	msg := "It looks like package '" + versionDetails.Package +
		"' does not exist in the '" + versionDetails.Repo + "' repository.\n" +
		"You can create the package by running the package-create command. For example:\n" +
		"jfrog bt pc " +
		versionDetails.Subject + "/" + versionDetails.Repo + "/" + versionDetails.Package +
		" --vcs-url=https://github.com/example"

	conf, err := config.ReadBintrayConf()
	if err != nil {
		return err
	}
	if conf.DefPackageLicense == "" {
		msg += " --licenses=Apache-2.0-example"
	}
	err = errorutils.CheckError(errors.New(msg))
	return err
}
