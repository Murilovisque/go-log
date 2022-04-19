package logs

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"

	logs "github.com/Murilovisque/logs/v3/internal"
)

var (
	globalLogger    logs.Logger
	levelSelected   logs.LoggerLevelMode = logs.LogDebugMode
	ErrInvalidLevel                      = errors.New("invalid logger level mode")
)

func init() {
	log.SetFlags(log.LstdFlags)
	err := initGlobalLogger(levelSelected, &logs.SimpleLogger{FieldsValues: []logs.FieldValue{}, LevelSelected: levelSelected})
	if err != nil {
		panic(err)
	}
}

func InitWithLogFile(level logs.LoggerLevelMode, filename string, fixedValues ...logs.FieldValue) error {
	l, err := newLoggerWithLogFile(level, filename, fixedValues...)
	if err != nil {
		return err
	}
	return initGlobalLogger(level, l)
}

func InitWithWriter(level logs.LoggerLevelMode, w io.Writer, fixedValues ...logs.FieldValue) error {
	return initGlobalLogger(level, newLoggerWithWriter(level, w, fixedValues...))
}

func NewChildLogger(fixedValues ...logs.FieldValue) Logger {
	globalFixedValues := globalLogger.FixedFieldsValues()[:]
	globalFixedValues = append(globalFixedValues, fixedValues...)
	l := logs.SimpleLogger{FieldsValues: globalFixedValues, LevelSelected: levelSelected}
	l.Init()
	return &l
}

func NewChildLoggerFrom(parentLogger Logger, fixedValues ...logs.FieldValue) Logger {
	parentFixedValues := parentLogger.FixedFieldsValues()[:]
	parentFixedValues = append(parentFixedValues, fixedValues...)
	l := logs.SimpleLogger{FieldsValues: parentFixedValues, LevelSelected: levelSelected}
	l.Init()
	return &l
}

func Close() {
	globalLogger.Close()
}

func StringToLoggerLevelMode(level string) (logs.LoggerLevelMode, error) {
	level = strings.ToUpper(level)
	for _, l := range logs.LogsMode {
		if string(l) == level {
			return l, nil
		}
	}
	return "", ErrInvalidLevel
}

func initGlobalLogger(level logs.LoggerLevelMode, l logs.Logger) error {
	levelSelected = level
	globalLogger = l
	globalLogger.Init()
	Infof("Log initialized with level %v", level)
	return nil
}

func newLoggerWithWriter(level logs.LoggerLevelMode, w io.Writer, fixedValues ...logs.FieldValue) logs.Logger {
	l := logs.SimpleLogger{FieldsValues: fixedValues[:], LevelSelected: level}
	l.SetWriter(w)
	return &l
}

func newLoggerWithLogFile(level logs.LoggerLevelMode, filename string, fixedValues ...logs.FieldValue) (logs.Logger, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return newLoggerWithWriter(level, f, fixedValues...), nil
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
