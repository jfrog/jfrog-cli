package curl

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type CurlCommand struct {
	arguments      []string
	executablePath string
	rtDetails      *config.ArtifactoryDetails
}

func NewCurlCommand() *CurlCommand {
	return &CurlCommand{}
}

func (curlCmd *CurlCommand) SetArguments(arguments []string) *CurlCommand {
	curlCmd.arguments = arguments
	return curlCmd
}

func (curlCmd *CurlCommand) SetExecutablePath(executablePath string) *CurlCommand {
	curlCmd.executablePath = executablePath
	return curlCmd
}

func (curlCmd *CurlCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *CurlCommand {
	curlCmd.rtDetails = rtDetails
	return curlCmd
}

func (curlCmd *CurlCommand) Run() error {
	// Get curl execution path.
	execPath, err := exec.LookPath("curl")
	if err != nil {
		return errorutils.CheckError(err)
	}
	curlCmd.SetExecutablePath(execPath)

	// If the command already includes credentials flag, return an error.
	if curlCmd.isCredentialsFlagExists() {
		return errorutils.CheckError(errors.New("Curl command must not include credentials flag (-u or --user)."))
	}

	// If the command already includes certificates flag, return an error.
	if curlCmd.rtDetails.ClientCertificatePath != "" && curlCmd.isCertificateFlagExists() {
		return errorutils.CheckError(errors.New("Curl command must not include certificate flag (--cert or --key)."))
	}

	// Get target url for the curl command.
	uriIndex, targetUri, err := curlCmd.buildCommandUrl(curlCmd.rtDetails.Url)
	if err != nil {
		return err
	}

	// Replace url argument with complete url.
	curlCmd.arguments[uriIndex] = targetUri

	cmdWithoutCreds := strings.Join(curlCmd.arguments, " ")
	// Add credentials to curl command.
	credentialsMessage, err := curlCmd.addCommandCredentials()

	// Run curl.
	log.Debug(fmt.Sprintf("Executing curl command: '%s %s'", cmdWithoutCreds, credentialsMessage))
	return gofrogcmd.RunCmd(curlCmd)
}

func (curlCmd *CurlCommand) addCommandCredentials() (string, error) {
	certificateHelpPrefix := ""

	if curlCmd.rtDetails.ClientCertificatePath != "" {
		curlCmd.arguments = append(curlCmd.arguments,
			"--cert", curlCmd.rtDetails.ClientCertificatePath,
			"--key", curlCmd.rtDetails.ClientCertificateKeyPath)
		certificateHelpPrefix = "--cert *** --key *** "
	}

	if curlCmd.rtDetails.ApiKey != "" {
		// Add access token header.
		tokenHeader := fmt.Sprintf("X-JFrog-Art-Api: %s", curlCmd.rtDetails.ApiKey)
		curlCmd.arguments = append(curlCmd.arguments, "-H", tokenHeader)

		return certificateHelpPrefix + "-H \"X-JFrog-Art-Api: ***\"", nil
	} else if curlCmd.rtDetails.AccessToken != "" {
		// Add access token header.
		tokenHeader := fmt.Sprintf("Authorization: Bearer %s", curlCmd.rtDetails.AccessToken)
		curlCmd.arguments = append(curlCmd.arguments, "-H", tokenHeader)

		return certificateHelpPrefix + "-H \"Authorization: Bearer ***\"", nil
	}

	// Add credentials flag to Command. In case of flag duplication, the latter is used by Curl.
	credFlag := fmt.Sprintf("-u%s:%s", curlCmd.rtDetails.User, curlCmd.rtDetails.Password)
	curlCmd.arguments = append(curlCmd.arguments, credFlag)

	return certificateHelpPrefix + "-u***:***", nil
}

func (curlCmd *CurlCommand) buildCommandUrl(artifactoryUrl string) (uriIndex int, uriValue string, err error) {
	// Find command's URL argument.
	// Representing the target API for the Curl command.
	uriIndex, uriValue = curlCmd.findUriValueAndIndex()
	if uriIndex == -1 {
		err = errorutils.CheckError(errors.New("Could not find argument in curl command."))
		return
	}

	// If user provided full-url, throw an error.
	if strings.HasPrefix(uriValue, "http://") || strings.HasPrefix(uriValue, "https://") {
		err = errorutils.CheckError(errors.New("Curl command must not include full-url, but only the REST API URI (e.g '/api/system/ping')."))
		return
	}

	// Trim '/' prefix if exists.
	if strings.HasPrefix(uriValue, "/") {
		uriValue = strings.TrimPrefix(uriValue, "/")
	}

	// Attach Artifactory's url to the api.
	uriValue = artifactoryUrl + uriValue

	return
}

// Returns Artifactory details
func (curlCmd *CurlCommand) GetArtifactoryDetails() (*config.ArtifactoryDetails, error) {
	// Get --server-id flag value from the command, and remove it.
	flagIndex, valueIndex, serverIdValue, err := utils.FindFlag("--server-id", curlCmd.arguments)
	if err != nil {
		return nil, err
	}
	utils.RemoveFlagFromCommand(&curlCmd.arguments, flagIndex, valueIndex)
	return config.GetArtifactorySpecificConfig(serverIdValue)
}

// Find the URL argument in the Curl Command.
// A command flag is prefixed by '-' or '--'.
// Use this method ONLY after removing all JFrog-CLI flags, i.e flags in the form: '--my-flag=value' are not allowed.
// An argument is any provided candidate which is not a flag or a flag value.
func (curlCmd *CurlCommand) findUriValueAndIndex() (int, string) {
	skipThisArg := false
	for index, arg := range curlCmd.arguments {
		// Check if shouldn't check current arg.
		if skipThisArg {
			skipThisArg = false
			continue
		}

		// If starts with '--', meaning a flag which its value is at next slot.
		if strings.HasPrefix(arg, "--") {
			skipThisArg = true
			continue
		}

		// Check if '-'.
		if strings.HasPrefix(arg, "-") {
			if len(arg) > 2 {
				// Meaning that this flag also contains its value.
				continue
			}
			// If reached here, means that the flag value is at the next arg.
			skipThisArg = true
			continue
		}

		// Found an argument
		return index, arg
	}

	// If reached here, didn't find an argument.
	return -1, ""
}

// Return true if the curl command includes credentials flag.
// The searched flags are not CLI flags.
func (curlCmd *CurlCommand) isCredentialsFlagExists() bool {
	for _, arg := range curlCmd.arguments {
		if strings.HasPrefix(arg, "-u") || arg == "--user" {
			return true
		}
	}

	return false
}

// Return true if the curl command includes certificates flag.
// The searched flags are not CLI flags.
func (curlCmd *CurlCommand) isCertificateFlagExists() bool {
	for _, arg := range curlCmd.arguments {
		if arg == "--cert" || arg == "--key" {
			return true
		}
	}

	return false
}

func (curlCmd *CurlCommand) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, curlCmd.executablePath)
	cmd = append(cmd, curlCmd.arguments...)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (curlCmd *CurlCommand) GetEnv() map[string]string {
	return map[string]string{}
}

func (curlCmd *CurlCommand) GetStdWriter() io.WriteCloser {
	return nil
}

func (curlCmd *CurlCommand) GetErrWriter() io.WriteCloser {
	return nil
}

func (curlCmd *CurlCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return curlCmd.rtDetails, nil
}

func (curlCmd *CurlCommand) CommandName() string {
	return "rt_curl"
}
