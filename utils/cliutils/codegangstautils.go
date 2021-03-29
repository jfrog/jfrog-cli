package cliutils

import (
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-client-go/utils"
	"sort"
	"strconv"
	"strings"
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
		for _, singleValue := range strings.Split(c.String(flagName), ";") {
			resultArray = append(resultArray, singleValue)
		}
	}
	return
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
