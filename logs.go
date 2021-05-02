package logs

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const (
	logInfoPrefix  = "INFO"
	logErrorPrefix = "ERROR"
	logFatalPrefix = "FATAL"
)

var (
	globalLogger *Logger
)

func init() {
	SetGlobalLogger(NewLogger())
	log.SetFlags(log.LstdFlags)
}

// Init to log using the Writer
func Init(writer io.Writer) {
	log.SetOutput(writer)
}

// InitLogFile to log using file
func InitLogFile(filename string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	Init(f)
	return nil
}

// NewLogger gets a new Logger
func NewLogger(fixedValues ...FieldValue) *Logger {
	return &Logger{fixedValues[:]}
}

// NewChildLogger gets a new Logger based in global Logger
func NewChildLogger(fixedValues ...FieldValue) *Logger {
	globalFixedValues := globalLogger.fixedValues[:]
	globalFixedValues = append(globalFixedValues, fixedValues...)
	return &Logger{globalFixedValues}
}

// SetGlobalLogger set global logger
func SetGlobalLogger(l *Logger) {
	globalLogger = l
}

// FixedValue gets a FieldValue
func FixedValue(key string, val interface{}) FieldValue {
	return FieldValue{Key: key, Val: val}
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

type Logger struct {
	fixedValues []FieldValue
}

func (l *Logger) Fatalf(message string, v ...interface{}) {
	log.Fatal(l.format(logFatalPrefix, message, v...))
}

func (l *Logger) Infof(message string, v ...interface{}) {
	log.Println(l.format(logInfoPrefix, message, v...))
}

func (l *Logger) Errorf(message string, v ...interface{}) {
	log.Println(l.format(logErrorPrefix, message, v...))
}

func (l *Logger) Fatal(message interface{}) {
	log.Fatal(l.print(logFatalPrefix, message))
}

func (l *Logger) Info(message interface{}) {
	log.Println(l.print(logInfoPrefix, message))
}

func (l *Logger) Error(message interface{}) {
	log.Println(l.print(logErrorPrefix, message))
}

func (l *Logger) print(level string, message interface{}) string {
	builder := l.logTemplate(level)
	builder.WriteString(fmt.Sprint(message))
	return builder.String()
}

func (l *Logger) format(level, message string, v ...interface{}) string {
	builder := l.logTemplate(level)
	builder.WriteString(fmt.Sprintf(message, v...))
	return builder.String()
}

func (l *Logger) logTemplate(level string) *strings.Builder {
	builder := strings.Builder{}
	builder.WriteString(level)
	for _, fv := range l.fixedValues {
		builder.WriteString(" [")
		builder.WriteString(fv.Key)
		builder.WriteString(": ")
		builder.WriteString(fmt.Sprint(fv.Val))
		builder.WriteString("]")
	}
	builder.WriteString(" * ")
	return &builder
}

type FieldValue struct {
	Key string
	Val interface{}
}
