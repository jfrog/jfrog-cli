package commands

import (
	"fmt"
	"github.com/JFrogDev/jfrog-cli-go/cliutils"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

func Config(details, defaultDetails *cliutils.BintrayDetails, interactive bool) *cliutils.BintrayDetails {
    if details == nil {
        details = new(cliutils.BintrayDetails)
    }
	if interactive {
	    if defaultDetails == nil {
            defaultDetails = cliutils.ReadBintrayConf()
	    }
		if details.User == "" {
			cliutils.ScanFromConsole("User", &details.User, defaultDetails.User)
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
			cliutils.ScanFromConsole("\nDefault package licenses",
			    &details.DefPackageLicenses, defaultDetails.DefPackageLicenses)
		}
	}
	cliutils.SaveBintrayConf(details)
	return details
}

func ShowConfig() {
	details := cliutils.ReadBintrayConf()
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
	cliutils.SaveBintrayConf(new(cliutils.BintrayDetails))
}

func GetConfig() *cliutils.BintrayDetails {
	return cliutils.ReadBintrayConf()
}