package tests

import (
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
)

func CreateBintrayDetails() *cliutils.BintrayDetails {
	return &cliutils.BintrayDetails{
		ApiUrl:             "https://api.bintray.com/",
		DownloadServerUrl:  "https://dl.bintray.com/",
		User:               "user",
		Key:                "api-key",
		DefPackageLicenses: "Apache-2.0"}
}