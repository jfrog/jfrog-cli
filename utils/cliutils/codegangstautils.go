package cliutils

import (
	"golang.org/x/exp/slices"
	"sort"
	"strconv"
	"strings"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-client-go/utils"
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

func GetThreadsCount(c *cli.Context) (threads int, err error) {
	return commonCliUtils.GetThreadsCount(c.String("threads"))
}

func ExtractCommand(c *cli.Context) []string {
	return slices.Clone(c.Args())
}

func GetSortedCommands(commands cli.CommandsByName) cli.CommandsByName {
	sort.Sort(commands)
	return commands
}
