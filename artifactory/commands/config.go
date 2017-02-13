package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/ioutils"
	"strings"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func Config(details, defaultDetails *config.ArtifactoryDetails, interactive,
    shouldEncPassword bool) (*config.ArtifactoryDetails, error) {

    if details == nil {
        details = new(config.ArtifactoryDetails)
    }
    var err error
	if interactive {
	    if defaultDetails == nil {
            defaultDetails, err = config.ReadArtifactoryConf()
            if err != nil {
                return nil, err
            }
	    }
		if details.Url == "" {
			ioutils.ScanFromConsole("Artifactory URL", &details.Url, defaultDetails.Url)
		}
		if strings.Index(details.Url, "ssh://") == 0 || strings.Index(details.Url, "SSH://") == 0 {
			err = readSshKeyPathFromConsole(details, defaultDetails)
            if err != nil {
                return nil, err
            }
		} else {
		    if details.ApiKey == "" && details.Password == "" {
				ioutils.ScanFromConsole("API key (leave empty for basic authentication)", &details.ApiKey, "")
		    }
			if details.ApiKey == "" {
				ioutils.ReadCredentialsFromConsole(details, defaultDetails)
			}
		}
	}
	err = checkSingleAuthMethod(details)
    if err != nil {
        return nil, err
    }

	details.Url = cliutils.AddTrailingSlashIfNeeded(details.Url)
	if shouldEncPassword {
		details, err = encryptPassword(details)
        if err != nil {
            return nil, err
        }
	}
	err = config.SaveArtifactoryConf(details)
	return details, err
}

func readSshKeyPathFromConsole(details, savedDetails *config.ArtifactoryDetails) error {
	if details.SshKeyPath == "" {
		ioutils.ScanFromConsole("SSH key file path", &details.SshKeyPath, savedDetails.SshKeyPath)
	}

	details.SshKeyPath = cliutils.ReplaceTildeWithUserHome(details.SshKeyPath)
	exists, err := fileutils.IsFileExists(details.SshKeyPath)
	if err != nil {
	    return err
	}
	if !exists {
		log.Warn("Could not find SSH key file at:", details.SshKeyPath)
	}
	return nil
}

func ShowConfig() error {
	details, err := config.ReadArtifactoryConf()
	if err != nil {
	    return err
	}
	if details.Url != "" {
		fmt.Println("Url: " + details.Url)
	}
	if details.ApiKey != "" {
		fmt.Println("API key: " + details.ApiKey)
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
	return nil
}

func ClearConfig() {
	config.SaveArtifactoryConf(new(config.ArtifactoryDetails))
}

func GetConfig() (*config.ArtifactoryDetails, error) {
	return config.ReadArtifactoryConf()
}

func encryptPassword(details *config.ArtifactoryDetails) (*config.ArtifactoryDetails, error) {
	if details.Password == "" {
		return details, nil
	}
	log.Info("\nEncrypting password...")
	response, encPassword, err := utils.GetEncryptedPasswordFromArtifactory(details)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
        case 409:
            message := "\nYour Artifactory server is not configured to encrypt passwords.\n" +
                "You may use \"art config --enc-password=false\""
            err = cliutils.CheckError(errors.New(message))
        case 200:
            details.Password = encPassword
            log.Info("Done encrypting password.")
        default:
            err = cliutils.CheckError(errors.New("\nArtifactory response: " + response.Status))
	}
	return details, err
}

func checkSingleAuthMethod(details *config.ArtifactoryDetails) (err error) {
	boolArr := []bool{details.User != "" && details.Password != "", details.ApiKey != "", details.SshKeyPath != ""}
	if (cliutils.SumTrueValues(boolArr) > 1) {
		err = cliutils.CheckError(errors.New("Only one authentication method is allowd: Username/Password, API key or RSA tokens."))
	}
	return
}

type ConfigFlags struct {
	ArtDetails   *config.ArtifactoryDetails
	Interactive  bool
	EncPassword  bool
}