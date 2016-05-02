package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"strings"
)

func Config(details, defaultDetails *config.ArtifactoryDetails, interactive,
    shouldEncPassword bool) *config.ArtifactoryDetails {

    if details == nil {
        details = new(config.ArtifactoryDetails)
    }
	if interactive {
	    if defaultDetails == nil {
            defaultDetails = config.ReadArtifactoryConf()
	    }
		if details.Url == "" {
			ioutils.ScanFromConsole("Artifactory URL", &details.Url, defaultDetails.Url)
		}
		if strings.Index(details.Url, "ssh://") == 0 || strings.Index(details.Url, "SSH://") == 0 {
			readSshKeyPathFromConsole(details, defaultDetails)
		} else {
			ioutils.ReadCredentialsFromConsole(details, defaultDetails)
		}
	}
	details.Url = cliutils.AddTrailingSlashIfNeeded(details.Url)
	if shouldEncPassword {
		details = encryptPassword(details)
	}
	config.SaveArtifactoryConf(details)
	return details
}

func readSshKeyPathFromConsole(details, savedDetails *config.ArtifactoryDetails) {
	if details.SshKeyPath == "" {
		ioutils.ScanFromConsole("SSH key file path", &details.SshKeyPath, savedDetails.SshKeyPath)
	}
	if !ioutils.IsFileExists(details.SshKeyPath) {
		fmt.Println("Warning: Could not find SSH key file at: " + details.SshKeyPath)
	}
}

func ShowConfig() {
	details := config.ReadArtifactoryConf()
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
	config.SaveArtifactoryConf(new(config.ArtifactoryDetails))
}

func GetConfig() *config.ArtifactoryDetails {
	return config.ReadArtifactoryConf()
}

func encryptPassword(details *config.ArtifactoryDetails) *config.ArtifactoryDetails {
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
