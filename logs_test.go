package logs

import (
	"strings"
	"testing"

	logs "github.com/Murilovisque/logs/v3/internal"
)

var (
	logWriter = logWriterTest{}
	sl        = logs.SimpleLogger{LevelSelected: logs.LogDebugMode}
)

func TestShouldLogBasedInGlobalLoggerWithFixedFields(t *testing.T) {
	InitWithWriter(logs.LogDebugMode, &logWriter, FixedFieldValue("reqid", "1"))
	Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO [reqid: 1] * teste txt 10\n")
	Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] * teste txt 10\n")

	l := NewChildLogger(FixedFieldValue("idtperson", "2"))
	l.Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO [reqid: 1] [idtperson: 2] * teste txt 10\n")
	l.Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] [idtperson: 2] * teste txt 10\n")
}

func TestShouldLogUntilWarn(t *testing.T) {
	InitWithWriter(logs.LogWarnMode, &logWriter)
	Error("teste")
	logWriter.assertLogMessage(t, "ERROR * teste\n")
	Warn("teste")
	logWriter.assertLogMessage(t, "WARN * teste\n")
	Info("teste")
	logWriter.assertLogMessage(t, "WARN * teste\n")
	Debug("teste")
	logWriter.assertLogMessage(t, "WARN * teste\n")
}

func setup(fixedValues ...logs.FieldValue) {
}

type logWriterTest struct {
	lastLog string
}

func (w *logWriterTest) Write(p []byte) (n int, err error) {
	w.lastLog = string(p)
	return len(p), nil
}

func (w *logWriterTest) assertLogMessage(t *testing.T, m string) {
	if !strings.HasSuffix(w.lastLog, m) {
		t.Fatalf("Expected '%s', but '%s' was logged\n", m, w.lastLog)
	}
}
