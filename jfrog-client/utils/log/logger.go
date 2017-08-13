package log

import (
	"os"
	"log"
)

var Logger Log
type LogLevelType int
const (
	ERROR LogLevelType = iota
	WARN
	INFO
	DEBUG
)

func init() {
	if Logger == nil {
		Logger = NewDefaultLogger()
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
	Logger = newLogger
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
	return Logger.GetLogLevel()
}

func Debug(a ...interface{}) {
	Logger.Debug(a...)
}

func Info(a ...interface{}) {
	Logger.Info(a...)
}

func Warn(a ...interface{}) {
	Logger.Warn(a...)
}

func Error(a ...interface{}) {
	Logger.Error(a...)
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

func GetCliLogLevel(logLevel string) LogLevelType {
	switch logLevel {
	case "ERROR":
		return ERROR
	case "WARN":
		return WARN
	case "DEBUG":
		return DEBUG
	default:
		return INFO
	}
}