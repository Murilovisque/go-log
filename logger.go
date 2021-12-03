package logs

import (
	logs "github.com/Murilovisque/logs/v2/internal"
)

const (
	LevelFatal = logs.LogFatalMode
	LevelError = logs.LogErrorMode
	LevelWarn = logs.LogWarnMode
	LevelInfo = logs.LogInfoMode
	LevelDebug = logs.LogDebugMode
)

type Logger interface {
	Fatalf(message string, v ...interface{})
	Infof(message string, v ...interface{})
	Errorf(message string, v ...interface{})
	Debugf(message string, v ...interface{})
	Warnf(message string, v ...interface{})
	Fatal(message interface{})
	Info(message interface{})
	Error(message interface{})
	Debug(message interface{})
	Warn(message interface{})
}

func FixedFieldValue(key string, val interface{}) logs.FieldValue {
	return logs.FieldValue{Key: key, Val: val}
}

