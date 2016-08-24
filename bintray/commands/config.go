package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

func Config(details, defaultDetails *config.BintrayDetails, interactive bool) (*config.BintrayDetails, error) {
    if details == nil {
        details = new(config.BintrayDetails)
    }
	if interactive {
	    if defaultDetails == nil {
	        var err error
            defaultDetails, err = config.ReadBintrayConf()
			if err != nil {
			    return nil, err
			}
	    }
		if details.User == "" {
			ioutils.ScanFromConsole("User", &details.User, defaultDetails.User)
		}
		if details.Key == "" {
			print("Key: ")
			byteKey, err := terminal.ReadPassword(int(syscall.Stdin))
			err = cliutils.CheckError(err)
			if err != nil {
			    return nil, err
			}
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
	return details, nil
}

func ShowConfig() error {
	details, err := config.ReadBintrayConf()
	if err != nil {
		return err
	}
	if details.User != "" {
		fmt.Println("User: " + details.User)
	}
	if details.User != "" {
		fmt.Println("User: " + details.User)
	}
	if details.Key != "" {
		fmt.Println("Key: ***")
	}
	if details.DefPackageLicenses != "" {
		fmt.Println("Default package license: " + details.DefPackageLicenses)
	}
	return nil
}

func ClearConfig() {
	config.SaveBintrayConf(new(config.BintrayDetails))
}

func GetConfig() (*config.BintrayDetails, error) {
	return config.ReadBintrayConf()
}