package logs

import (
	"io"
	"log"
	"os"
)

var (
	globalLogger Logger
)

func init() {
	globalLogger = NewLogger()
	log.SetFlags(log.LstdFlags)
}

// NewLogger gets a new Logger
func NewLogger(fixedValues ...FieldValue) Logger {
	return &SimpleLogger{fieldsValues: fixedValues[:]}
}

// NewLoggerWithWriter gets a new Logger with writer
func NewLoggerWithWriter(w io.Writer, fixedValues ...FieldValue) Logger {
	l := NewLogger(fixedValues...)
	l.SetWriter(w)
	return l
}

// NewLoggerWithLogFile to log using file
func NewLoggerWithLogFile(filename string, fixedValues ...FieldValue) (Logger, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return NewLoggerWithWriter(f, fixedValues...), nil
}

// NewChildLogger gets a new Logger based in global Logger
func NewChildLogger(fixedValues ...FieldValue) Logger {
	globalFixedValues := globalLogger.FixedFieldsValues()[:]
	globalFixedValues = append(globalFixedValues, fixedValues...)
	return NewLogger(globalFixedValues...)
}

func Init(l Logger) {
	globalLogger = l
	globalLogger.Init()
}


// Fatal logs using the globalLogger
func Fatal(message interface{}) {
	globalLogger.Fatal(message)
}

// Info logs using the globalLogger
func Info(message interface{}) {
	globalLogger.Info(message)
}

// Error logs using the globalLogger
func Error(message interface{}) {
	globalLogger.Error(message)
}

// Debug logs using the globalLogger
func Debug(message interface{}) {
	globalLogger.Debug(message)
}
// Warn logs using the globalLogger
func Warn(message interface{}) {
	globalLogger.Warn(message)
}

// Fatalf logs using the globalLogger
func Fatalf(message string, v ...interface{}) {
	globalLogger.Fatalf(message, v...)
}

// Infof logs using the globalLogger
func Infof(message string, v ...interface{}) {
	globalLogger.Infof(message, v...)
}

// Errorf logs using the globalLogger
func Errorf(message string, v ...interface{}) {
	globalLogger.Errorf(message, v...)
}

// Debugf logs using the globalLogger
func Debugf(message string, v ...interface{}) {
	globalLogger.Debugf(message, v...)
}

// Warnf logs using the globalLogger
func Warnf(message string, v ...interface{}) {
	globalLogger.Warnf(message, v...)
}

