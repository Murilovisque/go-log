package logs

import (
	"io"

	logs "github.com/Murilovisque/logs/v2/internal"
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

type loggerLevel struct {
	logErrorEnabled bool
	logWarnEnabled bool
	logInfoEnabled bool
	logDebugEnabled bool
	realLogger logs.Logger
}

func (l *loggerLevel) setLevel(level logs.LoggerLevelMode) {
	l.logErrorEnabled = anyLevelMatch(level, []logs.LoggerLevelMode{ logs.LogErrorMode, logs.LogWarnMode, logs.LogInfoMode, logs.LogDebugMode })
	l.logWarnEnabled = anyLevelMatch(level, []logs.LoggerLevelMode{ logs.LogWarnMode, logs.LogInfoMode, logs.LogDebugMode })
	l.logInfoEnabled = anyLevelMatch(level, []logs.LoggerLevelMode{ logs.LogInfoMode, logs.LogDebugMode })
	l.logDebugEnabled = anyLevelMatch(level, []logs.LoggerLevelMode{ logs.LogDebugMode })
}

func newLoggerLevel(level logs.LoggerLevelMode, loggerToWrap logs.Logger) (logs.Logger, error) {
	ll := loggerLevel{realLogger: loggerToWrap}
	ll.setLevel(level)
	return &ll, nil
}

func anyLevelMatch(level logs.LoggerLevelMode, allowedLevels []logs.LoggerLevelMode) bool {
	for _, l := range allowedLevels {
		if l == level {
			return true
		}
	}
	return false
}

func (l *loggerLevel) Fatalf(message string, v ...interface{}) {
	l.realLogger.Fatalf(message, v...)
}

func (l *loggerLevel) Infof(message string, v ...interface{}) {
	if l.logInfoEnabled {
		l.realLogger.Infof(message, v...)
	}
}

func (l *loggerLevel) Errorf(message string, v ...interface{}) {
	if l.logErrorEnabled {
		l.realLogger.Errorf(message, v...)
	}
}

func (l *loggerLevel) Debugf(message string, v ...interface{}) {
	if l.logDebugEnabled {
		l.realLogger.Debugf(message, v...)
	}
}

func (l *loggerLevel) Warnf(message string, v ...interface{}) {
	if l.logWarnEnabled {
		l.realLogger.Warnf(message, v...)
	}
}

func (l *loggerLevel) Fatal(message interface{}) {
	l.realLogger.Fatal(message)
}

func (l *loggerLevel) Info(message interface{}) {
	if l.logInfoEnabled {
		l.realLogger.Info(message)
	}
}

func (l *loggerLevel) Error(message interface{}) {
	if l.logErrorEnabled {
		l.realLogger.Error(message)
	}
}

func (l *loggerLevel) Debug(message interface{}) {
	if l.logDebugEnabled {
		l.realLogger.Debug(message)
	}
}

func (l *loggerLevel) Warn(message interface{}) {
	if l.logWarnEnabled {
		l.realLogger.Warn(message)
	}
}

func (l *loggerLevel) SetWriter(w io.Writer) {
	l.realLogger.SetWriter(w)
}

func (l *loggerLevel) Init() {
	l.realLogger.Init()
}

func (l *loggerLevel) FixedFieldsValues() []logs.FieldValue {
	return l.realLogger.FixedFieldsValues()
}
