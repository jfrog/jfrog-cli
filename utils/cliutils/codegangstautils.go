package cliutils

import (
	errors2 "errors"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"sort"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func GetIntFlagValue(c *cli.Context, flagName string, defValue int) (int, error) {
	if c.IsSet(flagName) {
		flagIntVal, err := strconv.Atoi(c.String(flagName))
		err = utils.CheckErrorWithMessage(err, "can't parse "+flagName+" value: "+c.String(flagName))
		return flagIntVal, err
	}
	return defValue, nil
}

func GetStringsArrFlagValue(c *cli.Context, flagName string) (resultArray []string) {
	if c.IsSet(flagName) {
		resultArray = append(resultArray, strings.Split(c.String(flagName), ";")...)
	}
	return
}

func GetThreadsCount(c *cli.Context, defaultNum int) (threads int, err error) {
	threads = defaultNum
	err = nil
	if c.String("threads") != "" {
		threads, err = strconv.Atoi(c.String("threads"))
		if err != nil || threads < 1 {
			err = errors.New("the '--threads' option should have a numeric positive value")
			return 0, err
		}
	}
	return threads, nil
}

func ExtractCommand(c *cli.Context) (command []string) {
	command = make([]string, len(c.Args()))
	copy(command, c.Args())
	return command
}

func GetSortedCommands(commands cli.CommandsByName) cli.CommandsByName {
	sort.Sort(commands)
	return commands
}

func GetRetries(c *cli.Context) (retries int, err error) {
	retries = Retries
	err = nil
	if c.String("retries") != "" {
		retries, err = strconv.Atoi(c.String("retries"))
		if err != nil {
			err = errors2.New("The '--retries' option should have a numeric value. " + GetDocumentationMessage())
			return 0, err
		}
	}

	return retries, nil
}

// GetRetryWaitTime extract the given '--retry-wait-time' value and validate that it has a numeric value and a 's'/'ms' suffix.
// The returned wait time's value is in milliseconds.
func GetRetryWaitTime(c *cli.Context) (waitMilliSecs int, err error) {
	waitMilliSecs = RetryWaitMilliSecs
	waitTimeStringValue := c.String("retry-wait-time")
	useSeconds := false
	if waitTimeStringValue != "" {
		if strings.HasSuffix(waitTimeStringValue, "ms") {
			waitTimeStringValue = strings.TrimSuffix(waitTimeStringValue, "ms")
		} else if strings.HasSuffix(waitTimeStringValue, "s") {
			useSeconds = true
			waitTimeStringValue = strings.TrimSuffix(waitTimeStringValue, "s")
		} else {
			err = getRetryWaitTimeVerificationError()
			return
		}
		waitMilliSecs, err = strconv.Atoi(waitTimeStringValue)
		if err != nil {
			err = getRetryWaitTimeVerificationError()
			return
		}
		// Convert seconds to milliseconds
		if useSeconds {
			waitMilliSecs = waitMilliSecs * 1000
		}
	}
	return
}

func getRetryWaitTimeVerificationError() error {
	return errorutils.CheckError(errors2.New("The '--retry-wait-time' option should have a numeric value with 's'/'ms' suffix. " + GetDocumentationMessage()))
}
