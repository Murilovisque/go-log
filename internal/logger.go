package logs

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

type LoggerLevelMode string

const (
	LogFatalMode LoggerLevelMode = "FATAL"
	LogErrorMode LoggerLevelMode = "ERROR"
	LogWarnMode  LoggerLevelMode = "WARN"
	LogInfoMode  LoggerLevelMode = "INFO"
	LogDebugMode LoggerLevelMode = "DEBUG"
)

var (
	LogsMode    = []LoggerLevelMode{LogFatalMode, LogErrorMode, LogWarnMode, LogInfoMode, LogDebugMode}
	builderPool = sync.Pool{
		New: func() interface{} {
			return new(strings.Builder)
		},
	}
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
	SetWriter(io.Writer)
	Init()
	FixedFieldsValues() []FieldValue
	Close()
}

type SimpleLogger struct {
	FieldsValues    []FieldValue
	LevelSelected   LoggerLevelMode
	fixedLogMessage string
	logErrorEnabled bool
	logWarnEnabled  bool
	logInfoEnabled  bool
	logDebugEnabled bool
}

func (l *SimpleLogger) Init() {
	if l.FieldsValues == nil {
		l.FieldsValues = []FieldValue{}
	}
	builder := builderPool.Get().(*strings.Builder)
	for _, fv := range l.FieldsValues {
		builder.WriteString(" [")
		builder.WriteString(fv.Key)
		builder.WriteString(": ")
		builder.WriteString(fmt.Sprint(fv.Val))
		builder.WriteString("]")
	}
	builder.WriteString(" * ")
	l.fixedLogMessage = builder.String()
	builder.Reset()
	builderPool.Put(builder)

	l.logErrorEnabled = anyLevelMatch(l.LevelSelected, []LoggerLevelMode{LogErrorMode, LogWarnMode, LogInfoMode, LogDebugMode})
	l.logWarnEnabled = anyLevelMatch(l.LevelSelected, []LoggerLevelMode{LogWarnMode, LogInfoMode, LogDebugMode})
	l.logInfoEnabled = anyLevelMatch(l.LevelSelected, []LoggerLevelMode{LogInfoMode, LogDebugMode})
	l.logDebugEnabled = anyLevelMatch(l.LevelSelected, []LoggerLevelMode{LogDebugMode})
}

func anyLevelMatch(level LoggerLevelMode, allowedLevels []LoggerLevelMode) bool {
	for _, l := range allowedLevels {
		if l == level {
			return true
		}
	}
	return false
}

func (l *SimpleLogger) Fatalf(message string, v ...interface{}) {
	log.Fatal(l.buildFormatedMessage(LogFatalMode, message, v...))
}

func (l *SimpleLogger) Infof(message string, v ...interface{}) {
	if l.logInfoEnabled {
		log.Println(l.buildFormatedMessage(LogInfoMode, message, v...))
	}
}

func (l *SimpleLogger) Errorf(message string, v ...interface{}) {
	if l.logErrorEnabled {
		log.Println(l.buildFormatedMessage(LogErrorMode, message, v...))
	}
}

func (l *SimpleLogger) Debugf(message string, v ...interface{}) {
	if l.logDebugEnabled {
		log.Println(l.buildFormatedMessage(LogDebugMode, message, v...))
	}
}

func (l *SimpleLogger) Warnf(message string, v ...interface{}) {
	if l.logWarnEnabled {
		log.Println(l.buildFormatedMessage(LogWarnMode, message, v...))
	}
}

func (l *SimpleLogger) Fatal(message interface{}) {
	log.Fatal(l.buildMessage(LogFatalMode, message))
}

func (l *SimpleLogger) Info(message interface{}) {
	if l.logInfoEnabled {
		log.Println(l.buildMessage(LogInfoMode, message))
	}
}

func (l *SimpleLogger) Error(message interface{}) {
	if l.logErrorEnabled {
		log.Println(l.buildMessage(LogErrorMode, message))
	}
}

func (l *SimpleLogger) Debug(message interface{}) {
	if l.logDebugEnabled {
		log.Println(l.buildMessage(LogDebugMode, message))
	}
}

func (l *SimpleLogger) Warn(message interface{}) {
	if l.logWarnEnabled {
		log.Println(l.buildMessage(LogWarnMode, message))
	}
}

func (l *SimpleLogger) FixedFieldsValues() []FieldValue {
	return l.FieldsValues
}

func (l *SimpleLogger) Close() {}

func (l *SimpleLogger) buildMessage(level LoggerLevelMode, message interface{}) string {
	return fmt.Sprintf("%s%s%v", level, l.fixedLogMessage, message)
}

func (l *SimpleLogger) buildFormatedMessage(level LoggerLevelMode, message string, v ...interface{}) string {
	return fmt.Sprintf("%s%s%s", level, l.fixedLogMessage, fmt.Sprintf(message, v...))
}

func (l *SimpleLogger) SetWriter(writer io.Writer) {
	log.SetOutput(writer)
}

type FieldValue struct {
	Key string
	Val interface{}
}
