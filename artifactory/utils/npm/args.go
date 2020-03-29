package npm

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"strconv"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

func ExtractNpmOptionsFromArgs(args []string) (threads int, jsonOutput bool, cleanArgs []string, buildConfig *utils.BuildConfiguration, err error) {
	threads = 3
	// Extract threads information from the args.
	flagIndex, valueIndex, numOfThreads, err := utils.FindFlag("--threads", args)
	if err != nil {
		return
	}
	utils.RemoveFlagFromCommand(&args, flagIndex, valueIndex)
	if numOfThreads != "" {
		threads, err = strconv.Atoi(numOfThreads)
		if err != nil {
			err = errorutils.CheckError(err)
			return
		}
	}

	// Since we use --json flag for retrieving the npm config for writing the temp .npmrc, json=true is written to the config list.
	// We don't want to force the json output for all users, so we check whether the json output was explicitly required.
	flagIndex, jsonOutput, err = utils.FindBooleanFlag("--json", args)
	if err != nil {
		return
	}
	// Since boolean flag might appear as --flag or --flag=value, the value index is the same as the flag index.
	utils.RemoveFlagFromCommand(&args, flagIndex, flagIndex)

	cleanArgs, buildConfig, err = utils.ExtractBuildDetailsFromArgs(args)
	return
}
