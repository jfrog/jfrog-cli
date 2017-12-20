package log

import (
	"io"
	"log"
	"os"
)

var Logger Log

type LevelType int

const (
	ERROR LevelType = iota
	WARN
	INFO
	DEBUG
)

func init() {
	if Logger == nil {
		Logger = NewLogger()
	}
}

func NewLogger() Log {
	logger := new(jfrogLogger)
	logLevel := os.Getenv("JFROG_CLI_LOG_LEVEL")
	if logLevel != "" {
		logger.SetLogLevel(GetCliLogLevel(logLevel))
	} else {
		logger.SetLogLevel(INFO)
	}
	logger.SetOutputWriter(os.Stdout)
	logger.SetStderrWriter(os.Stderr)
	return logger
}

type jfrogLogger struct {
	LogLevel  LevelType
	OutputLog *log.Logger
	DebugLog  *log.Logger
	InfoLog   *log.Logger
	WarnLog   *log.Logger
	ErrorLog  *log.Logger
}

func SetLogger(newLogger Log) {
	Logger = newLogger
}

func (logger *jfrogLogger) SetLogLevel(LevelEnum LevelType) {
	logger.LogLevel = LevelEnum
}

func (logger *jfrogLogger) SetOutputWriter(writer io.Writer) {
	logger.OutputLog = log.New(writer, "", 0)
}

func (logger *jfrogLogger) SetStderrWriter(writer io.Writer) {
	logger.DebugLog = log.New(writer, "[Debug] ", 0)
	logger.InfoLog = log.New(writer, "[Info] ", 0)
	logger.WarnLog = log.New(writer, "[Warn] ", 0)
	logger.ErrorLog = log.New(writer, "[Error] ", 0)
}

func GetLogLevel() LevelType {
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

func Output(a ...interface{}) {
	Logger.Output(a...)
}

func (logger jfrogLogger) GetLogLevel() LevelType {
	return logger.LogLevel
}

func (logger jfrogLogger) Debug(a ...interface{}) {
	if logger.GetLogLevel() >= DEBUG {
		logger.DebugLog.Println(a...)
	}
}

func (logger jfrogLogger) Info(a ...interface{}) {
	if logger.GetLogLevel() >= INFO {
		logger.InfoLog.Println(a...)
	}
}

func (logger jfrogLogger) Warn(a ...interface{}) {
	if logger.GetLogLevel() >= WARN {
		logger.WarnLog.Println(a...)
	}
}

func (logger jfrogLogger) Error(a ...interface{}) {
	if logger.GetLogLevel() >= ERROR {
		logger.ErrorLog.Println(a...)
	}
}

func (logger jfrogLogger) Output(a ...interface{}) {
	logger.OutputLog.Println(a...)
}

type Log interface {
	GetLogLevel() LevelType
	SetLogLevel(LevelType)
	SetOutputWriter(writer io.Writer)
	SetStderrWriter(writer io.Writer)

	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
	Output(a ...interface{})
}

func GetCliLogLevel(logLevel string) LevelType {
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
