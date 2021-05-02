package logs

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var logWriter = logWriterTest{}

func TestShouldLogSimpleMessage(t *testing.T) {
	setup()
	Info("teste")
	logWriter.assertLogMessage(t, "INFO * teste\n")
	Error("teste")
	logWriter.assertLogMessage(t, "ERROR * teste\n")
}

func TestShouldLogFormattedMessage(t *testing.T) {
	setup()
	Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO * teste txt 10\n")
	Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR * teste txt 10\n")
}

func TestShouldLogSimpleMessageWithFixedFields(t *testing.T) {
	setup()
	l := NewLogger(FixedValue("reqid", "1"), FixedValue("idtperson", "2"))
	l.Info("teste")
	logWriter.assertLogMessage(t, "INFO [reqid: 1] [idtperson: 2] * teste\n")
	l.Error("teste")
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] [idtperson: 2] * teste\n")
}

func TestShouldLogFormattedMessageWithFixedFields(t *testing.T) {
	setup()
	l := NewLogger(FixedValue("reqid", "1"), FixedValue("idtperson", "2"))
	l.Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO [reqid: 1] [idtperson: 2] * teste txt 10\n")
	l.Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] [idtperson: 2] * teste txt 10\n")
}

func TestShouldLogBasedInGlobalLoggerWithFixedFields(t *testing.T) {
	setup()
	SetGlobalLogger(NewLogger(FixedValue("reqid", "1")))
	Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO [reqid: 1] * teste txt 10\n")
	Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] * teste txt 10\n")

	l := NewChildLogger(FixedValue("idtperson", "2"))
	l.Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO [reqid: 1] [idtperson: 2] * teste txt 10\n")
	l.Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] [idtperson: 2] * teste txt 10\n")
}

func TestShouldLogInFile(t *testing.T) {
	setup()
	f, err := ioutil.TempFile("", "test-go-log.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	err = InitLogFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	Info("teste")
	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	const expected = "INFO * teste\n"
	if !strings.HasSuffix(string(b), expected) {
		t.Fatalf("Expected '%s', but '%s' was logged\n", expected, string(b))
	}
}

func setup() {
	SetGlobalLogger(NewLogger())
	Init(&logWriter)
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
		t.Fatalf("Expected '%s', but '%s' was logged\n", w.lastLog, m)
	}
}

