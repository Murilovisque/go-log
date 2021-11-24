package logs

import (
	"fmt"
	"io"
	"log"
	"strings"

)

const (
	logErrorMode = "ERROR"
	logWarnMode = "WARN"
	logInfoMode  = "INFO"
	logFatalMode = "FATAL"
	logDebugMode = "DEBUG"
)

var (
	logsMode = [5]string{logErrorMode, logWarnMode, logInfoMode, logFatalMode, logDebugMode}
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
}


type SimpleLogger struct {
	fieldsValues []FieldValue
	fixedLogMessage string
}

func (l *SimpleLogger) Fatalf(message string, v ...interface{}) {
	log.Fatal(l.buildFormatedMessage(logFatalMode, message, v...))
}

func (l *SimpleLogger) Infof(message string, v ...interface{}) {
	log.Println(l.buildFormatedMessage(logInfoMode, message, v...))
}

func (l *SimpleLogger) Errorf(message string, v ...interface{}) {
	log.Println(l.buildFormatedMessage(logErrorMode, message, v...))
}

func (l *SimpleLogger) Debugf(message string, v ...interface{}) {
	log.Println(l.buildFormatedMessage(logDebugMode, message, v...))
}

func (l *SimpleLogger) Warnf(message string, v ...interface{}) {
	log.Println(l.buildFormatedMessage(logWarnMode, message, v...))
}

func (l *SimpleLogger) Fatal(message interface{}) {
	log.Fatal(l.buildMessage(logFatalMode, message))
}

func (l *SimpleLogger) Info(message interface{}) {
	log.Println(l.buildMessage(logInfoMode, message))
}

func (l *SimpleLogger) Error(message interface{}) {
	log.Println(l.buildMessage(logErrorMode, message))
}

func (l *SimpleLogger) Debug(message interface{}) {
	log.Println(l.buildMessage(logDebugMode, message))
}

func (l *SimpleLogger) Warn(message interface{}) {
	log.Println(l.buildMessage(logWarnMode, message))
}

func (l *SimpleLogger) FixedFieldsValues() []FieldValue {
	return l.fieldsValues
}

func (l *SimpleLogger) buildMessage(level string, message interface{}) string {
	return fmt.Sprintf("%s%s%v", level, l.fixedLogMessage, message)
}

func (l *SimpleLogger) buildFormatedMessage(level, message string, v ...interface{}) string {
	return fmt.Sprintf("%s%s%s", level, l.fixedLogMessage, fmt.Sprintf(message, v...))
}

func (l *SimpleLogger) SetWriter(writer io.Writer) {
	log.SetOutput(writer)
}

func (l *SimpleLogger) Init() {
	builder := strings.Builder{}
	for _, fv := range l.fieldsValues {
		builder.WriteString(" [")
		builder.WriteString(fv.Key)
		builder.WriteString(": ")
		builder.WriteString(fmt.Sprint(fv.Val))
		builder.WriteString("]")
	}
	builder.WriteString(" * ")
	l.fixedLogMessage = builder.String()
}

type FieldValue struct {
	Key string
	Val interface{}
}

func FixedValue(key string, val interface{}) FieldValue {
	return FieldValue{Key: key, Val: val}
}
