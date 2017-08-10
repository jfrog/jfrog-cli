package log

import (
	"os"
	"log"
)

var logger Log
type LogLevelType int
const (
	ERROR LogLevelType = iota
	WARN
	INFO
	DEBUG
)

func init() {
	if logger == nil {
		logger = NewDefaultLogger()
	}
}

type JfrogLogger struct {
	LogLevel LogLevelType
	DebugLog *log.Logger
	InfoLog  *log.Logger
	WarnLog  *log.Logger
	ErrorLog *log.Logger
}

func SetLogger(newLogger Log) {
	logger = newLogger
}

func NewDefaultLogger() (logger *JfrogLogger) {
	logger = new(JfrogLogger)
	logger.SetLogLevel(INFO)
	return
}

func (logger *JfrogLogger) SetLogLevel(LevelEnum LogLevelType) {
	logger.LogLevel = LevelEnum
	logger.DebugLog = log.New(os.Stdout, "[Debug] ", 0)
	logger.InfoLog = log.New(os.Stdout, "[Info] ", 0)
	logger.WarnLog = log.New(os.Stdout, "[Warn] ", 0)
	logger.ErrorLog = log.New(os.Stderr, "[Error] ", 0)
}

func GetLogLevel() LogLevelType {
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

func (logger JfrogLogger) GetLogLevel() LogLevelType {
	return logger.LogLevel
}

func (logger JfrogLogger) Debug(a ...interface{}) {
	if logger.GetLogLevel() >= DEBUG {
		logger.DebugLog.Println(a...)
	}
}

func (logger JfrogLogger) Info(a ...interface{}) {
	if logger.GetLogLevel() >= INFO {
		logger.InfoLog.Println(a...)
	}
}

func (logger JfrogLogger) Warn(a ...interface{}) {
	if logger.GetLogLevel() >= WARN {
		logger.WarnLog.Println(a...)
	}
}

func (logger JfrogLogger) Error(a ...interface{}) {
	if logger.GetLogLevel() >= ERROR {
		logger.ErrorLog.Println(a...)
	}
}

type Log interface {
	GetLogLevel() LogLevelType
	SetLogLevel(LogLevelType)
	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
}
