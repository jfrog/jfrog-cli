package log

import (
	"os"
	"log"
)

var LogLevel = map[string]int{"ERROR": 0, "WARN":  1, "INFO":  2, "DEBUG": 3, }
var logger Log

func init() {
	logger = createLogger()
}

type CliLogger struct {
	LogLevel int
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
	logger.LogLevel = LogLevel["INFO"]
	if val, ok := LogLevel[os.Getenv("JFROG_CLI_LOG_LEVEL")]; ok {
		logger.LogLevel = val;
	}

	logger.DebugLog = log.New(os.Stdout, "[Debug] ", 0)
	logger.InfoLog = log.New(os.Stdout, "[Info] ", 0)
	logger.WarnLog = log.New(os.Stdout, "[Warn] ", 0)
	logger.ErrorLog = log.New(os.Stderr, "[Error] ", 0)
}

func GetLogLevel() int {
	return logger.GetLogLevel()
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

func (logger CliLogger) GetLogLevel() int {
	return logger.LogLevel
}

func (logger CliLogger) Debug(a ...interface{}) {
	if logger.LogLevel >= LogLevel["DEBUG"] {
		logger.DebugLog.Println(a...)
	}
}

func (logger CliLogger) Info(a ...interface{}) {
	if logger.LogLevel >= LogLevel["INFO"] {
		logger.InfoLog.Println(a...)
	}
}

func (logger CliLogger) Warn(a ...interface{}) {
	if logger.LogLevel >= LogLevel["WARN"] {
		logger.WarnLog.Println(a...)
	}
}

func (logger CliLogger) Error(a ...interface{}) {
	if logger.LogLevel >= LogLevel["ERROR"] {
		logger.ErrorLog.Println(a...)
	}
}


type Log interface {
	GetLogLevel() int
	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
}

