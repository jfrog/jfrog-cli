package log

import (
	"os"
	"log"
)

var logLevel = map[string]int{"ERROR": 0, "WARN":  1, "INFO":  2, "DEBUG": 3, }
var logger Log

func init() {
	logger = createLogger()
}

type CliLogger struct {
	logLevel int
	DebugLog *log.Logger
	InfoLog  *log.Logger
	WarnLog  *log.Logger
	ErrorLog *log.Logger
}

func createLogger() (logger *CliLogger) {
	logger = new(CliLogger)
	logger.setLevel()
	return
}

func (logger *CliLogger) setLevel() {
	logger.logLevel = logLevel["INFO"]
	if val, ok := logLevel[os.Getenv("JFROG_CLI_LOG_LEVEL")]; ok {
		logger.logLevel = val;
	}

	logger.DebugLog = log.New(os.Stdout, "[Debug:] ", 0)
	logger.InfoLog = log.New(os.Stdout, "[Info:] ", 0)
	logger.WarnLog = log.New(os.Stdout, "[Warn:] ", 0)
	logger.ErrorLog = log.New(os.Stdout, "[Error:] ", 0)
}

func Debug(a ...interface{}) {
	logger.Debug(a...)
}

func Info(a ...interface{}) {
	logger.Info(a...)
}

func Warn(a ...interface{}) {
	logger.Warn(a...)
}

func Error(a ...interface{}) {
	logger.Error(a...)
}

func (logger CliLogger) Debug(a ...interface{}) {
	if logger.logLevel >= logLevel["DEBUG"] {
		logger.DebugLog.Println(a...)
	}
}

func (logger CliLogger) Info(a ...interface{}) {
	if logger.logLevel >= logLevel["INFO"] {
		logger.InfoLog.Println(a...)
	}
}

func (logger CliLogger) Warn(a ...interface{}) {
	if logger.logLevel >= logLevel["WARN"] {
		logger.WarnLog.Println(a...)
	}
}

func (logger CliLogger) Error(a ...interface{}) {
	if logger.logLevel >= logLevel["ERROR"] {
		logger.ErrorLog.Println(a...)
	}
}

type Log interface {
	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
}

