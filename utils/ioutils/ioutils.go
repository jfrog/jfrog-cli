package ioutils

import (
	"bufio"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"strings"
	"syscall"
)

// @param allowUsingSavedPassword - Prevent changing username or url without changing the password.
// False iff the user changed the username or the url.
func ReadCredentialsFromConsole(details, savedDetails cliutils.Credentials, allowUsingSavedPassword bool) error {
	if details.GetUser() == "" {
		tempUser := ""
		ScanFromConsole("User", &tempUser, savedDetails.GetUser())
		details.SetUser(tempUser)
		allowUsingSavedPassword = false
	}
	if details.GetPassword() == "" {
		print("Password/API key: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		err = errorutils.CheckError(err)
		if err != nil {
			return err
		}
		details.SetPassword(string(bytePassword))
		if details.GetPassword() == "" && allowUsingSavedPassword {
			details.SetPassword(savedDetails.GetPassword())
		}
		// New-line required after the password input:
		fmt.Println()
	}
	return nil
}

func ScanFromConsole(caption string, scanInto *string, defaultValue string) {
	if defaultValue != "" {
		print(caption + " [" + defaultValue + "]: ")
	} else {
		print(caption + ": ")
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	*scanInto = scanner.Text()
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

func DoubleWinPathSeparator(filePath string) string {
	return strings.Replace(filePath, "\\", "\\\\", -1)
}

func UnixToWinPathSeparator(filePath string) string {
	return strings.Replace(filePath, "/", "\\\\", -1)
}

func PrepareFilePathForUnix(path string) string {
	if cliutils.IsWindows() {
		path = strings.Replace(path, "\\\\", "/", -1)
		path = strings.Replace(path, "\\", "/", -1)
	}
	return path
}
