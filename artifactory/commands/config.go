package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
	"golang.org/x/crypto/ssh/terminal"
	"strings"
	"syscall"
)

func Config(details, defaultDetails *cliutils.ArtifactoryDetails, interactive,
    shouldEncPassword bool) *cliutils.ArtifactoryDetails {

    if details == nil {
        details = new(cliutils.ArtifactoryDetails)
    }
	if interactive {
	    if defaultDetails == nil {
            defaultDetails = cliutils.ReadArtifactoryConf()
	    }
		if details.Url == "" {
			cliutils.ScanFromConsole("Artifactory URL", &details.Url, defaultDetails.Url)
		}
		if strings.Index(details.Url, "ssh://") == 0 || strings.Index(details.Url, "SSH://") == 0 {
			readSshKeyPathFromConsole(details, defaultDetails)
		} else {
			readCredentialsFromConsole(details, defaultDetails)
		}
	}
	details.Url = cliutils.AddTrailingSlashIfNeeded(details.Url)
	if shouldEncPassword {
		details = encryptPassword(details)
	}
	cliutils.SaveArtifactoryConf(details)
	return details
}

func readSshKeyPathFromConsole(details, savedDetails *cliutils.ArtifactoryDetails) {
	if details.SshKeyPath == "" {
		cliutils.ScanFromConsole("SSH key file path", &details.SshKeyPath, savedDetails.SshKeyPath)
	}
	if !cliutils.IsFileExists(details.SshKeyPath) {
		fmt.Println("Warning: Could not find SSH key file at: " + details.SshKeyPath)
	}
}

func readCredentialsFromConsole(details, savedDetails *cliutils.ArtifactoryDetails) {
	if details.User == "" {
		cliutils.ScanFromConsole("User", &details.User, savedDetails.User)
	}
	if details.Password == "" {
		print("Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		cliutils.CheckError(err)
		details.Password = string(bytePassword)
		if details.Password == "" {
			details.Password = savedDetails.Password
		}
	}
}

func ShowConfig() {
	details := cliutils.ReadArtifactoryConf()
	if details.Url != "" {
		fmt.Println("Url: " + details.Url)
	}
	if details.User != "" {
		fmt.Println("User: " + details.User)
	}
	if details.Password != "" {
		fmt.Println("Password: ***")
	}
	if details.SshKeyPath != "" {
		fmt.Println("SSH key file path: " + details.SshKeyPath)
	}
}

func ClearConfig() {
	cliutils.SaveArtifactoryConf(new(cliutils.ArtifactoryDetails))
}

func GetConfig() *cliutils.ArtifactoryDetails {
	return cliutils.ReadArtifactoryConf()
}

func encryptPassword(details *cliutils.ArtifactoryDetails) *cliutils.ArtifactoryDetails {
	if details.Password == "" {
		return details
	}
	fmt.Print("\nEncrypting password...")
	response, encPassword := utils.GetEncryptedPasswordFromArtifactory(details)
	switch response.StatusCode {
	case 409:
		cliutils.Exit(cliutils.ExitCodeError, "\nYour Artifactory server is not configured to encrypt passwords.\n"+
			"You may use \"art config --enc-password=false\"")
	case 200:
		details.Password = encPassword
		fmt.Println("Done.")
	default:
		cliutils.Exit(cliutils.ExitCodeError, "\nArtifactory response: "+response.Status)
	}
	return details
}
