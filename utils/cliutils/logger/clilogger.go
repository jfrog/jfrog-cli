package logger

import (
	"os"
	"fmt"
)

const (
	ERROR string = "ERROR"
	INFO = "INFO"
)

var Logger Log

func init() {
	Logger = CreateLogger()
}

type CliLogger struct {
	logLevel string
}

func CreateLogger() (logger *CliLogger) {
	logger = new(CliLogger)
	logger.setLevel()
	return
}

func (logger *CliLogger) setLevel() {
	if os.Getenv("JFROG_CLI_LOG_LEVEL") == ERROR {
		logger.logLevel = ERROR
	} else {
		logger.logLevel = INFO
	}
}

func (logger CliLogger) Info(a ...interface{}) {
	if logger.logLevel != ERROR {
		fmt.Println(a...)
	}
}

func (logger CliLogger) Error(a ...interface{}) {
	fmt.Println(a...)
}

type Log interface {
	Info(a ...interface{})
	Error(a ...interface{})
}

