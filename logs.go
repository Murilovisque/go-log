package logs

import (
	"io"
	"log"
	"os"

	logs "github.com/Murilovisque/logs/v2/internal"
	"github.com/Murilovisque/logs/v2/internal/rotating"
)

var (
	globalLogger logs.Logger
)

func init() {
	log.SetFlags(log.LstdFlags)
	err := initGlobalLogger(logs.LogDebugMode, &logs.SimpleLogger{FieldsValues: []logs.FieldValue{}})
	if err != nil {
		panic(err)
	}
}

func InitWithLogFile(level, filename string, fixedValues ...logs.FieldValue) error {
	l, err := newLoggerWithLogFile(filename, fixedValues...)
	if err != nil {
		return err
	}
	return initGlobalLogger(level, l)
}

func InitWithWriter(level string, w io.Writer, fixedValues ...logs.FieldValue) error {
	return initGlobalLogger(level, newLoggerWithWriter(w, fixedValues...))
}

func InitWithRotatingLogFile(level, filename string, rotatingScheme rotating.TimeRotatingScheme, amountOfFilesToRetain int, fixedValues ...logs.FieldValue) error {
	l, err := rotating.NewTimeRotatingLogger(filename, rotatingScheme, amountOfFilesToRetain, fixedValues...)
	if err != nil {
		return err
	}
	return initGlobalLogger(level, l)
}

func NewChildLogger(fixedValues ...logs.FieldValue) Logger {
	globalFixedValues := globalLogger.FixedFieldsValues()[:]
	globalFixedValues = append(globalFixedValues, fixedValues...)
	l := logs.SimpleLogger{FieldsValues: globalFixedValues[:]}
	l.Init()
	return &l
}

func initGlobalLogger(level string, l logs.Logger) error {
	var err error
	globalLogger, err = newLoggerLevel(level, l)
	if err != nil {
		return err
	}
	globalLogger.Init()
	return nil
}

func newLoggerWithWriter(w io.Writer, fixedValues ...logs.FieldValue) logs.Logger {
	l := logs.SimpleLogger{FieldsValues: fixedValues[:]}
	l.SetWriter(w)
	return &l
}

func newLoggerWithLogFile(filename string, fixedValues ...logs.FieldValue) (logs.Logger, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return newLoggerWithWriter(f, fixedValues...), nil
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

