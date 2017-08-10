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
	switch logLevel := os.Getenv("JFROG_CLI_LOG_LEVEL"); logLevel {
	case "ERROR":
		logger.SetLogLevel(log.ERROR)
	case "WARN":
		logger.SetLogLevel(log.WARN)
	case "DEBUG":
		logger.SetLogLevel(log.DEBUG)
	default:
		logger.SetLogLevel(log.INFO)
	}
	return logger
}
