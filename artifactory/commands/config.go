package commands

import (
    "os"
    "fmt"
    "bytes"
    "strings"
    "syscall"
    "io/ioutil"
    "encoding/json"
    "golang.org/x/crypto/ssh/terminal"
    "github.com/jFrogdev/jfrog-cli-go/cliutils"
    "github.com/jFrogdev/jfrog-cli-go/artifactory/utils"
)

func Config(details *utils.ArtifactoryDetails, interactive, shouldEncPassword bool) {
    if interactive {
        if details.Url == "" {
            print("Artifactory URL: ")
            fmt.Scanln(&details.Url)
        }

        if strings.Index(details.Url, "ssh://") == 0 || strings.Index(details.Url, "SSH://") == 0 {
            readSshKeyPathFromConsole(details)
        } else {
            readCredentialsFromConsole(details)
        }
    }
    details.Url = cliutils.AddTrailingSlashIfNeeded(details.Url)
    if shouldEncPassword {
        details = encryptPassword(details)
    }
    writeConfFile(details)
}

func readSshKeyPathFromConsole(details *utils.ArtifactoryDetails) {
    if details.SshKeyPath == "" {
        print("SSH key file path: ")
        fmt.Scanln(&details.SshKeyPath)
    }
    if !cliutils.IsFileExists(details.SshKeyPath) {
        fmt.Println("Warning: Could not find SSH key file at: " + details.SshKeyPath)
    }
}

func readCredentialsFromConsole(details *utils.ArtifactoryDetails) {
    if details.User == "" {
        print("User: ")
        fmt.Scanln(&details.User)
    }
    if details.Password == "" {
        print("Password: ")
        bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
        details.Password = string(bytePassword)
        cliutils.CheckError(err)
    }
}

func ShowConfig() {
    details := readConfFile()
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
    writeConfFile(new(utils.ArtifactoryDetails))
}

func GetConfig() *utils.ArtifactoryDetails {
    return readConfFile()
}

func encryptPassword(details *utils.ArtifactoryDetails) *utils.ArtifactoryDetails {
    if details.Password == "" {
        return details
    }
    response, encPassword := utils.GetEncryptedPasswordFromArtifactory(details)
    switch response.StatusCode {
        case 409:
            cliutils.Exit(cliutils.ExitCodeError, "\nYour Artifactory server is not configured to encrypt passwords.\n" +
                "You may use \"art config --enc-password=false\"")
        case 200:
            details.Password = encPassword
        default:
            cliutils.Exit(cliutils.ExitCodeError, "\nArtifactory response: " + response.Status)
    }
    return details
}

func getConFilePath() string {
    userDir := cliutils.GetHomeDir()
    if userDir == "" {
        cliutils.Exit(cliutils.ExitCodeError, "Couldn't find home directory. Make sure your HOME environment variable is set.")
    }
    confPath := userDir + "/.jfrog/"
    os.MkdirAll(confPath ,0777)
    return confPath + "art-cli.conf"
}

func writeConfFile(details *utils.ArtifactoryDetails) {
    confFilePath := getConFilePath()
    if !cliutils.IsFileExists(confFilePath) {
        out, err := os.Create(confFilePath)
        cliutils.CheckError(err)
        defer out.Close()
    }

    b, err := json.Marshal(&details)
    cliutils.CheckError(err)
    var content bytes.Buffer
    err = json.Indent(&content, b, "", "  ")
    cliutils.CheckError(err)

    ioutil.WriteFile(confFilePath,[]byte(content.String()), 0x777)
}

func readConfFile() *utils.ArtifactoryDetails {
    confFilePath := getConFilePath()
    details := new(utils.ArtifactoryDetails)
    if !cliutils.IsFileExists(confFilePath) {
        return details
    }
    content := cliutils.ReadFile(confFilePath)
    json.Unmarshal(content, &details)

    return details
}