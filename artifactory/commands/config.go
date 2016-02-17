package commands

import (
    "fmt"
    "strings"
    "syscall"
    "golang.org/x/crypto/ssh/terminal"
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
    "github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

func Config(details *cliutils.ArtifactoryDetails, interactive, shouldEncPassword bool) {
    if interactive {
        savedDetails := cliutils.ReadArtifactoryConf()

        if details.Url == "" {
            print("Artifactory URL [" + savedDetails.Url + "]: ")
            cliutils.ScanFromConsole(&details.Url, savedDetails.Url)
        }

        if strings.Index(details.Url, "ssh://") == 0 || strings.Index(details.Url, "SSH://") == 0 {
            readSshKeyPathFromConsole(details, savedDetails)
        } else {
            readCredentialsFromConsole(details, savedDetails)
        }
    }
    details.Url = cliutils.AddTrailingSlashIfNeeded(details.Url)
    if shouldEncPassword {
        details = encryptPassword(details)
    }
    cliutils.SaveArtifactoryConf(details)
}

func readSshKeyPathFromConsole(details, savedDetails *cliutils.ArtifactoryDetails) {
    if details.SshKeyPath == "" {
        print("SSH key file path [" + savedDetails.SshKeyPath + "]: ")
        cliutils.ScanFromConsole(&details.SshKeyPath, savedDetails.SshKeyPath)
    }
    if !cliutils.IsFileExists(details.SshKeyPath) {
        fmt.Println("Warning: Could not find SSH key file at: " + details.SshKeyPath)
    }
}

func readCredentialsFromConsole(details, savedDetails *cliutils.ArtifactoryDetails) {
    if details.User == "" {
        print("User: [" + savedDetails.User + "]: ")
        cliutils.ScanFromConsole(&details.User, savedDetails.User)
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
        fmt.Println("Password: " + details.Password)
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
            cliutils.Exit(cliutils.ExitCodeError, "\nYour Artifactory server is not configured to encrypt passwords.\n" +
                "You may use \"art config --enc-password=false\"")
        case 200:
            details.Password = encPassword
            fmt.Println("Done.")
        default:
            cliutils.Exit(cliutils.ExitCodeError, "\nArtifactory response: " + response.Status)
    }
    return details
}