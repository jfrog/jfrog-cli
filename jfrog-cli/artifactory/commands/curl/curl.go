package curl

import (
	"errors"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"os/exec"
	"strings"
)

func Execute(args []string) error {
	// Get curl execution path.
	execPath, err := exec.LookPath("curl")
	if err != nil {
		return errorutils.CheckError(err)
	}

	// Create curl command.
	command := &CurlCommand{ExecutablePath: execPath, Arguments: args}

	// Get server-id, remove from the command if exists.
	serverIdValue, err := command.getAndRemoveServerIdFromCommand()
	if err != nil {
		return err
	}

	// Get Artifactory details for this ID.
	artDetails, err := config.GetArtifactorySpecificConfig(serverIdValue)
	if err != nil {
		return err
	}

	// If the command already include credentials flag, return an error.
	if command.isCredentialsFlagExists() {
		return errorutils.CheckError(errors.New("Curl command must not include credentials flag (-u or --user)."))
	}

	// Get target url for the curl command.
	urlIndex, targetUrl, err := command.buildCommandUrl(artDetails.Url)
	if err != nil {
		return err
	}

	// Replace url argument with complete url.
	command.Arguments[urlIndex] = targetUrl

	log.Debug(fmt.Sprintf("Executing curl command: '%s -u***:***'", strings.Join(command.Arguments, " ")))
	// Add credentials flag to Command. In case of flag duplication, the latter is used by Curl.
	credFlag := fmt.Sprintf("-u%s:%s", artDetails.User, artDetails.Password)
	command.Arguments = append(command.Arguments, credFlag)

	// Run curl.
	return gofrogcmd.RunCmd(command)
}

func (curlCmd *CurlCommand) buildCommandUrl(artifactoryUrl string) (urlIndex int, urlValue string, err error) {
	// Find command's URL argument.
	// Representing the target API for the Curl command.
	urlIndex, urlValue = curlCmd.findUrlValueAndIndex()
	if urlIndex == -1 {
		err = errorutils.CheckError(errors.New("Could not find argument in curl command."))
		return
	}

	// If user provided full-url, throw an error.
	if strings.HasPrefix(urlValue, "http://") || strings.HasPrefix(urlValue, "https://") {
		err = errorutils.CheckError(errors.New("Curl command must not include full-url, provide only the REST API (e.g '/api/system/ping')."))
		return
	}

	// Trim '/' prefix if exists.
	if strings.HasPrefix(urlValue, "/") {
		urlValue = strings.TrimPrefix(urlValue, "/")
	}

	// Attach Artifactory's url to the api.
	urlValue = artifactoryUrl + urlValue

	return
}

// Get --server-id flag value from the command, and remove it.
func (curlCmd *CurlCommand) getAndRemoveServerIdFromCommand() (string, error) {
	// Get server id.
	serverIdFlagIndex, serverIdValueIndex, serverIdValue, err := curlCmd.findFlag("--server-id")
	if err != nil {
		return "", err
	}

	// Remove --server-id from Command if required.
	if serverIdFlagIndex != -1 {
		curlCmd.Arguments = append(curlCmd.Arguments[:serverIdFlagIndex], curlCmd.Arguments[serverIdValueIndex+1:]...)
	}

	return serverIdValue, nil
}

// Find the URL argument in the Curl Command.
// A command flag is prefixed by '-' or '--'.
// Use this method ONLY after removing all JFrog-CLI flags, i.e flags in the form: '--my-flag=value' are not allowed.
// An argument is any provided candidate which is not a flag or a flag value.
func (curlCmd *CurlCommand) findUrlValueAndIndex() (int, string) {
	skipThisArg := false
	for index, arg := range curlCmd.Arguments {
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
func (curlCmd *CurlCommand) isCredentialsFlagExists() (bool) {
	for _, arg := range curlCmd.Arguments {
		if strings.HasPrefix(arg, "-u") || arg == "--user" {
			return true
		}
	}

	return false
}

// Find value of required CLI flag in Command.
// If flag does not exist, returned indexes are -1 and error is nil.
// Return values:
// err - error if flag exists but failed to extract its value.
// flagIndex - index of flagName in Command.
// flagValueIndex - index in Command in which the value of the flag exists.
// flagValue - value of flagName.
func (curlCmd *CurlCommand) findFlag(flagName string) (flagIndex, flagValueIndex int, flagValue string, err error) {
	flagIndex = -1
	flagValueIndex = -1
	for index, arg := range curlCmd.Arguments {
		// Check current argument.
		if !strings.HasPrefix(arg, flagName) {
			continue
		}

		// Get flag value.
		flagValue, flagValueIndex, err = curlCmd.getFlagValueAndValueIndex(flagName, index)
		if err != nil {
			return
		}

		// If was not the correct flag, continue looking.
		if flagValueIndex == -1 {
			continue
		}

		// Return value.
		flagIndex = index
		return
	}

	// Flag not found.
	return
}

// Get the provided flag's value, and the index of the value.
// Value-index can either be same as flag's index, or the next one.
// Return error if flag is found, but couldn't extract value.
// If the provided index doesn't contain the searched flag, return flagIndex = -1.
func (curlCmd *CurlCommand) getFlagValueAndValueIndex(flagName string, flagIndex int) (flagValue string, flagValueIndex int, err error) {
	indexValue := curlCmd.Arguments[flagIndex]

	// Check if flag is in form '--server-id=myServer'
	indexValue = strings.TrimPrefix(indexValue, flagName)
	if strings.HasPrefix(indexValue, "=") {
		if len(indexValue) > 1 {
			return indexValue[1:], flagIndex, nil
		}
		return "", -1, errorutils.CheckError(errors.New(fmt.Sprintf("Flag %s is provided with empty value.", flagName)))
	}

	// Check if it is a different flag with same prefix, e.g --server-id-another
	if len(indexValue) > 0 {
		return "", -1, nil
	}

	// If reached here, expect the flag value in next argument.
	if len(curlCmd.Arguments) < flagIndex+2 {
		// Flag value does not exist.
		return "", -1, errorutils.CheckError(errors.New(fmt.Sprintf("Failed extracting value of provided flag: %s.", flagName)))
	}

	nextIndexValue := curlCmd.Arguments[flagIndex+1]
	// Don't allow next value to be a flag.
	if strings.HasPrefix(nextIndexValue, "-") {
		// Flag value does not exist.
		return "", -1, errorutils.CheckError(errors.New(fmt.Sprintf("Failed extracting value of provided flag: %s.", flagName)))
	}

	return nextIndexValue, flagIndex + 1, nil
}

type CurlCommand struct {
	Arguments      []string
	ExecutablePath string
}

func (curlCmd *CurlCommand) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, curlCmd.ExecutablePath)
	cmd = append(cmd, curlCmd.Arguments...)
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
