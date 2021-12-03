package logs

import (
	"strings"
	"testing"

)

var (
	logWriter = logWriterTest{}
	sl = SimpleLogger{ LevelSelected: LogDebugMode }
)

func TestShouldLogSimpleMessage(t *testing.T) {
	setup()
	sl.Info("teste")
	logWriter.assertLogMessage(t, "INFO * teste\n")
	sl.Error("teste")
	logWriter.assertLogMessage(t, "ERROR * teste\n")
}

func TestShouldLogFormattedMessage(t *testing.T) {
	setup()
	sl.Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO * teste txt 10\n")
	sl.Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR * teste txt 10\n")
}

func TestShouldLogSimpleMessageWithFixedFields(t *testing.T) {
	setup(FieldValue{"reqid", "1"}, FieldValue{"idtperson", "2"})
	sl.Info("teste")
	logWriter.assertLogMessage(t, "INFO [reqid: 1] [idtperson: 2] * teste\n")
	sl.Error("teste")
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] [idtperson: 2] * teste\n")
}

func TestShouldLogFormattedMessageWithFixedFields(t *testing.T) {
	setup(FieldValue{"reqid", "1"}, FieldValue{"idtperson", "2"})
	sl.Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO [reqid: 1] [idtperson: 2] * teste txt 10\n")
	sl.Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] [idtperson: 2] * teste txt 10\n")
}

func setup(fixedValues ...FieldValue) {
	sl = SimpleLogger{ FieldsValues: fixedValues, LevelSelected: LogDebugMode }
	sl.SetWriter(&logWriter)
	sl.Init()
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
