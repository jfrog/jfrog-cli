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
			print("User: [" + savedDetails.User + "]: ")
			cliutils.ScanFromConsole(&details.User, savedDetails.User)
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
	}
	cliutils.SaveBintrayConf(details)
}

func ShowConfig() {
	details := cliutils.ReadBintrayConf()
	if details.User != "" {
		fmt.Println("User: " + details.User)
	}
	if details.Key != "" {
		fmt.Println("Key: " + details.Key)
	}
}

func ClearConfig() {
	cliutils.SaveBintrayConf(new(cliutils.BintrayDetails))
}

func GetConfig() *cliutils.BintrayDetails {
	return cliutils.ReadBintrayConf()
}
