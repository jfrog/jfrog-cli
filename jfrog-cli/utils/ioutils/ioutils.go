package ioutils

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"os"
	"io"
)

func ReadCredentialsFromConsole(details, savedDetails cliutils.Credentials) error {
	if details.GetUser() == "" {
		tempUser := ""
		ScanFromConsole("User", &tempUser, savedDetails.GetUser())
		details.SetUser(tempUser)
	}
	if details.GetPassword() == "" {
		print("Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		err = errorutils.CheckError(err)
		if err != nil {
			return err
		}
		details.SetPassword(string(bytePassword))
		if details.GetPassword() == "" {
			details.SetPassword(savedDetails.GetPassword())
		}
	}
	return nil
}

func ScanFromConsole(caption string, scanInto *string, defaultValue string) {
	if defaultValue != "" {
		print(caption + " [" + defaultValue + "]: ")
	} else {
		print(caption + ": ")
	}
	fmt.Scanln(scanInto)
	if *scanInto == "" {
		*scanInto = defaultValue
	}
}

func CopyFile(src, dst string, fileMode os.FileMode) error {
	from, err := os.Open(src)
	if err != nil {
		return errorutils.CheckError(err)
	}
	defer from.Close()

	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, fileMode)
	if err != nil {
		return errorutils.CheckError(err)
	}
	defer to.Close()

	if _, err = io.Copy(to, from); err != nil {
		return errorutils.CheckError(err)
	}

	return errorutils.CheckError(os.Chmod(dst, fileMode))
}