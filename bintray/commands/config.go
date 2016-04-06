package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/Godeps/_workspace/src/golang.org/x/crypto/ssh/terminal"
	"syscall"
)

func Config(details, defaultDetails *config.BintrayDetails, interactive bool) *config.BintrayDetails {
    if details == nil {
        details = new(config.BintrayDetails)
    }
	if interactive {
	    if defaultDetails == nil {
            defaultDetails = config.ReadBintrayConf()
	    }
		if details.User == "" {
			ioutils.ScanFromConsole("User", &details.User, defaultDetails.User)
		}
		if details.Key == "" {
			print("Key: ")
			byteKey, err := terminal.ReadPassword(int(syscall.Stdin))
			cliutils.CheckError(err)
			details.Key = string(byteKey)
			if details.Key == "" {
				details.Key = defaultDetails.Key
			}
		}
		if details.DefPackageLicenses == "" {
			ioutils.ScanFromConsole("\nDefault package licenses",
			    &details.DefPackageLicenses, defaultDetails.DefPackageLicenses)
		}
	}
	config.SaveBintrayConf(details)
	return details
}

func ShowConfig() {
	details := config.ReadBintrayConf()
	if details.User != "" {
		fmt.Println("User: " + details.User)
	}
	if details.Key != "" {
		fmt.Println("Key: ***")
	}
	if details.DefPackageLicenses != "" {
		fmt.Println("Default package license: " + details.DefPackageLicenses)
	}
}

func ClearConfig() {
	config.SaveBintrayConf(new(config.BintrayDetails))
}

func GetConfig() *config.BintrayDetails {
	return config.ReadBintrayConf()
}