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
}

func NewCliLogger() log.Log {
	logger := log.NewDefaultLogger()
	loglevel := os.Getenv("JFROG_CLI_LOG_LEVEL")
	if loglevel == "ERROR" {
		logger.SetLogLevel(log.ERROR)
	}
	if loglevel == "WARN" {
		logger.SetLogLevel(log.WARN)
	}
	if loglevel == "INFO" {
		logger.SetLogLevel(log.INFO)
	}
	if loglevel == "DEBUG" {
		logger.SetLogLevel(log.DEBUG)
	}
	return logger
}
