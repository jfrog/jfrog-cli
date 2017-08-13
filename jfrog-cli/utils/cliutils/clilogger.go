package cliutils

import (
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

var CliLogger log.Log

func init() {
	if CliLogger == nil {
		CliLogger = NewCliLogger()
	}
	// TODO - Remove this when all the command uses the new CliLogger
	log.Logger.SetLogLevel(log.GetCliLogLevel(os.Getenv("JFROG_CLI_LOG_LEVEL")))
}

func NewCliLogger() log.Log {
	logger := log.NewDefaultLogger()
	logger.SetLogLevel(log.GetCliLogLevel(os.Getenv("JFROG_CLI_LOG_LEVEL")))
	return logger
}