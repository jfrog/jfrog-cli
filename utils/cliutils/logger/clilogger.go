package logger

import (
	"os"
	"log"
)

const (
	ERROR string = "ERROR"
	INFO = "INFO"
)

var Logger Log

func init() {
	Logger = createLogger()
}

type CliLogger struct {
	logLevel string
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func createLogger() (logger *CliLogger) {
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
	logger.InfoLog = log.New(os.Stdout, "[Info:] ", 0)
	logger.ErrorLog = log.New(os.Stderr, "[Error:] ", 0)
}

func (logger CliLogger) Info(a ...interface{}) {
	if logger.logLevel != ERROR {
		logger.InfoLog.Println(a...)
	}
}

func (logger CliLogger) Error(a ...interface{}) {
	logger.ErrorLog.Println(a...)
}

type Log interface {
	Info(a ...interface{})
	Error(a ...interface{})
}

