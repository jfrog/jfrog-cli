package tests

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
)

func CreateBintrayDetails() *config.BintrayDetails {
	return &config.BintrayDetails{
		ApiUrl:             "https://api.bintray.com/",
		DownloadServerUrl:  "https://dl.bintray.com/",
		User:               "user",
		Key:                "api-key",
		DefPackageLicenses: "Apache-2.0"}
}