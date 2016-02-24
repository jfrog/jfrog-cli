package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

func Config(details *cliutils.BintrayDetails, interactive bool) {
	if interactive {
		savedDetails := cliutils.ReadBintrayConf()

		if details.User == "" {
			cliutils.ScanFromConsole("User", &details.User, savedDetails.User)
		}
		if details.Key == "" {
			print("Key: ")
			byteKey, err := terminal.ReadPassword(int(syscall.Stdin))
			cliutils.CheckError(err)
			details.Key = string(byteKey)
			if details.Key == "" {
				details.Key = savedDetails.Key
			}
		}
		if details.DefPackageLicenses == "" {
			cliutils.ScanFromConsole("\nDefault package licenses",
			    &details.DefPackageLicenses, savedDetails.DefPackageLicenses)
		}
	}
	cliutils.SaveBintrayConf(details)
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